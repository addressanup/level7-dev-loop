# Standalone CLI v1 — Wave 2 Local Lifecycle

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-2` |
| Risk tier | `2` |
| Status | `approved` |
| Base commit | `b91d468f3f6fa56d8c9874c8ddc89b495494766b` |
| Feature flag | `features.local_lifecycle`, default `false` |

## Problem

The Wave 1 CLI is intentionally inert. A solo product founder still cannot adopt
a repository, create a proportionate change brief, recover local workflow context
after interruption, or inspect a change whose identity and scope come from Git.

## Scope

Implement the approved Tier 2 Wave 2 slice of `CLI-001`–`CLI-003`:

- add non-destructive `adopt`, concise `brief`, and read-only Git-derived
  `status` commands with deterministic text and JSON output;
- add a pure, deadlock-free lifecycle contract for all three risk tiers;
- add strict, bounded configuration and transient-state formats;
- use atomic local writes and a repository mutation lock;
- reconstruct accepted facts from Git, the brief, and minimal local state; and
- keep the new user-visible lifecycle default OFF.

This wave does not invoke agents or verification commands, create Git commits,
record authority, perform review, recommend readiness, merge, deploy, use the
network, or change protected governance controls. Runtime status cannot claim a
post-build state until a later wave supplies qualifying evidence.

### Command contract

- `l7 adopt [--enable-local-lifecycle] [--json]` adopts an existing non-bare Git
  repository with an initial commit. It creates configuration atomically and is
  idempotent only when an existing configuration is valid and equivalent.
- `l7 brief --id <id> --tier <1|2|3> --problem <text> --scope <path> ...`
  accepts repeatable acceptance, risk, rollback, and scope values. Tier 1 stores
  only disposable active context; Tier 2/3 creates the single permitted tracked
  change brief and refuses an existing destination.
- `l7 status [--json]` is read-only and reports repository identity, declared and
  changed scope, state, outcome, and one next action. Unsupported downstream
  transitions remain explicitly blocked rather than inferred.

## Exact implementation file set

- `.l7/config.json`
- `docs/artifacts/changes/standalone-cli-v1-wave-2.md`
- `cmd/l7/main.go`
- `cmd/l7/main_test.go`
- `internal/l7/domain/lifecycle.go`
- `internal/l7/domain/lifecycle_test.go`
- `internal/l7/domain/result.go`
- `internal/l7/domain/result_test.go`
- `internal/l7/app/lifecycle.go`
- `internal/l7/app/lifecycle_test.go`
- `internal/l7/app/app.go`
- `internal/l7/app/app_test.go`
- `internal/l7/adapter/localfile/file.go`
- `internal/l7/adapter/localfile/file_test.go`
- `internal/l7/adapter/config/config.go`
- `internal/l7/adapter/config/config_test.go`
- `internal/l7/adapter/config/brief.go`
- `internal/l7/adapter/config/brief_test.go`
- `internal/l7/adapter/git/repository.go`
- `internal/l7/adapter/git/repository_test.go`
- `internal/l7/adapter/state/store.go`
- `internal/l7/adapter/state/store_test.go`
- `internal/l7/presentation/output.go`
- `internal/l7/presentation/output_test.go`
- `README.md`

No Makefile, workflow, harness, controller, skill, plugin manifest, or agent
instruction file is in scope.

## Acceptance criteria

1. Tier 1 creates zero tracked governance artifacts; Tier 2/3 creates exactly one
   concise brief at `docs/artifacts/changes/<change-id>.md`.
2. Git commits and trees are the only candidate identity. Status deterministically
   reports root, common directory, base, head, tree, changed paths, tier, scope,
   state, outcome, and one actionable next step.
3. Scope expansion, stale ancestry, conflicting state, unsafe paths, corrupt or
   partial files, concurrent mutation, and repository identity drift fail closed.
4. Strict JSON rejects unknown or duplicate fields, trailing values, oversized
   input, symlinks, and non-regular files. Local mutations use atomic replacement
   and leave no accepted partial state after interruption.
5. Every valid lifecycle state has a next transition. Failed or stale assurance
   returns to `building`; tests and repository prose never become approval.
6. Wave 2 performs no agent, network, verification, review, commit, merge, or
   deployment effect and makes no claim that those capabilities exist.
7. Unit, application, temporary-Git integration, interruption, and adversarial
   tests pass. A 10,000-path status benchmark records the initial performance
   baseline without creating another governance artifact.
8. Production code adds no third-party dependencies. `make policy-check`,
   `make verify`, and `make cli-cross-build` pass, followed by normal review.

## Risks and mitigations

- **Filesystem substitution or corruption:** reject unsafe file types and links,
  parse bounded strict JSON, lock mutations, and use same-directory atomic writes.
- **Git drift or monorepo cost:** take and recheck explicit Git snapshots, parse
  bounded NUL-delimited output, avoid repository content scans, and benchmark a
  large changed-path fixture.
- **Scope or authority confusion:** normalize explicit relative paths, fail closed
  on expansion, keep Git canonical, and never read approval from prose or tests.
- **Incomplete product journey:** keep the feature default OFF and report later
  verification, review, readiness, and merge capabilities as unavailable.
- **State duplication:** store only an active selector and Tier 1 task context;
  reconstruct identity and changed scope from Git and the tracked brief.

## Rollback

Revert the small conventional commits in reverse order. In an adopted target
repository, rollback removes only the exact Level 7-created `.l7/config.json` and
`$(git rev-parse --git-common-dir)/l7/product/` paths after validating them; Wave
2 does not automate deletion. There is no database migration, external service,
dependency, remote state, Git-history rewrite, or deployment to reverse.
