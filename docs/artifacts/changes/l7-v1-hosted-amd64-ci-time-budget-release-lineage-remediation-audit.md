# Level 7 v1.0 Hosted amd64 CI Time-Budget and Release-Lineage Remediation - Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation` |
| Candidate commit | `a0e39b2fc8d6ad4a6d83cdec5fcbd079c2324c5e` |
| Candidate tree | `433c9c5929e59766b189d0d61f1593275542d22e` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |
| Audited at | `2026-09-02T07:46:16Z` |
| Implementation commit | `333ed7abec9d34eacc9cc8481b2194f4506db87a` |
| Implementation tree | `fa09d1581fcea377cedee133635e9ce534ddf682` |
| Approved brief | `4d56fe6729d679da3002ca86572c5e7d5838fb85` |
| Base commit | `c634e092b2f938ad3038be0632d162b2bdde41f3` |
| Base tree | `4a28b3ec2495566554cda8ab2462b3b41043b474` |

## Decision and independence

`GO` for only the exact verification successor commit and tree above. The
brief and implementer-run verification were treated as claims. Git identity and
topology, declared scope, workflow semantics, approval and evidence bindings,
preserved hosted failure, read-only reproducible identities, rollback, remote
state, and protected user state were inspected independently.

Product Owner Anup Pandey (`addressanup`) commissioned this one read-only audit
for reviewer `apbusinessidentity-tech`, GitHub User id `322178380`, distinct
from owner `addressanup` (User id `195550570`), PR author/release operator
`anup19950725` (User id `157838648`), and implementer/verifier `codex-root`.
The GitHub CLI identity was not switched and only GET requests were made. This
record is not a GitHub review, hosted result, owner `GO`, merge-readiness
decision, release authority, or publication authority.

## Requirement and evidence map

| Area | Independent assessment |
|---|---|
| Exact candidate | PASS - audit began from a clean branch head `a0e39b2fc8d6ad4a6d83cdec5fcbd079c2324c5e`, tree `433c9c5929e59766b189d0d61f1593275542d22e`. The designated audit path and matching external audit envelope were absent. |
| Authority and topology | PASS - the regular-file approval envelope has SHA-256 `e2f6a6a5bfc8694c7ee3bb28fe2e02f3447b00e177de7900b45493ba1d1431fb` and binds Product Owner Anup Pandey, implementer `codex-root`, this change ID, and brief `4d56fe6729d679da3002ca86572c5e7d5838fb85`. Exact base `c634e09...` is followed by the sole-path brief, two-path implementation `333ed7a...`, and sole-path verification record `a0e39b2...` in a direct linear chain. No prior authority or decision transfers. |
| Scope and history | PASS - base through candidate changes exactly four paths: the two permitted workflow files plus the brief and verification record, with no deletion or historical-record rewrite. The implementation commit changes only `.github/workflows/harness.yml` and `.github/workflows/release.yml`; `git diff --check` is clean. |
| Harness semantics | PASS - Ubuntu baseline/shadow jobs remain literal 15-minute jobs, the paired benchmark remains 20 minutes, and the macOS matrix retains `macos-15` arm64 at literal 15 and `macos-15-intel` amd64 at literal 25. The job consumes only the fixed matrix value. Names, runners, steps, commands, targets, ordering, concurrency, fail-fast behavior, and direct failure propagation are otherwise byte-identical to the base; no skip, retry, target reduction, `continue-on-error`, or swallowed failure was introduced. |
| Hosted failure lineage | PASS by GET-only GitHub evidence - PR #17 Harness run `33542744706` passed the identical tree; Intel job `99972747658` completed in 14m09s. Automatic main Harness run `33592522261`, attempt 1, was cancelled; Intel job `100129242555` exhausted the 15-minute ceiling after all eight fuzz targets passed, then skipped distribution completion, cross-build, and native lifecycle. The logs show fuzz completion roughly 3m43s later than the prior successful Intel run. Trusted-policy run `33593520198` was the expected push-event skip. No rerun is credited. |
| Time-budget proportionality | PASS - the prior successful Intel job needed about 3m10s after fuzz completion, while the failing same-tree run had already consumed about 14m15s at that boundary. A fixed 25-minute Intel-only limit provides bounded recovery margin without broadening another job or changing the workload. The new ceiling is still subject to fresh exact-head hosted proof. |
| Release lineage | PASS - the release workflow diff is exactly one constant replacement: `L7_RELEASE_BASE` changes from `66777352538a514b990ffca8fa290ca6de9584fd` to merged predecessor `c634e092b2f938ad3038be0632d162b2bdde41f3`. The workflow continues to require a future two-parent merge with that exact first parent, reviewed PR head as second parent, and merge tree equal to the PR-head tree. Actor, label, checks, reviews, environments, signing/notarization, provider-trial, owner authorization, absent-tag/release, exact-asset, attestation, and publish-once gates are byte-identical. |
| Implementer evidence | PASS within the read-only audit boundary - the verification record binds local `PASS` to implementation `333ed7a...`, tree `fa09d15...`, and truthfully distinguishes its initial missing-toolchain failure, evidence-path typo, authorized storage recovery, bootstrap, one replacement full verification, and one v1 candidate check. This audit did not rerun those commands. Current workflow digests match the record, and the retained A/B harness and CLI binaries match their claimed sizes and SHA-256 identities. |
| Reproducible outputs | PASS by read-only inspection - all four arm64/amd64 v1 input binaries and both candidate ZIPs match the recorded sizes and SHA-256 values. Both ZIPs pass archive integrity, and each embedded binary matches its source input and `CHECKSUMS.json` entry. These ignored local outputs corroborate, but do not replace, the recorded command results. |
| Remote and release boundary | PASS - live remote `main` remains exact base `c634e09...`; the remediation branch, associated PR, exact-candidate Actions runs, `v1.0.0` tag/release, and release-workflow dispatch remain absent. No push, review, merge, release, publication, installation, or deployment is credited. |
| Protected user state | PASS - the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remains a regular file, untouched and unstaged, at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Findings

No unresolved BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW finding remains within
the approved local remediation and audit boundary.

The workflow change repairs two specific stale controls without weakening the
workload: Intel receives additional bounded wall-clock budget, and a future
release preflight is rebound to the current merged first parent. The historical
timeout remains a cancelled failure and is not relabeled as success.

## Residual risks and claim boundary

- No hosted workflow has run on this exact candidate. The 25-minute ceiling has
  not yet been exercised on GitHub's Intel runner, so a fresh exact-head PR must
  prove that the amd64 job reaches and passes distribution, cross-build, and
  native v1 lifecycle and must record actual duration and margin.
- `actionlint`, full verification, fuzzing, v1 candidate checks, builds, and
  benchmarks were not rerun by this audit under its explicit read-only/no-check
  commission. The workflow diff and retained outputs were inspected directly;
  command execution remains implementer evidence until fresh hosted CI runs.
- Native amd64 behavior is established historically on the identical prior tree
  and is cross-built in the local record, but no native amd64 run exists for
  this exact workflow candidate.
- The local storage cleanup and bootstrap recovery cannot be reconstructed after
  the fact. Their failures and authorization stops are preserved in the bound
  record; current approval, toolchain-derived outputs, and workspace state are
  consistent with it.
- Release topology, reviews, environments, Apple secrets, provider trials,
  signing, notarization, assets, attestations, tag creation, and immutable
  publication have not been exercised for a future merge and remain blocked.
- The matching non-versioned audit envelope is intentionally absent from this
  commission. This tracked decision alone cannot advance controller state.

## Rollback and preservation

No remote or production effect exists. Before merge, rollback uses ordinary
history-preserving reverts in reverse order: this audit record, verification
`a0e39b2fc8d6ad4a6d83cdec5fcbd079c2324c5e`, implementation
`333ed7abec9d34eacc9cc8481b2194f4506db87a`, then brief
`4d56fe6729d679da3002ca86572c5e7d5838fb85`. Stop on conflict or any extra
path and confirm restoration of base tree
`4a28b3ec2495566554cda8ab2462b3b41043b474`.

After a separately authorized future merge, rollback requires an ordinary merge
revert under fresh authority while `v1.0.0` remains absent. Unexpected external
state must not be deleted. An immutable published release must never be moved or
overwritten; correction then requires a separately reviewed patch release.

## Next executable transition

After this sole-path audit successor and a separately created matching external
audit envelope validate as `reviewed`, stop for distinct Product Owner authority
to publish the exact audit head to a fresh Tier 3 PR. That PR must retain only
`l7-risk-tier-3` and earn fresh exact-head baseline, shadow, arm64, amd64, paired
benchmark, and trusted-policy success plus exact-head independent reviewer and
accountable-owner approvals before any separately authorized merge.

This decision authorizes no envelope creation, push, PR mutation or review,
workflow dispatch or rerun, merge, release dispatch, provider trial, signing,
notarization, tag, publication, installation, deployment, or cleanup.
