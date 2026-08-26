# Level 7 Dev Loop — Wave 1 Third-Successor Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-004` |
| Artifact type | Independent Principal Engineer release audit — `level7-dev-loop:l7-release` Mode B |
| Version | 0.4.0 |
| Date | 2026-08-26 |
| Accountable owner | Anup Pandey |
| Reviewer role | Fresh, structurally separate, read-only Principal Engineer reviewer |
| Audited candidate | Commit `3f6daebee32d16b38497e74235c25f3b6a443fe1`; tree `77e731b269612cbeee078a25fffde443b8fafbe5`; parent `47f8e7af4e942964f8a8046864fdcb8dc267ffa1` |
| Evidence-only child | Commit `cca98593bd63c7ebbd869a9ed34e41e71923f164`; tree `85d93ece84104ad227f5e3060b85858908d23675`; parent `3f6daebee32d16b38497e74235c25f3b6a443fe1` |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Source audit | Commit `910c21406b717da13a7bdfe8ac73357b5078b251`; artifact SHA-256 `27b255f6cbdaaf050c7268828f26cd24be9b5120aa38db85ea2fb63fc9288039` |
| Audit scope | Exact Wave 1 third-successor R3 development-gate candidate and evidence chain only |
| Verdict | **NO-GO** |

## 1. Decision

`W01-AC-012` is **not cleared** for the exact candidate. All supplied commit, tree, parent, artifact-digest, 34-row manifest, and 37-path closure identities reproduce, and `AUD-W01-012` is adequately closed. The `AUD-W01-013` remediation bounds directory enumeration before materialization, but it does not make the bounded failure deterministic: `(*os.File).ReadDir(n)` selects up to `n` entries in filesystem order and the controller sorts only that already-truncated subset. An oversized directory can therefore change the emitted failure subject, and a mixed directory can change the first rule between `SCOPE-346` and `SCOPE-348`, for the same repository path-and-byte set. This violates the explicit no-unordered-filesystem-traversal contract and leaves one unresolved `MEDIUM` finding.

Passing same-user local tests cannot waive that finding. The release rule in `wave-01-specification.md:291,310` and the design audit seam at `wave-01-design.md:321-322` require zero unresolved `BLOCKER`, `CRITICAL`, `HIGH`, or `MEDIUM` findings.

This audit authorizes no remediation, candidate change, merge, release, deployment, publication, hosted CI, external action, or later phase.

## 2. Audit method, authority, and repository map

The complete `level7-dev-loop:l7-release` skill and repository `AGENTS.md` were read before audit actions. Mode B was applied: the candidate, evidence, code, tests, configuration, Git state, caches, toolchains, remotes, and external systems were not mutated or executed. The only write is this registered `audit-only` reviewer path. The protected historical `docs/artifacts/principal-engineer-release-audit.md` was not edited.

The initial worktree and index were clean on local branch `feat/wave-01-build-control`; `git worktree list --porcelain` showed one worktree; `git remote -v` showed no remote. Read-only mapping of the candidate found 103 blobs:

| Area | Files / role |
|---|---|
| Root | 10 governance, module, harness, and plugin metadata files |
| `.claude-plugin/`, `.codex-plugin/`, `.github/` | 3 host/plugin/workflow records |
| `docs/` | 46 specification, design, foundation, audit, evidence, and planning artifacts |
| `harness/` | 12 phase, path, support, ownership, module, and digest registries |
| `internal/` | 16 Go files in exactly two packages: `internal/harness` and `internal/harness/buildcontrol` |
| `skills/` | 12 user-invocable prototype skills |
| `scripts/` | 3 harness scripts |
| `references/` | 1 reference file |

There is no `go.sum`, `vendor/`, `cmd/`, `internal/product/`, `pkg/`, product runtime, updater, or product dependency. Git modes are 100 regular `100644` blobs and three regular `100755` scripts; no symlink or submodule entry exists. A filesystem inventory excluding `.git` and `.cache` found no symlink, FIFO, socket, or device. `git diff --check ee181b7..3f6daeb` passed.

The audit inspected:

- `docs/artifacts/wave-01-change-contract.md`, `wave-01-specification.md`, `wave-01-design.md`, `wave-01-design-amendment.md`, `wave-01-approval.md`, `wave-01-module-identity-decision.md`, and the inert `wave-01-grant-ladder-amendment.md`;
- source audit `L7-AUD-W01-003` at commit `910c214`, remediation record `L7-REM-W01-003`, and evidence `L7-EVD-W01-004`;
- `Makefile`, `.github/workflows/harness.yml`, all Wave 1 registry/manifests, `go.mod`, the complete `internal/harness/buildcontrol` implementation and permanent tests, `internal/harness/proving_test.go`, and the import-boundary script; and
- the pinned Go 1.26.7 and 1.27.0 `os.File.ReadDir` implementation and the existing ignored reproducibility binaries, read-only.

## 3. Identity, digest, manifest, and closure reproduction

### 3.1 Git and artifact identities

`git show -s --format='tree=%T parent=%P subject=%s' <commit>` reproduced:

| Object | Reproduced identity |
|---|---|
| Approved base | `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate | `3f6daebee32d16b38497e74235c25f3b6a443fe1`; tree `77e731b269612cbeee078a25fffde443b8fafbe5`; parent `47f8e7af4e942964f8a8046864fdcb8dc267ffa1` |
| Evidence child | `cca98593bd63c7ebbd869a9ed34e41e71923f164`; tree `85d93ece84104ad227f5e3060b85858908d23675`; parent `3f6daebee32d16b38497e74235c25f3b6a443fe1` |
| Source audit | `910c21406b717da13a7bdfe8ac73357b5078b251`; tree `b5cd3064d63c8f3b138624bf80d7a006870ebc05`; parent `8f82512fd4eb3e24ba6033427badaa3439e06450` |
| `AUD-W01-012` fix | `b9a48d5f55abbab9eeab1f0a4f1a536351a13a6e`; parent `910c214`; expected conventional subject |
| `AUD-W01-013` fix | `47f8e7af4e942964f8a8046864fdcb8dc267ffa1`; parent `b9a48d5`; expected conventional subject |

`git show <commit>:<artifact> | shasum -a 256` reproduced:

| Artifact | SHA-256 |
|---|---|
| Candidate manifest | `561d8ab3e6480435fbe0a4baa377ba098349e3766b0098edf5371b598620ae69` |
| Remediation record | `fe8e4d3674a46d4342f239b0473c08803c0f3368a643eaa8bccdee9b21a70a93` |
| Source audit | `27b255f6cbdaaf050c7268828f26cd24be9b5120aa38db85ea2fb63fc9288039` |
| Evidence-only record | `ba2bdc8a34e8a79b286b4ad4dbb99d8710782ba9b600d79288bac1f6ab64797a` |
| Protected generic audit | `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c` |
| Protected generic remediation | `d124659ce655415252255e5e50cd117ba964d634720fd4d9ccd98a0dd9872ca2` |

### 3.2 Manifest and path closure

Independent shell parsing and Git-blob hashing reproduced:

- 34 bytewise-sorted candidate-manifest data rows and zero digest mismatches;
- exactly 37 base-to-candidate changed paths and exactly 37 data rows in `harness/wave-01-paths.tsv`, with zero set difference;
- after excluding only `docs/artifacts/wave-01-candidate.sha256`, `docs/artifacts/wave-01-evidence.md`, and `docs/artifacts/wave-01-audit.md`, the 37-path delta reduces to the exact 34 manifest paths, with zero set difference;
- 72 base-manifest rows, each reproducing against the approved base Git blob;
- 26 protected foundation-manifest rows, each reproducing against the candidate Git blob; and
- no deletion from the approved base.

`git diff --name-status 3f6daeb..cca9859` returned exactly:

```text
M	docs/artifacts/wave-01-evidence.md
```

The evidence child therefore differs from the candidate only at the registered evidence path. The candidate contains 103 blobs and 1,259,997 bytes. `git diff --shortstat ee181b7..3f6daeb` reproduces 37 files changed, 5,544 insertions, and 16 deletions.

Finding-specific commit scopes also reproduce: `b9a48d5` changes only `claims.go`, `claims_test.go`, `load.go`, `main.go`, and `testutil_test.go`; `47f8e7a` changes only `policy.go` and `policy_test.go`; candidate-binding commit `3f6daeb` changes only the remediation record and candidate manifest.

## 4. Specification, design, and control evaluation

The governing requirements are explicit:

- all repository data are untrusted; rooted paths and special-node rejection are mandatory (`wave-01-specification.md:218-223`);
- identical admitted inputs must yield identical semantic results, and parsing must not depend on unordered filesystem traversal (`wave-01-specification.md:224-228`);
- roots and file/byte/time bounds must be declared, and overflow must block without silent truncation (`wave-01-specification.md:230-234`);
- every intended adversary must fail under its stable rule (`wave-01-specification.md:252-261`); and
- one failed acceptance criterion or one material audit finding cannot be waived (`wave-01-specification.md:291,299-310`).

The design likewise requires lexically sorted deterministic diagnostics (`wave-01-design.md:136-163`), a complete no-follow repository walk (`wave-01-design.md:241-265`), stable intended-rule negative cases (`wave-01-design.md:351-363`), and zero unresolved material findings for audit admission (`wave-01-design.md:321-322`).

Independent source-derived checks reproduced 163 unique normative definitions and ownership sums of total 163, `V1.0=140`, `V1.x=18`, `Later=5`. The candidate contains 19 exact support-matrix rows, 12 user-invocable skills, 12 prototype dispositions, and 43 ownership controls. Permanent tests cover malformed, missing, duplicate, unknown, stale, unowned, protected, module, path-shape, support-claim, diagnostic-cap, phase, import, and no-repair cases. The grant amendment is referenced by governance/ownership controls but no implementation loads or activates it.

## 5. Source finding remediation disposition

| Source finding | Independent disposition |
|---|---|
| `AUD-W01-012` — prototype inventory root | **ADEQUATELY CLOSED.** `load.go:31-90` validates a canonical rooted relative path, opens through `os.OpenRoot` with `O_NONBLOCK|O_NOFOLLOW|O_DIRECTORY`, and compares pre/open/post identity, type, mode, size, and modification time around a bounded `ReadDir`. `claims.go:171-202` uses that acquisition before reading individual strict `SKILL.md` files. Direct and full-controller real symlink, FIFO, and replacement cases at `claims_test.go:316-395` require `BCTL-022`/`BCTL-023` within one second. |
| `AUD-W01-013` — repository enumeration bound | **NOT ADEQUATELY CLOSED.** `policy.go:17-26,203-328` adds fixed directory/file/byte/time/read-batch bounds and prevents whole-directory materialization. Rooted no-follow identity checks at `policy.go:331-408` are sound for the inspected handles. However, the capped subset is selected in filesystem order before sorting, so the required stable semantic result remains traversal-order-dependent. See `AUD-W01-016`. |

## 6. Recorded verification and independent inspection

The evidence records these exact-candidate commands:

| Command | Recorded result | Reproduced read-only evidence |
|---|---|---|
| `make verify GO_VERSION=1.26.7` | `PASS`, 11.02 s | SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da`; multiple existing `harness-a.test`/`harness-b.test` pairs in `.cache/repro/` match |
| `make verify GO_VERSION=1.27.0` | `PASS`, 11.33 s shadow | SHA-256 `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a`; multiple existing pairs match |

`Makefile:39-76,115-154` pins offline Go settings, repository-scoped cache locations, build-control/import/format/vet/typecheck/test/repeat-build steps, and the printed `internal/harness` test-binary digest. `.github/workflows/harness.yml:15-42` configures Ubuntu 24.04 baseline/shadow jobs with read-only checkout credentials and a 15-minute job timeout, but no hosted job was invoked.

The audit did not rerun either `make verify` command or any test because those commands intentionally create cache, temporary fixture, process, and reproducibility outputs beyond Mode B. Candidate test source and existing ignored binaries were inspected read-only. This is sufficient to evaluate implementation and evidence consistency, not to upgrade same-user local evidence to independent or hosted evidence.

## 7. Severity model and findings

| Severity | Audit meaning |
|---|---|
| `BLOCKER` | Candidate/evidence identity cannot be reproduced or reviewed reliably. |
| `CRITICAL` | Demonstrated catastrophic authority, integrity, or external-effect failure. |
| `HIGH` | Central R3 boundary or checkpoint criterion is materially bypassable or absent. |
| `MEDIUM` | Required correctness/reliability contract is incomplete and must be corrected before the checkpoint. |
| `LOW` | Bounded conformance/maintainability defect that does not independently block the checkpoint. |
| `INFO` | Verified fact, limitation, or future gate with no candidate defect established. |

Finding counts: `BLOCKER 0`, `CRITICAL 0`, `HIGH 0`, `MEDIUM 1`, `LOW 1`, `INFO 2`.

### `AUD-W01-016` — `MEDIUM` — Capped repository enumeration depends on the unordered pre-sort subset

**Affected acceptance:** deterministic/fail-closed portions of `W01-AC-006`; stable-rule permanent fixtures in `W01-AC-007`; consequently `W01-AC-012`.

**Evidence:** `policy.go:226-243` computes a sentinel-sized request capped at 1,027, calls `readDirectory`, and only then sorts returned entries. `policy.go:331-365` calls `directory.ReadDir(entryLimit)`. In both pinned toolchains, `.cache/toolchains/go1.26.7/src/os/dir.go:91-106` and the byte-identical Go 1.27.0 file state that `(*File).ReadDir(n)` returns at most `n`; only the separate package function `os.ReadDir(name)` at lines 109-125 reads all entries and sorts them. The remediation therefore sorts a filesystem-selected partial set, not the complete set or a stable lexical prefix.

`policy_test.go:200-227` creates only regular files and asserts the maximum request, `SCOPE-346`, and retained-file count. It does not compare exact diagnostic subject/output across creation orders or exercise mixed file/directory overflow. With 1,227 regular files, a different 1,027-entry subset can change the 513th sorted file and therefore the `SCOPE-346` subject. With root names ordered as directories before files and 513 directories plus 515 files, omitting a directory from the capped subset yields `SCOPE-346`, while omitting a file yields `SCOPE-348`.

**Impact:** the scanner remains memory-bounded and fails nonzero, but its semantic diagnostic can vary with unordered filesystem traversal for the same repository path-and-byte set. This directly violates `wave-01-specification.md:226-227` and the stable diagnostic/rule contract. A recovery action and test selection can therefore depend on ambient directory enumeration order.

**Required disposition:** detect a sentinel-filled/truncated batch before category-specific iteration and emit one directory-scoped, deterministic aggregate-bound rule, or use another bounded strategy that identifies a stable lexical set. Add permanent real-filesystem and injected-reader regressions for oversized all-file and mixed file/directory inventories created or supplied in different orders, asserting byte-identical rendered rule, subject, message, and exit result. Rebind, fully verify, freeze a new evidence child, and obtain a fresh independent Mode B audit.

### `AUD-W01-017` — `LOW` — Accepted fixture/effect description is stale

`wave-01-design.md:98-104` says test fixtures need only immutable `testing/fstest.MapFS` and no temporary files, external process, or clock. Current permanent regressions use `t.TempDir`, wall-clock deadlines, FIFO/symlink fixtures, and child processes (`claims_test.go:299-395`, `policy_test.go:171-272`, `testutil_test.go:150-180,217-245`). Evidence lines 67-81 truthfully disclose test-owned temporary roots and local child processes, which bounds the operational risk, but the accepted design is no longer an accurate effect/test architecture description. Record this evolution in an owner-approved successor amendment rather than rewriting protected historical design bytes.

### `AUD-W01-018` — `INFO` — Local verification was inspected, not rerun

Candidate/evidence bytes, committed tests, Make/workflow definitions, and matching ignored reproducibility binaries were inspected. No cache-writing or temporary-fixture command was rerun under Mode B. The recorded passes remain same-user local development evidence. The printed reproducibility digest binds `internal/harness`, not the `internal/harness/buildcontrol` test binary. The five-second scan deadline is checked around bounded syscalls but cannot portably cancel a filesystem syscall already in progress; evidence lines 160-162 disclose both limits.

### `AUD-W01-019` — `INFO` — Later qualification and release gates remain separate

Hosted Ubuntu CI, authenticated GitHub repository/account/remote identity, actual Codex/Claude lifecycle and differential conformance, Controlled Client qualification, protected evaluation, grant security review and adoption, pilot/stable grants, AP2/AP3, signing, TUF, promotion, publication, release, deployment, exposure, and monitoring remain `NOT_RUN` or absent. This Wave 1 development audit clears none of them.

## 8. Wave 1 acceptance disposition

| Acceptance | Independent disposition |
|---|---|
| `W01-AC-001` | `PASS`: independent parsing reproduced 163 unique definitions, one owner/allocation each, and totals `140/18/5`; permanent missing/duplicate/unknown/allocation fixtures exist. |
| `W01-AC-002` | `PASS`: 19 exact support rows preserve the two-product distinction, A0-A2/A3-A5 boundary, three proof profiles, and development-only status. |
| `W01-AC-003` | `PASS`: 12 invocable prototype skills have exactly 12 dispositions; all 26 protected foundation rows reproduce. |
| `W01-AC-004` | `PASS`: stable version, dual-host, enforcement, compatibility, support, and release promotion claims remain withheld. |
| `W01-AC-005` | `PASS`: priority/scope impact and accountable-approval rules remain source-bound; no metric/date waiver exists. |
| `W01-AC-006` | `FAIL`: `AUD-W01-012` is closed, but oversized repository failure semantics still depend on unordered traversal (`AUD-W01-016`). |
| `W01-AC-007` | `NOT SATISFIED`: the new inventory-root and pre-enumeration bound regressions are permanent, but no fixture proves exact deterministic output across oversized partial-batch order or mixed file/directory overflow. |
| `W01-AC-008` | `PASS` for local development identity only: module records agree, updater remains reserved, and GitHub ownership/location remains `USER_ASSERTED`, remote-absent, and unpublished. |
| `W01-AC-009` | `PASS` as an inert proposal only: controls preserve non-interchangeability and separate security/adoption approval; no implementation activates it. |
| `W01-AC-010` | `PASS`: 43 disjoint controls cover the authoritative ownership classes and deny candidate authority over protected controls. |
| `W01-AC-011` | `PASS` as recorded same-user local evidence only: exact baseline and shadow matrices are recorded green with environment/effect limits; hosted CI remains truthful as `NOT_RUN`. |
| `W01-AC-012` | `FAIL`: exact identities, manifest closure, evidence separation, and independent audit separation reproduce, but one unresolved `MEDIUM` finding prevents the R3 development gate. |
| `W01-AC-013` | `PASS` for the exact candidate: no dependency, product/runtime/updater path, prototype edit, unexpected registered path, grant activation, stable claim, external service effect, or bounded credential-pattern match was found. |

No aggregate passing result waives `W01-AC-006`, `W01-AC-007`, `W01-AC-012`, or `AUD-W01-016`.

## 9. Final boundary

The exact candidate is not eligible for Wave 1 checkpoint clearance. The next authorized action, if the owner chooses, is a narrowly scoped Mode C remediation of `AUD-W01-016` (and the non-blocking design amendment for `AUD-W01-017`), followed by new candidate/evidence binding, complete baseline/shadow verification, and another structurally separate Mode B audit. Nothing in this record authorizes those actions automatically.
