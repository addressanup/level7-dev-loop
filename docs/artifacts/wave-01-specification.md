# Level 7 Dev Loop — Wave 1 Specification

| Field | Value |
|---|---|
| Artifact ID | `L7-W01-SPEC-001` |
| Artifact type | Proposed implementation-independent wave specification |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 1 — Scope, traceability, and build-control transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Change contract | [`L7-W01-CC-001`](wave-01-change-contract.md) 0.1.0 |
| Source identity | Clean local `main` base commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Requirement owner | `L7-BL-001`, with harness/build integration ownership for Wave 1 prerequisites |
| Risk / maximum later effect | `R3 — high` / A2 local repository change only after separate exact approval |
| Current effect | A1 addition of this proposal record; no product or existing-file effect |
| Approval state | Proposed/AP0 when persisted; the current direct approval ends after presentation of the two-file proposal |
| Next gate | Approve or revise this specification and `L7-W01-CC-001`; an approval may authorize a design proposal only |

## 1. Purpose and success statement

This specification defines what Wave 1 must accomplish, how completion will be judged, and which conditions remain blocked. It intentionally does not select the final file layout, script structure, parser design, commit sequence, module name, grant issuer, or implementation mechanism; those belong in a later exact design proposal.

Wave 1 succeeds when one exact local candidate deterministically proves the approved requirement allocation and support boundary, transitions the inert Step 5 scope sentinel to a fail-closed phase-aware build-control gate with permanent adversarial fixtures, records the required ownership and decision boundaries, leaves the complete baseline/shadow harness green, and receives the required independent read-only audit. Success creates no product behavior and no support or release claim.

## 2. Normative language and precedence

`SHALL`, `SHALL NOT`, `SHOULD`, and `MAY` carry the meanings defined by `L7-REQ-001`. If this proposal conflicts with an enforced platform boundary, a non-waivable Level 7 safety invariant, an authenticated organization policy, or an approved foundation input, the higher-authority source controls and the affected work is `BLOCKED` pending an authorized revision.

The source of truth order for Wave 1 is:

1. exact owner authority for the current action;
2. approved requirements and release allocation;
3. approved architecture, technology, harness, and orchestration decisions;
4. the later owner-approved Wave 1 change contract, specification, and design; and
5. implementation, generated indexes, reports, model output, and repository prose as derived/untrusted evidence candidates.

No lower layer may silently widen scope or weaken a higher layer.

## 3. Scope and path contract

### 3.1 Maximum later implementation envelope

The later design may select exact files only within this maximum envelope:

| Path or path family | Permitted Wave 1 purpose | Constraint |
|---|---|---|
| `harness/` | Phase, module, import, ownership, support/claim, trace, and fixture control data | Existing approved manifests remain immutable history; new successor records must be distinguishable |
| `scripts/harness/` | Deterministic offline validators and phase-aware gate entry points | No network, credential, provider, package-manager, or product execution |
| `internal/harness/` | Harness-only positive/negative test support | Cannot become a product import or runtime dependency |
| `Makefile` | Invoke the approved local checks | Exact effects and cache paths must be declared in the design |
| `.github/workflows/harness.yml` | Invoke equivalent configured baseline/shadow checks | No new secret or write permission; hosted execution remains separately evidenced |
| `README.md` | Truthful current status and verification instructions | Cannot claim support, enforcement, compatibility, or release |
| `go.mod`, `harness/modules.lock.tsv` | Exact root-module identity transition only if separately decided | No production dependency; no updater activation; no unowned namespace |
| `docs/artifacts/` | New Wave 1 decision, candidate, evidence, audit, and successor records | Do not rewrite approved historical artifacts or manifests |

The design SHALL narrow this envelope to an exact allowlist for every work package. Any other tracked path is out of scope unless a revised contract/specification receives fresh owner approval.

### 3.2 Protected and forbidden paths

Wave 1 SHALL NOT modify `skills/`, `plugin.json`, `marketplace.json`, `references/WORKFLOW.md`, `.env.example`, `.go-version`, generated host packages, protected evaluation assets, signing/promotion assets, host configuration, or anything outside the canonical repository root. Product paths named by the current Step 5 sentinel remain absent until the successor gate and its fixtures are approved and passing; Wave 1 does not otherwise use them.

Git metadata changes, branches/worktrees, staging, commits, merges, remotes, tags, and hosted actions are not authorized by this specification. A later implementation approval must separately bind any required local Git effects.

## 4. Ordered work-package requirements

The six work packages are serial at integration. Read-only research may be parallel only if it does not create a correctness dependency or file mutation. One integration owner writes shared controls.

### 4.1 `W01-WP-01` — Normative trace and allocation validator

#### Functional requirements

1. The validator SHALL derive normative requirement IDs from the authoritative requirement-definition tables in `L7-REQ-001`, not from the summary total.
2. It SHALL recognize only the approved ID grammar, expand approved ranges deterministically, and reject malformed, reversed, overlapping, duplicated, or unknown IDs.
3. It SHALL derive the accountable-owner and allocation map from the approved backlog ownership source and prove exactly one owner plus one `V1.0`, `V1.x`, or `Later` allocation for every normative ID.
4. It SHALL prove totals of exactly 163 requirements, with `140 V1.0`, `18 V1.x`, and `5 Later` for the approved candidate.
5. Supporting backlog relationships MAY be many-to-many but SHALL NOT count as accountable ownership.
6. Output SHALL be deterministic, concise, and actionable: success reports source digests and totals; failure identifies each missing, duplicate, malformed, unknown, multiply owned, or unallocated ID.
7. The validator SHALL operate offline and SHALL NOT rewrite its sources or repair discrepancies automatically.

#### Required fixtures

- exact approved source;
- one missing ID;
- one duplicated definition;
- one malformed and one unknown ID;
- overlapping or duplicate owner ranges;
- zero owners and two owners;
- each allocation-total drift; and
- a changed hard-coded summary with unchanged source definitions, proving the source-derived result controls.

### 4.2 `W01-WP-02` — Support, claim, priority, and prototype freeze

#### Functional requirements

1. A versioned Wave 1 support/claim record SHALL distinguish current prototype state, intended v1 scope, development evidence, unsupported behavior, and release-blocking proof.
2. The v1 matrix SHALL state one local repository/worktree; separate Codex CLI and Claude Code advisory packages; a separately installed Controlled Client; A0–A2 execution only at the later proven ceiling; A3/A4 plan/handoff only; and no A5/background/self-modifying behavior.
3. Plugin installation SHALL be explicitly insufficient for mutation authority.
4. Launch proof profiles SHALL be exactly generic, feature/behavior change, and behavior-preserving refactor. A materially required unavailable profile SHALL yield `BLOCKED` or `NOT_EVALUATED`.
5. P0/P1/P2 allocations and every scope/priority change rule SHALL remain aligned with `L7-BKL-001`.
6. Each of the 12 current user-invocable prototype skills SHALL have exactly one proposed disposition: conform, replace, deprecate, or exclude, with target owner and cutover wave. Wave 1 SHALL NOT edit the skills.
7. Existing `1.0.0`, MIT, dual-host, compatibility, enforcement, and workflow claims SHALL be classified as prototype/unverified or withheld as applicable. Their mere presence in protected prototype metadata SHALL NOT become a stable claim.
8. The audit SHALL reject wording that implies product implementation, security qualification, controlled mutation, actual-host conformance, release, deployment, or support without the later required evidence.

#### Required fixtures

- approved truthful current status;
- plugin-install-implies-mutation claim;
- Codex evidence incorrectly promoting Claude or the reverse;
- generic proof incorrectly satisfying a missing specialist profile;
- A3/A4/A5 execution claim;
- prototype `1.0.0` metadata promoted as stable;
- missing or duplicate prototype disposition; and
- priority/allocation drift without impact approval.

### 4.3 `W01-WP-03` — Phase-aware build-control gate

#### Functional requirements

1. The successor SHALL replace, not delete or bypass, the current Foundation Step 5 product-path sentinel.
2. The approved Step 5 manifests, approvals, and evidence SHALL remain byte-preserved historical inputs. A current successor manifest SHALL bind the new candidate without expecting superseded live-path bytes to remain unchanged forever.
3. The gate SHALL admit only an explicit registered tuple of phase, path, module, import boundary, owner, and predecessor identity. Missing or ambiguous fields fail closed.
4. Unknown phase, unknown path, reserved module, unowned namespace, unowned shared file, stale predecessor, malformed registry, and forbidden import SHALL return nonzero with a stable rule identifier.
5. The gate SHALL preserve all applicable Step 5 protections: exact baseline/shadow pins, read-only CI permissions, no secret consumption, digest-pinned actions, zero unauthorized production dependencies, updater reservation, pure-boundary rules, `unsafe` prohibition, and test-harness isolation.
6. A phase transition SHALL NOT itself create a product path, enable mutation, change an effect ceiling, activate a grant, or create a support claim.
7. Boundary policy SHALL be in effect before the governed path/module/dependency exists. A check that first discovers a violation after admitting the capability is nonconforming.
8. The gate and its control data SHALL be deterministic, offline, diff-reviewable, and usable on the pinned baseline and shadow toolchains where Go inspection is required.
9. Failure SHALL leave the current stricter gate effective. No automatic fallback to a broader phase is permitted.

#### Transition compatibility

The design SHALL preserve an explicit way to verify the immutable Foundation Step 5 baseline and a distinct way to verify the Wave 1 successor. The two results SHALL not be conflated. The candidate manifest SHALL show the predecessor, the exact successor delta, and which gate is authoritative for the candidate.

### 4.4 `W01-WP-04` — Permanent boundary fixtures

Before any new product directory, module, dependency, or ownership class can be admitted, permanent fixtures SHALL cover at least:

| Fixture class | Required behavior |
|---|---|
| Authorized current phase | Known harness-only paths pass under the exact predecessor/successor binding |
| Unknown or missing phase | Fail closed |
| Reserved or wrong module | Fail with the applicable module-boundary rule |
| Unauthorized product path | Fail before package inspection can treat the path as admitted |
| Forbidden direct/transitive import | Fail with the stable boundary rule and path |
| External-module detour | Fail when a pure closure reaches an undeclared module |
| Test-harness import from product | Fail |
| `unsafe` import | Fail |
| Stale or mismatched predecessor/manifest | Fail; do not advance phase |
| Missing, duplicate, or unknown owner | Fail |
| Malformed/duplicate registry entry | Fail |
| Historical Step 5 record alteration | Fail |
| Updater path before its separate module gate | Fail |
| Attempted unsigned/test/env/repository mutation shortcut | Fail or remain structurally unavailable |

Fixtures SHALL be permanent repository tests, not transient probes used only during development. Negative fixtures SHALL prove that the control detects a deliberately broken candidate rather than only testing a success path.

### 4.5 `W01-WP-05` — Module and grant-ladder decisions

#### Root module identity

1. The accountable owner SHALL either confirm control of `continuallabs.ltd/level7-dev-loop` for the intended module use or approve a targeted technology decision naming a replacement.
2. Before that decision, no product import or support/publication claim may rely on the provisional namespace.
3. A replacement SHALL include impact on import paths, manifests, module registry, generated outputs, compatibility, and future updater separation. It SHALL NOT be performed as an incidental search-and-replace.
4. The updater module remains `reserved` with identity `UNSET` until Wave 10's separate module-aware design and audit.

#### Grant-ladder amendment proposal

1. Wave 1 SHALL draft an amendment proposal; it SHALL NOT issue, sign, install, verify as active, or consume any grant.
2. The amendment SHALL define distinct `qualification`, `evaluation`, `pilot`, and `stable` grant kinds with non-interchangeable audience, purpose, issuer/trust policy, candidate, host/model/platform tuple, target class/root, effect ceiling, issue/expiry, nonce, revocation, and policy digest bindings.
3. `qualification` SHALL target only Level-7-owned disposable synthetic fixture roots and yield development evidence only.
4. `evaluation` SHALL target only fresh evaluator-owned isolated case roots and remain unavailable to candidate authors.
5. `pilot` SHALL require prior C5 controlled conformance plus an exact consented cohort/root/expiry/revocation/observation authorization and yield C6 evidence only.
6. `stable` SHALL require C7 independent `GO` plus exact release authorization and root-owned local policy.
7. No repository boolean, environment variable, test mode, conversational phrase, unsigned object, or lower grant kind may verify as another grant or activate mutation.
8. Qualification/evaluation mode SHALL be structurally incapable of targeting an arbitrary user repository or leaving a dormant production bypass.
9. The proposal SHALL state the exact conflict with `TDR-013`, the clauses it would supersede, migration/compatibility effects, kill/degrade behavior, and why C2/pilot remain blocked if the amendment cannot be independently assured.
10. Activation requires a new digest-bound technology/backlog revision, independent security/boundary audit, and separate accountable-owner approval; none is implied by Wave 1 completion.

### 4.6 `W01-WP-06` — Change-control ownership

A versioned ownership record SHALL assign one accountable writer/integrator and one review rule to each shared control class:

| Control class | Required ownership rule |
|---|---|
| Requirements IDs and release allocation | Backlog/scope owner proposes; source-derived validator detects drift; owner approves material change |
| Semantic IDs, taxonomies, schemas, prompt/profile contracts | Future `BL-002` semantic owner; Wave 1 may only reserve the rule |
| Public evaluator protocol, truth schema, oracles, thresholds, coverage index | Future `BL-003` evaluator-governance owner; candidate authors have no write authority |
| Harness, phase/module/import policy, tool/dependency/action locks, CI, `Makefile` | Harness/build integrator |
| Feature public fixtures | Feature owner in a disjoint backlog-ID scope; evaluator owner integrates frozen indexes |
| Dependencies and module identities | Harness/build integrator under an approved technology decision and dependency evidence |
| Generated files/packages/indexes | Generator/integration owner only; no hand edit or reverse promotion |
| `docs/artifacts/` wave record and current status | Wave integration owner; approved historical records remain immutable |
| Existing skills/manifests/`WORKFLOW.md` | Protected until their Wave 7/Wave 10 cutover owner acts |
| Updater module and privileged channel | Future Wave 10 updater owner in a separate module; no root/core import |
| Protected cases, grants, signing, promotion, AP2/AP3 roots | Outside candidate repository and agent authority |

The ownership validator SHALL reject missing, duplicate, conflicting, or unauthorized ownership and SHALL make shared-file proposals serial through the named integrator.

## 5. External behavior and interfaces

Wave 1 introduces no product runtime, CLI, API, schema consumed by users, host package, provider interaction, network protocol, or controlled mutation interface.

Its only externally observable developer interfaces after later implementation are deterministic local validation commands and their exit status. The later design SHALL define exact command names and output schemas with these minimum semantics:

| Result | Required behavior |
|---|---|
| Success | Exit zero; identify validator/gate version, exact relevant source digests, phase, and bounded totals/result |
| Policy or validation failure | Nonzero; stable rule ID; exact offending source/path/ID; no mutation or auto-repair |
| Missing capability/input | Nonzero and `BLOCKED`/`NOT_EVALUATED` as appropriate; never a pass |
| Unexpected internal error | Nonzero; bounded secret-safe diagnostic; stricter gate remains effective |

Output SHALL be deterministic under the declared locale/timezone and SHALL not include secrets, raw environment values, credentials, hidden chain-of-thought, or terminal-control sequences.

## 6. Quality and nonfunctional requirements

### 6.1 Safety and security

- Repository content, manifests, docs, fixtures, model output, and tool output are untrusted inputs to validation.
- Validators SHALL use rooted repository-relative paths and reject path escape, unexpected symlink/special-node use, or out-of-envelope targets relevant to their checks.
- No validator may grant approval, lower risk, activate a grant, create a product capability, or issue a release verdict.
- Checks SHALL run offline by default with no credential access, provider call, package fetch, telemetry, or external sink.
- Diagnostics and fixtures SHALL contain no secret; seeded secret-shaped data, if later approved for a test, must remain synthetic and repository-local.

### 6.2 Determinism and portability

- Identical admitted inputs SHALL produce identical semantic results on the pinned baseline and shadow toolchains, aside from explicitly declared environment metadata.
- Parsing SHALL not depend on locale-specific collation, current time, network data, Git author identity, model output, or unordered filesystem traversal.
- Host-specific behavior is outside Wave 1; no macOS result may be promoted to the controlled Ubuntu or dual-host support matrices.

### 6.3 Performance and boundedness

- Each validator SHALL declare its scanned roots, file-count/byte/time limits, and failure behavior in the later design.
- Exceeding a bound SHALL return a bounded blocking result, not silently truncate and pass.
- No long-running/background process, daemon, watcher, or retry loop is permitted.

### 6.4 Accessibility and usability

- Diagnostics SHALL lead with the decision (`PASS`, `BLOCKED`, `NOT_EVALUATED`, or failure), identify the applicable rule, and give one recovery action.
- Results SHALL not rely on color, animation, terminal width, or icon-only meaning.
- Aggregate success SHALL not hide an individual safety failure.

### 6.5 Compatibility and preservation

- Existing developer harness commands SHALL either retain their documented meaning or receive an explicit, documented successor path and migration note.
- Historical Step 5 verification SHALL remain reproducible from its frozen inputs to the extent the already approved harness supports it.
- Unrelated user work, approved artifacts, prototype assets, and ignored ambient state SHALL not be overwritten or cleaned.

## 7. Verification specification

### 7.1 Verification layers

| Layer | Required evidence |
|---|---|
| Static/source | Formatting, shell syntax, registry/schema structure, exact path and dependency policy |
| Unit/deterministic | Range/ID parsing, allocation reduction, claim classification, phase admission, ownership resolution |
| Positive integration | Exact approved sources and registered Wave 1 harness-only candidate pass |
| Negative/adversarial | Every fixture class in §§4.1–4.4 fails for the intended stable rule |
| Degraded/interruption | Missing tool/input, stale digest, malformed registry, partial candidate, and bounded-resource failure remain fail closed |
| Baseline/shadow | Complete local baseline Go 1.26.7 verification blocks; Go 1.27.0 shadow is reported separately and cannot mask baseline failure |
| Manifest | One exact candidate manifest covers all changed files and immutable predecessors needed for the claim |
| Independent review | Read-only exact-digest audit of the first scope relaxation and its negative fixtures |

### 7.2 Test-effect boundary

No test was executed while authoring this proposal. Before any later test run, the owner-approved design/implementation action SHALL declare exact commands, repository cache/temp/log effects, network state, credentials, cleanup policy, and expected duration. Tests SHALL not use an external effect unless that effect receives a separate exact authorization.

The intended Wave 1 test envelope is local and offline: source reads plus declared writes under ignored repository-scoped cache/temp locations. It does not include bootstrap downloads, dependency installation, hosted CI, a provider/model call, host installation, a remote, a root operation, or cleanup of ambient host files.

### 7.3 Evidence rules

Each executed check SHALL record method/command, candidate commit/tree and dirty-state identity, tool version, OS/architecture, declared cache/effect locations, timestamp, result, relevant bounded output, producer, evidence state, and reproducibility limits. Planned or blocked checks remain `NOT_RUN`, `NOT_EVALUATED`, or `UNVERIFIED`; they are never inferred from file presence or a prior candidate.

## 8. Acceptance criteria and traceability

| Acceptance ID | Requirement | Source |
|---|---|---|
| `W01-AC-001` | Exactly 163 source-derived normative IDs have one accountable owner and allocation; totals are `140/18/5`. | `L7-BL-001` AC1; `L7-ORC-001` Wave 1 WP1 |
| `W01-AC-002` | The narrow v1 support matrix, two-product distinction, A0–A2/A3–A5 boundary, and three proof profiles are frozen truthfully. | `L7-BL-001` AC2–3 |
| `W01-AC-003` | All 12 prototype skills have one disposition and protected bytes remain unchanged. | `L7-BL-001` AC4; `L7-ORC-001` §§4.2, 8.4 |
| `W01-AC-004` | Existing prototype version/compatibility/enforcement claims are withheld from stable promotion. | `L7-BL-001` AC5 |
| `W01-AC-005` | Scope/priority changes require an impact diff and accountable approval; no metric/date waives a safety prerequisite. | `L7-BL-001` AC6 |
| `W01-AC-006` | The phase-aware successor preserves historical Step 5 evidence and fails closed on unknown/malformed/stale/unowned input. | `L7-ORC-001` `PW-01`, `PW-04`, Wave 1 WP3 |
| `W01-AC-007` | Permanent positive/negative fixtures land before any new governed capability. | `L7-ORC-001` Wave 1 WP4 and integration gate |
| `W01-AC-008` | Module identity receives an exact decision before product imports; updater identity remains reserved. | `L7-ORC-001` `PW-03`, `PW-06` |
| `W01-AC-009` | Grant-ladder amendment remains inert, non-interchangeable, separately auditable, and separately approvable. | `L7-ORC-001` `PW-05` and §4.1 |
| `W01-AC-010` | Shared control ownership is complete, unique, enforceable, and excludes candidate authority over protected assets. | `L7-ORC-001` §§10–12 |
| `W01-AC-011` | Later authorized baseline and shadow local verification pass with effects and limits recorded; hosted CI remains truthful. | `L7-HAR-001`; `L7-ORC-001` Wave 1 exit evidence |
| `W01-AC-012` | Exact candidate/evidence manifest and independent read-only scope-relaxation audit satisfy the R3 development gate. | `L7-ORC-001` §§12–13 |
| `W01-AC-013` | No dependency, product behavior, prototype edit, unexpected path, external effect, stable claim, or secret is introduced. | Change-contract invariants/non-goals; `L7-AUTH-002`, `L7-SAFE-003` |

No aggregate pass can waive one failed acceptance criterion.

## 9. Feature-flag and exposure decision

A runtime feature flag is `NOT_APPLICABLE` because Wave 1 creates no user-visible or production behavior. The phase-aware build-control registry is a deterministic admission control, not a product rollout flag and not an authorization mechanism. Its default behavior SHALL deny any phase/path/module/owner tuple not explicitly admitted by the approved candidate.

There is no rollout cohort, exposure percentage, telemetry metric, production observation, or removal schedule in Wave 1. Any later user-visible behavior requires its own default-OFF feature/exposure contract in the wave that creates it.

## 10. Failure, recovery, and stopping behavior

| Failure | Safe result | Required recovery/next action |
|---|---|---|
| Trace or allocation mismatch | `BLOCKED`; no scope freeze | Correct the source or mapping through an approved change; rerun against a new candidate |
| Claim/prototype disposition mismatch | `BLOCKED`; no support promotion | Correct the claim record or disposition proposal; keep prototype status |
| Phase-gate or fixture failure | Current stricter Step 5/product-path denial remains effective | Restore the last approved gate if changed; remediate in a new bounded candidate and re-audit |
| Missing module decision | Product imports remain blocked | Obtain exact owner/technology decision |
| Grant amendment not assured/approved | C2 synthetic mutation, pilot mutation, and stable mutation remain at their documented lower ceilings | Revise or defer; no alternate unsigned bypass |
| Partial or interrupted local change | `RECOVERY_REQUIRED`; freeze overlapping writers | Use the later design's exact preimage/rollback procedure; do not auto-clean or continue |
| Unexpected path/dependency/network/credential effect | Stop; candidate fails | Preserve evidence, revoke any affected authority if applicable, and obtain a new scoped plan |
| Audit Blocker/Critical/High/Medium | No Wave 1 checkpoint | Remediator creates a new candidate; independent reviewer re-audits exact bytes |

Every success, block, cancellation, recovery, or resume ends at one decision-first status with one permitted next action. Wave 2 never begins automatically.

## 11. Required later design decisions

Approval of this specification does not decide the following. The Wave 1 design proposal SHALL present and bind them explicitly:

1. exact implementation branch/worktree/base strategy and local Git effects;
2. exact per-work-package file allowlists within §3.1 and serial integration order;
3. validator input grammar, deterministic output schema, rule IDs, and parser boundaries;
4. phase-registry and successor-manifest shape, including how Foundation Step 5 remains independently verifiable;
5. permanent fixture layout and how negative fixtures cannot be mistaken for product source;
6. root module identity decision or targeted replacement proposal;
7. grant-ladder amendment artifact boundaries and independent security-review inputs;
8. exact baseline/shadow verification commands, declared cache/temp effects, and recovery;
9. exact candidate/evidence manifest and independent-review handoff; and
10. documentation/evidence records produced at each work-package and Wave 1 exit.

If any choice expands the contract's path, effect, risk, dependency, external boundary, or acceptance criteria, the contract and specification must be revised and reapproved before design continues.

## 12. Deliverables and completion state

A completed Wave 1 candidate is expected to contain, through later approved actions:

- deterministic trace and allocation validation;
- a versioned support/claim/prototype-disposition freeze;
- the phase-aware build-control successor plus permanent boundary fixtures;
- an exact module-identity decision record;
- an inert grant-ladder amendment proposal for a separate decision;
- a versioned shared-control ownership record;
- baseline/shadow local verification evidence;
- exact candidate and evidence manifests;
- a separate read-only exact-digest audit of the first scope relaxation; and
- a Wave 1 evidence/status record reporting actual results and limitations.

Artifact presence alone does not complete the wave. The checkpoint may be recorded as **build control ready; no product behavior yet** only when every acceptance criterion is satisfied and the accountable owner accepts the exact evidence-bound candidate. `PASS`, `GO`, stable support, dual-host support, security qualification, release, and deployment remain outside this wave.

## 13. Approval record

No approval is embedded in this proposal. The accountable owner may approve this specification together with `L7-W01-CC-001`, request revision, or reject it. Until a fresh exact decision is given, design and implementation are blocked.
