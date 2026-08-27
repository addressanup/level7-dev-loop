package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestBriefRenderParseRoundTrip(t *testing.T) {
	want := fixtureBrief()
	data, err := RenderBrief(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseBrief(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Tier != want.Tier || got.Base != want.Base || got.Path != want.Path || got.Problem != want.Problem || strings.Join(got.Scope, "|") != strings.Join(want.Scope, "|") || strings.Join(got.AcceptanceCriteria, "|") != strings.Join(want.AcceptanceCriteria, "|") || strings.Join(got.Risks, "|") != strings.Join(want.Risks, "|") || strings.Join(got.Rollback, "|") != strings.Join(want.Rollback, "|") {
		t.Fatalf("round trip mismatch\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestCreateBriefIsAtomicAndNeverOverwrites(t *testing.T) {
	root := physicalRoot(t)
	brief := fixtureBrief()
	if err := CreateBrief(root, brief); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(brief.Path)))
	if err != nil {
		t.Fatal(err)
	}
	brief.Problem = "replacement attempt"
	if err := CreateBrief(root, brief); !errors.Is(err, os.ErrExist) {
		t.Fatalf("second CreateBrief() error=%v, want os.ErrExist", err)
	}
	after, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(brief.Path)))
	if err != nil || string(after) != string(before) {
		t.Fatalf("brief overwritten: before=%q after=%q error=%v", before, after, err)
	}
	loaded, err := LoadBrief(root, fixtureBrief().Path)
	if err != nil || loaded.ID != fixtureBrief().ID {
		t.Fatalf("LoadBrief()=%+v error=%v", loaded, err)
	}
}

func TestBriefRejectsUnsafeOrIncompleteInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*domain.ChangeBrief)
	}{
		{"Tier 1 artifact", func(brief *domain.ChangeBrief) { brief.Tier = domain.TierRoutine }},
		{"unsafe ID", func(brief *domain.ChangeBrief) { brief.ID = "../escape"; brief.Path = BriefPath(brief.ID) }},
		{"abbreviated base", func(brief *domain.ChangeBrief) { brief.Base = "abc123" }},
		{"multiline problem", func(brief *domain.ChangeBrief) { brief.Problem = "first\nsecond" }},
		{"unsafe scope", func(brief *domain.ChangeBrief) { brief.Scope = []string{"../outside"} }},
		{"missing acceptance", func(brief *domain.ChangeBrief) { brief.AcceptanceCriteria = nil }},
		{"missing risk", func(brief *domain.ChangeBrief) { brief.Risks = nil }},
		{"missing rollback", func(brief *domain.ChangeBrief) { brief.Rollback = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			brief := fixtureBrief()
			test.mutate(&brief)
			if err := ValidateBrief(brief); err == nil {
				t.Fatal("ValidateBrief() unexpectedly passed")
			}
		})
	}
}

func TestParseBriefRejectsStructuralAmbiguity(t *testing.T) {
	valid, err := RenderBrief(fixtureBrief())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		data string
	}{
		{"duplicate heading", strings.Replace(string(valid), "## Rollback", "## Scope\n\n- `extra.go`\n\n## Rollback", 1)},
		{"unknown metadata", strings.Replace(string(valid), "| Base commit |", "| Unknown |", 1)},
		{"title mismatch", strings.Replace(string(valid), "# Level 7 change — example-change", "# Level 7 change — other", 1)},
		{"extra paragraph", strings.Replace(string(valid), "\n## Scope", "\nextra\n\n## Scope", 1)},
		{"missing newline", strings.TrimSuffix(string(valid), "\n")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseBrief([]byte(test.data)); err == nil {
				t.Fatal("ParseBrief() unexpectedly passed")
			}
		})
	}
}

func TestLoadBriefRejectsSymlinkAndOversizedInput(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(string, string) error
	}{
		{"symlink", func(root, destination string) error {
			target := filepath.Join(root, "target")
			if err := os.WriteFile(target, []byte("not a brief"), 0o600); err != nil {
				return err
			}
			return os.Symlink(target, destination)
		}},
		{"oversized", func(_ string, destination string) error {
			return os.WriteFile(destination, []byte(strings.Repeat("x", MaxBrief+1)), 0o600)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := physicalRoot(t)
			directory := filepath.Join(root, "docs", "artifacts", "changes")
			if err := os.MkdirAll(directory, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, "example-change.md")
			if err := test.setup(root, path); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadBrief(root, "docs/artifacts/changes/example-change.md"); err == nil {
				t.Fatal("LoadBrief() unexpectedly passed")
			}
		})
	}
}

func fixtureBrief() domain.ChangeBrief {
	return domain.ChangeBrief{
		ID:                 "example-change",
		Tier:               domain.TierProduct,
		Base:               strings.Repeat("a", 40),
		Path:               "docs/artifacts/changes/example-change.md",
		Problem:            "The current workflow loses concise change intent.",
		Scope:              []string{"README.md", "internal/example/**"},
		AcceptanceCriteria: []string{"Relevant tests pass.", "Status remains Git-derived."},
		Risks:              []string{"A stale base could misstate scope; reject non-ancestor bases."},
		Rollback:           []string{"Revert the implementation commits."},
	}
}
