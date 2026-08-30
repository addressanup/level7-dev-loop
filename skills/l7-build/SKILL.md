---
name: l7-build
description: >
  Implement one bounded change continuously through code, tests, repair, and
  truthful self-review with controls proportional to risk.
user-invocable: true
---

# Lean Build Loop

Implement one bounded change.

1. Inspect Git, tests, and the declared scope; classify Tier 1/2/3.
2. Tier 1: use the task/PR description and create no governance artifact.
3. Tier 2/3: use exactly one concise change brief. Solo Tier 3 adds stronger
   verification and rollback analysis, not a manufactured independent auditor.
4. Implement in small conventional commits. Add a default-OFF feature flag for
   appropriate user-visible production behavior.
5. Run relevant tests, lint/types, CI, and inspect the final diff.
6. Self-review truthfully and finish the complete local loop. Use `l7-release`
   only when team assurance or a real release decision explicitly requires it.

Do not create separate specification, design, candidate manifest, approval,
verification, audit, evidence, or history files in solo mode. Git and CI carry
identity and evidence. Pause only at an actual external-effect boundary.
