# Level 7 workflow

The default path is:

`brief → implement → test → review → merge`

## Risk tiers

| Tier | Planning artifact | Gates |
|---|---:|---|
| 1 — routine | 0 | Relevant tests and normal review |
| 2 — product | 1 change brief | Tests, normal review, default-OFF feature flag when appropriate |
| 3 — high risk/release | At most 3: brief, verification, audit | External owner approval and independent read-only audit |

Tier 3 covers production releases, authorization/security boundaries, destructive
or irreversible behavior, material migrations, and protected governance controls.

## States

Tier 1/2:

`planned → building → verified → reviewed → ready → merge`

Tier 3:

`planned → awaiting-owner-approval → building → verified → awaiting-independent-audit → reviewed → ready → merge`

New implementation commits after verification/review return to `building`. Audit
failure returns to `building`. Every state reports a concrete next action.

In CI, risk comes from exactly one maintainer-controlled `l7-risk-tier-1/2/3`
label. Tier 1 scope comes from an explicit `L7-Scope:` PR metadata field, never
from the candidate diff. Exact-head Harness success and non-author review are
required before Tier 1/2 can become `ready`; Tier 3 additionally requires owner
approval and a distinct bound auditor.

Use `l7-next` for routing. Use `l7-change` for live work, `l7-build` for an
approved build, `l7-release` only for Tier 3/release validation, and `l7-deploy`
after a bound GO decision.

Historical phase manifests and approval/audit chains remain in Git as records but
are deprecated inputs. New work does not update them.
