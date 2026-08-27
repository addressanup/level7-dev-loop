package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxInputBytes = 2 << 20

func decodeStrictJSON(name string, data []byte, target any) []finding {
	if len(data) == 0 || len(data) > maxInputBytes {
		return []finding{newFinding("INPUT-001", name, "structured input is empty or exceeds the size limit", "provide one bounded JSON record")}
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return []finding{newFinding("INPUT-002", name, err.Error(), "provide the exact structured input schema")}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return []finding{newFinding("INPUT-003", name, "structured input has trailing values", "retain exactly one JSON object")}
	}
	return nil
}

func readBounded(name string) ([]byte, []finding) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, []finding{newFinding("INPUT-010", name, err.Error(), "restore the required external input")}
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxInputBytes {
		return nil, []finding{newFinding("INPUT-011", name, "input is not a bounded regular file", "use a regular external authority record")}
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, []finding{newFinding("INPUT-010", name, err.Error(), "restore the required external input")}
	}
	return data, nil
}

func envInt(name string) int {
	value := os.Getenv(name)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return parsed
}

func envBool(name string) bool {
	value := strings.ToLower(os.Getenv(name))
	return value == "1" || value == "true" || value == "yes"
}

func splitScope(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func authorityPath(gitCommonDir, kind, changeID string) string {
	return filepath.Join(gitCommonDir, "l7", kind, fmt.Sprintf("%s.json", changeID))
}
