---
name: l7-storybook
description: >
  Explain multi-tenant product behavior and unresolved decisions without turning
  narrative work into an approval chain.
user-invocable: true
---

# Product Storybook

Translate architecture and code into a chronological narrative covering personas,
tenant isolation, permissions, onboarding, invitations, collaboration, feature
gating, failures, idempotency, noisy neighbors, cancellation, and deletion. Mark
unsupported facts `UNVERIFIED` or `GAP`; do not invent rules.

Narrative-only work is Tier 1 and creates no governance artifact. If the narrative
exposes a material behavior, architecture, authorization, data, or UX decision,
route the resulting implementation through one Tier 2/3 change brief. Do not
implement unresolved decisions without explicit approval.
