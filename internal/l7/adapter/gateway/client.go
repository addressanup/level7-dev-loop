// Package gateway implements explicitly configured OpenAI Responses and Anthropic Messages workers.
package gateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	maxHTTPBody = 8 << 20
	maxTurns    = 64
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Broker interface {
	Call(context.Context, string, []byte) ([]byte, error)
}

// QuotaError carries only the provider-advertised natural reset boundary. It
// intentionally excludes response bodies and credential-bearing request data.
type QuotaError struct{ ResetAtUTC string }

func (failure *QuotaError) Error() string { return "gateway quota is exhausted" }

type Client struct {
	http    Doer
	resolve func(string) (processadapter.Executable, error)
	run     func(context.Context, processadapter.Request) (processadapter.Result, error)
	now     func() time.Time
}

func New() Client {
	transport := &http.Transport{Proxy: nil}
	return NewWith(&http.Client{Transport: transport, Timeout: 5 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("gateway redirects are disabled") }}, processadapter.Resolve, (processadapter.Runner{}).Run, time.Now)
}

func NewWith(httpClient Doer, resolve func(string) (processadapter.Executable, error), run func(context.Context, processadapter.Request) (processadapter.Result, error), now func() time.Time) Client {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	if now == nil {
		now = time.Now
	}
	return Client{http: httpClient, resolve: resolve, run: run, now: now}
}

func (client Client) Probe(ctx context.Context, provider orchestrationconfig.Provider) domain.ProviderSnapshot {
	snapshot := domain.ProviderSnapshot{Schema: domain.OrchestrationSchema, ID: provider.ID, Kind: provider.Kind, Authentication: domain.AuthUnknown, CatalogComplete: false, Models: []domain.ModelCapability{}, ObservedAtUTC: client.now().UTC().Format(time.RFC3339), Next: "repair gateway configuration and retry l7 providers probe"}
	if client.http == nil || (provider.Kind != domain.ProviderKindOpenAIResponses && provider.Kind != domain.ProviderKindAnthropic) {
		snapshot.Diagnostic = "gateway client is unavailable"
		return snapshot
	}
	secret, err := client.credential(ctx, provider.Credential)
	if err != nil {
		snapshot.Authentication = domain.AuthUnauthenticated
		snapshot.Diagnostic = "gateway credential reference could not be resolved"
		return snapshot
	}
	models := append([]orchestrationconfig.Model{}, provider.Models...)
	catalogFailed := false
	if provider.CatalogURL != "" {
		discovered, catalogStatus, catalogErr := client.catalog(ctx, provider, secret)
		if catalogErr != nil {
			catalogFailed = true
			if catalogStatus == http.StatusUnauthorized || catalogStatus == http.StatusForbidden {
				snapshot.Authentication = domain.AuthUnauthenticated
				snapshot.Diagnostic = "gateway catalog rejected the referenced credential"
				return snapshot
			}
		} else {
			models = mergeModels(models, discovered)
			snapshot.CatalogComplete = true
		}
	}
	authenticated := false
	unauthorized := false
	quotaLimited := false
	quotaReset := ""
	for _, configured := range models {
		status, probeErr := client.probeModel(ctx, provider, configured.ID, secret)
		verified := probeErr == nil && status >= 200 && status < 300
		if verified {
			authenticated = true
		}
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			unauthorized = true
		}
		var quota *QuotaError
		if errors.As(probeErr, &quota) {
			quotaLimited, quotaReset = true, quota.ResetAtUTC
		}
		snapshot.Models = append(snapshot.Models, domain.ModelCapability{
			ID: configured.ID, DisplayName: configured.ID, ContextWindow: configured.ContextWindow,
			Languages:     append([]string{}, configured.Languages...),
			SupportsTools: configured.SupportsTools, SupportsEditing: configured.SupportsEditing, SupportsResume: configured.SupportsResume,
			Efforts: append([]domain.ReasoningEffort{}, configured.Efforts...), CostClass: configured.CostClass, LatencyClass: configured.LatencyClass, Verified: verified,
		})
	}
	if quotaLimited {
		snapshot.Quota = domain.QuotaState{Limited: true, ResetAtUTC: quotaReset, Source: "gateway"}
	}
	if authenticated {
		snapshot.Authentication = domain.AuthAuthenticated
		if catalogFailed {
			snapshot.Diagnostic = "gateway credential and configured models were verified; catalog retrieval failed closed"
		} else if provider.CatalogURL != "" {
			snapshot.Diagnostic = "gateway credential, catalog, and model capabilities were verified"
		} else {
			snapshot.Diagnostic = "gateway credential and configured model capabilities were verified"
		}
		snapshot.Next = "run l7 route explain"
	} else if unauthorized {
		snapshot.Authentication = domain.AuthUnauthenticated
		snapshot.Diagnostic = "gateway rejected the referenced credential"
	} else {
		snapshot.Diagnostic = "gateway did not verify any configured model"
	}
	return snapshot
}

type catalogDocument struct {
	Schema int                         `json:"schema"`
	Models []orchestrationconfig.Model `json:"models"`
	Data   []orchestrationconfig.Model `json:"data"`
}

func (client Client) catalog(ctx context.Context, provider orchestrationconfig.Provider, secret string) ([]orchestrationconfig.Model, int, error) {
	data, status, err := client.requestMethod(ctx, http.MethodGet, provider.CatalogURL, providerHeaders(provider.Kind, secret), nil)
	if err != nil {
		return nil, status, err
	}
	var document catalogDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil || decoder.Decode(&struct{}{}) != io.EOF || document.Schema != domain.OrchestrationSchema || (len(document.Models) == 0) == (len(document.Data) == 0) {
		return nil, status, errors.New("gateway catalog is malformed")
	}
	models := document.Models
	if len(models) == 0 {
		models = document.Data
	}
	if len(models) > 128 {
		return nil, status, errors.New("gateway catalog is unbounded")
	}
	seen := make(map[string]bool, len(models))
	for _, model := range models {
		if seen[model.ID] || orchestrationconfig.ValidateModel(model) != nil {
			return nil, status, errors.New("gateway catalog capabilities are invalid")
		}
		seen[model.ID] = true
	}
	return append([]orchestrationconfig.Model{}, models...), status, nil
}

func mergeModels(configured, discovered []orchestrationconfig.Model) []orchestrationconfig.Model {
	byID := make(map[string]orchestrationconfig.Model, len(configured)+len(discovered))
	for _, model := range discovered {
		byID[model.ID] = model
	}
	for _, model := range configured {
		byID[model.ID] = model
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	models := make([]orchestrationconfig.Model, 0, len(ids))
	for _, id := range ids {
		models = append(models, byID[id])
	}
	return models
}

func providerHeaders(kind domain.ProviderKind, secret string) map[string]string {
	headers := map[string]string{"content-type": "application/json"}
	if kind == domain.ProviderKindOpenAIResponses {
		headers["authorization"] = "Bearer " + secret
	} else {
		headers["x-api-key"] = secret
		headers["anthropic-version"] = "2023-06-01"
	}
	return headers
}

func (client Client) probeModel(ctx context.Context, provider orchestrationconfig.Provider, model, secret string) (int, error) {
	body := map[string]any{}
	headers := map[string]string{"content-type": "application/json"}
	if provider.Kind == domain.ProviderKindOpenAIResponses {
		body = map[string]any{"model": model, "input": "Reply exactly OK.", "max_output_tokens": 8, "store": false}
		headers["authorization"] = "Bearer " + secret
	} else {
		body = map[string]any{"model": model, "max_tokens": 8, "messages": []any{map[string]any{"role": "user", "content": "Reply exactly OK."}}}
		headers["x-api-key"] = secret
		headers["anthropic-version"] = "2023-06-01"
	}
	_, status, err := client.request(ctx, provider.Endpoint, headers, body)
	return status, err
}

func (client Client) credential(ctx context.Context, reference orchestrationconfig.Credential) (string, error) {
	if reference.Source == "env" {
		value, ok := os.LookupEnv(reference.Reference)
		if !ok || !safeSecret(value) {
			return "", errors.New("environment credential is unavailable")
		}
		return value, nil
	}
	if reference.Source != "keychain" {
		return "", errors.New("credential source is unsupported")
	}
	service, account, ok := strings.Cut(reference.Reference, "/")
	if !ok || service == "" || account == "" {
		return "", errors.New("Keychain reference is invalid")
	}
	executable, err := client.resolve("security")
	if err != nil {
		return "", errors.New("macOS Keychain client is unavailable")
	}
	result, err := client.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: []string{"find-generic-password", "-s", service, "-a", account, "-w"}, Directory: "/",
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 10 * time.Second,
	})
	if err != nil || result.ExitCode != 0 {
		return "", errors.New("Keychain credential lookup failed")
	}
	value := strings.TrimSpace(string(result.Stdout))
	if !safeSecret(value) {
		return "", errors.New("Keychain credential is invalid")
	}
	return value, nil
}

func (client Client) request(ctx context.Context, endpoint string, headers map[string]string, body any) ([]byte, int, error) {
	return client.requestMethod(ctx, http.MethodPost, endpoint, headers, body)
}

func (client Client) requestMethod(ctx context.Context, method, endpoint string, headers map[string]string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil || len(data) > maxHTTPBody {
			return nil, 0, errors.New("gateway request is invalid or unbounded")
		}
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, 0, errors.New("gateway request could not be created")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return nil, 0, errors.New("gateway request failed")
	}
	defer response.Body.Close()
	responseData, readErr := io.ReadAll(io.LimitReader(response.Body, maxHTTPBody+1))
	if readErr != nil || len(responseData) > maxHTTPBody {
		return nil, response.StatusCode, errors.New("gateway response is invalid or unbounded")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if response.StatusCode == http.StatusTooManyRequests {
			return nil, response.StatusCode, &QuotaError{ResetAtUTC: retryReset(response.Header, client.now())}
		}
		return nil, response.StatusCode, errors.New("gateway returned a non-success status")
	}
	if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "text/event-stream") {
		responseData, readErr = decodeEventStream(responseData)
		if readErr != nil {
			return nil, response.StatusCode, readErr
		}
	}
	if len(responseData) < 2 || !json.Valid(responseData) {
		return nil, response.StatusCode, errors.New("gateway returned malformed JSON")
	}
	return responseData, response.StatusCode, nil
}

func decodeEventStream(data []byte) ([]byte, error) {
	if len(data) < 2 || len(data) > maxHTTPBody {
		return nil, errors.New("gateway event stream is invalid or unbounded")
	}
	type anthropicBlock struct {
		value   map[string]any
		partial strings.Builder
	}
	blocks := make(map[int]*anthropicBlock)
	completed := []byte(nil)
	anthropicComplete := false
	events := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		if bytes.Equal(payload, []byte("[DONE]")) || len(payload) == 0 {
			continue
		}
		events++
		if events > 4096 || !json.Valid(payload) {
			return nil, errors.New("gateway event stream contains malformed data")
		}
		var event map[string]any
		if json.Unmarshal(payload, &event) != nil {
			return nil, errors.New("gateway event stream contains malformed data")
		}
		if response, ok := event["response"].(map[string]any); ok && event["type"] == "response.completed" {
			completed, _ = json.Marshal(response)
			continue
		}
		index, hasIndex := integerIndex(event["index"])
		switch event["type"] {
		case "content_block_start":
			value, ok := event["content_block"].(map[string]any)
			if !hasIndex || !ok || len(blocks) >= 1024 {
				return nil, errors.New("gateway event stream has an invalid content block")
			}
			blocks[index] = &anthropicBlock{value: value}
		case "content_block_delta":
			delta, ok := event["delta"].(map[string]any)
			block := blocks[index]
			partial, partialOK := delta["partial_json"].(string)
			if hasIndex && ok && block != nil && partialOK {
				if block.partial.Len()+len(partial) > 1<<20 {
					return nil, errors.New("gateway event stream tool input is unbounded")
				}
				block.partial.WriteString(partial)
			}
		case "message_stop":
			anthropicComplete = true
		}
	}
	if scanner.Err() != nil {
		return nil, errors.New("gateway event stream could not be read")
	}
	if len(completed) > 0 {
		return completed, nil
	}
	if len(blocks) == 0 || !anthropicComplete {
		return nil, errors.New("gateway event stream ended without a terminal response")
	}
	indices := make([]int, 0, len(blocks))
	for index := range blocks {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	content := make([]any, 0, len(indices))
	for _, index := range indices {
		block := blocks[index]
		if block.value["type"] == "tool_use" {
			if block.partial.Len() > 0 {
				var input map[string]any
				if json.Unmarshal([]byte(block.partial.String()), &input) != nil {
					return nil, errors.New("gateway event stream tool input is malformed")
				}
				block.value["input"] = input
			} else if _, ok := block.value["input"].(map[string]any); !ok {
				block.value["input"] = map[string]any{}
			}
		}
		content = append(content, block.value)
	}
	return json.Marshal(map[string]any{"content": content})
}

func integerIndex(value any) (int, bool) {
	number, ok := value.(float64)
	if !ok || number < 0 || number > 1024 || number != float64(int(number)) {
		return 0, false
	}
	return int(number), true
}

func retryReset(header http.Header, now time.Time) string {
	value := strings.TrimSpace(header.Get("Retry-After"))
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 && seconds <= 7*24*60*60 {
		return now.UTC().Add(time.Duration(seconds) * time.Second).Format(time.RFC3339)
	}
	if parsed, err := http.ParseTime(value); err == nil && parsed.After(now) {
		return parsed.UTC().Format(time.RFC3339)
	}
	return ""
}

func safeSecret(value string) bool {
	return len(value) >= 8 && len(value) <= 32<<10 && !strings.ContainsAny(value, "\x00\r\n")
}
