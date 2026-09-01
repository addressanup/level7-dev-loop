package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type doFunc func(*http.Request) (*http.Response, error)

func (function doFunc) Do(request *http.Request) (*http.Response, error) { return function(request) }

type brokerFunc func(context.Context, string, []byte) ([]byte, error)

func (function brokerFunc) Call(ctx context.Context, name string, input []byte) ([]byte, error) {
	return function(ctx, name, input)
}

func TestProbeGatewayVerifiesConfiguredModelWithoutLeakingCredential(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	client := NewWith(doFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "Bearer super-secret-value" {
			t.Fatal("missing auth header")
		}
		return response(200, `{"id":"response"}`), nil
	}), nil, nil, func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) })
	snapshot := client.Probe(context.Background(), openAIProvider())
	if snapshot.Authentication != domain.AuthAuthenticated || len(snapshot.Models) != 1 || !snapshot.Models[0].Verified || strings.Contains(snapshot.Diagnostic, "secret") {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestProbeGatewayDiscoversAndVerifiesStrictCatalog(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	requests := 0
	client := NewWith(doFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Authorization") != "Bearer super-secret-value" {
			t.Fatal("missing catalog auth header")
		}
		if request.Method == http.MethodGet && request.URL.String() == "https://gateway.example/v1/l7-models" {
			return response(200, `{"schema":1,"models":[{"id":"catalog-model","languages":["go","python"],"context_window":200000,"supports_tools":true,"supports_editing":true,"supports_resume":false,"efforts":["medium","high"],"cost_class":2,"latency_class":2}]}`), nil
		}
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL)
		}
		return response(200, `{"id":"probe"}`), nil
	}), nil, nil, time.Now)
	provider := openAIProvider()
	provider.CatalogURL = "https://gateway.example/v1/l7-models"
	provider.Models = nil
	snapshot := client.Probe(context.Background(), provider)
	if requests != 2 || snapshot.Authentication != domain.AuthAuthenticated || !snapshot.CatalogComplete || len(snapshot.Models) != 1 || snapshot.Models[0].ID != "catalog-model" || !snapshot.Models[0].Verified {
		t.Fatalf("requests=%d snapshot=%#v", requests, snapshot)
	}
}

func TestProbeGatewayRejectsMalformedCatalogWithoutLeakingBody(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		return response(200, `{"schema":1,"models":[{"id":"secret-super-secret-value"}]}`), nil
	}), nil, nil, time.Now)
	provider := openAIProvider()
	provider.CatalogURL = "https://gateway.example/v1/l7-models"
	provider.Models = nil
	snapshot := client.Probe(context.Background(), provider)
	if snapshot.CatalogComplete || strings.Contains(snapshot.Diagnostic, "secret") || strings.Contains(snapshot.Diagnostic, "super-secret-value") {
		t.Fatalf("unsafe catalog result: %#v", snapshot)
	}
}

func TestProbeGatewayDistinguishesCredentialAndModelRejection(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	unauthorized := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusUnauthorized, `{"error":"super-secret-value"}`), nil
	}), nil, nil, time.Now).Probe(context.Background(), openAIProvider())
	if unauthorized.Authentication != domain.AuthUnauthenticated || unauthorized.Models[0].Verified || strings.Contains(unauthorized.Diagnostic, "super-secret-value") {
		t.Fatalf("unsafe unauthorized snapshot: %#v", unauthorized)
	}
	rejected := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusBadRequest, `{"error":"unsupported model"}`), nil
	}), nil, nil, time.Now).Probe(context.Background(), openAIProvider())
	if rejected.Authentication != domain.AuthUnknown || rejected.Models[0].Verified || !strings.Contains(rejected.Diagnostic, "did not verify") {
		t.Fatalf("unexpected model rejection: %#v", rejected)
	}
}

func TestOpenAIWorkerUsesBrokerAndRequiresFinish(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	turn := 0
	client := NewWith(doFunc(func(request *http.Request) (*http.Response, error) {
		turn++
		data, _ := io.ReadAll(request.Body)
		if bytes.Contains(data, []byte("super-secret-value")) {
			t.Fatal("secret entered request JSON")
		}
		if turn == 1 {
			return response(200, `{"id":"r1","output":[{"type":"function_call","call_id":"c1","name":"read_file","arguments":"{\"path\":\"src/main.go\",\"offset\":0,\"limit\":100}"}]}`), nil
		}
		return response(200, `{"id":"r2","output":[{"type":"function_call","call_id":"c2","name":"finish","arguments":"{\"outcome\":\"complete\",\"summary\":\"Implemented.\",\"findings\":[]}"}]}`), nil
	}), nil, nil, time.Now)
	called := false
	result, err := client.Run(context.Background(), Assignment{Provider: openAIProvider(), Model: "model", Effort: domain.EffortMedium, Prompt: "Implement.", Reviewer: false}, brokerFunc(func(_ context.Context, name string, input []byte) ([]byte, error) {
		called = name == "read_file" && bytes.Contains(input, []byte("src/main.go"))
		return []byte(`{"ok":true,"content":"package main"}`), nil
	}))
	if err != nil || !called || result.Summary != "Implemented." || turn != 2 {
		t.Fatalf("result=%#v called=%t turn=%d err=%v", result, called, turn, err)
	}
}

func TestAnthropicWorkerEnforcesReviewerDecision(t *testing.T) {
	t.Setenv("L7_ANTHROPIC_KEY", "super-secret-value")
	provider := anthropicProvider()
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		return response(200, `{"content":[{"type":"tool_use","id":"tool1","name":"finish","input":{"outcome":"complete","summary":"Reviewed.","findings":[],"decision":"GO"}}]}`), nil
	}), nil, nil, time.Now)
	result, err := client.Run(context.Background(), Assignment{Provider: provider, Model: "model", Effort: domain.EffortHigh, Prompt: "Review.", Reviewer: true}, brokerFunc(func(context.Context, string, []byte) ([]byte, error) { t.Fatal("broker called"); return nil, nil }))
	if err != nil || result.Decision != domain.DecisionGO {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestOpenAIWorkerAcceptsCompletedEventStream(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		value := response(200, "event: response.completed\n"+`data: {"type":"response.completed","response":{"id":"r1","output":[{"type":"function_call","call_id":"c1","name":"finish","arguments":"{\"outcome\":\"complete\",\"summary\":\"Streamed.\",\"findings\":[]}"}]}}`+"\n\n")
		value.Header.Set("Content-Type", "text/event-stream")
		return value, nil
	}), nil, nil, time.Now)
	result, err := client.Run(context.Background(), Assignment{Provider: openAIProvider(), Model: "model", Effort: domain.EffortMedium, Prompt: "Implement."}, brokerFunc(func(context.Context, string, []byte) ([]byte, error) {
		t.Fatal("broker called")
		return nil, nil
	}))
	if err != nil || result.Summary != "Streamed." {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestAnthropicWorkerAcceptsToolUseEventStream(t *testing.T) {
	t.Setenv("L7_ANTHROPIC_KEY", "super-secret-value")
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		stream := strings.Join([]string{
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"tool1","name":"finish","input":{}}}`,
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"outcome\":\"complete\",\"summary\":\"Reviewed stream.\",\"findings\":[],\"decision\":\"GO\"}"}}`,
			`data: {"type":"content_block_stop","index":0}`,
			`data: {"type":"message_stop"}`,
		}, "\n\n") + "\n\n"
		value := response(200, stream)
		value.Header.Set("Content-Type", "text/event-stream; charset=utf-8")
		return value, nil
	}), nil, nil, time.Now)
	result, err := client.Run(context.Background(), Assignment{Provider: anthropicProvider(), Model: "model", Effort: domain.EffortHigh, Prompt: "Review.", Reviewer: true}, brokerFunc(func(context.Context, string, []byte) ([]byte, error) {
		t.Fatal("broker called")
		return nil, nil
	}))
	if err != nil || result.Decision != domain.DecisionGO || result.Summary != "Reviewed stream." {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestEventStreamInterruptionFailsClosed(t *testing.T) {
	stream := []byte("data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tool1\",\"name\":\"finish\",\"input\":{}}}\n\n")
	if _, err := decodeEventStream(stream); err == nil {
		t.Fatal("interrupted Anthropic stream was accepted")
	}
}

func TestKeychainErrorDoesNotExposeProcessOutput(t *testing.T) {
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) { t.Fatal("HTTP called"); return nil, nil }),
		func(string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/security", Digest: strings.Repeat("a", 64)}, nil
		},
		func(context.Context, processadapter.Request) (processadapter.Result, error) {
			return processadapter.Result{ExitCode: 1, Stderr: []byte("secret-value")}, nil
		}, time.Now)
	_, err := client.credential(context.Background(), orchestrationconfig.Credential{Source: "keychain", Reference: "service/account"})
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func TestGatewayQuotaCarriesOnlyNaturalResetBoundary(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	now := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	client := NewWith(doFunc(func(*http.Request) (*http.Response, error) {
		value := response(http.StatusTooManyRequests, `{"error":"quota for super-secret-value"}`)
		value.Header.Set("Retry-After", "120")
		return value, nil
	}), nil, nil, func() time.Time { return now })
	_, err := client.Run(context.Background(), Assignment{Provider: openAIProvider(), Model: "model", Effort: domain.EffortMedium, Prompt: "Implement."}, brokerFunc(func(context.Context, string, []byte) ([]byte, error) { return nil, nil }))
	var quota *QuotaError
	if !errors.As(err, &quota) || quota.ResetAtUTC != "2026-08-31T00:02:00Z" || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe quota result: quota=%#v err=%v", quota, err)
	}
}

func TestGatewayCancellationDoesNotExposeCredential(t *testing.T) {
	t.Setenv("L7_GATEWAY_KEY", "super-secret-value")
	client := NewWith(doFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	}), nil, nil, time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Run(ctx, Assignment{Provider: openAIProvider(), Model: "model", Effort: domain.EffortMedium, Prompt: "Implement."}, brokerFunc(func(context.Context, string, []byte) ([]byte, error) { return nil, nil }))
	if err == nil || strings.Contains(err.Error(), "super-secret-value") {
		t.Fatalf("unsafe cancellation error: %v", err)
	}
}

func FuzzDecodeEventStream(f *testing.F) {
	f.Add([]byte("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"r\",\"output\":[]}}\n\n"))
	f.Add([]byte("data: [DONE]\n\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxHTTPBody {
			t.Skip()
		}
		decoded, err := decodeEventStream(data)
		if err == nil && !json.Valid(decoded) {
			t.Fatalf("decoder returned non-JSON success: %q", decoded)
		}
	})
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func openAIProvider() orchestrationconfig.Provider {
	return orchestrationconfig.Provider{ID: "gateway", Kind: domain.ProviderKindOpenAIResponses, Enabled: true, Endpoint: "https://gateway.example/v1/responses", Credential: orchestrationconfig.Credential{Source: "env", Reference: "L7_GATEWAY_KEY"}, Models: []orchestrationconfig.Model{{ID: "model", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, Efforts: []domain.ReasoningEffort{domain.EffortMedium}, CostClass: 2, LatencyClass: 2}}}
}

func anthropicProvider() orchestrationconfig.Provider {
	return orchestrationconfig.Provider{ID: "anthropic", Kind: domain.ProviderKindAnthropic, Enabled: true, Endpoint: "https://gateway.example/v1/messages", Credential: orchestrationconfig.Credential{Source: "env", Reference: "L7_ANTHROPIC_KEY"}, Models: []orchestrationconfig.Model{{ID: "model", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, Efforts: []domain.ReasoningEffort{domain.EffortHigh}, CostClass: 3, LatencyClass: 3}}}
}

var _ = os.ErrNotExist
