package codex

import (
	"context"
	"errors"
	"strings"
	"testing"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestRunTranslatesBothRolesWithoutDangerousBypass(t *testing.T) {
	for _, test := range []struct {
		role    domain.ProviderRole
		sandbox string
		output  string
	}{
		{role: domain.RoleImplementer, sandbox: "workspace-write", output: `{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}`},
		{role: domain.RoleReviewer, sandbox: "read-only", output: `{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`},
	} {
		t.Run(string(test.role), func(t *testing.T) {
			var invocation processadapter.Request
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
				}
				invocation = request
				event := `{"type":"item.completed","item":{"id":"final","type":"agent_message","text":` + quoted(test.output) + `}}` + "\n"
				return processadapter.Result{ExitCode: 0, Stdout: []byte(event)}, nil
			}))
			response, err := adapter.Run(context.Background(), providerTask(test.role), 1<<20, 30)
			if err != nil || response.Role != test.role || response.Identity.Provider != domain.ProviderCodex {
				t.Fatalf("Run()=%+v error=%v", response, err)
			}
			joined := strings.Join(invocation.Arguments, " ")
			if !strings.Contains(joined, "--sandbox "+test.sandbox) || strings.Contains(joined, "dangerously") || invocation.Directory != "/repo" || !strings.Contains(string(invocation.Input), `"role": "`+string(test.role)+`"`) {
				t.Fatalf("invocation=%+v", invocation)
			}
		})
	}
}

func TestProbeDegradesUnknownVersion(t *testing.T) {
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte("codex-cli 0.150.0\n")}, nil
	}))
	identity, err := adapter.Probe(context.Background())
	if err != nil || identity.Capability != domain.CapabilityDegraded {
		t.Fatalf("Probe()=%+v error=%v", identity, err)
	}
}

func TestRunRejectsMalformedOrFailedEvents(t *testing.T) {
	outputs := []string{
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"},"unknown":true}` + "\n",
		`{"type":"turn.failed","error":{"message":"failed"}}` + "\n",
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"}}` + "\n" + `{"type":"item.completed","item":{"id":"two","type":"agent_message","text":"{}"}}` + "\n",
	}
	for _, output := range outputs {
		if _, err := parseEvents([]byte(output), domain.RoleImplementer); err == nil {
			t.Fatalf("parseEvents(%q) unexpectedly passed", output)
		}
	}
	if _, err := parseEvents([]byte{0xff}, domain.RoleImplementer); err == nil {
		t.Fatal("parseEvents accepted invalid UTF-8")
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

func BenchmarkParseEvents(b *testing.B) {
	output := []byte(`{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"No blocker.\",\"findings\":[],\"decision\":\"GO\"}"}}` + "\n")
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := parseEvents(output, domain.RoleReviewer); err != nil {
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
		Role: role, Provider: domain.ProviderCodex, RepositoryRoot: "/repo", ChangeID: "change", Tier: domain.TierHighRisk,
		Base: strings.Repeat("a", 40), Candidate: domain.CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)},
		Problem: "Implement the approved change.", Scope: []string{"internal/example/**"},
	}
}

func quoted(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range value {
		switch character {
		case '\\', '"':
			builder.WriteByte('\\')
			builder.WriteRune(character)
		case '\n':
			builder.WriteString(`\n`)
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
