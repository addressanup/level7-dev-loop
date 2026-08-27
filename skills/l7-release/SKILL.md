---
name: l7-release
description: >
  Validate a production release or other Tier 3 candidate with bound verification,
  an independent read-only audit, and a GO or NO-GO decision.
user-invocable: true
---

# Tier 3 Release Validation

Use only for production releases or genuinely high-risk changes.

1. Confirm the approved brief, scope, rollback, and exact Git candidate.
2. Run integration, data, security, migration, performance, and regression checks
   relevant to the risk. Write at most one verification record bound to Git.
3. In a separate read-only review, map the candidate, cite evidence, classify
   findings, and issue GO or NO-GO. Write at most one audit record bound to the
   verified commit/tree.
4. The auditor must be distinct from the implementer and owner and cannot
   self-certify remediation. New implementation commits invalidate verification
   and audit and return the change to `building`.

Do not create candidate manifests or approval/remediation chains. Recommend
`l7-deploy` only after a bound GO decision.
