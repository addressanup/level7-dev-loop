package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	armCrew  = "crew"
	armPlain = "plain"

	claimComplete = "complete"
	claimBlocked  = "blocked"
	claimNone     = "none"
	claimUnparsed = "unparsed"

	crewTarget    = "l7/crew"
	maxWaitSlice  = 120
	stopDeadline  = 90 * time.Second
	finalAnswerSc = `{"type":"object","properties":{"outcome":{"type":"string","enum":["complete","blocked"]},"summary":{"type":"string"}},"required":["outcome","summary"],"additionalProperties":false}`
)

type route struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Effort   string `json:"effort"`
	Reviewer string `json:"expected_reviewer,omitempty"`
}

type trialRecord struct {
	Task        string       `json:"task"`
	Arm         string       `json:"arm"`
	Round       int          `json:"round"`
	Sequence    int          `json:"sequence"`
	StartedUTC  string       `json:"started_utc"`
	WallSeconds float64      `json:"wall_seconds"`
	Directory   string       `json:"directory"`
	BaseCommit  string       `json:"base_commit"`
	Delivered   string       `json:"delivered_commit"`
	Claim       string       `json:"claim"`
	Terminal    string       `json:"terminal"`
	Model       string       `json:"model"`
	Outcome     string       `json:"outcome"`
	Invalid     string       `json:"invalid,omitempty"`
	Grade       gradeResult  `json:"grade"`
	Meter       meterSummary `json:"meter"`
}

func (record trialRecord) claimed() bool {
	return record.Claim == claimComplete || record.Claim == claimUnparsed
}

// harness holds what every trial of one invocation shares.
type harness struct {
	corpus  corpus
	l7      string
	codex   string
	meter   string
	ledger  string
	work    string
	temp    string
	pathEnv string
	route   route
	judge   grader
	l7Run   func(ctx context.Context, directory string, timeout time.Duration, arguments ...string) (envelope, error)
}

type envelope struct {
	Outcome string          `json:"outcome"`
	Code    string          `json:"code"`
	State   string          `json:"state"`
	Version string          `json:"version"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// environment is what Level 7 and both arms' agents see: the meter first on
// PATH and a private TMPDIR, the same for every trial.
func (h *harness) environment() []string {
	environment := baseEnvironment(h.pathEnv)
	if h.temp == "" {
		return environment
	}
	environment = slices.DeleteFunc(environment, func(entry string) bool { return strings.HasPrefix(entry, "TMPDIR=") })
	return append(environment, "TMPDIR="+h.temp)
}

func (h *harness) callL7(ctx context.Context, directory string, timeout time.Duration, arguments ...string) (envelope, error) {
	if h.l7Run != nil {
		return h.l7Run(ctx, directory, timeout, arguments...)
	}
	result, err := execute(ctx, commandSpec{Dir: directory, Env: h.environment(), Name: h.l7, Args: append(slices.Clone(arguments), "--json"), Timeout: timeout})
	label := "l7 " + strings.Join(arguments[:min(2, len(arguments))], " ")
	if err != nil {
		return envelope{}, fmt.Errorf("%s: %w", label, err)
	}
	if result.TimedOut {
		return envelope{}, fmt.Errorf("%s timed out", label)
	}
	var value envelope
	if err := json.Unmarshal(bytes.TrimSpace(result.Stdout), &value); err != nil || value.Outcome == "" {
		return envelope{}, fmt.Errorf("%s returned no JSON envelope: %s", label, tail(result.Stderr))
	}
	if value.Outcome != "PASS" {
		return value, fmt.Errorf("%s: %s %s", label, value.Outcome, value.Message)
	}
	return value, nil
}

func (h *harness) newTrial(value task, arm string, round, sequence int) (trialRecord, string, error) {
	directory := filepath.Join(h.work, "trials", value.ID, fmt.Sprintf("r%d-%s", round, arm))
	record := trialRecord{Task: value.ID, Arm: arm, Round: round, Sequence: sequence, StartedUTC: time.Now().UTC().Format(time.RFC3339), Directory: directory, Claim: claimNone}
	return record, directory, os.MkdirAll(directory, 0o755)
}

// finish grades the delivered commit, attaches the meter totals, and
// classifies the trial.
func (h *harness) finish(ctx context.Context, record trialRecord, value task, repository string, started time.Time, invalid string) trialRecord {
	record.WallSeconds = time.Since(started).Seconds()
	if summary, err := summarizeLedger(h.ledger, record.Directory); err == nil {
		record.Meter = summary
	} else if invalid == "" {
		invalid = "meter: " + err.Error()
	}
	if invalid == "" && record.Delivered != "" {
		files, irregular, err := treeFiles(ctx, repository, record.Delivered)
		if err != nil {
			record.Grade.ForbiddenEffects = append(record.Grade.ForbiddenEffects, "unreadable delivered tree: "+err.Error())
		} else {
			checkout := record.Grade.CheckoutFindings
			record.Grade = h.judge.grade(ctx, value, files, irregular, filepath.Join(record.Directory, "grading"))
			record.Grade.CheckoutFindings = checkout
		}
	}
	if invalid == "" && ctx.Err() != nil {
		invalid = "interrupted"
	}
	record.Outcome, record.Invalid = classify(record.claimed(), record.Grade, invalid)
	return record
}

func plainPrompt(value task) string {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Task: %s\n\n%s\n\nAcceptance criteria:\n", value.Title, value.Objective)
	for _, criterion := range value.Acceptance {
		fmt.Fprintf(&prompt, "- %s\n", criterion)
	}
	prompt.WriteString("\nYou may change only these paths:\n")
	for _, scope := range value.Paths {
		fmt.Fprintf(&prompt, "- %s\n", scope)
	}
	prompt.WriteString("\nCheck your work with:\n")
	for _, argv := range value.Verify {
		fmt.Fprintf(&prompt, "- %s\n", strings.Join(argv, " "))
	}
	prompt.WriteString("\nWhen you finish, answer with JSON: outcome is \"complete\" when the task is done and \"blocked\" otherwise, and summary says what you changed and why.\n")
	return prompt.String()
}

func crewObjective(value task) string {
	var objective strings.Builder
	fmt.Fprintf(&objective, "## Ship: %s\n\n%s\n\nPaths: %s\n", value.Title, value.Objective, strings.Join(value.Paths, ", "))
	for _, argv := range value.Verify {
		encoded, _ := json.Marshal(argv)
		fmt.Fprintf(&objective, "Verify: %s\n", encoded)
	}
	for _, criterion := range value.Acceptance {
		fmt.Fprintf(&objective, "Acceptance: %s\n", criterion)
	}
	return objective.String()
}

func (h *harness) runPlain(ctx context.Context, value task, round, sequence int) trialRecord {
	started := time.Now()
	record, directory, err := h.newTrial(value, armPlain, round, sequence)
	repository := filepath.Join(directory, "repo")
	if err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	if record.BaseCommit, err = materializeRepository(ctx, repository, value.repo); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	schema, answer := filepath.Join(directory, "final-answer.schema.json"), filepath.Join(directory, "final-answer.json")
	if err := os.WriteFile(schema, []byte(finalAnswerSc+"\n"), 0o644); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	record.Model = h.route.Provider + "/" + h.route.Model + "@" + h.route.Effort
	arguments := []string{
		"exec", "--json", "--color", "never", "-m", h.route.Model,
		"-c", "model_reasoning_effort=" + strconv.Quote(h.route.Effort), "-c", `approval_policy="never"`,
		"-s", "workspace-write", "-c", "sandbox_workspace_write.network_access=false",
		"-c", "sandbox_workspace_write.exclude_slash_tmp=true", "-c", "sandbox_workspace_write.exclude_tmpdir_env_var=true",
		"--output-schema", schema, "-o", answer, "-C", repository, "-",
	}
	timeout := time.Duration(h.corpus.protocol.Timeouts.Plain) * time.Second
	result, runErr := execute(ctx, commandSpec{Dir: repository, Env: h.environment(), Name: h.meter, Args: arguments, Stdin: []byte(plainPrompt(value)), Timeout: timeout})
	invalid := ""
	switch {
	case runErr != nil:
		invalid = "codex exec: " + runErr.Error()
	case result.TimedOut:
		record.Terminal = "timeout"
	case quotaFailure(result.Stdout) || quotaFailure(result.Stderr):
		invalid = "quota: codex exec reported a usage or rate limit"
	case result.ExitCode != 0:
		record.Terminal = fmt.Sprintf("exit %d: %s", result.ExitCode, tail(result.Stderr))
	default:
		record.Terminal, record.Claim = "exit 0", readClaim(answer, result.Stdout)
	}
	if invalid == "" {
		if record.Delivered, err = snapshotWorktree(ctx, repository); err != nil {
			invalid = "snapshot: " + err.Error()
		}
	}
	return h.finish(ctx, record, value, repository, started, invalid)
}

// readClaim takes the structured final answer from codex exec's last-message
// file, falling back to the last agent message in the event stream.
func readClaim(answerPath string, events []byte) string {
	text, err := os.ReadFile(answerPath)
	if err != nil || len(bytes.TrimSpace(text)) == 0 {
		text = nil
		for _, line := range bytes.Split(events, []byte("\n")) {
			var event struct {
				Type string `json:"type"`
				Item struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"item"`
			}
			if json.Unmarshal(line, &event) == nil && event.Type == "item.completed" && event.Item.Type == "agent_message" {
				text = []byte(event.Item.Text)
			}
		}
	}
	var answer struct {
		Outcome string `json:"outcome"`
		Summary string `json:"summary"`
	}
	if json.Unmarshal(bytes.TrimSpace(text), &answer) != nil || (answer.Outcome != claimComplete && answer.Outcome != claimBlocked) {
		return claimUnparsed
	}
	return answer.Outcome
}

func quotaFailure(output []byte) bool {
	for _, line := range bytes.Split(output, []byte("\n")) {
		var event struct {
			Type    string          `json:"type"`
			Message string          `json:"message"`
			Error   json.RawMessage `json:"error"`
		}
		if json.Unmarshal(line, &event) != nil || (event.Type != "error" && event.Type != "turn.failed") {
			continue
		}
		lower := strings.ToLower(event.Message + string(event.Error))
		if strings.Contains(lower, "usage limit") || strings.Contains(lower, "rate limit") || strings.Contains(lower, "quota") {
			return true
		}
	}
	return false
}

type crewTaskView struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	Attempt      int    `json:"attempt"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	Candidate    string `json:"candidate"`
	Verification string `json:"verification"`
	Message      string `json:"message"`
}

type crewStatusView struct {
	PlanID        string         `json:"plan_id"`
	Supervisor    bool           `json:"supervisor_running"`
	Tasks         []crewTaskView `json:"tasks"`
	OpenDecisions []struct {
		Kind string `json:"kind"`
	} `json:"open_decisions"`
	Token string `json:"token"`
}

func terminalCrewState(state string) bool {
	switch state {
	case "done", "cancelled", "needs-decision", "paused":
		return true
	}
	return false
}

func (h *harness) runCrew(ctx context.Context, value task, round, sequence int) trialRecord {
	started := time.Now()
	record, directory, err := h.newTrial(value, armCrew, round, sequence)
	repository := filepath.Join(directory, "repo")
	if err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	if record.BaseCommit, err = materializeRepository(ctx, repository, value.repo); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	if err := h.prepareCrew(ctx, repository); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	objective := filepath.Join(repository, ".git", "l7", "crew", "objectives", value.ID+".md")
	if err := os.MkdirAll(filepath.Dir(objective), 0o700); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	if err := os.WriteFile(objective, []byte(crewObjective(value)), 0o600); err != nil {
		return h.finish(ctx, record, value, repository, started, "setup: "+err.Error())
	}
	plan, err := h.callL7(ctx, repository, time.Minute, "crew", "plan", "--objective", objective)
	if err != nil {
		return h.finish(ctx, record, value, repository, started, err.Error())
	}
	var planned struct {
		ID     string `json:"id"`
		Digest string `json:"digest"`
		Tasks  []struct {
			AllowedPaths []string   `json:"allowed_paths"`
			Verification [][]string `json:"verification"`
		} `json:"tasks"`
	}
	if json.Unmarshal(plan.Data, &planned) != nil || planned.ID == "" || len(planned.Tasks) != 1 ||
		!slices.Equal(planned.Tasks[0].AllowedPaths, value.Paths) || !equalCommands(planned.Tasks[0].Verification, value.Verify) {
		return h.finish(ctx, record, value, repository, started, "crew plan does not match the task's scope and checks")
	}
	if _, err := h.callL7(ctx, repository, time.Minute, "crew", "start", "--plan", planned.ID, "--digest", planned.Digest, "--owner", "outcome-eval", "--role", "evaluator", "--confirm"); err != nil {
		return h.finish(ctx, record, value, repository, started, err.Error())
	}
	status, timedOut, quota, waitErr := h.awaitCrew(ctx, repository, started.Add(time.Duration(h.corpus.protocol.Timeouts.Crew)*time.Second))
	if err := h.stopCrew(context.WithoutCancel(ctx), repository); err != nil && waitErr == nil {
		waitErr = err
	}
	if waitErr != nil {
		return h.finish(ctx, record, value, repository, started, waitErr.Error())
	}
	if len(status.Tasks) != 1 {
		return h.finish(ctx, record, value, repository, started, "crew status does not report exactly one task")
	}
	final := status.Tasks[0]
	record.Model, record.Terminal = final.Provider+"/"+final.Model, final.State
	for _, open := range status.OpenDecisions {
		record.Terminal += ":" + open.Kind
	}
	invalid := ""
	switch {
	case final.State == "done":
		record.Claim = claimComplete
	case quota:
		invalid = "quota: the crew waited for a provider quota reset"
	case timedOut:
		record.Terminal, record.Claim = "timeout ("+final.State+")", claimNone
	case !terminalCrewState(final.State):
		invalid = "crew supervisor stopped while the task was " + final.State
	default:
		record.Claim = claimBlocked
	}
	record.Delivered = record.BaseCommit
	if delivered, err := git(ctx, repository, "rev-parse", "--verify", "--quiet", "refs/heads/"+crewTarget); err == nil && delivered != "" {
		record.Delivered = delivered
	}
	record.Grade.CheckoutFindings = checkoutFindings(ctx, repository, record.BaseCommit)
	return h.finish(ctx, record, value, repository, started, invalid)
}

// prepareCrew enables the crew in a fresh repository: the orchestration
// policy from onboard, features.crew on, Claude disabled, and a probe of the
// metered Codex.
func (h *harness) prepareCrew(ctx context.Context, repository string) error {
	if _, err := h.callL7(ctx, repository, time.Minute, "onboard", "--apply"); err != nil {
		return err
	}
	if err := enableCrewPolicy(filepath.Join(repository, ".l7", "orchestration.json")); err != nil {
		return err
	}
	probe, err := h.callL7(ctx, repository, 2*time.Minute, "providers", "probe")
	if err != nil {
		return err
	}
	var snapshots []struct {
		ID             string `json:"id"`
		Executable     string `json:"executable"`
		Authentication string `json:"authentication"`
		Models         []struct {
			Verified bool `json:"verified"`
		} `json:"models"`
	}
	if err := json.Unmarshal(probe.Data, &snapshots); err != nil {
		return errors.New("providers probe returned no snapshots")
	}
	for _, snapshot := range snapshots {
		if snapshot.ID != "codex-local" {
			continue
		}
		verified := 0
		for _, model := range snapshot.Models {
			if model.Verified {
				verified++
			}
		}
		if snapshot.Executable != h.meter || snapshot.Authentication != "authenticated" || verified < 2 {
			return fmt.Errorf("probed Codex is not the authenticated meter with two verified models (executable %s, %s, %d models)", snapshot.Executable, snapshot.Authentication, verified)
		}
		return nil
	}
	return errors.New("providers probe did not report codex-local")
}

func enableCrewPolicy(name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var policy map[string]any
	if err := decoder.Decode(&policy); err != nil {
		return fmt.Errorf("orchestration policy: %w", err)
	}
	features, ok := policy["features"].(map[string]any)
	providers, listed := policy["providers"].([]any)
	if !ok || !listed {
		return errors.New("orchestration policy lacks features or providers")
	}
	features["crew"] = true
	for _, entry := range providers {
		if provider, ok := entry.(map[string]any); ok && provider["id"] == "claude-local" {
			provider["enabled"] = false
		}
	}
	encoded, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(name, append(encoded, '\n'))
}

func (h *harness) crewStatus(ctx context.Context, repository string) (crewStatusView, error) {
	value, err := h.callL7(ctx, repository, time.Minute, "crew", "status")
	if err != nil {
		return crewStatusView{}, err
	}
	var status crewStatusView
	if err := json.Unmarshal(value.Data, &status); err != nil {
		return crewStatusView{}, errors.New("crew status data is malformed")
	}
	return status, nil
}

// awaitCrew waits until the task is done, needs a decision, is cancelled or
// paused, the supervisor stops, or the deadline passes. It never answers a
// decision.
func (h *harness) awaitCrew(ctx context.Context, repository string, deadline time.Time) (crewStatusView, bool, bool, error) {
	token, quota := "", false
	for {
		remaining := int(time.Until(deadline).Seconds())
		if remaining <= 0 {
			status, err := h.crewStatus(ctx, repository)
			return status, true, quota, err
		}
		arguments := []string{"crew", "wait", "--timeout", strconv.Itoa(min(maxWaitSlice, remaining))}
		if token != "" {
			arguments = append(arguments, "--since", token)
		}
		value, err := h.callL7(ctx, repository, time.Duration(min(maxWaitSlice, remaining)+60)*time.Second, arguments...)
		if err != nil {
			return crewStatusView{}, false, quota, err
		}
		var waited struct {
			Reason string         `json:"reason"`
			Status crewStatusView `json:"status"`
		}
		if json.Unmarshal(value.Data, &waited) != nil || len(waited.Status.Tasks) != 1 {
			return crewStatusView{}, false, quota, errors.New("crew wait data is malformed")
		}
		status := waited.Status
		token = status.Token
		state := status.Tasks[0].State
		quota = quota || state == "waiting-quota"
		if terminalCrewState(state) || !status.Supervisor {
			return status, false, quota, nil
		}
	}
}

// stopCrew cancels a still-running supervisor and waits until it has exited,
// so no worker keeps editing the trial while it is graded. The caller keeps
// the status it observed before stopping, so an open decision is reported
// rather than the cancellation.
func (h *harness) stopCrew(ctx context.Context, repository string) error {
	deadline, cancelled := time.Now().Add(stopDeadline), false
	for {
		current, err := h.crewStatus(ctx, repository)
		if err != nil {
			return err
		}
		if !current.Supervisor {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("crew supervisor did not stop")
		}
		if !cancelled {
			if _, err := h.callL7(ctx, repository, 2*time.Minute, "crew", "cancel"); err != nil {
				return err
			}
			cancelled = true
		}
		time.Sleep(time.Second)
	}
}

// checkoutFindings reports any change the crew made to the user's checkout:
// the branch, its commit, the index, or the working tree.
func checkoutFindings(ctx context.Context, repository, base string) []string {
	findings := []string{}
	if branch, err := git(ctx, repository, "symbolic-ref", "-q", "HEAD"); err != nil || branch != "refs/heads/main" {
		findings = append(findings, "checkout branch changed to "+branch)
	}
	if head, err := git(ctx, repository, "rev-parse", "refs/heads/main"); err != nil || head != base {
		findings = append(findings, "main moved to "+head)
	}
	if status, err := git(ctx, repository, "status", "--porcelain=v1", "--untracked-files=all"); err != nil || status != "" {
		findings = append(findings, "checkout has changes: "+strings.ReplaceAll(status, "\n", "; "))
	}
	return findings
}

func equalCommands(left, right [][]string) bool {
	return slices.EqualFunc(left, right, func(a, b []string) bool { return slices.Equal(a, b) })
}
