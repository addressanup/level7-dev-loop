package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

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
			"record_version": expected.recordVersion, "id": id, "surface": expected.surface, "current_state": expected.currentState,
			"v1_ceiling": expected.v1Ceiling, "claim_state": expected.claimState, "owner": expected.owner,
		})
	}
	rows[0]["claim_state"] = "supported"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"record_version": "wave-01-v1", "id": "unapproved", "surface": "unknown", "current_state": "planned", "v1_ceiling": "A5", "claim_state": "supported", "owner": "L7-BL-001"})
	rules := findingRules(validateSupportRows(rows))
	for _, rule := range []string{"CLAIM-201", "CLAIM-202", "CLAIM-203"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestSupportMatrixRejectsEveryRequiredBoundaryDrift(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"workspace-boundary", "plugin-authority", "development-evidence", "release-blocking-proof", "priority-p0", "priority-p1", "priority-p2", "priority-change-control"} {
		id := id
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			var rows []tsvRow
			for rowID, expected := range expectedSupport {
				row := tsvRow{
					"record_version": expected.recordVersion, "id": rowID, "surface": expected.surface, "current_state": expected.currentState,
					"v1_ceiling": expected.v1Ceiling, "claim_state": expected.claimState, "owner": expected.owner,
				}
				if rowID == id {
					row["claim_state"] = "promoted"
				}
				rows = append(rows, row)
			}
			if findings := validateSupportRows(rows); findingRules(findings)["CLAIM-203"] == 0 {
				t.Fatalf("%s drift was accepted: %+v", id, findings)
			}
		})
	}
}

func TestPermanentFalseClaimMatrix(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name  string
		id    string
		field string
		value string
	}{
		{"plugin-implies-mutation", "plugin-authority", "v1_ceiling", "mutation-authority"},
		{"codex-inherits-claude", "codex-advisory", "current_state", "promoted-by-Claude-evidence"},
		{"claude-inherits-codex", "claude-advisory", "current_state", "promoted-by-Codex-evidence"},
		{"a3-execution", "a3-a4-handoff", "v1_ceiling", "A3-execution"},
		{"a4-execution", "a3-a4-handoff", "v1_ceiling", "A4-execution"},
		{"a5-execution", "a5-autonomy", "v1_ceiling", "A5-execution"},
		{"stable-promotion", "stable-version-1.0", "claim_state", "supported"},
		{"dual-host-promotion", "dual-host-support", "claim_state", "supported"},
		{"enforcement-promotion", "enforcement-claim", "claim_state", "enforced"},
		{"generic-for-specialist", "proof-generic", "claim_state", "substitutes-specialist"},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var rows []tsvRow
			for id, expected := range expectedSupport {
				row := tsvRow{
					"record_version": expected.recordVersion, "id": id, "surface": expected.surface, "current_state": expected.currentState,
					"v1_ceiling": expected.v1Ceiling, "claim_state": expected.claimState, "owner": expected.owner,
				}
				if id == testCase.id {
					row[testCase.field] = testCase.value
				}
				rows = append(rows, row)
			}
			if findings := validateSupportRows(rows); findingRules(findings)["CLAIM-203"] == 0 {
				t.Fatalf("false claim %s was accepted: %+v", testCase.name, findings)
			}
		})
	}
}

func TestPriorityContractIsSourceBound(t *testing.T) {
	t.Parallel()
	data, findings := readStrictFile(repositoryRoot(t), "docs/artifacts/feature-backlog.md")
	if len(findings) != 0 {
		t.Fatalf("read backlog: %+v", findings)
	}
	if findings := validatePriorityContract(string(data)); len(findings) != 0 {
		t.Fatalf("approved priority contract findings: %+v", findings)
	}
	for priority := range expectedPriority {
		changed := strings.Replace(string(data), "| `"+priority+"` |", "| `PX` |", 1)
		if findings := validatePriorityContract(changed); findingRules(findings)["CLAIM-205"] == 0 {
			t.Fatalf("%s priority drift was accepted: %+v", priority, findings)
		}
	}
	changed := strings.Replace(string(data), requiredPriorityChangeRules[0], "priority changes are unrestricted", 1)
	if findings := validatePriorityContract(changed); findingRules(findings)["CLAIM-206"] == 0 {
		t.Fatalf("priority approval drift was accepted: %+v", findings)
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

func TestPrototypeInventoryRejectsMissingDuplicateAndUnknownSkills(t *testing.T) {
	t.Parallel()
	valid := []string{
		"l7-build", "l7-change", "l7-constitution", "l7-deploy", "l7-experience", "l7-geometry",
		"l7-greenfield", "l7-next", "l7-ops", "l7-release", "l7-review", "l7-storybook",
	}
	var rows []tsvRow
	for skill, expected := range expectedDispositions {
		rows = append(rows, tsvRow{"skill": skill, "disposition": expected.disposition, "target_owner": expected.targetOwner, "cutover": expected.cutover})
	}
	duplicate := append(append([]string(nil), valid...), "l7-build")
	if findings := validateDispositionRows(rows, duplicate); findingRules(findings)["CLAIM-226"] == 0 {
		t.Fatalf("duplicate prototype skill was accepted: %+v", findings)
	}
	missing := append([]string(nil), valid[:len(valid)-1]...)
	if findings := validateDispositionRows(rows, missing); findingRules(findings)["CLAIM-221"] == 0 {
		t.Fatalf("missing prototype skill was accepted: %+v", findings)
	}
	unknown := append([]string(nil), valid...)
	unknown[len(unknown)-1] = "l7-unknown"
	if findings := validateDispositionRows(rows, unknown); findingRules(findings)["CLAIM-224"] == 0 {
		t.Fatalf("unknown prototype skill was accepted: %+v", findings)
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

func TestStrictInputBoundsAreEnforcedBeforeParsing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, findings := readStrictFile(root, "missing"); findingRules(findings)["BCTL-010"] == 0 {
		t.Fatalf("missing required input was accepted: %+v", findings)
	}
	if err := os.WriteFile(filepath.Join(root, "oversize"), bytes.Repeat([]byte{'x'}, maxInputBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	data, findings := readStrictFile(root, "oversize")
	if data != nil || findingRules(findings)["BCTL-011"] == 0 {
		t.Fatalf("oversize input was not rejected before parsing: data=%d findings=%+v", len(data), findings)
	}

	atLimit := strings.Repeat("x\n", maxInputLines)
	if _, findings := validateStrictText("at-limit", []byte(atLimit)); findingRules(findings)["BCTL-015"] != 0 {
		t.Fatalf("line-count boundary was rejected: %+v", findings)
	}
	overLimit := atLimit + "x\n"
	if _, findings := validateStrictText("over-limit", []byte(overLimit)); findingRules(findings)["BCTL-015"] == 0 {
		t.Fatalf("line-count overflow was accepted: %+v", findings)
	}
}

func TestStrictInputRejectsSymlinkFIFOAndHardlinkBeforeRead(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target"), []byte("target\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(root, "symlink")); err != nil {
		t.Fatal(err)
	}
	if data, findings := readStrictFile(root, "symlink"); data != nil || findingRules(findings)["SCOPE-343"] == 0 {
		t.Fatalf("symlink input was consumed: data=%q findings=%+v", data, findings)
	}

	if err := os.Mkdir(filepath.Join(root, "real-directory"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "real-directory", "input"), []byte("inside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real-directory", filepath.Join(root, "directory-link")); err != nil {
		t.Fatal(err)
	}
	if data, findings := readStrictFile(root, "directory-link/input"); data != nil || findingRules(findings)["BCTL-022"] == 0 {
		t.Fatalf("intermediate symlink input was consumed: data=%q findings=%+v", data, findings)
	}

	if err := os.Link(filepath.Join(root, "target"), filepath.Join(root, "hardlink")); err != nil {
		t.Fatal(err)
	}
	if data, findings := readStrictFile(root, "target"); data != nil || findingRules(findings)["SCOPE-344"] == 0 {
		t.Fatalf("hardlinked input was consumed: data=%q findings=%+v", data, findings)
	}

	fifo := filepath.Join(root, "fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	finished := make(chan []finding, 1)
	go func() {
		_, findings := readStrictFile(root, "fifo")
		finished <- findings
	}()
	select {
	case findings := <-finished:
		if findingRules(findings)["SCOPE-343"] == 0 {
			t.Fatalf("FIFO did not fail for the stable shape rule: %+v", findings)
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO input blocked instead of failing before open")
	}
}

func TestStrictInputRejectsChangeDuringRead(t *testing.T) {
	root := t.TempDir()
	name := filepath.Join(root, "changing")
	if err := os.WriteFile(name, []byte("before\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, findings := readStrictFileWithHook(root, "changing", func() {
		if err := os.WriteFile(name, []byte("changed-and-longer\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	})
	if data != nil || findingRules(findings)["BCTL-023"] == 0 {
		t.Fatalf("changing input was accepted: data=%q findings=%+v", data, findings)
	}
}

func TestSkillInventoryRejectsEntryOverflowBeforeReadingSkills(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	skills := filepath.Join(root, "skills")
	if err := os.Mkdir(skills, 0o700); err != nil {
		t.Fatal(err)
	}
	for index := 0; index <= maxSkillEntries; index++ {
		if err := os.Mkdir(filepath.Join(skills, fmt.Sprintf("skill-%03d", index)), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if _, findings := loadSkillInventory(root); findingRules(findings)["CLAIM-213"] == 0 {
		t.Fatalf("skill-entry overflow was accepted: %+v", findings)
	}
}
