// Package plugininstall is the shared installer, updater, and state engine
// behind `l7 plugin`. Codex and Claude differ only through pluginhost
// adapters; release trust lives in pluginrelease and package validation in
// pluginpackage.
package plugininstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

const (
	stateSchema   = 1
	maxStateBytes = 1 << 20
)

// Layout resolves every installer-owned path under one user-owned home.
type Layout struct{ Home string }

func (layout Layout) Plugin() string             { return filepath.Join(layout.Home, "plugin") }
func (layout Layout) StateFile() string          { return filepath.Join(layout.Plugin(), "state.json") }
func (layout Layout) LockFile() string           { return filepath.Join(layout.Plugin(), "lock") }
func (layout Layout) JournalFile() string        { return filepath.Join(layout.Plugin(), "journal.json") }
func (layout Layout) Versions() string           { return filepath.Join(layout.Plugin(), "versions") }
func (layout Layout) Cache() string              { return filepath.Join(layout.Plugin(), "cache") }
func (layout Layout) Staging() string            { return filepath.Join(layout.Plugin(), "staging") }
func (layout Layout) VersionDir(v string) string { return filepath.Join(layout.Versions(), v) }
func (layout Layout) Root(v, host string) string { return filepath.Join(layout.VersionDir(v), host) }
func (layout Layout) Receipt(v, host string) string {
	return filepath.Join(layout.VersionDir(v), host+".receipt.json")
}
func (layout Layout) CLI() string                { return filepath.Join(layout.Home, "cli") }
func (layout Layout) CLIVersion(v string) string { return filepath.Join(layout.CLI(), v) }
func (layout Layout) CLIBinary(v string) string  { return filepath.Join(layout.CLIVersion(v), "l7") }
func (layout Layout) CLIReceipt(v string) string {
	return filepath.Join(layout.CLIVersion(v), "receipt.json")
}
func (layout Layout) CLILinkRecord() string   { return filepath.Join(layout.CLI(), "link") }
func (layout Layout) within(name string) bool { return withinDir(name, layout.Home) }
func (layout Layout) ownsVersionDir(v string) bool {
	return withinDir(layout.VersionDir(v), layout.Versions())
}

func withinDir(name, root string) bool {
	clean := filepath.Clean(name)
	return clean != filepath.Clean(root) && len(clean) > len(root) && clean[:len(root)] == filepath.Clean(root) && os.IsPathSeparator(clean[len(root)])
}

// Retained is a previously active, still verified installation kept for an
// offline rollback.
type Retained struct {
	Version string `json:"version"`
	Root    string `json:"root"`
	Receipt string `json:"receipt"`
}

// HostRecord is the installer's record of one managed host.
type HostRecord struct {
	Version             string    `json:"version"`
	Channel             string    `json:"channel"`
	Pinned              bool      `json:"pinned"`
	Root                string    `json:"root"`
	Receipt             string    `json:"receipt"`
	Marketplace         string    `json:"marketplace"`
	Plugin              string    `json:"plugin"`
	Scope               string    `json:"scope"`
	HostVersion         string    `json:"host_version"`
	InstalledAtUTC      string    `json:"installed_at_utc"`
	AcceptedUnsigned    bool      `json:"accepted_unsigned"`
	AcceptedUnqualified bool      `json:"accepted_unqualified"`
	Previous            *Retained `json:"previous,omitempty"`
}

func (record HostRecord) retained() Retained {
	return Retained{Version: record.Version, Root: record.Root, Receipt: record.Receipt}
}

// State is the installer's digest-bound record. Loading anything that does
// not round-trip exactly fails closed.
type State struct {
	Schema      int                   `json:"schema"`
	Repository  string                `json:"repository"`
	Hosts       map[string]HostRecord `json:"hosts"`
	StateSHA256 string                `json:"state_sha256"`
}

func emptyState() State {
	return State{Schema: stateSchema, Repository: pluginrelease.Repository, Hosts: map[string]HostRecord{}}
}

func (state State) hostNames() []string {
	names := make([]string, 0, len(state.Hosts))
	for name := range state.Hosts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// references reports whether any host's current or retained installation
// still uses a version (the CLI binary for that version is shared).
func (state State) references(version string) bool {
	for _, record := range state.Hosts {
		if record.Version == version || (record.Previous != nil && record.Previous.Version == version) {
			return true
		}
	}
	return false
}

// referencesRoot reports whether any record still points at one extraction.
func (state State) referencesRoot(root string) bool {
	for _, record := range state.Hosts {
		if record.Root == root || (record.Previous != nil && record.Previous.Root == root) {
			return true
		}
	}
	return false
}

func loadState(layout Layout) (State, bool, error) {
	data, err := localfile.Read(layout.StateFile(), maxStateBytes)
	if errors.Is(err, os.ErrNotExist) {
		return emptyState(), false, nil
	}
	if err != nil {
		return State{}, false, fmt.Errorf("installer state %s is unreadable: %w", layout.StateFile(), err)
	}
	var state State
	if err := localfile.DecodeJSON(data, &state); err != nil {
		return State{}, false, fmt.Errorf("installer state %s is corrupt (%v); move it away to start fresh, nothing was changed", layout.StateFile(), err)
	}
	want := state.StateSHA256
	state.StateSHA256 = ""
	digest, err := stateDigest(state)
	if err != nil || want != digest || state.Schema != stateSchema || state.Repository != pluginrelease.Repository {
		return State{}, false, fmt.Errorf("installer state %s failed its integrity check; move it away to start fresh, nothing was changed", layout.StateFile())
	}
	if state.Hosts == nil {
		state.Hosts = map[string]HostRecord{}
	}
	for name, record := range state.Hosts {
		if _, hostErr := parseHost(name); hostErr != nil || record.Version == "" || !layout.within(record.Root) || !layout.within(record.Receipt) || (record.Previous != nil && (!layout.within(record.Previous.Root) || record.Previous.Version == "")) {
			return State{}, false, fmt.Errorf("installer state %s records an unexpected host or path; move it away to start fresh", layout.StateFile())
		}
	}
	state.StateSHA256 = want
	return state, true, nil
}

func saveState(layout Layout, state State) error {
	state.StateSHA256 = ""
	digest, err := stateDigest(state)
	if err != nil {
		return err
	}
	state.StateSHA256 = digest
	data, err := localfile.EncodeJSON(state)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(layout.Plugin(), 0o700); err != nil {
		return err
	}
	if _, err := os.Lstat(layout.StateFile()); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(layout.StateFile(), data, 0o600)
	}
	return localfile.AtomicReplace(layout.StateFile(), data, 0o600)
}

func stateDigest(state State) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// Registration is one host-side registration the journal can restore.
type Registration struct {
	Version     string `json:"version"`
	Root        string `json:"root"`
	Marketplace string `json:"marketplace"`
	Plugin      string `json:"plugin"`
}

type Step struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Journal records one in-flight host registration change so an interrupted
// run can be recovered before any further mutation.
type Journal struct {
	Schema       int           `json:"schema"`
	Operation    string        `json:"operation"`
	Host         string        `json:"host"`
	StartedAtUTC string        `json:"started_at_utc"`
	From         *Registration `json:"from,omitempty"`
	To           *Registration `json:"to,omitempty"`
	Steps        []Step        `json:"steps"`
	Attention    string        `json:"attention,omitempty"`
	StateSHA256  string        `json:"state_sha256"`
}

func loadJournal(layout Layout) (Journal, bool, error) {
	data, err := localfile.Read(layout.JournalFile(), maxStateBytes)
	if errors.Is(err, os.ErrNotExist) {
		return Journal{}, false, nil
	}
	if err != nil {
		return Journal{}, false, fmt.Errorf("installer journal is unreadable: %w", err)
	}
	var journal Journal
	if err := localfile.DecodeJSON(data, &journal); err != nil {
		return Journal{}, false, fmt.Errorf("installer journal %s is corrupt (%v); nothing was changed", layout.JournalFile(), err)
	}
	want := journal.StateSHA256
	journal.StateSHA256 = ""
	digest, err := journalDigest(journal)
	if err != nil || digest != want || journal.Schema != stateSchema {
		return Journal{}, false, fmt.Errorf("installer journal %s failed its integrity check; nothing was changed", layout.JournalFile())
	}
	if _, err := parseHost(journal.Host); err != nil {
		return Journal{}, false, errors.New("installer journal names an unknown host")
	}
	journal.StateSHA256 = want
	return journal, true, nil
}

func writeJournal(layout Layout, journal Journal, create bool) error {
	journal.StateSHA256 = ""
	if journal.Steps == nil {
		journal.Steps = []Step{}
	}
	digest, err := journalDigest(journal)
	if err != nil {
		return err
	}
	journal.StateSHA256 = digest
	data, err := localfile.EncodeJSON(journal)
	if err != nil {
		return err
	}
	if create {
		return localfile.AtomicCreate(layout.JournalFile(), data, 0o600)
	}
	return localfile.AtomicReplace(layout.JournalFile(), data, 0o600)
}

func clearJournal(layout Layout) error {
	err := os.Remove(layout.JournalFile())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot clear installer journal: %w", err)
	}
	return nil
}

func journalDigest(journal Journal) (string, error) {
	data, err := json.Marshal(journal)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// CLIReceipt records one installed CLI binary.
type CLIReceipt struct {
	Schema       int    `json:"schema"`
	Version      string `json:"version"`
	Arch         string `json:"arch"`
	SHA256       string `json:"sha256"`
	SourceCommit string `json:"source_commit"`
}
