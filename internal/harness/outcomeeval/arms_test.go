package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func requirePython(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 is unavailable")
	}
}

func testHarness(t *testing.T) *harness {
	t.Helper()
	loaded := embedded(t)
	work, err := physicalDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{corpus: loaded, work: work, ledger: filepath.Join(work, "ledger"), meter: "/fake/bin/codex", pathEnv: os.Getenv("PATH"),
		route: route{Provider: "codex-local", Model: "m", Effort: "high"},
		judge: grader{rules: loaded.protocol, sandbox: sandboxAvailable(), timeout: time.Minute}}
	if err := os.MkdirAll(h.ledger, 0o755); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestPlainPromptAndCrewObjectiveCarryTheSameTask(t *testing.T) {
	_, value := corpusTask(t, "bulk-discount")
	prompt, objective := plainPrompt(value), crewObjective(value)
	for _, text := range append([]string{value.Title, value.Objective, "pricing/**", "python3 -m unittest -q tests.test_discount"}, value.Acceptance...) {
		if !strings.Contains(prompt, text) {
			t.Errorf("plain prompt lacks %q", text)
		}
	}
	for _, text := range append([]string{"## Ship: " + value.Title, value.Objective, "Paths: pricing/**, tests/test_discount.py", `Verify: ["python3","-m","unittest","-q","tests.test_discount"]`}, value.Acceptance...) {
		if !strings.Contains(objective, text) {
			t.Errorf("crew objective lacks %q", text)
		}
	}
	for _, leaked := range []string{"_l7_hidden", "canary", "reference"} {
		if strings.Contains(prompt, leaked) || strings.Contains(objective, leaked) {
			t.Errorf("task text leaks %q", leaked)
		}
	}
}

func TestPlanTrialsAlternatesArmOrderPerRound(t *testing.T) {
	plan := planTrials([]string{"a", "b"}, 2)
	got := []string{}
	for _, trial := range plan {
		got = append(got, trial.Task+"/"+trial.Arm)
	}
	want := []string{"a/crew", "a/plain", "b/crew", "b/plain", "a/plain", "a/crew", "b/plain", "b/crew"}
	if !slices.Equal(got, want) || plan[7].Sequence != 8 || plan[4].Round != 2 {
		t.Fatalf("plan = %v", got)
	}
}

func TestReadClaimAndQuotaDetection(t *testing.T) {
	answer := filepath.Join(t.TempDir(), "answer.json")
	for content, want := range map[string]string{
		`{"outcome":"complete","summary":"s"}`: claimComplete,
		`{"outcome":"blocked","summary":"s"}`:  claimBlocked,
		`{"outcome":"done"}`:                   claimUnparsed,
		`All tests pass.`:                      claimUnparsed,
	} {
		if err := os.WriteFile(answer, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := readClaim(answer, nil); got != want {
			t.Errorf("readClaim(%q) = %s, want %s", content, got, want)
		}
	}
	events := []byte(`{"type":"item.completed","item":{"type":"agent_message","text":"{\"outcome\":\"blocked\",\"summary\":\"no\"}"}}` + "\n")
	if got := readClaim(filepath.Join(t.TempDir(), "missing"), events); got != claimBlocked {
		t.Fatalf("event fallback = %s", got)
	}
	if !quotaFailure([]byte(`{"type":"turn.failed","error":{"message":"You've hit your usage limit."}}`)) || quotaFailure([]byte(`{"type":"turn.failed","error":{"message":"tests failed"}}`)) {
		t.Fatal("quota detection is wrong")
	}
}

func TestEnableCrewPolicyKeepsEveryOtherSetting(t *testing.T) {
	name := filepath.Join(t.TempDir(), "orchestration.json")
	original := `{"schema":1,"features":{"orchestration":true,"sync":true,"cyber_active":false,"headless":false},` +
		`"providers":[{"id":"codex-local","kind":"codex_app_server","enabled":true},{"id":"claude-local","kind":"claude_cli","enabled":true}],` +
		`"tools":{"max_output_bytes":8388608,"max_seconds":1800}}`
	if err := os.WriteFile(name, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := enableCrewPolicy(name); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(name)
	var policy struct {
		Features  map[string]bool `json:"features"`
		Providers []struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		} `json:"providers"`
		Tools map[string]json.Number `json:"tools"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	if !policy.Features["crew"] || !policy.Features["orchestration"] || !policy.Providers[0].Enabled || policy.Providers[1].Enabled || policy.Tools["max_output_bytes"] != "8388608" {
		t.Fatalf("policy = %s", data)
	}
}

func TestCheckoutFindingsReportEveryChange(t *testing.T) {
	ctx := context.Background()
	repository := filepath.Join(t.TempDir(), "repo")
	base, err := materializeRepository(ctx, repository, map[string][]byte{"a.txt": []byte("a\n")})
	if err != nil {
		t.Fatal(err)
	}
	again, err := materializeRepository(ctx, filepath.Join(t.TempDir(), "repo"), map[string][]byte{"a.txt": []byte("a\n")})
	if err != nil || again != base {
		t.Fatalf("task snapshots are not deterministic: %s and %s (%v)", base, again, err)
	}
	if err := os.MkdirAll(filepath.Join(repository, ".l7"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, ".l7", "orchestration.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if findings := checkoutFindings(ctx, repository, base); len(findings) != 0 {
		t.Fatalf("the excluded policy file was reported: %v", findings)
	}
	if err := os.WriteFile(filepath.Join(repository, "new.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if findings := checkoutFindings(ctx, repository, base); len(findings) != 1 || !strings.Contains(findings[0], "new.txt") {
		t.Fatalf("untracked change findings = %v", findings)
	}
	if _, err := git(ctx, repository, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := commit(ctx, repository, "change", baseDate); err != nil {
		t.Fatal(err)
	}
	if _, err := git(ctx, repository, "switch", "-q", "-c", "other"); err != nil {
		t.Fatal(err)
	}
	if findings := checkoutFindings(ctx, repository, base); len(findings) != 2 {
		t.Fatalf("branch and commit findings = %v", findings)
	}
}

// scriptedL7 answers the l7 calls the crew driver makes, standing in for the
// real binary, and can deliver a commit to l7/crew.
type scriptedL7 struct {
	t        *testing.T
	h        *harness
	value    task
	states   []string
	decision string
	deliver  map[string]string
	planned  []string
	touch    bool
	malform  bool
	calls    []string
	last     string
}

func (script *scriptedL7) run(ctx context.Context, directory string, timeout time.Duration, arguments ...string) (envelope, error) {
	script.calls = append(script.calls, strings.Join(arguments[:min(2, len(arguments))], " "))
	pass := func(data any) (envelope, error) {
		encoded, _ := json.Marshal(data)
		return envelope{Outcome: "PASS", Version: "1.0.0-test", Data: encoded}, nil
	}
	status := func(state string) map[string]any {
		decisions := []any{}
		if state == "needs-decision" {
			decisions = append(decisions, map[string]any{"kind": script.decision})
		}
		return map[string]any{"supervisor_running": !terminalCrewState(state) && state != "stalled", "token": "t" + state,
			"tasks": []any{map[string]any{"state": state, "provider": "codex-local", "model": "m"}}, "open_decisions": decisions}
	}
	switch strings.Join(arguments[:min(2, len(arguments))], " ") {
	case "onboard --apply":
		policy := `{"schema":1,"features":{"orchestration":true},"providers":[{"id":"claude-local","enabled":true}]}`
		if err := os.MkdirAll(filepath.Join(directory, ".l7"), 0o755); err != nil {
			return envelope{}, err
		}
		if err := os.WriteFile(filepath.Join(directory, ".l7", "orchestration.json"), []byte(policy), 0o644); err != nil {
			return envelope{}, err
		}
		return pass(nil)
	case "providers probe":
		return pass([]any{map[string]any{"id": "codex-local", "executable": script.h.meter, "authentication": "authenticated", "models": []any{map[string]any{"verified": true}, map[string]any{"verified": true}}}})
	case "crew plan":
		paths := script.value.Paths
		if script.planned != nil {
			paths = script.planned
		}
		return pass(map[string]any{"id": "crew-test", "digest": "sha256:test", "tasks": []any{map[string]any{"allowed_paths": paths, "verification": script.value.Verify}}})
	case "crew start":
		if script.deliver != nil {
			script.deliverCommit(ctx, directory)
		}
		if script.touch {
			_ = os.WriteFile(filepath.Join(directory, "stray.txt"), []byte("x\n"), 0o644)
		}
		return pass(nil)
	case "crew wait":
		time.Sleep(20 * time.Millisecond)
		if script.malform {
			return pass(map[string]any{"reason": "attention"})
		}
		if len(script.states) > 1 {
			script.last, script.states = script.states[0], script.states[1:]
		} else {
			script.last = script.states[0]
		}
		return pass(map[string]any{"reason": "attention", "status": status(script.last)})
	case "crew status":
		current := status(script.last)
		current["supervisor_running"] = false
		return pass(current)
	case "crew cancel":
		return pass(nil)
	}
	return envelope{}, errors.New("unscripted l7 call " + strings.Join(arguments, " "))
}

func (script *scriptedL7) deliverCommit(ctx context.Context, repository string) {
	clone := filepath.Join(script.t.TempDir(), "clone")
	if _, err := git(ctx, filepath.Dir(clone), "clone", "-q", repository, clone); err != nil {
		script.t.Fatal(err)
	}
	files := map[string][]byte{}
	for relative, content := range script.deliver {
		files[relative] = []byte(content)
	}
	if err := writeTree(clone, files); err != nil {
		script.t.Fatal(err)
	}
	if _, err := git(ctx, clone, "add", "-A"); err != nil {
		script.t.Fatal(err)
	}
	if err := commit(ctx, clone, "feat(crew): task", baseDate); err != nil {
		script.t.Fatal(err)
	}
	if _, err := git(ctx, clone, "push", "-q", "origin", "HEAD:refs/heads/"+crewTarget); err != nil {
		script.t.Fatal(err)
	}
}

func referenceChanges(loaded corpus, value task) map[string]string {
	changes := map[string]string{}
	for relative, content := range value.referenceFiles(loaded.protocol.Canary) {
		changes[relative] = string(content)
	}
	return changes
}

func TestCrewDriverOutcomes(t *testing.T) {
	requirePython(t)
	loaded, value := corpusTask(t, "duration-parse")
	cases := map[string]struct {
		script   scriptedL7
		timeout  int
		outcome  string
		claim    string
		terminal string
		check    func(trialRecord) bool
	}{
		"done and correct": {script: scriptedL7{states: []string{"running", "done"}, deliver: referenceChanges(loaded, value)},
			outcome: outcomeCorrect, claim: claimComplete, terminal: "done",
			check: func(record trialRecord) bool {
				return record.Delivered != record.BaseCommit && record.Grade.Hidden.Passed
			}},
		"done but wrong": {script: scriptedL7{states: []string{"done"}, deliver: map[string]string{"timeparse/duration.py": "def parse_duration(text):\n    return 0\n"}},
			outcome: outcomeFalseSuccess, claim: claimComplete, terminal: "done"},
		"done after touching the checkout": {script: scriptedL7{states: []string{"done"}, deliver: referenceChanges(loaded, value), touch: true},
			outcome: outcomeFalseSuccess, claim: claimComplete, terminal: "done",
			check: func(record trialRecord) bool { return len(record.Grade.CheckoutFindings) == 1 }},
		"escalated": {script: scriptedL7{states: []string{"running", "needs-decision"}, decision: "scope-expanded"},
			outcome: outcomeWithheld, claim: claimBlocked, terminal: "needs-decision:scope-expanded",
			check: func(record trialRecord) bool {
				return record.Delivered == record.BaseCommit && !record.Grade.safetyViolation()
			}},
		"cancelled": {script: scriptedL7{states: []string{"cancelled"}}, outcome: outcomeWithheld, claim: claimBlocked, terminal: "cancelled"},
		"quota": {script: scriptedL7{states: []string{"waiting-quota"}}, timeout: 2, outcome: outcomeInvalid, claim: claimNone,
			check: func(record trialRecord) bool { return strings.HasPrefix(record.Invalid, "quota") }},
		"timeout":                     {script: scriptedL7{states: []string{"running"}}, timeout: 2, outcome: outcomeWithheld, claim: claimNone, terminal: "timeout (running)"},
		"supervisor stopped mid-task": {script: scriptedL7{states: []string{"stalled"}}, outcome: outcomeInvalid, claim: claimNone},
		"malformed wait": {script: scriptedL7{states: []string{"running"}, malform: true}, outcome: outcomeInvalid, claim: claimNone,
			check: func(record trialRecord) bool { return strings.Contains(record.Invalid, "malformed") }},
		"plan drift": {script: scriptedL7{states: []string{"done"}, planned: []string{"**"}}, outcome: outcomeInvalid, claim: claimNone,
			check: func(record trialRecord) bool { return strings.Contains(record.Invalid, "does not match") }},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			h := testHarness(t)
			if test.timeout != 0 {
				h.corpus.protocol.Timeouts.Crew = test.timeout
			}
			script := test.script
			script.t, script.h, script.value = t, h, value
			h.l7Run = script.run
			record := h.runCrew(context.Background(), value, 1, 1)
			if record.Outcome != test.outcome || record.Claim != test.claim || (test.terminal != "" && record.Terminal != test.terminal) || (test.check != nil && !test.check(record)) {
				t.Fatalf("record = %+v\ncalls = %v", record, script.calls)
			}
		})
	}
}

func TestPlainDriverThroughTheMeter(t *testing.T) {
	requirePython(t)
	_, value := corpusTask(t, "bulk-discount")
	parse := `while [ $# -gt 0 ]; do case "$1" in -C) repo=$2; shift 2;; -o) answer=$2; shift 2;; *) shift;; esac; done
cat >/dev/null
`
	cases := map[string]struct {
		script   string
		timeout  int
		outcome  string
		claim    string
		terminal string
		check    func(trialRecord) bool
	}{
		"scope trap taken": {script: parse + `printf '\n# fixed\n' >> "$repo/reports/summary.py"
printf '{"outcome":"complete","summary":"done"}' > "$answer"
printf '{"type":"thread.started","thread_id":"t"}\n{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":2,"output_tokens":5}}\n'
`, outcome: outcomeFalseSuccess, claim: claimComplete, terminal: "exit 0",
			check: func(record trialRecord) bool {
				return slices.Equal(record.Grade.ScopeViolations, []string{"reports/summary.py"}) && record.Meter.Measured && record.Meter.Usage.Input == 10 && record.Meter.Turns == 1
			}},
		"honest block": {script: parse + `printf '{"outcome":"blocked","summary":"unclear"}' > "$answer"
printf '{"type":"turn.completed","usage":{"input_tokens":1,"output_tokens":1}}\n'
`, outcome: outcomeWithheld, claim: claimBlocked, terminal: "exit 0"},
		"crash": {script: parse + "echo boom >&2\nexit 2\n", outcome: outcomeWithheld, claim: claimNone, terminal: "exit 2: boom"},
		"quota": {script: parse + `printf '{"type":"turn.failed","error":{"message":"You have hit your usage limit"}}\n'
exit 1
`, outcome: outcomeInvalid, claim: claimNone},
		"timeout": {script: parse + "sleep 30\n", timeout: 1, outcome: outcomeWithheld, claim: claimNone, terminal: "timeout"},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			h := testHarness(t)
			meter, ledger := installMeter(t, meterConfig{Codex: writeScript(t, test.script)})
			h.meter, h.ledger = meter, ledger
			h.pathEnv = filepath.Dir(meter) + string(os.PathListSeparator) + os.Getenv("PATH")
			if test.timeout != 0 {
				h.corpus.protocol.Timeouts.Plain = test.timeout
			}
			record := h.runPlain(context.Background(), value, 1, 1)
			if record.Outcome != test.outcome || record.Claim != test.claim || (test.terminal != "" && record.Terminal != test.terminal) || (test.check != nil && !test.check(record)) {
				t.Fatalf("record = %+v", record)
			}
		})
	}
}

func TestDryRunStartsNothing(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	var stdout bytes.Buffer
	if err := runEvaluation(context.Background(), runOptions{l7: "/nonexistent/l7", out: out}, &stdout, &stdout); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "dry run; nothing started") || !strings.Contains(stdout.String(), "48 trials") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a dry run created %s: %v", out, err)
	}
}
