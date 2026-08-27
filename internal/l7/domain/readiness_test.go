package domain

import (
	"strings"
	"testing"
)

func TestEvaluateReadinessAcceptsExactTierThreeFacts(t *testing.T) {
	facts := readyFacts()
	decision := EvaluateReadiness(facts)
	if !decision.Ready || len(decision.Findings) != 0 || !ReadinessEvidenceValid(facts.Evidence) {
		t.Fatalf("EvaluateReadiness()=%+v", decision)
	}
}

func TestEvaluateReadinessFailsClosedForFalseReadyFacts(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ReadinessFacts)
		code string
	}{
		{"dirty", func(facts *ReadinessFacts) { facts.RepositoryClean = false }, "L7-READY-F008"},
		{"stale verification", func(facts *ReadinessFacts) { facts.VerificationCurrent = false }, "L7-READY-F008"},
		{"missing approval", func(facts *ReadinessFacts) { facts.ApprovalCurrent = false }, "L7-READY-F009"},
		{"self review", func(facts *ReadinessFacts) { facts.Evidence.Reviewer = facts.Evidence.Implementer }, "L7-READY-F010"},
		{"NO_GO", func(facts *ReadinessFacts) { facts.Evidence.ReviewDecision = DecisionNoGO }, "L7-READY-F007"},
		{"config drift", func(facts *ReadinessFacts) { facts.Evidence.ConfigurationDigest = "" }, "L7-READY-F004"},
		{"benchmark absent", func(facts *ReadinessFacts) { facts.Evidence.Checks[0].Benchmark = false }, "L7-READY-F013"},
		{"benchmark waived", func(facts *ReadinessFacts) { facts.Evidence.BenchmarkRequired = false }, "L7-READY-F011"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			facts := readyFacts()
			test.edit(&facts)
			decision := EvaluateReadiness(facts)
			if decision.Ready || !hasReadinessFinding(decision, test.code) {
				t.Fatalf("EvaluateReadiness()=%+v, want %s", decision, test.code)
			}
		})
	}
}

func readyFacts() ReadinessFacts {
	evidence := ReadinessEvidence{
		ChangeID: "product-change", Tier: TierHighRisk, Base: strings.Repeat("a", 40),
		Candidate:   CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)},
		BriefCommit: strings.Repeat("d", 40), ConfigurationDigest: strings.Repeat("e", 64),
		VerificationCommit: strings.Repeat("f", 40), ReviewCommit: strings.Repeat("1", 40),
		Scope:  []string{"internal/product/**"},
		Checks: []CheckResult{{Name: "benchmark", Benchmark: true, Passed: true, ExitCode: 0}},
		Owner:  "accountable-owner", Implementer: ProviderCodex, Reviewer: ProviderClaude,
		ReviewDecision: DecisionGO, BenchmarkRequired: true,
	}
	return ReadinessFacts{Evidence: evidence, PlanCurrent: true, RepositoryClean: true, ApprovalCurrent: true, VerificationCurrent: true, AuditCurrent: true}
}

func hasReadinessFinding(decision ReadinessDecision, code string) bool {
	for _, finding := range decision.Findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}
