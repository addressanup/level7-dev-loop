package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestControlledCommitAdvancesExpectedHeadWithExactPaths(t *testing.T) {
	repository, base := initializedRepository(t)
	configureIdentity(t, repository)
	writeFile(t, filepath.Join(repository, "change.txt"), "candidate")
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	pending, err := adapter.Pending(context.Background(), repository)
	if err != nil || strings.Join(pending.Paths, "|") != "change.txt" || pending.IndexDirty {
		t.Fatalf("Pending()=%+v error=%v", pending, err)
	}
	location, err := adapter.Commit(context.Background(), commitRequest(pending, []string{"change.txt"}))
	if err != nil || location.Head == base || location.Tree == pending.Tree {
		t.Fatalf("Commit()=%+v error=%v", location, err)
	}
	changed := strings.TrimSpace(runGit(t, repository, "diff", "--name-only", base, location.Head))
	if changed != "change.txt" {
		t.Fatalf("committed paths=%q", changed)
	}
	addition, err := adapter.PathCommit(context.Background(), repository, "change.txt")
	if err != nil || addition != location.Head {
		t.Fatalf("PathCommit()=%q error=%v", addition, err)
	}
	matches, err := adapter.CommitMatches(context.Background(), repository, location.Head, base, "feat(test): commit bounded change")
	if err != nil || !matches {
		t.Fatalf("CommitMatches()=%v error=%v", matches, err)
	}
	matches, err = adapter.CommitMatches(context.Background(), repository, location.Head, base, "fix(test): wrong subject")
	if err != nil || matches {
		t.Fatalf("wrong-subject CommitMatches()=%v error=%v", matches, err)
	}
}

func TestControlledCommitRejectsUnexpectedPathsAndStaleHead(t *testing.T) {
	repository, _ := initializedRepository(t)
	configureIdentity(t, repository)
	writeFile(t, filepath.Join(repository, "allowed.txt"), "allowed")
	writeFile(t, filepath.Join(repository, "outside.txt"), "outside")
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	pending, err := adapter.Pending(context.Background(), repository)
	if err != nil {
		t.Fatal(err)
	}
	request := commitRequest(pending, []string{"allowed.txt"})
	if _, err := adapter.Commit(context.Background(), request); err == nil || !strings.Contains(err.Error(), "exact commit path set") {
		t.Fatalf("scope-expanded Commit() error=%v", err)
	}
	request.Paths = append([]string{}, pending.Paths...)
	request.ExpectedCommit = strings.Repeat("f", 40)
	if _, err := adapter.Commit(context.Background(), request); err == nil || !strings.Contains(err.Error(), "changed before commit") {
		t.Fatalf("stale Commit() error=%v", err)
	}
}

func TestControlledCommitRetainsHookFailureAndUserWork(t *testing.T) {
	repository, _ := initializedRepository(t)
	configureIdentity(t, repository)
	writeFile(t, filepath.Join(repository, "change.txt"), "candidate")
	hook := filepath.Join(repository, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nexit 9\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	pending, err := adapter.Pending(context.Background(), repository)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Commit(context.Background(), commitRequest(pending, pending.Paths)); err == nil || !strings.Contains(err.Error(), "failed with exit") {
		t.Fatalf("hook-failed Commit() error=%v", err)
	}
	after, err := adapter.Pending(context.Background(), repository)
	if err != nil || strings.Join(after.Paths, "|") != "change.txt" {
		t.Fatalf("user work after hook failure=%+v error=%v", after, err)
	}
}

func TestPathSetDigestIsOrderIndependentAndFramed(t *testing.T) {
	left := PathSetDigest([]string{"ab", "c"})
	if left != PathSetDigest([]string{"c", "ab"}) || left == PathSetDigest([]string{"a", "bc"}) || len(left) != 64 || left == strings.Repeat("0", 64) {
		t.Fatalf("unexpected path-set digest %q", left)
	}
}

func commitRequest(pending domain.PendingChanges, paths []string) domain.CommitRequest {
	return domain.CommitRequest{
		Root: pending.Root, ExpectedCommit: pending.Head, ExpectedTree: pending.Tree,
		Paths: append([]string{}, paths...), Message: "feat(test): commit bounded change",
		MaxOutputBytes: DefaultMaxOutput, MaxPaths: DefaultMaxPaths, MaxCommandSeconds: 30,
	}
}

func configureIdentity(t *testing.T, repository string) {
	t.Helper()
	runGit(t, repository, "config", "user.name", "Level Seven")
	runGit(t, repository, "config", "user.email", "l7@example.invalid")
}
