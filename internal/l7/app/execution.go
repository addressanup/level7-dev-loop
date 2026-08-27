package app

import (
	"context"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type executionContext struct {
	location      domain.RepositoryLocation
	configuration domain.Configuration
	active        domain.ActiveChange
	brief         domain.ChangeBrief
	snapshot      domain.RepositorySnapshot
}

type evidenceView struct {
	run               domain.RunEvidence
	runFound          bool
	runValid          bool
	verification      domain.VerificationEvidence
	verificationFound bool
	verificationValid bool
	review            domain.ReviewEvidence
	reviewFound       bool
	reviewValid       bool
}

func (application Application) runChange(ctx context.Context, request domain.Request) domain.Result {
	if !application.executionPortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-003", string(request.Command), "unavailable", "provider execution is not configured in this build", "run l7 help")
	}
	if !request.Agent.Valid() || !domain.ConventionalSubject(request.CommitMessage) {
		return application.Invalid(string(request.Command), "run requires --agent codex|claude and one bounded conventional --message")
	}
	change, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil {
		return *blocked
	}
	status := application.statusFromContext(ctx, domain.Request{Command: domain.CommandStatus}, change)
	if status.Outcome == domain.OutcomeFailed || status.Code == "L7-SCOPE-001" {
		status.Command = string(request.Command)
		return status
	}
	if status.State == string(domain.StateVerified) || status.State == string(domain.StateAwaitingIndependentAudit) || status.State == string(domain.StateReviewed) || status.State == string(domain.StateReady) || status.State == string(domain.StateMerged) {
		return application.result(domain.OutcomeBlocked, "L7-RUN-004", string(request.Command), status.State, "the current candidate already has later lifecycle evidence", status.Next)
	}

	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then retry l7 run")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil || !sameExecutionContext(change, rechecked) {
		return application.failure("L7-GIT-002", request.Command, "changed", "repository, configuration, or active context changed before provider execution", "retry l7 run against stable intake")
	}
	change = rechecked

	briefCommit := change.active.Base
	if change.active.Kind == domain.ActiveBrief {
		briefCommit, err = application.ports.PathCommit(ctx, change.location.Root, change.active.BriefPath)
		if err != nil {
			return application.result(domain.OutcomeBlocked, "L7-BRIEF-005", string(request.Command), "unstable-intake", "the tracked change brief is not committed", "commit the exact change brief, then retry l7 run")
		}
	}
	baselineEvidence, err := application.loadEvidenceState(change.location.CommonDir)
	if err != nil {
		return application.failure("L7-STATE-005", request.Command, "invalid", "execution evidence is invalid: "+bounded(err.Error(), 512), "repair the unsafe local evidence before continuing")
	}
	existing, found := baselineEvidence.run, baselineEvidence.runFound
	recovering := found && existing.ChangeID == change.active.ID && existing.Candidate.Commit == ""
	if recovering && (existing.Provider.Provider != request.Agent || existing.CommitMessage != request.CommitMessage) {
		return application.result(domain.OutcomeBlocked, "L7-RUN-003", string(request.Command), "recovery-required", "an interrupted run is bound to a different provider or commit subject", "rerun l7 run with the original --agent and --message")
	}
	if found && existing.ChangeID == change.active.ID && existing.Candidate.Commit == change.location.Head && existing.Candidate.Tree == change.location.Tree {
		return application.result(domain.OutcomeBlocked, "L7-RUN-004", string(request.Command), string(domain.StateBuilding), "the current implementation run is already committed", "run l7 verify")
	}
	if !recovering {
		pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		if pendingErr != nil || !sameLocation(change.location, pending.RepositoryLocation) {
			return application.failure("L7-GIT-001", request.Command, "invalid", "cannot establish a stable provider boundary", "stabilize the Git worktree, then retry l7 run")
		}
		if len(pending.Paths) != 0 || pending.IndexDirty {
			return application.result(domain.OutcomeBlocked, "L7-RUN-001", string(request.Command), "dirty", "provider execution requires a clean worktree and index", "commit or restore pending work, then retry l7 run")
		}
	}
	if change.active.Tier == domain.TierHighRisk {
		approval, found, loadErr := application.ports.LoadApproval(change.location.CommonDir)
		if loadErr != nil {
			return application.failure("L7-AUTH-003", request.Command, "invalid", "owner approval state is invalid: "+bounded(loadErr.Error(), 512), "repair the unsafe local approval record before continuing")
		}
		if !found || !approvalMatches(approval, change.active.ID, request.Agent, briefCommit) {
			approval, err = application.ports.ConfirmApproval(ctx, change.active.ID, request.Agent, briefCommit)
			if err != nil {
				return application.operationError(ctx, request.Command, "L7-AUTH-002", "awaiting-owner-approval", "Tier 3 owner approval was not recorded: "+bounded(err.Error(), 512), "rerun l7 run from an interactive terminal and confirm the exact change")
			}
			if err := application.ports.SaveApproval(change.location.CommonDir, approval); err != nil {
				return application.failure("L7-AUTH-003", request.Command, "recovery-required", "owner approval could not be recorded: "+bounded(err.Error(), 512), "inspect .git/l7/product/approval.json before retrying")
			}
		}
	}
	rechecked, blocked = application.loadExecutionContext(ctx, request.Command)
	if blocked != nil || !sameExecutionContext(change, rechecked) {
		return application.failure("L7-GIT-002", request.Command, "changed", "repository, configuration, or active context changed during owner approval", "retry l7 run against stable intake")
	}
	change = rechecked
	if change.active.Kind == domain.ActiveBrief {
		recheckedBriefCommit, pathErr := application.ports.PathCommit(ctx, change.location.Root, change.active.BriefPath)
		if pathErr != nil || recheckedBriefCommit != briefCommit {
			return application.failure("L7-BRIEF-005", request.Command, "changed", "change brief commit changed during owner approval", "retry l7 run against the exact committed brief")
		}
	}
	recheckedEvidence, loadErr := application.loadEvidenceState(change.location.CommonDir)
	if loadErr != nil || !sameEvidenceState(baselineEvidence, recheckedEvidence) {
		return application.failure("L7-STATE-005", request.Command, "changed", "execution evidence changed during owner approval", "retry l7 run after reconstructing current state")
	}
	if recovering {
		result := application.resumeRun(ctx, request, change, existing)
		if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
			result = application.failure("L7-STATE-003", request.Command, result.State, "run completed but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
		}
		released = true
		return result
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || !sameLocation(change.location, pending.RepositoryLocation) {
		return application.failure("L7-GIT-001", request.Command, "invalid", "cannot establish a stable provider boundary", "stabilize the Git worktree, then retry l7 run")
	}
	if len(pending.Paths) != 0 || pending.IndexDirty {
		return application.result(domain.OutcomeBlocked, "L7-RUN-001", string(request.Command), "dirty", "provider execution requires a clean worktree and index", "commit or restore pending work, then retry l7 run")
	}

	task := providerTask(change, request.Agent, domain.RoleImplementer, domain.CandidateIdentity{Commit: change.location.Head, Tree: change.location.Tree})
	response, err := application.ports.RunProvider(ctx, task, change.configuration.MaxCommandOutputBytes, change.configuration.MaxCommandSeconds)
	if err != nil {
		return application.operationError(ctx, request.Command, "L7-PROVIDER-001", string(domain.StateBuilding), "implementer did not complete: "+bounded(err.Error(), 512), "resolve provider capability or process failure, then retry l7 run")
	}
	postProvider, contextBlocked := application.loadExecutionContext(ctx, request.Command)
	if contextBlocked != nil {
		return *contextBlocked
	}
	if !sameExecutionIntake(change, postProvider) {
		return application.failure("L7-GIT-003", request.Command, "recovery-required", "provider changed repository, configuration, or active identity", "inspect provider effects without resetting user work")
	}
	change = postProvider
	location, locateErr := application.ports.Locate(ctx, change.location.Root)
	pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if locateErr != nil || pendingErr != nil || !sameLocation(change.location, location) || !sameLocation(change.location, pending.RepositoryLocation) {
		return application.failure("L7-GIT-003", request.Command, "recovery-required", "provider changed Git identity or it cannot be reconstructed", "inspect Git history and worktree without resetting user work")
	}
	if pending.IndexDirty {
		return application.result(domain.OutcomeBlocked, "L7-GIT-004", string(request.Command), "recovery-required", "provider changed the Git index", "inspect and unstage the provider changes before retrying")
	}
	if len(pending.Paths) == 0 {
		return application.result(domain.OutcomeBlocked, "L7-RUN-002", string(request.Command), string(domain.StateBuilding), "provider completed without a scoped repository change", "clarify the task or inspect whether the change already exists")
	}
	postEvidence, evidenceErr := application.loadEvidenceState(change.location.CommonDir)
	if evidenceErr != nil || !sameEvidenceState(baselineEvidence, postEvidence) {
		return application.failure("L7-STATE-005", request.Command, "recovery-required", "provider changed Level 7 execution evidence", "inspect local evidence and provider effects before retrying")
	}
	approvalOK, approvalErr := application.approvalCurrent(ctx, change, request.Agent)
	if approvalErr != nil || !approvalOK {
		return application.result(domain.OutcomeBlocked, "L7-AUTH-001", string(request.Command), "recovery-required", "owner approval changed during provider execution", "restore current external owner approval before accepting provider work")
	}
	if response.Role != domain.RoleImplementer || response.Identity.Provider != request.Agent || response.Identity.Capability != domain.CapabilityAvailable {
		return application.failure("L7-PROVIDER-002", request.Command, string(domain.StateBuilding), "implementer returned an invalid provider identity", "inspect the provider adapter result before retrying")
	}
	expanded := domain.ExpandedPaths(change.active.Scope, pending.Paths, nil)
	if len(expanded) != 0 {
		result := application.result(domain.OutcomeBlocked, "L7-SCOPE-001", string(request.Command), string(domain.StateBuilding), "provider changed paths outside the declared scope", "restore the expanded paths or approve a revised change brief")
		result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, expanded)
		return result
	}
	runEvidence := domain.RunEvidence{
		ChangeID: change.active.ID, Provider: response.Identity,
		Parent:     domain.CandidateIdentity{Commit: change.location.Head, Tree: change.location.Tree},
		PathDigest: application.ports.PathSetDigest(pending.Paths), PathCount: len(pending.Paths), CommitMessage: request.CommitMessage,
	}
	if err := application.ports.SaveRun(change.location.CommonDir, runEvidence); err != nil {
		return application.failure("L7-STATE-005", request.Command, "recovery-required", "implementation finished but its commit transition was not recorded: "+bounded(err.Error(), 512), "preserve the worktree and repair run evidence before continuing")
	}
	result := application.commitRun(ctx, request, change, runEvidence, pending.Paths)
	if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
		result = application.failure("L7-STATE-003", request.Command, result.State, "run completed but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
	}
	released = true
	return result
}

func (application Application) verifyChange(ctx context.Context, request domain.Request) domain.Result {
	if !application.executionPortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-003", string(request.Command), "unavailable", "repository verification is not configured in this build", "run l7 help")
	}
	change, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil {
		return *blocked
	}
	if len(change.configuration.Verification) == 0 {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-004", string(request.Command), string(domain.StateBuilding), "repository verification is empty", "configure at least one explicit verification argv in .l7/config.json")
	}
	baselineEvidence, err := application.loadEvidenceState(change.location.CommonDir)
	if err != nil {
		return application.failure("L7-STATE-009", request.Command, "invalid", "execution evidence is invalid: "+bounded(err.Error(), 512), "repair the unsafe local evidence before continuing")
	}
	runEvidence, found := baselineEvidence.run, baselineEvidence.runFound
	if err != nil || !found || runEvidence.ChangeID != change.active.ID || runEvidence.Candidate.Commit == "" {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-005", string(request.Command), string(domain.StateBuilding), "no completed implementer binding exists for this change", "run l7 run --agent codex|claude --message <conventional-subject>")
	}
	runValid, validationErr := application.validRunEvidence(ctx, change, runEvidence)
	if validationErr != nil || !runValid {
		return application.failure("L7-STATE-005", request.Command, string(domain.StateBuilding), "implementer evidence does not match its Git commit", "inspect the candidate and rerun implementation before verification")
	}
	approvalOK, approvalErr := application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if approvalErr != nil || !approvalOK {
		return application.result(domain.OutcomeBlocked, "L7-AUTH-001", string(request.Command), string(domain.StateAwaitingOwnerApproval), "current owner approval does not bind the implementer and brief", "restore current external owner approval before verification")
	}
	existing, verificationFound := baselineEvidence.verification, baselineEvidence.verificationFound
	if verificationFound && existing.ChangeID == change.active.ID && existing.Candidate == runEvidence.Candidate {
		if change.active.Tier == domain.TierHighRisk && existing.VerificationCommit == "" {
			return application.withLockResumeVerification(ctx, request, change, baselineEvidence)
		}
		valid, validErr := application.validVerificationEvidence(ctx, change, runEvidence, existing)
		if validErr == nil && valid && verificationAtCurrent(change, existing) {
			return application.verificationResult(request, change, existing, "verification is already current")
		}
	}
	if change.location.Head != runEvidence.Candidate.Commit || change.location.Tree != runEvidence.Candidate.Tree {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-006", string(request.Command), string(domain.StateBuilding), "the implementer candidate is stale", "run l7 run for the current Git candidate before verification")
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(pending.Paths) != 0 || pending.IndexDirty || !sameLocation(change.location, pending.RepositoryLocation) {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-007", string(request.Command), string(domain.StateBuilding), "verification requires the exact clean implementer candidate", "commit or restore pending work, then retry l7 verify")
	}

	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then retry l7 verify")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	recheckedEvidence, evidenceErr := application.loadEvidenceState(change.location.CommonDir)
	pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	approvalOK, approvalErr = application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if blocked != nil || !sameExecutionContext(change, rechecked) || evidenceErr != nil || !sameEvidenceState(baselineEvidence, recheckedEvidence) || pendingErr != nil || len(pending.Paths) != 0 || pending.IndexDirty || !sameLocation(change.location, pending.RepositoryLocation) || approvalErr != nil || !approvalOK {
		return application.failure("L7-GIT-002", request.Command, "changed", "candidate changed before verification", "retry l7 verify against a stable candidate")
	}
	change = rechecked
	checks, runErr := application.ports.RunVerification(ctx, change.location.Root, change.configuration.Verification, change.configuration.MaxCommandOutputBytes, change.configuration.MaxCommandSeconds)
	if runErr != nil && ctx.Err() != nil {
		return application.operationError(ctx, request.Command, "L7-VERIFY-001", string(domain.StateBuilding), "repository verification failed: "+bounded(runErr.Error(), 512), "run l7 status to reconstruct the accepted state")
	}
	postVerification, contextBlocked := application.loadExecutionContext(ctx, request.Command)
	location, locateErr := application.ports.Locate(ctx, change.location.Root)
	pending, pendingErr = application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	postEvidence, evidenceErr := application.loadEvidenceState(change.location.CommonDir)
	approvalOK, approvalErr = application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if contextBlocked != nil || !sameExecutionContext(change, postVerification) || locateErr != nil || pendingErr != nil || !sameLocation(change.location, location) || !sameLocation(change.location, pending.RepositoryLocation) || len(pending.Paths) != 0 || pending.IndexDirty || evidenceErr != nil || !sameEvidenceState(baselineEvidence, postEvidence) || approvalErr != nil || !approvalOK {
		return application.failure("L7-VERIFY-008", request.Command, string(domain.StateBuilding), "candidate changed during verification", "inspect the worktree and rerun verification against a stable candidate")
	}
	change = postVerification
	if runErr != nil {
		result := application.operationError(ctx, request.Command, "L7-VERIFY-001", string(domain.StateBuilding), "repository verification failed: "+bounded(runErr.Error(), 512), "fix the failing check, commit remediation with l7 run, then retry l7 verify")
		result.Execution = &domain.ExecutionDetails{Role: domain.RoleImplementer, Provider: runEvidence.Provider.Provider, Commit: runEvidence.Candidate.Commit, Tree: runEvidence.Candidate.Tree, Checks: checks}
		return result
	}
	evidence := domain.VerificationEvidence{ChangeID: change.active.ID, Candidate: runEvidence.Candidate, Result: domain.DecisionGO, Checks: checks, ConfigurationDigest: change.configuration.Digest}
	if err := application.ports.SaveVerification(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-006", request.Command, "recovery-required", "passing verification was not recorded: "+bounded(err.Error(), 512), "preserve the candidate and repair verification evidence before continuing")
	}
	result := application.verificationResult(request, change, evidence, "repository verification passed")
	if change.active.Tier == domain.TierHighRisk {
		result = application.finishVerificationArtifact(ctx, request, change, evidence)
	}
	if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
		result = application.failure("L7-STATE-003", request.Command, result.State, "verification completed but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
	}
	released = true
	return result
}

func (application Application) reviewChange(ctx context.Context, request domain.Request) domain.Result {
	if !application.executionPortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-003", string(request.Command), "unavailable", "provider review is not configured in this build", "run l7 help")
	}
	if !request.Agent.Valid() {
		return application.Invalid(string(request.Command), "review requires --agent codex|claude")
	}
	change, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil {
		return *blocked
	}
	baselineEvidence, err := application.loadEvidenceState(change.location.CommonDir)
	if err != nil {
		return application.failure("L7-STATE-009", request.Command, "invalid", "execution evidence is invalid: "+bounded(err.Error(), 512), "repair the unsafe local evidence before continuing")
	}
	runEvidence, runFound := baselineEvidence.run, baselineEvidence.runFound
	if !runFound || runEvidence.ChangeID != change.active.ID || runEvidence.Candidate.Commit == "" {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-001", string(request.Command), string(domain.StateBuilding), "review requires a completed implementer binding", "run l7 run and l7 verify before review")
	}
	verification, verificationFound := baselineEvidence.verification, baselineEvidence.verificationFound
	if !verificationFound || verification.ChangeID != change.active.ID {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-002", string(request.Command), string(domain.StateBuilding), "review requires current verification", "run l7 verify against the exact implementer candidate")
	}
	verificationValid, validationErr := application.validVerificationEvidence(ctx, change, runEvidence, verification)
	if validationErr != nil || !verificationValid || !verificationAtCurrent(change, verification) {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-002", string(request.Command), string(domain.StateBuilding), "verification is stale or does not match Git", "run l7 verify against the current candidate")
	}
	approvalOK, approvalErr := application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if approvalErr != nil || !approvalOK {
		return application.result(domain.OutcomeBlocked, "L7-AUTH-001", string(request.Command), string(domain.StateAwaitingOwnerApproval), "current owner approval does not bind the implementer and brief", "restore current external owner approval before review")
	}
	if change.active.Tier == domain.TierHighRisk && request.Agent == runEvidence.Provider.Provider {
		other, _ := domain.OtherProvider(runEvidence.Provider.Provider)
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-003", string(request.Command), string(domain.StateAwaitingIndependentAudit), "Tier 3 reviewer must use the other provider", "run l7 review --agent "+string(other))
	}
	existing, reviewFound := baselineEvidence.review, baselineEvidence.reviewFound
	if reviewFound && existing.ChangeID == change.active.ID && existing.Candidate == verification.Candidate {
		if change.active.Tier == domain.TierHighRisk && existing.ReviewCommit == "" {
			if existing.Provider.Provider != request.Agent {
				return application.result(domain.OutcomeBlocked, "L7-REVIEW-004", string(request.Command), "recovery-required", "an interrupted audit is bound to another provider", "rerun l7 review with the original --agent")
			}
			return application.withLockResumeReview(ctx, request, change, baselineEvidence)
		}
		valid, validErr := application.validReviewEvidence(ctx, change, runEvidence, verification, existing)
		if validErr == nil && valid && reviewAtCurrent(change, existing) {
			return application.reviewResult(request, change, existing, "review is already current")
		}
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(pending.Paths) != 0 || pending.IndexDirty || !sameLocation(change.location, pending.RepositoryLocation) {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-005", string(request.Command), string(domain.StateBuilding), "review requires a clean verified candidate", "commit or restore pending work, then rerun verification")
	}

	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then retry l7 review")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	recheckedEvidence, evidenceErr := application.loadEvidenceState(change.location.CommonDir)
	pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	approvalOK, approvalErr = application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if blocked != nil || !sameExecutionContext(change, rechecked) || evidenceErr != nil || !sameEvidenceState(baselineEvidence, recheckedEvidence) || pendingErr != nil || len(pending.Paths) != 0 || pending.IndexDirty || !sameLocation(change.location, pending.RepositoryLocation) || approvalErr != nil || !approvalOK {
		return application.failure("L7-GIT-002", request.Command, "changed", "candidate changed before review", "rerun verification before review")
	}
	change = rechecked
	task := providerTask(change, request.Agent, domain.RoleReviewer, verification.Candidate)
	response, err := application.ports.RunProvider(ctx, task, change.configuration.MaxCommandOutputBytes, change.configuration.MaxCommandSeconds)
	if err != nil {
		return application.operationError(ctx, request.Command, "L7-PROVIDER-001", string(domain.StateAwaitingIndependentAudit), "reviewer did not complete: "+bounded(err.Error(), 512), "resolve provider capability or process failure, then retry l7 review")
	}
	postReview, contextBlocked := application.loadExecutionContext(ctx, request.Command)
	location, locateErr := application.ports.Locate(ctx, change.location.Root)
	pending, pendingErr = application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	postEvidence, evidenceErr := application.loadEvidenceState(change.location.CommonDir)
	approvalOK, approvalErr = application.approvalCurrent(ctx, change, runEvidence.Provider.Provider)
	if contextBlocked != nil || !sameExecutionContext(change, postReview) || locateErr != nil || pendingErr != nil || !sameLocation(change.location, location) || !sameLocation(change.location, pending.RepositoryLocation) || len(pending.Paths) != 0 || pending.IndexDirty || evidenceErr != nil || !sameEvidenceState(baselineEvidence, postEvidence) || approvalErr != nil || !approvalOK {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-007", string(request.Command), string(domain.StateBuilding), "read-only reviewer mutated the candidate", "inspect and restore reviewer mutations, then rerun verification")
	}
	change = postReview
	if response.Role != domain.RoleReviewer || response.Identity.Provider != request.Agent || response.Identity.Capability != domain.CapabilityAvailable || !response.Decision.Valid() || !domain.DistinctReviewer(change.active.Tier, runEvidence.Provider, response.Identity) {
		return application.failure("L7-REVIEW-006", request.Command, string(domain.StateBuilding), "reviewer identity or decision is invalid", "choose an eligible independent reviewer and retry")
	}
	evidence := domain.ReviewEvidence{ChangeID: change.active.ID, Provider: response.Identity, Candidate: verification.Candidate, Decision: response.Decision, Findings: append([]string{}, response.Findings...)}
	if err := application.ports.SaveReview(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-007", request.Command, "recovery-required", "review completed but was not recorded: "+bounded(err.Error(), 512), "preserve the candidate and repair review evidence before continuing")
	}
	result := application.reviewResult(request, change, evidence, "review completed")
	if change.active.Tier == domain.TierHighRisk {
		result = application.finishReviewArtifact(ctx, request, change, verification, evidence)
	}
	if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
		result = application.failure("L7-STATE-003", request.Command, result.State, "review completed but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
	}
	released = true
	return result
}

func (application Application) loadExecutionContext(ctx context.Context, command domain.Command) (executionContext, *domain.Result) {
	var change executionContext
	location, err := application.ports.Locate(ctx, application.cwd)
	if err != nil {
		result := application.failure("L7-REPO-001", command, "unavailable", "cannot inspect repository: "+bounded(err.Error(), 512), "run l7 "+string(command)+" from an adopted Git worktree")
		return change, &result
	}
	configuration, found, err := application.ports.LoadConfiguration(location.Root)
	if err != nil {
		result := application.failure("L7-CONFIG-001", command, "invalid", "repository configuration is invalid: "+bounded(err.Error(), 512), "repair .l7/config.json before continuing")
		return change, &result
	}
	if !found {
		result := application.result(domain.OutcomeBlocked, "L7-CONFIG-002", string(command), "unadopted", "repository has not been adopted", "run l7 adopt --enable-local-lifecycle")
		return change, &result
	}
	if !configuration.LocalLifecycle {
		result := application.result(domain.OutcomeBlocked, "L7-FLAG-001", string(command), "disabled", "local lifecycle behavior is default OFF", "run l7 adopt --enable-local-lifecycle")
		return change, &result
	}
	active, activeFound, err := application.ports.LoadActive(location.CommonDir)
	if err != nil {
		result := application.failure("L7-STATE-001", command, "invalid", "active context is invalid: "+bounded(err.Error(), 512), "repair or explicitly recover .git/l7/product/active.json")
		return change, &result
	}
	if !activeFound {
		result := application.result(domain.OutcomeBlocked, "L7-STATE-008", string(command), "idle", "no active Level 7 change exists", "run l7 brief with an explicit risk tier and scope")
		return change, &result
	}
	brief := domain.ChangeBrief{}
	storedActive := active
	if active.Kind == domain.ActiveBrief {
		brief, err = application.ports.LoadBrief(location.Root, active.BriefPath)
		if err != nil || brief.ID != active.ID {
			result := application.failure("L7-BRIEF-003", command, "invalid", "active change brief is missing, unsafe, or conflicting", "restore the exact tracked change brief before continuing")
			return change, &result
		}
		active.Tier = brief.Tier
		active.Base = brief.Base
		active.Problem = brief.Problem
		active.Scope = append([]string{}, brief.Scope...)
	}
	if active.Tier != domain.TierHighRisk && touchesProtected(active.Scope, configuration.ProtectedPaths) {
		result := application.result(domain.OutcomeBlocked, "L7-RISK-001", string(command), "risk-mismatch", "active scope intersects a protected control without Tier 3 classification", "replace the active change with an explicitly approved Tier 3 brief")
		return change, &result
	}
	snapshot, err := application.ports.Snapshot(ctx, location.Root, active.Base, configuration.MaxGitOutputBytes, configuration.MaxGitPaths)
	if err != nil {
		result := application.failure("L7-GIT-001", command, "invalid", "cannot reconstruct Git-derived change: "+bounded(err.Error(), 512), "restore an ancestor base and stable Git worktree")
		return change, &result
	}
	if !repositoryOutputFits(configuration.MaxCommandOutputBytes, snapshot, active.Scope) {
		result := application.result(domain.OutcomeBlocked, "L7-OUTPUT-001", string(command), "bounded", "Git-derived change status exceeds the configured command-output limit", "narrow the active scope or explicitly raise the bounded output limit")
		return change, &result
	}
	permitted := []string{}
	if active.BriefPath != "" {
		permitted = append(permitted, active.BriefPath)
	}
	if active.Tier == domain.TierHighRisk {
		permitted = append(permitted, verificationPath(active.ID), auditPath(active.ID))
	}
	expanded := domain.ExpandedPaths(active.Scope, snapshot.ChangedPaths, permitted)
	if len(expanded) != 0 {
		result := application.result(domain.OutcomeBlocked, "L7-SCOPE-001", string(command), string(domain.StateBuilding), "changed paths exceed the declared scope", "restore the expanded paths or begin a new appropriately scoped change")
		result.Repository = detailsForSnapshot(snapshot, active.ID, active.Tier, active.Scope, expanded)
		return change, &result
	}
	recheckedConfiguration, configFound, configErr := application.ports.LoadConfiguration(location.Root)
	recheckedActive, recheckedFound, activeErr := application.ports.LoadActive(location.CommonDir)
	if configErr != nil || !configFound || activeErr != nil || !recheckedFound || !sameConfiguration(configuration, recheckedConfiguration) || !sameStoredActive(active, recheckedActive) {
		result := application.failure("L7-STATE-004", command, "changed", "configuration or active context changed during reconstruction", "retry against stable local state")
		return change, &result
	}
	if storedActive.Kind == domain.ActiveBrief {
		recheckedBrief, briefErr := application.ports.LoadBrief(location.Root, storedActive.BriefPath)
		if briefErr != nil || !sameBrief(brief, recheckedBrief) {
			result := application.failure("L7-BRIEF-004", command, "changed", "change brief changed during reconstruction", "retry against a stable change brief")
			return change, &result
		}
	}
	change = executionContext{location: location, configuration: configuration, active: active, brief: brief, snapshot: snapshot}
	return change, nil
}

func (application Application) statusFromContext(ctx context.Context, request domain.Request, change executionContext) domain.Result {
	view, err := application.loadEvidenceView(ctx, change)
	if err != nil {
		result := application.failure("L7-STATE-009", request.Command, "invalid", "execution evidence cannot be reconstructed: "+bounded(err.Error(), 512), "repair the unsafe local evidence before continuing")
		result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
		return result
	}
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil || !sameExecutionContext(change, rechecked) {
		return application.failure("L7-GIT-002", request.Command, "changed", "repository changed during execution-state reconstruction", "retry l7 status against a stable candidate")
	}
	pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if pendingErr != nil || !sameLocation(change.location, pending.RepositoryLocation) {
		return application.failure("L7-GIT-001", request.Command, "invalid", "cannot establish exact pending-candidate state", "stabilize the Git worktree, then retry l7 status")
	}
	candidateClean := len(pending.Paths) == 0 && !pending.IndexDirty
	workStarted := false
	for _, changed := range change.snapshot.ChangedPaths {
		if domain.ScopeContains(change.active.Scope, changed) {
			workStarted = true
			break
		}
	}
	runCurrent := candidateClean && view.runValid && runAtCurrentChain(change, view)
	verificationCurrent := candidateClean && view.verificationValid && verificationAtCurrentChain(change, view)
	reviewCurrent := candidateClean && view.reviewValid && reviewAtCurrent(change, view.review)
	ownerCurrent := false
	if change.active.Tier == domain.TierHighRisk {
		briefCommit, briefErr := application.ports.PathCommit(ctx, change.location.Root, change.active.BriefPath)
		approval, approvalFound, approvalErr := application.ports.LoadApproval(change.location.CommonDir)
		if briefErr == nil && approvalErr == nil && approvalFound {
			expected := approval.Implementer
			if view.runFound && view.run.ChangeID == change.active.ID && view.run.Provider.Provider.Valid() {
				expected = view.run.Provider.Provider
			}
			ownerCurrent = approvalMatches(approval, change.active.ID, expected, briefCommit)
		}
	}
	readyCurrent := false
	readinessFound := false
	readinessEvidence := domain.ReadinessEvidence{}
	if application.ports.LoadReadiness != nil {
		readiness, found, readinessErr := application.ports.LoadReadiness(change.location.CommonDir)
		if readinessErr != nil {
			result := application.failure("L7-STATE-010", request.Command, "invalid", "readiness evidence cannot be reconstructed: "+bounded(readinessErr.Error(), 512), "repair the unsafe readiness evidence before continuing")
			result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
			return result
		}
		readinessFound = found
		if found {
			readinessEvidence = readiness
			currentFacts, factsErr := application.localReadinessFacts(ctx, change, view)
			readyCurrent = factsErr == nil && domain.EvaluateReadiness(currentFacts).Ready && sameReadinessEvidence(readiness, currentFacts.Evidence)
		}
	}
	mergedCurrent := false
	mergeFound := false
	mergeReceipt := domain.MergeReceipt{}
	if application.ports.LoadMerge != nil {
		merge, found, mergeErr := application.ports.LoadMerge(change.location.CommonDir)
		if mergeErr != nil {
			result := application.failure("L7-STATE-011", request.Command, "invalid", "merge receipt cannot be reconstructed: "+bounded(mergeErr.Error(), 512), "repair the unsafe merge receipt before continuing")
			result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
			return result
		}
		mergeFound, mergeReceipt = found, merge
		if found && readyCurrent && application.ports.MergeCurrent != nil && mergeReceiptMatchesReadiness(merge, readinessEvidence) {
			mergedCurrent, mergeErr = application.ports.MergeCurrent(ctx, change.location.Root, merge, change.configuration.MaxGitOutputBytes)
			if mergeErr != nil {
				result := application.failure("L7-MERGE-004", request.Command, "recovery-required", "merged target cannot be reconstructed: "+bounded(mergeErr.Error(), 512), "inspect Git and the merge receipt before continuing")
				result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
				return result
			}
		}
	}
	if change.active.Tier == domain.TierHighRisk && workStarted && !ownerCurrent {
		result := application.result(domain.OutcomeBlocked, "L7-AUTH-001", string(request.Command), string(domain.StateAwaitingOwnerApproval), "Tier 3 implementation exists without a current external owner-approval binding", "restore unapproved work or rerun l7 run from an interactive terminal for the exact approved brief")
		result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
		return result
	}
	stale := (view.runFound && !runCurrent) || (view.verificationFound && !verificationCurrent) || (view.reviewFound && !reviewCurrent) || (readinessFound && !readyCurrent) || (mergeFound && !mergedCurrent)
	rejected := reviewCurrent && view.review.Decision == domain.DecisionNoGO
	facts := domain.LifecycleFacts{
		Tier: change.active.Tier, PlanPresent: true, OwnerApprovalCurrent: ownerCurrent, WorkStarted: workStarted,
		VerificationCurrent: verificationCurrent, ReadyCurrent: readyCurrent, MergedCurrent: mergedCurrent, AssuranceRejected: rejected, AssuranceStale: stale && workStarted,
	}
	if change.active.Tier == domain.TierHighRisk {
		facts.IndependentAuditCurrent = reviewCurrent && view.review.Decision == domain.DecisionGO
	} else {
		facts.ReviewCurrent = reviewCurrent && view.review.Decision == domain.DecisionGO
	}
	state, valid := domain.DeriveLifecycle(facts)
	if !valid {
		result := application.failure("L7-LIFECYCLE-001", request.Command, "invalid", "active lifecycle facts conflict", "restore the last valid Git and local-state combination")
		result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
		return result
	}
	next, _ := domain.NextTransition(change.active.Tier, state)
	outcome := domain.OutcomePass
	code := "L7-STATUS-000"
	message := "Git-derived lifecycle state is current"
	switch state {
	case domain.StateAwaitingOwnerApproval:
		outcome, code, message = domain.OutcomeBlocked, "L7-AUTH-002", "Tier 3 is awaiting explicit external owner approval"
		next.Action = "run l7 run --agent codex|claude --message <conventional-subject> from an interactive terminal"
	case domain.StateBuilding:
		outcome, code = domain.OutcomeBlocked, "L7-BUILD-001"
		message = "implementation or assurance is incomplete or stale"
		next.Action = "run l7 run --agent codex|claude --message <conventional-subject>"
		if runCurrent && !verificationCurrent {
			next.Action = "run l7 verify"
		}
	case domain.StateVerified:
		next.Action = "run l7 review --agent codex|claude"
	case domain.StateAwaitingIndependentAudit:
		outcome, code, message = domain.OutcomeBlocked, "L7-REVIEW-002", "Tier 3 is awaiting the other provider's read-only audit"
		other, ok := domain.OtherProvider(view.run.Provider.Provider)
		if ok {
			next.Action = "run l7 review --agent " + string(other)
		}
	case domain.StateReviewed:
		message = "review is current; exact candidate readiness has not been recorded"
		next.Action = "run l7 ready"
	case domain.StateReady:
		message = "exact candidate readiness is current"
		next.Action = "run l7 merge --target <branch> and confirm the full candidate SHA"
	case domain.StateMerged:
		message = "local target contains the exact ready candidate"
		next.Action = "run l7 status to inspect the merged local ref"
	}
	result := application.result(outcome, code, string(request.Command), string(state), message, next.Action)
	result.Repository = detailsForSnapshot(change.snapshot, change.active.ID, change.active.Tier, change.active.Scope, nil)
	result.Execution = executionDetails(view)
	if readyCurrent {
		result.Readiness = readinessDetails(readinessEvidence, false, true)
		if mergedCurrent {
			result.Readiness.TargetRef = mergeReceipt.TargetRef
			result.Readiness.PreviousCommit = mergeReceipt.PreviousCommit
		}
	}
	return result
}

func (application Application) loadEvidenceView(ctx context.Context, change executionContext) (evidenceView, error) {
	view, err := application.loadEvidenceState(change.location.CommonDir)
	if err != nil {
		return view, err
	}
	if view.runFound && view.run.ChangeID == change.active.ID && view.run.Candidate.Commit != "" {
		view.runValid, err = application.validRunEvidence(ctx, change, view.run)
		if err != nil {
			return view, err
		}
	}
	if view.runValid && view.verificationFound && view.verification.ChangeID == change.active.ID {
		view.verificationValid, err = application.validVerificationEvidence(ctx, change, view.run, view.verification)
		if err != nil {
			return view, err
		}
	}
	if view.runValid && view.verificationValid && view.reviewFound && view.review.ChangeID == change.active.ID {
		view.reviewValid, err = application.validReviewEvidence(ctx, change, view.run, view.verification, view.review)
		if err != nil {
			return view, err
		}
	}
	return view, nil
}

func (application Application) loadEvidenceState(commonDirectory string) (evidenceView, error) {
	var view evidenceView
	var err error
	view.run, view.runFound, err = application.ports.LoadRun(commonDirectory)
	if err != nil {
		return view, err
	}
	view.verification, view.verificationFound, err = application.ports.LoadVerification(commonDirectory)
	if err != nil {
		return view, err
	}
	view.review, view.reviewFound, err = application.ports.LoadReview(commonDirectory)
	return view, err
}

func (application Application) validRunEvidence(ctx context.Context, change executionContext, evidence domain.RunEvidence) (bool, error) {
	if evidence.ChangeID != change.active.ID || evidence.Candidate.Commit == "" {
		return false, nil
	}
	parentTree, err := application.ports.CommitTree(ctx, change.location.Root, evidence.Parent.Commit)
	if err != nil {
		return false, err
	}
	candidateTree, err := application.ports.CommitTree(ctx, change.location.Root, evidence.Candidate.Commit)
	if err != nil {
		return false, err
	}
	matches, err := application.ports.CommitMatches(ctx, change.location.Root, evidence.Candidate.Commit, evidence.Parent.Commit, evidence.CommitMessage, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || !matches || parentTree != evidence.Parent.Tree || candidateTree != evidence.Candidate.Tree {
		return false, err
	}
	paths, err := application.ports.CommitPaths(ctx, change.location.Root, evidence.Parent.Commit, evidence.Candidate.Commit, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil {
		return false, err
	}
	return len(paths) == evidence.PathCount && application.ports.PathSetDigest(paths) == evidence.PathDigest && len(domain.ExpandedPaths(change.active.Scope, paths, nil)) == 0, nil
}

func (application Application) validVerificationEvidence(ctx context.Context, change executionContext, run domain.RunEvidence, evidence domain.VerificationEvidence) (bool, error) {
	if evidence.ChangeID != change.active.ID || evidence.Candidate != run.Candidate || evidence.Result != domain.DecisionGO || evidence.ConfigurationDigest == "" || evidence.ConfigurationDigest != change.configuration.Digest || !checksMatchConfiguration(evidence.Checks, change.configuration.Verification) {
		return false, nil
	}
	if change.active.Tier != domain.TierHighRisk {
		return evidence.VerificationCommit == "" && evidence.VerificationTree == "", nil
	}
	if evidence.VerificationCommit == "" || evidence.VerificationTree == "" {
		return false, nil
	}
	matches, err := application.ports.CommitMatches(ctx, change.location.Root, evidence.VerificationCommit, run.Candidate.Commit, verificationSubject(change.active.ID), change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || !matches {
		return false, err
	}
	paths, err := application.ports.CommitPaths(ctx, change.location.Root, run.Candidate.Commit, evidence.VerificationCommit, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(paths) != 1 || paths[0] != verificationPath(change.active.ID) {
		return false, err
	}
	tree, err := application.ports.CommitTree(ctx, change.location.Root, evidence.VerificationCommit)
	return err == nil && tree == evidence.VerificationTree, err
}

func (application Application) validReviewEvidence(ctx context.Context, change executionContext, run domain.RunEvidence, verification domain.VerificationEvidence, evidence domain.ReviewEvidence) (bool, error) {
	if evidence.ChangeID != change.active.ID || evidence.Candidate != verification.Candidate || !evidence.Decision.Valid() || !domain.DistinctReviewer(change.active.Tier, run.Provider, evidence.Provider) {
		return false, nil
	}
	if change.active.Tier != domain.TierHighRisk {
		return evidence.ReviewCommit == "" && evidence.ReviewTree == "", nil
	}
	if evidence.ReviewCommit == "" || evidence.ReviewTree == "" || verification.VerificationCommit == "" {
		return false, nil
	}
	matches, err := application.ports.CommitMatches(ctx, change.location.Root, evidence.ReviewCommit, verification.VerificationCommit, auditSubject(change.active.ID), change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || !matches {
		return false, err
	}
	paths, err := application.ports.CommitPaths(ctx, change.location.Root, verification.VerificationCommit, evidence.ReviewCommit, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(paths) != 1 || paths[0] != auditPath(change.active.ID) {
		return false, err
	}
	tree, err := application.ports.CommitTree(ctx, change.location.Root, evidence.ReviewCommit)
	return err == nil && tree == evidence.ReviewTree, err
}

func (application Application) resumeRun(ctx context.Context, request domain.Request, change executionContext, evidence domain.RunEvidence) domain.Result {
	if change.location.Head == evidence.Parent.Commit && change.location.Tree == evidence.Parent.Tree {
		pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		if err != nil || pending.IndexDirty || len(pending.Paths) != evidence.PathCount || application.ports.PathSetDigest(pending.Paths) != evidence.PathDigest || len(domain.ExpandedPaths(change.active.Scope, pending.Paths, nil)) != 0 {
			return application.result(domain.OutcomeBlocked, "L7-RUN-003", string(request.Command), "recovery-required", "interrupted provider changes no longer match their recorded path set", "inspect the worktree and preserve user changes before explicit recovery")
		}
		return application.commitRun(ctx, request, change, evidence, pending.Paths)
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(pending.Paths) != 0 || pending.IndexDirty {
		return application.result(domain.OutcomeBlocked, "L7-RUN-003", string(request.Command), "recovery-required", "interrupted commit recovery requires a clean worktree", "inspect Git and preserve user changes before explicit recovery")
	}
	matches, matchErr := application.ports.CommitMatches(ctx, change.location.Root, change.location.Head, evidence.Parent.Commit, evidence.CommitMessage, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	paths, pathErr := application.ports.CommitPaths(ctx, change.location.Root, evidence.Parent.Commit, change.location.Head, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if matchErr != nil || pathErr != nil || !matches || len(paths) != evidence.PathCount || application.ports.PathSetDigest(paths) != evidence.PathDigest {
		return application.result(domain.OutcomeBlocked, "L7-RUN-003", string(request.Command), "recovery-required", "current HEAD is not the interrupted controlled commit", "inspect Git history before explicit recovery")
	}
	evidence.Candidate = domain.CandidateIdentity{Commit: change.location.Head, Tree: change.location.Tree}
	if err := application.ports.SaveRun(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-005", request.Command, "recovery-required", "controlled commit exists but run evidence was not finalized", "repair run evidence, then run l7 status")
	}
	return application.runResult(request, change, evidence, "interrupted controlled commit recovered")
}

func (application Application) commitRun(ctx context.Context, request domain.Request, change executionContext, evidence domain.RunEvidence, paths []string) domain.Result {
	location, err := application.ports.Commit(ctx, domain.CommitRequest{
		Root: change.location.Root, ExpectedCommit: evidence.Parent.Commit, ExpectedTree: evidence.Parent.Tree,
		Paths: append([]string{}, paths...), Message: evidence.CommitMessage,
		MaxOutputBytes: change.configuration.MaxGitOutputBytes, MaxPaths: change.configuration.MaxGitPaths, MaxCommandSeconds: change.configuration.MaxCommandSeconds,
	})
	if err != nil {
		return application.operationError(ctx, request.Command, "L7-COMMIT-001", "recovery-required", "controlled implementation commit failed: "+bounded(err.Error(), 512), "inspect the index and worktree; preserve user changes before retrying")
	}
	evidence.Candidate = domain.CandidateIdentity{Commit: location.Head, Tree: location.Tree}
	if err := application.ports.SaveRun(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-005", request.Command, "recovery-required", "commit succeeded but run evidence was not finalized: "+bounded(err.Error(), 512), "rerun the identical l7 run command to recover the controlled commit")
	}
	change.location = location
	return application.runResult(request, change, evidence, "implementation completed in one controlled Git commit")
}

func (application Application) withLockResumeVerification(ctx context.Context, request domain.Request, change executionContext, baseline evidenceView) domain.Result {
	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait, then retry l7 verify")
	}
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	recheckedEvidence, loadErr := application.loadEvidenceState(change.location.CommonDir)
	approvalOK, approvalErr := application.approvalCurrent(ctx, change, baseline.run.Provider.Provider)
	if blocked != nil || !sameExecutionContext(change, rechecked) || loadErr != nil || !sameEvidenceState(baseline, recheckedEvidence) || approvalErr != nil || !approvalOK {
		_ = release()
		return application.failure("L7-GIT-002", request.Command, "changed", "candidate or verification evidence changed while acquiring the mutation lock", "retry l7 verify after reconstructing current state")
	}
	change, baseline = rechecked, recheckedEvidence
	evidence := baseline.verification
	result := application.finishVerificationArtifact(ctx, request, change, evidence)
	if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
		return application.failure("L7-STATE-003", request.Command, result.State, "verification recovered but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
	}
	return result
}

func (application Application) finishVerificationArtifact(ctx context.Context, request domain.Request, change executionContext, evidence domain.VerificationEvidence) domain.Result {
	if change.location.Head != evidence.Candidate.Commit {
		pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		matches, matchErr := application.ports.CommitMatches(ctx, change.location.Root, change.location.Head, evidence.Candidate.Commit, verificationSubject(change.active.ID), change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		paths, pathErr := application.ports.CommitPaths(ctx, change.location.Root, evidence.Candidate.Commit, change.location.Head, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		if pendingErr != nil || matchErr != nil || pathErr != nil || len(pending.Paths) != 0 || pending.IndexDirty || !matches || len(paths) != 1 || paths[0] != verificationPath(change.active.ID) {
			return application.result(domain.OutcomeBlocked, "L7-VERIFY-009", string(request.Command), "recovery-required", "current HEAD is not the interrupted verification-record commit", "inspect Git history and the verification artifact before explicit recovery")
		}
		evidence.VerificationCommit, evidence.VerificationTree = change.location.Head, change.location.Tree
		if err := application.ports.SaveVerification(change.location.CommonDir, evidence); err != nil {
			return application.failure("L7-STATE-006", request.Command, "recovery-required", "verification commit exists but evidence was not finalized", "repair verification evidence, then run l7 status")
		}
		return application.verificationResult(request, change, evidence, "interrupted verification-record commit recovered")
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(pending.Paths) != 0 || pending.IndexDirty {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-009", string(request.Command), "recovery-required", "verification-record commit requires a clean candidate", "restore unrelated work before retrying l7 verify")
	}
	path, err := application.ports.WriteVerification(change.location.Root, evidence, "local-verifier")
	if err != nil || path != verificationPath(change.active.ID) {
		return application.failure("L7-VERIFY-010", request.Command, "recovery-required", "verification artifact could not be written", "inspect the bounded verification artifact path before retrying")
	}
	pending, err = application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || pending.IndexDirty || len(pending.Paths) != 1 || pending.Paths[0] != path {
		return application.result(domain.OutcomeBlocked, "L7-VERIFY-010", string(request.Command), "recovery-required", "verification artifact did not produce one exact pending path", "inspect the worktree and preserve unrelated changes")
	}
	location, err := application.ports.Commit(ctx, domain.CommitRequest{
		Root: change.location.Root, ExpectedCommit: change.location.Head, ExpectedTree: change.location.Tree,
		Paths: []string{path}, Message: verificationSubject(change.active.ID), MaxOutputBytes: change.configuration.MaxGitOutputBytes,
		MaxPaths: change.configuration.MaxGitPaths, MaxCommandSeconds: change.configuration.MaxCommandSeconds,
	})
	if err != nil {
		return application.operationError(ctx, request.Command, "L7-COMMIT-002", "recovery-required", "verification-record commit failed: "+bounded(err.Error(), 512), "inspect the index and rerun l7 verify for recovery")
	}
	evidence.VerificationCommit, evidence.VerificationTree = location.Head, location.Tree
	if err := application.ports.SaveVerification(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-006", request.Command, "recovery-required", "verification record committed but evidence was not finalized", "rerun l7 verify to recover the exact commit")
	}
	change.location = location
	return application.verificationResult(request, change, evidence, "verification passed and its sole Tier 3 record was committed")
}

func (application Application) withLockResumeReview(ctx context.Context, request domain.Request, change executionContext, baseline evidenceView) domain.Result {
	release, err := application.ports.Acquire(change.location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait, then retry l7 review")
	}
	rechecked, blocked := application.loadExecutionContext(ctx, request.Command)
	recheckedEvidence, loadErr := application.loadEvidenceState(change.location.CommonDir)
	approvalOK, approvalErr := application.approvalCurrent(ctx, change, baseline.run.Provider.Provider)
	if blocked != nil || !sameExecutionContext(change, rechecked) || loadErr != nil || !sameEvidenceState(baseline, recheckedEvidence) || approvalErr != nil || !approvalOK {
		_ = release()
		return application.failure("L7-GIT-002", request.Command, "changed", "candidate or review evidence changed while acquiring the mutation lock", "retry l7 review after reconstructing current state")
	}
	change, baseline = rechecked, recheckedEvidence
	verification, evidence := baseline.verification, baseline.review
	result := application.finishReviewArtifact(ctx, request, change, verification, evidence)
	if releaseErr := release(); releaseErr != nil && result.Outcome == domain.OutcomePass {
		return application.failure("L7-STATE-003", request.Command, result.State, "review recovered but lock release reported: "+bounded(releaseErr.Error(), 512), "run l7 status before continuing")
	}
	return result
}

func (application Application) finishReviewArtifact(ctx context.Context, request domain.Request, change executionContext, verification domain.VerificationEvidence, evidence domain.ReviewEvidence) domain.Result {
	parent := verification.VerificationCommit
	if change.location.Head != parent {
		pending, pendingErr := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		matches, matchErr := application.ports.CommitMatches(ctx, change.location.Root, change.location.Head, parent, auditSubject(change.active.ID), change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		paths, pathErr := application.ports.CommitPaths(ctx, change.location.Root, parent, change.location.Head, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
		if pendingErr != nil || matchErr != nil || pathErr != nil || len(pending.Paths) != 0 || pending.IndexDirty || !matches || len(paths) != 1 || paths[0] != auditPath(change.active.ID) {
			return application.result(domain.OutcomeBlocked, "L7-REVIEW-008", string(request.Command), "recovery-required", "current HEAD is not the interrupted audit-record commit", "inspect Git history and the audit artifact before explicit recovery")
		}
		evidence.ReviewCommit, evidence.ReviewTree = change.location.Head, change.location.Tree
		if err := application.ports.SaveReview(change.location.CommonDir, evidence); err != nil {
			return application.failure("L7-STATE-007", request.Command, "recovery-required", "audit commit exists but review evidence was not finalized", "repair review evidence, then run l7 status")
		}
		return application.reviewResult(request, change, evidence, "interrupted audit-record commit recovered")
	}
	pending, err := application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || len(pending.Paths) != 0 || pending.IndexDirty {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-008", string(request.Command), "recovery-required", "audit-record commit requires a clean verified candidate", "restore unrelated work before retrying l7 review")
	}
	path, err := application.ports.WriteAudit(change.location.Root, evidence)
	if err != nil || path != auditPath(change.active.ID) {
		return application.failure("L7-REVIEW-009", request.Command, "recovery-required", "audit artifact could not be written", "inspect the bounded audit artifact path before retrying")
	}
	pending, err = application.ports.Pending(ctx, change.location.Root, change.configuration.MaxGitOutputBytes, change.configuration.MaxGitPaths)
	if err != nil || pending.IndexDirty || len(pending.Paths) != 1 || pending.Paths[0] != path {
		return application.result(domain.OutcomeBlocked, "L7-REVIEW-009", string(request.Command), "recovery-required", "audit artifact did not produce one exact pending path", "inspect the worktree and preserve unrelated changes")
	}
	location, err := application.ports.Commit(ctx, domain.CommitRequest{
		Root: change.location.Root, ExpectedCommit: parent, ExpectedTree: change.location.Tree,
		Paths: []string{path}, Message: auditSubject(change.active.ID), MaxOutputBytes: change.configuration.MaxGitOutputBytes,
		MaxPaths: change.configuration.MaxGitPaths, MaxCommandSeconds: change.configuration.MaxCommandSeconds,
	})
	if err != nil {
		return application.operationError(ctx, request.Command, "L7-COMMIT-003", "recovery-required", "audit-record commit failed: "+bounded(err.Error(), 512), "inspect the index and rerun l7 review for recovery")
	}
	evidence.ReviewCommit, evidence.ReviewTree = location.Head, location.Tree
	if err := application.ports.SaveReview(change.location.CommonDir, evidence); err != nil {
		return application.failure("L7-STATE-007", request.Command, "recovery-required", "audit record committed but evidence was not finalized", "rerun l7 review to recover the exact commit")
	}
	change.location = location
	return application.reviewResult(request, change, evidence, "review completed and its sole Tier 3 audit record was committed")
}

func (application Application) runResult(request domain.Request, change executionContext, evidence domain.RunEvidence, message string) domain.Result {
	result := application.result(domain.OutcomePass, "L7-RUN-000", string(request.Command), string(domain.StateBuilding), message, "run l7 verify")
	result.Repository = repositoryForExecution(change, evidence.Candidate)
	result.Execution = &domain.ExecutionDetails{Role: domain.RoleImplementer, Provider: evidence.Provider.Provider, Executable: evidence.Provider.Executable, Version: evidence.Provider.Version, Digest: evidence.Provider.Digest, Commit: evidence.Candidate.Commit, Tree: evidence.Candidate.Tree}
	return result
}

func (application Application) verificationResult(request domain.Request, change executionContext, evidence domain.VerificationEvidence, message string) domain.Result {
	state := domain.StateVerified
	next := "run l7 review --agent codex|claude"
	if change.active.Tier == domain.TierHighRisk {
		state = domain.StateAwaitingIndependentAudit
	}
	result := application.result(domain.OutcomePass, "L7-VERIFY-000", string(request.Command), string(state), message, next)
	result.Repository = repositoryForExecution(change, evidence.Candidate)
	result.Execution = &domain.ExecutionDetails{Commit: evidence.Candidate.Commit, Tree: evidence.Candidate.Tree, Checks: append([]domain.CheckResult{}, evidence.Checks...)}
	return result
}

func (application Application) reviewResult(request domain.Request, change executionContext, evidence domain.ReviewEvidence, message string) domain.Result {
	outcome, code, state, next := domain.OutcomePass, "L7-REVIEW-000", domain.StateReviewed, "retain the reviewed candidate for Wave 4 readiness evaluation"
	if evidence.Decision == domain.DecisionNoGO {
		outcome, code, state, next = domain.OutcomeBlocked, "L7-REVIEW-010", domain.StateBuilding, "run l7 run --agent codex|claude --message <conventional-subject> to remediate"
		message = "review returned NO_GO"
	}
	result := application.result(outcome, code, string(request.Command), string(state), message, next)
	result.Repository = repositoryForExecution(change, evidence.Candidate)
	result.Execution = &domain.ExecutionDetails{Role: domain.RoleReviewer, Provider: evidence.Provider.Provider, Executable: evidence.Provider.Executable, Version: evidence.Provider.Version, Digest: evidence.Provider.Digest, Commit: evidence.Candidate.Commit, Tree: evidence.Candidate.Tree, Decision: evidence.Decision}
	return result
}

func (application Application) operationError(ctx context.Context, command domain.Command, code, state, message, next string) domain.Result {
	if ctx.Err() != nil {
		return application.result(domain.OutcomeCancelled, "L7-CLI-003", string(command), "cancelled", "command cancelled", "run l7 status to reconstruct the accepted state")
	}
	return application.failure(code, command, state, message, next)
}

func providerTask(change executionContext, selected domain.Provider, role domain.ProviderRole, candidate domain.CandidateIdentity) domain.ProviderTask {
	return domain.ProviderTask{
		Role: role, Provider: selected, RepositoryRoot: change.location.Root, ChangeID: change.active.ID, Tier: change.active.Tier,
		Base: change.active.Base, Candidate: candidate, Problem: change.active.Problem, Scope: append([]string{}, change.active.Scope...),
		AcceptanceCriteria: append([]string{}, change.brief.AcceptanceCriteria...), Risks: append([]string{}, change.brief.Risks...), Rollback: append([]string{}, change.brief.Rollback...),
	}
}

func executionDetails(view evidenceView) *domain.ExecutionDetails {
	if view.reviewValid {
		return &domain.ExecutionDetails{Role: domain.RoleReviewer, Provider: view.review.Provider.Provider, Executable: view.review.Provider.Executable, Version: view.review.Provider.Version, Digest: view.review.Provider.Digest, Commit: view.review.Candidate.Commit, Tree: view.review.Candidate.Tree, Decision: view.review.Decision}
	}
	if view.verificationValid {
		return &domain.ExecutionDetails{Commit: view.verification.Candidate.Commit, Tree: view.verification.Candidate.Tree, Checks: append([]domain.CheckResult{}, view.verification.Checks...)}
	}
	if view.runValid {
		return &domain.ExecutionDetails{Role: domain.RoleImplementer, Provider: view.run.Provider.Provider, Executable: view.run.Provider.Executable, Version: view.run.Provider.Version, Digest: view.run.Provider.Digest, Commit: view.run.Candidate.Commit, Tree: view.run.Candidate.Tree}
	}
	return nil
}

func runAtCurrentChain(change executionContext, view evidenceView) bool {
	if change.location.Head == view.run.Candidate.Commit && change.location.Tree == view.run.Candidate.Tree {
		return true
	}
	if view.verificationValid && verificationAtCurrent(change, view.verification) {
		return true
	}
	return view.reviewValid && reviewAtCurrent(change, view.review)
}

func verificationAtCurrentChain(change executionContext, view evidenceView) bool {
	return verificationAtCurrent(change, view.verification) || (view.reviewValid && reviewAtCurrent(change, view.review))
}

func verificationAtCurrent(change executionContext, evidence domain.VerificationEvidence) bool {
	if change.active.Tier == domain.TierHighRisk {
		return change.location.Head == evidence.VerificationCommit && change.location.Tree == evidence.VerificationTree
	}
	return change.location.Head == evidence.Candidate.Commit && change.location.Tree == evidence.Candidate.Tree
}

func reviewAtCurrent(change executionContext, evidence domain.ReviewEvidence) bool {
	if change.active.Tier == domain.TierHighRisk {
		return change.location.Head == evidence.ReviewCommit && change.location.Tree == evidence.ReviewTree
	}
	return change.location.Head == evidence.Candidate.Commit && change.location.Tree == evidence.Candidate.Tree
}

func repositoryForExecution(change executionContext, candidate domain.CandidateIdentity) *domain.RepositoryDetails {
	return &domain.RepositoryDetails{
		Root: change.location.Root, CommonDir: change.location.CommonDir, ChangeID: change.active.ID, Tier: change.active.Tier,
		Base: change.active.Base, Head: change.location.Head, Tree: change.location.Tree, DeclaredScope: append([]string{}, change.active.Scope...), ChangedPaths: []string{}, ExpandedPaths: []string{},
	}
}

func sameExecutionContext(left, right executionContext) bool {
	return sameLocation(left.location, right.location) && sameConfiguration(left.configuration, right.configuration) && sameResolvedActive(left.active, right.active) && sameBrief(left.brief, right.brief) && sameSnapshot(left.snapshot, right.snapshot)
}

func sameResolvedActive(left, right domain.ActiveChange) bool {
	return left.Kind == right.Kind && left.ID == right.ID && left.Tier == right.Tier && left.Base == right.Base && left.Problem == right.Problem && left.BriefPath == right.BriefPath && sameStrings(left.Scope, right.Scope)
}

func sameVerificationEvidence(left, right domain.VerificationEvidence) bool {
	if left.ChangeID != right.ChangeID || left.Candidate != right.Candidate || left.Result != right.Result || left.ConfigurationDigest != right.ConfigurationDigest || left.VerificationCommit != right.VerificationCommit || left.VerificationTree != right.VerificationTree || len(left.Checks) != len(right.Checks) {
		return false
	}
	for index := range left.Checks {
		if left.Checks[index] != right.Checks[index] {
			return false
		}
	}
	return true
}

func sameReviewEvidence(left, right domain.ReviewEvidence) bool {
	return left.ChangeID == right.ChangeID && left.Provider == right.Provider && left.Candidate == right.Candidate && left.Decision == right.Decision && left.ReviewCommit == right.ReviewCommit && left.ReviewTree == right.ReviewTree && sameStrings(left.Findings, right.Findings)
}

func sameEvidenceState(left, right evidenceView) bool {
	return left.runFound == right.runFound && (!left.runFound || left.run == right.run) &&
		left.verificationFound == right.verificationFound && (!left.verificationFound || sameVerificationEvidence(left.verification, right.verification)) &&
		left.reviewFound == right.reviewFound && (!left.reviewFound || sameReviewEvidence(left.review, right.review))
}

func sameEvidenceView(left, right evidenceView) bool {
	return sameEvidenceState(left, right) && left.runValid == right.runValid && left.verificationValid == right.verificationValid && left.reviewValid == right.reviewValid
}

func sameExecutionIntake(left, right executionContext) bool {
	return sameLocation(left.location, right.location) && sameConfiguration(left.configuration, right.configuration) && sameResolvedActive(left.active, right.active) && sameBrief(left.brief, right.brief)
}

func (application Application) approvalCurrent(ctx context.Context, change executionContext, implementer domain.Provider) (bool, error) {
	if change.active.Tier != domain.TierHighRisk {
		return true, nil
	}
	briefCommit, err := application.ports.PathCommit(ctx, change.location.Root, change.active.BriefPath)
	if err != nil {
		return false, err
	}
	approval, found, err := application.ports.LoadApproval(change.location.CommonDir)
	if err != nil || !found {
		return false, err
	}
	return approvalMatches(approval, change.active.ID, implementer, briefCommit), nil
}

func approvalMatches(binding domain.ApprovalBinding, changeID string, implementer domain.Provider, briefCommit string) bool {
	return binding.ChangeID == changeID && binding.Actor != "" && binding.Implementer == implementer && binding.BriefCommit == briefCommit
}

func verificationSubject(changeID string) string {
	return "test(l7): record " + changeID + " verification"
}
func auditSubject(changeID string) string { return "docs(l7): record " + changeID + " audit" }
func verificationPath(changeID string) string {
	return "docs/artifacts/changes/" + changeID + "-verification.md"
}
func auditPath(changeID string) string { return "docs/artifacts/changes/" + changeID + "-audit.md" }

func (application Application) executionPortsAvailable() bool {
	return application.lifecyclePortsAvailable() && application.ports.Pending != nil && application.ports.PathCommit != nil && application.ports.Commit != nil && application.ports.CommitMatches != nil && application.ports.CommitPaths != nil && application.ports.CommitTree != nil && application.ports.PathSetDigest != nil && application.ports.ConfirmApproval != nil && application.ports.LoadApproval != nil && application.ports.SaveApproval != nil && application.ports.LoadRun != nil && application.ports.SaveRun != nil && application.ports.LoadVerification != nil && application.ports.SaveVerification != nil && application.ports.LoadReview != nil && application.ports.SaveReview != nil && application.ports.RunProvider != nil && application.ports.RunVerification != nil && application.ports.WriteVerification != nil && application.ports.WriteAudit != nil
}

func sameVerificationCommands(left, right []domain.VerificationCommand) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name || left[index].Benchmark != right[index].Benchmark || !sameStrings(left[index].Argv, right[index].Argv) {
			return false
		}
	}
	return true
}

func checksMatchConfiguration(checks []domain.CheckResult, commands []domain.VerificationCommand) bool {
	if len(checks) != len(commands) {
		return false
	}
	for index := range checks {
		if checks[index].Name != commands[index].Name || checks[index].Benchmark != commands[index].Benchmark || !checks[index].Passed || checks[index].ExitCode != 0 {
			return false
		}
	}
	return true
}
