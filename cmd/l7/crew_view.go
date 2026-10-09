package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	crewAttachWait = 90 * time.Second
	crewWatchPoll  = time.Second
)

// launchCrewSupervisor starts the detached supervisor; tests replace it.
var launchCrewSupervisor = startCrewSupervisor

func crewAttach(ctx context.Context, configuration orchestrationconfig.File, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	taskID, err := crewTaskOption("attach", arguments)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	plan, err := activeCrewPlan(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	checkpoint, err := store.RequestAttach(plan, taskID, store.SupervisorRunning(), time.Now())
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	deadline := time.Now().Add(crewAttachWait)
	for checkpoint.State != domain.CrewAttached {
		if supervised := store.SupervisorRunning(); !checkpoint.State.Active() || !supervised {
			if checkpoint, err = store.RequestAttach(plan, taskID, supervised, time.Now()); err != nil {
				return orchestrationEnvelope{}, err
			}
			continue
		}
		if time.Now().After(deadline) {
			return passEnvelope("crew attach", "L7-CREW-000", "attach-requested", "the supervisor stops the worker after its current merge step", "run l7 crew attach --task "+taskID+" again for the resume command", checkpoint), nil
		}
		select {
		case <-ctx.Done():
			return orchestrationEnvelope{}, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
		if checkpoint, err = store.Checkpoint(plan, taskID); err != nil {
			return orchestrationEnvelope{}, err
		}
	}
	command, resumable := resumeCommand(configuration, checkpoint)
	opened := false
	if socket := os.Getenv("TMUX"); resumable && socket != "" {
		opened = runTmux(ctx, tmuxSocket(socket), checkpoint.Worktree, "new-window", "-n", crewWindowName(plan.ID, taskID), "-c", checkpoint.Worktree, command) == nil
	}
	message := "task held for you; Level 7 changes nothing in it until you release it"
	if opened {
		message += "; the session opened in a new tmux window"
	}
	data := map[string]any{"task": checkpoint, "resume_command": command, "resumable": resumable, "opened_in_tmux": opened}
	return passEnvelope("crew attach", "L7-CREW-000", "attached", message, "when you are done, run l7 crew release --task "+taskID, data), nil
}

func crewRelease(location domain.RepositoryLocation, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	taskID, err := crewTaskOption("release", arguments)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	plan, err := activeCrewPlan(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	checkpoint, err := store.Release(plan, taskID, time.Now())
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	message := "task released; your changes are checked against scope, verified, reviewed, and merged as usual"
	pid, startErr := launchCrewSupervisor(store, location.Root)
	switch {
	case errors.Is(startErr, crew.ErrSupervisorRunning):
	case startErr != nil:
		return orchestrationEnvelope{}, fmt.Errorf("task released, but the supervisor did not start: %w", startErr)
	default:
		message += fmt.Sprintf("; supervisor %d started", pid)
	}
	return passEnvelope("crew release", "L7-CREW-000", string(checkpoint.State), message, "run l7 crew wait", checkpoint), nil
}

// resumeCommand returns the provider's native command for continuing the
// task's session in its worktree, and whether such a session exists.
func resumeCommand(configuration orchestrationconfig.File, checkpoint domain.CrewCheckpoint) (string, bool) {
	if checkpoint.Worktree == "" {
		return "no worktree exists yet; release the task to let a worker start it", false
	}
	directory := "cd " + shellQuote(checkpoint.Worktree)
	var kind domain.ProviderKind
	for _, provider := range configuration.Providers {
		if provider.ID == checkpoint.ProviderID {
			kind = provider.Kind
		}
	}
	switch {
	case kind == domain.ProviderKindCodexAppServer && sessionIDValid(checkpoint.SessionID):
		return directory + " && codex resume " + checkpoint.SessionID, true
	case kind == domain.ProviderKindClaudeCLI && sessionIDValid(checkpoint.SessionID):
		return directory + " && claude --resume " + checkpoint.SessionID, true
	case kind == domain.ProviderKindClaudeCLI:
		return directory + " && claude --continue", true
	default:
		return directory + "  # no resumable provider session; edit the files directly", false
	}
}

func sessionIDValid(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for index, character := range value {
		alphanumeric := (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9')
		if !alphanumeric && (index == 0 || (character != '-' && character != '_' && character != '.')) {
			return false
		}
	}
	return true
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'" }

func crewTaskOption(action string, arguments []string) (string, error) {
	if len(arguments) != 2 || arguments[0] != "--task" || arguments[1] == "" {
		return "", fmt.Errorf("crew %s requires exactly --task <id>", action)
	}
	return arguments[1], nil
}

// runCrewWatch redraws the crew board on every state change until each task
// has finished or the owner interrupts it. It never changes crew state.
func runCrewWatch(ctx context.Context, arguments []string, cwd string, stdout, stderr io.Writer, jsonOutput bool) int {
	fail := func(message, next string) int {
		return writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope("crew", "L7-ORCH-002", "failed", message, next))
	}
	if len(arguments) != 0 || jsonOutput {
		return fail("crew watch is an interactive board and accepts no options", "use l7 crew wait --json for machine-readable updates")
	}
	client, err := gitadapter.New("", gitadapter.DefaultMaxOutput, gitadapter.DefaultMaxPaths)
	if err != nil {
		return fail(err.Error(), "run l7 help")
	}
	location, err := client.Locate(ctx, cwd)
	if err != nil {
		return fail(err.Error(), "run l7 crew watch from inside the repository")
	}
	if _, err := requireOrchestration(location.Root); err != nil {
		return fail(err.Error(), "run l7 onboard --apply")
	}
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		return fail(err.Error(), "run l7 crew status")
	}
	view, err := watchCrew(ctx, store, stdout, isTerminal(stdout), crewWatchPoll)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fail(err.Error(), "run l7 crew status")
	}
	if crewOverallState(view) == "finished" {
		return writeOrchestration(stdout, stderr, false, passEnvelope("crew watch", "L7-CREW-000", "finished", "every task has finished", crewNext(view), nil))
	}
	return writeOrchestration(stdout, stderr, false, passEnvelope("crew watch", "L7-CREW-000", "stopped", "watch stopped; the crew keeps running", crewNext(view), nil))
}

func watchCrew(ctx context.Context, store crew.Store, stdout io.Writer, clear bool, poll time.Duration) (crewStatusView, error) {
	shown := ""
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		view, err := crewStatus(store)
		if err != nil {
			return view, err
		}
		if key := fmt.Sprintf("%s|%t|%d|%d", view.Token, view.Supervisor, view.SupervisorPID, len(view.OpenDecisions)); key != shown {
			if clear {
				fmt.Fprint(stdout, "\x1b[H\x1b[2J")
			}
			fmt.Fprint(stdout, renderCrewBoard(view))
			shown = key
		}
		if crewOverallState(view) == "finished" {
			return view, nil
		}
		select {
		case <-ctx.Done():
			return view, ctx.Err()
		case <-ticker.C:
		}
	}
}

func renderCrewBoard(view crewStatusView) string {
	var output strings.Builder
	supervisor := "supervisor stopped"
	if view.Supervisor {
		supervisor = fmt.Sprintf("supervisor running (pid %d)", view.SupervisorPID)
	}
	fmt.Fprintf(&output, "Level 7 crew %s -> %s   %s   %s\n\n", view.PlanID, view.TargetBranch, crewOverallState(view), supervisor)
	table := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "TASK\tSHAPE\tSTATE\tROUTE\tCHECK\tLATEST")
	for _, task := range view.Tasks {
		route, check, latest := "-", "-", task.Message
		if task.Provider != "" {
			route = task.Provider + "/" + task.Model
		}
		if task.Verification != "" {
			check = task.Verification
		}
		if task.Checks != "" {
			check = "pr " + task.Checks
		}
		if latest == "" {
			latest = task.Next
		}
		if task.PullRequest != "" && task.State == domain.CrewPROpen {
			latest = task.PullRequest
		}
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\n", strings.TrimPrefix(task.ID, view.PlanID+"-"), task.Shape, task.State, route, check, boardText(latest, 72))
	}
	_ = table.Flush()
	if count := len(view.OpenDecisions); count != 0 {
		fmt.Fprintf(&output, "\n%d open decision(s); first: %s\n", count, boardText(view.OpenDecisions[0].Question, 120))
	}
	fmt.Fprintf(&output, "\nNext: %s\n", boardText(crewNext(view), 200))
	return output.String()
}

func boardText(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	for limit > 0 && (value[limit]&0xC0) == 0x80 {
		limit--
	}
	return value[:limit] + "..."
}

func isTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// crewView builds one detached tmux session per plan: the board first, then
// a shell in each unfinished task's worktree.
func crewView(ctx context.Context, location domain.RepositoryLocation, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) != 0 {
		return orchestrationEnvelope{}, errors.New("crew view accepts no options")
	}
	view, err := crewStatus(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if _, err := processadapter.Resolve("tmux"); err != nil {
		return passEnvelope("crew view", "L7-CREW-000", "unavailable", "tmux is not installed; nothing was changed", "run l7 crew watch for the live board", nil), nil
	}
	socket := tmuxSocket(os.Getenv("TMUX"))
	session := "l7-" + view.PlanID
	if runTmux(ctx, socket, location.Root, "has-session", "-t", "="+session) != nil {
		executable, err := os.Executable()
		if err == nil {
			executable, err = filepath.EvalSymlinks(executable)
		}
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		if err := runTmux(ctx, socket, location.Root, "new-session", "-d", "-s", session, "-n", "board", "-c", location.Root, shellQuote(executable)+" crew watch"); err != nil {
			return orchestrationEnvelope{}, fmt.Errorf("create tmux session %s: %w", session, err)
		}
		for _, task := range view.Tasks {
			if task.State.Terminal() || task.Worktree == "" {
				continue
			}
			if err := runTmux(ctx, socket, location.Root, "new-window", "-d", "-t", session+":", "-n", crewWindowName(view.PlanID, task.ID), "-c", task.Worktree); err != nil {
				return orchestrationEnvelope{}, fmt.Errorf("open tmux window for %s: %w", task.ID, err)
			}
		}
	}
	attach := "tmux attach -t " + session
	if socket != "" {
		attach = "tmux switch-client -t " + session
	}
	return passEnvelope("crew view", "L7-CREW-000", "ready", "tmux session "+session+" shows the board and a shell in each unfinished task's worktree", attach, map[string]string{"session": session}), nil
}

func crewWindowName(planID, taskID string) string { return strings.TrimPrefix(taskID, planID+"-") }

func tmuxSocket(value string) string {
	socket, _, _ := strings.Cut(value, ",")
	if !filepath.IsAbs(socket) {
		return ""
	}
	return socket
}

func runTmux(ctx context.Context, socket, directory string, arguments ...string) error {
	tmux, err := processadapter.Resolve("tmux")
	if err != nil {
		return err
	}
	if socket != "" {
		arguments = append([]string{"-S", socket}, arguments...)
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: tmux.Path, Arguments: arguments, Directory: directory,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 1 << 20, Timeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("tmux exited %d", result.ExitCode)
	}
	return nil
}
