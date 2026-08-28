package claude

import (
	"context"
	"errors"
	"strings"
	"testing"

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
			var invocation processadapter.Request
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + " (Claude Code)\n")}, nil
				}
				invocation = request
				output := `{"type":"result","subtype":"success","is_error":false,"structured_output":` + test.terminal + `}` + "\n"
				return processadapter.Result{ExitCode: 0, Stdout: []byte(output)}, nil
			}))
			response, err := adapter.Run(context.Background(), providerTask(test.role), 1<<20, 30)
			if err != nil || response.Role != test.role || response.Identity.Provider != domain.ProviderClaude {
				t.Fatalf("Run()=%+v error=%v", response, err)
			}
			joined := strings.Join(invocation.Arguments, " ")
			if !strings.Contains(joined, "--permission-mode "+test.permission) || !strings.Contains(joined, "--tools "+test.tools) || !strings.Contains(joined, "--safe-mode") || strings.Contains(joined, "--bare") || !strings.Contains(joined, "--max-turns 64") || strings.Contains(joined, "bypassPermissions") || strings.Contains(joined, "dangerously-skip") || invocation.Directory != "/repo" || !strings.Contains(string(invocation.Input), `"role": "`+string(test.role)+`"`) {
				t.Fatalf("invocation=%+v", invocation)
			}
			if test.role == domain.RoleReviewer && strings.Contains(test.tools, "Bash") {
				t.Fatalf("reviewer tools include Bash: %q", test.tools)
			}
		})
	}
}

func TestProbeDegradesUnknownVersion(t *testing.T) {
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("claude"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte("2.2.0 (Claude Code)\n")}, nil
	}))
	identity, err := adapter.Probe(context.Background())
	if err != nil || identity.Capability != domain.CapabilityDegraded {
		t.Fatalf("Probe()=%+v error=%v", identity, err)
	}
}

func TestRunRejectsUnknownErrorOrProseResults(t *testing.T) {
	outputs := [][]byte{
		[]byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{},"unknown":true}`),
		[]byte(`{"type":"result","subtype":"error","is_error":true,"structured_output":{}}`),
		[]byte(`implemented`),
		[]byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]},"structured_output":{}}`),
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
	f.Add([]byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}`), false)
	f.Add([]byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}`), true)
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
	output := []byte(`{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}`)
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
