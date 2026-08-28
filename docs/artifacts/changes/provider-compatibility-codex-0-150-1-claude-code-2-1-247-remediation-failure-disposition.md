# Provider Compatibility Remediation Failure Disposition

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition` |
| Risk tier | `3` |
| Status | `proposed`; test implementation is not approved |
| Base commit | `17664b48c0284982b74f9fad71e011ac32cddaf9` |
| Base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Failed brief | `438375b2d8edcec0983f9ce4eb4654a222cabd68` |
| Failed candidate | `8fba20512d1b5710104ec4b031ae9ee0f41d16a5` |
| Failed candidate tree | `7943f38db45705ce9cc1da01fb600f57e518215f` |
| Rollback checkpoint | `1db877f071a9f86ae720d7ca57d5a5e1db886072` |
| Implementer | `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false` |

## Problem

Codex actual-host gate 1 passed for failed candidate `8fba20512d1b5710104ec4b031ae9ee0f41d16a5`,
tree `7943f38db45705ce9cc1da01fb600f57e518215f`. Claude actual-host gate 2 returned
`NO_GO`: the test-owned unknown-option controls unexpectedly exited successfully
for both implementer and reviewer.

For both roles, the exact role help invocation succeeded and the invalid
`--max-turns not-an-integer` control failed as required. Neither help surface
advertised `--max-turns`. Help advertisement is non-dispositive; those successful
observations do not cure either unknown-option failure.

The failed implementation and brief were reverted through five ordinary commits.
The rollback checkpoint has exact tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`,
matching both original base `51191ad6edc670a0e73c3d152484bd57785144f7`
and prior clean-baseline disposition head
`a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a`. The failed history remains in
ancestry. An initially malformed disposition brief was added at `d2008d96f970737af1656f606805720fc823b131`
and removed by `17664b48c0284982b74f9fad71e011ac32cddaf9`; neither commit supplies assurance.

## Scope

Keep the rollback-only decision and add only fake-runtime regression coverage
that locks the failed target versions in the degraded state before semantic
provider invocation:

- Codex `codex-cli 0.150.1` must degrade after its fake version probe and must
  produce zero semantic invocations.
- Claude `2.1.247` and `2.1.247 (Claude Code)` must each degrade after their fake
  version probe and must produce zero semantic invocations.

Do not change production code or create an alternate provider argv. A future
qualification must retain unknown-option rejection, typed `--max-turns 64`, all
argument and permission restrictions, strict role output schemas, cancellation,
cleanup, reviewer immutability, scope enforcement, and containment. These
observations authorize no weaker or conditional control.

The lifecycle stays default OFF. Gates 3 through 6 remain `NOT_RUN`; the
historical Codex pass cannot transfer. This change authorizes no provider,
version, help, prompt, stdin, model, network, retry, fallback, installation,
global-configuration, external-CI, remote, merge, release, deployment, or
publication action.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition-audit.md`

Modify:

- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/claude/adapter_test.go`

No other path is authorized. In particular, production source, README,
`.l7/config.json`, dependencies, workflows, skills, global provider
configuration, remotes, historical commits, and the user-owned untracked
foundation audit remain unchanged. Scope expansion requires a revised brief and
fresh accountable-owner approval.

## Acceptance criteria

1. The base resolves to the exact clean-baseline tree, and all failed brief,
   candidate, disposition, and malformed-brief commits remain in ancestry.
2. Fake-runtime tests prove each exact failed target spelling degrades and no
   semantic provider request follows its version probe.
3. No production argument, permission, schema, parser, cancellation, cleanup,
   reviewer, scope, containment, feature-flag, or compatibility code changes.
4. Repository-pinned `make verify`, `go test -race ./internal/l7/... ./cmd/l7`,
   and `make cli-cross-build` pass without executing build-tagged actual-host
   tests or any provider executable.
5. The final diff contains only the brief and the two declared test files before
   the sole verification record is added.
6. The verification record binds the exact test implementation commit and tree.
   A distinct, separately authorized read-only `l7-release` audit is required
   afterward; any implementation successor invalidates both records.
7. The disposition requires zero actual-host gates. Any future qualification is
   a new Tier 3 change, new candidate/tree, fresh authorization, and all gates
   from the beginning.

## Offline test strategy

Use injected fake resolvers and runners only. Count version-probe requests
separately from semantic requests and require exactly one probe plus zero semantic
requests for every target spelling. Run the targeted adapter packages first,
then the repository-pinned full verification, race suite, and declared macOS
cross-builds. Inspect ancestry, exact base tree, changed paths, diff hygiene,
tracked state, and the untouched unrelated untracked file before recording
verification.

No actual provider binary, version/help surface, prompt/stdin, model session,
network, retry, fallback, external CI, or global configuration participates.

## Risks and mitigations

- A fake test could count the version probe as semantic activity. Classify
  requests by exact `--version` argv and fail on every other invocation.
- A broad version matcher could admit an adjacent or suffixed target. Exercise
  all exact failed spellings and require degraded capability for each.
- Historical Gate 1 success could be mistaken for support. Keep it bound to the
  failed candidate and state the overall `NO_GO` and zero transferable gates.
- The missing help token could be mistaken for the failure. Record it only as a
  non-dispositive observation; the two successful unknown-option controls are
  the dispositive failures.
- A successor could weaken an unrelated control while adding tests. Prohibit
  production changes and fail closed on every undeclared path.
- The unrelated untracked foundation audit could be staged accidentally. Use
  exact pathspecs and leave it untouched.

## Rollback

Before audit, revert the test commit and then this brief commit. That returns to
base `17664b48c0284982b74f9fad71e011ac32cddaf9` and exact baseline tree
`0676df022d3a2c3ab46b0344213f9e5eff80fc73`. Preserve every failed candidate,
revert, malformed-brief, and recovery commit in history; do not reset, amend,
rebase, delete, or auto-clean user work.

Reversing the five provider disposition reverts would restore an explicitly
failed candidate, not a qualified one, and requires a separate Tier 3 brief and
approval. There is no migration, installation, remote, provider-side state,
global configuration, release, deployment, or publication to reverse.

## Commit sequence and approval boundary

Completed rollback and planning recovery:

1. Five history-preserving provider reverts ending at
   `1db877f071a9f86ae720d7ca57d5a5e1db886072`.
2. Invalid brief `d2008d96f970737af1656f606805720fc823b131`.
3. `17664b48c0284982b74f9fad71e011ac32cddaf9` —
   `revert(provider): remove invalid failure disposition brief`.
4. Commit this corrected brief and stop for fresh approval bound to its exact
   addition commit.

Only after that fresh approval may the two test files change and the local
approval envelope be recorded. Offline verification and its sole record follow
the test commit. Independent audit requires separate authorization. Actual-host
activity, merge, release, deployment, and publication remain unauthorized.
