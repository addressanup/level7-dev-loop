package plugininstall

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

const (
	fixtureCommit  = "ed1457f2c54e2308f757d7ea3bf4fcfa113575ec"
	fixtureTree    = "cb596fe99648ec545bb924258896434bc1c87e73"
	fixtureRunID   = "38042885014"
	fixtureCodexV  = "codex-cli 0.162.1"
	fixtureClaudeV = "2.1.291 (Claude Code)"
)

var fixtureSkills = []string{"l7-build", "l7-change", "l7-constitution", "l7-crew", "l7-cyber", "l7-deploy", "l7-experience", "l7-geometry", "l7-greenfield", "l7-headless", "l7-next", "l7-onboard", "l7-ops", "l7-release", "l7-review", "l7-storybook", "l7-sync"}

// fixtureLauncher mirrors the real launcher but answers --version itself so
// the architecture binaries can stay inert Mach-O stubs; the production
// verifyLauncher still runs it through /bin/sh.
func fixtureLauncher(version string) string {
	return "#!/bin/sh\nset -eu\nscript_dir=$(CDPATH='' cd -- \"$(dirname -- \"$0\")\" && pwd -P)\ncase $(uname -m) in\n  arm64|aarch64) platform=darwin-arm64 ;;\n  x86_64|amd64) platform=darwin-amd64 ;;\n  *) printf 'unsupported\\n' >&2; exit 1 ;;\nesac\nplatform_dir=\"$script_dir/$platform\"\n" +
		"if [ \"${1:-}\" = --version ]; then printf '%s\\n' '{\"schema\":4,\"outcome\":\"PASS\",\"code\":\"L7-CLI-000\",\"command\":\"version\",\"state\":\"available\",\"version\":\"" + version + "\",\"message\":\"fixture\",\"next\":\"\",\"details\":[]}'; exit 0; fi\n" +
		"PATH=\"$platform_dir:$PATH\"\nexport PATH\nexec \"$platform_dir/l7\" \"$@\"\n"
}

type fixtureBundle struct {
	Subject  string `json:"subject"`
	SHA256   string `json:"sha256"`
	Commit   string `json:"commit"`
	Workflow string `json:"workflow"`
	Tamper   string `json:"tamper,omitempty"`
}

// fixtureRelease is one synthetic release with both host archives.
type fixtureRelease struct {
	Tag        string
	Version    string
	Prerelease bool
	Manifest   string
	Assets     map[string][]byte
	Digests    map[string]string
	Commit     string
}

var (
	fixtureOnce     sync.Once
	fixtureReleases map[string]*fixtureRelease
	fixtureArch     string
	fixtureErr      error
)

// fixtures builds the release set once per test binary: stable 1.0.0 and
// 1.0.1, prerelease 1.1.0-dev, and 0.9.0 for downgrade tests.
func fixtures(t *testing.T) map[string]*fixtureRelease {
	t.Helper()
	fixtureOnce.Do(func() {
		fixtureArch = runtime.GOARCH
		if fixtureArch != "arm64" && fixtureArch != "amd64" {
			fixtureArch = "arm64"
		}
		fixtureReleases = map[string]*fixtureRelease{}
		for _, spec := range []struct {
			version    string
			prerelease bool
			commit     string
		}{{"0.9.0", false, "aa" + fixtureCommit[2:]}, {"1.0.0", false, fixtureCommit}, {"1.0.1", false, "bb" + fixtureCommit[2:]}, {"1.1.0-dev", true, "cc" + fixtureCommit[2:]}} {
			release, err := buildFixtureRelease(spec.version, spec.prerelease, spec.commit)
			if err != nil {
				fixtureErr = err
				return
			}
			fixtureReleases[spec.version] = release
		}
	})
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	return fixtureReleases
}

func buildFixtureRelease(version string, prerelease bool, commit string) (*fixtureRelease, error) {
	binaries := map[string][]byte{}
	for _, arch := range []string{"arm64", "amd64"} {
		binaries["darwin-"+arch+"/l7"] = machOStub(arch, "l7-"+version)
		binaries["darwin-"+arch+"/l7-embed"] = machOStub(arch, "embed-"+version)
	}
	inputs := map[string]string{}
	for name, data := range binaries {
		inputs[name] = digestHex(data)
	}
	release := &fixtureRelease{Tag: "v" + version, Version: version, Prerelease: prerelease, Assets: map[string][]byte{}, Digests: map[string]string{}, Commit: commit}
	for _, host := range []string{"codex", "claude"} {
		archive, err := buildFixtureArchive(host, version, binaries)
		if err != nil {
			return nil, err
		}
		name := "level7-dev-loop-" + version + "-" + host + ".zip"
		release.Assets[name] = archive
		release.Digests[name] = digestHex(archive)
	}
	claude, codex := "level7-dev-loop-"+version+"-claude.zip", "level7-dev-loop-"+version+"-codex.zip"
	sums := []byte(release.Digests[claude] + "  " + claude + "\n" + release.Digests[codex] + "  " + codex + "\n")
	release.Assets["SHA256SUMS"] = sums
	release.Digests["SHA256SUMS"] = digestHex(sums)
	artifacts := []any{
		map[string]any{"name": claude, "sha256": release.Digests[claude], "size": len(release.Assets[claude]), "developer_id_signature": "NOT_RUN", "notarization": map[string]any{"status": "NOT_RUN", "id": nil}},
		map[string]any{"name": codex, "sha256": release.Digests[codex], "size": len(release.Assets[codex]), "developer_id_signature": "NOT_RUN", "notarization": map[string]any{"status": "NOT_RUN", "id": nil}},
	}
	var manifest map[string]any
	if prerelease {
		release.Manifest = "UNSIGNED-PRERELEASE-MANIFEST.json"
		unsignedInputs := []any{}
		for _, path := range []string{"darwin-amd64/l7", "darwin-amd64/l7-embed", "darwin-arm64/l7", "darwin-arm64/l7-embed"} {
			unsignedInputs = append(unsignedInputs, map[string]string{"path": path, "sha256": inputs[path], "signature_state": "NONE"})
		}
		manifest = map[string]any{
			"schema": 1, "version": version, "tag": release.Tag, "channel": "development-candidate", "change_id": "l7-v1-explicitly-unsigned-prerelease",
			"candidate":            map[string]any{"base": fixtureTree, "commit": commit, "pull_request": 19, "tree": fixtureTree},
			"workflow":             map[string]any{"repository": pluginrelease.Repository, "run_id": fixtureRunID, "run_attempt": "1"},
			"developer_id_signing": "NOT_PERFORMED", "notarization": "NOT_PERFORMED", "support": "WITHHELD", "provider_execution": "NOT_RUN", "gatekeeper": "MAY_BLOCK",
			"artifacts": artifacts, "checksums": map[string]any{"name": "SHA256SUMS", "sha256": release.Digests["SHA256SUMS"], "size": len(sums)},
			"unsigned_inputs": unsignedInputs, "unsigned_packages": map[string]string{claude: release.Digests[claude], codex: release.Digests[codex]},
			"publication_blocked": true,
		}
	} else {
		release.Manifest = "RELEASE-MANIFEST.json"
		manifest = map[string]any{
			"schema": 1, "change_id": "unsigned-stable-release", "version": version, "tag": release.Tag, "release_channel": "stable",
			"policy": "owner-authorized-unsigned-stable-v1", "artifact_state": "unsigned-unnotarized-prepared", "publication_blocked": true, "gatekeeper": "MAY_BLOCK; NEVER_DISABLE_GLOBALLY",
			"candidate": map[string]any{"base": fixtureTree, "commit": commit, "tree": fixtureTree, "pull_request": 27},
			"workflow":  map[string]any{"repository": pluginrelease.Repository, "path": ".github/workflows/release.yml", "run_id": fixtureRunID, "run_attempt": "1"},
			"assurance": map[string]any{"mode": "solo", "owner": "addressanup", "operator": "addressanup"}, "authorization": "https://github.com/addressanup/level7-dev-loop/pull/27#issuecomment-1",
			"waivers":         map[string]string{"apple_signing_and_notarization": "WAIVED", "protected_holdout_evaluation": "WAIVED", "separate_operator_and_protected_approval": "WAIVED", "exact_asset_provider_trials": "WAIVED"},
			"qualification":   map[string]string{"developer_id_signing": "NOT_RUN", "apple_notarization": "NOT_RUN", "protected_holdout_evaluation": "NOT_RUN", "exact_asset_provider_trials": "NOT_RUN", "independent_code_audit": "NOT_RUN", "formal_support": "WITHHELD"},
			"reproducibility": map[string]string{"result": "PASS", "toolchain": "go1.26.7"}, "unsigned_inputs": inputs,
			"unsigned_packages": map[string]string{claude: release.Digests[claude], codex: release.Digests[codex]},
			"artifacts":         artifacts, "checksums": map[string]any{"name": "SHA256SUMS", "sha256": release.Digests["SHA256SUMS"], "size": len(sums)},
			"release_notes": map[string]string{"path": "docs/releases/v" + version + ".md", "sha256": strings.Repeat("0", 64)},
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	release.Assets[release.Manifest] = data
	release.Digests[release.Manifest] = digestHex(data)
	return release, nil
}

// machOStub is a minimal 64-bit Mach-O header with no load commands; it
// parses as the requested CPU type but is never executed.
func machOStub(arch, salt string) []byte {
	cpu := uint32(0x0100000c)
	if arch == "amd64" {
		cpu = 0x01000007
	}
	header := make([]byte, 32)
	binary.LittleEndian.PutUint32(header[0:], 0xfeedfacf)
	binary.LittleEndian.PutUint32(header[4:], cpu)
	binary.LittleEndian.PutUint32(header[8:], 0)
	binary.LittleEndian.PutUint32(header[12:], 2)
	return append(header, []byte("\n"+salt+"\n")...)
}

func buildFixtureArchive(host, version string, binaries map[string][]byte) ([]byte, error) {
	type entry struct {
		name string
		data []byte
		mode os.FileMode
	}
	entries := []entry{}
	for _, skill := range fixtureSkills {
		entries = append(entries, entry{"skills/" + skill + "/SKILL.md", []byte("---\nname: " + skill + "\n---\n# " + skill + " " + version + "\n"), 0o644})
	}
	for name, data := range binaries {
		entries = append(entries, entry{"bin/" + name, data, 0o755})
	}
	mcp := map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "./bin/l7", "args": []string{"mcp"}}}}
	manifestName := ".codex-plugin/plugin.json"
	marketplaceName := ".agents/plugins/marketplace.json"
	marketplace := map[string]any{"name": "level7-engineering-v1", "interface": map[string]any{"displayName": "Level 7 Engineering v1"}, "plugins": []any{map[string]any{"name": "level7-dev-loop", "source": map[string]any{"source": "local", "path": "."}, "policy": map[string]any{"installation": "AVAILABLE", "authentication": "ON_INSTALL"}, "category": "Developer Tools"}}}
	if host == "claude" {
		mcp = map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "${CLAUDE_PLUGIN_ROOT}/bin/l7", "args": []string{"mcp"}}}}
		manifestName = ".claude-plugin/plugin.json"
		marketplaceName = ".claude-plugin/marketplace.json"
		marketplace = map[string]any{"name": "level7-engineering-v1", "description": "fixture", "owner": map[string]any{"name": "Level 7 Engineering"}, "plugins": []any{map[string]any{"name": "level7-dev-loop", "source": ".", "description": "fixture", "version": version, "license": "MIT", "category": "Development Tools", "strict": true}}}
	}
	entries = append(entries,
		entry{"bin/l7", []byte(fixtureLauncher(version)), 0o755},
		entry{".mcp.json", mustJSON(mcp), 0o644},
		entry{"LICENSE", []byte("MIT fixture\n"), 0o644},
		entry{"README.md", []byte("# fixture " + version + "\n"), 0o644},
		entry{"CHANGELOG.md", []byte("# changes " + version + "\n"), 0o644},
		entry{"PERMISSIONS.json", mustJSON(map[string]any{"schema": 1, "version": version, "host": host}), 0o644},
		entry{"PROVENANCE.input.json", mustJSON(map[string]any{"schema": 2, "version": version, "host": host, "channel": "stable", "artifact_state": "package-input", "authority": "external-only", "release_blocked": true, "next": "fixture"}), 0o644},
		entry{manifestName, mustJSON(map[string]any{"name": "level7-dev-loop", "version": version, "description": "fixture"}), 0o644},
		entry{marketplaceName, mustJSON(marketplace), 0o644},
	)
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	checks := []any{}
	for _, current := range entries {
		checks = append(checks, map[string]any{"path": current.name, "sha256": digestHex(current.data), "size": len(current.data)})
	}
	entries = append(entries,
		entry{"CHECKSUMS.json", mustJSON(map[string]any{"schema": 1, "version": version, "files": checks}), 0o644},
		entry{"SBOM.spdx.json", mustJSON(map[string]any{"spdxVersion": "SPDX-2.3", "name": "level7-dev-loop-" + version + "-" + host}), 0o644},
	)
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, current := range entries {
		header := &zip.FileHeader{Name: current.name, Method: zip.Store}
		header.SetMode(current.mode)
		file, err := writer.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := file.Write(current.data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func mustJSON(value any) []byte {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(data, '\n')
}

func digestHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// releaseServer serves the fixture releases the way api.github.com and the
// release download endpoint do, with switchable failures.
type releaseServer struct {
	server   *httptest.Server
	mu       sync.Mutex
	releases []*fixtureRelease
	// Overrides inject failures: missing attestations, corrupt bytes, HTTP
	// status codes, tampered identities, and redirect destinations.
	corrupt       map[string]bool
	noAttestation map[string]bool
	status        map[string]int
	tamper        map[string]string
	redirectTo    string
	requests      []string
}

func newReleaseServer(t *testing.T, releases ...*fixtureRelease) *releaseServer {
	t.Helper()
	rs := &releaseServer{releases: releases, corrupt: map[string]bool{}, noAttestation: map[string]bool{}, status: map[string]int{}, tamper: map[string]string{}}
	rs.server = httptest.NewServer(http.HandlerFunc(rs.handle))
	t.Cleanup(rs.server.Close)
	return rs
}

func (rs *releaseServer) source() pluginrelease.Source {
	return pluginrelease.Source{APIBase: rs.server.URL, DownloadBase: rs.server.URL, AllowedHosts: []string{"127.0.0.1"}, InsecureLoopback: true}
}

func (rs *releaseServer) handle(response http.ResponseWriter, request *http.Request) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.requests = append(rs.requests, request.URL.Path)
	if request.Header.Get("Authorization") != "" || request.Header.Get("Cookie") != "" {
		http.Error(response, "credentials must never be sent", http.StatusBadRequest)
		return
	}
	path := request.URL.Path
	prefix := "/repos/" + pluginrelease.Repository
	switch {
	case path == prefix+"/releases":
		entries := []any{}
		for _, release := range rs.releases {
			entries = append(entries, rs.apiRelease(release))
		}
		writeJSON(response, entries)
	case strings.HasPrefix(path, prefix+"/releases/tags/"):
		tag := strings.TrimPrefix(path, prefix+"/releases/tags/")
		for _, release := range rs.releases {
			if release.Tag == tag {
				writeJSON(response, rs.apiRelease(release))
				return
			}
		}
		http.NotFound(response, request)
	case strings.HasPrefix(path, prefix+"/attestations/sha256:"):
		digest := strings.TrimPrefix(path, prefix+"/attestations/sha256:")
		if rs.noAttestation[digest] {
			writeJSON(response, map[string]any{"attestations": []any{}})
			return
		}
		for _, release := range rs.releases {
			for name, assetDigest := range release.Digests {
				if assetDigest == digest {
					workflow, _ := pluginrelease.WorkflowForManifest(release.Manifest)
					bundle := fixtureBundle{Subject: name, SHA256: digest, Commit: release.Commit, Workflow: workflow.Path, Tamper: rs.tamper[name]}
					writeJSON(response, map[string]any{"attestations": []any{map[string]any{"bundle": map[string]any{"mediaType": "application/vnd.dev.sigstore.bundle.v0.3+json", "l7test": bundle}, "repository_id": 1349427189}}})
					return
				}
			}
		}
		writeJSON(response, map[string]any{"attestations": []any{}})
	case strings.HasPrefix(path, "/"+pluginrelease.Repository+"/releases/download/"):
		rest := strings.TrimPrefix(path, "/"+pluginrelease.Repository+"/releases/download/")
		tag, asset, ok := strings.Cut(rest, "/")
		if !ok {
			http.NotFound(response, request)
			return
		}
		if rs.redirectTo != "" && !strings.HasPrefix(request.URL.Path, "/redirected") {
			http.Redirect(response, request, rs.redirectTo+path, http.StatusFound)
			return
		}
		rs.serveAsset(response, request, tag, asset)
	case strings.HasPrefix(path, "/redirected/"):
		rest := strings.TrimPrefix(path, "/redirected/"+pluginrelease.Repository+"/releases/download/")
		tag, asset, _ := strings.Cut(rest, "/")
		rs.serveAsset(response, request, tag, asset)
	default:
		http.NotFound(response, request)
	}
}

func (rs *releaseServer) serveAsset(response http.ResponseWriter, request *http.Request, tag, asset string) {
	if status := rs.status[asset]; status != 0 {
		http.Error(response, "injected", status)
		return
	}
	for _, release := range rs.releases {
		if release.Tag != tag {
			continue
		}
		data, ok := release.Assets[asset]
		if !ok {
			break
		}
		if rs.corrupt[asset] {
			data = append([]byte{}, data...)
			data[len(data)/2] ^= 0xff
		}
		response.Header().Set("Content-Length", fmt.Sprint(len(data)))
		_, _ = response.Write(data)
		return
	}
	http.NotFound(response, request)
}

func (rs *releaseServer) apiRelease(release *fixtureRelease) map[string]any {
	assets := []any{}
	names := make([]string, 0, len(release.Assets))
	for name := range release.Assets {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		assets = append(assets, map[string]any{"name": name, "size": len(release.Assets[name]), "browser_download_url": rs.server.URL + "/" + pluginrelease.Repository + "/releases/download/" + release.Tag + "/" + name})
	}
	return map[string]any{"tag_name": release.Tag, "draft": false, "prerelease": release.Prerelease, "immutable": true, "assets": assets}
}

func writeJSON(response http.ResponseWriter, value any) {
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(value)
}

func readAllString(reader io.Reader) string {
	data, _ := io.ReadAll(reader)
	return string(data)
}
