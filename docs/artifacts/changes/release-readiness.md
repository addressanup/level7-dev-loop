# Stable Release Readiness — Change Brief

| Field | Value |
|---|---|
| Change ID | `release-readiness` |
| Risk tier | `3` — release controls, package qualification, and publication boundary |
| Status | Active user authorized repository-local release preparation on 2026-10-10; release `NO_GO` until external gates pass |
| Base commit | `e17aa1797099e3fed0e4054d52ce8718ed3b61a7` |
| Base tree | `922d517cbeb847b9525d27a8175b30ffe262c62a` |
| Intended release | `1.0.0` / `stable`, tag `v1.0.0`; freeze a new exact candidate after preparation |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Assurance | `solo` — truthful self-review and Git/CI evidence; no independent code audit claimed |
| Current authority | Implement and verify the bounded repository-local scope; no workflow dispatch, signing, provider calls, protected merge, or publication |

## Problem

Current `main` passes engineering checks but is not a qualified stable release.
All six foundation steps have evidence in `docs/foundation/README.md`; no
foundation status changes are needed for this brief.

| Evidence or gate | Current disposition |
|---|---|
| Post-merge Harness run `38023004666` at the base above | PASS: Linux baseline/shadow and both macOS jobs; paired benchmark skipped on push by design |
| Live outcome run `20261010T034830Z-live` | 48/48 valid, no contamination; both arms 24/24 correct; zero false successes or safety violations |
| Phase 4 gate / improvement claim | `READY_FOR_BRIEF` / `INCONCLUSIVE`; 24 tied pairs, exact McNemar p = 1.0000 |
| Observed overhead | Crew median 3m4s and 70 turns; plain median 55s and 24 turns |
| Stable release preflight at the base | BLOCKED: fixed base `5c23038…`, two-parent merge requirement, and historical team identities do not match current squash/solo development |
| Exact stable assets and host/provider qualification | NOT_RUN for this candidate; no stable release workflow run was found |
| Release evaluation | BLOCKED: public corpus has no protected holdout; `L7-EVAL-007` remains unmet |
| Existing publication | Immutable unsigned, unnotarized, evaluation-only `v1.0.0-dev` at `245d514…`; it predates the crew changes and must not be replaced |

The public run is formative, Codex-only, and covers eight small Python tasks.
Twenty-four pairs detect only large differences, and prompt/process effects
are not separated. It establishes neither equivalence nor an improvement
claim, release-quality evaluation, or permission to auto-merge.

## Scope

Prepare a stable release of the existing product, including opt-in crew
Phases 1–3. Keep `features.crew` and `features.crew_pr` default OFF. Phase 4,
deployment, dependency changes, and evaluator/corpus/threshold changes are
outside this change.

Proposed repository-local preparation:

- Reconcile both stable workflow jobs with exact squash-merge lineage and solo
  assurance. Resolve the real forge owner/operator at preparation time; never
  fabricate a second reviewer or weaken hosted protection silently. Any
  enforced external separation-of-duties rule remains a blocker until the
  owner explicitly resolves it.
- Preserve manual-only dispatch, exact commit/tree/check binding, required
  PR benchmark and trusted policy checks, protected signing/publication
  boundaries, immutable assets, and refusal of existing tags/releases.
- Add offline preflight regression checks and wire them into the harness.
- Refresh stable notes, README, and changelog for the 17 skills, opt-in crew,
  owner-confirmed PR merging, measured limitations, and qualified host scope.
  Preserve the frozen v0.1.1 payload and all historical governance records.

Owner-controlled gates follow repository preparation: supply a protected
release holdout, authorize its evaluation, configure/verify hosted controls,
authorize signing/notarization and exact-asset provider trials, then authorize
publication against the final assets. None is authorized by this document.

## Exact implementation file set

These are the bounded paths for the authorized repository-local preparation;
expand them only through a separate scoped decision. The shared read-only
preflight script also runs its regression tests offline.

Add:

- `docs/artifacts/changes/release-readiness.md`
- `scripts/harness/check-stable-release-preflight.sh`

Modify:

- `.github/workflows/release.yml`
- `Makefile`
- `README.md`
- `CHANGELOG.md`
- `docs/releases/v1.0.0.md`

Delete: none. External holdout material, credentials, transcripts, and prepared
assets are not tracked implementation files.

## Acceptance criteria

1. This preparation changes only the declared paths. Preserve the user's untracked
   foundation audit, current product behavior, tags, releases,
   credentials, and hosted settings. Report readiness as `NO_GO` while any
   release gate remains unverified.
2. Later preflight tests accept an eligible exact squash-merged candidate and
   reject changed main/head/tree/checks, missing authority, protected-control
   drift, reused artifacts, and an existing tag or release. Apply the same
   invariants before preparation and immediately before publication. Never
   retry or bypass a failed gate.
3. The final repository candidate passes relevant tests, `make ci`, the new
   preflight check, `actionlint .github/workflows/release.yml`, and exact-head
   hosted baseline, shadow, both macOS, PR paired benchmark, and trusted policy
   checks. Do not substitute predecessor results for changed bytes.
4. From clean archived source, explicit `L7_CLI_VERSION=1.0.0` and
   `L7_PACKAGE_CHANNEL=stable` builds reproduce unsigned inputs and pass
   stable package inventory, checksum/SBOM, native CLI/MCP, upgrade, rollback,
   uninstall, and residue checks. Use scratch checkouts for local fuzzing;
   leave the unrelated untracked audit untouched.
5. Before release evaluation, at least 20% of its corpus is a protected holdout
   outside candidate-readable/writable scope, operated or released by an
   independent evaluator as `L7-EVAL-007` requires. The owner must supply that
   real isolation; solo self-review is not a substitute. Freeze protocol,
   versions, repetitions, safety thresholds, and cost measurement before use.
   Until then, no release-evaluation PASS or improvement claim applies.
6. Under separate effect authority, verify all four Developer ID signatures,
   require Accepted notarization for both final ZIPs, and attest the exact
   prepared assets. Observe Codex and Claude install/discovery/invocation,
   provider execution, removal, and residue on those exact archives from
   disposable roots. Record host versions and architecture; untested support
   cells remain NOT_RUN or unqualified.
7. Publish only after explicit owner authorization binds the final commit/tree,
   workflow run, archive/checksum/manifest digests, release notes, and protected
   publication approval. Revalidate every gate, then create the absent
   annotated `v1.0.0` tag and immutable release once. No installation or
   deployment follows publication.
8. Phase 4 is not a prerequisite for releasing Phases 1–3. It still needs its
   own owner-authorized Tier 3 brief and a test proving auto-merge stops when
   `main` breaks. `READY_FOR_BRIEF` grants neither implementation nor merge
   authority.

## Risks and mitigations

- **Stale authority or lineage:** bind controls to the actual final candidate
  and real identities; do not carry historical team approvals forward.
- **False release confidence:** keep engineering PASS, formative evaluation,
  protected-holdout qualification, and release GO distinct.
- **Changed final bytes or excessive claims:** requalify signed assets by
  digest and keep support limited to observed hosts; claim no improvement
  from the current inconclusive comparison.
- **Secret leakage or irreversible publication:** use protected credentials,
  disposable state, redacted evidence, guarded cleanup, and explicit authority
  at each external effect. Never put secrets or user transcripts in Git.

## Rollback

Before implementation, remove only this new brief if the owner rejects it.
Later preparation changes can be reverted without modifying product behavior,
the immutable prerelease, or frozen rollback payloads. Failed preparation must
publish nothing; report unexpected remote state and stop for cleanup authority.
After publication, never move the tag or replace immutable assets. Use a new
reviewed patch release or separately authorized withdrawal/security notice,
and retain v0.1.1 as the documented skills-only rollback.
