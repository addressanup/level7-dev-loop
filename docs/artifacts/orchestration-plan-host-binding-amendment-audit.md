# Level 7 Dev Loop — Codex Host-Binding Amendment Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-ORC-AMD-001` |
| Artifact type | Post-write, separate-context, read-only targeted governance audit; not a product, security, release, or deployment audit |
| Date | 2026-08-25 |
| Candidate | [`L7-AMD-ORC-001`](orchestration-plan-host-binding-amendment.md) 0.1.0 |
| Candidate SHA-256 audited | `6eaac34d871a2a9fc3e92e46730673d74084d1752e8485182ba558e55fd25f6b` |
| Audit authority | Owner message on 2026-08-25: “I approve l7-release in read-only audit mode for exact L7-AMD-ORC-001 SHA-256 6eaac34d871a2a9fc3e92e46730673d74084d1752e8485182ba558e55fd25f6b; no amendment edits, Wave 1 artifacts, or product code.” |
| Audit mode | `l7-release` Mode B; candidate/runtime/package state read-only; this audit record is the only durable evidence write, while the harness may refresh documented derived `.cache` artifacts |
| Candidate-declared risk | `R3` — authorization identity binding |
| Overall verdict | **NO_GO** |
| Activation state | **INERT / BLOCKED** — the §4 activation gate is not satisfied |
| Candidate mutation during audit | None |
| Scope exclusions preserved | No amendment remediation, Wave 1 artifacts, product or harness code, manifest/skill/package changes, Git mutation, release, or deployment |

## 1. Decision

`L7-AMD-ORC-001` is **not eligible for activation or exact owner approval in its audited form**. The exact runtime and package bindings were reproducible, and the adversarial binding review found the substitution controls fail-closed at the stated A1 ceiling. However, two unresolved High findings and one unresolved Medium finding prevent the candidate-required `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW` result.

This is a governance `NO_GO`, not evidence that the staged plugin package is corrupt. The candidate remains the unmodified object of this audit and retains SHA-256 `6eaac34d871a2a9fc3e92e46730673d74084d1752e8485182ba558e55fd25f6b`.

| Severity | Count |
|---|---:|
| Blocker | 0 |
| Critical | 0 |
| High | 2 |
| Medium | 1 |
| Low | 0 |
| Info | 2 |

## 2. Findings

### AUD-HB-001 — High — AP1 authority is not preserved across the mandatory new-thread boundary

**Evidence.** Amendment §4 requires an exact owner approval and then a new Codex thread for the fresh `/skills` selection and invocation. [`L7-REQ-001`](requirements.md) §6.6 defines `AP1` as **current-session confirmed**, classifies a persisted approval record as `AP0`, and requires the applicable approval to be reconfirmed immediately before mutation. The inherited Codex invocation form in [`L7-ORC-001`](orchestration-plan.md) §4.2 names only the parent plan section; it does not itself bind the amendment digest, this audit digest, nonce, target/two-file scope, exact host, A1 ceiling, validity window, one-attempt state, or sole-writer condition.

**Impact.** Approval issued before opening the required new thread becomes provenance in that new thread rather than current-session authority. A token submission that does not rebind the complete authorization tuple can be mistaken for an executable AP1 grant.

**Required correction.** Make the invocation-thread submission an exact current-session reconfirmation of the logical action, amendment SHA-256, audit SHA-256, nonce, canonical target and two-file scope, host binding, A1 effect ceiling, expiry, unused one-attempt state, and sole-writer condition. The preflight must explicitly revalidate that activation chain and current-session AP1 immediately before the bounded mutation.

### AUD-HB-002 — High — The R3 activation path omits structurally independent qualified review

**Evidence.** The amendment classifies itself as `R3`, but §4 and §8 permit activation after only a separate-context model audit and accountable-owner approval. The candidate correctly says that the model audit is not a qualified human security review, yet it does not add the missing qualified-review gate. [`L7-REQ-001`](requirements.md) `L7-FLOW-008` requires every R3 work item to have both an assurance case and structurally independent qualified review; §6.5–§6.6 and the confirmed constraints preserve that requirement. `L7-RISK-002` prohibits an authoring agent from approving its own risk downgrade.

**Impact.** The proposed activation sequence is weaker than the governing R3 profile. Neither this separate-context model audit nor owner approval alone closes that mandatory reviewer role.

**Required correction.** Before activation, require a digest-bound review by a qualified reviewer who is structurally independent of the candidate author and remediator, and record reviewer identity, qualification, reviewed evidence, findings, and decision. Alternatively, a lower risk classification requires new evidence and an accountable decision through a separate reassessment; it cannot be self-issued by the authoring agent.

### AUD-HB-003 — Medium — The exact token is required both before and after approval

**Evidence.** Amendment §4 step 1 requires the post-write audit to recompute **every** live binding in §3, which includes the exact `/skills`-inserted token. Section 4 step 4 requires that token to be selected freshly in a new thread **after** audit and approval, and §5 rejects a selection that predates approval.

**Impact.** A compliant pre-approval audit cannot attest the exact post-approval token, while any token it does inspect is deliberately invalid for the later invocation. The phase ordering makes the token row impossible to satisfy literally as part of this audit.

**Required correction.** Split §3 into audit-time static/package bindings and invocation-time dynamic bindings. Audit the canonical component identity and expected discovery mechanism now; bind and validate the exact inserted token only after approval in the invocation thread, before dispatch.

### AUD-HB-004 — Info — Replay and TOCTOU controls remain governance-only

The nonce, process observation, snapshots, and one-attempt consumption are useful controls but are not an OS sandbox, atomic transaction, cryptographic non-replay mechanism, or trusted monotonic counter. The amendment accurately discloses this limitation in §5 and confines its claim to one local A1 proposal. The limitation must remain visible in any successor.

### AUD-HB-005 — Info — Multiple Codex sessions were visible at audit time

The later sole-writer preflight is not currently satisfied because other Codex sessions/processes were visible. This is not an amendment-audit failure: no activation or Wave 1 attempt is authorized now, and §4 already requires a fresh observation and owner confirmation immediately before an approved bounded attempt. A successor must retain that fail-closed gate.

## 3. Independent perspective results

| Perspective | Result | Material conclusion |
|---|---|---|
| Authority and lineage | `NO_GO` | Found `AUD-HB-001` High and `AUD-HB-003` Medium; the approval lineage and token timing do not survive the required new-thread transition as written. |
| Operational and assurance-profile conformance | `NO_GO` | Found `AUD-HB-002` High; the self-declared R3 path lacks mandatory structurally independent qualified review. It also identified the token-ordering concern, consolidated above at Medium. |
| Adversarial binding and fail-closed behavior | `GO_FOR_AP1_LOCAL_CANDIDATE_REVIEW` | Found no Blocker–Low defect in substitution, aliasing, package closure, path escape, replay, concurrency, partial-write, or authority-inflation defenses at the stated A1 ceiling; retained the two informational caveats above. |

These are separate-context **model** perspectives. They are not qualified human/domain reviews. The favorable adversarial result cannot override unresolved High or Medium findings from the other perspectives.

## 4. Evidence reproduced

| Check | Audit result |
|---|---|
| Exact amendment bytes | `PASS`; SHA-256 `6eaac34d871a2a9fc3e92e46730673d74084d1752e8485182ba558e55fd25f6b` |
| Parent plan | `PASS`; SHA-256 `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| Parent candidate manifest | `PASS`; SHA-256 `da1fc881dd12f779f55af4745109511ce92a25fcf2c953b893008b08c6c8c109` |
| Parent audit | `PASS`; SHA-256 `9b6e294639419c7bf17af1ca6af5d329f83beb1facaaa1d3b10841706b7e4e91` |
| Parent approval | `PASS`; SHA-256 `475870d1623014a8c5fb69e03994833867a9344d8fbe5ae85fef9a85e60dbf1d` |
| Orchestration transitive input freeze | `PASS`; SHA-256 `ef17c49d7ceae115b476c2945fba4149f63094beade4cf8c0ba2d4cf652d2b16` |
| Foundation input freeze | `PASS`; SHA-256 `428100ade80a848c2ae5aaa4d1d93876f0c4322cdd56ba2b19a9196593ca31ca` |
| Historical repository manifest | `PASS` as immutable historical evidence; SHA-256 `b3b1c2ce4708899073e9168ecf909bd2a009b800131e608fbdff9c284519a4cf` |
| Staged and effective cached manifests | `PASS`; both SHA-256 `202be0ca3b6ba80685f2b6bb520e839419faacdb65a7726be96af1170ae7f3f3` |
| Staged and cached `l7-build` | `PASS`; both SHA-256 `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |
| Package closure | `PASS`; each package has exactly 13 regular single-link files, no symlinks or extras, and all 12 skill digests match the protected historical inputs |
| Package content-set digest | `PASS`; staged and cached sets both SHA-256 `b1241ed16cbc2e4a2c560591c56daeb2d72444da1e562aa474a62d0ab04abd04` |
| Ownership and mode | `PASS`; package entries owned by `anuppandey`, with no group/world-writable entry |
| Marketplace binding | `PASS`; normalized `level7-dev-loop@personal` local-source binding reproduced; marketplace whole-file evidence SHA-256 `fab99932b6790dfb3ab11945808f3a89469b1288e32d517af22edc7046047553` |
| Plugin registration | `PASS`; `level7-dev-loop@personal` installed and enabled at version `0.1.0` |
| Canonical component identity | `PASS`; installed component resolves to `level7-dev-loop:l7-build` with the staged/cached skill digest above |
| Exact post-approval `/skills` token | `NOT_EVALUATED`; no pre-approval token can satisfy the candidate's fresh post-approval selection rule, as recorded in `AUD-HB-003` |
| Host tuple | `PASS` for this local observation: `codex-cli 0.149.1`, macOS 26.5.2 build `25F84`, `arm64` |
| Canonical project/output paths | `PASS`; exact non-symlink project root and `docs/artifacts` binding observed |
| Git and permitted-output preconditions | `PASS`; Git absent and both Wave 1 output paths absent |
| Foundation harness | `PASS`; `make verify` completed successfully during the audit, with only documented derived `.cache` effects |

## 5. Assurance boundary and non-claims

This record establishes only that the exact candidate and its live local evidence were reviewed under the stated read-only scope. It does not:

- activate or approve the amendment;
- authorize a fresh `l7-build` invocation or create either Wave 1 artifact;
- qualify the reviewer role required by R3;
- establish product security, compatibility, controlled execution, release readiness, deployment readiness, or cross-host support; or
- authorize remediation, code, manifest, skill, package, marketplace, Git, environment, or infrastructure changes.

## 6. Required gate and exactly one next action

Status is `REMEDIATION_REQUIRED`. The next permitted action is an owner-approved `l7-release` remediation pass limited to `AUD-HB-001` through `AUD-HB-003` and this amendment. Any remediation creates a new candidate digest and requires a new exact-digest post-write audit. The resulting R3 activation chain must still include the qualified independent review identified in `AUD-HB-002` before owner activation approval.

Until those gates close, `L7-AMD-ORC-001` remains inert and Wave 1 remains blocked.
