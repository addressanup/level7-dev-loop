// Package authority captures explicit local owner approval outside repository text.
package authority

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	schemaVersion   = 1
	maxApprovalFile = 16 << 10
)

type approvalFile struct {
	Schema      int             `json:"schema"`
	ChangeID    string          `json:"change_id"`
	Actor       string          `json:"actor"`
	Implementer domain.Provider `json:"implementer"`
	BriefCommit string          `json:"brief_commit"`
	Source      string          `json:"source"`
}

type Terminal struct {
	input       io.Reader
	output      io.Writer
	interactive bool
	actor       string
}

func NewTerminal(input io.Reader, output io.Writer, interactive bool, actor string) Terminal {
	return Terminal{input: input, output: output, interactive: interactive, actor: actor}
}

func (terminal Terminal) Confirm(ctx context.Context, changeID string, implementer domain.Provider, briefCommit string) (domain.ApprovalBinding, error) {
	if err := ctx.Err(); err != nil {
		return domain.ApprovalBinding{}, err
	}
	binding := domain.ApprovalBinding{ChangeID: changeID, Actor: terminal.actor, Implementer: implementer, BriefCommit: briefCommit}
	if !terminal.interactive {
		return domain.ApprovalBinding{}, errors.New("Tier 3 approval requires an active terminal interaction")
	}
	if err := validateBinding(binding); err != nil {
		return domain.ApprovalBinding{}, err
	}
	if terminal.input == nil || terminal.output == nil {
		return domain.ApprovalBinding{}, errors.New("approval terminal is unavailable")
	}
	if _, err := fmt.Fprintf(terminal.output, "Approve Tier 3 change %q for implementer %q at brief %s. Type the full change ID to continue: ", changeID, implementer, briefCommit); err != nil {
		return domain.ApprovalBinding{}, errors.New("cannot write approval prompt")
	}
	reader := bufio.NewReader(io.LimitReader(terminal.input, 1024))
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return domain.ApprovalBinding{}, errors.New("cannot read approval response")
	}
	if err := ctx.Err(); err != nil {
		return domain.ApprovalBinding{}, err
	}
	if strings.TrimSpace(answer) != changeID {
		return domain.ApprovalBinding{}, errors.New("Tier 3 approval was not confirmed")
	}
	return binding, nil
}

func Load(commonDirectory string) (domain.ApprovalBinding, bool, error) {
	name, err := approvalPath(commonDirectory)
	if err != nil {
		return domain.ApprovalBinding{}, false, err
	}
	data, err := localfile.Read(name, maxApprovalFile)
	if errors.Is(err, os.ErrNotExist) {
		return domain.ApprovalBinding{}, false, nil
	}
	if err != nil {
		return domain.ApprovalBinding{}, false, err
	}
	var file approvalFile
	if err := localfile.DecodeJSON(data, &file); err != nil {
		return domain.ApprovalBinding{}, false, err
	}
	binding := domain.ApprovalBinding{ChangeID: file.ChangeID, Actor: file.Actor, Implementer: file.Implementer, BriefCommit: file.BriefCommit}
	if file.Schema != schemaVersion || file.Source != "active-terminal-interaction" {
		return domain.ApprovalBinding{}, false, errors.New("approval schema or source is invalid")
	}
	if err := validateBinding(binding); err != nil {
		return domain.ApprovalBinding{}, false, err
	}
	return binding, true, nil
}

func Save(commonDirectory string, binding domain.ApprovalBinding) error {
	if err := validateBinding(binding); err != nil {
		return err
	}
	if _, _, err := Load(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid approval: %w", err)
	}
	name, err := approvalPath(commonDirectory)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(filepath.Dir(name), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(approvalFile{Schema: schemaVersion, ChangeID: binding.ChangeID, Actor: binding.Actor, Implementer: binding.Implementer, BriefCommit: binding.BriefCommit, Source: "active-terminal-interaction"})
	if err != nil {
		return err
	}
	_, readErr := localfile.Read(name, maxApprovalFile)
	if errors.Is(readErr, os.ErrNotExist) {
		return localfile.AtomicCreate(name, data, 0o600)
	}
	if readErr != nil {
		return readErr
	}
	return localfile.AtomicReplace(name, data, 0o600)
}

func Current(binding domain.ApprovalBinding, changeID string, implementer domain.Provider, briefCommit string) bool {
	return validateBinding(binding) == nil && binding.ChangeID == changeID && binding.Implementer == implementer && binding.BriefCommit == briefCommit
}

func approvalPath(commonDirectory string) (string, error) {
	if !filepath.IsAbs(commonDirectory) {
		return "", errors.New("Git common directory must be absolute")
	}
	return filepath.Join(filepath.Clean(commonDirectory), "l7", "product", "approval.json"), nil
}

func validateBinding(binding domain.ApprovalBinding) error {
	if !safeToken(binding.ChangeID, 64) || !safeToken(binding.Actor, 128) || !binding.Implementer.Valid() || !fullGitID(binding.BriefCommit) {
		return errors.New("approval binding is invalid")
	}
	return nil
}

func safeToken(value string, maximum int) bool {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > maximum || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character == 0x7f || character < 0x20 {
			return false
		}
	}
	return true
}

func fullGitID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
