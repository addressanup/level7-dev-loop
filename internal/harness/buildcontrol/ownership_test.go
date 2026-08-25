package main

import "testing"

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
