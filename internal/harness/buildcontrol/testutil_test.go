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

func TestSuccessOutputBindsVersionAndExactSourceDigests(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	first, findings := loadSuccessSourceDigests(root)
	if len(findings) != 0 {
		t.Fatalf("source digest findings: %+v", findings)
	}
	second, findings := loadSuccessSourceDigests(root)
	if len(findings) != 0 || second != first {
		t.Fatalf("source digests are not deterministic: first=%q second=%q findings=%+v", first, second, findings)
	}
	for _, source := range successSources {
		data, readFindings := readStrictFile(root, source.path)
		if len(readFindings) != 0 {
			t.Fatalf("read %s: %+v", source.path, readFindings)
		}
		want := source.id + ":" + fileSHA256(data)
		if !strings.Contains(first, want) {
			t.Fatalf("source digest %q is absent from %q", want, first)
		}
	}
	line := formatSuccess(
		traceResult{total: 163, allocations: map[string]int{"V1.0": 140, "V1.x": 18, "Later": 5}},
		12,
		policyResult{phase: "wave-01", files: 100, changed: 36},
		42,
		first,
	)
	for _, required := range []string{"gate_version=" + buildControlVersion, "source_sha256=" + first, "phase=wave-01", "requirements=163"} {
		if !strings.Contains(line, required) {
			t.Fatalf("success output %q does not contain %q", line, required)
		}
	}
}
