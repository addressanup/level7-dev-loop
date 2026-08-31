// Package app implements Level 7 CLI use cases against the pure domain model.
package app

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const developmentVersion = "1.0.0-dev"

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
		result := application.result(domain.OutcomePass, "L7-CLI-000", string(request.Command), "available", "Level 7 multi-host orchestration development candidate", "run l7 onboard --status")
		result.Details = []string{
			"Usage: l7 <command> [options] [--json]",
			"Orchestration: onboard, providers, route, sync, cyber, headless, mcp",
			"Lifecycle compatibility: adopt, brief, run, verify, review, ready, merge, status",
			"All v1 features remain default OFF until l7 onboard --apply; Headless requires a separately confirmed manifest digest.",
		}
		return result
	case domain.CommandVersion:
		return application.result(domain.OutcomePass, "L7-CLI-000", string(request.Command), "available", "Level 7 multi-host orchestration development candidate", "run l7 onboard --status")
	case domain.CommandAdopt:
		return application.adopt(ctx, request)
	case domain.CommandBrief:
		return application.createBrief(ctx, request)
	case domain.CommandStatus:
		return application.status(ctx, request)
	case domain.CommandRun:
		return application.runChange(ctx, request)
	case domain.CommandVerify:
		return application.verifyChange(ctx, request)
	case domain.CommandReview:
		return application.reviewChange(ctx, request)
	case domain.CommandReady:
		return application.readyChange(ctx, request)
	case domain.CommandMerge:
		return application.mergeChange(ctx, request)
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
