package codexapp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "app-server" {
		fakeAppServer(os.Stdin, os.Stdout)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeAppServer speaks the subset of the app-server protocol Run uses. A
// prompt containing HANG keeps the turn open until turn/interrupt arrives.
func fakeAppServer(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		var request struct {
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			return
		}
		switch request.Method {
		case "initialize":
			_ = encoder.Encode(map[string]any{"id": 1, "result": map[string]any{}})
		case "thread/start", "thread/resume":
			_ = encoder.Encode(map[string]any{"id": 2, "result": map[string]any{"thread": map[string]any{"id": "thread-fake"}}})
		case "turn/start":
			_ = os.WriteFile("turn-started.marker", nil, 0o600)
			_ = encoder.Encode(map[string]any{"id": 3, "result": map[string]any{"turn": map[string]any{"id": "turn-fake"}}})
			input, _ := request.Params["input"].([]any)
			first, _ := input[0].(map[string]any)
			if text, _ := first["text"].(string); strings.Contains(text, "HANG") {
				continue
			}
			_ = encoder.Encode(map[string]any{"method": "item/completed", "params": map[string]any{"item": map[string]any{"type": "agentMessage", "text": `{"outcome":"complete","summary":"done"}`}}})
			_ = encoder.Encode(map[string]any{"method": "turn/completed", "params": map[string]any{"turn": map[string]any{"status": "completed"}}})
		case "turn/interrupt":
			turnID, _ := request.Params["turnId"].(string)
			_ = os.WriteFile("interrupted.marker", []byte(turnID), 0o600)
			_ = encoder.Encode(map[string]any{"id": 5, "result": map[string]any{}})
			_ = encoder.Encode(map[string]any{"method": "turn/completed", "params": map[string]any{"turn": map[string]any{"status": "interrupted"}}})
		}
	}
}

func fakeAssignment(t *testing.T, prompt string) Assignment {
	t.Helper()
	executable, err := os.Executable()
	if err == nil {
		executable, err = filepath.EvalSymlinks(executable)
	}
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return Assignment{Executable: executable, Root: root, Model: "fake-model", Effort: domain.EffortHigh, Prompt: prompt}
}

func TestRunCompletesATurnOverTheAppServerProtocol(t *testing.T) {
	result, err := Run(context.Background(), fakeAssignment(t, "fix the handler"))
	if err != nil || result.SessionID != "thread-fake" || result.TurnID != "turn-fake" || result.Status != "completed" || result.Summary != "done" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestRunInterruptsTheServerTurnWhenCancelled(t *testing.T) {
	assignment := fakeAssignment(t, "HANG until interrupted")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type outcome struct {
		result Result
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := Run(ctx, assignment)
		done <- outcome{result, err}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(assignment.Root, "turn-started.marker")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fake app-server never started the turn")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	var got outcome
	select {
	case got = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run ignored cancellation")
	}
	if !errors.Is(got.err, context.Canceled) || got.result.SessionID != "thread-fake" || got.result.Status != "interrupted" {
		t.Fatalf("cancelled run = %+v err=%v", got.result, got.err)
	}
	marker, err := os.ReadFile(filepath.Join(assignment.Root, "interrupted.marker"))
	if err != nil || string(marker) != "turn-fake" {
		t.Fatalf("the server turn was not interrupted: %q err=%v", marker, err)
	}
}
