package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestArchiveIsDeterministicAndPreservesExecutableMode(t *testing.T) {
	files := []file{{name: "bin/l7", data: []byte(launcher), mode: 0o755}, {name: "PERMISSIONS.json", data: permissions("codex"), mode: 0o644}}
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
	var permission map[string]any
	if err := json.Unmarshal(permissions("claude"), &permission); err != nil {
		t.Fatal(err)
	}
	network := permission["network"].(map[string]any)
	if network["default"] != "off" || permission["telemetry"] != false || permission["local_mcp"] != true {
		t.Fatalf("permissions=%v", permission)
	}
	var attestation map[string]any
	if err := json.Unmarshal(provenance("claude"), &attestation); err != nil || attestation["signed"] != false || attestation["release_blocked"] != true {
		t.Fatalf("attestation=%v err=%v", attestation, err)
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
	data := sbom("codex", []checksum{{Path: "bin/l7", SHA256: strings.Repeat("a", 64), Size: 1}})
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	packages, _ := document["packages"].([]any)
	files, _ := document["files"].([]any)
	relationships, _ := document["relationships"].([]any)
	if len(packages) != 2 || len(files) != 1 || len(relationships) != 3 || !bytes.Contains(data, []byte("gotreesitter@v0.24.0")) {
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
