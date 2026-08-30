package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestQualifyLifecycleSetPreflightsBothHostsBeforeMutation(t *testing.T) {
	packages := []builtPackage{
		testBuiltPackage(t, "codex", "0.1.0", "codex\n"),
		testBuiltPackage(t, "claude", "0.1.0", "claude\n"),
	}
	packages[1].CatalogDigest = strings.Repeat("0", 64)
	syncs := 0
	replaceLifecycleSyncDirectory(t, func(string) error {
		syncs++
		return nil
	})
	if qualification, err := qualifyLifecycleSet(packages); err == nil || len(qualification.Packages) != 0 {
		t.Fatalf("invalid second host passed preflight: qualification=%+v error=%v", qualification, err)
	}
	if syncs != 0 {
		t.Fatalf("preflight failure attempted %d lifecycle durability writes", syncs)
	}
}

func TestQualifyLifecycleSetRejectsMixedVersionsBeforeMutation(t *testing.T) {
	packages := []builtPackage{
		testBuiltPackage(t, "codex", "0.1.0", "codex\n"),
		testBuiltPackage(t, "claude", "0.1.1", "claude\n"),
	}
	syncs := 0
	replaceLifecycleSyncDirectory(t, func(string) error {
		syncs++
		return nil
	})
	if qualification, err := qualifyLifecycleSet(packages); err == nil || len(qualification.Packages) != 0 {
		t.Fatalf("mixed-version package set passed preflight: qualification=%+v error=%v", qualification, err)
	}
	if syncs != 0 {
		t.Fatalf("mixed-version preflight failure attempted %d lifecycle durability writes", syncs)
	}
}

func TestFixtureLifecycleInstallUpgradeRecoverRollbackRemove(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")

	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	sameArchiveDifferentVersion := base
	sameArchiveDifferentVersion.Version = "0.1.999"
	if err := installFixture(root, sameArchiveDifferentVersion, ""); err == nil {
		t.Fatal("same archive with a different version passed reinstall receipt binding")
	}
	if err := installFixture(root, base, ""); err != nil {
		t.Fatalf("idempotent reinstall: %v", err)
	}
	if err := installFixture(root, upgrade, "after-publish"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	receipt, err := loadLifecycleReceipt(root, "codex", true)
	if err != nil || receipt.ActiveDigest != base.ArchiveDigest {
		t.Fatalf("interruption changed active receipt=%+v error=%v", receipt, err)
	}
	if err := recoverFixture(root, "codex"); err != nil {
		t.Fatal(err)
	}
	receipt, err = loadLifecycleReceipt(root, "codex", true)
	if err != nil || receipt.ActiveDigest != upgrade.ArchiveDigest || receipt.PreviousDigest != base.ArchiveDigest || len(receipt.Packages) != 2 {
		t.Fatalf("recovered receipt=%+v error=%v", receipt, err)
	}
	if err := rollbackFixture(root, "codex"); err != nil {
		t.Fatal(err)
	}
	receipt, err = loadLifecycleReceipt(root, "codex", true)
	if err != nil || receipt.ActiveDigest != base.ArchiveDigest || receipt.PreviousDigest != upgrade.ArchiveDigest {
		t.Fatalf("rollback receipt=%+v error=%v", receipt, err)
	}
	preview, err := prepareRemoval(root, "codex")
	if err != nil || len(preview.Conflicts) != 0 || len(preview.OwnedPackages) != 2 {
		t.Fatalf("preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(root, "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(receiptPath(root, "codex")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("receipt remains: %v", err)
	}
}

func TestFixtureLifecycleBeforePublishRecoveryPreservesActive(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "claude", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, "before-publish"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	stage := pendingStageDirectory(root, upgrade.Host, upgrade.ArchiveDigest)
	if _, err := os.Stat(stage); err != nil {
		t.Fatalf("journal-bound stage was not preserved: %v", err)
	}
	partial := filepath.Join(stage, filepath.FromSlash(upgrade.Entries[0].Name))
	if err := os.WriteFile(partial, upgrade.Entries[0].Data[:1], 0o600); err != nil {
		t.Fatalf("simulate partial staged write: %v", err)
	}
	if err := os.Chmod(partial, 0o600); err != nil {
		t.Fatalf("mark partial staged write: %v", err)
	}
	if len(upgrade.Entries) > 1 {
		if err := os.Remove(filepath.Join(stage, filepath.FromSlash(upgrade.Entries[1].Name))); err != nil {
			t.Fatalf("simulate missing staged file: %v", err)
		}
	}
	if err := filepath.WalkDir(stage, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if strings.HasPrefix(entry.Name(), ".write-") {
			t.Fatalf("stage population used an unjournaled temporary path: %s", name)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, "claude"); err != nil {
		t.Fatal(err)
	}
	receipt, err := loadLifecycleReceipt(root, "claude", true)
	if err != nil || receipt.ActiveDigest != base.ArchiveDigest || len(receipt.Packages) != 1 {
		t.Fatalf("receipt=%+v error=%v", receipt, err)
	}
	if _, err := os.Stat(packageDirectory(root, "claude", upgrade.ArchiveDigest)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unpublished package remains: %v", err)
	}
	if _, err := os.Stat(stage); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("recovered lifecycle stage remains: %v", err)
	}
}

func TestFixtureLifecycleStageRecoveryPreservesUnownedConflict(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, "before-publish"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	conflict := filepath.Join(pendingStageDirectory(root, upgrade.Host, upgrade.ArchiveDigest), "unowned.txt")
	if err := os.WriteFile(conflict, []byte("preserve\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, upgrade.Host); err == nil {
		t.Fatal("stage recovery accepted an unowned conflict")
	}
	if data, err := os.ReadFile(conflict); err != nil || string(data) != "preserve\n" {
		t.Fatalf("blocked stage recovery changed conflict: data=%q error=%v", data, err)
	}
	if _, err := loadPending(root, upgrade.Host); err != nil {
		t.Fatalf("blocked stage recovery removed its journal: %v", err)
	}
	if err := os.Remove(conflict); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, upgrade.Host); err != nil {
		t.Fatalf("stage recovery after conflict removal: %v", err)
	}
	receipt, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil || receipt.ActiveDigest != base.ArchiveDigest {
		t.Fatalf("stage recovery changed active package: receipt=%+v error=%v", receipt, err)
	}
}

func TestFixtureLifecycleRecoversAfterReceiptPublication(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, "after-receipt"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	before, err := loadLifecycleReceipt(root, upgrade.Host, true)
	if err != nil || before.ActiveDigest != upgrade.ArchiveDigest || before.PreviousDigest != base.ArchiveDigest {
		t.Fatalf("published receipt=%+v error=%v", before, err)
	}
	if _, err := loadPending(root, upgrade.Host); err != nil {
		t.Fatalf("pending install was not preserved: %v", err)
	}
	if err := recoverFixture(root, upgrade.Host); err != nil {
		t.Fatal(err)
	}
	after, err := loadLifecycleReceipt(root, upgrade.Host, true)
	if err != nil || !equalLifecycleReceipt(before, after) {
		t.Fatalf("recovery changed committed receipt: before=%+v after=%+v error=%v", before, after, err)
	}
	if _, err := loadPending(root, upgrade.Host); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("pending install remains after recovery: %v", err)
	}
	if err := installFixture(root, upgrade, ""); err != nil {
		t.Fatalf("reinstall after recovery: %v", err)
	}
}

func TestFixtureLifecycleCommittedReceiptWithMissingPackageFailsClosed(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, "after-receipt"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	committed, err := loadLifecycleReceipt(root, upgrade.Host, true)
	if err != nil || committed.ActiveDigest != upgrade.ArchiveDigest {
		t.Fatalf("committed receipt=%+v error=%v", committed, err)
	}
	if err := os.RemoveAll(packageDirectory(root, upgrade.Host, upgrade.ArchiveDigest)); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, upgrade.Host); err == nil {
		t.Fatal("recovery abandoned an install committed by its receipt")
	}
	if _, err := loadPending(root, upgrade.Host); err != nil {
		t.Fatalf("failed recovery removed its install journal: %v", err)
	}
	after, err := loadLifecycleReceipt(root, upgrade.Host, true)
	if err != nil || !equalLifecycleReceipt(committed, after) {
		t.Fatalf("failed recovery changed committed receipt: before=%+v after=%+v error=%v", committed, after, err)
	}
	if err := verifyInstalledDirectory(root, packageDirectory(root, base.Host, base.ArchiveDigest), receiptPackageFromBuilt(base)); err != nil {
		t.Fatalf("failed recovery changed base package: %v", err)
	}
}

func TestFixtureLifecycleRejectsChangedCommittedReceiptBinding(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	before, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, "after-publish"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	pending, err := loadPending(root, upgrade.Host)
	if err != nil {
		t.Fatal(err)
	}
	pending.CommittedReceiptDigest = sha256Hex([]byte("substituted committed receipt binding\n"))
	if pending.CommittedReceiptDigest == pending.BaseReceiptDigest {
		t.Fatal("test committed receipt digest unexpectedly matches the base")
	}
	if err := writeLifecycleJSON(root, pendingRelative(upgrade.Host), pending); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, upgrade.Host); err == nil {
		t.Fatal("recovery accepted a changed committed receipt binding")
	}
	after, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil || !equalLifecycleReceipt(before, after) {
		t.Fatalf("blocked completion changed receipt: before=%+v after=%+v error=%v", before, after, err)
	}
	if _, err := loadPending(root, upgrade.Host); err != nil {
		t.Fatalf("blocked completion removed its install journal: %v", err)
	}
	for _, built := range []builtPackage{base, upgrade} {
		if err := verifyInstalledDirectory(root, packageDirectory(root, built.Host, built.ArchiveDigest), receiptPackageFromBuilt(built)); err != nil {
			t.Fatalf("blocked completion changed %s package: %v", built.Version, err)
		}
	}
}

func TestFixtureLifecycleRevalidatesProspectivePackageSet(t *testing.T) {
	for _, mode := range []string{"changed-base-file", "missing-base-package"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			base := testBuiltPackage(t, "claude", "0.1.0", "base\n")
			upgrade := testBuiltPackage(t, "claude", "0.1.1", "upgrade\n")
			if err := installFixture(root, base, ""); err != nil {
				t.Fatal(err)
			}
			before, err := loadLifecycleReceipt(root, base.Host, true)
			if err != nil {
				t.Fatal(err)
			}
			if err := installFixture(root, upgrade, "after-publish"); !errors.Is(err, errFixtureInterrupted) {
				t.Fatalf("interruption error=%v", err)
			}
			baseDirectory := packageDirectory(root, base.Host, base.ArchiveDigest)
			changedPath := filepath.Join(baseDirectory, filepath.FromSlash(base.Entries[0].Name))
			switch mode {
			case "changed-base-file":
				if err := os.WriteFile(changedPath, []byte("changed\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "missing-base-package":
				if err := os.RemoveAll(baseDirectory); err != nil {
					t.Fatal(err)
				}
			}
			if err := recoverFixture(root, upgrade.Host); err == nil {
				t.Fatal("recovery committed without its full prospective package set")
			}
			after, err := loadLifecycleReceipt(root, base.Host, true)
			if err != nil || !equalLifecycleReceipt(before, after) {
				t.Fatalf("blocked completion changed receipt: before=%+v after=%+v error=%v", before, after, err)
			}
			if _, err := loadPending(root, upgrade.Host); err != nil {
				t.Fatalf("blocked completion removed its install journal: %v", err)
			}
			if err := verifyInstalledDirectory(root, packageDirectory(root, upgrade.Host, upgrade.ArchiveDigest), receiptPackageFromBuilt(upgrade)); err != nil {
				t.Fatalf("blocked completion changed pending package: %v", err)
			}
			if mode == "changed-base-file" {
				if data, err := os.ReadFile(changedPath); err != nil || string(data) != "changed\n" {
					t.Fatalf("blocked completion changed conflicting base bytes: data=%q error=%v", data, err)
				}
			} else if _, err := os.Stat(baseDirectory); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("blocked completion recreated missing base package: %v", err)
			}
		})
	}
}

func TestFixtureLifecycleAbandonmentRequiresExactBaseReceipt(t *testing.T) {
	for _, mode := range []string{"missing", "same-active-mismatch"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			base := testBuiltPackage(t, "claude", "0.0.9", "base\n")
			active := testBuiltPackage(t, "claude", "0.1.0", "active\n")
			pendingPackage := testBuiltPackage(t, "claude", "0.1.1", "pending\n")
			if err := installFixture(root, base, ""); err != nil {
				t.Fatal(err)
			}
			if err := installFixture(root, active, ""); err != nil {
				t.Fatal(err)
			}
			if err := installFixture(root, pendingPackage, "before-publish"); !errors.Is(err, errFixtureInterrupted) {
				t.Fatalf("interruption error=%v", err)
			}
			stage := pendingStageDirectory(root, pendingPackage.Host, pendingPackage.ArchiveDigest)
			switch mode {
			case "missing":
				if err := os.Remove(receiptPath(root, pendingPackage.Host)); err != nil {
					t.Fatal(err)
				}
			case "same-active-mismatch":
				receipt, err := loadLifecycleReceipt(root, pendingPackage.Host, true)
				if err != nil {
					t.Fatal(err)
				}
				receipt.PreviousDigest = ""
				if err := writeLifecycleJSON(root, receiptRelative(pendingPackage.Host), receipt); err != nil {
					t.Fatal(err)
				}
			}
			if err := recoverFixture(root, pendingPackage.Host); err == nil {
				t.Fatal("recovery abandoned an install against a changed base receipt")
			}
			if _, err := loadPending(root, pendingPackage.Host); err != nil {
				t.Fatalf("failed recovery removed its install journal: %v", err)
			}
			if _, err := os.Stat(stage); err != nil {
				t.Fatalf("failed recovery changed its diagnostic stage: %v", err)
			}
		})
	}
}

func TestFixtureLifecycleConflictsFailClosed(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	packageRoot := packageDirectory(root, built.Host, built.ArchiveDigest)
	unowned := filepath.Join(packageRoot, "unowned.txt")
	if err := os.WriteFile(unowned, []byte("unowned\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	preview, err := prepareRemoval(root, built.Host)
	if err != nil || len(preview.Conflicts) == 0 {
		t.Fatalf("preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(root, built.Host); err == nil {
		t.Fatal("unowned-file removal passed")
	}
	if _, err := os.Stat(unowned); err != nil {
		t.Fatalf("unowned file changed: %v", err)
	}
	if err := os.Remove(unowned); err != nil {
		t.Fatal(err)
	}

	owned := filepath.Join(packageRoot, filepath.FromSlash(built.Entries[0].Name))
	if err := os.WriteFile(owned, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, built, ""); err == nil {
		t.Fatal("changed-owned reinstall passed")
	}
	if err := removeFixture(root, built.Host); err == nil {
		t.Fatal("changed-owned removal passed")
	}
}

func TestFixtureLifecycleUnownedEmptyDirectoryBlocksBeforeMutation(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	packageRoot := packageDirectory(root, built.Host, built.ArchiveDigest)
	empty := filepath.Join(packageRoot, "unowned-empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	receiptBefore, err := os.ReadFile(receiptPath(root, built.Host))
	if err != nil {
		t.Fatal(err)
	}
	ownedPath := filepath.Join(packageRoot, filepath.FromSlash(built.Entries[0].Name))
	ownedBefore, err := os.ReadFile(ownedPath)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := prepareRemoval(root, built.Host)
	if err != nil || len(preview.Conflicts) == 0 {
		t.Fatalf("preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(root, built.Host); err == nil {
		t.Fatal("unowned empty directory passed removal")
	}
	if _, err := os.Stat(empty); err != nil {
		t.Fatalf("blocked removal changed unowned directory: %v", err)
	}
	if receiptAfter, err := os.ReadFile(receiptPath(root, built.Host)); err != nil || string(receiptAfter) != string(receiptBefore) {
		t.Fatalf("blocked removal changed receipt: error=%v", err)
	}
	if ownedAfter, err := os.ReadFile(ownedPath); err != nil || string(ownedAfter) != string(ownedBefore) {
		t.Fatalf("blocked removal changed owned bytes: error=%v", err)
	}
	if _, err := loadPendingRemoval(root, built.Host); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("blocked removal wrote a journal: %v", err)
	}
}

func TestFixtureLifecycleInterruptedRemovalResumes(t *testing.T) {
	for _, fault := range []string{
		"after-removal-journal",
		"after-removal-rename",
		"after-first-owned-delete",
		"after-removal-tree",
		"after-removal-receipt",
	} {
		t.Run(fault, func(t *testing.T) {
			root := t.TempDir()
			base := testBuiltPackage(t, "claude", "0.1.0", "base\n")
			upgrade := testBuiltPackage(t, "claude", "0.1.1", "upgrade\n")
			if err := os.WriteFile(filepath.Join(root, "unowned.txt"), []byte("preserve\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := installFixture(root, base, ""); err != nil {
				t.Fatal(err)
			}
			if err := installFixture(root, upgrade, ""); err != nil {
				t.Fatal(err)
			}
			if err := removeFixtureWithFault(root, base.Host, fault); !errors.Is(err, errFixtureInterrupted) {
				t.Fatalf("fault=%s error=%v", fault, err)
			}
			if err := installFixture(root, upgrade, ""); err == nil {
				t.Fatal("install passed during pending removal")
			}
			if err := rollbackFixture(root, base.Host); err == nil {
				t.Fatal("rollback passed during pending removal")
			}
			if err := recoverFixture(root, base.Host); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{
				receiptPath(root, base.Host),
				pendingRemovalPath(root, base.Host),
				filepath.Join(root, "packages", base.Host),
				filepath.Join(root, "removing", base.Host),
			} {
				if _, err := os.Lstat(name); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("removal state remains at %s: %v", name, err)
				}
			}
			if data, err := os.ReadFile(filepath.Join(root, "unowned.txt")); err != nil || string(data) != "preserve\n" {
				t.Fatalf("unowned state changed: data=%q error=%v", data, err)
			}
		})
	}
}

func TestFixtureLifecycleRemovalRecoveryRejectsQuarantineConflict(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	if err := removeFixtureWithFault(root, built.Host, "after-removal-rename"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	conflict := filepath.Join(root, "removing", built.Host, built.ArchiveDigest, "unowned.txt")
	if err := os.WriteFile(conflict, []byte("preserve\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err == nil {
		t.Fatal("removal recovery accepted a quarantine conflict")
	}
	if data, err := os.ReadFile(conflict); err != nil || string(data) != "preserve\n" {
		t.Fatalf("blocked recovery changed conflict: data=%q error=%v", data, err)
	}
	if _, err := os.Stat(receiptPath(root, built.Host)); err != nil {
		t.Fatalf("blocked recovery removed receipt: %v", err)
	}
	if _, err := os.Stat(pendingRemovalPath(root, built.Host)); err != nil {
		t.Fatalf("blocked recovery removed journal: %v", err)
	}
	if err := os.Remove(conflict); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err != nil {
		t.Fatalf("recovery after conflict removal: %v", err)
	}
}

func TestFixtureLifecycleRemovalRecoveryRejectsMissingLeafUnderSymlink(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	if err := removeFixtureWithFault(root, built.Host, "after-removal-journal"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	relocated := filepath.Join(t.TempDir(), "packages")
	if err := os.Rename(filepath.Join(root, "packages"), relocated); err != nil {
		t.Fatal(err)
	}
	emptyTarget := t.TempDir()
	symlink := filepath.Join(root, "packages")
	if err := os.Symlink(emptyTarget, symlink); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err == nil {
		t.Fatal("removal recovery accepted a missing leaf under a symlinked parent")
	}
	if info, err := os.Lstat(symlink); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("blocked recovery changed symlink: info=%v error=%v", info, err)
	}
	if _, err := os.Stat(filepath.Join(relocated, built.Host, built.ArchiveDigest)); err != nil {
		t.Fatalf("blocked recovery changed relocated package: %v", err)
	}
	if _, err := os.Stat(receiptPath(root, built.Host)); err != nil {
		t.Fatalf("blocked recovery removed receipt: %v", err)
	}
	if _, err := os.Stat(pendingRemovalPath(root, built.Host)); err != nil {
		t.Fatalf("blocked recovery removed journal: %v", err)
	}
	if err := os.Remove(symlink); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(relocated, filepath.Join(root, "packages")); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err != nil {
		t.Fatalf("recovery after restoring physical package root: %v", err)
	}
}

func TestFixtureLifecycleRemovalRecoveryRejectsConflictingTrees(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	if err := removeFixtureWithFault(root, built.Host, "after-removal-journal"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	quarantine := filepath.Join(root, "removing", built.Host)
	if err := os.MkdirAll(quarantine, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err == nil {
		t.Fatal("removal recovery accepted simultaneous published and quarantined trees")
	}
	if _, err := os.Stat(packageDirectory(root, built.Host, built.ArchiveDigest)); err != nil {
		t.Fatalf("blocked recovery changed published package: %v", err)
	}
	if _, err := os.Stat(quarantine); err != nil {
		t.Fatalf("blocked recovery changed quarantine conflict: %v", err)
	}
	if err := os.Remove(quarantine); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, built.Host); err != nil {
		t.Fatalf("recovery after removing conflicting tree: %v", err)
	}
}

func TestFixtureLifecycleRemovalRecoveryRejectsReceiptMismatch(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "claude", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, ""); err != nil {
		t.Fatal(err)
	}
	if err := removeFixtureWithFault(root, base.Host, "after-removal-journal"); !errors.Is(err, errFixtureInterrupted) {
		t.Fatalf("interruption error=%v", err)
	}
	original, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil {
		t.Fatal(err)
	}
	changed := cloneLifecycleReceipt(original)
	changed.ActiveDigest, changed.PreviousDigest = changed.PreviousDigest, changed.ActiveDigest
	if err := writeLifecycleJSON(root, receiptRelative(base.Host), changed); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, base.Host); err == nil {
		t.Fatal("removal recovery accepted a receipt that diverged from its journal")
	}
	if _, err := os.Stat(packageDirectory(root, base.Host, base.ArchiveDigest)); err != nil {
		t.Fatalf("blocked recovery changed package bytes: %v", err)
	}
	if err := writeLifecycleJSON(root, receiptRelative(base.Host), original); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, base.Host); err != nil {
		t.Fatalf("recovery after restoring receipt: %v", err)
	}
}

func TestFixtureLifecycleMissingOrMalformedReceiptPreservesBytes(t *testing.T) {
	for _, mode := range []string{"missing", "unknown-field"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
			if err := installFixture(root, built, ""); err != nil {
				t.Fatal(err)
			}
			if mode == "missing" {
				if err := os.Remove(receiptPath(root, built.Host)); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(receiptPath(root, built.Host), []byte(`{"schema":1,"unknown":true}`+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := removeFixture(root, built.Host); err == nil {
				t.Fatal("invalid-receipt removal passed")
			}
			if _, err := os.Stat(packageDirectory(root, built.Host, built.ArchiveDigest)); err != nil {
				t.Fatalf("package bytes changed: %v", err)
			}
		})
	}
}

func TestFixtureLifecycleRejectsPackageSubstitution(t *testing.T) {
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	built.ArchiveDigest = "00" + built.ArchiveDigest[2:]
	if err := installFixture(t.TempDir(), built, ""); err == nil {
		t.Fatal("substituted package digest passed")
	}
}

func TestFixtureLifecycleRejectsInvalidFixtureVersionBeforeMutation(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "garbage.fixture", "owned\n")
	if err := installFixture(root, built, ""); err == nil {
		t.Fatal("invalid fixture version passed")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid fixture version wrote lifecycle state: entries=%v error=%v", entries, err)
	}
}

func TestLifecycleJSONWriterRejectsUnreadableRecord(t *testing.T) {
	root := t.TempDir()
	value := struct {
		Data string `json:"data"`
	}{Data: strings.Repeat("x", maximumFileSize)}
	if err := writeLifecycleJSON(root, "pending/oversized.json", value); err == nil {
		t.Fatal("oversized lifecycle record passed")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("oversized lifecycle record wrote state: entries=%v error=%v", entries, err)
	}
}

func TestRemoveEmptyParentsResumesPastMissingChild(t *testing.T) {
	root := t.TempDir()
	packageRoot := filepath.Join(root, "package")
	emptyAncestor := filepath.Join(packageRoot, "a")
	if err := os.MkdirAll(emptyAncestor, 0o755); err != nil {
		t.Fatal(err)
	}
	removeEmptyParents(filepath.Join(emptyAncestor, "already-missing"), packageRoot)
	if _, err := os.Stat(emptyAncestor); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty ancestor remains after resumed cleanup: %v", err)
	}
	if _, err := os.Stat(packageRoot); err != nil {
		t.Fatalf("cleanup crossed its stop boundary: %v", err)
	}
}

func TestFixtureLifecycleRejectsArchiveIdentityRelabeling(t *testing.T) {
	original := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	tests := []struct {
		name   string
		mutate func(*builtPackage)
	}{
		{name: "host", mutate: func(value *builtPackage) { value.Host = "claude" }},
		{name: "version", mutate: func(value *builtPackage) { value.Version = "0.1.999" }},
		{name: "catalog digest", mutate: func(value *builtPackage) { value.CatalogDigest = strings.Repeat("0", 64) }},
		{name: "source digest", mutate: func(value *builtPackage) { value.SourceDigest = strings.Repeat("0", 64) }},
		{name: "missing catalog", mutate: func(value *builtPackage) {
			value.Catalog = nil
			value.CatalogDigest = sha256Hex(value.Catalog)
			rewriteTestDistribution(t, value, func(metadata *distributionMetadata) {
				metadata.CatalogSHA256 = value.CatalogDigest
			})
		}},
		{name: "noncanonical catalog", mutate: func(value *builtPackage) {
			value.Catalog = append(value.Catalog, '\n')
			value.CatalogDigest = sha256Hex(value.Catalog)
			rewriteTestDistribution(t, value, func(metadata *distributionMetadata) {
				metadata.CatalogSHA256 = value.CatalogDigest
			})
		}},
		{name: "opposite manifest", mutate: func(value *builtPackage) {
			manifest, err := jsonBytes(claudeManifest{
				Schema: "https://json.schemastore.org/claude-code-plugin-manifest.json",
				Name:   "level7-dev-loop", DisplayName: "Level 7 Dev Loop", Version: value.Version,
			})
			if err != nil {
				t.Fatal(err)
			}
			entries := cloneEntries(value.Entries)
			entries = append(entries, archiveEntry{Name: ".claude-plugin/plugin.json", Data: manifest, Mode: 0o644})
			rebuildTestPackage(t, value, entries)
		}},
		{name: "distribution version", mutate: func(value *builtPackage) {
			entries := cloneEntries(value.Entries)
			for index := range entries {
				if entries[index].Name == "DISTRIBUTION.json" {
					var metadata distributionMetadata
					if err := decodeStrict(entries[index].Data, &metadata); err != nil {
						t.Fatal(err)
					}
					metadata.Version = "0.1.999"
					data, err := jsonBytes(metadata)
					if err != nil {
						t.Fatal(err)
					}
					entries[index].Data = data
				}
			}
			rebuildTestPackage(t, value, entries)
		}},
		{name: "manifest version", mutate: func(value *builtPackage) {
			entries := cloneEntries(value.Entries)
			for index := range entries {
				if entries[index].Name == ".codex-plugin/plugin.json" {
					var manifest codexManifest
					if err := decodeStrict(entries[index].Data, &manifest); err != nil {
						t.Fatal(err)
					}
					manifest.Version = "0.1.999"
					data, err := jsonBytes(manifest)
					if err != nil {
						t.Fatal(err)
					}
					entries[index].Data = data
				}
			}
			rebuildTestPackage(t, value, entries)
		}},
		{name: "projection version", mutate: func(value *builtPackage) {
			entries := cloneEntries(value.Entries)
			for index := range entries {
				if entries[index].Name == "COMPATIBILITY.json" {
					var compatibility compatibilityProjection
					if err := decodeStrict(entries[index].Data, &compatibility); err != nil {
						t.Fatal(err)
					}
					compatibility.PackageVersion = "0.1.999"
					data, err := jsonBytes(compatibility)
					if err != nil {
						t.Fatal(err)
					}
					entries[index].Data = data
				}
			}
			rebuildTestPackage(t, value, entries)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			value := original
			value.Entries = cloneEntries(original.Entries)
			value.Archive = append([]byte{}, original.Archive...)
			value.Catalog = append([]byte{}, original.Catalog...)
			test.mutate(&value)
			if err := installFixture(root, value, ""); err == nil {
				t.Fatal("relabelled package passed install")
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("rejected package wrote lifecycle state: entries=%v error=%v", entries, err)
			}
		})
	}
}

func rewriteTestDistribution(t *testing.T, built *builtPackage, mutate func(*distributionMetadata)) {
	t.Helper()
	entries := cloneEntries(built.Entries)
	found := false
	for index := range entries {
		if entries[index].Name != "DISTRIBUTION.json" {
			continue
		}
		var metadata distributionMetadata
		if err := decodeStrict(entries[index].Data, &metadata); err != nil {
			t.Fatal(err)
		}
		mutate(&metadata)
		data, err := jsonBytes(metadata)
		if err != nil {
			t.Fatal(err)
		}
		entries[index].Data = data
		found = true
	}
	if !found {
		t.Fatal("test package lacks distribution metadata")
	}
	rebuildTestPackage(t, built, entries)
}

func TestFixtureLifecycleRejectsInactiveDigestRelabeling(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "claude", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	if err := installFixture(root, upgrade, ""); err != nil {
		t.Fatal(err)
	}
	before, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil {
		t.Fatal(err)
	}
	inactive, ok := findReceiptPackage(before, base.ArchiveDigest)
	if !ok {
		t.Fatal("inactive package is absent from receipt")
	}
	inactive.Version = "0.1.999"
	baseReceiptDigest, err := lifecycleReceiptStateDigest(before)
	if err != nil {
		t.Fatal(err)
	}
	pending := pendingInstall{
		Schema:                 1,
		Host:                   base.Host,
		PreviousDigest:         upgrade.ArchiveDigest,
		BaseReceiptDigest:      baseReceiptDigest,
		CommittedReceiptDigest: sha256Hex([]byte("invalid committed receipt fixture\n")),
		Package:                inactive,
	}
	if err := writeLifecycleJSON(root, pendingRelative(base.Host), pending); err != nil {
		t.Fatal(err)
	}
	if err := recoverFixture(root, base.Host); err == nil {
		t.Fatal("inactive known digest was relabelled")
	}
	after, err := loadLifecycleReceipt(root, base.Host, true)
	if err != nil || !equalLifecycleReceipt(before, after) {
		t.Fatalf("blocked relabelling changed receipt: before=%+v after=%+v error=%v", before, after, err)
	}
	if _, err := loadPending(root, base.Host); err != nil {
		t.Fatalf("blocked relabelling removed diagnostic pending state: %v", err)
	}
}

func TestFixtureLifecycleMalformedRemovalJournalPreservesBytes(t *testing.T) {
	for _, mode := range []string{"unknown-field", "cross-host", "escaping-receipt"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
			if err := installFixture(root, built, ""); err != nil {
				t.Fatal(err)
			}
			receipt, err := loadLifecycleReceipt(root, built.Host, true)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "unknown-field":
				if err := writeRegular(root, pendingRemovalRelative(built.Host), []byte(`{"schema":1,"unknown":true}`+"\n")); err != nil {
					t.Fatal(err)
				}
			case "cross-host":
				if err := writeLifecycleJSON(root, pendingRemovalRelative(built.Host), pendingRemoval{Schema: 1, Host: "claude", Receipt: receipt}); err != nil {
					t.Fatal(err)
				}
			case "escaping-receipt":
				changed := cloneLifecycleReceipt(receipt)
				changed.Packages[0].Files[0].Path = "../outside"
				if err := writeLifecycleJSON(root, pendingRemovalRelative(built.Host), pendingRemoval{Schema: 1, Host: built.Host, Receipt: changed}); err != nil {
					t.Fatal(err)
				}
			}
			if err := recoverFixture(root, built.Host); err == nil {
				t.Fatal("malformed removal journal passed recovery")
			}
			if err := removeFixture(root, built.Host); err == nil {
				t.Fatal("malformed removal journal passed removal")
			}
			if _, err := os.Stat(packageDirectory(root, built.Host, built.ArchiveDigest)); err != nil {
				t.Fatalf("malformed journal changed package bytes: %v", err)
			}
			after, err := loadLifecycleReceipt(root, built.Host, true)
			if err != nil || !equalLifecycleReceipt(receipt, after) {
				t.Fatalf("malformed journal changed receipt: before=%+v after=%+v error=%v", receipt, after, err)
			}
		})
	}
}

func TestFixtureLifecycleRejectsReceiptReclassification(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	receipt, err := loadLifecycleReceipt(root, built.Host, true)
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte("changed and falsely receipted\n")
	changedPath := filepath.Join(packageDirectory(root, built.Host, built.ArchiveDigest), filepath.FromSlash(receipt.Packages[0].Files[0].Path))
	if err := os.WriteFile(changedPath, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	receipt.Packages[0].Files[0].Size = len(changed)
	receipt.Packages[0].Files[0].SHA256 = sha256Hex(changed)
	if err := writeLifecycleJSON(root, receiptRelative(built.Host), receipt); err != nil {
		t.Fatal(err)
	}
	preview, err := prepareRemoval(root, built.Host)
	if err != nil || len(preview.Conflicts) == 0 {
		t.Fatalf("tampered receipt preview=%+v error=%v", preview, err)
	}
	if err := removeFixture(root, built.Host); err == nil {
		t.Fatal("receipt-reclassified package bytes passed removal")
	}
	if data, err := os.ReadFile(changedPath); err != nil || string(data) != string(changed) {
		t.Fatalf("blocked removal changed reclassified bytes: data=%q error=%v", data, err)
	}
}

func TestFixtureLifecycleRejectsSymlinkedPackageRoot(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}
	hostRoot := filepath.Join(root, "packages", built.Host)
	relocated := filepath.Join(t.TempDir(), built.Host)
	if err := os.Rename(hostRoot, relocated); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(relocated, hostRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareRemoval(root, built.Host); err == nil {
		t.Fatal("symlinked package root passed removal preview")
	}
	if err := removeFixture(root, built.Host); err == nil {
		t.Fatal("symlinked package root passed removal")
	}
	if _, err := os.Stat(filepath.Join(relocated, built.ArchiveDigest)); err != nil {
		t.Fatalf("blocked removal changed relocated package: %v", err)
	}
}

func TestFixtureLifecycleDirectorySyncPreflightFailsClosed(t *testing.T) {
	testError := errors.New("test directory sync unsupported")

	t.Run("install", func(t *testing.T) {
		root := t.TempDir()
		built := testBuiltPackage(t, "codex", "0.1.0", "owned\n")
		realSync := lifecycleSyncDirectory
		calls := 0
		restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
			calls++
			if err := realSync(directory); err != nil {
				return err
			}
			return testError
		})

		err := installFixture(root, built, "")
		restore()
		if !errors.Is(err, testError) {
			t.Fatalf("install preflight error=%v", err)
		}
		if calls != 1 {
			t.Fatalf("install attempted %d directory syncs after its failed preflight", calls)
		}
		for _, name := range []string{
			pendingPath(root, built.Host),
			pendingStageDirectory(root, built.Host, built.ArchiveDigest),
			packageDirectory(root, built.Host, built.ArchiveDigest),
			receiptPath(root, built.Host),
		} {
			if lifecycleTestPathExists(t, name) {
				t.Fatalf("failed install preflight created lifecycle state at %s", name)
			}
		}
	})

	t.Run("removal", func(t *testing.T) {
		root := t.TempDir()
		built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
		if err := installFixture(root, built, ""); err != nil {
			t.Fatal(err)
		}
		receiptBefore, err := os.ReadFile(receiptPath(root, built.Host))
		if err != nil {
			t.Fatal(err)
		}
		realSync := lifecycleSyncDirectory
		calls := 0
		restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
			calls++
			if err := realSync(directory); err != nil {
				return err
			}
			return testError
		})

		err = removeFixture(root, built.Host)
		restore()
		if !errors.Is(err, testError) {
			t.Fatalf("removal preflight error=%v", err)
		}
		if calls != 1 {
			t.Fatalf("removal attempted %d directory syncs after its failed preflight", calls)
		}
		if lifecycleTestPathExists(t, pendingRemovalPath(root, built.Host)) ||
			lifecycleTestPathExists(t, filepath.Join(root, "removing", built.Host)) {
			t.Fatal("failed removal preflight created transaction state")
		}
		if err := verifyInstalledDirectory(root, packageDirectory(root, built.Host, built.ArchiveDigest), receiptPackageFromBuilt(built)); err != nil {
			t.Fatalf("failed removal preflight changed the package: %v", err)
		}
		receiptAfter, err := os.ReadFile(receiptPath(root, built.Host))
		if err != nil || string(receiptAfter) != string(receiptBefore) {
			t.Fatalf("failed removal preflight changed the receipt: error=%v", err)
		}
	})
}

func TestEnsureDurableRootPublishesEachCreatedComponent(t *testing.T) {
	parent := t.TempDir()
	outer := filepath.Join(parent, "outer")
	root := filepath.Join(outer, "host")
	type observation struct {
		directory    string
		outerPresent bool
		rootPresent  bool
	}
	realSync := lifecycleSyncDirectory
	observations := []observation{}
	restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
		if err := realSync(directory); err != nil {
			return err
		}
		observations = append(observations, observation{
			directory:    filepath.Clean(directory),
			outerPresent: lifecycleTestPathExists(t, outer),
			rootPresent:  lifecycleTestPathExists(t, root),
		})
		return nil
	})
	if err := ensureDurableRoot(root); err != nil {
		restore()
		t.Fatal(err)
	}
	restore()

	find := func(start int, match func(observation) bool) int {
		for index := start; index < len(observations); index++ {
			if match(observations[index]) {
				return index
			}
		}
		return -1
	}
	outerSync := find(0, func(value observation) bool {
		return value.directory == outer && value.outerPresent && !value.rootPresent
	})
	outerParentSync := find(outerSync+1, func(value observation) bool {
		return value.directory == parent && value.outerPresent && !value.rootPresent
	})
	rootSync := find(outerParentSync+1, func(value observation) bool {
		return value.directory == root && value.rootPresent
	})
	rootParentSync := find(rootSync+1, func(value observation) bool {
		return value.directory == outer && value.rootPresent
	})
	if outerSync < 0 || outerParentSync < 0 || rootSync < 0 || rootParentSync < 0 {
		t.Fatalf("missing root-publication barrier sequence: outer=%d outer-parent=%d root=%d root-parent=%d observations=%+v",
			outerSync, outerParentSync, rootSync, rootParentSync, observations)
	}
}

func TestFixtureLifecycleInstallDurabilityBarrierOrder(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}

	realSync := lifecycleSyncDirectory
	observations := []lifecycleSyncObservation{}
	restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
		if err := realSync(directory); err != nil {
			return err
		}
		observations = append(observations, observeLifecycleSync(t, root, upgrade, directory))
		return nil
	})
	if err := installFixture(root, upgrade, ""); err != nil {
		restore()
		t.Fatal(err)
	}
	restore()

	pendingDirectory := "pending"
	stageDirectory := lifecycleTestRelative(t, root, pendingStageDirectory(root, upgrade.Host, upgrade.ArchiveDigest))
	packageHostDirectory := filepath.ToSlash(filepath.Join("packages", upgrade.Host))
	journal := firstLifecycleSync(observations, 0, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == pendingDirectory && observation.PendingInstall &&
			!observation.FinalPackage && observation.ActiveDigest == base.ArchiveDigest
	})
	stage := firstLifecycleSync(observations, journal+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == stageDirectory && observation.PendingInstall && observation.Stage &&
			observation.OwnedFiles == len(upgrade.Entries) && !observation.FinalPackage && observation.ActiveDigest == base.ArchiveDigest
	})
	published := firstLifecycleSync(observations, stage+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == packageHostDirectory && observation.PendingInstall &&
			!observation.Stage && observation.FinalPackage && observation.ActiveDigest == base.ArchiveDigest
	})
	receipt := firstLifecycleSync(observations, published+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "receipts" && observation.PendingInstall &&
			observation.FinalPackage && observation.ActiveDigest == upgrade.ArchiveDigest
	})
	cleared := firstLifecycleSync(observations, receipt+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == pendingDirectory && !observation.PendingInstall &&
			observation.FinalPackage && observation.ActiveDigest == upgrade.ArchiveDigest
	})
	if journal < 0 || stage < 0 || published < 0 || receipt < 0 || cleared < 0 {
		t.Fatalf("missing ordered install durability barrier: journal=%d stage=%d publish=%d receipt=%d clear=%d observations=%+v",
			journal, stage, published, receipt, cleared, observations)
	}

	expectedStageDirectories := ownedDirectorySet(receiptPackageFromBuilt(upgrade).Files)
	expectedStageDirectories["."] = true
	observedStageDirectories := make(map[string]bool)
	stageRoot := pendingStageDirectory(root, upgrade.Host, upgrade.ArchiveDigest)
	for index := journal + 1; index < published; index++ {
		relative, err := filepath.Rel(stageRoot, filepath.Join(root, filepath.FromSlash(observations[index].Directory)))
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			observedStageDirectories[filepath.ToSlash(relative)] = true
		}
	}
	for directory := range expectedStageDirectories {
		if !observedStageDirectories[directory] {
			t.Errorf("staged directory %q had no durability barrier before package publication", directory)
		}
	}
}

func TestFixtureLifecycleInstallDirectorySyncFailuresRecover(t *testing.T) {
	testError := errors.New("test install directory sync failure")
	for _, phase := range []struct {
		name          string
		match         func(lifecycleSyncObservation, builtPackage, builtPackage) bool
		wantUpgrade   bool
		wantPending   bool
		wantPublished bool
	}{
		{
			name: "journal-publication",
			match: func(observation lifecycleSyncObservation, base, _ builtPackage) bool {
				return observation.Directory == "pending" && observation.PendingInstall &&
					!observation.Stage && !observation.FinalPackage && observation.ActiveDigest == base.ArchiveDigest
			},
			wantPending: true,
		},
		{
			name: "stage-tree",
			match: func(observation lifecycleSyncObservation, base, upgrade builtPackage) bool {
				return observation.Directory == pendingStageRelative(upgrade.Host, upgrade.ArchiveDigest) &&
					observation.PendingInstall && observation.Stage && observation.OwnedFiles == len(upgrade.Entries) &&
					!observation.FinalPackage && observation.ActiveDigest == base.ArchiveDigest
			},
			wantPending: true,
		},
		{
			name: "package-publication",
			match: func(observation lifecycleSyncObservation, base, upgrade builtPackage) bool {
				return observation.Directory == filepath.ToSlash(filepath.Join("packages", upgrade.Host)) &&
					observation.PendingInstall && observation.FinalPackage && !observation.Stage && observation.ActiveDigest == base.ArchiveDigest
			},
			wantUpgrade: true, wantPending: true, wantPublished: true,
		},
		{
			name: "receipt-publication",
			match: func(observation lifecycleSyncObservation, _, upgrade builtPackage) bool {
				return observation.Directory == "receipts" && observation.PendingInstall &&
					observation.FinalPackage && observation.ActiveDigest == upgrade.ArchiveDigest
			},
			wantUpgrade: true, wantPending: true, wantPublished: true,
		},
		{
			name: "journal-removal",
			match: func(observation lifecycleSyncObservation, _, upgrade builtPackage) bool {
				return observation.Directory == "pending" && !observation.PendingInstall &&
					observation.FinalPackage && observation.ActiveDigest == upgrade.ArchiveDigest
			},
			wantUpgrade: true, wantPublished: true,
		},
	} {
		t.Run(phase.name, func(t *testing.T) {
			root := t.TempDir()
			base := testBuiltPackage(t, "codex", "0.1.0", "base\n")
			upgrade := testBuiltPackage(t, "codex", "0.1.1", "upgrade\n")
			if err := installFixture(root, base, ""); err != nil {
				t.Fatal(err)
			}
			realSync := lifecycleSyncDirectory
			injected := false
			restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
				if err := realSync(directory); err != nil {
					return err
				}
				observation := observeLifecycleSync(t, root, upgrade, directory)
				if !injected && phase.match(observation, base, upgrade) {
					injected = true
					return testError
				}
				return nil
			})
			err := installFixture(root, upgrade, "")
			restore()
			if !injected || !errors.Is(err, testError) {
				t.Fatalf("injected=%v error=%v", injected, err)
			}
			if lifecycleTestPathExists(t, pendingPath(root, upgrade.Host)) != phase.wantPending {
				t.Fatalf("pending install presence does not match failed phase %q", phase.name)
			}
			if lifecycleTestPathExists(t, packageDirectory(root, upgrade.Host, upgrade.ArchiveDigest)) != phase.wantPublished {
				t.Fatalf("published package presence does not match failed phase %q", phase.name)
			}
			if err := recoverFixture(root, upgrade.Host); err != nil {
				t.Fatalf("recover after %s: %v", phase.name, err)
			}
			receipt, err := loadLifecycleReceipt(root, upgrade.Host, true)
			if err != nil {
				t.Fatal(err)
			}
			wantDigest := base.ArchiveDigest
			if phase.wantUpgrade {
				wantDigest = upgrade.ArchiveDigest
			}
			if receipt.ActiveDigest != wantDigest {
				t.Fatalf("recovery active digest=%q want=%q", receipt.ActiveDigest, wantDigest)
			}
			if lifecycleTestPathExists(t, pendingPath(root, upgrade.Host)) ||
				lifecycleTestPathExists(t, pendingStageDirectory(root, upgrade.Host, upgrade.ArchiveDigest)) {
				t.Fatal("install recovery left transaction state")
			}
		})
	}
}

func TestFixtureLifecycleRemovalDurabilityBarrierOrder(t *testing.T) {
	root := t.TempDir()
	built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
	if err := installFixture(root, built, ""); err != nil {
		t.Fatal(err)
	}

	realSync := lifecycleSyncDirectory
	observations := []lifecycleSyncObservation{}
	restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
		if err := realSync(directory); err != nil {
			return err
		}
		observations = append(observations, observeLifecycleSync(t, root, built, directory))
		return nil
	})
	if err := removeFixture(root, built.Host); err != nil {
		restore()
		t.Fatal(err)
	}
	restore()

	journal := firstLifecycleSync(observations, 0, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "pending-removals" && observation.PendingRemoval &&
			observation.SourceTree && !observation.QuarantineTree && observation.Receipt
	})
	sourceRename := firstLifecycleSync(observations, journal+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "packages" && observation.PendingRemoval &&
			!observation.SourceTree && observation.QuarantineTree && observation.Receipt
	})
	targetRename := firstLifecycleSync(observations, journal+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "removing" && observation.PendingRemoval &&
			!observation.SourceTree && observation.QuarantineTree && observation.Receipt &&
			observation.OwnedFiles == len(built.Entries)
	})
	firstDelete := firstLifecycleSync(observations, maxLifecycleTestIndex(sourceRename, targetRename)+1, func(observation lifecycleSyncObservation) bool {
		return observation.PendingRemoval && observation.QuarantineTree && observation.Receipt &&
			observation.OwnedFiles < len(built.Entries)
	})
	treeRemoval := firstLifecycleSync(observations, firstDelete+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "removing" && observation.PendingRemoval &&
			!observation.SourceTree && !observation.QuarantineTree && observation.Receipt
	})
	receiptRemoval := firstLifecycleSync(observations, treeRemoval+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "receipts" && observation.PendingRemoval &&
			!observation.SourceTree && !observation.QuarantineTree && !observation.Receipt
	})
	journalRemoval := firstLifecycleSync(observations, receiptRemoval+1, func(observation lifecycleSyncObservation) bool {
		return observation.Directory == "pending-removals" && !observation.PendingRemoval &&
			!observation.SourceTree && !observation.QuarantineTree && !observation.Receipt
	})
	if journal < 0 || sourceRename < 0 || targetRename < 0 || firstDelete < 0 || treeRemoval < 0 || receiptRemoval < 0 || journalRemoval < 0 {
		t.Fatalf("missing ordered removal durability barrier: journal=%d source-rename=%d target-rename=%d first-delete=%d tree=%d receipt=%d clear=%d observations=%+v",
			journal, sourceRename, targetRename, firstDelete, treeRemoval, receiptRemoval, journalRemoval, observations)
	}
	if targetRename >= sourceRename {
		t.Fatalf("cross-directory rename barriers were not destination-before-source: destination=%d source=%d", targetRename, sourceRename)
	}
	seenRemaining := make(map[int]bool)
	for index := maxLifecycleTestIndex(sourceRename, targetRename) + 1; index < treeRemoval; index++ {
		observation := observations[index]
		if observation.PendingRemoval && observation.Receipt && observation.QuarantineTree {
			seenRemaining[observation.OwnedFiles] = true
		}
	}
	for remaining := len(built.Entries) - 1; remaining >= 0; remaining-- {
		if !seenRemaining[remaining] {
			t.Errorf("owned-file deletion leaving %d files had no directory durability barrier", remaining)
		}
	}
}

func TestFixtureLifecycleRemovalDirectorySyncFailuresRecover(t *testing.T) {
	testError := errors.New("test removal directory sync failure")
	for _, phase := range []struct {
		name  string
		match func(lifecycleSyncObservation, builtPackage) bool
	}{
		{
			name: "journal-publication",
			match: func(observation lifecycleSyncObservation, _ builtPackage) bool {
				return observation.Directory == "pending-removals" && observation.PendingRemoval &&
					observation.SourceTree && !observation.QuarantineTree && observation.Receipt
			},
		},
		{
			name: "source-rename",
			match: func(observation lifecycleSyncObservation, _ builtPackage) bool {
				return observation.Directory == "packages" && observation.PendingRemoval &&
					!observation.SourceTree && observation.QuarantineTree && observation.Receipt
			},
		},
		{
			name: "target-rename",
			match: func(observation lifecycleSyncObservation, built builtPackage) bool {
				return observation.Directory == "removing" && observation.PendingRemoval &&
					!observation.SourceTree && observation.QuarantineTree && observation.Receipt && observation.OwnedFiles == len(built.Entries)
			},
		},
		{
			name: "owned-file-deletion",
			match: func(observation lifecycleSyncObservation, built builtPackage) bool {
				return observation.PendingRemoval && observation.QuarantineTree && observation.Receipt &&
					observation.OwnedFiles < len(built.Entries)
			},
		},
		{
			name: "package-tree-deletion",
			match: func(observation lifecycleSyncObservation, _ builtPackage) bool {
				return observation.Directory == "removing" && observation.PendingRemoval &&
					!observation.SourceTree && !observation.QuarantineTree && observation.Receipt
			},
		},
		{
			name: "receipt-deletion",
			match: func(observation lifecycleSyncObservation, _ builtPackage) bool {
				return observation.Directory == "receipts" && observation.PendingRemoval &&
					!observation.SourceTree && !observation.QuarantineTree && !observation.Receipt
			},
		},
		{
			name: "journal-deletion",
			match: func(observation lifecycleSyncObservation, _ builtPackage) bool {
				return observation.Directory == "pending-removals" && !observation.PendingRemoval &&
					!observation.SourceTree && !observation.QuarantineTree && !observation.Receipt
			},
		},
	} {
		t.Run(phase.name, func(t *testing.T) {
			root := t.TempDir()
			built := testBuiltPackage(t, "claude", "0.1.0", "owned\n")
			if err := installFixture(root, built, ""); err != nil {
				t.Fatal(err)
			}
			realSync := lifecycleSyncDirectory
			injected := false
			restore := replaceLifecycleSyncDirectory(t, func(directory string) error {
				if err := realSync(directory); err != nil {
					return err
				}
				observation := observeLifecycleSync(t, root, built, directory)
				if !injected && phase.match(observation, built) {
					injected = true
					return testError
				}
				return nil
			})
			err := removeFixture(root, built.Host)
			restore()
			if !injected || !errors.Is(err, testError) {
				t.Fatalf("injected=%v error=%v", injected, err)
			}
			if phase.name != "journal-deletion" && !lifecycleTestPathExists(t, pendingRemovalPath(root, built.Host)) {
				t.Fatalf("failed phase %q lost its visible recovery journal", phase.name)
			}
			if err := recoverFixture(root, built.Host); err != nil {
				t.Fatalf("recover after %s: %v", phase.name, err)
			}
			for _, name := range []string{
				pendingRemovalPath(root, built.Host),
				receiptPath(root, built.Host),
				filepath.Join(root, "packages", built.Host),
				filepath.Join(root, "removing", built.Host),
			} {
				if lifecycleTestPathExists(t, name) {
					t.Fatalf("removal recovery left state at %s", name)
				}
			}
		})
	}
}

// These observations prove the order in which the lifecycle asks the filesystem
// for namespace durability and how it handles reported failures. They do not
// simulate a reboot or establish physical persistence across power loss.
type lifecycleSyncObservation struct {
	Directory      string
	PendingInstall bool
	PendingRemoval bool
	Stage          bool
	FinalPackage   bool
	SourceTree     bool
	QuarantineTree bool
	Receipt        bool
	ActiveDigest   string
	OwnedFiles     int
}

func observeLifecycleSync(t *testing.T, root string, built builtPackage, directory string) lifecycleSyncObservation {
	t.Helper()
	observation := lifecycleSyncObservation{
		Directory:      lifecycleTestRelative(t, root, directory),
		PendingInstall: lifecycleTestPathExists(t, pendingPath(root, built.Host)),
		PendingRemoval: lifecycleTestPathExists(t, pendingRemovalPath(root, built.Host)),
		Stage:          lifecycleTestPathExists(t, pendingStageDirectory(root, built.Host, built.ArchiveDigest)),
		FinalPackage:   lifecycleTestPathExists(t, packageDirectory(root, built.Host, built.ArchiveDigest)),
		SourceTree:     lifecycleTestPathExists(t, filepath.Join(root, "packages", built.Host)),
		QuarantineTree: lifecycleTestPathExists(t, filepath.Join(root, "removing", built.Host)),
		Receipt:        lifecycleTestPathExists(t, receiptPath(root, built.Host)),
	}
	if observation.Receipt {
		receipt, err := loadLifecycleReceipt(root, built.Host, true)
		if err != nil {
			t.Fatalf("observe lifecycle receipt: %v", err)
		}
		observation.ActiveDigest = receipt.ActiveDigest
	}
	packageRoot := pendingStageDirectory(root, built.Host, built.ArchiveDigest)
	if observation.QuarantineTree {
		packageRoot = filepath.Join(root, "removing", built.Host, built.ArchiveDigest)
	}
	for _, entry := range built.Entries {
		if lifecycleTestPathExists(t, filepath.Join(packageRoot, filepath.FromSlash(entry.Name))) {
			observation.OwnedFiles++
		}
	}
	return observation
}

func replaceLifecycleSyncDirectory(t *testing.T, replacement func(string) error) func() {
	t.Helper()
	previous := lifecycleSyncDirectory
	restored := false
	restore := func() {
		if restored {
			return
		}
		lifecycleSyncDirectory = previous
		restored = true
	}
	lifecycleSyncDirectory = replacement
	t.Cleanup(restore)
	return restore
}

func firstLifecycleSync(observations []lifecycleSyncObservation, start int, match func(lifecycleSyncObservation) bool) int {
	if start < 0 {
		return -1
	}
	for index := start; index < len(observations); index++ {
		if match(observations[index]) {
			return index
		}
	}
	return -1
}

func lifecycleTestRelative(t *testing.T, root, name string) string {
	t.Helper()
	relative, err := filepath.Rel(root, name)
	if err != nil {
		t.Fatalf("make lifecycle test path %q relative to %q: %v", name, root, err)
	}
	return filepath.ToSlash(relative)
}

func lifecycleTestPathExists(t *testing.T, name string) bool {
	t.Helper()
	_, err := os.Lstat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	t.Fatalf("inspect lifecycle test path %s: %v", name, err)
	return false
}

func maxLifecycleTestIndex(first, second int) int {
	if first > second {
		return first
	}
	return second
}

func testBuiltPackage(t *testing.T, host, version, content string) builtPackage {
	t.Helper()
	manifestPath, catalogPath, _, err := packageIdentityPaths(host)
	if err != nil {
		t.Fatal(err)
	}
	var manifestData []byte
	if host == "codex" {
		manifestData, err = jsonBytes(codexManifest{Name: "level7-dev-loop", Version: version, Skills: "./skills/"})
	} else {
		manifestData, err = jsonBytes(claudeManifest{
			Schema: "https://json.schemastore.org/claude-code-plugin-manifest.json",
			Name:   "level7-dev-loop", DisplayName: "Level 7 Dev Loop", Version: version,
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	descriptor := packageDescriptor{
		Name: "level7-dev-loop", Version: version,
		Description: "One-intent solo development: inspect, implement, test, repair, self-review, and hand off with optional team assurance.",
		Author:      author{Name: "Level 7 Engineering"}, License: "MIT",
		Hosts: hosts{
			Codex:  hostDescriptor{Category: "Developer Tools"},
			Claude: hostDescriptor{Category: "Development Tools"},
		},
	}
	catalog, err := renderCatalog(descriptor, host)
	if err != nil {
		t.Fatal(err)
	}
	catalogDigest := sha256Hex(catalog)
	sourceDigest := sha256Hex([]byte(host + "\x00" + version + "\x00" + catalogDigest + "\x00" + content))
	distributionData, err := jsonBytes(distributionMetadata{
		Schema: 2, Name: "level7-dev-loop", Version: version, Channel: "stable", Host: host,
		ManifestPath: manifestPath, CatalogPath: catalogPath, CatalogSHA256: catalogDigest, SourceDigest: sourceDigest,
		Builder: builderVersion, SupportClaim: "WITHHELD", ActualHostGate: "SMOKE_TESTED",
	})
	if err != nil {
		t.Fatal(err)
	}
	compatibilityEntries := expectedCompatibilityEntries()
	compatibilityEntry := compatibilityEntries[0]
	if host == "claude" {
		compatibilityEntry = compatibilityEntries[1]
	}
	compatibilityData, err := jsonBytes(compatibilityProjection{
		Schema: 1, PackageVersion: version, ArtifactSchema: "lean-risk-v1", Entry: compatibilityEntry,
	})
	if err != nil {
		t.Fatal(err)
	}
	permissionData, err := jsonBytes(permissionsProjection{
		Schema: 1, PackageVersion: version, Host: host,
		Permissions:  permissions{WorkspaceBoundary: "Synthetic lifecycle fixtures remain inside the test-owned temporary directory."},
		SupportClaim: "WITHHELD",
	})
	if err != nil {
		t.Fatal(err)
	}
	provenanceData, err := jsonBytes(provenanceInput{
		Schema: 1, Unsigned: true, Package: "level7-dev-loop", Version: version, Host: host,
		SourceDigest: sourceDigest, Builder: builderVersion,
		Recipe: "offline deterministic standard-library package assembly", ExternalInputs: []string{},
		Claim: "unsigned package input; authenticity is not established",
	})
	if err != nil {
		t.Fatal(err)
	}
	sbomData, err := jsonBytes(makeSBOM(packageDescriptor{
		Name: "level7-dev-loop", Version: version, Repository: "https://github.com/addressanup/level7-dev-loop", License: "MIT",
	}, host, sourceDigest))
	if err != nil {
		t.Fatal(err)
	}
	entries := []archiveEntry{
		{Name: manifestPath, Data: manifestData, Mode: 0o644},
		{Name: "COMPATIBILITY.json", Data: compatibilityData, Mode: 0o644},
		{Name: "DISTRIBUTION.json", Data: distributionData, Mode: 0o644},
		{Name: "PERMISSIONS.json", Data: permissionData, Mode: 0o644},
		{Name: "PROVENANCE.input.json", Data: provenanceData, Mode: 0o644},
		{Name: "SBOM.spdx.json", Data: sbomData, Mode: 0o644},
		{Name: "skills/l7-next/SKILL.md", Data: []byte(content), Mode: 0o644},
	}
	inventoryData, err := jsonBytes(makeInventory(entries))
	if err != nil {
		t.Fatal(err)
	}
	entries = append(entries, archiveEntry{Name: "INVENTORY.json", Data: inventoryData, Mode: 0o644})
	archive, err := createArchive(entries)
	if err != nil {
		t.Fatal(err)
	}
	return builtPackage{
		Host: host, Version: version, Entries: entries, Archive: archive, SourceDigest: sourceDigest,
		ArchiveDigest: sha256Hex(archive), ArchiveName: host + ".zip", CatalogPath: catalogPath,
		Catalog: catalog, CatalogDigest: catalogDigest,
	}
}

func cloneLifecycleReceipt(receipt lifecycleReceipt) lifecycleReceipt {
	clone := receipt
	clone.Packages = append([]receiptPackage{}, receipt.Packages...)
	for index := range clone.Packages {
		clone.Packages[index].Files = append([]receiptFile{}, receipt.Packages[index].Files...)
	}
	return clone
}

func rebuildTestPackage(t *testing.T, built *builtPackage, entries []archiveEntry) {
	t.Helper()
	payload := make([]archiveEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Name != "INVENTORY.json" {
			payload = append(payload, entry)
		}
	}
	sort.Slice(payload, func(i, j int) bool { return payload[i].Name < payload[j].Name })
	inventoryData, err := jsonBytes(makeInventory(payload))
	if err != nil {
		t.Fatal(err)
	}
	payload = append(payload, archiveEntry{Name: "INVENTORY.json", Data: inventoryData, Mode: 0o644})
	sort.Slice(payload, func(i, j int) bool { return payload[i].Name < payload[j].Name })
	archive, err := createArchive(payload)
	if err != nil {
		t.Fatal(err)
	}
	built.Entries, built.Archive, built.ArchiveDigest = payload, archive, sha256Hex(archive)
}
