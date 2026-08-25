package main

import (
	"strings"
	"testing"
)

func TestCurrentPolicyContract(t *testing.T) {
	result, findings := checkPolicy(repositoryRoot(t))
	if len(findings) != 0 {
		t.Fatalf("policy findings: %+v", findings)
	}
	if result.phase != "wave-01" || result.files == 0 || result.changed == 0 {
		t.Fatalf("unexpected policy result: %+v", result)
	}
}

func TestPhaseRowsRejectMissingDuplicateAndChangedBindings(t *testing.T) {
	t.Parallel()
	valid := tsvRow{"phase": "wave-01", "state": "active", "base_commit": waveBaseCommit, "base_tree": waveBaseTree, "base_manifest": "harness/wave-01-base.sha256", "path_policy": "harness/wave-01-paths.tsv"}
	if _, findings := validatePhaseRows(nil); findingRules(findings)["SCOPE-300"] == 0 {
		t.Fatalf("missing phase findings: %+v", findings)
	}
	if _, findings := validatePhaseRows([]tsvRow{valid, valid}); findingRules(findings)["SCOPE-300"] == 0 {
		t.Fatalf("duplicate phase findings: %+v", findings)
	}
	changed := tsvRow{}
	for key, value := range valid {
		changed[key] = value
	}
	changed["base_tree"] = strings.Repeat("0", 40)
	if _, findings := validatePhaseRows([]tsvRow{changed}); findingRules(findings)["SCOPE-301"] == 0 {
		t.Fatalf("changed phase findings: %+v", findings)
	}
}

func TestPathRowsRejectUnknownDuplicateAndChangedRules(t *testing.T) {
	t.Parallel()
	var rows []tsvRow
	for relative, expected := range expectedWavePaths {
		rows = append(rows, tsvRow{"path": relative, "change": expected.change, "owner": expected.owner, "rule": expected.rule})
	}
	rows[0]["owner"] = "unknown"
	rows = append(rows, rows[1])
	rows = append(rows, tsvRow{"path": "unapproved/file", "change": "add", "owner": "wave-integrator", "rule": "SCOPE-321"})
	rules := findingRules(func() []finding { _, findings := validatePathRows(rows); return findings }())
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
