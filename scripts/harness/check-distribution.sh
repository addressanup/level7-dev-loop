#!/bin/sh

set -eu

fail()
{
	printf 'check-distribution: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 1 || fail 'usage: check-distribution.sh /absolute/path/to/go'
go_bin=$1
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
case $project_root in /*) ;; *) fail 'project root is not absolute' ;; esac

cd "$project_root"
"$go_bin" run -mod=readonly ./internal/harness/distribution --root "$project_root" --check
