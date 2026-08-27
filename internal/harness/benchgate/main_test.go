package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsPairedMediansAtTheThreshold(t *testing.T) {
	base := writeBenchmarkFixture(t, benchmarkOutput("BenchmarkParseStatus10000Paths", 100, 102, 98, 101, 99))
	candidate := writeBenchmarkFixture(t, benchmarkOutput("BenchmarkParseStatus10000Paths", 111, 110, 108, 109, 112))
	var stdout, stderr bytes.Buffer
	if exit := run([]string{base, candidate}, &stdout, &stderr); exit != 0 || stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
	for _, expected := range []string{"base_median_ns_op=100.000", "candidate_median_ns_op=110.000", "change_percent=+10.00", "result=PASS", "benchgate: PASS"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("stdout=%q missing %q", stdout.String(), expected)
		}
	}
}

func TestRunBlocksRegressionAboveTheThreshold(t *testing.T) {
	base := writeBenchmarkFixture(t, benchmarkOutput("BenchmarkSnapshot10000Paths", 100, 100, 100, 100, 100))
	candidate := writeBenchmarkFixture(t, benchmarkOutput("BenchmarkSnapshot10000Paths", 111, 111, 111, 111, 111))
	var stdout, stderr bytes.Buffer
	if exit := run([]string{base, candidate}, &stdout, &stderr); exit != 2 || !strings.Contains(stdout.String(), "result=BLOCKED") || !strings.Contains(stderr.String(), "accountable-owner") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
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
