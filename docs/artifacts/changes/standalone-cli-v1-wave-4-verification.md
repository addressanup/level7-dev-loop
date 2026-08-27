# Standalone CLI v1 Wave 4 — Verification

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-4` |
| Candidate commit | `c7d30046f95cc276f14864d502eabcabe917af97` |
| Candidate tree | `c65ddb6a756b790124a20059cabe0ef66c509439` |
| Result | `PASS` |
| Reviewer | `codex-root` |

## Checks

| Check | Result |
|---|---|
| `make verify` | PASS — the candidate-bound Tier 3 controller truthfully remained at `building`; offline dependency, import/effect boundary, formatting, shell, vet, typecheck, tagged actual-host compile-only, full test, and both reproducibility gates passed |
| Fake-provider end to end | PASS — Codex→Claude and Claude→Codex completed `run → verify → review → ready → merge` in disposable no-remote repositories; full-SHA prompting, exact local ref advancement, restart reconstruction, and bounded records were asserted |
| Readiness adversarial matrix | PASS — dirty/stale evidence, config drift, forged or self authority, `NO_GO`, missing audit/approval, duplicate/failing checks, and benchmark waiver/absence all failed closed; headless evaluation left an outside-Git directory unchanged |
| Merge and recovery matrix | PASS — unsafe, missing, occupied, divergent, and concurrently changed targets failed closed; compare-and-swap advanced only the explicit local ref; receipt-loss recovery never reset the ref |
| Process/provider hardening | PASS — cancellation, timeout, output flood, a session-escaped pipe holder, executable drift, reviewer mutation, and scope drift remained bounded; Claude reviewer argv exposed only `Read,Glob,Grep` under safe plan mode |
| 10,000-path scale | PASS — the exact-bound parser test and shared parser/snapshot benchmarks completed within configured path/output limits |
| `go test -race ./internal/l7/... ./cmd/l7` | PASS — pinned Go 1.26.7, macOS arm64, CGO-enabled race instrumentation |
| Parser fuzzing | PASS — five seconds each for provider-neutral, Codex, Claude, and trusted-CI decoders; 544,527 + 936,503 + 726,695 + 598,808 executions with no crash or contract violation |
| `actionlint` 1.7.12 | PASS — both workflow files passed local syntax, expression, and shell checks |
| `make cli-cross-build` | PASS — declared macOS arm64 and amd64 binaries built with pinned Go 1.26.7 |
| Scope and artifact hygiene | PASS — the implementation candidate has 48 changed paths, all inside the approved Wave 4 set; its successor adds only this verification record, no audit exists, `.l7/config.json` and module metadata are unchanged, and `git diff --check` passed |

## Paired benchmark gate

Five alternating base/candidate samples ran on the same Darwin arm64 host with
Go 1.26.7 and `-benchtime=250ms`. The protected gate compared medians against
base `be898482ce632a8faf60f97022f740cd24585063` and enforced the approved 10%
ceiling; no owner override was used.

| Benchmark | Base samples (ns/op) | Candidate samples (ns/op) | Base median | Candidate median | Change |
|---|---|---|---:|---:|---:|
| `BenchmarkParseStatus10000Paths` | 988327, 1039431, 1022151, 1068390, 1050206 | 1036079, 1021124, 991350, 988812, 981973 | 1,039,431 | 991,350 | -4.63% |
| `BenchmarkSnapshot10000Paths` | 109474903, 121016167, 146381416, 109805320, 119889208 | 122716361, 108126292, 132014042, 133231438, 110227333 | 119,889,208 | 122,716,361 | +2.36% |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| Reproducible CLI / macOS arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| Cross-built macOS amd64 CLI | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Validation boundary

The build-tagged provider probes and provider-order actual-host harness compiled
with `-run '^$'`; no real Codex or Claude process was launched. GitHub Actions,
live repository rules, macOS Intel runtime, signing, notarization, release,
publication, deployment, and merge of this Wave 4 branch are `NOT_RUN`. The
repository has no remote. Local ref effects occurred only inside disposable test
repositories. The accountable-owner benchmark exception was not exercised.

This is implementer-run technical verification, not independent audit, owner
approval, release approval, or merge authority. The next transition is one
separately authorized independent read-only audit of the verified successor.
