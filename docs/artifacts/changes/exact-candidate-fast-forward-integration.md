# Exact-Candidate Fast-Forward Integration

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Risk tier | `3` — protected GitHub branch controls and remote ref integration |
| Status | `controller-compatible remediation proposed after independent NO_GO`; remediation implementation is not approved |
| Base commit | `a178f047d8d0269ae2b1b0aa957ff3b65ff75116` |
| Base tree | `6dcade32b7b4d765dea7925fe0f2e7326088c216` |
| Current remote main | `be5c0c8f99b8ec55b42e1919533400fa0b41f46c`, tree `e6edcf5bbd01b11769ec4c1b3a848d47a24c69b6` |
| Target integration PR | `addressanup/level7-dev-loop#3`; exact head `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, tree `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0` |
| Accountable owner | GitHub user `apbusinessidentity-tech`; fresh approval of the exact brief commit is required before implementation |
| Implementer | `addressanup`, operated through `codex-root` |
| Feature flag | Not applicable; no product runtime behavior changes |
| Rejected candidate | `bd0cb9b0f7a39ba512e3a13d5d5a3da91ee25ff7`, tree `80643e1bfed3efb35beada46a750d91fc61775bf` |
| Preserved failed remediation | `eda83482d589915aa9a9ece8f1ac217535a33222`, tree `1cd45647f9cea91c6d232fd13d6c5a917c429060` |
| Planning reset | Revert `a178f047d8d0269ae2b1b0aa957ff3b65ff75116`; brief deletion `01992e1f7745ab9eb423e76d5969040b7373be56` |

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

A separately authorized independent read-only audit returned `NO_GO` for the
verified successor `bd0cb9b0...`. It found that required checks were ordered by
check-run ID instead of chronology, the approved no-force wording did not
authorize the exact lease option used by the implementation, mutable authority
was not fully refreshed after terminal confirmation, restoration state was
cleared before exact restoration was proved, and ambient
`GIT_CONFIG_PARAMETERS` remained available. No audit record was materialized.
The first remediation implementation was preserved at `eda83482...`, but the
controller correctly rejected approval of a modified brief because Tier 3
authority binds the brief's latest addition commit. Ordinary revert
`a178f047...` restored the exact approved-proposal tree, and `01992e1f...`
deleted only this brief so the same sole path could be re-added as a fresh
planning boundary. This revision proposes only the bounded remediation
contract; it grants no implementation or live-effect authority.

## Scope

Harden the existing repository-owned operator script and offline fake-command
contract test. The script may integrate only an explicitly named, already
reviewed pull request whose exact base, head, tree, approvals, required checks,
ancestry, and protection state all pass fail-closed preflight. It must require
active-terminal confirmation of the full candidate SHA immediately before the
external effect.

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

Modify after fresh owner approval:

- `scripts/harness/exact-fast-forward-integration.sh`
- `scripts/harness/check-exact-fast-forward-integration.sh`

Governance paths:

- `docs/artifacts/changes/exact-candidate-fast-forward-integration.md` — this
  sole re-added brief;
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-verification.md`
  — remove the stale record in the remediation implementation, then add one
  fresh record after verification; and
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-audit.md` —
  add only after an independent `GO`.

No other path is authorized. In particular, do not modify production packages,
provider adapters or tests, workflows, controller code, plugin or skill files,
repository merge settings, tracked Level 7 configuration, existing change
records, PR #3 or PR #1 branches, local `main`, the invalid implementation
branch/worktree, or the user-owned untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md`.

For the NO_GO remediation, the net base-to-proposal change may modify only this
same brief. After fresh exact-brief owner approval, the remediation
implementation may modify only the two harness scripts and remove the stale
verification record; the fresh verification successor may add that same
verification path back as the sole current record. `Makefile`, `README.md`,
production code, workflows, and every other tracked path must remain unchanged.

## Acceptance criteria

1. Original proposal `61a005cc...`, rejected candidate `bd0cb9b0...`, amended
   brief `159a5de1...`, failed remediation `eda83482...`, exact-tree revert
   `a178f047...`, and brief deletion `01992e1f...` remain immutable. Base
   `a178f047...` has tree `6dcade32...`, identical to `159a5de1...`; the net
   base-to-proposal diff modifies only this same brief. PR #3 and its source
   branch remain at their exact commit and tree.
2. Remediation implementation cannot begin until `apbusinessidentity-tech`
   approves the exact revised-brief commit. Repository prose, PR #3 approvals,
   approval of original brief `61a005cc...`, the independent NO_GO decision,
   and active user approval to create this remediation proposal do not transfer
   implementation authority.
3. The operator script accepts explicit repository, pull-request number,
   expected base/head/tree, accountable-owner, and auditor bindings; rejects
   missing, malformed, abbreviated, conflicting, or extra input; and targets
   only `refs/heads/main`.
4. Preflight proves the active GitHub actor is the sole repository administrator;
   the configured accountable-owner variable matches the requested owner; the
   PR author, owner, and auditor are distinct; and both approvals bind the exact
   head. After terminal confirmation and before protection mutation, it repeats
   the sole-administrator query and revalidates repository settings, absence of
   rulesets, exact open/unmerged PR state and identities, owner binding, both
   approvals, protection, source and main refs, and every required check.
5. Preflight proves remote `main`, PR base/head, source branch, candidate tree,
   merge base, ahead/behind counts, open/unmerged state, required check contexts,
   latest successful check conclusions, and trusted `evaluate` all match the
   exact request. For each context it selects the unique trusted-app run with
   the greatest non-null `started_at`, never check-run ID; equal greatest
   timestamps, missing timestamps, cancelled, skipped-only, stale,
   duplicate-ambiguous, or wrong-app evidence fails closed.
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
   expected head, and restores administrator enforcement. The sole permitted
   force-prefixed option is exactly
   `--force-with-lease=refs/heads/main:<full-expected-base>` as the explicit
   compare-and-swap guard after independent ancestry proof. Generic `--force`,
   `-f`, an unbound or empty lease, a leading-plus refspec, and every
   non-fast-forward update remain prohibited. It changes no other protection or
   repository field.
9. Restoration runs after success, command failure, or handled interruption.
   The recovery state remains armed after the restoration POST until a fresh
   GET proves the complete canonical protection contract equals the original.
   A failed POST, failed proof request, or mismatched protection is a blocking
   recovery state: report it prominently, perform no further integration, and
   name re-enabling administrator enforcement as the sole next action.
10. Postflight requires remote `main` and its tree to equal the requested head
    and tree, the PR source branch to remain unchanged, the original protection
    contract to be restored, and GitHub to report the PR merged indirectly. If
    GitHub does not record the indirect merge, the script must not close or edit
    the PR.
11. The offline contract test uses fake `gh` and `git` commands and covers valid
    ordering plus wrong actor, stale base/head/tree, missing approval, failed or
    ambiguous check, unsafe protection, nonterminal input, confirmation
    mismatch, ref race, push failure, restoration failure, and postcondition
    failure. It additionally covers competing same-app check runs whose ID and
    start-time order disagree, post-confirmation sole-admin/ruleset/PR changes,
    successful restoration POST with unchanged protection, and ambient Git
    configuration injection. It proves the one exact bound lease is required
    while generic or unbound force forms, merge API, PR-close API, branch
    deletion, and a second ref update are unreachable.
12. Repository-pinned `make verify`, the focused operator-script contract test,
    shell syntax, diff hygiene, exact scope, artifact budget, and clean
    tracked/index state pass before a fresh sole verification record is
    committed. The `bd0cb9b0...` verification remains historical but is stale
    for every remediation successor and must be absent from the remediation
    implementation tree before fresh verification.
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
  current protection contract, permit only the fully bound expected-old lease,
  prohibit every generic or non-fast-forward force form and merge API, and keep
  live use outside automated CI.
- **Stale authority after confirmation:** repeat every mutable authority,
  repository, ruleset, PR, review, check, protection, and ref query immediately
  before disabling enforcement.
- **Misordered check evidence:** choose the unique greatest trusted-app
  `started_at` per required context and fail on missing or tied chronology.
- **False restoration success:** keep recovery armed until a fresh canonical
  protection read proves exact restoration; otherwise emit only the blocking
  recovery action.
- **Ambient Git redirection:** execute Git with a minimal explicit environment
  that removes `GIT_CONFIG_PARAMETERS` and every other ambient configuration or
  transport override.
- **Historical identity loss:** never update either source branch or rewrite any
  commit. An ordinary revert is the only later tree rollback.
- **Unrelated user state:** work only in the isolated proposal worktree with
  exact pathspecs; never inspect or stage the protected untracked audit.

## Rollback

Before remediation, revert the re-added proposal and then deletion commit
`01992e1f...` to restore exact restart base tree `6dcade32...` without rewriting
history. Before any live effect, revert remediation and its evidence in reverse
order to return to that same tree. The older rejected candidate, failed
remediation, and its ordinary revert remain immutable ancestors. Any later
rollback of PR #3 itself follows its separate reverse-order ordinary-revert
contract; remote `main` is never reset or force-pushed.

For a started live operation, restoration of administrator enforcement is the
immediate recovery action. If the remote ref did not move, no Git rollback is
needed. If `main` reached the exact audited candidate, never reset or force-push
it; any product rollback follows PR #3's existing reverse-order ordinary-revert
contract as a separately approved Tier 3 change.

## Commit sequence and approval boundary

1. Preserve the complete existing chain from exact PR #3 head `f92c560c...`
   through rejected verification `bd0cb9b0...`, amended brief `159a5de1...`,
   and failed remediation `eda83482...`.
2. `a178f047...` — ordinary-revert the failed remediation and restore exact
   tree `6dcade32...`.
3. `01992e1f...` — delete only the brief to reset its controller addition
   identity; preserve all prior brief versions in Git history.
4. `docs(git): reestablish exact integration remediation boundary` — re-add
   only this same sole brief with `a178f047...` as its declared base. The net
   base-to-proposal diff changes only this brief.
5. Stop for fresh approval of that exact proposal commit from
   `apbusinessidentity-tech`. Every earlier approval does not transfer.

After that fresh approval only:

6. `fix(git): harden exact integration controls` — modify only the operator and
   contract test and remove the stale verification record.
7. Run offline verification and add one fresh verification record.
8. Obtain separate authorization for a fresh independent read-only
   `l7-release` audit that does not reuse the NO_GO decision as certification.
9. Commit the sole audit record only after `GO`.
10. Stop. Proposal publication, hosted checks/reviews, and any live integration
   execution require their own explicit transitions and exact-head bindings.

This brief authorizes no implementation or external effect by its presence.
