# Level 7 Dev Loop — Wave 1 Fourth Remediation Successor Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-005` |
| Artifact type | Local development evidence-only child for the fourth remediated successor candidate |
| Version | 0.5.0 |
| Date | 2026-08-26 |
| Status | **FOURTH SUCCESSOR VERIFIED LOCALLY — FRESH INDEPENDENT AUDIT REQUIRED** |
| Supersedes | `L7-EVD-W01-004` only for fourth-successor evidence; prior bytes remain at commit `cca9859` |
| Producer | Wave 1 remediator using `level7-dev-loop:l7-release` Mode C under accountable-owner authorization |
| Evidence state | Same-user local development evidence; not independent, hosted, protected, security-qualified, release, or support evidence |
| Fourth successor candidate commit | `a1f146cdb5b2f20a7852bcf490223541fe4c8986` |
| Fourth successor candidate tree | `db580f77234dc14289f22174760d6da9bf442891` |
| Fourth successor candidate parent | `4b092a65e74d713346975de5bb4d78d161ad2b0a` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md`, SHA-256 `e147c48f204118172e36d9e234e424516f7f1e06c638d4c3fac65dec0f08293b` |
| Design amendment | `docs/artifacts/wave-01-design-amendment.md`, SHA-256 `e10378f598098d5db8e9f20177324e917260e0ce016453903ac0159485526470` |
| Source audit | `L7-AUD-W01-004`, SHA-256 `16f11e11b466a78cb7bf758cff40b7e0f7e85057e73af217bb69649795003917`; preserved at commit `62e1a019eb6a75748c628c93102c41db81166d28` |
| Branch | Local `feat/wave-01-build-control`; clean at exact candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED`, remote-absent, and unpublished |
| Next gate | Fresh structurally separate Mode B audit of this exact candidate and evidence-only child |

## 1. Acyclic binding and closure

The reproduced chain is:

```text
approved base ee181b7
  -> original through third successor histories
  -> third successor candidate/evidence 3f6daeb / cca9859
  -> independent audit 62e1a01
  -> finding commits 6c04c53 / 4b092a6
  -> fourth successor candidate a1f146c + self-excluded candidate manifest
  -> this evidence-only child
  -> later fresh independent audit record
```

The fourth-successor manifest contains 34 bytewise-sorted data records. Relative to the approved base, the candidate changes exactly 37 registered paths. The manifest excludes exactly its own path, this evidence path, and the independent audit path, leaving an exact 34-row path and digest match. The current remediation record and successor design amendment are included in the manifest.

At candidate commit `a1f146cdb5b2f20a7852bcf490223541fe4c8986`:

- tracked inventory: 103 regular Git blobs and 1,269,272 bytes;
- base-to-candidate delta: 37 files, 5,693 insertions, and 17 deletions, with no deleted base path;
- only three `100755` regular script blobs differ from mode `100644`; no symlink or submodule entry exists;
- controller result: 163 requirements, allocations `140/18/5`, 12 prototype skills, 43 ownership controls, 103 files, and 37 changes;
- dependency/product state: no `go.sum`, `vendor/`, product command, updater, product runtime, or grant activation;
- candidate manifest SHA-256: `1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb`; and
- worktree/index state before both exact-candidate verification runs and before this evidence edit: clean.

This evidence file is excluded from the candidate manifest and does not embed its own digest or containing commit. The completion handoff reports the evidence SHA-256 and evidence-only commit/tree/parent. Any candidate or evidence byte change invalidates a later audit.

## 2. Fourth remediation lineage

| Commit | Purpose | Exact changed paths |
|---|---|---|
| `62e1a019eb6a75748c628c93102c41db81166d28` | Preserve `L7-AUD-W01-004` before candidate mutation | `docs/artifacts/wave-01-audit.md` only |
| `6c04c537fa3a1af2a0ba0ab3db469b99d8852593` | `fix(audit-AUD-W01-016): stabilize bounded enumeration` | `policy.go`, `policy_test.go` |
| `4b092a65e74d713346975de5bb4d78d161ad2b0a` | `fix(audit-AUD-W01-017): bind verifier temp effects` | `Makefile`, `wave-01-design-amendment.md`, `policy.go`, `testutil_test.go` |
| `a1f146cdb5b2f20a7852bcf490223541fe4c8986` | Bind `L7-REM-W01-004` and the regenerated self-excluded candidate manifest | `wave-01-audit-remediation.md`, `wave-01-candidate.sha256` only |

The source audit's one `MEDIUM` and one `LOW` actionable finding are addressed in separate commits with regression proof. An initial unbound placement of the temporary-root assertion outside the registered Wave 1 path set was detected and moved into `testutil_test.go` before the finding commit was finalized. The candidate retains the exact registered 37-path closure. This is the remediator's evidence, not independent closure; every prior audit remains correct for its exact candidate.

## 3. Environment and effects

| Item | Observed value |
|---|---|
| Verification timestamp | Baseline and shadow runs completed on 2026-08-26 before evidence authoring; handoff inventory observed at `2026-08-26T04:15:15Z` |
| Host | macOS 26.5.2 build 25F84; Darwin 25.5.0; arm64 |
| Baseline compiler | `go version go1.26.7 darwin/arm64` |
| Shadow compiler | `go version go1.27.0 darwin/arm64` |
| Make | GNU Make 3.81 |
| Git | 2.50.1, Apple Git-155 |
| Go policy | `GOENV=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`; repository-scoped telemetry fixed `off` |
| Temporary-root policy | `TMPDIR` equals `GOTMPDIR`, both fixed to ignored repository path `.cache/go/tmp`; permanent test proves `t.TempDir()` containment |
| Expected verifier writes | Ignored repository `.cache/go/`, `.cache/repro/`, and existing `.cache/toolchains/` only |
| Other local effects | No non-cache repository or ambient host file effect observed in this remediation; no cleanup of ambient files |
| External effects | No network, remote, hosted workflow, provider, host plugin, publication, release, deployment, or exposure action |

Real filesystem fixtures create only test-owned roots under repository `.cache/go/tmp` and exercise symlink, FIFO, replacement, and directory-enumeration behavior. Permanent process fixtures execute only the current local test binary or pinned local Go binaries. Package-boundary fixtures use offline module settings, local replacement modules where required, and test-owned repository-scoped cache roots. Go testing cleans `t.TempDir` roots; retained verifier cache/reproducibility state remains ignored under `.cache`.

## 4. Exact candidate verification

Both commands ran after candidate commit `a1f146cdb5b2f20a7852bcf490223541fe4c8986` with a clean worktree.

### 4.1 Baseline — Go 1.26.7

Command: `make verify GO_VERSION=1.26.7`

Result: `PASS` in 10.41 seconds wall time.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da
```

### 4.2 Shadow — Go 1.27.0

Command: `make verify GO_VERSION=1.27.0`

Result: `PASS` as local shadow development evidence in 10.72 seconds wall time.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible internal/harness test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 4.3 Fourth-remediation closure checks

| Check | Local evidence | Result |
|---|---|---|
| Stable aggregate bound | A sentinel-filled directory batch returns directory-scoped `SCOPE-338` before sorting or entry inspection | `PASS` |
| Real filesystem order | Oversized all-file roots created ascending and descending render byte-identical rule, subject, message, and recovery | `PASS` |
| Mixed subset order | Omitted-directory and omitted-file capped batches render the same complete result without calling `DirEntry.Info` | `PASS` |
| Cross-process result | Both mixed variants exit 1 with byte-identical `BLOCKED` output | `PASS` |
| Existing resource boundaries | 512 directory, 512 file, 8 MiB byte, 1,027-entry read, five-second deadline, strict-file, and diagnostic caps remain tested | `PASS` |
| Temporary-root containment | `TMPDIR == GOTMPDIR == <repository>/.cache/go/tmp`; `t.TempDir()` remains underneath | `PASS` |
| Successor design contract | Amendment 0.2.0 records actual real-filesystem/process/clock effects and `SCOPE-338` semantics | `PASS` |
| Candidate manifest | Exact 34-row path/digest reproduction with three non-circular exclusions | `PASS` |
| Deletion/whitespace | No base deletion; `git diff --check` | `PASS` |

The prior trace, claim, ownership, inventory-root, file-identity, scope, import, source-digest, no-repair, environment-isolation, deadline, and process suites also remain green. These results are same-user local development evidence and do not self-clear the audit.

## 5. Acceptance and audit disposition

| Acceptance | Remediation evidence disposition |
|---|---|
| `W01-AC-006` | Rooted fail-closed scanning now provides deterministic aggregate/count/byte/time-bound failures; independent adequacy decision pending |
| `W01-AC-007` | Permanent real/injected/process order regressions and temporary-effect proof are present; independent adequacy decision pending |
| `W01-AC-011` | Baseline and shadow local matrices green with repository-scoped effects; hosted CI remains `NOT_RUN` |
| `W01-AC-012` | **NOT CLEARED** — fresh independent Mode B audit of the exact fourth-successor/evidence tuple is mandatory |
| `W01-AC-013` | Registered 37-path closure, no dependency/product/grant activation, and offline effect boundary reproduce locally; independent decision pending |
| Other Wave 1 acceptance | Prior independent dispositions remain subject to the fresh exact-candidate audit and unchanged later gates |

No aggregate passing result overrides `W01-AC-012`. This writer does not issue a release verdict, release approval, support, security qualification, AP2/AP3, or deployment authority.

## 6. `NOT_RUN` and limitations

- Fresh independent Mode B fourth-successor audit: `NOT_RUN`.
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

1. candidate commit `a1f146cdb5b2f20a7852bcf490223541fe4c8986`, tree `db580f77234dc14289f22174760d6da9bf442891`, parent `4b092a65e74d713346975de5bb4d78d161ad2b0a`;
2. candidate manifest SHA-256 `1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb`;
3. remediation record SHA-256 `e147c48f204118172e36d9e234e424516f7f1e06c638d4c3fac65dec0f08293b` and design-amendment SHA-256 `e10378f598098d5db8e9f20177324e917260e0ce016453903ac0159485526470`;
4. source audit commit `62e1a019eb6a75748c628c93102c41db81166d28` and SHA-256 `16f11e11b466a78cb7bf758cff40b7e0f7e85057e73af217bb69649795003917`;
5. finding commits `6c04c537fa3a1af2a0ba0ab3db469b99d8852593` and `4b092a65e74d713346975de5bb4d78d161ad2b0a`;
6. baseline reproducible SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` and shadow SHA-256 `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a`; and
7. this evidence file's final SHA-256 plus evidence-only commit/tree/parent from the completion handoff.

The fresh reviewer must reproduce identities and closure, inspect implementation and permanent fixtures, independently decide whether `AUD-W01-016` and `AUD-W01-017` are closed, rescore all findings, evaluate every Wave 1 acceptance criterion, and issue exactly one Mode B verdict. No merge, release, deployment, publication, external action, or later phase is authorized by this evidence.
