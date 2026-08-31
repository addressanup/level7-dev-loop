package cyber

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestReadOnlyAuditFindsBoundedIssueAndPersistsReport(t *testing.T) {
	root := repository(t)
	write(t, root, "danger.go", "package danger\nimport \"os/exec\"\nfunc run(x string) { exec.Command(\"sh\", \"-c\", x) }\n")
	git(t, root, "add", "danger.go")
	git(t, root, "-c", "user.name=L7", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "danger")
	common := common(t, root)
	adapter := NewWith(nil, nil, func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) })
	report, err := adapter.Audit(context.Background(), root, common, orchestrationconfig.Default(), false)
	if err != nil || report.Mode != "read-only" || len(report.Findings) != 1 || report.Findings[0].CWE != "CWE-78" || strings.Contains(report.Findings[0].EvidenceDigest, "exec.Command") {
		t.Fatalf("report=%#v err=%v", report, err)
	}
	loaded, err := LoadReport(common, report.ID)
	if err != nil || loaded.RepositoryHead != report.RepositoryHead {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}

func TestActiveAuditFailsClosedWithoutExplicitFeatureAndDigest(t *testing.T) {
	root := repository(t)
	configuration := orchestrationconfig.Default()
	_, err := New().Audit(context.Background(), root, common(t, root), configuration, true)
	if err == nil || !strings.Contains(err.Error(), "default OFF") {
		t.Fatalf("error=%v", err)
	}
	configuration.Features.CyberActive = true
	_, err = New().Audit(context.Background(), root, common(t, root), configuration, true)
	if err == nil || !strings.Contains(err.Error(), "pinned Docker image") {
		t.Fatalf("error=%v", err)
	}
}

func TestActiveIsolationMountsOnlyRedactedDisposableCopy(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "safe.go", "package safe\n")
	write(t, root, ".env", "SECRET=value\n")
	write(t, root, "tracked.txt", "API_KEY=secret-value\n")
	runs := 0
	adapter := NewWith(func(name string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("a", 64)}, nil
	}, func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		runs++
		switch filepath.Base(request.Executable) {
		case "cosign":
			if strings.Contains(strings.Join(request.Arguments, " "), "secret") {
				t.Fatal("secret entered signature command")
			}
			return processadapter.Result{}, nil
		case "git":
			return processadapter.Result{Stdout: []byte("safe.go\x00.env\x00tracked.txt\x00")}, nil
		case "docker":
			arguments := strings.Join(request.Arguments, " ")
			for _, required := range []string{"--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--user=65532:65532", "--pids-limit=256"} {
				if !strings.Contains(arguments, required) {
					t.Fatalf("missing isolation argument %s: %s", required, arguments)
				}
			}
			if strings.Contains(arguments, "docker.sock") || strings.Contains(arguments, root) || request.Directory == root {
				t.Fatalf("original checkout or host socket exposed: %#v", request)
			}
			if _, err := os.Stat(filepath.Join(request.Directory, "safe.go")); err != nil {
				t.Fatal("tracked safe file missing from lab")
			}
			for _, excluded := range []string{".env", "tracked.txt"} {
				if _, err := os.Stat(filepath.Join(request.Directory, excluded)); !os.IsNotExist(err) {
					t.Fatalf("secret-like file entered lab: %s", excluded)
				}
			}
			if err := os.WriteFile(filepath.Join(request.Directory, ".l7-cyber-active.json"), []byte(`{"schema":1,"findings":[]}`), 0o600); err != nil {
				t.Fatal(err)
			}
			return processadapter.Result{}, nil
		default:
			t.Fatalf("unexpected executable: %s", request.Executable)
			return processadapter.Result{}, nil
		}
	}, time.Now)
	policy := orchestrationconfig.Default().Cyber
	policy.ImageDigest = "sha256:" + strings.Repeat("b", 64)
	findings, err := adapter.active(context.Background(), root, policy)
	if err != nil || runs != 3 || len(findings) != 0 {
		t.Fatalf("runs=%d findings=%#v err=%v", runs, findings, err)
	}
}

func TestRemediationIsSeparateTierThreeBrief(t *testing.T) {
	report := domainReport()
	brief, err := RemediationBrief(report)
	if err != nil || !strings.Contains(brief, "Risk tier | `3`") || !strings.Contains(brief, "fresh owner approval") || !strings.Contains(brief, report.Findings[0].ID) {
		t.Fatalf("brief=%q err=%v", brief, err)
	}
}

func domainReport() domain.SecurityReport {
	return domain.SecurityReport{Schema: domain.OrchestrationSchema, ID: "cyber-report", RepositoryHead: strings.Repeat("a", 40), Findings: []domain.SecurityFinding{{ID: "CYBER-1", Path: "main.go", Line: 1, Title: "Issue", Remediation: "Fix it.", VerificationTest: "Test it."}}}
}

func repository(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "-q")
	write(t, root, "README.md", "fixture\n")
	git(t, root, "add", "README.md")
	git(t, root, "-c", "user.name=L7", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "initial")
	return root
}

func common(t *testing.T, root string) string {
	t.Helper()
	value := strings.TrimSpace(git(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(value) {
		value = filepath.Join(root, value)
	}
	return value
}

func write(t *testing.T, root, relative, value string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	command.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return string(output)
}
