# Standalone CLI v1 Wave 4 — Independent Audit

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-4` |
| Candidate commit | `a92c9ac5d00f638da440c9879576bb8ed6090000` |
| Candidate tree | `3687b50353f3fc7adbe582af95d9623b7234a8bd` |
| Result | `GO` |
| Reviewer | `claude-wave4-independent-auditor` |

## Decision

`GO`. The distinct Claude Code auditor inspected implementation candidate
`c7d30046f95cc276f14864d502eabcabe917af97` and verified successor
`a92c9ac5d00f638da440c9879576bb8ed6090000` against approved base
`be898482ce632a8faf60f97022f740cd24585063`. It found no blocker or high-severity
issue, no scope expansion, no false-ready effect path, and no misleading claim
that an unobserved external capability had run. Three low residuals and five
notes do not invalidate the approved Wave 4 contract.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| Default OFF | PASS — `loadExecutionContext` blocks local readiness persistence and merge before lock or effect; the disabled-path test asserts zero receipt writes, confirmations, and ref advances |
| Exact evidence | PASS — the exact configuration bytes are SHA-256 bound; verification/check identity, candidate/tree, audit, distinct Tier 3 identities, and benchmark facts are re-evaluated |
| Restart truth | PASS — status reconstructs current readiness/merge from Git plus receipts and marks legacy or changed evidence stale with an executable transition |
| Terminal authority | PASS — merge rejects JSON/non-interactive input and requires an exact full candidate SHA after displaying the explicit target and old/new identities |
| Race closure | PASS — readiness and merge facts are recomputed under lock and after confirmation; target, worktree, ancestry, evidence, and compare-and-swap races fail closed |
| Sole merge effect | PASS — production `update-ref` ownership is statically pinned to `internal/l7/adapter/git/merge.go`; no checkout, push, fetch, reset, rebase, merge commit, delete, or remote effect was found |
| Pure headless path | PASS — only the strict bounded CI decoder and domain evaluator are composed; duplicate/unknown/trailing input fails and effectful imports are prohibited |
| Both provider orders | PASS by implementation/test inspection — disposable fake-provider tests cover both orders through ready and local merge without remotes |
| False readiness | PASS — dirty/stale/forged/self-review/`NO_GO`/config/check/benchmark/target/ref-race cases are covered and fail closed |
| Process and reviewer containment | PASS by source/test inspection — escaped inherited pipes have a second bound; Claude reviewer argv contains only `Read,Glob,Grep` in safe plan mode |
| Performance gate | PASS — paired sample sets, minimum count, medians, raw values, and the strict greater-than-10% block are enforced; the exact-head external owner marker cannot waive missing/cancelled evidence |
| Dependency and claim hygiene | PASS — zero production dependencies remain; the README and verification record explicitly mark providers, hosted CI, live rules, and Intel runtime unobserved |

The auditor also found all six Wave 3 audit findings closed, all 49 successor
paths contained by the approved Wave 4 file set, `.l7/config.json`, `go.mod`, and
`go.sum` unchanged, and the Tier 3 artifact budget intact before this sole audit
record was added.

## Findings

### `W4-AUD-001` — low — headless evaluation is not feature-flag gated

- `cmd/l7/main.go` composes only `Ports{DecodeCI}` for `ready --headless` and
  never loads repository configuration.
- `internal/l7/app/readiness.go` enters the pure headless evaluator before the
  local lifecycle feature check.
- `cmd/l7/readiness_test.go` intentionally proves a true decision outside Git.
- This is narrower than the brief's broad statement that every new visible
  behavior is gated, but it satisfies the explicit acceptance criterion that
  disabled configuration blocks readiness persistence and merge. The path has
  no prompt, state write, provider, network, or merge capability and cannot
  create a receipt consumed by local merge.

Disposition: accepted as a documentation tension without a safety effect; no
remediation required for this `GO`.

### `W4-AUD-002` — low — Claude read-only semantics remain untried on an actual host

- `internal/l7/adapter/claude/adapter.go` pins Claude Code `2.1.241`, uses
  `--safe-mode`, plan permission, and reviewer tools `Read,Glob,Grep`.
- Unknown versions or unsupported flags fail closed.
- README and verification evidence explicitly make no current provider-support
  claim and mark actual-host provider-order trials `NOT_RUN`.

Disposition: the argv-level claim is accurate and closes the Wave 3 residual;
real semantics remain separately gated. No remediation required for this `GO`.

### `W4-AUD-003` — low — benchmark comparison runs from the candidate checkout

- `scripts/harness/check-cli-benchmarks.sh` invokes the comparator with the
  candidate root, and the harness workflow runs the target from that checkout.
- The trusted policy controller is instead built from the trusted base.
- The comparator, script, workflow, and harness are protected paths that force
  Tier 3 classification and external owner/CODEOWNER review.

Disposition: the candidate/base asymmetry is recorded for future hardening; the
approved protected-control review is the current compensating control. No
remediation required for this `GO`.

### `W4-AUD-004` — note — terminal confirmation holds the mutation lock

`ConfirmMerge` performs a blocking terminal read while the repository mutation
lock is held; context cancellation is observed after that read returns. This
matches the pre-existing Tier 3 approval pattern and preserves post-confirmation
revalidation. It is a bounded interactive liveness/UX residual, not a safety or
Wave 4 regression.

### `W4-AUD-005` — note — approved paths were intentionally unused

Several brief-authorized test/adapter paths were not changed, including the
audit path until this record. Every actual changed path is authorized. Unused
authorization is not scope expansion and requires no action.

### `W4-AUD-006` — note — test-only transition helper omits stale merged recovery

`domain.TransitionAllowed` does not list `StateMerged → StateBuilding`, while
runtime derivation can return building for a stale merge receipt. The helper is
referenced only by lifecycle tests and no production path, so this is model
coherence debt without runtime effect.

### `W4-AUD-007` — note — identity separation is Tier 3 only

Tier 3 enforces distinct owner, implementer, and reviewer. Tier 1/2 allow the
same implementer/reviewer by the documented risk model, including in trusted
headless facts. This is intentional and not a Wave 4 defect.

### `W4-AUD-008` — note — execution evidence was inspected, not reproduced

The auditor did not rerun Go, Make, race, fuzz, actionlint, cross-build,
reproducibility, or benchmark commands. It confirmed four fuzz entry points and
recomputed the published benchmark arithmetic: parser `-4.63%` and snapshot
`+2.36%`, both inside the 10% threshold. Build hashes, host identity, timings,
and execution counts remain implementer-run technical evidence rather than
independent execution evidence.

## Read-only and actual-host boundary

The auditor used one authorized ephemeral Claude Code session with only `Read`,
`Glob`, and `Grep`; it spawned no subagents and used no Bash, Edit, Write,
browsing, web-fetch, or repository skill. The provider connection required for
that audit session is acknowledged; the auditor launched no additional Codex or
Claude product session and performed no provider-order actual-host trial.

It performed no test/build command, Git effect, Level 7 state write, merge,
deployment, release, publication, signing, notarization, hosted-CI query, or
Intel runtime trial. Immediately after the session, Codex `/root` compared the
pre/post worktree status, HEAD/tree, every local branch ref, remotes, and every
existing `.git/l7` file hash; the snapshots were identical.

This `GO` authorizes no merge, release, deployment, publication, provider-support
claim, or actual-host validation. It is the independent decision required before
the controller may advance to reviewed state.
