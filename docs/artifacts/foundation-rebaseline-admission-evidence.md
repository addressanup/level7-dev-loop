# Level 7 Dev Loop — Foundation Rebaseline Admission Evidence

| Field | Value |
|---|---|
| Artifact ID | `L7-FRB-ADM-EVD-001` |
| Version | `0.1.0` |
| Date | `2026-08-27` |
| State | `admitted-awaiting-assurance` |
| Gate 2 candidate | `L7-FRB-CAND-001` |
| Candidate-manifest SHA-256 | `97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40` |
| Source commit | `1c5c351f52f258d37ba48d8348e1cd883d2fb250` |
| Source tree | `b1fe4753b51b0da847d73b0ff64377fb2bda1434` |
| Admission control commit | `baac74bae2f4a60141213d86a360e73da22541de` |
| Admission control tree | `572e5f4941c8de0da64942b9f073550f0be2ba63` |
| Branch | `feat/foundation-rebaseline` |
| Maximum observed effect | `A2`; repository-local governance/control writes, local Git commit, cache, and temporary test effects |
| Network effect | `NONE`; harness environment set `L7_NETWORK=off`, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, and `GOAUTH=off` |
| Local verification | `PASS` |
| Independent admission assurance | `NOT_RUN` |
| Wave 2 evidence/audit | `ABSENT` |

## Outcome

The exact owner-approved Gate 2 payload was promoted byte-for-byte on the exact successor branch. The deterministic controller now reconstructs `foundation-rebaseline` as the sole active phase, binds the 174-file source inventory and key predecessors, enforces the exact 69-row path policy and its stage windows, records historical staleness without changing predecessor bytes, and reports `admitted-awaiting-assurance` when this evidence exists but the separately owned admission audit does not.

This local verification does not satisfy independent assurance. A genuinely separate reviewer must inspect the committed candidate read-only and write only `docs/artifacts/foundation-rebaseline-admission-audit.md`. Until a bound `GO` exists, no requirements candidate is authorized by repository state or this evidence.

## Verification results

| Check | Result | Evidence | Limitation |
|---|---|---|---|
| Exact source inventory | `PASS` | Raw SHA-256 comparison of every `git show 1c5c351:<path>` payload against `harness/foundation-rebaseline-base.sha256`; 174/174 matched; source tree `b1fe4753b51b0da847d73b0ff64377fb2bda1434` | Local repository and Git object store are not tamper-proof |
| Gate 2 bundle | `PASS` | Manifest hash `97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40`; four canonical payload digests matched | Confirms bytes and owner decision, not product correctness |
| Path envelope | `PASS` | Policy hash `09513eab93c254c50a5cae2704786a62a9d3a61f02103c93d28706f8c49f6ecc`; 69 sorted unique rows: 56 add, 13 modify | Later windows remain locked and unexecuted |
| Admission snapshot | `PASS` | Controller reported 186 regular files and 19 changes against the exact base, all in the admission window | Evidence file was intentionally added after this pre-evidence run and is rechecked below |
| Phase and ownership | `PASS` | One active `foundation-rebaseline` phase; 69 shared-control rows; evidence writer and audit writer are disjoint | Local role labels do not cryptographically prove human independence |
| Historical integrity | `PASS` | Key predecessor manifest matched the complete base; Wave 2 candidate retained without evidence or audit | Historical status is a successor classification, not a rewrite of old text |
| Negative controls | `PASS` | Tests rejected path/window expansion, premature requirements mutation, AP0 replay/tamper, fabricated Wave 2 completion, and candidate-writer self-audit | Separate reviewer must assess whether test coverage is sufficient |
| Offline full harness | `PASS` | `make verify`: install, lint, typecheck, tests, import boundaries, and reproducibility all passed | This is local integrator evidence, not separate assurance |
| Reproducible harness binary | `PASS` | Both builds matched SHA-256 `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` | Applies to the current harness test binary only |
| Diff hygiene | `PASS` | `git diff --check` returned no findings before the admission control commit | Does not replace semantic review |

## Admission control closure

The following SHA-256 records bind every path changed between the exact source commit and admission control commit `baac74bae2f4a60141213d86a360e73da22541de`. This evidence file is intentionally outside that pre-evidence closure and is bound by its enclosing successor commit for the separate reviewer.

```text
b025d1e4c97b2dd208e2047ee7bfe0ddce759e673be0c6822313019399f64369  docs/artifacts/foundation-rebaseline-approval.md
97556c4741d6b576079dd31a398ebf227d2a565cdc71ac3709163f07c54c8a40  docs/artifacts/foundation-rebaseline-candidate.sha256
8c297db289bd9f405ccdec9f33448fb81def5f6334ba70a23c0306cbd3aa68e8  docs/artifacts/foundation-rebaseline-change-contract.md
7621541e0319dc5a2c238ea57a3841b3bfb62a0bb3b46d8371bb3bb3ba54b23a  docs/artifacts/foundation-rebaseline-design.md
4a208c68d7ff559c52db53b5907f2f4bcec8226be459330f9e9e2bf6d406ba2b  docs/artifacts/foundation-rebaseline-history.md
ff9ca1d03a21533e000bb586796ea4367f95df59a321900b5044ce71d6dfebc9  docs/artifacts/foundation-rebaseline-specification.md
cab02a5d9235488893c0b4525a66ed8ec10d96fa457aa1c364f0379f0d6172de  harness/control-ownership.tsv
176f3725b6c801b34a3a3208efc8c6609b854fe483f8bc2c6cd167170261aa47  harness/foundation-rebaseline-base.sha256
13a455175a28dd76f7c6b0313e4868515bcbfe0c86fa536b3966d05ecb460ecb  harness/foundation-rebaseline-gates.tsv
09513eab93c254c50a5cae2704786a62a9d3a61f02103c93d28706f8c49f6ecc  harness/foundation-rebaseline-paths.tsv
3b8be699cb39824b0c56498bd2b71dad08035f8668c17eef6b6c4e2a17cdbd4d  harness/foundation-rebaseline-predecessors.sha256
45d9cb5cb52d7c5072e676a3ce653b7a7082a7b434863319c9794f4fb416e76e  harness/phases.tsv
c3b200008c5046fd071a885e8048fe9e0365be9b0c9f9b9ff36b997885fc4b9c  internal/harness/buildcontrol/foundation.go
9dd6b6fbbfd1b7a8f3dfaf21705e110283f64372d0a75f05c92f68d466a1ff87  internal/harness/buildcontrol/foundation_test.go
93d80ddf6c589a8a5864cab9d4f87490da165b82fc1e312df4ad50773dfb63ed  internal/harness/buildcontrol/main.go
3af56d8a02f6583e9d59157f7cca735f090f6eba5fdb59746d9eb8610004bdb2  internal/harness/buildcontrol/ownership.go
ee9e7b6d975d5def1a433c9ae23bce731c1d2dee9269dd082e2f966e9677f441  internal/harness/buildcontrol/ownership_test.go
bd302d142bb247eebbd8d12096912d7e28218843e742c6bda27dafc33f0b0ce4  internal/harness/buildcontrol/policy.go
51e0c6048145e5ee01f871c7c1e42b66057bbc6422c89ee53df355695dd29a70  internal/harness/buildcontrol/policy_test.go
```

## Truth boundary and residual risks

- What is current: the approved Concept Brief, this local foundation-governance phase, and its deterministic admission controls.
- What remains historical or stale: the former requirements-through-orchestration dependency chain and all persisted predecessor approvals as non-replayable AP0.
- What remains only planned or unverified: the corrected requirements, backlog, architecture, technology, support matrix, harness plan, orchestration, and every product/runtime capability.
- What is not claimed: no desktop client, CLI, host package, conductor, kernel, graph, RAG, learning loop, repair loop, adapter, installer, support tuple, A3/A4 workflow, or regulated pack was implemented or qualified.
- Residual risk: the integrator authored the controller and its public tests; therefore a separate read-only audit is mandatory before Gate 3 can begin.
- Recovery: any mismatch or `NO_GO` retains this evidence as history and returns to the earliest affected admission control through a new authorized commit; no history rewrite, reset, amend, or silent restoration is permitted.
