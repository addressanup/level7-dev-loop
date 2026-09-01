package discovery

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestDiscoverCodexUsesAppServerCapabilities(t *testing.T) {
	adapter := NewWith(
		func(name string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("a", 64)}, nil
		},
		func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
				return processadapter.Result{Stdout: []byte("codex-cli 9.9.9\n")}, nil
			}
			if !strings.Contains(string(request.Input), `"method":"account/rateLimits/read"`) {
				t.Fatal("Codex discovery omitted the app-server rate-limit capability")
			}
			return processadapter.Result{Stdout: []byte(
				`{"id":1,"result":{}}` + "\n" +
					`{"id":2,"result":{"account":{"type":"chatgpt"}}}` + "\n" +
					`{"id":3,"result":{"data":[{"id":"model-b","displayName":"Model B","contextWindow":200000,"supportedReasoningEfforts":[{"reasoningEffort":"low"},{"reasoningEffort":"high"}]},{"id":"model-a","context_window":128000,"supported_reasoning_efforts":["medium"]}]}}` + "\n" +
					`{"id":4,"result":{"rateLimits":{"primary":{"usedPercent":100,"resetsAt":1800000000}}}}` + "\n")}, nil
		},
		func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) },
	)
	configuration := orchestrationconfig.Default()
	configuration.Providers = configuration.Providers[:1]
	snapshots := adapter.Discover(context.Background(), configuration)
	if len(snapshots) != 1 || snapshots[0].Version != "codex-cli 9.9.9" || snapshots[0].Authentication != domain.AuthAuthenticated || !snapshots[0].CatalogComplete || len(snapshots[0].Models) != 2 || snapshots[0].Models[0].ID != "model-a" || !snapshots[0].Models[1].Verified || !snapshots[0].Quota.Limited || snapshots[0].Quota.ResetAtUTC != "2027-01-15T08:00:00Z" {
		t.Fatalf("unexpected snapshot: %#v", snapshots)
	}
}

func TestDiscoverClaudeLabelsCatalogNonExhaustive(t *testing.T) {
	adapter := NewWith(
		func(name string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("b", 64)}, nil
		},
		func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			if len(request.Arguments) == 1 {
				return processadapter.Result{Stdout: []byte("2.1.247 (Claude Code)\n")}, nil
			}
			return processadapter.Result{Stdout: []byte(`{"loggedIn":true,"authMethod":"claude.ai"}`)}, nil
		},
		time.Now,
	)
	configuration := orchestrationconfig.Default()
	configuration.Providers = configuration.Providers[1:]
	snapshot := adapter.Discover(context.Background(), configuration)[0]
	if snapshot.Authentication != domain.AuthAuthenticated || snapshot.CatalogComplete || len(snapshot.Models) != 3 || snapshot.Models[0].Verified || !strings.Contains(snapshot.Diagnostic, "not exhaustive") {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestProbeClaudeVerifiesCandidatesWithoutClaimingExhaustiveCatalog(t *testing.T) {
	calls := 0
	adapter := NewWith(nil, func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		calls++
		if len(request.Arguments) < 2 || request.Arguments[0] != "-p" || request.Arguments[len(request.Arguments)-1] != "" {
			t.Fatalf("unsafe Claude probe arguments: %#v", request.Arguments)
		}
		return processadapter.Result{Stdout: []byte(`{"result":"OK"}`)}, nil
	}, time.Now)
	snapshot := domain.ProviderSnapshot{
		Schema: domain.OrchestrationSchema, ID: "claude-local", Kind: domain.ProviderKindClaudeCLI,
		Executable: "/usr/bin/claude", Authentication: domain.AuthAuthenticated,
		Models: []domain.ModelCapability{{ID: "haiku"}, {ID: "sonnet"}},
	}
	probed := adapter.ProbeClaude(context.Background(), snapshot, "/tmp")
	if calls != 2 || probed.CatalogComplete || !probed.Models[0].Verified || !probed.Models[1].Verified || !strings.Contains(probed.Diagnostic, "not exhaustive") {
		t.Fatalf("calls=%d snapshot=%#v", calls, probed)
	}
}

func TestGatewayStartsUnverified(t *testing.T) {
	configuration := orchestrationconfig.Default()
	configuration.Providers = []orchestrationconfig.Provider{{
		ID: "gateway", Kind: domain.ProviderKindOpenAIResponses, Enabled: true, Endpoint: "https://example.invalid/v1/responses",
		CatalogURL: "https://example.invalid/v1/l7-models",
		Credential: orchestrationconfig.Credential{Source: "env", Reference: "GATEWAY_TOKEN"},
		Models:     []orchestrationconfig.Model{{ID: "model", Languages: []string{"*"}, ContextWindow: 10_000, SupportsTools: true, SupportsEditing: true, Efforts: []domain.ReasoningEffort{domain.EffortMedium}, CostClass: 2, LatencyClass: 2}},
	}}
	snapshot := NewWith(nil, nil, time.Now).Discover(context.Background(), configuration)[0]
	if snapshot.Authentication != domain.AuthUnknown || snapshot.CatalogComplete || len(snapshot.Models) != 1 || snapshot.Models[0].Verified {
		t.Fatalf("gateway was trusted before probe: %#v", snapshot)
	}
}
