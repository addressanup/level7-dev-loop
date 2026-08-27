# Standalone CLI v1 Wave 3 — Verification

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-3` |
| Candidate commit | `e2c0cdba4f0f5d02377695e179c5596148b6c423` |
| Candidate tree | `6c4269e06f4329e05e906421a58d729aad5225a0` |
| Result | `PASS` |
| Reviewer | `local-verifier` |

## Checks

| Check | Result |
|---|---|
| `make verify` | PASS — the candidate-bound Tier 3 policy check remained at `building`; offline module, import/effect boundary, format, shell, vet, typecheck, full test, and reproducibility gates passed |
| Fake-provider CLI integration | PASS — Codex→Claude and Claude→Codex completed `run → verify → review`; exact commits, restart reconstruction, default-OFF gating, and the three-artifact ceiling were asserted without network or credentials |
| Adversarial execution tests | PASS — cancellation, timeout, output flood, inherited child process, hook failure, dirty index, stale HEAD/tree, scope expansion, provider/reviewer mutation, self-review, local-evidence drift, and executable replacement fail closed |
| `go test -race ./internal/l7/... ./cmd/l7` | PASS — exact candidate, macOS arm64, CGO-enabled race instrumentation |
| Parser fuzzing | PASS — five seconds each for the neutral, Codex, and Claude terminal parsers; 702,774 + 1,228,383 + 1,026,642 executions with no crash or contract violation |
| `make cli-cross-build` | PASS — macOS arm64 and amd64 binaries built with pinned Go 1.26.7 |
| Actual-host test compilation | PASS — both build-tagged probes compiled with `-run '^$'`; no provider executable was launched |
| `git diff --check 9021c69d6c11d71c47510983b157ebc54db6cbf0..e2c0cdba4f0f5d02377695e179c5596148b6c423` | PASS — 41 implementation paths, no path outside the approved Wave 3 set; the user-owned untracked foundation audit was excluded |

## Benchmarks

Same-host Apple M3 Max medians used Go 1.26.7, `-benchtime=250ms`, and five
samples for shared 10,000-path benchmarks. Against base `9021c69d`:

| Benchmark | Base median | Candidate median | Change |
|---|---:|---:|---:|
| `BenchmarkParseStatus10000Paths` | 963,603 ns/op | 982,331 ns/op | +1.94% |
| `BenchmarkSnapshot10000Paths` | 102,257,986 ns/op | 100,877,306 ns/op | -1.35% |

Both remain inside the approved 10% regression threshold. Three-sample Wave 3
medians were 2,748,479 ns/op for supervised process dispatch, 8,399 ns/op for
Codex event parsing, 10,664 ns/op for Claude result parsing, 3,971 ns/op for the
neutral terminal protocol, 1,932 ns/op for verification dispatch, and 1,482
ns/op for in-memory lifecycle reconstruction.

## Reproducible identities and validation boundary

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| Reproducible CLI / macOS arm64 | `5c06c4ca96bc3170a921444e63bd5cef7e9c3226fd165b291ad5045f82d88968` |
| Cross-built macOS amd64 CLI | `aa5179a9ce0b350b17df015f3608d361c9c3026278a77f4ddd2ec442bf1a807f` |

Real Codex and Claude launches, credentials, network use, disposable actual-host
trials, and daemon/session-escape experiments are `NOT_RUN` because their
separate authorization was not granted. Adapter compatibility remains
provisional and bound only to fake fixtures for Codex `codex-cli 0.149.1` and
Claude Code `2.1.241`; no provider-support claim or overall `GO` follows from
this record. Readiness, merge, and deployment remain unavailable. This record is
technical verification evidence, not owner approval or independent-audit
authority.
