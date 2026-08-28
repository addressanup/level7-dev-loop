# Provider Compatibility No-Model Gate Harness — Verification

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness` |
| Candidate commit | `53c04388e96407fd361f974a76e9c5fcda29e0d3` |
| Candidate tree | `6167a55f4c8889a90dbd1b766a924e9d8cbabb0b` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-28T09:05:44Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the fresh repository-local external approval binds exact brief commit `8ab7412a63e34838c49027050fcf6e6f68b6e65c`; the controller selected this Tier 3 change, base `481adaaec967ac34b6b27cf78509b85d5c068abc`, and the candidate/tree above in `building` state with five changed paths |
| Base and preserved history | PASS — base `481adaaec967ac34b6b27cf78509b85d5c068abc` resolves to tree `d57a334696487b1d15557c9980e8a55c2dc4c930`; original base `51191ad6edc670a0e73c3d152484bd57785144f7`, clean-baseline disposition head `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a`, brief `438375b2d8edcec0983f9ce4eb4654a222cabd68`, failed candidate `8fba20512d1b5710104ec4b031ae9ee0f41d16a5` with tree `7943f38db45705ce9cc1da01fb600f57e518215f`, rollback-closure audit `5ce7972b6cbb2a257a0665efffa01fabddf15cd6`, inconsistent brief `9bcb10a63aebfe79bbddd347e16fb2d6f52ee240`, and its ordinary revert remain immutable ancestors |
| Exact candidate scope | PASS — the candidate adds only the approved brief and four declared test files: one untagged fake-runner file and one build-tagged actual-host file for each provider; no existing file changed, `git diff --check` passed, and the brief blob is byte-identical to its approved commit |
| Production compatibility and controls | PASS — existing production blobs are unchanged from the base; Codex remains exactly `codex-cli 0.149.1`, Claude remains exactly `2.1.241`, and `codex-cli 0.150.1`, `2.1.247`, and `2.1.247 (Claude Code)` still degrade after one fake version probe with zero semantic invocations. No argument, permission, output-schema, cancellation, cleanup, reviewer-immutability, scope, containment, or default-OFF control changed |
| Codex fake-runner qualification | PASS — tests derive fresh implementer and reviewer observations from the production argv, require one final stdin sentinel, replace only that sentinel for positive help, insert one test-owned unknown option immediately before help, perform one fake version request plus five bounded no-input observations, and fail closed on unsafe argv and unsuccessful or ambiguous outcomes |
| Claude fake-runner qualification | PASS — tests cover both exact target spellings, derive three fresh observations for each role from an argv containing exactly one `--max-turns 64` pair, change only `64` to `not-an-integer` for the typed negative control, perform one fake version request plus six bounded no-input observations, and require positive help success plus nonzero exits for both negative controls |
| Fail-closed diagnostics and postconditions | PASS — fake-runner tests reject wrong versions, failed help, successful unknown-option controls, successful invalid-value controls, empty or invalid UTF-8 output, overflow, timeout, runner errors, ambiguous negative exits, malformed or contaminated argv, missing/duplicate controls, and postcondition drift. The checked runners execute postconditions after earlier gate errors; actual-host tests also register source and executable cleanup checks before the first interface invocation |
| Help advertisement | PASS — bounded help advertisement is diagnostic only. Tests prove that missing advertisement does not reject a successful positive help result and that present advertisement cannot override either a successful or rejecting negative parser-control exit |
| Tagged actual-host compile boundary | PASS — both build-tagged harnesses compiled with `-tags l7_actual_provider -run '^$'`; no tagged test was selected, both future actual-host gates remain `NOT_RUN`, and no provider executable or interface was invoked |
| Targeted provider tests | PASS — repository-pinned, network-disabled tests matching `Qualification` passed for both Codex and Claude adapter packages using injected fake runners only |
| `make verify` | PASS — policy, offline module checks, import/effect boundaries, formatting, shell syntax, vet, type compilation, build-tagged compile-only coverage, complete tests, harness reproducibility, and CLI reproducibility passed |
| `go test -race ./internal/l7/... ./cmd/l7` | PASS — repository-pinned Go 1.26.7 with CGO-enabled race instrumentation on Darwin arm64 |
| `make cli-cross-build` | PASS — declared Darwin arm64 and amd64 binaries built with the pinned, offline toolchain |
| Artifact budget and rollback | PASS — the candidate has approved brief commit `8ab7412a63e34838c49027050fcf6e6f68b6e65c` as its parent, and that brief has exact base `481adaaec967ac34b6b27cf78509b85d5c068abc` as its parent. Before audit, ordinary reverts of this verification record, implementation, and brief restore tree `d57a334696487b1d15557c9980e8a55c2dc4c930`; after audit, its record is reverted first. The brief, this sole verification record, and one future audit remain within the Tier 3 artifact budget |
| Tracked and user-owned state | PASS — tracked worktree and index were clean after candidate verification; the unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Historical evidence and gate state

Historical Codex actual-host Gate 1 remains a pass bound only to failed
candidate `8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Historical Claude actual-host
Gate 2 remains `NO_GO`: both exact role help invocations succeeded, both
unknown-option parser controls unexpectedly exited successfully, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement remains non-dispositive;
the two successful unknown-option controls remain dispositive.

The new Codex no-model parser gate and Claude no-model parser gate are both
`NOT_RUN`. Historical evidence does not transfer to this harness, and neither a
historical nor future diagnostic result can change compatibility or support.

## Provider and verification boundary

No actual Codex or Claude executable, version/help surface, prompt, stdin, model
session, network, retry, fallback, provider installation, global configuration,
external CI, remote, merge, release, deployment, or publication participated.
All observed provider requests and responses in this verification came from
injected fake runners; build-tagged actual-host tests only compiled with zero
tests selected.

This is implementer-run technical verification, not an independent audit or
merge/release authority. The next transition is fresh explicit authorization
for one independent, read-only `l7-release` audit bound to the verification
successor commit/tree. Any implementation change invalidates this record and
requires fresh verification.
