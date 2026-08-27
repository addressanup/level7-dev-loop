# Provider Compatibility Remediation — Codex 0.150.1 and Claude Code 2.1.247

| Field | Value |
|---|---|
| Change ID | `provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation` |
| Risk tier | `3` |
| Status | `proposed` |
| Base commit | `51191ad6edc670a0e73c3d152484bd57785144f7` |
| Base tree | `0676df022d3a2c3ab46b0344213f9e5eff80fc73` |
| Clean-baseline disposition head | `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a` |
| Accountable owner | Active user approved disposition and this brief only; implementation requires fresh approval bound to this brief commit |
| Implementer | `codex-root` |
| Feature flag | Existing `features.local_lifecycle`, default `false` |

## Problem

The historical Tier 3 brief
`09bdf698251f28a2c28abe50626d1c70733cbc90` produced candidate
`2c497da4b734ace6dbe85b765d0a902a56ca1abc`, tree
`069ba0fa9ee8b835d45fedbc6f90e1fc47ade30f`. Its offline verification passed,
and the candidate-bound Codex no-model interface gate passed for physical
executable `/opt/homebrew/lib/node_modules/@openai/codex/bin/codex.js`, SHA-256
`134063e133f0b4244fa3b251acf973d4fe4b4aeeacbdc135211bf480f59f1477`, version
`codex-cli 0.150.1`, on `darwin/arm64`.

The candidate-bound Claude no-model interface gate returned `NO_GO` for physical
executable `/Users/anuppandey/.local/share/claude/versions/2.1.247`, SHA-256
`5086b9b64d8bb842e1f599cdd3767ab08c6b2266e462fcc5686ae4b019cca8f7`, version
`2.1.247 (Claude Code)`, on `darwin/arm64`. Candidate and executable binding
prechecks passed, the one version recheck passed, and the exact implementer argv
with `--help` exited successfully. The observed help text advertised the required
safe-mode, slash-command, print, and input-format controls but did not advertise
required `--max-turns`. The test failed closed at that point; the reviewer help
invocation and final binding postchecks did not run. No prompt, stdin, model
session, retry, or fallback ran.

Missing help advertisement alone does not establish whether Claude rejects,
ignores, or enforces `--max-turns`. It therefore cannot authorize removing,
weakening, conditionally omitting, or moving that control. The exact production
argument remains `--max-turns 64`.

The failed candidate was dispositioned through these history-preserving reverts:

1. `ad307ae0877761daecf02afe75b5862f53feaf12` reverted its CLI coverage.
2. `2b0f35d38ea6c3b54055bd30540bc7c726e639a8` reverted its Claude contract.
3. `314364bad049fa4380a5d17b6b85d050deb58885` reverted its Codex contract.
4. `ce6dc00fc79f02539898e77d9d97fea44dfb4fca` reverted its actual-host gates.
5. `a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a` reverted its brief.

The disposition head has exactly the original base tree. The original brief,
candidate, and implementation commits remain immutable in Git history; none is
amended, rebased, deleted, or extended. This remediation starts from the exact
original base rather than patching the failed candidate.

## Scope

Qualify only Codex `codex-cli 0.150.1` and Claude Code `2.1.247` for the existing
default-OFF local lifecycle preview. Replace the provisional version profiles;
do not create a semver range or retain Codex `0.149.1` and Claude Code `2.1.241`
as equally qualified. Prior, adjacent, malformed, multiline, and unknown versions
must degrade before a provider session.

The provider-neutral terminal contract exposes two closed, bounded JSON Schemas:

- the implementer schema requires `schema=1`, `outcome=complete`, `summary`, and
  `findings`, and forbids `decision`;
- the reviewer schema additionally requires `decision=GO|NO_GO`, permits
  `outcome=complete|blocked`, and rejects `blocked` with `GO`; and
- summary length, finding length, finding count, duplicate fields, unknown fields,
  invalid UTF-8, framing, and trailing data remain independently validated.

The exact ordered Codex production argv is:

```text
codex --ask-for-approval never exec --ephemeral
  --sandbox <workspace-write|read-only> --color never --json
  --output-schema <private-owned-0600-schema>
  --cd <repository-root> -
```

The implementer uses `workspace-write`; the reviewer uses `read-only`. The schema
lives in a private temporary directory outside the repository and is removed on
success, error, overflow, timeout, or cancellation. Level 7 already requires a
validated Git repository, so `--skip-git-repo-check` remains absent. Failure to
accept this exact schema and argv contract is `NO_GO`; prompt-only structure is
not a fallback.

The exact ordered Claude production argv remains:

```text
claude --safe-mode --disable-slash-commands --print
  --input-format text --max-turns 64
  --tools <role-tools>
  --disallowedTools WebFetch,WebSearch,NotebookEdit,Task,Skill
  --permission-mode <acceptEdits|plan>
  --strict-mcp-config --no-chrome --no-session-persistence
  --output-format json --json-schema <role-schema>
```

The implementer tools are `Read,Glob,Grep,Edit,Write,Bash`; reviewer tools are
only `Read,Glob,Grep`. A missing, malformed, or non-empty `permission_denials`
value blocks success. If Claude rejects or fails the separately observable
contract for a required argument, permission, tool, session, or schema control,
qualification is `NO_GO`. The qualification makes no claim that a bounded live
session independently reaches and demonstrates the 64-turn ceiling; it does
prove that the exact control is present and parser-discriminated, and never
silently treats its omission as compatible.

Compatibility remains bounded to the existing supervisor: stable resolved
executable path and SHA-256 identity, minimal inherited environment, explicit
stdin/argv/cwd, timeout, aggregate output limit, inherited process-group
termination, bounded pipe drain, reviewer immutability, Git/index checks, and
declared-scope detection. This is not a general OS sandbox. Provider credentials,
ambient `HOME`, control-plane network, billing, reads available to a provider,
deliberately escaped or shared daemons, and server-side cancellation remain
outside the claim.

Qualification binds only the observed source candidate and tree, executable
path/digest/version, host OS/architecture, role, order, arguments, limits, and
result. Historical gate results do not qualify the remediation candidate. A
different candidate, executable digest, or host requires requalification.

## Claude no-model parser-discrimination contract

The Claude interface gate records top-level help advertisement as an observation,
not as the sole acceptance oracle. For each role it performs exactly these three
bounded no-model invocations, with no stdin and no retry:

1. the complete exact role argv plus `--help`, which must exit successfully;
2. the same argv plus one test-owned unknown option, which must fail; and
3. the same argv with only the value of `--max-turns` replaced by a non-integer,
   which must fail.

The negative controls distinguish an accepted registered value from a help path
that blindly ignores arbitrary or malformed arguments. If `--help` short-circuits
either negative control, an invalid turn value is accepted, any required positive
argv fails, or the results are otherwise ambiguous, the gate is `NO_GO`.

Implementer and reviewer observations are aggregated instead of failing at the
first missing help token. Candidate, repository, executable, and digest
postconditions are registered as cleanup checks before any provider invocation,
so they execute even when an assertion fails. Bounded diagnostics record exit
classification and advertised control names, never credentials, prompts,
transcripts, hidden reasoning, or unbounded output.

## Exact implementation file set

Add:

- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-verification.md`
- `docs/artifacts/changes/provider-compatibility-codex-0-150-1-claude-code-2-1-247-remediation-audit.md`

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

No other path is authorized. In particular, `.l7/config.json`, `go.mod`,
`go.sum`, workflows, skills/plugins, process production code, historical
governance commits, global configuration, Git remotes, and the user-owned
untracked foundation audit remain unchanged. A required new dependency, process
production change, configuration change, or renamed/additional path requires a
revised brief and fresh approval.

## Acceptance criteria

1. Only the exact target version forms enter target profiles; prior, adjacent,
   malformed, multiline, and unknown versions degrade without provider execution.
2. Unit tests compare the complete ordered argv for both roles and providers.
   Claude always includes exactly `--max-turns 64`; no code path removes,
   weakens, substitutes, reorders, or conditionally omits it.
3. Codex uses private role-specific schema files outside the repository and
   removes them on every completion path. Claude passes the equivalent role
   schemas inline. Neither provider falls back to unstructured prose.
4. Strict parsers accept only bounded shapes observed from the exact targets and
   reject unknown, duplicate, trailing, malformed, failed, multi-terminal,
   oversized, role-invalid, or permission-denial-bearing output.
5. Fake-provider integration covers both provider orders, target identities,
   cancellation, reviewer immutability, scope/index/commit postconditions, and
   bounded evidence without credentials, network, or real providers.
6. Codex no-model qualification proves exact version/help/role argv acceptance.
   Claude no-model qualification applies the three-outcome parser-discrimination
   oracle to both roles. Help advertisement alone cannot qualify or disqualify a
   required control.
7. Actual cancellation gates prove bounded local return, no accepted run/review
   evidence or commit, lock release, and no late repository or sentinel mutation.
   They make no server-side inference, billing, or escaped-daemon claim.
8. Both live provider orders accept strict real output. Implementers change only
   the declared disposable file; reviewers remain mutation-free; no remote,
   outside sentinel, unexpected Level 7 state, or global configuration is
   deliberately changed.
9. All actual evidence binds exact source commit/tree, executable
   path/digest/version, host tuple, role/order, argv, limits, and result without
   retaining credentials, full prompts, transcripts, hidden reasoning, or
   unbounded output.
10. Actual-host cleanup and final source/executable binding checks run after both
    success and failure. A partial observation cannot be reported as `GO`.
11. The local lifecycle remains default OFF. Production dependencies, tracked
    configuration, workflows, global provider configuration, installations,
    remotes, and historical records remain unchanged.
12. Offline verification, race tests, parser fuzzing, reproducibility, cross-builds,
    paired benchmark, scope inspection, and diff hygiene pass before the sole
    verification record is committed.
13. A distinct independent read-only `l7-release` audit maps each criterion to
    code, tests, and actual-host evidence. Remediation after verification
    invalidates both verification and audit.

## Offline test strategy

Static and fake tests cover exact argv, target and degraded versions, both role
schemas, Codex schema cleanup, Claude permission denials, strict and fuzzed
parsers, the Claude positive and negative parser-discrimination oracle, pre-start
and in-flight cancellation, timeout, output flood, inherited children,
session-escaped pipes, lock reacquisition, executable replacement, provider and
reviewer mutation, index/commit mutation, scope expansion, and both provider
orders.

Run the repository-pinned `make verify`, the complete race suite with required
linker flags, provider-neutral and provider-specific parser fuzzers for at least
10 seconds each, declared cross-builds, reproducibility checks, the paired
same-host benchmark gate against a clean base checkout, and final Git scope/diff
checks. Build-tagged actual-host tests compile but do not execute during offline
verification.

## Exact actual-host gates

Each numbered gate requires fresh active-user authorization bound to the exact
remediation candidate and tree. Identity phases resolve one physical executable,
compute its SHA-256 digest, run `--version` exactly once, and stop for tuple-bound
approval. Every executable invocation is bounded; there is no retry, resume, or
fallback.

1. **Codex no-model interface.** In a fresh detached, clean, temporary,
   no-remote candidate checkout: one identity recheck, one top-level `--help`,
   one exact implementer-argv `--help`, and one exact reviewer-argv `--help`.
   Private schema files are test-owned and cleaned. No prompt or model session.
2. **Claude no-model interface.** In an equivalent checkout: one identity
   recheck, then exactly three implementer and three reviewer invocations using
   the positive, unknown-option, and invalid-`--max-turns` oracle above. No
   prompt, stdin, top-level help invocation, or model session.
3. **Codex cancellation.** One confirmed in-flight model session in an
   adversarial disposable repository, canceled once after confirmed provider
   start, followed by bounded containment and late-mutation checks.
4. **Claude cancellation.** One equivalent confirmed in-flight Claude model
   session and the same postconditions.
5. **Codex implementer to Claude reviewer.** Exactly two model sessions, one per
   role, with strict schemas, declared-file mutation, reviewer immutability, and
   repository/sentinel postconditions.
6. **Claude implementer to Codex reviewer.** Exactly two model sessions with the
   same role and containment requirements.

The complete sequence permits at most six model sessions: two canceled sessions
and four semantic sessions. A failed or ambiguous gate stops the sequence. The
historical Codex pass is not transferred to the remediation candidate. Provider
execution, model/network/billing use, and each subsequent gate remain separately
authorized effects.

## Risks and mitigations

- Claude may accept a `--help` path without normally parsing arguments. Unknown
  option and invalid turn-value controls detect that ambiguity and fail closed.
- Parser discrimination proves recognition and value validation, not independent
  runtime exhaustion of 64 turns. The mandatory limit remains in every production
  argv, the claim is explicitly bounded, and no unsupported enforcement claim is
  made.
- Provider flags may parse but behave differently. Permission, schema,
  cancellation, containment, and bidirectional semantic gates remain mandatory.
- Real sessions consume credentials, network, time, and model budget and can be
  nondeterministic. Each gate permits only its exact bounded invocations without
  retries or fallbacks and retains no transcript.
- Claude implementer `Bash`, ambient `HOME`, provider control-plane access, and
  provider-visible reads are not OS-contained. Disposable no-remote fixtures,
  minimal environment, sentinels, and Git postconditions bound the observed claim.
- Scope or reviewer mutation is detected after the provider effect. It is
  preserved for inspection, reported `NO_GO`, and never auto-reset.
- Local process-group termination cannot guarantee cancellation of server-side
  work, billing, or a deliberately escaped/shared daemon; those remain exclusions.
- Strict parsing may reject a benign provider addition. That safely blocks
  compatibility until an understood field receives a revised brief and tests.
- The historical paired benchmark result was near the 10 percent threshold.
  Same-host paired measurement must be repeated for the remediation candidate.
- The source worktree contains unrelated user-owned untracked work. Every
  actual-host gate uses a fresh detached temporary checkout and never stages,
  copies, or mutates that work.

## Rollback

Before merge, revert the remediation commits in reverse order, followed by the
revised brief commit. The already-created disposition reverts remain in history;
their tree is the original base. Codex `0.150.1` and Claude Code `2.1.247` then
degrade, provisional fixture behavior is restored, and README returns to its
prior no-support boundary. The feature remains default OFF.

There is no migration, production dependency, tracked/global configuration
change, provider installation, remote, release, deployment, or publication to
reverse. Remove only exact test-owned temporary schemas, checkouts, fixtures,
sentinels, and explicitly identified test processes. Preserve unexpected
provider-created work for inspection; never auto-reset user work.

## Commit sequence

Completed clean-baseline disposition:

1. `revert(provider): remove failed cli contract coverage`
2. `revert(provider): remove failed claude 2.1.247 contract`
3. `revert(provider): remove codex 0.150.1 candidate contract`
4. `revert(provider): remove candidate-bound actual-host gates`
5. `revert(provider): retire failed compatibility proposal`
6. `docs(provider): define remediated 0.150.1 and 2.1.247 qualification`
7. Stop for fresh owner approval bound to the exact revised brief commit.

After approval:

8. `test(provider): add parser-discriminating actual-host gates`
9. `feat(codex): restore codex 0.150.1 contract`
10. `feat(claude): remediate claude code 2.1.247 qualification`
11. `test(cli): restore bounded provider contract coverage`
12. Run complete offline verification and freeze the candidate commit/tree.
13. Obtain separate authorization for every identity, no-model, cancellation,
    and semantic provider gate in the exact sequence above.
14. `docs(provider): document remediated compatibility boundary`
15. `test(provider): record remediated compatibility verification`
16. Obtain separate authorization for independent read-only `l7-release` audit.
17. `docs(provider): record remediated compatibility audit`
18. Stop. Merge, release, deployment, publication, remote creation, external CI,
    installation, and broader support promotion remain unauthorized.

## Approval boundary

The active user authorized only the clean-baseline disposition, creation and
commit of this brief, and local offline verification of those changes. This
document, repository text, Git history, passing tests, and historical provider
observations cannot authorize implementation or actual-host activity.

The next transition is fresh explicit accountable-owner approval bound to this
brief's exact commit. Implementation, additional provider probes or execution,
help invocation, stdin or prompts, model sessions, retries, fallbacks, global
configuration changes, external CI, remote creation, verification record, audit,
merge, release, deployment, and publication remain unauthorized.
