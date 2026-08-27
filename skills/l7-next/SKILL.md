---
name: l7-next
description: >
  Inspect Git, tests, CI, and current change records, then propose exactly one
  next Level 7 skill without creating ceremony.
user-invocable: true
---

# Lean Conductor

Inspect Git status/history, relevant tests and CI, `docs/artifacts/changes/`, and
product context. Identify the current change and its risk tier.

Route exactly one next skill:

- New product or missing foundation: `l7-greenfield`
- Approved feature/change to implement: `l7-build` or `l7-change` if live
- Candidate needing release or Tier 3 validation: `l7-release`
- Production deployment: `l7-deploy`
- Post-launch operations: `l7-ops`
- Existing ad-hoc implementation: `l7-review`
- UI/UX work: `l7-experience`; spacing-only polish: `l7-geometry`
- Multi-tenant narrative: `l7-storybook`

Report evidence, current state, one next action, and why. Do not code or create an
artifact. Wait for approval before invoking the next skill.
