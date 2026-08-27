package app

import (
	"context"
	"errors"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func (application Application) readyChange(ctx context.Context, request domain.Request) domain.Result {
	if request.Headless {
		return application.result(domain.OutcomeBlocked, "L7-CAP-004", string(request.Command), "unavailable", "headless readiness is not configured in this build", "run l7 help")
	}
	if !application.readinessPortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-004", string(request.Command), "unavailable", "readiness persistence is not configured in this build", "run l7 help")
	}
	change, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil {
		return *blocked
	}
	view, err := application.loadEvidenceView(ctx, change)
	if err != nil {
		return application.failure("L7-STATE-009", request.Command, "invalid", "execution evidence cannot be reconstructed: "+bounded(err.Error(), 512), "repair the unsafe local evidence before continuing")
	}
	facts, err := application.localReadinessFacts(ctx, change, view)
	if err != nil {
		return application.failure("L7-READY-002", request.Command, "invalid", "readiness facts cannot be reconstructed: "+bounded(err.Error(), 512), "stabilize Git and local evidence, then retry l7 ready")
	}
	decision := domain.EvaluateReadiness(facts)
	if !decision.Ready {
		return application.blockedReadiness(request, change, view, facts, decision)
	}
	baseline, baselineFound, err := application.ports.LoadReadiness(change.location.CommonDir)
	if err != nil {
		return application.failure("L7-STATE-010", request.Command, "invalid", "readiness evidence is unsafe: "+bounded(err.Error(), 512), "repair or explicitly recover .git/l7/product/readiness.json")
	}

	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then retry l7 ready")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()

	rechecked, contextBlocked := application.loadExecutionContext(ctx, request.Command)
	recheckedView, evidenceErr := application.loadEvidenceView(ctx, rechecked)
	recheckedReceipt, recheckedFound, receiptErr := application.ports.LoadReadiness(change.location.CommonDir)
	if contextBlocked != nil || !sameExecutionContext(change, rechecked) || evidenceErr != nil || !sameEvidenceView(view, recheckedView) || receiptErr != nil || baselineFound != recheckedFound || (baselineFound && !sameReadinessEvidence(baseline, recheckedReceipt)) {
		return application.failure("L7-GIT-002", request.Command, "changed", "candidate or readiness evidence changed while acquiring the mutation lock", "retry l7 ready after reconstructing current state")
	}
	recheckedFacts, factErr := application.localReadinessFacts(ctx, rechecked, recheckedView)
	if factErr != nil || !sameReadinessFacts(facts, recheckedFacts) {
		return application.failure("L7-GIT-002", request.Command, "changed", "readiness facts changed before persistence", "retry l7 ready against a stable candidate")
	}
	decision = domain.EvaluateReadiness(recheckedFacts)
	if !decision.Ready {
		return application.blockedReadiness(request, rechecked, recheckedView, recheckedFacts, decision)
	}
	if err := application.ports.SaveReadiness(rechecked.location.CommonDir, recheckedFacts.Evidence); err != nil {
		return application.failure("L7-STATE-010", request.Command, "recovery-required", "readiness was not recorded: "+bounded(err.Error(), 512), "preserve the reviewed candidate and repair readiness evidence")
	}
	if err := release(); err != nil {
		released = true
		return application.failure("L7-STATE-003", request.Command, string(domain.StateReady), "readiness was recorded but lock release reported: "+bounded(err.Error(), 512), "run l7 status before continuing")
	}
	released = true
	return application.readinessResult(request, recheckedFacts.Evidence, false, "exact candidate readiness is current")
}

func (application Application) localReadinessFacts(ctx context.Context, change executionContext, view evidenceView) (domain.ReadinessFacts, error) {
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || !sameLocation(change.location, pending.RepositoryLocation) {
		return domain.ReadinessFacts{}, errors.New("cannot establish a stable clean repository")
	}
	clean := len(pending.Paths) == 0 && !pending.IndexDirty
	briefCommit := change.active.Base
	if change.active.Kind == domain.ActiveBrief {
		briefCommit, err = application.ports.PathCommit(ctx, change.location.Root, change.active.BriefPath)
		if err != nil {
			return domain.ReadinessFacts{}, errors.New("cannot resolve the tracked brief commit")
		}
	}
	owner := ""
	approvalCurrent := change.active.Tier != domain.TierHighRisk
	if change.active.Tier == domain.TierHighRisk {
		approval, found, approvalErr := application.ports.LoadApproval(change.location.CommonDir)
		if approvalErr != nil {
			return domain.ReadinessFacts{}, approvalErr
		}
		if found {
			owner = approval.Actor
			expected := approval.Implementer
			if view.runFound && view.run.Provider.Provider.Valid() {
				expected = view.run.Provider.Provider
			}
			approvalCurrent = approvalMatches(approval, change.active.ID, expected, briefCommit)
		}
	}
	runCurrent := clean && view.runValid && runAtCurrentChain(change, view)
	verificationCurrent := runCurrent && view.verificationValid && verificationAtCurrentChain(change, view)
	reviewCurrent := verificationCurrent && view.reviewValid && reviewAtCurrent(change, view.review) && view.review.Decision == domain.DecisionGO
	verificationCommit := ""
	reviewCommit := ""
	if view.verificationFound {
		verificationCommit = view.verification.Candidate.Commit
		if change.active.Tier == domain.TierHighRisk {
			verificationCommit = view.verification.VerificationCommit
		}
	}
	if view.reviewFound {
		reviewCommit = view.review.Candidate.Commit
		if change.active.Tier == domain.TierHighRisk {
			reviewCommit = view.review.ReviewCommit
		}
	}
	evidence := domain.ReadinessEvidence{
		ChangeID: change.active.ID, Tier: change.active.Tier, Base: change.active.Base,
		Candidate:   domain.CandidateIdentity{Commit: change.location.Head, Tree: change.location.Tree},
		BriefCommit: briefCommit, ConfigurationDigest: change.configuration.Digest,
		VerificationCommit: verificationCommit, ReviewCommit: reviewCommit,
		Scope: append([]string{}, change.active.Scope...), Owner: owner,
		ReviewDecision: view.review.Decision, BenchmarkRequired: change.active.Tier == domain.TierHighRisk || configurationRequiresBenchmark(change.configuration),
	}
	if view.runFound {
		evidence.Implementer = view.run.Provider.Provider
	}
	if view.verificationFound {
		evidence.Checks = append([]domain.CheckResult{}, view.verification.Checks...)
	}
	if view.reviewFound {
		evidence.Reviewer = view.review.Provider.Provider
	}
	return domain.ReadinessFacts{
		Evidence: evidence, PlanCurrent: true, RepositoryClean: clean,
		ApprovalCurrent: approvalCurrent, VerificationCurrent: verificationCurrent,
		ReviewCurrent: change.active.Tier != domain.TierHighRisk && reviewCurrent,
		AuditCurrent:  change.active.Tier == domain.TierHighRisk && reviewCurrent,
	}, nil
}

func (application Application) blockedReadiness(request domain.Request, change executionContext, view evidenceView, facts domain.ReadinessFacts, decision domain.ReadinessDecision) domain.Result {
	state := domain.StateReviewed
	if !facts.VerificationCurrent || (!facts.ReviewCurrent && !facts.AuditCurrent) {
		state = domain.StateBuilding
	}
	result := application.result(domain.OutcomeBlocked, "L7-READY-001", string(request.Command), string(state), "exact candidate readiness is blocked", "remediate the reported readiness findings, then run l7 ready")
	result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
	result.Execution = executionDetails(view)
	result.Readiness = readinessDetails(facts.Evidence, false, false)
	for _, finding := range decision.Findings {
		result.Details = append(result.Details, finding.Code+": "+finding.Message)
	}
	return result
}

func (application Application) readinessResult(request domain.Request, evidence domain.ReadinessEvidence, headless bool, message string) domain.Result {
	result := application.result(domain.OutcomePass, "L7-READY-000", string(request.Command), string(domain.StateReady), message, "run l7 merge --target <branch> and confirm the full candidate SHA")
	result.Readiness = readinessDetails(evidence, headless, true)
	return result
}

func readinessDetails(evidence domain.ReadinessEvidence, headless, ready bool) *domain.ReadinessDetails {
	return &domain.ReadinessDetails{
		Headless: headless, Ready: ready, Base: evidence.Base, Candidate: evidence.Candidate.Commit, Tree: evidence.Candidate.Tree,
		BriefCommit: evidence.BriefCommit, ConfigurationDigest: evidence.ConfigurationDigest,
		VerificationCommit: evidence.VerificationCommit, ReviewCommit: evidence.ReviewCommit,
		Owner: evidence.Owner, Implementer: evidence.Implementer, Reviewer: evidence.Reviewer,
		Checks: append([]domain.CheckResult{}, evidence.Checks...),
	}
}

func configurationRequiresBenchmark(configuration domain.Configuration) bool {
	for _, command := range configuration.Verification {
		if command.Benchmark {
			return true
		}
	}
	return false
}

func sameReadinessEvidence(left, right domain.ReadinessEvidence) bool {
	if left.ChangeID != right.ChangeID || left.Tier != right.Tier || left.Base != right.Base || left.Candidate != right.Candidate || left.BriefCommit != right.BriefCommit || left.ConfigurationDigest != right.ConfigurationDigest || left.VerificationCommit != right.VerificationCommit || left.ReviewCommit != right.ReviewCommit || left.Owner != right.Owner || left.Implementer != right.Implementer || left.Reviewer != right.Reviewer || left.ReviewDecision != right.ReviewDecision || left.BenchmarkRequired != right.BenchmarkRequired || !sameStrings(left.Scope, right.Scope) || len(left.Checks) != len(right.Checks) {
		return false
	}
	for index := range left.Checks {
		if left.Checks[index] != right.Checks[index] {
			return false
		}
	}
	return true
}

func sameReadinessFacts(left, right domain.ReadinessFacts) bool {
	return sameReadinessEvidence(left.Evidence, right.Evidence) && left.PlanCurrent == right.PlanCurrent && left.RepositoryClean == right.RepositoryClean && left.ApprovalCurrent == right.ApprovalCurrent && left.VerificationCurrent == right.VerificationCurrent && left.ReviewCurrent == right.ReviewCurrent && left.AuditCurrent == right.AuditCurrent
}

func (application Application) readinessPortsAvailable() bool {
	return application.executionPortsAvailable() && application.ports.LoadReadiness != nil && application.ports.SaveReadiness != nil
}
