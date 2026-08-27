// Package domain contains the side-effect-free Level 7 CLI model.
package domain

const ResultSchema = 3

type Outcome string

const (
	OutcomePass      Outcome = "PASS"
	OutcomeBlocked   Outcome = "BLOCKED"
	OutcomeFailed    Outcome = "FAILED"
	OutcomeCancelled Outcome = "CANCELLED"
)

type Command string

const (
	CommandHelp    Command = "help"
	CommandVersion Command = "version"
	CommandAdopt   Command = "adopt"
	CommandBrief   Command = "brief"
	CommandStatus  Command = "status"
	CommandRun     Command = "run"
	CommandVerify  Command = "verify"
	CommandReview  Command = "review"
)

type Request struct {
	Command              Command
	EnableLocalLifecycle bool
	ChangeID             string
	Tier                 RiskTier
	Problem              string
	Scope                []string
	AcceptanceCriteria   []string
	Risks                []string
	Rollback             []string
	Agent                Provider
	CommitMessage        string
}

type RepositoryDetails struct {
	Root          string
	CommonDir     string
	ChangeID      string
	Tier          RiskTier
	Base          string
	Head          string
	Tree          string
	DeclaredScope []string
	ChangedPaths  []string
	ExpandedPaths []string
}

type Result struct {
	Schema     int
	Outcome    Outcome
	Code       string
	Command    string
	State      string
	Version    string
	Message    string
	Next       string
	Details    []string
	Repository *RepositoryDetails
	Execution  *ExecutionDetails
}

type ExecutionDetails struct {
	Role       ProviderRole
	Provider   Provider
	Executable string
	Version    string
	Digest     string
	Commit     string
	Tree       string
	Decision   ReviewDecision
	Checks     []CheckResult
}

func (result Result) ExitCode() int {
	switch result.Outcome {
	case OutcomePass:
		return 0
	case OutcomeBlocked:
		return 2
	case OutcomeCancelled:
		return 130
	default:
		return 1
	}
}
