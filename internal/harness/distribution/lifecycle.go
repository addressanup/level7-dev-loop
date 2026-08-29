package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var errFixtureInterrupted = errors.New("injected fixture interruption")

type receiptFile struct {
	Path   string `json:"path"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type receiptPackage struct {
	Version string        `json:"version"`
	Digest  string        `json:"digest"`
	Files   []receiptFile `json:"files"`
}

type lifecycleReceipt struct {
	Schema         int              `json:"schema"`
	Host           string           `json:"host"`
	ActiveDigest   string           `json:"active_digest"`
	PreviousDigest string           `json:"previous_digest"`
	Packages       []receiptPackage `json:"packages"`
}

type pendingInstall struct {
	Schema         int            `json:"schema"`
	Host           string         `json:"host"`
	PreviousDigest string         `json:"previous_digest"`
	Package        receiptPackage `json:"package"`
}

type removalPreview struct {
	Host          string
	OwnedPackages []string
	Conflicts     []string
}

func qualifyLifecycleSet(packages []builtPackage) error {
	for _, built := range packages {
		parent, err := os.MkdirTemp("", "l7-distribution-lifecycle-")
		if err != nil {
			return err
		}
		qualificationErr := qualifyOneLifecycle(parent, built)
		cleanupErr := os.RemoveAll(parent)
		if err := errors.Join(qualificationErr, cleanupErr); err != nil {
			return fmt.Errorf("%s: %w", built.Host, err)
		}
	}
	return nil
}

func qualifyOneLifecycle(parent string, base builtPackage) error {
	hostRoot := filepath.Join(parent, "host")
	projectRoot := filepath.Join(parent, "project")
	if err := writeRegular(projectRoot, "canonical.md", []byte("preserve canonical project state\n")); err != nil {
		return err
	}
	if err := writeRegular(hostRoot, "unowned.txt", []byte("preserve unowned host state\n")); err != nil {
		return err
	}
	if err := installFixture(hostRoot, base, ""); err != nil {
		return fmt.Errorf("clean install: %w", err)
	}
	if err := installFixture(hostRoot, base, ""); err != nil {
		return fmt.Errorf("same-version reinstall: %w", err)
	}
	variant, err := syntheticLifecycleVariant(base)
	if err != nil {
		return err
	}
	if err := installFixture(hostRoot, variant, "after-publish"); !errors.Is(err, errFixtureInterrupted) {
		return fmt.Errorf("post-publish interruption did not stop: %v", err)
	}
	if err := recoverFixture(hostRoot, base.Host); err != nil {
		return fmt.Errorf("recover interrupted upgrade: %w", err)
	}
	receipt, err := loadLifecycleReceipt(hostRoot, base.Host, true)
	if err != nil || receipt.ActiveDigest != variant.ArchiveDigest {
		return fmt.Errorf("recovered active digest=%q error=%v", receipt.ActiveDigest, err)
	}
	if err := rollbackFixture(hostRoot, base.Host); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	receipt, err = loadLifecycleReceipt(hostRoot, base.Host, true)
	if err != nil || receipt.ActiveDigest != base.ArchiveDigest {
		return fmt.Errorf("rollback active digest=%q error=%v", receipt.ActiveDigest, err)
	}
	if err := installFixture(hostRoot, variant, ""); err != nil {
		return fmt.Errorf("upgrade after rollback: %w", err)
	}

	activeDirectory := packageDirectory(hostRoot, base.Host, variant.ArchiveDigest)
	unownedPath := filepath.Join(activeDirectory, "user-note.txt")
	if err := os.WriteFile(unownedPath, []byte("unowned\n"), 0o600); err != nil {
		return err
	}
	preview, err := prepareRemoval(hostRoot, base.Host)
	if err != nil || len(preview.Conflicts) == 0 {
		return fmt.Errorf("unowned conflict preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(hostRoot, base.Host); err == nil {
		return errors.New("removal accepted an unowned package file")
	}
	if _, err := os.Stat(unownedPath); err != nil {
		return errors.New("blocked removal changed the unowned file")
	}
	if err := os.Remove(unownedPath); err != nil {
		return err
	}

	owned := variant.Entries[0]
	ownedPath := filepath.Join(activeDirectory, filepath.FromSlash(owned.Name))
	if err := os.WriteFile(ownedPath, []byte("changed\n"), 0o644); err != nil {
		return err
	}
	preview, err = prepareRemoval(hostRoot, base.Host)
	if err != nil || len(preview.Conflicts) == 0 {
		return fmt.Errorf("changed-owned conflict preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(hostRoot, base.Host); err == nil {
		return errors.New("removal accepted a changed owned file")
	}
	if err := os.WriteFile(ownedPath, owned.Data, 0o644); err != nil {
		return err
	}
	preview, err = prepareRemoval(hostRoot, base.Host)
	if err != nil || len(preview.Conflicts) != 0 || len(preview.OwnedPackages) != 2 {
		return fmt.Errorf("clean removal preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(hostRoot, base.Host); err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	canonical, err := os.ReadFile(filepath.Join(projectRoot, "canonical.md"))
	if err != nil || string(canonical) != "preserve canonical project state\n" {
		return errors.New("lifecycle changed canonical project state")
	}
	unowned, err := os.ReadFile(filepath.Join(hostRoot, "unowned.txt"))
	if err != nil || string(unowned) != "preserve unowned host state\n" {
		return errors.New("lifecycle changed unowned host state")
	}

	missingReceiptRoot := filepath.Join(parent, "missing-receipt-host")
	if err := installFixture(missingReceiptRoot, base, ""); err != nil {
		return err
	}
	if err := os.Remove(receiptPath(missingReceiptRoot, base.Host)); err != nil {
		return err
	}
	if err := removeFixture(missingReceiptRoot, base.Host); err == nil {
		return errors.New("removal accepted a missing receipt")
	}
	if _, err := os.Stat(packageDirectory(missingReceiptRoot, base.Host, base.ArchiveDigest)); err != nil {
		return errors.New("missing-receipt removal changed package bytes")
	}
	return nil
}

func syntheticLifecycleVariant(original builtPackage) (builtPackage, error) {
	variant := original
	variant.Version = original.Version + ".fixture"
	entries := cloneEntries(original.Entries)
	var payload []archiveEntry
	foundChangelog := false
	for index := range entries {
		if entries[index].Name == "INVENTORY.json" {
			continue
		}
		if entries[index].Name == "CHANGELOG.md" {
			entries[index].Data = append(entries[index].Data, []byte("\nSynthetic lifecycle upgrade fixture.\n")...)
			foundChangelog = true
		}
		payload = append(payload, entries[index])
	}
	if !foundChangelog {
		return builtPackage{}, errors.New("fixture variant lacks changelog")
	}
	inventoryBytes, err := jsonBytes(makeInventory(payload))
	if err != nil {
		return builtPackage{}, err
	}
	for index := range entries {
		if entries[index].Name == "INVENTORY.json" {
			entries[index].Data = inventoryBytes
		}
	}
	archive, err := createArchive(entries)
	if err != nil {
		return builtPackage{}, err
	}
	if err := validateArchive(archive, entries); err != nil {
		return builtPackage{}, err
	}
	variant.Entries = entries
	variant.Archive = archive
	variant.ArchiveDigest = sha256Hex(archive)
	variant.ArchiveName = strings.TrimSuffix(original.ArchiveName, ".zip") + "-fixture.zip"
	return variant, nil
}

func installFixture(root string, built builtPackage, fault string) error {
	if fault != "" && fault != "before-publish" && fault != "after-publish" {
		return errors.New("unknown fixture interruption point")
	}
	if err := validateBuiltPackage(built); err != nil {
		return err
	}
	if _, err := os.Lstat(pendingPath(root, built.Host)); err == nil {
		return errors.New("pending lifecycle transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	receipt, err := loadLifecycleReceipt(root, built.Host, false)
	if err != nil {
		return err
	}
	if receipt.Schema != 0 && receipt.ActiveDigest == built.ArchiveDigest {
		return verifyInstalledDirectory(packageDirectory(root, built.Host, built.ArchiveDigest), receiptPackageFromBuilt(built))
	}

	packageReceipt := receiptPackageFromBuilt(built)
	finalDirectory := packageDirectory(root, built.Host, built.ArchiveDigest)
	if _, statErr := os.Lstat(finalDirectory); statErr == nil {
		if err := verifyInstalledDirectory(finalDirectory, packageReceipt); err != nil {
			return err
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	} else {
		packagesRoot := filepath.Dir(finalDirectory)
		if err := os.MkdirAll(packagesRoot, 0o755); err != nil {
			return err
		}
		stage, err := os.MkdirTemp(packagesRoot, ".stage-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(stage)
		for _, entry := range built.Entries {
			if err := writeRegular(stage, entry.Name, entry.Data); err != nil {
				return err
			}
		}
		if err := verifyInstalledDirectory(stage, packageReceipt); err != nil {
			return err
		}
		pending := pendingInstall{Schema: 1, Host: built.Host, PreviousDigest: receipt.ActiveDigest, Package: packageReceipt}
		if err := writeLifecycleJSON(root, pendingRelative(built.Host), pending); err != nil {
			return err
		}
		if fault == "before-publish" {
			return errFixtureInterrupted
		}
		if err := os.Rename(stage, finalDirectory); err != nil {
			return err
		}
		stage = ""
		if fault == "after-publish" {
			return errFixtureInterrupted
		}
		return completePending(root, built.Host)
	}

	pending := pendingInstall{Schema: 1, Host: built.Host, PreviousDigest: receipt.ActiveDigest, Package: packageReceipt}
	if err := writeLifecycleJSON(root, pendingRelative(built.Host), pending); err != nil {
		return err
	}
	if fault != "" {
		return errFixtureInterrupted
	}
	return completePending(root, built.Host)
}

func recoverFixture(root, host string) error {
	pending, err := loadPending(root, host)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	directory := packageDirectory(root, host, pending.Package.Digest)
	if _, err := os.Lstat(directory); errors.Is(err, os.ErrNotExist) {
		return os.Remove(pendingPath(root, host))
	} else if err != nil {
		return err
	}
	if err := verifyInstalledDirectory(directory, pending.Package); err != nil {
		return err
	}
	return completePending(root, host)
}

func completePending(root, host string) error {
	pending, err := loadPending(root, host)
	if err != nil {
		return err
	}
	receipt, err := loadLifecycleReceipt(root, host, false)
	if err != nil {
		return err
	}
	if receipt.Schema == 0 {
		receipt = lifecycleReceipt{Schema: 1, Host: host}
	}
	if receipt.ActiveDigest != pending.PreviousDigest {
		return errors.New("active receipt changed during pending install")
	}
	found := false
	for index := range receipt.Packages {
		if receipt.Packages[index].Digest == pending.Package.Digest {
			receipt.Packages[index] = pending.Package
			found = true
		}
	}
	if !found {
		receipt.Packages = append(receipt.Packages, pending.Package)
	}
	sort.Slice(receipt.Packages, func(i, j int) bool { return receipt.Packages[i].Digest < receipt.Packages[j].Digest })
	receipt.PreviousDigest = receipt.ActiveDigest
	receipt.ActiveDigest = pending.Package.Digest
	if err := writeLifecycleJSON(root, receiptRelative(host), receipt); err != nil {
		return err
	}
	return os.Remove(pendingPath(root, host))
}

func rollbackFixture(root, host string) error {
	if _, err := os.Lstat(pendingPath(root, host)); err == nil {
		return errors.New("cannot roll back with a pending transaction")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	receipt, err := loadLifecycleReceipt(root, host, true)
	if err != nil {
		return err
	}
	if receipt.PreviousDigest == "" || receipt.PreviousDigest == receipt.ActiveDigest {
		return errors.New("no prior package is available for rollback")
	}
	previous, ok := findReceiptPackage(receipt, receipt.PreviousDigest)
	if !ok {
		return errors.New("prior package receipt is missing")
	}
	if err := verifyInstalledDirectory(packageDirectory(root, host, previous.Digest), previous); err != nil {
		return err
	}
	receipt.ActiveDigest, receipt.PreviousDigest = receipt.PreviousDigest, receipt.ActiveDigest
	return writeLifecycleJSON(root, receiptRelative(host), receipt)
}

func prepareRemoval(root, host string) (removalPreview, error) {
	preview := removalPreview{Host: host}
	if _, err := os.Lstat(pendingPath(root, host)); err == nil {
		preview.Conflicts = append(preview.Conflicts, "pending lifecycle transaction")
	} else if !errors.Is(err, os.ErrNotExist) {
		return preview, err
	}
	receipt, err := loadLifecycleReceipt(root, host, true)
	if err != nil {
		return preview, err
	}
	known := make(map[string]bool, len(receipt.Packages))
	for _, installed := range receipt.Packages {
		known[installed.Digest] = true
		preview.OwnedPackages = append(preview.OwnedPackages, installed.Digest)
		if err := verifyInstalledDirectory(packageDirectory(root, host, installed.Digest), installed); err != nil {
			preview.Conflicts = append(preview.Conflicts, installed.Digest+": "+err.Error())
		}
	}
	sort.Strings(preview.OwnedPackages)
	packagesRoot := filepath.Join(root, "packages", host)
	entries, readErr := os.ReadDir(packagesRoot)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return preview, readErr
	}
	for _, entry := range entries {
		if !entry.IsDir() || !known[entry.Name()] {
			preview.Conflicts = append(preview.Conflicts, "unowned package entry: "+entry.Name())
		}
	}
	sort.Strings(preview.Conflicts)
	return preview, nil
}

func removeFixture(root, host string) error {
	preview, err := prepareRemoval(root, host)
	if err != nil {
		return err
	}
	if len(preview.Conflicts) != 0 {
		return fmt.Errorf("removal blocked by conflicts: %s", strings.Join(preview.Conflicts, "; "))
	}
	receipt, err := loadLifecycleReceipt(root, host, true)
	if err != nil {
		return err
	}
	for _, installed := range receipt.Packages {
		packageRoot := packageDirectory(root, host, installed.Digest)
		files := append([]receiptFile{}, installed.Files...)
		sort.Slice(files, func(i, j int) bool { return strings.Count(files[i].Path, "/") > strings.Count(files[j].Path, "/") })
		for _, file := range files {
			if err := os.Remove(filepath.Join(packageRoot, filepath.FromSlash(file.Path))); err != nil {
				return err
			}
			removeEmptyParents(filepath.Dir(filepath.Join(packageRoot, filepath.FromSlash(file.Path))), packageRoot)
		}
		if err := os.Remove(packageRoot); err != nil {
			return err
		}
	}
	if err := os.Remove(receiptPath(root, host)); err != nil {
		return err
	}
	removeEmptyParents(filepath.Dir(receiptPath(root, host)), root)
	removeEmptyParents(filepath.Join(root, "packages", host), root)
	return nil
}

func removeEmptyParents(directory, stop string) {
	for directory != stop && directory != "." && directory != string(filepath.Separator) {
		if err := os.Remove(directory); err != nil {
			return
		}
		directory = filepath.Dir(directory)
	}
}

func validateBuiltPackage(built builtPackage) error {
	if built.Host != "codex" && built.Host != "claude" {
		return errors.New("fixture package host is invalid")
	}
	if !versionPattern.MatchString(built.Version) && !strings.HasSuffix(built.Version, ".fixture") {
		return errors.New("fixture package version is invalid")
	}
	if sha256Hex(built.Archive) != built.ArchiveDigest {
		return errors.New("fixture package digest mismatch")
	}
	return validateArchive(built.Archive, built.Entries)
}

func receiptPackageFromBuilt(built builtPackage) receiptPackage {
	files := make([]receiptFile, 0, len(built.Entries))
	for _, entry := range built.Entries {
		files = append(files, receiptFile{Path: entry.Name, Size: len(entry.Data), SHA256: sha256Hex(entry.Data)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return receiptPackage{Version: built.Version, Digest: built.ArchiveDigest, Files: files}
}

func verifyInstalledDirectory(root string, receipt receiptPackage) error {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return errors.New("installed package directory is missing or unsafe")
	}
	expected := make(map[string]receiptFile, len(receipt.Files))
	for _, file := range receipt.Files {
		if err := validateArchiveName(file.Path); err != nil || file.Size < 1 || file.Size > maximumFileSize || len(file.SHA256) != 64 {
			return errors.New("installed package receipt contains an invalid file")
		}
		expected[file.Path] = file
	}
	seen := make(map[string]bool, len(expected))
	err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == root {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("installed package contains symlink %q", relative)
		}
		if entry.IsDir() {
			return nil
		}
		file, ok := expected[relative]
		if !ok {
			return fmt.Errorf("installed package contains unowned file %q", relative)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o644 || info.Size() != int64(file.Size) {
			return fmt.Errorf("installed file metadata changed: %q", relative)
		}
		data, err := os.ReadFile(name)
		if err != nil || sha256Hex(data) != file.SHA256 {
			return fmt.Errorf("installed file content changed: %q", relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return errors.New("installed package is missing an owned file")
	}
	return nil
}

func loadLifecycleReceipt(root, host string, required bool) (lifecycleReceipt, error) {
	if host != "codex" && host != "claude" {
		return lifecycleReceipt{}, errors.New("invalid lifecycle host")
	}
	data, err := os.ReadFile(receiptPath(root, host))
	if errors.Is(err, os.ErrNotExist) && !required {
		return lifecycleReceipt{}, nil
	}
	if err != nil {
		return lifecycleReceipt{}, err
	}
	if len(data) == 0 || len(data) > maximumFileSize {
		return lifecycleReceipt{}, errors.New("lifecycle receipt is empty or oversized")
	}
	var receipt lifecycleReceipt
	if err := decodeStrict(data, &receipt); err != nil {
		return lifecycleReceipt{}, err
	}
	if receipt.Schema != 1 || receipt.Host != host || len(receipt.ActiveDigest) != 64 || len(receipt.Packages) == 0 {
		return lifecycleReceipt{}, errors.New("lifecycle receipt identity is invalid")
	}
	if _, ok := findReceiptPackage(receipt, receipt.ActiveDigest); !ok {
		return lifecycleReceipt{}, errors.New("active package is absent from lifecycle receipt")
	}
	return receipt, nil
}

func loadPending(root, host string) (pendingInstall, error) {
	data, err := os.ReadFile(pendingPath(root, host))
	if err != nil {
		return pendingInstall{}, err
	}
	if len(data) == 0 || len(data) > maximumFileSize {
		return pendingInstall{}, errors.New("pending lifecycle record is empty or oversized")
	}
	var pending pendingInstall
	if err := decodeStrict(data, &pending); err != nil {
		return pendingInstall{}, err
	}
	if pending.Schema != 1 || pending.Host != host || len(pending.Package.Digest) != 64 || len(pending.Package.Files) == 0 {
		return pendingInstall{}, errors.New("pending lifecycle record identity is invalid")
	}
	return pending, nil
}

func writeLifecycleJSON(root, relative string, value any) error {
	data, err := jsonBytes(value)
	if err != nil {
		return err
	}
	return writeRegular(root, relative, data)
}

func findReceiptPackage(receipt lifecycleReceipt, digest string) (receiptPackage, bool) {
	for _, installed := range receipt.Packages {
		if installed.Digest == digest {
			return installed, true
		}
	}
	return receiptPackage{}, false
}

func receiptRelative(host string) string {
	return filepath.ToSlash(filepath.Join("receipts", host+".json"))
}

func pendingRelative(host string) string {
	return filepath.ToSlash(filepath.Join("pending", host+".json"))
}

func receiptPath(root, host string) string {
	return filepath.Join(root, filepath.FromSlash(receiptRelative(host)))
}

func pendingPath(root, host string) string {
	return filepath.Join(root, filepath.FromSlash(pendingRelative(host)))
}

func packageDirectory(root, host, digest string) string {
	return filepath.Join(root, "packages", host, digest)
}
