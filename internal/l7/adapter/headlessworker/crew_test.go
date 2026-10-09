package headlessworker

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type providerCall struct {
	reviewer bool
	session  string
	prompt   string
}

func crewSnapshots() []domain.ProviderSnapshot {
	model := func(id string, cost int) domain.ModelCapability {
		return domain.ModelCapability{ID: id, Languages: []string{"*"}, ContextWindow: 200_000, SupportsTools: true, SupportsEditing: true, SupportsResume: true, Efforts: []domain.ReasoningEffort{domain.EffortHigh}, CostClass: cost, LatencyClass: cost, Verified: true}
	}
	return []domain.ProviderSnapshot{
		{ID: "codex-local", Kind: domain.ProviderKindCodexAppServer, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{model("implementer", 1)}},
		{ID: "claude-local", Kind: domain.ProviderKindClaudeCLI, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{model("reviewer", 2)}},
	}
}

func crewFixture(t *testing.T, objective string) (CrewExecutor, domain.CrewPlan, string, string) {
	t.Helper()
	root, common, base := workerRepository(t)
	workerGit(t, root, "config", "user.name", "Level Seven")
	workerGit(t, root, "config", "user.email", "l7@example.invalid")
	workerGit(t, root, "branch", "l7/crew", base)
	executor, err := NewCrew(root, common, orchestrationconfig.Default())
	if err != nil {
		t.Fatal(err)
	}
	executor.snapshots = func() ([]domain.ProviderSnapshot, bool, error) { return crewSnapshots(), true, nil }
	plan, err := crew.NewPlannerWith(func() time.Time { return time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC) }).Plan(crew.PlanRequest{
		ObjectivePath: "crew.md", Objective: []byte(objective), BaseCommit: base, TargetBranch: "l7/crew", MaxWorkers: 2, RepairRounds: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	return executor, plan, root, base
}

const shipObjective = "## Ship: Fix handler\nPaths: api/**\nVerify: [\"true\"]\nAcceptance: handler returns 200\n"

// advanceTarget commits one file to l7/crew through a temporary worktree.
func advanceTarget(t *testing.T, root, relative, content string) string {
	t.Helper()
	worktree := filepath.Join(t.TempDir(), "target")
	workerGit(t, root, "worktree", "add", "-q", worktree, "l7/crew")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(worktree, relative)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, relative), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	workerGit(t, worktree, "add", relative)
	workerGit(t, worktree, "commit", "-q", "-m", "chore: advance target")
	head := strings.TrimSpace(workerGit(t, worktree, "rev-parse", "HEAD"))
	workerGit(t, root, "worktree", "remove", "--force", worktree)
	return head
}

func writeWorktreeFile(t *testing.T, worktree, relative, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(worktree, relative)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, relative), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCrewShipRepairsInSameSessionThenMergesAfterRebase(t *testing.T) {
	executor, plan, root, _ := crewFixture(t, shipObjective)
	var calls []providerCall
	executor.provider = func(_ context.Context, worktree string, _ domain.RouteDecision, prompt, session string, reviewer bool, _ []string, _ [][]string) (providerResult, error) {
		calls = append(calls, providerCall{reviewer: reviewer, session: session, prompt: prompt})
		if reviewer {
			return providerResult{SessionID: "review-1", Summary: "meets the criteria", Decision: domain.DecisionGO}, nil
		}
		content := "package api // bug\n"
		if len(calls) > 1 {
			content = "package api // fixed\n"
		}
		writeWorktreeFile(t, worktree, "api/handler.go", content)
		return providerResult{SessionID: "session-1", Summary: "changed the handler"}, nil
	}
	executor.verify = func(_ context.Context, worktree string, _ []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		data, _ := os.ReadFile(filepath.Join(worktree, "api", "handler.go"))
		if strings.Contains(string(data), "bug") {
			return []domain.CheckResult{{Name: "crew-01", ExitCode: 1, Code: "L7-VERIFY-001"}}, "stdout:\nexpected 200, got 500\n", context.DeadlineExceeded
		}
		return []domain.CheckResult{{Name: "crew-01", Passed: true}}, "", nil
	}
	task := plan.Tasks[0]
	built, err := executor.Build(context.Background(), plan, task, domain.CrewCheckpoint{})
	if err != nil || built.Kind != crew.OutcomeBuilt || len(built.CandidateCommit) != 40 {
		t.Fatalf("build = %+v err=%v", built, err)
	}
	if len(calls) != 3 || calls[0].reviewer || calls[1].reviewer || !calls[2].reviewer {
		t.Fatalf("expected implement, repair, review; got %+v", calls)
	}
	if calls[1].session != "session-1" || !strings.Contains(calls[1].prompt, "expected 200, got 500") || !strings.Contains(calls[1].prompt, "Never weaken") {
		t.Fatalf("repair must resume the same session with the failure output: %+v", calls[1])
	}
	moved := advanceTarget(t, root, "docs/notes.md", "notes\n")
	checkpoint := domain.CrewCheckpoint{CandidateCommit: built.CandidateCommit}
	merged, err := executor.Integrate(context.Background(), plan, task, checkpoint)
	if err != nil || merged.Kind != crew.OutcomeMerged {
		t.Fatalf("integrate = %+v err=%v", merged, err)
	}
	target := strings.TrimSpace(workerGit(t, root, "rev-parse", "l7/crew"))
	if target != merged.CandidateCommit || strings.TrimSpace(workerGit(t, root, "merge-base", "--is-ancestor", moved, target)) != "" {
		t.Fatalf("target %s must be the rebased candidate %s on top of %s", target, merged.CandidateCommit, moved)
	}
	if strings.TrimSpace(workerGit(t, root, "show", "l7/crew:api/handler.go")) != "package api // fixed" {
		t.Fatal("merged target does not contain the repaired handler")
	}
	if status := workerGit(t, root, "status", "--porcelain"); strings.TrimSpace(status) != "" {
		t.Fatalf("the user's checkout was modified: %q", status)
	}
}

func TestCrewShipStopsOnOutOfScopeChange(t *testing.T) {
	executor, plan, _, _ := crewFixture(t, shipObjective)
	executor.provider = func(_ context.Context, worktree string, _ domain.RouteDecision, _, _ string, _ bool, _ []string, _ [][]string) (providerResult, error) {
		writeWorktreeFile(t, worktree, "README.md", "rewritten\n")
		return providerResult{SessionID: "s", Summary: "done"}, nil
	}
	executor.verify = func(context.Context, string, []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		t.Fatal("verification ran for an out-of-scope change")
		return nil, "", nil
	}
	outcome, err := executor.Build(context.Background(), plan, plan.Tasks[0], domain.CrewCheckpoint{})
	if err != nil || outcome.Kind != crew.OutcomeDecision || outcome.Decision != domain.CrewDecisionScopeExpanded || !strings.Contains(outcome.Message, "README.md") {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
}

func TestCrewIntegrateTurnsConflictIntoDecision(t *testing.T) {
	executor, plan, root, _ := crewFixture(t, shipObjective)
	executor.provider = func(_ context.Context, worktree string, _ domain.RouteDecision, _, _ string, reviewer bool, _ []string, _ [][]string) (providerResult, error) {
		if reviewer {
			return providerResult{SessionID: "r", Summary: "ok", Decision: domain.DecisionGO}, nil
		}
		writeWorktreeFile(t, worktree, "api/handler.go", "package api // crew\n")
		return providerResult{SessionID: "s", Summary: "done"}, nil
	}
	executor.verify = func(context.Context, string, []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		return []domain.CheckResult{{Name: "crew-01", Passed: true}}, "", nil
	}
	built, err := executor.Build(context.Background(), plan, plan.Tasks[0], domain.CrewCheckpoint{})
	if err != nil || built.Kind != crew.OutcomeBuilt {
		t.Fatalf("build = %+v err=%v", built, err)
	}
	advanceTarget(t, root, "api/handler.go", "package api // other\n")
	outcome, err := executor.Integrate(context.Background(), plan, plan.Tasks[0], domain.CrewCheckpoint{})
	if err != nil || outcome.Kind != crew.OutcomeDecision || outcome.Decision != domain.CrewDecisionConflict || !strings.Contains(outcome.Message, "api/handler.go") {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	head := strings.TrimSpace(workerGit(t, built.Worktree, "rev-parse", "HEAD"))
	if head != built.CandidateCommit {
		t.Fatalf("aborted rebase must leave the candidate untouched: %s != %s", head, built.CandidateCommit)
	}
}

func TestCrewScoutIsReadOnly(t *testing.T) {
	executor, plan, _, _ := crewFixture(t, "## Scout: Why is CI slow?\nAcceptance: report names the slowest job\n")
	executor.provider = func(_ context.Context, _ string, route domain.RouteDecision, _, _ string, reviewer bool, scope []string, _ [][]string) (providerResult, error) {
		if !reviewer || len(scope) != 0 {
			t.Fatalf("scout must run read-only without write scope: reviewer=%t scope=%v", reviewer, scope)
		}
		return providerResult{SessionID: "scout", Summary: "The slowest job is lint (README.md:1).", Decision: domain.DecisionGO}, nil
	}
	outcome, err := executor.Build(context.Background(), plan, plan.Tasks[0], domain.CrewCheckpoint{})
	if err != nil || outcome.Kind != crew.OutcomeReported || !strings.Contains(outcome.Report, "# Why is CI slow?") || !strings.Contains(outcome.Report, "Result: answered") || !strings.Contains(outcome.Report, "README.md:1") {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	executor.provider = func(_ context.Context, worktree string, _ domain.RouteDecision, _, _ string, _ bool, _ []string, _ [][]string) (providerResult, error) {
		writeWorktreeFile(t, worktree, "scratch.txt", "x\n")
		return providerResult{SessionID: "scout", Summary: "done", Decision: domain.DecisionGO}, nil
	}
	outcome, err = executor.Build(context.Background(), plan, plan.Tasks[0], domain.CrewCheckpoint{Attempt: 1})
	if err != nil || outcome.Kind != crew.OutcomeDecision || outcome.Decision != domain.CrewDecisionScopeExpanded {
		t.Fatalf("a writing scout was accepted: %+v err=%v", outcome, err)
	}
}

func TestCrewReviewExcludesEveryModelThatImplementedTheAttempt(t *testing.T) {
	executor, plan, _, _ := crewFixture(t, shipObjective)
	task := plan.Tasks[0]
	var implementers, reviewers []string
	executor.provider = func(_ context.Context, worktree string, route domain.RouteDecision, _, _ string, reviewer bool, _ []string, _ [][]string) (providerResult, error) {
		if reviewer {
			reviewers = append(reviewers, route.ModelID)
			return providerResult{SessionID: "review", Summary: "ok", Decision: domain.DecisionGO}, nil
		}
		implementers = append(implementers, route.ModelID)
		content := "package api // partial\n"
		if route.ModelID == "reviewer" {
			content = "package api // fixed\n"
		}
		writeWorktreeFile(t, worktree, "api/handler.go", content)
		return providerResult{SessionID: "session-" + route.ModelID, Summary: "worked"}, nil
	}
	executor.verify = func(_ context.Context, worktree string, _ []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		data, _ := os.ReadFile(filepath.Join(worktree, "api", "handler.go"))
		if !strings.Contains(string(data), "fixed") {
			return []domain.CheckResult{{Name: "crew-01", ExitCode: 1, Code: "L7-VERIFY-001"}}, "stdout:\nstill broken\n", context.DeadlineExceeded
		}
		return []domain.CheckResult{{Name: "crew-01", Passed: true}}, "", nil
	}
	first, err := executor.Build(context.Background(), plan, task, domain.CrewCheckpoint{})
	if err != nil || first.Kind != crew.OutcomeFailed {
		t.Fatalf("first implementer must exhaust its repair rounds: %+v err=%v", first, err)
	}
	second, err := executor.Build(context.Background(), plan, task, domain.CrewCheckpoint{})
	if err != nil || second.Kind != crew.OutcomeDecision || second.Decision != domain.CrewDecisionBlocked || !strings.Contains(second.Message, "independent of every model") {
		t.Fatalf("both models implemented the attempt, so neither may review it: %+v err=%v implementers=%v reviewers=%v", second, err, implementers, reviewers)
	}
	if len(reviewers) != 0 || !reflect.DeepEqual(implementers, []string{"implementer", "implementer", "implementer", "reviewer"}) {
		t.Fatalf("implementers=%v reviewers=%v", implementers, reviewers)
	}
	snapshots := withoutModels(crewSnapshots(), []string{"codex-local/implementer"})
	if len(snapshots[0].Models) != 0 || len(snapshots[1].Models) != 1 || len(crewSnapshots()[0].Models) != 1 {
		t.Fatalf("withoutModels must filter a copy: %+v", snapshots)
	}
}

func TestCrewScopeProtectedRejectsControlPaths(t *testing.T) {
	for _, pattern := range []string{".github/**", ".github/workflows/ci.yml", ".l7/**", ".l7/orchestration.json", "AGENTS.md", "CLAUDE.md", ".git/**", "config/.env", "secrets/credentials.json"} {
		if _, found := CrewScopeProtected([]string{"docs/**", pattern}); !found {
			t.Fatalf("protected scope %q accepted", pattern)
		}
	}
	if pattern, found := CrewScopeProtected([]string{"docs/**", "api/**", ".github/ISSUE_TEMPLATE/bug.md"}); found {
		t.Fatalf("ordinary scope rejected: %q", pattern)
	}
}

func TestCrewCommitSubjectIsConventional(t *testing.T) {
	for _, title := range []string{"Fix login", "Fix\tlogin\x00now", strings.Repeat("Long title ", 40), "Überprüfung der Anmeldung"} {
		subject := crewCommitSubject(domain.CrewTask{ID: "crew-0123456789ab-t01", Title: title})
		if !domain.ConventionalSubject(subject) || !strings.HasPrefix(subject, "feat(crew): ") {
			t.Fatalf("subject %q for title %q is not conventional", subject, title)
		}
	}
}

func TestCrewPromptsCarryTaskContextAndRules(t *testing.T) {
	task := domain.CrewTask{ID: "crew-0123456789ab-t01", Shape: domain.CrewShip, Title: "Fix login", Objective: "Fix the flaky login test.", AcceptanceCriteria: []string{"passes 20 runs"}, AllowedPaths: []string{"auth/**"}, Verification: [][]string{{"go", "test", "./auth/..."}}}
	repair := crewRepairPrompt(task, "Check `go test ./auth/...` exited with code 1")
	for _, want := range []string{"Fix the cause, not the symptom", "exited with code 1", "auth/**", "go test ./auth/...", "passes 20 runs", "Never weaken"} {
		if !strings.Contains(repair, want) {
			t.Fatalf("repair prompt lacks %q:\n%s", want, repair)
		}
	}
	review := crewReviewPrompt(task, strings.Repeat("a", 40), strings.Repeat("b", 40))
	if !strings.Contains(review, "git diff "+strings.Repeat("a", 40)+" "+strings.Repeat("b", 40)) || !strings.Contains(review, "NO_GO") {
		t.Fatalf("review prompt is incomplete:\n%s", review)
	}
	if strings.Contains(crewScoutPrompt(domain.CrewTask{Title: "x", Objective: "y", AcceptanceCriteria: []string{"z"}}), "You may change") {
		t.Fatal("scout prompt offered write scope")
	}
}
