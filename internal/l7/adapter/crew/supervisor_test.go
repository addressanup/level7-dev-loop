package crew

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// TestHelperSupervisorProcess is the child process for the supervisor tests.
func TestHelperSupervisorProcess(t *testing.T) {
	common := os.Getenv("L7_CREW_HELPER_COMMON")
	if common == "" {
		return
	}
	store, err := Open(common)
	if err != nil {
		os.Exit(3)
	}
	lock, err := store.AcquireSupervisor()
	if err != nil {
		os.Exit(4)
	}
	defer lock.Close()
	if err := store.SaveSupervisor(SupervisorInfo{Schema: domain.CrewSchema, PID: os.Getpid(), StartedAtUTC: time.Now().UTC().Format(time.RFC3339)}); err != nil {
		os.Exit(5)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
	case <-time.After(30 * time.Second):
	}
	os.Exit(0)
}

func TestSupervisorStartsDetachedAndStopsOnRequest(t *testing.T) {
	store, _ := newTestStore(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("L7_CREW_HELPER_COMMON", filepath.Dir(filepath.Dir(store.root)))
	pid, err := store.StartSupervisor(executable, []string{"-test.run=^TestHelperSupervisorProcess$"}, directory, 10*time.Second)
	if err != nil || pid <= 1 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	t.Cleanup(func() { _, _ = store.StopSupervisor(5 * time.Second) })
	if !store.SupervisorRunning() {
		t.Fatal("started supervisor does not hold the lock")
	}
	if _, err := store.StartSupervisor(executable, []string{"-test.run=^TestHelperSupervisorProcess$"}, directory, time.Second); !errors.Is(err, ErrSupervisorRunning) {
		t.Fatalf("second supervisor start = %v", err)
	}
	running, err := store.StopSupervisor(10 * time.Second)
	if err != nil || !running || store.SupervisorRunning() {
		t.Fatalf("stop running=%t err=%v still=%t", running, err, store.SupervisorRunning())
	}
	if running, err := store.StopSupervisor(time.Second); err != nil || running {
		t.Fatalf("stopping an idle repository reported running=%t err=%v", running, err)
	}
}

func TestSupervisorReportsEarlyExit(t *testing.T) {
	store, _ := newTestStore(t)
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.StartSupervisor("/usr/bin/false", nil, directory, 5*time.Second); err == nil {
		t.Fatal("a supervisor that exited with failure was reported as started")
	}
	if _, err := store.StartSupervisor("relative/l7", nil, directory, time.Second); err == nil {
		t.Fatal("relative executable accepted")
	}
}

func TestCancelUnfinishedClosesDecisionsAndKeepsFinishedTasks(t *testing.T) {
	store, plan := newTestStore(t)
	ship, scout := plan.Tasks[0].ID, plan.Tasks[1].ID
	if _, err := store.Update(plan, scout, testNow, func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State, checkpoint.Next = domain.CrewDone, "done"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.OpenDecision(plan, ship, domain.CrewDecisionBlocked, "missing target", testNow); err != nil {
		t.Fatal(err)
	}
	cancelled, err := store.CancelUnfinished(plan, testNow)
	if err != nil || cancelled != 0 {
		t.Fatalf("cancelled=%d err=%v; the decided task is cancelled through its decision", cancelled, err)
	}
	checkpoints := states(t, store, plan)
	if checkpoints[ship].State != domain.CrewCancelled || checkpoints[scout].State != domain.CrewDone {
		t.Fatalf("states after cancel: ship=%s scout=%s", checkpoints[ship].State, checkpoints[scout].State)
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil || len(decisions) != 1 || decisions[0].Answer != domain.CrewAnswerCancel {
		t.Fatalf("open decision was not closed: %+v err=%v", decisions, err)
	}
}
