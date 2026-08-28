package codex

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
	codexQualificationTarget        = "codex-cli 0.150.1"
	codexQualificationUnknownOption = "--l7-qualification-unknown-option"
	codexQualificationMaxOutput     = 1 << 20
)

type codexQualificationRun func(context.Context, processadapter.Request) (processadapter.Result, error)

type codexQualificationCase struct {
	name         string
	arguments    []string
	wantExitZero bool
}

type codexQualificationDiagnostic struct {
	name       string
	advertised []string
}

type codexQualificationReport struct {
	version     string
	diagnostics []codexQualificationDiagnostic
}

func codexQualificationRoleCases(role domain.ProviderRole, base []string) ([]codexQualificationCase, error) {
	if !role.Valid() || len(base) == 0 {
		return nil, errors.New("Codex qualification role or argv is invalid")
	}
	sentinels := 0
	for index, argument := range base {
		if argument == "-" {
			sentinels++
			if index != len(base)-1 {
				return nil, errors.New("Codex production stdin sentinel is not final")
			}
		}
		if argument == "--help" || argument == codexQualificationUnknownOption {
			return nil, errors.New("Codex production argv contains a test-only control")
		}
	}
	if sentinels != 1 {
		return nil, errors.New("Codex production argv must contain exactly one final stdin sentinel")
	}

	positiveArguments := append([]string{}, base...)
	positiveArguments[len(positiveArguments)-1] = "--help"
	unknownArguments := append([]string{}, positiveArguments[:len(positiveArguments)-1]...)
	unknownArguments = append(unknownArguments, codexQualificationUnknownOption, "--help")
	return []codexQualificationCase{
		{name: string(role) + "-help", arguments: positiveArguments, wantExitZero: true},
		{name: string(role) + "-unknown-option", arguments: unknownArguments},
	}, nil
}

func runCodexQualification(ctx context.Context, run codexQualificationRun, executable, root string) (codexQualificationReport, error) {
	if run == nil || !filepath.IsAbs(executable) || !filepath.IsAbs(root) || strings.ContainsAny(executable+root, "\r\n\x00") {
		return codexQualificationReport{}, errors.New("Codex qualification runtime binding is invalid")
	}
	versionResult, err := run(ctx, processadapter.Request{
		Executable: executable, Arguments: []string{"--version"}, Directory: root,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 10 * time.Second,
	})
	version, err := codexQualificationVersion(versionResult, err)
	if err != nil {
		return codexQualificationReport{}, err
	}

	cases := []codexQualificationCase{{name: "top-level-help", arguments: []string{"--help"}, wantExitZero: true}}
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		roleCases, caseErr := codexQualificationRoleCases(role, arguments(role, root))
		if caseErr != nil {
			return codexQualificationReport{}, caseErr
		}
		cases = append(cases, roleCases...)
	}

	report := codexQualificationReport{version: version}
	for _, observation := range cases {
		result, runErr := run(ctx, processadapter.Request{
			Executable: executable, Arguments: append([]string{}, observation.arguments...), Directory: root,
			Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: codexQualificationMaxOutput, Timeout: 30 * time.Second,
		})
		output, outcomeErr := codexQualificationOutcome(observation, result, runErr)
		if outcomeErr != nil {
			return codexQualificationReport{}, fmt.Errorf("Codex %s observation: %w", observation.name, outcomeErr)
		}
		if observation.wantExitZero {
			report.diagnostics = append(report.diagnostics, codexQualificationDiagnostic{
				name: observation.name, advertised: codexQualificationAdvertised(output),
			})
		}
	}
	return report, nil
}

func runCodexQualificationChecked(ctx context.Context, run codexQualificationRun, executable, root string, postcheck func() error) (codexQualificationReport, error) {
	if postcheck == nil {
		return codexQualificationReport{}, errors.New("Codex qualification postcondition is missing")
	}
	report, qualificationErr := runCodexQualification(ctx, run, executable, root)
	postconditionErr := postcheck()
	if err := errors.Join(qualificationErr, postconditionErr); err != nil {
		return report, err
	}
	return report, nil
}

func codexQualificationVersion(result processadapter.Result, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("bounded Codex version invocation failed: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("Codex version exited %d", result.ExitCode)
	}
	output, err := codexQualificationOutput(result, 256)
	if err != nil {
		return "", fmt.Errorf("Codex version output: %w", err)
	}
	version := strings.TrimSpace(output)
	if version != codexQualificationTarget || strings.ContainsAny(version, "\r\n") {
		return "", fmt.Errorf("Codex version %q does not match the exact diagnostic target", version)
	}
	return version, nil
}

func codexQualificationOutcome(observation codexQualificationCase, result processadapter.Result, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("bounded no-model invocation failed: %w", err)
	}
	output, outputErr := codexQualificationOutput(result, codexQualificationMaxOutput)
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

func codexQualificationOutput(result processadapter.Result, maximum int) (string, error) {
	output := append(append([]byte{}, result.Stdout...), result.Stderr...)
	if len(output) == 0 || len(output) > maximum || !utf8.Valid(output) || strings.ContainsRune(string(output), 0) || strings.TrimSpace(string(output)) == "" {
		return "", errors.New("diagnostic output is empty, unbounded, or invalid")
	}
	return string(output), nil
}

func codexQualificationAdvertised(output string) []string {
	var advertised []string
	for _, control := range []string{"--ask-for-approval", "exec", "--ephemeral", "--sandbox", "--color", "--json", "--cd", "--skip-git-repo-check"} {
		if strings.Contains(output, control) {
			advertised = append(advertised, control)
		}
	}
	return advertised
}

func TestCodexQualificationCasesCopyProductionArgv(t *testing.T) {
	for _, role := range []domain.ProviderRole{domain.RoleImplementer, domain.RoleReviewer} {
		base := arguments(role, "/repo")
		original := append([]string{}, base...)
		observations, err := codexQualificationRoleCases(role, base)
		if err != nil || len(observations) != 2 {
			t.Fatalf("%s observations=%+v error=%v", role, observations, err)
		}
		if !slices.Equal(base, original) {
			t.Fatalf("%s base argv mutated: got=%v want=%v", role, base, original)
		}
		positive := append([]string{}, base...)
		positive[len(positive)-1] = "--help"
		if !observations[0].wantExitZero || !slices.Equal(observations[0].arguments, positive) {
			t.Fatalf("%s positive observation=%+v", role, observations[0])
		}
		unknown := append([]string{}, positive[:len(positive)-1]...)
		unknown = append(unknown, codexQualificationUnknownOption, "--help")
		if observations[1].wantExitZero || !slices.Equal(observations[1].arguments, unknown) || codexQualificationCount(observations[1].arguments, codexQualificationUnknownOption) != 1 {
			t.Fatalf("%s unknown observation=%+v", role, observations[1])
		}
	}
}

func TestCodexQualificationCasesRejectUnsafeBase(t *testing.T) {
	valid := arguments(domain.RoleImplementer, "/repo")
	for _, base := range [][]string{
		nil,
		append([]string{}, valid[:len(valid)-1]...),
		append(append([]string{}, valid...), "-"),
		append([]string{"-"}, valid[:len(valid)-1]...),
		append([]string{"--help"}, valid...),
		append([]string{codexQualificationUnknownOption}, valid...),
	} {
		if observations, err := codexQualificationRoleCases(domain.RoleImplementer, base); err == nil || observations != nil {
			t.Fatalf("unsafe Codex base accepted: %v", base)
		}
	}
	if observations, err := codexQualificationRoleCases(domain.ProviderRole("unknown"), valid); err == nil || observations != nil {
		t.Fatalf("invalid role accepted: %+v", observations)
	}
}

func TestCodexQualificationUsesOneVersionProbeAndNoInput(t *testing.T) {
	requests := 0
	run := func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		requests++
		if request.Executable != "/provider" || request.Directory != "/repo" || len(request.Input) != 0 || request.MaxOutputBytes < 64<<10 || request.Timeout <= 0 {
			t.Fatalf("unsafe request=%+v", request)
		}
		if slices.Equal(request.Arguments, []string{"--version"}) {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(codexQualificationTarget + "\n")}, nil
		}
		if codexQualificationCount(request.Arguments, codexQualificationUnknownOption) == 1 {
			return processadapter.Result{ExitCode: 2, Stderr: []byte("unknown option\n")}, nil
		}
		if codexQualificationCount(request.Arguments, "-") != 0 || codexQualificationCount(request.Arguments, "--help") != 1 {
			t.Fatalf("semantic or malformed Codex request=%v", request.Arguments)
		}
		return processadapter.Result{ExitCode: 0, Stdout: []byte("bounded help\n")}, nil
	}
	report, err := runCodexQualification(context.Background(), run, "/provider", "/repo")
	if err != nil || report.version != codexQualificationTarget || requests != 6 || len(report.diagnostics) != 3 {
		t.Fatalf("report=%+v requests=%d error=%v", report, requests, err)
	}
	for _, diagnostic := range report.diagnostics {
		if len(diagnostic.advertised) != 0 {
			t.Fatalf("missing help advertisement became dispositive: %+v", diagnostic)
		}
	}
}

func TestCodexQualificationOutcomesFailClosed(t *testing.T) {
	tests := []struct {
		name   string
		call   int
		result processadapter.Result
		err    error
	}{
		{name: "version runner error", call: 0, err: errors.New("runner failed")},
		{name: "version nonzero", call: 0, result: processadapter.Result{ExitCode: 2, Stderr: []byte("bad version\n")}},
		{name: "wrong version", call: 0, result: processadapter.Result{ExitCode: 0, Stdout: []byte("codex-cli 0.150.2\n")}},
		{name: "top help failed", call: 1, result: processadapter.Result{ExitCode: 2, Stderr: []byte("failed\n")}},
		{name: "empty help", call: 2, result: processadapter.Result{ExitCode: 0}},
		{name: "unknown accepted", call: 3, result: processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}},
		{name: "invalid utf8", call: 4, result: processadapter.Result{ExitCode: 0, Stdout: []byte{0xff}}},
		{name: "ambiguous negative exit", call: 5, result: processadapter.Result{ExitCode: -1, Stderr: []byte("ambiguous\n")}},
		{name: "output overflow", call: 2, result: processadapter.Result{ExitCode: -1}, err: processadapter.ErrOutputLimit},
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
					return processadapter.Result{ExitCode: 0, Stdout: []byte(codexQualificationTarget + "\n")}, nil
				}
				if codexQualificationCount(request.Arguments, codexQualificationUnknownOption) == 1 {
					return processadapter.Result{ExitCode: 2, Stderr: []byte("unknown option\n")}, nil
				}
				return processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}, nil
			}
			if report, err := runCodexQualification(context.Background(), run, "/provider", "/repo"); err == nil {
				t.Fatalf("unsafe outcome passed: %+v", report)
			}
		})
	}
}

func TestCodexQualificationPostconditionsAlwaysRun(t *testing.T) {
	postchecks := 0
	_, err := runCodexQualificationChecked(context.Background(), func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{}, errors.New("early runner failure")
	}, "/provider", "/repo", func() error {
		postchecks++
		return nil
	})
	if err == nil || postchecks != 1 {
		t.Fatalf("early failure error=%v postchecks=%d", err, postchecks)
	}

	postchecks = 0
	_, err = runCodexQualificationChecked(context.Background(), codexQualificationPassingRun, "/provider", "/repo", func() error {
		postchecks++
		return errors.New("source or executable identity drift")
	})
	if err == nil || !strings.Contains(err.Error(), "identity drift") || postchecks != 1 {
		t.Fatalf("postcondition error=%v postchecks=%d", err, postchecks)
	}
}

func TestCodexQualificationHelpAdvertisementIsNonDispositive(t *testing.T) {
	positive := codexQualificationCase{name: "positive", wantExitZero: true}
	if _, err := codexQualificationOutcome(positive, processadapter.Result{ExitCode: 0, Stdout: []byte("help without named controls\n")}, nil); err != nil {
		t.Fatalf("missing advertisement rejected a valid positive outcome: %v", err)
	}
	negative := codexQualificationCase{name: "negative"}
	advertised := []byte("--ask-for-approval --sandbox --json --cd\n")
	if _, err := codexQualificationOutcome(negative, processadapter.Result{ExitCode: 0, Stdout: advertised}, nil); err == nil {
		t.Fatal("advertised controls overrode a successful negative parser control")
	}
	if _, err := codexQualificationOutcome(negative, processadapter.Result{ExitCode: 2, Stderr: advertised}, nil); err != nil {
		t.Fatalf("advertised controls changed a rejecting parser outcome: %v", err)
	}
}

func codexQualificationPassingRun(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
	if slices.Equal(request.Arguments, []string{"--version"}) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte(codexQualificationTarget + "\n")}, nil
	}
	if codexQualificationCount(request.Arguments, codexQualificationUnknownOption) == 1 {
		return processadapter.Result{ExitCode: 2, Stderr: []byte("unknown option\n")}, nil
	}
	return processadapter.Result{ExitCode: 0, Stdout: []byte("help\n")}, nil
}

func codexQualificationCount(arguments []string, target string) int {
	count := 0
	for _, argument := range arguments {
		if argument == target {
			count++
		}
	}
	return count
}
