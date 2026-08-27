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
		{"stale plan", func(facts *ReadinessFacts) { facts.PlanCurrent = false }, "L7-READY-F008"},
		{"stale verification", func(facts *ReadinessFacts) { facts.VerificationCurrent = false }, "L7-READY-F008"},
		{"missing approval", func(facts *ReadinessFacts) { facts.ApprovalCurrent = false }, "L7-READY-F009"},
		{"missing audit", func(facts *ReadinessFacts) { facts.AuditCurrent = false }, "L7-READY-F009"},
		{"owner is implementer", func(facts *ReadinessFacts) { facts.Evidence.Owner = string(facts.Evidence.Implementer) }, "L7-READY-F010"},
		{"owner is reviewer", func(facts *ReadinessFacts) { facts.Evidence.Owner = string(facts.Evidence.Reviewer) }, "L7-READY-F010"},
		{"self review", func(facts *ReadinessFacts) { facts.Evidence.Reviewer = facts.Evidence.Implementer }, "L7-READY-F010"},
		{"NO_GO", func(facts *ReadinessFacts) { facts.Evidence.ReviewDecision = DecisionNoGO }, "L7-READY-F007"},
		{"config drift", func(facts *ReadinessFacts) { facts.Evidence.ConfigurationDigest = "" }, "L7-READY-F004"},
		{"candidate equals base", func(facts *ReadinessFacts) { facts.Evidence.Candidate.Commit = facts.Evidence.Base }, "L7-READY-F002"},
		{"missing brief", func(facts *ReadinessFacts) { facts.Evidence.BriefCommit = "" }, "L7-READY-F003"},
		{"unsafe scope", func(facts *ReadinessFacts) {
			facts.Evidence.Scope = []string{"internal/product/**", "internal/product/**"}
		}, "L7-READY-F005"},
		{"failing check", func(facts *ReadinessFacts) { facts.Evidence.Checks[0].Passed = false }, "L7-READY-F006"},
		{"duplicate check", func(facts *ReadinessFacts) {
			facts.Evidence.Checks = append(facts.Evidence.Checks, facts.Evidence.Checks[0])
		}, "L7-READY-F006"},
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

func TestMergeReceiptRequiresExactLocalRefAndBindings(t *testing.T) {
	receipt := MergeReceipt{
		ChangeID: "product-change", TargetRef: "refs/heads/main", PreviousCommit: strings.Repeat("a", 40),
		Candidate:           CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)},
		ConfigurationDigest: strings.Repeat("d", 64), VerificationCommit: strings.Repeat("e", 40), ReviewCommit: strings.Repeat("f", 40),
	}
	if !MergeReceiptValid(receipt) {
		t.Fatalf("valid receipt rejected: %+v", receipt)
	}
	for _, target := range []string{"main", "refs/tags/main", "refs/heads/../main", "refs/heads/main.lock", "refs/heads/space name"} {
		receipt.TargetRef = target
		if MergeReceiptValid(receipt) {
			t.Fatalf("unsafe target %q accepted", target)
		}
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
