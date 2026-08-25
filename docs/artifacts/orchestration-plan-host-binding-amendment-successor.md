# Level 7 Dev Loop — Codex Host-Binding Successor Amendment

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-ORC-002` |
| Artifact type | One-use successor to the consumed local Codex host-binding amendment |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT pending fresh exact-digest model audit, structurally independent qualified review, fresh current-session AP1, selection, and dispatch** |
| Parent plan | [`L7-ORC-001` 0.3.1](orchestration-plan.md), SHA-256 `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| Parent audit | [`L7-AUD-ORC-001`](orchestration-plan-audit.md), SHA-256 `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` |
| Parent approval | [`L7-APR-ORC-001`](orchestration-plan-approval.md), SHA-256 `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` |
| Parent input freeze | [`orchestration-inputs.sha256`](orchestration-inputs.sha256), SHA-256 `ef17c49d7ceae115b476c2945fba4149f63094beade4cf8c0ba2d4cf652d2b16` |
| Predecessor candidate | [`L7-AMD-ORC-001` 0.1.1](orchestration-plan-host-binding-amendment.md), SHA-256 `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` |
| Predecessor model audit | [`L7-AUD-ORC-AMD-002`](principal-engineer-release-audit.md), SHA-256 `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c` |
| Predecessor qualified review | [`L7-REV-ORC-AMD-001`](orchestration-plan-host-binding-amendment-qualified-review.md), SHA-256 `85187c07a4a44b249e373e75718f93f813401f6090a60a5f191b8e7a0b550e26` |
| Predecessor nonce/state | `L7-AMD-ORC-001-20260825-02` — **CONSUMED; non-replayable** |
| Draft authority | Current user-role message from Anup Pandey authorizing exactly one local A1 artifact-only draft of this inert successor, at exactly this no-overwrite path, with no Wave 1 artifacts or other mutation |
| Draft transport | `level7-dev-loop:l7-greenfield`; staged/cached SHA-256 `6c76a16af74b932733f3a1ea0838fef67fe2c5cbaf6a6aab22777949c8866609` |
| Logical action if later activated | `L7-FOUNDATION-START-WAVE-1` |
| Drafting effect / activated effect ceiling | A1 successor-governance record / one local A1 Wave 1 contract-and-specification proposal |
| Successor risk | `R3` authorization-identity and one-use transition binding |
| Fresh nonce | `L7-AMD-ORC-002-20260825-01` |
| Proposed validity ceiling | Earliest invalidation below, including 2026-09-01 23:59:59 Asia/Kathmandu |
| Next gate | Separate read-only model audit of these exact final bytes; no activation or Wave 1 dispatch |

## 1. Reason and predecessor terminal record

The predecessor route completed its review chain and received a current-session AP1, a fresh host-selected `$level7-dev-loop:l7-build` token, and a token-prefixed dispatch. Its invocation preflight recomputed the bound evidence and runtime identities, confirmed the output paths were absent, confirmed Git was absent, ran the pinned Foundation harness successfully, and froze an 85-entry non-cache workspace snapshot with SHA-256 `c157fac631a954e90301b3434c35e7f943e48f1494e824d99e10ffa7c2b238d7`.

Before either authorized output was created, the final orchestration wrapper failed to parse with:

`SyntaxError: Missing } in template expression`

No patch call executed. Read-only terminal checks established:

- `docs/artifacts/wave-01-change-contract.md` was absent;
- `docs/artifacts/wave-01-specification.md` was absent;
- Git remained absent;
- the 85-entry non-cache snapshot was byte-for-byte unchanged; and
- no partial output existed, so `RECOVERY_REQUIRED` did not apply.

Under `L7-AMD-ORC-001` §4, the dispatch had already transitioned its tuple from `UNUSED` to `IN_USE`. The syntax failure then transitioned it irrevocably to `CONSUMED`. The in-memory draft pair was never a repository artifact, approved candidate, authority source, or resumable payload. It MUST NOT be recovered, replayed, or treated as evidence that the two files existed.

This successor does not reinterpret the predecessor result as unused and does not repair or edit predecessor evidence. It provides a distinct, inert candidate with a fresh nonce and a fresh review/approval chain.

## 2. Narrow normative successor

This section has no effect until every activation gate in §5 closes.

Once activated, and only for the Codex CLI route, this successor would permit one new attempt at the same logical action, canonical target, two-file scope, and A1 ceiling as the predecessor. It supersedes only the predecessor candidate's availability for a future invocation. It does not supersede or erase the predecessor's terminal `CONSUMED` evidence.

Every other approved `L7-ORC-001` and predecessor boundary remains unchanged, including:

- one Wave 1 planning pair followed by a mandatory owner-approval stop;
- separate later stops for design and implementation;
- exact component discovery and fail-closed token preservation;
- no Git implementation while Git is absent;
- no product, harness, prompt, skill, manifest, dependency, generated-package, host, provider, network, root, publication, deployment, exposure, release, cleanup, self-modification, or autonomous effect;
- no automatic continuation; and
- no product, security, compatibility, release, or deployment assurance claim.

The predecessor audit and qualified review are historical provenance only. They do not audit, review, activate, or approve these successor bytes.

## 3. Exact inherited local binding

### 3.1 Audit-time and invocation-preflight bindings

Every binding below MUST be independently recomputed by the successor model audit and MUST match again during any later invocation preflight. Missing, stale, ambiguous, partially matching, or changed state is `BLOCKED`.

| Binding | Required value |
|---|---|
| Logical action | `L7-FOUNDATION-START-WAVE-1` |
| Canonical project root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Canonical output parent | `/Users/anuppandey/Desktop/level7-dev-loop/docs/artifacts`; the project root and each path component must resolve directly without a symlink |
| Permitted activated outputs | Only `docs/artifacts/wave-01-change-contract.md` and `docs/artifacts/wave-01-specification.md`, both new regular single-link files created without overwrite |
| Host observation | `codex-cli 0.149.1` on macOS 26.5.2 build `25F84`, `arm64` |
| Marketplace identity | Normalized `level7-dev-loop@personal` entry; local source `./plugins/level7-dev-loop`; policy `AVAILABLE / ON_INSTALL`; category `Developer Tools` |
| Resolved staged source | `/Users/anuppandey/plugins/level7-dev-loop`; `codex plugin list` must report `installed, enabled`, version `0.1.0` |
| Staged manifest | `/Users/anuppandey/plugins/level7-dev-loop/.codex-plugin/plugin.json`; SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Effective cached manifest | `/Users/anuppandey/.codex/plugins/cache/personal/level7-dev-loop/0.1.0/.codex-plugin/plugin.json`; SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Activated component | Fresh `/skills` discovery must expose exactly one result resolving to `level7-dev-loop:l7-build` |
| Staged/cached activated skill | Both `skills/l7-build/SKILL.md` files; SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package closure | Each staged/cached package has exactly 13 regular files—one manifest and 12 expected skill files—with zero symlinks, no hardlink count other than 1, no extra/missing file, local owner `anuppandey`, and no group/world write bit |
| Package content-set SHA-256 | Both packages: `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` over the predecessor-defined sorted content-set encoding |
| Historical repository manifest | `.codex-plugin/plugin.json` SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf`; historical evidence only |
| Marketplace observation | `/Users/anuppandey/.agents/plugins/marketplace.json` SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553`; evidence, not an immutable validity anchor |
| Parent-chain closure | Every entry in `orchestration-plan-candidate.sha256`, `orchestration-inputs.sha256`, and `harness/foundation-inputs.sha256` must verify |
| Harness | Pinned baseline `make verify` must pass immediately before the activated attempt; shadow and unrun environments retain their existing evidence labels |
| Git and outputs | Git must remain absent and both activated outputs must remain absent, including file, directory, symlink, and broken-symlink forms |
| Predecessor terminal state | Nonce `L7-AMD-ORC-001-20260825-02` remains `CONSUMED`; both predecessor outputs remain absent; no replay or state reset is permitted |
| Fresh nonce | `L7-AMD-ORC-002-20260825-01`; it must have no earlier or concurrent claim |

The `l7-greenfield` component is only the transport for drafting this inert successor. It is not the activated Wave 1 component and supplies no future authority.

### 3.2 Fresh activation evidence chain

A later activation MUST bind:

1. the exact final SHA-256 of this `L7-AMD-ORC-002` candidate;
2. a new final read-only model-audit record, expected artifact ID `L7-AUD-ORC-AMD-003`, that binds this exact candidate and returns at most `GO_FOR_R3_QUALIFIED_REVIEW`;
3. a new final structurally independent qualified-review record, expected artifact ID `L7-REV-ORC-AMD-002`, that binds both successor candidate and successor audit and returns at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`;
4. the fresh nonce `L7-AMD-ORC-002-20260825-01`;
5. a fresh-thread, original accountable-owner AP1 binding the complete successor tuple;
6. a post-AP1 fresh `/skills` selection whose unmodified returned token resolves uniquely to `level7-dev-loop:l7-build`; and
7. one original token-prefixed owner reconfirmation-and-dispatch message in that same fresh thread.

The new audit and qualified review require new no-overwrite artifacts at separately authorized paths. Their IDs and expected roles in this section do not authorize those writes.

### 3.3 One-attempt state

Before the new audit/review/AP1 chain closes, the successor attempt state is `NOT_AVAILABLE`, not `UNUSED`. A valid token-prefixed confirmation-and-dispatch message is the sole transition:

`NOT_AVAILABLE → UNUSED → IN_USE → CONSUMED`

The review chain and exact AP1 make the tuple eligible as `UNUSED`. Dispatch changes it once to `IN_USE`. Terminal success, block, failure, cancellation, ambiguity, conflict, or recovery changes it to `CONSUMED`. It never returns to `UNUSED`, and predecessor state cannot be borrowed.

### 3.4 Current sole-writer state

A later owner AP1 must freshly confirm that no other mutation-capable thread or process will write the canonical project during the bounded attempt. The coordinator must corroborate that statement with host-visible process and open-file checks. Hidden-writer absence remains a disclosed governance assumption; ambiguity blocks and consumes after dispatch.

## 4. Pre-write transport hardening

The prior failure was caused by orchestration-payload syntax, not a target-file conflict. The successor therefore adds these mandatory local transport checks without weakening one-use consumption:

1. Render both complete candidate file bodies in memory before the first write.
2. Validate required sections, exact output paths, terminal newlines, and the absence of any third patch target.
3. Keep final authority/path/process/snapshot revalidation in a separate completed tool step from mutation-payload construction.
4. Construct the mutation payload only from already-rendered data. Do not embed shell parameter expansion, repository content, or user content in a dynamically interpolated JavaScript template literal.
5. Use one direct host `apply_patch` call, or an equivalently inspectable no-overwrite primitive, containing exactly two add-file directives and no update/delete/move directive.
6. Before calling the mutation primitive, validate the complete payload structure in memory: one add directive for each exact authorized path, zero other targets, no overwrite directive, and both complete bodies.
7. Immediately before mutation, recheck the exact successor chain, fresh nonce ownership, expiry/revocation, canonical paths, output absence, and host-visible sole-writer state.
8. Any preparation, validation, syntax, transport, or tool error after dispatch still consumes the attempt. The coordinator MUST NOT repair, retry, reconstruct from the failed payload, or reinterpret the state as unused.

These checks are model/tool governance in a same-user mutable host. They do not make the two-file operation atomic or establish a hardened transaction.

## 5. Activation and one-use gate

This successor becomes active only when all of the following occur in order:

1. These exact final bytes receive a separate post-write, read-only model audit that recomputes the successor digest, all parent/predecessor/runtime/package bindings, the consumed-attempt evidence, the new nonce, the §4 transport controls, and every fail-closed rule.
2. That audit reports no unresolved Blocker, Critical, High, or Medium finding and returns only `GO_FOR_R3_QUALIFIED_REVIEW`. It cannot issue product, security, compatibility, release, or deployment assurance.
3. The exact successor and exact successor audit receive a read-only review by a named qualified human/domain reviewer structurally independent of this successor's author and every remediator. The final record must include identity, role, qualification evidence, independence/conflicts, methods, findings/dispositions, residual risk, scoped decision, and validity. It may return at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`.
4. Only after both review gates close does the accountable owner open a fresh Codex thread and, before selection, issue an exact current-session AP1 binding every literal successor value, both fresh review records and hashes, the new nonce, target, two-file scope, host/source identity, A1 ceiling, validity, unused state, and sole-writer condition.
5. After that AP1, a fresh `/skills` selection uniquely resolves and inserts the exact `level7-dev-loop:l7-build` token. The owner submits one original message beginning with that unmodified token and restating the complete successor AP1 tuple plus token/component resolution.
6. Immediately after dispatch and before any write, the coordinator transitions only this successor tuple to `IN_USE`; revalidates the complete chain, bindings, expiry, revocation, outputs, Git absence, harness, and sole-writer state; and captures a complete in-memory non-cache workspace snapshot.
7. The coordinator renders and validates the complete pair, satisfies §4, performs final pre-write revalidation, creates both outputs as one bounded no-overwrite change, and repeats the snapshot. Exactly the two new regular single-link files may differ. Every prior regular file, symlink, and directory type/mode must match; only derived directory metadata caused by the additions may be ignored.
8. The coordinator reports exact output hashes, terminal state `CONSUMED`, limitations, and the mandatory owner-approval stop. It does not continue to design or implementation.

Authority expires or is consumed at the earliest of:

- any terminal outcome of the one successor attempt;
- any partial output or unrelated delta;
- 2026-09-01 23:59:59 Asia/Kathmandu;
- any candidate, audit, review, nonce, AP1, token, host, plugin, manifest, skill, package, marketplace, path, parent/predecessor chain, scope, risk, effect, or authority mismatch;
- discovery that the predecessor was replayed or its consumed evidence changed; or
- owner revocation or supersession.

## 6. Fail-closed and recovery rules

Before dispatch, a missing, ambiguous, stale, duplicate, guessed, edited, reconstructed, or uninspectable identity blocks without making the successor tuple available. After dispatch, any block or failure consumes it.

If neither Wave 1 output exists after a terminal failure, the result is `BLOCKED / CONSUMED`. If either output exists without a verified complete pair and clean post-snapshot, the result is `RECOVERY_REQUIRED / CONSUMED`. No automatic overwrite, deletion, cleanup, completion, or retry is authorized.

The coordinator MUST NOT repair a mismatch by editing or reinstalling the plugin, marketplace, source, cache, repository manifest, skill, package, parent evidence, approval, environment, validator, or nonce; by changing hosts; by adding a cachebuster; or by weakening a check. Any repair requires explicit successor authority and a new exact-digest review chain.

## 7. Explicit non-authorization

Drafting, hashing, auditing, reviewing, or approving this artifact does not invoke Wave 1. The proposed successor remains inert until §5 closes.

It does not authorize:

- either Wave 1 output;
- design or implementation;
- Git initialization/import, branches, commits, merges, remotes, or release history;
- product or harness code, prompts, semantic workflows, skills, manifests, packages, dependencies, or generated outputs;
- provider/network/actual-host trials, root operations, protected infrastructure, publication, deployment, exposure, release, cleanup, or external effects;
- reuse of the predecessor AP1, token, review chain, or nonce; or
- product, security, compatibility, compliance, release, or deployment claims.

## 8. Draft-time evidence

Observed on 2026-08-25 at approximately 10:02 Asia/Kathmandu:

| Check | Draft result |
|---|---|
| Successor output path | `ABSENT`, including symlink form, before the draft |
| Wave 1 output paths | Both `ABSENT` |
| Git | `ABSENT` |
| Parent and predecessor hashes | Exact matches to §1 |
| Frozen manifests | `orchestration-plan-candidate.sha256`, `orchestration-inputs.sha256`, and `harness/foundation-inputs.sha256` all verified |
| Host | `codex-cli 0.149.1`; macOS 26.5.2 build `25F84`; `arm64` |
| Plugin registration | `level7-dev-loop@personal` installed and enabled at version `0.1.0` from the bound staged path |
| Stage/cache manifests | Both exact SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Stage/cache build skill | Both exact SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package closure | Each has 13 regular files, zero symlinks/extras/hardlinks, correct owner/mode; both content sets equal `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` |
| Foundation harness | `make verify` passed using pinned Go 1.26.7; repeat binary SHA-256 `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f` |
| Non-cache workspace | 85-entry snapshot SHA-256 `c157fac631a954e90301b3434c35e7f943e48f1494e824d99e10ffa7c2b238d7` before this one permitted draft |
| Foreign writer evidence | No non-CWD target file handle or active foreign target mutation command observed; hidden-writer absence is not proven |
| Successor model audit/review/AP1/token | `NOT_YET_CREATED / NOT_EVALUATED` by design |
| Successor final SHA-256 | Computed only after the no-overwrite write and reported in the handoff; not self-embedded |

These observations are local drafting evidence, not the required successor audit or qualified review.

## 9. Compact assurance case

| Element | Statement |
|---|---|
| Claim | This successor can be reviewed for one new local A1 contract/specification attempt without replaying the consumed predecessor; it cannot establish broader assurance. |
| Argument | Exact predecessor consumption is preserved; a new nonce and wholly fresh audit/review/AP1/selection chain prevent governance reuse; inherited host/package bindings remain fail closed; literal payload validation addresses the observed syntax-failure class; no-overwrite paths, one writer, snapshots, and irreversible consumption confine the attempted effect. |
| Evidence | Exact parent/predecessor hashes; terminal zero-output/unchanged-snapshot observation; draft-time runtime/package/manifests; green pinned harness; new nonce; §§3–6 controls. |
| Assumptions | Local hashes, clock, process, path, and link observations are accurate at check time; owner and future reviewer facts are truthful; hidden writers and same-user tampering are absent; host tool behavior matches observation. |
| Defeaters | Replay or alteration of predecessor evidence; hidden writer; same-user TOCTOU; path/token/component aliasing; clock rollback; compromised tooling; payload-validator defect; partial two-file write; or any binding drift. Each blocks or yields recovery and consumes after dispatch. |
| Residual risk | Governance and local tool checks are not cryptographic replay prevention, OS containment, a trusted counter, or an atomic transaction. A fresh attempt can still fail and be consumed. |
| Assurance ceiling | The model audit may advance only to qualified review; qualified review may advance only to AP1 consideration; neither may issue activation, product, security, release, or deployment `PASS/GO`. |

## 10. Exactly one next gate

After this candidate is written and its exact SHA-256 is reported, the only permitted next action is a separate, read-only model audit of exact `L7-AMD-ORC-002` version 0.1.0. The audit must bind the final digest, recompute every inherited and fresh binding, examine the predecessor terminal evidence and §4 transport mitigation, record all findings and dispositions, and stop at most at `GO_FOR_R3_QUALIFIED_REVIEW`.

No qualified review, AP1, skill selection, Wave 1 artifact, code, Git action, or external effect is authorized by this draft.
