package crew

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// ErrSupervisorRunning reports that the repository already has a supervisor.
var ErrSupervisorRunning = errors.New("a crew supervisor is already running for this repository")

func (store Store) SupervisorLogPath() string { return filepath.Join(store.root, "supervisor.log") }

// StartSupervisor launches executable as a detached session leader in
// directory, appending its output to the supervisor log. It returns once the
// child holds the supervisor lock or has already finished successfully.
func (store Store) StartSupervisor(executable string, arguments []string, directory string, wait time.Duration) (int, error) {
	if !filepath.IsAbs(executable) || !filepath.IsAbs(directory) {
		return 0, errors.New("supervisor executable and directory must be absolute")
	}
	if store.SupervisorRunning() {
		return 0, ErrSupervisorRunning
	}
	if err := localfile.EnsureDirectory(store.root, 0o700); err != nil {
		return 0, err
	}
	logPath := store.SupervisorLogPath()
	pid, exited, err := processadapter.StartDetached(executable, arguments, directory, logPath)
	if err != nil {
		return 0, fmt.Errorf("start crew supervisor: %w", err)
	}
	deadline := time.NewTimer(wait)
	defer deadline.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-exited:
			if err != nil {
				return 0, fmt.Errorf("crew supervisor exited early (%v); see %s", err, logPath)
			}
			return pid, nil
		case <-ticker.C:
			if store.SupervisorRunning() {
				return pid, nil
			}
		case <-deadline.C:
			_ = processadapter.Terminate(pid)
			return 0, fmt.Errorf("crew supervisor did not start within %s; see %s", wait, logPath)
		}
	}
}

// StopSupervisor asks the running supervisor to pause its tasks and waits for
// it to release the lock. It reports whether a supervisor was running.
func (store Store) StopSupervisor(wait time.Duration) (bool, error) {
	if !store.SupervisorRunning() {
		return false, nil
	}
	info, err := store.LoadSupervisor()
	if err != nil || info.Schema != domain.CrewSchema || info.PID <= 1 {
		return true, errors.New("a crew supervisor holds the lock but its process record is unreadable")
	}
	if err := processadapter.Terminate(info.PID); err != nil {
		return true, fmt.Errorf("signal crew supervisor %d: %w", info.PID, err)
	}
	deadline := time.Now().Add(wait)
	for store.SupervisorRunning() {
		if time.Now().After(deadline) {
			return true, fmt.Errorf("crew supervisor %d did not stop within %s", info.PID, wait)
		}
		time.Sleep(50 * time.Millisecond)
	}
	return true, nil
}

// Supervise holds the supervisor lock, records this process, and runs the
// active plan until no task can progress without the owner.
func Supervise(ctx context.Context, store Store, engine Engine, executor Executor, now time.Time) (domain.CrewPlan, error) {
	lock, err := store.AcquireSupervisor()
	if err != nil {
		return domain.CrewPlan{}, err
	}
	defer lock.Close()
	planID, err := store.ActivePlanID()
	if err != nil {
		return domain.CrewPlan{}, err
	}
	plan, err := store.LoadPlan(planID)
	if err != nil {
		return domain.CrewPlan{}, err
	}
	if err := store.SaveSupervisor(SupervisorInfo{Schema: domain.CrewSchema, PlanID: plan.ID, PID: os.Getpid(), StartedAtUTC: now.UTC().Format(time.RFC3339)}); err != nil {
		return plan, err
	}
	return plan, engine.Run(ctx, store, plan, executor)
}

// CancelUnfinished ends every task that has not finished and closes its open
// decisions. Worktrees and evidence are kept.
func (store Store) CancelUnfinished(plan domain.CrewPlan, now time.Time) (int, error) {
	decisions, err := store.Decisions(plan.ID)
	if err != nil {
		return 0, err
	}
	for _, decision := range decisions {
		if decision.Answer == "" && contains(decision.Options, domain.CrewAnswerCancel) {
			if _, _, err := store.Answer(plan, decision.ID, domain.CrewAnswerCancel, now); err != nil {
				return 0, err
			}
		}
	}
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return 0, err
	}
	cancelled := 0
	for _, task := range plan.Tasks {
		if checkpoints[task.ID].State.Terminal() {
			continue
		}
		if _, err := store.Update(plan, task.ID, now, func(checkpoint *domain.CrewCheckpoint) error {
			checkpoint.State, checkpoint.Message, checkpoint.Next = domain.CrewCancelled, "cancelled by the owner", "inspect the retained worktree; the task will not run again"
			return nil
		}); err != nil {
			return cancelled, err
		}
		cancelled++
	}
	return cancelled, nil
}
