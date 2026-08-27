package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunCommandContract(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		exit     int
		contains []string
	}{
		{"default help", nil, 0, []string{"PASS", `command="help"`, "Usage: l7"}},
		{"help flag", []string{"--help"}, 0, []string{"PASS", `command="help"`}},
		{"version flag", []string{"--version"}, 0, []string{"PASS", `command="version"`, `version="test-version"`}},
		{"status unavailable", []string{"status"}, 2, []string{"BLOCKED", "L7-STATUS-001", `state="unavailable"`}},
		{"unknown command", []string{"run"}, 1, []string{"FAILED", "L7-CLI-001", `command="run"`}},
		{"unknown flag", []string{"status", "--unsafe"}, 1, []string{"FAILED", "unknown flag"}},
		{"extra command", []string{"help", "status"}, 1, []string{"FAILED", "expected exactly one command"}},
		{"duplicate json", []string{"--json", "status", "--json"}, 1, []string{`"outcome":"FAILED"`, `"message":"duplicate --json flag"`}},
		{"json status", []string{"status", "--json"}, 2, []string{`{"schema":1`, `"outcome":"BLOCKED"`, `"details":[]`}},
	}
	previous := version
	version = "test-version"
	t.Cleanup(func() { version = previous })
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run(context.Background(), test.args, &stdout, &stderr)
			if exit != test.exit {
				t.Fatalf("exit=%d, want %d; stdout=%q stderr=%q", exit, test.exit, stdout.String(), stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", stderr.String())
			}
			for _, substring := range test.contains {
				if !strings.Contains(stdout.String(), substring) {
					t.Fatalf("stdout=%q, want substring %q", stdout.String(), substring)
				}
			}
		})
	}
}

func TestRunHonorsCancellationBeforeExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	if exit := run(ctx, []string{"status"}, &stdout, &stderr); exit != 130 || !strings.HasPrefix(stdout.String(), "CANCELLED ") || stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestRunOutputDoesNotDependOnWriterType(t *testing.T) {
	var buffer bytes.Buffer
	var builder strings.Builder
	if exit := run(context.Background(), []string{"version"}, &buffer, ioDiscard{}); exit != 0 {
		t.Fatalf("buffer exit=%d", exit)
	}
	if exit := run(context.Background(), []string{"version"}, &builder, ioDiscard{}); exit != 0 {
		t.Fatalf("builder exit=%d", exit)
	}
	if buffer.String() != builder.String() {
		t.Fatalf("writer-dependent output: buffer=%q builder=%q", buffer.String(), builder.String())
	}
}

func TestRunReportsOutputFailureOnStderr(t *testing.T) {
	var stderr bytes.Buffer
	if exit := run(context.Background(), []string{"version"}, errorWriter{}, &stderr); exit != 1 || !strings.Contains(stderr.String(), "L7-CLI-002") {
		t.Fatalf("exit=%d stderr=%q", exit, stderr.String())
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

type ioDiscard struct{}

func (ioDiscard) Write(data []byte) (int, error) { return len(data), nil }
