package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxSkillEntries = 64

type supportExpectation struct {
	recordVersion string
	surface       string
	currentState  string
	v1Ceiling     string
	claimState    string
	owner         string
}

type dispositionExpectation struct {
	disposition string
	targetOwner string
	cutover     string
}

var expectedSupport = map[string]supportExpectation{
	"codex-advisory":          {"wave-01-v1", "Codex-advisory-package", "prototype", "A0", "withheld", "L7-BL-012"},
	"claude-advisory":         {"wave-01-v1", "Claude-advisory-package", "prototype", "A0", "withheld", "L7-BL-013"},
	"controlled-client":       {"wave-01-v1", "separately-installed-Level-7-Controlled-Client", "planned", "A0-A2-gated", "withheld", "L7-BL-005"},
	"proof-generic":           {"wave-01-v1", "generic-change-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"proof-feature":           {"wave-01-v1", "feature-behavior-change-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"proof-refactor":          {"wave-01-v1", "behavior-preserving-refactor-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"a3-a4-handoff":           {"wave-01-v1", "Package-Deploy-Expose-handoff", "planned", "plan-handoff-only", "withheld", "L7-BL-011"},
	"a5-autonomy":             {"wave-01-v1", "background-self-modifying-remediation", "absent", "none", "excluded", "L7-BL-035"},
	"dual-host-support":       {"wave-01-v1", "Codex-and-Claude-support", "unproved", "none", "withheld", "L7-BL-014"},
	"stable-version-1.0":      {"wave-01-v1", "stable-1.0-release", "unproved", "none", "withheld", "L7-BL-042"},
	"workspace-boundary":      {"wave-01-v1", "one-local-repository-worktree", "development-only", "one-worktree", "required", "L7-BL-001"},
	"plugin-authority":        {"wave-01-v1", "plugin-installation", "insufficient", "no-mutation-authority", "withheld", "L7-BL-005"},
	"development-evidence":    {"wave-01-v1", "same-user-local-development-evidence", "available", "non-promoting", "unqualified", "L7-BL-010"},
	"release-blocking-proof":  {"wave-01-v1", "release-blocking-proof", "absent", "required-before-release", "withheld", "L7-BL-042"},
	"priority-p0":             {"wave-01-v1", "P0-v1.0-scope", "approved", "release-blocking", "source-bound", "L7-BL-001"},
	"priority-p1":             {"wave-01-v1", "P1-v1.x-scope", "approved", "post-v1.0", "source-bound", "L7-BL-001"},
	"priority-p2":             {"wave-01-v1", "P2-later-scope", "approved", "no-commitment", "source-bound", "L7-BL-001"},
	"priority-change-control": {"wave-01-v1", "scope-priority-change", "approval-required", "impact-diff+accountable-approval", "source-bound", "L7-BL-001"},
}

type priorityExpectation struct {
	meaning          string
	releaseTreatment string
}

var expectedPriority = map[string]priorityExpectation{
	"P0": {"Required for the v1.0 product promise or its safety/release boundary.", "A missing or failing P0 item blocks v1.0. Safety-critical failures cannot be waived by an aggregate score."},
	"P1": {"A v1.x product-family increment after the v1.0 slice proves usable and safe.", "Prioritized with pilot evidence; not smuggled into v1.0 through a broad skill."},
	"P2": {"Later candidate needing discovery, new authority, or a separate autonomy/product charter.", "No delivery commitment. Promotion requires explicit owner approval and re-risking."},
}

var requiredPriorityChangeRules = []string{
	"Priority is a mutable field; stable backlog IDs do not encode it. An item may move priority only with a recorded rationale, requirement-impact check, and approval. A P0 safety prerequisite cannot be demoted merely to meet a date.",
	"Scope or priority changes produce an impact diff and accountable approval; dates or aggregate metrics cannot waive a safety prerequisite.",
}

var expectedDispositions = map[string]dispositionExpectation{
	"l7-build":        {"replace", "L7-BL-007+L7-BL-008+L7-BL-009", "wave-08"},
	"l7-change":       {"replace", "L7-BL-011", "wave-09"},
	"l7-constitution": {"replace", "L7-BL-002+L7-BL-007", "wave-07"},
	"l7-deploy":       {"exclude", "L7-BL-011", "wave-09"},
	"l7-experience":   {"exclude", "L7-BL-019", "post-v1.0"},
	"l7-geometry":     {"deprecate", "L7-BL-019", "post-v1.0"},
	"l7-greenfield":   {"exclude", "L7-BL-016", "post-v1.0"},
	"l7-next":         {"conform", "L7-BL-007", "wave-07"},
	"l7-ops":          {"exclude", "L7-BL-022", "post-v1.0"},
	"l7-release":      {"replace", "L7-BL-010+L7-BL-042", "wave-13"},
	"l7-review":       {"replace", "L7-BL-010", "wave-09"},
	"l7-storybook":    {"exclude", "L7-BL-029", "post-v1.0"},
}

func checkClaims(root string) (int, []finding) {
	supportRows, findings := loadTSV(root, "harness/support-matrix.tsv", []string{"record_version", "id", "surface", "current_state", "v1_ceiling", "claim_state", "owner"})
	dispositionRows, dispositionLoadFindings := loadTSV(root, "harness/prototype-dispositions.tsv", []string{"skill", "disposition", "target_owner", "cutover"})
	findings = appendFindings(findings, dispositionLoadFindings...)
	if len(findings) != 0 {
		return 0, findings
	}
	findings = appendFindings(findings, validateSupportRows(supportRows)...)
	backlogData, backlogFindings := readStrictFile(root, "docs/artifacts/feature-backlog.md")
	findings = appendFindings(findings, backlogFindings...)
	if len(backlogFindings) == 0 {
		findings = appendFindings(findings, validatePriorityContract(string(backlogData))...)
	}
	inventory, inventoryFindings := loadSkillInventory(root)
	findings = appendFindings(findings, inventoryFindings...)
	findings = appendFindings(findings, validateDispositionRows(dispositionRows, inventory)...)
	return len(inventory), findings
}

func validateSupportRows(rows []tsvRow) []finding {
	seen := make(map[string]bool)
	var findings []finding
	for _, row := range rows {
		id := row["id"]
		if seen[id] {
			findings = appendFindings(findings, newFinding("CLAIM-201", id, "duplicate support-matrix ID", "retain exactly one row"))
			continue
		}
		seen[id] = true
		expected, ok := expectedSupport[id]
		if !ok {
			findings = appendFindings(findings, newFinding("CLAIM-202", id, "unknown support-matrix ID", "remove or approve the new claim surface"))
			continue
		}
		actual := supportExpectation{row["record_version"], row["surface"], row["current_state"], row["v1_ceiling"], row["claim_state"], row["owner"]}
		if actual != expected {
			findings = appendFindings(findings, newFinding("CLAIM-203", id, "support row differs from the approved Wave 1 claim", "restore the approved claim or obtain a new impact decision"))
		}
	}
	for id := range expectedSupport {
		if !seen[id] {
			findings = appendFindings(findings, newFinding("CLAIM-204", id, "required support-matrix row is missing", "restore the required row"))
		}
	}
	return findings
}

func validatePriorityContract(document string) []finding {
	actual := make(map[string]priorityExpectation)
	inPrioritySection := false
	for _, line := range strings.Split(document, "\n") {
		if line == "### 2.1 Priority" {
			inPrioritySection = true
			continue
		}
		if inPrioritySection && strings.HasPrefix(line, "### 2.2 ") {
			break
		}
		if !inPrioritySection {
			continue
		}
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) != 3 {
			continue
		}
		priority, ok := unquoteCodeCell(cells[0])
		if !ok {
			continue
		}
		if _, expected := expectedPriority[priority]; expected {
			actual[priority] = priorityExpectation{cells[1], cells[2]}
		}
	}
	var findings []finding
	for priority, expected := range expectedPriority {
		if actual[priority] != expected {
			findings = appendFindings(findings, newFinding("CLAIM-205", priority, "backlog priority semantics differ from the approved contract", "restore the source-bound P0/P1/P2 rule"))
		}
	}
	for _, rule := range requiredPriorityChangeRules {
		if !strings.Contains(document, rule) {
			findings = appendFindings(findings, newFinding("CLAIM-206", "priority-change-control", "backlog scope or priority approval rule is missing", "restore the source-bound impact and approval rule"))
		}
	}
	return findings
}

func loadSkillInventory(root string) ([]string, []finding) {
	directory, err := os.Open(filepath.Join(root, "skills"))
	if err != nil {
		return nil, []finding{newFinding("CLAIM-210", "skills", err.Error(), "restore the protected skill inventory")}
	}
	entries, readErr := directory.ReadDir(maxSkillEntries + 1)
	closeErr := directory.Close()
	if readErr != nil && readErr != io.EOF {
		return nil, []finding{newFinding("CLAIM-210", "skills", readErr.Error(), "restore the protected skill inventory")}
	}
	if closeErr != nil {
		return nil, []finding{newFinding("CLAIM-210", "skills", closeErr.Error(), "restore the protected skill inventory")}
	}
	if len(entries) > maxSkillEntries {
		return nil, []finding{newFinding("CLAIM-213", "skills", "skill directory exceeds the entry limit", "narrow the protected skill inventory")}
	}
	var inventory []string
	var findings []finding
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		relative := filepath.ToSlash(filepath.Join("skills", entry.Name(), "SKILL.md"))
		data, readFindings := readStrictFile(root, relative)
		if len(readFindings) != 0 {
			findings = appendFindings(findings, readFindings...)
			continue
		}
		name, invocable, parseFindings := parseSkillFrontmatter(relative, string(data))
		findings = appendFindings(findings, parseFindings...)
		if invocable {
			inventory = append(inventory, name)
		}
	}
	sort.Strings(inventory)
	return inventory, findings
}

func parseSkillFrontmatter(name, document string) (string, bool, []finding) {
	lines := strings.Split(document, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", false, []finding{newFinding("CLAIM-211", name, "skill frontmatter is missing", "restore the protected skill frontmatter")}
	}
	var skillName string
	invocable := false
	closed := false
	for _, line := range lines[1:] {
		if line == "---" {
			closed = true
			break
		}
		if strings.HasPrefix(line, "name: ") {
			skillName = strings.TrimSpace(strings.TrimPrefix(line, "name: "))
		}
		if line == "user-invocable: true" {
			invocable = true
		}
	}
	if !closed || skillName == "" {
		return "", false, []finding{newFinding("CLAIM-212", name, "skill frontmatter is incomplete", "restore the exact skill name and closing delimiter")}
	}
	return skillName, invocable, nil
}

func validateDispositionRows(rows []tsvRow, inventory []string) []finding {
	inventorySet := make(map[string]bool, len(inventory))
	for _, skill := range inventory {
		inventorySet[skill] = true
	}
	seen := make(map[string]bool)
	var findings []finding
	for _, row := range rows {
		skill := row["skill"]
		if seen[skill] {
			findings = appendFindings(findings, newFinding("CLAIM-220", skill, "duplicate prototype disposition", "retain exactly one disposition"))
			continue
		}
		seen[skill] = true
		if !inventorySet[skill] {
			findings = appendFindings(findings, newFinding("CLAIM-221", skill, "disposition names an unknown or non-invocable skill", "remove or approve the inventory change"))
			continue
		}
		expected, ok := expectedDispositions[skill]
		if !ok {
			findings = appendFindings(findings, newFinding("CLAIM-222", skill, "skill has no approved Wave 1 disposition", "add an approved disposition"))
			continue
		}
		actual := dispositionExpectation{row["disposition"], row["target_owner"], row["cutover"]}
		if actual != expected {
			findings = appendFindings(findings, newFinding("CLAIM-223", skill, "prototype disposition differs from the approved migration plan", "restore the approved disposition or obtain a new impact decision"))
		}
	}
	for _, skill := range inventory {
		if !seen[skill] {
			findings = appendFindings(findings, newFinding("CLAIM-224", skill, "user-invocable skill has no disposition", "add exactly one disposition"))
		}
	}
	if len(inventory) != 12 {
		findings = appendFindings(findings, newFinding("CLAIM-225", "prototype-inventory", fmt.Sprintf("derived user-invocable skill count is %d, want 12", len(inventory)), "restore or approve the prototype inventory"))
	}
	return findings
}
