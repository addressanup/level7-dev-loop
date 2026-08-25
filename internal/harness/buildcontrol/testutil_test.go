package main

import (
	"path/filepath"
	"strings"
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

func TestFindingNormalizationAndOrderAreDeterministic(t *testing.T) {
	t.Parallel()
	findings := []finding{
		newFinding("Z-002", "b", "second", "next"),
		newFinding("A-001", "z", "first", "next"),
		newFinding("A-001", "a", "first", "next"),
	}
	sortFindings(findings)
	if findings[0].subject != "a" || findings[1].subject != "z" || findings[2].rule != "Z-002" {
		t.Fatalf("unexpected finding order: %+v", findings)
	}
	normalized := safeASCII("line\n"+strings.Repeat("x", 300), 12)
	if normalized != "line?xxxxxxx" || len(normalized) != 12 {
		t.Fatalf("unexpected bounded normalization: %q", normalized)
	}
	if maxFindings <= 0 || maxMessageSize <= 0 {
		t.Fatal("diagnostic bounds must be positive")
	}
}

func TestFindingCollectionIsBoundedBeforeAppend(t *testing.T) {
	t.Parallel()
	var findings []finding
	for index := 0; index < maxCollectedFindings+100; index++ {
		findings = appendFindings(findings, newFinding("TEST-001", "fixture", "bounded", "none"))
	}
	if len(findings) != maxCollectedFindings {
		t.Fatalf("finding collection size: got %d, want %d", len(findings), maxCollectedFindings)
	}
}
