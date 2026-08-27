package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTierOneStateRoundTripContainsOnlyContinuityFacts(t *testing.T) {
	common := physicalCommonDirectory(t)
	want := tierOneActive()
	if err := Save(common, want); err != nil {
		t.Fatal(err)
	}
	got, exists, err := Load(common)
	if err != nil || !exists || got.Kind != want.Kind || got.ID != want.ID || got.Tier != want.Tier || got.Base != want.Base || got.Problem != want.Problem || strings.Join(got.Scope, "|") != strings.Join(want.Scope, "|") {
		t.Fatalf("Load()=%+v exists=%v error=%v, want %+v", got, exists, err, want)
	}
}

func TestBriefBackedStateDoesNotDuplicateGitOrBriefFields(t *testing.T) {
	common := physicalCommonDirectory(t)
	active := domain.ActiveChange{Kind: domain.ActiveBrief, ID: "product-change", BriefPath: "docs/artifacts/changes/product-change.md"}
	if err := Save(common, active); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(common, "l7", "product", "active.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"base"`, `"scope"`, `"problem"`, `"tier"`, `"head"`, `"tree"`, `"status"`, `"approval"`} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("brief-backed state duplicates %s: %s", forbidden, data)
		}
	}
	got, exists, err := Load(common)
	if err != nil || !exists || got.Kind != active.Kind || got.ID != active.ID || got.BriefPath != active.BriefPath || got.Tier != 0 || got.Base != "" || got.Problem != "" || len(got.Scope) != 0 {
		t.Fatalf("Load()=%+v exists=%v error=%v, want %+v", got, exists, err, active)
	}
}

func TestSaveReplacesOnlyValidStateAndPreservesCorruption(t *testing.T) {
	common := physicalCommonDirectory(t)
	if err := Save(common, tierOneActive()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(common, "l7", "product", "active.json")
	corrupt := []byte(`{"schema":1,"kind":"tier-1","kind":"brief"}`)
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Save(common, domain.ActiveChange{Kind: domain.ActiveBrief, ID: "next", BriefPath: "docs/artifacts/changes/next.md"}); err == nil || !strings.Contains(err.Error(), "refuse to replace invalid") {
		t.Fatalf("Save() error=%v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(corrupt) {
		t.Fatalf("corrupt state was overwritten: data=%q error=%v", after, err)
	}
}

func TestLoadRejectsMalformedUnsafeAndOversizedState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(string) error
	}{
		{"unknown field", func(path string) error {
			return os.WriteFile(path, []byte(`{"schema":1,"kind":"brief","change_id":"x","brief_path":"docs/artifacts/changes/x.md","unknown":true}`), 0o600)
		}},
		{"trailing", func(path string) error { return os.WriteFile(path, []byte(`{} {}`), 0o600) }},
		{"oversized", func(path string) error {
			return os.WriteFile(path, []byte(strings.Repeat("x", MaxActiveFile+1)), 0o600)
		}},
		{"symlink", func(path string) error {
			target := filepath.Join(filepath.Dir(path), "target")
			if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
				return err
			}
			return os.Symlink(target, path)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			common := physicalCommonDirectory(t)
			directory := filepath.Join(common, "l7", "product")
			if err := os.MkdirAll(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := test.setup(filepath.Join(directory, "active.json")); err != nil {
				t.Fatal(err)
			}
			if _, _, err := Load(common); err == nil {
				t.Fatal("Load() unexpectedly passed")
			}
		})
	}
}

func TestActiveStateValidationRejectsConflictingForms(t *testing.T) {
	tests := []activeFile{
		{Schema: SchemaVersion, Kind: domain.ActiveTierOne, ChangeID: "x", Tier: domain.TierRoutine, Base: strings.Repeat("a", 40), Problem: "problem", Scope: []string{"safe.go"}, BriefPath: "docs/artifacts/changes/x.md"},
		{Schema: SchemaVersion, Kind: domain.ActiveBrief, ChangeID: "x", Tier: domain.TierProduct, BriefPath: "docs/artifacts/changes/x.md"},
		{Schema: SchemaVersion, Kind: domain.ActiveBrief, ChangeID: "x", BriefPath: "docs/artifacts/changes/other.md"},
		{Schema: SchemaVersion, Kind: "mutable-status", ChangeID: "x"},
	}
	for _, file := range tests {
		if err := validateActive(file); err == nil {
			t.Fatalf("validateActive(%+v) unexpectedly passed", file)
		}
	}
}

func TestAcquireSerializesMutationAndDoesNotDeadlockAfterRelease(t *testing.T) {
	common := physicalCommonDirectory(t)
	first, err := Acquire(common)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(common); err == nil {
		t.Fatal("concurrent Acquire() unexpectedly passed")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Acquire(common)
	if err != nil {
		t.Fatalf("released lock deadlocked: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

func tierOneActive() domain.ActiveChange {
	return domain.ActiveChange{
		Kind:    domain.ActiveTierOne,
		ID:      "routine-fix",
		Tier:    domain.TierRoutine,
		Base:    strings.Repeat("a", 40),
		Problem: "Fix a low-risk defect.",
		Scope:   []string{"internal/example/**"},
	}
}

func physicalCommonDirectory(t *testing.T) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}
