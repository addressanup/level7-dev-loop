# Level 7 Dev Loop — Foundation Harness Audit Record

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-HAR-001` |
| Artifact type | Separate-context harness, isolation, supply-chain, and evidence review |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Status | Complete — separate-context model audit `GO` for Foundation Step 5 only |
| Version | 0.1.0 |
| Date | 2026-08-24 |
| Candidate | [`L7-HAR-001` implementation manifest](harness-candidate.sha256) |
| Candidate-manifest SHA-256 | `64bba1fcfe347d27a2b05df545b753bf7dc181383d99630493db4d7a47233592` |
| Candidate entries | 20; every digest independently rechecked `PASS` |
| Review mode | Read-only separate-context model review; primary agent performed every correction |
| Assurance | Useful engineering separation only; not human security/legal review, AP2/AP3, protected CI, actual-host proof, or release independence |
| Effect | A1 audit-record write only; no product/runtime/host effect |

## 1. Final verdict

**`GO` for the exact Foundation Step 5 candidate manifest above.** No blocker or high-severity finding remains. This verdict does not extend to a modified successor, product implementation, controlled mode, host compatibility, deployment, publication, or release.

## 2. Audit scope

The reviewer checked the approved Step 5 boundary against `l7-greenfield`, `L7-TEC-001` §§18, 20, 22, and 28, and the owner approval. Review covered repository scope, Go bootstrap authenticity, shell/Make/CI portability, ambient configuration, version/platform binding, concurrency, zero-dependency truth, the inert proving test, structured logging/privacy, import-boundary enforcement, protected-file preservation, and evidence language.

The reviewer did not edit the candidate, invoke Codex or Claude, call a model/provider, initialize Git, run hosted CI, create product packages, install outside the repository cache, publish, deploy, or release.

## 3. Material findings and corrections

| Finding | Initial severity | Correction | Final state |
|---|---:|---|---|
| `GOTELEMETRY=off` did not isolate Go telemetry; the toolchain gate initially returned `local`. | High | Redirect exact toolchains through repository-scoped `TEST_TELEMETRY_DIR`, write an `off` mode, assert its resolved directory/mode, and fail if the pinned Go source stops honoring the mechanism. | Corrected; both exact versions pass. |
| Baseline and shadow reproducibility jobs shared output/cache paths and raced. | High | Namespace both build caches and outputs by version and a per-invocation `mktemp` directory. | Corrected; simultaneous full verification passes. |
| Ambient `GOROOT`, `GOOS`, and `GOARCH` could alter the actual compiler/tool/target tuple while the dedicated gate passed. | High | Override the root, native target/host, architecture level, and related Go controls; assert the actual root/tool/host/target plus toolchain, offline, and cache settings. | Corrected; hostile inherited values are normalized. |
| The proving test accepted either Go version independent of the selected matrix role. | High | Link exact expected Go version, GOOS, and GOARCH into the inert test and compare all three to `runtime`. | Corrected. |
| Reused toolchain wording implied stronger authentication than a same-user writable cache provides. | Medium | Reauthenticate the official archive and detached signature on bootstrap reuse, label the extracted tree as not tree-verified, and require fresh extraction for future gate-bearing evidence. | Corrected claim; full cached-tree verification remains deferred. |
| The updater rule named a nonexistent repository package and a future nested updater module would be skipped by the root `go list`. | High | Add an updater allow-closure plus an explicit module registry. `cmd/l7up` is `reserved`/`UNSET`; its appearance fails until a later approved module-aware harness is installed. | Corrected for Step 5; future updater activation remains blocked. |
| Pure-package rules omitted `syscall`, `runtime`, indirect internal helpers, and external-module escape; `unsafe` was not globally rejected. | Medium | Inspect internal transitive closures, reject external dependencies in pure closures, add effect-family bans, and reject product imports of `unsafe` or the test-only harness. | Corrected at policy level. |
| Reproducibility hashing and Make paths were not fully portable to Linux or paths containing spaces. | Medium/Low | Add SHA-256 tool fallback, quote executable/project paths, and fix the `.go-version` lookup. | Corrected. |
| Approval metadata called source/config harness construction A1 even though the approved taxonomy defines it as A2. | Blocker for evidence accuracy | Correct the record to bounded A2 harness construction plus its A1 decision record, then update the protected digest. | Corrected; authorization scope itself was already explicit. |

## 4. Independently checked evidence

- exact Go 1.26.7 baseline install, lint, typecheck, test, repeat-build, and local CI target: `PASS`;
- exact Go 1.27.0 shadow full verification: `PASS`;
- concurrent baseline/shadow full verification: `PASS`;
- repeat-build test-binary SHA-256: baseline `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f`; shadow `9edf1547c82d18a356c0d85d00c848bed9159f1461c469cf1b49ed3085b420c3`;
- ShellCheck, actionlint, YAML parse, protected-input digests, zero-dependency assertion, and exact-one-test assertion: `PASS`;
- all four locked Go archive records match official release metadata;
- `actions/checkout` `v7.0.1` resolves to locked commit `3d3c42e5aac5ba805825da76410c181273ba90b1`;
- protected skills, prompts, manifests, and prior approved foundation artifacts: unchanged; and
- all 20 implementation-manifest entries and the manifest digest: `PASS`.

## 5. Residual limitations

| Limitation | Severity/state | Required later gate |
|---|---|---|
| A reused extracted Go tree is writable and not fully compared to a fresh extraction. | Medium, disclosed | Use fresh authenticated extraction and protected clean builders for gate-bearing evidence. |
| Ubuntu GitHub-hosted CI has not executed; its mutable label is not an exact production tuple. | Medium, `NOT_RUN` | Git repository/remote workflow plus exact protected platform evidence later. |
| Current positive import scan observes one inert package. Two transient negative probes proved reserved-updater and pure-effect rejection, but permanent adversarial fixtures do not yet exist. | Low | Add module-specific positive/negative fixtures before the first product package. |
| Updater module identity and its exact go-tuf/transitive allowlist are deliberately `UNSET`. | Low now; blocking on activation | Approved build decision plus module-aware checker before `cmd/l7up` exists. |
| Go's detached signature uses Google's shared Linux Packages Signing Authority rather than a Go-exclusive key. | Low, disclosed | Treat official SHA-256 and fresh protected-build provenance as separate evidence layers. |
| No Git repository exists. | Operational, observed | Initialize/import only through a separately approved owner decision; then produce commit/CI evidence. |

## 6. Gate

The exact candidate is eligible for the product owner's Foundation Step 5 decision. Owner approval may authorize Foundation Step 6 orchestration planning only; it cannot convert any `NOT_RUN`/`UNPROVED` item into `PASS` or authorize feature implementation.
