package state

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const MaxEvidenceFile = 256 << 10

type providerFile struct {
	Name       domain.Provider        `json:"name"`
	Executable string                 `json:"executable"`
	Version    string                 `json:"version"`
	Digest     string                 `json:"digest"`
	Capability domain.CapabilityState `json:"capability"`
}

type checkFile struct {
	Name      string `json:"name"`
	Benchmark bool   `json:"benchmark"`
	Passed    bool   `json:"passed"`
	ExitCode  int    `json:"exit_code"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type runFile struct {
	Schema        int          `json:"schema"`
	ChangeID      string       `json:"change_id"`
	Provider      providerFile `json:"provider"`
	ParentCommit  string       `json:"parent_commit"`
	ParentTree    string       `json:"parent_tree"`
	Commit        string       `json:"candidate_commit"`
	Tree          string       `json:"candidate_tree"`
	PathDigest    string       `json:"path_digest"`
	PathCount     int          `json:"path_count"`
	CommitMessage string       `json:"commit_message"`
}

type verificationFile struct {
	Schema              int                   `json:"schema"`
	ChangeID            string                `json:"change_id"`
	CandidateCommit     string                `json:"candidate_commit"`
	CandidateTree       string                `json:"candidate_tree"`
	Result              domain.ReviewDecision `json:"result"`
	Checks              []checkFile           `json:"checks"`
	ConfigurationDigest string                `json:"configuration_digest,omitempty"`
	VerificationCommit  string                `json:"verification_commit"`
	VerificationTree    string                `json:"verification_tree"`
}

type reviewFile struct {
	Schema          int                   `json:"schema"`
	ChangeID        string                `json:"change_id"`
	Provider        providerFile          `json:"provider"`
	CandidateCommit string                `json:"candidate_commit"`
	CandidateTree   string                `json:"candidate_tree"`
	Decision        domain.ReviewDecision `json:"decision"`
	Findings        []string              `json:"findings"`
	ReviewCommit    string                `json:"review_commit"`
	ReviewTree      string                `json:"review_tree"`
}

func LoadRun(commonDirectory string) (domain.RunEvidence, bool, error) {
	var file runFile
	found, err := loadEvidence(commonDirectory, "run.json", &file)
	if err != nil || !found {
		return domain.RunEvidence{}, found, err
	}
	if err := validateRun(file); err != nil {
		return domain.RunEvidence{}, false, err
	}
	return domain.RunEvidence{
		ChangeID: file.ChangeID, Provider: providerDomain(file.Provider),
		Parent:     domain.CandidateIdentity{Commit: file.ParentCommit, Tree: file.ParentTree},
		Candidate:  domain.CandidateIdentity{Commit: file.Commit, Tree: file.Tree},
		PathDigest: file.PathDigest, PathCount: file.PathCount, CommitMessage: file.CommitMessage,
	}, true, nil
}

func SaveRun(commonDirectory string, evidence domain.RunEvidence) error {
	file := runFile{
		Schema: domain.EvidenceSchema, ChangeID: evidence.ChangeID, Provider: providerFromDomain(evidence.Provider),
		ParentCommit: evidence.Parent.Commit, ParentTree: evidence.Parent.Tree,
		Commit: evidence.Candidate.Commit, Tree: evidence.Candidate.Tree,
		PathDigest: evidence.PathDigest, PathCount: evidence.PathCount, CommitMessage: evidence.CommitMessage,
	}
	if err := validateRun(file); err != nil {
		return err
	}
	if _, _, err := LoadRun(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid run evidence: %w", err)
	}
	return saveEvidence(commonDirectory, "run.json", file)
}

func LoadVerification(commonDirectory string) (domain.VerificationEvidence, bool, error) {
	var file verificationFile
	found, err := loadEvidence(commonDirectory, "verification.json", &file)
	if err != nil || !found {
		return domain.VerificationEvidence{}, found, err
	}
	if err := validateVerification(file); err != nil {
		return domain.VerificationEvidence{}, false, err
	}
	return verificationDomain(file), true, nil
}

func SaveVerification(commonDirectory string, evidence domain.VerificationEvidence) error {
	file := verificationFromDomain(evidence)
	if err := validateVerification(file); err != nil {
		return err
	}
	if _, _, err := LoadVerification(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid verification evidence: %w", err)
	}
	return saveEvidence(commonDirectory, "verification.json", file)
}

func LoadReview(commonDirectory string) (domain.ReviewEvidence, bool, error) {
	var file reviewFile
	found, err := loadEvidence(commonDirectory, "review.json", &file)
	if err != nil || !found {
		return domain.ReviewEvidence{}, found, err
	}
	if err := validateReview(file); err != nil {
		return domain.ReviewEvidence{}, false, err
	}
	return reviewDomain(file), true, nil
}

func SaveReview(commonDirectory string, evidence domain.ReviewEvidence) error {
	file := reviewFromDomain(evidence)
	if err := validateReview(file); err != nil {
		return err
	}
	if _, _, err := LoadReview(commonDirectory); err != nil {
		return fmt.Errorf("refuse to replace invalid review evidence: %w", err)
	}
	return saveEvidence(commonDirectory, "review.json", file)
}

func WriteVerificationArtifact(root string, evidence domain.VerificationEvidence, reviewer string) (string, error) {
	if !filepath.IsAbs(root) || !safeChangeID(evidence.ChangeID) || !safeLine(reviewer) || errCandidate(evidence.Candidate) != nil || evidence.Result != domain.DecisionGO {
		return "", errors.New("verification artifact input is invalid")
	}
	relative := "docs/artifacts/changes/" + evidence.ChangeID + "-verification.md"
	var document bytes.Buffer
	fmt.Fprintf(&document, "# %s — Verification\n\n", evidence.ChangeID)
	fmt.Fprintf(&document, "| Field | Value |\n|---|---|\n")
	fmt.Fprintf(&document, "| Change ID | `%s` |\n", evidence.ChangeID)
	fmt.Fprintf(&document, "| Candidate commit | `%s` |\n", evidence.Candidate.Commit)
	fmt.Fprintf(&document, "| Candidate tree | `%s` |\n", evidence.Candidate.Tree)
	fmt.Fprintf(&document, "| Configuration digest | `%s` |\n", evidence.ConfigurationDigest)
	fmt.Fprintf(&document, "| Result | `PASS` |\n")
	fmt.Fprintf(&document, "| Reviewer | `%s` |\n\n", reviewer)
	fmt.Fprintf(&document, "## Checks\n\n| Check | Result |\n|---|---|\n")
	for _, check := range evidence.Checks {
		result := "FAIL"
		if check.Passed {
			result = "PASS"
		}
		fmt.Fprintf(&document, "| `%s` | %s |\n", check.Name, result)
	}
	return relative, writeArtifact(root, relative, document.Bytes())
}

func WriteAuditArtifact(root string, evidence domain.ReviewEvidence) (string, error) {
	if !filepath.IsAbs(root) || !safeChangeID(evidence.ChangeID) || errCandidate(evidence.Candidate) != nil || !evidence.Decision.Valid() || !safeProvider(providerFromDomain(evidence.Provider)) {
		return "", errors.New("audit artifact input is invalid")
	}
	relative := "docs/artifacts/changes/" + evidence.ChangeID + "-audit.md"
	var document bytes.Buffer
	fmt.Fprintf(&document, "# %s — Independent Audit\n\n", evidence.ChangeID)
	fmt.Fprintf(&document, "| Field | Value |\n|---|---|\n")
	fmt.Fprintf(&document, "| Change ID | `%s` |\n", evidence.ChangeID)
	fmt.Fprintf(&document, "| Candidate commit | `%s` |\n", evidence.Candidate.Commit)
	fmt.Fprintf(&document, "| Candidate tree | `%s` |\n", evidence.Candidate.Tree)
	fmt.Fprintf(&document, "| Result | `%s` |\n", evidence.Decision)
	fmt.Fprintf(&document, "| Reviewer | `%s` |\n\n", evidence.Provider.Provider)
	fmt.Fprintf(&document, "## Findings\n\n")
	if len(evidence.Findings) == 0 {
		fmt.Fprintln(&document, "No findings recorded.")
	} else {
		for _, finding := range evidence.Findings {
			fmt.Fprintf(&document, "- %s\n", finding)
		}
	}
	return relative, writeArtifact(root, relative, document.Bytes())
}

func loadEvidence(commonDirectory, name string, destination any) (bool, error) {
	directory, err := productDirectory(commonDirectory)
	if err != nil {
		return false, err
	}
	data, err := localfile.Read(filepath.Join(directory, name), MaxEvidenceFile)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := localfile.DecodeJSON(data, destination); err != nil {
		return false, err
	}
	return true, nil
}

func saveEvidence(commonDirectory, name string, value any) error {
	directory, err := productDirectory(commonDirectory)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(directory, 0o700); err != nil {
		return err
	}
	path := filepath.Join(directory, name)
	data, err := localfile.EncodeJSON(value)
	if err != nil {
		return err
	}
	_, readErr := localfile.Read(path, MaxEvidenceFile)
	if errors.Is(readErr, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	}
	if readErr != nil {
		return fmt.Errorf("refuse to replace invalid evidence: %w", readErr)
	}
	return localfile.AtomicReplace(path, data, 0o600)
}

func writeArtifact(root, relative string, data []byte) error {
	name := filepath.Join(root, filepath.FromSlash(relative))
	if err := localfile.EnsureDirectory(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	_, err := localfile.Read(name, MaxEvidenceFile)
	if errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(name, data, 0o644)
	}
	if err != nil {
		return fmt.Errorf("refuse to replace unsafe artifact: %w", err)
	}
	return localfile.AtomicReplace(name, data, 0o644)
}

func validateRun(file runFile) error {
	if file.Schema != domain.EvidenceSchema || !safeChangeID(file.ChangeID) || !safeProvider(file.Provider) || !fullGitID(file.ParentCommit) || !fullGitID(file.ParentTree) || !hexDigest(file.PathDigest) || file.PathCount < 1 || file.PathCount > 1_000_000 || !domain.ConventionalSubject(file.CommitMessage) {
		return errors.New("run evidence is invalid")
	}
	if (file.Commit == "") != (file.Tree == "") || (file.Commit != "" && (!fullGitID(file.Commit) || !fullGitID(file.Tree))) {
		return errors.New("run evidence is invalid")
	}
	return nil
}

func validateVerification(file verificationFile) error {
	legacy := file.Schema == domain.EvidenceSchema && file.ConfigurationDigest == ""
	current := file.Schema == domain.VerificationEvidenceSchema && hexDigest(file.ConfigurationDigest)
	if (!legacy && !current) || !safeChangeID(file.ChangeID) || !fullGitID(file.CandidateCommit) || !fullGitID(file.CandidateTree) || file.Result != domain.DecisionGO || len(file.Checks) < 1 || len(file.Checks) > 32 {
		return errors.New("verification evidence is invalid")
	}
	if (file.VerificationCommit == "") != (file.VerificationTree == "") || (file.VerificationCommit != "" && (!fullGitID(file.VerificationCommit) || !fullGitID(file.VerificationTree))) {
		return errors.New("verification successor identity is invalid")
	}
	for _, check := range file.Checks {
		if !safeToken(check.Name, 64) || !safeToken(check.Code, 64) || !safeLine(check.Message) || !check.Passed || check.ExitCode != 0 {
			return errors.New("verification check is invalid")
		}
	}
	return nil
}

func validateReview(file reviewFile) error {
	if file.Schema != domain.EvidenceSchema || !safeChangeID(file.ChangeID) || !safeProvider(file.Provider) || !fullGitID(file.CandidateCommit) || !fullGitID(file.CandidateTree) || !file.Decision.Valid() || len(file.Findings) > 64 {
		return errors.New("review evidence is invalid")
	}
	if (file.ReviewCommit == "") != (file.ReviewTree == "") || (file.ReviewCommit != "" && (!fullGitID(file.ReviewCommit) || !fullGitID(file.ReviewTree))) {
		return errors.New("review successor identity is invalid")
	}
	for _, finding := range file.Findings {
		if !safeLine(finding) {
			return errors.New("review finding is invalid")
		}
	}
	return nil
}

func safeProvider(provider providerFile) bool {
	return provider.Name.Valid() && filepath.IsAbs(provider.Executable) && safeLine(provider.Executable) && safeLine(provider.Version) && hexDigest(provider.Digest) && provider.Capability == domain.CapabilityAvailable
}

func safeToken(value string, maximum int) bool {
	return len(value) > 0 && len(value) <= maximum && !strings.ContainsAny(value, "\r\n\x00")
}

func hexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func errCandidate(candidate domain.CandidateIdentity) error {
	if !fullGitID(candidate.Commit) || !fullGitID(candidate.Tree) {
		return errors.New("candidate identity is invalid")
	}
	return nil
}

func providerFromDomain(provider domain.ProviderIdentity) providerFile {
	return providerFile{Name: provider.Provider, Executable: provider.Executable, Version: provider.Version, Digest: provider.Digest, Capability: provider.Capability}
}

func providerDomain(provider providerFile) domain.ProviderIdentity {
	return domain.ProviderIdentity{Provider: provider.Name, Executable: provider.Executable, Version: provider.Version, Digest: provider.Digest, Capability: provider.Capability}
}

func verificationFromDomain(evidence domain.VerificationEvidence) verificationFile {
	checks := make([]checkFile, 0, len(evidence.Checks))
	for _, check := range evidence.Checks {
		checks = append(checks, checkFile{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
	}
	return verificationFile{Schema: domain.VerificationEvidenceSchema, ChangeID: evidence.ChangeID, CandidateCommit: evidence.Candidate.Commit, CandidateTree: evidence.Candidate.Tree, Result: evidence.Result, Checks: checks, ConfigurationDigest: evidence.ConfigurationDigest, VerificationCommit: evidence.VerificationCommit, VerificationTree: evidence.VerificationTree}
}

func verificationDomain(file verificationFile) domain.VerificationEvidence {
	checks := make([]domain.CheckResult, 0, len(file.Checks))
	for _, check := range file.Checks {
		checks = append(checks, domain.CheckResult{Name: check.Name, Benchmark: check.Benchmark, Passed: check.Passed, ExitCode: check.ExitCode, Code: check.Code, Message: check.Message})
	}
	return domain.VerificationEvidence{ChangeID: file.ChangeID, Candidate: domain.CandidateIdentity{Commit: file.CandidateCommit, Tree: file.CandidateTree}, Result: file.Result, Checks: checks, ConfigurationDigest: file.ConfigurationDigest, VerificationCommit: file.VerificationCommit, VerificationTree: file.VerificationTree}
}

func reviewFromDomain(evidence domain.ReviewEvidence) reviewFile {
	return reviewFile{Schema: domain.EvidenceSchema, ChangeID: evidence.ChangeID, Provider: providerFromDomain(evidence.Provider), CandidateCommit: evidence.Candidate.Commit, CandidateTree: evidence.Candidate.Tree, Decision: evidence.Decision, Findings: append([]string{}, evidence.Findings...), ReviewCommit: evidence.ReviewCommit, ReviewTree: evidence.ReviewTree}
}

func reviewDomain(file reviewFile) domain.ReviewEvidence {
	return domain.ReviewEvidence{ChangeID: file.ChangeID, Provider: providerDomain(file.Provider), Candidate: domain.CandidateIdentity{Commit: file.CandidateCommit, Tree: file.CandidateTree}, Decision: file.Decision, Findings: append([]string{}, file.Findings...), ReviewCommit: file.ReviewCommit, ReviewTree: file.ReviewTree}
}
