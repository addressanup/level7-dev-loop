package provider

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const fuzzTerminalMaxBytes = 64 << 10

func TestProbeDistinguishesInstalledFromCompatible(t *testing.T) {
	runtime := NewRuntime(
		func(string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/codex", Digest: strings.Repeat("a", 64)}, nil
		},
		func(context.Context, processadapter.Request) (processadapter.Result, error) {
			return processadapter.Result{ExitCode: 0, Stdout: []byte("codex-cli 0.149.1\n")}, nil
		},
	)
	identity, err := runtime.Probe(context.Background(), "codex", domain.ProviderCodex, []string{"--version"}, func(version string) bool { return version == "codex-cli 0.149.1" })
	if err != nil || identity.Capability != domain.CapabilityAvailable || identity.Version != "codex-cli 0.149.1" {
		t.Fatalf("Probe()=%+v error=%v", identity, err)
	}
	identity, err = runtime.Probe(context.Background(), "codex", domain.ProviderCodex, []string{"--version"}, func(string) bool { return false })
	if err != nil || identity.Capability != domain.CapabilityDegraded {
		t.Fatalf("degraded Probe()=%+v error=%v", identity, err)
	}
}

func TestInvokeRejectsExecutableReplacement(t *testing.T) {
	runtime := NewRuntime(
		func(string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/codex", Digest: strings.Repeat("b", 64)}, nil
		},
		func(context.Context, processadapter.Request) (processadapter.Result, error) {
			t.Fatal("run called after executable replacement")
			return processadapter.Result{}, nil
		},
	)
	identity := domain.ProviderIdentity{Provider: domain.ProviderCodex, Executable: "/usr/bin/codex", Digest: strings.Repeat("a", 64), Capability: domain.CapabilityAvailable}
	if _, err := runtime.Invoke(context.Background(), identity, "/repo", []string{"exec"}, []byte("task"), 1<<20, 30); err == nil || !strings.Contains(err.Error(), "identity changed") {
		t.Fatalf("Invoke() error=%v", err)
	}
}

func TestRenderAndParseProviderNeutralProtocol(t *testing.T) {
	task := domain.ProviderTask{Role: domain.RoleReviewer, Provider: domain.ProviderClaude, RepositoryRoot: "/repo", ChangeID: "change", Tier: domain.TierHighRisk, Base: strings.Repeat("a", 40), Candidate: domain.CandidateIdentity{Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40)}, Problem: "Review the change.", Scope: []string{"internal/example/**"}}
	prompt, err := RenderTask(task)
	if err != nil || !strings.Contains(string(prompt), `"role": "reviewer"`) || !strings.Contains(string(prompt), `"commit": "`+strings.Repeat("b", 40)+`"`) || !strings.Contains(string(prompt), "read-only audit") {
		t.Fatalf("RenderTask()=%q error=%v", prompt, err)
	}
	response, err := ParseTerminal([]byte(`{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`), domain.RoleReviewer)
	if err != nil || response.Decision != domain.DecisionGO || response.Role != domain.RoleReviewer {
		t.Fatalf("ParseTerminal()=%+v error=%v", response, err)
	}
	for _, data := range [][]byte{
		[]byte(`{"schema":1,"schema":1}`),
		[]byte(`{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[],"decision":"GO"}`),
		[]byte(`{"schema":1,"outcome":"blocked","summary":"Blocked.","findings":[],"decision":"GO"}`),
		append([]byte(`{"schema":1,"outcome":"complete","summary":"`), 0xff),
	} {
		role := domain.RoleReviewer
		if strings.Contains(string(data), "Implemented") {
			role = domain.RoleImplementer
		}
		if _, err := ParseTerminal(data, role); err == nil {
			t.Fatalf("ParseTerminal(%s) unexpectedly passed", data)
		}
	}
}

func TestProbePropagatesUnavailableExecutable(t *testing.T) {
	runtime := NewRuntime(func(string) (processadapter.Executable, error) {
		return processadapter.Executable{}, errors.New("missing")
	}, nil)
	identity, err := runtime.Probe(context.Background(), "codex", domain.ProviderCodex, []string{"--version"}, func(string) bool { return true })
	if err == nil || identity.Capability != domain.CapabilityUnavailable {
		t.Fatalf("Probe()=%+v error=%v", identity, err)
	}
}

func TestParseTerminalEnforcesProductionSizeBoundary(t *testing.T) {
	pad := func(data []byte, size int) []byte {
		t.Helper()
		if len(data) > size {
			t.Fatalf("fixture length %d exceeds boundary %d", len(data), size)
		}
		result := append([]byte(nil), data...)
		return append(result, bytes.Repeat([]byte(" "), size-len(result))...)
	}

	valid := []byte(`{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`)
	atLimit := pad(valid, MaxProviderPrompt)
	response, err := ParseTerminal(atLimit, domain.RoleReviewer)
	if err != nil || response.Decision != domain.DecisionGO {
		t.Fatalf("valid provider response at %d-byte boundary was rejected: response=%+v error=%v", MaxProviderPrompt, response, err)
	}

	invalidAtLimit := pad([]byte(`{"schema":1,"schema":1}`), MaxProviderPrompt)
	if _, err := ParseTerminal(invalidAtLimit, domain.RoleReviewer); err == nil {
		t.Fatalf("strict-invalid provider response at %d-byte boundary was accepted", MaxProviderPrompt)
	}

	aboveLimit := append(append([]byte(nil), atLimit...), ' ')
	if _, err := ParseTerminal(aboveLimit, domain.RoleReviewer); err == nil {
		t.Fatalf("provider response above %d-byte boundary was accepted", MaxProviderPrompt)
	}
}

func FuzzParseTerminal(f *testing.F) {
	f.Add([]byte(`{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}`), false)
	f.Add([]byte(`{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`), true)
	f.Add([]byte{0xff}, true)
	f.Fuzz(func(t *testing.T, data []byte, reviewer bool) {
		if len(data) > fuzzTerminalMaxBytes {
			t.Skip()
		}
		role := domain.RoleImplementer
		if reviewer {
			role = domain.RoleReviewer
		}
		response, err := ParseTerminal(data, role)
		if err == nil && (response.Role != role || (role == domain.RoleReviewer && !response.Decision.Valid()) || (role == domain.RoleImplementer && response.Decision != "")) {
			t.Fatalf("successful parse violated role contract: %+v", response)
		}
	})
}

func BenchmarkTerminalProtocolDecode(b *testing.B) {
	data := []byte(`{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}`)
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := ParseTerminal(data, domain.RoleReviewer); err != nil {
			b.Fatal(err)
		}
	}
}
