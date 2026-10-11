# Owner-Authorized Unsigned Stable Release v1.1.0 — Change Brief

| Field | Value |
|---|---|
| Change ID | `unsigned-stable-release-1-1-0` |
| Risk tier | `3` — release policy, package identity, and irreversible publication |
| Status | owner selected `1.1.0`, unsigned stable with all four gates re-waived for this version, and publication under the owner's own login on 2026-10-11 |
| Base commit | `111ab58776d863c7252aae462db6211794f460ee` |
| Accountable owner | `addressanup`, the actual authenticated repository owner |
| Assurance | `solo`; truthful self-review, no independent audit |
| Authority | prepare, merge, dispatch, and publish unsigned stable v1.1.0 under the owner's explicit per-version waiver decision; no hosted protection setting changes and no modification of existing releases |

## Problem

The merged `l7 plugin` installer is only reachable through a published
release, and every part of the stable release path is pinned to `1.0.0`:
package identity, the preparation workflow, the preflight, the publisher, and
the documentation. The preflight also requires the target tag and release to
be absent, so v1.0.0 cannot be reused. Signing, notarization, protected
holdout evaluation, and provider trials are still unavailable, and the
repository forbids carrying v1.0.0's waivers forward implicitly.

## Scope

Re-pin the release path to `1.1.0` with one version value per file so the
next bump is a one-line change: `releaseVersion` in `l7pack` and
`v1candidate`, `L7_RELEASE_VERSION`/`L7_RELEASE_TAG` and the dispatch
acknowledgement in `release.yml`, `VERSION` in the preflight, and `version` in
the publisher. Add the v1.1.0 release notes, record the release in the
changelog, and point the README's current-release references at v1.1.0 while
keeping v1.0.0, v1.0.0-dev, and v0.1.1 as unchanged history. Policy text,
waiver and qualification vocabulary, attestation checks, and publication
controls are unchanged; the owner's fresh explicit waiver authorization for
v1.1.0 is posted on the merged pull request, not encoded in candidate bytes.

## Exact implementation file set

Add:

- `docs/artifacts/changes/unsigned-stable-release-1-1-0.md`
- `docs/releases/v1.1.0.md`

Modify:

- `.github/workflows/release.yml`
- `CHANGELOG.md`
- `README.md`
- `cmd/l7pack/main.go`
- `cmd/l7pack/main_test.go`
- `internal/harness/v1candidate/main.go`
- `internal/harness/v1candidate/main_test.go`
- `scripts/harness/check-stable-release-preflight.py`
- `scripts/harness/publish-unsigned-stable.sh`

Delete: none.

## Acceptance criteria

1. Stable packaging accepts exactly `1.1.0`/`stable` (and the unchanged
   `1.0.0-dev` development default); `make ci` passes, including the
   preflight's offline regressions and the bootstrap check.
2. The preparation workflow, preflight, and publisher agree on `v1.1.0`
   names, the acknowledgement string, the artifact name, and the release
   notes digest, with no remaining `1.0.0` literal in the release path.
3. Publication still requires exact squash lineage, passing exact-head and
   post-merge checks, one Tier 3 label, the owner's raw waiver authorization
   comment for this exact candidate, immutable-release controls, and a single
   dispatch; the published release is immutable, non-prerelease, and titled
   UNSIGNED, NOT NOTARIZED with all qualification NOT_RUN and support WITHHELD.
4. After publication, `l7 plugin status --check` resolves v1.1.0 as the newest
   stable release and `l7 plugin update --all` can move a managed host to it.

## Risks and mitigations

Waived signing and qualification lower assurance by the owner's explicit
per-version choice; the notes, manifest, README, and CLI say so. A wrong pin
cannot publish: the preflight re-derives every name from one value and fails
closed on any drift. Publication is one-shot and verified before and after.

## Rollback

Before dispatch, revert this change normally. After a tag or draft exists,
stop without deleting, retrying, or moving anything and obtain separate
cleanup authority. After immutable publication, correct defects with a
separately reviewed release; v1.0.0 remains installable and pinnable
(`l7 plugin install --version 1.0.0`).
