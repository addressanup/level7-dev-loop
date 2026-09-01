// Package codex translates the provider-neutral contract to Codex CLI.
package codex

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// CompatibleVersion remains as a fixture baseline for upgrade tests. Runtime
// admission is capability-based and deliberately does not pin this value.
const CompatibleVersion = "codex-cli 0.149.1"

const maxEvents = 4096

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
	return adapter.runtime.Probe(ctx, "codex", domain.ProviderCodex, []string{"--version"}, func(version string) bool {
		return strings.HasPrefix(version, "codex-cli ") || strings.HasPrefix(version, "codex ")
	})
}

func (adapter Adapter) Run(ctx context.Context, task domain.ProviderTask, maxOutputBytes, maxSeconds int) (domain.ProviderResponse, error) {
	if task.Provider != domain.ProviderCodex {
		return domain.ProviderResponse{}, errors.New("Codex adapter received a task for another provider")
	}
	identity, err := adapter.Probe(ctx)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	if identity.Capability != domain.CapabilityAvailable {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, fmt.Errorf("Codex %q is installed but has no qualified adapter contract", identity.Version)
	}
	prompt, err := provider.RenderTask(task)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	result, err := adapter.runtime.Invoke(ctx, identity, task.RepositoryRoot, arguments(task.Role, task.RepositoryRoot), prompt, maxOutputBytes, maxSeconds)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	response, err := parseEvents(result.Stdout, task.Role)
	response.Identity = identity
	return response, err
}

func arguments(role domain.ProviderRole, root string) []string {
	sandbox := "workspace-write"
	if role == domain.RoleReviewer {
		sandbox = "read-only"
	}
	return []string{
		"--ask-for-approval", "never",
		"exec",
		"--ephemeral",
		"--sandbox", sandbox,
		"--color", "never",
		"--json",
		"--cd", root,
		"--skip-git-repo-check",
		"-",
	}
}

func parseEvents(output []byte, role domain.ProviderRole) (domain.ProviderResponse, error) {
	if !role.Valid() || len(output) == 0 || len(output) > provider.MaxProviderPrompt*64 || !utf8.Valid(output) {
		return domain.ProviderResponse{}, errors.New("Codex event stream is invalid")
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > maxEvents+1 {
		return domain.ProviderResponse{}, errors.New("Codex event count exceeds the contract")
	}
	var terminal string
	terminalCount := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.TrimSpace(line) != line {
			return domain.ProviderResponse{}, errors.New("Codex event framing is invalid")
		}
		var event map[string]any
		if err := localfile.DecodeJSON([]byte(line), &event); err != nil {
			return domain.ProviderResponse{}, fmt.Errorf("decode Codex event: %w", err)
		}
		if !onlyKeys(event, "type", "thread_id", "item", "usage", "error") {
			return domain.ProviderResponse{}, errors.New("Codex event contains an unknown field")
		}
		eventType, ok := event["type"].(string)
		if !ok || eventType == "" {
			return domain.ProviderResponse{}, errors.New("Codex event has no valid type")
		}
		switch eventType {
		case "thread.started", "turn.started", "item.started", "turn.completed":
		case "item.completed":
			item, ok := event["item"].(map[string]any)
			if !ok {
				return domain.ProviderResponse{}, errors.New("Codex completed item is invalid")
			}
			itemType, ok := item["type"].(string)
			if !ok || itemType == "" {
				return domain.ProviderResponse{}, errors.New("Codex completed item has no valid type")
			}
			if itemType != "agent_message" {
				continue
			}
			if !onlyKeys(item, "id", "type", "text") {
				return domain.ProviderResponse{}, errors.New("Codex terminal item contains an unknown field")
			}
			message, ok := item["text"].(string)
			if !ok {
				return domain.ProviderResponse{}, errors.New("Codex terminal message is invalid")
			}
			terminal = message
			terminalCount++
		case "error", "turn.failed":
			return domain.ProviderResponse{}, errors.New("Codex reported a failed terminal event")
		default:
			return domain.ProviderResponse{}, fmt.Errorf("unsupported Codex event type %q", eventType)
		}
	}
	if terminalCount != 1 {
		return domain.ProviderResponse{}, errors.New("Codex did not emit exactly one terminal agent message")
	}
	return provider.ParseTerminal([]byte(terminal), role)
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
