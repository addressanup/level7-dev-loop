#!/bin/sh

# Explicitly consequential. No defaults, retries, cleanup of remote state, or
# credential copying. Run only for the owner-authorized unsigned stable release.
set -eu
test "$#" = 6 || {
  echo 'usage: publish-unsigned-stable.sh COMMIT TREE RUN_ID ARTIFACT_ID NEW_ABSOLUTE_WORK_ROOT CONFIRMATION' >&2
  exit 1
}
script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)
CANDIDATE_COMMIT=$1
CANDIDATE_TREE=$2
RELEASE_RUN_ID=$3
artifact_id=$4
work=$5
WAIVER_ACKNOWLEDGEMENT=$6
test "$WAIVER_ACKNOWLEDGEMENT" = 'I accept unsigned, unnotarized, unqualified stable v1.0.0'
printf '%s\n' "$CANDIDATE_COMMIT" | grep -Eq '^[0-9a-f]{40}$'
printf '%s\n' "$CANDIDATE_TREE" | grep -Eq '^[0-9a-f]{40}$'
printf '%s\n' "$RELEASE_RUN_ID" | grep -Eq '^[0-9]+$'
printf '%s\n' "$artifact_id" | grep -Eq '^[0-9]+$'
case "$work" in /*) ;; *) echo 'work root must be absolute' >&2; exit 1 ;; esac
test ! -e "$work"
test ! -L "$work"
umask 077
mkdir -m 0700 "$work"
test "$(CDPATH='' cd "$work" && pwd -P)" = "$work"
mkdir -m 0700 "$work/assets"
export CANDIDATE_COMMIT CANDIDATE_TREE RELEASE_RUN_ID WAIVER_ACKNOWLEDGEMENT
RELEASE_RUN_ATTEMPT=1
export RELEASE_RUN_ATTEMPT
repo=addressanup/level7-dev-loop
cd "$project_root"

preflight=$(sh "$script_dir/check-stable-release-preflight.sh" --publish "$artifact_id")
L7_RELEASE_BASE=$(printf '%s' "$preflight" | jq -er .base)
AUTHORIZATION_REFERENCE=$(printf '%s' "$preflight" | jq -er .authorization)
export L7_RELEASE_BASE AUTHORIZATION_REFERENCE
artifact=$(gh api "repos/$repo/actions/artifacts/$artifact_id")
printf '%s' "$artifact" | jq -e \
  --arg id "$artifact_id" --arg commit "$CANDIDATE_COMMIT" --arg run "$RELEASE_RUN_ID" '
  (.id | tostring) == $id and .name == ("v1.0.0-prepared-" + $commit) and .expired == false and
  (.workflow_run.id | tostring) == $run and .workflow_run.head_sha == $commit and
  (.digest | type == "string" and test("^sha256:[0-9a-f]{64}$"))' >/dev/null
gh api "repos/$repo/actions/artifacts/$artifact_id/zip" > "$work/prepared-artifact.zip"
expected_artifact_sha=$(printf '%s' "$artifact" | jq -er '.digest | sub("^sha256:";"")')
test "$(shasum -a 256 "$work/prepared-artifact.zip" | awk '{print $1}')" = "$expected_artifact_sha"
sh "$script_dir/check-stable-release-preflight.sh" --extract-artifact "$work/prepared-artifact.zip" "$work/assets"
assets="$work/assets"
sh "$script_dir/check-stable-release-preflight.sh" --verify-assets "$assets"
test "$(jq -er .candidate.pull_request "$assets/RELEASE-MANIFEST.json")" = "$(printf '%s' "$preflight" | jq -er .pull_request)"
for name in level7-dev-loop-1.0.0-claude.zip level7-dev-loop-1.0.0-codex.zip SHA256SUMS RELEASE-MANIFEST.json
do
  gh attestation verify "$assets/$name" --repo "$repo" \
    --signer-workflow "$repo/.github/workflows/release.yml" --signer-digest "$CANDIDATE_COMMIT" \
    --source-ref refs/heads/main --source-digest "$CANDIDATE_COMMIT"
done
final_preflight=$(sh "$script_dir/check-stable-release-preflight.sh" --publish "$artifact_id")
test "$final_preflight" = "$preflight"
request="$work/create-release.json"
jq -n --rawfile body "$project_root/docs/releases/v1.0.0.md" --arg commit "$CANDIDATE_COMMIT" '
{tag_name:"v1.0.0",target_commitish:$commit,name:"Level 7 Dev Loop v1.0.0 — UNSIGNED, NOT NOTARIZED",
 body:$body,draft:true,prerelease:false,generate_release_notes:false,make_latest:"true"}' > "$request"
tag_object=$(gh api --method POST "repos/$repo/git/tags" \
  -f tag=v1.0.0 -f message='Level 7 Dev Loop v1.0.0 — UNSIGNED, NOT NOTARIZED; qualification WITHHELD' \
  -f object="$CANDIDATE_COMMIT" -f type=commit)
RELEASE_TAG_OBJECT=$(printf '%s' "$tag_object" | jq -er .sha)
export RELEASE_TAG_OBJECT
gh api --method POST "repos/$repo/git/refs" -f ref=refs/tags/v1.0.0 -f sha="$RELEASE_TAG_OBJECT" >/dev/null
release=$(gh api --method POST "repos/$repo/releases" --input "$request")
RELEASE_DRAFT_ID=$(printf '%s' "$release" | jq -er 'select(.draft == true and .tag_name == "v1.0.0") | .id')
export RELEASE_DRAFT_ID
gh release upload v1.0.0 --repo "$repo" "$assets/level7-dev-loop-1.0.0-codex.zip" \
  "$assets/level7-dev-loop-1.0.0-claude.zip" "$assets/SHA256SUMS" "$assets/RELEASE-MANIFEST.json"
remote=$(gh api "repos/$repo/releases/$RELEASE_DRAFT_ID")
expected='["RELEASE-MANIFEST.json","SHA256SUMS","level7-dev-loop-1.0.0-claude.zip","level7-dev-loop-1.0.0-codex.zip"]'
test "$(printf '%s' "$remote" | jq -c '[.assets[].name] | sort')" = "$expected"
test "$(printf '%s' "$remote" | jq '[.assets[] | select(.state == "uploaded" and .size > 0)] | length')" = 4
mkdir -m 0700 "$work/staged"
gh release download v1.0.0 --repo "$repo" --dir "$work/staged"
sh "$script_dir/check-stable-release-preflight.sh" --verify-assets "$work/staged"
for name in RELEASE-MANIFEST.json SHA256SUMS level7-dev-loop-1.0.0-claude.zip level7-dev-loop-1.0.0-codex.zip
do
  cmp "$assets/$name" "$work/staged/$name"
done
sh "$script_dir/check-stable-release-preflight.sh" --finalize "$artifact_id"
published=$(jq -n '{draft:false,make_latest:"true"}' | \
  gh api --method PATCH "repos/$repo/releases/$RELEASE_DRAFT_ID" --input -)
test "$(printf '%s' "$published" | jq -r .draft)" = false
final=$(gh api "repos/$repo/releases/$RELEASE_DRAFT_ID")
test "$(printf '%s' "$final" | jq -r .immutable)" = true
test "$(printf '%s' "$final" | jq -r .tag_name)" = v1.0.0
test "$(printf '%s' "$final" | jq -r .prerelease)" = false
test "$(printf '%s' "$final" | jq -c '[.assets[].name] | sort')" = "$expected"
printf '%s\n' 'Published immutable stable v1.0.0: UNSIGNED, NOT NOTARIZED; protected evaluation and provider trials NOT_RUN; formal support WITHHELD.'
