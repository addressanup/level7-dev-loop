package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func (adapter Adapter) Pending(ctx context.Context, workingDirectory string) (domain.PendingChanges, error) {
	location, err := adapter.Locate(ctx, workingDirectory)
	if err != nil {
		return domain.PendingChanges{}, err
	}
	status, err := adapter.run(ctx, location.Root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none", "--no-renames")
	if err != nil {
		return domain.PendingChanges{}, fmt.Errorf("read pending Git paths: %w", err)
	}
	paths, err := parseStatus(status, adapter.maxPaths)
	if err != nil {
		return domain.PendingChanges{}, err
	}
	indexDirty, err := statusIndexDirty(status)
	if err != nil {
		return domain.PendingChanges{}, err
	}
	rechecked, err := adapter.Locate(ctx, location.Root)
	if err != nil || !sameRepositoryLocation(location, rechecked) {
		return domain.PendingChanges{}, errors.New("Git identity changed while reading pending paths")
	}
	return domain.PendingChanges{RepositoryLocation: location, Paths: paths, IndexDirty: indexDirty}, nil
}

func (adapter Adapter) PathCommit(ctx context.Context, root, relative string) (string, error) {
	if !safeGitPath(relative) {
		return "", errors.New("path is unsafe")
	}
	commit, err := adapter.singleLine(ctx, root, "log", "--max-count=1", "--format=%H", "--diff-filter=A", "--", relative)
	if err != nil || !fullObjectID(commit) {
		return "", errors.New("path has no readable addition commit")
	}
	return commit, nil
}

func (adapter Adapter) Commit(ctx context.Context, request domain.CommitRequest) (domain.RepositoryLocation, error) {
	if err := validateCommitRequest(request); err != nil {
		return domain.RepositoryLocation{}, err
	}
	limited := adapter.WithLimits(request.MaxOutputBytes, request.MaxPaths)
	pending, err := limited.Pending(ctx, request.Root)
	if err != nil {
		return domain.RepositoryLocation{}, err
	}
	if pending.Head != request.ExpectedCommit || pending.Tree != request.ExpectedTree {
		return domain.RepositoryLocation{}, errors.New("expected Git candidate changed before commit")
	}
	if !samePathSet(pending.Paths, request.Paths) {
		return domain.RepositoryLocation{}, errors.New("pending Git paths do not match the exact commit path set")
	}
	input := pathspecInput(request.Paths)
	environment := append(processadapter.MinimalEnvironment(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_OPTIONAL_LOCKS=0",
	)
	stageArguments := []string{"-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-C", request.Root, "add", "--all", "--pathspec-from-file=-", "--pathspec-file-nul"}
	result, runErr := runGitEffect(ctx, adapter.binary, request, environment, stageArguments, input)
	if runErr != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("controlled Git staging did not complete: %w", runErr)
	}
	if result.ExitCode != 0 {
		diagnostic := strings.TrimSpace(string(append(result.Stdout, result.Stderr...)))
		return domain.RepositoryLocation{}, fmt.Errorf("controlled Git staging failed with exit %d: %s", result.ExitCode, boundedDiagnostic(diagnostic))
	}
	stagedOutput, err := limited.run(ctx, request.Root, "diff", "--cached", "--name-only", "-z", "--no-renames", "--diff-filter=ACDMRTUXB", "--")
	if err != nil {
		return domain.RepositoryLocation{}, errors.New("cannot verify the controlled Git index")
	}
	staged, err := parseNULPaths(stagedOutput, request.MaxPaths)
	if err != nil || !samePathSet(staged, request.Paths) {
		return domain.RepositoryLocation{}, errors.New("controlled Git index contains an unexpected path; recovery is required")
	}
	commitArguments := []string{"-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-C", request.Root, "commit", "--quiet", "-m", request.Message}
	result, runErr = runGitEffect(ctx, adapter.binary, request, environment, commitArguments, nil)
	if runErr != nil {
		return domain.RepositoryLocation{}, fmt.Errorf("controlled Git commit did not complete: %w", runErr)
	}
	if result.ExitCode != 0 {
		diagnostic := strings.TrimSpace(string(append(result.Stdout, result.Stderr...)))
		return domain.RepositoryLocation{}, fmt.Errorf("controlled Git commit failed with exit %d: %s", result.ExitCode, boundedDiagnostic(diagnostic))
	}
	location, err := limited.Locate(ctx, request.Root)
	if err != nil {
		return domain.RepositoryLocation{}, errors.New("commit completed but candidate identity cannot be reconstructed")
	}
	if location.Head == request.ExpectedCommit {
		return domain.RepositoryLocation{}, errors.New("Git commit did not advance HEAD")
	}
	parent, err := limited.singleLine(ctx, request.Root, "rev-parse", "HEAD^1")
	if err != nil || parent != request.ExpectedCommit {
		return domain.RepositoryLocation{}, errors.New("Git commit parent does not match the expected candidate")
	}
	changedOutput, err := limited.run(ctx, request.Root, "diff", "--name-only", "-z", "--no-renames", "--diff-filter=ACDMRTUXB", request.ExpectedCommit, location.Head, "--")
	if err != nil {
		return domain.RepositoryLocation{}, errors.New("commit completed but changed paths cannot be reconstructed")
	}
	changed, err := parseNULPaths(changedOutput, request.MaxPaths)
	if err != nil || !samePathSet(changed, request.Paths) {
		return domain.RepositoryLocation{}, errors.New("commit completed with an unexpected path set; recovery is required")
	}
	after, err := limited.Pending(ctx, request.Root)
	if err != nil || len(after.Paths) != 0 || after.IndexDirty {
		return domain.RepositoryLocation{}, errors.New("commit completed but the worktree or index is not clean; recovery is required")
	}
	return location, nil
}

func runGitEffect(ctx context.Context, binary string, request domain.CommitRequest, environment, arguments []string, input []byte) (processadapter.Result, error) {
	return (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: binary, Arguments: arguments, Input: input, Directory: request.Root,
		Environment: environment, MaxOutputBytes: request.MaxOutputBytes,
		Timeout: time.Duration(request.MaxCommandSeconds) * time.Second,
	})
}

func validateCommitRequest(request domain.CommitRequest) error {
	if request.Root == "" || !fullObjectID(request.ExpectedCommit) || !fullObjectID(request.ExpectedTree) || !domain.ConventionalSubject(request.Message) || len(request.Paths) < 1 || request.MaxOutputBytes < 64<<10 || request.MaxOutputBytes > 64<<20 || request.MaxPaths < 1 || len(request.Paths) > request.MaxPaths || request.MaxCommandSeconds < 1 || request.MaxCommandSeconds > 86400 {
		return errors.New("controlled Git commit request is invalid")
	}
	seen := make(map[string]bool)
	for _, relative := range request.Paths {
		if !safeGitPath(relative) || seen[relative] {
			return errors.New("controlled Git commit path set is invalid")
		}
		seen[relative] = true
	}
	return nil
}

func statusIndexDirty(data []byte) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}
	if data[len(data)-1] != 0 {
		return false, errors.New("Git status output is not NUL terminated")
	}
	for _, record := range bytes.Split(data[:len(data)-1], []byte{0}) {
		if len(record) < 4 {
			return false, errors.New("Git returned an invalid status record")
		}
		if record[0] != ' ' && record[0] != '?' {
			return true, nil
		}
	}
	return false, nil
}

func pathspecInput(paths []string) []byte {
	var input bytes.Buffer
	for _, relative := range paths {
		input.WriteString(relative)
		input.WriteByte(0)
	}
	return input.Bytes()
}

func samePathSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string{}, left...)
	rightCopy := append([]string{}, right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}

func sameRepositoryLocation(left, right domain.RepositoryLocation) bool {
	return left.Root == right.Root && left.CommonDir == right.CommonDir && left.Head == right.Head && left.Tree == right.Tree
}
