# l7-next Foundation Gate — Change Brief

| Field | Value |
|---|---|
| Change ID | `l7-next-foundation-gate` |
| Risk tier | `3` — protected plugin-instruction and workflow-contract change |
| Status | `approved` by the active user for this bounded repository-local implementation |
| Base commit | `245d5140f6dea7cc062c11bd51d499b22077c5f6` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `devin` |
| Assurance | `solo` — self-review only; no independent audit is claimed |

## Problem

`l7-next` starts implementing whatever objective it receives without checking
whether the product has a foundation to build on. The lean `l7-greenfield`
lists foundation topics but has no order, completion criteria, or progress
record, so nothing can tell whether a foundation is complete or where to resume
it. The original six-step sequence (commit `08c38b6`) had that order but was
removed by the lean rewrite.

## Scope

Make the product foundation a hard prerequisite of the default conductor:

- `l7-greenfield` again defines the original six ordered steps — requirements,
  feature backlog, architecture, technology selection, harness, and
  orchestration plan — with their original depth, explicit completion
  criteria, and the original rules;
- every step is assessed against existing repository evidence before anything
  is written; satisfying evidence is referenced rather than rewritten, and a
  step is complete when cited evidence meets all of its criteria, with no
  per-step sign-off;
- `docs/foundation/README.md` is the single foundation status record, and new
  foundation documents live under `docs/foundation/`, outside the
  `docs/artifacts/` governance budget;
- `l7-next` runs a foundation gate before any repository change. While any step
  is incomplete it blocks work of every tier, including routine fixes, reports
  each step's status, and applies `l7-greenfield` from the first incomplete
  step. It resumes the original objective only when all six steps are complete;
- read-only questions and the foundation work itself are not blocked. The gate
  belongs to `l7-next`; directly invoked specialized skills are not gated;
- `AGENTS.md`, the workflow reference, README, and changelog describe the gate.

Do not change Go code, the `l7` CLI or MCP behavior, CI, the policy controller,
distribution packaging, the frozen v0.1.1 payload under `plugins/`, historical
records, or the user-owned untracked audit file.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-next-foundation-gate.md`

Modify:

- `skills/l7-next/SKILL.md`
- `skills/l7-greenfield/SKILL.md`
- `AGENTS.md`
- `references/WORKFLOW.md`
- `README.md`
- `CHANGELOG.md`

## Acceptance criteria

1. Before any repository change, `l7-next` assesses the six `l7-greenfield`
   steps in order from `docs/foundation/README.md` and current repository
   evidence.
2. While any step is partial or missing, `l7-next` makes no repository change
   outside the foundation, for every risk tier. Read-only answers remain
   available.
3. A blocked run reports every step's status with evidence or gaps, names the
   first incomplete step, and continues `l7-greenfield` from that step without
   redoing complete steps or skipping ahead.
4. The user's original objective resumes only after all six steps are complete.
5. `l7-greenfield` defines the six steps with completion criteria matching the
   original sequence: requirements (problem, users, functional and
   non-functional requirements, constraints, metrics, risks); feature backlog
   (P0/P1/P2, dependencies, effort, acceptance criteria); architecture (three
   scored options, selected design, failure modes); technology selection
   (candidates, scores, stack, compatibility, pinned production versions);
   harness (layout, lint, types, tests, CI, logging, environment example,
   README, verified commands); orchestration plan (waves, shared files,
   parallelism limits).
6. Existing evidence anywhere in the repository can satisfy a criterion, and
   every criterion marked met cites its evidence path. Completion requires no
   per-step sign-off, but product intent is never invented and material
   decisions still pause for the user.
7. The status record lists each step's status, evidence paths, gaps, and the
   commit at which it was assessed complete. A status label without existing
   evidence never satisfies the gate; evidence changed since that commit, or a
   changed earlier step, triggers re-assessment.
8. Foundation documents are written under `docs/foundation/`, never under
   `docs/artifacts/`.
9. `AGENTS.md`, `references/WORKFLOW.md`, README, and the changelog describe the
   gate consistently, and README no longer implies that a first `l7-next` run on
   a repository without a foundation goes straight to implementation.
10. The frozen v0.1.1 payload and its pinned package digests, all Go code, CI,
    the policy controller, historical records, and the untracked audit file are
    unchanged.
11. `make policy-check`, `make verify`, and `git diff --check` pass on the exact
    implementation candidate.

## Risks and mitigations

- **Adoption friction (owner-accepted):** the first `l7-next` run in any
  repository without a complete foundation — including established products
  and this repository — stops all other work until all six steps are complete.
  Brownfield assessment reuses existing code, CI, and documents, so established
  repositories may already satisfy several steps; only missing criteria are
  written.
- **Self-certified foundation (owner-accepted):** evidence, not sign-off,
  decides completion. Criteria are concrete, every met criterion cites a path,
  product intent comes only from the user or existing evidence, and material
  product, architecture, and technology choices still pause for the user.
- **Bypass via status text:** a status label is never evidence. The gate checks
  that evidence still exists and re-assesses steps whose evidence changed since
  the commit at which they were assessed complete.
- **Urgent work blocked:** the gate lives in `l7-next` only. Directly invoking a
  specialized skill, such as `l7-ops` for incident containment, is not gated.
- **Budget collision:** foundation documents stay outside `docs/artifacts/`, so
  the policy controller's artifact budget does not count them.
- **Package divergence:** v0.1.1 rollback skills stay frozen; only canonical v1
  skills gain the gate, as recorded in the changelog.
- **Engine unaware of the gate:** `l7 onboard --status` and `l7 headless` do not
  evaluate the foundation; engine-level enforcement would be a separate change.

## Rollback

Revert the implementation commit, then this brief commit. No data migration,
external state, package digest, or release is involved. Foundation documents
that the new flow later creates in product repositories are ordinary
documentation that users may keep or delete.
