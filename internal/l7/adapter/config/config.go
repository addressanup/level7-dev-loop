// Package config owns strict tracked Level 7 repository configuration and briefs.
package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	SchemaVersion     = 1
	MaxConfiguration  = 64 << 10
	configurationPath = ".l7/config.json"
)

type File struct {
	Schema         int                   `json:"schema"`
	Features       Features              `json:"features"`
	Verification   []VerificationCommand `json:"verification"`
	Limits         Limits                `json:"limits"`
	ProtectedPaths []string              `json:"protected_paths"`
	Providers      Providers             `json:"providers"`
}

type Features struct {
	LocalLifecycle bool `json:"local_lifecycle"`
}

type VerificationCommand struct {
	Name      string   `json:"name"`
	Argv      []string `json:"argv"`
	Benchmark bool     `json:"benchmark"`
}

type Limits struct {
	MaxInputBytes         int `json:"max_input_bytes"`
	MaxGitOutputBytes     int `json:"max_git_output_bytes"`
	MaxGitPaths           int `json:"max_git_paths"`
	MaxCommandOutputBytes int `json:"max_command_output_bytes"`
	MaxCommandSeconds     int `json:"max_command_seconds"`
}

type Providers struct {
	Implementer string `json:"implementer"`
	Reviewer    string `json:"reviewer"`
}

func Default(localLifecycle bool) File {
	return File{
		Schema:         SchemaVersion,
		Features:       Features{LocalLifecycle: localLifecycle},
		Verification:   []VerificationCommand{},
		Limits:         Limits{MaxInputBytes: 1 << 20, MaxGitOutputBytes: 16 << 20, MaxGitPaths: 100_000, MaxCommandOutputBytes: 8 << 20, MaxCommandSeconds: 1800},
		ProtectedPaths: []string{},
		Providers:      Providers{},
	}
}

func (configuration File) Domain() domain.Configuration {
	return domain.Configuration{LocalLifecycle: configuration.Features.LocalLifecycle}
}

func Load(root string) (File, error) {
	var configuration File
	data, err := localfile.Read(filepath.Join(root, filepath.FromSlash(configurationPath)), MaxConfiguration)
	if err != nil {
		return configuration, err
	}
	if err := localfile.DecodeJSON(data, &configuration); err != nil {
		return File{}, err
	}
	if err := configuration.Validate(); err != nil {
		return File{}, err
	}
	return configuration, nil
}

func Adopt(root string, enableLocalLifecycle bool) (File, bool, error) {
	if !filepath.IsAbs(root) {
		return File{}, false, errors.New("repository root must be absolute")
	}
	configuration, err := Load(root)
	if err == nil {
		if enableLocalLifecycle && !configuration.Features.LocalLifecycle {
			configuration.Features.LocalLifecycle = true
			data, encodeErr := localfile.EncodeJSON(configuration)
			if encodeErr != nil {
				return File{}, false, encodeErr
			}
			if writeErr := localfile.AtomicReplace(filepath.Join(root, filepath.FromSlash(configurationPath)), data, 0o644); writeErr != nil {
				return File{}, false, writeErr
			}
			return configuration, true, nil
		}
		return configuration, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return File{}, false, err
	}

	directory := filepath.Join(root, ".l7")
	if err := localfile.EnsureDirectory(directory, 0o755); err != nil {
		return File{}, false, err
	}
	configuration = Default(enableLocalLifecycle)
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		return File{}, false, err
	}
	if err := localfile.AtomicCreate(filepath.Join(directory, "config.json"), data, 0o644); err != nil {
		return File{}, false, err
	}
	return configuration, true, nil
}

func (configuration File) Validate() error {
	if configuration.Schema != SchemaVersion {
		return fmt.Errorf("unsupported configuration schema %d", configuration.Schema)
	}
	if configuration.Limits.MaxInputBytes < 1024 || configuration.Limits.MaxInputBytes > 16<<20 {
		return errors.New("max_input_bytes is outside 1024..16777216")
	}
	if configuration.Limits.MaxGitOutputBytes < 64<<10 || configuration.Limits.MaxGitOutputBytes > 64<<20 {
		return errors.New("max_git_output_bytes is outside 65536..67108864")
	}
	if configuration.Limits.MaxGitPaths < 1 || configuration.Limits.MaxGitPaths > 1_000_000 {
		return errors.New("max_git_paths is outside 1..1000000")
	}
	if configuration.Limits.MaxCommandOutputBytes < 64<<10 || configuration.Limits.MaxCommandOutputBytes > 64<<20 {
		return errors.New("max_command_output_bytes is outside 65536..67108864")
	}
	if configuration.Limits.MaxCommandSeconds < 1 || configuration.Limits.MaxCommandSeconds > 86_400 {
		return errors.New("max_command_seconds is outside 1..86400")
	}
	if len(configuration.Verification) > 32 {
		return errors.New("verification command count exceeds 32")
	}
	seenCommands := make(map[string]bool)
	for _, command := range configuration.Verification {
		if !safeName(command.Name) {
			return fmt.Errorf("invalid verification command name %q", command.Name)
		}
		if seenCommands[command.Name] {
			return fmt.Errorf("duplicate verification command name %q", command.Name)
		}
		seenCommands[command.Name] = true
		if len(command.Argv) < 1 || len(command.Argv) > 64 {
			return fmt.Errorf("verification command %q must contain 1..64 argv values", command.Name)
		}
		for _, argument := range command.Argv {
			if !safeText(argument, 4096, false) {
				return fmt.Errorf("verification command %q contains an invalid argv value", command.Name)
			}
		}
	}
	if len(configuration.ProtectedPaths) > 256 {
		return errors.New("protected path count exceeds 256")
	}
	seenPaths := make(map[string]bool)
	for _, protected := range configuration.ProtectedPaths {
		if err := ValidateRepositoryPath(protected); err != nil {
			return fmt.Errorf("invalid protected path %q: %w", protected, err)
		}
		if seenPaths[protected] {
			return fmt.Errorf("duplicate protected path %q", protected)
		}
		seenPaths[protected] = true
	}
	for label, provider := range map[string]string{"implementer": configuration.Providers.Implementer, "reviewer": configuration.Providers.Reviewer} {
		if provider != "" && provider != "codex" && provider != "claude" {
			return fmt.Errorf("unsupported %s provider %q", label, provider)
		}
	}
	return nil
}

func ValidateRepositoryPath(value string) error {
	if !utf8.ValidString(value) || len(value) < 1 || len(value) > 512 || strings.ContainsRune(value, 0) || strings.Contains(value, "\\") || strings.Contains(value, "`") {
		return errors.New("path encoding or length is unsafe")
	}
	recursive := strings.HasSuffix(value, "/**")
	plain := value
	if recursive {
		plain = strings.TrimSuffix(value, "/**")
	}
	if plain == "" || strings.HasPrefix(plain, "/") || plain == "." || plain == ".." || strings.HasPrefix(plain, "../") || path.Clean(plain) != plain {
		return errors.New("path must be normalized and repository-relative")
	}
	if strings.ContainsAny(plain, "*?[") || (!recursive && strings.HasSuffix(value, "/")) {
		return errors.New("only an explicit trailing /** recursion marker is supported")
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return errors.New("path contains a control character")
		}
	}
	return nil
}

func safeName(value string) bool {
	return regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`).MatchString(value)
}

func safeText(value string, limit int, allowEmpty bool) bool {
	if !utf8.ValidString(value) || len(value) > limit || strings.ContainsRune(value, 0) || (!allowEmpty && value == "") {
		return false
	}
	return true
}
