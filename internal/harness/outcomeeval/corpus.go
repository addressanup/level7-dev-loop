package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

//go:embed corpus/protocol.json corpus/tasks/*.txtar
var embeddedCorpus embed.FS

const (
	hiddenDirectory  = "_l7_hidden/"
	maxCorpusFile    = 64 << 10
	maxArchiveBytes  = 256 << 10
	maxAcceptance    = 12
	maxTaskArguments = 16
)

var taskIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,40}$`)

type protocol struct {
	Schema         int           `json:"schema"`
	ID             string        `json:"id"`
	Version        string        `json:"version"`
	CorpusSHA256   string        `json:"corpus_sha256"`
	Tasks          []string      `json:"tasks"`
	TrialsPerTask  int           `json:"trials_per_task"`
	Arms           []string      `json:"arms"`
	Order          string        `json:"order"`
	Provider       string        `json:"provider"`
	RouteProfile   routeProfile  `json:"route_profile"`
	Timeouts       timeouts      `json:"timeouts_seconds"`
	Canary         string        `json:"canary"`
	ProtectedPaths []string      `json:"protected_paths"`
	Decision       decisionRules `json:"decision"`
	EvidenceClass  string        `json:"evidence_class"`
	Limitations    []string      `json:"limitations"`
}

type routeProfile struct {
	Complexity string `json:"complexity"`
	Risk       int    `json:"risk"`
	Context    int    `json:"context"`
}

type timeouts struct {
	Plain int `json:"plain"`
	Crew  int `json:"crew"`
	Check int `json:"check"`
}

type decisionRules struct {
	MinValidBasisPoints      int    `json:"min_valid_basis_points"`
	MaxCrewFalseSuccess      int    `json:"max_crew_false_success"`
	MaxCrewSafetyViolations  int    `json:"max_crew_safety_violations"`
	CrewAtLeastPlain         bool   `json:"crew_delivered_at_least_plain"`
	ImprovementTest          string `json:"improvement_test"`
	AlphaBasisPoints         int    `json:"alpha_basis_points"`
	RequireSafetyNonInferior bool   `json:"require_safety_non_inferior"`
}

type forbidden struct {
	Unchanged     []string            `json:"unchanged,omitempty"`
	NoAddedText   []string            `json:"no_added_text,omitempty"`
	RetainedLines map[string][]string `json:"retained_lines,omitempty"`
}

type task struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Category   string     `json:"category"`
	Objective  string     `json:"objective"`
	Acceptance []string   `json:"acceptance"`
	Paths      []string   `json:"paths"`
	Verify     [][]string `json:"verify"`
	Hidden     [][]string `json:"hidden"`
	Forbidden  forbidden  `json:"forbidden"`

	repo      map[string][]byte
	hidden    map[string][]byte
	reference map[string][]byte
}

type corpus struct {
	protocol       protocol
	protocolSHA256 string
	digest         string
	tasks          []task
}

func (c corpus) task(id string) (task, bool) {
	for _, candidate := range c.tasks {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return task{}, false
}

// loadCorpus decodes the frozen protocol and every task archive, and refuses a
// corpus whose digest, task list, or structure disagrees with the protocol.
func loadCorpus(files fs.FS) (corpus, error) {
	protocolData, err := fs.ReadFile(files, "corpus/protocol.json")
	if err != nil {
		return corpus{}, fmt.Errorf("read protocol: %w", err)
	}
	var loaded corpus
	if err := decodeStrict(protocolData, &loaded.protocol); err != nil {
		return corpus{}, fmt.Errorf("protocol: %w", err)
	}
	sum := sha256.Sum256(protocolData)
	loaded.protocolSHA256 = hex.EncodeToString(sum[:])
	if loaded.digest, err = corpusDigest(files); err != nil {
		return corpus{}, err
	}
	names, _ := fs.Glob(files, "corpus/tasks/*.txtar")
	byID := map[string]task{}
	for _, name := range names {
		data, err := fs.ReadFile(files, name)
		if err != nil {
			return corpus{}, err
		}
		parsed, err := parseTask(data)
		if err != nil {
			return corpus{}, fmt.Errorf("%s: %w", path.Base(name), err)
		}
		if parsed.ID+".txtar" != path.Base(name) {
			return corpus{}, fmt.Errorf("%s: task id %q does not match its file name", path.Base(name), parsed.ID)
		}
		byID[parsed.ID] = parsed
	}
	if loaded.protocol.CorpusSHA256 != loaded.digest {
		return corpus{}, fmt.Errorf("corpus digest %s does not match the protocol's %s", loaded.digest, loaded.protocol.CorpusSHA256)
	}
	if err := validateProtocol(loaded.protocol); err != nil {
		return corpus{}, err
	}
	if len(loaded.protocol.Tasks) != len(byID) {
		return corpus{}, errors.New("protocol task list and corpus archives differ")
	}
	for _, id := range loaded.protocol.Tasks {
		parsed, found := byID[id]
		if !found {
			return corpus{}, fmt.Errorf("protocol task %s has no archive", id)
		}
		if err := validateTask(parsed, loaded.protocol); err != nil {
			return corpus{}, fmt.Errorf("task %s: %w", id, err)
		}
		loaded.tasks = append(loaded.tasks, parsed)
	}
	for _, left := range loaded.tasks {
		for _, right := range loaded.tasks {
			if left.ID != right.ID && strings.Contains(left.Title, right.Title) {
				return corpus{}, fmt.Errorf("task title %q contains task title %q", left.Title, right.Title)
			}
		}
	}
	return loaded, nil
}

// corpusDigest binds every task archive's name, length, and bytes, in name
// order. The protocol records it, so editing a task without re-freezing the
// protocol fails closed.
func corpusDigest(files fs.FS) (string, error) {
	names, err := fs.Glob(files, "corpus/tasks/*.txtar")
	if err != nil || len(names) == 0 {
		return "", errors.New("corpus has no task archives")
	}
	digest := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(files, name)
		if err != nil || len(data) > maxArchiveBytes {
			return "", fmt.Errorf("task archive %s is unreadable or too large", name)
		}
		fmt.Fprintf(digest, "%s\x00%d\x00", path.Base(name), len(data))
		digest.Write(data)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func validateProtocol(value protocol) error {
	rules := value.Decision
	switch {
	case value.Schema != 1 || value.ID == "" || value.Version == "":
		return errors.New("protocol identity is incomplete")
	case value.TrialsPerTask < 1 || value.TrialsPerTask > 10:
		return errors.New("protocol trials_per_task must be 1 to 10")
	case strings.Join(value.Arms, ",") != "crew,plain" || value.Order != "alternate" || value.Provider != "codex":
		return errors.New("protocol arms, order, or provider is unsupported")
	case value.RouteProfile.Complexity == "" || value.RouteProfile.Risk < 1 || value.RouteProfile.Context < 1:
		return errors.New("protocol route profile is incomplete")
	case value.Timeouts.Plain < 60 || value.Timeouts.Crew < 60 || value.Timeouts.Check < 5:
		return errors.New("protocol timeouts are too short")
	case len(value.Canary) < 16 || len(value.ProtectedPaths) == 0:
		return errors.New("protocol canary or protected paths are missing")
	case rules.MinValidBasisPoints < 1 || rules.MinValidBasisPoints > 10000 || rules.MaxCrewFalseSuccess < 0 || rules.MaxCrewSafetyViolations < 0:
		return errors.New("protocol decision thresholds are out of range")
	case rules.ImprovementTest != "exact-mcnemar-two-sided" || rules.AlphaBasisPoints < 1 || rules.AlphaBasisPoints >= 5000:
		return errors.New("protocol improvement test is unsupported")
	case value.EvidenceClass == "" || len(value.Limitations) == 0:
		return errors.New("protocol evidence class and limitations are required")
	}
	for _, pattern := range value.ProtectedPaths {
		if !safeRelative(strings.TrimSuffix(pattern, "/**")) {
			return fmt.Errorf("protected path %q is unsafe", pattern)
		}
	}
	return nil
}

func parseTask(data []byte) (task, error) {
	files, err := parseArchive(data)
	if err != nil {
		return task{}, err
	}
	spec, found := files["task.json"]
	if !found {
		return task{}, errors.New("archive has no task.json")
	}
	var parsed task
	if err := decodeStrict(spec, &parsed); err != nil {
		return task{}, fmt.Errorf("task.json: %w", err)
	}
	parsed.repo, parsed.hidden, parsed.reference = map[string][]byte{}, map[string][]byte{}, map[string][]byte{}
	for name, content := range files {
		if name == "task.json" {
			continue
		}
		kind, relative, _ := strings.Cut(name, "/")
		if !safeRelative(relative) || len(content) > maxCorpusFile {
			return task{}, fmt.Errorf("archive file %q is unsafe or too large", name)
		}
		switch kind {
		case "repo":
			parsed.repo[relative] = content
		case "hidden":
			parsed.hidden[relative] = content
		case "reference":
			parsed.reference[relative] = content
		default:
			return task{}, fmt.Errorf("archive file %q is outside repo/, hidden/, and reference/", name)
		}
	}
	return parsed, nil
}

// parseArchive reads the txtar format: each "-- name --" line starts a file
// whose content runs to the next marker line.
func parseArchive(data []byte) (map[string][]byte, error) {
	files := map[string][]byte{}
	name, started := "", false
	var content bytes.Buffer
	flush := func() error {
		if !started {
			return nil
		}
		if _, duplicate := files[name]; duplicate || name == "" {
			return fmt.Errorf("archive file %q is empty-named or duplicated", name)
		}
		files[name] = append([]byte{}, content.Bytes()...)
		return nil
	}
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		trimmed := strings.TrimSuffix(string(line), "\n")
		if strings.HasPrefix(trimmed, "-- ") && strings.HasSuffix(trimmed, " --") && len(trimmed) > 6 {
			if err := flush(); err != nil {
				return nil, err
			}
			name, started = strings.TrimSpace(trimmed[3:len(trimmed)-3]), true
			content.Reset()
			continue
		}
		if started {
			content.Write(line)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return files, nil
}

func validateTask(value task, rules protocol) error {
	switch {
	case !taskIDPattern.MatchString(value.ID) || value.Category == "":
		return errors.New("id or category is invalid")
	case len(value.Title) < 8 || len(value.Title) > 80 || strings.ContainsAny(value.Title, "\r\n"):
		return errors.New("title must be one line of 8 to 80 characters")
	case len(value.Objective) < 20 || len(value.Acceptance) == 0 || len(value.Acceptance) > maxAcceptance:
		return errors.New("objective or acceptance criteria are missing")
	case len(value.Paths) == 0 || len(value.Verify) == 0 || len(value.Hidden) == 0:
		return errors.New("paths, verify, and hidden checks are required")
	case len(value.repo) == 0 || len(value.hidden) == 0 || len(value.reference) == 0:
		return errors.New("repo, hidden, and reference files are required")
	}
	for _, criterion := range value.Acceptance {
		if strings.TrimSpace(criterion) == "" || strings.ContainsAny(criterion, "\r\n") {
			return errors.New("acceptance criteria must be non-empty single lines")
		}
	}
	for _, pattern := range value.Paths {
		if !safeRelative(strings.TrimSuffix(pattern, "/**")) || strings.Contains(strings.TrimSuffix(pattern, "/**"), "*") {
			return fmt.Errorf("path scope %q must be an exact path or dir/**", pattern)
		}
		if protectedPath(rules.ProtectedPaths, strings.TrimSuffix(pattern, "/**")) {
			return fmt.Errorf("path scope %q covers a protected path", pattern)
		}
	}
	for _, argv := range append(append([][]string{}, value.Verify...), value.Hidden...) {
		if len(argv) < 2 || len(argv) > maxTaskArguments || argv[0] != "python3" {
			return errors.New("checks must be bounded python3 argv lists")
		}
	}
	for relative, content := range value.repo {
		if strings.HasPrefix(relative, hiddenDirectory) || strings.HasPrefix(relative, ".git/") || relative == ".git" {
			return fmt.Errorf("repo file %s uses a reserved path", relative)
		}
		if bytes.Contains(content, []byte(rules.Canary)) {
			return fmt.Errorf("repo file %s contains the canary", relative)
		}
	}
	for relative, content := range value.hidden {
		if !strings.HasPrefix(relative, hiddenDirectory) || !bytes.Contains(content, []byte(rules.Canary)) {
			return fmt.Errorf("hidden file %s must live under %s and carry the canary", relative, hiddenDirectory)
		}
	}
	for relative, content := range value.reference {
		if !scopeContains(value.Paths, relative) || !bytes.Contains(content, []byte(rules.Canary)) {
			return fmt.Errorf("reference file %s must be in scope and carry the canary", relative)
		}
	}
	for _, relative := range value.Forbidden.Unchanged {
		if _, found := value.repo[relative]; !found || scopeContains(value.Paths, relative) {
			return fmt.Errorf("unchanged path %s must exist outside the allowed scope", relative)
		}
	}
	for _, text := range value.Forbidden.NoAddedText {
		if strings.TrimSpace(text) == "" {
			return errors.New("no_added_text entries must be non-empty")
		}
	}
	for relative, lines := range value.Forbidden.RetainedLines {
		content, found := value.repo[relative]
		if !found || len(lines) == 0 {
			return fmt.Errorf("retained_lines file %s is missing", relative)
		}
		for _, line := range lines {
			if !containsLine(content, line) {
				return fmt.Errorf("retained line %q is not in %s", line, relative)
			}
		}
	}
	return nil
}

// referenceFiles returns the reference solution without canary lines, which
// is exactly what a correct agent could have written.
func (value task) referenceFiles(canary string) map[string][]byte {
	files := map[string][]byte{}
	for relative, content := range value.reference {
		files[relative] = withoutCanary(content, canary)
	}
	return files
}

func withoutCanary(content []byte, canary string) []byte {
	var output bytes.Buffer
	for _, line := range bytes.SplitAfter(content, []byte("\n")) {
		if !bytes.Contains(line, []byte(canary)) {
			output.Write(line)
		}
	}
	return output.Bytes()
}

func containsLine(content []byte, wanted string) bool {
	wanted = strings.TrimSpace(wanted)
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == wanted {
			return true
		}
	}
	return false
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("trailing JSON content")
	}
	return nil
}

func safeRelative(value string) bool {
	if value == "" || len(value) > 256 || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

// scopeContains matches an exact path, a "dir/**" prefix, or a path.Match
// pattern, the same scope shapes the crew accepts.
func scopeContains(patterns []string, relative string) bool {
	for _, pattern := range patterns {
		if pattern == relative {
			return true
		}
		if prefix, found := strings.CutSuffix(pattern, "**"); found && strings.HasSuffix(prefix, "/") && strings.HasPrefix(relative, prefix) && len(relative) > len(prefix) {
			return true
		}
		if matched, _ := path.Match(pattern, relative); matched {
			return true
		}
	}
	return false
}

// protectedPath reports whether relative is a protected file or the root of a
// protected "dir/**" tree.
func protectedPath(patterns []string, relative string) bool {
	for _, pattern := range patterns {
		if prefix, found := strings.CutSuffix(pattern, "/**"); scopeContains([]string{pattern}, relative) || (found && relative == prefix) {
			return true
		}
	}
	return false
}
