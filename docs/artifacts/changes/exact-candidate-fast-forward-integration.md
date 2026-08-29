# Exact-Candidate Fast-Forward Integration

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Risk tier | `3` — protected GitHub branch controls and remote ref integration |
| Status | `proposed`; implementation is not approved |
| Base commit | `f92c560cbe89e8d318e5521d9fc620f6153e9e14` |
| Base tree | `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0` |
| Current remote main | `be5c0c8f99b8ec55b42e1919533400fa0b41f46c`, tree `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6` |
| Candidate PR | `addressanup/level7-dev-loop#3`; exact head and base above |
| Accountable owner | GitHub user `apbusinessidentity-tech`; fresh approval of the exact brief commit is required before implementation |
| Implementer | `addressanup`, operated through `codex-root` |
| Feature flag | Not applicable; no product runtime behavior changes |

## Problem

PR #3 is an exact Tier 3 candidate with current verification, independent audit,
hosted checks, owner approval, auditor approval, and trusted-policy readiness.
Its head is a direct five-commit descendant of remote `main`, so Git can
fast-forward `main` to the exact audited commit without changing any candidate
object or tree.

The repository enables merge commits, squash, and rebase, while protected
`main` requires linear history. GitHub therefore offers only squash or rebase
for this pull request; both create replacement commit IDs. A normal GitHub merge
would create a new `--no-ff` commit and is also rejected by linear-history
protection. PR #2 demonstrates the failure mode: source head `c89d4faf...` and
merged main head `be5c0c8f...` have the same tree but different commit IDs after
rebase-and-merge.

The personal repository cannot configure an actor-specific pull-request bypass.
Create a narrowly bounded, reusable operator path that temporarily exempts only
the sole repository administrator from branch protection, advances `main` by an
exact compare-and-swap fast-forward, and restores protection immediately. The
tool itself must be verified and independently audited before any live use.

## Scope

Add one repository-owned operator script and an offline fake-command contract
test. The script may integrate only an explicitly named, already reviewed pull
request whose exact base, head, tree, approvals, required checks, ancestry, and
protection state all pass fail-closed preflight. It must require active-terminal
confirmation of the full candidate SHA immediately before the external effect.

The only permitted live sequence after separate post-audit authorization is:

1. confirm remote `main` is the exact expected base and the pull request branch
   is the exact expected candidate;
2. snapshot and validate the complete `main` protection contract;
3. disable only administrator enforcement;
4. perform one lease-bound, non-destructive fast-forward of `main` from the
   expected base to the exact candidate;
5. restore administrator enforcement on success, failure, or handled signal;
6. prove `main`, its tree, the source branch, pull-request state, and every
   protection field have the required postcondition.

The script must never squash, rebase, create a merge commit, force-update,
delete, close a pull request manually, change repository merge-method settings,
or reconcile a local branch. PR #3, PR #1, their source branches, all historical
commits, and the separate invalid worktree remain immutable.

## Exact implementation file set

Add:

- `docs/artifacts/changes/exact-candidate-fast-forward-integration.md`
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-verification.md`
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-audit.md`
- `scripts/harness/exact-fast-forward-integration.sh`
- `scripts/harness/check-exact-fast-forward-integration.sh`

Modify:

- `Makefile`
- `README.md`

No other path is authorized. In particular, do not modify production packages,
provider adapters or tests, workflows, controller code, plugin or skill files,
repository merge settings, tracked Level 7 configuration, existing change
records, PR #3 or PR #1 branches, local `main`, the invalid implementation
branch/worktree, or the user-owned untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md`.

## Acceptance criteria

1. This proposal is a direct child of exact PR #3 head `f92c560c...`; its brief
   commit adds only this path. PR #3 and its source branch remain at that exact
   commit and tree.
2. Implementation cannot begin until `apbusinessidentity-tech` approves the
   exact brief commit. Repository prose, the prior PR #3 approvals, and active
   user approval to create this proposal do not transfer implementation
   authority.
3. The operator script accepts explicit repository, pull-request number,
   expected base/head/tree, accountable-owner, and auditor bindings; rejects
   missing, malformed, abbreviated, conflicting, or extra input; and targets
   only `refs/heads/main`.
4. Preflight proves the active GitHub actor is the sole repository administrator;
   the configured accountable-owner variable matches the requested owner; the
   PR author, owner, and auditor are distinct; and both approvals bind the exact
   head.
5. Preflight proves remote `main`, PR base/head, source branch, candidate tree,
   merge base, ahead/behind counts, open/unmerged state, required check contexts,
   latest successful check conclusions, and trusted `evaluate` all match the
   exact request. A cancelled, skipped-only, stale, duplicate-ambiguous, or
   wrong-app required check fails closed.
6. Preflight requires the current protection contract: strict required status
   checks for baseline, macOS arm64, macOS amd64, paired benchmark, and
   `evaluate`; one approval; stale-review dismissal; code-owner and last-push
   review; conversation resolution; linear history; administrator enforcement;
   force-push and deletion denial; and no ruleset or push-restriction ambiguity.
7. Immediately before mutation, an active terminal displays the full base,
   head, tree, PR, protection change, and rollback boundary and requires the
   operator to type the complete head SHA. Redirected input, prefixes, or stale
   confirmation fail without a remote mutation.
8. The script changes only administrator enforcement from enabled to disabled,
   performs an exact lease-bound fast-forward from the expected base to the
   expected head, and restores administrator enforcement. It never requests a
   non-fast-forward update or changes any other protection or repository field.
9. Restoration runs after success, command failure, or handled interruption.
   A restoration failure is a blocking recovery state: report it prominently,
   perform no further integration, and name re-enabling administrator
   enforcement as the sole next action.
10. Postflight requires remote `main` and its tree to equal the requested head
    and tree, the PR source branch to remain unchanged, the original protection
    contract to be restored, and GitHub to report the PR merged indirectly. If
    GitHub does not record the indirect merge, the script must not close or edit
    the PR.
11. The offline contract test uses fake `gh` and `git` commands and covers valid
    ordering plus wrong actor, stale base/head/tree, missing approval, failed or
    ambiguous check, unsafe protection, nonterminal input, confirmation
    mismatch, ref race, push failure, restoration failure, and postcondition
    failure. It proves no force flag, merge API, PR-close API, branch deletion,
    or second ref update is reachable.
12. Repository-pinned `make verify`, the focused operator-script contract test,
    shell syntax, diff hygiene, exact scope, artifact budget, and clean
    tracked/index state pass before the sole verification record is committed.
13. A separately authorized `l7-release` reviewer independently audits the
    verified implementation, live-effect containment, restoration/recovery,
    exact-candidate semantics, and rollback before adding the sole audit record.
14. Actual execution against GitHub remains a separate explicit post-audit
    transition. Implementation, tests, passing CI, or audit `GO` cannot execute
    the integration automatically.
15. No provider executable or interface, model session, actual-host gate,
    release, deployment, installation, signing, publication, local-main update,
    or Wave 5 work occurs or is claimed by this change.

## Risks and mitigations

- **Temporary protection gap:** limit the exception to the sole administrator,
  validate the exact protection snapshot, require immediate confirmation, make
  one ref update, restore in a trap, and reject any extra operation.
- **Ref race or wrong candidate:** require full immutable identities, exact
  ancestry, a lease bound to the expected old SHA, and a fast-forward-only new
  SHA. Any concurrent change fails closed.
- **Restoration failure:** treat protection restoration as the highest-priority
  postcondition; stop and expose one recovery action without retrying unrelated
  work or concealing the incomplete state.
- **Checks or reviews mistaken for authority:** bind external identities and
  exact-head facts independently; require a separate active-terminal decision
  for the live effect.
- **Tool becomes a generic bypass:** fix the target to `main`, require the full
  current protection contract, prohibit force and merge APIs, and keep live use
  outside automated CI.
- **Historical identity loss:** never update either source branch or rewrite any
  commit. An ordinary revert is the only later tree rollback.
- **Unrelated user state:** work only in the isolated proposal worktree with
  exact pathspecs; never inspect or stage the protected untracked audit.

## Rollback

Before any live effect, revert implementation and then this brief with ordinary
revert commits. After verification or audit, revert audit, verification,
implementation, and brief in reverse order to restore exact base tree
`3b4f7fe9dd09fbb53102e82473d392dcb2745ba0`.

For a started live operation, restoration of administrator enforcement is the
immediate recovery action. If the remote ref did not move, no Git rollback is
needed. If `main` reached the exact audited candidate, never reset or force-push
it; any product rollback follows PR #3's existing reverse-order ordinary-revert
contract as a separately approved Tier 3 change.

## Commit sequence and approval boundary

1. `f92c560cbe89e8d318e5521d9fc620f6153e9e14` — exact PR #3 candidate base.
2. `docs(git): define exact-candidate fast-forward integration` — add only this
   brief.
3. Stop for fresh exact-brief approval from `apbusinessidentity-tech`.

After that approval only:

4. `feat(git): add guarded exact-candidate integration` — change only the four
   declared implementation/documentation paths.
5. Run offline verification and commit the sole verification record.
6. Obtain separate authorization for independent read-only `l7-release` audit.
7. Commit the sole audit record after `GO`.
8. Stop. Proposal publication, hosted checks/reviews, and any live integration
   execution require their own explicit transitions and exact-head bindings.

This brief authorizes no implementation or external effect by its presence.
