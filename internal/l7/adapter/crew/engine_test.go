package crew

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type fakeExecutor struct {
	build     func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error)
	integrate func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error)

	mu                  sync.Mutex
	builds, integrates  map[string]int
	activeBuilds        atomic.Int32
	activeIntegrates    atomic.Int32
	maxBuilds, maxMerge atomic.Int32
}

func (executor *fakeExecutor) Build(ctx context.Context, _ domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (Outcome, error) {
	executor.count(&executor.builds, task.ID)
	raise(&executor.maxBuilds, executor.activeBuilds.Add(1))
	defer executor.activeBuilds.Add(-1)
	if executor.build != nil {
		return executor.build(ctx, task, checkpoint)
	}
	if task.Shape == domain.CrewScout {
		return Outcome{Kind: OutcomeReported, Report: "# " + task.Title + "\n", ProviderID: "claude-local", ModelID: "m"}, nil
	}
	return Outcome{Kind: OutcomeBuilt, Worktree: "/w/" + task.ID, CandidateCommit: strings.Repeat("e", 40), Verification: "passed"}, nil
}

func (executor *fakeExecutor) Integrate(ctx context.Context, _ domain.CrewPlan, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (Outcome, error) {
	executor.count(&executor.integrates, task.ID)
	raise(&executor.maxMerge, executor.activeIntegrates.Add(1))
	defer executor.activeIntegrates.Add(-1)
	time.Sleep(5 * time.Millisecond)
	if executor.integrate != nil {
		return executor.integrate(ctx, task, checkpoint)
	}
	return Outcome{Kind: OutcomeMerged, CandidateCommit: checkpoint.CandidateCommit}, nil
}

func (executor *fakeExecutor) count(target *map[string]int, id string) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if *target == nil {
		*target = map[string]int{}
	}
	(*target)[id]++
}

func (executor *fakeExecutor) calls(target map[string]int, id string) int {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	return target[id]
}

func raise(maximum *atomic.Int32, value int32) {
	for {
		current := maximum.Load()
		if value <= current || maximum.CompareAndSwap(current, value) {
			return
		}
	}
}

func testEngine(wait func(context.Context, time.Time) error) Engine {
	engine := NewEngineWith(func() time.Time { return testNow }, wait)
	engine.poll = 5 * time.Millisecond
	return engine
}

func approvedCrew(t *testing.T, objective string, workers int) (Store, domain.CrewPlan) {
	t.Helper()
	store, _ := newTestStore(t)
	request := sampleRequest()
	request.Objective, request.MaxWorkers = []byte(objective), workers
	plan, err := fixedPlanner().Plan(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SavePlan(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Approve(plan, plan.Digest, "Anup", "owner", testNow); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(plan); err != nil {
		t.Fatal(err)
	}
	return store, plan
}

const parallelObjective = `## Ship: API
Paths: api/**
Verify: ["go","test","./api/..."]
Acceptance: api works
## Ship: Web
Paths: web/**
Verify: ["go","test","./web/..."]
Acceptance: web works
## Scout: CI
Acceptance: report explains CI time
`

func states(t *testing.T, store Store, plan domain.CrewPlan) map[string]domain.CrewCheckpoint {
	t.Helper()
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		t.Fatal(err)
	}
	return checkpoints
}

func TestEngineRunsTasksInParallelAndMergesOneAtATime(t *testing.T) {
	store, plan := approvedCrew(t, parallelObjective, 3)
	var arrived atomic.Int32
	executor := &fakeExecutor{}
	executor.build = func(_ context.Context, task domain.CrewTask, _ domain.CrewCheckpoint) (Outcome, error) {
		arrived.Add(1)
		deadline := time.Now().Add(2 * time.Second)
		for arrived.Load() < 3 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if task.Shape == domain.CrewScout {
			return Outcome{Kind: OutcomeReported, Report: "# CI\n"}, nil
		}
		return Outcome{Kind: OutcomeBuilt, Worktree: "/w/" + task.ID, CandidateCommit: strings.Repeat("e", 40), Verification: "passed"}, nil
	}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	if executor.maxBuilds.Load() != 3 {
		t.Fatalf("expected three concurrent builds, saw %d", executor.maxBuilds.Load())
	}
	if executor.maxMerge.Load() != 1 {
		t.Fatalf("merges must be serialized, saw %d concurrent", executor.maxMerge.Load())
	}
	for _, task := range plan.Tasks {
		checkpoint := states(t, store, plan)[task.ID]
		if checkpoint.State != domain.CrewDone {
			t.Fatalf("%s ended in %s: %s", task.ID, checkpoint.State, checkpoint.Message)
		}
	}
	if executor.calls(executor.integrates, plan.Tasks[2].ID) != 0 {
		t.Fatal("scout task reached the merge queue")
	}
	if checkpoint := states(t, store, plan)[plan.Tasks[0].ID]; checkpoint.Worktree != "/w/"+plan.Tasks[0].ID || checkpoint.Verification != "passed" {
		t.Fatalf("ship checkpoint did not record its build: %+v", checkpoint)
	}
}

func TestEngineNeverRunsOverlappingShipTasksTogether(t *testing.T) {
	objective := strings.Replace(parallelObjective, "Paths: web/**", "Paths: api/handler.go", 1)
	store, plan := approvedCrew(t, objective, 3)
	executor := &fakeExecutor{}
	executor.build = func(_ context.Context, task domain.CrewTask, _ domain.CrewCheckpoint) (Outcome, error) {
		time.Sleep(20 * time.Millisecond)
		if task.Shape == domain.CrewScout {
			return Outcome{Kind: OutcomeReported, Report: "# CI\n"}, nil
		}
		return Outcome{Kind: OutcomeBuilt, CandidateCommit: strings.Repeat("e", 40)}, nil
	}
	var overlap atomic.Int32
	var shipActive atomic.Int32
	inner := executor.build
	executor.build = func(ctx context.Context, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (Outcome, error) {
		if task.Shape == domain.CrewShip {
			raise(&overlap, shipActive.Add(1))
			defer shipActive.Add(-1)
		}
		return inner(ctx, task, checkpoint)
	}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	if overlap.Load() != 1 {
		t.Fatalf("overlapping ship tasks ran together (%d)", overlap.Load())
	}
	for _, checkpoint := range states(t, store, plan) {
		if checkpoint.State != domain.CrewDone {
			t.Fatalf("%s ended in %s", checkpoint.TaskID, checkpoint.State)
		}
	}
}

func TestEngineOpensNoProgressDecisionAfterRepeatedFailures(t *testing.T) {
	store, plan := approvedCrew(t, "## Ship: API\nPaths: api/**\nVerify: [\"go\",\"test\"]\nAcceptance: ok\n", 1)
	executor := &fakeExecutor{build: func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error) {
		return Outcome{Kind: OutcomeFailed, FailureSignature: "same", Message: "tests failed"}, nil
	}}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	taskID := plan.Tasks[0].ID
	if calls := executor.calls(executor.builds, taskID); calls != 3 {
		t.Fatalf("expected the three-failure breaker, saw %d builds", calls)
	}
	checkpoint := states(t, store, plan)[taskID]
	decisions, err := store.Decisions(plan.ID)
	if err != nil || checkpoint.State != domain.CrewNeedsDecision || len(decisions) != 1 || decisions[0].Kind != domain.CrewDecisionNoProgress {
		t.Fatalf("checkpoint=%+v decisions=%+v err=%v", checkpoint, decisions, err)
	}
}

func TestEngineWaitsForQuotaThenContinues(t *testing.T) {
	store, plan := approvedCrew(t, "## Scout: CI\nAcceptance: ok\n", 1)
	reset := testNow.Add(time.Hour)
	var waited time.Time
	executor := &fakeExecutor{}
	executor.build = func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error) {
		if executor.calls(executor.builds, plan.Tasks[0].ID) == 1 {
			return Outcome{Kind: OutcomeQuota, QuotaResetAtUTC: reset.Format(time.RFC3339), Message: "rate limited"}, errors.New("quota")
		}
		return Outcome{Kind: OutcomeReported, Report: "# CI\n"}, nil
	}
	engine := testEngine(func(_ context.Context, until time.Time) error {
		waited = until
		return nil
	})
	if err := engine.Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	if !waited.Equal(reset) || states(t, store, plan)[plan.Tasks[0].ID].State != domain.CrewDone {
		t.Fatalf("waited until %s; final %+v", waited, states(t, store, plan)[plan.Tasks[0].ID])
	}
}

func TestEngineTurnsIntegrationConflictIntoDecision(t *testing.T) {
	store, plan := approvedCrew(t, "## Ship: API\nPaths: api/**\nVerify: [\"go\",\"test\"]\nAcceptance: ok\n", 1)
	executor := &fakeExecutor{integrate: func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error) {
		return Outcome{Kind: OutcomeDecision, Decision: domain.CrewDecisionConflict, Message: "rebase conflict in api/handler.go"}, nil
	}}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil || len(decisions) != 1 || decisions[0].Kind != domain.CrewDecisionConflict || !strings.Contains(decisions[0].Question, "api/handler.go") {
		t.Fatalf("decisions = %+v err=%v", decisions, err)
	}
	if _, _, err := store.Answer(plan, decisions[0].ID, domain.CrewAnswerRetry, testNow); err != nil {
		t.Fatal(err)
	}
	executor.integrate = nil
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	if checkpoint := states(t, store, plan)[plan.Tasks[0].ID]; checkpoint.State != domain.CrewDone {
		t.Fatalf("retried task ended in %s", checkpoint.State)
	}
}

func TestEngineRebuildsWhenIntegrationNeedsAnotherReview(t *testing.T) {
	store, plan := approvedCrew(t, "## Ship: API\nPaths: api/**\nVerify: [\"go\",\"test\"]\nAcceptance: ok\n", 1)
	executor := &fakeExecutor{}
	executor.integrate = func(_ context.Context, task domain.CrewTask, checkpoint domain.CrewCheckpoint) (Outcome, error) {
		if executor.calls(executor.integrates, task.ID) == 1 {
			return Outcome{Kind: OutcomeBuilt, Message: "rebase changed the patch"}, nil
		}
		return Outcome{Kind: OutcomeMerged, CandidateCommit: checkpoint.CandidateCommit}, nil
	}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	taskID := plan.Tasks[0].ID
	checkpoint := states(t, store, plan)[taskID]
	if checkpoint.State != domain.CrewDone || checkpoint.RepeatedFailures != 0 || executor.calls(executor.builds, taskID) != 2 {
		t.Fatalf("rebuild must re-run Build without counting a failure: %+v builds=%d", checkpoint, executor.calls(executor.builds, taskID))
	}
	executor.integrate = func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error) {
		return Outcome{Kind: OutcomeBuilt, Message: "target moved again"}, nil
	}
	store, plan = approvedCrew(t, "## Ship: Web\nPaths: web/**\nVerify: [\"go\",\"test\"]\nAcceptance: ok\n", 1)
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil || len(decisions) != 1 || decisions[0].Kind != domain.CrewDecisionNoProgress {
		t.Fatalf("endless rebuilds must stop for a decision: %+v err=%v", decisions, err)
	}
}

func TestEngineCancellationPausesAndRunResumes(t *testing.T) {
	store, plan := approvedCrew(t, "## Scout: CI\nAcceptance: ok\n", 1)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	executor := &fakeExecutor{build: func(buildCtx context.Context, _ domain.CrewTask, _ domain.CrewCheckpoint) (Outcome, error) {
		close(started)
		<-buildCtx.Done()
		return Outcome{}, buildCtx.Err()
	}}
	done := make(chan error)
	go func() { done <- testEngine(nil).Run(ctx, store, plan, executor) }()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled run returned %v", err)
	}
	if checkpoint := states(t, store, plan)[plan.Tasks[0].ID]; checkpoint.State != domain.CrewPaused {
		t.Fatalf("cancelled task ended in %s", checkpoint.State)
	}
	executor.build = nil
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	if checkpoint := states(t, store, plan)[plan.Tasks[0].ID]; checkpoint.State != domain.CrewDone {
		t.Fatalf("resumed task ended in %s", checkpoint.State)
	}
}

func TestEngineRejectsUnexpectedOutcomesAndMissingApproval(t *testing.T) {
	store, plan := approvedCrew(t, "## Scout: CI\nAcceptance: ok\n", 1)
	executor := &fakeExecutor{build: func(context.Context, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error) {
		return Outcome{Kind: OutcomeBuilt}, nil
	}}
	if err := testEngine(nil).Run(context.Background(), store, plan, executor); err != nil {
		t.Fatal(err)
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil || len(decisions) != 1 || decisions[0].Kind != domain.CrewDecisionBlocked {
		t.Fatalf("a scout that built a candidate must stop for a decision: %+v err=%v", decisions, err)
	}
	unapproved, _ := newTestStore(t)
	other, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := testEngine(nil).Run(context.Background(), unapproved, other, executor); err == nil {
		t.Fatal("a crew without owner approval was run")
	}
}

func TestWaitReturnsOnAttentionIdleOrTimeout(t *testing.T) {
	store, plan := approvedCrew(t, parallelObjective, 3)
	result, err := Wait(context.Background(), store, plan, "", time.Second, time.Millisecond)
	if err != nil || result.Reason != "idle" {
		t.Fatalf("idle crew wait = %+v err=%v", result, err)
	}
	lock, err := store.AcquireSupervisor()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	update := func(index int, state domain.CrewState) {
		if _, err := store.Update(plan, plan.Tasks[index].ID, testNow, func(checkpoint *domain.CrewCheckpoint) error {
			checkpoint.State, checkpoint.Next = state, "continue"
			return nil
		}); err != nil {
			t.Error(err)
		}
	}
	token := result.Token
	update(0, domain.CrewRunning)
	result, err = Wait(context.Background(), store, plan, token, 20*time.Millisecond, time.Millisecond)
	if err != nil || result.Reason != "timeout" || len(result.Changed) != 1 {
		t.Fatalf("progress without attention must not wake the liaison: %+v err=%v", result, err)
	}
	token = result.Token
	updated := make(chan struct{})
	go func() {
		defer close(updated)
		time.Sleep(20 * time.Millisecond)
		update(1, domain.CrewDone)
	}()
	result, err = Wait(context.Background(), store, plan, token, 2*time.Second, time.Millisecond)
	<-updated
	if err != nil || result.Reason != "attention" || len(result.Changed) != 1 || result.Changed[0] != plan.Tasks[1].ID {
		t.Fatalf("finished task must wake the liaison: %+v err=%v", result, err)
	}
	if _, err := Wait(context.Background(), store, plan, "9", time.Second, time.Millisecond); err == nil {
		t.Fatal("foreign token accepted")
	}
	if _, err := Wait(context.Background(), store, plan, "", 2*time.Hour, time.Millisecond); err == nil {
		t.Fatal("unbounded wait accepted")
	}
}
