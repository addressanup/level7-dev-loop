package render

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type ProjectionKind string

const (
	ProjectionStockA0          ProjectionKind = "stock-a0"
	ProjectionControlledClient ProjectionKind = "controlled-client"
)

var inputIDPattern = regexp.MustCompile(`^L7-INPUT-[0-9]{3}$`)

type AuthoritativeInput struct {
	ID              string `json:"id"`
	Source          Digest `json:"source"`
	Version         string `json:"version"`
	Provenance      string `json:"provenance"`
	Trust           string `json:"trust"`
	Sensitivity     string `json:"sensitivity"`
	Freshness       string `json:"freshness"`
	InclusionReason string `json:"inclusion_reason"`
}

type PromptSection struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ObligationAccounting struct {
	ObligationID      string   `json:"obligation_id"`
	SourceRequirement string   `json:"source_requirement"`
	Disposition       string   `json:"disposition"`
	RuleSHA256        string   `json:"rule_sha256"`
	GraderIDs         []string `json:"grader_ids"`
	PublicCaseIDs     []string `json:"public_case_ids"`
}

type PromptIR struct {
	RecordMeta
	Sections             []PromptSection        `json:"sections"`
	Projection           ProjectionKind         `json:"projection"`
	WorkflowID           string                 `json:"workflow_id"`
	ProfileIDs           []string               `json:"profile_ids"`
	SourceDigests        []Digest               `json:"source_digests"`
	ObligationAccounting []ObligationAccounting `json:"obligation_accounting"`
	OutputSchemaID       string                 `json:"output_schema_id"`
}

type CompileRequest struct {
	Bundle     Bundle
	WorkflowID string
	ProfileIDs []string
	Projection ProjectionKind
	Goal       string
	Inputs     []AuthoritativeInput
}

type Compilation struct {
	IR                PromptIR
	Text              string
	SourceDigests     []Digest
	Accounting        []ObligationAccounting
	SourceSetSHA256   string
	IRSHA256          string
	TextSHA256        string
	AccountingSHA256  string
	CompilationSHA256 string
}

type renderedObligation struct {
	ID                string `json:"id"`
	SourceRequirement string `json:"source_requirement"`
	Criticality       string `json:"criticality"`
	Rule              string `json:"rule"`
	EnforcementLocus  string `json:"enforcement_locus"`
}

type goalSection struct {
	Goal       string   `json:"goal"`
	Transition string   `json:"transition"`
	WorkflowID string   `json:"workflow_id"`
	ProfileIDs []string `json:"profile_ids"`
}

type inputSection struct {
	Inputs    []AuthoritativeInput `json:"inputs"`
	Knowledge []Knowledge          `json:"knowledge"`
}

type invariantSection struct {
	Obligations       []renderedObligation `json:"obligations"`
	ProhibitedEffects []string             `json:"prohibited_effects"`
}

type authoritySection struct {
	Projection      ProjectionKind `json:"projection"`
	RiskFloor       string         `json:"risk_floor"`
	ApprovalFloor   string         `json:"approval_floor"`
	EffectCeiling   string         `json:"effect_ceiling"`
	Authority       []string       `json:"authority"`
	Capabilities    []string       `json:"capabilities"`
	Tools           []string       `json:"tools"`
	DeclarationOnly bool           `json:"declaration_only"`
}

type acceptanceSection struct {
	Accounting []ObligationAccounting `json:"obligation_accounting"`
	Success    []string               `json:"success"`
	Evidence   []string               `json:"evidence"`
}

type budgetSection struct {
	BudgetID            string   `json:"budget_id"`
	ToolCalls           int64    `json:"tool_calls"`
	Subagents           int64    `json:"subagents"`
	SubagentsRequired   int64    `json:"subagents_required"`
	SingleAgentFallback string   `json:"single_agent_fallback"`
	WallTimeSeconds     int64    `json:"wall_time_seconds"`
	Tokens              int64    `json:"tokens"`
	ContextBytes        int64    `json:"context_bytes"`
	ContextItems        int64    `json:"context_items"`
	OutputBytes         int64    `json:"output_bytes"`
	MonetaryMicroUSD    int64    `json:"monetary_micro_usd"`
	Retries             int64    `json:"retries_per_operation"`
	IdenticalFailures   int64    `json:"identical_failures"`
	Exhaustion          string   `json:"exhaustion"`
	Stopping            []string `json:"stopping"`
	Recovery            []string `json:"recovery"`
}

type outputSection struct {
	SchemaID string         `json:"schema_id"`
	Contract OutputContract `json:"contract"`
}

func Compile(request CompileRequest) (Compilation, []Diagnostic) {
	if diagnostics := Validate(request.Bundle); len(diagnostics) != 0 {
		return Compilation{}, diagnostics
	}

	workflow, workflowOK := findWorkflow(request.Bundle.Workflows, request.WorkflowID)
	if !workflowOK {
		return Compilation{}, []Diagnostic{newDiagnostic("SEM-186", request.WorkflowID, "compile request references an unknown workflow", "select the active reference workflow")}
	}
	var diagnostics []Diagnostic
	if request.Projection != ProjectionStockA0 && request.Projection != ProjectionControlledClient {
		diagnostics = addDiagnostic(diagnostics, "SEM-186", string(request.Projection), "projection is unknown", "use stock-a0 or controlled-client")
	}
	if !validCompileText(request.Goal, MaxStringBytes) {
		diagnostics = addDiagnostic(diagnostics, "SEM-190", "goal", "goal is empty, malformed, or exceeds 65536 bytes", "supply one bounded UTF-8 goal")
	}

	profileIDs := append([]string(nil), request.ProfileIDs...)
	if len(profileIDs) == 0 || len(profileIDs) > 3 {
		diagnostics = addDiagnostic(diagnostics, "SEM-186", "profiles", "profile selection is outside 1..3", "select a bounded active profile set")
	}
	diagnostics = appendDiagnostics(diagnostics, sortedUnique("compile:profiles", profileIDs)...)
	profiles := make([]Profile, 0, len(profileIDs))
	for _, id := range profileIDs {
		profile, ok := findProfile(request.Bundle.Profiles, id)
		if !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-186", id, "compile request references an unknown profile", "select an active approved profile")
			continue
		}
		profiles = append(profiles, profile)
	}
	for left := 0; left < len(profiles); left++ {
		for right := left + 1; right < len(profiles); right++ {
			if profilesConflict(profiles[left], profiles[right]) {
				diagnostics = addDiagnostic(diagnostics, "SEM-188", profiles[left].ID+"+"+profiles[right].ID, "selected profiles have an explicit applicability contraindication", "select a non-conflicting approved profile set")
			}
		}
	}

	inputs := cloneInputs(request.Inputs)
	if len(inputs) == 0 || len(inputs) > 256 {
		diagnostics = addDiagnostic(diagnostics, "SEM-190", "inputs", "authoritative input count is outside 1..256", "supply a bounded explicit input set")
	}
	contextBytes := len(request.Goal)
	previousID := ""
	seenPaths := make(map[string]bool, len(inputs))
	for index, input := range inputs {
		if len(input.ID) > 64 || !inputIDPattern.MatchString(input.ID) || len(input.Source.Path) > MaxStringBytes || !safeSourcePath(input.Source.Path) || !validSHA256(input.Source.SHA256) || len(input.Version) > 32 || !versionPattern.MatchString(input.Version) {
			diagnostics = addDiagnostic(diagnostics, "SEM-186", input.ID, "authoritative input identity, digest, or version is invalid", "supply a typed source-bound input")
		}
		if !validCompileText(input.Provenance, 160) || input.Trust != "authoritative" || !oneOf(input.Sensitivity, "public", "internal") || input.Freshness != "current" || !validCompileText(input.InclusionReason, 160) {
			diagnostics = addDiagnostic(diagnostics, "SEM-192", input.ID, "authoritative input provenance, trust, sensitivity, freshness, or inclusion reason is invalid", "supply current approved input metadata")
		}
		if index > 0 && input.ID <= previousID {
			diagnostics = addDiagnostic(diagnostics, "SEM-186", input.ID, "authoritative inputs are duplicated or not sorted by ID", "sort unique input IDs bytewise")
		}
		if seenPaths[input.Source.Path] {
			diagnostics = addDiagnostic(diagnostics, "SEM-186", input.Source.Path, "authoritative source path is duplicated", "retain one input per source path")
		}
		seenPaths[input.Source.Path] = true
		previousID = input.ID
		contextBytes += len(input.ID) + len(input.Source.Path) + len(input.Source.SHA256) + len(input.Version) + len(input.Provenance) + len(input.Trust) + len(input.Sensitivity) + len(input.Freshness) + len(input.InclusionReason)
	}
	if contextBytes > int(request.Bundle.Budgets[0].ContextBytes) {
		diagnostics = addDiagnostic(diagnostics, "SEM-190", "context_bytes", fmt.Sprintf("compile request context is %d bytes", contextBytes), "narrow the request below the contextual budget")
	}

	sourceDigests := append([]Digest(nil), request.Bundle.SourceDigests...)
	for index, digest := range sourceDigests {
		if len(digest.Path) > MaxStringBytes || !safeSourcePath(digest.Path) || !validSHA256(digest.SHA256) || index > 0 && digest.Path <= sourceDigests[index-1].Path {
			diagnostics = addDiagnostic(diagnostics, "SEM-191", digest.Path, "bundle source digest set is malformed or unordered", "decode the exact source bundle again")
		}
	}
	if len(sourceDigests) != 19 {
		diagnostics = addDiagnostic(diagnostics, "SEM-191", "source_digests", fmt.Sprintf("reference bundle has %d raw source identities, want 19", len(sourceDigests)), "decode the complete exact source bundle before compilation")
	}
	if len(diagnostics) != 0 {
		return Compilation{}, finishDiagnostics(diagnostics)
	}

	riskFloor, approvalFloor, effectCeiling := workflow.RiskFloor, workflow.ApprovalGate, workflow.EffectCeiling
	selectedObligations := make(map[string]bool)
	selectedReferences := make(map[string]bool)
	for _, profile := range profiles {
		riskFloor = higherRank(riskFloor, profile.RiskFloor)
		approvalFloor = higherRank(approvalFloor, profile.ApprovalFloor)
		effectCeiling = lowerRank(effectCeiling, profile.EffectCeiling)
		for _, id := range profile.ObligationIDs {
			selectedObligations[id] = true
		}
		for _, id := range profile.ReferenceIDs {
			selectedReferences[id] = true
		}
	}
	knowledgeIDs := sortedKeys(selectedReferences)
	knowledge := make([]Knowledge, 0, len(knowledgeIDs))
	for _, id := range knowledgeIDs {
		entry, ok := findKnowledge(request.Bundle.Knowledge, id)
		if !ok || !containsString(workflow.References, id) {
			diagnostics = addDiagnostic(diagnostics, "SEM-193", id, "selected profile knowledge is missing or outside the workflow reference set", "restore a selectable workflow-bound reference")
			continue
		}
		knowledge = append(knowledge, entry)
	}
	if len(diagnostics) != 0 {
		return Compilation{}, finishDiagnostics(diagnostics)
	}

	obligationByID := make(map[string]Obligation, len(request.Bundle.Obligations))
	for _, obligation := range request.Bundle.Obligations {
		obligationByID[obligation.ID] = obligation
	}
	accounting := make([]ObligationAccounting, 0, len(workflow.ObligationIDs))
	rendered := make([]renderedObligation, 0, len(workflow.ObligationIDs))
	renderedBytes := 0
	for _, id := range workflow.ObligationIDs {
		obligation, ok := obligationByID[id]
		if !ok {
			diagnostics = addDiagnostic(diagnostics, "SEM-184", id, "workflow obligation is dangling", "restore the exact source-derived obligation")
			continue
		}
		if obligation.Criticality != "noncritical" && strings.Contains(strings.ToLower(obligation.EnforcementLocus), "prompt-only") {
			diagnostics = addDiagnostic(diagnostics, "SEM-185", id, "prompt-only enforcement cannot support a material obligation claim", "bind deterministic enforcement or mark the obligation machine-only")
		}
		disposition := "rendered"
		if obligation.MachineOnly {
			disposition = "machine-only"
		} else if !selectedObligations[id] {
			disposition = "not-applicable"
		} else if !containsString(obligation.RequiredRenderers, string(request.Projection)) {
			diagnostics = addDiagnostic(diagnostics, "SEM-184", id, "selected projection is not an approved renderer for the obligation", "restore projection coverage without weakening the obligation")
		} else {
			if len(obligation.Rule) > MaxOutputBytes-renderedBytes {
				diagnostics = addDiagnostic(diagnostics, "SEM-190", id, "rendered obligation rules exceed the output budget", "narrow or separately version the semantic source")
				continue
			}
			renderedBytes += len(obligation.Rule)
			rendered = append(rendered, renderedObligation{
				ID:                obligation.ID,
				SourceRequirement: obligation.SourceRequirement,
				Criticality:       obligation.Criticality,
				Rule:              obligation.Rule,
				EnforcementLocus:  obligation.EnforcementLocus,
			})
		}
		accounting = append(accounting, ObligationAccounting{
			ObligationID:      obligation.ID,
			SourceRequirement: obligation.SourceRequirement,
			Disposition:       disposition,
			RuleSHA256:        sha256Hex([]byte(obligation.Rule)),
			GraderIDs:         append([]string(nil), obligation.GraderIDs...),
			PublicCaseIDs:     append([]string(nil), obligation.PublicCaseIDs...),
		})
	}
	if len(accounting) != 29 {
		diagnostics = addDiagnostic(diagnostics, "SEM-184", "obligation_accounting", fmt.Sprintf("compiled accounting has %d rows, want 29", len(accounting)), "restore exact workflow accounting")
	}
	if len(diagnostics) != 0 {
		return Compilation{}, finishDiagnostics(diagnostics)
	}

	authority := []string{"no-runtime-capability", "provider-neutral-generated-text"}
	capabilities := []string{}
	effectiveCeiling := "A0"
	declarationOnly := false
	if request.Projection == ProjectionControlledClient {
		authority = append([]string(nil), workflow.Authority...)
		capabilities = append([]string(nil), workflow.Capabilities...)
		effectiveCeiling = effectCeiling
		declarationOnly = true
	}
	sections := []PromptSection{
		{Name: promptSectionNames[0], Content: marshalPromptSection(goalSection{Goal: request.Goal, Transition: workflow.Lifecycle.Transition, WorkflowID: workflow.ID, ProfileIDs: append([]string(nil), profileIDs...)})},
		{Name: promptSectionNames[1], Content: marshalPromptSection(inputSection{Inputs: inputs, Knowledge: knowledge})},
		{Name: promptSectionNames[2], Content: marshalPromptSection(invariantSection{Obligations: rendered, ProhibitedEffects: []string{"A5", "background-behavior", "self-modification", "unapproved-authority-expansion"}})},
		{Name: promptSectionNames[3], Content: marshalPromptSection(authoritySection{Projection: request.Projection, RiskFloor: riskFloor, ApprovalFloor: approvalFloor, EffectCeiling: effectiveCeiling, Authority: authority, Capabilities: capabilities, Tools: []string{}, DeclarationOnly: declarationOnly})},
		{Name: promptSectionNames[4], Content: marshalPromptSection(acceptanceSection{Accounting: cloneAccounting(accounting), Success: append([]string(nil), workflow.Success...), Evidence: []string{"generated text is disposable and non-authoritative", "grader and public-case bindings are declarations until evaluated"}})},
		{Name: promptSectionNames[5], Content: marshalPromptSection(budgetSection{BudgetID: request.Bundle.Budgets[0].ID, ToolCalls: request.Bundle.Budgets[0].ToolCalls, Subagents: request.Bundle.Budgets[0].Subagents, SubagentsRequired: 0, SingleAgentFallback: request.Bundle.Delegations[0].SingleAgentFallback, WallTimeSeconds: request.Bundle.Budgets[0].WallTimeSeconds, Tokens: request.Bundle.Budgets[0].Tokens, ContextBytes: request.Bundle.Budgets[0].ContextBytes, ContextItems: request.Bundle.Budgets[0].ContextItems, OutputBytes: request.Bundle.Budgets[0].OutputBytes, MonetaryMicroUSD: request.Bundle.Budgets[0].MonetaryMicroUSD, Retries: request.Bundle.Budgets[0].RetriesPerOperation, IdenticalFailures: request.Bundle.Budgets[0].IdenticalFailures, Exhaustion: request.Bundle.Budgets[0].Exhaustion, Stopping: append([]string(nil), workflow.Stopping...), Recovery: append([]string(nil), workflow.Recovery...)})},
		{Name: promptSectionNames[6], Content: marshalPromptSection(outputSection{SchemaID: workflow.OutputSchema, Contract: request.Bundle.Outputs[0]})},
	}
	ir := PromptIR{
		RecordMeta:           compilationRecordMeta(),
		Sections:             sections,
		Projection:           request.Projection,
		WorkflowID:           workflow.ID,
		ProfileIDs:           append([]string(nil), profileIDs...),
		SourceDigests:        append([]Digest(nil), sourceDigests...),
		ObligationAccounting: cloneAccounting(accounting),
		OutputSchemaID:       workflow.OutputSchema,
	}
	text, renderDiagnostics := renderTemplate(request.Bundle.Template, sections)
	diagnostics = appendDiagnostics(diagnostics, renderDiagnostics...)
	if len(diagnostics) != 0 {
		return Compilation{}, finishDiagnostics(diagnostics)
	}
	if strings.Contains(text, "L7_SYNTHETIC_"+"CANARY_") {
		return Compilation{}, []Diagnostic{newDiagnostic("SEM-194", "compiled_output", "synthetic canary material would enter a generated projection", "remove the canary-bearing input and preserve nonleakage")}
	}

	irBytes := mustJSONLine(ir)
	accountingBytes := mustJSONLine(accounting)
	if len(irBytes) > MaxOutputBytes || len(accountingBytes) > MaxOutputBytes || len(text) > MaxOutputBytes {
		return Compilation{}, []Diagnostic{newDiagnostic("SEM-190", "compiled_output", "IR, accounting, or rendered text exceeds 131072 bytes", "narrow the bounded semantic request")}
	}
	irDigest := sha256Hex(irBytes)
	textDigest := sha256Hex([]byte(text))
	accountingDigest := sha256Hex(accountingBytes)
	frame := "L7-COMPILATION-v1\n" +
		"ir_sha256 " + irDigest + "\n" +
		"text_sha256 " + textDigest + "\n" +
		"accounting_sha256 " + accountingDigest + "\n"

	return Compilation{
		IR:                ir,
		Text:              text,
		SourceDigests:     append([]Digest(nil), sourceDigests...),
		Accounting:        cloneAccounting(accounting),
		SourceSetSHA256:   sourceSetSHA256(sourceDigests),
		IRSHA256:          irDigest,
		TextSHA256:        textDigest,
		AccountingSHA256:  accountingDigest,
		CompilationSHA256: sha256Hex([]byte(frame)),
	}, nil
}

func findWorkflow(workflows []Workflow, id string) (Workflow, bool) {
	for _, workflow := range workflows {
		if workflow.ID == id && workflow.Status == "active" {
			return workflow, true
		}
	}
	return Workflow{}, false
}

func findProfile(profiles []Profile, id string) (Profile, bool) {
	for _, profile := range profiles {
		if profile.ID == id && profile.Status == "active" {
			return profile, true
		}
	}
	return Profile{}, false
}

func findKnowledge(entries []Knowledge, id string) (Knowledge, bool) {
	for _, entry := range entries {
		if entry.ID == id && selectableKnowledge(entry) {
			return entry, true
		}
	}
	return Knowledge{}, false
}

func validCompileText(value string, limit int) bool {
	if value == "" || len(value) > limit || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character == '\r' || character == 0x7f || character < 0x20 && character != '\n' {
			return false
		}
	}
	return true
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return false
			}
		}
	}
	return true
}

func higherRank(left, right string) string {
	if rankValue(right) > rankValue(left) {
		return right
	}
	return left
}

func lowerRank(left, right string) string {
	if rankValue(right) < rankValue(left) {
		return right
	}
	return left
}

func rankValue(value string) int {
	if len(value) < 2 {
		return -1
	}
	return int(value[len(value)-1] - '0')
}

func containsString(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}

func profilesConflict(left, right Profile) bool {
	for _, contraindication := range left.Contraindications {
		if containsString(right.Applicability, contraindication) {
			return true
		}
	}
	for _, contraindication := range right.Contraindications {
		if containsString(left.Applicability, contraindication) {
			return true
		}
	}
	return false
}

func marshalPromptSection(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic("render: typed prompt section cannot fail JSON encoding")
	}
	return string(data)
}

func renderTemplate(template string, sections []PromptSection) (string, []Diagnostic) {
	rendered := template
	if len(sections) != len(promptSectionMarkers) {
		return "", []Diagnostic{newDiagnostic("SEM-180", "sections", "prompt IR does not have seven sections", "restore the exact ordered prompt IR")}
	}
	for index, section := range sections {
		if section.Name != promptSectionNames[index] || strings.Count(rendered, promptSectionMarkers[index]) != 1 {
			return "", []Diagnostic{newDiagnostic("SEM-180", section.Name, "prompt IR name or marker contract differs", "restore exact marker and section order")}
		}
		rendered = strings.Replace(rendered, promptSectionMarkers[index], section.Content, 1)
	}
	if strings.Contains(rendered, "{{L7:") {
		return "", []Diagnostic{newDiagnostic("SEM-180", "template", "unexpanded or unknown marker remains", "restore the narrow fixed marker grammar")}
	}
	return rendered, nil
}

func mustJSONLine(value any) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic("render: typed compilation cannot fail JSON encoding")
	}
	return append(data, '\n')
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func sourceSetSHA256(digests []Digest) string {
	var manifest strings.Builder
	for _, digest := range digests {
		manifest.WriteString(digest.SHA256)
		manifest.WriteString("  ")
		manifest.WriteString(digest.Path)
		manifest.WriteByte('\n')
	}
	return sha256Hex([]byte(manifest.String()))
}

func cloneInputs(inputs []AuthoritativeInput) []AuthoritativeInput {
	cloned := make([]AuthoritativeInput, len(inputs))
	copy(cloned, inputs)
	return cloned
}

func cloneAccounting(accounting []ObligationAccounting) []ObligationAccounting {
	cloned := make([]ObligationAccounting, len(accounting))
	for index, item := range accounting {
		cloned[index] = item
		cloned[index].GraderIDs = append([]string(nil), item.GraderIDs...)
		cloned[index].PublicCaseIDs = append([]string(nil), item.PublicCaseIDs...)
	}
	return cloned
}

func compilationRecordMeta() RecordMeta {
	return RecordMeta{
		ID:              "L7-WF-PROMPT-IR-REFERENCE",
		SchemaVersion:   "1.0.0",
		Version:         "1.0.0",
		Owner:           "semantic-owner",
		Reviewer:        "independent-readonly",
		ChangeGate:      "owner approval plus wave-02-design",
		Status:          "active",
		IntroducedBy:    "L7-W02-DES-001",
		Definition:      "Disposable typed prompt IR for the provider-neutral reference workflow.",
		Compatibility:   "major-version changes require a new stable record identity",
		Supersedes:      []string{},
		Replacement:     []string{},
		EarliestRemoval: "not-before-2.0.0",
		RetainedTests:   []string{"SEM-184", "SEM-188", "SEM-190", "SEM-191", "SEM-192", "SEM-193", "SEM-194"},
	}
}
