// Command l7pack assembles deterministic, unsigned v1 development candidates.
// It never publishes, signs, or upgrades the stable channel.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const candidateVersion = "1.0.0-dev"

var canonicalSkills = []string{
	"l7-build", "l7-change", "l7-constitution", "l7-cyber", "l7-deploy", "l7-experience", "l7-geometry", "l7-greenfield",
	"l7-headless", "l7-next", "l7-onboard", "l7-ops", "l7-release", "l7-review", "l7-storybook", "l7-sync",
}

type file struct {
	name string
	data []byte
	mode os.FileMode
}

type checksum struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("l7pack", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	input := flags.String("input", "", "prebuilt architecture input root")
	output := flags.String("output", "", "candidate output directory")
	if flags.Parse(arguments) != nil || flags.NArg() != 0 || *input == "" || *output == "" {
		return 2
	}
	physicalRoot, err := filepath.EvalSymlinks(*root)
	if err != nil {
		fmt.Fprintln(stderr, "l7pack: repository root is unavailable")
		return 1
	}
	physicalInput, err := filepath.EvalSymlinks(*input)
	if err != nil {
		fmt.Fprintln(stderr, "l7pack: prebuilt input is unavailable")
		return 1
	}
	outputPath, err := filepath.Abs(*output)
	if err != nil || outputPath == physicalRoot || strings.HasPrefix(physicalRoot, outputPath+string(filepath.Separator)) {
		fmt.Fprintln(stderr, "l7pack: output path is unsafe")
		return 1
	}
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		fmt.Fprintln(stderr, "l7pack: cannot create output")
		return 1
	}
	for _, host := range []string{"codex", "claude"} {
		files, err := packageFiles(physicalRoot, physicalInput, host)
		if err != nil {
			fmt.Fprintf(stderr, "l7pack: %s: %v\n", host, err)
			return 1
		}
		first, err := archive(files)
		if err != nil {
			fmt.Fprintf(stderr, "l7pack: %s archive: %v\n", host, err)
			return 1
		}
		second, _ := archive(files)
		if !bytes.Equal(first, second) {
			fmt.Fprintln(stderr, "l7pack: archive assembly is not reproducible")
			return 1
		}
		name := fmt.Sprintf("level7-dev-loop-%s-%s.zip", candidateVersion, host)
		if err := os.WriteFile(filepath.Join(outputPath, name), first, 0o644); err != nil {
			fmt.Fprintln(stderr, "l7pack: cannot write archive")
			return 1
		}
		digest := sha256.Sum256(first)
		fmt.Fprintf(stdout, "%s sha256=%x\n", name, digest)
	}
	return 0
}

func packageFiles(root, input, host string) ([]file, error) {
	if host != "codex" && host != "claude" {
		return nil, errors.New("unsupported host")
	}
	files := []file{}
	for _, skill := range canonicalSkills {
		data, err := boundedRead(filepath.Join(root, "skills", skill, "SKILL.md"), 1<<20)
		if err != nil || !bytes.Contains(data, []byte("name: "+skill+"\n")) {
			return nil, fmt.Errorf("canonical skill %s is invalid", skill)
		}
		files = append(files, file{name: "skills/" + skill + "/SKILL.md", data: data, mode: 0o644})
	}
	for _, arch := range []string{"arm64", "amd64"} {
		goBinaryPath := filepath.Join(input, "darwin-"+arch, "l7")
		embedPath := filepath.Join(input, "darwin-"+arch, "l7-embed")
		goBinary, err := boundedRead(goBinaryPath, 64<<20)
		if err != nil || validateMachO(goBinaryPath, arch) != nil {
			return nil, fmt.Errorf("darwin-%s Level 7 executable is invalid", arch)
		}
		embed, err := boundedRead(embedPath, 64<<20)
		if err != nil || validateMachO(embedPath, arch) != nil {
			return nil, fmt.Errorf("darwin-%s embedding helper is invalid", arch)
		}
		prefix := "bin/darwin-" + arch + "/"
		files = append(files, file{name: prefix + "l7", data: goBinary, mode: 0o755}, file{name: prefix + "l7-embed", data: embed, mode: 0o755})
	}
	license, err := boundedRead(filepath.Join(root, "LICENSE"), 1<<20)
	if err != nil {
		return nil, err
	}
	readme, err := boundedRead(filepath.Join(root, "README.md"), 4<<20)
	if err != nil {
		return nil, err
	}
	changelog, err := boundedRead(filepath.Join(root, "CHANGELOG.md"), 1<<20)
	if err != nil {
		return nil, err
	}
	files = append(files,
		file{name: "bin/l7", data: []byte(launcher), mode: 0o755},
		file{name: ".mcp.json", data: mcpConfiguration(host), mode: 0o644},
		file{name: "LICENSE", data: license, mode: 0o644},
		file{name: "README.md", data: readme, mode: 0o644},
		file{name: "CHANGELOG.md", data: changelog, mode: 0o644},
		file{name: "PERMISSIONS.json", data: permissions(host), mode: 0o644},
		file{name: "PROVENANCE.input.json", data: provenance(host), mode: 0o644},
	)
	manifestName := ".codex-plugin/plugin.json"
	if host == "claude" {
		manifestName = ".claude-plugin/plugin.json"
	}
	files = append(files, file{name: manifestName, data: manifest(host), mode: 0o644})
	checks := make([]checksum, 0, len(files))
	for _, current := range files {
		digest := sha256.Sum256(current.data)
		checks = append(checks, checksum{Path: current.name, SHA256: fmt.Sprintf("%x", digest), Size: len(current.data)})
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Path < checks[j].Path })
	files = append(files,
		file{name: "CHECKSUMS.json", data: jsonData(map[string]any{"schema": 1, "version": candidateVersion, "files": checks}), mode: 0o644},
		file{name: "SBOM.spdx.json", data: sbom(host, checks), mode: 0o644},
	)
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files, nil
}

func mcpConfiguration(host string) []byte {
	command := "./bin/l7"
	if host == "claude" {
		command = "${CLAUDE_PLUGIN_ROOT}/bin/l7"
	}
	return jsonData(map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": command, "args": []string{"mcp"}}}})
}

func manifest(host string) []byte {
	value := map[string]any{
		"name": "level7-dev-loop", "version": candidateVersion,
		"description": "Default-off multi-host orchestration for Codex, Claude Code, and configured API gateways.",
		"author":      map[string]any{"name": "Level 7 Engineering"}, "homepage": "https://github.com/addressanup/level7-dev-loop#readme",
		"repository": "https://github.com/addressanup/level7-dev-loop", "license": "MIT",
		"keywords": []string{"orchestration", "codex", "claude-code", "memory", "security", "headless"},
	}
	if host == "codex" {
		value["skills"] = "./skills/"
		value["mcpServers"] = "./.mcp.json"
		value["interface"] = map[string]any{
			"displayName": "Level 7 Orchestration", "shortDescription": "Route, remember, audit, and execute",
			"longDescription": "Local-first, default-off orchestration with provider discovery, explainable routing, private codebase memory, isolated Cyber audit, and durable Headless waves.",
			"developerName":   "Level 7 Engineering", "category": "Developer Tools",
			"capabilities":  []string{"Executable", "MCP", "Write"},
			"defaultPrompt": []string{"Use l7-onboard to inspect this project", "Use l7-sync to build private codebase memory", "Use l7-cyber for a read-only security audit"},
		}
	} else {
		value["$schema"] = "https://json.schemastore.org/claude-code-plugin-manifest.json"
		value["displayName"] = "Level 7 Orchestration"
	}
	return jsonData(value)
}

func permissions(host string) []byte {
	return jsonData(map[string]any{
		"schema": 1, "version": candidateVersion, "host": host,
		"bundled_executables": []string{"bin/l7", "bin/darwin-arm64/l7", "bin/darwin-arm64/l7-embed", "bin/darwin-amd64/l7", "bin/darwin-amd64/l7-embed"},
		"local_mcp":           true, "telemetry": false, "host_settings": false,
		"network":     map[string]any{"default": "off", "implicit": "configured gateway endpoint only", "other": "requires explicit policy; Cyber active container has no Internet"},
		"credentials": "environment-variable or macOS Keychain references only; secret values are never persisted",
		"workspace":   "repository scope plus disposable Git worktrees; private state remains under the Git common directory",
	})
}

func provenance(host string) []byte {
	return jsonData(map[string]any{
		"schema": 1, "version": candidateVersion, "host": host, "signed": false,
		"channel": "development-candidate", "release_blocked": true,
		"next": "obtain protected release identities, exact-candidate verification, independent read-only audit, and named owner GO before signing v1.0.0",
	})
}

func sbom(host string, checks []checksum) []byte {
	files := make([]any, 0, len(checks))
	relationships := []any{
		map[string]any{"spdxElementId": "SPDXRef-DOCUMENT", "relationshipType": "DESCRIBES", "relatedSpdxElement": "SPDXRef-Package-Level7"},
		map[string]any{"spdxElementId": "SPDXRef-Package-Level7", "relationshipType": "DEPENDS_ON", "relatedSpdxElement": "SPDXRef-Package-gotreesitter"},
	}
	for index, current := range checks {
		id := fmt.Sprintf("SPDXRef-File-%04d", index+1)
		files = append(files, map[string]any{
			"fileName": "./" + current.Path, "SPDXID": id,
			"checksums":        []any{map[string]any{"algorithm": "SHA256", "checksumValue": current.SHA256}},
			"licenseConcluded": "NOASSERTION", "copyrightText": "NOASSERTION",
		})
		relationships = append(relationships, map[string]any{"spdxElementId": "SPDXRef-Package-Level7", "relationshipType": "CONTAINS", "relatedSpdxElement": id})
	}
	return jsonData(map[string]any{
		"spdxVersion": "SPDX-2.3", "dataLicense": "CC0-1.0", "SPDXID": "SPDXRef-DOCUMENT",
		"name":              "level7-dev-loop-" + candidateVersion + "-" + host,
		"documentNamespace": "https://github.com/addressanup/level7-dev-loop/sbom/" + candidateVersion + "/" + host,
		"creationInfo":      map[string]any{"created": "1970-01-01T00:00:00Z", "creators": []string{"Tool: l7pack-1"}},
		"packages": []any{
			map[string]any{"name": "level7-dev-loop", "SPDXID": "SPDXRef-Package-Level7", "versionInfo": candidateVersion, "downloadLocation": "NOASSERTION", "filesAnalyzed": false, "licenseConcluded": "MIT", "licenseDeclared": "MIT", "copyrightText": "NOASSERTION"},
			map[string]any{"name": "github.com/odvcencio/gotreesitter", "SPDXID": "SPDXRef-Package-gotreesitter", "versionInfo": "v0.24.0", "downloadLocation": "https://github.com/odvcencio/gotreesitter", "filesAnalyzed": false, "licenseConcluded": "MIT", "licenseDeclared": "MIT", "copyrightText": "NOASSERTION", "externalRefs": []any{map[string]any{"referenceCategory": "PACKAGE-MANAGER", "referenceType": "purl", "referenceLocator": "pkg:golang/github.com/odvcencio/gotreesitter@v0.24.0"}}},
		},
		"files": files, "relationships": relationships,
	})
}

func archive(files []file) ([]byte, error) {
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, current := range files {
		header := &zip.FileHeader{Name: current.name, Method: zip.Deflate}
		header.SetMode(current.mode)
		header.SetModTime(time.Unix(0, 0).UTC())
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(current.data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func validateMachO(name, arch string) error {
	value, err := macho.Open(name)
	if err != nil {
		return err
	}
	defer value.Close()
	want := macho.CpuArm64
	if arch == "amd64" {
		want = macho.CpuAmd64
	}
	if value.Cpu != want {
		return errors.New("Mach-O architecture mismatch")
	}
	return nil
}

func boundedRead(name string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() < 1 || info.Size() > maximum {
		return nil, errors.New("input is not a bounded regular file")
	}
	return os.ReadFile(name)
}

func jsonData(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}

const launcher = `#!/bin/sh
set -eu
script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
case $(uname -m) in
  arm64|aarch64) platform=darwin-arm64 ;;
  x86_64|amd64) platform=darwin-amd64 ;;
  *) printf 'Level 7 v1 supports macOS arm64 and amd64 only.\n' >&2; exit 1 ;;
esac
platform_dir="$script_dir/$platform"
PATH="$platform_dir:$PATH"
export PATH
exec "$platform_dir/l7" "$@"
`
