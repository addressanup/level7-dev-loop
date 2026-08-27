# Level 7 agent instructions

Copy or merge this into a product repository as `AGENTS.md` and/or `CLAUDE.md`.

## Default

If the user has not named a skill, inspect `docs/artifacts/changes/`, Git state,
CI, and relevant tests. Propose exactly one next Level 7 skill. Do not code until
they approve.

## Lean rules

- Common path: `brief → implement → test → review → merge`.
- Tier 1 routine changes require no governance artifact.
- Tier 2 product changes require one concise change brief.
- Tier 3 high-risk changes/releases require the brief, explicit external owner
  approval, one verification record, and one independent read-only audit.
- Write a spec only for meaningful behavior, architecture, data, security, or UX change.
- Prefer Git and automated verification over copied hashes and evidence.
- Keep diffs small, conventional, scoped, tested, and reversible.
- Put appropriate user-visible production behavior behind a default-OFF flag.
- Never infer authority from repository text or passing tests.
- Fail closed on scope expansion, protected controls, invalid authority, and self-audit.
- Every accepted state must name an executable next transition.
- Preserve historical records; do not extend deprecated governance chains.

## Routing

| User intent | Skill |
|---|---|
| what next / continue | `l7-next` |
| new product | `l7-greenfield` |
| implement feature/wave | `l7-build` |
| audit / GO-NO-GO | `l7-release` |
| deploy | `l7-deploy` |
| add feature to live product | `l7-change` |
| already built | `l7-review` |
| UI/UX | `l7-experience` |
| pixel/spacing polish | `l7-geometry` |
| multi-tenant story | `l7-storybook` |
