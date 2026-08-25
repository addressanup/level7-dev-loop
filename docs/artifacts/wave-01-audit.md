# Level 7 Dev Loop — Wave 1 Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-001` |
| Artifact type | Independent read-only Principal Engineer audit; Level 7 Release Validation Mode B |
| Date | 2026-08-25 |
| Review mode | Structurally separate reviewer; exact frozen candidate and evidence inspection |
| Scope | Wave 1 build-control development checkpoint only |
| Verdict | **NO-GO — WAVE 1 DEVELOPMENT CHECKPOINT NOT CLEARED** |
| Production/release meaning | None; this is not a production-release, support, deployment, publication, AP2, or AP3 decision |
| Candidate remediation | Not authorized and not performed |
| Audit effect | This new uncommitted record only; no candidate, evidence, Git, environment, remote, or external mutation |

## 1. Decision

The exact Wave 1 candidate is **NO-GO for the Wave 1 development checkpoint**. The candidate and evidence identities reproduce exactly, the current snapshot remains product-free and dependency-free, and the recorded local baseline/shadow matrix is internally consistent. Those facts do not clear the checkpoint because one `HIGH` and three `MEDIUM` findings remain open.

This verdict applies only to the first phase-aware relaxation of the local development harness. It does not assess or authorize a product, controlled mutation, actual-host support, security qualification, release, publication, deployment, or production operation. No such product or operation exists in this candidate.

The approved Wave 1 design states that any unresolved Blocker, Critical, High, or Medium finding requires a successor candidate, full verification, and fresh independent audit. Therefore `W01-AC-012` is not satisfied by the mere presence of this audit record.

## 2. Exact identity binding

### 2.1 Approved base

| Identity | Reproduced value |
|---|---|
| Commit | `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4` |
| Tree | `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Base-manifest source | `harness/wave-01-base.sha256` from the implementation candidate |
| Base-manifest result | `72` rows; every row reproduced from the exact base Git object; zero digest or path-set mismatch |

### 2.2 Implementation candidate

| Identity | Reproduced value |
|---|---|
| Commit | `4bf485022aeb23cab51856a4749663d2d6f78619` |
| Tree | `6f59ca1d203e982b1d76b617a5209c85cb9d9d7e` |
| Parent | `3afde82cac5eaac346f6daae7fa2c0337073e323` |
| Manifest | `docs/artifacts/wave-01-candidate.sha256` |
| Manifest SHA-256 | `fb4f092fbbbaf85b8bc3e073e2853d2ce4ae0f06ea1d61ec92d3d6bbba6579bd` |
| Manifest result | `33` rows; zero digest mismatch |
| Base-to-candidate delta | `34` paths including the manifest itself; the manifest path exclusion leaves exactly the recorded `33` rows |
| Path-set result | Exact match between the base-to-candidate changed set and manifest rows after the approved manifest exclusion |
| Whitespace/deletion result | `git diff --check` passed; no base path was deleted |

### 2.3 Evidence-only child

| Identity | Reproduced value |
|---|---|
| Commit | `ce62a1f44ec4c2a28d0039df2a6ae2421f4d48a4` |
| Tree | `306decf8f2de507c8a56de8823e793f0e4ef8a52` |
| Parent | `4bf485022aeb23cab51856a4749663d2d6f78619` |
| Evidence file | `docs/artifacts/wave-01-evidence.md` |
| Evidence SHA-256 | `6f473496d3c2791fa996ce3bbea503621fcd6e875fa02e55e4238d4d94bba2dd` |
| Parent-to-child delta | Exactly `docs/artifacts/wave-01-evidence.md`; zero non-evidence path difference |

The digest-binding amendment's acyclic order is therefore reproduced:

```text
approved base
  -> implementation candidate + self-excluded candidate manifest
  -> evidence-only direct child
  -> this independent audit record
```

This audit binds both the exact implementation-candidate tuple and exact evidence-only tuple. Any change to candidate or evidence bytes invalidates this audit.

## 3. Authority and review method

The accountable owner authorized Mode B independent review and exactly one filesystem write: this uncommitted file. The reviewer did not modify a candidate file, evidence file, Git ref/index/config, cache, toolchain, environment, remote, or external system. No remediation, stage, commit, branch, merge, hosted workflow, network request, publication, release, or deployment occurred.

The audit mapped and inspected:

- the approved change contract, specification, design, digest-binding amendment, approval record, module decision, inert grant-ladder proposal, and evidence record;
- the approved base, phase/path/base/candidate manifests, support/prototype records, ownership record, module/action/toolchain/import locks, and historical Foundation controls;
- the complete base-to-candidate and candidate-to-evidence diffs;
- `internal/harness/buildcontrol` trace, claim, scope, ownership, file-shape, diagnostic, and test logic;
- `Makefile`, `scripts/harness/check-import-boundaries.sh`, `.github/workflows/harness.yml`, `README.md`, `go.mod`, and `harness/modules.lock.tsv`; and
- current repository status, remote absence, product/dependency absence, common secret patterns, and grant-amendment runtime non-use.

Representative read-only commands:

```text
git status --porcelain=v2 --branch --untracked-files=all
git log --format=... ee181b7..ce62a1f
git cat-file -p <base|candidate|evidence-commit>
git ls-tree -r --name-only <base|candidate>
git diff --name-status --stat ee181b7..4bf4850
git diff --name-only 4bf4850..ce62a1f
git diff --check ee181b7..4bf4850
git show <commit>:<path> | shasum -a 256
comm -3 <exact Git object path set> <manifest path set>
rg -n <contract, test, bound, effect, and secret-pattern queries>
nl -ba <reviewed source files>
```

One initial read-only base-manifest loop omitted braces around a shell variable adjacent to `:path`; Git rejected the malformed object expression and that attempt produced no validation evidence. The corrected `${commit}:${path}` command was then used and produced the `72`-row, zero-mismatch result above.

Per the audit authorization, no `make`, Go, shell test harness, hosted CI, or other command that intentionally writes caches or environments was run. The local baseline/shadow results below are cited from the exact evidence-only child, not independently re-executed.

## 4. Positive observations

The following observations are supported by direct source/object inspection:

- The root module transition is consistent across `go.mod`, `harness/modules.lock.tsv`, and the module-derived Makefile import path: `github.com/addressanup/level7-dev-loop`. The updater remains `reserved` at `cmd/l7up` with `UNSET` identity.
- The module ownership/location is correctly labeled `USER_ASSERTED`; there is no remote and no repository/account authentication, publication, compatibility, or support claim.
- `go.mod` has no `require`, `replace`, `exclude`, or `retract` directive. `go.sum`, `vendor/`, product commands, planned product internals, updater code, generated packages, and runtime grant code are absent.
- The candidate changes only the exact Wave 1 path set. Historical Foundation inputs, prototype skills/manifests, and `scripts/harness/check-foundation-scope.sh` remain byte-bound by the reproduced base/protected manifests.
- The controller has no command-line policy mode, repair mode, environment-selected phase, network import, `os/exec`, clock, randomness, or filesystem-write call. It reads fixed inputs and emits a decision.
- The active Makefile policy target invokes the standard-library-only controller with the pinned local toolchain; the predecessor shell checker is preserved but is not the active policy target.
- The import checker now rejects the exact `internal/harness` path and descendants while rejecting sibling-prefix aliases. Its current two-package graph contains no product package.
- The configured workflow retains read-only contents permission, pinned checkout, disabled credential persistence, blocking baseline, allowed-to-fail shadow, and no secret reference or `pull_request_target` trigger. It is configuration only and remains `NOT_RUN`.
- README/module/evidence wording consistently withholds product, support, controlled-mutation, security-qualification, hosted, publication, release, and deployment claims.
- The grant-ladder amendment is marked proposed/inert, is not read by runtime or harness logic, issues no authority, and still requires separate security/boundary review plus separate normative approval.
- A read-only common credential-pattern scan over the exact candidate reported no matching file. This bounded pattern check is not a comprehensive secret audit.

## 5. Severity model

| Severity | Audit meaning |
|---|---|
| `BLOCKER` | The candidate or evidence cannot be identified or reviewed reliably. |
| `CRITICAL` | A demonstrated path can create catastrophic authority, integrity, or external-effect failure. |
| `HIGH` | A central R3 boundary or explicit checkpoint criterion is materially unproved or bypassable. |
| `MEDIUM` | A required contract is incomplete or unreliable and must be corrected before this checkpoint. |
| `LOW` | A real bounded conformance/maintainability defect that does not independently block the checkpoint. |
| `INFO` | Verified fact, limitation, or future gate with no candidate defect established. |

Finding counts: `BLOCKER 0`, `CRITICAL 0`, `HIGH 1`, `MEDIUM 3`, `LOW 1`, `INFO 2`.

## 6. Findings

### `AUD-W01-001` — `HIGH` — Required permanent adversarial coverage is incomplete

**Affected acceptance:** `W01-AC-007`; consequently `W01-AC-012`.

The approved design §14.3 requires a table-driven negative suite for every listed category and states that each broken candidate must fail for its intended stable rule. The committed suite covers useful cases, including malformed/reversed requirement expressions, duplicate definitions/owners, claim/disposition row drift, missing/changed phase bindings, path-policy drift, add/modify/delete scope failures, manifest shape, file types/hardlinks, module/updater state, path aliases, ownership overlap, and protected-writer denial.

It does not provide permanent negative cases for all required categories. In particular:

- no summary-total tampering case exists, and `trace.go` does not parse or compare the displayed summary total despite design §7.4;
- missing definitions, overlapping requirement ranges, zero/two-owner states, and every allocation-total drift are not each exercised for their intended rules;
- plugin-to-mutation, cross-host inheritance, A3/A4/A5 execution, stable/dual/enforcement, generic-for-specialist substitution, and P0/P1/P2 drift are not represented as distinct broken fixtures;
- no deliberately broken package graph proves external-module detour, actual harness import, `unsafe`, or forbidden transitive-import rejection; the shell script contains only matcher positive/sibling-prefix controls, and the passing current two-package graph has no product package to exercise those failures; and
- cap exhaustion, process exit status, no-repair behavior, and repeat-run determinism do not have permanent negative/integration tests. The evidence records two repeated successful controller runs, but that is not the required permanent fixture.

No `testing/fstest.MapFS` fixture layer described in design §4.3 is present. Therefore the evidence statement that positive and adversarial trace/claim/scope/module/file/ownership/import cases satisfy `W01-AC-007` overstates the committed proof. Passing the current product-free graph does not prove that the correct import rule detects a future deliberately broken graph.

**Required disposition:** Add deterministic fixtures for every §14.3 category, assert the intended stable rule and nonzero behavior, rerun both full local toolchains, freeze a new manifest/evidence chain, and obtain a fresh independent audit. Do not add a product path merely to create these fixtures; use the approved isolated fixture mechanism or an explicitly reapproved equivalent.

### `AUD-W01-002` — `MEDIUM` — Declared resource bounds are not enforced before untrusted allocation/traversal

**Affected acceptance:** fail-closed/bounded portions of `W01-AC-006` and the Wave 1 nonfunctional contract; consequently `W01-AC-012`.

The design §17.3 says file count, total bytes, Markdown line length, TSV rows, expanded IDs, findings, and output bytes are constant-bounded and that exceeding a cap returns `BLOCKED`. The implementation does not consistently enforce those limits before consuming the resource:

- `readStrictFile` calls `os.ReadFile` before checking `maxInputBytes`, then `validateStrictText` calls `bytes.Split` before deciding the line count. An oversized untrusted file can be fully allocated before `BCTL-011`/`BCTL-015` is emitted.
- `loadSkillInventory` uses `os.ReadDir` without an entry cap before the repository scope check runs.
- `scanRepository` does not count or cap directories. After `maxRepositoryFiles`, it continues the complete walk and appends one finding per excess file; `maxFindings` limits printing, not finding allocation or traversal.
- `expandRequirementExpression` caps one range but has no cumulative expanded-ID cap. A bounded input line may contain many ranges and allocate far more IDs than the approved requirement maximum before validation rejects the final totals.
- repository file content is sized before `os.ReadFile`, but the stat/read sequence is not size-limited at the reader and is susceptible to size growth between inspection and read.

These defects do not demonstrate a current policy bypass or external effect. They do mean the controller may exhaust memory/time or fail to return the promised bounded `BLOCKED` result when repository content is adversarial, which is material for the first fail-closed scope gate.

**Required disposition:** Use limited/streaming reads, cap directory and file traversal with an early blocking stop, cap accumulated findings and expanded IDs before append, make size-change behavior fail closed, and add boundary-at-limit/over-limit permanent tests.

### `AUD-W01-003` — `MEDIUM` — The claim/support contract does not encode the complete approved boundary

**Affected acceptance:** `W01-AC-002` and `W01-AC-005`; consequently `W01-AC-012`.

The specification requires the versioned support/claim record to state one local repository/worktree, distinguish development evidence, unsupported behavior, and release-blocking proof, state that plugin installation is insufficient for mutation authority, and keep P0/P1/P2 scope/priority change rules aligned with the backlog.

`harness/support-matrix.tsv` and `expectedSupport` contain ten exact rows for advisory/controlled surfaces, three proof profiles, A3/A4 handoff, A5 exclusion, dual-host withholding, and stable withholding. They contain no row or field for the one-repository/worktree limit, plugin-install insufficiency, development-evidence state, release-blocking proof, or backlog P0/P1/P2 allocation/change rule. `claims.go` validates only equality with that incomplete ten-row constant plus prototype dispositions; it does not derive or compare the priority/change contract from `feature-backlog.md`.

The README and protected planning prose state several of these limits truthfully, and base-byte protection would reject an unauthorized edit to the existing backlog. That is not equivalent to the approved machine-validatable support/claim record or an intended-rule claim fixture. The evidence disposition for `W01-AC-005` cites phase/path/owner/module drift, which is not the criterion's required scope/priority impact-diff and accountable-approval rule.

**Required disposition:** Extend the approved record/schema and validator to cover the omitted v1 boundary and priority/change-control semantics, add source-derived or digest-bound comparison to the authoritative backlog, and add the missing false-claim fixtures before regenerating the candidate.

### `AUD-W01-004` — `MEDIUM` — Shared-control ownership is not complete against the authoritative ownership table

**Affected acceptance:** `W01-AC-010`; consequently `W01-AC-012`.

The specification requires ownership for requirement IDs/release allocation and all listed shared-control classes. The design further says `control-ownership.tsv` is cross-validated against the orchestration ownership table.

The implementation checks the TSV against a candidate-local hard-coded `expectedOwnership` map and cross-checks only the Wave 1 path-policy rows. It does not parse or digest-bind the authoritative orchestration §10 ownership table. More concretely, neither `control-ownership.tsv` nor `expectedOwnership` assigns the required requirement-ID/release-allocation source class covering `docs/artifacts/requirements.md` and `docs/artifacts/feature-backlog.md`; the existing `harness/` prefix does not cover those files. Several orchestration classes are represented only by narrower future path guesses rather than a source-table cross-check.

The current base manifest still protects those existing documents from Wave 1 byte drift, so no unauthorized current modification was observed. The recorded count of 31 self-consistent controls nevertheless does not prove that the required authoritative ownership set is complete.

**Required disposition:** Add the missing mandatory class mappings and validate the record against a source-derived or separately digest-bound authoritative ownership inventory, with missing/extra/source-drift fixtures.

### `AUD-W01-005` — `LOW` — Success output omits required gate version and source digests

**Affected contract:** Wave 1 specification §5 minimum success semantics.

The specification requires success to identify the validator/gate version and exact relevant source digests. `main.go` emits phase, counts, allocations, ownership, inventory, and changed-file totals, but no gate version or source digest. The evidence and manifest supply external binding for this exact review, so this is not a demonstrated acceptance bypass; it is still a direct developer-interface conformance gap.

**Required disposition:** Add stable gate-version and relevant source-digest fields in the deterministic output contract and cover their stability in tests.

### `AUD-W01-006` — `INFO` — Local test evidence reproduced as a record, not rerun by the reviewer

The exact evidence-only child records passing `make verify GO_VERSION=1.26.7` and `make verify GO_VERSION=1.27.0`, with reproducible test-binary hashes `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` and `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a`. It also records passing policy, import, type, unit, vet/format, repeat-build, manifest, static effect, and bounded secret-pattern checks.

Those results are same-user local development evidence. The reviewer intentionally did not rerun them because the audit authority prohibited commands that intentionally write caches or environments. Their exact record and identity are verified; their execution is not independently reproduced by this audit.

### `AUD-W01-007` — `INFO` — Grant proposal and external qualification remain separate gates

The grant-ladder amendment is coherent as an inert proposal and no Wave 1 code consumes it. This audit does not approve that proposal's security model, supersede `TDR-013`, authorize any grant kind, or clear controlled mutation. The proposal's required independent security/boundary review, schemas, issuer/root separation, protected infrastructure, qualification/evaluation/pilot/stable evidence, and owner decisions remain future `NOT_RUN` gates.

## 7. Acceptance disposition

| Acceptance | Independent disposition |
|---|---|
| `W01-AC-001` | `PASS` for the exact current snapshot: 163 source-derived IDs, one recorded owner/allocation each, totals `140/18/5`; test-completeness gap is tracked by `AUD-W01-001`. |
| `W01-AC-002` | `NOT SATISFIED` — the machine support/claim record omits required boundary semantics (`AUD-W01-003`). |
| `W01-AC-003` | `PASS` for the exact current snapshot: 12 invocable prototype skills have one disposition and protected bytes reproduce. |
| `W01-AC-004` | `PASS` for current candidate wording: prototype/stable/dual-host/enforcement promotion is withheld. |
| `W01-AC-005` | `NOT SATISFIED` — P0/P1/P2 and impact/approval change semantics are not validated by the claim contract (`AUD-W01-003`). |
| `W01-AC-006` | `NOT CLEARED` — historical bytes and nominal unknown/stale/unowned rules exist, but bounded fail-closed behavior is incomplete (`AUD-W01-002`). |
| `W01-AC-007` | `NOT SATISFIED` — mandatory permanent adversarial categories are absent (`AUD-W01-001`). |
| `W01-AC-008` | `PASS` for local development identity only: module records agree; updater is reserved; external GitHub identity remains `USER_ASSERTED`/`NOT_RUN`. |
| `W01-AC-009` | `PASS` as an inert proposal only; separate security review and normative approval remain `NOT_RUN`. |
| `W01-AC-010` | `NOT SATISFIED` — ownership is internally self-consistent but incomplete against the authoritative class set (`AUD-W01-004`). |
| `W01-AC-011` | `PASS` as recorded local evidence only; both pinned local matrices are recorded green, while hosted CI remains `NOT_RUN`. |
| `W01-AC-012` | `FAIL / NO-GO` — exact identities are bound and this audit ran, but unresolved High/Medium findings prevent the R3 checkpoint. |
| `W01-AC-013` | `PASS` for the exact candidate: no dependency, product/runtime path, prototype edit, unexpected candidate path, external effect, stable claim, or bounded secret-pattern match was observed. |

No aggregate passing count overrides an unsatisfied acceptance criterion.

## 8. Residual risks and `NOT_RUN` states

- Hosted GitHub Actions on Ubuntu 24.04: `NOT_RUN`; workflow inspection only.
- GitHub repository/account ownership, remote binding, publication identity, and package-source authentication: `NOT_RUN`; no remote exists.
- Controlled Ubuntu/Bubblewrap/provider/model/host qualification and actual Codex/Claude lifecycle/differential tests: `NOT_RUN`.
- Independent execution reproduction of the local baseline/shadow matrix: `NOT_RUN` by this reviewer; exact recorded evidence inspected only.
- Fuzz, property, race, fault-injection, hostile-scale, and concurrency/TOCTOU verification of the new controller: `NOT_RUN` unless specifically present in the recorded ordinary suite; no independent claim is made.
- Grant-ladder security/boundary review, normative adoption, schemas, issuers, signing, revocation infrastructure, qualification, protected evaluation, pilot, stable grant, and controlled mutation: `NOT_RUN` or absent.
- Protected evaluator, AP2/AP3, signing, TUF, promotion, release authorization, release build, publication, deployment, exposure, and production monitoring: `NOT_RUN` or absent.
- Local Git, filesystem, SHA-256 records, and same-user test evidence remain mutable by the local owner. They support exact reproducibility but do not create external trust or qualified assurance.
- The audit record is intentionally outside the candidate/evidence manifests. At creation it is an uncommitted reviewer-only path; its final digest/Git binding, any remediation authorization, and any later integration are separate decisions.

## 9. Required next transition

The candidate must remain frozen. Under separately granted remediation authority, a different writer may address `AUD-W01-001` through `AUD-W01-005`, with one finding-specific change set at a time and regression proof. Because the fixes alter candidate bytes, the remediator must produce a new implementation commit/tree, regenerate the candidate manifest, rerun the complete baseline and shadow local matrix, create a new evidence-only binding, and request a fresh structurally separate Mode B audit.

No merge, Wave 2 start, deployment, release, publication, or external action follows from this audit. A remediation pass must not self-issue `GO`.
