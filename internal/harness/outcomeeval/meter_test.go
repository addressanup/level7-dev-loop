package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func observeAll(tap *meterTap, lines ...string) {
	for _, line := range lines {
		tap.observe([]byte(line + "\n"))
	}
}

func TestMeterTapCountsOnlyTurnsStartedInItsProcess(t *testing.T) {
	tap := newMeterTap([]string{"app-server"}, "/w")
	observeAll(tap,
		`{"id":2,"result":{"thread":{"id":"thr"}}}`,
		`{"method":"thread/tokenUsage/updated","params":{"threadId":"thr","turnId":"resumed","tokenUsage":{"total":{"totalTokens":900},"last":{"inputTokens":800,"cachedInputTokens":0,"outputTokens":100,"reasoningOutputTokens":0,"totalTokens":900}}}}`,
		`{"id":3,"result":{"turn":{"id":"t1","status":"inProgress"}}}`,
		`{"method":"turn/started","params":{"threadId":"thr","turn":{"id":"t1"}}}`,
		`{"method":"thread/tokenUsage/updated","params":{"threadId":"thr","turnId":"t1","tokenUsage":{"total":{"totalTokens":1020},"last":{"inputTokens":100,"cachedInputTokens":40,"outputTokens":20,"reasoningOutputTokens":5,"totalTokens":120}}}}`,
		`{"method":"thread/tokenUsage/updated","params":{"threadId":"thr","turnId":"t1","tokenUsage":{"total":{"totalTokens":1080},"last":{"inputTokens":50,"cachedInputTokens":10,"outputTokens":10,"reasoningOutputTokens":0,"totalTokens":60}}}}`,
		`{"method":"item/completed","params":{"item":{"type":"agentMessage","text":"{}"}}}`,
		`{"method":"turn/completed","params":{"threadId":"thr","turn":{"id":"t1","status":"completed"}}}`,
		`{"method":"turn/completed","params":{"threadId":"thr","turn":{"id":"elsewhere","status":"completed"}}}`,
		`plain text from a wrapper`,
		`{"method": broken`,
	)
	record := tap.record
	want := usage{Input: 150, CachedInput: 50, Output: 30, ReasoningOutput: 5, Total: 180}
	if record.Turns != 1 || record.Usage != want || record.UsageEvents != 2 || record.Unattributed != 1 || record.Malformed != 1 {
		t.Fatalf("record = %+v", record)
	}
}

func TestMeterTapReadsExecUsage(t *testing.T) {
	tap := newMeterTap([]string{"exec", "--json"}, "/w")
	observeAll(tap,
		`{"type":"thread.started","thread_id":"thr"}`,
		`{"type":"turn.started"}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"{}"}}`,
		`{"type":"turn.completed","usage":{"input_tokens":300,"cached_input_tokens":100,"cache_write_input_tokens":0,"output_tokens":40,"reasoning_output_tokens":10}}`,
	)
	want := usage{Input: 300, CachedInput: 100, Output: 40, ReasoningOutput: 10, Total: 340}
	if tap.record.Mode != "exec" || tap.record.Turns != 1 || tap.record.Usage != want || tap.record.UsageEvents != 1 {
		t.Fatalf("record = %+v", tap.record)
	}
	other := newMeterTap([]string{"--version"}, "/")
	observeAll(other, `{"type":"turn.completed","usage":{"input_tokens":1}}`)
	if other.record.Mode != "other" || other.record.Turns != 0 {
		t.Fatalf("other = %+v", other.record)
	}
}

// installMeter copies the test binary as "codex"; TestMain then runs it as
// the meter.
func installMeter(t *testing.T, config meterConfig) (string, string) {
	t.Helper()
	directory, err := physicalDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bin, ledger := filepath.Join(directory, "bin"), filepath.Join(directory, "ledger")
	for _, name := range []string{bin, ledger} {
		if err := os.MkdirAll(name, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := copyExecutable(self, filepath.Join(bin, meterName)); err != nil {
		t.Fatal(err)
	}
	config.Ledger = ledger
	encoded, _ := json.Marshal(config)
	if err := os.WriteFile(filepath.Join(bin, meterConfigName), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(bin, meterName), ledger
}

func writeScript(t *testing.T, body string) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "real-codex")
	if err := os.WriteFile(name, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return name
}

func ledgerRecords(t *testing.T, ledger string) []meterRecord {
	t.Helper()
	entries, err := os.ReadDir(ledger)
	if err != nil {
		t.Fatal(err)
	}
	records := []meterRecord{}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(ledger, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var record meterRecord
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		records = append(records, record)
	}
	return records
}

func TestMeterRelaysCodexBytesAndExitStatus(t *testing.T) {
	script := writeScript(t, `printf '%s\n' "$*" >&2
printf '{"type":"turn.completed","usage":{"input_tokens":5,"cached_input_tokens":1,"output_tokens":2}}\n'
printf 'partial'
cat
exit 3
`)
	meter, ledger := installMeter(t, meterConfig{Codex: script})
	directory, _ := physicalDirectory(t.TempDir())
	command := exec.Command(meter, "exec", "--json", "-")
	command.Dir, command.Stdin = directory, strings.NewReader("prompt-bytes")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Fatalf("err = %v, stderr = %q", err, stderr.String())
	}
	if want := `{"type":"turn.completed","usage":{"input_tokens":5,"cached_input_tokens":1,"output_tokens":2}}` + "\npartialprompt-bytes"; stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if strings.TrimSpace(stderr.String()) != "exec --json -" {
		t.Fatalf("arguments reached Codex as %q", stderr.String())
	}
	records := ledgerRecords(t, ledger)
	if len(records) != 1 {
		t.Fatalf("records = %+v", records)
	}
	record := records[0]
	if record.Mode != "exec" || record.Directory != directory || record.Turns != 1 || record.Usage != (usage{Input: 5, CachedInput: 1, Output: 2, Total: 7}) || !record.Exited || record.ExitCode != 3 {
		t.Fatalf("record = %+v", record)
	}
}

// TestMeterSavesTheLedgerBeforeRelayingALine matters because Level 7 kills
// the app-server as soon as it reads turn/completed.
func TestMeterSavesTheLedgerBeforeRelayingALine(t *testing.T) {
	ledger := t.TempDir()
	tap := newMeterTap([]string{"app-server"}, "/w")
	store, err := newMeterLedger(ledger, tap.record)
	if err != nil {
		t.Fatal(err)
	}
	seen := []int{}
	output := writerFunc(func(line []byte) {
		if bytes.Contains(line, []byte("turn/completed")) {
			data, _ := os.ReadFile(store.path)
			var record meterRecord
			_ = json.Unmarshal(data, &record)
			seen = append(seen, record.Turns, record.UsageEvents)
		}
	})
	writer := &tappedWriter{tap: tap, ledger: store, output: output}
	lines := `{"id":3,"result":{"turn":{"id":"t1"}}}` + "\n" +
		`{"method":"thread/tokenUsage/updated","params":{"turnId":"t1","tokenUsage":{"last":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}}}` + "\n" +
		`{"method":"turn/completed","params":{"turn":{"id":"t1","status":"completed"}}}` + "\n"
	for _, chunk := range []string{lines[:17], lines[17:90], lines[90:]} {
		if _, err := writer.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 1 {
		t.Fatalf("ledger when turn/completed was relayed: turns, usage events = %v", seen)
	}
}

type writerFunc func([]byte)

func (write writerFunc) Write(data []byte) (int, error) {
	write(data)
	return len(data), nil
}

func TestSummarizeLedgerAttributesProcessesByDirectory(t *testing.T) {
	ledger := t.TempDir()
	trial := "/w/trials/slugify/r1-crew"
	write := func(name string, record meterRecord) {
		record.Schema = 1
		data, _ := json.Marshal(record)
		if err := os.WriteFile(filepath.Join(ledger, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("1.json", meterRecord{Mode: "app-server", Directory: trial + "/repo/.git/l7/crew/worktrees/t01", Turns: 1, UsageEvents: 2, Usage: usage{Input: 10, Total: 12}})
	write("2.json", meterRecord{Mode: "app-server", Directory: trial + "/repo/.git/l7/crew/worktrees/t01", Turns: 1, UsageEvents: 1, Usage: usage{Input: 5, Total: 6}})
	write("3.json", meterRecord{Mode: "app-server", Directory: "/", Turns: 0})
	write("4.json", meterRecord{Mode: "other", Directory: trial, Turns: 9})
	write("5.json", meterRecord{Mode: "exec", Directory: "/w/trials/slugify/r1-crew-other", Turns: 1, UsageEvents: 1})
	summary, err := summarizeLedger(ledger, trial)
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Measured || summary.Processes != 2 || summary.Turns != 2 || summary.Usage != (usage{Input: 15, Total: 18}) {
		t.Fatalf("summary = %+v", summary)
	}
	write("6.json", meterRecord{Mode: "app-server", Directory: trial, Turns: 1})
	if summary, _ := summarizeLedger(ledger, trial); summary.Measured {
		t.Fatalf("a turn without usage was reported as measured: %+v", summary)
	}
	if summary, _ := summarizeLedger(ledger, "/w/trials/none"); summary.Measured || summary.Processes != 0 {
		t.Fatalf("a trial without metered turns was reported as measured: %+v", summary)
	}
}

func TestFakeAppServerImplementsAndReviews(t *testing.T) {
	loaded, value := corpusTask(t, "duration-parse")
	directory := t.TempDir()
	turn := func(id int, sandbox string) string {
		encoded, _ := json.Marshal(map[string]any{"id": id, "method": "turn/start", "params": map[string]any{
			"threadId": "thr-1", "cwd": directory, "sandboxPolicy": map[string]any{"type": sandbox},
			"input": []any{map[string]any{"type": "text", "text": "Task t01: " + value.Title + "\n\nbrief"}},
		}})
		return string(encoded)
	}
	input := strings.Join([]string{
		`{"id":1,"method":"initialize","params":{}}`, `{"method":"initialized","params":{}}`,
		`{"id":2,"method":"account/read","params":{}}`, `{"id":3,"method":"model/list","params":{}}`,
		`{"id":4,"method":"thread/start","params":{}}`, turn(5, "workspaceWrite"), turn(6, "readOnly"),
		`{"id":7,"method":"unknown/method","params":{}}`,
	}, "\n") + "\n"
	var output bytes.Buffer
	if code := fakeAppServer(loaded, strings.NewReader(input), &output); code != 0 {
		t.Fatalf("exit %d", code)
	}
	messages, models, unsupported := []string{}, 0, false
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var line struct {
			ID     int `json:"id"`
			Method string
			Result struct {
				Account any   `json:"account"`
				Data    []any `json:"data"`
			}
			Error  any `json:"error"`
			Params struct {
				Item struct {
					Text string `json:"text"`
				} `json:"item"`
			}
		}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("%q: %v", scanner.Text(), err)
		}
		if line.ID == 3 {
			models = len(line.Result.Data)
		}
		if line.ID == 7 && line.Error != nil {
			unsupported = true
		}
		if line.Method == "item/completed" {
			messages = append(messages, line.Params.Item.Text)
		}
	}
	if models != 2 || !unsupported || len(messages) != 2 || !strings.Contains(messages[0], `"outcome":"complete"`) || !strings.Contains(messages[1], `"decision":"GO"`) {
		t.Fatalf("models=%d unsupported=%t messages=%q", models, unsupported, messages)
	}
	written, err := os.ReadFile(filepath.Join(directory, "timeparse", "duration.py"))
	if err != nil || !strings.Contains(string(written), "_PATTERN") || strings.Contains(string(written), loaded.protocol.Canary) {
		t.Fatalf("reference not applied without its canary: %v %q", err, written)
	}
}

func TestFakeExecAppliesTheMatchingReferenceOnly(t *testing.T) {
	loaded, value := corpusTask(t, "slugify")
	directory, answer := t.TempDir(), filepath.Join(t.TempDir(), "answer.json")
	var output bytes.Buffer
	if code := fakeExec(loaded, []string{"--json", "-C", directory, "-o", answer, "-"}, strings.NewReader(plainPrompt(value)), &output); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if claim := readClaim(answer, output.Bytes()); claim != claimComplete {
		t.Fatalf("claim = %s", claim)
	}
	if _, err := os.Stat(filepath.Join(directory, "textkit", "slug.py")); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	other := t.TempDir()
	fakeExec(loaded, []string{"-C", other, "-o", answer, "-"}, strings.NewReader("Task: something else"), &output)
	if claim := readClaim(answer, nil); claim != claimBlocked {
		t.Fatalf("an unmatched prompt was claimed as %s", claim)
	}
}
