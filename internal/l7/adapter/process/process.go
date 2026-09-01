// Package process owns bounded child-process discovery and supervision.
package process

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MaxExecutableBytes = 512 << 20
	MaxArguments       = 64
	MaxArgumentBytes   = 4096
	MaxInputBytes      = 1 << 20
	pipeDrainDelay     = 500 * time.Millisecond
)

var ErrOutputLimit = errors.New("process output limit reached")

type Executable struct {
	Path   string
	Digest string
}

type Request struct {
	Executable     string
	Arguments      []string
	Input          []byte
	Directory      string
	Environment    []string
	MaxOutputBytes int
	Timeout        time.Duration
}

type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

type Runner struct{}

func Resolve(name string) (Executable, error) {
	if name == "" || strings.ContainsAny(name, "\r\n\x00") {
		return Executable{}, errors.New("executable name is invalid")
	}
	resolved := name
	var err error
	if !filepath.IsAbs(resolved) {
		if strings.ContainsRune(resolved, filepath.Separator) {
			return Executable{}, errors.New("relative executable paths are unsupported")
		}
		resolved, err = exec.LookPath(resolved)
		if err != nil {
			return Executable{}, errors.New("executable is unavailable")
		}
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return Executable{}, errors.New("cannot resolve executable path")
	}
	resolved, err = filepath.EvalSymlinks(resolved)
	if err != nil {
		return Executable{}, errors.New("cannot resolve physical executable")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 || info.Size() < 1 || info.Size() > MaxExecutableBytes {
		return Executable{}, errors.New("executable is not a bounded regular executable file")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return Executable{}, errors.New("cannot open executable for identity")
	}
	hash := sha256.New()
	written, copyErr := io.Copy(hash, io.LimitReader(file, MaxExecutableBytes+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written != info.Size() {
		return Executable{}, errors.New("cannot read stable executable identity")
	}
	rechecked, err := os.Stat(resolved)
	if err != nil || !os.SameFile(info, rechecked) || info.Size() != rechecked.Size() || !info.ModTime().Equal(rechecked.ModTime()) {
		return Executable{}, errors.New("executable changed during identity calculation")
	}
	return Executable{Path: resolved, Digest: fmt.Sprintf("%x", hash.Sum(nil))}, nil
}

func MinimalEnvironment() []string {
	allowed := map[string]bool{
		"HOME": true, "PATH": true, "TMPDIR": true, "LANG": true, "LC_ALL": true,
		"TERM": true, "COLORTERM": true, "USER": true, "LOGNAME": true, "SHELL": true,
	}
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if found && allowed[key] && safeEnvironmentValue(value) {
			values[key] = value
		}
	}
	values["LC_ALL"] = "C"
	values["LANG"] = "C"
	values["GIT_TERMINAL_PROMPT"] = "0"
	values["NO_COLOR"] = "1"
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}

func (Runner) Run(ctx context.Context, request Request) (Result, error) {
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, request.Timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	command := exec.Command(request.Executable, request.Arguments...)
	command.Dir = request.Directory
	command.Env = append([]string{}, request.Environment...)
	command.Stdin = bytes.NewReader(request.Input)
	// A descendant can deliberately escape the supervised process group while
	// retaining inherited pipes. Bound the time Cmd.Wait may spend draining
	// those pipes after the direct child exits or is terminated.
	command.WaitDelay = pipeDrainDelay
	configureProcessGroup(command)

	overflow := make(chan struct{})
	budget := &sharedBudget{remaining: request.MaxOutputBytes, overflow: overflow}
	stdout := &boundedBuffer{budget: budget}
	stderr := &boundedBuffer{budget: budget}
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		return Result{}, fmt.Errorf("start child process: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()

	var waitErr error
	var terminalErr error
	select {
	case waitErr = <-done:
	case <-overflow:
		waitErr = stopProcessGroup(command.Process.Pid, done)
		terminalErr = ErrOutputLimit
	case <-ctx.Done():
		waitErr = stopProcessGroup(command.Process.Pid, done)
		terminalErr = ctx.Err()
	}
	result := Result{ExitCode: exitCode(command, waitErr), Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if terminalErr != nil {
		return result, terminalErr
	}
	if budget.Exceeded() {
		return result, ErrOutputLimit
	}
	if waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			return result, fmt.Errorf("wait for child process: %w", waitErr)
		}
	}
	return result, nil
}

func validateRequest(request Request) error {
	if !filepath.IsAbs(request.Executable) || !filepath.IsAbs(request.Directory) || request.MaxOutputBytes < 1024 || request.MaxOutputBytes > 64<<20 || request.Timeout < time.Millisecond || request.Timeout > 24*time.Hour || len(request.Input) > MaxInputBytes || len(request.Arguments) > MaxArguments || len(request.Environment) > 256 {
		return errors.New("process request bounds are invalid")
	}
	info, err := os.Stat(request.Executable)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return errors.New("process executable is unsafe")
	}
	directory, err := filepath.EvalSymlinks(request.Directory)
	if err != nil || directory != filepath.Clean(request.Directory) {
		return errors.New("process working directory is unsafe")
	}
	total := len(request.Input)
	for _, argument := range request.Arguments {
		if len(argument) > MaxArgumentBytes || strings.ContainsRune(argument, 0) {
			return errors.New("process argument is unsafe")
		}
		total += len(argument)
	}
	if total > MaxInputBytes {
		return errors.New("process input exceeds size limit")
	}
	seen := make(map[string]bool)
	for _, entry := range request.Environment {
		key, value, found := strings.Cut(entry, "=")
		if !found || key == "" || seen[key] || !safeEnvironmentKey(key) || !safeEnvironmentValue(value) {
			return errors.New("process environment is unsafe")
		}
		seen[key] = true
	}
	return nil
}

func safeEnvironmentKey(value string) bool {
	for _, character := range value {
		if (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return value != ""
}

func safeEnvironmentValue(value string) bool {
	return len(value) <= 32<<10 && !strings.ContainsRune(value, 0)
}

func exitCode(command *exec.Cmd, waitErr error) int {
	if command.ProcessState != nil {
		return command.ProcessState.ExitCode()
	}
	if waitErr != nil {
		return -1
	}
	return 0
}

type sharedBudget struct {
	mu        sync.Mutex
	remaining int
	exceeded  bool
	overflow  chan struct{}
	once      sync.Once
}

func (budget *sharedBudget) consume(data []byte) []byte {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	if len(data) <= budget.remaining {
		budget.remaining -= len(data)
		return data
	}
	accepted := data[:max(0, budget.remaining)]
	budget.remaining = 0
	budget.exceeded = true
	budget.once.Do(func() { close(budget.overflow) })
	return accepted
}

func (budget *sharedBudget) Exceeded() bool {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	return budget.exceeded
}

type boundedBuffer struct {
	mu     sync.Mutex
	data   bytes.Buffer
	budget *sharedBudget
}

func (buffer *boundedBuffer) Write(data []byte) (int, error) {
	accepted := buffer.budget.consume(data)
	buffer.mu.Lock()
	_, _ = buffer.data.Write(accepted)
	buffer.mu.Unlock()
	return len(data), nil
}

func (buffer *boundedBuffer) Bytes() []byte {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return append([]byte{}, buffer.data.Bytes()...)
}
