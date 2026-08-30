package main

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSensitiveUnprotectedLookingPathCannotUseTierOne(t *testing.T) {
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("internal/authorization/check.go", "package authorization\n")
	repository.commit("feat: authorization change")
	_, findings := runController(controllerOptions{Root: repository.root, BaseRef: base, HeadRef: "HEAD", ChangeID: "auth-change", Tier: tierRoutine, TierOneScope: []string{"internal/authorization/check.go"}})
	if rules(findings)["RISK-003"] == 0 {
		t.Fatalf("authorization change used Tier 1: %+v", findings)
	}
}

func TestTier3SoloFastPathNeedsNoSeparateOwnerOrAudit(t *testing.T) {
	repository, _ := tierThreeImplementation(t)
	options := controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"}

	report, findings := runController(options)
	if len(findings) != 0 || report.State != stateBuilding || report.Assurance != assuranceSolo {
		t.Fatalf("solo Tier 3 did not enter building: report=%+v findings=%+v", report, findings)
	}
	options.VerifiedRef = "HEAD"
	report, findings = runController(options)
	if len(findings) != 0 || report.State != stateVerified || !strings.Contains(report.Next, "self-review") {
		t.Fatalf("solo Tier 3 did not accept exact-head verification: report=%+v findings=%+v", report, findings)
	}
	options.ReviewRef = "HEAD"
	report, findings = runController(options)
	if len(findings) != 0 || report.State != stateReviewed {
		t.Fatalf("solo Tier 3 did not accept truthful self-review: report=%+v findings=%+v", report, findings)
	}
	options.ReadyRef = "HEAD"
	options.RequireReady = true
	report, findings = runController(options)
	if len(findings) != 0 || report.State != stateReady {
		t.Fatalf("solo Tier 3 did not become ready: report=%+v findings=%+v", report, findings)
	}
}

func TestTier3SoloRejectsEvidenceOnlyArtifacts(t *testing.T) {
	repository, _ := tierThreeImplementation(t)
	candidate := repository.rev("HEAD")
	tree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-verification.md", evidenceDocument("controller", candidate, tree, "PASS", "self"))
	repository.commit("docs: add redundant verification")

	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller"})
	if rules(findings)["ART-003"] == 0 || rules(findings)["ART-004"] == 0 {
		t.Fatalf("solo evidence-only artifact was accepted: %+v", findings)
	}
}

func TestInvalidAssuranceModeFailsClosed(t *testing.T) {
	repository, _ := tierThreeImplementation(t)
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: "pretend-independent"})
	if rules(findings)["ASSURANCE-001"] == 0 {
		t.Fatalf("invalid assurance mode was accepted: %+v", findings)
	}
}

func TestTier3RejectsMissingInvalidAndSelfIssuedApproval(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if rules(findings)["AUTH-001"] == 0 {
		t.Fatalf("missing approval accepted: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: "0000000000000000000000000000000000000000", Source: "active-user-interaction"})
	_, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if rules(findings)["AUTH-002"] == 0 {
		t.Fatalf("mismatched approval accepted: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "codex", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	_, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if rules(findings)["AUTH-003"] == 0 {
		t.Fatalf("self-issued approval accepted: %+v", findings)
	}
}

func TestTier3ValidatesApprovalBeforeImplementation(t *testing.T) {
	repository, briefCommit := tierThreeBrief(t)
	options := controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam}

	_, findings := runController(options)
	if rules(findings)["AUTH-001"] == 0 {
		t.Fatalf("brief-only Tier 3 change accepted without approval: %+v", findings)
	}

	repository.write(".git/l7/approvals/controller.json", "{not-json}\n")
	_, findings = runController(options)
	if rules(findings)["INPUT-002"] == 0 {
		t.Fatalf("brief-only Tier 3 change accepted malformed approval: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "codex", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	_, findings = runController(options)
	if rules(findings)["AUTH-003"] == 0 {
		t.Fatalf("brief-only Tier 3 change accepted self-issued approval: %+v", findings)
	}

	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	report, findings := runController(options)
	if len(findings) != 0 || report.State != stateBuilding {
		t.Fatalf("valid pre-build approval did not authorize implementation: report=%+v findings=%+v", report, findings)
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
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if rules(findings)["AUTH-004"] == 0 {
		t.Fatalf("mutated approved brief accepted: %+v", findings)
	}
}

func TestTier3ApprovalContinuesThroughVerificationAndIndependentAudit(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})

	report, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if len(findings) != 0 || report.State != stateBuilding {
		t.Fatalf("approved Tier 3 change deadlocked: report=%+v findings=%+v", report, findings)
	}
	implementationCommit := repository.rev("HEAD")
	implementationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-verification.md", evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci"))
	verificationCommit := repository.commit("test: verify controller")
	verificationTree := repository.rev("HEAD^{tree}")

	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if len(findings) != 0 || report.State != stateVerified {
		t.Fatalf("verified Tier 3 state is unreachable: report=%+v findings=%+v", report, findings)
	}
	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam, AuditRequestRef: "HEAD"})
	if len(findings) != 0 || report.State != stateAwaitingIndependentAudit {
		t.Fatalf("verified Tier 3 change did not enter audit: report=%+v findings=%+v", report, findings)
	}

	repository.write("docs/artifacts/changes/controller-audit.md", evidenceDocument("controller", verificationCommit, verificationTree, "GO", "auditor"))
	auditCommit := repository.commit("docs: record independent audit")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "auditor", CandidateCommit: verificationCommit, AuditCommit: auditCommit, Source: "independent-agent"})
	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if len(findings) != 0 || report.State != stateReviewed {
		t.Fatalf("reviewed Tier 3 state is unreachable: report=%+v findings=%+v", report, findings)
	}
	report, findings = runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam, ReadyRef: "HEAD", RequireReady: true})
	if len(findings) != 0 || report.State != stateReady {
		t.Fatalf("independently audited Tier 3 change not ready: report=%+v findings=%+v", report, findings)
	}
}

func TestBoundNoGoReturnsTier3ToBuilding(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	implementationCommit := repository.rev("HEAD")
	implementationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-verification.md", evidenceDocument("controller", implementationCommit, implementationTree, "PASS", "ci"))
	verificationCommit := repository.commit("test: verify controller")
	verificationTree := repository.rev("HEAD^{tree}")
	repository.write("docs/artifacts/changes/controller-audit.md", evidenceDocument("controller", verificationCommit, verificationTree, "NO_GO", "auditor"))
	auditCommit := repository.commit("docs: record no-go")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "auditor", CandidateCommit: verificationCommit, AuditCommit: auditCommit, Source: "independent-agent"})
	report, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if len(findings) != 0 || report.State != stateBuilding {
		t.Fatalf("bound NO_GO did not return to building: report=%+v findings=%+v", report, findings)
	}
}

func TestRemediationCanReachFreshAuditAfterHistoricalNoGo(t *testing.T) {
	repository, briefCommit := tierThreeImplementation(t)
	repository.authority("approvals", "controller", approvalEnvelope{Schema: 1, ChangeID: "controller", Actor: "owner", Implementer: "codex", BriefCommit: briefCommit, Source: "active-user-interaction"})
	firstCandidate := repository.rev("HEAD")
	firstTree := repository.rev("HEAD^{tree}")
	verificationPath := "docs/artifacts/changes/controller-verification.md"
	auditPath := "docs/artifacts/changes/controller-audit.md"
	repository.write(verificationPath, evidenceDocument("controller", firstCandidate, firstTree, "PASS", "ci"))
	firstVerification := repository.commit("test: first verification")
	firstVerificationTree := repository.rev("HEAD^{tree}")
	repository.write(auditPath, evidenceDocument("controller", firstVerification, firstVerificationTree, "NO_GO", "auditor"))
	firstAudit := repository.commit("docs: first no-go")
	repository.authority("audits", "controller", auditEnvelope{Schema: 1, ChangeID: "controller", Actor: "auditor", CandidateCommit: firstVerification, AuditCommit: firstAudit, Source: "independent-agent"})

	repository.remove(verificationPath)
	repository.write("internal/harness/buildcontrol/policy.go", "package main\n// remediated\n")
	remediatedCandidate := repository.commit("fix: remediate audit")
	remediatedTree := repository.rev("HEAD^{tree}")
	repository.write(verificationPath, evidenceDocument("controller", remediatedCandidate, remediatedTree, "PASS", "ci"))
	secondVerification := repository.commit("test: rebound verification")

	report, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam, AuditRequestRef: secondVerification})
	if len(findings) != 0 || report.State != stateAwaitingIndependentAudit {
		t.Fatalf("historical NO_GO deadlocked fresh audit: report=%+v findings=%+v", report, findings)
	}
}

func TestTierOneAndTwoReachVerifiedReviewedAndReady(t *testing.T) {
	for _, tier := range []riskTier{tierRoutine, tierProduct} {
		repository := newTestRepository(t)
		base := repository.rev("HEAD")
		options := controllerOptions{Root: repository.root, BaseRef: base, HeadRef: "HEAD", ChangeID: "normal", Tier: tierRoutine, TierOneScope: []string{"internal/change.go"}}
		if tier == tierProduct {
			repository.write("docs/artifacts/changes/normal.md", briefDocument("normal", tierProduct, base, "internal/change.go"))
			options = controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "normal"}
		}
		repository.write("internal/change.go", "package internal\n")
		repository.commit("feat: normal change")
		options.VerifiedRef = "HEAD"
		report, findings := runController(options)
		if len(findings) != 0 || report.State != stateVerified {
			t.Fatalf("tier %d verified state: report=%+v findings=%+v", tier, report, findings)
		}
		options.ReviewRef = "HEAD"
		report, findings = runController(options)
		if len(findings) != 0 || report.State != stateReviewed {
			t.Fatalf("tier %d reviewed state: report=%+v findings=%+v", tier, report, findings)
		}
		options.ReadyRef = "HEAD"
		report, findings = runController(options)
		if len(findings) != 0 || report.State != stateReady {
			t.Fatalf("tier %d ready state: report=%+v findings=%+v", tier, report, findings)
		}
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
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
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
	_, findings := runController(controllerOptions{Root: repository.root, HeadRef: "HEAD", ChangeID: "controller", Assurance: assuranceTeam})
	if rules(findings)["AUDIT-007"] == 0 {
		t.Fatalf("verification mutation in audit successor accepted: %+v", findings)
	}
}

func tierThreeImplementation(t *testing.T) (*testRepository, string) {
	t.Helper()
	repository, briefCommit := tierThreeBrief(t)
	repository.write("internal/harness/buildcontrol/policy.go", "package main\n")
	repository.commit("refactor: controller")
	return repository, briefCommit
}

func tierThreeBrief(t *testing.T) (*testRepository, string) {
	t.Helper()
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("docs/artifacts/changes/controller.md", briefDocument("controller", tierHighRisk, base,
		"internal/harness/buildcontrol/policy.go",
		"docs/artifacts/changes/controller-verification.md",
		"docs/artifacts/changes/controller-audit.md"))
	briefCommit := repository.commit("docs: approve controller brief")
	return repository, briefCommit
}
