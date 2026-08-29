# Exact-Candidate Fast-Forward Integration — Hosted-Policy Restart Verification

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Candidate commit | `d2d0b1287350c3eb7ed2956ccb6890b80b505279` |
| Candidate tree | `228083e557f3770480569618f93af0e8143a1c88` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T12:53:10Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Owner binding | PASS — current external interaction identifies accountable owner `apbusinessidentity-tech`, implementer `addressanup`, and exact proposal `b22b903cf1adb94c374a5a7225782930e58aa907`; the active-user-interaction envelope binds the same values. |
| Planning topology | PASS — reset base `1e381b5bf0bb024739cdc654d0d5aed5f128aed4` has the exact prior brief-absent tree. Its base-to-proposal diff adds only the brief; the trusted workflow's added-file query selects that path and Git resolves `b22b903...` as its addition commit. |
| Scope and controller | PASS — base-to-candidate changes exactly four paths: the sole added brief, two modified harness scripts, and deletion of the stale verification record. `BCTL-000` selected Tier 3, exact base/candidate/tree, `state=building`, and the verification transition. |
| `make verify` | PASS — controller, offline dependency checks, import/effect boundaries, formatting, focused contract, shell syntax, vet, type compilation, compile-only actual-host coverage, complete tests, harness reproducibility, and CLI reproducibility passed at the exact candidate. |
| Focused operator contract | PASS — the dedicated fake-command contract passed independently after the full gate: three malformed-input probes, twenty-three failure scenarios, and two ordered success scenarios. |
| Static and cross-build checks | PASS — `sh -n` and ShellCheck passed both scripts; `make cli-cross-build` produced Darwin arm64 and amd64 binaries with the pinned offline toolchain. |
| Critical containment | PASS — tests prove unique greatest trusted-app `started_at` selection, exact bound lease only, complete post-confirmation authority refresh, recovery armed through canonical restoration proof, Git under `env -i`, and initial plus post-confirmation rejection of automatic source-branch deletion. |
| Hygiene and artifact budget | PASS — `git diff --check` passed; tracked worktree and index were clean before this record; no audit record exists; the Tier 3 artifact budget remains within its maximum. |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |
| Operator script | `80e0f4226c46e367c9916b6b64773dd6d9715edcbfe975570c45d7334c8969fc` |
| Contract test | `3ae33f23aa7d8ee68ce89ecda02e89771cec94a3837780e965cf011f3b647fb6` |

Module and provider network access remained disabled. No provider executable,
prompt, model session, actual-host gate, retry, fallback, install, release,
deployment, signing, or publication participated. Compile-only actual-host
coverage selected no provider test.

## Hosted boundary and preservation

Read-only revalidation found remote `main` unchanged at
`be5c0c8f99b8ec55b42e1919533400fa0b41f46c`; PR #3 remains open, clean, and
unmerged at `f92c560cbe89e8d318e5521d9fc620f6153e9e14`; and automatic branch
deletion remains disabled. PR #4 remains the stale one-file proposal at
`83e5b78a16046cb89a76ac4c4df333f0b00eff41` against the unprotected
`a178f047...` base. It supplies no evidence for this successor.

The corrected proposal and candidate are local only. A dedicated ref at the new
reset base has not been created or protected, no pull request targets it, and no
candidate hosted checks or reviews have run. Those states are intentionally
`NOT_RUN`, not inherited from PR #4, and this verification makes no hosted
readiness or provider-support claim.

All rejected candidates, audits, reverts, and governance records remain
immutable ancestors. Local `main`, PR #1, PR #3, the invalid implementation
worktree, remote refs, branch protection, repository settings, and unrelated
user state were not changed.

## Rollback and next boundary

Ordinary-revert this verification record to return exactly to implementation
`d2d0b128...`. Then revert the implementation and proposal in reverse order to
return to reset tree `5eff6de7...`. Never reset, rebase, force-push, delete, or
rewrite a historical ref.

This is implementer-run technical verification, not an independent audit or
live-effect authorization. The sole next transition is a separately authorized
fresh independent read-only `l7-release` audit of the verification successor.
The auditor must be distinct from the accountable owner and implementer. Any
implementation change invalidates this record and requires fresh verification.
