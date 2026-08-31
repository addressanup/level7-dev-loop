#!/bin/sh

set -eu
umask 077
PATH=/usr/bin:/bin
export PATH

fail()
{
	printf 'bootstrap-modules: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 2 || fail 'usage: bootstrap-modules.sh /absolute/path/to/go VERSION'
go_bin=$1
go_version=$2

case $go_bin in /*) ;; *) fail 'Go binary path must be absolute' ;; esac
case $go_version in ''|*[!0-9.]*) fail 'Go version must contain only digits and periods' ;; esac

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
expected_go="$project_root/.cache/toolchains/go$go_version/bin/go"
test "$go_bin" = "$expected_go" || fail "Go binary is outside the pinned toolchain: $go_bin"
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"
test ! -L "$go_bin" || fail "Go binary is a symlink: $go_bin"

go_mod="$project_root/go.mod"
go_sum="$project_root/go.sum"
for input in "$go_mod" "$go_sum"; do
	test -f "$input" || fail "dependency input is missing: $input"
	test ! -L "$input" || fail "dependency input is a symlink: $input"
done

cache_root="$project_root/.cache/go"
module_cache="$cache_root/mod"
build_cache="$cache_root/build"
temp_dir="$cache_root/tmp"
telemetry_dir="$cache_root/telemetry"

for directory in "$cache_root" "$module_cache" "$build_cache" "$temp_dir" "$telemetry_dir"; do
	test -d "$directory" || fail "prepared cache directory is missing: $directory"
	test ! -L "$directory" || fail "prepared cache directory is a symlink: $directory"
	physical=$(CDPATH='' cd "$directory" && pwd -P)
	case $physical in "$project_root"/.cache/go/*|"$project_root"/.cache/go) ;; *) fail "cache directory escapes the repository: $directory" ;; esac
done

bootstrap_home=$(mktemp -d "$temp_dir/l7-bootstrap-home.XXXXXX")
case $bootstrap_home in "$temp_dir"/l7-bootstrap-home.*) ;; *) fail 'mktemp returned an unsafe bootstrap home' ;; esac
cleanup()
{
	case $bootstrap_home in "$temp_dir"/l7-bootstrap-home.*) rm -rf -- "$bootstrap_home" ;; *) return 1 ;; esac
}
trap cleanup EXIT HUP INT TERM

sha256_file()
{
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{ print $1 }'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{ print $1 }'
	else
		fail 'sha256sum or shasum is required'
	fi
}

go_mod_sha256=$(sha256_file "$go_mod")
go_sum_sha256=$(sha256_file "$go_sum")

verify_inputs()
{
	test -f "$go_mod" && test ! -L "$go_mod" || fail 'go.mod changed type during bootstrap'
	test -f "$go_sum" && test ! -L "$go_sum" || fail 'go.sum changed type during bootstrap'
	test "$(sha256_file "$go_mod")" = "$go_mod_sha256" || fail 'go.mod changed during bootstrap'
	test "$(sha256_file "$go_sum")" = "$go_sum_sha256" || fail 'go.sum changed during bootstrap'
}

run_go()
{
	mode=$1
	shift
	case $mode in
		network)
			proxy=https://proxy.golang.org
			sumdb=sum.golang.org
			;;
		offline)
			proxy=off
			sumdb=off
			;;
		*) fail "invalid Go execution mode: $mode" ;;
	esac
	status=0
	env -i \
		PATH=/usr/bin:/bin \
		HOME="$bootstrap_home" \
		TMPDIR="$temp_dir" \
		LC_ALL=C \
		TZ=UTC \
		GOENV=off \
		GOTOOLCHAIN=local \
		GOWORK=off \
		GO111MODULE=on \
		GOFLAGS= \
		GOEXPERIMENT= \
		GOFIPS140=off \
		CGO_ENABLED=0 \
		GOMODCACHE="$module_cache" \
		GOCACHE="$build_cache" \
		GOTMPDIR="$temp_dir" \
		GOPROXY="$proxy" \
		GOSUMDB="$sumdb" \
		GOPRIVATE= \
		GONOPROXY= \
		GONOSUMDB= \
		GOINSECURE= \
		GOVCS='*:off' \
		GOAUTH=off \
		GOTELEMETRY=off \
		TEST_TELEMETRY_DIR="$telemetry_dir" \
		"$go_bin" "$@" || status=$?
	verify_inputs
	if test "$status" -ne 0; then
		printf 'bootstrap-modules: go command failed with status %s: %s\n' "$status" "$*" >&2
		return "$status"
	fi
}

actual_version=$(run_go offline env GOVERSION)
test "$actual_version" = "go$go_version" || fail "toolchain reports $actual_version, expected go$go_version"
run_go network mod download all
run_go offline mod verify
run_go offline list -mod=readonly -m all >/dev/null
verify_inputs

printf 'bootstrap-modules: PASS (go%s; fixed proxy and checksum database; repository-local cache; downstream offline)\n' "$go_version"
