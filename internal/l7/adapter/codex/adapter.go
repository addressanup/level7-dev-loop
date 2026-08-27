// Package codex translates the provider-neutral contract to Codex CLI.
package codex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const CompatibleVersion = "codex-cli 0.150.1"

const maxEvents = 4096

type Adapter struct {
	runtime   provider.Runtime
	mkdirTemp func(string, string) (string, error)
	removeAll func(string) error
}

func New() Adapter {
	return NewWithRuntime(provider.NewRuntime(nil, nil))
}

func NewWithRuntime(runtime provider.Runtime) Adapter {
	return Adapter{runtime: runtime, mkdirTemp: os.MkdirTemp, removeAll: os.RemoveAll}
}

func (adapter Adapter) Probe(ctx context.Context) (domain.ProviderIdentity, error) {
	return adapter.runtime.Probe(ctx, "codex", domain.ProviderCodex, []string{"--version"}, func(version string) bool {
		return version == CompatibleVersion
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
	schemaPath, cleanupSchema, err := adapter.prepareTerminalSchema(task.Role, task.RepositoryRoot)
	if err != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, err
	}
	result, invokeErr := adapter.runtime.Invoke(ctx, identity, task.RepositoryRoot, arguments(task.Role, task.RepositoryRoot, schemaPath), prompt, maxOutputBytes, maxSeconds)
	cleanupErr := cleanupSchema()
	if invokeErr != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, errors.Join(invokeErr, cleanupErr)
	}
	if cleanupErr != nil {
		return domain.ProviderResponse{Identity: identity, Role: task.Role}, cleanupErr
	}
	response, err := parseEvents(result.Stdout, task.Role)
	response.Identity = identity
	return response, err
}

func (adapter Adapter) prepareTerminalSchema(role domain.ProviderRole, repositoryRoot string) (string, func() error, error) {
	if adapter.mkdirTemp == nil || adapter.removeAll == nil {
		return "", nil, errors.New("Codex terminal schema storage is not configured")
	}
	schema, err := provider.TerminalSchema(role)
	if err != nil {
		return "", nil, err
	}
	repositoryRoot, err = filepath.EvalSymlinks(repositoryRoot)
	if err != nil || !filepath.IsAbs(repositoryRoot) {
		return "", nil, errors.New("Codex repository root is not a physical absolute directory")
	}
	temporaryRoot, tempRootErr := filepath.EvalSymlinks(os.TempDir())
	if tempRootErr != nil || !filepath.IsAbs(temporaryRoot) {
		temporaryRoot = filepath.Dir(repositoryRoot)
	} else if inside, compareErr := pathWithin(repositoryRoot, temporaryRoot); compareErr != nil || inside {
		temporaryRoot = filepath.Dir(repositoryRoot)
	}
	directory, err := adapter.mkdirTemp(temporaryRoot, "l7-codex-schema-")
	if err != nil {
		return "", nil, fmt.Errorf("create private Codex schema directory: %w", err)
	}
	cleanupPath := directory
	cleanup := func() error {
		if err := adapter.removeAll(cleanupPath); err != nil {
			return fmt.Errorf("remove private Codex schema directory: %w", err)
		}
		return nil
	}
	fail := func(cause error) (string, func() error, error) {
		return "", nil, errors.Join(cause, cleanup())
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return fail(fmt.Errorf("set private Codex schema directory mode: %w", err))
	}
	physicalDirectory, err := filepath.EvalSymlinks(directory)
	if err != nil || !filepath.IsAbs(physicalDirectory) {
		return fail(errors.New("Codex schema directory is not a physical absolute directory"))
	}
	cleanupPath = physicalDirectory
	inside, err := pathWithin(repositoryRoot, physicalDirectory)
	if err != nil {
		return fail(fmt.Errorf("compare Codex schema and repository paths: %w", err))
	}
	if inside {
		return fail(errors.New("Codex schema directory must be outside the repository"))
	}
	info, err := os.Stat(physicalDirectory)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		return fail(errors.New("Codex schema directory is not private"))
	}
	path := filepath.Join(physicalDirectory, "terminal-schema.json")
	if err := localfile.AtomicCreate(path, []byte(schema), 0o600); err != nil {
		return fail(fmt.Errorf("create private Codex terminal schema: %w", err))
	}
	info, err = os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return fail(errors.New("Codex terminal schema file is not private"))
	}
	return path, cleanup, nil
}

func pathWithin(root, path string) (bool, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))), nil
}

func arguments(role domain.ProviderRole, root, schemaPath string) []string {
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
		"--output-schema", schemaPath,
		"--cd", root,
		"-",
	}
}

func parseEvents(output []byte, role domain.ProviderRole) (domain.ProviderResponse, error) {
	if !role.Valid() || len(output) == 0 || len(output) > provider.MaxProviderPrompt*64 || !utf8.Valid(output) {
		return domain.ProviderResponse{}, errors.New("Codex event stream is invalid")
	}
	output = []byte(strings.TrimSuffix(string(output), "\n"))
	if len(output) == 0 {
		return domain.ProviderResponse{}, errors.New("Codex event framing is invalid")
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > maxEvents {
		return domain.ProviderResponse{}, errors.New("Codex event count exceeds the contract")
	}
	var terminal string
	terminalCount := 0
	for _, line := range lines {
		if line == "" || strings.TrimSpace(line) != line {
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
