package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	stateadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
)

func TestRunCommandContract(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		exit     int
		contains []string
	}{
		{"default help", nil, 0, []string{"PASS", `command="help"`, "Usage: l7"}},
		{"help flag", []string{"--help"}, 0, []string{"PASS", `command="help"`}},
		{"version flag", []string{"--version"}, 0, []string{"PASS", `command="version"`, `version="test-version"`}},
		{"status default off", []string{"status"}, 2, []string{"BLOCKED", "L7-FLAG-001", `state="disabled"`}},
		{"unknown command", []string{"unknown"}, 1, []string{"FAILED", "L7-CLI-001", `command="unknown"`}},
		{"unknown flag", []string{"status", "--unsafe"}, 1, []string{"FAILED", "unknown flag"}},
		{"json after unknown flag", []string{"--unsafe", "--json"}, 1, []string{`"outcome":"FAILED"`, `"message":"unknown flag"`}},
		{"extra command", []string{"help", "status"}, 1, []string{"FAILED", "command does not accept options"}},
		{"too many arguments", make([]string, maxArguments+1), 1, []string{"FAILED", "too many arguments"}},
		{"oversized argument", []string{strings.Repeat("x", maxArgumentBytes+1)}, 1, []string{"FAILED", "argument exceeds size limit"}},
		{"duplicate json", []string{"--json", "status", "--json"}, 1, []string{`"outcome":"FAILED"`, `"message":"duplicate --json flag"`}},
		{"json status", []string{"status", "--json"}, 2, []string{`{"schema":4`, `"outcome":"BLOCKED"`, `"details":[]`, `"repository":{`}},
		{"adopt unknown option", []string{"adopt", "--unsafe"}, 1, []string{"FAILED", "unknown adopt option"}},
		{"brief invalid tier", []string{"brief", "--id", "x", "--tier", "4"}, 1, []string{"FAILED", "risk tier must be 1, 2, or 3"}},
		{"brief missing value", []string{"brief", "--scope"}, 1, []string{"FAILED", "missing its value"}},
		{"run missing options", []string{"run"}, 1, []string{"FAILED", "run requires --agent"}},
		{"run invalid provider", []string{"run", "--agent", "other", "--message", "feat: change"}, 1, []string{"FAILED", "run requires --agent"}},
		{"run missing value", []string{"run", "--agent"}, 1, []string{"FAILED", "missing its value"}},
		{"verify option", []string{"verify", "--unsafe"}, 1, []string{"FAILED", "unknown flag"}},
		{"review missing provider", []string{"review"}, 1, []string{"FAILED", "review requires --agent"}},
	}
	previous := version
	version = "test-version"
	t.Cleanup(func() { version = previous })
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run(context.Background(), test.args, &stdout, &stderr)
			if exit != test.exit {
				t.Fatalf("exit=%d, want %d; stdout=%q stderr=%q", exit, test.exit, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
			for _, substring := range test.contains {
				if !strings.Contains(stdout.String(), substring) {
					t.Fatalf("stdout=%q, want substring %q", stdout.String(), substring)
				}
			}
		})
	}
}

func TestTierOneEndToEndCreatesNoGovernanceArtifact(t *testing.T) {
	repository := cliRepository(t)
	assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
	commitAll(t, repository, "adopt Level 7")
	assertRun(t, repository, []string{
		"brief", "--id", "routine-fix", "--tier", "1", "--problem", "Fix a low-risk defect.", "--scope", "internal/example/**",
	}, 0, "L7-BRIEF-000")
	if _, err := os.Stat(filepath.Join(repository, "docs", "artifacts", "changes")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Tier 1 created governance artifact directory: %v", err)
	}
	assertRun(t, repository, []string{"status", "--json"}, 0, `"state":"planned"`)
	path := filepath.Join(repository, "internal", "example", "change.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertRun(t, repository, []string{"status"}, 2, `state="building"`)
}

func TestTierTwoEndToEndCreatesOneBriefAndRecoversAfterInterruption(t *testing.T) {
	repository := cliRepository(t)
	assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
	commitAll(t, repository, "adopt Level 7")
	arguments := tierTwoArguments()
	assertRun(t, repository, arguments, 0, "L7-BRIEF-000")
	briefPath := filepath.Join(repository, "docs", "artifacts", "changes", "product-feature.md")
	if info, err := os.Lstat(briefPath); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("Tier 2 brief info=%v error=%v", info, err)
	}
	matches, err := filepath.Glob(filepath.Join(repository, "docs", "artifacts", "changes", "*"))
	if err != nil || len(matches) != 1 || matches[0] != briefPath {
		t.Fatalf("Tier 2 artifacts=%v error=%v", matches, err)
	}
	activePath := filepath.Join(repository, ".git", "l7", "product", "active.json")
	if err := os.Remove(activePath); err != nil {
		t.Fatal(err)
	}
	assertRun(t, repository, arguments, 0, "L7-BRIEF-000")
	assertRun(t, repository, []string{"status", "--json"}, 0, `"state":"planned"`)
	data, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"base"`, `"head"`, `"tree"`, `"status"`} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("active pointer contains redundant %s: %s", forbidden, data)
		}
	}
}

func TestTierThreeEndToEndWaitsForOwnerAndRejectsPreapprovalWork(t *testing.T) {
	repository := cliRepository(t)
	assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
	commitAll(t, repository, "adopt Level 7")
	arguments := []string{
		"brief", "--id", "security-change", "--tier", "3", "--problem", "Change a protected control.",
		"--scope", "Makefile", "--accept", "Protection remains effective.", "--risk", "A bypass could be introduced.", "--rollback", "Revert the candidate.",
	}
	assertRun(t, repository, arguments, 0, "L7-BRIEF-000")
	assertRun(t, repository, []string{"status"}, 2, "L7-AUTH-002")
	if err := os.WriteFile(filepath.Join(repository, "Makefile"), []byte("all:\n\t@true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertRun(t, repository, []string{"status"}, 2, "L7-AUTH-001")
}

func TestBriefFailsClosedWhileRepositoryLockIsHeld(t *testing.T) {
	repository := cliRepository(t)
	assertRun(t, repository, []string{"adopt", "--enable-local-lifecycle"}, 0, "L7-ADOPT-000")
	commitAll(t, repository, "adopt Level 7")
	common := strings.TrimSpace(cliGit(t, repository, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(repository, common)
	}
	lock, err := stateadapter.Acquire(common)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	assertRun(t, repository, tierTwoArguments(), 2, "L7-STATE-002")
}

func assertRun(t *testing.T, repository string, arguments []string, wantExit int, contains string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exit := runAt(context.Background(), arguments, repository, &stdout, &stderr)
	if exit != wantExit || stderr.Len() != 0 || !strings.Contains(stdout.String(), contains) {
		t.Fatalf("runAt(%v) exit=%d want=%d stdout=%q stderr=%q contains=%q", arguments, exit, wantExit, stdout.String(), stderr.String(), contains)
	}
	return stdout.String()
}

func tierTwoArguments() []string {
	return []string{
		"brief", "--id", "product-feature", "--tier", "2", "--problem", "Add a bounded product feature.",
		"--scope", "internal/product/**", "--accept", "Relevant tests pass.", "--risk", "State could become stale.", "--rollback", "Revert the candidate.",
	}
}

func cliRepository(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cliGit(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	commitAll(t, root, "initial")
	return root
}

func commitAll(t *testing.T, repository, message string) {
	t.Helper()
	cliGit(t, repository, "add", "-A")
	cliGit(t, repository, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", message)
}

func cliGit(t *testing.T, repository string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repository}, arguments...)...)
	command.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return string(output)
}

func TestRunHonorsCancellationBeforeExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	if exit := run(ctx, []string{"status"}, &stdout, &stderr); exit != 130 || !strings.HasPrefix(stdout.String(), "CANCELLED ") || stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestRunOutputDoesNotDependOnWriterType(t *testing.T) {
	var buffer bytes.Buffer
	var builder strings.Builder
	if exit := run(context.Background(), []string{"version"}, &buffer, ioDiscard{}); exit != 0 {
		t.Fatalf("buffer exit=%d", exit)
	}
	if exit := run(context.Background(), []string{"version"}, &builder, ioDiscard{}); exit != 0 {
		t.Fatalf("builder exit=%d", exit)
	}
	if buffer.String() != builder.String() {
		t.Fatalf("writer-dependent output: buffer=%q builder=%q", buffer.String(), builder.String())
	}
}

func TestRunReportsOutputFailureOnStderr(t *testing.T) {
	for _, writer := range []interface{ Write([]byte) (int, error) }{errorWriter{}, shortWriter{}} {
		var stderr bytes.Buffer
		if exit := run(context.Background(), []string{"version"}, writer, &stderr); exit != 1 || !strings.Contains(stderr.String(), "L7-CLI-002") {
			t.Fatalf("writer=%T exit=%d stderr=%q", writer, exit, stderr.String())
		}
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

type shortWriter struct{}

func (shortWriter) Write(data []byte) (int, error) { return len(data) / 2, nil }

type ioDiscard struct{}

func (ioDiscard) Write(data []byte) (int, error) { return len(data), nil }
