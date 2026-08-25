# Level 7 Dev Loop — Git-Baseline Host-Binding Successor Amendment

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-ORC-003` |
| Artifact type | Inert Git-baseline successor to the consumed local Codex host-binding attempt |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT pending exact-digest model audit, structurally independent qualified review, fresh current-session AP1, selection, and dispatch** |
| Parent plan | [`L7-ORC-001` 0.3.1](orchestration-plan.md), SHA-256 `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| Parent audit | [`L7-AUD-ORC-001`](orchestration-plan-audit.md), SHA-256 `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` |
| Parent approval | [`L7-APR-ORC-001`](orchestration-plan-approval.md), SHA-256 `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` |
| Immediate predecessor | [`L7-AMD-ORC-002` 0.1.0](orchestration-plan-host-binding-amendment-successor.md), SHA-256 `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74` |
| Predecessor audit | [`L7-AUD-ORC-AMD-003`](orchestration-plan-host-binding-amendment-successor-audit.md), SHA-256 `8c9a495a7160c592da4aeb4964d93f21f29cc85d24653b9059ec8a0e22337c06` |
| Predecessor qualified review | [`L7-REV-ORC-AMD-002`](orchestration-plan-host-binding-amendment-successor-qualified-review.md), SHA-256 `4a0c8dab4c5e97bfde247df5dd2f065852ed14e63ace6705bc2b76aadc0374b8` |
| Predecessor nonce/state | `L7-AMD-ORC-002-20260825-01` — **CONSUMED; non-replayable** |
| Earlier nonce/state | `L7-AMD-ORC-001-20260825-02` — **CONSUMED; non-replayable** |
| Git setup authority | Current owner direction to set up Git first, following the proposed local `main` / baseline-commit / no-remote / no-extra-worktree strategy |
| Draft authority | Current owner direction to use `level7-dev-loop:l7-greenfield` for the previously described narrow successor amendment |
| Draft transport | `level7-dev-loop:l7-greenfield`; staged/cached SHA-256 `6c76a16af74b932733f3a1ea0838fef67fe2c5cbaf6a6aab22777949c8866609` |
| Logical action if later activated | `L7-FOUNDATION-START-WAVE-1` |
| Drafting effect / activated effect ceiling | A1 successor-governance record / one local A1 Wave 1 contract-and-specification proposal |
| Governing risk | `R3` authorization identity, source transition, and one-use binding |
| Fresh nonce | `L7-AMD-ORC-003-20260825-01` |
| Proposed validity ceiling | Earliest invalidation below, including 2026-09-01 23:59:59 Asia/Kathmandu |
| Candidate path | `docs/artifacts/orchestration-plan-host-binding-amendment-git-successor.md` |
| Next gate | Separate read-only model audit of the final candidate digest and containing Git commit |

## 1. Purpose and predecessor terminal record

The `L7-AMD-ORC-002` review chain closed and its exact fresh-thread AP1, post-AP1 `l7-build` selection, and token-prefixed dispatch made nonce `L7-AMD-ORC-002-20260825-01` eligible and then moved it from `UNUSED` to `IN_USE`.

During pre-mutation validation on 2026-08-25 at 10:59:09 Asia/Kathmandu, the coordinator recomputed the successor, audit, review, parent, manifest, marketplace, package, and skill hashes and verified all three protected hash manifests. A zsh loop then assigned to the shell-special lowercase variable `path`, which is tied to command lookup. The next unqualified `stat` invocation failed with exit code 127 and `zsh:40: command not found: stat`.

The preparation error occurred before `make verify`, the complete pre-snapshot, candidate-body construction, or any `apply_patch` call. A terminal read-only check at 10:59:33 Asia/Kathmandu confirmed both Wave 1 outputs absent. The correct result was `BLOCKED / CONSUMED`, not `RECOVERY_REQUIRED`.

The exact local session evidence is `/Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T10-56-30-01a03754-ab28-7f32-9f97-31de915de029.jsonl`. It is same-user mutable corroboration, not a signed ledger or trusted counter.

This successor preserves both earlier attempts as consumed. It does not retry, repair, resume, or reinterpret either nonce. It creates a distinct inert candidate with a new nonce, a new source identity, and a wholly fresh audit/review/AP1/selection chain.

## 2. Separately authorized Git baseline

Before this candidate was drafted, the owner separately directed local Git setup. The coordinator:

1. verified all protected input manifests;
2. ran pinned `make verify` successfully using Go 1.26.7;
3. initialized a local SHA-1 Git repository on branch `main`;
4. staged exactly the 62 existing non-cache Foundation files while `.cache/` remained ignored;
5. created root commit `08c38b69a2cd63b4adf27873756a09e363e0c5a4` with subject `chore(repo): establish foundation baseline`; and
6. confirmed a clean worktree, one worktree only, and zero remotes.

The root commit has tree `bcb254506102cc52386e14dc65414face95f4a6b`. It was authored and committed as `addressanup <address.anup@gmail.com>` at 2026-08-25T11:23:46+05:45.

This Git operation was separately owner-authorized and is already complete. This artifact records rather than grants that authority. It does not authorize a remote, additional worktree, branch protection, publication, push, rewrite, tag, merge, or release.

## 3. Narrow normative successor

This section has no activated effect until every gate in §6 closes.

Once activated, and only for the Codex CLI route, this successor would permit one new attempt at logical action `L7-FOUNDATION-START-WAVE-1`. The attempt may create only:

1. `docs/artifacts/wave-01-change-contract.md`; and
2. `docs/artifacts/wave-01-specification.md`.

Both must be new regular single-link files created without overwrite. They remain uncommitted proposal files at the mandatory owner-approval stop; the future invocation does not authorize `git add`, commit, branch creation, merge, remote creation, push, or any other Git mutation.

This successor supersedes only the availability and source-binding portions needed for a future attempt. It does not alter the immutable parent, predecessor, audit, review, or terminal records. All other `L7-ORC-001` boundaries remain in force, including:

- specification before design and implementation;
- one primary writer;
- no product, harness, prompt, skill, manifest, dependency, generated-package, provider, network, root, publication, deployment, exposure, release, cleanup, or self-modification effect;
- no automatic continuation; and
- no product, security, compatibility, compliance, release, or deployment assurance claim.

Git existence resolves only the mechanical absence recorded by pre-wave gate `PW-02`. Product implementation remains blocked until the owner approves the Wave 1 specification and design and every other applicable pre-wave decision closes.

## 4. Exact inherited and Git-aware binding

### 4.1 Project and Git source identity

Every future audit and invocation preflight MUST recompute these fields. Missing, dirty, divergent, ambiguous, partially matching, or changed state is `BLOCKED`.

| Binding | Required value |
|---|---|
| Canonical project root | `/Users/anuppandey/Desktop/level7-dev-loop`; every component resolves directly without a symlink |
| Canonical output parent | `/Users/anuppandey/Desktop/level7-dev-loop/docs/artifacts`; direct directories with no symlink component |
| Repository identity | Git repository rooted exactly at the canonical project root; object format `sha1` |
| Foundation baseline | Root commit `08c38b69a2cd63b4adf27873756a09e363e0c5a4`; tree `bcb254506102cc52386e14dc65414face95f4a6b`; exactly 62 tracked files |
| Branch strategy | Local `main`; one primary writer; exactly one worktree at the canonical root; zero remotes |
| Candidate lineage | This candidate must be committed as a direct single-file successor to the Foundation baseline; its final digest and containing commit are supplied post-write |
| Review lineage | Each later audit and qualified-review commit must be a linear descendant, must add only its separately authorized record, and must bind its exact parent |
| Invocation tip | Exact qualified-review-chain tip named in the future AP1; `main` points to it; no detached HEAD, merge, rewrite, replacement, shallow ambiguity, or unexpected object-format change |
| Clean admission state | No tracked worktree or index delta and no untracked non-ignored path before the future attempt |
| Ignored state | `.cache/` remains ignored and is excluded from source identity and complete workspace snapshots |
| Permitted output state | Both Wave 1 outputs absent in all forms before mutation; afterward they are the only two untracked non-ignored paths |
| Product surfaces | No reserved product path may appear before separately authorized Wave 1 implementation |

The complete before/after filesystem snapshot excludes `.git/` and `.cache/`. Git object, index, branch, tracked-tree, worktree, untracked, worktree-list, and remote state are verified separately with Git-native read-only checks. The filesystem snapshot records every other directory type/mode, regular-file path/mode/link-count/size/SHA-256, and symlink path/mode/target.

### 4.2 Host, plugin, and protected-input identity

| Binding | Required value |
|---|---|
| Logical action | `L7-FOUNDATION-START-WAVE-1` |
| Host | `codex-cli 0.149.1` on macOS 26.5.2 build `25F84`, `arm64` |
| Marketplace | `level7-dev-loop@personal`; source `local`; path `./plugins/level7-dev-loop`; `AVAILABLE / ON_INSTALL`; category `Developer Tools` |
| Staged source | `/Users/anuppandey/plugins/level7-dev-loop`; installed and enabled at version 0.1.0 |
| Staged/cached manifests | Both SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Activated component | Fresh `/skills` discovery exposes exactly one result resolving to `level7-dev-loop:l7-build` |
| Staged/cached `l7-build` | Both SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package closure | Each package has exactly 13 expected regular single-link files, zero symlinks/other nodes/extras/missing files, owner `anuppandey`, and no group/world-writable entry |
| Package content set | Both SHA-256 `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` over sorted `<digest><two spaces><relative path><LF>` records |
| Historical repository manifest | `.codex-plugin/plugin.json` SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf`; historical evidence only |
| Marketplace observation | `/Users/anuppandey/.agents/plugins/marketplace.json` SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553`; observation, not immutable anchor |
| Parent closure | Every entry in `orchestration-plan-candidate.sha256`, `orchestration-inputs.sha256`, and `harness/foundation-inputs.sha256` verifies |
| Harness | Pinned `make verify` passes immediately before the future attempt; ignored cache effects do not alter source identity |

### 4.3 Fresh activation evidence

A future activation MUST bind, in order:

1. the final SHA-256 and containing Git commit of this exact `L7-AMD-ORC-003` 0.1.0 candidate;
2. a separate final read-only model audit `L7-AUD-ORC-AMD-004`, including its SHA-256 and single-record commit, returning at most `GO_FOR_R3_QUALIFIED_REVIEW`;
3. a separate final structurally independent qualified review `L7-REV-ORC-AMD-003`, including its SHA-256 and single-record commit, returning at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`;
4. fresh nonce `L7-AMD-ORC-003-20260825-01`, with no earlier or concurrent claim;
5. a fresh-thread original accountable-owner AP1 issued after both review gates and before skill selection;
6. a post-AP1 fresh `/skills` selection whose unmodified host token resolves uniquely to `level7-dev-loop:l7-build`; and
7. one original token-prefixed owner reconfirmation-and-dispatch message in that same thread.

Before the review/AP1 chain closes, state is `NOT_AVAILABLE`. The only permitted sequence is:

`NOT_AVAILABLE → UNUSED → IN_USE → CONSUMED`

Review completion plus exact AP1 makes the tuple eligible as `UNUSED`. Dispatch alone enters `IN_USE`. Success, block, failure, preparation error, cancellation, ambiguity, conflict, or recovery consumes it permanently.

## 5. Preflight and transport hardening

The earlier parser failure and the immediate predecessor shell-state failure are different transport classes. A future attempt must address both without weakening one-use consumption:

1. Use short, inspectable read-only tool steps rather than one monolithic validation shell wrapper.
2. Do not assign shell-special or command-search variables, including `path`, `PATH`, `cdpath`, `CDPATH`, `fpath`, `FPATH`, `manpath`, or `MANPATH`.
3. Use direct executable paths for metadata probes where practical, including `/usr/bin/stat` on this bound host.
4. Keep every final authority/Git/path/process/snapshot revalidation step separate from mutation-payload construction.
5. Render both complete Wave 1 file bodies in memory before mutation and validate required sections, exact paths, and terminal newlines.
6. Do not embed shell expansion, repository content, or user content in a dynamically interpolated JavaScript template literal.
7. Validate exactly two add-file targets, zero update/delete/move targets, no overwrite, and no third path.
8. Use one direct inspectable `apply_patch` call with exactly two add-file directives.
9. Immediately before that call, recheck the complete authority chain, Git lineage/cleanliness, paths, output absence, expiry, nonce ownership, and visible sole-writer state.
10. Afterward, repeat Git-native checks and the complete non-`.git`/non-`.cache` snapshot. Exactly the two authorized new untracked regular single-link files may differ.

Any post-dispatch preparation, validation, syntax, transport, or tool error still consumes the attempt. The coordinator must not repair, retry, reconstruct, or resume it.

## 6. Activation and one-use gate

This successor becomes active only when all of the following occur in order:

1. The final candidate bytes and containing Git commit receive a separate read-only model audit that recomputes every parent, predecessor, Git, host, plugin, package, path, nonce, and transport binding.
2. That audit reports no unresolved Blocker, Critical, High, or Medium finding and returns only `GO_FOR_R3_QUALIFIED_REVIEW`.
3. A named qualified human/domain reviewer, structurally independent of the candidate author and every remediator, reviews the exact candidate/audit pair and returns at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`.
4. Only after both reviews, the accountable owner opens a fresh Codex thread and issues exact AP1 for the complete candidate, commit lineage, audit, review, nonce, target, scope, host/plugin identity, A1 ceiling, validity, unused state, and sole-writer condition.
5. Only after AP1, fresh `/skills` selection uniquely resolves the exact `l7-build` component and the owner sends one original token-prefixed reconfirmation-and-dispatch message.
6. The coordinator revalidates everything, runs pinned `make verify`, confirms a clean exact Git tip and absent outputs, corroborates visible sole-writer/open-file state, and captures the complete pre-snapshot.
7. The coordinator satisfies §5, creates the two outputs in one bounded no-overwrite patch, and verifies the exact two-file-only post-state.
8. The coordinator reports both output SHA-256 values, terminal `CONSUMED`, residual limitations, and the mandatory owner-approval stop. It performs no design, implementation, staging, or commit.

Authority expires or is consumed at the earliest of:

- any terminal outcome;
- any partial output or unrelated delta;
- 2026-09-01 23:59:59 Asia/Kathmandu;
- any candidate, commit, audit, review, nonce, AP1, token, host, plugin, marketplace, manifest, skill, package, path, Git, scope, risk, effect, or authority mismatch;
- predecessor replay or changed consumed evidence; or
- owner or reviewer revocation or supersession.

## 7. Fail-closed and recovery rules

Before dispatch, any missing, stale, duplicate, guessed, edited, reconstructed, dirty, divergent, ambiguous, or uninspectable identity blocks without making the tuple available. After dispatch, every block or failure consumes it.

If neither Wave 1 output exists after a terminal failure, report `BLOCKED / CONSUMED`. If either exists without a verified complete pair and clean Git/snapshot post-state, report `RECOVERY_REQUIRED / CONSUMED`. Do not overwrite, delete, stage, commit, clean up, complete, or retry under that tuple.

Hidden-writer absence, same-user session integrity, wall-clock correctness, and tool behavior remain disclosed assumptions. Git improves immutable source identity and diff visibility but is not an OS sandbox, atomic two-file transaction, trusted counter, cryptographic one-use grant, or proof that no hidden writer exists.

## 8. Explicit non-authorization

This draft and its later audit/review do not authorize either Wave 1 output. Even if activated, the action ceiling is only the uncommitted two-file proposal and mandatory stop.

Nothing here authorizes design, implementation, additional Git commits or branches, merge, remote, push, tag, history rewrite, product or harness code, prompts, skills, manifests, packages, dependencies, generated outputs, provider/network trials, root operations, protected infrastructure, publication, deployment, exposure, release, cleanup, self-modification, autonomous continuation, or broader assurance.

## 9. Draft-time evidence

Observed locally on 2026-08-25:

| Check | Result |
|---|---|
| Git baseline | `main` at root commit `08c38b69a2cd63b4adf27873756a09e363e0c5a4`, tree `bcb254506102cc52386e14dc65414face95f4a6b`, 62 tracked files |
| Git topology | Clean worktree/index, one worktree, zero remotes, SHA-1 object format |
| Foundation harness | Pinned `make verify` passed; reproducible binary SHA-256 `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f` |
| Parent/predecessor hashes | Exact matches to the metadata and §4 tables |
| Protected manifests | All entries in the three nested manifests verified |
| Host | `codex-cli 0.149.1`; macOS 26.5.2 build `25F84`; `arm64` |
| Plugin registration | `level7-dev-loop@personal` installed and enabled at 0.1.0 from the bound staged path |
| Stage/cache manifests and build skill | Exact required digests and byte equality |
| Package closure | Each package has 13 expected regular single-link files, no symlinks/other nodes, valid owner/mode, and content-set digest `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` |
| Predecessor failure | Session evidence records exit 127 before mutation and terminal `BLOCKED / CONSUMED` |
| Wave 1 outputs | Both absent |
| Candidate path | Absent immediately before this no-overwrite draft |
| New audit/review/AP1/token | `NOT_YET_CREATED / NOT_EVALUATED` by design |
| Candidate final SHA-256 and commit | Computed and reported only after the no-overwrite write and commit; not self-embedded |

## 10. Compact assurance case

| Element | Statement |
|---|---|
| Claim | A Git-bound successor can be reviewed for one fresh local A1 Wave 1 planning attempt without replaying either consumed predecessor. |
| Argument | Immutable Git lineage, exact host/package bindings, a new nonce and review chain, two-file no-overwrite scope, shell-state and parser hardening, one writer, snapshots, and irreversible consumption confine the proposal. |
| Evidence | Exact parent/predecessor hashes and session terminal record; clean root commit/tree; green pinned harness; current runtime/package closure; fresh nonce; §§4–7 controls. |
| Assumptions | Local Git/hashes/session/clock/process observations are accurate; owner/reviewer facts are truthful; hidden writers and same-user tampering are absent; tools behave as observed. |
| Defeaters | Rewritten ancestry, dirty or foreign delta, predecessor replay, nonce reuse, hidden writer, path/token/component aliasing, clock rollback, compromised tooling, or partial patch. |
| Residual risk | Git strengthens source identity and reviewability but does not provide runtime containment, atomic mutation, trusted identity, or cryptographic replay prevention. |
| Assurance ceiling | Model audit may advance only to qualified review; qualified review only to fresh AP1 consideration; neither activates Wave 1 or grants broader assurance. |

## 11. Exactly one next gate

After the final candidate digest and containing commit are reported, the only permitted next action is a separately authorized read-only model audit `L7-AUD-ORC-AMD-004`. It must bind the exact bytes and Git lineage, independently recompute every inherited and fresh binding, inspect both consumed attempts and both transport mitigations, record all findings and dispositions, and stop at most at `GO_FOR_R3_QUALIFIED_REVIEW`.

No qualified review, AP1, skill selection, Wave 1 output, design, implementation, or external effect is authorized by this draft.
