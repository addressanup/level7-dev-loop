package authority

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTerminalApprovalRequiresExactActiveConfirmation(t *testing.T) {
	commit := strings.Repeat("a", 40)
	var prompt bytes.Buffer
	terminal := NewTerminal(strings.NewReader("security-change\n"), &prompt, true, "accountable-owner")
	binding, err := terminal.Confirm(context.Background(), "security-change", domain.ProviderCodex, commit)
	if err != nil || binding.ChangeID != "security-change" || binding.Actor != "accountable-owner" || binding.BriefCommit != commit || !strings.Contains(prompt.String(), commit) {
		t.Fatalf("Confirm()=%+v prompt=%q error=%v", binding, prompt.String(), err)
	}
	for _, terminal := range []Terminal{
		NewTerminal(strings.NewReader("security-change\n"), &bytes.Buffer{}, false, "accountable-owner"),
		NewTerminal(strings.NewReader("wrong\n"), &bytes.Buffer{}, true, "accountable-owner"),
	} {
		if _, err := terminal.Confirm(context.Background(), "security-change", domain.ProviderCodex, commit); err == nil {
			t.Fatal("Confirm() accepted missing external approval")
		}
	}
}

func TestApprovalRoundTripAndFreshness(t *testing.T) {
	common := physicalDirectory(t)
	binding := domain.ApprovalBinding{ChangeID: "security-change", Actor: "accountable-owner", Implementer: domain.ProviderClaude, BriefCommit: strings.Repeat("b", 40)}
	if err := Save(common, binding); err != nil {
		t.Fatal(err)
	}
	loaded, found, err := Load(common)
	if err != nil || !found || loaded != binding || !Current(loaded, binding.ChangeID, binding.Implementer, binding.BriefCommit) {
		t.Fatalf("Load()=%+v found=%v error=%v", loaded, found, err)
	}
	if Current(loaded, binding.ChangeID, domain.ProviderCodex, binding.BriefCommit) {
		t.Fatal("approval remained current after implementer change")
	}
}

func TestInvalidApprovalIsPreserved(t *testing.T) {
	common := physicalDirectory(t)
	directory := filepath.Join(common, "l7", "product")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(directory, "approval.json")
	if err := os.WriteFile(name, []byte(`{"schema":1,"schema":2}`), 0o600); err != nil {
		t.Fatal(err)
	}
	binding := domain.ApprovalBinding{ChangeID: "security-change", Actor: "accountable-owner", Implementer: domain.ProviderCodex, BriefCommit: strings.Repeat("c", 40)}
	if err := Save(common, binding); err == nil {
		t.Fatal("Save() overwrote invalid approval")
	}
}

func physicalDirectory(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}
