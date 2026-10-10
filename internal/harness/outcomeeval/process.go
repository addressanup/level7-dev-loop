package main

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	maxCommandOutput = 32 << 20
	maxTreeFiles     = 2048
	maxTreeBytes     = 32 << 20
	tailBytes        = 1536
	sandboxExec      = "/usr/bin/sandbox-exec"
	baseDate         = "2026-01-01T00:00:00Z"
	evalIdentity     = "Level 7 Outcome Eval"
	evalEmail        = "outcome-eval@localhost"
)

type commandSpec struct {
	Dir     string
	Env     []string
	Name    string
	Args    []string
	Stdin   []byte
	Timeout time.Duration
}

type commandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	TimedOut bool
	Elapsed  time.Duration
}

// execute runs one bounded command in its own process group and kills the
// whole group on timeout or cancellation.
func execute(ctx context.Context, spec commandSpec) (commandResult, error) {
	if spec.Timeout <= 0 {
		return commandResult{}, errors.New("command timeout is required")
	}
	ctx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	command := exec.Command(spec.Name, spec.Args...)
	command.Dir, command.Env = spec.Dir, spec.Env
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 5 * time.Second
	stdout, stderr := &boundedBuffer{limit: maxCommandOutput}, &boundedBuffer{limit: maxCommandOutput}
	command.Stdout, command.Stderr = stdout, stderr
	if spec.Stdin != nil {
		command.Stdin = bytes.NewReader(spec.Stdin)
	}
	started := time.Now()
	if err := command.Start(); err != nil {
		return commandResult{}, fmt.Errorf("start %s: %w", filepath.Base(spec.Name), err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	var waitErr error
	result := commandResult{}
	select {
	case waitErr = <-done:
	case <-ctx.Done():
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		waitErr = <-done
		result.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
	}
	result.Stdout, result.Stderr, result.Elapsed = stdout.Bytes(), stderr.Bytes(), time.Since(started)
	result.ExitCode = command.ProcessState.ExitCode()
	if stdout.exceeded || stderr.exceeded {
		return result, errors.New("command output exceeded its bound")
	}
	if ctx.Err() != nil && !result.TimedOut {
		return result, ctx.Err()
	}
	var exitErr *exec.ExitError
	// A detached descendant may keep an output pipe open after the command
	// exits; WaitDelay bounds that wait, and the exit status still stands.
	if waitErr != nil && !errors.As(waitErr, &exitErr) && !errors.Is(waitErr, exec.ErrWaitDelay) && !result.TimedOut {
		return result, waitErr
	}
	return result, nil
}

type boundedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (value *boundedBuffer) Write(data []byte) (int, error) {
	if remaining := value.limit - value.buffer.Len(); len(data) > remaining {
		value.buffer.Write(data[:max(0, remaining)])
		value.exceeded = true
		return len(data), nil
	}
	return value.buffer.Write(data)
}

func (value *boundedBuffer) Bytes() []byte { return append([]byte{}, value.buffer.Bytes()...) }

// baseEnvironment keeps the same small variable set Level 7 gives provider
// processes, with PATH replaced.
func baseEnvironment(pathValue string) []string {
	environment := []string{"PATH=" + pathValue, "LANG=C", "LC_ALL=C", "NO_COLOR=1", "GIT_TERMINAL_PROMPT=0"}
	for _, key := range []string{"HOME", "TMPDIR", "USER", "LOGNAME", "SHELL", "TERM"} {
		if value, found := os.LookupEnv(key); found && !strings.ContainsAny(value, "\x00\r\n") {
			environment = append(environment, key+"="+value)
		}
	}
	return environment
}

func gitEnvironment(extra ...string) []string {
	environment := append(baseEnvironment(os.Getenv("PATH")), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	return append(environment, extra...)
}

func git(ctx context.Context, directory string, arguments ...string) (string, error) {
	result, err := execute(ctx, commandSpec{Dir: directory, Env: gitEnvironment(), Name: "git", Args: arguments, Timeout: time.Minute})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("git %s: %s", strings.Join(arguments, " "), tail(result.Stderr))
	}
	return strings.TrimSpace(string(result.Stdout)), nil
}

// materializeRepository writes a task's starting files and commits them with
// a fixed identity and date, so every trial of a task starts at the same
// commit. It returns that commit.
func materializeRepository(ctx context.Context, directory string, files map[string][]byte) (string, error) {
	if err := writeTree(directory, files); err != nil {
		return "", err
	}
	hooks := filepath.Join(filepath.Dir(directory), "no-hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		return "", err
	}
	steps := [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.name", evalIdentity},
		{"config", "user.email", evalEmail},
		{"config", "commit.gpgsign", "false"},
		{"config", "core.hooksPath", hooks},
		{"add", "-A"},
	}
	for _, step := range steps {
		if _, err := git(ctx, directory, step...); err != nil {
			return "", err
		}
	}
	if err := commit(ctx, directory, "chore: task snapshot", baseDate); err != nil {
		return "", err
	}
	exclude := filepath.Join(directory, ".git", "info", "exclude")
	if err := appendFile(exclude, ".l7/\n"); err != nil {
		return "", err
	}
	return git(ctx, directory, "rev-parse", "HEAD")
}

func commit(ctx context.Context, directory, message, date string) error {
	environment := gitEnvironment(
		"GIT_AUTHOR_NAME="+evalIdentity, "GIT_AUTHOR_EMAIL="+evalEmail, "GIT_AUTHOR_DATE="+date,
		"GIT_COMMITTER_NAME="+evalIdentity, "GIT_COMMITTER_EMAIL="+evalEmail, "GIT_COMMITTER_DATE="+date,
	)
	result, err := execute(ctx, commandSpec{Dir: directory, Env: environment, Name: "git", Args: []string{"commit", "-q", "--allow-empty", "--no-verify", "-m", message}, Timeout: time.Minute})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("git commit: %s", tail(result.Stderr))
	}
	return nil
}

// snapshotWorktree commits everything an agent left in a checkout, tracked or
// not (ignored files excepted), and returns that commit.
func snapshotWorktree(ctx context.Context, directory string) (string, error) {
	if _, err := git(ctx, directory, "add", "-A"); err != nil {
		return "", err
	}
	if err := commit(ctx, directory, "outcome-eval: delivered snapshot", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return "", err
	}
	return git(ctx, directory, "rev-parse", "HEAD")
}

// treeFiles reads a commit's files through git archive. Symlinks and other
// non-regular entries are reported, not followed.
func treeFiles(ctx context.Context, directory, commit string) (map[string][]byte, []string, error) {
	result, err := execute(ctx, commandSpec{Dir: directory, Env: gitEnvironment(), Name: "git", Args: []string{"archive", "--format=tar", commit}, Timeout: time.Minute})
	if err != nil {
		return nil, nil, err
	}
	if result.ExitCode != 0 {
		return nil, nil, fmt.Errorf("git archive: %s", tail(result.Stderr))
	}
	files, irregular := map[string][]byte{}, []string{}
	reader, total := tar.NewReader(bytes.NewReader(result.Stdout)), 0
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("read archive: %w", err)
		}
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeXGlobalHeader:
			continue
		case tar.TypeReg:
		default:
			irregular = append(irregular, header.Name)
			continue
		}
		if !safeRelative(header.Name) || len(files) >= maxTreeFiles || total+int(header.Size) > maxTreeBytes {
			return nil, nil, fmt.Errorf("delivered tree exceeds bounds at %q", header.Name)
		}
		data, err := io.ReadAll(io.LimitReader(reader, header.Size))
		if err != nil {
			return nil, nil, err
		}
		files[header.Name], total = data, total+len(data)
	}
	sort.Strings(irregular)
	return files, irregular, nil
}

func writeTree(directory string, files map[string][]byte) error {
	for _, relative := range sortedKeys(files) {
		if !safeRelative(relative) {
			return fmt.Errorf("unsafe tree path %q", relative)
		}
		target := filepath.Join(directory, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, files[relative], 0o644); err != nil {
			return err
		}
	}
	return nil
}

func appendFile(name, text string) error {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func writeAtomic(name string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(name), "."+filepath.Base(name)+".*")
	if err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		_ = os.Remove(temporary.Name())
		return err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporary.Name())
		return err
	}
	return os.Rename(temporary.Name(), name)
}

func physicalDirectory(name string) (string, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func within(name, root string) bool {
	relative, err := filepath.Rel(root, name)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func tail(data []byte) string {
	text := strings.ToValidUTF8(string(data), "?")
	if len(text) > tailBytes {
		text = "..." + text[len(text)-tailBytes:]
	}
	return strings.TrimSpace(text)
}

// sandboxAvailable reports whether checks can run under the macOS sandbox
// that denies network access and writes outside the grading copy.
func sandboxAvailable() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	info, err := os.Lstat(sandboxExec)
	return err == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0
}

func sandboxProfile(writable string) string {
	quoted := strings.ReplaceAll(strings.ReplaceAll(writable, `\`, `\\`), `"`, `\"`)
	return `(version 1)(allow default)(deny network*)(deny file-write*)` +
		`(allow file-write* (subpath "` + quoted + `") (literal "/dev/null") (literal "/dev/zero") (literal "/dev/tty") (regex #"^/dev/fd/"))`
}
