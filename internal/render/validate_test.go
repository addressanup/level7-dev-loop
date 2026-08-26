package render

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestRepositoryBundleValidatesAndDerivesExactObligations(t *testing.T) {
	bundle := loadSliceOneBundle(t)
	requireNoDiagnostics(t, Validate(bundle))

	requirementsData := readRepositoryFile(t, "docs/artifacts/requirements.md")
	backlogData := readRepositoryFile(t, "docs/artifacts/feature-backlog.md")
	requirements, diagnostics := DeriveWave02Requirements(requirementsData, backlogData)
	requireNoDiagnostics(t, diagnostics)

	expected := []string{
		"L7-AGENT-001", "L7-AGENT-002", "L7-AGENT-003",
		"L7-EVAL-001", "L7-EVAL-003", "L7-EVAL-004", "L7-EVAL-006", "L7-EVAL-008", "L7-EVAL-009",
		"L7-FLOW-001", "L7-FLOW-002", "L7-FLOW-003", "L7-FLOW-004", "L7-FLOW-005", "L7-FLOW-006", "L7-FLOW-008", "L7-FLOW-009", "L7-FLOW-010",
		"L7-HOST-001", "L7-HOST-005",
		"L7-KNOW-001", "L7-KNOW-002", "L7-KNOW-003", "L7-KNOW-004",
		"L7-NFR-033",
		"L7-PROMPT-001", "L7-PROMPT-002",
		"L7-SKILL-001", "L7-SKILL-002",
	}
	actual := make([]string, 0, len(requirements))
	ownerCounts := map[string]int{}
	for _, requirement := range requirements {
		actual = append(actual, requirement.ID)
		ownerCounts[requirement.Owner]++
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("derived requirement IDs = %v, want %v", actual, expected)
	}
	if ownerCounts["L7-BL-002"] != 23 || ownerCounts["L7-BL-003"] != 6 {
		t.Fatalf("derived owner counts = %v, want BL-002=23 BL-003=6", ownerCounts)
	}
	requireNoDiagnostics(t, ValidateRequirementCoverage(bundle, requirements))
}

func TestRequirementCoverageRejectsWeakeningAndInvention(t *testing.T) {
	requirements, diagnostics := DeriveWave02Requirements(
		readRepositoryFile(t, "docs/artifacts/requirements.md"),
		readRepositoryFile(t, "docs/artifacts/feature-backlog.md"),
	)
	requireNoDiagnostics(t, diagnostics)

	t.Run("weakened rule", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		bundle.Obligations[0].Rule = "weakened"
		requireRule(t, ValidateRequirementCoverage(bundle, requirements), "SEM-165")
	})
	t.Run("invented source", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		invented := bundle.Obligations[0]
		invented.ID = "L7-OBL-FAKE-999"
		invented.SourceRequirement = "L7-FAKE-999"
		bundle.Obligations = append(bundle.Obligations, invented)
		requireRule(t, ValidateRequirementCoverage(bundle, requirements), "SEM-166")
	})
	t.Run("missing source", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		bundle.Obligations = bundle.Obligations[1:]
		requireRule(t, ValidateRequirementCoverage(bundle, requirements), "SEM-163")
	})
}

func TestRecordIdentityVersionAndSupersessionRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
		rule   string
	}{
		{
			name: "invalid ID",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].ID = "lowercase"
			},
			rule: "SEM-120",
		},
		{
			name: "invalid version",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Version = "1.0"
			},
			rule: "SEM-121",
		},
		{
			name: "unknown status",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Status = "unknown"
			},
			rule: "SEM-123",
		},
		{
			name: "omitted lifecycle array",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Replacement = nil
			},
			rule: "SEM-124",
		},
		{
			name: "deprecated without replacement",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Status = "deprecated"
			},
			rule: "SEM-128",
		},
		{
			name: "stable ID redefinition",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[1].ID = bundle.Taxonomies[0].ID
				bundle.Taxonomies[1].Definition += " changed"
			},
			rule: "SEM-125",
		},
		{
			name: "supersession cycle",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Supersedes = []string{bundle.Taxonomies[1].ID}
				bundle.Taxonomies[1].Supersedes = []string{bundle.Taxonomies[0].ID}
			},
			rule: "SEM-126",
		},
		{
			name: "dangling replacement",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].Replacement = []string{"L7-TAX-UNKNOWN"}
			},
			rule: "SEM-127",
		},
		{
			name: "unordered retained tests",
			mutate: func(bundle *Bundle) {
				bundle.Taxonomies[0].RetainedTests = []string{"SEM-144", "SEM-140"}
			},
			rule: "SEM-129",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadSliceOneBundle(t)
			test.mutate(&bundle)
			requireRule(t, Validate(bundle), test.rule)
		})
	}
}

func TestTaxonomyMatrixAndLifecycleSemantics(t *testing.T) {
	fields := []string{"meaning", "entry", "exit", "failure", "blocked", "stale", "superseded"}
	for _, field := range fields {
		t.Run("missing "+field, func(t *testing.T) {
			bundle := loadSliceOneBundle(t)
			taxonomy := lifecycleTaxonomy(t, &bundle)
			switch field {
			case "meaning":
				taxonomy.Values[0].Meaning = ""
			case "entry":
				taxonomy.Values[0].Entry = ""
			case "exit":
				taxonomy.Values[0].Exit = ""
			case "failure":
				taxonomy.Values[0].Failure = ""
			case "blocked":
				taxonomy.Values[0].Blocked = ""
			case "stale":
				taxonomy.Values[0].Stale = ""
			case "superseded":
				taxonomy.Values[0].Superseded = ""
			}
			requireRule(t, Validate(bundle), "SEM-143")
		})
	}
	t.Run("value order", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		taxonomy := lifecycleTaxonomy(t, &bundle)
		taxonomy.Values[0], taxonomy.Values[1] = taxonomy.Values[1], taxonomy.Values[0]
		requireRule(t, Validate(bundle), "SEM-144")
	})
	t.Run("transition order", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		taxonomy := lifecycleTaxonomy(t, &bundle)
		taxonomy.AllowedTransitions[0], taxonomy.AllowedTransitions[1] = taxonomy.AllowedTransitions[1], taxonomy.AllowedTransitions[0]
		requireRule(t, Validate(bundle), "SEM-144")
	})
	t.Run("unknown transition", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		taxonomy := lifecycleTaxonomy(t, &bundle)
		taxonomy.AllowedTransitions[0].To = "unknown"
		requireRule(t, Validate(bundle), "SEM-143")
	})
	t.Run("duplicate invalid combination", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		taxonomy := lifecycleTaxonomy(t, &bundle)
		taxonomy.InvalidCombinations = []string{"L7-GUARD-NA-REASON", "L7-GUARD-NA-REASON"}
		requireRule(t, Validate(bundle), "SEM-129")
	})
}

func TestObligationRendererGraderAndEvidenceAccounting(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Obligation)
		rule   string
	}{
		{name: "grader", mutate: func(value *Obligation) { value.GraderIDs = nil }, rule: "SEM-154"},
		{name: "renderer", mutate: func(value *Obligation) { value.RequiredRenderers = nil }, rule: "SEM-155"},
		{name: "machine conflict", mutate: func(value *Obligation) { value.MachineOnly = true }, rule: "SEM-156"},
		{name: "public case", mutate: func(value *Obligation) { value.PublicCaseIDs = nil }, rule: "SEM-157"},
		{name: "unknown renderer", mutate: func(value *Obligation) { value.RequiredRenderers = []string{"unknown"} }, rule: "SEM-158"},
		{name: "omitted applicability", mutate: func(value *Obligation) { value.Applicability.Hosts = nil }, rule: "SEM-159"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadSliceOneBundle(t)
			test.mutate(&bundle.Obligations[0])
			requireRule(t, Validate(bundle), test.rule)
		})
	}
}

func TestGuardrailRosterAndKnowledgeSelectionMetadata(t *testing.T) {
	t.Run("guardrail roster", func(t *testing.T) {
		bundle := loadSliceOneBundle(t)
		bundle.Guardrails[0].ID = "L7-GUARD-INVENTED"
		requireRule(t, Validate(bundle), "SEM-148")
	})
	tests := []struct {
		name   string
		mutate func(*Knowledge)
	}{
		{name: "freshness", mutate: func(value *Knowledge) { value.FreshnessDays = 0 }},
		{name: "review order", mutate: func(value *Knowledge) { value.NextReview = "2020-01-01" }},
		{name: "invalid calendar date", mutate: func(value *Knowledge) { value.NextReview = "2027-02-30" }},
		{name: "freshness window", mutate: func(value *Knowledge) { value.NextReview = "2028-08-26" }},
		{name: "license", mutate: func(value *Knowledge) { value.License = "" }},
		{name: "restricted normative", mutate: func(value *Knowledge) { value.SourceStatus = "restricted"; value.Normative = true }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadSliceOneBundle(t)
			test.mutate(&bundle.Knowledge[0])
			requireRule(t, Validate(bundle), "SEM-149")
		})
	}
}

func TestProfileCompositionRules(t *testing.T) {
	bundle := loadSliceOneBundle(t)
	bundle.Descriptors = nil
	bundle.Profiles = validTestProfiles(bundle)
	requireNoDiagnostics(t, Validate(bundle))

	t.Run("floor and composition", func(t *testing.T) {
		changed := loadSliceOneBundle(t)
		changed.Descriptors = nil
		changed.Profiles = validTestProfiles(changed)
		changed.Profiles[0].Composition = "average-down"
		requireRule(t, Validate(changed), "SEM-177")
	})
	t.Run("unknown obligation", func(t *testing.T) {
		changed := loadSliceOneBundle(t)
		changed.Descriptors = nil
		changed.Profiles = validTestProfiles(changed)
		changed.Profiles[0].ObligationIDs = []string{"L7-OBL-UNKNOWN-999"}
		requireRule(t, Validate(changed), "SEM-178")
	})
	t.Run("roster", func(t *testing.T) {
		changed := loadSliceOneBundle(t)
		changed.Descriptors = nil
		changed.Profiles = validTestProfiles(changed)[:2]
		requireRule(t, Validate(changed), "SEM-179")
	})
}

func TestBudgetAndDelegationLimits(t *testing.T) {
	bundle := validTestWorkflowBundle(t)
	requireNoDiagnostics(t, Validate(bundle))

	t.Run("budget", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Budgets[0].ToolCalls = 65
		requireRule(t, Validate(changed), "SEM-175")
	})
	t.Run("delegation fallback", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Delegations[0].SingleAgentFallback = ""
		requireRule(t, Validate(changed), "SEM-176")
	})
	t.Run("delegation scope order", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Delegations[0].DisjointScope = []string{"z", "a"}
		requireRule(t, Validate(changed), "SEM-129")
	})
	t.Run("trigger normalization", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Workflows[0].PositiveTriggers = []string{"compile approved semantic contract", "compile approved semantic contract!"}
		requireRule(t, Validate(changed), "SEM-173")
	})
	t.Run("required workflow set", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Workflows[0].Inputs = nil
		requireRule(t, Validate(changed), "SEM-182")
	})
	t.Run("typed output", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Outputs[0].Owner = ""
		requireRule(t, Validate(changed), "SEM-181")
	})
	t.Run("template marker", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Template = strings.Replace(changed.Template, "{{L7:TYPED_OUTPUT}}", "{{L7:GOAL_TRANSITION}}", 1)
		requireRule(t, Validate(changed), "SEM-180")
	})
	t.Run("template prose", func(t *testing.T) {
		changed := validTestWorkflowBundle(t)
		changed.Template += "invented normative prose\n"
		requireRule(t, Validate(changed), "SEM-180")
	})
}

func TestSchemaDescriptorParity(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SchemaDescriptor)
		rule   string
	}{
		{
			name: "missing property",
			mutate: func(descriptor *SchemaDescriptor) {
				delete(descriptor.Properties, "kind")
			},
			rule: "SEM-132",
		},
		{
			name: "property type",
			mutate: func(descriptor *SchemaDescriptor) {
				property := descriptor.Properties["kind"]
				property.Type = "integer"
				descriptor.Properties["kind"] = property
			},
			rule: "SEM-134",
		},
		{
			name: "property description",
			mutate: func(descriptor *SchemaDescriptor) {
				property := descriptor.Properties["kind"]
				property.Description = ""
				descriptor.Properties["kind"] = property
			},
			rule: "SEM-134",
		},
		{
			name: "nested object openness",
			mutate: func(descriptor *SchemaDescriptor) {
				property := descriptor.Properties["values"]
				open := true
				property.Items.AdditionalProperties = &open
				descriptor.Properties["values"] = property
			},
			rule: "SEM-134",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadSliceOneBundle(t)
			descriptor := taxonomyDescriptor(t, &bundle)
			test.mutate(descriptor)
			requireRule(t, Validate(bundle), test.rule)
		})
	}
}

func TestDiagnosticsAreBoundedAndSorted(t *testing.T) {
	bundle := Bundle{Taxonomies: make([]Taxonomy, 100)}
	diagnostics := Validate(bundle)
	if len(diagnostics) == 0 || len(diagnostics) > MaxDiagnostics {
		t.Fatalf("diagnostic count = %d, want 1..%d", len(diagnostics), MaxDiagnostics)
	}
	for index, diagnostic := range diagnostics {
		if len(diagnostic.Rule)+len(diagnostic.Subject)+len(diagnostic.Message)+len(diagnostic.Next) > MaxDiagnosticBytes {
			t.Fatalf("diagnostic %d exceeds byte bound: %+v", index, diagnostic)
		}
		if index > 0 && diagnosticKey(diagnostics[index-1]) > diagnosticKey(diagnostic) {
			t.Fatalf("diagnostics are not sorted at %d: %+v", index, diagnostics)
		}
	}
}

func lifecycleTaxonomy(t *testing.T, bundle *Bundle) *Taxonomy {
	t.Helper()
	for index := range bundle.Taxonomies {
		if bundle.Taxonomies[index].Kind == "lifecycle" {
			return &bundle.Taxonomies[index]
		}
	}
	t.Fatal("lifecycle taxonomy not found")
	return nil
}

func taxonomyDescriptor(t *testing.T, bundle *Bundle) *SchemaDescriptor {
	t.Helper()
	for index := range bundle.Descriptors {
		if bundle.Descriptors[index].ID == "schemas/semantic/taxonomy.schema.json" {
			return &bundle.Descriptors[index]
		}
	}
	t.Fatal("taxonomy descriptor not found")
	return nil
}

func validTestProfiles(bundle Bundle) []Profile {
	ids := []string{
		"L7-PROF-BEHAVIOR-PRESERVING-REFACTOR",
		"L7-PROF-FEATURE-CHANGE",
		"L7-PROF-GENERIC",
	}
	profiles := make([]Profile, 0, len(ids))
	for _, id := range ids {
		profiles = append(profiles, Profile{
			RecordMeta:        testRecordMeta(id, "Test profile "+id+"."),
			Description:       "Typed profile used to validate composition.",
			Applicability:     []string{"wave-02"},
			Contraindications: []string{},
			ObligationIDs:     []string{bundle.Obligations[0].ID},
			RiskFloor:         "R1",
			EffectCeiling:     "A2",
			ApprovalFloor:     "AP1",
			ReferenceIDs:      []string{bundle.Knowledge[0].ID},
			BudgetID:          "L7-BUDGET-W02-DEV-001",
			Composition:       "obligation-union-highest-risk-approval-minimum-effect",
		})
	}
	return profiles
}

func validTestWorkflowBundle(t *testing.T) Bundle {
	t.Helper()
	bundle := loadSliceOneBundle(t)
	bundle.Descriptors = nil
	obligationIDs := make([]string, 0, len(bundle.Obligations))
	for _, obligation := range bundle.Obligations {
		obligationIDs = append(obligationIDs, obligation.ID)
	}
	sort.Strings(obligationIDs)
	bundle.Profiles = validTestProfiles(bundle)
	bundle.Workflows = []Workflow{{
		RecordMeta:       testRecordMeta("L7-WF-REFERENCE", "Reference workflow for deterministic semantic compilation."),
		Description:      "Compile provider-neutral Level 7 contracts: bounded semantic reference workflow.",
		PositiveTriggers: []string{"compile approved semantic contract"},
		NegativeTriggers: []string{"deploy production system"},
		Prerequisites:    []string{"approved source-bound semantic inputs"},
		Inputs:           []string{"goal and authoritative input digests"},
		Lifecycle: LifecycleRule{
			Entry:         "approved workflow and profiles are valid",
			Exit:          "both projections preserve obligation accounting",
			Transition:    "frame-to-verify semantic compilation",
			AllowedRepeat: []string{},
			AllowedSkip:   []string{},
		},
		Profiles:      append([]string(nil), exactProfileIDs...),
		ObligationIDs: obligationIDs,
		RiskFloor:     "R1",
		EffectCeiling: "A2",
		ApprovalGate:  "AP1",
		Authority:     []string{"repository-bounded-pure-compilation"},
		Capabilities:  []string{"controlled-client-declaration", "single-agent-fallback", "stock-a0-projection"},
		References:    []string{bundle.Knowledge[0].ID},
		Budget:        "L7-BUDGET-W02-DEV-001",
		OutputSchema:  "schemas/semantic/output.schema.json",
		Success:       []string{"typed compilation with exact accounting"},
		Failure:       []string{"bounded diagnostics and no partial result"},
		Stopping:      []string{"budget or repeated semantic failure"},
		Recovery:      []string{"correct authoritative semantic input"},
		Fixtures:      []string{"fixtures/public/bl-002/semantic-cases.json"},
	}}
	bundle.Budgets = []Budget{{
		RecordMeta:          testRecordMeta("L7-BUDGET-W02-DEV-001", "Contextual Wave 2 local development budget."),
		MeasurementScope:    "visible-transition",
		ToolCalls:           64,
		Subagents:           4,
		RetriesPerOperation: 2,
		IdenticalFailures:   2,
		WallTimeSeconds:     900,
		Tokens:              200000,
		ContextBytes:        1048576,
		ContextItems:        256,
		OutputBytes:         MaxOutputBytes,
		MonetaryMicroUSD:    0,
		Exhaustion:          "blocked",
		Recovery:            "narrow the bounded request",
	}}
	bundle.Delegations = []Delegation{{
		RecordMeta:          testRecordMeta("L7-EVAL-DELEGATION-001", "Validation-only optional delegation manifest."),
		Objective:           "validate one disjoint read-only scope",
		DisjointScope:       []string{"semantic-only"},
		Inputs:              []string{"bound-source-digests"},
		Authority:           "no-approval-power",
		EffectCeiling:       "A2",
		AllowedTools:        []string{"read-only-local"},
		BudgetID:            "L7-BUDGET-W02-DEV-001",
		OutputSchemaID:      "schemas/semantic/output.schema.json",
		Evidence:            []string{"typed-result"},
		Verifier:            "wave-integrator",
		IntegrationOwner:    "wave-integrator",
		Termination:         []string{"budget-exhausted"},
		SingleAgentFallback: "perform the same bounded validation serially",
	}}
	bundle.Outputs = []OutputContract{{
		RecordMeta:     testRecordMeta("L7-EVAL-OUTPUT-001", "Typed semantic decision envelope."),
		Decision:       []string{"blocked", "fail", "pass"},
		RuleIDs:        "sorted stable rule IDs",
		Scope:          "bound compilation scope",
		SourceIdentity: "raw source digests",
		Evidence:       "typed evidence state",
		Uncertainty:    "explicit uncertainty",
		Assumptions:    "explicit assumptions",
		Defeaters:      "explicit defeaters",
		ResidualRisk:   "highest residual risk",
		Blocker:        "first stable blocker",
		Owner:          "accountable-owner",
		NextAction:     "one permitted next action",
		Effect:         "observed effect class",
		Authority:      "bound authority",
		Diagnostics:    "bounded sorted diagnostics",
	}}
	bundle.Template = "# Level 7 semantic contract\n\n## Goal transition\n{{L7:GOAL_TRANSITION}}\n\n## Authoritative inputs\n{{L7:AUTHORITATIVE_INPUTS}}\n\n## Invariants and prohibited effects\n{{L7:INVARIANTS_PROHIBITED_EFFECTS}}\n\n## Authority, tools, capabilities, risk, and effect\n{{L7:AUTHORITY_TOOLS_CAPABILITIES_RISK_EFFECT}}\n\n## Acceptance and evidence\n{{L7:ACCEPTANCE_EVIDENCE}}\n\n## Budgets, stopping, and escalation\n{{L7:BUDGETS_STOPPING_ESCALATION}}\n\n## Typed output\n{{L7:TYPED_OUTPUT}}\n"
	return bundle
}

func testRecordMeta(id, definition string) RecordMeta {
	return RecordMeta{
		ID:              id,
		SchemaVersion:   "1.0.0",
		Version:         "1.0.0",
		Owner:           "semantic-owner",
		Reviewer:        "independent-readonly",
		ChangeGate:      "owner approval plus wave-02-design",
		Status:          "active",
		IntroducedBy:    "L7-W02-DES-001",
		Definition:      definition,
		Compatibility:   "major-version changes require a new stable record identity",
		Supersedes:      []string{},
		Replacement:     []string{},
		EarliestRemoval: "not-before-2.0.0",
		RetainedTests:   []string{"SEM-TEST"},
	}
}

func diagnosticKey(value Diagnostic) string {
	return strings.Join([]string{value.Rule, value.Subject, value.Message, value.Next}, "\x00")
}
