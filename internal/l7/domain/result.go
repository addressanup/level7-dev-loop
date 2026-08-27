// Package domain contains the side-effect-free Level 7 CLI model.
package domain

const ResultSchema = 1

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
	CommandStatus  Command = "status"
)

type Result struct {
	Schema  int
	Outcome Outcome
	Code    string
	Command string
	State   string
	Version string
	Message string
	Next    string
	Details []string
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
