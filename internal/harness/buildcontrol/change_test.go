package main

import "testing"

func TestEveryAcceptedStateHasAnExecutableNextTransition(t *testing.T) {
	states := []workflowState{statePlanned, stateBuilding, stateVerified, stateReviewed, stateReady}
	for _, tier := range []riskTier{tierRoutine, tierProduct} {
		for _, state := range states {
			next, action, ok := nextState(tier, state)
			if !ok || next == "" || action == "" {
				t.Fatalf("tier %d state %s deadlocks: next=%s action=%q ok=%v", tier, state, next, action, ok)
			}
		}
	}
	highRiskStates := []workflowState{statePlanned, stateAwaitingOwnerApproval, stateBuilding, stateVerified, stateAwaitingIndependentAudit, stateReviewed, stateReady}
	for _, state := range highRiskStates {
		next, action, ok := nextState(tierHighRisk, state)
		if !ok || next == "" || action == "" {
			t.Fatalf("Tier 3 state %s deadlocks: next=%s action=%q ok=%v", state, next, action, ok)
		}
	}
}

func TestApprovalAndRemediationTransitionsCannotDeadlock(t *testing.T) {
	if !validateTransition(tierHighRisk, stateAwaitingOwnerApproval, stateBuilding) {
		t.Fatal("owner approval cannot enter building")
	}
	for _, from := range []workflowState{stateVerified, stateReviewed, stateAwaitingIndependentAudit} {
		if !validateTransition(tierHighRisk, from, stateBuilding) {
			t.Fatalf("%s cannot return to building for remediation", from)
		}
	}
}

func TestBriefParserRequiresOneCompleteScopedRecord(t *testing.T) {
	document := briefDocument("feature-x", tierProduct, "0123456789012345678901234567890123456789", "internal/feature.go")
	brief, findings := parseBrief("docs/artifacts/changes/feature-x.md", []byte(document))
	if len(findings) != 0 || brief.ID != "feature-x" || brief.Tier != tierProduct || !scopeContains(brief.Scope, "internal/feature.go") {
		t.Fatalf("valid brief rejected: brief=%+v findings=%+v", brief, findings)
	}
	_, findings = parseBrief("docs/artifacts/changes/feature-x.md", []byte("# incomplete\n"))
	if len(findings) == 0 {
		t.Fatal("incomplete brief accepted")
	}
}
