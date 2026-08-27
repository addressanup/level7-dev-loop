// Package verify executes explicit repository verification argv through the shared supervisor.
package verify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type ResolveFunc func(string) (processadapter.Executable, error)
type RunFunc func(context.Context, processadapter.Request) (processadapter.Result, error)

type Runner struct {
	resolve ResolveFunc
	run     RunFunc
}

func New(resolve ResolveFunc, run RunFunc) Runner {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	return Runner{resolve: resolve, run: run}
}

func (runner Runner) Run(ctx context.Context, root string, commands []domain.VerificationCommand, maxOutputBytes, maxSeconds int) ([]domain.CheckResult, error) {
	if runner.resolve == nil || runner.run == nil || root == "" || len(commands) < 1 || len(commands) > 32 || maxOutputBytes < 64<<10 || maxOutputBytes > 64<<20 || maxSeconds < 1 || maxSeconds > 86400 {
		return nil, errors.New("verification runner configuration is invalid")
	}
	checks := make([]domain.CheckResult, 0, len(commands))
	executables := make(map[string]processadapter.Executable)
	for _, command := range commands {
		if err := ctx.Err(); err != nil {
			return checks, err
		}
		if command.Name == "" || len(command.Argv) < 1 {
			return checks, errors.New("verification command is incomplete")
		}
		executable, found := executables[command.Argv[0]]
		if !found {
			var err error
			executable, err = runner.resolve(command.Argv[0])
			if err != nil {
				check := domain.CheckResult{Name: command.Name, Benchmark: command.Benchmark, Passed: false, ExitCode: -1, Code: "L7-VERIFY-002", Message: "verification executable is unavailable"}
				checks = append(checks, check)
				return checks, fmt.Errorf("verification command %q cannot resolve its executable: %w", command.Name, err)
			}
			executables[command.Argv[0]] = executable
		}
		result, runErr := runner.run(ctx, processadapter.Request{
			Executable: executable.Path, Arguments: append([]string{}, command.Argv[1:]...), Directory: root,
			Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: maxOutputBytes,
			Timeout: time.Duration(maxSeconds) * time.Second,
		})
		check := domain.CheckResult{Name: command.Name, Benchmark: command.Benchmark, ExitCode: result.ExitCode}
		if runErr != nil {
			check.Code = "L7-VERIFY-003"
			check.Message = boundedMessage(runErr.Error())
			checks = append(checks, check)
			return checks, runErr
		}
		if result.ExitCode != 0 {
			check.Code = "L7-VERIFY-001"
			check.Message = boundedDiagnostic(result)
			checks = append(checks, check)
			return checks, fmt.Errorf("verification command %q failed with exit %d", command.Name, result.ExitCode)
		}
		check.Passed = true
		check.Code = "L7-VERIFY-000"
		check.Message = "command passed"
		checks = append(checks, check)
	}
	return checks, nil
}

func boundedDiagnostic(result processadapter.Result) string {
	diagnostic := strings.TrimSpace(string(append(append([]byte{}, result.Stderr...), result.Stdout...)))
	if diagnostic == "" {
		diagnostic = "command returned a nonzero exit status"
	}
	return boundedMessage(diagnostic)
}

func boundedMessage(value string) string {
	value = strings.Map(func(character rune) rune {
		if character == '\n' || character == '\r' || character == 0 || character == 0x7f || (character < 0x20 && character != '\t') {
			return ' '
		}
		return character
	}, value)
	value = strings.TrimSpace(value)
	if len(value) > 512 {
		value = value[:512]
	}
	return value
}
