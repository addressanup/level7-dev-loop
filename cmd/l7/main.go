package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	configadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/config"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	stateadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
	cliapp "github.com/addressanup/level7-dev-loop/internal/l7/app"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
	"github.com/addressanup/level7-dev-loop/internal/l7/presentation"
)

var version = "0.1.0-dev"

const (
	maxArguments     = 1024
	maxArgumentBytes = 4096
	maxInputBytes    = 1 << 20
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, arguments []string, stdout, stderr io.Writer) int {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "FAILED code=L7-CLI-004 message=%q\n", "cannot determine working directory")
		return 1
	}
	return runAt(ctx, arguments, workingDirectory, stdout, stderr)
}

func runAt(ctx context.Context, arguments []string, workingDirectory string, stdout, stderr io.Writer) int {
	baseApplication := cliapp.New(version)
	request, jsonOutput, invalid := parseArguments(arguments, baseApplication)
	application, compositionErr := lifecycleApplication(workingDirectory)
	result := domain.Result{}
	if invalid != nil {
		result = *invalid
	} else if compositionErr != nil && request.Command != domain.CommandHelp && request.Command != domain.CommandVersion {
		result = domain.Result{
			Schema: domain.ResultSchema, Outcome: domain.OutcomeFailed, Code: "L7-CLI-005", Command: string(request.Command), State: "unavailable", Version: version,
			Message: "cannot initialize Git adapter", Next: "install Git and run l7 help",
		}
	} else {
		if compositionErr != nil {
			application = baseApplication
		}
		result = application.ExecuteRequest(ctx, request)
	}

	var output []byte
	if jsonOutput {
		var err error
		output, err = presentation.JSON(result)
		if err != nil {
			fmt.Fprintf(stderr, "FAILED code=L7-CLI-002 message=%q\n", "cannot render JSON output")
			return 1
		}
	} else {
		output = presentation.Text(result)
	}
	if written, err := stdout.Write(output); err != nil || written != len(output) {
		fmt.Fprintf(stderr, "FAILED code=L7-CLI-002 message=%q\n", "cannot write command output")
		return 1
	}
	return result.ExitCode()
}

func lifecycleApplication(workingDirectory string) (cliapp.Application, error) {
	gitClient, err := gitadapter.New("", gitadapter.DefaultMaxOutput, gitadapter.DefaultMaxPaths)
	if err != nil {
		return cliapp.New(version), err
	}
	ports := cliapp.Ports{
		Locate: gitClient.Locate,
		Snapshot: func(ctx context.Context, cwd, base string, maxOutput, maxPaths int) (domain.RepositorySnapshot, error) {
			return gitClient.WithLimits(maxOutput, maxPaths).Snapshot(ctx, cwd, base)
		},
		LoadConfiguration: func(root string) (domain.Configuration, bool, error) {
			configuration, loadErr := configadapter.Load(root)
			if errors.Is(loadErr, os.ErrNotExist) {
				return domain.Configuration{}, false, nil
			}
			if loadErr != nil {
				return domain.Configuration{}, false, loadErr
			}
			return configuration.Domain(), true, nil
		},
		AdoptConfiguration: func(root string, enable bool) (domain.Configuration, bool, error) {
			configuration, changed, adoptErr := configadapter.Adopt(root, enable)
			return configuration.Domain(), changed, adoptErr
		},
		LoadActive: stateadapter.Load,
		SaveActive: stateadapter.Save,
		Acquire: func(commonDirectory string) (func() error, error) {
			lock, lockErr := stateadapter.Acquire(commonDirectory)
			if lockErr != nil {
				return nil, lockErr
			}
			return lock.Close, nil
		},
		EnsureBrief: configadapter.EnsureBrief,
		LoadBrief:   configadapter.LoadBrief,
	}
	return cliapp.NewLifecycle(version, workingDirectory, ports), nil
}

func parseArguments(arguments []string, application cliapp.Application) (domain.Request, bool, *domain.Result) {
	invalid := func(subject, message string, jsonOutput bool) (domain.Request, bool, *domain.Result) {
		result := application.Invalid(subject, message)
		return domain.Request{}, jsonOutput, &result
	}
	if len(arguments) > maxArguments {
		return invalid("", "too many arguments", false)
	}
	total := 0
	jsonCount := 0
	filtered := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		if len(argument) > maxArgumentBytes {
			return invalid("", "argument exceeds size limit", jsonCount > 0)
		}
		total += len(argument)
		if total > maxInputBytes {
			return invalid("", "argument input exceeds total size limit", jsonCount > 0)
		}
		if argument == "--json" {
			jsonCount++
			continue
		}
		filtered = append(filtered, argument)
	}
	jsonOutput := jsonCount > 0
	if jsonCount > 1 {
		return invalid("--json", "duplicate --json flag", true)
	}
	if len(filtered) == 0 {
		return domain.Request{Command: domain.CommandHelp}, jsonOutput, nil
	}
	if filtered[0] == "--help" || filtered[0] == "-h" {
		if len(filtered) != 1 {
			return invalid(strings.Join(filtered, " "), "help flag does not accept command arguments", jsonOutput)
		}
		return domain.Request{Command: domain.CommandHelp}, jsonOutput, nil
	}
	if filtered[0] == "--version" {
		if len(filtered) != 1 {
			return invalid(strings.Join(filtered, " "), "version flag does not accept command arguments", jsonOutput)
		}
		return domain.Request{Command: domain.CommandVersion}, jsonOutput, nil
	}
	if strings.HasPrefix(filtered[0], "-") {
		return invalid(filtered[0], "unknown flag", jsonOutput)
	}
	request := domain.Request{Command: domain.Command(filtered[0])}
	options := filtered[1:]
	switch request.Command {
	case domain.CommandHelp, domain.CommandVersion, domain.CommandStatus:
		if len(options) != 0 {
			if strings.HasPrefix(options[0], "-") {
				return invalid(options[0], "unknown flag", jsonOutput)
			}
			return invalid(strings.Join(options, " "), "command does not accept options", jsonOutput)
		}
	case domain.CommandAdopt:
		for _, option := range options {
			if option != "--enable-local-lifecycle" {
				return invalid(option, "unknown adopt option", jsonOutput)
			}
			if request.EnableLocalLifecycle {
				return invalid(option, "duplicate --enable-local-lifecycle flag", jsonOutput)
			}
			request.EnableLocalLifecycle = true
		}
	case domain.CommandBrief:
		seen := make(map[string]bool)
		for index := 0; index < len(options); index += 2 {
			flag := options[index]
			if index+1 >= len(options) {
				return invalid(flag, "brief option is missing its value", jsonOutput)
			}
			value := options[index+1]
			switch flag {
			case "--id":
				if seen[flag] {
					return invalid(flag, "duplicate brief option", jsonOutput)
				}
				request.ChangeID = value
				seen[flag] = true
			case "--tier":
				if seen[flag] {
					return invalid(flag, "duplicate brief option", jsonOutput)
				}
				tier, err := strconv.Atoi(value)
				if err != nil || tier < 1 || tier > 3 {
					return invalid(value, "risk tier must be 1, 2, or 3", jsonOutput)
				}
				request.Tier = domain.RiskTier(tier)
				seen[flag] = true
			case "--problem":
				if seen[flag] {
					return invalid(flag, "duplicate brief option", jsonOutput)
				}
				request.Problem = value
				seen[flag] = true
			case "--scope":
				request.Scope = append(request.Scope, value)
			case "--accept":
				request.AcceptanceCriteria = append(request.AcceptanceCriteria, value)
			case "--risk":
				request.Risks = append(request.Risks, value)
			case "--rollback":
				request.Rollback = append(request.Rollback, value)
			default:
				return invalid(flag, "unknown brief option", jsonOutput)
			}
		}
	default:
		if len(options) != 0 {
			return invalid(strings.Join(filtered, " "), "unknown command", jsonOutput)
		}
	}
	return request, jsonOutput, nil
}
