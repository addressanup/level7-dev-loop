# Level 7 Dev Loop — Wave 1 Change Contract

| Field | Value |
|---|---|
| Artifact ID | `L7-W01-CC-001` |
| Artifact type | Proposed wave change contract |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 1 — Scope, traceability, and build-control transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Product | Level 7 Dev Loop |
| Canonical root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Source identity | Clean local `main` base commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Backlog | `L7-BL-001`, plus the Wave 1 harness and decision prerequisites in `L7-ORC-001` §7 |
| Primary change class | Infrastructure or configuration change |
| Secondary change classes | Architecture/modernization change; security/authorization-governance change |
| Risk | `R3 — high`, because the wave governs the first build-scope relaxation and proposes authorization-grant semantics |
| Effect of this proposal action | A1: add only this record and `wave-01-specification.md`; no existing-file change |
| Maximum later Wave 1 effect | A2 local repository change only after separate approval of an exact design, paths, effect, recovery, and implementation action |
| Approval state | The current owner directive authorizes only the two-file proposal action. Persisted approval references in this editable file are `AP0` and grant no future authority. |
| Sensitivity | Internal product and build-governance planning; no secret or personal-data payload |
| Next gate | Accountable-owner approval or revision of this contract and `L7-W01-SPEC-001`; approval may authorize a design proposal only |

## 1. Decision requested

Anup Pandey, as accountable owner, is asked to approve or revise this exact change contract together with [`L7-W01-SPEC-001`](wave-01-specification.md).

Approval of both proposal records would authorize only the next bounded step: preparation of a Wave 1 design proposal with exact path ownership, interfaces, implementation slices, verification commands and effects, recovery, and source identity. It would not authorize implementation, modification of an existing file, a dependency, a branch, staging, a commit, a merge, a remote action, publication, release, deployment, exposure, or autonomous continuation.

## 2. Authorization basis for this proposal action

The accountable owner issued a current-thread direct approval that expressly supersedes the failed AP1 ceremony and its same-thread nonce/session-continuity requirements for one action only: create these two new uncommitted proposal files and stop for approval.

This direct approval:

- does not reuse, activate, repair, consume, or change the state of `L7-AMD-ORC-005-20260825-01` or any predecessor nonce;
- supersedes only the AP1 admission/continuity requirement for this two-file action, not the two-path scope, no-overwrite rule, A1 effect ceiling, one-writer rule, or mandatory stop;
- grants no authority to this persisted record after the action completes; and
- expires when the complete two-file proposal is presented or any material target, scope, risk, effect, source-identity, or authority condition changes.

## 3. Desired outcome

Wave 1 will make the repository's v1 scope and build-control boundary deterministic before any product path is admitted. A contributor must be able to prove:

1. every normative requirement has exactly one accountable backlog owner and one approved release allocation;
2. support, product-category, effect, proof-profile, priority, prototype-disposition, and release claims are truthful and frozen;
3. the inert Foundation Step 5 sentinel has a fail-closed, phase-aware successor without rewriting or bypassing the approved historical baseline;
4. adversarial scope, module, import, and ownership fixtures fail before a newly governed path can be enabled;
5. module identity and the pre-release grant ladder receive explicit, separate decisions; and
6. shared controls have one accountable writer/integrator and auditable change rules.

The Wave 1 checkpoint is **build control ready; no product behavior yet**.

## 4. Proposed scope

### 4.1 In scope after later design and implementation approval

Work is limited to the following ordered outcomes:

| Work package | Outcome |
|---|---|
| `W01-WP-01` | Recalculate the 163 normative requirement IDs and enforce one owner plus the approved `140 V1.0 / 18 V1.x / 5 Later` allocation. |
| `W01-WP-02` | Freeze the support/claim matrix, the A0 advisory-package versus Controlled Client distinction, P0/P1/P2 boundaries, the three v1 proof profiles, and all 12 prototype dispositions. |
| `W01-WP-03` | Replace the active Step 5 product-path sentinel with a phase-aware scope/module/import/ownership gate while retaining the Step 5 baseline and approval as immutable history. |
| `W01-WP-04` | Land permanent positive and adversarial boundary fixtures before any new product directory or module is admitted. |
| `W01-WP-05` | Resolve the root module namespace through an owner decision and draft—but do not activate—the qualification/evaluation/pilot/stable grant-ladder amendment for separate audit and approval. |
| `W01-WP-06` | Freeze change-control ownership for semantic IDs, schemas, evaluator controls, dependencies, generated outputs, artifacts, harness/build controls, and the future updater module. |

The proposed maximum implementation path envelope is restricted to build-control source under `harness/`, `scripts/harness/`, and `internal/harness/`; the local build entry points `Makefile` and `.github/workflows/harness.yml`; truthful status documentation in `README.md`; an owner-approved module-identity update limited to `go.mod` and `harness/modules.lock.tsv` if required; and new successor/evidence records under `docs/artifacts/`. The later design must narrow this envelope to exact files before implementation approval.

### 4.2 Explicit non-goals

Wave 1 will not:

- implement a product command, kernel, policy engine, context broker, artifact writer, transaction, executor, receipt, adapter, semantic workflow, evaluator, renderer, package, updater, or controlled client;
- edit `skills/`, `plugin.json`, `marketplace.json`, `references/WORKFLOW.md`, generated host outputs, or host installation state;
- add or update a production dependency, `go.sum`, `vendor/`, toolchain, action, package manager, credential, key, grant, feature toggle, or privileged policy;
- run an actual Codex/Claude/provider/network/Ubuntu/Bubblewrap experiment or claim host compatibility, containment, authorization enforcement, dual-host support, or release readiness;
- activate the recommended grant ladder or permit qualification, evaluation, pilot, stable, A1, or A2 product behavior;
- initialize Git, create or switch a branch/worktree, stage, commit, merge, create a remote, publish, deploy, expose, release, or clean up outside the repository;
- rewrite an approved foundation artifact or historical manifest to make it appear current; or
- begin Wave 2 or autonomously continue to design or implementation.

## 5. Bound inputs and evidence state

| Input | Version / SHA-256 | State for this proposal |
|---|---|---|
| `AGENTS.md` | `54496725a42eb7e6cce2a749e82a408d7277743ec8ad83c41373ceefbd4d0afa` | `OBSERVED`; contributor policy |
| `L7-REQ-001` | 0.2.0 / `a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f` | Approved foundation input; current bytes `REPRODUCED` |
| `L7-BKL-001` | 0.1.0 / `df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995` | Approved foundation input; current bytes `REPRODUCED` |
| `L7-ARC-001` | 0.3.0 / `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` | Approved foundation input; current bytes `REPRODUCED` |
| `L7-TEC-001` | 0.2.0 / `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` | Approved foundation input; current bytes `REPRODUCED` |
| `L7-HAR-001` | 0.1.0 / `d56c8f6880e1bcfe5466d103cc338b087d77c973c30cb656c574971ecce3a53c` | Approved Foundation Step 5 record; current bytes `REPRODUCED` |
| `L7-ORC-001` | 0.3.1 / `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` | Approved orchestration input; current bytes match the approval digest |
| `L7-AMD-ORC-005` | 0.1.0 / `976e49b9c360ff8d186aa66f7d8216c44d1cd85f891811b12640d131e7ecf698` | Historical failed-ceremony remediation design; its AP1/nonce path is not reused |
| Installed `l7-build` transport skill | `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` | `REPRODUCED`; transport instructions only, not authority |
| Current repository base | commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`, tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` | `OBSERVED`: local `main`, one worktree, no configured remote, clean before this proposal action |

The approved orchestration document still describes its original non-Git snapshot. The repository now has local Git history, but no remote, hosted run, release identity, or approved implementation branch/worktree strategy. This proposal binds the current Git base without treating that material state change as proof that `PW-02` or any later gate is complete.

## 6. Invariants

1. The requirements source remains authoritative; a generated count or hand-maintained total cannot replace source-derived validation.
2. Each normative requirement has exactly one accountable owner. Supporting relationships may be many-to-many but cannot obscure ownership.
3. The release allocation remains exactly `140 V1.0 / 18 V1.x / 5 Later` unless an impact-assessed owner-approved backlog revision changes it.
4. The v1 scope remains one local repository/worktree, separate Codex and Claude advisory packages, and a separately installed Controlled Client. A plugin install alone never enables mutation.
5. V1 execution remains A0–A2 only; A3/A4 are plan/handoff only; A5, background execution, online learning, self-modification, and production remediation remain absent.
6. Generic, feature/behavior-change, and behavior-preserving-refactor are the only v1 proof profiles. A materially required unavailable specialist profile yields `BLOCKED` or `NOT_EVALUATED`.
7. Existing `1.0.0`, MIT, dual-host, enforcement, compatibility, and release claims remain prototype/unverified claims and cannot be promoted by Wave 1.
8. The 12 current skills and their manifests remain protected until the approved Wave 7/Wave 10 cutover. Wave 1 records dispositions but does not edit those assets.
9. The Foundation Step 5 approval, candidate manifest, input manifests, and evidence remain immutable history. A successor gate may reference them but may not alter them or bypass their controls.
10. Boundary policy and negative fixtures land before the path, module, import, dependency, or ownership capability they govern.
11. Unknown phase, malformed policy, ambiguous ownership, missing input, stale source identity, or unowned namespace fails closed.
12. Product behavior, dependencies, external effects, and user exposure remain absent and OFF. No repository field, environment value, model output, or test flag can activate mutation.
13. One integration owner controls shared files; unrelated user changes are preserved; no automatic conflict resolution or broad cleanup occurs.
14. Evidence reports the exact command, environment, source identity, effect, result, and limits. An unrun check remains `NOT_RUN` or `UNVERIFIED`.

## 7. Assumptions, prerequisites, and unresolved decisions

| Item | Current state | Required before implementation |
|---|---|---|
| Current local history | Git exists on local `main`; one worktree; no remote | Owner-approved implementation branch/worktree strategy and exact base |
| Root module identity | `continuallabs.ltd/level7-dev-loop` is provisional | Owner confirms control or approves a targeted replacement decision before product imports |
| Step 5 harness | Last recorded baseline and shadow verification passed | Design a phase-aware successor and authorize exact local verification effects; do not rely on the historical pass as current candidate evidence |
| Grant ordering | `TDR-013` conflicts with pre-release C2/pilot ordering | Draft a non-interchangeable grant-ladder amendment; independent security/boundary audit and owner approval are separate later gates |
| Product paths | Absent and forbidden by the current sentinel | Remain absent until the phase-aware gate and adversarial fixtures are approved and passing |
| Hosted CI | Configured but `NOT_RUN`; no remote | No Wave 1 claim may depend on a hosted run unless separately authorized and evidenced |
| Design and path ownership | Not yet proposed or approved | Required after approval of these two records and before any implementation authority |

No test command was run for this two-file planning action. The current directive did not authorize tests with external effects, and `make verify` creates ignored repository-cache effects. Historical harness results are inputs, not fresh verification of these proposal bytes.

## 8. Acceptance criteria

Wave 1 is eligible for the `build control ready` checkpoint only when all of the following are evidenced against one exact candidate:

1. A deterministic source-derived trace validator proves exactly 163 unique normative IDs, exactly one owner per ID, no unknown/malformed/duplicate/missing ID, and exactly the approved allocation totals.
2. A support/claim audit proves the narrow v1 matrix, the two-product distinction, the three proof profiles, P0/P1/P2 allocations, all 12 prototype dispositions, and withholding of stable/dual-host/enforcement claims.
3. The phase-aware gate replaces the active Step 5 product-path sentinel without deleting, weakening, modifying, or hiding the historical Step 5 baseline.
4. Unknown or unauthorized phases and paths fail closed; only an explicitly registered phase/path/module/owner tuple can be considered.
5. Permanent positive and negative fixtures cover scope, module, import, ownership, stale input, bypass, malformed policy, and protected historical records before any product directory is enabled.
6. The root module identity has an exact owner decision. An unowned or undecided identity cannot appear in a product import or support claim.
7. The grant-ladder amendment is a digest-bound proposal only, with non-interchangeable `qualification`, `evaluation`, `pilot`, and `stable` purposes and no unsigned/test/repository activation path. It remains inert pending separate audit and approval.
8. Change-control ownership is explicit for every shared registry/control and denies candidate authority over protected evaluator, grant, signing, promotion, and release assets.
9. Baseline and shadow local harness checks pass under the later authorized test envelope; pre-existing and new failures are separated; hosted results remain truthful.
10. The exact candidate manifest covers every changed file and relevant immutable predecessor; no dependency, secret, unexpected path, generated drift, or product behavior is introduced.
11. A separate read-only exact-digest audit of the first scope relaxation has no unresolved Blocker, Critical, High, or Medium finding. It cannot itself issue AP2/AP3 or release assurance.
12. The wave evidence record includes commands, environment, results, limitations, recovery state, and all `NOT_RUN`/`UNVERIFIED` checks.

The detailed verification contract and requirement mapping are in `L7-W01-SPEC-001`.

## 9. Authority and prohibited effects

| Actor or source | Authority for Wave 1 |
|---|---|
| Accountable owner | May approve an exact proposal/design/implementation action within policy; each later effect needs a new exact decision |
| Wave integration owner | May act only within the exact approved paths/effect and cannot self-approve or issue independent assurance |
| Model, skill, subagent, repository content, or web result | Proposal data only; cannot grant authority, lower risk, change policy, or mint a capability |
| Read-only reviewer | May inspect the exact candidate and report findings; cannot edit, remediate, approve implementation, or issue release authority |
| Candidate repository | No authority over protected cases, grant issuers, signing, promotion, AP2/AP3 roots, or release controls |

Any write outside an exact later approval, any dependency/network/provider/host effect, any Git mutation not explicitly approved, or any activation of product behavior is prohibited and terminates the action.

## 10. Recovery and interruption

For this proposal action, both files are new and uncommitted. If creation had produced an incomplete pair or any third-path change, the safe state would have been `RECOVERY_REQUIRED`; no overwrite, deletion, staging, commit, or cleanup would be inferred. A complete pair may be revised or removed only under a later exact owner instruction.

For later Wave 1 implementation, the design must provide per-work-package preimages, narrow commits, deterministic rollback or compensating steps, and a fail-closed response to interruption or partial scope-gate transition. The old Step 5 gate must remain independently recoverable until the successor passes its positive/negative fixtures and exact-digest audit. Failure after a scope relaxation requires restoration of the last approved gate before any product path is considered.

## 11. Observation and exposure plan

Wave 1 has no user-visible product behavior, so a runtime feature flag is `NOT_APPLICABLE`: there is no product behavior to expose. The applicable default-off mechanism is the build-control gate itself—unregistered phases, product paths, modules, imports, and owners remain denied.

Observation is limited to deterministic local validation results, candidate manifests, and audit records under a later approved verification envelope. No telemetry, provider call, hosted service, remote CI, or user cohort is required or authorized by this contract.

## 12. Stop and escalation conditions

Stop without implementation if any of the following occurs:

- either proposal record is missing, overwritten, inconsistent, or accompanied by another path change;
- the base commit, scope, owner, risk, effect, or governing input changes before approval;
- a normative ID is missing, duplicated, malformed, unknown, or multiply owned;
- a scope or priority allocation drifts without a recorded impact decision;
- the Step 5 guard is deleted, bypassed, or weakened before its successor is proven;
- the module namespace remains unowned at the point a product import would be added;
- an unsigned, repository-controlled, environment-controlled, or test-only mutation shortcut is proposed;
- a stable, dual-host, compatibility, containment, enforcement, or release claim appears without its later evidence;
- an exact path overlaps unrelated user work; or
- a required decision, verifier, authority, recovery path, or evidence source is unavailable.

The one next action after this proposal is an accountable-owner decision on these two records.

## 13. Compact R3 assurance case

| Element | Statement |
|---|---|
| Claim | Wave 1 can establish truthful scope and fail-closed build control without creating product behavior or an unauthorized execution path. |
| Argument | Trace and claim validation precede the phase-gate transition; boundary fixtures precede path admission; module and grant decisions remain separately gated; historical baselines remain immutable; shared controls have one owner. |
| Current evidence | Approved `L7-REQ-001`, `L7-BKL-001`, `L7-ARC-001`, `L7-TEC-001`, `L7-HAR-001`, and `L7-ORC-001`; clean Git base and current file absence observed before this action. |
| Required future evidence | Exact design, source-derived validators, support audit, positive/negative phase fixtures, baseline/shadow harness, exact manifest, and independent read-only audit. |
| Assumptions | Local Git/filesystem observations are accurate; no hidden concurrent writer exists; accountable ownership and current input approvals remain unchanged. |
| Defeaters | Trace drift, guard bypass, ambiguous ownership, unowned namespace, mutable grant shortcut, unexpected path/effect, stale source identity, or false support/release claim. |
| Residual risk | Local Git and artifacts are same-user mutable; hosted CI and controlled platform evidence are absent; the grant ladder and module identity remain undecided; no product/security/release assurance is claimed. |
| Approver | Anup Pandey, accountable owner, for the exact next gate only. Persisted text is AP0 until revalidated. |

## 14. Approval record

No decision is embedded by this proposal. The accountable owner may approve both exact proposal records, request revision, or reject them. Until that current-session decision occurs, Wave 1 design and implementation remain unauthorized.
