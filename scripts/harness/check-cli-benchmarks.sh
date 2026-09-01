#!/bin/sh

set -eu

fail()
{
	printf 'check-cli-benchmarks: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 3 || fail 'usage: check-cli-benchmarks.sh /absolute/path/to/go BASE_ROOT CANDIDATE_ROOT'
go_bin=$1
base_root=$2
candidate_root=$3
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"
test -d "$base_root" || fail "base root is not a directory: $base_root"
test -d "$candidate_root" || fail "candidate root is not a directory: $candidate_root"
base_root=$(CDPATH='' cd "$base_root" && pwd -P)
candidate_root=$(CDPATH='' cd "$candidate_root" && pwd -P)
test "$base_root" != "$candidate_root" || fail 'base and candidate roots must differ'
test -f "$base_root/go.mod" || fail 'base root has no go.mod'
test -f "$candidate_root/go.mod" || fail 'candidate root has no go.mod'

base_module=$(cd "$base_root" && "$go_bin" list -m -f '{{.Path}}')
candidate_module=$(cd "$candidate_root" && "$go_bin" list -m -f '{{.Path}}')
test "$base_module" = "$candidate_module" || fail 'base and candidate module paths differ'

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=${temporary_parent%/}
benchmark_root=$(mktemp -d "$temporary_parent/l7-cli-benchmark.XXXXXX")
case $benchmark_root in "$temporary_parent"/l7-cli-benchmark.*) ;; *) fail 'mktemp returned an unsafe benchmark directory' ;; esac
cleanup()
{
	rm -rf -- "$benchmark_root"
}
trap cleanup EXIT HUP INT TERM
mkdir "$benchmark_root/cache" "$benchmark_root/tmp"
base_output=$benchmark_root/base.txt
candidate_output=$benchmark_root/candidate.txt
: >"$base_output"
: >"$candidate_output"

export GOENV=off GOTOOLCHAIN=local GOWORK=off GO111MODULE=on GOFLAGS=
export CGO_ENABLED=0 GOPROXY=off GOSUMDB=off GOVCS='*:off' GOAUTH=off
export GOCACHE=$benchmark_root/cache GOTMPDIR=$benchmark_root/tmp TMPDIR=$benchmark_root/tmp
export GOTELEMETRY=off LC_ALL=C TZ=UTC
unset GOPRIVATE GONOPROXY GONOSUMDB GOINSECURE GOEXPERIMENT GOFIPS140

run_benchmark()
{
	root=$1
	output=$2
	benchmark_pattern=$3
	benchmark_budget=$4
	(
		cd "$root"
		"$go_bin" test -mod=readonly -trimpath -buildvcs=false -run '^$' -bench "$benchmark_pattern" -benchtime="$benchmark_budget" -count=1 ./internal/l7/adapter/git
	) >>"$output"
}

run_sample()
{
	sample_root=$1
	sample_output=$2
	run_benchmark "$sample_root" "$sample_output" '^BenchmarkParseStatus10000Paths$' 250x
	run_benchmark "$sample_root" "$sample_output" '^BenchmarkSnapshot10000Paths$' 10x
}

for sample in 1 2 3 4 5; do
	case $sample in
		1|3|5)
			run_sample "$base_root" "$base_output"
			run_sample "$candidate_root" "$candidate_output"
			;;
		2|4)
			run_sample "$candidate_root" "$candidate_output"
			run_sample "$base_root" "$base_output"
			;;
	esac
done

printf 'check-cli-benchmarks: host=%s/%s toolchain=%s samples=5 parse_status_benchtime=250x snapshot_benchtime=10x\n' \
	"$("$go_bin" env GOHOSTOS)" "$("$go_bin" env GOHOSTARCH)" "$("$go_bin" env GOVERSION)"
(
	cd "$candidate_root"
	"$go_bin" run -mod=readonly ./internal/harness/benchgate --threshold-percent 10 --minimum-samples 5 "$base_output" "$candidate_output"
)
