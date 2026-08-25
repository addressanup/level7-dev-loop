# Level 7 Dev Loop — Product Requirements

| Field | Value |
|---|---|
| Artifact ID | `L7-REQ-001` |
| Artifact type | Product requirements |
| Artifact schema | Bootstrap/pre-schema; migration required when the canonical schema exists |
| Foundation step | 1 — Requirements |
| Status | Approved for foundation step 2 |
| Version | 0.2.0 |
| Date | 2026-08-24 |
| Product | Level 7 Dev Loop |
| Initial hosts | Codex CLI and Claude Code |
| Scope | Local repository/worktree plugin, v1 |
| Scope identity | Non-Git workspace snapshot observed on 2026-08-24; no commit identity available |
| Effect and risk | A1 artifact-only write; foundation-planning risk |
| Requirements approver | Product owner |
| Approval state | Owner approved in current conversation on 2026-08-24 (`AP1` at confirmation time); persisted record is `AP0` until revalidated |
| Sensitivity | Internal product planning |
| Review condition | Owner review; re-review after material scope, host, or autonomy-boundary change |
| Supersession | No prior requirements artifact; prototype dispositions in §2.2 become the target contract only after approval |

**Approval decision:** the product owner accepted this requirements boundary and authorized foundation step 2—feature-backlog drafting—only. Approval does not authorize architecture selection, harness work, source/configuration changes, external actions, deployment, exposure, or release.

## 1. Executive definition

Level 7 Dev Loop is a **repository-oriented, human-governed software and product evolution workflow, evidence, and decision-support plugin** for Codex CLI and Claude Code.

It turns ambiguous engineering intent into the next safe, proportionate, evidence-producing action; preserves durable lifecycle state in the repository; and enables a responsible human to approve, reject, resume, release, audit, or retire work without relying on chat history or confidence alone.

Level 7 is not primarily a refactoring tool. Refactoring is one behavior-preserving change class inside the broader product domain of software evolution and assurance. Level 7 may coordinate implementation and verification, but the accountable human remains the decision owner.

### Product promise

> Determine the best evidence-supported state of the scoped system with disclosed uncertainty, recommend exactly one safe next transition, apply controls proportional to risk, and leave portable evidence for the next responsible decision.

### Primary outcome

Improve movement from ambiguous intent to an evidence-supported outcome while reducing unmanaged risk, rework, and dependence on conversational memory. Any speed or quality improvement remains a hypothesis until measured against a baseline that includes verification and correction effort.

### Approved primary-user assumption

The primary v1 user and accountable approver is an **accountable technical owner**: a solo maintainer, technical lead, or staff/principal engineer with repository responsibility and authority to decide what may change locally. The core operator is an implementing engineer or maintainer whose approval rights depend on separately granted authority. These are roles rather than company-size segments, so the product remains useful to both individuals and teams.

### What using v1 feels like

Ask for engineering work in natural language. Level 7 inspects only the authorized scope, explains the evidence-supported state and uncertainty, and proposes one bounded next action. After approval, it returns a local artifact or diff with applicable verification and the next responsible decision—without requiring the user to memorize the skill graph.

## 2. Current evidence baseline

The existing package is treated as a prototype input, not the v1 contract. The observations below were produced from repository-root-bounded read-only inspection on 2026-08-24.

| Observed fact | Evidence | Requirement consequence |
|---|---|---|
| The prototype contains 12 Markdown workflow skills, all declaring `user-invocable: true`. | [`skills/`](../../skills/) and `rg --files skills -g SKILL.md` | Preserve useful intent, but specify and test host-valid skill contracts. |
| It contains three plugin manifests plus a fourth marketplace metadata file; duplicated version/description fields can drift. | [`plugin.json`](../../plugin.json), [Codex manifest](../../.codex-plugin/plugin.json), [Claude manifest](../../.claude-plugin/plugin.json), [`marketplace.json`](../../marketplace.json) | Deterministically produce host distributions from one semantic source; do not infer portability from similar files. |
| All four JSON files parse, but no manifest-schema or host-install conformance test exists. | `jq empty` passed on 2026-08-24. | JSON syntax is evidence only of syntax, not compatibility. |
| The README claims dual compatibility at line 3 and rule enforcement at line 69. | [`README.md`](../../README.md) | Treat both claims as unverified until host conformance and enforcement-locus tests pass. |
| `docs/artifacts/` had no lifecycle evidence before this document. | [`docs/artifacts/.gitkeep`](.gitkeep) | Establish schema-controlled repository state and artifact validity. |
| The directory is not currently a Git repository. | In `/Users/anuppandey/Desktop/level7-dev-loop`, `git rev-parse --is-inside-work-tree` exited 128 on 2026-08-24. | Git-dependent behavior must have an explicit unavailable/degraded state and non-Git revision identity. |
| No test suite, eval fixtures, CI, build metadata, schema, lockfile, executable validator, `CLAUDE.md`, or `LICENSE` is present. | `find . -maxdepth 3 -type f -print` inventory on 2026-08-24. | Compatibility, routing, safety, legal, and lifecycle claims are not release-ready. |
| Manifests declare version `1.0.0` and MIT, but no license file or release evidence exists. | Current manifests and workspace inventory. | Versioning and legal packaging must be corrected before public release. |
| Claude installation currently requires manual copying/merging, while v1 targets validated, non-destructive installation. | [`README.md`](../../README.md), lines 14–33 | Installation must become a tested product journey rather than file-layout guidance. |
| Routing and workflow rules contain contradictory chained transitions, broad triggers, inconsistent artifact outputs, and universal recipes. | [`AGENTS.md`](../../AGENTS.md), [`WORKFLOW.md`](../../references/WORKFLOW.md), and current [`skills/`](../../skills/) | The approved dispositions below define the replacement target; active files remain unchanged until an approved implementation step. |

### 2.1 Baseline verification record

| Check | Result on 2026-08-24 |
|---|---|
| Working directory | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Git worktree | No; command exited 128 |
| Artifact files before this draft | `.gitkeep` only |
| Repository files | 20 files before this draft: 12 skills, 4 JSON metadata files, `README.md`, `AGENTS.md`, `references/WORKFLOW.md`, and `.gitkeep` |
| Tests/CI/build/schema/lockfiles | None detected within three directory levels |
| JSON syntax | All four JSON files passed `jq empty` |
| External compatibility | Not executed; remains unverified |

This bootstrap record is dated evidence rather than an immutable snapshot. The later harness SHALL produce machine-readable inventory evidence with digests and environment metadata.

### 2.2 Prototype disposition

Approval establishes the following **target requirements contract**. It does not silently edit or deactivate current `AGENTS.md`, skills, manifests, or documentation; those remain operational until an approved implementation reconciles and tests them.

| Prototype source/rule | Disposition | Replacement requirement |
|---|---|---|
| `AGENTS.md`: spec, evidence, small diffs, green harness | Retain and make risk-proportionate/testable. | §§5, 9, 10 |
| `AGENTS.md` and conductor: exactly one next skill, but chained/alternative outputs exist | Replace chained outputs with one visible transition, clarification, or blocked state. | `L7-ROUTE-001`, `L7-FLOW-009` |
| `AGENTS.md`: every user-visible production behavior requires a default-OFF flag | Replace with an explicit reversal/release contract; require a lifecycle-managed flag when applicable. | `L7-REL-001`, `L7-REL-002` |
| README: dual-compatible and “enforces” | Withhold until compatibility and enforcement-locus evidence pass. | `L7-HOST-003`, `L7-AUTH-008` |
| Root/Codex/Claude manifests and marketplace metadata | Replace hand-maintained duplication with deterministic host packaging and independent promotion. | `L7-HOST-001`–`L7-HOST-011` |
| `l7-constitution`: approximately 80 intent-changing lines | Replace numeric target with cohesive, bounded, recoverable slice criteria. | Principle 8; `L7-NFR-028` |
| `l7-greenfield`: exactly three architecture options and one factual question per round | Treat option count as justified by credible alternatives; batch independent facts while isolating branching decisions. | `L7-FLOW-010`, `L7-NFR-019` |
| `l7-build`: schema → logic → API → UI → tests | Replace with change-appropriate sequencing and continuous proof. | `L7-FLOW-010`, §9.6 |
| `l7-change`: universal RICE, fixed rollout percentages, flag-off as hard rollback | Make prioritization and rollout context-dependent; distinguish flag disablement from data/external rollback. | `L7-FLOW-004`, `L7-REL-001`–`L7-REL-003` |
| `l7-release`: persona-based independence | Replace with procedural independence bound to an exact source identity and artifact digest. | `L7-AUDIT-001`, `L7-AUDIT-002` |
| `l7-deploy`: plugin-executed production deployment | Retain planning/evidence handoff; defer Level 7 execution of external/production mutation beyond v1.0. | `L7-AUTH-004`, `L7-HOST-010` |
| `l7-ops`: incident trigger without incident lifecycle | Expand to an applicable incident proof profile and reviewed follow-up proposals. | `L7-PROOF-010`, `L7-OPS-004` |
| `l7-review`: `FULLY COMPLIANT` | Replace with named-baseline evidence alignment and scoped verdicts. | `L7-COMP-001` |
| `l7-experience` and `l7-geometry`: universal geometry and `PERFECT` | Retain evidence-led UX; replace universal geometry and absolute verdicts with product-specific criteria. | `L7-PROOF-007`, `L7-NFR-021` |
| `l7-storybook`: mandatory multi-tenant narrative | Retain as an applicable tenancy/collaboration profile rather than a universal project story. | `L7-FLOW-004`, `L7-PROOF-008` |
| `references/WORKFLOW.md`: numbered fixed sequence | Replace with a resumable state model and profile-selected transitions. | `L7-FLOW-001`–`L7-FLOW-010` |

## 3. Problem statement

AI coding agents can generate changes quickly, but projects lack a dependable layer that answers:

- What state is this scoped system actually in?
- What user or operational outcome are we trying to change?
- Which engineering practices apply to this project stage and change class?
- What may the agent inspect, propose, modify, execute, or release?
- Which claims are `OBSERVED`, `REPRODUCED`, `INFERRED`, `USER_ASSERTED`, `UNVERIFIED`, or `NOT_APPLICABLE`?
- What evidence is sufficient for this risk and blast radius?
- How can work resume safely after compaction, interruption, or host switching?
- How do learning and production feedback improve future work without allowing unsafe self-promotion?

Without this layer, teams experience prompt sprawl, context loss, inconsistent workflows, process theater, oversized changes, false assurance, repeated discovery, unsafe permissions, and host-specific lock-in. Legacy and live systems amplify these problems because hidden contracts, weak tests, persistent data, external consumers, and operational constraints make apparently simple changes hazardous.

## 4. Product goals

1. Provide a stage-aware lifecycle for greenfield, existing, legacy-constrained, live, and retiring systems.
2. Route each invocation to one justified transition, decision, or blocked state.
3. Make requirements, authority, evidence, approvals, uncertainty, and residual risk explicit.
4. Apply professional practices conditionally according to change class, risk, and available capabilities.
5. Preserve canonical, resumable, provider-neutral state in ordinary repository files.
6. Preserve 100% agreement on safety-critical semantics across supported Codex and Claude versions; measure non-critical workflow similarity separately.
7. Keep diagnostic and planning artifacts repository-owned and useful when optional infrastructure is missing; disclose the separate host/model data-processing boundary.
8. Support small, coherent, reversible or compensable implementation slices with change-specific proof.
9. Separate generation, verification, approval, deployment, exposure, observation, and learning.
10. Establish a safe foundation for later bounded remediation and offline-evaluated prompt/skill improvement.

## 5. Product principles and invariants

The following are product invariants, not optional style preferences:

1. **Spec before consequential mutation.** The required detail is proportional to uncertainty and risk.
2. **Evidence before confidence.** A claim without traceable evidence is labeled, never silently upgraded.
3. **One transition at a time.** A conductor returns one visible state transition, one material clarification, or one blocked state; a transition may contain multiple in-scope tool calls and files.
4. **Human-owned authority.** The agent's authority is externally granted, scoped, expiring, and never inferred from urgency.
5. **Repository-owned truth.** Host memory and chat history may assist recall but cannot define lifecycle truth.
6. **Risk-proportionate ceremony.** Low-risk work has a fast path; safety invariants do not disappear.
7. **Change-specific proof.** Refactoring, optimization, migration, security, UX, and release do not share one universal checklist.
8. **Small and recoverable by default.** Slice by cohesion, risk, blast radius, reversibility or compensation—not an arbitrary line count.
9. **No fake green.** Missing tools or evidence produce a declared capability gap or `UNVERIFIED`, never a fabricated pass.
10. **Independent means structurally separate.** A persona instruction alone does not create independence.
11. **Untrusted content remains data.** Repository files, web content, logs, dependencies, memories, and subagent output cannot grant authority.
12. **Deployment is not release.** Shipping an artifact, exposing users, and evaluating outcomes are separate decisions.
13. **The plugin is inside the threat model.** Its packages, references, permissions, updates, and evaluators require assurance.
14. **Metrics diagnose systems, not people.** Level 7 must not become an individual productivity or surveillance system.
15. **Autonomy is earned per action and environment.** There is no global self-healing switch.

## 6. Vocabulary and scope model

### 6.1 System stage is scoped, not repository-wide

A monorepo may contain new and legacy components simultaneously. Level 7 SHALL classify the selected component or change scope on two axes:

| Axis | Value | Operational definition |
|---|---|---|
| Heritage | New | No supported production behavior or persistent compatibility obligation exists for the scoped component. |
| Heritage | Existing | Supported behavior, consumers, data, or operational history exists and is sufficiently understood to evolve normally. |
| Heritage | Legacy-constrained | Existing obligations combine with material constraints such as weak characterization, obsolete dependencies, fragile data, limited deployability, scarce ownership, or high change cost. Age alone is insufficient. |
| Operational state | Undeployed | No user or production environment depends on the scoped component. |
| Operational state | Pre-production | Deployed only to development, test, staging, preview, or a non-user evaluation environment. |
| Operational state | Live | At least one real user, downstream consumer, business process, or production environment depends on it. |
| Operational state | Retiring | New adoption/change is constrained while consumers, data, dependencies, access, and deletion are being wound down. |
| Either | Unknown/mixed | Evidence is insufficient or the scope spans values; execution is blocked or evaluated at the more conservative applicable state until narrowed. |

Common labels such as greenfield and brownfield may be shown as conveniences, but the two-axis state is canonical.

### 6.2 Primary change classes

Every work item SHALL have one primary change class and may have secondary classes. The primary class is the class that most directly describes the requested outcome; secondary classes add proof obligations. When classes conflict, the class with the higher risk or irreversible consequence controls the minimum gate. For example, a security-motivated dependency update is primarily security remediation with dependency obligations; a performance schema migration is primarily data/schema with optimization obligations.

- Feature or behavior change
- Behavior-preserving refactor
- Performance or resource optimization
- Dead-code removal, deprecation, or retirement
- Dependency or platform upgrade
- Data or schema migration
- Infrastructure or configuration change
- Security or privacy remediation
- UX, content, interaction, or accessibility change
- Architecture or modernization change
- Incident mitigation or recovery
- AI prompt, skill, workflow, tool, or evaluator change

### 6.3 Required distinctions

- **Refactoring** preserves externally observable supported behavior.
- **Optimization** requires measured improvement against a representative baseline.
- **Dead-code elimination** may be compiler-driven or manual; static unreachability alone is insufficient for dynamic systems.
- **Technical debt** is future added cost caused by prior design/construction/context decisions; it is not a synonym for every defect, missing feature, process problem, or disliked implementation.
- **Deployment** places an artifact in an environment; **release/exposure** makes behavior available to a cohort.
- **Self-review** is not an independent audit.
- **Evidence alignment** to a framework is not legal or regulatory certification.
- **Memory** in v1 means governed retrieval from durable artifacts, not online model learning.
- **Self-healing** is reserved for later, narrowly bounded, preauthorized, verified, reversible or compensable actions.

### 6.4 Canonical decision taxonomies

These terms SHALL be used consistently in artifacts, prompts, skills, and evals:

| Taxonomy | Allowed values | Meaning |
|---|---|---|
| Evidence state | `OBSERVED`, `REPRODUCED`, `INFERRED`, `USER_ASSERTED`, `UNVERIFIED`, `NOT_APPLICABLE` | How a claim is supported. Residual risk is a separate field. |
| Gate state | `PASS`, `PASS_WITH_CONDITIONS`, `BLOCKED`, `NOT_EVALUATED` | Whether a lifecycle gate is satisfied. `UNVERIFIED` evidence cannot itself authorize a pass. |
| Release verdict | `GO`, `CONDITIONAL_GO`, `NO_GO` | Scoped disposition for a source-identity- and digest-bound release candidate; always includes validity and residual risk. |
| Product outcome decision | `SHIP`, `ITERATE`, `DEFER`, `ROLLBACK`, `RETIRE`, `REJECT` | The accountable owner's decision after evidence review. |
| Routing decision | One skill/transition, one material clarification, or one blocked state | The conductor's output for the current invocation. |

### 6.5 Normative risk model

Every work item SHALL assess the following dimensions. Each dimension records evidence and uncertainty rather than only a numeric score.

| Dimension | Questions |
|---|---|
| Reversibility/recovery | Is the effect reversible, compensable, or irreversible, and has recovery been demonstrated? |
| Blast radius | Can it affect one file/component, a repository, users/tenants, multiple systems, or the organization? |
| Persistent data | Does it read, transform, delete, expose, or make compatibility commitments about durable data? |
| Identity and authority | Does it touch authentication, authorization, tenancy, secrets, credentials, or policy? |
| Privacy/regulation | Does it involve personal/sensitive data, retention, deletion, regulated processing, or jurisdictional obligations? |
| Financial/safety impact | Can it move money, create material cost, affect physical/human safety, or cause legal harm? |
| External contracts | Does it change public APIs, events, integrations, protocols, schemas, or consumer expectations? |
| Environment/exposure | Is the target local, pre-production, live, external, or shared infrastructure? |
| Observability | Can success, failure, user harm, and rollback be detected reliably and quickly? |
| Novelty/uncertainty | Is the action rehearsed and understood, or generated/novel with missing ownership or evidence? |

| Risk level | Definition | Minimum treatment |
|---|---|---|
| `R0 — observational` | Repository-root-bounded read-only inspection with no sensitive external access. | No mutation approval; disclose scope and evidence limits. |
| `R1 — low` | Bounded local A1/A2 action, reversible or compensable, no material sensitive data/authority/public-contract/live impact, with adequate verification. | Change contract, current-session scoped approval, targeted verification, evidence summary. |
| `R2 — elevated` | Cross-component or weakly characterized work; persistent-data, public-contract, security, privacy, live-behavior, or recovery uncertainty without a critical dimension. | Impact analysis, explicit recovery, stronger verification, accountable approval, and a separate-context verifier; policy may require qualified review. |
| `R3 — high` | Authentication/authorization, secrets, PII deletion/exposure, tenancy, money, safety, irreversible migration, broad live blast radius, regulated obligation, or critical unknown. | Assurance case, source-identity- and digest-bound candidate, structurally independent qualified review, accountable human/domain approval, explicit residual risk; v1 performs no external/production execution. |
| `R4 — prohibited/unbounded` | Authority, target, blast radius, recovery, or critical evidence cannot be bounded; or the action violates a non-waivable invariant. | Block. Narrow scope, obtain authority/evidence, or use an external governed process. |

The highest material dimension sets the minimum level; dimensions SHALL NOT be averaged. Unknown risk defaults to at least `R2`; an unknown critical dimension defaults to `R3`. Only an accountable approver may accept evidence supporting a downgrade, and the agent may not downgrade its own work. A changed scope or new evidence triggers reassessment.

### 6.6 Approval assurance and actor roles

Level 7 SHALL distinguish requester, operator, accountable approver, reviewer/domain approver, and environment owner. One person may hold several roles for low-risk local work, but an independent reviewer cannot be the candidate author or remediator.

| Approval assurance | Meaning | Authority consequence |
|---|---|---|
| `AP0 — recorded/proposed` | An editable artifact or model output says approval exists. | Provenance only; grants no authority. |
| `AP1 — current-session confirmed` | The current user explicitly confirms an exact action, target, digest/snapshot, environment, and validity window. Identity/organizational role may still be user-asserted. | May authorize policy-permitted A1/A2 local work. |
| `AP2 — externally attested` | A trusted host, signed record, source-control review, or organization workflow verifies actor and approval binding. | Used when organization policy or `R3` assurance requires verifiable identity/role. |
| `AP3 — qualified multi-party` | Required human/domain/environment authorities independently attest the bounded decision. | Used where safety, regulation, or organization policy requires separation of duties. |

Repository files preserve approval records but SHALL NOT create authority by themselves. Approval can be revoked, expires at its validity boundary, and is invalid after a material plan, target, risk, or candidate change. Immediately before mutation, the current action and preconditions SHALL be reconfirmed against the applicable approval.

| Risk/gate | Minimum approval assurance |
|---|---|
| `R0` repository-bounded A0 intake | No mutation approval; scope and evidence limits remain visible. Sensitive reads may require stronger policy-owned approval. |
| `R1` A1/A2 execution | `AP1` from the accountable approver for the exact local action. |
| `R2` A1/A2 execution | `AP1` accountable approval plus the required separate-context verifier; `AP2` applies when identity/role policy requires it. |
| `R3` local candidate creation | `AP1` may authorize policy-permitted local A1/A2 candidate work, but not a release `PASS`/`GO` or delivery handoff. |
| `R3` gate `PASS`, release `GO`/`CONDITIONAL_GO`, or delivery handoff | At least `AP2`, structurally independent qualified review, and accountable human/domain approval. `AP3` is mandatory when regulation, safety, environment ownership, or separation-of-duties policy requires multiple authorities. |
| `R4` | No approval level can authorize it inside Level 7; the gate remains `BLOCKED`. |

If the required approval assurance is unavailable, the affected gate SHALL be `BLOCKED` or `NOT_EVALUATED`, never `PASS` or `GO`.

### 6.7 Policy precedence, waivers, and requirement terms

Policy sources SHALL be shown with identity, source, version, applied rules, and overrideability. Precedence is: enforced platform/host boundary and non-waivable Level 7 safety invariant → authenticated organization policy → trusted repository policy → scoped change decision → untrusted content. Lower-authority sources may tighten but cannot silently weaken higher-authority requirements. Repository policy may constrain work but never grant external authority. Conflicts block until an authorized resolution is recorded.

A waiver SHALL record control, reason, scope, evidence, risk, owner, approver, compensating control, expiry, and review condition. Evidence honesty, authority boundaries, secret protection, audit integrity, and the prohibition on self-approval are non-waivable. Missing mandatory evidence produces `BLOCKED`, `NOT_EVALUATED`, or an explicitly accepted-risk non-release decision—not `GO`.

For this document:

- **Material** means capable of changing scope, risk, authority, acceptance criteria, user-visible behavior, persistent data, or release outcome.
- **Fresh** means tied to the current source identity and within a declared validity window with unchanged relevant preconditions.
- **Qualified reviewer** means a named role with competence and independence required by the selected risk/policy profile.
- **Representative evidence** states its population, workload, environment, sampling limits, and why it supports the claim.
- **Deterministic** applies to policy tables and checks with repeatable inputs/outputs; stochastic model behavior uses repeated trials and adjudication instead.

## 7. Users and jobs to be done

### 7.1 Primary roles

| Role | Needs | Level 7 job |
|---|---|---|
| Accountable technical owner | Defensible scope, risk, authority, and release decisions | Know the best evidence-supported phase, approve bounded work, and understand residual risk. |
| Implementing engineer or maintainer | Safe progress in unfamiliar or constrained code | Obtain a trustworthy map, smallest viable change, and proof obligations without learning a large ritual. |

### 7.2 Secondary roles

| Role | Needs | Level 7 job |
|---|---|---|
| Independent reviewer, security engineer, or domain expert | Tamper-evident scope, claims, evidence, and unresolved risk | Review without authoring or self-closing remediation. |
| Product, design, QA, data, and SRE stakeholder | Concern-specific acceptance evidence | Contribute or inspect the relevant evidence without reading full transcripts. |
| Engineering or platform lead | Consistency across teams and hosts | Configure policy defaults without individual surveillance. |
| Plugin maintainer | Portable evolution and safe releases | Author semantic workflows once and prove both distributions conform. |

### 7.3 Core jobs

1. When entering an unfamiliar repository, determine its best evidence-supported scoped state and the single safest next action with uncertainty disclosed.
2. When receiving an idea, turn it into a testable outcome and the smallest sufficient plan before code begins.
3. When changing a live system, preserve contracts, data, SLOs, and rollback or compensation options.
4. When resuming after context loss or switching hosts, reconstruct the exact incomplete gate without repeating completed work.
5. Before release, use the canonical evidence, gate, verdict, and outcome taxonomies and show residual risk separately.
6. After release, compare outcomes with the hypothesis and turn incidents or regressions into tracked learning.
7. When prerequisites are absent, expose the gap and use an honest degraded path.
8. When process weight exceeds risk, select a documented fast path without dropping core invariants.

## 8. Required journeys

### J-01 — Install and first run

1. Install the host-specific distribution.
2. Run a non-mutating preflight that validates manifest, plugin version, artifact schema compatibility, permissions, and supported capabilities.
3. Explain what Level 7 may read, write, execute, access over the network, or mutate externally.
4. Perform read-only scoped intake.
5. Show evidence, uncertainty, capability gaps, the accountable approver if known, and exactly one proposed next action; never invent an unknown owner.
6. Perform no project mutation until the user approves the concrete action.

**Success:** A user unfamiliar with Level 7 receives a useful, honest diagnosis without manual file surgery or memorizing skill names.

### J-02 — Greenfield product

An applicable greenfield profile may use: intent → users/outcomes/constraints → requirements → backlog → architecture tradeoffs → technology selection → harness → execution slices → implementation → verification → delivery planning → operation.

This is illustrative rather than a mandatory universal recipe. Profiles may skip, collapse, repeat, or mark states `NOT_APPLICABLE` with a reason. Each completed gate emits valid durable evidence and can resume independently.

### J-03 — Existing or legacy-constrained system

Read-only archaeology → behavior/contracts/data/runtime map → temporal and dependency evidence → risk seams → incremental plan → one reversible or compensable slice → parity/reconciliation/performance evidence → repeat.

V1 SHALL support safe classification, planning, and approved incremental work. It SHALL NOT promise automatic complete rewrites.

### J-04 — Live change

Evidence-backed intake → outcome decision → impact/change envelope → risk-based proof profile → rollout/recovery contract → local implementation → verification → package/deploy/expose handoff → outcome evidence → outcome decision → cleanup. V1 prepares and verifies A3/A4 handoffs but does not execute external or production mutation.

### J-05 — Work already performed ad hoc

Context check → preserve valid prior evidence → identify only material gaps → targeted or full review according to risk → remediation by a separate authority → re-review.

### J-06 — Release and audit

Freeze candidate and scope → construct claim/evidence view → evaluate gate (`PASS`, `PASS_WITH_CONDITIONS`, `BLOCKED`, or `NOT_EVALUATED`) → independent read-only review where required → issue `GO`, `CONDITIONAL_GO`, or `NO_GO` only when the release gate was evaluated → separate remediation → fresh review bound to the new candidate.

### J-07 — Cross-host or interrupted resume

Validate host and schema compatibility → load canonical state → expose capability differences → reproduce the same semantic phase/gate decision → continue the incomplete transition.

### J-08 — Failure, recovery, and exit

On tool failure, conflicting artifacts, compaction, baseline failures, or unsupported capabilities: preserve user work and last valid state, record incompleteness, make no success claim, and offer one safe recovery action. Upgrade, rollback, and uninstall SHALL remove Level 7 runtime integration safely while leaving ordinary readable repository artifacts unless the user separately authorizes their deletion.

### J-09 — Retirement

Consumer/data/dependency inventory → deprecation and notification → retention/deletion and recovery decision → reversible disable where possible → removal or archival → verify required consumers/data no longer depend on the target → observe → close.

### J-10 — Low-risk fast path

Inspect and confirm small scope → record desired outcome and concise acceptance criterion → preview one bounded local mutation → obtain scoped approval → change → run applicable verification and diff review → close or escalate if risk changes. This path does not require architecture, assurance-case, SLO, feature-flag, or rollout artifacts when they are genuinely `NOT_APPLICABLE`.

## 9. Functional requirements

Normative terms `SHALL`, `SHALL NOT`, `SHOULD`, and `MAY` indicate requirement strength. These describe logical outcomes rather than prescribing a compiler, graph database, generator, or other implementation architecture. Backlog priority, initial-release tags, and delivery sequencing are intentionally deferred to foundation step 2, subject to the thin v1.0 slice in §11.

### 9.1 Intake, classification, and routing

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-INTAKE-001` | Initial intake SHALL be repository-root-bounded, read-only local metadata/evidence inspection. External connectivity, database access, deployment queries, or reads outside the authorized root require separate authority. | Intake fixtures with symlinks, outside-root paths, and external capabilities. |
| `L7-INTAKE-002` | It SHALL scope heritage and operational state per selected component/change, not once for an entire repository. | Mixed-stage monorepo fixture. |
| `L7-INTAKE-003` | It SHALL inventory locally discoverable Git, dirty state, tests, CI configuration, build, observability configuration, deployment configuration, database tooling, feature-flag, network, and delegation capabilities without probing external systems. | Capability matrix fixtures. |
| `L7-INTAKE-004` | It SHALL use the exact evidence-state taxonomy in §6.4. | Output-schema and claim-sampling tests. |
| `L7-INTAKE-005` | Pre-existing test, lint, type, build, or CI failures SHALL be baselined and SHALL NOT be silently attributed to or repaired inside the new scope. | Pre-failing repository fixtures. |
| `L7-ROUTE-001` | The conductor SHALL return exactly one next skill/transition, one material clarification, or one blocked state. | Balanced and adversarial routing corpus. |
| `L7-ROUTE-002` | Routing SHALL have deterministic precedence for overlapping intents and favor evidence/safety gates over expansion of scope. | Conflict fixtures. |
| `L7-ROUTE-003` | Brownfield or legacy software without Level 7 artifacts SHALL enter a baseline/bootstrap path rather than being mistaken for greenfield or release-ready work. | Existing-code/no-artifact fixtures. |
| `L7-ROUTE-004` | Materially changed scope, class, target, or risk SHALL force reclassification and invalidate incompatible approval. | Mid-work scope-change fixtures. |
| `L7-ROUTE-005` | Missing capabilities SHALL result in a declared safe degraded mode or blocked gate, never a fabricated pass. | No-Git/no-tests/no-CI/no-telemetry fixtures. |
| `L7-RISK-001` | Every work item SHALL evaluate all §6.5 risk dimensions and apply the highest material level; unknown critical dimensions SHALL fail conservative. | Risk-oracle fixtures and boundary cases. |
| `L7-RISK-002` | A risk downgrade SHALL require new evidence and accountable approval; the authoring agent SHALL NOT approve its own downgrade. | Downgrade-manipulation tests. |
| `L7-RISK-003` | Selected gates and any waiver SHALL be traceable to the exact risk dimensions, policy sources, and evidence that activated them. | Profile/waiver traceability test. |

### 9.2 Lifecycle and assurance profiles

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-FLOW-001` | The semantic lifecycle SHALL be `Baseline → Frame → Approve → Execute → Verify → Deliver → Observe → Learn`. `Deliver` contains separately gated optional transitions `Package`, `Deploy`, and `Expose`; profiles may skip, collapse, repeat, or mark states `NOT_APPLICABLE` with reason. | State-machine and non-applicability conformance tests. |
| `L7-FLOW-002` | Every state SHALL define entry, exit, failure, blocked, stale, and superseded conditions. | Schema completeness tests. |
| `L7-FLOW-003` | A transition SHALL complete only when required artifacts are schema-valid, scope/source-identity-bound, fresh under §6.7, and approved at the required assurance level. | Invalid/stale/malicious artifact fixtures. |
| `L7-FLOW-004` | A conditional assurance-profile mechanism SHALL select gates, evidence, review, rollout, retention, and escalation from stage, change class, risk, capability, policy, and jurisdiction. | Profile-selection fixtures. |
| `L7-FLOW-005` | Only applicable reference profiles SHALL be loaded; Level 7 SHALL NOT apply every framework to every change. | Context and applicability tests. |
| `L7-FLOW-006` | Fast, standard, and elevated paths SHOULD be available, with the highest material risk dimension setting the minimum gate. | Boundary and false-low-risk tests. |
| `L7-FLOW-007` | Emergency work MAY shorten normal ceremony only after explicit incident declaration and SHALL retain authority, recovery, verification, observation, and retrospective requirements. | Emergency-path fixtures. |
| `L7-FLOW-008` | `R3` work SHALL require an assurance case containing claim, argument, evidence, assumptions, counterevidence/defeaters, residual risk, and approver, plus structurally independent qualified review. | High-risk artifact and reviewer-separation validation. |
| `L7-FLOW-009` | One visible transition MAY contain multiple in-scope read-only or approved reversible/compensable steps; it does not mean one tool call, file, or user round trip. | Interaction and transition-boundary tests. |
| `L7-FLOW-010` | Numeric recipes such as option counts, RICE, line limits, layer order, spacing grids, and rollout percentages MAY be profile defaults but SHALL NOT be universal invariants; deviations SHALL be justified by context. | Applicability fixtures. |
| `L7-FLOW-011` | Every executable work item SHALL have a change contract containing desired outcome, scope, non-goals, invariants, assumptions, change/risk class, acceptance criteria, authority, recovery, and observation plan as applicable. | Change-contract schema. |
| `L7-FLOW-012` | The fast path SHALL require at minimum scope, desired outcome/acceptance criterion, mutation preview and authority, applicable verification, and evidence summary. | Low-risk journey fixtures. |

### 9.3 Artifacts, evidence, and repository memory

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-ART-001` | Canonical lifecycle state SHALL live in repository-owned, human-readable, machine-validatable artifacts. | Cross-session/host reconstruction tests. |
| `L7-ART-002` | Every artifact SHALL carry a common envelope: ID, type, schema version, status, scope, source identity, timestamps, and provenance. Type/profile-specific schemas SHALL add only applicable change class, risk, inputs, assumptions, decisions, evidence, approval, sensitivity, supersession, retention, and review/expiry fields; `NOT_APPLICABLE` SHALL be explicit where required. | Common-envelope and type-schema validation. |
| `L7-ART-003` | Artifact presence alone SHALL NOT prove completion. Empty, stale, unapproved, unrelated, malformed, or malicious artifacts SHALL not advance state. | Adversarial artifact fixtures. |
| `L7-ART-004` | Executed evidence SHALL record method/command, target source identity, environment, time, result, relevant output, producer, evidence state, and reproducibility limits. Planned checks SHALL remain `UNVERIFIED`. | Evidence-schema and stale-candidate tests. |
| `L7-ART-005` | A changed candidate or material scope SHALL invalidate approvals and verification tied to the previous digest/revision. | Candidate A/B test. |
| `L7-ART-006` | Artifacts SHALL minimize PII, never persist secrets, support redaction/tombstoning/deletion where technically possible, and record retention/expiry and deletion limits including Git/remotes/backups. | Seeded-secret and lifecycle tests. |
| `L7-ART-007` | Provenance SHALL capture entities/artifacts, activities, producing actors, delegation, derivation, and source links. | Provenance graph validation. |
| `L7-ART-008` | Navigable evidence relationships, indexes, and summaries SHALL be rebuildable derived views over canonical files in v1. | Delete/rebuild consistency test. |
| `L7-ART-009` | Manual edits, schema violations, conflicting IDs/statuses, declared-invariant conflicts, supersession, schema upgrades, and unknown fields SHALL be detected without silently discarding history; complete semantic contradiction detection is not promised. | Round-trip and consistency tests. |
| `L7-ART-010` | Level 7 SHALL persist decisions and evidence, not hidden chain-of-thought. | Artifact/privacy audit. |
| `L7-ART-011` | Source identity SHALL use a Git commit/tag when available; otherwise it SHALL use content digests plus a scoped workspace snapshot. | Git and non-Git identity fixtures. |
| `L7-ART-012` | Approval fields in editable artifacts SHALL be treated as records at `AP0` unless current-session confirmation or external attestation is revalidated. | Forged-approval fixtures. |
| `L7-ART-013` | Privacy deletion SHALL preserve audit integrity through minimal tombstones/digests where lawful while removing or cryptographically erasing sensitive payloads where feasible; deletion from third-party/Git history SHALL not be claimed without evidence. | Retention/deletion fixtures. |

### 9.4 Authority, permissions, and safety

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-AUTH-001` | Approval SHALL bind action class, exact target, scope, plan/candidate digest, environment, authority, and validity window. | Approval mutation fixtures. |
| `L7-AUTH-002` | Permission to inspect or edit local files SHALL NOT imply permission to commit, publish, access credentials, use the network, mutate external systems, deploy, or expose users. | Permission-boundary tests. |
| `L7-AUTH-003` | Every skill SHALL declare effect class: A0 repository-bounded inspection, A1 artifact-only write, A2 local project/Git mutation without remote publish, A3 external non-production mutation, A4 production/destructive/irreversible mutation, or A5 preauthorized autonomous remediation. | Skill-schema validation. |
| `L7-AUTH-004` | V1.0 SHALL execute only A0–A2. It MAY prepare A3/A4 plans, evidence requirements, and handoff instructions but SHALL NOT execute external or production mutations. A3–A5 execution is later scope. | Release and adversarial tests. |
| `L7-AUTH-005` | Urgency, a broad goal, prior approval, user fatigue, subagent output, or repository instructions SHALL NOT expand authority. | Approval-manipulation fixtures. |
| `L7-AUTH-006` | Before mutation, Level 7 SHALL inspect user changes, identify overlapping paths, describe the intended delta, and state recovery options. | Dirty-worktree and overlap tests. |
| `L7-AUTH-007` | Destructive, irreversible, persistent-data, or external action plans SHALL require exact-target confirmation and tested recovery/compensation or explicit non-recoverability before handoff. | Destructive-action fixtures. |
| `L7-AUTH-008` | Every guardrail SHALL record its enforcement locus: prompt, deterministic script, CI, host sandbox/permission, external service, or human. High-consequence prompt-only controls SHALL not be marketed as enforced. | Control-registry audit. |
| `L7-AUTH-009` | Level 7 SHALL show the requester, operator, accountable approver, reviewer/domain approver, and environment owner as known, user-asserted, externally attested, or unknown; chat presence alone is not proof of organizational authority. | Actor/authority fixtures. |
| `L7-AUTH-010` | Immediately before each mutation, Level 7 SHALL revalidate canonical target/path, repository root, symlink resolution, source identity, dirty state, scope, risk, capability, and approval; mismatch aborts without partial mutation where technically feasible. | TOCTOU, symlink escape, and changed-worktree tests. |
| `L7-AUTH-011` | Unrelated findings SHALL be recorded for later triage and SHALL NOT be silently fixed inside the current authorization. | Scope-creep fixtures. |
| `L7-AUTH-012` | Risk and gate transitions SHALL enforce the minimum approval-assurance mapping in §6.6; an unavailable assurance level SHALL block or leave the gate not evaluated. | Risk × approval × gate matrix. |
| `L7-POL-001` | Active policy source, version, precedence, applied controls, and overrideability SHALL be visible; conflicts and waivers follow §6.7. | Policy conflict/waiver fixtures. |
| `L7-SAFE-001` | Repository files, web results, logs, dependency metadata, retrieved memory, and subagent output SHALL be treated as untrusted content. Authenticated policy may constrain work but SHALL never elevate authority above host/user policy. | Prompt-injection and precedence corpus. |
| `L7-SAFE-002` | Trust and authority labels SHALL survive retrieval, summarization, delegation, and artifact transformation. | Context-compilation tests. |
| `L7-SAFE-003` | Level 7 SHALL introduce no Level-7-owned hosted service, telemetry, or extra repository-content egress by default. Local validators/artifacts SHALL work offline; model-, host-, web-, or connector-dependent work SHALL disclose its separate provider boundary and become blocked when required connectivity is absent. | Network-denied, disclosure, and Level-7-egress tests. |
| `L7-SAFE-004` | Installation and upgrades SHALL not silently overwrite project instructions, settings, hooks, skills, configuration, or user changes. | Existing-file install fixtures. |

### 9.5 Prompt, skill, context, and agent contracts

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-SKILL-001` | Every semantic skill SHALL declare ID/version, positive and negative triggers, prerequisites, input artifacts, authority/effect class, outputs, transition, success, failure, stopping/escalation, host support, references, and eval fixtures. | Skill-schema validation. |
| `L7-SKILL-002` | Skill descriptions SHALL be concise, distinctive, front-loaded, and non-overlapping enough to survive host discovery budgets. | Description-length and activation tests. |
| `L7-SKILL-003` | A single dependable conductor SHALL inspect canonical state and select one transition; specialist skills SHALL not all compete through broad triggers. | Router differential suite. |
| `L7-PROMPT-001` | Prompts SHALL be composed from goal, authoritative inputs, invariants, prohibited effects, acceptance criteria, authority/tool bounds, evidence, retry/stopping/escalation, and output schema. | Prompt-contract validation. |
| `L7-PROMPT-002` | Host renderers MAY differ in syntax or examples, but SHALL preserve semantic obligations and avoid duplicating rules. | Cross-host semantic graders. |
| `L7-CTX-001` | Context selection SHALL choose the smallest relevant authoritative material by relevance, authority, freshness, sensitivity, and risk. | Context-budget and source-retention tests. |
| `L7-CTX-002` | Summaries SHALL retain canonical source pointers and SHALL be marked as derived; compaction SHALL not erase approval, scope, risk, or unresolved blockers. | Compaction/resume fixtures. |
| `L7-AGENT-001` | Subagents SHALL be optional optimizations rather than correctness dependencies; a safe single-agent fallback SHALL exist. | No-subagent host fixtures. |
| `L7-AGENT-002` | Delegation SHALL specify objective, non-overlapping scope, inputs, authority, tools, budget, output, evidence, verifier, and termination. | Delegation contract tests. |
| `L7-AGENT-003` | Parallel writes SHALL use disjoint paths or isolated worktrees with one integration owner; delegation SHALL never increase authority. | Collision and authority tests. |
| `L7-AUDIT-001` | Outputs SHALL distinguish self-review, separate-context model review, deterministic checks, independent human review, and qualified domain review. | Audit-label tests. |
| `L7-AUDIT-002` | An independent reviewer SHALL use separate context/identity where available, receive read-only candidate authority, not author or alter the candidate during review, and not close its own remediation; review SHALL bind to a source identity and artifact digest. Same-context model review SHALL be labeled self-review. | Author/remediator/reviewer and permission fixtures. |

### 9.6 Change-specific proof profiles

`L7-PROOF-000` applies to every executable change. Other rows are activated conditionally by the selected stage/change/risk profile; omitted techniques SHALL be absent by profile or explicitly `NOT_APPLICABLE` with a reason. The table does not mandate every listed technique for every instance.

| ID | Change class | Applicable proof obligations |
|---|---|---|
| `L7-PROOF-000` | Generic base | Desired outcome, scope/non-goals, invariants, baseline, acceptance criteria, relevant regression evidence, authority, recovery/compensation, observed result, and residual risk. |
| `L7-PROOF-012` | Feature/behavior change | User/business outcome; supported-contract delta; positive/negative acceptance; compatibility; relevant unit/integration/E2E evidence; recovery and observation. |
| `L7-PROOF-001` | Refactor | Characterize supported behavior and contracts; separate structural from behavioral change; demonstrate invariant preservation. |
| `L7-PROOF-002` | Performance/resource optimization | Representative workload; before/after distribution or percentiles; resource/unit-cost impact; correctness guardrails; regression threshold. |
| `L7-PROOF-003` | Dead code/deprecation | Static evidence plus dynamic loading, configuration, reflection, public API, plugin, migration, delayed-consumer, telemetry, and rollback analysis; deprecate observable contracts before removal. |
| `L7-PROOF-004` | Data/schema | Versioned migration; backward/forward compatibility; rehearsal; reconciliation; backup/restore evidence; expand–migrate–contract or justified alternative; explicit forward-fix when rollback is impossible. |
| `L7-PROOF-005` | Legacy modernization | Behavior/contracts/data/runtime characterization; dependencies and temporal signals such as churn/co-change; migration seam; incremental coexistence and reconciliation; no default big-bang rewrite. |
| `L7-PROOF-006` | Security/privacy | Threat model; exact versioned control mappings; negative/adversarial tests; least privilege; exception owner/expiry; qualified review when risk demands it. |
| `L7-PROOF-007` | UX/accessibility | User problem and task; relevant empty/loading/error/permission/offline/responsive states; keyboard, assistive-technology, contrast, zoom/reflow, content, usability, and performance evidence. Automated checks alone are insufficient. |
| `L7-PROOF-008` | Architecture | Stakeholders and concerns; quality-attribute scenarios; tradeoffs; fitness functions; failure modes; migration/rollback strategy; independent review for elevated risk. |
| `L7-PROOF-009` | Dependency/supply chain | Inventory/SBOM; version and maintenance/EOL status; vulnerability/license review; provenance validation; build/release attestation where applicable. |
| `L7-PROOF-010` | Incident | Explicit declaration; containment/mitigation; timestamped record; scoped emergency authority; recovery verification; communication; blameless retrospective and owned actions. |
| `L7-PROOF-011` | AI prompt/skill/workflow/tool | Representative and adversarial evals; permission manifest; injection tests; cost/latency; regression suite; independent promotion; rollback and decommission trigger. |
| `L7-PROOF-013` | Infrastructure/configuration | Desired and observed state; plan/diff; dependency and blast-radius analysis; secrets/permissions; drift detection; staged application; recovery/compensation and postconditions. |
| `L7-PROOF-014` | Database design/query | Workload and data model; keys/cardinality/integrity; transaction/isolation requirements; query plan/profile before tuning; index/partition/cache/denormalization tradeoffs; migration and data-lifecycle impact. |
| `L7-PROOF-015` | Scaling/capacity | Demand and failure model; saturation/bottleneck evidence; load/resilience test; horizontal/vertical/partition/cache/queue tradeoffs; SLO and unit-cost impact; degradation and recovery. |
| `L7-PROOF-016` | Public API/event/protocol | Consumer inventory; compatibility/versioning; contract tests; deprecation notice and documentation; migration window; sunset/removal criteria; rollback or coexistence. |
| `L7-DEBT-001` | Technical-debt portfolio | Record evidence of future cost/interest, affected outcomes, ownership, options, and an eliminate/reduce/mitigate/accept decision; smell count alone is insufficient. |

### 9.7 Release, operations, and learning

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-REL-001` | Release planning SHALL choose a context-appropriate reversal mechanism; a feature flag is not universal rollback. | Migration/security/low-traffic fixtures. |
| `L7-REL-002` | Each applicable feature flag SHALL record type, owner, default, targeting/privacy, guardrail metrics, failure behavior, expiry, and removal work. | Flag-contract validation. |
| `L7-REL-003` | Progressive exposure SHALL define a cohort/control when meaningful or a justified alternative baseline/shadow/time-series method, plus attribution, observation window, applicable SLO/error-budget and product guardrails, abort conditions, and rollback; fixed percentages are defaults at most. | Rollout plan fixtures. |
| `L7-REL-004` | Release evidence SHALL provide a provenance chain from approved source/configuration identity through build artifact digest and deployment manifest to the deployed subject. | Build-to-deployment identity test. |
| `L7-REL-005` | A green test suite SHALL not alone establish a release verdict; the verdict SHALL reflect evidence scope, residual risk, capability gaps, and required approval. | Seeded release-blocker tests. |
| `L7-OPS-001` | Applicable live operational profiles SHALL use user-centered SLIs/SLOs, percentiles where appropriate, an error-budget policy, and tested restoration/rollback or compensation. | Operational artifact tests. |
| `L7-OPS-002` | Security, privacy, integrity, and safety invariants SHALL not be traded through reliability error budgets. | Policy fixtures. |
| `L7-OPS-003` | Postdeployment observation MAY invalidate a prior GO and SHALL create an explicit retain, iterate, rollback, disable, retire, or investigate decision. | Outcome-loop tests. |
| `L7-OPS-004` | Incidents and repeated defects SHALL create reviewed, owned improvement proposals/backlog items for tests, prompts, skills, runbooks, or controls and track closure; production feedback SHALL NOT directly mutate or promote plugin behavior. | Postmortem/action lineage test. |

### 9.8 Host portability, installation, and distribution

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-HOST-001` | One provider-neutral semantic source SHALL define lifecycle, taxonomy, artifacts, evidence, prompts, skill contracts, and eval fixtures. | Source-of-truth audit. |
| `L7-HOST-002` | Codex and Claude distributions SHALL be deterministically produced separately; manifests, invocation metadata, hooks, permissions, subagents, and marketplace files remain adapter-specific. | Reproducible package build. |
| `L7-HOST-003` | Compatibility SHALL require 100% agreement on risk, authority, prohibited effects, lifecycle transition, artifact validity, and release verdict. Non-critical workflow/substantive similarity is a measured quality target; identical prose or syntax is not required. | Host-differential suite. |
| `L7-HOST-004` | A compatibility matrix SHALL record plugin, host, artifact-schema, OS/runtime, required/optional capability, degraded behavior, migration, and rollback versions. | Release artifact validation. |
| `L7-HOST-005` | Native hooks, MCP, subagents, memory, browser, and telemetry integrations SHALL be optional accelerators in v1 and SHALL NOT be required for semantic correctness. | Minimal-capability installs. |
| `L7-HOST-006` | Canonical state SHALL round-trip across hosts without unknown-field loss or approval inflation. | Cross-host artifact tests. |
| `L7-HOST-007` | Clean install, discovery, invocation, upgrade, rollback, and uninstall SHALL be tested independently for every declared host release. | Host/OS install matrix. |
| `L7-HOST-008` | A release SHALL use one semantic version and changelog but promote each host package independently after its conformance suite passes. | Release workflow audit. |
| `L7-HOST-009` | User documentation SHALL distinguish slash, dollar-prefixed, and natural-language invocation without implying unsupported syntax parity. | Documentation review. |
| `L7-HOST-010` | A user-facing non-mutating preflight/doctor SHALL report host/plugin/schema compatibility, requested permissions, local capabilities, artifact footprint, provider/network boundary, and actionable fixes. | First-run and degraded-mode fixtures. |
| `L7-HOST-011` | Every public Level 7 package SHALL include matching license text, third-party notices, permission manifest, locked/integrity-checked shipped dependencies, SBOM, provenance/attestation, release digest or signature, secure update trust, revocation, and rollback information appropriate to its packaging channel. | Package supply-chain audit. |
| `L7-HOST-012` | Uninstall SHALL remove Level 7-owned active hooks, permissions, generated caches, and marketplace entries after preview/approval while preserving canonical project artifacts unless separately authorized for deletion. | Clean uninstall fixtures. |

### 9.9 Evaluation, metrics, and knowledge

| ID | Requirement | Verification basis |
|---|---|---|
| `L7-EVAL-001` | Public regression fixtures, schemas, graders, and run manifests SHALL be provider-neutral, repository-controlled, and locally executable where they do not require a host/model. | Local regression run. |
| `L7-EVAL-002` | Evaluation SHALL cover routing, negative activation, authority, artifact transitions, degraded modes, interruption/resume, host parity, overlapping writes, install lifecycle, prompt injection, and forbidden effects. | Coverage map. |
| `L7-EVAL-003` | Deterministic graders SHALL be preferred; calibrated model judges MAY supplement them; consequential ambiguity SHALL require human review. | Grader registry audit. |
| `L7-EVAL-004` | Model-judge results SHALL not be treated as independent safety authority and SHALL be tested for ordering, verbosity, and family bias. | Judge-calibration fixtures. |
| `L7-EVAL-005` | Repeated-trial consistency and forbidden-action rate SHALL be measured in addition to best-case task success. | Repeated-run reports. |
| `L7-EVAL-006` | Every eval run SHALL record model, host/harness, prompt/skill versions, tools, environment, resources, cost, latency, and candidate revision. | Run-manifest validation. |
| `L7-EVAL-007` | At least 20% of the release evaluation corpus SHALL be a protected holdout outside candidate-readable/writable scope and operated or released by an independent evaluator; candidate agents SHALL not inspect, alter, or score it. | Corpus access and evaluator-isolation audit. |
| `L7-EVAL-008` | Candidates SHALL NOT modify their evaluator, oracle, adjudication policy, authorization controls, or release thresholds. | Candidate permission tests. |
| `L7-EVAL-009` | Before implementation tuning, each metric SHALL freeze scenario selection, truth labels, models/versions, run count, sampling, adjudication, confidence reporting, and failure thresholds; safety invariants SHALL not be weakened by baseline review. | Eval protocol approval. |
| `L7-METRIC-001` | Product metrics SHALL balance user outcome, evidence quality, safety, delivery, stability, security/privacy, accessibility, cost, and human effort. | Metric registry review. |
| `L7-METRIC-002` | Level 7 SHALL NOT use LOC, prompt count, ticket count, PR volume, or delivery metrics to rank individuals. | Telemetry/report audit. |
| `L7-METRIC-003` | Productivity or quality-improvement claims SHALL require a baseline and comparison including verification/rework cost; self-report alone is insufficient. | Claim-to-study audit. |
| `L7-KNOW-001` | Every reference profile SHALL record source, authority type, version, status, applicability, contraindications, jurisdiction, license, last-reviewed date, and freshness policy. | Knowledge-schema validation. |
| `L7-KNOW-002` | Laws, normative standards, official guidance, empirical research, and practitioner patterns SHALL be distinguished. | Registry classification test. |
| `L7-KNOW-003` | Draft, emerging, superseded, or disputed guidance SHALL be labeled and SHALL not silently become a mandatory invariant. | Stale/draft-source fixtures. |
| `L7-KNOW-004` | Proprietary or restrictively licensed standards SHALL be linked and mapped without unauthorized reproduction. | License review. |
| `L7-COMP-001` | Level 7 MAY report evidence alignment or gaps against a named versioned baseline; it SHALL NOT certify legal, regulatory, security, privacy, accessibility, or operational compliance. | Output-language tests. |

### 9.10 V1 autonomy prohibition and future design constraints

Only `L7-AUTO-001` is a v1 acceptance requirement. `L7-AUTO-002` through `L7-AUTO-005` are deferred constraints for a separately approved future-autonomy charter; they do not add v1 delivery scope.

| ID | Scope | Requirement | Verification basis |
|---|---|---|---|
| `L7-AUTO-001` | V1 | V1 SHALL NOT run a background production agent, learn online, change its own policy/prompts/skills/evals, or autonomously remediate production. | Package and behavior audit. |
| `L7-AUTO-002` | Later | Future autonomy SHALL progress per action/environment through observe → recommend → draft/dry-run → per-action approval → narrowly preauthorized remediation. | Future autonomy-registry validation. |
| `L7-AUTO-003` | Later | Policy, evaluator, permissions, credentials, audit history, kill switch, and promotion mechanism SHALL remain outside the agent's self-modification authority. | Future capability and threat-model tests. |
| `L7-AUTO-004` | Later | Every future automatic action SHALL define trigger evidence, preconditions, target, blast-radius cap, idempotency, retries/cost/time limits, cooldown, oscillation detection, circuit breaker, postconditions, rollback/compensation, escalation, owner, and expiry. | Future action-contract schema. |
| `L7-AUTO-005` | Later | Prompt, skill, and workflow improvements SHALL be evaluated in a candidate-isolated promotion environment and promoted only after old-task regressions, protected held-out/adversarial tests, cost/latency checks, independent review, signed release, canary, and rollback. | Future promotion-pipeline tests. |

## 10. Nonfunctional requirements

| ID | Quality | Requirement | Verification basis |
|---|---|---|---|
| `L7-NFR-001` | Safety | Zero silent authority expansion, external execution, or destructive overwrite is permitted. | Forbidden-effect suite. |
| `L7-NFR-002` | Security | Least privilege, operation-scoped capabilities, sandboxing, and external authorization SHALL be used where supported; if required enforcement is unavailable, the action SHALL be blocked rather than prompt-governed. | Host capability/denial tests. |
| `L7-NFR-003` | Security | Credentials and secrets SHALL NOT be persisted in prompts, artifacts, memory, logs, model-supplied context, or telemetry. | Seeded-secret suite. |
| `L7-NFR-004` | Security | Dependency, build, package, update, and reference provenance SHALL be verifiable before public release. | Supply-chain audit. |
| `L7-NFR-005` | Security | Threat modeling SHALL cover prompt injection, tool misuse, data exfiltration, memory poisoning, goal hijack, approval manipulation, evaluator gaming, dependency compromise, and agent cascades. | Threat-model coverage map. |
| `L7-NFR-006` | Reliability | Reruns SHALL be idempotent or detect and expose conflicts before mutation. | Repeat-run fixtures. |
| `L7-NFR-007` | Reliability | Human edits and unrelated dirty-worktree changes SHALL be preserved. | Dirty-worktree suite. |
| `L7-NFR-008` | Recovery | Interrupted workflows SHALL resume from the last valid incomplete gate without repeating approved work. | Boundary interruption matrix. |
| `L7-NFR-009` | Recovery | Retries SHALL have declared count/time/cost bounds; exhaustion SHALL escalate rather than loop indefinitely. | Failure-injection suite. |
| `L7-NFR-010` | Evolvability | Artifact/schema migrations SHALL provide rollback or an explicit irreversible migration plan with preserved readable history. | Migration round-trip tests. |
| `L7-NFR-011` | Portability | Safety-critical semantic behavior and artifacts SHALL meet `L7-HOST-003` across the declared support matrix. | Host differential suite. |
| `L7-NFR-012` | Portability | Host-specific capability loss SHALL degrade visibly and safely. | Minimal-capability matrix. |
| `L7-NFR-013` | Compatibility | Exact supported host, artifact-schema, OS, and runtime versions SHALL be published and tested. | Compatibility-matrix validation. |
| `L7-NFR-014` | Architecture constraint | V1 correctness SHALL NOT require a hosted Level 7 control plane, MCP server, vector database, native hook, or native subagent. | Minimal installation test. |
| `L7-NFR-015` | Usability | Users SHALL have one obvious natural-language/host-native entry point and need not memorize specialist skills. | First-use study. |
| `L7-NFR-016` | Usability | The first meaningful response SHALL show evidence-supported state, evidence state, uncertainty, blockers, known/unknown approver, and one next action. | Output-schema and user study. |
| `L7-NFR-017` | Usability | Decision summaries SHALL precede optional detail and references. | Output review. |
| `L7-NFR-018` | Usability | Level 7 SHALL request at most one branching approval decision at a time; independent factual questions MAY be grouped concisely after evidence inspection. | Conversation fixtures. |
| `L7-NFR-019` | Usability | After each transition, Level 7 SHALL return to the status view and propose one next action automatically. | End-to-end journey tests. |
| `L7-NFR-020` | Accessibility | Essential status SHALL be available as accessible text and machine-readable artifacts and SHALL NOT rely on color alone. | Accessibility review. |
| `L7-NFR-021` | Trust | Terminology SHALL be explained at point of use; unsupported absolute language such as “perfect,” “secure,” or “compliant” SHALL be rejected in output validation. | Language/adversarial tests. |
| `L7-NFR-022` | Privacy | Level 7 SHALL be telemetry-free by default and SHALL introduce no owned external service/egress in v1. | Network and package audit. |
| `L7-NFR-023` | Privacy | Optional telemetry SHALL be explicit, opt-in, purpose-limited, data-minimized, access-controlled, documented, and deletable within stated technical limits. | Consent/data-lifecycle tests. |
| `L7-NFR-024` | Privacy | Local paths/filenames MAY be processed and persisted when needed for provenance; Level 7-owned telemetry SHALL NOT collect raw source, prompts, secrets, paths, personal data, or individual productivity by default. | Telemetry schema audit. |
| `L7-NFR-025` | Privacy | Retention, access, redaction, tombstoning/deletion, sensitivity, legal-hold, and backup/Git limitations SHALL apply to memory and observability artifacts. | Data-lifecycle fixtures. |
| `L7-NFR-026` | Maintainability | Schemas, semantic skills, reference profiles, host adapters, eval fixtures, and packages SHALL be versioned with explicit compatibility. | Versioning audit. |
| `L7-NFR-027` | Supply chain | Level 7's shipped runtime dependencies SHALL use ecosystem-appropriate locks and integrity/update policy; consuming projects are governed by their applicable profile, not a universal exact-pin rule. | Lock/update tests. |
| `L7-NFR-028` | Maintainability | Changes SHALL be cohesive, bounded, reviewable, and reversible or compensable, with conventional history after Git initialization. | Change review. |
| `L7-NFR-029` | Evolvability | Public behavior/artifacts SHALL have compatibility, deprecation, migration, and sunset policies. | Compatibility tests and documentation review. |
| `L7-NFR-030` | Maintainability | Deterministically produced host files SHALL detect unsupported manual drift. | Regeneration-drift test. |
| `L7-NFR-031` | Performance | Skill-discovery and selected-skill context budgets SHALL be declared per supported host and tested. | Context-budget suite. |
| `L7-NFR-032` | Context quality | Context selection SHALL favor authoritative high-signal inputs over volume and retain source pointers. | Retrieval/context fixtures. |
| `L7-NFR-033` | Cost | Tool calls, subagents, retries, time, tokens, and monetary cost SHALL have policy-owned budgets and observability. | Budget-exhaustion tests. |
| `L7-NFR-034` | Quality tradeoff | Level 7 performance optimization SHALL NOT weaken safety, evidence, privacy, or accessibility invariants. | Regression policy tests. |
| `L7-NFR-035` | Scale | Before release, the declared support matrix SHALL publish repository-size/context/scan-time benchmarks with hardware, cache, host/model, and network conditions. | Performance benchmark report. |

## 11. Product release boundary

“Included” means semantic classification, guided artifact workflow, and tested A0–A2 behavior. It does not imply direct integration with or execution against every database, cloud, CI, observability, or deployment system.

### V1.0 launch-blocking thin slice

- One authorized local repository or worktree scope at a time
- Independently validated Codex CLI and Claude Code distributions
- Non-mutating preflight, permission/provider disclosure, and compatibility report
- Read-only intake and scoped classification across new, existing, legacy-constrained, live, and retiring components
- Change/risk classification, actor/approval assurance, policy precedence, and safe degraded modes
- One provider-neutral conductor that returns one transition/clarification/blocked state
- Common artifact envelope, canonical state, evidence taxonomy, provenance, staleness, and cross-host resumption
- `Baseline → Frame → Approve → Execute → Verify` for an approved local A0–A2 slice
- Generic, feature/behavior-change, and behavior-preserving-refactor proof profiles
- `Deliver → Observe → Learn` planning, handoff, and evidence ingestion without Level 7 external/production execution
- Procedural audit labels, mandatory independent review for `R3`, and stale-candidate invalidation
- Public regression plus protected release-holdout conformance suite
- Clean install, update, rollback, uninstall, licensing, permissions, SBOM, provenance, and release integrity
- No mandatory MCP, hooks, subagents, external memory, telemetry, or Level 7-hosted service

All other §9.6 profiles remain product-family requirements that SHALL apply when shipped, but they are not launch-blocking v1.0 scope. Foundation step 2 must tag every requirement `V1.0`, `V1.x`, or `Later` and preserve the safety prerequisites of any selected slice.

### Candidate v1.x increments, to prioritize in step 2

- Data/schema and database-design/query profiles
- Security/privacy and dependency/supply-chain project profiles
- UX/accessibility and architecture/modernization profiles
- Performance, scaling/capacity, and dead-code/deprecation profiles
- Infrastructure/configuration and public-contract profiles
- Incident and applicable live-operations profiles
- Richer local relationship indexes, organization policy packs, and integrations

### Explicit non-goals

- Replacing accountable engineers, product/design, security, QA/SRE, data, legal, or domain experts
- Certifying correctness, security, accessibility, privacy, performance, legal, or regulatory compliance
- Any Level 7-executed external, deployment, exposure, production, destructive, or irreversible mutation in v1.0
- Background self-healing or self-modifying production operation
- Automatic full legacy rewrite, guaranteed dead-code discovery, or guaranteed optimization
- Universal support for every coding agent, language, stack, CI provider, deployment platform, or jurisdiction
- Multi-repository orchestration or a mandatory Level 7 SaaS control plane
- Hidden remote memory, credential brokerage, or default outbound telemetry
- One universal quality score, coverage threshold, process recipe, geometry system, rollout percentage, or productivity ranking
- Replacing Git, CI, issue trackers, observability, design tools, or incident-management systems

## 12. Initial product and release metrics

Targets below are initial release hypotheses. Definitions and run protocols SHALL be frozen under `L7-EVAL-009` before tuning. Baseline review may refine non-safety targets but SHALL NOT weaken a safety invariant; no aggregate score may compensate for one.

### North-star outcome

An **initiated change** is a scoped change record accepted by the accountable owner for evaluation. The north-star is the proportion that reaches a product outcome decision from §6.4 with outcome evidence, complete decision/evidence lineage, proportionate ceremony, and no unauthorized action—considered alongside downstream rework, rollback, and user value so rapid rejection/deferment cannot game it.

### Release-blocking conformance targets

| Area | Initial v1 target |
|---|---|
| Clean compatibility | 100% install, discovery, invocation, update, rollback, and removal success on every declared host/OS/version combination. |
| Routing | At least 80 balanced, human-labeled scenarios; at least 95% correct primary routing decision across the frozen repeated-run protocol on both hosts. |
| High-risk routing | 100% required conservative routing/gate decision on at least 20 authorization, PII, migration, deletion, production, and prompt-injection cases; false-low-risk rate is separately release-blocking. |
| Semantic portability | 100% agreement on safety decision, lifecycle transition, and artifact schema; at least 90% agreement on substantive workflow outcome. |
| Unauthorized effects | Zero unauthorized code, Git, network, credential, deployment, destructive, or external mutations in at least 30 adversarial cases. |
| Evidence integrity | 100% of gate/release claims bound to scope, source identity, and traceable evidence; evidence is repeatable where possible and records limitations otherwise; zero fabricated command/test claims. |
| Stale evidence | 100% rejection of approvals and audit results after a material candidate change. |
| Resumption | 100% correct state reconstruction after interruption at every lifecycle boundary on both hosts. |
| Degraded modes | 100% correct gap/`UNVERIFIED` handling for fixtures lacking Git, tests, CI, telemetry, flags, network, or subagents. |
| Artifact validity | 100% schema-valid published artifacts with no dangling required references or seeded secrets. |
| Memory/injection safety | Zero execution of malicious instructions embedded in repository artifacts and zero seeded-secret persistence/transmission in the release suite. |
| Audit independence | Zero cases where the author or remediator can issue independent GO for its own candidate. |
| Seeded blockers | Zero false GO results; track false-positive blocks separately with an initial target no greater than 15%. |
| Knowledge hygiene | 100% of shipped reference profiles carry source/version/status/license/freshness metadata. |
| Existing-file safety | Zero silent overwrites of user instructions, settings, configuration, skills, hooks, or dirty-worktree changes. |

### Pilot outcome targets

- At least 90% of supported real-user clean installs pass first-run validation without manual file surgery; controlled declared host/OS configurations must still meet the 100% conformance gate above.
- Median time from install to evidence-backed phase/next-action diagnosis is no more than five minutes on a published fixture-size, hardware/cache, host/model, and network profile, excluding user decision time.
- At least 80% of at least 12 representative pilot users complete install, phase detection, interruption/resume, and one workflow without maintainer intervention. This is a formative usability threshold, not market-wide efficacy evidence.
- At least 80% can answer “what phase are we in, why, what is blocked, and who decides next?” from the first status view.
- At least 80% judge the requested ceremony proportionate to risk.
- At least 95% of sampled material claims link to evidence or carry an explicit unverified/user-asserted label.
- Cross-host interrupted workflows resume at the correct incomplete gate in at least 95% of real-user pilot attempts; controlled boundary fixtures remain at 100%.
- Median user decision interruptions from diagnosis to verified local diff for an eligible fast-path change are no more than two, excluding substantive product ambiguity.
- At least 60% of eligible pilot users complete a second Level 7 transition or change within 14 days.

### Outcome measurement guardrails

- Baseline before claiming improvement.
- Measure verification, review, rework, rollback, and intervention cost alongside generation speed.
- Compare request-to-approved-testable-plan time, missing acceptance/recovery criteria, human correction count, verification/review effort, rework, rollback, and escaped-change rate. Speed counts as improvement only when safety and quality are non-inferior.
- Track delivery throughput and instability together, per comparable product/service and over time.
- Treat correlations and self-report as signals rather than proof of causality.
- Never rank individual developers with Level 7 metrics.

## 13. Constraints and assumptions

### Confirmed constraints

- The target hosts are Codex CLI and Claude Code, each with distinct packaging and capability contracts.
- Project artifacts live in the consuming repository and remain readable without Level 7.
- Level 7-owned validators and artifacts must remain useful without a Level 7 service; model/host processing and optional web/connectors are separate disclosed network boundaries and may be unavailable.
- `R3` work requires structurally independent qualified review and accountable human/domain approval.
- User-visible live behavior requires an explicit release/reversal strategy; feature flags are one mechanism, not a universal one.
- Production behavior defaults to non-exposure until explicitly authorized.

### Working assumptions requiring later validation

- The accountable technical owner is the primary v1 user.
- English-language text workflows are sufficient for the initial release.
- One local repository/worktree at a time is an acceptable v1 scope.
- A human-readable Markdown layer plus machine-readable metadata can satisfy early artifact UX and validation needs.
- The initial product can remain skills-first with no mandatory MCP server or hosted UI.
- Initial operating-system support will be selected after the technology and compatibility evaluation.

## 14. Risk register

| ID | Risk | Impact | Mitigation requirement |
|---|---|---|---|
| `R-001` | Scope expands into an encyclopedia of all engineering knowledge. | Context bloat, contradictions, unusable ceremony. | Conditional assurance profiles, a thin v1.0 slice, and versioned references. |
| `R-002` | Generated artifacts create false assurance. | Unsafe approval or release. | Schema validity, claim/evidence typing, stale-evidence invalidation, assurance cases. |
| `R-003` | Shared files are mistaken for host portability. | One host silently routes or authorizes differently. | Separate adapters and differential semantic evals. |
| `R-004` | Prompt-only rules are described as enforcement. | Guardrail bypass and misleading marketing. | Enforcement-locus registry and deterministic/host controls. |
| `R-005` | Repository or memory content injects instructions. | Exfiltration, tool misuse, authority hijack. | Trust labels, bounded context selection, permission controls, adversarial suite. |
| `R-006` | Stale context or artifacts drive action. | Wrong target, duplicate work, invalid GO. | Revision/digest binding, freshness, leases/precondition rechecks, supersession. |
| `R-007` | Risk classifier underestimates high-impact work. | Inadequate evidence or review. | Highest-dimension rule, false-low-risk metrics, reclassification triggers. |
| `R-008` | Process weight drives abandonment or blind bypass. | Low adoption or approval fatigue. | Risk-proportionate paths, concise decision views, measured ceremony. |
| `R-009` | Model self-review is mistaken for independence. | Correlated blind spots and self-certification. | Explicit review taxonomy, separate identity/context, human/domain approval. |
| `R-010` | Rich memory/telemetry stores sensitive data. | Privacy, security, and retention harm. | No Level 7 telemetry by default, minimization, redaction, access, TTL, deletion/tombstone limits. |
| `R-011` | Multi-agent work creates races or consensus error. | Conflicting writes, cost, false confidence. | Optional delegation, one writer, isolation, budgets, independent integration. |
| `R-012` | Model or host upgrades change routing behavior. | Silent regression and unreproducible releases. | Compatibility matrix, pinned eval environment, repeated differential tests. |
| `R-013` | Visible evals are gamed by prompt/skill improvement. | Benchmark gains without real safety. | Held-out corpora, forbidden-effect tests, independent promotion. |
| `R-014` | Standards drift, conflict, or licensing is mishandled. | Outdated policy, false compliance, redistribution risk. | Reference metadata, freshness review, authority/status labels, license controls. |
| `R-015` | Feature flags accumulate or fail as rollback. | Dead configuration and false recoverability. | Flag lifecycle contract and alternative reversal profiles. |
| `R-016` | “Self-healing” is overclaimed or prematurely automated. | Production instability or unauthorized mutation. | V1 autonomy ceiling and action-level future contracts. |
| `R-017` | Metrics incentivize speed or surveillance. | Quality regression, gaming, organizational harm. | Balanced system metrics, privacy guardrails, no individual ranking. |
| `R-018` | Plugin supply chain is compromised. | Every managed project may be exposed. | Signed/provenance-verified releases, pinned dependencies, SBOM, secure update path. |
| `R-019` | The plugin cannot recover from its own schema evolution. | Lost approvals/history or lock-in. | Versioned migrations, round-trip unknown fields, rollback, readable artifacts. |
| `R-020` | Claims exceed demonstrated support. | Trust and adoption failure. | Evidence-qualified language and release claims limited to tested matrix. |
| `R-021` | Editable repository state forges or revives an approval. | Unauthorized mutation or false GO. | Approval assurance levels, current-action confirmation, expiry/revocation, external attestation where required. |
| `R-022` | A candidate inspects or changes its held-out evaluator. | Reward hacking and false release confidence. | Candidate-isolated protected holdout and independent evaluator authority. |

## 15. Research and standards basis

The later knowledge registry SHALL preserve exact versions, status, licenses, and applicability. This initial requirements document uses the following primary or authoritative anchors:

| Domain | Source | Requirement implication |
|---|---|---|
| Plugin and skill contracts | [OpenAI plugins](https://developers.openai.com/plugins/build/plugins), [OpenAI skills](https://developers.openai.com/plugins/build/skills), [Claude plugin reference](https://code.claude.com/docs/en/plugins-reference), [Claude skills](https://code.claude.com/docs/en/skills) | Semantic core with separate adapters, concise skill discovery, explicit capabilities. |
| Prompting and evaluation | [OpenAI model guidance](https://developers.openai.com/api/docs/guides/latest-model), [OpenAI evaluation guidance](https://developers.openai.com/api/docs/guides/evaluation-best-practices), [Anthropic prompting](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices), [Anthropic evaluation](https://platform.claude.com/docs/en/test-and-evaluate/develop-tests) | Lean semantic prompt contracts and provider-neutral, representative evals. |
| Software lifecycle | [ISO/IEC/IEEE 14764:2022](https://www.iso.org/standard/80710.html) | Refactoring is not the umbrella for maintenance, operation, migration, and retirement. |
| Product quality and architecture | [ISO/IEC 25010:2023](https://www.iso.org/standard/78176.html), [ISO/IEC/IEEE 42010:2022](https://www.iso.org/standard/74393.html) | Select relevant quality attributes and stakeholder concerns rather than one checklist. |
| Assurance and provenance | [ISO/IEC 15026-2](https://www.iso.org/standard/52926.html), [W3C PROV-DM](https://www.w3.org/TR/prov-dm/) | Risk-tiered claim/argument/evidence and traceable artifact lineage. |
| Delivery and reliability | [DORA metrics](https://dora.dev/guides/dora-metrics/), [SRE error-budget policy](https://sre.google/workbook/error-budget-policy/), [SRE canarying](https://sre.google/workbook/canarying-releases/), [SRE postmortems](https://sre.google/workbook/postmortem-culture/) | Balanced system outcomes, progressive exposure, rollback, and operational learning. |
| Secure development | [NIST SSDF](https://csrc.nist.gov/pubs/sp/800/218/final), [NIST SP 800-218A](https://csrc.nist.gov/pubs/sp/800/218/a/final), [NIST AI RMF](https://www.nist.gov/itl/ai-risk-management-framework) | Risk-based secure lifecycle and AI-specific governance. |
| Agent security and supply chain | [OWASP AI Agent Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/AI_Agent_Security_Cheat_Sheet.html), [SLSA 1.2](https://slsa.dev/spec/v1.2/) | Least authority, injection defenses, provenance, and hardened release path. |
| Accessibility and privacy | [WCAG 2.2](https://www.w3.org/TR/WCAG22/), [NIST Privacy Framework](https://www.nist.gov/privacy-framework) | Testable accessibility baseline and governed data lifecycle. |
| API/data evolution | [RFC 9745 — Deprecation](https://www.rfc-editor.org/rfc/rfc9745.html), [Evolutionary Database Design](https://www.martinfowler.com/articles/evodb.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/18/using-explain.html) | Separate deprecation from removal; version migrations; profile before query optimization. |
| Adaptive systems | [IBM autonomic computing blueprint](https://www.redbooks.ibm.com/redbooks/pdfs/sg246665.pdf), [Kubernetes Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), [Perpetual assurances](https://arxiv.org/abs/1903.04771) | Future bounded control loops need governance, verification, freshness, and rollback. |
| Real-world AI measurement | [DORA 2025 AI research](https://dora.dev/research/2025/dora-report/), [METR early-2025 developer study](https://metr.org/blog/2025-07-10-early-2025-ai-experienced-os-dev-study/) | Measure outcomes and verification cost rather than assuming or self-reporting acceleration. |

## 16. Requirements acceptance gate

This artifact is ready for owner review when:

- The product category, promise, accountable authority, and non-goals are unambiguous.
- Primary and secondary users and core jobs are identifiable without reading the skills.
- Greenfield, existing, legacy-constrained, live, and retiring entry conditions are covered.
- The risk dimensions, levels, unknown-risk behavior, approval assurance, policy precedence, and waiver rules are explicit.
- Functional and nonfunctional requirements are testable and internally consistent.
- Codex/Claude portability is behavioral and artifact-based rather than filename-based.
- Human authority, plugin authority, and prohibited autonomous actions are explicit.
- V1.0 is limited to A0–A2 execution; A3/A4 are planning/handoff only.
- Public regression and protected holdout evaluator boundaries are separated.
- Deployment and exposure are separate optional transitions inside `Deliver`.
- Prototype policies have explicit retain/replace dispositions without pretending the active files are already changed.
- Metrics contain safety invariants and privacy/anti-gaming constraints.
- Risks and working assumptions are visible.

Approval of this document authorizes **foundation step 2: feature backlog only**. It does not authorize architecture selection, harness work, product feature implementation, deployment, or release.

## 17. Decision record

| Date | Decision | Actor | Status |
|---|---|---|---|
| 2026-08-24 | Treat the current package as prototype evidence rather than a completed v1 contract. | Product owner | Approved |
| 2026-08-24 | Define Level 7 as a human-governed software/product evolution and assurance system. | Product owner | Approved |
| 2026-08-24 | Use accountable technical owner as primary user/approver and implementing engineer/maintainer as the core operator with separately granted authority. | Product owner | Approved |
| 2026-08-24 | Cap v1.0 execution at A0–A2; A3/A4 are planning and handoff only; exclude A5. | Product owner | Approved |
| 2026-08-24 | Build one semantic product with separate Codex and Claude distributions. | Product owner | Approved |
| 2026-08-24 | Use `Baseline → Frame → Approve → Execute → Verify → Deliver → Observe → Learn`, with separately gated Package/Deploy/Expose transitions. | Product owner | Approved |
| 2026-08-24 | Require max-dimension risk aggregation, structurally independent `R3` review, and approval assurance beyond editable repository text. | Product owner | Approved |
| 2026-08-24 | Separate public regression fixtures from a candidate-inaccessible protected release holdout. | Product owner | Approved |
| 2026-08-24 | Limit v1.0 to a thin A0–A2 end-to-end slice; prioritize additional proof profiles as v1.x increments. | Product owner | Approved |
| 2026-08-24 | Approve requirements v0.2.0 and authorize feature-backlog drafting only. | Product owner | Approved (`AP1` at confirmation time) |
