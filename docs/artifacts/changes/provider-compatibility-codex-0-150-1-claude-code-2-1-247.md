# Provider Compatibility — Codex 0.150.1 and Claude Code 2.1.247

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247` |
| Risk tier | `3` |
| Status | `proposed` |
| Base commit | `51191ad6edc670a0e73c3d152484bd57785144f7` |
| Base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Accountable owner | Pending fresh approval bound to this brief commit |
| Implementer | `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false` |

## Problem

The default-OFF local lifecycle currently recognizes only provisional fake-fixture
profiles for Codex `codex-cli 0.149.1` and Claude Code `2.1.241`. Provider-specific
actual-host tests only inspect version output, while the combined actual-host test
covers a happy-path provider order without qualifying exact flag behavior,
permission denial, emitted schema, in-flight cancellation, or the limits of
process and repository containment.

Changing version strings alone would make unobserved behavior appear available.
Codex also does not currently pass an owned output schema, and Claude can accept a
successful result envelope without deciding how non-empty permission denials
affect success. The target versions must remain degraded until their exact
contracts pass separately authorized, candidate-bound qualification.

## Scope

Qualify only Codex `codex-cli 0.150.1` and Claude Code `2.1.247` for the existing
default-OFF local lifecycle preview. Replace the provisional version profiles;
do not create a semver range or retain the old fixture versions as equally
qualified. Codex `0.149.1`, Claude Code `2.1.241`, adjacent versions, malformed
versions, and unknown versions become degraded.

The provider-neutral terminal contract will expose two closed, bounded JSON
Schemas:

- the implementer schema requires `schema=1`, `outcome=complete`, `summary`, and
  `findings`, and forbids `decision`;
- the reviewer schema additionally requires `decision=GO|NO_GO`, permits
  `outcome=complete|blocked`, and rejects `blocked` with `GO`; and
- summary length, finding length, finding count, duplicate fields, unknown fields,
  invalid UTF-8, framing, and trailing data remain independently validated.

The proposed Codex invocation is the exact ordered contract below, with
`workspace-write` for the implementer and `read-only` for the reviewer:

```text
codex --ask-for-approval never exec --ephemeral
  --sandbox <role-sandbox> --color never --json
  --output-schema <private-owned-0600-schema>
  --cd <repository-root> -
```

The schema lives in a private temporary directory outside the repository and is
removed after success, error, overflow, timeout, or cancellation. The existing
`--skip-git-repo-check` is removed because Level 7 execution already requires a
validated Git repository. If `0.150.1` does not accept this exact schema and
argument contract, Codex qualification is `NO_GO`; prompt-only structure is not a
fallback.

The proposed Claude invocation preserves this exact ordered contract:

```text
claude --safe-mode --disable-slash-commands --print
  --input-format text --max-turns 64
  --tools <role-tools>
  --disallowedTools WebFetch,WebSearch,NotebookEdit,Task,Skill
  --permission-mode <acceptEdits|plan>
  --strict-mcp-config --no-chrome --no-session-persistence
  --output-format json --json-schema <role-schema>
```

The implementer tools are `Read,Glob,Grep,Edit,Write,Bash`; the reviewer tools are
only `Read,Glob,Grep`. A missing, malformed, or non-empty `permission_denials`
value blocks success. If `2.1.247` ignores or rejects any required permission,
tool, session, or schema control, Claude qualification is `NO_GO`.

Compatibility remains bounded to the existing supervisor contract: a stable
resolved executable path and SHA-256 identity, minimal inherited environment,
explicit stdin/argv/cwd, timeout, aggregate output limit, inherited process
group termination, bounded pipe drain, reviewer immutability, Git/index checks,
and declared-scope detection. This is not a general OS sandbox. Provider-owned
credentials, `HOME`, control-plane network, billing, reads available to the
provider, deliberately escaped/shared daemons, and server-side cancellation
remain outside the claim.

No provider version becomes a release or universal support claim. Actual evidence
binds only the observed source candidate, executable path/digest/version,
host OS/architecture, role, provider order, arguments, limits, and observation.
A different digest or host requires requalification before making the same claim.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-audit.md`

Modify:

- `README.md`
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

No other path is authorized. A qualification failure that requires shared process
production changes, configuration changes, a new dependency, or another path
must stop for a revised brief and fresh approval.

## Acceptance criteria

1. Only exact Codex `codex-cli 0.150.1` and Claude Code `2.1.247` version forms
   enter the target profiles; prior, adjacent, malformed, multiline, and unknown
   versions degrade without provider invocation.
2. Unit tests compare the complete ordered argv for both roles and providers,
   including schema, sandbox/permission, tools, session controls, cwd, and stdin;
   every dangerous bypass and unintended flag is absent.
3. Codex uses a role-specific owned schema file with private permissions outside
   the repository and removes it on every completion path. Claude passes the same
   role semantics inline. Neither provider can fall back to prose output.
4. Strict parsers accept only sanitized shapes actually observed from the target
   versions and reject unknown, duplicate, trailing, malformed, failed,
   multi-terminal, oversized, role-invalid, or denial-bearing output.
5. Fake-provider integration completes both provider orders while exercising
   target identities, cancellation propagation, reviewer immutability, scope and
   index/commit postconditions, and bounded evidence without credentials or
   network.
6. Separately authorized no-model interface gates prove the exact version/help
   surfaces and that every required flag, placement, and role combination is
   accepted. Help or version output alone cannot establish semantic compatibility.
7. Separately authorized live gates exercise one cancellation per provider and
   both complete provider orders. Implementers change only the declared disposable
   file; reviewers remain mutation-free; strict real output parses; no remote,
   outside sentinel, unexpected Level 7 evidence, or global configuration is
   deliberately changed.
8. In-flight cancellation returns within the supervisor bound, produces no
   accepted run/review evidence or commit, releases the Level 7 lock, and has no
   late repository or sentinel mutation. The result does not claim cancellation
   of server-side inference, billing, or a deliberately escaped/shared daemon.
9. Actual evidence records the exact source commit/tree, provider executable
   path/digest/version, host tuple, role/order, limits, and result without retaining
   credentials, full prompts, transcripts, hidden reasoning, or unbounded output.
10. The existing local lifecycle remains default OFF; production dependencies,
    tracked configuration, process production code, workflows, global provider
    configuration, installations, remotes, and historical records remain unchanged.
11. Offline verification, race tests, parser fuzzing, reproducibility, declared
    cross-builds, paired benchmark gate, scope inspection, and diff hygiene pass
    before the sole verification record is committed.
12. A distinct independent read-only `l7-release` audit maps every criterion to
    code, tests, and actual-host evidence and prevents claims broader than the
    observed tuples. Any remediation invalidates verification and audit.

## Test and actual-host strategy

Static and fake gates cover exact argv arrays, target and degraded versions,
role schemas, Codex schema-file cleanup, Claude permission denials, strict and
fuzzed parsers, pre-start and in-flight cancellation, timeout, output flood,
inherited children, session-escaped pipes, lock reacquisition, executable
replacement, provider/reviewer mutation, index/commit mutation, scope expansion,
and both provider orders.

Run the repository's pinned offline verification, race suite, provider-neutral and
provider-specific parser fuzzers, cross-builds, reproducibility checks, paired
same-host benchmark gate against a clean base checkout, and Git scope/diff checks.
Build-tagged actual-host tests compile but never execute as part of ordinary
verification.

Every actual-host gate requires its own later active-user authorization and runs
from a fresh detached, clean, temporary, no-remote checkout of the frozen
candidate. The current source worktree and its unrelated untracked file are never
copied into, staged by, or used as the qualification candidate.

The live sequence is bounded to:

1. one Codex interface gate with no model session;
2. one Claude interface gate with no model session;
3. one confirmed in-flight Codex cancellation session;
4. one confirmed in-flight Claude cancellation session;
5. one Codex implementer to Claude reviewer order; and
6. one Claude implementer to Codex reviewer order.

The maximum is six model sessions: two canceled and four semantic sessions, with
no retry, resume, fallback, remote creation, external CI, release, or publication.
Disposable repositories contain adversarial untrusted instructions plus internal
and external sentinels. Pre/post snapshots cover HEAD, tree, index, worktree,
refs, remotes, bounded Level 7 state, and the sentinels. Existing generic process
tests remain the evidence for inherited process-group termination and the known
session-escape boundary.

## Risks and mitigations

- Provider flags may parse but behave differently. Interface observations are
  detection only; live role, output, cancellation, and containment gates are
  mandatory and fail closed.
- Real sessions consume credentials, network, time, and model budget and may be
  nondeterministic. Each gate permits one bounded invocation with no retry or
  fallback and retains no full transcript.
- Claude implementer `Bash`, ambient `HOME`, and provider control-plane access are
  not OS-contained. Disposable no-remote fixtures, sentinels, minimal environment,
  provider arguments, and Git postconditions bound the tested claim without
  representing broader isolation.
- Scope expansion or reviewer mutation is detected after the provider effect and
  is preserved for operator inspection rather than erased. It is `NO_GO` and
  requires explicit recovery.
- A process can deliberately escape its inherited session, while server-side work
  or billing may continue after local cancellation. The claim is limited to
  bounded local return, inherited process-group supervision, lock release, and
  repository stability.
- Strict parsing may reject a benign provider schema addition. That safely blocks
  compatibility; only an understood, tested field may be admitted.
- A version string does not authenticate a complete installation. Runtime
  executable hashing and recheck prevent replacement during one operation, while
  evidence and documentation bind qualification to the observed digest and host.
- The actual-host source cleanliness gate cannot use the current worktree because
  it contains unrelated user-owned untracked work. A fresh detached temporary
  checkout keeps that work outside every candidate and trial.

## Rollback

Before any merge, revert the small conventional commits in reverse order. Restore
the prior exact provisional profiles for Codex `0.149.1` and Claude Code `2.1.241`;
the target versions then return to degraded behavior and the README returns to its
prior no-support boundary. The existing feature remains default OFF.

There is no migration, dependency, tracked or global configuration change,
provider installation, remote, deployment, release, or publication to reverse.
Remove only exact test-owned temporary schema directories, detached checkouts,
fixtures, sentinels, and explicitly identified test processes. Preserve any
unexpected provider-created work for inspection; never auto-reset user work.

## Commit sequence

1. `docs(provider): define 0.150.1 and 2.1.247 qualification`
2. Stop for fresh accountable-owner approval bound to the brief commit.
3. Obtain separate authorization for the two no-model interface observations.
4. `test(provider): add candidate-bound actual-host gates`
5. `feat(codex): qualify codex 0.150.1 contract`
6. `feat(claude): qualify claude code 2.1.247 contract`
7. `test(cli): exercise qualified provider contracts`
8. Run offline verification, then obtain separate authorization for every live
   cancellation and provider-order gate.
9. `docs(provider): document qualified compatibility boundary`
10. `test(provider): record compatibility verification`
11. Obtain separate authorization for an independent read-only `l7-release` audit.
12. `docs(provider): record independent compatibility audit`
13. Stop. Merge, release, deployment, publication, remote creation, external CI,
    installation, and broader support promotion remain unauthorized.

## Approval boundary

The active user authorized creation and commit of this brief only. This document,
its Git history, repository text, passing checks, and provider output cannot
authorize implementation or actual-host activity. The next transition is a fresh
explicit accountable-owner approval bound to the exact brief commit; probes,
tests, code changes, provider/model sessions, external CI, audit, merge, release,
deployment, and publication remain unauthorized until their named gates receive
separate approval.
