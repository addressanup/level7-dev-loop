package toolbroker

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

func TestBrokerReadsOnlyDeclaredNonSecretPaths(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	broker, err := NewWith(root, policy(), nil, func(string) (processadapter.Executable, error) { return processadapter.Executable{}, os.ErrNotExist }, func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := broker.Call(context.Background(), "read_file", []byte(`{"path":"src/main.go","offset":0,"limit":100}`))
	if err != nil || !strings.Contains(string(data), "package main") {
		t.Fatalf("read=%s err=%v", data, err)
	}
	data, err = broker.Call(context.Background(), "read_file", []byte(`{"path":".env","offset":0,"limit":100}`))
	if err != nil || !strings.Contains(string(data), `"ok":false`) || strings.Contains(string(data), "SECRET") {
		t.Fatalf("secret response=%s err=%v", data, err)
	}
	data, err = broker.Call(context.Background(), "read_file", []byte(`{"path":"../outside","offset":0,"limit":100}`))
	if err != nil || !strings.Contains(string(data), `"ok":false`) {
		t.Fatalf("escape response=%s err=%v", data, err)
	}
}

func TestBrokerRunsOnlyExactAllowlistedCommand(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	called := false
	broker, err := NewWith(root, policy(), nil,
		func(name string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("a", 64)}, nil
		},
		func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			called = true
			if request.Executable != "/usr/bin/go" || strings.Join(request.Arguments, " ") != "test ./..." || request.Directory != root {
				t.Fatalf("unsafe request: %#v", request)
			}
			return processadapter.Result{Stdout: []byte("ok\n")}, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	data, err := broker.Call(context.Background(), "run_command", []byte(`{"name":"test"}`))
	if err != nil || !called || !strings.Contains(string(data), `"ok":true`) {
		t.Fatalf("data=%s called=%t err=%v", data, called, err)
	}
	called = false
	data, err = broker.Call(context.Background(), "run_command", []byte(`{"name":"shell"}`))
	if err != nil || called || !strings.Contains(string(data), `"ok":false`) {
		t.Fatalf("data=%s called=%t err=%v", data, called, err)
	}
}

func TestPatchPathParserRejectsEscape(t *testing.T) {
	if _, err := patchPaths([]byte("1\t0\t../../escape\x00")); err == nil {
		t.Fatal("patch escape accepted")
	}
	if _, err := patchPaths([]byte("0\t0\t\x00src/old.go\x00src/new.go\x00")); err == nil {
		t.Fatal("rename/copy path tuple accepted")
	}
	if _, err := patchPaths([]byte("-\t-\tsrc/image.bin\x00")); err == nil {
		t.Fatal("binary path record accepted")
	}
	paths, err := patchPaths([]byte("1\t1\tsrc/main.go\x00"))
	if err != nil || len(paths) != 1 || paths[0] != "src/main.go" {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
}

func FuzzPatchPathContainment(f *testing.F) {
	f.Add([]byte("1\t1\tsrc/main.go\x00"))
	f.Add([]byte("1\t0\t../../escape\x00"))
	f.Add([]byte("0\t0\t\x00src/old.go\x00src/new.go\x00"))
	f.Fuzz(func(t *testing.T, numstat []byte) {
		if len(numstat) > 1<<20 {
			t.Skip()
		}
		paths, err := patchPaths(numstat)
		if err != nil {
			return
		}
		for _, relative := range paths {
			clean := filepath.ToSlash(filepath.Clean(relative))
			if clean != relative || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "\\") {
				t.Fatalf("unsafe successful path %q", relative)
			}
		}
	})
}

func TestBrokerAppliesOnlyCanonicalGitPreflightPaths(t *testing.T) {
	root := initializedPatchRepository(t)
	broker, err := New(root, policy(), nil)
	if err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/src/main.go b/src/main.go\n" +
		"--- a/src/main.go\n+++ b/src/main.go\n@@ -1 +1 @@\n-package old\n+package updated\n"
	data, err := callPatch(t, broker, patch)
	if err != nil || !strings.Contains(string(data), `"ok":true`) || !strings.Contains(string(data), `"src/main.go"`) {
		t.Fatalf("data=%s err=%v", data, err)
	}
	contents, err := os.ReadFile(filepath.Join(root, "src", "main.go"))
	if err != nil || string(contents) != "package updated\n" {
		t.Fatalf("contents=%q err=%v", contents, err)
	}
}

func TestBrokerRejectsEveryUnsupportedPatchOperationBeforeMutation(t *testing.T) {
	cases := map[string]string{
		"rename":                           "diff --git a/src/main.go b/src/renamed.go\nsimilarity index 100%\nrename from src/main.go\nrename to src/renamed.go\n",
		"copy":                             "diff --git a/src/main.go b/src/copied.go\nsimilarity index 100%\ncopy from src/main.go\ncopy to src/copied.go\n",
		"mode-only":                        "diff --git a/src/main.go b/src/main.go\nold mode 100644\nnew mode 100755\n",
		"symlink":                          "diff --git a/src/link b/src/link\nnew file mode 120000\n--- /dev/null\n+++ b/src/link\n@@ -0,0 +1 @@\n+main.go\n",
		"submodule":                        "diff --git a/src/module b/src/module\nnew file mode 160000\n--- /dev/null\n+++ b/src/module\n@@ -0,0 +1 @@\n+Subproject commit 0000000000000000000000000000000000000000\n",
		"binary":                           "diff --git a/src/blob b/src/blob\nnew file mode 100644\nindex 0000000..1111111\nGIT binary patch\nliteral 1\nKcmZQz00IC2\n",
		"different source and destination": "diff --git a/src/main.go b/src/other.go\n--- a/src/main.go\n+++ b/src/other.go\n@@ -1 +1 @@\n-package old\n+package other\n",
		"outside manifest":                 "diff --git a/docs/escape.md b/docs/escape.md\nnew file mode 100644\n--- /dev/null\n+++ b/docs/escape.md\n@@ -0,0 +1 @@\n+escape\n",
		"secret":                           "diff --git a/src/credential.txt b/src/credential.txt\nnew file mode 100644\n--- /dev/null\n+++ b/src/credential.txt\n@@ -0,0 +1 @@\n+secret\n",
	}
	for name, patch := range cases {
		t.Run(name, func(t *testing.T) {
			root := initializedPatchRepository(t)
			broker, err := New(root, policy(), nil)
			if err != nil {
				t.Fatal(err)
			}
			data, err := callPatch(t, broker, patch)
			if err != nil || !strings.Contains(string(data), `"ok":false`) {
				t.Fatalf("data=%s err=%v", data, err)
			}
			contents, readErr := os.ReadFile(filepath.Join(root, "src", "main.go"))
			if readErr != nil || string(contents) != "package old\n" {
				t.Fatalf("rejected patch mutated source: contents=%q err=%v", contents, readErr)
			}
			for _, unexpected := range []string{"renamed.go", "copied.go", "other.go", "link", "module", "blob", "credential.txt"} {
				if _, statErr := os.Lstat(filepath.Join(root, "src", unexpected)); !os.IsNotExist(statErr) {
					t.Fatalf("rejected patch created %s: %v", unexpected, statErr)
				}
			}
			if _, statErr := os.Lstat(filepath.Join(root, "docs", "escape.md")); !os.IsNotExist(statErr) {
				t.Fatalf("rejected patch created out-of-scope file: %v", statErr)
			}
		})
	}
}

func TestBrokerRejectsProtectedPathEvenWhenManifestAllowsIt(t *testing.T) {
	root := initializedPatchRepository(t)
	if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte("all:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	brokerPolicy := policy()
	brokerPolicy.AllowedPaths = append(brokerPolicy.AllowedPaths, "Makefile")
	broker, err := New(root, brokerPolicy, nil)
	if err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/Makefile b/Makefile\n--- a/Makefile\n+++ b/Makefile\n@@ -1 +1 @@\n-all:\n+unsafe:\n"
	data, err := callPatch(t, broker, patch)
	contents, readErr := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil || !strings.Contains(string(data), `"ok":false`) || readErr != nil || string(contents) != "all:\n" {
		t.Fatalf("data=%s contents=%q err=%v readErr=%v", data, contents, err, readErr)
	}
}

func TestBrokerRollsBackPostconditionPathMismatch(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls := 0
	broker, err := NewWith(root, policy(), nil,
		func(string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/git", Digest: strings.Repeat("a", 64)}, nil
		},
		func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			calls++
			switch calls {
			case 1:
				return processadapter.Result{Stdout: []byte("1\t1\tsrc/main.go\x00")}, nil
			case 2:
				return processadapter.Result{}, nil
			case 3:
				return processadapter.Result{Stdout: []byte("1\t1\tsrc/other.go\x00")}, nil
			case 4:
				if strings.Join(request.Arguments, " ") != "apply --reverse --whitespace=error-all -" {
					t.Fatalf("unexpected rollback request: %v", request.Arguments)
				}
				return processadapter.Result{}, nil
			case 5:
				if strings.Join(request.Arguments, " ") != "apply --check --numstat -z --whitespace=error-all -" {
					t.Fatalf("unexpected rollback check: %v", request.Arguments)
				}
				return processadapter.Result{Stdout: []byte("1\t1\tsrc/main.go\x00")}, nil
			default:
				t.Fatalf("unexpected Git call %d", calls)
				return processadapter.Result{}, nil
			}
		})
	if err != nil {
		t.Fatal(err)
	}
	data, err := callPatch(t, broker, "diff --git a/src/main.go b/src/main.go\n--- a/src/main.go\n+++ b/src/main.go\n@@ -1 +1 @@\n-old\n+new\n")
	if err != nil || calls != 5 || !strings.Contains(string(data), `"ok":false`) || !strings.Contains(string(data), "rolled back") {
		t.Fatalf("data=%s calls=%d err=%v", data, calls, err)
	}
}

func TestBrokerRejectsExistingSymlinkTargetBeforeMutation(t *testing.T) {
	root := initializedPatchRepository(t)
	link := filepath.Join(root, "src", "link")
	if err := os.Symlink("main.go", link); err != nil {
		t.Fatal(err)
	}
	gitFixtureCommand(t, root, "add", "src/link")
	gitFixtureCommand(t, root, "commit", "--quiet", "-m", "symlink fixture")
	broker, err := New(root, policy(), nil)
	if err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/src/link b/src/link\nindex 1f95dfd..b8e2c18 120000\n--- a/src/link\n+++ b/src/link\n@@ -1 +1 @@\n-main.go\n\\ No newline at end of file\n+other.go\n\\ No newline at end of file\n"
	data, err := callPatch(t, broker, patch)
	target, readlinkErr := os.Readlink(link)
	contents, readErr := os.ReadFile(filepath.Join(root, "src", "main.go"))
	if err != nil || !strings.Contains(string(data), `"ok":false`) || readlinkErr != nil || target != "main.go" || readErr != nil || string(contents) != "package old\n" {
		t.Fatalf("data=%s target=%q contents=%q err=%v readlinkErr=%v readErr=%v", data, target, contents, err, readlinkErr, readErr)
	}
}

func TestReadOnlyReviewerRejectsPatch(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	broker, err := NewReadOnly(root, policy(), nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := broker.Call(context.Background(), "apply_patch", []byte(`{"patch":"--- /dev/null\n+++ b/src/new.go\n@@ -0,0 +1 @@\n+package src\n"}`))
	if err != nil || !strings.Contains(string(data), `"ok":false`) || !strings.Contains(string(data), "read-only") {
		t.Fatalf("data=%s err=%v", data, err)
	}
	data, err = broker.Call(context.Background(), "run_command", []byte(`{"name":"test"}`))
	if err != nil || !strings.Contains(string(data), `"ok":false`) || !strings.Contains(string(data), "read-only") {
		t.Fatalf("read-only command data=%s err=%v", data, err)
	}
}

func policy() orchestrationconfig.Tools {
	return orchestrationconfig.Tools{
		AllowedPaths:    []string{"src/**"},
		AllowedCommands: []orchestrationconfig.Command{{Name: "test", Argv: []string{"go", "test", "./..."}}},
		MaxOutputBytes:  1 << 20, MaxSeconds: 30,
	}
}

func initializedPatchRepository(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{{"init", "--quiet"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Level 7 Test"}, {"add", "src/main.go"}, {"commit", "--quiet", "-m", "fixture"}} {
		gitFixtureCommand(t, root, arguments...)
	}
	return root
}

func gitFixtureCommand(t *testing.T, root string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = root
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
}

func callPatch(t *testing.T, broker Broker, patch string) ([]byte, error) {
	t.Helper()
	arguments, err := json.Marshal(map[string]string{"patch": patch})
	if err != nil {
		t.Fatal(err)
	}
	return broker.Call(context.Background(), "apply_patch", arguments)
}
