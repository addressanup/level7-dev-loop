# Level 7 Dev Loop — Feature Backlog

| Field | Value |
|---|---|
| Artifact ID | `L7-BKL-001` |
| Artifact type | Feature backlog |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Foundation step | 2 — Feature backlog |
| Status | Approved for foundation step 3 |
| Version | 0.1.0 |
| Date | 2026-08-24 |
| Input | Approved [`L7-REQ-001`](requirements.md), version 0.2.0 |
| Product | Level 7 Dev Loop |
| Initial hosts | Codex CLI and Claude Code |
| Scope identity | Non-Git workspace snapshot observed on 2026-08-24; no commit identity available |
| Effect and risk | A1 artifact-only write; foundation-planning risk |
| Approval state | Owner approved in current conversation on 2026-08-24 (`AP1` at confirmation time); persisted record is `AP0` until revalidated |
| Sensitivity | Internal product planning |
| Next authorization if approved | Foundation step 3 — architecture only |

**Approval decision:** the product owner accepted the priority, dependency, effort, and acceptance boundaries and authorized foundation step 3—architecture drafting—only. Approval does not authorize technology selection, harness construction, skill or prompt edits, manifest changes, implementation, installation, deployment, exposure, or release.

## 1. Backlog outcome

The v1.0 critical path is a **governed end-to-end product slice**, not a rewrite of the 12 prototype skills. It first establishes the semantic, evidence, evaluation, and authority contracts that make a workflow trustworthy; then proves one complete local A0–A2 change journey; then renders and tests separate Codex and Claude distributions; finally it admits a release only through conformance, protected evaluation, pilot evidence, and independent review.

The minimum lovable v1.0 story is:

> On either supported host, an accountable technical owner can install Level 7 without file surgery, ask what to do next in a real local repository, understand the evidence-backed state and authority boundary, approve one proportionate feature or behavior-preserving refactor, receive a verified and resumable local diff, and hand off a release/observation decision with no fabricated evidence or unauthorized effect.

This backlog deliberately does **not** choose an implementation language, schema format, test framework, packaging mechanism, graph store, model, hosted service, or UI. Those are architecture and technology decisions for later approved foundation steps.

## 2. Prioritization and estimation contract

### 2.1 Priority

| Priority | Meaning | Release treatment |
|---|---|---|
| `P0` | Required for the v1.0 product promise or its safety/release boundary. | A missing or failing P0 item blocks v1.0. Safety-critical failures cannot be waived by an aggregate score. |
| `P1` | A v1.x product-family increment after the v1.0 slice proves usable and safe. | Prioritized with pilot evidence; not smuggled into v1.0 through a broad skill. |
| `P2` | Later candidate needing discovery, new authority, or a separate autonomy/product charter. | No delivery commitment. Promotion requires explicit owner approval and re-risking. |

Priority is a mutable field; stable backlog IDs do not encode it. An item may move priority only with a recorded rationale, requirement-impact check, and approval. A P0 safety prerequisite cannot be demoted merely to meet a date.

### 2.2 Relative effort

| Points | Planning interpretation |
|---|---|
| `1` | At most one focused engineering day after prerequisites exist. |
| `2` | Roughly two to three focused engineering days. |
| `3` | Roughly four to six focused engineering days. |
| `5` | Roughly one to two focused engineering weeks. |
| `8` | Roughly two to four focused engineering weeks. |
| `13` | Too large or uncertain for implementation commitment; split during architecture/planning. |

Points are coarse, pre-architecture estimates for relative sequencing, not dates or staffing promises. They include implementation, fixtures, documentation, and verification. Confidence is `H` (bounded and familiar), `M` (material architecture uncertainty), or `L` (external/empirical uncertainty). Parallel branches reduce elapsed time only when they do not weaken independent review or create shared-write races.

### 2.3 Common definition of done

Every implemented backlog item must address each dimension below. A genuinely irrelevant dimension carries a named `NOT_APPLICABLE` reason or points to the supporting accountable item; “as applicable” cannot silently erase an obligation.

- trace to its accountable requirements and preserve the approved scope/non-goals;
- declare outcome, effect class, risk, authority, prohibited effects, recovery, stopping, and escalation behavior;
- produce human-readable and machine-validatable evidence bound to the exact candidate/source identity;
- include positive, negative, boundary, degraded-capability, interruption, and adversarial fixtures rather than only a happy-path demonstration;
- work through the provider-neutral semantics; host-involved behavior receives both provisional-adapter smoke during the walking slices and final actual-host conformance before a support claim;
- preserve unrelated user work and never convert missing capability or evidence into a pass;
- document context, cost, latency, privacy, accessibility, compatibility, and support limits that apply;
- leave the harness green and add no failing forbidden-effect, fabricated-evidence, stale-approval, or seeded-secret case;
- update the relevant `docs/artifacts/` state and receive review proportional to risk.

An item is not done because files exist, a model says it is done, or a single demonstration succeeds.

## 3. Dependency strategy and release checkpoints

The critical dependency graph is:

```text
BL-001 Scope/support freeze
  ├── BL-002 Semantic contract ─┬── BL-040 Dual-host walking skeleton ─┐
  └── BL-003 Evaluation protocol┘                                      │
             └── BL-004 Artifact/state core                            │
                              │                                        │
                  BL-005 Safety kernel ─────────────────────────────────┘
                              │
                  BL-006 Doctor/intake/context
                               │
                  BL-007 Conductor and cutover
                               │
                  BL-008 Frame and Approve
                               │
                  BL-009 Execute and Verify
                               │
                  BL-010 Candidate assurance
                               │
                  BL-011 Deliver/Observe/Learn handoff
                               │
                  ┌────────────┴────────────┐
                  v                         v
             BL-012 Codex              BL-013 Claude
                  └────────────┬────────────┘
                               v
                  BL-014 Distribution lifecycle
                               │
                  BL-015 Controlled conformance
                               │
                  BL-041 Pilot/adoption validation
                               │
                  BL-042 Independent release decision
```

The graph shows **completion dependencies**, not a mandate to finish horizontal layers before demonstrating value. `BL-003` freezes labels, oracles, run rules, evaluator authority, and the holdout boundary before candidate prompt/skill tuning; it seeds only the tests possible at that point. Every subsequent feature owns the public fixtures for its behavior, and `BL-015` operates the complete suite and protected holdout against the frozen candidate.

`BL-040` is an actual-host, read-only walking skeleton used to expose Codex/Claude discovery, rendering, context, and permission constraints early. Each C0–C3 slice extends those provisional adapters and runs development smoke tests, but smoke evidence cannot satisfy final host support or conformance. `BL-012` and `BL-013` are completed only after the whole local lifecycle exists; they harden each adapter and prove the actual-host journey. Supply-chain and uninstall work are part of the product, not final polish.

| Checkpoint | Completed items | Cumulative points | User-visible capability | Promotion condition |
|---|---|---:|---|---|
| C−1 — Host walking skeleton | `BL-001`–`003`, `040` | 21 | Actual Codex and Claude development installs expose one read-only entry point; no support claim. | Host constraints recorded; forbidden-effect smoke passes. |
| C0 — Governed observer | `BL-001`–`006`, `040` | 55 | Safe preflight, scoped evidence, explicit uncertainty; no project mutation. | Semantic/eval/state/safety integration reviews pass. |
| C1 — Governed planner | through `BL-008` | 68 | One next transition and an approval-bound change contract. | No specialist or direct invocation bypass. |
| C2 — Local changer | through `BL-009` | 76 | One approved feature or refactor can produce a verified local diff. | A0–A2 boundary and negative suites pass. |
| C3 — Governed closer | through `BL-011` | 86 | Candidate assurance and non-executing delivery/observation handoff. | Stale/independence/release-language tests pass. |
| C4 — Dual-host beta | through `BL-014` | 109 | Installable, reversible Codex and Claude packages. | Both declared matrices pass independently. |
| C5 — Controlled candidate | through `BL-015` | 117 | The actual packages pass public, protected, and differential conformance. | Every controlled release blocker passes. |
| C6 — Adoption candidate | through `BL-041` | 125 | Staged and release-candidate pilots support the minimum lovable story. | Pilot thresholds and disclosed denominators pass. |
| C7 — v1.0 | through `BL-042` | 128 | Evidence-qualified stable release. | Exact candidate receives an independent digest-bound `GO`. |

## 4. P0 — v1.0 launch backlog

### 4.1 Summary

| ID | Product outcome | Depends on | Effort | Confidence |
|---|---|---:|---:|:---:|
| `L7-BL-001` | Freeze the supported v1 boundary and requirement ownership. | — | 3 | H |
| `L7-BL-002` | Define one provider-neutral semantic and knowledge contract. | 001 | 8 | M |
| `L7-BL-003` | Freeze a non-gameable evaluation protocol and holdout boundary. | 001 | 5 | L |
| `L7-BL-040` | Prove an early read-only walking skeleton on both actual hosts. | 002, 003 | 5 | M |
| `L7-BL-004` | Preserve canonical, valid, private, resumable artifact state. | 002, 003 | 13 | L |
| `L7-BL-005` | Enforce risk, authority, policy, and prohibited-effect boundaries. | 002, 003, 004 | 13 | L |
| `L7-BL-006` | Deliver a safe first five minutes: doctor, intake, classification, context. | 004, 005, 040 | 8 | M |
| `L7-BL-007` | Route through one conductor and close prototype bypasses. | 004, 005, 006 | 8 | M |
| `L7-BL-008` | Frame and approve one bounded, proportionate change contract. | 007 | 5 | M |
| `L7-BL-009` | Execute and verify the local generic/feature/refactor slice. | 008 | 8 | M |
| `L7-BL-010` | Produce digest-bound audit and independent assurance decisions. | 009 | 5 | M |
| `L7-BL-011` | Hand off Package/Deploy/Expose, Observe, and Learn safely. | 010 | 5 | M |
| `L7-BL-012` | Complete and validate the Codex adapter. | 011, 040 | 5 | M |
| `L7-BL-013` | Complete and validate the Claude adapter. | 011, 040 | 5 | M |
| `L7-BL-014` | Ship reversible distributions with supply-chain evidence. | 011, 012, 013 | 13 | L |
| `L7-BL-015` | Pass controlled differential conformance and protected evaluation. | 003, 011, 014 | 8 | L |
| `L7-BL-041` | Validate first-run, workflow, resumption, and return-use adoption. | 008, 009, 014, 015 | 8 | L |
| `L7-BL-042` | Make the independent v1.0 release decision. | 015, 041 | 3 | M |

P0 totals **128 coarse points plus protected-evaluator and pilot calendar dependencies**. That is a scope warning, not a delivery forecast: the v1.0 profile breadth is thin, but the dual-host safety/evidence product is not a small build. `L7-BL-004`, `005`, and `014` retain indivisible integration gates but must be decomposed after architecture. The controlled-evaluation, pilot, and release-decision outcomes are already separated because they have different owners, evidence, and clocks.

### `L7-BL-001` — Scope, support, and traceability freeze

**Outcome:** contributors can tell what v1.0 supports, what it only plans, what is deferred, and which backlog item is accountable for every normative requirement.

**Owns:** backlog governance and the release-boundary allocation in §8; no separate normative requirement family.

**Acceptance criteria:**

1. Every one of the 163 normative requirements in `L7-REQ-001` has exactly one accountable backlog owner and a `V1.0`, `V1.x`, or `Later` allocation; supporting relationships may be many-to-many but do not obscure ownership. A source-derived traceability check recalculates the count and fails on missing, duplicate, unknown, or malformed IDs so the hard-coded summary cannot silently drift.
2. The v1 support matrix says one local repository/worktree, Codex CLI and Claude Code, A0–A2 execution, and A3/A4 plan/handoff only; A5 and background/self-modifying behavior are excluded.
3. Generic, feature/behavior-change, and behavior-preserving-refactor are the only launch proof profiles. If another specialist profile is materially required but unavailable, the result is `BLOCKED` or `NOT_EVALUATED`, never `NOT_APPLICABLE` or a generic pass.
4. Each of the 12 current user-invocable prototype skills has an explicit conform, replace, deprecate, or exclude disposition before any v1 release; none is grandfathered into the release claim.
5. The existing `1.0.0`, dual-compatibility, and enforcement claims are withheld from stable promotion until `L7-BL-042` passes.
6. Scope or priority changes produce an impact diff and accountable approval; dates or aggregate metrics cannot waive a safety prerequisite.

### `L7-BL-002` — Provider-neutral semantic contract and knowledge spine

**Outcome:** prompts, skills, artifacts, lifecycle transitions, and hosts share one versioned meaning without forcing identical syntax.

**Owns requirements:** `L7-FLOW-001`–`006`, `008`–`010`; `L7-SKILL-001`–`002`; `L7-PROMPT-001`–`002`; `L7-AGENT-001`–`003`; `L7-HOST-001`, `005`; `L7-KNOW-001`–`004`; `L7-NFR-033`.

**Acceptance criteria:**

1. The canonical taxonomies for evidence, gate, release verdict, product decision, heritage/operational state, risk, approval assurance, effect class, lifecycle, and change class have versioned, machine-testable allowed values and definitions.
2. Invalid states and combinations—such as `GO` with an unresolved release blocker, `PASS` based only on `UNVERIFIED`, or A5 in v1—are rejected deterministically.
3. The semantic skill contract covers positive/negative triggers, prerequisites, inputs, authority/effect, outputs, transition, success/failure/stopping, host support, references, and fixtures; descriptions have measurable discovery budgets.
4. The prompt contract carries goal, authoritative inputs, invariants, prohibited effects, acceptance criteria, tool/authority bounds, evidence, retry/stopping/escalation, and output shape without hidden duplicated policy.
5. Optional delegation has explicit scope, authority, budget, evidence, verifier, and one-writer integration rules, while the same correctness is possible without a subagent. Policy-owned budgets cover tool calls, subagents, retries, wall time, tokens, and monetary cost; exhaustion stops or escalates rather than weakening a gate.
6. Every shipped reference entry records source, type/authority, version/status, applicability/contraindications, jurisdiction, license, freshness, and review date; draft, disputed, superseded, and restricted material is labeled rather than silently made normative.
7. Semantic-contract fixtures fail when a host renderer drops, weakens, or invents a safety-critical obligation.

### `L7-BL-003` — Evaluation protocol and protected-holdout trust boundary

**Outcome:** improvement is measured against frozen, representative rules that a candidate cannot inspect, rewrite, or relax.

**Owns requirements:** `L7-EVAL-001`, `003`–`004`, `006`, `008`–`009`. Release-time coverage, repeated-trial operation, and the protected corpus are owned by `L7-BL-015`.

**Acceptance criteria:**

1. Before prompt, skill, routing, or grader tuning, the public protocol freezes scenario-selection rules, truth-label schema, supported host/model version policy, run count, sampling, adjudication, confidence reporting, cost/latency recording, and failure thresholds.
2. A provider-neutral coverage map assigns routing/negative activation, authority, artifact transitions, degraded modes, interruption/resume, parity, write collisions, install lifecycle, injection, stale evidence, and forbidden effects to the features that create those test subjects. `BL-003` supplies the executable runner contract and initial semantic/safety fixtures; later backlog items must add their fixtures before they can complete.
3. Deterministic graders are used where the answer can be computed; model judges are calibrated for ordering, verbosity, and model-family bias and never serve as independent safety authority.
4. The protocol reserves at least 20% of release evaluation for a protected holdout outside candidate-readable/writable scope, with separate operator/evaluator authority and a documented leakage response. The independent evaluator owns or independently samples hidden cases and labels using a frozen stratification across hosts, risks, negative activation, forbidden effects, stale approval, secret handling, and degraded modes; `L7-BL-015` owns actual corpus operation and proof of the percentage/boundary.
5. Candidate permissions prohibit changes to evaluator code, oracles, truth labels, adjudication, authorization controls, and release thresholds; attempts fail and are recorded.
6. Seeded protocol/reference candidates that route low on high risk, fabricate evidence, leak a secret, accept stale approval, or perform a forbidden effect fail the available initial gates; each downstream feature adds a deliberately broken candidate proving its new grader detects the claimed fault.
7. Run manifests bind candidate revision, prompt/skill versions, model, host/harness, tools, environment, resources, cost, latency, and trial; best-case runs cannot hide consistency or forbidden-action failures.

### `L7-BL-040` — Dual-host read-only walking skeleton

**Outcome:** real Codex and Claude constraints influence the product before state and safety designs harden, without treating an early smoke path as host support.

**Owns:** early integration evidence for the host requirements owned by `L7-BL-002`, `004`, `006`, `012`–`015`; no separate normative requirement family.

**Acceptance criteria:**

1. One development-supported actual version of Codex and one of Claude Code can cleanly discover and invoke a provisional single Level 7 entry point in Level-7-provided synthetic, non-sensitive isolated fixtures; the observed syntax, context, permission, filesystem, and capability constraints become architecture inputs.
2. The walking path is A0 and read-only: it reports a minimal provider-neutral status/evidence payload, makes no project/Git/network/external mutation, and fails visibly if the host cannot preserve a safety-critical field.
3. The same semantic seed fixtures pass through both provisional renderers; differences are recorded rather than normalized away in prose or hidden in adapter state.
4. Negative smoke tests cover over-broad permission requests, undiscoverable/duplicate entry points, context truncation of safety fields, repository-instruction injection, and attempts to infer unsupported invocation parity.
5. Four to five consented formative first-run sessions test discovery, doctor/status comprehension, and permission/provider disclosure using only synthetic non-sensitive repositories and canary-only credentials. No participant repository/credential payload is persisted or reused; real-project sessions cannot begin until `L7-BL-005` and `006` pass. Participant background, host experience, intervention, and failures are recorded.
6. Output and documentation label this a development walking skeleton. It cannot satisfy final install, compatibility, semantic parity, security, or release evidence for `L7-BL-012`–`015`.

### `L7-BL-004` — Canonical artifact, state, evidence, and provenance core

**Outcome:** work resumes safely across interruption and hosts from repository-owned truth rather than chat memory.

**Owns requirements:** `L7-ART-001`–`013`; `L7-HOST-006`; `L7-NFR-006`–`010`, `025`.

**Acceptance criteria:**

1. A common envelope and applicable type schemas validate ID, type, schema version, status, scope, source identity, timestamps, provenance, evidence, approval assurance, sensitivity, freshness, supersession, retention, and review/expiry.
2. Presence does not imply validity: empty, malformed, unrelated, stale, unapproved, conflicting, manually forged, malicious, or superseded artifacts cannot advance state.
3. Evidence distinguishes planned from executed work and records method, source identity, environment, time, result/output, producer, evidence state, and reproducibility limits; no output may claim an unexecuted command.
4. Git identities and non-Git scoped content-digest snapshots both work; material candidate or scope changes invalidate prior verification and approval.
5. A fresh provider-neutral reference process reconstructs every lifecycle boundary from canonical files only, resumes at the last valid incomplete gate, and never repeats approved work or inflates `AP0` records. `L7-BL-012`, `013`, and `015` separately prove actual same-host and cross-host reconstruction.
6. Unknown fields round-trip, migrations preserve readable history and have rollback or an explicit irreversibility plan, reruns are idempotent or expose conflicts, and derived indexes can be deleted and rebuilt from canonical files.
7. Seeded secrets are never persisted; sensitivity, minimization, redaction, tombstoning/deletion, retention, and Git/remote/backup limitations are represented honestly.
8. Provenance records artifacts/entities, producing activity and actor, delegation, derivation, and source links; hidden chain-of-thought is neither requested nor persisted.

### `L7-BL-005` — Authority, risk, policy, and safety kernel

**Outcome:** Level 7 can explain and enforce what may happen, to exactly what target, under whose current authority.

**Owns requirements:** `L7-RISK-001`–`003`; `L7-AUTH-001`–`012`; `L7-POL-001`; `L7-SAFE-001`–`003`; `L7-AUTO-001`; `L7-NFR-001`–`003`, `005`, `022`, `024`.

**Acceptance criteria:**

1. Every work item evaluates every approved risk dimension, conservatively handles unknown critical dimensions, applies the highest material level, and requires independent evidence plus accountable approval for a downgrade.
2. Approval binds action/effect class, exact target/scope, plan or candidate digest, environment, approver/actor assurance, and validity window; editable approval text remains `AP0` until revalidated.
3. Every mutation path passes a non-bypassable enforceable admission boundary that rechecks canonical root/path and symlinks, source identity, dirty/overlapping changes, scope, risk, capability, approval, intended delta, and recovery immediately before mutation; mismatch aborts without silent partial work where technically feasible. Step 3 decides whether equivalent enforcement is centralized, layered, generated, or host-backed.
4. A0–A2 are the only executable v1 effects. Network, credentials, Git publish, external systems, deployment, exposure, destructive/irreversible actions, A3–A5, and unrelated fixes are denied or converted to an explicit handoff under the correct authority.
5. Policy source/version/precedence, applied controls, conflicts, waiver scope/expiry, actor roles, and each guardrail's enforcement locus are visible; high-consequence prompt-only rules are never marketed as enforced.
6. Repository/web/log/dependency/memory/subagent content remains untrusted data through retrieval, summary, delegation, and transformation; injection, goal-hijack, approval-manipulation, exfiltration, and agent-cascade fixtures cannot expand authority. Known secrets and unnecessary sensitive payloads are neither persisted nor included in selected host/model, web, connector, or subagent context. Each supported path discloses its prevention/redaction/denial controls, provider boundary, limitations and false-negative risk, and safe blocked mode; representative direct and transformed secret fixtures test the outcome without prescribing one detector.
7. Offline/local operation introduces no Level-7-owned service, telemetry, or extra repository-content egress; provider/network boundaries are disclosed and a missing required capability blocks safely.
8. The exact risk × gate × approval matrix is executable: R1 local action requires AP1; R2 requires AP1 plus the required verifier and AP2 when policy demands identity; R3 local candidate creation may use policy-permitted AP1, while R3 `PASS`, `GO`, `CONDITIONAL_GO`, or delivery handoff requires at least AP2 plus structurally independent qualified review and accountable human/domain approval; AP3 is mandatory where multiple authorities/separation policy applies; R4 always blocks. Missing assurance is `BLOCKED`/`NOT_EVALUATED`, and AP0/AP1 cannot be promoted by artifact or model text.
9. Artifact-state and authorization integration uses the same exact candidate identity: forged/stale/conflicting artifacts cannot create authority, and a state-transition plus mutation-admission mismatch fails without advancing state or partially mutating the target.
10. Seeded-secret, symlink escape, changed-target/TOCTOU, broad-goal, urgency, prior-approval, and user-fatigue cases produce zero silent authority expansion or unauthorized effect.

### `L7-BL-006` — Safe first five minutes: doctor, intake, classification, and context

**Outcome:** a new user quickly receives an evidence-backed state, uncertainty, blockers, authority status, and exactly one safe next action without project mutation.

**Owns requirements:** `L7-INTAKE-001`–`005`; `L7-CTX-001`–`002`; `L7-HOST-010`; `L7-NFR-032`.

**Acceptance criteria:**

1. Preflight reports host/plugin/schema compatibility, requested permissions, artifact footprint, provider/network boundary, locally discoverable capabilities, degradation, and actionable fixes without probing an external system.
2. Intake remains repository-root-bounded and read-only, resolves symlink/outside-root attempts safely, and inventories Git/dirty state, tests, CI/build, observability/deployment configuration, database tooling, flags, network, and delegation only from authorized local evidence.
3. Heritage and operational state are classified per selected component/change—including mixed monorepos—and uncertainty is explicit rather than coerced into greenfield, brownfield, or release-ready.
4. Pre-existing failures are baselined and never silently blamed on or repaired within the requested scope.
5. Selected context contains no known secret or unnecessary sensitive payload. The supported host path must prevent, redact, deny, or safely block before transmission at an enforceable point and disclose its provider boundary and limitations. Context then retains the smallest relevant, authoritative, fresh, sensitivity-appropriate inputs and source pointers; compaction preserves approval, scope, risk, blockers, and trust labels without reconstructing excluded material.
6. Provider-neutral first-response fixtures and development smoke through both `L7-BL-040` adapters show evidence state, uncertainty, capability gaps, known/user-asserted/unknown approver, and one next action in accessible text and machine-readable form; actual-host support claims remain gated by `L7-BL-012`–`015`.
7. On the published pilot profile, median install-to-diagnosis time is no more than five minutes excluding user decision time; larger/unsupported repositories report bounded scan behavior rather than hanging or silently truncating.

### `L7-BL-007` — Conductor, status experience, and prototype cutover

**Outcome:** users need one obvious entry point; every invocation follows the same state, authority, and routing controls.

**Owns requirements:** `L7-ROUTE-001`–`005`; `L7-SKILL-003`; `L7-NFR-015`–`021`, `031`.

**Acceptance criteria:**

1. The conductor returns exactly one state transition, one material clarification, or one blocked state with deterministic precedence for overlapping intents and safety gates over scope expansion.
2. In the provider-neutral reference suite, at least 80 balanced human-labeled scenarios meet the frozen 95% routing target. At least 20 balanced high-risk authorization, tenancy, PII, migration, deletion, production, injection, and mixed-class cases achieve 100% of the human-approved conservative **risk, route, and minimum-gate disposition**; blanket blocking or blanket R3 classification does not count. Zero false-low-risk remains separately release-blocking, and false-positive blocks are reported against the approved threshold. Provisional adapters run development smoke; `L7-BL-012`, `013`, and `015` own actual-host conformance.
3. New/existing/legacy-constrained/live/retiring, no-artifact brownfield, mixed-stage, missing-capability, scope-change, and negative-activation fixtures route correctly and invalidate incompatible approval when required.
4. Every current user-invocable prototype skill is either brought behind the conductor and common safety kernel or safely deprecated/excluded. Provider-neutral bypass tests plus development smoke through both walking adapters cover direct invocation, aliases, natural language, and host-native syntax; final adapters repeat them on actual hosts before any support claim.
5. The status view leads with phase, evidence, uncertainty, blockers, approver, and one next action; optional detail follows, terminology is explained, color is not required, and unsupported absolutes such as “perfect,” “secure,” or “compliant” fail output validation.
6. At most one branching approval decision is requested at a time; independent factual questions may be grouped after inspection, and every completed transition returns to status automatically.
7. Discovery and selected-skill context budgets are provisionally measured through `L7-BL-040` and finally declared/tested per supported host by `L7-BL-012`–`015`; a broad specialist description or context that achieves its budget by dropping safety semantics fails.

### `L7-BL-008` — Bounded Frame and Approve journey

**Outcome:** the user can turn ambiguous intent into a proportionate, approval-ready contract without prematurely editing the project.

**Owns requirements:** `L7-FLOW-011`–`012`.

**Acceptance criteria:**

1. Every executable item captures desired outcome, scope/non-goals, invariants, assumptions, primary/secondary change class, risk, acceptance criteria, authority, recovery/compensation, and observation as applicable.
2. The low-risk fast path still captures scope, one outcome/acceptance criterion, mutation preview/authority, applicable verification, and evidence summary; non-applicable architecture/SLO/flag/assurance artifacts require a reason rather than empty ceremony.
3. The approval view clearly separates `OBSERVED`, `INFERRED`, `USER_ASSERTED`, and `UNVERIFIED`, names capability gaps and residual risk, and previews the exact bounded delta.
4. Material change to target, scope, desired behavior, change class, risk, candidate, or recovery invalidates incompatible approval and returns to the appropriate earlier gate.
5. Ambiguous high-impact decisions block for one branching decision; low-impact independent facts can be gathered together without a forced one-question-per-round ritual.
6. Frame/Approve fixtures cover both feature and behavior-preserving refactor, dirty and non-Git workspaces, a genuinely small fast path, and a deceptively small high-risk request.
7. Eligible fast-path conversation fixtures reach a verified local diff with no more than two accountable-owner decision interruptions, excluding genuine product ambiguity; `L7-BL-041` confirms the real-user median rather than discovering the criterion for the first time.

### `L7-BL-009` — Local Execute and Verify thin slice

**Outcome:** an approved local feature or behavior-preserving refactor becomes a bounded diff with truthful change-specific proof.

**Owns requirements:** `L7-PROOF-000`, `001`, `012`.

**Acceptance criteria:**

1. Execution occurs only after the enforceable pre-mutation admission check/boundary defined by `L7-BL-005`, touches only approved scope, preserves unrelated changes, respects retries/cost/time bounds, and stops on material mismatch or risk escalation.
2. The generic profile records outcome, scope/non-goals, invariants, baseline, acceptance criteria, regression evidence, authority, recovery/compensation, observed result, and residual risk.
3. The feature profile proves supported-contract delta, positive and negative acceptance, compatibility, applicable unit/integration/end-to-end evidence, recovery, and observation.
4. The refactor profile first characterizes supported behavior/contracts, keeps structural and behavioral changes separate, and demonstrates invariant preservation; a behavior change cannot be relabeled as refactoring to reduce gates.
5. Missing tools or evidence remain `UNVERIFIED`; missing materially required data/security/UX/performance/operations or other specialist proof causes `BLOCKED`/`NOT_EVALUATED`, never a generic pass or unsupported `NOT_APPLICABLE`.
6. Provider-neutral render fixtures and development smoke through the evolving walking adapters preserve risk, authority, prohibited effects, lifecycle state, artifact validity, and gate result while allowing host-specific wording/tool syntax. Only `L7-BL-012`, `013`, and `015` may satisfy the actual-host parity claim.
7. Positive and adversarial journeys prove no out-of-scope edit, hidden Git/network/external action, fabricated test run, or silent repair of a pre-existing failure.

### `L7-BL-010` — Candidate assurance and independent audit

**Outcome:** a decision owner receives a scoped, digest-bound assurance view whose independence and limitations are explicit.

**Owns requirements:** `L7-AUDIT-001`–`002`; `L7-COMP-001`.

**Acceptance criteria:**

1. The candidate is frozen by exact source identity and artifact digest before audit; any material change makes the verdict stale and requires re-verification/review.
2. Outputs label self-review, separate-context model review, deterministic check, independent human review, and qualified domain review accurately; same-context role-play is self-review.
3. Independent review is read-only, uses separate identity/context where available, cannot author or alter the candidate, and cannot close its own remediation.
4. `R3` requires a claim/argument/evidence assurance case with assumptions, counterevidence/defeaters, residual risk, qualified independent review, and accountable/domain approval. Candidate creation may use policy-permitted AP1, but R3 `PASS`, `GO`, `CONDITIONAL_GO`, or delivery handoff requires at least AP2; AP3 applies when multiple authorities or separation of duties is required. Missing assurance stays `BLOCKED`/`NOT_EVALUATED`.
5. The author/remediator cannot issue independent `GO` for its own candidate, and a green test suite alone cannot erase residual risk or a capability gap.
6. Named frameworks yield scoped, versioned evidence-alignment/gap language, never certification or unsupported “compliant/secure/perfect” claims.

### `L7-BL-011` — Deliver, Observe, Learn, and retirement handoff

**Outcome:** Level 7 closes local work with the next accountable decision while keeping deployment, exposure, observation, and learning distinct.

**Owns requirements:** `L7-REL-001`–`005`; `L7-OPS-003`–`004`.

**Acceptance criteria:**

1. `Package`, `Deploy`, and `Expose` are separate optional gates under `Deliver`; v1 may prepare A3/A4 plans and evidence requirements but cannot execute an external/production action.
2. Handoff selects a context-appropriate reversal or compensation mechanism. If a flag applies, its type, owner, default, targeting/privacy, metrics, failure behavior, expiry, and removal work are explicit; flag disablement is not falsely described as data rollback.
3. Progressive-exposure plans define cohort/control or a justified alternative, attribution, observation window, applicable SLO/product guardrails, abort conditions, and rollback without universal percentages.
4. Release evidence links approved source/configuration through package digest and deployment manifest to the intended subject, while unexecuted steps remain planned/`UNVERIFIED`.
5. Outcome evidence leads to one accountable `SHIP`, `ITERATE`, `DEFER`, `ROLLBACK`, `RETIRE`, or `REJECT` decision; postdeployment evidence may invalidate an earlier `GO`.
6. Incidents and repeated defects create reviewed, owned improvement proposals with lineage; production feedback cannot directly mutate or promote the plugin.
7. Retirement handoff covers consumers, data, dependencies, notice, retention/deletion, recovery, disablement/removal, verification, observation, and closure as applicable.

### `L7-BL-012` — Codex adapter

**Outcome:** Codex users receive the canonical product semantics through a host-valid, minimal-capability adapter.

**Owns:** Codex-specific support for `L7-HOST-002`–`010`; accountable ownership of those requirements remains with `L7-BL-014`/`015` as mapped in §8.

**Acceptance criteria:**

1. A clean declared Codex version discovers and invokes the single entry point using documented native and natural-language syntax; unsupported syntax is not implied.
2. The adapter declares host/version/context/permission/capability limits, requests no broader authority than the selected effect, and safely degrades without optional MCP, hooks, browser, telemetry, memory, or subagents.
3. It reads and writes the canonical artifacts without semantic drift, approval inflation, unknown-field loss, or hidden host-only lifecycle state.
4. Codex-specific prompt/skill rendering passes the public safety, routing, interruption, context-budget, injection, forbidden-effect, and fixture suites.
5. Actual Codex journeys cover Observer through governed closer, including R3/AP blocking, digest-bound audit, stale-verdict invalidation, non-executing delivery handoff, clean-process same-host resume, and a cross-host handoff to Claude using only canonical files.
6. Adapter-specific errors are actionable and do not masquerade as product gate success; `BL-040` smoke evidence is rerun rather than grandfathered.

### `L7-BL-013` — Claude adapter

**Outcome:** Claude Code users receive the same canonical product semantics through a host-valid, minimal-capability adapter.

**Owns:** Claude-specific support for `L7-HOST-002`–`010`; accountable ownership of those requirements remains with `L7-BL-014`/`015` as mapped in §8.

**Acceptance criteria:**

1. A clean declared Claude Code version discovers and invokes the single entry point using documented native and natural-language syntax; unsupported syntax is not implied.
2. The adapter declares host/version/context/permission/capability limits, requests no broader authority than the selected effect, and safely degrades without optional MCP, hooks, browser, telemetry, memory, or subagents.
3. It reads and writes the canonical artifacts without semantic drift, approval inflation, unknown-field loss, or hidden host-only lifecycle state.
4. Claude-specific prompt/skill rendering passes the public safety, routing, interruption, context-budget, injection, forbidden-effect, and fixture suites.
5. Actual Claude journeys cover Observer through governed closer, including R3/AP blocking, digest-bound audit, stale-verdict invalidation, non-executing delivery handoff, clean-process same-host resume, and a cross-host handoff to Codex using only canonical files.
6. Adapter-specific errors are actionable and do not masquerade as product gate success; `BL-040` smoke evidence is rerun rather than grandfathered.

### `L7-BL-014` — Distribution lifecycle, compatibility, and supply chain

**Outcome:** each host package is reproducible, installable, reversible, legally complete, and independently promotable.

**Owns requirements:** `L7-SAFE-004`; `L7-HOST-002`, `004`, `007`–`009`, `011`–`012`; `L7-NFR-004`, `011`–`014`, `026`–`030`.

**Acceptance criteria:**

1. Separate Codex and Claude packages are generated deterministically from one semantic version/source; unsupported manual drift in generated files is detected.
2. The published matrix names exact plugin/host/schema/OS/runtime combinations, required/optional capabilities, safe degradation, migrations, rollback, and support limits.
3. Clean install, discovery, invocation, upgrade, schema migration, package rollback, and uninstall pass on every declared matrix entry; existing instructions/settings/hooks/skills/configuration and dirty work are never silently overwritten.
4. Uninstall previews and removes Level-7-owned active integration/caches/marketplace entries while preserving canonical project artifacts unless their deletion is separately authorized.
5. Every package has authenticated identity through its trusted packaging channel or a verified signing identity. License/notices, least-privilege permission manifest, dependency lock/integrity data, SBOM, provenance/attestation, and update metadata bind to the exact package digest and are verified before install/update. Negative tests reject modified/substituted bytes, mismatched SBOM/provenance, untrusted or revoked identity/key, stale metadata, and an unintended vulnerable downgrade; rotation, revocation, and safe rollback remain testable.
6. Host packages share one semantic version/changelog but promote independently only after their own conformance suite passes; before `L7-BL-042`, any individual-host promotion is explicitly beta/prerelease and cannot make a stable `1.0.0` or dual-host claim.
7. A minimal installation requires no Level 7 control plane, hosted service, MCP server, vector database, native hook, native subagent, external memory, or telemetry.
8. No package or documentation claims v1.0, dual compatibility, enforcement, security, or compliance beyond the exact evidence-qualified tested matrix.

### `L7-BL-015` — Controlled differential conformance and protected evaluation

**Outcome:** the actual release-candidate packages meet the frozen controlled safety, semantic, compatibility, and performance gates before the final pilot is asked to validate adoption; earlier sessions remain formative research.

**Owns requirements:** `L7-EVAL-002`, `005`, `007`; `L7-HOST-003`; `L7-METRIC-001`–`003`; `L7-NFR-034`–`035`.

**Supports/consumes:** the runner, grader, run-manifest, anti-tamper, and frozen-protocol contracts owned by `L7-BL-003`.

**Acceptance criteria:**

1. Every declared host/OS/version entry achieves 100% install/discovery/invocation/update/rollback/removal success and 100% safety-critical agreement on risk, authority, prohibited effects, lifecycle transition, artifact validity, and release verdict; substantive workflow agreement is at least 90%.
2. The frozen routing protocol achieves at least 95% on 80 or more balanced labeled cases. At least 20 balanced high-risk cases achieve 100% of the human-approved conservative risk, route, and minimum-gate disposition; blanket blocking/R3 classification is incorrect, zero false-low-risk remains separately release-blocking, and false-positive blocks are no greater than the approved initial 15% target.
3. At least 30 adversarial cases produce zero unauthorized local/Git/network/credential/deployment/destructive/external effect; evidence, staleness, resumption, degraded mode, artifact validity, injection/secret, independence, knowledge, and existing-file-safety release blockers all meet the exact targets in `L7-REQ-001` §12.
4. The protected holdout is operated independently, remains inaccessible to candidate agents, is at least 20% of the release corpus, and uses independently owned/sampled hidden cases and labels stratified across hosts, risk, negative activation, forbidden effects, stale approval, secrets, and degraded modes. It passes without a candidate-team change to cases, thresholds, oracles, evaluators, or adjudication.
5. Published scale/context/scan-time benchmarks identify repository profile, hardware, cache, host/model, network, and limitations; performance changes do not weaken safety, evidence, privacy, or accessibility.
6. Same-host and cross-host actual-package journeys cover every lifecycle boundary, R3/AP blocking, audit/handoff semantics, degraded modes, interrupted resumption from canonical artifacts only, and both adapter directions; mock or walking-skeleton results cannot substitute.
7. Product metrics balance outcome, evidence, safety, delivery, stability, security/privacy, accessibility, cost, rework, and human effort; no LOC/ticket/PR/prompt metric ranks individuals, and improvement claims require an approved baseline/comparison.
8. The frozen candidate is the actual package bytes, not a source-folder demonstration. Any unmet safety invariant, seeded blocker, holdout boundary, package-authenticity test, or host matrix entry fails controlled conformance regardless of aggregate performance.

### `L7-BL-041` — Staged pilot and adoption validation

**Outcome:** representative users can understand and complete the minimum lovable journey with proportionate ceremony and without maintainer file surgery.

**Owns:** the pilot outcome evidence in `L7-REQ-001` §12; no separate normative requirement family.

**Acceptance criteria:**

1. C−1/C0 formative research includes four to five consented first-run/doctor/status sessions; C1/C2 research includes six to eight framing, approval, fast-path, and local-change sessions. Findings become traced backlog decisions before the release-candidate pilot.
2. Before recruitment, the pilot protocol freezes minimum denominators for each declared host, relevant heritage cohort, contributor/non-contributor status, and both cross-host-resume directions. The final pilot uses the actual controlled-conformance packages with at least 12 representative users and meets those cell floors; undersized cells are `NOT_EVALUATED` and block C6/dual-host release rather than being pooled into an aggregate.
3. Sampling records participant role, contributor status, project/stage, host/model experience, environment, intervention, withdrawal/failure, and denominator. Results are reported overall and by meaningful subgroup without individual ranking.
4. All pilot targets in `L7-REQ-001` §12 are evaluated exactly: clean first run, ≤5-minute diagnosis profile, workflow completion, status comprehension, proportionate ceremony, evidence linkage, cross-host resume, and no more than two median decision interruptions for eligible fast-path work.
5. The eligible return-use cohort is observed for at least 14 days and the ≥60% second-transition/change hypothesis is reported with exclusions and loss to follow-up; an early enthusiastic session cannot substitute.
6. Research is consented and data-minimized, works without Level-7 telemetry, records maintainer intervention/rework/correction cost, and reports negative results rather than changing denominators or thresholds after observation.

### `L7-BL-042` — Independent v1.0 release review and decision

**Outcome:** an accountable owner promotes only the exact candidate whose controlled and pilot evidence supports a stable evidence-qualified release.

**Owns:** the roll-up v1.0 decision; no separate normative requirement family.

**Acceptance criteria:**

1. A frozen release packet binds the actual Codex and Claude package digests to controlled conformance, protected-holdout, pilot, provenance, compatibility, residual-risk, rollback/revocation, and open-condition evidence.
2. A structurally independent reviewer with read-only candidate authority issues a digest-bound `GO`, `CONDITIONAL_GO`, or `NO_GO`; the candidate author/remediator cannot close its own findings or decide promotion.
3. Any unmet safety invariant, seeded blocker, protected-evaluator boundary, package-authenticity condition, pilot release threshold, or declared host matrix entry forces `NO_GO` regardless of aggregate performance.
4. `CONDITIONAL_GO` cannot promote v1.0. Every condition must be evidenced and the unchanged or newly digested exact candidate must receive a fresh independent `GO`.
5. Stable `1.0.0` and dual-host claims require both packages and C7 `GO`; a single-host pass remains beta/prerelease.
6. Release notes accurately state supported scope/matrix, evidence, limitations, residual risk, provider boundary, non-goals, rollback/revocation, and the continued A0–A2 autonomy ceiling.

## 5. P1 — v1.x product-family increments

P1 discovery and design may begin when its technical prerequisites are stable, but implementation promotion requires the C7 v1.0 `GO` from `L7-BL-042`; the dependency column lists additional technical prerequisites. Every new assurance profile must use the semantic skill/profile schema, activate only when applicable, add at least six representative fixtures **per distinct shipped profile** spanning typical/edge/negative/`NOT_APPLICABLE`/blocked behavior, preserve both-host safety semantics, and produce canonical evidence. Profile absence must never be disguised as a pass.

| ID | Increment and outcome | Technical dependencies | Effort | Confidence | Acceptance boundary |
|---|---|---:|---:|:---:|---|
| `L7-BL-016` | Full greenfield foundation-to-build journey | 008, 009, 042 | 8 | L | Requirements → backlog → architecture → technology → harness → orchestration can resume and produce an approved local build slice; no mandatory fixed option count or layer order. |
| `L7-BL-017` | Data/schema and database design/query assurance | 004, 005, 009 | 8 | M | Migration compatibility/rehearsal/reconciliation/recovery and workload/model/key/integrity/isolation/query-plan tradeoffs are evidenced; owns `L7-PROOF-004`, `014`. |
| `L7-BL-018` | Security/privacy and dependency/project-supply-chain assurance | 005, 009, 014 | 13 | L | Must split before build into independently selectable security/privacy and dependency/supply-chain profiles, each with its own qualified review and fixtures; together they own `L7-PROOF-006`, `009`. |
| `L7-BL-019` | UX and accessibility assurance | 008, 009 | 8 | M | User task and all relevant states, keyboard/AT, contrast, zoom/reflow, content, usability and performance receive human-plus-automated evidence; owns `L7-PROOF-007`. |
| `L7-BL-020` | Legacy modernization, dead-code/deprecation, and public-contract retirement | 004, 009, 011 | 13 | L | Must split before build; covers dynamic/static reachability, consumers, data/runtime characterization, migration seams, coexistence, notices/windows/sunset/removal and recovery; owns `L7-PROOF-003`, `005`, `016`. |
| `L7-BL-021` | Performance, resource, scaling, and capacity assurance | 009, 011 | 8 | M | Representative baselines/distributions, correctness thresholds, demand/failure model, saturation/load/resilience, unit cost, degradation and recovery are required; owns `L7-PROOF-002`, `015`. |
| `L7-BL-022` | Infrastructure/configuration, incident, and live-operations assurance | 005, 011 | 13 | L | Must split before build; covers declared emergency authority, desired/observed state, plan/drift/blast radius/secrets, staged application, restoration, SLI/SLO/error-budget limits, postmortem and actions; owns `L7-FLOW-007`, `L7-PROOF-010`, `013`, `L7-OPS-001`–`002`. |
| `L7-BL-023` | Architecture and modernization assurance | 008, 009 | 5 | M | Stakeholders/concerns, quality-attribute scenarios, tradeoffs, fitness functions, failure modes, migration/recovery and independent elevated-risk review are evidenced; owns `L7-PROOF-008`. |
| `L7-BL-024` | Rich local knowledge and evidence-relationship navigation | 004, 015 | 8 | M | Rebuildable derived graphs/indexes expose provenance, supersession, dependencies, freshness and gaps without becoming a second source of truth or requiring a hosted/vector service. |
| `L7-BL-025` | Organization policy packs and attested approval integrations | 005, 014 | 8 | L | Versioned organization policy/precedence and AP2/AP3 attestation are optional, least-privilege, revocable, auditable, and never let repository text self-grant authority. |
| `L7-BL-026` | Optional host accelerators and multi-agent orchestration | 002, 012, 013, 041 | 8 | M | Hooks/MCP/subagents/memory improve a measured outcome but retain a correct single-agent/minimal-host path, disjoint writes, bounded budgets, one integrator, and unchanged authority. |
| `L7-BL-027` | AI prompt, skill, workflow, tool, and evaluator assurance | 003, 005, 015 | 8 | M | Representative/adversarial eval, permissions, injection, cost/latency, regression, independent promotion, rollback and decommission are required; owns `L7-PROOF-011`. |
| `L7-BL-028` | Technical-debt portfolio | 004, 007, 015 | 5 | M | Debt records future cost/interest evidence, outcomes, owner and eliminate/reduce/mitigate/accept decision; smell counts alone do not qualify; owns `L7-DEBT-001`. |
| `L7-BL-029` | Applicable tenancy and collaboration assurance | 005, 009, 023 | 8 | L | Implements the approved prototype disposition only when product evidence requires it; tests tenant isolation, role/permission boundaries, shared-state conflicts, auditability, migration, abuse and recovery rather than imposing a universal story. |

## 6. P2 — later candidates and autonomy horizon

P2 is not a promise. `L7-BL-035` requires explicit authorization to draft a future-autonomy charter with a new threat model, authority model, evaluation plan, ownership, kill switch, and release boundary. `L7-BL-036` and `043` require that charter to be approved; `L7-BL-037`–`039` additionally require their stated predecessors. Dependency on a completed epic is never enough by itself: promotion uses immutable registry lineage for the **same exact action/environment**, and evidence cannot be reused across a different action or environment. No P2 item may retroactively broaden v1 authority.

| ID | Candidate | Technical dependencies | Effort | Confidence | Promotion gate / acceptance boundary |
|---|---|---:|---:|:---:|---|
| `L7-BL-030` | Optional product telemetry | 005, 042 | 8 | L | Explicit opt-in, purpose limitation, minimization, access, retention/deletion and published utility must beat a telemetry-free alternative; owns `L7-NFR-023`. |
| `L7-BL-031` | Read-only external evidence connectors | 005, 014, 042 | 8 | M | Connector-specific authority, provider/egress disclosure, secret handling, provenance, rate/cost bounds, injection defense and offline degradation pass before connection. |
| `L7-BL-032` | Multi-repository/team coordination and dashboard | 004, 005, 024, 042 | 13 | L | New identity, concurrency, privacy, tenancy, consistency, authority and recovery model; repository files remain portable and no individual surveillance/ranking is introduced. |
| `L7-BL-033` | Third-party profile/skill ecosystem | 002, 003, 014, 042 | 13 | L | Signing, provenance, permission review, namespace/version conflicts, quarantine/revocation, compatibility, malicious-package tests and maintainer governance exist before external extensions load. |
| `L7-BL-034` | Additional hosts and localization | 002, 014, 042 | 8 | L | Each host/language has an owned matrix, semantic differential suite, installation lifecycle, context budget, accessible terminology and independent promotion; no “universal agent” claim. |
| `L7-BL-035` | Future autonomy charter and action registry | 005, 027, 042 | 13 | L | Per-action/environment ladder from observe to recommendation, dry-run, approval and narrow preauthorization; policy/evaluator/credentials/audit/kill switch stay outside self-modification; owns `L7-AUTO-002`–`003`. |
| `L7-BL-036` | A3 external non-production action pilot | 025, 031, 035 | 13 | L | Approved charter names exact systems/actions, service owner, sandbox, credentials, idempotency, quotas, compensation, postconditions, observation and immediate revocation. Before credentials/tool invocation, target/environment/window/digest-bound AP2 from the attested environment owner is current; AP3 applies when separation policy requires it, and revoked/stale/mismatched attestation blocks. No production access. |
| `L7-BL-043` | A4 controlled production-action pilot | 022, 025, 035, 036 | 13 | L | Must split by exact action/environment and require applicable specialist assurance, attested approval, live observation/incident controls, blast/recovery evidence, per-action confirmation and kill-switch exercise; it is not autonomous remediation. |
| `L7-BL-037` | A5 bounded autonomous-remediation controller | 022, 025, 035, 036, 043 | 13 | L | Must split by action/environment. Immutable lineage proves that the exact pair passed observe/recommend/dry-run, A3, and controlled-A4 thresholds; failed, stale, expired, or different-pair evidence blocks. Each action also has trigger evidence, preconditions, blast cap, idempotency, retry/cost/time bounds, cooldown/oscillation/circuit breaker, postconditions, recovery, escalation, owner and expiry; owns `L7-AUTO-004`. |
| `L7-BL-038` | Candidate-isolated improvement lab | 003, 014, 027, 035, 042 | 13 | L | Old-task regression, protected adversarial holdout, cost/latency, independent review, signed promotion, canary, rollback and decommission prevent production feedback from self-promoting behavior; owns `L7-AUTO-005`. |
| `L7-BL-039` | Constrained parameter/policy-learning research | 038 | 13 | L | Offline research only until a separately approved proof shows bounded objective, anti-gaming, privacy, interpretability, rollback, no policy/evaluator self-edit, and superiority to reviewed manual proposals. |

## 7. Explicitly excluded from this backlog

The following remain non-goals unless a future requirements revision explicitly adds them:

- legal, regulatory, security, privacy, accessibility, performance, or correctness certification;
- replacement of accountable engineering, product/design, QA/SRE, security, data, legal, or domain experts;
- automatic full legacy rewrites, guaranteed dead-code discovery, or guaranteed optimization;
- a mandatory SaaS control plane, hidden remote memory, credential brokerage, or default outbound telemetry;
- universal coverage of every host, stack, CI/deployment platform, organization, or jurisdiction;
- one universal quality score, recipe, rollout percentage, geometry system, or individual productivity ranking;
- replacement of Git, CI, issue tracking, observability, design, or incident-management systems.

## 8. Normative requirement ownership and release allocation

The table below assigns exactly one accountable backlog owner to each normative requirement from `L7-REQ-001`. Other features may implement or verify the same behavior, but ownership is not duplicated. The total is 163: 140 allocated to v1.0, 18 to v1.x, and 5 to Later.

| Requirement IDs | Accountable backlog owner | Allocation | Count |
|---|---|---:|---:|
| `L7-INTAKE-001`–`005` | `L7-BL-006` | V1.0 | 5 |
| `L7-ROUTE-001`–`005` | `L7-BL-007` | V1.0 | 5 |
| `L7-RISK-001`–`003` | `L7-BL-005` | V1.0 | 3 |
| `L7-FLOW-001`–`006`, `008`–`010` | `L7-BL-002` | V1.0 | 9 |
| `L7-FLOW-007` | `L7-BL-022` | V1.x | 1 |
| `L7-FLOW-011`–`012` | `L7-BL-008` | V1.0 | 2 |
| `L7-ART-001`–`013` | `L7-BL-004` | V1.0 | 13 |
| `L7-AUTH-001`–`012` | `L7-BL-005` | V1.0 | 12 |
| `L7-POL-001` | `L7-BL-005` | V1.0 | 1 |
| `L7-SAFE-001`–`003` | `L7-BL-005` | V1.0 | 3 |
| `L7-SAFE-004` | `L7-BL-014` | V1.0 | 1 |
| `L7-SKILL-001`–`002` | `L7-BL-002` | V1.0 | 2 |
| `L7-SKILL-003` | `L7-BL-007` | V1.0 | 1 |
| `L7-PROMPT-001`–`002` | `L7-BL-002` | V1.0 | 2 |
| `L7-CTX-001`–`002` | `L7-BL-006` | V1.0 | 2 |
| `L7-AGENT-001`–`003` | `L7-BL-002` | V1.0 | 3 |
| `L7-AUDIT-001`–`002` | `L7-BL-010` | V1.0 | 2 |
| `L7-PROOF-000`, `001`, `012` | `L7-BL-009` | V1.0 | 3 |
| `L7-PROOF-002`, `015` | `L7-BL-021` | V1.x | 2 |
| `L7-PROOF-003`, `005`, `016` | `L7-BL-020` | V1.x | 3 |
| `L7-PROOF-004`, `014` | `L7-BL-017` | V1.x | 2 |
| `L7-PROOF-006`, `009` | `L7-BL-018` | V1.x | 2 |
| `L7-PROOF-007` | `L7-BL-019` | V1.x | 1 |
| `L7-PROOF-008` | `L7-BL-023` | V1.x | 1 |
| `L7-PROOF-010`, `013` | `L7-BL-022` | V1.x | 2 |
| `L7-PROOF-011` | `L7-BL-027` | V1.x | 1 |
| `L7-DEBT-001` | `L7-BL-028` | V1.x | 1 |
| `L7-REL-001`–`005` | `L7-BL-011` | V1.0 | 5 |
| `L7-OPS-001`–`002` | `L7-BL-022` | V1.x | 2 |
| `L7-OPS-003`–`004` | `L7-BL-011` | V1.0 | 2 |
| `L7-HOST-001`, `005` | `L7-BL-002` | V1.0 | 2 |
| `L7-HOST-002`, `004`, `007`–`009`, `011`–`012` | `L7-BL-014` | V1.0 | 7 |
| `L7-HOST-003` | `L7-BL-015` | V1.0 | 1 |
| `L7-HOST-006` | `L7-BL-004` | V1.0 | 1 |
| `L7-HOST-010` | `L7-BL-006` | V1.0 | 1 |
| `L7-EVAL-001`, `003`–`004`, `006`, `008`–`009` | `L7-BL-003` | V1.0 | 6 |
| `L7-EVAL-002`, `005`, `007` | `L7-BL-015` | V1.0 | 3 |
| `L7-METRIC-001`–`003` | `L7-BL-015` | V1.0 | 3 |
| `L7-KNOW-001`–`004` | `L7-BL-002` | V1.0 | 4 |
| `L7-COMP-001` | `L7-BL-010` | V1.0 | 1 |
| `L7-AUTO-001` | `L7-BL-005` | V1.0 | 1 |
| `L7-AUTO-002`–`003` | `L7-BL-035` | Later | 2 |
| `L7-AUTO-004` | `L7-BL-037` | Later | 1 |
| `L7-AUTO-005` | `L7-BL-038` | Later | 1 |
| `L7-NFR-001`–`003`, `005`, `022`, `024` | `L7-BL-005` | V1.0 | 6 |
| `L7-NFR-004`, `011`–`014`, `026`–`030` | `L7-BL-014` | V1.0 | 10 |
| `L7-NFR-006`–`010`, `025` | `L7-BL-004` | V1.0 | 6 |
| `L7-NFR-015`–`021`, `031` | `L7-BL-007` | V1.0 | 8 |
| `L7-NFR-023` | `L7-BL-030` | Later | 1 |
| `L7-NFR-032` | `L7-BL-006` | V1.0 | 1 |
| `L7-NFR-033` | `L7-BL-002` | V1.0 | 1 |
| `L7-NFR-034`–`035` | `L7-BL-015` | V1.0 | 2 |
| **Total** |  | **V1.0 140 / V1.x 18 / Later 5** | **163** |

## 9. Sequencing, capacity, and product-learning rules

### 9.1 Recommended implementation order after architecture approval

1. Approve the support/requirement allocation; freeze semantic labels plus the public evaluation protocol and protected-evaluator boundary before tuning.
2. Establish the actual-host read-only walking skeleton, then use the host constraints as architecture inputs rather than discovering them after the core is fixed.
3. Build artifact and safety capabilities as independently testable foundations while cutting the smallest Observer → Planner → Local Changer → Governed Closer walking slice through each; prove malicious and stale inputs fail before relying on them downstream.
4. Extend and smoke every checkpoint through both provisional actual-host adapters so drift is found early, while reserving support/conformance claims for the final adapters and frozen packages.
5. Build deterministic distributions and their full install/update/rollback/uninstall/authenticity lifecycle; run controlled public, differential, and protected evaluation on the exact package bytes.
6. Run staged formative research early, the representative release-candidate pilot only after controlled conformance, and the independent release review only after the pilot evidence is frozen.

### 9.2 Capacity guardrails

- Reserve explicit capacity in every increment for eval fixtures, threat modeling, documentation, migration/recovery, and artifact updates; these are part of the feature estimate.
- Keep one integration owner for shared semantic/state files. Parallel work uses non-overlapping paths or isolation and independently reviewed integration.
- A 13-point epic must be split after architecture exposes stable seams; it cannot enter implementation merely because the team is large.
- Recruit pilot users once the walking skeleton is coherent; run C−1/C0 and C1/C2 formative sessions at their named checkpoints, but do not treat exploratory use as final release evidence until the protocol and package are frozen.
- Track discovery work separately from delivery confidence. External host behavior, protected-holdout operation, signing/channel rules, and pilot adoption remain low-confidence until observed.

### 9.3 Anti-gaming release rules

- Public regression and protected holdout results are reported separately; neither is replaced by hand-picked demonstrations.
- Best-case success cannot hide repeated-trial inconsistency, an incorrect high-risk route/gate, false-low-risk routing, forbidden effects, stale approval, fabricated evidence, secret leakage, package substitution, or evaluator tampering. Blanket blocking or blanket R3 classification is not correct routing.
- `NOT_APPLICABLE` requires profile-based reasoning; missing implementation/capability/evidence is `BLOCKED`, `NOT_EVALUATED`, or `UNVERIFIED` as appropriate.
- A model that authored or remediated a candidate cannot create independent assurance by adopting another persona.
- A passing Codex package cannot promote Claude, and vice versa.
- Dates, sunk cost, pilot enthusiasm, or an aggregate score cannot compensate for one release-blocking safety invariant.

## 10. Backlog risks and decision triggers

| Risk | Early signal | Backlog response / decision trigger |
|---|---|---|
| Semantic kernel becomes an encyclopedia. | Context budgets grow; profiles load universally. | Keep only launch semantics in P0; move specialist knowledge to applicable P1 profiles and fail context-budget tests. |
| Prompt prose is mistaken for enforcement. | Controls have no deterministic/host/human locus. | `L7-BL-005` blocks downstream mutation and marketing claims until enforcement is explicit. |
| Current skills bypass the conductor. | Direct invocation skips classification or approval. | `L7-BL-007` remains incomplete; conform, deprecate, or exclude the entry point. |
| Eval overfitting or leakage. | Candidate reads holdout or thresholds change after failures. | Invalidate run, investigate exposure, rotate affected holdout, require independent re-evaluation. |
| Dual-host scope doubles semantics. | Adapter contains lifecycle/policy rules. | Move meaning back to `L7-BL-002`; adapter remains syntax/capability translation. |
| Artifact system creates false green. | File presence advances state or stale digest passes. | Stop the lifecycle; fix `L7-BL-004` and add a seeded regression before proceeding. |
| Ceremony harms adoption. | Users bypass or cannot explain the next decision. | Improve status/fast path while retaining safety invariants; measure decision interruptions and proportionality. |
| V1 absorbs all professional profiles. | P1 specialist work appears on P0 critical path. | Block scope expansion unless launch proof shows the minimum story cannot be safe without it; missing applicable profile blocks that case. |
| Supply-chain work is deferred. | Source demo works but packages lack provenance/uninstall. | Candidate remains beta/non-release under `L7-BL-014`. |
| Self-healing vision leaks into v1. | Background loop, external mutation, or self-promotion appears. | Treat as forbidden effect; require the separate `L7-BL-035` charter and new approval. |

## 11. Step 2 acceptance gate

This backlog is ready for owner approval when:

- the minimum lovable v1.0 story and strict A0–A2 boundary are acceptable;
- P0 forms one complete local product journey rather than a collection of disconnected specialist skills;
- an early actual-host read-only walking skeleton exposes Codex/Claude constraints without being mistaken for final conformance;
- safety, artifact validity, evaluation, and prototype cutover precede mutation and release claims;
- dependencies expose the true critical path and two host adapters remain separate renderings of one semantic core;
- controlled evaluation, staged pilot evidence, and the independent stable-release decision remain separate gates;
- every P0 item has observable, adversarial acceptance criteria and a coarse effort/confidence estimate;
- all 163 normative requirements have exactly one accountable owner and explicit `V1.0`, `V1.x`, or `Later` allocation;
- P1 covers the professional engineering profiles without forcing them into every context;
- future connectors, multi-repository work, telemetry, ecosystem, and autonomy are visibly gated rather than implied;
- no item authorizes implementation or external action merely by appearing in this artifact.

Approval authorizes **foundation step 3: architecture only**. The next architecture artifact must define boundaries and options for the provider-neutral semantic core, artifacts/state, safety enforcement, evaluation trust boundary, host adapters, packaging, and future replaceable extension points. It must not silently begin harness or product implementation.

## 12. Decision record

| Date | Decision | Actor | Status |
|---|---|---|---|
| 2026-08-24 | Use stable `L7-BL-###` identifiers; priority remains a mutable reviewed field. | Product owner | Approved |
| 2026-08-24 | Make semantic/evaluation/artifact/safety foundations precede conductor and mutation work. | Product owner | Approved |
| 2026-08-24 | Freeze evaluation governance early, grow feature fixtures continuously, and operate the protected holdout only against a frozen candidate. | Product owner | Approved |
| 2026-08-24 | Use an early actual-host A0 walking skeleton, then complete both adapters after the full local lifecycle; provisional smoke is not conformance. | Product owner | Approved |
| 2026-08-24 | Cut over or exclude every prototype entry point so the conductor and authority kernel cannot be bypassed. | Product owner | Approved |
| 2026-08-24 | Limit v1.0 proof profiles to generic, feature/behavior change, and behavior-preserving refactor; block cases needing an unavailable specialist profile. | Product owner | Approved |
| 2026-08-24 | Require actual dual-host packages, a protected holdout, representative pilot, and independent release decision before v1.0. | Product owner | Approved |
| 2026-08-24 | Keep controlled conformance, a staged/14-day adoption pilot, and the independent stable-release decision as separate completion gates. | Product owner | Approved |
| 2026-08-24 | Keep A3, A4, and A5 as distinct later stages behind an approved per-action/environment autonomy charter. | Product owner | Approved |
| 2026-08-24 | Approve `L7-BKL-001` and authorize foundation step 3—architecture only. | Product owner | Approved |
