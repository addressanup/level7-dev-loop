// Package config owns strict tracked Level 7 repository configuration and briefs.
package config

import (
	"crypto/sha256"
	"encoding/hex"
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

var configurationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

type File struct {
	Schema         int                   `json:"schema"`
	Features       Features              `json:"features"`
	Verification   []VerificationCommand `json:"verification"`
	Limits         Limits                `json:"limits"`
	ProtectedPaths []string              `json:"protected_paths"`
	Providers      Providers             `json:"providers"`
	digest         string
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

type fileWire struct {
	Schema         *int                   `json:"schema"`
	Features       *featuresWire          `json:"features"`
	Verification   *[]VerificationCommand `json:"verification"`
	Limits         *limitsWire            `json:"limits"`
	ProtectedPaths *[]string              `json:"protected_paths"`
	Providers      *providersWire         `json:"providers"`
}

type featuresWire struct {
	LocalLifecycle *bool `json:"local_lifecycle"`
}

type limitsWire struct {
	MaxInputBytes         *int `json:"max_input_bytes"`
	MaxGitOutputBytes     *int `json:"max_git_output_bytes"`
	MaxGitPaths           *int `json:"max_git_paths"`
	MaxCommandOutputBytes *int `json:"max_command_output_bytes"`
	MaxCommandSeconds     *int `json:"max_command_seconds"`
}

type providersWire struct {
	Implementer *string `json:"implementer"`
	Reviewer    *string `json:"reviewer"`
}

func Default(localLifecycle bool) File {
	configuration := File{
		Schema:         SchemaVersion,
		Features:       Features{LocalLifecycle: localLifecycle},
		Verification:   []VerificationCommand{},
		Limits:         Limits{MaxInputBytes: 1 << 20, MaxGitOutputBytes: 16 << 20, MaxGitPaths: 100_000, MaxCommandOutputBytes: 8 << 20, MaxCommandSeconds: 1800},
		ProtectedPaths: []string{},
		Providers:      Providers{},
	}
	configuration.setCanonicalDigest()
	return configuration
}

func (configuration File) Domain() domain.Configuration {
	verification := make([]domain.VerificationCommand, 0, len(configuration.Verification))
	for _, command := range configuration.Verification {
		verification = append(verification, domain.VerificationCommand{
			Name:      command.Name,
			Argv:      append([]string{}, command.Argv...),
			Benchmark: command.Benchmark,
		})
	}
	return domain.Configuration{
		Digest:                configuration.configurationDigest(),
		LocalLifecycle:        configuration.Features.LocalLifecycle,
		Verification:          verification,
		MaxInputBytes:         configuration.Limits.MaxInputBytes,
		MaxGitOutputBytes:     configuration.Limits.MaxGitOutputBytes,
		MaxGitPaths:           configuration.Limits.MaxGitPaths,
		MaxCommandOutputBytes: configuration.Limits.MaxCommandOutputBytes,
		MaxCommandSeconds:     configuration.Limits.MaxCommandSeconds,
		ProtectedPaths:        append([]string{}, configuration.ProtectedPaths...),
		Implementer:           domain.Provider(configuration.Providers.Implementer),
		Reviewer:              domain.Provider(configuration.Providers.Reviewer),
	}
}

func Load(root string) (File, error) {
	var configuration File
	data, err := localfile.Read(filepath.Join(root, filepath.FromSlash(configurationPath)), MaxConfiguration)
	if err != nil {
		return configuration, err
	}
	configuration, err = decodeConfiguration(data)
	if err != nil {
		return File{}, err
	}
	if err := configuration.Validate(); err != nil {
		return File{}, err
	}
	configuration.digest = digestBytes(data)
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
			configuration.digest = digestBytes(data)
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
	configuration.digest = digestBytes(data)
	return configuration, true, nil
}

func (configuration File) configurationDigest() string {
	if configuration.digest != "" {
		return configuration.digest
	}
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func (configuration *File) setCanonicalDigest() {
	configuration.digest = configuration.configurationDigest()
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
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
	for _, entry := range []struct {
		label    string
		provider string
	}{
		{label: "implementer", provider: configuration.Providers.Implementer},
		{label: "reviewer", provider: configuration.Providers.Reviewer},
	} {
		if entry.provider != "" && entry.provider != "codex" && entry.provider != "claude" {
			return fmt.Errorf("unsupported %s provider %q", entry.label, entry.provider)
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
	return configurationNamePattern.MatchString(value)
}

func safeText(value string, limit int, allowEmpty bool) bool {
	if !utf8.ValidString(value) || len(value) > limit || strings.ContainsRune(value, 0) || (!allowEmpty && value == "") {
		return false
	}
	for _, character := range value {
		if character == 0x7f || character < 0x20 {
			return false
		}
	}
	return true
}

func decodeConfiguration(data []byte) (File, error) {
	var wire fileWire
	if err := localfile.DecodeJSON(data, &wire); err != nil {
		return File{}, err
	}
	if wire.Schema == nil || wire.Features == nil || wire.Features.LocalLifecycle == nil || wire.Verification == nil || wire.Limits == nil || wire.Limits.MaxInputBytes == nil || wire.Limits.MaxGitOutputBytes == nil || wire.Limits.MaxGitPaths == nil || wire.Limits.MaxCommandOutputBytes == nil || wire.Limits.MaxCommandSeconds == nil || wire.ProtectedPaths == nil || wire.Providers == nil || wire.Providers.Implementer == nil || wire.Providers.Reviewer == nil {
		return File{}, errors.New("configuration is missing a required field")
	}
	return File{
		Schema:         *wire.Schema,
		Features:       Features{LocalLifecycle: *wire.Features.LocalLifecycle},
		Verification:   append([]VerificationCommand{}, (*wire.Verification)...),
		Limits:         Limits{MaxInputBytes: *wire.Limits.MaxInputBytes, MaxGitOutputBytes: *wire.Limits.MaxGitOutputBytes, MaxGitPaths: *wire.Limits.MaxGitPaths, MaxCommandOutputBytes: *wire.Limits.MaxCommandOutputBytes, MaxCommandSeconds: *wire.Limits.MaxCommandSeconds},
		ProtectedPaths: append([]string{}, (*wire.ProtectedPaths)...),
		Providers:      Providers{Implementer: *wire.Providers.Implementer, Reviewer: *wire.Providers.Reviewer},
	}, nil
}
