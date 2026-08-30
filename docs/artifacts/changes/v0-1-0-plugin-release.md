# Level 7 Plugin v0.1.0 — Release Brief

| Field | Value |
|---|---|
| Change ID | `v0-1-0-plugin-release` |
| Risk tier | `3` — public package identity and release boundary |
| Status | `approved` by the active user for this bounded finish slice |
| Base commit | `1ffe87310320d10103aaa625d44d10d217652974` |
| Base tree | `53be6fe314c9ec57d12a82c2db3cd8a273d49d66` |
| Assurance | `solo`; automated verification and truthful self-review |
| Runtime feature flag | Not applicable — the package contributes instructions only |

## Problem

The plugin implementation and deterministic dual-host packaging are complete,
but the generated artifacts still identify themselves as an unreleased
development build. Neither final package has been installed, discovered, and
invoked on its claimed host. The repository also lacks a short user path for
installing, using, and removing the plugin.

Further feature or governance work would delay the product without improving
this release. The finish line is one stable package identity, one bounded smoke
on each claimed host, concise usage documentation, and a release candidate whose
exact bytes can be authorized at the publication boundary.

## Scope

This change will:

1. freeze the v0.1 feature set and promote only the instruction plugin packages
   to strict version `0.1.0` on the `stable` channel;
2. replace development-only marketplace and package wording while preserving
   deterministic archives, strict validation, reversible synthetic lifecycle
   checks, and fail-closed metadata binding;
3. build the final Codex and Claude archives, then—under separate immediate
   authorization—validate, add, install, discover, explicitly invoke
   `l7-next`, uninstall, and remove the marketplace in isolated disposable host
   state on the local macOS arm64 machine;
4. record only observed, version-bounded smoke facts after both final archives
   pass; formal support, signing, publication, and release authority remain
   withheld or unevaluated until their real boundaries; and
5. add a short install/use/remove guide and make the plugin/standalone-CLI
   distinction explicit.

The standalone Go CLI remains an unbundled `0.1.0-dev` preview. This slice does
not add features, providers, executables, hooks, MCP servers, network access,
telemetry, updater behavior, signing infrastructure, governance records, or a
new review/approval chain.

Actual provider invocation may use the user's existing Codex or Claude
credential through its normal host mechanism and may access the provider
network. It must not run until the exact executable, version, archive digest,
command, disposable paths, residuals, and cleanup are presented for immediate
authorization. Pushing, opening or merging a pull request, tagging, and creating
a release remain separate exact-candidate effects.

## Exact implementation file set

Add:

- `docs/artifacts/changes/v0-1-0-plugin-release.md`

Modify:

- `.claude-plugin/plugin.json`
- `.codex-plugin/plugin.json`
- `CHANGELOG.md`
- `README.md`
- `distribution/compatibility.json`
- `distribution/package.json`
- `internal/harness/distribution/lifecycle.go`
- `internal/harness/distribution/lifecycle_test.go`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`
- `internal/harness/distribution/qualification.go`
- `internal/harness/distribution/qualification_test.go`
- `marketplace.json`
- `plugin.json`

Do not modify skills, provider adapters, the standalone CLI version, policy or
workflow controls, dependencies, remotes, repository rules, historical records,
or user-owned files. Do not add a verification or audit artifact.

## Acceptance criteria

1. The descriptor accepts exactly canonical stable SemVer and the `stable`
   channel; prerelease, build-suffixed, leading-zero, or mismatched identities
   fail closed.
2. The changelog, four tracked manifests/catalogs, both host catalogs, package
   metadata, archives, and checksums bind version `0.1.0` and the stable
   marketplace identity.
3. Existing package structure, catalog binding, archive reproducibility,
   inventories, SBOM/provenance inputs, containment, interruption recovery,
   reinstall, upgrade, rollback, conflict preview, and removal checks pass.
4. The final Codex archive passes host validation, marketplace add, installation,
   discovery, explicit `$l7-next` invocation in a disposable Git repository,
   uninstall, marketplace removal, and bounded state cleanup.
5. The final Claude archive passes the equivalent lifecycle with explicit
   `/level7-dev-loop:l7-next` invocation and bounded cleanup.
6. Compatibility metadata records only the exact host versions, architecture,
   and smoke outcomes actually observed. Unobserved architectures stay
   `NOT_RUN`; support stays `WITHHELD`; signing and publication stay `NOT_RUN`;
   offline qualification stays unable to authorize release.
7. README instructions accurately cover install, use, update/reinstall, remove,
   the instruction-only permission boundary, and the separate experimental CLI.
8. Focused tests, race tests, full repository verification, declared cross-builds,
   manifest validation, diff hygiene, and final self-review pass.

## Risks and mitigations

- **False promotion:** stable version/channel coupling is strict, while host
  observations, support, signing, publication, and authority remain separate
  states.
- **Testing the wrong bytes:** each host smoke binds the archive SHA-256 and runs
  against an extracted final archive; any package-affecting repair requires a
  rebuild and repeat of both host smokes.
- **Host-state leakage:** tests use a no-remote disposable repository and
  isolated host configuration/cache roots where supported, followed by explicit
  uninstall, marketplace removal, and guarded deletion. macOS Keychain access
  and provider-side request logs cannot be isolated and are disclosed before
  authorization.
- **Scope growth:** only smoke blockers may be fixed. All new features, broad
  compatibility, signing, updater, pilots, and historical backlog items remain
  deferred.

## Rollback

Before publication, discard this isolated branch and its generated build output.
After publication, remove the release assets and revert the release commit; do
not rewrite existing tags without explicit destructive authority. Host-trial
state is removed during each smoke, and the plugin contributes no persistent
runtime service or data migration.
