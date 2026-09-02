# Level 7 v1.0.0-dev Explicitly Unsigned (No Developer ID) Prerelease — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-explicitly-unsigned-prerelease` |
| Risk tier | `3` |
| Approved brief commit | `92d750c2d3be83321f65c3ed9c9f0ce4f9dc50e7` |
| Base commit | `5c23038e38a07b4f91f8ef38bbf163e061857910` |
| Base tree | `240bc9509a02a1a71616b62e88068ed7783bc65b` |
| Candidate commit | `71a4b7cd60ac6545852dcb2156897d42411a8025` |
| Candidate tree | `1b5556148bea1e89f5bce0a70574145a408c2827` |
| Result | `FAIL` |
| Verification scope | Local technical verification; pinned bootstrap network only |
| Reviewer | `codex-root` (implementer-run verification) |
| Failed at | `2026-09-02T16:12:10Z` |
| Local host | `darwin/arm64` (macOS `26.5.2`) |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Authorization, lineage, and scope | PASS — active-user authorization from Product Owner Anup Pandey (`addressanup`) continued `l7-release` for candidate `71a4b7c…` / tree `1b55561…`. The candidate is the direct child of approved brief `92d750c…`; its only paths are `.github/workflows/release.yml`, `.github/workflows/unsigned-prerelease.yml`, `CHANGELOG.md`, `README.md`, and `docs/releases/v1.0.0-dev.md`. The tracked worktree and index were clean before verification |
| Authorized storage recovery | PASS — the exact non-symlink, regenerable Xcode DerivedData leaf `/Users/anuppandey/Library/Developer/Xcode/DerivedData/laalee-goneifujaobgbbfdzbylbzthnzit` contained no nested Git repository and measured `7,374,760 KiB`. The force-delete form was rejected by the safety layer before execution; the same exact target was then permanently removed with non-forced recursive deletion under separate active-user authority. Free space became `7,378,560 KiB`, above the 6.5 GiB prerequisite floor. No other cache or user path was removed; recovery is by Xcode regeneration only |
| Pinned bootstrap | PASS — exactly one `make bootstrap-ci GO_VERSION=1.26.7` ran from `2026-09-02T16:11:34Z` through `16:11:50Z` and exited `0`. GPG verification accepted Google signing subkey `0E225917414670F4442C250DFD533C07C264648F`; the authenticated Darwin/arm64 Go `1.26.7` toolchain was installed in the isolated worktree, all modules verified, and downstream module access was prepared offline |
| Post-bootstrap preflight | PASS — free space was `6,976,544 KiB`, above the 6.1 GiB floor; the Go executable reported `go1.26.7 darwin/arm64`; candidate commit, tree, parent, tracked worktree, and index remained exact and clean |
| Workflow lint | **FAIL** — the sole invocation `actionlint .github/workflows/release.yml .github/workflows/unsigned-prerelease.yml` ran from `2026-09-02T16:12:07Z` through `16:12:10Z` and exited `1`. ShellCheck reported `SC2129` at the run block beginning on workflow line 199 because seven `$GITHUB_OUTPUT` writes use individual redirects, plus four `SC2016` findings at the run block beginning on line 861 because single-quoted `printf` format strings intentionally contain literal Markdown backticks. Acceptance criterion 15 therefore did not pass |
| `git diff --check` | `NOT_RUN` — the sequence stopped at the first new failure |
| Semantic workflow verification | `NOT_RUN` as a formal verification result; prior implementation-time read-only review cannot replace the failed required lint gate |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | `NOT_RUN`; its one authorized attempt was not consumed |
| `make v1-candidate-check GO_VERSION=1.26.7` | `NOT_RUN`; it was conditional on all preceding checks and full verification passing |
| Workspace and external boundary | PASS — the failed command did not alter implementation bytes, the tracked candidate remained clean before this record, and generated prerequisites remained confined to ignored `.cache/`. The original checkout's unrelated foundation audit remained untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. No push, PR, hosted run, review, merge, repository administration, environment or secret change, dispatch, artifact, tag, release, publication, installation, or deployment occurred |

## Candidate file identities

| Path | SHA-256 |
|---|---|
| `.github/workflows/unsigned-prerelease.yml` | `a3163b501a4b2536213e21673c92dcbb0e1a4399672d7ab4926a405ca59a130a` |
| `docs/releases/v1.0.0-dev.md` | `9a288d082f50e810b6e0834e473d510451182b1f6ca17d9d7f75b2a7702a7fa9` |
| `.github/workflows/release.yml` | `669bf0db8c400f2a543c76a256dd048f1202327561c3f9757aacaffde6883f51` |
| `CHANGELOG.md` | `45a2272ec68372c958ec64478b7c1d59fc3e218270bc4b12d9abfc4c7a3d5614` |
| `README.md` | `80469a189a7fdc98e24f63aa8513d159c49f67dd104196d72596d6b68ed12b29` |

No binary, package, manifest, checksum, attestation, or hosted artifact identity
was produced because the sequence stopped before full verification.

## Verification boundary

| Boundary | State |
|---|---|
| Full local verification and v1 candidate check | `NOT_RUN` |
| Prerelease workflow dispatch and its archived-source builds | `NOT_RUN` |
| Runtime signature classification, manifest, attestations, and artifact upload | `NOT_RUN` |
| Actual Codex and Claude host lifecycle, provider/model execution, and macOS amd64 native execution | `NOT_RUN` |
| Hosted Harness, policy, PR, reviews, merge, and protected approval | `NOT_RUN` |
| Apple credentials, signing, hardened-runtime verification, and notarization | `NOT_RUN` |
| Push, tag, GitHub release, publication, production installation, or deployment | `NOT_RUN` |

This implementer-run `FAIL` is bound only to candidate `71a4b7c…` / tree
`1b55561…`. It is not an independent audit, hosted-readiness decision, owner
GO, merge authority, release authority, or publication authority. It cannot
advance to audit. Any implementation-byte change invalidates this record.

## Rollback and next transition

No remote or release state exists from this verification. A local rollback
reverts this record, then implementation commit `71a4b7cd60ac6545852dcb2156897d42411a8025`,
then brief commit `92d750c2d3be83321f65c3ed9c9f0ce4f9dc50e7`, and confirms exact base tree
`240bc9509a02a1a71616b62e88068ed7783bc65b`; history remains preserved.

The only next Level 7 transition is a separately approved `l7-change`
remediation of the recorded `actionlint` findings. No retry, implementation
edit, audit, push, PR, merge, release, publication, installation, deployment,
or additional cleanup is authorized by this record.
