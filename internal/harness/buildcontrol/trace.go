package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var summaryAllocationPattern = regexp.MustCompile(`^\*\*V1\.0 ([0-9]+) / V1\.x ([0-9]+) / Later ([0-9]+)\*\*$`)

type ownership struct {
	owner      string
	allocation string
}

type traceResult struct {
	total       int
	allocations map[string]int
}

func checkTrace(root string) (traceResult, []finding) {
	requirementsData, findings := readStrictFile(root, "docs/artifacts/requirements.md")
	backlogData, backlogFindings := readStrictFile(root, "docs/artifacts/feature-backlog.md")
	findings = appendFindings(findings, backlogFindings...)
	if len(findings) != 0 {
		return traceResult{}, findings
	}
	definitions, definitionFindings := parseRequirementDefinitions(string(requirementsData))
	owners, ownerFindings := parseRequirementOwnership(string(backlogData))
	findings = appendFindings(findings, definitionFindings...)
	findings = appendFindings(findings, ownerFindings...)
	result, validationFindings := validateTrace(definitions, owners)
	findings = appendFindings(findings, validationFindings...)
	return result, findings
}

func parseRequirementDefinitions(document string) (map[string]struct{}, []finding) {
	definitions := make(map[string]struct{})
	var findings []finding
	inNormativeRegion := false
	for lineIndex, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, "## 9. Functional requirements") {
			inNormativeRegion = true
			continue
		}
		if inNormativeRegion && strings.HasPrefix(line, "## 11.") {
			break
		}
		if !inNormativeRegion {
			continue
		}
		cell, ok := firstMarkdownCell(line)
		if !ok || !strings.Contains(cell, "L7-") {
			continue
		}
		id, ok := unquoteCodeCell(cell)
		if !ok || !requirementIDPattern.MatchString(id) {
			findings = appendFindings(findings, newFinding("TRACE-101", fmt.Sprintf("requirements.md:%d", lineIndex+1), "malformed normative requirement ID cell", "restore one exact backticked requirement ID"))
			continue
		}
		if _, duplicate := definitions[id]; duplicate {
			findings = appendFindings(findings, newFinding("TRACE-102", id, "duplicate normative requirement definition", "retain exactly one authoritative definition"))
			continue
		}
		if len(definitions) >= maxExpandedRequirementIDs {
			findings = appendFindings(findings, newFinding("TRACE-103", "requirements.md", "normative requirement definitions exceed the cumulative limit", "narrow the authoritative requirement set"))
			break
		}
		definitions[id] = struct{}{}
	}
	if !inNormativeRegion {
		findings = appendFindings(findings, newFinding("TRACE-100", "requirements.md", "normative requirement section is missing", "restore the approved section heading"))
	}
	return definitions, findings
}

func parseRequirementOwnership(document string) (map[string]ownership, []finding) {
	owners := make(map[string]ownership)
	var findings []finding
	inOwnershipRegion := false
	expandedCount := 0
	summarySeen := false
	declaredSummary := traceResult{allocations: map[string]int{}}
	for lineIndex, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, "## 8. Normative requirement ownership and release allocation") {
			inOwnershipRegion = true
			continue
		}
		if inOwnershipRegion && strings.HasPrefix(line, "## 9.") {
			break
		}
		if !inOwnershipRegion {
			continue
		}
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) == 0 || cells[0] == "Requirement IDs" || strings.HasPrefix(cells[0], "---") {
			continue
		}
		if cells[0] == "**Total**" {
			summarySeen = true
			if len(cells) != 4 {
				findings = appendFindings(findings, newFinding("TRACE-125", "feature-backlog.md", "summary row must contain exactly four cells", "restore the source-derived summary row"))
				continue
			}
			match := summaryAllocationPattern.FindStringSubmatch(cells[2])
			declaredTotal, totalErr := strconv.Atoi(strings.Trim(cells[3], "*"))
			if match == nil || totalErr != nil {
				findings = appendFindings(findings, newFinding("TRACE-125", "feature-backlog.md", "summary row is malformed", "restore the source-derived allocation and total summary"))
				continue
			}
			declaredSummary.allocations["V1.0"], _ = strconv.Atoi(match[1])
			declaredSummary.allocations["V1.x"], _ = strconv.Atoi(match[2])
			declaredSummary.allocations["Later"], _ = strconv.Atoi(match[3])
			declaredSummary.total = declaredTotal
			continue
		}
		if !strings.Contains(cells[0], "L7-") {
			continue
		}
		if len(cells) != 4 {
			findings = appendFindings(findings, newFinding("TRACE-120", fmt.Sprintf("feature-backlog.md:%d", lineIndex+1), "ownership row must contain exactly four cells", "restore the approved ownership row shape"))
			continue
		}
		ids, expressionFindings := expandRequirementExpressionBounded(cells[0], maxExpandedRequirementIDs-expandedCount)
		findings = appendFindings(findings, expressionFindings...)
		if len(expressionFindings) != 0 {
			continue
		}
		expandedCount += len(ids)
		owner, ownerOK := unquoteCodeCell(cells[1])
		if !ownerOK || !backlogOwnerPattern.MatchString(owner) {
			findings = appendFindings(findings, newFinding("TRACE-121", cells[1], "malformed accountable backlog owner", "use one exact L7-BL-### owner"))
			continue
		}
		allocation := cells[2]
		if allocation != "V1.0" && allocation != "V1.x" && allocation != "Later" {
			findings = appendFindings(findings, newFinding("TRACE-122", allocation, "unknown release allocation", "use V1.0, V1.x, or Later"))
			continue
		}
		declaredCount, err := strconv.Atoi(strings.Trim(cells[3], "*"))
		if err != nil || declaredCount != len(ids) {
			findings = appendFindings(findings, newFinding("TRACE-123", cells[0], "declared count does not match the expanded IDs", "correct the expression or count"))
		}
		for _, id := range ids {
			if previous, duplicate := owners[id]; duplicate {
				findings = appendFindings(findings, newFinding("TRACE-124", id, fmt.Sprintf("multiple accountable owners: %s and %s", previous.owner, owner), "retain exactly one accountable owner"))
				continue
			}
			owners[id] = ownership{owner: owner, allocation: allocation}
		}
	}
	if !inOwnershipRegion {
		findings = appendFindings(findings, newFinding("TRACE-119", "feature-backlog.md", "ownership section is missing", "restore the approved section heading"))
	}
	if !summarySeen {
		findings = appendFindings(findings, newFinding("TRACE-125", "feature-backlog.md", "ownership summary row is missing", "restore the source-derived summary row"))
	} else if declaredSummary.total != len(owners) {
		findings = appendFindings(findings, newFinding("TRACE-125", "total", fmt.Sprintf("displayed summary is %d, derived total is %d", declaredSummary.total, len(owners)), "correct the displayed source summary"))
	}
	derivedAllocations := map[string]int{"V1.0": 0, "V1.x": 0, "Later": 0}
	for _, owner := range owners {
		derivedAllocations[owner.allocation]++
	}
	for allocation, derived := range derivedAllocations {
		if declaredSummary.allocations[allocation] != derived {
			findings = appendFindings(findings, newFinding("TRACE-125", allocation, fmt.Sprintf("displayed summary is %d, derived allocation is %d", declaredSummary.allocations[allocation], derived), "correct the displayed source summary"))
		}
	}
	return owners, findings
}

func validateTrace(definitions map[string]struct{}, owners map[string]ownership) (traceResult, []finding) {
	result := traceResult{total: len(definitions), allocations: map[string]int{"V1.0": 0, "V1.x": 0, "Later": 0}}
	var findings []finding
	for id := range definitions {
		owner, ok := owners[id]
		if !ok {
			findings = appendFindings(findings, newFinding("TRACE-130", id, "normative requirement has no accountable owner", "add exactly one approved ownership record"))
			continue
		}
		result.allocations[owner.allocation]++
	}
	for id := range owners {
		if _, ok := definitions[id]; !ok {
			findings = appendFindings(findings, newFinding("TRACE-131", id, "ownership map contains an unknown requirement", "remove or define the unknown ID through an approved change"))
		}
	}
	if len(definitions) != 163 {
		findings = appendFindings(findings, newFinding("TRACE-132", "requirements", fmt.Sprintf("derived total is %d, want 163", len(definitions)), "restore the approved normative requirement set"))
	}
	expected := map[string]int{"V1.0": 140, "V1.x": 18, "Later": 5}
	for allocation, want := range expected {
		if got := result.allocations[allocation]; got != want {
			findings = appendFindings(findings, newFinding("TRACE-133", allocation, fmt.Sprintf("derived allocation is %d, want %d", got, want), "restore the approved release allocation"))
		}
	}
	return result, findings
}
