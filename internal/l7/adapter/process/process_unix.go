//go:build darwin || linux

package process

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const terminationGrace = 250 * time.Millisecond

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
		return <-done
	}
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
