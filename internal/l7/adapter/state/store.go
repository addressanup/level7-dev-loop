// Package state stores minimal disposable workflow context under Git's common directory.
package state

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	SchemaVersion = 1
	MaxActiveFile = 64 << 10
)

type activeFile struct {
	Schema    int               `json:"schema"`
	Kind      domain.ActiveKind `json:"kind"`
	ChangeID  string            `json:"change_id"`
	Tier      domain.RiskTier   `json:"tier,omitempty"`
	Base      string            `json:"base,omitempty"`
	Problem   string            `json:"problem,omitempty"`
	Scope     []string          `json:"scope,omitempty"`
	BriefPath string            `json:"brief_path,omitempty"`
}

func Load(commonDirectory string) (domain.ActiveChange, bool, error) {
	path, err := activePath(commonDirectory)
	if err != nil {
		return domain.ActiveChange{}, false, err
	}
	data, err := localfile.Read(path, MaxActiveFile)
	if errors.Is(err, os.ErrNotExist) {
		return domain.ActiveChange{}, false, nil
	}
	if err != nil {
		return domain.ActiveChange{}, false, err
	}
	return decodeActive(data)
}

func Save(commonDirectory string, active domain.ActiveChange) error {
	path, err := activePath(commonDirectory)
	if err != nil {
		return err
	}
	file, err := encodeActive(active)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	_, exists, loadErr := Load(commonDirectory)
	if loadErr != nil {
		return fmt.Errorf("refuse to replace invalid active state: %w", loadErr)
	}
	if exists {
		return localfile.AtomicReplace(path, file, 0o600)
	}
	return localfile.AtomicCreate(path, file, 0o600)
}

func Acquire(commonDirectory string) (*localfile.Lock, error) {
	directory, err := productDirectory(commonDirectory)
	if err != nil {
		return nil, err
	}
	if err := localfile.EnsureDirectory(directory, 0o700); err != nil {
		return nil, err
	}
	return localfile.AcquireLock(filepath.Join(directory, "lock"))
}

func encodeActive(active domain.ActiveChange) ([]byte, error) {
	file := activeFile{
		Schema:    SchemaVersion,
		Kind:      active.Kind,
		ChangeID:  active.ID,
		Tier:      active.Tier,
		Base:      active.Base,
		Problem:   active.Problem,
		Scope:     append([]string{}, active.Scope...),
		BriefPath: active.BriefPath,
	}
	if err := validateActive(file); err != nil {
		return nil, err
	}
	return localfile.EncodeJSON(file)
}

func decodeActive(data []byte) (domain.ActiveChange, bool, error) {
	var file activeFile
	if err := localfile.DecodeJSON(data, &file); err != nil {
		return domain.ActiveChange{}, false, err
	}
	if err := validateActive(file); err != nil {
		return domain.ActiveChange{}, false, err
	}
	return domain.ActiveChange{
		Kind:      file.Kind,
		ID:        file.ChangeID,
		Tier:      file.Tier,
		Base:      file.Base,
		Problem:   file.Problem,
		Scope:     append([]string{}, file.Scope...),
		BriefPath: file.BriefPath,
	}, true, nil
}

func validateActive(file activeFile) error {
	if file.Schema != SchemaVersion {
		return fmt.Errorf("unsupported active-state schema %d", file.Schema)
	}
	if !safeChangeID(file.ChangeID) {
		return errors.New("active change ID is invalid")
	}
	switch file.Kind {
	case domain.ActiveTierOne:
		if file.Tier != domain.TierRoutine || !fullGitID(file.Base) || !safeLine(file.Problem) || len(file.Scope) < 1 || len(file.Scope) > 256 || file.BriefPath != "" {
			return errors.New("Tier 1 active state is incomplete or conflicting")
		}
		seen := make(map[string]bool)
		for _, scoped := range file.Scope {
			if !safeRepositoryPath(scoped) || seen[scoped] {
				return errors.New("Tier 1 active scope is invalid")
			}
			seen[scoped] = true
		}
	case domain.ActiveBrief:
		if file.Tier != 0 || file.Base != "" || file.Problem != "" || len(file.Scope) != 0 || file.BriefPath != "docs/artifacts/changes/"+file.ChangeID+".md" {
			return errors.New("brief-backed active state duplicates or conflicts with the brief")
		}
	default:
		return errors.New("active state kind is invalid")
	}
	return nil
}

func activePath(commonDirectory string) (string, error) {
	directory, err := productDirectory(commonDirectory)
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "active.json"), nil
}

func productDirectory(commonDirectory string) (string, error) {
	if !filepath.IsAbs(commonDirectory) {
		return "", errors.New("Git common directory must be absolute")
	}
	return filepath.Join(filepath.Clean(commonDirectory), "l7", "product"), nil
}

func safeChangeID(value string) bool {
	if len(value) < 1 || len(value) > 64 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
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

func safeLine(value string) bool {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 2048 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character == '\n' || character == '\r' || character == 0 || character == 0x7f {
			return false
		}
	}
	return true
}

func safeRepositoryPath(value string) bool {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 512 || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "`") {
		return false
	}
	recursive := strings.HasSuffix(value, "/**")
	plain := strings.TrimSuffix(value, "/**")
	if plain == "" || plain == "." || plain == ".." || path.Clean(plain) != plain || strings.HasPrefix(plain, "../") || strings.ContainsAny(plain, "*?[") {
		return false
	}
	return recursive || !strings.HasSuffix(value, "/")
}
