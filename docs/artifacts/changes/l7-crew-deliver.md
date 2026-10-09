# L7 Crew Deliver — Change Brief

| Field | Value |
|---|---|
| Change ID | `l7-crew-deliver` |
| Risk tier | `3` — external effects (push, pull request, merge) and a workflow-contract change |
| Status | `approved` by the active user; amendment `docs/foundation/requirements-amendment-crew-delivery.md` |
| Base commit | `0a76492e25ea78fdab27ee9f8129a0d49b262154` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `devin` |
| Assurance | `solo` — self-review only; no independent audit is claimed |
| Roadmap | Phase 3 of 4: pull-request delivery, stacked on Phase 2 `l7-crew-watch` (PR #22) |

## Problem

The crew can only fast-forward a local branch. Owners who work through pull
requests must push each crew branch, open each pull request, and copy the
evidence by hand, which loses the handoff and invites merging unchecked work.

## Scope

- A default-OFF `features.crew_pr` flag, plus `crew.remote` (default
  `origin`) and `crew.merge_method` (`merge`, `squash`, or `rebase`; default
  `squash`), omitted from the policy file while unset.
- `l7 crew plan --deliver pr` records delivery mode, remote, and base branch
  (the branch checked out at planning) in the plan digest. Approval of that
  plan authorizes exactly the effects in amendment item 2.
- Each ship task starts from the freshly fetched remote base. After local
  verification and independent review, Level 7 pushes its `l7/tasks/*` branch
  (fast-forward only), opens or reuses one pull request against the base with
  the task handoff as the description, and applies `l7-risk-tier-2`. The task
  becomes `pr-open`; it keeps its write scope, so overlapping tasks wait.
- While any task is `pr-open`, the supervisor polls the pull request. A merged
  pull request finishes the task; a closed one cancels it; changed check
  results update the task and wake the liaison. A failed check opens a
  `ci-failed` decision. `retry` sends the failing check names and a bounded
  log tail to the implementer, then verifies, re-reviews, and pushes a new
  commit.
- `l7 crew merge --task <id> --head <sha> --confirm` merges one pull request
  at that exact head, with the configured method. It refuses unless the
  request is open, not a draft, at that head, and clean to merge, and unless
  at least one check ran and all passed. It never uses `--admin` or forge
  auto-merge.
- CLI and MCP (`deliver` on `plan`; a `merge` action), status and board
  showing pull requests and checks, and updates to the `l7-crew` skill,
  workflow reference, README, and changelog.

Out of scope: auto-merge (Phase 4), forges other than GitHub through `gh`,
force pushes, branch deletion, and any change to non-crew paths.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-crew-deliver.md`
- `docs/foundation/requirements-amendment-crew-delivery.md`
- `internal/l7/adapter/forge/forge.go`
- `internal/l7/adapter/forge/forge_test.go`
- `internal/l7/adapter/forge/forgetest/forgetest.go`
- `internal/l7/adapter/headlessworker/crew_deliver.go`
- `internal/l7/adapter/headlessworker/crew_deliver_test.go`
- `cmd/l7/crew_deliver.go`
- `cmd/l7/crew_deliver_test.go`

Modify:

- `docs/foundation/README.md`
- `internal/l7/domain/crew.go`
- `internal/l7/domain/crew_test.go`
- `internal/l7/adapter/orchestrationconfig/config.go`
- `internal/l7/adapter/orchestrationconfig/config_test.go`
- `internal/l7/adapter/crew/plan.go`
- `internal/l7/adapter/crew/plan_test.go`
- `internal/l7/adapter/crew/engine.go`
- `internal/l7/adapter/crew/engine_test.go`
- `internal/l7/adapter/headlessworker/crew.go`
- `internal/l7/adapter/headlessworker/crew_test.go`
- `cmd/l7/crew_cli.go`
- `cmd/l7/crew_view.go`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `skills/l7-crew/SKILL.md`
- `references/WORKFLOW.md`
- `README.md`
- `CHANGELOG.md`

## Acceptance criteria

1. With `features.crew_pr` OFF, `plan --deliver pr` and `merge` fail closed,
   and local delivery is unchanged.
2. A PR plan's digest binds delivery mode, remote, and base branch; changing
   any of them changes the plan identity.
3. Pushes go only to `refs/heads/l7/tasks/*` on the recorded remote, never
   with force. A remote branch with commits Level 7 did not push stops the
   task with a decision.
4. Each ship task has at most one open pull request. Restarts reuse it, and a
   closed request is never silently reopened.
5. Pull-request descriptions carry the objective, acceptance criteria, exact
   verification argv and result, review verdict and models, changed paths,
   and what was not verified.
6. Tracking maps merged to done, closed to cancelled, and a head moved by
   someone else to a decision. Check failures open a `ci-failed` decision,
   and `retry` repairs, re-verifies, re-reviews, and pushes a new commit.
7. `merge` refuses a wrong or stale head, a draft or non-open request, any
   state other than clean, no checks, or any failing or pending check. It
   passes `--match-head-commit` and never `--admin` or `--auto`.
8. Tests cover the above against a local bare remote and a stand-in `gh`;
   `make ci` stages, `make policy-check`, and `git diff --check` pass.

## Risks and mitigations

- **Publishing unreviewed code:** delivery happens only after local
  verification and independent review, and merge needs an exact-head owner
  action plus passing checks.
- **Remote side effects beyond the approval:** only `l7/tasks/*` refs are
  pushed, fast-forward only, and only to the recorded remote and base.
- **Credentials:** Git and `gh` use the owner's authentication; Level 7
  passes no token and stores none.
- **Long-running supervisor:** it stays only while pull requests are open,
  polls at a bounded interval, and stops on `l7 crew cancel`, which leaves
  pull requests open for the owner.

## Rollback

Cancel or finish PR-delivery crews, close any crew pull requests you do not
want, then revert the implementation commits, this brief, and the amendment.
`features.crew_pr` stays OFF by default.
