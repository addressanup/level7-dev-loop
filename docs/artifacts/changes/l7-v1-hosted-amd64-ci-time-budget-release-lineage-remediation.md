# Level 7 v1.0 Hosted amd64 CI Time-Budget and Release-Lineage Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation` |
| Risk tier | `3` — protected CI timing control and unreleased v1.0.0 workflow lineage |
| Status | `proposed`; implementation requires Product Owner approval of this exact brief commit |
| Base commit | `c634e092b2f938ad3038be0632d162b2bdde41f3` |
| Base tree | `4a28b3ec2495566554cda8ab2462b3b41043b474` |
| Product lineage | PR #17 merge with parents `66777352538a514b990ffca8fa290ca6de9584fd` and `70a4b50d2fb71a4af51d7566fcec39938ae2bc01` |
| Post-merge hosted evidence | Harness `33592522261`, attempt 1, amd64 job `100129242555`; Trusted policy `33593520198` was the designed non-applicable push-event skip |
| Release state | `v1.0.0` tag and GitHub release absent; release workflow never dispatched |
| Proposal branch | `codex/l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation`, rooted directly at the exact base |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root` |
| PR author / release operator | `anup19950725` |
| Independent auditor | `apbusinessidentity-tech` |
| Hosted assurance | team mode; sole trusted PR label `l7-risk-tier-3` |
| Next executable transition | Commit only this brief, then stop for exact Product Owner approval |

## Problem

The exact PR #17 tree passed its pull-request amd64 job in 14 minutes 9
seconds, leaving only 51 seconds under the fixed 15-minute job limit. The
identical merged tree later exhausted that limit on `main`. Bootstrap and every
reported check through all eight fixed fuzz targets passed, but GitHub cancelled
`make ci` during reproducibility work before CLI reproducibility and distribution
checks completed. The later cross-build and native v1 lifecycle steps therefore
never ran.

This is a hosted Intel job-budget failure, not a test assertion failure and not
release evidence. An unchanged rerun would not repair the boundary. The current
release workflow also binds the future release merge's first parent to
`6677735...`; any honest successor merge based on current `main` would fail that
lineage gate unless the constant is rebound.

## Scope

After approval of this exact brief, implementation may change only two
non-evidence paths:

1. `.github/workflows/harness.yml` may make the existing macOS matrix carry a
   literal per-entry timeout. `arm64` remains exactly 15 minutes and only
   `amd64` changes to exactly 25 minutes. The job must consume that fixed matrix
   value without changing either runner, check name, step, command, ordering,
   architecture, target, fuzz budget, benchmark control, concurrency rule, or
   failure propagation.
2. `.github/workflows/release.yml` may change only `L7_RELEASE_BASE` from
   `66777352538a514b990ffca8fa290ca6de9584fd` to
   `c634e092b2f938ad3038be0632d162b2bdde41f3`. Every actor, topology, tree,
   check, review, environment, signing, notarization, provider-trial,
   attestation, absent-tag/release, owner-approval, asset, and publish-once gate
   remains unchanged.

The 25-minute Intel ceiling is fixed and bounded. The cancelled run was already
roughly 3 minutes 44 seconds behind the prior run by fuzz completion, and the
prior remaining reproducibility, distribution, cross-build, and lifecycle work
projects an approximately 18-minute total on that runner profile. A 20-minute
ceiling would restore less than two minutes of margin and retain the same
fragility.

No unchanged workflow rerun, product code, Makefile target, test, dependency,
threshold, architecture, release environment, secret, branch protection, tag,
release, provider execution, installation, or deployment is in scope.

## Exact implementation file set

Declared path count: 5 (3 Add, 2 Modify, 0 Delete).

Add:

- `docs/artifacts/changes/l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation-audit.md`

Modify:

- `.github/workflows/harness.yml`
- `.github/workflows/release.yml`

Delete:

- None.

## Acceptance criteria

1. The proposal commit is the direct child of exact base `c634e092...`, adds
   only this brief, and changes no implementation, prior record, remote, PR,
   review, check, protection, credential, environment, tag, release, or
   deployment state.
2. Fresh Product Owner approval names this change ID, the exact proposal commit,
   base/tree, implementer `codex-root`, the two non-evidence paths, and the exact
   Intel-only 25-minute boundary. No predecessor approval or release authority
   transfers.
3. Implementation is the direct child of the approved brief and changes exactly
   the two declared workflow paths. Verification and audit are later sole-path
   commits; every historical record remains byte-identical.
4. Harness retains the existing Ubuntu 15-minute jobs and paired benchmark
   20-minute job. The macOS matrix retains both current runners and every current
   step, with `arm64=15` and `amd64=25` as literal values. No skip, retry,
   `continue-on-error`, target reduction, or swallowed failure is introduced.
5. The release-workflow diff is exactly the one lineage-constant replacement.
   It accepts only a future two-parent merge whose first parent is
   `c634e092...` and whose second parent and tree match the reviewed PR head.
6. `actionlint` passes both workflows, and a semantic/diff inspection proves
   the timeout inventory, check names, runners, commands, release gates, and all
   out-of-scope bytes remain exact.
7. From a clean exact implementation candidate with adequate local storage,
   `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` and
   `make v1-candidate-check GO_VERSION=1.26.7` each run once and pass without
   reduced targets or retry-until-green.
8. One verification record binds PASS or FAIL to the exact implementation
   commit/tree and observed evidence. Only PASS may advance to a separately
   commissioned independent read-only audit by `apbusinessidentity-tech`, which
   binds GO or NO_GO to the exact verification successor.
9. A fresh exact-head PR carries only `l7-risk-tier-3` and passes baseline,
   shadow, arm64, amd64, the unchanged paired benchmark, and trusted policy.
   The amd64 job must reach and pass distribution, cross-build, and native v1
   lifecycle work; its duration and remaining timeout margin are recorded.
10. A later merge requires separate protected-control authority and must have
    parents `[c634e092..., exact PR head]`, tree equal to the PR-head tree, and a
    fully successful automatic post-merge Harness. The stopped handoff cannot
    be resumed; any release preparation requires a new exact staged authority.
11. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and unstaged at SHA-256
    `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

## Risks and mitigations

- **Timeout masks a hang:** change only the Intel ceiling, keep every fail-fast
  command and hard target budget, prohibit reruns, and record actual duration.
- **Budget remains fragile or becomes broad:** use the evidence-backed fixed
  25-minute Intel value while retaining every other timeout exactly.
- **Release lineage drifts:** bind the one new accepted first parent and leave
  every other release preflight and publication gate unchanged.
- **A green PR hides mainline failure:** require a fresh automatic post-merge
  Harness success before any new release-recovery authority.
- **Historical or user state is damaged:** isolate the branch, preserve all
  records, and leave the unrelated original-checkout file untouched.

## Rollback

Before implementation, revert or discard only this proposal commit. Before
merge, revert later audit and verification records first, then implementation
and brief, and confirm restoration of exact tree
`4a28b3ec2495566554cda8ab2462b3b41043b474`.

After a future remediation merge but before publication, rollback requires a
separately authorized ordinary merge revert. Never rewrite history, move or
replace `v1.0.0`, or delete unexpected external state. An immutable published
release, if later authorized, must be corrected only by a new reviewed release.

## Current transition

Commit only this brief as a direct child of exact base `c634e092...`, then stop
for explicit Product Owner approval of that exact proposal commit. That approval
must bind the change ID, base/tree, implementer, the two implementation paths,
and the Intel-only 25-minute ceiling. It grants no push, PR mutation, review,
workflow rerun, merge, release dispatch, provider trial, signing, notarization,
tag, publication, installation, deployment, or cleanup authority.
