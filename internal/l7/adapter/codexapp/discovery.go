package codexapp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

// Discover performs the app-server handshake sequentially. Sending account
// and model requests before initialize completes is intentionally avoided:
// compatible servers may close at stdin EOF before processing pipelined work.
func Discover(ctx context.Context, executablePath string) ([]byte, error) {
	if !filepath.IsAbs(executablePath) {
		return nil, errors.New("Codex discovery executable is invalid")
	}
	executable, err := processadapter.Resolve(executablePath)
	if err != nil || executable.Path != executablePath {
		return nil, errors.New("Codex discovery executable identity is unavailable")
	}
	command := exec.CommandContext(ctx, executable.Path, "app-server")
	command.Dir = "/"
	command.Env = processadapter.MinimalEnvironment()
	command.WaitDelay = time.Second
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, errors.New("cannot open Codex discovery input")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, errors.New("cannot open Codex discovery output")
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, errors.New("cannot open Codex discovery diagnostic")
	}
	if err := command.Start(); err != nil {
		return nil, errors.New("cannot start Codex discovery")
	}
	diagnostic := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(io.LimitReader(stderr, maxProtocolData+1))
		if readErr != nil || len(data) > maxProtocolData {
			diagnostic <- errors.New("Codex discovery diagnostic exceeded bounds")
			return
		}
		diagnostic <- nil
	}()
	defer func() {
		_ = stdin.Close()
		if command.Process != nil {
			_ = command.Process.Kill()
		}
		_ = command.Wait()
		select {
		case <-diagnostic:
		default:
		}
	}()
	encoder := json.NewEncoder(stdin)
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), maxProtocolLine)
	var output bytes.Buffer
	total := 0
	send := func(value any) error {
		if err := encoder.Encode(value); err != nil {
			return errors.New("write Codex discovery request")
		}
		return nil
	}
	wait := func(id string) error {
		for scanner.Scan() {
			line := append([]byte{}, scanner.Bytes()...)
			total += len(line)
			if total > maxProtocolData {
				return errors.New("Codex discovery output exceeded bounds")
			}
			var value message
			if json.Unmarshal(line, &value) != nil {
				return errors.New("Codex discovery emitted malformed JSON")
			}
			if len(value.ID) > 0 && value.Method != "" {
				return errors.New("Codex discovery requested unsupported host authority")
			}
			output.Write(line)
			output.WriteByte('\n')
			if string(value.ID) == id {
				if value.Error != nil {
					return errors.New("Codex discovery request failed")
				}
				return nil
			}
		}
		if scanner.Err() != nil {
			return errors.New("Codex discovery framing exceeded limits")
		}
		return errors.New("Codex discovery closed before completion")
	}
	steps := []struct {
		id      string
		request map[string]any
	}{
		{"1", map[string]any{"id": 1, "method": "initialize", "params": map[string]any{"clientInfo": map[string]any{"name": "level7", "title": "Level 7", "version": "1.0.0"}}}},
		{"2", map[string]any{"id": 2, "method": "account/read", "params": map[string]any{"refreshToken": false}}},
		{"3", map[string]any{"id": 3, "method": "model/list", "params": map[string]any{}}},
		{"4", map[string]any{"id": 4, "method": "account/rateLimits/read", "params": map[string]any{}}},
	}
	for index, step := range steps {
		if err := send(step.request); err != nil || wait(step.id) != nil {
			return nil, errors.New("Codex discovery handshake failed")
		}
		if index == 0 {
			if err := send(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
				return nil, err
			}
		}
	}
	return output.Bytes(), nil
}
