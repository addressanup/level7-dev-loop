# Provider Compatibility Remediation Failure Disposition — Verification

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition` |
| Candidate commit | `83846ecea259e3f966060cc8a7d94c98de78638d` |
| Candidate tree | `ba549d80d842afbe32c7e38999d7dda7e6d27688` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-28T05:09:13Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the repository-local external approval binds brief addition commit `9e942b8e486a55c50f137793be0aec262eb713da`; the controller selected the exact Tier 3 change, base `17664b48c0284982b74f9fad71e011ac32cddaf9`, and candidate/tree above in `building` state |
| Clean-baseline and history | PASS — base and rollback checkpoints resolve to tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`; the failed brief/candidate, five reverts, invalid brief, and its recovery revert remain ancestors |
| Targeted fake-runtime adapter tests | PASS — Codex `codex-cli 0.150.1`, Claude `2.1.247`, and Claude `2.1.247 (Claude Code)` each made exactly one fake `--version` probe, degraded, and made zero semantic invocations |
| `make verify` | PASS — policy, offline module checks, import/effect boundaries, formatting, shell syntax, vet, type compilation, build-tagged actual-host compile-only coverage, complete tests, harness reproducibility, and CLI reproducibility passed |
| `go test -race ./internal/l7/... ./cmd/l7` | PASS — pinned Go 1.26.7 with CGO-enabled race instrumentation on Darwin arm64 |
| `make cli-cross-build` | PASS — declared Darwin arm64 and amd64 binaries built with the pinned toolchain |
| Scope and diff hygiene | PASS — the candidate changes only the approved brief and two adapter test files; `git diff --check` passed; production source, configuration, dependencies, workflows, remotes, and global provider configuration are unchanged |
| Tracked and user-owned state | PASS — tracked worktree and index remained clean after verification; the unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Verification boundary

No actual Codex or Claude executable, version/help surface, prompt, stdin, model
session, network, retry, fallback, provider installation, global configuration,
external CI, remote, merge, release, deployment, or publication participated.
Build-tagged actual-host tests compiled with no tests selected and did not run.
Historical Gate 1 remains bound only to failed candidate `8fba20512d1b5710104ec4b031ae9ee0f41d16a5`;
the overall provider qualification remains `NO_GO` and gates 3 through 6 remain
`NOT_RUN`.

This is implementer-run technical verification, not an independent audit or
merge/release authority. The next transition is one separately authorized,
independent, read-only `l7-release` audit of the verification successor. Any
implementation change invalidates this record and requires fresh verification.
