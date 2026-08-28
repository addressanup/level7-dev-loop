# Provider Compatibility Remediation Failure Disposition

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-failure-disposition` |
| Risk tier | `3` |
| Status | `dispositioned`; offline verification and audit are not authorized |
| Original base | `51191ad6edc670a0e73c3d152484bd57785144f7` |
| Original base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Prior clean-baseline disposition head | `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a` |
| Failed brief | `438375b2d8edcec0983f9ce4eb4654a222cabd68` |
| Failed candidate | `8fba20512d1b5710104ec4b031ae9ee0f41d16a5` |
| Failed candidate tree | `7943f38db45705ce9cc1da01fb600f57e518215f` |
| Disposition head | `1db877f071a9f86ae720d7ca57d5a5e1db886072` |
| Disposition tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Accountable owner | Active user approved the exact rollback-only proposal; the approval covered the five history-preserving reverts and this brief only |
| Feature flag | Existing `features.local_lifecycle`, default `false` |

## Decision and observations

The candidate is `NO_GO`. Codex actual-host gate 1 passed for the exact failed
candidate and tree, but that result is candidate-bound and cannot transfer.
Claude actual-host gate 2 returned `NO_GO` because the test-owned unknown-option
controls unexpectedly exited successfully for both implementer and reviewer.

For both roles, the exact role help invocation succeeded and the invalid
`--max-turns not-an-integer` control failed as required. Neither help surface
advertised `--max-turns`. Help advertisement is an observation only and is not
an acceptance or rejection oracle. The successful help and typed-value results
do not cure either unknown-option failure.

Gates 3 through 6 were not run because gate 2 stopped the sequence. No historical
gate result qualifies the disposition or any future candidate.

## Disposition and control boundary

The selected path is rollback only. Five ordinary revert commits preserve the
failed brief and implementation in ancestry while restoring the exact original
base tree. There was no reset, amend, rebase, history deletion, or alternate
weaker provider invocation.

Codex `codex-cli 0.150.1` and Claude Code `2.1.247` again degrade before provider
execution. Any future qualification must retain unknown-option rejection, typed
`--max-turns 64`, every argument and permission restriction, strict role output
schemas, cancellation, cleanup, reviewer immutability, scope enforcement, and
containment. These observations authorize no weakened or conditional control.

The default-OFF lifecycle boundary remains unchanged. The rollback makes no
provider-support, release, deployment, or publication claim.

## Exact file scope

The approved reverts restored these paths:

- `cmd/l7/actual_host_test.go`
- `cmd/l7/execution_test.go`
- `internal/l7/app/execution_test.go`
- `internal/l7/adapter/provider/contract.go`
- `internal/l7/adapter/provider/contract_test.go`
- `internal/l7/adapter/codex/adapter.go`
- `internal/l7/adapter/codex/adapter_test.go`
- `internal/l7/adapter/codex/actual_host_test.go`
- `internal/l7/adapter/claude/adapter.go`
- `internal/l7/adapter/claude/adapter_test.go`
- `internal/l7/adapter/claude/actual_host_test.go`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation.md`

This brief is the only added path. If separately authorized, the only additional
records are the same basename with `-verification.md` and `-audit.md` suffixes.
No README, configuration, dependency, workflow, skill, remote, global-provider,
or unrelated file is in scope. The user-owned untracked foundation audit remains
untouched.

## Pending offline verification

No test or verification record is authorized by the disposition approval. A
fresh approval bound to the commit containing this brief is required before:

1. rechecking that the disposition tree is exactly
   `0676df022d3a2c3ab46b0344213f9e5eff80fc73` and that the failed commits remain
   ancestors;
2. confirming the two failed target versions degrade without a semantic provider
   invocation;
3. running repository-pinned `make verify`, whose actual-host tests compile but
   do not execute;
4. running `go test -race ./internal/l7/... ./cmd/l7` and
   `make cli-cross-build`; and
5. inspecting final scope, diff hygiene, tracked state, and the untouched
   unrelated untracked file before committing the sole verification record.

The disposition requires zero provider, version, help, prompt, stdin, model,
network, retry, fallback, installation, global-configuration, remote, or external
CI invocations. A future provider qualification is a new Tier 3 change with a
fresh candidate, all six separately authorized gates, and no transferred pass.

## Risks and rollback

- The historical Codex pass could be mistaken for support. Every successor
  record must keep it bound to the failed candidate and state `NO_GO` overall.
- Restoring the baseline restores provisional fixture profiles, not an
  actual-host support claim; the lifecycle remains default OFF.
- Reversing this disposition would restore an explicitly failed candidate, not
  a qualified one. It requires a fresh Tier 3 brief and owner approval and must
  not occur through reset or automatic cleanup.
- Unexpected scope, changed ancestry, or a tree other than the exact baseline
  fails closed and requires a revised proposal.

There is no migration, provider installation, remote, global configuration,
deployment, release, or publication to reverse.

## Commit sequence

Completed disposition:

1. `e7eb297efa130659e54b4124127d05a09353483b` —
   `revert(provider): remove failed cli contract coverage`
2. `944b99baa5bb5958b8ba60a0afe9fa915e62b98c` —
   `revert(provider): remove failed claude 2.1.247 remediation`
3. `84eb7f2a7a2ae3b2f9bd865245b293c875858946` —
   `revert(provider): remove failed codex 0.150.1 remediation`
4. `759f539a31e62ea405af593e134185a0696dc6c5` —
   `revert(provider): remove failed parser-discrimination gates`
5. `1db877f071a9f86ae720d7ca57d5a5e1db886072` —
   `revert(provider): retire failed compatibility remediation`
6. Commit this brief and stop for fresh owner approval bound to its exact commit.

After fresh approval, run only the pending offline verification and commit its
single record. Obtain separate authorization for an independent read-only
`l7-release` audit. Merge, release, deployment, publication, provider activity,
external CI, remote creation, and global configuration changes remain
unauthorized.
