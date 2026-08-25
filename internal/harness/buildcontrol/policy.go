package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	waveBaseCommit           = "ee181b759c346055b0fb5b2fa1b3b1e676dd83e4"
	waveBaseTree             = "2f23a0810660995b6f562c361ab38cd4faafa3b3"
	selectedModule           = "github.com/addressanup/level7-dev-loop"
	legacyModule             = "continuallabs.ltd/level7-dev-loop"
	maxRepositoryDirectories = 512
	maxRepositoryFiles       = 512
	maxRepositoryBytes       = int64(8 << 20)
)

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var errRepositoryBound = errors.New("repository scan bound reached")

type pathExpectation struct {
	change string
	owner  string
	rule   string
}

type snapshotFile struct {
	digest  string
	regular bool
	links   uint64
}

type policyResult struct {
	phase   string
	files   int
	changed int
}

var expectedWavePaths = map[string]pathExpectation{
	".github/workflows/harness.yml":                      {"modify", "harness-integrator", "SCOPE-320"},
	"Makefile":                                           {"modify", "harness-integrator", "SCOPE-320"},
	"README.md":                                          {"modify", "wave-integrator", "SCOPE-320"},
	"docs/artifacts/wave-01-approval.md":                 {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-audit.md":                    {"audit-only", "independent-reviewer", "SCOPE-322"},
	"docs/artifacts/wave-01-audit-remediation.md":        {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-candidate.sha256":            {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-change-contract.md":          {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-design.md":                   {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-design-amendment.md":         {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-evidence.md":                 {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-grant-ladder-amendment.md":   {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-module-identity-decision.md": {"add", "wave-integrator", "SCOPE-321"},
	"docs/artifacts/wave-01-specification.md":            {"add", "wave-integrator", "SCOPE-321"},
	"go.mod":                                          {"conditional", "harness-integrator", "SCOPE-323"},
	"harness/control-ownership.tsv":                   {"add", "harness-integrator", "SCOPE-321"},
	"harness/phases.tsv":                              {"add", "harness-integrator", "SCOPE-321"},
	"harness/prototype-dispositions.tsv":              {"add", "harness-integrator", "SCOPE-321"},
	"harness/support-matrix.tsv":                      {"add", "harness-integrator", "SCOPE-321"},
	"harness/wave-01-base.sha256":                     {"add", "harness-integrator", "SCOPE-321"},
	"harness/wave-01-paths.tsv":                       {"add", "harness-integrator", "SCOPE-321"},
	"harness/modules.lock.tsv":                        {"conditional", "harness-integrator", "SCOPE-323"},
	"internal/harness/buildcontrol/claims.go":         {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/claims_test.go":    {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/load.go":           {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/main.go":           {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/markdown.go":       {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/nlink_unix.go":     {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/ownership.go":      {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/ownership_test.go": {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/policy.go":         {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/policy_test.go":    {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/report.go":         {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/testutil_test.go":  {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/trace.go":          {"add", "harness-integrator", "SCOPE-321"},
	"internal/harness/buildcontrol/trace_test.go":     {"add", "harness-integrator", "SCOPE-321"},
	"scripts/harness/check-import-boundaries.sh":      {"modify", "harness-integrator", "SCOPE-320"},
}

var approvedWaveInputs = map[string]string{
	"docs/artifacts/wave-01-change-contract.md":  "f53d06d2b02760bcf6ca958b72e4d2473cc52edc3f4a2cb1471cadbd4ab42afc",
	"docs/artifacts/wave-01-design.md":           "07953b2319635846505a018c3e4cc66705e0c263ab01b0a5c79e75cdaf1fb8e8",
	"docs/artifacts/wave-01-design-amendment.md": "934a9b0fb2839425401d77ef53d9a7914a14812f3eadd4fa10e4f770ebe12e29",
	"docs/artifacts/wave-01-specification.md":    "8715388fbe0185a3ae24d4c13d30704305a2393526fefcc71a82fce9bba119cc",
}

var forbiddenWaveProductPaths = []string{
	"cmd/l7", "cmd/l7up", "internal/supervisor", "internal/kernel", "internal/context", "internal/artifact", "internal/policy", "internal/transaction", "internal/executor", "internal/receipt", "internal/platform", "internal/adapter", "internal/channel", "internal/render", "internal/evaluator", "semantic", "schemas", "fixtures", "packages", "build/generated",
}

func checkPolicy(root string) (policyResult, []finding) {
	phaseRows, findings := loadTSV(root, "harness/phases.tsv", []string{"phase", "state", "base_commit", "base_tree", "base_manifest", "path_policy"})
	phase, phaseFindings := validatePhaseRows(phaseRows)
	findings = appendFindings(findings, phaseFindings...)
	pathRows, pathFindings := loadTSV(root, "harness/wave-01-paths.tsv", []string{"change", "path", "owner", "rule"})
	rules, validationFindings := validatePathRows(pathRows)
	findings = appendFindings(findings, pathFindings...)
	findings = appendFindings(findings, validationFindings...)
	baseData, baseReadFindings := readStrictFile(root, "harness/wave-01-base.sha256")
	findings = appendFindings(findings, baseReadFindings...)
	base, baseFindings := parseSHA256Manifest("harness/wave-01-base.sha256", baseData, true)
	findings = appendFindings(findings, baseFindings...)
	current, walkFindings := scanRepository(root)
	findings = appendFindings(findings, walkFindings...)
	changed, snapshotFindings := validateSnapshot(base, current, rules)
	findings = appendFindings(findings, snapshotFindings...)
	findings = appendFindings(findings, checkProtectedInputs(root)...)
	findings = appendFindings(findings, checkApprovedWaveInputs(root)...)
	finalCandidate := current["docs/artifacts/wave-01-candidate.sha256"].regular
	findings = appendFindings(findings, checkHarnessInvariants(root, finalCandidate)...)
	if finalCandidate {
		findings = appendFindings(findings, validateCandidateManifest(root, base, current, rules)...)
	}
	return policyResult{phase: phase, files: len(current), changed: changed}, findings
}

func validatePhaseRows(rows []tsvRow) (string, []finding) {
	if len(rows) != 1 {
		return "", []finding{newFinding("SCOPE-300", "harness/phases.tsv", fmt.Sprintf("phase registry has %d rows, want exactly one", len(rows)), "restore one active Wave 1 row")}
	}
	row := rows[0]
	if row["phase"] != "wave-01" || row["state"] != "active" || row["base_commit"] != waveBaseCommit || row["base_tree"] != waveBaseTree || row["base_manifest"] != "harness/wave-01-base.sha256" || row["path_policy"] != "harness/wave-01-paths.tsv" {
		return "", []finding{newFinding("SCOPE-301", "wave-01", "active phase binding differs from the approved base and policy", "restore the approved phase tuple")}
	}
	return row["phase"], nil
}

func validatePathRows(rows []tsvRow) (map[string]pathExpectation, []finding) {
	rules := make(map[string]pathExpectation)
	var findings []finding
	for _, row := range rows {
		relative := row["path"]
		if !safeRelativeASCIIPath(relative) {
			findings = appendFindings(findings, newFinding("SCOPE-310", relative, "path is not a canonical ASCII repository-relative path", "restore an exact approved path"))
			continue
		}
		if row["change"] != "add" && row["change"] != "modify" && row["change"] != "conditional" && row["change"] != "audit-only" {
			findings = appendFindings(findings, newFinding("SCOPE-311", relative, "unknown path change class", "use an approved change class"))
			continue
		}
		if _, duplicate := rules[relative]; duplicate {
			findings = appendFindings(findings, newFinding("SCOPE-312", relative, "duplicate path-policy row", "retain exactly one path rule"))
			continue
		}
		rules[relative] = pathExpectation{row["change"], row["owner"], row["rule"]}
	}
	for relative, expected := range expectedWavePaths {
		actual, ok := rules[relative]
		if !ok {
			findings = appendFindings(findings, newFinding("SCOPE-313", relative, "approved path is missing from the path policy", "restore the approved path rule"))
		} else if actual != expected {
			findings = appendFindings(findings, newFinding("SCOPE-314", relative, "path rule differs from the approved design", "restore the approved change, owner, and rule"))
		}
	}
	for relative := range rules {
		if _, ok := expectedWavePaths[relative]; !ok {
			findings = appendFindings(findings, newFinding("SCOPE-315", relative, "path policy contains an unapproved path", "remove it or obtain a new exact design approval"))
		}
	}
	return rules, findings
}

func parseSHA256Manifest(name string, data []byte, requireSorted bool) (map[string]string, []finding) {
	manifest := make(map[string]string)
	var findings []finding
	previous := ""
	for lineIndex, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || !sha256Pattern.MatchString(parts[0]) || !safeRelativeASCIIPath(parts[1]) {
			findings = appendFindings(findings, newFinding("SCOPE-330", fmt.Sprintf("%s:%d", name, lineIndex+1), "malformed SHA-256 manifest row", "restore lowercase SHA-256 and an exact relative path"))
			continue
		}
		if requireSorted && previous != "" && parts[1] <= previous {
			findings = appendFindings(findings, newFinding("SCOPE-331", parts[1], "candidate manifest paths are not bytewise sorted", "sort candidate paths bytewise"))
		}
		previous = parts[1]
		if _, duplicate := manifest[parts[1]]; duplicate {
			findings = appendFindings(findings, newFinding("SCOPE-332", parts[1], "duplicate manifest path", "retain one digest per path"))
			continue
		}
		manifest[parts[1]] = parts[0]
	}
	if len(manifest) == 0 {
		findings = appendFindings(findings, newFinding("SCOPE-333", name, "manifest contains no records", "restore the approved manifest"))
	}
	return manifest, findings
}

func scanRepository(root string) (map[string]snapshotFile, []finding) {
	current := make(map[string]snapshotFile)
	var findings []finding
	directoriesSeen := 0
	filesSeen := 0
	totalBytes := int64(0)
	err := filepath.WalkDir(root, func(fullPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			findings = appendFindings(findings, newFinding("SCOPE-340", filepath.ToSlash(fullPath), walkErr.Error(), "restore readable bounded repository state"))
			return errRepositoryBound
		}
		if fullPath == root {
			return nil
		}
		relativeOS, err := filepath.Rel(root, fullPath)
		if err != nil {
			findings = appendFindings(findings, newFinding("SCOPE-340", filepath.ToSlash(fullPath), err.Error(), "restore canonical repository paths"))
			return errRepositoryBound
		}
		relative := filepath.ToSlash(relativeOS)
		if entry.IsDir() && (relative == ".git" || relative == ".cache") {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			directoriesSeen++
			if directoriesSeen > maxRepositoryDirectories {
				findings = appendFindings(findings, newFinding("SCOPE-348", relative, "repository directory count exceeds the Wave 1 bound", "remove the unapproved paths or approve a bounded successor"))
				return errRepositoryBound
			}
			return nil
		}
		filesSeen++
		if filesSeen > maxRepositoryFiles {
			findings = appendFindings(findings, newFinding("SCOPE-346", relative, "repository file count exceeds the Wave 1 bound", "remove the unapproved paths or approve a bounded successor"))
			return errRepositoryBound
		}
		if !safeRelativeASCIIPath(relative) {
			findings = appendFindings(findings, newFinding("SCOPE-341", relative, "noncanonical or non-ASCII repository path", "remove or rename the unapproved path"))
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			findings = appendFindings(findings, newFinding("SCOPE-342", relative, err.Error(), "restore readable regular-file state"))
			return errRepositoryBound
		}
		links := uint64(0)
		linkCountKnown := false
		if info.Mode().IsRegular() {
			links, linkCountKnown = regularFileLinkCount(info)
		}
		shapeFindings := validateFileShape(relative, info.Mode(), links, linkCountKnown)
		findings = appendFindings(findings, shapeFindings...)
		if !info.Mode().IsRegular() {
			current[relative] = snapshotFile{}
			return nil
		}
		if info.Size() > maxRepositoryBytes-totalBytes {
			findings = appendFindings(findings, newFinding("SCOPE-347", relative, "repository bytes exceed the Wave 1 bound", "remove the unapproved data or approve a bounded successor"))
			return errRepositoryBound
		}
		file, bytesRead, readFindings := readRepositoryFile(fullPath, relative, info, maxRepositoryBytes-totalBytes)
		findings = appendFindings(findings, readFindings...)
		if len(readFindings) != 0 {
			return errRepositoryBound
		}
		totalBytes += bytesRead
		current[relative] = file
		return nil
	})
	if err != nil && !errors.Is(err, errRepositoryBound) {
		findings = appendFindings(findings, newFinding("SCOPE-340", root, err.Error(), "restore a readable repository tree"))
	}
	return current, findings
}

func readRepositoryFile(fullPath, relative string, walkInfo fs.FileInfo, byteLimit int64) (snapshotFile, int64, []finding) {
	file, err := os.Open(fullPath)
	if err != nil {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-345", relative, err.Error(), "restore readable candidate bytes")}
	}
	openedInfo, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-345", relative, statErr.Error(), "restore readable candidate bytes")}
	}
	if !os.SameFile(walkInfo, openedInfo) || !openedInfo.Mode().IsRegular() {
		_ = file.Close()
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-349", relative, "repository file changed during inspection", "retry from a stable canonical worktree")}
	}
	data, readErr := io.ReadAll(io.LimitReader(file, byteLimit+1))
	closedInfo, finalStatErr := file.Stat()
	closeErr := file.Close()
	if readErr != nil {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-345", relative, readErr.Error(), "restore readable candidate bytes")}
	}
	if finalStatErr != nil {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-345", relative, finalStatErr.Error(), "restore readable candidate bytes")}
	}
	if closeErr != nil {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-345", relative, closeErr.Error(), "restore readable candidate bytes")}
	}
	if int64(len(data)) > byteLimit {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-347", relative, "repository bytes exceed the Wave 1 bound", "remove the unapproved data or approve a bounded successor")}
	}
	if !os.SameFile(openedInfo, closedInfo) || openedInfo.Size() != closedInfo.Size() || closedInfo.Size() != int64(len(data)) {
		return snapshotFile{}, 0, []finding{newFinding("SCOPE-349", relative, "repository file changed during inspection", "retry from a stable canonical worktree")}
	}
	links, linkCountKnown := regularFileLinkCount(closedInfo)
	if shapeFindings := validateFileShape(relative, closedInfo.Mode(), links, linkCountKnown); len(shapeFindings) != 0 {
		return snapshotFile{}, 0, shapeFindings
	}
	return snapshotFile{digest: fileSHA256(data), regular: true, links: links}, int64(len(data)), nil
}

func validateFileShape(relative string, mode fs.FileMode, links uint64, linkCountKnown bool) []finding {
	if !mode.IsRegular() {
		return []finding{newFinding("SCOPE-343", relative, "repository entry is not a regular file", "remove the symlink or special node")}
	}
	if !linkCountKnown || links != 1 {
		return []finding{newFinding("SCOPE-344", relative, fmt.Sprintf("regular file link count is %d", links), "use one independent regular file")}
	}
	return nil
}

func validateSnapshot(base map[string]string, current map[string]snapshotFile, rules map[string]pathExpectation) (int, []finding) {
	var findings []finding
	changed := 0
	for relative, rule := range rules {
		_, inBase := base[relative]
		if (rule.change == "add" || rule.change == "audit-only") && inBase {
			findings = appendFindings(findings, newFinding("SCOPE-350", relative, "add-only path already exists in the approved base", "restore the approved base/path class"))
		}
		if (rule.change == "modify" || rule.change == "conditional") && !inBase {
			findings = appendFindings(findings, newFinding("SCOPE-351", relative, "modify path is absent from the approved base", "restore the approved base/path class"))
		}
	}
	for relative, digest := range base {
		file, ok := current[relative]
		if !ok {
			findings = appendFindings(findings, newFinding("SCOPE-352", relative, "approved base path is missing", "restore the exact base path"))
			continue
		}
		rule, mutable := rules[relative]
		if file.digest != digest {
			changed++
			if !mutable || (rule.change != "modify" && rule.change != "conditional") {
				findings = appendFindings(findings, newFinding("SCOPE-353", relative, "protected base bytes changed outside an approved modify rule", "restore the exact base bytes"))
			}
		}
	}
	for relative := range current {
		if _, inBase := base[relative]; inBase {
			continue
		}
		changed++
		rule, ok := rules[relative]
		if !ok || (rule.change != "add" && rule.change != "audit-only") {
			findings = appendFindings(findings, newFinding("SCOPE-354", relative, "repository contains an unapproved added path", "remove it or obtain a new exact path approval"))
		}
	}
	return changed, findings
}

func checkProtectedInputs(root string) []finding {
	data, findings := readStrictFile(root, "harness/foundation-inputs.sha256")
	if len(findings) != 0 {
		return findings
	}
	manifest, parseFindings := parseSHA256Manifest("harness/foundation-inputs.sha256", data, false)
	findings = appendFindings(findings, parseFindings...)
	for relative, expected := range manifest {
		content, readFindings := readStrictFile(root, relative)
		findings = appendFindings(findings, readFindings...)
		if len(readFindings) == 0 && fileSHA256(content) != expected {
			findings = appendFindings(findings, newFinding("SCOPE-360", relative, "protected foundation/prototype bytes changed", "restore the approved protected input"))
		}
	}
	return findings
}

func checkApprovedWaveInputs(root string) []finding {
	var findings []finding
	for relative, expected := range approvedWaveInputs {
		content, readFindings := readStrictFile(root, relative)
		findings = appendFindings(findings, readFindings...)
		if len(readFindings) == 0 && fileSHA256(content) != expected {
			findings = appendFindings(findings, newFinding("SCOPE-361", relative, "approved Wave 1 planning bytes changed", "restore the exact owner-approved input"))
		}
	}
	return findings
}

func checkHarnessInvariants(root string, finalCandidate bool) []finding {
	var findings []finding
	read := func(relative string) string {
		data, readFindings := readStrictFile(root, relative)
		findings = appendFindings(findings, readFindings...)
		return string(data)
	}
	version := read(".go-version")
	if version != "1.26.7\n" {
		findings = appendFindings(findings, newFinding("SCOPE-370", ".go-version", "baseline Go version changed", "restore Go 1.26.7"))
	}
	moduleData := read("go.mod")
	moduleRows, moduleFindings := loadTSV(root, "harness/modules.lock.tsv", []string{"role", "state", "directory", "module_path"})
	findings = appendFindings(findings, moduleFindings...)
	findings = appendFindings(findings, validateModuleInvariants(moduleData, moduleRows, finalCandidate)...)
	for _, relative := range []string{"go.sum", "vendor"} {
		if _, err := os.Lstat(filepath.Join(root, relative)); !os.IsNotExist(err) {
			findings = appendFindings(findings, newFinding("SCOPE-378", relative, "dependency artifact is forbidden in Wave 1", "remove it through an authorized recovery action"))
		}
	}
	findings = appendFindings(findings, checkForbiddenProductPaths(root)...)
	workflow := read(".github/workflows/harness.yml")
	for _, required := range []string{"permissions:", "contents: read", "persist-credentials: false", "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1", "go-version: 1.26.7", "go-version: 1.27.0", "experimental: false", "experimental: true", "Verify Wave 1 build controls offline", "run: make ci GO_VERSION=${{ matrix.go-version }}"} {
		if !strings.Contains(workflow, required) {
			findings = appendFindings(findings, newFinding("SCOPE-380", ".github/workflows/harness.yml", "configured CI lost a required safety or matrix binding", "restore the approved configured control"))
		}
	}
	if strings.Contains(workflow, "pull_request_target:") || strings.Contains(workflow, "secrets.") {
		findings = appendFindings(findings, newFinding("SCOPE-381", ".github/workflows/harness.yml", "configured CI contains a forbidden trigger or secret reference", "remove the forbidden control"))
	}
	makefile := read("Makefile")
	for _, required := range []string{"override CORE_MODULE_PATH :=", "$(PROJECT_ROOT)/harness/modules.lock.tsv", "override HARNESS_IMPORT_PATH := $(CORE_MODULE_PATH)/internal/harness", "build-control-check: toolchain-check", "\"$(GO)\" run -mod=readonly ./internal/harness/buildcontrol", "policy-check: build-control-check", "candidate-check: policy-check import-check"} {
		if !strings.Contains(makefile, required) {
			findings = appendFindings(findings, newFinding("SCOPE-383", "Makefile", "active build-control integration is incomplete", "restore the approved module-derived Wave 1 targets"))
		}
	}
	if strings.Contains(makefile, legacyModule) || strings.Contains(makefile, "check-foundation-scope.sh") {
		findings = appendFindings(findings, newFinding("SCOPE-384", "Makefile", "active harness retains a provisional module or predecessor policy entry", "use the module registry and Wave 1 controller"))
	}
	importCheck := read("scripts/harness/check-import-boundaries.sh")
	for _, required := range []string{"matches_prefix \"$package\" \"$harness_path\" && continue", "matches_prefix \"$imported\" \"$harness_path\""} {
		if !strings.Contains(importCheck, required) {
			findings = appendFindings(findings, newFinding("SCOPE-385", "scripts/harness/check-import-boundaries.sh", "harness descendant import denial is incomplete", "restore prefix denial for the entire harness subtree"))
		}
	}
	for relative, required := range map[string]string{
		"harness/signing-identities.lock.tsv": "EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796",
		"harness/ci-actions.lock.tsv":         "3d3c42e5aac5ba805825da76410c181273ba90b1",
		"harness/toolchains.lock.tsv":         "1.26.7",
	} {
		if !strings.Contains(read(relative), required) {
			findings = appendFindings(findings, newFinding("SCOPE-382", relative, "required frozen identity is missing", "restore the approved lock"))
		}
	}
	return findings
}

func checkForbiddenProductPaths(root string) []finding {
	var findings []finding
	for _, relative := range forbiddenWaveProductPaths {
		if _, err := os.Lstat(filepath.Join(root, relative)); !os.IsNotExist(err) {
			findings = appendFindings(findings, newFinding("SCOPE-379", relative, "product path exists during Wave 1 build control", "remove it through an authorized recovery action"))
		}
	}
	return findings
}

func validateModuleInvariants(moduleData string, moduleRows []tsvRow, finalCandidate bool) []finding {
	var findings []finding
	modulePath := ""
	for _, line := range strings.Split(moduleData, "\n") {
		if strings.HasPrefix(line, "module ") {
			modulePath = strings.TrimPrefix(line, "module ")
		}
	}
	if modulePath != legacyModule && modulePath != selectedModule {
		findings = appendFindings(findings, newFinding("SCOPE-372", "go.mod", "root module is neither the predecessor nor approved replacement", "use the approved GitHub module"))
	}
	if finalCandidate && modulePath != selectedModule {
		findings = appendFindings(findings, newFinding("SCOPE-373", "go.mod", "final candidate did not apply the approved module decision", "use the approved GitHub module"))
	}
	expectedGoMod := fmt.Sprintf("module %s\n\ngo 1.26.0\n\ntoolchain go1.26.7\n", modulePath)
	if moduleData != expectedGoMod {
		findings = appendFindings(findings, newFinding("SCOPE-374", "go.mod", "module file differs from the zero-dependency pinned Wave 1 shape", "restore the approved module, Go language, and toolchain lines"))
	}
	coreCount := 0
	updaterCount := 0
	for _, row := range moduleRows {
		if row["role"] == "core" {
			coreCount++
			if row["state"] != "active" || row["directory"] != "." || row["module_path"] != modulePath {
				findings = appendFindings(findings, newFinding("SCOPE-375", "core", "module registry does not match go.mod", "update both module identities together"))
			}
		}
		if row["role"] == "updater" {
			updaterCount++
			if row["state"] != "reserved" || row["directory"] != "cmd/l7up" || row["module_path"] != "UNSET" {
				findings = appendFindings(findings, newFinding("SCOPE-376", "updater", "updater must remain reserved with identity UNSET", "restore the Wave 10 reservation"))
			}
		}
		if row["role"] != "core" && row["role"] != "updater" {
			findings = appendFindings(findings, newFinding("SCOPE-377", row["role"], "module registry contains an unknown role", "retain only the active core and reserved updater"))
		}
	}
	if coreCount != 1 || updaterCount != 1 {
		findings = appendFindings(findings, newFinding("SCOPE-377", "harness/modules.lock.tsv", "module registry must contain one core and one updater row", "restore the exact module roles"))
	}
	expectedModules := fmt.Sprintf("# role\tstate\tdirectory\tmodule_path\ncore\tactive\t.\t%s\nupdater\treserved\tcmd/l7up\tUNSET\n", modulePath)
	var actualModules strings.Builder
	actualModules.WriteString("# role\tstate\tdirectory\tmodule_path\n")
	for _, row := range moduleRows {
		fmt.Fprintf(&actualModules, "%s\t%s\t%s\t%s\n", row["role"], row["state"], row["directory"], row["module_path"])
	}
	if actualModules.String() != expectedModules {
		findings = appendFindings(findings, newFinding("SCOPE-377", "harness/modules.lock.tsv", "module registry differs from the exact two-role Wave 1 shape", "restore the active core and reserved updater rows"))
	}
	return findings
}

func validateCandidateManifest(root string, base map[string]string, current map[string]snapshotFile, rules map[string]pathExpectation) []finding {
	manifestPath := "docs/artifacts/wave-01-candidate.sha256"
	data, findings := readStrictFile(root, manifestPath)
	if len(findings) != 0 {
		return findings
	}
	manifest, parseFindings := parseSHA256Manifest(manifestPath, data, true)
	findings = appendFindings(findings, parseFindings...)
	excluded := map[string]bool{
		manifestPath:                         true,
		"docs/artifacts/wave-01-evidence.md": true,
		"docs/artifacts/wave-01-audit.md":    true,
	}
	expected := make(map[string]string)
	for relative, file := range current {
		if excluded[relative] {
			continue
		}
		baseDigest, inBase := base[relative]
		if !inBase || file.digest != baseDigest {
			expected[relative] = file.digest
		}
	}
	for relative, digest := range expected {
		if manifest[relative] != digest {
			findings = appendFindings(findings, newFinding("SCOPE-390", relative, "candidate manifest is missing or has the wrong digest", "regenerate the exact candidate manifest"))
		}
	}
	for relative := range manifest {
		if _, ok := expected[relative]; !ok {
			findings = appendFindings(findings, newFinding("SCOPE-391", relative, "candidate manifest contains a noncandidate or excluded path", "regenerate the exact candidate manifest"))
		}
	}
	for relative, rule := range rules {
		if rule.change == "add" && relative != manifestPath && relative != "docs/artifacts/wave-01-evidence.md" {
			if _, ok := current[relative]; !ok {
				findings = appendFindings(findings, newFinding("SCOPE-392", relative, "final candidate is missing an approved required addition", "complete the approved candidate path"))
			}
		}
	}
	return findings
}

func safeRelativeASCIIPath(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || path.Clean(value) != value || value == "." {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
