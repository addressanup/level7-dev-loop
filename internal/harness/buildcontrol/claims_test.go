package main

import "testing"

func TestCurrentClaimContracts(t *testing.T) {
	t.Parallel()
	count, findings := checkClaims(repositoryRoot(t))
	if len(findings) != 0 {
		t.Fatalf("claim findings: %+v", findings)
	}
	if count != 12 {
		t.Fatalf("prototype count: got %d, want 12", count)
	}
}

func TestSupportMatrixRejectsFalseClaimsAndInventoryDrift(t *testing.T) {
	t.Parallel()
	rows := make([]tsvRow, 0, len(expectedSupport))
	for id, expected := range expectedSupport {
		rows = append(rows, tsvRow{
			"id": id, "surface": expected.surface, "current_state": expected.currentState,
			"v1_ceiling": expected.v1Ceiling, "claim_state": expected.claimState, "owner": expected.owner,
		})
	}
	rows[0]["claim_state"] = "supported"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"id": "unapproved", "surface": "unknown", "current_state": "planned", "v1_ceiling": "A5", "claim_state": "supported", "owner": "L7-BL-001"})
	rules := findingRules(validateSupportRows(rows))
	for _, rule := range []string{"CLAIM-201", "CLAIM-202", "CLAIM-203"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestPrototypeDispositionsRejectMissingDuplicateUnknownAndChangedRows(t *testing.T) {
	t.Parallel()
	inventory := []string{
		"l7-build", "l7-change", "l7-constitution", "l7-deploy", "l7-experience", "l7-geometry",
		"l7-greenfield", "l7-next", "l7-ops", "l7-release", "l7-review", "l7-storybook",
	}
	var rows []tsvRow
	for skill, expected := range expectedDispositions {
		if skill == "l7-next" {
			continue
		}
		rows = append(rows, tsvRow{"skill": skill, "disposition": expected.disposition, "target_owner": expected.targetOwner, "cutover": expected.cutover})
	}
	rows[0]["disposition"] = "conform"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"skill": "l7-unknown", "disposition": "exclude", "target_owner": "L7-BL-001", "cutover": "never"})
	rules := findingRules(validateDispositionRows(rows, inventory))
	for _, rule := range []string{"CLAIM-220", "CLAIM-221", "CLAIM-223", "CLAIM-224"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestStrictTSVRejectsHeaderBlankAndPaddedFields(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		data string
		rule string
	}{
		{"header", "wrong\tvalue\na\tb\n", "BCTL-018"},
		{"blank", "id\tvalue\na\tb\n\nc\td\n", "BCTL-017"},
		{"padded", "id\tvalue\na\t b\n", "BCTL-020"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, findings := parseTSV("fixture.tsv", []byte(testCase.data), []string{"id", "value"})
			if findingRules(findings)[testCase.rule] == 0 {
				t.Fatalf("findings %+v do not contain %s", findings, testCase.rule)
			}
		})
	}
}

func TestStrictTSVAcceptsExactCommentHeader(t *testing.T) {
	t.Parallel()
	rows, findings := parseTSV("fixture.tsv", []byte("# id\tvalue\na\tb\n"), []string{"id", "value"})
	if len(findings) != 0 || len(rows) != 1 || rows[0]["id"] != "a" || rows[0]["value"] != "b" {
		t.Fatalf("comment-header parse rows=%+v findings=%+v", rows, findings)
	}
}
