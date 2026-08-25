# Level 7 Dev Loop — Wave 1 Second-Successor Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-003` |
| Artifact type | Fresh independent read-only Principal Engineer audit; Level 7 Release Validation Mode B |
| Date | 2026-08-26 |
| Review mode | Structurally separate second-successor reviewer; exact frozen candidate and evidence inspection |
| Scope | Wave 1 build-control development checkpoint and clearance of `W01-AC-012` only |
| Verdict | **NO-GO** |
| Production/release meaning | None; this is not product support, deployment, publication, AP2, AP3, or production-release authority |
| Candidate remediation | Not authorized and not performed |
| Audit effect | This reviewer-owned file update only; no candidate, evidence, Git ref/index/config, environment, cache, toolchain, remote, or external mutation |

## 1. Decision

The exact Wave 1 second-successor candidate does **not** clear the R3 development checkpoint. Every supplied Git and artifact identity reproduces; all `34` candidate-manifest rows and the exact `37`-path closure reproduce; the evidence-only child changes only `docs/artifacts/wave-01-evidence.md`; and source inspection supports closure of `AUD-W01-008`.

Two `MEDIUM` correctness/reliability findings remain:

1. `AUD-W01-009` is not fully closed because the fixed protected prototype inventory directory is still opened with plain `os.Open` before repository shape validation. A FIFO at `skills/` can block the controller, and a symlink can expose out-of-root directory entries, before the later policy scan can return a stable failure.
2. Repository file/directory count limits are enforced only after `filepath.WalkDir` has materialized an entire directory through `os.ReadDir`; a large single directory can therefore consume memory and time before `SCOPE-346` or `SCOPE-348` is reached.

Neither defect creates a demonstrated pass, mutation, or authority widening. Both violate the mandatory bounded, fail-closed degraded behavior and permanent-fixture contracts for the first scope-relaxation gate. Under `wave-01-specification.md:230-234,252-261,282-291,299-310` and `wave-01-design.md:383-389,421-425`, either unresolved `MEDIUM` finding prevents `W01-AC-006`, `W01-AC-007`, and `W01-AC-012` from clearing. Passing recorded tests cannot waive those findings.

This decision is limited to the local Wave 1 build-control checkpoint. It does not assess or authorize product behavior, controlled mutation, host support, grant activation, security qualification, publication, release, deployment, exposure, or Wave 2.

## 2. Exact identity and closure reproduction

### 2.1 Git identities

| Identity | Reproduced value | Result |
|---|---|---|
| Approved base commit | `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4` | `PASS` |
| Approved base tree | `2f23a0810660995b6f562c361ab38cd4faafa3b3` | `PASS` |
| Candidate commit | `64eee794519e381a69d284c32cc35ac58897aa2f` | `PASS` |
| Candidate tree | `5829f97b1ef9e29a9c61d5872ed725920be65f84` | `PASS` |
| Candidate parent | `f5f197cd76db474f3c3e8085ea611255b54585fb` | `PASS` |
| Evidence-child commit / audit-start `HEAD` | `8f82512fd4eb3e24ba6033427badaa3439e06450` | `PASS` |
| Evidence-child tree | `a005a29728114fc7cf7f7d3243b4c988619d5836` | `PASS` |
| Evidence-child parent | `64eee794519e381a69d284c32cc35ac58897aa2f` | `PASS` |
| Source audit commit | `4a0685ad0e1c7950152f3af55c98789710f61693`; ancestor of the candidate | `PASS` |

The two finding-specific commits reproduce with the required subjects and scopes:

| Commit | Subject | Exact changed paths |
|---|---|---|
| `86a0a477f0418d8813657cbe2d1a6ba86085b97f` | `fix(audit-AUD-W01-008): stabilize capped diagnostics` | `report.go`, `testutil_test.go` |
| `f5f197cd76db474f3c3e8085ea611255b54585fb` | `fix(audit-AUD-W01-009): harden fixed input reads` | `load.go`, `claims_test.go` |
| `64eee794519e381a69d284c32cc35ac58897aa2f` | `docs(wave-01): bind second audit remediation` | only `wave-01-audit-remediation.md` and `wave-01-candidate.sha256` |

### 2.2 Artifact digests

All values were computed from the named Git object, not inferred from current prose:

| Artifact | Git object | Reproduced SHA-256 | Result |
|---|---|---|---|
| Candidate manifest | candidate | `60212b7bf49bc8a1435dd723ed72c45b110798317d602418f8a2f570abc44bdd` | `PASS` |
| Remediation record | candidate | `49ce412bd29791c3fde86cac0a1084ad19ff0d1c85187a51001ab1c949c406d2` | `PASS` |
| Evidence record | evidence child | `79a71e73c5f05fe83aa9a9e9a2c9db1d062a0c055b9c770b40d8c64fe4f2b6ed` | `PASS` |
| Source successor audit | `4a0685a` | `bc58479c7626dd88ad4937eabfa0482b8c1e11a95f2a9c95ded114b379a1ef1b` | `PASS` |

The protected generic skill-output record `docs/artifacts/principal-engineer-release-audit.md` remains byte-identical to the approved base at SHA-256 `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c`; it was inspected and not edited.

### 2.3 Manifest, path, and evidence closure

| Claim | Independent reproduction | Result |
|---|---|---|
| Candidate manifest | `34` data rows, bytewise path-sorted, zero digest mismatch against candidate blobs | `PASS` |
| Registered candidate delta | Base-to-candidate delta has exactly `37` paths; path set exactly equals the `37` data rows in `harness/wave-01-paths.tsv` | `PASS` |
| Three non-circular exclusions | Removing only `wave-01-candidate.sha256`, `wave-01-evidence.md`, and `wave-01-audit.md` from the `37`-path delta leaves exactly the `34` manifest paths | `PASS` |
| Evidence-only child | Candidate-to-child delta is exactly one modified path: `docs/artifacts/wave-01-evidence.md` | `PASS` |
| Approved base manifest | `72` rows; exact path closure over all base blobs; zero digest mismatch | `PASS` |
| Foundation/prototype manifest | `26` data rows; zero candidate digest mismatch | `PASS` |
| Protected paths | Zero base-to-candidate change across protected skills/manifests, `.go-version`, `.env.example`, `WORKFLOW.md`, Foundation locks, and the old Step 5 checker | `PASS` |
| Deletion and whitespace | Zero deleted base path; `git diff --check ee181b7..64eee79` returned no finding | `PASS` |

The candidate has `103` Git blobs totaling `1,245,127` bytes. Its base delta is `37` files, `5,167` insertions, `16` deletions, and no deletion. The only non-`100644` blob modes are three regular executable scripts; there is no symlink or submodule entry.

## 3. Repository map and audit method

At audit start the worktree was clean, on local branch `feat/wave-01-build-control`, at the evidence child. There is one worktree and no configured remote. The candidate repository comprises root build/plugin metadata plus `.github/`, `docs/`, `harness/`, `internal/`, `references/`, `scripts/`, and protected prototype `skills/`. The only Go package directories are `internal/harness` and `internal/harness/buildcontrol`.

Wave 1 implementation is confined to `internal/harness/buildcontrol`, strict harness data, the Makefile/configured-CI/import-check integration, and Wave 1 governance artifacts. The candidate has no `go.sum`, `vendor/`, product command, updater path, product runtime package, generated product/package path, grant runtime, or production dependency. `go.mod` and `harness/modules.lock.tsv` agree on `github.com/addressanup/level7-dev-loop`; updater identity remains `reserved`/`UNSET`. A bounded credential-pattern scan returned no candidate match; that is not a comprehensive secret audit.

The review inspected the complete `l7-release` skill and repository `AGENTS.md`, then inspected:

- the change contract, specification, design, digest-binding amendment, approval, module decision, inert grant proposal, source NO-GO audit, remediation record, candidate evidence, and registered ownership/path records;
- exact base, candidate, evidence-child, source-audit, and remediation Git objects;
- all production build-controller source, both finding-specific patches, permanent tests, `Makefile`, configured workflow, import gate, module state, README claims, and the relevant Go 1.26.7 `filepath.WalkDir` implementation; and
- recorded baseline/shadow evidence and existing ignored reproducibility artifacts, without executing cache-writing verification.

Representative read-only commands were:

```text
git status --short --untracked-files=all
git rev-parse <commit>^{commit|tree} ; git show -s --format='%H%n%T%n%P%n%s' <commit>
git diff --name-status --stat ee181b7..64eee79
git diff --name-status 64eee79..8f82512
git diff --check ee181b7..64eee79
git ls-tree -r -l <base|candidate|evidence-child>
git show <commit>:<path> | shasum -a 256
git diff-tree --no-commit-id --name-status -r <remediation-commit>
git worktree list --porcelain ; git remote -v
rg -n <identity, contract, source, test, special-node, bound, dependency, and claim queries>
nl -ba <reviewed source/test/artifact>
find .cache/repro ... -exec shasum -a 256 {} \;
```

Corrected in-memory shell loops independently hashed every candidate/base/foundation manifest row and compared sorted path sets. A separate read-only parser derived `163` unique definitions and `163` unique owner/allocation records, zero duplicates/missing/unknown IDs, and allocation `140/18/5`. Direct inventories reproduce `19` support rows, `12` invocable skills, `12` disposition rows, `43` ownership rows, and `17` authoritative orchestration ownership classes.

Two preliminary read-only shell probes were discarded: one used zsh's special `path` variable inside a loop, and one omitted braces around a `commit:path` expansion. Corrected probes produced the results above. Neither probe mutated the filesystem, Git, caches, settings, or environment.

## 4. Source NO-GO and remediation disposition

### 4.1 `AUD-W01-008` — adequately closed

`internal/harness/buildcontrol/report.go:32-69` now keeps the lexicographically smallest `51` findings regardless of arrival order and uses one fieldwise comparator for collection and rendering. This removes the prior dependency on unordered map traversal while preserving the memory cap. `internal/harness/buildcontrol/testutil_test.go:112-141` creates `163` missing-owner findings from a Go map in two distinct child processes, requires nonzero status and byte-identical output, checks the retained boundary, and requires `BCTL-099`.

The production algorithm, cap signaling, separate-process fixture, and finding-specific commit scope satisfy the source audit's required disposition for this exact candidate. No unresolved finding remains against `AUD-W01-008`.

### 4.2 `AUD-W01-009` — not fully closed

The remediation materially hardens regular fixed-file reads. `internal/harness/buildcontrol/load.go:27-149` validates canonical rooted paths; checks every component; rejects non-regular and multi-link inputs; uses `os.Root`, no-follow/nonblocking opening, and pre/open/post identity, mode, link-count, modification-time, and size comparisons. `internal/harness/buildcontrol/claims_test.go:232-297` covers final/intermediate symlink, FIFO, hardlink, and deterministic in-window mutation behavior.

That protection is not applied to every fixed authoritative input path. `checkClaims` calls `loadSkillInventory` at `claims.go:82-98`; `loadSkillInventory` opens `filepath.Join(root, "skills")` with plain `os.Open` at `claims.go:169-175`. `main` completes `checkClaims` before calling `checkPolicy` at `main.go:42-45`. Therefore a FIFO substituted at the fixed `skills/` inventory root can block in `os.Open` before either the strict reader or the repository scan reports `BLOCKED`. A symlinked `skills/` root can cause out-of-root directory entries to be consumed before the per-file strict reader rejects the intermediate link.

The new `TestStrictInputRejectsSymlinkFIFOAndHardlinkBeforeRead` invokes `readStrictFile` directly; `TestSkillInventoryRejectsEntryOverflowBeforeReadingSkills` at `claims_test.go:299-314` uses a normal directory and covers only the `65`-entry threshold. No committed fixture exercises `loadSkillInventory` or the full controller with a symlink/FIFO inventory root. The source audit's fixed-input, special-node, bounded completion, and end-to-end fixture requirements are therefore incomplete.

### 4.3 Prior finding matrix

| Source finding | Second-successor disposition |
|---|---|
| `AUD-W01-001` — adversarial matrix | `NOT FULLY CLEARED`: most listed cases exist, but fixed inventory-root special-node behavior and pre-materialization repository bounds lack end-to-end proof (`AUD-W01-012`, `AUD-W01-013`). |
| `AUD-W01-002` — resource bounds | `NOT FULLY CLEARED`: fixed-file, byte, line, expansion, finding, and post-enumeration counts are bounded; inventory-root opening and directory preloading remain unbounded. |
| `AUD-W01-003` — claim boundary | `CLEARED` for the exact candidate: the `19` exact claim rows, source-bound priority semantics, and distinct false-claim fixtures remain present. |
| `AUD-W01-004` — ownership completeness | `CLEARED` for the exact candidate: `43` disjoint controls cover all `17` authoritative orchestration classes, with distinct requirements/allocation owners. |
| `AUD-W01-005` — success schema | `CLEARED` for the exact candidate: version plus `12` exact source digests remain emitted and fixture-bound. |
| `AUD-W01-008` — capped failure determinism | `CLEARED` for the exact candidate. |
| `AUD-W01-009` — fixed input reads | `NOT FULLY CLEARED`; see `AUD-W01-012`. |

## 5. Recorded verification evidence and limits

The exact evidence child records:

| Recorded command | Recorded result | Reproduced recorded SHA-256 |
|---|---|---|
| `make verify GO_VERSION=1.26.7` | `PASS` | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| `make verify GO_VERSION=1.27.0` | `PASS` shadow development evidence | `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a` |

The evidence records zero dependencies; controller success with `12` source digests and the exact candidate-manifest digest; `163/140/18/5`, `12`, `43`, `103`, and `37` totals; a two-package import graph; compile/typecheck/unit/adversarial/vet/format/shell success; and repeat-build equality. Source inspection confirms the named tests are committed and normal `go test ./...` tests. Existing ignored `.cache/repro` contains `22` binaries matching each recorded digest.

This audit did not rerun `make`, Go build/test/vet, shell integration, or hosted CI because those commands intentionally write caches, temporary fixture roots, or environments outside the sole authorized audit-file mutation. The ignored cache is same-user mutable and does not independently prove run provenance. Also, `Makefile:145-152` computes the recorded reproducibility digest for `./internal/harness`, not for `internal/harness/buildcontrol`; the build-control package is compiled/tested elsewhere in `make verify`, but the reported binary digest is not a digest of the controller. These are evidence limitations, not independent test failures.

## 6. Severity model and findings

| Severity | Audit meaning |
|---|---|
| `BLOCKER` | Candidate/evidence identity cannot be reproduced or reviewed reliably. |
| `CRITICAL` | Demonstrated catastrophic authority, integrity, or external-effect failure. |
| `HIGH` | Central R3 boundary or checkpoint criterion is materially bypassable or absent. |
| `MEDIUM` | Required correctness/reliability contract is incomplete and must be corrected before the checkpoint. |
| `LOW` | Bounded conformance/maintainability defect that does not independently block the checkpoint. |
| `INFO` | Verified fact, limitation, or future gate with no candidate defect established. |

Finding counts: `BLOCKER 0`, `CRITICAL 0`, `HIGH 0`, `MEDIUM 2`, `LOW 0`, `INFO 2`.

### `AUD-W01-012` — `MEDIUM` — Prototype inventory root bypasses strict fixed-input validation

**Affected acceptance:** rooted/fail-closed/bounded portions of `W01-AC-006`; required special-node and degraded fixtures in `W01-AC-007`; consequently `W01-AC-012`.

**Evidence:** `claims.go:169-175` uses plain `os.Open` and `ReadDir` on `skills/`; `main.go:42-45` runs that path before `checkPolicy`; `claims_test.go:232-314` covers strict regular-file inputs and normal-directory entry overflow but not inventory-root type/link behavior.

**Impact:** a FIFO can block indefinitely before a stable nonzero diagnostic; a symlink can expose external directory-entry names before rejection. The current exact candidate contains a normal `skills/` directory, so no current external read or hang occurred. The defect is nevertheless material to the mandatory behavior under adversarial repository state.

**Required disposition:** acquire the fixed inventory root through a rooted, no-follow, nonblocking, identity-checked directory path; reject symlink/special/replaced directory state before enumeration; and add time-bounded `loadSkillInventory` plus end-to-end controller fixtures for symlink, FIFO/special-node, and replacement cases. Rebind and fully reverify a new candidate/evidence chain.

### `AUD-W01-013` — `MEDIUM` — Repository count bounds apply after whole-directory materialization

**Affected acceptance:** bounded/degraded portions of `W01-AC-006`; mandatory resource-bound fixtures in `W01-AC-007`; consequently `W01-AC-012`.

**Evidence:** `policy.go:199-267` calls `filepath.WalkDir`; its `filesSeen` and `directoriesSeen` caps execute only inside the callback. The pinned Go 1.26.7 implementation at `.cache/toolchains/go1.26.7/src/path/filepath/path.go:310-319` calls `os.ReadDir(path)` before iterating callbacks, which materializes and sorts the entire directory. `policy_test.go:170-197` creates `513` entries and proves the later rule fires, but it does not prove a bound before enumeration/allocation. There is no controller wall-clock deadline.

**Impact:** one directory with a sufficiently large entry set can consume memory and latency beyond the declared `512`-entry controller budget before `SCOPE-346`/`SCOPE-348` can return a bounded failure. This is denial of bounded diagnostic completion, not a demonstrated policy pass or mutation.

**Required disposition:** replace whole-directory materialization with rooted streaming/batched enumeration that stops before exceeding the global directory/file budget, define the required time/deadline behavior without creating a pass-on-timeout path, and add a fixture proving the pre-enumeration resource ceiling and stable nonzero result. Rebind and fully reverify a new candidate/evidence chain.

### `AUD-W01-014` — `INFO` — Recorded local verification was inspected, not rerun

Candidate/evidence bytes, committed tests, recorded digests, and matching ignored binary hashes were inspected. No cache-writing or temporary-fixture command was rerun under this audit authority. The recorded passes remain same-user local development evidence.

### `AUD-W01-015` — `INFO` — External qualification and release gates remain separate

Hosted Ubuntu CI, authenticated GitHub repository/account/remote identity, actual Codex/Claude lifecycle and differential conformance, Controlled Client qualification, protected evaluation, grant security review and normative adoption, pilot/stable grants, AP2/AP3, signing, TUF, promotion, publication, release, deployment, exposure, and monitoring remain `NOT_RUN` or absent. The inert proposal and this audit clear none of them.

## 7. Wave 1 acceptance disposition

| Acceptance | Independent disposition |
|---|---|
| `W01-AC-001` | `PASS`: independent parsing reproduced `163` definitions, one owner/allocation each, zero duplicate/missing/unknown IDs, totals `140/18/5`; permanent trace and summary-tamper fixtures are present. |
| `W01-AC-002` | `PASS`: the `19` exact support rows preserve the two-product, A0-A2/A3-A5, three-profile, development-only boundary. |
| `W01-AC-003` | `PASS`: `12` invocable prototype skills have exactly `12` dispositions; all protected manifest rows and bytes reproduce. |
| `W01-AC-004` | `PASS`: stable version, dual-host, enforcement, compatibility, support, and release promotion remain withheld. |
| `W01-AC-005` | `PASS`: P0/P1/P2 and scope/priority impact-approval rules remain source-bound. |
| `W01-AC-006` | `NOT CLEARED`: inventory-root special-node handling can block/consume before rejection, and repository count bounds apply after entire-directory materialization (`AUD-W01-012`, `AUD-W01-013`). |
| `W01-AC-007` | `NOT SATISFIED`: the permanent suite lacks the required end-to-end inventory-root special-node cases and a pre-materialization repository-bound case. |
| `W01-AC-008` | `PASS` for local development identity only: module records agree; updater is reserved; GitHub identity remains `USER_ASSERTED`, remote-absent, and unpublished. |
| `W01-AC-009` | `PASS` as an inert proposal only: non-interchangeability and separate review/adoption gates are explicit; no source loads or activates it. |
| `W01-AC-010` | `PASS`: `43` disjoint controls cover `17` authoritative ownership classes and exclude candidate authority over protected controls. |
| `W01-AC-011` | `PASS` as recorded same-user local evidence only: both exact matrices are recorded green with effects/limits; hosted CI remains truthful as `NOT_RUN`. |
| `W01-AC-012` | `FAIL`: exact identities and audit separation reproduce, but two unresolved `MEDIUM` findings prevent the R3 development gate. |
| `W01-AC-013` | `PASS` for the exact candidate: no dependency, product/runtime/updater path, prototype edit, unexpected registered path, external effect, stable claim, or bounded credential-pattern match was observed. |

No aggregate passing result waives an unsatisfied criterion or material audit finding.

## 8. Stopping boundary

The candidate and evidence child remain frozen. Correcting either material finding requires a separately authorized remediator, a new finding-specific candidate, regenerated manifest, complete baseline/shadow verification, a new evidence-only child, and another fresh independent audit. This reviewer performed no remediation and authorizes no merge, Wave 2 work, remote/hosted action, publication, release, deployment, exposure, grant activation, or external effect.
