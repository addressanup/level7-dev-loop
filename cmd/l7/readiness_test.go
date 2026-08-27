package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authorityadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/authority"
)

func TestHeadlessReadinessIsJSONOnlyAndLeavesNoLocalEffects(t *testing.T) {
	workingDirectory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadDir(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	terminal := authorityadapter.NewTerminal(nil, &stderr, false, "accountable-owner")
	exit := runAtWithTerminalAndInput(context.Background(), []string{"ready", "--headless", "--json"}, workingDirectory, &stdout, &stderr, terminal, strings.NewReader(headlessEnvelope(true)))
	if exit != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"code":"L7-READY-000"`) || !strings.Contains(stdout.String(), `"headless":true`) || !strings.Contains(stdout.String(), `"ready":true`) {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
	after, err := os.ReadDir(workingDirectory)
	if err != nil || len(before) != len(after) || len(after) != 0 {
		t.Fatalf("headless evaluator changed directory: before=%v after=%v error=%v", before, after, err)
	}
}

func TestHeadlessReadinessRejectsMalformedAndFalseReadyInput(t *testing.T) {
	workingDirectory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		input    string
		exit     int
		contains string
	}{
		{"malformed", `{"schema":1,"schema":2}`, 1, `"code":"L7-READY-003"`},
		{"false ready", headlessEnvelope(false), 2, `L7-READY-F008`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			terminal := authorityadapter.NewTerminal(nil, &stderr, false, "accountable-owner")
			exit := runAtWithTerminalAndInput(context.Background(), []string{"ready", "--headless", "--json"}, workingDirectory, &stdout, &stderr, terminal, strings.NewReader(test.input))
			if exit != test.exit || stderr.Len() != 0 || !strings.Contains(stdout.String(), test.contains) {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
			}
		})
	}

	var stdout, stderr bytes.Buffer
	exit := runAtWithTerminalAndInput(context.Background(), []string{"ready", "--headless"}, workingDirectory, &stdout, &stderr, authorityadapter.NewTerminal(nil, &stderr, false, "accountable-owner"), strings.NewReader(headlessEnvelope(true)))
	if exit != 1 || !strings.Contains(stdout.String(), "headless readiness requires --json") {
		t.Fatalf("non-JSON exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func headlessEnvelope(clean bool) string {
	cleanValue := "false"
	if clean {
		cleanValue = "true"
	}
	return `{"schema":1,"change_id":"product-change","tier":3,"base_commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","candidate_commit":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","candidate_tree":"cccccccccccccccccccccccccccccccccccccccc","brief_commit":"dddddddddddddddddddddddddddddddddddddddd","configuration_digest":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","verification_commit":"ffffffffffffffffffffffffffffffffffffffff","review_commit":"1111111111111111111111111111111111111111","scope":["internal/product/**"],"checks":[{"name":"benchmark","benchmark":true,"passed":true,"exit_code":0,"code":"L7-VERIFY-000","message":"passed"}],"owner":"accountable-owner","implementer":"codex","reviewer":"claude","review_decision":"GO","benchmark_required":true,"plan_current":true,"repository_clean":` + cleanValue + `,"approval_current":true,"verification_current":true,"review_current":false,"audit_current":true}`
}
