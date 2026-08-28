//go:build l7_actual_provider

package codex

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

func TestCodexQualificationActualHostNoModel(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL") != "gate:codex-cli-0.150.1" {
		t.Skip("set the exact separately approved Codex no-model gate token")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		t.Fatal(err)
	}
	actualCodexQualificationAssertSource(t, root)
	if expected := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_GOOS"); expected != runtime.GOOS {
		t.Fatalf("actual-host GOOS=%q want=%q", runtime.GOOS, expected)
	}
	if expected := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_GOARCH"); expected != runtime.GOARCH {
		t.Fatalf("actual-host GOARCH=%q want=%q", runtime.GOARCH, expected)
	}
	if version := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_VERSION"); version != codexQualificationTarget {
		t.Fatalf("authorized Codex version=%q want=%q", version, codexQualificationTarget)
	}

	authorizedPath := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_EXECUTABLE")
	physicalPath, err := filepath.EvalSymlinks(authorizedPath)
	expectedDigest := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_DIGEST")
	if err != nil || !filepath.IsAbs(physicalPath) || physicalPath != authorizedPath || !actualCodexQualificationHex(expectedDigest, 32) {
		t.Fatalf("authorized Codex executable tuple is invalid: path=%q error=%v", physicalPath, err)
	}
	resolved, err := processadapter.Resolve("codex")
	if err != nil || resolved.Path != physicalPath || resolved.Digest != expectedDigest {
		t.Fatalf("resolved Codex does not match authorization: %+v error=%v", resolved, err)
	}

	t.Cleanup(func() { actualCodexQualificationAssertSource(t, root) })
	t.Cleanup(func() { actualCodexQualificationAssertExecutable(t, physicalPath, expectedDigest) })
	report, err := runCodexQualificationChecked(context.Background(), (processadapter.Runner{}).Run, physicalPath, root, func() error {
		actualCodexQualificationAssertSource(t, root)
		actualCodexQualificationAssertExecutable(t, physicalPath, expectedDigest)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("diagnostic-only provider=codex version=%q executable=%q digest=%s advertised=%v", report.version, physicalPath, expectedDigest, report.diagnostics)
}

func actualCodexQualificationAssertSource(t *testing.T, root string) {
	t.Helper()
	authorizedRoot := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_SOURCE_ROOT")
	temporaryParent := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_TEMP_PARENT")
	physicalRoot, rootErr := filepath.EvalSymlinks(authorizedRoot)
	physicalParent, parentErr := filepath.EvalSymlinks(temporaryParent)
	inside, insideErr := actualCodexQualificationPathWithin(physicalParent, root)
	if rootErr != nil || parentErr != nil || insideErr != nil || !filepath.IsAbs(physicalRoot) || !filepath.IsAbs(physicalParent) || physicalRoot != root || physicalParent == root || !inside {
		t.Fatalf("actual-host source isolation is invalid: root=%q parent=%q errors=%v/%v/%v", physicalRoot, physicalParent, rootErr, parentErr, insideErr)
	}
	gitInfo, err := os.Lstat(filepath.Join(root, ".git"))
	if err != nil || !gitInfo.IsDir() {
		t.Fatalf("actual-host source must own a .git directory: info=%v error=%v", gitInfo, err)
	}
	common := actualCodexQualificationGit(t, root, "rev-parse", "--git-common-dir")
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common, err = filepath.EvalSymlinks(common)
	expectedCommon, expectedErr := filepath.EvalSymlinks(filepath.Join(root, ".git"))
	if err != nil || expectedErr != nil || common != expectedCommon {
		t.Fatalf("actual-host Git common directory=%q want=%q errors=%v/%v", common, expectedCommon, err, expectedErr)
	}
	detached := actualCodexQualificationGitResult(t, root, "symbolic-ref", "-q", "HEAD")
	if detached.ExitCode != 1 || strings.TrimSpace(string(append(detached.Stdout, detached.Stderr...))) != "" {
		t.Fatalf("actual-host source is not detached: %+v", detached)
	}
	candidate := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_CANDIDATE")
	tree := actualCodexQualificationRequired(t, "L7_AUTHORIZE_ACTUAL_CODEX_NO_MODEL_TREE")
	if !actualCodexQualificationHex(candidate, 20) || !actualCodexQualificationHex(tree, 20) || actualCodexQualificationGit(t, root, "rev-parse", "HEAD") != candidate || actualCodexQualificationGit(t, root, "rev-parse", "HEAD^{tree}") != tree {
		t.Fatal("actual-host source does not match the exact authorized commit/tree")
	}
	if status := actualCodexQualificationGit(t, root, "status", "--porcelain=v1", "--untracked-files=all"); status != "" {
		t.Fatalf("actual-host source is not clean: %q", status)
	}
	if remotes := actualCodexQualificationGit(t, root, "remote"); remotes != "" {
		t.Fatalf("actual-host source has remotes: %q", remotes)
	}
}

func actualCodexQualificationAssertExecutable(t *testing.T, expectedPath, expectedDigest string) {
	t.Helper()
	executable, err := processadapter.Resolve(expectedPath)
	if err != nil || executable.Path != expectedPath || executable.Digest != expectedDigest {
		t.Fatalf("Codex executable identity changed: %+v error=%v", executable, err)
	}
}

func actualCodexQualificationRequired(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		t.Fatalf("actual-host authorization must set safe exact %s", name)
	}
	return value
}

func actualCodexQualificationPathWithin(root, path string) (bool, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	return relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func actualCodexQualificationHex(value string, bytes int) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == bytes && value == strings.ToLower(value)
}

func actualCodexQualificationGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	result := actualCodexQualificationGitResult(t, root, arguments...)
	if result.ExitCode != 0 {
		t.Fatalf("git %s exited %d", strings.Join(arguments, " "), result.ExitCode)
	}
	output := append(append([]byte{}, result.Stdout...), result.Stderr...)
	if len(output) > 64<<10 || !utf8.Valid(output) {
		t.Fatalf("git %s returned invalid bounded output", strings.Join(arguments, " "))
	}
	return strings.TrimSpace(string(output))
}

func actualCodexQualificationGitResult(t *testing.T, root string, arguments ...string) processadapter.Result {
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
