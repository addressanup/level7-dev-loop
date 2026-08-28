# Provider No-Model Gate Harness — Mainline Independent Audit

| Field | Value |
|---|---|
| Change ID | `provider-no-model-gate-harness-mainline` |
| Candidate commit | `5d0c84e6218dba0496a02dce019bd99f11a617f1` |
| Candidate tree | `e37f4d9a68bb19ef472c8823a4c30d8b0d6be41a` |
| Result | `GO` |
| Reviewer | `anup19950725` |
| Audited at | `2026-08-28T17:54:34Z` |
| Verified implementation | `c39cc1ddf783a8b785eb8a0bb6377de083ffd04f` |
| Implementation tree | `b9b22c203e91b2bd59855e25dda54abca134c20f` |
| Brief commit | `dff13dceb481867844fb6f26f087367e7d36849d` |
| Prior invalid proposal | `a10759533e8a11201f29393698f1de71a8b4a6bf` |
| Base commit | `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` |
| Base tree | `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6` |

## Decision

`GO` for the exact local verification successor
`5d0c84e6218dba0496a02dce019bd99f11a617f1`, tree
`e37f4d9a68bb19ef472c8823a4c30d8b0d6be41a`.

The approved corrected brief, preserved invalid proposal, exact implementation
blobs, verification binding, authority, immutable production boundary,
fake-runner behavior, actual-host containment, historical provider facts,
artifact budget, post-verification state, and complete rollback were
independently inspected. All eleven acceptance criteria pass.

This decision validates only the test-only no-model harness on current mainline.
It does not run either actual-host gate, qualify or promote a provider version,
establish hosted exact-head readiness, or authorize publication, merge, release,
deployment, installation, or a provider-support claim.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Exact base and preserved proposal history | PASS — base `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` resolves to tree `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`. Invalid proposal `a10759533e8a11201f29393698f1de71a8b4a6bf` is its direct child; corrected brief `dff13dceb481867844fb6f26f087367e7d36849d` is the direct child of that proposal. The corrected commit deletes only the invalid brief path and adds the controller-conformant brief. Base-to-corrected-brief diff contains only `docs/artifacts/changes/provider-no-model-gate-harness-mainline.md`, while both proposal commits remain immutable ancestors. |
| 2. PR #1 preservation | PASS — remote branch `provider-compatibility-no-model-gate-harness` still resolves to head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`, tree `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb`. PR #1 remains open and retains its four original commits in order. No rebase, force-push, amend, merge, close, or source-branch mutation occurred. |
| 3. Exact implementation scope and source blobs | PASS — implementation `c39cc1ddf783a8b785eb8a0bb6377de083ffd04f` is the direct child of the corrected brief and adds only the four declared test files. Their blobs exactly match PR #1 head: Codex untagged `d2494c9a652cea4bad65934a42ad3f1a0c8b6b4b`, Codex actual-host `5f432011a0df78ab667c1c616ca92741435c3d01`, Claude untagged `28cbf68cea4b38e778604c8a76b88b8deed93aea`, and Claude actual-host `7950b07c8d6ff613269e0d266a9b2a69e84edbe7`. |
| 4. Production and protected-control immutability | PASS — base-to-implementation contains only the corrected brief and four added `_test.go` files. No existing file is modified or deleted. Production Codex remains pinned to `codex-cli 0.149.1`; Claude remains pinned to `2.1.241`. Unknown-option rejection, typed `--max-turns 64`, arguments, permissions, output schemas, cancellation, cleanup, reviewer immutability, scope, containment, dependencies, workflows, toolchain controls, and the default-OFF `features.local_lifecycle` flag remain byte-identical to the base. |
| 5. Fake-runner contract | PASS by independent source inspection — both harnesses derive observations from fresh production argv for implementer and reviewer roles without mutating the input slices. Codex preserves exactly one final stdin sentinel, replaces it only with help, and inserts one test-owned unknown option before help. Claude requires exactly one `--max-turns 64` pair, appends help, inserts one unknown option before help, and changes only `64` to `not-an-integer` for its typed negative control. Exact request counts, empty input, bounded time/output, version identity, nonempty valid UTF-8 diagnostics, positive and negative exit classifications, unsafe-argv rejection, checked postconditions, and fail-closed runner, timeout, overflow, malformed, ambiguous, and unexpectedly successful outcomes are enforced. Help advertisement is derived only as diagnostic metadata and cannot override an exit result. |
| 6. Actual-host containment and offline boundary | PASS — both actual-host files are guarded by `l7_actual_provider`, contain no `init` or `TestMain`, and require exact provider-specific tokens plus source-root, temporary-parent, OS, architecture, target version, executable path, digest, candidate, and tree bindings. They require an owned detached clean repository with no remotes and recheck source and executable identity after both successful and earlier-failing gate paths. The bound verification compiled tagged coverage with no selected test. No provider executable, version/help interface, prompt/stdin, model session, retry, fallback, installation, provider network activity, or global provider configuration participated. |
| 7. Bound offline verification and state | PASS on the sole implementer verification record — verification commit `5d0c84e6218dba0496a02dce019bd99f11a617f1`, tree `e37f4d9a68bb19ef472c8823a4c30d8b0d6be41a`, is the direct child of implementation `c39cc1ddf783a8b785eb8a0bb6377de083ffd04f`, tree `b9b22c203e91b2bd59855e25dda54abca134c20f`, and adds only `docs/artifacts/changes/provider-no-model-gate-harness-mainline-verification.md`. It records PASS for targeted qualification tests, repository-pinned `make verify`, the complete applicable race suite, cross-builds, controller policy, diff hygiene, source-blob identity, ancestry, artifact budget, and clean tracked/index state. Independent `git diff --check` and tracked/index inspection also pass. This audit did not rerun tests, builds, or the controller. |
| 8. Independent audit and post-verification immutability | PASS — the active user separately authorized this read-only `l7-release` audit. Accountable owner `apbusinessidentity-tech`, implementer and PR author `addressanup`, and designated auditor `anup19950725` are distinct GitHub identities. No implementation commit follows verification, and the audit worktree remains exactly at the verification commit/tree with clean tracked and index state. Faithful materialization may add only this audit path; any later implementation change invalidates both verification and audit. |
| 9. Hosted exact-head boundary | PASS as a fail-closed future transition — repository variable `L7_ACCOUNTABLE_OWNER` is exactly `apbusinessidentity-tech`. GitHub review `5053475994` records that owner approving exact corrected brief `dff13dceb481867844fb6f26f087367e7d36849d`; the older approval on `a10759533…` is dismissed. PR #3 remains at the corrected brief head, so its checks and review do not transfer to the local implementation, verification successor, or a future audit head. After separately authorized publication, the exact audit head must receive fresh required Harness checks, successful trusted policy, exact-head approval from `apbusinessidentity-tech`, and exact-head approval from independent auditor `anup19950725`. |
| 10. Actual-host gate state and non-promotion | PASS — both successor no-model gates remain `NOT_RUN`. The test-only harness has no compatibility-admission path. Even a future diagnostic `GO` cannot change a supported version or support claim without a separate Tier 3 promotion change. |
| 11. Claim and effect boundary | PASS — the verified candidate changes only governance and test files. It creates no release, deployment, installation, supported-plugin publication, provider-support claim, PR #1 update, or user-visible production exposure. This read-only audit created no such effect. |

## Exact harness and containment assessment

The Codex harness makes one exact-version request followed by bounded top-level
help and implementer/reviewer help and unknown-option observations. It requires
six requests in total, uses empty input for every request, and fails if a
negative parser control exits zero or ambiguously.

The Claude harness makes one exact-version request followed by implementer and
reviewer help, unknown-option, and invalid-`--max-turns` observations. It
requires seven requests in total, uses empty input for every request, and
requires both negative controls for each role to exit nonzero.

Both harnesses combine bounded stdout and stderr, reject empty, whitespace-only,
oversized, invalid UTF-8, or NUL-containing output, use minimal environments and
timeouts, and always execute the registered postcondition after an earlier gate
error. Help-surface advertisement is collected only after a valid positive help
result and is not used to classify a parser control.

The actual-host entry points additionally fail closed unless their exact
provider-specific authorization tuple is present. Source and executable
identity are checked before provider activity, through the checked
postcondition, and through test cleanup. They do not call production
compatibility admission and cannot by themselves promote support.

## Historical provider facts

Historical Codex actual-host Gate 1 remains `PASS` only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`.

Historical Claude actual-host Gate 2 remains `NO_GO`. Both exact implementer and
reviewer help invocations succeeded. Both unknown-option parser controls
unexpectedly exited successfully. Both invalid `--max-turns not-an-integer`
controls failed as required. Neither help surface advertised `--max-turns`.

Help advertisement remains non-dispositive; the two successful unknown-option
controls remain dispositive. None of those candidate-bound observations
transfers to this successor. Both successor gates remain `NOT_RUN`.

## Artifact budget and immutable successor boundary

Relative to the exact base, the verification successor contains:

- one current brief;
- one verification record; and
- four exact test files.

The invalid intermediate brief is absent from the final tree but preserved in
Git history. Adding only this audit record produces exactly one brief, one
verification record, and one audit record: the Tier 3 maximum of three artifact
paths.

All change after implementation is currently confined to the sole verification
record. A valid audit successor must be a direct child of
`5d0c84e6218dba0496a02dce019bd99f11a617f1` and add only
`docs/artifacts/changes/provider-no-model-gate-harness-mainline-audit.md`.
Any second changed path or implementation-byte change fails closed and returns
the change to fresh briefing, approval, implementation, and verification.

## Hosted transition boundary

PR #3 is still published only through corrected brief
`dff13dceb481867844fb6f26f087367e7d36849d`. Its existing owner approval is
valid for the implementation authority boundary but is not an exact-head
approval of the verification or a future audit successor. Its technical and
trusted-policy results also do not establish future exact-head readiness.

After separately authorized audit materialization and publication, the
resulting exact audit head requires:

1. fresh successful required Harness checks;
2. successful trusted `evaluate`;
3. exact-head approval from accountable owner `apbusinessidentity-tech`;
4. exact-head approval from independent auditor `anup19950725`;
5. an audit envelope whose actor is `anup19950725` and whose candidate is the
   exact verification commit; and
6. no implementation change after verification.

Old-head reviews, checks, local designation, this record, or historical
provider evidence cannot substitute for those hosted facts.

## Rollback proof

| State | Required reverse sequence | Independent result |
|---|---|---|
| Prior-proposal-only `a10759533…` | Prior invalid proposal | PASS — reverting `a10759533e8a11201f29393698f1de71a8b4a6bf` removes its sole brief and restores exact base tree `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`. |
| Corrected proposal `dff13dce…` | Corrected proposal, then prior invalid proposal | PASS — reverting `dff13dceb481867844fb6f26f087367e7d36849d` restores the invalid brief path and removes the corrected path; reverting `a10759533…` then removes that restored path and returns to the exact base tree. |
| Implementation `c39cc1dd…` | Implementation, corrected proposal, then prior invalid proposal | PASS — the implementation is the direct child of the corrected proposal and adds only four tests. Reverse order removes the tests before unwinding both proposal commits. |
| Verification `5d0c84e6…` | Verification, implementation, corrected proposal, then prior invalid proposal | PASS — the verification is the direct child of implementation and adds only its sole record. Reverse order removes every current artifact and restores the exact base tree. |
| Audit-only successor | Audit, verification, implementation, corrected proposal, then prior invalid proposal | PASS — faithful materialization adds only this audit path. Reverting it first preserves the same direct reverse proof, removes all state-specific records in dependency order, and restores exact base tree `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`. |

Every rollback must use ordinary revert commits, preserve history, fail closed
on conflicts or extra paths, and confirm the final tree equals
`e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`.

PR state, remote branches, reviews, checks, repository visibility, variables,
and other hosted evidence are outside repository-tree rollback and remain
unchanged unless separately authorized.

## Findings and severity

No unresolved findings.

| Severity | Count |
|---|---:|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |

Hosted exact-head checks, trusted policy, owner approval, and auditor approval
remain future transition gates, not unresolved implementation findings.

## Read-only boundary

This audit used only read-only Git object, tree, ancestry, diff, blob, source,
tracked/index status, GitHub PR, commit, branch, review, repository-variable,
identity, collaborator-permission, and repository-state inspection.

It did not edit files, refs, index, configuration, envelopes, remotes, PRs,
reviews, variables, collaborators, branch protection, or any other local or
remote state. It did not run tests, builds, `make`, the controller, provider
executables, version/help probes, prompts/stdin, model sessions, retries,
fallbacks, provider installation, provider network activity, CI, merge,
release, deployment, or publication.

The unrelated untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md` was deliberately
excluded from inspection and was not touched.

## Authorization boundary

This audit record is returned as text only. No repository materialization or
external transition is authorized by the read-only audit itself.

If separately and explicitly authorized, faithful materialization must add
exactly this sole file as the only tree change after
`5d0c84e6218dba0496a02dce019bd99f11a617f1`. The matching audit envelope must
use schema `1`, change ID `provider-no-model-gate-harness-mainline`, actor
`anup19950725`, candidate commit
`5d0c84e6218dba0496a02dce019bd99f11a617f1`, the resulting exact audit commit,
and source `independent-agent`.

This decision authorizes no provider execution, actual-host gate, version/help
invocation, prompt/stdin, model session, retry, fallback, implementation,
remediation, rollback, re-verification, promotion, support claim, hosted CI,
remote update, GitHub review, merge, release, deployment, installation, or
publication.
