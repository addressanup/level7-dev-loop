---
name: l7-greenfield
description: >
  Assess, start, or resume the six-step product foundation: requirements,
  feature backlog, architecture, technology selection, harness, and
  orchestration plan. Use for a new product, an incomplete foundation, or when
  the l7-next foundation gate holds an objective.
user-invocable: true
---

# Greenfield Foundation

The product foundation is six ordered steps. Assess them in order, resume at
the first step that is not complete, and never skip ahead.

## Sequence

| Step | Default document | Complete when evidence records |
|---|---|---|
| 1. Requirements | `docs/foundation/requirements.md` | Problem, target users, functional and non-functional requirements, constraints, measurable success metrics, and risks |
| 2. Feature backlog | `docs/foundation/feature-backlog.md` | P0/P1/P2 features, each with dependencies, effort, and acceptance criteria |
| 3. Architecture | `docs/foundation/architecture.md` | Three options scored against the requirements, the selected design with its rationale, and failure modes |
| 4. Technology selection | `docs/foundation/technology-selection.md` | Scored candidates, the selected stack, compatibility, and pinned production dependency versions |
| 5. Harness | `docs/foundation/harness.md` | Repository layout, lint, type checking, tests, CI, logging, README, and `.env.example` when the product reads environment configuration, with install, lint, typecheck, and test commands verified passing |
| 6. Orchestration plan | `docs/foundation/orchestration-plan.md` | Waves, shared files, and parallelism limits |

## Status record

`docs/foundation/README.md` is the single foundation status record. For each
step it lists the status (`complete`, `partial`, or `missing`), the evidence
path for every met criterion, the remaining gaps, and the commit at which the
step was assessed complete (`none` before the first commit). Create it at the
first assessment and update it whenever a status changes.

## Assess before writing

1. For each step in order, look for evidence anywhere in the repository: the
   status record, step documents, legacy foundation documents such as an
   earlier `docs/artifacts/requirements.md`, README, code, configuration, and
   CI.
2. Mark a step `complete` only when cited evidence meets every criterion,
   `partial` when some criteria are met, and `missing` when none are. A status
   label is never evidence by itself; confirm that cited evidence still exists.
3. Re-assess a complete step when its evidence changed after the commit at
   which it was assessed, or when an earlier step changed.
4. Reference evidence that already satisfies a criterion instead of rewriting
   it. Write or extend a step document only for missing criteria.

Evidence decides completion; do not ask the user to sign off a finished step.

## Rules

- Never invent product intent. Ask requirements discovery questions one at a
  time; unanswered items remain gaps.
- Pause for material product, architecture, technology, data, security, or UX
  decisions, such as choosing the architecture option or the stack.
- Do not write product features in the harness step beyond a proving test.
- Pin production dependency versions.
- Verify the harness with install, lint, typecheck, and test commands.
- True parallel waves need isolated branches or agents; otherwise sequence
  Wave 1 → 2 → 3.
- Keep foundation documents under `docs/foundation/`, never under
  `docs/artifacts/`. They are Tier 1 documentation. Classify harness setup and
  every later build normally: Tier 1 needs no artifact; Tier 2 and Tier 3 use
  one `docs/artifacts/changes/<change-id>.md` brief.

## Output

Report the status of all six steps, name the step in progress, write or extend
only that step's missing criteria, and update the status record. When all six
steps are complete, return to the objective that `l7-next` held. When invoked
directly, continue into `l7-build` for Wave 1 if the user's objective
authorizes building.
