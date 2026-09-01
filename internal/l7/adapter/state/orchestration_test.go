package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestProviderAndRouteStateRoundTrip(t *testing.T) {
	common, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	snapshots := []domain.ProviderSnapshot{{Schema: domain.OrchestrationSchema, ID: "codex", Kind: domain.ProviderKindCodexAppServer, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{{ID: "model", Languages: []string{"*"}, ContextWindow: 100_000, Efforts: []domain.ReasoningEffort{domain.EffortMedium}, CostClass: 2, LatencyClass: 2, Verified: true}}, Diagnostic: "ok", Next: "route"}}
	if err := SaveProviderSnapshots(common, snapshots); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := LoadProviderSnapshots(common)
	if err != nil || !found || len(loaded) != 1 || loaded[0].Models[0].ID != "model" {
		t.Fatalf("loaded=%#v found=%t err=%v", loaded, found, err)
	}
	decision := domain.RouteDecision{Schema: domain.OrchestrationSchema, TaskID: "task", ProviderID: "codex", ModelID: "model", Effort: domain.EffortMedium, Policy: "balanced", Candidates: []domain.RouteCandidate{}, Fallbacks: []string{}, Escalations: []string{}, Next: "start"}
	if err := SaveRouteDecision(common, decision); err != nil {
		t.Fatal(err)
	}
	loadedDecision, found, err := LoadRouteDecision(common)
	if err != nil || !found || loadedDecision.ModelID != "model" {
		t.Fatalf("decision=%#v found=%t err=%v", loadedDecision, found, err)
	}
}

func TestOrchestrationStateRejectsUnknownField(t *testing.T) {
	common, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(common, "l7", "orchestration", "providers")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "snapshot.json"), []byte(`{"schema":1,"providers":[],"unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadProviderSnapshots(common); err == nil {
		t.Fatal("unknown provider state field was accepted")
	}
}

func TestProviderStateRejectsIncompleteVerifiedCapability(t *testing.T) {
	common, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	snapshot := domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: "gateway", Kind: domain.ProviderKindOpenAIResponses, Authentication: domain.AuthAuthenticated, Models: []domain.ModelCapability{{ID: "unbounded", Verified: true}}, Diagnostic: "bad", Next: "repair"}
	if err := SaveProviderSnapshots(common, []domain.ProviderSnapshot{snapshot}); err == nil {
		t.Fatal("incomplete verified model capability was persisted")
	}
}
