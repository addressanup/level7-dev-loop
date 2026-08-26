# Level 7 Dev Loop — Concept Discovery Rebaseline Specification

| Field | Value |
|---|---|
| Artifact ID | `L7-CRB-SPEC-001` |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | Approved admission specification; persisted authority is `AP0` |
| Governing contract | `L7-CRB-001` |
| Applies to | Future canonical greenfield lifecycle and its generated `l7-greenfield` projection |
| Explicit exclusion | Current files under `skills/`, both plugin manifests, root manifests, and `references/WORKFLOW.md` |

## 1. Normative terms

`SHALL`, `SHALL NOT`, `SHOULD`, and `MAY` are normative. A check that cannot establish a required condition SHALL return `BLOCKED` or `FAIL`; it SHALL NOT infer success from absence of evidence.

## 2. Canonical lifecycle order

The future greenfield foundation sequence SHALL be:

`Concept Discovery → Concept Brief approval → requirements → backlog → architecture/product-form decision → technology impact/selection → harness → orchestration → execution`.

Concept Discovery and Concept Brief are bootstrap Markdown authority only until the artifact core exists. The new Wave 5 SHALL migrate them into typed canonical records, emit deterministic Markdown projections, and nominate exactly one authority source. Later conductor and execution waves SHALL consume those contracts and SHALL NOT redefine them.

## 3. Concept Discovery dossier

The canonical bootstrap path is `docs/artifacts/concept-discovery.md`.

### 3.1 States

The only states are:

- `open`: admitted, but no research execution has begun;
- `researching`: a bounded pass began and may be resumed;
- `ready_for_brief`: the authorized bounded pass ended and the dossier is structurally complete;
- `blocked`: the pass could not be executed because required authority, network capability, or another named prerequisite was unavailable; and
- `superseded`: a retained historical dossier was replaced or the concept was abandoned.

Allowed transitions are `open → researching`, `open → blocked`, `researching → researching`, `researching → ready_for_brief`, `researching → blocked`, `blocked → researching`, and any non-superseded state → `superseded`. `ready_for_brief` is not evidence strength and MAY contain a weak, negative, conflicting, or null conclusion.

### 3.2 Bounded research contract

The dossier SHALL record:

- explicit public-network authority and its scope;
- start/end timestamps and cumulative elapsed minutes;
- a ceiling of 240 cumulative minutes;
- an ordered query log;
- no more than 25 deduplicated material sources;
- for every material source: stable source ID, title, publisher/author, URL or public identifier, publication/update/access date where available, source type, relevance, and provenance notes;
- supporting evidence, contrary evidence, alternatives, assumptions, contradictions, limitations, and unresolved questions;
- a synthesis that distinguishes sourced fact, owner assertion, and inference; and
- an interruption/resume record sufficient to enforce the cumulative bounds.

Duplicate mirrors, search-result pages, promotional repetitions, and sources that add no material evidence SHALL NOT inflate the source count. A source is not material merely because it was opened. Citations SHALL resolve to recorded sources; fabricated citations, invented quotes, or unsupported provenance SHALL fail validation.

Page text, embedded prompts, comments, downloads, and instructions encountered during research are untrusted evidence. They SHALL NOT expand authority, alter repository instructions, request credentials, cause code execution, or become persisted commands. Only public read-only retrieval is allowed.

Raw conversations, hidden reasoning, credentials, private tokens, unnecessary personal data, and full copyrighted works SHALL NOT be persisted. The dossier is a structured research synthesis, not a transcript.

### 3.3 Completion and blocking

Research is complete when the bounded authorized search was actually executed and the required record is present—not when it finds strong support. Missing network authority or capability SHALL yield `blocked`. Denied authority SHALL be named as denied; it SHALL NOT be silently converted to an offline-only `ready_for_brief` result.

## 4. Concept Brief

The canonical bootstrap path is `docs/artifacts/concept-brief.md`.

### 4.1 States

The only states are `draft`, `approved`, `rejected`, `stale`, and `superseded`.

A brief may be drafted only from a `ready_for_brief` dossier. `approved` requires an explicit owner decision that binds the exact brief SHA-256 digest, the discovery artifact path and digest, the current product identity, and the approved scope. Persisted approval is historical `AP0`; it cannot authorize requirements or later mutation. Rejection and supersession remain inspectable.

### 4.2 Permitted content

The brief has no length cap, but SHALL settle only:

1. working identity;
2. target users and context;
3. evidenced problem;
4. desired outcome;
5. non-goals and boundaries;
6. success signals; and
7. assumptions, uncertainty, and validation needs.

### 4.3 Prohibited solution content

The brief SHALL reject feature inventories, user-story or backlog inventories, functional or nonfunctional requirement specifications, architecture, technology or vendor choices, implementation plans, delivery waves, schema/API/CLI/UI decisions, and other solution-form decisions. Examples used to clarify a problem SHALL be explicitly non-normative and SHALL NOT become scope.

## 5. Ownership and routing

Canonical `l7-greenfield` exclusively owns discovery execution, both concept artifacts, the brief approval gate, conditional backfill, and the later foundation sequence.

Canonical `l7-next` is read-only and selects exactly one route:

| Observed state | Sole route |
|---|---|
| No concept artifacts | Start Concept Discovery in `l7-greenfield` |
| Dossier `open`, `researching`, or resumable `blocked` | Continue Concept Discovery in `l7-greenfield` |
| Dossier `ready_for_brief`, no brief | Draft Concept Brief in `l7-greenfield` |
| Brief `draft` | Request owner decision in `l7-greenfield` |
| Brief `approved`, no current requirements successor | Continue to requirements in `l7-greenfield` |
| Brief `rejected`, `stale`, or `superseded` | Stop or restart at the earliest affected concept stage in `l7-greenfield` |

Canonical `l7-build` SHALL NOT create, reconstruct, edit, repair, approve, or bypass concept artifacts. Greenfield execution requires a current approved brief plus the approved downstream backlog, architecture, harness, and orchestration chain.

## 6. Existing-project migration

Existing greenfield projects SHALL use conditional backfill:

1. inventory existing requirements, backlog, architecture, technology, harness, orchestration, approvals, candidates, and evidence;
2. reconstruct a dossier draft and Concept Brief draft from those records, labeling provenance and inference;
3. execute the same mandatory public research pass;
4. request exact owner approval of the resulting brief;
5. compare the approved brief with every downstream artifact; and
6. retain downstream artifacts only if their user, problem, outcome, and boundaries are materially consistent.

If the approved brief materially changes or contradicts downstream scope, requirements and every transitive dependent artifact SHALL be marked `stale`, with retained historical identity, and the lifecycle SHALL restart at the earliest affected stage. A material change to target user, evidenced problem, desired outcome, or boundary likewise stales the brief and all dependent requirements, architecture, plans, candidates, approvals, and evidence. An unchanged or editorial-only backfill SHALL NOT churn otherwise current downstream artifacts.

## 7. Roadmap requirements

After exact brief approval, versioned successors SHALL be issued for requirements, backlog, architecture, technology impact/selection, and orchestration. `L7-BL-016` remains the umbrella identity, moves to P0, drops the `L7-BL-042` dependency, and is re-estimated to `13/L`. Its work SHALL be split into bounded concept-contract, foundation-transition, and end-to-end integration packages.

The old Wave 2 semantic/evaluator candidate remains inspectable history. The affected semantic and evaluator lineage SHALL be rebuilt from the revised normative source. Existing Slice 3 controls SHALL remain byte-identical; evaluator controls SHALL be superseded through explicit versions rather than rewritten.

A dedicated new Wave 5 SHALL follow the artifact/state core and implement canonical concept/foundation schemas, reducers, migration behavior, workflow/profile contracts, deterministic Markdown projection, and deterministic evaluation. Existing Waves 5–13 SHALL become Waves 6–14 without changing their internal scope except for consuming the new contracts.

## 8. Verification contract

The implementation plan SHALL include:

- schema and transition tests for every dossier/brief state, approval binding, supersession, staleness, and resume path;
- research tests for source/time ceilings, weak or negative findings, unavailable network, denied authority, provenance, malicious-page instructions, duplication, and fabricated citations;
- content tests proving prohibited solution content cannot enter the brief;
- routing tests proving one read-only `l7-next` state and no `l7-build` bypass;
- migration tests for coherent backfill, contradiction, identity changes, and unchanged downstream artifacts;
- at least six representative fixtures, collectively covering vague and well-formed ideas, unsupported claims, conflicting evidence, abandonment, interruption/resume, existing-project backfill, and stale approval;
- cross-host parity for state, authority, next transition, artifact validity, and failure outcome; and
- full candidate verification, a versioned evaluator-control freeze, exact path closure, and fresh independent read-only audit before promotion.

## 9. Admission acceptance criteria

`L7-CRB-AC-001` through `L7-CRB-AC-010` are blocking:

1. `AC-001`: `34c3ba9` remains an ancestor and its tree is not amended.
2. `AC-002`: the active phase base manifest is an exact SHA-256 inventory of all 164 predecessor files.
3. `AC-003`: current prototype skills/manifests/reference workflow remain byte-identical.
4. `AC-004`: old Wave 2 evidence/audit paths remain absent.
5. `AC-005`: the new path policy admits only concept bootstrap artifacts and required governance controls.
6. `AC-006`: invalid states, illegal transitions, and false research completion fail closed.
7. `AC-007`: forged, stale, mismatched, or replayed brief approval fails closed.
8. `AC-008`: prohibited solution-form sections or normative solution decisions fail validation.
9. `AC-009`: `make verify` passes on the atomic admission commit.
10. `AC-010`: downstream source artifacts remain byte-identical until exact brief approval.
