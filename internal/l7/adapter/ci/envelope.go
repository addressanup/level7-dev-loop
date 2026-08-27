// Package ci decodes the bounded, side-effect-free trusted-CI readiness contract.
package ci

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	EnvelopeSchema   = 1
	MaxEnvelopeBytes = 256 << 10
	MaxEnvelopeScope = 4096
)

type envelopeWire struct {
	Schema              *int                   `json:"schema"`
	ChangeID            *string                `json:"change_id"`
	Tier                *domain.RiskTier       `json:"tier"`
	BaseCommit          *string                `json:"base_commit"`
	CandidateCommit     *string                `json:"candidate_commit"`
	CandidateTree       *string                `json:"candidate_tree"`
	BriefCommit         *string                `json:"brief_commit"`
	ConfigurationDigest *string                `json:"configuration_digest"`
	VerificationCommit  *string                `json:"verification_commit"`
	ReviewCommit        *string                `json:"review_commit"`
	Scope               *[]string              `json:"scope"`
	Checks              *[]checkWire           `json:"checks"`
	Owner               *string                `json:"owner"`
	Implementer         *domain.Provider       `json:"implementer"`
	Reviewer            *domain.Provider       `json:"reviewer"`
	ReviewDecision      *domain.ReviewDecision `json:"review_decision"`
	BenchmarkRequired   *bool                  `json:"benchmark_required"`
	PlanCurrent         *bool                  `json:"plan_current"`
	RepositoryClean     *bool                  `json:"repository_clean"`
	ApprovalCurrent     *bool                  `json:"approval_current"`
	VerificationCurrent *bool                  `json:"verification_current"`
	ReviewCurrent       *bool                  `json:"review_current"`
	AuditCurrent        *bool                  `json:"audit_current"`
}

type checkWire struct {
	Name      *string `json:"name"`
	Benchmark *bool   `json:"benchmark"`
	Passed    *bool   `json:"passed"`
	ExitCode  *int    `json:"exit_code"`
	Code      *string `json:"code"`
	Message   *string `json:"message"`
}

func Decode(data []byte) (domain.ReadinessFacts, error) {
	if len(data) < 2 || len(data) > MaxEnvelopeBytes {
		return domain.ReadinessFacts{}, errors.New("trusted-CI envelope is empty or exceeds the size limit")
	}
	if err := rejectDuplicateFields(data); err != nil {
		return domain.ReadinessFacts{}, err
	}
	var wire envelopeWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return domain.ReadinessFacts{}, fmt.Errorf("decode trusted-CI envelope: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.ReadinessFacts{}, errors.New("trusted-CI envelope contains trailing data")
	}
	if !completeEnvelope(wire) || *wire.Schema != EnvelopeSchema || len(*wire.Scope) < 1 || len(*wire.Scope) > MaxEnvelopeScope || len(*wire.Checks) < 1 || len(*wire.Checks) > 32 {
		return domain.ReadinessFacts{}, errors.New("trusted-CI envelope is missing a required bounded field")
	}
	checks := make([]domain.CheckResult, 0, len(*wire.Checks))
	for _, check := range *wire.Checks {
		if check.Name == nil || check.Benchmark == nil || check.Passed == nil || check.ExitCode == nil || check.Code == nil || check.Message == nil {
			return domain.ReadinessFacts{}, errors.New("trusted-CI check is missing a required field")
		}
		checks = append(checks, domain.CheckResult{Name: *check.Name, Benchmark: *check.Benchmark, Passed: *check.Passed, ExitCode: *check.ExitCode, Code: *check.Code, Message: *check.Message})
	}
	evidence := domain.ReadinessEvidence{
		ChangeID: *wire.ChangeID, Tier: *wire.Tier, Base: *wire.BaseCommit,
		Candidate:   domain.CandidateIdentity{Commit: *wire.CandidateCommit, Tree: *wire.CandidateTree},
		BriefCommit: *wire.BriefCommit, ConfigurationDigest: *wire.ConfigurationDigest,
		VerificationCommit: *wire.VerificationCommit, ReviewCommit: *wire.ReviewCommit,
		Scope: append([]string{}, (*wire.Scope)...), Checks: checks, Owner: *wire.Owner,
		Implementer: *wire.Implementer, Reviewer: *wire.Reviewer, ReviewDecision: *wire.ReviewDecision,
		BenchmarkRequired: *wire.BenchmarkRequired,
	}
	return domain.ReadinessFacts{
		Evidence: evidence, PlanCurrent: *wire.PlanCurrent, RepositoryClean: *wire.RepositoryClean,
		ApprovalCurrent: *wire.ApprovalCurrent, VerificationCurrent: *wire.VerificationCurrent,
		ReviewCurrent: *wire.ReviewCurrent, AuditCurrent: *wire.AuditCurrent,
	}, nil
}

func completeEnvelope(wire envelopeWire) bool {
	return wire.Schema != nil && wire.ChangeID != nil && wire.Tier != nil && wire.BaseCommit != nil && wire.CandidateCommit != nil && wire.CandidateTree != nil && wire.BriefCommit != nil && wire.ConfigurationDigest != nil && wire.VerificationCommit != nil && wire.ReviewCommit != nil && wire.Scope != nil && wire.Checks != nil && wire.Owner != nil && wire.Implementer != nil && wire.Reviewer != nil && wire.ReviewDecision != nil && wire.BenchmarkRequired != nil && wire.PlanCurrent != nil && wire.RepositoryClean != nil && wire.ApprovalCurrent != nil && wire.VerificationCurrent != nil && wire.ReviewCurrent != nil && wire.AuditCurrent != nil
}

func rejectDuplicateFields(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := walkJSONValue(decoder); err != nil {
		return fmt.Errorf("validate trusted-CI JSON shape: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("trusted-CI envelope contains trailing data")
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] {
				return errors.New("trusted-CI envelope contains a duplicate or invalid object field")
			}
			seen[key] = true
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("trusted-CI object is not closed")
		}
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("trusted-CI array is not closed")
		}
	default:
		return errors.New("trusted-CI JSON has an invalid delimiter")
	}
	return nil
}
