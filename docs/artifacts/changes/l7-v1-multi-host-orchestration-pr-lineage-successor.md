# Level 7 v1.0 Multi-Host Orchestration — PR-Base Lineage Successor

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration-pr-lineage-successor` |
| Risk tier | `3` |
| Status | `proposed`; no integration authority exists until the Product Owner approves this exact brief commit |
| Base commit | `84bd69f90d366356b0ce1e1a392f906258f3de91` |
| Base tree | `84c2a105227f98089ea001f97473af79933bd743` |
| Integration target | `codex/l7-v1-orchestration` at commit `6e41f554b5447ffe4c081b000f3f143371cddffd`, tree `58b1b6065c046af8e87fdbcf5c2ea7151e2d2472` |
| Remote context | PR #15 targets the base above and remains at head `f10486aee9b361f9b2ec7e92782f30247b7b7a32`; no remote mutation is authorized |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root`; designated hosted PR author remains `anup19950725` |
| Trigger | Trusted-controller `GIT-003` when PR base `84bd69f…` is compared with the active brief base `f10486a…` |
| Assurance | Tier 3 team: fresh external owner approval, one exact-candidate verification record, one separately commissioned independent read-only audit, then a separate owner decision |
| Next executable transition | Stop for explicit Product Owner approval bound to this exact proposal commit |

## Problem

The cumulative v1.0 orchestration candidate is implemented, locally verified,
and independently reviewed through exact audit commit `6e41f554…`. Its latest
module-bootstrap remediation correctly declares base `f10486a…`, so local
policy evaluation can reach `reviewed` state when it adopts that brief base.

Hosted trusted policy instead evaluates PR #15 with explicit base
`84bd69f90d366356b0ce1e1a392f906258f3de91`. The mismatch fails closed as
`GIT-003`. The current audited `GO` covers the bootstrap repair; it does not
close this separately declared lineage blocker or establish merge readiness.

Editing an approved brief, weakening the controller, rewriting history, or
treating an old approval as authority for a revised scope is prohibited. A
fresh base-rooted successor is required.

## Scope

This proposal is created on an isolated branch rooted at the exact PR base. Its
proposal commit adds only this brief. It changes no implementation, policy,
workflow, test, package, prior record, external envelope, PR, or hosted state.

After fresh approval, one local integration merge may combine this proposal
with only the exact integration target. The merge must:

- preserve all 63 non-governance paths byte-for-byte from target `6e41f554…`;
- add this successor as the only live brief and reserve only its declared
  verification and audit records;
- retire exactly the nine superseded live governance paths listed below while
  preserving every prior commit and record in Git history; and
- leave policy code, PR base, product behavior, default-OFF boundaries,
  dependency identities, package outputs, and remote state unchanged.

The resulting base-visible candidate must contain exactly this successor's
three permitted records plus the 63 preserved non-governance paths. The trusted
controller must evaluate that candidate against explicit PR base `84bd69f…`
without `GIT-003`, scope expansion, artifact-budget failure, or stale-authority
acceptance.

This change does not supply the required trusted Tier 3 label, hosted exact-head
checks, GitHub reviews, owner `GO`, or merge authority. Those remain later,
separately authorized transitions.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-pr-lineage-successor-audit.md`
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

- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation-audit.md`
- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation-audit.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor-audit.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor.md`

The nine retirement paths are absent from both the declared base and intended
final tree, so they do not appear as base-visible changes. They are declared
here to constrain the integration merge exactly and prevent silent historical
or artifact-chain drift.

## Acceptance criteria

1. The proposal commit has exact parent/base `84bd69f…`, adds only this brief,
   and leaves both the implementation worktree and remote state unchanged.
2. Fresh Product Owner approval names this change ID and exact brief commit.
   No prior approval, verification, audit, review, or owner decision transfers.
3. Integration starts only from exact target commit `6e41f554…` and exact target
   tree `58b1b606…`; any drift requires a new proposal or owner decision.
4. The integration commit has the target and approved proposal as its two
   parents, removes only the nine declared live governance paths, and performs
   no rebase, reset, amend, squash, force update, or history rewrite.
5. Relative to base `84bd69f…`, the integrated tree changes exactly the new
   brief plus the 63 declared non-governance paths. Every non-governance mode
   and blob equals target `6e41f554…`; the old nine records remain reachable in
   immutable history but are not extended as live authority chains.
6. The active controller is not modified. An exact trusted-controller dry run
   with explicit `--base 84bd69f…`, the integrated head, Tier 3, and team mode
   reports no `GIT-003`, `SCOPE-*`, `ART-*`, or stale-authority finding.
7. `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` and
   `make v1-candidate-check GO_VERSION=1.26.7` pass at the exact integrated
   head using disposable test-owned roots. Existing frozen v0.1.1 and v1.0-dev
   package identities remain unchanged.
8. If verification passes, one successor verification commit adds only the
   declared verification record and binds `PASS` to the exact integrated
   commit and tree. It contains no implementation or governance-control drift.
9. An independent read-only audit is commissioned only through later explicit
   authority. Its sole record binds `GO` or `NO_GO` to the exact verification
   commit and tree; `codex-root`, Anup Pandey, and the PR author cannot satisfy
   the independent reviewer role.
10. The final Tier 3 artifact budget is exactly this brief, its verification
    record, and its audit record. Old external envelopes remain untouched and
    cannot authorize this successor.
11. PR #15, failed run `33427651169`, required risk-tier labeling, hosted
    exact-head checks, reviews, owner `GO`, and promotion remain truthful,
    separate gates. Local success cannot be credited as hosted success.
12. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and outside every scope, commit, index, worktree operation, and rollback.

## Risks and mitigations

- **Implementation drift during integration:** bind the exact target commit and
  tree; compare all 63 non-governance modes and blobs before testing or commit.
- **Historical-record loss:** use an additive two-parent merge, retire only the
  nine named live paths, and verify every old commit remains reachable.
- **Deprecated-chain reuse:** use a fresh change ID, brief addition, approval,
  verification, and audit; never edit or extend an old record.
- **Policy weakening disguised as remediation:** exclude policy and trusted
  workflow changes; prove the existing controller accepts the correct base.
- **False merge-readiness:** keep the trusted risk label, hosted checks, GitHub
  reviews, owner decision, push, and PR mutation outside this local repair.
- **Unrelated-state damage:** use exact pathspecs and recheck the original
  checkout's untracked artifact fingerprint after every transition.

## Rollback

Before integration, abandon or ordinarily revert only this proposal commit on
its isolated branch; the implementation branch remains at exact target
`6e41f554…`. After integration, ordinarily revert the integration merge with
the implementation parent as mainline, then confirm restoration of target tree
`58b1b606…`. After later verification or audit, revert audit, verification, and
integration in reverse order.

Every rollback stops on conflicts or unexpected paths. Historical commits,
external envelopes, PR #15, run `33427651169`, packages, releases, credentials,
user installations, and production state remain untouched.

## Current transition

1. Commit only this proposal on `codex/l7-v1-pr-lineage-successor` as a direct
   child of exact PR base `84bd69f…`.
2. Stop for explicit Product Owner approval bound to that exact brief commit.
3. Only after approval, record the matching external approval envelope,
   perform the one declared local integration merge, and run bounded exact-head
   verification including the explicit PR-base controller check.
4. If verification passes, commit only the declared verification record and
   stop for separate authority to commission an independent read-only audit.
5. Push, PR mutation or labeling, hosted execution, review, owner `GO`, merge,
   signing, release, publication, installation, and deployment remain
   separately gated and unauthorized.
