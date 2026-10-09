package headlessworker

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
	verifyadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/verify"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	stageStart     = "start"
	stageRepairing = "repairing"
	stageVerified  = "verified"
	stageCommitted = "committed"
	stageReviewed  = "reviewed"
	stageMerged    = "merged"

	maxFeedbackBytes = 12 << 10
	maxWorkerSteps   = 32
)

type providerFunc func(ctx context.Context, root string, route domain.RouteDecision, prompt, session string, reviewer bool, scope []string, commands [][]string) (providerResult, error)

// CrewExecutor runs crew tasks with the same routing, provider adapters,
// verification runner, reviewer separation, and compare-and-swap merge as
// Headless waves. It adds a repair loop that resumes the same session with
// verification output or reviewer findings, and a rebase step before merge.
type CrewExecutor struct {
	Executor
	worktrees string
	provider  providerFunc
	snapshots func() ([]domain.ProviderSnapshot, bool, error)
	verify    func(context.Context, string, []domain.VerificationCommand) ([]domain.CheckResult, string, error)
}

type crewProgress struct {
	Schema                 int                  `json:"schema"`
	PlanDigest             string               `json:"plan_digest"`
	TaskID                 string               `json:"task_id"`
	Attempt                int                  `json:"attempt"`
	Stage                  string               `json:"stage"`
	Worktree               string               `json:"worktree"`
	BaseCommit             string               `json:"base_commit"`
	ImplementationRoute    domain.RouteDecision `json:"implementation_route"`
	ImplementationSession  string               `json:"implementation_session"`
	ImplementationFailures int                  `json:"implementation_failures"`
	RepairRounds           int                  `json:"repair_rounds"`
	Feedback               string               `json:"feedback"`
	CandidateCommit        string               `json:"candidate_commit"`
	ReviewRoute            domain.RouteDecision `json:"review_route"`
	ReviewSession          string               `json:"review_session"`
	ReviewFailures         int                  `json:"review_failures"`
	ReviewedCommit         string               `json:"reviewed_commit"`
	ReviewedPatch          string               `json:"reviewed_patch"`
	UpdatedAtUTC           string               `json:"updated_at_utc"`
}

func NewCrew(root, common string, configuration orchestrationconfig.File) (CrewExecutor, error) {
	base, err := New(root, common, configuration)
	if err != nil {
		return CrewExecutor{}, err
	}
	store, err := crew.Open(common)
	if err != nil {
		return CrewExecutor{}, err
	}
	executor := CrewExecutor{Executor: base, worktrees: store.WorktreeRoot()}
	executor.provider = base.runProvider
	executor.snapshots = func() ([]domain.ProviderSnapshot, bool, error) { return state.LoadProviderSnapshots(common) }
	executor.verify = func(ctx context.Context, worktree string, commands []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		return verifyadapter.New(nil, nil).RunWithFailureTail(ctx, worktree, commands, configuration.Tools.MaxOutputBytes, configuration.Tools.MaxSeconds)
	}
	return executor, nil
}

func (executor CrewExecutor) Build(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (crew.Outcome, error) {
	if task.Shape == domain.CrewScout {
		return executor.scout(ctx, plan, task, checkpoint)
	}
	if task.Shape != domain.CrewShip {
		return decision(domain.CrewDecisionBlocked, "unknown crew task shape"), nil
	}
	target, err := executor.ref(ctx, plan.TargetBranch)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, "target branch "+plan.TargetBranch+" is unavailable"), nil
	}
	progress, err := executor.loadCrewProgress(plan, task, checkpoint.Attempt)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, err.Error()), nil
	}
	worktree, err := executor.crewWorktree(ctx, task, checkpoint.Attempt, target, progress.Worktree, false)
	if err != nil {
		return failed("worktree", err.Error()), nil
	}
	if progress.Worktree == "" {
		progress.Worktree, progress.BaseCommit, progress.Stage = worktree, target, stageStart
		if err := executor.saveCrewProgress(plan, &progress); err != nil {
			return failed("worker-checkpoint", err.Error()), nil
		}
	}
	snapshots, found, err := executor.snapshots()
	if err != nil || !found {
		return decision(domain.CrewDecisionBlocked, "verified provider snapshot is unavailable; run l7 providers probe"), nil
	}
	for step := 0; step < maxWorkerSteps; step++ {
		var outcome crew.Outcome
		var stop bool
		switch progress.Stage {
		case stageStart, stageRepairing:
			outcome, stop = executor.implement(ctx, plan, task, &progress, snapshots)
		case stageVerified:
			outcome, stop = executor.commitCandidate(ctx, plan, task, &progress)
		case stageCommitted:
			outcome, stop = executor.review(ctx, plan, task, &progress, snapshots)
		case stageReviewed, stageMerged:
			return crew.Outcome{
				Kind: crew.OutcomeBuilt, Worktree: progress.Worktree, CandidateCommit: progress.CandidateCommit, Verification: "passed",
				ProviderID: progress.ImplementationRoute.ProviderID, ModelID: progress.ImplementationRoute.ModelID, SessionID: progress.ImplementationSession,
				Message: "candidate verified and independently reviewed",
			}, nil
		default:
			return decision(domain.CrewDecisionBlocked, "crew worker checkpoint has an unknown stage"), nil
		}
		if stop {
			return outcome, nil
		}
	}
	return failed("worker-steps", "crew worker exceeded its step bound without finishing"), nil
}

func (executor CrewExecutor) implement(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, progress *crewProgress, snapshots []domain.ProviderSnapshot) (crew.Outcome, bool) {
	profile := domain.TaskProfile{
		Schema: domain.OrchestrationSchema, ID: task.ID + "/implement", Summary: task.Title, Complexity: domain.ComplexityC3,
		RiskTier: domain.TierProduct, ContextTokens: 64_000, NeedsTools: true, NeedsEditing: true, NeedsResume: true,
		PriorFailures: progress.ImplementationFailures,
	}
	route := routeForAttempt(profile, snapshots, progress.ImplementationFailures)
	if route.ProviderID == "" {
		return decision(domain.CrewDecisionBlocked, "no qualified implementation route is available"), true
	}
	if !sameRoute(progress.ImplementationRoute, route) {
		if progress.ImplementationRoute.ProviderID != "" {
			progress.ImplementationSession = ""
		}
		route.DecisionUTC = time.Now().UTC().Format(time.RFC3339)
		progress.ImplementationRoute = route
		_ = state.SaveRouteDecision(executor.common, route)
	}
	prompt := crewImplementationPrompt(task)
	if progress.Feedback != "" {
		prompt = crewRepairPrompt(task, progress.Feedback)
	}
	result, runErr := executor.provider(ctx, progress.Worktree, route, prompt, progress.ImplementationSession, false, task.AllowedPaths, task.Verification)
	if result.SessionID != "" {
		progress.ImplementationSession = result.SessionID
	}
	if result.Quota && result.Reset != "" {
		_ = executor.saveCrewProgress(plan, progress)
		return crew.Outcome{Kind: crew.OutcomeQuota, QuotaResetAtUTC: result.Reset, ProviderID: route.ProviderID, ModelID: route.ModelID, Message: "implementation provider reported a quota reset"}, true
	}
	if runErr != nil {
		progress.ImplementationFailures++
		_ = executor.saveCrewProgress(plan, progress)
		return withRoute(failed("implementation:"+route.ProviderID, "implementation session failed: "+runErr.Error()), route), true
	}
	pending, err := executor.git.Pending(ctx, progress.Worktree)
	if err != nil {
		return failed("pending", err.Error()), true
	}
	if pending.IndexDirty {
		return decision(domain.CrewDecisionScopeExpanded, "the worker changed the Git index outside the controlled commit"), true
	}
	for _, relative := range pending.Paths {
		if !domain.ScopeContains(task.AllowedPaths, relative) || protected(relative) {
			return decision(domain.CrewDecisionScopeExpanded, "the worker changed "+relative+", which is outside the approved scope or protected"), true
		}
	}
	if len(pending.Paths) == 0 {
		progress.ImplementationFailures++
		_ = executor.saveCrewProgress(plan, progress)
		return failed("no-change", "the worker finished without changing any file"), true
	}
	checks, tail, verifyErr := executor.verify(ctx, progress.Worktree, crewVerificationCommands(task.Verification))
	if problem, broken := verificationEnvironmentProblem(checks); broken {
		return decision(domain.CrewDecisionBlocked, problem), true
	}
	if verifyErr == nil && allPassed(checks) {
		progress.Stage, progress.Feedback, progress.RepairRounds = stageVerified, "", 0
		if err := executor.saveCrewProgress(plan, progress); err != nil {
			return failed("worker-checkpoint", err.Error()), true
		}
		return crew.Outcome{}, false
	}
	feedback := verificationFeedback(checks, task.Verification, tail, verifyErr)
	progress.Stage, progress.Feedback = stageRepairing, feedback
	if progress.RepairRounds < plan.RepairRounds {
		progress.RepairRounds++
		if err := executor.saveCrewProgress(plan, progress); err != nil {
			return failed("worker-checkpoint", err.Error()), true
		}
		return crew.Outcome{}, false
	}
	progress.ImplementationFailures++
	progress.RepairRounds = 0
	_ = executor.saveCrewProgress(plan, progress)
	return failed("verification:"+failedCheckName(checks), fmt.Sprintf("verification still failed after %d repair rounds", plan.RepairRounds)), true
}

func (executor CrewExecutor) commitCandidate(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, progress *crewProgress) (crew.Outcome, bool) {
	pending, err := executor.git.Pending(ctx, progress.Worktree)
	if err != nil {
		return failed("pending", err.Error()), true
	}
	if len(pending.Paths) == 0 && !pending.IndexDirty && progress.CandidateCommit != "" && pending.Head == progress.CandidateCommit {
		progress.Stage = stageCommitted
		return crew.Outcome{}, false
	}
	if pending.IndexDirty || len(pending.Paths) == 0 {
		return decision(domain.CrewDecisionBlocked, "verified changes are missing or the index changed before commit"), true
	}
	for _, relative := range pending.Paths {
		if !domain.ScopeContains(task.AllowedPaths, relative) || protected(relative) {
			return decision(domain.CrewDecisionScopeExpanded, "the worker changed "+relative+", which is outside the approved scope or protected"), true
		}
	}
	candidate, err := executor.git.Commit(ctx, domain.CommitRequest{
		Root: progress.Worktree, ExpectedCommit: pending.Head, ExpectedTree: pending.Tree, Paths: pending.Paths,
		Message: crewCommitSubject(task), MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes,
		MaxPaths: 100_000, MaxCommandSeconds: executor.configuration.Tools.MaxSeconds,
	})
	if err != nil {
		return failed("commit", err.Error()), true
	}
	progress.CandidateCommit, progress.Stage = candidate.Head, stageCommitted
	if err := executor.saveCrewProgress(plan, progress); err != nil {
		return failed("worker-checkpoint", err.Error()), true
	}
	return crew.Outcome{}, false
}

func (executor CrewExecutor) review(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, progress *crewProgress, snapshots []domain.ProviderSnapshot) (crew.Outcome, bool) {
	profile := domain.TaskProfile{
		Schema: domain.OrchestrationSchema, ID: task.ID + "/review", Summary: "independent review " + task.Title, Complexity: domain.ComplexityC3,
		RiskTier: domain.TierProduct, ContextTokens: 64_000, NeedsTools: true, IndependentReview: true,
		ImplementerProvider: progress.ImplementationRoute.ProviderID, ImplementerModel: progress.ImplementationRoute.ModelID,
		PriorFailures: progress.ReviewFailures,
	}
	route := routeForAttempt(profile, snapshots, progress.ReviewFailures)
	if route.ProviderID == "" {
		return decision(domain.CrewDecisionBlocked, "no independent qualified reviewer is available"), true
	}
	if !sameRoute(progress.ReviewRoute, route) {
		if progress.ReviewRoute.ProviderID != "" {
			progress.ReviewSession = ""
		}
		route.DecisionUTC = time.Now().UTC().Format(time.RFC3339)
		progress.ReviewRoute = route
	}
	result, runErr := executor.provider(ctx, progress.Worktree, route, crewReviewPrompt(task, progress.BaseCommit, progress.CandidateCommit), progress.ReviewSession, true, task.AllowedPaths, task.Verification)
	if result.SessionID != "" {
		progress.ReviewSession = result.SessionID
	}
	if result.Quota && result.Reset != "" {
		_ = executor.saveCrewProgress(plan, progress)
		return crew.Outcome{Kind: crew.OutcomeQuota, QuotaResetAtUTC: result.Reset, ProviderID: route.ProviderID, ModelID: route.ModelID, Message: "reviewer reported a quota reset"}, true
	}
	if runErr != nil {
		progress.ReviewFailures++
		_ = executor.saveCrewProgress(plan, progress)
		return withRoute(failed("review:"+route.ProviderID, "independent review session failed: "+runErr.Error()), route), true
	}
	after, err := executor.git.Pending(ctx, progress.Worktree)
	if err != nil || len(after.Paths) != 0 || after.IndexDirty || after.Head != progress.CandidateCommit {
		return decision(domain.CrewDecisionScopeExpanded, "the independent reviewer modified the candidate"), true
	}
	if result.Decision == domain.DecisionGO {
		patch, err := executor.patchDigest(ctx, progress.Worktree, progress.BaseCommit, progress.CandidateCommit)
		if err != nil {
			return failed("patch", err.Error()), true
		}
		progress.ReviewedCommit, progress.ReviewedPatch, progress.Stage = progress.CandidateCommit, patch, stageReviewed
		if err := executor.saveCrewProgress(plan, progress); err != nil {
			return failed("worker-checkpoint", err.Error()), true
		}
		return crew.Outcome{}, false
	}
	progress.Stage = stageRepairing
	progress.Feedback = "The independent reviewer returned NO_GO:\n" + bounded(result.Summary, 4096)
	if progress.RepairRounds < plan.RepairRounds {
		progress.RepairRounds++
		if err := executor.saveCrewProgress(plan, progress); err != nil {
			return failed("worker-checkpoint", err.Error()), true
		}
		return crew.Outcome{}, false
	}
	progress.ReviewFailures++
	progress.RepairRounds = 0
	_ = executor.saveCrewProgress(plan, progress)
	return failed("review:no-go", "independent review returned NO_GO: "+bounded(result.Summary, 512)), true
}

// Integrate fast-forwards the target to the reviewed candidate. If the target
// moved, it rebases first; a changed patch or failing verification sends the
// task back for another build step instead of merging.
func (executor CrewExecutor) Integrate(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (crew.Outcome, error) {
	progress, err := executor.loadCrewProgress(plan, task, checkpoint.Attempt)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, err.Error()), nil
	}
	if task.Shape != domain.CrewShip || (progress.Stage != stageReviewed && progress.Stage != stageMerged) || progress.ReviewedCommit == "" || progress.ReviewedCommit != progress.CandidateCommit {
		return failed("integrate-unreviewed", "the candidate is not reviewed at its current commit"), nil
	}
	target, err := executor.ref(ctx, plan.TargetBranch)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, "target branch "+plan.TargetBranch+" is unavailable"), nil
	}
	if target == progress.CandidateCommit {
		progress.Stage = stageMerged
		_ = executor.saveCrewProgress(plan, &progress)
		return crew.Outcome{Kind: crew.OutcomeMerged, CandidateCommit: target, Message: "target already contains the candidate"}, nil
	}
	descendant, err := executor.isAncestor(ctx, progress.Worktree, target, progress.CandidateCommit)
	if err != nil {
		return failed("ancestry", err.Error()), nil
	}
	if !descendant {
		rebased, conflicts, err := executor.rebase(ctx, progress.Worktree, target)
		if len(conflicts) != 0 {
			return decision(domain.CrewDecisionConflict, fmt.Sprintf("rebasing onto %s conflicts in: %s", short(target), strings.Join(conflicts, ", "))), nil
		}
		if err != nil {
			return failed("rebase", err.Error()), nil
		}
		progress.BaseCommit, progress.CandidateCommit = target, rebased
		checks, tail, verifyErr := executor.verify(ctx, progress.Worktree, crewVerificationCommands(task.Verification))
		if problem, broken := verificationEnvironmentProblem(checks); broken {
			_ = executor.saveCrewProgress(plan, &progress)
			return decision(domain.CrewDecisionBlocked, problem), nil
		}
		if verifyErr != nil || !allPassed(checks) {
			progress.Stage, progress.RepairRounds, progress.ReviewedCommit = stageRepairing, 0, ""
			progress.Feedback = "After rebasing onto the latest target, verification failed.\n" + verificationFeedback(checks, task.Verification, tail, verifyErr)
			if err := executor.saveCrewProgress(plan, &progress); err != nil {
				return failed("worker-checkpoint", err.Error()), nil
			}
			return crew.Outcome{Kind: crew.OutcomeBuilt, CandidateCommit: rebased, Verification: "failed", Message: "rebased candidate failed verification; repairing before merge"}, nil
		}
		patch, err := executor.patchDigest(ctx, progress.Worktree, target, rebased)
		if err != nil {
			return failed("patch", err.Error()), nil
		}
		if patch != progress.ReviewedPatch {
			progress.Stage, progress.ReviewedCommit = stageCommitted, ""
			if err := executor.saveCrewProgress(plan, &progress); err != nil {
				return failed("worker-checkpoint", err.Error()), nil
			}
			return crew.Outcome{Kind: crew.OutcomeBuilt, CandidateCommit: rebased, Verification: "passed", Message: "rebase changed the patch; re-reviewing before merge"}, nil
		}
		progress.ReviewedCommit = rebased
		if err := executor.saveCrewProgress(plan, &progress); err != nil {
			return failed("worker-checkpoint", err.Error()), nil
		}
	}
	paths, err := executor.git.CommitPaths(ctx, executor.root, target, progress.CandidateCommit)
	if err != nil {
		return failed("merge-paths", err.Error()), nil
	}
	for _, relative := range paths {
		if !domain.ScopeContains(task.AllowedPaths, relative) || protected(relative) {
			return decision(domain.CrewDecisionScopeExpanded, "the candidate changes "+relative+", which is outside the approved scope or protected"), nil
		}
	}
	merge := domain.MergeRequest{Root: executor.root, TargetBranch: plan.TargetBranch, ExpectedOld: target, Candidate: progress.CandidateCommit, MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes}
	inspected, err := executor.git.InspectMerge(ctx, merge)
	if err != nil {
		return failed("merge-inspect", err.Error()), nil
	}
	if !inspected.AlreadyAdvanced {
		if err := executor.git.AdvanceMerge(ctx, merge); err != nil {
			return failed("merge-cas", err.Error()), nil
		}
	}
	progress.Stage = stageMerged
	if err := executor.saveCrewProgress(plan, &progress); err != nil {
		return failed("worker-checkpoint", err.Error()), nil
	}
	return crew.Outcome{Kind: crew.OutcomeMerged, CandidateCommit: progress.CandidateCommit, Verification: "passed", Message: "fast-forwarded " + plan.TargetBranch + " to " + short(progress.CandidateCommit)}, nil
}

// scout runs a read-only investigation on a detached worktree through a
// native host. The reviewer contract keeps the provider read-only; its GO
// means the report satisfies the acceptance criteria.
func (executor CrewExecutor) scout(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (crew.Outcome, error) {
	target, err := executor.ref(ctx, plan.TargetBranch)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, "target branch "+plan.TargetBranch+" is unavailable"), nil
	}
	progress, err := executor.loadCrewProgress(plan, task, checkpoint.Attempt)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, err.Error()), nil
	}
	worktree, err := executor.crewWorktree(ctx, task, checkpoint.Attempt, target, progress.Worktree, true)
	if err != nil {
		return failed("worktree", err.Error()), nil
	}
	if progress.Worktree == "" {
		progress.Worktree, progress.BaseCommit, progress.Stage = worktree, target, stageStart
		if err := executor.saveCrewProgress(plan, &progress); err != nil {
			return failed("worker-checkpoint", err.Error()), nil
		}
	}
	snapshots, found, err := executor.snapshots()
	if err != nil || !found {
		return decision(domain.CrewDecisionBlocked, "verified provider snapshot is unavailable; run l7 providers probe"), nil
	}
	native := make([]domain.ProviderSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot.Kind == domain.ProviderKindCodexAppServer || snapshot.Kind == domain.ProviderKindClaudeCLI {
			native = append(native, snapshot)
		}
	}
	profile := domain.TaskProfile{
		Schema: domain.OrchestrationSchema, ID: task.ID + "/scout", Summary: task.Title, Complexity: domain.ComplexityC2,
		RiskTier: domain.TierProduct, ContextTokens: 64_000, NeedsTools: true, PriorFailures: progress.ImplementationFailures,
	}
	route := routeForAttempt(profile, native, progress.ImplementationFailures)
	if route.ProviderID == "" {
		return decision(domain.CrewDecisionBlocked, "no native Codex or Claude host is qualified for a read-only scout"), nil
	}
	if !sameRoute(progress.ImplementationRoute, route) {
		progress.ImplementationSession, progress.ImplementationRoute = "", route
	}
	result, runErr := executor.provider(ctx, worktree, route, crewScoutPrompt(task), progress.ImplementationSession, true, nil, nil)
	if result.SessionID != "" {
		progress.ImplementationSession = result.SessionID
	}
	if result.Quota && result.Reset != "" {
		_ = executor.saveCrewProgress(plan, &progress)
		return crew.Outcome{Kind: crew.OutcomeQuota, QuotaResetAtUTC: result.Reset, ProviderID: route.ProviderID, ModelID: route.ModelID, Message: "scout provider reported a quota reset"}, nil
	}
	if runErr != nil {
		progress.ImplementationFailures++
		_ = executor.saveCrewProgress(plan, &progress)
		return withRoute(failed("scout:"+route.ProviderID, "scout session failed: "+runErr.Error()), route), nil
	}
	after, err := executor.git.Pending(ctx, worktree)
	if err != nil || len(after.Paths) != 0 || after.IndexDirty || after.Head != target && after.Head != progress.BaseCommit {
		return decision(domain.CrewDecisionScopeExpanded, "the read-only scout changed its worktree"), nil
	}
	status := "answered"
	if result.Decision != domain.DecisionGO {
		status = "incomplete"
	}
	report := fmt.Sprintf("# %s\n\n- Result: %s\n- Provider: %s/%s\n- Base: %s\n\n%s\n", task.Title, status, route.ProviderID, route.ModelID, progress.BaseCommit, strings.TrimSpace(result.Summary))
	progress.Stage = stageReviewed
	_ = executor.saveCrewProgress(plan, &progress)
	return crew.Outcome{Kind: crew.OutcomeReported, Report: report, Worktree: worktree, ProviderID: route.ProviderID, ModelID: route.ModelID, SessionID: progress.ImplementationSession, Message: "scout report " + status}, nil
}

func (executor CrewExecutor) crewWorktree(ctx context.Context, task domain.CrewTask, attempt int, target, recorded string, detached bool) (string, error) {
	if err := localfile.EnsureDirectory(executor.worktrees, 0o700); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s-a%d", task.ID, attempt)
	worktree := filepath.Join(executor.worktrees, name)
	branch := "l7/tasks/" + name
	for _, candidate := range []string{recorded, worktree} {
		if candidate == "" {
			continue
		}
		physical, err := filepath.EvalSymlinks(candidate)
		if err == nil && physical == filepath.Clean(candidate) && executor.crewWorktreeMatches(ctx, physical, branch, detached) {
			return physical, nil
		}
	}
	if _, err := os.Lstat(worktree); err == nil {
		return "", errors.New("existing crew worktree does not match the task identity")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	arguments := []string{"worktree", "add", "-b", branch, worktree, target}
	if detached {
		arguments = []string{"worktree", "add", "--detach", worktree, target}
	}
	if _, err := executor.gitChecked(ctx, executor.root, arguments...); err != nil {
		return "", errors.New("cannot create disposable crew worktree")
	}
	return filepath.EvalSymlinks(worktree)
}

func (executor CrewExecutor) crewWorktreeMatches(ctx context.Context, worktree, branch string, detached bool) bool {
	root, err := executor.gitLine(ctx, worktree, "rev-parse", "--show-toplevel")
	if err != nil || root != worktree {
		return false
	}
	current, err := executor.gitLine(ctx, worktree, "symbolic-ref", "--short", "HEAD")
	if detached {
		return err != nil
	}
	return err == nil && current == branch
}

func (executor CrewExecutor) isAncestor(ctx context.Context, directory, ancestor, descendant string) (bool, error) {
	result, err := executor.gitResult(ctx, directory, nil, "merge-base", "--is-ancestor", ancestor, descendant)
	if err != nil {
		return false, err
	}
	switch result.ExitCode {
	case 0:
		return true, nil
	case 1:
		return false, nil
	default:
		return false, errors.New("cannot compare candidate and target ancestry")
	}
}

// rebase replays the task branch onto target. On conflict it aborts and
// returns the conflicting paths; nothing resolves conflicts automatically.
func (executor CrewExecutor) rebase(ctx context.Context, worktree, target string) (string, []string, error) {
	environment := append(processadapter.MinimalEnvironment(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_EDITOR=true")
	result, err := executor.gitResult(ctx, worktree, environment, "-c", "core.fsmonitor=false", "rebase", "--no-autostash", target)
	if err != nil {
		return "", nil, err
	}
	if result.ExitCode != 0 {
		conflicts, _ := executor.gitResult(ctx, worktree, nil, "diff", "--name-only", "--diff-filter=U", "-z")
		paths := nulPaths(conflicts.Stdout)
		_, _ = executor.gitResult(ctx, worktree, environment, "rebase", "--abort")
		if len(paths) == 0 {
			return "", nil, errors.New("rebase failed without reporting conflicting paths")
		}
		return "", paths, nil
	}
	head, err := executor.gitLine(ctx, worktree, "rev-parse", "HEAD")
	if err != nil || len(head) != 40 {
		return "", nil, errors.New("cannot read the rebased candidate")
	}
	return head, nil, nil
}

// patchDigest hashes the changed lines between base and candidate while
// ignoring hunk positions, so a rebase that only shifts context keeps its
// review.
func (executor CrewExecutor) patchDigest(ctx context.Context, directory, base, candidate string) (string, error) {
	result, err := executor.gitResult(ctx, directory, nil, "diff", "--no-color", "--no-ext-diff", "--no-renames", "--unified=0", base, candidate, "--")
	if err != nil || result.ExitCode != 0 {
		return "", errors.New("cannot read the candidate patch")
	}
	digest := sha256.New()
	scanner := bufio.NewScanner(bytes.NewReader(result.Stdout))
	scanner.Buffer(make([]byte, 64<<10), 8<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte("@@")) || bytes.HasPrefix(line, []byte("index ")) {
			continue
		}
		_, _ = digest.Write(line)
		_, _ = digest.Write([]byte{'\n'})
	}
	if scanner.Err() != nil {
		return "", errors.New("candidate patch exceeds its bounds")
	}
	return fmt.Sprintf("sha256:%x", digest.Sum(nil)), nil
}

func (executor CrewExecutor) gitResult(ctx context.Context, directory string, environment []string, arguments ...string) (processadapter.Result, error) {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return processadapter.Result{}, err
	}
	if environment == nil {
		environment = processadapter.MinimalEnvironment()
	}
	return (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: git.Path, Arguments: arguments, Directory: directory, Environment: environment,
		MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes, Timeout: time.Duration(executor.configuration.Tools.MaxSeconds) * time.Second,
	})
}

func (executor CrewExecutor) gitChecked(ctx context.Context, directory string, arguments ...string) (processadapter.Result, error) {
	result, err := executor.gitResult(ctx, directory, nil, arguments...)
	if err == nil && result.ExitCode != 0 {
		err = fmt.Errorf("git %s exited %d", arguments[0], result.ExitCode)
	}
	return result, err
}

func (executor CrewExecutor) loadCrewProgress(plan domain.CrewPlan, task domain.CrewTask, attempt int) (crewProgress, error) {
	path, err := executor.crewProgressPath(task.ID, attempt)
	if err != nil {
		return crewProgress{}, err
	}
	data, err := localfile.Read(path, 4<<20)
	if errors.Is(err, os.ErrNotExist) {
		return crewProgress{Schema: domain.CrewSchema, PlanDigest: plan.Digest, TaskID: task.ID, Attempt: attempt, ImplementationRoute: emptyRoute(), ReviewRoute: emptyRoute()}, nil
	}
	if err != nil {
		return crewProgress{}, err
	}
	var progress crewProgress
	if err := localfile.DecodeJSON(data, &progress); err != nil {
		return crewProgress{}, fmt.Errorf("decode crew worker checkpoint: %w", err)
	}
	if progress.Schema != domain.CrewSchema || progress.PlanDigest != plan.Digest || progress.TaskID != task.ID || progress.Attempt != attempt ||
		progress.ImplementationFailures < 0 || progress.ReviewFailures < 0 || progress.RepairRounds < 0 {
		return crewProgress{}, errors.New("crew worker checkpoint is invalid or belongs to another plan")
	}
	return progress, nil
}

func (executor CrewExecutor) saveCrewProgress(plan domain.CrewPlan, progress *crewProgress) error {
	progress.Schema, progress.PlanDigest = domain.CrewSchema, plan.Digest
	progress.Feedback = bounded(progress.Feedback, maxFeedbackBytes)
	progress.UpdatedAtUTC = time.Now().UTC().Format(time.RFC3339)
	path, err := executor.crewProgressPath(progress.TaskID, progress.Attempt)
	if err != nil {
		return err
	}
	return writePrivateJSON(path, progress, true)
}

func (executor CrewExecutor) crewProgressPath(taskID string, attempt int) (string, error) {
	if taskID == "" || strings.ContainsAny(taskID, "/\\\x00\r\n.") || attempt < 0 {
		return "", errors.New("crew worker identity is unsafe")
	}
	return filepath.Join(executor.common, "l7", "crew", "worker", fmt.Sprintf("%s-a%d.json", taskID, attempt)), nil
}

// CrewScopeProtected returns the first scope pattern that may cover a path a
// Tier 2 crew may never change: Git metadata, workflows, Level 7 policy, or
// agent instructions. Credential and .env files are also refused at runtime.
func CrewScopeProtected(patterns []string) (string, bool) {
	protectedScopes := []string{".git/**", ".github/workflows/**", ".l7/**", "AGENTS.md", "CLAUDE.md"}
	for _, pattern := range patterns {
		if protected(pattern) || domain.CrewPathScopesMayOverlap([]string{pattern}, protectedScopes) {
			return pattern, true
		}
	}
	return "", false
}

func crewVerificationCommands(commands [][]string) []domain.VerificationCommand {
	result := make([]domain.VerificationCommand, 0, len(commands))
	for index, argv := range commands {
		result = append(result, domain.VerificationCommand{Name: fmt.Sprintf("crew-%02d", index+1), Argv: append([]string{}, argv...)})
	}
	return result
}

func verificationFeedback(checks []domain.CheckResult, commands [][]string, tail string, runErr error) string {
	for index, check := range checks {
		if check.Passed {
			continue
		}
		argv := ""
		if index < len(commands) {
			argv = strings.Join(commands[index], " ")
		}
		return fmt.Sprintf("Check `%s` exited with code %d (%s).\n%s", argv, check.ExitCode, check.Code, bounded(tail, 2*verifyadapter.FailureTailBytes+64))
	}
	if runErr != nil {
		return "Verification could not run: " + bounded(runErr.Error(), 1024)
	}
	return "Verification did not report a passing result."
}

// verificationEnvironmentProblem reports checks that could not run at all,
// such as a missing or replaced executable. A worker cannot repair those.
func verificationEnvironmentProblem(checks []domain.CheckResult) (string, bool) {
	for _, check := range checks {
		if check.Code == "L7-VERIFY-002" {
			return "verification check " + check.Name + " cannot run: " + check.Message, true
		}
	}
	return "", false
}

func failedCheckName(checks []domain.CheckResult) string {
	for _, check := range checks {
		if !check.Passed {
			return fmt.Sprintf("%s:%d", check.Name, check.ExitCode)
		}
	}
	return "unknown"
}

func crewCommitSubject(task domain.CrewTask) string {
	title := strings.Join(strings.FieldsFunc(task.Title, func(character rune) bool { return character < 0x20 || character == 0x7f }), " ")
	title = strings.TrimSpace(bounded(title, 140))
	if title == "" {
		title = task.ID
	}
	return "feat(crew): " + title
}

func crewTaskBrief(task domain.CrewTask) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Task %s: %s\n\n%s\n\nAcceptance criteria:\n", task.ID, task.Title, task.Objective)
	for _, criterion := range task.AcceptanceCriteria {
		fmt.Fprintf(&output, "- %s\n", criterion)
	}
	if len(task.AllowedPaths) != 0 {
		output.WriteString("\nYou may change only these paths:\n")
		for _, path := range task.AllowedPaths {
			fmt.Fprintf(&output, "- %s\n", path)
		}
	}
	if len(task.Verification) != 0 {
		output.WriteString("\nLevel 7 runs these checks after each of your turns and returns any failure output to you:\n")
		for _, argv := range task.Verification {
			fmt.Fprintf(&output, "- %s\n", strings.Join(argv, " "))
		}
	}
	return output.String()
}

const crewWorkerRules = `Work like a careful senior engineer: find the root cause before changing code, keep the diff small, follow the repository's existing conventions, and add or update a test that fails without your change. Never weaken, skip, or delete a test to make checks pass. Do not push, release, deploy, use secrets, access the network, or change the Git index; Level 7 verifies and commits.`

func crewImplementationPrompt(task domain.CrewTask) string {
	return "You are one worker in a Level 7 crew. Implement exactly this task in this disposable worktree.\n\n" +
		crewTaskBrief(task) + "\n" + crewWorkerRules + "\n\nFinish with the required structured result summarizing what you changed and why."
}

func crewRepairPrompt(task domain.CrewTask, feedback string) string {
	return "You are one worker in a Level 7 crew. Your previous attempt at this task is in this worktree but did not pass. Fix the cause, not the symptom.\n\n" +
		"What failed:\n" + bounded(feedback, maxFeedbackBytes) + "\n\n" + crewTaskBrief(task) + "\n" + crewWorkerRules +
		"\n\nFinish with the required structured result summarizing what you changed and why."
}

func crewReviewPrompt(task domain.CrewTask, base, candidate string) string {
	return fmt.Sprintf("Independently and read-only review this Level 7 crew task at candidate %s against base %s; inspect `git diff %s %s`.\n\n%s\n"+
		"Check that the change meets every acceptance criterion, handles edge cases and error paths, stays within the allowed paths, keeps or strengthens tests, and adds no security, data-loss, or compatibility risk. "+
		"Do not modify files or Git state. Return GO only when the task is complete and safe; otherwise return NO_GO and put concrete, actionable findings in the summary.",
		candidate, base, base, candidate, crewTaskBrief(task))
}

func crewScoutPrompt(task domain.CrewTask) string {
	return "Investigate read-only and report for a Level 7 crew. Do not modify any file or Git state.\n\n" + crewTaskBrief(task) +
		"\nCite file paths and line numbers for every claim, separate what you verified from what you infer, and state what you could not determine. " +
		"Put the complete report in the summary (at most 4000 characters). Return GO when the report satisfies every acceptance criterion; otherwise return NO_GO and say what is missing."
}

func decision(kind domain.CrewDecisionKind, message string) crew.Outcome {
	return crew.Outcome{Kind: crew.OutcomeDecision, Decision: kind, Message: message}
}

func failed(stage, message string) crew.Outcome {
	digest := sha256.Sum256([]byte(stage))
	return crew.Outcome{Kind: crew.OutcomeFailed, FailureSignature: fmt.Sprintf("sha256:%x", digest), Message: stage + ": " + bounded(message, 1024)}
}

func withRoute(outcome crew.Outcome, route domain.RouteDecision) crew.Outcome {
	outcome.ProviderID, outcome.ModelID = route.ProviderID, route.ModelID
	return outcome
}

func short(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	for limit > 0 && (value[limit]&0xC0) == 0x80 {
		limit--
	}
	return value[:limit]
}
