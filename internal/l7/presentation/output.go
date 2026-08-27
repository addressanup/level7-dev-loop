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
