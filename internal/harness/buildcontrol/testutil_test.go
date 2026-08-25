package main

import (
	"path/filepath"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func findingRules(findings []finding) map[string]int {
	rules := make(map[string]int)
	for _, item := range findings {
		rules[item.rule]++
	}
	return rules
}
