# Level 7 Dev Loop — Wave 1 Candidate Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-EVD-W01-001` |
| Artifact type | Local development candidate evidence; evidence-only child record |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **CANDIDATE VERIFIED LOCALLY — INDEPENDENT AUDIT NOT RUN** |
| Producer | Wave 1 implementation writer using `level7-dev-loop:l7-build` under the current accountable-owner authorization |
| Evidence state | Local same-user development evidence; not independent, protected, hosted, security-qualified, release, or support evidence |
| Candidate commit | `4bf485022aeb23cab51856a4749663d2d6f78619` |
| Candidate tree | `6f59ca1d203e982b1d76b617a5209c85cb9d9d7e` |
| Candidate parent | `3afde82cac5eaac346f6daae7fa2c0337073e323` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate manifest | `docs/artifacts/wave-01-candidate.sha256`, SHA-256 `fb4f092fbbbaf85b8bc3e073e2853d2ce4ae0f06ea1d61ec92d3d6bbba6579bd` |
| Branch | Local `feat/wave-01-build-control`; clean at candidate verification |
| Module | `github.com/addressanup/level7-dev-loop`; GitHub location/ownership remains `USER_ASSERTED` and unpublished |
| Verification timestamp | 2026-08-25T17:05:21Z identity capture; final matrix and closure checks executed immediately before this record |
| Next gate | Separately authorized independent read-only `l7-release` audit of the exact candidate and evidence identities |

## 1. Evidence binding and closure

The candidate manifest contains 33 bytewise-sorted changed-file records. It excludes only:

- the manifest itself;
- this evidence-only record; and
- the absent later independent audit record.

Reproduction from the approved base enumerated 34 changed files including the manifest, excluded the three declared non-circular paths, recalculated every SHA-256 digest, and produced an exact zero-diff match with all 33 manifest rows.

At candidate commit `4bf4850`:

- tracked repository inventory: 100 files and 1,147,562 bytes;
- candidate delta: 34 files, 3,413 added lines, and 16 deleted lines;
- build-control inventory: 100 non-`.git`/non-`.cache` files and 34 changes relative to the approved base;
- deletion audit: `PASS` with no deleted base path;
- `go.sum`, `vendor/`, product command paths, updater path, evidence record, and audit record: absent; and
- `git status --porcelain=v2 --branch --untracked-files=all`: branch identity only, with no tracked or untracked candidate change.

This evidence file is intentionally not in the candidate manifest and does not contain its own digest or containing commit identity. The completion handoff must report this file's final SHA-256 and the evidence-only commit/tree/parent. A later audit must bind both the candidate tuple above and that evidence tuple.

## 2. Environment and declared effects

| Item | Observed value |
|---|---|
| Host | macOS 26.5.2 build 25F84; Darwin 25.5.0; arm64 |
| Baseline compiler | `go version go1.26.7 darwin/arm64` |
| Shadow compiler | `go version go1.27.0 darwin/arm64` |
| Make | GNU Make 3.81 |
| Git | 2.50.1, Apple Git-155 |
| Digest tool | `shasum` 6.02, SHA-256 mode |
| Go environment | `GOENV=off`, `GOTOOLCHAIN=local`, `GOWORK=off`, `CGO_ENABLED=0`, host OS/architecture fixed, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`, telemetry redirected and fixed `off` |
| Permitted writes | Ignored repository `.cache/go/`, `.cache/repro/`, and existing `.cache/toolchains/`; repository-scoped telemetry mode only |
| External effects | None observed or authorized; no remote, hosted workflow, provider, host plugin, publication, release, deployment, or exposure action |

The two final `make verify` commands ran concurrently and completed in an observed combined command-runner wall time of approximately 9.3 seconds. This is a local observation, not a benchmark or performance claim.

## 3. Final verification results

### 3.1 Baseline — Go 1.26.7

Command:

```text
make verify GO_VERSION=1.26.7
```

Result: `PASS`.

Bounded result summary:

```text
go: no module dependencies to download
all modules verified
PASS rule=BCTL-000 phase=wave-01 requirements=163 allocation=v1.0:140,v1.x:18,later:5 prototypes=12 ownership=31 files=100 changed=34
check-import-boundaries: PASS (2 package set)
internal/harness compile/type check: PASS
internal/harness/buildcontrol compile/type check: PASS
internal/harness tests: PASS
internal/harness/buildcontrol tests: PASS
repeat-build comparison: PASS
reproducible test-binary SHA-256: e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da
```

### 3.2 Shadow — Go 1.27.0

Command:

```text
make verify GO_VERSION=1.27.0
```

Result: `PASS` as local shadow evidence. The configured hosted shadow job remains allowed-to-fail and `NOT_RUN`.

Bounded result summary:

```text
go: no module dependencies to download
all modules verified
PASS rule=BCTL-000 phase=wave-01 requirements=163 allocation=v1.0:140,v1.x:18,later:5 prototypes=12 ownership=31 files=100 changed=34
check-import-boundaries: PASS (2 package set)
internal/harness compile/type check: PASS
internal/harness/buildcontrol compile/type check: PASS
internal/harness tests: PASS
internal/harness/buildcontrol tests: PASS
repeat-build comparison: PASS
reproducible test-binary SHA-256: da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a
```

### 3.3 Candidate closure checks

| Check | Method | Result |
|---|---|---|
| Candidate controller | Two identical `make policy-check GO_VERSION=1.26.7` invocations | `PASS`; output byte-identical |
| Requirement trace | Source-derived controller result | `PASS`: 163 unique IDs; `140/18/5` |
| Prototype disposition | Protected inventory plus `harness/prototype-dispositions.tsv` | `PASS`: 12 invocable skills, one approved disposition each |
| Shared ownership | Exact/prefix disjoint ownership plus path-policy cross-check | `PASS`: 31 controls |
| Manifest | Recompute all changed paths/digests from approved base and `diff -u` | `PASS`: 33 rows |
| Import boundary | `make import-check GO_VERSION=1.26.7` and built-in positive/sibling-prefix controls | `PASS`: 2-package set |
| Dependency/product absence | Controller plus explicit absence of `go.sum`, `vendor`, `cmd/l7`, and `cmd/l7up` | `PASS` |
| Secret pattern review | Valid 33-file manifest path set scanned for common private-key, AWS, GitHub, Slack, and Google credential forms | `PASS`; no match |
| Build-control effect review | Static import/call scan for process, network, randomness, clock, environment-policy, and filesystem mutation entry points | `PASS`; no match |
| Boundedness | Input bytes/lines/line length, repository file/byte counts, finding count, and message length are constant-bounded | `PASS` |
| File/path safety | Direct adversarial cases for aliases, symlink, pipe/device, and hardlink plus live repository scan | `PASS` |
| Protected ownership | Direct candidate-writer denial for protected controls | `PASS` |
| Deletion and whitespace | Base-to-candidate name-status plus `git diff --check` | `PASS` |

The pattern and static scans are defense-in-depth reviews, not comprehensive secret detection, sandbox proof, or an independent security audit.

## 4. Candidate commit chain

| Commit | Purpose |
|---|---|
| `8e8779d` | `docs(wave-01): bind approved build-control plan` |
| `44a853a` | `test(build-control): enforce scope and claim contracts` |
| `bc6c4b1` | `docs(wave-01): resolve candidate digest binding` |
| `ee42de2` | `feat(build-control): add fail-closed phase gate` |
| `c0044c6` | `chore(harness): activate wave 01 controls` |
| `3afde82` | `docs(wave-01): freeze implementation candidate` — superseded after pre-evidence coverage review |
| `4bf4850` | `test(build-control): cover boundary bypasses` — current exact candidate and manifest |

The current candidate is the complete snapshot at `4bf4850`; earlier commits are history and do not carry forward an independent result.

## 5. Failure, correction, and interruption record

| Event | State and response |
|---|---|
| Circular manifest/evidence/commit wording found before the phase gate | Implementation stopped. The accountable owner approved `L7-AMD-W01-DES-001`; candidate, evidence, and audit binding became an acyclic commit chain. |
| First Slice 2 test run failed | The validator incorrectly required legacy protected manifests to be bytewise sorted, did not recognize the existing comment-header module registry, and validated prefix ownership paths as exact paths. Parsers were narrowed to the actual approved formats; targeted and full tests then passed. |
| Repository disk pressure from ignored reproducibility outputs | Work stopped. The accountable owner explicitly authorized deletion of only `/Users/anuppandey/Desktop/level7-dev-loop/.cache/repro/*`; ignored reproducibility outputs were removed and later regenerated. No source, toolchain, or other cache was removed. |
| Initial candidate `3afde82` lacked direct adversarial cases for file shape and protected writer denial | Pre-evidence self-review treated the gap as candidate-invalidating. Direct cases and import-prefix controls were added, the manifest was regenerated, and the complete matrix was rerun for successor `4bf4850`. |
| Preliminary manifest/secret review command formulations were invalid | One temporary-file formulation was rejected before execution; two path-extraction formulations produced no valid evidence. They were discarded. Process-substitution manifest reproduction and an exact 33-path array scan produced the recorded passing results. |

No failed or superseded observation is represented as candidate evidence.

## 6. Acceptance disposition

| Acceptance | Candidate disposition |
|---|---|
| `W01-AC-001` | `PASS` — 163 source-derived IDs, one owner/allocation each, totals `140/18/5` |
| `W01-AC-002` | `PASS` — v1/support/effect/proof-profile matrix is explicit and withholds support |
| `W01-AC-003` | `PASS` — 12 skill dispositions; protected prototype bytes verified |
| `W01-AC-004` | `PASS` — stable/dual/enforcement claims withheld |
| `W01-AC-005` | `PASS` — exact path/phase/owner/module tuples fail closed on drift |
| `W01-AC-006` | `PASS` — successor active; historical Step 5 checker and evidence bytes preserved |
| `W01-AC-007` | `PASS` — positive and adversarial trace/claim/scope/module/file/ownership/import cases precede any product path |
| `W01-AC-008` | `PASS` for local candidate — exact root module selected; updater remains `reserved`/`UNSET`; external GitHub proof remains later |
| `W01-AC-009` | `PASS` as an inert proposal — grant kinds are non-interchangeable and no Wave 1 code parses or activates the artifact |
| `W01-AC-010` | `PASS` — ownership complete/disjoint; protected writer remains `external-denied` |
| `W01-AC-011` | `PASS` locally — baseline and shadow green; hosted CI `NOT_RUN` |
| `W01-AC-012` | `PARTIAL / AUDIT PENDING` — candidate manifest and evidence chain prepared; independent read-only audit is `NOT_RUN` and separately unauthorized |
| `W01-AC-013` | `PASS` — no dependency, product behavior, prototype edit, unexpected path, external effect, stable claim, or detected secret introduced |

Wave 1 is not complete because `W01-AC-012` still requires the independent R3 audit. No aggregate local result waives that gate.

## 7. Limits and `NOT_RUN` states

- Independent read-only scope/security audit: `NOT_RUN`.
- Hosted GitHub Actions on Ubuntu: `NOT_RUN`; the workflow is configured only.
- GitHub repository/account/remote authentication: `NOT_RUN`; the module decision is `USER_ASSERTED` only and no remote exists.
- Controlled Ubuntu/Bubblewrap/provider/host/model qualification: `NOT_RUN`.
- Codex/Claude actual-host discovery, compatibility, and lifecycle tests: `NOT_RUN`.
- Protected evaluator, grant issuer, signing, TUF, promotion, release, deployment, and exposure: `NOT_RUN` or absent.
- Product commands, runtime behavior, user-visible feature, feature flag, production dependency, updater, and grant activation: absent/`NOT_APPLICABLE` in Wave 1.
- Local Git and filesystem evidence is same-user mutable. SHA-256 and commit binding improve reproducibility but create no external trust, AP2/AP3 assurance, release verdict, or support claim.

## 8. Required audit handoff

The later reviewer must receive:

1. candidate commit `4bf485022aeb23cab51856a4749663d2d6f78619`, tree `6f59ca1d203e982b1d76b617a5209c85cb9d9d7e`, parent `3afde82cac5eaac346f6daae7fa2c0337073e323`;
2. candidate-manifest SHA-256 `fb4f092fbbbaf85b8bc3e073e2853d2ce4ae0f06ea1d61ec92d3d6bbba6579bd`;
3. the final evidence file SHA-256 and evidence-only commit/tree/parent reported in the completion handoff;
4. the approved planning, module decision, digest-binding amendment, inert grant proposal, base manifest, path policy, ownership/claim/trace data, tests, and this result/limit record; and
5. read-only authority only, with no candidate remediation, merge, publication, release, deployment, or exposure effect.

Any candidate or evidence change invalidates the corresponding identities and requires fresh verification and audit binding.
