# Level 7 v1.0 Hosted Bootstrap Transport Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-bootstrap-transport-remediation` |
| Risk tier | `3` — authenticated toolchain transport and unreleased v1.0.0 workflow lineage |
| Status | `proposed`; implementation requires Product Owner approval of this exact brief commit |
| Base commit | `66777352538a514b990ffca8fa290ca6de9584fd` |
| Base tree | `055583cef8181be59405443c2bb0ee14fc5e7690` |
| Integration target | current remote `main` at the same exact commit and tree |
| Product lineage | PR #16 merged with parents `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` and `b28f1cb036ff3014915f36f6894bc3f4e22b5c17` |
| Proposal branch | `codex/l7-v1-hosted-bootstrap-transport-remediation`, rooted directly at the exact base |
| Failed hosted evidence | Harness `33526156162`, attempt 1 arm64 job `99917396518` and attempt 2 arm64 job `99923803648`; Trusted policy `33527633971` and `33528069463` skipped for the main-push event |
| Release state | `v1.0.0` tag and GitHub release absent; release workflow not dispatched |
| Remote boundary | proposal branch remains local; no push, PR, review, check, or hosted-setting mutation is authorized |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root` |
| PR author / release operator | `anup19950725` |
| Independent auditor | `apbusinessidentity-tech` |
| Hosted assurance | team mode; sole trusted PR label `l7-risk-tier-3` |
| Next executable transition | Stop for explicit Product Owner approval bound to this exact proposal commit |

## Problem

The exact merged v1.0.0 candidate passed its PR Harness, including both macOS
architectures, but its automatic main Harness failed twice on macOS arm64
before any offline verification, cross-build, or host lifecycle test ran. Both
attempts ended in `make bootstrap-ci GO_VERSION=1.26.7` with `curl` status 56
and an HTTP 500 while fetching the pinned
`https://go.dev/dl/go1.26.7.darwin-arm64.tar.gz` archive. Attempt 1
was the original main-push job; the single separately authorized failed-job
rerun produced the same terminal result on a new runner. Baseline, shadow, and
macOS amd64 passed attempt 1. The paired benchmark was skipped on main push by
design, and both Trusted policy workflow-run jobs were skipped because policy
evaluates Harness completion only for pull-request events.

The bootstrap already retries selected HTTP responses, but it did not recover
from the observed receive failure. It also gives `curl` the final cache path
while a transfer is in progress. Repeating hosted jobs until one happens to
pass is not remediation, and the one authorized retry is exhausted.

The same bootstrap is used by the inert Harness and the manual stable-release
prepare job. A corrected candidate must be merged by a new PR on top of
`6677735…`; the current release workflow instead requires the earlier PR #16
base as the immediate first parent and would reject that successor. The
smallest viable remediation therefore makes transfer retry and cache commit
bounded and deterministic, tests that boundary offline, and rebinds only the
exact unreleased workflow lineage to this merged predecessor.

The approved stable-release brief `19ab613894a0b8ae4a8524d208477f0388ec7d28`,
implementation `8911e179f95165347946214420f93dbfa9a9f7dc`, verification
`d39109f268ea2ea742316a7710aa71c0bf2cc433`, audit
`b28f1cb036ff3014915f36f6894bc3f4e22b5c17`, hosted approvals, and owner GO
remain historical evidence for their exact bytes. None transfers to this
remediation or authorizes a release effect.

## Scope

After approval of this exact brief, implementation may change only four
non-evidence paths:

1. `scripts/harness/bootstrap-go.sh` may replace the implicit transfer retry
   with a literal maximum of four total attempts inside one bootstrap call.
   Only curl timeout status 28, receive status 56, and status 22 paired with
   final HTTP 408, 429, 500, 502, 503, or 504 may retry. Delays are fixed at
   1, 2, then 4 seconds, each locked file's aggregate transfer boundary remains
   at most 600 seconds, and no environment or candidate input may alter the
   policy.
2. Each attempt must write to a unique, same-directory, non-symlink temporary
   file. Only a successful transfer may atomically install the absent final
   cache path. Failure or signal removes the temporary file, preserves a
   pre-existing regular cache entry, reports the terminal status, and exits
   nonzero. There is no nested curl retry, blanket all-error retry, alternate
   mirror, swallowed error, or workflow retry.
3. The existing HTTPS host restrictions, TLS floor, locked URL, size and
   SHA-256 checks, exact signing fingerprints, detached-signature validation,
   archive-member containment, version check, disposable roots, and offline
   post-bootstrap execution remain mandatory and byte-equivalent in effect.
4. A new `scripts/harness/check-bootstrap-go.sh` must exercise the transport
   boundary without external network access using disposable fake-curl
   fixtures: receive failure then success, persistent receive failure, timeout,
   retryable HTTP status, non-retryable HTTP/TLS/local-write failure, exact
   attempt and delay counts, atomic installation, partial cleanup, pre-existing
   output preservation, symlink refusal, and ambient-policy rejection.
5. `Makefile` may add the focused check as a phony target and a dependency of
   the existing technical lint/verification path. It may not change any other
   target inventory, toolchain, timeout, fuzz, benchmark, or release command.
6. `.github/workflows/release.yml` may change only the exact release-base
   binding from `e88b18e…` to `6677735…`. The successor release candidate must
   still be the exact two-parent merge of that base and one reviewed PR head,
   with identical head/merge trees and all existing actor, label, check,
   review, environment, signing, provider-trial, tag, asset, attestation, and
   publication gates.

Harness and policy workflows, toolchain and signing-identity locks, archive
URLs and digests, dependencies, production code, benchmark driver/comparator,
10% threshold, five-sample minimum, architecture and target inventories,
branch protection, release environments, credentials, tags, releases, and all
historical records are outside implementation scope. A later benchmark
regression remains blocking unless `addressanup` makes the existing separate
exact-head GitHub decision; such acceptance is not implementation authority
and cannot bypass trusted policy.

## Exact implementation file set

Declared path count: 7 (4 Add, 3 Modify, 0 Delete).

Add:

- `docs/artifacts/changes/l7-v1-hosted-bootstrap-transport-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-bootstrap-transport-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-bootstrap-transport-remediation-audit.md`
- `scripts/harness/check-bootstrap-go.sh`

Modify:

- `.github/workflows/release.yml`
- `Makefile`
- `scripts/harness/bootstrap-go.sh`

Delete:

- None.

## Acceptance criteria

1. This proposal commit has exact sole parent/base `6677735…`, adds only this
   brief, and changes no implementation, prior record, envelope, worktree,
   remote, PR, check, protection, credential, tag, release, or deployment.
2. Fresh external Product Owner approval names this change ID, exact proposal
   commit, base/tree, implementer `codex-root`, and the four non-evidence paths.
   No prior approval, verification, audit, review, GO, or hosted result
   transfers.
3. Implementation is a direct descendant of the approved brief and changes
   exactly the four declared non-evidence paths. Verification and audit are
   later single-purpose commits; every historical record remains immutable.
4. The download policy performs at most four total curl invocations, uses only
   the literal retry statuses and delays in Scope, and stops each locked file's
   attempts within the fixed aggregate 600-second boundary. A non-retryable or
   exhausted failure exits nonzero without continuing to authentication,
   extraction, or tests.
5. No attempt writes the final cache path directly. Success installs one
   complete file atomically; failure and signals leave no temporary file or new
   final path. Existing paths and symlinks retain the current fail-closed
   behavior and are never overwritten.
6. All locked transport and authentication controls remain effective. A
   downloaded byte is never accepted without exact size, digest, signature
   identity, member-path, and toolchain-version validation; no fallback host or
   candidate-controlled retry/timeout input exists.
7. The focused offline check deterministically proves every named success and
   failure fixture, exact attempt/delay counts, terminal status propagation,
   atomicity, cleanup, and environment isolation on Darwin-compatible and
   Linux shell semantics. It performs no external request.
8. The Makefile retains every current verification target and adds only the
   focused bootstrap check. `sh -n` passes for every harness script, and
   `actionlint .github/workflows/release.yml` passes.
9. The release workflow differs only in its exact base binding and accepts only
   a later merge whose parents are `6677735…` and the exact remediation PR
   head. Manual dispatch, `anup19950725`, PR/tree/check/review validation,
   protected environments, signing/notarization, provider evidence, absent
   `v1.0.0` state, owner production approval, and publish-once behavior remain
   fail closed.
10. Local verification runs the focused check, actionlint,
    `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7`, and
    `make v1-candidate-check GO_VERSION=1.26.7` once from clean disposable
    roots. Any failure is recorded and stops; there is no retry-until-green.
11. A fresh PR for the exact final audit head uses sole label
    `l7-risk-tier-3` and passes baseline, shadow, macOS arm64, macOS amd64, the
    unchanged paired benchmark, and trusted policy. Local or predecessor
    results cannot substitute for exact-head hosted evidence.
12. If implementation verification passes, one verification commit adds only
    this change's verification record. Only after separate commissioning may
    `apbusinessidentity-tech` add one independent read-only audit record. A
    later exact-head owner approval and owner GO remain separately required
    before merge.
13. Release workflow dispatch, actual host/provider trials, provider network or
    model use, Apple credential/environment configuration, signing,
    notarization, tagging, publication, installation, and deployment remain
    unauthorized and `NOT_RUN`.
14. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and unstaged at SHA-256
    `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

## Risks and mitigations

- **Retry hides a permanent trust failure:** use a closed status allowlist,
  four total attempts, fixed delays, one aggregate limit, and immediate failure
  for certificate, protocol, local-write, lock, digest, or signature errors.
- **Partial cache is mistaken for authenticated input:** download beside the
  destination, atomically install only success, then retain every independent
  size, digest, signature, member, and version check.
- **Release lineage is broadened:** bind one exact merged predecessor and keep
  the existing two-parent topology, exact-tree, review, check, environment, and
  external-owner gates unchanged.
- **A green rerun is mistaken for remediation:** preserve both failed attempts,
  prohibit workflow retries, and require fresh exact-head PR evidence for the
  changed bytes.
- **Stale release authority:** changed bootstrap and workflow bytes invalidate
  predecessor verification, audit, approvals, and GO; require the full fresh
  Tier 3 sequence before any release effect.
- **User or historical-state damage:** use an isolated worktree, exact
  pathspecs, ordinary additive commits, and repeated fingerprint checks of the
  unrelated original-checkout file.

## Rollback

Before implementation, ordinarily revert only this proposal commit. After
implementation but before merge, revert later audit and verification records
first, then the implementation and brief, and confirm restoration of exact
base tree `055583cef8181be59405443c2bb0ee14fc5e7690`.

If a later remediation PR is merged but v1.0.0 remains unpublished, use a
separately authorized history-preserving revert of that merge and confirm main,
workflow lineage, and tree restoration. A failed release preparation must
publish nothing and must leave unexpected external state untouched pending
separate cleanup authority. An immutable published release must never be moved
or overwritten; correction then requires a separately approved patch release.

## Current transition

Commit only this brief as a direct child of exact base `6677735…`, then stop
for explicit Product Owner approval of that exact proposal commit. Only that
fresh approval may authorize `codex-root` to record the matching local envelope
and implement, locally test, and commit the four declared non-evidence paths.
Verification, independent audit, push, PR mutation or review, hosted execution
or rerun, owner GO, protected-branch merge, release workflow dispatch, provider
trials, signing, notarization, tagging, publication, installation, and
deployment remain separately gated and unauthorized.
