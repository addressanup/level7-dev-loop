#!/bin/sh
# Offline check of scripts/l7-bootstrap.sh: a fake curl serves fixture release
# assets, a fake gh verifies attestation bundles, and the extracted fixture
# l7 records how the bootstrap invoked it. No network, no real host.

set -eu
umask 077

fail()
{
	printf 'check-l7-bootstrap: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 0 || fail 'usage: check-l7-bootstrap.sh'
for tool in jq zip unzip shasum; do
	command -v "$tool" >/dev/null 2>&1 || command -v sha256sum >/dev/null 2>&1 || fail "$tool is required for the bootstrap check"
done
if command -v shasum >/dev/null 2>&1; then
	sha256() { shasum -a 256 "$1" | awk '{ print $1 }'; }
else
	sha256() { sha256sum "$1" | awk '{ print $1 }'; }
fi

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
bootstrap="$project_root/scripts/l7-bootstrap.sh"
test -f "$bootstrap" || fail 'bootstrap script is missing'
sh -n "$bootstrap" || fail 'bootstrap script has a syntax error'

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=$(CDPATH='' cd "${temporary_parent%/}" && pwd -P)
scratch=$(mktemp -d "$temporary_parent/l7-bootstrap-check.XXXXXX")
case $scratch in "$temporary_parent"/l7-bootstrap-check.*) ;; *) fail 'mktemp returned an unsafe directory' ;; esac
cleanup()
{
	case $scratch in "$temporary_parent"/l7-bootstrap-check.*) rm -rf -- "$scratch" ;; *) return 1 ;; esac
}
trap cleanup EXIT HUP INT TERM

arch=arm64
machine=arm64
case $(uname -m) in x86_64|amd64) arch=amd64; machine=x86_64 ;; esac
version=1.0.0
tag=v$version
commit=ed1457f2c54e2308f757d7ea3bf4fcfa113575ec
fixtures="$scratch/fixtures"
mkdir -p "$fixtures/package/bin/darwin-$arch" "$fixtures/package/bin/darwin-other" "$scratch/bin" "$scratch/home"

# The packaged l7 is a script here: the bootstrap binds it by digest, runs
# --version, and delegates the plugin installation to it.
cat >"$fixtures/package/bin/darwin-$arch/l7" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >>"${L7_HOME:?}/l7-invocations.log"
case $1 in
	--version) printf 'PASS version=1.0.0\n' ;;
	plugin) printf 'PASS plugin %s\n' "$*" ;;
esac
EOF
chmod 0755 "$fixtures/package/bin/darwin-$arch/l7"
printf 'other arch\n' >"$fixtures/package/bin/darwin-other/l7"
(cd "$fixtures/package" && zip -q -X -r "$fixtures/level7-dev-loop-$version-codex.zip" bin)
cp "$fixtures/level7-dev-loop-$version-codex.zip" "$fixtures/level7-dev-loop-$version-claude.zip"
input_digest=$(sha256 "$fixtures/package/bin/darwin-$arch/l7")
codex_digest=$(sha256 "$fixtures/level7-dev-loop-$version-codex.zip")
claude_digest=$(sha256 "$fixtures/level7-dev-loop-$version-claude.zip")
printf '%s  level7-dev-loop-%s-claude.zip\n%s  level7-dev-loop-%s-codex.zip\n' "$claude_digest" "$version" "$codex_digest" "$version" >"$fixtures/SHA256SUMS"
sums_digest=$(sha256 "$fixtures/SHA256SUMS")
sums_size=$(wc -c <"$fixtures/SHA256SUMS" | tr -d ' ')
jq -n --arg version "$version" --arg tag "$tag" --arg commit "$commit" --arg arch "darwin-$arch/l7" --arg input "$input_digest" \
	--arg codex "$codex_digest" --arg claude "$claude_digest" --arg sums "$sums_digest" --argjson sums_size "$sums_size" '
	{schema: 1, version: $version, tag: $tag, release_channel: "stable", policy: "owner-authorized-unsigned-stable-v1",
	 candidate: {commit: $commit, tree: $commit}, workflow: {repository: "addressanup/level7-dev-loop", path: ".github/workflows/release.yml", run_id: "1"},
	 qualification: {developer_id_signing: "NOT_RUN", apple_notarization: "NOT_RUN", formal_support: "WITHHELD"},
	 unsigned_inputs: {($arch): $input},
	 artifacts: [{name: ("level7-dev-loop-" + $version + "-claude.zip"), sha256: $claude}, {name: ("level7-dev-loop-" + $version + "-codex.zip"), sha256: $codex}],
	 checksums: {name: "SHA256SUMS", sha256: $sums, size: $sums_size}}' >"$fixtures/RELEASE-MANIFEST.json"
jq -n --arg tag "$tag" '{tag_name: $tag, draft: false, prerelease: false, assets: [{name: "RELEASE-MANIFEST.json"}, {name: "SHA256SUMS"}, {name: "level7-dev-loop-1.0.0-codex.zip"}, {name: "level7-dev-loop-1.0.0-claude.zip"}]}' >"$fixtures/release.json"

# Fake uname: the bootstrap targets macOS, and this check also runs on Linux
# CI, so the fixture PATH reports Darwin with the real machine architecture.
cat >"$scratch/bin/uname" <<UNAME
#!/bin/sh
case "\${1:-}" in
	-m) printf '%s\n' '$machine' ;;
	*) printf 'Darwin\n' ;;
esac
UNAME
chmod 0755 "$scratch/bin/uname"

# Fake curl: serves fixture assets by URL path and reports the effective URL
# the way the real bootstrap expects. CONTROL=missing-attestation drops the
# attestation lookups; CONTROL=foreign-redirect reports a foreign final URL.
cat >"$scratch/bin/curl" <<EOF
#!/bin/sh
output=''
url=''
while [ "\$#" -gt 0 ]; do
	case \$1 in
		--output) output=\$2; shift 2 ;;
		--write-out|--header|--max-redirs|--proto|--proto-redir|--max-filesize|--connect-timeout|--max-time) shift 2 ;;
		-*) shift ;;
		*) url=\$1; shift ;;
	esac
done
printf '%s\n' "\$url" >>'$scratch/curl.log'
control=\$(cat '$scratch/control' 2>/dev/null || true)
case \$url in
	https://api.github.com/repos/addressanup/level7-dev-loop/releases/latest|https://api.github.com/repos/addressanup/level7-dev-loop/releases/tags/$tag)
		cp '$fixtures/release.json' "\$output" ;;
	https://api.github.com/repos/addressanup/level7-dev-loop/attestations/sha256:*)
		if [ "\$control" = missing-attestation ]; then printf '{"attestations":[]}' >"\$output"; else
		digest=\${url##*sha256:}
		printf '{"attestations":[{"bundle":{"l7test":{"sha256":"%s","commit":"$commit"}}}]}' "\$digest" >"\$output"; fi ;;
	https://github.com/addressanup/level7-dev-loop/releases/download/$tag/*)
		asset=\${url##*/}
		test -f '$fixtures'/"\$asset" || exit 22
		cp '$fixtures'/"\$asset" "\$output" ;;
	*) exit 22 ;;
esac
if [ "\$control" = foreign-redirect ]; then printf 'https://evil.example/%s' "\${url##*/}"; else printf '%s' "\$url"; fi
EOF
chmod 0755 "$scratch/bin/curl"

# Fake gh: enforces the subject digest and --source-digest from the bundle,
# then prints the certificate shape the bootstrap re-checks with jq.
cat >"$scratch/bin/gh" <<EOF
#!/bin/sh
if [ "\$1" = --version ]; then printf 'gh version 2.101.0 (2026-09-15)\n'; exit 0; fi
test "\$1" = attestation && test "\$2" = verify || exit 2
file=\$3
shift 3
bundle=''
source_digest=''
while [ "\$#" -gt 0 ]; do
	case \$1 in
		--bundle) bundle=\$2; shift 2 ;;
		--source-digest) source_digest=\$2; shift 2 ;;
		--repo|--signer-workflow|--source-ref|--predicate-type|--format) shift 2 ;;
		*) shift ;;
	esac
done
test -n "\$GH_TOKEN" && { printf 'token must not be forwarded\n' >&2; exit 9; }
if command -v shasum >/dev/null 2>&1; then digest=\$(shasum -a 256 "\$file" | awk '{ print \$1 }'); else digest=\$(sha256sum "\$file" | awk '{ print \$1 }'); fi
expected=\$(jq -r '.l7test.sha256' "\$bundle" | head -n 1)
expected_commit=\$(jq -r '.l7test.commit' "\$bundle" | head -n 1)
test "\$digest" = "\$expected" && test "\$source_digest" = "\$expected_commit" || { printf 'verification failed\n' >&2; exit 1; }
jq -n --arg digest "\$digest" --arg name "\$(basename "\$file")" --arg commit "\$expected_commit" '[{verificationResult: {signature: {certificate: {
	sourceRepositoryDigest: \$commit, sourceRepositoryURI: "https://github.com/addressanup/level7-dev-loop", sourceRepositoryRef: "refs/heads/main",
	buildSignerURI: "https://github.com/addressanup/level7-dev-loop/.github/workflows/release.yml@refs/heads/main", runnerEnvironment: "github-hosted"}},
	statement: {predicateType: "https://slsa.dev/provenance/v1", subject: [{name: \$name, digest: {sha256: \$digest}}]}}}]'
EOF
chmod 0755 "$scratch/bin/gh"

run_bootstrap()
{
	# run_bootstrap HOME EXPECTED_STATUS ARGS...: run with a disposable home
	# and the fake tools first on PATH.
	home=$1
	expected=$2
	shift 2
	mkdir -p "$home"
	status=0
	HOME="$home" L7_HOME="$home/level7" L7_BIN_DIR="$home/bin" GH_TOKEN=must-not-cross \
		PATH="$scratch/bin:/usr/bin:/bin" TMPDIR="$scratch" \
		sh "$bootstrap" "$@" >"$home/output" 2>&1 || status=$?
	test "$status" -eq "$expected" || { cat "$home/output" >&2; fail "bootstrap exited $status, expected $expected ($*)"; }
}

# Consent is required before anything is placed.
run_bootstrap "$scratch/home/consent" 1 --host codex
grep -Fq -- '--accept-unsigned' "$scratch/home/consent/output" || fail 'missing-consent error does not name --accept-unsigned'
test ! -e "$scratch/home/consent/level7/cli" || fail 'refused bootstrap placed a CLI'
run_bootstrap "$scratch/home/consent2" 1 --host codex --accept-unsigned
grep -Fq -- '--accept-unqualified' "$scratch/home/consent2/output" || fail 'missing-consent error does not name --accept-unqualified'

# Dry run resolves and verifies but changes nothing.
run_bootstrap "$scratch/home/dry" 0 --host codex --accept-unsigned --accept-unqualified --dry-run
grep -Fq 'dry run; would install l7 1.0.0' "$scratch/home/dry/output" || fail 'dry run did not describe its plan'
test ! -e "$scratch/home/dry/level7/cli" || fail 'dry run placed a CLI'

# Success: verified CLI, receipt, link, cache, and delegated plugin installs.
run_bootstrap "$scratch/home/ok" 0 --host codex --host claude --accept-unsigned --accept-unqualified
cli="$scratch/home/ok/level7/cli/$version/l7"
test -x "$cli" || fail 'CLI was not installed'
test "$(sha256 "$cli")" = "$input_digest" || fail 'installed CLI does not match the attested input'
jq -e --arg version "$version" --arg commit "$commit" --arg digest "$input_digest" '.schema == 1 and .version == $version and .source_commit == $commit and .sha256 == $digest' "$scratch/home/ok/level7/cli/$version/receipt.json" >/dev/null || fail 'CLI receipt is wrong'
test "$(cat "$scratch/home/ok/level7/cli/link")" = "$scratch/home/ok/bin/l7" || fail 'link record is wrong'
test "$(readlink "$scratch/home/ok/bin/l7")" = "$cli" || fail 'l7 link does not point at the installed CLI'
test -f "$scratch/home/ok/level7/plugin/cache/$tag/level7-dev-loop-$version-codex.zip" || fail 'verified archive was not cached for the CLI'
grep -Fqx 'plugin install --host codex --accept-unsigned --accept-unqualified' "$scratch/home/ok/level7/l7-invocations.log" || fail 'codex plugin install was not delegated with the consent flags'
grep -Fqx 'plugin install --host claude --accept-unsigned --accept-unqualified' "$scratch/home/ok/level7/l7-invocations.log" || fail 'claude plugin install was not delegated'
if grep 'plugin install' "$scratch/home/ok/level7/l7-invocations.log" | grep -Fq -- '--version'; then fail 'bootstrap pinned the plugin without an explicit --version'; fi
grep -Fq 'add '"$scratch/home/ok/bin"' to your PATH' "$scratch/home/ok/output" || fail 'PATH advice missing'
grep -Fq 'l7 plugin update --all' "$scratch/home/ok/output" || fail 'update recipe missing'
grep -q 'attestations/sha256:' "$scratch/curl.log" || fail 'attestation bundles were never fetched'
test "$(grep -c 'attestations/sha256:' "$scratch/curl.log")" -ge 3 || fail 'manifest, checksums, and archive were not each attested'

# Re-running is idempotent and an explicit version pins the plugin.
run_bootstrap "$scratch/home/ok" 0 --host codex --version 1.0.0 --accept-unsigned --accept-unqualified
grep -Fq 'already installed' "$scratch/home/ok/output" || fail 'rerun did not reuse the installed CLI'
grep -Fqx 'plugin install --host codex --version 1.0.0 --accept-unsigned --accept-unqualified' "$scratch/home/ok/level7/l7-invocations.log" || fail 'explicit version was not forwarded as a pin'

# A foreign file at the link path is never replaced.
mkdir -p "$scratch/home/foreign/bin"
printf '#!/bin/sh\necho mine\n' >"$scratch/home/foreign/bin/l7"
run_bootstrap "$scratch/home/foreign" 1 --host codex --accept-unsigned --accept-unqualified
grep -Fq 'refusing to replace' "$scratch/home/foreign/output" || fail 'foreign link file was not refused'
grep -Fq 'echo mine' "$scratch/home/foreign/bin/l7" || fail 'foreign link file was replaced'

# Tampered archive bytes fail before extraction.
cp "$fixtures/level7-dev-loop-$version-codex.zip" "$scratch/codex.zip.orig"
printf 'x' >>"$fixtures/level7-dev-loop-$version-codex.zip"
run_bootstrap "$scratch/home/tamper" 1 --host codex --accept-unsigned --accept-unqualified
grep -Fq 'does not match the attested digest' "$scratch/home/tamper/output" || fail 'tampered archive was not detected'
test ! -e "$scratch/home/tamper/level7/cli" || fail 'tampered archive placed a CLI'
cp "$scratch/codex.zip.orig" "$fixtures/level7-dev-loop-$version-codex.zip"

# Missing attestations and foreign redirects fail closed.
printf 'missing-attestation' >"$scratch/control"
run_bootstrap "$scratch/home/noattest" 1 --host codex --accept-unsigned --accept-unqualified
grep -Fq 'no GitHub attestation exists' "$scratch/home/noattest/output" || fail 'missing attestation was not detected'
printf 'foreign-redirect' >"$scratch/control"
run_bootstrap "$scratch/home/redirect" 1 --host codex --accept-unsigned --accept-unqualified
grep -Fq "left GitHub's hosts" "$scratch/home/redirect/output" || fail 'foreign redirect was not detected'
rm -f "$scratch/control"

# Prereleases need an explicit version; unknown options are rejected.
run_bootstrap "$scratch/home/pre" 1 --host codex --channel prerelease --accept-unsigned --accept-unqualified
grep -Fq 'requires an explicit --version' "$scratch/home/pre/output" || fail 'prerelease without version accepted'
run_bootstrap "$scratch/home/bad" 1 --host browser
run_bootstrap "$scratch/home/bad2" 1 --bogus
run_bootstrap "$scratch/home/bad3" 1

printf 'check-l7-bootstrap: PASS (fake curl/gh; consent, dry run, install, rerun, pin, foreign link, tamper, missing attestation, foreign redirect)\n'
