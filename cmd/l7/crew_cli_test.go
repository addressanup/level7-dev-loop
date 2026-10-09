package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const cliObjective = `## Ship: Fix login
Paths: auth/**
Verify: ["go","test","./auth/..."]
Acceptance: login test passes 20 runs in a row
## Scout: CI time
Acceptance: report names the slowest job
`

func crewLocation(t *testing.T, crewEnabled bool) (domain.RepositoryLocation, orchestrationconfig.File) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	common := filepath.Join(root, ".git")
	if err := os.MkdirAll(filepath.Join(root, ".l7"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	configuration := orchestrationconfig.AppliedDefault()
	configuration.Features.Crew = crewEnabled
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(orchestrationconfig.Path(root), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "crew.md"), []byte(cliObjective), 0o600); err != nil {
		t.Fatal(err)
	}
	return domain.RepositoryLocation{Root: root, CommonDir: common, Head: strings.Repeat("a", 40)}, configuration
}

func TestCrewMutationsFailClosedWhileFeatureIsOff(t *testing.T) {
	location, _ := crewLocation(t, false)
	for _, arguments := range [][]string{{"plan", "--objective", "crew.md"}, {"start", "--plan", "x"}, {"answer", "--decision", "x", "--choice", "retry"}, {"resume"}, {"supervise"}} {
		if _, err := crewCommand(context.Background(), location, arguments); err == nil || !strings.Contains(err.Error(), "default OFF") {
			t.Fatalf("%v ran while crew was OFF: %v", arguments, err)
		}
	}
	envelope, err := crewCommand(context.Background(), location, []string{"status"})
	if err != nil || envelope.State != "idle" {
		t.Fatalf("read-only status must stay available: %+v err=%v", envelope, err)
	}
}

func TestCrewPlanIsIdempotentAndRejectsProtectedScopes(t *testing.T) {
	location, configuration := crewLocation(t, true)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	first, err := crewPlan(location, configuration, store, []string{"--objective", "crew.md"})
	if err != nil || first.State != "planned" || !strings.Contains(first.Message, "Nothing is pushed") {
		t.Fatalf("plan = %+v err=%v", first, err)
	}
	plan := first.Data.(domain.CrewPlan)
	if plan.TargetBranch != defaultCrewTarget || plan.MaxWorkers != 3 || plan.RepairRounds != 2 || !strings.Contains(first.Next, plan.Digest) {
		t.Fatalf("plan defaults = %+v next=%q", plan, first.Next)
	}
	second, err := crewPlan(location, configuration, store, []string{"--objective", filepath.Join(location.Root, "crew.md")})
	if err != nil || second.Data.(domain.CrewPlan).Digest != plan.Digest {
		t.Fatalf("re-planning the same objective must return the same plan: %+v err=%v", second, err)
	}
	if err := os.MkdirAll(store.ObjectiveRoot(), 0o700); err != nil {
		t.Fatal(err)
	}
	private := filepath.Join(store.ObjectiveRoot(), "sprint.md")
	if err := os.WriteFile(private, []byte("Sprint 2\n\n"+cliObjective), 0o600); err != nil {
		t.Fatal(err)
	}
	drafted, err := crewPlan(location, configuration, store, []string{"--objective", private})
	if err != nil || drafted.Data.(domain.CrewPlan).ObjectivePath != "crew-objectives/sprint.md" {
		t.Fatalf("private objective draft rejected: %+v err=%v", drafted, err)
	}
	if _, err := crewPlan(location, configuration, store, []string{"--objective", filepath.Join(t.TempDir(), "elsewhere.md")}); err == nil {
		t.Fatal("objective outside the repository and the private objective directory was accepted")
	}
	protected := strings.Replace(cliObjective, "Paths: auth/**", "Paths: auth/**, .github/**", 1)
	if err := os.WriteFile(filepath.Join(location.Root, "crew.md"), []byte(protected), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := crewPlan(location, configuration, store, []string{"--objective", "crew.md"}); err == nil || !strings.Contains(err.Error(), "protected") {
		t.Fatalf("protected scope accepted: %v", err)
	}
	for _, arguments := range [][]string{{}, {"--objective"}, {"--objective", "crew.md", "--objective", "x.md"}, {"--objective", "crew.md", "--push", "origin"}, {"--objective", "crew.md", "--command-json", "go test"}, {"--objective", "../outside.md"}} {
		if _, err := crewPlan(location, configuration, store, arguments); err == nil {
			t.Fatalf("plan accepted %v", arguments)
		}
	}
}

func TestCrewStatusReportsTasksDecisionsAndNextAction(t *testing.T) {
	location, configuration := crewLocation(t, true)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	planned, err := crewPlan(location, configuration, store, []string{"--objective", "crew.md"})
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
	scout := plan.Tasks[1].ID
	if _, err := store.Update(plan, scout, time.Now(), func(checkpoint *domain.CrewCheckpoint) error {
		checkpoint.State, checkpoint.Next = domain.CrewDone, "read the report"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	view, err := crewStatus(store)
	if err != nil || len(view.Tasks) != 2 || view.Tasks[1].Report == "" || view.Supervisor || view.Token != "0.1" {
		t.Fatalf("view = %+v err=%v", view, err)
	}
	if crewOverallState(view) != "stopped" || crewNext(view) != "run l7 crew resume" {
		t.Fatalf("a stopped crew must point at resume: %s / %s", crewOverallState(view), crewNext(view))
	}
	if _, err := store.OpenDecision(plan, plan.Tasks[0].ID, domain.CrewDecisionConflict, "rebase conflict in auth/login.go", time.Now()); err != nil {
		t.Fatal(err)
	}
	view, err = crewStatus(store)
	if err != nil || crewOverallState(view) != "needs-decision" || !strings.Contains(crewNext(view), "--choice <retry|restart|cancel>") || !strings.Contains(crewNext(view), "auth/login.go") {
		t.Fatalf("open decision must lead the next action: %s err=%v", crewNext(view), err)
	}
	envelope, err := crewCommand(context.Background(), location, []string{"decisions"})
	if err != nil || !strings.Contains(envelope.Next, "crew answer --decision") {
		t.Fatalf("decisions = %+v err=%v", envelope, err)
	}
	finished := crewStatusView{TargetBranch: "l7/crew", Tasks: []crewTaskView{{State: domain.CrewDone}, {State: domain.CrewCancelled}}}
	if crewOverallState(finished) != "finished" || crewNext(finished) != "review the merged work, then fast-forward your branch: git merge --ff-only l7/crew" {
		t.Fatalf("finished next = %q", crewNext(finished))
	}
	running := crewStatusView{Supervisor: true, Token: "3.4", Tasks: []crewTaskView{{State: domain.CrewRunning}}}
	if crewNext(running) != "run l7 crew wait --since 3.4" {
		t.Fatalf("running next = %q", crewNext(running))
	}
}

func TestCrewOptionValidation(t *testing.T) {
	location, _ := crewLocation(t, true)
	store, err := crew.Open(location.CommonDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"--timeout", "-1"}, {"--timeout", "3601"}, {"--timeout"}, {"--since", "1", "--follow", "x"}} {
		if _, err := crewWait(context.Background(), store, arguments); err == nil {
			t.Fatalf("wait accepted %v", arguments)
		}
	}
	for _, arguments := range [][]string{{}, {"--decision", "x"}, {"--decision", "x", "--choice", "retry", "--force", "y"}, {"--decision", "x", "--decision", "y"}} {
		if _, err := crewAnswer(location, store, arguments); err == nil {
			t.Fatalf("answer accepted %v", arguments)
		}
	}
	for _, arguments := range [][]string{{"--plan", "crew-0123456789ab"}, {"--plan", "p", "--digest", "d", "--owner", "o", "--role", "r"}, {"--plan", "p", "--remote", "origin", "--confirm"}} {
		if _, err := crewStart(context.Background(), location, store, arguments); err == nil {
			t.Fatalf("start accepted %v", arguments)
		}
	}
	for _, arguments := range [][]string{{"status", "--all"}, {"decisions", "x"}, {"resume", "now"}, {"cancel", "--force"}, {"deploy"}} {
		if _, err := crewCommand(context.Background(), location, arguments); err == nil {
			t.Fatalf("crew accepted %v", arguments)
		}
	}
}
