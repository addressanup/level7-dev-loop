---
name: l7-deploy
description: >
  Production deployment planning and execution checklist with rollback, smoke
  tests, and monitoring. Trigger when the user says deploy, release to
  production, or ship the audited candidate.
user-invocable: true
---

# Deploy

Do not deploy if the latest independent audit is NO-GO or has unmet CONDITIONAL GO items.

## Work

1. Confirm artifact equals audited commit/tag
2. Choose strategy: blue-green, canary, rolling, or documented alternative
3. Write rollback triggers and owner
4. Verify backups, migrations, secrets, and observability
5. Deploy only after explicit user approval
6. Smoke-test critical journeys
7. Watch error rate, latency, and business guardrails
8. Write `docs/artifacts/production-deployment-report.md`

Feature exposure is not the same as deploy. If a feature flag exists, deploy dark (flag OFF) unless the user explicitly releases it. Use `/l7-change` for progressive exposure.

Then recommend `/l7-ops`.
