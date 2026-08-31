package headless

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type executorFunc func(context.Context, domain.HeadlessManifest, domain.HeadlessWave, domain.HeadlessCheckpoint) (WaveOutcome, error)

func (function executorFunc) Execute(ctx context.Context, manifest domain.HeadlessManifest, wave domain.HeadlessWave, checkpoint domain.HeadlessCheckpoint) (WaveOutcome, error) {
	return function(ctx, manifest, wave, checkpoint)
}

func TestPlanRequiresMeasurableAcceptanceAndBindsDigest(t *testing.T) {
	planner := NewPlannerWith(func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) })
	request := PlanRequest{ObjectivePath: "concept.md", Objective: []byte("# Product\n\nAcceptance: Feature works end to end.\n"), BaseCommit: strings.Repeat("a", 40), TargetBranch: "main", AllowedPaths: []string{"internal/**"}, AllowedCommands: [][]string{{"go", "test", "./..."}}, ProviderPolicy: "balanced", NetworkPolicy: "gateway-only", LocalMerge: true}
	manifest, err := planner.Plan(request)
	if err != nil || manifest.Digest == "" || !manifest.StopBeforeDeploy || manifest.RiskCeiling != domain.TierProduct || len(manifest.Waves) != 1 {
		t.Fatalf("manifest=%#v err=%v", manifest, err)
	}
	request.Objective = []byte("# Product\n")
	if _, err := planner.Plan(request); err == nil {
		t.Fatal("objective without acceptance was accepted")
	}
}

func TestHeadlessCompletesMultipleWavesAndPersistsCheckpoint(t *testing.T) {
	common, _ := filepath.EvalSymlinks(t.TempDir())
	now := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	manifest := manifest(t, now, "## Wave one\nAcceptance: One.\n## Wave two\nAcceptance: Two.\n")
	if err := SaveManifest(common, manifest); err != nil {
		t.Fatal(err)
	}
	approval, err := Approve(common, manifest, manifest.Digest, "Anup Pandey", "Product Owner", now)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	checkpoint, err := NewEngineWith(func() time.Time { now = now.Add(time.Second); return now }, func(context.Context, time.Time) error { return nil }).Start(context.Background(), common, manifest, approval, executorFunc(func(_ context.Context, _ domain.HeadlessManifest, wave domain.HeadlessWave, _ domain.HeadlessCheckpoint) (WaveOutcome, error) {
		calls++
		return WaveOutcome{State: "complete", ProviderID: "codex", ModelID: "model", CandidateCommit: strings.Repeat(string(rune('a'+calls)), 40), Message: wave.ID + " done"}, nil
	}))
	if err != nil || calls != 2 || checkpoint.State != "complete" || !strings.Contains(checkpoint.Next, "release") {
		t.Fatalf("checkpoint=%#v calls=%d err=%v", checkpoint, calls, err)
	}
	loaded, err := LoadCheckpoint(common)
	if err != nil || loaded.Sequence != checkpoint.Sequence {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}

func TestHeadlessWaitsQuotaAndCircuitBreaksAfterThreeNoProgressFailures(t *testing.T) {
	common, _ := filepath.EvalSymlinks(t.TempDir())
	now := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	manifest := manifest(t, now, "Acceptance: Done.\n")
	if err := SaveManifest(common, manifest); err != nil {
		t.Fatal(err)
	}
	approval, _ := Approve(common, manifest, manifest.Digest, "Owner", "Product Owner", now)
	waits, calls := 0, 0
	engine := NewEngineWith(func() time.Time { return now }, func(context.Context, time.Time) error { waits++; return nil })
	checkpoint, err := engine.Start(context.Background(), common, manifest, approval, executorFunc(func(context.Context, domain.HeadlessManifest, domain.HeadlessWave, domain.HeadlessCheckpoint) (WaveOutcome, error) {
		calls++
		if calls == 1 {
			return WaveOutcome{State: "quota", QuotaResetAtUTC: now.Add(time.Hour).Format(time.RFC3339)}, nil
		}
		return WaveOutcome{State: "failed", FailureSignature: "same"}, nil
	}))
	if err != nil || waits != 1 || calls != 4 || checkpoint.State != "paused" || checkpoint.RepeatedFailures != 3 {
		t.Fatalf("checkpoint=%#v waits=%d calls=%d err=%v", checkpoint, waits, calls, err)
	}
}

func TestScopeExpansionPausesWithoutWideningManifest(t *testing.T) {
	common, _ := filepath.EvalSymlinks(t.TempDir())
	now := time.Now().UTC()
	manifest := manifest(t, now, "Acceptance: Done.\n")
	_ = SaveManifest(common, manifest)
	approval, _ := Approve(common, manifest, manifest.Digest, "Owner", "Product Owner", now)
	checkpoint, err := NewEngineWith(func() time.Time { return now }, nil).Start(context.Background(), common, manifest, approval, executorFunc(func(context.Context, domain.HeadlessManifest, domain.HeadlessWave, domain.HeadlessCheckpoint) (WaveOutcome, error) {
		return WaveOutcome{State: "scope-expanded", Message: "outside path"}, nil
	}))
	if err != nil || checkpoint.State != "paused" || !strings.Contains(checkpoint.Next, "revised manifest") {
		t.Fatalf("checkpoint=%#v err=%v", checkpoint, err)
	}
}

func manifest(t *testing.T, now time.Time, objective string) domain.HeadlessManifest {
	t.Helper()
	planner := NewPlannerWith(func() time.Time { return now })
	value, err := planner.Plan(PlanRequest{ObjectivePath: "features.md", Objective: []byte(objective), BaseCommit: strings.Repeat("a", 40), TargetBranch: "main", AllowedPaths: []string{"internal/**"}, AllowedCommands: [][]string{{"go", "test", "./..."}}, ProviderPolicy: "balanced", NetworkPolicy: "gateway-only", LocalMerge: true})
	if err != nil {
		t.Fatal(err)
	}
	return value
}
