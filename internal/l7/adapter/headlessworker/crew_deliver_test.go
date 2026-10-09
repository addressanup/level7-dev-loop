package headlessworker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge/forgetest"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && forgetest.Commands[os.Args[1]] {
		os.Exit(forgetest.Main(os.Args[1:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

type deliveryFixture struct {
	executor CrewExecutor
	plan     domain.CrewPlan
	task     domain.CrewTask
	root     string
	remote   string
	base     string
	prompts  []string
}

// newDeliveryFixture builds a repository whose origin is a local bare remote
// and whose gh is the test binary.
func newDeliveryFixture(t *testing.T) *deliveryFixture {
	t.Helper()
	root, common, _ := workerRepository(t)
	workerGit(t, root, "config", "user.name", "Level Seven")
	workerGit(t, root, "config", "user.email", "l7@example.invalid")
	branch := strings.TrimSpace(workerGit(t, root, "symbolic-ref", "--short", "HEAD"))
	remote := filepath.Join(t.TempDir(), "remote.git")
	workerGit(t, root, "init", "-q", "--bare", remote)
	workerGit(t, root, "remote", "add", "origin", remote)
	workerGit(t, root, "push", "-q", "origin", "HEAD:refs/heads/"+branch)
	fixture := &deliveryFixture{root: root, remote: remote}
	fixture.base = advanceRemote(t, fixture, branch, "docs/remote.md", "pushed by a teammate\n")
	configuration := orchestrationconfig.Default()
	configuration.Features.Crew, configuration.Features.CrewPR = true, true
	executor, err := NewCrew(root, common, configuration)
	if err != nil {
		t.Fatal(err)
	}
	executor.snapshots = func() ([]domain.ProviderSnapshot, bool, error) { return crewSnapshots(), true, nil }
	gh, err := os.Executable()
	if err == nil {
		gh, err = filepath.EvalSymlinks(gh)
	}
	if err != nil {
		t.Fatal(err)
	}
	executor.forge = func() (forge.Client, error) { return forge.New(gh, root) }
	executor.verify = func(_ context.Context, worktree string, _ []domain.VerificationCommand) ([]domain.CheckResult, string, error) {
		if data, _ := os.ReadFile(filepath.Join(worktree, "api", "handler.go")); !strings.Contains(string(data), "fixed") {
			return []domain.CheckResult{{Name: "crew-01", ExitCode: 1, Code: "L7-VERIFY-001"}}, "stdout:\nbroken\n", context.DeadlineExceeded
		}
		return []domain.CheckResult{{Name: "crew-01", Passed: true}}, "", nil
	}
	executor.provider = func(_ context.Context, worktree string, _ domain.RouteDecision, prompt, _ string, reviewer bool, _ []string, _ [][]string) (providerResult, error) {
		if reviewer {
			return providerResult{SessionID: "review", Summary: "meets the criteria", Decision: domain.DecisionGO}, nil
		}
		fixture.prompts = append(fixture.prompts, prompt)
		writeWorktreeFile(t, worktree, "api/handler.go", "package api // fixed "+strings.Repeat("!", len(fixture.prompts))+"\n")
		return providerResult{SessionID: "session-1", Summary: "fixed"}, nil
	}
	plan, err := crew.NewPlannerWith(func() time.Time { return time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC) }).Plan(crew.PlanRequest{
		ObjectivePath: "crew.md", Objective: []byte(shipObjective), BaseCommit: strings.TrimSpace(workerGit(t, root, "rev-parse", "HEAD")),
		TargetBranch: branch, MaxWorkers: 2, RepairRounds: 2, Delivery: domain.CrewDeliveryPR, Remote: "origin",
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture.executor, fixture.plan, fixture.task = executor, plan, plan.Tasks[0]
	return fixture
}

// advanceRemote commits to branch on the remote through a scratch clone, as
// someone else pushing would.
func advanceRemote(t *testing.T, fixture *deliveryFixture, branch, relative, content string) string {
	t.Helper()
	clone := filepath.Join(t.TempDir(), "clone")
	workerGit(t, fixture.root, "clone", "-q", "--branch", branch, fixture.remote, clone)
	writeWorktreeFile(t, clone, relative, content)
	workerGit(t, clone, "add", relative)
	workerGit(t, clone, "-c", "user.name=Teammate", "-c", "user.email=t@example.invalid", "commit", "-q", "-m", "docs: teammate change")
	workerGit(t, clone, "push", "-q", "origin", "HEAD:refs/heads/"+branch)
	return strings.TrimSpace(workerGit(t, clone, "rev-parse", "HEAD"))
}

func (fixture *deliveryFixture) state(t *testing.T) forgetest.State {
	t.Helper()
	state, err := forgetest.Load(fixture.root)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func (fixture *deliveryFixture) update(t *testing.T, change func(*forgetest.PullRequest)) {
	t.Helper()
	state := fixture.state(t)
	change(&state.PullRequests[0])
	if err := forgetest.Save(fixture.root, state); err != nil {
		t.Fatal(err)
	}
}

func (fixture *deliveryFixture) deliver(t *testing.T) crew.Outcome {
	t.Helper()
	built, err := fixture.executor.Build(context.Background(), fixture.plan, fixture.task, domain.CrewCheckpoint{})
	if err != nil || built.Kind != crew.OutcomeBuilt {
		t.Fatalf("build = %+v err=%v", built, err)
	}
	delivered, err := fixture.executor.Integrate(context.Background(), fixture.plan, fixture.task, domain.CrewCheckpoint{})
	if err != nil || delivered.Kind != crew.OutcomeDelivered {
		t.Fatalf("delivery = %+v err=%v", delivered, err)
	}
	return delivered
}

func TestCrewPullRequestDeliveryPushesTheTaskBranchAndOpensOnePullRequest(t *testing.T) {
	fixture := newDeliveryFixture(t)
	delivered := fixture.deliver(t)
	branch := "l7/tasks/" + fixture.task.ID + "-a0"
	if remote := strings.Fields(workerGit(t, fixture.root, "ls-remote", "origin", "refs/heads/"+branch)); len(remote) == 0 || remote[0] != delivered.CandidateCommit {
		t.Fatalf("remote task branch = %v, want %s", remote, delivered.CandidateCommit)
	}
	if parent := strings.TrimSpace(workerGit(t, fixture.root, "rev-parse", delivered.CandidateCommit+"^")); parent != fixture.base {
		t.Fatalf("the task did not start from the fetched remote base: parent %s, remote base %s", parent, fixture.base)
	}
	state := fixture.state(t)
	if len(state.PullRequests) != 1 || delivered.PullRequest != 1 || !forge.PullURLValid(delivered.PullRequestURL) {
		t.Fatalf("pull requests = %+v delivered=%+v", state.PullRequests, delivered)
	}
	pull := state.PullRequests[0]
	for _, want := range []string{"Level 7 crew task `" + fixture.task.ID + "`", "handler returns 200", "`true`", "GO from `claude-local/reviewer`", "implemented by `codex-local/implementer`", "`api/handler.go`", "Not verified here"} {
		if !strings.Contains(pull.Body, want) {
			t.Fatalf("handoff lacks %q:\n%s", want, pull.Body)
		}
	}
	if pull.BaseRefName != fixture.plan.TargetBranch || pull.HeadRefName != branch || len(pull.Labels) != 1 || pull.Labels[0] != "l7-risk-tier-2" || !strings.HasPrefix(pull.Title, "feat(crew): ") {
		t.Fatalf("pull request = %+v", pull)
	}
	again, err := fixture.executor.Integrate(context.Background(), fixture.plan, fixture.task, domain.CrewCheckpoint{})
	if err != nil || again.Kind != crew.OutcomeDelivered || again.PullRequest != 1 || len(fixture.state(t).PullRequests) != 1 {
		t.Fatalf("a repeated delivery must reuse the pull request: %+v err=%v", again, err)
	}
	resumed, err := fixture.executor.Build(context.Background(), fixture.plan, fixture.task, domain.CrewCheckpoint{State: domain.CrewRunning, Checks: "pending"})
	if err != nil || resumed.Kind != crew.OutcomeDelivered || len(fixture.prompts) != 1 {
		t.Fatalf("a delivered task must resume as delivered without the implementer: %+v prompts=%d err=%v", resumed, len(fixture.prompts), err)
	}
	for _, call := range fixture.state(t).Calls {
		for _, argument := range call {
			if argument == "--force" || argument == "--admin" {
				t.Fatalf("delivery used %s: %v", argument, call)
			}
		}
	}
}

func TestCrewPullRequestMergeChecksEveryPrecondition(t *testing.T) {
	fixture := newDeliveryFixture(t)
	delivered := fixture.deliver(t)
	head := delivered.CandidateCommit
	open := domain.CrewCheckpoint{State: domain.CrewPROpen, PullRequest: 1, CandidateCommit: head}
	status, err := fixture.executor.Track(context.Background(), fixture.plan, fixture.task, open)
	if err != nil || status.State != "OPEN" || status.Head != head || status.Checks != "none" || status.Failed {
		t.Fatalf("status = %+v err=%v", status, err)
	}
	passing := []map[string]any{{"__typename": "CheckRun", "name": "test", "status": "COMPLETED", "conclusion": "SUCCESS"}}
	refusals := map[string]func(){
		"no checks ran": func() {},
		"pending checks": func() {
			fixture.update(t, func(pull *forgetest.PullRequest) {
				pull.StatusCheckRollup = []map[string]any{{"name": "test", "status": "IN_PROGRESS"}}
			})
		},
		"not clean": func() { fixture.update(t, func(pull *forgetest.PullRequest) { pull.StatusCheckRollup = passing }) },
		"draft": func() {
			fixture.update(t, func(pull *forgetest.PullRequest) { pull.MergeStateStatus, pull.IsDraft = "CLEAN", true })
		},
		"moved head": func() {
			fixture.update(t, func(pull *forgetest.PullRequest) { pull.IsDraft, pull.HeadRefOid = false, strings.Repeat("9", 40) })
		},
		"closed": func() {
			fixture.update(t, func(pull *forgetest.PullRequest) { pull.HeadRefOid, pull.State = head, "CLOSED" })
		},
	}
	for _, name := range []string{"no checks ran", "pending checks", "not clean", "draft", "moved head", "closed"} {
		refusals[name]()
		if _, err := fixture.executor.MergePullRequest(context.Background(), fixture.plan, open, head); err == nil {
			t.Fatalf("merge accepted with %s", name)
		}
	}
	fixture.update(t, func(pull *forgetest.PullRequest) { pull.State = "OPEN" })
	for name, checkpoint := range map[string]domain.CrewCheckpoint{
		"stale head":      {State: domain.CrewPROpen, PullRequest: 1, CandidateCommit: strings.Repeat("8", 40)},
		"not pr-open":     {State: domain.CrewDone, PullRequest: 1, CandidateCommit: head},
		"no pull request": {State: domain.CrewPROpen, CandidateCommit: head},
	} {
		if _, err := fixture.executor.MergePullRequest(context.Background(), fixture.plan, checkpoint, head); err == nil {
			t.Fatalf("merge accepted with %s", name)
		}
	}
	if status, _ := fixture.executor.Track(context.Background(), fixture.plan, fixture.task, open); status.Checks != "passed: 1" {
		t.Fatalf("passing checks = %+v", status)
	}
	merged, err := fixture.executor.MergePullRequest(context.Background(), fixture.plan, open, head)
	if err != nil || merged.Kind != crew.OutcomeMerged {
		t.Fatalf("merge = %+v err=%v", merged, err)
	}
	state := fixture.state(t)
	if len(state.Merges) != 1 || state.Merges[0] != (forgetest.Merge{Number: 1, Method: "squash", Head: head}) {
		t.Fatalf("merges = %+v", state.Merges)
	}
	if status, _ := fixture.executor.Track(context.Background(), fixture.plan, fixture.task, open); status.State != "MERGED" {
		t.Fatalf("merged status = %+v", status)
	}
}

func TestCrewPullRequestFailedChecksRepairAndPushANewCommit(t *testing.T) {
	fixture := newDeliveryFixture(t)
	first := fixture.deliver(t)
	fixture.update(t, func(pull *forgetest.PullRequest) {
		pull.StatusCheckRollup = []map[string]any{{"__typename": "CheckRun", "name": "test", "status": "COMPLETED", "conclusion": "FAILURE", "detailsUrl": "https://github.com/owner/repo/actions/runs/5/job/77"}}
	})
	state := fixture.state(t)
	state.Logs["77"] = "--- FAIL: TestHandler (0.01s)\n    handler_test.go:12: got 500, want 200\n"
	if err := forgetest.Save(fixture.root, state); err != nil {
		t.Fatal(err)
	}
	failedChecks := domain.CrewCheckpoint{State: domain.CrewRunning, PullRequest: 1, CandidateCommit: first.CandidateCommit, Checks: "failed: test"}
	rebuilt, err := fixture.executor.Build(context.Background(), fixture.plan, fixture.task, failedChecks)
	if err != nil || rebuilt.Kind != crew.OutcomeBuilt || rebuilt.CandidateCommit == first.CandidateCommit {
		t.Fatalf("failed checks must produce a new reviewed candidate: %+v err=%v", rebuilt, err)
	}
	repair := fixture.prompts[len(fixture.prompts)-1]
	for _, want := range []string{"Checks failed on pull request #1", "got 500, want 200"} {
		if !strings.Contains(repair, want) {
			t.Fatalf("repair prompt lacks %q:\n%s", want, repair)
		}
	}
	updated, err := fixture.executor.Integrate(context.Background(), fixture.plan, fixture.task, failedChecks)
	if err != nil || updated.Kind != crew.OutcomeDelivered || updated.PullRequest != 1 || len(fixture.state(t).PullRequests) != 1 {
		t.Fatalf("the repair must update the same pull request: %+v err=%v", updated, err)
	}
	branch := "l7/tasks/" + fixture.task.ID + "-a0"
	if remote := strings.Fields(workerGit(t, fixture.root, "ls-remote", "origin", "refs/heads/"+branch)); remote[0] != rebuilt.CandidateCommit {
		t.Fatalf("remote head = %s, want %s", remote[0], rebuilt.CandidateCommit)
	}
	if parent := strings.TrimSpace(workerGit(t, fixture.root, "rev-parse", rebuilt.CandidateCommit+"^")); parent != first.CandidateCommit {
		t.Fatalf("the repair must build on the delivered commit with a fast-forward push: parent %s", parent)
	}
}

func TestCrewPullRequestRefusesForeignPushesAndClosedRequests(t *testing.T) {
	fixture := newDeliveryFixture(t)
	first := fixture.deliver(t)
	branch := "l7/tasks/" + fixture.task.ID + "-a0"
	advanceRemote(t, fixture, branch, "api/teammate.go", "package api\n")
	failedChecks := domain.CrewCheckpoint{State: domain.CrewRunning, PullRequest: 1, CandidateCommit: first.CandidateCommit, Checks: "failed: test"}
	if _, err := fixture.executor.Build(context.Background(), fixture.plan, fixture.task, failedChecks); err != nil {
		t.Fatal(err)
	}
	rejected, err := fixture.executor.Integrate(context.Background(), fixture.plan, fixture.task, failedChecks)
	if err != nil || rejected.Kind != crew.OutcomeDecision || !strings.Contains(rejected.Message, "was rejected") {
		t.Fatalf("a push over someone else's commit must stop for the owner: %+v err=%v", rejected, err)
	}
	fixture.update(t, func(pull *forgetest.PullRequest) { pull.State = "CLOSED" })
	closed, err := fixture.executor.Integrate(context.Background(), fixture.plan, fixture.task, failedChecks)
	if err != nil || closed.Kind != crew.OutcomeDecision || !strings.Contains(closed.Message, "does not reopen") || len(fixture.state(t).PullRequests) != 1 {
		t.Fatalf("a closed pull request must never be replaced: %+v err=%v", closed, err)
	}
}
