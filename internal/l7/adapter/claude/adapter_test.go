package claude

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/provider"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const actualClaudeUnknownOption = "--l7-unknown-option"

type actualClaudeInterfaceCase struct {
	name         string
	arguments    []string
	wantExitZero bool
}

func actualClaudeInterfaceCases(base []string) ([]actualClaudeInterfaceCase, error) {
	turns := -1
	for index, argument := range base {
		if argument == "--help" || argument == actualClaudeUnknownOption {
			return nil, errors.New("base Claude argv already contains a test-only interface option")
		}
		if argument != "--max-turns" {
			continue
		}
		if turns >= 0 || index+1 >= len(base) || base[index+1] != "64" {
			return nil, errors.New("base Claude argv must contain exactly one --max-turns 64 pair")
		}
		turns = index
	}
	if turns < 0 {
		return nil, errors.New("base Claude argv is missing --max-turns 64")
	}
	clone := func() []string { return append([]string{}, base...) }
	positive := append(clone(), "--help")
	unknown := append(clone(), actualClaudeUnknownOption, "--help")
	invalidTurns := clone()
	invalidTurns[turns+1] = "not-an-integer"
	invalidTurns = append(invalidTurns, "--help")
	return []actualClaudeInterfaceCase{
		{name: "positive", arguments: positive, wantExitZero: true},
		{name: "unknown-option", arguments: unknown},
		{name: "invalid-max-turns", arguments: invalidTurns},
	}, nil
}

func actualClaudeInterfaceOutcome(observation actualClaudeInterfaceCase, result processadapter.Result, err error) error {
	if err != nil {
		return fmt.Errorf("bounded no-model invocation failed: %w", err)
	}
	if observation.wantExitZero && result.ExitCode != 0 {
		return fmt.Errorf("expected successful exit, got %d", result.ExitCode)
	}
	if !observation.wantExitZero && result.ExitCode == 0 {
		return errors.New("negative parser control exited successfully")
	}
	return nil
}

func TestActualClaudeInterfaceCasesPreserveMaxTurnsAndRoleArgv(t *testing.T) {
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		base := arguments(role)
		original := append([]string{}, base...)
		observations, err := actualClaudeInterfaceCases(base)
		if err != nil || len(observations) != 3 {
			t.Fatalf("%s observations=%+v error=%v", role, observations, err)
		}
		if !slices.Equal(base, original) {
			t.Fatalf("%s base argv mutated: got=%v want=%v", role, base, original)
		}
		if !slices.Equal(observations[0].arguments[:len(base)], base) || observations[0].arguments[len(base)] != "--help" || !observations[0].wantExitZero {
			t.Fatalf("%s positive observation=%+v", role, observations[0])
		}
		if !slices.Equal(observations[1].arguments[:len(base)], base) || !slices.Equal(observations[1].arguments[len(base):], []string{actualClaudeUnknownOption, "--help"}) || observations[1].wantExitZero {
			t.Fatalf("%s unknown observation=%+v", role, observations[1])
		}
		if len(observations[2].arguments) != len(base)+1 || observations[2].arguments[len(base)] != "--help" || observations[2].wantExitZero {
			t.Fatalf("%s invalid turns observation=%+v", role, observations[2])
		}
		differences := 0
		for index := range base {
			if observations[2].arguments[index] != base[index] {
				differences++
				if base[index] != "64" || observations[2].arguments[index] != "not-an-integer" {
					t.Fatalf("%s invalid turns changed unexpected argument %d: got=%q want=%q", role, index, observations[2].arguments[index], base[index])
				}
			}
		}
		if differences != 1 {
			t.Fatalf("%s invalid turns differences=%d", role, differences)
		}
	}
}

func TestActualClaudeInterfaceCasesAndOutcomesFailClosed(t *testing.T) {
	for _, base := range [][]string{
		{"--print"},
		{"--max-turns"},
		{"--max-turns", "63"},
		{"--max-turns", "64", "--max-turns", "64"},
		{"--max-turns", "64", "--help"},
		{"--max-turns", "64", actualClaudeUnknownOption},
	} {
		if observations, err := actualClaudeInterfaceCases(base); err == nil || observations != nil {
			t.Fatalf("unsafe base argv accepted: %v", base)
		}
	}
	positive := actualClaudeInterfaceCase{name: "positive", wantExitZero: true}
	negative := actualClaudeInterfaceCase{name: "negative"}
	if err := actualClaudeInterfaceOutcome(positive, processadapter.Result{ExitCode: 0}, nil); err != nil {
		t.Fatalf("valid positive rejected: %v", err)
	}
	if err := actualClaudeInterfaceOutcome(negative, processadapter.Result{ExitCode: 2}, nil); err != nil {
		t.Fatalf("valid negative rejected: %v", err)
	}
	for _, test := range []struct {
		observation actualClaudeInterfaceCase
		result      processadapter.Result
		err         error
	}{
		{observation: positive, result: processadapter.Result{ExitCode: 2}},
		{observation: negative, result: processadapter.Result{ExitCode: 0}},
		{observation: positive, err: errors.New("runner failed")},
	} {
		if err := actualClaudeInterfaceOutcome(test.observation, test.result, test.err); err == nil {
			t.Fatalf("unsafe outcome accepted: %+v", test)
		}
	}
}

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
