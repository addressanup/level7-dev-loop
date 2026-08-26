# Level 7 Dev Loop — Foundation Rebaseline History and Supersession Record

| Field | Value |
|---|---|
| Artifact ID | `L7-FRB-HIST-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| State | `admission` |
| Governing Concept Brief | `L7-CB-001` v0.1.2; approved payload `7373151a35bccf79b4e31e38cfc9a555bab4cd3376767dc129114954867e9a1b` |
| Predecessor commit | `1c5c351f52f258d37ba48d8348e1cd883d2fb250` |
| Predecessor tree | `b1fe4753b51b0da847d73b0ff64377fb2bda1434` |
| Complete predecessor inventory | `harness/foundation-rebaseline-base.sha256` |
| Key predecessor inventory | `harness/foundation-rebaseline-predecessors.sha256` |
| Authority rule | This successor record classifies currentness; predecessor text and persisted `AP0` never grant replayable authority |

## 1. Decision

The approved Concept Brief materially changes the target audience, product purpose, product form, lifecycle, market-ready-v1 scope, autonomy/effect boundary, and evidence obligations assumed by the existing downstream foundation. Therefore the pre-rebaseline requirements-through-orchestration dependency chain is `historical_stale` from admission of this phase.

The old bytes remain preserved by Git and the exact base manifests. This record does not edit them, erase their approvals, or imply that the earlier work lacked value. It prevents their historical status text from being mistaken for current authority.

## 2. Current and historical classification

| Artifact or set | Exact identity/hash | Classification at admission | Reason | Successor |
|---|---|---|---|---|
| Concept Discovery | `L7-CD-001` v0.1.1; `6330342fe9929297252114bcec0636a45c7384ac149f039a17ac8716addd51c6` | `current_input` | Completed bounded research record | Remains immutable through this phase |
| Concept Brief | `L7-CB-001` v0.1.2; whole file `2851c30fd92030615915a5912a1882074cd7b971682b9068a6e52c9d401922ea`; payload `7373151a35bccf79b4e31e38cfc9a555bab4cd3376767dc129114954867e9a1b` | `current_approved_input` | Exact current-conversation approval persisted as AP0 | Governs all successor planning |
| Requirements | `L7-REQ-001`; `a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f` | `historical_stale` | Predates brief and conflicts with locked product definition/scope | `L7-REQ-002` at established path; `PENDING_GATE_3` |
| Feature backlog | `L7-BKL-001`; `df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995` | `historical_stale` | Depends on stale requirements and defers locked v1 capabilities | `L7-BKL-002`; `PENDING_GATE_4` |
| Architecture | `L7-ARC-001`; `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` | `historical_stale` | Depends on stale scope and narrower product form | `L7-ARC-002`; `PENDING_GATE_5` |
| Technology selection | `L7-TEC-001`; `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` | `historical_stale` | Go-only/no-database choices require reassessment against corrected product | `L7-TEC-002`; `PENDING_GATE_6` |
| Harness plan | `L7-HAR-001`; `d56c8f6880e1bcfe5466d103cc338b087d77c973c30cb656c574971ecce3a53c` | `historical_stale` | Does not verify the corrected product/foundation contract | `L7-HAR-002`; `PENDING_GATE_7` |
| Orchestration plan | `L7-ORC-001`; `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` | `historical_stale` | Dependency graph and waves follow stale foundation | `L7-ORC-002`; `PENDING_GATE_8` |
| Historical requirements/backlog/architecture/technology/harness/orchestration approvals and audits | Exact bytes in base manifest | `historical_ap0` | Evidence of past decisions; never replayable | Preserved unchanged |
| Wave 1 records and implementation | Exact bytes in base manifest | `protected_historical_input` | Useful control implementation and evidence; not the corrected product | Preserved unchanged |
| Wave 2 candidate | Manifest `7f6fd2d53c8897d26eef10a8350173aca1bf531cac1a4a56e276e35049cc5fd1` | `candidate_without_completion_evidence_or_audit` | Candidate exists; completion evidence and audit do not | Preserved unchanged; no completion claim |
| Semantic/evaluator sources, schemas, fixtures, and renderer | Exact bytes in base manifest | `protected_historical_input` | Useful research/implementation input whose normative source is stale | Preserved unchanged during foundation planning |
| Prototype skills and plugin manifests | Exact bytes in base manifest | `protected_historical_input` | Compatibility/research input, not final architecture | Preserved; new disposition at Gate 5 |
| `harness/support-matrix.tsv` | `efc339be312ac9f44956117eb1546d7b2999c2b335d38998b2ca267ba8934e58` | `historical_stale` | Narrow historical claim ceiling | New successor matrix at Gate 6 |
| `harness/prototype-dispositions.tsv` | `c24c0e4064575bd5394c9057ebc4b1ff7a52f1ce4d5586ae6b11a259791f1767` | `historical_stale` | Dispositions conflict with locked market-ready-v1 scope | New successor disposition at Gate 5 |

## 3. Required absences

At the predecessor commit, both paths are absent and must remain absent throughout foundation planning:

- `docs/artifacts/wave-02-evidence.md`
- `docs/artifacts/wave-02-audit.md`

Their absence means Wave 2 is not complete. No successor may create either file or fabricate equivalent historical evidence.

## 4. Successor update rule

When a stage receives exact owner approval, this document may append or update only the corresponding `Successor` field and its own version/state. Earlier classifications, hashes, approval facts, and Wave 2 truth are immutable. A material contradiction returns to the earliest affected stage and records downstream staleness without rewriting prior approval records.

## 5. Current claim boundary

At admission, only Concept Discovery, the Concept Brief, and local governance/harness controls are current. The corrected requirements, backlog, architecture, technology, qualification matrix, harness, orchestration, and product runtime do not yet exist as approved current artifacts. All product capabilities remain `PLANNED` or `UNVERIFIED`.
