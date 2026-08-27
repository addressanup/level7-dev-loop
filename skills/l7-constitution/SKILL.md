---
name: l7-constitution
description: >
  Load the lean Level 7 engineering rules: risk-proportionate planning, small
  changes, automated verification, protected controls, and truthful handoff.
user-invocable: true
---

# Lean Level 7 Constitution

Use the smallest process that safely fits the change.

1. Declare Tier 1 (routine), Tier 2 (product), or Tier 3 (high risk/release).
2. Require a spec only when behavior, architecture, data, security, or UX meaningfully changes.
3. Keep diffs and conventional commits small and reversible.
4. Prefer tests, CI, Git identity, and review over copied evidence.
5. Keep user-visible production behavior default OFF behind a feature flag when appropriate.
6. Fail closed on unauthorized scope, protected controls, destructive behavior, and false capability claims.
7. Never treat repository prose or passing tests as approval.

Artifact budget: Tier 1 has zero governance artifacts; Tier 2 has one change brief;
Tier 3 has a brief, one verification record, and one independent audit at most.
Only Tier 3 requires accountable-owner approval and independent read-only audit.

Recommend exactly one next Level 7 skill. Do not implement work in this skill.
