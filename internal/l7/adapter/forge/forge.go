// Package forge drives GitHub pull requests through the owner's own gh CLI.
// Level 7 passes no token and stores none; gh uses the owner's login.
package forge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	maxOutputBytes = 8 << 20
	maxTitleBytes  = 256
	maxBodyBytes   = 60_000
	commandTimeout = 90 * time.Second
)

var (
	objectID   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	jobURL     = regexp.MustCompile(`/actions/runs/[0-9]+/job/([0-9]+)`)
	pullNumber = regexp.MustCompile(`/pull/([0-9]+)$`)
	segment    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
)

// Client acts on one repository, named explicitly so gh never guesses
// between remotes.
type Client struct {
	executable string
	root       string
	repository string
}

type Check struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DetailsURL string `json:"details_url,omitempty"`
}

type PullRequest struct {
	Number      int     `json:"number"`
	URL         string  `json:"url"`
	State       string  `json:"state"`
	Draft       bool    `json:"draft"`
	Head        string  `json:"head"`
	HeadBranch  string  `json:"head_branch"`
	Base        string  `json:"base"`
	MergeState  string  `json:"merge_state"`
	MergeCommit string  `json:"merge_commit,omitempty"`
	Checks      []Check `json:"checks"`
}

// Discover finds gh on PATH for repository, given as HOST/OWNER/REPO.
func Discover(root, repository string) (Client, error) {
	executable, err := processadapter.Resolve("gh")
	if err != nil {
		return Client{}, errors.New("the gh CLI is not installed; pull-request delivery needs it")
	}
	return New(executable.Path, root, repository)
}

func New(executable, root, repository string) (Client, error) {
	if !filepath.IsAbs(executable) || !filepath.IsAbs(root) {
		return Client{}, errors.New("gh executable and repository root must be absolute")
	}
	if parts := strings.Split(repository, "/"); len(parts) != 3 || !segment.MatchString(parts[0]) || !segment.MatchString(parts[1]) || !segment.MatchString(parts[2]) {
		return Client{}, errors.New("forge repository must be HOST/OWNER/REPO")
	}
	resolved, err := processadapter.Resolve(executable)
	if err != nil || resolved.Path != executable {
		return Client{}, errors.New("gh executable identity is unavailable")
	}
	return Client{executable: executable, root: root, repository: repository}, nil
}

// RepositoryFromURL turns a remote URL such as https://github.com/o/r.git,
// git@github.com:o/r.git, or ssh://git@github.com/o/r into github.com/o/r.
func RepositoryFromURL(value string) (string, error) {
	path := ""
	switch {
	case strings.HasPrefix(value, "https://"):
		path = strings.TrimPrefix(value, "https://")
	case strings.HasPrefix(value, "ssh://"):
		path = strings.TrimPrefix(value, "ssh://")
	case strings.Contains(value, "@") && strings.Contains(value, ":") && !strings.Contains(value, "://"):
		path = strings.Replace(value, ":", "/", 1)
	default:
		return "", errors.New("the remote is not an https or ssh forge URL")
	}
	if _, rest, found := strings.Cut(path, "@"); found {
		path = rest
	}
	parts := strings.Split(strings.TrimSuffix(strings.TrimSuffix(path, "/"), ".git"), "/")
	if len(parts) != 3 {
		return "", errors.New("the remote URL does not name one host, owner, and repository")
	}
	host, _, _ := strings.Cut(parts[0], ":")
	repository := host + "/" + parts[1] + "/" + parts[2]
	for _, part := range []string{host, parts[1], parts[2]} {
		if !segment.MatchString(part) {
			return "", errors.New("the remote URL has an unsafe host, owner, or repository name")
		}
	}
	return repository, nil
}

// Authenticated reports whether gh can act for the owner.
func (client Client) Authenticated(ctx context.Context) error {
	if _, err := client.run(ctx, "auth", "status"); err != nil {
		return errors.New("gh is not authenticated; run gh auth login")
	}
	return nil
}

// FindByHead lists pull requests in any state whose head is branch.
func (client Client) FindByHead(ctx context.Context, branch string) ([]PullRequest, error) {
	if !domain.CrewBranchValid(branch) {
		return nil, errors.New("pull-request head branch is unsafe")
	}
	data, err := client.run(ctx, "pr", "list", "--head", branch, "--state", "all", "--limit", "20", "--json", "number,url,state,isDraft,headRefOid,headRefName,baseRefName")
	if err != nil {
		return nil, err
	}
	var listed []ghPullRequest
	if err := decode(data, &listed); err != nil {
		return nil, err
	}
	pulls := make([]PullRequest, 0, len(listed))
	for _, pull := range listed {
		if pull.HeadRefName == branch {
			pulls = append(pulls, pull.normalize())
		}
	}
	return pulls, nil
}

// Create opens a pull request from head into base and returns it.
func (client Client) Create(ctx context.Context, base, head, title, body string) (PullRequest, error) {
	if !domain.CrewBranchValid(base) || !domain.CrewBranchValid(head) || !text(title, maxTitleBytes) || !utf8.ValidString(body) || body == "" || len(body) > maxBodyBytes || strings.ContainsRune(body, 0) {
		return PullRequest{}, errors.New("pull request base, head, title, or body is invalid")
	}
	data, err := client.run(ctx, "pr", "create", "--base", base, "--head", head, "--title", title, "--body", body)
	if err != nil {
		return PullRequest{}, err
	}
	lines := strings.Fields(strings.TrimSpace(string(data)))
	if len(lines) == 0 {
		return PullRequest{}, errors.New("gh did not report the new pull request")
	}
	match := pullNumber.FindStringSubmatch(lines[len(lines)-1])
	if match == nil {
		return PullRequest{}, errors.New("gh reported an unexpected pull request URL")
	}
	number, _ := strconv.Atoi(match[1])
	return client.View(ctx, number)
}

// View reads one pull request with its check results.
func (client Client) View(ctx context.Context, number int) (PullRequest, error) {
	if number < 1 {
		return PullRequest{}, errors.New("pull request number is invalid")
	}
	data, err := client.run(ctx, "pr", "view", strconv.Itoa(number), "--json", "number,url,state,isDraft,headRefOid,headRefName,baseRefName,mergeStateStatus,mergeCommit,statusCheckRollup")
	if err != nil {
		return PullRequest{}, err
	}
	var pull ghPullRequest
	if err := decode(data, &pull); err != nil {
		return PullRequest{}, err
	}
	if pull.Number != number {
		return PullRequest{}, errors.New("gh returned a different pull request")
	}
	return pull.normalize(), nil
}

func (client Client) AddLabel(ctx context.Context, number int, label string) error {
	if number < 1 || !text(label, 64) {
		return errors.New("pull request label is invalid")
	}
	_, err := client.run(ctx, "pr", "edit", strconv.Itoa(number), "--add-label", label)
	return err
}

// Merge merges one pull request with method, only while its head is head.
// It never bypasses branch protection or enables auto-merge.
func (client Client) Merge(ctx context.Context, number int, method, head string) error {
	if number < 1 || !objectID.MatchString(head) || (method != "merge" && method != "squash" && method != "rebase") {
		return errors.New("pull request merge request is invalid")
	}
	_, err := client.run(ctx, "pr", "merge", strconv.Itoa(number), "--"+method, "--match-head-commit", head)
	return err
}

// FailedLog returns the end of a failed GitHub Actions job log, or "" when
// the check is not an Actions job or its log is unavailable.
func (client Client) FailedLog(ctx context.Context, check Check, limit int) string {
	match := jobURL.FindStringSubmatch(check.DetailsURL)
	if match == nil || limit < 1 {
		return ""
	}
	data, err := client.run(ctx, "run", "view", "--job", match[1], "--log-failed")
	if err != nil {
		return ""
	}
	log := strings.ToValidUTF8(string(data), "")
	log = strings.Map(func(character rune) rune {
		if character == '\n' || character == '\t' {
			return character
		}
		if character < 0x20 || character == 0x7f {
			return -1
		}
		return character
	}, log)
	if len(log) > limit {
		cut := len(log) - limit
		for cut < len(log) && !utf8.RuneStart(log[cut]) {
			cut++
		}
		log = log[cut:]
	}
	return strings.TrimSpace(log)
}

// Summarize reduces checks to passed, failed, pending, or none, with the
// names of failing checks. Skipped checks never ran, so checks that were all
// skipped summarize as none.
func Summarize(checks []Check) (string, []string) {
	failing, pending, passed := []string{}, 0, 0
	for _, check := range checks {
		switch check.Status {
		case "failed":
			failing = append(failing, check.Name)
		case "pending":
			pending++
		case "passed":
			passed++
		}
	}
	switch {
	case len(failing) != 0:
		return "failed", failing
	case pending != 0:
		return "pending", nil
	case passed == 0:
		return "none", nil
	default:
		return "passed", nil
	}
}

type ghCheck struct {
	Typename   string `json:"__typename"`
	Name       string `json:"name"`
	Context    string `json:"context"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	State      string `json:"state"`
	DetailsURL string `json:"detailsUrl"`
	TargetURL  string `json:"targetUrl"`
}

type ghPullRequest struct {
	Number           int    `json:"number"`
	URL              string `json:"url"`
	State            string `json:"state"`
	IsDraft          bool   `json:"isDraft"`
	HeadRefOid       string `json:"headRefOid"`
	HeadRefName      string `json:"headRefName"`
	BaseRefName      string `json:"baseRefName"`
	MergeStateStatus string `json:"mergeStateStatus"`
	MergeCommit      *struct {
		Oid string `json:"oid"`
	} `json:"mergeCommit"`
	StatusCheckRollup []ghCheck `json:"statusCheckRollup"`
}

func (pull ghPullRequest) normalize() PullRequest {
	normalized := PullRequest{
		Number: pull.Number, URL: pull.URL, State: pull.State, Draft: pull.IsDraft, Head: pull.HeadRefOid,
		HeadBranch: pull.HeadRefName, Base: pull.BaseRefName, MergeState: pull.MergeStateStatus, Checks: []Check{},
	}
	if pull.MergeCommit != nil {
		normalized.MergeCommit = pull.MergeCommit.Oid
	}
	for _, check := range pull.StatusCheckRollup {
		normalized.Checks = append(normalized.Checks, check.normalize())
	}
	return normalized
}

func (check ghCheck) normalize() Check {
	if check.Typename == "StatusContext" || (check.Name == "" && check.Context != "") {
		status := "failed"
		switch check.State {
		case "SUCCESS":
			status = "passed"
		case "PENDING", "EXPECTED":
			status = "pending"
		}
		return Check{Name: check.Context, Status: status, DetailsURL: check.TargetURL}
	}
	status := "failed"
	switch {
	case check.Status != "COMPLETED":
		status = "pending"
	case check.Conclusion == "SUCCESS" || check.Conclusion == "NEUTRAL":
		status = "passed"
	case check.Conclusion == "SKIPPED":
		status = "skipped"
	}
	return Check{Name: check.Name, Status: status, DetailsURL: check.DetailsURL}
}

func (client Client) run(ctx context.Context, arguments ...string) ([]byte, error) {
	if arguments[0] != "auth" {
		arguments = append(arguments, "--repo", client.repository)
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{
		Executable: client.executable, Arguments: arguments, Directory: client.root,
		Environment: environment(), MaxOutputBytes: maxOutputBytes, Timeout: commandTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("gh %s did not complete: %w", arguments[0], err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("gh %s %s failed: %s", arguments[0], arguments[1], diagnostic(result.Stderr))
	}
	return result.Stdout, nil
}

// environment is the minimal process environment plus the owner's own gh
// settings, so gh authenticates the way it does in the owner's shell.
func environment() []string {
	values := processadapter.MinimalEnvironment()
	for _, key := range []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GH_HOST", "GH_CONFIG_DIR", "XDG_CONFIG_HOME"} {
		if value := os.Getenv(key); value != "" && !strings.ContainsAny(value, "\x00\r\n") {
			values = append(values, key+"="+value)
		}
	}
	return append(values, "GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "GH_PAGER=cat")
}

func decode(data []byte, target any) error {
	if err := json.Unmarshal(data, target); err != nil {
		return errors.New("gh returned malformed JSON")
	}
	return nil
}

func diagnostic(stderr []byte) string {
	message := strings.Join(strings.Fields(strings.ToValidUTF8(string(stderr), "")), " ")
	if len(message) > 512 {
		cut := 512
		for cut > 0 && !utf8.RuneStart(message[cut]) {
			cut--
		}
		message = message[:cut]
	}
	if message == "" {
		return "no diagnostic"
	}
	return message
}

func text(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && utf8.ValidString(value) && !strings.ContainsAny(value, "\x00\r\n")
}

// PullURLValid accepts an https pull request URL.
func PullURLValid(value string) bool {
	rest, secure := strings.CutPrefix(value, "https://")
	host, path, found := strings.Cut(rest, "/")
	return secure && found && host != "" && len(value) <= 2048 && !strings.ContainsAny(value, " \t\r\n\x00?#") && pullNumber.MatchString("/"+path)
}
