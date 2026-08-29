# Exact-Candidate Fast-Forward Integration — Verification

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Candidate commit | `ae67880506e34aea277ae6fe0e5a78abd18850bb` |
| Candidate tree | `d3a03866e6590b14ded2e6034aab2474c9e52426` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T05:51:11Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval binding | PASS — GitHub review `5057008830` records accountable owner `apbusinessidentity-tech` approving exact brief commit `61a005cc85d4900c4210152198c3650592418e42`; the local `trusted-ci` envelope binds that actor, implementer `addressanup`, and the same brief commit. |
| Controller | PASS — `BCTL-000` selected Tier 3 change `exact-candidate-fast-forward-integration`, base `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, the candidate and tree above, five final changed paths, state `building`, and the verification transition. |
| Focused contract | PASS — `make exact-fast-forward-integration-check` used fake `gh` and `git` executables for three malformed-input probes, fourteen failure scenarios, and one ordered success scenario during `make verify`; three additional consecutive focused runs also passed. |
| Failure containment | PASS — wrong actor, stale base/head/tree, missing owner approval, failed check, wrong-app duplicate check, unsafe protection, nonterminal input, confirmation mismatch, ref race, push failure, restoration failure, and postcondition failure all failed at their declared boundaries. |
| Effect contract | PASS — the success scenario performed exactly one main ref update after administrator enforcement was disabled and before it was restored. It reached no generic force flag, leading-plus refspec, merge API, pull-request edit/close API, branch deletion, local ref update, retry, or fallback. |
| Restoration | PASS — push failure and handled ref-race paths attempted restoration; restoration failure produced the blocking sole recovery action. Postflight required byte-equivalent canonical protection, exact main commit/tree, unchanged source ref, indirect GitHub merge state, unchanged repository settings, and no rulesets. |
| Deterministic terminal driver | PASS — the fake-command harness keeps its input pipe open until the child reports `PASS` or `BLOCKED`, eliminating the observed BSD `script(1)` EOF race while preserving a real PTY and exact typed-SHA confirmation. |
| `make verify` | PASS — controller, offline modules, import/effect boundaries, formatting, the focused contract, shell syntax, vet, type compilation, compile-only actual-host coverage, complete tests, harness reproducibility, and CLI reproducibility passed at the candidate commit. |
| Static shell analysis | PASS — ShellCheck with POSIX `sh` semantics and `sh -n` passed both new executable scripts. |
| Cross-build | PASS — `make cli-cross-build` produced the declared Darwin arm64 and amd64 binaries with the pinned offline toolchain. |
| Scope and hygiene | PASS — the final base-to-candidate tree changes only `Makefile`, `README.md`, the sole brief, and the two declared scripts. The stale verification path is absent; `git diff --check` passed and the tracked worktree/index were clean before this record. |
| Artifact budget | PASS — the candidate tree contained one brief and no verification or audit record. This is the sole current verification record; one independent audit record remains within the Tier 3 maximum. |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |
| Operator script | `0b93d5afc5b961eaea6b10f0495b285042ef15a94d3753796b9f0b35e56511d5` |
| Contract test | `6c1a84d8a1e8b2d290e2774d619ee5a7713e33397dc64319364657a97bef1112` |

The authenticated bootstrap verified the pinned Darwin arm64 Go archive digest
`020a1e8224811be75163e920bc77e0926a1390a6aeea19bdcf23f74b9d749f6d`
and signing subkey `0E225917414670F4442C250DFD533C07C264648F` before installing it in the
repository-local ignored cache. Module and provider network access remained
disabled during verification.

## Preserved failed verification attempt

Commit `5e23951549428c91d51b45c6de8ee35629944ecc` recorded an earlier local
verification attempt for implementation `6ccde69749bd6e74371c02b2ff6ca7488df28465`.
A subsequent focused run exposed a nondeterministic EOF race in the BSD
`script(1)` test driver. The gate failed, that record became stale, and commit
`ae67880506e34aea277ae6fe0e5a78abd18850bb` removed it from the candidate tree
while correcting only the test driver. Both commits remain in history; neither
is promoted as current evidence.

## Live-effect boundary

The real operator script was not invoked. Every effect-path test substituted
fake `gh` and `git` commands inside a temporary directory. No branch protection,
remote ref, pull request, source branch, repository setting, local Git ref, or
provider state was mutated by verification.

The implementation requires an exact expected-old lease and independently
proves the candidate is a descendant of that old commit. It sends one full-SHA
refspec to `refs/heads/main`; no generic force option or leading-plus refspec is
reachable. The temporary clean bare repository suppresses ambient Git
configuration and hooks. Its askpass helper is limited to authenticated GitHub
transport, while the offline contract harness replaces it with fake commands.

No Codex or Claude executable, version/help surface, prompt, stdin, model
session, retry, fallback, actual-host test, provider installation, release,
deployment, signing, publication, or Wave 5 work participated. Compile-only
actual-host coverage selected no provider test.

## Hosted boundary

The corrected implementation and this verification record were not pushed
during local verification. PR #4 remained open at proposal head
`61a005cc85d4900c4210152198c3650592418e42`, directly based on exact PR #3
head `f92c560cbe89e8d318e5521d9fc620f6153e9e14`. PR #3 remained open,
`CLEAN`, and approved; remote `main` remained
`be5c0c8f99b8ec55b42e1919533400fa0b41f46c`.

Fresh hosted checks and exact-head review/audit bindings do not exist yet for
the verification successor and cannot be inferred from proposal-head results.

## Rollback and next boundary

The exact post-PR-#3 local chain is brief `61a005cc…`, implementation
`6ccde697…`, invalidated verification `5e239515…`, and remediation `ae678805…`.
All commits are preserved. Reverting the eventual audit and current verification
record, then those four commits in reverse order, restores exact PR #3 tree
`3b4f7fe9dd09fbb53102e82473d392dcb2745ba0` using ordinary revert commits.
Remote `main` must never be reset or force-pushed.

This is implementer-run technical verification, not an independent audit or
live-effect authorization. The sole next transition is separately authorized,
independent, read-only `l7-release` audit of the verification successor. Any
implementation change invalidates this record and requires fresh verification.
