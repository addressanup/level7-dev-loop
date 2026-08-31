// Package discovery observes native host and configured gateway capabilities.
package discovery

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/codexapp"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	maxDiscoveryOutput = 4 << 20
	discoveryTimeout   = 15 * time.Second
)

type ResolveFunc func(string) (processadapter.Executable, error)
type RunFunc func(context.Context, processadapter.Request) (processadapter.Result, error)

type Adapter struct {
	resolve ResolveFunc
	run     RunFunc
	codex   func(context.Context, string) ([]byte, error)
	now     func() time.Time
}

func New() Adapter {
	adapter := NewWith(processadapter.Resolve, (processadapter.Runner{}).Run, time.Now)
	adapter.codex = codexapp.Discover
	return adapter
}

func NewWith(resolve ResolveFunc, run RunFunc, now func() time.Time) Adapter {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	if now == nil {
		now = time.Now
	}
	return Adapter{resolve: resolve, run: run, codex: batchCodexDiscovery(run), now: now}
}

func (adapter Adapter) Discover(ctx context.Context, configuration orchestrationconfig.File) []domain.ProviderSnapshot {
	snapshots := make([]domain.ProviderSnapshot, 0, len(configuration.Providers))
	for _, provider := range configuration.Providers {
		if !provider.Enabled {
			continue
		}
		var snapshot domain.ProviderSnapshot
		switch provider.Kind {
		case domain.ProviderKindCodexAppServer:
			snapshot = adapter.discoverCodex(ctx, provider.ID)
		case domain.ProviderKindClaudeCLI:
			snapshot = adapter.discoverClaude(ctx, provider.ID)
		case domain.ProviderKindOpenAIResponses, domain.ProviderKindAnthropic:
			snapshot = gatewaySnapshot(provider)
		default:
			snapshot = unavailable(provider.ID, provider.Kind, "unsupported provider kind")
		}
		snapshot.ObservedAtUTC = adapter.now().UTC().Format(time.RFC3339)
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

// ProbeClaude verifies the deliberately non-exhaustive candidate aliases by
// asking the authenticated, user-installed CLI to perform one bounded,
// read-only turn per candidate. Discovery never calls this method implicitly.
func (adapter Adapter) ProbeClaude(ctx context.Context, snapshot domain.ProviderSnapshot, directory string) domain.ProviderSnapshot {
	if snapshot.Kind != domain.ProviderKindClaudeCLI || snapshot.Authentication != domain.AuthAuthenticated || snapshot.Executable == "" {
		return snapshot
	}
	verified := 0
	for index := range snapshot.Models {
		result, err := adapter.run(ctx, processadapter.Request{
			Executable: snapshot.Executable,
			Arguments: []string{
				"-p", "Reply exactly OK.", "--model", snapshot.Models[index].ID,
				"--safe-mode", "--disable-slash-commands", "--no-chrome",
				"--output-format", "json", "--max-turns", "1", "--permission-mode", "plan", "--tools", "",
			},
			Directory: directory, Environment: processadapter.MinimalEnvironment(),
			MaxOutputBytes: 256 << 10, Timeout: 2 * time.Minute,
		})
		if err == nil && result.ExitCode == 0 && len(bytes.TrimSpace(result.Stdout)) > 1 && json.Valid(bytes.TrimSpace(result.Stdout)) {
			snapshot.Models[index].Verified = true
			verified++
		}
	}
	snapshot.CatalogComplete = false
	if verified > 0 {
		snapshot.Diagnostic = fmt.Sprintf("Claude authenticated; %d candidate model aliases verified; catalog verified, not exhaustive", verified)
		snapshot.Next = "run l7 route explain"
	} else {
		snapshot.Diagnostic = "Claude authenticated but no candidate model alias was verified; catalog is not exhaustive"
		snapshot.Next = "check configured Claude model access and retry l7 providers probe"
	}
	return snapshot
}

func (adapter Adapter) discoverCodex(ctx context.Context, id string) domain.ProviderSnapshot {
	executable, err := adapter.resolve("codex")
	if err != nil {
		return unavailable(id, domain.ProviderKindCodexAppServer, "Codex executable is unavailable")
	}
	version, err := adapter.version(ctx, executable)
	if err != nil {
		return failedNative(id, domain.ProviderKindCodexAppServer, executable, "Codex version probe failed")
	}
	discoveryCtx, cancel := context.WithTimeout(ctx, discoveryTimeout)
	defer cancel()
	data, runErr := adapter.codex(discoveryCtx, executable.Path)
	snapshot := domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: id, Kind: domain.ProviderKindCodexAppServer, Executable: executable.Path, Version: version, ExecutableDigest: executable.Digest, Authentication: domain.AuthUnknown, Models: []domain.ModelCapability{}, Next: "run l7 providers probe"}
	if runErr != nil {
		snapshot.Diagnostic = "Codex app-server discovery failed"
		return snapshot
	}
	account, models, quota, err := parseCodexMessages(data)
	if err != nil {
		snapshot.Diagnostic = err.Error()
		return snapshot
	}
	if account {
		snapshot.Authentication = domain.AuthAuthenticated
	} else {
		snapshot.Authentication = domain.AuthUnauthenticated
	}
	snapshot.Models = models
	snapshot.Quota = quota
	snapshot.CatalogComplete = account && len(models) > 0
	if snapshot.CatalogComplete {
		snapshot.Diagnostic = "authenticated Codex model catalog observed"
		snapshot.Next = "run l7 route explain"
	} else if account {
		snapshot.Diagnostic = "Codex is authenticated but no model catalog was observed"
	}
	return snapshot
}

func batchCodexDiscovery(run RunFunc) func(context.Context, string) ([]byte, error) {
	return func(ctx context.Context, executable string) ([]byte, error) {
		input := []byte(strings.Join([]string{
			`{"id":1,"method":"initialize","params":{"clientInfo":{"name":"level7","title":"Level 7","version":"1.0.0"}}}`,
			`{"method":"initialized","params":{}}`,
			`{"id":2,"method":"account/read","params":{"refreshToken":false}}`,
			`{"id":3,"method":"model/list","params":{}}`,
			`{"id":4,"method":"account/rateLimits/read","params":{}}`,
		}, "\n") + "\n")
		result, err := run(ctx, processadapter.Request{
			Executable: executable, Arguments: []string{"app-server"}, Input: input, Directory: "/",
			Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: maxDiscoveryOutput, Timeout: discoveryTimeout,
		})
		if err != nil || result.ExitCode != 0 {
			return nil, errors.New("Codex batch discovery failed")
		}
		return result.Stdout, nil
	}
}

func (adapter Adapter) discoverClaude(ctx context.Context, id string) domain.ProviderSnapshot {
	executable, err := adapter.resolve("claude")
	if err != nil {
		return unavailable(id, domain.ProviderKindClaudeCLI, "Claude executable is unavailable")
	}
	version, err := adapter.version(ctx, executable)
	if err != nil {
		return failedNative(id, domain.ProviderKindClaudeCLI, executable, "Claude version probe failed")
	}
	result, runErr := adapter.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: []string{"auth", "status", "--json"}, Directory: "/",
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 256 << 10, Timeout: discoveryTimeout,
	})
	snapshot := domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: id, Kind: domain.ProviderKindClaudeCLI, Executable: executable.Path, Version: version, ExecutableDigest: executable.Digest, Authentication: domain.AuthUnknown, CatalogComplete: false, Models: []domain.ModelCapability{}, Next: "run l7 providers probe"}
	if runErr != nil || result.ExitCode != 0 {
		snapshot.Authentication = domain.AuthUnauthenticated
		snapshot.Diagnostic = "Claude authentication status is unavailable"
		return snapshot
	}
	var status map[string]any
	if len(result.Stdout) < 2 || len(result.Stdout) > 256<<10 || json.Unmarshal(bytes.TrimSpace(result.Stdout), &status) != nil {
		snapshot.Diagnostic = "Claude authentication status is malformed"
		return snapshot
	}
	if boolValue(status, "loggedIn", "authenticated", "isAuthenticated") {
		snapshot.Authentication = domain.AuthAuthenticated
	} else {
		snapshot.Authentication = domain.AuthUnauthenticated
		snapshot.Diagnostic = "Claude is not authenticated"
		return snapshot
	}
	// Claude Code does not expose an approved exhaustive subscription catalog.
	// These aliases are candidates only; providers probe verifies them before
	// routing and stores the resulting snapshot in Git-local state.
	for _, model := range []struct {
		id      string
		cost    int
		latency int
	}{
		{id: "haiku", cost: 1, latency: 1},
		{id: "sonnet", cost: 3, latency: 2},
		{id: "opus", cost: 5, latency: 4},
	} {
		snapshot.Models = append(snapshot.Models, domain.ModelCapability{
			ID: model.id, DisplayName: "Claude " + model.id, ContextWindow: 200_000,
			Languages:     []string{"*"},
			SupportsTools: true, SupportsEditing: true, SupportsResume: true,
			Efforts:   []domain.ReasoningEffort{domain.EffortLow, domain.EffortMedium, domain.EffortHigh},
			CostClass: model.cost, LatencyClass: model.latency, Verified: false,
		})
	}
	snapshot.Diagnostic = "Claude authenticated; model aliases are candidates pending explicit probe and the catalog is not exhaustive"
	return snapshot
}

func (adapter Adapter) version(ctx context.Context, executable processadapter.Executable) (string, error) {
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: []string{"--version"}, Directory: "/",
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: discoveryTimeout,
	})
	if err != nil || result.ExitCode != 0 {
		return "", errors.New("version process failed")
	}
	value := strings.TrimSpace(string(append(append([]byte{}, result.Stdout...), result.Stderr...)))
	if value == "" || len(value) > 256 || strings.ContainsAny(value, "\x00\r\n") {
		return "", errors.New("version is invalid")
	}
	return value, nil
}

func gatewaySnapshot(provider orchestrationconfig.Provider) domain.ProviderSnapshot {
	snapshot := domain.ProviderSnapshot{
		Schema: domain.OrchestrationSchema, ID: provider.ID, Kind: provider.Kind,
		Authentication: domain.AuthUnknown, CatalogComplete: false, Models: []domain.ModelCapability{},
		Diagnostic: "gateway is configured but has not been network-probed", Next: "run l7 providers probe",
	}
	for _, model := range provider.Models {
		snapshot.Models = append(snapshot.Models, domain.ModelCapability{
			ID: model.ID, DisplayName: model.ID, ContextWindow: model.ContextWindow,
			Languages:     append([]string{}, model.Languages...),
			SupportsTools: model.SupportsTools, SupportsEditing: model.SupportsEditing, SupportsResume: model.SupportsResume,
			Efforts: append([]domain.ReasoningEffort{}, model.Efforts...), CostClass: model.CostClass, LatencyClass: model.LatencyClass,
			Verified: false,
		})
	}
	return snapshot
}

func parseCodexMessages(data []byte) (bool, []domain.ModelCapability, domain.QuotaState, error) {
	if len(data) < 2 || len(data) > maxDiscoveryOutput {
		return false, nil, domain.QuotaState{}, errors.New("Codex discovery stream is empty or unbounded")
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	authenticated := false
	models := []domain.ModelCapability{}
	quota := domain.QuotaState{}
	seen := make(map[string]bool)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var message map[string]any
		if err := json.Unmarshal(line, &message); err != nil {
			return false, nil, domain.QuotaState{}, errors.New("Codex discovery emitted malformed JSON")
		}
		id := intValue(message["id"])
		result, _ := message["result"].(map[string]any)
		switch id {
		case 2:
			authenticated = result != nil && (result["account"] != nil || boolValue(result, "authenticated", "loggedIn"))
		case 3:
			for _, record := range objectList(result, "data", "models", "items") {
				model, ok := codexModel(record)
				if ok && !seen[model.ID] {
					seen[model.ID] = true
					models = append(models, model)
				}
			}
		case 4:
			quota = codexQuota(result)
		}
	}
	if err := scanner.Err(); err != nil {
		return false, nil, domain.QuotaState{}, errors.New("Codex discovery framing exceeds limits")
	}
	sortModels(models)
	return authenticated, models, quota, nil
}

func codexQuota(result map[string]any) domain.QuotaState {
	quota := domain.QuotaState{Source: "codex_app_server"}
	limits, _ := result["rateLimits"].(map[string]any)
	if limits == nil {
		limits = result
	}
	var latestReset int64
	for _, raw := range limits {
		window, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		used := intValue(first(window, "usedPercent", "used_percent", "percentUsed"))
		if used >= 100 || boolValue(window, "limited", "exhausted") {
			quota.Limited = true
		}
		reset := int64(intValue(first(window, "resetsAt", "resetAt", "reset_at")))
		if reset > latestReset {
			latestReset = reset
		}
	}
	if latestReset > 0 {
		quota.ResetAtUTC = time.Unix(latestReset, 0).UTC().Format(time.RFC3339)
	}
	return quota
}

func codexModel(record map[string]any) (domain.ModelCapability, bool) {
	id := stringValue(record, "id", "model", "slug")
	if id == "" || len(id) > 160 || strings.ContainsAny(id, "\x00\r\n") {
		return domain.ModelCapability{}, false
	}
	efforts := reasoningEfforts(record)
	if len(efforts) == 0 {
		efforts = []domain.ReasoningEffort{domain.EffortMedium}
	}
	window := intValue(first(record, "contextWindow", "context_window", "contextWindowTokens"))
	if window < 1024 {
		window = 128_000
	}
	return domain.ModelCapability{
		ID: id, DisplayName: stringValue(record, "displayName", "name"), ContextWindow: window,
		Languages:     []string{"*"},
		SupportsTools: true, SupportsEditing: true, SupportsResume: true, Efforts: efforts,
		CostClass: 3, LatencyClass: 3, Verified: true,
	}, true
}

func reasoningEfforts(record map[string]any) []domain.ReasoningEffort {
	value := first(record, "supportedReasoningEfforts", "supported_reasoning_efforts", "reasoningEfforts")
	items, _ := value.([]any)
	result := []domain.ReasoningEffort{}
	seen := make(map[domain.ReasoningEffort]bool)
	for _, item := range items {
		text, _ := item.(string)
		if object, ok := item.(map[string]any); ok {
			text = stringValue(object, "reasoningEffort", "effort", "value")
		}
		effort := domain.ReasoningEffort(text)
		if effort.Valid() && !seen[effort] {
			seen[effort] = true
			result = append(result, effort)
		}
	}
	return result
}

func unavailable(id string, kind domain.ProviderKind, diagnostic string) domain.ProviderSnapshot {
	return domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: id, Kind: kind, Authentication: domain.AuthUnavailable, Models: []domain.ModelCapability{}, Diagnostic: diagnostic, Next: "install or configure the provider, then run l7 providers probe"}
}

func failedNative(id string, kind domain.ProviderKind, executable processadapter.Executable, diagnostic string) domain.ProviderSnapshot {
	return domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: id, Kind: kind, Executable: executable.Path, ExecutableDigest: executable.Digest, Authentication: domain.AuthUnknown, Models: []domain.ModelCapability{}, Diagnostic: diagnostic, Next: "repair the provider installation, then run l7 providers probe"}
}

func first(object map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := object[key]; ok {
			return value
		}
	}
	return nil
}

func stringValue(object map[string]any, keys ...string) string {
	value, _ := first(object, keys...).(string)
	return value
}

func boolValue(object map[string]any, keys ...string) bool {
	value, _ := first(object, keys...).(bool)
	return value
}

func intValue(value any) int {
	switch number := value.(type) {
	case float64:
		return int(number)
	case json.Number:
		parsed, _ := strconv.Atoi(number.String())
		return parsed
	case int:
		return number
	default:
		return 0
	}
}

func objectList(object map[string]any, keys ...string) []map[string]any {
	if object == nil {
		return nil
	}
	items, _ := first(object, keys...).([]any)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if record, ok := item.(map[string]any); ok {
			result = append(result, record)
		}
	}
	return result
}

func sortModels(models []domain.ModelCapability) {
	for index := 1; index < len(models); index++ {
		value := models[index]
		cursor := index - 1
		for cursor >= 0 && models[cursor].ID > value.ID {
			models[cursor+1] = models[cursor]
			cursor--
		}
		models[cursor+1] = value
	}
}

func DebugMessage(snapshot domain.ProviderSnapshot) string {
	return fmt.Sprintf("%s %s auth=%s models=%d complete=%t", snapshot.ID, snapshot.Kind, snapshot.Authentication, len(snapshot.Models), snapshot.CatalogComplete)
}
