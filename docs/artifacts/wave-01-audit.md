# Level 7 Dev Loop — Wave 1 Fourth-Successor Independent Audit

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-W01-005` |
| Artifact type | Independent Principal Engineer release audit — `level7-dev-loop:l7-release` Mode B |
| Version | 0.5.0 |
| Date | 2026-08-26 |
| Accountable owner | Anup Pandey |
| Reviewer role | Fresh, structurally separate, read-only Principal Engineer reviewer |
| Approved base | Commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Audited candidate | Commit `a1f146cdb5b2f20a7852bcf490223541fe4c8986`; tree `db580f77234dc14289f22174760d6da9bf442891`; parent `4b092a65e74d713346975de5bb4d78d161ad2b0a` |
| Evidence-only child | Commit `06345424e455a57b06c183b38a8492d20580c2bf`; tree `15f69b94bb0d943e7979fc9df14aedf13c89181b`; parent is the audited candidate |
| Source independent audit | Commit `62e1a019eb6a75748c628c93102c41db81166d28`; SHA-256 `16f11e11b466a78cb7bf758cff40b7e0f7e85057e73af217bb69649795003917` |
| Finding remediation commits | `6c04c537fa3a1af2a0ba0ab3db469b99d8852593`; `4b092a65e74d713346975de5bb4d78d161ad2b0a` |
| Verdict | **NO-GO** |
| Wave 1 checkpoint | `W01-AC-012` is **not cleared** |

## 1. Decision

The exact base, candidate, evidence child, ancestry, trees, artifact digests, 34-row candidate manifest, and 37-path base-to-candidate closure all reproduce. The evidence child changes only `docs/artifacts/wave-01-evidence.md`. The fourth-successor implementation adequately closes `AUD-W01-016`: the aggregate batch condition is evaluated before sorting or entry inspection, and permanent regressions cover real-filesystem creation order, mixed capped subsets, and cross-process exit/output stability.

`AUD-W01-017` is not adequately closed. The approved successor effect model requires temporary fixtures and retained verifier state to remain physically below the repository's ignored `.cache` root. The implementation proves only lexical path equality and a lexical prefix. It neither rejects nor resolves symlinked cache components. Because `.cache` is excluded from the repository scanner, a symlink such as `.cache/go/tmp` can redirect `t.TempDir`, build, and test temporary writes outside the repository while the new regression passes. `make prepare` also writes through cache paths before the subsequent toolchain checks. This is an unresolved `MEDIUM` correctness finding. Under the approved no-material-finding threshold, the exact candidate cannot clear `W01-AC-012`.

Finding count: `BLOCKER 0`, `CRITICAL 0`, `HIGH 0`, `MEDIUM 1`, `LOW 1`, `INFO 2`.

## 2. Audit method and repository map

This review was read-only until this registered reviewer artifact was updated. All Git reads used `--no-optional-locks`. The initial and pre-write `git --no-optional-locks status --short` results were empty. No verifier, test, build, controller, mutation probe, network action, Git/index/ref mutation, cache cleanup, or hosted action was run.

Principal read-only commands included:

```text
git --no-optional-locks status --short
git --no-optional-locks rev-parse <commit>^{tree} <commit>^ <commit>^1
git --no-optional-locks show <commit>:<path> | shasum -a 256
git --no-optional-locks diff --name-status <base> <candidate>
git --no-optional-locks diff --numstat <candidate> <evidence-child>
git --no-optional-locks ls-tree -r -l <commit>
git --no-optional-locks diff --check <base> <candidate>
rg --files
rg -n <targeted-policy-and-effect-patterns> <candidate-files>
```

An independent read-only parser also checked manifest ordering, uniqueness, row hashes, set closure, ownership uniqueness, and fixture/accounting totals directly from committed bytes.

| Area | Audited map |
|---|---|
| Root/control | 10 root files plus 3 host/plugin/workflow control files |
| Documentation | 46 committed documentation files in the candidate |
| Harness data | 12 committed files in `harness/` |
| Go implementation | 16 files under `internal/harness/`; packages `harness` and `main` only |
| Skills | 12 strict skill packages; 163 definitions, 163 unique owners, no missing/unknown/duplicate ownership |
| Scripts/references | 3 harness scripts and 1 reference file |
| Candidate inventory | 103 blobs; 1,269,272 bytes; 100 mode `100644`, 3 mode `100755`; no symlink or submodule entry |
| Base delta | 37 files; 31 additions, 6 modifications; 5,693 insertions, 17 deletions; `git diff --check` clean |
| Product/dependency scope | No candidate `go.sum`, `vendor/`, `cmd/l7/`, updater implementation, or product package path |
| Git topology observed | One worktree on `feat/wave-01-build-control`; no configured remote |

## 3. Identity, digest, manifest, and closure reproduction

| Evidence | Independent result |
|---|---|
| Approved base | Commit and tree exactly reproduced: `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4` / `2f23a0810660995b6f562c361ab38cd4faafa3b3` |
| Candidate | Commit, tree, and parent exactly reproduced: `a1f146cdb5b2f20a7852bcf490223541fe4c8986` / `db580f77234dc14289f22174760d6da9bf442891` / `4b092a65e74d713346975de5bb4d78d161ad2b0a` |
| Evidence child | Commit, tree, and parent exactly reproduced: `06345424e455a57b06c183b38a8492d20580c2bf` / `15f69b94bb0d943e7979fc9df14aedf13c89181b` / candidate |
| Evidence-only delta | Exactly one modified path, `docs/artifacts/wave-01-evidence.md`; numstat `60 59` |
| Evidence SHA-256 | `8e1103ad832d27e016593e6d66ffc5c25015ef8974312d315e0401bfd887727b` |
| Candidate manifest SHA-256 | `1833ec87308735bdb5cbbe47f12c26c75596657a119c19d27c639ab9121c44cb` |
| Remediation record SHA-256 | `e147c48f204118172e36d9e234e424516f7f1e06c638d4c3fac65dec0f08293b` |
| Successor design amendment SHA-256 | `e10378f598098d5db8e9f20177324e917260e0ce016453903ac0159485526470` |
| Source audit | Commit `62e1a019eb6a75748c628c93102c41db81166d28`; artifact SHA-256 `16f11e11b466a78cb7bf758cff40b7e0f7e85057e73af217bb69649795003917` |
| `AUD-W01-016` remediation | Commit `6c04c537fa3a1af2a0ba0ab3db469b99d8852593`; parent is the source audit; exact subject `fix(audit-AUD-W01-016): stabilize bounded enumeration`; only `policy.go` and `policy_test.go` changed |
| `AUD-W01-017` remediation | Commit `4b092a65e74d713346975de5bb4d78d161ad2b0a`; parent is the `AUD-W01-016` remediation; exact subject `fix(audit-AUD-W01-017): bind verifier temp effects`; only `Makefile`, the design amendment, `policy.go`, and `testutil_test.go` changed |

The candidate manifest contains exactly 34 data rows. It is byte-sorted and unique. Every recorded row hash matches the corresponding blob read from the exact candidate commit. Removing the manifest's self-row exclusions — `docs/artifacts/wave-01-candidate.sha256`, `docs/artifacts/wave-01-evidence.md`, and the audit-only `docs/artifacts/wave-01-audit.md` — from the 37-path closure produces exactly those 34 paths.

`harness/wave-01-paths.tsv` contains exactly 37 data rows. Its path set equals `git diff --name-only` from approved base to candidate. Its declared change classes equal Git's result for every path: 31 `A`, 6 `M`, no deletion. The 72-row approved-base manifest independently reproduces against the base tree and covers that tree exactly. The 26-row foundation manifest also reproduces against the candidate.

## 4. Contract and architecture evaluation

The following committed authorities were inspected as a single contract:

- `docs/artifacts/wave-01-specification.md`, including untrusted-input, canonical-path, no-follow file-shape, offline/no-credential, deterministic-order, resource-bound, ignored-worktree, fixture/effect, acceptance, non-waiver, and material-audit clauses;
- `docs/artifacts/wave-01-design.md`, including deterministic rendering, no-follow acquisition, cache/effect claims, permanent negative regressions, and independent-audit threshold;
- `docs/artifacts/wave-01-design-amendment.md`, especially §4.2's physical repository-scoped effect envelope and §4.3's stable `SCOPE-338` semantics;
- `docs/artifacts/wave-01-change-contract.md`, approval, module-identity decision, grant-ladder amendment, source audit, and remediation record; and
- candidate implementation, permanent regressions, candidate manifest, path closure, and the evidence-only child.

The specification requires deterministic decisions independent of unordered filesystem traversal and blocks aggregate waiver of a material acceptance failure. The design amendment expressly permits real filesystem/process fixtures only inside a repository-scoped root and permits only the current test binary or pinned repository-local Go. These are release criteria, not documentary aspirations.

The implementation otherwise remains appropriately inert: no product command, updater, dependency graph, provider action, remote, publication, release, or deployment surface appears in the candidate. The grant amendment is referenced only as controlled artifact/ownership input and has no parser or activation path.

## 5. Source-audit remediation disposition

### `AUD-W01-016` — adequately closed

`internal/harness/buildcontrol/policy.go:226-247` computes a bounded `entryLimit`, performs the bounded read, and emits directory-scoped `SCOPE-338` when the returned batch reaches that limit. The branch breaks at lines 243-245. Sorting starts only at line 247 and entry iteration/inspection only at line 250. Therefore the full sentinel batch is rejected before subset sorting or `DirEntry.Info` inspection.

Permanent regressions provide the required independent dimensions:

- `TestRepositoryScanCapsSingleDirectoryReadBeforeEnumeration` (`policy_test.go:200-239`) creates 1,027 real files in ascending and descending order, requires the identical complete result, asserts the exact requested read cap, and asserts zero retained files;
- `TestRepositoryScanMixedEntryBatchIsOrderIndependent` (`policy_test.go:241-271`) supplies directory-omitted and file-omitted capped subsets and uses entries whose `Info` fails if inspected;
- `TestRepositoryEntryBatchFailureIsStableAcrossProcesses` (`policy_test.go:273-305`) executes separate child processes and requires exit code 1 with byte-identical rule, subject, message, and recovery output; and
- `syntheticRepositoryBatch` (`policy_test.go:307-343`) covers forward/reverse mixed entry construction.

The exact source order supplies the critical “before sorting” proof. The mixed-subset test makes premature `Info` observable but does not make a premature call to `Name` observable, so it would not alone detect moving the sort above the aggregate check. That is a non-material regression-strength limitation for this exact implementation.

### `AUD-W01-017` — not adequately closed

The amendment's §4.2 says temporary files and retained verifier effects must remain under the ignored repository-scoped root. The remediation adds useful lexical assertions but not a physical containment invariant:

- `Makefile:55-60` constructs `GOPATH`, `GOCACHE`, `GOMODCACHE`, `GOTMPDIR`, and `TMPDIR` from lexical repository paths.
- `Makefile:84-86` creates those paths and writes telemetry mode before `toolchain-check`; `Makefile:112-113` later compares strings only.
- `TestTemporaryRootsAreRepositoryScoped` (`testutil_test.go:25-37`) checks exact strings, absoluteness, `filepath.Clean`, and `strings.HasPrefix`. It does not use no-follow component acquisition, `Lstat`, `EvalSymlinks`, or file identity comparison.
- `policy.go:275-277` intentionally skips the top-level `.cache` directory, so its scanner cannot reject a redirected nested cache root.
- The pinned Go testing implementation ultimately creates temporary directories below `GOTMPDIR` using normal path traversal; a symlinked component is followed.

Consequently, an ignored same-user symlink at `.cache/go/tmp` can make the lexical `t.TempDir` assertion pass while placing files outside the repository. Similar cache/telemetry component redirection exists, and `prepare` writes before a later test could reject it. The currently observed `.cache`, `.cache/go`, temporary, telemetry, and repro paths were real directories, but mutable ambient state is not a candidate invariant. No mutating symlink probe was authorized or run.

## 6. Recorded verification and independent inspection

| Recorded exact-candidate evidence | Result |
|---|---|
| `make verify GO_VERSION=1.26.7` | `PASS`, 10.41 seconds, trace digest `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| `make verify GO_VERSION=1.27.0` | `PASS`, 10.72 seconds, trace digest `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a` |

The recorded evidence binds the exact candidate and records matching baseline/shadow verification for both pinned toolchains. Existing `.cache/repro` output includes matching current-digest pairs for all 13 verifier steps for each toolchain. The pinned toolchain `VERSION` files identify `go1.26.7` and `go1.27.0`; the inspected pinned `os/tempfile.go` bytes are identical between them for the relevant creation behavior.

The reviewer did not rerun `make verify` or tests because the audit authority prohibited commands that write caches, temporary environments, or fixtures. Passing recorded tests cannot waive the physical containment defect described above.

## 7. Severity model and findings

| Severity | Meaning used here |
|---|---|
| `BLOCKER` | Release cannot be evaluated safely or candidate identity/authority is invalid |
| `CRITICAL` | Immediate catastrophic integrity, security, or irreversible-effect exposure |
| `HIGH` | Major release-control failure with broad or likely impact |
| `MEDIUM` | Material contract/correctness failure that blocks the Wave 1 audit threshold |
| `LOW` | Real but non-material weakness for this checkpoint |
| `INFO` | Evidence boundary or later gate, not a candidate defect |

### `AUD-W01-020` — `MEDIUM` — Repository-scoped verifier effects are only lexically contained

The approved effect model requires physical repository containment. Symlinked ignored cache components can redirect temporary/cache/telemetry writes outside the repository while the Makefile and test assertions accept their lexical paths. Because preparation writes occur before later checks and `.cache` is deliberately outside policy traversal, this is an unresolved material correctness failure affecting `W01-AC-012` and `W01-AC-013`. It does not negate the directly observed, repository-contained baseline and shadow runs evaluated under `W01-AC-011`.

### `AUD-W01-021` — `LOW` — Process-fixture inventory and environment isolation are incomplete

The amendment permits only the current test binary or pinned local Go, yet `policy_test.go:487-492` directly executes `/bin/sh`, and `testutil_test.go:241-262` builds and executes a new controller binary. Those subprocesses inherit `os.Environ` before selected overrides (`policy_test.go:489`; `testutil_test.go:243,250`). Inspection found no candidate credential consumption, network call, or external sink, so this is a least-authority and documentary-accuracy weakness rather than a demonstrated material release effect.

### `AUD-W01-022` — `INFO` — Local verification evidence was inspected, not rerun

The recorded verifier outputs, candidate binding, and existing reproduction artifacts were inspected. The audit did not regenerate them under the read-only constraint. Existing build/test binaries are same-user local evidence; the trace digest binds controller sources and declared inputs, not the bytes of the already-built controller executable.

### `AUD-W01-023` — `INFO` — Hosted and later release gates remain outside this audit

No hosted qualification, merge, publication, release, deployment, exposure, monitoring, or later-phase gate was invoked. This artifact grants none of those authorities.

## 8. Wave 1 acceptance criteria

| Criterion | Independent result |
|---|---|
| `W01-AC-001` | `PASS` — independent source parsing reproduces exactly 163 normative IDs, one unique accountable owner/allocation per ID, and totals `140/18/5` |
| `W01-AC-002` | `PASS` — the 19-row narrow support matrix preserves the two-product distinction, A0–A2/A3–A5 boundary, and all three proof profiles without promotion |
| `W01-AC-003` | `PASS` — all 12 prototype skills have exactly one disposition, and every byte in the 26-row protected foundation manifest reproduces unchanged |
| `W01-AC-004` | `PASS` — stable version, compatibility, enforcement, and support-promotion claims remain withheld; candidate claims remain development/prototype scoped |
| `W01-AC-005` | `PASS` — priority/scope changes remain gated by an impact diff and accountable approval, and no metric or date is treated as a safety waiver |
| `W01-AC-006` | `PASS` — the phase-aware successor preserves historical Step 5 evidence and fails closed on unknown, malformed, stale, and unowned input; aggregate enumeration is now deterministic before inspection |
| `W01-AC-007` | `PASS` — permanent positive and adversarial fixtures cover the governed contract, including real-filesystem order, mixed subsets, cross-process output, and bounded failure, before any new governed capability |
| `W01-AC-008` | `PASS` — the exact module-identity decision precedes product imports, no product import exists, and updater identity remains reserved |
| `W01-AC-009` | `PASS` — the grant-ladder amendment remains inert, non-interchangeable, separately auditable, and separately approvable |
| `W01-AC-010` | `PASS` — 43 ownership controls are complete, unique, and disjoint for the governed classes and exclude candidate authority over protected assets |
| `W01-AC-011` | `PASS` — exact-candidate baseline and shadow local matrices are recorded green; observed temporary/cache roots were real repository directories, effects and reproducibility limits are recorded, and hosted CI remains truthfully `NOT_RUN`. `AUD-W01-021` remains a non-material process-description limitation |
| `W01-AC-012` | `FAIL` — exact candidate/evidence identities and manifest closure reproduce, but one unresolved `MEDIUM` finding violates the independent-audit clearance threshold |
| `W01-AC-013` | `FAIL` — no dependency, product behavior, prototype edit, unexpected registered path, stable claim, or secret was found, but physical no-external-effect containment is not enforced for ignored cache and temporary roots |

No aggregate passing result or recorded test pass waives `W01-AC-012`, `W01-AC-013`, or `AUD-W01-020`.

## 9. Authority boundary

This audit records the decision for the exact candidate only. It modifies no candidate code, test, configuration, evidence, manifest, design, protected historical audit, Git state, cache, toolchain, remote, or external system. It authorizes no merge, publication, release, deployment, exposure, remediation, or later-phase action.
