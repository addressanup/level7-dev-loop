package main

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type riskTier int

const (
	tierRoutine  riskTier = 1
	tierProduct  riskTier = 2
	tierHighRisk riskTier = 3
)

type assuranceMode string

const (
	assuranceSolo assuranceMode = "solo"
	assuranceTeam assuranceMode = "team"
)

func (mode assuranceMode) valid() bool {
	return mode == assuranceSolo || mode == assuranceTeam
}

type workflowState string

const (
	statePlanned                  workflowState = "planned"
	stateAwaitingOwnerApproval    workflowState = "awaiting-owner-approval"
	stateBuilding                 workflowState = "building"
	stateVerified                 workflowState = "verified"
	stateAwaitingIndependentAudit workflowState = "awaiting-independent-audit"
	stateReviewed                 workflowState = "reviewed"
	stateReady                    workflowState = "ready"
)

type changeBrief struct {
	ID         string
	Tier       riskTier
	BaseCommit string
	Path       string
	Scope      []string
}

type evidenceRecord struct {
	ChangeID        string
	CandidateCommit string
	CandidateTree   string
	Result          string
	Reviewer        string
}

type approvalEnvelope struct {
	Schema      int    `json:"schema"`
	ChangeID    string `json:"change_id"`
	Actor       string `json:"actor"`
	Implementer string `json:"implementer"`
	BriefCommit string `json:"brief_commit"`
	Source      string `json:"source"`
}

type auditEnvelope struct {
	Schema          int    `json:"schema"`
	ChangeID        string `json:"change_id"`
	Actor           string `json:"actor"`
	CandidateCommit string `json:"candidate_commit"`
	AuditCommit     string `json:"audit_commit"`
	Source          string `json:"source"`
}

var codeSpanPattern = regexp.MustCompile("`([^`]+)`")

func parseBrief(relative string, data []byte) (changeBrief, []finding) {
	document := string(data)
	fields := parseMarkdownTable(document)
	tierText := fields["Risk tier"]
	if index := strings.Index(tierText, " "); index >= 0 {
		tierText = tierText[:index]
	}
	tierValue, tierErr := strconv.Atoi(strings.Trim(tierText, "`"))
	brief := changeBrief{
		ID:         strings.Trim(fields["Change ID"], "`"),
		Tier:       riskTier(tierValue),
		BaseCommit: strings.Trim(fields["Base commit"], "`"),
		Path:       relative,
	}
	var findings []finding
	if brief.ID == "" || relative != "docs/artifacts/changes/"+brief.ID+".md" {
		findings = appendFindings(findings, newFinding("BRIEF-001", relative, "change ID is missing or does not match the filename", "use one stable lowercase change ID"))
	}
	if tierErr != nil || brief.Tier < tierProduct || brief.Tier > tierHighRisk {
		findings = appendFindings(findings, newFinding("BRIEF-002", relative, "change brief must declare risk tier 2 or 3", "declare the proportionate risk tier"))
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(brief.BaseCommit) {
		findings = appendFindings(findings, newFinding("BRIEF-003", relative, "base commit is not a full Git commit ID", "record the exact Git base commit"))
	}
	for _, heading := range []string{"## Problem", "## Scope", "## Acceptance criteria", "## Risks and mitigations", "## Rollback"} {
		if !hasExactHeading(document, heading) {
			findings = appendFindings(findings, newFinding("BRIEF-004", relative, "required section is missing: "+heading, "complete the one concise change brief"))
		}
	}
	brief.Scope = parseImplementationScope(document)
	brief.Scope = append(brief.Scope, relative)
	brief.Scope = uniqueSorted(brief.Scope)
	if len(brief.Scope) < 2 {
		findings = appendFindings(findings, newFinding("BRIEF-005", relative, "declared implementation scope is empty", "list exact files or bounded glob patterns"))
	}
	return brief, findings
}

func parseMarkdownTable(document string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(document, "\n") {
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}
		parts := strings.Split(line[1:len(line)-1], "|")
		if len(parts) != 2 {
			continue
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key != "" && key != "Field" && !strings.HasPrefix(key, "---") {
			fields[key] = value
		}
	}
	return fields
}

func parseImplementationScope(document string) []string {
	section := sectionBetween(document, "## Exact implementation file set", "## Acceptance criteria")
	var result []string
	for _, line := range strings.Split(section, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "-") {
			continue
		}
		lastDirectory := ""
		for _, match := range codeSpanPattern.FindAllStringSubmatch(line, -1) {
			candidate := match[1]
			if !strings.Contains(candidate, "/") && lastDirectory != "" {
				candidate = path.Join(lastDirectory, candidate)
			}
			if strings.Contains(candidate, "/") {
				lastDirectory = path.Dir(candidate)
			}
			if safeScopePattern(candidate) {
				result = append(result, candidate)
			}
		}
	}
	return uniqueSorted(result)
}

func parseEvidence(data []byte) evidenceRecord {
	fields := parseMarkdownTable(string(data))
	return evidenceRecord{
		ChangeID:        strings.Trim(fields["Change ID"], "`"),
		CandidateCommit: strings.Trim(fields["Candidate commit"], "`"),
		CandidateTree:   strings.Trim(fields["Candidate tree"], "`"),
		Result:          strings.Trim(fields["Result"], "`"),
		Reviewer:        strings.Trim(fields["Reviewer"], "`"),
	}
}

func hasExactHeading(document, heading string) bool {
	for _, line := range strings.Split(document, "\n") {
		if line == heading {
			return true
		}
	}
	return false
}

func sectionBetween(document, start, end string) string {
	startIndex := strings.Index(document, start+"\n")
	if startIndex < 0 {
		return ""
	}
	section := document[startIndex+len(start)+1:]
	if endIndex := strings.Index(section, end+"\n"); endIndex >= 0 {
		section = section[:endIndex]
	}
	return section
}

func safeScopePattern(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "..") || strings.Contains(value, "\\") {
		return false
	}
	if strings.Count(value, "*") > 1 || (strings.Contains(value, "*") && !strings.Contains(value, "/*/")) {
		return false
	}
	return true
}

func scopeContains(scope []string, relative string) bool {
	for _, pattern := range scope {
		if pattern == relative {
			return true
		}
		if strings.Contains(pattern, "*") {
			matched, _ := path.Match(pattern, relative)
			if matched {
				return true
			}
		}
	}
	return false
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func nextState(tier riskTier, assurance assuranceMode, state workflowState) (workflowState, string, bool) {
	if tier == tierHighRisk && assurance == assuranceTeam {
		switch state {
		case statePlanned:
			return stateAwaitingOwnerApproval, "request explicit accountable-owner approval", true
		case stateAwaitingOwnerApproval:
			return stateBuilding, "obtain explicit accountable-owner approval, then implement the approved scope", true
		case stateBuilding:
			return stateVerified, "implement or remediate the approved scope, then run verification bound to the Git candidate", true
		case stateVerified:
			return stateAwaitingIndependentAudit, "request an independent read-only audit", true
		case stateAwaitingIndependentAudit:
			return stateReviewed, "record the bound independent decision", true
		case stateReviewed:
			return stateReady, "confirm merge readiness", true
		case stateReady:
			return stateReady, "merge the reviewed Git candidate", true
		}
	}
	if assurance == assuranceSolo {
		switch state {
		case statePlanned:
			return stateBuilding, "implement the declared scope", true
		case stateBuilding:
			return stateVerified, "implement or remediate the declared scope, then run relevant tests and CI", true
		case stateVerified:
			return stateReviewed, "self-review the exact candidate without claiming independence", true
		case stateReviewed:
			return stateReady, "confirm merge readiness", true
		case stateReady:
			return stateReady, "merge the reviewed Git candidate", true
		}
	}
	switch state {
	case statePlanned:
		return stateBuilding, "implement the declared scope", true
	case stateBuilding:
		return stateVerified, "implement or remediate the declared scope, then run relevant tests and CI", true
	case stateVerified:
		return stateReviewed, "obtain normal review", true
	case stateReviewed:
		return stateReady, "confirm merge readiness", true
	case stateReady:
		return stateReady, "merge the reviewed Git candidate", true
	}
	return "", "", false
}

func validateTransition(tier riskTier, assurance assuranceMode, from, to workflowState) bool {
	next, _, ok := nextState(tier, assurance, from)
	if ok && next == to {
		return true
	}
	return (from == stateVerified || from == stateReviewed || from == stateAwaitingIndependentAudit) && to == stateBuilding
}

func (tier riskTier) String() string { return fmt.Sprintf("%d", tier) }
