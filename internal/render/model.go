package render

const (
	MaxFileBytes       = 262144
	MaxAggregateBytes  = 2097152
	MaxJSONDepth       = 32
	MaxObjectFields    = 128
	MaxStringBytes     = 65536
	MaxDiagnostics     = 64
	MaxDiagnosticBytes = 1024
	MaxOutputBytes     = 131072
)

type SourceFile struct {
	Path string
	Data []byte
}

type Digest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Diagnostic struct {
	Rule    string `json:"rule"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Next    string `json:"next"`
}

type RecordMeta struct {
	ID              string   `json:"id"`
	SchemaVersion   string   `json:"schema_version"`
	Version         string   `json:"version"`
	Owner           string   `json:"owner"`
	Reviewer        string   `json:"reviewer"`
	ChangeGate      string   `json:"change_gate"`
	Status          string   `json:"status"`
	IntroducedBy    string   `json:"introduced_by"`
	Definition      string   `json:"definition"`
	Compatibility   string   `json:"compatibility"`
	Supersedes      []string `json:"supersedes"`
	Replacement     []string `json:"replacement"`
	EarliestRemoval string   `json:"earliest_removal"`
	RetainedTests   []string `json:"retained_tests"`
}

type Transition struct {
	From string `json:"from"`
	To   string `json:"to"`
	Gate string `json:"gate"`
}

type TaxonomyValue struct {
	Value      string `json:"value"`
	Meaning    string `json:"meaning"`
	Entry      string `json:"entry"`
	Exit       string `json:"exit"`
	Failure    string `json:"failure"`
	Blocked    string `json:"blocked"`
	Stale      string `json:"stale"`
	Superseded string `json:"superseded"`
}

type Taxonomy struct {
	RecordMeta
	Kind                string          `json:"kind"`
	Values              []TaxonomyValue `json:"values"`
	AllowedTransitions  []Transition    `json:"allowed_transitions"`
	InvalidCombinations []string        `json:"invalid_combinations"`
}

type Applicability struct {
	Stages            []string `json:"stages"`
	Profiles          []string `json:"profiles"`
	Risks             []string `json:"risks"`
	Effects           []string `json:"effects"`
	Capabilities      []string `json:"capabilities"`
	Hosts             []string `json:"hosts"`
	Contraindications []string `json:"contraindications"`
}

type Obligation struct {
	RecordMeta
	SourceRequirement string        `json:"source_requirement"`
	Criticality       string        `json:"criticality"`
	Rationale         string        `json:"rationale"`
	Applicability     Applicability `json:"applicability"`
	Rule              string        `json:"rule"`
	EnforcementLocus  string        `json:"enforcement_locus"`
	RequiredRenderers []string      `json:"required_renderers"`
	MachineOnly       bool          `json:"machine_only"`
	GraderIDs         []string      `json:"grader_ids"`
	PublicCaseIDs     []string      `json:"public_case_ids"`
	EvidenceRule      string        `json:"evidence_rule"`
	Overrideability   string        `json:"overrideability"`
}

type Guardrail struct {
	RecordMeta
	Input            string   `json:"input"`
	Decision         string   `json:"decision"`
	FailureMode      string   `json:"failure_mode"`
	Recovery         string   `json:"recovery"`
	Proof            string   `json:"proof"`
	Criticality      string   `json:"criticality"`
	EnforcementLocus string   `json:"enforcement_locus"`
	GraderIDs        []string `json:"grader_ids"`
	Overrideability  string   `json:"overrideability"`
}

type Knowledge struct {
	RecordMeta
	Pointer           string   `json:"pointer"`
	SourceType        string   `json:"source_type"`
	AuthorityType     string   `json:"authority_type"`
	SourceVersion     string   `json:"source_version"`
	SourceDate        string   `json:"source_date"`
	SourceStatus      string   `json:"source_status"`
	Applicability     []string `json:"applicability"`
	Contraindications []string `json:"contraindications"`
	Jurisdiction      string   `json:"jurisdiction"`
	License           string   `json:"license"`
	UseRestriction    string   `json:"use_restriction"`
	FreshnessDays     int64    `json:"freshness_days"`
	LastReviewed      string   `json:"last_reviewed"`
	NextReview        string   `json:"next_review"`
	Normative         bool     `json:"normative"`
	Mapping           []string `json:"mapping"`
}

type LifecycleRule struct {
	Entry         string   `json:"entry"`
	Exit          string   `json:"exit"`
	Transition    string   `json:"transition"`
	AllowedRepeat []string `json:"allowed_repeat"`
	AllowedSkip   []string `json:"allowed_skip"`
}

type Workflow struct {
	RecordMeta
	Description      string        `json:"description"`
	PositiveTriggers []string      `json:"positive_triggers"`
	NegativeTriggers []string      `json:"negative_triggers"`
	Prerequisites    []string      `json:"prerequisites"`
	Inputs           []string      `json:"inputs"`
	Lifecycle        LifecycleRule `json:"lifecycle"`
	Profiles         []string      `json:"profiles"`
	ObligationIDs    []string      `json:"obligation_ids"`
	RiskFloor        string        `json:"risk_floor"`
	EffectCeiling    string        `json:"effect_ceiling"`
	ApprovalGate     string        `json:"approval_gate"`
	Authority        []string      `json:"authority"`
	Capabilities     []string      `json:"capabilities"`
	References       []string      `json:"references"`
	Budget           string        `json:"budget"`
	OutputSchema     string        `json:"output_schema"`
	Success          []string      `json:"success"`
	Failure          []string      `json:"failure"`
	Stopping         []string      `json:"stopping"`
	Recovery         []string      `json:"recovery"`
	Fixtures         []string      `json:"fixtures"`
}

type Profile struct {
	RecordMeta
	Description       string   `json:"description"`
	Applicability     []string `json:"applicability"`
	Contraindications []string `json:"contraindications"`
	ObligationIDs     []string `json:"obligation_ids"`
	RiskFloor         string   `json:"risk_floor"`
	EffectCeiling     string   `json:"effect_ceiling"`
	ApprovalFloor     string   `json:"approval_floor"`
	ReferenceIDs      []string `json:"reference_ids"`
	BudgetID          string   `json:"budget_id"`
	Composition       string   `json:"composition"`
}

type Budget struct {
	RecordMeta
	MeasurementScope    string `json:"measurement_scope"`
	ToolCalls           int64  `json:"tool_calls"`
	Subagents           int64  `json:"subagents"`
	RetriesPerOperation int64  `json:"retries_per_operation"`
	IdenticalFailures   int64  `json:"identical_failures"`
	WallTimeSeconds     int64  `json:"wall_time_seconds"`
	Tokens              int64  `json:"tokens"`
	ContextBytes        int64  `json:"context_bytes"`
	ContextItems        int64  `json:"context_items"`
	OutputBytes         int64  `json:"output_bytes"`
	MonetaryMicroUSD    int64  `json:"monetary_micro_usd"`
	Exhaustion          string `json:"exhaustion"`
	Recovery            string `json:"recovery"`
}

type Delegation struct {
	RecordMeta
	Objective           string   `json:"objective"`
	DisjointScope       []string `json:"disjoint_scope"`
	Inputs              []string `json:"inputs"`
	Authority           string   `json:"authority"`
	EffectCeiling       string   `json:"effect_ceiling"`
	AllowedTools        []string `json:"allowed_tools"`
	BudgetID            string   `json:"budget_id"`
	OutputSchemaID      string   `json:"output_schema_id"`
	Evidence            []string `json:"evidence"`
	Verifier            string   `json:"verifier"`
	IntegrationOwner    string   `json:"integration_owner"`
	Termination         []string `json:"termination"`
	SingleAgentFallback string   `json:"single_agent_fallback"`
}

type OutputContract struct {
	RecordMeta
	Decision       []string `json:"decision"`
	RuleIDs        string   `json:"rule_ids"`
	Scope          string   `json:"scope"`
	SourceIdentity string   `json:"source_identity"`
	Evidence       string   `json:"evidence"`
	Uncertainty    string   `json:"uncertainty"`
	Assumptions    string   `json:"assumptions"`
	Defeaters      string   `json:"defeaters"`
	ResidualRisk   string   `json:"residual_risk"`
	Blocker        string   `json:"blocker"`
	Owner          string   `json:"decision_owner"`
	NextAction     string   `json:"next_action"`
	Effect         string   `json:"effect"`
	Authority      string   `json:"authority"`
	Diagnostics    string   `json:"diagnostics"`
}

type SchemaProperty struct {
	Type                 string                    `json:"type"`
	Description          string                    `json:"description"`
	Pattern              string                    `json:"pattern,omitempty"`
	Enum                 []string                  `json:"enum,omitempty"`
	Minimum              *int64                    `json:"minimum,omitempty"`
	Maximum              *int64                    `json:"maximum,omitempty"`
	MinItems             *int64                    `json:"minItems,omitempty"`
	MaxItems             *int64                    `json:"maxItems,omitempty"`
	UniqueItems          *bool                     `json:"uniqueItems,omitempty"`
	AdditionalProperties *bool                     `json:"additionalProperties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Properties           map[string]SchemaProperty `json:"properties,omitempty"`
	Items                *SchemaProperty           `json:"items,omitempty"`
}

type SchemaDescriptor struct {
	Schema               string                    `json:"$schema"`
	ID                   string                    `json:"$id"`
	Title                string                    `json:"title"`
	Type                 string                    `json:"type"`
	AdditionalProperties bool                      `json:"additionalProperties"`
	Required             []string                  `json:"required"`
	Properties           map[string]SchemaProperty `json:"properties"`
}

type Bundle struct {
	Taxonomies    []Taxonomy
	Obligations   []Obligation
	Guardrails    []Guardrail
	Knowledge     []Knowledge
	Workflows     []Workflow
	Profiles      []Profile
	Budgets       []Budget
	Delegations   []Delegation
	Outputs       []OutputContract
	Descriptors   []SchemaDescriptor
	Template      string
	SourceDigests []Digest
}

type Requirement struct {
	ID    string
	Rule  string
	Owner string
}
