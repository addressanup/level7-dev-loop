package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	maximumArchiveFiles = 256
	maximumArchiveSize  = 8 << 20
)

var archiveTimestamp = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

type archiveEntry struct {
	Name string
	Data []byte
	Mode fs.FileMode
}

func createArchive(entries []archiveEntry) ([]byte, error) {
	if err := validateArchiveEntries(entries); err != nil {
		return nil, err
	}
	ordered := cloneEntries(entries)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })

	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, entry := range ordered {
		header := &zip.FileHeader{Name: entry.Name, Method: zip.Store}
		header.SetModTime(archiveTimestamp)
		header.SetMode(entry.Mode)
		part, err := writer.CreateHeader(header)
		if err != nil {
			writer.Close()
			return nil, err
		}
		if _, err := part.Write(entry.Data); err != nil {
			writer.Close()
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if output.Len() == 0 || output.Len() > maximumArchiveSize {
		return nil, errors.New("archive output exceeds its size boundary")
	}
	return output.Bytes(), nil
}

func validateArchive(data []byte, expected []archiveEntry) error {
	if len(data) == 0 || len(data) > maximumArchiveSize {
		return errors.New("archive is empty or oversized")
	}
	if err := validateArchiveEntries(expected); err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	if len(reader.File) != len(expected) || len(reader.File) > maximumArchiveFiles {
		return errors.New("archive file count does not match its allowlist")
	}
	want := make(map[string]archiveEntry, len(expected))
	for _, entry := range expected {
		want[entry.Name] = entry
	}
	seen := make(map[string]bool, len(expected))
	previous := ""
	total := 0
	for _, file := range reader.File {
		if err := validateArchiveName(file.Name); err != nil {
			return err
		}
		if seen[file.Name] || (previous != "" && file.Name <= previous) {
			return fmt.Errorf("archive contains duplicate or unsorted path %q", file.Name)
		}
		seen[file.Name] = true
		previous = file.Name
		expectedEntry, ok := want[file.Name]
		if !ok {
			return fmt.Errorf("archive contains undeclared path %q", file.Name)
		}
		if file.Method != zip.Store || !file.Mode().IsRegular() || file.Mode().Perm() != expectedEntry.Mode.Perm() ||
			!file.Modified.UTC().Equal(archiveTimestamp) || file.UncompressedSize64 != uint64(len(expectedEntry.Data)) {
			return fmt.Errorf("archive metadata mismatch for %q", file.Name)
		}
		if file.UncompressedSize64 > maximumFileSize || total+int(file.UncompressedSize64) > maximumArchiveSize {
			return fmt.Errorf("archive content exceeds bounds at %q", file.Name)
		}
		stream, err := file.Open()
		if err != nil {
			return err
		}
		content, readErr := io.ReadAll(io.LimitReader(stream, maximumFileSize+1))
		closeErr := stream.Close()
		if readErr != nil || closeErr != nil {
			return errors.Join(readErr, closeErr)
		}
		if !bytes.Equal(content, expectedEntry.Data) {
			return fmt.Errorf("archive content mismatch for %q", file.Name)
		}
		total += len(content)
	}
	return nil
}

func validateArchiveEntries(entries []archiveEntry) error {
	if len(entries) == 0 || len(entries) > maximumArchiveFiles {
		return errors.New("archive entry count is outside bounds")
	}
	seen := make(map[string]bool, len(entries))
	total := 0
	for _, entry := range entries {
		if err := validateArchiveName(entry.Name); err != nil {
			return err
		}
		if seen[entry.Name] {
			return fmt.Errorf("duplicate archive path %q", entry.Name)
		}
		seen[entry.Name] = true
		if entry.Mode != 0o644 || len(entry.Data) == 0 || len(entry.Data) > maximumFileSize || bytes.IndexByte(entry.Data, 0) >= 0 {
			return fmt.Errorf("archive entry %q has invalid mode or content", entry.Name)
		}
		total += len(entry.Data)
		if total > maximumArchiveSize {
			return errors.New("archive uncompressed content exceeds bounds")
		}
	}
	return nil
}

func validateArchiveName(name string) error {
	if name == "" || len(name) > 240 || strings.ContainsAny(name, "\\\x00\r\n") || strings.HasPrefix(name, "/") ||
		path.Clean(name) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("unsafe archive path %q", name)
	}
	for _, element := range strings.Split(name, "/") {
		if element == "" || element == "." || element == ".." {
			return fmt.Errorf("unsafe archive path element in %q", name)
		}
	}
	return nil
}

func cloneEntries(entries []archiveEntry) []archiveEntry {
	cloned := make([]archiveEntry, len(entries))
	for index, entry := range entries {
		cloned[index] = archiveEntry{Name: entry.Name, Data: append([]byte{}, entry.Data...), Mode: entry.Mode}
	}
	return cloned
}
