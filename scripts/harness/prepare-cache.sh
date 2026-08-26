#!/bin/sh

set -eu
umask 077

fail()
{
	printf 'prepare-cache: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 2 || fail 'usage: prepare-cache.sh /absolute/project/root GO_VERSION'
project_root=$1
go_version=$2

case $project_root in
	/*) ;;
	*) fail 'project root must be absolute' ;;
esac
case $go_version in
	''|*[!0-9.]*) fail 'Go version must contain only digits and periods' ;;
esac

test -d "$project_root" || fail "project root is not a directory: $project_root"
test ! -L "$project_root" || fail "project root is a symlink: $project_root"
physical_root=$(CDPATH='' cd "$project_root" && pwd -P)
test "$physical_root" = "$project_root" || fail "project root is not its canonical physical path: $project_root"

cache_root="$project_root/.cache"
go_root="$cache_root/go"
go_path="$go_root/path"
go_bin="$go_root/bin"
go_build="$go_root/build"
go_mod="$go_root/mod"
go_tmp="$go_root/tmp"
telemetry_root="$go_root/telemetry"
repro_root="$cache_root/repro"
toolchains_root="$cache_root/toolchains"
toolchain_root="$toolchains_root/go$go_version"
telemetry_mode="$telemetry_root/mode"

check_existing_directory()
{
	directory=$1
	test ! -L "$directory" || fail "refusing symlinked cache path: $directory"
	if test -e "$directory"; then
		test -d "$directory" || fail "cache path is not a directory: $directory"
		physical=$(CDPATH='' cd "$directory" && pwd -P)
		test "$physical" = "$directory" || fail "cache path escapes the physical repository root: $directory"
	fi
}

# Complete the read-only preflight before creating a directory or replacing the
# telemetry-mode file. Parent-first order prevents a symlinked ancestor from
# being traversed while inspecting a descendant.
for directory in \
	"$cache_root" \
	"$go_root" \
	"$go_path" \
	"$go_bin" \
	"$go_build" \
	"$go_mod" \
	"$go_tmp" \
	"$telemetry_root" \
	"$repro_root" \
	"$toolchains_root" \
	"$toolchain_root"
do
	check_existing_directory "$directory"
done
test ! -L "$telemetry_mode" || fail "refusing symlinked telemetry mode: $telemetry_mode"
if test -e "$telemetry_mode"; then
	test -f "$telemetry_mode" || fail "telemetry mode is not a regular file: $telemetry_mode"
fi

for directory in \
	"$cache_root" \
	"$go_root" \
	"$go_path" \
	"$go_bin" \
	"$go_build" \
	"$go_mod" \
	"$go_tmp" \
	"$telemetry_root" \
	"$repro_root"
do
	if test ! -d "$directory"; then
		mkdir "$directory"
	fi
	check_existing_directory "$directory"
done

mode_temp=$(mktemp "$telemetry_root/.mode.XXXXXX")
cleanup_mode_temp()
{
	rm -f "$mode_temp"
}
trap cleanup_mode_temp EXIT HUP INT TERM
printf 'off 2026-08-24\n' >"$mode_temp"
mv -f "$mode_temp" "$telemetry_mode"
mode_temp=''
trap - EXIT HUP INT TERM

test -f "$telemetry_mode" || fail "telemetry mode was not created: $telemetry_mode"
test ! -L "$telemetry_mode" || fail "telemetry mode became a symlink: $telemetry_mode"
