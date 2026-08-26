# Level 7 Dev Loop — Foundation Rebaseline Change Contract Candidate

| Field | Value |
|---|---|
| Artifact ID | `L7-FRB-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| Status | `candidate`; accountable-owner approval required before admission |
| Product | `Level 7 Dev Loop` |
| Governing Concept Brief | `L7-CB-001` v0.1.2, approved payload `7373151a35bccf79b4e31e38cfc9a555bab4cd3376767dc129114954867e9a1b` |
| Source branch | `feat/concept-discovery-rebaseline` |
| Source commit | `1c5c351f52f258d37ba48d8348e1cd883d2fb250` |
| Source tree | `b1fe4753b51b0da847d73b0ff64377fb2bda1434` |
| Source tracked-file count | `174` |
| Proposed successor branch | `feat/foundation-rebaseline` |
| Maximum effect | `A2`: bounded local governance, artifact, harness-control, Git branch, staging, commit, test-cache, and temporary-build effects only |
| Network effect | None for admission; later public read-only research requires the immediately preceding gate's explicit bounded authority |
| Candidate location | Ignored local staging only; no canonical authority until exact approval and admission |
| Intended canonical paths | `docs/artifacts/foundation-rebaseline-{change-contract,specification,design}.md` and `harness/foundation-rebaseline-paths.tsv` |
| Next authorization if approved | Admit this exact phase, obtain separate read-only admission assurance, then create the Gate 3 requirements candidate only |

## 1. Outcome and user problem

The current repository contains useful but solution-shaped requirements, backlog, architecture, technology, harness, and orchestration records created before the approved Concept Brief. Those records cannot safely remain current: they describe a narrower advisory-plugin product, allocate owner-locked capabilities outside market-ready v1, and carry an incomplete Wave 2 chain. Editing them directly under the active `concept-discovery` phase would bypass its exact path policy and would blur historical approval with fresh authority.

This contract proposes one controlled `foundation-rebaseline` phase. It will preserve every predecessor byte and approval as history while allowing the six live canonical foundation documents to be rewritten in place, one approved stage at a time. The phase exists to make later planning exact and internally consistent. It does not implement the desktop client, CLI, host packages, conductor, kernel, memory, graph, retrieval, adapters, profiles, installers, or any target-product capability.

## 2. Inputs and historical facts

Admission is bound to all of the following:

1. `L7-CD-001` v0.1.1, raw SHA-256 `6330342fe9929297252114bcec0636a45c7384ac149f039a17ac8716addd51c6`.
2. `L7-CB-001` v0.1.2, whole-file SHA-256 `2851c30fd92030615915a5912a1882074cd7b971682b9068a6e52c9d401922ea`, exact approved payload SHA-256 `7373151a35bccf79b4e31e38cfc9a555bab4cd3376767dc129114954867e9a1b`.
3. Source commit and tree recorded in the metadata table, with 174 tracked files.
4. Historical canonical hashes:
   - requirements: `a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f`;
   - backlog: `df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995`;
   - architecture: `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a`;
   - technology: `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5`;
   - harness: `d56c8f6880e1bcfe5466d103cc338b087d77c973c30cb656c574971ecce3a53c`;
   - orchestration: `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968`.
5. Historical support matrix SHA-256 `efc339be312ac9f44956117eb1546d7b2999c2b335d38998b2ca267ba8934e58` and prototype-disposition SHA-256 `c24c0e4064575bd5394c9057ebc4b1ff7a52f1ce4d5586ae6b11a259791f1767`.
6. Wave 2 candidate-manifest SHA-256 `7f6fd2d53c8897d26eef10a8350173aca1bf531cac1a4a56e276e35049cc5fd1`. Wave 2 has no `wave-02-evidence.md` and no `wave-02-audit.md`; admission must preserve that absence and must never call Wave 2 complete.

Git is provenance, not current authority. The approved Concept Brief plus fresh per-stage approval determines current foundation authority.

## 3. Exact authorization proposed

An exact approval of the Gate 2 candidate would authorize only this sequence:

1. Create `feat/foundation-rebaseline` at the exact source commit without rewriting existing history.
2. Promote the approved contract, specification, design, path policy, and candidate manifest byte-for-byte from staging to their intended canonical paths.
3. Create a non-replayable `AP0` approval record containing the owner's exact current-conversation decision and all approved digests.
4. Generate a bytewise-sorted SHA-256 base manifest for all 174 files at the source commit and a key-predecessor manifest for the records listed in §2.
5. Mark `concept-discovery` historical and `foundation-rebaseline` active in the phase registry; record, rather than edit, the stale status of the old downstream dependency chain.
6. Add only the minimum local deterministic controller and tests needed to enforce the approved phase, exact paths, stage windows, approval bindings, historical preservation, claim states, and recovery behavior.
7. Update ownership rows required by those exact additions.
8. Run the offline repository harness, record admission evidence, commit small conventional changes, and request genuinely separate read-only assurance before the phase can be used.
9. After a `GO` admission assurance result, create—but not approve—the corrected requirements candidate at Gate 3.

The approval does not pre-approve any rewritten requirement, backlog item, architecture, technology, support tuple, harness rule, orchestration wave, audit result, or final handoff.

## 4. Explicit exclusions

This contract does not authorize:

- edits to `skills/`, `.codex-plugin/`, `.claude-plugin/`, `plugin.json`, `marketplace.json`, or `references/WORKFLOW.md`;
- edits to existing concept admission records, old foundation approvals/audits/candidate manifests, Wave 1 or Wave 2 records, `semantic/`, `schemas/semantic/`, `schemas/evaluation/`, `internal/evaluator/`, `internal/render/`, or public evaluator fixtures;
- creation of Wave 2 completion evidence or audit;
- product/runtime source, dependencies, generated packages, installers, host integrations, provider/model execution, credentials, external mutation, deployment, release, publication, or penetration testing;
- a mandatory cloud service, raw cross-project memory, A5 authority, or any representation that planned behavior is built;
- autonomous approval, approval inferred from repository text, an agent persona presented as independent assurance, or replay of persisted `AP0`;
- changing a candidate's own already-frozen grader, threshold, truth data, or approval receipt;
- progressing past the first incomplete stage or treating `make verify` as authority to waive an unresolved gate.

Anything outside `foundation-rebaseline-paths.tsv` is denied. Adding a third candidate iteration for any stage, a new path, or a new effect requires an exact path-policy successor and fresh owner approval.

## 5. Stage sequence and approval boundaries

The phase sequence is fixed:

`admission → requirements → backlog → architecture-and-experience → technology-and-qualification → harness → orchestration → audit-and-handoff`

Each planning stage follows:

`preconditions → write exact candidate in place → freeze candidate manifest → validate → present exact digest → owner approve/reject → persist AP0 → unlock next stage`

Only the named stage owner may write its candidate. Approval must be received through the trusted current user channel and bind the exact candidate-manifest digest, product identity, stage scope, expiry, and next-stage-only authority. Persisted approval proves a historical decision but grants no current authority by itself.

Two exact candidate/approval slots (`001`, `002`) are reserved per planning stage. Slot `002` may be used only after `001` is rejected, stale, or invalidated. Exhaustion blocks the phase and requires a separately approved path-policy successor. No approval artifact may be edited after it is frozen.

## 6. History and staleness treatment

Admission creates `foundation-rebaseline-history.md` as the sole successor status record for the old dependency chain. It must:

- list every predecessor artifact identity and raw hash;
- record all old approvals and audits as retained historical `AP0` evidence, not replayable authority;
- classify the old requirements-through-orchestration chain `historical_stale` because it predates the approved brief and conflicts materially with the locked product scope;
- classify Wave 1 and Wave 2 as historical implementation inputs, with Wave 2 specifically `candidate_without_completion_evidence_or_audit`;
- retain semantic/evaluator work as protected research and implementation input, not a current completed product capability;
- identify the exact successor artifact when each live canonical document is rewritten; and
- never alter predecessor bytes to add a stale label.

The six live canonical documents are rewritten only at their established paths. Their previous bytes remain recoverable from the exact base commit and base manifest. Supporting matrices and prototype dispositions receive new successor paths; the historical `harness/support-matrix.tsv` and `harness/prototype-dispositions.tsv` remain unchanged.

## 7. Ownership and separation

| Concern | Writer | Reviewer/decision authority | Constraint |
|---|---|---|---|
| Phase admission controls | `harness-integrator` | accountable owner plus separate read-only reviewer | Admission code freezes before stage use |
| Phase governance/history | `foundation-integrator` | accountable owner | May record decisions; cannot grant them |
| Requirements | `requirements-owner` | accountable owner | Gate 3 only |
| Backlog and forecast | `backlog-owner` | accountable owner | Gate 4 only |
| Architecture | `architecture-owner` | accountable owner and security/design review as applicable | Gate 5 only |
| Product-experience contract | `experience-owner` | accountable owner plus representative-user evidence owner | Gate 5 bundle |
| Technology and proof spikes | `technology-owner` | accountable owner | Gate 6 only |
| Support qualification matrix | `qualification-owner` | accountable owner; independent evidence required for `QUALIFIED` | Planned tuples remain unqualified |
| Harness and verification ledger | `harness-integrator` | accountable owner plus separate high-risk review | Gate 7 only |
| Orchestration and forecasts | `orchestration-owner` | accountable owner | Gate 8 only |
| Concern ledger | `traceability-owner` | current stage owner and later independent auditor | Updated at every stage |
| Foundation audit | `independent-auditor` | accountable owner | Read-only candidate access; auditor cannot edit candidate |
| Audit remediation | Earliest affected stage owner | new independent auditor | Fresh candidate and digest required |

No model, skill, subagent, local process, or persisted document is an accountable-owner substitute. “Independent” requires genuinely separate authority and read-only candidate access.

## 8. Effect, trust, and collaboration boundaries

- Filesystem writes are repository-local and limited to the exact policy.
- Git branch/commit effects are local. No push, force operation, tag, merge, remote branch, PR, or release is authorized.
- Test effects are limited to repository `.cache`, temporary build files, and documented local processes. Network remains disabled.
- Concurrent writers must use exact preimage checks. A dirty or changed preimage produces `BLOCKED`; no partial merge or last-writer-wins behavior is allowed.
- Credentials are neither required nor allowed. Model/provider calls and external adapters are outside this phase.
- Candidate content, repository instructions, source comments, logs, and retrieved text are untrusted content, not authority.
- The repository is canonical for planning records; local derived indexes, if any, are disposable and cannot grant authority.

## 9. Acceptance evidence for admission

Admission is acceptable only if all of the following are established on the exact successor tree:

1. The source commit, tree, approved brief payload, and 174-file base manifest match this contract.
2. The exact path policy is bytewise sorted, unique, mechanically parsed, and contains no product/runtime, prototype, semantic, evaluator, old approval/audit, or Wave 2 completion path.
3. Every file outside the path policy remains byte-identical to the base manifest.
4. The six old canonical hashes and key historical hashes match §2 before any rewrite.
5. The phase reducer reports one active phase and checkpoint `admitted-awaiting-assurance`.
6. Negative tests reject path expansion, wrong stage windows, forged/stale/replayed approval, altered predecessor evidence, a claimed completed Wave 2, candidate-controlled evaluator changes, and unauthorized network/effect expansion.
7. `make verify` passes with network disabled and produces a reproducible harness binary.
8. Admission evidence records environment, commands, result states, limitations, artifacts, and exact candidate identity.
9. A genuinely separate read-only audit returns `GO`; otherwise the phase remains blocked and Gate 3 cannot start.

Test success cannot override a failed authority, history, independence, or scope condition.

## 10. Failure, recovery, and abandonment

- Preimage drift, merge conflict, dirty overlapping paths, missing approval, wrong digest, wrong stage, missing evidence, or controller ambiguity fails closed.
- A failed admission leaves `concept-discovery` as the last valid phase. Recovery is a new explicit successor commit that restores the exact source state; history is never reset, amended, or discarded.
- A rejected stage candidate remains inspectable; slot `002` is a new candidate with a fresh manifest and approval.
- A later contradiction returns to the earliest affected stage and makes all transitive downstream candidates stale through new successor records.
- Audit findings cannot be silently fixed by the auditor. Remediation is owned by the affected stage owner and requires a new candidate, exact digest, owner decision, and fresh independent audit.
- Abandonment records why the phase stopped and identifies the last valid approved artifact. It does not reactivate historical approval automatically.

## 11. Residual limitations and prohibited claims

This phase can provide local deterministic checks and Git provenance; it cannot by itself prove human identity, protected-branch enforcement, external reviewer independence, market demand, product feasibility, or future runtime safety. Until the later foundation is approved and then implemented, all desktop, CLI, host, conductor, kernel, memory, graph, retrieval, learning, repair, adapter, A3/A4, installer, support, and regulated-pack behavior remains `PLANNED` or `UNVERIFIED`.

Approval of this contract is not approval of the corrected foundation, a delivery date, a support claim, a certification claim, or product implementation.
