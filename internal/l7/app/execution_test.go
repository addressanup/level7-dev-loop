package app

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTierThreeFakeProvidersCompleteBothOrdersAndResume(t *testing.T) {
	for _, order := range []struct {
		implementer domain.Provider
		reviewer    domain.Provider
	}{
		{implementer: domain.ProviderCodex, reviewer: domain.ProviderClaude},
		{implementer: domain.ProviderClaude, reviewer: domain.ProviderCodex},
	} {
		t.Run(string(order.implementer)+"-to-"+string(order.reviewer), func(t *testing.T) {
			fixture := newExecutionFixture(domain.TierHighRisk)
			application := fixture.application()
			run := application.ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: order.implementer, CommitMessage: "feat(product): implement bounded change"})
			if run.Outcome != domain.OutcomePass || run.Code != "L7-RUN-000" || fixture.run.Provider.Provider != order.implementer || fixture.confirmations != 1 {
				t.Fatalf("run=%+v evidence=%+v confirmations=%d", run, fixture.run, fixture.confirmations)
			}
			verification := application.ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
			if verification.Outcome != domain.OutcomePass || verification.State != string(domain.StateAwaitingIndependentAudit) || fixture.verification.VerificationCommit == "" || fixture.verificationRuns != 1 {
				t.Fatalf("verification=%+v evidence=%+v runs=%d", verification, fixture.verification, fixture.verificationRuns)
			}
			review := application.ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: order.reviewer})
			if review.Outcome != domain.OutcomePass || review.State != string(domain.StateReviewed) || fixture.review.Provider.Provider != order.reviewer || fixture.review.ReviewCommit == "" {
				t.Fatalf("review=%+v evidence=%+v", review, fixture.review)
			}
			status := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
			if status.Outcome != domain.OutcomePass || status.State != string(domain.StateReviewed) || strings.Contains(status.State, "ready") || !strings.Contains(status.Message, "Wave 4") {
				t.Fatalf("restarted status=%+v", status)
			}
			if strings.Join(fixture.artifacts, "|") != verificationPath(fixture.active.ID)+"|"+auditPath(fixture.active.ID) {
				t.Fatalf("artifacts=%v", fixture.artifacts)
			}
		})
	}
}

func TestDisabledFlagBlocksEveryExecutionEffect(t *testing.T) {
	fixture := newExecutionFixture(domain.TierProduct)
	fixture.configuration.LocalLifecycle = false
	for _, request := range []domain.Request{
		{Command: domain.CommandRun, Agent: domain.ProviderCodex, CommitMessage: "feat(product): implement change"},
		{Command: domain.CommandVerify},
		{Command: domain.CommandReview, Agent: domain.ProviderClaude},
	} {
		result := fixture.application().ExecuteRequest(context.Background(), request)
		if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-FLAG-001" {
			t.Fatalf("request=%+v result=%+v", request, result)
		}
	}
	if fixture.providerRuns != 0 || fixture.verificationRuns != 0 || fixture.commits != 0 || fixture.acquisitions != 0 {
		t.Fatalf("effects provider=%d verification=%d commits=%d locks=%d", fixture.providerRuns, fixture.verificationRuns, fixture.commits, fixture.acquisitions)
	}
}

func TestRunRejectsProviderGitAndScopeMutation(t *testing.T) {
	for _, test := range []struct {
		name      string
		mutate    func(*executionFixture)
		wantCode  string
		wantPaths []string
	}{
		{name: "provider commit", mutate: func(fixture *executionFixture) { fixture.providerCommits = true }, wantCode: "L7-GIT-003"},
		{name: "scope expansion", mutate: func(fixture *executionFixture) { fixture.providerPaths = []string{"outside.txt"} }, wantCode: "L7-SCOPE-001", wantPaths: []string{"outside.txt"}},
		{name: "index mutation", mutate: func(fixture *executionFixture) { fixture.providerStages = true }, wantCode: "L7-GIT-004"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newExecutionFixture(domain.TierProduct)
			test.mutate(fixture)
			result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: domain.ProviderCodex, CommitMessage: "feat(product): implement change"})
			if result.Code != test.wantCode || fixture.commits != 0 {
				t.Fatalf("result=%+v commits=%d", result, fixture.commits)
			}
			if len(test.wantPaths) != 0 && (result.Repository == nil || strings.Join(result.Repository.ExpandedPaths, "|") != strings.Join(test.wantPaths, "|")) {
				t.Fatalf("expanded result=%+v", result)
			}
		})
	}
}

func TestInterruptedTransitionsResumeWithoutRelaunch(t *testing.T) {
	t.Run("run commit", func(t *testing.T) {
		fixture := newExecutionFixture(domain.TierProduct)
		fixture.pendingPaths = []string{"internal/product/change.go"}
		fixture.addOverall(fixture.pendingPaths...)
		fixture.run = domain.RunEvidence{
			ChangeID: fixture.active.ID, Provider: fixture.identity(domain.ProviderCodex),
			Parent:     domain.CandidateIdentity{Commit: fixture.location.Head, Tree: fixture.location.Tree},
			PathDigest: fixture.pathDigest(fixture.pendingPaths), PathCount: 1, CommitMessage: "feat(product): implement change",
		}
		fixture.runFound = true
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: domain.ProviderCodex, CommitMessage: fixture.run.CommitMessage})
		if result.Outcome != domain.OutcomePass || fixture.providerRuns != 0 || fixture.run.Candidate.Commit == "" {
			t.Fatalf("result=%+v providerRuns=%d evidence=%+v", result, fixture.providerRuns, fixture.run)
		}
	})

	t.Run("verification record", func(t *testing.T) {
		fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
		fixture.verification = domain.VerificationEvidence{ChangeID: fixture.active.ID, Candidate: fixture.run.Candidate, Result: domain.DecisionGO, Checks: fixture.passingChecks()}
		fixture.verificationFound = true
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
		if result.Outcome != domain.OutcomePass || fixture.verificationRuns != 0 || fixture.verification.VerificationCommit == "" {
			t.Fatalf("result=%+v verificationRuns=%d evidence=%+v", result, fixture.verificationRuns, fixture.verification)
		}
	})

	t.Run("audit record", func(t *testing.T) {
		fixture := completedVerificationFixture(domain.ProviderCodex)
		fixture.review = domain.ReviewEvidence{ChangeID: fixture.active.ID, Provider: fixture.identity(domain.ProviderClaude), Candidate: fixture.verification.Candidate, Decision: domain.DecisionGO}
		fixture.reviewFound = true
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
		if result.Outcome != domain.OutcomePass || fixture.providerRuns != 0 || fixture.review.ReviewCommit == "" {
			t.Fatalf("result=%+v providerRuns=%d evidence=%+v", result, fixture.providerRuns, fixture.review)
		}
	})
}

func TestVerificationFailureAndReviewerMutationCannotAdvance(t *testing.T) {
	fixture := completedRunFixture(domain.TierProduct, domain.ProviderCodex)
	fixture.verificationError = errors.New("tests failed")
	failed := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
	if failed.Outcome != domain.OutcomeFailed || failed.State != string(domain.StateBuilding) || fixture.verificationFound {
		t.Fatalf("failed verification=%+v evidence=%+v", failed, fixture.verification)
	}

	fixture = completedVerificationFixture(domain.ProviderCodex)
	fixture.providerPaths = []string{"internal/product/reviewer-write.go"}
	fixture.reviewerMutates = true
	mutated := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
	if mutated.Outcome != domain.OutcomeBlocked || mutated.Code != "L7-REVIEW-007" || fixture.reviewFound {
		t.Fatalf("mutating review=%+v evidence=%+v", mutated, fixture.review)
	}
}

func TestTierThreeRejectsSelfReviewBeforeProviderLaunch(t *testing.T) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderCodex})
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-REVIEW-003" || fixture.providerRuns != 0 {
		t.Fatalf("result=%+v providerRuns=%d", result, fixture.providerRuns)
	}
}

func TestReviewerNoGoReturnsToBuildingAndReconstructs(t *testing.T) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	fixture.providerDecision = domain.DecisionNoGO
	review := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
	if review.Outcome != domain.OutcomeBlocked || review.Code != "L7-REVIEW-010" || review.State != string(domain.StateBuilding) || !strings.Contains(review.Next, "l7 run") || !fixture.reviewFound {
		t.Fatalf("NO_GO review=%+v evidence=%+v", review, fixture.review)
	}
	status := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if status.Outcome != domain.OutcomeBlocked || status.Code != "L7-BUILD-001" || status.State != string(domain.StateBuilding) || !strings.Contains(status.Next, "l7 run") {
		t.Fatalf("NO_GO status=%+v", status)
	}
}

func TestDirtyCandidateCannotRetainCurrentAssurance(t *testing.T) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	fixture.pendingPaths = []string{"internal/product/change.go"}
	status := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
	if status.Outcome != domain.OutcomeBlocked || status.Code != "L7-BUILD-001" || status.State != string(domain.StateBuilding) {
		t.Fatalf("dirty candidate status=%+v", status)
	}
}

func TestTierThreeDirtyIntakeBlocksBeforeApproval(t *testing.T) {
	fixture := newExecutionFixture(domain.TierHighRisk)
	fixture.pendingPaths = []string{"internal/product/change.go"}
	fixture.addOverall(fixture.pendingPaths...)
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: domain.ProviderCodex, CommitMessage: "feat(product): implement change"})
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-RUN-001" || fixture.confirmations != 0 || fixture.providerRuns != 0 {
		t.Fatalf("result=%+v confirmations=%d providerRuns=%d", result, fixture.confirmations, fixture.providerRuns)
	}
}

func TestResumeRevalidatesContextAndEvidenceAfterLock(t *testing.T) {
	t.Run("verification context", func(t *testing.T) {
		fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
		fixture.verification = domain.VerificationEvidence{ChangeID: fixture.active.ID, Candidate: fixture.run.Candidate, Result: domain.DecisionGO, Checks: fixture.passingChecks()}
		fixture.verificationFound = true
		fixture.onAcquire = func() { fixture.configuration.MaxCommandSeconds++ }
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
		if result.Outcome != domain.OutcomeFailed || result.Code != "L7-GIT-002" || len(fixture.artifacts) != 0 {
			t.Fatalf("result=%+v artifacts=%v", result, fixture.artifacts)
		}
	})

	t.Run("review evidence", func(t *testing.T) {
		fixture := completedVerificationFixture(domain.ProviderCodex)
		fixture.review = domain.ReviewEvidence{ChangeID: fixture.active.ID, Provider: fixture.identity(domain.ProviderClaude), Candidate: fixture.verification.Candidate, Decision: domain.DecisionGO}
		fixture.reviewFound = true
		fixture.onAcquire = func() { fixture.review.Findings = []string{"changed while waiting"} }
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
		if result.Outcome != domain.OutcomeFailed || result.Code != "L7-GIT-002" || len(fixture.artifacts) != 1 {
			t.Fatalf("result=%+v artifacts=%v", result, fixture.artifacts)
		}
	})
}

func TestExecutionEffectsCannotRewriteLocalBindings(t *testing.T) {
	t.Run("implementer active state", func(t *testing.T) {
		fixture := newExecutionFixture(domain.TierProduct)
		fixture.onProvider = func(domain.ProviderTask) { fixture.active.ID = "provider-replacement" }
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: domain.ProviderCodex, CommitMessage: "feat(product): implement change"})
		if result.Outcome != domain.OutcomeFailed || fixture.commits != 0 || fixture.runFound {
			t.Fatalf("result=%+v commits=%d run=%+v", result, fixture.commits, fixture.run)
		}
	})

	t.Run("verification evidence", func(t *testing.T) {
		fixture := completedRunFixture(domain.TierProduct, domain.ProviderCodex)
		fixture.onVerification = func() { fixture.run.CommitMessage = "fix(product): replace binding" }
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
		if result.Outcome != domain.OutcomeFailed || result.Code != "L7-VERIFY-008" || fixture.verificationFound {
			t.Fatalf("result=%+v verification=%+v", result, fixture.verification)
		}
	})

	t.Run("review evidence", func(t *testing.T) {
		fixture := completedVerificationFixture(domain.ProviderCodex)
		fixture.onProvider = func(task domain.ProviderTask) {
			if task.Role == domain.RoleReviewer {
				fixture.run.CommitMessage = "fix(product): replace binding"
			}
		}
		result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
		if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-REVIEW-007" || fixture.reviewFound {
			t.Fatalf("result=%+v review=%+v", result, fixture.review)
		}
	})
}

func TestTierThreeAssuranceRequiresCurrentOwnerBinding(t *testing.T) {
	fixture := completedRunFixture(domain.TierHighRisk, domain.ProviderCodex)
	fixture.approvalFound = false
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
	if result.Outcome != domain.OutcomeBlocked || result.Code != "L7-AUTH-001" || fixture.verificationRuns != 0 {
		t.Fatalf("result=%+v verificationRuns=%d", result, fixture.verificationRuns)
	}
}

func TestReviewerTaskNamesExactVerifiedCandidate(t *testing.T) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	want := fixture.verification.Candidate
	if want.Commit == fixture.location.Head {
		t.Fatal("fixture does not distinguish product and verification commits")
	}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandReview, Agent: domain.ProviderClaude})
	if result.Outcome != domain.OutcomePass || fixture.lastTask.Candidate != want {
		t.Fatalf("result=%+v task=%+v want=%+v", result, fixture.lastTask, want)
	}
}

func BenchmarkStatusReconstruction(b *testing.B) {
	fixture := completedVerificationFixture(domain.ProviderCodex)
	application := fixture.application()
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		result := application.ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandStatus})
		if result.State != string(domain.StateAwaitingIndependentAudit) {
			b.Fatal(result)
		}
	}
}

type fakeCommit struct {
	parent  string
	tree    string
	subject string
	paths   []string
}

type executionFixture struct {
	ports                 Ports
	location              domain.RepositoryLocation
	configuration         domain.Configuration
	active                domain.ActiveChange
	brief                 domain.ChangeBrief
	overallPaths          []string
	pendingPaths          []string
	indexDirty            bool
	providerPaths         []string
	providerStages        bool
	providerCommits       bool
	reviewerMutates       bool
	providerDecision      domain.ReviewDecision
	verificationError     error
	run                   domain.RunEvidence
	runFound              bool
	verification          domain.VerificationEvidence
	verificationFound     bool
	review                domain.ReviewEvidence
	reviewFound           bool
	approval              domain.ApprovalBinding
	approvalFound         bool
	commitsByID           map[string]fakeCommit
	treesByCommit         map[string]string
	artifacts             []string
	providerRuns          int
	verificationRuns      int
	commits               int
	confirmations         int
	acquisitions          int
	onAcquire             func()
	onProvider            func(domain.ProviderTask)
	onVerification        func()
	lastTask              domain.ProviderTask
	nextIdentityCharacter byte
}

func newExecutionFixture(tier domain.RiskTier) *executionFixture {
	base := strings.Repeat("a", 40)
	baseTree := strings.Repeat("b", 40)
	briefCommit := strings.Repeat("c", 40)
	briefTree := strings.Repeat("d", 40)
	changeID := "product-change"
	briefPath := "docs/artifacts/changes/" + changeID + ".md"
	fixture := &executionFixture{
		location: domain.RepositoryLocation{Root: "/repo", CommonDir: "/repo/.git", Head: briefCommit, Tree: briefTree},
		configuration: domain.Configuration{
			LocalLifecycle: true, Verification: []domain.VerificationCommand{{Name: "test", Argv: []string{"make", "test"}}},
			MaxInputBytes: 1 << 20, MaxGitOutputBytes: 1 << 20, MaxGitPaths: 1000, MaxCommandOutputBytes: 1 << 20, MaxCommandSeconds: 30,
		},
		active: domain.ActiveChange{Kind: domain.ActiveBrief, ID: changeID, BriefPath: briefPath},
		brief: domain.ChangeBrief{
			ID: changeID, Tier: tier, Base: base, Path: briefPath, Problem: "Implement a bounded product change.", Scope: []string{"internal/product/**"},
			AcceptanceCriteria: []string{"Relevant tests pass."}, Risks: []string{"A regression is possible."}, Rollback: []string{"Revert the candidate."},
		},
		overallPaths: []string{briefPath}, providerPaths: []string{"internal/product/change.go"}, providerDecision: domain.DecisionGO,
		commitsByID:   map[string]fakeCommit{briefCommit: {parent: base, tree: briefTree, subject: "docs(product): add change brief", paths: []string{briefPath}}},
		treesByCommit: map[string]string{base: baseTree, briefCommit: briefTree}, nextIdentityCharacter: 'e',
	}
	fixture.ports = fixture.buildPorts()
	return fixture
}

func completedRunFixture(tier domain.RiskTier, implementer domain.Provider) *executionFixture {
	fixture := newExecutionFixture(tier)
	if tier == domain.TierHighRisk {
		fixture.approval = domain.ApprovalBinding{ChangeID: fixture.active.ID, Actor: "accountable-owner", Implementer: implementer, BriefCommit: fixture.location.Head}
		fixture.approvalFound = true
	}
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandRun, Agent: implementer, CommitMessage: "feat(product): implement change"})
	if result.Outcome != domain.OutcomePass {
		panic("fake run did not complete: " + result.Message)
	}
	fixture.providerRuns = 0
	return fixture
}

func completedVerificationFixture(implementer domain.Provider) *executionFixture {
	fixture := completedRunFixture(domain.TierHighRisk, implementer)
	result := fixture.application().ExecuteRequest(context.Background(), domain.Request{Command: domain.CommandVerify})
	if result.Outcome != domain.OutcomePass {
		panic("fake verification did not complete: " + result.Message)
	}
	fixture.verificationRuns = 0
	return fixture
}

func (fixture *executionFixture) application() Application {
	return NewLifecycle("test-version", "/repo", fixture.ports)
}

func (fixture *executionFixture) buildPorts() Ports {
	return Ports{
		Locate: func(context.Context, string) (domain.RepositoryLocation, error) { return fixture.location, nil },
		Snapshot: func(_ context.Context, _ string, base string, _, _ int) (domain.RepositorySnapshot, error) {
			return domain.RepositorySnapshot{RepositoryLocation: fixture.location, Base: base, ChangedPaths: append([]string{}, fixture.overallPaths...)}, nil
		},
		LoadConfiguration:  func(string) (domain.Configuration, bool, error) { return fixture.configuration, true, nil },
		AdoptConfiguration: func(string, bool) (domain.Configuration, bool, error) { return fixture.configuration, false, nil },
		LoadActive:         func(string) (domain.ActiveChange, bool, error) { return fixture.active, true, nil },
		SaveActive:         func(string, domain.ActiveChange) error { return nil },
		Acquire: func(string) (func() error, error) {
			fixture.acquisitions++
			if fixture.onAcquire != nil {
				fixture.onAcquire()
			}
			return func() error { return nil }, nil
		},
		EnsureBrief: func(string, domain.ChangeBrief) (bool, error) { return false, nil },
		LoadBrief:   func(string, string) (domain.ChangeBrief, error) { return fixture.brief, nil },
		Pending: func(context.Context, string, int, int) (domain.PendingChanges, error) {
			return domain.PendingChanges{RepositoryLocation: fixture.location, Paths: append([]string{}, fixture.pendingPaths...), IndexDirty: fixture.indexDirty}, nil
		},
		PathCommit: func(context.Context, string, string) (string, error) { return strings.Repeat("c", 40), nil },
		Commit:     fixture.commit,
		CommitMatches: func(_ context.Context, _ string, commit, parent, subject string, _, _ int) (bool, error) {
			stored, found := fixture.commitsByID[commit]
			return found && stored.parent == parent && stored.subject == subject, nil
		},
		CommitPaths: func(_ context.Context, _ string, parent, commit string, _, _ int) ([]string, error) {
			stored, found := fixture.commitsByID[commit]
			if !found || stored.parent != parent {
				return nil, errors.New("unknown commit")
			}
			return append([]string{}, stored.paths...), nil
		},
		CommitTree: func(_ context.Context, _ string, commit string) (string, error) {
			tree, found := fixture.treesByCommit[commit]
			if !found {
				return "", errors.New("unknown tree")
			}
			return tree, nil
		},
		PathSetDigest: fixture.pathDigest,
		ConfirmApproval: func(_ context.Context, changeID string, implementer domain.Provider, briefCommit string) (domain.ApprovalBinding, error) {
			fixture.confirmations++
			return domain.ApprovalBinding{ChangeID: changeID, Actor: "accountable-owner", Implementer: implementer, BriefCommit: briefCommit}, nil
		},
		LoadApproval: func(string) (domain.ApprovalBinding, bool, error) {
			return fixture.approval, fixture.approvalFound, nil
		},
		SaveApproval: func(_ string, approval domain.ApprovalBinding) error {
			fixture.approval, fixture.approvalFound = approval, true
			return nil
		},
		LoadRun: func(string) (domain.RunEvidence, bool, error) { return fixture.run, fixture.runFound, nil },
		SaveRun: func(_ string, evidence domain.RunEvidence) error {
			fixture.run, fixture.runFound = evidence, true
			return nil
		},
		LoadVerification: func(string) (domain.VerificationEvidence, bool, error) {
			return fixture.verification, fixture.verificationFound, nil
		},
		SaveVerification: func(_ string, evidence domain.VerificationEvidence) error {
			fixture.verification, fixture.verificationFound = evidence, true
			return nil
		},
		LoadReview: func(string) (domain.ReviewEvidence, bool, error) { return fixture.review, fixture.reviewFound, nil },
		SaveReview: func(_ string, evidence domain.ReviewEvidence) error {
			fixture.review, fixture.reviewFound = evidence, true
			return nil
		},
		RunProvider: fixture.runProvider,
		RunVerification: func(context.Context, string, []domain.VerificationCommand, int, int) ([]domain.CheckResult, error) {
			fixture.verificationRuns++
			if fixture.onVerification != nil {
				fixture.onVerification()
			}
			if fixture.verificationError != nil {
				return []domain.CheckResult{{Name: "test", Passed: false, ExitCode: 1, Code: "L7-VERIFY-001", Message: "tests failed"}}, fixture.verificationError
			}
			return fixture.passingChecks(), nil
		},
		WriteVerification: func(_ string, _ domain.VerificationEvidence, _ string) (string, error) {
			path := verificationPath(fixture.active.ID)
			fixture.pendingPaths = []string{path}
			fixture.addOverall(path)
			fixture.artifacts = append(fixture.artifacts, path)
			return path, nil
		},
		WriteAudit: func(_ string, _ domain.ReviewEvidence) (string, error) {
			path := auditPath(fixture.active.ID)
			fixture.pendingPaths = []string{path}
			fixture.addOverall(path)
			fixture.artifacts = append(fixture.artifacts, path)
			return path, nil
		},
	}
}

func (fixture *executionFixture) runProvider(_ context.Context, task domain.ProviderTask, _, _ int) (domain.ProviderResponse, error) {
	fixture.providerRuns++
	fixture.lastTask = task
	if fixture.onProvider != nil {
		fixture.onProvider(task)
	}
	paths := append([]string{}, fixture.providerPaths...)
	if task.Role == domain.RoleReviewer && !fixture.reviewerMutates {
		paths = nil
	}
	if fixture.providerCommits {
		parent := fixture.location.Head
		fixture.location = domain.RepositoryLocation{Root: fixture.location.Root, CommonDir: fixture.location.CommonDir, Head: strings.Repeat("9", 40), Tree: strings.Repeat("8", 40)}
		fixture.commitsByID[fixture.location.Head] = fakeCommit{parent: parent, tree: fixture.location.Tree, subject: "feat(provider): unauthorized commit", paths: append([]string{}, paths...)}
		fixture.treesByCommit[fixture.location.Head] = fixture.location.Tree
	}
	fixture.pendingPaths = append([]string{}, paths...)
	fixture.indexDirty = fixture.providerStages
	fixture.addOverall(paths...)
	decision := domain.ReviewDecision("")
	if task.Role == domain.RoleReviewer {
		decision = fixture.providerDecision
	}
	return domain.ProviderResponse{Identity: fixture.identity(task.Provider), Role: task.Role, Summary: "completed", Decision: decision}, nil
}

func (fixture *executionFixture) commit(_ context.Context, request domain.CommitRequest) (domain.RepositoryLocation, error) {
	if request.ExpectedCommit != fixture.location.Head || request.ExpectedTree != fixture.location.Tree || !testPathSetsEqual(request.Paths, fixture.pendingPaths) {
		return domain.RepositoryLocation{}, errors.New("stale or expanded commit")
	}
	fixture.commits++
	parent := fixture.location.Head
	commit := strings.Repeat(string(fixture.nextIdentityCharacter), 40)
	fixture.nextIdentityCharacter++
	tree := strings.Repeat(string(fixture.nextIdentityCharacter), 40)
	fixture.nextIdentityCharacter++
	fixture.location.Head, fixture.location.Tree = commit, tree
	fixture.commitsByID[commit] = fakeCommit{parent: parent, tree: tree, subject: request.Message, paths: append([]string{}, request.Paths...)}
	fixture.treesByCommit[commit] = tree
	fixture.pendingPaths = nil
	fixture.indexDirty = false
	return fixture.location, nil
}

func (fixture *executionFixture) identity(provider domain.Provider) domain.ProviderIdentity {
	version := "codex-cli 0.149.1"
	if provider == domain.ProviderClaude {
		version = "2.1.241 (Claude Code)"
	}
	return domain.ProviderIdentity{Provider: provider, Executable: "/usr/bin/" + string(provider), Version: version, Digest: strings.Repeat(string(provider[0]), 64), Capability: domain.CapabilityAvailable}
}

func (fixture *executionFixture) pathDigest(paths []string) string {
	ordered := append([]string{}, paths...)
	sort.Strings(ordered)
	return strings.Join(ordered, "\x00")
}

func (fixture *executionFixture) passingChecks() []domain.CheckResult {
	return []domain.CheckResult{{Name: "test", Passed: true, ExitCode: 0, Code: "L7-VERIFY-000", Message: "command passed"}}
}

func (fixture *executionFixture) addOverall(paths ...string) {
	for _, path := range paths {
		found := false
		for _, existing := range fixture.overallPaths {
			found = found || existing == path
		}
		if !found {
			fixture.overallPaths = append(fixture.overallPaths, path)
		}
	}
}

func testPathSetsEqual(left, right []string) bool {
	leftCopy, rightCopy := append([]string{}, left...), append([]string{}, right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	return strings.Join(leftCopy, "\x00") == strings.Join(rightCopy, "\x00")
}
