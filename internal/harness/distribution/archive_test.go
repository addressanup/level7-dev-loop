package main

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestArchiveIsCanonicalAndDeterministic(t *testing.T) {
	entries := []archiveEntry{
		{Name: "skills/z/SKILL.md", Data: []byte("z\n"), Mode: 0o644},
		{Name: ".codex-plugin/plugin.json", Data: []byte("{}\n"), Mode: 0o644},
		{Name: "LICENSE", Data: []byte("license\n"), Mode: 0o644},
	}
	first, err := createArchive(entries)
	if err != nil {
		t.Fatal(err)
	}
	second, err := createArchive(entries)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("archive changed error=%v", err)
	}
	if err := validateArchive(first, entries); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".codex-plugin/plugin.json", "LICENSE", "skills/z/SKILL.md"}
	for index, file := range reader.File {
		if file.Name != want[index] || file.Method != zip.Store || file.Mode().Perm() != 0o644 || !file.Modified.Equal(archiveTimestamp) {
			t.Fatalf("entry %d=%+v", index, file.FileHeader)
		}
	}
}

func TestArchiveRejectsUnsafeEntries(t *testing.T) {
	for _, name := range []string{"", "/absolute", "../escape", "a/../b", "a\\b", "a//b", "directory/", "line\nbreak", "."} {
		t.Run(strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			if _, err := createArchive([]archiveEntry{{Name: name, Data: []byte("x\n"), Mode: 0o644}}); err == nil {
				t.Fatalf("unsafe name %q passed", name)
			}
		})
	}
	if _, err := createArchive([]archiveEntry{
		{Name: "same", Data: []byte("a\n"), Mode: 0o644},
		{Name: "same", Data: []byte("b\n"), Mode: 0o644},
	}); err == nil {
		t.Fatal("duplicate archive path passed")
	}
	for _, entry := range []archiveEntry{
		{Name: "empty", Mode: 0o644},
		{Name: "executable", Data: []byte("x\n"), Mode: 0o755},
		{Name: "nul", Data: []byte{'x', 0}, Mode: 0o644},
	} {
		if _, err := createArchive([]archiveEntry{entry}); err == nil {
			t.Fatalf("unsafe entry passed: %+v", entry)
		}
	}
}

func TestArchiveValidationRejectsMutationAndMetadata(t *testing.T) {
	entry := archiveEntry{Name: "file.txt", Data: []byte("expected\n"), Mode: 0o644}
	archive, err := createArchive([]archiveEntry{entry})
	if err != nil {
		t.Fatal(err)
	}
	wrong := entry
	wrong.Data = []byte("different\n")
	if err := validateArchive(archive, []archiveEntry{wrong}); err == nil {
		t.Fatal("content substitution passed")
	}
	if err := validateArchive(archive, []archiveEntry{entry, {Name: "missing", Data: []byte("x\n"), Mode: 0o644}}); err == nil {
		t.Fatal("missing allowlisted file passed")
	}

	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	header := &zip.FileHeader{Name: "file.txt", Method: zip.Deflate}
	header.SetMode(0o644)
	header.SetModTime(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	part, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "expected\n"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := validateArchive(output.Bytes(), []archiveEntry{entry}); err == nil {
		t.Fatal("noncanonical archive metadata passed")
	}
}

func TestArchiveBounds(t *testing.T) {
	large := archiveEntry{Name: "large", Data: bytes.Repeat([]byte{'x'}, maximumFileSize+1), Mode: 0o644}
	if _, err := createArchive([]archiveEntry{large}); err == nil {
		t.Fatal("oversized entry passed")
	}
	entries := make([]archiveEntry, maximumArchiveFiles+1)
	for index := range entries {
		entries[index] = archiveEntry{Name: "f" + strings.Repeat("x", index%10) + string(rune('a'+index%26)), Data: []byte("x\n"), Mode: 0o644}
	}
	if _, err := createArchive(entries); err == nil {
		t.Fatal("oversized entry count passed")
	}
}
