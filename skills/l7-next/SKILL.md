---
name: l7-next
description: >
  Conduct one intent through inspection, implementation, verification, repair,
  self-review, and a review-ready handoff without exposing a skill graph.
user-invocable: true
---

# Solo Fast Conductor

Begin by invoking the same read-only project analysis as `$l7-onboard` status
(local MCP first, bundled CLI fallback). Preserve `l7-next` as the compatibility
entry point; do not recreate a separate project-state router.

Take ownership of the complete repository-local development loop:

`one intent → inspect → implement → test → repair → self-review → handoff`

1. Inspect repository instructions, Git state, the current change, relevant code,
   tests, and CI. Preserve unrelated user work.
2. Infer the smallest coherent result. Use specialized Level 7 skills internally
   when useful; never ask the user to select or approve a skill transition.
3. Classify actual risk. Solo assurance is default unless trusted repository
   configuration explicitly selects team assurance.
4. Implement ordinary repository-local reversible work continuously. Run fast
   targeted checks first, repair in-scope failures, then run broader checks.
5. Self-review the final diff for correctness, scope, security, data,
   compatibility, performance, accessibility, operations, and rollback as
   applicable. Label it truthfully as self-review.
6. If PR publication was explicitly authorized, apply the risk label and use
   hosted exact-head evidence. Otherwise stop at a review-ready local handoff.

Solo mode does not require an independent auditor or tracked verification/audit
record, including for Tier 3 repository-local work. Team mode may require a real
distinct reviewer; resolve their forge login before starting that review.

Stop only for a material decision or missing authority at an external,
destructive, irreversible, credentialed, production, publication, deployment,
release, or protected-branch merge boundary. Ask one precise question there.
