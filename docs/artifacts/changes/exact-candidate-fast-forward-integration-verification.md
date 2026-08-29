# Exact-Candidate Fast-Forward Integration — Automatic-Deletion Remediation Verification

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Candidate commit | `594f053aec9946bb7d0bc4c3334d9183634b670f` |
| Candidate tree | `f63a57e1de8d0c7c38bfbbb168b279d44960bbda` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T10:36:40Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval binding | PASS — GitHub review `5057548460` records accountable owner `apbusinessidentity-tech` approving exact restarted brief commit `83e5b78a16046cb89a76ac4c4df333f0b00eff41`; the local `trusted-ci` envelope binds that actor, implementer `addressanup`, and the same latest brief-addition commit. |
| Controller before verification | PASS — `BCTL-000` selected Tier 3 change `exact-candidate-fast-forward-integration`, base `a178f047d8d0269ae2b1b0aa957ff3b65ff75116`, the candidate and tree above, four changed paths, state `building`, and the candidate-bound verification transition. |
| Focused contract | PASS — `make verify` ran the fake-command contract once; three additional consecutive exact-commit runs also passed. Every run covered three malformed-input probes, twenty-three failure scenarios, and two ordered success scenarios. |
| Automatic source-branch deletion | PASS — repository validation requires `delete_branch_on_merge` to be exactly `false` during initial preflight and the complete post-confirmation refresh. Dedicated initially-enabled and post-confirmation-change scenarios fail with `automatic source-branch deletion is enabled` before administrator enforcement is disabled or any ref is updated. |
| Check chronology | PASS — competing trusted-app benchmark runs deliberately give the skipped run a greater numeric ID but an earlier `started_at`; the unique later-started successful run is selected. Missing or tied greatest timestamps and wrong-app duplicates fail closed. |
| Post-confirmation authority | PASS — repository settings, accountable-owner binding, sole-direct-admin identity, absence of rulesets, exact open PR state and identities, main/source refs, protection, reviews, checks, and legacy status duplication are freshly queried before enforcement is disabled. Dedicated post-confirmation repository, administrator, ruleset, PR, and automatic-deletion changes fail before mutation. |
| Restoration proof | PASS — restoration remains armed after POST until a fresh protection GET canonically equals the original contract. POST failure and successful-but-unapplied restoration both emit `BLOCKED RECOVERY` with re-enabling administrator enforcement as the sole next action. |
| Git containment | PASS — every Git invocation runs under `env -i` with an explicit minimal environment, isolated home/config/tmp paths, noninteractive token-backed askpass, and no ambient Git configuration or transport overrides. The injection scenario proves ambient Git configuration does not reach Git. |
| Exact lease | PASS — the only permitted force-prefixed argument is exactly `--force-with-lease=refs/heads/main:<full-expected-base>` after independent ancestry proof. The fake command requires one exact lease and main refspec while rejecting generic, unbound, inclusive, shorthand, and leading-plus force forms. |
| `make verify` | PASS — controller, offline modules, import/effect boundaries, formatting, the focused contract, shell syntax, vet, type compilation, compile-only actual-host coverage, complete tests, harness reproducibility, and CLI reproducibility passed at the exact candidate commit. |
| Static shell analysis | PASS — ShellCheck with POSIX `sh` semantics and `sh -n` passed both executable scripts. |
| Cross-build | PASS — `make cli-cross-build` produced the declared Darwin arm64 and amd64 binaries with the pinned offline toolchain. |
| Scope and hygiene | PASS — base-to-candidate changes only the sole brief, the two declared harness scripts, and deletion of the stale verification record. `git diff --check` passed, no audit exists, and the tracked worktree/index were clean before this record. |
| Artifact budget | PASS — the candidate tree contained one brief and no verification or audit record. This is the sole current verification record; one independent audit record remains within the Tier 3 maximum. |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |
| Operator script | `80e0f4226c46e367c9916b6b64773dd6d9715edcbfe975570c45d7334c8969fc` |
| Contract test | `3ae33f23aa7d8ee68ce89ecda02e89771cec94a3837780e965cf011f3b647fb6` |

Module and provider network access remained disabled during verification. No
provider executable, prompt, model session, actual-host gate, retry, fallback,
installation, or global provider configuration participated. Compile-only
actual-host coverage selected no provider test.

## Remediation and preserved history

The earlier independent read-only audit of verification successor
`5800205c9ce5ab8fcb9622b785bd6bd750b51f56`, tree
`d9855e1ff8446fc42ffa533dd7f09cbe160a265a`, returned `NO_GO` because the
operator accepted repository state with automatic source-branch deletion
enabled. No audit record or audit authority envelope was materialized.

Implementation `594f053aec9946bb7d0bc4c3334d9183634b670f` is the direct child
of that preserved verification successor. It changes only the two approved
harness scripts and deletes the stale verification record. The operator now
rejects both initially enabled automatic deletion and activation during the
post-confirmation refresh. The fake repository models both cases and proves
that neither reaches a protection mutation or ref update.

The five findings from the earlier rejected `bd0cb9b0...` candidate remain
closed: check selection uses a unique greatest valid `started_at`; the approved
brief explicitly authorizes only the exact expected-old lease; all mutable
authority is refreshed after confirmation; restoration stays armed until exact
canonical proof; and every Git invocation is isolated under `env -i`.

All rejected proposals, implementations, verification successors, ordinary
reverts, and governance records remain immutable ancestors. This record does
not supersede or rewrite that historical evidence.

## Live-effect and hosted boundary

The real operator script was not invoked. Every focused effect-path test used
fake `gh` and `git` commands inside an owned temporary directory. No branch
protection, remote ref, pull request, source branch, repository setting, local
Git ref, provider state, release, deployment, installation, signing, or
publication was changed by verification.

Read-only revalidation found remote `main` still at
`be5c0c8f99b8ec55b42e1919533400fa0b41f46c`; PR #3 remains open, clean,
approved, and unmerged at `f92c560cbe89e8d318e5521d9fc620f6153e9e14`;
and repository automatic branch deletion remains disabled. PR #4 remains open
at proposal head `83e5b78a16046cb89a76ac4c4df333f0b00eff41`.
This implementation and record remain local only. Proposal-head checks and
owner review `5057548460` establish implementation authority but not
exact-candidate hosted readiness.

## Rollback and next boundary

Reverting this verification record returns exactly to implementation
`594f053a...`. Reverting commits `594f053a...`, `5800205c...`, and
`d10857f...` in reverse order then returns to exact approved proposal
`83e5b78a...`; reverting that proposal and deletion `01992e1f...` in reverse
returns to restart base `a178f047...`, tree `6dcade32...`. Every step uses
ordinary revert commits and preserves history; remote `main` must never be
reset or force-pushed.

This is implementer-run technical verification, not an independent audit or
live-effect authorization. The sole next transition is a separately authorized,
fresh independent read-only `l7-release` audit of the resulting verification
successor. The auditor must be distinct from the implementer and owner, and no
GitHub reviewer identity is inferred here. Any implementation change
invalidates this record and requires fresh verification.
