package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// SessionAssignment is the version-independent Claude Code execution contract
// used by the v1 orchestrator. Candidate flags and models are verified during
// the explicit provider probe rather than admitted by an exact version pin.
type SessionAssignment struct {
	Executable string
	Root       string
	Model      string
	Effort     domain.ReasoningEffort
	Prompt     string
	SessionID  string
	Reviewer   bool
}

type SessionResult struct {
	SessionID  string
	Summary    string
	Decision   domain.ReviewDecision
	Quota      bool
	QuotaReset string
}

func RunSession(ctx context.Context, assignment SessionAssignment) (SessionResult, error) {
	if assignment.Executable == "" || assignment.Root == "" || assignment.Model == "" || !assignment.Effort.Valid() || assignment.Prompt == "" || len(assignment.Prompt) > 1<<20 {
		return SessionResult{}, errors.New("Claude session assignment is invalid")
	}
	executable, err := processadapter.Resolve(assignment.Executable)
	if err != nil || executable.Path != assignment.Executable {
		return SessionResult{}, errors.New("Claude executable identity is unavailable")
	}
	arguments := sessionArguments(assignment)
	result, runErr := (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: arguments, Directory: assignment.Root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 20, Timeout: 24 * time.Hour,
	})
	rawCombined := string(append(append([]byte{}, result.Stdout...), result.Stderr...))
	combined := strings.ToLower(rawCombined)
	if runErr != nil || result.ExitCode != 0 {
		quota := strings.Contains(combined, "rate limit") || strings.Contains(combined, "usage limit") || strings.Contains(combined, "quota")
		reset := ""
		if quota {
			reset = reportedQuotaReset(rawCombined, time.Now())
		}
		return SessionResult{Quota: quota, QuotaReset: reset}, errors.New("Claude session did not complete")
	}
	data := bytes.TrimSpace(result.Stdout)
	if len(data) < 2 || !json.Valid(data) {
		return SessionResult{}, errors.New("Claude session returned malformed JSON")
	}
	var envelope struct {
		Type             string          `json:"type"`
		Subtype          string          `json:"subtype"`
		IsError          bool            `json:"is_error"`
		SessionID        string          `json:"session_id"`
		Result           string          `json:"result"`
		StructuredOutput json.RawMessage `json:"structured_output"`
	}
	if json.Unmarshal(data, &envelope) != nil || envelope.Type != "result" || envelope.IsError || envelope.SessionID == "" || len(envelope.SessionID) > 256 || strings.ContainsAny(envelope.SessionID, "\x00\r\n") {
		return SessionResult{}, errors.New("Claude session result is invalid")
	}
	terminal, err := parseSessionTerminal(envelope.StructuredOutput, assignment.Reviewer)
	if err != nil {
		return SessionResult{SessionID: envelope.SessionID}, err
	}
	return SessionResult{SessionID: envelope.SessionID, Summary: terminal.Summary, Decision: terminal.Decision}, nil
}

func sessionArguments(assignment SessionAssignment) []string {
	permission, tools := "acceptEdits", "Read,Glob,Grep,Edit,Write"
	if assignment.Reviewer {
		permission, tools = "plan", "Read,Glob,Grep"
	}
	arguments := []string{
		"--safe-mode", "--disable-slash-commands", "--print",
		"--model", assignment.Model, "--effort", string(assignment.Effort),
		"--permission-mode", permission, "--tools", tools,
		"--disallowedTools", "Bash,WebFetch,WebSearch,NotebookEdit,Task,Skill,Agent,Chrome",
		"--no-chrome", "--output-format", "json", "--max-turns", "64",
		"--json-schema", terminalSchema,
	}
	if assignment.SessionID != "" {
		arguments = append(arguments, "--resume", assignment.SessionID)
	}
	return append(arguments, assignment.Prompt)
}

func reportedQuotaReset(value string, now time.Time) string {
	for _, field := range strings.Fields(value) {
		candidate := strings.Trim(field, "\"'()[]{}<>,;.")
		if parsed, err := time.Parse(time.RFC3339, candidate); err == nil && parsed.After(now) {
			return parsed.UTC().Format(time.RFC3339)
		}
	}
	fields := strings.Fields(strings.ToLower(value))
	for index := 0; index+3 < len(fields); index++ {
		if fields[index] != "retry" || fields[index+1] != "after" {
			continue
		}
		count, err := time.ParseDuration(fields[index+2] + durationSuffix(fields[index+3]))
		if err == nil && count > 0 && count <= 7*24*time.Hour {
			return now.UTC().Add(count).Format(time.RFC3339)
		}
	}
	return ""
}

func durationSuffix(value string) string {
	value = strings.Trim(value, "\"'()[]{}<>,;.")
	switch value {
	case "second", "seconds", "sec", "secs":
		return "s"
	case "minute", "minutes", "min", "mins":
		return "m"
	case "hour", "hours", "hr", "hrs":
		return "h"
	default:
		return value
	}
}

type sessionTerminal struct {
	Schema   int                   `json:"schema"`
	Outcome  string                `json:"outcome"`
	Summary  string                `json:"summary"`
	Findings []string              `json:"findings"`
	Decision domain.ReviewDecision `json:"decision,omitempty"`
}

func parseSessionTerminal(data []byte, reviewer bool) (sessionTerminal, error) {
	var terminal sessionTerminal
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if len(data) < 2 || decoder.Decode(&terminal) != nil || terminal.Schema != 1 || (terminal.Outcome != "complete" && terminal.Outcome != "blocked") || terminal.Summary == "" || len(terminal.Summary) > 4096 || len(terminal.Findings) > 64 {
		return terminal, errors.New("Claude structured terminal result is invalid")
	}
	for _, finding := range terminal.Findings {
		if finding == "" || len(finding) > 2048 || strings.ContainsAny(finding, "\x00\r\n") {
			return terminal, errors.New("Claude structured finding is invalid")
		}
	}
	if reviewer {
		if !terminal.Decision.Valid() || (terminal.Outcome == "blocked" && terminal.Decision == domain.DecisionGO) {
			return terminal, errors.New("Claude reviewer decision is invalid")
		}
	} else if terminal.Decision != "" {
		return terminal, errors.New("Claude implementer attempted to review itself")
	}
	return terminal, nil
}
