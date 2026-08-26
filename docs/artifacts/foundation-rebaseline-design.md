# Level 7 Dev Loop — Foundation Rebaseline Design Candidate

| Field | Value |
|---|---|
| Artifact ID | `L7-FRB-DES-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| Status | `candidate`; not admitted |
| Contract | `L7-FRB-001` |
| Specification | `L7-FRB-SPEC-001` |
| Source commit | `1c5c351f52f258d37ba48d8348e1cd883d2fb250` |
| Source tree | `b1fe4753b51b0da847d73b0ff64377fb2bda1434` |
| Design scope | Local foundation-governance phase only |

## 1. Design goals

The design must permit an intentionally large foundation correction without turning the owner mandate into one uncontrolled rewrite. It must preserve useful history, block stale authority, keep product implementation out, make each candidate exact, and keep ordinary solo/small-team use comprehensible.

The phase therefore uses:

1. one immutable Git/base-manifest preimage;
2. one exact finite path envelope with stage windows;
3. six in-place canonical successor documents;
4. immutable per-stage candidate and approval records;
5. one evolving concern ledger plus bounded supporting artifacts;
6. deterministic state reconstruction rather than file-presence authority;
7. separate admission and final read-only assurance; and
8. explicit recovery to the earliest affected stage.

## 2. Candidate staging and promotion

Before Gate 2 approval, the contract, specification, design, and path policy live only in `.cache/gates/foundation-rebaseline/`, which is ignored repository-local staging. This avoids an unauthorized new canonical path under the active `concept-discovery` policy. Staged files confer no authority.

The Gate 2 candidate manifest contains raw SHA-256 digests for those four files in bytewise intended-canonical-path order. The manifest's own raw SHA-256 is the owner approval target.

After exact approval:

1. verify that `HEAD`, tree, brief, and candidate bytes still match;
2. create the proposed successor branch at the exact preimage;
3. apply the four files byte-for-byte at their intended canonical paths;
4. add the exact candidate manifest and current-conversation approval receipt;
5. generate base/predecessor manifests and phase controls;
6. run the admission gates;
7. commit conventionally in small bounded changes; and
8. stop at `admitted-awaiting-assurance` until a separate reviewer returns a bound result.

Any staging/canonical digest mismatch blocks. Promotion is never a reinterpretation or rewrite of the approved candidate.

## 3. Phase records and data model

### 3.1 Static admission records

- `foundation-rebaseline-change-contract.md`: authority and effect boundary.
- `foundation-rebaseline-specification.md`: normative lifecycle and stage acceptance.
- `foundation-rebaseline-design.md`: implementation-neutral phase-control design.
- `foundation-rebaseline-candidate.sha256`: exact Gate 2 bundle manifest.
- `foundation-rebaseline-approval.md`: current owner decision persisted as non-replayable AP0.
- `foundation-rebaseline-history.md`: predecessor/staleness/successor record.
- `foundation-rebaseline-base.sha256`: all 174 source files.
- `foundation-rebaseline-predecessors.sha256`: key historical inputs and absences.
- `foundation-rebaseline-paths.tsv`: exact path/window/owner policy.
- `foundation-rebaseline-gates.tsv`: static checkpoint order, required records, and successor route.

### 3.2 Stage records

Each planning stage owns a canonical document or bundle, `candidate-001.sha256`, and `approval-001.md`. A rejected/stale candidate may use reserved slot `002`. Approval files are immutable after creation. Candidate manifests bind the changed canonical document, supporting files, current concern ledger, all upstream approval digests, and the frozen admission-control set.

### 3.3 Cross-cutting records

- `concern-capability-ledger.md` is the canonical concern traceability table and is updated only in declared windows.
- `delivery-forecast.md` holds staffing/throughput/range scenarios so backlog scope is not mistaken for a calendar promise.
- `product-experience-design-contract.md` makes the Level 7 desktop experience testable rather than a generic dashboard aspiration.
- `prototype-disposition.md` and `foundation-prototype-dispositions.tsv` bind legacy entry points to one conductor.
- `technology-research-ledger.md`, `technology-proof-spikes.md`, and the two support-matrix projections bind current technical evidence and qualification truth.
- `foundation-verification-ledger.md` maps requirements and fitness functions to evidence owners and suites.

Markdown records are human-reviewable canonical planning artifacts. TSV projections exist only where deterministic matrix validation materially helps. If a Markdown/TSV pair disagrees, the candidate fails; neither silently overrides the other.

## 4. State reconstruction

The local controller scans bounded regular files, validates the exact base and path envelope, then reconstructs state from records in dependency order:

```text
approved Concept Brief
        ↓
admission approval + exact bundle
        ↓
admission evidence + separate admission audit
        ↓
requirements candidate/approval
        ↓
backlog candidate/approval
        ↓
architecture + experience candidate/approval
        ↓
technology + qualification candidate/approval
        ↓
harness + verification candidate/approval
        ↓
orchestration candidate/approval
        ↓
full manifest → separate audit → final owner approval → handoff
```

The reducer never infers approval from file existence. It validates raw digests, exact predecessor relations, allowed state transitions, window closure, and non-stale upstream inputs. If more than one checkpoint appears current, if an upstream digest changes, or if an approval has no matching candidate, the result is `BLOCKED`.

## 5. In-place rewrite without historical loss

The live six foundation paths are intentionally reused so future agents have one canonical location. Before any rewrite, admission freezes the complete source tree and a focused predecessor manifest. `foundation-rebaseline-history.md` records old IDs, versions, hashes, approvals, audit status, deficiencies, and successor IDs.

The status model is:

| Record | Admission status | Mutation behavior |
|---|---|---|
| Approved Concept Brief | `current` | Omitted from path policy; immutable |
| Old requirements through orchestration | `historical_stale` | Live paths may be rewritten only in their stage; old bytes remain at base commit |
| Old approvals/audits/candidate manifests | `historical_ap0` | Omitted; byte-identical |
| Wave 1 | `historical_input` | Omitted; no completion claim change |
| Wave 2 | `candidate_without_completion_evidence_or_audit` | Omitted; missing evidence/audit remain missing |
| Semantic/evaluator work | `protected_historical_input` | Omitted; no built-product claim |
| Prototype skills/manifests | `protected_historical_input` | Omitted; disposition only in new successor table |
| Old support/disposition TSVs | `historical_stale` | Omitted; new successor TSV paths used |

This makes Git necessary provenance while preventing Git history or old prose from becoming current authority.

## 6. Windowed path enforcement

The path policy is the union of every bounded file that may change during the phase. Its `window` column prevents that union from becoming a phase-long broad write grant.

For each command, the controller receives or reconstructs one intended stage. It performs:

1. exact `HEAD`/base/phase validation;
2. bounded repository scan and file-shape validation;
3. preimage digest validation for every candidate target;
4. owner/window admission;
5. candidate delta validation;
6. post-write scan and manifest generation;
7. stage-specific tests; and
8. checkpoint reconstruction.

Admission-control source files may change only in `admission`. Their exact digests are then frozen into every later candidate manifest. A later stage that alters admission controls fails even though those paths appear in the phase's union policy.

Canonical document paths have one stage window. Cross-cutting ledgers have an explicit ordered window list and must preserve prior approved rows unless the current candidate records a supersession. Audit-owned files are writable only by the independent auditor; remediation files are writable only by the earliest affected stage owner and cannot change the audit record.

## 7. Approval presentation and receipt

The trusted presentation must show:

- artifact and stage identity;
- exact candidate-manifest SHA-256;
- changed paths and preimage hashes;
- material decisions and alternatives;
- assumptions, contradictions, evidence gaps, and unsupported claims;
- effect/risk ceiling;
- validation results and limitations;
- recovery/restart implications; and
- the one exact next-stage-only approval command.

The owner response is checked against the presented digest in the current user channel. The persisted receipt records that event as `AP0`, its expiry, scope, and non-replay rule. It cannot be used to approve a later candidate or execute product effects.

## 8. Ownership and conflict handling

One integrator owns the active candidate branch at a time. Specialist owners prepare content serially in the same worktree unless isolated branches with explicit merge ownership are approved. “Parallel” agents writing a shared worktree are not a supported phase mode.

Before mutation, the writer records exact preimages for every target and verifies no overlapping dirty change. If another writer or user changes a target, the current operation stops with a conflict report containing expected/actual hashes. It does not auto-merge governance or approval records. A resolved merge becomes a fresh candidate with a fresh digest.

Git operations are non-destructive: no reset, amend, force push, history rewrite, or silent checkout restoration. Recovery uses new commits or an explicit abandoned/superseded record.

## 9. Assurance separation

### 9.1 Admission assurance

Because admission changes the authority controller, local self-tests are necessary but insufficient. A separate reviewer receives the exact committed admission candidate read-only and verifies source identity, path closure, stage-window enforcement, history treatment, negative tests, and harness evidence. The reviewer writes only `foundation-rebaseline-admission-audit.md`. The integrator cannot produce a qualifying result.

### 9.2 Final foundation audit

The full candidate manifest freezes all current foundation artifacts and the admission-control digest set. The independent auditor cannot modify those paths and cannot change thresholds or truth data. A `NO_GO` audit remains immutable. Remediation occurs on a successor candidate, then a different fresh audit binding is required.

Local files cannot prove external human identity or independence cryptographically. The claim is limited to recorded channel/role separation and must be strengthened by repository/forge protections before any release use.

## 10. Failure modes and degraded behavior

| Failure | Result | Recovery |
|---|---|---|
| Source commit/tree differs | `BLOCKED` | Rebase is not inferred; create and approve a new Gate 2 candidate |
| Approved brief digest differs | `BLOCKED` | Return to Concept Brief staleness decision |
| Candidate staging digest differs | `FAIL` | Regenerate and re-present Gate 2 candidate |
| Unlisted or wrong-window path changes | `FAIL` | Remove only through authorized recovery or approve a successor policy |
| Dirty overlapping user change | `BLOCKED` | Owner resolves ownership; then create fresh preimages |
| Approval missing/mismatched/replayed | `BLOCKED` | Present the current exact candidate to owner |
| Old approval/evidence changes | `FAIL` | Restore in a new explicit successor; investigate integrity event |
| Wave 2 completion artifact appears | `FAIL` | Remove through authorized recovery; do not invent evidence |
| Admission tests pass but reviewer absent | `BLOCKED` | Obtain genuinely separate read-only assurance |
| Technology research unavailable | `BLOCKED` or truthful `UNVERIFIED` | Resume within separately approved bounds; never fabricate matrix status |
| Candidate slot `002` exhausted | `BLOCKED` | Approve a path-policy successor |
| Audit finds cross-stage contradiction | `NO_GO` | Return to earliest affected stage; stale downstream chain |
| Audit tool/environment unavailable | `BLOCKED` | Retain frozen candidate; do not self-audit |

No degraded mode allows protected writes, approval inference, false support, or product implementation.

## 11. Recovery model

Recovery is append-only in history:

- Before Gate 2 approval: delete ignored staging only; canonical repository remains at approved Concept Brief.
- During admission: if atomic checks fail before commit, revert only newly staged authorized paths; if committed, use a new recovery commit.
- During a planning candidate: retain or explicitly abandon the candidate; do not mutate its approval record.
- After an approved stage becomes stale: create a successor history record and use reserved candidate slot `002`.
- After audit `NO_GO`: remediate through the earliest affected stage and freeze a new full candidate.
- On phase abandonment: preserve all records and name the last valid checkpoint. Historical foundation does not automatically reactivate.

Base files are recoverable from `1c5c351`; derived candidate caches may be deleted and regenerated because canonical authority is repository-owned.

## 12. Admission implementation shape

The proposed admission adds `internal/harness/buildcontrol/foundation.go` and its tests, then minimally extends `main.go`, `policy.go`, and `ownership.go` plus their tests. It does not add a production package or dependency.

The foundation controller owns:

- parsing/validating the path and gate TSVs;
- base/predecessor manifest checks;
- phase/checkpoint reconstruction;
- stage-window and immutable-record enforcement;
- candidate/approval digest binding;
- historical and claim-state checks; and
- deterministic report output.

It reuses existing bounded scan, strict file, SHA-256, Markdown-table, finding, and reporting helpers. No artifact text is executed. Network is disabled by the Makefile environment. Repository scan bounds and non-regular/hardlink rejection remain in force.

## 13. Admission test design

Tests include:

1. exact path count, sort order, schema, owner, windows, and prohibited-prefix absence;
2. exact source commit/tree and 174-file base manifest;
3. approved brief/discovery binding;
4. sole active phase and every legal/illegal checkpoint transition;
5. historical predecessor hashes and missing Wave 2 evidence/audit;
6. old approval/audit/prototype/semantic/evaluator byte preservation;
7. candidate slot ordering and immutable approval receipts;
8. forged, stale, mismatched, expired, replayed, or pre-presented approval rejection;
9. wrong-stage path, unlisted path, case alias, symlink, hardlink, special file, oversized scan, and timeout rejection;
10. admission-control freeze after the admission window;
11. candidate inability to change its audit record, evaluator/control freeze, or claim-state rules;
12. planned-versus-built/support/certification negative claims;
13. deterministic findings across ordering and process runs; and
14. full `make verify`, including reproducible test binary output.

Fixtures are local and adversarial; they do not call providers or external systems.

## 14. Performance and resource bounds

Admission retains the existing repository scan limits unless an exact measured successor is separately approved. Parsing is single-pass over bounded UTF-8/ASCII control files. Hashing is streaming and repository-local. Tests use bounded temporary directories under `.cache` and timeouts. The controller opens no network socket, stores no credential, starts no daemon, and has no production shell authority.

If a future foundation candidate exceeds scan, file, manifest, or time bounds, the result is `BLOCKED` with a measured proposal; the controller does not silently lift limits.

## 15. Security analysis

Primary threats are approval replay, stale authority, path-envelope expansion, candidate-controlled evaluation, malicious artifact text, history rewriting, alias/hardlink escape, and false completion/support claims.

Controls are exact digest binding, current-channel approval, AP0 persistence, stage windows, base manifests, canonical paths, regular-file/single-link checks, omitted protected prefixes, frozen evaluator/control digests, independent read-only audit, explicit claim states, and non-destructive recovery.

Residual risks include local-user compromise, a malicious integrator changing both code and tests before external review, weak human review, and Git objects being garbage-collected or repository history being rewritten outside this phase. The design does not claim tamper-proof storage or verified human identity. Later product architecture must address signed releases and stronger trust anchors; this phase remains local planning governance.

## 16. Exit and handoff

Admission exits at `admitted` only after local evidence and separate read-only `GO`. The full phase exits only after all planning gates, a full independent `GO` audit, and final exact owner approval. The handoff names what is built (repository governance controls and plans), what remains planned (the entire product runtime), critical risks, forecast assumptions, and exactly one next Level 7 skill.

No product implementation begins under this design.
