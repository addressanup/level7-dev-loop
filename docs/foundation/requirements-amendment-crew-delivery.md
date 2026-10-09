# Requirements amendment: crew delivery

| Field | Value |
|---|---|
| Amends | `docs/artifacts/requirements.md` §11 (explicit non-goals) and §13 (constraints) |
| Decision | Active user, 2026-10-09: the crew delivers through pull requests, and later merges automatically when checks pass |
| Status | `approved` for the opt-in crew only; every other Level 7 path is unchanged |
| Applies from | `l7-crew-deliver` (PR delivery) onward; auto-merge needs its own brief |

## What changes

The requirements make "any Level 7-executed external ... mutation in v1.0" a
non-goal (§11) and assume one worktree at a time (§13). For the opt-in crew
only, they are amended as follows.

1. **Parallel worktrees.** The crew may run several tasks at once, each in its
   own disposable worktree, with one supervisor per repository and serialized
   integration. This supersedes the §13 single-worktree assumption for the
   crew.
2. **Pull-request delivery.** With `features.crew_pr` on and an owner-approved
   plan whose digest records the remote and the base branch, Level 7 may:
   - fetch the recorded base branch from the recorded remote;
   - push only the `l7/tasks/*` branches it created, fast-forward only, never
     with force and never to any other ref;
   - open one pull request per ship task against the recorded base, with the
     task handoff as its description;
   - apply exactly the `l7-risk-tier-2` label to those pull requests.

   Each push follows the same approved scope, protected-path refusal,
   verification, and independent review as local delivery.
3. **Owner-confirmed merge.** Level 7 merges a crew pull request only on an
   explicit owner action naming the exact head commit. It refuses unless the
   pull request is open, not a draft, at that head, and clean to merge, and
   unless at least one check ran and every check passed. It never bypasses
   branch protection (`--admin`) and never enables forge auto-merge.
4. **Auto-merge (decided direction, not yet built).** A later brief may merge
   automatically when checks pass, only under a separate per-project opt-in
   with all of these safeguards restated and tested: Tier 1 or 2 only; no
   protected path changed; at least one check ran and every required check
   passed at the exact head; independent review returned GO; the task came
   from the owner, not from issue or pull-request text; merges pin the head;
   and a circuit breaker pauses auto-merge when the base branch starts failing.

## What does not change

- No deployment, release, publication of packages, production mutation, or
  destructive remote operation (deleting, force-pushing, or rewriting remote
  history).
- No credential brokerage. Git and the `gh` CLI use the owner's own
  authentication; Level 7 stores no token.
- Every other Level 7 path keeps the §11 non-goal and stops before any
  external effect.

## Verification

The `l7-crew-deliver` brief turns items 2 and 3 into acceptance criteria and
tests them against a local bare remote and a stand-in `gh`. Live use against a
real forge needs the owner's go-ahead for the repository it targets.
