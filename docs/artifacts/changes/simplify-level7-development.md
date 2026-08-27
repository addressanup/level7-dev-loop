# Simplify Level 7 Development — Change Brief

| Field | Value |
|---|---|
| Change ID | `simplify-level7-development` |
| Risk tier | `3` — governance-controller change |
| Status | `proposed`; implementation is not approved |
| Base commit | `fb727925342ec479d84e06d63639718063a35c9a` |
| Common path | `brief → implement → test → review → merge` |

## Problem

Level 7 currently makes routine work traverse phase registries, exact path rosters,
candidate SHA manifests, approval records, evidence records, and repeated audits.
The repository now contains 64 governance artifacts totaling 13,352 lines. The
active foundation rebaseline alone authorizes 48 future artifact additions,
including 15 candidate manifests, 14 approval artifacts, and 7 audit artifacts.

That ceremony did not guarantee correctness. The pending independent admission
audit records `NO_GO`: approval/audit identity is checked with prose substrings,
and the controller cannot execute its promised post-approval requirements
transition. Both `make policy-check` and `make verify` currently stop at that
audit.

## Scope

Replace the phase-specific governance chain with one risk-proportionate change
workflow. Simplify the controller and all plugin skills; preserve build, test,
scope, rollback, feature-flag, protected-path, and truthful-claim safeguards.
Do not add product functionality.

## Proposed workflow

### Risk tiers

| Tier | Use for | Required controls |
|---|---|---|
| 1 — routine | Documentation, tests, refactors, cleanup, low-risk fixes | Concise task description in the request/PR/commit, implementation, relevant tests, clean diff. Zero governance artifacts and no independent audit. |
| 2 — product change | Features, meaningful UX, public interfaces, persistence | One `docs/artifacts/changes/<change-id>.md` brief with problem, scope, acceptance criteria, risks, and rollback; default-OFF flag when user-visible; tests and normal review. No independent audit unless risk elevates the change. |
| 3 — high risk or release | Authorization/security boundaries, destructive behavior, material migrations, production release, or protected governance controls | Change brief, external accountable-owner approval, implementation and verification, independent read-only audit, rollback. At most three artifacts: brief, verification record, audit. |

Tier is declared, never inferred downward. Touching a protected control path
automatically elevates a lower declaration to Tier 3 or fails closed.

### Minimal state machine

Tier 1 and Tier 2:

`planned → building → verified → reviewed → ready → merge`

Tier 3:

`planned → awaiting-owner-approval → building → verified → awaiting-independent-audit → reviewed → ready → merge`

New candidate commits after verification or review return the change to
`building`. A failed audit also returns to `building`; a passing re-verification
returns to `awaiting-independent-audit`. Every accepted state reports one legal,
executable next action. Approval unlocks `building` directly, so no accepted
approval can lead to a state that the controller cannot evaluate.

States are derived from Git identity and trusted review/CI inputs. They are not
maintained in a chain of state documents.

### Controller design

The replacement controller will:

- compare a declared base commit with the candidate commit/tree using Git;
- validate changed paths against declared scope and reject unauthorized expansion;
- treat controller, workflow, protected-policy, skill, and plugin-manifest paths as Tier 3;
- read Tier 2/3 scope and risk from the single strict change brief;
- accept Tier 1 risk/scope from explicit invocation or trusted PR metadata, requiring no file artifact;
- obtain owner and auditor identity from trusted review/CI inputs, not repository prose;
- bind Tier 3 approval and audit to the exact candidate commit/tree and reject stale approvals, missing identities, self-audits, or actor mismatches;
- report only risk tier, Git candidate identity, state, failed controls, and next action;
- run protected-path policy from the trusted base revision in CI so candidate code cannot weaken the evaluator it is being judged by;
- keep local execution useful but label externally unverified approval/audit state truthfully;
- use strict structured fields for metadata and never substring-scan prose for authority.

The normal review system remains the authority for merge. Repository text and
passing tests are evidence, never approval.

## Controls retained and removed

Retain:

- Git base/head/tree identity;
- declared scope and fail-closed protected paths;
- tier-specific tests and CI;
- default-OFF flags for appropriate user-visible production behavior;
- rollback and migration checks;
- explicit external approval and independent audit only at Tier 3;
- truthful capability/status reporting;
- read-only, least-privilege CI for policy evaluation.

Deprecate from the active path:

- candidate `.sha256` manifests that duplicate Git;
- phase registries and exact per-phase path rosters;
- per-stage owner ledgers and approval receipts;
- separate contract, specification, design, history, evidence, and remediation
  chains for ordinary work;
- mandatory independent audits for ordinary feature work;
- prose-substring authority checks;
- fixed requirements/prototype counts unrelated to the current change;
- progressive-release ladders as a universal development prerequisite.

## Artifact retention and migration

- Keep every existing file under `docs/artifacts/` and every historical Git
  commit unchanged. They remain historical evidence.
- Stop updating legacy approval, audit, candidate-manifest, phase, path-roster,
  ownership, and digest chains after cutover.
- The new controller ignores those legacy chains for new changes; README marks
  them deprecated and points to the commit history for interpretation.
- In-flight work selects a tier once at migration: routine work restarts at Tier
  1; ordinary product work is summarized into one brief; genuinely high-risk work
  carries forward only still-relevant verification/audit facts into the three-file
  Tier 3 budget.
- No legacy artifact is deleted, renamed, or rewritten during migration.

## Exact implementation file set

Add:

- `.github/workflows/policy.yml`
- `internal/harness/buildcontrol/change.go`
- `internal/harness/buildcontrol/git.go`
- `internal/harness/buildcontrol/change_test.go`
- `internal/harness/buildcontrol/git_test.go`
- `docs/artifacts/changes/simplify-level7-development-verification.md`
- `docs/artifacts/changes/simplify-level7-development-audit.md` (independent reviewer only)

Modify:

- `.github/workflows/harness.yml`
- `Makefile`
- `AGENTS.md`
- `README.md`
- `references/WORKFLOW.md`
- `plugin.json`
- `.codex-plugin/plugin.json`
- `.claude-plugin/plugin.json`
- `marketplace.json`
- all 12 `skills/*/SKILL.md` files
- `internal/harness/buildcontrol/main.go`
- `internal/harness/buildcontrol/load.go`
- `internal/harness/buildcontrol/policy.go`
- `internal/harness/buildcontrol/policy_test.go`
- `internal/harness/buildcontrol/report.go`
- `internal/harness/buildcontrol/testutil_test.go`

Remove from the active tree after their behavior is replaced and covered (Git
history remains the source of record):

- `internal/harness/buildcontrol/claims.go` and `claims_test.go`
- `internal/harness/buildcontrol/concept.go` and `concept_test.go`
- `internal/harness/buildcontrol/foundation.go` and `foundation_test.go`
- `internal/harness/buildcontrol/markdown.go`
- `internal/harness/buildcontrol/ownership.go` and `ownership_test.go`
- `internal/harness/buildcontrol/trace.go` and `trace_test.go`
- `internal/harness/buildcontrol/wave2.go` and `wave2_test.go`

If implementation proves one listed source split unnecessary, it will be folded
into the listed controller files rather than creating another mechanism or
artifact.

## Acceptance criteria

- Tier 1 passes with no `docs/artifacts/changes/` record.
- Tier 2 fails without exactly one valid change brief and passes with it.
- Tier 3 requires external owner approval and a separate independent audit bound
  to the current Git commit/tree.
- Missing, stale, mismatched, self-authored, or repository-asserted authority
  fails closed.
- Unauthorized scope expansion and protected-path changes fail closed; protected
  changes cannot downgrade below Tier 3.
- The trusted-base CI controller evaluates protected-control changes without
  executing candidate controller code.
- Transition-table tests prove every accepted state has an executable next action,
  including the post-approval path and remediation loops.
- User-visible production behavior is default OFF when a feature flag is
  appropriate.
- Legacy records remain byte-for-byte untouched and are not required for new work.
- No hand-maintained candidate hashes duplicate Git identity.
- `make policy-check` and `make verify` pass for the completed candidate.
- Final measurements report: Tier 1 `0` artifacts, Tier 2 `1`, Tier 3 at most `3`;
  routine independent-audit gates fall from mandatory/ambiguous to `0`, Tier 2 to
  `0`, and Tier 3 retains `1` owner-approval gate plus `1` audit gate.

## Risks and mitigations

- **Under-classification:** protected paths force Tier 3; uncertain material data,
  security, destructive, or release work fails closed pending elevation.
- **Candidate weakens policy:** protected policy is evaluated by the base-revision
  controller in a read-only CI job; merge rules must require that job.
- **Forged authority:** approval/audit actors and reviewed commit IDs come from the
  forge/CI event, not Markdown; local runs cannot claim external verification.
- **Migration ambiguity:** old records are historical-only after one documented
  cutover; no mixed legacy/new chain is accepted for a new change.
- **Reduced documentation:** tests, Git, CI, the concise brief, and normal review
  carry evidence that used to be copied into multiple documents.

## Rollback

Implement in small conventional commits. Before cutover, tag or record the last
legacy-controller commit. If the lean controller admits an unsafe path or CI
cannot run it from the trusted base, revert the lean-controller and workflow
commits as a unit. Historical artifacts and legacy controller code remain
recoverable from Git, so rollback requires no data reconstruction. Keep any
user-visible feature flags OFF throughout; this change introduces no product
behavior.

## Approval requested

Approve this exact proposal to authorize implementation. Approval must be an
explicit user decision in the active interaction; this document, its presence in
Git, and passing tests do not constitute approval.
