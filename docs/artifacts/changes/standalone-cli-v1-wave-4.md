# Standalone CLI v1 — Wave 4 Readiness and Confirmed Local Merge

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-4` |
| Risk tier | `3` |
| Status | `approved` |
| Base commit | `be898482ce632a8faf60f97022f740cd24585063` |
| Base tree | `3586155ea462d594ee75f9beeaee515025d3cbc7` |
| Accountable owner | `accountable-owner` — active user interaction |
| Implementer | `codex-root` |
| Feature flag | `features.local_lifecycle`, default `false` |

## Problem

Wave 3 can build, verify, and independently review an exact local candidate, but
it cannot decide release readiness from configuration-bound evidence, consume a
trusted headless CI decision, or advance a local target ref with immediate human
confirmation. Its audit also found residual process-containment, provider-mode,
evidence-binding, wording, and negative-path coverage gaps.

## Scope

Implement the approved Tier 3 Wave 4 slice of `CLI-008`–`CLI-010`:

- evaluate exact-candidate readiness from current Git, verification, review,
  audit, configuration, identity, and benchmark facts;
- accept one strict, bounded trusted-CI envelope in a side-effect-free headless
  evaluator;
- require active-terminal confirmation immediately before a controlled local,
  fast-forward-only, compare-and-swap ref update;
- reconstruct `ready` and `merged` lifecycle state from local records plus Git;
- close the six Wave 3 independent-audit findings; and
- extend adversarial, recovery, scale, architecture, and macOS CI coverage.

The existing feature remains default OFF and gates every new user-visible
behavior and mutation. This wave does not add `doctor`, P1/P2 features, remote
configuration, fetch/push, merge commits, rebase/reset, deployment, release,
publication, signing, notarization, automatic merge, global provider
configuration, or plugin installation. It creates no remote and makes no live
CI, Intel runtime, or provider claim without separately authorized evidence.

## Architecture and contracts

- Domain logic owns configuration-bound readiness, trusted-envelope facts,
  merge preconditions, and the `ready`/`merged` lifecycle transitions.
- Application orchestration recomputes readiness under the repository mutation
  lock and coordinates a confirmed merge transaction without owning Git or
  terminal mechanics.
- The authority adapter reads a full candidate SHA from the active terminal. It
  never accepts repository prose, stdin redirection, tests, or agent output as
  owner authority.
- The CI adapter decodes one size-bounded JSON document with strict fields and
  no network, prompt, provider, repository-write, or merge capability.
- The Git adapter validates explicit `refs/heads/*` targets, ancestry and
  worktree safety, then performs only `git update-ref <target> <candidate>
  <expected-old>` as the merge effect.
- The state adapter atomically stores bounded readiness and merge receipts under
  the Git common directory. Git remains canonical if interruption occurs after
  the ref update and before receipt persistence.
- The protected harness owns paired same-host benchmark comparison and the CI
  workflows exercise supported macOS architectures without asserting unobserved
  provider behavior.

### Command contract

- `l7 ready [--json]` requires a clean repository and current, exact candidate,
  tree, configuration digest, verification, independent `GO` review/audit,
  distinct identities, and required benchmark facts. It writes one bounded
  readiness receipt under the mutation lock; any changed fact makes it stale.
- `l7 ready --headless --json` reads one strict trusted-CI JSON envelope from
  bounded stdin and emits stable JSON plus a deterministic exit status. It uses
  the evaluator shipped by the trusted base and cannot prompt, write state,
  launch providers, access the network, or merge.
- `l7 merge --target <branch>` is interactive only. The target must resolve to
  one explicit local branch, equal the active readiness base, not be checked out
  by another worktree, and be an ancestor of the candidate. Under lock it shows
  old and new identities, requires the full candidate SHA, revalidates all
  facts, and advances the ref with compare-and-swap. It never checks out, pushes,
  rebases, resets, deletes, or creates a merge commit.
- `l7 status [--json]` reconstructs `ready` and `merged` from current receipts
  and Git, and names the next executable transition when a receipt is stale or
  an interrupted effect needs recovery.

### Evidence and compatibility contract

New verification evidence binds the exact configuration digest. Existing
schema-one evidence remains parseable for recovery but is stale for readiness.
Readiness and merge receipts are versioned, strictly decoded, bounded, and
replaceable only under the repository lock. No tracked `.l7/config.json`
migration is introduced.

The process runner applies a bounded `Cmd.WaitDelay` so a session-escaped
descendant retaining inherited pipes cannot retain the mutation lock forever.
The Claude reviewer removes Bash and uses an authenticated safe read-only mode;
actual semantics remain unclaimed until a separately authorized host trial.

## Exact implementation file set

Add:

- `docs/artifacts/changes/standalone-cli-v1-wave-4.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-4-verification.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-4-audit.md`
- `cmd/l7/readiness_test.go`
- `cmd/l7/actual_host_test.go`
- `internal/l7/domain/readiness.go`
- `internal/l7/domain/readiness_test.go`
- `internal/l7/app/readiness.go`
- `internal/l7/app/readiness_test.go`
- `internal/l7/adapter/authority/confirmation.go`
- `internal/l7/adapter/authority/confirmation_test.go`
- `internal/l7/adapter/ci/envelope.go`
- `internal/l7/adapter/ci/envelope_test.go`
- `internal/l7/adapter/git/merge.go`
- `internal/l7/adapter/git/merge_test.go`
- `internal/l7/adapter/state/readiness.go`
- `internal/l7/adapter/state/readiness_test.go`
- `internal/harness/benchgate/main.go`
- `internal/harness/benchgate/main_test.go`
- `scripts/harness/check-cli-benchmarks.sh`

Modify:

- `README.md`
- `Makefile`
- `.github/workflows/harness.yml`
- `.github/workflows/policy.yml`
- `cmd/l7/main.go`
- `cmd/l7/main_test.go`
- `cmd/l7/execution_test.go`
- `internal/l7/domain/result.go`
- `internal/l7/domain/result_test.go`
- `internal/l7/domain/lifecycle.go`
- `internal/l7/domain/lifecycle_test.go`
- `internal/l7/domain/execution.go`
- `internal/l7/domain/execution_test.go`
- `internal/l7/app/app.go`
- `internal/l7/app/app_test.go`
- `internal/l7/app/lifecycle.go`
- `internal/l7/app/lifecycle_test.go`
- `internal/l7/app/execution.go`
- `internal/l7/app/execution_test.go`
- `internal/l7/adapter/config/config.go`
- `internal/l7/adapter/config/config_test.go`
- `internal/l7/adapter/process/process.go`
- `internal/l7/adapter/process/process_unix.go`
- `internal/l7/adapter/process/process_test.go`
- `internal/l7/adapter/provider/contract.go`
- `internal/l7/adapter/provider/contract_test.go`
- `internal/l7/adapter/codex/adapter.go`
- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/codex/actual_host_test.go`
- `internal/l7/adapter/claude/adapter.go`
- `internal/l7/adapter/claude/adapter_test.go`
- `internal/l7/adapter/claude/actual_host_test.go`
- `internal/l7/adapter/git/repository_test.go`
- `internal/l7/adapter/state/evidence.go`
- `internal/l7/adapter/state/evidence_test.go`
- `internal/l7/presentation/output.go`
- `internal/l7/presentation/output_test.go`
- `harness/import-boundaries.tsv`
- `scripts/harness/check-import-boundaries.sh`

No other path is authorized. In particular, `.l7/config.json`, `go.mod`,
`go.sum`, skills/plugins, historical governance chains, global configuration,
Git remotes, and the user-owned untracked foundation audit remain unchanged.
Scope expansion or a renamed path requires a revised brief and fresh owner
approval.

## Acceptance criteria

1. Disabled configuration blocks readiness persistence and merge before any
   subprocess or Git mutation.
2. Readiness requires current verification bound to the exact configuration
   digest, exact candidate/tree, independent `GO`, audit, distinct identities,
   and benchmark facts.
3. Status reconstructs truthful `ready` and `merged` states after restart and
   treats legacy or changed evidence as stale with an executable recovery action.
4. Every merge effect requires immediate active-terminal confirmation of the
   full candidate SHA; repository text, redirected stdin, tests, or agent output
   cannot satisfy it.
5. Merge target, ancestry, worktree occupancy, expected old ref, readiness, and
   candidate are revalidated under lock; every race or divergence fails closed.
6. The only merge effect is an atomic local fast-forward `update-ref` compare-
   and-swap. There is no remote, checkout, push, merge-commit, rewrite, or delete
   capability.
7. Headless readiness uses strict bounded input and stable JSON/exit semantics,
   with no prompt, repository/product-state write, provider, network, or merge
   path.
8. Fake-provider end-to-end tests pass both provider orders through readiness
   and disposable local-merge recovery without credentials or network.
9. False readiness, self-review, `NO_GO`, stale/forged evidence, changed config,
   confirmation mismatch, divergence, concurrent ref changes, and interruptions
   all fail closed.
10. A session-escaped descendant retaining inherited pipes is bounded, and the
    Claude reviewer exposes no Bash tool in its safe read-only argument contract.
11. Paired same-host benchmark regressions above 10% fail unless explicitly
    accepted by the accountable owner.
12. Production dependencies remain zero, default-OFF behavior is preserved, and
    no provider, Intel, live-CI, release, or deployment claim is made without the
    separately authorized actual-host evidence.

## Risks and mitigations

- A ref update is durable and user-visible. Explicit target resolution, fast-
  forward ancestry, full-SHA confirmation, worktree checks, lock-scoped
  recomputation, and compare-and-swap contain the effect.
- Trusted-CI input is untrusted data to the candidate. A base-built evaluator,
  strict schema, bounded input, and a pure path prevent authority inference or
  candidate code execution in the decision step.
- Interruption can occur after the ref moves but before the receipt is written.
  Recovery consults Git and never auto-resets or rewrites the ref.
- Provider flags and authentication behavior can change. Adapter tests verify
  argv contracts; real support remains provisional until an authorized trial.
- Timing noise can create false regression decisions. The gate compares paired
  samples on the same host/toolchain and reports raw inputs and medians.
- No remote is configured, so workflow syntax tests cannot establish repository
  rules, hosted-runner availability, or external CI trust.

## Test, benchmark, and actual-host boundaries

Use domain/application tables, strict and fuzzed envelope decoders, terminal and
process helpers, temporary Git repositories/worktrees, fake providers, compiled
CLI tests, injected interruption/race failures, and import/effect-boundary
checks. Cover both provider orders, reviewer `NO_GO`, stale and forged evidence,
config drift, escaped descendants, confirmation mismatch, ref races,
divergence, receipt loss, 10,000-path repositories, and paired benchmark pass/
fail thresholds. Run formatting, lint, type checks, unit/integration tests, race
tests, fuzz/adversarial checks, reproducibility, cross-builds, and diff hygiene.

Real Codex→Claude or Claude→Codex sessions, provider network/model/billing,
actual GitHub Actions or repository rules, Intel runtime, signing/notarization,
release, publication, deployment, and merge require their own explicit gates.
Any later provider trial must use disposable no-remote repositories, bounded
budgets, exact executable/version/config digests, and must not retain credentials
or full transcripts. Build-tagged actual-host tests remain skipped by default.

## Rollback

Before merge, revert the small conventional commits in reverse order. The
feature remains default OFF; there is no tracked configuration migration,
dependency, remote, deployment, or publication. After any ref advance, never
reset automatically: preserve the receipt or reconstruct from Git and require a
separately approved recovery/revert.

## Commit sequence

Planned commits:

1. `docs(cli): define standalone cli wave four`
2. `test(harness): reserve wave four effect boundaries`
3. `fix(cli): close wave three execution residuals`
4. `feat(cli): bind exact candidate readiness`
5. `feat(cli): evaluate trusted headless readiness`
6. `feat(cli): confirm atomic fast-forward merges`
7. `test(cli): harden false-ready and scale gates`
8. `ci(cli): exercise macos release architectures`
9. `docs(cli): document wave four candidate`
10. one verification-only commit, followed by one independent audit-only commit
    after its separate authorization.
