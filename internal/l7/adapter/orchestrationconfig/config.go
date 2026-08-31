// Package orchestrationconfig owns the strict, tracked v1 orchestration policy.
package orchestrationconfig

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	SchemaVersion = 1
	MaxBytes      = 256 << 10
	relativePath  = ".l7/orchestration.json"
)

var (
	idPattern          = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	environmentPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,127}$`)
	digestPattern      = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type File struct {
	Schema    int        `json:"schema"`
	Features  Features   `json:"features"`
	Providers []Provider `json:"providers"`
	Routing   Routing    `json:"routing"`
	Tools     Tools      `json:"tools"`
	Memory    Memory     `json:"memory"`
	Cyber     Cyber      `json:"cyber"`
	Headless  Headless   `json:"headless"`
}

type Features struct {
	Orchestration bool `json:"orchestration"`
	Sync          bool `json:"sync"`
	CyberActive   bool `json:"cyber_active"`
	Headless      bool `json:"headless"`
}

type Provider struct {
	ID         string              `json:"id"`
	Kind       domain.ProviderKind `json:"kind"`
	Enabled    bool                `json:"enabled"`
	Endpoint   string              `json:"endpoint"`
	CatalogURL string              `json:"catalog_url"`
	Credential Credential          `json:"credential"`
	Models     []Model             `json:"models"`
}

type Credential struct {
	Source    string `json:"source"`
	Reference string `json:"reference"`
}

type Model struct {
	ID              string                   `json:"id"`
	Languages       []string                 `json:"languages"`
	ContextWindow   int                      `json:"context_window"`
	SupportsTools   bool                     `json:"supports_tools"`
	SupportsEditing bool                     `json:"supports_editing"`
	SupportsResume  bool                     `json:"supports_resume"`
	Efforts         []domain.ReasoningEffort `json:"efforts"`
	CostClass       int                      `json:"cost_class"`
	LatencyClass    int                      `json:"latency_class"`
}

type Routing struct {
	Policy                  string `json:"policy"`
	RequireIndependentAudit bool   `json:"require_independent_audit"`
}

type Tools struct {
	AllowedPaths    []string  `json:"allowed_paths"`
	AllowedCommands []Command `json:"allowed_commands"`
	MaxOutputBytes  int       `json:"max_output_bytes"`
	MaxSeconds      int       `json:"max_seconds"`
}

type Command struct {
	Name string   `json:"name"`
	Argv []string `json:"argv"`
}

type Memory struct {
	EmbeddingProvider string   `json:"embedding_provider"`
	Exclude           []string `json:"exclude"`
	MaxFileBytes      int      `json:"max_file_bytes"`
}

type Cyber struct {
	Runtime           string `json:"runtime"`
	Image             string `json:"image"`
	ImageDigest       string `json:"image_digest"`
	SignatureIdentity string `json:"signature_identity"`
	SignatureIssuer   string `json:"signature_issuer"`
	NetworkPolicy     string `json:"network_policy"`
}

type Headless struct {
	LocalMerge      bool `json:"local_merge"`
	RiskCeiling     int  `json:"risk_ceiling"`
	NoProgressLimit int  `json:"no_progress_limit"`
}

func Default() File {
	return File{
		Schema:   SchemaVersion,
		Features: Features{},
		Providers: []Provider{
			{ID: "codex-local", Kind: domain.ProviderKindCodexAppServer, Enabled: true, Models: []Model{}},
			{ID: "claude-local", Kind: domain.ProviderKindClaudeCLI, Enabled: true, Models: []Model{}},
		},
		Routing: Routing{Policy: "balanced", RequireIndependentAudit: true},
		Tools:   Tools{AllowedPaths: []string{}, AllowedCommands: []Command{}, MaxOutputBytes: 8 << 20, MaxSeconds: 1800},
		Memory: Memory{EmbeddingProvider: "apple_natural_language", Exclude: []string{
			".git/**", ".env", ".env.*", ".transcripts/**", "transcripts/**", "*.transcript", "*.chatlog",
			"node_modules/**", "vendor/**", "build/**", "dist/**",
		}, MaxFileBytes: 2 << 20},
		Cyber: Cyber{
			Runtime: "docker", Image: "ghcr.io/addressanup/level7-cyber:v1.0.0", ImageDigest: "",
			SignatureIdentity: "https://github.com/addressanup/level7-dev-loop/.github/workflows/release.yml@refs/tags/v1.0.0",
			SignatureIssuer:   "https://token.actions.githubusercontent.com", NetworkPolicy: "internal-only",
		},
		Headless: Headless{LocalMerge: false, RiskCeiling: 2, NoProgressLimit: 3},
	}
}

func AppliedDefault() File {
	configuration := Default()
	configuration.Features.Orchestration = true
	configuration.Features.Sync = true
	return configuration
}

func Path(root string) string { return filepath.Join(root, filepath.FromSlash(relativePath)) }

func Load(root string) (File, error) {
	var configuration File
	data, err := localfile.Read(Path(root), MaxBytes)
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

func Create(root string, applied bool) (File, bool, error) {
	if !filepath.IsAbs(root) {
		return File{}, false, errors.New("repository root must be absolute")
	}
	if existing, err := Load(root); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return File{}, false, err
	}
	configuration := Default()
	if applied {
		configuration = AppliedDefault()
	}
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		return File{}, false, err
	}
	directory := filepath.Join(root, ".l7")
	if err := localfile.EnsureDirectory(directory, 0o755); err != nil {
		return File{}, false, err
	}
	if err := localfile.AtomicCreate(Path(root), data, 0o644); err != nil {
		return File{}, false, err
	}
	return configuration, true, nil
}

// Apply explicitly enables the safe v1 starting features while preserving all
// user-configured provider, routing, tool, Cyber, and Headless policy. It is
// intentionally only called by the mutating Onboard action.
func Apply(root string) (File, bool, error) {
	configuration, err := Load(root)
	if errors.Is(err, os.ErrNotExist) {
		return Create(root, true)
	}
	if err != nil {
		return File{}, false, err
	}
	if configuration.Features.Orchestration && configuration.Features.Sync {
		return configuration, false, nil
	}
	configuration.Features.Orchestration = true
	configuration.Features.Sync = true
	if err := configuration.Validate(); err != nil {
		return File{}, false, err
	}
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		return File{}, false, err
	}
	if err := localfile.AtomicReplace(Path(root), data, 0o644); err != nil {
		return File{}, false, err
	}
	return configuration, true, nil
}

func (configuration File) Validate() error {
	if configuration.Schema != SchemaVersion {
		return fmt.Errorf("unsupported orchestration schema %d", configuration.Schema)
	}
	if configuration.Routing.Policy != "balanced" {
		return errors.New("routing policy must be balanced")
	}
	if len(configuration.Providers) > 32 {
		return errors.New("provider count exceeds 32")
	}
	seenProviders := make(map[string]bool)
	for _, provider := range configuration.Providers {
		if !idPattern.MatchString(provider.ID) || seenProviders[provider.ID] || !provider.Kind.Valid() {
			return fmt.Errorf("provider identity %q is invalid or duplicated", provider.ID)
		}
		seenProviders[provider.ID] = true
		if err := validateProvider(provider); err != nil {
			return fmt.Errorf("provider %q: %w", provider.ID, err)
		}
	}
	if len(configuration.Tools.AllowedPaths) > 256 || len(configuration.Tools.AllowedCommands) > 64 ||
		configuration.Tools.MaxOutputBytes < 64<<10 || configuration.Tools.MaxOutputBytes > 64<<20 ||
		configuration.Tools.MaxSeconds < 1 || configuration.Tools.MaxSeconds > 86_400 {
		return errors.New("tool policy bounds are invalid")
	}
	for _, relative := range configuration.Tools.AllowedPaths {
		if !safeRelativePattern(relative) {
			return fmt.Errorf("unsafe allowed path %q", relative)
		}
	}
	seenCommands := make(map[string]bool)
	for _, command := range configuration.Tools.AllowedCommands {
		if !idPattern.MatchString(command.Name) || seenCommands[command.Name] || len(command.Argv) == 0 || len(command.Argv) > 64 {
			return fmt.Errorf("allowed command %q is invalid or duplicated", command.Name)
		}
		seenCommands[command.Name] = true
		for _, argument := range command.Argv {
			if !safeText(argument, 4096) {
				return fmt.Errorf("allowed command %q contains an unsafe argument", command.Name)
			}
		}
	}
	if configuration.Memory.EmbeddingProvider != "apple_natural_language" || len(configuration.Memory.Exclude) > 256 ||
		configuration.Memory.MaxFileBytes < 4096 || configuration.Memory.MaxFileBytes > 16<<20 {
		return errors.New("memory policy is invalid")
	}
	for _, pattern := range configuration.Memory.Exclude {
		if !safeRelativePattern(pattern) {
			return fmt.Errorf("unsafe memory exclusion %q", pattern)
		}
	}
	if configuration.Cyber.Runtime != "docker" || configuration.Cyber.NetworkPolicy != "internal-only" ||
		!safeImage(configuration.Cyber.Image) || !safeText(configuration.Cyber.SignatureIdentity, 1024) || !safeText(configuration.Cyber.SignatureIssuer, 1024) ||
		(configuration.Cyber.ImageDigest != "" && !digestPattern.MatchString(configuration.Cyber.ImageDigest)) {
		return errors.New("Cyber isolation policy is invalid")
	}
	if configuration.Features.CyberActive && configuration.Cyber.ImageDigest == "" {
		return errors.New("active Cyber requires a pinned sha256 image digest")
	}
	if configuration.Headless.RiskCeiling != 2 || configuration.Headless.NoProgressLimit != 3 {
		return errors.New("Headless safety policy must keep the Tier 2 ceiling and three-failure pause")
	}
	return nil
}

func validateProvider(provider Provider) error {
	isGateway := provider.Kind == domain.ProviderKindOpenAIResponses || provider.Kind == domain.ProviderKindAnthropic
	if !isGateway {
		if provider.Endpoint != "" || provider.CatalogURL != "" || provider.Credential.Source != "" || provider.Credential.Reference != "" || len(provider.Models) != 0 {
			return errors.New("native host cannot declare gateway settings")
		}
		return nil
	}
	if !safeEndpoint(provider.Endpoint) || (provider.CatalogURL != "" && !safeEndpoint(provider.CatalogURL)) {
		return errors.New("gateway endpoint is invalid")
	}
	if provider.Credential.Source != "env" && provider.Credential.Source != "keychain" {
		return errors.New("credential source must be env or keychain")
	}
	if provider.Credential.Source == "env" && !environmentPattern.MatchString(provider.Credential.Reference) {
		return errors.New("environment credential reference is invalid")
	}
	if provider.Credential.Source == "keychain" && (!safeText(provider.Credential.Reference, 256) || !strings.Contains(provider.Credential.Reference, "/")) {
		return errors.New("Keychain credential reference must be service/account")
	}
	if (len(provider.Models) == 0 && provider.CatalogURL == "") || len(provider.Models) > 128 {
		return errors.New("gateway must declare models or a bounded catalog endpoint")
	}
	seen := make(map[string]bool)
	for _, model := range provider.Models {
		if seen[model.ID] {
			return fmt.Errorf("model %q is invalid or duplicated", model.ID)
		}
		if err := ValidateModel(model); err != nil {
			return err
		}
		seen[model.ID] = true
	}
	return nil
}

// ValidateModel applies the same fail-closed capability contract to configured
// and explicitly discovered gateway catalog entries.
func ValidateModel(model Model) error {
	if !safeText(model.ID, 160) || len(model.Languages) == 0 || len(model.Languages) > 128 || model.ContextWindow < 1024 || model.ContextWindow > 16_000_000 ||
		model.CostClass < 1 || model.CostClass > 5 || model.LatencyClass < 1 || model.LatencyClass > 5 || len(model.Efforts) == 0 || len(model.Efforts) > 8 {
		return fmt.Errorf("model %q is invalid", model.ID)
	}
	seenLanguages := make(map[string]bool)
	for _, language := range model.Languages {
		normalized := strings.ToLower(strings.TrimSpace(language))
		if !safeText(normalized, 64) || seenLanguages[normalized] || (normalized != "*" && !idPattern.MatchString(normalized)) {
			return fmt.Errorf("model %q has invalid or duplicate language %q", model.ID, language)
		}
		seenLanguages[normalized] = true
	}
	efforts := make(map[domain.ReasoningEffort]bool)
	for _, effort := range model.Efforts {
		if !effort.Valid() || efforts[effort] {
			return fmt.Errorf("model %q has invalid or duplicate effort", model.ID)
		}
		efforts[effort] = true
	}
	return nil
}

func safeEndpoint(value string) bool {
	if !safeText(value, 2048) || strings.ContainsAny(value, "?#\\") {
		return false
	}
	scheme := ""
	remainder := ""
	if strings.HasPrefix(value, "https://") {
		scheme, remainder = "https", strings.TrimPrefix(value, "https://")
	} else if strings.HasPrefix(value, "http://") {
		scheme, remainder = "http", strings.TrimPrefix(value, "http://")
	} else {
		return false
	}
	authority, _, _ := strings.Cut(remainder, "/")
	host, ok := safeURLAuthority(authority)
	if !ok {
		return false
	}
	if scheme == "https" {
		return true
	}
	normalized := strings.ToLower(host)
	return normalized == "localhost" || normalized == "127.0.0.1" || normalized == "::1"
}

func safeURLAuthority(authority string) (string, bool) {
	if authority == "" || len(authority) > 512 || strings.ContainsAny(authority, "@%") {
		return "", false
	}
	host := authority
	port := ""
	if strings.HasPrefix(authority, "[") {
		closing := strings.IndexByte(authority, ']')
		if closing < 0 || strings.Contains(authority[closing+1:], "]") {
			return "", false
		}
		host = authority[1:closing]
		rest := authority[closing+1:]
		if rest != "" {
			if !strings.HasPrefix(rest, ":") || len(rest) == 1 {
				return "", false
			}
			port = rest[1:]
		}
		if !safeIPv6(host) {
			return "", false
		}
	} else {
		if strings.ContainsAny(authority, "[]") || strings.Count(authority, ":") > 1 {
			return "", false
		}
		if value, found := strings.CutSuffix(authority, ":"); found {
			_ = value
			return "", false
		}
		if value, suffix, found := strings.Cut(authority, ":"); found {
			host, port = value, suffix
		}
		if !safeHostnameOrIPv4(host) {
			return "", false
		}
	}
	if port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 || strconv.Itoa(value) != port {
			return "", false
		}
	}
	return host, true
}

func safeHostnameOrIPv4(host string) bool {
	if host == "" || len(host) > 253 || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}
	labels := strings.Split(host, ".")
	if len(labels) == 4 {
		ipv4 := true
		numeric := true
		for _, label := range labels {
			value, err := strconv.Atoi(label)
			if err != nil || value < 0 || value > 255 || strconv.Itoa(value) != label {
				ipv4 = false
			}
			for _, character := range label {
				if character < '0' || character > '9' {
					numeric = false
				}
			}
		}
		if ipv4 {
			return true
		}
		if numeric {
			return false
		}
	} else {
		numeric := true
		for _, character := range host {
			if character < '0' || character > '9' {
				numeric = false
				break
			}
		}
		if numeric {
			return false
		}
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') &&
				(character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

func safeIPv6(host string) bool {
	if host == "" || strings.ContainsAny(host, ".%") || strings.Contains(host, ":::") || strings.Count(host, "::") > 1 {
		return false
	}
	double := strings.Contains(host, "::")
	parts := strings.Split(host, ":")
	if double {
		left, right, _ := strings.Cut(host, "::")
		parts = nil
		if left != "" {
			parts = append(parts, strings.Split(left, ":")...)
		}
		if right != "" {
			parts = append(parts, strings.Split(right, ":")...)
		}
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		if len(part) > 4 {
			return false
		}
		for _, character := range part {
			if (character < '0' || character > '9') && (character < 'a' || character > 'f') && (character < 'A' || character > 'F') {
				return false
			}
		}
	}
	if double {
		return len(parts) < 8
	}
	return len(parts) == 8
}

func safeImage(value string) bool {
	return len(value) >= 3 && len(value) <= 512 && safeText(value, 512) && !strings.ContainsAny(value, " '@\"")
}

func safeRelativePattern(value string) bool {
	if !safeText(value, 1024) || value == "" || filepath.IsAbs(value) || strings.HasPrefix(value, "../") || strings.Contains(value, "/../") || strings.Contains(value, "\\") {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(value))
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}

func safeText(value string, limit int) bool {
	return value != "" && len(value) <= limit && !strings.ContainsAny(value, "\x00\r\n")
}

func Digest(configuration File) (string, error) {
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum), nil
}
