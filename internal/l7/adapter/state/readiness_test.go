package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestReadinessEvidenceRoundTripsStrictly(t *testing.T) {
	common := physicalCommonDirectory(t)
	evidence := testReadinessEvidence()
	if err := SaveReadiness(common, evidence); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := LoadReadiness(common)
	if err != nil || !found || loaded.ChangeID != evidence.ChangeID || loaded.Candidate != evidence.Candidate || loaded.ConfigurationDigest != evidence.ConfigurationDigest || len(loaded.Checks) != 1 {
		t.Fatalf("LoadReadiness()=%+v found=%v error=%v", loaded, found, err)
	}
	data, err := os.ReadFile(filepath.Join(common, "l7", "product", "readiness.json"))
	if err != nil || len(data) > MaxEvidenceFile || strings.Contains(string(data), "transcript") {
		t.Fatalf("readiness data=%q error=%v", data, err)
	}
}

func TestReadinessEvidenceRejectsFalseReadyAndCorruption(t *testing.T) {
	common := physicalCommonDirectory(t)
	evidence := testReadinessEvidence()
	evidence.Reviewer = evidence.Implementer
	if err := SaveReadiness(common, evidence); err == nil {
		t.Fatal("SaveReadiness accepted self-review")
	}
	directory := filepath.Join(common, "l7", "product")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "readiness.json")
	corrupt := []byte(`{"schema":1,"schema":2}`)
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadReadiness(common); err == nil {
		t.Fatal("LoadReadiness accepted duplicate fields")
	}
	if err := SaveReadiness(common, testReadinessEvidence()); err == nil {
		t.Fatal("SaveReadiness overwrote corrupt evidence")
	}
}

func testReadinessEvidence() domain.ReadinessEvidence {
	return domain.ReadinessEvidence{
		ChangeID: "product-change", Tier: domain.TierHighRisk, Base: strings.Repeat("a", 40),
		Candidate:   domain.CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)},
		BriefCommit: strings.Repeat("d", 40), ConfigurationDigest: strings.Repeat("e", 64),
		VerificationCommit: strings.Repeat("f", 40), ReviewCommit: strings.Repeat("1", 40),
		Scope: []string{"internal/product/**"}, Checks: []domain.CheckResult{{Name: "benchmark", Benchmark: true, Passed: true}},
		Owner: "accountable-owner", Implementer: domain.ProviderCodex, Reviewer: domain.ProviderClaude,
		ReviewDecision: domain.DecisionGO, BenchmarkRequired: true,
	}
}
