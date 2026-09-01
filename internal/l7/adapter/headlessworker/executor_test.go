package headlessworker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestWorkerTransitionsPersistCheckpointAndAppendEvents(t *testing.T) {
	root, common, _ := workerRepository(t)
	executor, err := New(root, common, orchestrationconfig.Default())
	if err != nil {
		t.Fatal(err)
	}
	manifest := domain.HeadlessManifest{ID: "headless-test", Digest: "sha256:" + strings.Repeat("a", 64)}
	wave := domain.HeadlessWave{ID: "wave-01"}
	progress, err := executor.loadProgress(manifest, wave)
	if err != nil {
		t.Fatal(err)
	}
	progress.Worktree = root
	if err := executor.transition(manifest, wave, &progress, "worktree-ready", "route implementation"); err != nil {
		t.Fatal(err)
	}
	progress.ImplementationSession = "session-1"
	if err := executor.transition(manifest, wave, &progress, "implementation-complete", "verify candidate"); err != nil {
		t.Fatal(err)
	}
	loaded, err := executor.loadProgress(manifest, wave)
	if err != nil || loaded.Sequence != 2 || loaded.ImplementationSession != "session-1" || loaded.Next != "verify candidate" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
	for sequence := 1; sequence <= 2; sequence++ {
		path, _ := executor.progressPath(manifest.ID, wave.ID, fmt.Sprintf("events/%08d.json", sequence))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("event %d: %v", sequence, err)
		}
	}
}

func TestWorkerAdoptsBoundWorktreeAndRecoversControlledCandidate(t *testing.T) {
	root, common, base := workerRepository(t)
	executor, err := New(root, common, orchestrationconfig.Default())
	if err != nil {
		t.Fatal(err)
	}
	manifest := domain.HeadlessManifest{ID: "headless-test", Digest: "sha256:" + strings.Repeat("b", 64)}
	wave := domain.HeadlessWave{ID: "wave-01"}
	worktree, err := executor.worktree(context.Background(), manifest, wave, base, domain.HeadlessCheckpoint{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(worktree, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktree, "src", "feature.go"), []byte("package src\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workerGit(t, worktree, "add", "src/feature.go")
	workerGit(t, worktree, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "feat(headless): complete wave-01")
	candidate, recovered := executor.recoverCandidate(context.Background(), worktree, base, wave, []string{"src/**"})
	if !recovered || len(candidate) != 40 {
		t.Fatalf("candidate=%q recovered=%t", candidate, recovered)
	}
	adopted, err := executor.worktree(context.Background(), manifest, wave, base, domain.HeadlessCheckpoint{}, worktree)
	if err != nil || adopted != worktree {
		t.Fatalf("adopted=%q err=%v", adopted, err)
	}
	if _, recovered := executor.recoverCandidate(context.Background(), worktree, base, wave, []string{"other/**"}); recovered {
		t.Fatal("candidate outside the restored scope was accepted")
	}
}

func TestRouteForAttemptAdvancesDeterministicallyWithoutSelfReview(t *testing.T) {
	snapshots := []domain.ProviderSnapshot{
		{ID: "provider-a", Kind: domain.ProviderKindCodexAppServer, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{{ID: "model-a", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, SupportsResume: true, Efforts: []domain.ReasoningEffort{domain.EffortHigh}, CostClass: 1, LatencyClass: 1, Verified: true}}},
		{ID: "provider-b", Kind: domain.ProviderKindClaudeCLI, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{{ID: "model-b", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, SupportsResume: true, Efforts: []domain.ReasoningEffort{domain.EffortHigh}, CostClass: 2, LatencyClass: 2, Verified: true}}},
		{ID: "provider-c", Kind: domain.ProviderKindOpenAIResponses, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{{ID: "model-c", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, SupportsResume: true, Efforts: []domain.ReasoningEffort{domain.EffortHigh}, CostClass: 3, LatencyClass: 3, Verified: true}}},
	}
	task := domain.TaskProfile{Schema: domain.OrchestrationSchema, ID: "wave/implement", Complexity: domain.ComplexityC3, RiskTier: domain.TierProduct, ContextTokens: 64_000, NeedsTools: true, NeedsEditing: true, NeedsResume: true}
	first := routeForAttempt(task, snapshots, 0)
	second := routeForAttempt(task, snapshots, 1)
	last := routeForAttempt(task, snapshots, 99)
	if first.ProviderID != "provider-a" || second.ProviderID != "provider-b" || last.ProviderID != "provider-c" || len(second.Escalations) != 1 {
		t.Fatalf("first=%#v second=%#v last=%#v", first, second, last)
	}
	review := task
	review.ID, review.NeedsEditing, review.NeedsResume, review.IndependentReview = "wave/review", false, false, true
	review.ImplementerProvider, review.ImplementerModel = "provider-a", "model-a"
	if selected := routeForAttempt(review, snapshots, 0); selected.ProviderID == "provider-a" && selected.ModelID == "model-a" {
		t.Fatalf("implementer selected as reviewer: %#v", selected)
	}
}

func workerRepository(t *testing.T) (string, string, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	workerGit(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workerGit(t, root, "add", "README.md")
	workerGit(t, root, "-c", "user.name=Level Seven", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "initial")
	base := strings.TrimSpace(workerGit(t, root, "rev-parse", "HEAD"))
	common := strings.TrimSpace(workerGit(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		t.Fatal(err)
	}
	return root, common, base
}

func workerGit(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	command.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return string(output)
}
