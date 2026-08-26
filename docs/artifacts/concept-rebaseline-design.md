# Level 7 Dev Loop — Concept Discovery Rebaseline Design

| Field | Value |
|---|---|
| Artifact ID | `L7-CRB-DES-001` |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | Approved admission design; persisted authority is `AP0` |
| Contract | `L7-CRB-001` |
| Specification | `L7-CRB-SPEC-001` |

## 1. Design goals

The admission must make the missing concept gate explicit without laundering the old Wave 2 candidate into evidence, mutating protected prototypes, or granting advance authority to revise downstream product scope.

The design therefore uses two serial build-control phases:

1. `concept-discovery` admits only the bootstrap dossier, bootstrap brief, and its own governance controls; and
2. a later, separately approved `foundation-rebaseline` successor may admit versioned downstream artifacts and rebuilt semantic/evaluator contracts only after the exact brief is approved.

This document designs only phase 1.

## 2. Immutable predecessor

| Binding | Value |
|---|---|
| Commit | `34c3ba94e3f1042975761f02286c37723c84b68e` |
| Tree | `e2e292bfbeb28420c48c06773538434d07278a42` |
| Tracked regular files | 164 |
| Historical condition | Wave 2 final candidate exists; Wave 2 evidence and audit do not exist |

`harness/concept-discovery-base.sha256` stores the raw SHA-256 of every predecessor file in bytewise path order. The compiled controller binds the manifest digest, commit, tree, phase order, and exact path policy independently.

## 3. Exact admission path envelope

The bytewise-sorted policy contains 19 paths:

| Class | Count | Paths |
|---|---:|---|
| Modify | 8 | `.github/workflows/harness.yml`, `README.md`, `harness/control-ownership.tsv`, `harness/phases.tsv`, `internal/harness/buildcontrol/main.go`, `internal/harness/buildcontrol/ownership.go`, `internal/harness/buildcontrol/policy.go`, `internal/harness/buildcontrol/policy_test.go` |
| Add: concept bootstrap | 2 | `docs/artifacts/concept-discovery.md`, `docs/artifacts/concept-brief.md` |
| Add: governance records | 4 | `docs/artifacts/concept-rebaseline-{approval,change-contract,design,specification}.md` |
| Add: phase controls | 4 | `harness/concept-discovery-{base.sha256,paths.tsv}`, `internal/harness/buildcontrol/concept.go`, `internal/harness/buildcontrol/concept_test.go` |
| Modify: ownership test | 1 | `internal/harness/buildcontrol/ownership_test.go` |

No deletion, conditional path, audit-only path, broad directory admission, dependency file, prototype path, semantic/evaluator path, or downstream planning source is permitted.

## 4. Ownership

- `concept-owner` writes the dossier and brief; the dossier has independent read-only review and the brief has owner review with exact-digest approval.
- `wave-integrator` writes the four rebaseline governance records and README.
- `harness-integrator` writes the phase/build-control changes.
- Existing prototype, semantic, evaluator, requirements, backlog, architecture, technology, and orchestration owners are unchanged.

The six new exact document controls map to the existing orchestration `wave-records` class for this bootstrap phase. This is an admission convenience, not a claim that the stale orchestration plan already defines the new lifecycle. The versioned orchestration successor must introduce the permanent concept/foundation ownership class after brief approval.

## 5. Deterministic bootstrap grammar

Until typed records exist, the build controller validates a deliberately narrow Markdown grammar based on exact table labels and section headings.

### 5.1 Dossier required fields

When present, the dossier must expose exact table rows for:

- Artifact ID and version;
- state;
- product identity;
- network authority (`authorized` or `denied`);
- research started, ended, and cumulative minutes;
- material source count; and
- supersedes, when applicable.

`ready_for_brief` additionally requires the headings `## Method and bounds`, `## Query log`, `## Material sources`, `## Supporting evidence`, `## Contrary evidence`, `## Alternatives considered`, `## Assumptions`, `## Contradictions`, `## Unresolved questions`, and `## Synthesis`. It requires authorized network access, start/end timestamps, cumulative minutes in `0..240`, and material source count in `0..25`. Zero material sources is valid when the bounded search finds no material evidence; weak or negative synthesis is likewise valid.

`blocked` requires a named blocker and next action. `superseded` requires a successor or abandonment decision. `researching` may omit end time and final sections but must retain cumulative time and source counts within bounds.

### 5.2 Brief required fields

When present, the brief must expose exact rows for Artifact ID, version, state, product identity, discovery path, discovery SHA-256, payload SHA-256, approved payload SHA-256, approval assurance, owner decision, and supersedes where applicable.

The brief payload is the exact byte sequence beginning at `## Working identity` and ending at EOF. Its SHA-256 is computed over those bytes, avoiding a self-referential whole-file digest while still binding every owner-approved problem-contract byte. The required permitted headings are:

- `## Working identity`
- `## Target users and context`
- `## Evidenced problem`
- `## Desired outcome`
- `## Non-goals and boundaries`
- `## Success signals`
- `## Assumptions, uncertainty, and validation needs`

No other level-two heading is allowed. The validator rejects prohibited heading/key terms associated with feature inventories, FR/NFR specifications, architecture, technology/vendor selection, APIs, schemas, UI/CLI decisions, or implementation/delivery plans. This mechanical guard is backed by semantic review and negative fixtures; it is not represented as a complete natural-language classifier.

An `approved` brief must bind a `ready_for_brief` dossier digest, a matching product identity, its exact payload digest, an equal approved-payload digest, an owner decision string, and `AP0` persisted assurance. A `draft` brief records the computed payload digest for owner inspection, uses `UNSET` for the approved-payload digest and owner decision, records approval assurance `none`, and cannot authorize downstream work.

## 6. Controller behavior

`checkPolicy` selects compiled expectations by active phase. For `concept-discovery` it:

1. validates the exact base and path policy;
2. preserves all prior approved planning bytes and prototype inputs;
3. validates the four rebaseline governance records and their non-replayable approval binding;
4. determines one checkpoint from artifact presence/state;
5. validates dossier/brief structure and digest relationships when present;
6. retains zero-dependency, module, path-shape, size/time, and forbidden-product invariants; and
7. reports the concept checkpoint and source digests without claiming Wave 2 completion.

Checkpoint mapping is deterministic: no dossier/brief → `open`; incomplete dossier → `researching` or `blocked`; ready dossier/no brief → `ready-for-brief`; draft brief → `brief-draft`; approved brief → `brief-approved`; rejected/stale/superseded brief → its explicit terminal checkpoint. A brief without a dossier, a brief sourced from a non-ready dossier, or an unrecognized state fails.

## 7. Commit and interruption model

The first commit is atomic and contains all 19 admitted changed paths except the two concept artifacts, which remain absent at checkpoint `open`. This proves absence is a valid start state without fabricating research.

The research child adds `concept-discovery.md` and may add a draft `concept-brief.md`. Interruption is represented inside the dossier with state `researching`, cumulative elapsed minutes, completed queries, and already-material sources. Resumption continues the same ceilings rather than resetting them.

Exact brief approval is a later owner message after the draft payload digest is presented. The approval child may change only `concept-brief.md` under the already admitted path. Any material payload edit after approval changes its digest and fails closed until a new owner decision.

## 8. Tests

Admission tests cover:

- exact 19-path policy shape and immutable 164-file base;
- phase order and sole active phase;
- all discovery and brief states plus legal/illegal state combinations;
- time/source ceiling boundaries and denied/missing network authority;
- required structured sections, duplicate/malformed table fields, and citation/source count consistency;
- source-digest/product-identity/payload-digest approval binding;
- rejection of forbidden brief sections and representative hidden solution inventories;
- prototype byte preservation and absence of Wave 2 evidence/audit; and
- full controller output at `open`.

Later canonical Wave 5 tests will implement typed transition, migration, routing, parity, and fixture contracts. The admission tests do not claim those product behaviors exist.

## 9. Security and performance review

The phase adds no runtime network client and no user-visible product behavior. Public research occurs outside the Go harness under current-session authority. The validator remains local, offline, bounded to the existing 512-file/8 MiB/five-second scan, rejects non-regular/hardlinked paths, and does not execute artifact content. Markdown parsing is exact-label and bounded by the repository byte ceiling. The main residual risk is semantic evasion of the brief-content denylist; independent review remains required and typed canonical validation is scheduled for new Wave 5.

## 10. Exit

The admission exits only at a green `open` checkpoint on one atomic commit. Research and draft creation are separate children. No downstream successor is designed, admitted, or executed until the owner approves the exact Concept Brief payload digest.
