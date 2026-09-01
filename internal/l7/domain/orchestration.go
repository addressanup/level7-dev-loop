package domain

// OrchestrationSchema versions public orchestration records independently from
// the legacy local-lifecycle result schema.
const OrchestrationSchema = 1

type ProviderKind string

const (
	ProviderKindCodexAppServer  ProviderKind = "codex_app_server"
	ProviderKindClaudeCLI       ProviderKind = "claude_cli"
	ProviderKindOpenAIResponses ProviderKind = "openai_responses"
	ProviderKindAnthropic       ProviderKind = "anthropic_messages"
)

func (kind ProviderKind) Valid() bool {
	return kind == ProviderKindCodexAppServer || kind == ProviderKindClaudeCLI ||
		kind == ProviderKindOpenAIResponses || kind == ProviderKindAnthropic
}

type AuthenticationState string

const (
	AuthAuthenticated   AuthenticationState = "authenticated"
	AuthUnavailable     AuthenticationState = "unavailable"
	AuthUnauthenticated AuthenticationState = "unauthenticated"
	AuthUnknown         AuthenticationState = "unknown"
)

func (state AuthenticationState) Valid() bool {
	return state == AuthAuthenticated || state == AuthUnavailable ||
		state == AuthUnauthenticated || state == AuthUnknown
}

type ReasoningEffort string

const (
	EffortNone    ReasoningEffort = "none"
	EffortMinimal ReasoningEffort = "minimal"
	EffortLow     ReasoningEffort = "low"
	EffortMedium  ReasoningEffort = "medium"
	EffortHigh    ReasoningEffort = "high"
	EffortXHigh   ReasoningEffort = "xhigh"
	EffortMax     ReasoningEffort = "max"
	EffortUltra   ReasoningEffort = "ultra"
)

func (effort ReasoningEffort) Valid() bool {
	switch effort {
	case EffortNone, EffortMinimal, EffortLow, EffortMedium, EffortHigh, EffortXHigh, EffortMax, EffortUltra:
		return true
	default:
		return false
	}
}

type Complexity string

const (
	ComplexityC1 Complexity = "C1"
	ComplexityC2 Complexity = "C2"
	ComplexityC3 Complexity = "C3"
	ComplexityC4 Complexity = "C4"
)

func (complexity Complexity) Valid() bool {
	return complexity == ComplexityC1 || complexity == ComplexityC2 || complexity == ComplexityC3 || complexity == ComplexityC4
}

type ModelCapability struct {
	ID              string            `json:"id"`
	DisplayName     string            `json:"display_name"`
	Languages       []string          `json:"languages"`
	ContextWindow   int               `json:"context_window"`
	SupportsTools   bool              `json:"supports_tools"`
	SupportsEditing bool              `json:"supports_editing"`
	SupportsResume  bool              `json:"supports_resume"`
	Efforts         []ReasoningEffort `json:"efforts"`
	CostClass       int               `json:"cost_class"`
	LatencyClass    int               `json:"latency_class"`
	Verified        bool              `json:"verified"`
}

type QuotaState struct {
	Limited    bool   `json:"limited"`
	ResetAtUTC string `json:"reset_at_utc"`
	Source     string `json:"source"`
}

type ProviderSnapshot struct {
	Schema           int                 `json:"schema"`
	ID               string              `json:"id"`
	Kind             ProviderKind        `json:"kind"`
	Executable       string              `json:"executable"`
	Version          string              `json:"version"`
	ExecutableDigest string              `json:"executable_digest"`
	Authentication   AuthenticationState `json:"authentication"`
	CatalogComplete  bool                `json:"catalog_complete"`
	Models           []ModelCapability   `json:"models"`
	Quota            QuotaState          `json:"quota"`
	ObservedAtUTC    string              `json:"observed_at_utc"`
	Diagnostic       string              `json:"diagnostic"`
	Next             string              `json:"next"`
}

type TaskProfile struct {
	Schema              int        `json:"schema"`
	ID                  string     `json:"id"`
	Summary             string     `json:"summary"`
	Complexity          Complexity `json:"complexity"`
	RiskTier            RiskTier   `json:"risk_tier"`
	ContextTokens       int        `json:"context_tokens"`
	NeedsTools          bool       `json:"needs_tools"`
	NeedsEditing        bool       `json:"needs_editing"`
	NeedsResume         bool       `json:"needs_resume"`
	IndependentReview   bool       `json:"independent_review"`
	ImplementerProvider string     `json:"implementer_provider"`
	ImplementerModel    string     `json:"implementer_model"`
	Languages           []string   `json:"languages"`
	WorkKinds           []string   `json:"work_kinds"`
	PriorFailures       int        `json:"prior_failures"`
}

type RouteCandidate struct {
	ProviderID string          `json:"provider_id"`
	ModelID    string          `json:"model_id"`
	Qualified  bool            `json:"qualified"`
	Score      int             `json:"score"`
	Effort     ReasoningEffort `json:"effort"`
	Reasons    []string        `json:"reasons"`
}

type RouteDecision struct {
	Schema      int              `json:"schema"`
	TaskID      string           `json:"task_id"`
	ProviderID  string           `json:"provider_id"`
	ModelID     string           `json:"model_id"`
	Effort      ReasoningEffort  `json:"effort"`
	Policy      string           `json:"policy"`
	Candidates  []RouteCandidate `json:"candidates"`
	Fallbacks   []string         `json:"fallbacks"`
	Escalations []string         `json:"escalations"`
	DecisionUTC string           `json:"decision_utc"`
	Next        string           `json:"next"`
}

type MemoryNode struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	Language string `json:"language"`
	Digest   string `json:"digest"`
	Summary  string `json:"summary"`
	Line     int    `json:"line"`
}

type MemoryEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type MemoryGraph struct {
	Schema             int          `json:"schema"`
	RepositoryHead     string       `json:"repository_head"`
	EmbeddingProvider  string       `json:"embedding_provider"`
	EmbeddingRevision  int          `json:"embedding_revision"`
	EmbeddingDimension int          `json:"embedding_dimension"`
	Nodes              []MemoryNode `json:"nodes"`
	Edges              []MemoryEdge `json:"edges"`
	UpdatedAtUTC       string       `json:"updated_at_utc"`
	Next               string       `json:"next"`
}

type MemoryMatch struct {
	Node  MemoryNode `json:"node"`
	Score int        `json:"score"`
	Why   []string   `json:"why"`
}

type SecurityFinding struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Severity         string   `json:"severity"`
	CWE              string   `json:"cwe"`
	CVSS             string   `json:"cvss"`
	Confidence       string   `json:"confidence"`
	Exploitability   string   `json:"exploitability"`
	Path             string   `json:"path"`
	Line             int      `json:"line"`
	EvidenceDigest   string   `json:"evidence_digest"`
	Reproduction     []string `json:"reproduction"`
	Remediation      string   `json:"remediation"`
	VerificationTest string   `json:"verification_test"`
}

type SecurityReport struct {
	Schema         int               `json:"schema"`
	ID             string            `json:"id"`
	RepositoryHead string            `json:"repository_head"`
	Mode           string            `json:"mode"`
	Isolation      string            `json:"isolation"`
	Coverage       []string          `json:"coverage"`
	Findings       []SecurityFinding `json:"findings"`
	CreatedAtUTC   string            `json:"created_at_utc"`
	Redacted       bool              `json:"redacted"`
	Next           string            `json:"next"`
}

type HeadlessWave struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Scope              []string   `json:"scope"`
	AcceptanceCriteria []string   `json:"acceptance_criteria"`
	Verification       [][]string `json:"verification"`
}

type HeadlessManifest struct {
	Schema           int            `json:"schema"`
	ID               string         `json:"id"`
	ObjectivePath    string         `json:"objective_path"`
	ObjectiveDigest  string         `json:"objective_digest"`
	BaseCommit       string         `json:"base_commit"`
	TargetBranch     string         `json:"target_branch"`
	RiskCeiling      RiskTier       `json:"risk_ceiling"`
	AllowedPaths     []string       `json:"allowed_paths"`
	AllowedCommands  [][]string     `json:"allowed_commands"`
	ProviderPolicy   string         `json:"provider_policy"`
	NetworkPolicy    string         `json:"network_policy"`
	LocalMerge       bool           `json:"local_merge"`
	StopBeforeDeploy bool           `json:"stop_before_deploy"`
	Waves            []HeadlessWave `json:"waves"`
	Digest           string         `json:"digest"`
	CreatedAtUTC     string         `json:"created_at_utc"`
	Next             string         `json:"next"`
}

type HeadlessCheckpoint struct {
	Schema           int    `json:"schema"`
	RunID            string `json:"run_id"`
	ManifestDigest   string `json:"manifest_digest"`
	Sequence         int    `json:"sequence"`
	State            string `json:"state"`
	WaveID           string `json:"wave_id"`
	ProviderID       string `json:"provider_id"`
	ModelID          string `json:"model_id"`
	SessionID        string `json:"session_id"`
	Worktree         string `json:"worktree"`
	CandidateCommit  string `json:"candidate_commit"`
	FailureSignature string `json:"failure_signature"`
	RepeatedFailures int    `json:"repeated_failures"`
	QuotaResetAtUTC  string `json:"quota_reset_at_utc"`
	UpdatedAtUTC     string `json:"updated_at_utc"`
	Next             string `json:"next"`
}

type OrchestrationDetails struct {
	Providers  []ProviderSnapshot  `json:"providers"`
	Route      *RouteDecision      `json:"route"`
	Memory     *MemoryGraph        `json:"memory"`
	Matches    []MemoryMatch       `json:"matches"`
	Security   *SecurityReport     `json:"security"`
	Manifest   *HeadlessManifest   `json:"manifest"`
	Checkpoint *HeadlessCheckpoint `json:"checkpoint"`
}
