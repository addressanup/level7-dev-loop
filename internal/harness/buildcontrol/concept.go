package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	conceptBaseCommit         = "34c3ba94e3f1042975761f02286c37723c84b68e"
	conceptBaseTree           = "e2e292bfbeb28420c48c06773538434d07278a42"
	conceptBaseManifestSHA256 = "eed041b0cbaedbf17d2acecba8f56442d7e1a0cea1c8ecb4c78d3269bd3e96c4"
	conceptDossierPath        = "docs/artifacts/concept-discovery.md"
	conceptBriefPath          = "docs/artifacts/concept-brief.md"
	maxResearchMinutes        = 240
	maxMaterialSources        = 25
)

var expectedConceptPaths = map[string]pathExpectation{
	".github/workflows/harness.yml":                        {"modify", "harness-integrator", "SCOPE-620"},
	"README.md":                                            {"modify", "wave-integrator", "SCOPE-620"},
	"docs/artifacts/concept-brief.md":                      {"add", "concept-owner", "SCOPE-621"},
	"docs/artifacts/concept-discovery.md":                  {"add", "concept-owner", "SCOPE-621"},
	"docs/artifacts/concept-rebaseline-approval.md":        {"add", "wave-integrator", "SCOPE-621"},
	"docs/artifacts/concept-rebaseline-change-contract.md": {"add", "wave-integrator", "SCOPE-621"},
	"docs/artifacts/concept-rebaseline-design.md":          {"add", "wave-integrator", "SCOPE-621"},
	"docs/artifacts/concept-rebaseline-specification.md":   {"add", "wave-integrator", "SCOPE-621"},
	"harness/concept-discovery-base.sha256":                {"add", "harness-integrator", "SCOPE-621"},
	"harness/concept-discovery-paths.tsv":                  {"add", "harness-integrator", "SCOPE-621"},
	"harness/control-ownership.tsv":                        {"modify", "harness-integrator", "SCOPE-620"},
	"harness/phases.tsv":                                   {"modify", "harness-integrator", "SCOPE-620"},
	"internal/harness/buildcontrol/concept.go":             {"add", "harness-integrator", "SCOPE-621"},
	"internal/harness/buildcontrol/concept_test.go":        {"add", "harness-integrator", "SCOPE-621"},
	"internal/harness/buildcontrol/main.go":                {"modify", "harness-integrator", "SCOPE-620"},
	"internal/harness/buildcontrol/ownership.go":           {"modify", "harness-integrator", "SCOPE-620"},
	"internal/harness/buildcontrol/ownership_test.go":      {"modify", "harness-integrator", "SCOPE-620"},
	"internal/harness/buildcontrol/policy.go":              {"modify", "harness-integrator", "SCOPE-620"},
	"internal/harness/buildcontrol/policy_test.go":         {"modify", "harness-integrator", "SCOPE-620"},
}

var approvedConceptInputs = map[string]string{
	"docs/artifacts/concept-rebaseline-approval.md":        "cca3d19b5cf4af17e36841ee74fc58589b0386fecdae1a5b25747cd506aea161",
	"docs/artifacts/concept-rebaseline-change-contract.md": "691662df69d348312bed54045632e73c8260d951f2683101790a321c7544d936",
	"docs/artifacts/concept-rebaseline-design.md":          "83cf87d13834f11ee8b2d69200f0ed6a73c346abedb2dd7f65b8c1b334fe1296",
	"docs/artifacts/concept-rebaseline-specification.md":   "d3d957dce1571448c7a0f6bcc8d9c7afe2d952e6cec46b2c6de4259871a32871",
}

var conceptApprovalBindings = []string{
	"Artifact ID | `L7-APR-CRB-001`",
	"Status | **RECORDED AP0",
	"NO REPLAY**",
	"Accountable owner | Anup Pandey",
	"Source commit | `" + conceptBaseCommit + "`",
	"Source tree | `" + conceptBaseTree + "`",
	"Change contract SHA-256 | `691662df69d348312bed54045632e73c8260d951f2683101790a321c7544d936`",
	"Specification SHA-256 | `d3d957dce1571448c7a0f6bcc8d9c7afe2d952e6cec46b2c6de4259871a32871`",
	"Design SHA-256 | `83cf87d13834f11ee8b2d69200f0ed6a73c346abedb2dd7f65b8c1b334fe1296`",
	"Four cumulative hours and 25 deduplicated material public sources",
	"does not approve a not-yet-authored Concept Brief",
}

var (
	conceptArtifactIDPattern = regexp.MustCompile(`^L7-C(?:D|B)-[0-9]{3}$`)
	conceptVersionPattern    = regexp.MustCompile(`^(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)$`)
	materialSourceIDPattern  = regexp.MustCompile(`^SRC-[0-9]{3}$`)
	materialSourceReference  = regexp.MustCompile(`\bSRC-[0-9]{3}\b`)
	forbiddenBriefInventory  = regexp.MustCompile(`(?mi)^\s*(?:[-*]|[0-9]+\.)\s+(?:feature|functional requirement|nonfunctional requirement|architecture decision|technology choice|implementation step|api endpoint|schema|ui screen|cli command)\s*:`)
	forbiddenBriefIdentifier = regexp.MustCompile(`\b(?:L7-)?(?:FR|NFR)-[0-9]{2,}\b`)
)

var dossierStates = map[string]bool{
	"open": true, "researching": true, "ready_for_brief": true, "blocked": true, "superseded": true,
}

var briefStates = map[string]bool{
	"draft": true, "approved": true, "rejected": true, "stale": true, "superseded": true,
}

var requiredDossierSections = []string{
	"## Method and bounds",
	"## Query log",
	"## Material sources",
	"## Supporting evidence",
	"## Contrary evidence",
	"## Alternatives considered",
	"## Assumptions",
	"## Contradictions",
	"## Unresolved questions",
	"## Synthesis",
}

var permittedBriefSections = []string{
	"## Working identity",
	"## Target users and context",
	"## Evidenced problem",
	"## Desired outcome",
	"## Non-goals and boundaries",
	"## Success signals",
	"## Assumptions, uncertainty, and validation needs",
}

type dossierRecord struct {
	state           string
	productIdentity string
	digest          string
}

func checkConceptAdmission(root string, current map[string]snapshotFile) (string, []finding) {
	findings := checkConceptRebaselineApproval(root)
	for _, forbidden := range []string{wave02EvidencePath, wave02AuditPath} {
		if current[forbidden].regular {
			findings = appendFindings(findings, newFinding("SCOPE-630", forbidden, "stale Wave 2 scope gained an evidence or audit child", "remove the stale child and retain commit 34c3ba9 as unevidenced history"))
		}
	}

	dossierPresent := current[conceptDossierPath].regular
	briefPresent := current[conceptBriefPath].regular
	if !dossierPresent && !briefPresent {
		return "open", findings
	}
	if briefPresent && !dossierPresent {
		findings = appendFindings(findings, newFinding("CONCEPT-101", conceptBriefPath, "Concept Brief exists without its discovery dossier", "restore the source dossier or remove the premature brief"))
		return "invalid", findings
	}

	dossierData, readFindings := readStrictFile(root, conceptDossierPath)
	findings = appendFindings(findings, readFindings...)
	dossier, dossierFindings := validateConceptDossier(dossierData)
	findings = appendFindings(findings, dossierFindings...)
	if len(readFindings)+len(dossierFindings) != 0 {
		return "invalid", findings
	}
	if !briefPresent {
		return dossierCheckpoint(dossier.state), findings
	}

	briefData, briefReadFindings := readStrictFile(root, conceptBriefPath)
	findings = appendFindings(findings, briefReadFindings...)
	briefState, briefFindings := validateConceptBrief(briefData, dossierData, dossier)
	findings = appendFindings(findings, briefFindings...)
	if len(briefReadFindings)+len(briefFindings) != 0 {
		return "invalid", findings
	}
	return briefCheckpoint(briefState), findings
}

func checkConceptRebaselineApproval(root string) []finding {
	data, findings := readStrictFile(root, "docs/artifacts/concept-rebaseline-approval.md")
	if len(findings) != 0 {
		return findings
	}
	text := string(data)
	for _, binding := range conceptApprovalBindings {
		if !strings.Contains(text, binding) {
			findings = appendFindings(findings, newFinding("SCOPE-631", "docs/artifacts/concept-rebaseline-approval.md", "rebaseline approval lost an owner, source, scope, research, or artifact binding", "restore the exact non-replayable admission record"))
			break
		}
	}
	if strings.Contains(text, "RECORDED AP1") || strings.Contains(text, "authorizes replay") {
		findings = appendFindings(findings, newFinding("SCOPE-632", "docs/artifacts/concept-rebaseline-approval.md", "persisted admission claims replayable authority", "retain historical AP0 current-session semantics"))
	}
	return findings
}

func validateConceptDossier(data []byte) (dossierRecord, []finding) {
	document := string(data)
	fields, findings := parseConceptFields(conceptDossierPath, document)
	record := dossierRecord{state: fields["State"], productIdentity: fields["Product identity"], digest: fileSHA256(data)}

	if !conceptArtifactIDPattern.MatchString(fields["Artifact ID"]) || !strings.HasPrefix(fields["Artifact ID"], "L7-CD-") {
		findings = appendFindings(findings, newFinding("CONCEPT-110", "Artifact ID", "dossier has a missing or invalid artifact identity", "use a stable L7-CD-NNN identity"))
	}
	if !conceptVersionPattern.MatchString(fields["Version"]) {
		findings = appendFindings(findings, newFinding("CONCEPT-110", "Version", "dossier has a missing or invalid semantic version", "record an exact MAJOR.MINOR.PATCH version"))
	}
	if !dossierStates[record.state] {
		findings = appendFindings(findings, newFinding("CONCEPT-111", "State", "dossier state is not in the canonical state set", "use open, researching, ready_for_brief, blocked, or superseded"))
	}
	if record.productIdentity == "" || record.productIdentity == "UNSET" {
		findings = appendFindings(findings, newFinding("CONCEPT-112", "Product identity", "dossier does not bind a current product identity", "record the working product identity"))
	}

	minutes := boundedConceptInteger(fields["Cumulative minutes"], maxResearchMinutes, "Cumulative minutes", &findings)
	declaredSources := boundedConceptInteger(fields["Material source count"], maxMaterialSources, "Material source count", &findings)
	sourceIDs, sourceFindings := validateMaterialSources(document)
	findings = appendFindings(findings, sourceFindings...)
	if declaredSources >= 0 && declaredSources != len(sourceIDs) {
		findings = appendFindings(findings, newFinding("CONCEPT-116", "Material source count", fmt.Sprintf("dossier declares %d material sources but records %d unique source rows", declaredSources, len(sourceIDs)), "make the bounded count match the deduplicated source ledger"))
	}
	for _, reference := range materialSourceReference.FindAllString(document, -1) {
		if !sourceIDs[reference] {
			findings = appendFindings(findings, newFinding("CONCEPT-117", reference, "dossier cites a source ID absent from the material-source ledger", "add truthful provenance or remove the fabricated citation"))
		}
	}

	authority := fields["Network authority"]
	if authority != "authorized" && authority != "denied" {
		findings = appendFindings(findings, newFinding("CONCEPT-113", "Network authority", "dossier does not record authorized or denied public-network authority", "record the actual current-session authority"))
	}
	start, startOK := parseConceptTimestamp(fields["Research started"])
	end, endOK := parseConceptTimestamp(fields["Research ended"])

	switch record.state {
	case "open":
		if fields["Research started"] != "UNSET" || fields["Research ended"] != "UNSET" || minutes != 0 || declaredSources != 0 {
			findings = appendFindings(findings, newFinding("CONCEPT-120", "open", "open dossier claims research execution or collected sources", "move to researching or restore the unstarted state"))
		}
	case "researching":
		if authority != "authorized" || !startOK || fields["Research ended"] != "UNSET" {
			findings = appendFindings(findings, newFinding("CONCEPT-121", "researching", "researching state lacks authority/start or claims a final end", "record authorized resumable research state"))
		}
	case "ready_for_brief":
		if authority != "authorized" || !startOK || !endOK || (startOK && endOK && end.Before(start)) {
			findings = appendFindings(findings, newFinding("CONCEPT-122", "ready_for_brief", "ready dossier lacks an executed authorized time interval", "execute and record the bounded public research pass"))
		}
		for _, heading := range requiredDossierSections {
			if strings.Count(document, heading+"\n") != 1 {
				findings = appendFindings(findings, newFinding("CONCEPT-123", heading, "ready dossier is missing or duplicates a required structured section", "retain one complete structured research section"))
			}
		}
	case "blocked":
		if emptyConceptField(fields["Blocker"]) || emptyConceptField(fields["Next action"]) {
			findings = appendFindings(findings, newFinding("CONCEPT-124", "blocked", "blocked dossier lacks a concrete blocker or next action", "record the denied authority, missing capability, or other prerequisite"))
		}
	case "superseded":
		if emptyConceptField(fields["Disposition"]) {
			findings = appendFindings(findings, newFinding("CONCEPT-125", "superseded", "superseded dossier lacks its successor or abandonment disposition", "record why and by what the dossier was superseded"))
		}
	}
	return record, findings
}

func validateConceptBrief(data, dossierData []byte, dossier dossierRecord) (string, []finding) {
	document := string(data)
	fields, findings := parseConceptFields(conceptBriefPath, document)
	state := fields["State"]
	if !conceptArtifactIDPattern.MatchString(fields["Artifact ID"]) || !strings.HasPrefix(fields["Artifact ID"], "L7-CB-") {
		findings = appendFindings(findings, newFinding("CONCEPT-130", "Artifact ID", "brief has a missing or invalid artifact identity", "use a stable L7-CB-NNN identity"))
	}
	if !conceptVersionPattern.MatchString(fields["Version"]) {
		findings = appendFindings(findings, newFinding("CONCEPT-130", "Version", "brief has a missing or invalid semantic version", "record an exact MAJOR.MINOR.PATCH version"))
	}
	if !briefStates[state] {
		findings = appendFindings(findings, newFinding("CONCEPT-131", "State", "brief state is not in the canonical state set", "use draft, approved, rejected, stale, or superseded"))
	}
	if fields["Product identity"] != dossier.productIdentity {
		findings = appendFindings(findings, newFinding("CONCEPT-132", "Product identity", "brief product identity differs from its dossier", "bind both artifacts to the same current identity"))
	}
	if fields["Discovery path"] != conceptDossierPath || fields["Discovery SHA-256"] != fileSHA256(dossierData) {
		findings = appendFindings(findings, newFinding("CONCEPT-133", "Discovery source", "brief does not bind the exact canonical discovery dossier", "record the exact path and raw dossier digest"))
	}
	if (state == "draft" || state == "approved") && dossier.state != "ready_for_brief" {
		findings = appendFindings(findings, newFinding("CONCEPT-134", state, "brief was drafted or approved from a non-ready dossier", "complete the bounded research record first"))
	}

	headings := conceptLevelTwoHeadings(document)
	if len(headings) != len(permittedBriefSections) {
		findings = appendFindings(findings, newFinding("CONCEPT-135", conceptBriefPath, "brief does not contain exactly the seven permitted problem-contract sections", "remove solution sections and restore every permitted section"))
	} else {
		for index := range permittedBriefSections {
			if headings[index] != permittedBriefSections[index] {
				findings = appendFindings(findings, newFinding("CONCEPT-135", headings[index], "brief section is missing, reordered, or outside the permitted contract", "use only the seven canonical sections in order"))
				break
			}
		}
	}
	payloadStart := strings.Index(document, permittedBriefSections[0]+"\n")
	if payloadStart < 0 {
		findings = appendFindings(findings, newFinding("CONCEPT-136", "Payload SHA-256", "brief payload start is missing", "restore the Working identity section"))
	} else {
		payloadDigest := fileSHA256([]byte(document[payloadStart:]))
		if fields["Payload SHA-256"] != payloadDigest {
			findings = appendFindings(findings, newFinding("CONCEPT-136", "Payload SHA-256", "brief payload digest does not match the exact problem-contract bytes", "recompute the digest before requesting approval"))
		}
		if state == "approved" && fields["Approved payload SHA-256"] != payloadDigest {
			findings = appendFindings(findings, newFinding("CONCEPT-137", "Approved payload SHA-256", "approved brief is not bound to its current exact payload", "obtain a new owner decision for the current digest"))
		}
	}
	if forbiddenBriefInventory.MatchString(document) || forbiddenBriefIdentifier.MatchString(document) {
		findings = appendFindings(findings, newFinding("CONCEPT-138", conceptBriefPath, "brief contains a feature, requirement, architecture, technology, or implementation inventory", "move solution content to the gated downstream stage"))
	}

	switch state {
	case "draft":
		if fields["Approved payload SHA-256"] != "UNSET" || fields["Approval assurance"] != "none" || fields["Owner decision"] != "UNSET" {
			findings = appendFindings(findings, newFinding("CONCEPT-140", "draft", "draft brief claims an approval binding", "leave approval fields unset until the owner decides the exact digest"))
		}
	case "approved":
		if fields["Approval assurance"] != "AP0" || emptyConceptField(fields["Owner decision"]) || emptyConceptField(fields["Approved scope"]) {
			findings = appendFindings(findings, newFinding("CONCEPT-141", "approved", "approved brief lacks historical assurance, exact owner decision, or scope", "record the exact owner decision as non-replayable AP0"))
		}
	case "rejected", "stale", "superseded":
		if emptyConceptField(fields["Decision reason"]) {
			findings = appendFindings(findings, newFinding("CONCEPT-142", state, "terminal brief state lacks a decision or staleness reason", "record the reason and earliest restart stage"))
		}
	}
	return state, findings
}

func parseConceptFields(subject, document string) (map[string]string, []finding) {
	fields := make(map[string]string)
	var findings []finding
	for _, line := range strings.Split(document, "\n") {
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) != 2 || cells[0] == "Field" || strings.HasPrefix(cells[0], "---") {
			continue
		}
		key := conceptCellValue(cells[0])
		value := conceptCellValue(cells[1])
		if _, duplicate := fields[key]; duplicate {
			findings = appendFindings(findings, newFinding("CONCEPT-102", key, "concept artifact contains a duplicate metadata field", "retain one unambiguous field value"))
			continue
		}
		fields[key] = value
	}
	if len(fields) == 0 {
		findings = appendFindings(findings, newFinding("CONCEPT-103", subject, "concept artifact has no readable metadata table", "restore the structured artifact metadata"))
	}
	return fields, findings
}

func conceptCellValue(cell string) string {
	value := strings.TrimSpace(cell)
	if unquoted, ok := unquoteCodeCell(value); ok {
		return unquoted
	}
	for strings.HasPrefix(value, "**") && strings.HasSuffix(value, "**") && len(value) >= 4 {
		value = strings.TrimSpace(value[2 : len(value)-2])
	}
	return value
}

func boundedConceptInteger(value string, maximum int, subject string, findings *[]finding) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > maximum {
		*findings = appendFindings(*findings, newFinding("CONCEPT-114", subject, fmt.Sprintf("value is not an integer in 0..%d", maximum), "record the actual cumulative bounded value"))
		return -1
	}
	return parsed
}

func parseConceptTimestamp(value string) (time.Time, bool) {
	if value == "" || value == "UNSET" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil
}

func validateMaterialSources(document string) (map[string]bool, []finding) {
	sources := make(map[string]bool)
	var findings []finding
	for _, line := range strings.Split(document, "\n") {
		cells, ok := splitMarkdownRow(line)
		if !ok || len(cells) < 2 {
			continue
		}
		identifier := conceptCellValue(cells[0])
		if !materialSourceIDPattern.MatchString(identifier) {
			continue
		}
		if sources[identifier] {
			findings = appendFindings(findings, newFinding("CONCEPT-115", identifier, "material-source ledger contains a duplicate source ID", "deduplicate mirrors and retain one provenance record"))
			continue
		}
		sources[identifier] = true
		if len(cells) < 8 {
			findings = appendFindings(findings, newFinding("CONCEPT-115", identifier, "material source lacks the required provenance columns", "record title, publisher, URL, dates, source type, relevance, and provenance"))
		}
	}
	return sources, findings
}

func conceptLevelTwoHeadings(document string) []string {
	var headings []string
	for _, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			headings = append(headings, line)
		}
	}
	return headings
}

func emptyConceptField(value string) bool {
	return value == "" || value == "UNSET" || strings.EqualFold(value, "none")
}

func dossierCheckpoint(state string) string {
	switch state {
	case "ready_for_brief":
		return "ready-for-brief"
	default:
		return state
	}
}

func briefCheckpoint(state string) string {
	switch state {
	case "draft":
		return "brief-draft"
	case "approved":
		return "brief-approved"
	default:
		return "brief-" + state
	}
}

func validDossierTransition(from, to string) bool {
	if !dossierStates[from] || !dossierStates[to] {
		return false
	}
	if to == "superseded" && from != "superseded" {
		return true
	}
	allowed := map[string]map[string]bool{
		"open":        {"researching": true, "blocked": true},
		"researching": {"researching": true, "ready_for_brief": true, "blocked": true},
		"blocked":     {"researching": true},
	}
	return allowed[from][to]
}

func validBriefTransition(from, to string) bool {
	if !briefStates[from] || !briefStates[to] {
		return false
	}
	allowed := map[string]map[string]bool{
		"draft":    {"approved": true, "rejected": true, "stale": true, "superseded": true},
		"approved": {"stale": true, "superseded": true},
		"rejected": {"superseded": true},
		"stale":    {"superseded": true},
	}
	return allowed[from][to]
}
