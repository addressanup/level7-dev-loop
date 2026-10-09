// Package crew plans, records, and supervises parallel local crew tasks.
package crew

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const maxObjectiveBytes = 1 << 20

type PlanRequest struct {
	ObjectivePath       string
	Objective           []byte
	BaseCommit          string
	TargetBranch        string
	MaxWorkers          int
	RepairRounds        int
	DefaultVerification [][]string
}

type Planner struct{ now func() time.Time }

func NewPlanner() Planner { return Planner{now: time.Now} }

func NewPlannerWith(now func() time.Time) Planner {
	if now == nil {
		now = time.Now
	}
	return Planner{now: now}
}

// Plan turns an objective document into one digest-bound crew plan. Each task
// is a "## Ship: <title>" or "## Scout: <title>" section. Inside a section,
// "Paths:" lists comma-separated scopes, "Verify:" holds one JSON argv array,
// and "Acceptance:", "Success:", or "- [ ]" lines are acceptance criteria.
// Text before the first section is shared context for every task.
func (planner Planner) Plan(request PlanRequest) (domain.CrewPlan, error) {
	if len(request.Objective) < 2 || len(request.Objective) > maxObjectiveBytes || !utf8.Valid(request.Objective) {
		return domain.CrewPlan{}, errors.New("crew objective must be 2 bytes to 1 MiB of UTF-8 text")
	}
	if !safeObjectivePath(request.ObjectivePath) {
		return domain.CrewPlan{}, errors.New("crew objective path is unsafe")
	}
	objectiveDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(request.Objective))
	identity := sha256.Sum256([]byte(objectiveDigest + "\x00" + request.BaseCommit + "\x00" + request.TargetBranch))
	plan := domain.CrewPlan{
		Schema: domain.CrewSchema, ID: fmt.Sprintf("crew-%x", identity[:6]), ObjectivePath: request.ObjectivePath,
		ObjectiveDigest: objectiveDigest, BaseCommit: request.BaseCommit, TargetBranch: request.TargetBranch,
		RiskCeiling: domain.TierProduct, MaxWorkers: request.MaxWorkers, RepairRounds: request.RepairRounds, LocalOnly: true,
		CreatedAtUTC: planner.now().UTC().Format(time.RFC3339),
		Next:         "review every task scope, then approve this exact plan digest",
	}
	tasks, err := parseTasks(plan.ID, string(request.Objective), request.DefaultVerification)
	if err != nil {
		return domain.CrewPlan{}, err
	}
	plan.Tasks = tasks
	if problem := domain.CrewPlanProblem(plan); problem != "" {
		return domain.CrewPlan{}, errors.New(problem)
	}
	digest, err := PlanDigest(plan)
	if err != nil {
		return domain.CrewPlan{}, err
	}
	plan.Digest = digest
	return plan, nil
}

// PlanDigest binds every plan field except the digest itself.
func PlanDigest(plan domain.CrewPlan) (string, error) {
	plan.Digest = ""
	data, err := localfile.EncodeJSON(plan)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}

func ValidatePlan(plan domain.CrewPlan) error {
	if problem := domain.CrewPlanProblem(plan); problem != "" {
		return errors.New(problem)
	}
	digest, err := PlanDigest(plan)
	if err != nil || digest != plan.Digest {
		return errors.New("crew plan digest is stale")
	}
	return nil
}

type section struct {
	shape        domain.CrewShape
	title        string
	body         []string
	paths        []string
	verification [][]string
	acceptance   []string
}

func parseTasks(planID, objective string, defaults [][]string) ([]domain.CrewTask, error) {
	var sections []*section
	var preamble []string
	var current *section
	for number, line := range strings.Split(objective, "\n") {
		trimmed := strings.TrimSpace(line)
		if shape, title, ok := sectionHeading(trimmed); ok {
			current = &section{shape: shape, title: title}
			sections = append(sections, current)
			continue
		}
		if current == nil {
			preamble = append(preamble, line)
			continue
		}
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "paths:"):
			for _, value := range strings.Split(trimmed[len("paths:"):], ",") {
				if value = strings.TrimSpace(value); value != "" {
					current.paths = append(current.paths, value)
				}
			}
			continue
		case strings.HasPrefix(lower, "verify:"):
			var argv []string
			if err := json.Unmarshal([]byte(strings.TrimSpace(trimmed[len("verify:"):])), &argv); err != nil {
				return nil, fmt.Errorf("line %d: Verify: must be one JSON argv array", number+1)
			}
			current.verification = append(current.verification, argv)
			continue
		case strings.HasPrefix(lower, "acceptance:"), strings.HasPrefix(lower, "success:"):
			if value := strings.TrimSpace(trimmed[strings.Index(trimmed, ":")+1:]); value != "" {
				current.acceptance = append(current.acceptance, value)
			}
		case strings.HasPrefix(trimmed, "- [ ]"):
			if value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ]")); value != "" {
				current.acceptance = append(current.acceptance, value)
			}
		}
		current.body = append(current.body, line)
	}
	if len(sections) == 0 {
		return nil, errors.New(`crew objective has no "## Ship: <title>" or "## Scout: <title>" task sections`)
	}
	if len(sections) > domain.CrewMaxTasks {
		return nil, fmt.Errorf("crew objective has more than %d tasks", domain.CrewMaxTasks)
	}
	shared := strings.TrimSpace(strings.Join(preamble, "\n"))
	tasks := make([]domain.CrewTask, 0, len(sections))
	for index, parsed := range sections {
		body := strings.TrimSpace(strings.Join(parsed.body, "\n"))
		if body == "" {
			body = parsed.title
		}
		if shared != "" {
			body = "Context:\n" + shared + "\n\nTask:\n" + body
		}
		task := domain.CrewTask{
			ID: fmt.Sprintf("%s-t%02d", planID, index+1), Shape: parsed.shape, Title: parsed.title, Objective: body,
			AcceptanceCriteria: parsed.acceptance, AllowedPaths: []string{}, Verification: [][]string{},
		}
		if parsed.shape == domain.CrewShip {
			task.AllowedPaths = parsed.paths
			task.Verification = parsed.verification
			if len(task.Verification) == 0 {
				task.Verification = copyCommands(defaults)
			}
		} else if len(parsed.paths) != 0 || len(parsed.verification) != 0 {
			return nil, fmt.Errorf("scout task %q is read-only and cannot declare Paths: or Verify:", parsed.title)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func sectionHeading(line string) (domain.CrewShape, string, bool) {
	if !strings.HasPrefix(line, "## ") {
		return "", "", false
	}
	heading := strings.TrimSpace(line[3:])
	for _, shape := range []domain.CrewShape{domain.CrewShip, domain.CrewScout} {
		prefix := string(shape) + ":"
		if len(heading) > len(prefix) && strings.EqualFold(heading[:len(prefix)], prefix) {
			return shape, strings.TrimSpace(heading[len(prefix):]), true
		}
	}
	return "", "", false
}

func safeObjectivePath(value string) bool {
	return value != "" && len(value) <= 1024 && !strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "../") &&
		value != ".." && !strings.Contains(value, "/../") && !strings.ContainsAny(value, "\\\x00\r\n")
}

func copyCommands(values [][]string) [][]string {
	result := make([][]string, len(values))
	for index := range values {
		result[index] = append([]string{}, values[index]...)
	}
	return result
}
