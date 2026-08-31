package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadArchiveRejectsTraversalSymlinkDuplicateAndCollision(t *testing.T) {
	tests := map[string][]zipFixture{
		"traversal": {{Name: "../escape", Mode: 0o644, Data: []byte("escape\n")}},
		"absolute":  {{Name: "/escape", Mode: 0o644, Data: []byte("escape\n")}},
		"backslash": {{Name: `dir\escape`, Mode: 0o644, Data: []byte("escape\n")}},
		"symlink":   {{Name: "link", Mode: os.ModeSymlink | 0o777, Data: []byte("target")}},
		"duplicate": {
			{Name: "same", Mode: 0o644, Data: []byte("one")},
			{Name: "same", Mode: 0o644, Data: []byte("two")},
		},
		"file-directory collision": {
			{Name: "parent", Mode: 0o644, Data: []byte("file")},
			{Name: "parent/child", Mode: 0o644, Data: []byte("child")},
		},
		"unsorted": {
			{Name: "z", Mode: 0o644, Data: []byte("last")},
			{Name: "a", Mode: 0o644, Data: []byte("first")},
		},
	}
	for name, entries := range tests {
		t.Run(name, func(t *testing.T) {
			archive := writeFixtureArchive(t, entries)
			if _, err := loadArchive(archive, "codex", "test"); err == nil {
				t.Fatal("unsafe archive was accepted")
			}
		})
	}
}

func TestClosedInventoryRejectsSubstitution(t *testing.T) {
	pkg := fixturePackage("codex", stableVersion, map[string]fixtureFile{
		"a.txt": {Data: []byte("a\n"), Mode: 0o644},
		"b.txt": {Data: []byte("b\n"), Mode: 0o644},
	})
	declared := make([]declaredFile, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		declared = append(declared, declaredFile{Path: file.Path, Size: len(file.Data), SHA256: file.SHA256})
	}
	if err := validateDeclaredFiles(pkg, declared, nil, false); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func([]declaredFile){
		"path":   func(files []declaredFile) { files[0].Path = "substitute.txt" },
		"digest": func(files []declaredFile) { files[0].SHA256 = strings.Repeat("0", 64) },
		"size":   func(files []declaredFile) { files[0].Size++ },
		"missing": func(files []declaredFile) {
			files[0] = files[1]
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := append([]declaredFile{}, declared...)
			mutate(changed)
			if err := validateDeclaredFiles(pkg, changed, nil, false); err == nil {
				t.Fatal("substituted inventory was accepted")
			}
		})
	}
}

func TestStrictMetadataRejectsDuplicateFieldsAndTrailingData(t *testing.T) {
	var target struct {
		Schema int `json:"schema"`
	}
	for _, data := range [][]byte{
		[]byte(`{"schema":1,"schema":1}`),
		[]byte("{\"schema\":1}\n{\"schema\":1}"),
		[]byte(`{"schema":1,"unknown":true}`),
	} {
		if err := decodeStrict(data, &target); err == nil {
			t.Fatalf("unsafe metadata accepted: %s", data)
		}
	}
}

func TestNativeExecutionProfileDeniesAllNetworkOperations(t *testing.T) {
	if !strings.Contains(networkDenyProfile, "(deny network*)") || !strings.Contains(networkDenyProfile, "(allow default)") {
		t.Fatalf("unsafe sandbox profile: %q", networkDenyProfile)
	}
	if runtime.GOOS != "darwin" {
		t.Skip("macOS sandbox enforcement is exercised by the native architecture jobs")
	}
	probe := `import socket; client=socket.socket(); client.connect(("127.0.0.1", 1))`
	output, err := exec.Command("/usr/bin/sandbox-exec", "-p", networkDenyProfile, "/usr/bin/python3", "-c", probe).CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("Operation not permitted")) {
		t.Fatalf("network denial was not enforced: err=%v output=%s", err, output)
	}
}

func TestDisposableLifecycleUpgradeRollbackAndRemoval(t *testing.T) {
	stable, candidate := lifecyclePackages()
	root := filepath.Join(t.TempDir(), "lifecycle-codex")
	if err := installInitial(root, stable); err != nil {
		t.Fatal(err)
	}
	stableSnapshot := snapshotTree(t, filepath.Join(root, "active"))
	if err := upgradeInstallation(root, stable.Digest, candidate); err != nil {
		t.Fatal(err)
	}
	if err := verifyTree(filepath.Join(root, "active"), receiptPackage(candidate)); err != nil {
		t.Fatal(err)
	}
	if err := rollbackInstallation(root, candidate.Digest, stable.Digest); err != nil {
		t.Fatal(err)
	}
	if got := snapshotTree(t, filepath.Join(root, "active")); !bytes.Equal(got, stableSnapshot) {
		t.Fatal("rollback did not restore exact stable bytes and modes")
	}
	if err := removeInstallation(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("removal left lifecycle root: %v", err)
	}
}

func TestUpgradeRejectsStaleReceiptWithoutMutation(t *testing.T) {
	stable, candidate := lifecyclePackages()
	root := filepath.Join(t.TempDir(), "lifecycle-codex")
	if err := installInitial(root, stable); err != nil {
		t.Fatal(err)
	}
	before := snapshotTree(t, filepath.Join(root, "active"))
	receipt, err := loadReceipt(root)
	if err != nil {
		t.Fatal(err)
	}
	receiptData, err := os.ReadFile(filepath.Join(root, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	receiptData = bytes.Replace(receiptData, []byte(receipt.Active.Digest), []byte(strings.Repeat("0", 64)), 1)
	if err := os.WriteFile(filepath.Join(root, "receipt.json"), receiptData, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := upgradeInstallation(root, stable.Digest, candidate); err == nil {
		t.Fatal("upgrade accepted a stale receipt")
	}
	if got := snapshotTree(t, filepath.Join(root, "active")); !bytes.Equal(got, before) {
		t.Fatal("rejected stale-receipt upgrade mutated the active tree")
	}
	if pathExists(filepath.Join(root, "previous")) || pathExists(filepath.Join(root, "stage")) {
		t.Fatal("rejected stale-receipt upgrade created transaction state")
	}
}

func TestLifecycleRejectsChangedOwnedAndUnownedFiles(t *testing.T) {
	stable, _ := lifecyclePackages()
	tests := map[string]func(string) error{
		"changed owned": func(root string) error {
			return os.WriteFile(filepath.Join(root, "active", "a.txt"), []byte("changed\n"), 0o644)
		},
		"unowned": func(root string) error {
			return os.WriteFile(filepath.Join(root, "active", "unowned.txt"), []byte("unowned\n"), 0o600)
		},
	}
	for name, alter := range tests {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "lifecycle-codex")
			if err := installInitial(root, stable); err != nil {
				t.Fatal(err)
			}
			if err := alter(root); err != nil {
				t.Fatal(err)
			}
			if err := removeInstallation(root); err == nil {
				t.Fatal("removal accepted a package-tree conflict")
			}
			if !pathExists(root) {
				t.Fatal("blocked removal changed the installation")
			}
		})
	}
}

func TestVerifyTreeRejectsSymlinkSubstitution(t *testing.T) {
	stable, _ := lifecyclePackages()
	root := filepath.Join(t.TempDir(), "lifecycle-codex")
	if err := installInitial(root, stable); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("stable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	owned := filepath.Join(root, "active", "a.txt")
	if err := os.Remove(owned); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, owned); err != nil {
		t.Fatal(err)
	}
	if err := verifyTree(filepath.Join(root, "active"), receiptPackage(stable)); err == nil {
		t.Fatal("symlink substitution was accepted")
	}
}

type zipFixture struct {
	Name string
	Mode fs.FileMode
	Data []byte
}

func writeFixtureArchive(t *testing.T, entries []zipFixture) string {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.Name, Method: zip.Store}
		header.SetMode(entry.Mode)
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(entry.Data); err != nil {
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

type fixtureFile struct {
	Data []byte
	Mode fs.FileMode
}

func fixturePackage(host, version string, files map[string]fixtureFile) archivePackage {
	paths := make([]string, 0, len(files))
	for name := range files {
		paths = append(paths, name)
	}
	sortStrings(paths)
	pkg := archivePackage{Host: host, Version: version, Files: make([]archiveFile, 0, len(paths))}
	identity := sha256.New()
	for _, name := range paths {
		file := files[name]
		sum := sha256.Sum256(file.Data)
		digest := hex.EncodeToString(sum[:])
		pkg.Files = append(pkg.Files, archiveFile{Path: name, Data: append([]byte{}, file.Data...), Mode: file.Mode, SHA256: digest})
		_, _ = fmt.Fprintf(identity, "%s\x00%04o\x00%s\x00", name, file.Mode.Perm(), digest)
	}
	pkg.Digest = hex.EncodeToString(identity.Sum(nil))
	return pkg
}

func lifecyclePackages() (archivePackage, archivePackage) {
	stable := fixturePackage("codex", stableVersion, map[string]fixtureFile{
		"a.txt":             {Data: []byte("stable\n"), Mode: 0o644},
		"skills/x/SKILL.md": {Data: []byte("stable skill\n"), Mode: 0o644},
	})
	candidate := fixturePackage("codex", candidateVersion, map[string]fixtureFile{
		"a.txt":             {Data: []byte("candidate\n"), Mode: 0o644},
		"bin/l7":            {Data: []byte("#!/bin/sh\nexit 0\n"), Mode: 0o755},
		"skills/x/SKILL.md": {Data: []byte("candidate skill\n"), Mode: 0o644},
	})
	return stable, candidate
}

func snapshotTree(t *testing.T, root string) []byte {
	t.Helper()
	var output bytes.Buffer
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(&output, "%s\x00%04o\x00%d\x00", filepath.ToSlash(relative), info.Mode().Perm(), len(data))
		_, _ = output.Write(data)
		_ = output.WriteByte(0)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func sortStrings(values []string) {
	for index := 1; index < len(values); index++ {
		for current := index; current > 0 && values[current] < values[current-1]; current-- {
			values[current], values[current-1] = values[current-1], values[current]
		}
	}
}
