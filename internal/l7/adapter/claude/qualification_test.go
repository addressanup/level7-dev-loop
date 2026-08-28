package claude

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

const (
	claudeQualificationTargetBare    = "2.1.247"
	claudeQualificationTargetNamed   = "2.1.247 (Claude Code)"
	claudeQualificationUnknownOption = "--l7-qualification-unknown-option"
	claudeQualificationMaxOutput     = 1 << 20
)

type claudeQualificationRun func(context.Context, processadapter.Request) (processadapter.Result, error)

type claudeQualificationCase struct {
	name         string
	arguments    []string
	wantExitZero bool
}

type claudeQualificationDiagnostic struct {
	name       string
	advertised []string
}

type claudeQualificationReport struct {
	version     string
	diagnostics []claudeQualificationDiagnostic
}

func claudeQualificationRoleCases(role domain.ProviderRole, base []string) ([]claudeQualificationCase, error) {
	if !role.Valid() || len(base) == 0 {
		return nil, errors.New("Claude qualification role or argv is invalid")
	}
	turns := -1
	for index, argument := range base {
		if argument == "--help" || argument == claudeQualificationUnknownOption {
			return nil, errors.New("Claude production argv contains a test-only control")
		}
		if argument != "--max-turns" {
			continue
		}
		if turns >= 0 || index+1 >= len(base) || base[index+1] != "64" {
			return nil, errors.New("Claude production argv must contain exactly one --max-turns 64 pair")
		}
		turns = index
	}
	if turns < 0 {
		return nil, errors.New("Claude production argv is missing --max-turns 64")
	}

	positiveArguments := append(append([]string{}, base...), "--help")
	unknownArguments := append([]string{}, positiveArguments[:len(positiveArguments)-1]...)
	unknownArguments = append(unknownArguments, claudeQualificationUnknownOption, "--help")
	invalidArguments := append([]string{}, base...)
	invalidArguments[turns+1] = "not-an-integer"
	invalidArguments = append(invalidArguments, "--help")
	return []claudeQualificationCase{
		{name: string(role) + "-help", arguments: positiveArguments, wantExitZero: true},
		{name: string(role) + "-unknown-option", arguments: unknownArguments},
		{name: string(role) + "-invalid-max-turns", arguments: invalidArguments},
	}, nil
}

func runClaudeQualification(ctx context.Context, run claudeQualificationRun, executable, root string) (claudeQualificationReport, error) {
	if run == nil || !filepath.IsAbs(executable) || !filepath.IsAbs(root) || strings.ContainsAny(executable+root, "\r\n\x00") {
		return claudeQualificationReport{}, errors.New("Claude qualification runtime binding is invalid")
	}
	versionResult, err := run(ctx, processadapter.Request{
		Executable: executable, Arguments: []string{"--version"}, Directory: root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 10 * time.Second,
	})
	version, err := claudeQualificationVersion(versionResult, err)
	if err != nil {
		return claudeQualificationReport{}, err
	}

	var cases []claudeQualificationCase
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		roleCases, caseErr := claudeQualificationRoleCases(role, arguments(role))
		if caseErr != nil {
			return claudeQualificationReport{}, caseErr
		}
		cases = append(cases, roleCases...)
	}

	report := claudeQualificationReport{version: version}
	for _, observation := range cases {
		result, runErr := run(ctx, processadapter.Request{
			Executable: executable, Arguments: append([]string{}, observation.arguments...), Directory: root,
			Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: claudeQualificationMaxOutput, Timeout: 30 * time.Second,
		})
		output, outcomeErr := claudeQualificationOutcome(observation, result, runErr)
		if outcomeErr != nil {
			return claudeQualificationReport{}, fmt.Errorf("Claude %s observation: %w", observation.name, outcomeErr)
		}
		if observation.wantExitZero {
			report.diagnostics = append(report.diagnostics, claudeQualificationDiagnostic{
				name: observation.name, advertised: claudeQualificationAdvertised(output),
			})
		}
	}
	return report, nil
}

func runClaudeQualificationChecked(ctx context.Context, run claudeQualificationRun, executable, root string, postcheck func() error) (claudeQualificationReport, error) {
	if postcheck == nil {
		return claudeQualificationReport{}, errors.New("Claude qualification postcondition is missing")
	}
	report, qualificationErr := runClaudeQualification(ctx, run, executable, root)
	postconditionErr := postcheck()
	if err := errors.Join(qualificationErr, postconditionErr); err != nil {
		return report, err
	}
	return report, nil
}

func claudeQualificationVersion(result processadapter.Result, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("bounded Claude version invocation failed: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("Claude version exited %d", result.ExitCode)
	}
	output, err := claudeQualificationOutput(result, 256)
	if err != nil {
		return "", fmt.Errorf("Claude version output: %w", err)
	}
	version := strings.TrimSpace(output)
	if (version != claudeQualificationTargetBare && version != claudeQualificationTargetNamed) || strings.ContainsAny(version, "\r\n") {
		return "", fmt.Errorf("Claude version %q does not match an exact diagnostic target", version)
	}
	return version, nil
}

func claudeQualificationOutcome(observation claudeQualificationCase, result processadapter.Result, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("bounded no-model invocation failed: %w", err)
	}
	output, outputErr := claudeQualificationOutput(result, claudeQualificationMaxOutput)
	if outputErr != nil {
		return "", outputErr
	}
	if observation.wantExitZero && result.ExitCode != 0 {
		return "", fmt.Errorf("positive help exited %d", result.ExitCode)
	}
	if !observation.wantExitZero && result.ExitCode < 1 {
		return "", fmt.Errorf("negative parser control exited %d", result.ExitCode)
	}
	return output, nil
}

func claudeQualificationOutput(result processadapter.Result, maximum int) (string, error) {
	output := append(append([]byte{}, result.Stdout...), result.Stderr...)
	if len(output) == 0 || len(output) > maximum || !utf8.Valid(output) || strings.ContainsRune(string(output), 0) || strings.TrimSpace(string(output)) == "" {
		return "", errors.New("diagnostic output is empty, unbounded, or invalid")
	}
	return string(output), nil
}

func claudeQualificationAdvertised(output string) []string {
	var advertised []string
	for _, control := range []string{"--safe-mode", "--disable-slash-commands", "--print", "--input-format", "--max-turns", "--tools", "--disallowedTools", "--permission-mode", "--strict-mcp-config", "--no-chrome", "--no-session-persistence", "--output-format", "--json-schema"} {
		if strings.Contains(output, control) {
			advertised = append(advertised, control)
		}
	}
	return advertised
}

func TestClaudeQualificationCasesCopyProductionArgv(t *testing.T) {
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		base := arguments(role)
		original := append([]string{}, base...)
		observations, err := claudeQualificationRoleCases(role, base)
		if err != nil || len(observations) != 3 {
			t.Fatalf("%s observations=%+v error=%v", role, observations, err)
		}
		if !slices.Equal(base, original) {
			t.Fatalf("%s base argv mutated: got=%v want=%v", role, base, original)
		}
		positive := append(append([]string{}, base...), "--help")
		if !observations[0].wantExitZero || !slices.Equal(observations[0].arguments, positive) {
			t.Fatalf("%s positive observation=%+v", role, observations[0])
		}
		unknown := append([]string{}, positive[:len(positive)-1]...)
		unknown = append(unknown, claudeQualificationUnknownOption, "--help")
		if observations[1].wantExitZero || !slices.Equal(observations[1].arguments, unknown) || claudeQualificationCount(observations[1].arguments, claudeQualificationUnknownOption) != 1 {
			t.Fatalf("%s unknown observation=%+v", role, observations[1])
		}
		turns := slices.Index(base, "--max-turns")
		invalid := append([]string{}, base...)
		invalid[turns+1] = "not-an-integer"
		invalid = append(invalid, "--help")
		if observations[2].wantExitZero || !slices.Equal(observations[2].arguments, invalid) || claudeQualificationCount(observations[2].arguments, "not-an-integer") != 1 {
			t.Fatalf("%s invalid turns observation=%+v", role, observations[2])
		}
	}
}

func TestClaudeQualificationCasesRejectUnsafeBase(t *testing.T) {
	valid := arguments(domain.RoleImplementer)
	missing := append([]string{}, valid...)
	turns := slices.Index(missing, "--max-turns")
	missing = append(missing[:turns], missing[turns+2:]...)
	wrong := append([]string{}, valid...)
	wrong[turns+1] = "63"
	duplicate := append(append([]string{}, valid...), "--max-turns", "64")
	for _, base := range [][]string{
		nil,
		missing,
		wrong,
		duplicate,
		append([]string{"--help"}, valid...),
		append([]string{claudeQualificationUnknownOption}, valid...),
	} {
		if observations, err := claudeQualificationRoleCases(domain.RoleImplementer, base); err == nil || observations != nil {
			t.Fatalf("unsafe Claude base accepted: %v", base)
		}
	}
	if observations, err := claudeQualificationRoleCases(domain.ProviderRole("unknown"), valid); err == nil || observations != nil {
		t.Fatalf("invalid role accepted: %+v", observations)
	}
}

func TestClaudeQualificationUsesOneVersionProbeAndNoInput(t *testing.T) {
	requests := 0
	run := func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		requests++
		if request.Executable != "/provider" || request.Directory != "/repo" || len(request.Input) != 0 || request.MaxOutputBytes < 64<<10 || request.Timeout <= 0 {
			t.Fatalf("unsafe request=%+v", request)
		}
		if slices.Equal(request.Arguments, []string{"--version"}) {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(claudeQualificationTargetNamed + "\n")}, nil
		}
		unknown := claudeQualificationCount(request.Arguments, claudeQualificationUnknownOption) == 1
		invalid := claudeQualificationCount(request.Arguments, "not-an-integer") == 1
		if unknown || invalid {
			return processadapter.Result{ExitCode: 2, Stderr: []byte("parser rejected control\n")}, nil
		}
		if claudeQualificationCount(request.Arguments, "--help") != 1 || claudeQualificationCount(request.Arguments, "--max-turns") != 1 || claudeQualificationCount(request.Arguments, "64") != 1 {
			t.Fatalf("malformed Claude request=%v", request.Arguments)
		}
		return processadapter.Result{ExitCode: 0, Stdout: []byte("bounded help\n")}, nil
	}
	report, err := runClaudeQualification(context.Background(), run, "/provider", "/repo")
	if err != nil || report.version != claudeQualificationTargetNamed || requests != 7 || len(report.diagnostics) != 2 {
		t.Fatalf("report=%+v requests=%d error=%v", report, requests, err)
	}
	for _, diagnostic := range report.diagnostics {
		if len(diagnostic.advertised) != 0 {
			t.Fatalf("missing help advertisement became dispositive: %+v", diagnostic)
		}
	}
}

func TestClaudeQualificationAcceptsBothExactVersionSpellings(t *testing.T) {
	for _, target := range []string{claudeQualificationTargetBare, claudeQualificationTargetNamed} {
		result := processadapter.Result{ExitCode: 0, Stdout: []byte(target + "\n")}
		if version, err := claudeQualificationVersion(result, nil); err != nil || version != target {
			t.Fatalf("version=%q error=%v", version, err)
		}
	}
}

func TestClaudeQualificationOutcomesFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		call   int
		result processadapter.Result
		err    error
	}{
		{name: "version runner error", call: 0, err: errors.New("runner failed")},
		{name: "version nonzero", call: 0, result: processadapter.Result{ExitCode: 2, Stderr: []byte("bad version\n")}},
		{name: "wrong version", call: 0, result: processadapter.Result{ExitCode: 0, Stdout: []byte("2.1.248 (Claude Code)\n")}},
		{name: "implementer help failed", call: 1, result: processadapter.Result{ExitCode: 2, Stderr: []byte("failed\n")}},
		{name: "implementer unknown accepted", call: 2, result: processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}},
		{name: "implementer invalid turns accepted", call: 3, result: processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}},
		{name: "empty reviewer help", call: 4, result: processadapter.Result{ExitCode: 0}},
		{name: "reviewer unknown invalid utf8", call: 5, result: processadapter.Result{ExitCode: 2, Stderr: []byte{0xff}}},
		{name: "ambiguous reviewer invalid exit", call: 6, result: processadapter.Result{ExitCode: -1, Stderr: []byte("ambiguous\n")}},
		{name: "output overflow", call: 1, result: processadapter.Result{ExitCode: -1}, err: processadapter.ErrOutputLimit},
		{name: "timeout", call: 4, result: processadapter.Result{ExitCode: -1}, err: context.DeadlineExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			run := func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
				call := calls
				calls++
				if call == test.call {
					return test.result, test.err
				}
				if slices.Equal(request.Arguments, []string{"--version"}) {
					return processadapter.Result{ExitCode: 0, Stdout: []byte(claudeQualificationTargetNamed + "\n")}, nil
				}
				if claudeQualificationCount(request.Arguments, claudeQualificationUnknownOption) == 1 || claudeQualificationCount(request.Arguments, "not-an-integer") == 1 {
					return processadapter.Result{ExitCode: 2, Stderr: []byte("parser rejected control\n")}, nil
				}
				return processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}, nil
			}
			if report, err := runClaudeQualification(context.Background(), run, "/provider", "/repo"); err == nil {
				t.Fatalf("unsafe outcome passed: %+v", report)
			}
		})
	}
}

func TestClaudeQualificationPostconditionsAlwaysRun(t *testing.T) {
	postchecks := 0
	_, err := runClaudeQualificationChecked(context.Background(), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{}, errors.New("early runner failure")
	}, "/provider", "/repo", func() error {
		postchecks++
		return nil
	})
	if err == nil || postchecks != 1 {
		t.Fatalf("early failure error=%v postchecks=%d", err, postchecks)
	}

	postchecks = 0
	_, err = runClaudeQualificationChecked(context.Background(), claudeQualificationPassingRun, "/provider", "/repo", func() error {
		postchecks++
		return errors.New("source or executable identity drift")
	})
	if err == nil || !strings.Contains(err.Error(), "identity drift") || postchecks != 1 {
		t.Fatalf("postcondition error=%v postchecks=%d", err, postchecks)
	}
}

func TestClaudeQualificationHelpAdvertisementIsNonDispositive(t *testing.T) {
	positive := claudeQualificationCase{name: "positive", wantExitZero: true}
	if _, err := claudeQualificationOutcome(positive, processadapter.Result{ExitCode: 0, Stdout: []byte("help without named controls\n")}, nil); err != nil {
		t.Fatalf("missing advertisement rejected a valid positive outcome: %v", err)
	}
	negative := claudeQualificationCase{name: "negative"}
	advertised := []byte("--safe-mode --max-turns --permission-mode --json-schema\n")
	if _, err := claudeQualificationOutcome(negative, processadapter.Result{ExitCode: 0, Stdout: advertised}, nil); err == nil {
		t.Fatal("advertised controls overrode a successful negative parser control")
	}
	if _, err := claudeQualificationOutcome(negative, processadapter.Result{ExitCode: 2, Stderr: advertised}, nil); err != nil {
		t.Fatalf("advertised controls changed a rejecting parser outcome: %v", err)
	}
}

func claudeQualificationPassingRun(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
	if slices.Equal(request.Arguments, []string{"--version"}) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte(claudeQualificationTargetNamed + "\n")}, nil
	}
	if claudeQualificationCount(request.Arguments, claudeQualificationUnknownOption) == 1 || claudeQualificationCount(request.Arguments, "not-an-integer") == 1 {
		return processadapter.Result{ExitCode: 2, Stderr: []byte("parser rejected control\n")}, nil
	}
	return processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}, nil
}

func claudeQualificationCount(arguments []string, target string) int {
	count := 0
	for _, argument := range arguments {
		if argument == target {
			count++
		}
	}
	return count
}
