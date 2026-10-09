package crew

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

var testNow = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

func newTestStore(t *testing.T) (Store, domain.CrewPlan) {
	t.Helper()
	common, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store, err := Open(common)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SavePlan(plan); err != nil {
		t.Fatal(err)
	}
	return store, plan
}

func TestStorePlanApprovalAndActivation(t *testing.T) {
	store, plan := newTestStore(t)
	if err := store.SavePlan(plan); !errors.Is(err, os.ErrExist) {
		t.Fatalf("plan records must be create-only: %v", err)
	}
	loaded, err := store.LoadPlan(plan.ID)
	if err != nil || loaded.Digest != plan.Digest {
		t.Fatalf("loaded plan = %+v err=%v", loaded, err)
	}
	if _, err := store.Approve(plan, "sha256:"+strings.Repeat("0", 64), "Anup", "owner", testNow); err == nil {
		t.Fatal("approval of a different digest was accepted")
	}
	if _, err := store.Approve(plan, plan.Digest, "Anup\nX", "owner", testNow); err == nil {
		t.Fatal("multi-line approval identity was accepted")
	}
	approval, err := store.Approve(plan, plan.Digest, "Anup", "owner", testNow)
	if err != nil || approval.PlanDigest != plan.Digest {
		t.Fatalf("approval = %+v err=%v", approval, err)
	}
	if _, err := store.Approve(plan, plan.Digest, "Someone", "owner", testNow); err == nil {
		t.Fatal("approval was replaced")
	}
	if _, err := store.LoadApproval(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ActivePlanID(); !errors.Is(err, ErrNoActivePlan) {
		t.Fatalf("fresh store reported an active plan: %v", err)
	}
	if err := store.Activate(plan); err != nil {
		t.Fatal(err)
	}
	if active, err := store.ActivePlanID(); err != nil || active != plan.ID {
		t.Fatalf("active = %q err=%v", active, err)
	}
}

func TestActivateRefusesWhilePreviousCrewIsUnfinished(t *testing.T) {
	store, plan := newTestStore(t)
	if err := store.Activate(plan); err != nil {
		t.Fatal(err)
	}
	request := sampleRequest()
	request.BaseCommit = strings.Repeat("c", 40)
	next, err := fixedPlanner().Plan(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SavePlan(next); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(next); err == nil {
		t.Fatal("a second crew replaced one with unfinished tasks")
	}
	for _, task := range plan.Tasks {
		if _, err := store.Update(plan, task.ID, testNow, func(checkpoint *domain.CrewCheckpoint) error {
			checkpoint.State, checkpoint.Next = domain.CrewDone, "done"
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Activate(next); err != nil {
		t.Fatalf("finished crew blocked the next one: %v", err)
	}
}

func TestStoreUpdateAdvancesSequenceAndRecordsEvents(t *testing.T) {
	store, plan := newTestStore(t)
	taskID := plan.Tasks[0].ID
	checkpoints, err := store.Checkpoints(plan)
	if err != nil || checkpoints[taskID].State != domain.CrewQueued || checkpoints[taskID].Sequence != 0 {
		t.Fatalf("unstarted task = %+v err=%v", checkpoints[taskID], err)
	}
	for _, state := range []domain.CrewState{domain.CrewRunning, domain.CrewMerging} {
		if _, err := store.Update(plan, taskID, testNow, func(checkpoint *domain.CrewCheckpoint) error {
			checkpoint.State, checkpoint.Next = state, "continue"
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	checkpoints, err = store.Checkpoints(plan)
	if err != nil || checkpoints[taskID].Sequence != 2 || checkpoints[taskID].State != domain.CrewMerging {
		t.Fatalf("updated task = %+v err=%v", checkpoints[taskID], err)
	}
	for _, sequence := range []string{"00000001.json", "00000002.json"} {
		if _, err := os.Stat(store.taskPath(plan.ID, taskID, "events/"+sequence)); err != nil {
			t.Fatalf("event %s missing: %v", sequence, err)
		}
	}
	if _, err := store.Update(plan, plan.ID+"-t99", testNow, func(*domain.CrewCheckpoint) error { return nil }); err == nil {
		t.Fatal("unknown task accepted")
	}
	if _, err := store.Update(plan, taskID, testNow, func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State = "deploying"
		return nil
	}); err == nil {
		t.Fatal("invalid state accepted")
	}
	failure := errors.New("abort")
	if _, err := store.Update(plan, taskID, testNow, func(*domain.CrewCheckpoint) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("change error not returned: %v", err)
	}
}

func TestStoreDecisionsOpenOrderAndAnswer(t *testing.T) {
	store, plan := newTestStore(t)
	ship, scout := plan.Tasks[0].ID, plan.Tasks[1].ID
	low, err := store.OpenDecision(plan, scout, domain.CrewDecisionBlocked, "provider snapshot is missing", testNow)
	if err != nil {
		t.Fatal(err)
	}
	conflict, err := store.OpenDecision(plan, ship, domain.CrewDecisionConflict, "rebase conflict in auth/login.go", testNow)
	if err != nil || conflict.ID != ship+"-d01" {
		t.Fatalf("decision = %+v err=%v", conflict, err)
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil || len(decisions) != 2 || decisions[0].ID != conflict.ID || decisions[1].ID != low.ID {
		t.Fatalf("decisions must be ordered by impact: %+v err=%v", decisions, err)
	}
	if _, _, err := store.Answer(plan, conflict.ID, "merge-anyway", testNow); err == nil {
		t.Fatal("answer outside the offered options was accepted")
	}
	if _, _, err := store.Answer(plan, low.ID, domain.CrewAnswerRestart, testNow); err == nil {
		t.Fatal("restart was accepted for a decision that does not offer it")
	}
	if _, err := store.Update(plan, ship, testNow, func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.Worktree, checkpoint.CandidateCommit, checkpoint.SessionID = "/tmp/w", strings.Repeat("d", 40), "session"
		checkpoint.Next = "answer the decision"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	answered, checkpoint, err := store.Answer(plan, conflict.ID, domain.CrewAnswerRestart, testNow)
	if err != nil || answered.Answer != domain.CrewAnswerRestart {
		t.Fatalf("answer = %+v err=%v", answered, err)
	}
	if checkpoint.State != domain.CrewQueued || checkpoint.Attempt != 1 || checkpoint.Worktree != "" || checkpoint.CandidateCommit != "" || checkpoint.SessionID != "" {
		t.Fatalf("restart must queue a fresh attempt: %+v", checkpoint)
	}
	if _, _, err := store.Answer(plan, conflict.ID, domain.CrewAnswerCancel, testNow); err == nil {
		t.Fatal("a decision was answered twice")
	}
	_, checkpoint, err = store.Answer(plan, low.ID, domain.CrewAnswerCancel, testNow)
	if err != nil || checkpoint.State != domain.CrewCancelled {
		t.Fatalf("cancel = %+v err=%v", checkpoint, err)
	}
	decisions, err = store.Decisions(plan.ID)
	if err != nil || decisions[0].Answer == "" || decisions[1].Answer == "" {
		t.Fatalf("answered decisions = %+v err=%v", decisions, err)
	}
}

func TestTokenChangesWithProgress(t *testing.T) {
	store, plan := newTestStore(t)
	before, err := store.Checkpoints(plan)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(plan, plan.Tasks[1].ID, testNow, func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State, checkpoint.Next = domain.CrewRunning, "investigate"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	after, err := store.Checkpoints(plan)
	if err != nil {
		t.Fatal(err)
	}
	if Token(plan, before) != "0.0" || Token(plan, after) != "0.1" {
		t.Fatalf("token must list sequences in plan order: %s -> %s", Token(plan, before), Token(plan, after))
	}
	if _, err := parseToken(plan, "1"); err == nil {
		t.Fatal("token for a different task count was accepted")
	}
	if _, err := parseToken(plan, "1.-2"); err == nil {
		t.Fatal("negative sequence accepted")
	}
}

func TestSupervisorLockIsExclusive(t *testing.T) {
	store, _ := newTestStore(t)
	if store.SupervisorRunning() {
		t.Fatal("idle repository reported a running supervisor")
	}
	lock, err := store.AcquireSupervisor()
	if err != nil {
		t.Fatal(err)
	}
	if !store.SupervisorRunning() {
		t.Fatal("held supervisor lock was not detected")
	}
	if _, err := store.AcquireSupervisor(); err == nil {
		t.Fatal("a second supervisor acquired the lock")
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if store.SupervisorRunning() {
		t.Fatal("released supervisor lock still reported as running")
	}
}

func TestSaveReportIsBoundedToPlanTasks(t *testing.T) {
	store, plan := newTestStore(t)
	scout := plan.Tasks[1].ID
	if err := store.SaveReport(plan.ID, scout, "# Findings\n"); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveReport(plan.ID, scout, "# Updated findings\n"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.ReportPath(plan.ID, scout))
	if err != nil || string(data) != "# Updated findings\n" {
		t.Fatalf("report = %q err=%v", data, err)
	}
	for _, report := range []string{"", strings.Repeat("x", maxReportBytes+1), "nul\x00"} {
		if err := store.SaveReport(plan.ID, scout, report); err == nil {
			t.Fatalf("invalid report of %d bytes accepted", len(report))
		}
	}
	if err := store.SaveReport(plan.ID, "crew-ffffffffffff-t01", "x"); err == nil {
		t.Fatal("report for another plan's task accepted")
	}
}
