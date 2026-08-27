package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTierOneBriefUsesFastPathWithZeroGovernanceArtifacts(t *testing.T) {
	fixture := newLifecycleFixture()
	briefCalls := 0
	fixture.ports.EnsureBrief = func(string, domain.ChangeBrief) (bool, error) {
		briefCalls++
		return false, errors.New("Tier 1 must not create a brief")
	}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{
		Command:  domain.CommandBrief,
		ChangeID: "routine-fix",
		Tier:     domain.TierRoutine,
		Problem:  "Fix a low-risk internal defect.",
		Scope:    []string{"internal/example/**"},
	})
	if result.Outcome != domain.OutcomePass || result.State != string(domain.StatePlanned) || briefCalls != 0 || len(fixture.saved) != 1 || fixture.saved[0].Kind != domain.ActiveTierOne || fixture.saved[0].BriefPath != "" {
		t.Fatalf("Tier 1 result=%+v briefCalls=%d saved=%+v", result, briefCalls, fixture.saved)
	}
}

func TestTierTwoCreatesExactlyOneConciseBriefAndMinimalPointer(t *testing.T) {
	fixture := newLifecycleFixture()
	briefPath := "docs/artifacts/changes/product-feature.md"
	fixture.snapshotPaths = [][]string{{}, {}, {briefPath}}
	var captured domain.ChangeBrief
	fixture.ports.EnsureBrief = func(_ string, brief domain.ChangeBrief) (bool, error) {
		captured = brief
		fixture.brief = brief
		return true, nil
	}
	result := fixture.application().ExecuteRequest(context.Background(), productRequest())
	if result.Outcome != domain.OutcomePass || result.Code != "L7-BRIEF-000" || captured.Path != briefPath || captured.Tier != domain.TierProduct || len(fixture.saved) != 1 {
		t.Fatalf("Tier 2 result=%+v brief=%+v saved=%+v", result, captured, fixture.saved)
	}
	wantPointer := domain.ActiveChange{Kind: domain.ActiveBrief, ID: "product-feature", BriefPath: briefPath}
	got := fixture.saved[0]
	if got.Kind != wantPointer.Kind || got.ID != wantPointer.ID || got.BriefPath != wantPointer.BriefPath || got.Tier != 0 || got.Base != "" || len(got.Scope) != 0 {
		t.Fatalf("active pointer=%+v, want minimal %+v", got, wantPointer)
	}
}

func TestTierTwoBriefRequiresAllArtifactFields(t *testing.T) {
	fixture := newLifecycleFixture()
	fixture.ports.EnsureBrief = func(_ string, brief domain.ChangeBrief) (bool, error) {
		if len(brief.AcceptanceCriteria) == 0 || len(brief.Risks) == 0 || len(brief.Rollback) == 0 {
			return false, errors.New("tracked brief is incomplete")
		}
		return true, nil
	}
	request := productRequest()
	request.Rollback = nil
	result := fixture.application().ExecuteRequest(context.Background(), request)
	if result.Outcome != domain.OutcomeFailed || result.Code != "L7-BRIEF-002" || len(fixture.saved) != 0 {
		t.Fatalf("incomplete Tier 2 result=%+v saved=%+v", result, fixture.saved)
	}
}

func TestTierThreeCannotBuildBeforeExternalOwnerApproval(t *testing.T) {
	fixture := newLifecycleFixture()
	fixture.active = domain.ActiveChange{Kind: domain.ActiveBrief, ID: "security-change", BriefPath: "docs/artifacts/changes/security-change.md"}
	fixture.activeFound = true
	fixture.brief = domain.ChangeBrief{
		ID: "security-change", Tier: domain.TierHighRisk, Base: fixture.location.Head,
		Path: "docs/artifacts/changes/security-change.md", Problem: "Change authorization.",
		Scope: []string{"internal/auth/policy.go"}, AcceptanceCriteria: []string{"Authorization tests pass."},
		Risks: []string{"An authorization bypass is possible."}, Rollback: []string{"Revert the commit."},
	}
	fixture.snapshotPaths = [][]string{{fixture.brief.Path}}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if result.Outcome != domain.OutcomeBlocked || result.State != string(domain.StateAwaitingOwnerApproval) || result.Code != "L7-AUTH-002" {
		t.Fatalf("Tier 3 planned status=%+v", result)
	}
	fixture.snapshotIndex = 0
	fixture.snapshotPaths = [][]string{{fixture.brief.Path, "internal/auth/policy.go"}}
	result = fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-AUTH-001" || result.State != string(domain.StateAwaitingOwnerApproval) {
		t.Fatalf("Tier 3 unauthorized work status=%+v", result)
	}
}

func TestProtectedScopeElevatesBeforeAnyMutation(t *testing.T) {
	for _, scope := range [][]string{{"Makefile"}, {"internal/**"}, {"custom/protected/**"}} {
		fixture := newLifecycleFixture()
		fixture.configuration.ProtectedPaths = []string{"custom/protected/**"}
		request := productRequest()
		request.Scope = scope
		result := fixture.application().ExecuteRequest(context.Background(), request)
		if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-RISK-001" || len(fixture.saved) != 0 || fixture.acquisitions != 0 {
			t.Fatalf("scope=%v result=%+v saved=%+v acquisitions=%d", scope, result, fixture.saved, fixture.acquisitions)
		}
	}
}

func TestStatusFailsClosedOnScopeExpansion(t *testing.T) {
	fixture := plannedTierOneFixture()
	fixture.snapshotPaths = [][]string{{"internal/example/change.go", "outside.txt"}}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-SCOPE-001" || result.Repository == nil || strings.Join(result.Repository.ExpandedPaths, "|") != "outside.txt" {
		t.Fatalf("expanded status=%+v", result)
	}
}

func TestStatusDistinguishesPlannedFromUnsupportedVerification(t *testing.T) {
	fixture := plannedTierOneFixture()
	fixture.snapshotPaths = [][]string{{}}
	planned := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if planned.Outcome != domain.OutcomePass || planned.State != string(domain.StatePlanned) {
		t.Fatalf("planned status=%+v", planned)
	}
	fixture.snapshotIndex = 0
	fixture.snapshotPaths = [][]string{{"internal/example/change.go"}}
	building := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if building.Outcome != domain.OutcomeBlocked || building.State != string(domain.StateBuilding) || building.Code != "L7-CAP-002" || !strings.Contains(building.Message, "Wave 3") {
		t.Fatalf("building status=%+v", building)
	}
}

func TestBriefRecoveryCanCompleteAfterBriefWriteInterruption(t *testing.T) {
	fixture := newLifecycleFixture()
	path := "docs/artifacts/changes/product-feature.md"
	fixture.snapshotPaths = [][]string{{path}, {path}, {path}}
	fixture.ports.EnsureBrief = func(_ string, brief domain.ChangeBrief) (bool, error) {
		fixture.brief = brief
		return false, nil
	}
	result := fixture.application().ExecuteRequest(context.Background(), productRequest())
	if result.Outcome != domain.OutcomePass || len(fixture.saved) != 1 || fixture.saved[0].BriefPath != path {
		t.Fatalf("recovery result=%+v saved=%+v", result, fixture.saved)
	}
}

func TestBriefRefusesDirtyRepositoryBeforeLockOrWrite(t *testing.T) {
	fixture := newLifecycleFixture()
	fixture.snapshotPaths = [][]string{{"unrelated.txt"}}
	result := fixture.application().ExecuteRequest(context.Background(), productRequest())
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-BRIEF-001" || fixture.acquisitions != 0 || len(fixture.saved) != 0 {
		t.Fatalf("dirty result=%+v acquisitions=%d saved=%+v", result, fixture.acquisitions, fixture.saved)
	}
}

func TestStatusDetectsLocalStateDrift(t *testing.T) {
	fixture := plannedTierOneFixture()
	loadCount := 0
	fixture.ports.LoadActive = func(string) (domain.ActiveChange, bool, error) {
		loadCount++
		active := fixture.active
		if loadCount > 1 {
			active.ID = "replacement"
		}
		return active, true, nil
	}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if result.Outcome != domain.OutcomeFailed || result.Code != "L7-STATE-004" {
		t.Fatalf("drifting status=%+v", result)
	}
}

func TestLifecycleUseCaseStopsBeforePortsWhenCancelled(t *testing.T) {
	fixture := newLifecycleFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := fixture.application().ExecuteRequest(ctx, productRequest())
	if result.Outcome != domain.OutcomeCancelled || fixture.locates != 0 || fixture.acquisitions != 0 {
		t.Fatalf("cancelled result=%+v locates=%d acquisitions=%d", result, fixture.locates, fixture.acquisitions)
	}
}

type lifecycleFixture struct {
	ports         Ports
	location      domain.RepositoryLocation
	configuration domain.Configuration
	active        domain.ActiveChange
	activeFound   bool
	brief         domain.ChangeBrief
	snapshotPaths [][]string
	snapshotIndex int
	saved         []domain.ActiveChange
	locates       int
	acquisitions  int
}

func newLifecycleFixture() *lifecycleFixture {
	fixture := &lifecycleFixture{
		location:      domain.RepositoryLocation{Root: "/repo", CommonDir: "/repo/.git", Head: strings.Repeat("a", 40), Tree: strings.Repeat("b", 40)},
		configuration: domain.Configuration{LocalLifecycle: true, MaxGitOutputBytes: 1 << 20, MaxGitPaths: 1000},
		snapshotPaths: [][]string{{}, {}, {}},
	}
	fixture.ports = Ports{
		Locate: func(context.Context, string) (domain.RepositoryLocation, error) {
			fixture.locates++
			return fixture.location, nil
		},
		Snapshot: func(_ context.Context, _ string, base string, _, _ int) (domain.RepositorySnapshot, error) {
			index := fixture.snapshotIndex
			if index >= len(fixture.snapshotPaths) {
				index = len(fixture.snapshotPaths) - 1
			}
			fixture.snapshotIndex++
			return domain.RepositorySnapshot{RepositoryLocation: fixture.location, Base: base, ChangedPaths: append([]string{}, fixture.snapshotPaths[index]...)}, nil
		},
		LoadConfiguration: func(string) (domain.Configuration, bool, error) { return fixture.configuration, true, nil },
		AdoptConfiguration: func(_ string, enable bool) (domain.Configuration, bool, error) {
			configuration := fixture.configuration
			configuration.LocalLifecycle = configuration.LocalLifecycle || enable
			return configuration, true, nil
		},
		LoadActive: func(string) (domain.ActiveChange, bool, error) { return fixture.active, fixture.activeFound, nil },
		SaveActive: func(_ string, active domain.ActiveChange) error {
			fixture.saved = append(fixture.saved, active)
			fixture.active = active
			fixture.activeFound = true
			return nil
		},
		Acquire: func(string) (func() error, error) {
			fixture.acquisitions++
			return func() error { return nil }, nil
		},
		EnsureBrief: func(_ string, brief domain.ChangeBrief) (bool, error) {
			fixture.brief = brief
			return true, nil
		},
		LoadBrief: func(string, string) (domain.ChangeBrief, error) { return fixture.brief, nil },
	}
	return fixture
}

func (fixture *lifecycleFixture) application() Application {
	return NewLifecycle("test-version", "/repo", fixture.ports)
}

func plannedTierOneFixture() *lifecycleFixture {
	fixture := newLifecycleFixture()
	fixture.active = domain.ActiveChange{
		Kind: domain.ActiveTierOne, ID: "routine-fix", Tier: domain.TierRoutine,
		Base: fixture.location.Head, Problem: "Fix a routine defect.", Scope: []string{"internal/example/**"},
	}
	fixture.activeFound = true
	return fixture
}

func productRequest() domain.Request {
	return domain.Request{
		Command: domain.CommandBrief, ChangeID: "product-feature", Tier: domain.TierProduct,
		Problem: "Add a bounded product feature.", Scope: []string{"internal/product/**"},
		AcceptanceCriteria: []string{"Relevant tests pass."}, Risks: []string{"State could become stale."}, Rollback: []string{"Revert the commits."},
	}
}
