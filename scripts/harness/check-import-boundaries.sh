#!/bin/sh

set -eu

fail()
{
	printf 'check-import-boundaries: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 1 || fail 'usage: check-import-boundaries.sh /absolute/path/to/go'
go_bin=$1
test -x "$go_bin" || fail "Go binary is not executable: $go_bin"

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
policy_file="$project_root/harness/import-boundaries.tsv"
modules_file="$project_root/harness/modules.lock.tsv"
cd "$project_root"
module=$($go_bin list -m -f '{{.Path}}')
packages=$($go_bin list -f '{{.ImportPath}}' ./...)

core_record=$(awk -F '\t' '$1 == "core" && $2 == "active" { print }' "$modules_file")
core_count=$(printf '%s\n' "$core_record" | awk 'NF { count++ } END { print count + 0 }')
test "$core_count" -eq 1 || fail 'module registry must contain exactly one active core'
tab=$(printf '\t')
IFS=$tab read -r core_role core_state core_directory core_module <<EOF
$core_record
EOF
test "$core_role" = 'core' || fail 'active module role mismatch'
test "$core_state" = 'active' || fail 'active module state mismatch'
test "$core_directory" = '.' || fail 'active core directory must be repository root'
test "$module" = "$core_module" || fail "root module $module does not match registry $core_module"

while IFS="$tab" read -r module_role module_state module_directory module_path; do
	case $module_role in ''|'#'*) continue ;; esac
	if test "$module_state" = 'reserved'; then
		test ! -e "$project_root/$module_directory" || fail "BND-000: reserved $module_role module path exists; activate module-aware enforcement before adding it"
		test "$module_path" = 'UNSET' || fail "reserved $module_role module path must remain UNSET"
	fi
done <"$modules_file"

matches_prefix()
{
	case $1 in
		"$2"|"$2"/*) return 0 ;;
		*) return 1 ;;
	esac
}

while IFS="$tab" read -r mode source_prefix forbidden_prefix rule; do
	case $mode in ''|'#'*) continue ;; allowclosure|effect|transitive) ;; *) fail "unknown policy mode: $mode" ;; esac
	test -n "$source_prefix" || fail "missing source prefix in $rule"
	test -n "$forbidden_prefix" || fail "missing forbidden prefix in $rule"
	test -n "$rule" || fail 'missing boundary rule identifier'

	for package in $packages; do
		relative_package=${package#"$module/"}
		matches_prefix "$relative_package" "$source_prefix" || continue

		case $mode in
			allowclosure)
				test "$relative_package" = "$source_prefix" || fail "$rule: updater must remain one entry package, found $relative_package"
				dependencies=$($go_bin list -deps -f '{{.ImportPath}}' "$package")
				for dependency in $dependencies; do
					test "$dependency" = "$package" && continue
					if matches_prefix "$dependency" "$module/$forbidden_prefix"; then
						continue
					fi
					if test "$($go_bin list -f '{{.Standard}}' "$dependency")" = 'true'; then
						continue
					fi
					fail "$rule: $relative_package reaches non-allowlisted dependency $dependency"
				done
				imports=''
				;;
			effect)
				imports=''
				dependencies=$($go_bin list -deps -f '{{.ImportPath}}' "$package")
				for dependency in $dependencies; do
					if matches_prefix "$dependency" "$module"; then
						direct_imports=$($go_bin list -f '{{join .Imports " "}}' "$dependency")
						imports="$imports $direct_imports"
					elif test "$($go_bin list -f '{{.Standard}}' "$dependency")" = 'false'; then
						fail "BND-006: $relative_package reaches external module $dependency through a pure package closure"
					fi
				done
				;;
			transitive) imports=$($go_bin list -deps -f '{{.ImportPath}}' "$package") ;;
		esac
		for imported in $imports; do
			case $forbidden_prefix in
				os|net|time|math/rand|crypto/rand|runtime|syscall) forbidden_path=$forbidden_prefix ;;
				*) forbidden_path="$module/$forbidden_prefix" ;;
			esac
			if matches_prefix "$imported" "$forbidden_path"; then
				fail "$rule: $relative_package imports forbidden $imported ($mode)"
			fi
		done
	done
done <"$policy_file"

harness_path="$module/internal/harness"
for package in $packages; do
	case $package in "$module"/*) ;; *) continue ;; esac
	matches_prefix "$package" "$harness_path" && continue
	imports=$($go_bin list -f '{{join .Imports " "}} {{join .TestImports " "}} {{join .XTestImports " "}}' "$package")
	for imported in $imports; do
		if matches_prefix "$imported" "$harness_path"; then
			fail "BND-005: ${package#"$module/"} imports test/build-control-only harness package $imported"
		fi
		test "$imported" != 'unsafe' || fail "BND-007: ${package#"$module/"} imports unsafe"
	done
done

printf 'check-import-boundaries: PASS (%s package set)\n' "$(printf '%s\n' "$packages" | awk 'NF { count++ } END { print count + 0 }')"
