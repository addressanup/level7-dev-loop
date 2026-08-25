package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"
)

const (
	maxInputBytes = 2 << 20
	maxInputLines = 10000
	maxLineBytes  = 64 << 10
)

type tsvRow map[string]string

func readStrictFile(root, relative string) ([]byte, []finding) {
	return readStrictFileWithHook(root, relative, nil)
}

func readStrictDirectory(root, relative string, entryLimit int) ([]os.DirEntry, []finding) {
	return readStrictDirectoryWithHook(root, relative, entryLimit, nil)
}

func readStrictDirectoryWithHook(root, relative string, entryLimit int, beforeRead func()) ([]os.DirEntry, []finding) {
	if !safeRelativeASCIIPath(relative) {
		return nil, []finding{newFinding("BCTL-022", relative, "input path is not a canonical rooted relative path", "restore the fixed repository-relative input")}
	}
	rooted, err := os.OpenRoot(root)
	if err != nil {
		return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
	}
	expected, pathFindings := lstatStrictDirectory(rooted, relative)
	if len(pathFindings) != 0 {
		if closeErr := rooted.Close(); closeErr != nil {
			pathFindings = appendFindings(pathFindings, newFinding("BCTL-010", relative, closeErr.Error(), "restore readable rooted input state"))
		}
		return nil, pathFindings
	}

	directory, err := rooted.OpenFile(relative, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW|syscall.O_DIRECTORY, 0)
	if err != nil {
		_ = rooted.Close()
		return nil, []finding{inputChangedFinding(relative, err.Error())}
	}
	opened, statErr := directory.Stat()
	if statErr != nil {
		_ = directory.Close()
		_ = rooted.Close()
		return nil, []finding{newFinding("BCTL-010", relative, statErr.Error(), "restore the required input and rerun the gate")}
	}
	if !opened.IsDir() || !sameStrictDirectoryState(expected, opened) {
		_ = directory.Close()
		_ = rooted.Close()
		return nil, []finding{inputChangedFinding(relative, "directory identity or metadata changed before enumeration")}
	}

	if beforeRead != nil {
		beforeRead()
	}
	entries, readErr := directory.ReadDir(entryLimit)
	final, finalStatErr := directory.Stat()
	pathFinal, pathStatErr := rooted.Lstat(relative)
	directoryCloseErr := directory.Close()
	rootCloseErr := rooted.Close()
	if readErr != nil && readErr != io.EOF {
		return nil, []finding{newFinding("BCTL-010", relative, readErr.Error(), "restore the required input and rerun the gate")}
	}
	if finalStatErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, finalStatErr.Error(), "restore the required input and rerun the gate")}
	}
	if pathStatErr != nil {
		return nil, []finding{inputChangedFinding(relative, pathStatErr.Error())}
	}
	if directoryCloseErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, directoryCloseErr.Error(), "restore readable rooted input state")}
	}
	if rootCloseErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, rootCloseErr.Error(), "restore readable rooted input state")}
	}
	if !final.IsDir() || !pathFinal.IsDir() || !sameStrictDirectoryState(expected, final) || !sameStrictDirectoryState(final, pathFinal) {
		return nil, []finding{inputChangedFinding(relative, "directory identity or metadata changed during enumeration")}
	}
	return entries, nil
}

func readStrictFileWithHook(root, relative string, beforeRead func()) ([]byte, []finding) {
	if !safeRelativeASCIIPath(relative) {
		return nil, []finding{newFinding("BCTL-022", relative, "input path is not a canonical rooted relative path", "restore the fixed repository-relative input")}
	}
	rooted, err := os.OpenRoot(root)
	if err != nil {
		return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
	}
	expected, pathFindings := lstatStrictInput(rooted, relative)
	if len(pathFindings) != 0 {
		if closeErr := rooted.Close(); closeErr != nil {
			pathFindings = appendFindings(pathFindings, newFinding("BCTL-010", relative, closeErr.Error(), "restore readable rooted input state"))
		}
		return nil, pathFindings
	}
	if expected.Size() > maxInputBytes {
		_ = rooted.Close()
		return nil, []finding{newFinding("BCTL-011", relative, "input exceeds the byte limit", "narrow the authoritative input")}
	}

	file, err := rooted.OpenFile(relative, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		_ = rooted.Close()
		return nil, []finding{inputChangedFinding(relative, err.Error())}
	}
	opened, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		_ = rooted.Close()
		return nil, []finding{newFinding("BCTL-010", relative, statErr.Error(), "restore the required input and rerun the gate")}
	}
	if shapeFindings := strictInputShapeFindings(relative, opened); len(shapeFindings) != 0 {
		_ = file.Close()
		_ = rooted.Close()
		return nil, shapeFindings
	}
	if !sameStrictInputState(expected, opened) {
		_ = file.Close()
		_ = rooted.Close()
		return nil, []finding{inputChangedFinding(relative, "input identity or metadata changed before content read")}
	}

	if beforeRead != nil {
		beforeRead()
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxInputBytes+1))
	final, finalStatErr := file.Stat()
	pathFinal, pathStatErr := rooted.Lstat(relative)
	fileCloseErr := file.Close()
	rootCloseErr := rooted.Close()
	if readErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, readErr.Error(), "restore the required input and rerun the gate")}
	}
	if finalStatErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, finalStatErr.Error(), "restore the required input and rerun the gate")}
	}
	if pathStatErr != nil {
		return nil, []finding{inputChangedFinding(relative, pathStatErr.Error())}
	}
	if fileCloseErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, fileCloseErr.Error(), "restore the required input and rerun the gate")}
	}
	if rootCloseErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, rootCloseErr.Error(), "restore readable rooted input state")}
	}
	if len(data) > maxInputBytes {
		return nil, []finding{newFinding("BCTL-011", relative, "input exceeds the byte limit", "narrow the authoritative input")}
	}
	if shapeFindings := strictInputShapeFindings(relative, final); len(shapeFindings) != 0 {
		return nil, shapeFindings
	}
	if shapeFindings := strictInputShapeFindings(relative, pathFinal); len(shapeFindings) != 0 {
		return nil, shapeFindings
	}
	if !sameStrictInputState(expected, final) || !sameStrictInputState(final, pathFinal) || final.Size() != int64(len(data)) {
		return nil, []finding{inputChangedFinding(relative, "input identity, metadata, or size changed during content read")}
	}
	return validateStrictText(relative, data)
}

func lstatStrictInput(rooted *os.Root, relative string) (os.FileInfo, []finding) {
	components := strings.Split(relative, "/")
	for index := range components {
		name := strings.Join(components[:index+1], "/")
		info, err := rooted.Lstat(name)
		if err != nil {
			return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
		}
		if index < len(components)-1 {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return nil, []finding{newFinding("BCTL-022", name, "input path contains a symlink or nondirectory component", "restore a real rooted directory path")}
			}
			continue
		}
		if findings := strictInputShapeFindings(relative, info); len(findings) != 0 {
			return nil, findings
		}
		return info, nil
	}
	return nil, []finding{newFinding("BCTL-022", relative, "input path has no component", "restore the fixed repository-relative input")}
}

func lstatStrictDirectory(rooted *os.Root, relative string) (os.FileInfo, []finding) {
	components := strings.Split(relative, "/")
	for index := range components {
		name := strings.Join(components[:index+1], "/")
		info, err := rooted.Lstat(name)
		if err != nil {
			return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, []finding{newFinding("BCTL-022", name, "input path contains a symlink or nondirectory component", "restore a real rooted directory path")}
		}
	}
	info, err := rooted.Lstat(relative)
	if err != nil {
		return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
	}
	return info, nil
}

func strictInputShapeFindings(relative string, info os.FileInfo) []finding {
	links := uint64(0)
	linkCountKnown := false
	if info.Mode().IsRegular() {
		links, linkCountKnown = regularFileLinkCount(info)
	}
	return validateFileShape(relative, info.Mode(), links, linkCountKnown)
}

func sameStrictInputState(left, right os.FileInfo) bool {
	leftLinks, leftKnown := regularFileLinkCount(left)
	rightLinks, rightKnown := regularFileLinkCount(right)
	return os.SameFile(left, right) &&
		left.Mode() == right.Mode() &&
		left.Size() == right.Size() &&
		left.ModTime().Equal(right.ModTime()) &&
		leftKnown && rightKnown && leftLinks == 1 && rightLinks == 1
}

func sameStrictDirectoryState(left, right os.FileInfo) bool {
	return os.SameFile(left, right) &&
		left.IsDir() && right.IsDir() &&
		left.Mode() == right.Mode() &&
		left.Size() == right.Size() &&
		left.ModTime().Equal(right.ModTime())
}

func inputChangedFinding(relative, detail string) finding {
	return newFinding("BCTL-023", relative, "input changed during rooted inspection: "+detail, "retry from stable rooted input state")
}

func validateStrictText(name string, data []byte) ([]byte, []finding) {
	var findings []finding
	if len(data) > maxInputBytes {
		return nil, []finding{newFinding("BCTL-011", name, "input exceeds the byte limit", "narrow the authoritative input")}
	}
	if !utf8.Valid(data) {
		findings = appendFindings(findings, newFinding("BCTL-012", name, "input is not valid UTF-8", "encode the input as UTF-8"))
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		findings = appendFindings(findings, newFinding("BCTL-013", name, "input must end with one newline", "add the required final newline"))
	}
	if bytes.Contains(data, []byte{'\r'}) {
		findings = appendFindings(findings, newFinding("BCTL-014", name, "carriage returns are forbidden", "normalize the input to LF"))
	}
	lineNumber := 1
	lineBytes := 0
	lineLimitReported := false
	for _, value := range data {
		if value == '\n' {
			lineNumber++
			lineBytes = 0
			if lineNumber-1 > maxInputLines && !lineLimitReported {
				findings = appendFindings(findings, newFinding("BCTL-015", name, "input exceeds the line limit", "split or narrow the authoritative input"))
				lineLimitReported = true
			}
			continue
		}
		lineBytes++
		if lineBytes == maxLineBytes+1 {
			findings = appendFindings(findings, newFinding("BCTL-016", fmt.Sprintf("%s:%d", name, lineNumber), "line exceeds the byte limit", "shorten the line"))
		}
	}
	return data, findings
}

func parseTSV(name string, data []byte, expectedHeader []string) ([]tsvRow, []finding) {
	_, findings := validateStrictText(name, data)
	if len(findings) != 0 {
		return nil, findings
	}

	lines := strings.Split(string(data), "\n")
	headerSeen := false
	var rows []tsvRow
	for lineIndex, line := range lines[:len(lines)-1] {
		if !headerSeen && strings.HasPrefix(line, "#") {
			commentHeader := strings.TrimPrefix(line, "# ")
			if commentHeader != strings.Join(expectedHeader, "\t") {
				continue
			}
			line = commentHeader
		}
		if line == "" {
			findings = appendFindings(findings, newFinding("BCTL-017", fmt.Sprintf("%s:%d", name, lineIndex+1), "blank rows are forbidden", "remove the blank row"))
			continue
		}
		fields := strings.Split(line, "\t")
		if !headerSeen {
			headerSeen = true
			if len(fields) != len(expectedHeader) {
				findings = appendFindings(findings, newFinding("BCTL-018", name, "TSV header has the wrong field count", "restore the exact approved header"))
				continue
			}
			for index := range expectedHeader {
				if fields[index] != expectedHeader[index] {
					findings = appendFindings(findings, newFinding("BCTL-018", name, "TSV header differs from the approved schema", "restore the exact approved header"))
					break
				}
			}
			continue
		}
		if len(fields) != len(expectedHeader) {
			findings = appendFindings(findings, newFinding("BCTL-019", fmt.Sprintf("%s:%d", name, lineIndex+1), "TSV row has the wrong field count", "restore the exact row shape"))
			continue
		}
		row := make(tsvRow, len(fields))
		valid := true
		for index, field := range fields {
			if field == "" || strings.TrimSpace(field) != field || !isSafeField(field) {
				findings = appendFindings(findings, newFinding("BCTL-020", fmt.Sprintf("%s:%d", name, lineIndex+1), "TSV field is empty, padded, or contains a control character", "use a nonempty canonical field"))
				valid = false
				break
			}
			row[expectedHeader[index]] = field
		}
		if valid {
			rows = append(rows, row)
		}
	}
	if !headerSeen {
		findings = appendFindings(findings, newFinding("BCTL-021", name, "TSV header is missing", "restore the exact approved header"))
	}
	return rows, findings
}

func loadTSV(root, relative string, header []string) ([]tsvRow, []finding) {
	data, findings := readStrictFile(root, relative)
	if len(findings) != 0 {
		return nil, findings
	}
	return parseTSV(relative, data, header)
}

func isSafeField(value string) bool {
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func fileSHA256(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
