package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveIsDeterministicAndPreservesExecutableMode(t *testing.T) {
	identity := packageIdentity{Version: candidateVersion, Channel: candidateChannel}
	files := []file{{name: "bin/l7", data: []byte(launcher), mode: 0o755}, {name: "PERMISSIONS.json", data: permissions("codex", identity), mode: 0o644}}
	first, err := archive(files)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := archive(files)
	if !bytes.Equal(first, second) {
		t.Fatal("archive bytes changed for identical inputs")
	}
	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil || len(reader.File) != 2 || reader.File[0].Mode().Perm() != 0o755 {
		t.Fatalf("reader=%v err=%v", reader, err)
	}
}

func TestCandidateMetadataRemainsUnsignedAndDefaultOff(t *testing.T) {
	identity := packageIdentity{Version: candidateVersion, Channel: candidateChannel}
	var permission map[string]any
	if err := json.Unmarshal(permissions("claude", identity), &permission); err != nil {
		t.Fatal(err)
	}
	network := permission["network"].(map[string]any)
	if network["default"] != "off" || permission["telemetry"] != false || permission["local_mcp"] != true {
		t.Fatalf("permissions=%v", permission)
	}
	var attestation map[string]any
	if err := json.Unmarshal(provenance("claude", identity), &attestation); err != nil || attestation["release_blocked"] != true || attestation["authority"] != "external-only" {
		t.Fatalf("attestation=%v err=%v", attestation, err)
	}
	for _, forbidden := range []string{"signed", "notarized", "published", "release_ready", "approval"} {
		if _, ok := attestation[forbidden]; ok {
			t.Fatalf("candidate-controlled provenance contains %q: %v", forbidden, attestation)
		}
	}
}

func TestPackageIdentityAcceptsOnlyDeclaredPairs(t *testing.T) {
	for _, identity := range []packageIdentity{
		{Version: candidateVersion, Channel: candidateChannel},
		{Version: releaseVersion, Channel: releaseChannel},
	} {
		if observed, err := validatePackageIdentity(identity.Version, identity.Channel); err != nil || observed != identity {
			t.Fatalf("identity=%v observed=%v err=%v", identity, observed, err)
		}
	}
	for _, identity := range []packageIdentity{
		{Version: candidateVersion, Channel: releaseChannel},
		{Version: releaseVersion, Channel: candidateChannel},
		{Version: "v1.0.0", Channel: releaseChannel},
		{Version: "1.0.0+build", Channel: releaseChannel},
		{Version: releaseVersion, Channel: "Stable"},
	} {
		if _, err := validatePackageIdentity(identity.Version, identity.Channel); err == nil {
			t.Fatalf("unsupported identity was accepted: %v", identity)
		}
	}
}

func TestDecodedGoIdentityRejectsVersionAndArchitectureBinarySubstitution(t *testing.T) {
	path := "github.com/addressanup/level7-dev-loop/cmd/l7"
	module := "github.com/addressanup/level7-dev-loop"
	if err := validateDecodedGoIdentity(path, module, releaseVersion, releaseVersion); err != nil {
		t.Fatal(err)
	}
	for _, value := range []struct{ path, module, version string }{
		{path, module, candidateVersion},
		{"example.invalid/substitute", module, releaseVersion},
		{path, "example.invalid/substitute", releaseVersion},
	} {
		if err := validateDecodedGoIdentity(value.path, value.module, value.version, releaseVersion); err == nil {
			t.Fatalf("substituted binary identity passed: %+v", value)
		}
	}
}

func TestRunRejectsMixedIdentityBeforeWritingOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output")
	var stdout, stderr bytes.Buffer
	code := run([]string{"--input", "missing", "--output", output, "--version", releaseVersion, "--channel", candidateChannel}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "package identity") || stdout.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Lstat(output); !os.IsNotExist(err) {
		t.Fatalf("rejected identity created output: %v", err)
	}
}

func TestMarketplaceCatalogsResolveOnlyTheExtractedPackage(t *testing.T) {
	identity := packageIdentity{Version: releaseVersion, Channel: releaseChannel}
	var codex struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name   string `json:"name"`
			Source struct {
				Source string `json:"source"`
				Path   string `json:"path"`
			} `json:"source"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(marketplace("codex", identity), &codex); err != nil || codex.Name != v1MarketplaceName || len(codex.Plugins) != 1 ||
		codex.Plugins[0].Name != v1PackageName || codex.Plugins[0].Source.Source != "local" || codex.Plugins[0].Source.Path != "." {
		t.Fatalf("codex catalog=%+v err=%v", codex, err)
	}
	var claude struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Version string `json:"version"`
			Strict  bool   `json:"strict"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(marketplace("claude", identity), &claude); err != nil || claude.Name != v1MarketplaceName || len(claude.Plugins) != 1 ||
		claude.Plugins[0].Name != v1PackageName || claude.Plugins[0].Source != "." || claude.Plugins[0].Version != releaseVersion || !claude.Plugins[0].Strict {
		t.Fatalf("claude catalog=%+v err=%v", claude, err)
	}
	if marketplacePath("codex") != ".agents/plugins/marketplace.json" || marketplacePath("claude") != ".claude-plugin/marketplace.json" {
		t.Fatal("host catalog paths changed")
	}
}

func TestLauncherChoosesOnlySupportedArchitectures(t *testing.T) {
	if !bytes.Contains([]byte(launcher), []byte("darwin-arm64")) || !bytes.Contains([]byte(launcher), []byte("darwin-amd64")) || bytes.Contains([]byte(launcher), []byte("curl")) {
		t.Fatal("launcher contract is invalid")
	}
}

func TestMCPConfigurationUsesEachHostPluginRootContract(t *testing.T) {
	if bytes.Contains(mcpConfiguration("codex"), []byte("CLAUDE_PLUGIN_ROOT")) || !bytes.Contains(mcpConfiguration("codex"), []byte(`"command": "./bin/l7"`)) {
		t.Fatalf("Codex MCP path is not plugin-relative: %s", mcpConfiguration("codex"))
	}
	if !bytes.Contains(mcpConfiguration("claude"), []byte(`${CLAUDE_PLUGIN_ROOT}/bin/l7`)) {
		t.Fatalf("Claude MCP path does not use the supported plugin root substitution: %s", mcpConfiguration("claude"))
	}
}

func TestSBOMUsesSPDXFilesRelationshipsAndPinnedParserDependency(t *testing.T) {
	identity := packageIdentity{Version: releaseVersion, Channel: releaseChannel}
	data := sbom("codex", identity, []checksum{{Path: "bin/l7", SHA256: strings.Repeat("a", 64), Size: 1}})
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	packages, _ := document["packages"].([]any)
	files, _ := document["files"].([]any)
	relationships, _ := document["relationships"].([]any)
	if len(packages) != 2 || len(files) != 1 || len(relationships) != 3 || !bytes.Contains(data, []byte("gotreesitter@v0.24.0")) || document["name"] != "level7-dev-loop-1.0.0-codex" {
		t.Fatalf("incomplete SPDX document: %s", data)
	}
	if bytes.Contains(data, []byte(`"externalRefs": [{"path"`)) {
		t.Fatalf("file inventory was encoded as invalid package externalRefs: %s", data)
	}
}

func readZip(t *testing.T, value []byte, name string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(value), int64(len(value)))
	if err != nil {
		t.Fatal(err)
	}
	for _, current := range reader.File {
		if current.Name == name {
			file, err := current.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			return data
		}
	}
	t.Fatalf("missing %s", name)
	return nil
}
