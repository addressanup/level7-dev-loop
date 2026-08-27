package app

import (
	"context"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTierThreeReadinessPersistsAndReconstructsExactCandidate(t *testing.T) {
	fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
	fixture.configuration.Verification[0].Benchmark = true
	verification := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
	if verification.Outcome != domain.OutcomePass {
		t.Fatalf("verification=%+v", verification)
	}
	review := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
	if review.Outcome != domain.OutcomePass {
		t.Fatalf("review=%+v", review)
	}
	ready := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady})
	if ready.Outcome != domain.OutcomePass || ready.Code != "L7-READY-000" || ready.State != string(domain.StateReady) || fixture.readinessSaves != 1 || ready.Readiness == nil || ready.Readiness.Candidate != fixture.location.Head {
		t.Fatalf("ready=%+v evidence=%+v saves=%d", ready, fixture.readiness, fixture.readinessSaves)
	}
	status := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if status.Outcome != domain.OutcomePass || status.State != string(domain.StateReady) || !strings.Contains(status.Next, "l7 merge") {
		t.Fatalf("status=%+v", status)
	}
}

func TestTierThreeReadinessFailsClosedWithoutBenchmarkOrAfterConfigDrift(t *testing.T) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	if review := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude}); review.Outcome != domain.OutcomePass {
		t.Fatalf("review=%+v", review)
	}
	blocked := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady})
	if blocked.Outcome != domain.OutcomeBlocked || blocked.Code != "L7-READY-001" || !containsDetail(blocked.Details, "L7-READY-F013") || fixture.readinessSaves != 0 {
		t.Fatalf("blocked=%+v saves=%d", blocked, fixture.readinessSaves)
	}

	fixture.configuration.Verification[0].Benchmark = true
	fixture.configuration.Digest = strings.Repeat("2", 64)
	stale := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady})
	if stale.Outcome != domain.OutcomeBlocked || !containsDetail(stale.Details, "L7-READY-F008") || fixture.readinessSaves != 0 {
		t.Fatalf("stale=%+v saves=%d", stale, fixture.readinessSaves)
	}
}

func TestReadinessRevalidatesAfterLockAcquisition(t *testing.T) {
	fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
	fixture.configuration.Verification[0].Benchmark = true
	if result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify}); result.Outcome != domain.OutcomePass {
		t.Fatal(result)
	}
	if result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude}); result.Outcome != domain.OutcomePass {
		t.Fatal(result)
	}
	fixture.onAcquire = func() { fixture.configuration.Digest = strings.Repeat("3", 64) }
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady})
	if result.Outcome != domain.OutcomeFailed || result.Code != "L7-GIT-002" || fixture.readinessSaves != 0 {
		t.Fatalf("result=%+v saves=%d", result, fixture.readinessSaves)
	}
}

func containsDetail(details []string, code string) bool {
	for _, detail := range details {
		if strings.HasPrefix(detail, code+":") {
			return true
		}
	}
	return false
}
