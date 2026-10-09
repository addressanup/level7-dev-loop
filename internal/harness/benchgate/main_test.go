package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runGate(t *testing.T, base, candidate string, flags ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	exit := run(append(flags, writeBenchmarkFixture(t, base), writeBenchmarkFixture(t, candidate)), &stdout, &stderr)
	return exit, stdout.String(), stderr.String()
}

func TestRunPassesAPairedChangeExactlyAtTheThreshold(t *testing.T) {
	exit, stdout, stderr := runGate(t,
		benchmarkOutput("BenchmarkParseStatus10000Paths", 100, 200, 300, 400, 500, 600, 700, 800, 900),
		benchmarkOutput("BenchmarkParseStatus10000Paths", 110, 220, 330, 440, 550, 660, 770, 880, 990))
	if exit != 0 || stderr != "" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
	for _, expected := range []string{"paired_change_percent=+10.00", "slower_pairs=9/9", "result=PASS", "benchgate: PASS threshold_percent=10.00 minimum_slower_pairs=8"} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("stdout=%q missing %q", stdout, expected)
		}
	}
}

func TestRunBlocksAConsistentRegression(t *testing.T) {
	exit, stdout, stderr := runGate(t,
		benchmarkOutput("BenchmarkSnapshot10000Paths", 100, 104, 97, 101, 99, 103, 98, 102, 100),
		benchmarkOutput("BenchmarkSnapshot10000Paths", 116, 119, 111, 117, 113, 119, 112, 118, 115))
	if exit != 2 || !strings.Contains(stdout, "slower_pairs=9/9 result=BLOCKED") || !strings.Contains(stderr, "accountable-owner") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
}

func TestRunNeedsConsistentPairsToBlock(t *testing.T) {
	base := benchmarkOutput("BenchmarkSnapshot10000Paths", 100, 100, 100, 100, 100, 100, 100, 100, 100)
	if exit, stdout, _ := runGate(t, base, benchmarkOutput("BenchmarkSnapshot10000Paths", 120, 120, 120, 120, 120, 120, 120, 90, 90)); exit != 0 || !strings.Contains(stdout, "paired_change_percent=+20.00 slower_pairs=7/9 result=PASS") {
		t.Fatalf("a slowdown in seven of nine pairs must pass: exit=%d stdout=%q", exit, stdout)
	}
	if exit, stdout, _ := runGate(t, base, benchmarkOutput("BenchmarkSnapshot10000Paths", 120, 120, 120, 120, 120, 120, 120, 120, 90)); exit != 2 || !strings.Contains(stdout, "slower_pairs=8/9 result=BLOCKED") {
		t.Fatalf("a slowdown in eight of nine pairs must block: exit=%d stdout=%q", exit, stdout)
	}
}

func TestRunIgnoresOneSlowOutlierPair(t *testing.T) {
	exit, stdout, _ := runGate(t,
		benchmarkOutput("BenchmarkSnapshot10000Paths", 100, 100, 100, 100, 100, 100, 100, 100, 100),
		benchmarkOutput("BenchmarkSnapshot10000Paths", 101, 99, 100, 102, 98, 100, 101, 99, 250))
	if exit != 0 || !strings.Contains(stdout, "paired_change_percent=+0.00") {
		t.Fatalf("one outlier pair must not block: exit=%d stdout=%q", exit, stdout)
	}
}

// TestRunScoresCapturedRunnerNoise replays failed gate runs from PRs #22 and
// #23, where the benchmarked package was identical between base and
// candidate. Five samples with all pairs slower mirror the brief's table.
func TestRunScoresCapturedRunnerNoise(t *testing.T) {
	tests := []struct {
		name            string
		base, candidate []float64
		medianChange    string
		blocked         bool
	}{
		{"#22 first run Snapshot", []float64{131210908, 120760454, 98802167, 105979775, 108363767}, []float64{132105233, 130862846, 103028754, 112434617, 189048992}, "median_change_percent=+20.76", false},
		{"#22 re-run ParseStatus", []float64{1971090, 2011836, 1809567, 2655626, 1943835}, []float64{2128387, 2354792, 1941414, 2322868, 2194373}, "median_change_percent=+11.33", false},
		{"#22 re-run Snapshot", []float64{128684971, 133692617, 132460758, 159209758, 201888188}, []float64{138504508, 138893858, 207209296, 153174646, 179156767}, "median_change_percent=+14.57", false},
		{"#23 first run ParseStatus", []float64{1835680, 1641807, 1579209, 1654727, 1637944}, []float64{1907590, 1950125, 1998992, 1593184, 1751033}, "median_change_percent=+16.19", false},
		{"#23 re-run Snapshot", []float64{106842208, 83107988, 82602583, 83231183, 75282438}, []float64{108986625, 86769542, 93276079, 92962412, 83695208}, "median_change_percent=+11.86", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exit, stdout, stderr := runGate(t, benchmarkOutput("BenchmarkCaptured", test.base...), benchmarkOutput("BenchmarkCaptured", test.candidate...), "--minimum-samples", "5", "--minimum-slower-pairs", "5")
			if !strings.Contains(stdout, test.medianChange) {
				t.Fatalf("the old group-median rule result is not reported: %q", stdout)
			}
			if blocked := exit == 2; blocked != test.blocked || (exit != 0 && exit != 2) {
				t.Fatalf("exit=%d blocked=%t want blocked=%t stdout=%q stderr=%q", exit, blocked, test.blocked, stdout, stderr)
			}
		})
	}
}

func TestRunRejectsInvalidSlowerPairCounts(t *testing.T) {
	data := benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1, 1, 1, 1, 1)
	for _, flags := range [][]string{{"--minimum-slower-pairs", "0"}, {"--minimum-slower-pairs", "10"}, {"--minimum-samples", "5", "--minimum-slower-pairs", "6"}} {
		if exit, _, stderr := runGate(t, data, data, flags...); exit != 1 || !strings.Contains(stderr, "minimum-slower-pairs") {
			t.Fatalf("flags %v: exit=%d stderr=%q", flags, exit, stderr)
		}
	}
}

func TestRunRejectsUnpairedMalformedAndUndersampledData(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		candidate string
	}{
		{"different names", benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1), benchmarkOutput("BenchmarkTwo", 1, 1, 1, 1, 1)},
		{"different counts", benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1), benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1, 1)},
		{"undersampled", benchmarkOutput("BenchmarkOne", 1, 1, 1, 1), benchmarkOutput("BenchmarkOne", 1, 1, 1, 1)},
		{"missing metric", "BenchmarkOne-8 1 1 B/op\n", benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1)},
		{"unsafe name", "BenchmarkBad:name-8 1 1 ns/op\n", benchmarkOutput("BenchmarkOne", 1, 1, 1, 1, 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run([]string{writeBenchmarkFixture(t, test.base), writeBenchmarkFixture(t, test.candidate)}, &stdout, &stderr)
			if exit != 1 || stderr.Len() == 0 {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
			}
		})
	}
}

func TestParseBenchmarkOutputNormalizesCPUAndSortsComparisons(t *testing.T) {
	data := []byte(benchmarkOutput("BenchmarkZulu", 20, 21, 22, 23, 24) + benchmarkOutput("BenchmarkAlpha/sub=one", 10, 11, 12, 13, 14))
	samples, err := parseBenchmarkOutput(data)
	if err != nil || len(samples["BenchmarkAlpha/sub=one"]) != 5 || len(samples["BenchmarkZulu"]) != 5 {
		t.Fatalf("samples=%v error=%v", samples, err)
	}
	comparisons, err := compareBenchmarks(samples, samples, 5)
	if err != nil || len(comparisons) != 2 || comparisons[0].name != "BenchmarkAlpha/sub=one" || comparisons[1].name != "BenchmarkZulu" {
		t.Fatalf("comparisons=%+v error=%v", comparisons, err)
	}
}

func benchmarkOutput(name string, values ...float64) string {
	var output strings.Builder
	for _, value := range values {
		fmt.Fprintf(&output, "%s-12 100 %.3f ns/op 0 B/op 0 allocs/op\n", name, value)
	}
	return output.String()
}

func writeBenchmarkFixture(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "benchmark.txt")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
