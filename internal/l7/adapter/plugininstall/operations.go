package plugininstall

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginhost"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginpackage"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

// Request carries the parsed `l7 plugin` options.
type Request struct {
	Host              string
	All               bool
	Channel           string
	Version           string
	Unpin             bool
	AllowDowngrade    bool
	AcceptUnsigned    bool
	AcceptUnqualified bool
	DryRun            bool
	Yes               bool
	Check             bool
}

// HostReport is the outcome for one host.
type HostReport struct {
	Host     string   `json:"host"`
	Outcome  string   `json:"outcome"`
	Message  string   `json:"message"`
	From     string   `json:"from_version,omitempty"`
	To       string   `json:"to_version,omitempty"`
	Root     string   `json:"root,omitempty"`
	Pinned   bool     `json:"pinned"`
	Channel  string   `json:"channel,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Reload   string   `json:"reload,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// Report is one operation's result in a form the CLI renders directly.
type Report struct {
	Outcome string       `json:"outcome"`
	State   string       `json:"state"`
	Message string       `json:"message"`
	Next    string       `json:"next"`
	Hosts   []HostReport `json:"hosts,omitempty"`
	CLI     string       `json:"cli_version,omitempty"`
	Notes   []string     `json:"notes,omitempty"`
	Status  *Status      `json:"status,omitempty"`
}

const (
	OutcomePass    = "PASS"
	OutcomeBlocked = "BLOCKED"
	OutcomeFailed  = "FAILED"
)

type resolvedTarget struct {
	resolved pluginrelease.Resolved
	channel  pluginrelease.Channel
	pinned   bool
}

// Install installs or brings one host to the selected release.
func (manager *Manager) Install(ctx context.Context, request Request) (Report, error) {
	host, err := parseHost(request.Host)
	if err != nil || request.All {
		return Report{}, &Blocked{Reason: "install requires exactly one --host codex|claude", Next: "run l7 plugin install --host codex or --host claude"}
	}
	return manager.mutateHosts(ctx, request, []pluginhost.Host{host}, true, "install")
}

// Update updates the named host or every already-managed host. It never
// installs into a host the user did not select earlier.
func (manager *Manager) Update(ctx context.Context, request Request) (Report, error) {
	if request.All == (request.Host != "") {
		return Report{}, &Blocked{Reason: "update requires exactly one of --all or --host codex|claude", Next: "run l7 plugin update --all"}
	}
	var hosts []pluginhost.Host
	if request.All {
		state, _, err := loadState(manager.layout)
		if err != nil {
			return Report{}, err
		}
		for _, name := range state.hostNames() {
			host, _ := parseHost(name)
			hosts = append(hosts, host)
		}
		if len(hosts) == 0 {
			return Report{Outcome: OutcomePass, State: "nothing-managed", Message: "no host is managed by this installer; nothing to update", Next: "run l7 plugin install --host codex|claude"}, nil
		}
	} else {
		host, err := parseHost(request.Host)
		if err != nil {
			return Report{}, &Blocked{Reason: err.Error(), Next: "run l7 plugin update --host codex|claude"}
		}
		hosts = []pluginhost.Host{host}
	}
	return manager.mutateHosts(ctx, request, hosts, false, "update")
}

func (manager *Manager) mutateHosts(ctx context.Context, request Request, hosts []pluginhost.Host, allowFresh bool, operation string) (Report, error) {
	if err := manager.options.Platform.supported(); err != nil {
		return Report{}, err
	}
	if request.Channel != "" && !pluginrelease.Channel(request.Channel).Valid() {
		return Report{}, &Blocked{Reason: "channel must be stable or prerelease", Next: "pass --channel stable or --channel prerelease"}
	}
	if request.Version != "" {
		if _, err := pluginrelease.ParseVersion(request.Version); err != nil {
			return Report{}, &Blocked{Reason: "pinned version is invalid: " + err.Error(), Next: "pass --version MAJOR.MINOR.PATCH[-prerelease]"}
		}
	}
	unlock, err := manager.lock()
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = unlock() }()
	manager.clearStaleStaging()
	state, _, err := loadState(manager.layout)
	if err != nil {
		return Report{}, err
	}
	if err := manager.recover(ctx, state); err != nil {
		return Report{}, err
	}
	report := Report{Outcome: OutcomePass, State: "installed", CLI: manager.options.CLIVersion}
	if operation == "update" {
		report.State = "updated"
	}
	if request.DryRun {
		report.State = "planned"
	}
	changed := false
	blocked, failed := 0, 0
	var firstErr error
	for _, host := range hosts {
		hostReport, mutated, err := manager.applyHost(ctx, &state, host, request, allowFresh, operation)
		changed = changed || mutated
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			var blockedErr *Blocked
			if errors.As(err, &blockedErr) {
				blocked++
				hostReport.Outcome, hostReport.Message, hostReport.Error = "blocked", blockedErr.Reason, blockedErr.Next
			} else {
				failed++
				hostReport.Outcome, hostReport.Message = "failed", err.Error()
			}
		}
		if hostReport.Host == "" {
			hostReport.Host = string(host)
		}
		report.Hosts = append(report.Hosts, hostReport)
		if mutated {
			if err := saveState(manager.layout, state); err != nil {
				return Report{}, fmt.Errorf("host change succeeded but installer state could not be saved: %w", err)
			}
		}
	}
	if changed {
		pointed, err := manager.pointCLI(state)
		if err != nil {
			report.Notes = append(report.Notes, "CLI link was not updated: "+err.Error())
		} else if pointed != "" {
			report.Notes = append(report.Notes, "l7 CLI link now resolves to version "+pointed+"; the binary that ran this command was not modified")
		}
		report.Notes = append(report.Notes, manager.prune(state, pointed)...)
	}
	switch {
	case len(hosts) == 1 && !changed && firstErr != nil:
		return Report{}, firstErr
	case failed > 0:
		report.Outcome, report.State, report.Message, report.Next = OutcomeFailed, "partial", fmt.Sprintf("%d host operation(s) failed; successful hosts kept their new state and failed hosts kept their previous state", failed), "read the per-host errors, repair the cause, and rerun the same command"
	case blocked > 0:
		report.Outcome, report.State, report.Message, report.Next = OutcomeBlocked, "partial", fmt.Sprintf("%d host operation(s) were blocked; see the per-host reasons", blocked), "resolve the blocked reasons and rerun"
	default:
		report.Message = summarize(report.Hosts)
		report.Next = "run l7 plugin status"
		if !changed && !request.DryRun {
			report.State = "current"
		}
	}
	return report, nil
}

func summarize(reports []HostReport) string {
	parts := make([]string, 0, len(reports))
	for _, report := range reports {
		parts = append(parts, report.Host+": "+report.Message)
	}
	return strings.Join(parts, "; ")
}

func (manager *Manager) applyHost(ctx context.Context, state *State, host pluginhost.Host, request Request, allowFresh bool, operation string) (HostReport, bool, error) {
	report := HostReport{Host: string(host)}
	record, managed := state.Hosts[string(host)]
	if !managed && !allowFresh {
		return report, false, &Blocked{Reason: fmt.Sprintf("%s is not managed by this installer", host), Next: "run l7 plugin install --host " + string(host) + " first; update --all never installs into an unselected host"}
	}
	port, executable, err := manager.hostPort(host)
	if err != nil {
		return report, false, err
	}
	hostVersion, err := port.Version(ctx, executable)
	if err != nil {
		return report, false, &Blocked{Reason: err.Error(), Next: "update or repair the host, then retry"}
	}
	// Observe before resolving: a registration the installer does not own is
	// adopted (migrated) only when it is complete, and never deleted from disk.
	observation, err := port.Observe(ctx, executable, pluginpackage.MarketplaceName, pluginpackage.PluginName)
	if err != nil {
		return report, false, err
	}
	var from *Registration
	migrated := ""
	switch {
	case managed:
		from = &Registration{Version: record.Version, Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin}
	case observation.MarketplaceRegistered && observation.PluginInstalled && observation.MarketplaceRoot != "":
		from = &Registration{Version: observation.PluginVersion, Root: observation.MarketplaceRoot, Marketplace: pluginpackage.MarketplaceName, Plugin: pluginpackage.PluginName}
		migrated = fmt.Sprintf("replaced the manually registered %s %s marketplace at %s; that directory was left in place", observation.PluginVersion, pluginpackage.MarketplaceName, observation.MarketplaceRoot)
	case observation.MarketplaceRegistered || observation.PluginInstalled:
		return report, false, &Blocked{Reason: fmt.Sprintf("%s has a partial %s registration that this installer does not manage", host, pluginpackage.MarketplaceName), Next: "remove it with the host's own plugin and marketplace commands, then rerun install"}
	}
	channel := pluginrelease.ChannelStable
	if managed && record.Channel != "" {
		channel = pluginrelease.Channel(record.Channel)
	}
	if request.Channel != "" {
		channel = pluginrelease.Channel(request.Channel)
	}
	exact := request.Version
	pinned := request.Version != ""
	if exact == "" && managed && record.Pinned {
		if request.Unpin {
			pinned = false
		} else if request.All {
			report.Outcome, report.Message, report.From, report.Pinned, report.Channel = "skipped-pinned", "pinned at "+record.Version+"; update --all leaves pins alone", record.Version, true, record.Channel
			return report, false, nil
		} else {
			return report, false, &Blocked{Reason: fmt.Sprintf("%s is pinned at %s", host, record.Version), Next: "pass --version <exact> to move the pin or --unpin to follow the " + string(channel) + " channel"}
		}
	}
	report.Channel, report.Pinned = string(channel), pinned
	target, err := manager.resolveTarget(ctx, channel, exact)
	if err != nil {
		return report, false, err
	}
	targetVersion := target.resolved.Contract.Version.String()
	report.To = targetVersion
	report.Warnings = target.resolved.Contract.Qualification.Warnings()
	repair := false
	if managed {
		report.From = record.Version
		installed, err := pluginrelease.ParseVersion(record.Version)
		if err != nil {
			return report, false, fmt.Errorf("installer state records an unparsable version %q", record.Version)
		}
		comparison := target.resolved.Contract.Version.Compare(installed)
		if comparison == 0 {
			if registrationMatches(observation, from) {
				report.Outcome, report.Message, report.Root = "current", "already at "+record.Version+"; nothing to do", record.Root
				if record.Pinned != pinned || record.Channel != string(channel) {
					if request.DryRun {
						report.Message += " (pin or channel change not saved in dry run)"
						return report, false, nil
					}
					record.Pinned, record.Channel = pinned, string(channel)
					state.Hosts[string(host)] = record
					report.Message = "already at " + record.Version + "; recorded pin/channel change only"
					return report, true, nil
				}
				return report, false, nil
			}
			repair = true
			if observation.MarketplaceRegistered && observation.PluginInstalled && observation.MarketplaceRoot != "" {
				from = &Registration{Version: observation.PluginVersion, Root: observation.MarketplaceRoot, Marketplace: pluginpackage.MarketplaceName, Plugin: pluginpackage.PluginName}
			} else if observation.MarketplaceRegistered || observation.PluginInstalled {
				return report, false, &Blocked{Reason: fmt.Sprintf("%s reports a partial %s registration that differs from the installer record", host, pluginpackage.MarketplaceName), Next: "remove it with the host's own plugin and marketplace commands, then rerun"}
			} else {
				from = nil
			}
		}
		if comparison < 0 && !request.AllowDowngrade {
			return report, false, &Blocked{Reason: fmt.Sprintf("%s would move %s from %s down to %s", operation, host, record.Version, targetVersion), Next: "pass --allow-downgrade to accept the downgrade explicitly"}
		}
	}
	if err := requireConsent(target.resolved.Contract.Qualification, request); err != nil {
		return report, false, err
	}
	if request.DryRun {
		report.Outcome, report.Message = "planned", fmt.Sprintf("dry run: would %s %s %s (%s, %s channel) from release %s at commit %s without changing anything", operation, host, targetVersion, target.resolved.Contract.Workflow.Path, channel, target.resolved.Release.Tag, target.resolved.Contract.SourceCommit)
		if migrated != "" {
			report.Message += "; would also replace the manually registered marketplace at " + observation.MarketplaceRoot
		}
		report.Root = manager.layout.Root(targetVersion, string(host))
		return report, false, nil
	}
	stage, cleanup, err := manager.stage()
	if err != nil {
		return report, false, err
	}
	defer cleanup()
	root, receiptPath, receipt, err := manager.ensureVersion(ctx, target.resolved, host, stage, acceptance{request.AcceptUnsigned, request.AcceptUnqualified})
	if err != nil {
		return report, false, err
	}
	report.Root = root
	to := &Registration{Version: targetVersion, Root: root, Marketplace: receipt.Marketplace, Plugin: receipt.Plugin}
	if err := manager.switchRegistration(ctx, port, executable, operation, from, to, observation); err != nil {
		return report, false, err
	}
	next := HostRecord{
		Version: targetVersion, Channel: string(channel), Pinned: pinned, Root: root, Receipt: receiptPath,
		Marketplace: receipt.Marketplace, Plugin: receipt.Plugin, Scope: pluginhost.Scope, HostVersion: hostVersion,
		InstalledAtUTC: manager.options.Now().UTC().Format(time.RFC3339), AcceptedUnsigned: request.AcceptUnsigned, AcceptedUnqualified: request.AcceptUnqualified,
	}
	if managed && !repair {
		retained := record.retained()
		next.Previous = &retained
	} else if managed {
		next.Previous = record.Previous
		next.AcceptedUnsigned, next.AcceptedUnqualified = record.AcceptedUnsigned || request.AcceptUnsigned, record.AcceptedUnqualified || request.AcceptUnqualified
	}
	state.Hosts[string(host)] = next
	commit := target.resolved.Contract.SourceCommit[:12]
	switch {
	case repair:
		report.Outcome = "repaired"
		report.Message = fmt.Sprintf("re-registered %s %s from its verified installation (release %s, source commit %s)", host, targetVersion, target.resolved.Release.Tag, commit)
	case managed:
		report.Outcome = "updated"
		report.Message = fmt.Sprintf("updated %s from %s to %s (release %s, source commit %s); %s retained for rollback", host, record.Version, targetVersion, target.resolved.Release.Tag, commit, record.Version)
	default:
		report.Outcome = "installed"
		report.Message = fmt.Sprintf("installed %s %s from release %s (verified: archive digest, GitHub attestation, source commit %s)", host, targetVersion, target.resolved.Release.Tag, commit)
		if migrated != "" {
			report.Message += "; " + migrated
		}
	}
	report.Reload = host.ReloadAdvice()
	return report, true, nil
}

func (manager *Manager) resolveTarget(ctx context.Context, channel pluginrelease.Channel, exact string) (resolvedTarget, error) {
	if manager.options.Releases == nil {
		return resolvedTarget{}, &Blocked{Reason: "network access is disabled (L7_NETWORK=off), so no release can be resolved", Next: "unset L7_NETWORK or run offline commands only (status, rollback, uninstall)"}
	}
	stage, cleanup, err := manager.stage()
	if err != nil {
		return resolvedTarget{}, err
	}
	defer cleanup()
	var releases []pluginrelease.Release
	if exact != "" {
		release, err := manager.options.Releases.ReleaseByTag(ctx, "v"+exact)
		if err != nil {
			return resolvedTarget{}, fmt.Errorf("release v%s: %w", exact, err)
		}
		releases = []pluginrelease.Release{release}
	} else if releases, err = manager.options.Releases.Candidates(ctx); err != nil {
		return resolvedTarget{}, err
	}
	release, err := pluginrelease.Select(releases, pluginrelease.Selection{Channel: channel, Exact: exact})
	if err != nil {
		return resolvedTarget{}, &Blocked{Reason: err.Error(), Next: "choose an available release or channel"}
	}
	resolved, err := manager.options.Releases.Resolve(ctx, release, stage)
	if err != nil {
		return resolvedTarget{}, err
	}
	return resolvedTarget{resolved: resolved, channel: channel, pinned: exact != ""}, nil
}

func requireConsent(qualification pluginrelease.Qualification, request Request) error {
	var missing []string
	if qualification.Unsigned() && !request.AcceptUnsigned {
		missing = append(missing, "--accept-unsigned")
	}
	if qualification.Unqualified() && !request.AcceptUnqualified {
		missing = append(missing, "--accept-unqualified")
	}
	if len(missing) == 0 {
		return nil
	}
	return &Blocked{
		Reason: "this release requires explicit acceptance: " + strings.Join(qualification.Warnings(), "; ") + " (policy " + qualification.Policy + "); --yes does not waive trust checks",
		Next:   "rerun with " + strings.Join(missing, " ") + " after reading the release notes",
	}
}

// Rollback re-registers the retained previous version without network use.
func (manager *Manager) Rollback(ctx context.Context, request Request) (Report, error) {
	host, err := parseHost(request.Host)
	if err != nil || request.All {
		return Report{}, &Blocked{Reason: "rollback requires exactly one --host codex|claude", Next: "run l7 plugin rollback --host codex|claude"}
	}
	unlock, err := manager.lock()
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = unlock() }()
	state, _, err := loadState(manager.layout)
	if err != nil {
		return Report{}, err
	}
	if err := manager.recover(ctx, state); err != nil {
		return Report{}, err
	}
	record, managed := state.Hosts[string(host)]
	if !managed {
		return Report{}, &Blocked{Reason: fmt.Sprintf("%s is not managed by this installer", host), Next: "run l7 plugin install --host " + string(host)}
	}
	if record.Previous == nil {
		return Report{}, &Blocked{Reason: fmt.Sprintf("%s has no retained previous version to roll back to", host), Next: "install or update first; one previous verified version is retained after each update"}
	}
	receipt, err := pluginpackage.LoadReceipt(record.Previous.Receipt)
	if err != nil {
		return Report{}, fmt.Errorf("retained %s is unusable: %w", record.Previous.Version, err)
	}
	if err := pluginpackage.VerifyTree(record.Previous.Root, receipt.Files); err != nil {
		return Report{}, fmt.Errorf("retained %s failed verification; refusing to activate it: %w", record.Previous.Version, err)
	}
	if err := manager.options.VerifyExecutable(ctx, filepath.Join(record.Previous.Root, "bin", "l7"), record.Previous.Version); err != nil {
		return Report{}, fmt.Errorf("retained %s executable did not run offline: %w", record.Previous.Version, err)
	}
	port, executable, err := manager.hostPort(host)
	if err != nil {
		return Report{}, err
	}
	hostVersion, err := port.Version(ctx, executable)
	if err != nil {
		return Report{}, &Blocked{Reason: err.Error(), Next: "update or repair the host, then retry"}
	}
	hostReport := HostReport{Host: string(host), From: record.Version, To: record.Previous.Version, Root: record.Previous.Root, Pinned: true, Channel: record.Channel, Warnings: receipt.Warnings}
	if request.DryRun {
		hostReport.Outcome, hostReport.Message = "planned", fmt.Sprintf("dry run: would roll %s back from %s to retained %s without network access", host, record.Version, record.Previous.Version)
		return Report{Outcome: OutcomePass, State: "planned", Message: hostReport.Message, Next: "rerun without --dry-run", Hosts: []HostReport{hostReport}, CLI: manager.options.CLIVersion}, nil
	}
	observation, err := port.Observe(ctx, executable, record.Marketplace, record.Plugin)
	if err != nil {
		return Report{}, err
	}
	from := &Registration{Version: record.Version, Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin}
	to := &Registration{Version: record.Previous.Version, Root: record.Previous.Root, Marketplace: receipt.Marketplace, Plugin: receipt.Plugin}
	if err := manager.switchRegistration(ctx, port, executable, "rollback", from, to, observation); err != nil {
		return Report{}, err
	}
	current := record.retained()
	state.Hosts[string(host)] = HostRecord{
		Version: record.Previous.Version, Channel: record.Channel, Pinned: true, Root: record.Previous.Root, Receipt: record.Previous.Receipt,
		Marketplace: receipt.Marketplace, Plugin: receipt.Plugin, Scope: pluginhost.Scope, HostVersion: hostVersion,
		InstalledAtUTC: manager.options.Now().UTC().Format(time.RFC3339), AcceptedUnsigned: record.AcceptedUnsigned, AcceptedUnqualified: record.AcceptedUnqualified,
		Previous: &current,
	}
	if err := saveState(manager.layout, state); err != nil {
		return Report{}, fmt.Errorf("rollback succeeded but installer state could not be saved: %w", err)
	}
	hostReport.Outcome = "rolled-back"
	hostReport.Message = fmt.Sprintf("rolled %s back from %s to retained %s and pinned it there; %s stays retained to roll forward", host, record.Version, record.Previous.Version, record.Version)
	hostReport.Reload = host.ReloadAdvice()
	report := Report{Outcome: OutcomePass, State: "rolled-back", Message: hostReport.Message, Next: "run l7 plugin update --host " + string(host) + " --unpin to follow the channel again", Hosts: []HostReport{hostReport}, CLI: manager.options.CLIVersion}
	if pointed, err := manager.pointCLI(state); err != nil {
		report.Notes = append(report.Notes, "CLI link was not updated: "+err.Error())
	} else if pointed != "" {
		report.Notes = append(report.Notes, "l7 CLI link now resolves to version "+pointed)
	}
	return report, nil
}

// Uninstall removes one host's registration and the versions no host still
// references. The CLI and installer state survive for other hosts.
func (manager *Manager) Uninstall(ctx context.Context, request Request) (Report, error) {
	host, err := parseHost(request.Host)
	if err != nil || request.All {
		return Report{}, &Blocked{Reason: "uninstall requires exactly one --host codex|claude", Next: "run l7 plugin uninstall --host codex|claude"}
	}
	unlock, err := manager.lock()
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = unlock() }()
	state, _, err := loadState(manager.layout)
	if err != nil {
		return Report{}, err
	}
	if err := manager.recover(ctx, state); err != nil {
		return Report{}, err
	}
	record, managed := state.Hosts[string(host)]
	if !managed {
		return Report{}, &Blocked{Reason: fmt.Sprintf("%s is not managed by this installer; a manually registered Level 7 marketplace is left alone", host), Next: "remove a manual registration with the host's own plugin and marketplace commands"}
	}
	if !request.Yes {
		if manager.options.Confirm == nil {
			return Report{}, &Blocked{Reason: "uninstall needs confirmation and no terminal is available", Next: "rerun with --yes to confirm removal of the " + string(host) + " plugin registration"}
		}
		confirmed, err := manager.options.Confirm(fmt.Sprintf("Remove Level 7 %s from %s and delete its unreferenced installation files? [y/N] ", record.Version, host))
		if err != nil || !confirmed {
			return Report{}, &Blocked{Reason: "uninstall was not confirmed", Next: "rerun and answer y, or pass --yes"}
		}
	}
	hostReport := HostReport{Host: string(host), From: record.Version, Root: record.Root}
	if request.DryRun {
		hostReport.Outcome, hostReport.Message = "planned", fmt.Sprintf("dry run: would unregister %s %s and remove versions no host references", host, record.Version)
		return Report{Outcome: OutcomePass, State: "planned", Message: hostReport.Message, Next: "rerun without --dry-run", Hosts: []HostReport{hostReport}, CLI: manager.options.CLIVersion}, nil
	}
	port, executable, err := manager.hostPort(host)
	if err != nil {
		return Report{}, err
	}
	observation, err := port.Observe(ctx, executable, record.Marketplace, record.Plugin)
	if err != nil {
		return Report{}, err
	}
	from := &Registration{Version: record.Version, Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin}
	if err := manager.switchRegistration(ctx, port, executable, "uninstall", from, nil, observation); err != nil {
		return Report{}, err
	}
	delete(state.Hosts, string(host))
	if err := saveState(manager.layout, state); err != nil {
		return Report{}, fmt.Errorf("uninstall succeeded but installer state could not be saved: %w", err)
	}
	hostReport.Outcome = "uninstalled"
	hostReport.Message = fmt.Sprintf("removed the Level 7 plugin and marketplace registration from %s", host)
	hostReport.Reload = host.ReloadAdvice()
	report := Report{Outcome: OutcomePass, State: "uninstalled", Message: hostReport.Message, Hosts: []HostReport{hostReport}, CLI: manager.options.CLIVersion}
	pointed := ""
	if len(state.Hosts) > 0 {
		if pointed, err = manager.pointCLI(state); err != nil {
			report.Notes = append(report.Notes, "CLI link was not updated: "+err.Error())
		}
		report.Next = "run l7 plugin status"
	} else {
		if link, linkManaged, _ := manager.cliLink(); linkManaged {
			report.Notes = append(report.Notes, "no host remains managed; the l7 CLI link "+link+" and "+manager.layout.Home+" were kept so you can reinstall; remove them yourself to finish")
			if current, err := manager.currentCLIVersion(link); err == nil {
				pointed = current
			}
		}
		report.Next = "run l7 plugin install --host codex|claude to reinstall, or delete " + manager.layout.Home + " to remove everything"
	}
	report.Notes = append(report.Notes, manager.prune(state, pointed)...)
	return report, nil
}

func (manager *Manager) currentCLIVersion(link string) (string, error) {
	target, err := readLink(link)
	if err != nil {
		return "", err
	}
	version := filepath.Base(filepath.Dir(target))
	if _, err := pluginrelease.ParseVersion(version); err != nil {
		return "", err
	}
	return version, nil
}
