#!/bin/sh

set -eu
umask 077
PATH=/usr/bin:/bin
export PATH

fail()
{
	printf 'check-bootstrap-modules: %s\n' "$*" >&2
	exit 1
}

test "$#" -eq 0 || fail 'usage: check-bootstrap-modules.sh'

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
source_script="$project_root/scripts/harness/bootstrap-modules.sh"
test -x "$source_script" || fail 'bootstrap-modules.sh is missing or not executable'

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=${temporary_parent%/}
temporary_parent=$(CDPATH='' cd "$temporary_parent" && pwd -P)
scratch=$(mktemp -d "$temporary_parent/l7-bootstrap-modules.XXXXXX")
case $scratch in "$temporary_parent"/l7-bootstrap-modules.*) ;; *) fail 'mktemp returned an unsafe test directory' ;; esac
cleanup()
{
	case $scratch in "$temporary_parent"/l7-bootstrap-modules.*) rm -rf -- "$scratch" ;; *) return 1 ;; esac
}
trap cleanup EXIT HUP INT TERM

make_fixture()
{
	label=$1
	behavior=$2
	fixture="$scratch/$label"
	mkdir -p \
		"$fixture/scripts/harness" \
		"$fixture/.cache/toolchains/go1.26.7/bin" \
		"$fixture/.cache/go/mod" \
		"$fixture/.cache/go/build" \
		"$fixture/.cache/go/tmp" \
		"$fixture/.cache/go/telemetry"
	cp "$source_script" "$fixture/scripts/harness/bootstrap-modules.sh"
	cp "$project_root/go.mod" "$fixture/go.mod"
	cp "$project_root/go.sum" "$fixture/go.sum"
	printf '%s\n' "$behavior" >"$fixture/behavior"
	printf 'off 2026-08-24\n' >"$fixture/.cache/go/telemetry/mode"
	cat >"$fixture/.cache/toolchains/go1.26.7/bin/go" <<EOF
#!/bin/sh
set -eu
behavior=\$(cat '$fixture/behavior')
args=
for argument in "\$@"; do args="\${args}\${argument}|"; done
printf 'args=%s proxy=%s sumdb=%s vcs=%s auth=%s private=%s noproxy=%s nosumdb=%s insecure=%s home=%s modcache=%s secret=%s https_proxy=%s path=%s\n' \
	"\$args" "\$GOPROXY" "\$GOSUMDB" "\$GOVCS" "\$GOAUTH" "\$GOPRIVATE" "\$GONOPROXY" "\$GONOSUMDB" "\$GOINSECURE" \
	"\$HOME" "\$GOMODCACHE" "\${SECRET_TOKEN-unset}" "\${HTTPS_PROXY-unset}" "\$PATH" >>'$fixture/calls.log'
case "\$*" in
	'env GOVERSION') printf 'go1.26.7\n' ;;
	'mod download')
		if test "\$behavior" = mutate; then printf '\n// mutation\n' >>'$fixture/go.mod'; fi
		if test "\$behavior" = fail-download; then exit 23; fi
		;;
	'mod verify')
		if test "\$behavior" = fail-verify; then exit 24; fi
		;;
	'list -mod=readonly -m all') printf 'github.com/addressanup/level7-dev-loop\n' ;;
	*) exit 99 ;;
esac
EOF
	chmod 0700 "$fixture/.cache/toolchains/go1.26.7/bin/go"
}

make_fixture success success
cp "$fixture/go.mod" "$fixture/go.mod.before"
cp "$fixture/go.sum" "$fixture/go.sum.before"
HOME=/sensitive/home \
GOPROXY=https://example.invalid \
GOSUMDB=evil.invalid \
GOPRIVATE=secret.example \
GOAUTH=netrc \
HTTPS_PROXY=http://proxy.invalid \
SECRET_TOKEN=must-not-cross \
	"$fixture/scripts/harness/bootstrap-modules.sh" \
	"$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output"
grep -Fq 'bootstrap-modules: PASS' "$fixture/output" || fail 'success fixture did not pass'
test "$(wc -l <"$fixture/calls.log" | tr -d ' ')" -eq 3 || fail 'unexpected Go command count'
grep -Fq 'args=env|GOVERSION| proxy=off sumdb=off vcs=*:off auth=off private= noproxy= nosumdb= insecure=' "$fixture/calls.log" || fail 'toolchain check was not offline'
grep -Fq 'args=mod|download| proxy=https://proxy.golang.org sumdb=sum.golang.org vcs=*:off auth=off private= noproxy= nosumdb= insecure=' "$fixture/calls.log" || fail 'download environment is not fixed'
grep -Fq 'args=mod|verify| proxy=off sumdb=off vcs=*:off auth=off' "$fixture/calls.log" || fail 'module verification was not offline'
grep -Fq "home=$fixture/.cache/go/tmp/l7-bootstrap-home." "$fixture/calls.log" || fail 'bootstrap home is not disposable and repository-local'
grep -Fq "modcache=$fixture/.cache/go/mod secret=unset https_proxy=unset path=/usr/bin:/bin" "$fixture/calls.log" || fail 'ambient environment crossed the bootstrap boundary'
if find "$fixture/.cache/go/tmp" -type d -name 'l7-bootstrap-home.*' | grep -q .; then fail 'disposable bootstrap home was not removed'; fi
cmp "$fixture/go.mod.before" "$fixture/go.mod" || fail 'success fixture changed go.mod'
cmp "$fixture/go.sum.before" "$fixture/go.sum" || fail 'success fixture changed go.sum'

make_fixture fail-download fail-download
cp "$fixture/go.mod" "$fixture/go.mod.before"
cp "$fixture/go.sum" "$fixture/go.sum.before"
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output" 2>&1; then
	fail 'download failure was accepted'
fi
grep -Fq 'go command failed with status 23: mod download' "$fixture/output" || fail 'download failure was not reported'
if grep -Fq 'args=mod|verify|' "$fixture/calls.log"; then fail 'verification ran after failed download'; fi
cmp "$fixture/go.mod.before" "$fixture/go.mod" || fail 'failed download changed go.mod'
cmp "$fixture/go.sum.before" "$fixture/go.sum" || fail 'failed download changed go.sum'

make_fixture fail-verify fail-verify
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output" 2>&1; then
	fail 'module verification failure was accepted'
fi
grep -Fq 'go command failed with status 24: mod verify' "$fixture/output" || fail 'verification failure was not reported'

make_fixture mutate mutate
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output" 2>&1; then
	fail 'dependency-file mutation was accepted'
fi
grep -Fq 'go.mod changed during bootstrap' "$fixture/output" || fail 'dependency-file mutation was not reported'

make_fixture missing-sum success
rm "$fixture/go.sum"
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output" 2>&1; then
	fail 'missing go.sum was accepted'
fi
grep -Fq 'dependency input is missing' "$fixture/output" || fail 'missing go.sum was not reported'

make_fixture symlink-sum success
mv "$fixture/go.sum" "$fixture/go.sum.target"
ln -s go.sum.target "$fixture/go.sum"
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 >"$fixture/output" 2>&1; then
	fail 'symlinked go.sum was accepted'
fi
grep -Fq 'dependency input is a symlink' "$fixture/output" || fail 'symlinked go.sum was not reported'

make_fixture invalid-input success
if (cd "$fixture" && scripts/harness/bootstrap-modules.sh .cache/toolchains/go1.26.7/bin/go 1.26.7) >"$fixture/output" 2>&1; then
	fail 'relative Go binary was accepted'
fi
if "$fixture/scripts/harness/bootstrap-modules.sh" "$fixture/.cache/toolchains/go1.26.7/bin/go" 1.26.7 extra >"$fixture/output" 2>&1; then
	fail 'extra argument was accepted'
fi

printf 'check-bootstrap-modules: PASS (fixed environment; 3 commands; 6 negative fixtures; no network)\n'
