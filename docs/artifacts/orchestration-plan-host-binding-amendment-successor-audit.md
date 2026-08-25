# Level 7 Dev Loop — Codex Host-Binding Successor Amendment Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-ORC-AMD-003` |
| Artifact type | Separate-session, read-only targeted governance audit of one inert local A1 successor candidate; not a product, security, compatibility, release, or deployment audit |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Date | 2026-08-25 |
| Status | **FINAL** |
| Audit mode | `level7-dev-loop:l7-release` Mode B; candidate, code, Git, harness, plugin, skill, manifest, package, marketplace, host, and environment remained read-only; this no-overwrite audit record is the sole authorized write |
| Audit authority | Current original user-role authorization from Anup Pandey for exactly one local A1 artifact-only audit of exact `L7-AMD-ORC-002` 0.1.0 SHA-256 `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74`, creating only this exact no-overwrite path |
| Candidate | [`L7-AMD-ORC-002`](orchestration-plan-host-binding-amendment-successor.md) 0.1.0 |
| Candidate SHA-256 audited | `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74` |
| Candidate-declared risk / maximum activated effect | `R3` authorization-identity and one-use transition binding / one local `A1` two-record proposal |
| Generic Mode B decision | **CONDITIONAL GO**, solely to the mandatory R3 qualified-review gate |
| Candidate-protocol verdict | **GO_FOR_R3_QUALIFIED_REVIEW** |
| Activation state | **INERT / BLOCKED** pending the exact audit-record digest, structurally independent qualified review, fresh-thread AP1, post-AP1 selection and dispatch, and invocation-time preflight |
| Exact invocation token or chip | `NOT_EVALUATED` by design; no `l7-build` token was selected, reconstructed, or dispatched during this audit |
| Audit-record SHA-256 | Computed after the no-overwrite write and reported in the completion handoff; not self-embedded |

## 1. Decision

The exact `L7-AMD-ORC-002` 0.1.0 candidate at SHA-256 `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74` may advance exactly one gate: the structurally independent R3 qualified review required by candidate §5. This audit found no unresolved Blocker, Critical, High, Medium, or Low finding and therefore returns the candidate's maximum permitted favorable result, `GO_FOR_R3_QUALIFIED_REVIEW`.

This is not activation or release approval. It does not authorize either Wave 1 output, establish successor AP1 or sole-writer state, select an invocation token, or issue product, security, compatibility, compliance, release, or deployment assurance. Within the generic `l7-release` vocabulary, `CONDITIONAL GO` means only that the exact candidate and this exact audit record may proceed to qualified review. The successor remains `NOT_AVAILABLE` for dispatch.

| Severity | Count |
|---|---:|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |
| INFO | 4 |

## 2. Scope, repository map, and method

The audit mapped the canonical project before concluding. Immediately before this record was created:

- `/Users/anuppandey/Desktop/level7-dev-loop` was not a Git repository and no ancestor supplied a Git worktree;
- the non-`.cache` tree contained 60 regular files and zero symlinks;
- `docs/artifacts/` contained 23 regular files, including the candidate and historical governance chain;
- this audit path and both permitted Wave 1 output paths were absent in file, directory, symlink, and broken-symlink forms; and
- the repository surface comprised Foundation documents, immutable hash manifests, the pinned harness, 12 prototype skills, plugin metadata, and no Wave 1 contract or specification.

The audit then:

1. Read the complete candidate and governing `l7-release` Mode B instructions.
2. Recomputed the exact candidate, parent, predecessor, audit, review, manifest, plugin, skill, package, and marketplace digests.
3. Verified every entry in `orchestration-plan-candidate.sha256`, `orchestration-inputs.sha256`, and `harness/foundation-inputs.sha256`.
4. Recomputed canonical paths and link state, package inventories, ownership, modes, link counts, expected skill closure, and the predecessor-defined package content-set encoding.
5. Reobserved the CLI, OS, architecture, marketplace entry, installed plugin registration, current host skill catalog, Git state, and output absence.
6. Examined the exact predecessor candidate/audit/review chain and the local session evidence for its AP1, dispatch, preflight, harness run, snapshot, parser failure, and terminal state.
7. Searched repository and local session evidence for earlier or concurrent use of the successor nonce.
8. Audited candidate §§3.2–7 for ordering, one-use state, transport hardening, failure consumption, recovery, expiry, and explicit non-authorization.

No network, external host, provider, mutation trial, Git operation, test refresh, plugin action, selection of `l7-build`, or Wave 1 action was performed.

## 3. Exact candidate, parent, and predecessor closure

| Binding | Recomputed result | Disposition |
|---|---|---|
| Successor identity | Candidate metadata is `L7-AMD-ORC-002` version 0.1.0; file is a regular single-link file owned by `anuppandey`, ends in LF, and hashes to `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74` | **PASS** |
| Parent plan | [`L7-ORC-001` 0.3.1](orchestration-plan.md) hashes to `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` | **PASS** |
| Parent audit | [`L7-AUD-ORC-001`](orchestration-plan-audit.md) hashes to `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` and binds the exact parent | **PASS** |
| Parent approval | [`L7-APR-ORC-001`](orchestration-plan-approval.md) hashes to `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` and authorizes only `L7-FOUNDATION-START-WAVE-1` to the two-record A1 ceiling | **PASS** |
| Parent candidate manifest | `orchestration-plan-candidate.sha256` hashes to `da1fc881dd12f779f55af4745109511ce92a25fcf2c953b893008b08c6c8c109`; all three listed entries verify | **PASS** |
| Parent input freeze | `orchestration-inputs.sha256` hashes to `ef17c49d7ceae115b476c2945fba4149f63094beade4cf8c0ba2d4cf652d2b16`; all five listed entries verify | **PASS** |
| Foundation input freeze | `harness/foundation-inputs.sha256` hashes to `428100ade80a848c2ae5aaa4d1d93876f0c4322cdd56ba2b19a9196593ca31ca`; all 26 protected entries verify | **PASS** |
| Predecessor candidate | [`L7-AMD-ORC-001` 0.1.1](orchestration-plan-host-binding-amendment.md) hashes to `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` | **PASS** |
| Predecessor model audit | [`L7-AUD-ORC-AMD-002`](principal-engineer-release-audit.md) hashes to `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c`, is final, and returned `GO_FOR_R3_QUALIFIED_REVIEW` for exact predecessor bytes | **PASS** |
| Predecessor qualified review | [`L7-REV-ORC-AMD-001`](orchestration-plan-host-binding-amendment-qualified-review.md) hashes to `85187c07a4a44b249e373e75718f93f813401f6090a60a5f191b8e7a0b550e26`, is final, and returned `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW` for the exact predecessor/audit pair | **PASS** |
| Historical pre-remediation audit | [`L7-AUD-ORC-AMD-001`](orchestration-plan-host-binding-amendment-audit.md) remains immutable historical evidence at `80fe801897d3f65a433a9c4b584301ea83457e61c441474b6d0b8bc7f69c9ddb`; it does not audit the successor | **PASS / HISTORICAL** |
| Draft transport | Staged and cached `l7-greenfield` both hash to `6c76a16af74b932733f3a1ea0838fef67fe2c5cbaf6a6aab22777949c8866609`; candidate correctly gives that transport no activation authority | **PASS** |
| Validity | Local date is 2026-08-25 Asia/Kathmandu, before the candidate ceiling of 2026-09-01 23:59:59; no local revocation or superseding successor was found | **PASS AT OBSERVATION** |

The logical action, two-output ceiling, mandatory post-proposal stop, no-Git implementation block, and non-authorization boundaries agree across the parent plan §4.2 and §§11/19/20, parent approval, predecessor, and successor. The successor changes only availability of a future one-use local attempt; it does not rewrite the consumed predecessor record or expand the action.

## 4. Recomputed §3.1 audit-time and preflight bindings

| Required binding | Audit result | Disposition |
|---|---|---|
| Logical action | Exact `L7-FOUNDATION-START-WAVE-1`; matches parent plan and approval | **PASS** |
| Canonical project root | `realpath` is exactly `/Users/anuppandey/Desktop/level7-dev-loop`; `/Users`, account home, `Desktop`, and project components are direct, not symlinks | **PASS** |
| Canonical output parent | `realpath` is exactly `/Users/anuppandey/Desktop/level7-dev-loop/docs/artifacts`; `docs` and `artifacts` are direct directories, not symlinks | **PASS** |
| Permitted activated outputs | Only `docs/artifacts/wave-01-change-contract.md` and `docs/artifacts/wave-01-specification.md`; both are absent in all forms | **PASS AT OBSERVATION** |
| Host tuple | `codex-cli 0.149.1`; macOS 26.5.2 build `25F84`; `arm64` | **PASS** |
| Marketplace identity | Exactly one personal entry named `level7-dev-loop`; local source `./plugins/level7-dev-loop`; `AVAILABLE` / `ON_INSTALL`; category `Developer Tools` | **PASS** |
| Resolved staged source | `/Users/anuppandey/plugins/level7-dev-loop` resolves directly; `codex plugin list` reports `level7-dev-loop@personal`, `installed, enabled`, version 0.1.0 at that path | **PASS** |
| Staged manifest | `/Users/anuppandey/plugins/level7-dev-loop/.codex-plugin/plugin.json` hashes to `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` and parses as the expected plugin/version/interface | **PASS** |
| Effective cached manifest | `/Users/anuppandey/.codex/plugins/cache/personal/level7-dev-loop/0.1.0/.codex-plugin/plugin.json` hashes to the same `202be0…f3f3`; staged and cached bytes compare equal | **PASS** |
| Activated component | The current host-provided audit-session skill catalog exposes exactly one canonical `level7-dev-loop:l7-build` component. No invocation token/chip was selected; a later post-AP1 `/skills` selection remains mandatory | **PASS FOR AUDIT-TIME IDENTITY / TOKEN NOT_EVALUATED** |
| Staged/cached activated skill | Both `skills/l7-build/SKILL.md` files hash to `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` and compare equal | **PASS** |
| Package closure | Each staged/cached package contains exactly 13 regular files: one manifest and the expected 12 `skills/*/SKILL.md` files; zero symlinks, other non-directory nodes, extras, missing files, or files with link count other than 1 | **PASS** |
| Historical skill closure | All 12 staged and all 12 cached skill digests match their protected `harness/foundation-inputs.sha256` entries | **PASS** |
| Package ownership and modes | Every package entry is owned by `anuppandey`; no directory or file is group- or world-writable | **PASS** |
| Package content-set digest | Both packages recompute to `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` over the `LC_ALL=C` sorted `<digest><two spaces><relative path><LF>` encoding | **PASS** |
| Historical repository manifest | `.codex-plugin/plugin.json` hashes to `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf`; retained only as historical evidence | **PASS / HISTORICAL** |
| Marketplace observation | `/Users/anuppandey/.agents/plugins/marketplace.json` hashes to `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553` | **PASS AT OBSERVATION** |
| Parent-chain closure | All three nested manifests and every listed entry verify, as detailed in §3 | **PASS** |
| Harness binding | The protected harness closure and pin are exact. The predecessor invocation's local session evidence records `make verify` exit 0 on pinned Go 1.26.7 and reproducible binary SHA-256 `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f`. A new pass is temporally required immediately before any successor attempt; it cannot be inherited or pre-satisfied by this audit | **PASS AS FAIL-CLOSED PROTOCOL / FUTURE RUN REQUIRED** |
| Git and outputs | `git rev-parse --show-toplevel` reports no repository; both Wave 1 outputs remain absent in all forms | **PASS AT OBSERVATION** |
| Predecessor terminal state | Local session evidence independently supports nonce `L7-AMD-ORC-001-20260825-02` as dispatched and terminally `CONSUMED`; both outputs remain absent; no replay/reset evidence was found | **PASS** |
| Fresh nonce | Repository search finds `L7-AMD-ORC-002-20260825-01` only in the successor candidate. Session search finds one original draft authorization and zero user-role messages combining that nonce with `l7-build`; no earlier dispatch or concurrent claim is visible | **PASS WITH DISCLOSED LOCAL-OBSERVABILITY LIMIT** |

No mismatch was repaired or normalized. Every mutable observation must match again at the later invocation preflight.

## 5. Predecessor consumed-attempt evidence

The predecessor terminal account is corroborated by the local session record `/Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T09-40-41-01a0370f-3ee2-7842-bcc0-83d91a141357.jsonl`, not merely repeated from the successor candidate:

| Event | Reproduced evidence |
|---|---|
| Current-session AP1 | Original user-role message at 2026-08-25T03:55:58.971Z binds predecessor 0.1.1, its review chain, nonce, action, target, scope, host/source tuple, A1 ceiling, validity, state, and sole-writer condition |
| Fresh dispatch | Original user-role message at 2026-08-25T03:59:38.332Z begins with `$level7-dev-loop:l7-build` and restates the predecessor tuple after AP1 |
| Bound preflight | Subsequent read-only checks recomputed hashes, package/runtime state, Git/output absence, and process observations |
| Harness | `make verify` completed at 2026-08-25T04:02:25.699Z with exit 0; policy, import, format/vet/type, proving test, and reproducibility stages passed using pinned Go 1.26.7 |
| Frozen workspace | Snapshot output at 2026-08-25T04:02:53.597Z records 85 entries and SHA-256 `c157fac631a954e90301b3434c35e7f943e48f1494e824d99e10ffa7c2b238d7` |
| Render preparation | Both complete candidate bodies were rendered into session memory before the attempted final wrapper |
| Terminal failure | The tool call at 2026-08-25T04:10:00.943Z contained one intended nested `tools.apply_patch` call for the two Wave 1 paths, but the JavaScript wrapper failed parsing with `SyntaxError: Missing } in template expression`; parse failure occurred before evaluation, so the nested patch primitive did not execute |
| Terminal recheck | Output at 2026-08-25T04:10:43.615Z reports both Wave 1 paths `ABSENT`, Git `ABSENT`, `snapshot_unchanged=true`, `attempt_state=CONSUMED`, and `retry_same_tuple=PROHIBITED` |

Under predecessor §4, submission of the valid token-prefixed message had already moved the tuple to `IN_USE`. The later parser failure is a terminal failure and therefore consumes the tuple even though it produced no repository output. Because neither file exists, the correct terminal condition is `BLOCKED / CONSUMED`, not `RECOVERY_REQUIRED`. The successor preserves rather than resets that state.

## 6. Fresh-chain, transport, state, and fail-closed review

| Control | Assessment | Disposition |
|---|---|---|
| Fresh audit binding | Candidate requires `L7-AUD-ORC-AMD-003` to bind these exact bytes and issue at most `GO_FOR_R3_QUALIFIED_REVIEW`; this record satisfies the ID, scope, and ceiling, with its digest necessarily supplied post-write | **PASS SUBJECT TO POST-WRITE HASH** |
| Independent qualified review | A new `L7-REV-ORC-AMD-002` must bind both exact candidate and exact audit, identify a qualified human/domain reviewer structurally independent of the successor author and every remediator, disclose conflicts, and issue at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW` | **REQUIRED / NOT YET SATISFIED** |
| Fresh AP1 ordering | Only after both reviews may the accountable owner open a fresh thread and issue a complete original AP1 before selection | **PASS AS FAIL-CLOSED RULE / NOT YET SATISFIED** |
| Selection and dispatch | Only a post-AP1 fresh `/skills` result uniquely resolving to `level7-dev-loop:l7-build`, followed by one original token-prefixed owner message in the same thread, can dispatch | **PASS AS FAIL-CLOSED RULE / NOT_EVALUATED** |
| State machine | `NOT_AVAILABLE → UNUSED → IN_USE → CONSUMED` is ordered consistently: reviews plus AP1 make the tuple eligible; dispatch alone enters `IN_USE`; every terminal outcome consumes; no state returns to `UNUSED` | **PASS** |
| Sole writer | Candidate requires a fresh owner assertion plus host-visible process/open-file corroboration immediately before the future attempt and treats hidden-writer absence as an assumption; ambiguity blocks and consumes after dispatch | **PASS AS DEFERRED RULE / NOT YET SATISFIED** |
| Root-cause mitigation | §4 separates completed revalidation from payload construction, requires complete in-memory bodies, bans shell/repository/user interpolation inside dynamic JavaScript template literals, validates exactly two add-only targets, uses one inspectable no-overwrite primitive, and makes every post-dispatch preparation or syntax error terminal | **PASS** |
| No-overwrite and partial failure | Both outputs must be new regular single-link files; neither may be overwritten. Zero output after failure is `BLOCKED / CONSUMED`; any incomplete or unverified pair is `RECOVERY_REQUIRED / CONSUMED`; cleanup, completion, and retry are forbidden | **PASS** |
| Snapshot confinement | Complete before/after non-cache snapshots permit only the exact two new files and preserve all prior regular-file, symlink, directory type/mode, and link evidence, subject to disclosed same-user TOCTOU and derived directory metadata | **PASS WITH RESIDUAL RISK** |
| Expiry and invalidation | Earliest terminal outcome, partial/unrelated delta, 2026-09-01 23:59:59 Asia/Kathmandu, any identity/binding mismatch, predecessor replay/change, or owner revocation/supersession invalidates or consumes authority | **PASS** |
| Explicit non-authorization | Draft, hash, audit, review, or approval cannot invoke Wave 1; Git, code, design, implementation, package/host/environment mutation, external effects, predecessor reuse, and broader assurance claims remain prohibited | **PASS** |

The controls directly address the observed parser-failure class while accurately declining atomicity, confinement, and cryptographic replay claims. No control silently treats the audit as activation.

## 7. Findings and dispositions

### AUD-SUC-001 — INFO — Predecessor terminal evidence is locally mutable

The exact local session record corroborates the AP1, dispatch, harness, snapshot, parser failure, zero-output result, and consumed state, but it is same-user mutable evidence without a signed ledger, Git history, trusted counter, or external attestation.

**Disposition:** Retain as disclosed residual risk. Candidate §§1, 3.1, 5, 6, and 9 preserve the predecessor as consumed, require fresh successor identities, and invalidate on replay or evidence drift. This limitation does not block advancement to qualified review but cannot support broader assurance.

### AUD-SUC-002 — INFO — Current audit does not pre-satisfy the future harness gate

The current no-other-file/environment authorization does not permit refreshing `.cache`, which `make verify` necessarily uses. This audit therefore does not claim a fresh current harness execution. It independently verifies the protected harness/input closure and corroborates the predecessor's exact successful run from session evidence.

**Disposition:** Correctly defer execution. Candidate §§3.1 and 5 require a new pinned `make verify` pass immediately before the future activated attempt. Failure or inability to run it must block; the prior result and this audit cannot be inherited as invocation evidence.

### AUD-SUC-003 — INFO — Transport and confinement remain governance-only

The new payload checks materially address the observed interpolation/parser class, but `apply_patch`, process checks, snapshots, a wall clock, and model-managed state do not form an atomic two-file transaction, OS sandbox, trusted counter, or cryptographic non-replay mechanism. A same-user TOCTOU, hidden writer, compromised tool, or partial primitive can still defeat the attempt.

**Disposition:** Retain and escalate on observation. Candidate §§4–6 and 9 accurately disclose this limit, require fail-closed consumption, and cap the effect at one local A1 proposal. Any observed ambiguity or partial pair blocks or enters recovery; it cannot be repaired under the same tuple.

### AUD-SUC-004 — INFO — Nonce uniqueness and sole-writer absence are bounded observations

Repository and session searches show one successor draft authorization and no successor `l7-build` dispatch, but they cannot prove absence of deleted evidence or hidden/concurrent same-user activity. Current sole-writer eligibility is intentionally not established by this audit.

**Disposition:** Retain the successor state as `NOT_AVAILABLE`. The qualified review, fresh-thread AP1, post-AP1 selection, original dispatch, nonce recheck, process/open-file checks, and immediate pre-write revalidation must all occur later. Any uncertainty is fail-closed.

No unresolved BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW finding remains within this audit's scope.

## 8. Commands and test evidence

Representative read-only commands and observations were:

```text
pwd
rg --files -g '!.cache/**'
rg -n / rg -l --fixed-strings ...
git rev-parse --show-toplevel
shasum -a 256 ...
shasum -a 256 -c docs/artifacts/orchestration-plan-candidate.sha256
shasum -a 256 -c docs/artifacts/orchestration-inputs.sha256
shasum -a 256 -c harness/foundation-inputs.sha256
realpath ...
find ...
stat -f ...
cmp -s ...
jq ...
codex --version
codex plugin list
sw_vers
uname -m
date / TZ=Asia/Kathmandu date
```

The package content-set digests were recomputed by hashing each regular file, emitting `<SHA-256><two spaces><root-relative path><LF>` in `LC_ALL=C` path order, and hashing that stream. No temporary evidence file was created.

Current-session `make verify` was **NOT RUN** because it writes derived `.cache` state and the audit authority forbids modifying any other file or environment. The exact predecessor session's 8.6-second successful run is cited only as historical/terminal evidence; it is not relabeled as a current or future successor pass. Hosted CI, cross-host tests, network tests, and product tests were also **NOT RUN / NOT APPLICABLE** to this inert artifact-only audit.

## 9. Assurance boundary and exactly one next gate

This audit establishes only that the exact successor candidate, its immutable parent/predecessor chain, current local audit-time bindings, consumed-attempt evidence, and fail-closed protocol support progression to the mandatory R3 qualified-review gate. It does not establish activation readiness, current sole-writer status, product behavior, security qualification, compatibility, release readiness, deployment readiness, portability, or external independence.

The only next gate is a separately authorized, no-overwrite, structurally independent qualified-review record `L7-REV-ORC-AMD-002` that binds:

1. exact candidate `L7-AMD-ORC-002` 0.1.0 SHA-256 `85f5295ff86e325d333e4c4f8ec2faca3fc78196fc48b1a43ef0a2940534ba74`; and
2. exact audit `L7-AUD-ORC-AMD-003` plus the post-write SHA-256 of this record.

That reviewer may return at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`. This audit does not authorize creation of that record. Until it exists and every later gate closes in order, the successor remains inert and both Wave 1 outputs remain blocked.
