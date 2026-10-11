//go:build l7_actual_provider

package plugininstall

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginhost"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

// TestActualHostsInstallIntoIsolatedProfiles drives the real codex, claude,
// and gh executables against the real GitHub release, but only inside
// explicitly authorized, empty host profiles. It never reads or writes the
// user's own ~/.codex or ~/.claude, and it spends no provider quota: only
// plugin-management commands run.
func TestActualHostsInstallIntoIsolatedProfiles(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_PLUGIN_HOSTS") != "install-real-release-into-isolated-profiles" {
		t.Skip("set L7_AUTHORIZE_ACTUAL_PLUGIN_HOSTS=install-real-release-into-isolated-profiles with isolated profile directories")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("actual host validation requires macOS")
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	profiles := map[string]string{"CODEX_HOME": os.Getenv("L7_ACTUAL_PLUGIN_CODEX_HOME"), "CLAUDE_CONFIG_DIR": os.Getenv("L7_ACTUAL_PLUGIN_CLAUDE_CONFIG_DIR")}
	for variable, directory := range profiles {
		if directory == "" || !filepath.IsAbs(directory) {
			t.Fatalf("%s must name an absolute isolated profile directory", variable)
		}
		physical, err := filepath.EvalSymlinks(directory)
		if err != nil {
			t.Fatalf("%s=%s must exist: %v", variable, directory, err)
		}
		for _, real := range []string{filepath.Join(userHome, ".codex"), filepath.Join(userHome, ".claude")} {
			realPhysical, err := filepath.EvalSymlinks(real)
			if err != nil {
				realPhysical = filepath.Clean(real)
			}
			if physical == realPhysical || strings.HasPrefix(physical, realPhysical+string(filepath.Separator)) {
				t.Fatalf("%s=%s points into the user's real host profile; refusing", variable, directory)
			}
		}
		entries, err := os.ReadDir(physical)
		if err != nil || len(entries) != 0 {
			t.Fatalf("%s=%s must be an empty directory (entries=%d err=%v)", variable, directory, len(entries), err)
		}
		t.Setenv(variable, physical)
	}
	home := t.TempDir()
	manager, err := New(Options{
		Home:       filepath.Join(home, "level7"),
		Platform:   Platform{OS: runtime.GOOS, Arch: runtime.GOARCH},
		CLIVersion: "actual-host-test",
		Releases:   pluginrelease.NewClient(nil, pluginrelease.DefaultSource(), pluginrelease.NewVerifier(nil, nil), "actual-host-test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, host := range pluginhost.All() {
		report, err := manager.Install(ctx, Request{Host: string(host), Version: "1.0.0", AcceptUnsigned: true, AcceptUnqualified: true})
		if err != nil {
			t.Fatalf("%s install: %v", host, err)
		}
		t.Logf("%s install: %s", host, report.Message)
		if report.Hosts[0].Outcome != "installed" {
			t.Fatalf("%s outcome=%s", host, report.Hosts[0].Outcome)
		}
	}
	status, err := manager.Status(ctx, Request{})
	if err != nil {
		t.Fatal(err)
	}
	for _, hostStatus := range status.Status.Hosts {
		if !hostStatus.Managed || hostStatus.HostObservation == nil || !hostStatus.HostObservation.PluginInstalled || hostStatus.HostDiagnostic != "" {
			t.Fatalf("%s status after install: %+v", hostStatus.Host, hostStatus)
		}
		t.Logf("%s host=%s plugin=%s root=%s", hostStatus.Host, hostStatus.HostVersion, hostStatus.HostObservation.PluginVersion, hostStatus.HostObservation.MarketplaceRoot)
	}
	for _, host := range pluginhost.All() {
		report, err := manager.Uninstall(ctx, Request{Host: string(host), Yes: true})
		if err != nil {
			t.Fatalf("%s uninstall: %v", host, err)
		}
		t.Logf("%s uninstall: %s", host, report.Message)
	}
	status, err = manager.Status(ctx, Request{})
	if err != nil || len(status.Status.Hosts) != 2 {
		t.Fatal(err)
	}
	for _, hostStatus := range status.Status.Hosts {
		if hostStatus.Managed {
			t.Fatalf("%s still managed after uninstall", hostStatus.Host)
		}
	}
}
