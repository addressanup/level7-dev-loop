// Package codexapp runs authenticated Codex app-server sessions without
// depending on an exact CLI version string.
package codexapp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	maxProtocolLine = 4 << 20
	maxProtocolData = 64 << 20
)

type Assignment struct {
	Executable string
	Root       string
	Model      string
	Effort     domain.ReasoningEffort
	Prompt     string
	SessionID  string
	Reviewer   bool
}

type Result struct {
	SessionID   string
	TurnID      string
	Status      string
	Summary     string
	QuotaReset  string
	FailureCode string
	Decision    domain.ReviewDecision
}

type message struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Result map[string]any  `json:"result,omitempty"`
	Error  map[string]any  `json:"error,omitempty"`
	Params map[string]any  `json:"params,omitempty"`
}

func Run(ctx context.Context, assignment Assignment) (Result, error) {
	if !filepath.IsAbs(assignment.Executable) || !filepath.IsAbs(assignment.Root) || assignment.Model == "" || !assignment.Effort.Valid() || assignment.Prompt == "" || len(assignment.Prompt) > 1<<20 {
		return Result{}, errors.New("Codex app-server assignment is invalid")
	}
	physical, err := filepath.EvalSymlinks(assignment.Root)
	if err != nil || physical != filepath.Clean(assignment.Root) {
		return Result{}, errors.New("Codex app-server root is unsafe")
	}
	executable, err := processadapter.Resolve(assignment.Executable)
	if err != nil || executable.Path != assignment.Executable {
		return Result{}, errors.New("Codex executable identity is unavailable")
	}
	command := exec.CommandContext(ctx, executable.Path, "app-server")
	command.Dir = physical
	command.Env = processadapter.MinimalEnvironment()
	command.WaitDelay = time.Second
	stdin, err := command.StdinPipe()
	if err != nil {
		return Result{}, errors.New("cannot open Codex app-server input")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return Result{}, errors.New("cannot open Codex app-server output")
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return Result{}, errors.New("cannot open Codex app-server diagnostic")
	}
	if err := command.Start(); err != nil {
		return Result{}, errors.New("cannot start Codex app-server")
	}
	diagnostic := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(io.LimitReader(stderr, maxProtocolData+1))
		if readErr != nil || len(data) > maxProtocolData {
			diagnostic <- errors.New("Codex app-server diagnostic exceeded bounds")
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
	total := 0
	send := func(value any) error {
		if err := encoder.Encode(value); err != nil {
			return errors.New("write Codex app-server request")
		}
		return nil
	}
	read := func() (message, error) {
		if !scanner.Scan() {
			if scanner.Err() != nil {
				return message{}, errors.New("Codex app-server framing exceeded limits")
			}
			return message{}, errors.New("Codex app-server closed before completion")
		}
		total += len(scanner.Bytes())
		if total > maxProtocolData {
			return message{}, errors.New("Codex app-server output exceeded bounds")
		}
		var value message
		if json.Unmarshal(scanner.Bytes(), &value) != nil {
			return message{}, errors.New("Codex app-server emitted malformed JSON")
		}
		if len(value.ID) > 0 && value.Method != "" {
			return message{}, errors.New("Codex app-server requested unsupported host authority")
		}
		return value, nil
	}
	waitID := func(id string) (message, error) {
		for {
			value, readErr := read()
			if readErr != nil {
				return value, readErr
			}
			if string(value.ID) != id {
				continue
			}
			if value.Error != nil {
				return value, fmt.Errorf("Codex app-server request %s failed", id)
			}
			return value, nil
		}
	}
	if err := send(map[string]any{"method": "initialize", "id": 1, "params": map[string]any{"clientInfo": map[string]any{"name": "level7", "title": "Level 7", "version": "1.0.0"}}}); err != nil {
		return Result{}, err
	}
	if _, err := waitID("1"); err != nil {
		return Result{}, err
	}
	if err := send(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return Result{}, err
	}
	threadMethod := "thread/start"
	threadParams := map[string]any{"model": assignment.Model, "cwd": physical, "approvalPolicy": "never", "sandbox": sandboxName(assignment.Reviewer), "serviceName": "level7"}
	if assignment.SessionID != "" {
		threadMethod = "thread/resume"
		threadParams = map[string]any{"threadId": assignment.SessionID, "model": assignment.Model, "cwd": physical, "approvalPolicy": "never", "sandbox": sandboxName(assignment.Reviewer)}
	}
	if err := send(map[string]any{"method": threadMethod, "id": 2, "params": threadParams}); err != nil {
		return Result{}, err
	}
	threadResponse, err := waitID("2")
	if err != nil {
		return Result{}, err
	}
	thread, _ := threadResponse.Result["thread"].(map[string]any)
	threadID, _ := thread["id"].(string)
	if threadID == "" || len(threadID) > 256 || strings.ContainsAny(threadID, "\x00\r\n") {
		return Result{}, errors.New("Codex app-server returned an invalid thread")
	}
	policy := sandboxPolicy(physical, assignment.Reviewer)
	turnParams := map[string]any{
		"threadId": threadID, "input": []any{map[string]any{"type": "text", "text": assignment.Prompt}},
		"cwd": physical, "approvalPolicy": "never", "sandboxPolicy": policy, "model": assignment.Model,
		"effort": assignment.Effort, "summary": "concise",
		"outputSchema": terminalOutputSchema(assignment.Reviewer),
	}
	if err := send(map[string]any{"method": "turn/start", "id": 3, "params": turnParams}); err != nil {
		return Result{}, err
	}
	turnResponse, err := waitID("3")
	if err != nil {
		return Result{SessionID: threadID}, err
	}
	turn, _ := turnResponse.Result["turn"].(map[string]any)
	turnID, _ := turn["id"].(string)
	result := Result{SessionID: threadID, TurnID: turnID, Status: "inProgress"}
	for {
		value, readErr := read()
		if readErr != nil {
			return result, readErr
		}
		if value.Method == "item/completed" {
			item, _ := value.Params["item"].(map[string]any)
			if item["type"] == "agentMessage" {
				if text, ok := item["text"].(string); ok && len(text) <= 16<<10 {
					result.Summary = text
				}
			}
		}
		if value.Method != "turn/completed" {
			continue
		}
		completed, _ := value.Params["turn"].(map[string]any)
		status, _ := completed["status"].(string)
		result.Status = status
		if status == "completed" {
			terminal, terminalErr := parseTerminal(result.Summary, assignment.Reviewer)
			if terminalErr != nil {
				return result, terminalErr
			}
			result.Summary, result.Decision = terminal.Summary, terminal.Decision
			return result, nil
		}
		failure, _ := completed["error"].(map[string]any)
		result.FailureCode = codexFailureCode(failure)
		if result.FailureCode == "quota" {
			reset, quotaErr := rateLimitReset(send, read)
			if quotaErr == nil {
				result.QuotaReset = reset
			}
		}
		return result, errors.New("Codex turn did not complete successfully")
	}
}

type terminalResult struct {
	Outcome  string                `json:"outcome"`
	Summary  string                `json:"summary"`
	Decision domain.ReviewDecision `json:"decision,omitempty"`
}

func terminalOutputSchema(reviewer bool) map[string]any {
	properties := map[string]any{
		"outcome": map[string]any{"type": "string", "enum": []string{"complete", "blocked"}},
		"summary": map[string]any{"type": "string"},
	}
	required := []string{"outcome", "summary"}
	if reviewer {
		properties["decision"] = map[string]any{"type": "string", "enum": []string{"GO", "NO_GO"}}
		required = append(required, "decision")
	}
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}

func parseTerminal(value string, reviewer bool) (terminalResult, error) {
	var terminal terminalResult
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	if value == "" || decoder.Decode(&terminal) != nil || (terminal.Outcome != "complete" && terminal.Outcome != "blocked") || terminal.Summary == "" || len(terminal.Summary) > 4096 {
		return terminal, errors.New("Codex terminal result is invalid")
	}
	if reviewer {
		if !terminal.Decision.Valid() || (terminal.Outcome == "blocked" && terminal.Decision == domain.DecisionGO) {
			return terminal, errors.New("Codex reviewer decision is invalid")
		}
	} else if terminal.Decision != "" {
		return terminal, errors.New("Codex implementer attempted to review itself")
	}
	return terminal, nil
}

func sandboxName(reviewer bool) string {
	if reviewer {
		return "readOnly"
	}
	return "workspaceWrite"
}

func sandboxPolicy(root string, reviewer bool) map[string]any {
	if reviewer {
		return map[string]any{"type": "readOnly", "networkAccess": false}
	}
	return map[string]any{
		"type": "workspaceWrite", "writableRoots": []string{root}, "networkAccess": false,
		"excludeSlashTmp": true, "excludeTmpdirEnvVar": true,
	}
}

func codexFailureCode(value map[string]any) string {
	data, _ := json.Marshal(value)
	lower := strings.ToLower(string(data))
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "usage limit") || strings.Contains(lower, "quota") {
		return "quota"
	}
	return "provider"
}

func rateLimitReset(send func(any) error, read func() (message, error)) (string, error) {
	if err := send(map[string]any{"method": "account/rateLimits/read", "id": 4}); err != nil {
		return "", err
	}
	for {
		value, err := read()
		if err != nil {
			return "", err
		}
		if string(value.ID) != "4" {
			continue
		}
		limits, _ := value.Result["rateLimits"].(map[string]any)
		var latest int64
		for _, key := range []string{"primary", "secondary"} {
			window, _ := limits[key].(map[string]any)
			if raw, ok := window["resetsAt"].(float64); ok && int64(raw) > latest {
				latest = int64(raw)
			}
		}
		if latest < 1 {
			return "", errors.New("Codex rate limit reset is unavailable")
		}
		return time.Unix(latest, 0).UTC().Format(time.RFC3339), nil
	}
}
