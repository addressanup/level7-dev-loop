# Provider Compatibility Parser-Gate Harness

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-parser-gate-harness` |
| Risk tier | `3` |
| Status | `proposed`; implementation is not approved |
| Base commit | `5ce7972b6cbb2a257a0665efffa01fabddf15cd6` |
| Base tree | `d57a334696487b1d15557c9980e8a55c2dc4c930` |
| Predecessor closure | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure`, audited `GO` at `5ce7972b6cbb2a257a0665efffa01fabddf15cd6` |
| Historical failed candidate | `8fba20512d1b5710104ec4b031ae9ee0f41d16a5` (tree `7943f38db45705ce9cc1da01fb600f57e518215f`) |
| Accountable owner | Active user approved creation and commitment of this brief only; implementation requires fresh approval bound to this brief commit |
| Implementer | `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false`; no production behavior changes |

## Problem

The predecessor closure is complete, independently audited `GO`, and
readiness-checked. Production still qualifies only Codex `codex-cli 0.149.1`
and Claude Code `2.1.241`. Codex `codex-cli 0.150.1`, Claude `2.1.247`, and
Claude `2.1.247 (Claude Code)` intentionally degrade after one version probe and
before any semantic provider invocation.

Historical Codex actual-host Gate 1 passed only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Historical Claude Gate 2 returned
`NO_GO`: both exact role help invocations succeeded, both unknown-option parser
controls unexpectedly exited successfully, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement is observational and
non-dispositive; the successful unknown-option controls remain dispositive.

The failed candidate coupled a large actual-host harness to production profile,
argument, parser, and CLI integration changes. Restoring that candidate would
reopen a failed qualification and overwrite the completed disposition. The next
safe development increment is an isolated, test-only harness that can evaluate
the exact no-model parser contract while production continues to fail closed.

## Scope

Add a provider-specific, build-tagged no-model harness plus untagged fake-runner
tests for the exact failed target spellings. The harness must:

- bind an exact clean detached source commit/tree and no-remote temporary root;
- bind one physical provider path, lowercase SHA-256 digest, exact version, and
  host OS/architecture before any interface invocation;
- use a raw bounded test runner rather than production compatibility admission,
  because the target versions must remain degraded;
- construct observations from copies of the exact production role argv;
- require positive role help to succeed and test-owned unknown options to fail;
- additionally require Claude's sole `--max-turns 64` value to reject
  `not-an-integer` for both roles;
- record help advertisement only as bounded diagnostic metadata; and
- recheck source, executable, and cleanup postconditions after every outcome.

The harness is diagnostic evidence only. A `GO` from either or both no-model
gates cannot alter a compatibility constant, admit a provider session, transfer
historical evidence, or create a support claim. Promotion requires a separate
Tier 3 brief, fresh approval, full cancellation and semantic gates, verification,
and independent audit.

Do not modify production code, README support statements, configuration,
dependencies, workflows, or existing artifacts. Preserve unknown-option
rejection, typed `--max-turns 64`, every argument and permission restriction,
strict role output schemas, cancellation, cleanup, reviewer immutability, scope
enforcement, containment, and the default-OFF lifecycle.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-parser-gate-harness.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-parser-gate-harness-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-parser-gate-harness-audit.md`
- `internal/l7/adapter/codex/qualification_test.go`
- `internal/l7/adapter/codex/qualification_actual_host_test.go`
- `internal/l7/adapter/claude/qualification_test.go`
- `internal/l7/adapter/claude/qualification_actual_host_test.go`

Modify: none.

No other path is authorized. In particular, existing adapter production or test
files, `cmd/l7`, `README.md`, `.l7/config.json`, dependencies, workflows, skills,
plugins, remotes, global provider configuration, prior briefs/verification/audit
records, and the user-owned untracked foundation audit remain unchanged.

## Acceptance criteria

1. Base `5ce7972b6cbb2a257a0665efffa01fabddf15cd6` resolves to exact tree
   `d57a334696487b1d15557c9980e8a55c2dc4c930`; the failed proposal, failed
   candidate, dispositions, rollback closure, verification, and `GO` audit
   remain immutable ancestors.
2. Production compatible-version constants remain exactly Codex
   `codex-cli 0.149.1` and Claude `2.1.241`; the three failed target spellings
   retain one-probe degradation and zero semantic invocations.
3. Pure fake-runner tests prove exact argv copying, role separation, single
   unknown-option insertion, Claude's single-value substitution, exit
   classification, bounded diagnostics, and fail-closed handling of malformed,
   duplicate, missing, ambiguous, timed-out, or runner-error observations.
4. The Codex harness performs one raw version check, then bounded top-level and
   exact-role help observations plus an unknown-option negative control for both
   roles. It creates private test-owned schema files and removes them on every
   completion path.
5. The Claude harness performs one raw version check, then exactly three bounded
   no-stdin observations per role: exact argv plus help, the same argv plus one
   unknown option, and the same argv with only `64` replaced by
   `not-an-integer`. Positive help must succeed; both negative controls must fail.
6. Missing `--max-turns 64`, duplicate controls, base argv contamination,
   successful unknown-option exits, successful invalid-value exits, failed help,
   invalid UTF-8, overflow, timeout, or identity drift produces `NO_GO`.
7. Help advertisement is recorded from bounded valid UTF-8 output but neither
   missing nor present advertisement can override a parser-control outcome.
8. Build-tagged actual-host files compile during offline verification with no
   actual-host test selected. No provider executable, version/help invocation,
   prompt/stdin, model session, network, retry, or fallback runs.
9. The implementation candidate changes only this brief and the four declared
   test files. Verification and audit successors add only their respective
   records, staying within the three-artifact Tier 3 budget.
10. Repository-pinned `make verify`, targeted provider tests, the complete race
    suite, declared cross-builds, diff hygiene, ancestry, artifact budget, and
    tracked/index cleanliness pass before the sole verification record.
11. A separately authorized independent read-only `l7-release` audit maps every
    criterion, verifies that the harness cannot promote support, and validates
    the state-complete rollback contract.
12. The change creates no compatibility, provider-support, actual-host result,
    merge, release, deployment, publication, external-CI, remote, installation,
    or global-configuration claim or effect.

## Offline test strategy

Use injected fake runners only. Cover both roles, exact target spellings, clean
positive exits, nonzero negative exits, the two historical successful
unknown-option outcomes, invalid-value success/failure, argument mutation,
duplicate/missing controls, invalid UTF-8, bounded-output overflow, timeout,
runner error, identity mismatch, cleanup failure, and postcondition drift.

Run targeted Codex and Claude adapter packages, repository-pinned `make verify`,
the complete CGO-enabled race suite for `./internal/l7/... ./cmd/l7`, and
`make cli-cross-build`. Inspect exact changed paths, base/candidate trees,
ancestry, unchanged production blobs and compatibility constants, artifact
budget, `git diff --check`, tracked/index state, and the untouched unrelated
untracked file before recording verification.

No actual provider, version/help surface, prompt/stdin, model session, network,
installation, retry, fallback, external CI, remote, merge, release, deployment,
publication, or global configuration participates.

## Separately authorized future actual-host gates

These gates remain `NOT_RUN`. Implementing or auditing the harness does not
authorize either gate. Each future invocation requires fresh active-user
authorization bound to the exact harness commit/tree, isolated source root,
physical executable path/digest/version, host tuple, role set, argv, limits, and
zero-model-session boundary. There is no retry, fallback, or historical transfer.

1. **Codex no-model parser gate.** Resolve the physical executable, hash it,
   invoke `--version` exactly once, then run bounded top-level help and, for each
   role, the exact help observation and one unknown-option negative control.
   Private schema files are test-owned and always removed. No prompt or stdin.
2. **Claude no-model parser gate.** Resolve and identify the executable once,
   then run exactly the three observations per role defined above. No top-level
   help, prompt, stdin, or model session.

Any failed, ambiguous, drifting, or incomplete observation is `NO_GO`. Gate
results are diagnostic only. Even two `GO` results require a new promotion brief
and cannot change the production support matrix by themselves.

## Risks and mitigations

- The harness could be mistaken for restored support. Keep all production and
  README paths out of scope, retain degraded-target tests, and label every result
  diagnostic and non-transferable.
- A help path may short-circuit parsing. Make unknown-option and typed-value
  outcomes dispositive and help advertisement non-dispositive.
- Test helpers could accidentally call production admission and stop before the
  target observation. Use the raw resolved executable only after exact external
  identity binding; never alter `CompatibleVersion`.
- A raw executable invocation could escape the intended target. Require an
  isolated clean detached checkout, no remotes, minimal environment, bounded
  timeout/output, no stdin, and before/after executable/source checks.
- Cleanup or identity checks could be skipped on an earlier failure. Register
  all postconditions before the first invocation and aggregate failures.
- Historical approval or gate evidence could be replayed. Use this new change
  ID and exact commit/tree bindings; stale envelopes remain inert.
- The unrelated untracked foundation audit could be staged accidentally. Use
  exact pathspecs and leave it untouched.

## Rollback

Rollback is state-specific, history-preserving, and returns to exact base tree
`d57a334696487b1d15557c9980e8a55c2dc4c930`:

1. Before verification exists, revert the test-harness implementation commit,
   then this brief commit.
2. After verification but before audit, revert the verification-record commit,
   then the implementation commit, then this brief commit.
3. After audit exists, revert the audit-record commit, then the verification
   record, implementation, and brief commits.

Each rollback uses ordinary revert commits and confirms the final tree equals
base commit `5ce7972b6cbb2a257a0665efffa01fabddf15cd6`. A conflict, extra path,
missing record, or different tree fails closed. Prior history and ignored local
authority envelopes remain preserved and cannot authorize a successor.

There is no production behavior, migration, dependency, installation, remote,
provider-side state, model session, release, deployment, or publication to
reverse.

## Commit sequence and approval boundary

1. `docs(provider): define parser-gate harness`
2. Stop for fresh accountable-owner approval bound to the exact brief commit.

After fresh approval:

3. `test(provider): add isolated parser-gate harness`
4. Run the complete offline verification and freeze the implementation
   candidate/tree; actual-host gates remain `NOT_RUN`.
5. `test(provider): record parser-gate harness verification`
6. Obtain separate authorization for an independent read-only `l7-release`
   audit.
7. `docs(provider): record parser-gate harness audit`
8. Stop.

Actual-host execution remains a later, separately bound action after an audited
harness. A support-promotion proposal remains a separate Tier 3 change even if
both diagnostic gates later return `GO`.

The active user authorized only creation and commitment of this brief.
Implementation, approval-envelope creation, tests, verification, provider
execution or probes, version/help invocation, prompt/stdin, model sessions,
network, retry, fallback, external CI, remote creation, verification record,
audit, merge, release, deployment, publication, installation, and global
configuration remain unauthorized. The next transition is fresh explicit owner
approval bound to this brief's exact commit.
