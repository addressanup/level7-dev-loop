# Provider Compatibility No-Model Gate Harness — Independent Audit

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness` |
| Candidate commit | `0af3c0e28e58ba28770b83e0e9159583e23a93dc` |
| Candidate tree | `bfc41c4667e31bb31d31a6b5ee1aef2310f8d3ce` |
| Result | `GO` |
| Reviewer | `l7-release-independent-auditor` |
| Audited at | `2026-08-28T09:12:55Z` |
| Verified implementation | `53c04388e96407fd361f974a76e9c5fcda29e0d3` |
| Implementation tree | `6167a55f4c8889a90dbd1b766a924e9d8cbabb0b` |
| Base commit | `481adaaec967ac34b6b27cf78509b85d5c068abc` |
| Base tree | `d57a334696487b1d15557c9980e8a55c2dc4c930` |

## Decision

`GO`. The exact approved brief, implementation, verification successor, Git
lineage, authority binding, five-path implementation scope, production
boundaries, fake-runner behavior, tagged actual-host containment, diagnostic-only
contract, artifact budget, historical evidence, and state-complete rollback were
independently inspected. All twelve acceptance criteria pass.

This decision validates the harness bytes only. It does not run or authorize
either actual-host gate and cannot promote either provider version.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Base and immutable history | PASS — base `481adaaec967ac34b6b27cf78509b85d5c068abc` resolves to tree `d57a334696487b1d15557c9980e8a55c2dc4c930`. Original base `51191ad6…`, disposition `a3b40cbe…`, failed brief `438375b2…`, failed candidate `8fba2051…`, rollback-closure audit `5ce7972b…`, inconsistent brief `9bcb10a6…`, and its ordinary revert remain ancestors. |
| 2. Production compatibility and degradation | PASS — production adapter blobs are byte-identical to the base. Codex remains exactly `codex-cli 0.149.1`; Claude remains exactly `2.1.241`. Existing regression tests retain one-probe degradation and zero semantic invocations for `codex-cli 0.150.1`, `2.1.247`, and `2.1.247 (Claude Code)`. |
| 3. Pure fake-runner behavior | PASS by source inspection — untagged tests derive fresh observations from production `arguments` for both roles, preserve input slices, enforce exact mutations and request counts, classify positive and negative exits, bound combined diagnostics, and reject unsafe argv, wrong versions, empty/invalid/overflow output, timeouts, runner errors, and ambiguous exits. |
| 4. Codex harness | PASS — one raw exact-version request precedes bounded top-level help plus implementer/reviewer help and unknown-option observations. Each role starts from current production argv, requires exactly one final `-`, replaces only that sentinel with `--help`, and inserts exactly one test-owned unknown option immediately before help. Every request has empty stdin; no schema file or production admission path exists. |
| 5. Claude harness | PASS — one raw exact-version request precedes exactly three observations for each role. Positive help appends only `--help`; unknown-option inserts exactly one test-owned option immediately before help; invalid-value changes only the unique `--max-turns 64` value to `not-an-integer` before appending help. Positive help must exit zero and both negative controls must exit nonzero. |
| 6. Fail-closed controls and postconditions | PASS — malformed, missing, duplicate, contaminated, failed, unexpectedly successful, empty, invalid UTF-8, overflowing, timed-out, runner-error, and ambiguous observations return errors. Checked runners execute postconditions after earlier gate errors; actual-host tests register source and executable cleanup checks before the first provider interface invocation. Identity or cleanup failure prevents a successful gate result. |
| 7. Help advertisement | PASS — advertisement is derived only from bounded valid UTF-8 positive-help output and retained as diagnostic metadata. Tests prove missing advertisement does not reject valid positive help and present advertisement cannot override a successful negative control or alter a rejecting negative result. |
| 8. Offline actual-host boundary | PASS — actual-host files are protected by `l7_actual_provider`; repository verification compiles them with `-run '^$'`, selecting no test. There is no `init` or `TestMain` side path. The verification record states that no provider executable, version/help interface, prompt/stdin, model session, network, retry, or fallback ran. Both new gates remain `NOT_RUN`. |
| 9. Exact scope and artifact budget | PASS — base-to-implementation adds only the approved brief and four declared test files; no existing file changes. Implementation-to-verification adds only the sole verification record. The future audit is the third and final permitted Tier 3 artifact. `git diff --check` passes. |
| 10. Bound offline verification | PASS on the sole implementer record — targeted qualification tests, repository-pinned `make verify`, the complete race suite, Darwin arm64/amd64 cross-builds, ancestry, scope, artifact, diff, and state checks report PASS against implementation `53c04388…`, tree `6167a55f…`. This strictly read-only audit did not rerun tests, builds, or the controller. |
| 11. Independence, non-promotion, and rollback | PASS — fresh external approval binds unchanged brief addition `8ab7412a63e34838c49027050fcf6e6f68b6e65c`; owner `accountable-owner`, implementer `codex-root`, and reviewer `l7-release-independent-auditor` are distinct. No harness result reaches production compatibility admission. The direct commit chain and state-specific reverse sequences restore the exact base tree without orphaning records. |
| 12. Claim and effect boundary | PASS — all implementation code is test-only. The change creates no compatibility, provider-support, actual-host result, merge, release, deployment, publication, external-CI, remote, installation, or global-configuration claim or effect. Future diagnostic success still requires a separate Tier 3 promotion change. |

## Actual-host containment

Both tagged tests fail closed unless their exact provider-specific gate token is
present. Before any provider interface invocation they require:

- an authorized absolute physical source root inside an authorized temporary
  parent;
- an owned `.git` directory and matching Git common directory;
- detached exact candidate and tree identities;
- a clean source root with no remotes;
- exact host OS and architecture;
- an absolute, non-symlinked physical provider path;
- a lowercase SHA-256 executable digest matching fresh resolution; and
- the exact authorized target version spelling.

The raw process runner uses a minimal environment, bounded output, bounded
timeouts, an empty stdin reader, and process-group cancellation. It does not call
production compatibility admission. Source and executable checks run through the
checked postcondition after success or earlier gate error and are also registered
as test cleanups.

Codex would make one version request and five no-input observations. Claude
would make one version request and six no-input observations. Neither sequence
contains a prompt, semantic task, retry, or fallback. These are future
diagnostic gates only and were not invoked by implementation, verification, or
this audit.

## Historical evidence and gate state

Historical Codex Gate 1 remains `PASS` only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`.

Historical Claude Gate 2 remains `NO_GO` because both implementer and reviewer
unknown-option controls unexpectedly exited successfully. Both exact role help
calls succeeded, both invalid `--max-turns not-an-integer` controls failed as
required, and neither help surface advertised `--max-turns`. Help advertisement
is non-dispositive; the two successful unknown-option controls remain
dispositive.

The new Codex no-model parser gate and Claude no-model parser gate are both
`NOT_RUN`. Historical evidence cannot transfer to them, and diagnostic results
cannot change production compatibility or support.

## Rollback proof

| State | Reverse sequence | Independent result |
|---|---|---|
| Pre-verification implementation `53c04388…` | Implementation, then brief | PASS — the implementation is the direct child of brief `8ab7412a…`, which is the direct child of base `481adaae…`; ordinary reverse reverts restore tree `d57a3346…`. |
| Post-verification `0af3c0e2…` | Verification, implementation, then brief | PASS — the verification record is the direct and only successor change after implementation, so reverse order removes all current change artifacts and restores `d57a3346…`. |
| Post-audit audit-only successor | Audit, verification, implementation, then brief | PASS — faithful mechanical materialization adds only this audit record; reverting it first preserves the same direct reverse proof and leaves no orphaned verification or audit artifact. |

Every rollback must use ordinary revert commits, preserve history, fail closed on
conflict or scope expansion, and confirm exact final tree
`d57a334696487b1d15557c9980e8a55c2dc4c930`.

## Findings and severity

No unresolved findings.

| Severity | Count |
|---|---|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |

## Read-only and materialization boundary

The auditor used only read-only Git object, tree, ancestry, diff, source,
authority, artifact, and status inspection. It did not edit files, index, refs,
authority envelopes, configuration, remotes, or external systems. It did not run
tests, builds, controller or provider executables, version/help probes,
prompts/stdin, model sessions, network activity, retries, fallbacks, CI, merge,
release, deployment, or publication.

Status confirmed only that
`docs/artifacts/foundation-rebaseline-admission-audit.md` remains untracked. Its
contents were not inspected or touched.

This `GO` authorizes only faithful mechanical creation and commitment of this
exact sole audit record as the only repository-tree change after
`0af3c0e28e58ba28770b83e0e9159583e23a93dc`, followed by creation of the
matching external audit envelope with schema `1`, change ID
`provider-compatibility-codex-0-150-1-claude-code-2-1-247-no-model-gate-harness`,
actor `l7-release-independent-auditor`, candidate commit
`0af3c0e28e58ba28770b83e0e9159583e23a93dc`, the resulting audit commit, and
source `independent-agent`.

It authorizes no actual-host gate, provider execution, probe, version/help
invocation, prompt/stdin, model session, network, retry, fallback, implementation,
remediation, rollback, re-verification, compatibility promotion, support claim,
external CI, remote creation, configuration change, merge, release, deployment,
installation, publication, or any second repository file. After faithful
materialization, stop.
