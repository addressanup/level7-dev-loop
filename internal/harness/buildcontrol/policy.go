package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type controllerOptions struct {
	Root         string
	BaseRef      string
	HeadRef      string
	ChangeID     string
	Tier         riskTier
	TierOneScope []string
	RequireReady bool
}

type controllerReport struct {
	Tier      riskTier
	ChangeID  string
	Base      string
	Head      string
	Tree      string
	State     workflowState
	Next      string
	PathCount int
}

func (report controllerReport) String() string {
	return fmt.Sprintf("PASS rule=BCTL-000 version=%s tier=%s change=%s base=%s head=%s tree=%s state=%s changed=%d next=%q",
		buildControlVersion, report.Tier, report.ChangeID, report.Base, report.Head, report.Tree, report.State, report.PathCount, report.Next)
}

var protectedExact = map[string]bool{
	".claude-plugin/plugin.json": true,
	".codex-plugin/plugin.json":  true,
	"AGENTS.md":                  true,
	"Makefile":                   true,
	"marketplace.json":           true,
	"plugin.json":                true,
	"references/WORKFLOW.md":     true,
}

var protectedPrefixes = []string{
	".github/workflows/",
	"harness/",
	"internal/harness/buildcontrol/",
	"scripts/harness/",
	"skills/",
}

func runController(options controllerOptions) (controllerReport, []finding) {
	repository := gitRepository{root: options.Root}
	head, err := repository.resolve(options.HeadRef)
	if err != nil {
		return controllerReport{}, []finding{newFinding("GIT-001", options.HeadRef, err.Error(), "provide a valid candidate Git commit")}
	}
	tree, err := repository.tree(head)
	if err != nil {
		return controllerReport{}, []finding{newFinding("GIT-002", head, err.Error(), "restore the candidate Git tree")}
	}

	brief, briefCommit, briefFindings := discoverBrief(repository, head, options.ChangeID)
	findings := appendFindings(nil, briefFindings...)
	tier := options.Tier
	baseRef := options.BaseRef
	scope := append([]string(nil), options.TierOneScope...)
	changeID := options.ChangeID
	if brief.Path != "" {
		tier, baseRef, scope, changeID = brief.Tier, brief.BaseCommit, brief.Scope, brief.ID
		if options.Tier != 0 && options.Tier != brief.Tier {
			findings = appendFindings(findings, newFinding("RISK-001", changeID, "explicit risk tier conflicts with the change brief", "use the brief tier or elevate it"))
		}
		if options.BaseRef != "" {
			requestedBase, resolveErr := repository.resolve(options.BaseRef)
			if resolveErr != nil || requestedBase != brief.BaseCommit {
				findings = appendFindings(findings, newFinding("GIT-003", options.BaseRef, "base revision conflicts with the change brief", "evaluate the exact declared base commit"))
			}
		}
	}
	if tier < tierRoutine || tier > tierHighRisk {
		findings = appendFindings(findings, newFinding("RISK-002", changeID, "risk tier is not declared", "declare Tier 1 explicitly or add one Tier 2/3 brief"))
	}
	if tier == tierRoutine && brief.Path != "" {
		findings = appendFindings(findings, newFinding("ART-001", brief.Path, "Tier 1 must not create a governance artifact", "remove the artifact or elevate the change"))
	}
	if tier > tierRoutine && brief.Path == "" {
		findings = appendFindings(findings, newFinding("ART-002", changeID, "Tier 2/3 requires exactly one change brief", "add docs/artifacts/changes/<change-id>.md"))
	}
	if baseRef == "" {
		findings = appendFindings(findings, newFinding("GIT-004", changeID, "base Git revision is missing", "declare the exact base commit"))
		return controllerReport{}, findings
	}
	base, err := repository.resolve(baseRef)
	if err != nil {
		findings = appendFindings(findings, newFinding("GIT-004", baseRef, err.Error(), "declare a valid base Git commit"))
		return controllerReport{}, findings
	}
	if !repository.isAncestor(base, head) {
		findings = appendFindings(findings, newFinding("GIT-005", base, "declared base is not an ancestor of the candidate", "use the actual change base"))
	}
	changes, err := repository.changed(base, head)
	if err != nil {
		findings = appendFindings(findings, newFinding("GIT-006", changeID, err.Error(), "restore a readable Git change set"))
		return controllerReport{}, findings
	}
	if tier == tierRoutine && len(scope) == 0 {
		findings = appendFindings(findings, newFinding("SCOPE-001", changeID, "Tier 1 scope is missing", "pass a concise comma-separated scope"))
	}
	findings = appendFindings(findings, validateScopeAndRisk(tier, scope, changes)...)
	findings = appendFindings(findings, validateArtifactBudget(tier, brief, changes)...)

	state, stateFindings := evaluateState(repository, head, tree, brief, briefCommit, changes)
	findings = appendFindings(findings, stateFindings...)
	_, next, ok := nextState(tier, state)
	if !ok {
		findings = appendFindings(findings, newFinding("STATE-001", string(state), "controller produced an invalid state", "restore an executable workflow state"))
	}
	if options.RequireReady && state != stateReady {
		findings = appendFindings(findings, newFinding("STATE-002", string(state), "change is not ready to merge", next))
	}
	if len(findings) != 0 {
		return controllerReport{}, findings
	}
	return controllerReport{Tier: tier, ChangeID: changeID, Base: base, Head: head, Tree: tree, State: state, Next: next, PathCount: len(changes)}, nil
}

func discoverBrief(repository gitRepository, head, requestedID string) (changeBrief, string, []finding) {
	paths, err := repository.list(head, "docs/artifacts/changes")
	if err != nil {
		return changeBrief{}, "", []finding{newFinding("BRIEF-010", "docs/artifacts/changes", err.Error(), "restore readable change briefs")}
	}
	type candidate struct {
		brief  changeBrief
		commit string
	}
	var candidates []candidate
	var findings []finding
	for _, relative := range paths {
		if !strings.HasSuffix(relative, ".md") || strings.HasSuffix(relative, "-verification.md") || strings.HasSuffix(relative, "-audit.md") {
			continue
		}
		data, showErr := repository.show(head, relative)
		if showErr != nil {
			findings = appendFindings(findings, newFinding("BRIEF-011", relative, showErr.Error(), "restore the brief Git blob"))
			continue
		}
		brief, parseFindings := parseBrief(relative, data)
		if requestedID != "" && brief.ID != requestedID {
			continue
		}
		commit, commitErr := repository.additionCommit(head, relative)
		if commitErr != nil {
			findings = appendFindings(findings, newFinding("BRIEF-012", relative, commitErr.Error(), "retain the brief in Git history"))
			continue
		}
		if repository.isAncestor(brief.BaseCommit, head) && repository.isAncestor(commit, head) {
			findings = appendFindings(findings, parseFindings...)
			candidates = append(candidates, candidate{brief, commit})
		}
	}
	if len(candidates) == 0 {
		return changeBrief{}, "", findings
	}
	sort.Slice(candidates, func(i, j int) bool {
		return repository.isAncestor(candidates[i].commit, candidates[j].commit) && candidates[i].commit != candidates[j].commit
	})
	selected := candidates[len(candidates)-1]
	return selected.brief, selected.commit, findings
}

func validateScopeAndRisk(tier riskTier, scope []string, changes []changedPath) []finding {
	var findings []finding
	for _, change := range changes {
		if !scopeContains(scope, change.Path) {
			findings = appendFindings(findings, newFinding("SCOPE-002", change.Path, "changed path is outside declared scope", "remove the path or explicitly revise the change scope"))
		}
		if isProtected(change.Path) && tier != tierHighRisk {
			findings = appendFindings(findings, newFinding("RISK-003", change.Path, "protected control change requires Tier 3", "elevate the change and obtain owner approval plus independent audit"))
		}
	}
	return findings
}

func isProtected(relative string) bool {
	if protectedExact[relative] {
		return true
	}
	for _, prefix := range protectedPrefixes {
		if strings.HasPrefix(relative, prefix) {
			return true
		}
	}
	return false
}

func validateArtifactBudget(tier riskTier, brief changeBrief, changes []changedPath) []finding {
	var artifacts []string
	for _, change := range changes {
		if strings.HasPrefix(change.Path, "docs/artifacts/") {
			artifacts = append(artifacts, change.Path)
		}
	}
	allowed := map[string]bool{}
	if brief.Path != "" {
		allowed[brief.Path] = true
		allowed["docs/artifacts/changes/"+brief.ID+"-verification.md"] = true
		allowed["docs/artifacts/changes/"+brief.ID+"-audit.md"] = true
	}
	var findings []finding
	for _, relative := range artifacts {
		if !allowed[relative] {
			findings = appendFindings(findings, newFinding("ART-003", relative, "governance artifact is outside the tier budget", "keep evidence in Git, CI, tests, or the permitted records"))
		}
	}
	limit := 0
	if tier == tierProduct {
		limit = 1
	} else if tier == tierHighRisk {
		limit = 3
	}
	if len(artifacts) > limit {
		findings = appendFindings(findings, newFinding("ART-004", brief.ID, fmt.Sprintf("change has %d governance artifacts; tier limit is %d", len(artifacts), limit), "remove redundant governance files"))
	}
	return findings
}

func evaluateState(repository gitRepository, head, tree string, brief changeBrief, briefCommit string, changes []changedPath) (workflowState, []finding) {
	implementation := 0
	verificationPath := ""
	auditPath := ""
	if brief.ID != "" {
		verificationPath = "docs/artifacts/changes/" + brief.ID + "-verification.md"
		auditPath = "docs/artifacts/changes/" + brief.ID + "-audit.md"
	}
	present := make(map[string]bool)
	for _, change := range changes {
		present[change.Path] = change.Status != "D"
		if change.Path != brief.Path && change.Path != verificationPath && change.Path != auditPath {
			implementation++
		}
	}
	if implementation == 0 {
		if brief.Tier == tierHighRisk {
			return stateAwaitingOwnerApproval, nil
		}
		return statePlanned, nil
	}
	if brief.Tier != tierHighRisk {
		if os.Getenv("L7_REVIEW_REF") == head && os.Getenv("L7_VERIFIED_REF") == head {
			return stateReady, nil
		}
		if os.Getenv("L7_VERIFIED_REF") == head {
			return stateVerified, nil
		}
		return stateBuilding, nil
	}

	approval, approvalFindings := loadApproval(repository, brief, briefCommit)
	if len(approvalFindings) != 0 {
		return stateAwaitingOwnerApproval, approvalFindings
	}
	if approval.Actor == approval.Implementer {
		return stateAwaitingOwnerApproval, []finding{newFinding("AUTH-003", approval.Actor, "accountable-owner approval is self-issued by the implementer", "obtain an explicit decision from a distinct accountable owner")}
	}
	if !present[verificationPath] {
		return stateBuilding, nil
	}
	_, verificationCommit, verificationFindings := validateVerification(repository, head, verificationPath, auditPath, brief.ID)
	if len(verificationFindings) != 0 {
		return stateBuilding, verificationFindings
	}
	if !present[auditPath] {
		return stateAwaitingIndependentAudit, nil
	}
	auditFindings := validateAudit(repository, head, tree, auditPath, brief, approval, verificationCommit)
	if len(auditFindings) != 0 {
		return stateAwaitingIndependentAudit, auditFindings
	}
	return stateReady, nil
}

func loadApproval(repository gitRepository, brief changeBrief, briefCommit string) (approvalEnvelope, []finding) {
	commonDir, err := repository.commonDir()
	if err != nil {
		return approvalEnvelope{}, []finding{newFinding("AUTH-001", brief.ID, err.Error(), "restore access to external approval context")}
	}
	name := authorityPath(commonDir, "approvals", brief.ID)
	data, findings := readBounded(name)
	if len(findings) != 0 {
		return approvalEnvelope{}, []finding{newFinding("AUTH-001", brief.ID, "external accountable-owner approval is absent", "record explicit approval outside candidate-controlled repository text")}
	}
	var approval approvalEnvelope
	if decodeFindings := decodeStrictJSON(name, data, &approval); len(decodeFindings) != 0 {
		return approvalEnvelope{}, decodeFindings
	}
	if approval.Schema != 1 || approval.ChangeID != brief.ID || approval.BriefCommit != briefCommit || approval.Actor == "" || approval.Implementer == "" || (approval.Source != "active-user-interaction" && approval.Source != "trusted-ci") {
		return approvalEnvelope{}, []finding{newFinding("AUTH-002", brief.ID, "approval identity, source, or brief binding is invalid", "obtain fresh external approval for the exact brief commit")}
	}
	return approval, nil
}

func validateVerification(repository gitRepository, head, verificationPath, auditPath, changeID string) (evidenceRecord, string, []finding) {
	data, err := repository.show(head, verificationPath)
	if err != nil {
		return evidenceRecord{}, "", []finding{newFinding("VERIFY-001", verificationPath, err.Error(), "restore the verification record")}
	}
	record := parseEvidence(data)
	candidate, err := repository.resolve(record.CandidateCommit)
	if err != nil || record.ChangeID != changeID || record.Result != "PASS" {
		return record, "", []finding{newFinding("VERIFY-002", verificationPath, "verification record has an invalid change, candidate, or result", "bind PASS evidence to the exact implementation commit")}
	}
	candidateTree, treeErr := repository.tree(candidate)
	if treeErr != nil || candidateTree != record.CandidateTree {
		return record, "", []finding{newFinding("VERIFY-003", verificationPath, "verification tree does not match Git", "use the Git-derived candidate tree")}
	}
	verificationCommit, err := repository.additionCommit(head, verificationPath)
	if err != nil || !repository.isAncestor(candidate, verificationCommit) {
		return record, "", []finding{newFinding("VERIFY-004", verificationPath, "verification record is not a successor of its candidate", "record verification after testing the candidate")}
	}
	postChanges, err := repository.changed(candidate, head)
	if err != nil {
		return record, "", []finding{newFinding("VERIFY-005", verificationPath, err.Error(), "restore a readable verification lineage")}
	}
	for _, change := range postChanges {
		if change.Path != verificationPath && change.Path != auditPath {
			return record, "", []finding{newFinding("VERIFY-006", change.Path, "implementation changed after verification", "rerun verification against the new candidate")}
		}
	}
	return record, verificationCommit, nil
}

func validateAudit(repository gitRepository, head, headTree, auditPath string, brief changeBrief, approval approvalEnvelope, verificationCommit string) []finding {
	data, err := repository.show(head, auditPath)
	if err != nil {
		return []finding{newFinding("AUDIT-001", auditPath, err.Error(), "restore the independent audit record")}
	}
	record := parseEvidence(data)
	if record.ChangeID != brief.ID || record.CandidateCommit != verificationCommit || record.Result != "GO" {
		return []finding{newFinding("AUDIT-002", auditPath, "audit is not a GO decision bound to the verified candidate", "obtain a fresh independent read-only audit")}
	}
	verifiedTree, treeErr := repository.tree(verificationCommit)
	if treeErr != nil || record.CandidateTree != verifiedTree {
		return []finding{newFinding("AUDIT-003", auditPath, "audit tree does not match the verified Git candidate", "bind the audit to the exact Git tree")}
	}
	commonDir, commonErr := repository.commonDir()
	if commonErr != nil {
		return []finding{newFinding("AUDIT-004", brief.ID, commonErr.Error(), "restore external audit authority")}
	}
	name := authorityPath(commonDir, "audits", brief.ID)
	envelopeData, envelopeFindings := readBounded(name)
	if len(envelopeFindings) != 0 {
		return []finding{newFinding("AUDIT-004", brief.ID, "external independent-audit authority is absent", "provide a trusted audit envelope outside repository prose")}
	}
	var envelope auditEnvelope
	if decodeFindings := decodeStrictJSON(name, envelopeData, &envelope); len(decodeFindings) != 0 {
		return decodeFindings
	}
	if envelope.Schema != 1 || envelope.ChangeID != brief.ID || envelope.CandidateCommit != verificationCommit || envelope.AuditCommit != head || envelope.Actor == "" || envelope.Actor == approval.Actor || envelope.Actor == approval.Implementer || envelope.Actor != record.Reviewer || (envelope.Source != "independent-agent" && envelope.Source != "trusted-ci") {
		return []finding{newFinding("AUDIT-005", brief.ID, "audit identity, independence, source, or Git binding is invalid", "obtain a separate audit bound to the verified candidate and audit commit")}
	}
	if tree, treeErr := repository.tree(head); treeErr != nil || tree != headTree {
		return []finding{newFinding("AUDIT-006", head, "audit successor tree changed during evaluation", "retry from a stable Git candidate")}
	}
	return nil
}
