package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type Assignment struct {
	Provider orchestrationconfig.Provider
	Model    string
	Effort   domain.ReasoningEffort
	Prompt   string
	Reviewer bool
}

type Result struct {
	Summary  string
	Findings []string
	Decision domain.ReviewDecision
}

func (client Client) Run(ctx context.Context, assignment Assignment, broker Broker) (Result, error) {
	if client.http == nil || broker == nil || assignment.Prompt == "" || len(assignment.Prompt) > 1<<20 || !assignment.Effort.Valid() {
		return Result{}, errors.New("gateway assignment is invalid")
	}
	configured := false
	for _, model := range assignment.Provider.Models {
		if model.ID == assignment.Model {
			configured = true
			break
		}
	}
	if !configured {
		return Result{}, errors.New("gateway model is not configured")
	}
	secret, err := client.credential(ctx, assignment.Provider.Credential)
	if err != nil {
		return Result{}, err
	}
	if assignment.Provider.Kind == domain.ProviderKindOpenAIResponses {
		return client.runOpenAI(ctx, assignment, broker, secret)
	}
	if assignment.Provider.Kind == domain.ProviderKindAnthropic {
		return client.runAnthropic(ctx, assignment, broker, secret)
	}
	return Result{}, errors.New("gateway protocol is unsupported")
}

func (client Client) runOpenAI(ctx context.Context, assignment Assignment, broker Broker, secret string) (Result, error) {
	input := any(assignment.Prompt)
	previous := ""
	for turn := 0; turn < maxTurns; turn++ {
		body := map[string]any{"model": assignment.Model, "input": input, "tools": openAITools(), "tool_choice": "auto", "store": false, "reasoning": map[string]any{"effort": assignment.Effort}}
		if previous != "" {
			body["previous_response_id"] = previous
		}
		data, _, err := client.request(ctx, assignment.Provider.Endpoint, map[string]string{"content-type": "application/json", "authorization": "Bearer " + secret}, body)
		if err != nil {
			return Result{}, err
		}
		var response map[string]any
		if json.Unmarshal(data, &response) != nil {
			return Result{}, errors.New("OpenAI response is malformed")
		}
		previous, _ = response["id"].(string)
		calls, err := openAICalls(response)
		if err != nil {
			return Result{}, err
		}
		if len(calls) == 0 {
			return Result{}, errors.New("OpenAI worker stopped without finish")
		}
		outputs := make([]any, 0, len(calls))
		for _, call := range calls {
			if call.Name == "finish" {
				return finishResult(call.Arguments, assignment.Reviewer)
			}
			output, err := broker.Call(ctx, call.Name, call.Arguments)
			if err != nil {
				return Result{}, errors.New("gateway broker call failed")
			}
			outputs = append(outputs, map[string]any{"type": "function_call_output", "call_id": call.ID, "output": string(output)})
		}
		input = outputs
	}
	return Result{}, errors.New("OpenAI worker exceeded the turn limit")
}

func (client Client) runAnthropic(ctx context.Context, assignment Assignment, broker Broker, secret string) (Result, error) {
	messages := []any{map[string]any{"role": "user", "content": assignment.Prompt}}
	for turn := 0; turn < maxTurns; turn++ {
		body := map[string]any{"model": assignment.Model, "max_tokens": 4096, "messages": messages, "tools": anthropicTools(), "tool_choice": map[string]any{"type": "auto"}, "output_config": map[string]any{"effort": assignment.Effort}}
		data, _, err := client.request(ctx, assignment.Provider.Endpoint, map[string]string{"content-type": "application/json", "x-api-key": secret, "anthropic-version": "2023-06-01"}, body)
		if err != nil {
			return Result{}, err
		}
		var response map[string]any
		if json.Unmarshal(data, &response) != nil {
			return Result{}, errors.New("Anthropic response is malformed")
		}
		content, ok := response["content"].([]any)
		if !ok || len(content) == 0 {
			return Result{}, errors.New("Anthropic worker returned no content")
		}
		calls, err := anthropicCalls(content)
		if err != nil {
			return Result{}, err
		}
		if len(calls) == 0 {
			return Result{}, errors.New("Anthropic worker stopped without finish")
		}
		toolResults := make([]any, 0, len(calls))
		for _, call := range calls {
			if call.Name == "finish" {
				return finishResult(call.Arguments, assignment.Reviewer)
			}
			output, err := broker.Call(ctx, call.Name, call.Arguments)
			if err != nil {
				return Result{}, errors.New("gateway broker call failed")
			}
			toolResults = append(toolResults, map[string]any{"type": "tool_result", "tool_use_id": call.ID, "content": string(output)})
		}
		messages = append(messages, map[string]any{"role": "assistant", "content": content}, map[string]any{"role": "user", "content": toolResults})
	}
	return Result{}, errors.New("Anthropic worker exceeded the turn limit")
}

type toolCall struct {
	ID        string
	Name      string
	Arguments []byte
}

func openAICalls(response map[string]any) ([]toolCall, error) {
	items, ok := response["output"].([]any)
	if !ok {
		return nil, errors.New("OpenAI output is invalid")
	}
	calls := []toolCall{}
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok || object["type"] != "function_call" {
			continue
		}
		id, _ := object["call_id"].(string)
		name, _ := object["name"].(string)
		arguments, _ := object["arguments"].(string)
		if id == "" || name == "" || !json.Valid([]byte(arguments)) {
			return nil, errors.New("OpenAI tool call is invalid")
		}
		calls = append(calls, toolCall{ID: id, Name: name, Arguments: []byte(arguments)})
	}
	return calls, nil
}

func anthropicCalls(content []any) ([]toolCall, error) {
	calls := []toolCall{}
	for _, item := range content {
		object, ok := item.(map[string]any)
		if !ok || object["type"] != "tool_use" {
			continue
		}
		id, _ := object["id"].(string)
		name, _ := object["name"].(string)
		input, ok := object["input"].(map[string]any)
		if id == "" || name == "" || !ok {
			return nil, errors.New("Anthropic tool call is invalid")
		}
		data, err := json.Marshal(input)
		if err != nil {
			return nil, errors.New("Anthropic tool arguments are invalid")
		}
		calls = append(calls, toolCall{ID: id, Name: name, Arguments: data})
	}
	return calls, nil
}

func finishResult(arguments []byte, reviewer bool) (Result, error) {
	var terminal struct {
		Outcome  string                `json:"outcome"`
		Summary  string                `json:"summary"`
		Findings []string              `json:"findings"`
		Decision domain.ReviewDecision `json:"decision"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(arguments)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&terminal) != nil || (terminal.Outcome != "complete" && terminal.Outcome != "blocked") || terminal.Summary == "" || len(terminal.Summary) > 4096 || len(terminal.Findings) > 64 {
		return Result{}, errors.New("gateway finish payload is invalid")
	}
	for _, finding := range terminal.Findings {
		if finding == "" || len(finding) > 2048 || strings.ContainsAny(finding, "\x00\r\n") {
			return Result{}, errors.New("gateway finding is invalid")
		}
	}
	if reviewer {
		if !terminal.Decision.Valid() || (terminal.Outcome == "blocked" && terminal.Decision == domain.DecisionGO) {
			return Result{}, errors.New("reviewer finish decision is invalid")
		}
	} else if terminal.Decision != "" {
		return Result{}, errors.New("implementer attempted to review itself")
	}
	return Result{Summary: terminal.Summary, Findings: append([]string{}, terminal.Findings...), Decision: terminal.Decision}, nil
}

func toolDefinitions(openAI bool) []any {
	definitions := []struct {
		name, description string
		properties        map[string]any
		required          []string
	}{
		{"list_files", "List bounded files under a scoped repository path.", map[string]any{"path": map[string]any{"type": "string"}}, []string{"path"}},
		{"read_file", "Read a bounded slice of a scoped file.", map[string]any{"path": map[string]any{"type": "string"}, "offset": map[string]any{"type": "integer"}, "limit": map[string]any{"type": "integer"}}, []string{"path", "offset", "limit"}},
		{"search", "Search literal text in scoped files.", map[string]any{"path": map[string]any{"type": "string"}, "query": map[string]any{"type": "string"}}, []string{"path", "query"}},
		{"apply_patch", "Apply one validated unified patch inside scope.", map[string]any{"patch": map[string]any{"type": "string"}}, []string{"patch"}},
		{"git_status", "Read bounded Git status.", map[string]any{}, []string{}},
		{"git_diff", "Read bounded Git diff.", map[string]any{}, []string{}},
		{"run_command", "Run one exact configured command by name.", map[string]any{"name": map[string]any{"type": "string"}}, []string{"name"}},
		{"memory_query", "Query private Level 7 repository memory.", map[string]any{"query": map[string]any{"type": "string"}}, []string{"query"}},
		{"finish", "Return the bounded terminal result.", map[string]any{"outcome": map[string]any{"type": "string", "enum": []string{"complete", "blocked"}}, "summary": map[string]any{"type": "string"}, "findings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "decision": map[string]any{"type": "string", "enum": []string{"GO", "NO_GO"}}}, []string{"outcome", "summary", "findings"}},
	}
	tools := make([]any, 0, len(definitions))
	for _, definition := range definitions {
		schema := map[string]any{"type": "object", "additionalProperties": false, "properties": definition.properties, "required": definition.required}
		if openAI {
			tools = append(tools, map[string]any{"type": "function", "name": definition.name, "description": definition.description, "parameters": schema, "strict": true})
		} else {
			tools = append(tools, map[string]any{"name": definition.name, "description": definition.description, "input_schema": schema})
		}
	}
	return tools
}

func openAITools() []any    { return toolDefinitions(true) }
func anthropicTools() []any { return toolDefinitions(false) }

func DebugResult(result Result) string {
	return fmt.Sprintf("summary=%q findings=%d decision=%s", result.Summary, len(result.Findings), result.Decision)
}
