package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	meterName       = "codex"
	meterConfigName = "codex.meter.json"
	maxMeterLine    = 64 << 20
)

// meterConfig sits next to the meter binary. Level 7 starts providers with a
// minimal environment, so the meter cannot rely on environment variables.
type meterConfig struct {
	Codex  string `json:"codex,omitempty"`
	Ledger string `json:"ledger"`
	Fake   bool   `json:"fake,omitempty"`
}

type usage struct {
	Input           int64 `json:"input"`
	CachedInput     int64 `json:"cached_input"`
	Output          int64 `json:"output"`
	ReasoningOutput int64 `json:"reasoning_output"`
	Total           int64 `json:"total"`
}

func (value *usage) add(other usage) {
	value.Input += other.Input
	value.CachedInput += other.CachedInput
	value.Output += other.Output
	value.ReasoningOutput += other.ReasoningOutput
	value.Total += other.Total
}

// meterRecord is one metered codex process. Usage counts only turns that
// started in this process, so a resumed thread's earlier turns are not
// counted twice.
type meterRecord struct {
	Schema       int    `json:"schema"`
	PID          int    `json:"pid"`
	Mode         string `json:"mode"`
	Directory    string `json:"directory"`
	StartedUTC   string `json:"started_utc"`
	UpdatedUTC   string `json:"updated_utc"`
	Turns        int    `json:"turns"`
	Usage        usage  `json:"usage"`
	UsageEvents  int    `json:"usage_events"`
	Unattributed int    `json:"unattributed_usage_events"`
	Malformed    int    `json:"malformed_lines"`
	Exited       bool   `json:"exited"`
	ExitCode     int    `json:"exit_code"`
}

type meterTap struct {
	record  meterRecord
	started map[string]bool
}

func newMeterTap(arguments []string, directory string) *meterTap {
	mode := "other"
	if len(arguments) > 0 && (arguments[0] == "app-server" || arguments[0] == "exec") {
		mode = arguments[0]
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return &meterTap{
		record:  meterRecord{Schema: 1, PID: os.Getpid(), Mode: mode, Directory: directory, StartedUTC: now, UpdatedUTC: now},
		started: map[string]bool{},
	}
}

type appServerLine struct {
	Method string `json:"method"`
	Result *struct {
		Turn *struct {
			ID string `json:"id"`
		} `json:"turn"`
	} `json:"result"`
	Params *struct {
		TurnID string `json:"turnId"`
		Turn   *struct {
			ID string `json:"id"`
		} `json:"turn"`
		TokenUsage *struct {
			Last *struct {
				Input           int64 `json:"inputTokens"`
				CachedInput     int64 `json:"cachedInputTokens"`
				Output          int64 `json:"outputTokens"`
				ReasoningOutput int64 `json:"reasoningOutputTokens"`
				Total           int64 `json:"totalTokens"`
			} `json:"last"`
		} `json:"tokenUsage"`
	} `json:"params"`
}

type execLine struct {
	Type  string `json:"type"`
	Usage *struct {
		Input           int64 `json:"input_tokens"`
		CachedInput     int64 `json:"cached_input_tokens"`
		Output          int64 `json:"output_tokens"`
		ReasoningOutput int64 `json:"reasoning_output_tokens"`
	} `json:"usage"`
}

// observe records turns and usage from one output line and reports whether
// the record changed.
func (tap *meterTap) observe(line []byte) bool {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return false
	}
	record := &tap.record
	switch record.Mode {
	case "app-server":
		var message appServerLine
		if json.Unmarshal(line, &message) != nil {
			record.Malformed++
			return true
		}
		if message.Result != nil && message.Result.Turn != nil && message.Result.Turn.ID != "" {
			tap.started[message.Result.Turn.ID] = true
		}
		if message.Params == nil {
			return false
		}
		switch message.Method {
		case "turn/started":
			if message.Params.Turn != nil {
				tap.started[message.Params.Turn.ID] = true
			}
		case "turn/completed":
			if message.Params.Turn != nil && tap.started[message.Params.Turn.ID] {
				record.Turns++
				return true
			}
		case "thread/tokenUsage/updated":
			if message.Params.TokenUsage == nil || message.Params.TokenUsage.Last == nil {
				record.Malformed++
				return true
			}
			if !tap.started[message.Params.TurnID] {
				record.Unattributed++
				return true
			}
			last := message.Params.TokenUsage.Last
			record.Usage.add(usage{Input: last.Input, CachedInput: last.CachedInput, Output: last.Output, ReasoningOutput: last.ReasoningOutput, Total: last.Total})
			record.UsageEvents++
			return true
		}
	case "exec":
		var event execLine
		if json.Unmarshal(line, &event) != nil {
			record.Malformed++
			return true
		}
		if event.Type == "turn.completed" {
			record.Turns++
			if event.Usage != nil {
				used := event.Usage
				record.Usage.add(usage{Input: used.Input, CachedInput: used.CachedInput, Output: used.Output, ReasoningOutput: used.ReasoningOutput, Total: used.Input + used.Output})
				record.UsageEvents++
			}
			return true
		}
	}
	return false
}

type meterLedger struct {
	path string
}

func newMeterLedger(directory string, record meterRecord) (meterLedger, error) {
	if !filepath.IsAbs(directory) {
		return meterLedger{}, errors.New("meter ledger directory must be absolute")
	}
	ledger := meterLedger{path: filepath.Join(directory, fmt.Sprintf("%d-%d.json", record.PID, time.Now().UnixNano()))}
	return ledger, ledger.save(record)
}

func (ledger meterLedger) save(record meterRecord) error {
	record.UpdatedUTC = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return writeAtomic(ledger.path, append(data, '\n'))
}

// meterMain runs as "codex". It relays the real Codex byte for byte and saves
// its ledger before relaying each line, because Level 7 kills the app-server
// as soon as a turn completes.
func meterMain(arguments []string, stdin *os.File, stdout, stderr io.Writer) int {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(stderr, "outcome-eval meter: cannot locate itself")
		return 125
	}
	var config meterConfig
	data, err := os.ReadFile(filepath.Join(filepath.Dir(self), meterConfigName))
	if err != nil || decodeStrict(data, &config) != nil || config.Ledger == "" || (!config.Fake && !filepath.IsAbs(config.Codex)) {
		fmt.Fprintln(stderr, "outcome-eval meter: configuration is missing or invalid")
		return 125
	}
	directory, _ := os.Getwd()
	tap := newMeterTap(arguments, directory)
	ledger, err := newMeterLedger(config.Ledger, tap.record)
	if err != nil {
		fmt.Fprintln(stderr, "outcome-eval meter: cannot write its ledger")
		return 125
	}
	writer := &tappedWriter{tap: tap, ledger: ledger, output: stdout}
	exitCode := 0
	if config.Fake {
		exitCode = fakeCodex(arguments, stdin, writer, stderr)
	} else {
		exitCode = relayCodex(config.Codex, arguments, stdin, writer, stderr)
	}
	writer.flush()
	tap.record.Exited, tap.record.ExitCode = true, exitCode
	_ = ledger.save(tap.record)
	return exitCode
}

func relayCodex(codex string, arguments []string, stdin *os.File, writer *tappedWriter, stderr io.Writer) int {
	command := exec.Command(codex, arguments...)
	command.Stdin, command.Stderr = stdin, stderr
	pipe, err := command.StdoutPipe()
	if err != nil {
		fmt.Fprintln(stderr, "outcome-eval meter: cannot open the Codex output")
		return 125
	}
	if err := command.Start(); err != nil {
		fmt.Fprintln(stderr, "outcome-eval meter: cannot start Codex")
		return 127
	}
	signals := make(chan os.Signal, 4)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signals)
	go func() {
		for received := range signals {
			_ = command.Process.Signal(received)
		}
	}()
	reader := bufio.NewReaderSize(pipe, 64<<10)
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			if _, err := writer.Write(line); err != nil {
				_ = command.Process.Kill()
				break
			}
		}
		if readErr != nil {
			break
		}
	}
	err = command.Wait()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if err != nil {
		return 125
	}
	return 0
}

// tappedWriter observes complete lines, saves the ledger when they change it,
// and only then forwards them.
type tappedWriter struct {
	tap     *meterTap
	ledger  meterLedger
	output  io.Writer
	partial []byte
}

func (writer *tappedWriter) Write(data []byte) (int, error) {
	writer.partial = append(writer.partial, data...)
	for {
		index := bytes.IndexByte(writer.partial, '\n')
		if index < 0 {
			if len(writer.partial) > maxMeterLine {
				return 0, errors.New("codex output line exceeds the meter bound")
			}
			return len(data), nil
		}
		line := writer.partial[:index+1]
		if writer.tap.observe(line) {
			_ = writer.ledger.save(writer.tap.record)
		}
		if _, err := writer.output.Write(line); err != nil {
			return 0, err
		}
		writer.partial = append([]byte{}, writer.partial[index+1:]...)
	}
}

func (writer *tappedWriter) flush() {
	if len(writer.partial) > 0 {
		if writer.tap.observe(writer.partial) {
			_ = writer.ledger.save(writer.tap.record)
		}
		_, _ = writer.output.Write(writer.partial)
		writer.partial = nil
	}
}

// fakeCodex stands in for Codex in smoke runs. It answers the app-server
// handshake and turns, and codex exec, by applying the matching task's
// reference solution, so the whole pipeline runs without model calls.
func fakeCodex(arguments []string, stdin io.Reader, output io.Writer, stderr io.Writer) int {
	loaded, err := loadCorpus(embeddedCorpus)
	if err != nil {
		fmt.Fprintln(stderr, "outcome-eval fake codex:", err)
		return 125
	}
	switch {
	case len(arguments) == 1 && arguments[0] == "--version":
		fmt.Fprintln(output, "codex-cli 0.0.0-outcome-eval-fake")
		return 0
	case len(arguments) == 1 && arguments[0] == "app-server":
		return fakeAppServer(loaded, stdin, output)
	case len(arguments) > 0 && arguments[0] == "exec":
		return fakeExec(loaded, arguments[1:], stdin, output)
	}
	fmt.Fprintln(stderr, "outcome-eval fake codex: unsupported invocation")
	return 2
}

var fakeTurnUsage = map[string]int64{"inputTokens": 1000, "cachedInputTokens": 0, "outputTokens": 200, "reasoningOutputTokens": 0, "totalTokens": 1200}

func fakeAppServer(loaded corpus, stdin io.Reader, output io.Writer) int {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 64<<10), 8<<20)
	encoder := json.NewEncoder(output)
	sequence := 0
	model := func(id string) map[string]any {
		return map[string]any{"id": id, "displayName": id, "supportedReasoningEfforts": []string{"low", "medium", "high"}}
	}
	for scanner.Scan() {
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil || len(request.ID) == 0 {
			continue
		}
		reply := func(result any) { _ = encoder.Encode(map[string]any{"id": request.ID, "result": result}) }
		notify := func(method string, params any) {
			_ = encoder.Encode(map[string]any{"method": method, "params": params})
		}
		switch request.Method {
		case "initialize":
			reply(map[string]any{"userAgent": "outcome-eval-fake"})
		case "account/read":
			reply(map[string]any{"account": map[string]any{"type": "fake"}})
		case "model/list":
			reply(map[string]any{"data": []any{model("fake-implementer"), model("fake-reviewer")}})
		case "account/rateLimits/read":
			reply(map[string]any{"rateLimits": map[string]any{}})
		case "thread/start", "thread/resume":
			var params struct {
				ThreadID string `json:"threadId"`
			}
			_ = json.Unmarshal(request.Params, &params)
			if params.ThreadID == "" {
				sequence++
				params.ThreadID = fmt.Sprintf("thr-fake-%d", sequence)
			}
			reply(map[string]any{"thread": map[string]any{"id": params.ThreadID}})
		case "turn/start":
			var params struct {
				ThreadID string `json:"threadId"`
				Cwd      string `json:"cwd"`
				Input    []struct {
					Text string `json:"text"`
				} `json:"input"`
				SandboxPolicy struct {
					Type string `json:"type"`
				} `json:"sandboxPolicy"`
			}
			_ = json.Unmarshal(request.Params, &params)
			sequence++
			turnID := fmt.Sprintf("turn-fake-%d", sequence)
			prompt := ""
			for _, input := range params.Input {
				prompt += input.Text
			}
			reply(map[string]any{"turn": map[string]any{"id": turnID, "status": "inProgress"}})
			notify("turn/started", map[string]any{"threadId": params.ThreadID, "turn": map[string]any{"id": turnID}})
			text := fakeTurn(loaded, prompt, params.Cwd, params.SandboxPolicy.Type == "readOnly")
			notify("item/completed", map[string]any{"threadId": params.ThreadID, "turnId": turnID, "item": map[string]any{"type": "agentMessage", "id": "msg-" + turnID, "text": text}})
			notify("thread/tokenUsage/updated", map[string]any{"threadId": params.ThreadID, "turnId": turnID, "tokenUsage": map[string]any{"total": fakeTurnUsage, "last": fakeTurnUsage}})
			notify("turn/completed", map[string]any{"threadId": params.ThreadID, "turn": map[string]any{"id": turnID, "status": "completed"}})
		case "turn/interrupt":
			reply(map[string]any{})
		default:
			_ = encoder.Encode(map[string]any{"id": request.ID, "error": map[string]any{"code": -32601, "message": "unsupported by the outcome-eval fake"}})
		}
	}
	return 0
}

func fakeExec(loaded corpus, arguments []string, stdin io.Reader, output io.Writer) int {
	directory, lastMessage := "", ""
	for index := 0; index+1 < len(arguments); index++ {
		switch arguments[index] {
		case "-C", "--cd":
			index++
			directory = arguments[index]
		case "-o", "--output-last-message":
			index++
			lastMessage = arguments[index]
		}
	}
	prompt, _ := io.ReadAll(io.LimitReader(stdin, 1<<20))
	text := fakeTurn(loaded, string(prompt), directory, false)
	encoder := json.NewEncoder(output)
	_ = encoder.Encode(map[string]any{"type": "thread.started", "thread_id": "thr-fake-exec"})
	_ = encoder.Encode(map[string]any{"type": "turn.started"})
	_ = encoder.Encode(map[string]any{"type": "item.completed", "item": map[string]any{"id": "item_0", "type": "agent_message", "text": text}})
	_ = encoder.Encode(map[string]any{"type": "turn.completed", "usage": map[string]any{"input_tokens": 1000, "cached_input_tokens": 0, "output_tokens": 200}})
	if lastMessage != "" {
		if err := os.WriteFile(lastMessage, []byte(text), 0o644); err != nil {
			return 1
		}
	}
	return 0
}

func fakeTurn(loaded corpus, prompt, directory string, reviewer bool) string {
	if reviewer {
		return `{"outcome":"complete","summary":"Outcome-eval fake review: the change matches the task.","decision":"GO"}`
	}
	matches := []task{}
	for _, candidate := range loaded.tasks {
		if strings.Contains(prompt, candidate.Title) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) != 1 || !filepath.IsAbs(directory) {
		return `{"outcome":"blocked","summary":"Outcome-eval fake: no single corpus task matches this prompt."}`
	}
	if err := writeTree(directory, matches[0].referenceFiles(loaded.protocol.Canary)); err != nil {
		return `{"outcome":"blocked","summary":"Outcome-eval fake: cannot write the reference solution."}`
	}
	return `{"outcome":"complete","summary":"Outcome-eval fake: applied the reference solution."}`
}

type meterSummary struct {
	Measured     bool  `json:"measured"`
	Processes    int   `json:"processes"`
	Turns        int   `json:"turns"`
	Usage        usage `json:"usage"`
	Unattributed int   `json:"unattributed_usage_events"`
	Malformed    int   `json:"malformed_lines"`
}

// summarizeLedger totals the metered app-server and exec processes that ran
// inside directory. Usage counts as measured only when every process that
// ran a turn reported usage for its turns.
func summarizeLedger(ledgerDirectory, directory string) (meterSummary, error) {
	entries, err := os.ReadDir(ledgerDirectory)
	if err != nil {
		return meterSummary{}, err
	}
	summary, incomplete := meterSummary{}, false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ledgerDirectory, entry.Name()))
		if err != nil {
			return meterSummary{}, err
		}
		var record meterRecord
		if err := json.Unmarshal(data, &record); err != nil || record.Schema != 1 {
			return meterSummary{}, fmt.Errorf("meter record %s is malformed", entry.Name())
		}
		if record.Mode == "other" || !within(record.Directory, directory) {
			continue
		}
		summary.Processes++
		summary.Turns += record.Turns
		summary.Usage.add(record.Usage)
		summary.Unattributed += record.Unattributed
		summary.Malformed += record.Malformed
		incomplete = incomplete || record.UsageEvents < record.Turns
	}
	summary.Measured = summary.Turns > 0 && !incomplete && summary.Malformed == 0
	return summary, nil
}
