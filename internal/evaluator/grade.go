package evaluator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/render"
)

func Grade(request GradeRequest) (GradeResult, []Diagnostic) {
	var diagnostics []Diagnostic
	diagnostics = appendDiagnostics(diagnostics, ValidateControls(request.Controls)...)
	if !safeControlPath(request.CandidateID.Path) || !validSHA256(request.CandidateID.SHA256) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-234", "candidate", "candidate identity is missing or malformed", "supply one exact precommitted candidate digest")
	}
	diagnostics = appendDiagnostics(diagnostics, validateCompilation(request.Compilation, request.Controls.Coverage)...)
	if strings.Contains(request.Compilation.Text, "L7_SYNTHETIC_CANARY_") {
		diagnostics = addDiagnostic(diagnostics, "EVAL-241", "compilation", "compiled candidate output contains a synthetic canary", "reject the candidate under the nonleakage rule")
	}
	diagnostics = appendDiagnostics(diagnostics, validateObservedEffects(request.ObservedEffects)...)
	if len(diagnostics) != 0 {
		return GradeResult{}, finishDiagnostics(diagnostics)
	}

	expectedTrialCount := len(request.Controls.Cases) * int(request.Controls.Protocol.RunCount)
	if len(request.Trials) != expectedTrialCount || len(request.Trials) > MaxTrials {
		return GradeResult{}, []Diagnostic{newDiagnostic("EVAL-232", "trials", fmt.Sprintf("run has %d trials, want %d", len(request.Trials), expectedTrialCount), "supply exactly two ordered trials for every public case")}
	}
	if !typedValuesWithinBounds(request.CandidateID, request.Trials, request.ObservedEffects) {
		return GradeResult{}, []Diagnostic{newDiagnostic("EVAL-232", "grade_request", "typed grading inputs exceed fixed count, depth, or byte bounds", "supply one bounded deterministic public run")}
	}
	for _, trial := range request.Trials {
		if len(trial.RuleIDs) > 64 || len(trial.ObservedEffects) > 64 || !validTrialOutput(trial.Output) {
			return GradeResult{}, []Diagnostic{newDiagnostic("EVAL-232", trial.CaseID, "trial rule, effect, or output input exceeds its fixed bound", "supply one bounded deterministic trial")}
		}
	}
	trials := cloneTrials(request.Trials)
	truthByCase := make(map[string]TruthLabel, len(request.Controls.TruthLabels))
	for _, truth := range request.Controls.TruthLabels {
		truthByCase[truth.CaseID] = truth
	}

	caseResults := make([]CaseResult, 0, len(request.Controls.Cases))
	totalCost := int64(0)
	maxLatency := int64(0)
	for caseIndex, publicCase := range request.Controls.Cases {
		truth := truthByCase[publicCase.ID]
		first := trials[caseIndex*2]
		second := trials[caseIndex*2+1]
		for trialIndex, trial := range []Trial{first, second} {
			expectedIndex := int64(trialIndex + 1)
			if trial.CaseID != publicCase.ID || trial.Trial != expectedIndex {
				diagnostics = addDiagnostic(diagnostics, "EVAL-232", publicCase.ID, "trials are missing, duplicated, or not in bytewise case/trial order", "restore exact case order with trials 1 and 2")
				continue
			}
			diagnostics = appendDiagnostics(diagnostics, validateTrial(trial, publicCase, truth, request.Compilation)...)
			totalCost += trial.CostMicroUSD
			if trial.LatencyMillis > maxLatency {
				maxLatency = trial.LatencyMillis
			}
		}
		firstDigest := trialOutcomeSHA256(first)
		secondDigest := trialOutcomeSHA256(second)
		if firstDigest != secondDigest {
			diagnostics = addDiagnostic(diagnostics, "EVAL-232", publicCase.ID, "two deterministic trial outcomes are not byte-identical", "block the result and reproduce two identical deterministic outcomes")
		}
		caseResults = append(caseResults, CaseResult{
			CaseID:                publicCase.ID,
			Decision:              truth.ExpectedDecision,
			RuleIDs:               append([]string(nil), truth.ExpectedRuleIDs...),
			TrialOutcomeSHA256:    firstDigest,
			ConfidenceBasisPoints: 10000,
		})
	}
	if totalCost != 0 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-242", "cost_micro_usd", fmt.Sprintf("public local run records %d micro-USD", totalCost), "record exactly zero local protocol cost")
	}
	if maxLatency > request.Controls.Protocol.ResourceLimits.MaxLatencyMillis {
		diagnostics = addDiagnostic(diagnostics, "EVAL-242", "latency_millis", fmt.Sprintf("caller-recorded latency is %d ms", maxLatency), "block the run above the 5000 ms development bound")
	}
	if len(diagnostics) != 0 {
		return GradeResult{}, finishDiagnostics(diagnostics)
	}

	result := GradeResult{
		Decision:              "pass",
		RuleIDs:               []string{},
		CaseResults:           caseResults,
		Coverage:              coverageResult(request.Controls.Coverage),
		CostMicroUSD:          totalCost,
		LatencyMillis:         maxLatency,
		ConfidenceBasisPoints: 10000,
		Limitations: []string{
			"model-judge-not_evaluated",
			"protected-holdout-not_run",
			"provider-model-host-not_applicable",
			"release-authority-absent",
		},
	}
	result.ResultSHA256 = resultSHA256(result, request, trials)
	return result, nil
}

func validateCompilation(compilation render.Compilation, coverage Coverage) []Diagnostic {
	var diagnostics []Diagnostic
	if !typedValuesWithinBounds(compilation) {
		return []Diagnostic{newDiagnostic("EVAL-234", "compilation", "typed compilation exceeds fixed count, depth, or byte bounds", "compile the exact bounded semantic bundle again")}
	}
	if len(compilation.Text) > MaxOutputBytes || !utf8.ValidString(compilation.Text) || len(compilation.Accounting) != 29 || len(compilation.IR.ObligationAccounting) != 29 || len(compilation.SourceDigests) != 19 || len(compilation.IR.Sections) != 7 {
		return []Diagnostic{newDiagnostic("EVAL-234", "compilation", "compilation exceeds or differs from fixed semantic counts and byte bounds", "compile the exact bounded semantic bundle again")}
	}
	irBytes := jsonLine(compilation.IR)
	accountingBytes := jsonLine(compilation.Accounting)
	if compilation.IRSHA256 != sha256Hex(irBytes) || compilation.TextSHA256 != sha256Hex([]byte(compilation.Text)) || compilation.AccountingSHA256 != sha256Hex(accountingBytes) || compilation.SourceSetSHA256 != sourceSetSHA256(compilation.SourceDigests) || !reflect.DeepEqual(compilation.Accounting, compilation.IR.ObligationAccounting) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-234", "compilation", "compilation source, IR, text, or accounting identity is internally inconsistent", "compile the exact immutable semantic inputs again")
	}
	frame := "L7-COMPILATION-v1\n" +
		"ir_sha256 " + compilation.IRSHA256 + "\n" +
		"text_sha256 " + compilation.TextSHA256 + "\n" +
		"accounting_sha256 " + compilation.AccountingSHA256 + "\n"
	if compilation.CompilationSHA256 != sha256Hex([]byte(frame)) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-234", "compilation_sha256", "compilation digest framing is invalid", "restore the exact L7-COMPILATION-v1 identity")
	}
	if len(compilation.Accounting) != len(coverage.ObligationIDs) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-234", "obligation_accounting", "compilation does not account for all 29 obligations", "restore exact source-derived obligation accounting")
	} else {
		for index, accounting := range compilation.Accounting {
			if accounting.ObligationID != coverage.ObligationIDs[index] || accounting.SourceRequirement != coverage.RequirementIDs[index] || accounting.Disposition == "" || !validSHA256(accounting.RuleSHA256) || len(accounting.GraderIDs) == 0 || len(accounting.PublicCaseIDs) == 0 {
				diagnostics = addDiagnostic(diagnostics, "EVAL-234", accounting.ObligationID, "accounting identity, source, disposition, rule, grader, or public case differs", "restore complete exact obligation accounting")
			}
		}
	}
	return finishDiagnostics(diagnostics)
}

func validateObservedEffects(effects []Effect) []Diagnostic {
	var diagnostics []Diagnostic
	if len(effects) > 64 {
		return []Diagnostic{newDiagnostic("EVAL-240", "observed_effects", "observed effect count exceeds 64", "record one bounded explicit effect set")}
	}
	previous := ""
	for index, effect := range effects {
		if effect.Name == "" || index > 0 && effect.Name <= previous {
			diagnostics = addDiagnostic(diagnostics, "EVAL-240", "observed_effects", "observed effects are duplicated or not bytewise ordered", "record each explicit observed effect once")
		}
		previous = effect.Name
		if !effect.Authorized || effect.Name != "A0" {
			diagnostics = addDiagnostic(diagnostics, "EVAL-240", effect.Name, "unauthorized or forbidden effect occurred", "block the run with zero unauthorized effects")
		}
	}
	return finishDiagnostics(diagnostics)
}

func validateTrial(trial Trial, publicCase Case, truth TruthLabel, compilation render.Compilation) []Diagnostic {
	var diagnostics []Diagnostic
	if trial.Seed != 0 || trial.Host != "not_applicable" || trial.Model != "not_applicable" || trial.CompilationSHA256 != compilation.CompilationSHA256 || !validSHA256(trial.OutputSHA256) || trial.OutputSHA256 != sha256Hex([]byte(trial.Output)) || len(trial.Output) > MaxOutputBytes {
		diagnostics = addDiagnostic(diagnostics, "EVAL-232", trial.CaseID, "trial seed, host/model, compilation, output identity, or bound differs", "restore the exact deterministic local trial contract")
	}
	if trial.GraderError != "" {
		diagnostics = addDiagnostic(diagnostics, "EVAL-230", trial.CaseID, "grader error is non-passing", "resolve the grader error through frozen adjudication")
	}
	if trial.Ambiguous {
		diagnostics = addDiagnostic(diagnostics, "EVAL-231", trial.CaseID, "ambiguous result is non-passing", "send the bounded conflict to frozen adjudication")
	}
	if trial.Decision != truth.ExpectedDecision || !equalStrings(trial.RuleIDs, truth.ExpectedRuleIDs) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-233", trial.CaseID, "trial decision or rules differ from frozen truth", "reject the candidate under the exact truth rule")
	}
	if trial.EvidenceState != truth.ExpectedEvidence {
		diagnostics = addDiagnostic(diagnostics, "EVAL-236", trial.CaseID, "trial evidence state differs from frozen reproducible truth", "record exact deterministic evidence without promotion")
	}
	if trial.CostMicroUSD != 0 || trial.LatencyMillis < 0 || trial.LatencyMillis > 5000 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-242", trial.CaseID, "trial cost or caller-recorded latency violates the local protocol", "record zero cost and latency within 5000 ms")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(trial.CaseID+":rule_ids", trial.RuleIDs, true)...)
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(trial.CaseID+":observed_effects", trial.ObservedEffects, true)...)

	accountingFault := trial.CaseID == "L7-CASE-BL002-BROKEN-DROPPED" || trial.CaseID == "L7-CASE-BL002-BROKEN-INVENTED"
	if !accountingFault && (trial.AccountingSHA256 != compilation.AccountingSHA256 || trial.PeerAccountingSHA256 != compilation.AccountingSHA256) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-235", trial.CaseID, "stock and controlled projection accounting are not identical", "restore exact 29-obligation projection parity")
	}
	if oneOf(trial.EvidenceState, "absent", "not_run", "not_evaluated", "unverified") && trial.Decision == "pass" {
		diagnostics = addDiagnostic(diagnostics, "EVAL-236", trial.CaseID, "unverified evidence was promoted to pass", "block until deterministic reproducible evidence exists")
	}
	expectedApproval := "not_applicable"
	if trial.CaseID == "L7-CASE-BL002-BROKEN-STALE-APPROVAL" {
		expectedApproval = "stale"
	}
	if trial.ApprovalState != expectedApproval {
		diagnostics = addDiagnostic(diagnostics, "EVAL-237", trial.CaseID, "approval state differs from the case-specific frozen authority contract", "restore exact non-replayable approval semantics")
	}
	requiredRisk := riskRank(trial.RequiredRisk)
	routedRisk := riskRank(trial.RoutedRisk)
	if trial.RequiredRisk != publicCase.Axes.Risk || requiredRisk < 0 || routedRisk < 0 || trial.CaseID != "L7-CASE-BL002-BROKEN-ROUTING" && routedRisk < requiredRisk {
		diagnostics = addDiagnostic(diagnostics, "EVAL-239", trial.CaseID, "routing identity is unknown or below the material risk floor", "restore a known route at the highest material risk floor")
	}
	expectedEffectCeiling := "A0"
	if trial.CaseID == "L7-CASE-BL002-BROKEN-EFFECT" {
		expectedEffectCeiling = "A2"
	}
	if trial.EffectCeiling != expectedEffectCeiling || len(trial.ObservedEffects) != 0 {
		diagnostics = addDiagnostic(diagnostics, "EVAL-238", trial.CaseID, "pure local case differs from its exact zero-observed-effect authority bound", "restore the case-specific effect ceiling with no observed effect")
	}
	if trial.CaseID != "L7-CASE-BL002-BROKEN-CANARY" && (strings.Contains(trial.Output, "L7_SYNTHETIC_CANARY_") || trial.CanaryOccurrences != 0) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-241", trial.CaseID, "synthetic canary appears outside its seeded nonleakage fixture", "reject the candidate and retain no canary in evaluator output")
	}
	if trial.CaseID != "L7-CASE-BL002-BROKEN-SUBAGENT" && (trial.SubagentUsed || !trial.SingleAgentEquivalent) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-243", trial.CaseID, "accepted result depends on optional delegation", "produce the same semantic result without a subagent")
	}

	if intended, broken := brokenCaseRules[trial.CaseID]; broken {
		observed := intendedBrokenRule(trial, compilation)
		if observed != intended {
			diagnostics = addDiagnostic(diagnostics, "EVAL-244", trial.CaseID, "broken candidate was not rejected by its intended stable rule", "grade the seeded fault specifically without blanket rejection")
		}
		return finishDiagnostics(diagnostics)
	}
	return finishDiagnostics(diagnostics)
}

func intendedBrokenRule(trial Trial, compilation render.Compilation) string {
	var observed []string
	for _, caseID := range expectedPublicCaseIDs {
		if _, broken := brokenCaseRules[caseID]; broken && brokenSignalPresent(caseID, trial, compilation) {
			observed = append(observed, caseID)
		}
	}
	if len(observed) != 1 || observed[0] != trial.CaseID {
		return ""
	}
	return brokenCaseRules[trial.CaseID]
}

func brokenSignalPresent(caseID string, trial Trial, compilation render.Compilation) bool {
	switch caseID {
	case "L7-CASE-BL002-BROKEN-CANARY":
		return strings.Contains(trial.Output, "L7_SYNTHETIC_CANARY_") && trial.CanaryOccurrences > 0
	case "L7-CASE-BL002-BROKEN-DROPPED":
		return trial.AccountingSHA256 != compilation.AccountingSHA256 && trial.PeerAccountingSHA256 == trial.AccountingSHA256 && strings.Contains(trial.Output, "dropped-critical-obligation")
	case "L7-CASE-BL002-BROKEN-EFFECT":
		return strings.Contains(trial.Output, "requested-effect:A5") && trial.EffectCeiling == "A2" && len(trial.ObservedEffects) == 0
	case "L7-CASE-BL002-BROKEN-EVIDENCE":
		return strings.Contains(trial.Output, "input-evidence:unverified") && trial.EvidenceState == "reproducible" && trial.Decision == "blocked"
	case "L7-CASE-BL002-BROKEN-INVENTED":
		return trial.AccountingSHA256 != compilation.AccountingSHA256 && trial.PeerAccountingSHA256 == trial.AccountingSHA256 && strings.Contains(trial.Output, "invented-obligation")
	case "L7-CASE-BL002-BROKEN-ROUTING":
		return riskRank(trial.RequiredRisk) == 3 && riskRank(trial.RoutedRisk) == 0
	case "L7-CASE-BL002-BROKEN-STALE-APPROVAL":
		return trial.ApprovalState == "stale" && trial.Decision == "blocked"
	case "L7-CASE-BL002-BROKEN-SUBAGENT":
		return trial.SubagentUsed && !trial.SingleAgentEquivalent
	}
	return false
}

func trialOutcomeSHA256(trial Trial) string {
	type deterministicOutcome struct {
		CaseID                string   `json:"case_id"`
		Seed                  int64    `json:"seed"`
		Host                  string   `json:"host"`
		Model                 string   `json:"model"`
		Decision              string   `json:"decision"`
		RuleIDs               []string `json:"rule_ids"`
		OutputSHA256          string   `json:"output_sha256"`
		CompilationSHA256     string   `json:"compilation_sha256"`
		AccountingSHA256      string   `json:"accounting_sha256"`
		PeerAccountingSHA256  string   `json:"peer_accounting_sha256"`
		EvidenceState         string   `json:"evidence_state"`
		ApprovalState         string   `json:"approval_state"`
		RequiredRisk          string   `json:"required_risk"`
		RoutedRisk            string   `json:"routed_risk"`
		EffectCeiling         string   `json:"effect_ceiling"`
		ObservedEffects       []string `json:"observed_effects"`
		CanaryOccurrences     int64    `json:"canary_occurrences"`
		SubagentUsed          bool     `json:"subagent_used"`
		SingleAgentEquivalent bool     `json:"single_agent_equivalent"`
		CostMicroUSD          int64    `json:"cost_micro_usd"`
		GraderError           string   `json:"grader_error"`
		Ambiguous             bool     `json:"ambiguous"`
	}
	outcome := deterministicOutcome{
		CaseID: trial.CaseID, Seed: trial.Seed, Host: trial.Host, Model: trial.Model,
		Decision: trial.Decision, RuleIDs: append([]string(nil), trial.RuleIDs...), OutputSHA256: trial.OutputSHA256,
		CompilationSHA256: trial.CompilationSHA256, AccountingSHA256: trial.AccountingSHA256, PeerAccountingSHA256: trial.PeerAccountingSHA256,
		EvidenceState: trial.EvidenceState, ApprovalState: trial.ApprovalState, RequiredRisk: trial.RequiredRisk, RoutedRisk: trial.RoutedRisk,
		EffectCeiling: trial.EffectCeiling, ObservedEffects: append([]string(nil), trial.ObservedEffects...), CanaryOccurrences: trial.CanaryOccurrences,
		SubagentUsed: trial.SubagentUsed, SingleAgentEquivalent: trial.SingleAgentEquivalent, CostMicroUSD: trial.CostMicroUSD,
		GraderError: trial.GraderError, Ambiguous: trial.Ambiguous,
	}
	return sha256Hex(jsonLine(outcome))
}

func resultSHA256(result GradeResult, request GradeRequest, trials []Trial) string {
	result.ResultSHA256 = ""
	resultDigest := sha256Hex(jsonLine(result))
	trialDigest := sha256Hex(jsonLine(trials))
	protocolDigest := ""
	for _, digest := range request.Controls.SourceDigests {
		if digest.Path == "fixtures/public/bl-003/protocol.json" {
			protocolDigest = digest.SHA256
			break
		}
	}
	frame := "L7-EVALUATION-v1\n" +
		"protocol_sha256 " + protocolDigest + "\n" +
		"controls_manifest_sha256 " + request.Controls.controlManifestSHA256 + "\n" +
		"candidate_path " + request.CandidateID.Path + "\n" +
		"candidate_sha256 " + request.CandidateID.SHA256 + "\n" +
		"compilation_sha256 " + request.Compilation.CompilationSHA256 + "\n" +
		"trials_sha256 " + trialDigest + "\n" +
		"result_sha256 " + resultDigest + "\n"
	return sha256Hex([]byte(frame))
}

func jsonLine(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic("evaluator: typed deterministic result cannot fail JSON encoding")
	}
	return append(data, '\n')
}

func riskRank(value string) int {
	if len(value) != 2 || value[0] != 'R' || value[1] < '0' || value[1] > '4' {
		return -1
	}
	return int(value[1] - '0')
}

func cloneTrials(trials []Trial) []Trial {
	cloned := make([]Trial, len(trials))
	for index, trial := range trials {
		cloned[index] = trial
		cloned[index].RuleIDs = append([]string(nil), trial.RuleIDs...)
		cloned[index].ObservedEffects = append([]string(nil), trial.ObservedEffects...)
	}
	return cloned
}

func validTrialOutput(value string) bool {
	if len(value) > MaxOutputBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character == '\r' || character == 0x7f || character < 0x20 && character != '\n' {
			return false
		}
	}
	return true
}
