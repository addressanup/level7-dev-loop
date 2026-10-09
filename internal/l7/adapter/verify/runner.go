// Package verify executes explicit repository verification argv through the shared supervisor.
package verify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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

// FailureTailBytes bounds each output stream in a repair-feedback tail.
const FailureTailBytes = 4 << 10

func (runner Runner) Run(ctx context.Context, root string, commands []domain.VerificationCommand, maxOutputBytes, maxSeconds int) ([]domain.CheckResult, error) {
	checks, _, err := runner.RunWithFailureTail(ctx, root, commands, maxOutputBytes, maxSeconds)
	return checks, err
}

// RunWithFailureTail behaves like Run and also returns the end of the failing
// command's stdout and stderr, each bounded by FailureTailBytes, so a worker
// can repair the failure. The tail is empty when every command passed.
func (runner Runner) RunWithFailureTail(ctx context.Context, root string, commands []domain.VerificationCommand, maxOutputBytes, maxSeconds int) ([]domain.CheckResult, string, error) {
	if runner.resolve == nil || runner.run == nil || root == "" || len(commands) < 1 || len(commands) > 32 || maxOutputBytes < 64<<10 || maxOutputBytes > 64<<20 || maxSeconds < 1 || maxSeconds > 86400 {
		return nil, "", errors.New("verification runner configuration is invalid")
	}
	checks := make([]domain.CheckResult, 0, len(commands))
	executables := make(map[string]processadapter.Executable)
	for _, command := range commands {
		if err := ctx.Err(); err != nil {
			return checks, "", err
		}
		if command.Name == "" || len(command.Argv) < 1 {
			return checks, "", errors.New("verification command is incomplete")
		}
		executable, found := executables[command.Argv[0]]
		if !found {
			var err error
			executable, err = runner.resolve(command.Argv[0])
			if err != nil {
				check := domain.CheckResult{Name: command.Name, Benchmark: command.Benchmark, Passed: false, ExitCode: -1, Code: "L7-VERIFY-002", Message: "verification executable is unavailable"}
				checks = append(checks, check)
				return checks, "", fmt.Errorf("verification command %q cannot resolve its executable: %w", command.Name, err)
			}
			executables[command.Argv[0]] = executable
		}
		rechecked, resolveErr := runner.resolve(executable.Path)
		if resolveErr != nil || rechecked.Path != executable.Path || rechecked.Digest != executable.Digest {
			check := domain.CheckResult{Name: command.Name, Benchmark: command.Benchmark, Passed: false, ExitCode: -1, Code: "L7-VERIFY-002", Message: "verification executable identity changed"}
			checks = append(checks, check)
			return checks, "", fmt.Errorf("verification command %q executable identity changed", command.Name)
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
			return checks, failureTail(result), runErr
		}
		if result.ExitCode != 0 {
			check.Code = "L7-VERIFY-001"
			check.Message = boundedDiagnostic(result)
			checks = append(checks, check)
			return checks, failureTail(result), fmt.Errorf("verification command %q failed with exit %d", command.Name, result.ExitCode)
		}
		check.Passed = true
		check.Code = "L7-VERIFY-000"
		check.Message = "command passed"
		checks = append(checks, check)
	}
	return checks, "", nil
}

// failureTail keeps line structure for a worker while dropping control
// characters and invalid UTF-8.
func failureTail(result processadapter.Result) string {
	var tail strings.Builder
	for _, stream := range []struct {
		name string
		data []byte
	}{{"stdout", result.Stdout}, {"stderr", result.Stderr}} {
		text := strings.TrimSpace(sanitizeTail(stream.data))
		if text == "" {
			continue
		}
		if len(text) > FailureTailBytes {
			cut := len(text) - FailureTailBytes
			for cut < len(text) && !utf8.RuneStart(text[cut]) {
				cut++
			}
			text = "…" + text[cut:]
		}
		fmt.Fprintf(&tail, "%s:\n%s\n", stream.name, text)
	}
	return tail.String()
}

func sanitizeTail(data []byte) string {
	return strings.Map(func(character rune) rune {
		switch {
		case character == '\n' || character == '\t':
			return character
		case character == '\r':
			return '\n'
		case character < 0x20 || character == 0x7f || character == utf8.RuneError:
			return -1
		default:
			return character
		}
	}, strings.ToValidUTF8(string(data), ""))
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
