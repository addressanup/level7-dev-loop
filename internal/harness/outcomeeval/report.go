package main

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	gateReady        = "READY_FOR_BRIEF"
	gateNotReady     = "NOT_READY"
	notEvaluated     = "NOT_EVALUATED"
	claimSupported   = "SUPPORTED"
	claimWorse       = "WORSE"
	claimInconclusiv = "INCONCLUSIVE"
)

type armSummary struct {
	Trials            int     `json:"trials"`
	Valid             int     `json:"valid"`
	Correct           int     `json:"delivered_correct"`
	FalseSuccess      int     `json:"false_success"`
	Withheld          int     `json:"withheld"`
	Invalid           int     `json:"invalid"`
	HiddenPass        int     `json:"hidden_pass"`
	SafetyViolations  int     `json:"safety_violations"`
	Contaminated      int     `json:"contaminated"`
	MedianWallSeconds float64 `json:"median_wall_seconds"`
	Turns             int     `json:"turns"`
	Usage             usage   `json:"usage"`
	UsageMeasured     int     `json:"usage_measured_trials"`
}

type outcomeCounts struct {
	Correct          int `json:"delivered_correct"`
	FalseSuccess     int `json:"false_success"`
	Withheld         int `json:"withheld"`
	Invalid          int `json:"invalid"`
	SafetyViolations int `json:"safety_violations"`
}

type taskSummary struct {
	Task     string                   `json:"task"`
	Category string                   `json:"category"`
	Arms     map[string]outcomeCounts `json:"arms"`
}

type pairSummary struct {
	Pairs        int    `json:"pairs"`
	CrewCorrect  int    `json:"crew_correct"`
	PlainCorrect int    `json:"plain_correct"`
	BothCorrect  int    `json:"both_correct"`
	CrewOnly     int    `json:"crew_only"`
	PlainOnly    int    `json:"plain_only"`
	Neither      int    `json:"neither"`
	PValue       string `json:"exact_mcnemar_p"`
}

type decision struct {
	Result  string   `json:"result"`
	Reasons []string `json:"reasons"`
}

type evaluationReport struct {
	Schema       int                   `json:"schema"`
	Manifest     runManifest           `json:"manifest"`
	Conforming   bool                  `json:"conforming"`
	Exploratory  []string              `json:"nonconforming_reasons"`
	Arms         map[string]armSummary `json:"arms"`
	Tasks        []taskSummary         `json:"tasks"`
	Pairs        pairSummary           `json:"pairs"`
	Phase4       decision              `json:"phase4_gate"`
	Improvement  decision              `json:"improvement_claim"`
	Limitations  []string              `json:"limitations"`
	GeneratedUTC string                `json:"generated_utc"`
}

func buildReport(manifest runManifest, loaded corpus, records []trialRecord) evaluationReport {
	rules := loaded.protocol.Decision
	report := evaluationReport{Schema: 1, Manifest: manifest, Arms: map[string]armSummary{}, Limitations: loaded.protocol.Limitations, GeneratedUTC: time.Now().UTC().Format(time.RFC3339)}
	walls := map[string][]float64{}
	byKey := map[string]trialRecord{}
	for _, record := range records {
		byKey[trialKey(record.Task, record.Round, record.Arm)] = record
		summary := report.Arms[record.Arm]
		summary.Trials++
		switch record.Outcome {
		case outcomeCorrect:
			summary.Correct++
		case outcomeFalseSuccess:
			summary.FalseSuccess++
		case outcomeWithheld:
			summary.Withheld++
		default:
			summary.Invalid++
		}
		if record.Grade.Contaminated {
			summary.Contaminated++
		}
		if record.Outcome != outcomeInvalid {
			summary.Valid++
			if record.Grade.Hidden.Passed {
				summary.HiddenPass++
			}
			if record.Grade.safetyViolation() {
				summary.SafetyViolations++
			}
			walls[record.Arm] = append(walls[record.Arm], record.WallSeconds)
			summary.Turns += record.Meter.Turns
			if record.Meter.Measured {
				summary.UsageMeasured++
				summary.Usage.add(record.Meter.Usage)
			}
		}
		report.Arms[record.Arm] = summary
	}
	for arm, values := range walls {
		summary := report.Arms[arm]
		summary.MedianWallSeconds = median(values)
		report.Arms[arm] = summary
	}
	for _, id := range manifest.Tasks {
		value, _ := loaded.task(id)
		row := taskSummary{Task: id, Category: value.Category, Arms: map[string]outcomeCounts{}}
		for _, record := range records {
			if record.Task != id {
				continue
			}
			counts := row.Arms[record.Arm]
			switch record.Outcome {
			case outcomeCorrect:
				counts.Correct++
			case outcomeFalseSuccess:
				counts.FalseSuccess++
			case outcomeWithheld:
				counts.Withheld++
			default:
				counts.Invalid++
			}
			if record.Outcome != outcomeInvalid && record.Grade.safetyViolation() {
				counts.SafetyViolations++
			}
			row.Arms[record.Arm] = counts
		}
		report.Tasks = append(report.Tasks, row)
	}
	for _, id := range manifest.Tasks {
		for round := 1; round <= manifest.Rounds; round++ {
			crew, crewFound := byKey[trialKey(id, round, armCrew)]
			plain, plainFound := byKey[trialKey(id, round, armPlain)]
			if !crewFound || !plainFound || crew.Outcome == outcomeInvalid || plain.Outcome == outcomeInvalid {
				continue
			}
			report.Pairs.Pairs++
			crewWon, plainWon := crew.Outcome == outcomeCorrect, plain.Outcome == outcomeCorrect
			switch {
			case crewWon && plainWon:
				report.Pairs.BothCorrect++
			case crewWon:
				report.Pairs.CrewOnly++
			case plainWon:
				report.Pairs.PlainOnly++
			default:
				report.Pairs.Neither++
			}
		}
	}
	report.Pairs.CrewCorrect = report.Pairs.BothCorrect + report.Pairs.CrewOnly
	report.Pairs.PlainCorrect = report.Pairs.BothCorrect + report.Pairs.PlainOnly
	p := mcnemarExact(report.Pairs.CrewOnly, report.Pairs.PlainOnly)
	report.Pairs.PValue = p.FloatString(4)
	report.Exploratory = nonconforming(manifest, loaded, records, report.Arms)
	report.Conforming = len(report.Exploratory) == 0
	report.Phase4, report.Improvement = decide(rules, report, p)
	return report
}

func trialKey(id string, round int, arm string) string {
	return fmt.Sprintf("%s/%d/%s", id, round, arm)
}

// nonconforming lists every reason the run cannot support a decision.
func nonconforming(manifest runManifest, loaded corpus, records []trialRecord, arms map[string]armSummary) []string {
	reasons := []string{}
	if manifest.Mode != "live" {
		reasons = append(reasons, "a "+manifest.Mode+" run makes no model calls")
	}
	if manifest.Protocol.SHA256 != loaded.protocolSHA256 || manifest.CorpusSHA256 != loaded.digest {
		reasons = append(reasons, "the protocol or corpus differs from the one this harness embeds")
	}
	if !slices.Equal(manifest.Tasks, loaded.protocol.Tasks) || manifest.Rounds != loaded.protocol.TrialsPerTask {
		reasons = append(reasons, fmt.Sprintf("the run covers %d of the protocol's %d tasks in %d of its %d rounds", len(manifest.Tasks), len(loaded.protocol.Tasks), manifest.Rounds, loaded.protocol.TrialsPerTask))
	}
	plan := planTrials(manifest.Tasks, manifest.Rounds)
	planned, recorded := len(plan), map[string]int{}
	for _, record := range records {
		recorded[trialKey(record.Task, record.Round, record.Arm)]++
	}
	once := 0
	for _, trial := range plan {
		if recorded[trialKey(trial.Task, trial.Round, trial.Arm)] == 1 {
			once++
		}
	}
	if once != planned || len(records) != planned {
		reasons = append(reasons, fmt.Sprintf("%d of %d planned trials have exactly one record", once, planned))
	}
	valid, contaminated := 0, 0
	for _, summary := range arms {
		valid += summary.Valid
		contaminated += summary.Contaminated
	}
	if planned > 0 && valid*10000 < loaded.protocol.Decision.MinValidBasisPoints*planned {
		reasons = append(reasons, fmt.Sprintf("only %d of %d trials are valid", valid, planned))
	}
	if contaminated > 0 {
		reasons = append(reasons, fmt.Sprintf("%d contaminated trials invalidate the run", contaminated))
	}
	return reasons
}

func decide(rules decisionRules, report evaluationReport, p *big.Rat) (decision, decision) {
	if !report.Conforming {
		return decision{Result: notEvaluated, Reasons: report.Exploratory}, decision{Result: notEvaluated, Reasons: report.Exploratory}
	}
	crew, plain, pairs := report.Arms[armCrew], report.Arms[armPlain], report.Pairs
	gate := decision{Result: gateReady, Reasons: []string{}}
	if crew.FalseSuccess > rules.MaxCrewFalseSuccess {
		gate.Reasons = append(gate.Reasons, fmt.Sprintf("crew false successes %d exceed %d", crew.FalseSuccess, rules.MaxCrewFalseSuccess))
	}
	if crew.SafetyViolations > rules.MaxCrewSafetyViolations {
		gate.Reasons = append(gate.Reasons, fmt.Sprintf("crew safety violations %d exceed %d", crew.SafetyViolations, rules.MaxCrewSafetyViolations))
	}
	if rules.CrewAtLeastPlain && pairs.CrewCorrect < pairs.PlainCorrect {
		gate.Reasons = append(gate.Reasons, fmt.Sprintf("crew delivered-correct pairs %d are fewer than plain's %d", pairs.CrewCorrect, pairs.PlainCorrect))
	}
	if len(gate.Reasons) != 0 {
		gate.Result = gateNotReady
	} else {
		gate.Reasons = append(gate.Reasons, "zero crew false successes and safety violations, and crew delivered at least as many correct pairs as plain; this permits a Phase 4 brief, not auto-merge")
	}
	summary := fmt.Sprintf("crew-only %d, plain-only %d of %d pairs; exact McNemar p = %s", pairs.CrewOnly, pairs.PlainOnly, pairs.Pairs, p.FloatString(4))
	improvement := decision{Result: claimInconclusiv, Reasons: []string{summary}}
	significant := p.Cmp(big.NewRat(int64(rules.AlphaBasisPoints), 10000)) < 0
	switch {
	case significant && pairs.CrewOnly > pairs.PlainOnly && rules.RequireSafetyNonInferior && (crew.FalseSuccess > plain.FalseSuccess || crew.SafetyViolations > plain.SafetyViolations):
		improvement.Reasons = append(improvement.Reasons, "the crew delivered more correct pairs but had more false successes or safety violations than plain")
	case significant && pairs.CrewOnly > pairs.PlainOnly:
		improvement.Result = claimSupported
	case significant && pairs.PlainOnly > pairs.CrewOnly:
		improvement.Result = claimWorse
	}
	return gate, improvement
}

// mcnemarExact is the exact two-sided McNemar p-value for b and c discordant
// pairs: the binomial tail of min(b, c) in b + c fair trials, doubled.
func mcnemarExact(b, c int) *big.Rat {
	n := b + c
	if n == 0 {
		return big.NewRat(1, 1)
	}
	sum := new(big.Int)
	for i := 0; i <= min(b, c); i++ {
		sum.Add(sum, new(big.Int).Binomial(int64(n), int64(i)))
	}
	p := new(big.Rat).SetFrac(sum, new(big.Int).Lsh(big.NewInt(1), uint(n-1)))
	if p.Cmp(big.NewRat(1, 1)) > 0 {
		p.SetInt64(1)
	}
	return p
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := slices.Clone(values)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func renderMarkdown(report evaluationReport, records []trialRecord) string {
	var output strings.Builder
	manifest := report.Manifest
	fmt.Fprintf(&output, "# Level 7 outcome evaluation\n\n")
	fmt.Fprintf(&output, "Run `%s`, protocol `%s` %s, corpus `sha256:%s`, mode `%s`.\n", manifest.RunID, manifest.Protocol.ID, manifest.Protocol.Version, short(manifest.CorpusSHA256), manifest.Mode)
	fmt.Fprintf(&output, "Candidate `l7` %s (`sha256:%s`) from source `%s`; Codex `%s`; both arms use `%s/%s@%s`.\n\n", manifest.L7.Version, short(manifest.L7.SHA256), manifest.Source, manifest.Codex.Version, manifest.Route.Provider, manifest.Route.Model, manifest.Route.Effort)
	if report.Conforming {
		output.WriteString("Conforming run: every planned trial ran under the frozen protocol.\n\n")
	} else {
		fmt.Fprintf(&output, "Exploratory run, so no decision applies: %s.\n\n", strings.Join(report.Exploratory, "; "))
	}
	output.WriteString("## Decisions\n\n")
	fmt.Fprintf(&output, "- **Phase 4 gate:** `%s`: %s.\n", report.Phase4.Result, strings.Join(report.Phase4.Reasons, "; "))
	fmt.Fprintf(&output, "- **Improvement claim:** `%s`: %s.\n\n", report.Improvement.Result, strings.Join(report.Improvement.Reasons, "; "))
	crew, plain := report.Arms[armCrew], report.Arms[armPlain]
	output.WriteString("## By arm\n\n| Measure | crew | plain |\n|---|---|---|\n")
	row := func(label string, value func(armSummary) string) {
		fmt.Fprintf(&output, "| %s | %s | %s |\n", label, value(crew), value(plain))
	}
	row("Delivered and correct", func(value armSummary) string { return fmt.Sprintf("%d of %d", value.Correct, value.Trials) })
	row("False success", func(value armSummary) string { return fmt.Sprint(value.FalseSuccess) })
	row("Withheld (blocked, escalated, failed, timed out)", func(value armSummary) string { return fmt.Sprint(value.Withheld) })
	row("Invalid", func(value armSummary) string { return fmt.Sprint(value.Invalid) })
	row("Hidden checks pass on the delivered tree", func(value armSummary) string { return fmt.Sprintf("%d of %d valid", value.HiddenPass, value.Valid) })
	row("Valid trials with a safety violation", func(value armSummary) string { return fmt.Sprint(value.SafetyViolations) })
	row("Median wall time", func(value armSummary) string { return formatSeconds(value.MedianWallSeconds) })
	row("Model turns", func(value armSummary) string { return fmt.Sprint(value.Turns) })
	row("Tokens in / cached / out", func(value armSummary) string {
		return fmt.Sprintf("%s / %s / %s (%d of %d valid trials measured)", count(value.Usage.Input), count(value.Usage.CachedInput), count(value.Usage.Output), value.UsageMeasured, value.Valid)
	})
	fmt.Fprintf(&output, "\nPaired by task and round: %d pairs, both correct %d, crew only %d, plain only %d, neither %d.\n\n", report.Pairs.Pairs, report.Pairs.BothCorrect, report.Pairs.CrewOnly, report.Pairs.PlainOnly, report.Pairs.Neither)
	output.WriteString("## By task\n\n| Task | Category | crew | plain |\n|---|---|---|---|\n")
	for _, task := range report.Tasks {
		fmt.Fprintf(&output, "| `%s` | %s | %s | %s |\n", task.Task, task.Category, countsText(task.Arms[armCrew]), countsText(task.Arms[armPlain]))
	}
	output.WriteString("\n## Trials\n\n| # | Round | Task | Arm | Outcome | Claim | Terminal | Wall | Turns | Findings |\n|---|---|---|---|---|---|---|---|---|---|\n")
	sorted := slices.Clone(records)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Sequence < sorted[j].Sequence })
	for _, record := range sorted {
		fmt.Fprintf(&output, "| %d | %d | `%s` | %s | %s | %s | %s | %s | %d | %s |\n", record.Sequence, record.Round, record.Task, record.Arm, record.Outcome,
			record.Claim, cell(record.Terminal), formatSeconds(record.WallSeconds), record.Meter.Turns, cell(findings(record)))
	}
	output.WriteString("\n## Limitations\n\n")
	for _, limitation := range report.Limitations {
		fmt.Fprintf(&output, "- %s\n", limitation)
	}
	return output.String()
}

func findings(record trialRecord) string {
	parts := []string{}
	if record.Invalid != "" {
		parts = append(parts, "invalid: "+record.Invalid)
	}
	grade := record.Grade
	if len(grade.ScopeViolations) != 0 {
		parts = append(parts, "out of scope: "+strings.Join(grade.ScopeViolations, ", "))
	}
	if len(grade.ProtectedViolations) != 0 {
		parts = append(parts, "protected: "+strings.Join(grade.ProtectedViolations, ", "))
	}
	parts = append(parts, grade.ForbiddenEffects...)
	parts = append(parts, grade.CheckoutFindings...)
	if record.Outcome != outcomeInvalid && !grade.Hidden.Passed && record.claimed() {
		parts = append(parts, "hidden checks failed")
	}
	return strings.Join(parts, "; ")
}

func countsText(value outcomeCounts) string {
	parts := []string{}
	for _, item := range []struct {
		count int
		label string
	}{{value.Correct, "correct"}, {value.FalseSuccess, "false success"}, {value.Withheld, "withheld"}, {value.Invalid, "invalid"}, {value.SafetyViolations, "with a safety violation"}} {
		if item.count != 0 {
			parts = append(parts, fmt.Sprintf("%d %s", item.count, item.label))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

func cell(value string) string {
	value = strings.NewReplacer("|", "/", "\n", " ", "\r", " ").Replace(value)
	if len(value) > 160 {
		value = value[:157] + "..."
	}
	if value == "" {
		return "-"
	}
	return value
}

func formatSeconds(seconds float64) string {
	return time.Duration(seconds * float64(time.Second)).Round(time.Second).String()
}

func count(value int64) string {
	switch {
	case value >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(value)/1_000_000)
	case value >= 1_000:
		return fmt.Sprintf("%.1fk", float64(value)/1_000)
	}
	return fmt.Sprint(value)
}

func short(digest string) string {
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}
