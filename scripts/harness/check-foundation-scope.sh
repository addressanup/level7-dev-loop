#!/bin/sh

set -eu

fail()
{
	printf 'check-foundation-scope: %s\n' "$*" >&2
	exit 1
}

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
cd "$project_root"

manifest='harness/foundation-inputs.sha256'
test -f "$manifest" || fail "missing digest manifest: $manifest"

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

while read -r expected path extra; do
	case $expected in ''|'#'*) continue ;; esac
	test -z "${extra:-}" || fail "malformed digest record for $path"
	test -f "$path" || fail "protected input is missing: $path"
	actual=$(sha256_file "$path")
	test "$actual" = "$expected" || fail "protected input changed: $path"
done <"$manifest"

for path in \
	cmd/l7 \
	cmd/l7up \
	internal/supervisor \
	internal/kernel \
	internal/context \
	internal/artifact \
	internal/policy \
	internal/transaction \
	internal/executor \
	internal/receipt \
	internal/platform \
	internal/adapter \
	internal/channel \
	internal/render \
	internal/evaluator \
	semantic \
	schemas \
	fixtures \
	packages \
	build/generated
do
	test ! -e "$path" || fail "product path exists during inert Foundation Step 5: $path"
done

test ! -e go.sum || fail 'go.sum is unexpected while production module count is zero'
test ! -e vendor || fail 'vendor is unexpected while production module count is zero'
awk '/^[[:space:]]*require([[:space:]]|\()/ { found = 1 } END { exit found }' go.mod || fail 'production requirements are forbidden in the inert harness'

test_count=$(awk '/^func Test[A-Za-z0-9_]*\(/ { count++ } END { print count + 0 }' internal/harness/*_test.go)
test "$test_count" -eq 1 || fail "expected exactly one proving test, found $test_count"

baseline=$(sed -n '1p' .go-version)
test "$baseline" = '1.26.7' || fail '.go-version must pin baseline Go 1.26.7'
grep -Fqx 'toolchain go1.26.7' go.mod || fail 'go.mod toolchain pin changed'
grep -Fqx 'module continuallabs.ltd/level7-dev-loop' go.mod || fail 'provisional root module identity changed'
grep -Fqx 'core	active	.	continuallabs.ltd/level7-dev-loop' harness/modules.lock.tsv || fail 'active core module registry changed'
grep -Fqx 'updater	reserved	cmd/l7up	UNSET' harness/modules.lock.tsv || fail 'updater must remain a reserved module until its boundary harness is approved'

toolchain_records=$(awk -F '\t' '$1 !~ /^#/ && NF { count++ } END { print count + 0 }' harness/toolchains.lock.tsv)
test "$toolchain_records" -eq 4 || fail "expected four platform toolchain records, found $toolchain_records"
awk -F '\t' '
	$1 !~ /^#/ && (NF != 9 || $7 !~ /^[0-9a-f]{64}$/ || $8 !~ /^https:\/\/go\.dev\/dl\// || $9 !~ /^https:\/\/go\.dev\/dl\/.*\.asc$/) { bad = 1 }
	END { exit bad }
' harness/toolchains.lock.tsv || fail 'malformed toolchain lock'
grep -Fq 'go-version: 1.26.7' .github/workflows/harness.yml || fail 'CI baseline Go version is missing'
grep -Fq 'go-version: 1.27.0' .github/workflows/harness.yml || fail 'CI shadow Go version is missing'
grep -Fq 'experimental: false' .github/workflows/harness.yml || fail 'CI baseline must remain blocking'
grep -Fq 'experimental: true' .github/workflows/harness.yml || fail 'CI shadow must remain nonblocking'
grep -Fq 'permissions:' .github/workflows/harness.yml || fail 'CI permissions block is missing'
grep -Fq 'contents: read' .github/workflows/harness.yml || fail 'CI contents permission must remain read-only'
grep -Fq 'persist-credentials: false' .github/workflows/harness.yml || fail 'CI checkout credentials must not persist'
if grep -Fq 'pull_request_target:' .github/workflows/harness.yml; then fail 'pull_request_target is forbidden'; fi
if grep -Fq 'secrets.' .github/workflows/harness.yml; then fail 'CI harness must not consume secrets'; fi
awk '
	$1 == "uses:" {
		split($2, parts, "@")
		if (parts[2] !~ /^[0-9a-f]{40}$/) bad = 1
	}
	END { exit bad }
' .github/workflows/harness.yml || fail 'every CI action must use a full commit digest'

identity_records=$(awk -F '\t' '$1 == "go-archive" { count++ } END { print count + 0 }' harness/signing-identities.lock.tsv)
test "$identity_records" -eq 1 || fail 'expected one go-archive signing identity record'
grep -Fq 'EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796' harness/signing-identities.lock.tsv || fail 'Go signing primary fingerprint is missing'
grep -Fq '0E225917414670F4442C250DFD533C07C264648F' harness/signing-identities.lock.tsv || fail 'Go signing subkey fingerprint is missing'

checkout_sha=$(awk -F '\t' '$1 == "actions/checkout" && $2 == "v7.0.1" { print $3 }' harness/ci-actions.lock.tsv)
test "$checkout_sha" = '3d3c42e5aac5ba805825da76410c181273ba90b1' || fail 'checkout action lock changed'
grep -Fq "uses: actions/checkout@$checkout_sha" .github/workflows/harness.yml || fail 'CI checkout action is not digest-pinned to the lock'

printf 'check-foundation-scope: PASS\n'
