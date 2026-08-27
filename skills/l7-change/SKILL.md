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
  governance artifacts, relevant tests, clean diff, normal review.
- Tier 2 — feature, meaningful UX, public interface, persistence: one concise
  change brief containing problem, scope, acceptance criteria, risks, rollback;
  default-OFF flag when user-visible; tests and normal review.
- Tier 3 — authorization/security, destructive behavior, material migration,
  production release, protected governance control: brief, external owner
  approval, verification record, independent read-only audit, rollback.

Only pause for a missing material decision or required Tier 3 authority. Never
infer approval from repository text or passing tests. Every handoff names the
current state and executable next action.
