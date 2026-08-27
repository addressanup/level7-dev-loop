package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestExecutionEvidenceRoundTripsWithoutTranscript(t *testing.T) {
	common := physicalCommonDirectory(t)
	provider := testProvider()
	run := domain.RunEvidence{ChangeID: "product-change", Provider: provider, Candidate: testCandidate("a", "b")}
	if err := SaveRun(common, run); err != nil {
		t.Fatal(err)
	}
	loadedRun, found, err := LoadRun(common)
	if err != nil || !found || loadedRun.ChangeID != run.ChangeID || loadedRun.Provider.Digest != provider.Digest || loadedRun.Candidate != run.Candidate {
		t.Fatalf("LoadRun()=%+v found=%v error=%v", loadedRun, found, err)
	}

	verification := domain.VerificationEvidence{ChangeID: run.ChangeID, Candidate: run.Candidate, Result: domain.DecisionGO, Checks: []domain.CheckResult{{Name: "test", Passed: true, ExitCode: 0, Code: "L7-VERIFY-000", Message: "passed"}}}
	if err := SaveVerification(common, verification); err != nil {
		t.Fatal(err)
	}
	loadedVerification, found, err := LoadVerification(common)
	if err != nil || !found || loadedVerification.Candidate != verification.Candidate || len(loadedVerification.Checks) != 1 {
		t.Fatalf("LoadVerification()=%+v found=%v error=%v", loadedVerification, found, err)
	}

	review := domain.ReviewEvidence{ChangeID: run.ChangeID, Provider: domain.ProviderIdentity{Provider: domain.ProviderClaude, Executable: "/usr/bin/claude", Version: "2.1.241", Digest: strings.Repeat("d", 64), Capability: domain.CapabilityAvailable}, Candidate: run.Candidate, Decision: domain.DecisionGO, Findings: []string{"No blocking finding."}}
	if err := SaveReview(common, review); err != nil {
		t.Fatal(err)
	}
	loadedReview, found, err := LoadReview(common)
	if err != nil || !found || loadedReview.Decision != domain.DecisionGO || len(loadedReview.Findings) != 1 {
		t.Fatalf("LoadReview()=%+v found=%v error=%v", loadedReview, found, err)
	}

	for _, name := range []string{"run.json", "verification.json", "review.json"} {
		data, readErr := os.ReadFile(filepath.Join(common, "l7", "product", name))
		if readErr != nil || strings.Contains(string(data), "transcript") || strings.Contains(string(data), "reasoning") {
			t.Fatalf("%s data=%q error=%v", name, data, readErr)
		}
	}
}

func TestEvidenceRejectsCorruptionAndDoesNotOverwriteIt(t *testing.T) {
	common := physicalCommonDirectory(t)
	directory := filepath.Join(common, "l7", "product")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "run.json")
	corrupt := []byte(`{"schema":1,"schema":2}`)
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadRun(common); err == nil {
		t.Fatal("LoadRun() accepted duplicate JSON fields")
	}
	if err := SaveRun(common, domain.RunEvidence{ChangeID: "product-change", Provider: testProvider(), Candidate: testCandidate("a", "b")}); err == nil {
		t.Fatal("SaveRun() overwrote corrupt evidence")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(corrupt) {
		t.Fatalf("corruption changed: %q error=%v", after, err)
	}
}

func TestTrackedEvidenceArtifactsAreBoundAndBounded(t *testing.T) {
	root := physicalCommonDirectory(t)
	verification := domain.VerificationEvidence{ChangeID: "product-change", Candidate: testCandidate("a", "b"), Result: domain.DecisionGO, Checks: []domain.CheckResult{{Name: "test", Passed: true}}}
	path, err := WriteVerificationArtifact(root, verification, "local-verifier")
	if err != nil || path != "docs/artifacts/changes/product-change-verification.md" {
		t.Fatalf("WriteVerificationArtifact() path=%q error=%v", path, err)
	}
	review := domain.ReviewEvidence{ChangeID: "product-change", Provider: domain.ProviderIdentity{Provider: domain.ProviderClaude, Executable: "/usr/bin/claude", Version: "2.1.241", Digest: strings.Repeat("d", 64), Capability: domain.CapabilityAvailable}, Candidate: testCandidate("c", "d"), Decision: domain.DecisionNoGO, Findings: []string{"A blocking issue remains."}}
	path, err = WriteAuditArtifact(root, review)
	if err != nil || path != "docs/artifacts/changes/product-change-audit.md" {
		t.Fatalf("WriteAuditArtifact() path=%q error=%v", path, err)
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil || !strings.Contains(string(data), "| Result | `NO_GO` |") {
		t.Fatalf("audit data=%q error=%v", data, err)
	}
}

func testProvider() domain.ProviderIdentity {
	return domain.ProviderIdentity{Provider: domain.ProviderCodex, Executable: "/usr/bin/codex", Version: "0.149.1", Digest: strings.Repeat("c", 64), Capability: domain.CapabilityAvailable}
}

func testCandidate(commit, tree string) domain.CandidateIdentity {
	return domain.CandidateIdentity{Commit: strings.Repeat(commit, 40), Tree: strings.Repeat(tree, 40)}
}
