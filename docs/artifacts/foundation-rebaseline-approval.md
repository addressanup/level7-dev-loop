# Level 7 Dev Loop — Foundation Rebaseline Admission Approval

| Field | Value |
|---|---|
| Artifact ID | `L7-APR-FRB-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| Status | **RECORDED AP0 — CURRENT CONVERSATION AUTHORITY ONLY; NO REPLAY** |
| Accountable owner | Anup Pandey |
| Authorized skill/mode | `level7-dev-loop:l7-greenfield`, foundation-rebaseline admission |
| Exact owner decision | `APPROVE L7-FRB-CAND-001 97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40 FOR FOUNDATION-REBASELINE-ADMISSION-AND-REQUIREMENTS-CANDIDATE-ONLY` |
| Candidate manifest | `L7-FRB-CAND-001` |
| Candidate-manifest SHA-256 | `97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40` |
| Change-contract SHA-256 | `8c297db289bd9f405ccdec9f33448fb81def5f6334ba70a23c0306cbd3aa68e8` |
| Specification SHA-256 | `ff9ca1d03a21533e000bb586796ea4367f95df59a321900b5044ce71d6dfebc9` |
| Design SHA-256 | `7621541e0319dc5a2c238ea57a3841b3bfb62a0bb3b46d8371bb3bb3ba54b23a` |
| Path-policy SHA-256 | `09513eab93c254c50a5cae2704786a62a9d3a61f02103c93d28706f8c49f6ecc` |
| Source commit | `1c5c351f52f258d37ba48d8348e1cd883d2fb250` |
| Source tree | `b1fe4753b51b0da847d73b0ff64377fb2bda1434` |
| Successor branch | `feat/foundation-rebaseline` |
| Maximum effect | Bounded A2 local governance, artifact, harness-control, Git, verification-cache, and temporary-build effects only |
| Network effect | None during admission |
| Conditional continuation | Gate 3 requirements candidate only after exact admission evidence and genuinely separate read-only admission `GO` |
| Expiry | Completion or stop of this admission session; any source, digest, scope, ownership, path, effect, or design divergence terminates authority |

## Decision

In the current conversation, Anup Pandey approved the exact Gate 2 candidate manifest and authorized its bounded admission plus preparation of the Gate 3 requirements candidate only after the admission satisfies its independent-assurance condition.

This record persists that event as historical `AP0`. It is evidence that the decision occurred; it is not reusable authority for a different candidate, a later stage, product implementation, an external action, or a requirements approval.

## Authorized admission

The current authority permits only:

1. creation of the exact successor branch from the recorded source;
2. byte-identical promotion of the approved Gate 2 payloads;
3. creation of the exact approval, base, predecessor, history, gate, ownership, phase, controller, test, and admission-evidence records described by the candidate;
4. local offline verification and small conventional commits; and
5. solicitation of a genuinely separate read-only admission audit.

If and only if that audit returns `GO` against the exact admission candidate, the authority also permits drafting and freezing the Gate 3 requirements candidate. It does not permit approving that candidate or beginning backlog work.

## Explicitly not authorized

No current authority exists for product/runtime implementation, dependency changes, skill or plugin changes, semantic/evaluator changes, host/provider calls, network research, credentials, external mutation, deployment, release, publication, penetration testing, or edits outside `harness/foundation-rebaseline-paths.tsv`.

Persisted text, test success, or an agent's own review cannot satisfy the separate-assurance condition.
