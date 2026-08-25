# Level 7 agent instructions

Copy or merge this into the product repo as `AGENTS.md` and/or `CLAUDE.md`.

## Default

If the user has not named a skill, inspect `docs/artifacts/`, git state, and tests.
Propose exactly one next skill from the Level 7 Dev Loop plugin.
Do not code until they approve.

## Standing rules

- Spec before code
- Small diffs and conventional commits
- Evidence for claims
- Keep harness/CI green
- High-risk changes need independent audit
- Design changes need problem + acceptance criteria
- User-visible production behavior uses a feature flag, default OFF
- Update `docs/artifacts/` when a phase completes

## Routing

| User intent | Skill |
|---|---|
| what next / continue | l7-next |
| new product | l7-greenfield |
| implement feature/wave | l7-build |
| audit / GO-NO-GO | l7-release |
| deploy | l7-deploy |
| add feature to live product | l7-change |
| we already built it | l7-review |
| UI/UX | l7-experience |
| pixel/spacing perfection | l7-geometry |
| multi-tenant story | l7-storybook |
