package domain

const EvidenceSchema = 1

type Provider string

const (
	ProviderCodex  Provider = "codex"
	ProviderClaude Provider = "claude"
)

func (provider Provider) Valid() bool {
	return provider == ProviderCodex || provider == ProviderClaude
}

func OtherProvider(provider Provider) (Provider, bool) {
	switch provider {
	case ProviderCodex:
		return ProviderClaude, true
	case ProviderClaude:
		return ProviderCodex, true
	default:
		return "", false
	}
}

type ProviderRole string

const (
	RoleImplementer ProviderRole = "implementer"
	RoleReviewer    ProviderRole = "reviewer"
)

func (role ProviderRole) Valid() bool {
	return role == RoleImplementer || role == RoleReviewer
}

type CapabilityState string

const (
	CapabilityAvailable   CapabilityState = "available"
	CapabilityUnavailable CapabilityState = "unavailable"
	CapabilityDegraded    CapabilityState = "degraded"
)

type ProviderIdentity struct {
	Provider   Provider
	Executable string
	Version    string
	Digest     string
	Capability CapabilityState
}

type CandidateIdentity struct {
	Commit string
	Tree   string
}

type PendingChanges struct {
	RepositoryLocation
	Paths      []string
	IndexDirty bool
}

type CommitRequest struct {
	Root              string
	ExpectedCommit    string
	ExpectedTree      string
	Paths             []string
	Message           string
	MaxOutputBytes    int
	MaxPaths          int
	MaxCommandSeconds int
}

type ProviderTask struct {
	Role               ProviderRole
	Provider           Provider
	RepositoryRoot     string
	ChangeID           string
	Tier               RiskTier
	Base               string
	Candidate          CandidateIdentity
	Problem            string
	Scope              []string
	AcceptanceCriteria []string
	Risks              []string
	Rollback           []string
}

type ReviewDecision string

const (
	DecisionGO   ReviewDecision = "GO"
	DecisionNoGO ReviewDecision = "NO_GO"
)

func (decision ReviewDecision) Valid() bool {
	return decision == DecisionGO || decision == DecisionNoGO
}

type ProviderResponse struct {
	Identity ProviderIdentity
	Role     ProviderRole
	Summary  string
	Findings []string
	Decision ReviewDecision
}

type VerificationCommand struct {
	Name      string
	Argv      []string
	Benchmark bool
}

type CheckResult struct {
	Name      string
	Benchmark bool
	Passed    bool
	ExitCode  int
	Code      string
	Message   string
}

type ApprovalBinding struct {
	ChangeID    string
	Actor       string
	Implementer Provider
	BriefCommit string
}

type RunEvidence struct {
	ChangeID      string
	Provider      ProviderIdentity
	Parent        CandidateIdentity
	Candidate     CandidateIdentity
	PathDigest    string
	PathCount     int
	CommitMessage string
}

type VerificationEvidence struct {
	ChangeID           string
	Candidate          CandidateIdentity
	Result             ReviewDecision
	Checks             []CheckResult
	VerificationCommit string
	VerificationTree   string
}

type ReviewEvidence struct {
	ChangeID     string
	Provider     ProviderIdentity
	Candidate    CandidateIdentity
	Decision     ReviewDecision
	Findings     []string
	ReviewCommit string
	ReviewTree   string
}

func ConventionalSubject(value string) bool {
	if len(value) < 8 || len(value) > 160 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	colon := -1
	for index, character := range []byte(value) {
		if character == '\n' || character == '\r' || character == 0 || character == 0x7f || character < 0x20 {
			return false
		}
		if character == ':' && colon == -1 {
			colon = index
		}
	}
	if colon < 3 || colon+2 >= len(value) || value[colon+1] != ' ' {
		return false
	}
	for _, character := range []byte(value[:colon]) {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '(' && character != ')' && character != '-' && character != '_' && character != '!' {
			return false
		}
	}
	return value[colon+2] != ' '
}

func DistinctReviewer(tier RiskTier, implementer, reviewer ProviderIdentity) bool {
	if !implementer.Provider.Valid() || !reviewer.Provider.Valid() {
		return false
	}
	if tier == TierHighRisk {
		return implementer.Provider != reviewer.Provider
	}
	return true
}
