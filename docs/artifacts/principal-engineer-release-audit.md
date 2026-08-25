# Level 7 Dev Loop — Principal Engineer Release Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-ORC-AMD-002` |
| Artifact type | Post-remediation, separate-context, read-only targeted governance audit; not a product, security, release, or deployment audit |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Date | 2026-08-25 |
| Status | **FINAL** |
| Audit mode | `l7-release` Mode B; candidate/runtime/package state read-only; this record is the only durable evidence write, while `make verify` may refresh documented derived `.cache` artifacts |
| Audit authority | Owner message on 2026-08-25: “I approve l7-release in read-only audit mode for exact L7-AMD-ORC-001 version 0.1.1 SHA-256 5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7; no amendment edits, Wave 1 artifacts, or product code.” |
| Candidate | [`L7-AMD-ORC-001`](orchestration-plan-host-binding-amendment.md) 0.1.1 |
| Candidate SHA-256 audited | `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` |
| Prior audit | [`L7-AUD-ORC-AMD-001`](orchestration-plan-host-binding-amendment-audit.md), SHA-256 `80fe801897d3f65a433a9c4b584301ea83457e61c441474b6d0b8bc7f69c9ddb`, `NO_GO` for candidate 0.1.0 only |
| Candidate-declared risk / maximum activated effect | `R3` authorization-identity binding / one local `A1` two-record proposal |
| Mode B classification | **CONDITIONAL GO**, solely to the mandatory R3 qualified-review gate |
| Candidate-protocol verdict | **GO_FOR_R3_QUALIFIED_REVIEW** |
| Activation state | **INERT / BLOCKED** pending qualified review, fresh-thread AP1, post-approval token selection, and invocation-time preflight |
| Exact invocation token or chip | `NOT_EVALUATED` by design; none was selected, inserted, reconstructed, or authorized during this audit |
| Audit-record SHA-256 | Computed after the final write and reported in the completion handoff; not self-embedded |

## 1. Decision

The exact `L7-AMD-ORC-001` 0.1.1 candidate at SHA-256 `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` may advance **one gate only**: a structurally independent review by the named qualified human/domain reviewer required by amendment §4. This audit found no unresolved Blocker, Critical, High, or Medium finding. It therefore returns the only favorable result the candidate permits at this phase: `GO_FOR_R3_QUALIFIED_REVIEW`.

This result does not activate or approve the amendment, authorize Wave 1, establish current-session AP1, select an invocation token, or issue product/security/release/deployment assurance. The candidate remains inert. Within the generic `l7-release` three-way vocabulary, `CONDITIONAL GO` means only that the exact candidate and this exact audit record are eligible for the next qualified-review gate; it is not a release or activation `GO`.

| Severity | Count |
|---|---:|
| Blocker | 0 |
| Critical | 0 |
| High | 0 |
| Medium | 0 |
| Low | 0 |
| Info | 2 |

## 2. Prior material-finding dispositions

| Prior finding | Disposition | Reproduced evidence |
|---|---|---|
| `AUD-HB-001` — AP1 authority lost across the new-thread boundary | **CLOSED** | Amendment §3.2 now requires AP1 in the same fresh invocation thread, after both review gates and before token selection; it binds the full evidence chain, action, target, two-file scope, exact host/source identity, A1 ceiling, validity, unused attempt, and sole-writer condition. Section 4 steps 4–7 repeat those bindings before dispatch and before each write. Persisted, inherited, quoted, delegated, or earlier-thread approval is explicitly `AP0` and fail-closed. |
| `AUD-HB-002` — R3 path omitted structurally independent qualified review | **CLOSED** | Amendment §3.2 and §4 step 3 require a named qualified human/domain reviewer who is structurally independent from the author and every remediator. The digest-bound record must carry identity, role, qualification evidence, independence/conflicts, methods/evidence, findings/dispositions, residual risk, scoped decision, and validity. Section 4 and §8 prohibit author/remediator self-review, self-downgrade, activation approval, and release/activation assurance. |
| `AUD-HB-003` — exact token was required both before and after approval | **CLOSED** | Amendment §3 is now phase-separated. Section 3.1 contains audit-time component/package/discovery bindings; §3.2 makes the exact token an invocation-time binding. Section 4 steps 1, 4, and 5 require this audit to leave the token `NOT_EVALUATED`, require fresh-thread AP1 first, and only then permit a fresh `/skills` selection and exact user-role dispatch. Audit-time, stale, guessed, reconstructed, edited, or earlier-thread selections are non-authorizing. |

The current full candidate was audited independently. Candidate 0.1.0 bytes are not retained as a separate file in the repository, so an exact byte-for-byte remediation diff is not reproducible from the current tree; this audit does not rely on such a diff. It instead binds and evaluates all current 0.1.1 bytes, the immutable prior audit record, the governing requirements, and the live §3.1 evidence.

## 3. Informational findings and residual limits

### AUD-HB-004 — Info — Replay, confinement, and TOCTOU controls remain governance-only

The nonce, source/digest checks, one-attempt state, process observation, no-overwrite rules, and workspace snapshots reduce accidental and confused-deputy risk, but they are not an OS sandbox, atomic two-file transaction, cryptographic non-replay mechanism, trusted clock, or trusted monotonic counter. Amendment §§4–8 disclose this accurately, cap the effect at one local A1 proposal, and make every mismatch fail closed. This retained limitation does not block the next qualified-review gate and must not be removed or inflated into a security claim.

### AUD-HB-005 — Info — Multiple Codex-related sessions/processes remain visible

Multiple Codex CLI/app-server sessions and host processes were visible through `ps` during this audit. Current sole-writer eligibility is therefore not established for an invocation now. That is not an audit-time candidate failure: amendment §§3.2 and 4 deliberately defer the owner confirmation and host-visible sole-writer recheck until the bounded invocation, immediately before mutation. Any remaining ambiguity at that phase is `BLOCKED` and consumes the attempt after dispatch as specified.

## 4. Evidence reproduced

| Check | Audit result |
|---|---|
| Exact candidate identity | `PASS`; metadata reports `L7-AMD-ORC-001` 0.1.1 and recomputed SHA-256 is `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` |
| Parent plan | `PASS`; SHA-256 `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| Parent candidate manifest | `PASS`; SHA-256 `da1fc881dd12f779f55af4745109511ce92a25fcf2c953b893008b08c6c8c109`; every listed file verified |
| Parent audit | `PASS`; SHA-256 `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` |
| Parent approval | `PASS`; SHA-256 `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` |
| Orchestration transitive input freeze | `PASS`; SHA-256 `ef17c49d7ceae115b476c2945fba4149f63094beade4cf8c0ba2d4cf652d2b16`; every listed file verified |
| Foundation input freeze | `PASS`; SHA-256 `428100ade80a848c2ae5aaa4d1d93876f0c4322cdd56ba2b19a9196593ca31ca`; all 26 protected entries verified |
| Prior 0.1.0 audit record | `PASS` as immutable historical evidence; SHA-256 `80fe801897d3f65a433a9c4b584301ea83457e61c441474b6d0b8bc7f69c9ddb` |
| Historical repository manifest | `PASS` as immutable historical evidence; SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf` |
| Development-manifest delta | `PASS`; the staged manifest differs from the historical manifest only by version `1.0.0 → 0.1.0` and the three required `interface` additions described by the candidate |
| Staged and effective cached manifests | `PASS`; valid JSON, byte-identical, both SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Marketplace binding | `PASS`; normalized `level7-dev-loop@personal`, local source `./plugins/level7-dev-loop`, `AVAILABLE` / `ON_INSTALL`, category `Developer Tools`; whole-file SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553` |
| Plugin registration | `PASS`; `codex plugin list` reports `level7-dev-loop@personal` as `installed, enabled`, version `0.1.0`, staged path `/Users/anuppandey/plugins/level7-dev-loop` |
| Staged/cached package closure | `PASS`; each exact root contains 13 regular files, zero symlinks, zero other non-directory file types, and no extras |
| Package hardlinks, ownership, and mode | `PASS`; every package file has `nlink == 1`; every entry is owned by `anuppandey`; no entry is group- or world-writable |
| Package path containment | `PASS`; every checked source/cache ancestor is a real directory at the declared absolute path, with no symlink component |
| All 12 skills | `PASS`; every staged and cached skill digest matches its protected `harness/foundation-inputs.sha256` entry |
| `l7-build` skill | `PASS`; staged and cached SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package content-set digest | `PASS`; staged and cached packages both reproduce SHA-256 `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` over the specified sorted encoding |
| Canonical component/discovery route | `PASS` for audit-time identity; the current host skill catalog exposes one canonical `level7-dev-loop:l7-build` component, consistent with the unique installed plugin and exact staged/cached source |
| Exact invocation token/chip | `NOT_EVALUATED`; deliberately not selected because a valid token can exist only after both review gates and fresh-thread AP1 |
| Host tuple | `PASS` for the bound local observation: `codex-cli 0.149.1`, macOS 26.5.2 build `25F84`, `arm64` |
| Canonical project/output paths | `PASS`; `/Users/anuppandey/Desktop/level7-dev-loop` and its `docs/artifacts` output parent resolve exactly and have no root-down symlink component |
| Git and output preconditions | `PASS` for audit-time state; Git is absent and both permitted Wave 1 output paths, including broken-symlink forms, are absent |
| Foundation harness | `PASS`; `make verify` exited 0 on Go 1.26.7, including offline install/verify/tidy-diff, policy, import, format, vet, type/compile, proving test, and two-build reproducibility comparison |
| Repository integrity around verification | `PASS`; the complete non-`.cache` file-content digest remained `b049ac899ce5c1717dd526af8a85bed6369aee4735f8fa6222c87320ca5ff8a7` before and after `make verify`; the non-`.cache` path/type/mode/link/size digest remained `967e6989971938eca8e8b6fad9fbe0d114b9263c6fbfe866685455a5a1e4de90` |
| Hosted CI and portability | `NOT_RUN` / `UNPROVED`; correctly outside this one-host local amendment audit |

## 5. Commands and test evidence

The audit used read-only inspection commands except for the one required audit-record creation and the harness's documented derived `.cache` refresh:

```text
shasum -a 256 docs/artifacts/orchestration-plan-host-binding-amendment.md
shasum -a 256 -c docs/artifacts/orchestration-plan-candidate.sha256
shasum -a 256 -c docs/artifacts/orchestration-inputs.sha256
shasum -a 256 -c harness/foundation-inputs.sha256
codex --version
codex plugin list
sw_vers
uname -m
realpath ...
find ...
stat -f ...
diff -u ...
jq ...
ps -axo ...
make verify
```

`git status --short --branch` returned the expected fatal result because neither this directory nor an ancestor is a Git repository. No Git mutation was attempted.

`make verify` completed in approximately 10.6 seconds and reported:

```text
go: no module dependencies to download
all modules verified
check-foundation-scope: PASS
check-import-boundaries: PASS (1 package set)
ok continuallabs.ltd/level7-dev-loop/internal/harness [no tests to run]
ok continuallabs.ltd/level7-dev-loop/internal/harness
reproducible harness binary SHA-256 1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f
```

Passing tests support only the bounded evidence above; they do not independently establish the verdict.

## 6. Assurance boundary and non-claims

This audit establishes only that the exact 0.1.1 candidate, its parent lineage, its audit-time local host/package bindings, and its fail-closed activation protocol support progression to the mandatory R3 qualified-review gate. It does not:

- activate or approve the amendment;
- satisfy or substitute for structurally independent qualified review;
- establish current-session AP1, one-attempt ownership, or sole-writer state;
- select, validate, preserve, or authorize an invocation token;
- authorize either Wave 1 artifact, design, implementation, Git, code, dependency, package, marketplace, host, infrastructure, publication, release, or deployment change; or
- establish controlled execution, product security, compatibility, release readiness, deployment readiness, or cross-host support.

Any candidate, parent, audit, review, approval, nonce, token, package, marketplace, CLI, OS, architecture, path, scope, expiry, revocation, or concurrency mismatch invalidates this result for activation use and must fail closed.

## 7. Exactly one next gate

Obtain a read-only review by a **named, evidence-qualified human/domain reviewer who is structurally independent of the amendment author and every remediator**. That durable record must bind exact candidate `L7-AMD-ORC-001` 0.1.1 SHA-256 `5684f9cf46f25998e324ce3863351890172b9626751895a36d8a9c3b093883e7` and exact model-audit `L7-AUD-ORC-AMD-002` plus its post-write SHA-256; satisfy every reviewer, evidence, finding, residual-risk, decision, and validity field in amendment §4; and issue at most `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW`.

Until that qualified-review gate closes, the amendment remains inert and Wave 1 remains blocked.
