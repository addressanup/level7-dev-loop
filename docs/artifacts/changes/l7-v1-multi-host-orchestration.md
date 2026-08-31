# Level 7 v1.0 Multi-Host Orchestration — Change Brief

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration` |
| Risk tier | `3` — provider authentication, agent execution, security testing, durable autonomy, plugin permissions, and stable release |
| Status | `superseding proposal`; verification and every later transition require fresh Product Owner approval bound to this exact revision commit |
| Base commit | `84bd69f90d366356b0ce1e1a392f906258f3de91` |
| Base tree | `84c2a105227f98089ea001f97473af79933bd743` |
| Prior approved brief | Commit `a9038285d612dec6d1c496c3f9a69fed9ca75f74`, tree `0512a917b7a029616e19af708fa346d9ab344002`; preserved in Git history, but its approval does not transfer to this revision |
| Implementation commit | `dc1fdb6a88d173a78e375bb8d54691c999d12e39` |
| Implementation tree | `a0f6355725067f89b63b621edcea4e682078be65` |
| Target | Stable `v1.0.0` for macOS 13+ on `arm64` and `amd64` |
| Development packages | Codex SHA-256 `efb0dbd1d6ed75d158860675d50332c69f30ed163999f35ebcbbaca1b3ae1eed`; Claude SHA-256 `58b3d6ad462abb2e25355b58a35621434144f6d89c4c270c54aad56faa8380fe` |
| Accountable owner | Anup Pandey, Product Owner; approval of this exact revision commit is pending |
| Implementer | `codex-root` |
| Assurance | Tier 3 external owner approval, exact-candidate verification, independent read-only audit, and owner GO before release |
| Next executable transition | Stop for explicit Product Owner approval of this exact revision commit; only then run exact-head verification and create one bound verification record |

## Problem

Level 7 v0.1.1 is an instruction-only plugin. The repository contains a local
CLI foundation and provider experiments, but the stable product cannot discover
and compare authenticated hosts, route work by capability and effort, operate
API-compatible models through a bounded tool broker, maintain durable codebase
memory, run isolated security confirmation, or continue a multi-wave objective
across provider sessions.

Users currently have to select hosts and carry context manually. Exact client
version pins also turn compatible host upgrades into degraded states. Copying a
handoff through a UI is neither headless nor crash-safe, and active security
testing without a real isolation boundary would create unacceptable host risk.

The Product Owner approved the prior brief at `a9038285d612dec6d1c496c3f9a69fed9ca75f74`,
and implementation was committed at `dc1fdb6a88d173a78e375bb8d54691c999d12e39`.
The prior brief omitted the controller-required exact implementation file-set
section, so `make verify` correctly failed closed with `BRIEF-005` and
`SCOPE-002`. This revision changes governance prose only, preserves the prior
brief and approval as historical records, and requires fresh approval before
verification or any later transition.

## Outcome

Ship one bundled, local-first orchestration engine for Codex, Claude Code, and
configured OpenAI Responses/Anthropic Messages-compatible gateways. The engine
discovers truthful capabilities, selects an appropriate qualified model and
effort, exposes explainable routing, maintains private repository memory, runs
bounded security analysis, and can execute an approved Tier 1/2 feature-wave
manifest through local merge while stopping before publication, release, or
deployment.

The stable plugin remains safe by default: all new product behavior is OFF until
explicit onboarding, raw state stays under `.git/l7`, credentials remain owned
by the native host or an environment/Keychain reference, and unsupported or
uncertain authority fails closed.

## Scope

### Runtime and plugin surfaces

- Produce one Go `l7` engine with CLI and local MCP modes, packaged for macOS
  `arm64` and `amd64`, plus a small Swift helper that uses Apple Natural
  Language sentence embeddings.
- Add skills `l7-onboard`, `l7-sync`, `l7-cyber`, and `l7-headless` to both host
  packages. Preserve `l7-next` as the backward-compatible read-only conductor.
- Add CLI/MCP contracts for onboarding, provider discovery and probing, route
  explanation, memory synchronization and query, Cyber audit/export/remediation
  planning, and Headless plan/start/status/resume/cancel.
- Preserve strict `.l7/config.json` schema 1. Add strict, tracked,
  default-OFF `.l7/orchestration.json`; store derived or sensitive state only
  under `.git/l7`.
- Change plugin permission declarations truthfully for bundled executables,
  local MCP, and explicitly configured network use. Do not add telemetry,
  hooks, ambient credential discovery, or host-setting mutation.

### Providers, routing, and worker tools

- Replace exact-version provider admission with semantic capability probing.
  Use Codex app-server account/model/session capabilities and the installed
  Claude CLI's authenticated status without reading either host's credentials.
- Report Codex's advertised model catalog and Claude's verified, non-exhaustive
  candidates without claiming unsupported subscription enumeration.
- Support configuration-defined OpenAI Responses-compatible and Anthropic
  Messages-compatible gateways. Credentials are environment or macOS Keychain
  references only; values are never serialized.
- Normalize authentication, models, effort levels, context/tool/edit support,
  cost/latency hints, quota state, and resume support. Route by hard capability
  gates followed by a deterministic balanced score, with explainable rejection,
  selection, fallback, and escalation records.
- Give gateway models parity through a Level 7 tool broker limited to scoped
  read/search, memory query, patching, Git status/diff, configured exact-argv
  commands, verification, and completion. Reject shell composition, path or
  symlink escape, ambient network tools, unbounded output, and self-audit.

### Sync and Onboard

- Build a content-addressed repository graph for files, symbols, modules,
  imports, calls, tests, commits, decisions, briefs, runs, verification, and
  findings under `.git/l7/memory`.
- Parse Go with the standard library and JavaScript/TypeScript/Python with
  pinned Tree-sitter grammars; use a generic file/Git graph elsewhere.
- Combine structural, lexical, recency, and local Apple `NLEmbedding` signals.
  Store the embedding revision/dimensions, rebuild derived indexes on mismatch,
  and keep structural retrieval available with a truthful degraded warning.
- Exclude ignored/generated/binary content, environment files, credentials,
  likely secrets, and transcripts before indexing or embedding.
- Make `l7 onboard` inspect state before effects, explain the next transitions,
  and create orchestration configuration only after explicit application.

### Cyber

- Run read-only attack-surface, trust-boundary, data-flow, dependency, secret,
  SAST, IaC/container, authentication, authorization, and exploit-hypothesis
  analysis first.
- Require explicit active mode and a disposable Docker/OrbStack-compatible
  container or configured VM for exploit confirmation. Use a pinned signed OCI
  image, a disposable repository copy, non-root execution, resource limits, no
  host credentials/sockets, and engine-created internal networks only.
- Keep complete evidence in `.git/l7/security`; export redacted Markdown/JSON
  only on request. Never target the Internet or another machine in v1.
- Make remediation a separate Tier 3 change brief. A Cyber run cannot patch its
  own finding or represent its own remediation review as independent.

### Headless

- Use native provider sessions and provider-neutral durable handoffs, never
  clipboard, keystroke, or `/new` UI automation.
- Accept a finalized `concept.md` or feature-wave file with measurable
  acceptance criteria. Produce a canonical manifest bound to the objective,
  base commit, allowed paths/commands, Tier 2 ceiling, tests, provider policy,
  target branch, local-merge policy, network policy, and stop-before-deploy
  boundary.
- Require active owner confirmation of the manifest digest once. Execute only
  in-scope Tier 1/2 waves through disposable worktree, implementation,
  verification, independent review, and exact-head CAS local merge.
- Persist event and checkpoint state around transitions. Resume across process
  restarts; wait for natural provider quota reset without an arbitrary total
  execution cap; never buy or redeem capacity. Pause after three identical
  no-progress failures at one checkpoint.
- Pause on scope expansion, protected paths, Tier 3 work, secrets, destructive
  actions, target-branch divergence, invalid authority, remote push, release,
  publication, or deployment.

## Explicit boundaries

- No Windows or Linux product support in v1.
- No hosted Level 7 account, credential broker, telemetry, background daemon,
  remote memory, remote embeddings, or automatic secret ingestion.
- No exhaustive Claude subscription model claim and no bundled Claude Agent SDK.
- No unrestricted gateway shell, browser, network, or host filesystem access.
- No active Cyber execution without a disposable container/VM; no Internet or
  off-machine penetration target.
- No Headless remote push, protected-branch merge, release, publication, or
  deployment.
- No silent `.l7/config.json` migration and no rewrite of historical governance
  artifacts or deprecated evidence chains.
- The existing untracked
  `docs/artifacts/foundation-rebaseline-admission-audit.md` in the source
  checkout remains untouched and outside this worktree.

## Exact implementation file set

Implementation commit `dc1fdb6a88d173a78e375bb8d54691c999d12e39`
changes exactly these 56 paths relative to approved brief commit
`a9038285d612dec6d1c496c3f9a69fed9ca75f74`. This superseding brief remains the
single governance path and changes no implementation byte.

Add:

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

Delete: none.

## Acceptance criteria

1. Both plugin packages install one canonical v1 engine and the four new skills;
   CLI and MCP expose versioned human/JSON contracts and default all new effects
   OFF.
2. Strict legacy configuration remains readable and unchanged. Orchestration
   configuration rejects unknown, duplicate, malformed, secret-bearing, or
   unsafe values and can be removed to roll back to v0.1.1 behavior.
3. Codex and Claude authenticated-host probes tolerate compatible client version
   drift, never read credential files, and report only observed capabilities.
   Missing auth, unsupported models/efforts, malformed streams, and quota limits
   remain truthful non-success states.
4. Gateway protocol fixtures prove streaming/tool loops, model rejection,
   authentication failure, timeout/cancellation, output bounds, redaction, and
   scoped mutation in disposable worktrees for both supported wire protocols.
5. Routing is deterministic for the same snapshot and task, applies C1-C4 effort
   mapping and risk floors, records every candidate decision, escalates boundedly,
   and cannot accept an implementer's own audit.
6. Sync golden tests cover Go, JavaScript, TypeScript, Python, generic fallback,
   incremental invalidation, Apple embedding revisions, hybrid retrieval,
   excluded secrets, concurrent runs, and corrupt-index recovery.
7. Cyber read-only analysis works without a container. Active mode fails closed
   without isolation and cannot access host credentials, the original checkout,
   the Internet, or off-machine targets when isolation is available.
8. Headless fixtures prove crash recovery at every transition, provider failover,
   natural quota-reset waiting, no-progress pause, scope/protected-path blocking,
   branch-divergence blocking, multi-wave verification/review, and exact-head
   local merge while stopping before external effects.
9. Repository unit, fuzz, race, adversarial, policy, distribution, compatibility,
   reproducibility, SBOM, and benchmark checks pass. macOS 13+ `arm64` and
   `amd64` packages pass install, CLI, MCP, and upgrade/rollback conformance.
10. Stable release is withheld until protected release identities, provenance,
    signed attestations, an exact-candidate verification record, an independent
    read-only audit, and a named owner GO bind the final head. Passing tests or
    repository text cannot substitute for those authorities.

## Sequencing

1. Reconcile the accepted lifecycle/durability foundation from the existing
   remediation work without merging its branch or copying stale governance.
2. Land strict orchestration contracts, durable state, CLI/MCP shell, and
   architecture-specific packaging behind default-OFF configuration.
3. Land native provider discovery, normalized capabilities, routing, and route
   evidence; then add gateway protocol workers and the bounded tool broker.
4. Land Sync and Onboard, followed by Cyber and then Headless. Each slice keeps
   the complete repository green and names its next executable transition.
5. Run full conformance and performance qualification, create exact-candidate
   verification, obtain independent read-only audit and owner GO, then publish
   `v1.0.0` as a separate authorized release effect.

## Risks and mitigations

- **Credential or source disclosure:** credential references only, centralized
  redaction, pre-index secret exclusion, adversarial fixtures, and no raw prompt
  or transcript persistence.
- **Provider drift or false capability:** semantic probes, versioned snapshots,
  explicit degradation, contract fixtures, and no inference from help text alone.
- **Agent tool escape:** canonical paths, symlink rejection, exact argv, bounded
  worktrees/processes, deny-by-default network, and adversarial containment tests.
- **Unsafe autonomous progress:** digest-bound manifest approval, Tier 2 ceiling,
  CAS transitions, durable checkpoints, independent review, and hard pauses at
  scope, authority, protected, destructive, and external-effect boundaries.
- **Cyber harm:** read-only first, signed disposable lab image, no host secrets or
  off-machine targets, and separate remediation authority.
- **Memory poisoning or staleness:** source digests, typed provenance, atomic index
  replacement, secret exclusion, model-revision invalidation, and rebuildable
  derived data.
- **Package and dependency expansion:** pinned sources/checksums/licenses,
  reproducible per-architecture builds, SBOMs, vulnerability review, and rollback
  to the retained v0.1.1 instruction-only release.

## Rollback

Before approval, revert only this superseding brief commit to restore the exact
`dc1fdb6a88d173a78e375bb8d54691c999d12e39` tree and its historical brief
bytes. Before release, revert the brief correction, implementation, and original
brief commits in reverse order, or delete the isolated worktree; the v0.1.1
packages and original checkout remain unchanged. After installation,
disable/remove `.l7/orchestration.json`, uninstall v1, reinstall v0.1.1, and
remove only package-manager-owned v1 binaries plus derived `.git/l7` v1 state
after explicit confirmation. No credential, remote, production, or deployment
cleanup is required because v1 does not own those effects.

## Current transition

1. Commit this governance-only revision as the direct child of
   `dc1fdb6a88d173a78e375bb8d54691c999d12e39`.
2. Stop for explicit approval from Anup Pandey bound to that exact revision
   commit. The approval bound to `a9038285d612dec6d1c496c3f9a69fed9ca75f74`
   does not transfer.
3. After approval, run `make verify` against the exact revision head and, only
   on `PASS`, create its single candidate-bound verification record.
4. Independent audit, push, signing, release, publication, merge, and deployment
   remain separate, unauthorized transitions.
