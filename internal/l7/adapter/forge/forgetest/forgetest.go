// Package forgetest is a stand-in gh CLI for tests. A test binary calls Main
// from TestMain when it is invoked with gh arguments; state lives in a JSON
// file inside the repository's .git directory.
package forgetest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

// Commands lists the first arguments that route a test binary to Main.
var Commands = map[string]bool{"pr": true, "auth": true, "run": true}

type PullRequest struct {
	Number            int              `json:"number"`
	URL               string           `json:"url"`
	State             string           `json:"state"`
	IsDraft           bool             `json:"isDraft"`
	HeadRefOid        string           `json:"headRefOid"`
	HeadRefName       string           `json:"headRefName"`
	BaseRefName       string           `json:"baseRefName"`
	MergeStateStatus  string           `json:"mergeStateStatus"`
	MergeCommit       map[string]any   `json:"mergeCommit,omitempty"`
	StatusCheckRollup []map[string]any `json:"statusCheckRollup"`
	Title             string           `json:"title"`
	Body              string           `json:"body"`
	Labels            []string         `json:"labels"`
}

type Merge struct {
	Number int    `json:"number"`
	Method string `json:"method"`
	Head   string `json:"head"`
}

type State struct {
	Unauthenticated bool              `json:"unauthenticated"`
	Heads           map[string]string `json:"heads"`
	PullRequests    []PullRequest     `json:"pull_requests"`
	Merges          []Merge           `json:"merges"`
	Logs            map[string]string `json:"logs"`
	Calls           [][]string        `json:"calls"`
}

func StatePath(root string) string { return filepath.Join(root, ".git", "fake-gh.json") }

func Load(root string) (State, error) {
	state := State{Heads: map[string]string{}, Logs: map[string]string{}}
	data, err := os.ReadFile(StatePath(root))
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func Save(root string, state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(root), data, 0o600)
}

// Main runs one gh invocation against the state in the current directory.
func Main(arguments []string, stdout, stderr io.Writer) int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	state, err := Load(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	state.Calls = append(state.Calls, arguments)
	code := handle(&state, arguments, stdout, stderr)
	if err := Save(root, state); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return code
}

func handle(state *State, arguments []string, stdout, stderr io.Writer) int {
	command := strings.Join(arguments[:min(2, len(arguments))], " ")
	flags := map[string]string{}
	for index := 2; index < len(arguments); index++ {
		if strings.HasPrefix(arguments[index], "--") && index+1 < len(arguments) && !strings.HasPrefix(arguments[index+1], "--") {
			flags[arguments[index]] = arguments[index+1]
			index++
		} else {
			flags[arguments[index]] = ""
		}
	}
	switch command {
	case "auth status":
		if state.Unauthenticated {
			fmt.Fprintln(stderr, "You are not logged into any GitHub hosts.")
			return 1
		}
		return 0
	case "pr list":
		listed := []PullRequest{}
		for _, pull := range state.PullRequests {
			if pull.HeadRefName == flags["--head"] {
				listed = append(listed, pull)
			}
		}
		return write(stdout, listed)
	case "pr create":
		head := state.Heads[flags["--head"]]
		if head == "" {
			head = remoteHead(flags["--head"])
		}
		if head == "" {
			fmt.Fprintln(stderr, "pull request create failed: head branch was not pushed")
			return 1
		}
		number := len(state.PullRequests) + 1
		state.PullRequests = append(state.PullRequests, PullRequest{
			Number: number, URL: fmt.Sprintf("https://github.com/owner/repo/pull/%d", number), State: "OPEN",
			HeadRefOid: head, HeadRefName: flags["--head"], BaseRefName: flags["--base"], MergeStateStatus: "BLOCKED",
			StatusCheckRollup: []map[string]any{}, Title: flags["--title"], Body: flags["--body"], Labels: []string{},
		})
		fmt.Fprintf(stdout, "https://github.com/owner/repo/pull/%d\n", number)
		return 0
	case "pr view", "pr edit", "pr merge":
		number, _ := strconv.Atoi(arguments[2])
		pull := find(state, number)
		if pull == nil {
			fmt.Fprintln(stderr, "no pull requests found")
			return 1
		}
		switch arguments[1] {
		case "view":
			return write(stdout, pull)
		case "edit":
			pull.Labels = append(pull.Labels, flags["--add-label"])
			return 0
		}
		if _, ok := flags["--admin"]; ok {
			fmt.Fprintln(stderr, "admin merges are not allowed in tests")
			return 1
		}
		if flags["--match-head-commit"] != pull.HeadRefOid {
			fmt.Fprintln(stderr, "head commit does not match")
			return 1
		}
		method := ""
		for _, candidate := range []string{"merge", "squash", "rebase"} {
			if _, ok := flags["--"+candidate]; ok {
				method = candidate
			}
		}
		pull.State, pull.MergeCommit = "MERGED", map[string]any{"oid": strings.Repeat("f", 40)}
		state.Merges = append(state.Merges, Merge{Number: number, Method: method, Head: pull.HeadRefOid})
		return 0
	case "run view":
		fmt.Fprint(stdout, state.Logs[flags["--job"]])
		return 0
	default:
		fmt.Fprintf(stderr, "unsupported fake gh command %q\n", command)
		return 1
	}
}

// remoteHead reads branch from the origin remote of the current repository,
// as GitHub would see the pushed branch.
func remoteHead(branch string) string {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return ""
	}
	directory, err := os.Getwd()
	if err != nil {
		return ""
	}
	result, err := (processadapter.Runner{}).Run(context.Background(), processadapter.Request{
		Executable: git.Path, Arguments: []string{"ls-remote", "origin", "refs/heads/" + branch}, Directory: directory,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second,
	})
	if err != nil || result.ExitCode != 0 {
		return ""
	}
	head, _, _ := strings.Cut(string(result.Stdout), "\t")
	return strings.TrimSpace(head)
}

func find(state *State, number int) *PullRequest {
	for index := range state.PullRequests {
		if state.PullRequests[index].Number == number {
			return &state.PullRequests[index]
		}
	}
	return nil
}

func write(stdout io.Writer, value any) int {
	if err := json.NewEncoder(stdout).Encode(value); err != nil {
		return 1
	}
	return 0
}
