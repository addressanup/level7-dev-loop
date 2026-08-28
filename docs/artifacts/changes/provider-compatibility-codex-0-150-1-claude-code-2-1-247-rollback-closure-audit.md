# Provider Compatibility Rollback Closure — Independent Audit

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure` |
| Candidate commit | `b107b581087bcafa774f417010e1f38d9cb2a486` |
| Candidate tree | `8a4de28108e5c3d486594a52992b5b012243d9f4` |
| Result | `GO` |
| Reviewer | `l7-release-independent-auditor` |
| Audited at | `2026-08-28T08:27:53Z` |
| Verified implementation | `4828519b7b021467826a0a906dc0551501bd610f` |
| Implementation tree | `1852959d7a69c889a9f062a61d02b620c2d35cea` |
| Base commit | `7e72988a189f51121931ea55a57209668ff1e37c` |
| Base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |

## Decision

`GO`. The exact verification commit/tree, approved brief, authority binding,
ancestry, scope, fake-runtime tests, sole verification record, artifact budget,
historical actual-host facts, and all three rollback states were independently
inspected. All eight acceptance criteria pass. The prior `FD-AUD-001` finding is
closed without a production-code change.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Exact base and preserved history | PASS — base `7e72988a189f51121931ea55a57209668ff1e37c` resolves to tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`, matching original base `51191ad6…`, clean-baseline head `a3b40cbe…`, and recovery base `17664b48…`. The original, failed, prior verification, prior `NO_GO` audit, and all disposition commits remain ancestors of the audited commit. |
| 2. Exact degraded-target behavior | PASS by source inspection — Codex `codex-cli 0.150.1`, Claude `2.1.247`, and Claude `2.1.247 (Claude Code)` each return their exact spelling through one fake `--version` probe, require degraded capability and the fail-closed contract error, and require zero semantic invocations. The rejection occurs in the shared pre-role path, before role-specific argument construction. |
| 3. Production controls unchanged | PASS — base-to-implementation changes are the approved brief plus 41 added test lines in the two adapter test files. Production arguments, permissions, typed `--max-turns 64`, schemas, parser behavior, cancellation, cleanup, reviewer restrictions, scope, containment, compatible-version constants, and the default-OFF feature flag are unchanged. |
| 4. Offline verification | PASS on bound implementer evidence — the sole verification record reports PASS for targeted fake-runtime tests, repository-pinned `make verify`, `go test -race ./internal/l7/... ./cmd/l7`, and `make cli-cross-build`. The pinned Makefile keeps build-tagged actual-host coverage compile-only with no selected tests. This read-only audit did not rerun tests or builds. |
| 5. Scope and artifact budget | PASS — implementation candidate `4828519b…` changes only the brief and two declared test files. Verification commit `b107b581…` adds only the sole verification record. The audit record is the one remaining authorized artifact, producing exactly the Tier 3 maximum of three records. No production, configuration, dependency, workflow, skill, remote, or global-provider path changed. |
| 6. Verification and independent authority | PASS — verification binds implementation `4828519b7b021467826a0a906dc0551501bd610f` and tree `1852959d7a69c889a9f062a61d02b620c2d35cea`; its direct successor and audit target is `b107b581087bcafa774f417010e1f38d9cb2a486`, tree `8a4de28108e5c3d486594a52992b5b012243d9f4`. Fresh approval binds brief addition `afed168a9f7313e8f7dad9fa8aa97b7a64155588`. Owner `accountable-owner`, implementer `codex-root`, and reviewer `l7-release-independent-auditor` are distinct. |
| 7. State-complete rollback | PASS — the brief commit is the direct child of the exact base, the test commit is the direct child of the brief, and the verification record is the direct child of the test commit. The declared reverse sequences remove every record present in each state and restore exact base tree `0676df02…`; an audit-only successor preserves the same proof after its audit record is reverted first. |
| 8. Claim and effect boundary | PASS — the change creates no provider-support, actual-host, merge, release, deployment, publication, external-CI, remote, or global-configuration claim or effect. Historical actual-host evidence remains expressly candidate-bound and non-transferable. |

## Rollback closure

| State | Required reverse sequence | Independent result |
|---|---|---|
| Pre-verification implementation `4828519b…` | Tests, then brief | PASS — direct-parent reversal returns to `7e72988a…` and tree `0676df02…`. |
| Post-verification `b107b581…` | Verification record, tests, then brief | PASS — removes the sole verification evidence before its candidate and brief, leaving no orphaned record and restoring `0676df02…`. |
| Post-audit audit-only successor | Audit record, verification record, tests, then brief | PASS — provided mechanical materialization adds only this audit record, reverse order removes every state-specific artifact and restores `0676df02…`. |

All rollback steps require ordinary revert commits, preserved ancestry, fail-closed
conflict handling, and final exact-tree confirmation. External authority
envelopes remain inert historical state and cannot qualify a future candidate.

## Historical actual-host boundary

Codex Gate 1 remains a pass bound only to failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Claude Gate 2 remains `NO_GO`
because the unknown-option controls unexpectedly exited successfully for both
implementer and reviewer roles.

Both exact role help invocations succeeded, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement remains non-dispositive.
Gates 3 through 6 remain `NOT_RUN`, and no gate transfers to this change.

## Finding closure and severity

`FD-AUD-001` — `CLOSED`. The corrected contract explicitly removes audit and
verification successors before reverting tests and the brief, covering
pre-verification, post-verification, and post-audit states.

| Severity | Unresolved count |
|---|---|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |
| NOTE | 0 |

## Read-only and materialization boundary

The auditor used read-only Git object, ancestry, tree, diff, source, approval,
artifact, and policy inspection. It did not alter files, refs, index, worktree,
`.git/l7` state, configuration, remotes, or external systems. It did not run
tests, builds, controller or provider executables, version/help probes,
prompts/stdin, model sessions, network activity, retries, fallbacks, merges,
releases, deployments, or publication. The unrelated untracked foundation audit
was not inspected or touched.

This `GO` authorizes only mechanical commitment of this exact sole audit record
as the only tree change after
`b107b581087bcafa774f417010e1f38d9cb2a486`, followed by creation of the matching
external audit envelope with schema `1`, change ID
`provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure`,
actor `l7-release-independent-auditor`, candidate commit
`b107b581087bcafa774f417010e1f38d9cb2a486`, the resulting audit commit, and
source `independent-agent`.

It authorizes no implementation, rollback, remediation, re-verification, merge,
release, deployment, publication, provider-support claim, actual-host activity,
external CI, remote creation, or global-configuration change. After faithful
materialization, stop.
