# Level 7 v1 Stable Release — Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-stable-release` |
| Candidate commit | `d39109f268ea2ea742316a7710aa71c0bf2cc433` |
| Candidate tree | `ad1eb3cf18fc37dcb23a9de32cc3f922bd9484e6` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |
| Audited at | `2026-09-01T14:05:35Z` |
| Base commit | `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` |
| Brief commit | `19ab613894a0b8ae4a8524d208477f0388ec7d28` |
| Implementation commit | `8911e179f95165347946214420f93dbfa9a9f7dc` |
| Implementation tree | `1a2b5fa792def7950cf974d7f56f3c11f36f257d` |

## Decision and independence

`GO` is granted only for the exact verification candidate and tree above. The
approved brief, implementation, verification record, release workflow,
packager, validator, focused tests, controller state, rollback, and authority
boundaries were independently inspected. The approved nine-path
implementation scope is exact, the verification is bound to that
implementation, and no implementation byte follows verification.

Reviewer `apbusinessidentity-tech` is distinct from Product Owner
`addressanup`, implementer/verifier `codex-root`, and PR author
`anup19950725`. The repository approval envelope binds the current change,
brief, and implementer to an active Product Owner interaction. This record does
not claim a GitHub review, hosted result, owner GO, merge authority, release
authority, or publication authority.

## Independent evidence

| Area | Assessment |
|---|---|
| Topology and scope | PASS — base `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` is followed by the sole-path brief `19ab613894a0b8ae4a8524d208477f0388ec7d28`, the exact nine-path implementation `8911e179f95165347946214420f93dbfa9a9f7dc`, and the sole-path verification candidate. The implementation changes only `.github/workflows/release.yml`, `CHANGELOG.md`, `Makefile`, `README.md`, `cmd/l7pack/main.go`, `cmd/l7pack/main_test.go`, `docs/releases/v1.0.0.md`, `internal/harness/v1candidate/main.go`, and `internal/harness/v1candidate/main_test.go`. There are no deletions or undeclared implementation paths. |
| Stable identity boundary | PASS — development remains the default and release-blocked. Exact stable mode requires command-line `L7_CLI_VERSION=1.0.0` and `L7_PACKAGE_CHANNEL=stable`; an ambient stable version is rejected. Unsupported or mismatched version/channel pairs fail before packaging. |
| Package and catalog boundary | PASS — source and tests enforce the closed, sorted host-specific inventories, local self-contained marketplace catalogs, exact version/channel/host pairing, checksums, SPDX metadata, strict paths and sizes, and provenance limited to observed package inputs with `authority: external-only` and `release_blocked: true`. Candidate metadata contains no approval, signing, notarization, publication, or release-ready authority. |
| Validator and lifecycle | PASS — focused package tests exercise identity substitution, cross-host mismatch, catalog binding, unsafe content, strict metadata, upgrade, rollback, removal, native CLI, and MCP behavior using disposable roots. An independent clean Git archive test of `./cmd/l7pack` and `./internal/harness/v1candidate` passed offline with the pinned Go toolchain and read-only modules. |
| Deterministic unsigned artifacts | PASS — the retained verification artifacts match the recorded SHA-256 values, including stable Codex `6f67333bc1117acee1a3fc44a1f1b819e9a131b1446f32902596c87ae8fb733c` and stable Claude `9a0369ab68362fb29f9811d389fbf92f99141646f523145956f185941702351c`. Archive inspection confirmed the exact inventories, stable manifests, host catalogs, and blocked external-authority provenance. The heavyweight reproducibility run was not repeated. |
| Manual release workflow | PASS by static inspection — the workflow has only `workflow_dispatch`, fixed concurrency without cancellation, explicit time limits and minimal permissions, and five action uses pinned to full commit hashes. It binds commit and tree to the exact merged-main PR lineage, risk label, exact-head required checks including the unchanged benchmark, distinct owner and auditor approvals, protected signing/production environments, absent tag/release state, and immutable-release configuration. No retry-until-green, automatic trigger, `continue-on-error`, signing fallback, benchmark exception, or candidate-controlled acceptance path is present. `actionlint` passed. |
| Signing and preparation design | PASS by static inspection — preparation uses clean archived source and guarded temporary roots, compares two unsigned builds before signing, imports protected Apple material into a disposable keychain, signs all four Mach-O inputs with hardened runtime and secure timestamps, verifies packaged signatures, submits each host ZIP exactly once with `notarytool --wait`, requires `Accepted`, prepares exactly four bound assets, and attests them. Cleanup is trapped for success and failure; certificate and notary-key files are removed and keychain deletion is attempted before the runner exits. None of these external operations was run by this audit. |
| Provider and owner boundary | PASS by design — publication requires an exact-run, exact-attempt, exact-commit/tree and exact-asset provider-trial record for both hosts, a separate exact Product Owner authorization, and protected `v1-production` approval by `addressanup`. Candidate archives cannot satisfy these gates. Actual host/provider/model execution remains separately authorized external evidence. |
| Publication and rollback | PASS by static inspection — the production job revalidates lineage, checks, reviews, environments, attestations, manifest, digests, and the exact four-asset set before creating an absent annotated `v1.0.0` tag and draft release. It refuses existing or mismatched state, verifies uploaded bytes, publishes once, and enables immutability. Unexpected partial external state is not overwritten or silently repaired. |
| Protected controls | PASS — the harness workflow, trusted policy workflow, benchmark driver and comparator, controller policy, dependencies, and `AGENTS.md` are byte-identical to the base. No threshold, target, architecture, controller, or authority-policy weakening is in scope. |
| Controller and hygiene | PASS — a disposable offline controller check reported Tier 3 state `awaiting-independent-audit` and next transition `record the bound independent decision`. Candidate `git diff --check`, tracked worktree, and index were clean before this record. The required audit path was absent. |

## Findings and residual risks

No unresolved implementation findings were identified.

| Severity | Count |
|---|---:|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |

The following are residual external boundaries, not locally established facts:

- fresh exact-head hosted checks, trusted-policy evaluation, required GitHub
  reviews, and protected-environment configuration and approvals;
- native macOS amd64 execution, actual Codex and Claude marketplace/provider
  trials, provider network/model behavior, and disposable-host residue checks;
- Developer ID credential suitability, all four live signatures, Apple
  notarization acceptance, keychain cleanup outcome, artifact attestations, and
  exact final asset digests; and
- tag creation, immutable GitHub release publication, installation from public
  assets, and post-publication compatibility claims.

The signing trap deliberately tolerates a failed `security delete-keychain`
command so that early failures and subsequent credential-file cleanup are not
masked. The keychain remains confined to `RUNNER_TEMP` on an ephemeral hosted
runner, and no credential file enters the prepared artifact. A future hosted
run must still observe the cleanup and protected-secret boundary; this audit
does not convert the best-effort cleanup command into executed evidence.

The verification record truthfully marks every external item above `NOT_RUN`.
Predecessor checks, reviews, benchmarks, provider trials, or release artifacts
do not transfer to the future exact audit successor.

## Preservation and rollback

This audit preserves every historical record and all implementation bytes. The
original checkout's unrelated untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched
and unstaged at SHA-256
`9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

Before merge or publication, rollback is the ordinary reverse-order revert of
this audit record, verification `d39109f268ea2ea742316a7710aa71c0bf2cc433`,
implementation `8911e179f95165347946214420f93dbfa9a9f7dc`, and brief
`19ab613894a0b8ae4a8524d208477f0388ec7d28`, preserving history and confirming
the restored tree equals base tree
`9b6550c11e9666ab25166af6a5fb7560cdcc8cdf`.

A failed preparation must publish nothing. Any unexpected tag, draft, asset,
or hosted state must be reported and left untouched pending separate cleanup
authority. After an immutable public `v1.0.0` release, do not move, delete, or
replace that tag or its assets; correct defects only through a separately
approved patch release.

## Executable next transition

`codex-root` must validate that the committed audit is the direct child of the
exact candidate, changes only this audit path, and retains the bound `GO`
fields; it may then record the matching local audit envelope for reviewer
`apbusinessidentity-tech` and candidate
`d39109f268ea2ea742316a7710aa71c0bf2cc433`.

That transition creates no hosted review or result and authorizes no push,
merge, tag, signing, notarization, provider execution, release, publication,
installation, or deployment. Every later external step requires its separately
defined authority and exact-head evidence.
