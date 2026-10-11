package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/plugininstall"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

const pluginUsage = `Usage: l7 plugin <command> [options] [--json]

Commands:
  install   --host codex|claude [--channel stable|prerelease] [--version X.Y.Z]
            [--accept-unsigned] [--accept-unqualified] [--dry-run]
  update    (--all | --host codex|claude) [--channel stable|prerelease] [--version X.Y.Z]
            [--unpin] [--allow-downgrade] [--accept-unsigned] [--accept-unqualified] [--dry-run]
  status    [--check]
  rollback  --host codex|claude [--dry-run]
  uninstall --host codex|claude [--yes] [--dry-run]

Install and update download one GitHub release of addressanup/level7-dev-loop over HTTPS,
verify SHA-256 digests and GitHub artifact attestations with gh (>= 2.67.0, no login),
extract into a new versioned directory under $L7_HOME (default ~/.local/share/level7),
and register it with the host's own plugin commands. Stable is the default channel;
prereleases, exact versions, downgrades, unsigned binaries, and withheld-support
releases each need their own flag. status, rollback, and uninstall work offline.
A new host session is needed before a changed plugin is loaded.`

var pluginFlagsWithValue = map[string]bool{"--host": true, "--channel": true, "--version": true}

var pluginFlagsBySubcommand = map[string]map[string]bool{
	"install":   {"--host": true, "--channel": true, "--version": true, "--accept-unsigned": true, "--accept-unqualified": true, "--dry-run": true},
	"update":    {"--all": true, "--host": true, "--channel": true, "--version": true, "--unpin": true, "--allow-downgrade": true, "--accept-unsigned": true, "--accept-unqualified": true, "--dry-run": true},
	"status":    {"--check": true},
	"rollback":  {"--host": true, "--dry-run": true},
	"uninstall": {"--host": true, "--yes": true, "--dry-run": true},
}

func runPluginCommand(ctx context.Context, arguments []string, input io.Reader, stdout, stderr io.Writer, jsonOutput bool) int {
	if len(arguments) == 0 || arguments[0] == "--help" || arguments[0] == "-h" || arguments[0] == "help" {
		envelope := passEnvelope("plugin", "L7-PLUGIN-000", "help", "Level 7 plugin installer", "run l7 plugin status", strings.Split(pluginUsage, "\n"))
		return writeOrchestration(stdout, stderr, jsonOutput, envelope)
	}
	subcommand := arguments[0]
	allowed, ok := pluginFlagsBySubcommand[subcommand]
	if !ok {
		return writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope("plugin", "L7-PLUGIN-001", "invalid", "unknown plugin command "+subcommand, "run l7 plugin --help"))
	}
	request, err := parsePluginRequest(arguments[1:], allowed)
	if err != nil {
		return writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope("plugin "+subcommand, "L7-PLUGIN-001", "invalid", err.Error(), "run l7 plugin --help"))
	}
	manager, err := newPluginManager(input, stderr)
	if err != nil {
		return writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope("plugin "+subcommand, "L7-PLUGIN-003", "failed", err.Error(), "repair the installer home and retry"))
	}
	var report plugininstall.Report
	switch subcommand {
	case "install":
		report, err = manager.Install(ctx, request)
	case "update":
		report, err = manager.Update(ctx, request)
	case "status":
		report, err = manager.Status(ctx, request)
	case "rollback":
		report, err = manager.Rollback(ctx, request)
	case "uninstall":
		report, err = manager.Uninstall(ctx, request)
	}
	command := "plugin " + subcommand
	if err != nil {
		var blocked *plugininstall.Blocked
		if errors.As(err, &blocked) {
			envelope := failedEnvelope(command, "L7-PLUGIN-002", "blocked", blocked.Reason, blocked.Next)
			envelope.Outcome = "BLOCKED"
			return writeOrchestration(stdout, stderr, jsonOutput, envelope)
		}
		return writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope(command, "L7-PLUGIN-003", "failed", err.Error(), "nothing was left half-changed; read the error, repair the cause, and rerun l7 plugin status"))
	}
	envelope := passEnvelope(command, "L7-PLUGIN-000", report.State, report.Message, report.Next, report)
	switch report.Outcome {
	case plugininstall.OutcomeBlocked:
		envelope.Outcome, envelope.Code = "BLOCKED", "L7-PLUGIN-002"
	case plugininstall.OutcomeFailed:
		envelope.Outcome, envelope.Code = "FAILED", "L7-PLUGIN-003"
	}
	return writeOrchestration(stdout, stderr, jsonOutput, envelope)
}

func parsePluginRequest(arguments []string, allowed map[string]bool) (plugininstall.Request, error) {
	request := plugininstall.Request{}
	seen := map[string]bool{}
	for index := 0; index < len(arguments); index++ {
		flag := arguments[index]
		if !allowed[flag] {
			return request, fmt.Errorf("unknown or unsupported option %s", flag)
		}
		if seen[flag] {
			return request, fmt.Errorf("duplicate option %s", flag)
		}
		seen[flag] = true
		value := ""
		if pluginFlagsWithValue[flag] {
			if index+1 >= len(arguments) || strings.HasPrefix(arguments[index+1], "-") {
				return request, fmt.Errorf("option %s is missing its value", flag)
			}
			value = arguments[index+1]
			index++
		}
		switch flag {
		case "--host":
			request.Host = value
		case "--channel":
			request.Channel = value
		case "--version":
			request.Version = value
		case "--all":
			request.All = true
		case "--unpin":
			request.Unpin = true
		case "--allow-downgrade":
			request.AllowDowngrade = true
		case "--accept-unsigned":
			request.AcceptUnsigned = true
		case "--accept-unqualified":
			request.AcceptUnqualified = true
		case "--dry-run":
			request.DryRun = true
		case "--yes":
			request.Yes = true
		case "--check":
			request.Check = true
		}
	}
	return request, nil
}

// newPluginManager composes the production installer: a user-owned home,
// the running platform, the fixed-repository release client (disabled under
// L7_NETWORK=off), the real host adapters, and a terminal confirmation only
// when both stdin and stderr are terminals.
func newPluginManager(input io.Reader, stderr io.Writer) (*plugininstall.Manager, error) {
	home := os.Getenv("L7_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.New("cannot determine the home directory; set L7_HOME to an absolute path")
		}
		home = filepath.Join(userHome, ".local", "share", "level7")
	}
	if !filepath.IsAbs(home) {
		return nil, errors.New("L7_HOME must be an absolute path")
	}
	options := plugininstall.Options{
		Home:       home,
		Platform:   plugininstall.Platform{OS: runtime.GOOS, Arch: runtime.GOARCH},
		CLIVersion: version,
	}
	if os.Getenv("L7_NETWORK") != "off" {
		options.Releases = pluginrelease.NewClient(nil, pluginrelease.DefaultSource(), pluginrelease.NewVerifier(nil, nil), version)
	}
	if terminalAvailable(input) && terminalAvailable(stderr) {
		reader := bufio.NewReader(input)
		options.Confirm = func(prompt string) (bool, error) {
			if _, err := io.WriteString(stderr, prompt); err != nil {
				return false, err
			}
			line, err := reader.ReadString('\n')
			if err != nil && line == "" {
				return false, err
			}
			answer := strings.ToLower(strings.TrimSpace(line))
			return answer == "y" || answer == "yes", nil
		}
	}
	return plugininstall.New(options)
}
