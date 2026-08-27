// Package app implements Level 7 CLI use cases against the pure domain model.
package app

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const developmentVersion = "0.1.0-dev"

const maxCommandBytes = 160

type Application struct {
	version string
	cwd     string
	ports   Ports
}

func New(version string) Application {
	if strings.TrimSpace(version) == "" {
		version = developmentVersion
	}
	return Application{version: version}
}

func NewLifecycle(version, workingDirectory string, ports Ports) Application {
	application := New(version)
	application.cwd = workingDirectory
	application.ports = ports
	return application
}

func (application Application) Execute(ctx context.Context, command domain.Command) domain.Result {
	return application.ExecuteRequest(ctx, domain.Request{Command: command})
}

func (application Application) ExecuteRequest(ctx context.Context, request domain.Request) domain.Result {
	if err := ctx.Err(); err != nil {
		return application.result(domain.OutcomeCancelled, "L7-CLI-003", string(request.Command), "cancelled", "command cancelled before execution", "run l7 help")
	}

	switch request.Command {
	case domain.CommandHelp:
		result := application.result(domain.OutcomePass, "L7-CLI-000", string(request.Command), "available", "Level 7 CLI local lifecycle preview", "run l7 status")
		result.Details = []string{
			"Usage: l7 <command> [options] [--json]",
			"Commands: help, version, adopt, brief, status",
			"Lifecycle behavior remains default OFF until explicitly enabled during adopt.",
		}
		return result
	case domain.CommandVersion:
		return application.result(domain.OutcomePass, "L7-CLI-000", string(request.Command), "available", "Level 7 CLI local lifecycle preview", "run l7 status")
	case domain.CommandAdopt:
		return application.adopt(ctx, request)
	case domain.CommandBrief:
		return application.createBrief(ctx, request)
	case domain.CommandStatus:
		return application.status(ctx, request)
	default:
		return application.Invalid(string(request.Command), "unknown command")
	}
}

func (application Application) Invalid(command, message string) domain.Result {
	return application.result(domain.OutcomeFailed, "L7-CLI-001", bounded(command, maxCommandBytes), "invalid", message, "run l7 help")
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

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit]
}
