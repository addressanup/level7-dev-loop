---
name: l7-build
description: >
  Implement one approved change through the lean brief, implement, test, review,
  merge path with controls proportional to risk.
user-invocable: true
---

# Lean Build Loop

Implement one bounded change.

1. Inspect Git, tests, and the declared scope; classify Tier 1/2/3.
2. Tier 1: use the task/PR description and create no governance artifact.
3. Tier 2/3: use exactly one concise change brief. Tier 3 must have external
   accountable-owner approval before implementation.
4. Implement in small conventional commits. Add a default-OFF feature flag for
   appropriate user-visible production behavior.
5. Run relevant tests, lint/types, CI, and inspect the final diff.
6. Use normal review for Tier 1/2. Route Tier 3 to `l7-release` for independent audit.

Do not create separate specification, design, candidate manifest, approval,
evidence, or history files. Git and CI carry identity and routine evidence.
