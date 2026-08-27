package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTier1FastPathNeedsNoGovernanceArtifact(t *testing.T) {
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("internal/refactor.go", "package internal\n")
	repository.commit("refactor: simplify")

	report, findings := runController(controllerOptions{Root: repository.root, BaseRef: base, HeadRef: "HEAD", ChangeID: "routine-refactor", Tier: tierRoutine, TierOneScope: []string{"internal/refactor.go"}})
	if len(findings) != 0 || report.Tier != tierRoutine || report.State != stateBuilding {
		t.Fatalf("Tier 1 fast path failed: report=%+v findings=%+v", report, findings)
	}
	if _, err := os.Stat(filepath.Join(repository.root, "docs", "artifacts")); !os.IsNotExist(err) {
		t.Fatalf("Tier 1 created governance artifacts: %v", err)
	}
}

func TestTier1IgnoresHistoricalChangeBriefs(t *testing.T) {
	repository := newTestRepository(t)
	oldBase := repository.rev("HEAD")
	repository.write("docs/artifacts/changes/old-feature.md", briefDocument("old-feature", tierProduct, oldBase, "internal/old.go"))
	repository.write("internal/old.go", "package internal\n")
	repository.commit("feat: historical feature")
	base := repository.rev("HEAD")
	repository.write("docs/note.md", "note\n")
	repository.commit("docs: routine note")

	report, findings := runController(controllerOptions{Root: repository.root, BaseRef: base, HeadRef: "HEAD", ChangeID: "routine-note", Tier: tierRoutine, TierOneScope: []string{"docs/note.md"}})
	if len(findings) != 0 || report.Tier != tierRoutine {
		t.Fatalf("historical brief blocked Tier 1: report=%+v findings=%+v", report, findings)
	}
}

func TestTier2RequiresExactlyOneChangeBrief(t *testing.T) {
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("internal/feature.go", "package internal\n")
	repository.commit("feat: feature")
	_, findings := runController(controllerOptions{Root: repository.root, BaseRef: base, HeadRef: "HEAD", ChangeID: "feature", Tier: tierProduct, TierOneScope: []string{"internal/feature.go"}})
	if rules(findings)["ART-002"] == 0 {
		t.Fatalf("Tier 2 without brief was accepted: %+v", findings)
	}

	repository = newTestRepository(t)
	base = repository.rev("HEAD")
	repository.write("docs/artifacts/changes/feature.md", briefDocument("feature", tierProduct, base, "internal/feature.go"))
	repository.write("internal/feature.go", "package internal\n")
	repository.commit("feat: feature with brief")
	report, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "feature"})
	if len(findings) != 0 || report.Tier != tierProduct {
		t.Fatalf("Tier 2 brief path failed: report=%+v findings=%+v", report, findings)
	}
}

func TestProtectedPathsElevateToTier3AndScopeExpansionFailsClosed(t *testing.T) {
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("docs/artifacts/changes/control.md", briefDocument("control", tierProduct, base, "Makefile"))
	repository.write("Makefile", "verify:\n\t@true\n")
	repository.write("outside.txt", "expanded\n")
	repository.commit("feat: unsafe control")
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "control"})
	if rules(findings)["RISK-003"] == 0 || rules(findings)["SCOPE-002"] == 0 {
		t.Fatalf("protected path or scope expansion accepted: %+v", findings)
	}
}

func TestTier3RejectsMissingInvalidAndSelfIssuedApproval(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUTH-001"] == 0 {
		t.Fatalf("missing approval accepted: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: "0000000000000000000000000000000000000000", Source: "active-user-interaction"})
	_, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUTH-002"] == 0 {
		t.Fatalf("mismatched approval accepted: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "codex", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	_, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUTH-003"] == 0 {
		t.Fatalf("self-issued approval accepted: %+v", findings)
	}
}

func TestTier3RejectsBriefMutationAfterApproval(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	name := "docs/artifacts/changes/controller.md"
	data, err := os.ReadFile(filepath.Join(repository.root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	repository.write(name, string(data)+"\nExpanded after approval.\n")
	repository.commit("docs: mutate approved brief")
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUTH-004"] == 0 {
		t.Fatalf("mutated approved brief accepted: %+v", findings)
	}
}

func TestTier3ApprovalContinuesThroughVerificationAndIndependentAudit(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})

	report, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if len(findings) != 0 || report.State != stateBuilding {
		t.Fatalf("approved Tier 3 change deadlocked: report=%+v findings=%+v", report, findings)
	}
	implementationCommit := repository.rev("HEAD")
	implementationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-verification.md", evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci"))
	verificationCommit := repository.commit("test: verify controller")
	verificationTree := repository.rev("HEAD^{tree}")

	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if len(findings) != 0 || report.State != stateAwaitingIndependentAudit {
		t.Fatalf("verified Tier 3 change did not await audit: report=%+v findings=%+v", report, findings)
	}

	repository.write("docs/artifacts/changes/controller-audit.md", evidenceDocument("controller", verificationCommit, verificationTree, "GO", "auditor"))
	auditCommit := repository.commit("docs: record independent audit")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "auditor", CandidateCommit: verificationCommit, AuditCommit: auditCommit, Source: "independent-agent"})
	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", RequireReady: true})
	if len(findings) != 0 || report.State != stateReady {
		t.Fatalf("independently audited Tier 3 change not ready: report=%+v findings=%+v", report, findings)
	}
}

func TestTier3RejectsSelfAudit(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	implementationCommit := repository.rev("HEAD")
	implementationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-verification.md", evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci"))
	verificationCommit := repository.commit("test: verify controller")
	verificationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-audit.md", evidenceDocument("controller", verificationCommit, verificationTree, "GO", "codex"))
	auditCommit := repository.commit("docs: self audit")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "codex", CandidateCommit: verificationCommit, AuditCommit: auditCommit, Source: "independent-agent"})
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUDIT-005"] == 0 {
		t.Fatalf("self-audit accepted: %+v", findings)
	}
}

func TestTier3RejectsVerificationMutationInAuditSuccessor(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	implementationCommit := repository.rev("HEAD")
	implementationTree := repository.rev("HEAD^{tree}")
	verificationPath := "docs/artifacts/changes/controller-verification.md"
	repository.write(verificationPath, evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci"))
	verificationCommit := repository.commit("test: verify controller")
	verificationTree := repository.rev("HEAD^{tree}")
	repository.write(verificationPath, evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci")+"\nChanged after verification.\n")
	repository.write("docs/artifacts/changes/controller-audit.md", evidenceDocument("controller", verificationCommit, verificationTree, "GO", "auditor"))
	auditCommit := repository.commit("docs: audit with evidence mutation")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "auditor", CandidateCommit: verificationCommit, AuditCommit: auditCommit, Source: "independent-agent"})
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["AUDIT-007"] == 0 {
		t.Fatalf("verification mutation in audit successor accepted: %+v", findings)
	}
}

func tierThreeImplementation(t *testing.T) (*testRepository, string) {
	t.Helper()
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("docs/artifacts/changes/controller.md", briefDocument("controller", tierHighRisk, base,
		"internal/harness/buildcontrol/policy.go",
		"docs/artifacts/changes/controller-verification.md",
		"docs/artifacts/changes/controller-audit.md"))
	briefCommit := repository.commit("docs: approve controller brief")
	repository.write("internal/harness/buildcontrol/policy.go", "package main\n")
	repository.commit("refactor: controller")
	return repository, briefCommit
}
