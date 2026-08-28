# Hosted CI Readiness Remediation — Independent Audit

| Field | Value |
|---|---|
| Change ID | `hosted-ci-readiness-remediation` |
| Candidate commit | `9007ff0dd4bd516bf2ea68b370511db5ec7fa643` |
| Candidate tree | `5fd541a21f2e55df968342975bc03c8e54ac22f4` |
| Result | `GO` |
| Reviewer | `anup19950725` |
| Audited at | `2026-08-28T15:15:25Z` |
| Verified implementation | `228af7bbd29e1836ff8db59b2aab256af0b9fb9f` |
| Implementation tree | `04283aa24cfb010d539afad18c2a8e4ccf439fdf` |
| Brief commit | `55e97c93b65c96f4f281ea090816b3205784cae5` |
| Prior audit commit | `7da1de520734fc284adc13e52f1c2b0b32473db6` |
| Prior audit tree | `d38820be45fa945957c3f8e27e7a0937f57a8ad8` |
| Base commit | `481adaaec967ac34b6b27cf78509b85d5c068abc` |
| Base tree | `d57a334696487b1d15557c9980e8a55c2dc4c930` |

## Decision

`GO` for a history-preserving audit-only identity successor.

The active user explicitly designated GitHub user `anup19950725` as the
independent auditor. That GitHub account exists and is distinct from PR author
`addressanup`, local accountable owner `accountable-owner`, and local
implementer `codex-root`.

A direct successor of prior audit commit
`7da1de520734fc284adc13e52f1c2b0b32473db6` may replace only the final-tree
contents of
`docs/artifacts/changes/hosted-ci-readiness-remediation-audit.md`. The prior
audit commit and its former reviewer binding remain immutable historical Git
evidence. The successor must not amend, rebase, delete, or rewrite that commit
and must add or modify no second path.

This replacement keeps exactly one audit-record pathname in the final tree and
exactly three Tier 3 artifact paths. The verified implementation remains
`228af7bbd29e1836ff8db59b2aab256af0b9fb9f`; verification commit
`9007ff0dd4bd516bf2ea68b370511db5ec7fa643` remains the audit candidate.

This decision does not transfer hosted evidence to a future successor head.
Required technical checks passed on old head `7da1de52…`; any replacement
commit creates a new exact head whose checks and reviews must be obtained
afresh. Trusted `evaluate` remains blocked because no accountable-owner GitHub
identity is configured or approved. Neither `addressanup` nor `anup19950725`
may be treated as that owner.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Exact base and preserved history | PASS — base `481adaaec967ac34b6b27cf78509b85d5c068abc` resolves to tree `d57a334696487b1d15557c9980e8a55c2dc4c930`. Brief `55e97c93…`, implementation `228af7bb…`, verification `9007ff0d…`, and prior audit `7da1de52…` form a direct history-preserving chain. A successor changes only the existing audit path and leaves every predecessor commit reachable. |
| 2. Hermetic test identity | PASS by source inspection and bound verification — only the test-owned `commit-tree` call has command-local `user.useConfigOnly=true`, `user.name=Level Seven`, and `user.email=l7@example.invalid`. Production Git code and the shared test helper remain unchanged. |
| 3. Production Git behavior unchanged | PASS — production repository discovery, snapshot, commit, merge, identity, and authorization blobs are byte-identical to the base. An audit-only successor cannot alter them. |
| 4. Exact Intel lock record | PASS — implementation added exactly the approved baseline Go `1.26.7` `darwin/amd64` record and no other lock-row change. |
| 5. Exact five-tuple control | PASS with the recorded deprecated-checker boundary — the checker requires the five approved unique tuples and rejects missing, duplicate, malformed, shifted, or extra rows. Verification exercised that block in the original valid foundation context. The deprecated whole-tree checker is not an active `make verify` or hosted `make ci` gate. |
| 6. Actual Intel bootstrap | PASS on old audit head `7da1de52…` — PR #2 run `33163760559`, job `98824175970`, checked out that exact head on `macos-15-intel`. The log records detached-signature verification with signing key `0E225917414670F4442C250DFD533C07C264648F`, a good Google signing-authority signature, and `bootstrap-go: installed authenticated development toolchain (baseline, darwin/amd64, ...)`. The subsequent `make ci` and declared cross-build completed successfully. This evidence remains bound to `7da1de52…` and is not exact-head evidence for a replacement commit. |
| 7. Historical records unchanged | PASS — historical digest manifests, provider records, and governance artifacts remain byte-identical to the base. Replacing the final-tree audit contents does not erase the old audit blob or commit. |
| 8. Workflow and protection preservation | PASS — workflows, action digests, permissions, runners, timeouts, baseline/shadow policy, benchmark logic, bootstrap, and `Makefile` remain unchanged. Main protection still requires baseline, macOS arm64, macOS amd64, benchmark, and `evaluate`, plus review protections. |
| 9. Exact audit-successor hosted checks | PASS only for old head `7da1de52…`; `NOT_CURRENT` for any replacement successor. On old head, `Go 1.26.7 (baseline)`, `CLI macOS 15 (arm64)`, `CLI macOS 15 (amd64)`, and `CLI paired benchmark gate` all succeeded. Nonblocking `Go 1.27.0 (shadow)` also succeeded. A replacement commit invalidates those checks for exact-head readiness and must receive a fresh complete run. |
| 10. Trusted external authority boundary | PASS as fail-closed behavior; readiness remains blocked — old-head trusted run `33163761635` failed with `AUTH-001 external accountable-owner approval is absent`. Repository variable `L7_ACCOUNTABLE_OWNER` is unset, PR #2 has no reviews, and its author is `addressanup`. Designating `anup19950725` as auditor neither constitutes an exact-head GitHub approval nor supplies an accountable owner. A future owner must be distinct from both `addressanup` and `anup19950725`. |
| 11. Bound offline verification | PASS on the sole implementer verification record — verification remains bound to implementation `228af7bb…`, tree `04283aa2…`, and reports all approved offline checks passing. No implementation path changed after verification. |
| 12. Independent audit | PASS — this fresh read-only audit remains bound to verification commit `9007ff0dd4bd516bf2ea68b370511db5ec7fa643`, tree `5fd541a21f2e55df968342975bc03c8e54ac22f4`, and records designated reviewer `anup19950725`. A committed successor is valid only with a new external audit envelope naming that same reviewer and the resulting exact successor commit. |
| 13. Scope, at-most-one-record rule, and artifact budget | PASS — the final base-to-successor diff contains the brief, verification record, existing audit-record path, and three implementation files. There is one brief path, one verification path, and one audit path. Modifying that audit path does not create a fourth artifact or a second final-tree audit record. Verification-to-successor changes remain confined to that single audit path. |
| 14. Claim and effect boundary | PASS — this audit performed no provider invocation, version/help probe, prompt/stdin, model session, implementation, build, test, retry, fallback, configuration mutation, GitHub review, merge, release, deployment, or publication. Historical provider evidence and no-model gate states remain unchanged. |

## Audit-only successor validity

The trusted controller’s final-state rules permit this direct successor:

- artifact budgeting operates on changed paths from the declared base, so the
  final tree still contains three permitted governance paths;
- the verification record’s addition commit remains
  `9007ff0dd4bd516bf2ea68b370511db5ec7fa643`;
- the audit record remains bound to that verification commit and its exact tree;
- all changes after verification remain confined to the one audit-record path;
- no implementation commit follows verification; and
- the successor audit envelope must bind its actor and audit commit to the new
  final state.

The current external audit envelope binds former actor
`l7-release-independent-auditor` to prior audit commit `7da1de52…`. It cannot
qualify the successor and must not be relabeled as if it had always named
`anup19950725`.

If separately authorized after the successor exists, a new envelope must contain:

- schema `1`;
- change ID `hosted-ci-readiness-remediation`;
- actor `anup19950725`;
- candidate commit `9007ff0dd4bd516bf2ea68b370511db5ec7fa643`;
- audit commit equal to the new exact successor; and
- source `independent-agent` or truthfully derived `trusted-ci`.

Its actor must remain distinct from the applicable approval-envelope owner and
implementer. The local approval envelope currently binds owner
`accountable-owner`, implementer `codex-root`, and brief `55e97c93…`, which are
both distinct from `anup19950725`.

For hosted policy, the implementer is PR author `addressanup`. A future trusted
owner must be a third identity distinct from `addressanup` and
`anup19950725`.

## Hosted evidence

PR `addressanup/level7-dev-loop#2` was open, mergeable, and blocked at old head
`7da1de520734fc284adc13e52f1c2b0b32473db6`.

| Old-head gate | Result |
|---|---|
| `Go 1.26.7 (baseline)` | `PASS` |
| `Go 1.27.0 (shadow)` | `PASS`; remains nonblocking |
| `CLI macOS 15 (arm64)` | `PASS` |
| `CLI macOS 15 (amd64)` | `PASS`, including authenticated Intel bootstrap |
| `CLI paired benchmark gate` | `PASS` |
| trusted `evaluate` | `FAIL` at `AUTH-001` |
| Exact-head GitHub reviews | None |

The successful technical results establish that the remediation fixed both
original hosted defects at old head `7da1de52…`. They do not survive as
exact-head readiness evidence after a new audit-only commit.

After any successor is committed and published, the following are required on
that exact new head:

1. fresh successful baseline, macOS arm64, macOS amd64, and paired benchmark
   checks;
2. truthful recording of the nonblocking shadow result;
3. an exact-head approval submitted by GitHub user `anup19950725`;
4. an exact-head approval by a separately configured accountable-owner GitHub
   user distinct from PR author `addressanup` and auditor `anup19950725`;
5. a trusted-CI audit envelope whose actor matches this record; and
6. successful trusted `evaluate`.

Designation, this record, old-head checks, or a local envelope cannot substitute
for those hosted facts.

## Provider and control preservation

Historical Codex Gate 1 remains `PASS` only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`.

Historical Claude Gate 2 remains `NO_GO`: both exact implementer and reviewer
help invocations succeeded, both unknown-option parser controls unexpectedly
exited successfully, both invalid `--max-turns not-an-integer` controls failed
as required, and neither help surface advertised `--max-turns`. Help
advertisement remains non-dispositive; the successful unknown-option controls
remain dispositive.

The audited Codex and Claude no-model parser gates remain `NOT_RUN`.
Unknown-option rejection, typed `--max-turns` enforcement, argument
construction, permissions, output schemas, cancellation, cleanup, reviewer
immutability, scope, and containment remain unchanged.

## Rollback proof

| State | Required reverse sequence | Independent result |
|---|---|---|
| Pre-audit verification `9007ff0d…` | Verification, implementation, then brief | PASS — direct-parent reversal restores exact base tree `d57a334696487b1d15557c9980e8a55c2dc4c930`. |
| Prior-audit head `7da1de52…` | Prior audit, verification, implementation, then brief | PASS — reverting the prior audit deletes the sole audit path before removing its verification and implementation lineage. |
| Audit-identity successor | Replacement audit, prior audit `7da1de52…`, verification, implementation, then brief | PASS — reverting the replacement first restores the prior audit contents; reverting `7da1de52…` then removes the audit path. The remaining reverse sequence restores the exact base tree without orphaning an audit or verification record. |

Reverting only the replacement audit is not a complete rollback: it restores the
old audit record. A full post-successor rollback must therefore include both
audit commits in newest-to-oldest order, use ordinary revert commits, fail
closed on conflict or extra paths, and verify exact final tree
`d57a334696487b1d15557c9980e8a55c2dc4c930`.

Remote visibility, PR state, checks, reviews, variables, protection rules, and
other hosted evidence are outside repository-tree rollback.

## Findings and severity

`AUD-ID-001` — `CLOSED BY THIS REPLACEMENT`. The prior final-tree audit used
reviewer `l7-release-independent-auditor`, which cannot match the designated
GitHub reviewer login `anup19950725` in a trusted exact-head audit envelope.
Replacing only the existing audit record corrects the final-tree identity while
preserving the prior audit commit.

| Severity | Closed | Unresolved |
|---|---:|---:|
| BLOCKER | 1 | 0 |
| CRITICAL | 0 | 0 |
| HIGH | 0 | 0 |
| MEDIUM | 0 | 0 |
| LOW | 0 | 0 |

Hosted readiness remains blocked by the absent accountable-owner identity,
missing exact-head reviews, and the need for fresh successor-head checks. Those
are external transition gates, not unresolved implementation defects.

## Read-only boundary

This audit used read-only Git object, tree, ancestry, diff, source, policy,
artifact, authority-envelope, worktree-status, GitHub PR, check-run, workflow-log,
review, repository-variable, user, collaborator-permission, and branch-protection
inspection.

It did not edit files, index, refs, envelopes, configuration, remotes, PR
content, reviews, variables, collaborators, protection rules, or other GitHub
state. It did not run tests, builds, `make`, the controller, providers,
version/help probes, prompts/stdin, model sessions, CI, retries, fallbacks,
merge, release, deployment, or publication.

The remediation worktree and index remained clean. Primary-worktree status
reported only
`docs/artifacts/foundation-rebaseline-admission-audit.md` as untracked; its
contents were not inspected or touched.

## Authorization boundary

This record is returned as text only. It does not authorize creating or
committing the successor, replacing an authority envelope, pushing a branch,
updating PR #2, triggering or rerunning CI, requesting or submitting a GitHub
review, configuring an accountable owner, changing collaborator permissions or
branch protection, merging, releasing, deploying, or publishing.

Any materialization or hosted transition requires separate explicit
authorization bound to the resulting exact Git head.
