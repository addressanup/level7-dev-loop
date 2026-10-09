package forge

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge/forgetest"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && forgetest.Commands[os.Args[1]] {
		os.Exit(forgetest.Main(os.Args[1:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

func fakeClient(t *testing.T) (Client, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err == nil {
		executable, err = filepath.EvalSymlinks(executable)
	}
	if err != nil {
		t.Fatal(err)
	}
	client, err := New(executable, root, "github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	return client, root
}

func TestRepositoryComesFromTheRemoteURL(t *testing.T) {
	for value, want := range map[string]string{
		"https://github.com/owner/repo.git":         "github.com/owner/repo",
		"https://github.com/owner/repo":             "github.com/owner/repo",
		"git@github.com:owner/repo.git":             "github.com/owner/repo",
		"ssh://git@github.com/owner/repo.git":       "github.com/owner/repo",
		"ssh://git@ghe.example.com:2222/team/r.git": "ghe.example.com/team/r",
		"https://token@github.com/owner/repo.git":   "github.com/owner/repo",
	} {
		if got, err := RepositoryFromURL(value); err != nil || got != want {
			t.Fatalf("%s -> %q err=%v, want %q", value, got, err, want)
		}
	}
	for _, value := range []string{"/tmp/remote.git", "file:///tmp/r.git", "https://github.com/owner", "https://github.com/a/b/c", "git@github.com:-x/repo.git", "https://github.com/owner/re po.git"} {
		if got, err := RepositoryFromURL(value); err == nil {
			t.Fatalf("%s accepted as %q", value, got)
		}
	}
	executable, _ := os.Executable()
	if _, err := New(executable, t.TempDir(), "owner/repo"); err == nil {
		t.Fatal("a repository without a host was accepted")
	}
}

func updateState(t *testing.T, root string, change func(*forgetest.State)) {
	t.Helper()
	state, err := forgetest.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	change(&state)
	if err := forgetest.Save(root, state); err != nil {
		t.Fatal(err)
	}
}

func TestPullRequestLifecycleThroughGh(t *testing.T) {
	client, root := fakeClient(t)
	head, branch := strings.Repeat("a", 40), "l7/tasks/crew-0123456789ab-t01-a0"
	updateState(t, root, func(state *forgetest.State) { state.Heads[branch] = head })
	if err := client.Authenticated(context.Background()); err != nil {
		t.Fatal(err)
	}
	created, err := client.Create(context.Background(), "main", branch, "feat(crew): Fix login", "## Handoff\n\nverified")
	if err != nil || created.Number != 1 || created.State != "OPEN" || created.Head != head || created.Base != "main" || !PullURLValid(created.URL) {
		t.Fatalf("created = %+v err=%v", created, err)
	}
	found, err := client.FindByHead(context.Background(), branch)
	if err != nil || len(found) != 1 || found[0].Number != 1 {
		t.Fatalf("found = %+v err=%v", found, err)
	}
	if err := client.AddLabel(context.Background(), 1, "l7-risk-tier-2"); err != nil {
		t.Fatal(err)
	}
	updateState(t, root, func(state *forgetest.State) {
		state.PullRequests[0].MergeStateStatus = "CLEAN"
		state.PullRequests[0].StatusCheckRollup = []map[string]any{
			{"__typename": "CheckRun", "name": "test", "status": "COMPLETED", "conclusion": "SUCCESS", "detailsUrl": "https://github.com/owner/repo/actions/runs/7/job/8"},
			{"__typename": "StatusContext", "context": "ci/external", "state": "SUCCESS"},
		}
	})
	viewed, err := client.View(context.Background(), 1)
	if err != nil || viewed.MergeState != "CLEAN" || len(viewed.Checks) != 2 {
		t.Fatalf("viewed = %+v err=%v", viewed, err)
	}
	if summary, failing := Summarize(viewed.Checks); summary != "passed" || len(failing) != 0 {
		t.Fatalf("summary = %s %v", summary, failing)
	}
	if err := client.Merge(context.Background(), 1, "squash", strings.Repeat("b", 40)); err == nil {
		t.Fatal("merge at a different head was accepted")
	}
	if err := client.Merge(context.Background(), 1, "squash", head); err != nil {
		t.Fatal(err)
	}
	state, err := forgetest.Load(root)
	if err != nil || len(state.Merges) != 1 || state.Merges[0] != (forgetest.Merge{Number: 1, Method: "squash", Head: head}) || state.PullRequests[0].Labels[0] != "l7-risk-tier-2" {
		t.Fatalf("state = %+v err=%v", state, err)
	}
	for _, call := range state.Calls {
		for _, argument := range call {
			if argument == "--admin" || argument == "--auto" {
				t.Fatalf("gh was asked to bypass protection: %v", call)
			}
		}
		if call[0] != "auth" && strings.Join(call[len(call)-2:], " ") != "--repo github.com/owner/repo" {
			t.Fatalf("gh call does not name its repository: %v", call)
		}
	}
	merged, err := client.View(context.Background(), 1)
	if err != nil || merged.State != "MERGED" || merged.MergeCommit != strings.Repeat("f", 40) {
		t.Fatalf("merged = %+v err=%v", merged, err)
	}
}

func TestCheckNormalizationAndSummary(t *testing.T) {
	checks := []ghCheck{
		{Typename: "CheckRun", Name: "build", Status: "IN_PROGRESS"},
		{Typename: "CheckRun", Name: "lint", Status: "COMPLETED", Conclusion: "NEUTRAL"},
		{Typename: "CheckRun", Name: "docs", Status: "COMPLETED", Conclusion: "SKIPPED"},
		{Typename: "CheckRun", Name: "test", Status: "COMPLETED", Conclusion: "TIMED_OUT"},
		{Typename: "StatusContext", Context: "external", State: "PENDING"},
		{Typename: "StatusContext", Context: "deploy-preview", State: "ERROR"},
	}
	normalized := []Check{}
	for _, check := range checks {
		normalized = append(normalized, check.normalize())
	}
	statuses := []string{}
	for _, check := range normalized {
		statuses = append(statuses, check.Status)
	}
	if !reflect.DeepEqual(statuses, []string{"pending", "passed", "skipped", "failed", "pending", "failed"}) {
		t.Fatalf("statuses = %v", statuses)
	}
	if summary, failing := Summarize(normalized); summary != "failed" || !reflect.DeepEqual(failing, []string{"test", "deploy-preview"}) {
		t.Fatalf("summary = %s %v", summary, failing)
	}
	if summary, _ := Summarize(normalized[:3]); summary != "pending" {
		t.Fatalf("pending summary = %s", summary)
	}
	if summary, _ := Summarize(nil); summary != "none" {
		t.Fatalf("no-check summary = %s", summary)
	}
	if summary, _ := Summarize(normalized[2:3]); summary != "none" {
		t.Fatalf("checks that were all skipped never ran: %s", summary)
	}
}

func TestFailedLogIsBoundedToActionsJobs(t *testing.T) {
	client, root := fakeClient(t)
	updateState(t, root, func(state *forgetest.State) {
		state.Logs["42"] = strings.Repeat("noise\n", 500) + "FAIL: TestLogin\x1b[0m expected 200\n"
	})
	log := client.FailedLog(context.Background(), Check{Name: "test", Status: "failed", DetailsURL: "https://github.com/owner/repo/actions/runs/9/job/42"}, 64)
	if len(log) > 64 || !strings.Contains(log, "expected 200") || strings.Contains(log, "\x1b") {
		t.Fatalf("log = %q", log)
	}
	if log := client.FailedLog(context.Background(), Check{DetailsURL: "https://ci.example.com/build/42"}, 64); log != "" {
		t.Fatalf("a non-Actions check fetched a log: %q", log)
	}
}

func TestInvalidRequestsNeverReachGh(t *testing.T) {
	client, root := fakeClient(t)
	updateState(t, root, func(state *forgetest.State) { state.Unauthenticated = true })
	if err := client.Authenticated(context.Background()); err == nil {
		t.Fatal("an unauthenticated gh was accepted")
	}
	before, _ := forgetest.Load(root)
	for _, request := range []func() error{
		func() error { _, err := client.Create(context.Background(), "main", "--force", "t", "b"); return err },
		func() error {
			_, err := client.Create(context.Background(), "main", "l7/tasks/x", "two\nlines", "b")
			return err
		},
		func() error {
			_, err := client.Create(context.Background(), "main", "l7/tasks/x", "t", "nul\x00")
			return err
		},
		func() error { _, err := client.View(context.Background(), 0); return err },
		func() error { return client.Merge(context.Background(), 1, "admin", strings.Repeat("a", 40)) },
		func() error { return client.Merge(context.Background(), 1, "squash", "HEAD") },
		func() error { _, err := client.FindByHead(context.Background(), "refs/heads/main"); return err },
	} {
		if request() == nil {
			t.Fatal("an invalid gh request was accepted")
		}
	}
	after, _ := forgetest.Load(root)
	if len(after.Calls) != len(before.Calls) {
		t.Fatalf("invalid requests reached gh: %v", after.Calls[len(before.Calls):])
	}
	for _, value := range []string{"https://github.com/o/r/pull/12", "https://ghe.example.com/team/r/pull/3"} {
		if !PullURLValid(value) {
			t.Fatalf("URL %q rejected", value)
		}
	}
	for _, value := range []string{"http://github.com/o/r/pull/1", "https://github.com/o/r/issues/1", "https://github.com/o/r/pull/1?x=1", "file:///pull/1"} {
		if PullURLValid(value) {
			t.Fatalf("URL %q accepted", value)
		}
	}
}
