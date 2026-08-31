package codexapp

import (
	"encoding/json"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestParseTerminalSeparatesImplementerAndReviewer(t *testing.T) {
	implementer, err := parseTerminal(`{"outcome":"complete","summary":"implemented"}`, false)
	if err != nil || implementer.Decision != "" {
		t.Fatalf("implementer=%+v err=%v", implementer, err)
	}
	reviewer, err := parseTerminal(`{"outcome":"complete","summary":"reviewed","decision":"GO"}`, true)
	if err != nil || reviewer.Decision != domain.DecisionGO {
		t.Fatalf("reviewer=%+v err=%v", reviewer, err)
	}
	if _, err := parseTerminal(`{"outcome":"complete","summary":"self review","decision":"GO"}`, false); err == nil {
		t.Fatal("implementer self-review passed")
	}
	if _, err := parseTerminal(`{"outcome":"blocked","summary":"blocked","decision":"GO"}`, true); err == nil {
		t.Fatal("blocked reviewer GO passed")
	}
	if _, err := parseTerminal(`{"outcome":"complete","summary":"reviewed","decision":"GO","unknown":true}`, true); err == nil {
		t.Fatal("unknown terminal field passed")
	}
}

func TestRateLimitResetUsesReportedWindowWithoutConsumingCredit(t *testing.T) {
	sent := map[string]any{}
	reset, err := rateLimitReset(func(value any) error {
		sent = value.(map[string]any)
		return nil
	}, func() (message, error) {
		return message{ID: json.RawMessage(`4`), Result: map[string]any{"rateLimits": map[string]any{"primary": map[string]any{"resetsAt": float64(1_800_000_000)}}}}, nil
	})
	if err != nil || reset == "" || sent["method"] != "account/rateLimits/read" {
		t.Fatalf("reset=%q sent=%v err=%v", reset, sent, err)
	}
}

func TestSandboxPolicyMatchesAppServerSchemaAndDeniesNetwork(t *testing.T) {
	implementation := sandboxPolicy("/worktree", false)
	if implementation["type"] != "workspaceWrite" || implementation["networkAccess"] != false || implementation["excludeSlashTmp"] != true || implementation["excludeTmpdirEnvVar"] != true || implementation["readOnlyAccess"] != nil {
		t.Fatalf("implementation policy=%#v", implementation)
	}
	roots, ok := implementation["writableRoots"].([]string)
	if !ok || len(roots) != 1 || roots[0] != "/worktree" {
		t.Fatalf("writable roots=%#v", implementation["writableRoots"])
	}
	reviewer := sandboxPolicy("/worktree", true)
	if reviewer["type"] != "readOnly" || reviewer["networkAccess"] != false || reviewer["writableRoots"] != nil {
		t.Fatalf("review policy=%#v", reviewer)
	}
}

func FuzzParseTerminal(f *testing.F) {
	f.Add(`{"outcome":"complete","summary":"ok"}`, false)
	f.Add(`{"outcome":"complete","summary":"ok","decision":"GO"}`, true)
	f.Fuzz(func(t *testing.T, value string, reviewer bool) {
		terminal, err := parseTerminal(value, reviewer)
		if err == nil && (terminal.Summary == "" || (reviewer && !terminal.Decision.Valid()) || (!reviewer && terminal.Decision != "")) {
			t.Fatalf("invalid successful terminal: %+v", terminal)
		}
	})
}
