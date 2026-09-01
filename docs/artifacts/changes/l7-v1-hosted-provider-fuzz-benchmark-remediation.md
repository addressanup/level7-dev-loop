# Level 7 v1.0 Hosted Provider Fuzz and Benchmark Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-provider-fuzz-benchmark-remediation` |
| Risk tier | `3` |
| Status | `proposed`; implementation requires fresh Product Owner approval of this exact brief commit |
| Base commit | `84bd69f90d366356b0ce1e1a392f906258f3de91` |
| Base tree | `84c2a105227f98089ea001f97473af79933bd743` |
| Integration target | `codex/l7-v1-orchestration` at commit `902a66010fe7fb4a56afd9bb1799970f83201eb9`, tree `88693537e3b1e2fe3885f295c7e510472ab5fbd4` |
| Proposal branch | `codex/l7-v1-hosted-provider-fuzz-benchmark-remediation`, rooted directly at the exact base |
| Remote context | PR #15, exact base and target above, sole trusted label `l7-risk-tier-3`; no remote mutation is authorized |
| Failed hosted evidence | Harness run `33488797941`: macOS amd64 job `99795090967`, benchmark job `99795091106`; Trusted policy run `33489813624`, evaluate check `99798417548` |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root`; hosted PR author remains `anup19950725` |
| Independent auditor | `apbusinessidentity-tech` |
| Assurance | Tier 3 team: fresh external owner approval, exact-candidate verification, separately commissioned independent read-only audit, then a separate owner decision |
| Next executable transition | Stop for explicit Product Owner approval bound to this exact proposal commit |

## Problem

Harness run `33488797941` on exact target `902a660…` passed Go 1.26.7,
Go 1.27, and macOS arm64. The macOS amd64 job failed closed in unchanged
`FuzzParseTerminal`: the prior strict-configuration repair completed its
`10000x` budget in 0.828 seconds, then the provider parser reached 17,868
executions and returned `context deadline exceeded` under its five-second
wall-clock budget. The parser admits production payloads up to 1 MiB and its
random fuzz callback has no smaller per-iteration bound.

The paired benchmark job also failed closed. `BenchmarkParseStatus10000Paths`
passed at -8.39%, while `BenchmarkSnapshot10000Paths` was +11.18% against the
unchanged 10% threshold. The benchmark implementation, tests, driver, and gate
are byte-identical at the exact base and target, and the five Snapshot samples
range widely on both sides. The result remains a real hosted blocker, but does
not justify weakening the threshold or inventing candidate-controlled
acceptance. The smallest repair makes the workload fixed-count and symmetric;
it does not accept the observed regression.

Trusted policy run `33489813624` separately failed `AUTH-001` because no
exact-head accountable-owner GitHub approval exists and `STATE-002` because the
candidate is not ready. PR #15 has no reviews. Local approval, verification,
or audit envelopes cannot satisfy those later hosted authority boundaries.

The approved predecessor brief `f5691891874ca9b88e6090ec2854f45ed7d49e3f`,
two-parent integration `830cec9037e1852deee0a1ca5ebd391dbaa06c1f`,
verification `ef34d34c419a048c57bfdbe61c752eef52bd45bd`, and independent
audit/current target `902a66010fe7fb4a56afd9bb1799970f83201eb9` remain
historical evidence only. None authorizes changed bytes.

## Scope

This base-rooted proposal commit adds only this brief. After fresh approval,
one additive two-parent integration may combine this proposal with only the
exact target above, retire only the target's current live brief, verification,
and audit from the integration tree, and preserve every retired record and
commit in Git history.

Relative to the target, implementation may modify only three files:

- the provider contract test, to cap random fuzz mutations at a literal 64 KiB
  while adding deterministic valid, strict-invalid, and oversized cases at the
  exact 1 MiB production parser boundary;
- the fuzz driver, to give provider `FuzzParseTerminal` a literal `10000x`
  execution budget while retaining strict configuration at `10000x`, the other
  six targets at `5s`, all eight targets, direct exit propagation, and the
  two-minute hard safety timeout; and
- the benchmark driver, to retain both exact benchmarks and five alternating
  same-host base/candidate pairs while replacing the shared adaptive `250ms`
  work budget with literal fixed workloads of `250x` for ParseStatus and `10x`
  for Snapshot.

Production code, dependencies, Makefile entry points, workflows, benchmark
implementation and fixtures, benchmark comparator, 10% threshold, five-sample
minimum, trusted controller and policy, product behavior, default-OFF effects,
packages, and remote state are outside target-relative implementation scope.
There is no retry, swallowed timeout, target reduction, architecture bypass,
environment-selected budget, blanket timeout increase, threshold weakening, or
candidate-controlled acceptance path.

A later exact-head benchmark failure remains blocking unless Product Owner
`addressanup` makes a separately authorized GitHub `APPROVED` review containing
the exact line `L7-Benchmark-Regression-Accepted: <exact-later-head>`. Such a
decision is not implementation authority, does not turn the benchmark job
green, does not satisfy independent review, and does not bypass the trusted
policy.

## Exact implementation file set

Declared path count: 71 (48 Add, 20 Modify, 3 Delete).

Add:

- `docs/artifacts/changes/l7-v1-hosted-provider-fuzz-benchmark-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-provider-fuzz-benchmark-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-provider-fuzz-benchmark-remediation-audit.md`
- `cmd/l7-embed/main.swift`
- `cmd/l7/mcp_server.go`
- `cmd/l7/mcp_server_test.go`
- `cmd/l7/orchestration_cli.go`
- `cmd/l7pack/main.go`
- `cmd/l7pack/main_test.go`
- `go.sum`
- `internal/harness/v1candidate/main.go`
- `internal/harness/v1candidate/main_test.go`
- `internal/l7/adapter/claude/worker.go`
- `internal/l7/adapter/codexapp/discovery.go`
- `internal/l7/adapter/codexapp/discovery_test.go`
- `internal/l7/adapter/codexapp/worker.go`
- `internal/l7/adapter/codexapp/worker_test.go`
- `internal/l7/adapter/cyber/cyber.go`
- `internal/l7/adapter/cyber/cyber_test.go`
- `internal/l7/adapter/discovery/discovery.go`
- `internal/l7/adapter/discovery/discovery_test.go`
- `internal/l7/adapter/gateway/client.go`
- `internal/l7/adapter/gateway/gateway_test.go`
- `internal/l7/adapter/gateway/worker.go`
- `internal/l7/adapter/headless/headless.go`
- `internal/l7/adapter/headless/headless_test.go`
- `internal/l7/adapter/headlessworker/executor.go`
- `internal/l7/adapter/headlessworker/executor_test.go`
- `internal/l7/adapter/memory/apple.go`
- `internal/l7/adapter/memory/memory.go`
- `internal/l7/adapter/memory/memory_test.go`
- `internal/l7/adapter/orchestrationconfig/config.go`
- `internal/l7/adapter/orchestrationconfig/config_test.go`
- `internal/l7/adapter/state/orchestration.go`
- `internal/l7/adapter/state/orchestration_test.go`
- `internal/l7/adapter/toolbroker/broker.go`
- `internal/l7/adapter/toolbroker/broker_test.go`
- `internal/l7/domain/orchestration.go`
- `internal/l7/domain/routing.go`
- `internal/l7/domain/routing_test.go`
- `scripts/harness/bootstrap-modules.sh`
- `scripts/harness/check-bootstrap-modules.sh`
- `scripts/harness/check-l7-fuzz.sh`
- `scripts/harness/check-v1-conformance.sh`
- `skills/l7-cyber/SKILL.md`
- `skills/l7-headless/SKILL.md`
- `skills/l7-onboard/SKILL.md`
- `skills/l7-sync/SKILL.md`

Modify:

- `.github/workflows/harness.yml`
- `CHANGELOG.md`
- `Makefile`
- `README.md`
- `cmd/l7/main.go`
- `go.mod`
- `harness/import-boundaries.tsv`
- `internal/harness/distribution/main.go`
- `internal/l7/adapter/claude/adapter.go`
- `internal/l7/adapter/claude/adapter_test.go`
- `internal/l7/adapter/codex/adapter.go`
- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/process/process.go`
- `internal/l7/adapter/process/process_test.go`
- `internal/l7/adapter/provider/contract_test.go`
- `internal/l7/app/app.go`
- `internal/l7/domain/result.go`
- `scripts/harness/check-cli-benchmarks.sh`
- `scripts/harness/check-import-boundaries.sh`
- `skills/l7-next/SKILL.md`

Delete from the live integration tree only; preserve in Git history:

- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation-audit.md`

## Acceptance criteria

1. The proposal commit has exact sole parent/base `84bd69f…`, adds only this
   brief, and changes no implementation, evidence, prior record, envelope,
   implementation worktree, PR, or remote state.
2. Fresh external Product Owner approval names this change ID, exact brief
   commit, implementer `codex-root`, and exact target. No predecessor approval,
   verification, audit, review, owner decision, or hosted result transfers.
3. Integration starts only from target `902a660…`, tree `88693537…`, and has
   that target and the approved proposal as its two parents. Target drift stops
   the transition.
4. Target-to-integration changes exactly seven paths: add this brief; delete
   only the three current live governance records; modify only the provider
   test, fuzz driver, and benchmark driver named in Scope. No history rewrite,
   rebase, reset, amend, squash, or force update is permitted.
5. Base-to-integration changes exactly 66 paths: this brief plus 65 declared
   non-governance paths. Final Tier 3 turnover adds only this successor's
   verification and audit records, for 68 base-visible paths and the exact
   71-path declared contract including the three target-live retirements.
6. Random provider-parser mutations above 64 KiB are skipped only in the fuzz
   callback. Deterministic cases exercise valid and strict-invalid inputs at
   exactly `MaxProviderPrompt` and rejection at `MaxProviderPrompt+1`.
   Production parser limits and behavior remain byte-identical.
7. The fuzz driver keeps all eight exact targets, pinned offline toolchains,
   clean archived source, disposable cache/corpus/temp/telemetry roots,
   `CGO_ENABLED=1`, serialized fuzzing, and the two-minute hard timeout. Both
   named targets use `10000x`; the other six retain `5s`; any setup, fuzz,
   boundary, or timeout failure remains terminal.
8. The benchmark driver runs exactly five alternating base/candidate samples
   for both unchanged targets with fixed `250x` ParseStatus and `10x` Snapshot
   workloads. It retains offline disposable roots and invokes the unchanged
   comparator with literal threshold 10 and minimum samples 5.
9. Focused verification completes provider `FuzzParseTerminal` at `10000x` at
   least 20 consecutive times, passes the deterministic 1 MiB boundaries, and
   runs the fixed-work paired benchmark plan once. A failed or blocked result
   is recorded and stops; there is no retry-until-green.
10. The exact integrated head passes
    `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` and
    `make v1-candidate-check GO_VERSION=1.26.7` from disposable test-owned
    roots. Local results are not credited as hosted evidence.
11. The trusted controller, policy, workflow benchmark threshold, comparator,
    and external marker behavior remain byte-identical. A fresh matching local
    owner envelope may authorize implementation but cannot satisfy the absent
    GitHub owner or independent-review envelopes.
12. If verification passes, one verification commit adds only this successor's
    record and binds `PASS` to the exact integration commit/tree. A separately
    authorized independent read-only audit by `apbusinessidentity-tech` then
    adds only this successor's audit record and binds `GO` or `NO_GO` to the
    exact verification successor.
13. A later separately authorized push requires fresh exact-head baseline,
    shadow, macOS arm64/amd64, paired benchmark, and trusted-policy results.
    If the paired benchmark remains failed, only an explicit exact-head GitHub
    decision by `addressanup` using the existing policy marker may accept it;
    the present brief approval cannot.
14. All historical records remain immutable and reachable. Only the three
    named current live records are retired; no deprecated governance chain is
    extended.
15. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and unstaged with SHA-256
    `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

## Risks and mitigations

- **Fuzz stability mistaken for coverage:** retain all eight targets and exact
  production-limit deterministic cases while bounding only random mutations.
- **Benchmark noise mistaken for acceptance:** use fixed symmetric work and
  one fail-closed comparison; keep the 10% threshold and comparator unchanged.
- **Candidate-controlled policy:** add no acceptance input; any later exception
  must be an exact-head accountable-owner GitHub decision consumed by the
  unchanged trusted-base policy.
- **Scope or history drift:** bind exact identities, use one additive merge,
  compare every other target mode/blob, and retire only the named live trio.
- **Stale readiness:** preserve both failed runs and require fresh approval,
  verification, audit, hosted checks, GitHub reviews, and owner decision after
  changed bytes.
- **User-state damage:** use exact pathspecs and recheck the unrelated file's
  status and fingerprint after every transition.

## Rollback

Before integration, ordinarily revert only the proposal commit; the
implementation branch remains at exact target `902a660…`. After integration,
ordinarily revert later audit and verification commits first, then revert the
two-parent integration with the target parent as mainline and confirm exact
restoration of tree `88693537e3b1e2fe3885f295c7e510472ab5fbd4`.

Every rollback stops on conflict or unexpected path. Historical commits and
records, external envelopes, PR #15, failed hosted runs, packages, credentials,
installations, releases, and production state remain untouched.

## Current transition

Commit only this proposal as a direct child of exact base `84bd69f…`, then stop
for explicit Product Owner approval of that exact brief commit. Only after that
approval may `codex-root` record the matching local external envelope and
perform the declared integration and three-file remediation. Verification,
independent audit, benchmark acceptance, owner GO, push, PR mutation or review,
hosted execution or rerun, protected-branch merge, signing, release,
publication, installation, and deployment remain separately gated and
unauthorized.
