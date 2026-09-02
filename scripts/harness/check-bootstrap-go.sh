#!/bin/sh

set -eu
umask 077
PATH=/usr/bin:/bin
export PATH

fail()
{
	printf 'check-bootstrap-go: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 0 || fail 'usage: check-bootstrap-go.sh'

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
source_script="$project_root/scripts/harness/bootstrap-go.sh"
test -x "$source_script" || fail 'bootstrap-go.sh is missing or not executable'

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=${temporary_parent%/}
temporary_parent=$(CDPATH='' cd "$temporary_parent" && pwd -P)
scratch=$(mktemp -d "$temporary_parent/l7-bootstrap-go.XXXXXX")
case $scratch in "$temporary_parent"/l7-bootstrap-go.*) ;; *) fail 'mktemp returned an unsafe test directory' ;; esac
cleanup()
{
	case $scratch in "$temporary_parent"/l7-bootstrap-go.*) rm -rf -- "$scratch" ;; *) return 1 ;; esac
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

archive_url=https://go.dev/dl/go1.26.7.darwin-arm64.tar.gz

make_fixture()
{
	label=$1
	behavior=$2
	fixture="$scratch/$label"
	mkdir -p "$fixture/scripts/harness" "$fixture/harness" "$fixture/fakebin"
	cp "$source_script" "$fixture/scripts/harness/bootstrap-go.sh"
	cp "$project_root/harness/toolchains.lock.tsv" "$fixture/harness/toolchains.lock.tsv"
	cp "$project_root/harness/signing-identities.lock.tsv" "$fixture/harness/signing-identities.lock.tsv"
	chmod 0700 "$fixture/scripts/harness/bootstrap-go.sh"
	printf '%s\n' "$behavior" >"$fixture/behavior"

	cat >"$fixture/fakebin/uname" <<'EOF'
#!/bin/sh
case ${1:-} in
	-s) printf 'Darwin\n' ;;
	-m) printf 'arm64\n' ;;
	*) exit 64 ;;
esac
EOF
	cat >"$fixture/fakebin/sleep" <<'EOF'
#!/bin/sh
set -eu
test "$#" -eq 1
printf '%s\n' "$1" >>"$L7_FIXTURE_ROOT/sleeps.log"
EOF
	cat >"$fixture/fakebin/date" <<'EOF'
#!/bin/sh
set -eu
if test "$(cat "$L7_FIXTURE_ROOT/behavior")" != deadline; then
	exec /bin/date "$@"
fi
count=0
if test -f "$L7_FIXTURE_ROOT/date.count"; then read -r count <"$L7_FIXTURE_ROOT/date.count"; fi
count=$((count + 1))
printf '%s\n' "$count" >"$L7_FIXTURE_ROOT/date.count"
case $count in
	1|2) printf '1000\n' ;;
	*) printf '1600\n' ;;
esac
EOF
	cat >"$fixture/fakebin/curl" <<'EOF'
#!/bin/sh
set -eu
: "${L7_FIXTURE_ROOT:?}"
behavior=$(cat "$L7_FIXTURE_ROOT/behavior")
count=0
if test -f "$L7_FIXTURE_ROOT/calls.count"; then read -r count <"$L7_FIXTURE_ROOT/calls.count"; fi
count=$((count + 1))
printf '%s\n' "$count" >"$L7_FIXTURE_ROOT/calls.count"
for argument in "$@"; do
	printf 'call=%s arg=%s\n' "$count" "$argument" >>"$L7_FIXTURE_ROOT/arguments.log"
done

output=
url=
max_time=
while test "$#" -gt 0; do
	case $1 in
		--output)
			test "$#" -ge 2
			output=$2
			shift 2
			;;
		--max-time)
			test "$#" -ge 2
			max_time=$2
			shift 2
			;;
		--write-out)
			test "$#" -ge 2
			shift 2
			;;
		*)
			url=$1
			shift
			;;
	esac
done
test -n "$output"
test -n "$url"
test -n "$max_time"
printf 'call=%s output=%s url=%s max_time=%s\n' "$count" "$output" "$url" "$max_time" >>"$L7_FIXTURE_ROOT/transfers.log"

case $url in
	*.tar.gz) kind=archive ;;
	*.tar.gz.asc) kind=signature ;;
	*) kind=key ;;
esac

if test "$kind" != archive; then
	printf 'partial-%s\n' "$count" >"$output"
	printf '000'
	exit 60
fi

case $behavior in
	receive-then-success)
		if test "$count" -eq 1; then
			printf 'partial-%s\n' "$count" >"$output"
			printf '500'
			exit 56
		fi
		;;
	persistent-receive|deadline)
		printf 'partial-%s\n' "$count" >"$output"
		printf '500'
		exit 56
		;;
	timeout-then-success)
		if test "$count" -eq 1; then
			printf 'partial-%s\n' "$count" >"$output"
			printf '000'
			exit 28
		fi
		;;
	http-*-then-success)
		if test "$count" -eq 1; then
			http_status=${behavior#http-}
			http_status=${http_status%-then-success}
			printf 'partial-%s\n' "$count" >"$output"
			printf '%s' "$http_status"
			exit 22
		fi
		;;
	http-nonretry)
		printf 'partial-%s\n' "$count" >"$output"
		printf '404'
		exit 22
		;;
	tls)
		printf 'partial-%s\n' "$count" >"$output"
		printf '000'
		exit 60
		;;
	local-write)
		printf 'partial-%s\n' "$count" >"$output"
		printf '000'
		exit 23
		;;
	signal)
		printf 'partial-%s\n' "$count" >"$output"
		kill -TERM "$PPID"
		/bin/sleep 1
		printf '000'
		exit 56
		;;
	stop-after-cache)
		exit 99
		;;
	*) exit 98 ;;
esac

printf 'complete-%s\n' "$count" >"$output"
printf '200'
exit 0
EOF
	chmod 0700 "$fixture/fakebin/uname" "$fixture/fakebin/sleep" "$fixture/fakebin/date" "$fixture/fakebin/curl"
	mkdir "$fixture/ambient-home"
	printf 'retry-all-errors\nretry 99\nmax-time 9999\n' >"$fixture/ambient-home/.curlrc"
}

run_fixture()
{
	if env \
		PATH="$fixture/fakebin:/usr/bin:/bin" \
		L7_FIXTURE_ROOT="$fixture" \
		L7_BOOTSTRAP_ATTEMPTS=99 \
		L7_BOOTSTRAP_RETRY_DELAY=0 \
		L7_BOOTSTRAP_MAX_TIME=9999 \
		CURL_RETRY=99 \
		CURL_HOME="$fixture/ambient-home" \
		HOME="$fixture/ambient-home" \
		HTTPS_PROXY=http://proxy.invalid \
		"$fixture/scripts/harness/bootstrap-go.sh" 1.26.7 >"$fixture/output.log" 2>&1
	then
		fixture_status=0
	else
		fixture_status=$?
	fi
}

archive_path()
{
	printf '%s\n' "$fixture/.cache/go-downloads/go1.26.7.darwin-arm64.tar.gz"
}

assert_call_count()
{
	expected=$1
	actual=0
	if test -f "$fixture/calls.count"; then read -r actual <"$fixture/calls.count"; fi
	test "$actual" -eq "$expected" || fail "$fixture: got $actual curl calls, want $expected"
}

assert_archive_call_count()
{
	expected=$1
	actual=$(awk -v expected="url=$archive_url" '$3 == expected { count++ } END { print count + 0 }' "$fixture/transfers.log")
	test "$actual" -eq "$expected" || fail "$fixture: got $actual archive calls, want $expected"
}

assert_no_partials()
{
	if find "$fixture/.cache/go-downloads" -name '*.partial.*' -print -quit 2>/dev/null | grep -q .; then
		fail "$fixture: partial download survived"
	fi
}

assert_one_second_delay()
{
	printf '1\n' >"$fixture/expected-sleeps"
	cmp "$fixture/expected-sleeps" "$fixture/sleeps.log" >/dev/null || fail "$fixture: retry delay was not exactly one second"
}

assert_transport_arguments()
{
	for required in --disable --fail --silent --show-error --location --proto --proto-redir --tlsv1.2 --connect-timeout --max-time --write-out --output; do
		grep -Fq "arg=$required" "$fixture/arguments.log" || fail "$fixture: missing curl argument $required"
	done
	awk '
		$1 != previous { if ($2 != "arg=--disable") bad = 1; previous = $1 }
		END { exit bad }
	' "$fixture/arguments.log" || fail "$fixture: curl did not disable ambient configuration first"
	if grep -Eq 'arg=--retry($|=)|arg=--retry-all-errors($|=)' "$fixture/arguments.log"; then
		fail "$fixture: curl contains nested or blanket retry"
	fi
	awk '
		$4 !~ /^max_time=[0-9]+$/ { bad = 1; next }
		{
			split($4, pair, "=")
			if (pair[2] < 1 || pair[2] > 600) bad = 1
		}
		END { exit bad }
	' "$fixture/transfers.log" || fail "$fixture: curl max-time escaped the fixed aggregate boundary"
}

assert_unique_archive_temporaries()
{
	expected=$1
	awk -v expected="url=$archive_url" '$3 == expected { sub(/^output=/, "", $2); print $2 }' "$fixture/transfers.log" >"$fixture/archive-temporaries"
	actual=$(sort -u "$fixture/archive-temporaries" | wc -l | tr -d ' ')
	test "$actual" -eq "$expected" || fail "$fixture: archive attempts did not use $expected unique temporary paths"
	while IFS= read -r temporary; do
		case $temporary in
			"$(archive_path)".partial.*) ;;
			*) fail "$fixture: curl wrote outside a same-directory temporary path: $temporary" ;;
		esac
	done <"$fixture/archive-temporaries"
}

make_fixture receive-then-success receive-then-success
run_fixture
test "$fixture_status" -eq 60 || fail 'receive-then-success did not stop at the later TLS fixture'
assert_call_count 3
assert_archive_call_count 2
assert_one_second_delay
assert_transport_arguments
assert_unique_archive_temporaries 2
test "$(cat "$(archive_path)")" = complete-2 || fail 'successful retry did not atomically install the complete archive'
grep -Fq 'curl=56 http=500' "$fixture/output.log" || fail 'receive failure status was not reported'
assert_no_partials

make_fixture persistent-receive persistent-receive
run_fixture
test "$fixture_status" -eq 56 || fail 'persistent receive failure did not preserve curl status 56'
assert_call_count 4
assert_archive_call_count 4
printf '1\n2\n4\n' >"$fixture/expected-sleeps"
cmp "$fixture/expected-sleeps" "$fixture/sleeps.log" >/dev/null || fail 'persistent receive delays were not exactly 1, 2, 4 seconds'
assert_unique_archive_temporaries 4
test ! -e "$(archive_path)" || fail 'persistent receive failure installed a final archive'
grep -Fq 'download failed after 4 attempts (curl=56 http=500)' "$fixture/output.log" || fail 'exhausted receive failure was not reported'
assert_no_partials

make_fixture timeout-then-success timeout-then-success
run_fixture
test "$fixture_status" -eq 60 || fail 'timeout-then-success did not stop at the later TLS fixture'
assert_call_count 3
assert_archive_call_count 2
assert_one_second_delay
test "$(cat "$(archive_path)")" = complete-2 || fail 'timeout retry did not install the complete archive'
grep -Fq 'curl=28 http=000' "$fixture/output.log" || fail 'timeout status was not reported'
assert_no_partials

for http_status in 408 429 500 502 503 504; do
	make_fixture "http-$http_status" "http-$http_status-then-success"
	run_fixture
	test "$fixture_status" -eq 60 || fail "HTTP $http_status retry did not stop at the later TLS fixture"
	assert_call_count 3
	assert_archive_call_count 2
	assert_one_second_delay
	test "$(cat "$(archive_path)")" = complete-2 || fail "HTTP $http_status retry did not install the complete archive"
	grep -Fq "curl=22 http=$http_status" "$fixture/output.log" || fail "HTTP $http_status status was not reported"
	assert_no_partials
done

make_fixture http-nonretry http-nonretry
run_fixture
test "$fixture_status" -eq 22 || fail 'non-retryable HTTP failure did not preserve curl status 22'
assert_call_count 1
assert_archive_call_count 1
test ! -s "$fixture/sleeps.log" 2>/dev/null || fail 'non-retryable HTTP failure slept'
test ! -e "$(archive_path)" || fail 'non-retryable HTTP failure installed a final archive'
grep -Fq 'download failure is not retryable (curl=22 http=404)' "$fixture/output.log" || fail 'non-retryable HTTP status was not reported'
assert_no_partials

make_fixture tls tls
run_fixture
test "$fixture_status" -eq 60 || fail 'TLS failure did not preserve curl status 60'
assert_call_count 1
test ! -s "$fixture/sleeps.log" 2>/dev/null || fail 'TLS failure was retried'
grep -Fq 'download failure is not retryable (curl=60 http=000)' "$fixture/output.log" || fail 'TLS failure was not reported'
assert_no_partials

make_fixture local-write local-write
run_fixture
test "$fixture_status" -eq 23 || fail 'local-write failure did not preserve curl status 23'
assert_call_count 1
test ! -s "$fixture/sleeps.log" 2>/dev/null || fail 'local-write failure was retried'
grep -Fq 'download failure is not retryable (curl=23 http=000)' "$fixture/output.log" || fail 'local-write failure was not reported'
assert_no_partials

make_fixture deadline deadline
run_fixture
test "$fixture_status" -eq 28 || fail 'aggregate deadline failure did not return timeout status 28'
assert_call_count 1
test ! -s "$fixture/sleeps.log" 2>/dev/null || fail 'aggregate deadline failure slept past its boundary'
grep -Fq 'download aggregate deadline exceeded before retry delay' "$fixture/output.log" || fail 'aggregate deadline failure was not reported'
assert_no_partials

make_fixture preexisting stop-after-cache
mkdir -p "$fixture/.cache/go-downloads"
printf 'preserved-cache\n' >"$(archive_path)"
run_fixture
test "$fixture_status" -eq 60 || fail 'pre-existing cache fixture did not stop at the later TLS fixture'
assert_call_count 1
test "$(cat "$(archive_path)")" = preserved-cache || fail 'pre-existing regular cache was changed'
assert_archive_call_count 0
assert_no_partials

make_fixture symlink stop-after-cache
mkdir -p "$fixture/.cache/go-downloads"
printf 'symlink-target\n' >"$fixture/cache-target"
ln -s "$fixture/cache-target" "$(archive_path)"
run_fixture
test "$fixture_status" -eq 1 || fail 'symlinked cache path was accepted'
test -L "$(archive_path)" || fail 'symlinked cache path was replaced'
test "$(cat "$fixture/cache-target")" = symlink-target || fail 'symlink target was changed'
assert_call_count 0
grep -Fq 'refusing symlinked download' "$fixture/output.log" || fail 'symlink refusal was not reported'
assert_no_partials

make_fixture signal signal
run_fixture
test "$fixture_status" -eq 143 || fail 'signal fixture did not preserve TERM status 143'
test ! -e "$(archive_path)" || fail 'signal fixture installed a final archive'
assert_no_partials

printf 'check-bootstrap-go: PASS (4-attempt fixed policy; 16 offline transport fixtures; atomic cache; cleanup)\n'
