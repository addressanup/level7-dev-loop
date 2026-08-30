package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path"
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
	Schema                 int            `json:"schema"`
	Host                   string         `json:"host"`
	PreviousDigest         string         `json:"previous_digest"`
	BaseReceiptDigest      string         `json:"base_receipt_digest"`
	CommittedReceiptDigest string         `json:"committed_receipt_digest"`
	Package                receiptPackage `json:"package"`
}

type pendingRemoval struct {
	Schema  int              `json:"schema"`
	Host    string           `json:"host"`
	Receipt lifecycleReceipt `json:"receipt"`
}

type removalPreview struct {
	Host          string
	OwnedPackages []string
	Conflicts     []string
}

type lifecycleQualificationPackage struct {
	Host          string
	ArchiveSHA256 string
	CatalogSHA256 string
}

type lifecycleQualification struct {
	Packages []lifecycleQualificationPackage
}

func qualifyLifecycleSet(packages []builtPackage) (lifecycleQualification, error) {
	if err := validateQualificationPackageOrder(packages); err != nil {
		return lifecycleQualification{}, fmt.Errorf("preflight package set: %w", err)
	}
	qualified := lifecycleQualification{Packages: make([]lifecycleQualificationPackage, 0, len(packages))}
	for _, built := range packages {
		parent, err := os.MkdirTemp("", "l7-distribution-lifecycle-")
		if err != nil {
			return lifecycleQualification{}, err
		}
		qualificationErr := qualifyOneLifecycle(parent, built)
		cleanupErr := os.RemoveAll(parent)
		if err := errors.Join(qualificationErr, cleanupErr); err != nil {
			return lifecycleQualification{}, fmt.Errorf("%s: %w", built.Host, err)
		}
		qualified.Packages = append(qualified.Packages, lifecycleQualificationPackage{
			Host: built.Host, ArchiveSHA256: built.ArchiveDigest, CatalogSHA256: built.CatalogDigest,
		})
	}
	return qualified, nil
}

func validateLifecycleQualification(qualification lifecycleQualification, packages []builtPackage) error {
	if len(qualification.Packages) != len(packages) || len(packages) != 2 {
		return errors.New("fixture lifecycle identity count is invalid")
	}
	for index, built := range packages {
		observed := qualification.Packages[index]
		if observed.Host != built.Host || observed.ArchiveSHA256 != built.ArchiveDigest ||
			observed.CatalogSHA256 != built.CatalogDigest || !sha256Pattern.MatchString(observed.ArchiveSHA256) ||
			!sha256Pattern.MatchString(observed.CatalogSHA256) {
			return fmt.Errorf("fixture lifecycle identity %d does not bind the %s package", index, built.Host)
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
	var err error
	variant.Version, err = nextStablePatchVersion(original.Version)
	if err != nil {
		return builtPackage{}, err
	}
	variant.Catalog = append([]byte{}, original.Catalog...)
	if original.Host == "claude" {
		var catalog claudeCatalog
		if err := decodeStrict(variant.Catalog, &catalog); err != nil || len(catalog.Plugins) != 1 || catalog.Plugins[0].Version != original.Version {
			return builtPackage{}, errors.New("fixture variant Claude catalog identity is inconsistent")
		}
		catalog.Plugins[0].Version = variant.Version
		variant.Catalog, err = jsonBytes(catalog)
		if err != nil {
			return builtPackage{}, err
		}
	}
	variant.CatalogDigest = sha256Hex(variant.Catalog)
	entries := cloneEntries(original.Entries)
	manifestPath, _, _, err := packageIdentityPaths(original.Host)
	if err != nil {
		return builtPackage{}, err
	}
	variantSourceDigest := sha256Hex([]byte("synthetic-lifecycle-variant\x00" + original.ArchiveDigest + "\x00" + variant.Version))
	var payload []archiveEntry
	found := make(map[string]bool)
	for index := range entries {
		switch entries[index].Name {
		case manifestPath:
			if original.Host == "codex" {
				var manifest codexManifest
				if err := decodeStrict(entries[index].Data, &manifest); err != nil {
					return builtPackage{}, err
				}
				manifest.Version = variant.Version
				entries[index].Data, err = jsonBytes(manifest)
			} else {
				var manifest claudeManifest
				if err := decodeStrict(entries[index].Data, &manifest); err != nil {
					return builtPackage{}, err
				}
				manifest.Version = variant.Version
				entries[index].Data, err = jsonBytes(manifest)
			}
			if err != nil {
				return builtPackage{}, err
			}
			found["manifest"] = true
		case "COMPATIBILITY.json":
			var compatibility compatibilityProjection
			if err := decodeStrict(entries[index].Data, &compatibility); err != nil {
				return builtPackage{}, err
			}
			if compatibility.PackageVersion != original.Version || compatibility.Entry.Host != original.Host {
				return builtPackage{}, errors.New("fixture variant compatibility identity is inconsistent")
			}
			compatibility.PackageVersion = variant.Version
			entries[index].Data, err = jsonBytes(compatibility)
			if err != nil {
				return builtPackage{}, err
			}
			found["compatibility"] = true
		case "DISTRIBUTION.json":
			var metadata distributionMetadata
			if err := decodeStrict(entries[index].Data, &metadata); err != nil {
				return builtPackage{}, err
			}
			if metadata.Host != original.Host || metadata.Version != original.Version {
				return builtPackage{}, errors.New("fixture variant distribution identity is inconsistent")
			}
			metadata.Version = variant.Version
			metadata.CatalogSHA256 = variant.CatalogDigest
			metadata.SourceDigest = variantSourceDigest
			entries[index].Data, err = jsonBytes(metadata)
			if err != nil {
				return builtPackage{}, err
			}
			found["distribution"] = true
		case "PERMISSIONS.json":
			var permission permissionsProjection
			if err := decodeStrict(entries[index].Data, &permission); err != nil {
				return builtPackage{}, err
			}
			if permission.PackageVersion != original.Version || permission.Host != original.Host {
				return builtPackage{}, errors.New("fixture variant permission identity is inconsistent")
			}
			permission.PackageVersion = variant.Version
			entries[index].Data, err = jsonBytes(permission)
			if err != nil {
				return builtPackage{}, err
			}
			found["permissions"] = true
		case "PROVENANCE.input.json":
			var provenance provenanceInput
			if err := decodeStrict(entries[index].Data, &provenance); err != nil {
				return builtPackage{}, err
			}
			if provenance.Version != original.Version || provenance.Host != original.Host {
				return builtPackage{}, errors.New("fixture variant provenance identity is inconsistent")
			}
			provenance.Version = variant.Version
			provenance.SourceDigest = variantSourceDigest
			entries[index].Data, err = jsonBytes(provenance)
			if err != nil {
				return builtPackage{}, err
			}
			found["provenance"] = true
		case "SBOM.spdx.json":
			var sbom sbomDocument
			if err := decodeStrict(entries[index].Data, &sbom); err != nil {
				return builtPackage{}, err
			}
			marker := "/sbom/" + original.Host + "/"
			markerIndex := strings.LastIndex(sbom.DocumentNamespace, marker)
			if markerIndex < 0 || len(sbom.Packages) != 1 || sbom.Packages[0].VersionInfo != original.Version || len(sbom.Packages[0].Checksums) != 1 {
				return builtPackage{}, errors.New("fixture variant SBOM identity is inconsistent")
			}
			sbom.Name = "level7-dev-loop-" + original.Host + "-" + variant.Version
			sbom.DocumentNamespace = sbom.DocumentNamespace[:markerIndex] + marker + variantSourceDigest
			sbom.Packages[0].VersionInfo = variant.Version
			sbom.Packages[0].Checksums[0].ChecksumValue = variantSourceDigest
			entries[index].Data, err = jsonBytes(sbom)
			if err != nil {
				return builtPackage{}, err
			}
			found["sbom"] = true
		case "CHANGELOG.md":
			entries[index].Data = append(entries[index].Data, []byte("\nSynthetic lifecycle upgrade fixture.\n")...)
			found["changelog"] = true
		}
		if entries[index].Name != "INVENTORY.json" {
			payload = append(payload, entries[index])
		}
	}
	for _, required := range []string{"manifest", "compatibility", "distribution", "permissions", "provenance", "sbom", "changelog"} {
		if !found[required] {
			return builtPackage{}, errors.New("fixture variant lacks identity metadata or changelog")
		}
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
	variant.SourceDigest = variantSourceDigest
	variant.ArchiveDigest = sha256Hex(archive)
	variant.ArchiveName = strings.TrimSuffix(original.ArchiveName, ".zip") + "-fixture.zip"
	if err := validateBuiltPackage(variant); err != nil {
		return builtPackage{}, err
	}
	return variant, nil
}

func nextStablePatchVersion(version string) (string, error) {
	if !versionPattern.MatchString(version) {
		return "", errors.New("stable package version is invalid")
	}
	components := strings.Split(version, ".")
	if len(components) != 3 {
		return "", errors.New("stable package patch version is missing")
	}
	patch, ok := new(big.Int).SetString(components[2], 10)
	if !ok || patch.Sign() < 0 {
		return "", errors.New("stable package patch version cannot be advanced")
	}
	patch.Add(patch, big.NewInt(1))
	return components[0] + "." + components[1] + "." + patch.String(), nil
}

func installFixture(root string, built builtPackage, fault string) error {
	if fault != "" && fault != "before-publish" && fault != "after-publish" && fault != "after-receipt" {
		return errors.New("unknown fixture interruption point")
	}
	if err := validateBuiltPackage(built); err != nil {
		return err
	}
	if _, err := loadPendingRemoval(root, built.Host); err == nil {
		return errors.New("pending removal transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, err := loadPending(root, built.Host); err == nil {
		return errors.New("pending lifecycle transaction requires recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	receipt, err := loadLifecycleReceipt(root, built.Host, false)
	if err != nil {
		return err
	}
	if receipt.Schema != 0 && receipt.ActiveDigest == built.ArchiveDigest {
		installed, ok := findReceiptPackage(receipt, built.ArchiveDigest)
		wanted := receiptPackageFromBuilt(built)
		if !ok || !equalReceiptPackage(installed, wanted) {
			return errors.New("same-digest reinstall does not match its ownership receipt")
		}
		if err := verifyInstallBaseState(root, built.Host, receipt); err != nil {
			return err
		}
		if err := ensureDurableRoot(root); err != nil {
			return err
		}
		for _, relative := range []string{pendingRelative(built.Host), pendingRemovalRelative(built.Host), receiptRelative(built.Host)} {
			if err := syncLifecycleRecordParent(root, relative); err != nil {
				return err
			}
		}
		return syncDirectoryTree(root, filepath.Join("packages", built.Host))
	}

	packageReceipt := receiptPackageFromBuilt(built)
	pending, nextReceipt, err := newPendingInstall(receipt, built.Host, packageReceipt)
	if err != nil {
		return err
	}
	if err := validateLifecycleJSONBound(pending); err != nil {
		return err
	}
	if err := validateLifecycleJSONBound(nextReceipt); err != nil {
		return err
	}
	if err := verifyInstallBaseState(root, built.Host, receipt); err != nil {
		return err
	}
	finalDirectory := packageDirectory(root, built.Host, built.ArchiveDigest)
	stage := pendingStageDirectory(root, built.Host, built.ArchiveDigest)
	stagePresent, err := physicalDirectoryState(root, pendingStageRelative(built.Host, built.ArchiveDigest))
	if err != nil {
		return err
	}
	if stagePresent {
		return errors.New("journal-free lifecycle stage requires manual recovery")
	}
	if err := ensureDurableRoot(root); err != nil {
		return err
	}
	for _, relative := range []string{pendingRelative(built.Host), pendingRemovalRelative(built.Host), receiptRelative(built.Host)} {
		if err := syncLifecycleRecordParent(root, relative); err != nil {
			return err
		}
	}
	if err := syncNearestExistingDirectory(root, filepath.Join(root, "packages", built.Host)); err != nil {
		return err
	}
	if _, statErr := os.Lstat(finalDirectory); statErr == nil {
		if err := verifyInstalledDirectory(root, finalDirectory, packageReceipt); err != nil {
			return err
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	} else {
		if err := writeLifecycleJSON(root, pendingRelative(built.Host), pending); err != nil {
			return err
		}
		if err := makeDurableDirectory(root, pendingStageRelative(built.Host, built.ArchiveDigest)); err != nil {
			return err
		}
		for _, entry := range built.Entries {
			if err := writeStagedRegular(stage, entry); err != nil {
				return err
			}
		}
		if err := verifyInstalledDirectory(root, stage, packageReceipt); err != nil {
			return err
		}
		if fault == "before-publish" {
			return errFixtureInterrupted
		}
		if err := renameDurably(root, pendingStageRelative(built.Host, built.ArchiveDigest), filepath.Join("packages", built.Host, built.ArchiveDigest)); err != nil {
			return err
		}
		if fault == "after-publish" {
			return errFixtureInterrupted
		}
		return completePending(root, built.Host, fault)
	}

	if err := writeLifecycleJSON(root, pendingRelative(built.Host), pending); err != nil {
		return err
	}
	if fault == "before-publish" || fault == "after-publish" {
		return errFixtureInterrupted
	}
	return completePending(root, built.Host, fault)
}

func recoverFixture(root, host string) error {
	pending, installErr := loadPending(root, host)
	removal, removalErr := loadPendingRemoval(root, host)
	installPresent := installErr == nil
	removalPresent := removalErr == nil
	if installErr != nil && !errors.Is(installErr, os.ErrNotExist) {
		return installErr
	}
	if removalErr != nil && !errors.Is(removalErr, os.ErrNotExist) {
		return removalErr
	}
	if !installPresent && !removalPresent {
		if _, err := os.Lstat(root); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
	}
	if err := ensureDurableRoot(root); err != nil {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRelative(host)); err != nil {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRemovalRelative(host)); err != nil {
		return err
	}
	if installPresent && removalPresent {
		return errors.New("conflicting install and removal transactions require manual recovery")
	}
	if removalPresent {
		return completePendingRemoval(root, host, removal, "")
	}
	if !installPresent {
		if err := syncLifecycleRecordParent(root, receiptRelative(host)); err != nil {
			return err
		}
		return syncNearestExistingDirectory(root, filepath.Join(root, "packages", host))
	}
	if err := ensurePhysicalDirectory(root, filepath.Join("packages", host), false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			receipt, receiptErr := loadLifecycleReceipt(root, host, false)
			if receiptErr != nil {
				return receiptErr
			}
			if err := verifyPendingBaseReceipt(receipt, pending); err != nil {
				return err
			}
			if err := verifyInstallBaseState(root, host, receipt); err != nil {
				return err
			}
			if err := syncNearestExistingDirectory(root, filepath.Join(root, "packages", host)); err != nil {
				return err
			}
			return removeDurably(root, pendingRelative(host), true)
		}
		return err
	}
	directoryRelative := filepath.Join("packages", host, pending.Package.Digest)
	directoryPresent, err := physicalDirectoryState(root, directoryRelative)
	if err != nil {
		return err
	}
	stagePresent, err := physicalDirectoryState(root, pendingStageRelative(host, pending.Package.Digest))
	if err != nil {
		return err
	}
	if !directoryPresent {
		receipt, err := loadLifecycleReceipt(root, host, false)
		if err != nil {
			return err
		}
		if err := verifyPendingBaseReceipt(receipt, pending); err != nil {
			return err
		}
		if stagePresent {
			if err := syncDirectoryTree(root, filepath.Dir(pendingStageRelative(host, pending.Package.Digest))); err != nil {
				return err
			}
			if err := cleanupPendingStage(root, host, pending.Package); err != nil {
				return err
			}
		} else if err := syncNearestExistingDirectory(root, filepath.Dir(packageDirectory(root, host, pending.Package.Digest))); err != nil {
			return err
		}
		if err := verifyInstallBaseState(root, host, receipt); err != nil {
			return err
		}
		return removeDurably(root, pendingRelative(host), true)
	}
	if stagePresent {
		return errors.New("pending install has both staged and published package trees")
	}
	if err := syncRenameDirectories(root, pendingStageRelative(host, pending.Package.Digest), directoryRelative); err != nil {
		return err
	}
	directory := filepath.Join(root, directoryRelative)
	if err := verifyInstalledDirectory(root, directory, pending.Package); err != nil {
		return err
	}
	return completePending(root, host, "")
}

func verifyInstallBaseState(root, host string, receipt lifecycleReceipt) error {
	hostRelative := filepath.Join("packages", host)
	present, err := physicalDirectoryState(root, hostRelative)
	if err != nil {
		return err
	}
	if receipt.Schema == 0 {
		if !present {
			return nil
		}
		entries, err := os.ReadDir(filepath.Join(root, hostRelative))
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return errors.New("receipt-free package root contains unowned state")
		}
		return nil
	}
	if !present {
		return errors.New("receipt-bound package root is missing")
	}
	return verifyReceiptPackageSet(root, hostRelative, receipt, false)
}

func cleanupPendingStage(root, host string, receipt receiptPackage) error {
	stage := pendingStageDirectory(root, host, receipt.Digest)
	if err := verifyPendingStageRemainder(root, stage, receipt); err != nil {
		return err
	}
	files := append([]receiptFile{}, receipt.Files...)
	sort.Slice(files, func(i, j int) bool {
		firstDepth, secondDepth := strings.Count(files[i].Path, "/"), strings.Count(files[j].Path, "/")
		if firstDepth == secondDepth {
			return files[i].Path < files[j].Path
		}
		return firstDepth > secondDepth
	})
	for _, file := range files {
		exists, err := validatePendingStageFile(stage, file)
		if err != nil {
			return err
		}
		name := filepath.Join(stage, filepath.FromSlash(file.Path))
		fileRelative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if err := removeDurably(root, fileRelative, !exists); err != nil {
			return err
		}
		if err := removeEmptyParents(filepath.Dir(name), stage); err != nil {
			return err
		}
	}
	stageRelative, err := filepath.Rel(root, stage)
	if err != nil {
		return err
	}
	return removeDurably(root, stageRelative, true)
}

func writeStagedRegular(stage string, entry archiveEntry) (result error) {
	if err := validateArchiveEntries([]archiveEntry{entry}); err != nil {
		return err
	}
	clean, err := cleanRelativePath(entry.Name)
	if err != nil {
		return fmt.Errorf("unsafe staged package path %q: %w", entry.Name, err)
	}
	if err := ensureDurableDirectory(stage, filepath.Dir(clean)); err != nil {
		return err
	}
	target := filepath.Join(stage, clean)
	within, err := filepath.Rel(stage, target)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return fmt.Errorf("staged package path escapes root: %q", entry.Name)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	fileOpen := true
	fileLinked := true
	defer func() {
		if fileOpen {
			result = errors.Join(result, file.Close())
		}
		if fileLinked {
			result = errors.Join(result, removeDurably(stage, clean, true))
		}
	}()
	if _, err := file.Write(entry.Data); err != nil {
		return err
	}
	if err := file.Chmod(entry.Mode.Perm()); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		fileOpen = false
		return err
	}
	fileOpen = false
	if err := lifecycleSyncDirectory(filepath.Dir(target)); err != nil {
		return err
	}
	fileLinked = false
	return nil
}

func completePending(root, host, fault string) error {
	pending, err := loadPending(root, host)
	if err != nil {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRelative(host)); err != nil {
		return err
	}
	receipt, err := loadLifecycleReceipt(root, host, false)
	if err != nil {
		return err
	}
	if receipt.ActiveDigest == pending.Package.Digest {
		currentDigest, digestErr := lifecycleReceiptStateDigest(receipt)
		if digestErr != nil || currentDigest != pending.CommittedReceiptDigest {
			return errors.New("pending install conflicts with the committed receipt state")
		}
		installed, ok := findReceiptPackage(receipt, pending.Package.Digest)
		if receipt.PreviousDigest != pending.PreviousDigest || !ok || !equalReceiptPackage(installed, pending.Package) {
			return errors.New("pending install conflicts with the committed receipt")
		}
		if err := verifyReceiptPackageSet(root, filepath.Join("packages", host), receipt, false); err != nil {
			return err
		}
		if err := syncDirectoryTree(root, filepath.Join("packages", host)); err != nil {
			return err
		}
		if err := syncLifecycleRecordParent(root, receiptRelative(host)); err != nil {
			return err
		}
		return removeDurably(root, pendingRelative(host), true)
	}
	nextReceipt, err := prospectiveLifecycleReceipt(receipt, pending)
	if err != nil {
		return err
	}
	nextDigest, err := lifecycleReceiptStateDigest(nextReceipt)
	if err != nil {
		return err
	}
	if nextDigest != pending.CommittedReceiptDigest {
		return errors.New("pending install committed receipt binding changed")
	}
	if err := verifyReceiptPackageSet(root, filepath.Join("packages", host), nextReceipt, false); err != nil {
		return err
	}
	if err := syncDirectoryTree(root, filepath.Join("packages", host)); err != nil {
		return err
	}
	if err := writeLifecycleJSON(root, receiptRelative(host), nextReceipt); err != nil {
		return err
	}
	if fault == "after-receipt" {
		return errFixtureInterrupted
	}
	return removeDurably(root, pendingRelative(host), true)
}

func newPendingInstall(receipt lifecycleReceipt, host string, installed receiptPackage) (pendingInstall, lifecycleReceipt, error) {
	baseDigest, err := lifecycleReceiptStateDigest(receipt)
	if err != nil {
		return pendingInstall{}, lifecycleReceipt{}, err
	}
	pending := pendingInstall{
		Schema:            1,
		Host:              host,
		PreviousDigest:    receipt.ActiveDigest,
		BaseReceiptDigest: baseDigest,
		Package:           installed,
	}
	next, err := prospectiveLifecycleReceipt(receipt, pending)
	if err != nil {
		return pendingInstall{}, lifecycleReceipt{}, err
	}
	pending.CommittedReceiptDigest, err = lifecycleReceiptStateDigest(next)
	if err != nil {
		return pendingInstall{}, lifecycleReceipt{}, err
	}
	if err := validatePendingInstall(pending, host); err != nil {
		return pendingInstall{}, lifecycleReceipt{}, err
	}
	return pending, next, nil
}

func prospectiveLifecycleReceipt(receipt lifecycleReceipt, pending pendingInstall) (lifecycleReceipt, error) {
	if err := validatePendingInstallBase(pending, pending.Host); err != nil {
		return lifecycleReceipt{}, err
	}
	if err := verifyPendingBaseReceipt(receipt, pending); err != nil {
		return lifecycleReceipt{}, err
	}
	next := receipt
	if next.Schema == 0 {
		next = lifecycleReceipt{Schema: 1, Host: pending.Host}
	} else {
		next.Packages = append([]receiptPackage{}, receipt.Packages...)
	}
	if next.Host != pending.Host || next.ActiveDigest != pending.PreviousDigest {
		return lifecycleReceipt{}, errors.New("active receipt changed during pending install")
	}
	found := false
	for index := range next.Packages {
		if next.Packages[index].Digest == pending.Package.Digest {
			if !equalReceiptPackage(next.Packages[index], pending.Package) {
				return lifecycleReceipt{}, errors.New("pending install relabels a known package digest")
			}
			found = true
		}
	}
	if !found {
		next.Packages = append(next.Packages, pending.Package)
	}
	sort.Slice(next.Packages, func(i, j int) bool { return next.Packages[i].Digest < next.Packages[j].Digest })
	next.PreviousDigest = next.ActiveDigest
	next.ActiveDigest = pending.Package.Digest
	if err := validateLifecycleReceipt(next, pending.Host); err != nil {
		return lifecycleReceipt{}, err
	}
	return next, nil
}

func verifyPendingBaseReceipt(receipt lifecycleReceipt, pending pendingInstall) error {
	digest, err := lifecycleReceiptStateDigest(receipt)
	if err != nil {
		return err
	}
	if digest != pending.BaseReceiptDigest || receipt.ActiveDigest != pending.PreviousDigest {
		return errors.New("pending install base receipt changed")
	}
	return nil
}

func rollbackFixture(root, host string) error {
	if _, err := loadPendingRemoval(root, host); err == nil {
		return errors.New("cannot roll back with a pending removal")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, err := loadPending(root, host); err == nil {
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
	if err := verifyInstalledDirectory(root, packageDirectory(root, host, previous.Digest), previous); err != nil {
		return err
	}
	if err := ensureDurableRoot(root); err != nil {
		return err
	}
	for _, relative := range []string{pendingRelative(host), pendingRemovalRelative(host)} {
		if err := syncLifecycleRecordParent(root, relative); err != nil {
			return err
		}
	}
	if err := syncDirectoryTree(root, filepath.Join("packages", host)); err != nil {
		return err
	}
	receipt.ActiveDigest, receipt.PreviousDigest = receipt.PreviousDigest, receipt.ActiveDigest
	return writeLifecycleJSON(root, receiptRelative(host), receipt)
}

func prepareRemoval(root, host string) (removalPreview, error) {
	preview := removalPreview{Host: host}
	if _, err := loadPending(root, host); err == nil {
		preview.Conflicts = append(preview.Conflicts, "pending lifecycle transaction")
	} else if !errors.Is(err, os.ErrNotExist) {
		return preview, err
	}
	if _, err := loadPendingRemoval(root, host); err == nil {
		preview.Conflicts = append(preview.Conflicts, "pending removal transaction requires recovery")
		sort.Strings(preview.Conflicts)
		return preview, nil
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
		if err := verifyInstalledDirectory(root, packageDirectory(root, host, installed.Digest), installed); err != nil {
			preview.Conflicts = append(preview.Conflicts, installed.Digest+": "+err.Error())
		}
	}
	sort.Strings(preview.OwnedPackages)
	packagesRoot := filepath.Join(root, "packages", host)
	if err := ensurePhysicalDirectory(root, filepath.Join("packages", host), false); err != nil {
		return preview, err
	}
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
	return removeFixtureWithFault(root, host, "")
}

func removeFixtureWithFault(root, host, fault string) error {
	if fault != "" && fault != "after-removal-journal" && fault != "after-removal-rename" &&
		fault != "after-first-owned-delete" && fault != "after-removal-tree" && fault != "after-removal-receipt" {
		return errors.New("unknown fixture removal interruption point")
	}
	if _, err := loadPending(root, host); err == nil {
		return errors.New("cannot remove with a pending install")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := ensureDurableRoot(root); err != nil {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRelative(host)); err != nil {
		return err
	}

	removal, err := loadPendingRemoval(root, host)
	if errors.Is(err, os.ErrNotExist) {
		preview, previewErr := prepareRemoval(root, host)
		if previewErr != nil {
			return previewErr
		}
		if len(preview.Conflicts) != 0 {
			return fmt.Errorf("removal blocked by conflicts: %s", strings.Join(preview.Conflicts, "; "))
		}
		receipt, receiptErr := loadLifecycleReceipt(root, host, true)
		if receiptErr != nil {
			return receiptErr
		}
		removal = pendingRemoval{Schema: 1, Host: host, Receipt: receipt}
		if err := writeLifecycleJSON(root, pendingRemovalRelative(host), removal); err != nil {
			return err
		}
		if fault == "after-removal-journal" {
			return errFixtureInterrupted
		}
	} else if err != nil {
		return err
	}
	return completePendingRemoval(root, host, removal, fault)
}

func completePendingRemoval(root, host string, removal pendingRemoval, fault string) error {
	if err := validatePendingRemoval(removal, host); err != nil {
		return err
	}
	if _, err := loadPending(root, host); err == nil {
		return errors.New("pending install conflicts with removal recovery")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRemovalRelative(host)); err != nil {
		return err
	}
	if err := syncLifecycleRecordParent(root, pendingRelative(host)); err != nil {
		return err
	}

	receipt, err := loadLifecycleReceipt(root, host, false)
	if err != nil {
		return err
	}
	receiptPresent := receipt.Schema != 0
	if receiptPresent && !equalLifecycleReceipt(receipt, removal.Receipt) {
		return errors.New("lifecycle receipt changed during pending removal")
	}

	sourceRelative := filepath.Join("packages", host)
	quarantineRelative := filepath.Join("removing", host)
	sourcePresent, err := physicalDirectoryState(root, sourceRelative)
	if err != nil {
		return err
	}
	quarantinePresent, err := physicalDirectoryState(root, quarantineRelative)
	if err != nil {
		return err
	}
	if sourcePresent && quarantinePresent {
		return errors.New("pending removal has both published and quarantined package trees")
	}
	if (sourcePresent || quarantinePresent) && !receiptPresent {
		return errors.New("pending removal package tree exists without its bound receipt")
	}

	if sourcePresent {
		if err := verifyReceiptPackageSet(root, sourceRelative, removal.Receipt, false); err != nil {
			return err
		}
		if err := ensureDurableDirectory(root, "removing"); err != nil {
			return err
		}
		if err := renameDurably(root, sourceRelative, quarantineRelative); err != nil {
			return err
		}
		sourcePresent = false
		quarantinePresent = true
		if fault == "after-removal-rename" {
			return errFixtureInterrupted
		}
	} else if quarantinePresent {
		if err := syncRenameDirectories(root, sourceRelative, quarantineRelative); err != nil {
			return err
		}
	} else {
		if err := syncNearestExistingDirectory(root, filepath.Join(root, filepath.Dir(sourceRelative))); err != nil {
			return err
		}
		if err := syncNearestExistingDirectory(root, filepath.Join(root, filepath.Dir(quarantineRelative))); err != nil {
			return err
		}
	}

	if quarantinePresent {
		if err := verifyReceiptPackageSet(root, quarantineRelative, removal.Receipt, true); err != nil {
			return err
		}
		removedFiles := 0
		quarantineRoot := filepath.Join(root, quarantineRelative)
		for _, installed := range removal.Receipt.Packages {
			packageRoot := filepath.Join(quarantineRoot, installed.Digest)
			files := append([]receiptFile{}, installed.Files...)
			sort.Slice(files, func(i, j int) bool {
				firstDepth, secondDepth := strings.Count(files[i].Path, "/"), strings.Count(files[j].Path, "/")
				if firstDepth == secondDepth {
					return files[i].Path < files[j].Path
				}
				return firstDepth > secondDepth
			})
			for _, file := range files {
				exists, err := validateRemovalFile(packageRoot, file)
				if err != nil {
					return err
				}
				name := filepath.Join(packageRoot, filepath.FromSlash(file.Path))
				fileRelative, err := filepath.Rel(root, name)
				if err != nil {
					return err
				}
				if err := removeDurably(root, fileRelative, !exists); err != nil {
					return err
				}
				if exists {
					removedFiles++
					if fault == "after-first-owned-delete" && removedFiles == 1 {
						return errFixtureInterrupted
					}
				}
				if err := removeEmptyParents(filepath.Dir(name), packageRoot); err != nil {
					return err
				}
			}
			packageRelative, err := filepath.Rel(root, packageRoot)
			if err != nil {
				return err
			}
			if err := removeDurably(root, packageRelative, true); err != nil {
				return err
			}
		}
		if err := removeDurably(root, quarantineRelative, true); err != nil {
			return err
		}
		quarantinePresent = false
		if fault == "after-removal-tree" {
			return errFixtureInterrupted
		}
	}

	receipt, err = loadLifecycleReceipt(root, host, false)
	if err != nil {
		return err
	}
	if receipt.Schema != 0 {
		if !equalLifecycleReceipt(receipt, removal.Receipt) {
			return errors.New("lifecycle receipt changed before removal finalization")
		}
		if err := removeDurably(root, receiptRelative(host), false); err != nil {
			return err
		}
		if fault == "after-removal-receipt" {
			return errFixtureInterrupted
		}
	}
	for _, directory := range []string{
		filepath.Join(root, "packages"),
		filepath.Join(root, "removing"),
		filepath.Dir(receiptPath(root, host)),
	} {
		if err := removeEmptyParents(directory, root); err != nil {
			return err
		}
	}
	if err := removeDurably(root, pendingRemovalRelative(host), true); err != nil {
		return err
	}
	return nil
}

func removeEmptyParents(directory, stop string) error {
	for directory != stop && directory != "." && directory != string(filepath.Separator) {
		relative, err := filepath.Rel(stop, directory)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("empty-parent cleanup escapes its stop boundary")
		}
		if err := removeDurably(stop, relative, true); err != nil {
			info, statErr := os.Lstat(directory)
			if statErr == nil && info.IsDir() && info.Mode()&fs.ModeSymlink == 0 {
				entries, readErr := os.ReadDir(directory)
				if readErr == nil && len(entries) != 0 {
					return nil
				}
			}
			return err
		}
		directory = filepath.Dir(directory)
	}
	return nil
}

func syncLifecycleRecordParent(root, relative string) error {
	clean, err := cleanRelativePath(filepath.ToSlash(relative))
	if err != nil {
		return err
	}
	return syncNearestExistingDirectory(root, filepath.Join(root, filepath.Dir(clean)))
}

func validateBuiltPackage(built builtPackage) error {
	if built.Host != "codex" && built.Host != "claude" {
		return errors.New("fixture package host is invalid")
	}
	if !versionPattern.MatchString(built.Version) {
		return errors.New("fixture package version is invalid")
	}
	if !sha256Pattern.MatchString(built.ArchiveDigest) || sha256Hex(built.Archive) != built.ArchiveDigest {
		return errors.New("fixture package digest mismatch")
	}
	if !sha256Pattern.MatchString(built.SourceDigest) || !sha256Pattern.MatchString(built.CatalogDigest) ||
		sha256Hex(built.Catalog) != built.CatalogDigest {
		return errors.New("fixture package source or catalog digest mismatch")
	}
	if err := validateArchive(built.Archive, built.Entries); err != nil {
		return err
	}
	return validatePackageIdentity(built)
}

func validatePackageIdentity(built builtPackage) error {
	manifestPath, catalogPath, oppositeManifest, err := packageIdentityPaths(built.Host)
	if err != nil {
		return err
	}
	if built.CatalogPath != catalogPath {
		return errors.New("fixture package catalog path does not match its host")
	}
	manifestData, manifestCount := packageEntry(built.Entries, manifestPath)
	distributionData, distributionCount := packageEntry(built.Entries, "DISTRIBUTION.json")
	_, oppositeCount := packageEntry(built.Entries, oppositeManifest)
	if manifestCount != 1 || distributionCount != 1 || oppositeCount != 0 {
		return errors.New("fixture package host manifest or distribution metadata is missing, duplicate, or cross-host")
	}

	var metadata distributionMetadata
	if err := decodeStrict(distributionData, &metadata); err != nil {
		return fmt.Errorf("decode fixture distribution metadata: %w", err)
	}
	if metadata.Schema != 2 || metadata.Name != "level7-dev-loop" || metadata.Version != built.Version ||
		metadata.Channel != "stable" || metadata.Host != built.Host || metadata.ManifestPath != manifestPath ||
		metadata.CatalogPath != catalogPath || metadata.CatalogSHA256 != built.CatalogDigest ||
		metadata.SourceDigest != built.SourceDigest ||
		metadata.Builder != builderVersion || metadata.SupportClaim != "WITHHELD" || metadata.ActualHostGate != "SMOKE_TESTED" {
		return errors.New("fixture distribution metadata does not bind the package identity")
	}
	if err := validateCatalogIdentity(built, metadata.Name); err != nil {
		return err
	}

	if built.Host == "codex" {
		var manifest codexManifest
		if err := decodeStrict(manifestData, &manifest); err != nil {
			return fmt.Errorf("decode Codex fixture manifest: %w", err)
		}
		if manifest.Name != metadata.Name || manifest.Version != built.Version || manifest.Skills != "./skills/" {
			return errors.New("Codex fixture manifest does not bind the package identity")
		}
	} else {
		var manifest claudeManifest
		if err := decodeStrict(manifestData, &manifest); err != nil {
			return fmt.Errorf("decode Claude fixture manifest: %w", err)
		}
		if manifest.Schema != "https://json.schemastore.org/claude-code-plugin-manifest.json" || manifest.Name != metadata.Name ||
			manifest.Version != built.Version || manifest.DisplayName == "" {
			return errors.New("Claude fixture manifest does not bind the package identity")
		}
	}
	return validatePackageProjections(built, metadata)
}

func validateCatalogIdentity(built builtPackage, packageName string) error {
	if len(built.Catalog) == 0 || len(built.Catalog) > maximumFileSize {
		return errors.New("fixture package catalog is empty or unbounded")
	}
	expectedSource := "./plugins/" + packageName
	if built.Host == "codex" {
		var catalog codexCatalog
		if err := decodeStrict(built.Catalog, &catalog); err != nil {
			return fmt.Errorf("decode Codex fixture catalog: %w", err)
		}
		canonical, err := jsonBytes(catalog)
		if err != nil || !bytes.Equal(canonical, built.Catalog) {
			return errors.New("Codex fixture catalog is not canonical")
		}
		if catalog.Name != "level7-engineering" || catalog.Interface.DisplayName != "Level 7 Engineering" ||
			len(catalog.Plugins) != 1 || catalog.Plugins[0].Name != packageName || catalog.Plugins[0].Source.Source != "local" ||
			catalog.Plugins[0].Source.Path != expectedSource || catalog.Plugins[0].Policy.Installation != "AVAILABLE" ||
			catalog.Plugins[0].Policy.Authentication != "ON_INSTALL" || catalog.Plugins[0].Category != "Developer Tools" {
			return errors.New("Codex fixture catalog does not bind the package identity")
		}
		return nil
	}
	var catalog claudeCatalog
	if err := decodeStrict(built.Catalog, &catalog); err != nil {
		return fmt.Errorf("decode Claude fixture catalog: %w", err)
	}
	canonical, err := jsonBytes(catalog)
	if err != nil || !bytes.Equal(canonical, built.Catalog) {
		return errors.New("Claude fixture catalog is not canonical")
	}
	if catalog.Name != "level7-engineering" ||
		catalog.Description != "One-intent solo development: inspect, implement, test, repair, self-review, and hand off with optional team assurance." ||
		catalog.Owner.Name != "Level 7 Engineering" || len(catalog.Plugins) != 1 ||
		catalog.Plugins[0].Name != packageName || catalog.Plugins[0].Source != expectedSource ||
		catalog.Plugins[0].Description != "One-intent solo development: inspect, implement, test, repair, self-review, and hand off with optional team assurance." ||
		catalog.Plugins[0].Version != built.Version || catalog.Plugins[0].License != "MIT" ||
		catalog.Plugins[0].Category != "Development Tools" || !catalog.Plugins[0].Strict {
		return errors.New("Claude fixture catalog does not bind the package identity")
	}
	return nil
}

func validatePackageProjections(built builtPackage, metadata distributionMetadata) error {
	compatibilityData, compatibilityCount := packageEntry(built.Entries, "COMPATIBILITY.json")
	permissionData, permissionCount := packageEntry(built.Entries, "PERMISSIONS.json")
	provenanceData, provenanceCount := packageEntry(built.Entries, "PROVENANCE.input.json")
	sbomData, sbomCount := packageEntry(built.Entries, "SBOM.spdx.json")
	inventoryData, inventoryCount := packageEntry(built.Entries, "INVENTORY.json")
	if compatibilityCount != 1 || permissionCount != 1 || provenanceCount != 1 || sbomCount != 1 || inventoryCount != 1 {
		return errors.New("fixture package identity projections are missing or duplicate")
	}

	var compatibility compatibilityProjection
	if err := decodeStrict(compatibilityData, &compatibility); err != nil {
		return fmt.Errorf("decode fixture compatibility projection: %w", err)
	}
	expectedCompatibility := expectedCompatibilityEntries()
	expectedEntry := expectedCompatibility[0]
	if built.Host == "claude" {
		expectedEntry = expectedCompatibility[1]
	}
	if compatibility.Schema != 1 || compatibility.PackageVersion != built.Version ||
		compatibility.ArtifactSchema != "lean-risk-v1" || !equalCompatibilityEntry(compatibility.Entry, expectedEntry) {
		return errors.New("fixture compatibility projection does not bind the package identity")
	}

	var permission permissionsProjection
	if err := decodeStrict(permissionData, &permission); err != nil {
		return fmt.Errorf("decode fixture permissions projection: %w", err)
	}
	grants := permission.Permissions
	if permission.Schema != 1 || permission.PackageVersion != built.Version || permission.Host != built.Host ||
		permission.SupportClaim != "WITHHELD" || grants.Level7Network || grants.BundledExecutable || grants.MCPServer ||
		grants.Hook || grants.HostSetting || grants.Telemetry || strings.TrimSpace(grants.WorkspaceBoundary) == "" {
		return errors.New("fixture permissions projection does not bind the inert package identity")
	}

	var provenance provenanceInput
	if err := decodeStrict(provenanceData, &provenance); err != nil {
		return fmt.Errorf("decode fixture provenance input: %w", err)
	}
	if provenance.Schema != 1 || !provenance.Unsigned || provenance.Package != metadata.Name ||
		provenance.Version != built.Version || provenance.Host != built.Host || provenance.SourceDigest != metadata.SourceDigest ||
		provenance.Builder != metadata.Builder || provenance.Recipe != "offline deterministic standard-library package assembly" ||
		provenance.ExternalInputs == nil || len(provenance.ExternalInputs) != 0 ||
		provenance.Claim != "unsigned package input; authenticity is not established" {
		return errors.New("fixture provenance input does not bind the package identity")
	}

	var sbom sbomDocument
	if err := decodeStrict(sbomData, &sbom); err != nil {
		return fmt.Errorf("decode fixture SBOM: %w", err)
	}
	namespaceSuffix := "/sbom/" + built.Host + "/" + metadata.SourceDigest
	if sbom.SPDXVersion != "SPDX-2.3" || sbom.Name != metadata.Name+"-"+built.Host+"-"+built.Version ||
		!strings.HasSuffix(sbom.DocumentNamespace, namespaceSuffix) || len(sbom.Packages) != 1 ||
		sbom.Packages[0].Name != metadata.Name+"-"+built.Host || sbom.Packages[0].VersionInfo != built.Version ||
		len(sbom.Packages[0].Checksums) != 1 || sbom.Packages[0].Checksums[0].Algorithm != "SHA256" ||
		sbom.Packages[0].Checksums[0].ChecksumValue != metadata.SourceDigest {
		return errors.New("fixture SBOM does not bind the package identity")
	}

	var inventoryDocument inventory
	if err := decodeStrict(inventoryData, &inventoryDocument); err != nil {
		return fmt.Errorf("decode fixture inventory: %w", err)
	}
	if inventoryDocument.Schema != 1 || inventoryDocument.Scope != "all archive files except INVENTORY.json" || len(inventoryDocument.Files) != len(built.Entries)-1 {
		return errors.New("fixture inventory identity or size is invalid")
	}
	expected := make(map[string]archiveEntry, len(built.Entries)-1)
	for _, entry := range built.Entries {
		if entry.Name != "INVENTORY.json" {
			expected[entry.Name] = entry
		}
	}
	previous := ""
	for _, file := range inventoryDocument.Files {
		entry, ok := expected[file.Path]
		if !ok || (previous != "" && file.Path <= previous) || file.Mode != "0644" || file.Size != len(entry.Data) || file.SHA256 != sha256Hex(entry.Data) {
			return errors.New("fixture inventory does not exactly bind its archive entries")
		}
		previous = file.Path
		delete(expected, file.Path)
	}
	if len(expected) != 0 {
		return errors.New("fixture inventory omits an archive entry")
	}
	return nil
}

func packageIdentityPaths(host string) (string, string, string, error) {
	switch host {
	case "codex":
		return ".codex-plugin/plugin.json", ".agents/plugins/marketplace.json", ".claude-plugin/plugin.json", nil
	case "claude":
		return ".claude-plugin/plugin.json", ".claude-plugin/marketplace.json", ".codex-plugin/plugin.json", nil
	default:
		return "", "", "", errors.New("fixture package host is invalid")
	}
}

func packageEntry(entries []archiveEntry, name string) ([]byte, int) {
	var data []byte
	count := 0
	for _, entry := range entries {
		if entry.Name == name {
			data = entry.Data
			count++
		}
	}
	return data, count
}

func receiptPackageFromBuilt(built builtPackage) receiptPackage {
	files := make([]receiptFile, 0, len(built.Entries))
	for _, entry := range built.Entries {
		files = append(files, receiptFile{Path: entry.Name, Size: len(entry.Data), SHA256: sha256Hex(entry.Data)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return receiptPackage{Version: built.Version, Digest: built.ArchiveDigest, Files: files}
}

func verifyInstalledDirectory(lifecycleRoot, directory string, receipt receiptPackage) error {
	if err := validateReceiptPackage(receipt); err != nil {
		return err
	}
	relativeDirectory, err := filepath.Rel(lifecycleRoot, directory)
	if err != nil || relativeDirectory == ".." || strings.HasPrefix(relativeDirectory, ".."+string(filepath.Separator)) {
		return errors.New("installed package directory escapes its lifecycle root")
	}
	if err := ensurePhysicalDirectory(lifecycleRoot, relativeDirectory, false); err != nil {
		return errors.New("installed package directory is missing or unsafe")
	}
	expected := make(map[string]receiptFile, len(receipt.Files))
	for _, file := range receipt.Files {
		expected[file.Path] = file
	}
	expectedDirectories := ownedDirectorySet(receipt.Files)
	seen := make(map[string]bool, len(expected))
	archiveEntries := make([]archiveEntry, 0, len(expected))
	err = filepath.WalkDir(directory, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == directory {
			return nil
		}
		relative, err := filepath.Rel(directory, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("installed package contains symlink %q", relative)
		}
		if entry.IsDir() {
			if !expectedDirectories[relative] {
				return fmt.Errorf("installed package contains unowned directory %q", relative)
			}
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
		archiveEntries = append(archiveEntries, archiveEntry{Name: relative, Data: data, Mode: 0o644})
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return errors.New("installed package is missing an owned file")
	}
	reconstructed, err := createArchive(archiveEntries)
	if err != nil || sha256Hex(reconstructed) != receipt.Digest {
		return errors.New("installed package bytes do not match the receipt digest")
	}
	return nil
}

func verifyReceiptPackageSet(lifecycleRoot, hostRelative string, receipt lifecycleReceipt, allowMissing bool) error {
	if err := validateLifecycleReceipt(receipt, receipt.Host); err != nil {
		return err
	}
	present, err := physicalDirectoryState(lifecycleRoot, hostRelative)
	if err != nil {
		return err
	}
	if !present {
		if allowMissing {
			return nil
		}
		return errors.New("receipt package set is missing")
	}
	hostRoot := filepath.Join(lifecycleRoot, hostRelative)
	expected := make(map[string]receiptPackage, len(receipt.Packages))
	for _, installed := range receipt.Packages {
		expected[installed.Digest] = installed
	}
	entries, err := os.ReadDir(hostRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, ok := expected[entry.Name()]; !ok || !entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("package set contains unowned entry %q", entry.Name())
		}
	}
	for _, installed := range receipt.Packages {
		packageRoot := filepath.Join(hostRoot, installed.Digest)
		if allowMissing {
			if err := verifyRemovalRemainder(lifecycleRoot, packageRoot, installed); err != nil {
				return err
			}
			continue
		}
		if err := verifyInstalledDirectory(lifecycleRoot, packageRoot, installed); err != nil {
			return err
		}
	}
	return nil
}

func verifyRemovalRemainder(lifecycleRoot, directory string, receipt receiptPackage) error {
	if err := validateReceiptPackage(receipt); err != nil {
		return err
	}
	relativeDirectory, err := filepath.Rel(lifecycleRoot, directory)
	if err != nil || relativeDirectory == ".." || strings.HasPrefix(relativeDirectory, ".."+string(filepath.Separator)) {
		return errors.New("removal package directory escapes its lifecycle root")
	}
	present, err := physicalDirectoryState(lifecycleRoot, relativeDirectory)
	if err != nil || !present {
		return err
	}
	expected := make(map[string]receiptFile, len(receipt.Files))
	for _, file := range receipt.Files {
		expected[file.Path] = file
	}
	expectedDirectories := ownedDirectorySet(receipt.Files)
	return filepath.WalkDir(directory, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == directory {
			return nil
		}
		relative, err := filepath.Rel(directory, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("removal package contains symlink %q", relative)
		}
		if entry.IsDir() {
			if !expectedDirectories[relative] {
				return fmt.Errorf("removal package contains unowned directory %q", relative)
			}
			return nil
		}
		file, ok := expected[relative]
		if !ok {
			return fmt.Errorf("removal package contains unowned file %q", relative)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o644 || info.Size() != int64(file.Size) {
			return fmt.Errorf("removal file metadata changed: %q", relative)
		}
		data, err := os.ReadFile(name)
		if err != nil || sha256Hex(data) != file.SHA256 {
			return fmt.Errorf("removal file content changed: %q", relative)
		}
		return nil
	})
}

func verifyPendingStageRemainder(lifecycleRoot, directory string, receipt receiptPackage) error {
	if err := validateReceiptPackage(receipt); err != nil {
		return err
	}
	relativeDirectory, err := filepath.Rel(lifecycleRoot, directory)
	if err != nil || relativeDirectory == ".." || strings.HasPrefix(relativeDirectory, ".."+string(filepath.Separator)) {
		return errors.New("pending stage directory escapes its lifecycle root")
	}
	present, err := physicalDirectoryState(lifecycleRoot, relativeDirectory)
	if err != nil || !present {
		return err
	}
	expected := make(map[string]receiptFile, len(receipt.Files))
	for _, file := range receipt.Files {
		expected[file.Path] = file
	}
	expectedDirectories := ownedDirectorySet(receipt.Files)
	return filepath.WalkDir(directory, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == directory {
			return nil
		}
		relative, err := filepath.Rel(directory, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("pending stage contains symlink %q", relative)
		}
		if entry.IsDir() {
			if !expectedDirectories[relative] {
				return fmt.Errorf("pending stage contains unowned directory %q", relative)
			}
			return nil
		}
		file, ok := expected[relative]
		if !ok {
			return fmt.Errorf("pending stage contains unowned file %q", relative)
		}
		info, err := entry.Info()
		if err != nil || !validPendingStageFileInfo(info, file) {
			return fmt.Errorf("pending stage file metadata is unsafe: %q", relative)
		}
		return nil
	})
}

func validatePendingStageFile(stage string, file receiptFile) (bool, error) {
	parent := filepath.Dir(filepath.FromSlash(file.Path))
	if err := ensurePhysicalDirectory(stage, parent, false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	name := filepath.Join(stage, filepath.FromSlash(file.Path))
	info, err := os.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !validPendingStageFileInfo(info, file) {
		return false, fmt.Errorf("pending stage file metadata is unsafe: %q", file.Path)
	}
	return true, nil
}

func validPendingStageFileInfo(info fs.FileInfo, file receiptFile) bool {
	mode := info.Mode()
	return mode.IsRegular() && mode&fs.ModeSymlink == 0 &&
		(mode.Perm() == 0o600 || mode.Perm() == 0o644) &&
		info.Size() >= 0 && info.Size() <= int64(file.Size)
}

func validateRemovalFile(packageRoot string, file receiptFile) (bool, error) {
	parent := filepath.Dir(filepath.FromSlash(file.Path))
	if err := ensurePhysicalDirectory(packageRoot, parent, false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	name := filepath.Join(packageRoot, filepath.FromSlash(file.Path))
	info, err := os.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 || info.Mode().Perm() != 0o644 || info.Size() != int64(file.Size) {
		return false, fmt.Errorf("removal file metadata changed: %q", file.Path)
	}
	data, err := os.ReadFile(name)
	if err != nil || sha256Hex(data) != file.SHA256 {
		return false, fmt.Errorf("removal file content changed: %q", file.Path)
	}
	return true, nil
}

func ownedDirectorySet(files []receiptFile) map[string]bool {
	directories := make(map[string]bool)
	for _, file := range files {
		for directory := path.Dir(file.Path); directory != "."; directory = path.Dir(directory) {
			directories[directory] = true
		}
	}
	return directories
}

func physicalDirectoryState(root, relative string) (bool, error) {
	clean, err := cleanRelativePath(filepath.ToSlash(relative))
	if err != nil {
		return false, err
	}
	if err := ensurePhysicalDirectory(root, filepath.Dir(clean), false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	info, err := os.Lstat(filepath.Join(root, clean))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return false, errors.New("lifecycle transaction path is not a physical directory")
	}
	if err := ensurePhysicalDirectory(root, clean, false); err != nil {
		return false, err
	}
	return true, nil
}

func loadLifecycleReceipt(root, host string, required bool) (lifecycleReceipt, error) {
	if host != "codex" && host != "claude" {
		return lifecycleReceipt{}, errors.New("invalid lifecycle host")
	}
	data, err := readRegularBounded(root, receiptRelative(host))
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
	if err := validateLifecycleReceipt(receipt, host); err != nil {
		return lifecycleReceipt{}, err
	}
	return receipt, nil
}

func loadPending(root, host string) (pendingInstall, error) {
	if host != "codex" && host != "claude" {
		return pendingInstall{}, errors.New("invalid pending lifecycle host")
	}
	data, err := readRegularBounded(root, pendingRelative(host))
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
	if err := validatePendingInstall(pending, host); err != nil {
		return pendingInstall{}, err
	}
	return pending, nil
}

func validatePendingInstall(pending pendingInstall, host string) error {
	if err := validatePendingInstallBase(pending, host); err != nil {
		return err
	}
	if !sha256Pattern.MatchString(pending.CommittedReceiptDigest) || pending.CommittedReceiptDigest == pending.BaseReceiptDigest {
		return errors.New("pending lifecycle committed receipt binding is invalid")
	}
	return nil
}

func validatePendingInstallBase(pending pendingInstall, host string) error {
	if pending.Schema != 1 || pending.Host != host ||
		(pending.PreviousDigest != "" && !sha256Pattern.MatchString(pending.PreviousDigest)) ||
		pending.PreviousDigest == pending.Package.Digest || !sha256Pattern.MatchString(pending.BaseReceiptDigest) {
		return errors.New("pending lifecycle record identity is invalid")
	}
	if err := validateReceiptPackage(pending.Package); err != nil {
		return err
	}
	return nil
}

func lifecycleReceiptStateDigest(receipt lifecycleReceipt) (string, error) {
	if receipt.Schema == 0 {
		if receipt.Host != "" || receipt.ActiveDigest != "" || receipt.PreviousDigest != "" || len(receipt.Packages) != 0 {
			return "", errors.New("absent lifecycle receipt contains state")
		}
		return sha256Hex([]byte("level7-lifecycle-receipt-absent-v1\n")), nil
	}
	if err := validateLifecycleReceipt(receipt, receipt.Host); err != nil {
		return "", err
	}
	data, err := lifecycleJSONBytes(receipt)
	if err != nil {
		return "", err
	}
	return sha256Hex(data), nil
}

func loadPendingRemoval(root, host string) (pendingRemoval, error) {
	if host != "codex" && host != "claude" {
		return pendingRemoval{}, errors.New("invalid pending removal host")
	}
	data, err := readRegularBounded(root, pendingRemovalRelative(host))
	if err != nil {
		return pendingRemoval{}, err
	}
	if len(data) == 0 || len(data) > maximumFileSize {
		return pendingRemoval{}, errors.New("pending removal record is empty or oversized")
	}
	var removal pendingRemoval
	if err := decodeStrict(data, &removal); err != nil {
		return pendingRemoval{}, err
	}
	if err := validatePendingRemoval(removal, host); err != nil {
		return pendingRemoval{}, err
	}
	return removal, nil
}

func validatePendingRemoval(removal pendingRemoval, host string) error {
	if removal.Schema != 1 || removal.Host != host {
		return errors.New("pending removal record identity is invalid")
	}
	if err := validateLifecycleReceipt(removal.Receipt, host); err != nil {
		return fmt.Errorf("pending removal receipt: %w", err)
	}
	return nil
}

func validateLifecycleReceipt(receipt lifecycleReceipt, host string) error {
	if receipt.Schema != 1 || receipt.Host != host || !sha256Pattern.MatchString(receipt.ActiveDigest) ||
		len(receipt.Packages) == 0 || len(receipt.Packages) > maximumArchiveFiles ||
		(receipt.PreviousDigest != "" && (!sha256Pattern.MatchString(receipt.PreviousDigest) || receipt.PreviousDigest == receipt.ActiveDigest)) {
		return errors.New("lifecycle receipt identity is invalid")
	}
	previous := ""
	for _, installed := range receipt.Packages {
		if err := validateReceiptPackage(installed); err != nil {
			return err
		}
		if previous != "" && installed.Digest <= previous {
			return errors.New("lifecycle receipt packages are duplicate or unsorted")
		}
		previous = installed.Digest
	}
	if _, ok := findReceiptPackage(receipt, receipt.ActiveDigest); !ok {
		return errors.New("active package is absent from lifecycle receipt")
	}
	if receipt.PreviousDigest != "" {
		if _, ok := findReceiptPackage(receipt, receipt.PreviousDigest); !ok {
			return errors.New("previous package is absent from lifecycle receipt")
		}
	}
	return nil
}

func validateReceiptPackage(receipt receiptPackage) error {
	if !versionPattern.MatchString(receipt.Version) || !sha256Pattern.MatchString(receipt.Digest) || len(receipt.Files) == 0 || len(receipt.Files) > maximumArchiveFiles {
		return errors.New("lifecycle package receipt identity is invalid")
	}
	previous := ""
	for _, file := range receipt.Files {
		if err := validateArchiveName(file.Path); err != nil || file.Size < 1 || file.Size > maximumFileSize || !sha256Pattern.MatchString(file.SHA256) {
			return errors.New("installed package receipt contains an invalid file")
		}
		if previous != "" && file.Path <= previous {
			return errors.New("installed package receipt paths are duplicate or unsorted")
		}
		previous = file.Path
	}
	return nil
}

func equalReceiptPackage(first, second receiptPackage) bool {
	if first.Version != second.Version || first.Digest != second.Digest || len(first.Files) != len(second.Files) {
		return false
	}
	for index := range first.Files {
		if first.Files[index] != second.Files[index] {
			return false
		}
	}
	return true
}

func equalLifecycleReceipt(first, second lifecycleReceipt) bool {
	if first.Schema != second.Schema || first.Host != second.Host || first.ActiveDigest != second.ActiveDigest ||
		first.PreviousDigest != second.PreviousDigest || len(first.Packages) != len(second.Packages) {
		return false
	}
	for index := range first.Packages {
		if !equalReceiptPackage(first.Packages[index], second.Packages[index]) {
			return false
		}
	}
	return true
}

func writeLifecycleJSON(root, relative string, value any) error {
	data, err := lifecycleJSONBytes(value)
	if err != nil {
		return err
	}
	return writeDurableRegular(root, relative, data)
}

func validateLifecycleJSONBound(value any) error {
	_, err := lifecycleJSONBytes(value)
	return err
}

func lifecycleJSONBytes(value any) ([]byte, error) {
	data, err := jsonBytes(value)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maximumFileSize {
		return nil, errors.New("lifecycle JSON record is empty or oversized")
	}
	return data, nil
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

func pendingRemovalRelative(host string) string {
	return filepath.ToSlash(filepath.Join("pending-removals", host+".json"))
}

func pendingStageRelative(host, digest string) string {
	return filepath.ToSlash(filepath.Join("packages", host, ".stage-"+digest))
}

func receiptPath(root, host string) string {
	return filepath.Join(root, filepath.FromSlash(receiptRelative(host)))
}

func pendingPath(root, host string) string {
	return filepath.Join(root, filepath.FromSlash(pendingRelative(host)))
}

func pendingRemovalPath(root, host string) string {
	return filepath.Join(root, filepath.FromSlash(pendingRemovalRelative(host)))
}

func pendingStageDirectory(root, host, digest string) string {
	return filepath.Join(root, filepath.FromSlash(pendingStageRelative(host, digest)))
}

func packageDirectory(root, host, digest string) string {
	return filepath.Join(root, "packages", host, digest)
}
