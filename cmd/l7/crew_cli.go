package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/headlessworker"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	defaultCrewTarget      = "l7/crew"
	defaultCrewWaitSeconds = 540
	crewSupervisorStart    = 10 * time.Second
	crewSupervisorStop     = 60 * time.Second
)

type crewTaskView struct {
	ID           string           `json:"id"`
	Shape        domain.CrewShape `json:"shape"`
	Title        string           `json:"title"`
	State        domain.CrewState `json:"state"`
	Attempt      int              `json:"attempt"`
	Provider     string           `json:"provider"`
	Model        string           `json:"model"`
	Worktree     string           `json:"worktree"`
	Candidate    string           `json:"candidate"`
	Verification string           `json:"verification"`
	Report       string           `json:"report,omitempty"`
	Message      string           `json:"message"`
	Next         string           `json:"next"`
}

type crewStatusView struct {
	PlanID        string                `json:"plan_id"`
	PlanDigest    string                `json:"plan_digest"`
	TargetBranch  string                `json:"target_branch"`
	BaseCommit    string                `json:"base_commit"`
	MaxWorkers    int                   `json:"max_workers"`
	Supervisor    bool                  `json:"supervisor_running"`
	SupervisorPID int                   `json:"supervisor_pid,omitempty"`
	Tasks         []crewTaskView        `json:"tasks"`
	OpenDecisions []domain.CrewDecision `json:"open_decisions"`
	Token         string                `json:"token"`
}

func crewCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) == 0 {
		return orchestrationEnvelope{}, errors.New(crewUsage)
	}
	configuration, err := requireOrchestration(location.Root)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	action, options := arguments[0], arguments[1:]
	switch action {
	case "plan", "start", "answer", "resume", "supervise", "attach", "release":
		if !configuration.Features.Crew {
			return orchestrationEnvelope{}, errors.New("crew is default OFF; set features.crew to true in .l7/orchestration.json")
		}
	}
	switch action {
	case "attach":
		return crewAttach(ctx, configuration, store, options)
	case "release":
		return crewRelease(location, store, options)
	case "view":
		return crewView(ctx, location, store, options)
	case "plan":
		return crewPlan(location, configuration, store, options)
	case "start":
		return crewStart(ctx, location, store, options)
	case "status":
		if len(options) != 0 {
			return orchestrationEnvelope{}, errors.New("crew status accepts no options")
		}
		view, err := crewStatus(store)
		if errors.Is(err, crew.ErrNoActivePlan) {
			return passEnvelope("crew status", "L7-CREW-000", "idle", "no crew plan is active", "run l7 crew plan --objective <file>", nil), nil
		}
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		return passEnvelope("crew status", "L7-CREW-000", crewOverallState(view), "crew status loaded from private state", crewNext(view), view), nil
	case "wait":
		return crewWait(ctx, store, options)
	case "decisions":
		if len(options) != 0 {
			return orchestrationEnvelope{}, errors.New("crew decisions accepts no options")
		}
		plan, err := activeCrewPlan(store)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		decisions, err := store.Decisions(plan.ID)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		next := "no decision is open"
		for _, decision := range decisions {
			if decision.Answer == "" {
				next = fmt.Sprintf("run l7 crew answer --decision %s --choice <%s>", decision.ID, strings.Join(decision.Options, "|"))
				break
			}
		}
		return passEnvelope("crew decisions", "L7-CREW-000", "listed", "open decisions are listed first, highest impact first", next, decisions), nil
	case "answer":
		return crewAnswer(location, store, options)
	case "resume":
		if len(options) != 0 {
			return orchestrationEnvelope{}, errors.New("crew resume accepts no options")
		}
		plan, err := activeCrewPlan(store)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		if _, err := store.LoadApproval(plan); err != nil {
			return orchestrationEnvelope{}, fmt.Errorf("crew plan lacks current owner approval: %w", err)
		}
		pid, err := launchCrewSupervisor(store, location.Root)
		if errors.Is(err, crew.ErrSupervisorRunning) {
			return passEnvelope("crew resume", "L7-CREW-000", "running", "the crew supervisor is already running", "run l7 crew wait", nil), nil
		}
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		return passEnvelope("crew resume", "L7-CREW-000", "running", fmt.Sprintf("crew supervisor %d resumed unfinished tasks", pid), "run l7 crew wait", map[string]int{"supervisor_pid": pid}), nil
	case "cancel":
		if len(options) != 0 {
			return orchestrationEnvelope{}, errors.New("crew cancel accepts no options")
		}
		plan, err := activeCrewPlan(store)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		if _, err := store.StopSupervisor(crewSupervisorStop); err != nil {
			return orchestrationEnvelope{}, err
		}
		cancelled, err := store.CancelUnfinished(plan, time.Now())
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		message := fmt.Sprintf("crew stopped; %d unfinished task(s) cancelled; worktrees and evidence kept", cancelled)
		return passEnvelope("crew cancel", "L7-CREW-000", "cancelled", message, "inspect "+store.WorktreeRoot()+" and the target branch "+plan.TargetBranch, nil), nil
	case "supervise":
		if len(options) != 0 {
			return orchestrationEnvelope{}, errors.New("crew supervise accepts no options")
		}
		executor, err := headlessworker.NewCrew(location.Root, location.CommonDir, configuration)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		plan, err := crew.Supervise(ctx, store, crew.NewEngine(), executor, time.Now())
		if err != nil && !errors.Is(err, context.Canceled) {
			return orchestrationEnvelope{}, err
		}
		return passEnvelope("crew supervise", "L7-CREW-000", "stopped", "crew supervisor finished "+plan.ID, "run l7 crew status", nil), nil
	default:
		return orchestrationEnvelope{}, errors.New(crewUsage)
	}
}

const crewUsage = "crew requires plan, start, status, wait, watch, view, decisions, answer, attach, release, resume, or cancel"

func crewPlan(location domain.RepositoryLocation, configuration orchestrationconfig.File, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	limits := configuration.EffectiveCrew()
	request := crew.PlanRequest{BaseCommit: location.Head, TargetBranch: defaultCrewTarget, MaxWorkers: limits.MaxWorkers, RepairRounds: limits.RepairRounds}
	seen := map[string]bool{}
	for index := 0; index < len(arguments); index += 2 {
		flag := arguments[index]
		if index+1 >= len(arguments) {
			return orchestrationEnvelope{}, fmt.Errorf("crew plan option %s is missing its value", flag)
		}
		value := arguments[index+1]
		if flag != "--command-json" && seen[flag] {
			return orchestrationEnvelope{}, fmt.Errorf("duplicate crew plan option %s", flag)
		}
		seen[flag] = true
		switch flag {
		case "--objective":
			request.ObjectivePath = filepath.ToSlash(value)
		case "--target":
			request.TargetBranch = value
		case "--command-json":
			var argv []string
			if json.Unmarshal([]byte(value), &argv) != nil || len(argv) == 0 {
				return orchestrationEnvelope{}, errors.New("--command-json must be one non-empty JSON argv array")
			}
			request.DefaultVerification = append(request.DefaultVerification, argv)
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown crew plan option %s", flag)
		}
	}
	if request.ObjectivePath == "" {
		return orchestrationEnvelope{}, errors.New("crew plan requires --objective <file>")
	}
	objectivePath := filepath.Join(location.Root, filepath.FromSlash(request.ObjectivePath))
	if filepath.IsAbs(request.ObjectivePath) {
		objectivePath = filepath.Clean(request.ObjectivePath)
		if relative, err := filepath.Rel(store.ObjectiveRoot(), objectivePath); err == nil && !strings.HasPrefix(relative, "..") && !strings.Contains(relative, string(filepath.Separator)) {
			request.ObjectivePath = "crew-objectives/" + relative
		} else if relative, err := filepath.Rel(location.Root, objectivePath); err == nil && !strings.HasPrefix(relative, "..") {
			request.ObjectivePath = filepath.ToSlash(relative)
		} else {
			return orchestrationEnvelope{}, errors.New("crew objective must be inside the repository or " + store.ObjectiveRoot())
		}
	}
	data, err := localfile.Read(objectivePath, 1<<20)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	request.Objective = data
	plan, err := crew.NewPlanner().Plan(request)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	for _, task := range plan.Tasks {
		if pattern, found := headlessworker.CrewScopeProtected(task.AllowedPaths); found {
			return orchestrationEnvelope{}, fmt.Errorf("task %q scope %s covers a protected path; protected work is Tier 3 and outside the crew", task.Title, pattern)
		}
	}
	if err := store.SavePlan(plan); errors.Is(err, os.ErrExist) {
		if plan, err = store.LoadPlan(plan.ID); err != nil {
			return orchestrationEnvelope{}, err
		}
	} else if err != nil {
		return orchestrationEnvelope{}, err
	}
	warning := fmt.Sprintf("WARNING: approval starts up to %d autonomous workers in parallel. Each ship task runs in its own worktree, is verified and independently reviewed, then fast-forwards local branch %s. Nothing is pushed, published, released, or deployed.", plan.MaxWorkers, plan.TargetBranch)
	next := "run l7 crew start --plan " + plan.ID + " --digest " + plan.Digest + " --owner <name> --role <role> --confirm"
	return passEnvelope("crew plan", "L7-CREW-000", "planned", warning, next, plan), nil
}

func crewStart(ctx context.Context, location domain.RepositoryLocation, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	values := map[string]string{}
	confirmed := false
	for index := 0; index < len(arguments); index++ {
		if arguments[index] == "--confirm" {
			confirmed = true
			continue
		}
		if index+1 >= len(arguments) || values[arguments[index]] != "" {
			return orchestrationEnvelope{}, errors.New("crew start option is missing or duplicated")
		}
		switch arguments[index] {
		case "--plan", "--digest", "--owner", "--role":
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown crew start option %s", arguments[index])
		}
		values[arguments[index]] = arguments[index+1]
		index++
	}
	if !confirmed || values["--plan"] == "" || values["--digest"] == "" || values["--owner"] == "" || values["--role"] == "" {
		return orchestrationEnvelope{}, errors.New("crew start requires --plan, --digest, --owner, --role, and --confirm")
	}
	plan, err := store.LoadPlan(values["--plan"])
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if plan.BaseCommit != location.Head {
		return orchestrationEnvelope{}, errors.New("crew plan base is stale; re-plan against the exact current head")
	}
	if err := ensureCrewTarget(ctx, location.Root, plan.TargetBranch, plan.BaseCommit); err != nil {
		return orchestrationEnvelope{}, err
	}
	if _, err := store.Approve(plan, values["--digest"], values["--owner"], values["--role"], time.Now()); errors.Is(err, os.ErrExist) {
		if _, err := store.LoadApproval(plan); err != nil || values["--digest"] != plan.Digest {
			return orchestrationEnvelope{}, errors.New("an approval for a different digest already exists for this plan")
		}
	} else if err != nil {
		return orchestrationEnvelope{}, err
	}
	if err := store.Activate(plan); err != nil {
		return orchestrationEnvelope{}, err
	}
	pid, err := launchCrewSupervisor(store, location.Root)
	if err != nil && !errors.Is(err, crew.ErrSupervisorRunning) {
		return orchestrationEnvelope{}, err
	}
	message := fmt.Sprintf("crew approved by %s and supervisor %d started", values["--owner"], pid)
	return passEnvelope("crew start", "L7-CREW-000", "running", message, "run l7 crew wait", map[string]any{"plan_id": plan.ID, "supervisor_pid": pid, "target_branch": plan.TargetBranch}), nil
}

func crewWait(ctx context.Context, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	since, timeout := "", defaultCrewWaitSeconds
	for index := 0; index < len(arguments); index += 2 {
		if index+1 >= len(arguments) {
			return orchestrationEnvelope{}, fmt.Errorf("crew wait option %s is missing its value", arguments[index])
		}
		switch arguments[index] {
		case "--since":
			since = arguments[index+1]
		case "--timeout":
			value, err := strconv.Atoi(arguments[index+1])
			if err != nil || value < 0 || value > 3600 {
				return orchestrationEnvelope{}, errors.New("--timeout must be 0 to 3600 seconds")
			}
			timeout = value
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown crew wait option %s", arguments[index])
		}
	}
	plan, err := activeCrewPlan(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	result, err := crew.Wait(ctx, store, plan, since, time.Duration(timeout)*time.Second, time.Second)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	view, err := crewStatus(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	view.Token = result.Token
	next := crewNext(view)
	if result.Reason == "timeout" {
		next = "run l7 crew wait --since " + result.Token
	}
	data := map[string]any{"reason": result.Reason, "changed": result.Changed, "status": view}
	return passEnvelope("crew wait", "L7-CREW-000", result.Reason, "crew wait returned: "+result.Reason, next, data), nil
}

func crewAnswer(location domain.RepositoryLocation, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	values := map[string]string{}
	for index := 0; index < len(arguments); index += 2 {
		if index+1 >= len(arguments) || values[arguments[index]] != "" || (arguments[index] != "--decision" && arguments[index] != "--choice") {
			return orchestrationEnvelope{}, errors.New("crew answer requires exactly --decision <id> and --choice <option>")
		}
		values[arguments[index]] = arguments[index+1]
	}
	if values["--decision"] == "" || values["--choice"] == "" {
		return orchestrationEnvelope{}, errors.New("crew answer requires exactly --decision <id> and --choice <option>")
	}
	plan, err := activeCrewPlan(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	decision, checkpoint, err := store.Answer(plan, values["--decision"], values["--choice"], time.Now())
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	message := "decision " + decision.ID + " answered: " + decision.Answer
	if checkpoint.State == domain.CrewQueued {
		pid, startErr := launchCrewSupervisor(store, location.Root)
		switch {
		case errors.Is(startErr, crew.ErrSupervisorRunning):
			message += "; the running supervisor will pick the task up"
		case startErr != nil:
			return orchestrationEnvelope{}, fmt.Errorf("%s, but the supervisor did not start: %w", message, startErr)
		default:
			message += fmt.Sprintf("; supervisor %d started", pid)
		}
	}
	return passEnvelope("crew answer", "L7-CREW-000", string(checkpoint.State), message, "run l7 crew wait", map[string]any{"decision": decision, "task": checkpoint}), nil
}

func crewStatus(store crew.Store) (crewStatusView, error) {
	plan, err := activeCrewPlan(store)
	if err != nil {
		return crewStatusView{}, err
	}
	checkpoints, err := store.Checkpoints(plan)
	if err != nil {
		return crewStatusView{}, err
	}
	decisions, err := store.Decisions(plan.ID)
	if err != nil {
		return crewStatusView{}, err
	}
	view := crewStatusView{
		PlanID: plan.ID, PlanDigest: plan.Digest, TargetBranch: plan.TargetBranch, BaseCommit: plan.BaseCommit, MaxWorkers: plan.MaxWorkers,
		Supervisor: store.SupervisorRunning(), Tasks: []crewTaskView{}, OpenDecisions: []domain.CrewDecision{}, Token: crew.Token(plan, checkpoints),
	}
	if info, err := store.LoadSupervisor(); err == nil && view.Supervisor && info.PlanID == plan.ID {
		view.SupervisorPID = info.PID
	}
	for _, task := range plan.Tasks {
		checkpoint := checkpoints[task.ID]
		item := crewTaskView{
			ID: task.ID, Shape: task.Shape, Title: task.Title, State: checkpoint.State, Attempt: checkpoint.Attempt,
			Provider: checkpoint.ProviderID, Model: checkpoint.ModelID, Worktree: checkpoint.Worktree, Candidate: checkpoint.CandidateCommit,
			Verification: checkpoint.Verification, Message: checkpoint.Message, Next: checkpoint.Next,
		}
		if task.Shape == domain.CrewScout && checkpoint.State == domain.CrewDone {
			item.Report = store.ReportPath(plan.ID, task.ID)
		}
		view.Tasks = append(view.Tasks, item)
	}
	for _, decision := range decisions {
		if decision.Answer == "" {
			view.OpenDecisions = append(view.OpenDecisions, decision)
		}
	}
	return view, nil
}

func crewOverallState(view crewStatusView) string {
	finished := true
	for _, task := range view.Tasks {
		finished = finished && task.State.Terminal()
	}
	switch {
	case finished:
		return "finished"
	case len(view.OpenDecisions) != 0:
		return "needs-decision"
	case view.Supervisor:
		return "running"
	default:
		return "stopped"
	}
}

func crewNext(view crewStatusView) string {
	if len(view.OpenDecisions) != 0 {
		decision := view.OpenDecisions[0]
		return fmt.Sprintf("ask the owner, then run l7 crew answer --decision %s --choice <%s>: %s", decision.ID, strings.Join(decision.Options, "|"), decision.Question)
	}
	switch crewOverallState(view) {
	case "finished":
		return "review the merged work, then fast-forward your branch: git merge --ff-only " + view.TargetBranch
	case "stopped":
		return "run l7 crew resume"
	default:
		return "run l7 crew wait --since " + view.Token
	}
}

func activeCrewPlan(store crew.Store) (domain.CrewPlan, error) {
	planID, err := store.ActivePlanID()
	if err != nil {
		return domain.CrewPlan{}, err
	}
	return store.LoadPlan(planID)
}

func startCrewSupervisor(store crew.Store, root string) (int, error) {
	executable, err := os.Executable()
	if err != nil {
		return 0, err
	}
	if executable, err = filepath.EvalSymlinks(executable); err != nil {
		return 0, err
	}
	return store.StartSupervisor(executable, []string{"crew", "supervise"}, root, crewSupervisorStart)
}

// ensureCrewTarget prepares the local target branch so the crew builds on the
// approved base: it creates a missing branch at base, fast-forwards one that is
// behind base, keeps one that already contains base, and refuses a diverged or
// checked-out target.
func ensureCrewTarget(ctx context.Context, root, branch, base string) error {
	if !domain.CrewBranchValid(branch) {
		return errors.New("crew target branch is unsafe")
	}
	git, err := processadapter.Resolve("git")
	if err != nil {
		return err
	}
	run := func(arguments ...string) (processadapter.Result, error) {
		return (processadapter.Runner{}).Run(ctx, processadapter.Request{Executable: git.Path, Arguments: arguments, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 8 << 20, Timeout: 30 * time.Second})
	}
	ref := "refs/heads/" + branch
	current, err := run("show-ref", "--verify", "--hash", ref)
	if err != nil {
		return err
	}
	if current.ExitCode != 0 {
		created, err := run("branch", "--no-track", branch, base)
		if err != nil || created.ExitCode != 0 {
			return fmt.Errorf("cannot create local crew target branch %s", branch)
		}
	} else if target := strings.TrimSpace(string(current.Stdout)); target != base {
		behind, err := run("merge-base", "--is-ancestor", target, base)
		if err != nil {
			return err
		}
		ahead, err := run("merge-base", "--is-ancestor", base, target)
		if err != nil {
			return err
		}
		switch {
		case behind.ExitCode == 0:
			client, err := gitadapter.New("", gitadapter.DefaultMaxOutput, gitadapter.DefaultMaxPaths)
			if err != nil {
				return err
			}
			if err := client.AdvanceMerge(ctx, domain.MergeRequest{Root: root, TargetBranch: branch, ExpectedOld: target, Candidate: base, MaxOutputBytes: 8 << 20}); err != nil {
				return fmt.Errorf("cannot fast-forward crew target branch %s to the approved base: %w", branch, err)
			}
		case ahead.ExitCode != 0:
			return fmt.Errorf("crew target branch %s has diverged from HEAD; merge it into your branch or delete it before starting a new crew", branch)
		}
	}
	listing, err := run("worktree", "list", "--porcelain")
	if err != nil || listing.ExitCode != 0 {
		return errors.New("cannot inspect checked-out worktree branches")
	}
	for _, line := range strings.Split(string(listing.Stdout), "\n") {
		if line == "branch refs/heads/"+branch {
			return fmt.Errorf("crew target branch %s is checked out; use a branch no worktree has checked out, such as %s", branch, defaultCrewTarget)
		}
	}
	return nil
}
