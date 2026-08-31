package domain

import "testing"

func TestRouteChoosesLeastCostQualifiedModelAndEffort(t *testing.T) {
	task := TaskProfile{Schema: OrchestrationSchema, ID: "task", Complexity: ComplexityC2, RiskTier: TierProduct, ContextTokens: 20_000, NeedsTools: true, NeedsEditing: true}
	snapshots := []ProviderSnapshot{
		{Schema: OrchestrationSchema, ID: "expensive", Authentication: AuthAuthenticated, Models: []ModelCapability{{ID: "a", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, Efforts: []ReasoningEffort{EffortMedium, EffortHigh}, CostClass: 4, LatencyClass: 2, Verified: true}}},
		{Schema: OrchestrationSchema, ID: "efficient", Authentication: AuthAuthenticated, Models: []ModelCapability{{ID: "b", Languages: []string{"go"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, Efforts: []ReasoningEffort{EffortLow, EffortMedium, EffortHigh}, CostClass: 2, LatencyClass: 2, Verified: true}}},
	}
	decision := Route(task, snapshots)
	if decision.ProviderID != "efficient" || decision.ModelID != "b" || decision.Effort != EffortMedium || len(decision.Fallbacks) != 1 {
		t.Fatalf("unexpected decision: %#v", decision)
	}
}

func TestRouteFailsClosedOnLanguageAndRaisesSensitiveWorkFloor(t *testing.T) {
	task := TaskProfile{ID: "security", Complexity: ComplexityC1, RiskTier: TierProduct, Languages: []string{"rust"}, WorkKinds: []string{"security"}}
	snapshots := []ProviderSnapshot{{ID: "gateway", Authentication: AuthAuthenticated, Models: []ModelCapability{
		{ID: "go-only", Languages: []string{"go"}, Efforts: []ReasoningEffort{EffortHigh}, CostClass: 1, LatencyClass: 1, Verified: true},
		{ID: "polyglot", Languages: []string{"*"}, Efforts: []ReasoningEffort{EffortLow, EffortHigh}, CostClass: 2, LatencyClass: 2, Verified: true},
	}}}
	decision := Route(task, snapshots)
	if decision.ProviderID != "gateway" || decision.ModelID != "polyglot" || decision.Effort != EffortHigh {
		t.Fatalf("unexpected sensitive route: %#v", decision)
	}
	if decision.Candidates[1].Qualified || decision.Candidates[1].Reasons[0] != "language rust is not advertised" {
		t.Fatalf("language constraint did not fail closed: %#v", decision.Candidates)
	}
}

func TestRouteAppliesRiskFloorAndRejectsSelfAudit(t *testing.T) {
	task := TaskProfile{Schema: OrchestrationSchema, ID: "audit", Complexity: ComplexityC1, RiskTier: TierHighRisk, IndependentReview: true, ImplementerProvider: "codex", ImplementerModel: "primary"}
	snapshots := []ProviderSnapshot{
		{ID: "codex", Authentication: AuthAuthenticated, Models: []ModelCapability{{ID: "primary", Efforts: []ReasoningEffort{EffortHigh}, CostClass: 1, LatencyClass: 1, Verified: true}, {ID: "other", Efforts: []ReasoningEffort{EffortMedium}, CostClass: 1, LatencyClass: 1, Verified: true}}},
		{ID: "claude", Authentication: AuthAuthenticated, Models: []ModelCapability{{ID: "reviewer", Efforts: []ReasoningEffort{EffortHigh}, CostClass: 2, LatencyClass: 2, Verified: true}}},
	}
	decision := Route(task, snapshots)
	if decision.ProviderID != "claude" || decision.Effort != EffortHigh {
		t.Fatalf("unexpected audit route: %#v", decision)
	}
	if decision.Candidates[1].Qualified || decision.Candidates[2].Qualified {
		t.Fatalf("unsafe candidates qualified: %#v", decision.Candidates)
	}
}

func TestRouteFailsClosedWithoutVerifiedCapability(t *testing.T) {
	decision := Route(TaskProfile{ID: "x", Complexity: ComplexityC1, RiskTier: TierRoutine}, []ProviderSnapshot{{ID: "p", Authentication: AuthAuthenticated, Models: []ModelCapability{{ID: "m", Efforts: []ReasoningEffort{EffortLow}, CostClass: 1, LatencyClass: 1}}}})
	if decision.ProviderID != "" || len(decision.Candidates) != 1 || decision.Candidates[0].Qualified {
		t.Fatalf("unverified route was selected: %#v", decision)
	}
}
