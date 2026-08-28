# Provider Compatibility No-Model Gate Harness — Mainline Successor

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness-mainline-successor` |
| Risk tier | `3` — provider qualification and protected test controls |
| Status | `proposed`; implementation is not approved |
| Base commit | `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` |
| Base tree | `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6` |
| Source evidence | PR `addressanup/level7-dev-loop#1`, head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`, tree `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb` |
| Selected path | Clean-mainline successor; preserve PR #1 and its four-commit history unchanged |
| Accountable owner | GitHub user `apbusinessidentity-tech`; exact-brief approval is required before implementation |
| Implementer | `addressanup`, operated through `codex-root` |
| Independent auditor | Must be distinct from owner and implementer; anticipated GitHub reviewer `anup19950725` requires fresh audit authorization and exact-head review |
| Feature flag | Existing `features.local_lifecycle`, default `false`; no production behavior changes |

## Problem and disposition

PR #1 contains an independently audited, test-only no-model parser-gate harness,
but its exact head is based on `481adaaec967ac34b6b27cf78509b85d5c068abc`.
Current `main` is five commits ahead after the separately audited hosted-CI
remediation. PR #1 is therefore four commits ahead and five commits behind its
merge base. Its old verification, audit, reviews, and CI results cannot qualify
a new exact head.

Create a new branch from the exact current `main` and carry forward only the
four verified harness test blobs. Do not rebase, merge into, force-push, amend,
close, or otherwise mutate PR #1 or its source branch. Its brief, implementation,
verification, audit, and failed hosted runs remain read-only historical evidence.

This successor does not promote provider support. Production continues to
qualify only Codex `codex-cli 0.149.1` and Claude Code `2.1.241`. Codex
`codex-cli 0.150.1`, Claude `2.1.247`, and Claude `2.1.247 (Claude Code)` remain
degraded before semantic invocation.

Historical Codex actual-host Gate 1 remains bound only to failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Historical Claude Gate 2 remains
`NO_GO`: both exact role help invocations succeeded, both unknown-option parser
controls unexpectedly exited successfully, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement remains non-dispositive.

The successor Codex and Claude no-model parser gates remain `NOT_RUN`.

## Scope

After exact-brief approval, add the following files byte-for-byte from source
head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`:

- `internal/l7/adapter/codex/qualification_test.go` — blob `d2494c9a652cea4bad65934a42ad3f1a0c8b6b4b`
- `internal/l7/adapter/codex/qualification_actual_host_test.go` — blob `5f432011a0df78ab667c1c616ca92741435c3d01`
- `internal/l7/adapter/claude/qualification_test.go` — blob `28cbf68cea4b38e778604c8a76b88b8deed93aea`
- `internal/l7/adapter/claude/qualification_actual_host_test.go` — blob `7950b07c8d6ff613269e0d266a9b2a69e84edbe7`

Add exactly these three governance paths over the life of this change:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness-mainline-successor.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness-mainline-successor-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness-mainline-successor-audit.md`

Modify no existing file. In particular, do not change production adapters,
compatibility profiles, CLI behavior, arguments, permissions, output schemas,
cancellation, cleanup, reviewer immutability, scope or containment controls,
workflows, toolchain locks, bootstrap logic, dependencies, skills, plugin
manifests, configuration, historical records, remotes, or global configuration.

## Acceptance criteria

1. The branch starts at exact base and tree above. PR #1 head/tree and its four
   commits remain unchanged and reachable on its existing remote branch.
2. The implementation adds only this approved brief and the four declared test
   files. Each test file has the exact source blob ID listed above; any required
   byte change fails closed and requires a revised brief.
3. Production blobs and compatible-version constants are unchanged from the
   base. Unknown-option rejection, typed `--max-turns 64`, all argument and
   permission restrictions, strict role output validation, cancellation,
   cleanup, reviewer immutability, scope, containment, and the default-OFF flag
   remain intact.
4. Untagged fake-runner tests retain exact argv derivation, role separation,
   negative parser controls, bounded diagnostics, postcondition checks, and
   fail-closed classification. Help advertisement remains diagnostic only.
5. Build-tagged actual-host files compile with no selected actual-host test.
   No provider executable, version/help interface, prompt/stdin, model session,
   retry, fallback, installation, network provider activity, or global provider
   configuration participates in implementation or offline verification.
6. Repository-pinned targeted provider tests, `make verify`, the complete
   applicable race suite, `make cli-cross-build`, diff hygiene, blob identity,
   ancestry, artifact budget, and tracked/index cleanliness pass before the
   sole verification record is committed.
7. A separately authorized independent read-only `l7-release` audit binds the
   verification successor, maps every criterion, and adds only the sole audit
   record. Any implementation change after verification invalidates it.
8. If published, the exact future audit head receives fresh required Harness
   checks, trusted policy, accountable-owner approval from
   `apbusinessidentity-tech`, and independent auditor approval. Old-head checks
   and reviews do not transfer.
9. Both no-model actual-host gates remain `NOT_RUN`. Even future diagnostic
   `GO` results cannot alter production compatibility or support without a
   separate Tier 3 promotion change.
10. No release, deployment, installation, publication, provider-support claim,
    update to PR #1, or user-visible production exposure occurs in this change.

## Risks and rollback

- Mainline drift could alter the harness contract. Exact blob carry-forward and
  full verification fail closed on incompatibility.
- Historical evidence could be mistaken for current authority. All verification,
  audit, CI, and reviews must bind the new exact commits.
- A parser observation could be used to weaken production controls. The harness
  remains test-only and cannot reach compatibility admission.
- The unrelated user-owned untracked
  `docs/artifacts/foundation-rebaseline-admission-audit.md` could be staged.
  Use exact pathspecs and leave it untouched and uninspected.

Rollback is history-preserving. Before verification, revert implementation and
then this brief. After verification, revert the verification record,
implementation, and brief. After audit, revert audit, verification,
implementation, and brief. Each sequence uses ordinary revert commits, rejects
conflicts or extra paths, and confirms the final tree equals
`e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6`. PR #1 and hosted evidence remain
outside this rollback and unchanged.

## Commit sequence and approval boundary

1. `be5c0c8f99b8ec55b42e1919533400fa0b41f46c` — exact current main.
2. `docs(provider): define mainline no-model harness successor` — add only this brief.
3. Stop for exact-brief approval from `apbusinessidentity-tech`.
4. `test(provider): add mainline no-model gate harness` — add only the four exact test blobs.
5. Run offline verification and freeze the implementation commit/tree.
6. `test(provider): record mainline no-model harness verification` — add the sole verification record.
7. Obtain separate authorization for an independent read-only `l7-release` audit.
8. `docs(provider): record mainline no-model harness audit` — add the sole audit record.
9. Stop. Remote publication, PR creation, hosted CI, reviews, actual-host gates,
   promotion, merge, release, deployment, and installation require their own
   explicit exact-head transitions.

This brief commit authorizes no implementation by itself. Repository text,
historical approvals, and passing tests cannot substitute for fresh external
accountable-owner approval bound to the exact brief commit.
