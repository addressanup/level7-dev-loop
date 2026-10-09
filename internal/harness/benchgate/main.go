package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultThresholdPercent   = 10.0
	defaultMinimumSamples     = 9
	defaultMinimumSlowerPairs = 8
	maxBenchmarkBytes         = 8 << 20
	maxBenchmarkNames         = 128
	maxSamplesPerBenchmark    = 100
)

type benchmarkSamples map[string][]float64

// benchmarkComparison pairs sample i of the base with sample i of the
// candidate. The checker script runs each pair back to back, so a pair sees
// the same machine state and its ratio cancels drift between pairs.
type benchmarkComparison struct {
	name                string
	baseSamples         []float64
	candidateSamples    []float64
	baseMedian          float64
	candidateMedian     float64
	medianChangePercent float64
	medianRatio         float64
	slowerPairs         int
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("benchgate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	threshold := flags.Float64("threshold-percent", defaultThresholdPercent, "maximum allowed median paired regression percentage")
	minimumSamples := flags.Int("minimum-samples", defaultMinimumSamples, "minimum paired samples required per benchmark")
	minimumSlowerPairs := flags.Int("minimum-slower-pairs", defaultMinimumSlowerPairs, "slower pairs required, with the threshold, to block")
	if err := flags.Parse(arguments); err != nil {
		return 1
	}
	if flags.NArg() != 2 {
		fmt.Fprintln(stderr, "benchgate: usage: benchgate [--threshold-percent 10] [--minimum-samples 9] [--minimum-slower-pairs 8] BASE.txt CANDIDATE.txt")
		return 1
	}
	if math.IsNaN(*threshold) || math.IsInf(*threshold, 0) || *threshold < 0 || *threshold > 1000 {
		fmt.Fprintln(stderr, "benchgate: threshold-percent must be finite and between 0 and 1000")
		return 1
	}
	if *minimumSamples < 1 || *minimumSamples > maxSamplesPerBenchmark {
		fmt.Fprintf(stderr, "benchgate: minimum-samples must be between 1 and %d\n", maxSamplesPerBenchmark)
		return 1
	}
	if *minimumSlowerPairs < 1 || *minimumSlowerPairs > *minimumSamples {
		fmt.Fprintln(stderr, "benchgate: minimum-slower-pairs must be between 1 and minimum-samples")
		return 1
	}

	base, err := loadBenchmarkFile(flags.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "benchgate: invalid base benchmark data: %v\n", err)
		return 1
	}
	candidate, err := loadBenchmarkFile(flags.Arg(1))
	if err != nil {
		fmt.Fprintf(stderr, "benchgate: invalid candidate benchmark data: %v\n", err)
		return 1
	}
	comparisons, err := compareBenchmarks(base, candidate, *minimumSamples)
	if err != nil {
		fmt.Fprintf(stderr, "benchgate: incomparable benchmark data: %v\n", err)
		return 1
	}

	blocked := false
	for _, comparison := range comparisons {
		result := "PASS"
		if comparison.medianRatio > 1+*threshold/100 && comparison.slowerPairs >= *minimumSlowerPairs {
			result = "BLOCKED"
			blocked = true
		}
		fmt.Fprintf(
			stdout,
			"benchmark=%s samples=%d base_samples_ns_op=%s candidate_samples_ns_op=%s base_median_ns_op=%.3f candidate_median_ns_op=%.3f median_change_percent=%+.2f paired_change_percent=%+.2f slower_pairs=%d/%d result=%s\n",
			comparison.name,
			len(comparison.baseSamples),
			formatSamples(comparison.baseSamples),
			formatSamples(comparison.candidateSamples),
			comparison.baseMedian,
			comparison.candidateMedian,
			comparison.medianChangePercent,
			(comparison.medianRatio-1)*100,
			comparison.slowerPairs,
			len(comparison.baseSamples),
			result,
		)
	}
	if blocked {
		fmt.Fprintf(stderr, "benchgate: BLOCKED threshold_percent=%.2f minimum_slower_pairs=%d benchmark_count=%d; explicit accountable-owner acceptance must occur outside candidate-controlled inputs\n", *threshold, *minimumSlowerPairs, len(comparisons))
		return 2
	}
	fmt.Fprintf(stdout, "benchgate: PASS threshold_percent=%.2f minimum_slower_pairs=%d benchmark_count=%d\n", *threshold, *minimumSlowerPairs, len(comparisons))
	return 0
}

func loadBenchmarkFile(path string) (benchmarkSamples, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	limited := io.LimitReader(file, maxBenchmarkBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maxBenchmarkBytes {
		return nil, errors.New("input is empty or exceeds the size limit")
	}
	return parseBenchmarkOutput(data)
}

func parseBenchmarkOutput(data []byte) (benchmarkSamples, error) {
	samples := make(benchmarkSamples)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			return nil, errors.New("benchmark record is incomplete")
		}
		name, ok := normalizeBenchmarkName(fields[0])
		if !ok {
			return nil, fmt.Errorf("unsafe benchmark name %q", fields[0])
		}
		metric := -1
		for index, field := range fields {
			if field == "ns/op" {
				if metric != -1 {
					return nil, fmt.Errorf("benchmark %s repeats ns/op", name)
				}
				metric = index
			}
		}
		if metric < 1 {
			return nil, fmt.Errorf("benchmark %s has no ns/op metric", name)
		}
		value, err := strconv.ParseFloat(fields[metric-1], 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
			return nil, fmt.Errorf("benchmark %s has an invalid ns/op value", name)
		}
		if _, found := samples[name]; !found && len(samples) >= maxBenchmarkNames {
			return nil, errors.New("benchmark count exceeds the size limit")
		}
		if len(samples[name]) >= maxSamplesPerBenchmark {
			return nil, fmt.Errorf("benchmark %s exceeds the sample limit", name)
		}
		samples[name] = append(samples[name], value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(samples) == 0 {
		return nil, errors.New("input contains no Go benchmark records")
	}
	return samples, nil
}

func normalizeBenchmarkName(value string) (string, bool) {
	if len(value) < len("Benchmark")+1 || len(value) > 256 || !strings.HasPrefix(value, "Benchmark") {
		return "", false
	}
	if separator := strings.LastIndexByte(value, '-'); separator > len("Benchmark") && decimal(value[separator+1:]) {
		value = value[:separator]
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') {
			continue
		}
		switch character {
		case '/', '_', '-', '.', '=':
			continue
		default:
			return "", false
		}
	}
	return value, true
}

func decimal(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func compareBenchmarks(base, candidate benchmarkSamples, minimumSamples int) ([]benchmarkComparison, error) {
	if len(base) != len(candidate) {
		return nil, errors.New("base and candidate benchmark sets differ")
	}
	names := make([]string, 0, len(base))
	for name, baseValues := range base {
		candidateValues, found := candidate[name]
		if !found {
			return nil, fmt.Errorf("candidate is missing benchmark %s", name)
		}
		if len(baseValues) != len(candidateValues) {
			return nil, fmt.Errorf("benchmark %s has unpaired sample counts", name)
		}
		if len(baseValues) < minimumSamples {
			return nil, fmt.Errorf("benchmark %s has %d samples; need at least %d", name, len(baseValues), minimumSamples)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	comparisons := make([]benchmarkComparison, 0, len(names))
	for _, name := range names {
		baseMedian := median(base[name])
		candidateMedian := median(candidate[name])
		ratios := make([]float64, len(base[name]))
		slower := 0
		for index, baseValue := range base[name] {
			ratios[index] = candidate[name][index] / baseValue
			if ratios[index] > 1 {
				slower++
			}
		}
		comparisons = append(comparisons, benchmarkComparison{
			name: name, baseSamples: append([]float64(nil), base[name]...), candidateSamples: append([]float64(nil), candidate[name]...),
			baseMedian: baseMedian, candidateMedian: candidateMedian, medianChangePercent: ((candidateMedian / baseMedian) - 1) * 100,
			medianRatio: median(ratios), slowerPairs: slower,
		})
	}
	return comparisons, nil
}

func median(values []float64) float64 {
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func formatSamples(values []float64) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = strconv.FormatFloat(value, 'f', 3, 64)
	}
	return strings.Join(parts, ",")
}
