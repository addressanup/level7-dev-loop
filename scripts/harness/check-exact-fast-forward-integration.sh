#!/bin/sh

set -eu

fail()
{
	printf 'check-exact-fast-forward-integration: %s\n' "$*" >&2
	exit 1
}

script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
operator=$script_dir/exact-fast-forward-integration.sh
test -x "$operator" || fail "operator is not executable: $operator"
sh -n "$operator"

base=1111111111111111111111111111111111111111
head=2222222222222222222222222222222222222222
tree=3333333333333333333333333333333333333333
race=4444444444444444444444444444444444444444
repo=fixture-owner/fixture-repository
pull_request=7
accountable_owner=owner-reviewer
auditor=audit-user

temporary_parent=${TMPDIR:-/tmp}
temporary_parent=${temporary_parent%/}
test -d "$temporary_parent" || temporary_parent=/tmp
case $temporary_parent in ''|/) fail 'unsafe temporary parent' ;; esac
test_root=$(mktemp -d "$temporary_parent/l7-exact-fast-forward-test.XXXXXX")
case $test_root in "$temporary_parent"/l7-exact-fast-forward-test.*) ;; *) fail 'mktemp returned an unsafe path' ;; esac
cleanup()
{
	rm -rf -- "$test_root"
}
trap cleanup EXIT HUP INT TERM

fake_bin=$test_root/bin
state_dir=$test_root/state
mkdir "$fake_bin" "$state_dir"

cat >"$fake_bin/gh" <<'EOF'
#!/bin/sh
set -eu

printf 'gh %s\n' "$*" >>"$FAKE_STATE_DIR/log"

if test "${1:-}" = auth; then
	printf '%s\n' fixture-token
	exit 0
fi
test "${1:-}" = api || exit 91
shift

method=GET
endpoint=
while test "$#" -gt 0; do
	case $1 in
		--hostname|-H)
			test "$#" -ge 2 || exit 92
			shift 2
			;;
		--method)
			test "$#" -ge 2 || exit 92
			method=$2
			shift 2
			;;
		--paginate|--slurp)
			shift
			;;
		*)
			test -z "$endpoint" || exit 92
			endpoint=$1
			shift
			;;
	esac
done

main_sha=$(sed -n '1p' "$FAKE_STATE_DIR/main")
admins=$(sed -n '1p' "$FAKE_STATE_DIR/admins")

case "$method:$endpoint" in
	GET:user)
		actor=fixture-admin
		test "$FAKE_SCENARIO" != wrong_actor || actor=wrong-admin
		printf '{"login":"%s"}\n' "$actor"
		;;
	GET:repos/$FAKE_REPO)
		jq -n --arg repo "$FAKE_REPO" '{
			full_name:$repo, owner:{type:"User"}, fork:false,
			default_branch:"main", archived:false, disabled:false,
			allow_merge_commit:true, allow_squash_merge:true,
			allow_rebase_merge:true, allow_auto_merge:false,
			delete_branch_on_merge:false, allow_update_branch:false
		}'
		;;
	GET:repos/$FAKE_REPO/actions/variables/L7_ACCOUNTABLE_OWNER)
		jq -n --arg value "$FAKE_OWNER" '{name:"L7_ACCOUNTABLE_OWNER",value:$value}'
		;;
	GET:repos/$FAKE_REPO/collaborators\?*)
		jq -n --arg actor fixture-admin '[[
			{login:$actor,permissions:{admin:true}},
			{login:"owner-reviewer",permissions:{admin:false}}
		]]'
		;;
	GET:repos/$FAKE_REPO/rulesets\?*)
		printf '[[]]\n'
		;;
	GET:repos/$FAKE_REPO/pulls/$FAKE_PR)
		pr_base=$FAKE_BASE
		pr_head=$FAKE_HEAD
		test "$FAKE_SCENARIO" != stale_base || pr_base=$FAKE_RACE
		test "$FAKE_SCENARIO" != stale_head || pr_head=$FAKE_RACE
		state=open
		merged=false
		merged_at=null
		merge_commit=null
		if test "$main_sha" = "$FAKE_HEAD" && test "$FAKE_SCENARIO" != postcondition_failure; then
			state=closed
			merged=true
			merged_at='"2026-08-29T00:00:00Z"'
			merge_commit='"'$FAKE_HEAD'"'
		fi
		jq -n \
			--argjson number "$FAKE_PR" \
			--arg state "$state" \
			--argjson merged "$merged" \
			--argjson merged_at "$merged_at" \
			--argjson merge_commit "$merge_commit" \
			--arg repo "$FAKE_REPO" \
			--arg base "$pr_base" \
			--arg head "$pr_head" '{
				number:$number, state:$state, merged:$merged,
				merged_at:$merged_at, merge_commit_sha:$merge_commit,
				draft:false, mergeable:true, mergeable_state:"clean",
				user:{login:"pr-author"},
				base:{ref:"main",sha:$base,repo:{full_name:$repo}},
				head:{ref:"candidate",sha:$head,repo:{full_name:$repo}}
			}'
		;;
	GET:repos/$FAKE_REPO/git/ref/heads/main)
		jq -n --arg sha "$main_sha" '{ref:"refs/heads/main",object:{type:"commit",sha:$sha}}'
		;;
	GET:repos/$FAKE_REPO/git/matching-refs/heads/candidate)
		jq -n --arg sha "$FAKE_HEAD" '[{ref:"refs/heads/candidate",object:{type:"commit",sha:$sha}}]'
		;;
	GET:repos/$FAKE_REPO/git/commits/$FAKE_HEAD)
		candidate_tree=$FAKE_TREE
		test "$FAKE_SCENARIO" != stale_tree || candidate_tree=$FAKE_RACE
		jq -n --arg sha "$FAKE_HEAD" --arg tree "$candidate_tree" '{sha:$sha,tree:{sha:$tree}}'
		;;
	GET:repos/$FAKE_REPO/compare/$FAKE_BASE...$FAKE_HEAD)
		jq -n --arg base "$FAKE_BASE" --arg head "$FAKE_HEAD" '{
			status:"ahead",ahead_by:1,behind_by:0,total_commits:1,
			base_commit:{sha:$base},merge_base_commit:{sha:$base},
			commits:[{sha:$head}]
		}'
		;;
	GET:repos/$FAKE_REPO/branches/main/protection)
		linear=true
		test "$FAKE_SCENARIO" != unsafe_protection || linear=false
		jq -n --argjson admins "$admins" --argjson linear "$linear" '{
			required_status_checks:{
				strict:true,
				contexts:[
					"Go 1.26.7 (baseline)",
					"CLI macOS 15 (arm64)",
					"CLI macOS 15 (amd64)",
					"CLI paired benchmark gate",
					"evaluate"
				],
				checks:[
					{context:"Go 1.26.7 (baseline)",app_id:15368},
					{context:"CLI macOS 15 (arm64)",app_id:15368},
					{context:"CLI macOS 15 (amd64)",app_id:15368},
					{context:"CLI paired benchmark gate",app_id:15368},
					{context:"evaluate",app_id:15368}
				]
			},
			required_pull_request_reviews:{
				dismiss_stale_reviews:true,
				require_code_owner_reviews:true,
				require_last_push_approval:true,
				required_approving_review_count:1
			},
			required_conversation_resolution:{enabled:true},
			required_linear_history:{enabled:$linear},
			enforce_admins:{enabled:$admins},
			allow_force_pushes:{enabled:false},
			allow_deletions:{enabled:false},
			block_creations:{enabled:false},
			required_signatures:{enabled:false},
			lock_branch:{enabled:false},
			allow_fork_syncing:{enabled:false}
		}'
		;;
	GET:repos/$FAKE_REPO/pulls/$FAKE_PR/reviews\?*)
		if test "$FAKE_SCENARIO" = missing_approval; then
			jq -n --arg head "$FAKE_HEAD" '[[
				{id:2,user:{login:"audit-user"},state:"APPROVED",commit_id:$head}
			]]'
		else
			jq -n --arg head "$FAKE_HEAD" '[[
				{id:1,user:{login:"owner-reviewer"},state:"APPROVED",commit_id:$head},
				{id:2,user:{login:"audit-user"},state:"APPROVED",commit_id:$head}
			]]'
		fi
		;;
	GET:repos/$FAKE_REPO/commits/$FAKE_HEAD/check-runs\?*)
		baseline=success
		test "$FAKE_SCENARIO" != failed_check || baseline=failure
		ambiguous=false
		test "$FAKE_SCENARIO" != ambiguous_check || ambiguous=true
		jq -n \
			--arg head "$FAKE_HEAD" \
			--arg repo "$FAKE_REPO" \
			--arg baseline "$baseline" \
			--argjson pr "$FAKE_PR" \
			--argjson ambiguous "$ambiguous" '
			def run($id; $name; $conclusion; $app; $slug): {
				id:$id, name:$name, head_sha:$head, status:"completed",
				conclusion:$conclusion, app:{id:$app,slug:$slug},
				details_url:("https://github.com/" + $repo + "/actions/runs/1/job/" + ($id|tostring)),
				pull_requests:[{number:$pr,head:{sha:$head}}]
			};
			[
				{
					total_count:(if $ambiguous then 6 else 5 end),
					check_runs:([
						run(10;"Go 1.26.7 (baseline)";$baseline;15368;"github-actions"),
						run(20;"CLI macOS 15 (arm64)";"success";15368;"github-actions"),
						run(30;"CLI macOS 15 (amd64)";"success";15368;"github-actions"),
						run(40;"CLI paired benchmark gate";"success";15368;"github-actions"),
						run(50;"evaluate";"success";15368;"github-actions")
					] + if $ambiguous then [run(60;"Go 1.26.7 (baseline)";"success";999;"other-app")] else [] end)
				}
			]'
		;;
	GET:repos/$FAKE_REPO/commits/$FAKE_HEAD/status)
		printf '{"statuses":[]}\n'
		;;
	DELETE:repos/$FAKE_REPO/branches/main/protection/enforce_admins)
		printf '%s\n' false >"$FAKE_STATE_DIR/admins"
		if test "$FAKE_SCENARIO" = ref_race; then
			printf '%s\n' "$FAKE_RACE" >"$FAKE_STATE_DIR/main"
		fi
		;;
	POST:repos/$FAKE_REPO/branches/main/protection/enforce_admins)
		if test "$FAKE_SCENARIO" = restoration_failure; then
			exit 93
		fi
		printf '%s\n' true >"$FAKE_STATE_DIR/admins"
		;;
	*)
		printf 'unexpected fake gh request: %s %s\n' "$method" "$endpoint" >&2
		exit 94
		;;
esac
EOF

cat >"$fake_bin/git" <<'EOF'
#!/bin/sh
set -eu

printf 'git %s\n' "$*" >>"$FAKE_STATE_DIR/log"

subcommand=
for argument in "$@"; do
	case $argument in
		check-ref-format|init|fetch|rev-parse|merge-base|ls-remote|push)
			subcommand=$argument
			break
			;;
	esac
done

case $subcommand in
	check-ref-format)
		exit 0
		;;
	init)
		last=
		for argument in "$@"; do last=$argument; done
		mkdir -p "$last"
		;;
	fetch)
		exit 0
		;;
	rev-parse)
		case "$*" in
			*'^{tree}'*) printf '%s\n' "$FAKE_TREE" ;;
			*) printf '%s\n' "$FAKE_HEAD" ;;
		esac
		;;
	merge-base)
		exit 0
		;;
	ls-remote)
		printf '%s\trefs/heads/main\n' "$(sed -n '1p' "$FAKE_STATE_DIR/main")"
		printf '%s\trefs/heads/candidate\n' "$FAKE_HEAD"
		;;
	push)
		lease=false
		refspecs=0
		for argument in "$@"; do
			case $argument in
				--force|-f|+*) exit 95 ;;
				--force-with-lease=refs/heads/main:$FAKE_BASE) lease=true ;;
				$FAKE_HEAD:refs/heads/main) refspecs=$((refspecs + 1)) ;;
			esac
		done
		test "$lease" = true || exit 96
		test "$refspecs" -eq 1 || exit 97
		if test "$FAKE_SCENARIO" = push_failure; then
			exit 98
		fi
		printf '%s\n' "$FAKE_HEAD" >"$FAKE_STATE_DIR/main"
		printf 'ok refs/heads/main\n'
		;;
	*)
		printf 'unexpected fake git request: %s\n' "$*" >&2
		exit 99
		;;
esac
EOF

cat >"$test_root/invoke.sh" <<'EOF'
#!/bin/sh
exec "$L7_OPERATOR" \
	--repo "$FAKE_REPO" \
	--pr "$FAKE_PR" \
	--expected-base "$FAKE_BASE" \
	--expected-head "$FAKE_HEAD" \
	--expected-tree "$FAKE_TREE" \
	--accountable-owner "$FAKE_OWNER" \
	--auditor "$FAKE_AUDITOR"
EOF

chmod 700 "$fake_bin/gh" "$fake_bin/git" "$test_root/invoke.sh"

export PATH="$fake_bin:$PATH"
export FAKE_STATE_DIR="$state_dir"
export FAKE_BASE="$base" FAKE_HEAD="$head" FAKE_TREE="$tree" FAKE_RACE="$race"
export FAKE_REPO="$repo" FAKE_PR="$pull_request" FAKE_OWNER="$accountable_owner" FAKE_AUDITOR="$auditor"
export L7_OPERATOR="$operator"

case_output=$test_root/output
case_status=0

run_case()
{
	FAKE_SCENARIO=$1
	mode=$2
	confirmation=$3
	export FAKE_SCENARIO
	printf '%s\n' "$base" >"$state_dir/main"
	printf '%s\n' true >"$state_dir/admins"
	: >"$state_dir/log"
	: >"$case_output"
	set +e
	if test "$mode" = terminal; then
		case $(uname -s) in
			Darwin)
				(sleep 1; printf '%s\n' "$confirmation") | script -q /dev/null "$test_root/invoke.sh" >"$case_output" 2>&1
				;;
			Linux)
				(sleep 1; printf '%s\n' "$confirmation") | script -q -e -c "$test_root/invoke.sh" /dev/null >"$case_output" 2>&1
				;;
			*)
				set -e
				fail 'contract test requires Darwin or Linux script(1) PTY support'
				;;
		esac
	else
		printf '%s\n' "$confirmation" | "$test_root/invoke.sh" >"$case_output" 2>&1
	fi
	case_status=$?
	set -e
}

assert_no_mutation()
{
	if grep -Eq 'gh .*--method DELETE .*protection/enforce_admins' "$state_dir/log"; then
		fail "$1 reached administrator-enforcement mutation"
	fi
	if grep -Eq 'git .* push( |$)' "$state_dir/log"; then
		fail "$1 reached a ref update"
	fi
}

assert_restoration_attempted()
{
	grep -Eq 'gh .*--method DELETE .*protection/enforce_admins' "$state_dir/log" || fail "$1 did not open the expected protection window"
	grep -Eq 'gh .*--method POST .*protection/enforce_admins' "$state_dir/log" || fail "$1 did not attempt protection restoration"
}

expect_failure()
{
	scenario=$1
	mode=$2
	confirmation=$3
	expected_message=$4
	mutation=$5
	run_case "$scenario" "$mode" "$confirmation"
	test "$case_status" -ne 0 || fail "$scenario unexpectedly succeeded"
	grep -Fq "$expected_message" "$case_output" || {
		sed -n '1,160p' "$case_output" >&2
		fail "$scenario did not report: $expected_message"
	}
	case $mutation in
		none) assert_no_mutation "$scenario" ;;
		restored) assert_restoration_attempted "$scenario" ;;
		*) fail "unknown mutation expectation: $mutation" ;;
	esac
}

expect_argument_failure()
{
	label=$1
	shift
	: >"$case_output"
	set +e
	"$operator" "$@" >"$case_output" 2>&1
	status=$?
	set -e
	test "$status" -ne 0 || fail "$label unexpectedly accepted malformed input"
}

expect_argument_failure abbreviated-sha \
	--repo "$repo" --pr "$pull_request" --expected-base "$base" \
	--expected-head 2222 --expected-tree "$tree" \
	--accountable-owner "$accountable_owner" --auditor "$auditor"
expect_argument_failure duplicate-input \
	--repo "$repo" --repo "$repo" --pr "$pull_request" \
	--expected-base "$base" --expected-head "$head" --expected-tree "$tree" \
	--accountable-owner "$accountable_owner" --auditor "$auditor"
expect_argument_failure extra-input \
	--repo "$repo" --pr "$pull_request" --expected-base "$base" \
	--expected-head "$head" --expected-tree "$tree" \
	--accountable-owner "$accountable_owner" --auditor "$auditor" --unexpected value

expect_failure wrong_actor terminal "$head" 'active actor is not the sole direct repository administrator' none
expect_failure stale_base terminal "$head" 'pull request is not the exact open, clean, unmerged candidate' none
expect_failure stale_head terminal "$head" 'pull request is not the exact open, clean, unmerged candidate' none
expect_failure stale_tree terminal "$head" 'candidate commit or tree does not match the exact request' none
expect_failure missing_approval terminal "$head" 'latest review from owner-reviewer is not an exact-head approval' none
expect_failure failed_check terminal "$head" 'required check is absent, stale, ambiguous, wrong-app, or unsuccessful: Go 1.26.7 (baseline)' none
expect_failure ambiguous_check terminal "$head" 'required check is absent, stale, ambiguous, wrong-app, or unsuccessful: Go 1.26.7 (baseline)' none
expect_failure unsafe_protection terminal "$head" 'main protection does not match the approved contract' none
expect_failure valid nonterminal "$head" 'active terminal input and output are required' none
expect_failure valid terminal "$race" 'confirmation did not equal the complete candidate SHA' none
expect_failure ref_race terminal "$head" 'remote main is not the required commit' restored
expect_failure push_failure terminal "$head" 'lease-bound fast-forward push failed or lost the ref race' restored
expect_failure restoration_failure terminal "$head" 'BLOCKED RECOVERY: administrator enforcement may be disabled' restored
expect_failure postcondition_failure terminal "$head" 'GitHub did not report the exact pull request as indirectly merged' restored

run_case valid terminal "$head"
test "$case_status" -eq 0 || { sed -n '1,200p' "$case_output" >&2; fail 'valid contract scenario failed'; }
grep -Fq 'exact-fast-forward-integration: PASS' "$case_output" || fail 'valid scenario did not report PASS'

disable_line=$(grep -n 'gh .*--method DELETE .*protection/enforce_admins' "$state_dir/log" | cut -d: -f1)
push_line=$(grep -n 'git .* push ' "$state_dir/log" | cut -d: -f1)
restore_line=$(grep -n 'gh .*--method POST .*protection/enforce_admins' "$state_dir/log" | cut -d: -f1)
test -n "$disable_line" && test -n "$push_line" && test -n "$restore_line" || fail 'valid scenario omitted a controlled effect'
test "$disable_line" -lt "$push_line" && test "$push_line" -lt "$restore_line" || fail 'valid effect ordering is unsafe'
test "$(grep -c 'git .* push ' "$state_dir/log")" -eq 1 || fail 'valid scenario reached more than one remote ref update'

if grep -Eq 'git .* push .*([[:space:]]--force([[:space:]]|$)|[[:space:]]-f([[:space:]]|$)|[[:space:]]\+)' "$state_dir/log"; then
	fail 'valid scenario used a generic force flag or leading-plus refspec'
fi
if grep -Eq 'gh .*repos/.*/merges([[:space:]]|$)' "$state_dir/log"; then
	fail 'valid scenario reached the merge API'
fi
if grep -Eq 'gh .*--method (PATCH|PUT|DELETE) .*pulls/' "$state_dir/log"; then
	fail 'valid scenario edited or closed the pull request'
fi
if grep -Eq 'gh .*--method DELETE .*git/(refs|matching-refs)' "$state_dir/log"; then
	fail 'valid scenario reached branch deletion'
fi
test "$(sed -n '1p' "$state_dir/main")" = "$head" || fail 'valid scenario did not leave main at the exact head'
test "$(sed -n '1p' "$state_dir/admins")" = true || fail 'valid scenario did not restore administrator enforcement'

printf 'check-exact-fast-forward-integration: PASS (3 malformed-input probes, 14 failure scenarios, 1 ordered success scenario)\n'
