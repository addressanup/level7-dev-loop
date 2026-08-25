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
	surface      string
	currentState string
	v1Ceiling    string
	claimState   string
	owner        string
}

type dispositionExpectation struct {
	disposition string
	targetOwner string
	cutover     string
}

var expectedSupport = map[string]supportExpectation{
	"codex-advisory":     {"Codex-advisory-package", "prototype", "A0", "withheld", "L7-BL-012"},
	"claude-advisory":    {"Claude-advisory-package", "prototype", "A0", "withheld", "L7-BL-013"},
	"controlled-client":  {"Level-7-Controlled-Client", "planned", "A0-A2-gated", "withheld", "L7-BL-005"},
	"proof-generic":      {"generic-change-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"proof-feature":      {"feature-behavior-change-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"proof-refactor":     {"behavior-preserving-refactor-profile", "planned", "A0-A2-gated", "withheld", "L7-BL-009"},
	"a3-a4-handoff":      {"Package-Deploy-Expose-handoff", "planned", "plan-handoff-only", "withheld", "L7-BL-011"},
	"a5-autonomy":        {"background-self-modifying-remediation", "absent", "none", "excluded", "L7-BL-035"},
	"dual-host-support":  {"Codex-and-Claude-support", "unproved", "none", "withheld", "L7-BL-014"},
	"stable-version-1.0": {"stable-1.0-release", "unproved", "none", "withheld", "L7-BL-042"},
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
	supportRows, findings := loadTSV(root, "harness/support-matrix.tsv", []string{"id", "surface", "current_state", "v1_ceiling", "claim_state", "owner"})
	dispositionRows, dispositionLoadFindings := loadTSV(root, "harness/prototype-dispositions.tsv", []string{"skill", "disposition", "target_owner", "cutover"})
	findings = appendFindings(findings, dispositionLoadFindings...)
	if len(findings) != 0 {
		return 0, findings
	}
	findings = appendFindings(findings, validateSupportRows(supportRows)...)
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
		actual := supportExpectation{row["surface"], row["current_state"], row["v1_ceiling"], row["claim_state"], row["owner"]}
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
