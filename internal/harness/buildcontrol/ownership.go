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

type orchestrationOwnershipExpectation struct {
	scope string
	owner string
}

var expectedOwnership = map[string]ownershipExpectation{
	"ci-workflow":                      {"exact", ".github/workflows/harness.yml", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"makefile":                         {"exact", "Makefile", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"readme":                           {"exact", "README.md", "wave-integrator", "owner-review", "owner+design"},
	"module":                           {"exact", "go.mod", "harness-integrator", "independent-readonly", "owner+module-decision"},
	"wave-approval":                    {"exact", "docs/artifacts/wave-01-approval.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-audit":                       {"exact", "docs/artifacts/wave-01-audit.md", "independent-reviewer", "owner-review", "separate-audit"},
	"wave-audit-remediation":           {"exact", "docs/artifacts/wave-01-audit-remediation.md", "wave-integrator", "independent-readonly", "fresh-independent-audit"},
	"wave-candidate":                   {"exact", "docs/artifacts/wave-01-candidate.sha256", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-contract":                    {"exact", "docs/artifacts/wave-01-change-contract.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-design":                      {"exact", "docs/artifacts/wave-01-design.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-design-amendment":            {"exact", "docs/artifacts/wave-01-design-amendment.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-evidence":                    {"exact", "docs/artifacts/wave-01-evidence.md", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-grant-amendment":             {"exact", "docs/artifacts/wave-01-grant-ladder-amendment.md", "wave-integrator", "independent-readonly", "separate-security-decision"},
	"wave-module-decision":             {"exact", "docs/artifacts/wave-01-module-identity-decision.md", "wave-integrator", "owner-review", "owner+module-decision"},
	"wave-specification":               {"exact", "docs/artifacts/wave-01-specification.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-02-approval":                 {"exact", "docs/artifacts/wave-02-approval.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-02-audit":                    {"exact", "docs/artifacts/wave-02-audit.md", "independent-reviewer", "owner-review", "separate-audit"},
	"wave-02-candidate":                {"exact", "docs/artifacts/wave-02-candidate.sha256", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-02-contract":                 {"exact", "docs/artifacts/wave-02-change-contract.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-02-design":                   {"exact", "docs/artifacts/wave-02-design.md", "wave-integrator", "owner-review", "owner+design"},
	"wave-02-evidence":                 {"exact", "docs/artifacts/wave-02-evidence.md", "wave-integrator", "independent-readonly", "owner+scope-audit"},
	"wave-02-specification":            {"exact", "docs/artifacts/wave-02-specification.md", "wave-integrator", "owner-review", "owner+design"},
	"concept-brief":                    {"exact", "docs/artifacts/concept-brief.md", "concept-owner", "owner-review", "owner+digest-approval"},
	"concept-discovery":                {"exact", "docs/artifacts/concept-discovery.md", "concept-owner", "independent-readonly", "bounded-public-research"},
	"concept-rebaseline-approval":      {"exact", "docs/artifacts/concept-rebaseline-approval.md", "wave-integrator", "owner-review", "owner+design"},
	"concept-rebaseline-contract":      {"exact", "docs/artifacts/concept-rebaseline-change-contract.md", "wave-integrator", "owner-review", "owner+design"},
	"concept-rebaseline-design":        {"exact", "docs/artifacts/concept-rebaseline-design.md", "wave-integrator", "owner-review", "owner+design"},
	"concept-rebaseline-specification": {"exact", "docs/artifacts/concept-rebaseline-specification.md", "wave-integrator", "owner-review", "owner+design"},
	"requirements-source":              {"exact", "docs/artifacts/requirements.md", "requirements-owner", "owner-review", "owner+requirements-decision"},
	"release-allocation":               {"exact", "docs/artifacts/feature-backlog.md", "backlog-owner", "owner-review", "owner+impact-decision"},
	"harness-data":                     {"prefix", "harness/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"harness-scripts":                  {"prefix", "scripts/harness/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"harness-code":                     {"prefix", "internal/harness/buildcontrol/", "harness-integrator", "independent-readonly", "owner+scope-audit"},
	"prototype-skills":                 {"prefix", "skills/", "protected-prototype-owner", "wave-07-owner", "wave-07-cutover"},
	"codex-manifest":                   {"prefix", ".codex-plugin/", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"claude-manifest":                  {"prefix", ".claude-plugin/", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"root-plugin":                      {"exact", "plugin.json", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"marketplace":                      {"exact", "marketplace.json", "protected-prototype-owner", "wave-10-owner", "wave-10-cutover"},
	"workflow-reference":               {"exact", "references/WORKFLOW.md", "protected-prototype-owner", "wave-07-owner", "wave-07-cutover"},
	"semantic-source":                  {"prefix", "semantic/", "semantic-owner", "independent-readonly", "owner+wave-02-design"},
	"semantic-schema":                  {"prefix", "schemas/semantic/", "semantic-owner", "independent-readonly", "owner+wave-02-design"},
	"evaluation-schema":                {"prefix", "schemas/evaluation/", "evaluator-owner", "independent-readonly", "separate-evaluator-freeze"},
	"artifact-schema":                  {"prefix", "schemas/artifact/", "state-owner", "independent-readonly", "future-wave"},
	"semantic-render":                  {"prefix", "internal/render/", "semantic-owner", "independent-readonly", "owner+wave-02-design"},
	"safety-policy":                    {"prefix", "internal/policy/", "safety-owner", "independent-readonly", "future-wave"},
	"context-safety":                   {"prefix", "internal/context/", "context-owner", "independent-readonly", "future-wave"},
	"transaction-plane":                {"prefix", "internal/transaction/", "effect-plane-owner", "independent-readonly", "future-wave"},
	"executor-plane":                   {"prefix", "internal/executor/", "effect-plane-owner", "independent-readonly", "future-wave"},
	"receipt-plane":                    {"prefix", "internal/receipt/", "effect-plane-owner", "independent-readonly", "future-wave"},
	"conductor-source":                 {"prefix", "internal/conductor/", "conductor-owner", "independent-readonly", "future-wave"},
	"codex-adapter":                    {"prefix", "internal/adapter/codex/", "codex-owner", "independent-readonly", "future-wave"},
	"claude-adapter":                   {"prefix", "internal/adapter/claude/", "claude-owner", "independent-readonly", "future-wave"},
	"semantic-fixtures":                {"prefix", "fixtures/public/bl-002/", "semantic-owner", "evaluator-owner", "evaluator-integration"},
	"evaluator-fixtures":               {"prefix", "fixtures/public/bl-003/", "evaluator-owner", "independent-readonly", "separate-evaluator-freeze"},
	"feature-fixtures":                 {"prefix", "fixtures/public/features/", "feature-owner", "evaluator-owner", "future-feature-integration"},
	"public-evaluator":                 {"prefix", "internal/evaluator/", "evaluator-owner", "independent-readonly", "separate-evaluator-freeze"},
	"generated-build":                  {"prefix", "build/generated/", "generator-owner", "independent-readonly", "future-wave"},
	"generated-packages":               {"prefix", "packages/", "generator-owner", "independent-readonly", "wave-10-cutover"},
	"distribution-source":              {"prefix", "internal/distribution/", "distribution-owner", "independent-readonly", "future-wave"},
	"updater":                          {"prefix", "cmd/l7up/", "updater-owner", "independent-readonly", "wave-10-cutover"},
	"protected-controls":               {"prefix", "protected/", "external-denied", "external-denied", "external-governance"},
}

var expectedOrchestrationOwnership = map[string]orchestrationOwnershipExpectation{
	"harness-build":        {"`go.mod`, `go.sum`, `vendor/`, `Makefile`, `.github/`, `harness/`, tool/dependency locks", "Harness/build integrator"},
	"semantic-contract":    {"Taxonomy, lifecycle, workflow/profile schema, obligation registry, prompt contract", "`BL-002` semantic owner"},
	"evaluator-governance": {"Evaluator protocol, truth schema, oracles, thresholds, coverage index", "`BL-003` evaluator-governance owner"},
	"feature-fixtures":     {"Feature public fixtures", "Feature owner in a disjoint backlog-ID directory; evaluator owner integrates frozen indexes"},
	"state-contract":       {"Canonical record schemas, migrations, digests, reducer/state contracts", "`BL-004` state owner"},
	"safety-contract":      {"Risk, effect, AP, policy, waiver, capability, grant, and guardrail contracts", "`BL-005` safety owner"},
	"context-safety":       {"Context/source/sink safety rules", "Safety/context owner; intake/presentation cannot redefine them"},
	"effect-plane":         {"Rooted transactions, executor, receipt, recovery", "Effect-plane owner under safety interfaces"},
	"conductor":            {"Conductor routing and prototype cutover", "`BL-007` conductor owner"},
	"codex-adapter":        {"Codex adapter/overlay", "`BL-012` Codex owner"},
	"claude-adapter":       {"Claude adapter/overlay", "`BL-013` Claude owner"},
	"generated":            {"Generated packages/indexes", "Generator/integration owner only"},
	"distribution":         {"Version/changelog/inventory/package lifecycle", "`BL-014` distribution owner"},
	"updater":              {"Privileged updater module, updater-owned channel code, locks, dependency graph, and separate CI", "`BL-014` updater owner; no root/core import and no shared core-module dependency lock"},
	"wave-records":         {"`docs/artifacts/` wave record and current status index", "Wave integration owner"},
	"prototype-assets":     {"Existing skills/manifests/`WORKFLOW.md`", "Protected until their approved cutover owner acts"},
	"protected-controls":   {"Protected cases, signing/promotion, AP2/AP3 roots, capability-grant issuers", "Outside candidate repository/agent authority"},
}

var orchestrationClassForControl = map[string]string{
	"ci-workflow": "harness-build", "makefile": "harness-build", "module": "harness-build", "harness-data": "harness-build", "harness-scripts": "harness-build", "harness-code": "harness-build",
	"requirements-source": "semantic-contract", "semantic-source": "semantic-contract", "semantic-schema": "semantic-contract", "semantic-render": "semantic-contract",
	"public-evaluator": "evaluator-governance", "evaluation-schema": "evaluator-governance", "evaluator-fixtures": "evaluator-governance", "semantic-fixtures": "feature-fixtures", "feature-fixtures": "feature-fixtures", "artifact-schema": "state-contract", "safety-policy": "safety-contract", "context-safety": "context-safety",
	"transaction-plane": "effect-plane", "executor-plane": "effect-plane", "receipt-plane": "effect-plane", "conductor-source": "conductor", "codex-adapter": "codex-adapter", "claude-adapter": "claude-adapter",
	"generated-build": "generated", "generated-packages": "generated", "distribution-source": "distribution", "updater": "updater",
	"readme": "wave-records", "release-allocation": "wave-records", "wave-approval": "wave-records", "wave-audit": "wave-records", "wave-audit-remediation": "wave-records", "wave-candidate": "wave-records", "wave-contract": "wave-records", "wave-design": "wave-records", "wave-design-amendment": "wave-records", "wave-evidence": "wave-records", "wave-grant-amendment": "wave-records", "wave-module-decision": "wave-records", "wave-specification": "wave-records", "wave-02-approval": "wave-records", "wave-02-audit": "wave-records", "wave-02-candidate": "wave-records", "wave-02-contract": "wave-records", "wave-02-design": "wave-records", "wave-02-evidence": "wave-records", "wave-02-specification": "wave-records", "concept-brief": "wave-records", "concept-discovery": "wave-records", "concept-rebaseline-approval": "wave-records", "concept-rebaseline-contract": "wave-records", "concept-rebaseline-design": "wave-records", "concept-rebaseline-specification": "wave-records",
	"prototype-skills": "prototype-assets", "codex-manifest": "prototype-assets", "claude-manifest": "prototype-assets", "root-plugin": "prototype-assets", "marketplace": "prototype-assets", "workflow-reference": "prototype-assets",
	"protected-controls": "protected-controls",
}

func checkOwnership(root string) (int, []finding) {
	rows, findings := loadTSV(root, "harness/control-ownership.tsv", []string{"control", "path_kind", "path", "writer", "reviewer", "change_gate"})
	rules, validationFindings := validateOwnershipRows(rows)
	findings = appendFindings(findings, validationFindings...)
	phase, phaseFindings := loadValidatedActivePhase(root)
	findings = appendFindings(findings, phaseFindings...)
	pathRows, pathFindings := loadTSV(root, phase.pathPolicy, []string{"change", "path", "owner", "rule"})
	findings = appendFindings(findings, pathFindings...)
	findings = appendFindings(findings, crossCheckPathOwnership(pathRows, rules)...)
	orchestrationData, orchestrationFindings := readStrictFile(root, "docs/artifacts/orchestration-plan.md")
	findings = appendFindings(findings, orchestrationFindings...)
	if len(orchestrationFindings) == 0 {
		findings = appendFindings(findings, crossCheckOrchestrationOwnership(string(orchestrationData), rules)...)
	}
	return len(rules), findings
}

func crossCheckOrchestrationOwnership(document string, rules map[string]ownershipExpectation) []finding {
	actual := make(map[string]orchestrationOwnershipExpectation)
	inOwnershipSection := false
	for _, line := range strings.Split(document, "\n") {
		if line == "## 10. Shared-file ownership" {
			inOwnershipSection = true
			continue
		}
		if inOwnershipSection && strings.HasPrefix(line, "## 11. ") {
			break
		}
		if !inOwnershipSection {
			continue
		}
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) != 2 || cells[0] == "Scope" || strings.HasPrefix(cells[0], "---") {
			continue
		}
		matchedClass := ""
		for class, expected := range expectedOrchestrationOwnership {
			if cells[0] == expected.scope {
				matchedClass = class
				break
			}
		}
		if matchedClass == "" {
			actual["unknown:"+cells[0]] = orchestrationOwnershipExpectation{cells[0], cells[1]}
			continue
		}
		actual[matchedClass] = orchestrationOwnershipExpectation{cells[0], cells[1]}
	}

	var findings []finding
	if !inOwnershipSection {
		findings = appendFindings(findings, newFinding("OWN-420", "orchestration-plan.md", "shared-file ownership section is missing", "restore the authoritative ownership table"))
	}
	for class, expected := range expectedOrchestrationOwnership {
		if actual[class] != expected {
			findings = appendFindings(findings, newFinding("OWN-421", class, "authoritative orchestration ownership differs from the approved source", "restore or approve the source ownership rule"))
		}
	}
	for class := range actual {
		if strings.HasPrefix(class, "unknown:") {
			findings = appendFindings(findings, newFinding("OWN-422", class, "authoritative orchestration table contains an unmapped ownership class", "map the new class through an approved ownership change"))
		}
	}
	covered := make(map[string]bool)
	for control := range rules {
		class, ok := orchestrationClassForControl[control]
		if !ok {
			findings = appendFindings(findings, newFinding("OWN-423", control, "control has no authoritative orchestration ownership class", "bind the control to one source ownership class"))
			continue
		}
		covered[class] = true
	}
	for class := range expectedOrchestrationOwnership {
		if !covered[class] {
			findings = appendFindings(findings, newFinding("OWN-424", class, "authoritative orchestration ownership class has no local control", "add one disjoint local ownership mapping"))
		}
	}
	return findings
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
