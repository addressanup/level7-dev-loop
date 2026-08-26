package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	wave02BaseCommit         = "c35bf4b6e4a38ca54899882a7e3c574d03d1df85"
	wave02BaseTree           = "eb60ac4d167df96ba02822c458cb81493e05537b"
	wave02BaseManifestSHA256 = "f7749b42b3886e438a6db17a89cfaba3da24e59bd168cf2a757fcd0e51c5780c"
	wave02CandidateManifest  = "docs/artifacts/wave-02-candidate.sha256"
	wave02EvidencePath       = "docs/artifacts/wave-02-evidence.md"
	wave02AuditPath          = "docs/artifacts/wave-02-audit.md"
	wave02EvaluatorManifest  = "harness/wave-02-evaluator-controls.sha256"
)

var expectedWave02Paths = map[string]pathExpectation{
	".github/workflows/harness.yml":                       {"modify", "harness-integrator", "SCOPE-520"},
	"README.md":                                           {"modify", "wave-integrator", "SCOPE-520"},
	"docs/artifacts/wave-02-approval.md":                  {"add", "wave-integrator", "SCOPE-521"},
	"docs/artifacts/wave-02-audit.md":                     {"audit-only", "independent-reviewer", "SCOPE-522"},
	"docs/artifacts/wave-02-candidate.sha256":             {"add", "wave-integrator", "SCOPE-521"},
	"docs/artifacts/wave-02-change-contract.md":           {"add", "wave-integrator", "SCOPE-521"},
	"docs/artifacts/wave-02-design.md":                    {"add", "wave-integrator", "SCOPE-521"},
	"docs/artifacts/wave-02-evidence.md":                  {"add", "wave-integrator", "SCOPE-521"},
	"docs/artifacts/wave-02-specification.md":             {"add", "wave-integrator", "SCOPE-521"},
	"fixtures/public/bl-002/broken-candidates.json":       {"add", "semantic-owner", "SCOPE-523"},
	"fixtures/public/bl-002/semantic-cases.json":          {"add", "semantic-owner", "SCOPE-523"},
	"fixtures/public/bl-003/adjudication.json":            {"add", "evaluator-owner", "SCOPE-524"},
	"fixtures/public/bl-003/cases.json":                   {"add", "evaluator-owner", "SCOPE-524"},
	"fixtures/public/bl-003/coverage.json":                {"add", "evaluator-owner", "SCOPE-524"},
	"fixtures/public/bl-003/grader-registry.json":         {"add", "evaluator-owner", "SCOPE-524"},
	"fixtures/public/bl-003/protocol.json":                {"add", "evaluator-owner", "SCOPE-524"},
	"fixtures/public/bl-003/truth-labels.json":            {"add", "evaluator-owner", "SCOPE-524"},
	"harness/control-ownership.tsv":                       {"modify", "harness-integrator", "SCOPE-520"},
	"harness/import-boundaries.tsv":                       {"modify", "harness-integrator", "SCOPE-520"},
	"harness/phases.tsv":                                  {"modify", "harness-integrator", "SCOPE-520"},
	"harness/wave-02-base.sha256":                         {"add", "harness-integrator", "SCOPE-521"},
	"harness/wave-02-evaluator-controls.sha256":           {"add", "harness-integrator", "SCOPE-521"},
	"harness/wave-02-paths.tsv":                           {"add", "harness-integrator", "SCOPE-521"},
	"internal/evaluator/coverage.go":                      {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/coverage_test.go":                 {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/doc.go":                           {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/grade.go":                         {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/grade_test.go":                    {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/model.go":                         {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/validate.go":                      {"add", "evaluator-owner", "SCOPE-524"},
	"internal/evaluator/validate_test.go":                 {"add", "evaluator-owner", "SCOPE-524"},
	"internal/harness/buildcontrol/main.go":               {"modify", "harness-integrator", "SCOPE-520"},
	"internal/harness/buildcontrol/ownership.go":          {"modify", "harness-integrator", "SCOPE-520"},
	"internal/harness/buildcontrol/ownership_test.go":     {"modify", "harness-integrator", "SCOPE-520"},
	"internal/harness/buildcontrol/policy.go":             {"modify", "harness-integrator", "SCOPE-520"},
	"internal/harness/buildcontrol/policy_test.go":        {"modify", "harness-integrator", "SCOPE-520"},
	"internal/harness/buildcontrol/wave2.go":              {"add", "harness-integrator", "SCOPE-521"},
	"internal/harness/buildcontrol/wave2_test.go":         {"add", "harness-integrator", "SCOPE-521"},
	"internal/render/compile.go":                          {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/compile_test.go":                     {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/decode.go":                           {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/decode_test.go":                      {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/doc.go":                              {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/model.go":                            {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/validate.go":                         {"add", "semantic-owner", "SCOPE-523"},
	"internal/render/validate_test.go":                    {"add", "semantic-owner", "SCOPE-523"},
	"schemas/evaluation/adjudication.schema.json":         {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/case.schema.json":                 {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/coverage.schema.json":             {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/grader.schema.json":               {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/protocol.schema.json":             {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/run-manifest.schema.json":         {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/evaluation/truth-label.schema.json":          {"add", "evaluator-owner", "SCOPE-524"},
	"schemas/semantic/budget.schema.json":                 {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/delegation.schema.json":             {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/guardrail.schema.json":              {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/knowledge.schema.json":              {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/obligation.schema.json":             {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/output.schema.json":                 {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/profile.schema.json":                {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/prompt-ir.schema.json":              {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/taxonomy.schema.json":               {"add", "semantic-owner", "SCOPE-523"},
	"schemas/semantic/workflow.schema.json":               {"add", "semantic-owner", "SCOPE-523"},
	"semantic/profiles/behavior-preserving-refactor.json": {"add", "semantic-owner", "SCOPE-523"},
	"semantic/profiles/feature-change.json":               {"add", "semantic-owner", "SCOPE-523"},
	"semantic/profiles/generic.json":                      {"add", "semantic-owner", "SCOPE-523"},
	"semantic/taxonomy/guardrails.json":                   {"add", "semantic-owner", "SCOPE-523"},
	"semantic/taxonomy/knowledge.json":                    {"add", "semantic-owner", "SCOPE-523"},
	"semantic/taxonomy/obligations.json":                  {"add", "semantic-owner", "SCOPE-523"},
	"semantic/taxonomy/registry.json":                     {"add", "semantic-owner", "SCOPE-523"},
	"semantic/workflows/reference/contract.json":          {"add", "semantic-owner", "SCOPE-523"},
	"semantic/workflows/reference/prompt.md.tmpl":         {"add", "semantic-owner", "SCOPE-523"},
}

var wave02EvaluatorControlPaths = []string{
	"fixtures/public/bl-003/adjudication.json",
	"fixtures/public/bl-003/cases.json",
	"fixtures/public/bl-003/coverage.json",
	"fixtures/public/bl-003/grader-registry.json",
	"fixtures/public/bl-003/protocol.json",
	"fixtures/public/bl-003/truth-labels.json",
	"internal/evaluator/coverage.go",
	"internal/evaluator/coverage_test.go",
	"internal/evaluator/doc.go",
	"internal/evaluator/grade.go",
	"internal/evaluator/grade_test.go",
	"internal/evaluator/model.go",
	"internal/evaluator/validate.go",
	"internal/evaluator/validate_test.go",
	"schemas/evaluation/adjudication.schema.json",
	"schemas/evaluation/case.schema.json",
	"schemas/evaluation/coverage.schema.json",
	"schemas/evaluation/grader.schema.json",
	"schemas/evaluation/protocol.schema.json",
	"schemas/evaluation/run-manifest.schema.json",
	"schemas/evaluation/truth-label.schema.json",
}

var approvedWave02Inputs = map[string]string{
	"docs/artifacts/wave-02-change-contract.md": "367dab50ee994b21eb2503ab7538c9687546d4e55a4275c563a87b80973eaaf4",
	"docs/artifacts/wave-02-specification.md":   "3cb7304e18bf1320160252ac4b74b7321e714728cb5079cb4e24d7e45bc6eb5d",
	"docs/artifacts/wave-02-design.md":          "febff4ba9cdaa17700724004aba7f1edf78cfd52b3f3e42baea2f609d0de5e55",
	"docs/artifacts/wave-01-audit.md":           "491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f",
}

var wave02ApprovalBindings = []string{
	"Artifact ID | `L7-APR-W02-001`",
	"Status | **RECORDED AP0",
	"NO REPLAY**",
	"Accountable owner | Anup Pandey",
	"Authorized implementation | Wave 2 Slices 0, 1, and 2 only",
	"Source commit | `" + wave02BaseCommit + "`",
	"Source tree | `" + wave02BaseTree + "`",
	"Source parent | `b77c4f02a2fcee7af782301699379342e19b7aa3`",
	"Local `main` | `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`",
	"Change contract SHA-256 | `367dab50ee994b21eb2503ab7538c9687546d4e55a4275c563a87b80973eaaf4`",
	"Specification SHA-256 | `3cb7304e18bf1320160252ac4b74b7321e714728cb5079cb4e24d7e45bc6eb5d`",
	"Design SHA-256 | `febff4ba9cdaa17700724004aba7f1edf78cfd52b3f3e42baea2f609d0de5e55`",
	"Wave 1 independent GO audit SHA-256 | `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f`",
}

func checkWave02Admission(root string, base map[string]string, current map[string]snapshotFile, rules map[string]pathExpectation, finalCandidate bool) []finding {
	var findings []finding
	findings = appendFindings(findings, checkWave02Approval(root)...)
	findings = appendFindings(findings, checkWave02EvaluatorFreeze(root, current)...)
	if finalCandidate {
		findings = appendFindings(findings, validateCandidateManifest(root, base, current, rules, candidateClosure{
			manifestPath: wave02CandidateManifest,
			evidencePath: wave02EvidencePath,
			auditPath:    wave02AuditPath,
			expectedRows: 69,
		})...)
	}
	return findings
}

func checkWave02Approval(root string) []finding {
	data, findings := readStrictFile(root, "docs/artifacts/wave-02-approval.md")
	if len(findings) != 0 {
		return findings
	}
	text := string(data)
	for _, binding := range wave02ApprovalBindings {
		if !strings.Contains(text, binding) {
			findings = appendFindings(findings, newFinding("SCOPE-550", "docs/artifacts/wave-02-approval.md", "approval record lost an exact owner, source, artifact, scope, or AP0 binding", "restore the exact non-replayable implementation record"))
			break
		}
	}
	if strings.Contains(text, "AP1") || strings.Contains(text, "authorizes replay") {
		findings = appendFindings(findings, newFinding("SCOPE-551", "docs/artifacts/wave-02-approval.md", "persisted approval text claims replayable authority", "retain AP0 current-conversation-only semantics"))
	}
	return findings
}

func checkWave02EvaluatorFreeze(root string, current map[string]snapshotFile) []finding {
	present := 0
	for _, relative := range wave02EvaluatorControlPaths {
		if current[relative].regular {
			present++
		}
	}
	manifestPresent := current[wave02EvaluatorManifest].regular
	if present == 0 && !manifestPresent {
		return nil
	}
	if present != len(wave02EvaluatorControlPaths) || !manifestPresent {
		return []finding{newFinding("SCOPE-560", wave02EvaluatorManifest, fmt.Sprintf("partial evaluator freeze has %d of %d controls and manifest=%t", present, len(wave02EvaluatorControlPaths), manifestPresent), "land or recover the entire evaluator-control set under separate authority")}
	}
	data, findings := readStrictFile(root, wave02EvaluatorManifest)
	manifest, parseFindings := parseSHA256Manifest(wave02EvaluatorManifest, data, true)
	findings = appendFindings(findings, parseFindings...)
	if len(manifest) != len(wave02EvaluatorControlPaths) {
		findings = appendFindings(findings, newFinding("SCOPE-561", wave02EvaluatorManifest, fmt.Sprintf("evaluator manifest has %d rows, want %d", len(manifest), len(wave02EvaluatorControlPaths)), "freeze exactly the approved evaluator-control paths"))
	}
	expected := make(map[string]bool, len(wave02EvaluatorControlPaths))
	for _, relative := range wave02EvaluatorControlPaths {
		expected[relative] = true
		file := current[relative]
		if manifest[relative] != file.digest {
			findings = appendFindings(findings, newFinding("SCOPE-562", relative, "evaluator-control digest is missing or changed", "restore or separately version the frozen evaluator control"))
		}
	}
	for relative := range manifest {
		if !expected[relative] {
			findings = appendFindings(findings, newFinding("SCOPE-563", relative, "evaluator manifest contains a non-control path", "retain only the exact 21 evaluator controls"))
		}
	}
	return findings
}

func wave02CandidatePresent(root string) bool {
	info, err := os.Lstat(root + "/" + wave02CandidateManifest)
	return err == nil && info.Mode().IsRegular()
}
