// Package app implements Level 7 CLI use cases against the pure domain model.
package app

import (
	"context"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const developmentVersion = "0.1.0-dev"

type Application struct {
	version string
}

func New(version string) Application {
	if strings.TrimSpace(version) == "" {
		version = developmentVersion
	}
	return Application{version: version}
}

func (application Application) Execute(ctx context.Context, command domain.Command) domain.Result {
	if err := ctx.Err(); err != nil {
		return application.result(domain.OutcomeCancelled, "L7-CLI-003", string(command), "cancelled", "command cancelled before execution", "run l7 help")
	}

	switch command {
	case domain.CommandHelp:
		result := application.result(domain.OutcomePass, "L7-CLI-000", string(command), "available", "Level 7 CLI proving shell", "run l7 status")
		result.Details = []string{
			"Usage: l7 <command> [--json]",
			"Commands: help, version, status",
		}
		return result
	case domain.CommandVersion:
		return application.result(domain.OutcomePass, "L7-CLI-000", string(command), "available", "Level 7 CLI proving shell", "run l7 status")
	case domain.CommandStatus:
		return application.result(domain.OutcomeBlocked, "L7-STATUS-001", string(command), "unavailable", "repository workflow status is not implemented in Wave 1", "run l7 help")
	default:
		return application.Invalid(string(command), "unknown command")
	}
}

func (application Application) Invalid(command, message string) domain.Result {
	return application.result(domain.OutcomeFailed, "L7-CLI-001", command, "invalid", message, "run l7 help")
}

func (application Application) result(outcome domain.Outcome, code, command, state, message, next string) domain.Result {
	return domain.Result{
		Schema:  domain.ResultSchema,
		Outcome: outcome,
		Code:    code,
		Command: command,
		State:   state,
		Version: application.version,
		Message: message,
		Next:    next,
	}
}
