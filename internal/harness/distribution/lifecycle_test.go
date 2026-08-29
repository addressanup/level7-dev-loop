package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureLifecycleInstallUpgradeRecoverRollbackRemove(t *testing.T) {
	root := t.TempDir()
	base := testBuiltPackage(t, "codex", "0.1.0-dev.5", "base\n")
	upgrade := testBuiltPackage(t, "codex", "0.1.0-dev.6", "upgrade\n")

	if err := installFixture(root, base, ""); err != nil {
		t.Fatal(err)
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

func testBuiltPackage(t *testing.T, host, version, content string) builtPackage {
	t.Helper()
	entries := []archiveEntry{
		{Name: "MANIFEST.json", Data: []byte("{}\n"), Mode: 0o644},
		{Name: "skills/l7-next/SKILL.md", Data: []byte(content), Mode: 0o644},
	}
	archive, err := createArchive(entries)
	if err != nil {
		t.Fatal(err)
	}
	return builtPackage{Host: host, Version: version, Entries: entries, Archive: archive, ArchiveDigest: sha256Hex(archive), ArchiveName: host + ".zip"}
}
