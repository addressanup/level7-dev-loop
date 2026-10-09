package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestCrewPathPatternsAcceptOnlyExactAndRecursiveForms(t *testing.T) {
	for _, pattern := range []string{"README.md", "internal/widget/**", "cmd/l7/main.go", "docs/**"} {
		if !CrewPathPatternValid(pattern) {
			t.Fatalf("valid pattern rejected: %q", pattern)
		}
	}
	for _, pattern := range []string{"", "**", "/etc/passwd", "../x", "a/../b", "./a", "a//b", "a/*.go", "a/**/b", "a\\b", "a/?", "a/[x]", "dir/", "a/**/**"} {
		if CrewPathPatternValid(pattern) {
			t.Fatalf("unsafe or unsupported pattern accepted: %q", pattern)
		}
	}
}

func TestCrewPathScopesMayOverlap(t *testing.T) {
	cases := []struct {
		left, right []string
		overlap     bool
	}{
		{[]string{"a.go"}, []string{"a.go"}, true},
		{[]string{"a.go"}, []string{"b.go"}, false},
		{[]string{"internal/widget/**"}, []string{"internal/widget/file.go"}, true},
		{[]string{"internal/widget/file.go"}, []string{"internal/widget/**"}, true},
		{[]string{"internal/widget/**"}, []string{"internal/widgets/file.go"}, false},
		{[]string{"internal/widget/**"}, []string{"internal/widgets/**"}, false},
		{[]string{"internal/**"}, []string{"internal/widget/**"}, true},
		{[]string{"internal/widget/**"}, []string{"internal/**"}, true},
		{[]string{"internal/widget/**"}, []string{"internal/widget"}, false},
		{[]string{"docs/**", "cmd/a.go"}, []string{"pkg/**", "cmd/a.go"}, true},
		{nil, []string{"anything/**"}, false},
	}
	for _, test := range cases {
		if got := CrewPathScopesMayOverlap(test.left, test.right); got != test.overlap {
			t.Fatalf("overlap(%v, %v) = %v, want %v", test.left, test.right, got, test.overlap)
		}
	}
}

func TestNextCrewAdmissionsRespectsWorkersAndScopes(t *testing.T) {
	tasks := []CrewTask{
		{ID: "t1", Shape: CrewShip, AllowedPaths: []string{"api/**"}},
		{ID: "t2", Shape: CrewShip, AllowedPaths: []string{"api/handler.go"}},
		{ID: "t3", Shape: CrewScout},
		{ID: "t4", Shape: CrewShip, AllowedPaths: []string{"web/**"}},
		{ID: "t5", Shape: CrewShip, AllowedPaths: []string{"docs/**"}},
	}
	queued := map[string]CrewState{"t1": CrewQueued, "t2": CrewQueued, "t3": CrewQueued, "t4": CrewQueued, "t5": CrewQueued}
	if got := NextCrewAdmissions(tasks, queued, 3); !reflect.DeepEqual(got, []string{"t1", "t3", "t4"}) {
		t.Fatalf("admissions = %v; overlapping t2 must wait and plan order must hold", got)
	}
	running := map[string]CrewState{"t1": CrewWaitingQuota, "t2": CrewQueued, "t3": CrewMerging, "t4": CrewQueued, "t5": CrewQueued}
	if got := NextCrewAdmissions(tasks, running, 3); !reflect.DeepEqual(got, []string{"t4"}) {
		t.Fatalf("admissions = %v; waiting and merging tasks hold slots and scope", got)
	}
	settled := map[string]CrewState{"t1": CrewPaused, "t2": CrewQueued, "t3": CrewDone, "t4": CrewNeedsDecision, "t5": CrewCancelled}
	if got := NextCrewAdmissions(tasks, settled, 1); !reflect.DeepEqual(got, []string{"t2"}) {
		t.Fatalf("admissions = %v; paused, decided, and terminal tasks free their scope", got)
	}
	if got := NextCrewAdmissions(tasks, queued, 0); len(got) != 0 {
		t.Fatalf("admissions with no workers = %v", got)
	}
}

func TestCrewStateClassification(t *testing.T) {
	for _, state := range []CrewState{CrewRunning, CrewMerging, CrewWaitingQuota} {
		if !state.Active() || state.Terminal() || state.NeedsAttention() {
			t.Fatalf("%s should be active only", state)
		}
	}
	for _, state := range []CrewState{CrewNeedsDecision, CrewPaused} {
		if state.Active() || state.Terminal() || !state.NeedsAttention() {
			t.Fatalf("%s should need attention without being terminal", state)
		}
	}
	if !CrewDone.Terminal() || !CrewDone.NeedsAttention() || !CrewCancelled.Terminal() || CrewCancelled.NeedsAttention() {
		t.Fatal("terminal classification is wrong")
	}
	if CrewState("unknown").Valid() || CrewQueued.Active() {
		t.Fatal("unknown state accepted or queued task counted as active")
	}
}

func TestCrewDecisionOptionsAreBounded(t *testing.T) {
	options, impact, ok := CrewDecisionFor(CrewDecisionConflict)
	if !ok || impact != 2 || !reflect.DeepEqual(options, []string{CrewAnswerRetry, CrewAnswerRestart, CrewAnswerCancel}) {
		t.Fatalf("conflict decision = %v %d %v", options, impact, ok)
	}
	if options, impact, ok := CrewDecisionFor(CrewDecisionTierThree); !ok || impact != 3 || !reflect.DeepEqual(options, []string{CrewAnswerCancel}) {
		t.Fatal("Tier 3 work is outside the crew: highest impact, cancel only")
	}
	if options, _, _ := CrewDecisionFor(CrewDecisionScopeExpanded); options[0] != CrewAnswerRestart {
		t.Fatal("a scope expansion must offer a fresh restart first")
	}
	if _, _, ok := CrewDecisionFor("push"); ok {
		t.Fatal("unknown decision kind accepted")
	}
}

func TestCrewPlanProblemFailsClosed(t *testing.T) {
	valid := validCrewPlan()
	if problem := CrewPlanProblem(valid); problem != "" {
		t.Fatalf("valid plan rejected: %s", problem)
	}
	mutations := map[string]func(*CrewPlan){
		"schema":           func(plan *CrewPlan) { plan.Schema = 2 },
		"plan id":          func(plan *CrewPlan) { plan.ID = "crew-XYZ" },
		"digest":           func(plan *CrewPlan) { plan.ObjectiveDigest = "md5:abc" },
		"base":             func(plan *CrewPlan) { plan.BaseCommit = "HEAD" },
		"branch":           func(plan *CrewPlan) { plan.TargetBranch = "refs/heads/main" },
		"tier":             func(plan *CrewPlan) { plan.RiskCeiling = TierHighRisk },
		"workers":          func(plan *CrewPlan) { plan.MaxWorkers = 5 },
		"repair":           func(plan *CrewPlan) { plan.RepairRounds = -1 },
		"remote delivery":  func(plan *CrewPlan) { plan.LocalOnly = false },
		"no tasks":         func(plan *CrewPlan) { plan.Tasks = nil },
		"duplicate task":   func(plan *CrewPlan) { plan.Tasks[1].ID = plan.Tasks[0].ID },
		"foreign task id":  func(plan *CrewPlan) { plan.Tasks[0].ID = "crew-000000000000-t01" },
		"shape":            func(plan *CrewPlan) { plan.Tasks[0].Shape = "deploy" },
		"title":            func(plan *CrewPlan) { plan.Tasks[0].Title = " " },
		"acceptance":       func(plan *CrewPlan) { plan.Tasks[0].AcceptanceCriteria = nil },
		"ship paths":       func(plan *CrewPlan) { plan.Tasks[0].AllowedPaths = nil },
		"glob path":        func(plan *CrewPlan) { plan.Tasks[0].AllowedPaths = []string{"api/*.go"} },
		"ship verify":      func(plan *CrewPlan) { plan.Tasks[0].Verification = nil },
		"newline argument": func(plan *CrewPlan) { plan.Tasks[0].Verification = [][]string{{"go", "test\n./..."}} },
		"scout writes":     func(plan *CrewPlan) { plan.Tasks[1].AllowedPaths = []string{"docs/**"} },
		"scout verifies":   func(plan *CrewPlan) { plan.Tasks[1].Verification = [][]string{{"go", "test"}} },
	}
	for name, mutate := range mutations {
		plan := validCrewPlan()
		mutate(&plan)
		if CrewPlanProblem(plan) == "" {
			t.Fatalf("%s mutation was accepted", name)
		}
	}
}

func TestCrewBranchValidation(t *testing.T) {
	for _, branch := range []string{"l7/crew", "integration", "team/feature-1"} {
		if !CrewBranchValid(branch) {
			t.Fatalf("valid branch rejected: %q", branch)
		}
	}
	for _, branch := range []string{"", "-x", "a..b", "a b", "x.lock", "refs/heads/x", "a//b", "/a", "a/", "a@{1}", "a.", strings.Repeat("a", 256)} {
		if CrewBranchValid(branch) {
			t.Fatalf("unsafe branch accepted: %q", branch)
		}
	}
}

func validCrewPlan() CrewPlan {
	return CrewPlan{
		Schema: CrewSchema, ID: "crew-0123456789ab", ObjectivePath: "crew.md",
		ObjectiveDigest: "sha256:" + strings.Repeat("a", 64), BaseCommit: strings.Repeat("b", 40),
		TargetBranch: "l7/crew", RiskCeiling: TierProduct, MaxWorkers: 3, RepairRounds: 2, LocalOnly: true,
		Tasks: []CrewTask{
			{ID: "crew-0123456789ab-t01", Shape: CrewShip, Title: "Fix login", Objective: "Fix the flaky login test.", AcceptanceCriteria: []string{"login test passes 20 times"}, AllowedPaths: []string{"auth/**"}, Verification: [][]string{{"go", "test", "./auth/..."}}},
			{ID: "crew-0123456789ab-t02", Shape: CrewScout, Title: "Investigate", Objective: "Find why CI is slow.", AcceptanceCriteria: []string{"report names the slowest jobs"}},
		},
		Digest: "sha256:" + strings.Repeat("c", 64),
	}
}
