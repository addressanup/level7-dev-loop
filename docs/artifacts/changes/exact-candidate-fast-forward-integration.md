# Exact-Candidate Fast-Forward Integration — Hosted Policy Restart

| Field | Value |
|---|---|
| Change ID | `exact-candidate-fast-forward-integration` |
| Risk tier | `3` — protected GitHub controls and remote ref integration |
| Status | `hosted-policy-compatible remediation proposed`; implementation is not approved |
| Base commit | `1e381b5bf0bb024739cdc654d0d5aed5f128aed4` |
| Base tree | `5eff6de773e26cf3cb126853d158de83fddf3793` |
| Target integration PR | `addressanup/level7-dev-loop#3`; exact candidate `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, tree `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0` |
| Accountable owner | GitHub user `apbusinessidentity-tech`; fresh approval of the exact proposal commit is required before implementation |
| Implementer | GitHub user `addressanup`, operated through `codex-root` |
| Feature flag | Not applicable; no product runtime behavior changes |
| Latest rejected successor | `d4bbed9af10c7abb90db84403578bbc99d32cdf3`, tree `1295cd59d6629e860323177d5ed0eba7ee4b8461` |

## Problem

PR #3 is a verified and audited Tier 3 Git candidate whose commit identity must
be preserved. GitHub's permitted pull-request merge methods cannot advance
protected `main` to that exact commit: squash and rebase replace commits, while
a generated merge commit also changes the candidate identity. The bounded
operator therefore uses an exact compare-and-swap fast-forward and restores the
complete protection contract around that single effect.

The latest local verification successor `d4bbed9a...` closed the operator
findings from earlier audits, including chronological check selection, exact
lease authorization, post-confirmation refresh, restoration proof, ambient Git
containment, and automatic source-branch deletion rejection. A fresh independent
audit nevertheless returned `NO_GO` for two hosted-control defects:

1. its declared base `a178f047...` already contained this brief, so the trusted
   workflow's added-file-only derivation produced no approval or audit envelope;
2. the dedicated hosted base branch was unprotected and writable, so it was not
   a durable boundary for the controller built from the pull request base SHA.

Commit `1e381b5b...` preserves the rejected chain as immutable ancestors while
aggregate-reverting the active proposal, implementation, and verification
commits to the exact brief-absent tree of `01992e1f...`. This revision is the
only added path relative to that new reset commit. It repairs the planning and
hosted-publication contract; it grants no implementation or remote authority.

## Scope

After fresh accountable-owner approval, reapply the already bounded operator
hardening only to the repository-owned operator and its offline fake-command
contract test. Remove the stale verification record in the implementation
commit, run repository-pinned verification, and then add one fresh verification
record. A separately authorized independent reviewer may add the sole audit
record only after an exact-successor `GO`.

Any later hosted publication is a separate transition. Before a pull request
uses a dedicated base, that ref must point exactly to this declared base commit
and tree and must be protected with administrator enforcement, pull-request
review, linear history, force-push denial, deletion denial, and no bypass or
direct-push ambiguity. The trusted controller, workflow, bootstrap recipe, and
toolchain inputs at that ref must equal the reviewed trusted baseline. A ref,
protection, ruleset, permission, or controller mismatch fails closed.

The hosted proposal must demonstrate that the exact base-to-proposal diff adds
this brief, that the workflow's `--diff-filter=A` query selects exactly this
path, and that Git identifies the proposal commit as this path's latest addition
commit. The base must remain at the declared SHA throughout fresh exact-head
checks and reviews. Existing checks, approvals, audit decisions, or envelopes
from any predecessor do not transfer.

No workflow, controller, production package, provider adapter, skill, plugin,
local `main`, historical branch, or unrelated user file may change. No remote
ref, pull request, protection rule, ruleset, review, check, release, deployment,
or provider process is touched by this planning proposal.

## Exact implementation file set

After fresh exact-proposal owner approval, the complete permitted file set is:

- `scripts/harness/exact-fast-forward-integration.sh`
- `scripts/harness/check-exact-fast-forward-integration.sh`
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-verification.md`
- `docs/artifacts/changes/exact-candidate-fast-forward-integration-audit.md`

The brief itself is the sole proposal artifact and is added before
implementation. The verification path is modified or removed only by the
implementation and fresh verification steps; the audit path is added only after
an independent `GO`. Every other path is outside scope, including the user-owned
untracked audit under `docs/artifacts/`.

## Acceptance criteria

1. The complete historical chain through `d4bbed9a...` remains immutable.
   Reset commit `1e381b5b...` is its descendant, has tree `5eff6de7...`, and is
   byte-identical to the prior brief-absent planning tree `01992e1f...`.
2. The proposal commit is a direct descendant of `1e381b5b...`; its net diff
   adds only this brief. Both the trusted workflow's exact added-file query and
   Git addition history resolve this path and the exact proposal commit.
3. Implementation cannot begin until `apbusinessidentity-tech` explicitly
   approves the exact proposal commit outside candidate-controlled text. The
   implementer remains `addressanup`; self-approval and every predecessor
   approval or local envelope fail closed.
4. Before any hosted candidate evaluation, a dedicated base ref points exactly
   to `1e381b5b...` and is protected against direct, force, deletion, and bypass
   updates with administrator enforcement. Its exact ref, protection, rulesets,
   permissions, trusted workflow/controller blobs, and bootstrap inputs are
   captured and revalidated after PR targeting and for every decisive run.
5. A hosted trusted-policy run is admissible only when its event base SHA is
   `1e381b5b...`, its derived brief path and addition commit equal this proposal,
   and its owner and auditor envelopes come from fresh exact-head GitHub reviews.
   Any base movement, stale run, missing envelope, or derivation ambiguity blocks.
6. The operator accepts only full explicit repository, PR, base, head, tree,
   owner, and auditor bindings; rejects malformed, conflicting, abbreviated, or
   extra input; targets only `refs/heads/main`; and preserves all argument,
   permission, schema, cancellation, cleanup, and containment controls.
7. Preflight and the post-confirmation refresh prove the sole administrator,
   repository settings, absence of conflicting rulesets, exact open/unmerged PR
   state, distinct identities, exact-head approvals, required checks, protection,
   source branch, main ref, ancestry, ahead/behind counts, and candidate tree.
8. Required checks are selected from the trusted app by unique greatest
   non-null `started_at`. Missing or tied chronology, wrong app, stale evidence,
   cancellation, skip-only evidence, or a non-successful latest run fails closed.
9. The active terminal displays the full effect and accepts only the complete
   candidate SHA. Redirected input, a prefix, stale confirmation, or any changed
   mutable fact performs no remote mutation.
10. The sole permitted update is a fast-forward of `main` with exactly
    `--force-with-lease=refs/heads/main:<full-expected-base>` after independent
    ancestry proof. Generic force, an unbound lease, a leading-plus refspec,
    merge/close APIs, source-branch deletion, a second ref update, and automatic
    `delete_branch_on_merge` behavior are rejected before and after confirmation.
11. Restoration remains armed until a fresh read proves the complete canonical
    protection contract. Failure or mismatch stops with re-enabling administrator
    enforcement as the sole recovery action. Postflight proves main/head/tree,
    source-branch preservation, restored protection, and indirect PR merge state.
12. Git executes under a minimal explicit environment that excludes ambient
    configuration and transport overrides. The offline fake-command contract
    covers every positive sequence and negative/race/restoration condition above.
13. Repository-pinned `make verify`, the focused contract test, shell syntax,
    shell lint, diff hygiene, exact scope, artifact budget, and clean tracked and
    index state pass before the fresh verification record is committed.
14. A separately authorized `l7-release` reviewer independently audits the exact
    verified successor. Passing tests, hosted CI, or audit `GO` cannot execute
    the operator; live use requires a later explicit transition.
15. PR #3, PR #1, their source branches, local `main`, the invalid implementation
    worktree, and all historical commits remain untouched. No provider executable,
    model session, actual-host gate, release, deployment, install, signing,
    publication, or Wave 5 work occurs or is claimed.

## Risks and mitigations

- **Missing trusted envelopes:** make the brief a true added path from its exact
  hosted base and prove both workflow derivation commands before implementation.
- **Mutable trusted base:** protect the exact dedicated ref before use, enforce
  administrators, deny force/deletion/bypass routes, and reject any later change
  in ref, protection, rulesets, permissions, controller, or event base identity.
- **Stale authority:** bind owner approval, independent audit, checks, and active
  terminal confirmation to their exact commits and require fresh decisions after
  every proposal, implementation, verification, or topology change.
- **Protection gap or ref race:** validate the full contract twice, use one
  expected-old lease-bound fast-forward, keep recovery armed, and prove the exact
  restored postcondition before success.
- **Generic bypass or ambient redirection:** fix the target and command grammar,
  prohibit alternate mutation APIs, and run Git with an explicit minimal
  environment.
- **Historical or user-state loss:** use ordinary additive commits only, isolate
  the worktree, stage exact paths, and never inspect or modify unrelated untracked
  content.

## Rollback

Before implementation, ordinary-revert the proposal commit to return exactly to
reset tree `5eff6de7...`. After implementation, revert audit, verification,
implementation, and proposal commits in reverse order. Never reset, rebase,
force-push, delete, or rewrite a historical ref.

If a separately authorized live operation ever starts, restoring administrator
enforcement is the immediate recovery action. If `main` did not move, no ref
rollback is needed. If it reached the exact audited PR #3 candidate, never reset
or force-push it; any product rollback is a separately approved reverse-order
ordinary-revert change.

## Commit sequence and approval boundary

1. `1e381b5b...` — preserve all predecessors and aggregate-revert the active
   chain to the exact prior brief-absent tree.
2. Add only this corrected brief, then stop for fresh exact-commit approval from
   `apbusinessidentity-tech`.
3. After that approval only, reapply the bounded operator hardening to the two
   scripts and remove the stale verification record.
4. Verify locally and add the sole fresh verification record.
5. Obtain separate authorization for a fresh independent read-only
   `l7-release` audit; add an audit record only after `GO`.
6. Stop. Protected-base creation or validation, proposal publication, PR
   targeting, hosted checks/reviews, and live integration each require explicit
   later authority.

This brief authorizes no implementation or external effect by its presence.
