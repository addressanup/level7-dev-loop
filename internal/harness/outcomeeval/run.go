package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"
)

type binaryInfo struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256,omitempty"`
	Version string `json:"version"`
}

type protocolRef struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

type runManifest struct {
	Schema          int         `json:"schema"`
	RunID           string      `json:"run_id"`
	Protocol        protocolRef `json:"protocol"`
	CorpusSHA256    string      `json:"corpus_sha256"`
	Mode            string      `json:"mode"`
	Tasks           []string    `json:"tasks"`
	Rounds          int         `json:"rounds"`
	Source          string      `json:"source"`
	L7              binaryInfo  `json:"l7"`
	Codex           binaryInfo  `json:"codex"`
	Route           route       `json:"route"`
	Host            string      `json:"host"`
	Python          string      `json:"python"`
	GradingSandbox  bool        `json:"grading_sandbox"`
	WorkDirectories []string    `json:"work_directories"`
	StartedUTC      string      `json:"started_utc"`
	UpdatedUTC      string      `json:"updated_utc"`
	Complete        bool        `json:"complete"`
}

type plannedTrial struct {
	Sequence int
	Round    int
	Task     string
	Arm      string
}

// planTrials alternates the arm order per round, so drift in time, load, or
// quota affects both arms of a pair alike.
func planTrials(tasks []string, rounds int) []plannedTrial {
	plan := []plannedTrial{}
	for round := 1; round <= rounds; round++ {
		arms := []string{armCrew, armPlain}
		if round%2 == 0 {
			arms = []string{armPlain, armCrew}
		}
		for _, id := range tasks {
			for _, arm := range arms {
				plan = append(plan, plannedTrial{Sequence: len(plan) + 1, Round: round, Task: id, Arm: arm})
			}
		}
	}
	return plan
}

type runOptions struct {
	l7      string
	codex   string
	out     string
	tasks   []string
	rounds  int
	smoke   bool
	confirm bool
	resume  string
}

func runEvaluation(ctx context.Context, options runOptions, stdout, stderr io.Writer) error {
	loaded, err := loadCorpus(embeddedCorpus)
	if err != nil {
		return err
	}
	manifest, records, runDirectory, err := openRun(loaded, options)
	if err != nil {
		return err
	}
	plan := planTrials(manifest.Tasks, manifest.Rounds)
	done := map[string]bool{}
	for _, record := range records {
		done[trialKey(record.Task, record.Round, record.Arm)] = true
	}
	remaining := []plannedTrial{}
	for _, trial := range plan {
		if !done[trialKey(trial.Task, trial.Round, trial.Arm)] {
			remaining = append(remaining, trial)
		}
	}
	fmt.Fprintf(stdout, "outcome-eval: %s run %s: %d tasks x %d rounds x 2 arms = %d trials, %d to run\n", manifest.Mode, manifest.RunID, len(manifest.Tasks), manifest.Rounds, len(plan), len(remaining))
	fmt.Fprintf(stdout, "outcome-eval: timeouts %ds plain, %ds crew; each crew trial runs an implementer and a reviewer, plus any repairs\n", loaded.protocol.Timeouts.Plain, loaded.protocol.Timeouts.Crew)
	if manifest.Mode == "live" && !options.confirm {
		fmt.Fprintln(stdout, "outcome-eval: dry run; nothing started. Add --confirm to run it, which spends the owner's Codex quota and may take hours.")
		return nil
	}
	if options.resume == "" {
		if err := os.MkdirAll(filepath.Dir(runDirectory), 0o755); err != nil {
			return err
		}
		if err := os.Mkdir(runDirectory, 0o755); err != nil {
			return fmt.Errorf("create run directory: %w", err)
		}
	}
	h, err := newHarness(ctx, loaded, options, manifest.Mode)
	if err != nil {
		return err
	}
	if err := h.discoverRoute(ctx, &manifest); err != nil {
		return err
	}
	manifest.WorkDirectories = append(manifest.WorkDirectories, h.work)
	if err := saveManifest(runDirectory, manifest); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "outcome-eval: route %s/%s@%s (expected reviewer %s); trial repositories in %s\n", manifest.Route.Provider, manifest.Route.Model, manifest.Route.Effort, manifest.Route.Reviewer, h.work)
	for _, trial := range remaining {
		if ctx.Err() != nil {
			break
		}
		value, _ := loaded.task(trial.Task)
		var record trialRecord
		if trial.Arm == armCrew {
			record = h.runCrew(ctx, value, trial.Round, trial.Sequence)
		} else {
			record = h.runPlain(ctx, value, trial.Round, trial.Sequence)
		}
		if ctx.Err() != nil {
			fmt.Fprintf(stdout, "outcome-eval: interrupted during trial %d; it was not recorded\n", trial.Sequence)
			break
		}
		if err := appendTrial(runDirectory, record); err != nil {
			return err
		}
		records = append(records, record)
		fmt.Fprintf(stdout, "[%d/%d] round %d %-16s %-5s %-17s %s, %d turns%s\n", trial.Sequence, len(plan), trial.Round, trial.Task, trial.Arm, record.Outcome,
			formatSeconds(record.WallSeconds), record.Meter.Turns, inlineFindings(record))
	}
	manifest.Complete = len(records) == len(plan)
	if err := saveManifest(runDirectory, manifest); err != nil {
		return err
	}
	report := buildReport(manifest, loaded, records)
	if err := writeReport(runDirectory, report, records); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "outcome-eval: Phase 4 gate %s; improvement claim %s\n", report.Phase4.Result, report.Improvement.Result)
	fmt.Fprintf(stdout, "outcome-eval: report %s\n", filepath.Join(runDirectory, "report.md"))
	if !manifest.Complete {
		fmt.Fprintf(stdout, "outcome-eval: %d of %d trials recorded; continue with --resume %s\n", len(records), len(plan), runDirectory)
	}
	return ctx.Err()
}

func inlineFindings(record trialRecord) string {
	if text := findings(record); text != "" {
		return " (" + cell(text) + ")"
	}
	return ""
}

// openRun starts a new run record or reopens one for --resume, refusing to
// mix trials from a different protocol, corpus, or selection.
func openRun(loaded corpus, options runOptions) (runManifest, []trialRecord, string, error) {
	if options.resume != "" {
		manifest, records, directory, err := loadRun(options.resume)
		if err != nil {
			return runManifest{}, nil, "", err
		}
		if manifest.Protocol.SHA256 != loaded.protocolSHA256 || manifest.CorpusSHA256 != loaded.digest {
			return runManifest{}, nil, "", errors.New("the run used a different protocol or corpus; it cannot be resumed with this harness")
		}
		if (manifest.Mode == "smoke") != options.smoke || len(options.tasks) != 0 || options.rounds != 0 {
			return runManifest{}, nil, "", errors.New("--resume takes its mode, tasks, and rounds from the run; pass only --resume, --l7, --codex, --confirm, and --smoke for a smoke run")
		}
		return manifest, records, directory, nil
	}
	tasks := loaded.protocol.Tasks
	if len(options.tasks) != 0 {
		tasks = options.tasks
		for _, id := range tasks {
			if _, found := loaded.task(id); !found {
				return runManifest{}, nil, "", fmt.Errorf("unknown task %q", id)
			}
		}
	}
	rounds := loaded.protocol.TrialsPerTask
	if options.rounds != 0 {
		rounds = options.rounds
	}
	if rounds < 1 || rounds > 10 {
		return runManifest{}, nil, "", errors.New("--rounds must be 1 to 10")
	}
	mode := "live"
	if options.smoke {
		mode = "smoke"
	}
	now := time.Now().UTC()
	manifest := runManifest{
		Schema: 1, RunID: now.Format("20060102T150405Z") + "-" + mode,
		Protocol:     protocolRef{ID: loaded.protocol.ID, Version: loaded.protocol.Version, SHA256: loaded.protocolSHA256},
		CorpusSHA256: loaded.digest, Mode: mode, Tasks: tasks, Rounds: rounds,
		StartedUTC: now.Format(time.RFC3339), WorkDirectories: []string{},
	}
	out, err := filepath.Abs(options.out)
	if err != nil {
		return runManifest{}, nil, "", err
	}
	return manifest, nil, filepath.Join(out, manifest.RunID), nil
}

func newHarness(ctx context.Context, loaded corpus, options runOptions, mode string) (*harness, error) {
	if !filepath.IsAbs(options.l7) {
		return nil, errors.New("--l7 must be an absolute path to the l7 binary under evaluation")
	}
	if info, err := os.Stat(options.l7); err != nil || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("l7 binary %s is not an executable file", options.l7)
	}
	for _, tool := range []string{"git", "python3"} {
		if _, err := exec.LookPath(tool); err != nil {
			return nil, fmt.Errorf("%s is required on PATH", tool)
		}
	}
	sandboxed := sandboxAvailable()
	if mode == "live" && !sandboxed {
		return nil, errors.New("live runs grade untrusted code and need the macOS sandbox at " + sandboxExec)
	}
	work, err := workDirectory(ctx)
	if err != nil {
		return nil, err
	}
	h := &harness{corpus: loaded, l7: options.l7, work: work, ledger: filepath.Join(work, "ledger"), temp: filepath.Join(work, "tmp"),
		judge: grader{rules: loaded.protocol, sandbox: sandboxed, timeout: time.Duration(loaded.protocol.Timeouts.Check) * time.Second}}
	bin := filepath.Join(work, "bin")
	for _, directory := range []string{bin, h.ledger, h.temp} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return nil, err
		}
	}
	config := meterConfig{Ledger: h.ledger, Fake: mode == "smoke"}
	if mode == "live" {
		codex := options.codex
		if codex == "" {
			if codex, err = exec.LookPath("codex"); err != nil {
				return nil, errors.New("codex is not on PATH; pass --codex")
			}
		}
		if config.Codex, err = filepath.EvalSymlinks(codex); err != nil || !filepath.IsAbs(config.Codex) {
			return nil, fmt.Errorf("cannot resolve the Codex executable %s", codex)
		}
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	h.meter = filepath.Join(bin, meterName)
	if err := copyExecutable(self, h.meter); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(bin, meterConfigName), append(encoded, '\n'), 0o644); err != nil {
		return nil, err
	}
	h.pathEnv = bin + string(os.PathListSeparator) + os.Getenv("PATH")
	h.codex = config.Codex
	if config.Fake {
		h.codex = "fake"
	}
	return h, ctx.Err()
}

// workDirectory creates the private directory for trial repositories outside
// the source repository, whose corpus holds the hidden checks. make points
// TMPDIR inside the repository, so that case falls back to /tmp.
func workDirectory(ctx context.Context) (string, error) {
	root, err := physicalDirectory(os.TempDir())
	if err != nil {
		return "", err
	}
	source := ""
	if current, err := os.Getwd(); err == nil {
		if top, err := git(ctx, current, "rev-parse", "--show-toplevel"); err == nil {
			source, _ = physicalDirectory(top)
		}
	}
	if source != "" && within(root, source) {
		if root, err = physicalDirectory("/tmp"); err != nil {
			return "", err
		}
	}
	created, err := os.MkdirTemp(root, "l7-outcome-eval-")
	if err != nil {
		return "", err
	}
	work, err := physicalDirectory(created)
	if err != nil {
		return "", err
	}
	if source != "" && within(work, source) {
		return "", fmt.Errorf("trial directory %s is inside the source repository", work)
	}
	return work, nil
}

// discoverRoute probes the metered Codex in a scratch repository and asks
// Level 7 which model and effort its first crew implementer would use. Both
// arms then run on that route for the whole run.
func (h *harness) discoverRoute(ctx context.Context, manifest *runManifest) error {
	repository := filepath.Join(h.work, "setup", "repo")
	if _, err := materializeRepository(ctx, repository, map[string][]byte{"README.md": []byte("# outcome-eval setup\n")}); err != nil {
		return err
	}
	if err := h.prepareCrew(ctx, repository); err != nil {
		return err
	}
	profile := h.corpus.protocol.RouteProfile
	common := []string{"--complexity", profile.Complexity, "--risk", fmt.Sprint(profile.Risk), "--context", fmt.Sprint(profile.Context), "--tools"}
	explain := func(arguments ...string) (envelope, route, error) {
		value, err := h.callL7(ctx, repository, time.Minute, append([]string{"route", "explain"}, arguments...)...)
		if err != nil {
			return value, route{}, err
		}
		var data struct {
			Decision struct {
				Provider string `json:"provider_id"`
				Model    string `json:"model_id"`
				Effort   string `json:"effort"`
			} `json:"decision"`
		}
		if json.Unmarshal(value.Data, &data) != nil || data.Decision.Model == "" {
			return value, route{}, errors.New("route explain returned no decision")
		}
		return value, route{Provider: data.Decision.Provider, Model: data.Decision.Model, Effort: data.Decision.Effort}, nil
	}
	implementer, chosen, err := explain(append(append([]string{"--task", "outcome-eval-implementer"}, common...), "--edit", "--resume")...)
	if err != nil {
		return err
	}
	_, reviewer, err := explain(append(append([]string{"--task", "outcome-eval-reviewer"}, common...), "--review", "--implementer-provider", chosen.Provider, "--implementer-model", chosen.Model)...)
	if err != nil {
		return err
	}
	chosen.Reviewer = reviewer.Provider + "/" + reviewer.Model + "@" + reviewer.Effort
	if chosen.Provider != "codex-local" {
		return fmt.Errorf("the crew's first implementer route is %s, not Codex", chosen.Provider)
	}
	if manifest.Route.Model != "" && manifest.Route != chosen {
		return fmt.Errorf("the route changed since the run started (%s/%s@%s, now %s/%s@%s); start a new run", manifest.Route.Provider, manifest.Route.Model, manifest.Route.Effort, chosen.Provider, chosen.Model, chosen.Effort)
	}
	h.route, manifest.Route = chosen, chosen
	digest, err := fileSHA256(h.l7)
	if err != nil {
		return err
	}
	if manifest.L7.SHA256 != "" && manifest.L7.SHA256 != digest {
		return errors.New("the l7 binary differs from the one this run started with; start a new run")
	}
	codex := binaryInfo{Path: h.codex, Version: h.codexVersion(ctx)}
	if manifest.Codex.Version != "" && manifest.Codex.Version != codex.Version {
		return fmt.Errorf("Codex changed since the run started (%s, now %s); start a new run", manifest.Codex.Version, codex.Version)
	}
	manifest.L7, manifest.Codex = binaryInfo{Path: h.l7, SHA256: digest, Version: implementer.Version}, codex
	manifest.Source = sourceRevision(ctx)
	manifest.Host = runtime.GOOS + "/" + runtime.GOARCH + " " + runtime.Version()
	manifest.Python = commandVersion(ctx, "python3", "--version")
	manifest.GradingSandbox = h.judge.sandbox
	manifest.UpdatedUTC = time.Now().UTC().Format(time.RFC3339)
	return nil
}

func (h *harness) codexVersion(ctx context.Context) string {
	result, err := execute(ctx, commandSpec{Dir: h.work, Env: h.environment(), Name: h.meter, Args: []string{"--version"}, Timeout: time.Minute})
	if err != nil || result.ExitCode != 0 {
		return "unknown"
	}
	return strings.TrimSpace(string(result.Stdout))
}

func commandVersion(ctx context.Context, name string, arguments ...string) string {
	result, err := execute(ctx, commandSpec{Env: baseEnvironment(os.Getenv("PATH")), Name: name, Args: arguments, Timeout: time.Minute})
	if err != nil || result.ExitCode != 0 {
		return "unknown"
	}
	return strings.TrimSpace(string(append(result.Stdout, result.Stderr...)))
}

func sourceRevision(ctx context.Context) string {
	directory, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	head, err := git(ctx, directory, "rev-parse", "HEAD")
	if err != nil {
		return "unknown"
	}
	if status, err := git(ctx, directory, "status", "--porcelain"); err != nil || status != "" {
		return head + "+dirty"
	}
	return head
}

func fileSHA256(name string) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyExecutable(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o755)
}

func saveManifest(directory string, manifest runManifest) error {
	manifest.UpdatedUTC = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(directory, "manifest.json"), append(data, '\n'))
}

func appendTrial(directory string, record trialRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(directory, "trials.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func loadRun(name string) (runManifest, []trialRecord, string, error) {
	directory, err := filepath.Abs(name)
	if err != nil {
		return runManifest{}, nil, "", err
	}
	var manifest runManifest
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil || json.Unmarshal(data, &manifest) != nil || manifest.Schema != 1 {
		return runManifest{}, nil, "", fmt.Errorf("cannot read %s", filepath.Join(directory, "manifest.json"))
	}
	records, err := readTrials(directory)
	return manifest, records, directory, err
}

func readTrials(directory string) ([]trialRecord, error) {
	data, err := os.ReadFile(filepath.Join(directory, "trials.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	records := []trialRecord{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var record trialRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("trials.jsonl line %d is malformed", len(records)+1)
		}
		records = append(records, record)
	}
	return records, scanner.Err()
}

func writeReport(directory string, report evaluationReport, records []trialRecord) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(directory, "report.json"), append(data, '\n')); err != nil {
		return err
	}
	return writeAtomic(filepath.Join(directory, "report.md"), []byte(renderMarkdown(report, records)))
}

// truthCheck proves a task's labels: its base fails the visible and hidden
// checks, and its reference passes both within the allowed paths.
func truthCheck(ctx context.Context, judge grader, value task, work string) error {
	base := judge.grade(ctx, value, value.repo, nil, filepath.Join(work, value.ID, "base"))
	if base.Visible.Error != "" || base.Hidden.Error != "" {
		return fmt.Errorf("base checks could not run: %s%s", base.Visible.Error, base.Hidden.Error)
	}
	if base.Visible.Passed || base.Hidden.Passed {
		return fmt.Errorf("base passes its checks (visible %t, hidden %t)", base.Visible.Passed, base.Hidden.Passed)
	}
	solved := map[string][]byte{}
	for relative, content := range value.repo {
		solved[relative] = content
	}
	for relative, content := range value.referenceFiles(judge.rules.Canary) {
		solved[relative] = content
	}
	reference := judge.grade(ctx, value, solved, nil, filepath.Join(work, value.ID, "reference"))
	switch {
	case !reference.Visible.Passed || !reference.Hidden.Passed:
		return fmt.Errorf("reference fails its checks: visible %t %s; hidden %t %s", reference.Visible.Passed, reference.Visible.Tail+reference.Visible.Error, reference.Hidden.Passed, reference.Hidden.Tail+reference.Hidden.Error)
	case reference.safetyViolation() || reference.Contaminated:
		return fmt.Errorf("reference violates scope or forbidden effects: %+v", reference)
	case len(reference.ChangedPaths) == 0 || !slices.Equal(reference.ChangedPaths, sortedKeys(value.reference)):
		return fmt.Errorf("reference changes %v, not exactly its files", reference.ChangedPaths)
	}
	return nil
}
