# Provider Compatibility Rollback Closure — Verification

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure` |
| Candidate commit | `4828519b7b021467826a0a906dc0551501bd610f` |
| Candidate tree | `1852959d7a69c889a9f062a61d02b620c2d35cea` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-28T07:05:50Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the fresh repository-local external approval binds brief addition commit `afed168a9f7313e8f7dad9fa8aa97b7a64155588`; the controller selected the exact Tier 3 change, base `7e72988a189f51121931ea55a57209668ff1e37c`, and candidate/tree above in `building` state |
| Clean disposition and history | PASS — original base `51191ad6edc670a0e73c3d152484bd57785144f7`, clean-baseline head `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a`, recovery base `17664b48c0284982b74f9fad71e011ac32cddaf9`, and disposition head `7e72988a189f51121931ea55a57209668ff1e37c` all resolve to tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`; prior brief, implementation, verification, audit, and disposition commits remain ancestors |
| Exact restored test state | PASS — both adapter test files are byte-identical to prior test commit `83846ecea259e3f966060cc8a7d94c98de78638d`; no production file changed |
| Targeted fake-runtime adapter tests | PASS — Codex `codex-cli 0.150.1`, Claude `2.1.247`, and Claude `2.1.247 (Claude Code)` each made exactly one fake `--version` probe, degraded, and made zero semantic invocations |
| `make verify` | PASS — policy, offline module checks, import/effect boundaries, formatting, shell syntax, vet, type compilation, build-tagged actual-host compile-only coverage, complete tests, harness reproducibility, and CLI reproducibility passed |
| `go test -race ./internal/l7/... ./cmd/l7` | PASS — repository-pinned Go 1.26.7 with CGO-enabled race instrumentation on Darwin arm64 |
| `make cli-cross-build` | PASS — declared Darwin arm64 and amd64 binaries built with the pinned toolchain |
| Rollback contract | PASS — candidate `4828519b7b021467826a0a906dc0551501bd610f` has brief commit `afed168a9f7313e8f7dad9fa8aa97b7a64155588` as its parent, and that brief has exact disposition base `7e72988a189f51121931ea55a57209668ff1e37c` as its parent; reverting tests then brief therefore restores tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`, while the declared later-state sequences first remove every audit and verification successor |
| Scope and diff hygiene | PASS — the candidate changes only the approved brief and two adapter test files; `git diff --check` passed; the brief bytes are unchanged from the approval binding; production source, configuration, dependencies, workflows, remotes, and global provider configuration are unchanged |
| Tracked and user-owned state | PASS — tracked worktree and index remained clean after verification; the unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Provider and verification boundary

No actual Codex or Claude executable, version/help surface, prompt, stdin, model
session, network, retry, fallback, provider installation, global configuration,
external CI, remote, merge, release, deployment, or publication participated.
Build-tagged actual-host tests compiled with no tests selected and did not run.

Historical Codex actual-host Gate 1 remains a pass bound only to failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Claude Gate 2 remains `NO_GO`
because both implementer and reviewer unknown-option controls unexpectedly exited
successfully. Both exact role help invocations succeeded, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`; help advertisement remains non-dispositive.
Gates 3 through 6 remain `NOT_RUN`.

This is implementer-run technical verification, not an independent audit or
merge/release authority. The next transition is one separately authorized,
independent, read-only `l7-release` audit of the verification successor. Any
implementation change invalidates this record and requires fresh verification.
