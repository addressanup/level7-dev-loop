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
	want := "{\"schema\":1,\"outcome\":\"BLOCKED\",\"code\":\"L7-STATUS-001\",\"command\":\"status\",\"state\":\"unavailable\",\"version\":\"test-version\",\"message\":\"not implemented\",\"next\":\"run l7 help\",\"details\":[\"first\"]}\n"
	got, err := JSON(fixtureResult())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("JSON() mismatch\ngot:  %q\nwant: %q", got, want)
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
