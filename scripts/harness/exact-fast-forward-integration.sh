#!/bin/sh

set -eu

program=exact-fast-forward-integration
github_api_version=2022-11-28
github_actions_app_id=15368
required_checks_json='["Go 1.26.7 (baseline)","CLI macOS 15 (arm64)","CLI macOS 15 (amd64)","CLI paired benchmark gate","evaluate"]'

fail()
{
	printf '%s: BLOCKED: %s\n' "$program" "$*" >&2
	exit 1
}

usage()
{
	cat >&2 <<'EOF'
usage: exact-fast-forward-integration.sh \
  --repo OWNER/REPOSITORY \
  --pr NUMBER \
  --expected-base FULL_SHA \
  --expected-head FULL_SHA \
  --expected-tree FULL_SHA \
  --accountable-owner GITHUB_LOGIN \
  --auditor GITHUB_LOGIN
EOF
	exit 2
}

require_command()
{
	command -v "$1" >/dev/null 2>&1 || fail "required command is unavailable: $1"
}

is_full_sha()
{
	printf '%s\n' "$1" | grep -Eq '^[0-9a-f]{40}$'
}

is_github_login()
{
	test "${#1}" -le 39 &&
		printf '%s\n' "$1" | grep -Eq '^[A-Za-z0-9]([A-Za-z0-9-]{0,37}[A-Za-z0-9])?$'
}

repo=
pull_request=
expected_base=
expected_head=
expected_tree=
accountable_owner=
auditor=

while test "$#" -gt 0; do
	option=$1
	shift
	case $option in
		--repo)
			test -z "$repo" || fail 'duplicate --repo'
			test "$#" -gt 0 || usage
			repo=$1
			shift
			;;
		--pr)
			test -z "$pull_request" || fail 'duplicate --pr'
			test "$#" -gt 0 || usage
			pull_request=$1
			shift
			;;
		--expected-base)
			test -z "$expected_base" || fail 'duplicate --expected-base'
			test "$#" -gt 0 || usage
			expected_base=$1
			shift
			;;
		--expected-head)
			test -z "$expected_head" || fail 'duplicate --expected-head'
			test "$#" -gt 0 || usage
			expected_head=$1
			shift
			;;
		--expected-tree)
			test -z "$expected_tree" || fail 'duplicate --expected-tree'
			test "$#" -gt 0 || usage
			expected_tree=$1
			shift
			;;
		--accountable-owner)
			test -z "$accountable_owner" || fail 'duplicate --accountable-owner'
			test "$#" -gt 0 || usage
			accountable_owner=$1
			shift
			;;
		--auditor)
			test -z "$auditor" || fail 'duplicate --auditor'
			test "$#" -gt 0 || usage
			auditor=$1
			shift
			;;
		*)
			usage
			;;
	esac
done

test -n "$repo" || usage
test -n "$pull_request" || usage
test -n "$expected_base" || usage
test -n "$expected_head" || usage
test -n "$expected_tree" || usage
test -n "$accountable_owner" || usage
test -n "$auditor" || usage

printf '%s\n' "$repo" | grep -Eq '^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$' || fail 'repository must be exactly OWNER/REPOSITORY'
case $pull_request in
	''|0|0*|*[!0-9]*) fail 'pull-request number must be a positive canonical integer' ;;
esac
is_full_sha "$expected_base" || fail 'expected base must be one full lowercase SHA'
is_full_sha "$expected_head" || fail 'expected head must be one full lowercase SHA'
is_full_sha "$expected_tree" || fail 'expected tree must be one full lowercase SHA'
test "$expected_base" != "$expected_head" || fail 'expected base and head must differ'
is_github_login "$accountable_owner" || fail 'accountable owner is not a canonical GitHub login'
is_github_login "$auditor" || fail 'auditor is not a canonical GitHub login'
test "$accountable_owner" != "$auditor" || fail 'accountable owner and auditor must differ'

require_command gh
require_command git
require_command jq
require_command grep
require_command mktemp
require_command env

gh_bin=$(command -v gh)
git_bin=$(command -v git)
env_bin=$(command -v env)
remote_url=https://github.com/$repo.git
protection_endpoint=repos/$repo/branches/main/protection
admins_endpoint=$protection_endpoint/enforce_admins

export GH_PROMPT_DISABLED=1 GIT_TERMINAL_PROMPT=0 LC_ALL=C
unset GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES
unset GIT_CONFIG GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT GIT_CONFIG_KEY_0 GIT_CONFIG_VALUE_0
unset GIT_SSH GIT_SSH_COMMAND GIT_PROXY_COMMAND

api()
{
	"$gh_bin" api --hostname github.com \
		-H 'Accept: application/vnd.github+json' \
		-H "X-GitHub-Api-Version: $github_api_version" "$@"
}

api_pages()
{
	api --paginate --slurp "$1"
}

canonical_json()
{
	jq -cS .
}

validate_repository()
{
	printf '%s\n' "$1" | jq -e --arg repo "$repo" '
		.full_name == $repo and
		.owner.type == "User" and
		.fork == false and
		.default_branch == "main" and
		.archived == false and
		.disabled == false
	' >/dev/null || fail 'repository identity or state is outside the approved contract'
	printf '%s\n' "$1" | jq -e '.delete_branch_on_merge == false' >/dev/null ||
		fail 'automatic source-branch deletion is enabled'
}

repository_contract_for()
{
	printf '%s\n' "$1" | jq -cS '{
		full_name, default_branch, archived, disabled, fork,
		allow_merge_commit, allow_squash_merge, allow_rebase_merge,
		allow_auto_merge, delete_branch_on_merge, allow_update_branch
	}'
}

validate_collaborators()
{
	printf '%s\n' "$1" | jq -e --arg actor "$actor" '
		([.[][] | select(.permissions.admin == true) | .login] | unique) == [$actor]
	' >/dev/null || fail 'active actor is not the sole direct repository administrator'
}

validate_rulesets()
{
	printf '%s\n' "$1" | jq -e '[.[][]] | length == 0' >/dev/null ||
		fail 'repository rulesets make the bypass scope ambiguous'
}

validate_pull()
{
	printf '%s\n' "$1" | jq -e \
		--arg repo "$repo" \
		--arg base "$expected_base" \
		--arg head "$expected_head" \
		--argjson pr "$pull_request" '
			.number == $pr and
			.state == "open" and
			.merged == false and
			.draft == false and
			.mergeable == true and
			.mergeable_state == "clean" and
			.base.ref == "main" and
			.base.repo.full_name == $repo and
			.base.sha == $base and
			.head.repo.full_name == $repo and
			.head.sha == $head
		' >/dev/null || fail 'pull request is not the exact open, clean, unmerged candidate'
}

validate_combined_status()
{
	printf '%s\n' "$1" | jq -e --argjson required "$required_checks_json" '
		[.statuses[]? | .context as $context | select(any($required[]; . == $context))] | length == 0
	' >/dev/null || fail 'a legacy status duplicates a required check context'
}

validate_protection()
{
	protection_json=$1
	expected_admin_state=$2
	printf '%s\n' "$protection_json" | jq -e \
		--argjson required "$required_checks_json" \
		--argjson app "$github_actions_app_id" \
		--argjson admins "$expected_admin_state" '
			.required_status_checks.strict == true and
			([.required_status_checks.contexts[]] | sort) == ($required | sort) and
			([.required_status_checks.checks[] | {context, app_id}] | sort_by(.context)) ==
				([$required[] | {context: ., app_id: $app}] | sort_by(.context)) and
			.required_pull_request_reviews.dismiss_stale_reviews == true and
			.required_pull_request_reviews.require_code_owner_reviews == true and
			.required_pull_request_reviews.require_last_push_approval == true and
			.required_pull_request_reviews.required_approving_review_count == 1 and
			.required_conversation_resolution.enabled == true and
			.required_linear_history.enabled == true and
			.enforce_admins.enabled == $admins and
			.allow_force_pushes.enabled == false and
			.allow_deletions.enabled == false and
			.block_creations.enabled == false and
			.required_signatures.enabled == false and
			.lock_branch.enabled == false and
			.allow_fork_syncing.enabled == false and
			.restrictions == null
		' >/dev/null || fail "main protection does not match the approved contract (enforce_admins=$expected_admin_state)"
}

assert_main_and_source()
{
	required_main=$1
	main_ref=$(api "repos/$repo/git/ref/heads/main")
	printf '%s\n' "$main_ref" | jq -e --arg sha "$required_main" '
		.ref == "refs/heads/main" and .object.type == "commit" and .object.sha == $sha
	' >/dev/null || fail "remote main is not the required commit: $required_main"

	source_refs=$(api "repos/$repo/git/matching-refs/heads/$head_ref")
	printf '%s\n' "$source_refs" | jq -e --arg ref "refs/heads/$head_ref" --arg sha "$expected_head" '
		[.[] | select(.ref == $ref and .object.type == "commit" and .object.sha == $sha)] | length == 1
	' >/dev/null || fail 'pull-request source branch no longer names the exact candidate'
}

require_latest_approval()
{
	review_actor=$1
	printf '%s\n' "$reviews" | jq -e --arg actor "$review_actor" --arg head "$expected_head" '
		[.[] | select(.user.login == $actor)]
		| sort_by(.id)
		| last
		| .state == "APPROVED" and .commit_id == $head
	' >/dev/null || fail "latest review from $review_actor is not an exact-head approval"
}

validate_check()
{
	check_name=$1
	printf '%s\n' "$check_runs" | jq -e \
		--arg name "$check_name" \
		--arg head "$expected_head" \
		--arg repo "$repo" \
		--argjson pr "$pull_request" \
		--argjson app "$github_actions_app_id" '
			[.[] | select(.name == $name)] as $named |
			[$named[] | select(.app.id == $app)] as $trusted |
			($named | length) > 0 and
			([$named[] | select(.app.id != $app)] | length) == 0 and
			($trusted | length) > 0 and
			($trusted | map(.id) | unique | length) == ($trusted | length) and
			($trusted | all(.started_at | type == "string" and
				test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"))) and
			(($trusted | map(.started_at) | max) as $latest_started |
				[$trusted[] | select(.started_at == $latest_started)] as $latest |
				($latest | length) == 1 and
				($latest[0] |
					.head_sha == $head and
					.status == "completed" and
					.conclusion == "success" and
					.app.slug == "github-actions" and
					(.details_url | startswith("https://github.com/" + $repo + "/actions/runs/")) and
					(any(.pull_requests[]?; .number == $pr and .head.sha == $head))))
		' >/dev/null || fail "required check is absent, stale, ambiguous, wrong-app, or unsuccessful: $check_name"
}

operation_root=
admins_may_be_disabled=false
restoration_attempted=false
restoration_failed=false
protection_restored=

restore_admins()
{
	test "$admins_may_be_disabled" = true || return 0
	test "$restoration_attempted" = false || return 1
	restoration_attempted=true
	if ! api --method POST "$admins_endpoint" >/dev/null; then
		restoration_failed=true
		return 1
	fi
	restored_json=$(api "$protection_endpoint") || {
		restoration_failed=true
		return 1
	}
	restored_contract=$(printf '%s\n' "$restored_json" | canonical_json) || {
		restoration_failed=true
		return 1
	}
	if test "$restored_contract" != "$protection_original"; then
		restoration_failed=true
		return 1
	fi
	protection_restored=$restored_contract
	admins_may_be_disabled=false
	return 0
}

cleanup()
{
	status=$?
	trap - EXIT HUP INT TERM
	if test "$admins_may_be_disabled" = true && test "$restoration_attempted" = false; then
		restore_admins || :
	fi
	if test "$admins_may_be_disabled" = true; then
		restoration_failed=true
	fi
	if test -n "$operation_root"; then
		case $operation_root in
			*/l7-exact-fast-forward.*) rm -rf -- "$operation_root" ;;
			*) printf '%s: refused unsafe temporary cleanup path: %s\n' "$program" "$operation_root" >&2 ;;
		esac
	fi
	if test "$restoration_failed" = true; then
		printf '%s: BLOCKED RECOVERY: administrator enforcement may be disabled. Sole next action: re-enable administrator enforcement on %s main.\n' "$program" "$repo" >&2
		status=70
	fi
	exit "$status"
}

trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

actor_json=$(api user)
actor=$(printf '%s\n' "$actor_json" | jq -er '.login') || fail 'could not resolve the active GitHub actor'

repository_json=$(api "repos/$repo")
validate_repository "$repository_json"
repository_contract=$(repository_contract_for "$repository_json") ||
	fail 'repository settings could not be canonicalized'

configured_owner=$(api "repos/$repo/actions/variables/L7_ACCOUNTABLE_OWNER" | jq -er '.value') || fail 'repository accountable-owner variable is unavailable'
test "$configured_owner" = "$accountable_owner" || fail 'requested accountable owner does not match L7_ACCOUNTABLE_OWNER'

collaborator_pages=$(api_pages "repos/$repo/collaborators?affiliation=direct&per_page=100")
validate_collaborators "$collaborator_pages"

ruleset_pages=$(api_pages "repos/$repo/rulesets?per_page=100")
validate_rulesets "$ruleset_pages"

pull_json=$(api "repos/$repo/pulls/$pull_request")
validate_pull "$pull_json"
head_ref=$(printf '%s\n' "$pull_json" | jq -er '.head.ref') || fail 'pull-request source ref is unavailable'
pr_author=$(printf '%s\n' "$pull_json" | jq -er '.user.login') || fail 'pull-request author is unavailable'
test "$head_ref" != main || fail 'pull-request source ref cannot be main'
test "$pr_author" != "$accountable_owner" || fail 'pull-request author and accountable owner must differ'
test "$pr_author" != "$auditor" || fail 'pull-request author and auditor must differ'

assert_main_and_source "$expected_base"

candidate_commit=$(api "repos/$repo/git/commits/$expected_head")
printf '%s\n' "$candidate_commit" | jq -e --arg sha "$expected_head" --arg tree "$expected_tree" '
	.sha == $sha and .tree.sha == $tree
' >/dev/null || fail 'candidate commit or tree does not match the exact request'

comparison=$(api "repos/$repo/compare/$expected_base...$expected_head")
printf '%s\n' "$comparison" | jq -e --arg base "$expected_base" --arg head "$expected_head" '
	.status == "ahead" and
	.ahead_by > 0 and
	.behind_by == 0 and
	.base_commit.sha == $base and
	.merge_base_commit.sha == $base and
	.commits[-1].sha == $head and
	.total_commits == .ahead_by
' >/dev/null || fail 'candidate is not an exact fast-forward descendant of the expected base'

protection_json=$(api "$protection_endpoint")
validate_protection "$protection_json" true
protection_original=$(printf '%s\n' "$protection_json" | canonical_json)

review_pages=$(api_pages "repos/$repo/pulls/$pull_request/reviews?per_page=100")
reviews=$(printf '%s\n' "$review_pages" | jq -c '[.[][]]')
require_latest_approval "$accountable_owner"
require_latest_approval "$auditor"

check_pages=$(api_pages "repos/$repo/commits/$expected_head/check-runs?per_page=100")
check_runs=$(printf '%s\n' "$check_pages" | jq -c '[.[].check_runs[]]')
validate_check 'Go 1.26.7 (baseline)'
validate_check 'CLI macOS 15 (arm64)'
validate_check 'CLI macOS 15 (amd64)'
validate_check 'CLI paired benchmark gate'
validate_check 'evaluate'

combined_status=$(api "repos/$repo/commits/$expected_head/status")
validate_combined_status "$combined_status"

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=${temporary_parent%/}
test -d "$temporary_parent" || temporary_parent=/tmp
case $temporary_parent in ''|/) fail 'unsafe temporary parent' ;; esac
operation_root=$(mktemp -d "$temporary_parent/l7-exact-fast-forward.XXXXXX")
case $operation_root in "$temporary_parent"/l7-exact-fast-forward.*) ;; *) fail 'mktemp returned an unsafe path' ;; esac
bare_repo=$operation_root/repository.git
askpass=$operation_root/askpass.sh
git_home=$operation_root/home
git_tmp=$operation_root/tmp
git_token_file=$operation_root/token
umask 077
mkdir "$git_home" "$git_tmp"
"$gh_bin" auth token --hostname github.com >"$git_token_file" ||
	fail 'could not obtain a GitHub token for the isolated Git transport'
test -s "$git_token_file" || fail 'GitHub returned an empty token for the isolated Git transport'
cat >"$askpass" <<'EOF'
#!/bin/sh
case ${1:-} in
	*Username*) printf '%s\n' x-access-token ;;
	*Password*) cat "$L7_GIT_TOKEN_FILE" ;;
	*) exit 1 ;;
esac
EOF
chmod 700 "$askpass"

clean_git()
{
	"$env_bin" -i \
		HOME="$git_home" \
		PATH=/usr/bin:/bin:/usr/sbin:/sbin \
		TMPDIR="$git_tmp" \
		LC_ALL=C \
		GIT_CONFIG_NOSYSTEM=1 \
		GIT_CONFIG_SYSTEM=/dev/null \
		GIT_CONFIG_GLOBAL=/dev/null \
		GIT_TERMINAL_PROMPT=0 \
		GIT_ASKPASS="$askpass" \
		SSH_ASKPASS="$askpass" \
		L7_GIT_TOKEN_FILE="$git_token_file" \
		"$git_bin" "$@"
}

clean_git check-ref-format "refs/heads/$head_ref" >/dev/null 2>&1 || fail 'pull-request source ref is malformed'
clean_git init --bare --quiet "$bare_repo"
clean_git -C "$bare_repo" fetch --quiet --no-tags "$remote_url" \
	"refs/heads/$head_ref:refs/l7/candidate"
test "$(clean_git -C "$bare_repo" rev-parse refs/l7/candidate)" = "$expected_head" || fail 'fetched source branch is not the exact candidate'
test "$(clean_git -C "$bare_repo" rev-parse "$expected_head^{tree}")" = "$expected_tree" || fail 'fetched candidate tree is not exact'
clean_git -C "$bare_repo" merge-base --is-ancestor "$expected_base" "$expected_head" || fail 'fetched candidate is not a fast-forward descendant'

test -t 0 && test -t 1 && test -t 2 || fail 'active terminal input and output are required'
cat <<EOF
Exact-candidate integration is ready for final confirmation.

Repository:        $repo
Pull request:      #$pull_request
Remote ref:        refs/heads/main
Expected old SHA:  $expected_base
Candidate SHA:     $expected_head
Candidate tree:    $expected_tree
Protection change: temporarily disable administrator enforcement only
Rollback boundary: restore administrator enforcement before any postflight

Type the complete candidate SHA to continue:
EOF
IFS= read -r confirmation || fail 'confirmation input ended before a full SHA was read'
test "$confirmation" = "$expected_head" || fail 'confirmation did not equal the complete candidate SHA'

# Revalidate every remote authority and mutable prerequisite after the human
# decision and before opening the narrow protection window.
test "$(api user | jq -er '.login')" = "$actor" || fail 'active GitHub actor changed after confirmation'

refreshed_repository_json=$(api "repos/$repo")
validate_repository "$refreshed_repository_json"
refreshed_repository_contract=$(repository_contract_for "$refreshed_repository_json") ||
	fail 'repository settings could not be canonicalized after confirmation'
test "$refreshed_repository_contract" = "$repository_contract" ||
	fail 'repository settings changed after confirmation'

test "$(api "repos/$repo/actions/variables/L7_ACCOUNTABLE_OWNER" | jq -er '.value')" = "$accountable_owner" || fail 'accountable-owner binding changed after confirmation'

refreshed_collaborator_pages=$(api_pages "repos/$repo/collaborators?affiliation=direct&per_page=100")
validate_collaborators "$refreshed_collaborator_pages"

refreshed_ruleset_pages=$(api_pages "repos/$repo/rulesets?per_page=100")
validate_rulesets "$refreshed_ruleset_pages"

refreshed_pull_json=$(api "repos/$repo/pulls/$pull_request")
validate_pull "$refreshed_pull_json"
test "$(printf '%s\n' "$refreshed_pull_json" | jq -er '.head.ref')" = "$head_ref" ||
	fail 'pull-request source ref changed after confirmation'
test "$(printf '%s\n' "$refreshed_pull_json" | jq -er '.user.login')" = "$pr_author" ||
	fail 'pull-request author changed after confirmation'

assert_main_and_source "$expected_base"
test "$(api "$protection_endpoint" | canonical_json)" = "$protection_original" || fail 'protection changed after confirmation'

review_pages=$(api_pages "repos/$repo/pulls/$pull_request/reviews?per_page=100")
reviews=$(printf '%s\n' "$review_pages" | jq -c '[.[][]]')
require_latest_approval "$accountable_owner"
require_latest_approval "$auditor"

check_pages=$(api_pages "repos/$repo/commits/$expected_head/check-runs?per_page=100")
check_runs=$(printf '%s\n' "$check_pages" | jq -c '[.[].check_runs[]]')
validate_check 'Go 1.26.7 (baseline)'
validate_check 'CLI macOS 15 (arm64)'
validate_check 'CLI macOS 15 (amd64)'
validate_check 'CLI paired benchmark gate'
validate_check 'evaluate'

combined_status=$(api "repos/$repo/commits/$expected_head/status")
validate_combined_status "$combined_status"

admins_may_be_disabled=true
api --method DELETE "$admins_endpoint" >/dev/null || fail 'administrator enforcement disable request failed or was ambiguous'

protection_disabled=$(api "$protection_endpoint")
validate_protection "$protection_disabled" false
expected_disabled=$(printf '%s\n' "$protection_original" | jq -cS '.enforce_admins.enabled = false')
test "$(printf '%s\n' "$protection_disabled" | canonical_json)" = "$expected_disabled" || fail 'a protection field other than administrator enforcement changed'
assert_main_and_source "$expected_base"

remote_refs=$(clean_git ls-remote --refs "$remote_url" refs/heads/main "refs/heads/$head_ref")
test "$(printf '%s\n' "$remote_refs" | awk '$2 == "refs/heads/main" { print $1 }')" = "$expected_base" || fail 'remote main changed before the lease-bound update'
test "$(printf '%s\n' "$remote_refs" | awk -v ref="refs/heads/$head_ref" '$2 == ref { print $1 }')" = "$expected_head" || fail 'source branch changed before the lease-bound update'

# The full expected-old lease is the compare-and-swap. The candidate has already
# been proven to descend from that exact old SHA, so this can only be a
# fast-forward. No generic force option, leading-plus refspec, merge API, or
# additional refspec is used.
clean_git -C "$bare_repo" -c core.hooksPath=/dev/null push --porcelain --atomic \
	"--force-with-lease=refs/heads/main:$expected_base" \
	"$remote_url" "$expected_head:refs/heads/main" || fail 'lease-bound fast-forward push failed or lost the ref race'

restore_admins || exit 70

test "$protection_restored" = "$protection_original" || fail 'complete branch protection was not restored exactly'
assert_main_and_source "$expected_head"

post_commit=$(api "repos/$repo/git/commits/$expected_head")
printf '%s\n' "$post_commit" | jq -e --arg sha "$expected_head" --arg tree "$expected_tree" '
	.sha == $sha and .tree.sha == $tree
' >/dev/null || fail 'remote main reached a commit with the wrong tree'

post_pull=$(api "repos/$repo/pulls/$pull_request")
printf '%s\n' "$post_pull" | jq -e \
	--arg repo "$repo" \
	--arg head "$expected_head" \
	--arg source "$head_ref" '
		.state == "closed" and
		.merged == true and
		.merged_at != null and
		.merge_commit_sha == $head and
		.base.ref == "main" and
		.base.repo.full_name == $repo and
		.head.ref == $source and
		.head.repo.full_name == $repo and
		.head.sha == $head
	' >/dev/null || fail 'GitHub did not report the exact pull request as indirectly merged; it was not edited or closed by this tool'

post_repository=$(api "repos/$repo" | jq -cS '{
	full_name, default_branch, archived, disabled, fork,
	allow_merge_commit, allow_squash_merge, allow_rebase_merge,
	allow_auto_merge, delete_branch_on_merge, allow_update_branch
}')
test "$post_repository" = "$repository_contract" || fail 'repository merge or identity settings changed during integration'

post_rulesets=$(api_pages "repos/$repo/rulesets?per_page=100")
printf '%s\n' "$post_rulesets" | jq -e '[.[][]] | length == 0' >/dev/null || fail 'repository rulesets changed during integration'

printf '%s: PASS repository=%s pr=%s main=%s tree=%s protection=restored source=unchanged\n' \
	"$program" "$repo" "$pull_request" "$expected_head" "$expected_tree"
