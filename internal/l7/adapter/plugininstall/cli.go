package plugininstall

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginpackage"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

// cliLink reads the bootstrap's link record. Without it the running binary
// is a developer build and the installer leaves the CLI alone.
func (manager *Manager) cliLink() (string, bool, error) {
	data, err := localfile.Read(manager.layout.CLILinkRecord(), 4096)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("CLI link record is unreadable: %w", err)
	}
	link := strings.TrimSpace(string(data))
	if !filepath.IsAbs(link) || strings.ContainsAny(link, "\n\x00") {
		return "", false, errors.New("CLI link record does not name an absolute path")
	}
	return link, true, nil
}

// ensureCLIBinary stores the package's CLI for the running architecture
// beside the other retained CLI versions. It never touches the running
// binary or the link; pointCLI does that after the host change succeeds.
func (manager *Manager) ensureCLIBinary(pkg pluginpackage.Package, contract pluginrelease.Contract, _ acceptance) error {
	if _, managed, err := manager.cliLink(); err != nil || !managed {
		return err
	}
	version := contract.Version.String()
	if _, err := manager.loadCLIReceipt(version); err == nil {
		return nil
	}
	relative := "darwin-" + manager.options.Platform.Arch + "/l7"
	var data []byte
	for _, file := range pkg.Files {
		if file.Path == "bin/"+relative {
			data = file.Data
		}
	}
	digest := sha256.Sum256(data)
	if len(data) == 0 || hex.EncodeToString(digest[:]) != contract.Inputs[relative] {
		return errors.New("package CLI binary does not match the attested release input")
	}
	if err := localfile.EnsureDirectory(manager.layout.CLI(), 0o755); err != nil {
		return err
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	staged := filepath.Join(manager.layout.CLI(), ".staging-"+hex.EncodeToString(random[:]))
	if err := os.Mkdir(staged, 0o755); err != nil {
		return err
	}
	cleanup := func() { _ = os.RemoveAll(staged) }
	binary := filepath.Join(staged, "l7")
	if err := localfile.AtomicCreate(binary, data, 0o755); err != nil {
		cleanup()
		return err
	}
	receipt := CLIReceipt{Schema: 1, Version: version, Arch: manager.options.Platform.Arch, SHA256: hex.EncodeToString(digest[:]), SourceCommit: contract.SourceCommit}
	encoded, err := localfile.EncodeJSON(receipt)
	if err != nil {
		cleanup()
		return err
	}
	if err := localfile.AtomicCreate(filepath.Join(staged, "receipt.json"), encoded, 0o644); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(staged, manager.layout.CLIVersion(version)); err != nil {
		cleanup()
		return fmt.Errorf("cannot place CLI %s: %w", version, err)
	}
	return nil
}

func (manager *Manager) loadCLIReceipt(version string) (CLIReceipt, error) {
	data, err := localfile.Read(manager.layout.CLIReceipt(version), 64<<10)
	if err != nil {
		return CLIReceipt{}, err
	}
	var receipt CLIReceipt
	if err := localfile.DecodeJSON(data, &receipt); err != nil || receipt.Schema != 1 || receipt.Version != version || !pluginrelease.ValidDigest(receipt.SHA256) {
		return CLIReceipt{}, errors.New("CLI receipt is invalid")
	}
	digest, _, err := pluginrelease.FileDigest(manager.layout.CLIBinary(version), 256<<20)
	if err != nil || digest != receipt.SHA256 {
		return CLIReceipt{}, errors.New("CLI binary does not match its receipt")
	}
	return receipt, nil
}

// pointCLI repoints the user-facing link at the newest installed version
// without rewriting any binary file; the previous binary stays retained.
func (manager *Manager) pointCLI(state State) (string, error) {
	link, managed, err := manager.cliLink()
	if err != nil || !managed {
		return "", err
	}
	newest := ""
	var newestVersion pluginrelease.Version
	for _, record := range state.Hosts {
		version, err := pluginrelease.ParseVersion(record.Version)
		if err != nil {
			continue
		}
		if newest == "" || version.Compare(newestVersion) > 0 {
			newest, newestVersion = record.Version, version
		}
	}
	if newest == "" {
		return "", nil
	}
	if _, err := manager.loadCLIReceipt(newest); err != nil {
		return "", fmt.Errorf("CLI %s is not retained: %w", newest, err)
	}
	target := manager.layout.CLIBinary(newest)
	if current, err := os.Readlink(link); err == nil && current == target {
		return newest, nil
	}
	if info, err := os.Lstat(link); err == nil {
		existing, readErr := os.Readlink(link)
		if info.Mode()&os.ModeSymlink == 0 || readErr != nil || !withinDir(existing, manager.layout.CLI()) {
			return "", fmt.Errorf("%s is not a Level 7 CLI link; refusing to replace it", link)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := localfile.EnsureDirectory(filepath.Dir(link), 0o755); err != nil {
		return "", err
	}
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	temporary := link + ".l7-" + hex.EncodeToString(random[:])
	if err := os.Symlink(target, temporary); err != nil {
		return "", err
	}
	if err := os.Rename(temporary, link); err != nil {
		_ = os.Remove(temporary)
		return "", err
	}
	return newest, nil
}

func (manager *Manager) removeCLIVersion(version string) error {
	if _, err := manager.loadCLIReceipt(version); err != nil {
		return err
	}
	link, managed, err := manager.cliLink()
	if err != nil {
		return err
	}
	if managed {
		if current, err := os.Readlink(link); err == nil && current == manager.layout.CLIBinary(version) {
			return errors.New("the CLI link still points at this version")
		}
	}
	directory := manager.layout.CLIVersion(version)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != "l7" && entry.Name() != "receipt.json" {
			return fmt.Errorf("%s contains unexpected entry %q", directory, entry.Name())
		}
	}
	return os.RemoveAll(directory)
}
