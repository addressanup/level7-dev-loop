# Level 7 Dev Loop — Wave 1 Fifth-Successor Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-006` |
| Artifact type | Independent Principal Engineer release audit — `level7-dev-loop:l7-release` Mode B |
| Version | 0.6.0 |
| Date | 2026-08-26 |
| Accountable owner | Anup Pandey |
| Reviewer role | Fresh, structurally separate, read-only Principal Engineer reviewer |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Audited candidate | Commit `af5bcd225c6b3d3911a8f04284dc22984de743f1`; tree `be170459dcd6c081b72ce16255fc67dd63b2f926`; parent `be689bfee83b6ae69e4911c977507de051ec3a0f` |
| Evidence-only child | Commit `b77c4f02a2fcee7af782301699379342e19b7aa3`; tree `cd9584ce4d91c1f3d378d8a6a064dcec13a3d9b1`; parent is the audited candidate |
| Source NO-GO audit | Commit `813eead7e19876d0ebe22745936fa2e107a9bbf0`; SHA-256 `b455cf313a44b96a3da023ae622db4ad17534b982b046070acf3777df5e83f51` |
| Finding remediation commits | `e86378df7af068215f9582f6a40093cc5cd940c7`; `cc465c2fc8f605ec171e17676f3385ee7aa2df91`; formatting-only follow-up `be689bfee83b6ae69e4911c977507de051ec3a0f` |
| Verdict | **GO** |
| Wave 1 checkpoint | `W01-AC-012` is **cleared for this exact development tuple** |

## 1. Decision

The exact approved base, fifth-successor candidate, evidence-only child, ancestry, trees, artifact digests, 35-row candidate manifest, and 38-path base-to-candidate closure all reproduce. The evidence child changes only `docs/artifacts/wave-01-evidence.md`. The preserved failed candidate `b311be5e6d5e6070d238be8f5f118fcb4510626d` remains disclosed history and is not confused with the audited candidate.

Independent source and regression inspection finds `AUD-W01-016`, `AUD-W01-017`, `AUD-W01-020`, and `AUD-W01-021` adequately closed. In particular, aggregate sentinel batches fail before subset sorting or entry inspection; verifier effect roots are physically contained and rejected before the first write when redirected; `t.TempDir` is checked physically; and every permanent `exec.Command` fixture receives the deterministic allowlisted environment.

One `LOW` evidence/regression precision gap remains: the evidence overstates the cache-preparation cases individually exercised by two tests. The implementation's generic preflight covers those roots, and the exact observed roots are physically contained, so this is not a demonstrated bypass or material acceptance failure.

Finding count: `BLOCKER 0`, `CRITICAL 0`, `HIGH 0`, `MEDIUM 0`, `LOW 1`, `INFO 2`.

With exact identity and closure reproduced and no unresolved `BLOCKER`, `CRITICAL`, `HIGH`, or `MEDIUM` finding, `W01-AC-012` clears and the Mode B verdict is **GO**. This verdict is limited to the exact local development tuple. It is not merge, publication, support, security qualification, release, deployment, or external-effect authority.

## 2. Audit method and repository map

The reviewer read the complete `level7-dev-loop:l7-release` skill and repository `AGENTS.md`, selected Mode B only, and mapped the repository before reaching a conclusion. Inspection remained read-only until this registered audit-only artifact was replaced. The initial Git status was clean. No verifier, test, build, preparation script, controller, mutation probe, network action, hosted action, cache/toolchain change, Git index/ref change, or external-system action was run.

Principal read-only commands included:

```text
git --no-optional-locks status --short
git --no-optional-locks rev-parse <commit>^{tree} <commit>^ <commit>^1
git --no-optional-locks show <commit>:<path> | shasum -a 256
git --no-optional-locks diff --name-status <base> <candidate>
git --no-optional-locks diff --numstat <candidate> <evidence-child>
git --no-optional-locks ls-tree -r -l <commit>
git --no-optional-locks diff --check <base> <candidate>
git --no-optional-locks worktree list --porcelain
git --no-optional-locks remote -v
find <repository-cache-roots> -maxdepth 0 -exec stat ...
realpath <repository-cache-and-toolchain-roots>
rg --files
rg -n 'exec\.Command|Cmd\.Env|processFixtureEnvironment' internal/harness
```

Independent read-only parsing checked manifest ordering, uniqueness, row hashes, set closure, requirement ownership, allocation totals, prototype/support inventories, and change classes directly from committed bytes.

| Area | Audited map |
|---|---|
| Root and controls | `Makefile`, module/version controls, host/plugin manifests, workflow, status, license, ignore rules, and agent instructions |
| Documentation | 46 committed candidate documentation files, including specification, original design, successor amendment, acceptance/change records, source-audit lineage, remediation, candidate manifest, and evidence exclusions |
| Harness data | 12 committed files under `harness/`, including exact phase/path/module/import/ownership/support/prototype control data |
| Go implementation | 16 files under `internal/harness/`; packages `harness` and `buildcontrol` only |
| Skills | 12 protected prototype skill packages; no candidate edit to their governed bytes |
| Scripts and reference | Four harness scripts, including `prepare-cache.sh`, plus the protected workflow reference |
| Candidate inventory | 104 regular Git blobs; 1,284,743 bytes; 100 mode `100644`, 4 mode `100755`; no symlink, special-node, or submodule entry |
| Base delta | Exactly 38 paths: 32 added, 6 modified, no deletion; 6,055 insertions, 19 deletions; `git diff --check` clean |
| Product/dependency scope | No `go.sum`, `vendor/`, product command, updater implementation, runtime product package, provider action, grant activation, or new production dependency |
| Repository topology | One worktree on the local feature branch and no configured remote |

## 3. Identity, digest, manifest, and closure reproduction

| Evidence | Independent result |
|---|---|
| Approved base | Commit/tree exactly reproduced: `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4` / `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate | Commit/tree/parent exactly reproduced: `af5bcd225c6b3d3911a8f04284dc22984de743f1` / `be170459dcd6c081b72ce16255fc67dd63b2f926` / `be689bfee83b6ae69e4911c977507de051ec3a0f` |
| Evidence child | Commit/tree/parent exactly reproduced: `b77c4f02a2fcee7af782301699379342e19b7aa3` / `cd9584ce4d91c1f3d378d8a6a064dcec13a3d9b1` / candidate |
| Evidence-only delta | Exactly one modified path, `docs/artifacts/wave-01-evidence.md`; numstat `78 67` |
| Evidence SHA-256 | `1d350436398fad8f53a6221fc1c1f2e64ac9bfa0f1b8c5317f1003c1a198b98c` |
| Candidate manifest SHA-256 | `eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84` |
| Remediation record SHA-256 | `23e5ad6491f4e64e0e429ffc29156a8dbd9fbe9f8f55346f43b04197ebf7847d` |
| Successor design amendment SHA-256 | `7162d7a05117374c0994f9a721e9930f0b27ec8527ccd51352a8749bf7119b67` |
| Source audit | Commit `813eead7e19876d0ebe22745936fa2e107a9bbf0`; artifact SHA-256 `b455cf313a44b96a3da023ae622db4ad17534b982b046070acf3777df5e83f51` |
| Physical-containment remediation | Commit `e86378df7af068215f9582f6a40093cc5cd940c7`; `AUD-W01-020` finding-specific change |
| Process-isolation remediation | Commit `cc465c2fc8f605ec171e17676f3385ee7aa2df91`; `AUD-W01-021` finding-specific change |
| Failed bind and follow-up | Failed candidate `b311be5e6d5e6070d238be8f5f118fcb4510626d` remains in history; `be689bfee83b6ae69e4911c977507de051ec3a0f` changes only `policy.go` formatting before the final bind |

`docs/artifacts/wave-01-candidate.sha256` contains one header plus exactly 35 bytewise-sorted, unique data rows. Every recorded SHA-256 matches the corresponding blob read from the exact candidate. Removing exactly the three circular/audit exclusions — the manifest itself, `docs/artifacts/wave-01-evidence.md`, and `docs/artifacts/wave-01-audit.md` — from the 38-path closure produces exactly those 35 paths.

`harness/wave-01-paths.tsv` contains exactly 38 data rows. Its path set equals the base-to-candidate Git diff. Every declared change class matches Git: 32 `A`, 6 `M`, and no deletion. The approved-base manifest contains exactly 72 rows, every digest reproduces, and its path set covers the base tree exactly. The 26-row protected foundation manifest also reproduces against the candidate.

The candidate controller inputs independently parse to exactly 163 normative definitions, 163 unique accountable owners, no missing, duplicate, unknown, or multiply owned ID, and allocations `140/18/5`. The support matrix has 19 rows, all 12 invocable prototype skills have exactly one disposition, and all 43 ownership controls are present and uniquely assigned.

## 4. Contract, architecture, and effect-boundary evaluation

The audited authority chain comprises the specification, original design, successor design amendment, change contract/acceptance criteria, source NO-GO audit, remediation record, implementation, Makefile/script boundary, permanent regressions, candidate manifest, path closure, and evidence-only child.

The specification defines deterministic, offline, fail-closed control behavior (`docs/artifacts/wave-01-specification.md:116-155`), treats repository content as untrusted and requires rooted/no-escape handling (`:214-220`), and makes passing aggregates unable to waive a material finding (`:285-306`). The original design supplies the source-derived controller and no-follow repository policy. The approved successor amendment requires aggregate rejection before sorting/inspection (`docs/artifacts/wave-01-design-amendment.md:58-82`), physical pre-write containment (`:84-86`), and deterministic allowlisted process fixtures (`:88-90`). The evidence records the exact candidate and local matrices but expressly remains same-user development evidence (`docs/artifacts/wave-01-evidence.md:9-23,88-146,161-175`).

The implementation remains appropriately inert. It adds no product interface, updater, dependency, provider/network operation, production subprocess, mutable grant, support claim, or release path. The module decision and grant-ladder amendment remain documentary controls only. No protected skill, plugin, marketplace, workflow reference, or foundation-manifest byte changed.

The Makefile exports repository-local `TMPDIR`, `GOTMPDIR`, Go cache, telemetry, reproduction, and toolchain paths and delegates the first preparation effect to the cache script (`Makefile:55-60,84-85`). `scripts/harness/prepare-cache.sh` resolves the canonical repository root (`:24-27`), declares the complete component inventory (`:29-40`), preflights physical type and containment for `.cache`, Go path/bin/build/module/temp/telemetry, reproduction, toolchain root, selected toolchain, and telemetry-mode paths before any write (`:42-74`), begins directory creation only afterward (`:76-91`), and replaces telemetry mode only after validation (`:93-105`).

Read-only inspection found `.cache`, `.cache/go` and its declared components, `.cache/repro`, `.cache/toolchains`, both selected toolchains, and telemetry state as ordinary repository-contained filesystem objects. The telemetry mode file contains the expected `off 2026-08-24`. These observations corroborate, but do not substitute for, the implementation invariant.

## 5. Inherited finding disposition

### `AUD-W01-016` — adequately closed

`internal/harness/buildcontrol/policy.go:227-246` detects a sentinel-filled bounded directory batch and returns directory-scoped `SCOPE-338`. Sorting begins only at line 248 and entry iteration/`Info()` only at line 265. Thus filesystem subset order cannot select a file- or directory-specific subject before the aggregate rejection.

Permanent regressions independently cover the required dimensions:

- `TestRepositoryScanCapsSingleDirectoryReadBeforeEnumeration` creates oversized real directories in ascending and descending order and requires the exact same complete result (`internal/harness/buildcontrol/policy_test.go:200-239`).
- `TestRepositoryScanMixedEntryBatchIsOrderIndependent` supplies file-omitted and directory-omitted capped subsets and uses entries that fail if `Info()` is called (`policy_test.go:241-271`).
- `TestRepositoryEntryBatchFailureIsStableAcrossProcesses` requires separate child processes to exit 1 with byte-identical output (`policy_test.go:273-305`).
- `syntheticRepositoryBatch` constructs both mixed orders (`policy_test.go:307-343`).

The exact source ordering proves failure before sorting; the mixed batch proves failure before entry inspection. The real-filesystem and cross-process tests prove the rendered rule, subject, message, recovery, and exit behavior are deterministic.

### `AUD-W01-017` — adequately closed

The approved amendment now truthfully and narrowly defines real filesystem/process fixtures as repository-scoped verifier effects and preserves the no-product/no-provider boundary (`docs/artifacts/wave-01-design-amendment.md:58-90`). The evidence and implementation match that model: repository-local ignored effects are declared, preparation is fail-before-write, process inputs are allowlisted, and later hosted/external gates remain `NOT_RUN`. This supersedes the prior documentary mismatch without widening production authority.

### `AUD-W01-020` — adequately closed

Physical containment is enforced by the preparation boundary, not inferred from lexical prefixes. `prepare-cache.sh:42-74` applies the same no-symlink, expected-type, realpath-below-root preflight to every declared existing component before line 76 can write. It explicitly includes the toolchain root and selected toolchain. Parent-first post-creation checks and telemetry-mode replacement preserve the invariant.

`TestTemporaryRootsAreRepositoryScoped` uses `Lstat`, `EvalSymlinks`, and `filepath.Rel` and creates `t.TempDir()` only after proving the configured root is physically below the repository (`internal/harness/testutil_test.go:26-63`). `TestPrepareCacheCreatesPhysicalRepositoryDirectories` and `TestPrepareCacheRejectsRedirectedComponentsBeforeWriting` exercise safe preparation and redirect rejection, including unchanged external targets and absence of a later sibling after failure (`testutil_test.go:74-188`). The exact live roots also resolve physically below the repository. Same-user races remain a declared local-evidence limitation, not an unacknowledged claim.

### `AUD-W01-021` — adequately closed

The candidate has seven permanent `exec.Command` call sites. Every one assigns `Cmd.Env` from the single deterministic `processFixtureEnvironment` helper before execution:

1. repository preparation shell: `internal/harness/testutil_test.go:68-69`;
2. current test binary for capped-trace behavior: `testutil_test.go:429-430`;
3. pinned Go building the test-owned controller: `testutil_test.go:492-494`;
4. controller execution: `testutil_test.go:499-501`;
5. controller argument behavior: `testutil_test.go:513-515`;
6. current test binary for enumeration output: `internal/harness/buildcontrol/policy_test.go:289-290`; and
7. fixed shell/import fixture using pinned Go: `policy_test.go:487-494`.

The helper, fixed offline controls, explicit repository-path allowlist, override precedence, and reordered/secret-shaped ambient-environment regression are at `internal/harness/testutil_test.go:199-297`. The child path is fixed to `/usr/bin:/bin`; home, credential, provider, proxy, and unrelated ambient values are excluded; exact helper overrides win. Repository search found no production `os/exec` site.

## 6. Recorded exact-candidate verification

The audit did not rerun write-producing commands. It inspected the committed evidence and existing exact-candidate outputs under the read-only Mode B authority.

| Recorded command/evidence | Result |
|---|---|
| `/usr/bin/time -p make verify GO_VERSION=1.26.7` | `PASS`, exit 0, real `12.40s`; reproducible internal/harness test-binary SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| `/usr/bin/time -p make verify GO_VERSION=1.27.0` | `PASS`, exit 0, real `12.39s`; reproducible internal/harness test-binary SHA-256 `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a` |

The evidence binds both matrices to candidate `af5bcd225c6b3d3911a8f04284dc22984de743f1` and manifest digest `eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84` (`docs/artifacts/wave-01-evidence.md:88-146`). Existing `.cache/repro` artifacts independently contain byte-identical `harness-a.test`/`harness-b.test` pairs with the recorded SHA for each toolchain.

Passing tests alone did not control the decision. The verdict also rests on source ordering, complete command-site inventory, pre-write effect flow, physical path checks, committed identity/digest closure, and the limitations below.

## 7. Severity model and findings

| Severity | Meaning used here |
|---|---|
| `BLOCKER` | Candidate identity/authority is invalid or the release cannot be evaluated safely |
| `CRITICAL` | Immediate catastrophic integrity, security, or irreversible-effect exposure |
| `HIGH` | Major release-control failure with broad or likely impact |
| `MEDIUM` | Material contract/correctness failure that blocks the Wave 1 audit threshold |
| `LOW` | Real but non-material weakness for this checkpoint |
| `INFO` | Evidence boundary or later gate, not a candidate defect |

### `AUD-W01-024` — `LOW` — Preparation evidence overstates per-path regression coverage

`docs/artifacts/wave-01-evidence.md:132` says the safe-preparation test resolves every created cache/temp/telemetry/toolchain directory, but `TestPrepareCacheCreatesPhysicalRepositoryDirectories` lists paths only through `.cache/repro` (`internal/harness/testutil_test.go:80-95`); it does not individually assert `.cache/toolchains` or the selected toolchain. Evidence line 133 says the redirect test covers the selected toolchain, while its case table (`testutil_test.go:111-157`) covers `.cache`, Go temp, telemetry mode, and toolchain root but not the selected toolchain; it also does not isolate separate Go build-cache and telemetry-directory redirect cases.

The generic implementation preflights all of those omitted components before its first write, the selected-toolchain logic is directly inspectable, the exact local roots are physically contained, and no bypass was found. The gap is therefore evidence precision and future regression granularity, not a material defect. A later evidence maintenance change should narrow the claim or add explicit cases; this audit does not authorize that mutation.

### `AUD-W01-025` — `INFO` — Local verification was inspected, not rerun

Mode B's mutation constraint prohibited rerunning cache/temp-writing verification. Existing results are same-user mutable local evidence. The reported reproducibility SHA binds the `internal/harness` test binary, not every verifier/controller executable. Filesystem syscalls are not cancellable hard wall-clock boundaries, and same-user races cannot be eliminated by preflight alone. The artifact makes no stronger claim.

### `AUD-W01-026` — `INFO` — Hosted and later release gates remain outside this audit

Hosted CI, actual host-package compatibility, protected evaluation, grant security review/adoption, AP2/AP3, signing, promotion, merge, publication, release, deployment, exposure, and monitoring remain `NOT_RUN`, absent, or separately gated. This GO does not clear them.

## 8. Wave 1 acceptance criteria

| Criterion | Independent result |
|---|---|
| `W01-AC-001` | `PASS` — source parsing reproduces exactly 163 normative IDs, one unique accountable owner/allocation per ID, and totals `140/18/5` |
| `W01-AC-002` | `PASS` — the 19-row support matrix preserves the two-product distinction, A0–A2/A3–A5 boundary, and all three proof profiles without promotion |
| `W01-AC-003` | `PASS` — all 12 prototype skills have exactly one disposition, and every byte in the 26-row protected foundation manifest reproduces |
| `W01-AC-004` | `PASS` — stable version, compatibility, enforcement, and support-promotion claims remain withheld; current claims remain development/prototype scoped |
| `W01-AC-005` | `PASS` — priority/scope changes remain gated by impact evidence and accountable approval; no date or metric acts as a safety waiver |
| `W01-AC-006` | `PASS` — the phase-aware successor preserves historical Step 5 evidence and fails closed on unknown, malformed, stale, unowned, and over-bound input; sentinel batches reject before sorting/inspection |
| `W01-AC-007` | `PASS` — permanent positive/adversarial fixtures cover the governed boundary, real-filesystem order, mixed subsets, cross-process output, physical containment, and process isolation. `AUD-W01-024` records a non-material per-path coverage/wording gap |
| `W01-AC-008` | `PASS` — the exact root-module decision precedes product imports, no product import exists, and updater identity remains reserved |
| `W01-AC-009` | `PASS` — the grant-ladder amendment remains inert, non-interchangeable, separately auditable, and separately approvable |
| `W01-AC-010` | `PASS` — 43 ownership controls are complete, unique, and disjoint for governed classes and exclude candidate authority over protected assets |
| `W01-AC-011` | `PASS` — exact-candidate baseline and shadow local matrices are recorded green; cache/temp/toolchain roots inspected for this audit are physically repository-contained; effects and limitations are recorded; hosted CI remains truthfully `NOT_RUN` |
| `W01-AC-012` | `PASS` — exact candidate/evidence identities, ancestry, artifact digests, 35-row manifest, and 38-path closure reproduce, and no unresolved `BLOCKER`, `CRITICAL`, `HIGH`, or `MEDIUM` finding remains |
| `W01-AC-013` | `PASS` — registered closure, no dependency/product/prototype/grant activation, offline effect boundary, pre-write physical containment, and deterministic process environment all reproduce |

No aggregate test result waives a material finding. The one unresolved `LOW` and two `INFO` findings do not meet the approved `W01-AC-012` blocking threshold.

## 9. Authority boundary

The source NO-GO audit remains byte-preserved at commit `813eead7e19876d0ebe22745936fa2e107a9bbf0`; its conclusions remain correct for its exact candidate. This audit independently decides only the exact fifth-successor tuple named above. It modifies no candidate code, test, configuration, evidence, manifest, design, protected historical record, Git index/ref, cache, toolchain, remote, or external system. It authorizes no merge, publication, support claim, security qualification, release, deployment, exposure, remediation, or later-phase action.

The audit file intentionally does not embed its own SHA-256. The completion handoff records that digest after the sole authorized write and verifies that `docs/artifacts/wave-01-audit.md` is the only changed path.
