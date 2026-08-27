// Package presentation renders deterministic Level 7 CLI results.
package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type wireResult struct {
	Schema     int             `json:"schema"`
	Outcome    domain.Outcome  `json:"outcome"`
	Code       string          `json:"code"`
	Command    string          `json:"command"`
	State      string          `json:"state"`
	Version    string          `json:"version"`
	Message    string          `json:"message"`
	Next       string          `json:"next"`
	Details    []string        `json:"details"`
	Repository *wireRepository `json:"repository,omitempty"`
	Execution  *wireExecution  `json:"execution,omitempty"`
	Readiness  *wireReadiness  `json:"readiness,omitempty"`
}

type wireExecution struct {
	Role       domain.ProviderRole   `json:"role,omitempty"`
	Provider   domain.Provider       `json:"provider,omitempty"`
	Executable string                `json:"executable,omitempty"`
	Version    string                `json:"provider_version,omitempty"`
	Digest     string                `json:"executable_digest,omitempty"`
	Commit     string                `json:"candidate_commit,omitempty"`
	Tree       string                `json:"candidate_tree,omitempty"`
	Decision   domain.ReviewDecision `json:"decision,omitempty"`
	Checks     []wireCheck           `json:"checks"`
}

type wireCheck struct {
	Name      string `json:"name"`
	Benchmark bool   `json:"benchmark"`
	Passed    bool   `json:"passed"`
	ExitCode  int    `json:"exit_code"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type wireReadiness struct {
	Headless            bool            `json:"headless"`
	Ready               bool            `json:"ready"`
	Base                string          `json:"base"`
	Candidate           string          `json:"candidate_commit"`
	Tree                string          `json:"candidate_tree"`
	BriefCommit         string          `json:"brief_commit,omitempty"`
	ConfigurationDigest string          `json:"configuration_digest"`
	VerificationCommit  string          `json:"verification_commit"`
	ReviewCommit        string          `json:"review_commit"`
	Owner               string          `json:"owner,omitempty"`
	Implementer         domain.Provider `json:"implementer"`
	Reviewer            domain.Provider `json:"reviewer"`
	TargetRef           string          `json:"target_ref,omitempty"`
	PreviousCommit      string          `json:"previous_commit,omitempty"`
	Checks              []wireCheck     `json:"checks"`
}

type wireRepository struct {
	Root          string          `json:"root"`
	CommonDir     string          `json:"common_dir"`
	ChangeID      string          `json:"change_id,omitempty"`
	Tier          domain.RiskTier `json:"tier,omitempty"`
	Base          string          `json:"base,omitempty"`
	Head          string          `json:"head"`
	Tree          string          `json:"tree"`
	DeclaredScope []string        `json:"declared_scope"`
	ChangedPaths  []string        `json:"changed_paths"`
	ExpandedPaths []string        `json:"expanded_paths"`
}

func Text(result domain.Result) []byte {
	var output bytes.Buffer
	fmt.Fprintf(&output, "%s code=%s command=%s state=%s version=%s message=%s next=%s\n",
		result.Outcome,
		strconv.Quote(result.Code),
		strconv.Quote(result.Command),
		strconv.Quote(result.State),
		strconv.Quote(result.Version),
		strconv.Quote(result.Message),
		strconv.Quote(result.Next),
	)
	for _, detail := range result.Details {
		fmt.Fprintf(&output, "detail=%s\n", strconv.Quote(detail))
	}
	if repository := result.Repository; repository != nil {
		fmt.Fprintf(&output, "repository_root=%s\n", strconv.Quote(repository.Root))
		fmt.Fprintf(&output, "git_common_dir=%s\n", strconv.Quote(repository.CommonDir))
		if repository.ChangeID != "" {
			fmt.Fprintf(&output, "change_id=%s\n", strconv.Quote(repository.ChangeID))
		}
		if repository.Tier.Valid() {
			fmt.Fprintf(&output, "risk_tier=%d\n", repository.Tier)
		}
		if repository.Base != "" {
			fmt.Fprintf(&output, "base=%s\n", strconv.Quote(repository.Base))
		}
		fmt.Fprintf(&output, "head=%s\n", strconv.Quote(repository.Head))
		fmt.Fprintf(&output, "tree=%s\n", strconv.Quote(repository.Tree))
		writeValues(&output, "declared_scope", repository.DeclaredScope)
		writeValues(&output, "changed_path", repository.ChangedPaths)
		writeValues(&output, "expanded_path", repository.ExpandedPaths)
	}
	if execution := result.Execution; execution != nil {
		if execution.Role.Valid() {
			fmt.Fprintf(&output, "execution_role=%s\n", strconv.Quote(string(execution.Role)))
		}
		if execution.Provider.Valid() {
			fmt.Fprintf(&output, "provider=%s\n", strconv.Quote(string(execution.Provider)))
		}
		for _, value := range []struct {
			label string
			value string
		}{
			{label: "provider_executable", value: execution.Executable},
			{label: "provider_version", value: execution.Version},
			{label: "provider_digest", value: execution.Digest},
			{label: "candidate_commit", value: execution.Commit},
			{label: "candidate_tree", value: execution.Tree},
			{label: "review_decision", value: string(execution.Decision)},
		} {
			if value.value != "" {
				fmt.Fprintf(&output, "%s=%s\n", value.label, strconv.Quote(value.value))
			}
		}
		for _, check := range execution.Checks {
			fmt.Fprintf(&output, "check_name=%s check_passed=%t check_exit=%d check_code=%s check_message=%s\n", strconv.Quote(check.Name), check.Passed, check.ExitCode, strconv.Quote(check.Code), strconv.Quote(check.Message))
		}
	}
	if readiness := result.Readiness; readiness != nil {
		fmt.Fprintf(&output, "readiness_headless=%t\n", readiness.Headless)
		fmt.Fprintf(&output, "readiness_ready=%t\n", readiness.Ready)
		for _, value := range []struct {
			label string
			value string
		}{
			{label: "readiness_base", value: readiness.Base},
			{label: "readiness_candidate", value: readiness.Candidate},
			{label: "readiness_tree", value: readiness.Tree},
			{label: "readiness_brief_commit", value: readiness.BriefCommit},
			{label: "readiness_configuration_digest", value: readiness.ConfigurationDigest},
			{label: "readiness_verification_commit", value: readiness.VerificationCommit},
			{label: "readiness_review_commit", value: readiness.ReviewCommit},
			{label: "readiness_owner", value: readiness.Owner},
			{label: "readiness_implementer", value: string(readiness.Implementer)},
			{label: "readiness_reviewer", value: string(readiness.Reviewer)},
			{label: "merge_target_ref", value: readiness.TargetRef},
			{label: "merge_previous_commit", value: readiness.PreviousCommit},
		} {
			if value.value != "" {
				fmt.Fprintf(&output, "%s=%s\n", value.label, strconv.Quote(value.value))
			}
		}
		for _, check := range readiness.Checks {
			fmt.Fprintf(&output, "readiness_check=%s benchmark=%t passed=%t\n", strconv.Quote(check.Name), check.Benchmark, check.Passed)
		}
	}
	return output.Bytes()
}

func JSON(result domain.Result) ([]byte, error) {
	details := append([]string{}, result.Details...)
	wire := wireResult{
		Schema:  result.Schema,
		Outcome: result.Outcome,
		Code:    result.Code,
		Command: result.Command,
		State:   result.State,
		Version: result.Version,
		Message: result.Message,
		Next:    result.Next,
		Details: details,
	}
	if repository := result.Repository; repository != nil {
		wire.Repository = &wireRepository{
			Root:          repository.Root,
			CommonDir:     repository.CommonDir,
			ChangeID:      repository.ChangeID,
			Tier:          repository.Tier,
			Base:          repository.Base,
			Head:          repository.Head,
			Tree:          repository.Tree,
			DeclaredScope: append([]string{}, repository.DeclaredScope...),
			ChangedPaths:  append([]string{}, repository.ChangedPaths...),
			ExpandedPaths: append([]string{}, repository.ExpandedPaths...),
		}
	}
	if execution := result.Execution; execution != nil {
		checks := make([]wireCheck, 0, len(execution.Checks))
		for _, check := range execution.Checks {
			checks = append(checks, wireCheck{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
		}
		wire.Execution = &wireExecution{
			Role: execution.Role, Provider: execution.Provider, Executable: execution.Executable, Version: execution.Version,
			Digest: execution.Digest, Commit: execution.Commit, Tree: execution.Tree, Decision: execution.Decision, Checks: checks,
		}
	}
	if readiness := result.Readiness; readiness != nil {
		checks := make([]wireCheck, 0, len(readiness.Checks))
		for _, check := range readiness.Checks {
			checks = append(checks, wireCheck{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
		}
		wire.Readiness = &wireReadiness{
			Headless: readiness.Headless, Ready: readiness.Ready, Base: readiness.Base,
			Candidate: readiness.Candidate, Tree: readiness.Tree, BriefCommit: readiness.BriefCommit,
			ConfigurationDigest: readiness.ConfigurationDigest, VerificationCommit: readiness.VerificationCommit,
			ReviewCommit: readiness.ReviewCommit, Owner: readiness.Owner, Implementer: readiness.Implementer,
			Reviewer: readiness.Reviewer, TargetRef: readiness.TargetRef, PreviousCommit: readiness.PreviousCommit,
			Checks: checks,
		}
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func writeValues(output *bytes.Buffer, label string, values []string) {
	for _, value := range values {
		fmt.Fprintf(output, "%s=%s\n", label, strconv.Quote(value))
	}
}
