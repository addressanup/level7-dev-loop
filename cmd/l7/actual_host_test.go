//go:build l7_actual_provider && darwin

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authorityadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/authority"
	configadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/config"
)

func TestActualHostProviderOrder(t *testing.T) {
	order := os.Getenv("L7_AUTHORIZE_ACTUAL_CLI_ORDER")
	if order != "codex-to-claude" && order != "claude-to-codex" {
		t.Skip("set one separately approved L7_AUTHORIZE_ACTUAL_CLI_ORDER; each provider order requires its own actual-host gate")
	}
	sourceRoot, err := filepath.EvalSymlinks(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	expectedCandidate := os.Getenv("L7_AUTHORIZE_ACTUAL_CLI_CANDIDATE")
	actualCandidate := strings.TrimSpace(cliGit(t, sourceRoot, "rev-parse", "HEAD"))
	if expectedCandidate == "" || expectedCandidate != actualCandidate || len(expectedCandidate) != 40 {
		t.Fatal("actual-host authorization must bind the exact full source candidate commit")
	}
	if status := strings.TrimSpace(cliGit(t, sourceRoot, "status", "--porcelain=v1", "--untracked-files=normal")); status != "" {
		t.Fatalf("actual-host source candidate is not clean: %q", status)
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		t.Fatal("actual-host provider execution requires an active terminal for immediate owner confirmation")
	}
	defer tty.Close()

	implementer, reviewer := "codex", "claude"
	if order == "claude-to-codex" {
		implementer, reviewer = reviewer, implementer
	}
	repository := cliRepository(t)
	if remotes := strings.TrimSpace(cliGit(t, repository, "remote")); remotes != "" {
		t.Fatalf("disposable actual-host repository unexpectedly has remotes: %q", remotes)
	}
	cliGit(t, repository, "config", "user.name", "Level Seven Actual Host")
	cliGit(t, repository, "config", "user.email", "actual-host@example.invalid")
	assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
	configureActualHost(t, repository, implementer, reviewer)
	commitAll(t, repository, "chore: adopt actual-host Level 7 fixture")
	assertRun(t, repository, []string{
		"brief", "--id", "actual-host-provider-order", "--tier", "3",
		"--problem", "Create actual-host.txt containing exactly: Level 7 actual-host validation.",
		"--scope", "actual-host.txt", "--accept", "The exact file exists and bounded verification passes.",
		"--risk", "A real provider session consumes credentials, network, time, and model budget.",
		"--rollback", "Delete the disposable no-remote repository.",
	}, 0, "L7-BRIEF-000")
	commitAll(t, repository, "docs(l7): add actual-host provider-order brief")

	terminal := authorityadapter.NewTerminal(tty, tty, true, "accountable-owner")
	stdout := executionRun(t, repository, []string{"run", "--agent", implementer, "--message", "test(cli): exercise actual host provider order", "--json"}, terminal, 0)
	if !strings.Contains(stdout, `"provider":"`+implementer+`"`) {
		t.Fatalf("implementer output=%s", stdout)
	}
	stdout = executionRun(t, repository, []string{"verify", "--json"}, authorityadapter.NewTerminal(nil, tty, false, "accountable-owner"), 0)
	if !strings.Contains(stdout, `"checks":[{"name":"actual-host-file"`) {
		t.Fatalf("verification output=%s", stdout)
	}
	stdout = executionRun(t, repository, []string{"review", "--agent", reviewer, "--json"}, authorityadapter.NewTerminal(nil, tty, false, "accountable-owner"), 0)
	if !strings.Contains(stdout, `"provider":"`+reviewer+`"`) || !strings.Contains(stdout, `"decision":"GO"`) {
		t.Fatalf("reviewer output=%s", stdout)
	}
	if remotes := strings.TrimSpace(cliGit(t, repository, "remote")); remotes != "" {
		t.Fatalf("actual-host provider order created a remote: %q", remotes)
	}
	assertBoundedRuntimeStateBeforeReadiness(t, repository)
}

func configureActualHost(t *testing.T, repository, implementer, reviewer string) {
	t.Helper()
	configuration, err := configadapter.Load(repository)
	if err != nil {
		t.Fatal(err)
	}
	configuration.Providers.Implementer = implementer
	configuration.Providers.Reviewer = reviewer
	expected := []byte("Level 7 actual-host validation.\n")
	if err := os.WriteFile(filepath.Join(repository, ".l7", "actual-host.expected"), expected, 0o600); err != nil {
		t.Fatal(err)
	}
	configuration.Verification = []configadapter.VerificationCommand{{
		Name: "actual-host-file", Argv: []string{"/usr/bin/cmp", ".l7/actual-host.expected", "actual-host.txt"}, Benchmark: false,
	}}
	configuration.Limits.MaxCommandOutputBytes = 1 << 20
	configuration.Limits.MaxCommandSeconds = 300
	data, err := json.MarshalIndent(configuration, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".l7", "config.json"), append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertBoundedRuntimeStateBeforeReadiness(t *testing.T, repository string) {
	t.Helper()
	directory := filepath.Join(repository, ".git", "l7", "product")
	for _, name := range []string{"active.json", "approval.json", "run.json", "verification.json", "review.json"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if len(data) > 256<<10 || bytes.Contains(data, []byte("transcript")) || bytes.Contains(data, []byte("reasoning")) {
			t.Fatalf("unsafe %s", name)
		}
	}
}
