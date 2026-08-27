package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type testRepository struct {
	t    *testing.T
	root string
}

func newTestRepository(t *testing.T) *testRepository {
	t.Helper()
	repository := &testRepository{t: t, root: t.TempDir()}
	repository.git("init", "-q")
	repository.git("config", "user.name", "Level 7 Test")
	repository.git("config", "user.email", "level7@example.test")
	repository.write("README.md", "fixture\n")
	repository.commit("chore: base")
	return repository
}

func (repository *testRepository) write(relative, content string) {
	repository.t.Helper()
	name := filepath.Join(repository.root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		repository.t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		repository.t.Fatal(err)
	}
}

func (repository *testRepository) remove(relative string) {
	repository.t.Helper()
	if err := os.Remove(filepath.Join(repository.root, filepath.FromSlash(relative))); err != nil {
		repository.t.Fatal(err)
	}
}

func (repository *testRepository) git(arguments ...string) string {
	repository.t.Helper()
	command := exec.Command("git", append([]string{"-C", repository.root}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		repository.t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
	return string(output)
}

func (repository *testRepository) commit(message string) string {
	repository.t.Helper()
	repository.git("add", "-A")
	repository.git("commit", "-q", "-m", message)
	return repository.rev("HEAD")
}

func (repository *testRepository) rev(ref string) string {
	repository.t.Helper()
	return trim(repository.git("rev-parse", ref))
}

func (repository *testRepository) authority(kind, changeID string, value any) {
	repository.t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		repository.t.Fatal(err)
	}
	name := filepath.Join(repository.root, ".git", "l7", kind, changeID+".json")
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		repository.t.Fatal(err)
	}
	if err := os.WriteFile(name, append(data, '\n'), 0o600); err != nil {
		repository.t.Fatal(err)
	}
}

func trim(value string) string {
	for len(value) > 0 && (value[len(value)-1] == '\n' || value[len(value)-1] == '\r') {
		value = value[:len(value)-1]
	}
	return value
}

func briefDocument(id string, tier riskTier, base string, scope ...string) string {
	document := "# Change\n\n| Field | Value |\n|---|---|\n| Change ID | `" + id + "` |\n| Risk tier | `" + tier.String() + "` |\n| Base commit | `" + base + "` |\n\n## Problem\n\nProblem.\n\n## Scope\n\nScope.\n\n## Exact implementation file set\n\nAdd:\n\n"
	for _, relative := range scope {
		document += "- `" + relative + "`\n"
	}
	return document + "\n## Acceptance criteria\n\n- Works.\n\n## Risks and mitigations\n\n- Risk.\n\n## Rollback\n\nRevert.\n"
}

func evidenceDocument(changeID, commit, tree, result, reviewer string) string {
	return "# Evidence\n\n| Field | Value |\n|---|---|\n| Change ID | `" + changeID + "` |\n| Candidate commit | `" + commit + "` |\n| Candidate tree | `" + tree + "` |\n| Result | `" + result + "` |\n| Reviewer | `" + reviewer + "` |\n"
}

func rules(findings []finding) map[string]int {
	result := make(map[string]int)
	for _, item := range findings {
		result[item.rule]++
	}
	return result
}
