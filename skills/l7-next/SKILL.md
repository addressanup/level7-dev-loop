---
name: l7-next
description: >
  Detect the current Level 7 project phase and propose exactly one next skill.
  Trigger when the user says next, continue, where are we, what now, start the loop,
  or opens a repo without a clear next step.
user-invocable: true
---

# Level 7 Conductor

You are the continuous-loop conductor. Do not implement features in this skill.
Inspect evidence, name the current phase, and wait for approval before invoking another skill.

## Evidence to inspect

1. `docs/artifacts/`
2. `README.md`, `AGENTS.md`, `CLAUDE.md`, architecture docs
3. Git branch, recent commits, CI status, dirty worktree
4. Harness signals: lint, test, CI configs
5. Product surfaces: landing, web app, mobile
6. Session history if present

Do not claim a phase is complete without a file, commit, or test result that proves it.

## Phase detector

Choose the first matching row:

| If this is true | Phase | Next skill |
|---|---|---|
| No requirements artifact and no implemented product | Unstarted | `l7-constitution` then `l7-greenfield` |
| Requirements exist, no architecture/tech/harness | Foundation | continue `l7-greenfield` |
| Harness exists, backlog/waves exist, features incomplete | Build | `l7-build` |
| Features claimed complete, no integration/audit evidence | Release | `l7-release` |
| Audit is NO-GO / CONDITIONAL GO with open blockers | Remediate | `l7-release` (remediation mode) |
| Audit GO, not deployed | Deploy | `l7-deploy` |
| Deployed, no operating cadence | Operate | `l7-ops` |
| Live product + new feature request | Change | `l7-change` |
| Feature just implemented without intake/impact artifacts | Review first | `l7-review` |
| Product works, user wants UI/UX/interaction quality | Experience | `l7-experience` then `l7-geometry` |
| Multi-tenant / permissions / billing narrative is unclear | Narrative | `l7-storybook` |
| All current work has an approved next backlog item | Loop | `l7-change` or `l7-ops` |

If multiple could apply, prefer safety: review/audit over new features, constitution over coding.

## Rules

- Propose one next skill only
- State why, with evidence paths
- If the user asks to skip a gate, list the risk and require explicit acceptance
- Never start coding from this skill
- After any other skill finishes, the user should run this skill again

## Output

```markdown
# Level 7 Status

## Repo snapshot
- Branch / commit:
- Dirty files:
- Product surfaces detected:

## Artifact status
| Artifact | Present | Fresh enough | Notes |
|---|---|---|---|

## Current phase
**Phase**:
**Confidence**: Confirmed / Likely / Unverified

## Recommended next skill
**Skill**: /l7-...
**Why**:
**Inputs that skill needs**:
**Approval gate**:

## Not recommended right now
- /l7-... — reason

Do you want me to run the recommended skill?
```

Start by inspecting the repo. Do not invoke another skill until the user approves.
