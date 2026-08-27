package localfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicCreateAndReplaceExposeOnlyCompleteFiles(t *testing.T) {
	directory := physicalTempDir(t)
	path := filepath.Join(directory, "state.json")
	if err := AtomicCreate(path, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := AtomicCreate(path, []byte("collision\n"), 0o600); !errors.Is(err, os.ErrExist) {
		t.Fatalf("second AtomicCreate() error=%v, want os.ErrExist", err)
	}
	if err := AtomicReplace(path, []byte("second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := Read(path, 64)
	if err != nil || string(data) != "second\n" {
		t.Fatalf("Read()=(%q,%v), want second", data, err)
	}
	matches, err := filepath.Glob(filepath.Join(directory, ".l7-write-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("atomic temporary residue=%v error=%v", matches, err)
	}
}

func TestReadRejectsUnsafeOrOversizedFiles(t *testing.T) {
	directory := physicalTempDir(t)
	regular := filepath.Join(directory, "regular")
	if err := os.WriteFile(regular, []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(regular, 4); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("oversized Read() error=%v", err)
	}
	link := filepath.Join(directory, "link")
	if err := os.Symlink(regular, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(link, 64); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("symlink Read() error=%v", err)
	}
	if _, err := Read(directory, 64); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory Read() error=%v", err)
	}
}

func TestEnsureDirectoryRejectsSymlinkedComponents(t *testing.T) {
	root := physicalTempDir(t)
	realDirectory := filepath.Join(root, "real")
	if err := os.Mkdir(realDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	linkedDirectory := filepath.Join(root, "linked")
	if err := os.Symlink(realDirectory, linkedDirectory); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDirectory(filepath.Join(linkedDirectory, "nested"), 0o700); err == nil || !strings.Contains(err.Error(), "unsafe directory") {
		t.Fatalf("EnsureDirectory() error=%v", err)
	}
	created := filepath.Join(root, "safe", "nested")
	if err := EnsureDirectory(created, 0o700); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(created); err != nil || !info.IsDir() {
		t.Fatalf("created directory info=%v error=%v", info, err)
	}
}

func TestDecodeJSONIsStrictAtEveryObjectDepth(t *testing.T) {
	type nested struct {
		Enabled bool `json:"enabled"`
	}
	type document struct {
		Schema int    `json:"schema"`
		Nested nested `json:"nested"`
	}
	tests := []struct {
		name string
		data string
	}{
		{"duplicate top-level", `{"schema":1,"schema":1,"nested":{"enabled":true}}`},
		{"duplicate nested", `{"schema":1,"nested":{"enabled":true,"enabled":false}}`},
		{"unknown", `{"schema":1,"nested":{"enabled":true},"extra":1}`},
		{"non-canonical case", `{"Schema":1,"nested":{"enabled":true}}`},
		{"trailing", `{"schema":1,"nested":{"enabled":true}} {}`},
		{"incomplete", `{"schema":1,"nested":`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result document
			if err := DecodeJSON([]byte(test.data), &result); err == nil {
				t.Fatalf("DecodeJSON(%s) unexpectedly passed", test.data)
			}
		})
	}
	var valid document
	if err := DecodeJSON([]byte(`{"schema":1,"nested":{"enabled":true}}`), &valid); err != nil || valid.Schema != 1 || !valid.Nested.Enabled {
		t.Fatalf("valid DecodeJSON()=%+v error=%v", valid, err)
	}
}

func TestRepositoryLockIsExclusiveAndCrashReleaseIsKernelOwned(t *testing.T) {
	directory := physicalTempDir(t)
	path := filepath.Join(directory, "lock")
	first, err := AcquireLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireLock(path); err == nil || !strings.Contains(err.Error(), "mutation is active") {
		t.Fatalf("contending AcquireLock() error=%v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("lock remained deadlocked after release: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryLockRejectsSymlinkSubstitution(t *testing.T) {
	directory := physicalTempDir(t)
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(target, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(directory, "lock")
	if err := os.Symlink(target, lockPath); err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireLock(lockPath); err == nil {
		t.Fatalf("symlinked AcquireLock() error=%v", err)
	}
}

func TestAnchoredDirectoryDetectsPathReplacement(t *testing.T) {
	parent := physicalTempDir(t)
	directory := filepath.Join(parent, "state")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	root, info, err := openAnchoredDirectory(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := os.Rename(directory, filepath.Join(parent, "moved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := revalidateAnchoredDirectory(root, directory, info); err == nil || !strings.Contains(err.Error(), "identity changed") {
		t.Fatalf("revalidateAnchoredDirectory() error=%v", err)
	}
}

func TestEncodeJSONIsDeterministicAndNewlineTerminated(t *testing.T) {
	data, err := EncodeJSON(struct {
		Schema int `json:"schema"`
	}{Schema: 1})
	if err != nil || string(data) != "{\n  \"schema\": 1\n}\n" {
		t.Fatalf("EncodeJSON()=%q error=%v", data, err)
	}
}

func physicalTempDir(t *testing.T) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}
