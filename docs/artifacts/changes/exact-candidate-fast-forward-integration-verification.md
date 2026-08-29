# Exact-Candidate Fast-Forward Integration — Restarted Remediation Verification

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Candidate commit | `d10857f887d8d142c81715a63e884dbcd704c193` |
| Candidate tree | `d5f0ac303e161b3f630f5009d2a6f8bed7b634b0` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T09:54:20Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval binding | PASS — GitHub review `5057548460` records accountable owner `apbusinessidentity-tech` approving exact restarted brief commit `83e5b78a16046cb89a76ac4c4df333f0b00eff41`; the local `trusted-ci` envelope binds that actor, implementer `addressanup`, and the same latest brief-addition commit. |
| Controller | PASS — `BCTL-000` selected Tier 3 change `exact-candidate-fast-forward-integration`, base `a178f047d8d0269ae2b1b0aa957ff3b65ff75116`, the candidate and tree above, four changed paths, state `building`, and the verification transition. |
| Focused contract | PASS — `make exact-fast-forward-integration-check` used fake `gh` and `git` executables for three malformed-input probes, twenty-one failure scenarios, and two ordered success scenarios during `make verify`; three additional consecutive focused runs also passed. |
| Check chronology | PASS — competing trusted-app benchmark runs deliberately give the skipped run a greater numeric ID but an earlier `started_at`; the later-started successful run is selected. Missing or tied greatest timestamps and wrong-app duplicates fail closed. |
| Post-confirmation authority | PASS — repository settings, accountable-owner binding, sole-direct-admin identity, absence of rulesets, exact open PR state and identities, main/source refs, protection, reviews, checks, and legacy status duplication are all freshly queried before enforcement is disabled. Dedicated post-confirmation repository, administrator, ruleset, and PR changes fail before mutation. |
| Restoration proof | PASS — restoration remains armed after POST until a fresh protection GET canonically equals the original contract. POST failure and successful-but-unapplied restoration both emit `BLOCKED RECOVERY` with re-enabling administrator enforcement as the sole next action. |
| Git containment | PASS — every Git invocation runs under `env -i` with an explicit minimal environment, isolated home/config/tmp paths, noninteractive token-backed askpass, and no ambient Git configuration or transport overrides. The injection scenario proves `GIT_CONFIG_PARAMETERS`, `GIT_CONFIG_COUNT`, and key/value entries do not reach Git. |
| Exact lease | PASS — the only permitted force-prefixed argument is exactly `--force-with-lease=refs/heads/main:<full-expected-base>` after independent ancestry proof. The fake command requires one exact lease and one exact main refspec while rejecting generic, unbound, inclusive, shorthand, and leading-plus force forms. |
| `make verify` | PASS — controller, offline modules, import/effect boundaries, formatting, the focused contract, shell syntax, vet, type compilation, compile-only actual-host coverage, complete tests, harness reproducibility, and CLI reproducibility passed at the candidate commit. |
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
| Operator script | `17b6009e58041b926a4f2eae2a0b42cba0cebe86fe7f7c32d8fb8c3e329ccf5f` |
| Contract test | `0b038b49969802ac181f1ec982a5207a94ba496bdb4614a3b48ef5a1cac9936c` |

Module and provider network access remained disabled during verification. No
provider executable, prompt, model session, actual-host gate, retry, fallback,
installation, or global provider configuration participated. Compile-only
actual-host coverage selected no provider test.

## Preserved failed history and restart boundary

The complete history remains immutable:

1. `bd0cb9b0f7a39ba512e3a13d5d5a3da91ee25ff7` — rejected verification
   successor audited `NO_GO`.
2. `159a5de1beda5f4d5036cdccaebe30fcf45f0944` — first amended remediation
   brief, approved by the accountable owner but not representable as a fresh
   brief-addition authority boundary.
3. `eda83482d589915aa9a9ece8f1ac217535a33222` — preserved failed remediation
   commit whose controller rejected the amended-brief approval binding.
4. `a178f047d8d0269ae2b1b0aa957ff3b65ff75116` — ordinary revert restoring exact
   tree `6dcade32b7b4d765dea7925fe0f2e7326088c216`.
5. `01992e1f7745ab9eb423e76d5969040b7373be56` — brief-only deletion resetting
   the controller addition identity.
6. `83e5b78a16046cb89a76ac4c4df333f0b00eff41` — sole brief re-addition and
   approved planning boundary.
7. `d10857f887d8d142c81715a63e884dbcd704c193` — current verified remediation.

The two script blobs at the current candidate exactly match the independently
identified remediation bytes preserved at `eda83482...`, but they now follow a
controller-recognized and externally approved planning boundary. The stale
verification record is absent from the candidate tree and remains available in
history only.

## Live-effect boundary

The real operator script was not invoked. Every focused effect-path test used
fake `gh` and `git` commands inside an owned temporary directory. No branch
protection, remote ref, pull request, source branch, repository setting, local
Git ref, provider state, release, deployment, installation, signing, or
publication was changed by verification.

The exact expected-old lease remains a compare-and-swap guard, not permission
for a non-fast-forward update: immutable ancestry is proved independently, the
candidate is the sole full-SHA refspec, and all generic force forms remain
unreachable.

## Hosted boundary

PR #4 remains open at proposal head
`83e5b78a16046cb89a76ac4c4df333f0b00eff41`; this remediation candidate and
record are local only. Proposal-head checks and review `5057548460` establish
implementation authority but do not provide exact-candidate hosted readiness.

PR #3 remains open, `CLEAN`, and approved at
`f92c560cbe89e8d318e5521d9fc620f6153e9e14`. Remote `main` remains
`be5c0c8f99b8ec55b42e1919533400fa0b41f46c`. Fresh hosted checks and exact-head
owner/auditor approvals do not exist for this verification successor and
cannot be inferred from older results.

## Rollback and next boundary

Reverting the eventual audit and this verification record, then remediation
`d10857f...`, returns to exact approved proposal `83e5b78a...`. Reverting that
brief re-addition and deletion `01992e1f...` in reverse returns to restart base
`a178f047...`, tree `6dcade32...`. Every step uses ordinary revert commits;
remote `main` must never be reset or force-pushed.

This is implementer-run technical verification, not an independent audit or
live-effect authorization. The sole next transition is a separately authorized,
independent, read-only `l7-release` audit of the verification successor. Any
implementation change invalidates this record and requires fresh verification.
