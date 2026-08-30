# Plugin Package Release Qualification — Offline Binding

| Field | Value |
|---|---|
| Change ID | `plugin-package-release-qualification` |
| Risk tier | `3` — distribution identity and release-readiness semantics |
| Status | `approved` for this bounded repository-local implementation by the active user |
| Base commit | `c17a4c7db2aa4ef0266430d947ae32d4456f51a8` |
| Base tree | `c20a20c79e9ca41fa8932baef5d8da89fcc42a44` |
| Assurance | `solo`; stronger automated verification and truthful self-review |
| Runtime feature flag | Not applicable — offline build validation only |

## Problem

The deterministic Codex and Claude development archives do not currently bind
their marketplace catalog bytes into package metadata or source identity. A
valid archive could therefore be paired with a substituted catalog without the
package validator detecting that mismatch. The distribution check also reports
digests only as text and does not expose a deterministic machine-readable
decision that distinguishes offline package qualification from release
readiness.

## Scope

This change will:

1. advance the development package identity to `0.1.0-dev.6` and accept only a
   canonical `MAJOR.MINOR.PATCH-dev.N` version for the development channel;
2. bind each exact catalog path and SHA-256 digest into schema-2
   `DISTRIBUTION.json`, the package source digest, reproducibility checks, and
   package validation;
3. reject catalog mutation, substitution, cross-host pairing, and any
   inconsistent package/catalog identity before synthetic lifecycle mutation;
4. add an offline qualification report with ordered per-host source, archive,
   and catalog digests plus explicit external-gate states; and
5. expose that report as canonical JSON only through `--check --json`.

The archive/catalog relation is Level 7 build evidence only. The provider hosts
are not documented to consume `DISTRIBUTION.json` or enforce its digest, so this
change cannot prove which bytes a real host resolves, caches, installs, or
enables.

The implementation does not run a real host or provider, install a plugin,
access the network, sign an artifact, publish, deploy, evaluate owner authority,
or authorize a release. Codex and Claude remain separately unqualified at the
actual-host boundary. Support remains withheld and `release_ready` is always
false in this offline evaluator.

## Exact implementation file set

Add:

- `docs/artifacts/changes/plugin-package-release-qualification.md`
- `internal/harness/distribution/qualification.go`
- `internal/harness/distribution/qualification_test.go`

Modify:

- `.claude-plugin/plugin.json`
- `.codex-plugin/plugin.json`
- `CHANGELOG.md`
- `README.md`
- `distribution/package.json`
- `internal/harness/distribution/lifecycle.go`
- `internal/harness/distribution/lifecycle_test.go`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`
- `marketplace.json`
- `plugin.json`

Do not modify compatibility claims, support matrices, provider adapters,
workflows, policy controls, skills, dependencies, remotes, or historical
records. Do not add a verification or audit artifact.

## Acceptance criteria

1. Development versions accept only canonical `*-dev.N` values with a positive
   ordinal and reject stable, release-candidate, arbitrary, zero, or
   leading-zero forms.
2. The descriptor, changelog heading, four generated root metadata files, both
   host packages, and Claude catalog bind `0.1.0-dev.6`.
3. Schema-2 `DISTRIBUTION.json` binds the exact host catalog path and digest,
   and each source digest includes the named catalog bytes.
4. Package validation rejects changed catalog bytes or digest, missing catalog
   data, cross-host catalogs, and inconsistent manifest/catalog identities.
5. Two clean builds reproduce archive and catalog bytes and their independent
   digests; existing archive, inventory, provenance, SBOM, permission, and
   reversible synthetic lifecycle checks still pass.
6. `--check --json` emits one canonical report with ordered Codex and Claude
   results, explicitly internal offline checks `PASS`, host validation,
   marketplace lifecycle, installation, activation, behavioral activation,
   provider execution, signing, and publication `NOT_RUN`, support `WITHHELD`,
   authority `NOT_EVALUATED`, and `release_ready=false`.
7. JSON mode rejects ambiguous arguments, writes no repository or host state,
   and cannot accept approval or promotion input.
8. Focused tests, race tests, the full offline verification suite, declared
   cross-builds, diff hygiene, and final self-review pass without making an
   external-effect or release claim.

## Risks and mitigations

- **Catalog/package split:** the builder includes the catalog in source identity,
  and the validator recomputes its digest and matches schema-2 metadata before
  any lifecycle fixture writes.
- **Semantic promotion:** the report schema admits only the exact withheld and
  not-run states and never consumes authority input.
- **Compatibility drift:** host target and support declarations remain
  unchanged; the version bump represents development package bytes only.
- **Later host isolation:** any separately authorized host trial must bind the
  bytes actually resolved and use disposable host state (`CODEX_HOME`, or the
  Claude configuration, plugin-cache, and temporary roots) while preserving the
  documented limits of those isolation controls.

## Rollback

Before integration, discard this isolated branch. After later integration,
revert this change as one reviewed unit. No host, provider, user configuration,
publication, signing, deployment, or migration state requires cleanup.
