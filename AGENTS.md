# Level 7 agent instructions

Copy or merge this into a product repository as `AGENTS.md` and/or `CLAUDE.md`.

## Default conductor

Treat one natural-language objective as authorization for ordinary
repository-local, reversible work within its stated or reasonably implied
scope. Run the complete loop without asking the user to select or approve skill
transitions:

`inspect → implement → test → repair → self-review → review-ready handoff`

Use `l7-next` as the conductor even when the user does not name a skill. It may
apply the specialized Level 7 skills internally, but the skill graph is not a
user interface. Do not stop after proposing a plan when implementation was
requested.

Pause only for a material product decision or missing authority at an external,
destructive, irreversible, credentialed, production, publication, deployment,
release, or protected-branch merge boundary. Consolidate that boundary into one
precise question.

## Assurance and risk

Solo assurance is the default. Set trusted `L7_ASSURANCE_MODE=team` only when a
repository actually has distinct accountable-owner and reviewer identities.

- Tier 1 routine changes use the task/PR, relevant tests, a clean diff, and zero
  governance artifacts.
- Tier 2 product changes use one concise change brief, tests, and truthful
  self-review.
- Tier 3 high-risk repository changes use one concise brief, stronger relevant
  verification, truthful self-review, rollback reasoning, and explicit authority
  at any real effect boundary.
- Solo mode never requires or claims an independent audit. It creates no tracked
  verification or audit records; Git and CI are the evidence.
- Team mode may require a distinct owner and independent reviewer. Resolve the
  real forge login before review and bind review evidence to the exact head.

Risk and collaboration topology are separate: protected controls still elevate
to Tier 3, but they do not manufacture a second person for a solo developer.

## Lean rules

- Keep diffs small, conventional, scoped, tested, and reversible.
- Write a spec only for meaningful behavior, architecture, data, security, or UX
  changes.
- Put appropriate user-visible production behavior behind a default-OFF flag.
- Prefer Git and automated verification over copied hashes and evidence commits.
- Never represent self-review as independent review.
- Fail closed on scope expansion, invalid authority, destructive effects, and
  fabricated success.
- In CI, require one trusted risk-tier label; apply it automatically when PR
  publication is authorized.
- Preserve historical records; do not extend deprecated governance chains on the
  solo path.

## Specialized skills

`l7-next` is the only default entry point. Use these as internal execution lenses
without asking the user to approve the routing decision:

| Intent | Internal skill |
|---|---|
| new product | `l7-greenfield` |
| implement feature/wave | `l7-build` |
| optional team/release audit | `l7-release` |
| deploy | `l7-deploy` |
| add feature to live product | `l7-change` |
| review existing work | `l7-review` |
| UI/UX | `l7-experience` |
| pixel/spacing polish | `l7-geometry` |
| multi-tenant story | `l7-storybook` |
