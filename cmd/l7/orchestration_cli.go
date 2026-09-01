package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/cyber"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/discovery"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/gateway"
	gitadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/git"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/headless"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/headlessworker"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/memory"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/state"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type orchestrationEnvelope struct {
	Schema  int    `json:"schema"`
	Outcome string `json:"outcome"`
	Code    string `json:"code"`
	Command string `json:"command"`
	State   string `json:"state"`
	Version string `json:"version"`
	Message string `json:"message"`
	Next    string `json:"next"`
	Data    any    `json:"data,omitempty"`
}

func runOrchestrationCommand(ctx context.Context, arguments []string, cwd string, input io.Reader, stdout, stderr io.Writer) (bool, int) {
	commandCandidate := ""
	for _, argument := range arguments {
		if argument != "--json" {
			commandCandidate = argument
			break
		}
	}
	if !orchestrationCommand(commandCandidate) {
		return false, 0
	}
	filtered, jsonOutput, err := orchestrationArguments(arguments)
	if err != nil {
		return true, writeOrchestration(stdout, stderr, jsonOutput, failedEnvelope("", "L7-ORCH-001", "invalid", err.Error(), "run l7 help"))
	}
	if len(filtered) == 0 {
		return false, 0
	}
	command := filtered[0]
	if command == "mcp" {
		if len(filtered) != 1 || jsonOutput {
			return true, writeOrchestration(stdout, stderr, false, failedEnvelope("mcp", "L7-MCP-001", "invalid", "mcp accepts no options", "invoke l7 mcp over stdio"))
		}
		if input == nil {
			return true, writeOrchestration(stdout, stderr, false, failedEnvelope("mcp", "L7-MCP-001", "invalid", "mcp requires stdio input", "invoke l7 mcp over stdio"))
		}
		return true, serveMCP(ctx, cwd, input, stdout, stderr)
	}
	envelope, runErr := executeOrchestration(ctx, command, filtered[1:], cwd)
	if runErr != nil {
		envelope = failedEnvelope(command, "L7-ORCH-002", "failed", runErr.Error(), nextForFailure(command))
	}
	return true, writeOrchestration(stdout, stderr, jsonOutput, envelope)
}

func orchestrationArguments(arguments []string) ([]string, bool, error) {
	if len(arguments) > maxArguments {
		return nil, false, errors.New("too many arguments")
	}
	filtered := make([]string, 0, len(arguments))
	jsonOutput := false
	for _, argument := range arguments {
		if len(argument) > maxArgumentBytes || strings.ContainsRune(argument, 0) {
			return nil, jsonOutput, errors.New("argument is invalid or exceeds its size limit")
		}
		if argument == "--json" {
			if jsonOutput {
				return nil, true, errors.New("duplicate --json flag")
			}
			jsonOutput = true
			continue
		}
		filtered = append(filtered, argument)
	}
	return filtered, jsonOutput, nil
}

func orchestrationCommand(command string) bool {
	switch command {
	case "onboard", "providers", "route", "sync", "cyber", "headless", "mcp":
		return true
	default:
		return false
	}
}

func executeOrchestration(ctx context.Context, command string, arguments []string, cwd string) (orchestrationEnvelope, error) {
	gitClient, err := gitadapter.New("", gitadapter.DefaultMaxOutput, gitadapter.DefaultMaxPaths)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	location, err := gitClient.Locate(ctx, cwd)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	switch command {
	case "onboard":
		return onboardCommand(ctx, location, arguments)
	case "providers":
		return providersCommand(ctx, location, arguments)
	case "route":
		return routeCommand(location, arguments)
	case "sync":
		return syncCommand(ctx, location, arguments)
	case "cyber":
		return cyberCommand(ctx, location, arguments)
	case "headless":
		return headlessCommand(ctx, location, arguments)
	default:
		return orchestrationEnvelope{}, errors.New("unknown orchestration command")
	}
}

func onboardCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) != 1 || (arguments[0] != "--status" && arguments[0] != "--apply") {
		return orchestrationEnvelope{}, errors.New("onboard requires exactly --status or --apply")
	}
	if arguments[0] == "--apply" {
		configuration, changed, err := orchestrationconfig.Apply(location.Root)
		if err != nil {
			return orchestrationEnvelope{}, err
		}
		message := "orchestration policy already enabled"
		if changed {
			message = "created or updated tracked .l7/orchestration.json"
		}
		return passEnvelope("onboard", "L7-ONBOARD-000", "applied", message, "run l7 providers probe", map[string]any{"configuration": configuration, "changed": changed, "repository": location}), nil
	}
	configuration, err := orchestrationconfig.Load(location.Root)
	configured := true
	if errors.Is(err, os.ErrNotExist) {
		configuration = orchestrationconfig.Default()
		configured = false
	} else if err != nil {
		return orchestrationEnvelope{}, err
	}
	providers := discovery.New().Discover(ctx, configuration)
	_, providerState, _ := state.LoadProviderSnapshots(location.CommonDir)
	_, memoryErr := memory.LoadGraph(location.CommonDir)
	next := "run l7 onboard --apply"
	stateName := "inspection-complete"
	if configured && configuration.Features.Orchestration {
		next = "run l7 providers probe"
		if providerState {
			next = "run l7 sync --incremental"
		}
	}
	data := map[string]any{
		"repository": location, "configured": configured, "configuration": configuration,
		"providers": providers, "provider_snapshot": providerState, "memory_available": memoryErr == nil,
		"transitions": []string{"onboard", "providers probe", "sync", "brief", "implement", "verify", "independent review", "local merge", "release", "deploy"},
	}
	return passEnvelope("onboard", "L7-ONBOARD-000", stateName, "read-only project and host inspection completed", next, data), nil
}

func providersCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) != 1 || (arguments[0] != "list" && arguments[0] != "probe") {
		return orchestrationEnvelope{}, errors.New("providers requires list or probe")
	}
	configuration, err := requireOrchestration(location.Root)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	discoverer := discovery.New()
	snapshots := discoverer.Discover(ctx, configuration)
	if arguments[0] == "probe" {
		for index, snapshot := range snapshots {
			switch snapshot.Kind {
			case domain.ProviderKindClaudeCLI:
				snapshots[index] = discoverer.ProbeClaude(ctx, snapshot, location.Root)
			case domain.ProviderKindOpenAIResponses, domain.ProviderKindAnthropic:
				for _, provider := range configuration.Providers {
					if provider.ID == snapshot.ID {
						snapshots[index] = gateway.New().Probe(ctx, provider)
						break
					}
				}
			}
		}
		if err := state.SaveProviderSnapshots(location.CommonDir, snapshots); err != nil {
			return orchestrationEnvelope{}, err
		}
		return passEnvelope("providers probe", "L7-PROVIDER-000", "probed", "explicit provider and configured gateway probes completed", "run l7 route explain", snapshots), nil
	}
	if saved, found, loadErr := state.LoadProviderSnapshots(location.CommonDir); loadErr == nil && found {
		snapshots = saved
	}
	return passEnvelope("providers list", "L7-PROVIDER-000", "listed", "provider capabilities listed without gateway traffic", "run l7 providers probe to refresh verified capabilities", snapshots), nil
}

func routeCommand(location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) == 0 || arguments[0] != "explain" {
		return orchestrationEnvelope{}, errors.New("route requires explain")
	}
	task := domain.TaskProfile{Schema: domain.OrchestrationSchema, ID: "interactive-task", Summary: "interactive Level 7 task", Complexity: domain.ComplexityC2, RiskTier: domain.TierProduct, NeedsTools: true, NeedsEditing: true}
	seen := map[string]bool{}
	for index := 1; index < len(arguments); index++ {
		flag := arguments[index]
		if flag == "--tools" || flag == "--edit" || flag == "--resume" || flag == "--review" {
			if seen[flag] {
				return orchestrationEnvelope{}, fmt.Errorf("duplicate %s", flag)
			}
			seen[flag] = true
			switch flag {
			case "--tools":
				task.NeedsTools = true
			case "--edit":
				task.NeedsEditing = true
			case "--resume":
				task.NeedsResume = true
			case "--review":
				task.IndependentReview = true
			}
			continue
		}
		if index+1 >= len(arguments) || seen[flag] {
			return orchestrationEnvelope{}, fmt.Errorf("route option %s is missing or duplicated", flag)
		}
		seen[flag] = true
		value := arguments[index+1]
		index++
		switch flag {
		case "--task":
			task.ID, task.Summary = value, value
		case "--complexity":
			task.Complexity = domain.Complexity(strings.ToUpper(value))
		case "--risk":
			tier, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return orchestrationEnvelope{}, errors.New("risk must be 1, 2, or 3")
			}
			task.RiskTier = domain.RiskTier(tier)
		case "--context":
			count, parseErr := strconv.Atoi(value)
			if parseErr != nil || count < 0 {
				return orchestrationEnvelope{}, errors.New("context must be a non-negative token count")
			}
			task.ContextTokens = count
		case "--language":
			task.Languages = append(task.Languages, value)
		case "--work-kind":
			task.WorkKinds = append(task.WorkKinds, value)
		case "--prior-failures":
			count, parseErr := strconv.Atoi(value)
			if parseErr != nil || count < 0 {
				return orchestrationEnvelope{}, errors.New("prior failures must be a non-negative count")
			}
			task.PriorFailures = count
		case "--implementer-provider":
			task.ImplementerProvider = value
		case "--implementer-model":
			task.ImplementerModel = value
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown route option %s", flag)
		}
	}
	if task.ID == "" || len(task.ID) > 128 || !task.Complexity.Valid() || !task.RiskTier.Valid() {
		return orchestrationEnvelope{}, errors.New("route task profile is invalid")
	}
	snapshots, found, err := state.LoadProviderSnapshots(location.CommonDir)
	if err != nil || !found {
		return orchestrationEnvelope{}, errors.New("verified provider snapshot is unavailable; run l7 providers probe")
	}
	decision := domain.Route(task, snapshots)
	decision.DecisionUTC = time.Now().UTC().Format(time.RFC3339)
	if err := state.SaveRouteDecision(location.CommonDir, decision); err != nil {
		return orchestrationEnvelope{}, err
	}
	outcome, stateName, message := "PASS", "selected", "qualified model and effort selected"
	if decision.ProviderID == "" {
		outcome, stateName, message = "BLOCKED", "unroutable", "no authenticated verified model satisfies the task profile"
	}
	envelope := passEnvelope("route explain", "L7-ROUTE-000", stateName, message, decision.Next, map[string]any{"task": task, "decision": decision})
	envelope.Outcome = outcome
	return envelope, nil
}

func syncCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) == 0 {
		return orchestrationEnvelope{}, errors.New("sync requires --incremental, --rebuild, or --query <text>")
	}
	configuration, err := requireOrchestration(location.Root)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if !configuration.Features.Sync {
		return orchestrationEnvelope{}, errors.New("Sync is default OFF; enable it with l7 onboard --apply")
	}
	adapter := memory.New(memory.NewAppleEmbedder())
	if arguments[0] == "--query" {
		if len(arguments) != 2 {
			return orchestrationEnvelope{}, errors.New("sync --query requires one bounded query")
		}
		matches, queryErr := adapter.Query(ctx, location.CommonDir, arguments[1], 20)
		if queryErr != nil {
			return orchestrationEnvelope{}, queryErr
		}
		return passEnvelope("sync --query", "L7-SYNC-000", "queried", "hybrid private memory query completed", "use the returned node IDs as bounded context", matches), nil
	}
	if len(arguments) != 1 || (arguments[0] != "--incremental" && arguments[0] != "--rebuild") {
		return orchestrationEnvelope{}, errors.New("sync requires exactly --incremental or --rebuild")
	}
	graph, syncErr := adapter.Sync(ctx, location.Root, location.CommonDir, configuration.Memory)
	if syncErr != nil {
		return orchestrationEnvelope{}, syncErr
	}
	return passEnvelope("sync "+arguments[0], "L7-SYNC-000", "synchronized", "Git-bound codebase memory graph and indexes synchronized", graph.Next, graph), nil
}

func cyberCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	configuration, err := requireOrchestration(location.Root)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if len(arguments) > 0 && arguments[0] == "remediate" {
		if len(arguments) != 3 || arguments[1] != "--report" {
			return orchestrationEnvelope{}, errors.New("cyber remediate requires --report <id>")
		}
		report, loadErr := cyber.LoadReport(location.CommonDir, arguments[2])
		if loadErr != nil {
			return orchestrationEnvelope{}, loadErr
		}
		brief, briefErr := cyber.RemediationBrief(report)
		if briefErr != nil {
			return orchestrationEnvelope{}, briefErr
		}
		path := filepath.Join(location.Root, "docs", "artifacts", "changes", report.ID+"-remediation.md")
		if err := localfile.EnsureDirectory(filepath.Dir(path), 0o755); err != nil {
			return orchestrationEnvelope{}, err
		}
		if err := localfile.AtomicCreate(path, []byte(brief), 0o644); err != nil {
			return orchestrationEnvelope{}, err
		}
		return passEnvelope("cyber remediate", "L7-CYBER-000", "brief-created", "separate Tier 3 remediation brief created; no source code was modified", "obtain fresh named owner approval before remediation", map[string]string{"report_id": report.ID, "brief": path}), nil
	}
	active := false
	export := ""
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--active":
			if active {
				return orchestrationEnvelope{}, errors.New("duplicate --active")
			}
			active = true
		case "--export":
			if export != "" || index+1 >= len(arguments) {
				return orchestrationEnvelope{}, errors.New("--export requires markdown or json")
			}
			export = arguments[index+1]
			index++
			if export != "markdown" && export != "json" {
				return orchestrationEnvelope{}, errors.New("--export requires markdown or json")
			}
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown cyber option %s", arguments[index])
		}
	}
	report, auditErr := cyber.New().Audit(ctx, location.Root, location.CommonDir, configuration, active)
	if auditErr != nil {
		return orchestrationEnvelope{}, auditErr
	}
	data := any(report)
	if export == "markdown" {
		data = map[string]string{"format": "markdown", "report": cyber.ExportMarkdown(report)}
	} else if export == "json" {
		redacted, exportErr := cyber.ExportJSON(report)
		if exportErr != nil {
			return orchestrationEnvelope{}, exportErr
		}
		data = json.RawMessage(redacted)
	}
	return passEnvelope("cyber", "L7-CYBER-000", report.Mode, "Cyber audit completed without remediation mutation", report.Next, data), nil
}

func headlessCommand(ctx context.Context, location domain.RepositoryLocation, arguments []string) (orchestrationEnvelope, error) {
	if len(arguments) == 0 {
		return orchestrationEnvelope{}, errors.New("headless requires plan, start, status, resume, or cancel")
	}
	configuration, err := requireOrchestration(location.Root)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	switch arguments[0] {
	case "status":
		if len(arguments) != 1 {
			return orchestrationEnvelope{}, errors.New("headless status accepts no options")
		}
		checkpoint, loadErr := headless.LoadCheckpoint(location.CommonDir)
		if loadErr != nil {
			return orchestrationEnvelope{}, loadErr
		}
		return passEnvelope("headless status", "L7-HEADLESS-000", checkpoint.State, "durable Headless checkpoint loaded", checkpoint.Next, checkpoint), nil
	case "cancel":
		if len(arguments) != 1 {
			return orchestrationEnvelope{}, errors.New("headless cancel accepts no options")
		}
		checkpoint, cancelErr := headless.Cancel(location.CommonDir, time.Now())
		if cancelErr != nil {
			return orchestrationEnvelope{}, cancelErr
		}
		return passEnvelope("headless cancel", "L7-HEADLESS-000", checkpoint.State, "Headless run cancelled; evidence retained", checkpoint.Next, checkpoint), nil
	case "plan":
		return headlessPlan(location, configuration, arguments[1:])
	case "start":
		return headlessStart(ctx, location, configuration, arguments[1:])
	case "resume":
		if len(arguments) != 1 {
			return orchestrationEnvelope{}, errors.New("headless resume accepts no options")
		}
		if !configuration.Features.Headless {
			return orchestrationEnvelope{}, errors.New("Headless is default OFF; enable it explicitly in .l7/orchestration.json")
		}
		executor, executorErr := headlessworker.New(location.Root, location.CommonDir, configuration)
		if executorErr != nil {
			return orchestrationEnvelope{}, executorErr
		}
		checkpoint, resumeErr := headless.NewEngine().Resume(ctx, location.CommonDir, executor)
		if resumeErr != nil {
			return orchestrationEnvelope{}, resumeErr
		}
		return passEnvelope("headless resume", "L7-HEADLESS-000", checkpoint.State, "Headless run resumed from durable state", checkpoint.Next, checkpoint), nil
	default:
		return orchestrationEnvelope{}, errors.New("headless requires plan, start, status, resume, or cancel")
	}
}

func headlessPlan(location domain.RepositoryLocation, configuration orchestrationconfig.File, arguments []string) (orchestrationEnvelope, error) {
	request := headless.PlanRequest{BaseCommit: location.Head, ProviderPolicy: "balanced", NetworkPolicy: "gateway-only", LocalMerge: configuration.Headless.LocalMerge}
	for index := 0; index < len(arguments); index++ {
		flag := arguments[index]
		if flag == "--local-merge" {
			request.LocalMerge = true
			continue
		}
		if index+1 >= len(arguments) {
			return orchestrationEnvelope{}, fmt.Errorf("Headless option %s is missing its value", flag)
		}
		value := arguments[index+1]
		index++
		switch flag {
		case "--objective":
			request.ObjectivePath = value
		case "--target":
			request.TargetBranch = value
		case "--allow-path":
			request.AllowedPaths = append(request.AllowedPaths, value)
		case "--command-json":
			var argv []string
			if json.Unmarshal([]byte(value), &argv) != nil {
				return orchestrationEnvelope{}, errors.New("--command-json must be one JSON argv array")
			}
			request.AllowedCommands = append(request.AllowedCommands, argv)
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown Headless plan option %s", flag)
		}
	}
	if request.ObjectivePath == "" {
		return orchestrationEnvelope{}, errors.New("Headless plan requires --objective <concept.md|feature-manifest>")
	}
	objectivePath := request.ObjectivePath
	if !filepath.IsAbs(objectivePath) {
		objectivePath = filepath.Join(location.Root, filepath.FromSlash(objectivePath))
	}
	data, err := localfile.Read(objectivePath, 1<<20)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	request.Objective = data
	manifest, err := headless.NewPlanner().Plan(request)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if err := headless.SaveManifest(location.CommonDir, manifest); err != nil {
		return orchestrationEnvelope{}, err
	}
	warning := "WARNING: approval starts an autonomous multi-session loop through all approved Tier 1/2 waves and local merges; it stops before push, release, or deployment"
	return passEnvelope("headless plan", "L7-HEADLESS-000", "planned", warning, "run l7 headless start --run "+manifest.ID+" --digest "+manifest.Digest+" --owner <name> --role <role> --confirm", manifest), nil
}

func headlessStart(ctx context.Context, location domain.RepositoryLocation, configuration orchestrationconfig.File, arguments []string) (orchestrationEnvelope, error) {
	if !configuration.Features.Headless {
		return orchestrationEnvelope{}, errors.New("Headless is default OFF; enable it explicitly in .l7/orchestration.json")
	}
	values := map[string]string{}
	confirmed := false
	for index := 0; index < len(arguments); index++ {
		if arguments[index] == "--confirm" {
			confirmed = true
			continue
		}
		if index+1 >= len(arguments) || values[arguments[index]] != "" {
			return orchestrationEnvelope{}, errors.New("Headless start option is missing or duplicated")
		}
		values[arguments[index]] = arguments[index+1]
		index++
	}
	if !confirmed || values["--run"] == "" || values["--digest"] == "" || values["--owner"] == "" || values["--role"] == "" {
		return orchestrationEnvelope{}, errors.New("Headless start requires --run, --digest, --owner, --role, and --confirm")
	}
	manifest, err := headless.LoadManifest(location.CommonDir, values["--run"])
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	if manifest.BaseCommit != location.Head {
		return orchestrationEnvelope{}, errors.New("Headless manifest base is stale; re-plan against the exact current head")
	}
	approval, err := headless.Approve(location.CommonDir, manifest, values["--digest"], values["--owner"], values["--role"], time.Now())
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	executor, err := headlessworker.New(location.Root, location.CommonDir, configuration)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	checkpoint, err := headless.NewEngine().Start(ctx, location.CommonDir, manifest, approval, executor)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	return passEnvelope("headless start", "L7-HEADLESS-000", checkpoint.State, "Headless approval recorded and execution checkpoint started", checkpoint.Next, checkpoint), nil
}

func requireOrchestration(root string) (orchestrationconfig.File, error) {
	configuration, err := orchestrationconfig.Load(root)
	if errors.Is(err, os.ErrNotExist) {
		return configuration, errors.New("orchestration policy is absent; run l7 onboard --apply")
	}
	if err != nil {
		return configuration, err
	}
	if !configuration.Features.Orchestration {
		return configuration, errors.New("orchestration is default OFF; run l7 onboard --apply")
	}
	return configuration, nil
}

func passEnvelope(command, code, stateName, message, next string, data any) orchestrationEnvelope {
	return orchestrationEnvelope{Schema: domain.OrchestrationSchema, Outcome: "PASS", Code: code, Command: command, State: stateName, Version: version, Message: message, Next: next, Data: data}
}

func failedEnvelope(command, code, stateName, message, next string) orchestrationEnvelope {
	return orchestrationEnvelope{Schema: domain.OrchestrationSchema, Outcome: "FAILED", Code: code, Command: command, State: stateName, Version: version, Message: message, Next: next}
}

func nextForFailure(command string) string {
	switch command {
	case "onboard":
		return "repair Git or orchestration policy, then run l7 onboard --status"
	case "providers":
		return "run l7 onboard --status and repair the provider diagnostic"
	case "route":
		return "run l7 providers probe before retrying route explanation"
	case "sync":
		return "repair private memory state or run l7 sync --rebuild"
	case "cyber":
		return "repair the reported isolation or policy blocker and retry"
	case "headless":
		return "run l7 headless status and preserve the existing checkpoint"
	default:
		return "run l7 help"
	}
}

func writeOrchestration(stdout, stderr io.Writer, jsonOutput bool, envelope orchestrationEnvelope) int {
	data, err := json.Marshal(envelope)
	if err != nil {
		fmt.Fprintln(stderr, "FAILED code=L7-ORCH-003 message=render-failed")
		return 1
	}
	if !jsonOutput {
		fmt.Fprintf(stdout, "%s code=%q command=%q state=%q version=%q message=%q next=%q\n", envelope.Outcome, envelope.Code, envelope.Command, envelope.State, envelope.Version, envelope.Message, envelope.Next)
		if envelope.Data != nil {
			payload, marshalErr := json.MarshalIndent(envelope.Data, "", "  ")
			if marshalErr != nil {
				fmt.Fprintln(stderr, "FAILED code=L7-ORCH-003 message=render-failed")
				return 1
			}
			_, _ = fmt.Fprintf(stdout, "data=%s\n", payload)
		}
	} else {
		data = append(data, '\n')
		if written, writeErr := stdout.Write(data); writeErr != nil || written != len(data) {
			fmt.Fprintln(stderr, "FAILED code=L7-ORCH-003 message=write-failed")
			return 1
		}
	}
	switch envelope.Outcome {
	case "PASS":
		return 0
	case "BLOCKED":
		return 2
	default:
		return 1
	}
}
