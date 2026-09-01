#!/bin/sh

set -eu

fail()
{
	printf 'check-l7-fuzz: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 1 || fail 'usage: check-l7-fuzz.sh /absolute/path/to/go'
go_bin=$1
case $go_bin in /*) ;; *) fail 'Go binary path must be absolute' ;; esac
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
test -z "$(git -C "$project_root" status --porcelain --untracked-files=all)" || fail 'worktree and index must be clean for exact-head fuzzing'
GOENV=off
GOTOOLCHAIN=local
GOWORK=off
GOPROXY=off
GOSUMDB=off
GOVCS='*:off'
GOAUTH=off
export GOENV GOTOOLCHAIN GOWORK GOPROXY GOSUMDB GOVCS GOAUTH
mkdir -p "$project_root/build"
scratch=$(mktemp -d "$project_root/build/l7-fuzz.XXXXXX")
cleanup()
{
	case $scratch in "$project_root"/build/l7-fuzz.*) rm -rf -- "$scratch" ;; *) return 1 ;; esac
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM
mkdir "$scratch/repository" "$scratch/cache" "$scratch/tmp" "$scratch/telemetry"

git -C "$project_root" archive --format=tar HEAD | tar -xf - -C "$scratch/repository"
cd "$scratch/repository"

inventory='
./internal/l7/adapter/ci FuzzDecode 5s
./internal/l7/adapter/toolbroker FuzzPatchPathContainment 5s
./internal/l7/adapter/claude FuzzParseResult 5s
./internal/l7/adapter/codex FuzzParseEvents 5s
./internal/l7/adapter/codexapp FuzzParseTerminal 5s
./internal/l7/adapter/orchestrationconfig FuzzStrictConfigurationDecode 10000x
./internal/l7/adapter/provider FuzzParseTerminal 5s
./internal/l7/adapter/gateway FuzzDecodeEventStream 5s'

expected_plan=$(printf '%s\n' "$inventory" | awk 'NF == 3' | LC_ALL=C sort)
expected_inventory=$(printf '%s\n' "$expected_plan" | awk '{print $1, $2}')
test "$(printf '%s\n' "$expected_plan" | awk '$2 == "FuzzStrictConfigurationDecode" && $3 == "10000x" {count++} END {print count + 0}')" -eq 1 || fail 'strict configuration fuzz budget changed'
test "$(printf '%s\n' "$expected_plan" | awk '$2 != "FuzzStrictConfigurationDecode" && $3 == "5s" {count++} END {print count + 0}')" -eq 7 || fail 'unaffected fuzz budgets changed'
module=$(GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" "$go_bin" list -m -f '{{.Path}}')
observed_inventory=''
packages=$(GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" "$go_bin" list -mod=readonly ./internal/l7/...)
for import_path in $packages; do
	case $import_path in "$module"/*) package=./${import_path#"$module"/} ;; *) fail "unexpected package identity: $import_path" ;; esac
	targets=$(CGO_ENABLED=1 GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" \
		"$go_bin" test -mod=readonly -trimpath -buildvcs=false -run '^$' -list '^Fuzz' "$package" | sed -n '/^Fuzz/p')
	for target in $targets; do
		observed_inventory="${observed_inventory}${package} ${target}
"
	done
done
observed_inventory=$(printf '%s' "$observed_inventory" | awk 'NF == 2' | LC_ALL=C sort)
test "$observed_inventory" = "$expected_inventory" || fail "repository fuzz inventory changed:\nobserved:\n$observed_inventory\nexpected:\n$expected_inventory"

printf '%s\n' "$expected_plan" | while read -r package target budget; do
	test -n "$package" || continue
	CGO_ENABLED=1 GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" \
		"$go_bin" test -mod=readonly -trimpath -buildvcs=false -run '^$' -fuzz "^$target$" -fuzztime="$budget" -parallel=1 -timeout=2m "$package"
done

printf 'check-l7-fuzz: PASS (8 fixed targets; strict configuration 10000x; 7 targets 5s each; pinned offline toolchain; disposable corpus root)\n'
