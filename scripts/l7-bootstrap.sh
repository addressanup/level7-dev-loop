#!/bin/sh
# Level 7 one-time bootstrap: install a verified `l7` CLI into a user-owned
# prefix, then let that CLI install the plugin into the selected hosts.
#
# Review this file before running it. It needs curl, unzip, jq, shasum or
# sha256sum, and the GitHub CLI (gh >= 2.67.0, no login). It never uses sudo,
# never evaluates remote content, never edits shell startup files, never
# touches Gatekeeper or quarantine, and downloads only from the fixed
# repository addressanup/level7-dev-loop over HTTPS.
#
# Usage:
#   sh l7-bootstrap.sh --host codex [--host claude] [--version X.Y.Z]
#       [--channel stable|prerelease] [--accept-unsigned] [--accept-unqualified]
#       [--dry-run]
#
# Environment:
#   L7_HOME     installer home (default ~/.local/share/level7)
#   L7_BIN_DIR  directory for the `l7` link (default ~/.local/bin)

set -eu

repository='addressanup/level7-dev-loop'
api='https://api.github.com'
download='https://github.com'
source_ref='refs/heads/main'
predicate='https://slsa.dev/provenance/v1'
minimum_gh='2.67.0'

fail()
{
	printf 'l7-bootstrap: %s\n' "$*" >&2
	exit 1
}

hosts=''
version=''
channel='stable'
accept_unsigned=0
accept_unqualified=0
dry_run=0
while [ "$#" -gt 0 ]; do
	case $1 in
		--host)
			[ "$#" -ge 2 ] || fail '--host needs codex or claude'
			case $2 in codex|claude) ;; *) fail "unknown host $2 (codex or claude)" ;; esac
			hosts="$hosts $2"
			shift 2
			;;
		--version)
			[ "$#" -ge 2 ] || fail '--version needs an exact version such as 1.0.0'
			version=$2
			shift 2
			;;
		--channel)
			[ "$#" -ge 2 ] || fail '--channel needs stable or prerelease'
			case $2 in stable|prerelease) ;; *) fail "unknown channel $2" ;; esac
			channel=$2
			shift 2
			;;
		--accept-unsigned) accept_unsigned=1; shift ;;
		--accept-unqualified) accept_unqualified=1; shift ;;
		--dry-run) dry_run=1; shift ;;
		-h|--help) sed -n '2,19p' "$0"; exit 0 ;;
		*) fail "unknown option $1" ;;
	esac
done
[ -n "$hosts" ] || fail 'pass at least one --host codex or --host claude'
if [ "$channel" = prerelease ] && [ -z "$version" ]; then
	fail 'the prerelease channel requires an explicit --version'
fi
case $version in
	''|[0-9]*.[0-9]*.[0-9]*) ;;
	*) fail "--version must be MAJOR.MINOR.PATCH[-prerelease], got $version" ;;
esac

for tool in curl unzip jq gh; do
	command -v "$tool" >/dev/null 2>&1 || fail "$tool is required; install it first (gh: https://cli.github.com, no GitHub login is needed)"
done
if command -v shasum >/dev/null 2>&1; then
	sha256() { shasum -a 256 "$1" | awk '{ print $1 }'; }
elif command -v sha256sum >/dev/null 2>&1; then
	sha256() { sha256sum "$1" | awk '{ print $1 }'; }
else
	fail 'shasum or sha256sum is required'
fi
gh_version=$(gh --version 2>/dev/null | awk 'NR == 1 { print $3 }')
case $gh_version in
	[0-9]*.[0-9]*.[0-9]*) ;;
	*) fail 'gh --version did not report a version; repair the GitHub CLI installation' ;;
esac
newest=$(printf '%s\n%s\n' "$minimum_gh" "$gh_version" | sort -t. -k1,1n -k2,2n -k3,3n | tail -n 1)
[ "$newest" = "$gh_version" ] || fail "gh $gh_version is older than $minimum_gh; update the GitHub CLI"

[ "$(uname -s)" = Darwin ] || fail 'Level 7 v1 packages target macOS only'
case $(uname -m) in
	arm64|aarch64) arch=arm64 ;;
	x86_64|amd64) arch=amd64 ;;
	*) fail "unsupported architecture $(uname -m)" ;;
esac

L7_HOME=${L7_HOME:-$HOME/.local/share/level7}
L7_BIN_DIR=${L7_BIN_DIR:-$HOME/.local/bin}
case $L7_HOME in /*) ;; *) fail 'L7_HOME must be an absolute path' ;; esac
case $L7_BIN_DIR in /*) ;; *) fail 'L7_BIN_DIR must be an absolute path' ;; esac

work=$(mktemp -d "${TMPDIR:-/tmp}/l7-bootstrap.XXXXXX")
cleanup() { rm -rf -- "$work"; }
trap cleanup EXIT
trap 'exit 130' HUP INT TERM

fetch()
{
	# fetch URL FILE: bounded HTTPS download with no credentials; the final
	# URL of the same request must stay on GitHub's own hosts.
	final=$(curl --fail --silent --show-error --location --max-redirs 5 --proto '=https' --proto-redir '=https' \
		--max-filesize 268435456 --connect-timeout 30 --max-time 900 \
		--header 'Accept: application/vnd.github+json' --header 'X-GitHub-Api-Version: 2022-11-28' \
		--header 'User-Agent: level7-dev-loop-bootstrap' \
		--output "$2" --write-out '%{url_effective}' "$1") || fail "download failed: $1"
	case $final in
		https://github.com/*|https://api.github.com/*|https://objects.githubusercontent.com/*|https://release-assets.githubusercontent.com/*) ;;
		*) rm -f -- "$2"; fail "download left GitHub's hosts: $final" ;;
	esac
}

attest()
{
	# attest FILE SUBJECT COMMIT WORKFLOW: fetch the GitHub attestation bundles
	# for the file's digest and verify them offline with gh.
	digest=$(sha256 "$1")
	fetch "$api/repos/$repository/attestations/sha256:$digest" "$work/$2.attestations.json"
	jq -c '.attestations[].bundle' "$work/$2.attestations.json" >"$work/$2.bundles.jsonl"
	[ -s "$work/$2.bundles.jsonl" ] || fail "no GitHub attestation exists for $2 (digest $digest)"
	env -u GH_TOKEN -u GITHUB_TOKEN GH_PROMPT_DISABLED=1 GH_NO_UPDATE_NOTIFIER=1 GH_PAGER=cat \
		gh attestation verify "$1" --bundle "$work/$2.bundles.jsonl" --repo "$repository" \
		--signer-workflow "$repository/$4" --source-ref "$source_ref" --source-digest "$3" \
		--predicate-type "$predicate" --deny-self-hosted-runners --format json >"$work/$2.verify.json" \
		|| fail "GitHub attestation verification failed for $2"
	jq -e --arg digest "$digest" --arg name "$2" --arg commit "$3" --arg repo "https://github.com/$repository" \
		--arg signer "https://github.com/$repository/$4@$source_ref" \
		'map(select(.verificationResult.signature.certificate.sourceRepositoryDigest == $commit
			and .verificationResult.signature.certificate.sourceRepositoryURI == $repo
			and .verificationResult.signature.certificate.sourceRepositoryRef == "refs/heads/main"
			and .verificationResult.signature.certificate.buildSignerURI == $signer
			and .verificationResult.signature.certificate.runnerEnvironment == "github-hosted"
			and (.verificationResult.statement.subject | any(.name == $name and .digest.sha256 == $digest)))) | length > 0' \
		"$work/$2.verify.json" >/dev/null || fail "attestation for $2 does not carry the trusted identity"
}

# 1. Resolve the release.
if [ -n "$version" ]; then
	fetch "$api/repos/$repository/releases/tags/v$version" "$work/release.json"
else
	fetch "$api/repos/$repository/releases/latest" "$work/release.json"
fi
tag=$(jq -er '.tag_name' "$work/release.json") || fail 'release record is malformed'
jq -e '.draft == false' "$work/release.json" >/dev/null || fail "$tag is a draft"
prerelease=$(jq -r '.prerelease' "$work/release.json")
if [ "$prerelease" = true ] && [ "$channel" != prerelease ]; then
	fail "$tag is a prerelease; pass --channel prerelease --version ${tag#v} to select it explicitly"
fi
if [ "$prerelease" = true ]; then
	manifest_name='UNSIGNED-PRERELEASE-MANIFEST.json'
	workflow='.github/workflows/unsigned-prerelease.yml'
else
	manifest_name='RELEASE-MANIFEST.json'
	workflow='.github/workflows/release.yml'
fi
jq -e --arg name "$manifest_name" '.assets | map(.name) | index($name) != null' "$work/release.json" >/dev/null \
	|| fail "$tag does not attach $manifest_name"

# 2. Verify the manifest, then everything the manifest binds.
fetch "$download/$repository/releases/download/$tag/$manifest_name" "$work/$manifest_name"
commit=$(jq -er '.candidate.commit' "$work/$manifest_name") || fail 'manifest lacks a source commit'
case $commit in [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]*) ;; *) fail 'manifest source commit is invalid' ;; esac
[ "$(jq -r '.tag' "$work/$manifest_name")" = "$tag" ] || fail 'manifest tag does not match the release'
[ "$(jq -r '.workflow.repository' "$work/$manifest_name")" = "$repository" ] || fail 'manifest names another repository'
if [ "$prerelease" != true ]; then
	[ "$(jq -r '.workflow.path' "$work/$manifest_name")" = "$workflow" ] || fail 'manifest names another workflow'
fi
release_version=$(jq -er '.version' "$work/$manifest_name") || fail 'manifest lacks a version'
attest "$work/$manifest_name" "$manifest_name" "$commit" "$workflow"

if [ "$prerelease" = true ]; then
	signing=$(jq -r '.developer_id_signing' "$work/$manifest_name")
	support=$(jq -r '.support' "$work/$manifest_name")
	input_digest=$(jq -er --arg path "darwin-$arch/l7" '.unsigned_inputs[] | select(.path == $path) | .sha256' "$work/$manifest_name") || fail 'manifest lacks the CLI input digest'
else
	signing=$(jq -r '.qualification.developer_id_signing' "$work/$manifest_name")
	support=$(jq -r '.qualification.formal_support' "$work/$manifest_name")
	input_digest=$(jq -er --arg path "darwin-$arch/l7" '.unsigned_inputs[$path]' "$work/$manifest_name") || fail 'manifest lacks the CLI input digest'
fi
if [ "$signing" != PASS ] && [ "$accept_unsigned" != 1 ]; then
	fail "$tag binaries are UNSIGNED and NOT NOTARIZED (developer_id_signing=$signing); macOS Gatekeeper may block them. Read the release notes, then rerun with --accept-unsigned. Never disable Gatekeeper globally."
fi
if [ "$support" != SUPPORTED ] && [ "$accept_unqualified" != 1 ]; then
	fail "$tag formal support is $support and release qualification is NOT_RUN. Read the release notes, then rerun with --accept-unqualified."
fi

set -- $hosts
first_host=$1
archive_name="level7-dev-loop-$release_version-$first_host.zip"
archive_digest=$(jq -er --arg name "$archive_name" '.artifacts[] | select(.name == $name) | .sha256' "$work/$manifest_name") || fail "manifest lacks $archive_name"
checksum_digest=$(jq -er '.checksums.sha256' "$work/$manifest_name") || fail 'manifest lacks the checksum digest'

fetch "$download/$repository/releases/download/$tag/SHA256SUMS" "$work/SHA256SUMS"
[ "$(sha256 "$work/SHA256SUMS")" = "$checksum_digest" ] || fail 'SHA256SUMS does not match the attested manifest'
attest "$work/SHA256SUMS" SHA256SUMS "$commit" "$workflow"
grep -Fqx "$archive_digest  $archive_name" "$work/SHA256SUMS" || fail "SHA256SUMS does not list $archive_name with the manifest digest"

if [ "$dry_run" = 1 ]; then
	printf 'l7-bootstrap: dry run; would install l7 %s (%s, commit %s) into %s and link %s/l7, then install the plugin for:%s\n' \
		"$release_version" "$tag" "$commit" "$L7_HOME" "$L7_BIN_DIR" "$hosts"
	exit 0
fi

fetch "$download/$repository/releases/download/$tag/$archive_name" "$work/$archive_name"
[ "$(sha256 "$work/$archive_name")" = "$archive_digest" ] || fail "$archive_name does not match the attested digest"
attest "$work/$archive_name" "$archive_name" "$commit" "$workflow"

# 3. Extract only the CLI for this architecture and bind it to the manifest.
unzip -p "$work/$archive_name" "bin/darwin-$arch/l7" >"$work/l7" || fail 'cannot extract the CLI from the archive'
[ "$(sha256 "$work/l7")" = "$input_digest" ] || fail 'extracted CLI does not match the attested release input'
chmod 0755 "$work/l7"

# 4. Place the CLI without touching any running binary; link atomically.
cli_dir="$L7_HOME/cli/$release_version"
mkdir -p "$L7_HOME/cli" "$L7_HOME/plugin/cache/$tag" "$L7_BIN_DIR"
if [ -e "$cli_dir/l7" ] && [ "$(sha256 "$cli_dir/l7")" = "$input_digest" ]; then
	printf 'l7-bootstrap: l7 %s is already installed at %s\n' "$release_version" "$cli_dir/l7"
else
	staged=$(mktemp -d "$L7_HOME/cli/.staging-XXXXXX")
	cp "$work/l7" "$staged/l7"
	chmod 0755 "$staged/l7"
	printf '{\n  "schema": 1,\n  "version": "%s",\n  "arch": "%s",\n  "sha256": "%s",\n  "source_commit": "%s"\n}\n' \
		"$release_version" "$arch" "$input_digest" "$commit" >"$staged/receipt.json"
	if [ -e "$cli_dir" ]; then
		rm -rf -- "$staged"
		fail "$cli_dir exists but does not hold the attested CLI; move it away and rerun"
	fi
	mv "$staged" "$cli_dir"
fi
printf '%s\n' "$L7_BIN_DIR/l7" >"$L7_HOME/cli/link"
if [ -e "$L7_BIN_DIR/l7" ] && [ ! -L "$L7_BIN_DIR/l7" ]; then
	fail "$L7_BIN_DIR/l7 exists and is not a Level 7 link; refusing to replace it (set L7_BIN_DIR to another directory)"
fi
temporary_link="$L7_BIN_DIR/.l7-link-$$"
ln -s "$cli_dir/l7" "$temporary_link"
mv -f "$temporary_link" "$L7_BIN_DIR/l7"
if [ ! -e "$L7_HOME/plugin/cache/$tag/$archive_name" ]; then
	mv "$work/$archive_name" "$L7_HOME/plugin/cache/$tag/$archive_name"
fi

"$cli_dir/l7" --version >/dev/null || fail 'the installed l7 did not run; if macOS Gatekeeper blocked it, stop here and never disable Gatekeeper globally'
printf 'l7-bootstrap: installed l7 %s (%s, source commit %s) at %s\n' "$release_version" "$tag" "$commit" "$cli_dir/l7"

# 5. Install the plugin through the same verified CLI. An explicit --version
# pins the plugin; otherwise the CLI follows the channel so `update --all`
# keeps working.
install_flags=''
[ -n "$version" ] && install_flags="--version $version"
[ "$channel" = prerelease ] && install_flags="$install_flags --channel prerelease"
[ "$accept_unsigned" = 1 ] && install_flags="$install_flags --accept-unsigned"
[ "$accept_unqualified" = 1 ] && install_flags="$install_flags --accept-unqualified"
status=0
for host in "$@"; do
	# shellcheck disable=SC2086
	L7_HOME="$L7_HOME" "$cli_dir/l7" plugin install --host "$host" $install_flags || status=$?
done

case ":$PATH:" in
	*":$L7_BIN_DIR:"*) ;;
	*) printf 'l7-bootstrap: add %s to your PATH to run l7 directly (this script does not edit shell startup files)\n' "$L7_BIN_DIR" ;;
esac
printf 'l7-bootstrap: update later with: l7 plugin update --all --accept-unsigned --accept-unqualified\n'
exit "$status"
