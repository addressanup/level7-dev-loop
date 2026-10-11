// Package pluginhost drives the Codex and Claude Code plugin managers through
// their own command-line interfaces. It never edits host configuration files
// directly, and every mutation is confirmed through the host's list commands.
package pluginhost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

type Host string

const (
	Codex  Host = "codex"
	Claude Host = "claude"

	// Scope is the only installation scope Level 7 manages.
	Scope = "user"

	maxOutputBytes = 4 << 20
	commandTimeout = 3 * time.Minute
	probeTimeout   = 30 * time.Second
)

// Minimum host versions are the adapter fixture versions recorded in
// distribution/compatibility.json: the oldest hosts whose plugin commands
// Level 7 has observed. Newer hosts are accepted with a status note.
var minimumVersions = map[Host]string{Codex: "0.149.1", Claude: "2.1.241"}

func Parse(value string) (Host, error) {
	switch Host(value) {
	case Codex, Claude:
		return Host(value), nil
	default:
		return "", fmt.Errorf("host must be codex or claude, got %q", value)
	}
}

func All() []Host { return []Host{Claude, Codex} }

func (host Host) String() string { return string(host) }

// ReloadAdvice tells the user how the running host picks up a change; the
// installer never claims a live session already sees the new plugin.
func (host Host) ReloadAdvice() string {
	if host == Claude {
		return "start a new Claude Code session or run /reload-plugins; the currently running session has not loaded this change"
	}
	return "start a new Codex task; the currently running Codex session has not loaded this change"
}

func (host Host) executableName() string { return string(host) }

type ResolveFunc func(string) (processadapter.Executable, error)
type RunFunc func(context.Context, processadapter.Request) (processadapter.Result, error)

// Adapter is one host's command translator.
type Adapter struct {
	host    Host
	resolve ResolveFunc
	run     RunFunc
}

func New(host Host, resolve ResolveFunc, run RunFunc) Adapter {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	return Adapter{host: host, resolve: resolve, run: run}
}

func (adapter Adapter) Host() Host { return adapter.host }

// Executable locates the host CLI with an actionable error when absent.
func (adapter Adapter) Executable() (processadapter.Executable, error) {
	executable, err := adapter.resolve(adapter.host.executableName())
	if err != nil {
		product := "Codex CLI"
		if adapter.host == Claude {
			product = "Claude Code"
		}
		return processadapter.Executable{}, fmt.Errorf("%s executable %q was not found on PATH; install %s yourself (Level 7 never installs hosts) or skip this host", product, adapter.host.executableName(), product)
	}
	return executable, nil
}

// Version probes the host and enforces the minimum known plugin-capable
// version. It returns the host's own version string for status output.
func (adapter Adapter) Version(ctx context.Context, executable processadapter.Executable) (string, error) {
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: []string{"--version"}, Directory: "/",
		Environment: hostEnvironment(), MaxOutputBytes: 64 << 10, Timeout: probeTimeout,
	})
	if err != nil || result.ExitCode != 0 {
		return "", fmt.Errorf("%s --version failed; repair the host installation before installing the plugin", adapter.host)
	}
	display, version, err := parseVersion(adapter.host, string(result.Stdout))
	if err != nil {
		return "", err
	}
	if compareVersions(version, minimumVersions[adapter.host]) < 0 {
		return "", fmt.Errorf("%s %s is older than the minimum plugin-capable version %s; update the host first", adapter.host, version, minimumVersions[adapter.host])
	}
	return display, nil
}

func parseVersion(host Host, output string) (string, string, error) {
	line, _, _ := strings.Cut(strings.TrimSpace(output), "\n")
	fields := strings.Fields(line)
	if len(fields) == 0 || !utf8.ValidString(line) || len(line) > 256 {
		return "", "", fmt.Errorf("%s --version output is unrecognized", host)
	}
	version := fields[0]
	if host == Codex {
		if len(fields) < 2 || fields[0] != "codex-cli" {
			return "", "", errors.New("codex --version output is unrecognized; expected \"codex-cli X.Y.Z\"")
		}
		version = fields[1]
	}
	version = strings.TrimPrefix(version, "v")
	if !plainVersion(version) {
		return "", "", fmt.Errorf("%s --version reports an unrecognized version %q", host, version)
	}
	return line, version, nil
}

func plainVersion(value string) bool {
	parts := strings.Split(strings.SplitN(value, "-", 2)[0], ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" || len(part) > 9 {
			return false
		}
		for _, character := range part {
			if character < '0' || character > '9' {
				return false
			}
		}
	}
	return true
}

func compareVersions(left, right string) int {
	leftParts := strings.Split(strings.SplitN(left, "-", 2)[0], ".")
	rightParts := strings.Split(strings.SplitN(right, "-", 2)[0], ".")
	for index := 0; index < 3; index++ {
		leftValue, _ := strconv.Atoi(leftParts[index])
		rightValue, _ := strconv.Atoi(rightParts[index])
		if leftValue != rightValue {
			if leftValue < rightValue {
				return -1
			}
			return 1
		}
	}
	return 0
}

// MarketplaceAdd registers a verified extraction root as a local marketplace.
func (adapter Adapter) MarketplaceAdd(ctx context.Context, executable processadapter.Executable, root string) error {
	if !filepath.IsAbs(root) {
		return errors.New("marketplace root must be absolute")
	}
	arguments := []string{"plugin", "marketplace", "add", root, "--json"}
	if adapter.host == Claude {
		arguments = []string{"plugin", "marketplace", "add", root, "--scope", Scope, "--json"}
	}
	return adapter.mutate(ctx, executable, "marketplace add", arguments)
}

func (adapter Adapter) MarketplaceRemove(ctx context.Context, executable processadapter.Executable, marketplace string) error {
	arguments := []string{"plugin", "marketplace", "remove", marketplace, "--json"}
	if adapter.host == Claude {
		arguments = []string{"plugin", "marketplace", "remove", marketplace, "--scope", Scope, "--json"}
	}
	return adapter.mutate(ctx, executable, "marketplace remove", arguments)
}

func (adapter Adapter) PluginAdd(ctx context.Context, executable processadapter.Executable, plugin, marketplace string) error {
	arguments := []string{"plugin", "add", plugin + "@" + marketplace, "--json"}
	if adapter.host == Claude {
		arguments = []string{"plugin", "install", plugin + "@" + marketplace, "--scope", Scope, "--json"}
	}
	return adapter.mutate(ctx, executable, "plugin install", arguments)
}

// PluginRemove uninstalls the plugin. Claude's persistent plugin data
// directory is always kept; deleting user data is not an installer decision.
func (adapter Adapter) PluginRemove(ctx context.Context, executable processadapter.Executable, plugin, marketplace string) error {
	arguments := []string{"plugin", "remove", plugin + "@" + marketplace, "--json"}
	if adapter.host == Claude {
		arguments = []string{"plugin", "uninstall", plugin + "@" + marketplace, "--scope", Scope, "--keep-data", "--json"}
	}
	return adapter.mutate(ctx, executable, "plugin uninstall", arguments)
}

func (adapter Adapter) mutate(ctx context.Context, executable processadapter.Executable, operation string, arguments []string) error {
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: arguments, Directory: "/",
		Environment: hostEnvironment(), MaxOutputBytes: maxOutputBytes, Timeout: commandTimeout,
	})
	if err != nil {
		return fmt.Errorf("%s %s did not complete: %w", adapter.host, operation, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("%s %s exited %d: %s", adapter.host, operation, result.ExitCode, diagnostic(append(result.Stderr, result.Stdout...)))
	}
	return nil
}

// Observation is what the host itself reports about Level 7's registration.
type Observation struct {
	MarketplaceRegistered bool   `json:"marketplace_registered"`
	MarketplaceRoot       string `json:"marketplace_root,omitempty"`
	PluginInstalled       bool   `json:"plugin_installed"`
	PluginVersion         string `json:"plugin_version,omitempty"`
	PluginEnabled         bool   `json:"plugin_enabled"`
	PluginPath            string `json:"plugin_path,omitempty"`
}

// Observe reads the host's marketplace and plugin lists. Both commands are
// local; Codex's plugin list is scoped to the marketplace so no remote
// catalog is contacted.
func (adapter Adapter) Observe(ctx context.Context, executable processadapter.Executable, marketplace, plugin string) (Observation, error) {
	data, err := adapter.query(ctx, executable, []string{"plugin", "marketplace", "list", "--json"})
	if err != nil {
		return Observation{}, err
	}
	observation := Observation{}
	if adapter.host == Codex {
		var listing struct {
			Marketplaces []struct {
				Name string `json:"name"`
				Root string `json:"root"`
			} `json:"marketplaces"`
		}
		if err := json.Unmarshal(data, &listing); err != nil {
			return Observation{}, errors.New("codex marketplace list is malformed")
		}
		for _, entry := range listing.Marketplaces {
			if entry.Name == marketplace {
				observation.MarketplaceRegistered, observation.MarketplaceRoot = true, entry.Root
			}
		}
	} else {
		var listing []struct {
			Name            string `json:"name"`
			Path            string `json:"path"`
			InstallLocation string `json:"installLocation"`
		}
		if err := json.Unmarshal(data, &listing); err != nil {
			return Observation{}, errors.New("claude marketplace list is malformed")
		}
		for _, entry := range listing {
			if entry.Name == marketplace {
				observation.MarketplaceRegistered = true
				observation.MarketplaceRoot = entry.Path
				if observation.MarketplaceRoot == "" {
					observation.MarketplaceRoot = entry.InstallLocation
				}
			}
		}
	}
	if adapter.host == Codex {
		if !observation.MarketplaceRegistered {
			return observation, nil
		}
		data, err = adapter.query(ctx, executable, []string{"plugin", "list", "--json", "--marketplace", marketplace})
		if err != nil {
			return Observation{}, err
		}
		var listing struct {
			Installed []struct {
				PluginID string `json:"pluginId"`
				Version  string `json:"version"`
				Enabled  bool   `json:"enabled"`
				Source   struct {
					Path string `json:"path"`
				} `json:"source"`
			} `json:"installed"`
		}
		if err := json.Unmarshal(data, &listing); err != nil {
			return Observation{}, errors.New("codex plugin list is malformed")
		}
		for _, entry := range listing.Installed {
			if entry.PluginID == plugin+"@"+marketplace {
				observation.PluginInstalled, observation.PluginVersion, observation.PluginEnabled, observation.PluginPath = true, entry.Version, entry.Enabled, entry.Source.Path
			}
		}
		return observation, nil
	}
	data, err = adapter.query(ctx, executable, []string{"plugin", "list", "--json"})
	if err != nil {
		return Observation{}, err
	}
	var listing []struct {
		ID          string `json:"id"`
		Version     string `json:"version"`
		Scope       string `json:"scope"`
		Enabled     bool   `json:"enabled"`
		InstallPath string `json:"installPath"`
	}
	if err := json.Unmarshal(data, &listing); err != nil {
		return Observation{}, errors.New("claude plugin list is malformed")
	}
	for _, entry := range listing {
		if entry.ID == plugin+"@"+marketplace && entry.Scope == Scope {
			observation.PluginInstalled, observation.PluginVersion, observation.PluginEnabled, observation.PluginPath = true, entry.Version, entry.Enabled, entry.InstallPath
		}
	}
	return observation, nil
}

func (adapter Adapter) query(ctx context.Context, executable processadapter.Executable, arguments []string) ([]byte, error) {
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: arguments, Directory: "/",
		Environment: hostEnvironment(), MaxOutputBytes: maxOutputBytes, Timeout: commandTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%s %s did not complete: %w", adapter.host, strings.Join(arguments[:3], " "), err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("%s %s exited %d: %s", adapter.host, strings.Join(arguments[:3], " "), result.ExitCode, diagnostic(result.Stderr))
	}
	data := lastJSONValue(result.Stdout)
	if data == nil {
		return nil, fmt.Errorf("%s %s returned no JSON", adapter.host, strings.Join(arguments[:3], " "))
	}
	return data, nil
}

// lastJSONValue tolerates human lines printed before a JSON document and
// returns the final complete JSON value in the output.
func lastJSONValue(output []byte) []byte {
	trimmed := strings.TrimSpace(string(output))
	if json.Valid([]byte(trimmed)) {
		return []byte(trimmed)
	}
	offset := 0
	for offset < len(trimmed) {
		candidate := trimmed[offset:]
		if (candidate[0] == '{' || candidate[0] == '[') && json.Valid([]byte(candidate)) {
			return []byte(candidate)
		}
		next := strings.IndexByte(candidate, '\n')
		if next < 0 {
			break
		}
		offset += next + 1
	}
	return nil
}

// hostEnvironment keeps the minimal environment plus the hosts' own
// configuration-location variables so an isolated profile can be targeted.
func hostEnvironment() []string {
	values := processadapter.MinimalEnvironment()
	for _, key := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME"} {
		if value := os.Getenv(key); value != "" && !strings.ContainsAny(value, "\x00\r\n") {
			values = append(values, key+"="+value)
		}
	}
	return values
}

func diagnostic(output []byte) string {
	message := strings.Join(strings.Fields(strings.ToValidUTF8(string(output), "")), " ")
	if len(message) > 512 {
		cut := 512
		for cut > 0 && !utf8.RuneStart(message[cut]) {
			cut--
		}
		message = message[:cut]
	}
	if message == "" {
		return "no diagnostic"
	}
	return message
}
