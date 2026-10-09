# L7 Crew Watch — Change Brief

| Field | Value |
|---|---|
| Change ID | `l7-crew-watch` |
| Risk tier | `3` — skill and workflow-contract change; owner takeover of autonomous worker sessions |
| Status | `approved` by the active user for this bounded repository-local implementation |
| Base commit | `e13ebf789b8feff117e817fbf58e48aad1ec631e` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `devin` |
| Assurance | `solo` — self-review only; no independent audit is claimed |
| Roadmap | Phase 2 of 4: visible crew (this brief), stacked on Phase 1 `l7-crew-core` (PR #21) |

## Problem

Phase 1 crew workers run out of sight. The owner sees state transitions
through `status` and `wait`, but cannot watch the crew live, and cannot step
into a stuck task without cancelling it and losing the worker's session.

## Scope

- `l7 crew watch` redraws a live board of tasks, states, routes, checks, and
  next actions whenever crew state changes. It reads local state only, makes
  no model calls, and exits when every task has finished.
- `l7 crew view` creates a detached tmux session when tmux is installed: the
  board in the first window and a shell in each unfinished task's worktree.
  It prints the attach command. Without tmux it says so and points to
  `watch`.
- `l7 crew attach --task <id>` takes one task out of automation. The
  supervisor stops a running worker by cancelling its provider process; a
  queued, paused, or decision-blocked task is held at once and its open
  decision is closed. The task becomes `attached`. It keeps its worktree,
  session, and write scope (overlapping ship tasks keep waiting) but frees its
  worker slot. The command returns the provider's native resume command for
  the task's session, to run inside the worktree: `codex resume <id>` or
  `claude --resume <id>`. Inside tmux it opens that command in a new window.
- `l7 crew release --task <id>` hands an attached task back. Level 7 treats
  the owner's edits as part of the candidate: it unstages index changes,
  checks changed paths and any owner commits against the task scope and
  protected paths, runs the exact verification, then commits, obtains
  independent review, and merges through the same serialized queue. A failing
  check goes back to the implementer session with its output.
- CLI and MCP `attach` and `release` actions; `watch` and `view` are terminal
  features. Updates to the `l7-crew` skill, workflow reference, README, and
  changelog.
- Stopping a Codex worker sends `turn/interrupt` and waits a bounded time for
  the turn to end. The first live attach showed that killing the local
  `codex app-server` client does not stop a codex-cli 0.162 turn: its
  managed daemon finished the turn 28 seconds later. Headless cancellation
  benefits too.

Level 7 persists no transcript or model output. Watching shows Level 7's own
state; full transcripts stay in the provider's own session store and are
reached through its native resume command.

Out of scope: streaming model output into Level 7, pull requests, auto-merge,
hooks, away mode, and remote hosts.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-crew-watch.md`
- `cmd/l7/crew_view.go`
- `cmd/l7/crew_view_test.go`
- `internal/l7/adapter/codexapp/session_test.go`

Modify:

- `internal/l7/adapter/codexapp/worker.go`
- `internal/l7/domain/crew.go`
- `internal/l7/domain/crew_test.go`
- `internal/l7/adapter/crew/store.go`
- `internal/l7/adapter/crew/store_test.go`
- `internal/l7/adapter/crew/engine.go`
- `internal/l7/adapter/crew/engine_test.go`
- `internal/l7/adapter/headlessworker/crew.go`
- `internal/l7/adapter/headlessworker/crew_test.go`
- `cmd/l7/crew_cli.go`
- `cmd/l7/orchestration_cli.go`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `skills/l7-crew/SKILL.md`
- `references/WORKFLOW.md`
- `README.md`
- `CHANGELOG.md`

## Acceptance criteria

1. `watch` redraws only when crew state changes, makes no model calls,
   exits successfully when every task is terminal, and leaves the crew
   running when interrupted.
2. `view` creates or reuses one tmux session per plan with the board and one
   window per unfinished task at its worktree. Without tmux it reports that
   and changes nothing.
3. Attaching a running task stops its worker within the supervisor's poll
   interval and marks it `attached` while other tasks continue. Attaching a
   queued, paused, or decision-blocked task takes effect at once and closes
   its open decision. Finished tasks cannot be attached. An attached ship task
   holds its scope but not a worker slot.
4. The returned resume command matches the task's provider kind and session.
   Gateway routes have no interactive resume, and the command says so.
5. `release` accepts only an attached task, queues it, and starts the
   supervisor when none is running. Owner edits are unstaged, re-checked
   against scope and protected paths (including owner commits), verified,
   committed, reviewed, and merged as in Phase 1. A failing check returns to
   the implementer with its output; an untouched task resumes where it was.
6. Level 7 writes no transcript or model output for watching or takeover.
7. CLI and MCP expose `attach` and `release`, and both fail closed while
   `features.crew` is OFF.
8. Domain, store, engine, git-backed executor, and CLI tests cover the above;
   `make ci` stages, `make policy-check`, and `git diff --check` pass.

## Risks and mitigations

- **Owner edits bypass worker limits:** every owner change is re-checked
  against scope, protected paths, verification, and independent review before
  it can merge.
- **Cancelling an in-flight worker:** Codex turns are interrupted through the
  protocol, and other providers' processes are stopped by the process
  adapter. Partial edits stay in the worktree, where the owner sees them and
  `release` re-checks them.
- **tmux missing or different:** tmux is optional; without it Level 7 prints
  the commands.
- **Resuming a non-interactive session:** Level 7 only prints the provider's
  documented resume command; the provider owns that session.

## Rollback

Release or cancel attached tasks first, because a Phase 1 binary rejects the
`attached` state. Then revert the implementation commits and this brief. The
feature stays behind the default-OFF `features.crew` flag.
