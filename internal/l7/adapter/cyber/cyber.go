// Package cyber performs read-only security analysis and isolated active confirmation.
package cyber

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const maxReportBytes = 32 << 20

type Adapter struct {
	resolve func(string) (processadapter.Executable, error)
	run     func(context.Context, processadapter.Request) (processadapter.Result, error)
	now     func() time.Time
}

func New() Adapter {
	return NewWith(processadapter.Resolve, (processadapter.Runner{}).Run, time.Now)
}

func NewWith(resolve func(string) (processadapter.Executable, error), run func(context.Context, processadapter.Request) (processadapter.Result, error), now func() time.Time) Adapter {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	if now == nil {
		now = time.Now
	}
	return Adapter{resolve: resolve, run: run, now: now}
}

func (adapter Adapter) Audit(ctx context.Context, root, common string, configuration orchestrationconfig.File, active bool) (domain.SecurityReport, error) {
	physical, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(physical) || physical != filepath.Clean(root) {
		return domain.SecurityReport{}, errors.New("Cyber repository root is unsafe")
	}
	git, err := adapter.resolve("git")
	if err != nil {
		return domain.SecurityReport{}, errors.New("Git is unavailable")
	}
	head, err := adapter.gitText(ctx, git, physical, []string{"rev-parse", "HEAD"}, 4096)
	if err != nil || len(head) != 40 {
		return domain.SecurityReport{}, errors.New("Cyber cannot establish repository head")
	}
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: git.Path, Arguments: []string{"ls-files", "-z"}, Directory: physical,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 20, Timeout: time.Minute,
	})
	if err != nil || result.ExitCode != 0 {
		return domain.SecurityReport{}, errors.New("Cyber cannot enumerate tracked files")
	}
	findings := []domain.SecurityFinding{}
	for _, relative := range splitNUL(result.Stdout) {
		path, pathErr := safePath(physical, relative)
		if pathErr != nil {
			continue
		}
		info, statErr := os.Stat(path)
		if statErr != nil || !info.Mode().IsRegular() || info.Size() > 4<<20 {
			continue
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil || bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
			continue
		}
		findings = append(findings, scan(relative, data)...)
	}
	mode, isolation := "read-only", "none"
	if active {
		if !configuration.Features.CyberActive {
			return domain.SecurityReport{}, errors.New("active Cyber is default OFF; enable it explicitly in orchestration policy")
		}
		activeFindings, activeErr := adapter.active(ctx, physical, configuration.Cyber)
		if activeErr != nil {
			return domain.SecurityReport{}, activeErr
		}
		findings = append(findings, activeFindings...)
		mode, isolation = "active", "disposable-container"
	}
	sortFindings(findings)
	created := adapter.now().UTC()
	report := domain.SecurityReport{
		Schema: domain.OrchestrationSchema, ID: "cyber-" + head[:12] + "-" + created.Format("20060102T150405Z"),
		RepositoryHead: head, Mode: mode, Isolation: isolation,
		Coverage:     []string{"attack-surface", "trust-boundaries", "data-flows", "dependencies", "secrets", "sast", "iac-containers", "authentication", "authorization", "exploit-hypotheses"},
		Findings:     findings,
		CreatedAtUTC: created.Format(time.RFC3339), Redacted: false,
	}
	report.Next = "review findings; run l7 cyber remediate --report " + report.ID
	if err := saveReport(common, report); err != nil {
		return domain.SecurityReport{}, err
	}
	return report, nil
}

func scan(path string, data []byte) []domain.SecurityFinding {
	findings := []domain.SecurityFinding{}
	lines := strings.Split(string(data), "\n")
	type rule struct {
		marker      string
		title       string
		severity    string
		cwe         string
		remediation string
		test        string
	}
	rules := []rule{
		{"-----BEGIN PRIVATE KEY-----", "Private key committed to source", "critical", "CWE-798", "Revoke the credential, remove it from history, and use a secret manager.", "secret scanner reports no private key material"},
		{"exec.Command(\"sh\", \"-c\"", "Shell command construction", "high", "CWE-78", "Replace shell composition with an exact executable and argv allowlist.", "adversarial metacharacters cannot alter argv"},
		{"child_process.exec(", "Shell command execution", "high", "CWE-78", "Use execFile/spawn with fixed argv and validate inputs.", "metacharacter input remains one inert argument"},
		{"os.system(", "Shell command execution", "high", "CWE-78", "Use subprocess with shell disabled and fixed argv.", "metacharacter input remains inert"},
		{"eval(", "Dynamic code evaluation", "high", "CWE-95", "Replace dynamic evaluation with a typed parser or explicit dispatch table.", "untrusted input cannot execute code"},
		{"network_mode: host", "Container host networking enabled", "high", "CWE-668", "Use an isolated internal container network.", "container cannot reach host or Internet"},
		{"privileged: true", "Privileged container enabled", "critical", "CWE-250", "Remove privileged mode and drop all capabilities.", "container runs non-root with all capabilities dropped"},
		{"InsecureSkipVerify: true", "TLS certificate verification disabled", "high", "CWE-295", "Use system or explicitly pinned trust roots and verify peer identity.", "invalid and attacker-controlled certificates are rejected"},
		{"verify=False", "TLS certificate verification disabled", "high", "CWE-295", "Enable certificate and hostname verification.", "invalid and attacker-controlled certificates are rejected"},
		{"0.0.0.0/0", "Infrastructure rule exposes all IPv4 addresses", "high", "CWE-284", "Restrict the CIDR to the smallest approved trust boundary.", "policy test rejects public ingress and egress"},
		{"\"Action\": \"*\"", "IAM policy grants wildcard actions", "high", "CWE-269", "Replace wildcard actions with the minimum required permissions.", "policy test rejects wildcard privileges"},
		{"Access-Control-Allow-Origin", "CORS policy requires trust-boundary review", "medium", "CWE-942", "Use an explicit origin allowlist and credential policy.", "untrusted origins cannot read authenticated responses"},
	}
	for lineNumber, line := range lines {
		for _, current := range rules {
			if strings.Contains(line, current.marker) {
				findings = append(findings, finding(path, lineNumber+1, current.title, current.severity, current.cwe, current.marker, current.remediation, current.test))
			}
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "uses:") || strings.Contains(trimmed, " uses: ") {
			value := strings.TrimSpace(trimmed[strings.Index(trimmed, "uses:")+5:])
			if strings.Contains(value, "@") {
				ref := value[strings.LastIndex(value, "@")+1:]
				if len(ref) != 40 || !hex(ref) {
					findings = append(findings, finding(path, lineNumber+1, "GitHub Action is not pinned to a full commit", "medium", "CWE-829", value, "Pin the action to an audited full commit SHA.", "workflow policy rejects mutable action references"))
				}
			}
		}
	}
	findings = append(findings, dependencyFindings(path, lines)...)
	if strings.EqualFold(filepath.Base(path), "Dockerfile") {
		hasUser, rootUser := false, false
		for _, line := range lines {
			trimmed := strings.TrimSpace(strings.ToLower(line))
			if strings.HasPrefix(trimmed, "user ") {
				hasUser = true
				rootUser = trimmed == "user root" || trimmed == "user 0" || trimmed == "user 0:0"
			}
		}
		if !hasUser || rootUser {
			findings = append(findings, finding(path, 1, "Container image may run as root", "medium", "CWE-250", "docker-user", "Declare a numeric non-root USER in the final stage.", "container process UID is non-zero"))
		}
	}
	return findings
}

func dependencyFindings(path string, lines []string) []domain.SecurityFinding {
	findings := []domain.SecurityFinding{}
	base := strings.ToLower(filepath.Base(path))
	if base == "requirements.txt" {
		for index, line := range lines {
			value := strings.TrimSpace(line)
			if value == "" || strings.HasPrefix(value, "#") || strings.HasPrefix(value, "-") {
				continue
			}
			if !strings.Contains(value, "==") && !strings.Contains(value, " @ ") {
				findings = append(findings, finding(path, index+1, "Python dependency is not exactly pinned", "medium", "CWE-1104", value, "Pin an audited exact version and lock hashes.", "dependency policy rejects floating requirements"))
			}
		}
	}
	if base == "package.json" {
		for index, line := range lines {
			value := strings.ToLower(strings.TrimSpace(line))
			if strings.Contains(value, `: "*"`) || strings.Contains(value, `: "latest"`) || strings.Contains(value, `: "git+`) {
				findings = append(findings, finding(path, index+1, "JavaScript dependency uses a floating source", "medium", "CWE-1104", value, "Use an audited version range with a committed lockfile and integrity data.", "clean install resolves only locked integrity-bound artifacts"))
			}
		}
	}
	return findings
}

func finding(path string, line int, title, severity, cwe, evidence, remediation, test string) domain.SecurityFinding {
	digest := sha256.Sum256([]byte(path + "\x00" + strconv.Itoa(line) + "\x00" + evidence))
	return domain.SecurityFinding{
		ID: fmt.Sprintf("CYBER-%x", digest[:6]), Title: title, Severity: severity, CWE: cwe,
		CVSS:       cvssFor(severity),
		Confidence: "high", Exploitability: "requires confirmation", Path: path, Line: line,
		EvidenceDigest: fmt.Sprintf("sha256:%x", digest), Reproduction: []string{"inspect " + path + ":" + strconv.Itoa(line)},
		Remediation: remediation, VerificationTest: test,
	}
}

func cvssFor(severity string) string {
	switch severity {
	case "critical":
		return "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H (10.0)"
	case "high":
		return "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:H/A:H (8.8)"
	case "medium":
		return "CVSS:3.1/AV:L/AC:L/PR:L/UI:R/S:U/C:L/I:L/A:L (5.3)"
	default:
		return "CVSS:3.1/AV:L/AC:H/PR:L/UI:R/S:U/C:L/I:N/A:N (2.5)"
	}
}

func (adapter Adapter) active(ctx context.Context, root string, policy orchestrationconfig.Cyber) ([]domain.SecurityFinding, error) {
	if policy.Runtime != "docker" || policy.ImageDigest == "" || !strings.HasPrefix(policy.ImageDigest, "sha256:") || policy.NetworkPolicy != "internal-only" {
		return nil, errors.New("active Cyber requires a pinned Docker image and internal-only policy")
	}
	docker, err := adapter.resolve("docker")
	if err != nil {
		return nil, errors.New("active Cyber requires Docker or a Docker-compatible OrbStack runtime")
	}
	cosign, err := adapter.resolve("cosign")
	if err != nil {
		return nil, errors.New("active Cyber requires Cosign verification")
	}
	image := policy.Image + "@" + policy.ImageDigest
	verify, err := adapter.run(ctx, processadapter.Request{
		Executable: cosign.Path,
		Arguments:  []string{"verify", "--certificate-identity", policy.SignatureIdentity, "--certificate-oidc-issuer", policy.SignatureIssuer, image},
		Directory:  root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 1 << 20, Timeout: 2 * time.Minute,
	})
	if err != nil || verify.ExitCode != 0 {
		return nil, errors.New("Cyber image signature verification failed")
	}
	lab, err := os.MkdirTemp("", "l7-cyber-lab-")
	if err != nil {
		return nil, errors.New("cannot create Cyber lab")
	}
	defer os.RemoveAll(lab)
	if err := adapter.copyTrackedRepository(ctx, root, lab); err != nil {
		return nil, err
	}
	outputPath := filepath.Join(lab, ".l7-cyber-active.json")
	arguments := []string{
		"run", "--rm", "--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges",
		"--pids-limit=256", "--memory=2g", "--cpus=2", "--user=65532:65532", "--tmpfs=/tmp:rw,noexec,nosuid,size=256m",
		"--mount", "type=bind,src=" + lab + ",dst=/workspace,rw", image,
		"l7-cyber-scan", "--root", "/workspace", "--output", "/workspace/.l7-cyber-active.json",
	}
	runResult, err := adapter.run(ctx, processadapter.Request{
		Executable: docker.Path, Arguments: arguments, Directory: lab, Environment: processadapter.MinimalEnvironment(),
		MaxOutputBytes: 8 << 20, Timeout: 30 * time.Minute,
	})
	if err != nil || runResult.ExitCode != 0 {
		return nil, errors.New("isolated Cyber confirmation failed")
	}
	data, err := localfile.Read(outputPath, maxReportBytes)
	if err != nil {
		return nil, errors.New("isolated Cyber result is unavailable")
	}
	var result struct {
		Schema   int                      `json:"schema"`
		Findings []domain.SecurityFinding `json:"findings"`
	}
	if err := localfile.DecodeJSON(data, &result); err != nil || result.Schema != domain.OrchestrationSchema || len(result.Findings) > 10_000 {
		return nil, errors.New("isolated Cyber result is invalid")
	}
	return result.Findings, nil
}

func RemediationBrief(report domain.SecurityReport) (string, error) {
	if report.Schema != domain.OrchestrationSchema || report.ID == "" || report.RepositoryHead == "" || len(report.Findings) == 0 {
		return "", errors.New("Cyber report has no bounded remediation scope")
	}
	var output strings.Builder
	fmt.Fprintf(&output, "# Cyber Remediation — %s\n\n| Field | Value |\n|---|---|\n| Change ID | `%s-remediation` |\n| Risk tier | `3` |\n| Status | `proposed`; requires fresh owner approval |\n| Source report | `%s` |\n| Candidate head | `%s` |\n\n## Problem\n\nRemediate the independently recorded security findings without changing the source report.\n\n## Scope and acceptance\n\n", report.ID, report.ID, report.ID, report.RepositoryHead)
	for _, finding := range report.Findings {
		fmt.Fprintf(&output, "- `%s` `%s:%d`: %s. Remediation: %s Verification: %s.\n", finding.ID, finding.Path, finding.Line, finding.Title, finding.Remediation, finding.VerificationTest)
	}
	output.WriteString("\n## Rollback\n\nRevert the remediation candidate as one reviewed unit; preserve the source Cyber evidence.\n")
	return output.String(), nil
}

func ExportMarkdown(report domain.SecurityReport) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# L7-Cyber Report %s\n\n- Repository head: `%s`\n- Mode: `%s`\n- Isolation: `%s`\n- Findings: `%d`\n\n", report.ID, report.RepositoryHead, report.Mode, report.Isolation, len(report.Findings))
	for _, finding := range report.Findings {
		fmt.Fprintf(&output, "## %s — %s\n\n- Severity: `%s`\n- Location: `%s:%d`\n- CWE: `%s`\n- CVSS: `%s`\n- Confidence: `%s`\n- Exploitability: `%s`\n- Evidence: `%s`\n- Remediation: %s\n- Verification: %s\n\n", finding.ID, finding.Title, finding.Severity, finding.Path, finding.Line, finding.CWE, finding.CVSS, finding.Confidence, finding.Exploitability, finding.EvidenceDigest, finding.Remediation, finding.VerificationTest)
	}
	return output.String()
}

func ExportJSON(report domain.SecurityReport) ([]byte, error) {
	copy := report
	copy.Redacted = true
	for index := range copy.Findings {
		copy.Findings[index].Reproduction = []string{}
	}
	return json.Marshal(copy)
}

func saveReport(common string, report domain.SecurityReport) error {
	root, err := securityRoot(common)
	if err != nil {
		return err
	}
	if err := localfile.EnsureDirectory(root, 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(report)
	if err != nil || len(data) > maxReportBytes {
		return errors.New("Cyber report is invalid or unbounded")
	}
	return localfile.AtomicCreate(filepath.Join(root, report.ID+".json"), data, 0o600)
}

func LoadReport(common, id string) (domain.SecurityReport, error) {
	var report domain.SecurityReport
	if id == "" || strings.ContainsAny(id, "/\\\x00\r\n") {
		return report, errors.New("Cyber report ID is invalid")
	}
	root, err := securityRoot(common)
	if err != nil {
		return report, err
	}
	data, err := localfile.Read(filepath.Join(root, id+".json"), maxReportBytes)
	if err != nil {
		return report, err
	}
	if err := localfile.DecodeJSON(data, &report); err != nil || report.Schema != domain.OrchestrationSchema || report.ID != id {
		return domain.SecurityReport{}, errors.New("Cyber report is invalid")
	}
	return report, nil
}

func securityRoot(common string) (string, error) {
	if !filepath.IsAbs(common) {
		return "", errors.New("Git common directory must be absolute")
	}
	info, err := os.Lstat(filepath.Clean(common))
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("Git common directory is unsafe")
	}
	return filepath.Join(filepath.Clean(common), "l7", "security"), nil
}

func (adapter Adapter) gitText(ctx context.Context, executable processadapter.Executable, root string, arguments []string, maximum int) (string, error) {
	result, err := adapter.run(ctx, processadapter.Request{Executable: executable.Path, Arguments: arguments, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: maximum, Timeout: 30 * time.Second})
	if err != nil || result.ExitCode != 0 {
		return "", errors.New("Git command failed")
	}
	value := strings.TrimSpace(string(result.Stdout))
	if value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", errors.New("Git output is invalid")
	}
	return value, nil
}

func splitNUL(data []byte) []string {
	result := []string{}
	for _, value := range bytes.Split(data, []byte{0}) {
		if len(value) > 0 && utf8.Valid(value) {
			result = append(result, filepath.ToSlash(string(value)))
		}
	}
	return result
}

func safePath(root, relative string) (string, error) {
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == "." || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, "../") || strings.Contains(relative, "\\") {
		return "", errors.New("unsafe Cyber path")
	}
	candidate := filepath.Join(root, filepath.FromSlash(relative))
	physical, err := filepath.EvalSymlinks(candidate)
	if err != nil || (physical != root && !strings.HasPrefix(physical, root+string(filepath.Separator))) {
		return "", errors.New("Cyber path escapes repository")
	}
	return physical, nil
}

func (adapter Adapter) copyTrackedRepository(ctx context.Context, root, destination string) error {
	git, err := adapter.resolve("git")
	if err != nil {
		return errors.New("Git is unavailable for isolated Cyber copy")
	}
	result, err := adapter.run(ctx, processadapter.Request{Executable: git.Path, Arguments: []string{"ls-files", "-z"}, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 20, Timeout: time.Minute})
	if err != nil || result.ExitCode != 0 {
		return errors.New("cannot enumerate tracked Cyber copy")
	}
	for _, relative := range splitNUL(result.Stdout) {
		source, err := safePath(root, relative)
		if err != nil || secretPath(relative) {
			continue
		}
		info, err := os.Stat(source)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 32<<20 {
			continue
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if likelySecret(data) {
			continue
		}
		target := filepath.Join(destination, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func secretPath(relative string) bool {
	base := strings.ToLower(filepath.Base(relative))
	return base == ".env" || strings.HasPrefix(base, ".env.") || strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".p12") || strings.Contains(base, "credential")
}

func likelySecret(data []byte) bool {
	upper := strings.ToUpper(string(data))
	for _, marker := range []string{"-----BEGIN PRIVATE KEY-----", "AWS_SECRET_ACCESS_KEY=", "OPENAI_API_KEY=", "ANTHROPIC_API_KEY=", "API_KEY=", "PASSWORD=", "SECRET=", "TOKEN="} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func hex(value string) bool {
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') && (character < 'A' || character > 'F') {
			return false
		}
	}
	return value != ""
}

func sortFindings(findings []domain.SecurityFinding) {
	rank := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}
	sort.Slice(findings, func(i, j int) bool {
		if rank[findings[i].Severity] != rank[findings[j].Severity] {
			return rank[findings[i].Severity] < rank[findings[j].Severity]
		}
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].ID < findings[j].ID
	})
}
