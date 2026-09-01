// Package headlessworker composes routing, disposable worktrees, provider
// sessions, verification, independent review, and exact-ref local merges.
package headlessworker

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	claudeadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/claude"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/codexapp"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/gateway"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/headless"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/memory"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/toolbroker"
	verifyadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/verify"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type Executor struct {
	root          string
	common        string
	configuration orchestrationconfig.File
	git           gitadapter.Adapter
}

type providerResult struct {
	SessionID string
	Summary   string
	Decision  domain.ReviewDecision
	Quota     bool
	Reset     string
}

type workerProgress struct {
	Schema                 int                  `json:"schema"`
	ManifestDigest         string               `json:"manifest_digest"`
	WaveID                 string               `json:"wave_id"`
	Sequence               int                  `json:"sequence"`
	Stage                  string               `json:"stage"`
	TargetCommit           string               `json:"target_commit"`
	Worktree               string               `json:"worktree"`
	ImplementationRoute    domain.RouteDecision `json:"implementation_route"`
	ImplementationSession  string               `json:"implementation_session"`
	ImplementationFailures int                  `json:"implementation_failures"`
	CandidateCommit        string               `json:"candidate_commit"`
	ReviewRoute            domain.RouteDecision `json:"review_route"`
	ReviewSession          string               `json:"review_session"`
	ReviewFailures         int                  `json:"review_failures"`
	ReviewApproved         bool                 `json:"review_approved"`
	Merged                 bool                 `json:"merged"`
	UpdatedAtUTC           string               `json:"updated_at_utc"`
	Next                   string               `json:"next"`
}

type workerEvent struct {
	Schema       int    `json:"schema"`
	Manifest     string `json:"manifest_digest"`
	WaveID       string `json:"wave_id"`
	Sequence     int    `json:"sequence"`
	Stage        string `json:"stage"`
	CreatedAtUTC string `json:"created_at_utc"`
	Next         string `json:"next"`
}

func New(root, common string, configuration orchestrationconfig.File) (Executor, error) {
	if !filepath.IsAbs(root) || !filepath.IsAbs(common) || configuration.Validate() != nil {
		return Executor{}, errors.New("Headless executor configuration is invalid")
	}
	gitClient, err := gitadapter.New("", gitadapter.DefaultMaxOutput, gitadapter.DefaultMaxPaths)
	if err != nil {
		return Executor{}, err
	}
	return Executor{root: root, common: common, configuration: configuration, git: gitClient}, nil
}

func (executor Executor) Execute(ctx context.Context, manifest domain.HeadlessManifest, wave domain.HeadlessWave, checkpoint domain.HeadlessCheckpoint) (headless.WaveOutcome, error) {
	if !manifest.LocalMerge || manifest.RiskCeiling != domain.TierProduct || !manifest.StopBeforeDeploy || len(wave.AcceptanceCriteria) == 0 || len(wave.Verification) == 0 {
		return headless.WaveOutcome{State: "tier3", Message: "wave exceeds the approved Tier 2 Headless contract"}, nil
	}
	target, err := executor.ref(ctx, manifest.TargetBranch)
	if err != nil {
		return executor.failure("target-ref", err), nil
	}
	progress, err := executor.loadProgress(manifest, wave)
	if err != nil {
		return executor.failure("worker-checkpoint", err), nil
	}
	if progress.TargetCommit != "" && progress.TargetCommit != target && progress.CandidateCommit != target {
		return headless.WaveOutcome{State: "scope-expanded", Message: "target branch diverged from the durable wave checkpoint"}, nil
	}
	if progress.TargetCommit == "" && checkpoint.Sequence <= 1 && target != manifest.BaseCommit {
		return headless.WaveOutcome{State: "scope-expanded", Message: "target branch diverged from the approved base"}, nil
	}
	if progress.Merged && target == progress.CandidateCommit {
		return progress.completeOutcome(), nil
	}
	base := target
	if progress.TargetCommit != "" {
		base = progress.TargetCommit
	}
	worktree, err := executor.worktree(ctx, manifest, wave, base, checkpoint, progress.Worktree)
	if err != nil {
		return executor.failure("worktree", err), nil
	}
	if progress.Worktree == "" {
		progress.TargetCommit, progress.Worktree = base, worktree
		if err := executor.transition(manifest, wave, &progress, "worktree-ready", "route the implementation assignment"); err != nil {
			return executor.failure("worker-checkpoint", err), nil
		}
	}
	snapshots, found, err := state.LoadProviderSnapshots(executor.common)
	if err != nil || !found {
		return executor.failure("provider-snapshot", errors.New("verified provider snapshot is unavailable")), nil
	}
	implementationTask := domain.TaskProfile{
		Schema: domain.OrchestrationSchema, ID: manifest.ID + "/" + wave.ID + "/implement", Summary: wave.Title,
		Complexity: domain.ComplexityC3, RiskTier: domain.TierProduct, ContextTokens: 64_000,
		NeedsTools: true, NeedsEditing: true, NeedsResume: true, PriorFailures: progress.ImplementationFailures,
	}
	implementationRoute := routeForAttempt(implementationTask, snapshots, progress.ImplementationFailures)
	if progress.ImplementationRoute.ProviderID == "" || !sameRoute(progress.ImplementationRoute, implementationRoute) {
		implementationRoute.DecisionUTC = time.Now().UTC().Format(time.RFC3339)
		if implementationRoute.ProviderID == "" {
			return headless.WaveOutcome{State: "blocked", Message: "no qualified implementation route"}, nil
		}
		if progress.ImplementationRoute.ProviderID != "" {
			progress.ImplementationSession = ""
		}
		progress.ImplementationRoute = implementationRoute
		_ = state.SaveRouteDecision(executor.common, implementationRoute)
		stage := "implementation-routed"
		if progress.ImplementationFailures > 0 {
			stage = "implementation-failover"
		}
		if err := executor.transition(manifest, wave, &progress, stage, "start or resume the implementation session"); err != nil {
			return executor.failure("worker-checkpoint", err), nil
		}
	}
	implementationRoute = progress.ImplementationRoute
	if progress.CandidateCommit == "" {
		if candidate, recovered := executor.recoverCandidate(ctx, worktree, base, wave, manifest.AllowedPaths); recovered {
			progress.CandidateCommit = candidate
			if err := executor.transition(manifest, wave, &progress, "candidate-committed", "route an independent read-only reviewer"); err != nil {
				return executor.failureWithRoute("worker-checkpoint", implementationRoute, progress.ImplementationSession, worktree, err), nil
			}
		} else {
			session := progress.ImplementationSession
			if session == "" && checkpoint.ProviderID == implementationRoute.ProviderID && checkpoint.ModelID == implementationRoute.ModelID {
				session = checkpoint.SessionID
			}
			implemented, runErr := executor.runProvider(ctx, worktree, implementationRoute, implementationPrompt(manifest, wave), session, false, manifest.AllowedPaths, manifest.AllowedCommands)
			if implemented.SessionID != "" {
				progress.ImplementationSession = implemented.SessionID
			}
			if implemented.Quota && implemented.Reset != "" {
				_ = executor.transition(manifest, wave, &progress, "implementation-quota", "wait for the natural provider quota reset")
				return headless.WaveOutcome{State: "quota", ProviderID: implementationRoute.ProviderID, ModelID: implementationRoute.ModelID, SessionID: implemented.SessionID, Worktree: worktree, QuotaResetAtUTC: implemented.Reset, Message: "provider reported a natural quota reset"}, nil
			}
			if runErr != nil {
				progress.ImplementationFailures++
				if err := executor.transition(manifest, wave, &progress, "implementation-failed", "retry with the next qualified provider fallback"); err != nil {
					return executor.failureWithRoute("worker-checkpoint", implementationRoute, implemented.SessionID, worktree, err), nil
				}
				return executor.providerFailure("implementation", implementationRoute, implemented, worktree, runErr), nil
			}
			if err := executor.transition(manifest, wave, &progress, "implementation-complete", "validate scope and run exact verification commands"); err != nil {
				return executor.failureWithRoute("worker-checkpoint", implementationRoute, implemented.SessionID, worktree, err), nil
			}
			pending, pendingErr := executor.git.Pending(ctx, worktree)
			if pendingErr != nil {
				return executor.failureWithRoute("pending", implementationRoute, implemented.SessionID, worktree, pendingErr), nil
			}
			if pending.IndexDirty {
				return headless.WaveOutcome{State: "scope-expanded", ProviderID: implementationRoute.ProviderID, ModelID: implementationRoute.ModelID, SessionID: implemented.SessionID, Worktree: worktree, Message: "provider changed the Git index outside the controlled commit transition"}, nil
			}
			if len(pending.Paths) == 0 {
				progress.ImplementationFailures++
				_ = executor.transition(manifest, wave, &progress, "implementation-no-progress", "retry with the next qualified provider fallback")
				return executor.failureWithRoute("no-change", implementationRoute, implemented.SessionID, worktree, errors.New("provider completed without repository progress")), nil
			}
			for _, relative := range pending.Paths {
				if !allowed(relative, manifest.AllowedPaths) || protected(relative) {
					return headless.WaveOutcome{State: "scope-expanded", ProviderID: implementationRoute.ProviderID, ModelID: implementationRoute.ModelID, SessionID: implemented.SessionID, Worktree: worktree, Message: "provider changed a path outside the approved manifest: " + relative}, nil
				}
			}
			checks, verifyErr := verifyadapter.New(nil, nil).Run(ctx, worktree, verificationCommands(wave.Verification), executor.configuration.Tools.MaxOutputBytes, executor.configuration.Tools.MaxSeconds)
			if verifyErr != nil || !allPassed(checks) {
				progress.ImplementationFailures++
				_ = executor.transition(manifest, wave, &progress, "verification-failed", "retry with the next qualified provider fallback")
				return executor.failureWithRoute("verification", implementationRoute, implemented.SessionID, worktree, errors.New("wave verification failed")), nil
			}
			if err := executor.transition(manifest, wave, &progress, "verification-complete", "create the controlled candidate commit"); err != nil {
				return executor.failureWithRoute("worker-checkpoint", implementationRoute, implemented.SessionID, worktree, err), nil
			}
			candidate, commitErr := executor.git.Commit(ctx, domain.CommitRequest{
				Root: worktree, ExpectedCommit: pending.Head, ExpectedTree: pending.Tree, Paths: pending.Paths,
				Message: "feat(headless): complete " + wave.ID, MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes,
				MaxPaths: gitadapter.DefaultMaxPaths, MaxCommandSeconds: executor.configuration.Tools.MaxSeconds,
			})
			if commitErr != nil {
				return executor.failureWithRoute("commit", implementationRoute, implemented.SessionID, worktree, commitErr), nil
			}
			progress.CandidateCommit = candidate.Head
			if err := executor.transition(manifest, wave, &progress, "candidate-committed", "route an independent read-only reviewer"); err != nil {
				return executor.failureWithRoute("worker-checkpoint", implementationRoute, implemented.SessionID, worktree, err), nil
			}
		}
	}
	reviewTask := domain.TaskProfile{
		Schema: domain.OrchestrationSchema, ID: manifest.ID + "/" + wave.ID + "/review", Summary: "independent review " + wave.Title,
		Complexity: domain.ComplexityC3, RiskTier: domain.TierProduct, ContextTokens: 64_000,
		NeedsTools: true, IndependentReview: true, ImplementerProvider: implementationRoute.ProviderID, ImplementerModel: implementationRoute.ModelID,
		PriorFailures: progress.ReviewFailures,
	}
	reviewRoute := routeForAttempt(reviewTask, snapshots, progress.ReviewFailures)
	if progress.ReviewRoute.ProviderID == "" || !sameRoute(progress.ReviewRoute, reviewRoute) {
		reviewRoute.DecisionUTC = time.Now().UTC().Format(time.RFC3339)
		if reviewRoute.ProviderID == "" {
			return headless.WaveOutcome{State: "blocked", ProviderID: implementationRoute.ProviderID, ModelID: implementationRoute.ModelID, SessionID: progress.ImplementationSession, Worktree: worktree, CandidateCommit: progress.CandidateCommit, Message: "no independent qualified reviewer is available"}, nil
		}
		if progress.ReviewRoute.ProviderID != "" {
			progress.ReviewSession = ""
		}
		progress.ReviewRoute = reviewRoute
		stage := "review-routed"
		if progress.ReviewFailures > 0 {
			stage = "review-failover"
		}
		if err := executor.transition(manifest, wave, &progress, stage, "start or resume the independent read-only review"); err != nil {
			return executor.failureWithRoute("worker-checkpoint", reviewRoute, progress.ReviewSession, worktree, err), nil
		}
	}
	reviewRoute = progress.ReviewRoute
	if !progress.ReviewApproved {
		reviewed, reviewErr := executor.runProvider(ctx, worktree, reviewRoute, reviewPrompt(manifest, wave, base, progress.CandidateCommit), progress.ReviewSession, true, manifest.AllowedPaths, manifest.AllowedCommands)
		if reviewed.SessionID != "" {
			progress.ReviewSession = reviewed.SessionID
		}
		if reviewed.Quota && reviewed.Reset != "" {
			_ = executor.transition(manifest, wave, &progress, "review-quota", "wait for the natural reviewer quota reset")
			return headless.WaveOutcome{State: "quota", ProviderID: reviewRoute.ProviderID, ModelID: reviewRoute.ModelID, SessionID: reviewed.SessionID, Worktree: worktree, CandidateCommit: progress.CandidateCommit, QuotaResetAtUTC: reviewed.Reset, Message: "reviewer reported a natural quota reset"}, nil
		}
		if reviewErr != nil || reviewed.Decision != domain.DecisionGO {
			progress.ReviewFailures++
			if err := executor.transition(manifest, wave, &progress, "review-failed", "retry with the next independent qualified provider fallback"); err != nil {
				return executor.failureWithRoute("worker-checkpoint", reviewRoute, reviewed.SessionID, worktree, err), nil
			}
			return executor.providerFailure("independent-review", reviewRoute, reviewed, worktree, errors.New("independent reviewer did not return GO")), nil
		}
		progress.ReviewApproved = true
		if err := executor.transition(manifest, wave, &progress, "review-approved", "verify reviewer immutability and compare-and-swap the local target"); err != nil {
			return executor.failureWithRoute("worker-checkpoint", reviewRoute, reviewed.SessionID, worktree, err), nil
		}
	}
	afterReview, err := executor.git.Pending(ctx, worktree)
	if err != nil || len(afterReview.Paths) != 0 || afterReview.IndexDirty || afterReview.Head != progress.CandidateCommit {
		return headless.WaveOutcome{State: "scope-expanded", ProviderID: reviewRoute.ProviderID, ModelID: reviewRoute.ModelID, SessionID: progress.ReviewSession, Worktree: worktree, CandidateCommit: progress.CandidateCommit, Message: "independent reviewer mutated the candidate"}, nil
	}
	merge := domain.MergeRequest{Root: executor.root, TargetBranch: manifest.TargetBranch, ExpectedOld: base, Candidate: progress.CandidateCommit, MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes}
	mergeTarget, err := executor.git.InspectMerge(ctx, merge)
	if err != nil {
		return headless.WaveOutcome{State: "scope-expanded", ProviderID: reviewRoute.ProviderID, ModelID: reviewRoute.ModelID, SessionID: progress.ReviewSession, Worktree: worktree, CandidateCommit: progress.CandidateCommit, Message: "target branch diverged or is checked out; local merge paused"}, nil
	}
	if !mergeTarget.AlreadyAdvanced {
		if err := executor.git.AdvanceMerge(ctx, merge); err != nil {
			return executor.failureWithRoute("cas-merge", reviewRoute, progress.ReviewSession, worktree, err), nil
		}
	}
	progress.Merged = true
	if err := executor.transition(manifest, wave, &progress, "locally-merged", "advance to the next approved wave or stop before release"); err != nil {
		return executor.failureWithRoute("worker-checkpoint", reviewRoute, progress.ReviewSession, worktree, err), nil
	}
	return progress.completeOutcome(), nil
}

func (executor Executor) worktree(ctx context.Context, manifest domain.HeadlessManifest, wave domain.HeadlessWave, target string, checkpoint domain.HeadlessCheckpoint, recorded string) (string, error) {
	base := filepath.Join(executor.common, "l7", "headless", "worktrees")
	if err := localfile.EnsureDirectory(base, 0o700); err != nil {
		return "", err
	}
	name := manifest.ID + "-" + wave.ID
	worktree := filepath.Join(base, name)
	branch := "codex/l7-headless-" + manifest.ID + "-" + wave.ID
	for _, candidate := range []string{recorded, checkpoint.Worktree, worktree} {
		if candidate == "" {
			continue
		}
		physical, resolveErr := filepath.EvalSymlinks(candidate)
		if resolveErr == nil && physical == filepath.Clean(candidate) && executor.validWorktree(ctx, physical, branch, target) {
			return physical, nil
		}
	}
	if _, err := os.Lstat(worktree); err == nil {
		return "", errors.New("existing Headless worktree does not match the durable wave identity")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	git, err := processadapter.Resolve("git")
	if err != nil {
		return "", err
	}
	result, runErr := (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: git.Path, Arguments: []string{"worktree", "add", "-b", branch, worktree, target}, Directory: executor.root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: executor.configuration.Tools.MaxOutputBytes, Timeout: time.Duration(executor.configuration.Tools.MaxSeconds) * time.Second,
	})
	if runErr != nil || result.ExitCode != 0 {
		return "", errors.New("cannot create disposable Headless worktree")
	}
	return filepath.EvalSymlinks(worktree)
}

func (executor Executor) validWorktree(ctx context.Context, worktree, branch, target string) bool {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return false
	}
	runner := processadapter.Runner{}
	request := func(arguments ...string) (processadapter.Result, error) {
		return runner.Run(ctx, processadapter.Request{Executable: git.Path, Arguments: arguments, Directory: worktree, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second})
	}
	root, err := request("rev-parse", "--show-toplevel")
	if err != nil || root.ExitCode != 0 || strings.TrimSpace(string(root.Stdout)) != worktree {
		return false
	}
	current, err := request("symbolic-ref", "--short", "HEAD")
	if err != nil || current.ExitCode != 0 || strings.TrimSpace(string(current.Stdout)) != branch {
		return false
	}
	ancestor, err := request("merge-base", "--is-ancestor", target, "HEAD")
	return err == nil && ancestor.ExitCode == 0
}

func (executor Executor) loadProgress(manifest domain.HeadlessManifest, wave domain.HeadlessWave) (workerProgress, error) {
	path, err := executor.progressPath(manifest.ID, wave.ID, "checkpoint.json")
	if err != nil {
		return workerProgress{}, err
	}
	data, err := localfile.Read(path, 4<<20)
	if errors.Is(err, os.ErrNotExist) {
		return workerProgress{Schema: domain.OrchestrationSchema, ManifestDigest: manifest.Digest, WaveID: wave.ID, ImplementationRoute: emptyRoute(), ReviewRoute: emptyRoute()}, nil
	}
	if err != nil {
		return workerProgress{}, err
	}
	var progress workerProgress
	if err := localfile.DecodeJSON(data, &progress); err != nil {
		return workerProgress{}, fmt.Errorf("decode durable Headless wave checkpoint: %w", err)
	}
	if progress.Schema != domain.OrchestrationSchema || progress.ManifestDigest != manifest.Digest || progress.WaveID != wave.ID || progress.Sequence < 0 || progress.ImplementationFailures < 0 || progress.ImplementationFailures > 4096 || progress.ReviewFailures < 0 || progress.ReviewFailures > 4096 {
		return workerProgress{}, errors.New("durable Headless wave checkpoint is invalid or stale")
	}
	return progress, nil
}

func emptyRoute() domain.RouteDecision {
	return domain.RouteDecision{Candidates: []domain.RouteCandidate{}, Fallbacks: []string{}, Escalations: []string{}}
}

func (executor Executor) transition(manifest domain.HeadlessManifest, wave domain.HeadlessWave, progress *workerProgress, stage, next string) error {
	if progress == nil || stage == "" || next == "" || strings.ContainsAny(stage+next, "\x00\r\n") {
		return errors.New("Headless worker transition is invalid")
	}
	progress.Schema, progress.ManifestDigest, progress.WaveID = domain.OrchestrationSchema, manifest.Digest, wave.ID
	progress.Sequence++
	progress.Stage, progress.Next, progress.UpdatedAtUTC = stage, next, time.Now().UTC().Format(time.RFC3339)
	checkpointPath, err := executor.progressPath(manifest.ID, wave.ID, "checkpoint.json")
	if err != nil {
		return err
	}
	if err := writePrivateJSON(checkpointPath, progress, true); err != nil {
		return err
	}
	eventPath, err := executor.progressPath(manifest.ID, wave.ID, fmt.Sprintf("events/%08d.json", progress.Sequence))
	if err != nil {
		return err
	}
	event := workerEvent{Schema: domain.OrchestrationSchema, Manifest: manifest.Digest, WaveID: wave.ID, Sequence: progress.Sequence, Stage: stage, CreatedAtUTC: progress.UpdatedAtUTC, Next: next}
	return writePrivateJSON(eventPath, event, false)
}

func (executor Executor) progressPath(runID, waveID, relative string) (string, error) {
	for _, value := range []string{runID, waveID} {
		if value == "" || value == "." || value == ".." || strings.ContainsAny(value, "/\\\x00\r\n") {
			return "", errors.New("Headless worker identity is unsafe")
		}
	}
	root := filepath.Join(executor.common, "l7", "headless", runID, "waves", waveID)
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func writePrivateJSON(path string, value any, replace bool) error {
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > 4<<20 {
		return errors.New("Headless worker record is invalid or unbounded")
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	} else if err != nil {
		return err
	} else if !replace {
		return os.ErrExist
	}
	return localfile.AtomicReplace(path, data, 0o600)
}

func (progress workerProgress) completeOutcome() headless.WaveOutcome {
	return headless.WaveOutcome{State: "complete", ProviderID: progress.ImplementationRoute.ProviderID, ModelID: progress.ImplementationRoute.ModelID, SessionID: progress.ImplementationSession, Worktree: progress.Worktree, CandidateCommit: progress.CandidateCommit, Message: "wave implemented, verified, independently reviewed, and locally merged"}
}

func (executor Executor) recoverCandidate(ctx context.Context, worktree, base string, wave domain.HeadlessWave, allowedPaths []string) (string, bool) {
	pending, err := executor.git.Pending(ctx, worktree)
	if err != nil || pending.IndexDirty || len(pending.Paths) != 0 || pending.Head == base || len(pending.Head) != 40 {
		return "", false
	}
	parent, err := executor.gitLine(ctx, worktree, "rev-parse", pending.Head+"^")
	if err != nil || parent != base {
		return "", false
	}
	subject, err := executor.gitLine(ctx, worktree, "show", "-s", "--format=%s", pending.Head)
	if err != nil || subject != "feat(headless): complete "+wave.ID {
		return "", false
	}
	git, err := processadapter.Resolve("git")
	if err != nil {
		return "", false
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{Executable: git.Path, Arguments: []string{"diff", "--name-only", "-z", base, pending.Head, "--"}, Directory: worktree, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 8 << 20, Timeout: 30 * time.Second})
	if err != nil || result.ExitCode != 0 {
		return "", false
	}
	paths := nulPaths(result.Stdout)
	if len(paths) == 0 {
		return "", false
	}
	for _, relative := range paths {
		if !allowed(relative, allowedPaths) || protected(relative) {
			return "", false
		}
	}
	return pending.Head, true
}

func (executor Executor) gitLine(ctx context.Context, root string, arguments ...string) (string, error) {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return "", err
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{Executable: git.Path, Arguments: arguments, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second})
	value := strings.TrimSpace(string(result.Stdout))
	if err != nil || result.ExitCode != 0 || value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", errors.New("Git recovery query failed")
	}
	return value, nil
}

func nulPaths(data []byte) []string {
	result := []string{}
	for _, value := range strings.Split(string(data), "\x00") {
		if value != "" {
			result = append(result, filepath.ToSlash(value))
		}
	}
	return result
}

func (executor Executor) runProvider(ctx context.Context, root string, route domain.RouteDecision, prompt, session string, reviewer bool, scope []string, commands [][]string) (providerResult, error) {
	snapshots, _, err := state.LoadProviderSnapshots(executor.common)
	if err != nil {
		return providerResult{}, err
	}
	var snapshot domain.ProviderSnapshot
	for _, candidate := range snapshots {
		if candidate.ID == route.ProviderID {
			snapshot = candidate
			break
		}
	}
	switch snapshot.Kind {
	case domain.ProviderKindCodexAppServer:
		result, runErr := codexapp.Run(ctx, codexapp.Assignment{Executable: snapshot.Executable, Root: root, Model: route.ModelID, Effort: route.Effort, Prompt: prompt, SessionID: session, Reviewer: reviewer})
		return providerResult{SessionID: result.SessionID, Summary: result.Summary, Decision: result.Decision, Quota: result.FailureCode == "quota", Reset: result.QuotaReset}, runErr
	case domain.ProviderKindClaudeCLI:
		result, runErr := claudeadapter.RunSession(ctx, claudeadapter.SessionAssignment{Executable: snapshot.Executable, Root: root, Model: route.ModelID, Effort: route.Effort, Prompt: prompt, SessionID: session, Reviewer: reviewer})
		return providerResult{SessionID: result.SessionID, Summary: result.Summary, Decision: result.Decision, Quota: result.Quota, Reset: result.QuotaReset}, runErr
	case domain.ProviderKindOpenAIResponses, domain.ProviderKindAnthropic:
		provider, ok := configuredProvider(executor.configuration, route.ProviderID)
		if !ok {
			return providerResult{}, errors.New("gateway provider configuration is unavailable")
		}
		if !configuredModel(provider, route.ModelID) {
			model, verified := snapshotModel(snapshot, route.ModelID)
			if !verified {
				return providerResult{}, errors.New("gateway model is not bound to the verified provider snapshot")
			}
			provider.Models = append(provider.Models, model)
		}
		policy := executor.configuration.Tools
		policy.AllowedPaths = append([]string{}, scope...)
		policy.AllowedCommands = manifestCommands(commands)
		query := func(value string) ([]byte, error) {
			matches, queryErr := memory.New(memory.NewAppleEmbedder()).Query(ctx, executor.common, value, 20)
			if queryErr != nil {
				return nil, queryErr
			}
			return json.Marshal(matches)
		}
		var broker toolbroker.Broker
		if reviewer {
			broker, err = toolbroker.NewReadOnly(root, policy, query)
		} else {
			broker, err = toolbroker.New(root, policy, query)
		}
		if err != nil {
			return providerResult{}, err
		}
		result, runErr := gateway.New().Run(ctx, gateway.Assignment{Provider: provider, Model: route.ModelID, Effort: route.Effort, Prompt: prompt, Reviewer: reviewer}, broker)
		providerResult := providerResult{Summary: result.Summary, Decision: result.Decision}
		var quota *gateway.QuotaError
		if errors.As(runErr, &quota) {
			providerResult.Quota, providerResult.Reset = true, quota.ResetAtUTC
		}
		return providerResult, runErr
	default:
		return providerResult{}, errors.New("selected provider kind is unsupported")
	}
}

func manifestCommands(commands [][]string) []orchestrationconfig.Command {
	result := make([]orchestrationconfig.Command, 0, len(commands))
	for index, argv := range commands {
		result = append(result, orchestrationconfig.Command{Name: fmt.Sprintf("verify-%02d", index+1), Argv: append([]string{}, argv...)})
	}
	return result
}

func configuredProvider(configuration orchestrationconfig.File, id string) (orchestrationconfig.Provider, bool) {
	for _, provider := range configuration.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return orchestrationconfig.Provider{}, false
}

func configuredModel(provider orchestrationconfig.Provider, id string) bool {
	for _, model := range provider.Models {
		if model.ID == id {
			return true
		}
	}
	return false
}

func snapshotModel(snapshot domain.ProviderSnapshot, id string) (orchestrationconfig.Model, bool) {
	for _, model := range snapshot.Models {
		if model.ID != id || !model.Verified {
			continue
		}
		return orchestrationconfig.Model{
			ID: model.ID, Languages: append([]string{}, model.Languages...), ContextWindow: model.ContextWindow,
			SupportsTools: model.SupportsTools, SupportsEditing: model.SupportsEditing, SupportsResume: model.SupportsResume,
			Efforts: append([]domain.ReasoningEffort{}, model.Efforts...), CostClass: model.CostClass, LatencyClass: model.LatencyClass,
		}, true
	}
	return orchestrationconfig.Model{}, false
}

func (executor Executor) ref(ctx context.Context, branch string) (string, error) {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return "", err
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{Executable: git.Path, Arguments: []string{"show-ref", "--verify", "--hash", "refs/heads/" + branch}, Directory: executor.root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second})
	value := strings.TrimSpace(string(result.Stdout))
	if err != nil || result.ExitCode != 0 || len(value) != 40 {
		return "", errors.New("Headless target branch is unavailable")
	}
	return value, nil
}

func implementationPrompt(manifest domain.HeadlessManifest, wave domain.HeadlessWave) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Implement exactly Headless %s / %s in this disposable worktree. Stay within Tier 2, do not push, release, deploy, use secrets, access the network, or change the Git index. Level 7 will verify and commit.\n\nAllowed paths:\n", manifest.ID, wave.ID)
	for _, value := range manifest.AllowedPaths {
		fmt.Fprintf(&output, "- %s\n", value)
	}
	output.WriteString("\nAcceptance criteria:\n")
	for _, value := range wave.AcceptanceCriteria {
		fmt.Fprintf(&output, "- %s\n", value)
	}
	output.WriteString("\nFinish with the required structured result after implementing and self-checking the wave.")
	return output.String()
}

func reviewPrompt(manifest domain.HeadlessManifest, wave domain.HeadlessWave, base, candidate string) string {
	return fmt.Sprintf("Independently and read-only audit Headless %s / %s at exact candidate %s against base %s. Inspect acceptance, correctness, security, and verification evidence. Do not modify files or Git state. Return GO only when the declared wave is complete and safe; otherwise return NO_GO with concise findings.", manifest.ID, wave.ID, candidate, base)
}

func verificationCommands(commands [][]string) []domain.VerificationCommand {
	result := make([]domain.VerificationCommand, 0, len(commands))
	for index, argv := range commands {
		result = append(result, domain.VerificationCommand{Name: fmt.Sprintf("headless-%02d", index+1), Argv: append([]string{}, argv...)})
	}
	return result
}

func allPassed(checks []domain.CheckResult) bool {
	if len(checks) == 0 {
		return false
	}
	for _, check := range checks {
		if !check.Passed {
			return false
		}
	}
	return true
}

func routeForAttempt(task domain.TaskProfile, snapshots []domain.ProviderSnapshot, attempt int) domain.RouteDecision {
	decision := domain.Route(task, snapshots)
	if decision.ProviderID == "" || attempt <= 0 {
		return decision
	}
	qualified := make([]domain.RouteCandidate, 0, len(decision.Candidates))
	for _, candidate := range decision.Candidates {
		if candidate.Qualified {
			qualified = append(qualified, candidate)
		}
	}
	if len(qualified) == 0 {
		return decision
	}
	if attempt >= len(qualified) {
		attempt = len(qualified) - 1
	}
	selected := qualified[attempt]
	decision.ProviderID, decision.ModelID, decision.Effort = selected.ProviderID, selected.ModelID, selected.Effort
	decision.Fallbacks = decision.Fallbacks[:0]
	for index := attempt + 1; index < len(qualified); index++ {
		candidate := qualified[index]
		decision.Fallbacks = append(decision.Fallbacks, candidate.ProviderID+"/"+candidate.ModelID+"@"+string(candidate.Effort))
	}
	decision.Escalations = append(decision.Escalations, fmt.Sprintf("advanced to qualified route %d after prior failure", attempt+1))
	decision.Next = "start the escalated provider assignment"
	return decision
}

func sameRoute(left, right domain.RouteDecision) bool {
	return left.ProviderID == right.ProviderID && left.ModelID == right.ModelID && left.Effort == right.Effort
}

func allowed(relative string, patterns []string) bool {
	for _, pattern := range patterns {
		if relative == pattern {
			return true
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "**")
			if strings.HasPrefix(relative, prefix) && len(relative) > len(prefix) {
				return true
			}
		}
		if matched, _ := path.Match(pattern, relative); matched {
			return true
		}
	}
	return false
}

func protected(relative string) bool {
	value := filepath.ToSlash(relative)
	return value == ".git" || strings.HasPrefix(value, ".git/") || value == ".l7/config.json" || value == ".l7/orchestration.json" || value == "AGENTS.md" || value == "CLAUDE.md" || strings.HasPrefix(value, ".github/workflows/") || strings.Contains(strings.ToLower(filepath.Base(value)), "credential") || strings.HasPrefix(strings.ToLower(filepath.Base(value)), ".env")
}

func (executor Executor) failure(stage string, err error) headless.WaveOutcome {
	return executor.failureWithRoute(stage, domain.RouteDecision{}, "", "", err)
}

func (executor Executor) failureWithRoute(stage string, route domain.RouteDecision, session, worktree string, err error) headless.WaveOutcome {
	digest := sha256.Sum256([]byte(stage + "\x00" + err.Error()))
	return headless.WaveOutcome{State: "failed", ProviderID: route.ProviderID, ModelID: route.ModelID, SessionID: session, Worktree: worktree, FailureSignature: fmt.Sprintf("sha256:%x", digest), Message: stage + " failed"}
}

func (executor Executor) providerFailure(stage string, route domain.RouteDecision, result providerResult, worktree string, err error) headless.WaveOutcome {
	return executor.failureWithRoute(stage, route, result.SessionID, worktree, err)
}
