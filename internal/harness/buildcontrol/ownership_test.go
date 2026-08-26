package main

import (
	"strings"
	"testing"
)

func TestCurrentOwnershipContract(t *testing.T) {
	t.Parallel()
	count, findings := checkOwnership(repositoryRoot(t))
	if len(findings) != 0 {
		t.Fatalf("ownership findings: %+v", findings)
	}
	if count != len(expectedOwnership) {
		t.Fatalf("ownership count: got %d, want %d", count, len(expectedOwnership))
	}
}

func TestOwnershipRejectsMissingDuplicateChangedAndUnknownRows(t *testing.T) {
	t.Parallel()
	var rows []tsvRow
	for control, expected := range expectedOwnership {
		if control == "readme" {
			continue
		}
		rows = append(rows, tsvRow{"control": control, "path_kind": expected.pathKind, "path": expected.path, "writer": expected.writer, "reviewer": expected.reviewer, "change_gate": expected.changeGate})
	}
	rows[0]["writer"] = "candidate-author"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"control": "unknown", "path_kind": "exact", "path": "unknown", "writer": "candidate-author", "reviewer": "self", "change_gate": "none"})
	_, findings := validateOwnershipRows(rows)
	rules := findingRules(findings)
	for _, rule := range []string{"OWN-401", "OWN-404", "OWN-405", "OWN-406"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestOwnershipRejectsOverlappingPathsAndWriterMismatch(t *testing.T) {
	t.Parallel()
	rules := map[string]ownershipExpectation{
		"prefix": {pathKind: "prefix", path: "docs/", writer: "one", reviewer: "review", changeGate: "gate"},
		"exact":  {pathKind: "exact", path: "docs/file", writer: "two", reviewer: "review", changeGate: "gate"},
	}
	if !ownershipPathsOverlap(rules["prefix"], rules["exact"]) {
		t.Fatal("overlap was not detected")
	}
	pathRows := []tsvRow{{"path": "docs/other", "owner": "wrong", "change": "add", "rule": "SCOPE-321"}}
	findings := crossCheckPathOwnership(pathRows, map[string]ownershipExpectation{"prefix": rules["prefix"]})
	if findingRules(findings)["OWN-411"] == 0 {
		t.Fatalf("writer mismatch findings: %+v", findings)
	}
}

func TestOwnershipRejectsCandidateWriterForProtectedControls(t *testing.T) {
	t.Parallel()
	var rows []tsvRow
	for control, expected := range expectedOwnership {
		writer := expected.writer
		if control == "protected-controls" {
			writer = "candidate-author"
		}
		rows = append(rows, tsvRow{"control": control, "path_kind": expected.pathKind, "path": expected.path, "writer": writer, "reviewer": expected.reviewer, "change_gate": expected.changeGate})
	}
	_, findings := validateOwnershipRows(rows)
	if rules := findingRules(findings); rules["OWN-405"] == 0 {
		t.Fatalf("candidate protected-control ownership was accepted: %+v", findings)
	}
}

func TestOwnershipIsCrossCheckedAgainstAuthoritativeOrchestrationSource(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	rows, loadFindings := loadTSV(root, "harness/control-ownership.tsv", []string{"control", "path_kind", "path", "writer", "reviewer", "change_gate"})
	if len(loadFindings) != 0 {
		t.Fatalf("load ownership: %+v", loadFindings)
	}
	rules, validationFindings := validateOwnershipRows(rows)
	if len(validationFindings) != 0 {
		t.Fatalf("validate ownership: %+v", validationFindings)
	}
	data, readFindings := readStrictFile(root, "docs/artifacts/orchestration-plan.md")
	if len(readFindings) != 0 {
		t.Fatalf("read orchestration plan: %+v", readFindings)
	}
	if findings := crossCheckOrchestrationOwnership(string(data), rules); len(findings) != 0 {
		t.Fatalf("approved source cross-check findings: %+v", findings)
	}

	changed := strings.Replace(string(data), expectedOrchestrationOwnership["codex-adapter"].owner, "candidate author", 1)
	if findings := crossCheckOrchestrationOwnership(changed, rules); findingRules(findings)["OWN-421"] == 0 {
		t.Fatalf("authoritative owner drift was accepted: %+v", findings)
	}

	incomplete := make(map[string]ownershipExpectation)
	for control, rule := range rules {
		if orchestrationClassForControl[control] != "codex-adapter" {
			incomplete[control] = rule
		}
	}
	if findings := crossCheckOrchestrationOwnership(string(data), incomplete); findingRules(findings)["OWN-424"] == 0 {
		t.Fatalf("missing source-class coverage was accepted: %+v", findings)
	}
}

func TestRequirementAndAllocationSourcesHaveDistinctOwners(t *testing.T) {
	t.Parallel()
	for control, want := range map[string]ownershipExpectation{
		"requirements-source": {"exact", "docs/artifacts/requirements.md", "requirements-owner", "owner-review", "owner+requirements-decision"},
		"release-allocation":  {"exact", "docs/artifacts/feature-backlog.md", "backlog-owner", "owner-review", "owner+impact-decision"},
	} {
		if got := expectedOwnership[control]; got != want {
			t.Fatalf("%s ownership: got %+v, want %+v", control, got, want)
		}
	}
	if len(orchestrationClassForControl) != len(expectedOwnership) {
		t.Fatalf("source mapping count: got %d, want %d", len(orchestrationClassForControl), len(expectedOwnership))
	}
}

func TestConceptBootstrapOwnershipIsExactAndSeparated(t *testing.T) {
	t.Parallel()
	for control, want := range map[string]ownershipExpectation{
		"concept-discovery": {"exact", conceptDossierPath, "concept-owner", "independent-readonly", "bounded-public-research"},
		"concept-brief":     {"exact", conceptBriefPath, "concept-owner", "owner-review", "owner+digest-approval"},
	} {
		if got := expectedOwnership[control]; got != want {
			t.Fatalf("%s ownership: got %+v, want %+v", control, got, want)
		}
		if orchestrationClassForControl[control] != "wave-records" {
			t.Fatalf("%s is not mapped to the bootstrap wave-record class", control)
		}
	}
	if ownershipPathsOverlap(expectedOwnership["concept-discovery"], expectedOwnership["concept-brief"]) {
		t.Fatal("concept dossier and brief ownership paths overlap")
	}
}

func TestWave02SemanticEvaluatorAndFutureFeatureOwnershipIsDisjoint(t *testing.T) {
	t.Parallel()
	for _, removed := range []string{"schema-source", "public-fixtures"} {
		if _, ok := expectedOwnership[removed]; ok {
			t.Fatalf("broad overlapping ownership fallback %s remains", removed)
		}
	}
	for control, want := range map[string]ownershipExpectation{
		"semantic-schema":    {"prefix", "schemas/semantic/", "semantic-owner", "independent-readonly", "owner+wave-02-design"},
		"evaluation-schema":  {"prefix", "schemas/evaluation/", "evaluator-owner", "independent-readonly", "separate-evaluator-freeze"},
		"artifact-schema":    {"prefix", "schemas/artifact/", "state-owner", "independent-readonly", "future-wave"},
		"semantic-fixtures":  {"prefix", "fixtures/public/bl-002/", "semantic-owner", "evaluator-owner", "evaluator-integration"},
		"evaluator-fixtures": {"prefix", "fixtures/public/bl-003/", "evaluator-owner", "independent-readonly", "separate-evaluator-freeze"},
		"feature-fixtures":   {"prefix", "fixtures/public/features/", "feature-owner", "evaluator-owner", "future-feature-integration"},
	} {
		if got := expectedOwnership[control]; got != want {
			t.Errorf("%s ownership: got %+v, want %+v", control, got, want)
		}
	}
	rows := []tsvRow{
		{"path": "schemas/semantic/test.json", "owner": "semantic-owner"},
		{"path": "schemas/evaluation/test.json", "owner": "evaluator-owner"},
		{"path": "fixtures/public/features/test.json", "owner": "feature-owner"},
	}
	if findings := crossCheckPathOwnership(rows, expectedOwnership); len(findings) != 0 {
		t.Fatalf("disjoint Wave 2 partitions: %+v", findings)
	}
}

func TestWave02CandidateCannotOwnEvaluatorControls(t *testing.T) {
	t.Parallel()
	var rows []tsvRow
	for control, expected := range expectedOwnership {
		writer := expected.writer
		if control == "evaluation-schema" || control == "evaluator-fixtures" || control == "public-evaluator" {
			writer = "candidate-author"
		}
		rows = append(rows, tsvRow{"control": control, "path_kind": expected.pathKind, "path": expected.path, "writer": writer, "reviewer": expected.reviewer, "change_gate": expected.changeGate})
	}
	_, findings := validateOwnershipRows(rows)
	if findingRules(findings)["OWN-405"] != 3 {
		t.Fatalf("candidate evaluator ownership findings: %+v", findings)
	}
}
