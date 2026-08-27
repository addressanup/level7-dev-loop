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
	Schema  int            `json:"schema"`
	Outcome domain.Outcome `json:"outcome"`
	Code    string         `json:"code"`
	Command string         `json:"command"`
	State   string         `json:"state"`
	Version string         `json:"version"`
	Message string         `json:"message"`
	Next    string         `json:"next"`
	Details []string       `json:"details"`
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
	return output.Bytes()
}

func JSON(result domain.Result) ([]byte, error) {
	details := append([]string{}, result.Details...)
	data, err := json.Marshal(wireResult{
		Schema:  result.Schema,
		Outcome: result.Outcome,
		Code:    result.Code,
		Command: result.Command,
		State:   result.State,
		Version: result.Version,
		Message: result.Message,
		Next:    result.Next,
		Details: details,
	})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
