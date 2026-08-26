package render

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var sliceTwoSourcePaths = []string{
	"schemas/semantic/budget.schema.json",
	"schemas/semantic/delegation.schema.json",
	"schemas/semantic/output.schema.json",
	"schemas/semantic/profile.schema.json",
	"schemas/semantic/prompt-ir.schema.json",
	"schemas/semantic/workflow.schema.json",
	"semantic/profiles/behavior-preserving-refactor.json",
	"semantic/profiles/feature-change.json",
	"semantic/profiles/generic.json",
	"semantic/workflows/reference/contract.json",
	"semantic/workflows/reference/prompt.md.tmpl",
}

type semanticFixtureDocument struct {
	SchemaVersion string                `json:"schema_version"`
	Cases         []semanticFixtureCase `json:"cases"`
}

type semanticFixtureCase struct {
	ID                        string               `json:"id"`
	FixtureKind               string               `json:"fixture_kind"`
	Scenario                  string               `json:"scenario"`
	WorkflowID                string               `json:"workflow_id"`
	ProfileIDs                []string             `json:"profile_ids"`
	Projection                ProjectionKind       `json:"projection"`
	Goal                      string               `json:"goal"`
	Inputs                    []AuthoritativeInput `json:"inputs"`
	RequestState              string               `json:"request_state"`
	ApplicabilityContext      string               `json:"applicability_context"`
	ExpectedDecision          string               `json:"expected_decision"`
	ExpectedRuleID            string               `json:"expected_rule_id"`
	ExpectedCompilationSHA256 string               `json:"expected_compilation_sha256"`
}

type brokenFixtureDocument struct {
	SchemaVersion string                   `json:"schema_version"`
	Candidates    []brokenFixtureCandidate `json:"candidates"`
}

type brokenFixtureCandidate struct {
	ID               string `json:"id"`
	FixtureKind      string `json:"fixture_kind"`
	FaultClass       string `json:"fault_class"`
	IntendedRuleID   string `json:"intended_rule_id"`
	SyntheticInput   string `json:"synthetic_input"`
	ExpectedDecision string `json:"expected_decision"`
}

func TestCompleteSemanticBundleDecodesAndValidates(t *testing.T) {
	bundle := loadCompleteBundle(t)
	requireNoDiagnostics(t, Validate(bundle))
	if got, want := len(bundle.SourceDigests), 19; got != want {
		t.Fatalf("source digest count = %d, want %d", got, want)
	}
	if len(bundle.Workflows) != 1 || len(bundle.Profiles) != 3 || len(bundle.Budgets) != 1 || len(bundle.Delegations) != 1 || len(bundle.Outputs) != 1 || len(bundle.Descriptors) != 10 {
		t.Fatalf("complete semantic bundle shape: workflows=%d profiles=%d budgets=%d delegations=%d outputs=%d descriptors=%d", len(bundle.Workflows), len(bundle.Profiles), len(bundle.Budgets), len(bundle.Delegations), len(bundle.Outputs), len(bundle.Descriptors))
	}
}

func TestSemanticDescriptorVersionsAndPromptIRBoundsAreStrict(t *testing.T) {
	bundle := loadCompleteBundle(t)
	checkedPatterns := 0
	for _, descriptor := range bundle.Descriptors {
		for _, field := range []string{"schema_version", "version"} {
			pattern := descriptor.Properties[field].Pattern
			if pattern == "" {
				continue
			}
			checkedPatterns++
			expression, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("%s %s pattern: %v", descriptor.ID, field, err)
			}
			if !expression.MatchString("1.0.0") || expression.MatchString("1x0x0") || expression.MatchString("01.0.0") {
				t.Fatalf("%s %s pattern does not enforce canonical MAJOR.MINOR.PATCH: %q", descriptor.ID, field, pattern)
			}
		}
	}
	if checkedPatterns != 12 {
		t.Fatalf("strict Slice 2 semantic-version pattern count = %d, want 12", checkedPatterns)
	}

	for _, descriptor := range bundle.Descriptors {
		if descriptor.ID != "schemas/semantic/prompt-ir.schema.json" {
			continue
		}
		expected := map[string][2]int64{
			"obligation_accounting": {29, 29},
			"profile_ids":           {1, 3},
			"sections":              {7, 7},
			"source_digests":        {19, 19},
		}
		for field, bounds := range expected {
			property := descriptor.Properties[field]
			if property.MinItems == nil || property.MaxItems == nil || *property.MinItems != bounds[0] || *property.MaxItems != bounds[1] {
				t.Fatalf("prompt IR %s bounds differ from %d..%d", field, bounds[0], bounds[1])
			}
		}
		return
	}
	t.Fatal("prompt IR descriptor not found")
}

func TestSliceTwoSchemaCollectionsHaveApprovedBounds(t *testing.T) {
	bundle := loadCompleteBundle(t)
	descriptors := make(map[string]SchemaDescriptor, len(bundle.Descriptors))
	for _, descriptor := range bundle.Descriptors {
		descriptors[descriptor.ID] = descriptor
	}
	checks := []struct {
		descriptor string
		field      string
		minimum    int64
		maximum    int64
	}{
		{"schemas/semantic/delegation.schema.json", "allowed_tools", 1, 64},
		{"schemas/semantic/delegation.schema.json", "disjoint_scope", 1, 64},
		{"schemas/semantic/output.schema.json", "decision", 4, 4},
		{"schemas/semantic/profile.schema.json", "applicability", 1, 16},
		{"schemas/semantic/profile.schema.json", "obligation_ids", 1, 29},
		{"schemas/semantic/prompt-ir.schema.json", "obligation_accounting", 29, 29},
		{"schemas/semantic/prompt-ir.schema.json", "profile_ids", 1, 3},
		{"schemas/semantic/prompt-ir.schema.json", "sections", 7, 7},
		{"schemas/semantic/prompt-ir.schema.json", "source_digests", 19, 19},
		{"schemas/semantic/workflow.schema.json", "fixtures", 2, 2},
		{"schemas/semantic/workflow.schema.json", "negative_triggers", 1, 16},
		{"schemas/semantic/workflow.schema.json", "obligation_ids", 29, 29},
		{"schemas/semantic/workflow.schema.json", "positive_triggers", 1, 16},
		{"schemas/semantic/workflow.schema.json", "profiles", 3, 3},
	}
	for _, check := range checks {
		property := descriptors[check.descriptor].Properties[check.field]
		if property.MinItems == nil || property.MaxItems == nil || *property.MinItems != check.minimum || *property.MaxItems != check.maximum {
			t.Fatalf("%s %s bounds differ from %d..%d", check.descriptor, check.field, check.minimum, check.maximum)
		}
	}
}

func TestCompileBothProjectionsPreserveObligationAccounting(t *testing.T) {
	bundle := loadCompleteBundle(t)
	stockRequest := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
	controlledRequest := validCompileRequest(bundle, ProjectionControlledClient, "L7-PROF-GENERIC")
	stock, diagnostics := Compile(stockRequest)
	requireNoDiagnostics(t, diagnostics)
	controlled, diagnostics := Compile(controlledRequest)
	requireNoDiagnostics(t, diagnostics)

	if !reflect.DeepEqual(stock.Accounting, controlled.Accounting) {
		t.Fatal("projection changed obligation accounting")
	}
	if len(stock.Accounting) != 29 {
		t.Fatalf("accounting rows = %d, want 29", len(stock.Accounting))
	}
	machineOnly := 0
	rendered := 0
	for _, item := range stock.Accounting {
		switch item.Disposition {
		case "machine-only":
			machineOnly++
		case "rendered":
			rendered++
		default:
			t.Fatalf("unknown accounting disposition: %+v", item)
		}
	}
	if machineOnly != 6 || rendered != 23 {
		t.Fatalf("accounting dispositions: machine-only=%d rendered=%d", machineOnly, rendered)
	}
	if stock.Text == controlled.Text || stock.CompilationSHA256 == controlled.CompilationSHA256 {
		t.Fatal("distinct projections produced an identical framed compilation")
	}
	if stock.AccountingSHA256 != controlled.AccountingSHA256 || stock.SourceSetSHA256 != controlled.SourceSetSHA256 {
		t.Fatal("projection changed source or accounting identity")
	}
	requireCompilationDigests(t, stock)
	requireCompilationDigests(t, controlled)
}

func TestCompileComposesProfileUnionAndFloors(t *testing.T) {
	bundle := loadCompleteBundle(t)
	request := validCompileRequest(bundle, ProjectionControlledClient, "L7-PROF-GENERIC")
	request.ProfileIDs = []string{"L7-PROF-FEATURE-CHANGE", "L7-PROF-GENERIC"}
	result, diagnostics := Compile(request)
	requireNoDiagnostics(t, diagnostics)

	var authority authoritySection
	if err := json.Unmarshal([]byte(result.IR.Sections[3].Content), &authority); err != nil {
		t.Fatalf("decode authority section: %v", err)
	}
	if authority.RiskFloor != "R2" || authority.ApprovalFloor != "AP2" || authority.EffectCeiling != "A0" {
		t.Fatalf("profile floors were averaged or weakened: %+v", authority)
	}
	accounted := make(map[string]bool, len(result.Accounting))
	for _, item := range result.Accounting {
		accounted[item.ObligationID] = true
	}
	for _, profile := range bundle.Profiles {
		if !containsString(request.ProfileIDs, profile.ID) {
			continue
		}
		for _, id := range profile.ObligationIDs {
			if !accounted[id] {
				t.Fatalf("selected profile union obligation %s is absent from accounting", id)
			}
		}
	}
	var inputs inputSection
	if err := json.Unmarshal([]byte(result.IR.Sections[1].Content), &inputs); err != nil {
		t.Fatalf("decode composed knowledge section: %v", err)
	}
	if len(inputs.Knowledge) != 4 {
		t.Fatalf("composed profile knowledge count = %d, want 4", len(inputs.Knowledge))
	}
}

func TestCompileMarksNonApplicableProfileObligations(t *testing.T) {
	bundle := loadCompleteBundle(t)
	result, diagnostics := Compile(validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-BEHAVIOR-PRESERVING-REFACTOR"))
	requireNoDiagnostics(t, diagnostics)
	counts := make(map[string]int)
	for _, item := range result.Accounting {
		counts[item.Disposition]++
	}
	if counts["rendered"] != 17 || counts["machine-only"] != 6 || counts["not-applicable"] != 6 || len(result.Accounting) != 29 {
		t.Fatalf("behavior-preserving accounting dispositions = %+v", counts)
	}
}

func TestCompileIsDeterministicAndCopiesMutableInputs(t *testing.T) {
	bundle := loadCompleteBundle(t)
	request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
	first, diagnostics := Compile(request)
	requireNoDiagnostics(t, diagnostics)
	second, diagnostics := Compile(request)
	requireNoDiagnostics(t, diagnostics)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("identical compile request produced different output")
	}

	request.ProfileIDs[0] = "mutated"
	request.Inputs[0].InclusionReason = "mutated"
	bundle.SourceDigests[0].Path = "mutated"
	if first.IR.ProfileIDs[0] != "L7-PROF-GENERIC" || strings.Contains(first.IR.Sections[1].Content, "mutated") || first.SourceDigests[0].Path == "mutated" {
		t.Fatal("compilation retained caller-owned mutable slices")
	}
}

func TestCompilePreservesAuthoritativeMetadataAndCompleteBudget(t *testing.T) {
	bundle := loadCompleteBundle(t)
	request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
	result, diagnostics := Compile(request)
	requireNoDiagnostics(t, diagnostics)

	var inputs inputSection
	if err := json.Unmarshal([]byte(result.IR.Sections[1].Content), &inputs); err != nil {
		t.Fatalf("decode authoritative input section: %v", err)
	}
	if !reflect.DeepEqual(inputs.Inputs, request.Inputs) {
		t.Fatalf("authoritative metadata changed: got %+v want %+v", inputs.Inputs, request.Inputs)
	}
	if len(inputs.Knowledge) != 3 {
		t.Fatalf("generic profile knowledge count = %d, want 3", len(inputs.Knowledge))
	}
	for _, entry := range inputs.Knowledge {
		if entry.ID == "L7-KNOW-LAW-001" || entry.ID == "L7-KNOW-STANDARD-001" || !selectableKnowledge(entry) {
			t.Fatalf("non-applicable or unsafe knowledge was rendered: %+v", entry)
		}
	}

	var budget budgetSection
	if err := json.Unmarshal([]byte(result.IR.Sections[5].Content), &budget); err != nil {
		t.Fatalf("decode budget section: %v", err)
	}
	want := bundle.Budgets[0]
	if budget.ToolCalls != want.ToolCalls || budget.Subagents != want.Subagents || budget.SubagentsRequired != 0 || budget.SingleAgentFallback != bundle.Delegations[0].SingleAgentFallback || budget.WallTimeSeconds != want.WallTimeSeconds || budget.Tokens != want.Tokens || budget.ContextBytes != want.ContextBytes || budget.ContextItems != want.ContextItems || budget.OutputBytes != want.OutputBytes || budget.MonetaryMicroUSD != want.MonetaryMicroUSD || budget.Retries != want.RetriesPerOperation || budget.IdenticalFailures != want.IdenticalFailures {
		t.Fatalf("compiled budget omitted or changed a required ceiling: %+v", budget)
	}
}

func TestCompileRejectsBrokenObligationAccounting(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
		rule   string
	}{
		{
			name: "dropped",
			mutate: func(bundle *Bundle) {
				bundle.Obligations = bundle.Obligations[1:]
			},
			rule: "SEM-150",
		},
		{
			name: "weakened renderer",
			mutate: func(bundle *Bundle) {
				index := firstRenderedObligation(t, *bundle)
				bundle.Obligations[index].RequiredRenderers = []string{"controlled-client"}
			},
			rule: "SEM-184",
		},
		{
			name: "duplicated",
			mutate: func(bundle *Bundle) {
				bundle.Obligations = append(bundle.Obligations, bundle.Obligations[0])
			},
			rule: "SEM-125",
		},
		{
			name: "invented",
			mutate: func(bundle *Bundle) {
				invented := bundle.Obligations[0]
				invented.ID = "L7-OBL-FAKE-999"
				invented.SourceRequirement = "L7-FAKE-999"
				bundle.Obligations = append(bundle.Obligations, invented)
			},
			rule: "SEM-150",
		},
		{
			name: "prose only",
			mutate: func(bundle *Bundle) {
				index := firstRenderedObligation(t, *bundle)
				bundle.Obligations[index].EnforcementLocus = "prompt-only"
			},
			rule: "SEM-185",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadCompleteBundle(t)
			test.mutate(&bundle)
			_, diagnostics := Compile(validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC"))
			requireRule(t, diagnostics, test.rule)
		})
	}
}

func TestCompileMachineOnlyObligationsNeverBecomeProse(t *testing.T) {
	bundle := loadCompleteBundle(t)
	var machineOnly Obligation
	for _, obligation := range bundle.Obligations {
		if obligation.MachineOnly {
			machineOnly = obligation
			break
		}
	}
	if machineOnly.ID == "" {
		t.Fatal("machine-only obligation not found")
	}
	compilation, diagnostics := Compile(validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC"))
	requireNoDiagnostics(t, diagnostics)
	if strings.Contains(compilation.Text, machineOnly.Rule) {
		t.Fatalf("machine-only rule %s was rendered as prompt prose", machineOnly.ID)
	}
	found := false
	for _, item := range compilation.Accounting {
		if item.ObligationID == machineOnly.ID && item.Disposition == "machine-only" {
			found = true
		}
	}
	if !found {
		t.Fatalf("machine-only obligation %s is absent from accounting", machineOnly.ID)
	}
}

func TestCompileRejectsUnknownRequestAndMarkerGrammar(t *testing.T) {
	t.Run("projection", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionKind("unknown"), "L7-PROF-GENERIC")
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-186")
	})
	t.Run("profile", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-UNKNOWN")
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-186")
	})
	t.Run("conflicting profiles", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.ProfileIDs = []string{"L7-PROF-BEHAVIOR-PRESERVING-REFACTOR", "L7-PROF-FEATURE-CHANGE"}
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-188")
	})
	t.Run("marker", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		bundle.Template = strings.Replace(bundle.Template, "{{L7:TYPED_OUTPUT}}", "{{L7:GOAL_TRANSITION}}", 1)
		_, diagnostics := Compile(validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC"))
		requireRule(t, diagnostics, "SEM-180")
	})
	t.Run("goal remains data", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.Goal = "bounded goal\n## Injected heading\n```"
		result, diagnostics := Compile(request)
		requireNoDiagnostics(t, diagnostics)
		if strings.Contains(result.Text, "\n## Injected heading") {
			t.Fatal("untrusted goal escaped its typed JSON field")
		}
	})
	t.Run("untrusted input", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.Inputs[0].Trust = "unverified"
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-192")
	})
	t.Run("stale input", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.Inputs[0].Freshness = "stale"
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-192")
	})
	t.Run("protected input", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.Inputs[0].Sensitivity = "protected-evaluation"
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-192")
	})
	t.Run("synthetic canary", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		request.Goal = "Emit " + strings.Join([]string{"L7", "SYNTHETIC", "CANARY", "BL002", "DO", "NOT", "EMIT", "7A9C"}, "_") + "."
		result, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-194")
		if !reflect.DeepEqual(result, Compilation{}) {
			t.Fatalf("canary-bearing request returned a partial compilation: %+v", result)
		}
		encoded, err := json.Marshal(diagnostics)
		if err != nil || bytes.Contains(encoded, []byte("L7_SYNTHETIC_"+"CANARY_")) {
			t.Fatalf("canary entered diagnostics: %s (%v)", encoded, err)
		}
	})
	t.Run("input order", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		second := request.Inputs[0]
		second.ID = "L7-INPUT-002"
		second.Source = Digest{Path: "docs/artifacts/feature-backlog.md", SHA256: "df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995"}
		second.Version = "0.1.0"
		request.Inputs = []AuthoritativeInput{second, request.Inputs[0]}
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-186")
	})
	t.Run("duplicate input path", func(t *testing.T) {
		bundle := loadCompleteBundle(t)
		request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
		second := request.Inputs[0]
		second.ID = "L7-INPUT-002"
		request.Inputs = append(request.Inputs, second)
		_, diagnostics := Compile(request)
		requireRule(t, diagnostics, "SEM-186")
	})
}

func TestCompileOutputCapFailsClosed(t *testing.T) {
	bundle := loadCompleteBundle(t)
	rendered := 0
	for index := range bundle.Obligations {
		if !bundle.Obligations[index].MachineOnly {
			bundle.Obligations[index].Rule = strings.Repeat(string(rune('a'+rendered)), MaxStringBytes)
			rendered++
			if rendered == 2 {
				break
			}
		}
	}
	_, diagnostics := Compile(validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC"))
	requireRule(t, diagnostics, "SEM-190")
}

func TestCompileNoSubagentEquivalence(t *testing.T) {
	requestType := reflect.TypeOf(CompileRequest{})
	for index := 0; index < requestType.NumField(); index++ {
		name := strings.ToLower(requestType.Field(index).Name)
		if strings.Contains(name, "subagent") || strings.Contains(name, "delegation") {
			t.Fatalf("pure compiler correctness depends on optional delegation field %s", requestType.Field(index).Name)
		}
	}
	bundle := loadCompleteBundle(t)
	request := validCompileRequest(bundle, ProjectionStockA0, "L7-PROF-GENERIC")
	results := make([]Compilation, 0, 2)
	for _, delegationAvailable := range []bool{false, true} {
		// Availability is intentionally not an input to the pure compiler.
		_ = delegationAvailable
		result, diagnostics := Compile(request)
		requireNoDiagnostics(t, diagnostics)
		results = append(results, result)
	}
	if results[0].CompilationSHA256 != results[1].CompilationSHA256 || !reflect.DeepEqual(results[0].Accounting, results[1].Accounting) {
		t.Fatal("optional delegation availability changed semantic correctness")
	}
}

func TestPublicBL002FixturesAreClosedAndComplete(t *testing.T) {
	var semantic semanticFixtureDocument
	decodeFixtureExact(t, "fixtures/public/bl-002/semantic-cases.json", &semantic)
	if semantic.SchemaVersion != "1.0.0" || len(semantic.Cases) != 9 {
		t.Fatalf("semantic fixture shape: version=%s cases=%d", semantic.SchemaVersion, len(semantic.Cases))
	}
	scenarios := make([]string, 0, len(semantic.Cases))
	previousID := ""
	for _, fixture := range semantic.Cases {
		if fixture.ID <= previousID {
			t.Fatalf("semantic fixture IDs are not sorted: %s after %s", fixture.ID, previousID)
		}
		previousID = fixture.ID
		scenarios = append(scenarios, fixture.Scenario)
		if fixture.WorkflowID == "" || len(fixture.ProfileIDs) == 0 || fixture.Goal == "" || len(fixture.Inputs) == 0 || fixture.RequestState == "" || fixture.ApplicabilityContext == "" {
			t.Fatalf("semantic fixture contract is incomplete: %+v", fixture)
		}
		for _, input := range fixture.Inputs {
			if !inputIDPattern.MatchString(input.ID) || !safeSourcePath(input.Source.Path) || !validSHA256(input.Source.SHA256) || !versionPattern.MatchString(input.Version) || input.Provenance == "" || input.Trust != "authoritative" || !oneOf(input.Sensitivity, "public", "internal") || input.Freshness != "current" || input.InclusionReason == "" {
				t.Fatalf("semantic fixture input metadata is incomplete: %+v", input)
			}
		}
		if fixture.ExpectedDecision != "pass" && fixture.FixtureKind != "broken" {
			t.Fatalf("negative fixture lacks fixture_kind=broken: %+v", fixture)
		}
		if fixture.ExpectedDecision == "pass" {
			if fixture.FixtureKind != "valid" || fixture.ExpectedRuleID != "" || fixture.ExpectedCompilationSHA256 == "" {
				t.Fatalf("passing fixture outcome contract is malformed: %+v", fixture)
			}
			bundle := loadCompleteBundle(t)
			result, diagnostics := Compile(CompileRequest{Bundle: bundle, WorkflowID: fixture.WorkflowID, ProfileIDs: fixture.ProfileIDs, Projection: fixture.Projection, Goal: fixture.Goal, Inputs: fixture.Inputs})
			requireNoDiagnostics(t, diagnostics)
			t.Logf("fixture=%s compilation_sha256=%s", fixture.ID, result.CompilationSHA256)
			if fixture.ExpectedCompilationSHA256 == "" || result.CompilationSHA256 != fixture.ExpectedCompilationSHA256 {
				t.Errorf("%s compilation digest = %s, want %s", fixture.ID, result.CompilationSHA256, fixture.ExpectedCompilationSHA256)
			}
		} else if fixture.ExpectedRuleID == "" || fixture.ExpectedCompilationSHA256 != "" {
			t.Fatalf("negative fixture outcome contract is malformed: %+v", fixture)
		}
	}
	sort.Strings(scenarios)
	expectedScenarios := []string{"applicable-profile", "boundary", "context-exhaustion", "degraded", "interruption", "invalid", "non-applicable-profile", "serialization-order", "valid"}
	if !reflect.DeepEqual(scenarios, expectedScenarios) {
		t.Fatalf("semantic fixture scenarios = %v, want %v", scenarios, expectedScenarios)
	}

	var broken brokenFixtureDocument
	decodeFixtureExact(t, "fixtures/public/bl-002/broken-candidates.json", &broken)
	if broken.SchemaVersion != "1.0.0" || len(broken.Candidates) != 8 {
		t.Fatalf("broken fixture shape: version=%s candidates=%d", broken.SchemaVersion, len(broken.Candidates))
	}
	faults := make([]string, 0, len(broken.Candidates))
	previousID = ""
	for _, fixture := range broken.Candidates {
		if fixture.ID <= previousID || fixture.FixtureKind != "broken" || fixture.ExpectedDecision != "blocked" || fixture.IntendedRuleID == "" || fixture.SyntheticInput == "" {
			t.Fatalf("invalid broken fixture: %+v", fixture)
		}
		previousID = fixture.ID
		faults = append(faults, fixture.FaultClass)
	}
	sort.Strings(faults)
	expectedFaults := []string{
		"correctness-dependent-on-subagent",
		"dropped-or-weakened-critical-obligation",
		"fabricated-evidence-or-unverified-pass",
		"false-low-routing-for-high-risk",
		"forbidden-effect-or-authority-expansion",
		"invented-obligation-or-unsupported-approval",
		"stale-approval-treated-as-current",
		"synthetic-canary-leakage",
	}
	if !reflect.DeepEqual(faults, expectedFaults) {
		t.Fatalf("broken fixture faults = %v, want %v", faults, expectedFaults)
	}
}

func TestRepositoryCompilationDigests(t *testing.T) {
	bundle := loadCompleteBundle(t)
	for _, projection := range []ProjectionKind{ProjectionStockA0, ProjectionControlledClient} {
		result, diagnostics := Compile(validCompileRequest(bundle, projection, "L7-PROF-GENERIC"))
		requireNoDiagnostics(t, diagnostics)
		t.Logf("projection=%s source_set_sha256=%s ir_sha256=%s text_sha256=%s accounting_sha256=%s compilation_sha256=%s", projection, result.SourceSetSHA256, result.IRSHA256, result.TextSHA256, result.AccountingSHA256, result.CompilationSHA256)
	}
}

func loadCompleteBundle(t *testing.T) Bundle {
	t.Helper()
	paths := append(append([]string(nil), sliceOneSourcePaths...), sliceTwoSourcePaths...)
	sort.Strings(paths)
	files := make([]SourceFile, 0, len(paths))
	for _, path := range paths {
		files = append(files, SourceFile{Path: path, Data: readRepositoryFile(t, path)})
	}
	bundle, diagnostics := Decode(files)
	requireNoDiagnostics(t, diagnostics)
	return bundle
}

func validCompileRequest(bundle Bundle, projection ProjectionKind, profileID string) CompileRequest {
	return CompileRequest{
		Bundle:     bundle,
		WorkflowID: "L7-WF-REFERENCE",
		ProfileIDs: []string{profileID},
		Projection: projection,
		Goal:       "Compile the approved provider-neutral semantic contract.",
		Inputs: []AuthoritativeInput{{
			ID:              "L7-INPUT-001",
			Source:          Digest{Path: "docs/artifacts/requirements.md", SHA256: "a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f"},
			Version:         "0.2.0",
			Provenance:      "Approved artifact L7-REQ-001 in the Wave 2 source tuple.",
			Trust:           "authoritative",
			Sensitivity:     "internal",
			Freshness:       "current",
			InclusionReason: "Bind the approved normative source.",
		}},
	}
}

func requireCompilationDigests(t *testing.T, compilation Compilation) {
	t.Helper()
	irDigest := sha256Hex(mustJSONLine(compilation.IR))
	textDigest := sha256Hex([]byte(compilation.Text))
	accountingDigest := sha256Hex(mustJSONLine(compilation.Accounting))
	frame := "L7-COMPILATION-v1\n" +
		"ir_sha256 " + irDigest + "\n" +
		"text_sha256 " + textDigest + "\n" +
		"accounting_sha256 " + accountingDigest + "\n"
	if compilation.IRSHA256 != irDigest || compilation.TextSHA256 != textDigest || compilation.AccountingSHA256 != accountingDigest || compilation.CompilationSHA256 != sha256Hex([]byte(frame)) || compilation.SourceSetSHA256 != sourceSetSHA256(compilation.SourceDigests) {
		t.Fatalf("compilation digest framing mismatch: %+v", compilation)
	}
}

func decodeFixtureExact(t *testing.T, path string, target any) {
	t.Helper()
	data := readRepositoryFile(t, path)
	requireNoDiagnostics(t, validateSourceBytes(path, data))
	requireNoDiagnostics(t, scanJSON(path, data))
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("trailing fixture data in %s: %v", path, err)
	}
}

func firstRenderedObligation(t *testing.T, bundle Bundle) int {
	t.Helper()
	for index, obligation := range bundle.Obligations {
		if !obligation.MachineOnly {
			return index
		}
	}
	t.Fatal("rendered obligation not found")
	return -1
}
