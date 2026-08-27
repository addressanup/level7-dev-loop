package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authorityadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/authority"
	configadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/config"
)

func TestFakeProviderEndToEndInBothOrders(t *testing.T) {
	for _, order := range []struct {
		implementer string
		reviewer    string
	}{
		{implementer: "codex", reviewer: "claude"},
		{implementer: "claude", reviewer: "codex"},
	} {
		t.Run(order.implementer+"-to-"+order.reviewer, func(t *testing.T) {
			repository := cliRepository(t)
			fakeDirectory := installFakeProviders(t)
			t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
			cliGit(t, repository, "config", "user.name", "Level Seven")
			cliGit(t, repository, "config", "user.email", "l7@example.invalid")

			assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
			configureFakeVerification(t, repository)
			commitAll(t, repository, "chore: adopt Level 7")
			cliGit(t, repository, "branch", "release-target")
			brief := []string{
				"brief", "--id", "provider-change", "--tier", "3", "--problem", "Implement a fake-provider change.",
				"--scope", "internal/product/**", "--accept", "Fake verification passes.", "--risk", "Provider behavior could drift.", "--rollback", "Revert the candidate.",
			}
			assertRun(t, repository, brief, 0, "L7-BRIEF-000")
			commitAll(t, repository, "docs(product): add provider change brief")
			assertRun(t, repository, []string{"status"}, 2, "L7-AUTH-002")

			var prompt bytes.Buffer
			terminal := authorityadapter.NewTerminal(strings.NewReader("provider-change\n"), &prompt, true, "accountable-owner")
			stdout := executionRun(t, repository, []string{"run", "--agent", order.implementer, "--message", "feat(product): implement fake change", "--json"}, terminal, 0)
			if !strings.Contains(stdout, `"provider":"`+order.implementer+`"`) || !strings.Contains(prompt.String(), "provider-change") {
				t.Fatalf("run stdout=%s prompt=%q", stdout, prompt.String())
			}
			assertRun(t, repository, []string{"status"}, 2, `next="run l7 verify"`)

			stdout = executionRun(t, repository, []string{"verify", "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			if !strings.Contains(stdout, `"state":"awaiting-independent-audit"`) || !strings.Contains(stdout, `"checks":[{"name":"test"`) {
				t.Fatalf("verify stdout=%s", stdout)
			}
			assertRun(t, repository, []string{"status"}, 2, "L7-REVIEW-002")

			stdout = executionRun(t, repository, []string{"review", "--agent", order.reviewer, "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			if !strings.Contains(stdout, `"state":"reviewed"`) || !strings.Contains(stdout, `"provider":"`+order.reviewer+`"`) || !strings.Contains(stdout, `"decision":"GO"`) {
				t.Fatalf("review stdout=%s", stdout)
			}
			assertRun(t, repository, []string{"status"}, 0, `state="reviewed"`)

			stdout = executionRun(t, repository, []string{"ready", "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			if !strings.Contains(stdout, `"state":"ready"`) || !strings.Contains(stdout, `"configuration_digest"`) || !strings.Contains(stdout, `"benchmark":true`) {
				t.Fatalf("ready stdout=%s", stdout)
			}
			candidate := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD"))
			prompt.Reset()
			terminal = authorityadapter.NewTerminal(strings.NewReader(candidate+"\n"), &prompt, true, "accountable-owner")
			stdout = executionRun(t, repository, []string{"merge", "--target", "release-target"}, terminal, 0)
			if !strings.Contains(stdout, `state="merged"`) || !strings.Contains(stdout, `merge_target_ref="refs/heads/release-target"`) || !strings.Contains(prompt.String(), candidate) {
				t.Fatalf("merge stdout=%s prompt=%q", stdout, prompt.String())
			}
			if target := strings.TrimSpace(cliGit(t, repository, "show-ref", "--hash", "refs/heads/release-target")); target != candidate {
				t.Fatalf("release-target=%s want=%s", target, candidate)
			}
			assertRun(t, repository, []string{"status"}, 0, `state="merged"`)

			artifacts, err := filepath.Glob(filepath.Join(repository, "docs", "artifacts", "changes", "provider-change*.md"))
			if err != nil || len(artifacts) != 3 {
				t.Fatalf("artifacts=%v error=%v", artifacts, err)
			}
			log := cliGit(t, repository, "log", "-3", "--format=%s")
			for _, subject := range []string{"docs(l7): record provider-change audit", "test(l7): record provider-change verification", "feat(product): implement fake change"} {
				if !strings.Contains(log, subject) {
					t.Fatalf("log=%q missing %q", log, subject)
				}
			}
			assertBoundedRuntimeState(t, repository)
		})
	}
}

func executionRun(t *testing.T, repository string, arguments []string, terminal authorityadapter.Terminal, wantExit int) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exit := runAtWithTerminal(context.Background(), arguments, repository, &stdout, &stderr, terminal)
	if exit != wantExit || stderr.Len() != 0 {
		t.Fatalf("runAtWithTerminal(%v) exit=%d want=%d stdout=%q stderr=%q", arguments, exit, wantExit, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func configureFakeVerification(t *testing.T, repository string) {
	t.Helper()
	configuration, err := configadapter.Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	configuration.Verification = []configadapter.VerificationCommand{{Name: "test", Argv: []string{"fake-verify"}, Benchmark: true}}
	data, err := json.MarshalIndent(configuration, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(repository, ".l7", "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func installFakeProviders(t *testing.T) string {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	scripts := map[string]string{
		"codex": `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf '%s\n' 'codex-cli 0.149.1'
  exit 0
fi
review=0
for argument in "$@"; do
  if [ "$argument" = "read-only" ]; then review=1; fi
done
/bin/cat >/dev/null
if [ "$review" -eq 0 ]; then
  /bin/mkdir -p internal/product
  printf '%s\n' 'package product' > internal/product/change.go
  printf '%s\n' '{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"Implemented.\",\"findings\":[]}"}}'
else
  printf '%s\n' '{"type":"item.completed","item":{"id":"final","type":"agent_message","text":"{\"schema\":1,\"outcome\":\"complete\",\"summary\":\"No blocker.\",\"findings\":[],\"decision\":\"GO\"}"}}'
fi
`,
		"claude": `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf '%s\n' '2.1.241 (Claude Code)'
  exit 0
fi
review=0
previous=''
for argument in "$@"; do
  if [ "$previous" = "--permission-mode" ] && [ "$argument" = "plan" ]; then review=1; fi
  previous="$argument"
done
/bin/cat >/dev/null
if [ "$review" -eq 0 ]; then
  /bin/mkdir -p internal/product
  printf '%s\n' 'package product' > internal/product/change.go
  printf '%s\n' '{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}'
else
  printf '%s\n' '{"type":"result","subtype":"success","is_error":false,"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}'
fi
`,
		"fake-verify": "#!/bin/sh\nexit 0\n",
	}
	for name, content := range scripts {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func assertBoundedRuntimeState(t *testing.T, repository string) {
	t.Helper()
	directory := filepath.Join(repository, ".git", "l7", "product")
	for _, name := range []string{"active.json", "approval.json", "run.json", "verification.json", "review.json", "readiness.json", "merge.json"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if len(data) > 256<<10 || strings.Contains(string(data), "transcript") || strings.Contains(string(data), "reasoning") {
			t.Fatalf("unsafe %s: %q", name, data)
		}
	}
}
