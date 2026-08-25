# Level 7 Dev Loop — Same-Thread AP1 Remediation Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-ORC-AMD-006` |
| Artifact type | Fresh, separate, independent `l7-release` Mode B Principal Engineer audit |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **FINAL — UNTRACKED BY EXPLICIT AUDIT SCOPE; REQUIRES A LATER SEPARATELY AUTHORIZED SINGLE-RECORD GIT BINDING BEFORE USE** |
| Audit authority | Anup Pandey's user-role request for a fresh, separate, read-only audit of the exact `AUD-ACT-001` remediation, authorizing only this one no-overwrite audit record plus ignored `.cache` verifier effects |
| Candidate audited | `L7-AMD-ORC-005` 0.1.0 at `orchestration-plan-host-binding-amendment-git-successor-ap1-thread-remediation.md`, SHA-256 `976e49b9c360ff8d186aa66f7d8216c44d1cd85f891811b12640d131e7ecf698` |
| Companion audited | `L7-REM-AUD-ACT-001` 0.1.0 at `release-audit-remediation-aud-act-001.md`, SHA-256 `a4982a70bd713a99208fc4c30f4bc981a9e2044cb5a04611ae90f6ef752dd246` |
| Candidate Git binding | Commit `2f7d020704a1d393156529cf30581a6d1ad7148f`; tree `7a33b9140a3ffc581510c52fd3ef4bea2d2ad274`; parent `d6a9e0357db308a81139eeecf4fc1a06bb26928a`; exact two-record addition |
| Source finding | `AUD-ACT-001 — HIGH — attempted dispatch lacked an earlier same-thread AP1`; source Mode B decision `NO-GO` |
| Overall Mode B decision | **GO — ONLY FOR THE REMEDIATED INERT GOVERNANCE CANDIDATE TO ADVANCE ONE GATE** |
| Separate candidate-protocol result | **GO_FOR_R3_QUALIFIED_REVIEW** |
| Audit-record SHA-256 | Computed after this no-overwrite write and reported in the handoff; not self-embedded |

## 1. Executive decision and scope ceiling

The exact candidate, companion, and containing Git commit pass this independent Mode B audit for one purpose only: presentation to the later, separately authorized, structurally independent qualified-review gate required by candidate section 4.1.

There are zero unresolved `BLOCKER`, `CRITICAL`, `HIGH`, or `MEDIUM` findings. One `LOW` local-provenance risk is explicitly accepted for this narrow local A1 governance gate, and the remaining findings are `INFO`. Passing the Foundation harness does not supply authorization evidence and was not used as a substitute for the session-ordering analysis.

This result does not issue qualified review, `AP1`, skill selection, dispatch, nonce availability, Wave 1 authority, implementation authority, release approval, deployment approval, or product/security assurance. Any favorable result here permits only the later independent qualified-review gate. This audit record itself is not activation-chain evidence until its exact final bytes receive the candidate-required, separately authorized, single-record Git binding.

## 2. Independent method and repository map

The audit was performed from canonical root `/Users/anuppandey/Desktop/level7-dev-loop`. The repository contains 70 tracked paths at `HEAD`, including Foundation inputs and harness code plus 33 `docs/artifacts` paths. The parent contains 68 tracked paths; the two-path increase is exactly the remediation candidate and companion. No remediator conclusion was accepted without reproducing the material underlying evidence.

Representative commands, all run locally and read-only except for documented ignored `.cache` effects from the sole verifier, were:

```sh
pwd -P
realpath .
git status --porcelain=v2 --branch --untracked-files=all
git show -s --format='%H%n%T%n%P%n%s' HEAD
git rev-parse --show-object-format
git rev-parse --is-shallow-repository
git worktree list --porcelain
git remote -v
git for-each-ref --format='%(refname) %(objectname)' refs/replace
git diff-tree --no-commit-id --name-status -r d6a9e0357db308a81139eeecf4fc1a06bb26928a 2f7d020704a1d393156529cf30581a6d1ad7148f
git fsck --full --strict --no-progress
shasum -a 256 <artifact-or-package-path>
shasum -a 256 -c <protected-manifest>
stat -f '%N|%HT|mode=%Sp (%OLp)|links=%l|owner=%Su|group=%Sg' <path>
git show HEAD:<path> | shasum -a 256
test ! -e <output> && test ! -L <output>
git log --all --full-history --format='%H' -- <output>
rg -l -F <identity-or-nonce> /Users/anuppandey/.codex/sessions /Users/anuppandey/.codex/archived_sessions
sed -n '<line>p' <session.jsonl> | jq ...
codex --version
sw_vers
uname -m
jq -e . <manifest-or-marketplace>
make verify
```

## 3. Repository, Git, artifact, and manifest evidence

| Check | Independently reproduced result | Disposition |
|---|---|---|
| Canonical repository | `pwd`, `pwd -P`, `git rev-parse --show-toplevel`, and `realpath .` all returned `/Users/anuppandey/Desktop/level7-dev-loop`; root-down components are direct directories and the repository/output parent contain no symlink component | **PASS** |
| Branch and exact tip | Clean local `main` at `2f7d020704a1d393156529cf30581a6d1ad7148f` | **PASS** |
| Tree / parent / subject | `7a33b9140a3ffc581510c52fd3ef4bea2d2ad274` / sole parent `d6a9e0357db308a81139eeecf4fc1a06bb26928a` / `fix(audit-AUD-ACT-001): enforce same-thread AP1 dispatch` | **PASS** |
| Topology | SHA-1; non-shallow; one canonical worktree; sole ref `refs/heads/main`; zero remotes and replacement refs | **PASS** |
| Commit scope | Exactly two additions and no other delta: the candidate and companion; modes `100644`; 169 + 119 lines; all 68 predecessor tracked paths therefore preserved byte-for-byte | **PASS** |
| Git integrity | `git fsck --full --strict --no-progress` exited 0 with no output; no garbage was reported | **PASS** |
| Candidate | SHA-256 `976e49b9c360ff8d186aa66f7d8216c44d1cd85f891811b12640d131e7ecf698`; UTF-8 text/plain regular file; mode 0644; link count 1; size 15474; committed/worktree digests equal | **PASS** |
| Companion | SHA-256 `a4982a70bd713a99208fc4c30f4bc981a9e2044cb5a04611ae90f6ef752dd246`; UTF-8 text/plain regular file; mode 0644; link count 1; size 10483; committed/worktree digests equal | **PASS** |
| Direct predecessor | `L7-AMD-ORC-004` SHA-256 `c1901549b66405a8d837746fcc83f56d04097108b24ba3e02301c484634f6f25` | **PASS / PRESERVED** |
| Predecessor audit | `L7-AUD-ORC-AMD-005` SHA-256 `4c04e6e966f9270f5061937657fac0ce304de943cfcb2a0cbde6aa9ce62b5a85` | **PASS / PRESERVED** |
| Predecessor qualified review | `L7-REV-ORC-AMD-004` SHA-256 `f9301c9902ec15e62063ed962347ab640764c7cf31f9a05e7f812ed09b5ec763` | **PASS / PRESERVED** |
| Predecessor remediation companion | `L7-REM-AUD-GIT-001` SHA-256 `d124659ce655415252255e5e50cd117ba964d634720fd4d9ccd98a0dd9872ca2` | **PASS / PRESERVED** |
| Existing principal audit | `principal-engineer-release-audit.md` SHA-256 `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c` | **PASS / PRESERVED** |
| Earlier predecessor chain | Recomputed candidate/audit/review hashes `1013bcf7…df8d`, `93687b0f…384e`, `5684f9cf…83e7`, `80fe8018…9ddb`, `85187c07…50e26`, `85f5295f…ba74`, `8c9a495a…7c06`, and `4a0c8dab…374b8`; exact two-add-only delta proves no predecessor rewrite | **PASS / PRESERVED** |
| Protected manifests | All 54 entries in `orchestration-plan-candidate.sha256`, `orchestration-inputs.sha256`, `foundation-inputs.sha256`, and `harness-candidate.sha256` returned `OK` | **PASS** |
| Wave 1 outputs | Both `wave-01-change-contract.md` and `wave-01-specification.md` are absent under `test -e` and `test -L`, absent from `HEAD`, and have zero commits/objects in reachable Git history | **PASS** |

The reachable history is the seven-commit linear chain from root `08c38b69a2cd63b4adf27873756a09e363e0c5a4` through exact `HEAD`; there are no alternate reachable refs to conceal a forbidden path.

## 4. Original session and `AUD-ACT-001` reproduction

The original event facts were independently reconstructed from the JSONL records rather than from the remediator's summary:

1. AP1 session `/Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T14-55-07-01a0382f-1f65-7450-81e6-f37832aaa111.jsonl` line 1 binds session ID `01a0382f-1f65-7450-81e6-f37832aaa111`. Line 9 is an original `user` message at `2026-08-25T09:10:10.005Z` issuing AP1 for `L7-AMD-ORC-004-20260825-01`; its observed line SHA-256 is `458bac9af7dc30a8f5122835ad5ccc3a96f69465e8cc13daea75d37deeef18a1`.
2. That session's assistant line 154 admitted the nonce as `UNUSED`, explicitly not `IN_USE`. User line 160 then asked “what next?”, and lines 165/187 show an intervening `l7-next` routing pass. This history is preserved; the present audit does not rewrite the old AP1 as nonexistent.
3. Dispatch session `/Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T18-36-39-01a038f9-f027-7853-b06c-bdf3e3ebaf25.jsonl` line 1 binds the distinct session ID `01a038f9-f027-7853-b06c-bdf3e3ebaf25`. Line 9, at `2026-08-25T12:51:55.805Z`, is its first substantive user instruction and is the token-prefixed attempted dispatch. It says “the exact AP1 above” even though no earlier AP1 exists in that transcript, and the nonce itself is split after `L7-AMD-`; its observed line SHA-256 is `5208df27dc0ee0547b856c12bc1015313b230ef8cf6483e48271f4b35bb958c2`. Line 11 is the host-injected `l7-build` skill record, not authority.
4. Dispatch-session lines 67 and 75 correctly block and consume the attempted dispatch. The line-68 command checks both file and symlink forms, and line 70 reports both Wave 1 outputs `ABSENT`; the observed line-70 SHA-256 is `f5bf7978807894a3cf54dbbd62f4f30ad2c0f4a521bc54340e073598e428291b`. Inspection of all tool calls in that session found no `apply_patch` or output-writing call.
5. Source-audit session `/Users/anuppandey/.codex/sessions/2026/08/25/rollout-2026-08-25T19-00-53-01a03910-1f16-7e31-871d-a67f9c00cd11.jsonl` line 9 is the user-role Mode B request and explicitly prohibits remediation and Git mutation. Line 238, at `2026-08-25T13:26:18.479Z`, independently issues `NO-GO` with `AUD-ACT-001 — HIGH`; its currently observed line SHA-256 is `55f45566331dae5a6310db4aa2980efa106b7d9c29081a7247785d37a2774127`.

The old nonce validly reached only the local `UNUSED` admission reported in its AP1 session. There is no valid same-thread `UNUSED → IN_USE` transition: the attempted dispatch occurred in another session after an intervening routing pass. Under the predecessor's fail-closed rule and the confirmed attempted dispatch, `L7-AMD-ORC-004-20260825-01` remains **BLOCKED / CONSUMED / NON-REPLAYABLE**.

## 5. Fresh identity and nonce searches

The search covered the exact worktree, every reachable commit, 355 visible active-session JSONL files, 39 archived-session JSONL files, and the visible local session history.

| Search | Result | Disposition |
|---|---|---|
| Proposed audit ID `L7-AUD-ORC-AMD-006` | Repository/history occurrence is only candidate line 79's expected identity. Session occurrences are the prior handoff proposal, this audit request, and inherited current audit-worker context; none is a finalized claim. No archived-session match exists. | **AVAILABLE FOR THIS NO-OVERWRITE RECORD** |
| Audit target path | Absent in file and symlink forms and absent from reachable Git history immediately before this write | **PASS** |
| Proposed nonce `L7-AMD-ORC-005-20260825-01` | Repository occurrences are proposal/protocol mechanics in the candidate and companion. Visible session occurrences are the Mode C `NOT_AVAILABLE` handoff and this audit context. No activation, ownership, AP1, selection, dispatch, `UNUSED`, or `IN_USE` claim was found; no archived-session match exists. | **NOT_AVAILABLE** |

These are local observations, not proof against deleted, hidden, external, inaccessible, or perfectly concurrent claims. Any later gate must repeat the fail-closed search.

## 6. Correction adequacy

The correction adequately closes `AUD-ACT-001` at the documentary protocol level:

| Required correction | Candidate evidence | Result |
|---|---|---|
| Exact `AP1_SESSION_ID == DISPATCH_SESSION_ID` | Lines 90 and 93 record both IDs and require exact byte equality before accepting authority from dispatch | **PASS** |
| Original user-role AP1 before selection and dispatch in the same transcript | Lines 89 and 92-94 require the original AP1 as the first substantive user instruction, same-thread selection/dispatch, transcript inspection, and strict event order | **PASS** |
| No new/resumed/forked thread or intervening `l7-next` | Lines 90-91 require remaining in the exact thread and prohibit new/resumed/forked/reconstructed flow, intervening skill/“what next” routing, and `l7-next` | **PASS** |
| No relative-reference substitute | Lines 94 and 96 classify quoted, summarized, inherited, or cross-session AP1 as AP0 and prohibit “above,” “previous,” “earlier,” or equivalent substitution | **PASS** |
| Exact contiguous nonce binding | Lines 96 and 132 require the exact nonce contiguously in dispatch and reject the historical split/relative form | **PASS** |
| Consumption after continuity loss or attempted dispatch | Lines 91, 100-106, and 144 make post-AP1 continuity loss and any attempted dispatch terminal and non-replayable | **PASS** |
| Preserve inherited controls | Line 54 binds the exact `L7-AMD-ORC-004` predecessor SHA and every non-conflicting scope, identity, Git, host, path, preflight, transport, no-overwrite, snapshot, writer, validity, recovery, and non-authorization control | **PASS** |
| Restate critical boundaries | Lines 110-119 require fresh preflight, one direct two-add no-overwrite patch, no third path, and complete post-state checks; lines 142-152 restate fail-closed recovery, limitations, and explicit non-authorization | **PASS** |

Blanket inheritance is anchored to an exact immutable predecessor digest, and the critical mutation/recovery controls are restated rather than left only implicit. A mismatch at any inherited binding remains `BLOCKED`.

## 7. Host, plugin, marketplace, package, path, validity, and output state

| Binding | Independently observed result | Disposition |
|---|---|---|
| Host | `/opt/homebrew/bin/codex`; `codex-cli 0.149.1`; macOS 26.5.2 build 25F84; arm64 | **PASS AT OBSERVATION** |
| Marketplace | `/Users/anuppandey/.agents/plugins/marketplace.json`, valid JSON, SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553`; exactly one personal `level7-dev-loop` local entry at `./plugins/level7-dev-loop`, `AVAILABLE / ON_INSTALL`, category `Developer Tools` | **PASS AT OBSERVATION** |
| Enabled state | `/Users/anuppandey/.codex/config.toml` records `[plugins."level7-dev-loop@personal"] enabled = true` | **PASS AT OBSERVATION** |
| Staged/cache roots | Canonical direct directories `/Users/anuppandey/plugins/level7-dev-loop` and `/Users/anuppandey/.codex/plugins/cache/personal/level7-dev-loop/0.1.0`; owner `anuppandey`; no symlink escape | **PASS** |
| Staged/cache manifests | Valid JSON, byte-identical, version 0.1.0, each SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` | **PASS** |
| Staged/cache `l7-build` | Byte-identical, frontmatter name `l7-build`, `user-invocable: true`, each SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` | **PASS STATIC IDENTITY ONLY** |
| Package closure | Each root contains exactly the same 13 regular mode-0644 single-link files owned by `anuppandey`, zero symlinks/other nodes/extras/missing files, no group/world-writable entry; content-set SHA-256 `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` | **PASS** |
| Historical repository manifest | `.codex-plugin/plugin.json` is valid JSON at SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf`; historical evidence only, not the active package manifest | **PASS AS HISTORICAL INPUT** |
| Candidate validity | Observed `2026-08-25T19:39:48+0545` / epoch `1787666088`, before ceiling `2026-09-01T23:59:59+0545` / epoch `1788286499` | **PASS WITH UNTRUSTED-LOCAL-CLOCK LIMIT** |
| Output state | Both Wave 1 outputs absent in filesystem/symlink forms and reachable history | **PASS** |

The current injected skill catalog exposes the expected `level7-dev-loop:l7-build` identity from the cached 0.1.0 package. This observation is not a post-AP1 `/skills` selection and creates no selection or dispatch authority.

## 8. Sole pinned offline verifier

Exactly one permitted verifier command was run:

```text
COMMAND: make verify
EXIT STATUS: 0
MATERIAL OUTPUT:
go: no module dependencies to download
all modules verified
check-foundation-scope: PASS
check-import-boundaries: PASS (1 package set)
ok continuallabs.ltd/level7-dev-loop/internal/harness 0.616s [no tests to run]
ok continuallabs.ltd/level7-dev-loop/internal/harness 0.272s
REPRODUCIBLE BINARY SHA-256: 1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f
POST-VERIFIER GIT STATUS: clean main at 2f7d020704a1d393156529cf30581a6d1ad7148f; zero non-ignored entries; only ignored .cache/
```

`Makefile` lines 35-71 pin the repository-local Go toolchain and disable toolchain substitution, proxy, checksum database, VCS fetches, auth, telemetry, and network-intent variables. The verifier proves the inert Foundation harness and repeat-build property. It does not test AP1/session ordering and did not influence the authorization conclusion beyond confirming the inherited harness precondition.

## 9. Findings, severities, and dispositions

### `AUD-R6-GIT-001 — INFO — Exact Git and remediation identity reproduced`

Commit, tree, parent, subject, topology, integrity, canonical worktree, clean state, and exact two-add delta all match. Candidate and companion physical/committed identities match.

**Disposition:** `CLOSED / PASS`.

### `AUD-R6-EVD-001 — INFO — Protected predecessors, manifests, and output absence reproduced`

Every predecessor path is preserved by the exact two-add delta; the directly bound predecessor digests match; all 54 protected manifest entries verify; both forbidden outputs are absent from filesystem, symlink forms, `HEAD`, and reachable history.

**Disposition:** `CLOSED / PASS`.

### `AUD-R6-ACT-001 — INFO — Historical cross-thread failure and terminal old-nonce state reproduced`

The original AP1 and attempted dispatch are user-role events in unequal sessions. The old AP1 reached only `UNUSED`; no valid same-thread `UNUSED → IN_USE` transition occurred. The attempt is terminal with both outputs absent.

**Disposition:** `CLOSED / PASS`; preserve `L7-AMD-ORC-004-20260825-01` as `BLOCKED / CONSUMED / NON-REPLAYABLE`.

### `AUD-R6-NONCE-001 — INFO — Fresh nonce has no visible authority claim`

Exact repository/history/session searches found proposal and audit mechanics only for `L7-AMD-ORC-005-20260825-01`.

**Disposition:** `CLOSED / PASS WITH LOCAL-OBSERVABILITY LIMIT`; preserve `NOT_AVAILABLE`.

### `AUD-R6-CTRL-001 — INFO — Same-thread correction is adequate`

The candidate requires exact session-ID equality, original same-transcript AP1 ordering, uninterrupted thread continuity, no relative substitute, contiguous nonce binding, terminal consumption, and inherited/repeated mutation and recovery controls.

**Disposition:** `CLOSED / PASS` for documentary protocol correction.

### `AUD-R6-HOST-001 — INFO — Live host/plugin/package bindings match`

Host, marketplace, enabled state, staged/cache manifests, `l7-build`, package closure, canonical paths, validity, and output state match the inherited bindings.

**Disposition:** `CLOSED / PASS AT OBSERVATION`; repeat at every later gate.

### `AUD-R6-TEST-001 — INFO — Pinned offline Foundation verifier passes`

The sole `make verify` exited 0 and reproduced binary SHA-256 `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f` without non-ignored repository effects.

**Disposition:** `CLOSED / PASS`; explicitly not authorization evidence.

### `AUD-R6-SRC-001 — LOW — Source audit provenance is same-user mutable and not contemporaneously Git-bound`

The source audit exists only as an owner-writable local JSONL assistant record. Its initiating request prohibited Git mutation; overwriting the existing tracked standard Mode B artifact would have violated that constraint, and no alternative durable record path or commit was authorized. The present line digest is an observation, not a historical or append-only anchor. The committed companion is a later remediation account, not independent contemporaneous audit evidence.

**Disposition:** `ACCEPTED RESIDUAL FOR THIS NARROW GATE`. This audit independently reproduced the original event roles, session IDs, ordering, state, and output absence, and the successor is inert and fail-closed. Do not treat the source record as signed, append-only, externally attested, or sufficient for any higher assurance. This new audit record must receive its own separately authorized Git binding before later use.

### `AUD-R6-REG-001 — INFO — Regression proof is documentary/procedural, not executable`

Candidate section 6 and the companion provide the exact unequal historical fixture and future evidence contract. `Makefile` and the Go harness contain no transcript-ordering regression test. This is adequate to review an inert advisory governance-document correction and advance only to qualified review because `README.md` lines 3-7 expressly deny controlled enforcement and candidate sections 6 and 8 limit the assurance claim.

**Disposition:** `ACCEPTED FOR THE DOCUMENTARY GOVERNANCE SCOPE ONLY`. The proof does not establish coordinator compliance, atomic one-use enforcement, confinement, hidden-writer absence, product behavior, security qualification, release readiness, or deployment safety. Executable or host-enforced and independently durable proof is required before any controlled-execution/product/release assurance claim, while fresh original-transcript checks remain mandatory for every attempted activation.

### Severity summary

| Severity | Count | Unresolved |
|---|---:|---:|
| BLOCKER | 0 | 0 |
| CRITICAL | 0 | 0 |
| HIGH | 0 | 0 |
| MEDIUM | 0 | 0 |
| LOW | 1 | 0; accepted residual |
| INFO | 8 | 0 |

## 10. Residual and local-observability risks

- Git is local SHA-1 with no remote, signed-commit requirement, protected branch, external timestamp, or organizational attestation.
- Session JSONL, marketplace, config, package cache, filesystem, process visibility, and wall clock are same-user mutable observations.
- Searches cannot exclude deleted, hidden, inaccessible, external, or perfectly concurrent identity/nonce/ownership claims.
- No local process/open-file check proves absence of a hidden writer or establishes an OS containment boundary.
- Session-ID equality and nonce consumption are protocol checks, not a trusted counter, signed authorization ledger, atomic transaction, or cryptographic one-use grant.
- `make verify` covers the Foundation harness only and can refresh ignored `.cache`; it does not prove authorization lineage or runtime enforcement.
- This untracked audit record has no Git object identity until a later separately authorized exact single-record commit.

Every later gate must freshly reproduce the exact candidate/companion/audit bytes, Git lineage, nonce/ID uniqueness, current session events, host/package state, validity, output absence, and sole-writer assumption. Any mismatch fails closed.

## 11. Final decisions and mandatory stop

**MODE B DECISION: GO — scoped solely to advancing the exact inert remediation candidate one gate.**

**CANDIDATE-PROTOCOL RESULT: GO_FOR_R3_QUALIFIED_REVIEW.**

Any favorable result permits only the later independent qualified-review gate. It does not perform or authorize that review, its Git binding, AP1, skill selection, Wave 1 work, implementation, release, deployment, cleanup, or any external effect.

Stop here. Do not perform the next gate in this audit turn.
