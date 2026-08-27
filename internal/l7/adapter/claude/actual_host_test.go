//go:build l7_actual_provider

package claude

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestActualHostInterface(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_CLAUDE") != "interface" {
		t.Skip("set L7_AUTHORIZE_ACTUAL_CLAUDE=interface only under the separately approved no-model interface gate")
	}
	sourceRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err == nil {
		sourceRoot, err = filepath.EvalSymlinks(sourceRoot)
	}
	if err != nil {
		t.Fatal(err)
	}
	assertActualClaudeSourceIsolation(t, sourceRoot)
	expectedCandidate := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_CANDIDATE")
	expectedTree := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_TREE")
	assertActualClaudeSource(t, sourceRoot, expectedCandidate, expectedTree)
	if expected := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_GOOS"); expected != runtime.GOOS {
		t.Fatalf("actual-host GOOS=%q want=%q", runtime.GOOS, expected)
	}
	if expected := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_GOARCH"); expected != runtime.GOARCH {
		t.Fatalf("actual-host GOARCH=%q want=%q", runtime.GOARCH, expected)
	}

	identity, err := New().Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	expectedVersion := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_VERSION")
	expectedExecutable := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_EXECUTABLE")
	expectedDigest := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_DIGEST")
	versionOK := expectedVersion == CompatibleVersion || expectedVersion == CompatibleVersion+" (Claude Code)"
	if !versionOK || identity.Capability != domain.CapabilityAvailable || identity.Version != expectedVersion || identity.Executable != expectedExecutable || identity.Digest != expectedDigest || !filepath.IsAbs(expectedExecutable) || !actualHostHex(identity.Digest, 32) {
		t.Fatalf("actual-host Claude identity does not match the authorized available tuple: %+v", identity)
	}
	assertActualHostExecutable(t, identity.Executable, identity.Digest)

	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		invocationArguments, err := arguments(role)
		if err != nil {
			t.Fatal(err)
		}
		result, err := (processadapter.Runner{}).Run(context.Background(), processadapter.Request{
			Executable:     identity.Executable,
			Arguments:      append(invocationArguments, "--help"),
			Directory:      sourceRoot,
			Environment:    processadapter.MinimalEnvironment(),
			MaxOutputBytes: 1 << 20,
			Timeout:        30 * time.Second,
		})
		if err != nil || result.ExitCode != 0 {
			t.Fatalf("Claude %s no-model argument/help interface failed: exit=%d error=%v", role, result.ExitCode, err)
		}
		help := string(append(append([]byte{}, result.Stdout...), result.Stderr...))
		for _, flag := range []string{"--safe-mode", "--disable-slash-commands", "--print", "--input-format", "--max-turns", "--tools", "--disallowedTools", "--permission-mode", "--strict-mcp-config", "--no-chrome", "--no-session-persistence", "--output-format", "--json-schema"} {
			if !strings.Contains(help, flag) {
				t.Fatalf("Claude %s help surface does not advertise required flag %s", role, flag)
			}
		}
	}
	assertActualHostExecutable(t, identity.Executable, identity.Digest)
	assertActualClaudeSource(t, sourceRoot, expectedCandidate, expectedTree)
	t.Logf("provider=%s version=%q executable=%q digest=%s capability=%s", identity.Provider, identity.Version, identity.Executable, identity.Digest, identity.Capability)
}

func assertActualClaudeSourceIsolation(t *testing.T, sourceRoot string) {
	t.Helper()
	authorizedRoot := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_SOURCE_ROOT")
	temporaryParent := requiredActualClaudeEnvironment(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_TEMP_PARENT")
	if !filepath.IsAbs(authorizedRoot) || !filepath.IsAbs(temporaryParent) {
		t.Fatal("actual-host authorization must bind absolute source and temporary-parent paths")
	}
	var err error
	authorizedRoot, err = filepath.EvalSymlinks(authorizedRoot)
	if err != nil || authorizedRoot != sourceRoot {
		t.Fatalf("authorized source root=%q actual=%q error=%v", authorizedRoot, sourceRoot, err)
	}
	temporaryParent, err = filepath.EvalSymlinks(temporaryParent)
	if err != nil {
		t.Fatalf("resolve authorized temporary parent: %v", err)
	}
	inside, err := actualClaudePathWithin(temporaryParent, sourceRoot)
	if err != nil || !inside || temporaryParent == sourceRoot {
		t.Fatalf("source root %q is not an isolated child of %q: %v", sourceRoot, temporaryParent, err)
	}
	gitInfo, err := os.Lstat(filepath.Join(sourceRoot, ".git"))
	if err != nil || !gitInfo.IsDir() {
		t.Fatalf("actual-host source must have its own .git directory: info=%v error=%v", gitInfo, err)
	}
	common := actualHostGit(t, sourceRoot, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(sourceRoot, common)
	}
	common, err = filepath.EvalSymlinks(common)
	expectedCommon, expectedErr := filepath.EvalSymlinks(filepath.Join(sourceRoot, ".git"))
	if err != nil || expectedErr != nil || common != expectedCommon {
		t.Fatalf("actual-host source common directory=%q want=%q errors=%v/%v", common, expectedCommon, err, expectedErr)
	}
	command := exec.Command("git", "-C", sourceRoot, "symbolic-ref", "-q", "HEAD")
	command.Env = processadapter.MinimalEnvironment()
	output, commandErr := command.CombinedOutput()
	var exitError *exec.ExitError
	if !errors.As(commandErr, &exitError) || exitError.ExitCode() != 1 || len(output) != 0 {
		t.Fatalf("actual-host source must be detached: error=%v output=%q", commandErr, output)
	}
}

func assertActualClaudeSource(t *testing.T, sourceRoot, candidate, tree string) {
	t.Helper()
	if !actualHostHex(candidate, 20) || !actualHostHex(tree, 20) || actualHostGit(t, sourceRoot, "rev-parse", "HEAD") != candidate || actualHostGit(t, sourceRoot, "rev-parse", "HEAD^{tree}") != tree {
		t.Fatal("actual-host authorization must bind the exact full source candidate commit and tree")
	}
	if status := actualHostGit(t, sourceRoot, "status", "--porcelain=v1", "--untracked-files=all"); status != "" {
		t.Fatalf("actual-host source candidate is not clean: %q", status)
	}
	if remotes := actualHostGit(t, sourceRoot, "remote"); remotes != "" {
		t.Fatalf("actual-host source candidate unexpectedly has remotes: %q", remotes)
	}
}

func requiredActualClaudeEnvironment(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		t.Fatalf("actual-host authorization must set a safe exact %s binding", name)
	}
	return value
}

func actualClaudePathWithin(root, path string) (bool, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))), nil
}

func assertActualHostExecutable(t *testing.T, expectedPath, expectedDigest string) {
	t.Helper()
	executable, err := processadapter.Resolve(expectedPath)
	if err != nil || executable.Path != expectedPath || executable.Digest != expectedDigest {
		t.Fatalf("actual-host Claude executable identity changed: %+v error=%v", executable, err)
	}
}

func actualHostGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	command.Env = processadapter.MinimalEnvironment()
	output, err := command.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(arguments, " "), err)
	}
	return strings.TrimSpace(string(output))
}

func actualHostHex(value string, bytes int) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytes && value == strings.ToLower(value)
}
