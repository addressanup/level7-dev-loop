# Provider No-Model Gate Harness — Mainline

| Field | Value |
|---|---|
| Change ID | `provider-no-model-gate-harness-mainline` |
| Risk tier | `3` — provider qualification and protected test controls |
| Status | `proposed`; implementation is not approved |
| Base commit | `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` |
| Base tree | `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6` |
| Prior proposal | `a10759533e8a11201f29393698f1de71a8b4a6bf`; preserved but removed from the final tree because its headings and scope were not controller-readable |
| Source evidence | PR `addressanup/level7-dev-loop#1`, head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`, tree `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb` |
| Selected path | History-preserving proposal correction and clean-mainline carry-forward; PR #1 remains unchanged |
| Accountable owner | GitHub user `apbusinessidentity-tech`; fresh exact-brief approval is required before implementation |
| Implementer | `addressanup`, operated through `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false`; no production behavior changes |

## Problem

PR #1 contains an independently audited, test-only no-model parser-gate harness,
but its head is four commits ahead and five commits behind current `main`. Its
old verification, audit, reviews, and CI results cannot qualify a new exact head.

The first mainline successor proposal at `a10759533e8a11201f29393698f1de71a8b4a6bf`
used combined section headings and omitted the controller's exact implementation
file-set section. The repository controller correctly rejected that brief before
full verification. Preserve that commit in history, remove its invalid brief
from the final tree, and use this replacement as the sole current change brief.

Carry forward only the four already verified harness test blobs onto exact
current `main`. Do not rebase, force-push, amend, close, or otherwise mutate PR
#1 or its source branch.

Production continues to qualify only Codex `codex-cli 0.149.1` and Claude Code
`2.1.241`. Codex `codex-cli 0.150.1`, Claude `2.1.247`, and Claude
`2.1.247 (Claude Code)` remain degraded before semantic invocation.

Historical Codex actual-host Gate 1 remains bound only to failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Historical Claude Gate 2 remains
`NO_GO`: both exact role help invocations succeeded, both unknown-option parser
controls unexpectedly exited successfully, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement remains non-dispositive.
Both successor no-model actual-host gates remain `NOT_RUN`.

## Scope

Create a clean-mainline successor that adds only the exact test harness, this
brief, one verification record, and one audit record. The proposal-correction
commit removes the invalid intermediate brief path and adds this brief; relative
to the declared base, only this brief remains.

No existing production, test, workflow, toolchain, plugin, skill, configuration,
dependency, or historical artifact may be modified.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-no-model-gate-harness-mainline.md`
- `docs/artifacts/changes/provider-no-model-gate-harness-mainline-verification.md`
- `docs/artifacts/changes/provider-no-model-gate-harness-mainline-audit.md`
- `internal/l7/adapter/codex/qualification_test.go`
- `internal/l7/adapter/codex/qualification_actual_host_test.go`
- `internal/l7/adapter/claude/qualification_test.go`
- `internal/l7/adapter/claude/qualification_actual_host_test.go`

The four test files must be byte-for-byte identical to these source blobs at
`f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`:

- `internal/l7/adapter/codex/qualification_test.go` — `d2494c9a652cea4bad65934a42ad3f1a0c8b6b4b`
- `internal/l7/adapter/codex/qualification_actual_host_test.go` — `5f432011a0df78ab667c1c616ca92741435c3d01`
- `internal/l7/adapter/claude/qualification_test.go` — `28cbf68cea4b38e778604c8a76b88b8deed93aea`
- `internal/l7/adapter/claude/qualification_actual_host_test.go` — `7950b07c8d6ff613269e0d266a9b2a69e84edbe7`

Modify: none.

## Acceptance criteria

1. Exact base and tree above remain the change base. Proposal `a10759533…` and
   its parent relationship to the base remain immutable history; the corrected
   proposal's final base diff contains only this replacement brief.
2. PR #1 head `f0e9f54c…`, tree `82bace9a…`, and its four commits remain
   unchanged and reachable on the existing remote branch.
3. Implementation adds only the four declared test files, each with the exact
   blob ID above. Any required byte change fails closed and requires a revised
   brief and fresh approval.
4. Production blobs and compatible-version constants remain unchanged from the
   base. Unknown-option rejection, typed `--max-turns 64`, argument and
   permission restrictions, output validation, cancellation, cleanup, reviewer
   immutability, scope, containment, and the default-OFF flag remain intact.
5. Untagged fake-runner tests retain exact argv derivation, role separation,
   negative parser controls, bounded diagnostics, postcondition checks, and
   fail-closed classification. Help advertisement remains diagnostic only.
6. Build-tagged actual-host files compile with no selected actual-host test. No
   provider executable, version/help interface, prompt/stdin, model session,
   retry, fallback, provider installation, or provider network activity runs.
7. Targeted provider tests, repository-pinned `make verify`, the complete
   applicable race suite, `make cli-cross-build`, controller policy, diff
   hygiene, blob identity, ancestry, artifact budget, and clean tracked/index
   state pass before the sole verification record is committed.
8. A separately authorized independent read-only `l7-release` audit binds the
   verification successor and adds only the sole audit record. Any later
   implementation change invalidates verification and audit.
9. Any published audit head receives fresh required Harness checks, trusted
   policy, exact-head approval from `apbusinessidentity-tech`, and exact-head
   approval from a distinct independent auditor. Prior reviews do not transfer.
10. Both no-model actual-host gates remain `NOT_RUN`. Even future diagnostic
    `GO` results cannot promote support without a separate Tier 3 change.
11. No release, deployment, installation, supported-plugin publication,
    provider-support claim, PR #1 update, or user-visible exposure occurs.

## Risks and mitigations

- **Mainline drift:** exact blob carry-forward plus full verification fails
  closed on any incompatibility.
- **Stale authority:** all approval, verification, audit, CI, and reviews bind
  the new exact commits; `a10759533…` approval cannot transfer.
- **Control weakening:** the harness remains test-only and cannot reach
  compatibility admission; production controls are compared to the base.
- **History loss:** use only additive commits and ordinary reverts; never amend,
  rebase, force-push, or delete PR #1 or either proposal commit.
- **Unrelated user state:** leave
  `docs/artifacts/foundation-rebaseline-admission-audit.md` untouched,
  uninspected, and unstaged through exact pathspecs.

## Rollback

Rollback is history-preserving and must restore exact base tree
`e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`.

- Before implementation: revert the corrected-brief commit, then revert prior
  proposal `a10759533e8a11201f29393698f1de71a8b4a6bf`.
- After implementation: revert implementation, corrected brief, then prior
  proposal `a10759533…`.
- After verification: revert verification, implementation, corrected brief,
  then prior proposal.
- After audit: revert audit, verification, implementation, corrected brief,
  then prior proposal.

Every step uses ordinary revert commits, rejects conflicts or extra paths, and
confirms the final tree equals the exact base tree. PR #1, remote reviews,
hosted checks, repository visibility, and other external evidence remain outside
tree rollback and unchanged.

## Commit sequence and approval boundary

1. `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` — exact current main.
2. `a10759533e8a11201f29393698f1de71a8b4a6bf` — preserved invalid proposal.
3. `docs(provider): correct mainline harness brief contract` — replace only the
   invalid brief path with this controller-conformant brief.
4. Stop for fresh exact-brief approval from `apbusinessidentity-tech`.
5. `test(provider): add mainline no-model gate harness` — add only four exact blobs.
6. Run offline verification and freeze the implementation commit/tree.
7. `test(provider): record mainline no-model harness verification` — add the
   sole verification record.
8. Obtain separate authorization for independent read-only `l7-release` audit.
9. `docs(provider): record mainline no-model harness audit` — add the sole audit.
10. Stop. Actual-host gates, promotion, merge, release, deployment, installation,
    and supported publication require separate exact-head transitions.

This brief commit authorizes no implementation by itself. Repository text,
historical approvals, and passing tests cannot substitute for fresh external
accountable-owner approval bound to the exact corrected brief commit.
