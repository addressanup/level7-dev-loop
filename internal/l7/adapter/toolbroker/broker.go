// Package toolbroker exposes the bounded repository tools available to gateway workers.
package toolbroker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

const (
	maxToolInput  = 1 << 20
	maxListFiles  = 100_000
	maxSearchHits = 512
)

type MemoryQueryFunc func(string) ([]byte, error)

type Broker struct {
	root        string
	policy      orchestrationconfig.Tools
	resolve     func(string) (processadapter.Executable, error)
	run         func(context.Context, processadapter.Request) (processadapter.Result, error)
	memoryQuery MemoryQueryFunc
	readOnly    bool
}

// NewReadOnly composes the same bounded broker while rejecting every patch
// request. It is used for independent reviewers.
func NewReadOnly(root string, policy orchestrationconfig.Tools, memoryQuery MemoryQueryFunc) (Broker, error) {
	broker, err := New(root, policy, memoryQuery)
	broker.readOnly = true
	return broker, err
}

type ToolResult struct {
	OK       bool     `json:"ok"`
	Message  string   `json:"message"`
	Paths    []string `json:"paths"`
	Content  string   `json:"content"`
	ExitCode int      `json:"exit_code"`
}

func New(root string, policy orchestrationconfig.Tools, memoryQuery MemoryQueryFunc) (Broker, error) {
	return NewWith(root, policy, memoryQuery, processadapter.Resolve, (processadapter.Runner{}).Run)
}

func NewWith(root string, policy orchestrationconfig.Tools, memoryQuery MemoryQueryFunc, resolve func(string) (processadapter.Executable, error), run func(context.Context, processadapter.Request) (processadapter.Result, error)) (Broker, error) {
	physical, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(physical) || physical != filepath.Clean(root) {
		return Broker{}, errors.New("tool broker root must be an existing physical absolute path")
	}
	info, err := os.Stat(physical)
	if err != nil || !info.IsDir() {
		return Broker{}, errors.New("tool broker root is not a directory")
	}
	if resolve == nil || run == nil || policy.MaxOutputBytes < 64<<10 || policy.MaxOutputBytes > 64<<20 || policy.MaxSeconds < 1 || policy.MaxSeconds > 86_400 {
		return Broker{}, errors.New("tool broker policy is invalid")
	}
	return Broker{root: physical, policy: policy, resolve: resolve, run: run, memoryQuery: memoryQuery}, nil
}

func (broker Broker) Call(ctx context.Context, name string, arguments []byte) ([]byte, error) {
	if len(arguments) > maxToolInput || !json.Valid(arguments) {
		return nil, errors.New("tool arguments are invalid or unbounded")
	}
	var result ToolResult
	var err error
	switch name {
	case "list_files":
		var input pathInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.listFiles(input.Path)
		}
	case "read_file":
		var input readInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.readFile(input)
		}
	case "search":
		var input searchInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.search(input)
		}
	case "apply_patch":
		if broker.readOnly {
			err = errors.New("read-only reviewer cannot apply patches")
			break
		}
		var input patchInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.applyPatch(ctx, input.Patch)
		}
	case "git_status":
		var input emptyInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.git(ctx, []string{"status", "--short", "--untracked-files=all"})
		}
	case "git_diff":
		var input emptyInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.git(ctx, []string{"diff", "--no-ext-diff", "--no-textconv", "--"})
		}
	case "run_command":
		if broker.readOnly {
			err = errors.New("read-only reviewer cannot run configured processes")
			break
		}
		var input commandInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.runCommand(ctx, input.Name)
		}
	case "memory_query":
		var input queryInput
		if err = decode(arguments, &input); err == nil {
			result, err = broker.queryMemory(input.Query)
		}
	default:
		err = errors.New("unknown broker tool")
	}
	if err != nil {
		return encode(ToolResult{OK: false, Message: bounded(err.Error(), 512), Paths: []string{}})
	}
	return encode(result)
}

type emptyInput struct{}
type pathInput struct {
	Path string `json:"path"`
}
type readInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}
type searchInput struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}
type patchInput struct {
	Patch string `json:"patch"`
}
type commandInput struct {
	Name string `json:"name"`
}
type queryInput struct {
	Query string `json:"query"`
}

func (broker Broker) listFiles(relative string) (ToolResult, error) {
	base, relative, err := broker.resolvePath(relative, false)
	if err != nil {
		return ToolResult{}, err
	}
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		return ToolResult{}, errors.New("list path is not a directory")
	}
	paths := []string{}
	err = filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(broker.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if rel == ".git" || strings.HasPrefix(rel, ".git/") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() || secretPath(rel) || !broker.allowed(rel) {
			return nil
		}
		paths = append(paths, rel)
		if len(paths) > maxListFiles {
			return errors.New("file listing exceeds limit")
		}
		return nil
	})
	if err != nil {
		return ToolResult{}, err
	}
	sort.Strings(paths)
	return ToolResult{OK: true, Message: "bounded file listing", Paths: paths, Content: relative}, nil
}

func (broker Broker) readFile(input readInput) (ToolResult, error) {
	path, relative, err := broker.resolvePath(input.Path, true)
	if err != nil {
		return ToolResult{}, err
	}
	if input.Offset < 0 || input.Limit < 1 || input.Limit > broker.policy.MaxOutputBytes {
		return ToolResult{}, errors.New("read bounds are invalid")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{}, errors.New("cannot read requested file")
	}
	if input.Offset > len(data) {
		return ToolResult{}, errors.New("read offset exceeds file")
	}
	end := input.Offset + input.Limit
	if end > len(data) {
		end = len(data)
	}
	return ToolResult{OK: true, Message: "bounded file content", Paths: []string{relative}, Content: string(data[input.Offset:end])}, nil
}

func (broker Broker) search(input searchInput) (ToolResult, error) {
	if input.Query == "" || len(input.Query) > 4096 || strings.ContainsRune(input.Query, 0) {
		return ToolResult{}, errors.New("search query is invalid")
	}
	listing, err := broker.listFiles(input.Path)
	if err != nil {
		return ToolResult{}, err
	}
	var output strings.Builder
	hits := 0
	for _, relative := range listing.Paths {
		path := filepath.Join(broker.root, filepath.FromSlash(relative))
		data, err := os.ReadFile(path)
		if err != nil || bytes.IndexByte(data, 0) >= 0 {
			continue
		}
		for index, line := range strings.Split(string(data), "\n") {
			if !strings.Contains(line, input.Query) {
				continue
			}
			fmt.Fprintf(&output, "%s:%d:%s\n", relative, index+1, bounded(line, 2048))
			hits++
			if hits >= maxSearchHits || output.Len() >= broker.policy.MaxOutputBytes {
				return ToolResult{OK: true, Message: "search result truncated at bounds", Paths: []string{}, Content: bounded(output.String(), broker.policy.MaxOutputBytes)}, nil
			}
		}
	}
	return ToolResult{OK: true, Message: "bounded search results", Paths: []string{}, Content: output.String()}, nil
}

func (broker Broker) applyPatch(ctx context.Context, patch string) (ToolResult, error) {
	if patch == "" || len(patch) > maxToolInput || strings.ContainsRune(patch, 0) {
		return ToolResult{}, errors.New("patch is empty or unbounded")
	}
	paths, err := patchPaths(patch)
	if err != nil {
		return ToolResult{}, err
	}
	for _, relative := range paths {
		if !broker.allowed(relative) || secretPath(relative) {
			return ToolResult{}, fmt.Errorf("patch path %q is outside the manifest", relative)
		}
		if _, _, err := broker.resolvePath(relative, false); err != nil && !errors.Is(err, os.ErrNotExist) {
			return ToolResult{}, err
		}
	}
	git, err := broker.resolve("git")
	if err != nil {
		return ToolResult{}, errors.New("Git is unavailable")
	}
	for _, arguments := range [][]string{{"apply", "--check", "--whitespace=error-all", "-"}, {"apply", "--whitespace=error-all", "-"}} {
		result, runErr := broker.run(ctx, processadapter.Request{
			Executable: git.Path, Arguments: arguments, Input: []byte(patch), Directory: broker.root,
			Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: broker.policy.MaxOutputBytes, Timeout: time.Duration(broker.policy.MaxSeconds) * time.Second,
		})
		if runErr != nil || result.ExitCode != 0 {
			return ToolResult{}, errors.New("Git rejected the bounded patch")
		}
	}
	return ToolResult{OK: true, Message: "patch applied", Paths: paths}, nil
}

func (broker Broker) git(ctx context.Context, arguments []string) (ToolResult, error) {
	git, err := broker.resolve("git")
	if err != nil {
		return ToolResult{}, errors.New("Git is unavailable")
	}
	result, runErr := broker.run(ctx, processadapter.Request{
		Executable: git.Path, Arguments: arguments, Directory: broker.root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: broker.policy.MaxOutputBytes, Timeout: time.Duration(broker.policy.MaxSeconds) * time.Second,
	})
	if runErr != nil || result.ExitCode != 0 {
		return ToolResult{}, errors.New("Git tool failed")
	}
	return ToolResult{OK: true, Message: "bounded Git output", Paths: []string{}, Content: string(result.Stdout), ExitCode: result.ExitCode}, nil
}

func (broker Broker) runCommand(ctx context.Context, name string) (ToolResult, error) {
	var command orchestrationconfig.Command
	found := false
	for _, candidate := range broker.policy.AllowedCommands {
		if candidate.Name == name {
			command, found = candidate, true
			break
		}
	}
	if !found || len(command.Argv) == 0 {
		return ToolResult{}, errors.New("command is not allowlisted")
	}
	executable, err := broker.resolve(command.Argv[0])
	if err != nil {
		return ToolResult{}, errors.New("allowlisted executable is unavailable")
	}
	result, runErr := broker.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: append([]string{}, command.Argv[1:]...), Directory: broker.root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: broker.policy.MaxOutputBytes, Timeout: time.Duration(broker.policy.MaxSeconds) * time.Second,
	})
	if runErr != nil {
		return ToolResult{}, errors.New("allowlisted command failed or exceeded bounds")
	}
	content := string(append(append([]byte{}, result.Stdout...), result.Stderr...))
	return ToolResult{OK: result.ExitCode == 0, Message: "allowlisted command completed", Paths: []string{}, Content: content, ExitCode: result.ExitCode}, nil
}

func (broker Broker) queryMemory(query string) (ToolResult, error) {
	if broker.memoryQuery == nil || query == "" || len(query) > 4096 || strings.ContainsRune(query, 0) {
		return ToolResult{}, errors.New("memory query is unavailable or invalid")
	}
	data, err := broker.memoryQuery(query)
	if err != nil {
		return ToolResult{}, errors.New("memory query failed")
	}
	if len(data) > broker.policy.MaxOutputBytes {
		data = data[:broker.policy.MaxOutputBytes]
	}
	return ToolResult{OK: true, Message: "bounded memory matches", Paths: []string{}, Content: string(data)}, nil
}

func (broker Broker) resolvePath(relative string, requireRegular bool) (string, string, error) {
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == "." {
		relative = ""
	}
	if filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, "../") || strings.Contains(relative, "\\") || strings.HasPrefix(relative, ".git/") || relative == ".git" || secretPath(relative) || (relative != "" && !broker.allowed(relative)) {
		return "", "", errors.New("path is outside the bounded repository scope")
	}
	candidate := filepath.Join(broker.root, filepath.FromSlash(relative))
	if candidate != broker.root && !strings.HasPrefix(candidate, broker.root+string(filepath.Separator)) {
		return "", "", errors.New("path escapes repository root")
	}
	physical, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			parent, parentErr := filepath.EvalSymlinks(filepath.Dir(candidate))
			if parentErr != nil || (parent != broker.root && !strings.HasPrefix(parent, broker.root+string(filepath.Separator))) {
				return "", "", errors.New("new path parent is unsafe")
			}
			return candidate, relative, os.ErrNotExist
		}
		return "", "", errors.New("cannot resolve path safely")
	}
	if physical != broker.root && !strings.HasPrefix(physical, broker.root+string(filepath.Separator)) {
		return "", "", errors.New("resolved path escapes repository root")
	}
	if requireRegular {
		info, err := os.Stat(physical)
		if err != nil || !info.Mode().IsRegular() || info.Size() > int64(broker.policy.MaxOutputBytes)*64 {
			return "", "", errors.New("path is not a bounded regular file")
		}
	}
	return physical, relative, nil
}

func (broker Broker) allowed(relative string) bool {
	for _, declared := range broker.policy.AllowedPaths {
		if declared == relative {
			return true
		}
		if strings.HasSuffix(declared, "/**") {
			prefix := strings.TrimSuffix(declared, "**")
			if strings.HasPrefix(relative, prefix) && len(relative) > len(prefix) {
				return true
			}
		}
	}
	return false
}

func patchPaths(patch string) ([]string, error) {
	seen := make(map[string]bool)
	paths := []string{}
	for _, line := range strings.Split(patch, "\n") {
		if !strings.HasPrefix(line, "+++ ") && !strings.HasPrefix(line, "--- ") {
			continue
		}
		value := strings.TrimSpace(line[4:])
		if value == "/dev/null" {
			continue
		}
		if strings.HasPrefix(value, "a/") || strings.HasPrefix(value, "b/") {
			value = value[2:]
		}
		value = filepath.ToSlash(filepath.Clean(value))
		if value == "." || value == ".." || filepath.IsAbs(value) || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") {
			return nil, errors.New("patch contains an unsafe path")
		}
		if !seen[value] {
			seen[value] = true
			paths = append(paths, value)
		}
	}
	if len(paths) == 0 || len(paths) > 256 {
		return nil, errors.New("patch has no bounded file set")
	}
	sort.Strings(paths)
	return paths, nil
}

func secretPath(relative string) bool {
	base := strings.ToLower(filepath.Base(relative))
	return base == ".env" || strings.HasPrefix(base, ".env.") || strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".p12") || strings.Contains(base, "credential")
}

func decode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("tool arguments do not match the strict schema")
	}
	if decoder.Decode(&struct{}{}) == nil {
		return errors.New("tool arguments contain trailing data")
	}
	return nil
}

func encode(result ToolResult) ([]byte, error) {
	if result.Paths == nil {
		result.Paths = []string{}
	}
	return json.Marshal(result)
}

func bounded(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[:maximum]
}
