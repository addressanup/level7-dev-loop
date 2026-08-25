# Level 7 Dev Loop — Git-Baseline Host-Binding Authorization-Lineage Remediation Successor

| Field | Value |
|---|---|
| Artifact ID | L7-AMD-ORC-004 |
| Artifact type | Inert finding-specific successor to rejected candidate L7-AMD-ORC-003 and the retired local host-binding attempts |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT pending a fresh independent Mode B audit of final bytes and Git lineage, a later structurally independent qualified review, fresh AP1, post-AP1 selection, and dispatch** |
| Parent plan | L7-ORC-001 0.3.1 at orchestration-plan.md, SHA-256 a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968 |
| Parent audit | L7-AUD-ORC-001 at orchestration-plan-audit.md, SHA-256 9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91 |
| Parent approval | L7-APR-ORC-001 at orchestration-plan-approval.md, SHA-256 475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d |
| Immediate record predecessor | L7-AMD-ORC-003 0.1.0 at orchestration-plan-host-binding-amendment-git-successor.md, SHA-256 1013bcf73463e11bd11b7b8d744dd6ae55085f6a7d95559efa8a9a2ac9a5df8d; commit 1141c9dd92f437574983abd40448e0113388b4f8; tree dd4b90859214634d90c60b2b5e851215af7c62e7 |
| Predecessor audit | L7-AUD-ORC-AMD-004 at orchestration-plan-host-binding-amendment-git-successor-audit.md, SHA-256 93687b0f66c47cd81ad6678791b78de4f2e8b0a76f76df8bc21f2d253d55384e; NO-GO with AUD-GIT-001 HIGH |
| Predecessor audit commit | 408decb636add15bac42e2eeeed5582d21c3d0f7; tree fa865e615863245335d1dc13b276f2c84160d8f0; parent 1141c9dd92f437574983abd40448e0113388b4f8; single-record audit delta |
| Remediation companion | release-audit-remediation.md |
| Remediation authority | Anup Pandey's current original user-role instruction authorizing only AUD-GIT-001 remediation, the audit-record commit, these two no-overwrite records, pinned offline verification, and one finding-specific commit |
| Remediation transport | level7-dev-loop:l7-release Mode C; staged/cached skill SHA-256 92e1fb180e63b4002414c349ef9ac8d6b00e312c8b9e866f9311346007fcec8f |
| Logical action if later activated | L7-FOUNDATION-START-WAVE-1 |
| Drafting effect / activated effect ceiling | A1 successor-governance record / one local A1 Wave 1 contract-and-specification proposal |
| Governing risk | R3 authorization identity, source transition, and one-use binding |
| Fresh proposed nonce | L7-AMD-ORC-004-20260825-01 |
| Proposed validity ceiling | Earliest invalidation below, including 2026-09-01 23:59:59 Asia/Kathmandu |
| Candidate path | docs/artifacts/orchestration-plan-host-binding-amendment-git-successor-remediation.md |
| Next gate | Fresh, separately authorized, independent Mode B audit of this final candidate, its companion remediation record, both digests, and their containing Git commit |

## 1. Finding-specific correction and immutable predecessor record

L7-AMD-ORC-003 remains byte-for-byte immutable. Its line 32 says that the L7-AMD-ORC-002 attempt had an exact fresh-thread AP1 before post-AP1 selection and dispatch. Audit L7-AUD-ORC-AMD-004 found no inspectable original AP1 before selection in the candidate-named session and no separate user-role AP1 in the visible exact-nonce search. This successor does not repeat, repair, infer, or reconstruct that missing authority.

The inspectable evidence supports only the following bounded account of the L7-AMD-ORC-002 attempt:

1. The candidate-named session is /Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T10-56-30-01a03754-ab28-7f32-9f97-31de915de029.jsonl.
2. Its first substantive user instruction is a token-prefixed attempted dispatch at JSONL line 9. That message claims an earlier AP1, but no earlier inspectable AP1 message appears in the session.
3. During pre-mutation validation, a zsh loop assigned the shell-special lowercase variable path, which is tied to command lookup. The next unqualified stat invocation terminated with exit code 127 and zsh:40: command not found: stat.
4. The failure occurred before make verify, the complete pre-snapshot, candidate-body construction, or any apply_patch call.
5. Terminal evidence records both Wave 1 outputs absent.
6. These facts justify irreversible anti-replay treatment. They do not establish that the tuple validly became UNUSED or validly transitioned from UNUSED to IN_USE.

The local session record is same-user mutable corroboration, not a signed ledger, trusted identity source, or trusted counter.

### 1.1 Nonce dispositions

| Nonce | Evidence-bounded disposition |
|---|---|
| L7-AMD-ORC-001-20260825-02 | BLOCKED / CONSUMED / NON-REPLAYABLE after a verified original AP1, token-prefixed dispatch, parser failure before mutation, and both outputs absent |
| L7-AMD-ORC-002-20260825-01 | **PERMANENTLY RETIRED / SUPERSEDED / NON-REPLAYABLE** after a token-prefixed attempted dispatch, terminal exit-127 shell failure, and both outputs absent; no inspectable original AP1 before selection and no asserted valid UNUSED-to-IN_USE transition |
| L7-AMD-ORC-003-20260825-01 | Never available because its candidate received NO-GO; now superseded and non-replayable without AP1, selection, dispatch, or state transition |
| L7-AMD-ORC-004-20260825-01 | Fresh proposed nonce; repository and visible active/archived session search found no prior or concurrent activation, ownership, AP1, selection, or dispatch claim; remains NOT_AVAILABLE |

Neither predecessor attempt is retried, repaired, resumed, or reinterpreted. No predecessor authority or state is borrowed by this candidate.

## 2. Git baseline and remediation lineage

The separately owner-authorized Git baseline remains:

1. local branch main;
2. SHA-1 root commit 08c38b69a2cd63b4adf27873756a09e363e0c5a4;
3. root tree bcb254506102cc52386e14dc65414face95f4a6b with exactly 62 tracked files;
4. one worktree at the canonical root; and
5. zero remotes and zero replacement refs.

Candidate L7-AMD-ORC-003 is the direct single-file child of that root. Its audit is the direct single-record child commit 408decb636add15bac42e2eeeed5582d21c3d0f7. This candidate and release-audit-remediation.md must be committed together as exactly one two-record direct child of that audit commit, with no other tracked, index, worktree, or untracked non-ignored delta. Final record digests, commit identity, and tree are computed only after the write and commit and are reported in the completion handoff.

This record preserves all earlier commits and artifacts. It does not authorize a remote, another worktree, branch protection, publication, push, rewrite, tag, merge, release, or any Git mutation beyond the one expressly authorized two-record remediation commit.

## 3. Narrow normative successor

This section has no activated effect until every gate in section 6 closes.

Once activated, and only for the Codex CLI route, this successor would permit one new attempt at logical action L7-FOUNDATION-START-WAVE-1. The attempt may create only:

1. docs/artifacts/wave-01-change-contract.md; and
2. docs/artifacts/wave-01-specification.md.

Both must be new regular single-link files created without overwrite. They remain uncommitted proposal files at the mandatory owner-approval stop. The future invocation does not authorize git add, commit, branch creation, merge, remote creation, push, or any other Git mutation.

This successor changes only the inaccurate predecessor authorization lineage and establishes a new inert source identity and nonce. It does not alter immutable parent, predecessor, audit, review, commit, or terminal evidence. All other L7-ORC-001 boundaries remain in force, including:

- specification before design and implementation;
- one primary writer;
- no product, harness, prompt, skill, manifest, dependency, generated-package, provider, network, root, publication, deployment, exposure, release, cleanup, or self-modification effect;
- no automatic continuation; and
- no product, security, compatibility, compliance, release, or deployment assurance claim.

Git existence resolves only the mechanical absence recorded by pre-wave gate PW-02. Product implementation remains blocked until the owner approves the Wave 1 specification and design and every other applicable pre-wave decision closes.

## 4. Exact inherited and Git-aware binding

### 4.1 Project and Git source identity

Every future audit and invocation preflight MUST recompute these fields. Missing, dirty, divergent, ambiguous, partially matching, or changed state is BLOCKED.

| Binding | Required value |
|---|---|
| Canonical project root | /Users/anuppandey/Desktop/level7-dev-loop; every component resolves directly without a symlink |
| Canonical output parent | /Users/anuppandey/Desktop/level7-dev-loop/docs/artifacts; direct directories with no symlink component |
| Repository identity | Git repository rooted exactly at the canonical project root; object format sha1 |
| Foundation baseline | Root commit 08c38b69a2cd63b4adf27873756a09e363e0c5a4; tree bcb254506102cc52386e14dc65414face95f4a6b; exactly 62 tracked files |
| Rejected candidate | Commit 1141c9dd92f437574983abd40448e0113388b4f8; tree dd4b90859214634d90c60b2b5e851215af7c62e7; parent Foundation baseline; exact single-file addition |
| Recorded audit | Commit 408decb636add15bac42e2eeeed5582d21c3d0f7; tree fa865e615863245335d1dc13b276f2c84160d8f0; parent rejected candidate; exact single-record addition |
| Branch strategy | Local main; one primary writer; exactly one worktree at the canonical root; zero remotes |
| Remediation candidate lineage | This candidate and release-audit-remediation.md must share one direct two-file child commit of the recorded audit; their final digests and containing commit are supplied post-commit |
| Review lineage | Each future audit and qualified-review commit must be a linear descendant, add only its separately authorized record, and bind its exact parent |
| Invocation tip | Exact qualified-review-chain tip named in the future AP1; main points to it; no detached HEAD, merge, rewrite, replacement, shallow ambiguity, or unexpected object-format change |
| Clean admission state | No tracked worktree or index delta and no untracked non-ignored path before the future attempt |
| Ignored state | .cache remains ignored and is excluded from source identity and complete workspace snapshots |
| Permitted output state | Both Wave 1 outputs absent in all forms before mutation; afterward they are the only two untracked non-ignored paths |
| Product surfaces | No reserved product path may appear before separately authorized Wave 1 implementation |

The complete before/after filesystem snapshot excludes .git and .cache. Git object, index, branch, tracked-tree, worktree, untracked, worktree-list, and remote state are verified separately with Git-native read-only checks. The filesystem snapshot records every other directory type/mode, regular-file path/mode/link-count/size/SHA-256, and symlink path/mode/target.

### 4.2 Host, plugin, and protected-input identity

| Binding | Required value |
|---|---|
| Logical action | L7-FOUNDATION-START-WAVE-1 |
| Host | codex-cli 0.149.1 on macOS 26.5.2 build 25F84, arm64 |
| Marketplace | level7-dev-loop@personal; source local; path ./plugins/level7-dev-loop; AVAILABLE / ON_INSTALL; category Developer Tools |
| Staged source | /Users/anuppandey/plugins/level7-dev-loop; installed and enabled at version 0.1.0 |
| Staged/cached manifests | Both SHA-256 202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3 |
| Activated component | Fresh /skills discovery exposes exactly one result resolving to level7-dev-loop:l7-build |
| Staged/cached l7-build | Both SHA-256 ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f |
| Package closure | Each package has exactly 13 expected regular single-link files, zero symlinks/other nodes/extras/missing files, owner anuppandey, and no group/world-writable entry |
| Package content set | Both SHA-256 b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04 over sorted digest, two spaces, relative path, and LF records |
| Historical repository manifest | .codex-plugin/plugin.json SHA-256 b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf; historical evidence only |
| Marketplace observation | /Users/anuppandey/.agents/plugins/marketplace.json SHA-256 fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553; observation, not immutable anchor |
| Parent closure | Every entry in orchestration-plan-candidate.sha256, orchestration-inputs.sha256, and harness/foundation-inputs.sha256 verifies |
| Harness | Pinned offline make verify passes immediately before the future attempt; ignored cache effects do not alter source identity |

### 4.3 Fresh activation evidence

A future activation MUST bind, in order:

1. the final SHA-256 and containing two-record Git commit of this exact L7-AMD-ORC-004 0.1.0 candidate and its companion remediation record;
2. a fresh, separately authorized, independent read-only Mode B audit of those exact bytes and Git lineage, including its final SHA-256 and single-record commit, returning at most GO_FOR_R3_QUALIFIED_REVIEW;
3. a separate final structurally independent qualified review, including its SHA-256 and single-record commit, returning at most GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW;
4. fresh nonce L7-AMD-ORC-004-20260825-01, with no earlier or concurrent claim;
5. a fresh-thread original accountable-owner AP1 issued after both review gates and before skill selection;
6. a post-AP1 fresh /skills selection whose unmodified host token resolves uniquely to level7-dev-loop:l7-build; and
7. one original token-prefixed owner reconfirmation-and-dispatch message in that same thread.

Before the review/AP1 chain closes, state is NOT_AVAILABLE. The only permitted sequence for this new tuple is:

NOT_AVAILABLE → UNUSED → IN_USE → CONSUMED

Review completion plus exact AP1 makes only this tuple eligible as UNUSED. Dispatch alone enters IN_USE. Success, block, failure, preparation error, cancellation, ambiguity, conflict, or recovery consumes it permanently. No predecessor state, assertion, or authority can satisfy any transition.

## 5. Preflight and transport hardening

The earlier parser failure and the later shell-state failure are different transport classes. A future attempt must address both without weakening one-use consumption:

1. Use short, inspectable read-only tool steps rather than one monolithic validation shell wrapper.
2. Do not assign shell-special or command-search variables, including path, PATH, cdpath, CDPATH, fpath, FPATH, manpath, or MANPATH.
3. Use direct executable paths for metadata probes where practical, including /usr/bin/stat on this bound host.
4. Keep every final authority/Git/path/process/snapshot revalidation step separate from mutation-payload construction.
5. Render both complete Wave 1 file bodies in memory before mutation and validate required sections, exact paths, and terminal newlines.
6. Do not embed shell expansion, repository content, or user content in a dynamically interpolated JavaScript template literal.
7. Validate exactly two add-file targets, zero update/delete/move targets, no overwrite, and no third path.
8. Use one direct inspectable apply_patch call with exactly two add-file directives.
9. Immediately before that call, recheck the complete authority chain, Git lineage/cleanliness, paths, output absence, expiry, nonce ownership, and visible sole-writer state.
10. Afterward, repeat Git-native checks and the complete non-.git/non-.cache snapshot. Exactly the two authorized new untracked regular single-link files may differ.

Any post-dispatch preparation, validation, syntax, transport, or tool error still consumes the future attempt. The coordinator must not repair, retry, reconstruct, or resume it.

## 6. Activation and one-use gate

This successor becomes active only when all of the following occur in order:

1. The final candidate and companion bytes and their containing Git commit receive a fresh independent Mode B audit that recomputes every parent, predecessor, Git, host, plugin, package, path, nonce, and transport binding.
2. That audit reports no unresolved Blocker, Critical, High, or Medium finding and returns only GO_FOR_R3_QUALIFIED_REVIEW.
3. A named qualified human/domain reviewer, structurally independent of the candidate author, this remediator, and the fresh auditor, reviews the exact candidate/audit pair and returns at most GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW.
4. Only after both reviews, the accountable owner opens a fresh Codex thread and issues exact AP1 for the complete candidate, commit lineage, audit, review, nonce, target, scope, host/plugin identity, A1 ceiling, validity, unused state, and sole-writer condition.
5. Only after AP1, fresh /skills selection uniquely resolves the exact l7-build component and the owner sends one original token-prefixed reconfirmation-and-dispatch message.
6. The coordinator revalidates everything, runs pinned offline make verify, confirms a clean exact Git tip and absent outputs, corroborates visible sole-writer/open-file state, and captures the complete pre-snapshot.
7. The coordinator satisfies section 5, creates the two outputs in one bounded no-overwrite patch, and verifies the exact two-file-only post-state.
8. The coordinator reports both output SHA-256 values, terminal CONSUMED, residual limitations, and the mandatory owner-approval stop. It performs no design, implementation, staging, or commit.

Authority expires or is consumed at the earliest of:

- any terminal outcome;
- any partial output or unrelated delta;
- 2026-09-01 23:59:59 Asia/Kathmandu;
- any candidate, commit, audit, review, nonce, AP1, token, host, plugin, marketplace, manifest, skill, package, path, Git, scope, risk, effect, or authority mismatch;
- predecessor replay or changed retired/consumed evidence; or
- owner or reviewer revocation or supersession.

## 7. Fail-closed and recovery rules

Before dispatch, any missing, stale, duplicate, guessed, edited, reconstructed, dirty, divergent, ambiguous, or uninspectable identity blocks without making the tuple available. After dispatch, every block or failure consumes it.

If neither Wave 1 output exists after a terminal failure, report BLOCKED / CONSUMED. If either exists without a verified complete pair and clean Git/snapshot post-state, report RECOVERY_REQUIRED / CONSUMED. Do not overwrite, delete, stage, commit, clean up, complete, or retry under that tuple.

Hidden-writer absence, same-user session integrity, wall-clock correctness, and tool behavior remain disclosed assumptions. Git improves immutable source identity and diff visibility but is not an OS sandbox, atomic two-file transaction, trusted counter, cryptographic one-use grant, or proof that no hidden writer exists.

## 8. Explicit non-authorization

This remediation draft and its future audit/review do not authorize either Wave 1 output. Even if later activated, the action ceiling is only the uncommitted two-file proposal and mandatory stop.

Nothing here authorizes qualified review in this session, AP1, skill selection, Wave 1 work, design, implementation, additional Git commits or branches, merge, remote, push, tag, history rewrite, product or harness code, prompts, skills, manifests, packages, dependencies, generated outputs, provider/network trials, root operations, protected infrastructure, publication, deployment, exposure, release, cleanup, self-modification, autonomous continuation, or broader assurance.

## 9. Remediation-time evidence

Observed locally on 2026-08-25:

| Check | Result |
|---|---|
| Admission Git tip | Audit commit 408decb636add15bac42e2eeeed5582d21c3d0f7, tree fa865e615863245335d1dc13b276f2c84160d8f0, parent 1141c9dd92f437574983abd40448e0113388b4f8; clean tracked worktree/index |
| Audit commit scope | Exactly one addition: orchestration-plan-host-binding-amendment-git-successor-audit.md |
| Candidate and audit bytes | Exact supplied SHA-256 values reproduced |
| Git topology | Local main, one worktree, zero remotes/replacements, SHA-1, non-shallow; git fsck --full --strict passed |
| Foundation baseline | Exact root commit/tree and 62-file count reproduced |
| Protected manifests | Every entry in all three nested manifests verified |
| Canonical paths | Project root and output parent resolve exactly; no project-root-down symlink outside excluded .git/.cache |
| Wave 1 outputs | Both absent in file and symlink forms |
| Host/plugin/package | All values in section 4.2 reproduced; staged and cached packages are byte-identical, each with 13 valid files and the required content-set digest |
| Prior attempted dispatch | Session evidence records token-prefixed dispatch, exit 127 before mutation, and both outputs absent, but no inspectable original AP1 before selection |
| New ID and nonce | Before adoption, exact repository and Git-history search found no occurrence. The only substantive pre-adoption session sources were the prior auditor's proposed prompt and the current owner's Mode C authorization; later assistant/tool echoes are remediation mechanics. No activation, ownership, AP1, selection, or dispatch claim was found |
| Foundation harness | Pinned offline make verify completed at 2026-08-25 12:57:52 Asia/Kathmandu with exit 0: no module dependencies to download; all modules verified; Foundation scope PASS; import boundaries PASS; both Go test invocations passed; reproducible binary SHA-256 1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f |
| Final candidate/report SHA-256 and commit | Computed after final record bytes and the authorized two-record commit; not self-embedded |

## 10. Compact assurance case

| Element | Statement |
|---|---|
| Claim | This inert successor corrects AUD-GIT-001 without changing or laundering immutable predecessor evidence. |
| Argument | It rejects the unsupported AP1 and state-transition claim, preserves the observable dispatch/failure/output facts, retires predecessor nonces, establishes a searched fresh nonce, and retains every inherited scope, Git, host, package, path, transport, validity, and fail-closed control. |
| Evidence | Exact candidate and audit hashes/commits; audit finding; session event ordering; clean Git topology; protected manifests; absent outputs; current runtime/package closure; exact nonce search; companion remediation record. |
| Assumptions | Local Git, hashes, sessions, clock, process observations, and tools are accurate; owner/reviewer facts are truthful; hidden writers and same-user tampering are absent. |
| Defeaters | Rewritten ancestry, dirty or foreign delta, missing evidence, predecessor replay, nonce claim/reuse, hidden writer, path/token/component aliasing, clock rollback, compromised tooling, partial patch, or any mismatched binding. |
| Residual risk | Git and hashes strengthen reviewability but do not supply trusted identity, immutable sessions, runtime containment, atomic mutation, trusted time/counter, or cryptographic replay prevention. |
| Assurance ceiling | This Mode C record cannot clear its own finding or issue GO; only a fresh independent Mode B audit may assess progression to qualified review. |

## 11. Exactly one next gate

After both final record digests and their containing commit are reported, the only permitted next action is a separately authorized fresh independent level7-dev-loop:l7-release Mode B audit. It must bind this candidate, release-audit-remediation.md, their exact bytes and two-record commit, independently reproduce the correction and every inherited binding, record all findings and dispositions, and stop at most at GO_FOR_R3_QUALIFIED_REVIEW.

No qualified review, AP1, skill selection, Wave 1 output, design, implementation, release, deployment, or external effect is authorized by this record.
