// Package headless owns durable manifest approval and multi-wave checkpoints.
package headless

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const maxRecordBytes = 16 << 20

type Planner struct{ now func() time.Time }

type PlanRequest struct {
	ObjectivePath   string
	Objective       []byte
	BaseCommit      string
	TargetBranch    string
	AllowedPaths    []string
	AllowedCommands [][]string
	ProviderPolicy  string
	NetworkPolicy   string
	LocalMerge      bool
}

type Approval struct {
	Schema         int    `json:"schema"`
	RunID          string `json:"run_id"`
	ManifestDigest string `json:"manifest_digest"`
	Actor          string `json:"actor"`
	Role           string `json:"role"`
	ApprovedAtUTC  string `json:"approved_at_utc"`
	Source         string `json:"source"`
}

type Event struct {
	Schema    int    `json:"schema"`
	RunID     string `json:"run_id"`
	Sequence  int    `json:"sequence"`
	State     string `json:"state"`
	WaveID    string `json:"wave_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at_utc"`
	Next      string `json:"next"`
}

type WaveOutcome struct {
	State            string
	ProviderID       string
	ModelID          string
	SessionID        string
	Worktree         string
	CandidateCommit  string
	FailureSignature string
	QuotaResetAtUTC  string
	Message          string
}

type Executor interface {
	Execute(context.Context, domain.HeadlessManifest, domain.HeadlessWave, domain.HeadlessCheckpoint) (WaveOutcome, error)
}

type WaitFunc func(context.Context, time.Time) error

type Engine struct {
	now  func() time.Time
	wait WaitFunc
}

func NewPlanner() Planner { return Planner{now: time.Now} }
func NewPlannerWith(now func() time.Time) Planner {
	if now == nil {
		now = time.Now
	}
	return Planner{now: now}
}
func NewEngine() Engine { return Engine{now: time.Now, wait: waitUntil} }
func NewEngineWith(now func() time.Time, wait WaitFunc) Engine {
	if now == nil {
		now = time.Now
	}
	if wait == nil {
		wait = waitUntil
	}
	return Engine{now: now, wait: wait}
}

func (planner Planner) Plan(request PlanRequest) (domain.HeadlessManifest, error) {
	if !safeObjectivePath(request.ObjectivePath) || len(request.Objective) < 2 || len(request.Objective) > 1<<20 || !utf8.Valid(request.Objective) || len(request.BaseCommit) != 40 || !hex(request.BaseCommit) || !safeBranch(request.TargetBranch) || request.ProviderPolicy != "balanced" || request.NetworkPolicy != "gateway-only" || len(request.AllowedPaths) == 0 || len(request.AllowedPaths) > 256 || len(request.AllowedCommands) == 0 || len(request.AllowedCommands) > 64 {
		return domain.HeadlessManifest{}, errors.New("Headless planning input is incomplete or unsafe")
	}
	for _, path := range request.AllowedPaths {
		if !safePattern(path) {
			return domain.HeadlessManifest{}, fmt.Errorf("unsafe allowed path %q", path)
		}
	}
	for _, command := range request.AllowedCommands {
		if len(command) == 0 || len(command) > 64 {
			return domain.HeadlessManifest{}, errors.New("Headless command bounds are invalid")
		}
		for _, argument := range command {
			if argument == "" || len(argument) > 4096 || strings.ContainsAny(argument, "\x00\r\n") {
				return domain.HeadlessManifest{}, errors.New("Headless command argument is invalid")
			}
		}
	}
	waves, err := parseWaves(string(request.Objective), request.AllowedPaths, request.AllowedCommands)
	if err != nil {
		return domain.HeadlessManifest{}, err
	}
	objectiveDigest := fmt.Sprintf("sha256:%x", sha256.Sum256(request.Objective))
	created := planner.now().UTC()
	manifest := domain.HeadlessManifest{
		Schema: domain.OrchestrationSchema, ID: "headless-" + objectiveDigest[7:19], ObjectivePath: request.ObjectivePath,
		ObjectiveDigest: objectiveDigest, BaseCommit: request.BaseCommit, TargetBranch: request.TargetBranch,
		RiskCeiling: domain.TierProduct, AllowedPaths: sortedCopy(request.AllowedPaths), AllowedCommands: copyCommands(request.AllowedCommands),
		ProviderPolicy: request.ProviderPolicy, NetworkPolicy: request.NetworkPolicy, LocalMerge: request.LocalMerge,
		StopBeforeDeploy: true, Waves: waves, CreatedAtUTC: created.Format(time.RFC3339), Next: "review the warning and approve this manifest digest",
	}
	digest, err := manifestDigest(manifest)
	if err != nil {
		return domain.HeadlessManifest{}, err
	}
	manifest.Digest = digest
	return manifest, nil
}

func SaveManifest(common string, manifest domain.HeadlessManifest) error {
	if err := validateManifest(manifest); err != nil {
		return err
	}
	return save(common, manifest.ID, "manifest.json", manifest, false)
}

func LoadManifest(common, runID string) (domain.HeadlessManifest, error) {
	var manifest domain.HeadlessManifest
	if err := load(common, runID, "manifest.json", &manifest); err != nil {
		return manifest, err
	}
	if err := validateManifest(manifest); err != nil {
		return domain.HeadlessManifest{}, err
	}
	return manifest, nil
}

func Approve(common string, manifest domain.HeadlessManifest, digest, actor, role string, now time.Time) (Approval, error) {
	if err := validateManifest(manifest); err != nil {
		return Approval{}, err
	}
	if digest != manifest.Digest || actor == "" || role == "" || len(actor) > 256 || len(role) > 256 || strings.ContainsAny(actor+role, "\x00\r\n") {
		return Approval{}, errors.New("Headless approval does not bind the manifest and owner")
	}
	approval := Approval{Schema: domain.OrchestrationSchema, RunID: manifest.ID, ManifestDigest: digest, Actor: actor, Role: role, ApprovedAtUTC: now.UTC().Format(time.RFC3339), Source: "active-owner-interaction"}
	if err := save(common, manifest.ID, "approval.json", approval, false); err != nil {
		return Approval{}, err
	}
	return approval, nil
}

func LoadApproval(common, runID string) (Approval, error) {
	var approval Approval
	if err := load(common, runID, "approval.json", &approval); err != nil {
		return approval, err
	}
	if approval.Schema != domain.OrchestrationSchema || approval.RunID != runID || approval.ManifestDigest == "" || approval.Actor == "" || approval.Role == "" || approval.Source != "active-owner-interaction" {
		return Approval{}, errors.New("Headless approval is invalid")
	}
	return approval, nil
}

func (engine Engine) Start(ctx context.Context, common string, manifest domain.HeadlessManifest, approval Approval, executor Executor) (domain.HeadlessCheckpoint, error) {
	if executor == nil || approval.RunID != manifest.ID || approval.ManifestDigest != manifest.Digest || approval.Source != "active-owner-interaction" || !manifest.LocalMerge || !manifest.StopBeforeDeploy || manifest.RiskCeiling != domain.TierProduct {
		return domain.HeadlessCheckpoint{}, errors.New("Headless start lacks current bounded authority")
	}
	checkpoint := domain.HeadlessCheckpoint{Schema: domain.OrchestrationSchema, RunID: manifest.ID, ManifestDigest: manifest.Digest, Sequence: 0, State: "running", UpdatedAtUTC: engine.now().UTC().Format(time.RFC3339), Next: "execute first feature wave"}
	if err := saveCheckpoint(common, checkpoint); err != nil {
		return checkpoint, err
	}
	if err := engine.event(common, checkpoint, "Headless run started"); err != nil {
		return checkpoint, err
	}
	return engine.run(ctx, common, manifest, checkpoint, executor)
}

func (engine Engine) Resume(ctx context.Context, common string, executor Executor) (domain.HeadlessCheckpoint, error) {
	checkpoint, err := LoadCheckpoint(common)
	if err != nil {
		return checkpoint, err
	}
	manifest, err := LoadManifest(common, checkpoint.RunID)
	if err != nil {
		return checkpoint, err
	}
	approval, err := LoadApproval(common, checkpoint.RunID)
	if err != nil {
		return checkpoint, err
	}
	if approval.ManifestDigest != manifest.Digest || checkpoint.ManifestDigest != manifest.Digest || checkpoint.State == "complete" || checkpoint.State == "cancelled" {
		return checkpoint, errors.New("Headless checkpoint cannot be resumed")
	}
	checkpoint.State = "running"
	checkpoint.UpdatedAtUTC = engine.now().UTC().Format(time.RFC3339)
	checkpoint.Next = "resume the current feature wave"
	if err := saveCheckpoint(common, checkpoint); err != nil {
		return checkpoint, err
	}
	return engine.run(ctx, common, manifest, checkpoint, executor)
}

func Cancel(common string, now time.Time) (domain.HeadlessCheckpoint, error) {
	checkpoint, err := LoadCheckpoint(common)
	if err != nil {
		return checkpoint, err
	}
	if checkpoint.State == "complete" {
		return checkpoint, errors.New("completed Headless run cannot be cancelled")
	}
	checkpoint.State = "cancelled"
	checkpoint.Sequence++
	checkpoint.UpdatedAtUTC = now.UTC().Format(time.RFC3339)
	checkpoint.Next = "inspect retained worktrees and start a new approved manifest"
	if err := saveCheckpoint(common, checkpoint); err != nil {
		return checkpoint, err
	}
	return checkpoint, nil
}

func (engine Engine) run(ctx context.Context, common string, manifest domain.HeadlessManifest, checkpoint domain.HeadlessCheckpoint, executor Executor) (domain.HeadlessCheckpoint, error) {
	start := waveIndex(manifest.Waves, checkpoint.WaveID)
	if checkpoint.WaveID == "" {
		start = 0
	}
	for index := start; index < len(manifest.Waves); {
		if err := ctx.Err(); err != nil {
			checkpoint.State = "paused"
			checkpoint.Next = "run l7 headless resume after cancellation is cleared"
			_ = saveCheckpoint(common, checkpoint)
			return checkpoint, err
		}
		wave := manifest.Waves[index]
		checkpoint.WaveID = wave.ID
		checkpoint.Sequence++
		checkpoint.UpdatedAtUTC = engine.now().UTC().Format(time.RFC3339)
		checkpoint.Next = "execute and verify wave " + wave.ID
		if err := saveCheckpoint(common, checkpoint); err != nil {
			return checkpoint, err
		}
		if err := engine.event(common, checkpoint, "wave execution dispatched"); err != nil {
			return checkpoint, err
		}
		outcome, err := executor.Execute(ctx, manifest, wave, checkpoint)
		if err != nil {
			outcome.State = "failed"
			if outcome.FailureSignature == "" {
				outcome.FailureSignature = fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(err.Error())))
			}
		}
		checkpoint.ProviderID, checkpoint.ModelID, checkpoint.SessionID, checkpoint.Worktree, checkpoint.CandidateCommit = outcome.ProviderID, outcome.ModelID, outcome.SessionID, outcome.Worktree, outcome.CandidateCommit
		switch outcome.State {
		case "complete":
			checkpoint.FailureSignature = ""
			checkpoint.RepeatedFailures = 0
			checkpoint.QuotaResetAtUTC = ""
			checkpoint.Sequence++
			index++
			if index < len(manifest.Waves) {
				checkpoint.WaveID = manifest.Waves[index].ID
				checkpoint.Next = "execute next feature wave"
			} else {
				checkpoint.State = "complete"
				checkpoint.WaveID = wave.ID
				checkpoint.Next = "inspect local merges; release and deployment remain separate"
			}
		case "quota":
			reset, parseErr := time.Parse(time.RFC3339, outcome.QuotaResetAtUTC)
			if parseErr != nil || !reset.After(engine.now()) {
				checkpoint.State = "paused"
				checkpoint.Next = "repair missing provider quota reset metadata, then resume"
				if err := saveCheckpoint(common, checkpoint); err != nil {
					return checkpoint, err
				}
				return checkpoint, errors.New("provider quota reset is invalid")
			}
			checkpoint.State = "waiting-quota"
			checkpoint.QuotaResetAtUTC = reset.UTC().Format(time.RFC3339)
			checkpoint.Next = "wait for the natural provider reset"
			if err := saveCheckpoint(common, checkpoint); err != nil {
				return checkpoint, err
			}
			if err := engine.wait(ctx, reset); err != nil {
				checkpoint.State = "paused"
				checkpoint.Next = "run l7 headless resume after the natural reset"
				_ = saveCheckpoint(common, checkpoint)
				return checkpoint, err
			}
			checkpoint.State = "running"
			checkpoint.QuotaResetAtUTC = ""
		case "blocked", "scope-expanded", "tier3", "external-effect":
			checkpoint.State = "paused"
			checkpoint.Next = "obtain new authority or create a revised manifest; do not widen this run"
			if err := saveCheckpoint(common, checkpoint); err != nil {
				return checkpoint, err
			}
			_ = engine.event(common, checkpoint, outcome.Message)
			return checkpoint, nil
		default:
			if checkpoint.FailureSignature == outcome.FailureSignature {
				checkpoint.RepeatedFailures++
			} else {
				checkpoint.FailureSignature = outcome.FailureSignature
				checkpoint.RepeatedFailures = 1
			}
			if checkpoint.RepeatedFailures >= 3 {
				checkpoint.State = "paused"
				checkpoint.Next = "inspect the repeated failure and resume only after the cause changes"
				if err := saveCheckpoint(common, checkpoint); err != nil {
					return checkpoint, err
				}
				_ = engine.event(common, checkpoint, "no-progress circuit breaker opened")
				return checkpoint, nil
			}
			checkpoint.State = "running"
			checkpoint.Next = "retry the same wave with recorded failure context"
		}
		if err := saveCheckpoint(common, checkpoint); err != nil {
			return checkpoint, err
		}
		if checkpoint.State == "complete" {
			_ = engine.event(common, checkpoint, "all locally merged waves completed")
			return checkpoint, nil
		}
	}
	return checkpoint, nil
}

func LoadCheckpoint(common string) (domain.HeadlessCheckpoint, error) {
	var pointer struct {
		Schema int    `json:"schema"`
		RunID  string `json:"run_id"`
	}
	root, err := root(common)
	if err != nil {
		return domain.HeadlessCheckpoint{}, err
	}
	data, err := localfile.Read(filepath.Join(root, "active.json"), maxRecordBytes)
	if err != nil {
		return domain.HeadlessCheckpoint{}, err
	}
	if err := localfile.DecodeJSON(data, &pointer); err != nil || pointer.Schema != domain.OrchestrationSchema || !safeRunID(pointer.RunID) {
		return domain.HeadlessCheckpoint{}, errors.New("Headless active pointer is invalid")
	}
	var checkpoint domain.HeadlessCheckpoint
	if err := load(common, pointer.RunID, "checkpoint.json", &checkpoint); err != nil {
		return checkpoint, err
	}
	if checkpoint.Schema != domain.OrchestrationSchema || checkpoint.RunID != pointer.RunID || checkpoint.ManifestDigest == "" || checkpoint.Sequence < 0 || checkpoint.State == "" || checkpoint.Next == "" {
		return domain.HeadlessCheckpoint{}, errors.New("Headless checkpoint is invalid")
	}
	return checkpoint, nil
}

func saveCheckpoint(common string, checkpoint domain.HeadlessCheckpoint) error {
	if !safeRunID(checkpoint.RunID) || checkpoint.Schema != domain.OrchestrationSchema {
		return errors.New("Headless checkpoint identity is invalid")
	}
	if err := save(common, checkpoint.RunID, "checkpoint.json", checkpoint, true); err != nil {
		return err
	}
	pointer := struct {
		Schema int    `json:"schema"`
		RunID  string `json:"run_id"`
	}{Schema: domain.OrchestrationSchema, RunID: checkpoint.RunID}
	root, err := root(common)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(root, 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(pointer)
	if err != nil {
		return err
	}
	path := filepath.Join(root, "active.json")
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	} else if err != nil {
		return err
	}
	return localfile.AtomicReplace(path, data, 0o600)
}

func (engine Engine) event(common string, checkpoint domain.HeadlessCheckpoint, message string) error {
	event := Event{Schema: domain.OrchestrationSchema, RunID: checkpoint.RunID, Sequence: checkpoint.Sequence, State: checkpoint.State, WaveID: checkpoint.WaveID, Message: bounded(message, 1024), CreatedAt: engine.now().UTC().Format(time.RFC3339), Next: checkpoint.Next}
	return save(common, checkpoint.RunID, fmt.Sprintf("events/%08d.json", event.Sequence), event, false)
}

func parseWaves(objective string, allowed []string, commands [][]string) ([]domain.HeadlessWave, error) {
	lines := strings.Split(objective, "\n")
	sections := []struct {
		title string
		body  []string
	}{}
	current := struct {
		title string
		body  []string
	}{title: "Objective"}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && (strings.Contains(strings.ToLower(trimmed), "wave") || strings.Contains(strings.ToLower(trimmed), "feature")) {
			if len(current.body) > 0 {
				sections = append(sections, current)
			}
			current = struct {
				title string
				body  []string
			}{title: strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))}
			continue
		}
		current.body = append(current.body, line)
	}
	if len(current.body) > 0 {
		sections = append(sections, current)
	}
	waves := []domain.HeadlessWave{}
	for index, section := range sections {
		acceptance := []string{}
		for _, line := range section.body {
			trimmed := strings.TrimSpace(line)
			lower := strings.ToLower(trimmed)
			if strings.HasPrefix(lower, "acceptance:") || strings.HasPrefix(lower, "success:") || strings.HasPrefix(trimmed, "- [ ]") {
				value := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "- [ ]"), "Acceptance:"), "Success:"))
				if value != "" {
					acceptance = append(acceptance, value)
				}
			}
		}
		if len(acceptance) == 0 {
			continue
		}
		wave := domain.HeadlessWave{ID: fmt.Sprintf("wave-%02d", index+1), Title: section.title, Scope: sortedCopy(allowed), AcceptanceCriteria: acceptance, Verification: copyCommands(commands)}
		waves = append(waves, wave)
	}
	if len(waves) == 0 {
		return nil, errors.New("Headless objective has no measurable Acceptance:, Success:, or checklist criteria")
	}
	return waves, nil
}

func validateManifest(manifest domain.HeadlessManifest) error {
	if manifest.Schema != domain.OrchestrationSchema || !safeRunID(manifest.ID) || !safeObjectivePath(manifest.ObjectivePath) || !strings.HasPrefix(manifest.ObjectiveDigest, "sha256:") || len(manifest.BaseCommit) != 40 || !hex(manifest.BaseCommit) || !safeBranch(manifest.TargetBranch) || manifest.RiskCeiling != domain.TierProduct || manifest.ProviderPolicy != "balanced" || manifest.NetworkPolicy != "gateway-only" || !manifest.StopBeforeDeploy || len(manifest.Waves) == 0 || len(manifest.Waves) > 128 || manifest.Digest == "" {
		return errors.New("Headless manifest is invalid")
	}
	digest, err := manifestDigest(manifest)
	if err != nil || digest != manifest.Digest {
		return errors.New("Headless manifest digest is stale")
	}
	return nil
}
func manifestDigest(manifest domain.HeadlessManifest) (string, error) {
	copy := manifest
	copy.Digest = ""
	data, err := localfile.EncodeJSON(copy)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}
func save(common, runID, relative string, value any, replace bool) error {
	base, err := root(common)
	if err != nil {
		return err
	}
	if !safeRunID(runID) {
		return errors.New("Headless run ID is invalid")
	}
	path := filepath.Join(base, runID, filepath.FromSlash(relative))
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > maxRecordBytes {
		return errors.New("Headless record is invalid or unbounded")
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	} else if err != nil {
		return err
	} else if !replace {
		return os.ErrExist
	}
	return localfile.AtomicReplace(path, data, 0o600)
}
func load(common, runID, relative string, target any) error {
	base, err := root(common)
	if err != nil {
		return err
	}
	if !safeRunID(runID) {
		return errors.New("Headless run ID is invalid")
	}
	data, err := localfile.Read(filepath.Join(base, runID, filepath.FromSlash(relative)), maxRecordBytes)
	if err != nil {
		return err
	}
	return localfile.DecodeJSON(data, target)
}
func root(common string) (string, error) {
	if !filepath.IsAbs(common) {
		return "", errors.New("Git common directory must be absolute")
	}
	info, err := os.Lstat(filepath.Clean(common))
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("Git common directory is unsafe")
	}
	return filepath.Join(filepath.Clean(common), "l7", "headless"), nil
}
func waitUntil(ctx context.Context, reset time.Time) error {
	duration := time.Until(reset)
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func waveIndex(waves []domain.HeadlessWave, id string) int {
	for index, wave := range waves {
		if wave.ID == id {
			return index
		}
	}
	return 0
}
func safeObjectivePath(value string) bool {
	value = filepath.ToSlash(filepath.Clean(value))
	return value != "." && value != ".." && !filepath.IsAbs(value) && !strings.HasPrefix(value, "../") && !strings.Contains(value, "\\") && len(value) <= 1024
}
func safePattern(value string) bool {
	value = filepath.ToSlash(filepath.Clean(value))
	return value != "." && value != ".." && !filepath.IsAbs(value) && !strings.HasPrefix(value, "../") && !strings.Contains(value, "\\") && len(value) <= 1024
}
func safeBranch(value string) bool {
	return value != "" && len(value) <= 255 && !strings.ContainsAny(value, " ~^:?*[\\\x00\r\n") && !strings.HasPrefix(value, "-") && !strings.Contains(value, "..") && !strings.HasSuffix(value, ".lock")
}
func safeRunID(value string) bool {
	if !strings.HasPrefix(value, "headless-") || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
			return false
		}
	}
	return true
}
func hex(value string) bool {
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return value != ""
}
func sortedCopy(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}
func copyCommands(values [][]string) [][]string {
	result := make([][]string, len(values))
	for index := range values {
		result[index] = append([]string{}, values[index]...)
	}
	return result
}
func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit]
}
