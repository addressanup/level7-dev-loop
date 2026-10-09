package headlessworker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	stageDelivered = "delivered"
	crewRiskLabel  = "l7-risk-tier-2"
	ciLogBytes     = 4 << 10
	maxHandoff     = 16 << 10
)

// startPoint returns the commit a new ship worktree starts from: the local
// target branch, or the freshly fetched remote base for pull-request plans.
func (executor CrewExecutor) startPoint(ctx context.Context, plan domain.CrewPlan) (string, error) {
	if plan.Delivery != domain.CrewDeliveryPR {
		target, err := executor.ref(ctx, plan.TargetBranch)
		if err != nil {
			return "", errors.New("target branch " + plan.TargetBranch + " is unavailable")
		}
		return target, nil
	}
	tracking := "refs/remotes/" + plan.Remote + "/" + plan.TargetBranch
	result, err := executor.gitResult(ctx, executor.root, remoteEnvironment(), "fetch", "--no-tags", "--quiet", plan.Remote, "+refs/heads/"+plan.TargetBranch+":"+tracking)
	if err != nil || result.ExitCode != 0 {
		return "", fmt.Errorf("cannot fetch %s from %s: %s", plan.TargetBranch, plan.Remote, gitDiagnostic(result.Stderr, err))
	}
	head, err := executor.gitLine(ctx, executor.root, "rev-parse", "--verify", tracking+"^{commit}")
	if err != nil || len(head) != 40 {
		return "", errors.New("cannot read the fetched base " + tracking)
	}
	return head, nil
}

// deliver pushes the reviewed candidate to its task branch and opens or
// reuses the task's one pull request. Pushes are fast-forward only.
func (executor CrewExecutor) deliver(ctx context.Context, plan domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (crew.Outcome, error) {
	progress, err := executor.loadCrewProgress(plan, task, checkpoint.Attempt)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, err.Error()), nil
	}
	if task.Shape != domain.CrewShip || (progress.Stage != stageReviewed && progress.Stage != stageDelivered) || progress.ReviewedCommit == "" || progress.ReviewedCommit != progress.CandidateCommit {
		return failed("deliver-unreviewed", "the candidate is not reviewed at its current commit"), nil
	}
	paths, err := executor.git.CommitPaths(ctx, executor.root, progress.BaseCommit, progress.CandidateCommit)
	if err != nil {
		return failed("deliver-paths", err.Error()), nil
	}
	for _, relative := range paths {
		if !domain.ScopeContains(task.AllowedPaths, relative) || protected(relative) {
			return decision(domain.CrewDecisionScopeExpanded, "the candidate changes "+relative+", which is outside the approved scope or protected"), nil
		}
	}
	client, err := executor.forge(ctx, plan.Remote)
	if err != nil {
		return decision(domain.CrewDecisionBlocked, err.Error()), nil
	}
	branch := crewTaskBranch(task, checkpoint.Attempt)
	pulls, err := client.FindByHead(ctx, branch)
	if err != nil {
		return failed("forge-list", err.Error()), nil
	}
	var open *forge.PullRequest
	closed := 0
	for index := range pulls {
		switch pulls[index].State {
		case "OPEN":
			open = &pulls[index]
		case "MERGED":
			if pulls[index].Head == progress.CandidateCommit {
				progress.Stage = stageMerged
				_ = executor.saveCrewProgress(plan, &progress)
				return crew.Outcome{Kind: crew.OutcomeMerged, CandidateCommit: progress.CandidateCommit, Message: "pull request #" + strconv.Itoa(pulls[index].Number) + " was already merged"}, nil
			}
		default:
			closed = pulls[index].Number
		}
	}
	if open == nil && closed != 0 {
		return decision(domain.CrewDecisionBlocked, "pull request #"+strconv.Itoa(closed)+" for this task was closed; Level 7 does not reopen it"), nil
	}
	result, err := executor.gitResult(ctx, executor.root, remoteEnvironment(), "push", "--porcelain", plan.Remote, progress.CandidateCommit+":refs/heads/"+branch)
	if err != nil {
		return failed("push", err.Error()), nil
	}
	if result.ExitCode != 0 {
		return decision(domain.CrewDecisionBlocked, "pushing "+branch+" to "+plan.Remote+" was rejected: "+gitDiagnostic(result.Stderr, nil)), nil
	}
	message := "updated pull request"
	if open == nil {
		created, err := client.Create(ctx, plan.TargetBranch, branch, crewCommitSubject(task), crewHandoff(plan, task, progress, paths))
		if err != nil {
			return failed("pull-request", err.Error()), nil
		}
		open, message = &created, "opened pull request"
		if err := client.AddLabel(ctx, created.Number, crewRiskLabel); err != nil {
			message += "; the " + crewRiskLabel + " label was not applied"
		}
	}
	progress.Stage, progress.PullRequest, progress.PullRequestURL = stageDelivered, open.Number, open.URL
	if err := executor.saveCrewProgress(plan, &progress); err != nil {
		return failed("worker-checkpoint", err.Error()), nil
	}
	return delivered(progress, fmt.Sprintf("%s #%d at %s", message, open.Number, short(progress.CandidateCommit))), nil
}

func delivered(progress crewProgress, message string) crew.Outcome {
	return crew.Outcome{
		Kind: crew.OutcomeDelivered, PullRequest: progress.PullRequest, PullRequestURL: progress.PullRequestURL,
		CandidateCommit: progress.CandidateCommit, Worktree: progress.Worktree, Verification: "passed",
		ProviderID: progress.ImplementationRoute.ProviderID, ModelID: progress.ImplementationRoute.ModelID, SessionID: progress.ImplementationSession,
		Message: message,
	}
}

// ciRepair turns failed pull-request checks into repair feedback for the
// implementer, including the end of up to two failed job logs.
func (executor CrewExecutor) ciRepair(ctx context.Context, plan domain.CrewPlan, progress *crewProgress, checks string) (crew.Outcome, bool) {
	feedback := fmt.Sprintf("Checks failed on pull request #%d at %s: %s", progress.PullRequest, short(progress.CandidateCommit), strings.TrimPrefix(checks, "failed: "))
	if client, err := executor.forge(ctx, plan.Remote); err == nil {
		if pull, err := client.View(ctx, progress.PullRequest); err == nil {
			added := 0
			for _, check := range pull.Checks {
				if check.Status != "failed" || added == 2 {
					continue
				}
				if log := client.FailedLog(ctx, check, ciLogBytes); log != "" {
					feedback += "\n\nEnd of the failed log for " + check.Name + ":\n" + log
					added++
				}
			}
		}
	}
	progress.Stage, progress.RepairRounds, progress.ReviewedCommit, progress.Feedback = stageRepairing, 0, "", bounded(feedback, maxFeedbackBytes)
	if err := executor.saveCrewProgress(plan, progress); err != nil {
		return failed("worker-checkpoint", err.Error()), true
	}
	return crew.Outcome{}, false
}

// Track reads the pull request of a delivered task.
func (executor CrewExecutor) Track(ctx context.Context, plan domain.CrewPlan, _ domain.CrewTask, checkpoint domain.CrewCheckpoint) (crew.PullRequestStatus, error) {
	if checkpoint.PullRequest < 1 {
		return crew.PullRequestStatus{}, errors.New("the task has no pull request")
	}
	client, err := executor.forge(ctx, plan.Remote)
	if err != nil {
		return crew.PullRequestStatus{}, err
	}
	pull, err := client.View(ctx, checkpoint.PullRequest)
	if err != nil {
		return crew.PullRequestStatus{}, err
	}
	summary, failing := forge.Summarize(pull.Checks)
	return crew.PullRequestStatus{State: pull.State, Head: pull.Head, Checks: checkSummary(summary, failing, pull.Checks), Failed: summary == "failed"}, nil
}

// MergePullRequest merges the task's pull request at head after checking
// every precondition again against the forge. It never bypasses branch
// protection.
func (executor CrewExecutor) MergePullRequest(ctx context.Context, plan domain.CrewPlan, checkpoint domain.CrewCheckpoint, head string) (crew.Outcome, error) {
	switch {
	case plan.Delivery != domain.CrewDeliveryPR:
		return crew.Outcome{}, errors.New("this crew delivers to a local branch; there is no pull request to merge")
	case checkpoint.State != domain.CrewPROpen || checkpoint.PullRequest < 1:
		return crew.Outcome{}, errors.New("the task has no open pull request")
	case head != checkpoint.CandidateCommit:
		return crew.Outcome{}, fmt.Errorf("head %s is not the delivered candidate %s", short(head), short(checkpoint.CandidateCommit))
	}
	client, err := executor.forge(ctx, plan.Remote)
	if err != nil {
		return crew.Outcome{}, err
	}
	pull, err := client.View(ctx, checkpoint.PullRequest)
	if err != nil {
		return crew.Outcome{}, err
	}
	summary, failing := forge.Summarize(pull.Checks)
	switch {
	case pull.State != "OPEN":
		return crew.Outcome{}, fmt.Errorf("pull request #%d is %s", pull.Number, strings.ToLower(pull.State))
	case pull.Draft:
		return crew.Outcome{}, fmt.Errorf("pull request #%d is a draft", pull.Number)
	case pull.Head != head:
		return crew.Outcome{}, fmt.Errorf("pull request #%d head is %s, not %s", pull.Number, short(pull.Head), short(head))
	case summary != "passed":
		return crew.Outcome{}, fmt.Errorf("pull request #%d checks are not all passing: %s", pull.Number, checkSummary(summary, failing, pull.Checks))
	case pull.MergeState != "CLEAN":
		return crew.Outcome{}, fmt.Errorf("GitHub reports pull request #%d as %s, not clean to merge", pull.Number, strings.ToLower(pull.MergeState))
	}
	if err := client.Merge(ctx, pull.Number, executor.configuration.EffectiveCrew().MergeMethod, head); err != nil {
		return crew.Outcome{}, err
	}
	merged, err := client.View(ctx, pull.Number)
	if err != nil || merged.State != "MERGED" {
		return crew.Outcome{}, fmt.Errorf("gh reported a merge, but pull request #%d is not merged", pull.Number)
	}
	return crew.Outcome{Kind: crew.OutcomeMerged, CandidateCommit: head, Message: fmt.Sprintf("merged pull request #%d at %s", pull.Number, short(head))}, nil
}

func checkSummary(summary string, failing []string, checks []forge.Check) string {
	switch summary {
	case "failed":
		return "failed: " + bounded(strings.Join(failing, ", "), 256)
	case "pending":
		pending := 0
		for _, check := range checks {
			if check.Status == "pending" {
				pending++
			}
		}
		return fmt.Sprintf("pending: %d of %d", pending, len(checks))
	case "passed":
		return fmt.Sprintf("passed: %d", len(checks))
	default:
		return "none"
	}
}

func crewTaskBranch(task domain.CrewTask, attempt int) string {
	return fmt.Sprintf("l7/tasks/%s-a%d", task.ID, attempt)
}

// crewHandoff is the pull-request description: what was asked, what was
// checked and by whom, and what was not verified.
func crewHandoff(plan domain.CrewPlan, task domain.CrewTask, progress crewProgress, paths []string) string {
	var body strings.Builder
	fmt.Fprintf(&body, "## Level 7 crew task `%s`\n\n**%s**\n\n%s\n\n### Acceptance criteria\n\n", task.ID, task.Title, bounded(strings.TrimSpace(task.Objective), 4<<10))
	for _, criterion := range task.AcceptanceCriteria {
		fmt.Fprintf(&body, "- %s\n", criterion)
	}
	fmt.Fprintf(&body, "\n### Evidence\n\n- Verification passed at `%s`:\n", short(progress.CandidateCommit))
	for _, argv := range task.Verification {
		fmt.Fprintf(&body, "  - `%s`\n", strings.Join(argv, " "))
	}
	fmt.Fprintf(&body, "- Review: GO from `%s/%s`, a model that did not implement this task (implemented by `%s`).\n",
		progress.ReviewRoute.ProviderID, progress.ReviewRoute.ModelID, strings.Join(progress.Contributors, "`, `"))
	fmt.Fprintf(&body, "- Changed paths, all inside the approved scope `%s`:\n", strings.Join(task.AllowedPaths, "`, `"))
	for _, relative := range paths {
		fmt.Fprintf(&body, "  - `%s`\n", relative)
	}
	fmt.Fprintf(&body, "\n### Not verified here\n\n- This repository's checks on this pull request. Level 7 tracks them and merges only on an explicit `l7 crew merge` at an exact head.\n\nOpened by the Level 7 crew from plan `%s` (Tier 2 ceiling).\n", plan.ID)
	return bounded(body.String(), maxHandoff)
}

// remoteEnvironment lets git reach a remote the way the owner's shell does,
// including an SSH agent, without passing any other secret.
func remoteEnvironment() []string {
	values := processadapter.MinimalEnvironment()
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" && !strings.ContainsAny(socket, "\x00\r\n") {
		values = append(values, "SSH_AUTH_SOCK="+socket)
	}
	return values
}

func gitDiagnostic(stderr []byte, err error) string {
	message := strings.Join(strings.Fields(strings.ToValidUTF8(string(stderr), "")), " ")
	if message == "" && err != nil {
		message = err.Error()
	}
	if message == "" {
		message = "no diagnostic"
	}
	return bounded(message, 512)
}
