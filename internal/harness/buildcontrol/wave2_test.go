package main

import (
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWave02ExactPolicyRosterAndClosures(t *testing.T) {
	t.Parallel()
	if len(expectedWave02Paths) != 72 {
		t.Fatalf("Wave 2 path count: got %d, want 72", len(expectedWave02Paths))
	}
	counts := map[string]int{}
	for _, rule := range expectedWave02Paths {
		counts[rule.change]++
	}
	if counts["modify"] != 10 || counts["add"] != 61 || counts["audit-only"] != 1 {
		t.Fatalf("Wave 2 change classes: %+v", counts)
	}

	rows, findings := loadTSV(repositoryRoot(t), "harness/wave-02-paths.tsv", []string{"change", "path", "owner", "rule"})
	if len(findings) != 0 {
		t.Fatalf("load Wave 2 policy: %+v", findings)
	}
	rules, findings := validatePathRows(rows, expectedWave02Paths)
	if len(findings) != 0 || len(rules) != 72 {
		t.Fatalf("validate Wave 2 policy: rows=%d findings=%+v", len(rules), findings)
	}

	finalCandidatePaths := 0
	manifestRows := 0
	evidenceChildPaths := 0
	auditChildPaths := 0
	for relative, rule := range rules {
		if rule.change == "audit-only" {
			auditChildPaths++
			continue
		}
		if relative == wave02EvidencePath {
			evidenceChildPaths++
			continue
		}
		finalCandidatePaths++
		if relative != wave02CandidateManifest {
			manifestRows++
		}
	}
	evidenceChildPaths += finalCandidatePaths
	auditChildPaths += evidenceChildPaths
	if finalCandidatePaths != 70 || manifestRows != 69 || evidenceChildPaths != 71 || auditChildPaths != 72 {
		t.Fatalf("Wave 2 closures: candidate=%d manifest=%d evidence=%d audit=%d", finalCandidatePaths, manifestRows, evidenceChildPaths, auditChildPaths)
	}
}

func TestWave02BaseManifestBindsExactPredecessor(t *testing.T) {
	t.Parallel()
	data, findings := readStrictFile(repositoryRoot(t), "harness/wave-02-base.sha256")
	if len(findings) != 0 {
		t.Fatalf("read Wave 2 base: %+v", findings)
	}
	if digest := fileSHA256(data); digest != wave02BaseManifestSHA256 {
		t.Fatalf("Wave 2 base digest: got %s, want %s", digest, wave02BaseManifestSHA256)
	}
	manifest, findings := parseSHA256Manifest("harness/wave-02-base.sha256", data, true)
	if len(findings) != 0 || len(manifest) != 104 {
		t.Fatalf("Wave 2 base inventory: rows=%d findings=%+v", len(manifest), findings)
	}
	for _, required := range []string{"docs/artifacts/wave-01-audit.md", "docs/artifacts/wave-01-evidence.md", "harness/wave-01-paths.tsv"} {
		if manifest[required] == "" {
			t.Errorf("Wave 2 base lacks predecessor path %s", required)
		}
	}
}

func TestWave02ApprovalIsExactAndNonReplayable(t *testing.T) {
	t.Parallel()
	if findings := checkWave02Approval(repositoryRoot(t)); len(findings) != 0 {
		t.Fatalf("Wave 2 approval findings: %+v", findings)
	}
}

func TestWave02EvaluatorControlsLandAtomically(t *testing.T) {
	t.Parallel()
	partial := map[string]snapshotFile{wave02EvaluatorControlPaths[0]: {digest: strings.Repeat("a", 64), regular: true, links: 1}}
	if rules := findingRules(checkWave02EvaluatorFreeze(t.TempDir(), partial)); rules["SCOPE-560"] == 0 {
		t.Fatalf("partial evaluator controls were accepted: %+v", rules)
	}

	fixture := fstest.MapFS{}
	current := make(map[string]snapshotFile, len(wave02EvaluatorControlPaths)+1)
	paths := append([]string(nil), wave02EvaluatorControlPaths...)
	sort.Strings(paths)
	var manifest strings.Builder
	for _, relative := range paths {
		data := []byte(relative + "\n")
		digest := fileSHA256(data)
		fixture[relative] = &fstest.MapFile{Data: data}
		current[relative] = snapshotFile{digest: digest, regular: true, links: 1}
		manifest.WriteString(digest + "  " + relative + "\n")
	}
	fixture[wave02EvaluatorManifest] = &fstest.MapFile{Data: []byte(manifest.String())}
	current[wave02EvaluatorManifest] = snapshotFile{digest: fileSHA256([]byte(manifest.String())), regular: true, links: 1}
	root := materializeMapFS(t, fixture)
	if findings := checkWave02EvaluatorFreeze(root, current); len(findings) != 0 {
		t.Fatalf("complete evaluator freeze findings: %+v", findings)
	}

	tampered := make(map[string]snapshotFile, len(current))
	for relative, file := range current {
		tampered[relative] = file
	}
	file := tampered[paths[0]]
	file.digest = strings.Repeat("f", 64)
	tampered[paths[0]] = file
	if rules := findingRules(checkWave02EvaluatorFreeze(root, tampered)); rules["SCOPE-562"] == 0 {
		t.Fatalf("changed evaluator control was accepted: %+v", rules)
	}
}

func TestWave02UnknownPathAndProductPrefixRemainDenied(t *testing.T) {
	t.Parallel()
	base := map[string]string{"protected": strings.Repeat("a", 64)}
	current := map[string]snapshotFile{
		"protected":        {digest: strings.Repeat("a", 64), regular: true, links: 1},
		"semantic/unknown": {digest: strings.Repeat("b", 64), regular: true, links: 1},
	}
	if _, findings := validateSnapshot(base, current, expectedWave02Paths); findingRules(findings)["SCOPE-354"] == 0 {
		t.Fatalf("unknown semantic path was accepted: %+v", findings)
	}
	root := materializeMapFS(t, fstest.MapFS{"cmd/l7/main.go": {Data: []byte("package main\n")}})
	if findings := checkForbiddenProductPaths(root); findingRules(findings)["SCOPE-379"] == 0 {
		t.Fatalf("forbidden product prefix was accepted: %+v", findings)
	}
}
