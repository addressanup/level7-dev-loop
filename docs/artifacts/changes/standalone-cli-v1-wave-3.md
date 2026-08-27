# Standalone CLI v1 — Wave 3 Verification and Provider Execution

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-3` |
| Risk tier | `3` |
| Status | `approved` |
| Base commit | `9021c69d6c11d71c47510983b157ebc54db6cbf0` |
| Accountable owner | `accountable-owner` — active user interaction |
| Implementer | `codex-root` |
| Feature flag | `features.local_lifecycle`, default `false` |

## Problem

The Wave 2 preview can reconstruct a bounded local change but cannot execute the
repository's own checks, create a controlled candidate commit, invoke either
supported coding client, survive a running subprocess interruption, or bind an
independent review to the exact verified candidate.

## Scope

Implement the approved Tier 3 Wave 3 slice of `CLI-004`–`CLI-007`:

- run explicit technology-neutral verification argv with bounded environment,
  output, timeout, cancellation, and exact-candidate freshness;
- create conventional Git commits containing only the validated declared paths
  against an unchanged expected HEAD;
- add one provider-neutral contract plus isolated Codex and Claude adapters;
- distinguish installed, compatible, degraded, and unavailable capabilities;
- supervise inherited process groups and preserve recoverable state on failure;
- orchestrate `implement → verify → review` in both provider orders; and
- reconstruct implementer, verification, and reviewer bindings after restart.

The existing `features.local_lifecycle` default-OFF gate covers every Wave 3
effect. Each effect additionally requires an explicit `run`, `verify`, or
`review` command. This wave does not add `doctor`, readiness, merge, CI policy,
deployment, provider installation, global configuration, or automatic fallback.
It makes no provider support claim without separately authorized actual-host
evidence.

### Command contract

- `l7 run --agent codex|claude --message <conventional-subject> [--json]`
  requires stable intake and, for Tier 3, current external terminal approval.
  It launches one bounded implementer, rejects provider-created commits or scope
  expansion, then creates one exact-path commit against the expected HEAD.
- `l7 verify [--json]` runs the non-empty configured verification array in order
  as exact argv without a shell. Results bind to the current commit and tree and
  become stale after any candidate change. Tier 3 creates the sole tracked
  verification record; Tier 1/2 assurance remains local.
- `l7 review --agent codex|claude [--json]` requires current verification. Tier
  3 requires the provider distinct from the implementer and runs read-only. Its
  decision binds the verified candidate; `NO_GO`, mutation, or stale identity
  returns to `building`.
- `l7 status [--json]` reconstructs Wave 3 facts but stops at `reviewed` and
  truthfully reports that readiness and merge remain unavailable until Wave 4.

### Adapter contract

- The process adapter alone supervises arbitrary child processes. It accepts an
  absolute validated executable, explicit argv/stdin/cwd, an allowlisted
  environment, and configured limits; it returns one bounded terminal result.
- The Git adapter owns Git invocation and commit effects. Commits use an expected
  old HEAD, explicit path set, bounded conventional subject, normal repository
  hooks, and post-effect identity/scope checks. It never merges or rewrites
  history and never destroys recovery data automatically.
- Provider adapters expose the same probe, implement, and read-only review
  contract while keeping provider flags and event schemas isolated. Installation
  or help output alone is not compatibility evidence.
- Provider output is untrusted observation data. Git supplies changed paths and
  candidate identity; external interaction supplies owner authority; recorded
  implementer/provider identity supplies the independence precondition.

## Exact implementation file set

Add:

- `docs/artifacts/changes/standalone-cli-v1-wave-3.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-3-verification.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-3-audit.md`
- `cmd/l7/execution_test.go`
- `internal/l7/domain/execution.go`
- `internal/l7/domain/execution_test.go`
- `internal/l7/app/execution.go`
- `internal/l7/app/execution_test.go`
- `internal/l7/adapter/authority/approval.go`
- `internal/l7/adapter/authority/approval_test.go`
- `internal/l7/adapter/process/process.go`
- `internal/l7/adapter/process/process_unix.go`
- `internal/l7/adapter/process/process_test.go`
- `internal/l7/adapter/verify/runner.go`
- `internal/l7/adapter/verify/runner_test.go`
- `internal/l7/adapter/provider/contract.go`
- `internal/l7/adapter/provider/contract_test.go`
- `internal/l7/adapter/codex/adapter.go`
- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/codex/actual_host_test.go`
- `internal/l7/adapter/claude/adapter.go`
- `internal/l7/adapter/claude/adapter_test.go`
- `internal/l7/adapter/claude/actual_host_test.go`
- `internal/l7/adapter/git/commit.go`
- `internal/l7/adapter/git/commit_test.go`
- `internal/l7/adapter/state/evidence.go`
- `internal/l7/adapter/state/evidence_test.go`

Modify:

- `README.md`
- `cmd/l7/main.go`
- `cmd/l7/main_test.go`
- `internal/l7/domain/result.go`
- `internal/l7/domain/result_test.go`
- `internal/l7/domain/lifecycle.go`
- `internal/l7/domain/lifecycle_test.go`
- `internal/l7/app/app.go`
- `internal/l7/app/app_test.go`
- `internal/l7/app/lifecycle.go`
- `internal/l7/app/lifecycle_test.go`
- `internal/l7/adapter/config/config.go`
- `internal/l7/adapter/config/config_test.go`
- `internal/l7/presentation/output.go`
- `internal/l7/presentation/output_test.go`
- `harness/import-boundaries.tsv`
- `scripts/harness/check-import-boundaries.sh`

No other path is authorized. In particular, `.l7/config.json`, `Makefile`, CI
workflows, plugin/skill files, dependencies, global Codex configuration, and
historical governance records remain unchanged. Any additional or renamed path
requires a revised brief and fresh owner approval.

## Acceptance criteria

1. Disabled configuration blocks all Wave 3 effects before subprocess or Git
   mutation.
2. Verification uses exact configured argv and bounded process controls; its
   result is current only for the exact commit/tree and configuration snapshot.
3. Cancellation, timeout, output overflow, invalid structured output, and
   inherited child processes cannot produce success or retain the mutation lock.
4. Controlled commits contain only declared paths, require the expected HEAD,
   retain hooks, and cannot merge, amend, reset, rebase, or rewrite history.
5. Codex and Claude meet one neutral contract without claiming flag, permission,
   session, sandbox, or version parity. Unknown required semantics block.
6. Fake-provider end-to-end tests pass for Codex→Claude and Claude→Codex with
   interruption between every adjacent state and no network or credentials.
7. Tier 3 enforces distinct owner, implementer, and reviewer identities plus one
   brief, one verification record, and one independent audit record.
8. Candidate mutation, review mutation, self-review, `NO_GO`, stale evidence,
   scope expansion, executable replacement, and concurrent HEAD/index change
   fail closed and return an executable recovery action.
9. Status never reports `ready`; no Wave 3 package or command exposes merge.
10. Production dependencies remain zero. Relevant tests, fuzz/adversarial
    coverage, benchmarks, `make verify`, import/effect boundaries, cross-builds,
    diff hygiene, separately authorized actual-host evidence, and an independent
    read-only security/process audit pass before a `GO` decision.

## Risks and mitigations

- Repository verification is explicitly authorized arbitrary code execution;
  bounds and supervision are not represented as a general OS sandbox.
- Provider path-level mutation prevention depends on tested provider sandbox
  behavior; Git postconditions refuse scope expansion but do not erase it.
- Process groups contain inherited descendants. Deliberate daemon/session escape
  is a residual macOS limitation that the audit must evaluate without overclaim.
- Executable replacement, ambient-path ambiguity, provider updates, hostile
  events, output flooding, prompt injection, hooks, and Git races fail closed.
- Level 7 makes no network call and passes no API keys. Real providers retain
  custody of their own credentials, network access, and billing behavior.

## Test and benchmark strategy

Use domain/application tables, temporary Git repositories, helper subprocesses,
fake providers, strict parser tests, fuzzing, compiled CLI tests, and adversarial
failure injection. Cover hangs, signals, floods, invalid UTF-8, child processes,
secret environment stripping, hook failure, index/HEAD races, stale bindings,
and both provider orders. Retain the current 10,000-path Git benchmarks and add
bounded supervisor/event/orchestration measurements. A same-host/toolchain
regression above 10% requires explicit acceptance; provider wall time is
observational rather than a release-quality proxy.

Actual-host validation is separately authorized: inert executable/interface
probes establish detection only; real model/network trials use disposable
fixtures, exercise both orders, capture no credentials or full transcripts, and
bind claims only to the observed executable digest/version and macOS/architecture.
Missing or incompatible providers block the corresponding support claim.

## Rollback

Revert the small conventional commits in reverse order. The feature remains
default OFF and introduces no configuration migration, dependency, remote state,
merge, deployment, or global installation. Failed operations retain user work
and exact recovery diagnostics rather than resetting it. Remove only validated
Wave 3 runtime records from their exact Git-common-directory product paths.
