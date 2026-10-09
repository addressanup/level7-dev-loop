# L7 Crew Core — Change Brief

| Field | Value |
|---|---|
| Change ID | `l7-crew-core` |
| Risk tier | `3` — new skill, workflow-contract change, and concurrent autonomous local execution |
| Status | `approved` by the active user for this bounded repository-local implementation |
| Base commit | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `devin` |
| Assurance | `solo` — self-review only; no independent audit is claimed |
| Roadmap | Phase 1 of 4: crew core (this brief), visible crew, PR delivery, auto-merge on green |

## Problem

Level 7 runs one objective at a time. `l7-next` works inline in the user's
checkout, and `l7 headless` runs waves strictly in sequence behind a single
repository-wide checkpoint (`headless.LoadCheckpoint` reads one `active.json`).
A user with several independent tasks must run them one after another or juggle
sessions by hand.

Inside a Headless wave there is also no repair loop. The Claude worker has no
shell (`sessionArguments` disallows `Bash`), and a failed verification switches
to the next provider with the original prompt and none of the failure output.
After three identical failures the run pauses.

## Scope

Add an opt-in crew: the user talks to one liaison, which plans independent
tasks, runs several at once in disposable worktrees, supervises them, and
reports results and open decisions.

- A default-OFF `features.crew` flag. When false it is omitted from
  `.l7/orchestration.json`, so older strict decoders still load the file.
- Two task shapes. `ship` implements in a disposable worktree, verifies,
  obtains review from a different provider/model, and fast-forwards a local
  target branch. `scout` investigates read-only and leaves a private report; it
  never commits or merges.
- Durable per-task state under `.git/l7/crew/`, independent of Headless state,
  with restart recovery for every task.
- A scheduler that runs at most `crew.max_workers` tasks at once (default 3,
  range 1–4). Ship tasks whose path scopes may overlap never run at the same
  time; the later one waits.
- An in-task repair loop. After each implementer turn Level 7 runs the task's
  exact allowlisted verification argv. On failure it resumes the same session
  with the failing command, exit code, and a bounded output tail, up to
  `crew.repair_rounds` times (default 2), before counting a failure and failing
  over. Workers still never get a shell.
- A serialized merge queue. If the target moved since the task started, the
  candidate is rebased onto the current target, verification re-runs at the new
  head, and review re-runs when the rebased patch differs. A rebase conflict
  becomes an open decision; nothing resolves conflicts automatically.
- One supervisor process per repository, started detached by `l7 crew start`
  and stopped by `l7 crew cancel`. It exits when no runnable task remains.
- Liaison surfaces with matching CLI and MCP (`l7_v1_crew`) actions: `plan`,
  `start`, `status`, `wait`, `decisions`, `answer`, `resume`, and `cancel`.
  `wait` blocks inside the engine until a task needs attention or the timeout
  passes, with no model calls.
- A new `l7-crew` skill for the liaison, plus `AGENTS.md`, workflow, README, and
  changelog updates.
- A fix to the shared Codex app-server adapter found by the first live crew
  run: codex-cli 0.162 rejects camelCase `thread/start` and `thread/resume`
  sandbox modes, so every Codex session failed (Headless included). Send the
  kebab-case modes the server requires and keep its bounded error message.

Out of scope for this phase: tmux views and session takeover, pushing, pull
requests, any remote or forge operation, automatic merging on a forge, hooks,
away mode, and remote hosts. Headless behavior does not change apart from the
shared Codex adapter fix above. The frozen
v0.1.1 payload under `plugins/` and `distribution/package.json` do not change.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-crew-core.md`
- `docs/foundation/README.md`
- `internal/l7/domain/crew.go`
- `internal/l7/domain/crew_test.go`
- `internal/l7/adapter/crew/plan.go`
- `internal/l7/adapter/crew/plan_test.go`
- `internal/l7/adapter/crew/store.go`
- `internal/l7/adapter/crew/store_test.go`
- `internal/l7/adapter/crew/engine.go`
- `internal/l7/adapter/crew/engine_test.go`
- `internal/l7/adapter/crew/supervisor.go`
- `internal/l7/adapter/crew/supervisor_test.go`
- `internal/l7/adapter/headlessworker/crew.go`
- `internal/l7/adapter/headlessworker/crew_test.go`
- `cmd/l7/crew_cli.go`
- `cmd/l7/crew_cli_test.go`
- `skills/l7-crew/SKILL.md`

Modify:

- `internal/l7/adapter/codexapp/worker.go`
- `internal/l7/adapter/codexapp/worker_test.go`
- `internal/l7/adapter/process/process_unix.go`
- `internal/l7/adapter/process/process_test.go`
- `internal/l7/adapter/verify/runner.go`
- `internal/l7/adapter/verify/runner_test.go`
- `internal/l7/adapter/orchestrationconfig/config.go`
- `internal/l7/adapter/orchestrationconfig/config_test.go`
- `cmd/l7/orchestration_cli.go`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `cmd/l7pack/main.go`
- `internal/harness/v1candidate/main.go`
- `AGENTS.md`
- `references/WORKFLOW.md`
- `README.md`
- `CHANGELOG.md`

## Acceptance criteria

1. `features.crew` defaults OFF and is omitted from the policy file while
   false. Every mutating `l7 crew` action fails closed while it is OFF.
2. `l7 crew plan` records each task's shape, objective, acceptance criteria,
   target branch, and digest; ship tasks also record allowed paths and exact
   verification argv, and scout tasks record no write scope. `l7 crew start`
   requires the plan digest, owner, role, and `--confirm`.
3. No more than `crew.max_workers` tasks run at once. Ship tasks with possibly
   overlapping scopes never run concurrently, and scouts never block ship
   tasks.
4. Each task runs in its own worktree and branch under `.git/l7/crew/`. The
   user's checkout and index are never modified. A change outside a task's
   allowed paths, or to a protected path, stops only that task.
5. A failed verification resumes the same session with bounded failure
   context up to `crew.repair_rounds` times before the task counts a failure
   and fails over to the next qualified provider.
6. Every ship candidate needs a `GO` from a reviewer route whose model differs
   from the implementer's before it can merge. At this Tier 2 ceiling the
   reviewer may come from the same provider, as in Headless routing.
7. Merges are serialized. A moved target triggers rebase, re-verification at
   the new head, and re-review when the patch changed. Conflicts become open
   decisions. The target branch only advances by compare-and-swap
   fast-forward and is never checked out.
8. Scout tasks leave a private report and never commit or merge.
9. Killing the supervisor and running `l7 crew resume` continues every
   unfinished task from its last checkpoint without repeating completed
   stages. A second supervisor for the same repository refuses to start.
10. `l7 crew wait --timeout <seconds>` returns as soon as any task needs
    attention, or at the timeout, and makes no model calls. `l7 crew status`
    lists every task's shape, state, route, worktree, candidate, verification
    result, and next action. `l7 crew decisions` lists open decisions by impact
    and `l7 crew answer` resolves one.
11. No push, pull request, remote, release, or deployment action exists in
    this phase.
12. CLI and MCP expose the same crew actions. The `l7-crew` skill tells the
    liaison never to edit the repository itself, to get one approval for the
    plan, to use `wait` between updates, and to walk decisions one at a time.
13. Existing Headless behavior and tests are unchanged, except that the shared
    Codex adapter now sends the thread sandbox modes codex-cli 0.162 accepts.
14. `AGENTS.md`, `references/WORKFLOW.md`, README, and the changelog describe
    the crew. The v1 package skill lists include `l7-crew`; the v0.1.1 payload
    is unchanged.
15. `go vet ./...`, `go test ./...`, the race check on the new packages,
    `make ci`, `make policy-check`, and `git diff --check` pass on the exact
    candidate.

## Risks and mitigations

- **Concurrency faults on one repository:** each task owns its directories,
  worktree, and branch; one supervisor lock per repository; one merge lock;
  target updates use Git's compare-and-swap `update-ref`; new packages run
  under the race detector.
- **Cost and quota growth:** `max_workers` and `repair_rounds` are bounded,
  each task keeps the three-failure pause, and quota resets are waited out per
  task.
- **Stale review after rebase:** verification re-runs at every new head and
  review re-runs whenever the rebased patch differs from the reviewed one.
- **Prompt injection steering a worker:** workers have no shell, changed paths
  are checked against scope and protected paths, review comes from a separate
  route, and every effect stays local.
- **Runaway background process:** the supervisor holds a lock, records its
  process ID in status, exits when no runnable task remains, and stops on
  `l7 crew cancel`, which preserves worktrees and evidence.
- **Failure output reaching a provider:** verification runs with the minimal
  environment, the tail is bounded, and it goes only to the session that
  already has repository access.
- **Policy-file compatibility:** new fields are omitted while unset, so a
  rollback to an older binary still reads the file.
- **Requirements drift:** `docs/artifacts/requirements.md` §13 assumes one
  worktree at a time; the crew is opt-in and local-only, and the currency note
  in `docs/foundation/README.md` records the remaining gap for later phases.

## Rollback

Revert the implementation commits, then this brief. The feature is default OFF,
the policy schema version is unchanged, and Headless is untouched. Private
crew state under `.git/l7/crew/` is uncommitted; the owner may inspect and
delete it deliberately. No remote, release, or deployment state is created.
