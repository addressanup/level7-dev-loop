package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFixtureLifecycleInstallUpgradeRecoverRollbackRemove(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")

	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
	}
	sameArchiveDifferentVersion := base
	sameArchiveDifferentVersion.Version = "0.1.0-dev.999"
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
	base := testBuiltPackage(t, "claude", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.0-dev.6", "upgrade\n")
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
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")
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
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")
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
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")
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
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")
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
			base := testBuiltPackage(t, "claude", "0.1.0-dev.5", "base\n")
			upgrade := testBuiltPackage(t, "claude", "0.1.0-dev.6", "upgrade\n")
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
			base := testBuiltPackage(t, "claude", "0.1.0-dev.4", "base\n")
			active := testBuiltPackage(t, "claude", "0.1.0-dev.5", "active\n")
			pendingPackage := testBuiltPackage(t, "claude", "0.1.0-dev.6", "pending\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
			base := testBuiltPackage(t, "claude", "0.1.0-dev.5", "base\n")
			upgrade := testBuiltPackage(t, "claude", "0.1.0-dev.6", "upgrade\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "claude", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	base := testBuiltPackage(t, "claude", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.0-dev.6", "upgrade\n")
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
			built := testBuiltPackage(t, "claude", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	original := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
	tests := []struct {
		name   string
		mutate func(*builtPackage)
	}{
		{name: "host", mutate: func(value *builtPackage) { value.Host = "claude" }},
		{name: "version", mutate: func(value *builtPackage) { value.Version = "0.1.0-dev.999" }},
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
					metadata.Version = "0.1.0-dev.999"
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
					manifest.Version = "0.1.0-dev.999"
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
					compatibility.PackageVersion = "0.1.0-dev.999"
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

func TestFixtureLifecycleRejectsInactiveDigestRelabeling(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "claude", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "claude", "0.1.0-dev.6", "upgrade\n")
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
	inactive.Version = "0.1.0-dev.999"
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
			built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "codex", "0.1.0-dev.5", "owned\n")
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
	built := testBuiltPackage(t, "claude", "0.1.0-dev.5", "owned\n")
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
	sourceDigest := sha256Hex([]byte(host + "\x00" + version + "\x00" + content))
	distributionData, err := jsonBytes(distributionMetadata{
		Schema: 1, Name: "level7-dev-loop", Version: version, Channel: "development", Host: host,
		ManifestPath: manifestPath, CatalogPath: catalogPath, SourceDigest: sourceDigest,
		Builder: builderVersion, SupportClaim: "WITHHELD", ActualHostGate: "NOT_RUN",
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
		Claim: "development input only; authenticity and release promotion are not established",
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
		Host: host, Version: version, Entries: entries, Archive: archive, ArchiveDigest: sha256Hex(archive),
		ArchiveName: host + ".zip", CatalogPath: catalogPath,
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
