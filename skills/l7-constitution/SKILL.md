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
7. Require explicit authority only at a real external or irreversible effect
   boundary; a concrete user request authorizes ordinary reversible local work.
8. Never claim that self-review is independent.

Solo artifact budget: Tier 1 has zero governance artifacts; Tier 2 and Tier 3
have at most one change brief. Git and CI carry verification and review evidence.
Team assurance is an explicit opt-in for repositories with genuinely distinct
owner and reviewer identities.

Apply these rules inside the active conductor. Do not turn them into another
approval checkpoint.
