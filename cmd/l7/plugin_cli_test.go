package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The plugin commands must work without Git and without touching the user's
// real hosts: PATH is emptied so no host executable resolves.
func TestPluginCommandContract(t *testing.T) {
	home := t.TempDir()
	t.Setenv("L7_HOME", filepath.Join(home, "level7"))
	t.Setenv("L7_NETWORK", "off")
	t.Setenv("PATH", filepath.Join(home, "empty-bin"))
	if err := os.Mkdir(filepath.Join(home, "empty-bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		args     []string
		exit     int
		contains []string
	}{
		{"help", []string{"plugin"}, 0, []string{"PASS", `command="plugin"`, "Usage: l7 plugin"}},
		{"help flag", []string{"plugin", "--help"}, 0, []string{"PASS", "install   --host codex|claude"}},
		{"unknown subcommand", []string{"plugin", "bogus"}, 1, []string{"FAILED", "L7-PLUGIN-001", "unknown plugin command bogus"}},
		{"unknown option", []string{"plugin", "status", "--host", "codex"}, 1, []string{"FAILED", "L7-PLUGIN-001", "unknown or unsupported option --host"}},
		{"missing value", []string{"plugin", "install", "--host"}, 1, []string{"FAILED", "missing its value"}},
		{"duplicate option", []string{"plugin", "install", "--host", "codex", "--host", "claude"}, 1, []string{"FAILED", "duplicate option --host"}},
		{"install without host", []string{"plugin", "install"}, 2, []string{"BLOCKED", "L7-PLUGIN-002", "exactly one --host"}},
		{"update ambiguity", []string{"plugin", "update", "--all", "--host", "codex"}, 2, []string{"BLOCKED", "exactly one of --all or --host"}},
		{"status offline", []string{"plugin", "status"}, 0, []string{"PASS", `state="not-installed"`, "was not found on PATH"}},
		{"status check needs network", []string{"plugin", "status", "--check"}, 2, []string{"BLOCKED", "L7_NETWORK=off"}},
		{"rollback unmanaged", []string{"plugin", "rollback", "--host", "claude"}, 2, []string{"BLOCKED", "not managed"}},
		{"uninstall unmanaged", []string{"plugin", "uninstall", "--host", "codex", "--yes"}, 2, []string{"BLOCKED", "not managed"}},
		{"update all nothing managed", []string{"plugin", "update", "--all"}, 0, []string{"PASS", `state="nothing-managed"`}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run(context.Background(), test.args, &stdout, &stderr)
			if exit != test.exit {
				t.Fatalf("exit=%d, want %d; stdout=%q stderr=%q", exit, test.exit, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
			for _, substring := range test.contains {
				if !strings.Contains(stdout.String(), substring) {
					t.Fatalf("stdout=%q lacks %q", stdout.String(), substring)
				}
			}
		})
	}
	var stdout, stderr bytes.Buffer
	if exit := run(context.Background(), []string{"plugin", "status", "--json"}, &stdout, &stderr); exit != 0 {
		t.Fatalf("json status exit=%d stderr=%q", exit, stderr.String())
	}
	var envelope struct {
		Schema  int    `json:"schema"`
		Outcome string `json:"outcome"`
		Code    string `json:"code"`
		Command string `json:"command"`
		Data    struct {
			Status struct {
				Home  string `json:"home"`
				Hosts []struct {
					Host    string `json:"host"`
					Managed bool   `json:"managed"`
				} `json:"hosts"`
				NetworkChecked bool `json:"network_checked"`
			} `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope.Outcome != "PASS" || envelope.Code != "L7-PLUGIN-000" || envelope.Command != "plugin status" || len(envelope.Data.Status.Hosts) != 2 || envelope.Data.Status.NetworkChecked {
		t.Fatalf("json status envelope: %s error=%v", stdout.String(), err)
	}
	if !strings.HasPrefix(envelope.Data.Status.Home, filepath.Join(home, "level7")) && !strings.Contains(envelope.Data.Status.Home, "level7") {
		t.Fatalf("status home %q is not the L7_HOME override", envelope.Data.Status.Home)
	}
	stdout.Reset()
	if exit := run(context.Background(), []string{"plugin", "install", "--host", "codex", "--accept-unsigned", "--accept-unqualified", "--json"}, &stdout, &stderr); exit != 2 || !strings.Contains(stdout.String(), `"outcome":"BLOCKED"`) {
		t.Fatalf("install without a host executable: exit=%d stdout=%q", exit, stdout.String())
	}
	if entries, _ := os.ReadDir(filepath.Join(home, "level7", "plugin")); len(entries) > 1 {
		t.Fatalf("a blocked install created state: %v", entries)
	}
}

func TestPluginHelpIsListedInTopLevelHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exit := run(context.Background(), nil, &stdout, &stderr); exit != 0 || !strings.Contains(stdout.String(), "plugin install|update|status|rollback|uninstall") {
		t.Fatalf("top-level help lacks the plugin installer: exit=%d %q", exit, stdout.String())
	}
}

func TestPluginRequestParsing(t *testing.T) {
	request, err := parsePluginRequest([]string{"--host", "claude", "--channel", "prerelease", "--version", "1.1.0-dev", "--allow-downgrade", "--unpin", "--accept-unsigned", "--accept-unqualified", "--dry-run"}, pluginFlagsBySubcommand["update"])
	if err != nil || request.Host != "claude" || request.Channel != "prerelease" || request.Version != "1.1.0-dev" || !request.AllowDowngrade || !request.Unpin || !request.AcceptUnsigned || !request.AcceptUnqualified || !request.DryRun {
		t.Fatalf("request=%+v err=%v", request, err)
	}
	if _, err := parsePluginRequest([]string{"--version", "--dry-run"}, pluginFlagsBySubcommand["install"]); err == nil {
		t.Fatal("flag consumed as a value")
	}
	if _, err := parsePluginRequest([]string{"--yes"}, pluginFlagsBySubcommand["install"]); err == nil {
		t.Fatal("--yes accepted where it would be meaningless")
	}
}
