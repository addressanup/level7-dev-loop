package toolbroker

import (
	"context"
	"os"
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
	if _, err := patchPaths("--- a/ok\n+++ b/../../escape\n"); err == nil {
		t.Fatal("patch escape accepted")
	}
	paths, err := patchPaths("--- a/src/main.go\n+++ b/src/main.go\n")
	if err != nil || len(paths) != 1 || paths[0] != "src/main.go" {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
}

func FuzzPatchPathContainment(f *testing.F) {
	f.Add("--- a/src/main.go\n+++ b/src/main.go\n")
	f.Add("--- a/ok\n+++ b/../../escape\n")
	f.Fuzz(func(t *testing.T, patch string) {
		if len(patch) > 1<<20 {
			t.Skip()
		}
		paths, err := patchPaths(patch)
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
