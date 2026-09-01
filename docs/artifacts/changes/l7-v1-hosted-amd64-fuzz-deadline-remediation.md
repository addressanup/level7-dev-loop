# Level 7 v1.0 Hosted amd64 Fuzz Deadline Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-amd64-fuzz-deadline-remediation` |
| Risk tier | `3` |
| Status | `proposed`; no integration or implementation authority exists until the Product Owner approves this exact brief commit |
| Base commit | `84bd69f90d366356b0ce1e1a392f906258f3de91` |
| Base tree | `84c2a105227f98089ea001f97473af79933bd743` |
| Integration target | `codex/l7-v1-orchestration` at commit `53bbe3b106f6bcea89078e474c2a63daf6c28d56`, tree `643de970a7fb68a66ccc1eed90871e34ea0a4358` |
| Remote context | PR #15 is open at the exact target head and base above with only `l7-risk-tier-3`; Harness run `33481638065` and workflow-triggered Trusted policy run `33482573666` failed; no remote mutation is authorized |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root`; designated hosted PR author remains `anup19950725` |
| Trigger | Exact-head Harness job `CLI macOS 15 (amd64)` (`99772342269`) failed in `FuzzStrictConfigurationDecode` with `context deadline exceeded` after the fuzz worker stopped making progress |
| Assurance | Tier 3 team: fresh external owner approval, one exact-candidate verification record, one separately commissioned independent read-only audit, then a separate owner decision |
| Next executable transition | Stop for explicit Product Owner approval bound to this exact proposal commit |

## Problem

Exact candidate `53bbe3b…` passed the baseline Go job, shadow Go job,
macOS arm64 job, and paired benchmark job in Harness run `33481638065`.
The required macOS amd64 job failed closed while running the existing offline
fuzz gate. `FuzzStrictConfigurationDecode` completed 35,808 executions, then
made no progress for approximately three seconds and returned
`context deadline exceeded` when the five-second wall-clock fuzz budget ended.

The successful exact-head baseline, shadow, and arm64 logs show that the same
target can stop making progress after as few as three to seven executions and
still exit successfully. The evidence therefore does not support treating this
as an amd64 product failure or solving it with an unchanged rerun. It exposes a
load-sensitive interaction between production-bound mutated configuration payloads
and wall-clock fuzz cancellation. Swallowing the error, retrying until green,
raising timeouts alone, or reducing the eight-target inventory would weaken the
gate rather than repair it.

The current PR-base lineage successor, its verification, independent `GO`, and
owner decision remain exact historical evidence for target `53bbe3b…`; none
transfers to changed test or harness bytes. Trusted policy also reports the
separate missing hosted review envelope (`AUTH-001`). This remediation does not
supply that review or override any failed hosted result.

## Scope

This proposal is created on an isolated branch rooted at the exact PR base. Its
proposal commit adds only this brief and changes no implementation, test,
workflow, policy, prior record, external envelope, PR, or hosted state.

After fresh approval, one local integration merge may combine this proposal
with only exact target `53bbe3b…`. Relative to that target, implementation may
change only:

- `internal/l7/adapter/orchestrationconfig/config_test.go`, to put a fixed
  fuzz-only bound on mutated payloads while adding deterministic production-
  boundary cases; and
- `scripts/harness/check-l7-fuzz.sh`, to give the affected target a fixed
  source-controlled execution-count budget instead of wall-clock cancellation.

The integration must preserve all 61 other non-governance modes and blobs
byte-for-byte, add this successor as the only live brief, reserve only its
declared verification and audit records, and retire only the three current live
PR-base-lineage governance paths listed below. Every retired record and commit
must remain reachable in Git history. Production decoding, product behavior,
the eight-target inventory, offline isolation, serialized fuzzing, global
safety timeout, workflow, trusted controller, dependency identities, packages,
default-OFF boundaries, and remote state are outside implementation scope.

The intended repair must not accept a timeout as success. It must retain
adversarial coverage at the production configuration size boundary through
deterministic cases while keeping random mutation iterations individually
bounded. No retry loop, conditional architecture bypass, environment override,
or hosted-result substitution is permitted.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation-audit.md`
- `cmd/l7-embed/main.swift`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `cmd/l7/orchestration_cli.go`
- `cmd/l7pack/main.go`
- `cmd/l7pack/main_test.go`
- `go.sum`
- `internal/harness/v1candidate/main.go`
- `internal/harness/v1candidate/main_test.go`
- `internal/l7/adapter/claude/worker.go`
- `internal/l7/adapter/codexapp/discovery.go`
- `internal/l7/adapter/codexapp/discovery_test.go`
- `internal/l7/adapter/codexapp/worker.go`
- `internal/l7/adapter/codexapp/worker_test.go`
- `internal/l7/adapter/cyber/cyber.go`
- `internal/l7/adapter/cyber/cyber_test.go`
- `internal/l7/adapter/discovery/discovery.go`
- `internal/l7/adapter/discovery/discovery_test.go`
- `internal/l7/adapter/gateway/client.go`
- `internal/l7/adapter/gateway/gateway_test.go`
- `internal/l7/adapter/gateway/worker.go`
- `internal/l7/adapter/headless/headless.go`
- `internal/l7/adapter/headless/headless_test.go`
- `internal/l7/adapter/headlessworker/executor.go`
- `internal/l7/adapter/headlessworker/executor_test.go`
- `internal/l7/adapter/memory/apple.go`
- `internal/l7/adapter/memory/memory.go`
- `internal/l7/adapter/memory/memory_test.go`
- `internal/l7/adapter/orchestrationconfig/config.go`
- `internal/l7/adapter/orchestrationconfig/config_test.go`
- `internal/l7/adapter/state/orchestration.go`
- `internal/l7/adapter/state/orchestration_test.go`
- `internal/l7/adapter/toolbroker/broker.go`
- `internal/l7/adapter/toolbroker/broker_test.go`
- `internal/l7/domain/orchestration.go`
- `internal/l7/domain/routing.go`
- `internal/l7/domain/routing_test.go`
- `scripts/harness/bootstrap-modules.sh`
- `scripts/harness/check-bootstrap-modules.sh`
- `scripts/harness/check-l7-fuzz.sh`
- `scripts/harness/check-v1-conformance.sh`
- `skills/l7-cyber/SKILL.md`
- `skills/l7-headless/SKILL.md`
- `skills/l7-onboard/SKILL.md`
- `skills/l7-sync/SKILL.md`

Modify:

- `.github/workflows/harness.yml`
- `CHANGELOG.md`
- `Makefile`
- `README.md`
- `cmd/l7/main.go`
- `go.mod`
- `harness/import-boundaries.tsv`
- `internal/harness/distribution/main.go`
- `internal/l7/adapter/claude/adapter.go`
- `internal/l7/adapter/claude/adapter_test.go`
- `internal/l7/adapter/codex/adapter.go`
- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/process/process.go`
- `internal/l7/adapter/process/process_test.go`
- `internal/l7/app/app.go`
- `internal/l7/domain/result.go`
- `scripts/harness/check-import-boundaries.sh`
- `skills/l7-next/SKILL.md`

Delete from the live integration tree only; preserve in Git history:

- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor-audit.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor.md`

The three retirement paths are absent from both the declared base and intended
final tree, so they do not appear as base-visible changes. They are declared to
constrain the integration turnover and prevent a deprecated governance chain
from being extended.

## Acceptance criteria

1. The proposal commit has exact parent/base `84bd69f…`, adds only this brief,
   and leaves the implementation worktree and all remote state unchanged.
2. Fresh Product Owner approval names this change ID and exact brief commit.
   No prior approval, verification, audit, review, or owner decision transfers.
3. Integration starts only from exact target commit `53bbe3b…` and tree
   `643de970…`; any target drift requires a fresh proposal or owner decision.
4. The integration commit has the target and approved proposal as its two
   parents. Relative to the target it changes exactly six paths: adds this
   brief, removes the three declared live governance records, and modifies only
   the strict configuration fuzz target and fuzz driver named in Scope. It
   performs no rebase, reset, amend, squash, force update, or history rewrite.
5. Relative to base `84bd69f…`, the integrated tree changes exactly 64 paths:
   this brief plus the 63 declared non-governance paths. The two remediation
   blobs differ from target only as approved; every other non-governance mode
   and blob equals target `53bbe3b…` exactly.
6. The strict configuration fuzz target uses an exact source-controlled 64 KiB
   fuzz-only payload ceiling below the production `MaxBytes` boundary.
   Deterministic cases still exercise valid and invalid inputs at that exact
   256 KiB boundary and prove oversized file admission remains rejected. No
   production decoder, validator, limit, schema, or behavior changes.
7. The fuzz driver keeps all eight exact targets, pinned offline toolchains,
   clean archived source, disposable cache/corpus/temporary/telemetry roots,
   `CGO_ENABLED=1`, `-parallel=1`, and the two-minute hard timeout. The affected
   target uses the literal Go `-fuzztime=10000x` execution budget;
   it has no wall-clock success race, retry, skip-on-error, environment override,
   or architecture exception. The other seven target budgets are unchanged.
8. Focused diagnostics pass the affected fuzz target at least 20 consecutive
   times with no stalled-progress deadline, then the unchanged integrated head
   passes `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` and
   `make v1-candidate-check GO_VERSION=1.26.7` using disposable test-owned
   roots. Local evidence is not credited as hosted amd64 evidence.
9. The existing trusted controller and policy workflow remain byte-identical.
   With a fresh matching external approval envelope, an explicit Tier 3 team
   dry run against PR base `84bd69f…` reports no `GIT-*`, `SCOPE-*`, `ART-*`,
   stale-authority, or policy-weakening finding.
10. If verification passes, one verification commit adds only this successor's
    declared verification record and binds `PASS` to the exact integrated
    commit and tree. It cannot reuse predecessor evidence.
11. An independent read-only audit is commissioned only through later explicit
    authority. Its sole record binds `GO` or `NO_GO` to the exact verification
    commit and tree; `codex-root`, Anup Pandey, and PR author `anup19950725`
    cannot satisfy the independent reviewer role.
12. The final Tier 3 artifact budget is exactly this brief, its verification
    record, and its audit record. The three retired records and all earlier
    records remain immutable in history and cannot authorize this successor.
13. Existing orchestration behavior, default-OFF features, dependency files,
    package contents, unsigned release-blocked provenance, and frozen v0.1.1
    identities remain unchanged.
14. A later separately authorized push requires fresh exact-head baseline,
    shadow, macOS arm64/amd64, benchmark, and trusted-policy results. GitHub
    owner and independent-review envelopes must bind that later exact head;
    current `AUTH-001` remains unresolved until then.
15. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched,
    unstaged, and outside every scope, commit, worktree operation, and rollback.

## Risks and mitigations

- **Coverage dilution disguised as stability:** retain all eight targets, use a
  fixed minimum execution count for the affected target, and cover the full
  production size boundary with deterministic cases.
- **Timeout masking:** reject retry-until-green, swallowed exit codes, raised
  timeout-only fixes, environment overrides, and architecture bypasses; every
  real fuzz or boundary failure remains terminal.
- **Implementation drift:** bind the exact target and permit only two target
  blobs to change; compare the other 61 non-governance blobs and modes before
  testing or commit.
- **Historical-record loss:** use an additive two-parent integration, retire
  only the three named live records, and verify all predecessors remain
  reachable in immutable history.
- **Stale assurance:** require fresh approval, verification, audit, owner
  decision, hosted checks, and exact-head GitHub reviews after changed bytes.
- **False merge-readiness:** preserve Harness run `33481638065` and policy run
  `33482573666` as failures; local success cannot override either.
- **Unrelated-state damage:** use exact pathspecs and recheck the original
  checkout's untracked artifact fingerprint after every transition.

## Rollback

Before integration, ordinarily revert only this proposal commit; the
implementation branch remains at exact target `53bbe3b…` and no remote state is
changed. After integration, ordinarily revert later audit and verification
commits first, then revert the integration merge with the target parent as
mainline and confirm restoration of tree `643de970…`.

Every rollback stops on conflict or unexpected path. Historical commits,
retired records, external envelopes, PR #15, failed workflow runs, packages,
credentials, user installations, releases, and production state remain
untouched.

## Current transition

1. Commit only this proposal on
   `codex/l7-v1-hosted-amd64-fuzz-deadline-remediation` as a direct child of
   exact PR base `84bd69f…`.
2. Stop for explicit Product Owner approval bound to that exact brief commit.
3. Only after approval, record the matching external approval envelope,
   perform the one declared local integration and two-file remediation, and run
   bounded exact-head verification.
4. If verification passes, commit only the declared verification record and
   stop for separate authority to commission an independent read-only audit.
5. Hosted rerun, push, PR mutation or review, owner `GO`, protected-branch
   merge, signing, release, publication, installation, and deployment remain
   separately gated and unauthorized.
