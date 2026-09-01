# Level 7 v1.0 Hosted Provider Fuzz and Benchmark Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-provider-fuzz-benchmark-remediation` |
| Candidate commit | `0277b90cf01cb47120b30ac4f1f0bb8bdd00e864` |
| Candidate tree | `0721dc7c17fdfc35103b3f3a91f6a8f095306662` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-09-01T10:18:21Z` |
| Brief commit | `cbfe8e33f96e105a6d15b6db6035b25e927b4984` |
| Integration parents | Target `902a66010fe7fb4a56afd9bb1799970f83201eb9`; proposal `cbfe8e33f96e105a6d15b6db6035b25e927b4984` |

## Checks

| Check | Result |
|---|---|
| Approval binding | PASS — the active-user envelope names Product Owner Anup Pandey, implementer `codex-root`, this change ID, and exact brief commit `cbfe8e33f96e105a6d15b6db6035b25e927b4984`. No predecessor authority or assurance was reused. |
| Proposal boundary | PASS — the proposal has sole parent/base `84bd69f90d366356b0ce1e1a392f906258f3de91`, tree `316a2f0f20534e667e4ee1f8ae336d1916ee1d1f`, and adds only the approved brief. |
| Integration topology | PASS — candidate `0277b90cf01cb47120b30ac4f1f0bb8bdd00e864` is an unsigned additive merge with exact first parent `902a66010fe7fb4a56afd9bb1799970f83201eb9` and exact second parent `cbfe8e33f96e105a6d15b6db6035b25e927b4984`. |
| Exact scope and turnover | PASS — target-to-candidate changes exactly seven paths: add this brief; retire only the predecessor live brief, verification, and audit; modify only the provider contract test, fuzz driver, and benchmark driver. Base-to-candidate changes exactly 66 paths. All retired records remain reachable in Git history; this record and the reserved audit record were absent while candidate tests ran. |
| Preservation boundary | PASS — every other target mode and blob is unchanged. Production code, dependencies, Makefile entry points, workflows, benchmark implementation and comparator, 10% threshold, five-sample minimum, controller, policy, packages, and default-OFF behavior remain byte-identical to target. |
| Provider boundary | PASS — deterministic cases accept valid terminal JSON and reject strict-invalid terminal JSON at exactly `MaxProviderPrompt`, then reject a valid payload at `MaxProviderPrompt+1`. Only random fuzz mutations above the literal 64 KiB fuzz ceiling are skipped; production parser behavior is unchanged. |
| Fuzz driver contract | PASS — all eight targets remain. Strict configuration and provider terminal use literal `10000x`; the other six retain `5s`. Offline archived source, disposable roots, serialized fuzzing, direct failure propagation, and the two-minute hard limit remain enforced. |
| Initial infrastructure stop | PRESERVED — the first focused attempt passed the deterministic boundary and repetitions 1–4, then repetition 5 stopped during compilation with `no space left on device`. No benchmark or later gate ran, the disposable root was removed, and the failure was not swallowed or retried in place. |
| Authorized environmental remediation | PASS — after Product Owner authorization, only exact ignored reproducible caches in the implementation worktree were removed. Pinned toolchains, offline module/download inputs, source, history, and user-owned state were preserved. The complete fixed plan then restarted once in a fresh disposable root with a shared build cache. |
| Focused provider verification | PASS — the exact archived candidate passed `TestParseTerminalEnforcesProductionSizeBoundary` in 0.673s, then completed 20 of 20 consecutive `FuzzParseTerminal` invocations at literal `10000x`. Every invocation executed at least 10,000 cases (200,127 total reported executions); no timeout or target reduction occurred. |
| Fixed-work paired benchmark | PASS — the gate ran exactly once on darwin/arm64 with Go 1.26.7, exact base `84bd69f90d366356b0ce1e1a392f906258f3de91`, five alternating pairs, ParseStatus `250x`, Snapshot `10x`, unchanged 10% threshold, and minimum five samples. ParseStatus moved from median 1,151,331 ns/op to 1,099,460 ns/op (`-4.51%`); Snapshot moved from 135,098,579 ns/op to 139,997,221 ns/op (`+3.63%`). Both passed without regression acceptance. |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | PASS — exact-base control (`state=building`, `changed=66`), module integrity, import closure and boundaries, offline bootstrap, formatting, shell syntax, vet, compilation, unit tests, race checks, all eight fixed fuzz targets, reproducible builds, and distribution compatibility passed. Provider terminal completed `10000x` in 0.706s inside the fixed fuzz inventory. |
| `make v1-candidate-check GO_VERSION=1.26.7` | PASS — Darwin arm64/amd64 binaries and both host archives reproduced; native arm64 Codex and Claude lifecycle, CLI, MCP, upgrade, rollback, removal, and disposable-root conformance passed. Packages remain unsigned and release-blocked. |
| Remote and hosted boundary | PRESERVED — `origin/codex/l7-v1-orchestration` remains `902a66010fe7fb4a56afd9bb1799970f83201eb9`. No push, PR mutation or review, hosted run or rerun, benchmark acceptance, independent audit, owner `GO`, protected merge, signing, release, publication, real host installation, or deployment occurred. Failed Harness run `33488797941` and Trusted policy run `33489813624` remain unresolved hosted evidence; local results do not replace their authority or review envelopes. |
| User-owned state | PASS — the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged with SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| `internal/l7/adapter/provider/contract_test.go` | `4362b1c79d36d98beb1d742ef2382b9ad973b3003eee4fccc83cb655312c1efe` |
| `scripts/harness/check-l7-fuzz.sh` | `38acc8c391213d2c621b0624af3f079c2ab77df1b232a6725cb7ac43ef734bf9` |
| `scripts/harness/check-cli-benchmarks.sh` | `812784f3a52abc96ccd994911619ea27de018fe3f24c5b20a3722e01a3b3363b` |
| Unchanged `internal/harness/benchgate/main.go` | `849ae7b93a4f926c0b1e61f45038c10ae4335f4298013286c18e778d333a8596` |
| Unchanged `internal/harness/buildcontrol/policy.go` | `fe094db22289213dfa9c577d713895b6a1b24348ee22853d50e828dbd24629ac` |
| Reproducible harness test binary | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Reproducible native CLI | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| `level7-dev-loop-1.0.0-dev-codex.zip` | `ce6ead70d4cbec718c2737bddcd0abe7e5cc984b81549f8188af5008c7b4f1fd` |
| `level7-dev-loop-1.0.0-dev-claude.zip` | `d7ae491f869ee0346a5437500272d26fd90a472c494d141ea129d6addab4df3d` |

## Evidence boundary

The local arm64 run verifies the remediation and local candidate contract. It
does not establish hosted macOS amd64 execution, a GitHub accountable-owner
approval, a GitHub review, or trusted-policy readiness. Any later benchmark
regression acceptance remains a separate exact-head Product Owner GitHub
decision and is neither implementation authority nor a policy bypass.

This is implementer-run verification, not an independent audit, owner `GO`,
merge authorization, or release authority.

## Next executable transition

Stop for separate Product Owner authority to commission one independent
read-only `l7-release` audit of the exact verification-record successor commit
and tree by `apbusinessidentity-tech`, distinct from Product Owner Anup Pandey,
implementer `codex-root`, and PR author `anup19950725`. Push, PR mutation or
review, hosted execution or rerun, benchmark acceptance, owner `GO`,
protected-branch merge, signing, release, publication, real host installation,
and deployment remain unauthorized.
