package evaluator

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestCoverageClosesExactSourceDerived29By29Roster(t *testing.T) {
	controls := loadRepositoryControls(t)
	result := coverageResult(controls.Coverage)
	if result.Requirements != 29 || result.Obligations != 29 || result.Axes < int64(len(requiredCoverageAxes)) || !result.Complete {
		t.Fatalf("coverage result = %+v", result)
	}

	var document struct {
		Obligations []struct {
			ID                string   `json:"id"`
			SourceRequirement string   `json:"source_requirement"`
			GraderIDs         []string `json:"grader_ids"`
			PublicCaseIDs     []string `json:"public_case_ids"`
		} `json:"obligations"`
	}
	if err := json.Unmarshal(readRepositoryFile(t, "semantic/taxonomy/obligations.json"), &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Obligations) != 29 {
		t.Fatalf("semantic obligations = %d, want 29", len(document.Obligations))
	}
	for index, obligation := range document.Obligations {
		entry := controls.Coverage.Axes[index]
		if entry.RequirementID != obligation.SourceRequirement || entry.ObligationID != obligation.ID || !contains(entry.CaseIDs, obligation.PublicCaseIDs[0]) || !contains(entry.GraderIDs, obligation.GraderIDs[0]) {
			t.Fatalf("coverage/source parity at %d: coverage=%+v obligation=%+v", index, entry, obligation)
		}
	}

	axisSet := map[string]bool{}
	for _, entry := range controls.Coverage.Axes {
		axisSet[entry.Axis] = true
	}
	for _, axis := range requiredCoverageAxes {
		if !axisSet[axis] {
			t.Fatalf("required coverage axis %s is missing", axis)
		}
	}
}

func TestCoverageLinksEverySeededBrokenCandidateByIntendedRule(t *testing.T) {
	controls := loadRepositoryControls(t)
	linkedCases := map[string]bool{}
	linkedGraders := map[string]bool{}
	for _, entry := range controls.Coverage.Axes {
		for _, caseID := range entry.CaseIDs {
			linkedCases[caseID] = true
		}
		for _, graderID := range entry.GraderIDs {
			linkedGraders[graderID] = true
		}
	}
	for caseID := range brokenCaseRules {
		if !linkedCases[caseID] {
			t.Errorf("broken case %s lacks a source-owned coverage link", caseID)
		}
	}
	for caseID := range semanticCaseRules {
		if !linkedCases[caseID] {
			t.Errorf("semantic case %s lacks a source-owned coverage link", caseID)
		}
	}
	for _, required := range []string{"L7-EGR-CANARY-NONLEAK", "L7-EGR-EVIDENCE-TRUTH", "L7-EGR-NO-SUBAGENT", "L7-EGR-OBLIGATION-ACCOUNTING", "L7-EGR-ROUTING-FLOOR", "L7-EGR-STALE-APPROVAL"} {
		if !linkedGraders[required] {
			t.Errorf("intended-rule grader %s lacks a coverage link", required)
		}
	}
}

func TestCoverageDuplicateMissingAndDanglingLinksFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Controls)
		rule   string
	}{
		{name: "duplicate owner", rule: "EVAL-228", mutate: func(controls *Controls) {
			controls.Coverage.Axes[1].RequirementID = controls.Coverage.Axes[0].RequirementID
		}},
		{name: "missing owner", rule: "EVAL-228", mutate: func(controls *Controls) { controls.Coverage.Axes = controls.Coverage.Axes[1:] }},
		{name: "dangling case", rule: "EVAL-229", mutate: func(controls *Controls) { controls.Coverage.Axes[0].CaseIDs = []string{"L7-CASE-UNKNOWN"} }},
		{name: "dangling truth", rule: "EVAL-229", mutate: func(controls *Controls) { controls.Coverage.Axes[0].TruthIDs = []string{"L7-TRUTH-UNKNOWN"} }},
		{name: "dangling grader", rule: "EVAL-229", mutate: func(controls *Controls) { controls.Coverage.Axes[0].GraderIDs = []string{"L7-EGR-UNKNOWN"} }},
		{name: "non-supporting grader", rule: "EVAL-229", mutate: func(controls *Controls) {
			controls.Coverage.Axes[0].GraderIDs = []string{"L7-EGR-MODEL-JUDGE-SUPPLEMENTAL"}
		}},
		{name: "owning case axis mismatch", rule: "EVAL-229", mutate: func(controls *Controls) {
			controls.Cases[0].Axes.Scenario = "different-axis"
		}},
		{name: "missing broken fixture coverage", rule: "EVAL-229", mutate: func(controls *Controls) {
			for index := range controls.Coverage.Axes {
				var retained []string
				for _, caseID := range controls.Coverage.Axes[index].CaseIDs {
					if caseID != "L7-CASE-BL002-BROKEN-CANARY" {
						retained = append(retained, caseID)
					}
				}
				controls.Coverage.Axes[index].CaseIDs = retained
			}
		}},
		{name: "missing required axis", rule: "EVAL-229", mutate: func(controls *Controls) { controls.Coverage.Axes[0].Axis = "semantic-obligation" }},
		{name: "wrong feature", rule: "EVAL-228", mutate: func(controls *Controls) { controls.Coverage.Axes[0].Feature = "L7-BL-003" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controls := loadRepositoryControls(t)
			test.mutate(&controls)
			requireRule(t, ValidateCoverage(controls.Coverage, controls.Cases, controls.TruthLabels, controls.Graders), test.rule)
		})
	}
}

func TestCoverageDiagnosticsAreLexicallyDeterministic(t *testing.T) {
	controls := loadRepositoryControls(t)
	controls.Coverage.Axes[0].CaseIDs = []string{"L7-CASE-Z", "L7-CASE-A"}
	controls.Coverage.Axes[1].TruthIDs = []string{"L7-TRUTH-Z"}
	controls.Coverage.Axes[2].GraderIDs = []string{"L7-EGR-Z"}
	first := ValidateCoverage(controls.Coverage, controls.Cases, controls.TruthLabels, controls.Graders)
	second := ValidateCoverage(controls.Coverage, controls.Cases, controls.TruthLabels, controls.Graders)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("identical malformed coverage produced different diagnostics")
	}
	keys := make([]string, len(first))
	for index, diagnostic := range first {
		keys[index] = diagnostic.Rule + "\x00" + diagnostic.Subject + "\x00" + diagnostic.Message
	}
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)
	if !reflect.DeepEqual(keys, sorted) {
		t.Fatalf("diagnostics are not lexical: %v", keys)
	}
}

func TestCoveragePreservesFutureInstallPlaceholderAndNoSubagentSemantics(t *testing.T) {
	controls := loadRepositoryControls(t)
	placeholder := false
	noSubagent := false
	for _, entry := range controls.Coverage.Axes {
		if entry.Axis == "install-lifecycle-placeholder" {
			placeholder = entry.Feature == "L7-BL-003" && len(entry.CaseIDs) > 0 && len(entry.TruthIDs) > 0 && len(entry.GraderIDs) > 0
		}
		if contains(entry.CaseIDs, "L7-CASE-BL002-BROKEN-SUBAGENT") && contains(entry.GraderIDs, "L7-EGR-NO-SUBAGENT") {
			noSubagent = true
		}
	}
	if !placeholder || !noSubagent {
		t.Fatalf("placeholder=%t no_subagent=%t", placeholder, noSubagent)
	}
}
