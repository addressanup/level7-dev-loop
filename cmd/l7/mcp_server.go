package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const mcpProtocolVersion = "2025-11-25"

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func serveMCP(ctx context.Context, cwd string, input io.Reader, output, diagnostic io.Writer) int {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return 130
		}
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var request mcpRequest
		if err := json.Unmarshal(line, &request); err != nil || request.JSONRPC != "2.0" || request.Method == "" {
			_ = encoder.Encode(mcpResponse{JSONRPC: "2.0", Error: &mcpError{Code: -32700, Message: "invalid JSON-RPC request"}})
			continue
		}
		if len(request.ID) == 0 || string(request.ID) == "null" {
			continue
		}
		response := handleMCP(ctx, cwd, request)
		if err := encoder.Encode(response); err != nil {
			fmt.Fprintln(diagnostic, "L7 MCP output failed")
			return 1
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(diagnostic, "L7 MCP input exceeded framing bounds")
		return 1
	}
	return 0
}

func handleMCP(ctx context.Context, cwd string, request mcpRequest) mcpResponse {
	response := mcpResponse{JSONRPC: "2.0", ID: request.ID}
	switch request.Method {
	case "initialize":
		protocol := mcpProtocolVersion
		var initialize struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(request.Params, &initialize) == nil && supportedMCPVersion(initialize.ProtocolVersion) {
			protocol = initialize.ProtocolVersion
		}
		response.Result = map[string]any{
			"protocolVersion": protocol,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "level7", "version": version},
			"instructions":    "Level 7 v1 tools are local-first, policy-bound, and stop before deployment.",
		}
	case "ping":
		response.Result = map[string]any{}
	case "tools/list":
		response.Result = map[string]any{"tools": mcpTools()}
	case "tools/call":
		var call struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if json.Unmarshal(request.Params, &call) != nil || call.Name == "" {
			response.Error = &mcpError{Code: -32602, Message: "invalid tool call"}
			return response
		}
		command, arguments, err := mcpToolArguments(call.Name, call.Arguments)
		if err != nil {
			response.Error = &mcpError{Code: -32602, Message: err.Error()}
			return response
		}
		envelope, executeErr := executeOrchestration(ctx, command, arguments, cwd)
		if executeErr != nil {
			envelope = failedEnvelope(command, "L7-MCP-002", "failed", executeErr.Error(), nextForFailure(command))
		}
		text, _ := json.Marshal(envelope)
		response.Result = map[string]any{
			"content":           []any{map[string]any{"type": "text", "text": string(text)}},
			"structuredContent": envelope,
			"isError":           envelope.Outcome != "PASS",
		}
	default:
		response.Error = &mcpError{Code: -32601, Message: "method not found"}
	}
	return response
}

func supportedMCPVersion(value string) bool {
	switch value {
	case "2024-11-05", "2025-03-26", "2025-06-18", "2025-11-25":
		return true
	default:
		return false
	}
}

func mcpTools() []mcpTool {
	object := func(properties map[string]any, required ...string) map[string]any {
		value := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
		if len(required) > 0 {
			value["required"] = required
		}
		return value
	}
	enum := func(values ...string) map[string]any { return map[string]any{"type": "string", "enum": values} }
	text := func() map[string]any { return map[string]any{"type": "string", "minLength": 1, "maxLength": 4096} }
	return []mcpTool{
		{Name: "l7_v1_onboard", Description: "Inspect project state or explicitly apply the default-off orchestration policy.", InputSchema: object(map[string]any{"action": enum("status", "apply")}, "action")},
		{Name: "l7_v1_provider_discovery", Description: "List local capabilities without gateway traffic or explicitly probe configured providers.", InputSchema: object(map[string]any{"action": enum("list", "probe")}, "action")},
		{Name: "l7_v1_route_explain", Description: "Persist and explain capability filtering and balanced model/effort routing.", InputSchema: object(map[string]any{"task": text(), "complexity": enum("C1", "C2", "C3", "C4"), "risk": map[string]any{"type": "integer", "minimum": 1, "maximum": 3}, "context": map[string]any{"type": "integer", "minimum": 0}, "tools": map[string]any{"type": "boolean"}, "edit": map[string]any{"type": "boolean"}, "resume": map[string]any{"type": "boolean"}, "review": map[string]any{"type": "boolean"}, "language": text(), "work_kind": enum("security", "architecture", "release", "implementation", "review"), "prior_failures": map[string]any{"type": "integer", "minimum": 0}, "implementer_provider": text(), "implementer_model": text()}, "task")},
		{Name: "l7_v1_memory", Description: "Incrementally synchronize, rebuild, or query private Git-bound codebase memory.", InputSchema: object(map[string]any{"action": enum("incremental", "rebuild", "query"), "query": text()}, "action")},
		{Name: "l7_v1_cyber", Description: "Run read-only Cyber audit, explicitly isolated active confirmation, or produce a separate remediation brief.", InputSchema: object(map[string]any{"action": enum("audit", "remediate"), "active": map[string]any{"type": "boolean"}, "export": enum("markdown", "json"), "report": text()}, "action")},
		{Name: "l7_v1_headless", Description: "Plan and operate a durable, approved Headless feature-wave lifecycle that stops before deployment.", InputSchema: object(map[string]any{"action": enum("plan", "start", "status", "resume", "cancel"), "objective": text(), "target": text(), "allow_paths": map[string]any{"type": "array", "items": text(), "maxItems": 256}, "commands": map[string]any{"type": "array", "items": map[string]any{"type": "array", "items": text(), "maxItems": 64}, "maxItems": 64}, "local_merge": map[string]any{"type": "boolean"}, "run": text(), "digest": text(), "owner": text(), "role": text(), "confirm": map[string]any{"type": "boolean"}}, "action")},
	}
}

func mcpToolArguments(name string, values map[string]any) (string, []string, error) {
	allowed := func(keys ...string) error {
		set := map[string]bool{}
		for _, key := range keys {
			set[key] = true
		}
		for key := range values {
			if !set[key] {
				return fmt.Errorf("unknown tool argument %s", key)
			}
		}
		return nil
	}
	stringValue := func(key string) string {
		value, _ := values[key].(string)
		return value
	}
	boolValue := func(key string) bool {
		value, _ := values[key].(bool)
		return value
	}
	switch name {
	case "l7_v1_onboard":
		if err := allowed("action"); err != nil {
			return "", nil, err
		}
		action := stringValue("action")
		if action != "status" && action != "apply" {
			return "", nil, errors.New("onboard action must be status or apply")
		}
		return "onboard", []string{"--" + action}, nil
	case "l7_v1_provider_discovery":
		if err := allowed("action"); err != nil {
			return "", nil, err
		}
		action := stringValue("action")
		if action != "list" && action != "probe" {
			return "", nil, errors.New("provider action must be list or probe")
		}
		return "providers", []string{action}, nil
	case "l7_v1_route_explain":
		if err := allowed("task", "complexity", "risk", "context", "tools", "edit", "resume", "review", "language", "work_kind", "prior_failures", "implementer_provider", "implementer_model"); err != nil {
			return "", nil, err
		}
		arguments := []string{"explain", "--task", stringValue("task")}
		for _, pair := range [][2]string{{"complexity", "--complexity"}, {"language", "--language"}, {"work_kind", "--work-kind"}, {"implementer_provider", "--implementer-provider"}, {"implementer_model", "--implementer-model"}} {
			if value := stringValue(pair[0]); value != "" {
				arguments = append(arguments, pair[1], value)
			}
		}
		for _, pair := range [][2]string{{"tools", "--tools"}, {"edit", "--edit"}, {"resume", "--resume"}, {"review", "--review"}} {
			if boolValue(pair[0]) {
				arguments = append(arguments, pair[1])
			}
		}
		for _, pair := range [][2]string{{"risk", "--risk"}, {"context", "--context"}, {"prior_failures", "--prior-failures"}} {
			if number, ok := values[pair[0]].(float64); ok {
				arguments = append(arguments, pair[1], fmt.Sprintf("%.0f", number))
			}
		}
		return "route", arguments, nil
	case "l7_v1_memory":
		if err := allowed("action", "query"); err != nil {
			return "", nil, err
		}
		action := stringValue("action")
		switch action {
		case "incremental", "rebuild":
			return "sync", []string{"--" + action}, nil
		case "query":
			if stringValue("query") == "" {
				return "", nil, errors.New("memory query is required")
			}
			return "sync", []string{"--query", stringValue("query")}, nil
		default:
			return "", nil, errors.New("memory action is invalid")
		}
	case "l7_v1_cyber":
		if err := allowed("action", "active", "export", "report"); err != nil {
			return "", nil, err
		}
		if stringValue("action") == "remediate" {
			if stringValue("report") == "" {
				return "", nil, errors.New("Cyber report is required")
			}
			return "cyber", []string{"remediate", "--report", stringValue("report")}, nil
		}
		if stringValue("action") != "audit" {
			return "", nil, errors.New("Cyber action is invalid")
		}
		arguments := []string{}
		if boolValue("active") {
			arguments = append(arguments, "--active")
		}
		if value := stringValue("export"); value != "" {
			arguments = append(arguments, "--export", value)
		}
		return "cyber", arguments, nil
	case "l7_v1_headless":
		if err := allowed("action", "objective", "target", "allow_paths", "commands", "local_merge", "run", "digest", "owner", "role", "confirm"); err != nil {
			return "", nil, err
		}
		action := stringValue("action")
		if action == "status" || action == "resume" || action == "cancel" {
			return "headless", []string{action}, nil
		}
		if action == "plan" {
			arguments := []string{"plan", "--objective", stringValue("objective"), "--target", stringValue("target")}
			if boolValue("local_merge") {
				arguments = append(arguments, "--local-merge")
			}
			if paths, ok := values["allow_paths"].([]any); ok {
				for _, raw := range paths {
					value, _ := raw.(string)
					arguments = append(arguments, "--allow-path", value)
				}
			}
			if commands, ok := values["commands"].([]any); ok {
				for _, raw := range commands {
					encoded, _ := json.Marshal(raw)
					arguments = append(arguments, "--command-json", string(encoded))
				}
			}
			return "headless", arguments, nil
		}
		if action == "start" {
			arguments := []string{"start", "--run", stringValue("run"), "--digest", stringValue("digest"), "--owner", stringValue("owner"), "--role", stringValue("role")}
			if boolValue("confirm") {
				arguments = append(arguments, "--confirm")
			}
			return "headless", arguments, nil
		}
		return "", nil, errors.New("Headless action is invalid")
	default:
		return "", nil, errors.New("unknown Level 7 v1 tool")
	}
}
