package main

import "testing"

func TestGitIdentityAndChangeSetComeFromGit(t *testing.T) {
	repository := newTestRepository(t)
	base := repository.rev("HEAD")
	repository.write("docs/note.md", "note\n")
	head := repository.commit("docs: note")

	git := gitRepository{root: repository.root}
	resolved, err := git.resolve(head)
	if err != nil || resolved != head {
		t.Fatalf("resolve: got %q err=%v", resolved, err)
	}
	tree, err := git.tree(head)
	if err != nil || len(tree) != 40 {
		t.Fatalf("tree: got %q err=%v", tree, err)
	}
	changes, err := git.changed(base, head)
	if err != nil || len(changes) != 1 || changes[0].Path != "docs/note.md" || changes[0].Status != "A" {
		t.Fatalf("changes: got %+v err=%v", changes, err)
	}
}

func TestSafeGitPathRejectsAliasesAndControlCharacters(t *testing.T) {
	for _, value := range []string{"", "/absolute", "../escape", "a/../b", "a\\b", "a\nb"} {
		if safeGitPath(value) {
			t.Errorf("unsafe path accepted: %q", value)
		}
	}
	if !safeGitPath("docs/artifacts/changes/change.md") {
		t.Fatal("canonical path rejected")
	}
}
