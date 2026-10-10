package plugininstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginpackage"
)

func TestCleanInstallBothHostsRequiresConsentAndIsIdempotent(t *testing.T) {
	env := newTestEnv(t)
	env.bootstrapCLI()

	_, err := env.manager.Install(env.ctx, Request{Host: "codex"})
	blocked := mustBlocked(t, err, "--accept-unsigned --accept-unqualified")
	if !strings.Contains(blocked.Reason, "UNSIGNED and NOT NOTARIZED") || !strings.Contains(blocked.Reason, "WITHHELD") || !strings.Contains(blocked.Reason, "--yes does not waive") {
		t.Fatalf("consent refusal is not informative: %q", blocked.Reason)
	}
	if env.registered("codex", "1.0.1") || exists(env.manager.layout.StateFile()) {
		t.Fatal("refused install mutated host or installer state")
	}
	_, err = env.manager.Install(env.ctx, Request{Host: "codex", AcceptUnsigned: true})
	mustBlocked(t, err, "--accept-unqualified")

	report, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != OutcomePass || len(report.Hosts) != 1 || report.Hosts[0].Outcome != "installed" || report.Hosts[0].To != "1.0.1" {
		t.Fatalf("unexpected install report: %+v", report)
	}
	if !strings.Contains(report.Hosts[0].Reload, "new Codex task") || !strings.Contains(report.Hosts[0].Message, "attestation") {
		t.Fatalf("install report lacks reload advice or verification summary: %+v", report.Hosts[0])
	}
	if !env.registered("codex", "1.0.1") {
		t.Fatalf("codex host state does not show the verified root: %+v", env.hostState("codex"))
	}
	state := env.installerState()
	record := state.Hosts["codex"]
	if record.Version != "1.0.1" || record.Channel != "stable" || record.Pinned || record.Previous != nil || !record.AcceptedUnsigned || !record.AcceptedUnqualified || record.Scope != "user" {
		t.Fatalf("unexpected installer record: %+v", record)
	}
	receipt, err := pluginpackage.LoadReceipt(record.Receipt)
	if err != nil || receipt.SourceCommit != "bb"+fixtureCommit[2:] || receipt.Workflow != ".github/workflows/release.yml" || !receipt.Unsigned || !receipt.Unqualified {
		t.Fatalf("receipt is incomplete: %+v error=%v", receipt, err)
	}
	if err := pluginpackage.VerifyTree(record.Root, receipt.Files); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(env.home, ".local", "bin", "l7")
	if target, err := os.Readlink(link); err != nil || target != env.manager.layout.CLIBinary("1.0.1") {
		t.Fatalf("CLI link target=%q error=%v", target, err)
	}
	if _, err := env.manager.loadCLIReceipt("1.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, found := state.Hosts["claude"]; found {
		t.Fatal("install touched the unselected host")
	}

	logBefore := env.hostLog("codex")
	report, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	if err != nil || report.Hosts[0].Outcome != "current" {
		t.Fatalf("repeat install was not a no-op: %+v error=%v", report, err)
	}
	if countLines(env.hostLog("codex"), "plugin marketplace add") != countLines(logBefore, "plugin marketplace add") {
		t.Fatal("repeat install re-registered the marketplace")
	}

	report, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude"}))
	if err != nil || report.Hosts[0].Outcome != "installed" {
		t.Fatalf("claude install: %+v error=%v", report, err)
	}
	if !env.registered("claude", "1.0.1") || !strings.Contains(env.hostLog("claude"), "plugin install level7-dev-loop@level7-engineering-v1 --scope user --json") {
		t.Fatalf("claude registration did not use the scoped install command: %s", env.hostLog("claude"))
	}

	env.server.server.Close()
	status, err := env.offline().Status(env.ctx, Request{})
	if err != nil {
		t.Fatal(err)
	}
	if status.Outcome != OutcomePass || status.Status == nil || len(status.Status.Hosts) != 2 {
		t.Fatalf("offline status: %+v", status)
	}
	for _, hostStatus := range status.Status.Hosts {
		if !hostStatus.Managed || hostStatus.Version != "1.0.1" || hostStatus.Verification == nil || hostStatus.Verification.Tree != "verified" || len(hostStatus.Warnings) != 2 || hostStatus.HostObservation == nil || !hostStatus.HostObservation.PluginInstalled || hostStatus.HostDiagnostic != "" {
			t.Fatalf("host status is incomplete: %+v", hostStatus)
		}
		if !strings.Contains(hostStatus.Availability, "not checked") {
			t.Fatalf("offline status implied a network check: %+v", hostStatus)
		}
	}
	if !status.Status.CLI.Managed || status.Status.CLI.LinkTarget != env.manager.layout.CLIBinary("1.0.1") {
		t.Fatalf("CLI status: %+v", status.Status.CLI)
	}
	if _, err := env.offline().Status(env.ctx, Request{Check: true}); err == nil {
		t.Fatal("status --check succeeded without a release client")
	}
}

func TestStatusCheckReportsAvailableUpdates(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	if _, err := env.manager.Update(env.ctx, Request{Host: "claude", Unpin: true, AcceptUnsigned: true, AcceptUnqualified: true}); err != nil {
		t.Fatal(err)
	}
	report, err := env.manager.Status(env.ctx, Request{Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Status.Checked || report.State != "current" {
		t.Fatalf("status --check: %+v", report)
	}
	for _, hostStatus := range report.Status.Hosts {
		switch hostStatus.Host {
		case "codex":
			if !hostStatus.Pinned || hostStatus.UpdateAvailable || !strings.Contains(hostStatus.Availability, "pinned at 1.0.0") || !strings.Contains(hostStatus.Availability, "1.0.1") {
				t.Fatalf("pinned codex availability: %+v", hostStatus)
			}
		case "claude":
			if hostStatus.Version != "1.0.1" || hostStatus.UpdateAvailable || hostStatus.Availability != "current" {
				t.Fatalf("claude availability: %+v", hostStatus)
			}
		}
	}
	if _, err := env.manager.Update(env.ctx, Request{Host: "codex", Unpin: true, Version: "", AcceptUnsigned: true, AcceptUnqualified: true}); err != nil {
		t.Fatal(err)
	}
}

func TestInstallFromCacheRevalidatesBytes(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	cached := filepath.Join(env.manager.layout.Cache(), "v1.0.0", "level7-dev-loop-1.0.0-claude.zip")
	if err := os.MkdirAll(filepath.Dir(cached), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cached, []byte("stale cache content"), 0o600); err != nil {
		t.Fatal(err)
	}
	requests := len(env.server.requests)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	downloaded := false
	for _, path := range env.server.requests[requests:] {
		if strings.HasSuffix(path, "level7-dev-loop-1.0.0-claude.zip") {
			downloaded = true
		}
	}
	if !downloaded {
		t.Fatal("a stale cache entry was trusted without re-downloading")
	}
	if digest, _, err := localfileDigest(cached); err != nil || digest != fixtures(t)["1.0.0"].Digests["level7-dev-loop-1.0.0-claude.zip"] {
		t.Fatalf("cache was not refreshed with verified bytes: %v", err)
	}
}

func localfileDigest(path string) (string, int64, error) {
	data, err := localfile.Read(path, 256<<20)
	if err != nil {
		return "", 0, err
	}
	return digestHex(data), int64(len(data)), nil
}
