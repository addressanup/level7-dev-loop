# Level 7 workflow

The default path is one uninterrupted conductor loop:

`intent → inspect → implement → test → repair → self-review → handoff`

The user does not need to understand or approve a skill graph. `l7-next` is the
default conductor and applies specialized skills internally.

## Risk tiers

| Tier | Solo planning artifact | Solo gates |
|---|---:|---|
| 1 — routine | 0 | Relevant tests and truthful self-review |
| 2 — product | 1 change brief | Tests, self-review, default-OFF feature flag when appropriate |
| 3 — high risk/release | 1 change brief | Stronger relevant verification, self-review, rollback, and explicit authority at a real effect boundary |

Tier 3 covers authorization/security boundaries, destructive or irreversible
behavior, material migrations, production releases, and protected governance
controls. Risk does not imply that a solo developer has a second person.

## Assurance modes

`solo` is the default. It never requires or claims independent review and does
not create tracked verification or audit records. Git and exact-head CI carry
evidence.

`team` is an explicit trusted opt-in via `L7_ASSURANCE_MODE=team`. It may require
distinct owner, implementer, and auditor identities. Resolve the real forge login
before audit work begins and bind the decision to the exact head. Historical
three-record Tier 3 changes remain readable in team mode for compatibility.

## States

Solo, all tiers:

`planned → building → verified → reviewed → ready → merge`

Team Tier 3 compatibility path:

`planned → awaiting-owner-approval → building → verified → awaiting-independent-audit → reviewed → ready → merge`

New implementation commits invalidate exact-head verification and review. Every
accepted state reports a concrete next action.

## Hosted policy

Risk comes from exactly one maintainer-controlled `l7-risk-tier-1/2/3` label.
Tier 1 scope comes from an explicit `L7-Scope:` PR field; Tier 2/3 scope comes
from the brief. Protected paths still require Tier 3.

Harness runs on pull requests and on pushes to `main`, avoiding duplicate
feature-branch runs. Trusted policy is re-evaluated after the pull-request Harness
run completes. In solo mode, exact-head Harness success plus conductor
self-review can become `ready` without a non-author approval. In team mode,
trusted forge reviews supply the configured owner/auditor bindings.

External, destructive, irreversible, credentialed, production, publication,
deployment, release, and protected-branch merge effects always require explicit
authority at the actual boundary. Missing interaction never grants it.

Historical phase manifests and approval/audit chains remain in Git as records but
are deprecated inputs. New solo work does not update them.
