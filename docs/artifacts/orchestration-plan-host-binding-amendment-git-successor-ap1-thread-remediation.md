# Level 7 Dev Loop — Same-Thread AP1 Remediation Successor

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-ORC-005` |
| Artifact type | Inert finding-specific successor to `L7-AMD-ORC-004` after the consumed cross-thread activation attempt |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT / NOT_AVAILABLE pending fresh independent Mode B audit, qualified review, exact same-thread AP1, post-AP1 selection, and dispatch** |
| Immediate predecessor | `L7-AMD-ORC-004` 0.1.0 at `orchestration-plan-host-binding-amendment-git-successor-remediation.md`, SHA-256 `c1901549b66405a8d837746fcc83f56d04097108b24ba3e02301c484634f6f25` |
| Predecessor candidate commit | `b45aa6c66f316132614eee7b5c3f7a369a3d1fcc`; tree `5dd5da46b3ab4dc9ec6c7c71025b6955d6ce3238`; parent `408decb636add15bac42e2eeeed5582d21c3d0f7` |
| Predecessor audit | `L7-AUD-ORC-AMD-005` 0.1.0, SHA-256 `4c04e6e966f9270f5061937657fac0ce304de943cfcb2a0cbde6aa9ce62b5a85`, commit `47649eeeaf8644f1218233a2eb1364f65099cdad` |
| Predecessor qualified review | `L7-REV-ORC-AMD-004` 0.1.0, SHA-256 `f9301c9902ec15e62063ed962347ab640764c7cf31f9a05e7f812ed09b5ec763`, commit `d6a9e0357db308a81139eeecf4fc1a06bb26928a` |
| Source finding | `AUD-ACT-001` — HIGH — attempted dispatch lacked an earlier same-thread AP1; Mode B decision `NO-GO` |
| Source audit evidence | Original assistant Mode B record at JSONL line 238 in session `01a03910-1f16-7e31-871d-a67f9c00cd11`; same-user mutable session evidence, not a Git-bound audit artifact |
| Remediation companion | `release-audit-remediation-aud-act-001.md` |
| Logical action if later activated | `L7-FOUNDATION-START-WAVE-1` |
| Maximum effect if later activated | One local A1 proposal consisting only of the Wave 1 change contract and specification, followed by a mandatory owner-approval stop |
| Fresh proposed nonce | `L7-AMD-ORC-005-20260825-01` |
| Proposed validity ceiling | Earliest invalidation below, including 2026-09-01 23:59:59 Asia/Kathmandu |
| Candidate path | `docs/artifacts/orchestration-plan-host-binding-amendment-git-successor-ap1-thread-remediation.md` |
| Next gate | Fresh-session independent `level7-dev-loop:l7-release` Mode B audit of these exact final bytes, companion bytes, and their containing Git commit |

## 1. Confirmed incident and immutable evidence

`AUD-ACT-001` is reproduced from original event order, not inferred from the failed coordinator's conclusion:

1. Session `01a0382f-1f65-7450-81e6-f37832aaa111` contains the accountable-owner AP1 at JSONL line 9, timestamp `2026-08-25T09:10:10.005Z`.
2. Session `01a038f9-f027-7853-b06c-bdf3e3ebaf25` is a different session. Its first substantive user instruction at JSONL line 9, timestamp `2026-08-25T12:51:55.805Z`, is the token-prefixed attempted dispatch.
3. The dispatch phrase refers to “the exact AP1 above,” but that dispatch session has no earlier original AP1 event.
4. The host-injected `l7-build` skill record follows in the dispatch session at JSONL line 11; it cannot import authority from the AP1 session.
5. No `apply_patch` or other output-creation call occurred. The direct read-only terminal check at dispatch-session JSONL line 70 records both Wave 1 outputs absent.
6. Current filesystem checks and reachable Git history also show both outputs absent.

The earlier AP1 and its admission result are not rewritten or denied. They did not satisfy the predecessor's separate requirement that AP1, post-AP1 selection, and dispatch occur in the same thread. No valid `UNUSED → IN_USE` transition is claimed for the attempted dispatch. Conservative anti-replay treatment still makes its nonce terminal and non-replayable.

All predecessor documents, commits, and session evidence remain unchanged. This successor does not reconstruct missing same-thread authority or reinterpret the failed attempt as resumable.

## 2. Nonce dispositions

| Nonce | Evidence-bounded disposition |
|---|---|
| `L7-AMD-ORC-001-20260825-02` | **BLOCKED / CONSUMED / NON-REPLAYABLE** |
| `L7-AMD-ORC-002-20260825-01` | **PERMANENTLY RETIRED / SUPERSEDED / NON-REPLAYABLE**; no verified pre-selection AP1 or valid transition is reconstructed |
| `L7-AMD-ORC-003-20260825-01` | Never available; superseded and non-replayable |
| `L7-AMD-ORC-004-20260825-01` | **BLOCKED / CONSUMED / NON-REPLAYABLE** after the cross-thread attempted dispatch; both outputs absent; no valid `UNUSED → IN_USE` transition claimed |
| `L7-AMD-ORC-005-20260825-01` | Fresh searched proposal only; **NOT_AVAILABLE** until every new review and authorization gate closes |

No prior nonce, AP1, token, selection, dispatch, admission result, or state is reusable by this successor.

## 3. Inherited immutable bindings

This successor incorporates every non-conflicting scope, source-identity, Git, host/plugin/package, path, preflight, transport-hardening, no-overwrite, snapshot, one-writer, validity, fail-closed, recovery, and non-authorization control from exact `L7-AMD-ORC-004` 0.1.0 at the SHA-256 above. If wording differs, this successor's stricter same-thread rules control. Missing or mismatched inherited evidence is `BLOCKED`.

The remediation admission baseline is:

| Binding | Required value |
|---|---|
| Canonical root | `/Users/anuppandey/Desktop/level7-dev-loop`; no symlink component |
| Admission Git tip | `d6a9e0357db308a81139eeecf4fc1a06bb26928a` |
| Admission tree | `b947196a3ff52650cdde50d83a87b26018850648` |
| Admission parent | `47649eeeaf8644f1218233a2eb1364f65099cdad` |
| Branch/topology | Local `main`; SHA-1; one canonical worktree; zero remotes and replacement refs; non-shallow |
| Remediation commit | One direct child of the admission tip adding only this candidate and its companion remediation record |
| Clean future admission | Exact future audit/review-chain tip, clean index and tracked worktree, and no untracked non-ignored path |
| Permitted outputs | Only `docs/artifacts/wave-01-change-contract.md` and `docs/artifacts/wave-01-specification.md`; both absent before mutation |
| Host/plugin bindings | Exact values inherited from `L7-AMD-ORC-004` §4.2, freshly recomputed at every later audit and invocation |
| Harness | Pinned offline `make verify` passes immediately before a future attempted dispatch; cache effects remain ignored |

Final candidate/companion SHA-256 values, remediation commit identity, tree, parent, and exact two-file delta are computed after finalization and reported in the Mode C completion handoff. They are mandatory inputs to the next Mode B audit.

## 4. New same-thread activation protocol

### 4.1 Review gates

Before AP1, the exact final candidate and companion bytes and their exact remediation commit MUST receive:

1. a fresh, separate, independent Mode B audit, expected new identity `L7-AUD-ORC-AMD-006`, with no unresolved Blocker, Critical, High, or Medium finding and a verdict of at most `GO_FOR_R3_QUALIFIED_REVIEW`; and
2. a later named, evidence-qualified, structurally independent review, expected new identity `L7-REV-ORC-AMD-005`, returning at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW` and receiving its own separately authorized single-record Git binding.

Neither proposed identity is authority. A collision, earlier claim, wrong digest, wrong commit, wrong parent, mixed delta, or missing durable record blocks before AP1.

### 4.2 AP1 session continuity

After both review gates close, the only admissible sequence is:

1. The accountable owner opens one fresh Codex thread at the exact qualified-review-chain Git tip.
2. Before any skill selection, the owner issues an original AP1 as the first substantive user instruction. It binds the complete candidate/audit/review chain, their hashes and commits, nonce `L7-AMD-ORC-005-20260825-01`, logical action, canonical root, exact two-file scope, host/plugin identity, A1 ceiling, validity, output absence, unused state, and sole-writer condition.
3. The coordinator records the current `session_meta.payload.session_id` as `AP1_SESSION_ID`, verifies AP1 admission read-only, and reports it explicitly. The admission response MUST say: **remain in this exact thread; do not open or resume another thread; do not run `l7-next`; use `/skills` here for the next message only**.
4. No intervening skill invocation, “what next” routing pass, copied prompt in another thread, resumed thread, fork, or reconstructed AP1 is admissible. Before AP1 admission, ambiguity blocks while the nonce remains `NOT_AVAILABLE`. After AP1 admission, any such continuity break invalidates the admission and permanently consumes the proposed nonce without authorizing an output.
5. Still in the same thread, the owner uses fresh `/skills` selection resolving uniquely to `level7-dev-loop:l7-build` and sends one original token-prefixed reconfirmation-and-dispatch message.
6. Before accepting any authority from that message, the coordinator obtains the current session ID as `DISPATCH_SESSION_ID` and requires exact byte equality with `AP1_SESSION_ID`.
7. The coordinator inspects the current session transcript and requires the original AP1 user event to precede the token-prefixed dispatch user event. A quoted, summarized, linked, inherited, assistant-authored, or cross-session AP1 is `AP0` and non-authorizing.

The dispatch message MUST bind the exact nonce contiguously and MUST NOT use “above,” “previous,” “earlier,” or another relative reference as a substitute for the inspected same-thread AP1 event.

### 4.3 One-use transition

Before all review and AP1-session checks close, state is `NOT_AVAILABLE`. Only a verified same-thread AP1 admission makes the tuple eligible as `UNUSED`. Only the exact same-thread token-prefixed dispatch may move it to `IN_USE`.

The required sequence is:

`NOT_AVAILABLE → UNUSED → IN_USE → CONSUMED`

Any post-AP1 continuity break or token-prefixed attempted dispatch naming the nonce is conservatively terminal even if malformed or unauthorized. Success, block, mismatch, preparation error, syntax/transport/tool failure, cancellation, ambiguity, conflict, recovery, or partial output consumes it permanently. No transition may be repaired, replayed, resumed, or inferred after the event.

## 5. Bounded output and transport rules

If and only if section 4 closes, the coordinator must freshly reproduce all inherited bindings, run the pinned offline verifier, confirm output absence and clean exact Git state, corroborate visible sole-writer/open-file state, and capture the complete pre-snapshot before mutation.

The only permitted mutation is one direct no-overwrite `apply_patch` call that adds exactly:

1. `docs/artifacts/wave-01-change-contract.md`; and
2. `docs/artifacts/wave-01-specification.md`.

Both complete bodies must be rendered and validated in memory first. No update, delete, move, overwrite, third path, staging, commit, design, or implementation is permitted. Every shell-special-name prohibition and direct-executable rule from the predecessor remains mandatory.

After the patch, Git-native and complete non-`.git`/non-`.cache` snapshot checks must prove that exactly those two new regular single-link files differ. The coordinator then reports both SHA-256 values, terminal `CONSUMED`, residual limitations, and the mandatory owner-approval stop.

## 6. Regression contract for `AUD-ACT-001`

Every future audit and attempted activation MUST produce the following evidence table from original session records:

| Assertion | Required result |
|---|---|
| AP1 event role | `user` |
| Dispatch event role | `user` |
| `AP1_SESSION_ID == DISPATCH_SESSION_ID` | `PASS` before any authority is accepted from dispatch |
| AP1 event order | Strictly earlier than skill selection and dispatch |
| AP1 content | Exact complete bindings; no relative-reference substitution |
| Dispatch content | Exact contiguous nonce and logical action; original host token |
| Cross-thread or missing event | `BLOCKED`; attempted dispatch consumes the named nonce |

The historical regression fixture is the known unequal pair:

- AP1 session: `01a0382f-1f65-7450-81e6-f37832aaa111`;
- attempted-dispatch session: `01a038f9-f027-7853-b06c-bdf3e3ebaf25`.

Any verifier that reports those unequal values as same-thread must fail. Passing the regression contract proves only event-order handling; it does not prove identity, hidden-writer absence, confinement, product behavior, or release readiness.

## 7. Fail-closed recovery

Before AP1 admission, any missing, stale, duplicate, guessed, edited, reconstructed, dirty, divergent, ambiguous, or uninspectable evidence blocks without making the tuple available. After AP1 admission, loss of same-thread continuity consumes the tuple. After any attempted dispatch, every terminal outcome consumes it.

If neither output exists after a terminal failure, report `BLOCKED / CONSUMED`. If either exists without a verified complete pair and exact two-file-only post-state, report `RECOVERY_REQUIRED / CONSUMED`. Do not overwrite, delete, stage, commit, clean up, complete, or retry under that tuple.

Local sessions, Git, hashes, clock, process checks, marketplace/cache state, and filesystem observations remain same-user mutable. They are not an OS sandbox, signed authorization ledger, trusted counter/time source, atomic two-file transaction, cryptographic one-use grant, or proof of hidden-writer absence.

## 8. Explicit non-authorization

This Mode C successor does not clear `AUD-ACT-001`, issue GO or CONDITIONAL GO, authorize its proposed audit/review identities, issue AP1, select a skill, dispatch `l7-build`, make the nonce available, create either Wave 1 output, or authorize design, implementation, product/harness/plugin changes, additional Git effects, network/provider trials, release, deployment, cleanup, or external effects.

## 9. Compact assurance case

| Element | Statement |
|---|---|
| Claim | The inert successor corrects the cross-thread ambiguity without editing or laundering the failed-attempt evidence. |
| Argument | It permanently consumes the failed nonce, creates a searched fresh proposal, binds AP1 and dispatch to one explicit session ID, prohibits intervening routing and relative references, and preserves inherited fail-closed/no-overwrite controls. |
| Evidence | Original AP1/dispatch session IDs and event order; terminal output-absence check; clean Git tip; exact predecessor hashes; Mode B `AUD-ACT-001`; companion remediation record. |
| Assumptions | Local sessions and tools are accurate; same-user tampering, hidden writers, clock rollback, and unobservable concurrent claims are absent. |
| Defeaters | Session-ID mismatch, missing original AP1, relative-reference substitution, ID/nonce collision, changed lineage, dirty state, hidden writer, partial output, or any inherited binding mismatch. |
| Assurance ceiling | Mode C correction only; a fresh independent Mode B audit must decide whether the finding is adequately remediated. |

## 10. Exactly one next gate

After final SHA-256 values and the exact two-record remediation commit are reported, stop. The only next action is a fresh-session independent `level7-dev-loop:l7-release` Mode B audit of this exact candidate, companion, commit, failure evidence, and regression contract.

No qualified review, AP1, skill selection, Wave 1 work, release, or deployment may occur before that audit independently clears progression.
