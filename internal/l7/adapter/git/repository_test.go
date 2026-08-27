package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocateAndSnapshotUseGitIdentityAndBoundedChangedScope(t *testing.T) {
	repository, base := initializedRepository(t)
	writeFile(t, filepath.Join(repository, "committed.txt"), "committed")
	runGit(t, repository, "add", "committed.txt")
	runGit(t, repository, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "add committed path")
	writeFile(t, filepath.Join(repository, "README.md"), "dirty")
	writeFile(t, filepath.Join(repository, "untracked.txt"), "untracked")

	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	location, err := adapter.Locate(context.Background(), repository)
	if err != nil {
		t.Fatal(err)
	}
	if location.Root != repository || location.CommonDir != filepath.Join(repository, ".git") || !fullObjectID(location.Head) || !fullObjectID(location.Tree) {
		t.Fatalf("Locate()=%+v", location)
	}
	snapshot, err := adapter.Snapshot(context.Background(), repository, base)
	if err != nil {
		t.Fatal(err)
	}
	want := "README.md|committed.txt|untracked.txt"
	if strings.Join(snapshot.ChangedPaths, "|") != want || snapshot.Base != base || snapshot.Head != location.Head || snapshot.Tree != location.Tree {
		t.Fatalf("Snapshot()=%+v, want paths %s", snapshot, want)
	}
}

func TestLocateRejectsOutsideBareAndUnbornRepositories(t *testing.T) {
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	t.Run("outside", func(t *testing.T) {
		if _, err := adapter.Locate(context.Background(), physicalTemp(t)); err == nil {
			t.Fatal("Locate() unexpectedly accepted a non-repository")
		}
	})
	t.Run("bare", func(t *testing.T) {
		root := filepath.Join(physicalTemp(t), "bare.git")
		runGit(t, filepath.Dir(root), "init", "-q", "--bare", root)
		if _, err := adapter.Locate(context.Background(), root); err == nil || !strings.Contains(err.Error(), "worktree") {
			t.Fatalf("bare Locate() error=%v", err)
		}
	})
	t.Run("unborn", func(t *testing.T) {
		root := physicalTemp(t)
		runGit(t, root, "init", "-q")
		if _, err := adapter.Locate(context.Background(), root); err == nil || !strings.Contains(err.Error(), "initial commit") {
			t.Fatalf("unborn Locate() error=%v", err)
		}
	})
}

func TestSnapshotRejectsInvalidMissingAndNonAncestorBases(t *testing.T) {
	repository, _ := initializedRepository(t)
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	if _, err := adapter.Snapshot(context.Background(), repository, "--help"); err == nil || !strings.Contains(err.Error(), "full lowercase") {
		t.Fatalf("argument-like base error=%v", err)
	}
	missing := strings.Repeat("f", 40)
	if _, err := adapter.Snapshot(context.Background(), repository, missing); err == nil || !strings.Contains(err.Error(), "available Git commit") {
		t.Fatalf("missing base error=%v", err)
	}
	unrelated := strings.TrimSpace(runGit(t, repository, "commit-tree", "HEAD^{tree}", "-m", "unrelated"))
	if _, err := adapter.Snapshot(context.Background(), repository, unrelated); err == nil || !strings.Contains(err.Error(), "not an ancestor") {
		t.Fatalf("non-ancestor base error=%v", err)
	}
}

func TestSnapshotFailsClosedOnPathAndOutputBounds(t *testing.T) {
	repository, base := initializedRepository(t)
	for index := 0; index < 3; index++ {
		writeFile(t, filepath.Join(repository, fmt.Sprintf("path-%d", index)), "x")
	}
	adapter := testAdapter(t, DefaultMaxOutput, 2)
	if _, err := adapter.Snapshot(context.Background(), repository, base); err == nil || !strings.Contains(err.Error(), "path count") {
		t.Fatalf("path-bounded Snapshot() error=%v", err)
	}
	adapter = testAdapter(t, 8, DefaultMaxPaths)
	if _, err := adapter.Snapshot(context.Background(), repository, base); err == nil || !strings.Contains(err.Error(), "output exceeds") {
		t.Fatalf("output-bounded Snapshot() error=%v", err)
	}
}

func TestSnapshotRejectsUnsafeGitPath(t *testing.T) {
	repository, base := initializedRepository(t)
	writeFile(t, filepath.Join(repository, "unsafe\npath"), "x")
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	if _, err := adapter.Snapshot(context.Background(), repository, base); err == nil || !strings.Contains(err.Error(), "unsafe status path") {
		t.Fatalf("unsafe-path Snapshot() error=%v", err)
	}
}

func TestLocateHonorsCancellationBeforeGit(t *testing.T) {
	repository, _ := initializedRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	adapter := testAdapter(t, DefaultMaxOutput, DefaultMaxPaths)
	if _, err := adapter.Locate(ctx, repository); err == nil || err != context.Canceled {
		t.Fatalf("cancelled Locate() error=%v", err)
	}
}

func TestPorcelainParserRejectsAmbiguousRecords(t *testing.T) {
	for _, data := range [][]byte{
		[]byte("?? missing-terminator"),
		[]byte("short\x00"),
		[]byte("!! ignored\x00"),
		[]byte("?? ../escape\x00"),
	} {
		if _, err := parseStatus(data, 10); err == nil {
			t.Fatalf("parseStatus(%q) unexpectedly passed", data)
		}
	}
}

func BenchmarkParseStatus10000Paths(b *testing.B) {
	var fixture strings.Builder
	for index := 0; index < 10_000; index++ {
		fmt.Fprintf(&fixture, "?? packages/component-%05d/file.go%c", index, 0)
	}
	data := []byte(fixture.String())
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		paths, err := parseStatus(data, 10_000)
		if err != nil || len(paths) != 10_000 {
			b.Fatalf("parseStatus() paths=%d error=%v", len(paths), err)
		}
	}
}

func testAdapter(t *testing.T, maxOutput, maxPaths int) Adapter {
	t.Helper()
	adapter, err := New("", maxOutput, maxPaths)
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func initializedRepository(t *testing.T) (string, string) {
	t.Helper()
	root := physicalTemp(t)
	runGit(t, root, "init", "-q")
	writeFile(t, filepath.Join(root, "README.md"), "initial")
	runGit(t, root, "add", "README.md")
	runGit(t, root, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "initial")
	return root, strings.TrimSpace(runGit(t, root, "rev-parse", "HEAD"))
}

func physicalTemp(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	command.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return string(output)
}
