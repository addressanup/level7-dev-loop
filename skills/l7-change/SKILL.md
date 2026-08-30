---
name: l7-change
description: >
  Change a live product with one risk-tiered workflow: brief when needed,
  implementation, tests, review, and merge or safe release.
user-invocable: true
---

# Lean Live Change

Preserve contracts, data, and SLOs while keeping the common path fast:

`brief → implement → test → review → merge`

Classify first:

- Tier 1 — docs, tests, refactors, cleanup, low-risk fixes: concise task, zero
  governance artifacts, relevant tests, clean diff, truthful self-review.
- Tier 2 — feature, meaningful UX, public interface, persistence: one concise
  change brief containing problem, scope, acceptance criteria, risks, rollback;
  default-OFF flag when user-visible; tests and truthful self-review.
- Tier 3 — authorization/security, destructive behavior, material migration,
  production release, protected governance control: one brief, stronger
  verification, truthful self-review, rollback, and explicit authority at the
  real external or irreversible effect boundary.

Solo mode creates no verification/audit record and requires no independent
auditor. Team assurance is opt-in and binds actual forge identities before
review. Only pause for a missing material decision or effect-boundary authority.
Complete all ordinary repository-local reversible work first.
