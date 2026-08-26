# Level 7 Dev Loop — Concept Discovery Rebaseline Admission Approval

| Field | Value |
|---|---|
| Artifact ID | `L7-APR-CRB-001` |
| Artifact type | Persisted admission-approval record |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | **RECORDED AP0 — CURRENT CONVERSATION AUTHORITY ONLY; NO REPLAY** |
| Accountable owner | Anup Pandey |
| Authorized skill/mode | `level7-dev-loop:l7-build`, one approved lifecycle correction |
| Authorized implementation | Atomic `concept-discovery` admission, then the bounded public research pass and draft Concept Brief; downstream revision remains gated by a later exact brief approval |
| Branch | `feat/concept-discovery-rebaseline` |
| Source commit | `34c3ba94e3f1042975761f02286c37723c84b68e` |
| Source tree | `e2e292bfbeb28420c48c06773538434d07278a42` |
| Change contract SHA-256 | `691662df69d348312bed54045632e73c8260d951f2683101790a321c7544d936` |
| Specification SHA-256 | `d3d957dce1571448c7a0f6bcc8d9c7afe2d952e6cec46b2c6de4259871a32871` |
| Design SHA-256 | `83cf87d13834f11ee8b2d69200f0ed6a73c346abedb2dd7f65b8c1b334fe1296` |
| Public research ceiling | Four cumulative hours and 25 deduplicated material public sources |
| Maximum effect | A2 local repository edits, repository-scoped verifier cache/temp effects, staging, and conventional commits on this successor branch; public network reads only for Concept Discovery |
| Expiry | Completion or stop of this implementation session; any material identity, scope, path, ownership, effect, or design divergence terminates authority |

## Decision

In the current conversation, Anup Pandey supplied the complete “Add Concept Discovery to the Canonical Greenfield Lifecycle” plan and directed the implementation to be carried through in a fresh context. That current-session direction authorizes the exact admission contract, specification, and design identified above, including the explicitly bounded public research pass.

This persisted record is evidence of that decision at `AP0`; it is not replayable authority. It does not approve a not-yet-authored Concept Brief, does not authorize an agent to approve on the owner’s behalf, and does not authorize requirements, backlog, architecture, technology, orchestration, semantic, evaluator, prototype, release, deployment, or external mutation work.

## Exact boundary

The atomic admission commit may contain only the 19 paths in `harness/concept-discovery-paths.tsv`, with the two bootstrap artifacts absent at the initial `open` checkpoint. Later discovery children may add only those two artifacts. The owner must receive the exact draft brief payload digest and explicitly decide it before any downstream artifact is admitted or changed.

Commit `34c3ba9` remains historical and unevidenced. This decision explicitly forbids creating `docs/artifacts/wave-02-evidence.md` or `docs/artifacts/wave-02-audit.md` for its stale scope.
