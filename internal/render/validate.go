package render

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	semanticIDPattern  = regexp.MustCompile(`^L7-(TAX|OBL|GUARD|KNOW|WF|PROF|BUDGET|EVAL|CASE|TRUTH|EGR|COV|ADJ)-[A-Z0-9]+(-[A-Z0-9]+)*$`)
	requirementPattern = regexp.MustCompile(`^L7-[A-Z]+-[0-9]{3}$`)
	versionPattern     = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	datePattern        = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
)

var expectedTaxonomyValues = map[string][]string{
	"approval-assurance":  {"AP0", "AP1", "AP2", "AP3"},
	"capability":          {"available", "unavailable", "degraded", "unsupported", "not_applicable", "unverified"},
	"change-class":        {"documentation", "test", "infrastructure", "architecture", "feature", "security", "data", "dependency", "distribution", "operations"},
	"effect":              {"A0", "A1", "A2", "A3", "A4", "A5"},
	"evidence-state":      {"absent", "not_run", "not_evaluated", "unverified", "observed", "reproducible", "independently_verified", "invalidated", "stale", "superseded"},
	"gate-result":         {"pass", "fail", "blocked", "not_applicable"},
	"heritage":            {"prototype", "generated", "canonical", "deprecated", "retired"},
	"lifecycle":           {"baseline", "frame", "approve", "execute", "verify", "deliver", "observe", "learn", "package", "deploy", "expose"},
	"operational-state":   {"draft", "active", "blocked", "recovery_required", "superseded", "retired"},
	"product-decision":    {"proceed", "revise", "defer", "stop"},
	"reference-authority": {"law", "normative_standard", "official_guidance", "empirical_research", "practitioner_pattern"},
	"reference-status":    {"current", "draft", "emerging", "disputed", "superseded", "stale", "restricted"},
	"release-verdict":     {"go", "conditional_go", "no_go"},
	"risk":                {"R0", "R1", "R2", "R3", "R4"},
	"sensitivity":         {"public", "internal", "confidential", "restricted", "secret", "protected-evaluation"},
}

var requiredGuardrailIDs = []string{
	"L7-GUARD-A5-V1",
	"L7-GUARD-AP0-CURRENT",
	"L7-GUARD-EFFECT-CEILING",
	"L7-GUARD-EVALUATOR-CONTROL",
	"L7-GUARD-GO-BLOCKER",
	"L7-GUARD-NA-REASON",
	"L7-GUARD-NO-SUBAGENT",
	"L7-GUARD-OBLIGATION-ACCOUNTING",
	"L7-GUARD-PASS-UNVERIFIED",
	"L7-GUARD-RISK-FLOOR",
	"L7-GUARD-SECRET-NONLEAK",
	"L7-GUARD-UNKNOWN-SUCCESS",
}

var expectedLifecycleTransitions = []Transition{
	{From: "baseline", To: "frame", Gate: "approved lifecycle order"},
	{From: "frame", To: "approve", Gate: "framing evidence and owner decision"},
	{From: "approve", To: "execute", Gate: "current approval at the required assurance"},
	{From: "execute", To: "verify", Gate: "bounded implementation is complete"},
	{From: "verify", To: "deliver", Gate: "required verification is reproducible"},
	{From: "deliver", To: "observe", Gate: "delivery evidence exists"},
	{From: "observe", To: "learn", Gate: "outcome evidence exists"},
	{From: "learn", To: "baseline", Gate: "approved learning updates the next baseline"},
	{From: "deliver", To: "package", Gate: "explicit distribution approval"},
	{From: "package", To: "deploy", Gate: "explicit deployment approval"},
	{From: "deploy", To: "expose", Gate: "explicit exposure approval"},
	{From: "package", To: "observe", Gate: "reason-bearing package-only transition"},
	{From: "deploy", To: "observe", Gate: "reason-bearing deploy-without-exposure transition"},
}

var exactProfileIDs = []string{
	"L7-PROF-BEHAVIOR-PRESERVING-REFACTOR",
	"L7-PROF-FEATURE-CHANGE",
	"L7-PROF-GENERIC",
}

var promptSectionNames = []string{
	"goal_transition",
	"authoritative_inputs",
	"invariants_prohibited_effects",
	"authority_tools_capabilities_risk_effect",
	"acceptance_evidence",
	"budgets_stopping_escalation",
	"typed_output",
}

var promptSectionMarkers = []string{
	"{{L7:GOAL_TRANSITION}}",
	"{{L7:AUTHORITATIVE_INPUTS}}",
	"{{L7:INVARIANTS_PROHIBITED_EFFECTS}}",
	"{{L7:AUTHORITY_TOOLS_CAPABILITIES_RISK_EFFECT}}",
	"{{L7:ACCEPTANCE_EVIDENCE}}",
	"{{L7:BUDGETS_STOPPING_ESCALATION}}",
	"{{L7:TYPED_OUTPUT}}",
}

var descriptorAdditionalFields = map[string][]string{
	"budget":     {"measurement_scope", "tool_calls", "subagents", "retries_per_operation", "identical_failures", "wall_time_seconds", "tokens", "context_bytes", "context_items", "output_bytes", "monetary_micro_usd", "exhaustion", "recovery"},
	"delegation": {"objective", "disjoint_scope", "inputs", "authority", "effect_ceiling", "allowed_tools", "budget_id", "output_schema_id", "evidence", "verifier", "integration_owner", "termination", "single_agent_fallback"},
	"guardrail":  {"input", "decision", "failure_mode", "recovery", "proof", "criticality", "enforcement_locus", "grader_ids", "overrideability"},
	"knowledge":  {"pointer", "source_type", "authority_type", "source_version", "source_date", "source_status", "applicability", "contraindications", "jurisdiction", "license", "use_restriction", "freshness_days", "last_reviewed", "next_review", "normative", "mapping"},
	"obligation": {"source_requirement", "criticality", "rationale", "applicability", "rule", "enforcement_locus", "required_renderers", "machine_only", "grader_ids", "public_case_ids", "evidence_rule", "overrideability"},
	"output":     {"decision", "rule_ids", "scope", "source_identity", "evidence", "uncertainty", "assumptions", "defeaters", "residual_risk", "blocker", "decision_owner", "next_action", "effect", "authority", "diagnostics"},
	"profile":    {"description", "applicability", "contraindications", "obligation_ids", "risk_floor", "effect_ceiling", "approval_floor", "reference_ids", "budget_id", "composition"},
	"prompt-ir":  {"sections", "projection", "workflow_id", "profile_ids", "source_digests", "obligation_accounting", "output_schema_id"},
	"taxonomy":   {"kind", "values", "allowed_transitions", "invalid_combinations"},
	"workflow":   {"description", "positive_triggers", "negative_triggers", "prerequisites", "inputs", "lifecycle", "profiles", "obligation_ids", "risk_floor", "effect_ceiling", "approval_gate", "authority", "capabilities", "references", "budget", "output_schema", "success", "failure", "stopping", "recovery", "fixtures"},
}

var commonDescriptorFields = []string{
	"change_gate", "compatibility", "definition", "earliest_removal", "id", "introduced_by", "owner", "replacement", "retained_tests", "reviewer", "schema_version", "status", "supersedes", "version",
}

func Validate(bundle Bundle) []Diagnostic {
	var diagnostics []Diagnostic
	diagnostics = appendDiagnostics(diagnostics, validateRecordIdentities(bundle)...)
	diagnostics = appendDiagnostics(diagnostics, validateTaxonomies(bundle.Taxonomies)...)
	diagnostics = appendDiagnostics(diagnostics, validateObligations(bundle.Obligations)...)
	diagnostics = appendDiagnostics(diagnostics, validateGuardrails(bundle.Guardrails)...)
	diagnostics = appendDiagnostics(diagnostics, validateKnowledge(bundle.Knowledge)...)
	diagnostics = appendDiagnostics(diagnostics, validateWorkflows(bundle)...)
	diagnostics = appendDiagnostics(diagnostics, validateProfiles(bundle)...)
	diagnostics = appendDiagnostics(diagnostics, validateDescriptors(bundle)...)
	return finishDiagnostics(diagnostics)
}

func ValidateRequirementCoverage(bundle Bundle, requirements []Requirement) []Diagnostic {
	var diagnostics []Diagnostic
	if len(requirements) != 29 {
		diagnostics = addDiagnostic(diagnostics, "SEM-160", "requirements", fmt.Sprintf("source-derived Wave 2 ownership has %d requirements, want 29", len(requirements)), "restore approved BL-002/BL-003 ownership")
	}
	requirementByID := make(map[string]Requirement, len(requirements))
	for _, requirement := range requirements {
		if previous, duplicate := requirementByID[requirement.ID]; duplicate {
			diagnostics = addDiagnostic(diagnostics, "SEM-161", requirement.ID, "source-derived requirement is duplicated for owners "+previous.Owner+" and "+requirement.Owner, "retain one approved accountable owner")
		}
		requirementByID[requirement.ID] = requirement
	}
	obligationBySource := make(map[string]Obligation, len(bundle.Obligations))
	for _, obligation := range bundle.Obligations {
		if _, duplicate := obligationBySource[obligation.SourceRequirement]; duplicate {
			diagnostics = addDiagnostic(diagnostics, "SEM-162", obligation.SourceRequirement, "multiple obligations claim one source requirement", "retain one derived obligation")
		}
		obligationBySource[obligation.SourceRequirement] = obligation
	}
	for _, id := range sortedKeys(requirementByID) {
		requirement := requirementByID[id]
		obligation, ok := obligationBySource[id]
		if !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-163", id, "source-owned requirement has no obligation", "add the mechanically derived obligation")
			continue
		}
		if obligation.ID != obligationID(id) {
			diagnostics = addDiagnostic(diagnostics, "SEM-164", obligation.ID, "obligation ID is not mechanically derived from its source", "insert OBL after the L7 prefix")
		}
		if obligation.Rule != requirement.Rule {
			diagnostics = addDiagnostic(diagnostics, "SEM-165", id, "obligation rule differs from the authoritative requirement text", "restore the exact source-derived rule")
		}
	}
	for _, source := range sortedKeys(obligationBySource) {
		if _, ok := requirementByID[source]; !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-166", source, "obligation invents an unowned Wave 2 source", "remove the invented obligation or approve source ownership")
		}
	}
	return finishDiagnostics(diagnostics)
}

func DeriveWave02Requirements(requirementsData, backlogData []byte) ([]Requirement, []Diagnostic) {
	var diagnostics []Diagnostic
	if len(requirementsData) == 0 || len(requirementsData) > MaxAggregateBytes || len(backlogData) == 0 || len(backlogData) > MaxAggregateBytes {
		return nil, []Diagnostic{newDiagnostic("SEM-160", "requirements/backlog", "authoritative source is empty or over the bounded input size", "restore the approved bounded sources")}
	}
	definitions := make(map[string]string)
	inRequirements := false
	for lineNumber, line := range strings.Split(string(requirementsData), "\n") {
		if strings.HasPrefix(line, "## 9. Functional requirements") {
			inRequirements = true
			continue
		}
		if inRequirements && strings.HasPrefix(line, "## 11.") {
			break
		}
		if !inRequirements {
			continue
		}
		cells, ok := markdownCells(line)
		if !ok || len(cells) < 2 || !strings.HasPrefix(cells[0], "`L7-") {
			continue
		}
		id := strings.Trim(cells[0], "`")
		if !requirementPattern.MatchString(id) {
			diagnostics = addDiagnostic(diagnostics, "SEM-160", fmt.Sprintf("requirements.md:%d", lineNumber+1), "malformed normative requirement ID", "restore the exact source row")
			continue
		}
		if _, duplicate := definitions[id]; duplicate {
			diagnostics = addDiagnostic(diagnostics, "SEM-161", id, "duplicate normative requirement definition", "retain one authoritative definition")
			continue
		}
		ruleColumn := 1
		if strings.HasPrefix(id, "L7-NFR-") {
			if len(cells) < 3 {
				diagnostics = addDiagnostic(diagnostics, "SEM-160", fmt.Sprintf("requirements.md:%d", lineNumber+1), "nonfunctional requirement row lacks its normative statement", "restore the exact source row")
				continue
			}
			ruleColumn = 2
		}
		definitions[id] = cells[ruleColumn]
	}
	if !inRequirements {
		diagnostics = addDiagnostic(diagnostics, "SEM-160", "requirements.md", "normative requirement section is missing", "restore the approved source heading")
	}

	owned := make(map[string]string)
	inOwnership := false
	for lineNumber, line := range strings.Split(string(backlogData), "\n") {
		if strings.HasPrefix(line, "## 8. Normative requirement ownership and release allocation") {
			inOwnership = true
			continue
		}
		if inOwnership && strings.HasPrefix(line, "## 9.") {
			break
		}
		if !inOwnership {
			continue
		}
		cells, ok := markdownCells(line)
		if !ok || len(cells) != 4 {
			continue
		}
		owner := strings.Trim(cells[1], "`")
		if owner != "L7-BL-002" && owner != "L7-BL-003" {
			continue
		}
		ids, err := expandRequirementExpression(cells[0])
		if err != nil {
			diagnostics = addDiagnostic(diagnostics, "SEM-160", fmt.Sprintf("feature-backlog.md:%d", lineNumber+1), err.Error(), "restore the approved ownership expression")
			continue
		}
		for _, id := range ids {
			if previous, duplicate := owned[id]; duplicate {
				diagnostics = addDiagnostic(diagnostics, "SEM-161", id, "requirement has duplicate Wave 2 owners "+previous+" and "+owner, "retain one accountable owner")
				continue
			}
			owned[id] = owner
		}
	}
	if !inOwnership {
		diagnostics = addDiagnostic(diagnostics, "SEM-160", "feature-backlog.md", "ownership section is missing", "restore the approved source heading")
	}
	ids := make([]string, 0, len(owned))
	for id := range owned {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	requirements := make([]Requirement, 0, len(ids))
	for _, id := range ids {
		rule, ok := definitions[id]
		if !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-163", id, "owned Wave 2 requirement has no normative definition", "restore the approved requirement source")
			continue
		}
		requirements = append(requirements, Requirement{ID: id, Rule: rule, Owner: owned[id]})
	}
	if len(requirements) != 29 {
		diagnostics = addDiagnostic(diagnostics, "SEM-160", "requirements", fmt.Sprintf("derived BL-002/BL-003 total is %d, want 29", len(requirements)), "restore the approved source ownership")
	}
	if len(diagnostics) != 0 {
		return nil, finishDiagnostics(diagnostics)
	}
	return requirements, nil
}

func validateRecordIdentities(bundle Bundle) []Diagnostic {
	type record struct {
		meta RecordMeta
	}
	var records []record
	for _, value := range bundle.Taxonomies {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Obligations {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Guardrails {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Knowledge {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Workflows {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Profiles {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Budgets {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Delegations {
		records = append(records, record{value.RecordMeta})
	}
	for _, value := range bundle.Outputs {
		records = append(records, record{value.RecordMeta})
	}
	var diagnostics []Diagnostic
	byID := make(map[string]RecordMeta, len(records))
	for _, record := range records {
		meta := record.meta
		if !semanticIDPattern.MatchString(meta.ID) || len(meta.ID) > 64 {
			diagnostics = addDiagnostic(diagnostics, "SEM-120", meta.ID, "record ID violates the stable ASCII grammar or length", "use one approved namespace-scoped ID")
		}
		if !versionPattern.MatchString(meta.SchemaVersion) || !versionPattern.MatchString(meta.Version) {
			diagnostics = addDiagnostic(diagnostics, "SEM-121", meta.ID, "schema or record version is not MAJOR.MINOR.PATCH", "use a canonical semantic version")
		}
		if meta.Owner == "" || meta.Reviewer == "" || meta.ChangeGate == "" || meta.Definition == "" || meta.IntroducedBy == "" || meta.Compatibility == "" || meta.EarliestRemoval == "" {
			diagnostics = addDiagnostic(diagnostics, "SEM-122", meta.ID, "record identity, owner, definition, compatibility, or lifecycle field is empty", "complete the common record envelope")
		}
		if !oneOf(meta.Status, "draft", "active", "deprecated", "superseded", "retired") {
			diagnostics = addDiagnostic(diagnostics, "SEM-123", meta.ID, "record status is unknown", "use one exact lifecycle status")
		}
		if meta.Supersedes == nil || meta.Replacement == nil || len(meta.RetainedTests) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-124", meta.ID, "supersession, replacement, or retained-test state is missing", "record explicit arrays and retained coverage")
		}
		if (meta.Status == "deprecated" || meta.Status == "superseded") && len(meta.Replacement) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-128", meta.ID, "deprecated or superseded record has no replacement", "bind an explicit compatible replacement")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(meta.ID+":supersedes", meta.Supersedes)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(meta.ID+":replacement", meta.Replacement)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(meta.ID+":retained_tests", meta.RetainedTests)...)
		if previous, duplicate := byID[meta.ID]; duplicate {
			message := "duplicate stable record ID"
			if previous.Definition != meta.Definition || previous.Version != meta.Version {
				message = "stable ID is redefined with different meaning or version"
			}
			diagnostics = addDiagnostic(diagnostics, "SEM-125", meta.ID, message, "retain one compatible record identity")
		}
		byID[meta.ID] = meta
	}
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var visit func(string)
	visit = func(id string) {
		if visiting[id] {
			diagnostics = addDiagnostic(diagnostics, "SEM-126", id, "supersession graph contains a cycle", "restore an acyclic replacement history")
			return
		}
		if visited[id] {
			return
		}
		visited[id] = true
		visiting[id] = true
		for _, target := range byID[id].Supersedes {
			if _, ok := byID[target]; !ok {
				diagnostics = addDiagnostic(diagnostics, "SEM-127", id, "supersedes references an unknown record "+target, "restore a retained predecessor record")
				continue
			}
			visit(target)
		}
		for _, target := range byID[id].Replacement {
			if _, ok := byID[target]; !ok {
				diagnostics = addDiagnostic(diagnostics, "SEM-127", id, "replacement references an unknown record "+target, "restore a retained replacement record")
			}
		}
		visiting[id] = false
	}
	for _, id := range sortedKeys(byID) {
		visit(id)
	}
	return finishDiagnostics(diagnostics)
}

func validateTaxonomies(taxonomies []Taxonomy) []Diagnostic {
	if len(taxonomies) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	if len(taxonomies) != len(expectedTaxonomyValues) {
		diagnostics = addDiagnostic(diagnostics, "SEM-140", "taxonomy", fmt.Sprintf("registry has %d families, want %d", len(taxonomies), len(expectedTaxonomyValues)), "restore the exact initial taxonomy families")
	}
	seen := make(map[string]bool)
	for _, taxonomy := range taxonomies {
		if seen[taxonomy.Kind] {
			diagnostics = addDiagnostic(diagnostics, "SEM-141", taxonomy.Kind, "duplicate taxonomy family", "retain one canonical family")
		}
		seen[taxonomy.Kind] = true
		expected, ok := expectedTaxonomyValues[taxonomy.Kind]
		if !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-142", taxonomy.Kind, "unknown taxonomy family", "remove the invented family")
			continue
		}
		actual := make([]string, 0, len(taxonomy.Values))
		for _, value := range taxonomy.Values {
			actual = append(actual, value.Value)
			if value.Meaning == "" || value.Entry == "" || value.Exit == "" || value.Failure == "" || value.Blocked == "" || value.Stale == "" || value.Superseded == "" {
				diagnostics = addDiagnostic(diagnostics, "SEM-143", taxonomy.ID+":"+value.Value, "taxonomy value lacks lifecycle semantics", "complete entry, exit, failure, blocked, stale, and superseded meaning")
			}
		}
		if strings.Join(actual, "\x00") != strings.Join(expected, "\x00") {
			diagnostics = addDiagnostic(diagnostics, "SEM-144", taxonomy.Kind, "taxonomy values or their required order differ", "restore the exact initial values")
		}
		for _, transition := range taxonomy.AllowedTransitions {
			if transition.Gate == "" || !oneOf(transition.From, expected...) || !oneOf(transition.To, expected...) {
				diagnostics = addDiagnostic(diagnostics, "SEM-143", taxonomy.ID+":"+transition.From+"->"+transition.To, "transition is untyped or lacks a reason-bearing gate", "restore a typed gated lifecycle transition")
			}
		}
		if taxonomy.Kind == "lifecycle" && !equalTransitions(taxonomy.AllowedTransitions, expectedLifecycleTransitions) {
			diagnostics = addDiagnostic(diagnostics, "SEM-144", taxonomy.Kind+":transitions", "lifecycle transitions or their approved order differ", "restore the exact initial transition matrix")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(taxonomy.ID+":invalid_combinations", taxonomy.InvalidCombinations)...)
	}
	for _, kind := range sortedKeys(expectedTaxonomyValues) {
		if !seen[kind] {
			diagnostics = addDiagnostic(diagnostics, "SEM-145", kind, "required taxonomy family is missing", "restore the canonical family")
		}
	}
	return finishDiagnostics(diagnostics)
}

func equalTransitions(actual, expected []Transition) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range actual {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}

func validateObligations(obligations []Obligation) []Diagnostic {
	if len(obligations) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	if len(obligations) != 29 {
		diagnostics = addDiagnostic(diagnostics, "SEM-150", "obligations", fmt.Sprintf("ledger has %d obligations, want 29", len(obligations)), "derive one obligation per approved source requirement")
	}
	seenSources := make(map[string]bool)
	for _, obligation := range obligations {
		if !requirementPattern.MatchString(obligation.SourceRequirement) || obligation.ID != obligationID(obligation.SourceRequirement) {
			diagnostics = addDiagnostic(diagnostics, "SEM-151", obligation.ID, "obligation/source identity is malformed or not mechanically derived", "insert OBL after the L7 prefix")
		}
		if seenSources[obligation.SourceRequirement] {
			diagnostics = addDiagnostic(diagnostics, "SEM-152", obligation.SourceRequirement, "duplicate obligation source", "retain one source-derived obligation")
		}
		seenSources[obligation.SourceRequirement] = true
		if !oneOf(obligation.Criticality, "safety-critical", "material", "noncritical") || obligation.Rationale == "" || obligation.Rule == "" || obligation.EnforcementLocus == "" || obligation.EvidenceRule == "" || obligation.Overrideability == "" {
			diagnostics = addDiagnostic(diagnostics, "SEM-153", obligation.ID, "obligation semantics or criticality are incomplete", "complete the canonical obligation rule")
		}
		if obligation.Criticality != "noncritical" && len(obligation.GraderIDs) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-154", obligation.ID, "material obligation has no grader", "bind at least one deterministic grader")
		}
		if obligation.Criticality != "noncritical" && len(obligation.RequiredRenderers) == 0 && !obligation.MachineOnly {
			diagnostics = addDiagnostic(diagnostics, "SEM-155", obligation.ID, "material obligation has neither renderer nor machine-only disposition", "bind a renderer or explicit machine-only enforcement")
		}
		if obligation.MachineOnly && len(obligation.RequiredRenderers) != 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-156", obligation.ID, "machine-only obligation also claims a prose renderer", "choose one explicit accounting disposition")
		}
		if obligation.Criticality != "noncritical" && len(obligation.PublicCaseIDs) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-157", obligation.ID, "material obligation has no public case binding", "bind a public semantic case")
		}
		for _, renderer := range obligation.RequiredRenderers {
			if renderer != "controlled-client" && renderer != "stock-a0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-158", obligation.ID, "unknown renderer "+renderer, "use an approved projection")
			}
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(obligation.ID+":renderers", obligation.RequiredRenderers)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(obligation.ID+":graders", obligation.GraderIDs)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(obligation.ID+":cases", obligation.PublicCaseIDs)...)
		diagnostics = appendDiagnostics(diagnostics, validateApplicability(obligation.ID, obligation.Applicability)...)
	}
	return finishDiagnostics(diagnostics)
}

func validateGuardrails(guardrails []Guardrail) []Diagnostic {
	if len(guardrails) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	actual := make([]string, 0, len(guardrails))
	for _, guardrail := range guardrails {
		actual = append(actual, guardrail.ID)
		if guardrail.Input == "" || guardrail.Decision == "" || guardrail.FailureMode == "" || guardrail.Recovery == "" || guardrail.Proof == "" || guardrail.EnforcementLocus == "" || guardrail.Overrideability == "" || !oneOf(guardrail.Criticality, "safety-critical", "material", "noncritical") {
			diagnostics = addDiagnostic(diagnostics, "SEM-146", guardrail.ID, "guardrail decision, failure, recovery, proof, or authority is incomplete", "complete the guardrail contract")
		}
		if guardrail.Criticality != "noncritical" && len(guardrail.GraderIDs) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-147", guardrail.ID, "material guardrail has no grader", "bind a deterministic grader")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(guardrail.ID+":graders", guardrail.GraderIDs)...)
	}
	sort.Strings(actual)
	if strings.Join(actual, "\x00") != strings.Join(requiredGuardrailIDs, "\x00") {
		diagnostics = addDiagnostic(diagnostics, "SEM-148", "guardrails", "initial guardrail roster differs from the exact 12 rules", "restore every approved invalid-combination rule")
	}
	return finishDiagnostics(diagnostics)
}

func validateKnowledge(entries []Knowledge) []Diagnostic {
	if len(entries) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	authorities := make(map[string]bool)
	for _, entry := range entries {
		authorities[entry.AuthorityType] = true
		sourceDay, sourceOK := dateOrdinal(entry.SourceDate)
		lastDay, lastOK := dateOrdinal(entry.LastReviewed)
		nextDay, nextOK := dateOrdinal(entry.NextReview)
		if entry.Pointer == "" || entry.SourceType == "" || entry.SourceVersion == "" || !sourceOK || !lastOK || !nextOK || entry.Jurisdiction == "" || entry.License == "" || entry.UseRestriction == "" || entry.FreshnessDays <= 0 || entry.FreshnessDays > 3650 {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", entry.ID, "knowledge authority, date, license, freshness, or use metadata is incomplete", "complete the license-safe reference record")
		}
		if !oneOf(entry.AuthorityType, expectedTaxonomyValues["reference-authority"]...) || !oneOf(entry.SourceStatus, expectedTaxonomyValues["reference-status"]...) {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", entry.ID, "knowledge authority or status category is unknown", "use the canonical reference taxonomy")
		}
		if sourceOK && lastOK && sourceDay > lastDay {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", entry.ID, "source date follows its recorded review", "restore a truthful review chronology")
		}
		if lastOK && nextOK && (nextDay < lastDay || nextDay-lastDay > entry.FreshnessDays) {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", entry.ID, "next review is nonmonotonic or exceeds the freshness window", "restore the bounded review policy")
		}
		if entry.Normative && oneOf(entry.SourceStatus, "draft", "disputed", "superseded", "stale", "restricted") {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", entry.ID, "unsafe source status is silently normative", "mark it non-normative or restore current authority")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.ID+":applicability", entry.Applicability)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.ID+":contraindications", entry.Contraindications)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(entry.ID+":mapping", entry.Mapping)...)
	}
	for _, authority := range expectedTaxonomyValues["reference-authority"] {
		if !authorities[authority] {
			diagnostics = addDiagnostic(diagnostics, "SEM-149", authority, "knowledge registry does not distinguish this authority class", "add a labeled license-safe metadata record")
		}
	}
	return finishDiagnostics(diagnostics)
}

func validateWorkflows(bundle Bundle) []Diagnostic {
	if len(bundle.Workflows)+len(bundle.Budgets)+len(bundle.Delegations)+len(bundle.Outputs) == 0 && bundle.Template == "" {
		return nil
	}
	var diagnostics []Diagnostic
	if len(bundle.Workflows) != 1 || len(bundle.Budgets) != 1 || len(bundle.Delegations) != 1 || len(bundle.Outputs) != 1 || bundle.Template == "" {
		diagnostics = addDiagnostic(diagnostics, "SEM-170", "reference-workflow", "reference contract must contain one workflow, budget, delegation, output, and template", "restore the exact reference contract")
		return diagnostics
	}
	workflow := bundle.Workflows[0]
	if workflow.ID != "L7-WF-REFERENCE" || workflow.Status != "active" || len(workflow.Description) < 40 || len(workflow.Description) > 240 {
		diagnostics = addDiagnostic(diagnostics, "SEM-171", workflow.ID, "workflow identity or discovery description violates its budget", "use the concise reference workflow description")
	}
	clause := workflow.Description
	if index := strings.IndexAny(clause, ":."); index >= 0 {
		clause = clause[:index+1]
	}
	if len(clause) > 80 {
		diagnostics = addDiagnostic(diagnostics, "SEM-171", workflow.ID, "front-loaded capability clause exceeds 80 bytes", "shorten the leading capability clause")
	}
	if len(workflow.PositiveTriggers) == 0 || len(workflow.PositiveTriggers) > 16 || len(workflow.NegativeTriggers) == 0 || len(workflow.NegativeTriggers) > 16 {
		diagnostics = addDiagnostic(diagnostics, "SEM-172", workflow.ID, "positive or negative trigger count is outside 1..16", "restore the bounded trigger inventory")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":positive", workflow.PositiveTriggers)...)
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":negative", workflow.NegativeTriggers)...)
	positive := make(map[string]bool)
	for _, trigger := range workflow.PositiveTriggers {
		if len(trigger) == 0 || len(trigger) > 160 {
			diagnostics = addDiagnostic(diagnostics, "SEM-172", workflow.ID, "positive trigger exceeds its byte budget", "shorten the discovery trigger")
		}
		normalized := normalizeTrigger(trigger)
		if normalized == "" || positive[normalized] {
			diagnostics = addDiagnostic(diagnostics, "SEM-173", trigger, "positive trigger is empty or collides after normalization", "make normalized discovery triggers unique")
		}
		positive[normalized] = true
	}
	negative := make(map[string]bool)
	for _, trigger := range workflow.NegativeTriggers {
		if len(trigger) == 0 || len(trigger) > 160 {
			diagnostics = addDiagnostic(diagnostics, "SEM-172", workflow.ID, "negative trigger exceeds its byte budget", "shorten the discovery trigger")
		}
		normalized := normalizeTrigger(trigger)
		if normalized == "" || positive[normalized] || negative[normalized] {
			diagnostics = addDiagnostic(diagnostics, "SEM-173", trigger, "negative trigger collides after normalization or with a positive trigger", "make normalized discovery triggers unique and disjoint")
		}
		negative[normalized] = true
	}
	obligationIDs := make([]string, 0, len(bundle.Obligations))
	for _, obligation := range bundle.Obligations {
		obligationIDs = append(obligationIDs, obligation.ID)
	}
	sort.Strings(obligationIDs)
	if len(workflow.ObligationIDs) != 29 || !equalStrings(workflow.ObligationIDs, obligationIDs) {
		diagnostics = addDiagnostic(diagnostics, "SEM-174", workflow.ID, fmt.Sprintf("workflow accounts for %d obligations, want 29", len(workflow.ObligationIDs)), "restore the complete obligation set")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":obligations", workflow.ObligationIDs)...)
	profileSet := make(map[string]bool, len(bundle.Profiles))
	for _, profile := range bundle.Profiles {
		profileSet[profile.ID] = true
	}
	if !equalStrings(workflow.Profiles, exactProfileIDs) {
		diagnostics = addDiagnostic(diagnostics, "SEM-174", workflow.ID+":profiles", "workflow profile set differs from the exact initial roster", "restore the three approved profiles")
	}
	for _, id := range workflow.Profiles {
		if !profileSet[id] {
			diagnostics = addDiagnostic(diagnostics, "SEM-178", workflow.ID+":"+id, "workflow references an unknown profile", "restore an active approved profile")
		}
	}
	knowledgeSet := make(map[string]Knowledge, len(bundle.Knowledge))
	for _, knowledge := range bundle.Knowledge {
		knowledgeSet[knowledge.ID] = knowledge
	}
	for _, id := range workflow.References {
		knowledge, ok := knowledgeSet[id]
		if !ok || !selectableKnowledge(knowledge) {
			diagnostics = addDiagnostic(diagnostics, "SEM-178", workflow.ID+":"+id, "workflow reference is missing, stale, restricted, or unlicensed", "select current license-safe knowledge metadata")
		}
	}
	workflowSets := []struct {
		name   string
		values []string
	}{
		{name: "authority", values: workflow.Authority},
		{name: "capabilities", values: workflow.Capabilities},
		{name: "failure", values: workflow.Failure},
		{name: "fixtures", values: workflow.Fixtures},
		{name: "inputs", values: workflow.Inputs},
		{name: "prerequisites", values: workflow.Prerequisites},
		{name: "profiles", values: workflow.Profiles},
		{name: "recovery", values: workflow.Recovery},
		{name: "references", values: workflow.References},
		{name: "stopping", values: workflow.Stopping},
		{name: "success", values: workflow.Success},
	}
	for _, set := range workflowSets {
		name, values := set.name, set.values
		if len(values) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-182", workflow.ID+":"+name, "required workflow set is empty", "restore the bounded workflow contract")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":"+name, values)...)
	}
	if workflow.Lifecycle.Entry == "" || workflow.Lifecycle.Exit == "" || workflow.Lifecycle.Transition == "" || workflow.Lifecycle.AllowedRepeat == nil || workflow.Lifecycle.AllowedSkip == nil {
		diagnostics = addDiagnostic(diagnostics, "SEM-182", workflow.ID+":lifecycle", "workflow lifecycle rule is incomplete", "restore entry, exit, transition, repeat, and skip semantics")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":allowed_repeat", workflow.Lifecycle.AllowedRepeat)...)
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(workflow.ID+":allowed_skip", workflow.Lifecycle.AllowedSkip)...)
	if !oneOf(workflow.RiskFloor, "R0", "R1", "R2", "R3", "R4") || !oneOf(workflow.EffectCeiling, "A0", "A1", "A2", "A3", "A4") || !oneOf(workflow.ApprovalGate, "AP0", "AP1", "AP2", "AP3") || workflow.Budget == "" || workflow.OutputSchema == "" {
		diagnostics = addDiagnostic(diagnostics, "SEM-182", workflow.ID, "workflow risk, effect, approval, budget, or output contract is invalid", "restore the typed authority envelope")
	}
	budget := bundle.Budgets[0]
	if budget.ID != "L7-BUDGET-W02-DEV-001" || budget.Status != "active" || budget.MeasurementScope == "" || budget.ToolCalls != 64 || budget.Subagents != 4 || budget.RetriesPerOperation != 2 || budget.IdenticalFailures != 2 || budget.WallTimeSeconds != 900 || budget.Tokens != 200000 || budget.ContextBytes != 1048576 || budget.ContextItems != 256 || budget.OutputBytes != MaxOutputBytes || budget.MonetaryMicroUSD != 0 || budget.Exhaustion != "blocked" || budget.Recovery == "" || workflow.Budget != budget.ID {
		diagnostics = addDiagnostic(diagnostics, "SEM-175", budget.ID, "reference budget differs from the contextual approved ceilings", "restore the exact Wave 2 development budget")
	}
	delegation := bundle.Delegations[0]
	if delegation.Status != "active" || delegation.Objective == "" || delegation.Authority == "" || delegation.BudgetID != budget.ID || delegation.EffectCeiling != "A2" || delegation.OutputSchemaID != workflow.OutputSchema || delegation.Verifier == "" || delegation.IntegrationOwner != "wave-integrator" || delegation.SingleAgentFallback == "" {
		diagnostics = addDiagnostic(diagnostics, "SEM-176", delegation.ID, "delegation lacks bounded authority, one integrator, or single-agent fallback", "restore optional non-authoritative delegation")
	}
	delegationSets := []struct {
		name   string
		values []string
	}{
		{name: "allowed_tools", values: delegation.AllowedTools},
		{name: "evidence", values: delegation.Evidence},
		{name: "inputs", values: delegation.Inputs},
		{name: "scope", values: delegation.DisjointScope},
		{name: "termination", values: delegation.Termination},
	}
	for _, set := range delegationSets {
		name, values := set.name, set.values
		if len(values) == 0 {
			diagnostics = addDiagnostic(diagnostics, "SEM-176", delegation.ID+":"+name, "delegation set is empty", "restore the bounded validation-only manifest")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(delegation.ID+":"+name, values)...)
	}
	output := bundle.Outputs[0]
	if output.Status != "active" || len(output.Decision) == 0 || output.RuleIDs == "" || output.Scope == "" || output.SourceIdentity == "" || output.Evidence == "" || output.Uncertainty == "" || output.Assumptions == "" || output.Defeaters == "" || output.ResidualRisk == "" || output.Blocker == "" || output.Owner == "" || output.NextAction == "" || output.Effect == "" || output.Authority == "" || output.Diagnostics == "" {
		diagnostics = addDiagnostic(diagnostics, "SEM-181", output.ID, "typed output decision envelope is incomplete", "restore every required decision-first field")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique(output.ID+":decision", output.Decision)...)
	diagnostics = appendDiagnostics(diagnostics, validatePromptTemplate(bundle.Template)...)
	return finishDiagnostics(diagnostics)
}

func validateProfiles(bundle Bundle) []Diagnostic {
	if len(bundle.Profiles) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	actual := make([]string, 0, len(bundle.Profiles))
	obligations := make(map[string]bool, len(bundle.Obligations))
	for _, obligation := range bundle.Obligations {
		obligations[obligation.ID] = true
	}
	for _, profile := range bundle.Profiles {
		actual = append(actual, profile.ID)
		if profile.Status != "active" || profile.Description == "" || len(profile.Applicability) == 0 || profile.Contraindications == nil || len(profile.ObligationIDs) == 0 || !oneOf(profile.RiskFloor, "R0", "R1", "R2", "R3", "R4") || !oneOf(profile.EffectCeiling, "A0", "A1", "A2", "A3", "A4") || !oneOf(profile.ApprovalFloor, "AP0", "AP1", "AP2", "AP3") || profile.BudgetID != "L7-BUDGET-W02-DEV-001" || profile.Composition != "obligation-union-highest-risk-approval-minimum-effect" {
			diagnostics = addDiagnostic(diagnostics, "SEM-177", profile.ID, "profile floor, ceiling, or composition rule is invalid", "restore set-union/highest-floor composition")
		}
		for _, id := range profile.ObligationIDs {
			if !obligations[id] {
				diagnostics = addDiagnostic(diagnostics, "SEM-178", profile.ID+":"+id, "profile references an unknown obligation", "restore a source-owned obligation ID")
			}
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(profile.ID+":obligations", profile.ObligationIDs)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(profile.ID+":applicability", profile.Applicability)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(profile.ID+":contraindications", profile.Contraindications)...)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(profile.ID+":references", profile.ReferenceIDs)...)
		for _, id := range profile.ReferenceIDs {
			found := false
			for _, knowledge := range bundle.Knowledge {
				if knowledge.ID == id && selectableKnowledge(knowledge) {
					found = true
					break
				}
			}
			if !found {
				diagnostics = addDiagnostic(diagnostics, "SEM-178", profile.ID+":"+id, "profile reference is missing, stale, restricted, or unlicensed", "select current license-safe knowledge metadata")
			}
		}
	}
	sort.Strings(actual)
	if !equalStrings(actual, exactProfileIDs) {
		diagnostics = addDiagnostic(diagnostics, "SEM-179", "profiles", "initial profile roster differs from generic, feature-change, and behavior-preserving-refactor", "restore the exact launch profiles")
	}
	return finishDiagnostics(diagnostics)
}

func validatePromptTemplate(template string) []Diagnostic {
	var diagnostics []Diagnostic
	allowed := map[string]bool{
		"":                                     true,
		"---":                                  true,
		"# Level 7 semantic contract":          true,
		"## Goal transition":                   true,
		"## Authoritative inputs":              true,
		"## Invariants and prohibited effects": true,
		"## Authority, tools, capabilities, risk, and effect": true,
		"## Acceptance and evidence":                          true,
		"## Budgets, stopping, and escalation":                true,
		"## Typed output":                                     true,
	}
	previous := -1
	for _, marker := range promptSectionMarkers {
		if strings.Count(template, marker) != 1 {
			diagnostics = addDiagnostic(diagnostics, "SEM-180", marker, "prompt marker is missing or repeated", "restore each exact marker once")
			continue
		}
		position := strings.Index(template, marker)
		if position <= previous {
			diagnostics = addDiagnostic(diagnostics, "SEM-180", marker, "prompt markers are reordered", "restore the seven-section order")
		}
		previous = position
	}
	markerSet := make(map[string]bool, len(promptSectionMarkers))
	for _, marker := range promptSectionMarkers {
		markerSet[marker] = true
	}
	for lineNumber, line := range strings.Split(strings.TrimSuffix(template, "\n"), "\n") {
		if !allowed[line] && !markerSet[line] {
			diagnostics = addDiagnostic(diagnostics, "SEM-180", fmt.Sprintf("prompt.md.tmpl:%d", lineNumber+1), "template contains prose or unsupported host syntax", "retain only fixed headings, separators, blank lines, and exact markers")
		}
	}
	return finishDiagnostics(diagnostics)
}

func selectableKnowledge(knowledge Knowledge) bool {
	lastDay, lastOK := dateOrdinal(knowledge.LastReviewed)
	nextDay, nextOK := dateOrdinal(knowledge.NextReview)
	return knowledge.Status == "active" && knowledge.Pointer != "" && knowledge.License != "" && knowledge.UseRestriction != "" && knowledge.FreshnessDays > 0 && lastOK && nextOK && nextDay >= lastDay && nextDay-lastDay <= knowledge.FreshnessDays && !oneOf(knowledge.SourceStatus, "draft", "disputed", "superseded", "stale", "restricted")
}

func dateOrdinal(value string) (int64, bool) {
	if !datePattern.MatchString(value) {
		return 0, false
	}
	year, yearErr := strconv.Atoi(value[0:4])
	month, monthErr := strconv.Atoi(value[5:7])
	day, dayErr := strconv.Atoi(value[8:10])
	if yearErr != nil || monthErr != nil || dayErr != nil || year == 0 || month < 1 || month > 12 {
		return 0, false
	}
	daysInMonth := [...]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	leap := year%4 == 0 && (year%100 != 0 || year%400 == 0)
	if leap {
		daysInMonth[2] = 29
	}
	if day < 1 || day > daysInMonth[month] {
		return 0, false
	}
	previousYear := int64(year - 1)
	ordinal := previousYear*365 + previousYear/4 - previousYear/100 + previousYear/400
	for currentMonth := 1; currentMonth < month; currentMonth++ {
		ordinal += int64(daysInMonth[currentMonth])
	}
	return ordinal + int64(day), true
}

func equalStrings(actual, expected []string) bool {
	return strings.Join(actual, "\x00") == strings.Join(expected, "\x00")
}

func validateDescriptors(bundle Bundle) []Diagnostic {
	if len(bundle.Descriptors) == 0 {
		return nil
	}
	var diagnostics []Diagnostic
	requiredKinds := make(map[string]bool)
	if len(bundle.Taxonomies) != 0 {
		requiredKinds["taxonomy"] = true
	}
	if len(bundle.Obligations) != 0 {
		requiredKinds["obligation"] = true
	}
	if len(bundle.Guardrails) != 0 {
		requiredKinds["guardrail"] = true
	}
	if len(bundle.Knowledge) != 0 {
		requiredKinds["knowledge"] = true
	}
	if len(bundle.Workflows) != 0 {
		for _, kind := range []string{"workflow", "budget", "delegation", "output", "prompt-ir"} {
			requiredKinds[kind] = true
		}
	}
	if len(bundle.Profiles) != 0 {
		requiredKinds["profile"] = true
	}
	seen := make(map[string]bool)
	for _, descriptor := range bundle.Descriptors {
		kind := strings.TrimSuffix(strings.TrimPrefix(descriptor.ID, "schemas/semantic/"), ".schema.json")
		extras, ok := descriptorAdditionalFields[kind]
		if !ok || descriptor.ID != "schemas/semantic/"+kind+".schema.json" || descriptor.Schema != "https://json-schema.org/draft/2020-12/schema" || descriptor.Title == "" || descriptor.Type != "object" || descriptor.AdditionalProperties {
			diagnostics = addDiagnostic(diagnostics, "SEM-130", descriptor.ID, "schema descriptor identity or closed-object contract is invalid", "restore the fixed local descriptor")
			continue
		}
		if seen[kind] {
			diagnostics = addDiagnostic(diagnostics, "SEM-131", descriptor.ID, "duplicate schema descriptor", "retain one descriptor per semantic kind")
		}
		seen[kind] = true
		expected := append(append([]string(nil), commonDescriptorFields...), extras...)
		sort.Strings(expected)
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(descriptor.ID+":required", descriptor.Required)...)
		actual := append([]string(nil), descriptor.Required...)
		sort.Strings(actual)
		if strings.Join(actual, "\x00") != strings.Join(expected, "\x00") || len(descriptor.Properties) != len(expected) {
			diagnostics = addDiagnostic(diagnostics, "SEM-132", descriptor.ID, "required fields or property table differ from the authoritative Go contract", "restore descriptor parity")
		}
		for _, property := range expected {
			value, ok := descriptor.Properties[property]
			if !ok {
				diagnostics = addDiagnostic(diagnostics, "SEM-132", descriptor.ID+":"+property, "required property descriptor is missing", "restore descriptor parity")
				continue
			}
			if value.Type != expectedSchemaPropertyType(kind, property) {
				diagnostics = addDiagnostic(diagnostics, "SEM-134", descriptor.ID+":"+property, "property type differs from the authoritative Go contract", "restore the exact typed property")
			}
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(descriptor.ID+":"+property, value)...)
		}
	}
	for _, kind := range sortedKeys(requiredKinds) {
		if !seen[kind] {
			diagnostics = addDiagnostic(diagnostics, "SEM-133", kind, "applicable schema descriptor is missing", "add the exact local descriptor")
		}
	}
	return finishDiagnostics(diagnostics)
}

func expectedSchemaPropertyType(kind, name string) string {
	if oneOf(name, "replacement", "retained_tests", "supersedes") {
		return "array"
	}
	arrays := map[string][]string{
		"delegation": {"allowed_tools", "disjoint_scope", "evidence", "inputs", "termination"},
		"guardrail":  {"grader_ids"},
		"knowledge":  {"applicability", "contraindications", "mapping"},
		"obligation": {"grader_ids", "public_case_ids", "required_renderers"},
		"output":     {"decision"},
		"profile":    {"applicability", "contraindications", "obligation_ids", "reference_ids"},
		"prompt-ir":  {"obligation_accounting", "profile_ids", "sections", "source_digests"},
		"taxonomy":   {"allowed_transitions", "invalid_combinations", "values"},
		"workflow":   {"authority", "capabilities", "failure", "fixtures", "inputs", "negative_triggers", "obligation_ids", "positive_triggers", "prerequisites", "profiles", "recovery", "references", "stopping", "success"},
	}
	if oneOf(name, arrays[kind]...) {
		return "array"
	}
	if kind == "obligation" && name == "applicability" || kind == "workflow" && name == "lifecycle" {
		return "object"
	}
	if kind == "obligation" && name == "machine_only" || kind == "knowledge" && name == "normative" {
		return "boolean"
	}
	if kind == "knowledge" && name == "freshness_days" || kind == "budget" && oneOf(name, "context_bytes", "context_items", "identical_failures", "monetary_micro_usd", "output_bytes", "retries_per_operation", "subagents", "tokens", "tool_calls", "wall_time_seconds") {
		return "integer"
	}
	return "string"
}

func validateSchemaProperty(subject string, property SchemaProperty) []Diagnostic {
	var diagnostics []Diagnostic
	if property.Description == "" {
		diagnostics = addDiagnostic(diagnostics, "SEM-134", subject, "property description is empty", "restore the license-safe local interface description")
	}
	switch property.Type {
	case "array":
		if property.Items == nil || property.MinItems == nil || property.MaxItems == nil || *property.MinItems < 0 || *property.MaxItems < *property.MinItems {
			diagnostics = addDiagnostic(diagnostics, "SEM-134", subject, "array property lacks bounded item semantics", "restore item, minimum, and maximum bounds")
		}
		if property.Items != nil {
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(subject+":items", *property.Items)...)
		}
	case "object":
		if property.AdditionalProperties == nil || *property.AdditionalProperties || property.Required == nil || len(property.Properties) != len(property.Required) {
			diagnostics = addDiagnostic(diagnostics, "SEM-134", subject, "critical object is open or differs from its required property table", "restore the closed nested object contract")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(subject+":required", property.Required)...)
		for _, name := range property.Required {
			nested, ok := property.Properties[name]
			if !ok {
				diagnostics = addDiagnostic(diagnostics, "SEM-134", subject+":"+name, "nested required property is missing", "restore the closed nested object contract")
				continue
			}
			diagnostics = appendDiagnostics(diagnostics, validateSchemaProperty(subject+":"+name, nested)...)
		}
	case "integer":
		if property.Minimum == nil || property.Maximum == nil || *property.Minimum < 0 || *property.Maximum < *property.Minimum {
			diagnostics = addDiagnostic(diagnostics, "SEM-134", subject, "integer property lacks nonnegative fixed bounds", "restore exact integer bounds")
		}
	case "boolean", "string":
	default:
		diagnostics = addDiagnostic(diagnostics, "SEM-134", subject, "property uses an unsupported type", "use the authoritative zero-dependency type")
	}
	return finishDiagnostics(diagnostics)
}

func validateApplicability(subject string, applicability Applicability) []Diagnostic {
	var diagnostics []Diagnostic
	dimensions := []struct {
		name   string
		values []string
	}{
		{name: "capabilities", values: applicability.Capabilities},
		{name: "contraindications", values: applicability.Contraindications},
		{name: "effects", values: applicability.Effects},
		{name: "hosts", values: applicability.Hosts},
		{name: "profiles", values: applicability.Profiles},
		{name: "risks", values: applicability.Risks},
		{name: "stages", values: applicability.Stages},
	}
	for _, dimension := range dimensions {
		name, values := dimension.name, dimension.values
		if values == nil {
			diagnostics = addDiagnostic(diagnostics, "SEM-159", subject+":"+name, "applicability dimension is omitted", "record an explicit possibly-empty set")
		}
		diagnostics = appendDiagnostics(diagnostics, sortedUnique(subject+":"+name, values)...)
	}
	return diagnostics
}

func sortedUnique(subject string, values []string) []Diagnostic {
	var diagnostics []Diagnostic
	previous := ""
	for index, value := range values {
		if value == "" || (index > 0 && value <= previous) {
			diagnostics = addDiagnostic(diagnostics, "SEM-129", subject, "set-valued array is empty, duplicated, or not bytewise sorted", "sort unique stable values")
			break
		}
		previous = value
	}
	return diagnostics
}

func sortedKeys[Value any](values map[string]Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func obligationID(requirement string) string {
	return strings.Replace(requirement, "L7-", "L7-OBL-", 1)
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func markdownCells(line string) ([]string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
		return nil, false
	}
	parts := strings.Split(strings.Trim(trimmed, "|"), "|")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts, true
}

func expandRequirementExpression(expression string) ([]string, error) {
	parts := strings.Split(expression, ", ")
	var ids []string
	prefix := ""
	for _, raw := range parts {
		part := strings.ReplaceAll(raw, "`", "")
		bounds := strings.Split(part, "–")
		if len(bounds) > 2 {
			return nil, fmt.Errorf("malformed requirement range %q", raw)
		}
		normalize := func(value string) (string, error) {
			if requirementPattern.MatchString(value) {
				prefix = value[:len(value)-3]
				return value, nil
			}
			if len(value) == 3 && prefix != "" {
				if _, err := strconv.Atoi(value); err == nil {
					return prefix + value, nil
				}
			}
			return "", fmt.Errorf("malformed requirement token %q", value)
		}
		start, err := normalize(bounds[0])
		if err != nil {
			return nil, err
		}
		if len(bounds) == 1 {
			ids = append(ids, start)
			continue
		}
		end, err := normalize(bounds[1])
		if err != nil {
			return nil, err
		}
		if start[:len(start)-3] != end[:len(end)-3] {
			return nil, fmt.Errorf("cross-family range %q", raw)
		}
		first, _ := strconv.Atoi(start[len(start)-3:])
		last, _ := strconv.Atoi(end[len(end)-3:])
		if last < first || last-first > 512 {
			return nil, fmt.Errorf("reversed or oversized requirement range %q", raw)
		}
		for value := first; value <= last; value++ {
			ids = append(ids, prefix+fmt.Sprintf("%03d", value))
		}
	}
	return ids, nil
}

func normalizeTrigger(value string) string {
	stop := map[string]bool{"a": true, "an": true, "and": true, "for": true, "in": true, "of": true, "on": true, "or": true, "the": true, "to": true, "with": true}
	var cleaned strings.Builder
	for _, character := range strings.ToLower(value) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == ' ' {
			cleaned.WriteRune(character)
		}
	}
	var words []string
	for _, word := range strings.Fields(cleaned.String()) {
		if !stop[word] {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}
