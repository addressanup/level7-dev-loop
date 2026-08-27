package git

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestAtomicMergeAdvancesOnlyTheExplicitLocalTarget(t *testing.T) {
	repository, base, candidate := mergeRepository(t)
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	request := domain.MergeRequest{Root: repository, TargetBranch: "release-target", ExpectedOld: base, Candidate: candidate, MaxOutputBytes: DefaultMaxOutput}
	target, err := adapter.InspectMerge(context.Background(), request)
	if err != nil || target.Ref != "refs/heads/release-target" || target.CurrentCommit != base || target.AlreadyAdvanced {
		t.Fatalf("InspectMerge()=%+v error=%v", target, err)
	}
	if err := adapter.AdvanceMerge(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runGit(t, repository, "show-ref", "--hash", "refs/heads/release-target")); got != candidate {
		t.Fatalf("target=%q want=%q", got, candidate)
	}
	if got := strings.TrimSpace(runGit(t, repository, "rev-parse", "HEAD")); got != candidate {
		t.Fatalf("candidate worktree HEAD changed=%q", got)
	}
	recovered, err := adapter.InspectMerge(context.Background(), request)
	if err != nil || !recovered.AlreadyAdvanced {
		t.Fatalf("recovery InspectMerge()=%+v error=%v", recovered, err)
	}
	tree := strings.TrimSpace(runGit(t, repository, "rev-parse", candidate+"^{tree}"))
	receipt := domain.MergeReceipt{ChangeID: "change", TargetRef: target.Ref, PreviousCommit: base, Candidate: domain.CandidateIdentity{Commit: candidate, Tree: tree}, ConfigurationDigest: strings.Repeat("a", 64), VerificationCommit: strings.Repeat("b", 40), ReviewCommit: strings.Repeat("c", 40)}
	current, err := adapter.MergeCurrent(context.Background(), repository, receipt, DefaultMaxOutput)
	if err != nil || !current {
		t.Fatalf("MergeCurrent()=%v error=%v", current, err)
	}
}

func TestMergeRejectsCheckedOutDivergentAndUnsafeTargets(t *testing.T) {
	t.Run("checked out", func(t *testing.T) {
		repository, base, candidate := mergeRepository(t)
		worktree := filepath.Join(physicalTemp(t), "target")
		runGit(t, repository, "worktree", "add", "-q", worktree, "release-target")
		request := domain.MergeRequest{Root: repository, TargetBranch: "release-target", ExpectedOld: base, Candidate: candidate, MaxOutputBytes: DefaultMaxOutput}
		if _, err := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths).InspectMerge(context.Background(), request); err == nil || !strings.Contains(err.Error(), "checked out") {
			t.Fatalf("checked-out InspectMerge() error=%v", err)
		}
	})
	t.Run("divergent", func(t *testing.T) {
		repository, base, candidate := mergeRepository(t)
		tree := strings.TrimSpace(runGit(t, repository, "rev-parse", base+"^{tree}"))
		divergent := strings.TrimSpace(runGit(t, repository, "commit-tree", tree, "-m", "divergent"))
		runGit(t, repository, "update-ref", "refs/heads/release-target", divergent, base)
		request := domain.MergeRequest{Root: repository, TargetBranch: "release-target", ExpectedOld: base, Candidate: candidate, MaxOutputBytes: DefaultMaxOutput}
		if _, err := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths).InspectMerge(context.Background(), request); err == nil || !strings.Contains(err.Error(), "diverged") {
			t.Fatalf("divergent InspectMerge() error=%v", err)
		}
	})
	t.Run("unsafe", func(t *testing.T) {
		repository, base, candidate := mergeRepository(t)
		adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
		for _, branch := range []string{"--help", "refs/heads/release-target", "../escape", "space name", "missing"} {
			request := domain.MergeRequest{Root: repository, TargetBranch: branch, ExpectedOld: base, Candidate: candidate, MaxOutputBytes: DefaultMaxOutput}
			if _, err := adapter.InspectMerge(context.Background(), request); err == nil {
				t.Fatalf("unsafe target %q passed", branch)
			}
		}
	})
}

func TestMergeCompareAndSwapRejectsConcurrentRefChange(t *testing.T) {
	repository, base, candidate := mergeRepository(t)
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	request := domain.MergeRequest{Root: repository, TargetBranch: "release-target", ExpectedOld: base, Candidate: candidate, MaxOutputBytes: DefaultMaxOutput}
	if _, err := adapter.InspectMerge(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	tree := strings.TrimSpace(runGit(t, repository, "rev-parse", base+"^{tree}"))
	concurrent := strings.TrimSpace(runGit(t, repository, "commit-tree", tree, "-m", "concurrent"))
	runGit(t, repository, "update-ref", "refs/heads/release-target", concurrent, base)
	if err := adapter.AdvanceMerge(context.Background(), request); err == nil {
		t.Fatal("AdvanceMerge accepted concurrent target change")
	}
	if got := strings.TrimSpace(runGit(t, repository, "show-ref", "--hash", "refs/heads/release-target")); got != concurrent {
		t.Fatalf("concurrent ref overwritten=%q want=%q", got, concurrent)
	}
}

func mergeRepository(t *testing.T) (string, string, string) {
	t.Helper()
	repository, base := initializedRepository(t)
	configureIdentity(t, repository)
	runGit(t, repository, "branch", "release-target", base)
	runGit(t, repository, "checkout", "-q", "-b", "candidate")
	writeFile(t, filepath.Join(repository, "candidate.txt"), "candidate")
	runGit(t, repository, "add", "candidate.txt")
	runGit(t, repository, "commit", "-q", "-m", "feat: candidate")
	candidate := strings.TrimSpace(runGit(t, repository, "rev-parse", "HEAD"))
	return repository, base, candidate
}
