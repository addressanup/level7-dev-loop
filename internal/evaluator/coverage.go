package evaluator

import (
	"fmt"
	"strings"
)

var requiredCoverageAxes = []string{
	"authority",
	"budget-exhaustion",
	"degraded-modes",
	"forbidden-effects",
	"injection",
	"install-lifecycle-placeholder",
	"interruption-resume",
	"lifecycle-transitions",
	"parity-obligations",
	"routing-negative-activation",
	"secret-handling",
	"stale-evidence-approval",
	"write-collision-semantics",
}

func ValidateCoverage(coverage Coverage, cases []Case, truths []TruthLabel, graders []Grader) []Diagnostic {
	var diagnostics []Diagnostic
	if !typedValuesWithinBounds(coverage, cases, truths, graders) {
		return []Diagnostic{newDiagnostic("EVAL-200", "coverage", "typed coverage inputs exceed fixed count, depth, or byte bounds", "supply one bounded public coverage graph")}
	}
	diagnostics = appendDiagnostics(diagnostics, validateMeta(coverage.RecordMeta, "L7-COV-PUBLIC-001")...)
	expectedObligations := make([]string, len(expectedRequirementIDs))
	for index, requirement := range expectedRequirementIDs {
		expectedObligations[index] = obligationID(requirement)
	}
	if !equalStrings(coverage.RequirementIDs, expectedRequirementIDs) || !equalStrings(coverage.ObligationIDs, expectedObligations) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-227", coverage.ID, "coverage roster differs from the exact 29 approved requirements and derived obligations", "restore exact source-derived coverage rosters")
	}
	if len(coverage.Axes) != len(expectedRequirementIDs) {
		diagnostics = addDiagnostic(diagnostics, "EVAL-228", coverage.ID, fmt.Sprintf("coverage has %d owning entries, want 29", len(coverage.Axes)), "restore one owning entry per requirement and obligation")
	}

	caseSet := make(map[string]bool, len(cases))
	caseByID := make(map[string]Case, len(cases))
	truthSet := make(map[string]bool, len(truths))
	graderSet := make(map[string]bool, len(graders))
	graderByID := make(map[string]Grader, len(graders))
	for _, item := range cases {
		caseSet[item.ID] = true
		caseByID[item.ID] = item
	}
	for _, truth := range truths {
		truthSet[truth.ID] = true
	}
	for _, grader := range graders {
		graderSet[grader.ID] = true
		graderByID[grader.ID] = grader
	}

	requirementOwners := make(map[string]int)
	obligationOwners := make(map[string]int)
	axisSet := make(map[string]bool)
	coveredCases := make(map[string]bool)
	coveredTruths := make(map[string]bool)
	brokenGraderCoverage := make(map[string]bool)
	links := 0
	previousRequirement := ""
	for index, entry := range coverage.Axes {
		if index > 0 && entry.RequirementID <= previousRequirement {
			diagnostics = addDiagnostic(diagnostics, "EVAL-228", entry.RequirementID, "coverage owning entries are duplicated or not bytewise requirement ordered", "sort unique owning entries by requirement ID")
		}
		previousRequirement = entry.RequirementID
		requirementOwners[entry.RequirementID]++
		obligationOwners[entry.ObligationID]++
		axisSet[entry.Axis] = true
		expectedFeature := "L7-BL-002"
		if strings.HasPrefix(entry.RequirementID, "L7-EVAL-") {
			expectedFeature = "L7-BL-003"
		}
		expectedCase := strings.Replace(entry.RequirementID, "L7-", "L7-CASE-", 1)
		if entry.ObligationID != obligationID(entry.RequirementID) || entry.Feature != expectedFeature || entry.Axis == "" || !contains(entry.CaseIDs, expectedCase) || len(entry.CaseIDs) == 0 || len(entry.TruthIDs) == 0 || len(entry.GraderIDs) == 0 {
			diagnostics = addDiagnostic(diagnostics, "EVAL-228", entry.RequirementID, "owning entry has the wrong obligation, feature, axis, owning case, truth, or grader", "restore its exact source-derived owning links")
		}
		owningCase, owningCaseExists := caseByID[expectedCase]
		owningTruthLinked := false
		if owningCaseExists {
			for _, truthID := range owningCase.TruthIDs {
				if contains(entry.TruthIDs, truthID) {
					owningTruthLinked = true
					break
				}
			}
		}
		if !owningCaseExists || owningCase.FeatureOwner != expectedFeature || owningCase.Axes.Scenario != entry.Axis || !owningTruthLinked {
			diagnostics = addDiagnostic(diagnostics, "EVAL-229", entry.RequirementID, "owning case, feature, axis, or truth link does not close semantically", "restore the exact source-owned case and truth support")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.RequirementID+":case_ids", entry.CaseIDs, false)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.RequirementID+":truth_ids", entry.TruthIDs, false)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.RequirementID+":grader_ids", entry.GraderIDs, false)...)
		for _, caseID := range entry.CaseIDs {
			links++
			coveredCases[caseID] = true
			if !caseSet[caseID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-229", entry.RequirementID+":"+caseID, "coverage case link is dangling", "restore a registered supporting case")
			}
		}
		for _, truthID := range entry.TruthIDs {
			links++
			coveredTruths[truthID] = true
			if !truthSet[truthID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-229", entry.RequirementID+":"+truthID, "coverage truth link is dangling", "restore a registered supporting truth")
			}
		}
		supportingGrader := false
		for _, graderID := range entry.GraderIDs {
			links++
			if !graderSet[graderID] {
				diagnostics = addDiagnostic(diagnostics, "EVAL-229", entry.RequirementID+":"+graderID, "coverage grader link is dangling", "restore a registered supporting grader")
				continue
			}
			grader := graderByID[graderID]
			truthSupport := false
			for _, truthID := range entry.TruthIDs {
				if contains(grader.TruthIDs, truthID) {
					truthSupport = true
					break
				}
			}
			if grader.Class == "deterministic" && contains(grader.ObligationIDs, entry.ObligationID) && truthSupport {
				supportingGrader = true
			}
		}
		if !supportingGrader {
			diagnostics = addDiagnostic(diagnostics, "EVAL-229", entry.RequirementID, "coverage lacks a deterministic grader supporting its obligation and linked truth", "restore one valid source-owned grader link")
		}
		for _, caseID := range entry.CaseIDs {
			if intendedGrader, broken := brokenCaseGraders[caseID]; broken && contains(entry.GraderIDs, intendedGrader) {
				brokenGraderCoverage[caseID] = true
			}
		}
	}
	if links > MaxCoverageLinks {
		diagnostics = addDiagnostic(diagnostics, "EVAL-228", coverage.ID, "coverage link count exceeds 2048", "narrow the bounded public coverage map")
	}
	for _, requirement := range expectedRequirementIDs {
		if requirementOwners[requirement] != 1 || obligationOwners[obligationID(requirement)] != 1 {
			diagnostics = addDiagnostic(diagnostics, "EVAL-228", requirement, "requirement or derived obligation does not have exactly one owning entry", "restore exact 29/29 ownership parity")
		}
	}
	for _, axis := range requiredCoverageAxes {
		if !axisSet[axis] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-229", axis, "required semantic/safety coverage axis is missing", "map the axis to one source-owned coverage entry")
		}
	}
	for _, caseID := range expectedPublicCaseIDs {
		_, broken := brokenCaseRules[caseID]
		_, semantic := semanticCaseRules[caseID]
		if !broken && !semantic {
			continue
		}
		truthID := strings.Replace(caseID, "L7-CASE-", "L7-TRUTH-", 1)
		if !coveredCases[caseID] || !coveredTruths[truthID] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-229", caseID, "public semantic or broken fixture lacks complete case/truth coverage", "restore its source-owned public coverage links")
		}
		if broken && !brokenGraderCoverage[caseID] {
			diagnostics = addDiagnostic(diagnostics, "EVAL-229", caseID, "broken fixture lacks co-located intended stable-rule grader coverage", "restore the exact fault-specific case/grader link")
		}
	}
	return finishDiagnostics(diagnostics)
}

func coverageResult(coverage Coverage) CoverageResult {
	axisSet := make(map[string]bool)
	for _, entry := range coverage.Axes {
		axisSet[entry.Axis] = true
	}
	complete := len(coverage.RequirementIDs) == 29 && len(coverage.ObligationIDs) == 29 && len(coverage.Axes) == 29
	for _, axis := range requiredCoverageAxes {
		complete = complete && axisSet[axis]
	}
	return CoverageResult{
		Requirements: int64(len(coverage.RequirementIDs)),
		Obligations:  int64(len(coverage.ObligationIDs)),
		Axes:         int64(len(axisSet)),
		Complete:     complete,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
