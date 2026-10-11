# Release Finalize Reads the Staged Draft by ID — Change Brief

| Field | Value |
|---|---|
| Change ID | `release-finalize-draft-by-id` |
| Risk tier | `3` — publication control path of the stable release |
| Status | implemented after the v1.1.0 publication exposed the defect; pull request open |
| Base commit | `72e6c1010b80a214d07166f748bb73693759b392` |
| Accountable owner | `addressanup` |
| Assurance | `solo`; truthful self-review, no independent audit |
| Authority | repository-local fix and pull request; no release effect |

## Problem

`publish-unsigned-stable.sh` creates the annotated tag and a draft release,
uploads and re-verifies the assets, then runs the preflight `--finalize`
phase before flipping the draft live. That phase read the staged release
with `GET /releases/tags/<tag>`, which GitHub does not serve for drafts, so
finalize failed closed with `release control read failed: gh` after every
verification had passed. During the v1.1.0 publication the owner authorized
completing the remaining steps by release ID with the same assertions; the
offline fake forge had served the draft by tag and therefore never caught it.

## Scope

In the finalize phase, require `RELEASE_DRAFT_ID` (the publisher already
exports it) and read `GET /releases/<id>`; `validate()` keeps binding that
release to the expected ID, draft state, tag name, target commit, and notes
body. The fake forge now refuses to serve a draft by tag name, like GitHub,
and serves `releases/<id>` only for the staged draft, so the end-to-end
publisher regression fails with the old read and passes with the new one.
No other phase, control, or publication behavior changes.

## Exact implementation file set

Add:

- `docs/artifacts/changes/release-finalize-draft-by-id.md`

Modify:

- `scripts/harness/check-stable-release-preflight.py`

Delete: none.

## Acceptance criteria

1. `make stable-release-preflight-check` passes, and the publisher
   end-to-end regression fails if the finalize read is reverted to the tag
   lookup.
2. Finalize still rejects a release whose ID, draft flag, prerelease flag,
   tag name, target commit, or body drifts from the staged candidate.

## Risks and mitigations

The fix touches only a read in the last verification before publication;
all assertions on the returned object are unchanged and covered by the
offline regression suite. It cannot create, modify, or delete remote state.

## Rollback

Revert the commit; the previous behavior fails closed before publication,
exactly as observed for v1.1.0.
