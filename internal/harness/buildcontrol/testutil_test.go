package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func findingRules(findings []finding) map[string]int {
	rules := make(map[string]int)
	for _, item := range findings {
		rules[item.rule]++
	}
	return rules
}

func materializeMapFS(t *testing.T, fixture fstest.MapFS) string {
	t.Helper()
	root := t.TempDir()
	err := fs.WalkDir(fixture, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		data, err := fs.ReadFile(fixture, name)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		t.Fatalf("materialize fixture: %v", err)
	}
	return root
}

func TestFindingNormalizationAndOrderAreDeterministic(t *testing.T) {
	t.Parallel()
	findings := []finding{
		newFinding("Z-002", "b", "second", "next"),
		newFinding("A-001", "z", "first", "next"),
		newFinding("A-001", "a", "first", "next"),
	}
	sortFindings(findings)
	if findings[0].subject != "a" || findings[1].subject != "z" || findings[2].rule != "Z-002" {
		t.Fatalf("unexpected finding order: %+v", findings)
	}
	normalized := safeASCII("line\n"+strings.Repeat("x", 300), 12)
	if normalized != "line?xxxxxxx" || len(normalized) != 12 {
		t.Fatalf("unexpected bounded normalization: %q", normalized)
	}
	if maxFindings <= 0 || maxMessageSize <= 0 {
		t.Fatal("diagnostic bounds must be positive")
	}
}

func TestFindingCollectionIsBoundedBeforeAppend(t *testing.T) {
	t.Parallel()
	var findings []finding
	for index := 0; index < maxCollectedFindings+100; index++ {
		findings = appendFindings(findings, newFinding("TEST-001", "fixture", "bounded", "none"))
	}
	if len(findings) != maxCollectedFindings {
		t.Fatalf("finding collection size: got %d, want %d", len(findings), maxCollectedFindings)
	}
}

func TestRenderedFindingsAreOrderedCappedAndRepeatable(t *testing.T) {
	t.Parallel()
	var findings []finding
	for index := maxCollectedFindings + 100; index >= 0; index-- {
		findings = appendFindings(findings, newFinding("TEST-001", strings.Repeat("x", index%120), strings.Repeat("m", maxMessageSize), "repair"))
	}
	first := renderFindings(append([]finding(nil), findings...))
	second := renderFindings(append([]finding(nil), findings...))
	if first != second {
		t.Fatal("rendered findings differ across identical runs")
	}
	if len(first) > maxOutputBytes {
		t.Fatalf("rendered findings use %d bytes, limit %d", len(first), maxOutputBytes)
	}
	if !strings.Contains(first, "rule=BCTL-099") {
		t.Fatalf("finding-cap diagnostic is absent: %q", first)
	}
}

func TestSuccessOutputBindsVersionAndExactSourceDigests(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	first, findings := loadSuccessSourceDigests(root)
	if len(findings) != 0 {
		t.Fatalf("source digest findings: %+v", findings)
	}
	second, findings := loadSuccessSourceDigests(root)
	if len(findings) != 0 || second != first {
		t.Fatalf("source digests are not deterministic: first=%q second=%q findings=%+v", first, second, findings)
	}
	for _, source := range successSources {
		data, readFindings := readStrictFile(root, source.path)
		if len(readFindings) != 0 {
			t.Fatalf("read %s: %+v", source.path, readFindings)
		}
		want := source.id + ":" + fileSHA256(data)
		if !strings.Contains(first, want) {
			t.Fatalf("source digest %q is absent from %q", want, first)
		}
	}
	line := formatSuccess(
		traceResult{total: 163, allocations: map[string]int{"V1.0": 140, "V1.x": 18, "Later": 5}},
		12,
		policyResult{phase: "wave-01", files: 100, changed: 36},
		42,
		first,
	)
	for _, required := range []string{"gate_version=" + buildControlVersion, "source_sha256=" + first, "phase=wave-01", "requirements=163"} {
		if !strings.Contains(line, required) {
			t.Fatalf("success output %q does not contain %q", line, required)
		}
	}
}

func TestControllerExitNoRepairEnvironmentIsolationAndRepeatDeterminism(t *testing.T) {
	root := repositoryRoot(t)
	before, beforeFindings := scanRepository(root)
	if len(beforeFindings) != 0 {
		t.Fatalf("pre-run snapshot: %+v", beforeFindings)
	}
	temp := t.TempDir()
	binary := filepath.Join(temp, "build-control")
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	build := exec.Command(goBinary, "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", binary, "./internal/harness/buildcontrol")
	build.Dir = root
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "GOFLAGS=")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build controller: %v\n%s", err, output)
	}
	run := func(extraEnvironment ...string) (string, error) {
		command := exec.Command(binary)
		command.Dir = root
		command.Env = append(os.Environ(), extraEnvironment...)
		output, err := command.CombinedOutput()
		return string(output), err
	}
	first, err := run()
	if err != nil || !strings.HasPrefix(first, "PASS rule=BCTL-000 ") {
		t.Fatalf("controller success exit: err=%v output=%q", err, first)
	}
	second, err := run("L7_PHASE=wave-99", "L7_APPROVED=true", "L7_TEST_MODE=repair")
	if err != nil || second != first {
		t.Fatalf("environment affected fixed policy: err=%v first=%q second=%q", err, first, second)
	}
	argument := exec.Command(binary, "unexpected")
	argument.Dir = root
	output, err := argument.CombinedOutput()
	if err == nil {
		t.Fatalf("argument failure exited zero: %q", output)
	}
	exitError, ok := err.(*exec.ExitError)
	if !ok || exitError.ExitCode() == 0 || !strings.Contains(string(output), "rule=BCTL-001") {
		t.Fatalf("argument failure lacks stable nonzero result: err=%v output=%q", err, output)
	}
	after, afterFindings := scanRepository(root)
	if len(afterFindings) != 0 {
		t.Fatalf("post-run snapshot: %+v", afterFindings)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("controller repaired or mutated repository bytes")
	}
}
