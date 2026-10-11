package plugininstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
)

func TestUpdateAllHonoursPinsAndTouchesOnlyManagedHosts(t *testing.T) {
	env := newTestEnv(t)
	env.bootstrapCLI()
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	report, err := env.manager.Update(env.ctx, accepted(Request{All: true}))
	if err != nil || report.Outcome != OutcomePass || len(report.Hosts) != 1 || report.Hosts[0].Outcome != "skipped-pinned" {
		t.Fatalf("update --all with a pin: %+v error=%v", report, err)
	}
	if _, found := env.installerState().Hosts["claude"]; found || env.hostState("claude").marketplace(pluginMarketplace) != nil {
		t.Fatal("update --all installed into an unselected host")
	}
	_, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "--unpin")

	report, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true}))
	if err != nil || report.Hosts[0].Outcome != "updated" || report.Hosts[0].From != "1.0.0" || report.Hosts[0].To != "1.0.1" {
		t.Fatalf("unpinned update: %+v error=%v", report, err)
	}
	record := env.installerState().Hosts["codex"]
	if record.Pinned || record.Previous == nil || record.Previous.Version != "1.0.0" || !env.registered("codex", "1.0.1") {
		t.Fatalf("update did not retain the previous version: %+v", record)
	}
	if !exists(record.Previous.Root) || !exists(env.manager.layout.CLIBinary("1.0.0")) {
		t.Fatal("retained version or CLI binary was pruned")
	}
	if target, _ := os.Readlink(filepath.Join(env.home, ".local", "bin", "l7")); target != env.manager.layout.CLIBinary("1.0.1") {
		t.Fatalf("CLI link was not advanced: %q", target)
	}

	report, err = env.manager.Update(env.ctx, accepted(Request{All: true}))
	if err != nil || report.Hosts[0].Outcome != "current" {
		t.Fatalf("update --all at the newest version was not a no-op: %+v error=%v", report, err)
	}

	env.server.server.Close()
	rollback, err := env.offline().Rollback(env.ctx, Request{Host: "codex"})
	if err != nil || rollback.Outcome != OutcomePass || rollback.Hosts[0].Outcome != "rolled-back" || rollback.Hosts[0].To != "1.0.0" {
		t.Fatalf("offline rollback: %+v error=%v", rollback, err)
	}
	record = env.installerState().Hosts["codex"]
	if record.Version != "1.0.0" || !record.Pinned || record.Previous == nil || record.Previous.Version != "1.0.1" || !env.registered("codex", "1.0.0") {
		t.Fatalf("rollback state: %+v", record)
	}
	if target, _ := os.Readlink(filepath.Join(env.home, ".local", "bin", "l7")); target != env.manager.layout.CLIBinary("1.0.0") {
		t.Fatalf("CLI link did not follow the rollback: %q", target)
	}
	rollback, err = env.offline().Rollback(env.ctx, Request{Host: "codex"})
	if err != nil || rollback.Hosts[0].To != "1.0.1" {
		t.Fatalf("roll forward: %+v error=%v", rollback, err)
	}
	_, err = env.offline().Rollback(env.ctx, Request{Host: "claude"})
	mustBlocked(t, err, "not managed")
}

func TestPrereleaseAndDowngradeNeedExplicitSelection(t *testing.T) {
	env := newTestEnv(t)
	plan, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", DryRun: true}))
	if err != nil || plan.Hosts[0].Outcome != "planned" || plan.Hosts[0].To != "1.0.1" {
		t.Fatalf("stable dry run: %+v error=%v", plan, err)
	}
	if exists(env.manager.layout.StateFile()) || env.hostState("codex").marketplace(pluginMarketplace) != nil || exists(env.manager.layout.Versions()) {
		t.Fatal("dry run changed state")
	}
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.1.0-dev"}))
	mustBlocked(t, err, "--channel prerelease")
	report, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.1.0-dev", Channel: "prerelease"}))
	if err != nil || report.Hosts[0].To != "1.1.0-dev" || !report.Hosts[0].Pinned {
		t.Fatalf("explicit prerelease install: %+v error=%v", report, err)
	}
	if receipt := env.installerState().Hosts["codex"]; receipt.Channel != "prerelease" {
		t.Fatalf("channel not recorded: %+v", receipt)
	}
	_, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", Channel: "stable", Unpin: true}))
	mustBlocked(t, err, "--allow-downgrade")
	report, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", Channel: "stable", Unpin: true, AllowDowngrade: true}))
	if err != nil || report.Hosts[0].To != "1.0.1" {
		t.Fatalf("explicit downgrade: %+v error=%v", report, err)
	}
	plan, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", Channel: "prerelease"}))
	if err != nil || plan.Hosts[0].To != "1.1.0-dev" || plan.Hosts[0].Outcome != "updated" {
		t.Fatalf("prerelease channel update: %+v error=%v", plan, err)
	}
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "9.9.9"}))
	mustFail(t, err, "release v9.9.9")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "nope"}))
	mustBlocked(t, err, "pinned version is invalid")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude", Channel: "nightly"}))
	mustBlocked(t, err, "stable or prerelease")
}

func TestUpdateAllContinuesPastOneHostFailureAndRestoresIt(t *testing.T) {
	env := newTestEnv(t)
	for _, host := range []string{"codex", "claude"} {
		if _, err := env.manager.Install(env.ctx, accepted(Request{Host: host, Version: "1.0.0"})); err != nil {
			t.Fatal(err)
		}
	}
	env.control("claude", "fail=plugin-add\nfailroot=/versions/1.0.1/")
	report, err := env.manager.Update(env.ctx, accepted(Request{All: true, Unpin: true}))
	if err != nil || report.Outcome != OutcomeFailed || len(report.Hosts) != 2 {
		t.Fatalf("partial update: %+v error=%v", report, err)
	}
	outcomes := map[string]HostReport{}
	for _, hostReport := range report.Hosts {
		outcomes[hostReport.Host] = hostReport
	}
	if outcomes["codex"].Outcome != "updated" || outcomes["claude"].Outcome != "failed" || !strings.Contains(outcomes["claude"].Message, "previous registration was restored and verified") {
		t.Fatalf("per-host outcomes: %+v", outcomes)
	}
	state := env.installerState()
	if state.Hosts["codex"].Version != "1.0.1" || state.Hosts["claude"].Version != "1.0.0" || !env.registered("codex", "1.0.1") || !env.registered("claude", "1.0.0") {
		t.Fatalf("state after partial failure: %+v", state.Hosts)
	}
	if exists(env.manager.layout.JournalFile()) {
		t.Fatal("journal was left behind after a verified restoration")
	}
	env.control("claude", "")
	report, err = env.manager.Update(env.ctx, accepted(Request{All: true, Unpin: true}))
	if err != nil || report.Outcome != OutcomePass || outcomesOf(report)["claude"] != "updated" || outcomesOf(report)["codex"] != "current" {
		t.Fatalf("retry after repair: %+v error=%v", report, err)
	}
	if !exists(env.manager.layout.Root("1.0.0", "codex")) || !exists(env.manager.layout.Root("1.0.0", "claude")) {
		t.Fatal("retained previous versions were pruned while still referenced")
	}
}

func outcomesOf(report Report) map[string]string {
	outcomes := map[string]string{}
	for _, hostReport := range report.Hosts {
		outcomes[hostReport.Host] = hostReport.Outcome
	}
	return outcomes
}

func TestRestorationFailureKeepsJournalAndBlocksUntilRecovered(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	env.control("codex", "fail=marketplace-add")
	_, err := env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true}))
	mustFail(t, err, "journal")
	if !exists(env.manager.layout.JournalFile()) {
		t.Fatal("journal was dropped although restoration failed")
	}
	if env.installerState().Hosts["codex"].Version != "1.0.0" {
		t.Fatal("state advanced although the host change failed")
	}
	status, err := env.manager.Status(env.ctx, Request{})
	if err != nil || status.State != "needs-recovery" || status.Status.Transaction == nil || status.Status.Transaction.Attention == "" {
		t.Fatalf("status did not surface the journal: %+v error=%v", status, err)
	}
	_, err = env.manager.Rollback(env.ctx, Request{Host: "codex"})
	mustFail(t, err, "recovery")
	env.control("codex", "")
	report, err := env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true}))
	if err != nil || report.Hosts[0].Outcome != "updated" || exists(env.manager.layout.JournalFile()) {
		t.Fatalf("recovery then update: %+v error=%v", report, err)
	}
	if !env.registered("codex", "1.0.1") {
		t.Fatal("host registration is not at the updated version after recovery")
	}
}

func TestInterruptedTransactionIsRecoveredBeforeMutation(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	record := env.installerState().Hosts["claude"]
	// Simulate a crash after the old registration was removed but before the
	// new one was added: the host has nothing registered and a journal remains.
	env.setHostState("claude", fakeHostState{})
	journal := Journal{
		Schema: 1, Operation: "update", Host: "claude", StartedAtUTC: time.Now().UTC().Format(time.RFC3339),
		From:  &Registration{Version: "1.0.0", Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin},
		To:    &Registration{Version: "1.0.1", Root: env.manager.layout.Root("1.0.1", "claude"), Marketplace: record.Marketplace, Plugin: record.Plugin},
		Steps: []Step{{Name: "plugin-remove", Status: "done"}, {Name: "marketplace-remove", Status: "done"}, {Name: "marketplace-add", Status: "pending"}, {Name: "plugin-add", Status: "pending"}},
	}
	if err := writeJournal(env.manager.layout, journal, true); err != nil {
		t.Fatal(err)
	}
	status, _ := env.manager.Status(env.ctx, Request{})
	if status.State != "needs-recovery" {
		t.Fatalf("status state=%q", status.State)
	}
	report, err := env.manager.Update(env.ctx, accepted(Request{Host: "claude", DryRun: true, Unpin: true}))
	if err != nil || report.Hosts[0].Outcome != "planned" {
		t.Fatalf("dry run after recovery: %+v error=%v", report, err)
	}
	if !env.registered("claude", "1.0.0") || exists(env.manager.layout.JournalFile()) {
		t.Fatalf("recovery did not restore the previous registration: %+v", env.hostState("claude"))
	}
	// A completed transaction whose journal survived is finished, not undone.
	env.setHostState("claude", fakeHostState{})
	if _, err := env.manager.Update(env.ctx, accepted(Request{Host: "claude", Unpin: true})); err != nil {
		t.Fatal(err)
	}
	record = env.installerState().Hosts["claude"]
	journal.To = &Registration{Version: record.Version, Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin}
	if err := writeJournal(env.manager.layout, journal, true); err != nil {
		t.Fatal(err)
	}
	env.setHostState("claude", fakeHostState{})
	if _, err := env.manager.Update(env.ctx, accepted(Request{Host: "claude", Version: "1.0.1", DryRun: true})); err != nil {
		t.Fatal(err)
	}
	if !env.registered("claude", "1.0.1") {
		t.Fatal("completed transaction was not re-applied during recovery")
	}
}

func TestConcurrentMutationIsRefusedWhileLockIsHeld(t *testing.T) {
	env := newTestEnv(t)
	if err := localfile.EnsureDirectory(env.manager.layout.Plugin(), 0o700); err != nil {
		t.Fatal(err)
	}
	lock, err := localfile.AcquireLock(env.manager.layout.LockFile())
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "lock")
	if _, err := env.manager.Status(env.ctx, Request{}); err != nil {
		t.Fatalf("status must not need the lock: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"})); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallRemovesOnlyOwnedRegistrationsAndVersions(t *testing.T) {
	env := newTestEnv(t)
	env.bootstrapCLI()
	for _, host := range []string{"codex", "claude"} {
		if _, err := env.manager.Install(env.ctx, accepted(Request{Host: host, Version: "1.0.0"})); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true})); err != nil {
		t.Fatal(err)
	}
	legacy := fakeHostState{Marketplaces: []fakeMarketplace{{Name: "level7-engineering", Root: "/frozen/v0.1.1"}}, Plugins: []fakePlugin{{Name: "level7-dev-loop", Marketplace: "level7-engineering"}}}
	current := env.hostState("codex")
	legacy.Marketplaces = append(legacy.Marketplaces, current.Marketplaces...)
	legacy.Plugins = append(legacy.Plugins, current.Plugins...)
	env.setHostState("codex", legacy)

	_, err := env.offline().Uninstall(env.ctx, Request{Host: "codex"})
	mustBlocked(t, err, "--yes")
	report, err := env.offline().Uninstall(env.ctx, Request{Host: "codex", Yes: true})
	if err != nil || report.Outcome != OutcomePass || report.Hosts[0].Outcome != "uninstalled" {
		t.Fatalf("uninstall codex: %+v error=%v", report, err)
	}
	after := env.hostState("codex")
	if after.marketplace("level7-engineering") == nil || len(after.Plugins) != 1 || after.Plugins[0].Marketplace != "level7-engineering" || after.marketplace(pluginMarketplace) != nil {
		t.Fatalf("uninstall touched the frozen legacy registration or left ours: %+v", after)
	}
	state := env.installerState()
	if _, found := state.Hosts["codex"]; found || state.Hosts["claude"].Version != "1.0.0" {
		t.Fatalf("state after uninstall: %+v", state.Hosts)
	}
	if exists(env.manager.layout.Root("1.0.1", "codex")) || exists(env.manager.layout.Root("1.0.0", "codex")) {
		t.Fatal("codex extractions survived although nothing references them")
	}
	if !exists(env.manager.layout.Root("1.0.0", "claude")) || !exists(env.manager.layout.CLIBinary("1.0.0")) {
		t.Fatal("claude's installation or the shared CLI was removed")
	}
	if !env.registered("claude", "1.0.0") {
		t.Fatal("claude registration was disturbed")
	}
	report, err = env.offline().Uninstall(env.ctx, Request{Host: "claude", Yes: true})
	if err != nil || len(report.Notes) == 0 || !strings.Contains(strings.Join(report.Notes, " "), "kept") {
		t.Fatalf("final uninstall: %+v error=%v", report, err)
	}
	if !exists(filepath.Join(env.home, ".local", "bin", "l7")) || !exists(env.manager.layout.StateFile()) {
		t.Fatal("uninstall deleted the shared CLI link or installer state")
	}
	_, err = env.offline().Uninstall(env.ctx, Request{Host: "claude", Yes: true})
	mustBlocked(t, err, "not managed")
}

func TestUninstallWithTerminalConfirmation(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"})); err != nil {
		t.Fatal(err)
	}
	manager := env.offline()
	answers := []bool{false, true}
	manager.options.Confirm = func(prompt string) (bool, error) {
		if !strings.Contains(prompt, "codex") {
			t.Fatalf("prompt=%q", prompt)
		}
		answer := answers[0]
		answers = answers[1:]
		return answer, nil
	}
	_, err := manager.Uninstall(env.ctx, Request{Host: "codex"})
	mustBlocked(t, err, "not confirmed")
	if !env.registered("codex", "1.0.1") {
		t.Fatal("declined confirmation changed the host")
	}
	if _, err := manager.Uninstall(env.ctx, Request{Host: "codex"}); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationAdoptsManualV1AndLeavesLegacyMarketplaceAlone(t *testing.T) {
	env := newTestEnv(t)
	manual := filepath.Join(env.home, "manual-v1-codex")
	if err := os.MkdirAll(filepath.Join(manual, ".agents", "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(manual, ".codex-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manual, ".agents", "plugins", "marketplace.json"), []byte(`{"name":"level7-engineering-v1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manual, ".codex-plugin", "plugin.json"), []byte(`{"name":"level7-dev-loop","version":"1.0.0-dev"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	env.setHostState("codex", fakeHostState{
		Marketplaces: []fakeMarketplace{{Name: "level7-engineering", Root: "/frozen/v0.1.1"}, {Name: pluginMarketplace, Root: manual}},
		Plugins:      []fakePlugin{{Name: pluginName, Marketplace: "level7-engineering"}, {Name: pluginName, Marketplace: pluginMarketplace}},
	})
	report, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", DryRun: true}))
	if err != nil || !strings.Contains(report.Hosts[0].Message, manual) {
		t.Fatalf("migration dry run: %+v error=%v", report, err)
	}
	report, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	if err != nil || report.Hosts[0].Outcome != "installed" || !strings.Contains(report.Hosts[0].Message, "left in place") {
		t.Fatalf("migration install: %+v error=%v", report, err)
	}
	after := env.hostState("codex")
	if after.marketplace("level7-engineering") == nil || after.marketplace("level7-engineering").Root != "/frozen/v0.1.1" || !env.registered("codex", "1.0.1") {
		t.Fatalf("migration disturbed the legacy marketplace or failed to register: %+v", after)
	}
	if !exists(filepath.Join(manual, ".codex-plugin", "plugin.json")) {
		t.Fatal("migration deleted the manual extraction")
	}
	if previous := env.installerState().Hosts["codex"].Previous; previous != nil {
		t.Fatalf("a manual root must not become a rollback target: %+v", previous)
	}

	env.setHostState("claude", fakeHostState{Marketplaces: []fakeMarketplace{{Name: pluginMarketplace, Root: manual}}})
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude"}))
	mustBlocked(t, err, "partial")
}

func TestMigrationFailureRestoresManualRegistration(t *testing.T) {
	env := newTestEnv(t)
	manual := filepath.Join(env.home, "manual-v1-claude")
	if err := os.MkdirAll(filepath.Join(manual, ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manual, ".claude-plugin", "marketplace.json"), []byte(`{"name":"level7-engineering-v1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manual, ".claude-plugin", "plugin.json"), []byte(`{"name":"level7-dev-loop","version":"1.0.0-dev"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	env.setHostState("claude", fakeHostState{Marketplaces: []fakeMarketplace{{Name: pluginMarketplace, Root: manual}}, Plugins: []fakePlugin{{Name: pluginName, Marketplace: pluginMarketplace}}})
	env.control("claude", "fail=plugin-add\nfailroot=/versions/")
	_, err := env.manager.Install(env.ctx, accepted(Request{Host: "claude"}))
	mustFail(t, err, "restored")
	after := env.hostState("claude")
	if after.marketplace(pluginMarketplace) == nil || after.marketplace(pluginMarketplace).Root != manual || len(after.Plugins) != 1 {
		t.Fatalf("manual registration was not restored: %+v", after)
	}
	if _, found := env.installerState().Hosts["claude"]; found {
		t.Fatal("failed migration recorded a managed host")
	}
}
