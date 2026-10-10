package plugininstall

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginhost"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginpackage"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

// Platform is the running operating system and architecture.
type Platform struct{ OS, Arch string }

func (platform Platform) supported() error {
	if platform.OS != "darwin" || (platform.Arch != "arm64" && platform.Arch != "amd64") {
		return &Blocked{Reason: fmt.Sprintf("Level 7 v1 packages target macOS arm64 and amd64 only; this platform is %s/%s", platform.OS, platform.Arch), Next: "install on a supported macOS machine"}
	}
	return nil
}

// Releases is the network port; pluginrelease.Client implements it.
type Releases interface {
	Candidates(ctx context.Context) ([]pluginrelease.Release, error)
	ReleaseByTag(ctx context.Context, tag string) (pluginrelease.Release, error)
	Resolve(ctx context.Context, release pluginrelease.Release, stage string) (pluginrelease.Resolved, error)
	FetchArchive(ctx context.Context, resolved pluginrelease.Resolved, host, stage, cache string) (string, pluginrelease.Evidence, error)
}

// HostPort is one host adapter; pluginhost.Adapter implements it.
type HostPort interface {
	Host() pluginhost.Host
	Executable() (processadapter.Executable, error)
	Version(ctx context.Context, executable processadapter.Executable) (string, error)
	MarketplaceAdd(ctx context.Context, executable processadapter.Executable, root string) error
	MarketplaceRemove(ctx context.Context, executable processadapter.Executable, marketplace string) error
	PluginAdd(ctx context.Context, executable processadapter.Executable, plugin, marketplace string) error
	PluginRemove(ctx context.Context, executable processadapter.Executable, plugin, marketplace string) error
	Observe(ctx context.Context, executable processadapter.Executable, marketplace, plugin string) (pluginhost.Observation, error)
}

// Options compose a Manager. Tests inject every port; production wiring
// lives in cmd/l7.
type Options struct {
	Home       string
	Platform   Platform
	CLIVersion string
	Releases   Releases
	Hosts      func(pluginhost.Host) HostPort
	// VerifyExecutable runs an extracted launcher offline and confirms it
	// reports the expected version before any host registration.
	VerifyExecutable func(ctx context.Context, launcher, version string) error
	// Confirm answers a non-trust confirmation prompt; nil means no terminal.
	Confirm func(prompt string) (bool, error)
	Now     func() time.Time
}

// Blocked is a refusal the user can resolve: missing consent, prerequisites,
// pins, downgrades, or unsupported platforms. Nothing was mutated.
type Blocked struct {
	Reason string
	Next   string
}

func (blocked *Blocked) Error() string { return blocked.Reason }

type Manager struct {
	layout  Layout
	options Options
}

func New(options Options) (*Manager, error) {
	if !filepath.IsAbs(options.Home) {
		return nil, errors.New("installer home must be an absolute path")
	}
	if err := os.MkdirAll(options.Home, 0o755); err != nil {
		return nil, fmt.Errorf("installer home %s is unusable: %w", options.Home, err)
	}
	physical, err := filepath.EvalSymlinks(options.Home)
	if err == nil {
		err = localfile.ValidateDirectory(physical)
	}
	if err != nil {
		return nil, fmt.Errorf("installer home %s cannot be resolved: %w", options.Home, err)
	}
	options.Home = physical
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Hosts == nil {
		options.Hosts = func(host pluginhost.Host) HostPort { return pluginhost.New(host, nil, nil) }
	}
	if options.VerifyExecutable == nil {
		options.VerifyExecutable = verifyLauncher
	}
	return &Manager{layout: Layout{Home: filepath.Clean(options.Home)}, options: options}, nil
}

func (manager *Manager) Layout() Layout { return manager.layout }

func parseHost(value string) (pluginhost.Host, error) { return pluginhost.Parse(value) }

// lock serializes every mutation; a second concurrent mutation fails
// immediately instead of waiting on a half-written transaction.
func (manager *Manager) lock() (func() error, error) {
	if err := localfile.EnsureDirectory(manager.layout.Plugin(), 0o700); err != nil {
		return nil, err
	}
	lock, err := localfile.AcquireLock(manager.layout.LockFile())
	if err != nil {
		return nil, &Blocked{Reason: "another l7 plugin operation holds the installer lock", Next: "wait for it to finish, then retry"}
	}
	return lock.Close, nil
}

// stage creates one disposable directory for this run's downloads.
func (manager *Manager) stage() (string, func(), error) {
	if err := localfile.EnsureDirectory(manager.layout.Staging(), 0o700); err != nil {
		return "", nil, err
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", nil, err
	}
	path := filepath.Join(manager.layout.Staging(), hex.EncodeToString(random[:]))
	if err := os.Mkdir(path, 0o700); err != nil {
		return "", nil, fmt.Errorf("cannot create staging directory: %w", err)
	}
	return path, func() { _ = os.RemoveAll(path) }, nil
}

// clearStaleStaging removes leftovers of interrupted runs. The staging root
// is created and written only by this installer.
func (manager *Manager) clearStaleStaging() {
	entries, err := os.ReadDir(manager.layout.Staging())
	if err != nil {
		return
	}
	for _, entry := range entries {
		_ = os.RemoveAll(filepath.Join(manager.layout.Staging(), entry.Name()))
	}
}

func (manager *Manager) hostPort(host pluginhost.Host) (HostPort, processadapter.Executable, error) {
	port := manager.options.Hosts(host)
	if port == nil {
		return nil, processadapter.Executable{}, fmt.Errorf("no adapter for host %s", host)
	}
	executable, err := port.Executable()
	if err != nil {
		return nil, processadapter.Executable{}, &Blocked{Reason: err.Error(), Next: "install the host or choose the other host"}
	}
	return port, executable, nil
}

// recover finishes or reverses an interrupted registration change. It runs
// under the lock before every mutation and refuses to proceed on failure.
func (manager *Manager) recover(ctx context.Context, state State) error {
	journal, found, err := loadJournal(manager.layout)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	host, _ := parseHost(journal.Host)
	port, executable, err := manager.hostPort(host)
	if err != nil {
		return fmt.Errorf("an interrupted %s %s transaction needs recovery but %w", journal.Host, journal.Operation, err)
	}
	record, managed := state.Hosts[journal.Host]
	target := journal.From
	if journal.To != nil && managed && record.Version == journal.To.Version && record.Root == journal.To.Root {
		target = journal.To
	}
	if err := manager.settle(ctx, port, executable, journal, target); err != nil {
		journal.Attention = err.Error()
		_ = writeJournal(manager.layout, journal, false)
		return fmt.Errorf("recovery of the interrupted %s %s transaction failed: %w; the journal is retained at %s and the next install or update retries", journal.Host, journal.Operation, err, manager.layout.JournalFile())
	}
	return clearJournal(manager.layout)
}

// settle makes the host registration equal target (nil means unregistered)
// and proves it through the host's own listing.
func (manager *Manager) settle(ctx context.Context, port HostPort, executable processadapter.Executable, journal Journal, target *Registration) error {
	marketplace, plugin := registrationNames(journal)
	observation, err := port.Observe(ctx, executable, marketplace, plugin)
	if err != nil {
		return err
	}
	if registrationMatches(observation, target) {
		return nil
	}
	if observation.PluginInstalled {
		if err := port.PluginRemove(ctx, executable, plugin, marketplace); err != nil {
			return err
		}
	}
	if observation.MarketplaceRegistered {
		if err := port.MarketplaceRemove(ctx, executable, marketplace); err != nil {
			return err
		}
	}
	if target != nil {
		if err := port.MarketplaceAdd(ctx, executable, target.Root); err != nil {
			return err
		}
		if err := port.PluginAdd(ctx, executable, target.Plugin, target.Marketplace); err != nil {
			return err
		}
	}
	observation, err = port.Observe(ctx, executable, marketplace, plugin)
	if err != nil {
		return err
	}
	if !registrationMatches(observation, target) {
		return errors.New("host listing does not show the expected registration after restoration")
	}
	return nil
}

func registrationNames(journal Journal) (string, string) {
	for _, registration := range []*Registration{journal.To, journal.From} {
		if registration != nil {
			return registration.Marketplace, registration.Plugin
		}
	}
	return pluginpackage.MarketplaceName, pluginpackage.PluginName
}

func registrationMatches(observation pluginhost.Observation, target *Registration) bool {
	if target == nil {
		return !observation.MarketplaceRegistered && !observation.PluginInstalled
	}
	return observation.MarketplaceRegistered && observation.PluginInstalled && samePath(observation.MarketplaceRoot, target.Root) && observation.PluginVersion == target.Version
}

func samePath(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	if filepath.Clean(left) == filepath.Clean(right) {
		return true
	}
	leftPhysical, leftErr := filepath.EvalSymlinks(left)
	rightPhysical, rightErr := filepath.EvalSymlinks(right)
	return leftErr == nil && rightErr == nil && leftPhysical == rightPhysical
}

// switchRegistration moves the host from one registration to another under a
// journal. Removal steps run only for registrations the host actually lists,
// and any failure restores the previous registration and reports it.
func (manager *Manager) switchRegistration(ctx context.Context, port HostPort, executable processadapter.Executable, operation string, from, to *Registration, observation pluginhost.Observation) error {
	if _, found, err := loadJournal(manager.layout); err != nil || found {
		return errors.New("an installer journal already exists; recovery must run first")
	}
	host := port.Host()
	journal := Journal{Schema: stateSchema, Operation: operation, Host: string(host), StartedAtUTC: manager.options.Now().UTC().Format(time.RFC3339), From: from, To: to}
	if from != nil && observation.PluginInstalled {
		journal.Steps = append(journal.Steps, Step{Name: "plugin-remove", Status: "pending"})
	}
	if from != nil && observation.MarketplaceRegistered {
		journal.Steps = append(journal.Steps, Step{Name: "marketplace-remove", Status: "pending"})
	}
	if to != nil {
		journal.Steps = append(journal.Steps, Step{Name: "marketplace-add", Status: "pending"}, Step{Name: "plugin-add", Status: "pending"})
	}
	if err := writeJournal(manager.layout, journal, true); err != nil {
		return err
	}
	marketplace, plugin := registrationNames(journal)
	runErr := manager.runSteps(ctx, port, executable, &journal, from, to)
	if runErr == nil {
		observation, err := port.Observe(ctx, executable, marketplace, plugin)
		if err != nil {
			runErr = err
		} else if !registrationMatches(observation, to) {
			runErr = fmt.Errorf("%s does not list the expected registration after the change", host)
		}
	}
	if runErr == nil {
		return clearJournal(manager.layout)
	}
	if restoreErr := manager.settle(ctx, port, executable, journal, from); restoreErr != nil {
		journal.Attention = restoreErr.Error()
		_ = writeJournal(manager.layout, journal, false)
		return fmt.Errorf("%s %s failed: %w; restoring the previous registration also failed (%v); the journal at %s is retained and the next install or update retries recovery", host, operation, runErr, restoreErr, manager.layout.JournalFile())
	}
	if err := clearJournal(manager.layout); err != nil {
		return err
	}
	return fmt.Errorf("%s %s failed: %w; the previous registration was restored and verified", host, operation, runErr)
}

func (manager *Manager) runSteps(ctx context.Context, port HostPort, executable processadapter.Executable, journal *Journal, from, to *Registration) error {
	for index := range journal.Steps {
		var err error
		switch journal.Steps[index].Name {
		case "plugin-remove":
			err = port.PluginRemove(ctx, executable, from.Plugin, from.Marketplace)
		case "marketplace-remove":
			err = port.MarketplaceRemove(ctx, executable, from.Marketplace)
		case "marketplace-add":
			err = port.MarketplaceAdd(ctx, executable, to.Root)
		case "plugin-add":
			err = port.PluginAdd(ctx, executable, to.Plugin, to.Marketplace)
		}
		if err != nil {
			journal.Steps[index].Status = "failed"
			_ = writeJournal(manager.layout, *journal, false)
			return err
		}
		journal.Steps[index].Status = "done"
		if err := writeJournal(manager.layout, *journal, false); err != nil {
			return err
		}
	}
	return nil
}

// ensureVersion returns a verified extraction root for the host, reusing a
// receipted directory or staging, validating, and placing a fresh one.
func (manager *Manager) ensureVersion(ctx context.Context, resolved pluginrelease.Resolved, host pluginhost.Host, stage string, accepted acceptance) (string, string, pluginpackage.Receipt, error) {
	version := resolved.Contract.Version.String()
	root := manager.layout.Root(version, string(host))
	receiptPath := manager.layout.Receipt(version, string(host))
	archive, err := resolved.Contract.Archive(string(host))
	if err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	receipt, receiptErr := pluginpackage.LoadReceipt(receiptPath)
	_, rootErr := os.Lstat(root)
	rootExists := rootErr == nil || !errors.Is(rootErr, os.ErrNotExist)
	switch {
	case receiptErr == nil && rootExists:
		if receipt.ArchiveSHA256 != archive.SHA256 {
			return "", "", pluginpackage.Receipt{}, fmt.Errorf("%s was installed from a different archive than release %s now publishes; move it away before continuing", root, resolved.Release.Tag)
		}
		if err := pluginpackage.VerifyTree(root, receipt.Files); err != nil {
			return "", "", pluginpackage.Receipt{}, fmt.Errorf("retained installation %s failed verification: %w", root, err)
		}
		return root, receiptPath, receipt, nil
	case rootExists:
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("%s exists but is not an installer-managed extraction (no valid receipt); move it away before continuing", root)
	default:
		// A receipt without its directory is an interrupted placement; the
		// receipt is the installer's own file and is safe to redo.
		if err := os.Remove(receiptPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", "", pluginpackage.Receipt{}, fmt.Errorf("cannot clear orphan receipt %s: %w", receiptPath, err)
		}
	}
	archivePath, evidence, err := manager.options.Releases.FetchArchive(ctx, resolved, string(host), stage, manager.layout.Cache())
	if err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	pkg, err := pluginpackage.Load(archivePath, string(host), version)
	if err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("%s: %w", archive.Name, err)
	}
	if pkg.Digest != archive.SHA256 || pkg.Digest != evidence.SubjectSHA256 {
		return "", "", pluginpackage.Receipt{}, errors.New("archive bytes changed between verification and validation")
	}
	identity, err := pluginpackage.Validate(pkg, manager.options.Platform.Arch, resolved.Contract.Inputs)
	if err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("%s: %w", archive.Name, err)
	}
	stagedRoot := filepath.Join(stage, string(host))
	if err := pluginpackage.Extract(stagedRoot, pkg); err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	receipt = pluginpackage.Receipt{
		Schema: 1, Host: string(host), Version: version, Tag: resolved.Release.Tag, Channel: string(resolved.Contract.Channel),
		ArchiveName: archive.Name, ArchiveSHA256: pkg.Digest, SourceCommit: resolved.Contract.SourceCommit, Workflow: resolved.Contract.Workflow.Path,
		Policy: resolved.Contract.Qualification.Policy, Unsigned: resolved.Contract.Qualification.Unsigned(), Unqualified: resolved.Contract.Qualification.Unqualified(),
		Warnings: resolved.Contract.Qualification.Warnings(), VerifiedAtUTC: manager.options.Now().UTC().Format(time.RFC3339),
		Marketplace: identity.Marketplace, Plugin: identity.Plugin, Files: pluginpackage.Owned(pkg),
	}
	if err := pluginpackage.VerifyTree(stagedRoot, receipt.Files); err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("staged extraction failed verification: %w", err)
	}
	if err := manager.options.VerifyExecutable(ctx, filepath.Join(stagedRoot, "bin", "l7"), version); err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("the extracted %s executable did not run offline: %w (if macOS Gatekeeper blocked it, stop; never disable Gatekeeper globally)", host, err)
	}
	if err := manager.ensureCLIBinary(pkg, resolved.Contract, accepted); err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	if err := localfile.EnsureDirectory(manager.layout.VersionDir(version), 0o755); err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	stagedReceipt := filepath.Join(stage, string(host)+".receipt.json")
	if err := pluginpackage.WriteReceipt(stagedReceipt, receipt); err != nil {
		return "", "", pluginpackage.Receipt{}, err
	}
	if err := os.Link(stagedReceipt, receiptPath); err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("cannot place ownership receipt: %w", err)
	}
	if err := os.Rename(stagedRoot, root); err != nil {
		_ = os.Remove(receiptPath)
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("cannot place verified extraction: %w", err)
	}
	if err := pluginpackage.VerifyTree(root, receipt.Files); err != nil {
		return "", "", pluginpackage.Receipt{}, fmt.Errorf("placed extraction failed verification: %w", err)
	}
	return root, receiptPath, receipt, nil
}

type acceptance struct{ unsigned, unqualified bool }

// prune removes extraction roots and CLI binaries no host references, after
// proving each still matches its receipt.
func (manager *Manager) prune(state State, keepCLI string) []string {
	var notes []string
	entries, err := os.ReadDir(manager.layout.Versions())
	if err == nil {
		for _, entry := range entries {
			version := entry.Name()
			if !entry.IsDir() || !manager.layout.ownsVersionDir(version) {
				continue
			}
			if err := manager.pruneVersionDir(state, version); err != nil {
				notes = append(notes, fmt.Sprintf("retained %s: %v", manager.layout.VersionDir(version), err))
			}
		}
	}
	entries, err = os.ReadDir(manager.layout.CLI())
	if err == nil {
		for _, entry := range entries {
			version := entry.Name()
			if !entry.IsDir() || state.references(version) || version == keepCLI || version == manager.options.CLIVersion {
				continue
			}
			if _, err := pluginrelease.ParseVersion(version); err != nil {
				continue
			}
			if err := manager.removeCLIVersion(version); err != nil {
				notes = append(notes, fmt.Sprintf("retained %s: %v", manager.layout.CLIVersion(version), err))
			}
		}
	}
	return notes
}

// pruneVersionDir removes each host root under one version that no record
// references, then the version directory once nothing remains in it.
func (manager *Manager) pruneVersionDir(state State, version string) error {
	directory := manager.layout.VersionDir(version)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	remaining := 0
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".receipt.json") {
			continue
		}
		root := filepath.Join(directory, name)
		if !entry.IsDir() {
			return fmt.Errorf("%s contains unexpected entry %q", directory, name)
		}
		if state.referencesRoot(root) {
			remaining++
			continue
		}
		receipt, err := pluginpackage.LoadReceipt(manager.layout.Receipt(version, name))
		if err != nil {
			return fmt.Errorf("%s has no valid receipt", root)
		}
		if err := pluginpackage.RemoveTree(root, receipt.Files); err != nil {
			return err
		}
		_ = os.Remove(manager.layout.Receipt(version, name))
	}
	if remaining > 0 {
		return nil
	}
	entries, err = os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".receipt.json") {
			return fmt.Errorf("%s contains unexpected entry %q", directory, entry.Name())
		}
		_ = os.Remove(filepath.Join(directory, entry.Name()))
	}
	return os.Remove(directory)
}

func verifyLauncher(ctx context.Context, launcher, version string) error {
	info, err := os.Lstat(launcher)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return errors.New("launcher is missing or not executable")
	}
	environment := []string{"HOME=/", "PATH=/usr/bin:/bin", "LC_ALL=C", "LANG=C", "L7_NETWORK=off", "L7_TELEMETRY=off", "NO_COLOR=1"}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: launcher, Arguments: []string{"--version", "--json"}, Directory: filepath.Dir(launcher),
		Environment: environment, MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("launcher exited %d: %s", result.ExitCode, strings.TrimSpace(string(result.Stderr)))
	}
	var envelope struct {
		Outcome string `json:"outcome"`
		Command string `json:"command"`
		Version string `json:"version"`
	}
	if json.Unmarshal(result.Stdout, &envelope) != nil || envelope.Outcome != "PASS" || envelope.Command != "version" || envelope.Version != version {
		return fmt.Errorf("launcher reported %q instead of version %s", strings.TrimSpace(string(result.Stdout)), version)
	}
	return nil
}
