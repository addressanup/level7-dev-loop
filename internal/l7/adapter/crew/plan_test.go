package crew

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const sampleObjective = `# Sprint

This is a Go service; keep handlers small.

## Ship: Fix the flaky login test
The login test fails one run in ten.
Paths: auth/**, auth_test.go
Verify: ["go","test","./auth/..."]
Acceptance: the login test passes 20 runs in a row
- [ ] no other test regresses

## Scout: Why is CI slow?
Find the slowest CI jobs and explain why.
Success: the report names the three slowest jobs with evidence
`

func fixedPlanner() Planner {
	return NewPlannerWith(func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) })
}

func sampleRequest() PlanRequest {
	return PlanRequest{
		ObjectivePath: "crew.md", Objective: []byte(sampleObjective), BaseCommit: strings.Repeat("a", 40),
		TargetBranch: "l7/crew", MaxWorkers: 3, RepairRounds: 2,
	}
}

func TestPlanParsesShipAndScoutTasks(t *testing.T) {
	plan, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("planned crew is invalid: %v", err)
	}
	if len(plan.Tasks) != 2 || !plan.LocalOnly || plan.RiskCeiling != domain.TierProduct {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	ship, scout := plan.Tasks[0], plan.Tasks[1]
	if ship.ID != plan.ID+"-t01" || ship.Shape != domain.CrewShip || ship.Title != "Fix the flaky login test" {
		t.Fatalf("ship identity = %+v", ship)
	}
	if !reflect.DeepEqual(ship.AllowedPaths, []string{"auth/**", "auth_test.go"}) || !reflect.DeepEqual(ship.Verification, [][]string{{"go", "test", "./auth/..."}}) {
		t.Fatalf("ship scope = %v verification = %v", ship.AllowedPaths, ship.Verification)
	}
	if !reflect.DeepEqual(ship.AcceptanceCriteria, []string{"the login test passes 20 runs in a row", "no other test regresses"}) {
		t.Fatalf("ship acceptance = %v", ship.AcceptanceCriteria)
	}
	if !strings.HasPrefix(ship.Objective, "Context:\n# Sprint") || strings.Contains(ship.Objective, "Paths:") || strings.Contains(ship.Objective, "Verify:") {
		t.Fatalf("ship objective must carry shared context without scope lines: %q", ship.Objective)
	}
	if scout.Shape != domain.CrewScout || len(scout.AllowedPaths) != 0 || len(scout.Verification) != 0 || len(scout.AcceptanceCriteria) != 1 {
		t.Fatalf("scout = %+v", scout)
	}
	again, err := fixedPlanner().Plan(sampleRequest())
	if err != nil || again.Digest != plan.Digest || again.ID != plan.ID {
		t.Fatalf("planning is not deterministic: %s vs %s (%v)", again.Digest, plan.Digest, err)
	}
}

func TestPlanIdentityBindsBaseAndTarget(t *testing.T) {
	first, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	moved := sampleRequest()
	moved.BaseCommit = strings.Repeat("b", 40)
	second, err := fixedPlanner().Plan(moved)
	if err != nil || second.ID == first.ID {
		t.Fatalf("a new base must create a new plan identity: %s %s %v", first.ID, second.ID, err)
	}
}

func TestPullRequestPlanBindsDeliveryAndRemote(t *testing.T) {
	local, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	request := sampleRequest()
	request.TargetBranch, request.Delivery, request.Remote = "main", domain.CrewDeliveryPR, "origin"
	pullRequests, err := fixedPlanner().Plan(request)
	if err != nil || pullRequests.Delivery != domain.CrewDeliveryPR || pullRequests.Remote != "origin" || pullRequests.LocalOnly || ValidatePlan(pullRequests) != nil {
		t.Fatalf("pull-request plan = %+v err=%v", pullRequests, err)
	}
	request.Remote = "upstream"
	other, err := fixedPlanner().Plan(request)
	if err != nil || other.ID == pullRequests.ID || other.Digest == pullRequests.Digest {
		t.Fatalf("a different remote must change the plan identity: %s %s", other.ID, pullRequests.ID)
	}
	request.Delivery, request.Remote = domain.CrewDeliveryLocal, ""
	sameBranchLocal, err := fixedPlanner().Plan(request)
	if err != nil || sameBranchLocal.ID == pullRequests.ID {
		t.Fatalf("delivery mode must change the plan identity: %s %s err=%v", sameBranchLocal.ID, pullRequests.ID, err)
	}
	if local.Delivery != "" || local.Remote != "" || !local.LocalOnly {
		t.Fatalf("local plans keep their Phase 1 shape: %+v", local)
	}
	tampered := pullRequests
	tampered.Remote = "upstream"
	if ValidatePlan(tampered) == nil {
		t.Fatal("a changed remote kept a valid digest")
	}
	for _, bad := range []PlanRequest{{Delivery: "deploy"}, {Delivery: domain.CrewDeliveryLocal, Remote: "origin"}, {Delivery: domain.CrewDeliveryPR, Remote: "-x"}} {
		invalid := sampleRequest()
		invalid.Delivery, invalid.Remote = bad.Delivery, bad.Remote
		if _, err := fixedPlanner().Plan(invalid); err == nil {
			t.Fatalf("delivery %+v accepted", bad)
		}
	}
}

func TestPlanUsesDefaultVerificationForShipTasks(t *testing.T) {
	request := sampleRequest()
	request.Objective = []byte("## Ship: Add docs\nPaths: docs/**\nAcceptance: docs build\n")
	request.DefaultVerification = [][]string{{"make", "docs"}}
	plan, err := fixedPlanner().Plan(request)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Tasks[0].Verification, [][]string{{"make", "docs"}}) {
		t.Fatalf("default verification not applied: %v", plan.Tasks[0].Verification)
	}
	request.DefaultVerification[0][0] = "rm"
	if plan.Tasks[0].Verification[0][0] != "make" {
		t.Fatal("plan aliases the caller's verification slice")
	}
}

func TestPlanRejectsUnsafeOrIncompleteObjectives(t *testing.T) {
	var many strings.Builder
	for index := 0; index <= domain.CrewMaxTasks; index++ {
		fmt.Fprintf(&many, "## Scout: task %d\nAcceptance: done\n", index)
	}
	cases := map[string]func(*PlanRequest){
		"no sections":      func(request *PlanRequest) { request.Objective = []byte("just prose\n") },
		"scout with paths": func(request *PlanRequest) { request.Objective = []byte("## Scout: x\nPaths: a/**\nAcceptance: y\n") },
		"scout verifies": func(request *PlanRequest) {
			request.Objective = []byte("## Scout: x\nVerify: [\"go\"]\nAcceptance: y\n")
		},
		"ship without path": func(request *PlanRequest) {
			request.Objective = []byte("## Ship: x\nVerify: [\"go\"]\nAcceptance: y\n")
		},
		"ship no verify": func(request *PlanRequest) { request.Objective = []byte("## Ship: x\nPaths: a/**\nAcceptance: y\n") },
		"bad verify json": func(request *PlanRequest) {
			request.Objective = []byte("## Ship: x\nPaths: a/**\nVerify: go test\nAcceptance: y\n")
		},
		"glob path": func(request *PlanRequest) {
			request.Objective = []byte("## Ship: x\nPaths: a/*.go\nVerify: [\"go\"]\nAcceptance: y\n")
		},
		"no acceptance":    func(request *PlanRequest) { request.Objective = []byte("## Scout: x\nlook around\n") },
		"too many tasks":   func(request *PlanRequest) { request.Objective = []byte(many.String()) },
		"unsafe objective": func(request *PlanRequest) { request.ObjectivePath = "../outside.md" },
		"absolute path":    func(request *PlanRequest) { request.ObjectivePath = "/etc/crew.md" },
		"invalid utf8":     func(request *PlanRequest) { request.Objective = []byte{0xff, 0xfe} },
		"short base":       func(request *PlanRequest) { request.BaseCommit = "abc" },
		"checked branch":   func(request *PlanRequest) { request.TargetBranch = "refs/heads/main" },
		"too many workers": func(request *PlanRequest) { request.MaxWorkers = 9 },
	}
	for name, mutate := range cases {
		request := sampleRequest()
		mutate(&request)
		if _, err := fixedPlanner().Plan(request); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestValidatePlanDetectsTampering(t *testing.T) {
	plan, err := fixedPlanner().Plan(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	plan.Tasks[0].AllowedPaths = append(plan.Tasks[0].AllowedPaths, "internal/**")
	if err := ValidatePlan(plan); err == nil {
		t.Fatal("widened task scope kept a valid digest")
	}
}
