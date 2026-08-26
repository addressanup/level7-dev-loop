package main

import (
	"fmt"
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

func TestTemporaryRootsAreRepositoryScoped(t *testing.T) {
	t.Parallel()
	want := filepath.Join(repositoryRoot(t), ".cache", "go", "tmp")
	temporaryRoot := os.Getenv("TMPDIR")
	goTemporaryRoot := os.Getenv("GOTMPDIR")
	if temporaryRoot != want || goTemporaryRoot != want || !filepath.IsAbs(temporaryRoot) {
		t.Fatalf("temporary roots are not bound to repository GOTMPDIR: TMPDIR=%q GOTMPDIR=%q want=%q", temporaryRoot, goTemporaryRoot, want)
	}
	physicalWant := physicalDirectory(t, want)
	if physicalWant != want {
		t.Fatalf("temporary root %q resolves outside its lexical repository path: %q", want, physicalWant)
	}
	testTemporaryRoot := physicalDirectory(t, t.TempDir())
	relative, err := filepath.Rel(physicalWant, testTemporaryRoot)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		t.Fatalf("testing temporary root %q is outside physical root %q: relative=%q err=%v", testTemporaryRoot, physicalWant, relative, err)
	}
}

func physicalDirectory(t *testing.T, directory string) string {
	t.Helper()
	info, err := os.Lstat(directory)
	if err != nil {
		t.Fatalf("lstat physical directory %q: %v", directory, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("physical directory %q has mode %v", directory, info.Mode())
	}
	physical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatalf("resolve physical directory %q: %v", directory, err)
	}
	physical, err = filepath.Abs(physical)
	if err != nil {
		t.Fatalf("make physical directory absolute %q: %v", physical, err)
	}
	return filepath.Clean(physical)
}

func runPrepareCacheFixture(t *testing.T, root string) (string, error) {
	t.Helper()
	script := filepath.Join(repositoryRoot(t), "scripts", "harness", "prepare-cache.sh")
	command := exec.Command("/bin/sh", script, root, "1.26.7")
	command.Env = []string{"PATH=/usr/bin:/bin", "LC_ALL=C", "TZ=UTC"}
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestPrepareCacheCreatesPhysicalRepositoryDirectories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if output, err := runPrepareCacheFixture(t, root); err != nil {
		t.Fatalf("prepare cache fixture: %v\n%s", err, output)
	}
	for _, relative := range []string{
		".cache",
		".cache/go",
		".cache/go/path",
		".cache/go/bin",
		".cache/go/build",
		".cache/go/mod",
		".cache/go/tmp",
		".cache/go/telemetry",
		".cache/repro",
	} {
		directory := filepath.Join(root, filepath.FromSlash(relative))
		if physical := physicalDirectory(t, directory); physical != directory {
			t.Fatalf("prepared cache directory %q resolves to %q", directory, physical)
		}
	}
	mode, err := os.ReadFile(filepath.Join(root, ".cache", "go", "telemetry", "mode"))
	if err != nil {
		t.Fatal(err)
	}
	if string(mode) != "off 2026-08-24\n" {
		t.Fatalf("unexpected telemetry mode: %q", mode)
	}
}

func TestPrepareCacheRejectsRedirectedComponentsBeforeWriting(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		setup func(*testing.T, string, string)
	}{
		{
			name: "cache-root",
			setup: func(t *testing.T, root, outside string) {
				if err := os.Symlink(outside, filepath.Join(root, ".cache")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "temporary-root",
			setup: func(t *testing.T, root, outside string) {
				if err := os.MkdirAll(filepath.Join(root, ".cache", "go"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(root, ".cache", "go", "tmp")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "telemetry-mode",
			setup: func(t *testing.T, root, outside string) {
				telemetry := filepath.Join(root, ".cache", "go", "telemetry")
				if err := os.MkdirAll(telemetry, 0o700); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(outside, "mode-target")
				if err := os.WriteFile(target, []byte("sentinel\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, filepath.Join(telemetry, "mode")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "toolchain-root",
			setup: func(t *testing.T, root, outside string) {
				if err := os.Mkdir(filepath.Join(root, ".cache"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(root, ".cache", "toolchains")); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			outside := t.TempDir()
			testCase.setup(t, root, outside)
			output, err := runPrepareCacheFixture(t, root)
			if err == nil || !strings.Contains(output, "prepare-cache: refusing symlinked") {
				t.Fatalf("redirected cache component was accepted: err=%v output=%q", err, output)
			}
			entries, readErr := os.ReadDir(outside)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if testCase.name == "telemetry-mode" {
				data, readErr := os.ReadFile(filepath.Join(outside, "mode-target"))
				if readErr != nil || string(data) != "sentinel\n" {
					t.Fatalf("redirected telemetry target changed: data=%q err=%v", data, readErr)
				}
				if len(entries) != 1 {
					t.Fatalf("redirected telemetry directory changed: %v", entries)
				}
			} else if len(entries) != 0 {
				t.Fatalf("redirected cache directory received writes: %v", entries)
			}
			if _, statErr := os.Lstat(filepath.Join(root, ".cache", "repro")); !os.IsNotExist(statErr) {
				t.Fatalf("prepare wrote before rejecting redirect: %v", statErr)
			}
		})
	}
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

func copyRepositoryFixture(t *testing.T) string {
	t.Helper()
	source := repositoryRoot(t)
	target := t.TempDir()
	err := filepath.WalkDir(source, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, name)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if entry.IsDir() && (relative == ".git" || relative == ".cache") {
			return filepath.SkipDir
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.Mkdir(destination, 0o700)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, data, info.Mode().Perm())
	})
	if err != nil {
		t.Fatalf("copy repository fixture: %v", err)
	}
	return target
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

func TestCappedFailingTraceIsRepeatDeterministicAcrossProcesses(t *testing.T) {
	const helperEnvironment = "L7_AUD_W01_008_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		definitions := make(map[string]struct{}, 163)
		for index := 1; index <= 163; index++ {
			definitions[fmt.Sprintf("L7-TEST-%03d", index)] = struct{}{}
		}
		_, findings := validateTrace(definitions, nil)
		fmt.Print(renderFindings(findings))
		os.Exit(1)
	}

	run := func() (string, error) {
		command := exec.Command(os.Args[0], "-test.run=^TestCappedFailingTraceIsRepeatDeterministicAcrossProcesses$")
		command.Env = append(os.Environ(), helperEnvironment+"=1")
		output, err := command.CombinedOutput()
		return string(output), err
	}

	first, firstErr := run()
	second, secondErr := run()
	if firstErr == nil || secondErr == nil {
		t.Fatalf("over-cap trace unexpectedly exited zero: first=%v second=%v", firstErr, secondErr)
	}
	if first != second {
		t.Fatalf("over-cap trace output differs across processes:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if !strings.Contains(first, "rule=BCTL-099") || !strings.Contains(first, "subject=L7-TEST-050") || strings.Contains(first, "subject=L7-TEST-051") {
		t.Fatalf("over-cap trace retained the wrong deterministic subset: %s", first)
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
