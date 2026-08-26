package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFoundationAdmissionExactPolicyAndImmutableBase(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	expectations, rows, findings := loadFoundationPathPolicy(root)
	if len(findings) != 0 {
		t.Fatalf("foundation path policy findings: %+v", findings)
	}
	if len(rows) != expectedFoundationPathRows || len(expectations) != expectedFoundationPathRows {
		t.Fatalf("foundation policy rows=%d expectations=%d want=%d", len(rows), len(expectations), expectedFoundationPathRows)
	}
	counts := map[string]int{}
	for _, expectation := range expectations {
		counts[expectation.change]++
	}
	if counts["add"] != 56 || counts["modify"] != 13 {
		t.Fatalf("foundation change classes: %+v", counts)
	}

	data, readFindings := readStrictFile(root, "harness/foundation-rebaseline-base.sha256")
	if len(readFindings) != 0 || fileSHA256(data) != foundationBaseManifestSHA256 {
		t.Fatalf("foundation base findings=%+v digest=%s", readFindings, fileSHA256(data))
	}
	base, parseFindings := parseSHA256Manifest("harness/foundation-rebaseline-base.sha256", data, true)
	if len(parseFindings) != 0 || len(base) != expectedFoundationBaseFiles {
		t.Fatalf("foundation base rows=%d findings=%+v", len(base), parseFindings)
	}
	for _, required := range []string{conceptBriefPath, conceptDossierPath, wave02CandidateManifest, "skills/l7-greenfield/SKILL.md"} {
		if base[required] == "" {
			t.Errorf("foundation base lacks %s", required)
		}
	}
}

func TestFoundationGate2BundleApprovalHistoryAndPredecessorsAreBound(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	if findings := checkFoundationCandidateManifest(root); len(findings) != 0 {
		t.Fatalf("candidate manifest findings: %+v", findings)
	}
	if findings := checkFoundationApproval(root); len(findings) != 0 {
		t.Fatalf("approval findings: %+v", findings)
	}
	if findings := checkFoundationHistory(root); len(findings) != 0 {
		t.Fatalf("history findings: %+v", findings)
	}
	if findings := checkFoundationGateRegistry(root); len(findings) != 0 {
		t.Fatalf("gate registry findings: %+v", findings)
	}
	baseData, readFindings := readStrictFile(root, "harness/foundation-rebaseline-base.sha256")
	if len(readFindings) != 0 {
		t.Fatalf("base read findings: %+v", readFindings)
	}
	base, parseFindings := parseSHA256Manifest("harness/foundation-rebaseline-base.sha256", baseData, true)
	if len(parseFindings) != 0 {
		t.Fatalf("base parse findings: %+v", parseFindings)
	}
	current, scanFindings := scanRepository(root)
	if len(scanFindings) != 0 {
		t.Fatalf("repository scan findings: %+v", scanFindings)
	}
	if findings := checkFoundationBaseAndPredecessors(root, base, current); len(findings) != 0 {
		t.Fatalf("predecessor findings: %+v", findings)
	}
	for _, forbidden := range []string{wave02EvidencePath, wave02AuditPath} {
		if _, err := os.Lstat(filepath.Join(root, forbidden)); !os.IsNotExist(err) {
			t.Fatalf("stale Wave 2 child exists: %s err=%v", forbidden, err)
		}
	}
}

func TestFoundationWindowRejectsPrematureAndAcceptsAdmissionChanges(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	expectations, _, findings := loadFoundationPathPolicy(root)
	if len(findings) != 0 {
		t.Fatalf("foundation path findings: %+v", findings)
	}
	base := map[string]string{
		"docs/artifacts/requirements.md": "old",
	}
	current := map[string]snapshotFile{
		"docs/artifacts/requirements.md":                   {digest: "new", regular: true, links: 1},
		"docs/artifacts/foundation-rebaseline-approval.md": {digest: "added", regular: true, links: 1},
	}
	windowFindings := validateFoundationWindow(base, current, expectations, "admission")
	rules := findingRules(windowFindings)
	if rules["FRB-SCOPE-112"] != 1 {
		t.Fatalf("wrong-window findings: %+v", windowFindings)
	}
	if windowFindings[0].subject != "docs/artifacts/requirements.md" {
		t.Fatalf("admission path was rejected instead of premature requirements: %+v", windowFindings)
	}
}

func TestFoundationPathRowsRejectExpansionOwnerAndWindowDrift(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	_, rows, findings := loadFoundationPathPolicy(root)
	if len(findings) != 0 {
		t.Fatalf("load path rows: %+v", findings)
	}
	changed := make([]tsvRow, len(rows))
	for index, row := range rows {
		changed[index] = tsvRow{}
		for key, value := range row {
			changed[index][key] = value
		}
	}
	changed[0]["owner"] = "candidate-author"
	changed[1]["window"] = "audit,admission"
	changed = append(changed, tsvRow{"change": "add", "path": "skills/escape/SKILL.md", "owner": "foundation-integrator", "window": "admission", "rule": "FRB-SCOPE-001"})
	_, validationFindings := validateFoundationPathRows(changed)
	rules := findingRules(validationFindings)
	for _, rule := range []string{"FRB-SCOPE-102", "FRB-SCOPE-107", "FRB-SCOPE-109", "FRB-SCOPE-111"} {
		if rules[rule] == 0 {
			t.Errorf("findings %+v lack %s", validationFindings, rule)
		}
	}
}

func TestFoundationApprovalRejectsTamperReplayAndScopeExpansion(t *testing.T) {
	t.Parallel()
	data, findings := readStrictFile(repositoryRoot(t), "docs/artifacts/foundation-rebaseline-approval.md")
	if len(findings) != 0 {
		t.Fatalf("read approval: %+v", findings)
	}
	tampered := []byte(strings.Replace(string(data), foundationCandidateSHA256, strings.Repeat("0", 64), 1))
	if rules := findingRules(validateFoundationApproval(tampered)); rules["FRB-APR-101"] == 0 || rules["FRB-APR-102"] == 0 {
		t.Fatalf("digest tamper was accepted: %+v", rules)
	}
	replay := append([]byte(nil), data...)
	replay = []byte(strings.Replace(string(replay), "historical `AP0`", "RECORDED AP1 and replayable authority", 1))
	if rules := findingRules(validateFoundationApproval(replay)); rules["FRB-APR-103"] == 0 {
		t.Fatalf("approval replay was accepted: %+v", rules)
	}
}

func TestFoundationSeparateAdmissionAuditCannotBeSelfAsserted(t *testing.T) {
	t.Parallel()
	fixture := []byte("# Audit\n\n| Field | Value |\n|---|---|\n| Artifact ID | `L7-FRB-ADM-AUD-001` |\n| Decision | `GO` |\n| Role | `independent-auditor` |\n| Access | `read-only candidate` |\n| Candidate-manifest SHA-256 | `" + foundationCandidateSHA256 + "` |\n| Candidate writer excluded | `true` |\n| Reviewer | `foundation-integrator` |\n\nself-audit\n")
	temp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(temp, "docs/artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(temp, foundationAdmissionAuditPath), fixture, 0o600); err != nil {
		t.Fatal(err)
	}
	if rules := findingRules(checkFoundationAdmissionAudit(temp)); rules["FRB-AUD-102"] == 0 {
		t.Fatalf("candidate-writer self-audit was accepted: %+v", rules)
	}
}

func TestFoundationCurrentCheckpointNeverInfersAdmissionAudit(t *testing.T) {
	root := repositoryRoot(t)
	if _, err := os.Lstat(filepath.Join(root, foundationAdmissionAuditPath)); !os.IsNotExist(err) {
		t.Fatalf("independent admission audit unexpectedly exists: %v", err)
	}
	result, findings := checkPolicy(root)
	if len(findings) != 0 {
		t.Fatalf("policy findings: %+v", findings)
	}
	want := "admission-in-progress"
	if foundationOptionalRegular(root, foundationAdmissionEvidencePath) {
		want = "admitted-awaiting-assurance"
	}
	if result.phase != "foundation-rebaseline" || result.checkpoint != want {
		t.Fatalf("unexpected foundation checkpoint: %+v want=%s", result, want)
	}
}
