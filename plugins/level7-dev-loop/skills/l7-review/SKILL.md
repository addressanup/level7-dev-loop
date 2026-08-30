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

- Tier 1: run targeted checks and truthful self-review; create no governance
  artifact.
- Tier 2: ensure one concise brief exists and matches the implementation.
- Tier 3 solo: run stronger relevant checks, adversarial self-review, and inspect
  rollback/effect boundaries without requiring a separate auditor.
- Tier 3 team: use a distinct real forge reviewer only when trusted team
  assurance is configured.

Report only evidenced gaps with severity and exact remediation. Do not run a full
compliance audit for documentation, tests, refactors, cleanup, or ordinary
features. Never describe self-review as independent. Do not remediate unless the
user requested a fix.
