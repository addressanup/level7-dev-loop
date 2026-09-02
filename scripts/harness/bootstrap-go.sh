#!/bin/sh

set -eu
umask 077

fail()
{
	printf 'bootstrap-go: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 1 || fail 'usage: bootstrap-go.sh VERSION'
version=$1

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
lock_file="$project_root/harness/toolchains.lock.tsv"
identity_lock_file="$project_root/harness/signing-identities.lock.tsv"

case $(uname -s) in
	Darwin) goos=darwin ;;
	Linux) goos=linux ;;
	*) fail "unsupported bootstrap operating system: $(uname -s)" ;;
esac

case $(uname -m) in
	arm64|aarch64) goarch=arm64 ;;
	x86_64|amd64) goarch=amd64 ;;
	*) fail "unsupported bootstrap architecture: $(uname -m)" ;;
esac

record=$(awk -F '\t' -v version="$version" -v goos="$goos" -v goarch="$goarch" '
	$1 !~ /^#/ && $2 == version && $3 == goos && $4 == goarch { print }
' "$lock_file")
record_count=$(printf '%s\n' "$record" | awk 'NF { count++ } END { print count + 0 }')
test "$record_count" -eq 1 || fail "lock must contain exactly one $version $goos/$goarch record"

tab=$(printf '\t')
IFS=$tab read -r role locked_version locked_goos locked_goarch filename expected_size expected_sha256 archive_url signature_url <<EOF
$record
EOF
test -n "${signature_url:-}" || fail 'malformed toolchain lock record'

test "$locked_version" = "$version" || fail 'version lock mismatch'
test "$locked_goos" = "$goos" || fail 'operating-system lock mismatch'
test "$locked_goarch" = "$goarch" || fail 'architecture lock mismatch'
case $expected_sha256 in
	*[!0-9a-f]*|'') fail 'malformed SHA-256 lock' ;;
esac
test "${#expected_sha256}" -eq 64 || fail 'malformed SHA-256 length'
case $archive_url in https://go.dev/dl/*) ;; *) fail 'archive URL is not on go.dev' ;; esac
case $signature_url in https://go.dev/dl/*.asc) ;; *) fail 'signature URL is not on go.dev' ;; esac

identity_record=$(awk -F '\t' '$1 == "go-archive" { print }' "$identity_lock_file")
identity_count=$(printf '%s\n' "$identity_record" | awk 'NF { count++ } END { print count + 0 }')
test "$identity_count" -eq 1 || fail 'signing identity lock must contain exactly one go-archive record'
IFS=$tab read -r identity_purpose primary_fingerprint signing_fingerprint key_url identity_scope <<EOF
$identity_record
EOF
test "$identity_purpose" = 'go-archive' || fail 'signing identity purpose mismatch'
test "$identity_scope" = 'shared-google-linux-packages-authority' || fail 'unexpected signing identity scope'
test "$primary_fingerprint" = 'EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796' || fail 'primary fingerprint lock changed'
test "$signing_fingerprint" = '0E225917414670F4442C250DFD533C07C264648F' || fail 'signing fingerprint lock changed'
test "$key_url" = 'https://dl.google.com/linux/linux_signing_key.pub' || fail 'signing key URL lock changed'

cache_root="$project_root/.cache"
download_root="$cache_root/go-downloads"
toolchain_root="$cache_root/toolchains"
receipt_root="$cache_root/go-bootstrap-receipts"
telemetry_root="$cache_root/go-bootstrap-telemetry"
destination="$toolchain_root/go$version"
stage="$toolchain_root/.go$version.$goos-$goarch.staging"

for path in "$cache_root" "$download_root" "$toolchain_root" "$receipt_root" "$telemetry_root"; do
	test ! -L "$path" || fail "refusing symlinked cache path: $path"
done
mkdir -p "$download_root" "$toolchain_root" "$receipt_root" "$telemetry_root"
printf 'off 2026-08-24\n' >"$telemetry_root/mode"

receipt="$destination/.l7-bootstrap-receipt"
reuse_destination=0
if test -d "$destination"; then
	test ! -L "$destination" || fail "refusing symlinked toolchain: $destination"
	test -x "$destination/bin/go" || fail "incomplete existing toolchain: $destination"
	test -f "$receipt" || fail "existing toolchain has no bootstrap receipt: $destination"
	grep -Fqx "version=go$version" "$receipt" || fail 'existing toolchain version receipt mismatch'
	grep -Fqx "archive_sha256=$expected_sha256" "$receipt" || fail 'existing toolchain archive receipt mismatch'
	grep -Fqx "signing_primary_fingerprint=$primary_fingerprint" "$receipt" || fail 'existing toolchain primary-signing receipt mismatch'
	grep -Fqx "signing_subkey_fingerprint=$signing_fingerprint" "$receipt" || fail 'existing toolchain subkey-signing receipt mismatch'
	reuse_destination=1
else
	test ! -e "$destination" || fail "destination exists but is not a directory: $destination"
	test ! -e "$stage" || fail "staging path exists; inspect it before retrying: $stage"
fi

archive="$download_root/$filename"
signature="$download_root/$filename.asc"
public_key="$download_root/google-linux-signing-key.pub"
download_temporary=

cleanup_download()
{
	test -n "$download_temporary" || return 0
	case $download_temporary in
		"$download_root"/*.partial.*) rm -f -- "$download_temporary" ;;
		*) return 1 ;;
	esac
}

trap cleanup_download EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

download_failure()
{
	status=$1
	shift
	printf 'bootstrap-go: %s\n' "$*" >&2
	exit "$status"
}

download()
{
	url=$1
	output=$2
	test ! -L "$output" || fail "refusing symlinked download: $output"
	if test -e "$output"; then
		test -f "$output" || fail "existing download is not a regular file: $output"
		return
	fi

	download_started_at=$(date +%s)
	download_attempt=1
	while test "$download_attempt" -le 4; do
		download_now=$(date +%s)
		download_remaining=$((600 - (download_now - download_started_at)))
		test "$download_remaining" -gt 0 || download_failure 28 \
			"download aggregate deadline exceeded before attempt $download_attempt/4 (curl=28 http=000): $url"

		download_temporary=$(mktemp "$output.partial.XXXXXX") || fail "could not create download temporary file: $output"
		case $download_temporary in
			"$output".partial.*) ;;
			*) fail "mktemp returned an unexpected download path: $download_temporary" ;;
		esac
		test -f "$download_temporary" || fail "download temporary path is not a regular file: $download_temporary"
		test ! -L "$download_temporary" || fail "download temporary path is a symlink: $download_temporary"

		download_curl_status=0
		download_http_status=$(curl \
			--disable \
			--fail \
			--silent \
			--show-error \
			--location \
			--proto '=https' \
			--proto-redir '=https' \
			--tlsv1.2 \
			--connect-timeout 20 \
			--max-time "$download_remaining" \
			--write-out '%{http_code}' \
			--output "$download_temporary" \
			"$url") || download_curl_status=$?
		case $download_http_status in
			[0-9][0-9][0-9]) ;;
			*) download_http_status=000 ;;
		esac

		if test "$download_curl_status" -eq 0; then
			test ! -e "$output" || fail "download path appeared during transfer: $output"
			ln "$download_temporary" "$output" || fail "could not atomically install download: $output"
			rm -f -- "$download_temporary" || fail "could not remove installed download temporary file: $download_temporary"
			download_temporary=
			return
		fi

		download_retry=0
		case $download_curl_status in
			28|56) download_retry=1 ;;
			22)
				case $download_http_status in
					408|429|500|502|503|504) download_retry=1 ;;
				esac
			;;
		esac
		printf 'bootstrap-go: download attempt %s/4 failed (curl=%s http=%s): %s\n' \
			"$download_attempt" "$download_curl_status" "$download_http_status" "$url" >&2
		rm -f -- "$download_temporary" || fail "could not remove failed download temporary file: $download_temporary"
		download_temporary=

		if test "$download_retry" -ne 1; then
			download_failure "$download_curl_status" \
				"download failure is not retryable (curl=$download_curl_status http=$download_http_status): $url"
		fi
		if test "$download_attempt" -eq 4; then
			download_failure "$download_curl_status" \
				"download failed after 4 attempts (curl=$download_curl_status http=$download_http_status): $url"
		fi

		case $download_attempt in
			1) download_delay=1 ;;
			2) download_delay=2 ;;
			3) download_delay=4 ;;
			*) fail 'invalid download retry state' ;;
		esac
		download_now=$(date +%s)
		download_remaining=$((600 - (download_now - download_started_at)))
		test "$download_remaining" -gt "$download_delay" || download_failure 28 \
			"download aggregate deadline exceeded before retry delay (curl=28 http=$download_http_status): $url"
		printf 'bootstrap-go: retrying download in %s second(s)\n' "$download_delay" >&2
		sleep "$download_delay"
		download_attempt=$((download_attempt + 1))
	done
}

download "$archive_url" "$archive"
download "$signature_url" "$signature"
download "$key_url" "$public_key"

case $goos in
	darwin) actual_size=$(stat -f '%z' "$archive") ;;
	linux) actual_size=$(stat -c '%s' "$archive") ;;
esac
test "$actual_size" = "$expected_size" || fail "archive size mismatch: got $actual_size, want $expected_size"

if command -v sha256sum >/dev/null 2>&1; then
	actual_sha256=$(sha256sum "$archive" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
	actual_sha256=$(shasum -a 256 "$archive" | awk '{ print $1 }')
else
	fail 'sha256sum or shasum is required'
fi
test "$actual_sha256" = "$expected_sha256" || fail 'archive SHA-256 mismatch'

command -v gpg >/dev/null 2>&1 || fail 'gpg is required for detached-signature verification'
command -v gpgv >/dev/null 2>&1 || fail 'gpgv is required for detached-signature verification'
gpg_home="$receipt_root/go$version.$goos-$goarch.gnupg"
test ! -L "$gpg_home" || fail "refusing symlinked GnuPG home: $gpg_home"
mkdir -p "$gpg_home"
chmod 0700 "$gpg_home"
key_fingerprints=$(gpg --batch --no-options --homedir "$gpg_home" --with-colons --show-keys --fingerprint --fingerprint "$public_key" 2>/dev/null | awk -F: '$1 == "fpr" { print $10 }')
printf '%s\n' "$key_fingerprints" | grep -Fqx "$primary_fingerprint" || fail 'Google signing primary fingerprint missing'
printf '%s\n' "$key_fingerprints" | grep -Fqx "$signing_fingerprint" || fail 'Go signing subkey fingerprint missing'

keyring="$receipt_root/go$version.$goos-$goarch.keyring.gpg"
status_file="$receipt_root/go$version.$goos-$goarch.gpg-status"
members_file="$receipt_root/go$version.$goos-$goarch.archive-members"
test ! -L "$keyring" || fail "refusing symlinked keyring receipt: $keyring"
test ! -L "$status_file" || fail "refusing symlinked signature receipt: $status_file"
test ! -L "$members_file" || fail "refusing symlinked member receipt: $members_file"

if test ! -e "$keyring"; then
	gpg --batch --no-options --homedir "$gpg_home" --yes --dearmor --output "$keyring" "$public_key"
fi
gpgv --homedir "$gpg_home" --status-fd=1 --keyring "$keyring" "$signature" "$archive" >"$status_file"
L7_PRIMARY_FINGERPRINT=$primary_fingerprint \
L7_SIGNING_FINGERPRINT=$signing_fingerprint \
awk '
	$1 == "[GNUPG:]" &&
	$2 == "VALIDSIG" &&
	$3 == ENVIRON["L7_SIGNING_FINGERPRINT"] {
		for (i = 4; i <= NF; i++) {
			if ($i == ENVIRON["L7_PRIMARY_FINGERPRINT"]) valid = 1
		}
	}
	END { exit !valid }
' "$status_file" || fail 'detached signature identity mismatch'

tar -tzf "$archive" >"$members_file"
awk '
	$0 !~ /^go(\/|$)/ || $0 ~ /(^|\/)\.\.(\/|$)/ { bad = 1 }
	END { exit bad }
' "$members_file" || fail 'archive contains an unexpected member path'

if test "$reuse_destination" -eq 1; then
	actual_version=$(GOROOT="$destination" GOENV=off GOTOOLCHAIN=local GOWORK=off GOFLAGS='' GOOS="$goos" GOARCH="$goarch" GOEXPERIMENT='' TEST_TELEMETRY_DIR="$telemetry_root" "$destination/bin/go" env GOVERSION)
	test "$actual_version" = "go$version" || fail "existing binary reports $actual_version"
	printf 'bootstrap-go: archive and signature reauthenticated; extracted local cache was not tree-verified (%s, %s/%s, %s)\n' "$role" "$goos" "$goarch" "$destination"
	exit 0
fi

mkdir "$stage"
tar -xzf "$archive" -C "$stage"
test -x "$stage/go/bin/go" || fail 'archive did not contain the Go binary'
actual_version=$(GOROOT="$stage/go" GOENV=off GOTOOLCHAIN=local GOWORK=off GOFLAGS='' GOOS="$goos" GOARCH="$goarch" GOEXPERIMENT='' TEST_TELEMETRY_DIR="$telemetry_root" "$stage/go/bin/go" env GOVERSION)
test "$actual_version" = "go$version" || fail "extracted binary reports $actual_version"
mv "$stage/go" "$destination"
rmdir "$stage"

{
	printf 'version=go%s\n' "$version"
	printf 'role=%s\n' "$role"
	printf 'goos=%s\n' "$goos"
	printf 'goarch=%s\n' "$goarch"
	printf 'archive_sha256=%s\n' "$expected_sha256"
	printf 'signing_primary_fingerprint=%s\n' "$primary_fingerprint"
	printf 'signing_subkey_fingerprint=%s\n' "$signing_fingerprint"
} >"$receipt"

printf 'bootstrap-go: installed authenticated development toolchain (%s, %s/%s, %s)\n' "$role" "$goos" "$goarch" "$destination"
