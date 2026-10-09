package crew

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type OutcomeKind string

const (
	// OutcomeBuilt means a ship candidate is verified, reviewed, and ready to
	// integrate.
	OutcomeBuilt OutcomeKind = "built"
	// OutcomeMerged means the candidate fast-forwarded the target branch.
	OutcomeMerged OutcomeKind = "merged"
	// OutcomeReported means a scout report is ready.
	OutcomeReported OutcomeKind = "reported"
	// OutcomeDelivered means a pull request is open for the candidate.
	OutcomeDelivered OutcomeKind = "delivered"
	OutcomeQuota     OutcomeKind = "quota"
	OutcomeDecision  OutcomeKind = "decision"
	OutcomeFailed    OutcomeKind = "failed"
)

type Outcome struct {
	Kind             OutcomeKind
	Decision         domain.CrewDecisionKind
	ProviderID       string
	ModelID          string
	SessionID        string
	Worktree         string
	CandidateCommit  string
	Verification     string
	FailureSignature string
	QuotaResetAtUTC  string
	Report           string
	Message          string
	PullRequest      int
	PullRequestURL   string
}

// PullRequestStatus is the forge's view of one delivered task.
type PullRequestStatus struct {
	// State is OPEN, MERGED, or CLOSED.
	State string
	Head  string
	// Checks summarizes check results, such as "passed: 3".
	Checks string
	Failed bool
}

// Executor performs the repository and provider work for one task. Build must
// be resumable from the checkpoint it receives. Integrate is only called for a
// built ship task and never concurrently with another Integrate. Track reads
// the pull request of a task in pr-open.
type Executor interface {
	Build(context.Context, domain.CrewPlan, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error)
	Integrate(context.Context, domain.CrewPlan, domain.CrewTask, domain.CrewCheckpoint) (Outcome, error)
	Track(context.Context, domain.CrewPlan, domain.CrewTask, domain.CrewCheckpoint) (PullRequestStatus, error)
}

// errNotQueued skips an admission whose task changed state after scheduling.
var errNotQueued = errors.New("crew task is no longer queued")

// errUnchanged lets an update leave a checkpoint as it is.
var errUnchanged = errors.New("crew checkpoint unchanged")

type Engine struct {
	now             func() time.Time
	wait            func(context.Context, time.Time) error
	poll            time.Duration
	track           time.Duration
	noProgressLimit int
	failureLimit    int
}

func NewEngine() Engine { return NewEngineWith(nil, nil) }

func NewEngineWith(now func() time.Time, wait func(context.Context, time.Time) error) Engine {
	if now == nil {
		now = time.Now
	}
	if wait == nil {
		wait = waitUntil
	}
	return Engine{now: now, wait: wait, poll: 2 * time.Second, track: time.Minute, noProgressLimit: 3, failureLimit: 8}
}

// Run supervises plan until no task can make progress without the owner. The
// caller must hold the supervisor lock. Cancelling ctx pauses running tasks so
// that a later Run resumes them from their checkpoints.
func (engine Engine) Run(ctx context.Context, store Store, plan domain.CrewPlan, executor Executor) error {
	if executor == nil {
		return errors.New("crew executor is required")
	}
	if err := ValidatePlan(plan); err != nil {
		return err
	}
	if _, err := store.LoadApproval(plan); err != nil {
		return fmt.Errorf("crew plan lacks current owner approval: %w", err)
	}
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return err
	}
	for id, checkpoint := range checkpoints {
		switch {
		case checkpoint.State.Active() && checkpoint.AttachRequested:
			if _, err := store.Update(plan, id, engine.now(), attachTo(id)); err != nil {
				return err
			}
		case checkpoint.State.Active() || checkpoint.State == domain.CrewPaused:
			if _, err := store.Update(plan, id, engine.now(), func(next *domain.CrewCheckpoint) error {
				next.State, next.Message, next.Next = domain.CrewQueued, "recovered after an interrupted supervisor", "resume from the recorded checkpoint"
				return nil
			}); err != nil {
				return err
			}
		}
	}
	var merge sync.Mutex
	finished := make(chan string)
	running := make(map[string]context.CancelFunc)
	ticker := time.NewTicker(engine.poll)
	defer ticker.Stop()
	var loopErr error
	var lastTrack time.Time
	for {
		tracking := false
		if ctx.Err() == nil && loopErr == nil {
			loopErr = engine.admit(ctx, store, plan, executor, &merge, running, finished)
		}
		if ctx.Err() == nil && loopErr == nil {
			tracking, lastTrack = engine.trackPullRequests(ctx, store, plan, executor, lastTrack)
		}
		if len(running) == 0 && !tracking {
			if loopErr != nil {
				return loopErr
			}
			return ctx.Err()
		}
		select {
		case id := <-finished:
			running[id]()
			delete(running, id)
		case <-ctx.Done():
			for len(running) > 0 {
				id := <-finished
				running[id]()
				delete(running, id)
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (engine Engine) admit(ctx context.Context, store Store, plan domain.CrewPlan, executor Executor, merge *sync.Mutex, running map[string]context.CancelFunc, finished chan<- string) error {
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return err
	}
	states := make(map[string]domain.CrewState, len(checkpoints))
	for id, checkpoint := range checkpoints {
		states[id] = checkpoint.State
		cancel, isRunning := running[id]
		switch {
		case isRunning && checkpoint.AttachRequested && checkpoint.State != domain.CrewMerging:
			cancel()
		case checkpoint.State.Active() && !isRunning && checkpoint.AttachRequested:
			states[id] = domain.CrewAttached
			_, _ = store.Update(plan, id, engine.now(), attachTo(id))
		case checkpoint.State.Active() && !isRunning:
			states[id] = domain.CrewPaused
			_, _ = store.Update(plan, id, engine.now(), func(next *domain.CrewCheckpoint) error {
				next.State, next.Message, next.Next = domain.CrewPaused, "worker exited without recording a final state", "run l7 crew resume"
				return nil
			})
		}
	}
	for _, id := range domain.NextCrewAdmissions(plan.Tasks, states, plan.MaxWorkers) {
		task, _ := planTask(plan, id)
		if _, err := store.Update(plan, id, engine.now(), func(next *domain.CrewCheckpoint) error {
			if next.State != domain.CrewQueued {
				return errNotQueued
			}
			next.State, next.Message, next.Next = domain.CrewRunning, "worker started", "implement, verify, and review the task"
			return nil
		}); errors.Is(err, errNotQueued) {
			continue
		} else if err != nil {
			return err
		}
		taskContext, cancel := context.WithCancel(ctx)
		running[id] = cancel
		go func(task domain.CrewTask) {
			engine.runTask(taskContext, store, plan, task, executor, merge)
			finished <- task.ID
		}(task)
	}
	return nil
}

// trackPullRequests polls the pull requests of pr-open tasks at most once per
// track interval. It reports whether any task is waiting on a pull request,
// which keeps the supervisor alive.
func (engine Engine) trackPullRequests(ctx context.Context, store Store, plan domain.CrewPlan, executor Executor, last time.Time) (bool, time.Time) {
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return false, last
	}
	open := []domain.CrewTask{}
	for _, task := range plan.Tasks {
		if checkpoints[task.ID].State == domain.CrewPROpen {
			open = append(open, task)
		}
	}
	if len(open) == 0 {
		return false, last
	}
	if !last.IsZero() && time.Since(last) < engine.track {
		return true, last
	}
	for _, task := range open {
		checkpoint := checkpoints[task.ID]
		status, err := executor.Track(ctx, plan, task, checkpoint)
		if err != nil {
			continue
		}
		engine.applyPullRequest(store, plan, task.ID, checkpoint, status)
	}
	return true, time.Now()
}

func (engine Engine) applyPullRequest(store Store, plan domain.CrewPlan, taskID string, checkpoint domain.CrewCheckpoint, status PullRequestStatus) {
	// The forge call took time; act only if the task is still waiting on the
	// same delivered head, so an attach or merge made meanwhile wins.
	current, err := store.Checkpoint(plan, taskID)
	if err != nil || current.State != domain.CrewPROpen || current.CandidateCommit != checkpoint.CandidateCommit {
		return
	}
	checkpoint = current
	pull := "pull request #" + strconv.Itoa(checkpoint.PullRequest)
	switch {
	case status.State == "MERGED":
		engine.record(store, plan, taskID, Outcome{Message: pull + " was merged"}, domain.CrewDone, "pull the base branch to get the change")
	case status.State == "CLOSED":
		engine.record(store, plan, taskID, Outcome{Message: pull + " was closed without merging"}, domain.CrewCancelled, "the task will not run again; its branch and worktree are kept")
	case status.State != "OPEN":
		return
	case status.Head != checkpoint.CandidateCommit:
		engine.decide(store, plan, taskID, domain.CrewDecisionBlocked, pull+" head moved to "+shortCommit(status.Head)+" outside Level 7")
	case status.Checks != checkpoint.Checks:
		updated, err := store.Update(plan, taskID, engine.now(), func(next *domain.CrewCheckpoint) error {
			if next.State != domain.CrewPROpen {
				return errUnchanged
			}
			next.Checks, next.Message = status.Checks, pull+" checks: "+status.Checks
			return nil
		})
		if err == nil && status.Failed && updated.State == domain.CrewPROpen {
			engine.decide(store, plan, taskID, domain.CrewDecisionCIFailed, "checks failed on "+pull+": "+status.Checks)
		}
	}
}

func shortCommit(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

// runTask drives one task until it finishes, needs the owner, or its context
// ends. The context ends when the supervisor stops or the owner attaches.
func (engine Engine) runTask(ctx context.Context, store Store, plan domain.CrewPlan, task domain.CrewTask, executor Executor, merge *sync.Mutex) {
	failures, rebuilds := 0, 0
	for {
		if ctx.Err() != nil {
			engine.halt(store, plan, task.ID, Outcome{})
			return
		}
		checkpoint, err := store.Checkpoint(plan, task.ID)
		if err != nil || !checkpoint.State.Active() {
			return
		}
		outcome, buildErr := executor.Build(ctx, plan, task, checkpoint)
		outcome = normalizeOutcome(outcome, buildErr)
		if ctx.Err() != nil {
			engine.halt(store, plan, task.ID, outcome)
			return
		}
		switch {
		case outcome.Kind == OutcomeReported && task.Shape == domain.CrewScout:
			if err := store.SaveReport(plan.ID, task.ID, outcome.Report); err != nil {
				engine.decide(store, plan, task.ID, domain.CrewDecisionBlocked, "scout report could not be saved: "+err.Error())
				return
			}
			engine.record(store, plan, task.ID, outcome, domain.CrewDone, "read the report at "+store.ReportPath(plan.ID, task.ID))
			return
		case outcome.Kind == OutcomeDelivered && task.Shape == domain.CrewShip:
			engine.deliver(store, plan, task.ID, outcome)
			return
		case outcome.Kind == OutcomeBuilt && task.Shape == domain.CrewShip:
			checkpoint = engine.record(store, plan, task.ID, outcome, domain.CrewMerging, "wait for the merge queue")
			if checkpoint.TaskID == "" {
				engine.pause(store, plan, task.ID, "could not record the built candidate before merging")
				return
			}
			// A started merge step runs to completion so that stopping the
			// supervisor or attaching never leaves a half-applied rebase.
			merge.Lock()
			integrated, integrateErr := executor.Integrate(context.WithoutCancel(ctx), plan, task, checkpoint)
			merge.Unlock()
			outcome = normalizeOutcome(integrated, integrateErr)
			switch outcome.Kind {
			case OutcomeMerged:
				engine.record(store, plan, task.ID, outcome, domain.CrewDone, "fast-forward your branch from "+plan.TargetBranch+" when ready")
				return
			case OutcomeDelivered:
				engine.deliver(store, plan, task.ID, outcome)
				return
			case OutcomeBuilt:
				// The target moved and the rebased candidate needs another
				// review or repair before it may merge.
				rebuilds++
				if rebuilds > engine.failureLimit {
					engine.decide(store, plan, task.ID, domain.CrewDecisionNoProgress, "the target kept moving; the candidate was rebuilt "+strconv.Itoa(rebuilds-1)+" times without merging")
					return
				}
				engine.record(store, plan, task.ID, outcome, domain.CrewRunning, "rebuild the candidate before merging")
				continue
			}
			if !engine.handle(ctx, store, plan, task, outcome, &failures) {
				return
			}
		default:
			if !engine.handle(ctx, store, plan, task, outcome, &failures) {
				return
			}
		}
	}
}

// handle applies a non-terminal outcome and reports whether the task should
// keep running.
func (engine Engine) handle(ctx context.Context, store Store, plan domain.CrewPlan, task domain.CrewTask, outcome Outcome, failures *int) bool {
	switch outcome.Kind {
	case OutcomeQuota:
		reset, err := time.Parse(time.RFC3339, outcome.QuotaResetAtUTC)
		if err != nil || !reset.After(engine.now()) {
			engine.decide(store, plan, task.ID, domain.CrewDecisionBlocked, "provider reported a quota limit without a valid future reset time")
			return false
		}
		outcome.QuotaResetAtUTC = reset.UTC().Format(time.RFC3339)
		engine.record(store, plan, task.ID, outcome, domain.CrewWaitingQuota, "wait for the natural provider reset")
		if err := engine.wait(ctx, reset); err != nil {
			engine.halt(store, plan, task.ID, Outcome{})
			return false
		}
		engine.record(store, plan, task.ID, Outcome{Message: "quota reset reached"}, domain.CrewRunning, "resume the task")
		return true
	case OutcomeDecision:
		kind := outcome.Decision
		if _, _, ok := domain.CrewDecisionFor(kind); !ok {
			kind = domain.CrewDecisionBlocked
		}
		engine.record(store, plan, task.ID, outcome, domain.CrewRunning, "open an owner decision")
		engine.decide(store, plan, task.ID, kind, outcome.Message)
		return false
	case OutcomeFailed:
		*failures++
		checkpoint := engine.record(store, plan, task.ID, outcome, domain.CrewRunning, "retry with recorded failure context")
		if checkpoint.RepeatedFailures >= engine.noProgressLimit || *failures >= engine.failureLimit {
			engine.decide(store, plan, task.ID, domain.CrewDecisionNoProgress, fmt.Sprintf("%d failed attempts without progress; last: %s", *failures, outcome.Message))
			return false
		}
		return true
	default:
		engine.decide(store, plan, task.ID, domain.CrewDecisionBlocked, fmt.Sprintf("executor returned unexpected outcome %q for a %s task", outcome.Kind, task.Shape))
		return false
	}
}

// record stores outcome fields with a new state. Failure signatures are
// counted here so the breaker survives restarts.
func (engine Engine) record(store Store, plan domain.CrewPlan, taskID string, outcome Outcome, state domain.CrewState, next string) domain.CrewCheckpoint {
	checkpoint, err := store.Update(plan, taskID, engine.now(), func(checkpoint *domain.CrewCheckpoint) error {
		applyOutcome(checkpoint, outcome)
		checkpoint.OwnerEdited = false
		if state.Terminal() {
			checkpoint.AttachRequested = false
		}
		if outcome.Kind == OutcomeFailed {
			if checkpoint.FailureSignature == outcome.FailureSignature {
				checkpoint.RepeatedFailures++
			} else {
				checkpoint.FailureSignature, checkpoint.RepeatedFailures = outcome.FailureSignature, 1
			}
		} else if state == domain.CrewDone || state == domain.CrewMerging {
			checkpoint.FailureSignature, checkpoint.RepeatedFailures = "", 0
		}
		checkpoint.QuotaResetAtUTC = ""
		if state == domain.CrewWaitingQuota {
			checkpoint.QuotaResetAtUTC = outcome.QuotaResetAtUTC
		}
		checkpoint.State, checkpoint.Message, checkpoint.Next = state, outcome.Message, next
		return nil
	})
	if err != nil {
		return domain.CrewCheckpoint{}
	}
	return checkpoint
}

// deliver records an open pull request. Checks start pending because the
// candidate head is new to the forge.
func (engine Engine) deliver(store Store, plan domain.CrewPlan, taskID string, outcome Outcome) {
	next := fmt.Sprintf("review pull request #%d, then run l7 crew merge --task %s --head %s --confirm", outcome.PullRequest, taskID, outcome.CandidateCommit)
	_, _ = store.Update(plan, taskID, engine.now(), func(checkpoint *domain.CrewCheckpoint) error {
		applyOutcome(checkpoint, outcome)
		checkpoint.OwnerEdited, checkpoint.FailureSignature, checkpoint.RepeatedFailures = false, "", 0
		checkpoint.State, checkpoint.Checks, checkpoint.Message, checkpoint.Next = domain.CrewPROpen, "pending", outcome.Message, next
		return nil
	})
}

func (engine Engine) pause(store Store, plan domain.CrewPlan, taskID, message string) {
	_, _ = store.Update(plan, taskID, engine.now(), func(checkpoint *domain.CrewCheckpoint) error {
		if checkpoint.State.Terminal() || checkpoint.State == domain.CrewNeedsDecision || checkpoint.State == domain.CrewAttached {
			return errUnchanged
		}
		checkpoint.State, checkpoint.Message, checkpoint.Next = domain.CrewPaused, message, "run l7 crew resume"
		return nil
	})
}

// halt records a task whose context ended. An owner takeover request
// attaches it; otherwise the supervisor stopped and the task pauses for
// resume. Finished, decided, and attached tasks keep their state.
func (engine Engine) halt(store Store, plan domain.CrewPlan, taskID string, outcome Outcome) {
	_, _ = store.Update(plan, taskID, engine.now(), func(checkpoint *domain.CrewCheckpoint) error {
		if checkpoint.State.Terminal() || checkpoint.State == domain.CrewNeedsDecision || checkpoint.State == domain.CrewAttached {
			return errUnchanged
		}
		applyOutcome(checkpoint, outcome)
		if checkpoint.AttachRequested {
			return attachTo(taskID)(checkpoint)
		}
		checkpoint.State, checkpoint.Message, checkpoint.Next = domain.CrewPaused, "supervisor stopped; resume continues from the checkpoint", "run l7 crew resume"
		return nil
	})
}

func attachTo(taskID string) func(*domain.CrewCheckpoint) error {
	return func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State, checkpoint.AttachRequested = domain.CrewAttached, false
		checkpoint.Message, checkpoint.Next = "held by the owner", "work in the task worktree, then run l7 crew release --task "+taskID
		return nil
	}
}

func applyOutcome(checkpoint *domain.CrewCheckpoint, outcome Outcome) {
	for target, value := range map[*string]string{
		&checkpoint.ProviderID: outcome.ProviderID, &checkpoint.ModelID: outcome.ModelID, &checkpoint.SessionID: outcome.SessionID,
		&checkpoint.Worktree: outcome.Worktree, &checkpoint.CandidateCommit: outcome.CandidateCommit, &checkpoint.Verification: outcome.Verification,
		&checkpoint.PullRequestURL: outcome.PullRequestURL,
	} {
		if value != "" {
			*target = value
		}
	}
	if outcome.PullRequest > 0 {
		checkpoint.PullRequest = outcome.PullRequest
	}
}

func (engine Engine) decide(store Store, plan domain.CrewPlan, taskID string, kind domain.CrewDecisionKind, question string) {
	if strings.TrimSpace(question) == "" {
		question = "the task cannot continue without an owner decision"
	}
	if _, err := store.OpenDecision(plan, taskID, kind, bounded(question, 2048), engine.now()); err != nil {
		engine.pause(store, plan, taskID, "could not record a decision: "+err.Error())
	}
}

func normalizeOutcome(outcome Outcome, err error) Outcome {
	if err != nil && (outcome.Kind == "" || outcome.Kind == OutcomeBuilt || outcome.Kind == OutcomeMerged || outcome.Kind == OutcomeReported) {
		outcome.Kind = OutcomeFailed
		if outcome.Message == "" {
			outcome.Message = err.Error()
		}
	}
	if outcome.Kind == "" {
		outcome.Kind = OutcomeFailed
		outcome.Message = "executor returned no outcome"
	}
	if outcome.Kind == OutcomeFailed && outcome.FailureSignature == "" {
		outcome.FailureSignature = fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(outcome.Message)))
	}
	return outcome
}

type WaitResult struct {
	Reason      string                           `json:"reason"`
	Token       string                           `json:"token"`
	Changed     []string                         `json:"changed"`
	Checkpoints map[string]domain.CrewCheckpoint `json:"checkpoints"`
}

// Token lists each task's checkpoint sequence in plan order.
func Token(plan domain.CrewPlan, checkpoints map[string]domain.CrewCheckpoint) string {
	parts := make([]string, len(plan.Tasks))
	for index, task := range plan.Tasks {
		parts[index] = strconv.Itoa(checkpoints[task.ID].Sequence)
	}
	return strings.Join(parts, ".")
}

func parseToken(plan domain.CrewPlan, token string) ([]int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != len(plan.Tasks) {
		return nil, errors.New("wait token does not match this crew plan")
	}
	sequences := make([]int, len(parts))
	for index, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return nil, errors.New("wait token is malformed")
		}
		sequences[index] = value
	}
	return sequences, nil
}

// Wait blocks until a task that changed since token needs attention, no
// supervisor is running, or timeout passes. An empty token waits for the next
// change from now. It reads only local state. Without a supervisor nothing
// will change, so stale running states from a crashed supervisor report idle.
func Wait(ctx context.Context, store Store, plan domain.CrewPlan, token string, timeout, poll time.Duration) (WaitResult, error) {
	if timeout < 0 || timeout > time.Hour || poll <= 0 {
		return WaitResult{}, errors.New("wait timeout must be 0 to 3600 seconds")
	}
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return WaitResult{}, err
	}
	if token == "" {
		token = Token(plan, checkpoints)
	}
	baseline, err := parseToken(plan, token)
	if err != nil {
		return WaitResult{}, err
	}
	deadline := time.Now().Add(timeout)
	for {
		changed, attention := []string{}, false
		for index, task := range plan.Tasks {
			checkpoint := checkpoints[task.ID]
			if checkpoint.Sequence != baseline[index] {
				changed = append(changed, task.ID)
				attention = attention || checkpoint.State.NeedsAttention()
			}
		}
		result := WaitResult{Token: Token(plan, checkpoints), Changed: changed, Checkpoints: checkpoints}
		switch {
		case attention:
			result.Reason = "attention"
			return result, nil
		case !store.SupervisorRunning():
			result.Reason = "idle"
			return result, nil
		case !time.Now().Before(deadline):
			result.Reason = "timeout"
			return result, nil
		}
		timer := time.NewTimer(min(poll, time.Until(deadline)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return WaitResult{}, ctx.Err()
		case <-timer.C:
		}
		if checkpoints, err = store.Checkpoints(plan); err != nil {
			return WaitResult{}, err
		}
	}
}

func planTask(plan domain.CrewPlan, id string) (domain.CrewTask, bool) {
	for _, task := range plan.Tasks {
		if task.ID == id {
			return task, true
		}
	}
	return domain.CrewTask{}, false
}

func waitUntil(ctx context.Context, reset time.Time) error {
	timer := time.NewTimer(time.Until(reset))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
