package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestResumeCommandMatchesProviderKind(t *testing.T) {
	configuration := orchestrationconfig.Default()
	worktree := "/repo/.git/l7/crew/worktrees/crew-0123456789ab-t01-a0"
	cases := []struct {
		checkpoint domain.CrewCheckpoint
		want       string
		resumable  bool
	}{
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "codex-local", SessionID: "019a-thread"}, "cd '" + worktree + "' && codex resume 019a-thread", true},
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "claude-local", SessionID: "4f7c2e"}, "cd '" + worktree + "' && claude --resume 4f7c2e", true},
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "claude-local"}, "cd '" + worktree + "' && claude --continue", true},
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "codex-local"}, "cd '" + worktree + "'  # no resumable provider session; edit the files directly", false},
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "codex-local", SessionID: "x; rm -rf ~"}, "cd '" + worktree + "'  # no resumable provider session; edit the files directly", false},
		{domain.CrewCheckpoint{Worktree: worktree, ProviderID: "openai-gateway", SessionID: "abc"}, "cd '" + worktree + "'  # no resumable provider session; edit the files directly", false},
		{domain.CrewCheckpoint{ProviderID: "codex-local", SessionID: "abc"}, "no worktree exists yet; release the task to let a worker start it", false},
	}
	for _, test := range cases {
		command, resumable := resumeCommand(configuration, test.checkpoint)
		if command != test.want || resumable != test.resumable {
			t.Fatalf("checkpoint=%+v command=%q resumable=%t", test.checkpoint, command, resumable)
		}
	}
	if quoted := shellQuote("/tmp/it's here"); quoted != `'/tmp/it'\''s here'` {
		t.Fatalf("quoted = %s", quoted)
	}
}

func TestRenderCrewBoardShowsTasksDecisionsAndNext(t *testing.T) {
	view := crewStatusView{
		PlanID: "crew-0123456789ab", TargetBranch: "l7/crew", Supervisor: true, SupervisorPID: 4242, Token: "3.1",
		Tasks: []crewTaskView{
			{ID: "crew-0123456789ab-t01", Shape: domain.CrewShip, State: domain.CrewAttached, Provider: "codex-local", Model: "gpt-6-astra", Message: "held by the owner"},
			{ID: "crew-0123456789ab-t02", Shape: domain.CrewScout, State: domain.CrewDone, Verification: "", Message: "scout report\nanswered"},
		},
		OpenDecisions: []domain.CrewDecision{{ID: "crew-0123456789ab-t03-d01", Question: "rebase conflict in api/handler.go", Options: []string{"retry", "restart", "cancel"}}},
	}
	board := renderCrewBoard(view)
	for _, want := range []string{"crew-0123456789ab -> l7/crew", "supervisor running (pid 4242)", "t01", "attached", "codex-local/gpt-6-astra", "scout report answered", "1 open decision(s)", "Next: ask the owner"} {
		if !strings.Contains(board, want) {
			t.Fatalf("board lacks %q:\n%s", want, board)
		}
	}
	if long := boardText(strings.Repeat("é", 100), 51); len(long) > 54 || !strings.HasSuffix(long, "...") {
		t.Fatalf("board text is not bounded: %q", long)
	}
	held := crewStatusView{PlanID: view.PlanID, Tasks: []crewTaskView{view.Tasks[0], {ID: "crew-0123456789ab-t02", State: domain.CrewQueued}}}
	if crewOverallState(held) != "held" || crewNext(held) != "the owner holds crew-0123456789ab-t01; when done, run l7 crew release --task crew-0123456789ab-t01" {
		t.Fatalf("a held crew must point at release, not resume: %s / %s", crewOverallState(held), crewNext(held))
	}
}

func approvedCLICrew(t *testing.T) (domain.RepositoryLocation, crew.Store, domain.CrewPlan) {
	t.Helper()
	location, configuration := crewLocation(t, true)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	planned, err := crewPlan(context.Background(), location, configuration, store, []string{"--objective", "crew.md"})
	if err != nil {
		t.Fatal(err)
	}
	plan := planned.Data.(domain.CrewPlan)
	if _, err := store.Approve(plan, plan.Digest, "Anup", "owner", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.Activate(plan); err != nil {
		t.Fatal(err)
	}
	return location, store, plan
}

func TestCrewAttachAndReleaseThroughTheCLI(t *testing.T) {
	location, store, plan := approvedCLICrew(t)
	launches := 0
	original := launchCrewSupervisor
	launchCrewSupervisor = func(crew.Store, string) (int, error) { launches++; return 4242, nil }
	t.Cleanup(func() { launchCrewSupervisor = original })
	t.Setenv("TMUX", "")
	ship := plan.Tasks[0].ID
	attached, err := crewCommand(context.Background(), location, []string{"attach", "--task", ship})
	if err != nil || attached.State != "attached" || !strings.Contains(attached.Next, "release --task "+ship) {
		t.Fatalf("attach = %+v err=%v", attached, err)
	}
	data := attached.Data.(map[string]any)
	if data["resumable"] != false || data["opened_in_tmux"] != false {
		t.Fatalf("a task without a worktree must not offer a resume command: %+v", data)
	}
	view, err := crewStatus(store)
	if err != nil || view.Tasks[0].State != domain.CrewAttached {
		t.Fatalf("status after attach = %+v err=%v", view, err)
	}
	released, err := crewCommand(context.Background(), location, []string{"release", "--task", ship})
	if err != nil || released.State != string(domain.CrewQueued) || launches != 1 || !strings.Contains(released.Message, "supervisor 4242 started") {
		t.Fatalf("release = %+v launches=%d err=%v", released, launches, err)
	}
	if _, err := crewCommand(context.Background(), location, []string{"release", "--task", ship}); err == nil {
		t.Fatal("a queued task was released")
	}
	launchCrewSupervisor = func(crew.Store, string) (int, error) { return 0, crew.ErrSupervisorRunning }
	if _, err := crewCommand(context.Background(), location, []string{"attach", "--task", ship}); err != nil {
		t.Fatal(err)
	}
	if released, err := crewCommand(context.Background(), location, []string{"release", "--task", ship}); err != nil || strings.Contains(released.Message, "started") {
		t.Fatalf("a running supervisor must pick the task up: %+v err=%v", released, err)
	}
	for _, arguments := range [][]string{{"attach"}, {"attach", "--task"}, {"attach", "--task", ship, "--force"}, {"release", "--id", ship}, {"attach", "--task", "crew-ffffffffffff-t01"}} {
		if _, err := crewCommand(context.Background(), location, arguments); err == nil {
			t.Fatalf("crew accepted %v", arguments)
		}
	}
}

func TestCrewAttachAndReleaseFailClosedWhileFeatureIsOff(t *testing.T) {
	location, _ := crewLocation(t, false)
	for _, action := range []string{"attach", "release"} {
		if _, err := crewCommand(context.Background(), location, []string{action, "--task", "crew-0123456789ab-t01"}); err == nil || !strings.Contains(err.Error(), "default OFF") {
			t.Fatalf("%s ran while crew was OFF: %v", action, err)
		}
	}
}

func TestWatchCrewDrawsOnChangeAndStopsWhenFinishedOrInterrupted(t *testing.T) {
	_, store, plan := approvedCLICrew(t)
	var output bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	view, err := watchCrew(ctx, store, &output, false, 5*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || crewOverallState(view) == "finished" || strings.Count(output.String(), "Level 7 crew") != 1 {
		t.Fatalf("an unchanged crew must draw once until interrupted: err=%v frames=%d", err, strings.Count(output.String(), "Level 7 crew"))
	}
	for _, task := range plan.Tasks {
		if _, err := store.Update(plan, task.ID, time.Now(), func(checkpoint *domain.CrewCheckpoint) error {
			checkpoint.State, checkpoint.Next = domain.CrewDone, "done"
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	output.Reset()
	view, err = watchCrew(context.Background(), store, &output, true, 5*time.Millisecond)
	if err != nil || crewOverallState(view) != "finished" || !strings.HasPrefix(output.String(), "\x1b[H\x1b[2J") || !strings.Contains(output.String(), "finished") {
		t.Fatalf("a finished crew must draw once and return: err=%v output=%q", err, output.String())
	}
}

func TestTmuxSocketComesFromTheTMUXVariable(t *testing.T) {
	if socket := tmuxSocket("/private/tmp/tmux-501/default,1234,0"); socket != "/private/tmp/tmux-501/default" {
		t.Fatalf("socket = %q", socket)
	}
	if socket := tmuxSocket("relative,1,0"); socket != "" {
		t.Fatalf("relative socket accepted: %q", socket)
	}
	if name := crewWindowName("crew-0123456789ab", "crew-0123456789ab-t02"); name != "t02" {
		t.Fatalf("window name = %q", name)
	}
}
