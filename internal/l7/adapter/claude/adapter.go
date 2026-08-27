// Package claude translates the provider-neutral contract to Claude Code.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const CompatibleVersion = "2.1.241"

const terminalSchema = `{"type":"object","additionalProperties":false,"required":["schema","outcome","summary","findings"],"properties":{"schema":{"const":1},"outcome":{"enum":["complete","blocked"]},"summary":{"type":"string"},"findings":{"type":"array","items":{"type":"string"}},"decision":{"enum":["GO","NO_GO"]}}}`

type Adapter struct {
	runtime provider.Runtime
}

func New() Adapter {
	return NewWithRuntime(provider.NewRuntime(nil, nil))
}

func NewWithRuntime(runtime provider.Runtime) Adapter {
	return Adapter{runtime: runtime}
}

func (adapter Adapter) Probe(ctx context.Context) (domain.ProviderIdentity, error) {
	return adapter.runtime.Probe(ctx, "claude", domain.ProviderClaude, []string{"--version"}, func(version string) bool {
		return version == CompatibleVersion || version == CompatibleVersion+" (Claude Code)"
	})
}

func (adapter Adapter) Run(ctx context.Context, task domain.ProviderTask, maxOutputBytes, maxSeconds int) (domain.ProviderResponse, error) {
	if task.Provider != domain.ProviderClaude {
		return domain.ProviderResponse{}, errors.New("Claude adapter received a task for another provider")
	}
	identity, err := adapter.Probe(ctx)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	if identity.Capability != domain.CapabilityAvailable {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, fmt.Errorf("Claude %q is installed but has no qualified adapter contract", identity.Version)
	}
	prompt, err := provider.RenderTask(task)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	result, err := adapter.runtime.Invoke(ctx, identity, task.RepositoryRoot, arguments(task.Role), prompt, maxOutputBytes, maxSeconds)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	response, err := parseResult(result.Stdout, task.Role)
	response.Identity = identity
	return response, err
}

func arguments(role domain.ProviderRole) []string {
	permission := "acceptEdits"
	tools := "Read,Glob,Grep,Edit,Write,Bash"
	if role == domain.RoleReviewer {
		permission = "plan"
		tools = "Read,Glob,Grep,Bash"
	}
	return []string{
		"--bare",
		"--disable-slash-commands",
		"--print",
		"--input-format", "text",
		"--max-turns", "64",
		"--tools", tools,
		"--disallowedTools", "WebFetch,WebSearch,NotebookEdit,Task,Skill",
		"--permission-mode", permission,
		"--strict-mcp-config",
		"--no-chrome",
		"--no-session-persistence",
		"--output-format", "json",
		"--json-schema", terminalSchema,
	}
}

func parseResult(output []byte, role domain.ProviderRole) (domain.ProviderResponse, error) {
	if !role.Valid() || len(output) == 0 || len(output) > provider.MaxProviderPrompt*64 || !utf8.Valid(output) {
		return domain.ProviderResponse{}, errors.New("Claude result envelope is invalid")
	}
	output = bytes.TrimSuffix(output, []byte("\n"))
	if len(output) == 0 || strings.TrimSpace(string(output)) != string(output) {
		return domain.ProviderResponse{}, errors.New("Claude result framing is invalid")
	}
	var envelope map[string]any
	if err := localfile.DecodeJSON(output, &envelope); err != nil {
		return domain.ProviderResponse{}, fmt.Errorf("decode Claude result: %w", err)
	}
	if !onlyKeys(envelope, "type", "subtype", "is_error", "duration_ms", "duration_api_ms", "num_turns", "result", "session_id", "total_cost_usd", "usage", "modelUsage", "permission_denials", "structured_output") {
		return domain.ProviderResponse{}, errors.New("Claude result contains an unknown field")
	}
	resultType, typeOK := envelope["type"].(string)
	subtype, subtypeOK := envelope["subtype"].(string)
	isError, errorOK := envelope["is_error"].(bool)
	structured, structuredOK := envelope["structured_output"].(map[string]any)
	if !typeOK || resultType != "result" || !subtypeOK || subtype != "success" || !errorOK || isError || !structuredOK {
		return domain.ProviderResponse{}, errors.New("Claude did not return one successful structured result")
	}
	terminal, err := json.Marshal(structured)
	if err != nil {
		return domain.ProviderResponse{}, errors.New("Claude structured result cannot be encoded")
	}
	return provider.ParseTerminal(terminal, role)
}

func onlyKeys(value map[string]any, allowed ...string) bool {
	set := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		set[key] = true
	}
	for key := range value {
		if !set[key] {
			return false
		}
	}
	return true
}
