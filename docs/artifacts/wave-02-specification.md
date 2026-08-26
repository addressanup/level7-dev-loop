# Level 7 Dev Loop — Wave 2 Specification

| Field | Value |
|---|---|
| Artifact ID | `L7-W02-SPEC-001` |
| Artifact type | Proposed implementation specification |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 2 — Provider-neutral semantic and evaluation foundation |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Product | Level 7 Dev Loop |
| Source identity | Commit `c35bf4b6e4a38ca54899882a7e3c574d03d1df85`; tree `eb60ac4d167df96ba02822c458cb81493e05537b` |
| Change contract | [`L7-W02-CC-001`](wave-02-change-contract.md) 0.1.0 |
| Backlog scope | `L7-BL-002`, `L7-BL-003` |
| Normative requirement scope | 29 unique IDs: 23 semantic/knowledge/budget and 6 evaluation-governance requirements |
| Risk/effect ceiling | `R3`; proposal A1 only, later implementation at most A2 after separate exact approval |
| Predecessor evidence | `L7-AUD-W01-006` `GO`, SHA-256 `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f` |
| Next gate | Owner approval or revision of this specification and change contract; approval may authorize design only |

## 1. Outcome and decision boundary

Wave 2 succeeds when one exact local candidate freezes provider-neutral semantic contracts and public evaluation governance before prompt tuning or host implementation. It must prove that invalid semantic combinations fail, every critical obligation is rendered and graded, the compiler rejects dropped or invented obligations, the public protocol is frozen before tuning, candidate authors cannot edit evaluator controls, and correctness does not depend on subagents.

Success establishes **semantic and public-evaluation interfaces for later C−1 work**. It does not establish a user-facing product, canonical lifecycle-state engine, policy kernel, controlled mutation, host package, provider/model compatibility, protected evaluation result, security qualification, support, release, deployment, or exposure.

This specification defines required behavior and evidence. It does not approve an implementation, exact file layout, dependency, branch mutation, test effect, or external action. Those belong to a later exact design and separate implementation decision.

## 2. Source-of-truth order and normative scope

If sources conflict, interpretation follows this order:

1. approved `L7-REQ-001` normative requirements;
2. approved `L7-BKL-001` ownership/allocation and `BL-002`/`BL-003` acceptance;
3. approved architecture invariants, trust zones, dependency rules, and evaluator separation;
4. approved conditional technology decisions, without promoting unproved dependency/host claims;
5. approved orchestration Wave 2 scope, sequencing, exit evidence, and kill conditions;
6. the independently audited Wave 1 predecessor and active fail-closed gate; and
7. an owner-approved final version of this change contract/specification and later design.

The 29 uniquely owned Wave 2 requirements are:

```text
L7-FLOW-001..006, L7-FLOW-008..010
L7-SKILL-001..002
L7-PROMPT-001..002
L7-AGENT-001..003
L7-HOST-001, L7-HOST-005
L7-KNOW-001..004
L7-NFR-033
L7-EVAL-001, L7-EVAL-003..004, L7-EVAL-006, L7-EVAL-008..009
```

No definition may be copied into a Wave 2 registry and silently diverge from its approved requirement. The later implementation must derive and verify this exact ownership set from source rather than trusting this hand-written list.

## 3. Scope and path boundary

### 3.1 Permitted capability classes after later approval

Wave 2 may create only provider-neutral, local, offline development contracts and their public deterministic validators/fixtures:

- semantic taxonomy, lifecycle, obligation, guardrail, workflow, profile, prompt-intermediate, typed-output, knowledge, capability, degradation, budget, loop, stopping, and delegation definitions;
- pure deterministic compiler/renderer interfaces that account for safety obligations but do not generate or install host packages;
- public evaluation protocol, case, truth-label, coverage, grader, adjudication, run-manifest, trial, confidence, cost, and latency contracts;
- public local deterministic fixtures and seeded broken candidates;
- candidate-denied ownership/policy for frozen public evaluator controls; and
- the bounded successor phase/path/ownership gate, local verification, manifests, evidence, and independent audit required to admit those paths.

### 3.2 Prohibited capability classes

Wave 2 SHALL NOT create or enable:

- `cmd/l7`, `cmd/l7up`, a runtime API/CLI, canonical project-state writer/reducer, policy/admission kernel, transaction, executor, receipt, supervisor, host adapter, package, updater, installer, daemon, background process, or user-visible behavior;
- host-specific semantic forks, rendered Codex/Claude packages, edits to prototype skills/manifests, actual host discovery/invocation, or a host support claim;
- model/provider calls, prompt tuning runs, hosted CI, protected evaluation, external evaluator infrastructure, AWS, network egress, credentials, secrets, signing, publication, release, deployment, or exposure;
- protected cases, labels, thresholds, evaluator credentials, detailed protected results, release policy, AP2/AP3 roots, signing, or promotion material inside candidate-readable paths;
- mutation/grant activation, A1/A2 product behavior, feature exposure, or an environment/test/repository bypass;
- a new production dependency, vendored module, `go.sum`, toolchain, action, or package manager without a separate exact approved decision; or
- Wave 3 behavior or evidence.

### 3.3 Fail-closed predecessor transition

The active Wave 1 gate is the last admitted policy. A Wave 2 design SHALL:

1. bind the exact predecessor commit/tree and immutable Wave 1 manifests/evidence/audit;
2. add a distinct successor phase and exact path policy rather than rewriting Wave 1 history;
3. reject unknown phase, path, change class, owner, module, import, source digest, or stale predecessor;
4. place negative fixtures and ownership checks in effect before or atomically with each new semantic/evaluator path;
5. keep protected evaluator/grant/signing/release prefixes denied outside candidate authority; and
6. leave the Wave 1 gate recoverable if transition verification fails.

These two proposal files are not themselves proof that the successor is admitted. The proposal worktree must not be described as a passing candidate.

## 4. Required semantic contracts

### 4.1 Identity, version, compatibility, and deprecation

Every canonical semantic record SHALL carry at least:

| Field | Required semantics |
|---|---|
| Stable ID | Unique, ASCII, namespace-scoped, case-sensitive identity; never reused with changed meaning |
| Schema/record version | Explicit machine-readable version with supported-reader compatibility rule |
| Owner | Exactly one accountable control owner; reviewer/change gate is explicit |
| Definition | One provider-neutral normative meaning; examples are non-normative unless explicitly labeled |
| Status | At least active, draft, deprecated, superseded, or retired with deterministic admissibility |
| Introduced/superseded by | Source/version identity and replacement link where applicable |
| Compatibility | Explicit backward/forward/read compatibility; unknown required fields fail |
| Deprecation | Replacement, warning/error behavior, earliest removal gate, and retained fixture coverage |

Ordering, serialization, and diagnostics SHALL be deterministic. Duplicate IDs, incompatible redefinition, unknown critical fields, cycles in supersession, and missing owners fail closed.

### 4.2 Canonical taxonomy registry

The registry SHALL freeze machine-testable values and definitions for at least:

- lifecycle stage and transition;
- evidence state and evidence authority;
- gate/check result;
- release verdict and product decision;
- heritage and operational state;
- risk class, effect class, approval assurance, and change class;
- freshness, stale, superseded, blocked, recovery, and non-applicability semantics;
- capability/support/degradation state; and
- sensitivity/reference-authority/status categories.

The lifecycle baseline is exactly `Baseline → Frame → Approve → Execute → Verify → Deliver → Observe → Learn`. `Package`, `Deploy`, and `Expose` are separately gated optional transitions within `Deliver`; their presence in the taxonomy grants no capability or authority.

Profiles may skip, collapse, repeat, or mark a stage non-applicable only through an allowed, reason-bearing rule. Every state defines entry, exit, failure, blocked, stale, and superseded conditions. Transition completion requires schema-valid, scope/source-bound, fresh, sufficiently approved artifacts.

At minimum, validation SHALL reject:

- `GO` while an applicable release blocker remains unresolved;
- `PASS` supported only by `UNVERIFIED`, `NOT_RUN`, or `NOT_EVALUATED` input;
- A5 or background/self-modifying behavior in v1;
- an effect above the granted capability or approved risk/gate floor;
- an `AP0` editable record treated as current approval;
- a non-applicable stage without a validated reason;
- a lower-risk profile overriding a higher material risk dimension; and
- an unknown taxonomy value interpreted as success.

### 4.3 Obligation and guardrail ledgers

Every semantic obligation SHALL include:

| Field | Required semantics |
|---|---|
| Obligation ID/version | Stable identity and compatibility/deprecation state |
| Source | One or more exact normative requirement/decision pointers |
| Owner and criticality | Accountable owner; safety-critical/material/noncritical class with rationale |
| Applicability | Stages, profiles, risks, effects, capabilities, hosts, and contraindications |
| Rule | Provider-neutral required/prohibited behavior |
| Enforcement locus | Kernel/writer/executor, host boundary, CI/evaluator, external authority, human confirmation, or prompt-only guidance |
| Renderer requirement | Exact projections in which it must appear or explicit machine-only disposition |
| Grader/evidence | Required public grader/fixture/evidence type and failure semantics |
| Overrideability | Non-waivable or exact typed waiver conditions; safety truth/authority boundaries cannot be waived |
| Lifecycle | Introduced, changed, deprecated, superseded, and retained-test state |

A guardrail record additionally binds its input, decision, failure mode, recovery, and proof. High-consequence prompt-only guidance SHALL be labeled unenforced and cannot support an enforcement claim.

Every safety-critical obligation must have at least one required renderer and one grader or deterministic machine-only enforcement proof. The validator fails on missing, duplicate, dangling, cyclic, silently downgraded, or multiply defined obligations.

### 4.4 Workflow and profile contracts

Every semantic workflow/profile SHALL declare:

- stable ID/version and obligation IDs;
- concise positive and negative triggers;
- prerequisites and canonical input artifacts;
- lifecycle entry/exit/transition and allowed repeat/skip behavior;
- risk inputs/floor, effect ceiling, approval/gate requirements, authority, and tool/capability bounds;
- authoritative context sources, projection rules, sensitivity/freshness handling, and applicable knowledge profiles;
- invariants, prohibited effects, success, failure, blocked, stale, superseded, stopping, recovery, and escalation behavior;
- output artifact and typed decision-first response contract;
- host capability/support requirements without claiming actual host support; and
- public positive, negative, boundary, degraded, interruption, adversarial, and no-subagent fixtures.

Fast, standard, and elevated paths may differ in ceremony only; the highest material risk dimension sets the minimum gate. Numeric recipes are contextual defaults, never universal safety invariants. Multiple applicable profiles compose by obligation union and the highest risk/gate, not by averaging requirements down.

### 4.5 Prompt intermediate representation and compiler

The provider-neutral prompt intermediate representation SHALL order and preserve:

1. bounded goal and current transition;
2. authoritative inputs with identity, provenance, trust, sensitivity, freshness, and inclusion reason;
3. invariants and prohibited effects;
4. authority, tool, capability, risk, and effect bounds;
5. acceptance criteria and proof/evidence obligations;
6. retry, tool-call, subagent, wall-time, token, monetary-cost, stopping, and escalation budgets; and
7. exact typed output shape with decision, evidence, uncertainty, blocker, owner, and one next action.

The compiler SHALL consume immutable semantic registries plus one declared workflow/profile projection and produce deterministic stock-A0 and Controlled Client semantic projections. Wave 2 outputs are interfaces/reference projections only, not installable host packages.

For every compilation the compiler SHALL:

- account for every applicable critical obligation by stable ID;
- reject omitted, weakened, duplicated, unknown, or invented critical obligations;
- reject hidden policy duplicated only in free-form prompt prose;
- preserve machine-only obligations without pretending they were rendered prose;
- use deterministic ordering and bounded output;
- emit source/version/digest and obligation-accounting metadata;
- request no hidden chain-of-thought; and
- perform no network, host, credential, installation, mutation, or ambient discovery effect.

Host-specific syntax and examples belong to later overlays. Removing optional hooks, MCP, browser, memory, telemetry, or subagents cannot alter semantic correctness.

### 4.6 Typed output contract

Every decision-first output schema SHALL distinguish at least:

- decision/result and applicable stable rule IDs;
- exact scope/source/candidate identity;
- evidence state, evidence source/authority, limits, and freshness;
- uncertainty, assumptions, counterevidence/defeaters, and residual risk where applicable;
- blocker, recovery/next action, and accountable decision owner;
- actual effect/effect ceiling and authority state; and
- bounded secret-safe diagnostics.

Unknown critical values, malformed output, contradictory decision/evidence, or missing required fields fail admission. Color, terminal width, icon, prose ordering, or model confidence cannot change semantic meaning.

## 5. Knowledge, budget, loop, and delegation contracts

### 5.1 Knowledge/reference registry

Every entry SHALL record source/pointer, source type and authority, version/date/status, applicability, contraindications, jurisdiction, license/use restriction, freshness/review policy, last review, owner, and normative/non-normative disposition.

The registry distinguishes law, normative standard, official guidance, empirical research, and practitioner pattern. Draft, emerging, disputed, superseded, stale, or restricted material remains labeled and cannot silently become a mandatory invariant. Proprietary/restricted standards are linked and mapped without unauthorized reproduction.

Only applicable, freshness-valid profiles may be selected. Missing or unsafe-to-use required knowledge returns `BLOCKED`/`NOT_EVALUATED`, never a fabricated or generic pass.

### 5.2 Budgets, retries, stopping, and degradation

Contracts SHALL define policy-owned maxima for tool calls, subagents, retries, wall time, tokens, context bytes/items, output size, and monetary cost. Each maximum includes measurement source, scope, exhaustion result, recovery/escalation, and whether a smaller host limit lowers capability.

Budget exhaustion, repeated identical failure, oscillation, cancellation, missing capability, or uncertain effect stops or degrades safely. It never reduces risk, approval, evidence, evaluator, or safety gates. Defaults must be justified by applicable evidence rather than copied as universal recipes.

### 5.3 Delegation

Delegation is optional and never a correctness dependency. A delegation contract SHALL bind objective, disjoint scope, inputs/projection, authority/effect ceiling, allowed tools, budgets, output schema, evidence, verifier, integration owner, and termination.

Subagents receive no approval capability, credentials, protected evaluator data, authority expansion, or overlapping direct write access. Parallel writes require disjoint paths or isolated workspaces and one writer/integrator. Every acceptance fixture must have a no-subagent execution with the same semantic result.

## 6. Public evaluation governance

### 6.1 Protocol freeze

Before prompt, workflow, skill, routing, grader, or threshold tuning, one versioned public protocol SHALL freeze:

- scenario selection/stratification and coverage ownership;
- truth-label schema, provenance, owner, version, and change policy;
- supported host/model/version fields as labels only, without asserting support;
- run/trial count, seed/randomness policy, sampling, repeated-trial aggregation, and consistency reporting;
- deterministic grader selection and model-judge calibration/authority limits;
- adjudication, ambiguity, confidence, counterexample, and human-review rules;
- cost, latency, resource, timeout, cancellation, and output-bound recording; and
- per-invariant failure thresholds plus the rule that safety invariants cannot be averaged or tuned away.

Any material protocol/control change creates a new version and invalidates affected results. A candidate cannot choose the protocol version after observing results.

### 6.2 Case and truth-label contracts

A public case SHALL bind stable case ID/version, owning backlog feature, scenario/risk/effect/profile axes, input fixture identity, allowed capabilities/tools/effects, prohibited effects, expected output schema, truth-label IDs, grader IDs, resource bounds, isolation assumptions, sensitivity, and deterministic setup/teardown limitations.

A truth label SHALL bind stable ID/version, exact case/protocol, expected semantic outcome and evidence, authority/owner, source/rationale, adjudication state, compatibility, and exposure class. Candidate output cannot alter truth.

Fixtures SHALL use synthetic, non-sensitive data. Secret-leak tests use only clearly synthetic canary values and must assert nonappearance in outputs/logs/artifacts.

### 6.3 Grader registry

Each grader SHALL declare stable ID/version, owner, deterministic/model/human class, accepted input/output schemas, covered obligation/truth IDs, result semantics, bounds, error behavior, calibration/adjudication needs, and authority limitations.

Deterministic graders are required whenever the answer is computable. Model judges may supplement ordering, clarity, or other ambiguous quality only after calibration for ordering, verbosity, and model-family bias; their output is evidence and never independent safety authority. Grader error or ambiguity cannot become a pass.

### 6.4 Coverage map and seeded broken candidates

The coverage map SHALL assign at least routing/negative activation, lifecycle transitions, stale evidence/approval, authority, forbidden effects, degraded modes, interruption/resume, parity obligations, write-collision semantics, install-lifecycle placeholders, injection, secret handling, and budget exhaustion to one owning feature plus obligation, case, truth, and grader IDs.

Wave 2 SHALL include seeded public broken candidates that:

- drop or weaken a critical obligation;
- invent an obligation or unsupported approval/evidence;
- route low on high risk;
- treat stale approval as current;
- fabricate evidence or turn `UNVERIFIED` into `PASS`;
- leak a synthetic secret/canary;
- attempt a forbidden effect or authority expansion; and
- make correctness depend on a subagent.

Each broken candidate must fail for the intended stable rule. Feature teams add later cases through evaluator-governed integration; they cannot weaken a failing oracle.

### 6.5 Run-manifest contract

Every run manifest SHALL bind:

- candidate/source and relevant semantic/workflow/prompt/skill/protocol/grader versions and digests;
- exact host/model/harness/tool/environment identifiers or explicit `NOT_APPLICABLE`/`UNVERIFIED` state;
- case/truth selection, trial, seed, start/end/duration, resource budgets, cost, latency, and termination;
- allowed/observed effects and isolation/effect limitations;
- per-case and aggregate results without hiding forbidden-action failures or cherry-picking best runs;
- producer/evidence authority, adjudication, uncertainty, and invalidation state; and
- secret-safe diagnostics and exact reproduction limits.

Missing identity, malformed results, unexpected effect, grader failure, inconsistent trial accounting, or threshold violation is non-passing.

### 6.6 Candidate-denied public controls

The design SHALL separate:

1. feature-owned ordinary public cases;
2. frozen evaluator-control code, oracles, truth labels, authorization controls, adjudication, coverage indexes, and thresholds; and
3. the external protected evaluation/release plane.

Candidate and remediation writers SHALL have no authority to change class 2 while fixing the candidate. Authorized public-control changes require a separate owner/reviewer transition, new digest/version, affected-result invalidation, and audit. Attempts to cross the boundary fail and are recorded.

### 6.7 Protected-holdout contract

Wave 2 specifies but does not instantiate the protected plane. The contract SHALL require:

- at least the approved release-holdout proportion outside candidate/runtime/author/remediator read, list, and write scope;
- separate operator/evaluator authority and independent case/label sampling under frozen stratification;
- fresh isolated credential-free candidate workspace per case and bounded inputs/outputs/resources/egress;
- external credentials, labels, thresholds, detailed results, adjudication, and release policy;
- bounded aggregate feedback only, exposure/tamper detection, case rotation, invalidation, rate/submission controls, and human exposure response; and
- exact candidate/package/protocol/host/model/environment/trial/cost/latency binding in later evaluator attestations.

No hidden/protected case, credential, detailed result, operator implementation, or infrastructure is created or claimed by Wave 2.

## 7. Deterministic interfaces and diagnostics

Local Wave 2 validators/compilers/graders SHALL be offline, deterministic, bounded, non-mutating with respect to product state, and explicit about allowed repository cache/temp effects. Their common decision-first result SHALL include:

| Outcome | Required behavior |
|---|---|
| `PASS` | Exit zero; stable tool/version, source digests, protocol/schema versions, bounded totals, and exact result |
| Validation/policy failure | Nonzero; stable rule ID, exact record/path/ID, bounded message, one recovery action; no repair |
| Missing capability/input | Nonzero `BLOCKED`/`NOT_EVALUATED`/`UNVERIFIED` as applicable; never inferred success |
| Resource exhaustion | Nonzero bounded failure; no truncation-and-pass or gate reduction |
| Unexpected internal error | Nonzero secret-safe diagnostic; predecessor gate/control remains effective |

Output SHALL be byte-stable for identical admitted inputs under declared locale/timezone and ordering. It SHALL contain no credentials, ambient environment dump, terminal control sequence, hidden reasoning, protected data, or unbounded source content.

## 8. Quality attributes

### 8.1 Safety and integrity

- Schemas, semantic files, templates, fixtures, manifests, repository content, and tool/model output are untrusted inputs.
- Readers use rooted repository-relative regular-file inputs, reject path escape and unexpected link/node types, and declare size/count/depth/time limits.
- Semantic validation cannot grant approval, capability, release verdict, evaluator authority, or mutation.
- Candidate-controlled fields cannot override owners, criticality, enforcement locus, truth, adjudication, thresholds, or protected boundaries.
- Synthetic secrets/canaries never become credentials and must not survive output/log/artifact paths.

### 8.2 Determinism and portability

- Identical canonical input bytes produce identical semantic results and obligation accounting on the pinned baseline and shadow Go toolchains.
- Results do not depend on map iteration, filesystem order, locale, wall clock, Git user identity, network, host/model output, ambient environment, or absolute path.
- Time, randomness, host, model, resource, and cost facts enter only through explicit interfaces/manifests.
- macOS local development evidence cannot become Ubuntu/host/provider/support evidence.

### 8.3 Boundedness and performance

- The design SHALL set exact file/count/byte/depth/output/time/trial/context/tool/retry/subagent/token/cost bounds and fail behavior before implementation.
- Bounds reject before unbounded allocation, traversal, rendering, or output where practicable.
- No daemon, watcher, background task, implicit retry, or network fallback is permitted.

### 8.4 Usability and accessibility

- Diagnostics lead with decision and stable rule, name the exact subject, and provide one recovery action.
- Typed status preserves decision, evidence, uncertainty, blocker, owner, and next action independent of prose, color, terminal width, or icons.
- Concision or context budgets cannot remove safety-critical obligations.

### 8.5 Compatibility and preservation

- Existing Wave 1 verification and history remain distinguishable and reproducible from frozen inputs.
- A successor does not relabel Wave 1 tests/results as Wave 2 evidence.
- Unknown required schema fields fail; compatible readers preserve explicitly permitted unknown noncritical data only if the design proves safe semantics.
- Generated/reference outputs are disposable and cannot be reverse-promoted into authored truth.

## 9. Verification specification

### 9.1 Required layers

| Layer | Required evidence |
|---|---|
| Source/trace | Exact 29-ID derivation, unique ownership, schema/ID/version grammar, source digests, path/change ownership |
| Unit/table | Taxonomy transitions/combinations, registry validation, obligation accounting, applicability/profile composition, budgets, knowledge, protocol, manifests, graders |
| Property/metamorphic | Input ordering, serialization ordering, duplicate/unknown fields, omitted/invented obligations, profile composition, trial aggregation |
| Positive integration | Exact admitted Wave 2 semantic/evaluator candidate compiles and evaluates deterministically under the successor gate |
| Negative/adversarial | Every invalid combination, boundary violation, seeded broken candidate, candidate-control mutation, stale predecessor, and malformed registry fails for the intended stable rule |
| No-subagent | The same accepted semantic result and evidence are produced with delegation unavailable |
| Degraded/interruption | Missing inputs/tools/capabilities, exhaustion, partial transition, stale evidence, cancellation, grader error, and interrupted candidate remain fail closed |
| Baseline/shadow | Complete pinned Go 1.26.7 matrix blocking; Go 1.27.0 shadow separate and unable to mask baseline failure |
| Manifest/closure | Exact predecessor, source, path-policy, candidate, evidence, exclusions, file modes/types, dependency state, and no-secret closure |
| Independent review | Fresh structurally separate read-only audit of exact candidate/evidence; zero unresolved Blocker/Critical/High/Medium |

### 9.2 Mandatory fixture inventory

The later design SHALL assign exact tests for at least:

- duplicate/unknown/redefined IDs, incompatible versions, cycles, missing owners, and invalid deprecation;
- every invalid combination in §4.2 plus boundary values for risk/effect/approval/evidence;
- missing/duplicate/downgraded/dangling renderer and grader mappings;
- dropped, weakened, duplicated, unknown, and invented prompt obligations;
- applicable versus universal profile/reference loading and context-budget exhaustion;
- retry/no-progress/oscillation/cancellation/budget exhaustion without gate reduction;
- delegation authority expansion, overlapping writers, malformed manifest, and no-subagent equivalence;
- draft/stale/disputed/restricted knowledge classification and license-safe mapping;
- protocol change invalidation, truth/case mismatch, trial accounting, grader error, adjudication, and model-judge authority limits;
- candidate attempts to edit frozen evaluator code/oracles/truth/adjudication/authorization/thresholds;
- all seeded broken candidates in §6.4;
- unknown Wave 2 phase/path/owner/import and stale predecessor;
- path escape, symlink/special-node, oversized/deep inputs, output caps, terminal controls, and synthetic-secret nonleakage; and
- exact candidate/evidence manifest and independent-audit separation.

### 9.3 Test-effect boundary

No test or build command was run while authoring this proposal. Before a later run, the owner-approved design/implementation action SHALL declare exact commands, pinned toolchains, repository cache/temp/log/reproducibility effects, environment allowlist, network denial, credentials, cleanup, duration, and interruption behavior.

The intended development envelope is local and offline with declared ignored repository-scoped effects. It excludes downloads, new dependencies, hosted CI, provider/model/host calls, protected evaluation, external credentials, installation, remote Git, signing, publication, release, deployment, exposure, and ambient cleanup.

### 9.4 Evidence rules

Every executed check SHALL bind method/command, exact candidate/commit/tree/dirty state, input/source/protocol/schema digests, toolchain/OS/architecture, environment/effect roots, time/duration, result, stable rule output, producer/authority, and limitations. `NOT_RUN`, `NOT_EVALUATED`, `UNVERIFIED`, `BLOCKED`, and `RECOVERY_REQUIRED` remain distinct from `PASS`.

Passing tests cannot waive an authority, identity, candidate-control, protected-boundary, or material audit finding.

## 10. Acceptance criteria and traceability

| Acceptance ID | Requirement | Source mapping |
|---|---|---|
| `W02-AC-001` | Exactly the 29 Wave 2 normative IDs derive from approved source, have one owner/allocation, and map to tests/evidence with no unknown/duplicate/missing ID. | `BL-002`, `BL-003`; Wave 1 trace invariant |
| `W02-AC-002` | Versioned taxonomies/lifecycle reject invalid states/combinations and preserve entry, exit, blocked, stale, superseded, and reasoned non-applicability semantics. | `FLOW-001..003` |
| `W02-AC-003` | Every critical obligation/guardrail has one meaning, owner, enforcement locus, renderer, grader, version, and deprecation state; provider-neutral source is singular. | `HOST-001`; architecture `AI-04`; `BL-002` AC1/7 |
| `W02-AC-004` | Workflow/profile contracts cover triggers, prerequisites, inputs, lifecycle, applicability, risk/gates, authority/effects, outputs, stopping, host capability, references, and fixtures without universal-profile or numeric-recipe misuse. | `FLOW-004..006`, `FLOW-009..010`, `SKILL-001..002` |
| `W02-AC-005` | Prompt IR/compiler preserves complete bounded obligations and typed output; omitted/weakened/duplicated/invented critical obligations fail for stock-A0 and Controlled Client projections. | `PROMPT-001..002`, `HOST-001`, `BL-002` AC3/4/7 |
| `W02-AC-006` | Budgets, loops, stopping, degradation, optional delegation, one-writer integration, and no-subagent correctness are machine-tested; optional host accelerators cannot change correctness. | `AGENT-001..003`, `HOST-005`, `NFR-033` |
| `W02-AC-007` | Every knowledge/reference entry carries authority/type/version/status/applicability/contraindication/jurisdiction/license/freshness/review metadata and cannot silently promote stale/restricted guidance. | `KNOW-001..004`, `FLOW-005` |
| `W02-AC-008` | Public protocol/case/truth/run-manifest contracts are provider-neutral, locally executable where applicable, frozen before tuning, and bind exact candidate/environment/trial/resource/cost/latency identity. | `EVAL-001`, `EVAL-006`, `EVAL-009` |
| `W02-AC-009` | Deterministic graders cover computable outcomes; calibrated model judges expose ordering/verbosity/family bias and never act as independent safety authority. | `EVAL-003..004` |
| `W02-AC-010` | Coverage and seeded broken candidates prove detection of dropped/invented obligations, fabricated evidence, stale approval, secret leakage, unsafe routing, forbidden effects, and subagent dependence. | `BL-003` AC2/6; `BL-002` AC7 |
| `W02-AC-011` | Candidate/remediation writers cannot modify frozen evaluator code, oracles, truth labels, adjudication, authorization controls, or thresholds; protected boundary remains external and uninstantiated. | `EVAL-008`; `BL-003` AC4/5 |
| `W02-AC-012` | Source-bound Wave 2 phase/path/ownership transition preserves exact Wave 1 history and fails closed before any unknown/unowned semantic/evaluator path is admitted. | `L7-ORC-001` Wave 2 dependency; `L7-W01-DES-001` future seam |
| `W02-AC-013` | Full baseline/shadow local matrices, exact manifests/closure/effects/limits, and fresh independent exact-digest audit pass with no unresolved Blocker/Critical/High/Medium. | `FLOW-008`; R3 development gate |
| `W02-AC-014` | No dependency, product command/runtime, host package/prototype edit, protected asset, credential, external action, mutation capability, host/support/stable claim, release, or deployment is introduced. | Change-contract non-goals; architecture trust boundary |

No aggregate pass may waive one failed criterion.

## 11. Feature-flag and exposure decision

A runtime feature flag is `NOT_APPLICABLE`: Wave 2 creates no user-visible or production behavior. The fail-closed phase/path gate and candidate-denied evaluator-control partition are admission/governance controls, not rollout flags or product authorization.

There is no cohort, exposure percentage, production telemetry, provider observation, or removal schedule. Any later user-visible behavior requires its own default-OFF feature/exposure contract.

## 12. Failure, recovery, and stopping behavior

| Failure | Safe result | Required recovery/next action |
|---|---|---|
| Trace/owner/allocation mismatch | `BLOCKED`; no semantic freeze | Correct approved source/mapping under owner decision; rerun on a new candidate |
| Invalid/ambiguous semantic record | Nonzero; record excluded | Correct schema/source; never infer or default to success |
| Missing/dropped/invented critical obligation | Compilation/evaluation fails | Correct authored semantic source or renderer/grader mapping; do not patch generated/prose output |
| Budget/capability exhaustion | `BLOCKED`/`NOT_EVALUATED` or lower capability | Escalate or narrow scope; never lower gate/risk/evidence requirements |
| Protocol/truth/grader change | Affected evidence invalidated | Version/freeze controls through separate authority; rerun exact candidate |
| Candidate evaluator-control write | Candidate fails; possible evidence invalidation | Preserve attempt evidence; restore frozen control through evaluator owner; re-audit |
| Protected data/credential exposure | Stop; invalidate affected run/material | Contain/rotate externally, assess exposure, replace cases as required; no candidate remediation of hidden truth |
| Partial phase/path transition | Last audited Wave 1 gate remains authoritative | Restore exact predecessor or complete only under approved preimage/recovery plan |
| Unexpected dependency/path/network/host/provider/external effect | Stop; candidate fails | Preserve evidence, assess/revoke authority if applicable, obtain new scoped decision |
| Audit Blocker/Critical/High/Medium | No Wave 2 checkpoint | Finding-specific remediation candidate followed by fresh independent audit |

Every result ends with one decision-first state and one permitted next action. Wave 3 never begins automatically.

## 13. Required design decisions

Approval of this specification does not decide implementation. The Wave 2 design proposal SHALL bind:

1. exact branch/base/predecessor-manifest strategy and how the audit-only Wave 1 child is preserved;
2. an acyclic phase/path-policy transition and exact per-slice path/change/owner allowlists;
3. exact semantic, obligation, guardrail, workflow, profile, prompt-IR, output, knowledge, protocol, case, truth, grader, coverage, adjudication, and run-manifest schemas;
4. stable ID grammars, version/compatibility/deprecation rules, source grammar, and deterministic diagnostics/rule IDs;
5. serialization/canonicalization boundaries and what validates with zero new production dependencies versus what remains a future interface;
6. pure compiler/renderer and public evaluator package APIs, dependency directions, source/output digest model, and generated-output non-authority;
7. exact public evaluator-control partition, writers/reviewers/change gate, candidate-denial enforcement, invalidation, and separate feature-fixture integration path;
8. protected-holdout contract artifact without cases, labels, credentials, operator code, or infrastructure;
9. exact taxonomy invalid-combination matrix, obligation criticality/locus rules, and renderer/grader coverage accounting;
10. exact protocol values for selection, trials, sampling, adjudication, confidence, model-judge calibration, resource/cost/latency capture, and failure thresholds;
11. exact fixture/broken-candidate inventory and how negative controls cannot be mistaken for product or protected source;
12. exact file/count/byte/depth/context/output/time/tool/retry/subagent/token/cost bounds and pre-allocation failure behavior;
13. serial implementation slices, preimages, one-writer ownership, interruption/recovery, and last-green checkpoints;
14. exact baseline/shadow verification commands and declared repository cache/temp/reproducibility effects;
15. candidate/evidence manifest construction, non-circular exclusions, no-secret/dependency/mode/type checks, and independent audit handoff; and
16. documentation/status outputs and the explicit Wave 3 stop.

Any decision that expands scope, paths, effects, dependencies, external boundary, product behavior, authority, risk, or acceptance requires contract/specification revision and reapproval before design continues.

## 14. Deliverables and completion state

A completed Wave 2 candidate is expected to contain, through later separately approved actions:

- versioned provider-neutral taxonomy/lifecycle registries;
- obligation and guardrail ledgers with renderer/grader accounting;
- semantic workflow/profile, prompt-IR, typed-output, knowledge, budget, loop, degradation, and delegation contracts;
- pure deterministic compiler interfaces and reference projections, not host packages;
- frozen public evaluation protocol, cases/truth/run/grader/coverage/adjudication contracts;
- seeded public broken candidates and no-subagent fixtures;
- candidate-denied public evaluator-control ownership plus an external protected-holdout contract;
- source-bound Wave 2 phase/path/ownership policy and permanent boundary fixtures;
- baseline/shadow local evidence, exact candidate/evidence manifests, and effect/limit records;
- fresh independent exact-digest audit; and
- a Wave 2 status record that reports every `PASS`, failure, `BLOCKED`, `NOT_RUN`, `NOT_EVALUATED`, and residual limitation truthfully.

Artifact presence alone does not complete the wave. The checkpoint may be recorded only when every `W02-AC-*` criterion passes against one exact candidate and a fresh independent audit finds no unresolved Blocker/Critical/High/Medium.

## 15. Approval record

No approval is embedded in this specification. The accountable owner may approve it together with `L7-W02-CC-001`, request revision, or reject it. Until a fresh exact decision is given, Wave 2 design and implementation are blocked.
