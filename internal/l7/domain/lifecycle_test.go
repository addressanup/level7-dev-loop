package domain

import "testing"

func TestEveryValidLifecycleStateHasAnExecutableNextTransition(t *testing.T) {
	for _, test := range []struct {
		tier   RiskTier
		states []LifecycleState
	}{
		{TierRoutine, []LifecycleState{StatePlanned, StateBuilding, StateVerified, StateReviewed, StateReady}},
		{TierProduct, []LifecycleState{StatePlanned, StateBuilding, StateVerified, StateReviewed, StateReady}},
		{TierHighRisk, []LifecycleState{StatePlanned, StateAwaitingOwnerApproval, StateBuilding, StateVerified, StateAwaitingIndependentAudit, StateReviewed, StateReady}},
	} {
		for _, state := range test.states {
			next, ok := NextTransition(test.tier, state)
			if !ok || next.State == "" || next.Action == "" {
				t.Fatalf("tier %d state %q deadlocks: next=%+v ok=%v", test.tier, state, next, ok)
			}
		}
	}
}

func TestLifecycleTransitionRulesAreRiskProportionate(t *testing.T) {
	if !TransitionAllowed(TierProduct, StatePlanned, StateBuilding) {
		t.Fatal("Tier 2 planned state cannot enter building")
	}
	if TransitionAllowed(TierProduct, StatePlanned, StateAwaitingOwnerApproval) {
		t.Fatal("Tier 2 incorrectly requires owner approval")
	}
	if !TransitionAllowed(TierHighRisk, StatePlanned, StateAwaitingOwnerApproval) {
		t.Fatal("Tier 3 bypassed owner-approval state")
	}
	if TransitionAllowed(TierHighRisk, StateAwaitingOwnerApproval, StateVerified) {
		t.Fatal("Tier 3 owner-approval state bypassed implementation")
	}
	if !TransitionAllowed(TierHighRisk, StateVerified, StateAwaitingIndependentAudit) {
		t.Fatal("Tier 3 verified state bypassed independent audit")
	}
	for _, state := range []LifecycleState{StateVerified, StateAwaitingIndependentAudit, StateReviewed, StateReady} {
		if !TransitionAllowed(TierHighRisk, state, StateBuilding) {
			t.Fatalf("Tier 3 %q cannot return to building for remediation", state)
		}
	}
}

func TestDeriveLifecycleUsesOnlyConsistentCurrentFacts(t *testing.T) {
	tests := []struct {
		name  string
		facts LifecycleFacts
		state LifecycleState
		ok    bool
	}{
		{"Tier 1 planned", LifecycleFacts{Tier: TierRoutine, PlanPresent: true}, StatePlanned, true},
		{"Tier 2 building", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true}, StateBuilding, true},
		{"Tier 2 verified", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, VerificationCurrent: true}, StateVerified, true},
		{"Tier 2 reviewed", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, VerificationCurrent: true, ReviewCurrent: true}, StateReviewed, true},
		{"Tier 2 ready", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, VerificationCurrent: true, ReviewCurrent: true, ReadyCurrent: true}, StateReady, true},
		{"Tier 3 awaits approval", LifecycleFacts{Tier: TierHighRisk, PlanPresent: true}, StateAwaitingOwnerApproval, true},
		{"Tier 3 approved", LifecycleFacts{Tier: TierHighRisk, PlanPresent: true, OwnerApprovalCurrent: true}, StateBuilding, true},
		{"Tier 3 awaits audit", LifecycleFacts{Tier: TierHighRisk, PlanPresent: true, OwnerApprovalCurrent: true, WorkStarted: true, VerificationCurrent: true}, StateAwaitingIndependentAudit, true},
		{"Tier 3 reviewed", LifecycleFacts{Tier: TierHighRisk, PlanPresent: true, OwnerApprovalCurrent: true, WorkStarted: true, VerificationCurrent: true, IndependentAuditCurrent: true}, StateReviewed, true},
		{"stale assurance remediates", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, VerificationCurrent: true, ReviewCurrent: true, AssuranceStale: true}, StateBuilding, true},
		{"missing plan", LifecycleFacts{Tier: TierProduct}, "", false},
		{"review without verification", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, ReviewCurrent: true}, "", false},
		{"Tier 3 work without approval", LifecycleFacts{Tier: TierHighRisk, PlanPresent: true, WorkStarted: true}, "", false},
		{"self-contradictory assurance", LifecycleFacts{Tier: TierProduct, PlanPresent: true, WorkStarted: true, AssuranceRejected: true, AssuranceStale: true}, "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state, ok := DeriveLifecycle(test.facts)
			if state != test.state || ok != test.ok {
				t.Fatalf("DeriveLifecycle()=(%q,%v), want (%q,%v)", state, ok, test.state, test.ok)
			}
		})
	}
}

func TestRiskTierValidation(t *testing.T) {
	for tier := RiskTier(0); tier <= 4; tier++ {
		want := tier >= TierRoutine && tier <= TierHighRisk
		if tier.Valid() != want {
			t.Fatalf("tier %d Valid()=%v, want %v", tier, tier.Valid(), want)
		}
	}
}

func TestScopeContainmentIsExactOrExplicitlyRecursive(t *testing.T) {
	scope := []string{"README.md", "internal/widget/**"}
	for _, path := range []string{"README.md", "internal/widget/file.go", "internal/widget/nested/file.go"} {
		if !ScopeContains(scope, path) {
			t.Fatalf("scope does not contain %q", path)
		}
	}
	for _, path := range []string{"README.md.bak", "internal/widget", "internal/widgets/file.go", "other.go"} {
		if ScopeContains(scope, path) {
			t.Fatalf("scope unexpectedly contains %q", path)
		}
	}
}

func TestExpandedPathsAllowsOnlyDeclaredAndProductOwnedRecords(t *testing.T) {
	got := ExpandedPaths(
		[]string{"internal/widget/**"},
		[]string{"internal/widget/file.go", "docs/artifacts/changes/example.md", "outside.txt"},
		[]string{"docs/artifacts/changes/example.md"},
	)
	if len(got) != 1 || got[0] != "outside.txt" {
		t.Fatalf("ExpandedPaths()=%v, want [outside.txt]", got)
	}
}
