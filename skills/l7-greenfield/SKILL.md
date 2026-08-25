---
name: l7-greenfield
description: >
  Start or continue a greenfield product: requirements, backlog, architecture,
  technology selection, harness, and wave plan. Trigger for new project, from scratch,
  bootstrap, or missing docs/artifacts foundation files.
user-invocable: true
---

# Greenfield Foundation

Run the next incomplete foundation step only. Save each artifact under `docs/artifacts/`. Get approval before the following step.

## Sequence

1. Requirements document — problem, users, FR/NFR, constraints, metrics, risks
2. Feature backlog — P0/P1/P2, dependencies, effort, acceptance criteria
3. Architecture — 3 options, scored, selected design, failure modes
4. Technology selection — candidates, scores, stack, compatibility
5. Harness — repo layout, lint, types, tests, CI, logging, `.env.example`, README
6. Orchestration plan — waves, shared files, parallelism limits

## Rules

- Ask discovery questions one at a time for requirements.
- Do not write product features in the harness step beyond a proving test.
- Pin production dependency versions.
- Verify harness with install, lint, typecheck, and test commands.
- True parallel waves need isolated branches/agents. Otherwise sequence Wave 1 → 2 → 3.

## Output

State which step you are on, write that artifact, then ask for approval before the next step.

When the orchestration plan is approved, tell the user to run `/l7-build` for Wave 1.
