package crew

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	maxRecordBytes  = 4 << 20
	maxReportBytes  = 1 << 20
	stateLockWait   = 10 * time.Second
	stateLockPeriod = 20 * time.Millisecond
)

// ErrNoActivePlan reports that no crew plan has been started in this
// repository.
var ErrNoActivePlan = errors.New("no crew plan is active; run l7 crew plan and l7 crew start")

type Approval struct {
	Schema        int    `json:"schema"`
	PlanID        string `json:"plan_id"`
	PlanDigest    string `json:"plan_digest"`
	Actor         string `json:"actor"`
	Role          string `json:"role"`
	ApprovedAtUTC string `json:"approved_at_utc"`
	Source        string `json:"source"`
}

type SupervisorInfo struct {
	Schema       int    `json:"schema"`
	PlanID       string `json:"plan_id"`
	PID          int    `json:"pid"`
	StartedAtUTC string `json:"started_at_utc"`
}

// Store keeps crew records in private Git-bound state:
//
//	<git-common>/l7/crew/active.json
//	<git-common>/l7/crew/plans/<plan>/{plan,approval}.json
//	<git-common>/l7/crew/plans/<plan>/tasks/<task>/{checkpoint.json,report.md,events/}
//	<git-common>/l7/crew/plans/<plan>/decisions/<decision>.json
//	<git-common>/l7/crew/worktrees/<task>-a<attempt>/
type Store struct{ root string }

func Open(common string) (Store, error) {
	if !filepath.IsAbs(common) {
		return Store{}, errors.New("Git common directory must be absolute")
	}
	clean := filepath.Clean(common)
	info, err := os.Lstat(clean)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return Store{}, errors.New("Git common directory is unsafe")
	}
	return Store{root: filepath.Join(clean, "l7", "crew")}, nil
}

func (store Store) WorktreeRoot() string { return filepath.Join(store.root, "worktrees") }

// ObjectiveRoot is a private place for a liaison to draft crew objectives
// without touching the user's checkout.
func (store Store) ObjectiveRoot() string { return filepath.Join(store.root, "objectives") }

func (store Store) SavePlan(plan domain.CrewPlan) error {
	if err := ValidatePlan(plan); err != nil {
		return err
	}
	return store.write(store.planPath(plan.ID, "plan.json"), plan, false)
}

func (store Store) LoadPlan(planID string) (domain.CrewPlan, error) {
	var plan domain.CrewPlan
	if !domain.CrewPlanIDValid(planID) {
		return plan, errors.New("crew plan ID is invalid")
	}
	if err := store.read(store.planPath(planID, "plan.json"), &plan); err != nil {
		return domain.CrewPlan{}, err
	}
	if plan.ID != planID {
		return domain.CrewPlan{}, errors.New("crew plan identity does not match its path")
	}
	if err := ValidatePlan(plan); err != nil {
		return domain.CrewPlan{}, err
	}
	return plan, nil
}

// Approve binds the owner's confirmation to the exact plan digest. Approval is
// recorded once and never replaced.
func (store Store) Approve(plan domain.CrewPlan, digest, actor, role string, now time.Time) (Approval, error) {
	if err := ValidatePlan(plan); err != nil {
		return Approval{}, err
	}
	if digest != plan.Digest || !boundedLine(actor, 256) || !boundedLine(role, 256) {
		return Approval{}, errors.New("crew approval does not bind the exact plan digest, owner, and role")
	}
	approval := Approval{Schema: domain.CrewSchema, PlanID: plan.ID, PlanDigest: digest, Actor: actor, Role: role, ApprovedAtUTC: now.UTC().Format(time.RFC3339), Source: "active-owner-interaction"}
	if err := store.write(store.planPath(plan.ID, "approval.json"), approval, false); err != nil {
		return Approval{}, err
	}
	return approval, nil
}

func (store Store) LoadApproval(plan domain.CrewPlan) (Approval, error) {
	var approval Approval
	if err := store.read(store.planPath(plan.ID, "approval.json"), &approval); err != nil {
		return approval, err
	}
	if approval.Schema != domain.CrewSchema || approval.PlanID != plan.ID || approval.PlanDigest != plan.Digest ||
		!boundedLine(approval.Actor, 256) || !boundedLine(approval.Role, 256) || approval.Source != "active-owner-interaction" {
		return Approval{}, errors.New("crew approval is invalid or does not match the plan")
	}
	return approval, nil
}

// Activate makes plan the repository's active crew. A previous plan must have
// no unfinished task.
func (store Store) Activate(plan domain.CrewPlan) error {
	return store.withStateLock(func() error {
		if current, err := store.activePlanID(); err == nil && current != plan.ID {
			previous, loadErr := store.LoadPlan(current)
			if loadErr != nil {
				return loadErr
			}
			checkpoints, loadErr := store.loadCheckpoints(previous)
			if loadErr != nil {
				return loadErr
			}
			for _, checkpoint := range checkpoints {
				if !checkpoint.State.Terminal() {
					return fmt.Errorf("crew %s still has unfinished tasks; finish or cancel it first", current)
				}
			}
		} else if err != nil && !errors.Is(err, ErrNoActivePlan) {
			return err
		}
		pointer := struct {
			Schema int    `json:"schema"`
			PlanID string `json:"plan_id"`
		}{domain.CrewSchema, plan.ID}
		return store.write(filepath.Join(store.root, "active.json"), pointer, true)
	})
}

// Records that are replaced in place are read under the state lock, because a
// strict read fails closed when an atomic replace races with it.
func (store Store) ActivePlanID() (string, error) {
	var planID string
	err := store.withStateLock(func() (err error) {
		planID, err = store.activePlanID()
		return err
	})
	return planID, err
}

func (store Store) activePlanID() (string, error) {
	var pointer struct {
		Schema int    `json:"schema"`
		PlanID string `json:"plan_id"`
	}
	if err := store.read(filepath.Join(store.root, "active.json"), &pointer); errors.Is(err, os.ErrNotExist) {
		return "", ErrNoActivePlan
	} else if err != nil {
		return "", err
	}
	if pointer.Schema != domain.CrewSchema || !domain.CrewPlanIDValid(pointer.PlanID) {
		return "", errors.New("crew active pointer is invalid")
	}
	return pointer.PlanID, nil
}

// Checkpoints returns every task's checkpoint; a task without one is queued.
func (store Store) Checkpoints(plan domain.CrewPlan) (map[string]domain.CrewCheckpoint, error) {
	var checkpoints map[string]domain.CrewCheckpoint
	err := store.withStateLock(func() (err error) {
		checkpoints, err = store.loadCheckpoints(plan)
		return err
	})
	return checkpoints, err
}

func (store Store) Checkpoint(plan domain.CrewPlan, taskID string) (domain.CrewCheckpoint, error) {
	var checkpoint domain.CrewCheckpoint
	err := store.withStateLock(func() (err error) {
		checkpoint, err = store.loadCheckpoint(plan, taskID)
		return err
	})
	return checkpoint, err
}

// Update applies change to one task checkpoint under the state lock, advances
// its sequence, and appends an event. change receives the current checkpoint.
func (store Store) Update(plan domain.CrewPlan, taskID string, now time.Time, change func(*domain.CrewCheckpoint) error) (domain.CrewCheckpoint, error) {
	var updated domain.CrewCheckpoint
	err := store.withStateLock(func() error {
		current, err := store.loadCheckpoint(plan, taskID)
		if err != nil {
			return err
		}
		next := current
		if err := change(&next); err != nil {
			return err
		}
		next.Schema, next.PlanID, next.PlanDigest, next.TaskID = domain.CrewSchema, plan.ID, plan.Digest, taskID
		next.Sequence = current.Sequence + 1
		next.Message = bounded(next.Message, 2048)
		next.UpdatedAtUTC = now.UTC().Format(time.RFC3339)
		if problem := checkpointProblem(plan, next); problem != "" {
			return errors.New(problem)
		}
		if err := store.write(store.taskPath(plan.ID, taskID, "checkpoint.json"), next, true); err != nil {
			return err
		}
		event := store.taskPath(plan.ID, taskID, fmt.Sprintf("events/%08d.json", next.Sequence))
		if err := store.write(event, next, false); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		updated = next
		return nil
	})
	return updated, err
}

// OpenDecision records a new decision for a task and moves the task to
// needs-decision in the same locked step.
func (store Store) OpenDecision(plan domain.CrewPlan, taskID string, kind domain.CrewDecisionKind, question string, now time.Time) (domain.CrewDecision, error) {
	options, impact, ok := domain.CrewDecisionFor(kind)
	if !ok || !boundedLine(question, 2048) {
		return domain.CrewDecision{}, errors.New("crew decision kind or question is invalid")
	}
	var decision domain.CrewDecision
	_, err := store.Update(plan, taskID, now, func(checkpoint *domain.CrewCheckpoint) error {
		existing, err := store.loadDecisions(plan.ID)
		if err != nil {
			return err
		}
		count := 0
		for _, item := range existing {
			if item.TaskID == taskID {
				count++
			}
		}
		decision = domain.CrewDecision{
			Schema: domain.CrewSchema, ID: fmt.Sprintf("%s-d%02d", taskID, count+1), TaskID: taskID, Kind: kind,
			Question: question, Options: options, Impact: impact, CreatedAtUTC: now.UTC().Format(time.RFC3339),
		}
		if err := store.write(store.planPath(plan.ID, "decisions/"+decision.ID+".json"), decision, false); err != nil {
			return err
		}
		checkpoint.State = domain.CrewNeedsDecision
		checkpoint.Message = question
		checkpoint.Next = "answer decision " + decision.ID + " with one of: " + strings.Join(options, ", ")
		return nil
	})
	return decision, err
}

// Answer resolves one open decision. retry re-queues the task from its
// checkpoint, restart re-queues it on a fresh attempt from the current target,
// and cancel ends it while preserving its worktree.
func (store Store) Answer(plan domain.CrewPlan, decisionID, answer string, now time.Time) (domain.CrewDecision, domain.CrewCheckpoint, error) {
	var decision domain.CrewDecision
	var checkpoint domain.CrewCheckpoint
	err := store.withStateLock(func() error {
		decisions, err := store.loadDecisions(plan.ID)
		if err != nil {
			return err
		}
		found := false
		for _, item := range decisions {
			if item.ID == decisionID {
				decision, found = item, true
			}
		}
		if !found {
			return fmt.Errorf("crew decision %s does not exist", decisionID)
		}
		if decision.Answer != "" {
			return fmt.Errorf("crew decision %s was already answered", decisionID)
		}
		if !contains(decision.Options, answer) {
			return fmt.Errorf("answer must be one of: %s", strings.Join(decision.Options, ", "))
		}
		current, err := store.loadCheckpoint(plan, decision.TaskID)
		if err != nil {
			return err
		}
		if current.State != domain.CrewNeedsDecision {
			return fmt.Errorf("task %s is not waiting for a decision", decision.TaskID)
		}
		decision.Answer, decision.AnsweredAtUTC = answer, now.UTC().Format(time.RFC3339)
		if err := store.write(store.planPath(plan.ID, "decisions/"+decision.ID+".json"), decision, true); err != nil {
			return err
		}
		next := current
		switch answer {
		case domain.CrewAnswerRetry:
			next.State, next.Next = domain.CrewQueued, "retry from the recorded checkpoint when a worker slot is free"
		case domain.CrewAnswerRestart:
			next = domain.CrewCheckpoint{Sequence: current.Sequence, Attempt: current.Attempt + 1, State: domain.CrewQueued, Next: "restart on a fresh worktree from the current target"}
		case domain.CrewAnswerCancel:
			next.State, next.Next = domain.CrewCancelled, "inspect the retained worktree; the task will not run again"
		}
		next.Schema, next.PlanID, next.PlanDigest, next.TaskID = domain.CrewSchema, plan.ID, plan.Digest, decision.TaskID
		next.Sequence = current.Sequence + 1
		next.Message = "decision " + decision.ID + " answered: " + answer
		next.FailureSignature, next.RepeatedFailures = "", 0
		next.UpdatedAtUTC = now.UTC().Format(time.RFC3339)
		if err := store.write(store.taskPath(plan.ID, decision.TaskID, "checkpoint.json"), next, true); err != nil {
			return err
		}
		event := store.taskPath(plan.ID, decision.TaskID, fmt.Sprintf("events/%08d.json", next.Sequence))
		if err := store.write(event, next, false); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		checkpoint = next
		return nil
	})
	return decision, checkpoint, err
}

// Decisions lists open decisions first, highest impact first, then answered
// decisions; ties keep creation order.
func (store Store) Decisions(planID string) ([]domain.CrewDecision, error) {
	var decisions []domain.CrewDecision
	err := store.withStateLock(func() (err error) {
		decisions, err = store.loadDecisions(planID)
		return err
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(decisions, func(i, j int) bool {
		left, right := decisions[i], decisions[j]
		if (left.Answer == "") != (right.Answer == "") {
			return left.Answer == ""
		}
		if left.Impact != right.Impact {
			return left.Impact > right.Impact
		}
		return left.ID < right.ID
	})
	return decisions, nil
}

func (store Store) SaveReport(planID, taskID, report string) error {
	if !domain.CrewPlanIDValid(planID) || !strings.HasPrefix(taskID, planID+"-t") || report == "" || len(report) > maxReportBytes || strings.ContainsRune(report, 0) {
		return errors.New("crew report is invalid or unbounded")
	}
	path := store.taskPath(planID, taskID, "report.md")
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, []byte(report), 0o600)
	} else if err != nil {
		return err
	}
	return localfile.AtomicReplace(path, []byte(report), 0o600)
}

func (store Store) ReportPath(planID, taskID string) string {
	return store.taskPath(planID, taskID, "report.md")
}

// AcquireSupervisor takes the repository's single supervisor lock.
func (store Store) AcquireSupervisor() (*localfile.Lock, error) {
	if err := localfile.EnsureDirectory(store.root, 0o700); err != nil {
		return nil, err
	}
	lock, err := localfile.AcquireLock(filepath.Join(store.root, "supervisor.lock"))
	if err != nil {
		return nil, errors.New("a crew supervisor is already running for this repository")
	}
	return lock, nil
}

// SupervisorRunning reports whether another process holds the supervisor
// lock. The kernel releases the lock when that process exits.
func (store Store) SupervisorRunning() bool {
	lock, err := store.AcquireSupervisor()
	if err != nil {
		return true
	}
	_ = lock.Close()
	return false
}

func (store Store) SaveSupervisor(info SupervisorInfo) error {
	return store.withStateLock(func() error {
		return store.write(filepath.Join(store.root, "supervisor.json"), info, true)
	})
}

func (store Store) LoadSupervisor() (SupervisorInfo, error) {
	var info SupervisorInfo
	err := store.withStateLock(func() error {
		return store.read(filepath.Join(store.root, "supervisor.json"), &info)
	})
	return info, err
}

func (store Store) loadCheckpoints(plan domain.CrewPlan) (map[string]domain.CrewCheckpoint, error) {
	checkpoints := make(map[string]domain.CrewCheckpoint, len(plan.Tasks))
	for _, task := range plan.Tasks {
		checkpoint, err := store.loadCheckpoint(plan, task.ID)
		if err != nil {
			return nil, err
		}
		checkpoints[task.ID] = checkpoint
	}
	return checkpoints, nil
}

func (store Store) loadCheckpoint(plan domain.CrewPlan, taskID string) (domain.CrewCheckpoint, error) {
	if !planHasTask(plan, taskID) {
		return domain.CrewCheckpoint{}, fmt.Errorf("task %s is not part of crew %s", taskID, plan.ID)
	}
	var checkpoint domain.CrewCheckpoint
	err := store.read(store.taskPath(plan.ID, taskID, "checkpoint.json"), &checkpoint)
	if errors.Is(err, os.ErrNotExist) {
		return domain.CrewCheckpoint{Schema: domain.CrewSchema, PlanID: plan.ID, PlanDigest: plan.Digest, TaskID: taskID, State: domain.CrewQueued, Next: "wait for a worker slot"}, nil
	}
	if err != nil {
		return domain.CrewCheckpoint{}, err
	}
	if problem := checkpointProblem(plan, checkpoint); problem != "" {
		return domain.CrewCheckpoint{}, errors.New(problem)
	}
	return checkpoint, nil
}

func (store Store) loadDecisions(planID string) ([]domain.CrewDecision, error) {
	if !domain.CrewPlanIDValid(planID) {
		return nil, errors.New("crew plan ID is invalid")
	}
	entries, err := os.ReadDir(store.planPath(planID, "decisions"))
	if errors.Is(err, os.ErrNotExist) {
		return []domain.CrewDecision{}, nil
	}
	if err != nil {
		return nil, err
	}
	decisions := make([]domain.CrewDecision, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var decision domain.CrewDecision
		if err := store.read(store.planPath(planID, "decisions/"+entry.Name()), &decision); err != nil {
			return nil, err
		}
		if decision.Schema != domain.CrewSchema || decision.ID+".json" != entry.Name() || !strings.HasPrefix(decision.TaskID, planID+"-t") {
			return nil, fmt.Errorf("crew decision %s is invalid", entry.Name())
		}
		decisions = append(decisions, decision)
	}
	return decisions, nil
}

func checkpointProblem(plan domain.CrewPlan, checkpoint domain.CrewCheckpoint) string {
	switch {
	case checkpoint.Schema != domain.CrewSchema || checkpoint.PlanID != plan.ID || checkpoint.PlanDigest != plan.Digest:
		return "crew checkpoint does not belong to the current plan digest"
	case !planHasTask(plan, checkpoint.TaskID):
		return "crew checkpoint names an unknown task"
	case !checkpoint.State.Valid() || checkpoint.Sequence < 0 || checkpoint.Attempt < 0 || checkpoint.Attempt > 64 || checkpoint.RepeatedFailures < 0:
		return "crew checkpoint state or counters are invalid"
	case !boundedLine(checkpoint.Next, 1024):
		return "crew checkpoint has no next action"
	}
	return ""
}

// stateMutex serialises updates within one process; the file lock serialises
// them between the supervisor and CLI processes.
var stateMutex sync.Mutex

// withStateLock serialises read-modify-write of crew records. The file lock is
// non-blocking, so contention from another process is retried for a bounded
// time.
func (store Store) withStateLock(action func() error) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if err := localfile.EnsureDirectory(store.root, 0o700); err != nil {
		return err
	}
	deadline := time.Now().Add(stateLockWait)
	for {
		lock, err := localfile.AcquireLock(filepath.Join(store.root, "state.lock"))
		if err == nil {
			actionErr := action()
			closeErr := lock.Close()
			if actionErr != nil {
				return actionErr
			}
			return closeErr
		}
		if time.Now().After(deadline) {
			return errors.New("crew state is locked by another Level 7 process")
		}
		time.Sleep(stateLockPeriod)
	}
}

func (store Store) planPath(planID, relative string) string {
	return filepath.Join(store.root, "plans", planID, filepath.FromSlash(relative))
}

func (store Store) taskPath(planID, taskID, relative string) string {
	return store.planPath(planID, "tasks/"+taskID+"/"+relative)
}

func (store Store) write(path string, value any, replace bool) error {
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > maxRecordBytes {
		return errors.New("crew record is invalid or unbounded")
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

func (store Store) read(path string, target any) error {
	data, err := localfile.Read(path, maxRecordBytes)
	if err != nil {
		return err
	}
	return localfile.DecodeJSON(data, target)
}

func planHasTask(plan domain.CrewPlan, taskID string) bool {
	for _, task := range plan.Tasks {
		if task.ID == taskID {
			return true
		}
	}
	return false
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func boundedLine(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && !strings.ContainsAny(value, "\x00\r\n")
}

func bounded(value string, limit int) string {
	value = strings.Map(func(character rune) rune {
		if character == '\n' || character == '\r' || character == 0 {
			return ' '
		}
		return character
	}, value)
	if len(value) <= limit {
		return value
	}
	for limit > 0 && (value[limit]&0xC0) == 0x80 {
		limit--
	}
	return value[:limit]
}
