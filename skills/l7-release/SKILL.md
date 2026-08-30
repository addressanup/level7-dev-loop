---
name: l7-release
description: >
  Validate a real release boundary or perform opt-in team assurance without
  imposing an independent-audit gate on solo development.
user-invocable: true
---

# Release and Optional Team Assurance

Use this skill for an actual release decision or when trusted repository
configuration explicitly selects team assurance. It is not a mandatory stop for
solo Tier 3 repository work.

1. Confirm the exact Git candidate, scope, relevant risks, rollback, and the
   specific release or team-assurance boundary.
2. Run integration, data, security, migration, performance, and regression checks
   proportional to that boundary. Use Git and CI evidence; do not add a tracked
   verification commit in solo mode.
3. In solo mode, perform adversarial self-review and label it `self-review`; never
   call it independent. A release/deploy/publication effect still requires the
   user's explicit authority.
4. In team mode, resolve the real forge reviewer login before audit work starts.
   The reviewer must be distinct from owner and implementer, review the exact
   head read-only, and use that same login in any bound decision record.

Do not fabricate reviewer identities, create evidence-only candidate commits on
the solo path, or rerun unchanged checks. A new implementation commit invalidates
the corresponding exact-head evidence.
