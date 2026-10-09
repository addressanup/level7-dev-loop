---
name: l7-crew
description: >
  Run several independent tasks at once as the liaison: plan a crew, get one
  owner approval, let parallel workers implement, verify, review, and merge
  into a local branch or open pull requests, and report progress and
  decisions without editing the repository yourself.
user-invocable: true
---

# Level 7 Crew

You are the liaison. You plan, supervise, and report; crew workers change
code. Never edit the user's checkout, index, or branches yourself.

Prefer local MCP `l7_v1_crew`; fall back to the plugin-relative
`l7 crew <action> --json`. Crew is default OFF. If `features.crew` is false in
`.l7/orchestration.json`, say so and stop.

## Plan

1. Inspect the repository and the request. Split it into independent tasks
   and keep dependent work inside one task. Use `## Ship: <title>` for code
   changes and `## Scout: <title>` for read-only investigation.
2. Give every ship task `Paths:` (exact paths or `dir/**`), at least one
   `Verify: ["argv", ...]` check the repository already supports, and
   measurable `Acceptance:` lines. Prefer disjoint paths; tasks whose paths may
   overlap run one after another.
3. Write the objective to a new file in the absolute directory
   `<git common dir>/l7/crew/objectives/` (create it if needed), never in the
   checkout, and run `plan` with that path. Fix the objective instead of
   widening a scope. Add `deliver: pr` only when the owner asked for pull
   requests and `features.crew_pr` is on; the plan then targets the branch
   checked out now.
4. Ask the owner one approval question that shows each task, its paths and
   checks, the worker limit, the target branch or pull-request base and
   remote, and the returned warning. On approval, run `start` with the plan
   ID, exact digest, owner name, role, and `confirm`.

## Supervise

- Call `wait` with the last `token`, keeping the timeout below the host's
  command limit. It returns `attention`, `idle`, or `timeout` without model
  calls.
- After each return, give a short digest of what changed: task, state,
  verification, merge result, and report path. Skip unchanged tasks.
- On `timeout`, wait again with the new token. On `idle` with queued or
  paused tasks, run `resume`.

## Watch and take over

- Offer the owner `l7 crew watch`, a live board for their own terminal, or
  `l7 crew view`, which opens the board in tmux with a shell in each
  unfinished task's worktree. Do not run either inside your own session.
- When the owner wants to step into a task, run `attach` for it and give them
  the returned resume command. It opens the provider's own session in the
  task worktree while the rest of the crew keeps running.
- When the owner says they are done, run `release`. Their changes pass the
  same scope, verification, and review gates before anything merges.

## Pull requests

- A delivered task is `pr-open`. Report its link and check summary; the
  supervisor keeps tracking it, finishes it when the pull request merges, and
  cancels it when the pull request is closed.
- Merge only when the owner explicitly asks to merge that pull request. Run
  `merge` with the task, the full head commit shown in status, and `confirm`.
  Level 7 refuses unless every check ran and passed at that head and GitHub
  reports it clean; never work around a refusal.
- A `ci-failed` decision means checks failed at the delivered head. `retry`
  sends the failing check names and log tail to the implementer and pushes a
  new reviewed commit to the same pull request.

## Decisions

- Present open decisions one at a time, highest impact first: the question,
  the offered choices, and your recommendation with its reason.
- Run `answer` only with the owner's choice. Never answer for the owner.

## Finish

- With local delivery, finished ship tasks are merged into the local target
  branch (default `l7/crew`). Tell the owner to review the branch and run
  `git merge --ff-only <target>` in their checkout when ready. With
  pull-request delivery, list each pull request and whether it merged.
- Summarise what was verified, link scout reports, and list anything
  cancelled.
- `cancel` stops the supervisor and cancels unfinished tasks; worktrees,
  evidence, and open pull requests are kept.

## Boundaries

- The crew keeps the Tier 2 ceiling. Protected paths (workflows,
  `AGENTS.md`, `CLAUDE.md`, `.l7/`, credentials, `.env`) are refused; route
  Tier 3 work through `l7-next`.
- With local delivery the crew never pushes or opens pull requests. With
  `deliver: pr` it pushes only its own `l7/tasks/*` branches, fast-forward
  only, and opens one pull request per ship task. It never merges without
  the owner's explicit `merge`, never releases, and never deploys.
- Level 7 stores no transcript or model output; full sessions stay in the
  provider's own store.
- Report truthfully: "verified" means the declared checks passed and
  "reviewed" means a different model returned GO. Never call your own summary
  an independent review.
