package claude

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestRunTranslatesBothRolesWithoutBypassPermissions(t *testing.T) {
	for _, test := range []struct {
		role       domain.ProviderRole
		permission string
		tools      string
		terminal   string
	}{
		{role: domain.RoleImplementer, permission: "acceptEdits", tools: "Read,Glob,Grep,Edit,Write,Bash", terminal: `{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}`},
		{role: domain.RoleReviewer, permission: "plan", tools: "Read,Glob,Grep", terminal: `{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`},
	} {
		t.Run(string(test.role), func(t *testing.T) {
			schema, err := provider.TerminalSchema(test.role)
			if err != nil {
				t.Fatal(err)
			}
			var invocation processadapter.Request
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + " (Claude Code)\n")}, nil
				}
				invocation = request
				output := `{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":` + test.terminal + `}` + "\n"
				return processadapter.Result{ExitCode: 0, Stdout: []byte(output)}, nil
			}))
			response, err := adapter.Run(context.Background(), providerTask(test.role), 1<<20, 30)
			if err != nil || response.Role != test.role || response.Identity.Provider != domain.ProviderClaude {
				t.Fatalf("Run()=%+v error=%v", response, err)
			}
			expectedArguments := []string{
				"--safe-mode",
				"--disable-slash-commands",
				"--print",
				"--input-format", "text",
				"--max-turns", "64",
				"--tools", test.tools,
				"--disallowedTools", "WebFetch,WebSearch,NotebookEdit,Task,Skill",
				"--permission-mode", test.permission,
				"--strict-mcp-config",
				"--no-chrome",
				"--no-session-persistence",
				"--output-format", "json",
				"--json-schema", schema,
			}
			joined := strings.Join(invocation.Arguments, " ")
			if !slices.Equal(invocation.Arguments, expectedArguments) || strings.Contains(joined, "--bare") || strings.Contains(joined, "bypassPermissions") || strings.Contains(joined, "dangerously-skip") || invocation.Directory != "/repo" || !strings.Contains(string(invocation.Input), `"role": "`+string(test.role)+`"`) {
				t.Fatalf("invocation=%+v", invocation)
			}
		})
	}
}

func TestProbeAcceptsOnlyExactTargetVersionSpellings(t *testing.T) {
	for _, test := range []struct {
		name            string
		version         string
		expectedVersion string
		capability      domain.CapabilityState
	}{
		{name: "bare", version: CompatibleVersion, expectedVersion: CompatibleVersion, capability: domain.CapabilityAvailable},
		{name: "product suffix", version: CompatibleVersion + " (Claude Code)", expectedVersion: CompatibleVersion + " (Claude Code)", capability: domain.CapabilityAvailable},
		{name: "old", version: "2.1.241", expectedVersion: "2.1.241", capability: domain.CapabilityDegraded},
		{name: "adjacent lower", version: "2.1.246 (Claude Code)", expectedVersion: "2.1.246 (Claude Code)", capability: domain.CapabilityDegraded},
		{name: "adjacent higher", version: "2.1.248", expectedVersion: "2.1.248", capability: domain.CapabilityDegraded},
		{name: "prefixed", version: "v" + CompatibleVersion, expectedVersion: "v" + CompatibleVersion, capability: domain.CapabilityDegraded},
		{name: "extended suffix", version: CompatibleVersion + " (Claude Code) extra", expectedVersion: CompatibleVersion + " (Claude Code) extra", capability: domain.CapabilityDegraded},
		{name: "padded", version: " " + CompatibleVersion + " ", capability: domain.CapabilityDegraded},
		{name: "multiline", version: CompatibleVersion + "\nextra", capability: domain.CapabilityDegraded},
	} {
		t.Run(test.name, func(t *testing.T) {
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
				return processadapter.Result{ExitCode: 0, Stdout: []byte(test.version + "\n")}, nil
			}))
			identity, err := adapter.Probe(context.Background())
			if err != nil || identity.Version != test.expectedVersion || identity.Capability != test.capability {
				t.Fatalf("Probe()=%+v error=%v", identity, err)
			}
		})
	}
}

func TestRunRejectsMissingMalformedOrNonEmptyPermissionDenials(t *testing.T) {
	for _, output := range [][]byte{
		[]byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":null,"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":"none","structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":{},"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[{"tool_name":"Bash","tool_use_id":"tool-1","tool_input":{}}],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
	} {
		if _, err := parseResult(output, domain.RoleImplementer); err == nil {
			t.Fatalf("parseResult(%s) unexpectedly accepted permission denials", output)
		}
	}
}

func TestRunRejectsUnknownFailedDuplicateTrailingOrProseResults(t *testing.T) {
	outputs := [][]byte{
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{},"unknown":true}`),
		[]byte(`{"type":"result","subtype":"error","is_error":true,"permission_denials":[],"structured_output":{}}`),
		[]byte(`implemented`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]},"structured_output":{}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}} {}`),
		[]byte("\n" + `{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}` + "\n\n"),
	}
	for _, output := range outputs {
		if _, err := parseResult(output, domain.RoleImplementer); err == nil {
			t.Fatalf("parseResult(%s) unexpectedly passed", output)
		}
	}
	if _, err := parseResult([]byte{0xff}, domain.RoleImplementer); err == nil {
		t.Fatal("parseResult accepted invalid UTF-8")
	}
}

func TestRunDoesNotSemanticallyInvokeDegradedVersion(t *testing.T) {
	invocations := 0
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		invocations++
		return processadapter.Result{ExitCode: 0, Stdout: []byte("2.1.241 (Claude Code)\n")}, nil
	}))
	response, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer), 1<<20, 30)
	if err == nil || response.Identity.Capability != domain.CapabilityDegraded || invocations != 1 {
		t.Fatalf("Run()=%+v error=%v invocations=%d", response, err, invocations)
	}
}

func TestArgumentsRejectInvalidRole(t *testing.T) {
	if result, err := arguments(domain.ProviderRole("invalid")); err == nil || result != nil {
		t.Fatalf("arguments(invalid)=%v error=%v", result, err)
	}
}

func TestRunPropagatesInFlightCancellationWithoutAcceptingOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	invocations := 0
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(ctx context.Context, request processadapter.Request) (processadapter.Result, error) {
		invocations++
		if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
		}
		close(started)
		<-ctx.Done()
		output := `{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Must not be accepted.","findings":[]}}`
		return processadapter.Result{ExitCode: -1, Stdout: []byte(output)}, ctx.Err()
	}))
	type result struct {
		response domain.ProviderResponse
		err      error
	}
	finished := make(chan result, 1)
	go func() {
		response, err := adapter.Run(ctx, providerTask(domain.RoleImplementer), 1<<20, 30)
		finished <- result{response: response, err: err}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Claude invocation did not start within the cancellation bound")
	}
	cancel()
	var got result
	select {
	case got = <-finished:
	case <-time.After(time.Second):
		t.Fatal("Claude invocation did not return within the cancellation bound")
	}
	if !errors.Is(got.err, context.Canceled) || got.response.Role != domain.RoleImplementer || got.response.Identity.Capability != domain.CapabilityAvailable || got.response.Summary != "" || len(got.response.Findings) != 0 || invocations != 2 {
		t.Fatalf("cancelled Run()=%+v error=%v invocations=%d", got.response, got.err, invocations)
	}
}

func TestRunPropagatesPreCancelledContextWithoutSemanticInvocation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var requests []processadapter.Request
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(ctx context.Context, request processadapter.Request) (processadapter.Result, error) {
		requests = append(requests, request)
		return processadapter.Result{}, ctx.Err()
	}))
	response, err := adapter.Run(ctx, providerTask(domain.RoleImplementer), 1<<20, 30)
	if !errors.Is(err, context.Canceled) || response.Role != domain.RoleImplementer || response.Identity.Capability != domain.CapabilityDegraded || len(requests) != 1 || !slices.Equal(requests[0].Arguments, []string{"--version"}) {
		t.Fatalf("pre-cancelled Run()=%+v error=%v requests=%+v", response, err, requests)
	}
}

func TestRunPropagatesUnavailableExecutable(t *testing.T) {
	adapter := NewWithRuntime(provider.NewRuntime(func(string) (processadapter.Executable, error) {
		return processadapter.Executable{}, errors.New("missing")
	}, nil))
	response, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer), 1<<20, 30)
	if err == nil || response.Identity.Capability != domain.CapabilityUnavailable {
		t.Fatalf("Run()=%+v error=%v", response, err)
	}
}

func FuzzParseResult(f *testing.F) {
	f.Add([]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`), false)
	f.Add([]byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}`), true)
	f.Add([]byte{0xff}, false)
	f.Fuzz(func(t *testing.T, data []byte, reviewer bool) {
		role := domain.RoleImplementer
		if reviewer {
			role = domain.RoleReviewer
		}
		response, err := parseResult(data, role)
		if err == nil && (response.Role != role || (role == domain.RoleReviewer && !response.Decision.Valid()) || (role == domain.RoleImplementer && response.Decision != "")) {
			t.Fatalf("successful parse violated role contract: %+v", response)
		}
	})
}

func BenchmarkParseResult(b *testing.B) {
	output := []byte(`{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}`)
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := parseResult(output, domain.RoleReviewer); err != nil {
			b.Fatal(err)
		}
	}
}

func fakeResolve(name string) provider.ResolveFunc {
	return func(string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("a", 64)}, nil
	}
}

func providerTask(role domain.ProviderRole) domain.ProviderTask {
	return domain.ProviderTask{
		Role: role, Provider: domain.ProviderClaude, RepositoryRoot: "/repo", ChangeID: "change", Tier: domain.TierHighRisk,
		Base: strings.Repeat("a", 40), Candidate: domain.CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)},
		Problem: "Implement the approved change.", Scope: []string{"internal/example/**"},
	}
}
