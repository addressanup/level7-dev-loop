package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strings"
)

type changedPath struct {
	Status string
	Path   string
}

type gitRepository struct{ root string }

func (repository gitRepository) run(arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", repository.root}, arguments...)...)
	output, err := command.Output()
	if err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s: %s", strings.Join(arguments, " "), strings.TrimSpace(string(failure.Stderr)))
		}
		return nil, err
	}
	return output, nil
}

func (repository gitRepository) resolve(ref string) (string, error) {
	output, err := repository.run("rev-parse", "--verify", ref+"^{commit}")
	return strings.TrimSpace(string(output)), err
}

func (repository gitRepository) tree(ref string) (string, error) {
	output, err := repository.run("rev-parse", "--verify", ref+"^{tree}")
	return strings.TrimSpace(string(output)), err
}

func (repository gitRepository) show(ref, relative string) ([]byte, error) {
	if !safeGitPath(relative) {
		return nil, fmt.Errorf("unsafe Git path %q", relative)
	}
	output, err := repository.run("show", ref+":"+relative)
	if len(output) > maxInputBytes {
		return nil, fmt.Errorf("Git object %s exceeds the input limit", relative)
	}
	return output, err
}

func (repository gitRepository) changed(base, head string) ([]changedPath, error) {
	output, err := repository.run("diff", "--name-status", "--no-renames", "-z", base+".."+head)
	if err != nil {
		return nil, err
	}
	fields := bytes.Split(output, []byte{0})
	var result []changedPath
	for index := 0; index+1 < len(fields); index += 2 {
		if len(fields[index]) == 0 {
			continue
		}
		status, relative := string(fields[index]), string(fields[index+1])
		if !safeGitPath(relative) {
			return nil, fmt.Errorf("unsafe changed path %q", relative)
		}
		result = append(result, changedPath{Status: status, Path: relative})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func (repository gitRepository) list(ref, prefix string) ([]string, error) {
	output, err := repository.run("ls-tree", "-r", "-z", "--name-only", ref, "--", prefix)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, item := range bytes.Split(output, []byte{0}) {
		if len(item) == 0 {
			continue
		}
		relative := string(item)
		if !safeGitPath(relative) {
			return nil, fmt.Errorf("unsafe tree path %q", relative)
		}
		result = append(result, relative)
	}
	sort.Strings(result)
	return result, nil
}

func (repository gitRepository) additionCommit(head, relative string) (string, error) {
	output, err := repository.run("log", "--diff-filter=A", "-1", "--format=%H", head, "--", relative)
	if err != nil {
		return "", err
	}
	commit := strings.TrimSpace(string(output))
	if commit == "" {
		return "", fmt.Errorf("no addition commit for %s", relative)
	}
	return commit, nil
}

func (repository gitRepository) commonDir() (string, error) {
	output, err := repository.run("rev-parse", "--path-format=absolute", "--git-common-dir")
	return strings.TrimSpace(string(output)), err
}

func (repository gitRepository) isAncestor(base, head string) bool {
	command := exec.Command("git", "-C", repository.root, "merge-base", "--is-ancestor", base, head)
	return command.Run() == nil
}

func safeGitPath(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || path.Clean(value) != value {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if component == "." || component == ".." || component == "" {
			return false
		}
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}
