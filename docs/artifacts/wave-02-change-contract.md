# Level 7 Dev Loop — Wave 2 Change Contract

| Field | Value |
|---|---|
| Artifact ID | `L7-W02-CC-001` |
| Artifact type | Proposed wave change contract |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 2 — Provider-neutral semantic and evaluation foundation |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Product | Level 7 Dev Loop |
| Canonical root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Planning branch | `feat/wave-02-semantic-evaluation` |
| Source identity | Commit `c35bf4b6e4a38ca54899882a7e3c574d03d1df85`; tree `eb60ac4d167df96ba02822c458cb81493e05537b` |
| Predecessor checkpoint | Wave 1 exact development tuple independently audited `GO`; audit SHA-256 `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f` |
| Backlog | `L7-BL-002` and `L7-BL-003`; 29 uniquely owned normative requirements |
| Primary change class | Architecture/modernization and feature/behavior change |
| Secondary change class | Security/evaluation-governance change |
| Risk | `R3 — high`, because this wave defines safety-critical semantics, evaluator controls, and the first product-source path admission |
| Effect of this proposal action | One local planning branch ref created from the exact predecessor, then A1 addition of only this record and `wave-02-specification.md`; no existing-file edit, staging, or commit |
| Maximum later Wave 2 effect | A2 local repository change only after separate approval of an exact design, paths, effects, recovery, and implementation action |
| Approval state | The current owner directive authorizes the planning branch and this exact two-file proposal action. Persisted text is `AP0` and grants no replay or implementation authority. |
| Sensitivity | Internal product, semantic-policy, and evaluator-governance planning; no secret, credential, protected case, or personal-data payload |
| Next gate | Accountable-owner approval or revision of this contract and `L7-W02-SPEC-001`; approval may authorize a design proposal only |

## 1. Decision requested

Anup Pandey, as accountable owner, is asked to approve or revise this exact change contract together with [`L7-W02-SPEC-001`](wave-02-specification.md).

Approval of both proposal records would authorize only the next bounded step: preparation of a Wave 2 design proposal that binds exact schemas, path ownership, public evaluator controls, compiler interfaces, implementation slices, verification effects, recovery, predecessor identity, candidate/evidence construction, and independent-audit seam. It would not authorize implementation, staging, a commit, a dependency, a merge, a remote or hosted action, a provider/model/host trial, publication, release, deployment, exposure, or autonomous continuation.

## 2. Authorization and predecessor basis

The accountable owner approved starting `level7-dev-loop:l7-build` for Wave 2 on the expressly stated boundary: preserve the exact Wave 1 GO audit, then produce only the Wave 2 change contract and specification before stopping for approval.

That bounded action has these conditions:

- Wave 1 audit bytes were first preserved in commit `c35bf4b6e4a38ca54899882a7e3c574d03d1df85`;
- the new branch `feat/wave-02-semantic-evaluation` starts exactly at that commit; local `main` remains unchanged at `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`;
- no merge, integration, remote, hosted CI, release, deployment, or external action is implied;
- only `docs/artifacts/wave-02-change-contract.md` and `docs/artifacts/wave-02-specification.md` may be added during this proposal action;
- both files remain uncommitted until the owner decides on their exact bytes; and
- the authority expires when the complete pair is presented or when source identity, scope, risk, effect, ownership, or governing input changes materially.

The active Wave 1 path policy intentionally does not admit new Wave 2 paths. Therefore these two owner-authorized planning records are not represented as a passing Wave 2 candidate. The last committed predecessor remains the green, audited checkpoint. A later design must provide an acyclic, fail-closed transition that registers the exact planning and implementation paths before any product-source capability is treated as admitted.

## 3. Desired outcome

Wave 2 will freeze one provider-neutral semantic source and one public evaluation-governance contract before prompt prose, host packages, or product behavior are tuned. A contributor must be able to prove that:

1. lifecycle, evidence, gate, risk, effect, approval, decision, and change-class terms have one versioned machine-testable meaning;
2. every safety-critical semantic obligation has one stable ID, owner, enforcement locus, renderer requirement, grader, version, and deprecation state;
3. workflow, profile, prompt intermediate representation, typed output, knowledge, budget, stopping, delegation, capability, and degradation contracts cannot silently drop or invent safety obligations;
4. the public evaluation protocol, truth-label schema, run-manifest contract, deterministic grader rules, coverage map, adjudication, trial, cost, and latency semantics are frozen before tuning;
5. candidate authors cannot modify frozen evaluator controls, oracles, truth labels, adjudication, authorization controls, or thresholds;
6. the protected-holdout boundary is specified as external and separately operated without placing protected cases or credentials in this repository; and
7. seeded broken candidates prove the available public gates reject fabricated evidence, stale approval, secret leakage, unsafe routing, forbidden effects, and dropped obligations.

The Wave 2 checkpoint is **semantic and public-evaluation interfaces frozen for C−1; no host support, controlled mutation, or release claim**.

## 4. Proposed scope

### 4.1 Ordered work packages after later design and implementation approval

| Work package | Outcome |
|---|---|
| `W02-WP-01` | Add a source-bound Wave 2 phase/path/ownership successor with negative fixtures while preserving the entire Wave 1 candidate, evidence, and audit lineage. |
| `W02-WP-02` | Freeze canonical taxonomy and lifecycle registries, including valid/invalid cross-taxonomy combinations, versioning, deprecation, and stable IDs. |
| `W02-WP-03` | Freeze obligation and guardrail ledgers with owner, criticality, enforcement locus, renderer requirement, grader, evidence, overrideability, version, and deprecation state. |
| `W02-WP-04` | Define semantic workflow/profile schemas, prompt intermediate representation, typed output contracts, and deterministic compiler interfaces for stock-A0 and Controlled Client projections without generating host packages. |
| `W02-WP-05` | Define knowledge/reference metadata plus budget, stopping, retry, loop, delegation, capability, degradation, and no-subagent correctness contracts. |
| `W02-WP-06` | Freeze the public evaluation protocol: cases, truth labels, coverage, run manifests, grader registry, adjudication, repeated trials, confidence, cost/latency, and failure thresholds. |
| `W02-WP-07` | Establish candidate-denied evaluator-control ownership and specify the external protected-holdout/operator/exposure-response contract without adding hidden cases, credentials, or protected infrastructure. |
| `W02-WP-08` | Land provider-neutral positive, negative, boundary, degraded, interruption, and seeded-broken-candidate fixtures; freeze exact candidate/evidence manifests and obtain independent read-only audit. |

The later design must narrow the maximum implementation envelope to exact paths and change classes. The proposed envelope is limited to:

- new Wave 2 proposal/design/approval/candidate/evidence/audit records under `docs/artifacts/`;
- successor phase, path, module, import, ownership, and source-digest controls under `harness/` and `internal/harness/buildcontrol/`, plus only the exact required `Makefile`, workflow, script, or `README.md` changes;
- provider-neutral authored sources under `semantic/taxonomy/`, `semantic/workflows/`, and `semantic/profiles/`;
- public contract schemas under `schemas/`;
- public deterministic cases under `fixtures/public/`;
- pure compiler/renderer interfaces under `internal/render/`; and
- the public local evaluator and deterministic graders under `internal/evaluator/`.

No path in that envelope is admitted by this proposal. Boundary policy and negative fixtures must precede or land atomically with the first governed path, and the exact design must assign one writer/reviewer to every shared control.

### 4.2 Explicit non-goals

Wave 2 will not:

- create `cmd/l7`, `cmd/l7up`, a user-facing CLI/API, a host package, generated Codex/Claude output, a Controlled Client, kernel, transaction writer, executor, receipt boundary, adapter, installer, updater, or daemon;
- edit the 12 protected prototype skills, `.codex-plugin/`, `.claude-plugin/`, root/marketplace manifests, `references/WORKFLOW.md`, or host installation/configuration;
- tune production prompt prose against evaluation outcomes or claim prompt, host, model, provider, compatibility, containment, security, or support quality;
- add protected/hidden cases, protected labels, evaluator credentials, release thresholds, signing, promotion, grant activation, or external evaluator infrastructure to the candidate repository;
- run an actual Codex, Claude, provider, network, hosted CI, AWS, Ubuntu, Bubblewrap, signing, publication, release, deployment, or exposure action;
- add or update a production dependency, `go.sum`, `vendor/`, toolchain, action, or module identity without a separately justified and owner-approved design decision;
- implement canonical lifecycle state, policy admission, A1/A2 mutation, host rendering, package generation, or distribution behavior owned by later waves;
- modify or relabel any approved Wave 1 artifact, candidate manifest, evidence, audit, or historical source; or
- begin Wave 3 or continue from specification to design/implementation without the required new decision.

## 5. Bound inputs and evidence state

| Input | Version / SHA-256 | State for this proposal |
|---|---|---|
| `AGENTS.md` | `54496725a42eb7e6cce2a749e82a408d7277743ec8ad83c41373ceefbd4d0afa` | `OBSERVED`; contributor policy |
| `L7-REQ-001` | 0.2.0 / `a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f` | Approved foundation input; current bytes reproduced |
| `L7-BKL-001` | 0.1.0 / `df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995` | Approved foundation input; owns `BL-002` and `BL-003` |
| `L7-ARC-001` | 0.2.0 / `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` | Approved architecture input; current bytes reproduced |
| `L7-TEC-001` | 0.2.0 / `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` | Approved conditional technology baseline; no dependency or support claim inferred |
| `L7-HAR-001` | 0.1.0 / `d56c8f6880e1bcfe5466d103cc338b087d77c973c30cb656c574971ecce3a53c` | Approved foundation harness record |
| `L7-ORC-001` | 0.3.1 / `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` | Approved orchestration input; Wave 2 scope and ordering source |
| Wave 1 evidence | `L7-EVD-W01-006` / `1d350436398fad8f53a6221fc1c1f2e64ac9bfa0f1b8c5317f1003c1a198b98c` | Exact local candidate evidence; same-user and local |
| Wave 1 independent audit | `L7-AUD-W01-006` / `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f` | `GO` for exact Wave 1 tuple; one LOW and two INFO findings remain explicit |
| Wave 1 source design | `L7-W01-DES-001` / `07953b2319635846505a018c3e4cc66705e0c263ab01b0a5c79e75cdaf1fb8e8` | Defines the fail-closed future transition seam |
| Shared control ownership | `harness/control-ownership.tsv` / `5f043166e9d698ceba278e22ce182a396faefd5eac929ac988dc6f25660fa8d8` | Reserves semantic, schema, public-fixture, public-evaluator, and protected-control ownership classes |
| Installed `l7-build` transport skill | `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` | Reproduced transport instructions; not authority |
| Current local predecessor | commit `c35bf4b6e4a38ca54899882a7e3c574d03d1df85`, tree `eb60ac4d167df96ba02822c458cb81493e05537b` | Audited Wave 1 lineage plus audit-only preservation commit; branch created clean; no remote |

## 6. Invariants

1. One provider-neutral semantic source owns safety-critical meaning. Host syntax, examples, or future overlays may translate but cannot weaken, duplicate, or invent policy.
2. Every semantic and evaluator-control record has a stable ID, explicit version, owner, status, and compatibility/deprecation behavior. Reuse of an ID with changed meaning is forbidden.
3. Invalid combinations fail deterministically. `GO` with an unresolved release blocker, `PASS` derived only from `UNVERIFIED`, or A5 in v1 can never validate.
4. Every critical obligation has at least one required renderer and grader. Missing, duplicated, weakened, or invented obligations block compilation/evaluation.
5. Model output, prompt prose, fixtures, repository content, and test output are untrusted data. They cannot grant approval, lower risk, alter truth, or activate a capability.
6. The public evaluation protocol freezes scenario selection, truth-label semantics, versions, runs, sampling, adjudication, confidence, cost/latency, and failure thresholds before tuning.
7. Candidate authors and remediators cannot write frozen evaluator code, oracles, truth labels, adjudication, authorization controls, or thresholds. A control change versions/invalidates affected evidence.
8. Protected cases, labels, credentials, detailed results, release policy, signing, promotion, AP2/AP3 roots, and external evaluator authority remain outside candidate read/list/write scope.
9. Deterministic graders are authoritative where results are computable. A calibrated model judge may supplement but is never independent safety or release authority.
10. Delegation is optional. No acceptance result may depend on a subagent; delegation never expands authority and parallel writes require disjoint scope plus one integrator.
11. Budgets for tools, subagents, retries, wall time, tokens, and money are policy-owned; exhaustion stops or escalates and cannot reduce a gate.
12. Knowledge records distinguish normative, official, empirical, and practitioner sources; draft, disputed, superseded, restricted, and stale material remains labeled.
13. Wave 1 history remains immutable. A new phase/path policy references its predecessor and cannot rewrite it to appear current.
14. Boundary policy and adversarial fixtures are effective before a new semantic/evaluator path is admitted. Unknown phase/path/owner/control remains denied.
15. All local validation is deterministic and offline under declared repository-scoped effects. Unrun host, provider, hosted, protected, or release checks remain `NOT_RUN`/`NOT_EVALUATED`.
16. Wave 2 creates no user-visible behavior, mutation capability, host support, compatibility, security qualification, publication, release, deployment, or support claim.

## 7. Preconditions and unresolved design decisions

| Item | Current state | Required before implementation |
|---|---|---|
| Wave 1 dependency | Exact fifth-successor audit is `GO`; `W01-AC-012` cleared | Bind this exact predecessor into the Wave 2 base/manifest; do not substitute mutable branch state |
| Branch strategy | Wave 2 branch stacks on audited Wave 1 head; `main` remains unchanged | Design exact commit slices, recovery, candidate freeze, and later integration boundary |
| Phase/path admission | Active gate has exactly one Wave 1 row and denies all new paths | Design a source-bound successor row/path policy plus negative fixtures; no universal future-path allowance |
| Bootstrap planning admission | These two proposal files are owner-authorized but not in the active candidate path policy | Design an acyclic admission method; do not call the proposal worktree a passing candidate |
| Semantic namespaces | Requirement IDs exist; semantic/obligation/guardrail/schema IDs are not frozen | Define grammars, owners, versioning, compatibility, deprecation, and collision rules |
| Serialization/schema runtime | JSON 2020-12/JCS is selected; strict runtime dependency qualification belongs to later exact decisions | Decide what Wave 2 can validate with zero new production dependencies and what remains an interface/fixture |
| Compiler boundary | Go-rendered semantic contracts are selected; no compiler/package exists | Define pure inputs/outputs/errors, deterministic ordering, obligation accounting, and generated-output non-authority |
| Evaluator controls | Ownership classes are reserved; physical partition and invalidation rules are not designed | Define exact public-control paths, candidate-denial enforcement, versioning, and independent review seam |
| Protected holdout | Required later; no external store, role, corpus, credential, or operator exists | Specify an external contract only; keep implementation and cases `NOT_RUN`/absent |
| Context and cost budgets | Normative categories exist; exact launch values are not frozen | Design explicit bounded defaults, measurement, exhaustion behavior, and justified applicability |
| Existing audit residuals | `AUD-W01-024` LOW and two INFO boundaries remain | Preserve them truthfully; do not silently remediate or relabel them as Wave 2 evidence |
| Hosted/external evidence | No remote; hosted CI and all provider/host/protected gates are `NOT_RUN` | No Wave 2 acceptance may depend on them |

No test, build, preparation, bootstrap, network, or external command is authorized or run for this two-file proposal. Historical exact-candidate verification is predecessor evidence, not verification of these proposal bytes.

## 8. Acceptance criteria

Wave 2 is eligible for its checkpoint only when one exact candidate proves all of the following:

1. All 29 `BL-002`/`BL-003` normative requirements are source-derived, uniquely owned, allocated, and traced to at least one Wave 2 acceptance test.
2. Versioned taxonomies and lifecycle contracts reject every specified invalid state/combination deterministically and preserve explicit blocked, stale, superseded, and non-applicable semantics.
3. Every safety-critical obligation and guardrail has one canonical meaning plus owner, criticality, enforcement locus, renderer requirement, grader, evidence rule, version, and deprecation status.
4. Workflow/profile/prompt/typed-output contracts contain the complete required semantic fields and cannot hide duplicated safety policy in prose.
5. The deterministic compiler rejects omitted, weakened, duplicated, unknown, and invented critical obligations for both stock-A0 and Controlled Client projections; it generates no host package in Wave 2.
6. Knowledge, budget, stopping, retry, loop, delegation, capability, degradation, and no-subagent contracts are machine-testable; exhaustion or missing capability never lowers a gate.
7. The public evaluation protocol is frozen before tuning and binds cases, truth labels, candidate/version identity, run/trial/environment/resource data, adjudication, confidence, cost/latency, and thresholds.
8. Deterministic graders cover computable safety outcomes; model-judge calibration and authority limits are explicit and cannot substitute for independent safety review.
9. A coverage map assigns every required semantic/safety axis to an owning feature and grader; seeded broken candidates fail for dropped obligations, fabricated evidence, stale approval, secret leakage, unsafe routing, and forbidden effects.
10. Candidate-denied evaluator controls are physically and policy separated from feature-owned fixtures; attempts to change frozen controls fail and invalidate affected evidence.
11. The protected-holdout/operator/exposure-response contract is complete but protected cases, labels, credentials, detailed results, and infrastructure remain outside the repository and `NOT_RUN`.
12. The Wave 2 phase/path/ownership transition preserves Wave 1 history, denies unknown/unowned paths, and has permanent positive/negative fixtures before semantic/evaluator paths pass.
13. Baseline and shadow local verification, exact candidate/evidence manifests, source digests, path closure, declared effects/limits, and a fresh independent read-only audit reproduce with no unresolved Blocker/Critical/High/Medium finding.
14. No dependency, host package, product command, prototype edit, protected asset, credential, provider/network/host action, mutation capability, stable claim, release, or deployment is introduced.

The detailed requirement and test mapping is normative in `L7-W02-SPEC-001`. No aggregate pass may waive a failed safety, authority, evaluator-control, or identity criterion.

## 9. Authority and prohibited effects

| Actor or source | Authority for Wave 2 |
|---|---|
| Accountable owner | May approve an exact proposal, design, or implementation action; each later effect requires a fresh exact decision |
| Semantic owner | May author semantic IDs/contracts only within an approved path set; cannot alter evaluator controls or self-approve |
| Evaluator-governance owner | May integrate separately approved public evaluator controls; cannot remediate the candidate or issue release authority |
| Feature fixture owner | May add ordinary fixtures in an assigned scope; cannot change frozen truth, oracles, adjudication, or thresholds |
| Harness/build integrator | May implement only the approved phase/path/build transition; cannot grant product or evaluator authority |
| Model, skill, subagent, repository, prompt, fixture, or tool output | Proposal/untrusted data only; cannot approve, lower risk, change truth, mint authority, or widen effects |
| Independent read-only reviewer | May audit exact candidate/evidence bytes and issue the scoped development verdict; cannot remediate, sign, promote, or deploy |
| Protected evaluator/operator | Remains external and unimplemented in Wave 2; candidate and remediator have no read/list/write authority |

Any write outside a later exact path/effect approval, any hidden/protected material in the repository, any dependency or external effect not separately authorized, or any attempt to let a candidate own its oracle terminates the action.

## 10. Recovery and interruption

For this proposal action, both files are new and uncommitted. If the pair is incomplete, inconsistent, overwritten, accompanied by another worktree change, or based on a changed predecessor, the safe result is `RECOVERY_REQUIRED`; no deletion, overwrite, staging, commit, merge, or cleanup is inferred.

The later design must define exact preimages, one-writer slices, fail-before-admission transition ordering, fixture-first or atomic path registration, candidate manifests, interruption points, compensating steps, and restoration of the last audited Wave 1 gate if a Wave 2 transition fails. Partial semantic/evaluator controls cannot become authoritative.

## 11. Observation and exposure

Wave 2 has no user-visible or production behavior, so a runtime feature flag is `NOT_APPLICABLE`. The applicable default-deny mechanism is the active phase/path/ownership gate plus the candidate-denied evaluator-control partition.

Observation is limited to deterministic local schema/compiler/grader results, candidate manifests, and audit records under a later approved effect envelope. There is no cohort, provider call, host session, protected evaluation, telemetry sink, deployment, exposure percentage, or production metric.

## 12. Stop and escalation conditions

Stop without design or implementation if:

- either proposal file is missing, inconsistent, or accompanied by a third worktree change;
- the predecessor commit/tree, Wave 1 audit digest, backlog scope, risk, authority, or effect boundary changes;
- any of the 29 owned requirement IDs is missing, duplicated, unknown, or assigned outside `BL-002`/`BL-003` without an approved impact decision;
- the design would admit semantic/evaluator paths before the successor policy and negative fixtures;
- policy is duplicated only in prompt prose or a renderer may silently omit/invent an obligation;
- candidate authors can change their evaluator, truth, adjudication, authorization, or thresholds;
- a protected case, label, credential, release policy, or external evaluator secret enters candidate-readable scope;
- a universal profile/context load exceeds the later approved budget or budget exhaustion weakens a gate;
- a dependency, host/provider/network/host-package/mutation/product path, or external effect is proposed without separate authority;
- Wave 1 historical bytes or audit conclusions are rewritten or relabeled; or
- an exact owner, reviewer, recovery path, verifier, or evidence source is unavailable.

The one next action after this proposal is the accountable owner's exact decision on these two files.

## 13. Compact R3 assurance case

| Element | Statement |
|---|---|
| Claim | Wave 2 can freeze provider-neutral semantics and public evaluation governance without creating host/product behavior or letting a candidate own its safety oracle. |
| Argument | One obligation source precedes rendering; public protocol freezes before tuning; evaluator-control ownership is candidate-denied; protected evaluation remains external; boundary fixtures precede path admission; all outputs bind the audited predecessor. |
| Current evidence | Approved requirements/backlog/architecture/technology/orchestration; audited Wave 1 GO tuple; reserved ownership classes; clean predecessor before this two-file action. |
| Required future evidence | Exact design, source-bound phase transition, schema/compiler/grader fixtures, seeded broken candidates, baseline/shadow verification, exact manifests, and fresh independent audit. |
| Assumptions | Same-user local Git/artifacts remain mutable; no remote/protected evaluator exists; selected serialization/dependency behavior is not inferred without implementation evidence. |
| Defeaters | Obligation drift, prose-only policy, dropped/invented obligation, invalid state accepted, candidate-controlled oracle, hidden-data exposure, unbounded context/cost, unexpected path/effect/dependency, stale predecessor, or material audit finding. |
| Residual risk | Host/provider behavior, protected evaluation, dependency conformance, controlled mutation, support, and release assurance remain later `NOT_RUN` gates. |
| Decision owner | Anup Pandey for this exact next gate only; persisted text remains `AP0`. |

## 14. Approval boundary

No approval is embedded in this proposal. The accountable owner may approve both exact proposal files, request revision, or reject them. Until a fresh exact decision is given, Wave 2 design and all implementation are blocked.
