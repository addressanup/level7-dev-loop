---
name: l7-review
description: >
  Review an ad-hoc implementation with targeted, risk-proportionate checks and
  escalate only concrete high-risk gaps.
user-invocable: true
---

# Lean Change Review

Inspect the diff, task intent, tests, CI, feature flags, and rollback. Classify the
change Tier 1/2/3 from actual impact.

- Tier 1: run targeted checks and normal review; create no governance artifact.
- Tier 2: ensure one concise brief exists and matches the implementation.
- Tier 3: require external owner approval, bound verification, and independent
  read-only audit through `l7-release`.

Report only evidenced gaps with severity and exact remediation. Do not run a full
compliance audit for documentation, tests, refactors, cleanup, or ordinary
features. Do not remediate unless the user requested a fix.
