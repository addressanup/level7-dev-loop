---
name: l7-deploy
description: >
  Deploy an audited Tier 3 candidate with explicit approval, rollback, smoke
  tests, and monitoring.
user-invocable: true
---

# Lean Deploy

Production deployment is Tier 3. Require the exact Git candidate, external owner
approval, passing verification, independent GO audit, rollback triggers, and an
accountable operator.

Deploy dark with user-visible flags OFF unless exposure is explicitly authorized.
Verify backups/migrations/secrets/observability as relevant, deploy, smoke-test
critical journeys, and monitor guardrails. Reuse the change brief and verification
record; do not create a separate deployment artifact unless a concrete
operational or regulatory risk requires it.

Recommend `l7-ops` after the deployment is stable.
