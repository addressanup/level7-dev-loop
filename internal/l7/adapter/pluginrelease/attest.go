package pluginrelease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

const (
	// MinimumGHVersion is the oldest GitHub CLI accepted as a verifier: it
	// carries --source-digest/--source-ref and the fix for the attestation
	// exit-status defect (GHSA-fgw4-v983-mgp8).
	MinimumGHVersion = "2.67.0"
	ghOutputBytes    = 8 << 20
	ghTimeout        = 3 * time.Minute
	oidcIssuer       = "https://token.actions.githubusercontent.com"
)

type ResolveFunc func(string) (processadapter.Executable, error)
type RunFunc func(context.Context, processadapter.Request) (processadapter.Result, error)

// Verifier runs `gh attestation verify --bundle` against locally fetched
// bundles. It needs no GitHub login and is never given a token.
type Verifier struct {
	resolve ResolveFunc
	run     RunFunc
}

func NewVerifier(resolve ResolveFunc, run RunFunc) Verifier {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	return Verifier{resolve: resolve, run: run}
}

// Expectation is the exact identity an attestation must carry.
type Expectation struct {
	Workflow      Workflow
	SourceCommit  string
	SubjectName   string
	SubjectSHA256 string
}

// Evidence summarizes one successful verification for state and status.
type Evidence struct {
	Verifier       string `json:"verifier"`
	SignerIdentity string `json:"signer_identity"`
	SourceCommit   string `json:"source_commit"`
	RunInvocation  string `json:"run_invocation"`
	Timestamp      string `json:"timestamp"`
	SubjectSHA256  string `json:"subject_sha256"`
}

// Prerequisite resolves gh and checks its version. The error text is the
// exact remediation a user needs.
func (verifier Verifier) Prerequisite(ctx context.Context) (processadapter.Executable, string, error) {
	executable, err := verifier.resolve("gh")
	if err != nil {
		return processadapter.Executable{}, "", errors.New("the GitHub CLI (gh) is required to verify release attestations; install gh " + MinimumGHVersion + " or newer (https://cli.github.com) and ensure it is on PATH; no GitHub login is needed")
	}
	result, err := verifier.run(ctx, processadapter.Request{
		Executable: executable.Path, Arguments: []string{"--version"}, Directory: "/",
		Environment: ghEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second,
	})
	if err != nil || result.ExitCode != 0 {
		return processadapter.Executable{}, "", errors.New("gh --version failed; repair the GitHub CLI installation before installing")
	}
	version, err := parseGHVersion(string(result.Stdout))
	if err != nil {
		return processadapter.Executable{}, "", err
	}
	minimum, _ := ParseVersion(MinimumGHVersion)
	parsed, err := ParseVersion(version)
	if err != nil || parsed.Compare(minimum) < 0 {
		return processadapter.Executable{}, "", fmt.Errorf("gh %s is too old; install gh %s or newer to verify attestations", version, MinimumGHVersion)
	}
	return executable, version, nil
}

func parseGHVersion(output string) (string, error) {
	line, _, _ := strings.Cut(strings.TrimSpace(output), "\n")
	fields := strings.Fields(line)
	if len(fields) < 3 || fields[0] != "gh" || fields[1] != "version" {
		return "", errors.New("gh --version output is unrecognized")
	}
	version := strings.TrimPrefix(fields[2], "v")
	if _, err := ParseVersion(version); err != nil {
		return "", errors.New("gh --version reports an unrecognized version")
	}
	return version, nil
}

// Verify binds artifact bytes to the trusted workflow, repository, ref, and
// exact source commit. Both gh's enforcement and an independent check of the
// returned certificate and statement must agree.
func (verifier Verifier) Verify(ctx context.Context, gh processadapter.Executable, artifactPath, bundlePath string, expect Expectation) (Evidence, error) {
	if !filepath.IsAbs(artifactPath) || !filepath.IsAbs(bundlePath) || !validCommit(expect.SourceCommit) || !ValidDigest(expect.SubjectSHA256) || expect.SubjectName == "" || expect.Workflow.Path == "" {
		return Evidence{}, errors.New("attestation expectation is incomplete")
	}
	for _, path := range []string{artifactPath, bundlePath} {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return Evidence{}, errors.New("attestation input is not a regular file")
		}
	}
	arguments := []string{
		"attestation", "verify", artifactPath,
		"--bundle", bundlePath,
		"--repo", Repository,
		"--signer-workflow", Repository + "/" + expect.Workflow.Path,
		"--source-ref", SourceRef,
		"--source-digest", expect.SourceCommit,
		"--predicate-type", PredicateType,
		"--deny-self-hosted-runners",
		"--format", "json",
	}
	result, err := verifier.run(ctx, processadapter.Request{
		Executable: gh.Path, Arguments: arguments, Directory: filepath.Dir(artifactPath),
		Environment: ghEnvironment(), MaxOutputBytes: ghOutputBytes, Timeout: ghTimeout,
	})
	if err != nil {
		return Evidence{}, fmt.Errorf("gh attestation verify did not complete: %w", err)
	}
	if result.ExitCode != 0 {
		return Evidence{}, fmt.Errorf("gh attestation verify rejected %s: %s", expect.SubjectName, diagnostic(result.Stderr))
	}
	return matchVerification(result.Stdout, expect, gh.Path)
}

type verificationResult struct {
	VerificationResult struct {
		Signature struct {
			Certificate map[string]string `json:"certificate"`
		} `json:"signature"`
		Statement struct {
			PredicateType string `json:"predicateType"`
			Subject       []struct {
				Name   string            `json:"name"`
				Digest map[string]string `json:"digest"`
			} `json:"subject"`
		} `json:"statement"`
		VerifiedTimestamps []struct {
			Type      string `json:"type"`
			URI       string `json:"uri"`
			Timestamp string `json:"timestamp"`
		} `json:"verifiedTimestamps"`
	} `json:"verificationResult"`
}

func matchVerification(output []byte, expect Expectation, ghPath string) (Evidence, error) {
	if len(output) == 0 || len(output) > ghOutputBytes || !utf8.Valid(output) {
		return Evidence{}, errors.New("gh attestation verify returned no usable JSON")
	}
	if err := validateJSONShape(output); err != nil {
		return Evidence{}, errors.New("gh attestation verify returned malformed JSON")
	}
	var results []verificationResult
	if err := json.Unmarshal(output, &results); err != nil || len(results) == 0 {
		return Evidence{}, errors.New("gh attestation verify returned no verification result")
	}
	want := map[string]string{
		"issuer":                   oidcIssuer,
		"buildSignerURI":           expect.Workflow.SignerIdentity(),
		"subjectAlternativeName":   expect.Workflow.SignerIdentity(),
		"sourceRepositoryURI":      "https://github.com/" + Repository,
		"sourceRepositoryDigest":   expect.SourceCommit,
		"sourceRepositoryRef":      SourceRef,
		"githubWorkflowRepository": Repository,
		"runnerEnvironment":        "github-hosted",
	}
	var failures []string
	for _, result := range results {
		verification := result.VerificationResult
		mismatch := ""
		for key, value := range want {
			if verification.Signature.Certificate[key] != value {
				mismatch = key
				break
			}
		}
		if mismatch != "" {
			failures = append(failures, "certificate "+mismatch+" mismatch")
			continue
		}
		if verification.Statement.PredicateType != PredicateType {
			failures = append(failures, "predicate type mismatch")
			continue
		}
		subjectFound := false
		for _, subject := range verification.Statement.Subject {
			if subject.Name == expect.SubjectName && subject.Digest["sha256"] == expect.SubjectSHA256 {
				subjectFound = true
			}
		}
		if !subjectFound {
			failures = append(failures, "statement does not cover "+expect.SubjectName)
			continue
		}
		if len(verification.VerifiedTimestamps) == 0 {
			failures = append(failures, "no verified timestamp")
			continue
		}
		return Evidence{
			Verifier:       "gh attestation verify --bundle (" + ghPath + ")",
			SignerIdentity: verification.Signature.Certificate["buildSignerURI"],
			SourceCommit:   verification.Signature.Certificate["sourceRepositoryDigest"],
			RunInvocation:  verification.Signature.Certificate["runInvocationURI"],
			Timestamp:      verification.VerifiedTimestamps[0].Timestamp,
			SubjectSHA256:  expect.SubjectSHA256,
		}, nil
	}
	return Evidence{}, fmt.Errorf("no attestation for %s matched the trusted identity: %s", expect.SubjectName, strings.Join(failures, "; "))
}

// ghEnvironment passes the user's gh configuration location through, never a
// token, so verification behaves identically with or without a login.
func ghEnvironment() []string {
	values := processadapter.MinimalEnvironment()
	for _, key := range []string{"GH_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME"} {
		if value := os.Getenv(key); value != "" && !strings.ContainsAny(value, "\x00\r\n") {
			values = append(values, key+"="+value)
		}
	}
	return append(values, "GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "GH_PAGER=cat")
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
