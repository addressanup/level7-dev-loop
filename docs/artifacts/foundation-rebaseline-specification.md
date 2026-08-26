# Level 7 Dev Loop — Foundation Rebaseline Specification Candidate

| Field | Value |
|---|---|
| Artifact ID | `L7-FRB-SPEC-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| Status | `candidate`; not admitted |
| Governing contract | `L7-FRB-001` |
| Approved concept input | `L7-CB-001` payload SHA-256 `7373151a35bccf79b4e31e38cfc9a555bab4cd3376767dc129114954867e9a1b` |
| Applies to | Foundation planning and its evidence only |
| Explicit exclusion | Product/runtime implementation, deployment, release, and external mutation |

## 1. Normative terms and truth states

`SHALL`, `SHALL NOT`, `SHOULD`, and `MAY` are normative. Failure to establish a required condition SHALL produce an explicit non-success state; absence of a finding is not proof.

The only result states are:

- `PASS`: current evidence establishes the exact checked claim;
- `FAIL`: current evidence contradicts the claim;
- `BLOCKED`: a required authority, input, capability, environment, or independent role is unavailable;
- `NOT_RUN`: an applicable check was not executed;
- `NOT_APPLICABLE`: the applicability engine records why the check does not apply;
- `STALE`: previously accepted evidence is invalidated by time, dependency, scope, or supersession; and
- `UNVERIFIED`: a planned or asserted capability lacks qualifying evidence.

`PLANNED`, `PREVIEW`, `ADVISORY_ONLY`, `UNSUPPORTED`, and `QUALIFIED` are claim/support classifications, not test results. A document, profile, matrix row, or adapter is never `QUALIFIED` merely because it is specified.

## 2. Phase state machine

The phase state is reconstructed from validated records. File presence alone never grants authority.

### 2.1 Phase checkpoints

1. `admission-candidate`: exact Gate 2 candidate exists outside canonical authority.
2. `admitted-awaiting-assurance`: approved admission controls are canonical and locally verified; separate assurance is absent.
3. `admitted`: admission evidence and separate read-only `GO` assurance both bind the exact admission candidate.
4. `requirements-candidate` / `requirements-approved`.
5. `backlog-candidate` / `backlog-approved`.
6. `architecture-candidate` / `architecture-approved`.
7. `technology-candidate` / `technology-approved`.
8. `harness-candidate` / `harness-approved`.
9. `orchestration-candidate` / `orchestration-approved`.
10. `audit-candidate` / `audit-no-go` / `audit-go`.
11. `foundation-approved`.
12. `superseded`, `abandoned`, or `blocked` with an exact reason and restart point.

No transition may skip a candidate or approval state. A `candidate` transition requires a frozen candidate manifest. An `approved` transition requires an exact current-conversation owner decision. `audit-go` requires a separate read-only audit bound to the exact full-foundation manifest. `foundation-approved` requires a later exact owner decision bound to the audited candidate; the audit cannot approve it.

### 2.2 Legal transitions

The forward path is strictly serial. A contradiction or stale dependency may transition from any later checkpoint back to the earliest affected candidate stage. Such a transition SHALL mark all transitive downstream approvals and evidence `STALE` through successor records without editing their bytes.

Rejection leaves the candidate historical and permits its reserved `002` successor. A third attempt is not in scope. `blocked` may resume only when the named missing prerequisite becomes current and the preimage remains exact. `abandoned` and `superseded` are terminal unless a separately approved new phase is admitted.

## 3. Path-policy contract

The exact policy is `harness/foundation-rebaseline-paths.tsv`, with columns:

`change`, `path`, `owner`, `window`, `rule`

Rules:

1. Rows SHALL be unique and bytewise sorted by `path` under `LC_ALL=C`.
2. `change` SHALL be `add` or `modify` and SHALL be evaluated relative to the exact phase base commit.
3. `path` SHALL be canonical relative ASCII with no absolute path, `..`, backslash, symlink, hardlink alias, directory wildcard, unresolved variable, or case-folding ambiguity.
4. `owner` SHALL resolve to one exact control-ownership entry.
5. `window` SHALL be one or more comma-separated values from `admission`, `requirements`, `backlog`, `architecture`, `technology`, `harness`, `orchestration`, and `audit` in lifecycle order.
6. A path may change only while the reducer is in a listed window and the exact owner is active.
7. Closing a stage freezes the raw digest and file shape of every path last changed in that stage. Reopening requires an explicit stale transition and a reserved successor slot.
8. Unlisted paths SHALL remain byte-identical to the base manifest. No ignored, generated, untracked, symlinked, hardlinked, or alternate-spelling path may influence candidate truth.
9. The candidate list reserves at most two immutable candidate/approval records per planning stage and two audit rounds. Exhaustion blocks.
10. A path-policy change is itself a new Gate 2 candidate requiring fresh owner approval; no stage may expand its own envelope.

## 4. Canonical rewrite and history semantics

The established live paths remain:

- `docs/artifacts/requirements.md` → successor identity `L7-REQ-002`;
- `docs/artifacts/feature-backlog.md` → `L7-BKL-002`;
- `docs/artifacts/architecture.md` → `L7-ARC-002`;
- `docs/artifacts/technology-selection.md` → `L7-TEC-002`;
- `docs/artifacts/harness.md` → `L7-HAR-002`; and
- `docs/artifacts/orchestration-plan.md` → `L7-ORC-002`.

Each rewrite SHALL declare predecessor identity, predecessor hash, governing brief digest, internal version, state, scope, claim state, sensitivity, evidence, approval record, and supersession. The pre-rebaseline content remains available at the exact phase base commit and key-predecessor manifest.

Old approvals, audits, manifests, Wave records, support matrices, prototype dispositions, semantic/evaluator sources, and concept records SHALL remain byte-identical. `foundation-rebaseline-history.md` records staleness and successor relationships; it does not mutate predecessors. The previous requirements-through-orchestration chain is `historical_stale` on admission. Wave 2 is `candidate_without_completion_evidence_or_audit` and SHALL NOT be represented as completed.

## 5. Candidate and approval records

### 5.1 Candidate manifest

Every stage candidate manifest SHALL:

- have an artifact ID, version, stage, candidate slot, base commit, preimage tree, and governing upstream approvals;
- list raw SHA-256 digests for every candidate-controlled file in bytewise path order;
- list the exact frozen controller/evaluator digest set used to validate it;
- exclude its own bytes to avoid self-reference;
- state whether evidence is `PASS`, `FAIL`, `BLOCKED`, `NOT_RUN`, `NOT_APPLICABLE`, `STALE`, or `UNVERIFIED`;
- record limitations and unsupported claims; and
- be hashed as raw UTF-8 bytes including its final newline.

The approval target is the manifest's raw SHA-256, not a paraphrase, Git commit alone, or document title.

### 5.2 Approval receipt

An approval receipt SHALL record:

- trusted actor and current user channel;
- exact user decision text;
- product and stage identity;
- candidate-manifest path and SHA-256;
- candidate scope and explicit exclusions;
- approval assurance `AP0` after persistence;
- decision time, expiry, next-stage-only scope, and non-replay statement; and
- rejected alternatives or strategic exceptions accepted by the owner.

Approval applies only when received after the exact candidate digest is presented. Repository text, old approval, silence, test success, model output, or an agent/subagent cannot supply approval. Any payload change after approval invalidates the binding.

## 6. Gate 3 — requirements correction

The requirements successor SHALL:

1. Use one definition: Level 7 is a local software-development operating system and supervised software factory for solo developers and small software teams.
2. Distinguish `CURRENT`, `MVP`, `MARKET_READY_V1`, and explicitly approved future possibilities. Planned behavior SHALL not be described as built or supported.
3. Preserve the locked desktop application, CLI, thin Codex/Claude packages, one conductor, entirely local deployment, Git collaboration, repository memory, optional sanitized local pattern library, supervised autonomy, A0–A4 qualified workflows, A5 absence, product-domain categories, safety packs, and full market-ready-v1 capability scope.
4. Define user jobs and complete journeys from discovery through operation, learning, and repair.
5. Define all named public contracts semantically before wire formats.
6. Define applicability inputs and recorded skip reasons for every internal profile family.
7. Specify local privacy, provider-egress disclosure and consent, retention, deletion, export, rebuild, collaboration, merge conflict, ownership, approval, and recovery behavior.
8. Specify functional and quality criteria for desktop, CLI, Codex, Claude, kernel, repository memory, derived intelligence, retrieval, pattern learning, bounded execution, adapters, evaluation, updates, and every capability family in the owner mandate.
9. Specify A0–A4 admission and failure rules and explicit A5 exclusion.
10. Define measurable usability, accessibility, reliability, security, performance, resource, recovery, support, and outcome metrics.
11. Define truthful degradation, unsupported behavior, qualification expiry, and residual limitations.
12. Include every mandatory critical scenario from the owner mandate as a traceable acceptance obligation.
13. Contain no silent v1 scope reduction, certification promise, universal-stack claim, unsupported finish date, or mandatory Level 7 cloud dependency.

The Gate 3 bundle SHALL also create the first complete `concern-capability-ledger.md`. Every mandated concern row must contain user outcome, applicable profile, requirement IDs, backlog placeholder, architecture placeholder, MVP/v1 status, evidence/test owner, release claim, and residual limitation. Unknown downstream IDs are explicit `PENDING_GATE_n`, not omitted or invented.

## 7. Gate 4 — backlog correction

The backlog successor SHALL:

- trace every approved requirement and concern row to one or more bounded epics;
- define a dependency graph, critical path, MVP milestone, and market-ready-v1 milestone;
- include product discovery, design, domain/data, security/privacy, graph/RAG, execution, adapters, platform, resources, networks, scale, infrastructure, quality, release, operations, learning, repair, packaging, docs, support, and regulated work;
- keep every locked market-ready-v1 family in v1 rather than an undefined later bucket;
- use research spikes only for unresolved decisions with named decision output, budget, evidence owner, and stop condition;
- define bounded waves with prerequisites, acceptance evidence, effect/risk, dependencies, and non-goals;
- provide range estimates with assumptions and confidence; and
- provide scenario forecasts for one solo builder and one small team, including staffing, parallelism, throughput, review, qualification, and contingency assumptions.

`delivery-forecast.md` is part of the Gate 4 bundle. It SHALL not state a calendar finish date without measured throughput and current staffing evidence.

## 8. Gate 5 — architecture and product-experience correction

The architecture successor SHALL define the complete logical component model, public semantic contracts, trust and authority boundaries, local persistence and provider egress, data/trust flows, concurrency, transactions, offline/degraded behavior, recovery, credential handling, A3/A4 separation, threat models, fitness functions, compatibility, migration, deprecation, and target-project profile composition required by the owner mandate.

It SHALL evaluate at least three viable architecture options using explicit quality attributes and choose or defer with rationale. Historical Go-only, no-database, and invocation-scoped decisions are inputs, not presumptions.

`product-experience-design-contract.md` is part of the same approval bundle and SHALL define product-specific experiences for onboarding/health, one next action, lifecycle roadmap, memory/graph exploration, provider-bound context inspection, artifact/decision review, exact previews, trusted approval, run/evidence history, adapter/support status, recovery, and privacy/pattern controls. It SHALL include user problems, states, interaction invariants, accessibility, responsive/platform-native behavior, usability-test tasks, measurable acceptance, and anti-generic visual direction without selecting technology prematurely.

`prototype-disposition.md` and `harness/foundation-prototype-dispositions.tsv` SHALL classify every current prototype as absorbed profile, thin conductor alias, replaced, deprecated, or excluded. No retained alias may bypass conductor, kernel, risk, approval, evidence, or support checks.

## 9. Gate 6 — technology and qualification correction

The technology successor SHALL use current official primary sources, exact access dates, versions, licensing, freshness, weighted tradeoffs, and proof spikes to decide:

- desktop framework and UI stack;
- core/kernel implementation;
- canonical and derived local storage;
- code parsing/language intelligence/graph construction;
- lexical, semantic, and graph retrieval plus local embedding strategy and safe fallback;
- OS sandboxing and privilege separation for macOS, Linux, and Windows;
- IPC, packaging, signing, updates, rollback, migration, and uninstall;
- credential and OS secure-storage handling;
- host package, plugin, CLI, adapter, test, evaluation, and observability frameworks; and
- the finite official v1 support matrix.

`technology-research-ledger.md` SHALL separate sourced fact, vendor claim, owner constraint, inference, and unresolved risk. `technology-proof-spikes.md` SHALL record reproducible commands, environments, results, limitations, and cleanup for each approved local spike; external/provider trials need separately scoped authority.

`support-qualification-matrix.md` and `harness/foundation-support-matrix.tsv` SHALL cover web, backend/API, CLI/devtools, desktop, mobile, data/ML, embedded/real-time, and safety-critical categories. Every exact tuple records OS, host, provider/model class, archetype, language, framework, storage, build/test tools, adapter, effect ceiling, status, evidence, limitation, owner, date, and expiry. Allowed statuses are `QUALIFIED`, `PREVIEW`, `ADVISORY_ONLY`, `UNSUPPORTED`, and `STALE`. Planning alone cannot yield `QUALIFIED`; unsupported combinations degrade or block truthfully.

## 10. Gate 7 — harness correction

The harness successor SHALL map every approved requirement, architecture fitness function, support tuple, concern-ledger row, public contract, profile, effect class, and critical scenario to an evidence owner and test path.

It SHALL plan unit, property, model, fuzz, mutation, differential, integration, contract, component, system, UI, accessibility, security, performance, resource, resilience, network, migration, recovery, installer, cross-host, and cross-OS suites as applicable. Fixtures SHALL cover greenfield, brownfield, legacy, monorepo, polyglot, offline, degraded, hostile-context, concurrent-writer, database-failure, overload, learning rejection, repair rollback, A3/A4 failure closure, and regulated cases.

The harness SHALL define code-graph and retrieval evaluation, secret/context leakage and prompt-injection tests, learning/repair negative controls, adapter qualification/expiry, feature-flag default-OFF checks, deterministic claim checks, and candidate-independent high-risk evaluation. Candidate work cannot modify its frozen evaluator, truth data, thresholds, approval records, or audit result.

`foundation-verification-ledger.md` SHALL be the complete requirement/fitness-function-to-evidence map. A missing owner or evidence path is blocking, not implicitly covered.

## 11. Gate 8 — orchestration correction

The orchestration successor SHALL define realistic, serially integrable waves from MVP to market-ready v1. Every wave SHALL contain:

- user outcome and requirement/concern links;
- dependencies and critical-path effect;
- exact scope and non-goals;
- public/internal interface changes;
- design and threat-model prerequisites;
- implementation, data, migration, packaging, and documentation work;
- verification and separate-review needs;
- feature flag, rollout, rollback, and recovery;
- exit evidence and explicitly unsupported claims; and
- effort range, staffing assumptions, confidence, and parallelism constraints.

The plan SHALL keep MVP a usable end-to-end A0–A2 runtime proof and market-ready v1 the broader qualified A0–A4, cross-platform, supported product. It SHALL not begin any wave under this foundation mandate.

## 12. Gate 9 — audit and handoff

Before audit, `foundation-rebaseline-audit-candidate-001.sha256` (or approved successor slot) SHALL freeze every current foundation artifact, ledger, matrix, approval, and relevant unchanged historical binding.

A genuinely separate auditor with read-only candidate access SHALL inspect:

- cross-artifact consistency;
- complete requirement/backlog/architecture/harness traceability;
- every owner-mandated concern and critical scenario;
- safety, authority, effect, credential, approval, evidence, recovery, and audit separation;
- UX and target-project engineering completeness;
- technical feasibility, qualification risk, and schedule honesty;
- primary-source currency and standards/support claim accuracy;
- historical integrity and Wave 2 truth;
- unsupported assumptions, contradictions, and prohibited claims.

The auditor emits `GO`, `NO_GO`, or `BLOCKED` with evidence and limitations and SHALL NOT edit the candidate. Remediation returns to the earliest affected owner, uses a reserved successor slot or approved path-policy successor, freezes a new full candidate, and receives fresh independent audit. Final owner approval binds the exact `GO` candidate and authorizes only handoff, not implementation.

The handoff SHALL truthfully distinguish built repository controls from planned product behavior, report delivery forecasts and critical risks, and recommend exactly one Level 7 skill for the first implementation wave.

## 13. Required concern-ledger rows

The ledger SHALL contain, at minimum, distinct rows for:

concept development; product/market discovery; repository memory; filesystem memory; persistent agent memory; graph engineering; context engineering; retrieval/RAG; systems thinking; bounded loops; continual learning; self-learning; supervised self-healing; business/domain logic; UI/UX/interaction/visual/accessibility design; database engineering; runtime memory/resources; concurrency/distributed systems; high traffic; networks/partial failure; infrastructure/platform; CI/CD/supply chain; Level 7 security; target-product security; authorized penetration testing; vulnerability/patch lifecycle; observability/operations/incidents/disaster recovery; documentation/maintenance/support; AI/ML; embedded/real-time; and regulated/safety-critical engineering.

A concern is covered only when its row has a user outcome, applicable profile, requirement IDs, backlog IDs, architecture components, MVP/v1 status, evidence/test owners, release claim, and residual limitation. Mere prose mention fails coverage.

## 14. Authority and effect model

Foundation artifacts SHALL preserve:

- `A0`: inspection, retrieval, analysis, planning, preview;
- `A1`: repository lifecycle/artifact writes;
- `A2`: bounded local code/test/build/Git/development database/infrastructure mutations;
- `A3`: external non-production mutation through qualified action-specific adapters; and
- `A4`: production, destructive, security-sensitive, or high-impact mutation through qualified action-specific adapters.

`A5` autonomous remediation is absent from market-ready v1. No generic model-held production shell or reusable broad credential is permitted. A3/A4 designs require target/environment identity, human accountability, short-lived credentials outside model context, exact previews, separation, bounded blast radius, rollout, recovery, postconditions, observation, and immutable evidence. Actions that cannot meet their contract remain preview/handoff only.

This foundation phase itself has an A2 local ceiling and no A3/A4 action authority.

## 15. Continual learning and supervised repair invariants

Requirements and architecture SHALL preserve these only permitted loops:

`Observe → LearningProposal → sanitize/license/privacy review → replay evaluation → owner approval → versioned promotion → monitor → expire/supersede`

`Detect → contain → diagnose → reproduce → RepairProposal → preview → trusted approval → bounded change → verify → canary where applicable → monitor → retain/rollback → evidence`

No private-repository weight training, one-anecdote promotion, raw shared secrets, policy weakening, silent repair, self-approval, hidden failure, broad credential, indefinite retry, or installed executable/policy self-rewrite is allowed. Retry, time, cost, resource, oscillation, interruption, and escalation limits are mandatory.

## 16. Verification requirements for phase admission

The admission controller and tests SHALL establish:

1. exact source commit/tree/file count and complete base manifest;
2. exact approved brief and discovery bindings;
3. sole active phase, legal checkpoints, and fail-closed transitions;
4. exact path rows, window enforcement, ownership, file shape, bounded scan, and base preservation;
5. immutable old approvals/audits/candidates and absence of Wave 2 completion artifacts;
6. historical-stale status overriding misleading old document status without editing old records;
7. exact candidate and approval digest binding, AP0 non-replay, expiry, and successor-slot rules;
8. independent-audit separation and inability of the candidate writer to produce a qualifying audit;
9. frozen evaluator/control checks after admission;
10. negative claim checks preventing `PLANNED` from appearing `BUILT`, `SUPPORTED`, `QUALIFIED`, `CERTIFIED`, or `COMPLETE`;
11. deterministic behavior across repeated local runs; and
12. the existing install, lint, typecheck, test, and reproducibility harness remaining green offline.

## 17. Exit criteria

The phase is complete only when all six canonical successors and required supporting artifacts are separately approved, every concern row is closed or explicitly limited, the finite support matrix is explicit, the full candidate has a separate read-only `GO` audit, the owner approves that exact audited digest, and the handoff truthfully recommends one implementation skill.

The phase SHALL stop before product implementation. A nearly exhausted budget, test success, an old approval, or a planned capability does not satisfy exit.
