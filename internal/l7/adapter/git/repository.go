// Package git reads canonical repository and candidate identity through explicit Git argv.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	DefaultMaxOutput = 16 << 20
	DefaultMaxPaths  = 100_000
	maxDiagnostic    = 8 << 10
)

var errOutputLimit = errors.New("bounded Git output limit reached")

type Adapter struct {
	binary    string
	maxOutput int
	maxPaths  int
}

func (adapter Adapter) WithLimits(maxOutput, maxPaths int) Adapter {
	if maxOutput > 0 {
		adapter.maxOutput = maxOutput
	}
	if maxPaths > 0 {
		adapter.maxPaths = maxPaths
	}
	return adapter
}

func New(binary string, maxOutput, maxPaths int) (Adapter, error) {
	if binary == "" {
		located, err := exec.LookPath("git")
		if err != nil {
			return Adapter{}, errors.New("Git executable is unavailable")
		}
		binary = located
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		return Adapter{}, fmt.Errorf("resolve Git executable: %w", err)
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return Adapter{}, fmt.Errorf("resolve physical Git executable: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return Adapter{}, errors.New("Git executable is not a regular executable file")
	}
	if maxOutput < 1 {
		maxOutput = DefaultMaxOutput
	}
	if maxPaths < 1 {
		maxPaths = DefaultMaxPaths
	}
	return Adapter{binary: absolute, maxOutput: maxOutput, maxPaths: maxPaths}, nil
}

func (adapter Adapter) Locate(ctx context.Context, workingDirectory string) (domain.RepositoryLocation, error) {
	if err := ctx.Err(); err != nil {
		return domain.RepositoryLocation{}, err
	}
	if workingDirectory == "" {
		return domain.RepositoryLocation{}, errors.New("working directory is empty")
	}
	absoluteWorkingDirectory, err := filepath.Abs(workingDirectory)
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("resolve working directory: %w", err)
	}
	inside, err := adapter.singleLine(ctx, absoluteWorkingDirectory, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("inspect Git worktree: %w", err)
	}
	if inside != "true" {
		return domain.RepositoryLocation{}, errors.New("working directory is not inside a Git worktree")
	}
	bare, err := adapter.singleLine(ctx, absoluteWorkingDirectory, "rev-parse", "--is-bare-repository")
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("inspect bare-repository state: %w", err)
	}
	if bare != "false" {
		return domain.RepositoryLocation{}, errors.New("bare Git repositories are unsupported")
	}
	rootOutput, err := adapter.singleLine(ctx, absoluteWorkingDirectory, "rev-parse", "--show-toplevel")
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("discover Git root: %w", err)
	}
	root, err := physicalDirectory(rootOutput)
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("validate Git root: %w", err)
	}
	commonOutput, err := adapter.singleLine(ctx, root, "rev-parse", "--git-common-dir")
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("discover Git common directory: %w", err)
	}
	if !filepath.IsAbs(commonOutput) {
		commonOutput = filepath.Join(root, commonOutput)
	}
	commonDirectory, err := physicalDirectory(commonOutput)
	if err != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("validate Git common directory: %w", err)
	}
	head, tree, err := adapter.identity(ctx, root)
	if err != nil {
		return domain.RepositoryLocation{}, err
	}
	recheckedHead, recheckedTree, err := adapter.identity(ctx, root)
	if err != nil || head != recheckedHead || tree != recheckedTree {
		return domain.RepositoryLocation{}, errors.New("Git identity changed during repository discovery")
	}
	return domain.RepositoryLocation{Root: root, CommonDir: commonDirectory, Head: head, Tree: tree}, nil
}

func (adapter Adapter) Snapshot(ctx context.Context, workingDirectory, base string) (domain.RepositorySnapshot, error) {
	if !fullObjectID(base) {
		return domain.RepositorySnapshot{}, errors.New("base must be a full lowercase Git object ID")
	}
	location, err := adapter.Locate(ctx, workingDirectory)
	if err != nil {
		return domain.RepositorySnapshot{}, err
	}
	resolvedBase, err := adapter.singleLine(ctx, location.Root, "rev-parse", "--verify", base+"^{commit}")
	if err != nil || resolvedBase != base {
		return domain.RepositorySnapshot{}, errors.New("base does not identify an available Git commit")
	}
	ancestor, err := adapter.isAncestor(ctx, location.Root, base, location.Head)
	if err != nil {
		return domain.RepositorySnapshot{}, err
	}
	if !ancestor {
		return domain.RepositorySnapshot{}, errors.New("base commit is not an ancestor of HEAD")
	}
	committedOutput, err := adapter.run(ctx, location.Root, "diff", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--diff-filter=ACDMRTUXB", base, location.Head, "--")
	if err != nil {
		return domain.RepositorySnapshot{}, fmt.Errorf("read committed Git paths: %w", err)
	}
	committed, err := parseNULPaths(committedOutput, adapter.maxPaths)
	if err != nil {
		return domain.RepositorySnapshot{}, err
	}
	statusOutput, err := adapter.run(ctx, location.Root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none", "--no-renames")
	if err != nil {
		return domain.RepositorySnapshot{}, fmt.Errorf("read Git worktree status: %w", err)
	}
	working, err := parseStatus(statusOutput, adapter.maxPaths)
	if err != nil {
		return domain.RepositorySnapshot{}, err
	}
	changed, err := combinePaths(adapter.maxPaths, committed, working)
	if err != nil {
		return domain.RepositorySnapshot{}, err
	}
	recheckedHead, recheckedTree, err := adapter.identity(ctx, location.Root)
	if err != nil || location.Head != recheckedHead || location.Tree != recheckedTree {
		return domain.RepositorySnapshot{}, errors.New("Git identity changed during status reconstruction")
	}
	return domain.RepositorySnapshot{RepositoryLocation: location, Base: base, ChangedPaths: changed}, nil
}

func (adapter Adapter) identity(ctx context.Context, root string) (string, string, error) {
	data, err := adapter.run(ctx, root, "rev-parse", "HEAD^{commit}", "HEAD^{tree}")
	if err != nil {
		return "", "", errors.New("repository has no readable initial commit")
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != 2 || !fullObjectID(lines[0]) || !fullObjectID(lines[1]) {
		return "", "", errors.New("Git returned an invalid commit/tree identity")
	}
	return lines[0], lines[1], nil
}

func (adapter Adapter) isAncestor(ctx context.Context, root, base, head string) (bool, error) {
	command := adapter.command(ctx, root, "merge-base", "--is-ancestor", base, head)
	var diagnostic limitedBuffer
	diagnostic.limit = maxDiagnostic
	command.Stdout = &diagnostic
	command.Stderr = &diagnostic
	err := command.Run()
	if diagnostic.exceeded {
		return false, errors.New("Git ancestry diagnostic exceeds size limit")
	}
	if err == nil {
		return true, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
		return false, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return false, fmt.Errorf("check Git ancestry: %w: %s", err, boundedDiagnostic(diagnostic.String()))
}

func (adapter Adapter) singleLine(ctx context.Context, root string, arguments ...string) (string, error) {
	data, err := adapter.run(ctx, root, arguments...)
	if err != nil {
		return "", err
	}
	value := strings.TrimSuffix(string(data), "\n")
	if value == "" || strings.ContainsAny(value, "\r\n\x00") || !utf8.ValidString(value) {
		return "", errors.New("Git returned an invalid single-line value")
	}
	return value, nil
}

func (adapter Adapter) run(ctx context.Context, root string, arguments ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	command := adapter.command(ctx, root, arguments...)
	stdout := limitedBuffer{limit: adapter.maxOutput}
	stderr := limitedBuffer{limit: maxDiagnostic}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if stdout.exceeded {
		return nil, fmt.Errorf("Git output exceeds %d bytes", adapter.maxOutput)
	}
	if stderr.exceeded {
		return nil, errors.New("Git diagnostic exceeds size limit")
	}
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("Git command failed: %w: %s", err, boundedDiagnostic(stderr.String()))
	}
	return append([]byte{}, stdout.Bytes()...), nil
}

func (adapter Adapter) command(ctx context.Context, root string, arguments ...string) *exec.Cmd {
	argv := []string{"-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "--no-optional-locks", "-C", root}
	argv = append(argv, arguments...)
	command := exec.CommandContext(ctx, adapter.binary, argv...)
	command.Env = []string{
		"LC_ALL=C",
		"LANG=C",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	}
	return command
}

func parseNULPaths(data []byte, maximum int) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	if data[len(data)-1] != 0 {
		return nil, errors.New("Git path output is not NUL terminated")
	}
	records := bytes.Split(data[:len(data)-1], []byte{0})
	if len(records) > maximum {
		return nil, fmt.Errorf("Git path count exceeds %d", maximum)
	}
	paths := make([]string, 0, len(records))
	for _, record := range records {
		value := string(record)
		if !safeGitPath(value) {
			return nil, errors.New("Git returned an unsafe repository path")
		}
		paths = append(paths, value)
	}
	return paths, nil
}

func parseStatus(data []byte, maximum int) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	if data[len(data)-1] != 0 {
		return nil, errors.New("Git status output is not NUL terminated")
	}
	records := bytes.Split(data[:len(data)-1], []byte{0})
	if len(records) > maximum {
		return nil, fmt.Errorf("Git path count exceeds %d", maximum)
	}
	paths := make([]string, 0, len(records))
	for _, record := range records {
		if len(record) < 4 || record[2] != ' ' || !validStatusCode(record[0]) || !validStatusCode(record[1]) || (record[0] == ' ' && record[1] == ' ') || record[0] == '!' || record[1] == '!' {
			return nil, errors.New("Git returned an invalid porcelain status record")
		}
		value := string(record[3:])
		if !safeGitPath(value) {
			return nil, errors.New("Git returned an unsafe status path")
		}
		paths = append(paths, value)
	}
	return paths, nil
}

func validStatusCode(value byte) bool {
	return strings.ContainsRune(" MADRCU?!", rune(value))
}

func combinePaths(maximum int, groups ...[]string) ([]string, error) {
	seen := make(map[string]bool)
	for _, group := range groups {
		for _, value := range group {
			seen[value] = true
			if len(seen) > maximum {
				return nil, fmt.Errorf("combined Git path count exceeds %d", maximum)
			}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func safeGitPath(value string) bool {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 4096 || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "`") {
		return false
	}
	if value == "." || value == ".." || path.Clean(value) != value || strings.HasPrefix(value, "../") {
		return false
	}
	for _, character := range value {
		if character == 0 || character == 0x7f || character < 0x20 {
			return false
		}
	}
	return true
}

func fullObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func physicalDirectory(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n\x00") || !utf8.ValidString(value) {
		return "", errors.New("directory contains an unsafe character")
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	physical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(physical)
	if err != nil || !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	return filepath.Clean(physical), nil
}

func boundedDiagnostic(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 512 {
		value = value[:512]
	}
	return value
}

type limitedBuffer struct {
	data     bytes.Buffer
	limit    int
	exceeded bool
}

func (buffer *limitedBuffer) Write(data []byte) (int, error) {
	if buffer.limit-buffer.data.Len() < len(data) {
		remaining := buffer.limit - buffer.data.Len()
		if remaining > 0 {
			_, _ = buffer.data.Write(data[:remaining])
		}
		buffer.exceeded = true
		return remaining, errOutputLimit
	}
	return buffer.data.Write(data)
}

func (buffer *limitedBuffer) Bytes() []byte { return buffer.data.Bytes() }

func (buffer *limitedBuffer) String() string { return buffer.data.String() }
