---
name: l7-greenfield
description: >
  Establish the minimum product foundation needed to build safely: problem,
  acceptance criteria, architecture decisions, stack, harness, and backlog.
user-invocable: true
---

# Lean Greenfield Foundation

Discover enough to build, not a document chain. Capture the product problem,
users, constraints, success measures, risks, architecture decisions, technology,
test/CI harness, and first backlog in the smallest useful product documents.

Before implementation, classify the first build:

- Tier 1: no governance artifact.
- Tier 2: one `docs/artifacts/changes/<change-id>.md` brief.
- Tier 3: the same brief plus external owner approval before building.

Do not require separate approval after every foundation section. Pause only for a
material product, architecture, data, security, or UX decision. Keep the harness
green and recommend `l7-build` when the first scoped change is ready.
