# Level 7 Dev Loop — Wave 1 Fifth Remediation Successor Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-006` |
| Artifact type | Local development evidence-only child for the fifth remediated successor candidate |
| Version | 0.6.0 |
| Date | 2026-08-26 |
| Status | **FIFTH SUCCESSOR VERIFIED LOCALLY — FRESH INDEPENDENT AUDIT REQUIRED** |
| Supersedes | `L7-EVD-W01-005` only for fifth-successor evidence; prior bytes remain at commit `06345424e455a57b06c183b38a8492d20580c2bf` |
| Producer | Wave 1 remediator using `level7-dev-loop:l7-release` Mode C under accountable-owner authorization |
| Evidence state | Same-user local development evidence; not independent, hosted, protected, security-qualified, release, or support evidence |
| Fifth successor candidate commit | `af5bcd225c6b3d3911a8f04284dc22984de743f1` |
| Fifth successor candidate tree | `be170459dcd6c081b72ce16255fc67dd63b2f926` |
| Fifth successor candidate parent | `be689bfee83b6ae69e4911c977507de051ec3a0f` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md`, SHA-256 `23e5ad6491f4e64e0e429ffc29156a8dbd9fbe9f8f55346f43b04197ebf7847d` |
| Design amendment | `docs/artifacts/wave-01-design-amendment.md`, SHA-256 `7162d7a05117374c0994f9a721e9930f0b27ec8527ccd51352a8749bf7119b67` |
| Source NO-GO audit | `L7-AUD-W01-005`, SHA-256 `b455cf313a44b96a3da023ae622db4ad17534b982b046070acf3777df5e83f51`; preserved at commit `813eead7e19876d0ebe22745936fa2e107a9bbf0` |
| Branch | Local `feat/wave-01-build-control`; clean at exact-candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED`, remote-absent, and unpublished |
| Next gate | Fresh structurally separate Mode B audit of this exact candidate and evidence-only child |

## 1. Acyclic binding and closure

The reproduced chain is:

```text
approved base ee181b7
  -> original through fourth successor histories
  -> fourth successor candidate/evidence a1f146c / 0634542
  -> independent NO-GO audit 813eead
  -> finding commits e86378d / cc465c2
  -> failed candidate bind b311be5
  -> formatting-only finding follow-up be689bf
  -> fifth successor candidate af5bcd2 + self-excluded candidate manifest
  -> this evidence-only child
  -> later fresh independent audit record
```

The fifth-successor manifest contains 35 bytewise-sorted data records. Relative to the approved base, the candidate changes exactly 38 registered paths. The manifest excludes exactly its own path, this evidence path, and the independent audit path, leaving an exact 35-row path and digest match. The current remediation record and successor design amendment are included in the manifest.

At candidate commit `af5bcd225c6b3d3911a8f04284dc22984de743f1`:

- tracked inventory: 104 regular Git blobs and 1,284,743 bytes;
- base-to-candidate delta: 38 files, 6,055 insertions, and 19 deletions; 32 paths added, 6 modified, and no base path deleted;
- modes: 100 `100644` and 4 `100755` regular blobs; no symlink or submodule entry exists;
- controller result: 163 requirements, allocations `140/18/5`, 12 prototype skills, 43 ownership controls, 104 files, and 38 changes;
- dependency/product state: no `go.sum`, `vendor/`, product command, updater, product runtime, or grant activation;
- candidate manifest SHA-256: `eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84`; and
- worktree/index state before both exact-candidate verification runs and before this evidence edit: clean.

This evidence file is excluded from the candidate manifest and does not embed its own digest or containing commit. The completion handoff reports the evidence SHA-256 and evidence-only commit/tree/parent. Any candidate or evidence byte change invalidates a later audit.

## 2. Fifth remediation lineage

| Commit | Purpose | Exact changed paths |
|---|---|---|
| `813eead7e19876d0ebe22745936fa2e107a9bbf0` | Preserve source `L7-AUD-W01-005` before candidate mutation | `docs/artifacts/wave-01-audit.md` only |
| `e86378df7af068215f9582f6a40093cc5cd940c7` | `fix(audit-AUD-W01-020): contain verifier effects` | `Makefile`, design amendment, path policy, `policy.go`, `testutil_test.go`, and new `scripts/harness/prepare-cache.sh` |
| `cc465c2fc8f605ec171e17676f3385ee7aa2df91` | `fix(audit-AUD-W01-021): isolate process fixtures` | design amendment, `policy.go`, `policy_test.go`, and `testutil_test.go` |
| `b311be5e6d5e6070d238be8f5f118fcb4510626d` | First candidate bind, preserved after baseline verification failed at `format-check` | remediation record and candidate manifest only |
| `be689bfee83b6ae69e4911c977507de051ec3a0f` | `fix(audit-AUD-W01-021): format policy binding` | `policy.go` only; one `gofmt` alignment change |
| `af5bcd225c6b3d3911a8f04284dc22984de743f1` | Rebind `L7-REM-W01-005` and the regenerated self-excluded candidate manifest | remediation record and candidate manifest only |

The source audit's `MEDIUM` `AUD-W01-020` and `LOW` `AUD-W01-021` findings are addressed with finding-specific commits and permanent regression proof. The failed `b311be5` candidate and its exit-2 baseline result remain visible in the lineage; no failed result is relabeled as final evidence. The candidate retains the exact registered 38-path closure. This is the remediator's evidence, not independent closure, and every prior audit remains correct for its exact candidate.

## 3. Environment and effects

| Item | Observed value |
|---|---|
| Verification timestamp | Baseline and shadow runs completed on 2026-08-26 before evidence authoring; handoff inventory observed at `2026-08-26T05:16:28Z` |
| Host | macOS 26.5.2 build 25F84; Darwin 25.5.0; arm64 |
| Baseline compiler | `go version go1.26.7 darwin/arm64` |
| Shadow compiler | `go version go1.27.0 darwin/arm64` |
| Make | GNU Make 3.81 |
| Git | 2.50.1, Apple Git-155 |
| Go policy | `GOENV=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`; repository-scoped telemetry fixed `off` |
| Temporary-root policy | `TMPDIR` equals `GOTMPDIR`, both fixed to ignored repository path `.cache/go/tmp`; permanent tests resolve physical paths and prove `t.TempDir()` containment |
| Preparation policy | The first verifier write is delegated to `scripts/harness/prepare-cache.sh`, which preflights every existing effect/toolchain component and then creates and rechecks repository-local directories parent-first |
| Expected verifier writes | Ignored repository `.cache/go/`, `.cache/repro/`, `.cache/toolchains/`, and `.cache/telemetry/` only |
| Other local effects | No non-cache repository or ambient host file effect observed in this remediation; no cleanup of ambient files |
| External effects | No network, remote, hosted workflow, provider, host plugin, publication, release, deployment, or exposure action |

Real filesystem fixtures create only test-owned roots under repository `.cache/go/tmp` and exercise symlink, FIFO, replacement, enumeration, and physical-containment behavior. The fixed process inventory is the current local test binary, a pinned repository-local Go binary, fixed `/bin/sh` for repository-owned harness scripts, and a test-owned controller built by pinned Go. Each `exec.Command` fixture sets `Cmd.Env` from deterministic fixed offline controls plus an explicit repository-path allowlist and exact helper overrides. Ambient home, credential, provider, proxy, and unrelated values are excluded; child `PATH` is `/usr/bin:/bin`.

## 4. Exact candidate verification

Both complete commands ran after candidate commit `af5bcd225c6b3d3911a8f04284dc22984de743f1` with a clean worktree.

### 4.1 Baseline — Go 1.26.7

Command: `/usr/bin/time -p make verify GO_VERSION=1.26.7`

Result: `PASS`; process exit 0; `real 12.40`, `user 22.54`, `sys 9.66` seconds.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=104; changed=38
check-import-boundaries: PASS (2 package set)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da
```

### 4.2 Shadow — Go 1.27.0

Command: `/usr/bin/time -p make verify GO_VERSION=1.27.0`

Result: `PASS` as local shadow development evidence; process exit 0; `real 12.39`, `user 23.78`, `sys 10.44` seconds.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=104; changed=38
check-import-boundaries: PASS (2 package set)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 4.3 Finding-specific and closure checks

| Check | Local evidence | Result |
|---|---|---|
| Physical safe preparation | `TestPrepareCacheCreatesPhysicalRepositoryDirectories` resolves every created cache/temp/telemetry/toolchain directory to its exact repository path and checks telemetry mode | `PASS` under both pinned toolchains |
| Redirect rejection before writes | `TestPrepareCacheRejectsRedirectedComponentsBeforeWriting` covers redirected `.cache`, temp, telemetry-mode, toolchain root, and selected toolchain; external targets remain unchanged and no later cache sibling is created | `PASS` under both pinned toolchains |
| Temporary-root containment | Strengthened `TestTemporaryRootsAreRepositoryScoped` uses `Lstat`, `EvalSymlinks`, and `filepath.Rel` rather than lexical prefix comparison | `PASS` under both pinned toolchains |
| Environment isolation | `TestProcessFixtureEnvironmentIsAllowlistedAndDeterministic` proves reordered ambient inputs render identically, excludes home/secret-shaped variables, prevents ambient proxy override, fixes `PATH`, and preserves exact override precedence | `PASS` under both pinned toolchains |
| Process inventory | Every permanent `exec.Command` site assigns the allowlisted environment; import, controller, current-test-binary, pinned-Go, and repository-script fixtures remain green | `PASS` |
| Stable aggregate bound | Sentinel-filled repository batches return directory-scoped `SCOPE-338` before sorting or entry inspection | `PASS` |
| Real filesystem order | Oversized all-file roots created ascending and descending render byte-identical rule, subject, message, and recovery | `PASS` |
| Mixed subset order | Omitted-directory and omitted-file capped batches render the same complete result without calling `DirEntry.Info` | `PASS` |
| Cross-process result | Both mixed variants exit 1 with byte-identical `BLOCKED` output | `PASS` |
| Candidate manifest | Exact 35-row path/digest reproduction with three non-circular exclusions; exact 38-path path-policy closure | `PASS` |
| Deletion/whitespace | 32 added and 6 modified paths, no base deletion; `git diff --check` | `PASS` |

Before final candidate binding, the finding-specific physical-containment selection passed in `0.542s` / `0.544s` and the process-environment selection passed in `0.973s` / `0.954s` under Go 1.26.7 / Go 1.27.0 respectively. The complete candidate-bound matrices above independently reran the whole package suites.

The prior trace, claim, ownership, inventory-root, file-identity, scope, import, source-digest, no-repair, deadline, and process suites also remain green. These results are same-user local development evidence and do not self-clear the audit.

## 5. Acceptance and audit disposition

| Acceptance | Remediation evidence disposition |
|---|---|
| `W01-AC-006` | Stable fail-closed enumeration and aggregate/count/byte/time-bound regressions remain green; independent adequacy decision pending |
| `W01-AC-007` | Permanent real/injected/process-order, physical-containment, and environment-isolation regressions are present and green; independent adequacy decision pending |
| `W01-AC-011` | Baseline and shadow local matrices are green with physically repository-scoped effects; hosted CI remains `NOT_RUN` |
| `W01-AC-012` | **NOT CLEARED** — fresh independent Mode B audit of the exact fifth-successor/evidence tuple is mandatory |
| `W01-AC-013` | Registered 38-path closure, no dependency/product/grant activation, and offline effect boundary reproduce locally; independent decision pending |
| Other Wave 1 acceptance | Prior independent dispositions remain subject to the fresh exact-candidate audit and unchanged later gates |

No aggregate passing result overrides `W01-AC-012`. This writer does not issue a release verdict, release approval, support, security qualification, AP2/AP3, or deployment authority.

## 6. `NOT_RUN` and limitations

- Fresh independent Mode B fifth-successor audit: `NOT_RUN`.
- The first frozen candidate bind `b311be5e6d5e6070d238be8f5f118fcb4510626d` failed its baseline matrix with exit 2 after build-control and import checks because `policy.go` was not `gofmt`-formatted. It remains immutable failed evidence and is not the verified candidate.
- Hosted GitHub Actions on Ubuntu 24.04: `NOT_RUN`; configuration inspected locally only.
- GitHub repository/account/remote authentication and publication identity: `NOT_RUN`; no remote exists.
- Actual Codex/Claude package discovery, compatibility, lifecycle, and cross-host conformance: `NOT_RUN`.
- Controlled Client qualification, provider/model/host/platform evaluation, protected holdout, pilot, and stable grant gates: `NOT_RUN` or absent.
- Grant security/boundary review and normative adoption: `NOT_RUN`; the proposal remains inert.
- AP2/AP3, signing, TUF, promotion, release, deployment, exposure, and monitoring: `NOT_RUN` or absent.
- Local Git/filesystem evidence remains same-user mutable and creates no external trust or production claim.
- The reproducibility SHA-256 emitted by `make verify` binds the `internal/harness` test binary. The build-control package is compiled and tested by the matrix, but that printed digest is not a binary digest of `internal/harness/buildcontrol`.
- The five-second scanner checks its deadline before and after bounded filesystem operations; portable local filesystem syscalls do not provide a cancellable hard wall-clock guarantee while a syscall is in progress.
- Physical containment rejects symlinked existing components but cannot eliminate same-user races or mutable local evidence; a fresh read-only audit must inspect that residual boundary independently.

## 7. Fresh-audit handoff

The next structurally separate reviewer receives:

1. candidate commit `af5bcd225c6b3d3911a8f04284dc22984de743f1`, tree `be170459dcd6c081b72ce16255fc67dd63b2f926`, parent `be689bfee83b6ae69e4911c977507de051ec3a0f`;
2. candidate manifest SHA-256 `eb88b29cd0c7543fb2f4a8945d1f8e369fc95e8e2a2c067af9f28bf0f3e79c84`;
3. remediation record SHA-256 `23e5ad6491f4e64e0e429ffc29156a8dbd9fbe9f8f55346f43b04197ebf7847d` and design-amendment SHA-256 `7162d7a05117374c0994f9a721e9930f0b27ec8527ccd51352a8749bf7119b67`;
4. source audit commit `813eead7e19876d0ebe22745936fa2e107a9bbf0` and SHA-256 `b455cf313a44b96a3da023ae622db4ad17534b982b046070acf3777df5e83f51`;
5. finding commits `e86378df7af068215f9582f6a40093cc5cd940c7` and `cc465c2fc8f605ec171e17676f3385ee7aa2df91`, formatting-only follow-up `be689bfee83b6ae69e4911c977507de051ec3a0f`, and preserved failed candidate `b311be5e6d5e6070d238be8f5f118fcb4510626d`;
6. baseline reproducible SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` and shadow SHA-256 `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a`; and
7. this evidence file's final SHA-256 plus evidence-only commit/tree/parent from the completion handoff.

The fresh reviewer must reproduce identities, the 35-row manifest, and the 38-path closure; inspect source audit, remediation, implementation, and permanent regressions; independently decide whether `AUD-W01-020` and `AUD-W01-021` are adequately closed and whether prior `AUD-W01-016`/`AUD-W01-017` closure remains valid; rescore all findings; evaluate every Wave 1 acceptance criterion; and issue exactly one Mode B verdict. No merge, release, deployment, publication, external action, or later phase is authorized by this evidence.
