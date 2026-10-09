package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// pullRequestLocation is a real repository on branch main with an origin.
func pullRequestLocation(t *testing.T) (domain.RepositoryLocation, orchestrationconfig.File) {
	t.Helper()
	location, configuration := crewLocation(t, true)
	if err := os.RemoveAll(location.CommonDir); err != nil {
		t.Fatal(err)
	}
	cliGit(t, location.Root, "init", "-q", "-b", "main")
	cliGit(t, location.Root, "add", "crew.md")
	cliGit(t, location.Root, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "initial")
	cliGit(t, location.Root, "remote", "add", "origin", filepath.Join(t.TempDir(), "remote.git"))
	location.Head = strings.TrimSpace(cliGit(t, location.Root, "rev-parse", "HEAD"))
	configuration.Features.CrewPR = true
	return location, configuration
}

func TestCrewPullRequestPlanIsGatedAndTargetsTheCheckedOutBranch(t *testing.T) {
	location, configuration := pullRequestLocation(t)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	off := configuration
	off.Features.CrewPR = false
	if _, err := crewPlan(context.Background(), location, off, store, []string{"--objective", "crew.md", "--deliver", "pr"}); err == nil || !strings.Contains(err.Error(), "default OFF") {
		t.Fatalf("pull-request planning ran while crew_pr was OFF: %v", err)
	}
	planned, err := crewPlan(context.Background(), location, configuration, store, []string{"--objective", "crew.md", "--deliver", "pr"})
	if err != nil {
		t.Fatal(err)
	}
	plan := planned.Data.(domain.CrewPlan)
	if plan.Delivery != domain.CrewDeliveryPR || plan.Remote != "origin" || plan.TargetBranch != "main" || plan.LocalOnly || !strings.Contains(planned.Message, "pull request into main") || !strings.Contains(planned.Message, "Nothing merges until you run l7 crew merge") {
		t.Fatalf("pull-request plan = %+v message=%q", plan, planned.Message)
	}
	for _, arguments := range [][]string{
		{"--objective", "crew.md", "--deliver", "pr", "--target", "release"},
		{"--objective", "crew.md", "--deliver", "github"},
	} {
		if _, err := crewPlan(context.Background(), location, configuration, store, arguments); err == nil {
			t.Fatalf("crew plan accepted %v", arguments)
		}
	}
	missing := configuration
	missing.Crew.Remote = "upstream"
	if _, err := crewPlan(context.Background(), location, missing, store, []string{"--objective", "crew.md", "--deliver", "pr"}); err == nil || !strings.Contains(err.Error(), "upstream") {
		t.Fatalf("a plan for a missing remote was accepted: %v", err)
	}
	cliGit(t, location.Root, "checkout", "-q", "--detach")
	if _, err := crewPlan(context.Background(), location, configuration, store, []string{"--objective", "crew.md", "--deliver", "pr"}); err == nil || !strings.Contains(err.Error(), "detached") {
		t.Fatalf("a detached HEAD was planned for pull requests: %v", err)
	}
}

func TestCrewPullRequestStartRefusesWithoutGh(t *testing.T) {
	location, configuration := pullRequestLocation(t)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	planned, err := crewPlan(context.Background(), location, configuration, store, []string{"--objective", "crew.md", "--deliver", "pr"})
	if err != nil {
		t.Fatal(err)
	}
	plan := planned.Data.(domain.CrewPlan)
	original := discoverForge
	discoverForge = func(string) (forge.Client, error) { return forge.Client{}, errors.New("the gh CLI is not installed") }
	t.Cleanup(func() { discoverForge = original })
	launched := false
	originalLaunch := launchCrewSupervisor
	launchCrewSupervisor = func(crew.Store, string) (int, error) { launched = true; return 1, nil }
	t.Cleanup(func() { launchCrewSupervisor = originalLaunch })
	arguments := []string{"--plan", plan.ID, "--digest", plan.Digest, "--owner", "Anup", "--role", "owner", "--confirm"}
	if _, err := crewStart(context.Background(), location, configuration, store, arguments); err == nil || !strings.Contains(err.Error(), "gh") {
		t.Fatalf("a pull-request crew started without gh: %v", err)
	}
	if _, err := store.LoadApproval(plan); err == nil || launched {
		t.Fatal("a refused start still recorded approval or launched the supervisor")
	}
	off := configuration
	off.Features.CrewPR = false
	if _, err := crewStart(context.Background(), location, off, store, arguments); err == nil || !strings.Contains(err.Error(), "default OFF") {
		t.Fatalf("a pull-request crew started while crew_pr was OFF: %v", err)
	}
}

type fakeMerger struct {
	head string
	err  error
}

func (merger *fakeMerger) MergePullRequest(_ context.Context, _ domain.CrewPlan, checkpoint domain.CrewCheckpoint, head string) (crew.Outcome, error) {
	merger.head = head
	if merger.err != nil {
		return crew.Outcome{}, merger.err
	}
	return crew.Outcome{Kind: crew.OutcomeMerged, Message: "merged pull request #3"}, nil
}

func TestCrewMergeRecordsOnlyAConfirmedExactHeadMerge(t *testing.T) {
	location, store, plan := approvedCLICrew(t)
	configuration := orchestrationconfig.AppliedDefault()
	configuration.Features.Crew, configuration.Features.CrewPR = true, true
	head := strings.Repeat("c", 40)
	ship := plan.Tasks[0].ID
	if _, err := store.Update(plan, ship, time.Now(), func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State, checkpoint.PullRequest, checkpoint.CandidateCommit, checkpoint.Next = domain.CrewPROpen, 3, head, "review pull request #3"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	merger := &fakeMerger{err: errors.New("pull request #3 checks are not all passing: pending: 1 of 2")}
	original := newCrewMerger
	newCrewMerger = func(domain.RepositoryLocation, orchestrationconfig.File) (pullRequestMerger, error) {
		return merger, nil
	}
	t.Cleanup(func() { newCrewMerger = original })
	for _, arguments := range [][]string{
		{"--task", ship, "--head", head},
		{"--task", ship, "--head", "abc123", "--confirm"},
		{"--head", head, "--confirm"},
		{"--task", ship, "--head", head, "--confirm", "--admin"},
	} {
		if _, err := crewMerge(context.Background(), location, configuration, store, arguments); err == nil {
			t.Fatalf("crew merge accepted %v", arguments)
		}
	}
	confirmed := []string{"--task", ship, "--head", head, "--confirm"}
	off := configuration
	off.Features.CrewPR = false
	if _, err := crewMerge(context.Background(), location, off, store, confirmed); err == nil || !strings.Contains(err.Error(), "default OFF") {
		t.Fatalf("merge ran while crew_pr was OFF: %v", err)
	}
	if _, err := crewMerge(context.Background(), location, configuration, store, confirmed); err == nil || !strings.Contains(err.Error(), "pending") {
		t.Fatalf("a refused merge was reported as success: %v", err)
	}
	if checkpoint, _ := store.Checkpoint(plan, ship); checkpoint.State != domain.CrewPROpen {
		t.Fatalf("a refused merge changed the task: %+v", checkpoint)
	}
	merger.err = nil
	merged, err := crewMerge(context.Background(), location, configuration, store, confirmed)
	if err != nil || merged.State != "merged" || merger.head != head {
		t.Fatalf("merge = %+v head=%s err=%v", merged, merger.head, err)
	}
	if checkpoint, _ := store.Checkpoint(plan, ship); checkpoint.State != domain.CrewDone || checkpoint.Message != "merged pull request #3" {
		t.Fatalf("merged task = %+v", checkpoint)
	}
	view := crewStatusView{PlanID: plan.ID, Supervisor: true, Tasks: []crewTaskView{{ID: ship, Shape: domain.CrewShip, State: domain.CrewPROpen, PullRequest: "https://github.com/o/r/pull/3", Checks: "passed: 2", Next: "review pull request #3, then run l7 crew merge"}}}
	if board := renderCrewBoard(view); !strings.Contains(board, "pr passed: 2") || !strings.Contains(board, "https://github.com/o/r/pull/3") || crewNext(view) != "review pull request #3, then run l7 crew merge" {
		t.Fatalf("board does not show the pull request:\n%s\nnext=%s", board, crewNext(view))
	}
}
