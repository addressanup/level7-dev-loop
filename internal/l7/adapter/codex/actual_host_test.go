//go:build l7_actual_provider

package codex

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestActualHostInterfaceBinding(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_CODEX") != "interface:codex-cli-0.150.1" {
		t.Skip("set the exact separately approved L7_AUTHORIZE_ACTUAL_CODEX interface token")
	}
	sourceRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err == nil {
		sourceRoot, err = filepath.EvalSymlinks(sourceRoot)
	}
	if err != nil {
		t.Fatal(err)
	}
	assertActualCodexSourceIsolation(t, sourceRoot)
	expectedCandidate := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_CANDIDATE")
	expectedTree := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_TREE")
	assertActualCodexSource(t, sourceRoot, expectedCandidate, expectedTree)
	if expected := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_GOOS"); expected != runtime.GOOS {
		t.Fatalf("actual-host GOOS=%q want=%q", runtime.GOOS, expected)
	}
	if expected := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_GOARCH"); expected != runtime.GOARCH {
		t.Fatalf("actual-host GOARCH=%q want=%q", runtime.GOARCH, expected)
	}
	expectedExecutable, err := filepath.EvalSymlinks(requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_EXECUTABLE"))
	if err != nil || !filepath.IsAbs(expectedExecutable) {
		t.Fatalf("actual-host executable binding is not a physical absolute path: %q error=%v", expectedExecutable, err)
	}
	expectedDigest := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_DIGEST")
	if !actualCodexHex(expectedDigest, 32) {
		t.Fatal("actual-host authorization must bind an exact lowercase SHA-256 digest")
	}
	resolved, err := processadapter.Resolve("codex")
	if err != nil || resolved.Path != expectedExecutable || resolved.Digest != expectedDigest {
		t.Fatalf("resolved provider does not match the pre-authorized executable binding: %+v error=%v", resolved, err)
	}
	adapter := New()
	identity, err := adapter.Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if identity.Executable != expectedExecutable || identity.Digest != expectedDigest || identity.Version != requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_VERSION") || identity.Version != CompatibleVersion || identity.Capability != domain.CapabilityAvailable {
		t.Fatalf("provider identity does not match the authorized executable tuple: %+v", identity)
	}
	if err := actualCodexHelp(adapter, identity, sourceRoot, []string{"--help"}, "--ask-for-approval", "exec"); err != nil {
		t.Fatal(err)
	}
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		schemaPath, cleanup, err := adapter.prepareTerminalSchema(role, sourceRoot)
		if err != nil {
			t.Fatal(err)
		}
		helpArguments := arguments(role, sourceRoot, schemaPath)
		helpArguments[len(helpArguments)-1] = "--help"
		helpErr := actualCodexHelp(adapter, identity, sourceRoot, helpArguments, "--ephemeral", "--sandbox", "--color", "--json", "--output-schema", "--cd")
		if err := errors.Join(helpErr, cleanup()); err != nil {
			t.Fatalf("%s interface contract: %v", role, err)
		}
	}
	assertActualCodexSource(t, sourceRoot, expectedCandidate, expectedTree)
	t.Logf("provider=%s version=%q executable=%q digest=%s capability=%s", identity.Provider, identity.Version, identity.Executable, identity.Digest, identity.Capability)
}

func assertActualCodexSourceIsolation(t *testing.T, sourceRoot string) {
	t.Helper()
	authorizedRoot := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_SOURCE_ROOT")
	temporaryParent := requiredActualEnvironment(t, "L7_AUTHORIZE_ACTUAL_CODEX_TEMP_PARENT")
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
	inside, err := pathWithin(temporaryParent, sourceRoot)
	if err != nil || !inside || temporaryParent == sourceRoot {
		t.Fatalf("source root %q is not an isolated child of %q: %v", sourceRoot, temporaryParent, err)
	}
	gitInfo, err := os.Lstat(filepath.Join(sourceRoot, ".git"))
	if err != nil || !gitInfo.IsDir() {
		t.Fatalf("actual-host source must have its own .git directory: info=%v error=%v", gitInfo, err)
	}
	common := actualGit(t, sourceRoot, "rev-parse", "--git-common-dir")
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

func assertActualCodexSource(t *testing.T, sourceRoot, candidate, tree string) {
	t.Helper()
	actualCandidate := actualGit(t, sourceRoot, "rev-parse", "HEAD")
	if !actualCodexHex(candidate, 20) || actualCandidate != candidate {
		t.Fatalf("actual-host source candidate=%q want exact full candidate=%q", actualCandidate, candidate)
	}
	actualTree := actualGit(t, sourceRoot, "rev-parse", "HEAD^{tree}")
	if !actualCodexHex(tree, 20) || actualTree != tree {
		t.Fatalf("actual-host source tree=%q want exact tree=%q", actualTree, tree)
	}
	if status := actualGit(t, sourceRoot, "status", "--porcelain=v1", "--untracked-files=all"); status != "" {
		t.Fatalf("actual-host source candidate is not clean: %q", status)
	}
	if remotes := actualGit(t, sourceRoot, "remote"); remotes != "" {
		t.Fatalf("actual-host source candidate has remotes: %q", remotes)
	}
}

func actualCodexHelp(adapter Adapter, identity domain.ProviderIdentity, root string, arguments []string, required ...string) error {
	result, err := adapter.runtime.Invoke(context.Background(), identity, root, arguments, nil, 1<<20, 30)
	if err != nil {
		return fmt.Errorf("bounded no-model Codex help invocation failed: %w", err)
	}
	output := string(append(append([]byte{}, result.Stdout...), result.Stderr...))
	if output == "" || !utf8.ValidString(output) {
		return errors.New("Codex help surface is empty or invalid UTF-8")
	}
	for _, value := range required {
		if !strings.Contains(output, value) {
			return fmt.Errorf("Codex help surface does not declare required argument %q", value)
		}
	}
	return nil
}

func requiredActualEnvironment(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		t.Fatalf("actual-host authorization must set a safe exact %s binding", name)
	}
	return value
}

func actualCodexHex(value string, bytes int) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytes && value == strings.ToLower(value)
}

func actualGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("inspect actual-host source with Git: %v: %s", err, output)
	}
	if len(output) > 1<<20 {
		t.Fatal("actual-host Git inspection exceeded its output bound")
	}
	return strings.TrimSpace(string(output))
}
