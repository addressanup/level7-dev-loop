---
name: l7-storybook
description: >
  Write the Product Storybook and end-to-end system narrative: personas,
  tenant isolation, collaboration, feature gating, worst-day errors, and
  deletion. Trigger for storybook, multi-tenant narrative, permissions story,
  or align on product behavior.
user-invocable: true
---

# Product Storybook

Translate architecture and code into a chronological human narrative. Do not invent rules. Mark UNVERIFIED or GAP when evidence is missing.

## Required sections

1. Persona matrix: platform admin, Tenant A enterprise, Tenant B free, owner / power / restricted roles, permission table
2. Onboarding, branding, tenant isolation, including cross-tenant URL access
3. Invite lifecycle, simultaneous edit, real-time update
4. Feature gating and API-level block for free-tier bypass
5. Network drop + double-click idempotency; noisy-neighbor isolation
6. Cancel + delete space: files, rows, logs, backups, grace period
7. Contradictions, unverified questions, product decisions

Use concrete names and Given/When/Then for complex logic.
Write `docs/artifacts/product-storybook.md`.
Do not implement changes until the user approves decisions in section 7.

Then `/l7-next`.
