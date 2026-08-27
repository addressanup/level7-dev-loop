package presentation

import (
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestTextIsDecisionFirstAndEscapesUntrustedValues(t *testing.T) {
	result := fixtureResult()
	result.Message = "line one\n\"line two\""
	want := "BLOCKED code=\"L7-STATUS-001\" command=\"status\" state=\"unavailable\" version=\"test-version\" message=\"line one\\n\\\"line two\\\"\" next=\"run l7 help\"\n" +
		"detail=\"first\"\n"
	if got := string(Text(result)); got != want {
		t.Fatalf("Text() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestJSONHasStableSchemaAndFieldOrder(t *testing.T) {
	want := "{\"schema\":4,\"outcome\":\"BLOCKED\",\"code\":\"L7-STATUS-001\",\"command\":\"status\",\"state\":\"unavailable\",\"version\":\"test-version\",\"message\":\"not implemented\",\"next\":\"run l7 help\",\"details\":[\"first\"]}\n"
	got, err := JSON(fixtureResult())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("JSON() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestExecutionDetailsAreStableInTextAndJSON(t *testing.T) {
	result := fixtureResult()
	result.Execution = &domain.ExecutionDetails{
		Role: domain.RoleReviewer, Provider: domain.ProviderClaude, Executable: "/usr/bin/claude", Version: "2.1.241", Digest: strings.Repeat("a", 64),
		Commit: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40), Decision: domain.DecisionGO,
		Checks: []domain.CheckResult{{Name: "test", Passed: true, ExitCode: 0, Code: "L7-VERIFY-000", Message: "command passed"}},
	}
	text := string(Text(result))
	for _, value := range []string{`execution_role="reviewer"`, `provider="claude"`, `review_decision="GO"`, `check_name="test" check_passed=true`} {
		if !strings.Contains(text, value) {
			t.Fatalf("Text()=%q, want %q", text, value)
		}
	}
	data, err := JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{`"execution":{"role":"reviewer","provider":"claude"`, `"decision":"GO"`, `"checks":[{"name":"test","benchmark":false,"passed":true`} {
		if !strings.Contains(string(data), value) {
			t.Fatalf("JSON()=%s, want %s", data, value)
		}
	}
}

func TestRepositoryDetailsAreStableAndEscapedInTextAndJSON(t *testing.T) {
	result := fixtureResult()
	result.Repository = &domain.RepositoryDetails{
		Root: "/repo with space", CommonDir: "/repo with space/.git", ChangeID: "change", Tier: domain.TierProduct,
		Base: "base", Head: "head", Tree: "tree", DeclaredScope: []string{"internal/**"}, ChangedPaths: []string{"internal/file.go"}, ExpandedPaths: []string{"line\nbreak"},
	}
	text := string(Text(result))
	for _, value := range []string{`repository_root="/repo with space"`, `risk_tier=2`, `declared_scope="internal/**"`, `expanded_path="line\nbreak"`} {
		if !strings.Contains(text, value) {
			t.Fatalf("Text()=%q, want %q", text, value)
		}
	}
	data, err := JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{`"repository":{"root":"/repo with space"`, `"tier":2`, `"declared_scope":["internal/**"]`, `"expanded_paths":["line\nbreak"]`} {
		if !strings.Contains(string(data), value) {
			t.Fatalf("JSON()=%s, want %s", data, value)
		}
	}
}

func TestReadinessDetailsAreStableInTextAndJSON(t *testing.T) {
	result := fixtureResult()
	result.Readiness = &domain.ReadinessDetails{
		Headless: true, Ready: true, Base: strings.Repeat("a", 40), Candidate: strings.Repeat("b", 40), Tree: strings.Repeat("c", 40),
		BriefCommit: strings.Repeat("d", 40), ConfigurationDigest: strings.Repeat("e", 64), VerificationCommit: strings.Repeat("f", 40),
		ReviewCommit: strings.Repeat("1", 40), Owner: "accountable-owner", Implementer: domain.ProviderCodex, Reviewer: domain.ProviderClaude,
		Checks: []domain.CheckResult{{Name: "benchmark", Benchmark: true, Passed: true}},
	}
	text := string(Text(result))
	for _, value := range []string{"readiness_headless=true", "readiness_ready=true", `readiness_implementer="codex"`, `readiness_check="benchmark" benchmark=true passed=true`} {
		if !strings.Contains(text, value) {
			t.Fatalf("Text()=%q, want %q", text, value)
		}
	}
	data, err := JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{`"readiness":{"headless":true,"ready":true`, `"implementer":"codex"`, `"checks":[{"name":"benchmark","benchmark":true,"passed":true`} {
		if !strings.Contains(string(data), value) {
			t.Fatalf("JSON()=%s, want %s", data, value)
		}
	}
}

func TestJSONUsesAnEmptyArrayForNoDetails(t *testing.T) {
	result := fixtureResult()
	result.Details = nil
	got, err := JSON(result)
	if err != nil {
		t.Fatal(err)
	}
	want := "\"details\":[]"
	if !strings.Contains(string(got), want) {
		t.Fatalf("JSON()=%s, want %s", got, want)
	}
}

func fixtureResult() domain.Result {
	return domain.Result{
		Schema:  domain.ResultSchema,
		Outcome: domain.OutcomeBlocked,
		Code:    "L7-STATUS-001",
		Command: "status",
		State:   "unavailable",
		Version: "test-version",
		Message: "not implemented",
		Next:    "run l7 help",
		Details: []string{"first"},
	}
}
