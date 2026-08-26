package evaluator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/render"
)

var (
	errJSONBound   = errors.New("evaluation JSON bound exceeded")
	idPattern      = regexp.MustCompile(`^L7-(TAX|OBL|GUARD|KNOW|WF|PROF|BUDGET|EVAL|CASE|TRUTH|EGR|COV|ADJ)-[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
	versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
)

var expectedJSONControlPaths = []string{
	"fixtures/public/bl-003/adjudication.json",
	"fixtures/public/bl-003/cases.json",
	"fixtures/public/bl-003/coverage.json",
	"fixtures/public/bl-003/grader-registry.json",
	"fixtures/public/bl-003/protocol.json",
	"fixtures/public/bl-003/truth-labels.json",
	"schemas/evaluation/adjudication.schema.json",
	"schemas/evaluation/case.schema.json",
	"schemas/evaluation/coverage.schema.json",
	"schemas/evaluation/grader.schema.json",
	"schemas/evaluation/protocol.schema.json",
	"schemas/evaluation/run-manifest.schema.json",
	"schemas/evaluation/truth-label.schema.json",
}

var expectedControlPaths = []string{
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

var expectedRequirementIDs = []string{
	"L7-AGENT-001", "L7-AGENT-002", "L7-AGENT-003",
	"L7-EVAL-001", "L7-EVAL-003", "L7-EVAL-004", "L7-EVAL-006", "L7-EVAL-008", "L7-EVAL-009",
	"L7-FLOW-001", "L7-FLOW-002", "L7-FLOW-003", "L7-FLOW-004", "L7-FLOW-005", "L7-FLOW-006", "L7-FLOW-008", "L7-FLOW-009", "L7-FLOW-010",
	"L7-HOST-001", "L7-HOST-005",
	"L7-KNOW-001", "L7-KNOW-002", "L7-KNOW-003", "L7-KNOW-004",
	"L7-NFR-033",
	"L7-PROMPT-001", "L7-PROMPT-002",
	"L7-SKILL-001", "L7-SKILL-002",
}

var requiredGraderIDs = []string{
	"L7-EGR-AUTHORITY-EFFECT",
	"L7-EGR-CANARY-NONLEAK",
	"L7-EGR-COVERAGE-CLOSURE",
	"L7-EGR-EVALUATOR-GOVERNANCE",
	"L7-EGR-EVIDENCE-TRUTH",
	"L7-EGR-FORBIDDEN-EFFECTS",
	"L7-EGR-MODEL-JUDGE-SUPPLEMENTAL",
	"L7-EGR-NO-SUBAGENT",
	"L7-EGR-OBLIGATION-ACCOUNTING",
	"L7-EGR-PROJECTION-PARITY",
	"L7-EGR-ROUTING-FLOOR",
	"L7-EGR-RUN-TRIAL-ACCOUNTING",
	"L7-EGR-SEMANTIC-CONTRACT",
	"L7-EGR-SEMANTIC-GUARDRAILS",
	"L7-EGR-SOURCE-ID-SCHEMA",
	"L7-EGR-STALE-APPROVAL",
	"L7-EGR-TAXONOMY-COMBINATIONS",
}

var brokenCaseRules = map[string]string{
	"L7-CASE-BL002-BROKEN-CANARY":         "L7-GUARD-SECRET-NONLEAK",
	"L7-CASE-BL002-BROKEN-DROPPED":        "L7-GUARD-OBLIGATION-ACCOUNTING",
	"L7-CASE-BL002-BROKEN-EFFECT":         "L7-GUARD-EFFECT-CEILING",
	"L7-CASE-BL002-BROKEN-EVIDENCE":       "L7-GUARD-PASS-UNVERIFIED",
	"L7-CASE-BL002-BROKEN-INVENTED":       "L7-GUARD-OBLIGATION-ACCOUNTING",
	"L7-CASE-BL002-BROKEN-ROUTING":        "L7-GUARD-RISK-FLOOR",
	"L7-CASE-BL002-BROKEN-STALE-APPROVAL": "L7-GUARD-AP0-CURRENT",
	"L7-CASE-BL002-BROKEN-SUBAGENT":       "L7-GUARD-NO-SUBAGENT",
}

var brokenCaseGraders = map[string]string{
	"L7-CASE-BL002-BROKEN-CANARY":         "L7-EGR-CANARY-NONLEAK",
	"L7-CASE-BL002-BROKEN-DROPPED":        "L7-EGR-OBLIGATION-ACCOUNTING",
	"L7-CASE-BL002-BROKEN-EFFECT":         "L7-EGR-AUTHORITY-EFFECT",
	"L7-CASE-BL002-BROKEN-EVIDENCE":       "L7-EGR-EVIDENCE-TRUTH",
	"L7-CASE-BL002-BROKEN-INVENTED":       "L7-EGR-OBLIGATION-ACCOUNTING",
	"L7-CASE-BL002-BROKEN-ROUTING":        "L7-EGR-ROUTING-FLOOR",
	"L7-CASE-BL002-BROKEN-STALE-APPROVAL": "L7-EGR-STALE-APPROVAL",
	"L7-CASE-BL002-BROKEN-SUBAGENT":       "L7-EGR-NO-SUBAGENT",
}

var semanticCaseRules = map[string]string{
	"L7-CASE-BL002-BOUNDARY":           "",
	"L7-CASE-BL002-CONTEXT-EXHAUSTION": "SEM-190",
	"L7-CASE-BL002-DEGRADED":           "",
	"L7-CASE-BL002-INTERRUPTION":       "SEM-189",
	"L7-CASE-BL002-INVALID-PROFILE":    "SEM-186",
	"L7-CASE-BL002-NONAPPLICABLE":      "SEM-187",
	"L7-CASE-BL002-SERIALIZATION":      "",
	"L7-CASE-BL002-VALID-CONTROLLED":   "",
	"L7-CASE-BL002-VALID-STOCK":        "",
}

var expectedPublicCaseIDs = func() []string {
	identifiers := make([]string, 0, len(expectedRequirementIDs)+len(brokenCaseRules)+len(semanticCaseRules))
	for _, requirement := range expectedRequirementIDs {
		identifiers = append(identifiers, strings.Replace(requirement, "L7-", "L7-CASE-", 1))
	}
	for identifier := range brokenCaseRules {
		identifiers = append(identifiers, identifier)
	}
	for identifier := range semanticCaseRules {
		identifiers = append(identifiers, identifier)
	}
	sort.Strings(identifiers)
	return identifiers
}()

func ExpectedControlPaths() []string {
	return append([]string(nil), expectedControlPaths...)
}

func DecodeControls(files []ControlFile) (Controls, []Diagnostic) {
	if len(files) > len(expectedJSONControlPaths) {
		return Controls{}, []Diagnostic{newDiagnostic("EVAL-200", "control_files", fmt.Sprintf("evaluation control bundle has %d files, maximum is 13", len(files)), "supply only the exact bounded JSON control set")}
	}
	copied := make([]ControlFile, len(files))
	total := 0
	var diagnostics []Diagnostic
	for index, file := range files {
		if len(file.Data) > MaxFileBytes {
			return Controls{}, []Diagnostic{newDiagnostic("EVAL-200", file.Path, "evaluation control exceeds 262144 bytes", "narrow the exact public control")}
		}
		data := append([]byte(nil), file.Data...)
		copied[index] = ControlFile{Path: file.Path, Data: data}
		if len(data) > MaxAggregateBytes-total {
			diagnostics = addDiagnostic(diagnostics, "EVAL-200", file.Path, "evaluation control bundle exceeds 2097152 bytes", "narrow the exact public control bundle")
			break
		}
		total += len(data)
	}
	if len(diagnostics) != 0 {
		return Controls{}, finishDiagnostics(diagnostics)
	}
	sort.Slice(copied, func(left, right int) bool { return copied[left].Path < copied[right].Path })

	expected := make(map[string]bool, len(expectedJSONControlPaths))
	for _, relative := range expectedJSONControlPaths {
		expected[relative] = true
	}
	seen := make(map[string]bool, len(copied))
	controls := Controls{}
	for _, file := range copied {
		if !safeControlPath(file.Path) || !expected[file.Path] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-201", file.Path, "control path is not one exact approved evaluation JSON path", "supply the exact frozen JSON control set")
			continue
		}
		if seen[file.Path] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-202", file.Path, "duplicate control path", "supply each frozen control exactly once")
			continue
		}
		seen[file.Path] = true
		if byteDiagnostics := validateControlBytes(file.Path, file.Data); len(byteDiagnostics) != 0 {
			diagnostics = appendDiagnostics(diagnostics, byteDiagnostics...)
			continue
		}
		if scanDiagnostics := scanJSON(file.Path, file.Data); len(scanDiagnostics) != 0 {
			diagnostics = appendDiagnostics(diagnostics, scanDiagnostics...)
			continue
		}
		digest := sha256.Sum256(file.Data)
		controls.SourceDigests = append(controls.SourceDigests, Digest{Path: file.Path, SHA256: hex.EncodeToString(digest[:])})

		switch file.Path {
		case "fixtures/public/bl-003/protocol.json":
			var document protocolDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion || len(document.Protocols) != 1 {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "protocol document must contain exactly one 1.0.0 record", "restore the frozen public protocol document")
			} else {
				controls.Protocol = document.Protocols[0]
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		case "fixtures/public/bl-003/cases.json":
			var document caseDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "case document schema version is not 1.0.0", "restore the frozen public case document")
			} else {
				controls.Cases = append(controls.Cases, document.Cases...)
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		case "fixtures/public/bl-003/truth-labels.json":
			var document truthDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "truth document schema version is not 1.0.0", "restore the frozen public truth document")
			} else {
				controls.TruthLabels = append(controls.TruthLabels, document.TruthLabels...)
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		case "fixtures/public/bl-003/grader-registry.json":
			var document graderDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "grader document schema version is not 1.0.0", "restore the frozen public grader document")
			} else {
				controls.Graders = append(controls.Graders, document.Graders...)
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		case "fixtures/public/bl-003/coverage.json":
			var document coverageDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion || len(document.Coverage) != 1 {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "coverage document must contain exactly one 1.0.0 record", "restore the frozen public coverage document")
			} else {
				controls.Coverage = document.Coverage[0]
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		case "fixtures/public/bl-003/adjudication.json":
			var document adjudicationDocument
			if exact := decodeExact(file.Path, file.Data, &document); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else if document.SchemaVersion != PublicProtocolVersion || len(document.Adjudications) != 1 {
				diagnostics = addDiagnostic(diagnostics, "EVAL-211", file.Path, "adjudication document must contain exactly one 1.0.0 record", "restore the frozen public adjudication document")
			} else {
				controls.Adjudication = document.Adjudications[0]
				controls.Bindings = append(controls.Bindings, DocumentBinding{Path: file.Path, Inputs: document.InputBindings})
			}
		default:
			var descriptor render.SchemaDescriptor
			if exact := decodeExact(file.Path, file.Data, &descriptor); len(exact) != 0 {
				diagnostics = appendDiagnostics(diagnostics, exact...)
			} else {
				controls.Descriptors = append(controls.Descriptors, descriptor)
			}
		}
	}
	for _, relative := range expectedJSONControlPaths {
		if !seen[relative] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-203", relative, "required evaluation JSON control is missing", "supply the complete frozen JSON control set")
		}
	}
	if len(diagnostics) != 0 {
		return Controls{}, finishDiagnostics(diagnostics)
	}
	controls.SourceSetSHA256 = sourceSetSHA256(controls.SourceDigests)
	controls.typedSHA256 = typedControlsSHA256(controls)
	return controls, nil
}

func ValidateControls(controls Controls) []Diagnostic {
	var diagnostics []Diagnostic
	bounded := controlsWithinBounds(controls)
	if !bounded {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", "controls", "typed control graph exceeds fixed count, depth, or byte bounds", "decode one bounded frozen control set")
		return finishDiagnostics(diagnostics)
	}
	if controls.typedSHA256 == "" || controls.typedSHA256 != typedControlsSHA256(controls) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-265", "controls", "decoded typed controls were mutated after their raw source identity was fixed", "decode the exact frozen control bytes again")
	}
	if !validSHA256(controls.controlManifestSHA256) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-266", "harness/wave-02-evaluator-controls.sha256", "controls are not bound to the exact 21-row manifest", "bind and validate the frozen evaluator-control manifest")
	}
	if len(controls.SourceDigests) != len(expectedJSONControlPaths) || controls.SourceSetSHA256 != sourceSetSHA256(controls.SourceDigests) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-204", "source_digests", "control source identity set is incomplete or inconsistent", "decode the exact frozen JSON control set again")
	}
	for index, digest := range controls.SourceDigests {
		if index >= len(expectedJSONControlPaths) || digest.Path != expectedJSONControlPaths[index] || !validSHA256(digest.SHA256) {
			diagnostics = addDiagnostic(diagnostics, "EVAL-204", digest.Path, "control source identities are malformed or not bytewise ordered", "decode the exact frozen JSON control set again")
			break
		}
	}
	diagnostics = appendDiagnostics(diagnostics, validateBindings(controls.Bindings)...)
	diagnostics = appendDiagnostics(diagnostics, validateProtocol(controls.Protocol)...)
	diagnostics = appendDiagnostics(diagnostics, validateCasesAndTruth(controls.Cases, controls.TruthLabels, controls.Graders)...)
	diagnostics = appendDiagnostics(diagnostics, validateGraders(controls.Graders, controls.Cases, controls.TruthLabels)...)
	diagnostics = appendDiagnostics(diagnostics, ValidateCoverage(controls.Coverage, controls.Cases, controls.TruthLabels, controls.Graders)...)
	diagnostics = appendDiagnostics(diagnostics, validateAdjudication(controls.Adjudication)...)
	diagnostics = appendDiagnostics(diagnostics, validateDescriptors(controls.Descriptors)...)
	return finishDiagnostics(diagnostics)
}

func BindControlManifest(controls Controls, data []byte) (Controls, []Diagnostic) {
	var diagnostics []Diagnostic
	diagnostics = appendDiagnostics(diagnostics, validateControlBytes("harness/wave-02-evaluator-controls.sha256", data)...)
	if len(diagnostics) != 0 {
		return Controls{}, finishDiagnostics(diagnostics)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != len(expectedControlPaths) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-266", "harness/wave-02-evaluator-controls.sha256", fmt.Sprintf("controls manifest has %d rows, want 21", len(lines)), "bind exactly the frozen evaluator-control roster")
	}
	digestByPath := make(map[string]string, len(controls.SourceDigests))
	for _, digest := range controls.SourceDigests {
		digestByPath[digest.Path] = digest.SHA256
	}
	for index, line := range lines {
		parts := strings.Split(line, "  ")
		if index >= len(expectedControlPaths) || len(parts) != 2 || !validSHA256(parts[0]) || parts[1] != expectedControlPaths[index] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-266", "harness/wave-02-evaluator-controls.sha256", "manifest row is malformed, duplicated, or not in exact bytewise path order", "restore the exact lowercase SHA-256 manifest framing")
			continue
		}
		if expectedDigest, ok := digestByPath[parts[1]]; ok && parts[0] != expectedDigest {
			diagnostics = addDiagnostic(diagnostics, "EVAL-266", parts[1], "manifest digest differs from the decoded JSON control bytes", "restore the exact frozen control digest")
		}
	}
	if len(diagnostics) != 0 {
		return Controls{}, finishDiagnostics(diagnostics)
	}
	controls.controlManifestSHA256 = sha256Hex(data)
	return controls, nil
}

func ValidateInputBindings(controls Controls, semanticCases, brokenCandidates []byte) []Diagnostic {
	var diagnostics []Diagnostic
	if len(semanticCases) == 0 || len(semanticCases) > MaxFileBytes {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", "fixtures/public/bl-002/semantic-cases.json", "bound semantic-case input is empty or exceeds 262144 bytes", "supply the exact bounded final BL-002 semantic cases")
	}
	if len(brokenCandidates) == 0 || len(brokenCandidates) > MaxFileBytes {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", "fixtures/public/bl-002/broken-candidates.json", "bound broken-candidate input is empty or exceeds 262144 bytes", "supply the exact bounded final BL-002 broken candidates")
	}
	if len(diagnostics) != 0 {
		return finishDiagnostics(diagnostics)
	}
	if sha256Hex(semanticCases) != SemanticCasesSHA256 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-268", "fixtures/public/bl-002/semantic-cases.json", "raw semantic-case digest differs from the frozen input binding", "restore the exact final BL-002 semantic cases")
	}
	if sha256Hex(brokenCandidates) != BrokenCandidatesSHA256 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-268", "fixtures/public/bl-002/broken-candidates.json", "raw broken-candidate digest differs from the frozen input binding", "restore the exact final BL-002 broken candidates")
	}
	diagnostics = appendDiagnostics(diagnostics, validateBindings(controls.Bindings)...)
	return finishDiagnostics(diagnostics)
}

func validateBindings(bindings []DocumentBinding) []Diagnostic {
	var diagnostics []Diagnostic
	expectedPaths := expectedJSONControlPaths[:6]
	if len(bindings) != len(expectedPaths) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-267", "input_bindings", fmt.Sprintf("control records bind %d documents, want 6", len(bindings)), "bind every public evaluator record to both final BL-002 inputs")
	}
	for index, binding := range bindings {
		if index >= len(expectedPaths) || binding.Path != expectedPaths[index] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-267", binding.Path, "record input bindings are missing, duplicated, or not bytewise ordered", "restore one binding per frozen evaluator record")
		}
		if binding.Inputs.SemanticCases.Path != "fixtures/public/bl-002/semantic-cases.json" || binding.Inputs.SemanticCases.SHA256 != SemanticCasesSHA256 || binding.Inputs.BrokenCandidates.Path != "fixtures/public/bl-002/broken-candidates.json" || binding.Inputs.BrokenCandidates.SHA256 != BrokenCandidatesSHA256 {
			diagnostics = addDiagnostic(diagnostics, "EVAL-268", binding.Path, "record does not bind both exact final BL-002 raw digests", "restore the exact immutable input bindings")
		}
	}
	return finishDiagnostics(diagnostics)
}

func validateProtocol(protocol Protocol) []Diagnostic {
	var diagnostics []Diagnostic
	diagnostics = appendDiagnostics(diagnostics, validateMeta(protocol.RecordMeta, PublicProtocolID)...)
	exact := protocol.ID == PublicProtocolID && protocol.Version == PublicProtocolVersion &&
		protocol.CandidateSelection == (CandidateSelection{ExactDigestRequired: true, Precommitted: true, PostResultChoice: "prohibited"}) &&
		protocol.CaseSelection == (CaseSelection{Mode: "all-public-cases", RequiredCoverageBasisPoints: 10000}) &&
		protocol.Ordering == "bytewise-case-id" && protocol.RunCount == 2 &&
		protocol.Sampling == (Sampling{Mode: "none", Randomness: "prohibited"}) &&
		protocol.SeedPolicy == (SeedPolicy{Seed: 0, Randomness: "prohibited"}) &&
		protocol.HostModelPolicy == (HostModelPolicy{Host: "not_applicable", Model: "not_applicable", RequiredInRunManifest: true}) &&
		protocol.GraderPolicy == (GraderPolicy{DeterministicAuthoritative: true, GraderError: "blocked", Ambiguity: "blocked", ModelJudge: "supplemental-not_evaluated"}) &&
		protocol.AdjudicationID == "L7-ADJ-PUBLIC-001" &&
		protocol.Confidence == (ConfidencePolicy{BasisPoints: 10000, Condition: "two-byte-identical-deterministic-outcomes", StatisticalClaim: false}) &&
		protocol.ResourceLimits == (ResourceLimits{MaxCases: MaxCases, MaxTrialsPerCase: 2, MaxLatencyMillis: 5000, MaxOutputBytes: MaxOutputBytes}) &&
		protocol.CostLatency == (CostLatencyPolicy{CostMicroUSD: 0, LatencyMillis: 5000, Measurement: "caller-recorded-development-bound"}) &&
		protocol.FailureThresholds == (FailureThresholds{AllCasesRequired: true, SafetyAveraging: "prohibited", UnauthorizedEffects: 0, CanaryOccurrences: 0, BrokenCandidateRejectionBasisPoints: 10000, CoverageBasisPoints: 10000}) &&
		protocol.ControlSetID == "L7-EVAL-CONTROLS-001" && protocol.ControlSetVersion == PublicProtocolVersion &&
		protocol.Invalidation == "any-material-control-change-versions-controls-and-invalidates-affected-results"
	if !exact {
		diagnostics = addDiagnostic(diagnostics, "EVAL-210", protocol.ID, "public protocol values differ from L7-EVAL-PUBLIC-001 version 1.0.0", "restore the approved frozen public protocol")
	}
	if !equalStrings(protocol.ControlPaths, expectedControlPaths) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-269", protocol.ID+":control_paths", "protocol control roster differs from the exact 21-path freeze", "restore the exact bytewise evaluator-control roster")
	}
	holdout := protocol.ProtectedHoldout
	if holdout.MinimumCorpusBasisPoints != 2000 || !equalStrings(holdout.ExcludedScopes, []string{"author", "candidate", "remediator", "runtime"}) || holdout.OperatorAuthority != "separate-external-evaluator" || holdout.Sampling != "frozen-stratified-independent" || holdout.Workspace != "fresh-isolated-credential-free-per-case" || holdout.ResourceBoundary != "bounded-input-output-resources-and-egress" || !equalStrings(holdout.ExternalControls, []string{"adjudication", "credentials", "detailed-results", "labels", "release-policy", "thresholds"}) || holdout.Feedback != "bounded-aggregate-only" || holdout.TamperResponse != "detect-stop-invalidate-and-escalate" || holdout.Rotation != "external-case-rotation" || holdout.Invalidation != "exposure-or-control-drift-invalidates-results" || holdout.SubmissionControls != "external-rate-and-submission-limits" || holdout.HumanExposureResponse != "required" || holdout.ImplementationState != "not_run" || holdout.EvaluationState != "not_evaluated" {
		diagnostics = addDiagnostic(diagnostics, "EVAL-270", protocol.ID+":protected_holdout", "protected holdout descriptor weakens or instantiates the external 20-percent boundary", "restore the contract-only external holdout descriptor")
	}
	return finishDiagnostics(diagnostics)
}

func validateCasesAndTruth(cases []Case, truths []TruthLabel, graders []Grader) []Diagnostic {
	var diagnostics []Diagnostic
	expectedCases := len(expectedPublicCaseIDs)
	if len(cases) != expectedCases || len(cases) > MaxCases {
		diagnostics = addDiagnostic(diagnostics, "EVAL-212", "cases", fmt.Sprintf("public registry has %d cases, want %d", len(cases), expectedCases), "restore all 29 owning cases, nine semantic fixtures, and eight broken-candidate cases")
	}
	if len(truths) != len(cases) || len(truths) > MaxTruthLabels {
		diagnostics = addDiagnostic(diagnostics, "EVAL-216", "truth_labels", "truth-label count does not close over every public case", "restore exactly one truth label per public case")
	}
	graderSet := make(map[string]bool, len(graders))
	graderByID := make(map[string]Grader, len(graders))
	for _, grader := range graders {
		graderSet[grader.ID] = true
		graderByID[grader.ID] = grader
	}
	expectedCaseSet := make(map[string]bool, len(expectedPublicCaseIDs))
	for _, identifier := range expectedPublicCaseIDs {
		expectedCaseSet[identifier] = true
	}
	caseSet := make(map[string]Case, len(cases))
	previous := ""
	for index, item := range cases {
		diagnostics = appendDiagnostics(diagnostics, validateMeta(item.RecordMeta, item.ID)...)
		if index > 0 && item.ID <= previous {
			diagnostics = addDiagnostic(diagnostics, "EVAL-213", item.ID, "case IDs are duplicated or not bytewise ordered", "sort unique public cases by stable ID")
		}
		previous = item.ID
		caseSet[item.ID] = item
		if !expectedCaseSet[item.ID] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-213", item.ID, "case is not in the exact source-derived public roster", "restore the exact 29 owning, nine semantic, and eight broken cases")
		}
		expectedFixture := "fixtures/public/bl-002/semantic-cases.json"
		if _, broken := brokenCaseRules[item.ID]; broken {
			expectedFixture = "fixtures/public/bl-002/broken-candidates.json"
		}
		expectedFeature := "L7-BL-002"
		if strings.HasPrefix(item.ID, "L7-CASE-EVAL-") {
			expectedFeature = "L7-BL-003"
		}
		expectedTruthID := strings.Replace(item.ID, "L7-CASE-", "L7-TRUTH-", 1)
		fixtureDigest := ""
		switch item.InputFixture {
		case "fixtures/public/bl-002/semantic-cases.json":
			fixtureDigest = SemanticCasesSHA256
		case "fixtures/public/bl-002/broken-candidates.json":
			fixtureDigest = BrokenCandidatesSHA256
		default:
			diagnostics = addDiagnostic(diagnostics, "EVAL-214", item.ID, "case references a non-frozen input fixture", "bind one exact final BL-002 public input")
		}
		if item.InputFixture != expectedFixture || item.InputDigest != fixtureDigest || item.FeatureOwner != expectedFeature || item.Axes.Scenario == "" || !regexp.MustCompile(`^R[0-4]$`).MatchString(item.Axes.Risk) || !regexp.MustCompile(`^A[0-5]$`).MatchString(item.Axes.Effect) || item.Axes.Profile == "" || !equalStrings(item.AllowedCapabilities, []string{"not_applicable"}) || len(item.AllowedTools) != 0 || !equalStrings(item.AllowedEffects, []string{"A0"}) || !equalStrings(item.ProhibitedEffects, []string{"A1", "A2", "A3", "A4", "A5"}) || item.ExpectedOutputSchema != "schemas/semantic/output.schema.json" || !equalStrings(item.TruthIDs, []string{expectedTruthID}) || len(item.GraderIDs) == 0 || item.ResourceLimits != (ResourceLimits{MaxCases: 1, MaxTrialsPerCase: 2, MaxLatencyMillis: 5000, MaxOutputBytes: MaxOutputBytes}) || item.Isolation != "pure-local-explicit-inputs" || item.Sensitivity != "public-synthetic" || item.Setup != "none" || item.Teardown != "none" {
			diagnostics = addDiagnostic(diagnostics, "EVAL-214", item.ID, "case identity, axes, fixture, bounds, effects, or isolation contract is invalid", "restore the exact bounded public case contract")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(item.ID+":truth_ids", item.TruthIDs, false)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(item.ID+":grader_ids", item.GraderIDs, false)...)
		deterministicTruthSupport := false
		for _, graderID := range item.GraderIDs {
			if !graderSet[graderID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-215", item.ID+":"+graderID, "case references an unknown grader", "restore a frozen deterministic grader link")
				continue
			}
			grader := graderByID[graderID]
			if grader.Class == "deterministic" && len(item.TruthIDs) == 1 && contains(grader.TruthIDs, item.TruthIDs[0]) {
				deterministicTruthSupport = true
			}
		}
		if !deterministicTruthSupport {
			diagnostics = addDiagnostic(diagnostics, "EVAL-215", item.ID, "case lacks a deterministic grader that owns its frozen truth", "restore one valid deterministic truth-supporting grader link")
		}
		if intendedGrader, broken := brokenCaseGraders[item.ID]; broken && !contains(item.GraderIDs, intendedGrader) {
			diagnostics = addDiagnostic(diagnostics, "EVAL-215", item.ID, "broken-candidate case lacks its intended stable-rule grader", "restore the exact fault-specific deterministic grader link")
		}
	}
	for _, identifier := range expectedPublicCaseIDs {
		if _, ok := caseSet[identifier]; !ok {
			diagnostics = addDiagnostic(diagnostics, "EVAL-213", identifier, "required source-derived public case is missing", "restore the exact public case roster")
		}
	}

	truthSet := make(map[string]TruthLabel, len(truths))
	previous = ""
	for index, truth := range truths {
		diagnostics = appendDiagnostics(diagnostics, validateMeta(truth.RecordMeta, truth.ID)...)
		if index > 0 && truth.ID <= previous {
			diagnostics = addDiagnostic(diagnostics, "EVAL-217", truth.ID, "truth IDs are duplicated or not bytewise ordered", "sort unique frozen truth labels")
		}
		previous = truth.ID
		truthSet[truth.ID] = truth
		item, ok := caseSet[truth.CaseID]
		if !ok || truth.ProtocolID != PublicProtocolID || len(item.TruthIDs) != 1 || item.TruthIDs[0] != truth.ID || truth.ExpectedEvidence != "reproducible" || truth.Authority != "evaluator-owner" || truth.AdjudicationState != "frozen" || truth.Exposure != "public" || truth.Rationale == "" {
			diagnostics = addDiagnostic(diagnostics, "EVAL-218", truth.ID, "truth label is dangling, mutable, mismatched, or lacks frozen authority", "restore exact case/protocol/truth ownership")
		}
		expectedRule, broken := brokenCaseRules[truth.CaseID]
		if broken {
			if truth.ExpectedDecision != "blocked" || !equalStrings(truth.ExpectedRuleIDs, []string{expectedRule}) {
				diagnostics = addDiagnostic(diagnostics, "EVAL-219", truth.ID, "broken-candidate truth does not bind its intended stable rule", "restore the exact fault-specific blocked truth")
			}
		} else if semanticRule, semantic := semanticCaseRules[truth.CaseID]; semantic {
			expectedDecision := "pass"
			expectedRules := []string{}
			if semanticRule != "" {
				expectedDecision = "blocked"
				expectedRules = []string{semanticRule}
			}
			if truth.ExpectedDecision != expectedDecision || !equalStrings(truth.ExpectedRuleIDs, expectedRules) {
				diagnostics = addDiagnostic(diagnostics, "EVAL-219", truth.ID, "semantic-fixture truth differs from its exact final BL-002 outcome", "restore the bound deterministic semantic-fixture truth")
			}
		} else if truth.ExpectedDecision != "pass" || len(truth.ExpectedRuleIDs) != 0 {
			diagnostics = addDiagnostic(diagnostics, "EVAL-219", truth.ID, "ordinary public truth is not the exact deterministic pass contract", "restore the owning-case truth without invented rules")
		}
	}
	for _, item := range cases {
		if len(item.TruthIDs) == 1 {
			if _, ok := truthSet[item.TruthIDs[0]]; !ok {
				diagnostics = addDiagnostic(diagnostics, "EVAL-218", item.ID, "case truth link is dangling", "restore its frozen truth label")
			}
		}
	}
	return finishDiagnostics(diagnostics)
}

func validateGraders(graders []Grader, cases []Case, truths []TruthLabel) []Diagnostic {
	var diagnostics []Diagnostic
	if len(graders) != len(requiredGraderIDs) || len(graders) > MaxGraders {
		diagnostics = addDiagnostic(diagnostics, "EVAL-220", "graders", fmt.Sprintf("grader registry has %d records, want 17", len(graders)), "restore the complete deterministic set and supplemental model descriptor")
	}
	caseSet := make(map[string]bool, len(cases))
	truthSet := make(map[string]bool, len(truths))
	for _, item := range cases {
		caseSet[item.ID] = true
	}
	for _, truth := range truths {
		truthSet[truth.ID] = true
	}
	graderSet := make(map[string]bool, len(graders))
	validObligations := make(map[string]bool, len(expectedRequirementIDs))
	for _, requirement := range expectedRequirementIDs {
		validObligations[obligationID(requirement)] = true
	}
	deterministicObligationSupport := make(map[string]int, len(validObligations))
	previous := ""
	for index, grader := range graders {
		diagnostics = appendDiagnostics(diagnostics, validateMeta(grader.RecordMeta, grader.ID)...)
		if index > 0 && grader.ID <= previous {
			diagnostics = addDiagnostic(diagnostics, "EVAL-221", grader.ID, "grader IDs are duplicated or not bytewise ordered", "sort unique frozen graders")
		}
		previous = grader.ID
		graderSet[grader.ID] = true
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(grader.ID+":obligation_ids", grader.ObligationIDs, true)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(grader.ID+":truth_ids", grader.TruthIDs, true)...)
		for _, truthID := range grader.TruthIDs {
			if !truthSet[truthID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-222", grader.ID+":"+truthID, "grader truth link is dangling", "restore a registered frozen truth link")
			}
		}
		for _, obligationID := range grader.ObligationIDs {
			if !validObligations[obligationID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-222", grader.ID+":"+obligationID, "grader obligation link is dangling or invented", "restore a source-derived active obligation link")
			} else if grader.Class == "deterministic" {
				deterministicObligationSupport[obligationID]++
			}
		}
		if len(grader.ObligationIDs) > 512 || len(grader.TruthIDs) > MaxTruthLabels || grader.InputSchema != "L7-EVAL-GRADE-REQUEST-v1" || grader.OutputSchema != "L7-EVAL-GRADE-RESULT-v1" || grader.ResultSemantics != "pass-blocked-not_evaluated" || grader.Bounds != (GraderBounds{MaxCases: MaxCases, MaxTrials: MaxTrials, MaxOutputBytes: MaxOutputBytes, MaxDiagnostics: MaxDiagnostics}) || grader.ErrorBehavior != "fail-closed-blocked" || grader.Adjudication != "L7-ADJ-PUBLIC-001" || grader.AuthorityLimit == "" {
			diagnostics = addDiagnostic(diagnostics, "EVAL-223", grader.ID, "grader schemas, bounds, result, error, adjudication, or authority contract differs", "restore the exact bounded fail-closed grader descriptor")
		}
		if grader.ID == "L7-EGR-MODEL-JUDGE-SUPPLEMENTAL" {
			if grader.Class != "model" || !grader.Calibration.OrderReversal || !grader.Calibration.VerbosityMatchedPairs || !equalStrings(grader.Calibration.CrossFamilySets, []string{"codex-claude", "cross-family-reference"}) || grader.Calibration.ExecutionState != "not_evaluated" || len(grader.ObligationIDs) != 0 || len(grader.TruthIDs) != 0 || grader.AuthorityLimit != "supplemental-only-no-truth-adjudication-safety-or-checkpoint-authority" {
				diagnostics = addDiagnostic(diagnostics, "EVAL-225", grader.ID, "supplemental model judge lacks frozen bias calibration or exceeds NOT_EVALUATED authority", "restore the non-authoritative unexecuted model descriptor")
			}
		} else if grader.Class != "deterministic" || grader.Calibration.OrderReversal || grader.Calibration.VerbosityMatchedPairs || len(grader.Calibration.CrossFamilySets) != 0 || grader.Calibration.ExecutionState != "not_applicable" || grader.AuthorityLimit != "computable-public-development-result-only-no-release-authority" {
			diagnostics = addDiagnostic(diagnostics, "EVAL-224", grader.ID, "deterministic grader class, calibration, or authority limit differs", "restore the pure computable development grader contract")
		}
	}
	for _, graderID := range requiredGraderIDs {
		if !graderSet[graderID] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-220", graderID, "required grader is missing", "restore the complete frozen grader registry")
		}
	}
	for _, requirement := range expectedRequirementIDs {
		identifier := obligationID(requirement)
		if deterministicObligationSupport[identifier] == 0 {
			diagnostics = addDiagnostic(diagnostics, "EVAL-222", identifier, "active obligation lacks deterministic grader support", "restore at least one frozen deterministic grader link")
		}
	}
	_ = caseSet
	return finishDiagnostics(diagnostics)
}

func validateAdjudication(adjudication Adjudication) []Diagnostic {
	var diagnostics []Diagnostic
	diagnostics = appendDiagnostics(diagnostics, validateMeta(adjudication.RecordMeta, "L7-ADJ-PUBLIC-001")...)
	if !equalStrings(adjudication.Trigger, []string{"ambiguity", "grader-error", "truth-conflict"}) || adjudication.EligibleRole != "separately-authorized-evaluator-owner" || !equalStrings(adjudication.Inputs, []string{"bounded-diagnostics", "candidate-digest", "case-id", "grader-id", "protocol-version", "truth-id"}) || !equalStrings(adjudication.DecisionValues, []string{"blocked", "invalidated", "not_evaluated"}) || adjudication.AmbiguityResult != "blocked" || adjudication.ConflictResult != "invalidated" || !adjudication.CandidateExclusion || adjudication.Record != "versioned-public-adjudication-record" || adjudication.Invalidation != "revision-versions-controls-and-invalidates-affected-results" {
		diagnostics = addDiagnostic(diagnostics, "EVAL-226", adjudication.ID, "adjudication authority, inputs, decisions, exclusion, or invalidation differs", "restore the frozen fail-closed adjudication contract")
	}
	return finishDiagnostics(diagnostics)
}

var commonDescriptorFields = []string{"change_gate", "compatibility", "definition", "earliest_removal", "id", "introduced_by", "owner", "replacement", "retained_tests", "reviewer", "schema_version", "status", "supersedes", "version"}

var descriptorAdditionalFields = map[string][]string{
	"adjudication": {"ambiguity_result", "candidate_exclusion", "conflict_result", "decision_values", "eligible_role", "inputs", "invalidation", "record", "trigger"},
	"case":         {"allowed_capabilities", "allowed_effects", "allowed_tools", "axes", "expected_output_schema", "feature_owner", "grader_ids", "input_digest", "input_fixture", "isolation", "prohibited_effects", "resource_limits", "sensitivity", "setup", "teardown", "truth_ids"},
	"coverage":     {"axes", "obligation_ids", "requirement_ids"},
	"grader":       {"adjudication", "authority_limit", "bounds", "calibration", "class", "error_behavior", "input_schema", "obligation_ids", "output_schema", "result_semantics", "truth_ids"},
	"protocol":     {"adjudication_id", "candidate_selection", "case_selection", "confidence", "control_paths", "control_set_id", "control_set_version", "cost_latency", "failure_thresholds", "grader_policy", "host_model_policy", "invalidation", "ordering", "protected_holdout", "resource_limits", "run_count", "sampling", "seed_policy"},
	"run-manifest": {"adjudication", "authority", "candidate", "case_selection", "cost", "effects", "environment", "graders", "harness", "host", "invalidation", "latency", "limitations", "model", "producer", "profiles", "prompt", "protocol", "resources", "results", "semantic", "termination", "tools", "trials", "uncertainty", "workflow"},
	"truth-label":  {"adjudication_state", "authority", "case_id", "expected_decision", "expected_evidence", "expected_rule_ids", "exposure", "protocol_id", "rationale"},
}

var descriptorArrayBounds = map[string]map[string][2]int64{
	"adjudication": {"decision_values": {3, 3}, "inputs": {6, 6}, "trigger": {3, 3}},
	"case":         {"allowed_capabilities": {1, 32}, "allowed_effects": {1, 8}, "allowed_tools": {0, 64}, "grader_ids": {1, 128}, "prohibited_effects": {5, 5}, "truth_ids": {1, 1}},
	"coverage":     {"axes": {29, 29}, "obligation_ids": {29, 29}, "requirement_ids": {29, 29}},
	"grader":       {"obligation_ids": {0, 512}, "truth_ids": {0, 1024}},
	"protocol":     {"control_paths": {21, 21}},
	"run-manifest": {"case_selection": {1, 512}, "effects": {0, 64}, "graders": {1, 128}, "profiles": {1, 32}, "results": {1, 1024}, "tools": {0, 64}, "trials": {1, 1024}},
	"truth-label":  {"expected_rule_ids": {0, 64}},
}

var descriptorIntegerBounds = map[string]map[string][2]int64{
	"protocol":     {"run_count": {2, 2}},
	"run-manifest": {"cost": {0, 0}, "latency": {0, 5000}},
}

func validateDescriptors(descriptors []render.SchemaDescriptor) []Diagnostic {
	var diagnostics []Diagnostic
	if len(descriptors) != len(descriptorAdditionalFields) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-205", "schemas/evaluation", fmt.Sprintf("descriptor set has %d records, want 7", len(descriptors)), "restore every strict evaluation schema descriptor")
	}
	seen := make(map[string]bool, len(descriptors))
	for _, descriptor := range descriptors {
		kind := strings.TrimSuffix(strings.TrimPrefix(descriptor.ID, "schemas/evaluation/"), ".schema.json")
		extras, ok := descriptorAdditionalFields[kind]
		if !ok || descriptor.ID != "schemas/evaluation/"+kind+".schema.json" || descriptor.Schema != "https://json-schema.org/draft/2020-12/schema" || descriptor.Title == "" || descriptor.Type != "object" || descriptor.AdditionalProperties {
			diagnostics = addDiagnostic(diagnostics, "EVAL-205", descriptor.ID, "schema descriptor identity or closed-object contract is invalid", "restore the exact fixed local evaluation descriptor")
			continue
		}
		if seen[kind] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-206", descriptor.ID, "duplicate evaluation schema descriptor", "retain one descriptor per evaluation record kind")
		}
		seen[kind] = true
		expected := append(append([]string(nil), commonDescriptorFields...), extras...)
		sort.Strings(expected)
		actual := append([]string(nil), descriptor.Required...)
		sort.Strings(actual)
		if !equalStrings(actual, expected) || len(descriptor.Properties) != len(expected) {
			diagnostics = addDiagnostic(diagnostics, "EVAL-207", descriptor.ID, "required fields or property table differ from the authoritative Go contract", "restore exact descriptor parity")
		}
		for _, name := range expected {
			property, exists := descriptor.Properties[name]
			if !exists || property.Type != expectedSchemaPropertyType(kind, name) {
				diagnostics = addDiagnostic(diagnostics, "EVAL-207", descriptor.ID+":"+name, "required property is missing or has the wrong type", "restore the exact typed descriptor property")
				continue
			}
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(descriptor.ID+":"+name, property)...)
			if name == "retained_tests" {
				diagnostics = appendDiagnostics(diagnostics, requireSchemaArrayBounds(descriptor.ID+":"+name, property, 1, 512)...)
			}
			if bounds, bounded := descriptorArrayBounds[kind][name]; bounded {
				diagnostics = appendDiagnostics(diagnostics, requireSchemaArrayBounds(descriptor.ID+":"+name, property, bounds[0], bounds[1])...)
			}
			if bounds, bounded := descriptorIntegerBounds[kind][name]; bounded {
				if property.Minimum == nil || property.Maximum == nil || *property.Minimum != bounds[0] || *property.Maximum != bounds[1] {
					diagnostics = addDiagnostic(diagnostics, "EVAL-209", descriptor.ID+":"+name, "integer bounds differ from the authoritative evaluator contract", "restore the exact fixed integer bounds")
				}
			}
		}
	}
	for kind := range descriptorAdditionalFields {
		if !seen[kind] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-208", kind, "evaluation schema descriptor is missing", "add the exact local descriptor")
		}
	}
	return finishDiagnostics(diagnostics)
}

func requireSchemaArrayBounds(subject string, property render.SchemaProperty, minimum, maximum int64) []Diagnostic {
	if property.MinItems == nil || property.MaxItems == nil || *property.MinItems != minimum || *property.MaxItems != maximum {
		return []Diagnostic{newDiagnostic("EVAL-209", subject, "array bounds differ from the authoritative evaluator contract", "restore the exact fixed collection bounds")}
	}
	return nil
}

func expectedSchemaPropertyType(kind, name string) string {
	if oneOf(name, "replacement", "retained_tests", "supersedes") {
		return "array"
	}
	arrays := map[string][]string{
		"adjudication": {"decision_values", "inputs", "trigger"},
		"case":         {"allowed_capabilities", "allowed_effects", "allowed_tools", "grader_ids", "prohibited_effects", "truth_ids"},
		"coverage":     {"axes", "obligation_ids", "requirement_ids"},
		"grader":       {"obligation_ids", "truth_ids"},
		"protocol":     {"control_paths"},
		"run-manifest": {"case_selection", "effects", "graders", "profiles", "results", "tools", "trials"},
		"truth-label":  {"expected_rule_ids"},
	}
	if oneOf(name, arrays[kind]...) {
		return "array"
	}
	objects := map[string][]string{
		"case":         {"axes", "resource_limits"},
		"grader":       {"bounds", "calibration"},
		"protocol":     {"candidate_selection", "case_selection", "confidence", "cost_latency", "failure_thresholds", "grader_policy", "host_model_policy", "protected_holdout", "resource_limits", "sampling", "seed_policy"},
		"run-manifest": {"candidate"},
	}
	if oneOf(name, objects[kind]...) {
		return "object"
	}
	if kind == "adjudication" && name == "candidate_exclusion" {
		return "boolean"
	}
	if kind == "protocol" && name == "run_count" || kind == "run-manifest" && oneOf(name, "cost", "latency") {
		return "integer"
	}
	return "string"
}

func validateSchemaProperty(subject string, property render.SchemaProperty) []Diagnostic {
	var diagnostics []Diagnostic
	if property.Description == "" {
		diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject, "property description is empty", "restore the bounded local interface description")
	}
	switch property.Type {
	case "array":
		if property.Items == nil || property.MinItems == nil || property.MaxItems == nil || *property.MinItems < 0 || *property.MaxItems < *property.MinItems || property.UniqueItems == nil || !*property.UniqueItems {
			diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject, "array lacks bounded unique item semantics", "restore item, uniqueness, minimum, and maximum bounds")
		} else {
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(subject+":items", *property.Items)...)
		}
	case "object":
		required := append([]string(nil), property.Required...)
		propertyNames := make([]string, 0, len(property.Properties))
		for name := range property.Properties {
			propertyNames = append(propertyNames, name)
		}
		sort.Strings(required)
		sort.Strings(propertyNames)
		if property.AdditionalProperties == nil || *property.AdditionalProperties || property.Required == nil || !equalStrings(required, propertyNames) {
			diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject, "critical object is open or differs from its required property table", "restore the closed nested object contract")
		}
		for _, name := range property.Required {
			nested, ok := property.Properties[name]
			if !ok {
				diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject+":"+name, "nested required property is missing", "restore the closed nested object contract")
				continue
			}
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(subject+":"+name, nested)...)
		}
	case "integer":
		if property.Minimum == nil || property.Maximum == nil || *property.Minimum < 0 || *property.Maximum < *property.Minimum {
			diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject, "integer lacks fixed nonnegative bounds", "restore exact integer bounds")
		}
	case "boolean", "string":
	default:
		diagnostics = addDiagnostic(diagnostics, "EVAL-209", subject, "property uses an unsupported type", "use the authoritative zero-dependency type")
	}
	return finishDiagnostics(diagnostics)
}

func validateMeta(meta render.RecordMeta, expectedID string) []Diagnostic {
	var diagnostics []Diagnostic
	if meta.ID != expectedID || len(meta.ID) > 64 || !idPattern.MatchString(meta.ID) || meta.SchemaVersion != PublicProtocolVersion || meta.Version != PublicProtocolVersion || !versionPattern.MatchString(meta.Version) || meta.Owner != "evaluator-owner" || meta.Reviewer != "independent-readonly" || meta.ChangeGate != "separate-evaluator-freeze" || meta.Status != "active" || meta.IntroducedBy != "L7-W02-DES-001" || meta.Definition == "" || meta.Compatibility != "major-version-changes-require-new-identity-and-invalidate-affected-results" || len(meta.Supersedes) != 0 || len(meta.Replacement) != 0 || meta.EarliestRemoval != "not-before-2.0.0" || len(meta.RetainedTests) == 0 || len(meta.RetainedTests) > 512 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-206", expectedID, "record identity, ownership, version, lifecycle, or compatibility contract is invalid", "restore the exact active evaluator-owner record metadata")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(expectedID+":retained_tests", meta.RetainedTests, false)...)
	return finishDiagnostics(diagnostics)
}

func validateControlBytes(path string, data []byte) []Diagnostic {
	var diagnostics []Diagnostic
	if len(data) == 0 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "control is empty or exceeds 262144 bytes", "supply one bounded nonempty control")
		return finishDiagnostics(diagnostics)
	}
	if len(data) > MaxFileBytes {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "control is empty or exceeds 262144 bytes", "supply one bounded nonempty control")
		return finishDiagnostics(diagnostics)
	}
	if !utf8.Valid(data) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "control is not valid UTF-8", "encode the control as UTF-8")
	}
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "UTF-8 BOM is forbidden", "remove the byte-order mark")
	}
	if len(data) == 0 || data[len(data)-1] != '\n' || len(data) > 1 && data[len(data)-2] == '\n' {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "control must end in exactly one LF", "normalize the final newline")
	}
	for _, value := range data {
		if value == '\r' || value == 0x7f || value < 0x20 && value != '\n' {
			diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, "raw terminal or control bytes are forbidden", "escape data and normalize controls")
			break
		}
	}
	return finishDiagnostics(diagnostics)
}

func scanJSON(path string, data []byte) []Diagnostic {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var diagnostics []Diagnostic
	if err := scanJSONValue(decoder, path, 0, &diagnostics); err != nil && !errors.Is(err, errJSONBound) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, err.Error(), "supply one strict bounded JSON value")
	}
	if len(diagnostics) == 0 {
		if token, err := decoder.Token(); err != io.EOF {
			if err == nil {
				diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, fmt.Sprintf("trailing JSON token %v", token), "remove trailing JSON data")
			} else {
				diagnostics = addDiagnostic(diagnostics, "EVAL-200", path, err.Error(), "supply one strict bounded JSON value")
			}
		}
	}
	return finishDiagnostics(diagnostics)
}

func scanJSONValue(decoder *json.Decoder, subject string, depth int, diagnostics *[]Diagnostic) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch value := token.(type) {
	case json.Delim:
		if depth >= MaxJSONDepth {
			*diagnostics = addDiagnostic(*diagnostics, "EVAL-200", subject, "JSON nesting exceeds 32 containers", "flatten the evaluation control")
			return errJSONBound
		}
		switch value {
		case '{':
			seen := make(map[string]bool)
			fields := 0
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				fields++
				if fields > MaxObjectFields || len(key) > MaxStringBytes {
					*diagnostics = addDiagnostic(*diagnostics, "EVAL-200", subject, "object field count or key length exceeds its bound", "narrow the evaluation record")
				}
				if seen[key] {
					*diagnostics = addDiagnostic(*diagnostics, "EVAL-201", subject+":"+key, "duplicate object key", "retain one unambiguous field")
				}
				seen[key] = true
				if err := scanJSONValue(decoder, subject, depth+1, diagnostics); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			items := 0
			for decoder.More() {
				items++
				if items > MaxCoverageLinks {
					*diagnostics = addDiagnostic(*diagnostics, "EVAL-200", subject, "array contains more than 2048 items", "narrow the bounded evaluation collection")
				}
				if err := scanJSONValue(decoder, subject, depth+1, diagnostics); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", value)
		}
	case string:
		if len(value) > MaxStringBytes {
			*diagnostics = addDiagnostic(*diagnostics, "EVAL-200", subject, "string exceeds 65536 bytes", "shorten the bounded evaluation value")
		}
	case json.Number:
		if strings.ContainsAny(value.String(), ".eE") {
			*diagnostics = addDiagnostic(*diagnostics, "EVAL-200", subject, "fractional or exponent-form numbers are forbidden", "use a bounded base-10 integer")
		}
	}
	return nil
}

func decodeExact(path string, data []byte, target any) []Diagnostic {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return []Diagnostic{newDiagnostic("EVAL-202", path, err.Error(), "match the exact typed evaluation contract")}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return []Diagnostic{newDiagnostic("EVAL-200", path, "JSON contains a trailing value", "retain exactly one JSON value")}
	}
	return nil
}

func safeControlPath(value string) bool {
	if value == "" || len(value) > MaxStringBytes || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "//") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return true
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func sourceSetSHA256(digests []Digest) string {
	var manifest strings.Builder
	for _, digest := range digests {
		manifest.WriteString(digest.SHA256)
		manifest.WriteString("  ")
		manifest.WriteString(digest.Path)
		manifest.WriteByte('\n')
	}
	return sha256Hex([]byte(manifest.String()))
}

func typedControlsSHA256(controls Controls) string {
	typed := struct {
		Protocol        Protocol                  `json:"protocol"`
		Cases           []Case                    `json:"cases"`
		TruthLabels     []TruthLabel              `json:"truth_labels"`
		Graders         []Grader                  `json:"graders"`
		Coverage        Coverage                  `json:"coverage"`
		Adjudication    Adjudication              `json:"adjudication"`
		Descriptors     []render.SchemaDescriptor `json:"descriptors"`
		Bindings        []DocumentBinding         `json:"bindings"`
		SourceDigests   []Digest                  `json:"source_digests"`
		SourceSetSHA256 string                    `json:"source_set_sha256"`
	}{
		Protocol: controls.Protocol, Cases: controls.Cases, TruthLabels: controls.TruthLabels,
		Graders: controls.Graders, Coverage: controls.Coverage, Adjudication: controls.Adjudication,
		Descriptors: controls.Descriptors, Bindings: controls.Bindings,
		SourceDigests: controls.SourceDigests, SourceSetSHA256: controls.SourceSetSHA256,
	}
	return sha256Hex(jsonLine(typed))
}

func controlsWithinBounds(controls Controls) bool {
	return typedValuesWithinBounds(controls)
}

func typedValuesWithinBounds(values ...any) bool {
	items := 0
	bytesSeen := 0
	for _, value := range values {
		if !boundedTypedValue(reflect.ValueOf(value), 0, &items, &bytesSeen) {
			return false
		}
	}
	return true
}

func boundedTypedValue(value reflect.Value, depth int, items, bytesSeen *int) bool {
	if depth > MaxJSONDepth || *items > 65536 || *bytesSeen > MaxAggregateBytes {
		return false
	}
	(*items)++
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return true
		}
		return boundedTypedValue(value.Elem(), depth+1, items, bytesSeen)
	case reflect.Struct:
		if value.NumField() > MaxObjectFields {
			return false
		}
		for index := 0; index < value.NumField(); index++ {
			if !boundedTypedValue(value.Field(index), depth+1, items, bytesSeen) {
				return false
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Len() > MaxCoverageLinks {
			return false
		}
		for index := 0; index < value.Len(); index++ {
			if !boundedTypedValue(value.Index(index), depth+1, items, bytesSeen) {
				return false
			}
		}
	case reflect.Map:
		if value.Len() > MaxObjectFields {
			return false
		}
		iterator := value.MapRange()
		for iterator.Next() {
			if !boundedTypedValue(iterator.Key(), depth+1, items, bytesSeen) || !boundedTypedValue(iterator.Value(), depth+1, items, bytesSeen) {
				return false
			}
		}
	case reflect.String:
		length := value.Len()
		if length > MaxStringBytes || length > MaxAggregateBytes-*bytesSeen {
			return false
		}
		*bytesSeen += length
	}
	return *items <= 65536 && *bytesSeen <= MaxAggregateBytes
}

func obligationID(requirement string) string {
	return strings.Replace(requirement, "L7-", "L7-OBL-", 1)
}

func equalStrings(actual, expected []string) bool {
	return strings.Join(actual, "\x00") == strings.Join(expected, "\x00")
}

func sortedUnique(subject string, values []string, allowEmpty bool) []Diagnostic {
	if len(values) == 0 && allowEmpty {
		return nil
	}
	var diagnostics []Diagnostic
	previous := ""
	for index, value := range values {
		if value == "" || index > 0 && value <= previous {
			diagnostics = addDiagnostic(diagnostics, "EVAL-203", subject, "set-valued array is empty, duplicated, or not bytewise sorted", "sort unique stable values")
			break
		}
		previous = value
	}
	return diagnostics
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func newDiagnostic(rule, subject, message, next string) Diagnostic {
	return Diagnostic{Rule: boundedASCII(rule, 32), Subject: boundedASCII(subject, 160), Message: boundedASCII(message, MaxDiagnosticBytes), Next: boundedASCII(next, MaxDiagnosticBytes)}
}

func addDiagnostic(current []Diagnostic, rule, subject, message, next string) []Diagnostic {
	return appendDiagnostics(current, newDiagnostic(rule, subject, message, next))
}

func appendDiagnostics(current []Diagnostic, additional ...Diagnostic) []Diagnostic {
	combined := make([]Diagnostic, 0, len(current)+len(additional))
	for _, item := range current {
		combined = append(combined, newDiagnostic(item.Rule, item.Subject, item.Message, item.Next))
	}
	for _, item := range additional {
		combined = append(combined, newDiagnostic(item.Rule, item.Subject, item.Message, item.Next))
	}
	sort.Slice(combined, func(left, right int) bool { return diagnosticLess(combined[left], combined[right]) })

	bounded := make([]Diagnostic, 0, min(len(combined), MaxDiagnostics))
	totalBytes := 0
	for _, item := range combined {
		itemBytes := diagnosticBytes(item)
		if len(bounded) == MaxDiagnostics || totalBytes+itemBytes > MaxDiagnosticsBytes {
			break
		}
		bounded = append(bounded, item)
		totalBytes += itemBytes
	}
	return bounded
}

func diagnosticBytes(item Diagnostic) int {
	return len(item.Rule) + len(item.Subject) + len(item.Message) + len(item.Next)
}

func finishDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	sort.Slice(diagnostics, func(left, right int) bool { return diagnosticLess(diagnostics[left], diagnostics[right]) })
	return diagnostics
}

func diagnosticLess(left, right Diagnostic) bool {
	if left.Rule != right.Rule {
		return left.Rule < right.Rule
	}
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Message != right.Message {
		return left.Message < right.Message
	}
	return left.Next < right.Next
}

func boundedASCII(value string, limit int) string {
	if index := strings.Index(value, "L7_SYNTHETIC_CANARY_"); index >= 0 {
		value = value[:index] + "[REDACTED_CANARY]"
	}
	var result strings.Builder
	for _, character := range value {
		if character >= 0x20 && character <= 0x7e {
			result.WriteRune(character)
		} else {
			result.WriteByte('?')
		}
		if result.Len() >= limit {
			break
		}
	}
	return result.String()
}
