package plugininstall

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginhost"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginpackage"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

// Status is the offline view of every managed host plus the CLI. Available
// versions are filled only by an explicit --check.
type Status struct {
	Schema       int                `json:"schema"`
	Home         string             `json:"home"`
	ManagedPaths map[string]string  `json:"managed_paths"`
	CLI          CLIStatus          `json:"cli"`
	Hosts        []HostStatus       `json:"hosts"`
	Transaction  *TransactionStatus `json:"incomplete_transaction,omitempty"`
	Checked      bool               `json:"network_checked"`
	Notes        []string           `json:"notes,omitempty"`
}

type CLIStatus struct {
	RunningVersion string   `json:"running_version"`
	Managed        bool     `json:"managed"`
	Link           string   `json:"link,omitempty"`
	LinkTarget     string   `json:"link_target,omitempty"`
	Retained       []string `json:"retained_versions,omitempty"`
}

type HostStatus struct {
	Host             string                  `json:"host"`
	Managed          bool                    `json:"managed"`
	Version          string                  `json:"version,omitempty"`
	Channel          string                  `json:"channel,omitempty"`
	Pinned           bool                    `json:"pinned"`
	Scope            string                  `json:"scope,omitempty"`
	Root             string                  `json:"root,omitempty"`
	Receipt          string                  `json:"receipt,omitempty"`
	Marketplace      string                  `json:"marketplace,omitempty"`
	Plugin           string                  `json:"plugin,omitempty"`
	InstalledAtUTC   string                  `json:"installed_at_utc,omitempty"`
	Verification     *VerificationStatus     `json:"verification,omitempty"`
	Warnings         []string                `json:"warnings,omitempty"`
	HostExecutable   string                  `json:"host_executable,omitempty"`
	HostVersion      string                  `json:"host_version,omitempty"`
	HostObservation  *pluginhost.Observation `json:"host_observation,omitempty"`
	HostDiagnostic   string                  `json:"host_diagnostic,omitempty"`
	Previous         *Retained               `json:"previous,omitempty"`
	AvailableVersion string                  `json:"available_version,omitempty"`
	UpdateAvailable  bool                    `json:"update_available"`
	Availability     string                  `json:"availability"`
}

type VerificationStatus struct {
	Tree          string `json:"tree"`
	ArchiveName   string `json:"archive_name"`
	ArchiveSHA256 string `json:"archive_sha256"`
	SourceCommit  string `json:"source_commit"`
	Workflow      string `json:"workflow"`
	Policy        string `json:"policy"`
	VerifiedAtUTC string `json:"verified_at_utc"`
}

type TransactionStatus struct {
	Operation    string `json:"operation"`
	Host         string `json:"host"`
	StartedAtUTC string `json:"started_at_utc"`
	Attention    string `json:"attention,omitempty"`
}

// Status reports installed and, with --check, available versions. It takes
// no lock and performs no network access unless request.Check is set.
func (manager *Manager) Status(ctx context.Context, request Request) (Report, error) {
	state, _, err := loadState(manager.layout)
	if err != nil {
		return Report{}, err
	}
	status := Status{
		Schema: stateSchema, Home: manager.layout.Home, Hosts: []HostStatus{},
		ManagedPaths: map[string]string{
			"state": manager.layout.StateFile(), "journal": manager.layout.JournalFile(), "versions": manager.layout.Versions(),
			"cache": manager.layout.Cache(), "cli": manager.layout.CLI(),
		},
		CLI: CLIStatus{RunningVersion: manager.options.CLIVersion},
	}
	if link, managed, err := manager.cliLink(); err != nil {
		status.Notes = append(status.Notes, err.Error())
	} else if managed {
		status.CLI.Managed, status.CLI.Link = true, link
		if target, err := readLink(link); err == nil {
			status.CLI.LinkTarget = target
		}
		if entries, err := os.ReadDir(manager.layout.CLI()); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					if _, err := manager.loadCLIReceipt(entry.Name()); err == nil {
						status.CLI.Retained = append(status.CLI.Retained, entry.Name())
					}
				}
			}
		}
	}
	if journal, found, err := loadJournal(manager.layout); err != nil {
		status.Notes = append(status.Notes, err.Error())
	} else if found {
		status.Transaction = &TransactionStatus{Operation: journal.Operation, Host: journal.Host, StartedAtUTC: journal.StartedAtUTC, Attention: journal.Attention}
		status.Notes = append(status.Notes, "an interrupted "+journal.Host+" "+journal.Operation+" is journaled; the next install, update, rollback, or uninstall restores the previous registration before continuing")
	}
	var releases []pluginrelease.Release
	if request.Check {
		if manager.options.Releases == nil {
			return Report{}, &Blocked{Reason: "network access is disabled (L7_NETWORK=off); status --check cannot query releases", Next: "run l7 plugin status without --check"}
		}
		releases, err = manager.options.Releases.Candidates(ctx)
		if err != nil {
			return Report{}, err
		}
		status.Checked = true
	}
	updates := 0
	for _, host := range pluginhost.All() {
		hostStatus := HostStatus{Host: string(host), Availability: "not checked; run l7 plugin status --check"}
		record, managed := state.Hosts[string(host)]
		if managed {
			hostStatus.Managed = true
			hostStatus.Version, hostStatus.Channel, hostStatus.Pinned, hostStatus.Scope = record.Version, record.Channel, record.Pinned, record.Scope
			hostStatus.Root, hostStatus.Receipt, hostStatus.Marketplace, hostStatus.Plugin = record.Root, record.Receipt, record.Marketplace, record.Plugin
			hostStatus.InstalledAtUTC, hostStatus.Previous = record.InstalledAtUTC, record.Previous
			if receipt, err := pluginpackage.LoadReceipt(record.Receipt); err != nil {
				hostStatus.Verification = &VerificationStatus{Tree: "FAILED: " + err.Error()}
			} else {
				verification := &VerificationStatus{Tree: "verified", ArchiveName: receipt.ArchiveName, ArchiveSHA256: receipt.ArchiveSHA256, SourceCommit: receipt.SourceCommit, Workflow: receipt.Workflow, Policy: receipt.Policy, VerifiedAtUTC: receipt.VerifiedAtUTC}
				if err := pluginpackage.VerifyTree(record.Root, receipt.Files); err != nil {
					verification.Tree = "FAILED: " + err.Error()
				}
				hostStatus.Verification = verification
				hostStatus.Warnings = receipt.Warnings
			}
		}
		port := manager.options.Hosts(host)
		if executable, err := port.Executable(); err != nil {
			hostStatus.HostDiagnostic = err.Error()
		} else {
			hostStatus.HostExecutable = executable.Path
			if version, err := port.Version(ctx, executable); err != nil {
				hostStatus.HostDiagnostic = err.Error()
			} else {
				hostStatus.HostVersion = version
			}
			if managed && hostStatus.HostExecutable != "" {
				if observation, err := port.Observe(ctx, executable, record.Marketplace, record.Plugin); err != nil {
					hostStatus.HostDiagnostic = err.Error()
				} else {
					hostStatus.HostObservation = &observation
					if !registrationMatches(observation, &Registration{Version: record.Version, Root: record.Root, Marketplace: record.Marketplace, Plugin: record.Plugin}) {
						hostStatus.HostDiagnostic = "host listing does not match the installer record; run l7 plugin update --host " + string(host) + " --version " + record.Version + " to re-register"
					}
				}
			}
		}
		if request.Check && managed {
			channel := pluginrelease.Channel(record.Channel)
			if !channel.Valid() {
				channel = pluginrelease.ChannelStable
			}
			selection := pluginrelease.Selection{Channel: channel}
			if record.Pinned {
				selection.Exact = record.Version
			}
			if release, err := pluginrelease.Select(releases, selection); err != nil {
				hostStatus.Availability = err.Error()
			} else {
				hostStatus.AvailableVersion = release.Version.String()
				installed, parseErr := pluginrelease.ParseVersion(record.Version)
				switch {
				case record.Pinned:
					hostStatus.Availability = "pinned at " + record.Version + "; newest " + string(channel) + " release is " + newestOf(releases, channel)
				case parseErr == nil && release.Version.Compare(installed) > 0:
					hostStatus.UpdateAvailable = true
					updates++
					hostStatus.Availability = "update available: " + release.Version.String()
				default:
					hostStatus.Availability = "current"
				}
			}
		} else if request.Check {
			hostStatus.Availability = "not installed; newest stable release is " + newestOf(releases, pluginrelease.ChannelStable)
		}
		status.Hosts = append(status.Hosts, hostStatus)
	}
	report := Report{Outcome: OutcomePass, State: "reported", CLI: manager.options.CLIVersion, Status: &status, Next: "run l7 plugin status --check to query available releases"}
	managedCount := len(state.Hosts)
	switch {
	case status.Transaction != nil:
		report.State = "needs-recovery"
		report.Message = "an interrupted transaction is journaled"
		report.Next = "run the install or update again to recover, then l7 plugin status"
	case managedCount == 0:
		report.State = "not-installed"
		report.Message = "no host is managed by this installer"
		report.Next = "run l7 plugin install --host codex|claude"
	case updates > 0:
		report.State = "update-available"
		report.Message = fmt.Sprintf("%d host(s) can be updated", updates)
		report.Next = "run l7 plugin update --all with the acceptance flags the release requires"
	default:
		report.Message = fmt.Sprintf("%d host(s) managed", managedCount)
		if status.Checked {
			report.State, report.Next = "current", "nothing to do"
		}
	}
	for _, hostStatus := range status.Hosts {
		if hostStatus.Verification != nil && strings.HasPrefix(hostStatus.Verification.Tree, "FAILED") {
			report.Outcome, report.State, report.Message = OutcomeFailed, "verification-failed", hostStatus.Host+" installation no longer matches its receipt"
			report.Next = "run l7 plugin rollback --host " + hostStatus.Host + " or reinstall with l7 plugin update --host " + hostStatus.Host + " --version " + hostStatus.Version
		}
	}
	return report, nil
}

func newestOf(releases []pluginrelease.Release, channel pluginrelease.Channel) string {
	release, err := pluginrelease.Select(releases, pluginrelease.Selection{Channel: channel})
	if err != nil {
		return "none"
	}
	return release.Version.String()
}

func readLink(link string) (string, error) {
	info, err := os.Lstat(link)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return "", errors.New("not a symlink")
	}
	target, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(link), target)
	}
	return target, nil
}
