package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authorityadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/authority"
	configadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/config"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	stateadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
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
			configureFakeVerification(t, repository, order.implementer, order.reviewer)
			commitAll(t, repository, "chore: adopt Level 7")
			cliGit(t, repository, "branch", "release-target")
			initialRefs := fakeRefNames(t, repository)
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
			assertFakeProviderBinding(t, stdout, fakeDirectory, order.implementer)
			if !strings.Contains(prompt.String(), "provider-change") {
				t.Fatalf("run stdout=%s prompt=%q", stdout, prompt.String())
			}
			candidate := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD"))
			candidateTree := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD^{tree}"))
			productData, err := os.ReadFile(filepath.Join(repository, "internal", "product", "change.go"))
			if err != nil || string(productData) != "package product\n" || strings.TrimSpace(cliGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")) != "" {
				t.Fatalf("candidate=%s tree=%s product=%q error=%v", candidate, candidateTree, productData, err)
			}
			assertRun(t, repository, []string{"status"}, 2, `next="run l7 verify"`)

			stdout = executionRun(t, repository, []string{"verify", "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			if !strings.Contains(stdout, `"state":"awaiting-independent-audit"`) || !strings.Contains(stdout, `"checks":[{"name":"test"`) {
				t.Fatalf("verify stdout=%s", stdout)
			}
			assertRun(t, repository, []string{"status"}, 2, "L7-REVIEW-002")

			stdout = executionRun(t, repository, []string{"review", "--agent", order.reviewer, "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			assertFakeProviderBinding(t, stdout, fakeDirectory, order.reviewer)
			if !strings.Contains(stdout, `"state":"reviewed"`) || !strings.Contains(stdout, `"decision":"GO"`) || !strings.Contains(stdout, `"candidate_commit":"`+candidate+`"`) || !strings.Contains(stdout, `"candidate_tree":"`+candidateTree+`"`) {
				t.Fatalf("review stdout=%s", stdout)
			}
			if after, err := os.ReadFile(filepath.Join(repository, "internal", "product", "change.go")); err != nil || !bytes.Equal(after, productData) {
				t.Fatalf("reviewer changed product bytes: data=%q error=%v", after, err)
			}
			assertRun(t, repository, []string{"status"}, 0, `state="reviewed"`)

			stdout = executionRun(t, repository, []string{"ready", "--json"}, authorityadapter.NewTerminal(nil, &prompt, false, "accountable-owner"), 0)
			if !strings.Contains(stdout, `"state":"ready"`) || !strings.Contains(stdout, `"configuration_digest"`) || !strings.Contains(stdout, `"benchmark":true`) {
				t.Fatalf("ready stdout=%s", stdout)
			}
			readyCandidate := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD"))
			prompt.Reset()
			terminal = authorityadapter.NewTerminal(strings.NewReader(readyCandidate+"\n"), &prompt, true, "accountable-owner")
			stdout = executionRun(t, repository, []string{"merge", "--target", "release-target"}, terminal, 0)
			if !strings.Contains(stdout, `state="merged"`) || !strings.Contains(stdout, `merge_target_ref="refs/heads/release-target"`) || !strings.Contains(prompt.String(), readyCandidate) {
				t.Fatalf("merge stdout=%s prompt=%q", stdout, prompt.String())
			}
			if target := strings.TrimSpace(cliGit(t, repository, "show-ref", "--hash", "refs/heads/release-target")); target != readyCandidate {
				t.Fatalf("release-target=%s want=%s", target, readyCandidate)
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
			if refs := fakeRefNames(t, repository); strings.Join(refs, "\x00") != strings.Join(initialRefs, "\x00") {
				t.Fatalf("refs=%v want names=%v", refs, initialRefs)
			}
			if remotes := strings.TrimSpace(cliGit(t, repository, "remote")); remotes != "" {
				t.Fatalf("fake provider created remotes: %q", remotes)
			}
			if status := strings.TrimSpace(cliGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")); status != "" {
				t.Fatalf("fake lifecycle left repository dirty: %q", status)
			}
			assertBoundedRuntimeState(t, repository)
		})
	}
}

func TestFakeProviderCancellationLeavesNoAcceptedEvidenceAndReleasesLock(t *testing.T) {
	for _, provider := range []string{"codex", "claude"} {
		t.Run(provider, func(t *testing.T) {
			repository := cliRepository(t)
			fakeDirectory := installFakeProviders(t)
			t.Setenv("PATH", fakeDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
			temporaryDirectory, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("TMPDIR", temporaryDirectory)
			block := filepath.Join(temporaryDirectory, "l7-fake-provider-block-"+provider)
			started := filepath.Join(temporaryDirectory, "l7-fake-provider-started-"+provider)
			if err := os.WriteFile(block, []byte("block\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			cliGit(t, repository, "config", "user.name", "Level Seven")
			cliGit(t, repository, "config", "user.email", "l7@example.invalid")
			assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
			reviewer := "claude"
			if provider == "claude" {
				reviewer = "codex"
			}
			configureFakeVerification(t, repository, provider, reviewer)
			commitAll(t, repository, "chore: adopt Level 7")
			assertRun(t, repository, []string{
				"brief", "--id", "provider-cancellation", "--tier", "2", "--problem", "Exercise fake provider cancellation.",
				"--scope", "internal/product/**", "--accept", "Cancellation creates no accepted evidence.", "--risk", "A child could outlive cancellation.", "--rollback", "Delete the disposable fixture.",
			}, 0, "L7-BRIEF-000")
			commitAll(t, repository, "docs(product): add cancellation brief")
			baselineHead := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD"))
			baselineRefs := fakeRefNames(t, repository)

			ctx, cancel := context.WithCancel(context.Background())
			type commandResult struct {
				exit           int
				stdout, stderr string
			}
			done := make(chan commandResult, 1)
			go func() {
				var stdout, stderr bytes.Buffer
				exit := runAtWithTerminal(ctx, []string{"run", "--agent", provider, "--message", "feat(product): exercise cancellation", "--json"}, repository, &stdout, &stderr, authorityadapter.NewTerminal(nil, &stderr, false, "accountable-owner"))
				done <- commandResult{exit: exit, stdout: stdout.String(), stderr: stderr.String()}
			}()

			deadline := time.NewTimer(5 * time.Second)
			ticker := time.NewTicker(10 * time.Millisecond)
			providerStarted := false
			var early *commandResult
			for !providerStarted && early == nil {
				select {
				case result := <-done:
					early = &result
				case <-ticker.C:
					_, statErr := os.Stat(started)
					providerStarted = statErr == nil
				case <-deadline.C:
					providerStarted = false
					early = &commandResult{exit: -1, stderr: "timed out waiting for fake provider start"}
				}
			}
			ticker.Stop()
			if !deadline.Stop() {
				select {
				case <-deadline.C:
				default:
				}
			}
			cancel()
			if early != nil {
				if early.exit == -1 {
					select {
					case <-done:
					case <-time.After(5 * time.Second):
					}
				}
				t.Fatalf("provider did not reach cancellable state: %+v", *early)
			}

			var result commandResult
			select {
			case result = <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("cancelled fake provider did not return within the supervisor bound")
			}
			if result.exit != 130 || result.stderr != "" || !strings.Contains(result.stdout, `"outcome":"CANCELLED"`) || !strings.Contains(result.stdout, `"code":"L7-CLI-003"`) {
				t.Fatalf("cancelled result=%+v", result)
			}
			if head := strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD")); head != baselineHead {
				t.Fatalf("cancelled provider changed HEAD: got=%s want=%s", head, baselineHead)
			}
			if status := strings.TrimSpace(cliGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")); status != "" {
				t.Fatalf("cancelled fake provider changed repository: %q", status)
			}
			if refs := fakeRefNames(t, repository); strings.Join(refs, "\x00") != strings.Join(baselineRefs, "\x00") {
				t.Fatalf("cancelled provider changed refs: got=%v want=%v", refs, baselineRefs)
			}
			common := strings.TrimSpace(cliGit(t, repository, "rev-parse", "--git-common-dir"))
			if !filepath.IsAbs(common) {
				common = filepath.Join(repository, common)
			}
			for _, name := range []string{"run.json", "verification.json", "review.json", "readiness.json", "merge.json"} {
				_, statErr := os.Stat(filepath.Join(common, "l7", "product", name))
				if !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("cancelled provider created accepted %s: %v", name, statErr)
				}
			}
			lock, err := stateadapter.Acquire(common)
			if err != nil {
				t.Fatalf("cancelled provider retained mutation lock: %v", err)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			startedData, err := os.ReadFile(started)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(750 * time.Millisecond)
			if after, err := os.ReadFile(started); err != nil || !bytes.Equal(after, startedData) || strings.TrimSpace(cliGit(t, repository, "rev-parse", "HEAD")) != baselineHead || strings.TrimSpace(cliGit(t, repository, "status", "--porcelain=v1", "--untracked-files=all")) != "" {
				t.Fatalf("late mutation after cancellation: marker=%q error=%v", after, err)
			}
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

func configureFakeVerification(t *testing.T, repository, implementer, reviewer string) {
	t.Helper()
	configuration, err := configadapter.Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	configuration.Verification = []configadapter.VerificationCommand{{Name: "test", Argv: []string{"fake-verify"}, Benchmark: true}}
	configuration.Providers.Implementer = implementer
	configuration.Providers.Reviewer = reviewer
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
  printf '%s\n' 'codex-cli 0.150.1'
  exit 0
fi
if [ -f "$TMPDIR/l7-fake-provider-block-codex" ]; then
  printf '%s\n' 'started' > "$TMPDIR/l7-fake-provider-started-codex"
  /bin/sleep 300
  exit 1
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
  printf '%s\n' '2.1.247 (Claude Code)'
  exit 0
fi
if [ -f "$TMPDIR/l7-fake-provider-block-claude" ]; then
  printf '%s\n' 'started' > "$TMPDIR/l7-fake-provider-started-claude"
  /bin/sleep 300
  exit 1
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
  printf '%s\n' '{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"Implemented.","findings":[]}}'
else
  printf '%s\n' '{"type":"result","subtype":"success","is_error":false,"permission_denials":[],"structured_output":{"schema":1,"outcome":"complete","summary":"No blocker.","findings":[],"decision":"GO"}}'
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

func assertFakeProviderBinding(t *testing.T, output, directory, provider string) {
	t.Helper()
	executable, err := processadapter.Resolve(filepath.Join(directory, provider))
	if err != nil {
		t.Fatal(err)
	}
	version := "codex-cli 0.150.1"
	if provider == "claude" {
		version = "2.1.247 (Claude Code)"
	}
	for _, binding := range []string{
		`"provider":"` + provider + `"`,
		`"executable":"` + executable.Path + `"`,
		`"provider_version":"` + version + `"`,
		`"executable_digest":"` + executable.Digest + `"`,
	} {
		if !strings.Contains(output, binding) {
			t.Fatalf("provider output=%s missing binding=%q", output, binding)
		}
	}
}

func fakeRefNames(t *testing.T, repository string) []string {
	t.Helper()
	output := strings.TrimSpace(cliGit(t, repository, "for-each-ref", "--format=%(refname)"))
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
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
