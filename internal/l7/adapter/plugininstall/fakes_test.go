package plugininstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fake executables are POSIX shell scripts (jq for JSON) so that each of
// the hundreds of spawns per test run costs milliseconds, also under -race.

type fakeMarketplace struct {
	Name string `json:"name"`
	Root string `json:"root"`
}

type fakePlugin struct {
	Name        string `json:"name"`
	Marketplace string `json:"marketplace"`
}

// fakeHostState is the JSON document the fake hosts keep under HOME.
type fakeHostState struct {
	Marketplaces []fakeMarketplace `json:"marketplaces"`
	Plugins      []fakePlugin      `json:"plugins"`
}

func (state fakeHostState) marketplace(name string) *fakeMarketplace {
	for index := range state.Marketplaces {
		if state.Marketplaces[index].Name == name {
			return &state.Marketplaces[index]
		}
	}
	return nil
}

func writeFakes(t *testing.T, directory string) {
	t.Helper()
	for name, script := range map[string]string{
		"gh":     fakeGHScript,
		"codex":  strings.ReplaceAll(fakeHostScript, "__HOST__", "codex"),
		"claude": strings.ReplaceAll(fakeHostScript, "__HOST__", "claude"),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// fakeGHScript emulates `gh --version` and `gh attestation verify --bundle`.
// The bundle fixture carries the identity the attestation claims; the fake
// enforces the subject digest and --source-digest like the real CLI and then
// prints the certificate shape the verifier re-checks.
const fakeGHScript = `#!/bin/sh
set -eu
if command -v shasum >/dev/null 2>&1; then
	sha256() { shasum -a 256 "$1" | awk '{ print $1 }'; }
else
	sha256() { sha256sum "$1" | awk '{ print $1 }'; }
fi
if [ "${1:-}" = --version ]; then
	version=2.101.0
	if [ -f "$HOME/.fake-gh.control" ]; then
		configured=$(sed -n 's/^version=//p' "$HOME/.fake-gh.control" | head -n 1)
		[ -n "$configured" ] && version=$configured
	fi
	printf 'gh version %s (2026-09-15)\nhttps://github.com/cli/cli/releases/tag/v%s\n' "$version" "$version"
	exit 0
fi
[ "${1:-}" = attestation ] && [ "${2:-}" = verify ] || { printf 'unsupported fake gh invocation\n' >&2; exit 2; }
file=$3
shift 3
bundle=''; repo=''; signer=''; ref=''; source_digest=''; predicate=''; format=''; deny=0
while [ "$#" -gt 0 ]; do
	case $1 in
		--bundle) bundle=$2; shift 2 ;;
		--repo) repo=$2; shift 2 ;;
		--signer-workflow) signer=$2; shift 2 ;;
		--source-ref) ref=$2; shift 2 ;;
		--source-digest) source_digest=$2; shift 2 ;;
		--predicate-type) predicate=$2; shift 2 ;;
		--format) format=$2; shift 2 ;;
		--deny-self-hosted-runners) deny=1; shift ;;
		*) printf 'unexpected flag %s\n' "$1" >&2; exit 2 ;;
	esac
done
[ "$repo" = addressanup/level7-dev-loop ] && [ "$ref" = refs/heads/main ] && [ "$format" = json ] && [ "$predicate" = https://slsa.dev/provenance/v1 ] && [ "$deny" = 1 ] \
	|| { printf 'fake gh received unexpected enforcement flags\n' >&2; exit 2; }
[ -z "${GH_TOKEN:-}${GITHUB_TOKEN:-}" ] || { printf 'a token was forwarded to gh\n' >&2; exit 9; }
digest=$(sha256 "$file")
results=$(jq -c --arg digest "$digest" --arg source "$source_digest" --arg signer "$signer" --arg repo "$repo" '
	select(.l7test.sha256 == $digest and .l7test.commit == $source and ($repo + "/" + .l7test.workflow) == $signer) | .l7test |
	(("https://github.com/" + $repo + "/" + .workflow + "@refs/heads/main") as $identity |
	{verificationResult: {
		signature: {certificate: ({issuer: "https://token.actions.githubusercontent.com", buildSignerURI: $identity, subjectAlternativeName: $identity,
			sourceRepositoryURI: ("https://github.com/" + $repo), sourceRepositoryDigest: .commit, sourceRepositoryRef: "refs/heads/main",
			githubWorkflowRepository: $repo, runnerEnvironment: "github-hosted",
			runInvocationURI: ("https://github.com/" + $repo + "/actions/runs/38042885014/attempts/1")}
			+ (if (.tamper // "") != "" then {(.tamper): "https://github.com/evil/mirror"} else {} end))},
		statement: {predicateType: "https://slsa.dev/provenance/v1", subject: [{name: .subject, digest: {sha256: .sha256}}]},
		verifiedTimestamps: [{type: "Tlog", uri: "https://rekor.sigstore.dev", timestamp: "2026-10-10T15:38:12+05:45"}]}})' "$bundle" | jq -s -c '.')
if [ -z "$results" ] || [ "$results" = '[]' ]; then
	printf 'X Verification failed: no attestations matched the provided identity\n' >&2
	exit 1
fi
printf '%s\n' "$results"
`

// fakeHostScript emulates the Codex and Claude plugin managers: a JSON state
// file under HOME, the list shapes the real hosts print, host-specific verbs,
// and failure injection through a control file (fail=<step>, failroot=<substr>,
// version=<string>).
const fakeHostScript = `#!/bin/sh
set -eu
host=__HOST__
state="$HOME/.fake-$host.json"
control="$HOME/.fake-$host.control"
printf '%s\n' "$*" >>"$HOME/.fake-$host.log"
[ -f "$state" ] || printf '{"marketplaces":[],"plugins":[]}\n' >"$state"
ctl() { if [ -f "$control" ]; then sed -n "s/^$1=//p" "$control" | head -n 1; fi; }
fail_step=$(ctl fail)
fail_root=$(ctl failroot)
should_fail() {
	[ "$fail_step" = "$1" ] || [ "$fail_step" = all ] || return 1
	if [ -n "$fail_root" ]; then case $2 in *"$fail_root"*) ;; *) return 1 ;; esac; fi
	printf 'fake %s: injected failure for %s\n' "$host" "$1" >&2
	return 0
}
root_of() { jq -r --arg name "$1" '.marketplaces[] | select(.name == $name) | .root' "$state" | head -n 1; }
has_flag() { flag=$1; shift; for argument in "$@"; do [ "$argument" = "$flag" ] && return 0; done; return 1; }
has_scope() { previous=''; for argument in "$@"; do if [ "$previous" = --scope ] && [ "$argument" = user ]; then return 0; fi; previous=$argument; done; return 1; }
manifest_version() {
	if [ "$host" = codex ]; then file="$1/.codex-plugin/plugin.json"; else file="$1/.claude-plugin/plugin.json"; fi
	if [ -f "$file" ]; then jq -r '.version // ""' "$file"; fi
}
catalog_name() {
	if [ "$host" = codex ]; then file="$1/.agents/plugins/marketplace.json"; else file="$1/.claude-plugin/marketplace.json"; fi
	if [ -f "$file" ]; then jq -r '.name // ""' "$file"; fi
}
save() { jq -c "$@" "$state" >"$state.tmp" && mv "$state.tmp" "$state"; }

if [ "${1:-}" = --version ]; then
	configured=$(ctl version)
	if [ -n "$configured" ]; then printf '%s\n' "$configured"
	elif [ "$host" = codex ]; then printf 'codex-cli 0.162.1\n'
	else printf '2.1.291 (Claude Code)\n'; fi
	exit 0
fi
[ "${1:-}" = plugin ] || { printf 'fake %s: unsupported command\n' "$host" >&2; exit 2; }
case "${2:-}" in
	marketplace)
		case "${3:-}" in
			list)
				if [ "$host" = codex ]; then
					jq -c '{marketplaces: [.marketplaces[] | {name: .name, root: .root, marketplaceSource: {sourceType: "local", source: .root}}]}' "$state"
				else
					printf 'Loading marketplaces...\n'
					jq -c '[.marketplaces[] | {name: .name, source: "directory", path: .root, installLocation: .root}]' "$state"
				fi ;;
			add)
				root=${4:?}
				if should_fail marketplace-add "$root"; then exit 1; fi
				if [ "$host" = claude ] && ! has_scope "$@"; then printf 'missing --scope user\n' >&2; exit 2; fi
				name=$(catalog_name "$root")
				[ -n "$name" ] || { printf 'no marketplace manifest found\n' >&2; exit 1; }
				[ -z "$(root_of "$name")" ] || { printf 'marketplace "%s" already exists\n' "$name" >&2; exit 1; }
				save --arg name "$name" --arg root "$root" '.marketplaces += [{name: $name, root: $root}]'
				printf '{"ok":true,"name":"%s"}\n' "$name" ;;
			remove)
				name=${4:?}
				if should_fail marketplace-remove "$(root_of "$name")"; then exit 1; fi
				[ -n "$(root_of "$name")" ] || { printf 'marketplace "%s" is not configured\n' "$name" >&2; exit 1; }
				[ "$(jq -r --arg name "$name" '[.plugins[] | select(.marketplace == $name)] | length' "$state")" = 0 ] \
					|| { printf 'marketplace "%s" still has installed plugins\n' "$name" >&2; exit 1; }
				save --arg name "$name" '.marketplaces |= map(select(.name != $name))'
				printf '{"ok":true}\n' ;;
			*) exit 2 ;;
		esac ;;
	list)
		scoped=''
		previous=''
		for argument in "$@"; do if [ "$previous" = --marketplace ]; then scoped=$argument; fi; previous=$argument; done
		if [ "$host" = codex ] && [ -n "$scoped" ] && [ -z "$(root_of "$scoped")" ]; then
			printf 'marketplace "%s" is not configured\n' "$scoped" >&2
			exit 1
		fi
		entries="$state.entries.$$"
		: >"$entries"
		jq -r '.plugins[] | "\(.name)\t\(.marketplace)"' "$state" | while IFS="$(printf '\t')" read -r name marketplace; do
			[ -z "$scoped" ] || [ "$marketplace" = "$scoped" ] || continue
			root=$(root_of "$marketplace")
			version=$(manifest_version "$root")
			if [ "$host" = codex ]; then
				jq -n -c --arg name "$name" --arg marketplace "$marketplace" --arg version "$version" --arg root "$root" \
					'{pluginId: ($name + "@" + $marketplace), name: $name, marketplaceName: $marketplace, version: $version, installed: true, enabled: true, source: {source: "local", path: $root}}' >>"$entries"
			else
				jq -n -c --arg name "$name" --arg marketplace "$marketplace" --arg version "$version" --arg root "$root" --arg home "$HOME" \
					'{id: ($name + "@" + $marketplace), version: $version, scope: "user", enabled: true, installPath: ($home + "/.claude/plugins/cache/" + $marketplace + "/" + $name + "/" + $version), readFromFolder: $root}' >>"$entries"
			fi
		done
		if [ "$host" = codex ]; then
			jq -s -c '{installed: ., available: []}' "$entries"
		else
			jq -s -c '.' "$entries"
		fi
		rm -f "$entries" ;;
	add|install)
		if [ "$host" = codex ] && [ "$2" != add ]; then printf 'wrong install verb for host\n' >&2; exit 2; fi
		if [ "$host" = claude ] && [ "$2" != install ]; then printf 'wrong install verb for host\n' >&2; exit 2; fi
		selector=${3:?}
		name=${selector%@*}
		marketplace=${selector#*@}
		if should_fail plugin-add "$(root_of "$marketplace")"; then exit 1; fi
		[ -n "$(root_of "$marketplace")" ] || { printf 'plugin "%s" not found in any configured marketplace\n' "$selector" >&2; exit 1; }
		if [ "$host" = claude ] && ! has_scope "$@"; then printf 'missing --scope user\n' >&2; exit 2; fi
		[ "$(jq -r --arg name "$name" --arg marketplace "$marketplace" '[.plugins[] | select(.name == $name and .marketplace == $marketplace)] | length' "$state")" = 0 ] \
			|| { printf 'already installed\n' >&2; exit 1; }
		save --arg name "$name" --arg marketplace "$marketplace" '.plugins += [{name: $name, marketplace: $marketplace}]'
		printf '{"ok":true}\n' ;;
	remove|uninstall)
		if [ "$host" = codex ] && [ "$2" != remove ]; then exit 2; fi
		if [ "$host" = claude ] && [ "$2" != uninstall ]; then exit 2; fi
		if [ "$host" = claude ] && ! has_flag --keep-data "$@"; then printf 'plugin data must be kept\n' >&2; exit 2; fi
		selector=${3:?}
		name=${selector%@*}
		marketplace=${selector#*@}
		if should_fail plugin-remove "$(root_of "$marketplace")"; then exit 1; fi
		[ "$(jq -r --arg name "$name" --arg marketplace "$marketplace" '[.plugins[] | select(.name == $name and .marketplace == $marketplace)] | length' "$state")" != 0 ] \
			|| { printf 'plugin is not installed\n' >&2; exit 1; }
		save --arg name "$name" --arg marketplace "$marketplace" '.plugins |= map(select(.name != $name or .marketplace != $marketplace))'
		printf '{"ok":true}\n' ;;
	*) printf 'fake %s: unsupported command\n' "$host" >&2; exit 2 ;;
esac
`
