# Level 7 Dev Loop — Wave 1 Remediation Successor Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-002` |
| Artifact type | Local development evidence-only child for the remediated successor candidate |
| Version | 0.2.0 |
| Date | 2026-08-25 |
| Status | **SUCCESSOR VERIFIED LOCALLY — FRESH INDEPENDENT AUDIT REQUIRED** |
| Supersedes | `L7-EVD-W01-001` only for successor-candidate evidence; original bytes remain at commit `ce62a1f` |
| Producer | Wave 1 remediator using `level7-dev-loop:l7-release` Mode C under fresh accountable-owner authorization |
| Evidence state | Same-user local development evidence; not independent, hosted, protected, security-qualified, release, or support evidence |
| Successor candidate commit | `808fc16b1c13c96e30c66f08a97dd2a014b31db0` |
| Successor candidate tree | `bd914fdcb88aad54be5acf19c60dbcbb35f2d709` |
| Successor candidate parent | `f8788906a17a525c47435a69876171f21daef545` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `d8fcfbd3d78fa104449c7edcc9d68444fdd8a1a776e844993a41b60527f0acb0` |
| Remediation record | `docs/artifacts/wave-01-audit-remediation.md`, SHA-256 `0a5f3895b263f8f7acf82611d066d35f8f8093257f9baf9b61c3349766f5be28` |
| Source NO-GO audit | `docs/artifacts/wave-01-audit.md`, SHA-256 `54e03a8a20ced5a944b8e2506dd424581f9fbf009bf2711955cb07535d8a1224` |
| Branch | Local `feat/wave-01-build-control`; clean at exact candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED`, remote-absent, and unpublished |
| Next gate | Fresh structurally separate Mode B audit of this exact candidate and evidence-only child |

## 1. Acyclic binding and closure

The reproduced chain is:

```text
approved base ee181b7
  -> original implementation candidate 4bf4850
  -> original evidence child ce62a1f
  -> preserved NO-GO audit record 42ea8ea
  -> finding-specific remediation commits
  -> successor candidate 808fc16 + self-excluded candidate manifest
  -> this evidence-only child
  -> later fresh independent audit record
```

The successor candidate manifest contains 34 bytewise-sorted records. Relative to the approved base, the candidate changes 37 paths. The manifest excludes exactly its own path, this evidence path, and the independent audit path, leaving an exact 34-row path and digest match. The remediation record is included in the manifest.

At candidate commit `808fc16`:

- tracked non-`.git`/non-`.cache` inventory: 103 regular files and 1,241,419 bytes;
- base-to-candidate delta: 37 files, 4,977 insertions, and 16 deletions, with no deleted base path;
- controller result: 163 requirements, allocations `140/18/5`, 12 prototype skills, 43 ownership controls, 103 files, and 37 changes;
- dependency/product state: no `go.sum`, `vendor/`, product command, updater, product runtime, or grant activation;
- candidate manifest SHA-256: `d8fcfbd3d78fa104449c7edcc9d68444fdd8a1a776e844993a41b60527f0acb0`; and
- worktree/index state before verification and before this evidence edit: clean.

This evidence file is intentionally excluded from the candidate manifest and does not embed its own digest or containing commit. The completion handoff reports the evidence SHA-256 and evidence-only commit/tree/parent. Any candidate or evidence byte change invalidates a later audit.

## 2. Remediation lineage

| Commit | Purpose |
|---|---|
| `42ea8ea` | Preserve the independent NO-GO audit before candidate mutation |
| `b92003f` | `fix(audit-AUD-W01-002): bound controller resource use` |
| `f0a9491` | `fix(audit-AUD-W01-003): complete claim boundary record` |
| `2a58891` | `fix(audit-AUD-W01-004): bind ownership to source` |
| `916602a` | `fix(audit-AUD-W01-005): bind success output sources` |
| `f878890` | `fix(audit-AUD-W01-001): complete adversarial matrix` |
| `808fc16` | Bind the remediation record, registered path/ownership controls, and final candidate manifest |

The source audit's one `HIGH`, three `MEDIUM`, and one `LOW` actionable findings are addressed in separate finding commits with regression proof. This statement is the remediator's evidence, not independent closure. The old audit remains NO-GO for the old candidate.

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

The permanent package-boundary fixtures build and inspect isolated local modules with the pinned Go binary, local replacement modules where needed, offline module settings, and test-owned temporary/cache roots. They do not create product packages in the repository or contact external systems.

## 4. Exact candidate verification

The following commands ran after candidate commit `808fc16` with a clean worktree.

### 4.1 Baseline — Go 1.26.7

```text
make verify GO_VERSION=1.26.7
```

Result: `PASS`.

```text
no module dependencies; modules verified
gate_version=wave-01-v1; 12 exact source digests reported
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
phase=wave-01; requirements=163; allocation=140/18/5
prototypes=12; ownership=43; files=103; changed=37
import boundaries: PASS (2-package current graph)
compile/typecheck/unit/adversarial/vet/format/shell checks: PASS
repeat-build comparison: PASS
reproducible test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 4.3 Remediation closure checks

| Check | Local evidence | Result |
|---|---|---|
| Resource bounds | Limited reads; bounded skill entries, directory/file/byte traversal, expansion, and finding collection; at/over tests | `PASS` |
| Trace matrix | Missing/duplicate/malformed/unknown/reversed/overlap, zero/two owners, all allocation drifts, and displayed-summary tampering | `PASS` |
| Claim matrix | Plugin authority, both cross-host directions, A3/A4/A5, stable, dual, enforcement, generic/specialist, P0/P1/P2, and prototype inventory/dispositions | `PASS` |
| Scope/file matrix | Missing/unknown/duplicate phase, stale base, manifest drift, add/modify/delete, aliases, symlink/special/hardlink, protected bytes, updater reservation | `PASS` |
| Import matrix | Real isolated broken graphs for external detour, harness import, `unsafe`, and forbidden transitive import | `PASS` |
| Ownership | Requirement/allocation sources, 43 disjoint controls, and all 17 orchestration §10 classes source-crosschecked | `PASS` |
| Process semantics | Stable nonzero failure, zero success, bounded ordered output, cap, no repair, environment non-authority, repeated-run determinism | `PASS` |
| Success schema | Gate version plus exact digests for requirements, backlog, support, dispositions, phase, paths, base, candidate, ownership, orchestration, modules, and imports | `PASS` |
| Candidate manifest | Exact 34-row path/digest reproduction with three non-circular exclusions | `PASS` |
| Deletion/whitespace | No base deletion; `git diff --check` | `PASS` |

These results remain same-user local development evidence. They do not self-clear the audit.

## 5. Acceptance and audit disposition

| Acceptance | Remediation evidence disposition |
|---|---|
| `W01-AC-002`, `005`, `006`, `007`, `010` | Remediation and local regression evidence present; independent adequacy decision pending |
| `W01-AC-011` | Baseline and shadow local matrices green; hosted CI remains `NOT_RUN` |
| `W01-AC-012` | **NOT CLEARED** — fresh independent Mode B audit of the exact successor/evidence tuple is mandatory |
| Other Wave 1 acceptance | Prior local evidence remains subject to fresh audit and unchanged later gates |

No aggregate passing result overrides `W01-AC-012`. This writer does not issue `GO`, `CONDITIONAL GO`, release approval, support, security qualification, AP2/AP3, or deployment authority.

## 6. `NOT_RUN` and residual boundaries

- Fresh independent Mode B successor audit: `NOT_RUN`.
- Hosted GitHub Actions on Ubuntu 24.04: `NOT_RUN`; configuration inspected locally only.
- GitHub repository/account/remote authentication and publication identity: `NOT_RUN`; no remote exists.
- Actual Codex/Claude package discovery, compatibility, lifecycle, and cross-host conformance: `NOT_RUN`.
- Controlled-client qualification, provider/model/host/platform evaluation, protected holdout, pilot, and stable grant gates: `NOT_RUN` or absent.
- Grant security/boundary review and normative adoption: `NOT_RUN`; the proposal remains inert.
- AP2/AP3, signing, TUF, promotion, release, deployment, exposure, and monitoring: `NOT_RUN` or absent.
- Local Git/filesystem evidence remains same-user mutable and creates no external trust or production claim.

## 7. Fresh-audit handoff

The next structurally separate reviewer receives:

1. successor candidate commit `808fc16b1c13c96e30c66f08a97dd2a014b31db0`, tree `bd914fdcb88aad54be5acf19c60dbcbb35f2d709`, parent `f8788906a17a525c47435a69876171f21daef545`;
2. candidate manifest SHA-256 `d8fcfbd3d78fa104449c7edcc9d68444fdd8a1a776e844993a41b60527f0acb0`;
3. remediation record SHA-256 `0a5f3895b263f8f7acf82611d066d35f8f8093257f9baf9b61c3349766f5be28`;
4. this evidence file's final SHA-256 and evidence-only commit/tree/parent from the completion handoff;
5. source audit SHA-256 `54e03a8a20ced5a944b8e2506dd424581f9fbf009bf2711955cb07535d8a1224`; and
6. read-only authority only, with no candidate remediation, merge, publication, release, deployment, or external effect.

The reviewer must reproduce identities, inspect each finding disposition and regression proof, and issue a new verdict without inheriting the remediator's conclusion.
