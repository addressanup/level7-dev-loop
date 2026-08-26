package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	foundationBaseCommit                = "1c5c351f52f258d37ba48d8348e1cd883d2fb250"
	foundationBaseTree                  = "b1fe4753b51b0da847d73b0ff64377fb2bda1434"
	foundationBaseManifestSHA256        = "176f3725b6c801b34a3a3208efc8c6609b854fe483f8bc2c6cd167170261aa47"
	foundationPathPolicySHA256          = "09513eab93c254c50a5cae2704786a62a9d3a61f02103c93d28706f8c49f6ecc"
	foundationCandidateSHA256           = "97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40"
	foundationApprovalSHA256            = "b025d1e4c97b2dd208e2047ee7bfe0ddce759e673be0c6822313019399f64369"
	foundationHistorySHA256             = "4a208c68d7ff559c52db53b5907f2f4bcec8226be459330f9e9e2bf6d406ba2b"
	foundationGateRegistrySHA256        = "13a455175a28dd76f7c6b0313e4868515bcbfe0c86fa536b3966d05ecb460ecb"
	foundationPredecessorManifestSHA256 = "3b8be699cb39824b0c56498bd2b71dad08035f8668c17eef6b6c4e2a17cdbd4d"
	foundationAdmissionEvidencePath     = "docs/artifacts/foundation-rebaseline-admission-evidence.md"
	foundationAdmissionAuditPath        = "docs/artifacts/foundation-rebaseline-admission-audit.md"
	foundationCandidatePath             = "docs/artifacts/foundation-rebaseline-candidate.sha256"
	foundationPathPolicyPath            = "harness/foundation-rebaseline-paths.tsv"
	foundationGateRegistryPath          = "harness/foundation-rebaseline-gates.tsv"
	foundationPredecessorManifestPath   = "harness/foundation-rebaseline-predecessors.sha256"
	expectedFoundationBaseFiles         = 174
	expectedFoundationPathRows          = 69
	expectedFoundationPredecessorRows   = 29
	expectedFoundationGateRows          = 8
)

type foundationPathExpectation struct {
	pathExpectation
	windows map[string]bool
	window  string
}

var foundationWindows = []string{
	"admission", "requirements", "backlog", "architecture", "technology", "harness", "orchestration", "audit",
}

var foundationOwners = map[string]bool{
	"architecture-owner": true, "backlog-owner": true, "experience-owner": true,
	"foundation-integrator": true, "harness-integrator": true, "independent-auditor": true,
	"orchestration-owner": true, "qualification-owner": true, "requirements-owner": true,
	"technology-owner": true, "traceability-owner": true,
}

var foundationApprovedBundle = map[string]string{
	"docs/artifacts/foundation-rebaseline-change-contract.md": "8c297db289bd9f405ccdec9f33448fb81def5f6334ba70a23c0306cbd3aa68e8",
	"docs/artifacts/foundation-rebaseline-design.md":          "7621541e0319dc5a2c238ea57a3841b3bfb62a0bb3b46d8371bb3bb3ba54b23a",
	"docs/artifacts/foundation-rebaseline-specification.md":   "ff9ca1d03a21533e000bb586796ea4367f95df59a321900b5044ce71d6dfebc9",
	foundationPathPolicyPath:                                  foundationPathPolicySHA256,
}

var foundationStaticDigests = map[string]string{
	foundationCandidatePath:                            foundationCandidateSHA256,
	"docs/artifacts/foundation-rebaseline-approval.md": foundationApprovalSHA256,
	"docs/artifacts/foundation-rebaseline-history.md":  foundationHistorySHA256,
	foundationGateRegistryPath:                         foundationGateRegistrySHA256,
	foundationPredecessorManifestPath:                  foundationPredecessorManifestSHA256,
}

var foundationRequiredAdmissionAdds = []string{
	"docs/artifacts/foundation-rebaseline-approval.md",
	foundationCandidatePath,
	"docs/artifacts/foundation-rebaseline-change-contract.md",
	"docs/artifacts/foundation-rebaseline-design.md",
	"docs/artifacts/foundation-rebaseline-history.md",
	"docs/artifacts/foundation-rebaseline-specification.md",
	"harness/foundation-rebaseline-base.sha256",
	foundationGateRegistryPath,
	foundationPathPolicyPath,
	foundationPredecessorManifestPath,
	"internal/harness/buildcontrol/foundation.go",
	"internal/harness/buildcontrol/foundation_test.go",
}

var foundationProhibitedPathPrefixes = []string{
	".claude-plugin/", ".codex-plugin/", "build/", "cmd/", "fixtures/", "internal/adapter/",
	"internal/conductor/", "internal/context/", "internal/distribution/", "internal/evaluator/",
	"internal/executor/", "internal/platform/", "internal/policy/", "internal/receipt/",
	"internal/render/", "internal/state/", "internal/supervisor/", "internal/transaction/",
	"packages/", "protected/", "references/", "schemas/", "semantic/", "skills/",
}

var commitPattern = regexp.MustCompile(`\b[0-9a-f]{40}\b`)

func loadFoundationPathPolicy(root string) (map[string]foundationPathExpectation, []tsvRow, []finding) {
	data, findings := readStrictFile(root, foundationPathPolicyPath)
	if len(findings) != 0 {
		return nil, nil, findings
	}
	if fileSHA256(data) != foundationPathPolicySHA256 {
		findings = appendFindings(findings, newFinding("FRB-SCOPE-101", foundationPathPolicyPath, "path policy differs from the exact owner-approved Gate 2 payload", "restore the approved path-policy bytes or obtain a fresh exact approval"))
	}
	rows, parseFindings := parseTSV(foundationPathPolicyPath, data, []string{"change", "path", "owner", "window", "rule"})
	findings = appendFindings(findings, parseFindings...)
	expectations, validationFindings := validateFoundationPathRows(rows)
	findings = appendFindings(findings, validationFindings...)
	return expectations, rows, findings
}

func validateFoundationPathRows(rows []tsvRow) (map[string]foundationPathExpectation, []finding) {
	expectations := make(map[string]foundationPathExpectation)
	var findings []finding
	if len(rows) != expectedFoundationPathRows {
		findings = appendFindings(findings, newFinding("FRB-SCOPE-102", foundationPathPolicyPath, fmt.Sprintf("path policy has %d rows, want exactly %d", len(rows), expectedFoundationPathRows), "restore the exact approved finite path envelope"))
	}
	previous := ""
	changeCounts := map[string]int{}
	for _, row := range rows {
		relative := row["path"]
		if !safeRelativeASCIIPath(relative) {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-103", relative, "foundation path is not canonical ASCII", "restore an exact approved repository-relative path"))
			continue
		}
		if previous != "" && relative <= previous {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-104", relative, "foundation path policy is not in strict bytewise order", "sort the exact policy rows by path"))
		}
		previous = relative
		if _, duplicate := expectations[relative]; duplicate {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-105", relative, "duplicate foundation path-policy row", "retain exactly one approved path row"))
			continue
		}
		change := row["change"]
		if change != "add" && change != "modify" {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-106", relative, "foundation path uses an unapproved change class", "use add or modify exactly as approved"))
		}
		changeCounts[change]++
		if !foundationOwners[row["owner"]] {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-107", relative, "foundation path has an unknown accountable owner", "restore an approved stage owner"))
		}
		windows, windowFindings := parseFoundationWindows(relative, row["window"])
		findings = appendFindings(findings, windowFindings...)
		if !regexp.MustCompile(`^FRB-SCOPE-00[1-7]$`).MatchString(row["rule"]) {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-108", relative, "foundation path has an unknown scope rule", "restore an approved FRB-SCOPE-001 through FRB-SCOPE-007 rule"))
		}
		for _, prefix := range foundationProhibitedPathPrefixes {
			if strings.HasPrefix(relative, prefix) {
				findings = appendFindings(findings, newFinding("FRB-SCOPE-109", relative, "foundation policy enters a protected or product/runtime prefix", "remove it and obtain a new exact approval for any changed envelope"))
			}
		}
		expectations[relative] = foundationPathExpectation{
			pathExpectation: pathExpectation{change: change, owner: row["owner"], rule: row["rule"]},
			windows:         windows,
			window:          row["window"],
		}
	}
	if changeCounts["add"] != 56 || changeCounts["modify"] != 13 {
		findings = appendFindings(findings, newFinding("FRB-SCOPE-110", foundationPathPolicyPath, fmt.Sprintf("change classes are add=%d modify=%d, want add=56 modify=13", changeCounts["add"], changeCounts["modify"]), "restore the exact approved change classes"))
	}
	return expectations, findings
}

func parseFoundationWindows(relative, value string) (map[string]bool, []finding) {
	allowed := make(map[string]int, len(foundationWindows))
	for index, window := range foundationWindows {
		allowed[window] = index
	}
	parsed := make(map[string]bool)
	previous := -1
	var findings []finding
	for _, window := range strings.Split(value, ",") {
		index, ok := allowed[window]
		if !ok || parsed[window] || index <= previous {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-111", relative, "foundation stage windows are unknown, duplicated, or out of lifecycle order", "restore the exact approved ordered window list"))
			continue
		}
		parsed[window] = true
		previous = index
	}
	return parsed, findings
}

func genericFoundationRules(expectations map[string]foundationPathExpectation) map[string]pathExpectation {
	rules := make(map[string]pathExpectation, len(expectations))
	for relative, expectation := range expectations {
		rules[relative] = expectation.pathExpectation
	}
	return rules
}

func validateFoundationWindow(base map[string]string, current map[string]snapshotFile, expectations map[string]foundationPathExpectation, activeWindow string) []finding {
	var findings []finding
	for relative, expectation := range expectations {
		baseDigest, inBase := base[relative]
		file, inCurrent := current[relative]
		changed := (inBase && (!inCurrent || file.digest != baseDigest)) || (!inBase && inCurrent)
		if changed && !expectation.windows[activeWindow] {
			findings = appendFindings(findings, newFinding("FRB-SCOPE-112", relative, "path changed outside its approved foundation stage window", "restore the exact preimage or return to the earliest authorized stage"))
		}
	}
	return findings
}

func checkFoundationAdmission(root string, base map[string]string, current map[string]snapshotFile, expectations map[string]foundationPathExpectation) (string, []finding) {
	var findings []finding
	findings = appendFindings(findings, validateFoundationWindow(base, current, expectations, "admission")...)
	for _, relative := range foundationRequiredAdmissionAdds {
		if !current[relative].regular {
			findings = appendFindings(findings, newFinding("FRB-ADM-101", relative, "required admission record or controller file is absent", "complete the exact approved admission payload"))
		}
	}
	for relative, expected := range foundationApprovedBundle {
		findings = appendFindings(findings, checkFoundationDigest(root, relative, expected, "FRB-ADM-102")...)
	}
	for relative, expected := range foundationStaticDigests {
		findings = appendFindings(findings, checkFoundationDigest(root, relative, expected, "FRB-ADM-103")...)
	}
	findings = appendFindings(findings, checkFoundationCandidateManifest(root)...)
	findings = appendFindings(findings, checkFoundationBaseAndPredecessors(root, base, current)...)
	findings = appendFindings(findings, checkFoundationGateRegistry(root)...)
	findings = appendFindings(findings, checkFoundationApproval(root)...)
	findings = appendFindings(findings, checkFoundationHistory(root)...)

	conceptCheckpoint, conceptFindings := checkConceptAdmission(root, current)
	findings = appendFindings(findings, conceptFindings...)
	if conceptCheckpoint != "brief-approved" {
		findings = appendFindings(findings, newFinding("FRB-ADM-104", conceptBriefPath, "foundation admission is not based on the exact approved Concept Brief checkpoint", "restore the approved Concept Discovery and Concept Brief preimage"))
	}

	evidencePresent := current[foundationAdmissionEvidencePath].regular
	auditPresent := current[foundationAdmissionAuditPath].regular
	if !evidencePresent {
		if auditPresent {
			findings = appendFindings(findings, newFinding("FRB-ADM-105", foundationAdmissionAuditPath, "admission audit exists before local admission evidence", "remove the premature audit through authorized recovery"))
		}
		return "admission-in-progress", findings
	}
	findings = appendFindings(findings, checkFoundationAdmissionEvidence(root)...)
	if !auditPresent {
		return "admitted-awaiting-assurance", findings
	}
	findings = appendFindings(findings, checkFoundationAdmissionAudit(root)...)
	return "admitted", findings
}

func checkFoundationDigest(root, relative, expected, rule string) []finding {
	data, findings := readStrictFile(root, relative)
	if len(findings) == 0 && fileSHA256(data) != expected {
		findings = appendFindings(findings, newFinding(rule, relative, "record differs from its frozen admission digest", "restore the exact approved or admitted bytes"))
	}
	return findings
}

func checkFoundationCandidateManifest(root string) []finding {
	data, findings := readStrictFile(root, foundationCandidatePath)
	if len(findings) != 0 {
		return findings
	}
	manifest, parseFindings := parseSHA256Manifest(foundationCandidatePath, data, true)
	findings = appendFindings(findings, parseFindings...)
	if len(manifest) != len(foundationApprovedBundle) {
		findings = appendFindings(findings, newFinding("FRB-ADM-106", foundationCandidatePath, fmt.Sprintf("Gate 2 bundle has %d rows, want %d", len(manifest), len(foundationApprovedBundle)), "restore the exact owner-approved candidate manifest"))
	}
	for relative, expected := range foundationApprovedBundle {
		if manifest[relative] != expected {
			findings = appendFindings(findings, newFinding("FRB-ADM-107", relative, "Gate 2 candidate manifest lost an exact approved payload binding", "restore the approved manifest row"))
		}
	}
	return findings
}

func checkFoundationBaseAndPredecessors(root string, base map[string]string, current map[string]snapshotFile) []finding {
	var findings []finding
	if len(base) != expectedFoundationBaseFiles {
		findings = appendFindings(findings, newFinding("FRB-HIST-101", "harness/foundation-rebaseline-base.sha256", fmt.Sprintf("base inventory has %d rows, want %d", len(base), expectedFoundationBaseFiles), "restore the complete exact predecessor inventory"))
	}
	data, readFindings := readStrictFile(root, foundationPredecessorManifestPath)
	findings = appendFindings(findings, readFindings...)
	if len(readFindings) == 0 {
		manifest, parseFindings := parseSHA256Manifest(foundationPredecessorManifestPath, data, true)
		findings = appendFindings(findings, parseFindings...)
		if len(manifest) != expectedFoundationPredecessorRows {
			findings = appendFindings(findings, newFinding("FRB-HIST-102", foundationPredecessorManifestPath, fmt.Sprintf("key predecessor manifest has %d rows, want %d", len(manifest), expectedFoundationPredecessorRows), "restore the exact key-predecessor closure"))
		}
		for relative, digest := range manifest {
			if base[relative] != digest {
				findings = appendFindings(findings, newFinding("FRB-HIST-103", relative, "key predecessor digest differs from the complete base inventory", "restore the exact predecessor binding"))
			}
		}
	}
	for _, forbidden := range []string{wave02EvidencePath, wave02AuditPath} {
		if current[forbidden].regular {
			findings = appendFindings(findings, newFinding("FRB-HIST-104", forbidden, "stale Wave 2 gained fabricated completion evidence or audit", "remove the unauthorized child and retain Wave 2 as an unevidenced candidate"))
		}
	}
	return findings
}

func checkFoundationGateRegistry(root string) []finding {
	rows, findings := loadTSV(root, foundationGateRegistryPath, []string{"order", "stage", "window", "owner", "candidate_001", "candidate_002", "approval_001", "approval_002", "next"})
	if len(rows) != expectedFoundationGateRows {
		findings = appendFindings(findings, newFinding("FRB-GATE-101", foundationGateRegistryPath, fmt.Sprintf("gate registry has %d rows, want %d", len(rows), expectedFoundationGateRows), "restore the exact eight-stage serial lifecycle"))
	}
	for index, row := range rows {
		if row["order"] != strconv.Itoa(index+1) {
			findings = appendFindings(findings, newFinding("FRB-GATE-102", foundationGateRegistryPath, "gate registry order is not the exact serial lifecycle", "restore contiguous order 1 through 8"))
			break
		}
	}
	return findings
}

func checkFoundationApproval(root string) []finding {
	data, findings := readStrictFile(root, "docs/artifacts/foundation-rebaseline-approval.md")
	if len(findings) != 0 {
		return findings
	}
	return appendFindings(findings, validateFoundationApproval(data)...)
}

func validateFoundationApproval(data []byte) []finding {
	var findings []finding
	if fileSHA256(data) != foundationApprovalSHA256 {
		findings = appendFindings(findings, newFinding("FRB-APR-101", "docs/artifacts/foundation-rebaseline-approval.md", "approval receipt differs from the exact current-conversation record", "restore the frozen AP0 receipt"))
	}
	text := string(data)
	required := []string{
		"Artifact ID | `L7-APR-FRB-001`", "RECORDED AP0", "NO REPLAY", "Accountable owner | Anup Pandey",
		"APPROVE L7-FRB-CAND-001 " + foundationCandidateSHA256 + " FOR FOUNDATION-REBASELINE-ADMISSION-AND-REQUIREMENTS-CANDIDATE-ONLY",
		"Candidate-manifest SHA-256 | `" + foundationCandidateSHA256 + "`", "Source commit | `" + foundationBaseCommit + "`",
		"Source tree | `" + foundationBaseTree + "`", "Network effect | None during admission",
		"only after exact admission evidence and genuinely separate read-only admission `GO`", "not reusable authority",
	}
	for _, fragment := range required {
		if !strings.Contains(text, fragment) {
			findings = appendFindings(findings, newFinding("FRB-APR-102", "docs/artifacts/foundation-rebaseline-approval.md", "approval lost an exact owner, digest, source, scope, or non-replay binding", "restore the exact current-conversation AP0 receipt"))
			break
		}
	}
	if strings.Contains(text, "RECORDED AP1") || strings.Contains(text, "self-audit satisfies") || strings.Contains(text, "replayable authority") {
		findings = appendFindings(findings, newFinding("FRB-APR-103", "docs/artifacts/foundation-rebaseline-approval.md", "persisted approval claims replay or self-assurance authority", "retain historical AP0 semantics and separate assurance"))
	}
	return findings
}

func checkFoundationHistory(root string) []finding {
	data, findings := readStrictFile(root, "docs/artifacts/foundation-rebaseline-history.md")
	if len(findings) != 0 {
		return findings
	}
	text := string(data)
	for _, fragment := range []string{
		"`historical_stale` from admission", "`candidate_without_completion_evidence_or_audit`",
		"Wave 2 is not complete", "All product capabilities remain `PLANNED` or `UNVERIFIED`",
		"Predecessor commit | `" + foundationBaseCommit + "`", "Predecessor tree | `" + foundationBaseTree + "`",
	} {
		if !strings.Contains(text, fragment) {
			findings = appendFindings(findings, newFinding("FRB-HIST-105", "docs/artifacts/foundation-rebaseline-history.md", "successor history lost a staleness, Wave 2, source, or claim-state binding", "restore the admitted history record"))
			break
		}
	}
	return findings
}

func checkFoundationAdmissionEvidence(root string) []finding {
	data, findings := readStrictFile(root, foundationAdmissionEvidencePath)
	if len(findings) != 0 {
		return findings
	}
	text := string(data)
	required := []string{
		"Artifact ID | `L7-FRB-ADM-EVD-001`", "State | `admitted-awaiting-assurance`",
		"Gate 2 candidate | `L7-FRB-CAND-001`", "Candidate-manifest SHA-256 | `" + foundationCandidateSHA256 + "`",
		"Source commit | `" + foundationBaseCommit + "`", "Source tree | `" + foundationBaseTree + "`",
		"Local verification | `PASS`", "Independent admission assurance | `NOT_RUN`",
		"does not satisfy independent assurance", "Wave 2 evidence/audit | `ABSENT`",
	}
	for _, fragment := range required {
		if !strings.Contains(text, fragment) {
			findings = appendFindings(findings, newFinding("FRB-EVD-101", foundationAdmissionEvidencePath, "admission evidence lacks an exact candidate, source, result, limitation, or Wave 2 binding", "restore truthful bounded local evidence"))
			break
		}
	}
	if !commitPattern.MatchString(text) {
		findings = appendFindings(findings, newFinding("FRB-EVD-102", foundationAdmissionEvidencePath, "admission evidence lacks a committed candidate identity", "record the exact local admission commit"))
	}
	return findings
}

func checkFoundationAdmissionAudit(root string) []finding {
	data, findings := readStrictFile(root, foundationAdmissionAuditPath)
	if len(findings) != 0 {
		return findings
	}
	text := string(data)
	required := []string{
		"Artifact ID | `L7-FRB-ADM-AUD-001`", "Decision | `GO`", "Role | `independent-auditor`",
		"Access | `read-only candidate`", "Candidate-manifest SHA-256 | `" + foundationCandidateSHA256 + "`",
		"Candidate writer excluded | `true`",
	}
	for _, fragment := range required {
		if !strings.Contains(text, fragment) {
			findings = appendFindings(findings, newFinding("FRB-AUD-101", foundationAdmissionAuditPath, "admission audit lacks a separate-role, read-only, candidate, or GO binding", "obtain a genuinely separate bound read-only review"))
			break
		}
	}
	if strings.Contains(text, "Reviewer | `foundation-integrator`") || strings.Contains(text, "self-audit") {
		findings = appendFindings(findings, newFinding("FRB-AUD-102", foundationAdmissionAuditPath, "candidate writer is represented as independent assurance", "use genuinely separate review authority"))
	}
	return findings
}

func foundationOptionalRegular(root, relative string) bool {
	info, err := os.Lstat(filepath.Join(root, relative))
	return err == nil && info.Mode().IsRegular()
}
