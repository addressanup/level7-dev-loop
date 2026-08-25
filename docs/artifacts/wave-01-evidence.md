# Level 7 Dev Loop — Wave 1 Second Remediation Successor Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-003` |
| Artifact type | Local development evidence-only child for the second remediated successor candidate |
| Version | 0.3.0 |
| Date | 2026-08-26 |
| Status | **SECOND SUCCESSOR VERIFIED LOCALLY — FRESH INDEPENDENT AUDIT REQUIRED** |
| Supersedes | `L7-EVD-W01-002` only for second-successor evidence; prior bytes remain at commit `c0f834f` |
| Producer | Wave 1 remediator using `level7-dev-loop:l7-release` Mode C under accountable-owner authorization |
| Evidence state | Same-user local development evidence; not independent, hosted, protected, security-qualified, release, or support evidence |
| Second successor candidate commit | `64eee794519e381a69d284c32cc35ac58897aa2f` |
| Second successor candidate tree | `5829f97b1ef9e29a9c61d5872ed725920be65f84` |
| Second successor candidate parent | `f5f197cd76db474f3c3e8085ea611255b54585fb` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md`, SHA-256 `49ce412bd29791c3fde86cac0a1084ad19ff0d1c85187a51001ab1c949c406d2` |
| Source successor NO-GO audit | `docs/artifacts/wave-01-audit.md`, SHA-256 `bc58479c7626dd88ad4937eabfa0482b8c1e11a95f2a9c95ded114b379a1ef1b` |
| Branch | Local `feat/wave-01-build-control`; clean at exact candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED`, remote-absent, and unpublished |
| Next gate | Fresh structurally separate Mode B audit of this exact candidate and evidence-only child |

## 1. Acyclic binding and closure

The reproduced chain is:

```text
approved base ee181b7
  -> original candidate/evidence 4bf4850 / ce62a1f
  -> first preserved NO-GO audit 42ea8ea
  -> first remediation commits and successor candidate/evidence 808fc16 / c0f834f
  -> second preserved NO-GO audit 4a0685a
  -> finding-specific remediation commits 86a0a47 / f5f197c
  -> second successor candidate 64eee79 + self-excluded candidate manifest
  -> this evidence-only child
  -> later fresh independent audit record
```

The second-successor manifest contains 34 bytewise-sorted records. Relative to the approved base, the candidate changes exactly 37 registered paths. The manifest excludes exactly its own path, this evidence path, and the independent audit path, leaving an exact 34-row path and digest match. The current remediation record is included in the manifest.

At candidate commit `64eee79`:

- tracked inventory: 103 regular Git blobs and 1,245,127 bytes;
- base-to-candidate delta: 37 files, 5,167 insertions, and 16 deletions, with no deleted base path;
- controller result: 163 requirements, allocations `140/18/5`, 12 prototype skills, 43 ownership controls, 103 files, and 37 changes;
- dependency/product state: no `go.sum`, `vendor/`, product command, updater, product runtime, or grant activation;
- candidate manifest SHA-256: `60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd`; and
- worktree/index state before both exact-candidate verification runs and before this evidence edit: clean.

This evidence file is excluded from the candidate manifest and does not embed its own digest or containing commit. The completion handoff reports the evidence SHA-256 and evidence-only commit/tree/parent. Any candidate or evidence byte change invalidates a later audit.

## 2. Second remediation lineage

| Commit | Purpose |
|---|---|
| `4a0685a` | Preserve `L7-AUD-W01-002`, the independent second NO-GO audit, before candidate mutation |
| `86a0a47` | `fix(audit-AUD-W01-008): stabilize capped diagnostics` |
| `f5f197c` | `fix(audit-AUD-W01-009): harden fixed input reads` |
| `64eee79` | Bind `L7-REM-W01-002` and the regenerated self-excluded candidate manifest |

The source audit's two `MEDIUM` actionable findings are addressed in separate commits with regression proof. This is the remediator's evidence, not independent closure. Both prior NO-GO audits remain correct for their exact candidates.

## 3. Environment and effects

| Item | Observed value |
|---|---|
| Host | macOS 26.5.2 build 25F84; Darwin 25.5.0; arm64 |
| Baseline compiler | `go version go1.26.7 darwin/arm64` |
| Shadow compiler | `go version go1.27.0 darwin/arm64` |
| Make | GNU Make 3.81 |
| Git | 2.50.1, Apple Git-155 |
| Go policy | `GOENV=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`; repository-scoped telemetry fixed `off` |
| Permitted writes | Ignored repository `.cache/go/`, `.cache/repro/`, and existing `.cache/toolchains/` only |
| External effects | None observed or authorized; no remote, network fetch, hosted workflow, provider, host plugin, publication, release, deployment, or exposure action |

The permanent process fixture executes only the current local test binary in a child process. Filesystem fixtures use test-owned temporary roots. The package-boundary fixtures use pinned local Go binaries, offline module settings, local replacement modules where required, and test-owned cache roots.

## 4. Exact candidate verification

The following commands ran after candidate commit `64eee79` with a clean worktree.

### 4.1 Baseline — Go 1.26.7

```text
make verify GO_VERSION=1.26.7
```

Result: `PASS`.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible test-binary SHA-256: e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da
```

### 4.2 Shadow — Go 1.27.0

```text
make verify GO_VERSION=1.27.0
```

Result: `PASS` as local shadow development evidence.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
candidate-manifest source digest=60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 4.3 Second-remediation closure checks

| Check | Local evidence | Result |
|---|---|---|
| Deterministic capped failure | Bounded collection retains the lexicographically smallest 51 findings regardless of arrival order | `PASS` |
| Separate-process repeat failure | Two independent child processes generate 163 unordered missing-owner findings, exit nonzero, and emit byte-identical capped output | `PASS` |
| Fixed-input shape before consumption | Real final/intermediate symlink, FIFO, and hardlink fixtures return stable findings without consuming or blocking | `PASS` |
| Fixed-input rooted stability | `os.Root`, no-follow/nonblocking open, opened/path identity, mode, link, modification-time, and size checks surround the bounded read | `PASS` |
| Change in read window | Deterministic in-window mutation returns `BCTL-023` and no bytes | `PASS` |
| Candidate manifest | Exact 34-row path/digest reproduction with three non-circular exclusions | `PASS` |
| Deletion/whitespace | No base deletion; `git diff --check` | `PASS` |

The prior trace, claim, ownership, scope, import, source-digest, no-repair, environment-isolation, and process suites also remain green. These results are same-user local development evidence and do not self-clear the audit.

## 5. Acceptance and audit disposition

| Acceptance | Remediation evidence disposition |
|---|---|
| `W01-AC-006`, `W01-AC-007` | Targeted remediation and permanent regression evidence present; independent adequacy decision pending |
| `W01-AC-011` | Baseline and shadow local matrices green; hosted CI remains `NOT_RUN` |
| `W01-AC-012` | **NOT CLEARED** — fresh independent Mode B audit of the exact second-successor/evidence tuple is mandatory |
| Other Wave 1 acceptance | Prior independent dispositions remain subject to the fresh exact-candidate audit and unchanged later gates |

No aggregate passing result overrides `W01-AC-012`. This writer does not issue `GO`, `CONDITIONAL GO`, release approval, support, security qualification, AP2/AP3, or deployment authority.

## 6. `NOT_RUN` and residual boundaries

- Fresh independent Mode B second-successor audit: `NOT_RUN`.
- Hosted GitHub Actions on Ubuntu 24.04: `NOT_RUN`; configuration inspected locally only.
- GitHub repository/account/remote authentication and publication identity: `NOT_RUN`; no remote exists.
- Actual Codex/Claude package discovery, compatibility, lifecycle, and cross-host conformance: `NOT_RUN`.
- Controlled Client qualification, provider/model/host/platform evaluation, protected holdout, pilot, and stable grant gates: `NOT_RUN` or absent.
- Grant security/boundary review and normative adoption: `NOT_RUN`; the proposal remains inert.
- AP2/AP3, signing, TUF, promotion, release, deployment, exposure, and monitoring: `NOT_RUN` or absent.
- Local Git/filesystem evidence remains same-user mutable and creates no external trust or production claim.

## 7. Fresh-audit handoff

The next structurally separate reviewer receives:

1. candidate commit `64eee794519e381a69d284c32cc35ac58897aa2f`, tree `5829f97b1ef9e29a9c61d5872ed725920be65f84`, parent `f5f197cd76db474f3c3e8085ea611255b54585fb`;
2. candidate manifest SHA-256 `60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd`;
3. remediation record SHA-256 `49ce412bd29791c3fde86cac0a1084ad19ff0d1c85187a51001ab1c949c406d2`;
4. this evidence file's final SHA-256 and evidence-only commit/tree/parent from the completion handoff;
5. source audit commit `4a0685a` and SHA-256 `bc58479c7626dd88ad4937eabfa0482b8c1e11a95f2a9c95ded114b379a1ef1b`; and
6. read-only authority only, with no candidate remediation, merge, publication, release, deployment, or external effect.

The reviewer must reproduce identities, inspect both finding dispositions and regression proofs, and issue a new verdict without inheriting the remediator's conclusion.
