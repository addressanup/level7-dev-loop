# Hosted CI Readiness Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `hosted-ci-readiness-remediation` |
| Candidate commit | `228af7bbd29e1836ff8db59b2aab256af0b9fb9f` |
| Candidate tree | `04283aa24cfb010d539afad18c2a8e4ccf439fdf` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-28T10:04:42Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — external active-user approval binds immutable brief commit `55e97c93b65c96f4f281ea090816b3205784cae5`; owner `accountable-owner` and implementer `codex-root` are distinct. The controller selected Tier 3, exact base `481adaaec967ac34b6b27cf78509b85d5c068abc`, candidate/tree above, four changed paths, state `building`, and the verification transition. |
| Lineage and scope | PASS — base tree is `d57a334696487b1d15557c9980e8a55c2dc4c930`; brief tree is `5cef591260b1403e877f6d79fc43f1c8b6003edb`; the brief is the candidate's grandparent and implementation parent. Base-to-candidate adds only the approved brief and modifies only the three approved implementation files. The approved brief blob is unchanged and `git diff --check` passes. |
| Hermetic Git setup | PASS — the formerly failing snapshot test gives only its test-owned `commit-tree` request `user.useConfigOnly=true`, `user.name=Level Seven`, and `user.email=l7@example.invalid`. The focused test and complete Git adapter package pass with global/system Git configuration disabled. No repository, global, or system Git configuration is written, and the general test runner is unchanged. |
| Clean-runner `make ci` | PASS — the exact hosted technical target passed end to end with author/committer environment removed, global/system Git configuration disabled, and `user.useConfigOnly=true`. This directly covers the Ubuntu failure condition without relying on host identity. |
| Intel lock record | PASS by exact data inspection — the candidate adds one baseline `1.26.7` `darwin/amd64` row with filename `go1.26.7.darwin-amd64.tar.gz`, size `67852067`, SHA-256 `92e8b34bff3c89ab16404c595669ac8cb004cc2f676dcbd1f5b87a6b8def3b47`, and the approved Go archive/signature URLs. The lock now contains exactly five non-comment records. |
| Five-tuple foundation control | PASS — the modified checker passed in a disposable archive of its original valid foundation commit `08c38b69a2cd63b4adf27873756a09e363e0c5a4` with only the candidate checker and lock overlaid. Separate disposable fixtures proved missing, duplicate, malformed, version-shifted, and extra records all fail closed. |
| Historical checker boundary | RECORDED — direct execution in the current product tree stops before the changed matrix block because the deprecated Step-5 `foundation-inputs.sha256` expects its historical `AGENTS.md` blob. The same mismatch exists at the declared change base; neither `make verify` nor hosted `make ci` invokes this deprecated whole-tree check. Historical manifests were not rewritten. |
| Shell syntax | PASS — both the modified foundation checker and unchanged bootstrap parse with `sh -n`. |
| `make verify` | PASS — Tier 3 policy, offline module checks, import/effect boundaries, formatting, shell syntax, vet, type compilation, compile-only actual-provider coverage, complete tests, harness reproducibility, and CLI reproducibility all passed. |
| Race suite | PASS — repository-pinned Go `1.26.7` with CGO race instrumentation passed `./internal/l7/... ./cmd/l7` on Darwin arm64. |
| Cross-build | PASS — `make cli-cross-build` produced the declared Darwin arm64 and amd64 CLI binaries with the pinned offline toolchain. |
| Out-of-scope controls | PASS — both workflows, `Makefile`, the bootstrap, all historical `harness/*.sha256` and `docs/artifacts/*.sha256` manifests, provider code/tests, compatibility profiles, dependencies, plugins, skills, and production source are unchanged from the exact base. No hosted check was removed, renamed, allowed to fail, or given a different runner, permission, timeout, action digest, branch-protection rule, or baseline/shadow policy. |
| Provider disposition | PASS — historical Codex Gate 1 and Claude Gate 2 facts remain candidate-bound; the audited no-model parser gates remain `NOT_RUN`. No unknown-option, typed `--max-turns`, argument, permission, output-schema, cancellation, cleanup, reviewer-immutability, scope, or containment control changed. |
| Hosted and authority boundary | PASS — this verification makes no hosted-runner claim. No new branch/PR was published and every new hosted gate remains `NOT_RUN`. The absent distinct GitHub accountable-owner and auditor/reviewer identities remain a separate fail-closed prerequisite for trusted `evaluate`. |
| Artifact budget and rollback | PASS — candidate parent is the brief commit and brief parent is the exact base. This sole future verification record and one future audit keep Tier 3 at three artifacts. Ordinary reverse-order reverts restore the exact base tree without changing historical provider or remote state. |
| Tracked and user-owned state | PASS — the isolated candidate worktree and index were clean after verification. The primary worktree still contains only the unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md`, which remained untouched and unstaged. |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

The local Go toolchain reports `go1.26.7` on `darwin/arm64`. Its copied,
repository-owned bootstrap receipt binds archive SHA-256
`020a1e8224811be75163e920bc77e0926a1390a6aeea19bdcf23f74b9d749f6d`
and existing signing fingerprints
`EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796` and
`0E225917414670F4442C250DFD533C07C264648F`. Verification reused that
authenticated cache offline; it did not bootstrap or execute an Intel host.

## Hosted-runner and external-authority state

All gates below are `NOT_RUN` for candidate
`228af7bbd29e1836ff8db59b2aab256af0b9fb9f` and for this future verification
successor:

- Ubuntu `Go 1.26.7 (baseline)`;
- Ubuntu `Go 1.27.0 (shadow)`;
- `CLI macOS 15 (arm64)`;
- `CLI macOS 15 (amd64)` and its authenticated Intel bootstrap;
- `CLI paired benchmark gate`; and
- trusted `evaluate`.

Historical public PR #1 results at audited provider head `f0e9f54c…` remain
evidence only. They do not transfer to this candidate. Publishing a remediation
branch, opening or updating a PR, running hosted CI, configuring
`L7_ACCOUNTABLE_OWNER`, or requesting GitHub reviews requires separate explicit
authorization. Even complete technical success cannot make Tier 3 ready without
real exact-head owner and independent auditor/reviewer identities distinct from
PR author `addressanup` and from each other.

## Preserved provider evidence

Historical Codex actual-host Gate 1 remains a pass only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`.

Historical Claude actual-host Gate 2 remains `NO_GO`: both exact implementer and
reviewer help invocations succeeded, both unknown-option parser controls
unexpectedly exited successfully, both invalid `--max-turns not-an-integer`
controls failed as required, and neither help surface advertised `--max-turns`.
Help advertisement remains non-dispositive; the successful unknown-option
controls remain dispositive.

The audited Codex and Claude no-model parser gates remain `NOT_RUN`. No provider
executable, version/help surface, prompt/stdin, model session, retry, fallback,
installation, compatibility promotion, or global provider configuration
participated in this implementation or verification. Build-tagged provider
tests compiled with no selected tests.

## Rollback proof

The exact implementation chain is:

1. base `481adaaec967ac34b6b27cf78509b85d5c068abc`, tree
   `d57a334696487b1d15557c9980e8a55c2dc4c930`;
2. brief `55e97c93b65c96f4f281ea090816b3205784cae5`, tree
   `5cef591260b1403e877f6d79fc43f1c8b6003edb`; and
3. implementation `228af7bbd29e1836ff8db59b2aab256af0b9fb9f`, tree
   `04283aa24cfb010d539afad18c2a8e4ccf439fdf`.

Before audit, revert this verification-record commit, implementation, and brief
in that order. After audit, revert its record first. Each sequence uses ordinary
revert commits and must restore the exact base tree. The audited provider branch,
PR #1, public repository, branch protection, historical manifests, reviews, and
other remote state remain outside this rollback.

## Verification boundary

This is implementer-run offline evidence, not an independent audit, hosted-runner
result, review, merge, release, deployment, or publication authority. The next
transition is separate explicit authorization for one independent read-only
`l7-release` audit bound to the verification successor commit/tree. Any
implementation change invalidates this record and requires fresh verification.
