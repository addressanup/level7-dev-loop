---
name: l7-review
description: >
  After an ad-hoc feature implementation, first check whether the Level 7 workflow
  was already followed. Only then run a full compliance audit if needed. Trigger
  when a feature was just built without the formal loop, or the user asks did you
  follow the process.
user-invocable: true
---

# Review Before Assuming Process Failure

Default to context check. Do not modify code in that mode.

## Mode A — Context check (default)

Inspect session context, artifacts, git history, tests, flags, and CI.
Classify each area: Evidenced / Partial / Not evidenced / N/A.

Recommend exactly one:

- NO ADDITIONAL AUDIT NEEDED
- RUN TARGETED GAP CHECKS
- RUN FULL COMPLIANCE AUDIT
- PAUSE OR DISABLE UNTIL REVIEWED

Wait for approval.

## Mode B — Full compliance audit

Only if Mode A recommended it or the user approved it.

Score requirements, architecture, spec, implementation discipline, tests, flags, observability, docs, independent audit, and outcome plan.

Classify FULLY COMPLIANT / MINOR GAPS / NON-COMPLIANT / UNVERIFIED.
Produce a remediation plan. Do not remediate until approved.

High-risk areas: auth, authorization, migrations, payments, PII, tenancy, deletion, AI tools.

Then recommend `/l7-change` for rollout or `/l7-release` if Tier 3.
