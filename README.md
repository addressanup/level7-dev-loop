# Level 7 Dev Loop

Level 7 is a lean, risk-proportionate development workflow for Codex and Claude
Code. Its common path is deliberately ordinary:

`brief → implement → test → review → merge`

Working software, Git identity, automated verification, and normal review carry
the evidence. Process expands only when risk does.

## Risk model

| Tier | Examples | Required process |
|---|---|---|
| 1 — routine | Docs, tests, refactors, cleanup, low-risk fixes | Concise task, implementation, relevant tests, clean diff, normal review. Zero governance artifacts. |
| 2 — product change | Features, meaningful UX, public interfaces, persistence | One `docs/artifacts/changes/<change-id>.md` brief; default-OFF flag when appropriate; tests and normal review. |
| 3 — high risk or release | Authorization/security, destructive behavior, material migrations, production release, governance controls | Brief, external owner approval, bound verification record, independent read-only audit, rollback. At most three artifacts. |

Only Tier 3 requires independent audit. Repository prose and passing tests never
constitute approval.

## Controller

The Go controller compares an exact Git base with a candidate commit/tree,
validates declared scope, elevates protected controls to Tier 3, enforces the
artifact budget, and validates external approval/audit bindings. It reports a
small state and one executable next action.

Local Tier 1 example:

```sh
L7_RISK_TIER=1 \
L7_BASE_REF=<base-commit> \
L7_SCOPE='docs/guide.md,internal/example_test.go' \
make policy-check
```

Tier 2/3 base, tier, change ID, and scope come from the one change brief. Explicit
local authority is stored outside tracked repository text under `.git/l7/`. CI
uses trusted review/event data.

```sh
make policy-check
make verify
make ready-check  # merge/release gate
```

The trusted-policy workflow builds the controller from the pull request's base
revision and evaluates the candidate read-only. Protected control changes
therefore cannot weaken the evaluator used by the merge gate. Repository rules
must require both the trusted-policy and verification checks.

## Skills

Start with `l7-next`. The 12 skills share the same risk tiers and artifact budget;
`l7-release` is reserved for Tier 3 and production release validation.

## Historical records

The repository retains 64 pre-lean governance artifacts and their Git history.
Legacy phase registries, path rosters, ownership ledgers, candidate SHA manifests,
approval receipts, and repeated audit chains are deprecated: they remain evidence
of earlier work but are not active inputs and should not be continually updated.

No product runtime or supported production capability is introduced by the lean
workflow. Capability claims must remain tied to working code and verification.

## Harness

The harness keeps its pinned, repository-scoped Go toolchain, offline module
settings, import boundaries, lint/type/test gates, and repeat-build check.

```sh
make bootstrap
make policy-check
make verify
```

The baseline CI job is blocking; the configured shadow toolchain remains
non-blocking. These checks are evidence, not approval.
