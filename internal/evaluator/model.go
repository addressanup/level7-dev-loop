package evaluator

import "github.com/addressanup/level7-dev-loop/internal/render"

const (
	MaxFileBytes        = 262144
	MaxAggregateBytes   = 2097152
	MaxJSONDepth        = 32
	MaxObjectFields     = 128
	MaxStringBytes      = 65536
	MaxCases            = 512
	MaxTruthLabels      = 1024
	MaxGraders          = 128
	MaxCoverageLinks    = 2048
	MaxTrials           = 1024
	MaxDiagnostics      = 64
	MaxDiagnosticBytes  = 1024
	MaxDiagnosticsBytes = 65536
	MaxOutputBytes      = 131072

	PublicProtocolID       = "L7-EVAL-PUBLIC-001"
	PublicProtocolVersion  = "1.0.0"
	SemanticCasesSHA256    = "35fc9f26ccc3b7440a44bf4bdb3e73e10dca53733673bab8dc58e657cc9cab06"
	BrokenCandidatesSHA256 = "8b8330b8a6390ea008dd9bd72f4dd5befe1f6973a98403fa836545c7a21ae56f"
)

type Digest = render.Digest

type ControlFile struct {
	Path string
	Data []byte
}

type Diagnostic struct {
	Rule    string `json:"rule"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Next    string `json:"next"`
}

type InputBindings struct {
	SemanticCases    Digest `json:"semantic_cases"`
	BrokenCandidates Digest `json:"broken_candidates"`
}

type DocumentBinding struct {
	Path   string
	Inputs InputBindings
}

type CandidateSelection struct {
	ExactDigestRequired bool   `json:"exact_digest_required"`
	Precommitted        bool   `json:"precommitted"`
	PostResultChoice    string `json:"post_result_choice"`
}

type CaseSelection struct {
	Mode                        string `json:"mode"`
	RequiredCoverageBasisPoints int64  `json:"required_coverage_basis_points"`
}

type Sampling struct {
	Mode       string `json:"mode"`
	Randomness string `json:"randomness"`
}

type SeedPolicy struct {
	Seed       int64  `json:"seed"`
	Randomness string `json:"randomness"`
}

type HostModelPolicy struct {
	Host                  string `json:"host"`
	Model                 string `json:"model"`
	RequiredInRunManifest bool   `json:"required_in_run_manifest"`
}

type GraderPolicy struct {
	DeterministicAuthoritative bool   `json:"deterministic_authoritative"`
	GraderError                string `json:"grader_error"`
	Ambiguity                  string `json:"ambiguity"`
	ModelJudge                 string `json:"model_judge"`
}

type ConfidencePolicy struct {
	BasisPoints      int64  `json:"basis_points"`
	Condition        string `json:"condition"`
	StatisticalClaim bool   `json:"statistical_claim"`
}

type ResourceLimits struct {
	MaxCases         int64 `json:"max_cases"`
	MaxTrialsPerCase int64 `json:"max_trials_per_case"`
	MaxLatencyMillis int64 `json:"max_latency_millis"`
	MaxOutputBytes   int64 `json:"max_output_bytes"`
}

type CostLatencyPolicy struct {
	CostMicroUSD  int64  `json:"cost_micro_usd"`
	LatencyMillis int64  `json:"latency_millis"`
	Measurement   string `json:"measurement"`
}

type FailureThresholds struct {
	AllCasesRequired                    bool   `json:"all_cases_required"`
	SafetyAveraging                     string `json:"safety_averaging"`
	UnauthorizedEffects                 int64  `json:"unauthorized_effects"`
	CanaryOccurrences                   int64  `json:"canary_occurrences"`
	BrokenCandidateRejectionBasisPoints int64  `json:"broken_candidate_rejection_basis_points"`
	CoverageBasisPoints                 int64  `json:"coverage_basis_points"`
}

type ProtectedHoldout struct {
	MinimumCorpusBasisPoints int64    `json:"minimum_corpus_basis_points"`
	ExcludedScopes           []string `json:"excluded_scopes"`
	OperatorAuthority        string   `json:"operator_authority"`
	Sampling                 string   `json:"sampling"`
	Workspace                string   `json:"workspace"`
	ResourceBoundary         string   `json:"resource_boundary"`
	ExternalControls         []string `json:"external_controls"`
	Feedback                 string   `json:"feedback"`
	TamperResponse           string   `json:"tamper_response"`
	Rotation                 string   `json:"rotation"`
	Invalidation             string   `json:"invalidation"`
	SubmissionControls       string   `json:"submission_controls"`
	HumanExposureResponse    string   `json:"human_exposure_response"`
	ImplementationState      string   `json:"implementation_state"`
	EvaluationState          string   `json:"evaluation_state"`
}

type Protocol struct {
	render.RecordMeta
	CandidateSelection CandidateSelection `json:"candidate_selection"`
	CaseSelection      CaseSelection      `json:"case_selection"`
	Ordering           string             `json:"ordering"`
	RunCount           int64              `json:"run_count"`
	Sampling           Sampling           `json:"sampling"`
	SeedPolicy         SeedPolicy         `json:"seed_policy"`
	HostModelPolicy    HostModelPolicy    `json:"host_model_policy"`
	GraderPolicy       GraderPolicy       `json:"grader_policy"`
	AdjudicationID     string             `json:"adjudication_id"`
	Confidence         ConfidencePolicy   `json:"confidence"`
	ResourceLimits     ResourceLimits     `json:"resource_limits"`
	CostLatency        CostLatencyPolicy  `json:"cost_latency"`
	FailureThresholds  FailureThresholds  `json:"failure_thresholds"`
	ControlSetID       string             `json:"control_set_id"`
	ControlSetVersion  string             `json:"control_set_version"`
	ControlPaths       []string           `json:"control_paths"`
	Invalidation       string             `json:"invalidation"`
	ProtectedHoldout   ProtectedHoldout   `json:"protected_holdout"`
}

type CaseAxes struct {
	Scenario string `json:"scenario"`
	Risk     string `json:"risk"`
	Effect   string `json:"effect"`
	Profile  string `json:"profile"`
}

type Case struct {
	render.RecordMeta
	FeatureOwner         string         `json:"feature_owner"`
	Axes                 CaseAxes       `json:"axes"`
	InputFixture         string         `json:"input_fixture"`
	InputDigest          string         `json:"input_digest"`
	AllowedCapabilities  []string       `json:"allowed_capabilities"`
	AllowedTools         []string       `json:"allowed_tools"`
	AllowedEffects       []string       `json:"allowed_effects"`
	ProhibitedEffects    []string       `json:"prohibited_effects"`
	ExpectedOutputSchema string         `json:"expected_output_schema"`
	TruthIDs             []string       `json:"truth_ids"`
	GraderIDs            []string       `json:"grader_ids"`
	ResourceLimits       ResourceLimits `json:"resource_limits"`
	Isolation            string         `json:"isolation"`
	Sensitivity          string         `json:"sensitivity"`
	Setup                string         `json:"setup"`
	Teardown             string         `json:"teardown"`
}

type TruthLabel struct {
	render.RecordMeta
	CaseID            string   `json:"case_id"`
	ProtocolID        string   `json:"protocol_id"`
	ExpectedDecision  string   `json:"expected_decision"`
	ExpectedRuleIDs   []string `json:"expected_rule_ids"`
	ExpectedEvidence  string   `json:"expected_evidence"`
	Authority         string   `json:"authority"`
	Rationale         string   `json:"rationale"`
	AdjudicationState string   `json:"adjudication_state"`
	Exposure          string   `json:"exposure"`
}

type GraderBounds struct {
	MaxCases       int64 `json:"max_cases"`
	MaxTrials      int64 `json:"max_trials"`
	MaxOutputBytes int64 `json:"max_output_bytes"`
	MaxDiagnostics int64 `json:"max_diagnostics"`
}

type Calibration struct {
	OrderReversal         bool     `json:"order_reversal"`
	VerbosityMatchedPairs bool     `json:"verbosity_matched_pairs"`
	CrossFamilySets       []string `json:"cross_family_sets"`
	ExecutionState        string   `json:"execution_state"`
}

type Grader struct {
	render.RecordMeta
	Class           string       `json:"class"`
	InputSchema     string       `json:"input_schema"`
	OutputSchema    string       `json:"output_schema"`
	ObligationIDs   []string     `json:"obligation_ids"`
	TruthIDs        []string     `json:"truth_ids"`
	ResultSemantics string       `json:"result_semantics"`
	Bounds          GraderBounds `json:"bounds"`
	ErrorBehavior   string       `json:"error_behavior"`
	Calibration     Calibration  `json:"calibration"`
	Adjudication    string       `json:"adjudication"`
	AuthorityLimit  string       `json:"authority_limit"`
}

type CoverageAxis struct {
	Axis          string   `json:"axis"`
	RequirementID string   `json:"requirement_id"`
	ObligationID  string   `json:"obligation_id"`
	Feature       string   `json:"feature"`
	CaseIDs       []string `json:"case_ids"`
	TruthIDs      []string `json:"truth_ids"`
	GraderIDs     []string `json:"grader_ids"`
}

type Coverage struct {
	render.RecordMeta
	RequirementIDs []string       `json:"requirement_ids"`
	ObligationIDs  []string       `json:"obligation_ids"`
	Axes           []CoverageAxis `json:"axes"`
}

type Adjudication struct {
	render.RecordMeta
	Trigger            []string `json:"trigger"`
	EligibleRole       string   `json:"eligible_role"`
	Inputs             []string `json:"inputs"`
	DecisionValues     []string `json:"decision_values"`
	AmbiguityResult    string   `json:"ambiguity_result"`
	ConflictResult     string   `json:"conflict_result"`
	CandidateExclusion bool     `json:"candidate_exclusion"`
	Record             string   `json:"record"`
	Invalidation       string   `json:"invalidation"`
}

type RunTrial struct {
	CaseID       string `json:"case_id"`
	Trial        int64  `json:"trial"`
	Seed         int64  `json:"seed"`
	DurationMS   int64  `json:"duration_millis"`
	CostMicroUSD int64  `json:"cost_micro_usd"`
	Decision     string `json:"decision"`
	OutputSHA256 string `json:"output_sha256"`
}

type RunManifest struct {
	render.RecordMeta
	Candidate     Digest     `json:"candidate"`
	Semantic      string     `json:"semantic"`
	Workflow      string     `json:"workflow"`
	Profiles      []string   `json:"profiles"`
	Prompt        string     `json:"prompt"`
	Protocol      string     `json:"protocol"`
	Graders       []string   `json:"graders"`
	Host          string     `json:"host"`
	Model         string     `json:"model"`
	Harness       string     `json:"harness"`
	Tools         []string   `json:"tools"`
	Environment   string     `json:"environment"`
	CaseSelection []string   `json:"case_selection"`
	Trials        []RunTrial `json:"trials"`
	Resources     string     `json:"resources"`
	Cost          int64      `json:"cost"`
	Latency       int64      `json:"latency"`
	Termination   string     `json:"termination"`
	Effects       []string   `json:"effects"`
	Results       []string   `json:"results"`
	Producer      string     `json:"producer"`
	Authority     string     `json:"authority"`
	Adjudication  string     `json:"adjudication"`
	Uncertainty   string     `json:"uncertainty"`
	Invalidation  string     `json:"invalidation"`
	Limitations   string     `json:"limitations"`
}

type Controls struct {
	Protocol              Protocol
	Cases                 []Case
	TruthLabels           []TruthLabel
	Graders               []Grader
	Coverage              Coverage
	Adjudication          Adjudication
	Descriptors           []render.SchemaDescriptor
	Bindings              []DocumentBinding
	SourceDigests         []Digest
	SourceSetSHA256       string
	typedSHA256           string
	controlManifestSHA256 string
}

type Effect struct {
	Name       string `json:"name"`
	Authorized bool   `json:"authorized"`
}

type Trial struct {
	CaseID                string   `json:"case_id"`
	Trial                 int64    `json:"trial"`
	Seed                  int64    `json:"seed"`
	Host                  string   `json:"host"`
	Model                 string   `json:"model"`
	Decision              string   `json:"decision"`
	RuleIDs               []string `json:"rule_ids"`
	Output                string   `json:"output"`
	OutputSHA256          string   `json:"output_sha256"`
	CompilationSHA256     string   `json:"compilation_sha256"`
	AccountingSHA256      string   `json:"accounting_sha256"`
	PeerAccountingSHA256  string   `json:"peer_accounting_sha256"`
	EvidenceState         string   `json:"evidence_state"`
	ApprovalState         string   `json:"approval_state"`
	RequiredRisk          string   `json:"required_risk"`
	RoutedRisk            string   `json:"routed_risk"`
	EffectCeiling         string   `json:"effect_ceiling"`
	ObservedEffects       []string `json:"observed_effects"`
	CanaryOccurrences     int64    `json:"canary_occurrences"`
	SubagentUsed          bool     `json:"subagent_used"`
	SingleAgentEquivalent bool     `json:"single_agent_equivalent"`
	CostMicroUSD          int64    `json:"cost_micro_usd"`
	LatencyMillis         int64    `json:"latency_millis"`
	GraderError           string   `json:"grader_error"`
	Ambiguous             bool     `json:"ambiguous"`
}

type GradeRequest struct {
	CandidateID     Digest
	Compilation     render.Compilation
	Controls        Controls
	Trials          []Trial
	ObservedEffects []Effect
}

type CaseResult struct {
	CaseID                string   `json:"case_id"`
	Decision              string   `json:"decision"`
	RuleIDs               []string `json:"rule_ids"`
	TrialOutcomeSHA256    string   `json:"trial_outcome_sha256"`
	ConfidenceBasisPoints int64    `json:"confidence_basis_points"`
}

type CoverageResult struct {
	Requirements int64 `json:"requirements"`
	Obligations  int64 `json:"obligations"`
	Axes         int64 `json:"axes"`
	Complete     bool  `json:"complete"`
}

type GradeResult struct {
	Decision              string         `json:"decision"`
	RuleIDs               []string       `json:"rule_ids"`
	CaseResults           []CaseResult   `json:"case_results"`
	Coverage              CoverageResult `json:"coverage"`
	CostMicroUSD          int64          `json:"cost_micro_usd"`
	LatencyMillis         int64          `json:"latency_millis"`
	ConfidenceBasisPoints int64          `json:"confidence_basis_points"`
	Limitations           []string       `json:"limitations"`
	ResultSHA256          string         `json:"result_sha256"`
}

type protocolDocument struct {
	SchemaVersion string        `json:"schema_version"`
	InputBindings InputBindings `json:"input_bindings"`
	Protocols     []Protocol    `json:"protocols"`
}

type caseDocument struct {
	SchemaVersion string        `json:"schema_version"`
	InputBindings InputBindings `json:"input_bindings"`
	Cases         []Case        `json:"cases"`
}

type truthDocument struct {
	SchemaVersion string        `json:"schema_version"`
	InputBindings InputBindings `json:"input_bindings"`
	TruthLabels   []TruthLabel  `json:"truth_labels"`
}

type graderDocument struct {
	SchemaVersion string        `json:"schema_version"`
	InputBindings InputBindings `json:"input_bindings"`
	Graders       []Grader      `json:"graders"`
}

type coverageDocument struct {
	SchemaVersion string        `json:"schema_version"`
	InputBindings InputBindings `json:"input_bindings"`
	Coverage      []Coverage    `json:"coverage"`
}

type adjudicationDocument struct {
	SchemaVersion string         `json:"schema_version"`
	InputBindings InputBindings  `json:"input_bindings"`
	Adjudications []Adjudication `json:"adjudications"`
}
