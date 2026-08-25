package main

import (
	"fmt"
	"strings"
)

type ownershipExpectation struct {
	pathKind   string
	path       string
	writer     string
	reviewer   string
	changeGate string
}

var expectedOwnership = map[string]ownershipExpectation{
	"ci-workflow":           {"exact", ".github/workflows/harness.yml", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"makefile":              {"exact", "Makefile", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"readme":                {"exact", "README.md", "wave-integrator", "owner-review", "owner+design"},
	"module":                {"exact", "go.mod", "harness-integrator", "independent-readonly", "owner+module-decision"},
	"wave-approval":         {"exact", "docs/artifacts/wave-01-approval.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-audit":            {"exact", "docs/artifacts/wave-01-audit.md", "independent-reviewer", "owner-review", "separate-audit"},
	"wave-candidate":        {"exact", "docs/artifacts/wave-01-candidate.sha256", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-contract":         {"exact", "docs/artifacts/wave-01-change-contract.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-design":           {"exact", "docs/artifacts/wave-01-design.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-design-amendment": {"exact", "docs/artifacts/wave-01-design-amendment.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-evidence":         {"exact", "docs/artifacts/wave-01-evidence.md", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-grant-amendment":  {"exact", "docs/artifacts/wave-01-grant-ladder-amendment.md", "wave-integrator", "independent-readonly", "separate-security-decision"},
	"wave-module-decision":  {"exact", "docs/artifacts/wave-01-module-identity-decision.md", "wave-integrator", "owner-review", "owner+module-decision"},
	"wave-specification":    {"exact", "docs/artifacts/wave-01-specification.md", "wave-integrator", "owner-review", "owner+design"},
	"harness-data":          {"prefix", "harness/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"harness-scripts":       {"prefix", "scripts/harness/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"harness-code":          {"prefix", "internal/harness/buildcontrol/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"prototype-skills":      {"prefix", "skills/", "protected-prototype-owner", "wave-07-owner", "wave-07-cutover"},
	"codex-manifest":        {"prefix", ".codex-plugin/", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"claude-manifest":       {"prefix", ".claude-plugin/", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"root-plugin":           {"exact", "plugin.json", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"marketplace":           {"exact", "marketplace.json", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"workflow-reference":    {"exact", "references/WORKFLOW.md", "protected-prototype-owner", "wave-07-owner", "wave-07-cutover"},
	"semantic-source":       {"prefix", "semantic/", "semantic-owner", "independent-readonly", "future-wave"},
	"schema-source":         {"prefix", "schemas/", "state-owner", "independent-readonly", "future-wave"},
	"public-fixtures":       {"prefix", "fixtures/", "feature-owner", "evaluator-owner", "future-wave"},
	"public-evaluator":      {"prefix", "internal/evaluator/", "evaluator-owner", "independent-readonly", "future-wave"},
	"generated-build":       {"prefix", "build/generated/", "generator-owner", "independent-readonly", "future-wave"},
	"generated-packages":    {"prefix", "packages/", "generator-owner", "independent-readonly", "wave-10-cutover"},
	"updater":               {"prefix", "cmd/l7up/", "updater-owner", "independent-readonly", "wave-10-cutover"},
	"protected-controls":    {"prefix", "protected/", "external-denied", "external-denied", "external-governance"},
}

func checkOwnership(root string) (int, []finding) {
	rows, findings := loadTSV(root, "harness/control-ownership.tsv", []string{"control", "path_kind", "path", "writer", "reviewer", "change_gate"})
	rules, validationFindings := validateOwnershipRows(rows)
	findings = appendFindings(findings, validationFindings...)
	pathRows, pathFindings := loadTSV(root, "harness/wave-01-paths.tsv", []string{"change", "path", "owner", "rule"})
	findings = appendFindings(findings, pathFindings...)
	findings = appendFindings(findings, crossCheckPathOwnership(pathRows, rules)...)
	return len(rules), findings
}

func validateOwnershipRows(rows []tsvRow) (map[string]ownershipExpectation, []finding) {
	rules := make(map[string]ownershipExpectation)
	var findings []finding
	for _, row := range rows {
		control := row["control"]
		if _, duplicate := rules[control]; duplicate {
			findings = appendFindings(findings, newFinding("OWN-401", control, "duplicate control ownership row", "retain exactly one control owner"))
			continue
		}
		actual := ownershipExpectation{row["path_kind"], row["path"], row["writer"], row["reviewer"], row["change_gate"]}
		if actual.pathKind != "exact" && actual.pathKind != "prefix" {
			findings = appendFindings(findings, newFinding("OWN-402", control, "unknown ownership path kind", "use exact or prefix"))
			continue
		}
		canonicalPath := actual.path
		if actual.pathKind == "prefix" {
			canonicalPath = strings.TrimSuffix(actual.path, "/")
		}
		if !safeRelativeASCIIPath(canonicalPath) || (actual.pathKind == "prefix" && canonicalPath+"/" != actual.path) {
			findings = appendFindings(findings, newFinding("OWN-403", control, "ownership path is noncanonical", "use an exact ASCII path and a trailing slash for prefixes"))
			continue
		}
		rules[control] = actual
		expected, ok := expectedOwnership[control]
		if !ok {
			findings = appendFindings(findings, newFinding("OWN-404", control, "unknown shared-control class", "remove it or obtain an approved ownership change"))
		} else if actual != expected {
			findings = appendFindings(findings, newFinding("OWN-405", control, "ownership row differs from the approved design", "restore the approved writer, reviewer, and gate"))
		}
	}
	for control := range expectedOwnership {
		if _, ok := rules[control]; !ok {
			findings = appendFindings(findings, newFinding("OWN-406", control, "required shared-control ownership is missing", "restore the required ownership row"))
		}
	}
	controls := sortedKeys(rules)
	for leftIndex, leftControl := range controls {
		left := rules[leftControl]
		for _, rightControl := range controls[leftIndex+1:] {
			right := rules[rightControl]
			if ownershipPathsOverlap(left, right) {
				findings = appendFindings(findings, newFinding("OWN-407", fmt.Sprintf("%s+%s", leftControl, rightControl), "ownership paths overlap", "make shared-control ownership paths disjoint"))
			}
		}
	}
	return rules, findings
}

func ownershipPathsOverlap(left, right ownershipExpectation) bool {
	if left.pathKind == "exact" && right.pathKind == "exact" {
		return left.path == right.path
	}
	if left.pathKind == "prefix" && right.pathKind == "prefix" {
		return strings.HasPrefix(left.path, right.path) || strings.HasPrefix(right.path, left.path)
	}
	if left.pathKind == "prefix" {
		return strings.HasPrefix(right.path, left.path)
	}
	return strings.HasPrefix(left.path, right.path)
}

func crossCheckPathOwnership(pathRows []tsvRow, rules map[string]ownershipExpectation) []finding {
	var findings []finding
	for _, row := range pathRows {
		relative := row["path"]
		var matches []ownershipExpectation
		for _, rule := range rules {
			if (rule.pathKind == "exact" && rule.path == relative) || (rule.pathKind == "prefix" && strings.HasPrefix(relative, rule.path)) {
				matches = append(matches, rule)
			}
		}
		if len(matches) != 1 {
			findings = appendFindings(findings, newFinding("OWN-410", relative, fmt.Sprintf("path resolves to %d ownership rules, want one", len(matches)), "assign one disjoint accountable writer"))
			continue
		}
		if matches[0].writer != row["owner"] {
			findings = appendFindings(findings, newFinding("OWN-411", relative, "path-policy owner differs from shared-control writer", "restore one consistent writer"))
		}
	}
	return findings
}
