package domain

import "strings"

// CrewSchema versions crew plans, checkpoints, and decisions independently
// from Headless records.
const CrewSchema = 1

const (
	CrewMaxTasks        = 32
	CrewMaxWorkers      = 4
	CrewMaxRepairRounds = 4
)

type CrewShape string

const (
	CrewShip  CrewShape = "ship"
	CrewScout CrewShape = "scout"
)

func (shape CrewShape) Valid() bool { return shape == CrewShip || shape == CrewScout }

type CrewState string

const (
	CrewQueued        CrewState = "queued"
	CrewRunning       CrewState = "running"
	CrewMerging       CrewState = "merging"
	CrewWaitingQuota  CrewState = "waiting-quota"
	CrewNeedsDecision CrewState = "needs-decision"
	CrewPaused        CrewState = "paused"
	CrewDone          CrewState = "done"
	CrewCancelled     CrewState = "cancelled"
)

func (state CrewState) Valid() bool {
	switch state {
	case CrewQueued, CrewRunning, CrewMerging, CrewWaitingQuota, CrewNeedsDecision, CrewPaused, CrewDone, CrewCancelled:
		return true
	default:
		return false
	}
}

// Active states occupy a worker slot and the task's path scope.
func (state CrewState) Active() bool {
	return state == CrewRunning || state == CrewMerging || state == CrewWaitingQuota
}

func (state CrewState) Terminal() bool { return state == CrewDone || state == CrewCancelled }

// NeedsAttention marks the states that end a liaison wait.
func (state CrewState) NeedsAttention() bool {
	return state == CrewNeedsDecision || state == CrewPaused || state == CrewDone
}

type CrewTask struct {
	ID                 string     `json:"id"`
	Shape              CrewShape  `json:"shape"`
	Title              string     `json:"title"`
	Objective          string     `json:"objective"`
	AcceptanceCriteria []string   `json:"acceptance_criteria"`
	AllowedPaths       []string   `json:"allowed_paths"`
	Verification       [][]string `json:"verification"`
}

type CrewPlan struct {
	Schema          int        `json:"schema"`
	ID              string     `json:"id"`
	ObjectivePath   string     `json:"objective_path"`
	ObjectiveDigest string     `json:"objective_digest"`
	BaseCommit      string     `json:"base_commit"`
	TargetBranch    string     `json:"target_branch"`
	RiskCeiling     RiskTier   `json:"risk_ceiling"`
	MaxWorkers      int        `json:"max_workers"`
	RepairRounds    int        `json:"repair_rounds"`
	LocalOnly       bool       `json:"local_only"`
	Tasks           []CrewTask `json:"tasks"`
	Digest          string     `json:"digest"`
	CreatedAtUTC    string     `json:"created_at_utc"`
	Next            string     `json:"next"`
}

type CrewCheckpoint struct {
	Schema           int       `json:"schema"`
	PlanID           string    `json:"plan_id"`
	PlanDigest       string    `json:"plan_digest"`
	TaskID           string    `json:"task_id"`
	Sequence         int       `json:"sequence"`
	Attempt          int       `json:"attempt"`
	State            CrewState `json:"state"`
	ProviderID       string    `json:"provider_id"`
	ModelID          string    `json:"model_id"`
	SessionID        string    `json:"session_id"`
	Worktree         string    `json:"worktree"`
	CandidateCommit  string    `json:"candidate_commit"`
	Verification     string    `json:"verification"`
	FailureSignature string    `json:"failure_signature"`
	RepeatedFailures int       `json:"repeated_failures"`
	QuotaResetAtUTC  string    `json:"quota_reset_at_utc"`
	Message          string    `json:"message"`
	UpdatedAtUTC     string    `json:"updated_at_utc"`
	Next             string    `json:"next"`
}

type CrewDecisionKind string

const (
	CrewDecisionConflict      CrewDecisionKind = "rebase-conflict"
	CrewDecisionScopeExpanded CrewDecisionKind = "scope-expanded"
	CrewDecisionTierThree     CrewDecisionKind = "tier3"
	CrewDecisionNoProgress    CrewDecisionKind = "no-progress"
	CrewDecisionBlocked       CrewDecisionKind = "blocked"
)

const (
	CrewAnswerRetry   = "retry"
	CrewAnswerRestart = "restart"
	CrewAnswerCancel  = "cancel"
)

type CrewDecision struct {
	Schema        int              `json:"schema"`
	ID            string           `json:"id"`
	TaskID        string           `json:"task_id"`
	Kind          CrewDecisionKind `json:"kind"`
	Question      string           `json:"question"`
	Options       []string         `json:"options"`
	Impact        int              `json:"impact"`
	Answer        string           `json:"answer"`
	CreatedAtUTC  string           `json:"created_at_utc"`
	AnsweredAtUTC string           `json:"answered_at_utc"`
}

// CrewDecisionFor returns the options and impact for one decision kind. Retry
// continues from the recorded checkpoint, restart begins a fresh attempt from
// the current target, and cancel ends the task. Tier 3 work is outside the
// crew, so it can only be cancelled and routed elsewhere.
func CrewDecisionFor(kind CrewDecisionKind) ([]string, int, bool) {
	switch kind {
	case CrewDecisionConflict:
		return []string{CrewAnswerRetry, CrewAnswerRestart, CrewAnswerCancel}, 2, true
	case CrewDecisionScopeExpanded:
		return []string{CrewAnswerRestart, CrewAnswerRetry, CrewAnswerCancel}, 3, true
	case CrewDecisionTierThree:
		return []string{CrewAnswerCancel}, 3, true
	case CrewDecisionNoProgress:
		return []string{CrewAnswerRetry, CrewAnswerRestart, CrewAnswerCancel}, 1, true
	case CrewDecisionBlocked:
		return []string{CrewAnswerRetry, CrewAnswerCancel}, 1, true
	default:
		return nil, 0, false
	}
}

// CrewPlanProblem returns the first structural reason a plan cannot run, or ""
// when it is valid. Digest correctness is checked by the adapter.
func CrewPlanProblem(plan CrewPlan) string {
	switch {
	case plan.Schema != CrewSchema:
		return "unsupported crew plan schema"
	case !CrewPlanIDValid(plan.ID):
		return "crew plan ID is invalid"
	case !hasPrefix(plan.ObjectiveDigest, "sha256:") || len(plan.ObjectiveDigest) != 71:
		return "crew objective digest is invalid"
	case !crewObjectID(plan.BaseCommit):
		return "crew base commit is not a full Git object ID"
	case !CrewBranchValid(plan.TargetBranch):
		return "crew target branch is unsafe"
	case plan.RiskCeiling != TierProduct:
		return "crew runs keep the Tier 2 ceiling"
	case plan.MaxWorkers < 1 || plan.MaxWorkers > CrewMaxWorkers:
		return "crew worker limit is out of range"
	case plan.RepairRounds < 0 || plan.RepairRounds > CrewMaxRepairRounds:
		return "crew repair rounds are out of range"
	case !plan.LocalOnly:
		return "crew delivery is local-only in this version"
	case len(plan.Tasks) == 0 || len(plan.Tasks) > CrewMaxTasks:
		return "crew plan must contain 1 to 32 tasks"
	}
	seen := make(map[string]bool, len(plan.Tasks))
	for _, task := range plan.Tasks {
		if seen[task.ID] {
			return "crew task ID is duplicated: " + task.ID
		}
		seen[task.ID] = true
		if problem := CrewTaskProblem(plan.ID, task); problem != "" {
			return problem
		}
	}
	return ""
}

func CrewTaskProblem(planID string, task CrewTask) string {
	if !hasPrefix(task.ID, planID+"-t") || !crewIdentifier(task.ID) {
		return "crew task ID is invalid: " + task.ID
	}
	if !task.Shape.Valid() {
		return "crew task shape must be ship or scout: " + task.ID
	}
	if !crewText(task.Title, 200) || !crewText(task.Objective, 1<<16) {
		return "crew task title or objective is missing or unbounded: " + task.ID
	}
	if len(task.AcceptanceCriteria) == 0 || len(task.AcceptanceCriteria) > 64 {
		return "crew task needs 1 to 64 acceptance criteria: " + task.ID
	}
	for _, criterion := range task.AcceptanceCriteria {
		if !crewText(criterion, 2048) {
			return "crew acceptance criterion is invalid: " + task.ID
		}
	}
	if task.Shape == CrewScout {
		if len(task.AllowedPaths) != 0 || len(task.Verification) != 0 {
			return "scout tasks are read-only and declare no paths or verification: " + task.ID
		}
		return ""
	}
	if len(task.AllowedPaths) == 0 || len(task.AllowedPaths) > 256 {
		return "ship task needs 1 to 256 allowed paths: " + task.ID
	}
	for _, pattern := range task.AllowedPaths {
		if !CrewPathPatternValid(pattern) {
			return "ship task path is not an exact path or dir/** pattern: " + pattern
		}
	}
	if len(task.Verification) == 0 || len(task.Verification) > 64 {
		return "ship task needs 1 to 64 verification commands: " + task.ID
	}
	for _, argv := range task.Verification {
		if len(argv) == 0 || len(argv) > 64 {
			return "ship task verification argv is invalid: " + task.ID
		}
		for _, argument := range argv {
			if argument == "" || len(argument) > 4096 || strings.ContainsAny(argument, "\x00\r\n") {
				return "ship task verification argument is invalid: " + task.ID
			}
		}
	}
	return ""
}

// CrewPathPatternValid accepts the same exact and recursive forms as
// ScopeContains: "path/to/file" or "dir/**".
func CrewPathPatternValid(pattern string) bool {
	if pattern == "" || len(pattern) > 1024 || hasPrefix(pattern, "/") || strings.ContainsAny(pattern, "\\\x00\r\n?[") {
		return false
	}
	body := pattern
	if hasSuffix(pattern, "/**") {
		body = pattern[:len(pattern)-3]
	}
	if body == "" || strings.Contains(body, "*") {
		return false
	}
	for _, segment := range strings.Split(body, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

// CrewPathScopesMayOverlap reports whether any path could be matched by both
// scopes. It is exact for the supported pattern forms.
func CrewPathScopesMayOverlap(left, right []string) bool {
	for _, a := range left {
		for _, b := range right {
			if crewPatternsOverlap(a, b) {
				return true
			}
		}
	}
	return false
}

func crewPatternsOverlap(a, b string) bool {
	aRecursive, bRecursive := hasSuffix(a, "/**"), hasSuffix(b, "/**")
	switch {
	case !aRecursive && !bRecursive:
		return a == b
	case aRecursive && bRecursive:
		aPrefix, bPrefix := a[:len(a)-2], b[:len(b)-2]
		return hasPrefix(aPrefix, bPrefix) || hasPrefix(bPrefix, aPrefix)
	case aRecursive:
		return ScopeContains([]string{a}, b)
	default:
		return ScopeContains([]string{b}, a)
	}
}

// NextCrewAdmissions returns the queued tasks to start now, in plan order. It
// never exceeds maxWorkers active tasks and never runs ship tasks whose scopes
// may overlap at the same time. Scout tasks hold no write scope.
func NextCrewAdmissions(tasks []CrewTask, states map[string]CrewState, maxWorkers int) []string {
	occupied := []string{}
	slots := maxWorkers
	for _, task := range tasks {
		if !states[task.ID].Active() {
			continue
		}
		slots--
		if task.Shape == CrewShip {
			occupied = append(occupied, task.AllowedPaths...)
		}
	}
	admitted := []string{}
	for _, task := range tasks {
		if slots <= 0 {
			break
		}
		if states[task.ID] != CrewQueued {
			continue
		}
		if task.Shape == CrewShip {
			if CrewPathScopesMayOverlap(task.AllowedPaths, occupied) {
				continue
			}
			occupied = append(occupied, task.AllowedPaths...)
		}
		admitted = append(admitted, task.ID)
		slots--
	}
	return admitted
}

func CrewPlanIDValid(value string) bool {
	return hasPrefix(value, "crew-") && len(value) == 17 && crewIdentifier(value) && crewHex(value[5:])
}

func CrewBranchValid(value string) bool {
	return value != "" && len(value) <= 255 && !strings.ContainsAny(value, " ~^:?*[\\\x00\r\n") &&
		!hasPrefix(value, "-") && !hasPrefix(value, "/") && !hasSuffix(value, "/") &&
		!strings.Contains(value, "..") && !strings.Contains(value, "//") && !strings.Contains(value, "@{") &&
		!hasSuffix(value, ".lock") && !hasSuffix(value, ".") && !hasPrefix(value, "refs/")
}

func crewIdentifier(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
			return false
		}
	}
	return true
}

func crewObjectID(value string) bool { return len(value) == 40 && crewHex(value) }

func crewHex(value string) bool {
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return value != ""
}

func crewText(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && !strings.ContainsRune(value, 0)
}
