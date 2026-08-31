# Level 7 v1.0 Multi-Host Orchestration — Scope Successor

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration-successor` |
| Risk tier | `3` — provider authentication, agent execution, security testing, durable autonomy, protected controls, and release qualification |
| Status | `proposed`; integration, verification, and every later transition require fresh Product Owner approval bound to this exact addition commit |
| Base commit | `84bd69f90d366356b0ce1e1a392f906258f3de91` |
| Base tree | `84c2a105227f98089ea001f97473af79933bd743` |
| Historical approved brief | Commit `a9038285d612dec6d1c496c3f9a69fed9ca75f74`, tree `0512a917b7a029616e19af708fa346d9ab344002`; preserved, but its approval does not transfer |
| Implementation commit | `dc1fdb6a88d173a78e375bb8d54691c999d12e39` |
| Implementation tree | `a0f6355725067f89b63b621edcea4e682078be65` |
| Historical in-place correction | Commit `0047641d6b825c3272ee6fd57aca93885e6aa75f`, tree `e314a0437a9ae0bfbd2e1812bb07c26b60ea385f`; preserved, but not used as machine-bound approval authority |
| Integration target | Branch `codex/l7-v1-orchestration` at exact head `0047641d6b825c3272ee6fd57aca93885e6aa75f`; any drift requires a new decision |
| Development packages | Codex SHA-256 `efb0dbd1d6ed75d158860675d50332c69f30ed163999f35ebcbbaca1b3ae1eed`; Claude SHA-256 `58b3d6ad462abb2e25355b58a35621434144f6d89c4c270c54aad56faa8380fe` |
| Accountable owner | Anup Pandey, Product Owner; approval of this exact proposal commit is pending |
| Implementer | `codex-root` |
| Assurance | Tier 3 external owner approval, one exact-candidate verification record, one separately commissioned independent read-only audit, and owner GO before release |
| Next executable transition | Stop for explicit Product Owner approval of this exact proposal commit |

## Problem

The approved v1 orchestration brief at `a9038285d612dec6d1c496c3f9a69fed9ca75f74`
defined the intended product but omitted the controller-required exact
implementation file set. Implementation commit
`dc1fdb6a88d173a78e375bb8d54691c999d12e39` therefore remained blocked by
`BRIEF-005` and `SCOPE-002` despite passing technical verification.

The history-preserving in-place correction at
`0047641d6b825c3272ee6fd57aca93885e6aa75f` fixed the 56 implementation paths,
and that exact head passed `make verify` plus the reproducible v1 package gate.
It cannot support the required tracked verification record: the approved file
set omitted that record and its audit successor, while the controller binds
approval to a brief's first-add commit, still `a9038285…`. Adding evidence would
therefore fail closed on scope, artifact budget, or approval binding.

This proposal creates a controller-recognized brief addition while preserving
both prior brief versions and all implementation history. No old approval,
passing test, repository prose, or historical provider observation transfers as
approval of this successor.

## Scope

Preserve the original product intent: one default-OFF, local-first engine for
Codex, Claude Code, and configured compatible gateways; deterministic
capability routing; bounded gateway tools; private repository Sync; isolated
Cyber analysis; resumable Headless execution; macOS 13+ arm64/amd64 packages;
and hard stops before remote push, protected merge, release, publication, or
deployment.

The proposal commit is a direct child of `a9038285…`. It removes that brief from
the live proposal tree and adds this successor. Relative to the declared base,
only this successor brief exists; the historical brief remains reachable in
Git. No implementation, test, policy, workflow, skill, package, or configuration
byte changes in this proposal.

After fresh approval, integrate this proposal into the exact target head. The
integration may resolve only the old-brief delete/modify conflict by removing
the old live path and retaining this successor. Every implementation blob must
remain identical to `dc1fdb6a88d173a78e375bb8d54691c999d12e39`.

Fresh exact-head verification may then add the single declared verification
record. An independent read-only audit remains a later, separately authorized
transition and may add only the single declared audit record.

Explicit boundaries remain unchanged:

- all orchestration, Sync, Cyber active mode, and Headless effects default OFF;
- credentials remain native-host owned or environment/Keychain references;
- no unrestricted shell, filesystem, browser, network, or credential access;
- no Windows/Linux product support, hosted account, telemetry, daemon, remote
  memory, off-machine penetration target, or automatic secret ingestion;
- no verification record before fresh approval and a passing exact-head run;
- no audit without separate commissioning and no self-audit claim; and
- no push, signing, release, publication, merge to a protected branch, or
  deployment under this proposal.

The unrelated untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md` in the source checkout
remains untouched and outside every orchestration worktree.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-successor-audit.md`
- `cmd/l7-embed/main.swift`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `cmd/l7/orchestration_cli.go`
- `cmd/l7pack/main.go`
- `cmd/l7pack/main_test.go`
- `go.sum`
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
- `skills/l7-cyber/SKILL.md`
- `skills/l7-headless/SKILL.md`
- `skills/l7-onboard/SKILL.md`
- `skills/l7-sync/SKILL.md`

Modify:

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

Delete relative to the declared base: none. The proposal's removal of the old
brief path replaces a file that is absent from the base and remains preserved at
`a9038285d612dec6d1c496c3f9a69fed9ca75f74`.

## Acceptance criteria

1. The proposal is a direct child of `a9038285…`; its commit diff deletes only
   the old brief and adds only this successor, while base-to-proposal changes
   only this successor path.
2. Commits `a9038285…`, `dc1fdb6a…`, and `0047641d…` remain immutable and
   reachable. No reset, amend, rebase, force-push, or history rewrite occurs.
3. Fresh Product Owner approval binds this successor's exact addition commit.
   The old approval envelope remains historical and cannot authorize integration.
4. Integration starts only from exact target `0047641d…`; its final base diff
   contains this brief plus the exact 56 implementation paths above and no old
   live brief. Any other target or path fails closed.
5. Every implementation blob equals `dc1fdb6a…`; all new behavior remains
   default OFF and the original brief's product, security, compatibility,
   durability, packaging, performance, and rollback acceptance criteria remain
   mandatory.
6. `make verify` and `make v1-candidate-check` pass on the exact integrated head.
   The two package SHA-256 values remain identical to the table above.
7. The verification successor adds only its declared record, binds `PASS` to the
   exact integrated commit/tree, and makes no implementation change.
8. A separately commissioned independent reviewer may audit only the verified
   successor. The audit record binds `GO` or `NO_GO` to that exact verification
   commit/tree and cannot be produced by `codex-root`.
9. Team artifact budget remains exactly one brief, one verification record, and
   one audit record. Missing, extra, stale, self-issued, or mismatched authority
   and evidence fail closed.
10. No provider support promotion, protected merge, push, signing, release,
    publication, installation, deployment, or production effect occurs.

## Risks and mitigations

- **Stale authority:** use the new change ID and exact addition commit; reject
  the approval bound to `a9038285…` and the human-only correction at `0047641d…`.
- **Scope or artifact recurrence:** declare both evidence paths and all 56
  implementation paths; run team policy before and after each transition.
- **Implementation drift during integration:** require exact target identity,
  allow only the old-brief conflict resolution, and compare every non-governance
  blob to `dc1fdb6a…`.
- **History loss:** preserve all commits and use only additive commits plus an
  ordinary local integration merge after approval; never amend or rebase.
- **Fabricated assurance:** implementation tests are not independent review;
  commission a distinct read-only auditor only through a later authorization.
- **Unrelated user-state damage:** use exact worktree and pathspecs and recheck
  the source-checkout audit fingerprint at every transition.

## Rollback

Before integration, revert this proposal commit on its isolated branch to
restore the exact `a9038285…` tree. After integration, revert the local
integration merge with its implementation branch as mainline to restore the
exact `0047641d…` tree. After verification or audit, revert audit, verification,
and integration in reverse order.

Every rollback uses ordinary revert commits, stops on conflicts or extra paths,
preserves all historical records, and confirms the expected Git tree. Remote
reviews, checks, branches, releases, packages, and deployments remain unchanged
because this proposal authorizes none of those effects.

## Current transition

1. Commit this proposal-only replacement on
   `codex/l7-v1-orchestration-brief-successor`.
2. Stop for explicit Product Owner approval bound to that exact proposal commit.
3. Only after approval, record the matching external approval envelope,
   integrate into exact `0047641d…`, and rerun exact-head verification.
4. On `PASS`, commit only the declared verification record and stop.
5. Independent audit and every remote, signing, release, publication, protected
   merge, installation, or deployment effect require later authority.
