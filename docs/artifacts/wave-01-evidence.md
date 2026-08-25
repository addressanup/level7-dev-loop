# Level 7 Dev Loop — Wave 1 Third Remediation Successor Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-004` |
| Artifact type | Local development evidence-only child for the third remediated successor candidate |
| Version | 0.4.0 |
| Date | 2026-08-26 |
| Status | **THIRD SUCCESSOR VERIFIED LOCALLY — FRESH INDEPENDENT AUDIT REQUIRED** |
| Supersedes | `L7-EVD-W01-003` only for third-successor evidence; prior bytes remain at commit `8f82512` |
| Producer | Wave 1 remediator using `level7-dev-loop:l7-release` Mode C under accountable-owner authorization |
| Evidence state | Same-user local development evidence; not independent, hosted, protected, security-qualified, release, or support evidence |
| Third successor candidate commit | `3f6daebee32d16b38497e74235c25f3b6a443fe1` |
| Third successor candidate tree | `77e731b269612cbeee078a25fffde443b8fafbe5` |
| Third successor candidate parent | `47f8e7af4e942964f8a8046864fdcb8dc267ffa1` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md`, SHA-256 `fe8e4d3674a46d4342f239b0473c08803c0f3368a643eaa8bccdee9b21a70a93` |
| Source third NO-GO audit | `docs/artifacts/wave-01-audit.md`, SHA-256 `27b255f6cbdaaf050c7268828f26cd24be9b5120aa38db85ea2fb63fc9288039`; preserved at commit `910c21406b717da13a7bdfe8ac73357b5078b251` |
| Branch | Local `feat/wave-01-build-control`; clean at exact candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED`, remote-absent, and unpublished |
| Next gate | Fresh structurally separate Mode B audit of this exact candidate and evidence-only child |

## 1. Acyclic binding and closure

The reproduced chain is:

```text
approved base ee181b7
  -> original candidate/evidence 4bf4850 / ce62a1f
  -> first NO-GO audit 42ea8ea
  -> first successor candidate/evidence 808fc16 / c0f834f
  -> second NO-GO audit 4a0685a
  -> second successor candidate/evidence 64eee79 / 8f82512
  -> third NO-GO audit 910c214
  -> finding commits b9a48d5 / 47f8e7a
  -> third successor candidate 3f6daeb + self-excluded candidate manifest
  -> this evidence-only child
  -> later fresh independent audit record
```

The third-successor manifest contains 34 bytewise-sorted records. Relative to the approved base, the candidate changes exactly 37 registered paths. The manifest excludes exactly its own path, this evidence path, and the independent audit path, leaving an exact 34-row path and digest match. The current remediation record is included in the manifest.

At candidate commit `3f6daebee32d16b38497e74235c25f3b6a443fe1`:

- tracked inventory: 103 regular Git blobs and 1,259,997 bytes;
- base-to-candidate delta: 37 files, 5,544 insertions, and 16 deletions, with no deleted base path;
- only three `100755` regular script blobs differ from mode `100644`; no symlink or submodule entry exists;
- controller result: 163 requirements, allocations `140/18/5`, 12 prototype skills, 43 ownership controls, 103 files, and 37 changes;
- dependency/product state: no `go.sum`, `vendor/`, product command, updater, product runtime, or grant activation;
- candidate manifest SHA-256: `561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69`; and
- worktree/index state before both exact-candidate verification runs and before this evidence edit: clean.

This evidence file is excluded from the candidate manifest and does not embed its own digest or containing commit. The completion handoff reports the evidence SHA-256 and evidence-only commit/tree/parent. Any candidate or evidence byte change invalidates a later audit.

## 2. Third remediation lineage

| Commit | Purpose | Exact changed paths |
|---|---|---|
| `910c21406b717da13a7bdfe8ac73357b5078b251` | Preserve `L7-AUD-W01-003`, the independent third NO-GO audit, before candidate mutation | `docs/artifacts/wave-01-audit.md` only |
| `b9a48d5f55abbab9eeab1f0a4f1a536351a13a6e` | `fix(audit-AUD-W01-012): harden prototype inventory root` | `claims.go`, `claims_test.go`, `load.go`, `main.go`, `testutil_test.go` |
| `47f8e7af4e942964f8a8046864fdcb8dc267ffa1` | `fix(audit-AUD-W01-013): bound repository enumeration` | `policy.go`, `policy_test.go` |
| `3f6daebee32d16b38497e74235c25f3b6a443fe1` | Bind `L7-REM-W01-003` and the regenerated self-excluded candidate manifest | `wave-01-audit-remediation.md`, `wave-01-candidate.sha256` only |

The source audit's two `MEDIUM` actionable findings are addressed in separate commits with regression proof. This is the remediator's evidence, not independent closure. Every prior NO-GO remains correct for its exact candidate.

## 3. Environment and effects

| Item | Observed value |
|---|---|
| Host | macOS 26.5.2 build 25F84; Darwin 25.5.0; arm64 |
| Baseline compiler | `go version go1.26.7 darwin/arm64` |
| Shadow compiler | `go version go1.27.0 darwin/arm64` |
| Make | GNU Make 3.81 |
| Git | 2.50.1, Apple Git-155 |
| Go policy | `GOENV=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`; repository-scoped telemetry fixed `off` |
| Expected verifier writes | Ignored repository `.cache/go/`, `.cache/repro/`, and existing `.cache/toolchains/` only |
| Other local effect | One unsolicited ignored Finder `.DS_Store` appeared at the repository root during remediation and was moved intact to `/tmp/level7-wave01-dsstore.4p7NA4/root.DS_Store`; it was not deleted or included in either exact verification |
| External effects | No network, remote, hosted workflow, provider, host plugin, publication, release, deployment, or exposure action |

The permanent process fixture executes only the current local test binary in a child process. Filesystem fixtures use test-owned temporary roots. Package-boundary fixtures use pinned local Go binaries, offline module settings, local replacement modules where required, and test-owned cache roots.

## 4. Exact candidate verification

Both commands ran after candidate commit `3f6daebee32d16b38497e74235c25f3b6a443fe1` with a clean worktree.

### 4.1 Baseline — Go 1.26.7

Command: `make verify GO_VERSION=1.26.7`

Result: `PASS` in 11.02 seconds wall time.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da
```

### 4.2 Shadow — Go 1.27.0

Command: `make verify GO_VERSION=1.27.0`

Result: `PASS` as local shadow development evidence in 11.33 seconds wall time.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 4.3 Third-remediation closure checks

| Check | Local evidence | Result |
|---|---|---|
| Inventory root shape | Rooted component/type validation and `O_NONBLOCK`, `O_NOFOLLOW`, `O_DIRECTORY` acquisition precede enumeration | `PASS` |
| Inventory root stability | Opened/path pre/post identity, mode, modification-time, and size checks surround the bounded read | `PASS` |
| Direct degraded fixtures | Real symlink, FIFO, and replacement cases return `BCTL-022`/`BCTL-023` within one second | `PASS` |
| End-to-end degraded fixtures | Full controller path repeats the symlink, FIFO, and replacement cases within one second | `PASS` |
| Pre-enumeration resource ceiling | Rooted `ReadDir(n)` never requests more than 1,027 entries and a larger single-directory fixture returns `SCOPE-346` | `PASS` |
| Scan deadline | Five-second deadline returns `SCOPE-339` and cannot pass or continue directory I/O | `PASS` |
| Existing resource boundaries | 512 directories, 512 files, 8 MiB repository bytes, strict-file bytes/lines, and diagnostic caps remain tested | `PASS` |
| Candidate manifest | Exact 34-row path/digest reproduction with three non-circular exclusions | `PASS` |
| Deletion/whitespace | No base deletion; `git diff --check` | `PASS` |

The prior trace, claim, ownership, scope, import, source-digest, no-repair, environment-isolation, and process suites also remain green. These results are same-user local development evidence and do not self-clear the audit.

## 5. Acceptance and audit disposition

| Acceptance | Remediation evidence disposition |
|---|---|
| `W01-AC-006` | Rooted, fail-closed, count/byte/time-bounded remediation and permanent fixtures are present; independent adequacy decision pending |
| `W01-AC-007` | Required inventory special-node/replacement, end-to-end controller, pre-enumeration ceiling, and deadline fixtures are present; independent adequacy decision pending |
| `W01-AC-011` | Baseline and shadow local matrices green; hosted CI remains `NOT_RUN` |
| `W01-AC-012` | **NOT CLEARED** — fresh independent Mode B audit of the exact third-successor/evidence tuple is mandatory |
| Other Wave 1 acceptance | Prior independent dispositions remain subject to the fresh exact-candidate audit and unchanged later gates |

No aggregate passing result overrides `W01-AC-012`. This writer does not issue `GO`, `CONDITIONAL GO`, release approval, support, security qualification, AP2/AP3, or deployment authority.

## 6. `NOT_RUN` and limitations

- Fresh independent Mode B third-successor audit: `NOT_RUN`.
- Hosted GitHub Actions on Ubuntu 24.04: `NOT_RUN`; configuration inspected locally only.
- GitHub repository/account/remote authentication and publication identity: `NOT_RUN`; no remote exists.
- Actual Codex/Claude package discovery, compatibility, lifecycle, and cross-host conformance: `NOT_RUN`.
- Controlled Client qualification, provider/model/host/platform evaluation, protected holdout, pilot, and stable grant gates: `NOT_RUN` or absent.
- Grant security/boundary review and normative adoption: `NOT_RUN`; the proposal remains inert.
- AP2/AP3, signing, TUF, promotion, release, deployment, exposure, and monitoring: `NOT_RUN` or absent.
- Local Git/filesystem evidence remains same-user mutable and creates no external trust or production claim.
- The reproducibility SHA-256 emitted by `make verify` binds the `internal/harness` test binary. The build-control package is compiled and tested by the matrix, but that printed digest is not a binary digest of `internal/harness/buildcontrol`.
- The five-second scanner checks its deadline before and after bounded filesystem operations; portable local filesystem syscalls do not provide a cancellable hard wall-clock guarantee while a syscall is in progress.

## 7. Fresh-audit handoff

The next structurally separate reviewer receives:

1. candidate commit `3f6daebee32d16b38497e74235c25f3b6a443fe1`, tree `77e731b269612cbeee078a25fffde443b8fafbe5`, parent `47f8e7af4e942964f8a8046864fdcb8dc267ffa1`;
2. candidate manifest SHA-256 `561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69`;
3. remediation record SHA-256 `fe8e4d3674a46d4342f239b0473c08803c0f3368a643eaa8bccdee9b21a70a93`;
4. source audit commit `910c21406b717da13a7bdfe8ac73357b5078b251` and SHA-256 `27b255f6cbdaaf050c7268828f26cd24be9b5120aa38db85ea2fb63fc9288039`;
5. finding commits `b9a48d5f55abbab9eeab1f0a4f1a536351a13a6e` and `47f8e7af4e942964f8a8046864fdcb8dc267ffa1`;
6. baseline reproducible SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` and shadow SHA-256 `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a`; and
7. this evidence file's final SHA-256 plus evidence-only commit/tree/parent from the completion handoff.

The fresh reviewer must reproduce identities and closure, inspect implementation and permanent fixtures, independently decide whether `AUD-W01-012` and `AUD-W01-013` are closed, rescore all findings, evaluate every Wave 1 acceptance criterion, and issue exactly one Mode B verdict. No merge, release, deployment, publication, external action, or later phase is authorized by this evidence.
