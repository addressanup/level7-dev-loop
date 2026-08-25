# Level 7 Dev Loop — Wave 1 Successor Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-002` |
| Artifact type | Fresh independent read-only Principal Engineer audit; Level 7 Release Validation Mode B |
| Date | 2026-08-26 |
| Review mode | Structurally separate successor reviewer; exact frozen candidate and evidence inspection |
| Scope | Wave 1 build-control development checkpoint only |
| Verdict | **NO-GO — SUCCESSOR WAVE 1 DEVELOPMENT CHECKPOINT NOT CLEARED** |
| Production/release meaning | None; this is not a production-release, support, deployment, publication, AP2, or AP3 decision |
| Candidate remediation | Not authorized and not performed |
| Audit effect | This reviewer-owned file update only; no candidate, evidence, Git ref/index/config, environment, cache, remote, or external mutation |

## 1. Decision

The exact remediated successor is **NO-GO for the Wave 1 development checkpoint**. All supplied candidate, evidence, manifest, remediation, and historical-audit identities reproduce. The registered path and ownership closure is exact; the candidate remains product-free and dependency-free; and the remediation adds substantial source and regression coverage. Those facts do not clear the checkpoint because two `MEDIUM` correctness/reliability findings remain open:

1. capped failure output is not repeat-deterministic because findings are truncated while iterating unordered maps and sorted only afterward; and
2. fixed authoritative inputs are opened before regular-file, single-link, and no-symlink validation, so a symlink can be consumed and a FIFO can block before the later repository scan rejects it.

The approved design says any unresolved `BLOCKER`, `CRITICAL`, `HIGH`, or `MEDIUM` finding requires another successor, complete verification, new evidence binding, and fresh independent audit. Consequently `W01-AC-006`, `W01-AC-007`, and `W01-AC-012` are not cleared.

This verdict applies only to the first phase-aware relaxation of the local development harness. It does not assess or authorize a product, controlled mutation, actual-host support, grant activation, security qualification, publication, release, deployment, or production operation.

## 2. Exact identity binding

### 2.1 Approved base and preserved source audit

| Identity | Reproduced value |
|---|---|
| Approved-base commit | `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4` |
| Approved-base tree | `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Base manifest | `harness/wave-01-base.sha256`; `72` rows |
| Base-manifest result | Every row reproduced from the exact base Git object; zero digest or path-set mismatch |
| Original NO-GO audit commit | `42ea8ea5457627090c9d728d3a0469f7696b6d9d` |
| Original NO-GO audit SHA-256 | `54e03a8a20ced5a944b8e2506dd424581f9fbf009bf2711955cb07535d8a1224` |
| Historical preservation | Original audit bytes reproduce at `42ea8ea`; that commit is an ancestor of the successor |

The generic skill-output record `docs/artifacts/principal-engineer-release-audit.md` remains a protected historical base input. Its current SHA-256 is `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c`; it was inspected and not edited.

### 2.2 Successor implementation candidate

| Identity | Reproduced value |
|---|---|
| Commit | `808fc16b1c13c96e30c66f08a97dd2a014b31db0` |
| Tree | `bd914fdcb88aad54be5acf19c60dbcbb35f2d709` |
| Parent | `f8788906a17a525c47435a69876171f21daef545` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256` |
| Candidate-manifest SHA-256 | `d8fcfbd3d78fa104449c7edcc9d68444fdd8a1a776e844993a41b60527f0acb0` |
| Manifest result | `34` rows; zero digest mismatch |
| Base-to-candidate delta | Exactly `37` paths, `4,977` insertions, `16` deletions, and zero deleted base path |
| Path closure | The `37` changed paths exactly equal the `37` data rows in `harness/wave-01-paths.tsv` |
| Manifest closure | Removing only the approved manifest/evidence/audit exclusions from the `37`-path delta leaves exactly the `34` manifest rows |
| Candidate inventory | `103` Git files and `1,241,419` blob bytes; no symlink or submodule entry |
| Whitespace result | `git diff --check ee181b7..808fc16` returned no finding |

### 2.3 Evidence-only child and remediation record

| Identity | Reproduced value |
|---|---|
| Evidence commit / current `HEAD` | `c0f834f7318081d99d0dcbbc2b82d907961be633` |
| Evidence tree | `55536582355517969b802be11856398fafe3b3e5` |
| Evidence parent | `808fc16b1c13c96e30c66f08a97dd2a014b31db0` |
| Evidence SHA-256 | `60122e391bb486cc1156998a5febbd5fab1f42daea8203da65968edfee7fc8eb` |
| Parent-to-child delta | Exactly one modified path: `docs/artifacts/wave-01-evidence.md` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md` |
| Remediation-record SHA-256 | `0a5f3895b263f8f7acf82611d066d35f8f8093257f9baf9b61c3349766f5be28` |

The reproduced acyclic chain is:

```text
approved base ee181b7
  -> original candidate/evidence 4bf4850 / ce62a1f
  -> preserved NO-GO audit 42ea8ea
  -> finding-specific remediation commits
  -> successor candidate + self-excluded manifest 808fc16
  -> evidence-only child c0f834f
  -> this fresh reviewer-only audit record
```

Any successor-candidate or evidence byte change invalidates this audit.

## 3. Authority, repository map, and review method

The accountable owner authorized a fresh Mode B review and exactly one filesystem mutation: update the registered `docs/artifacts/wave-01-audit.md`. `wave-01-design.md` §§4.3 and 15.3, `harness/wave-01-paths.tsv`, and `harness/control-ownership.tsv` bind that path to the independent reviewer and reserve it as `audit-only`. The worktree was clean immediately before this audit edit. There is one worktree, on local branch `feat/wave-01-build-control`, with no configured remote.

The repository map at the exact successor contains eight top-level control/documentation/source families: plugin metadata, configured CI, `docs/`, `harness/`, `internal/`, `references/`, `scripts/`, and protected prototype `skills/`, plus root manifests and build files. Wave 1 source is confined to `internal/harness/buildcontrol`; its only current package peer is the inert `internal/harness` proving package. No product command/runtime, updater, grant runtime, generated package, `go.sum`, `vendor/`, production dependency, or forbidden Wave 1 product path exists.

The audit inspected:

- the change contract, specification, design, digest-binding amendment, approval, module decision, inert grant amendment, source NO-GO audit, remediation record, and successor evidence;
- exact approved-base, candidate, evidence, and remediation Git objects and all base/candidate/evidence path transitions;
- both SHA-256 manifests, the phase/path policy, support and prototype records, module/import locks, protected-input bindings, and all `43` ownership controls;
- the authoritative `17`-row orchestration ownership table and its local coverage mapping, including distinct requirement and release-allocation source owners;
- all production build-controller source, Makefile/CI/import-check integration, README claim boundaries, and the committed unit/adversarial/process fixtures; and
- the recorded baseline/shadow verification outputs and limitations without rerunning cache-writing commands.

Representative read-only commands were:

```text
git status --porcelain=v2 --branch --untracked-files=all
git show -s --format=... <candidate|evidence|audit-commit>
git ls-tree -r[l] <base|candidate|evidence>
git diff --name-status --stat ee181b7..808fc16
git diff --name-status 808fc16..c0f834f
git diff --check ee181b7..808fc16
git show <commit>:<path> | shasum -a 256
comm -3 <manifest-or-policy path set> <exact Git-object delta path set>
rg -n <contract, source, fixture, effect, and bounded secret-pattern queries>
nl -ba <reviewed source and artifact files>
```

One initial read-only manifest loop used zsh's special `path` variable and invalidated command lookup inside that shell. Its output was discarded and made no mutation. The corrected loop used `file_path`; it reproduced all `34` candidate-manifest rows and all `72` base-manifest rows with zero mismatch.

Per the audit authorization, no `make`, Go test/build, shell harness, hosted CI, or other command that intentionally writes caches or environments was run. Test execution results below are exact recorded same-user local development evidence, not independently re-executed evidence.

## 4. Positive observations and remediation disposition

Direct object/source inspection supports these observations:

- `b92003f` added limited fixed-file reads, bounded skill enumeration, directory/file/byte scan stops, cumulative requirement-expansion caps, pre-append finding caps, and corresponding boundary fixtures.
- `f0a9491` versioned and expanded `harness/support-matrix.tsv` to `19` rows covering one-worktree, plugin authority, local development evidence, release-blocking proof, enforcement withholding, and P0/P1/P2/change-control semantics; `claims.go` now checks the authoritative backlog priority contract.
- `2a58891` added distinct requirement/allocation ownership and source-checks all `17` orchestration ownership classes; the successor has `43` disjoint local controls after the registered remediation record was added.
- `916602a` adds `gate_version=wave-01-v1` and deterministic SHA-256 fields for `12` named gate inputs on the success path.
- `f878890` adds source-summary tampering, trace/claim/scope/manifest/updater/diagnostic/process cases and real isolated broken-package graphs for external detour, harness import, `unsafe`, and forbidden transitive import.
- `808fc16` binds the remediation record, path/ownership controls, and regenerated manifest; its changes are exactly represented in the successor manifest.
- The module transition is consistent across `go.mod`, `harness/modules.lock.tsv`, and the Makefile-derived harness import. The updater remains `reserved`/`UNSET`; GitHub identity remains correctly labeled `USER_ASSERTED`, unpublished, remote-absent, and unsupported.
- Foundation/prototype assets and `scripts/harness/check-foundation-scope.sh` show no base-to-successor change. The grant amendment remains inert and is referenced only as governance/path/ownership data.
- A bounded common credential-pattern scan over the exact candidate returned no match. This is not a comprehensive secret audit.

Original finding disposition after fresh source review:

| Original finding | Successor disposition |
|---|---|
| `AUD-W01-001` — adversarial matrix | **NOT FULLY CLEARED**: the listed functional categories are substantially present, but capped failing-output repeat determinism and end-to-end fixed-input special-node behavior remain unproved/incorrect (`AUD-W01-008`, `AUD-W01-009`). |
| `AUD-W01-002` — resource bounds | **NOT FULLY CLEARED**: byte, line, traversal, expansion, and finding caps improved, but a fixed-input FIFO can block before those repository-shape controls run (`AUD-W01-009`). |
| `AUD-W01-003` — claim boundary | **CLEARED for this exact successor**: record, source binding, and distinct false-claim fixtures are present. |
| `AUD-W01-004` — ownership completeness | **CLEARED for this exact successor**: required local sources and all authoritative orchestration classes are bound and covered. |
| `AUD-W01-005` — success schema | **CLEARED for this exact successor**: version and `12` exact source digests are emitted and fixture-bound. |

## 5. Recorded test evidence

The exact evidence-only child records these post-candidate commands:

| Recorded command | Recorded result | Reproducible test-binary SHA-256 |
|---|---|---|
| `make verify GO_VERSION=1.26.7` | `PASS` | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| `make verify GO_VERSION=1.27.0` | `PASS` shadow development evidence | `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a` |

The record reports zero dependencies; passing controller and two-package import graph; compile/typecheck/unit/adversarial/vet/format/shell success; repeat-build equality; and the exact candidate-manifest/path totals. Source inspection confirms that the committed suite includes the named trace, claim, scope, ownership, import-graph, source-digest, exit-status, environment-non-authority, no-repair, and rendered-output tests. Findings below explain why those passing tests do not prove the complete required behavior.

## 6. Severity model

| Severity | Audit meaning |
|---|---|
| `BLOCKER` | The candidate or evidence cannot be identified or reviewed reliably. |
| `CRITICAL` | A demonstrated path can create catastrophic authority, integrity, or external-effect failure. |
| `HIGH` | A central R3 boundary or explicit checkpoint criterion is materially unproved or bypassable. |
| `MEDIUM` | A required contract is incomplete or unreliable and must be corrected before this checkpoint. |
| `LOW` | A real bounded conformance/maintainability defect that does not independently block the checkpoint. |
| `INFO` | Verified fact, limitation, or future gate with no candidate defect established. |

Finding counts: `BLOCKER 0`, `CRITICAL 0`, `HIGH 0`, `MEDIUM 2`, `LOW 0`, `INFO 2`.

## 7. Findings

### `AUD-W01-008` — `MEDIUM` — Capped failure output is not repeat-deterministic

**Affected acceptance:** deterministic/degraded portions of `W01-AC-006` and mandatory diagnostic/repeat-run fixtures in `W01-AC-007`; consequently `W01-AC-012`.

`report.go` caps collection at `51` findings in `appendFindings` (lines 9–40), but sorting occurs only later in `renderFindings` (lines 56–78). Several validators generate findings directly from Go maps before that cap: for example, `validateTrace` iterates the definition and owner maps (trace.go lines 173–198), and the policy and ownership checks similarly iterate expected/current/rule maps. Go map iteration order is unspecified.

This makes the retained failure subset unstable. A concrete source path is a valid authoritative requirements document paired with an empty ownership map: `validateTrace` has `163` missing-owner subjects, but only the first map-iteration-selected `51` findings survive. Rendering sorts that already truncated subset; separate identical executions can therefore report different requirement subjects even though both return `BLOCKED`.

The new test `TestRenderedFindingsAreOrderedCappedAndRepeatable` supplies one deterministic preordered slice and calls the renderer twice. `TestControllerExitNoRepairEnvironmentIsolationAndRepeatDeterminism` repeats only the successful current candidate. Neither test drives a greater-than-cap failing controller state across separate executions. The evidence claim that cap and repeated-run determinism are proven therefore exceeds the committed failure-path fixture.

This defect does not demonstrate a pass, authority widening, or external effect: the gate remains nonzero. It does violate the explicit deterministic diagnostic contract and makes remediation evidence unreliable when many adversarial findings coexist.

**Required disposition:** establish deterministic ordering before collection truncation (or iterate only sorted keys), retain the cap diagnostic, and add a permanent greater-than-cap failing-controller fixture that compares exact output and nonzero status across separate executions. Rerun both pinned matrices and bind a new candidate/evidence chain.

### `AUD-W01-009` — `MEDIUM` — Fixed inputs are consumed before file-shape validation

**Affected acceptance:** fail-closed, rooted-input, boundedness, and special-node portions of `W01-AC-006` and `W01-AC-007`; consequently `W01-AC-012`.

`readStrictFile` constructs a rooted lexical path and immediately calls `os.Open`, then applies `io.LimitReader` (load.go lines 23–40). It does not first reject a symlink, non-regular node, or multiple-link regular file, and it does not compare the opened handle to a pre-open `Lstat` identity. `main` calls trace and claim checks before `checkPolicy` performs the complete repository scan (main.go lines 37–51). The later scan's `validateFileShape` rejection therefore occurs after fixed inputs have already been opened or consumed.

Consequences are concrete:

- a fixed-input symlink such as `docs/artifacts/requirements.md` can cause bytes outside the repository root to be read (bounded to `maxInputBytes+1`) before the symlink is later reported; and
- replacing a fixed input with a FIFO can block in `os.Open` before `io.LimitReader` or the repository scan runs, so the promised bounded `BLOCKED` result is not reached.

The committed `TestFileShapeRejectsSymlinkSpecialAndHardlink` calls `validateFileShape` as a pure unit test. `TestStrictInputBoundsAreEnforcedBeforeParsing` uses a regular temporary file. Neither exercises the actual fixed-input read order with a symlink, FIFO, or hardlink. The source therefore does not satisfy the specification's rooted-path/special-node rule or the required end-to-end fixture merely because the later inventory scan would reject the node if execution reached it.

No such special node exists in the exact candidate or current clean worktree, and no current pass bypass is demonstrated. The defect is nevertheless material to the first fail-closed scope gate because adversarial repository state can make the controller consume an out-of-root target or hang instead of returning its stable failure.

**Required disposition:** make the fixed-input reader reject non-regular, symlinked, or multiple-link inputs before content consumption; bind the opened handle to the expected rooted file and verify stable pre/post identity and size; and add real fixed-input symlink, FIFO, hardlink, and change-during-read fixtures that complete without hanging and assert the intended stable nonzero rule. Rerun both pinned matrices and bind a new candidate/evidence chain.

### `AUD-W01-010` — `INFO` — Local test evidence was inspected, not rerun

The exact evidence child records both pinned local verification matrices as green with the test-binary digests listed in §5. This reviewer verified the evidence file's bytes, containing Git identity, source/fixture presence, and internal consistency. Execution was intentionally not reproduced because the audit authority prohibited commands that intentionally write caches or environments. These results remain same-user local development evidence.

### `AUD-W01-011` — `INFO` — External qualification and release gates remain separate

Hosted Ubuntu CI, GitHub account/repository/remote authentication, actual Codex/Claude lifecycle and differential conformance, Controlled Client qualification, protected evaluation, grant security review and normative adoption, pilot/stable grants, AP2/AP3, signing, TUF, promotion, publication, release, deployment, exposure, and monitoring remain `NOT_RUN` or absent. The inert grant proposal and this audit clear none of those gates.

## 8. Acceptance disposition

| Acceptance | Independent successor disposition |
|---|---|
| `W01-AC-001` | `PASS` for the exact successor snapshot: 163 source-derived IDs, one recorded owner/allocation each, totals `140/18/5`, and summary-tamper fixtures are present. |
| `W01-AC-002` | `PASS` for the exact successor: the `19`-row support record and validator cover the approved narrow claim boundary. |
| `W01-AC-003` | `PASS`: 12 invocable prototype skills have exactly one disposition and protected bytes reproduce. |
| `W01-AC-004` | `PASS`: current prototype/stable/dual-host/enforcement promotion remains withheld. |
| `W01-AC-005` | `PASS`: P0/P1/P2 and impact/approval semantics are source-bound for the exact backlog bytes. |
| `W01-AC-006` | `NOT CLEARED`: ordinary bounds and stable rules improved, but fixed-input special-node handling can consume outside-root bytes or block, and capped failure diagnostics are nondeterministic (`AUD-W01-008`, `AUD-W01-009`). |
| `W01-AC-007` | `NOT SATISFIED`: the expanded suite lacks an end-to-end fixed-input special-node case and a greater-than-cap repeated failing-controller case (`AUD-W01-008`, `AUD-W01-009`). |
| `W01-AC-008` | `PASS` for local development identity only: module records agree; updater is reserved; external GitHub identity remains `USER_ASSERTED`/`NOT_RUN`. |
| `W01-AC-009` | `PASS` as an inert proposal only; separate security review and normative approval remain `NOT_RUN`. |
| `W01-AC-010` | `PASS` for the exact successor: 43 disjoint controls include requirement/allocation sources and cover all 17 authoritative orchestration classes. |
| `W01-AC-011` | `PASS` as recorded local evidence only; baseline and shadow matrices are recorded green while hosted CI remains `NOT_RUN`. |
| `W01-AC-012` | `FAIL / NO-GO`: identities and independent review are bound, but two unresolved `MEDIUM` findings prevent the R3 development checkpoint. |
| `W01-AC-013` | `PASS` for the exact candidate: no dependency, product/runtime path, prototype edit, unexpected candidate path, external effect, stable claim, or bounded credential-pattern match was observed. |

No aggregate passing count or green recorded suite overrides an unsatisfied acceptance criterion.

## 9. Required next transition

The successor candidate and evidence must remain frozen. A separately authorized remediator may address `AUD-W01-008` and `AUD-W01-009` with finding-specific regression proof. Because either fix changes candidate bytes, remediation requires a new implementation commit/tree, regenerated manifest, complete baseline and shadow local verification, a new evidence-only child binding, and another fresh structurally separate Mode B audit.

No merge, Wave 2 start, deployment, release, publication, remote, hosted action, grant activation, or external action follows from this audit. This Mode B reviewer performed no remediation and does not authorize any next effect.
