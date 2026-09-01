#!/bin/sh

set -eu

fail()
{
	printf 'check-v1-conformance: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 1 || fail 'usage: check-v1-conformance.sh /absolute/path/to/go'
go_bin=$1
case $go_bin in /*) ;; *) fail 'Go binary path must be absolute' ;; esac
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"
test "$(uname -s)" = Darwin || fail 'native conformance requires macOS'
case $(uname -m) in arm64|aarch64|x86_64|amd64) ;; *) fail 'unsupported macOS architecture' ;; esac

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
test -z "$(git -C "$project_root" status --porcelain --untracked-files=all)" || fail 'worktree and index must be clean for exact-head conformance'
GOENV=off
GOTOOLCHAIN=local
GOWORK=off
GOPROXY=off
GOSUMDB=off
GOVCS='*:off'
GOAUTH=off
export GOENV GOTOOLCHAIN GOWORK GOPROXY GOSUMDB GOVCS GOAUTH

mkdir -p "$project_root/build"
scratch=$(mktemp -d "$project_root/build/v1-conformance.XXXXXX")
cleanup()
{
	case $scratch in "$project_root"/build/v1-conformance.*) rm -rf -- "$scratch" ;; *) return 1 ;; esac
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM
mkdir "$scratch/repository" "$scratch/work" "$scratch/cache" "$scratch/tmp" "$scratch/telemetry"
git -C "$project_root" archive --format=tar HEAD | tar -xf - -C "$scratch/repository"

(
	cd "$scratch/repository"
	CGO_ENABLED=0 GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" \
		L7_NETWORK=off L7_TELEMETRY=off GOPROXY=off GOSUMDB=off GOVCS='*:off' GOAUTH=off \
		"$go_bin" run -mod=readonly ./internal/harness/distribution --root "$scratch/repository" --output "$scratch/repository/build/distributions"
	CGO_ENABLED=0 GOCACHE="$scratch/cache" GOTMPDIR="$scratch/tmp" TMPDIR="$scratch/tmp" TEST_TELEMETRY_DIR="$scratch/telemetry" \
		L7_NETWORK=off L7_TELEMETRY=off GOPROXY=off GOSUMDB=off GOVCS='*:off' GOAUTH=off \
		"$go_bin" run -mod=readonly ./internal/harness/v1candidate \
		--candidate "$project_root/build/v1-candidate" \
		--stable "$scratch/repository/build/distributions" \
		--work "$scratch/work"
)

test -z "$(find "$scratch/work" -mindepth 1 -print -quit)" || fail 'conformance harness left disposable lifecycle state'
test -z "$(git -C "$project_root" status --porcelain --untracked-files=all)" || fail 'conformance changed repository state'
printf 'check-v1-conformance: PASS (Codex and Claude; native CLI/MCP; stable upgrade/rollback/removal; disposable roots)\n'
