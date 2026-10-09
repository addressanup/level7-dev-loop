//go:build darwin || linux

package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	terminationGrace  = 250 * time.Millisecond
	finalizationGrace = pipeDrainDelay + terminationGrace
)

func NotifyContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}

func configureProcessGroup(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func stopProcessGroup(pid int, done <-chan error) error {
	_ = signalProcessGroup(pid, syscall.SIGTERM)
	timer := time.NewTimer(terminationGrace)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		_ = signalProcessGroup(pid, syscall.SIGKILL)
		finalization := time.NewTimer(finalizationGrace)
		defer finalization.Stop()
		select {
		case err := <-done:
			return err
		case <-finalization.C:
			return errors.New("process group did not finalize within the bounded drain delay")
		}
	}
}

// StartDetached launches executable as a new session leader that outlives the
// caller. Its stdin is /dev/null and its output is appended to logPath. The
// returned channel receives the exit result if the caller is still running.
func StartDetached(executable string, arguments []string, directory, logPath string) (int, <-chan error, error) {
	if !filepath.IsAbs(executable) || !filepath.IsAbs(directory) || !filepath.IsAbs(logPath) {
		return 0, nil, errors.New("detached process paths must be absolute")
	}
	if info, err := os.Lstat(logPath); err == nil && !info.Mode().IsRegular() {
		return 0, nil, errors.New("detached process log is not a regular file")
	}
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, nil, fmt.Errorf("open detached process log: %w", err)
	}
	defer log.Close()
	null, err := os.Open(os.DevNull)
	if err != nil {
		return 0, nil, err
	}
	defer null.Close()
	command := exec.Command(executable, arguments...)
	command.Dir, command.Stdin, command.Stdout, command.Stderr = directory, null, log, log
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return 0, nil, fmt.Errorf("start detached process: %w", err)
	}
	exited := make(chan error, 1)
	go func() { exited <- command.Wait() }()
	return command.Process.Pid, exited, nil
}

// Terminate asks one process to stop. A process that already exited is not an
// error.
func Terminate(pid int) error {
	if pid < 2 {
		return errors.New("invalid process ID")
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

func signalProcessGroup(pid int, signal syscall.Signal) error {
	if pid < 1 {
		return errors.New("invalid process group")
	}
	err := syscall.Kill(-pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
