# Level 7 v1.0 Hosted Provider Fuzz and Benchmark Remediation - Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-provider-fuzz-benchmark-remediation` |
| Candidate commit | `1756aaa6efa60a8c43128ed86c4dc4643f10b2b2` |
| Candidate tree | `1dc6c325c00bbfef9d2cc3fa12e85ceffb6a94f1` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |

## Decision and independence

`GO` for only the exact local remediation verification commit and tree above.
The repair bounds only random provider-parser fuzz mutations, restores a fixed
execution budget for that affected target, adds deterministic production-limit
coverage, and makes the unchanged paired benchmarks run fixed symmetric work.
It does not accept the historical failures or weaken production, timeout,
benchmark, controller, policy, hosted, or authority boundaries.

Product Owner Anup Pandey (`addressanup`) separately commissioned this Tier 3
team-assurance audit after the forge reviewer login was resolved as exactly
`apbusinessidentity-tech`. That reviewer is distinct from owner `addressanup`,
implementer and verifier `codex-root`, and PR author `anup19950725`. The brief
and implementer verification were treated as claims to test, not conclusions.

The audit was independently reasoned and candidate-read-only until this sole
record was written. It is not a GitHub PR review, hosted result, owner `GO`,
merge-readiness decision, benchmark-regression acceptance, or authority for a
push, PR mutation, hosted run, protected merge, signing, release, publication,
installation, or deployment.

## Requirement and evidence map

| Area | Independent assessment |
|---|---|
| Exact candidate | PASS - the audit began at a clean `codex/l7-v1-orchestration` HEAD `1756aaa6efa60a8c43128ed86c4dc4643f10b2b2`, tree `1dc6c325c00bbfef9d2cc3fa12e85ceffb6a94f1`. Its sole parent is integration `0277b90cf01cb47120b30ac4f1f0bb8bdd00e864`, and it adds only the verification record. The audit path was absent. |
| Proposal and topology | PASS - proposal `cbfe8e33f96e105a6d15b6db6035b25e927b4984`, tree `316a2f0f20534e667e4ee1f8ae336d1916ee1d1f`, has sole parent/base `84bd69f90d366356b0ce1e1a392f906258f3de91`, tree `84c2a105227f98089ea001f97473af79933bd743`, and adds only the brief. Integration `0277b90cf01cb47120b30ac4f1f0bb8bdd00e864`, tree `0721dc7c17fdfc35103b3f3a91f6a8f095306662`, has exact parents target `902a66010fe7fb4a56afd9bb1799970f83201eb9` then proposal `cbfe8e33f96e105a6d15b6db6035b25e927b4984`. |
| Exact scope and history | PASS - target-to-integration changes exactly seven paths: add the successor brief, retire only the three predecessor live records, and modify only the provider test, fuzz driver, and benchmark driver. Base-to-integration changes 66 paths; base-to-verification changes 67. The three retired records remain readable at target `902a66010fe7fb4a56afd9bb1799970f83201eb9` and reachable in immutable history. `git diff --check` passed. |
| Provider boundary | PASS - `internal/l7/adapter/provider/contract_test.go` has SHA-256 `4362b1c79d36d98beb1d742ef2382b9ad973b3003eee4fccc83cb655312c1efe`. Only random mutations above literal `64 << 10` skip inside the fuzz callback. Deterministic tests accept valid and reject duplicate-key strict-invalid JSON at exact production `MaxProviderPrompt` (1 MiB), then reject the valid payload at `MaxProviderPrompt+1`. Production `contract.go` is byte-identical to target. |
| Fuzz fail-closed contract | PASS - `scripts/harness/check-l7-fuzz.sh` has SHA-256 `38acc8c391213d2c621b0624af3f079c2ab77df1b232a6725cb7ac43ef734bf9`. All eight targets remain inventoried; strict configuration and provider terminal use literal `10000x`; the other six retain `5s`. Clean archived source, pinned offline toolchain settings, disposable cache/corpus/temp/telemetry roots, `CGO_ENABLED=1`, serialization, direct exit propagation, and literal two-minute hard timeouts remain. No retry, swallowed timeout, target reduction, environment-selected budget, blanket timeout increase, or architecture bypass exists. Shell syntax passed. |
| Independent focused diagnostics | PASS - from an exact candidate Git archive in disposable `/private/tmp` roots, pinned offline Go 1.26.7 passed `TestParseTerminalEnforcesProductionSizeBoundary` in 0.653s. A separate invocation of `FuzzParseTerminal` with `CGO_ENABLED=1`, `-parallel=1`, `-fuzztime=10000x`, and `-timeout=2m` completed all 10,000 executions and passed in 0.719s. No repository path or persistent cache was written. |
| Benchmark fail-closed contract | PASS by implementation and evidence inspection - `scripts/harness/check-cli-benchmarks.sh` has SHA-256 `812784f3a52abc96ccd994911619ea27de018fe3f24c5b20a3722e01a3b3363b`. It runs exactly five alternating base/candidate pairs and both unchanged benchmark names with literal `250x` ParseStatus and `10x` Snapshot work. Offline disposable roots and direct failure propagation remain. The unchanged comparator has SHA-256 `849ae7b93a4f926c0b1e61f45038c10ae4335f4298013286c18e778d333a8596` and is still called with literal threshold 10 and minimum samples 5; no candidate-controlled acceptance input was added. The paired benchmark was not rerun during this audit. |
| Verification evidence | PASS within its local evidence boundary - the record binds `PASS` to integration `0277b90cf01cb47120b30ac4f1f0bb8bdd00e864`, tree `0721dc7c17fdfc35103b3f3a91f6a8f095306662`. It records a terminal first infrastructure failure rather than an in-place retry, an authorized one-time clean restart, 20 consecutive affected-target passes totaling 200,127 reported executions, one fixed-work paired benchmark pass, team `make verify`, and `v1-candidate-check`. Its reported benchmark medians independently recompute to ParseStatus `-4.51%` and Snapshot `+3.63%`. Heavyweight unchanged verification was not rerun. |
| Controller and policy | PASS - target and candidate retain identical blobs for Harness workflow `228cad9e42b9c23a0d79a6c582aa918c0569748c`, trusted policy workflow `e8afb29b310d5ab6270f667b45efcd4d22a4e604`, comparator `1960a749da40f3655d268080c044bb3a38b031b7`, controller policy `3cde1e7b4bdd1ee80e4a6015f3e4c45e79da6c75`, Makefile, `go.mod`, and `go.sum`. The future exact-head benchmark marker can recognize an external owner decision only after a failed benchmark; it neither changes candidate bytes nor supplies the separate reviewer/readiness boundary. |
| Historical hosted evidence | PRESERVED - Harness run `33488797941` is bound to target `902a66010fe7fb4a56afd9bb1799970f83201eb9`: baseline, shadow, and arm64 passed; amd64 failed terminally when provider `FuzzParseTerminal` stalled at 17,868 executions and returned `context deadline exceeded`; the paired gate failed Snapshot at `+11.18%` while ParseStatus passed at `-8.39%`. Trusted policy run `33489813624` separately failed `AUTH-001` and `STATE-002`. Neither run was rerun or represented as remediated-head evidence. |
| Hosted and authority boundary | PRESERVED - PR #15 remains open at remote head `902a66010fe7fb4a56afd9bb1799970f83201eb9`, base `84bd69f90d366356b0ce1e1a392f906258f3de91`, author `anup19950725`, sole trusted label `l7-risk-tier-3`, and no reviews. Forge identities resolve as distinct users. The available authenticated CLI session is `addressanup`, so no GitHub review or remote mutation was attempted or claimed for reviewer `apbusinessidentity-tech`. |
| User-owned state | PASS - the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged with SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Findings

No unresolved BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW finding remains within
the approved local remediation scope and its evidence boundary. The candidate
does not retry until green, swallow a timeout, reduce the target inventory,
bypass production architecture, weaken a blanket timeout or benchmark
threshold, or make benchmark acceptance candidate-controlled.

## Residual risks and claim boundary

- Exact remediated-head hosted execution remains `NOT_RUN`. In particular,
  macOS amd64 behavior and a fresh paired benchmark remain unqualified until a
  separately authorized exact-head push and hosted run complete.
- The 64 KiB fuzz-only ceiling intentionally removes random mutation coverage
  between 64 KiB and the 1 MiB production limit. Exact-limit valid,
  strict-invalid, and oversized cases cover the admission boundary
  deterministically, not the full upper interval randomly.
- Fixed `10000x` work removes the affected wall-clock cancellation race but
  does not guarantee duration. The unchanged two-minute hard timeout correctly
  remains terminal if an execution or corpus stalls.
- The independent audit did not rerun the paired benchmark and therefore
  relies on the verifier's recorded medians while independently confirming the
  arithmetic, driver, benchmark definitions, comparator, and fail-closed gate.
  A future hosted exact-head comparison remains required.
- A possible future exact-head benchmark-regression acceptance is a separate
  accountable-owner GitHub decision. It is not implementation authority, an
  independent review, a green benchmark result, or a trusted-policy bypass.
- No real provider/model session, host installation, signing, publication,
  release, or deployment was exercised or authorized by this audit.

## Rollback and preservation

No live or remote effect requires rollback. If this audit-only commit is later
abandoned, ordinarily revert only that commit. A deeper authorized rollback
must proceed in reverse order: verification
`1756aaa6efa60a8c43128ed86c4dc4643f10b2b2`, then integration merge
`0277b90cf01cb47120b30ac4f1f0bb8bdd00e864` using target
`902a66010fe7fb4a56afd9bb1799970f83201eb9` as mainline. Confirm restoration
of exact target tree `88693537e3b1e2fe3885f295c7e510472ab5fbd4` and stop on any conflict or
unexpected path. Preserve the base-rooted proposal, retired records and
commits, external envelopes, failed hosted runs, PR #15, packages,
credentials, user installations, releases, and production state.

## Next executable transition

Return this exact audit commit to `codex-root` for sole-record and decision
validation. Under the Product Owner's bounded commission, `codex-root` may
then record the matching local `.git/l7/audits` envelope for reviewer
`apbusinessidentity-tech` and the exact candidate above, evaluate the resulting
local team-assurance state, and stop. This audit does not authorize owner
`GO`, push, PR mutation or review, hosted execution or rerun, benchmark
acceptance, protected-branch merge, signing, release, publication,
installation, or deployment.
