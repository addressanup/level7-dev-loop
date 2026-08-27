package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestHeadlessReadinessUsesOnlyThePureDecoderPort(t *testing.T) {
	calls := 0
	application := NewLifecycle("test-version", "", Ports{DecodeCI: func(data []byte) (domain.ReadinessFacts, error) {
		calls++
		if string(data) != "trusted" {
			return domain.ReadinessFacts{}, errors.New("unexpected input")
		}
		return readyFactsForApp(), nil
	}})
	result := application.ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady, Headless: true, Input: []byte("trusted")})
	if result.Outcome != domain.OutcomePass || result.Code != "L7-READY-000" || calls != 1 || result.Readiness == nil || !result.Readiness.Headless {
		t.Fatalf("result=%+v calls=%d", result, calls)
	}
}

func readyFactsForApp() domain.ReadinessFacts {
	evidence := testReadinessEvidenceForApp()
	return domain.ReadinessFacts{Evidence: evidence, PlanCurrent: true, RepositoryClean: true, ApprovalCurrent: true, VerificationCurrent: true, AuditCurrent: true}
}

func testReadinessEvidenceForApp() domain.ReadinessEvidence {
	return domain.ReadinessEvidence{
		ChangeID: "product-change", Tier: domain.TierHighRisk,
		Base:                "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Candidate:           domain.CandidateIdentity{Commit: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Tree: "cccccccccccccccccccccccccccccccccccccccc"},
		BriefCommit:         "dddddddddddddddddddddddddddddddddddddddd",
		ConfigurationDigest: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		VerificationCommit:  "ffffffffffffffffffffffffffffffffffffffff", ReviewCommit: "1111111111111111111111111111111111111111",
		Scope: []string{"internal/product/**"}, Checks: []domain.CheckResult{{Name: "benchmark", Benchmark: true, Passed: true}},
		Owner: "accountable-owner", Implementer: domain.ProviderCodex, Reviewer: domain.ProviderClaude,
		ReviewDecision: domain.DecisionGO, BenchmarkRequired: true,
	}
}

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

func TestConfirmedMergeAdvancesOnceAndStatusReconstructsMerged(t *testing.T) {
	fixture := completedReadyFixture()
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
	if result.Outcome != domain.OutcomePass || result.Code != "L7-MERGE-000" || result.State != string(domain.StateMerged) || fixture.mergeConfirmations != 1 || fixture.mergeAdvances != 1 || !fixture.mergeFound || result.Readiness == nil || result.Readiness.TargetRef != "refs/heads/release-target" {
		t.Fatalf("result=%+v confirmations=%d advances=%d receipt=%+v", result, fixture.mergeConfirmations, fixture.mergeAdvances, fixture.merge)
	}
	status := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if status.Outcome != domain.OutcomePass || status.State != string(domain.StateMerged) || status.Readiness == nil || status.Readiness.TargetRef != "refs/heads/release-target" {
		t.Fatalf("status=%+v", status)
	}
	idempotent := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
	if idempotent.Outcome != domain.OutcomePass || fixture.mergeConfirmations != 1 || fixture.mergeAdvances != 1 {
		t.Fatalf("idempotent=%+v confirmations=%d advances=%d", idempotent, fixture.mergeConfirmations, fixture.mergeAdvances)
	}
}

func TestMergeConfirmationAndPostConfirmationRaceFailClosed(t *testing.T) {
	t.Run("confirmation", func(t *testing.T) {
		fixture := completedReadyFixture()
		fixture.mergeConfirmationError = errors.New("mismatch")
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
		if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-AUTH-003" || fixture.mergeAdvances != 0 || fixture.mergeFound {
			t.Fatalf("result=%+v advances=%d receipt=%+v", result, fixture.mergeAdvances, fixture.merge)
		}
	})
	t.Run("race", func(t *testing.T) {
		fixture := completedReadyFixture()
		fixture.onConfirmMerge = func() { fixture.mergeTargetCommit = strings.Repeat("9", 40) }
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
		if result.Outcome != domain.OutcomeFailed || result.Code != "L7-MERGE-002" || fixture.mergeAdvances != 0 || fixture.mergeFound {
			t.Fatalf("result=%+v advances=%d receipt=%+v", result, fixture.mergeAdvances, fixture.merge)
		}
	})
}

func TestMergeRecoversRefEffectWhenReceiptPersistenceWasInterrupted(t *testing.T) {
	fixture := completedReadyFixture()
	fixture.mergeSaveError = errors.New("interrupted receipt")
	first := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
	if first.Outcome != domain.OutcomeFailed || first.Code != "L7-STATE-011" || fixture.mergeTargetCommit != fixture.readiness.Candidate.Commit || fixture.mergeFound || fixture.mergeAdvances != 1 {
		t.Fatalf("first=%+v target=%s receipt=%+v advances=%d", first, fixture.mergeTargetCommit, fixture.merge, fixture.mergeAdvances)
	}
	fixture.mergeSaveError = nil
	second := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandMerge, TargetBranch: "release-target"})
	if second.Outcome != domain.OutcomePass || fixture.mergeAdvances != 1 || fixture.mergeConfirmations != 2 || !fixture.mergeFound || !strings.Contains(second.Message, "recovered") {
		t.Fatalf("second=%+v confirmations=%d advances=%d receipt=%+v", second, fixture.mergeConfirmations, fixture.mergeAdvances, fixture.merge)
	}
}

func completedReadyFixture() *executionFixture {
	fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
	fixture.configuration.Verification[0].Benchmark = true
	if result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify}); result.Outcome != domain.OutcomePass {
		panic("fake verification did not complete: " + result.Message)
	}
	if result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude}); result.Outcome != domain.OutcomePass {
		panic("fake review did not complete: " + result.Message)
	}
	if result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReady}); result.Outcome != domain.OutcomePass {
		panic("fake readiness did not complete: " + result.Message)
	}
	fixture.mergeConfirmations = 0
	return fixture
}

func containsDetail(details []string, code string) bool {
	for _, detail := range details {
		if strings.HasPrefix(detail, code+":") {
			return true
		}
	}
	return false
}
