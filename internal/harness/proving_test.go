package harness

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"runtime"
	"testing"
)

func TestHarnessContract(t *testing.T) {
	t.Parallel()

	if expectedGoVersion == "unset" || expectedGOOS == "unset" || expectedGOARCH == "unset" {
		t.Fatal("harness toolchain identity was not bound at link time")
	}
	if version := runtime.Version(); version != expectedGoVersion {
		t.Fatalf("Go version: got %q, want %q", version, expectedGoVersion)
	}
	if runtime.GOOS != expectedGOOS {
		t.Fatalf("GOOS: got %q, want %q", runtime.GOOS, expectedGOOS)
	}
	if runtime.GOARCH != expectedGOARCH {
		t.Fatalf("GOARCH: got %q, want %q", runtime.GOARCH, expectedGOARCH)
	}

	var output bytes.Buffer
	handler := slog.NewJSONHandler(&output, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	})
	logger := slog.New(handler)
	logger.Info(
		"harness proof",
		"schema", "l7.harness.proof.v1",
		"status", "PASS",
		"effect", "none",
		"telemetry", false,
	)

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode structured log: %v", err)
	}

	want := map[string]any{
		"level":     "INFO",
		"msg":       "harness proof",
		"schema":    "l7.harness.proof.v1",
		"status":    "PASS",
		"effect":    "none",
		"telemetry": false,
	}
	if len(event) != len(want) {
		t.Fatalf("structured log keys: got %v, want exactly %v", event, want)
	}
	for key, expected := range want {
		if got := event[key]; got != expected {
			t.Fatalf("structured log %s: got %v, want %v", key, got, expected)
		}
	}
}
