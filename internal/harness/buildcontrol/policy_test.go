package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestCurrentPolicyContract(t *testing.T) {
	result, findings := checkPolicy(repositoryRoot(t))
	if len(findings) != 0 {
		t.Fatalf("policy findings: %+v", findings)
	}
	if result.phase != "wave-02" || result.checkpoint != "in-progress" || result.files == 0 || result.changed == 0 {
		t.Fatalf("unexpected policy result: %+v", result)
	}
}

func TestPhaseRowsRejectMissingDuplicateAndChangedBindings(t *testing.T) {
	t.Parallel()
	valid := make([]tsvRow, 0, len(expectedPhases))
	for _, phase := range expectedPhases {
		valid = append(valid, tsvRow{"phase": phase.phase, "state": phase.state, "base_commit": phase.baseCommit, "base_tree": phase.baseTree, "base_manifest": phase.baseManifest, "path_policy": phase.pathPolicy})
	}
	if _, findings := validatePhaseRows(nil); findingRules(findings)["SCOPE-300"] == 0 {
		t.Fatalf("missing phase findings: %+v", findings)
	}
	if _, findings := validatePhaseRows(append(append([]tsvRow(nil), valid...), valid[1])); findingRules(findings)["SCOPE-300"] == 0 {
		t.Fatalf("duplicate phase findings: %+v", findings)
	}
	changed := append([]tsvRow(nil), valid...)
	changed[1] = tsvRow{}
	for key, value := range valid[1] {
		changed[1][key] = value
	}
	changed[1]["base_tree"] = strings.Repeat("0", 40)
	if _, findings := validatePhaseRows(changed); findingRules(findings)["SCOPE-301"] == 0 {
		t.Fatalf("changed phase findings: %+v", findings)
	}
	unknown := append([]tsvRow(nil), valid...)
	unknown[1] = tsvRow{}
	for key, value := range valid[1] {
		unknown[1][key] = value
	}
	unknown[1]["phase"] = "wave-unknown"
	if _, findings := validatePhaseRows(unknown); findingRules(findings)["SCOPE-301"] == 0 {
		t.Fatalf("unknown phase findings: %+v", findings)
	}
}

func TestPathRowsRejectUnknownDuplicateAndChangedRules(t *testing.T) {
	t.Parallel()
	var rows []tsvRow
	for relative, expected := range expectedWave02Paths {
		rows = append(rows, tsvRow{"path": relative, "change": expected.change, "owner": expected.owner, "rule": expected.rule})
	}
	rows[0]["owner"] = "unknown"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"path": "unapproved/file", "change": "add", "owner": "wave-integrator", "rule": "SCOPE-321"})
	rules := findingRules(func() []finding { _, findings := validatePathRows(rows, expectedWave02Paths); return findings }())
	for _, rule := range []string{"SCOPE-312", "SCOPE-314", "SCOPE-315"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestSnapshotRejectsMissingProtectedAndUnauthorizedPaths(t *testing.T) {
	t.Parallel()
	base := map[string]string{"protected": "aaa", "mutable": "bbb"}
	current := map[string]snapshotFile{
		"protected": {digest: "changed", regular: true, links: 1},
		"extra":     {digest: "ccc", regular: true, links: 1},
	}
	rules := map[string]pathExpectation{"mutable": {change: "modify", owner: "owner", rule: "SCOPE-320"}}
	_, findings := validateSnapshot(base, current, rules)
	ruleCounts := findingRules(findings)
	for _, rule := range []string{"SCOPE-352", "SCOPE-353", "SCOPE-354"} {
		if ruleCounts[rule] == 0 {
			t.Errorf("findings %+v do not contain %s", findings, rule)
		}
	}
}

func TestManifestRejectsMalformedDuplicateAndUnsortedRows(t *testing.T) {
	t.Parallel()
	data := []byte("# fixture\n" + strings.Repeat("a", 64) + "  z\n" + strings.Repeat("b", 64) + "  a\n" + strings.Repeat("c", 64) + "  a\n" + "bad  x\n")
	_, findings := parseSHA256Manifest("fixture", data, true)
	rules := findingRules(findings)
	for _, rule := range []string{"SCOPE-330", "SCOPE-331", "SCOPE-332"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestFileShapeRejectsSymlinkSpecialAndHardlink(t *testing.T) {
	t.Parallel()
	for name, mode := range map[string]fs.FileMode{
		"symlink": fs.ModeSymlink,
		"pipe":    fs.ModeNamedPipe,
		"device":  fs.ModeDevice,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if rules := findingRules(validateFileShape("candidate", mode, 0, false)); rules["SCOPE-343"] == 0 {
				t.Fatalf("nonregular mode %v was accepted: %+v", mode, rules)
			}
		})
	}
	if rules := findingRules(validateFileShape("candidate", 0, 2, true)); rules["SCOPE-344"] == 0 {
		t.Fatalf("hardlinked regular file was accepted: %+v", rules)
	}
	if findings := validateFileShape("candidate", 0, 1, true); len(findings) != 0 {
		t.Fatalf("single-link regular file findings: %+v", findings)
	}
}

func TestModuleInvariantsRejectWrongDependencyReservedAndExtraStates(t *testing.T) {
	t.Parallel()
	validRows := []tsvRow{
		{"role": "core", "state": "active", "directory": ".", "module_path": selectedModule},
		{"role": "updater", "state": "reserved", "directory": "cmd/l7up", "module_path": "UNSET"},
	}
	validModule := "module " + selectedModule + "\n\ngo 1.26.0\n\ntoolchain go1.26.7\n"
	if findings := validateModuleInvariants(validModule, validRows, true); len(findings) != 0 {
		t.Fatalf("valid module findings: %+v", findings)
	}

	legacyModuleData := "module " + legacyModule + "\n\ngo 1.26.0\n\ntoolchain go1.26.7\n"
	legacyRows := []tsvRow{
		{"role": "core", "state": "active", "directory": ".", "module_path": legacyModule},
		validRows[1],
	}
	if rules := findingRules(validateModuleInvariants(legacyModuleData, legacyRows, true)); rules["SCOPE-373"] == 0 {
		t.Fatalf("final legacy module was accepted: %+v", rules)
	}

	withDependency := validModule + "\nrequire example.invalid/dependency v1.0.0\n"
	if rules := findingRules(validateModuleInvariants(withDependency, validRows, true)); rules["SCOPE-374"] == 0 {
		t.Fatalf("dependency-bearing module was accepted: %+v", rules)
	}

	badRows := []tsvRow{
		validRows[0],
		{"role": "updater", "state": "active", "directory": "cmd/l7up", "module_path": "example.invalid/updater"},
		{"role": "extra", "state": "active", "directory": "extra", "module_path": "example.invalid/extra"},
	}
	rules := findingRules(validateModuleInvariants(validModule, badRows, true))
	for _, rule := range []string{"SCOPE-376", "SCOPE-377"} {
		if rules[rule] == 0 {
			t.Errorf("rules %+v do not contain %s", rules, rule)
		}
	}
}

func TestSafeRelativeASCIIPathRejectsAliases(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "/absolute", "../escape", "a/../b", "a\\b", "sp ace", "unicodé"} {
		if safeRelativeASCIIPath(value) {
			t.Errorf("path %q unexpectedly accepted", value)
		}
	}
	if !safeRelativeASCIIPath("docs/artifacts/file.md") {
		t.Fatal("canonical path rejected")
	}
}

func TestRepositoryScanStopsAtDirectoryAndFileBounds(t *testing.T) {
	t.Parallel()
	t.Run("directories", func(t *testing.T) {
		root := t.TempDir()
		for index := 0; index <= maxRepositoryDirectories; index++ {
			if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("d-%03d", index)), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		_, findings := scanRepository(root)
		if findingRules(findings)["SCOPE-348"] == 0 {
			t.Fatalf("directory traversal overflow was accepted: %+v", findings)
		}
	})
	t.Run("files", func(t *testing.T) {
		root := t.TempDir()
		for index := 0; index <= maxRepositoryFiles; index++ {
			name := filepath.Join(root, fmt.Sprintf("f-%03d", index))
			if err := os.WriteFile(name, nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		_, findings := scanRepository(root)
		if findingRules(findings)["SCOPE-346"] == 0 {
			t.Fatalf("file traversal overflow was accepted: %+v", findings)
		}
	})
}

func TestRepositoryScanCapsSingleDirectoryReadBeforeEnumeration(t *testing.T) {
	t.Parallel()
	const expected = "BLOCKED rule=SCOPE-338 subject=. message=\"repository entry count exceeds the combined Wave 1 bound\" next=\"remove the unapproved paths or approve a bounded successor\"\n"
	for _, testCase := range []struct {
		name       string
		descending bool
	}{{"ascending-creation", false}, {"descending-creation", true}} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			for offset := 0; offset < maxRepositoryReadBatch; offset++ {
				index := offset
				if testCase.descending {
					index = maxRepositoryReadBatch - 1 - offset
				}
				name := filepath.Join(root, fmt.Sprintf("f-%04d", index))
				if err := os.WriteFile(name, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			maximumRequested := 0
			reader := func(rooted *os.Root, relative string, entryLimit int) ([]os.DirEntry, []finding) {
				if entryLimit > maximumRequested {
					maximumRequested = entryLimit
				}
				return readRepositoryDirectory(rooted, relative, entryLimit)
			}
			fixedNow := func() time.Time { return time.Unix(0, 0) }
			current, findings := scanRepositoryWithDependencies(root, fixedNow, reader)
			if actual := renderFindings(findings); actual != expected {
				t.Fatalf("oversized single-directory result differs:\nwant:\n%s\ngot:\n%s", expected, actual)
			}
			if maximumRequested != maxRepositoryReadBatch {
				t.Fatalf("directory read requested %d entries, want exact limit %d", maximumRequested, maxRepositoryReadBatch)
			}
			if len(current) != 0 {
				t.Fatalf("scan retained %d files before aggregate bound rejection", len(current))
			}
		})
	}
}

func TestRepositoryScanMixedEntryBatchIsOrderIndependent(t *testing.T) {
	t.Parallel()
	const expected = "BLOCKED rule=SCOPE-338 subject=. message=\"repository entry count exceeds the combined Wave 1 bound\" next=\"remove the unapproved paths or approve a bounded successor\"\n"
	testCases := []struct {
		name               string
		directories, files int
		reverse            bool
	}{
		{"directory-omitted", 512, 515, false},
		{"file-omitted", 513, 514, true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			batch := syntheticRepositoryBatch(testCase.directories, testCase.files, testCase.reverse)
			reader := func(_ *os.Root, relative string, entryLimit int) ([]os.DirEntry, []finding) {
				if relative != "." || entryLimit != maxRepositoryReadBatch {
					t.Fatalf("unexpected directory request: relative=%q limit=%d", relative, entryLimit)
				}
				return batch, nil
			}
			fixedNow := func() time.Time { return time.Unix(0, 0) }
			current, findings := scanRepositoryWithDependencies(t.TempDir(), fixedNow, reader)
			if actual := renderFindings(findings); actual != expected {
				t.Fatalf("mixed entry result differs:\nwant:\n%s\ngot:\n%s", expected, actual)
			}
			if len(current) != 0 {
				t.Fatalf("scan retained %d entries before aggregate bound rejection", len(current))
			}
		})
	}
}

func TestRepositoryEntryBatchFailureIsStableAcrossProcesses(t *testing.T) {
	const helperEnvironment = "L7_AUD_W01_016_HELPER"
	if variant := os.Getenv(helperEnvironment); variant != "" {
		directories, files, reverse := 512, 515, false
		if variant == "file-omitted" {
			directories, files, reverse = 513, 514, true
		}
		batch := syntheticRepositoryBatch(directories, files, reverse)
		reader := func(*os.Root, string, int) ([]os.DirEntry, []finding) { return batch, nil }
		fixedNow := func() time.Time { return time.Unix(0, 0) }
		_, findings := scanRepositoryWithDependencies(".", fixedNow, reader)
		fmt.Print(renderFindings(findings))
		os.Exit(1)
	}

	run := func(variant string) (string, error) {
		command := exec.Command(os.Args[0], "-test.run=^TestRepositoryEntryBatchFailureIsStableAcrossProcesses$")
		command.Env = processFixtureEnvironment(helperEnvironment + "=" + variant)
		output, err := command.CombinedOutput()
		return string(output), err
	}
	first, firstErr := run("directory-omitted")
	second, secondErr := run("file-omitted")
	firstExit, firstOK := firstErr.(*exec.ExitError)
	secondExit, secondOK := secondErr.(*exec.ExitError)
	if !firstOK || !secondOK || firstExit.ExitCode() != 1 || secondExit.ExitCode() != 1 {
		t.Fatalf("aggregate bound did not exit one: first=%v second=%v", firstErr, secondErr)
	}
	const expected = "BLOCKED rule=SCOPE-338 subject=. message=\"repository entry count exceeds the combined Wave 1 bound\" next=\"remove the unapproved paths or approve a bounded successor\"\n"
	if first != expected || second != expected {
		t.Fatalf("aggregate bound output differs across processes:\nwant:\n%s\nfirst:\n%s\nsecond:\n%s", expected, first, second)
	}
}

type syntheticDirEntry struct {
	name      string
	directory bool
}

func (entry syntheticDirEntry) Name() string { return entry.name }
func (entry syntheticDirEntry) IsDir() bool  { return entry.directory }
func (entry syntheticDirEntry) Type() fs.FileMode {
	if entry.directory {
		return fs.ModeDir
	}
	return 0
}
func (entry syntheticDirEntry) Info() (fs.FileInfo, error) {
	return nil, fmt.Errorf("entry %s was inspected before aggregate bound rejection", entry.name)
}

func syntheticRepositoryBatch(directories, files int, reverse bool) []os.DirEntry {
	entries := make([]os.DirEntry, 0, directories+files)
	appendEntries := func(prefix string, count int, directory bool) {
		for offset := 0; offset < count; offset++ {
			index := offset
			if reverse {
				index = count - 1 - offset
			}
			entries = append(entries, syntheticDirEntry{name: fmt.Sprintf("%s-%04d", prefix, index), directory: directory})
		}
	}
	if reverse {
		appendEntries("f", files, false)
		appendEntries("d", directories, true)
	} else {
		appendEntries("d", directories, true)
		appendEntries("f", files, false)
	}
	return entries
}

func TestRepositoryScanDeadlineFailsClosedBeforeFurtherIO(t *testing.T) {
	t.Parallel()
	base := time.Unix(0, 0)
	clockCalls := 0
	now := func() time.Time {
		clockCalls++
		if clockCalls == 1 {
			return base
		}
		return base.Add(maxRepositoryScanTime + time.Nanosecond)
	}
	readCalls := 0
	reader := func(*os.Root, string, int) ([]os.DirEntry, []finding) {
		readCalls++
		return nil, nil
	}
	_, findings := scanRepositoryWithDependencies(t.TempDir(), now, reader)
	if findingRules(findings)["SCOPE-339"] == 0 || readCalls != 0 {
		t.Fatalf("deadline did not fail closed before directory I/O: reads=%d findings=%+v", readCalls, findings)
	}
}

func TestRepositoryFileReadHonorsByteBoundary(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	name := filepath.Join(root, "fixture")
	if err := os.WriteFile(name, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(name)
	if err != nil {
		t.Fatal(err)
	}
	rooted, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer rooted.Close()
	if _, size, findings := readRepositoryFile(rooted, "fixture", info, 4); len(findings) != 0 || size != 4 {
		t.Fatalf("at-limit repository read failed: size=%d findings=%+v", size, findings)
	}
	if _, _, findings := readRepositoryFile(rooted, "fixture", info, 3); findingRules(findings)["SCOPE-347"] == 0 {
		t.Fatalf("over-limit repository read was accepted: %+v", findings)
	}
}

func TestCandidateManifestRejectsWrongAndExcludedRows(t *testing.T) {
	t.Parallel()
	manifestPath := wave02CandidateManifest
	root := materializeMapFS(t, fstest.MapFS{
		manifestPath: {Data: []byte(strings.Repeat("b", 64) + "  changed\n" + strings.Repeat("c", 64) + "  " + wave02AuditPath + "\n")},
	})
	base := map[string]string{}
	current := map[string]snapshotFile{
		"changed":    {digest: strings.Repeat("a", 64), regular: true, links: 1},
		manifestPath: {regular: true, links: 1},
	}
	rules := map[string]pathExpectation{
		"changed":    {change: "add", owner: "wave-integrator", rule: "SCOPE-321"},
		manifestPath: {change: "add", owner: "wave-integrator", rule: "SCOPE-321"},
	}
	findings := validateCandidateManifest(root, base, current, rules, candidateClosure{manifestPath: manifestPath, evidencePath: wave02EvidencePath, auditPath: wave02AuditPath, expectedRows: 1})
	for _, rule := range []string{"SCOPE-390", "SCOPE-391"} {
		if findingRules(findings)[rule] == 0 {
			t.Fatalf("candidate-manifest fixture lacks %s: %+v", rule, findings)
		}
	}
}

func TestUpdaterPathAndNonASCIIPolicyFailForIntendedRules(t *testing.T) {
	t.Parallel()
	root := materializeMapFS(t, fstest.MapFS{"cmd/l7up/main.go": {Data: []byte("package main\n")}})
	if findings := checkForbiddenProductPaths(root); findingRules(findings)["SCOPE-379"] == 0 {
		t.Fatalf("reserved updater path was accepted: %+v", findings)
	}
	rows := []tsvRow{{"change": "add", "path": "unicodé/file", "owner": "wave-integrator", "rule": "SCOPE-321"}}
	if _, findings := validatePathRows(rows, map[string]pathExpectation{}); findingRules(findings)["SCOPE-310"] == 0 {
		t.Fatalf("non-ASCII path was accepted: %+v", findings)
	}
}

func TestImportBoundaryBrokenPackageGraphs(t *testing.T) {
	script, err := os.ReadFile(filepath.Join(repositoryRoot(t), "scripts/harness/check-import-boundaries.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name       string
		policy     string
		files      map[string]string
		expected   string
		moduleTail string
	}{
		{
			name: "external-module-detour", policy: "effect\tinternal/kernel\tos\tBND-004\n", expected: "BND-006",
			moduleTail: "require example.test/external v0.0.0\nreplace example.test/external => ./thirdparty\n",
			files: map[string]string{
				"internal/kernel/kernel.go": "package kernel\nimport _ \"example.test/external/ext\"\n",
				"thirdparty/go.mod":         "module example.test/external\n\ngo 1.26.0\n",
				"thirdparty/ext/ext.go":     "package ext\n",
			},
		},
		{
			name: "harness-import", expected: "BND-005",
			files: map[string]string{
				"internal/harness/helper/helper.go": "package helper\n",
				"internal/product/product.go":       "package product\nimport _ \"example.test/fixture/internal/harness/helper\"\n",
			},
		},
		{
			name: "unsafe-import", expected: "BND-007",
			files: map[string]string{
				"internal/product/product.go": "package product\nimport \"unsafe\"\nvar _ unsafe.Pointer\n",
			},
		},
		{
			name: "forbidden-transitive-import", policy: "transitive\tinternal/executor\tinternal/render\tBND-001\n", expected: "BND-001",
			files: map[string]string{
				"internal/executor/executor.go": "package executor\nimport _ \"example.test/fixture/internal/render\"\n",
				"internal/render/render.go":     "package render\n",
			},
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			fixture := fstest.MapFS{
				"go.mod":                                     {Data: []byte("module example.test/fixture\n\ngo 1.26.0\n\n" + testCase.moduleTail)},
				"harness/modules.lock.tsv":                   {Data: []byte("# role\tstate\tdirectory\tmodule_path\ncore\tactive\t.\texample.test/fixture\nupdater\treserved\tcmd/l7up\tUNSET\n")},
				"harness/import-boundaries.tsv":              {Data: []byte("# mode\tsource_prefix\tforbidden_prefix\trule\n" + testCase.policy)},
				"scripts/harness/check-import-boundaries.sh": {Data: script},
			}
			for name, data := range testCase.files {
				fixture[name] = &fstest.MapFile{Data: []byte(data)}
			}
			root := materializeMapFS(t, fixture)
			cache := t.TempDir()
			for _, directory := range []string{"go-cache", "mod-cache", "go-path", "go-tmp", "telemetry"} {
				if err := os.Mkdir(filepath.Join(cache, directory), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
			command := exec.Command("/bin/sh", filepath.Join(root, "scripts/harness/check-import-boundaries.sh"), goBinary)
			command.Dir = root
			command.Env = processFixtureEnvironment(
				"GOENV=off", "GOTOOLCHAIN=local", "GOWORK=off", "GO111MODULE=on", "GOFLAGS=-mod=readonly", "CGO_ENABLED=0",
				"GOPROXY=off", "GOSUMDB=off", "GOCACHE="+filepath.Join(cache, "go-cache"), "GOMODCACHE="+filepath.Join(cache, "mod-cache"),
				"GOPATH="+filepath.Join(cache, "go-path"), "GOTMPDIR="+filepath.Join(cache, "go-tmp"), "TMPDIR="+filepath.Join(cache, "go-tmp"),
				"TEST_TELEMETRY_DIR="+filepath.Join(cache, "telemetry"), "GOVCS=*:off", "GOAUTH=off", "GIT_TERMINAL_PROMPT=0",
			)
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), testCase.expected) {
				t.Fatalf("broken graph did not fail for %s: err=%v output=%s", testCase.expected, err, output)
			}
		})
	}
}
