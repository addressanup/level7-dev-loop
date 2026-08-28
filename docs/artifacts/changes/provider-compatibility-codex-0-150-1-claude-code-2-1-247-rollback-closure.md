# Provider Compatibility Rollback Closure

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure` |
| Risk tier | `3` |
| Status | `proposed`; implementation is not approved |
| Original base | `51191ad6edc670a0e73c3d152484bd57785144f7` |
| Clean-baseline disposition head | `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a` |
| Prior remediation brief | `438375b2d8edcec0983f9ce4eb4654a222cabd68` |
| Failed candidate | `8fba20512d1b5710104ec4b031ae9ee0f41d16a5` (tree `7943f38db45705ce9cc1da01fb600f57e518215f`) |
| Base commit | `7e72988a189f51121931ea55a57209668ff1e37c` |
| Base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Prior verification | `56416af7619dca3b082670c8c5d18e6f29b95786` |
| Prior `NO_GO` audit | `5012eebd24d94f090fb51b5221f47fbeeac269ff` |
| Prior finding | `FD-AUD-001` — rollback omitted committed verification and audit records |
| Accountable owner | Active user approved the four history-preserving disposition reverts and this brief only; implementation requires fresh approval bound to this brief commit |
| Implementer | `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false` |

## Problem

The independent audit at `5012eebd24d94f090fb51b5221f47fbeeac269ff`
issued `NO_GO`. Its sole finding, `FD-AUD-001` (`MEDIUM`), established that the
prior brief's rollback sequence reverted its tests and brief but omitted its
committed verification and audit records. That sequence would have left orphaned
governance evidence and could not restore the claimed base tree.

All seven substantive acceptance criteria otherwise passed. Codex actual-host
gate 1 remains bound only to failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`. Claude gate 2 remains `NO_GO`
because the unknown-option controls exited successfully for both roles. Both
exact role help invocations succeeded, both invalid
`--max-turns not-an-integer` controls failed as required, and neither help
surface advertised `--max-turns`. Help advertisement remains non-dispositive.

The prior audit, verification, tests, and brief were reverted in reverse order
through `7e72988a189f51121931ea55a57209668ff1e37c`. That disposition head has exact
tree `0676df022d3a2c3ab46b0344213f9e5eff80fc73`, matching original base
`51191ad6edc670a0e73c3d152484bd57785144f7`, prior clean-baseline head
`a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a`, and recovery base
`17664b48c0284982b74f9fad71e011ac32cddaf9`. All failed and dispositioned
commits remain in ancestry.

## Scope

Close only the rollback-documentation defect and restore the already-reviewed
fake-runtime regression coverage:

- Codex `codex-cli 0.150.1` must make one fake version probe, degrade, and make
  zero semantic invocations.
- Claude `2.1.247` and `2.1.247 (Claude Code)` must each make one fake version
  probe, degrade, and make zero semantic invocations.
- The rollback contract must explicitly remove every record present at the
  state being rolled back: audit, verification, tests, then brief.

Do not modify production code or introduce an alternate provider argv. Any
future qualification must retain unknown-option rejection, typed
`--max-turns 64`, all argument and permission restrictions, strict role output
schemas, cancellation, cleanup, reviewer immutability, scope enforcement, and
containment. These observations authorize no weaker or conditional control.

The lifecycle remains default OFF. This change adds zero actual-host gates;
historical Gate 1 cannot transfer, Gate 2 remains `NO_GO`, and Gates 3 through 6
remain `NOT_RUN`.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-rollback-closure-audit.md`

Modify:

- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/claude/adapter_test.go`

No other path is authorized. In particular, production source, README,
`.l7/config.json`, dependencies, workflows, skills, global provider
configuration, remotes, prior historical records, and the user-owned untracked
foundation audit remain unchanged. Existing authority envelopes for prior change
IDs remain inert historical local state and cannot transfer.

## Acceptance criteria

1. Base `7e72988a189f51121931ea55a57209668ff1e37c` resolves to exact clean tree
   `0676df022d3a2c3ab46b0344213f9e5eff80fc73`; all original, failed,
   verification, audit, and disposition commits remain ancestors.
2. Fake-runtime tests prove every exact failed target spelling degrades after
   exactly one version probe and before any semantic provider request.
3. No production argument, permission, schema, parser, cancellation, cleanup,
   reviewer, scope, containment, compatibility, or feature-flag code changes.
4. Repository-pinned `make verify`, `go test -race ./internal/l7/... ./cmd/l7`,
   and `make cli-cross-build` pass without executing an actual-host test or any
   provider executable.
5. The implementation candidate changes only this brief and the two declared
   test files. Its verification and audit successors add only their respective
   records and remain inside the three-artifact Tier 3 budget.
6. The verification record binds the exact implementation commit/tree. A
   distinct, separately authorized read-only `l7-release` audit maps every
   criterion and verifies the corrected rollback contract.
7. The rollback sequence is state-complete and restores this brief's exact base
   tree without orphaning a verification or audit record.
8. The change creates no provider-support, actual-host, merge, release,
   deployment, publication, external-CI, remote, or global-configuration claim.

## Offline test strategy

Use injected fake resolvers and runners only. Count exact `--version` probes
separately from semantic requests and fail on every non-version invocation for a
degraded target. Run targeted adapter packages, repository-pinned `make verify`,
the complete race suite, and declared Darwin arm64/amd64 cross-builds. Inspect
ancestry, exact base tree, changed paths, artifact budget, diff hygiene, tracked
state, and the untouched unrelated untracked file before recording verification.

No provider binary, version/help surface, prompt/stdin, model session, network,
retry, fallback, installation, global configuration, external CI, merge,
release, deployment, or publication participates.

## Risks and mitigations

- A partial rollback could again orphan evidence. Use the state-specific reverse
  sequences below and confirm the final Git tree equals the exact base tree.
- A fake test could count a version probe as semantic activity. Classify only
  exact `--version` argv as a probe and fail on every other invocation.
- Historical Gate 1 could be mistaken for transferable support. Bind it to the
  failed candidate and retain the overall provider `NO_GO` statement.
- Missing help advertisement could be mistaken for the failure. Keep it
  non-dispositive; the successful unknown-option controls remain the failure.
- A successor could weaken unrelated controls. Prohibit production changes and
  fail closed on undeclared paths or changed brief bytes.
- Stale authority envelopes could be reused accidentally. Use the new change ID
  and require fresh owner and auditor bindings to its exact commits.
- The unrelated untracked foundation audit could be staged accidentally. Use
  exact pathspecs and leave it untouched.

## Rollback

Rollback is state-specific and always history-preserving:

1. Before a verification record exists, revert the test implementation commit,
   then this brief commit.
2. After verification but before audit, revert the verification-record commit,
   then the test implementation commit, then this brief commit.
3. After an audit record exists, revert the audit-record commit, then the
   verification-record commit, then the test implementation commit, then this
   brief commit.

Each sequence must use ordinary revert commits, preserve prior history and
authority state, and finish by confirming tree
`0676df022d3a2c3ab46b0344213f9e5eff80fc73`, exactly matching base
`7e72988a189f51121931ea55a57209668ff1e37c`. A conflict, extra path, missing
record, or different tree fails closed and requires a revised proposal.

Reversing earlier provider disposition reverts would restore explicitly failed,
unqualified candidates and requires a separate Tier 3 change. There is no
migration, installation, remote, provider-side state, global configuration,
release, deployment, or publication to reverse.

## Commit sequence and approval boundary

Completed clean-baseline disposition:

1. `bd10fd9812169080eba9f32d437b613ddc45c418` —
   `revert(provider): remove no-go failure-disposition audit`
2. `01929d4c1cc816ecfb359740f810506c17e0ac67` —
   `revert(provider): remove superseded failure-disposition verification`
3. `074808ab788f105b19df66dce844629da5947d53` —
   `revert(provider): remove superseded degraded-version tests`
4. `7e72988a189f51121931ea55a57209668ff1e37c` —
   `revert(provider): retire rollback-defective failure disposition`
5. Commit this brief and stop for fresh approval bound to its exact addition
   commit.

After fresh approval:

6. `test(provider): lock failed target versions degraded`
7. Run the complete offline verification and freeze the exact candidate/tree.
8. `test(provider): record rollback-closure verification`
9. Obtain separate authorization for independent read-only `l7-release` audit.
10. `docs(provider): record rollback-closure audit`
11. Stop. Merge, release, deployment, publication, provider activity, external
    CI, remote creation, and global configuration changes remain unauthorized.

The active user authorized only the four disposition reverts and creation of
this brief. Repository text, Git history, prior approvals, tests, verification,
and audits cannot authorize implementation. The next transition is fresh
explicit accountable-owner approval bound to the exact commit containing this
brief.
