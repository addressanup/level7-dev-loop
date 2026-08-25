# Level 7 Dev Loop — Codex Host-Binding Amendment

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-ORC-001` |
| Artifact type | Narrow successor amendment to the Foundation Step 6 orchestration plan |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Version | 0.1.1 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT pending fresh post-write model audit, structurally independent qualified review, and exact current-session owner approval** |
| Parent plan | [`L7-ORC-001`](orchestration-plan.md) 0.3.1 |
| Parent-plan SHA-256 | `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| Parent candidate manifest | [`orchestration-plan-candidate.sha256`](orchestration-plan-candidate.sha256) SHA-256 `da1fc881dd12f779f55af4745109511ce92a25fcf2c953b893008b08c6c8c109` |
| Parent audit | [`L7-AUD-ORC-001`](orchestration-plan-audit.md) SHA-256 `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` |
| Parent approval | [`L7-APR-ORC-001`](orchestration-plan-approval.md) SHA-256 `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` |
| Transitive input freeze | [`orchestration-inputs.sha256`](orchestration-inputs.sha256) SHA-256 `ef17c49d7ceae115b476c2945fba4149f63094beade4cf8c0ba2d4cf652d2b16` |
| Draft authority | Owner message on 2026-08-25: “I approve l7-greenfield for the narrow L7-ORC-001 host-manifest binding amendment only; no Wave 1 artifacts or product code.” |
| Pre-remediation candidate | This artifact version 0.1.0, SHA-256 `6eaac34d871a2a9fc3e92e46730673d74084d1752e8485182ba558e55fd25f6b` |
| Pre-remediation audit | [`L7-AUD-ORC-AMD-001`](orchestration-plan-host-binding-amendment-audit.md) SHA-256 `80fe801897d3f65a433a9c4b584301ea83457e61c441474b6d0b8bc7f69c9ddb`; `NO_GO` |
| Remediation authority | Owner message on 2026-08-25 approving `l7-release` remediation for exact `L7-AUD-ORC-AMD-001` SHA-256 `80fe801897d3f65a433a9c4b584301ea83457e61c441474b6d0b8bc7f69c9ddb` and `AUD-HB-001` through `AUD-HB-003` only, editing only this artifact with no Wave 1 artifacts or product code |
| Remediation scope | `AUD-HB-001` current-session AP1; `AUD-HB-002` R3 qualified review; `AUD-HB-003` audit-time versus invocation-time token phasing |
| Drafting effect / activated action ceiling | A1 local governance record / A1 local Wave 1 contract-and-specification records only |
| Amendment risk | `R3`: this changes an authorization identity binding, even though its maximum activated effect remains A1 |
| Assurance ceiling | A separate-context model audit may issue at most `GO_FOR_R3_QUALIFIED_REVIEW`; a conforming qualified review may issue at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`; neither is a product, security, compatibility, release, or deployment verdict |
| Nonce | `L7-AMD-ORC-001-20260825-02` |

## 1. Reason for the amendment

The approved `L7-ORC-001` §4.2 binds Codex invocation to the historical repository manifest frozen in [`harness/foundation-inputs.sha256`](../../harness/foundation-inputs.sha256). That manifest has SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf` and version `1.0.0`. The current Codex ingestion validator rejects it because `interface.longDescription`, `interface.developerName`, and `interface.capabilities` are absent.

Under a separate owner-approved local bootstrap repair, a development package was staged and installed with only these manifest differences:

- version `1.0.0` became the truthful development version `0.1.0`; and
- the three required interface fields were added, with the long description explicitly denying controlled-execution, security-qualification, and release-assurance claims.

The staged and effective cached manifests both have SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3`. All 12 staged and cached skill files remain byte-identical to their historical entries in `harness/foundation-inputs.sha256`; in particular, `l7-build` remains `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f`.

A separate Codex thread correctly rejected the first Wave 1 attempt because the installed and frozen manifest digests differed. It wrote nothing. This amendment closes only that bootstrap identity mismatch. It does not weaken the fail-closed rule or retroactively authorize the rejected attempt.

Version 0.1.1 changes only the inert activation protocol to remediate `AUD-HB-001` through `AUD-HB-003`. The pre-remediation `NO_GO` audit is immutable evidence for version 0.1.0; it does not audit or approve these successor bytes.

## 2. Narrow normative amendment

This section has **no effect** until the activation gate in §4 is satisfied.

Once activated, and for the **Codex CLI route only**, the phrase “the host-specific manifest recorded in the approved transitive input freeze” in `L7-ORC-001` §4.2 and the corresponding host-manifest portion of `L7-APR-ORC-001`'s bound source identity SHALL mean the effective installed development-manifest binding defined in §3 of this amendment. The historical repository manifest and every existing approved plan, manifest, audit, and approval remain immutable evidence. They are not edited, deleted, re-hashed, or represented as the installed manifest.

If this amendment conflicts with `L7-ORC-001`, this amendment governs only that one Codex runtime-manifest referent. Every other `L7-ORC-001` clause remains unchanged, including:

- plugin and skill discovery, exact-token preservation, and fail-closed collision behavior;
- the Codex IDE exclusion and the unchanged Claude Code route;
- Wave 1-only scope and the contract/specification, design, and implementation approval stops;
- the no-Git implementation block;
- all network, provider, dependency, root, publication, deployment, cleanup, and external-effect prohibitions; and
- the ban on automatic continuation.

The manifest capability label `Write` is descriptive UI metadata. It grants no authority, raises no effect ceiling, and cannot bypass any repository or owner gate.

## 3. Exact local runtime binding by phase

### 3.1 Audit-time and preflight bindings

The post-write model audit MUST independently recompute every binding in §3.1, and every binding MUST match again during invocation preflight. These checks establish the canonical component and discovery route but do not establish the invocation-token binding in §3.2. Unknown, missing, ambiguous, stale, or partially matching state is `BLOCKED`.

| Binding | Required value |
|---|---|
| Logical action | `L7-FOUNDATION-START-WAVE-1` |
| Canonical project root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Canonical output parent | `/Users/anuppandey/Desktop/level7-dev-loop/docs/artifacts`; the project root and this directory must resolve exactly to these paths without a symlink from the project root downward |
| Permitted outputs | Only `docs/artifacts/wave-01-change-contract.md` and `docs/artifacts/wave-01-specification.md` |
| Host observation | `codex-cli 0.149.1` on macOS 26.5.2 build `25F84`, `arm64` |
| Marketplace identity | Normalized entry `level7-dev-loop@personal`, source `local`, source path `./plugins/level7-dev-loop`, policy `AVAILABLE` / `ON_INSTALL`, category `Developer Tools` |
| Resolved staged source | `/Users/anuppandey/plugins/level7-dev-loop`; `codex plugin list` must report `installed, enabled`, version `0.1.0` |
| Staged manifest | `/Users/anuppandey/plugins/level7-dev-loop/.codex-plugin/plugin.json`; SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Effective cached manifest | `/Users/anuppandey/.codex/plugins/cache/personal/level7-dev-loop/0.1.0/.codex-plugin/plugin.json`; SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Canonical component and discovery route | `/skills` discovery must expose exactly one result whose source resolves to canonical component `level7-dev-loop:l7-build`; its staged and cached skill bytes must match the digest below. The post-write audit verifies canonical identity, unique discoverability, and the host-native selection mechanism. It does not attest an invocation token or chip; anything exposed during audit is non-authorizing and cannot be reused. |
| Staged and cached build skill | Both `skills/l7-build/SKILL.md` files; SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package closure | Staged and cached packages each contain exactly 13 regular files—one manifest plus the 12 expected `skills/*/SKILL.md` files—with no symlink, hardlink (`nlink` must equal 1), or extra file; every skill digest matches its corresponding historical `harness/foundation-inputs.sha256` entry |
| Package content-set digest | Both packages produce SHA-256 `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` over the `LC_ALL=C` sorted regular-file list encoded as `<file SHA-256><two spaces><root-relative path><LF>` |
| Ownership and mode | Staged and cached package entries are owned by local user `anuppandey`; no package directory or file may be group- or world-writable |
| Historical manifest | Repository `.codex-plugin/plugin.json`; SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf`; historical evidence only |
| Marketplace observation | `/Users/anuppandey/.agents/plugins/marketplace.json` currently has SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553`; this whole-file digest is evidence, not a validity anchor, so an unrelated personal-plugin entry cannot silently redefine the normalized binding above |

### 3.2 Invocation-time dynamic bindings

The exact invocation token or skill chip is deliberately `NOT_EVALUATED` by the pre-approval post-write model audit. Every row below becomes mandatory only in the fresh invocation thread and MUST be checked before dispatch or, where stated, immediately after dispatch and before the first write.

| Dynamic binding | Required value |
|---|---|
| Activation evidence chain | Exact final amendment SHA-256; exact post-remediation model-audit artifact ID and SHA-256 bound to that candidate; exact qualified-review-record ID and SHA-256 bound to both; and nonce `L7-AMD-ORC-001-20260825-02`. Every record must be final, mutually consistent, unexpired, unrevoked, and unsuperseded. |
| Qualified reviewer | The qualified-review record must identify the named human/domain reviewer and role, evidence-supported authorization/governance qualification for this Codex host/plugin binding, structural independence from the amendment author and every remediator, conflicts, evidence and methods reviewed, every finding and disposition, residual risk, scoped decision, and validity window. Missing, stale, `UNVERIFIED`, conflicting, or mismatched evidence is `BLOCKED`. |
| Current-session AP1 | After the model and qualified-review gates close, the accountable owner must issue an exact approval in the same fresh Codex thread that will perform discovery and dispatch. It must bind the complete activation evidence chain, logical action, canonical target and two-file scope, exact §3.1 host/runtime/source identity, A1 effect ceiling, validity window, unused one-attempt state, and current sole-writer condition. Approval from another thread, an artifact, quotation, summary, inherited context, assistant message, or delegate is `AP0` provenance only. |
| Exact selected token | Only after that same-thread AP1, a fresh `/skills` selection must return exactly one result resolving to `level7-dev-loop:l7-build` and insert or display its exact `$…` token or host-native skill chip. The operator must verify the resolution against §3.1 and preserve the inserted identity exactly; it cannot be typed, edited, reconstructed, stripped, extended, guessed, or replaced after dispatch. |
| Current-session reconfirmation and dispatch | In that same thread, the accountable owner must submit one original user-role message beginning with the unmodified inserted token and restating every literal value in the AP1 authorization tuple, including the exact token/component resolution. Placeholders, abbreviations such as “approved §4.2,” or an uninspectable original message are invalid. This message alone reconfirms AP1 immediately before mutation and dispatches the attempt. |
| One-attempt state | The exact tuple must be `UNUSED` immediately before submission, become `IN_USE` exactly once through that dispatch, and have no earlier or concurrent claim. Any ambiguity, duplicate, block, failure, cancellation, recovery, or terminal success makes it `CONSUMED`; it never returns to `UNUSED`. |
| Current sole-writer state | The owner confirmation and host-visible check must both support that no other mutation-capable thread or process will write the canonical project during the bounded attempt. Hidden-writer absence remains an explicit governance assumption, not a mechanically proven fact. |

An audit-time, pre-approval, earlier-thread, stale, missing, duplicate, truncated, uninspectable, or mismatching selection is `BLOCKED`. Selection itself does not consume the attempt; submission of the exact confirmation-and-dispatch message does.

The absolute paths make this a one-host local development amendment. It is not portable compatibility evidence and cannot support a Codex, Claude, OS, architecture, installation, or distribution claim.

## 4. Activation and one-use gate

This amendment becomes active only when **all** of the following occur in order:

1. Its final bytes receive a separate post-write, read-only model audit that independently recomputes the amendment digest, all parent-chain digests, and every audit-time binding in §3.1, and audits the §3.2 discovery, selection, authority, state-transition, and fail-closed rules. The exact §3.2 invocation token or chip is intentionally `NOT_EVALUATED` at this phase; no audit-time or pre-approval token may become invocation evidence.
2. That audit returns only `GO_FOR_R3_QUALIFIED_REVIEW`, with no unresolved Blocker, Critical, High, or Medium finding, and binds the exact amendment SHA-256. A separate-context model audit is not the qualified review and cannot issue a product, security, release, or deployment `PASS` or `GO`.
3. The exact amendment and exact model-audit record then receive a read-only review by a named qualified human/domain reviewer who is structurally independent of the amendment author and every remediator. The durable review record MUST bind both SHA-256 digests and record the reviewer's identity and named role; evidence-supported authorization/governance qualification for this Codex host/plugin binding; independence and conflict declaration; evidence and methods reviewed; every finding and disposition; residual risk; scoped decision; and validity/expiry. Missing, stale, `UNVERIFIED`, conflicting, or mismatched evidence, or any unresolved Blocker, Critical, High, or Medium finding, is `BLOCKED`. A conforming decision may issue at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`, never product, security, release, or deployment assurance.
4. Only after both review gates close does the accountable owner open a fresh Codex thread. Before any invocation-token selection, that owner issues the exact current-session AP1 defined in §3.2. The approval must bind the final amendment SHA-256, model-audit ID and SHA-256, qualified-review-record ID and SHA-256, reviewer identity and role, nonce, logical action, canonical target and two-file scope, exact host/runtime/source identity, A1 effect ceiling, validity window, unused one-attempt state, and current sole-writer condition. The present draft/remediation authority cannot approve bytes or review evidence that did not yet exist, and an approval issued in another thread is `AP0` provenance only.
5. After that same-thread AP1, `/skills` freshly discovers the unique §3.1 component and inserts its exact token. The accountable owner then sends the §3.2 original user-role confirmation-and-dispatch message in the same thread. The previously rejected invocation and every audit-time, pre-approval, earlier-thread, guessed, reconstructed, or edited selection are non-replayable and cannot be resumed or reinterpreted as approved.
6. Immediately after dispatch and before the first write, the coordinator revalidates the complete activation evidence chain, current-session AP1 message, nonce, every §3.1 binding, the dispatched §3.2 token-to-component resolution without making a replacement selection, validity/revocation/supersession state, and the one-attempt transition from `UNUSED` to `IN_USE`. It also revalidates absence of Git, absence of both output files including broken symlinks, and the green Foundation harness. Because Git/worktree isolation is absent, it rechecks host-visible process/session state, records the owner's current sole-writer confirmation as an assumption, and captures an in-memory workspace snapshot outside `.cache/`. The snapshot records each regular file's root-relative path, kind, mode, link count, size, and SHA-256; each symlink's path, kind, mode, and target; and each directory's path, kind, and mode. Hidden-writer absence is not mechanically provable. Unknown or conflicting state stops and consumes the attempt.
7. The coordinator renders and validates the complete pair before the first write. Immediately before each no-overwrite file creation it revalidates the exact activation/AP1 tuple, expiry, revocation, attempt ownership, and host-visible sole-writer state. It creates both files as one bounded no-overwrite change and then repeats the snapshot. Exactly two new regular, single-link files at the permitted canonical paths may differ; every previously recorded regular-file and symlink field and every directory's type/mode must match. Only derived directory metadata necessarily caused by those two additions may be ignored. A mismatch before the first write is `BLOCKED`; a sole first file is permitted only as transient coordinator-owned state during that same uninterrupted change, and any mismatch, interruption, conflict, foreign mutation, or failure after it exists is `RECOVERY_REQUIRED`. Either result consumes the attempt.

The exact approval tuple has governance states `UNUSED → IN_USE → CONSUMED`. The §3.2 user-role submission is the sole permitted `UNUSED → IN_USE` transition. Any duplicate or concurrent submission, uncertain prior state, terminal success, block, failure, cancellation, or recovery result transitions it to `CONSUMED`; there is no retry or transition back to `UNUSED` under the same approval.

For this `R3` candidate, no authoring or remediating agent may serve as the qualified reviewer, close its own qualified-review findings, approve activation, or issue an activation/release `PASS`, `GO`, or `CONDITIONAL_GO`. `R3` remains the governing floor unless a separate reassessment presents new risk-dimension evidence and an accountable approver accepts it. An author or remediator cannot accept its own downgrade. Any accepted reclassification is a material successor that invalidates the prior model audit, qualified review, and approval and requires a fresh exact-digest chain.

After activation, authority expires or is consumed at the earliest of:

- the terminal result of the one fresh invocation attempt, including successful creation of both proposed records and the mandatory owner-approval stop;
- any partial output, which becomes `RECOVERY_REQUIRED` and cannot be retried under the same approval;
- 2026-09-01 23:59:59 Asia/Kathmandu;
- any amendment, model-audit, qualified-review, reviewer, nonce, AP1, token, staged, cached, skill, package-inventory, marketplace-binding, CLI, OS, architecture, project-root, parent-chain, scope, risk, effect, or authority mismatch; or
- owner revocation or supersession.

## 5. Fail-closed and recovery rules

The coordinator MUST stop before beginning the bounded proposal if either §3.1 discovery or §3.2 invocation-time selection is missing, duplicate, ambiguous, truncated, uninspectable, or mismatched; if the dispatched identity was guessed, typed, edited, reconstructed, selected before current-session AP1, selected in another thread, or replaced after dispatch; if the invocation-thread approval exists only as persisted `AP0`, quoted text, assistant output, summary, inherited context, delegation, or another thread's message; if the fresh selection, owner reconfirmation, and dispatch are not in the same thread; if the exact original user-role message is no longer inspectable; if the one-attempt transition cannot be uniquely attributed to the current dispatch; if the model-audit or qualified-review gate is absent, stale, mismatched, unqualified, non-independent, or unresolved; if the stage and cache disagree; if a package file has `nlink != 1`; if the cache or source escapes the bound paths through a symlink; if an extra or missing package file exists; if ownership or write permissions violate §3.1; if the installed plugin is disabled; if another mutation-capable writer exists; if the project root or output parent does not resolve to the exact canonical path in §3.1; if either output exists at preflight as any file type, symlink, or broken symlink; if the project root, host surface, or component belongs to another path/host; or if any digest, nonce, review, approval, token, state, or environment field is stale. After preflight, each output must be a new regular file with `nlink == 1` at its canonical path and may exist only when created by the same still-running bounded action described in §4; any foreign, stale, overwritten, linked, escaped, or abandoned output, or any unrelated workspace-snapshot delta, is `RECOVERY_REQUIRED`.

It MUST NOT repair a mismatch by editing the marketplace, source, cache, repository manifest, skill, approval, environment, or validator; by adding a cachebuster; by reinstalling; by switching host surfaces; or by weakening a check. Any such repair needs new explicit authority and a successor binding.

Cancellation preserves all existing user and repository state. No destructive reset, cache deletion, plugin removal, cleanup, or automatic retry is authorized. The correct result is a decision-first `BLOCKED` or `RECOVERY_REQUIRED` status with exactly one next action.

These controls are currently enforced by model instructions, local tool observations, and human approval in a same-user mutable environment. They are not an OS sandbox, cryptographic transaction, trusted monotonic counter, or hardened one-use grant. The nonce correlates evidence but cannot by itself prevent replay if an actor deletes or rewrites editable outputs or records; the one-attempt rule is therefore a governance invariant, not a cryptographic guarantee.

## 6. Explicit non-authorization

Drafting, auditing, reviewing, or approving this amendment does not itself invoke Wave 1. Even after activation, it authorizes only a fresh host-native request to propose the two records named in §3.1 and then stop.

It never authorizes design, Git initialization/import, branches, commits, merges, product or harness code, prompts, semantic workflows, skill or manifest changes, generated packages, dependencies, host/provider/network trials, root operations, protected infrastructure, controlled mutation, publication, deployment, exposure, release, cleanup outside the repository, self-modification, or autonomous behavior.

## 7. Evidence at drafting time

| Check | Result |
|---|---|
| Existing approved parent files | Byte-identical to the digests in the metadata table |
| Historical Foundation input closure | `PASS`; `harness/foundation-inputs.sha256` remains unchanged |
| Staged and cached manifest validation | `PASS` with the local plugin validator |
| Stage/cache manifest equality | `PASS`; both SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Package inventory and symlinks | `PASS`; 13 regular files in each package, no symlinks or extras |
| Package hardlinks and permissions | `PASS`; every package file has `nlink == 1`, is owned by `anuppandey`, and is not group- or world-writable |
| Project/output canonicalization | `PASS`; project root and `docs/artifacts` resolve to the exact paths in §3.1 and are not symlinks |
| Skill equality | `PASS`; all 12 staged and cached skill digests match the protected source entries |
| Plugin registration | `PASS`; `level7-dev-loop@personal` reports `installed, enabled`, version `0.1.0` |
| Previous invocation | Correctly `BLOCKED` on the manifest mismatch; zero files written |
| Pre-remediation model audit | `NO_GO` for version 0.1.0 at the exact digest in the metadata table; its three material findings are the sole scope of version 0.1.1 and it does not audit these bytes |
| Finding-level remediation checks | `PASS`; three separate-context read-only checks found `AUD-HB-001`, `AUD-HB-002`, and `AUD-HB-003` textually remediated; these checks are regression evidence, not the required fresh Mode B audit or qualified review |
| Exact post-approval `/skills` token or chip | `NOT_EVALUATED` by design; no valid invocation binding exists until the §4 review and same-thread AP1 gates close |
| Structurally independent qualified reviewer | `NOT_ESTABLISHED`; activation remains `BLOCKED` unless and until the exact §4 record exists |
| Git | Absent |
| Wave 1 artifacts and product code | Absent |
| `make verify` | `PASS` after the version 0.1.1 remediation |
| Hosted CI / cross-host compatibility | `NOT_RUN` / `UNPROVED` |

These drafting and remediation observations are not the required fresh post-write model audit or qualified review and cannot activate this amendment.

## 8. Compact assurance case

| Element | Statement |
|---|---|
| Claim | Activating this amendment can be acceptably reviewed for one local A1 contract/specification proposal; it cannot establish product security, compatibility, release, or deployment assurance. |
| Argument | Immutable parent evidence plus an exact stage/cache/component binding prevents the known manifest substitution; phase-separated token selection prevents pre-approval token reuse; qualified review and same-thread AP1 preserve R3 governance; canonical no-overwrite paths, one writer, before/after snapshots, and one-attempt consumption confine the authorized effect; every mismatch fails closed. |
| Evidence | Parent and runtime digests in §§1–3; validator and package-closure results; protected skill equality; canonical-path, ownership, link-count, and permission checks; green `make verify`; the prior correct block; the pre-remediation `NO_GO` and its dispositions; and the still-required fresh exact-digest model audit and qualified-review record. |
| Assumptions | Local hashing and `lstat` observations are accurate at check time; the qualified reviewer is competent and independent as recorded; the accountable owner and sole writer follow the protocol; Codex resolves the displayed component to the observed effective cache; the same-user environment is not already maliciously subverted. |
| Defeaters | Same-user TOCTOU or cache replacement after validation; hidden concurrent writers; path, token, or marketplace aliasing; clock rollback; deletion of editable evidence followed by replay; compromised Codex/tooling; or inability to perform one bounded no-overwrite operation with partial-write detection. Any observed defeater blocks or yields `RECOVERY_REQUIRED`. |
| Residual risk | The action remains governance-controlled in a mutable local environment. Snapshot comparison and one-attempt semantics reduce accidental/confused-deputy risk but do not provide cryptographic non-replay, OS-level confinement, or qualified security assurance. |
| Required reviewer and approver | A separate-context read-only model audit may issue only `GO_FOR_R3_QUALIFIED_REVIEW`; it does not satisfy qualified review. A named qualified human/domain reviewer, structurally independent of the candidate author and every remediator, must then issue the digest-bound scoped record required by §4 and may issue at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`. Only after both reviews may the accountable product owner issue same-thread AP1 for the exact local A1 candidate. An author/remediator cannot fill the qualified-review or approval role, close its own findings, self-downgrade, or self-issue activation/release `PASS` or `GO`; no role may inflate these results into product, security, release, or deployment assurance. |

## 9. Next gate

The only permitted next action after this proposed artifact is a separate, read-only model audit of its exact final digest, the §3.1 bindings, and the §3.2 mechanism and fail-closed rules. That audit MUST record the exact post-approval token or chip as `NOT_EVALUATED` by design. A clean model audit must stop for the structurally independent qualified review required by §4; a supporting qualified-review record must stop for a fresh-thread, exact current-session owner decision and post-approval token selection. Until the model audit, qualified review, and same-thread AP1/dispatch gates close in order, Wave 1 remains `BLOCKED`.
