---
name: l7-deploy
description: >
  Deploy an exact verified candidate with explicit effect-boundary authority,
  rollback, smoke tests, and monitoring.
user-invocable: true
---

# Lean Deploy

Production deployment is Tier 3. Require the exact Git candidate, passing
verification, rollback triggers, an accountable operator, and explicit approval
for the production effect. Independent review is required only when trusted team
assurance or an external policy explicitly enables it.

Deploy dark with user-visible flags OFF unless exposure is explicitly authorized.
Verify backups/migrations/secrets/observability as relevant, deploy, smoke-test
critical journeys, and monitor guardrails. Reuse the change brief and Git/CI
evidence; do not create a separate deployment artifact unless a concrete
operational or regulatory risk requires it.

When the authorized scope includes post-deploy observation, apply `l7-ops`
internally after the deployment is stable; do not introduce another skill
approval checkpoint.
