package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentTraceContract(t *testing.T) {
	t.Parallel()
	result, findings := checkTrace(repositoryRoot(t))
	if len(findings) != 0 {
		t.Fatalf("trace findings: %+v", findings)
	}
	if result.total != 163 || result.allocations["V1.0"] != 140 || result.allocations["V1.x"] != 18 || result.allocations["Later"] != 5 {
		t.Fatalf("unexpected trace result: %+v", result)
	}
}

func TestRequirementExpressionExpansion(t *testing.T) {
	t.Parallel()
	ids, findings := expandRequirementExpression("`L7-FLOW-001`–`003`, `005`, `008`–`009`")
	if len(findings) != 0 {
		t.Fatalf("unexpected findings: %+v", findings)
	}
	want := "L7-FLOW-001,L7-FLOW-002,L7-FLOW-003,L7-FLOW-005,L7-FLOW-008,L7-FLOW-009"
	if got := strings.Join(ids, ","); got != want {
		t.Fatalf("expanded IDs: got %q, want %q", got, want)
	}
}

func TestRequirementExpressionRejectsMalformedAndReversedRanges(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		expression string
		rule       string
	}{
		{"prefixless", "`001`–`003`", "TRACE-110"},
		{"malformed", "`L7-FLOW-1`", "TRACE-110"},
		{"reversed", "`L7-FLOW-009`–`003`", "TRACE-111"},
		{"empty-token", "`L7-FLOW-001`,", "TRACE-110"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, findings := expandRequirementExpression(testCase.expression)
			if findingRules(findings)[testCase.rule] == 0 {
				t.Fatalf("findings %+v do not contain %s", findings, testCase.rule)
			}
		})
	}
}

func TestRequirementExpressionEnforcesCumulativeExpansionBound(t *testing.T) {
	t.Parallel()
	ids, findings := expandRequirementExpressionBounded("`L7-FLOW-000`–`009`", 10)
	if len(findings) != 0 || len(ids) != 10 {
		t.Fatalf("at-limit expansion failed: ids=%d findings=%+v", len(ids), findings)
	}
	ids, findings = expandRequirementExpressionBounded("`L7-FLOW-000`–`009`", 9)
	if ids != nil || findingRules(findings)["TRACE-112"] == 0 {
		t.Fatalf("over-limit expansion was accepted: ids=%d findings=%+v", len(ids), findings)
	}

	backlog := "# Test\n\n## 8. Normative requirement ownership and release allocation\n\n" +
		"| Requirement IDs | Accountable backlog owner | Allocation | Count |\n|---|---|---|---|\n" +
		"| `L7-A-000`–`299` | `L7-BL-001` | V1.0 | 300 |\n" +
		"| `L7-B-000`–`299` | `L7-BL-002` | V1.0 | 300 |\n\n## 9. Stop\n"
	owners, findings := parseRequirementOwnership(backlog)
	if len(owners) != 300 || findingRules(findings)["TRACE-112"] == 0 {
		t.Fatalf("document-wide expansion bound failed: owners=%d findings=%+v", len(owners), findings)
	}
}

func TestTraceRejectsDuplicateMissingUnknownAndAllocationDrift(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	requirementsData, err := os.ReadFile(filepath.Join(root, "docs/artifacts/requirements.md"))
	if err != nil {
		t.Fatal(err)
	}
	backlogData, err := os.ReadFile(filepath.Join(root, "docs/artifacts/feature-backlog.md"))
	if err != nil {
		t.Fatal(err)
	}

	definitions, definitionFindings := parseRequirementDefinitions(string(requirementsData))
	owners, ownerFindings := parseRequirementOwnership(string(backlogData))
	if len(definitionFindings)+len(ownerFindings) != 0 {
		t.Fatalf("fixture source findings: %+v %+v", definitionFindings, ownerFindings)
	}

	delete(owners, "L7-INTAKE-001")
	owners["L7-UNKNOWN-999"] = ownership{owner: "L7-BL-001", allocation: "V1.0"}
	owners["L7-INTAKE-002"] = ownership{owner: "L7-BL-006", allocation: "Later"}
	_, findings := validateTrace(definitions, owners)
	rules := findingRules(findings)
	for _, rule := range []string{"TRACE-130", "TRACE-131", "TRACE-133"} {
		if rules[rule] == 0 {
			t.Errorf("findings %+v do not contain %s", findings, rule)
		}
	}

	duplicateLine := "| `L7-INTAKE-001` | Duplicate | Duplicate |\n"
	duplicated := strings.Replace(string(requirementsData), "## 9. Functional requirements\n", "## 9. Functional requirements\n"+duplicateLine, 1)
	_, duplicateFindings := parseRequirementDefinitions(duplicated)
	if findingRules(duplicateFindings)["TRACE-102"] == 0 {
		t.Fatalf("duplicate findings: %+v", duplicateFindings)
	}
}

func TestTraceRejectsMalformedDefinitionAndDuplicateOwner(t *testing.T) {
	t.Parallel()
	requirements := "# Test\n\n## 9. Functional requirements\n\n| ID | Requirement | Verify |\n|---|---|---|\n| `L7-FLOW-1` | Bad | Bad |\n\n## 11. Stop\n"
	_, findings := parseRequirementDefinitions(requirements)
	if findingRules(findings)["TRACE-101"] == 0 {
		t.Fatalf("malformed definition findings: %+v", findings)
	}

	backlog := "# Test\n\n## 8. Normative requirement ownership and release allocation\n\n| Requirement IDs | Accountable backlog owner | Allocation | Count |\n|---|---|---|---|\n| `L7-FLOW-001` | `L7-BL-002` | V1.0 | 1 |\n| `L7-FLOW-001` | `L7-BL-003` | V1.0 | 1 |\n\n## 9. Stop\n"
	_, ownerFindings := parseRequirementOwnership(backlog)
	if findingRules(ownerFindings)["TRACE-124"] == 0 {
		t.Fatalf("duplicate owner findings: %+v", ownerFindings)
	}
}
