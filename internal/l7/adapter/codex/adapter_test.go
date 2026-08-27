package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
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
			repository := t.TempDir()
			var invocation processadapter.Request
			var schemaPath string
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
				}
				invocation = request
				if len(request.Arguments) != 14 {
					t.Fatalf("arguments=%q", request.Arguments)
				}
				schemaPath = request.Arguments[10]
				inside, err := pathWithin(repository, schemaPath)
				if err != nil || inside {
					t.Fatalf("schemaPath=%q inside=%v error=%v", schemaPath, inside, err)
				}
				directoryInfo, err := os.Stat(filepath.Dir(schemaPath))
				if err != nil || !directoryInfo.IsDir() || directoryInfo.Mode().Perm() != 0o700 {
					t.Fatalf("schema directory info=%v error=%v", directoryInfo, err)
				}
				fileInfo, err := os.Lstat(schemaPath)
				if err != nil || !fileInfo.Mode().IsRegular() || fileInfo.Mode().Perm() != 0o600 {
					t.Fatalf("schema file info=%v error=%v", fileInfo, err)
				}
				data, err := os.ReadFile(schemaPath)
				if err != nil {
					t.Fatal(err)
				}
				wantSchema, err := provider.TerminalSchema(test.role)
				if err != nil || string(data) != wantSchema {
					t.Fatalf("schema=%q want=%q error=%v", data, wantSchema, err)
				}
				event := `{"type":"item.completed","item":{"id":"final","type":"agent_message","text":` + quoted(test.output) + `}}` + "\n"
				return processadapter.Result{ExitCode: 0, Stdout: []byte(event)}, nil
			}))
			response, err := adapter.Run(context.Background(), providerTask(test.role, repository), 1<<20, 30)
			if err != nil || response.Role != test.role || response.Identity.Provider != domain.ProviderCodex {
				t.Fatalf("Run()=%+v error=%v", response, err)
			}
			wantArguments := []string{
				"--ask-for-approval", "never",
				"exec",
				"--ephemeral",
				"--sandbox", test.sandbox,
				"--color", "never",
				"--json",
				"--output-schema", schemaPath,
				"--cd", repository,
				"-",
			}
			if !slices.Equal(invocation.Arguments, wantArguments) || strings.Contains(strings.Join(invocation.Arguments, " "), "dangerously") || strings.Contains(strings.Join(invocation.Arguments, " "), "skip-git-repo-check") || invocation.Directory != repository || !strings.Contains(string(invocation.Input), `"role": "`+string(test.role)+`"`) {
				t.Fatalf("invocation=%+v", invocation)
			}
			if _, err := os.Stat(schemaPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("schema file survived Run: %v", err)
			}
			if _, err := os.Stat(filepath.Dir(schemaPath)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("schema directory survived Run: %v", err)
			}
		})
	}
}

func TestProbeQualifiesOnlyExactTargetVersion(t *testing.T) {
	for _, test := range []struct {
		name       string
		version    string
		capability domain.CapabilityState
	}{
		{name: "target", version: CompatibleVersion + "\n", capability: domain.CapabilityAvailable},
		{name: "prior", version: "codex-cli 0.149.1\n", capability: domain.CapabilityDegraded},
		{name: "adjacent lower", version: "codex-cli 0.150.0\n", capability: domain.CapabilityDegraded},
		{name: "adjacent upper", version: "codex-cli 0.150.2\n", capability: domain.CapabilityDegraded},
		{name: "malformed", version: "Codex 0.150.1\n", capability: domain.CapabilityDegraded},
		{name: "padded", version: " codex-cli 0.150.1 \n", capability: domain.CapabilityDegraded},
		{name: "multiline", version: "codex-cli 0.150.1\nextra\n", capability: domain.CapabilityDegraded},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
				calls++
				return processadapter.Result{ExitCode: 0, Stdout: []byte(test.version)}, nil
			}))
			identity, err := adapter.Probe(context.Background())
			if err != nil || identity.Capability != test.capability || calls != 1 {
				t.Fatalf("Probe()=%+v calls=%d error=%v", identity, calls, err)
			}
		})
	}
}

func TestRunDoesNotInvokeDegradedVersion(t *testing.T) {
	calls := 0
	adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		calls++
		if calls > 1 {
			t.Fatal("degraded Codex version was invoked")
		}
		return processadapter.Result{ExitCode: 0, Stdout: []byte("codex-cli 0.149.1\n")}, nil
	}))
	response, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, t.TempDir()), 1<<20, 30)
	if err == nil || calls != 1 || response.Identity.Capability != domain.CapabilityDegraded {
		t.Fatalf("Run()=%+v calls=%d error=%v", response, calls, err)
	}
}

func TestRunRemovesSchemaAfterInvocationErrorAndCancellation(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "output error", err: processadapter.ErrOutputLimit},
		{name: "timeout", err: context.DeadlineExceeded},
		{name: "cancellation", err: context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := t.TempDir()
			var schemaPath string
			adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
				}
				schemaPath = request.Arguments[10]
				if _, err := os.Stat(schemaPath); err != nil {
					t.Fatalf("schema missing during invocation: %v", err)
				}
				return processadapter.Result{}, test.err
			}))
			if _, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, repository), 1<<20, 30); !errors.Is(err, test.err) {
				t.Fatalf("Run() error=%v want=%v", err, test.err)
			}
			if schemaPath == "" {
				t.Fatal("provider invocation did not receive a schema path")
			}
			if _, err := os.Stat(filepath.Dir(schemaPath)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("schema directory survived failed Run: %v", err)
			}
		})
	}
}

func TestRunFailsClosedForSchemaStorageErrors(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		creationErr := errors.New("schema storage unavailable")
		calls := 0
		adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
			calls++
			return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
		}))
		adapter.mkdirTemp = func(string, string) (string, error) { return "", creationErr }
		if _, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, t.TempDir()), 1<<20, 30); !errors.Is(err, creationErr) || calls != 1 {
			t.Fatalf("Run() calls=%d error=%v", calls, err)
		}
	})

	t.Run("repository local temp directory", func(t *testing.T) {
		repository := t.TempDir()
		temporaryRoot := filepath.Join(repository, "tmp")
		if err := os.Mkdir(temporaryRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		calls := 0
		adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(context.Context, processadapter.Request) (processadapter.Result, error) {
			calls++
			return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
		}))
		adapter.mkdirTemp = func(_ string, pattern string) (string, error) {
			return os.MkdirTemp(temporaryRoot, pattern)
		}
		if _, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, repository), 1<<20, 30); err == nil || !strings.Contains(err.Error(), "outside the repository") || calls != 1 {
			t.Fatalf("Run() calls=%d error=%v", calls, err)
		}
		entries, err := os.ReadDir(temporaryRoot)
		if err != nil || len(entries) != 0 {
			t.Fatalf("schema directory was not cleaned: entries=%v error=%v", entries, err)
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		cleanupErr := errors.New("cleanup failed")
		repository := t.TempDir()
		var schemaPath string
		adapter := NewWithRuntime(provider.NewRuntime(fakeResolve("codex"), func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
				return processadapter.Result{ExitCode: 0, Stdout: []byte(CompatibleVersion + "\n")}, nil
			}
			schemaPath = request.Arguments[10]
			event := `{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"Implemented.\",\"findings\":[]}"}}` + "\n"
			return processadapter.Result{ExitCode: 0, Stdout: []byte(event)}, nil
		}))
		adapter.removeAll = func(path string) error {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			return cleanupErr
		}
		if _, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, repository), 1<<20, 30); !errors.Is(err, cleanupErr) {
			t.Fatalf("Run() error=%v", err)
		}
		if _, err := os.Stat(filepath.Dir(schemaPath)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("test cleanup did not remove schema directory: %v", err)
		}
	})
}

func TestRunRejectsMalformedOrFailedEvents(t *testing.T) {
	outputs := []string{
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"},"unknown":true}` + "\n",
		`{"type":"turn.failed","error":{"message":"failed"}}` + "\n",
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"}}` + "\n" + `{"type":"item.completed","item":{"id":"two","type":"agent_message","text":"{}"}}` + "\n",
		"\n" + `{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"}}` + "\n",
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"}}` + "\n\n",
		`{"type":"item.completed","item":{"id":"one","type":"agent_message","text":"{}"}}` + "\nprose",
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
	response, err := adapter.Run(context.Background(), providerTask(domain.RoleImplementer, t.TempDir()), 1<<20, 30)
	if err == nil || response.Identity.Capability != domain.CapabilityUnavailable {
		t.Fatalf("Run()=%+v error=%v", response, err)
	}
}

func FuzzParseEvents(f *testing.F) {
	f.Add([]byte(`{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"Implemented.\",\"findings\":[]}"}}`+"\n"), false)
	f.Add([]byte(`{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"No blocker.\",\"findings\":[],\"decision\":\"GO\"}"}}`+"\n"), true)
	f.Add([]byte{0xff}, false)
	f.Fuzz(func(t *testing.T, data []byte, reviewer bool) {
		role := domain.RoleImplementer
		if reviewer {
			role = domain.RoleReviewer
		}
		response, err := parseEvents(data, role)
		if err == nil && (response.Role != role || (role == domain.RoleReviewer && !response.Decision.Valid()) || (role == domain.RoleImplementer && response.Decision != "")) {
			t.Fatalf("successful parse violated role contract: %+v", response)
		}
	})
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

func providerTask(role domain.ProviderRole, root string) domain.ProviderTask {
	return domain.ProviderTask{
		Role: role, Provider: domain.ProviderCodex, RepositoryRoot: root, ChangeID: "change", Tier: domain.TierHighRisk,
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
