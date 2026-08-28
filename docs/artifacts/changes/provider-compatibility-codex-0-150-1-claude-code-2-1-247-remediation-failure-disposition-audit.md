# Provider Compatibility Remediation Failure Disposition — Independent Audit

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition` |
| Candidate commit | `56416af7619dca3b082670c8c5d18e6f29b95786` |
| Candidate tree | `88eb7f69475cc8c1ac804882cbe35d4bfec4f459` |
| Result | `NO_GO` |
| Reviewer | `l7-release-independent-auditor` |
| Audited at | `2026-08-28T05:25:41Z` |
| Verified implementation | `83846ecea259e3f966060cc8a7d94c98de78638d` |
| Implementation tree | `ba549d80d842afbe32c7e38999d7dda7e6d27688` |

## Decision

`NO_GO`. The exact verification commit and tree, approval binding, ancestry,
scope, test-only implementation, fake-runtime behavior, verification record,
artifact budget, and actual-host boundary were inspected independently. Seven
acceptance criteria pass, but one unresolved medium-severity rollback defect
prevents Tier 3 acceptance.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Clean baseline and preserved history | PASS — base `17664b48c0284982b74f9fad71e011ac32cddaf9` resolves to tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`; original base `51191ad6…`, clean-baseline head `a3b40cbe…`, failed brief `438375b2…`, failed candidate `8fba2051…`, five ordinary reverts through `1db877f0…`, malformed brief `d2008d96…`, and its recovery revert remain ancestors. |
| 2. Exact degraded target coverage | PASS by source inspection — Codex `codex-cli 0.150.1` and Claude `2.1.247` plus `2.1.247 (Claude Code)` each return the exact failed version through one fake `--version` probe, require degraded capability and the fail-closed contract error, and require zero semantic invocations. |
| 3. No production-control change | PASS — base-to-implementation changes are only the approved brief and 41 added test lines in the two adapter test files. Production arguments, permissions, schemas, parser behavior, cancellation, cleanup, reviewer controls, scope, containment, compatibility constants, and the default-OFF feature flag are unchanged. |
| 4. Offline verification | PASS on bound implementer evidence — the sole verification record reports PASS for targeted fake-runtime tests, repository-pinned `make verify`, the full race suite, and Darwin arm64/amd64 cross-builds. The Makefile confirms actual-host tests are compile-only under `-run '^$'`. This read-only audit did not rerun write-producing tests or builds. |
| 5. Exact scope | PASS — implementation candidate changes only the brief and two declared adapter test files; its only successor adds the sole verification record. `git diff --check` is clean and no undeclared production, configuration, dependency, workflow, remote, or global-provider path changed. |
| 6. Verification and independence binding | PASS — verification binds implementation `83846ecea259e3f966060cc8a7d94c98de78638d` and tree `ba549d80d842afbe32c7e38999d7dda7e6d27688`; verification commit `56416af7619dca3b082670c8c5d18e6f29b95786` is its direct audit target. Approval binds brief addition `9e942b8e486a55c50f137793be0aec262eb713da`; owner `accountable-owner`, implementer `codex-root`, and reviewer `l7-release-independent-auditor` are distinct. |
| 7. Actual-host boundary | PASS — the disposition adds zero actual-host gates. Historical Codex Gate 1 remains candidate-bound, Claude Gate 2 remains `NO_GO`, Gates 3–6 remain `NOT_RUN`, and help advertisement remains non-dispositive. No target version is qualified by this change. |

## Findings

### `FD-AUD-001` — MEDIUM — rollback omits the committed verification record

The approved Rollback section says to revert test commit
`83846ecea259e3f966060cc8a7d94c98de78638d` and then brief commit
`9e942b8e486a55c50f137793be0aec262eb713da`, claiming that this returns to base
`17664b48c0284982b74f9fad71e011ac32cddaf9` and exact tree
`0676df022d3a2c3ab46b0344213f9e5eff80fc73`.

At the audited state, the verification record added by
`56416af7619dca3b082670c8c5d18e6f29b95786` is also present. Reverting only the
test and brief commits leaves that artifact in the tree, so the documented
sequence cannot restore the asserted base tree and leaves verification evidence
without its governing brief and candidate.

Disposition: unresolved and decision-blocking. A fresh approved proposal must
preserve history and explicitly include reverse-order removal of the audit
record if committed, verification commit `56416af…`, test commit `83846ece…`,
and brief commit `9e942b8e…`, followed by confirmation of tree `0676df02…`. No
production remediation is indicated.

## Severity summary

| Severity | Count |
|---|---|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 1 |
| LOW | 0 |
| NOTE | 0 |

## Read-only and actual-host boundary

The auditor used read-only Git object, ancestry, tree, diff, source, approval
envelope, and policy inspection. It did not write or alter files, Git refs,
`.git/l7` authority state, configuration, remotes, or external systems. It did
not run tests/builds, invoke any provider executable, execute version/help
probes, provide prompts/stdin, create a model session, retry/fallback, merge,
release, deploy, or publish. The unrelated untracked foundation audit was not
inspected or touched.

This `NO_GO` authorizes only mechanical commitment of this exact sole audit
record and creation of its matching external audit envelope, with reviewer
`l7-release-independent-auditor`, candidate
`56416af7619dca3b082670c8c5d18e6f29b95786`, tree
`88eb7f69475cc8c1ac804882cbe35d4bfec4f459`, and the resulting audit commit. It
authorizes no implementation, rollback, remediation, re-verification, merge,
release, deployment, publication, provider-support claim, or actual-host
activity.
