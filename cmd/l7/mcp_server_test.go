package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
)

func TestMCPNegotiatesVersionAndListsVersionedTools(t *testing.T) {
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"missing","params":{}}`,
	}, "\n") + "\n"
	var output, diagnostic bytes.Buffer
	if code := serveMCP(context.Background(), t.TempDir(), strings.NewReader(input), &output, &diagnostic); code != 0 || diagnostic.Len() != 0 {
		t.Fatalf("code=%d output=%s diagnostic=%s", code, output.String(), diagnostic.String())
	}
	responses := decodeMCPResponses(t, output.Bytes())
	if len(responses) != 3 {
		t.Fatalf("responses=%#v", responses)
	}
	initialize, _ := responses[0].Result.(map[string]any)
	if initialize["protocolVersion"] != "2025-11-25" {
		t.Fatalf("initialize=%#v", initialize)
	}
	listed, _ := responses[1].Result.(map[string]any)
	tools, _ := listed["tools"].([]any)
	if len(tools) != 6 {
		t.Fatalf("tools=%#v", tools)
	}
	for _, raw := range tools {
		tool, _ := raw.(map[string]any)
		if !strings.HasPrefix(tool["name"].(string), "l7_v1_") {
			t.Fatalf("unversioned tool: %#v", tool)
		}
	}
	if responses[2].Error == nil || responses[2].Error.Code != -32601 {
		t.Fatalf("unknown method response=%#v", responses[2])
	}
}

func TestMCPToolCallReturnsStructuredAndTextEnvelope(t *testing.T) {
	repository := cliRepository(t)
	configuration := orchestrationconfig.AppliedDefault()
	configuration.Providers = []orchestrationconfig.Provider{}
	data, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repository, ".l7"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".l7", "orchestration.json"), append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	input := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"l7_v1_onboard","arguments":{"action":"status"}}}` + "\n"
	var output bytes.Buffer
	if code := serveMCP(context.Background(), repository, strings.NewReader(input), &output, &bytes.Buffer{}); code != 0 {
		t.Fatalf("code=%d output=%s", code, output.String())
	}
	responses := decodeMCPResponses(t, output.Bytes())
	result, _ := responses[0].Result.(map[string]any)
	content, _ := result["content"].([]any)
	structured, _ := result["structuredContent"].(map[string]any)
	if len(content) != 1 || structured["command"] != "onboard" || result["isError"] != false {
		t.Fatalf("tool result=%#v", result)
	}
	text, _ := content[0].(map[string]any)
	if !strings.Contains(text["text"].(string), `"next"`) {
		t.Fatalf("text fallback=%#v", text)
	}
}

func TestMCPRejectsUnknownToolArguments(t *testing.T) {
	_, _, err := mcpToolArguments("l7_v1_onboard", map[string]any{"action": "status", "secret": "value"})
	if err == nil || !strings.Contains(err.Error(), "unknown tool argument") {
		t.Fatalf("error=%v", err)
	}
}

func decodeMCPResponses(t *testing.T, data []byte) []mcpResponse {
	t.Helper()
	responses := []mcpResponse{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var response mcpResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		responses = append(responses, response)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return responses
}
