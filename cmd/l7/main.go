package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	authorityadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/authority"
	claudeadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/claude"
	codexadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/codex"
	configadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/config"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	stateadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
	verifyadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/verify"
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
	ctx, cancel := processadapter.NotifyContext(context.Background())
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, arguments []string, stdout, stderr io.Writer) int {
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "FAILED code=L7-CLI-004 message=%q\n", "cannot determine working directory")
		return 1
	}
	terminal := authorityadapter.NewTerminal(os.Stdin, stderr, terminalAvailable(os.Stdin) && terminalAvailable(stderr), "accountable-owner")
	return runAtWithTerminal(ctx, arguments, workingDirectory, stdout, stderr, terminal)
}

func runAt(ctx context.Context, arguments []string, workingDirectory string, stdout, stderr io.Writer) int {
	terminal := authorityadapter.NewTerminal(nil, stderr, false, "accountable-owner")
	return runAtWithTerminal(ctx, arguments, workingDirectory, stdout, stderr, terminal)
}

func runAtWithTerminal(ctx context.Context, arguments []string, workingDirectory string, stdout, stderr io.Writer, terminal authorityadapter.Terminal) int {
	baseApplication := cliapp.New(version)
	request, jsonOutput, invalid := parseArguments(arguments, baseApplication)
	application, compositionErr := lifecycleApplication(workingDirectory, terminal)
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

func lifecycleApplication(workingDirectory string, terminal authorityadapter.Terminal) (cliapp.Application, error) {
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
		Pending: func(ctx context.Context, cwd string, maxOutput, maxPaths int) (domain.PendingChanges, error) {
			return gitClient.WithLimits(maxOutput, maxPaths).Pending(ctx, cwd)
		},
		PathCommit: gitClient.PathCommit,
		Commit:     gitClient.Commit,
		CommitMatches: func(ctx context.Context, root, commit, parent, subject string, maxOutput, maxPaths int) (bool, error) {
			return gitClient.WithLimits(maxOutput, maxPaths).CommitMatches(ctx, root, commit, parent, subject)
		},
		CommitPaths: func(ctx context.Context, root, parent, commit string, maxOutput, maxPaths int) ([]string, error) {
			return gitClient.WithLimits(maxOutput, maxPaths).CommitPaths(ctx, root, parent, commit)
		},
		CommitTree:       gitClient.CommitTree,
		PathSetDigest:    gitadapter.PathSetDigest,
		ConfirmApproval:  terminal.Confirm,
		LoadApproval:     authorityadapter.Load,
		SaveApproval:     authorityadapter.Save,
		LoadRun:          stateadapter.LoadRun,
		SaveRun:          stateadapter.SaveRun,
		LoadVerification: stateadapter.LoadVerification,
		SaveVerification: stateadapter.SaveVerification,
		LoadReview:       stateadapter.LoadReview,
		SaveReview:       stateadapter.SaveReview,
		LoadReadiness:    stateadapter.LoadReadiness,
		SaveReadiness:    stateadapter.SaveReadiness,
		RunProvider: func(ctx context.Context, task domain.ProviderTask, maxOutput, maxSeconds int) (domain.ProviderResponse, error) {
			switch task.Provider {
			case domain.ProviderCodex:
				return codexadapter.New().Run(ctx, task, maxOutput, maxSeconds)
			case domain.ProviderClaude:
				return claudeadapter.New().Run(ctx, task, maxOutput, maxSeconds)
			default:
				return domain.ProviderResponse{}, errors.New("unsupported provider")
			}
		},
		RunVerification:   verifyadapter.New(nil, nil).Run,
		WriteVerification: stateadapter.WriteVerificationArtifact,
		WriteAudit:        stateadapter.WriteAuditArtifact,
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
	case domain.CommandHelp, domain.CommandVersion, domain.CommandStatus, domain.CommandVerify, domain.CommandReady:
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
	case domain.CommandRun:
		seen := make(map[string]bool)
		for index := 0; index < len(options); index += 2 {
			flag := options[index]
			if index+1 >= len(options) {
				return invalid(flag, "run option is missing its value", jsonOutput)
			}
			if seen[flag] {
				return invalid(flag, "duplicate run option", jsonOutput)
			}
			value := options[index+1]
			switch flag {
			case "--agent":
				request.Agent = domain.Provider(value)
			case "--message":
				request.CommitMessage = value
			default:
				return invalid(flag, "unknown run option", jsonOutput)
			}
			seen[flag] = true
		}
		if !request.Agent.Valid() || request.CommitMessage == "" {
			return invalid(strings.Join(filtered, " "), "run requires --agent codex|claude and --message", jsonOutput)
		}
	case domain.CommandReview:
		if len(options) != 2 || options[0] != "--agent" || !domain.Provider(options[1]).Valid() {
			return invalid(strings.Join(filtered, " "), "review requires --agent codex|claude", jsonOutput)
		}
		request.Agent = domain.Provider(options[1])
	default:
		if len(options) != 0 {
			return invalid(strings.Join(filtered, " "), "unknown command", jsonOutput)
		}
	}
	return request, jsonOutput, nil
}

func terminalAvailable(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
