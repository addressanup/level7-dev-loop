# Solo Fast Development Loop — Transition Brief

| Field | Value |
|---|---|
| Change ID | `solo-fast-development-loop` |
| Risk tier | `3` — protected workflow, controller, and plugin-instruction cutover |
| Status | `approved` by the active user for this bounded repository-local implementation |
| Base commit | `79a0739202942970536cc29782ad1b3952e7d15e` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `codex-root` |

## Problem

Level 7 reduced artifact volume but still makes a solo developer approve a
sequence of skills, wait for a mandatory independent audit, create
verification/audit commits that move the candidate identity, and discover forge
reviewer constraints only after publication. Feature-branch pushes and pull
requests also start duplicate Harness runs, while trusted policy starts before
its exact-head prerequisites can exist. The result is a leaner controller with a
slow user journey.

Risk, collaboration topology, and external-effect authority are separate
concerns. A high-risk repository-local change still needs stronger testing and
rollback reasoning, but a solo developer cannot truthfully manufacture an
independent person or GitHub identity. External, destructive, irreversible,
credentialed, production, publication, deployment, and release actions must
continue to stop at their actual authority boundary.

## Scope

Make the plugin's default experience a one-intent solo conductor:

- one natural-language request authorizes ordinary repository-local reversible
  inspection, implementation, testing, repair, and truthful self-review;
- `l7-next` conducts that complete loop and uses specialized skills internally
  without asking the user to approve skill transitions;
- solo assurance is the default and never claims independent review;
- Tier 3 remains a risk classification, but solo Tier 3 uses one brief plus
  exact-head automated verification and self-review, with no separate owner,
  verifier, or auditor identity and no verification/audit evidence commits;
- team assurance remains an explicit repository opt-in and retains distinct
  owner/auditor gates;
- release, deployment, publication, destructive, credential, and production
  effects still require explicit authority at the effect boundary;
- hosted Harness runs once for a feature candidate, and trusted policy performs
  its final evaluation after the Harness run completes;
- team-mode audit instructions resolve and bind the real forge reviewer login
  before the audit starts.

Do not change product runtime behavior, provider adapters, distribution
lifecycle behavior, secrets, deployments, repository rules, remotes, or any
historical artifact. Do not claim that solo self-review is independent.

## Exact implementation file set

Add:

- `docs/artifacts/changes/solo-fast-development-loop.md`

Modify:

- `.env.example`
- `.github/workflows/harness.yml`
- `.github/workflows/policy.yml`
- `.codex-plugin/plugin.json`
- `.claude-plugin/plugin.json`
- `AGENTS.md`
- `CHANGELOG.md`
- `README.md`
- `distribution/package.json`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`
- `marketplace.json`
- `plugin.json`
- `references/WORKFLOW.md`
- `skills/*/SKILL.md`
- `internal/harness/buildcontrol/change.go`
- `internal/harness/buildcontrol/change_test.go`
- `internal/harness/buildcontrol/main.go`
- `internal/harness/buildcontrol/policy.go`
- `internal/harness/buildcontrol/policy_test.go`
- `internal/harness/buildcontrol/testutil_test.go`

Do not add transition verification or audit records. Local checks bind to the
exact implementation head. Because GitHub evaluates trusted policy from the base
branch, the first hosted cutover may still encounter the previous team-only audit
gate. Cross that bootstrap boundary only through a specifically authorized
administrator merge/bypass of the exact CI-green head when repository rules
permit it; never fabricate an auditor or evidence record.

## Acceptance criteria

1. The default agent journey is one intent followed by inspect, implement, test,
   repair, self-review, and review-ready handoff without skill-selection or plan
   approval pauses.
2. Solo assurance is the explicit default. Reports name it, self-review is never
   described as independent, and Tier 3 solo changes require no distinct owner
   or auditor identity.
3. Solo Tier 1 uses zero governance artifacts; solo Tier 2 and Tier 3 use at most
   one concise brief. Verification and review remain Git/CI evidence and do not
   move the candidate with evidence-only commits.
4. Team assurance is opt-in through trusted configuration and preserves exact
   candidate owner/auditor separation. Its audit flow binds a real forge login
   before audit work begins.
5. Protected paths still elevate risk to Tier 3 and scope expansion still fails
   closed in both assurance modes.
6. External, destructive, irreversible, credentialed, production, publication,
   deployment, merge, and release effects still require precise authority at the
   actual effect boundary.
7. Feature-branch publication starts one Harness candidate run rather than both
   push and pull-request runs. Trusted policy receives a final exact-head trigger
   when Harness completes and remains fail-closed for missing checks or invalid
   risk metadata.
8. Existing historical Tier 3 record chains remain readable under team
   assurance for compatibility, but they are deprecated for new solo work.
9. Focused controller tests, workflow lint, manifest validation, full repository
   verification, and diff hygiene pass.
10. No current Wave 5 PR state, user-owned untracked artifact, remote, repository
    rule, deployment, or production state is changed by this implementation.

## Risks and mitigations

- **Solo under-assurance:** protected paths remain Tier 3, exact-head tests and
  scope controls remain mandatory, and external effects still stop for explicit
  authority. Output states the assurance mode and never labels self-review as
  independent.
- **Candidate weakens policy:** trusted CI continues to run the base-revision
  controller. Assurance mode comes from trusted repository configuration, not
  candidate prose.
- **Team regression:** team mode retains distinct owner, implementer, and auditor
  identities and existing bound-record compatibility tests.
- **Premature policy success:** required Harness contexts remain separate branch
  protections, and final trusted policy re-evaluates the exact head after Harness
  completion.
- **Migration ambiguity:** solo is the documented default; teams opt in
  explicitly. Historical artifacts are preserved and not extended on the solo
  path. The one-time hosted bootstrap boundary is documented as an explicit
  protected-branch action, not disguised as independent assurance.

## Rollback

Revert this cutover as one reviewed unit. Restore the previous agent routing,
team-only Tier 3 controller semantics, workflow triggers, documentation, skills,
and manifest descriptions. Historical records remain unchanged, so rollback
requires no data reconstruction or external cleanup. No remote, deployment,
production, or user configuration state is created by the implementation.
