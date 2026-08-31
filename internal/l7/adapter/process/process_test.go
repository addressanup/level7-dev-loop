package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunnerReturnsBoundedOutputAndExitCode(t *testing.T) {
	request := helperRequest(t, "echo")
	result, err := (Runner{}).Run(context.Background(), request)
	if err != nil || result.ExitCode != 0 || string(result.Stdout) != "hello\n" || string(result.Stderr) != "diagnostic\n" {
		t.Fatalf("Run()=%+v error=%v", result, err)
	}
	request = helperRequest(t, "exit")
	result, err = (Runner{}).Run(context.Background(), request)
	if err != nil || result.ExitCode != 7 {
		t.Fatalf("nonzero Run()=%+v error=%v", result, err)
	}
}

func TestRunnerCancelsCompleteInheritedProcessGroup(t *testing.T) {
	request := helperRequest(t, "fork")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	result, err := (Runner{}).Run(ctx, request)
	if !errors.Is(err, context.DeadlineExceeded) || result.ExitCode == 0 {
		t.Fatalf("cancelled Run()=%+v error=%v", result, err)
	}
	childText := strings.TrimSpace(string(result.Stdout))
	childPID, parseErr := strconv.Atoi(childText)
	if parseErr != nil || childPID < 1 {
		t.Fatalf("child pid=%q error=%v", childText, parseErr)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		err := syscall.Kill(childPID, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("child process %d survived process-group cancellation", childPID)
}

func TestRunnerStopsOutputFloodAtAggregateLimit(t *testing.T) {
	request := helperRequest(t, "flood")
	request.MaxOutputBytes = 4096
	result, err := (Runner{}).Run(context.Background(), request)
	if !errors.Is(err, ErrOutputLimit) || len(result.Stdout)+len(result.Stderr) > request.MaxOutputBytes {
		t.Fatalf("flood Run() bytes=%d result=%+v error=%v", len(result.Stdout)+len(result.Stderr), result, err)
	}
}

func TestRunnerBoundsSessionEscapedInheritedPipes(t *testing.T) {
	request := helperRequest(t, "escape-pipes")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	started := time.Now()
	result, err := (Runner{}).Run(ctx, request)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > 2*time.Second {
		t.Fatalf("escaped Run() elapsed=%s result=%+v error=%v", time.Since(started), result, err)
	}
	childText := strings.TrimSpace(string(result.Stdout))
	childPID, parseErr := strconv.Atoi(childText)
	if parseErr != nil || childPID < 1 {
		t.Fatalf("escaped child pid=%q error=%v", childText, parseErr)
	}
	_ = syscall.Kill(childPID, syscall.SIGKILL)
}

func TestMinimalEnvironmentStripsAmbientSecrets(t *testing.T) {
	t.Setenv("L7_TEST_SECRET_TOKEN", "do-not-pass")
	t.Setenv("USER", "level-seven-user")
	foundUser := false
	for _, entry := range MinimalEnvironment() {
		if strings.Contains(entry, "SECRET") || strings.Contains(entry, "do-not-pass") {
			t.Fatalf("secret inherited in %q", entry)
		}
		foundUser = foundUser || entry == "USER=level-seven-user"
	}
	if !foundUser {
		t.Fatal("non-secret host identity required by authenticated CLIs was dropped")
	}
}

func TestResolvePinsPhysicalExecutableIdentity(t *testing.T) {
	executable, err := Resolve(os.Args[0])
	if err != nil || !filepath.IsAbs(executable.Path) || len(executable.Digest) != 64 {
		t.Fatalf("Resolve()=%+v error=%v", executable, err)
	}
}

func BenchmarkBoundedOutput(b *testing.B) {
	request := helperRequestB(b, "echo")
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := (Runner{}).Run(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func TestProcessHelper(t *testing.T) {
	if os.Getenv("L7_PROCESS_HELPER") != "1" {
		return
	}
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || separator+1 >= len(os.Args) {
		os.Exit(90)
	}
	switch os.Args[separator+1] {
	case "echo":
		fmt.Fprintln(os.Stdout, "hello")
		fmt.Fprintln(os.Stderr, "diagnostic")
	case "exit":
		os.Exit(7)
	case "hang":
		time.Sleep(24 * time.Hour)
	case "flood":
		data := strings.Repeat("x", 1<<20)
		_, _ = fmt.Fprint(os.Stdout, data)
		time.Sleep(24 * time.Hour)
	case "fork":
		command := exec.Command(os.Args[0], "-test.run=TestProcessHelper", "--", "hang")
		command.Env = append(os.Environ(), "L7_PROCESS_HELPER=1")
		if err := command.Start(); err != nil {
			os.Exit(91)
		}
		fmt.Fprintln(os.Stdout, command.Process.Pid)
		time.Sleep(24 * time.Hour)
	case "escape-pipes":
		command := exec.Command(os.Args[0], "-test.run=TestProcessHelper", "--", "escaped-hang")
		command.Env = append(os.Environ(), "L7_PROCESS_HELPER=1")
		if err := command.Start(); err != nil {
			os.Exit(93)
		}
		marker := filepath.Join(".", ".escaped-session-ready")
		deadline := time.Now().Add(time.Second)
		for {
			if _, err := os.Stat(marker); err == nil {
				break
			}
			if time.Now().After(deadline) {
				os.Exit(95)
			}
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Fprintln(os.Stdout, command.Process.Pid)
		time.Sleep(24 * time.Hour)
	case "escaped-hang":
		if _, err := syscall.Setsid(); err != nil {
			os.Exit(94)
		}
		if err := os.WriteFile(filepath.Join(".", ".escaped-session-ready"), []byte("ready\n"), 0o600); err != nil {
			os.Exit(96)
		}
		time.Sleep(24 * time.Hour)
	default:
		os.Exit(92)
	}
	os.Exit(0)
}

func helperRequest(t *testing.T, mode string) Request {
	t.Helper()
	return helperRequestFor(t.TempDir(), mode)
}

func helperRequestB(b *testing.B, mode string) Request {
	b.Helper()
	return helperRequestFor(b.TempDir(), mode)
}

func helperRequestFor(directory, mode string) Request {
	executable, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	directory, err = filepath.EvalSymlinks(directory)
	if err != nil {
		panic(err)
	}
	return Request{
		Executable:     executable,
		Arguments:      []string{"-test.run=TestProcessHelper", "--", mode},
		Directory:      directory,
		Environment:    append(MinimalEnvironment(), "L7_PROCESS_HELPER=1"),
		MaxOutputBytes: 64 << 10,
		Timeout:        5 * time.Second,
	}
}
