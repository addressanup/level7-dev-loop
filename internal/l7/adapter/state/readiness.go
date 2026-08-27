package state

import (
	"errors"
	"fmt"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type readinessFile struct {
	Schema              int                   `json:"schema"`
	ChangeID            string                `json:"change_id"`
	Tier                domain.RiskTier       `json:"tier"`
	BaseCommit          string                `json:"base_commit"`
	CandidateCommit     string                `json:"candidate_commit"`
	CandidateTree       string                `json:"candidate_tree"`
	BriefCommit         string                `json:"brief_commit,omitempty"`
	ConfigurationDigest string                `json:"configuration_digest"`
	VerificationCommit  string                `json:"verification_commit"`
	ReviewCommit        string                `json:"review_commit"`
	Scope               []string              `json:"scope"`
	Checks              []checkFile           `json:"checks"`
	Owner               string                `json:"owner,omitempty"`
	Implementer         domain.Provider       `json:"implementer"`
	Reviewer            domain.Provider       `json:"reviewer"`
	ReviewDecision      domain.ReviewDecision `json:"review_decision"`
	BenchmarkRequired   bool                  `json:"benchmark_required"`
}

type mergeFile struct {
	Schema              int    `json:"schema"`
	ChangeID            string `json:"change_id"`
	TargetRef           string `json:"target_ref"`
	PreviousCommit      string `json:"previous_commit"`
	CandidateCommit     string `json:"candidate_commit"`
	CandidateTree       string `json:"candidate_tree"`
	ConfigurationDigest string `json:"configuration_digest"`
	VerificationCommit  string `json:"verification_commit"`
	ReviewCommit        string `json:"review_commit"`
}

func LoadReadiness(commonDirectory string) (domain.ReadinessEvidence, bool, error) {
	var file readinessFile
	found, err := loadEvidence(commonDirectory, "readiness.json", &file)
	if err != nil || !found {
		return domain.ReadinessEvidence{}, found, err
	}
	evidence := readinessDomain(file)
	if file.Schema != domain.ReadinessEvidenceSchema || !domain.ReadinessEvidenceValid(evidence) {
		return domain.ReadinessEvidence{}, false, errors.New("readiness evidence is invalid")
	}
	return evidence, true, nil
}

func SaveReadiness(commonDirectory string, evidence domain.ReadinessEvidence) error {
	if !domain.ReadinessEvidenceValid(evidence) {
		return errors.New("readiness evidence is invalid")
	}
	if _, _, err := LoadReadiness(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid readiness evidence: %w", err)
	}
	return saveEvidence(commonDirectory, "readiness.json", readinessFromDomain(evidence))
}

func LoadMerge(commonDirectory string) (domain.MergeReceipt, bool, error) {
	var file mergeFile
	found, err := loadEvidence(commonDirectory, "merge.json", &file)
	if err != nil || !found {
		return domain.MergeReceipt{}, found, err
	}
	receipt := domain.MergeReceipt{
		ChangeID: file.ChangeID, TargetRef: file.TargetRef, PreviousCommit: file.PreviousCommit,
		Candidate:           domain.CandidateIdentity{Commit: file.CandidateCommit, Tree: file.CandidateTree},
		ConfigurationDigest: file.ConfigurationDigest, VerificationCommit: file.VerificationCommit, ReviewCommit: file.ReviewCommit,
	}
	if file.Schema != domain.MergeReceiptSchema || !domain.MergeReceiptValid(receipt) {
		return domain.MergeReceipt{}, false, errors.New("merge receipt is invalid")
	}
	return receipt, true, nil
}

func SaveMerge(commonDirectory string, receipt domain.MergeReceipt) error {
	if !domain.MergeReceiptValid(receipt) {
		return errors.New("merge receipt is invalid")
	}
	if _, _, err := LoadMerge(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid merge receipt: %w", err)
	}
	file := mergeFile{
		Schema: domain.MergeReceiptSchema, ChangeID: receipt.ChangeID, TargetRef: receipt.TargetRef,
		PreviousCommit: receipt.PreviousCommit, CandidateCommit: receipt.Candidate.Commit, CandidateTree: receipt.Candidate.Tree,
		ConfigurationDigest: receipt.ConfigurationDigest, VerificationCommit: receipt.VerificationCommit, ReviewCommit: receipt.ReviewCommit,
	}
	return saveEvidence(commonDirectory, "merge.json", file)
}

func readinessFromDomain(evidence domain.ReadinessEvidence) readinessFile {
	checks := make([]checkFile, 0, len(evidence.Checks))
	for _, check := range evidence.Checks {
		checks = append(checks, checkFile{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
	}
	return readinessFile{
		Schema: domain.ReadinessEvidenceSchema, ChangeID: evidence.ChangeID, Tier: evidence.Tier,
		BaseCommit: evidence.Base, CandidateCommit: evidence.Candidate.Commit, CandidateTree: evidence.Candidate.Tree,
		BriefCommit: evidence.BriefCommit, ConfigurationDigest: evidence.ConfigurationDigest,
		VerificationCommit: evidence.VerificationCommit, ReviewCommit: evidence.ReviewCommit,
		Scope: append([]string{}, evidence.Scope...), Checks: checks, Owner: evidence.Owner,
		Implementer: evidence.Implementer, Reviewer: evidence.Reviewer, ReviewDecision: evidence.ReviewDecision,
		BenchmarkRequired: evidence.BenchmarkRequired,
	}
}

func readinessDomain(file readinessFile) domain.ReadinessEvidence {
	checks := make([]domain.CheckResult, 0, len(file.Checks))
	for _, check := range file.Checks {
		checks = append(checks, domain.CheckResult{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
	}
	return domain.ReadinessEvidence{
		ChangeID: file.ChangeID, Tier: file.Tier, Base: file.BaseCommit,
		Candidate:   domain.CandidateIdentity{Commit: file.CandidateCommit, Tree: file.CandidateTree},
		BriefCommit: file.BriefCommit, ConfigurationDigest: file.ConfigurationDigest,
		VerificationCommit: file.VerificationCommit, ReviewCommit: file.ReviewCommit,
		Scope: append([]string{}, file.Scope...), Checks: checks, Owner: file.Owner,
		Implementer: file.Implementer, Reviewer: file.Reviewer, ReviewDecision: file.ReviewDecision,
		BenchmarkRequired: file.BenchmarkRequired,
	}
}
