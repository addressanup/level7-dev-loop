package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxInputBytes = 2 << 20
	maxInputLines = 10000
	maxLineBytes  = 64 << 10
)

type tsvRow map[string]string

func readStrictFile(root, relative string) ([]byte, []finding) {
	fullPath := filepath.Join(root, filepath.FromSlash(relative))
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, []finding{newFinding("BCTL-010", relative, err.Error(), "restore the required input and rerun the gate")}
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxInputBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, readErr.Error(), "restore the required input and rerun the gate")}
	}
	if closeErr != nil {
		return nil, []finding{newFinding("BCTL-010", relative, closeErr.Error(), "restore the required input and rerun the gate")}
	}
	if len(data) > maxInputBytes {
		return nil, []finding{newFinding("BCTL-011", relative, "input exceeds the byte limit", "narrow the authoritative input")}
	}
	return validateStrictText(relative, data)
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
