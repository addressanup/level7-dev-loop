//go:build l7_actual_provider

package claude

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

func TestClaudeQualificationActualHostNoModel(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL") != "gate:claude-code-2.1.247" {
		t.Skip("set the exact separately approved Claude no-model gate token")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		t.Fatal(err)
	}
	actualClaudeQualificationAssertSource(t, root)
	if expected := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_GOOS"); expected != runtime.GOOS {
		t.Fatalf("actual-host GOOS=%q want=%q", runtime.GOOS, expected)
	}
	if expected := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_GOARCH"); expected != runtime.GOARCH {
		t.Fatalf("actual-host GOARCH=%q want=%q", runtime.GOARCH, expected)
	}
	version := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_VERSION")
	if version != claudeQualificationTargetBare && version != claudeQualificationTargetNamed {
		t.Fatalf("authorized Claude version=%q is not an exact diagnostic target", version)
	}

	authorizedPath := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_EXECUTABLE")
	physicalPath, err := filepath.EvalSymlinks(authorizedPath)
	expectedDigest := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_DIGEST")
	if err != nil || !filepath.IsAbs(physicalPath) || physicalPath != authorizedPath || !actualClaudeQualificationHex(expectedDigest, 32) {
		t.Fatalf("authorized Claude executable tuple is invalid: path=%q error=%v", physicalPath, err)
	}
	resolved, err := processadapter.Resolve("claude")
	if err != nil || resolved.Path != physicalPath || resolved.Digest != expectedDigest {
		t.Fatalf("resolved Claude does not match authorization: %+v error=%v", resolved, err)
	}

	t.Cleanup(func() { actualClaudeQualificationAssertSource(t, root) })
	t.Cleanup(func() { actualClaudeQualificationAssertExecutable(t, physicalPath, expectedDigest) })
	report, err := runClaudeQualificationChecked(context.Background(), (processadapter.Runner{}).Run, physicalPath, root, func() error {
		actualClaudeQualificationAssertSource(t, root)
		actualClaudeQualificationAssertExecutable(t, physicalPath, expectedDigest)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.version != version {
		t.Fatalf("observed Claude version=%q want authorized=%q", report.version, version)
	}
	t.Logf("diagnostic-only provider=claude version=%q executable=%q digest=%s advertised=%v", report.version, physicalPath, expectedDigest, report.diagnostics)
}

func actualClaudeQualificationAssertSource(t *testing.T, root string) {
	t.Helper()
	authorizedRoot := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_SOURCE_ROOT")
	temporaryParent := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_TEMP_PARENT")
	physicalRoot, rootErr := filepath.EvalSymlinks(authorizedRoot)
	physicalParent, parentErr := filepath.EvalSymlinks(temporaryParent)
	inside, insideErr := actualClaudeQualificationPathWithin(physicalParent, root)
	if rootErr != nil || parentErr != nil || insideErr != nil || !filepath.IsAbs(physicalRoot) || !filepath.IsAbs(physicalParent) || physicalRoot != root || physicalParent == root || !inside {
		t.Fatalf("actual-host source isolation is invalid: root=%q parent=%q errors=%v/%v/%v", physicalRoot, physicalParent, rootErr, parentErr, insideErr)
	}
	gitInfo, err := os.Lstat(filepath.Join(root, ".git"))
	if err != nil || !gitInfo.IsDir() {
		t.Fatalf("actual-host source must own a .git directory: info=%v error=%v", gitInfo, err)
	}
	common := actualClaudeQualificationGit(t, root, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common, err = filepath.EvalSymlinks(common)
	expectedCommon, expectedErr := filepath.EvalSymlinks(filepath.Join(root, ".git"))
	if err != nil || expectedErr != nil || common != expectedCommon {
		t.Fatalf("actual-host Git common directory=%q want=%q errors=%v/%v", common, expectedCommon, err, expectedErr)
	}
	detached := actualClaudeQualificationGitResult(t, root, "symbolic-ref", "-q", "HEAD")
	if detached.ExitCode != 1 || strings.TrimSpace(string(append(detached.Stdout, detached.Stderr...))) != "" {
		t.Fatalf("actual-host source is not detached: %+v", detached)
	}
	candidate := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_CANDIDATE")
	tree := actualClaudeQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CLAUDE_NO_MODEL_TREE")
	if !actualClaudeQualificationHex(candidate, 20) || !actualClaudeQualificationHex(tree, 20) || actualClaudeQualificationGit(t, root, "rev-parse", "HEAD") != candidate || actualClaudeQualificationGit(t, root, "rev-parse", "HEAD^{tree}") != tree {
		t.Fatal("actual-host source does not match the exact authorized commit/tree")
	}
	if status := actualClaudeQualificationGit(t, root, "status", "--porcelain=v1", "--untracked-files=all"); status != "" {
		t.Fatalf("actual-host source is not clean: %q", status)
	}
	if remotes := actualClaudeQualificationGit(t, root, "remote"); remotes != "" {
		t.Fatalf("actual-host source has remotes: %q", remotes)
	}
}

func actualClaudeQualificationAssertExecutable(t *testing.T, expectedPath, expectedDigest string) {
	t.Helper()
	executable, err := processadapter.Resolve(expectedPath)
	if err != nil || executable.Path != expectedPath || executable.Digest != expectedDigest {
		t.Fatalf("Claude executable identity changed: %+v error=%v", executable, err)
	}
}

func actualClaudeQualificationRequired(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		t.Fatalf("actual-host authorization must set safe exact %s", name)
	}
	return value
}

func actualClaudeQualificationPathWithin(root, path string) (bool, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	return relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func actualClaudeQualificationHex(value string, bytes int) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytes && value == strings.ToLower(value)
}

func actualClaudeQualificationGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	result := actualClaudeQualificationGitResult(t, root, arguments...)
	if result.ExitCode != 0 {
		t.Fatalf("git %s exited %d", strings.Join(arguments, " "), result.ExitCode)
	}
	output := append(append([]byte{}, result.Stdout...), result.Stderr...)
	if len(output) > 64<<10 || !utf8.Valid(output) {
		t.Fatalf("git %s returned invalid bounded output", strings.Join(arguments, " "))
	}
	return strings.TrimSpace(string(output))
}

func actualClaudeQualificationGitResult(t *testing.T, root string, arguments ...string) processadapter.Result {
	t.Helper()
	git, err := processadapter.Resolve("git")
	if err != nil {
		t.Fatal(err)
	}
	environment := append(processadapter.MinimalEnvironment(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	result, err := (processadapter.Runner{}).Run(context.Background(), processadapter.Request{
		Executable: git.Path, Arguments: append([]string{}, arguments...), Directory: root,
		Environment: environment, MaxOutputBytes: 64 << 10, Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(arguments, " "), err)
	}
	return result
}
