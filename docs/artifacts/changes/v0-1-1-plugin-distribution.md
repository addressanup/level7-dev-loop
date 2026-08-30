# Level 7 Plugin v0.1.1 — Direct Marketplace Distribution Brief

| Field | Value |
|---|---|
| Change ID | `v0-1-1-plugin-distribution` |
| Risk tier | `3` — protected public plugin manifests and release identity |
| Status | `authorized` for bounded local implementation by the active user |
| Base commit | `9306d9f14c0ac9ecae58ac33667039b1a8f35703` |
| Base tree | `6528437b7d2020a594058b5d047c7acb97ae297a` |
| Assurance | `solo`; automated verification and truthful self-review |
| Runtime feature flag | Not applicable — distribution-only, with unchanged instruction behavior |

## Problem

The released `v0.1.0` plugin is usable, but installing it from GitHub requires a
source checkout, a pinned Go bootstrap, and local marketplace generation. The
repository does not contain the standard Codex and Claude marketplace catalogs
at the paths their Git-backed installers discover. Pointing either catalog at
the repository root would also copy the development harness and documentation
into the plugin cache.

The product feature set is finished. This patch should make the existing 12
skills installable through normal host marketplace commands without adding or
changing plugin behavior.

## Scope

This change will:

1. publish standard committed catalogs at `.agents/plugins/marketplace.json`
   and `.claude-plugin/marketplace.json`;
2. point both catalogs at one physical `plugins/level7-dev-loop/` payload that
   contains only the Codex and Claude manifests, 12 exact skill files, concise
   package README, changelog, and MIT license;
3. promote only the instruction-plugin package identity to stable `0.1.1`,
   leaving the standalone Go CLI at `0.1.0-dev`;
4. extend the deterministic distribution checker so the committed catalogs and
   clean plugin payload must exactly match the canonical descriptor and source
   skills, reject extra payload files, and remain aligned with generated release
   marketplaces; and
5. replace the build-first installation guide with direct, tag-pinned Codex and
   Claude marketplace commands, while retaining build-from-source instructions
   for maintainers.

No skill behavior, provider adapter, hook, MCP server, executable, network
capability, telemetry, updater, workflow, protected policy controller, or
dependency is added or changed. Public-directory submissions, pushing, pull
requests, merging, tagging, and GitHub Release publication remain separate
external effects.

## Exact implementation file set

Add:

- `.agents/plugins/marketplace.json`
- `.claude-plugin/marketplace.json`
- `plugins/level7-dev-loop/.codex-plugin/plugin.json`
- `plugins/level7-dev-loop/.claude-plugin/plugin.json`
- `plugins/level7-dev-loop/README.md`
- `plugins/level7-dev-loop/CHANGELOG.md`
- `plugins/level7-dev-loop/LICENSE`
- `plugins/level7-dev-loop/skills/*/SKILL.md`
- `docs/artifacts/changes/v0-1-1-plugin-distribution.md`

Modify:

- `.codex-plugin/plugin.json`
- `.claude-plugin/plugin.json`
- `plugin.json`
- `marketplace.json`
- `distribution/package.json`
- `distribution/compatibility.json`
- `CHANGELOG.md`
- `README.md`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`
- `internal/harness/distribution/qualification.go`
- `internal/harness/distribution/qualification_test.go`

Do not modify `skills/**`, provider adapters, the standalone CLI version,
workflows, policy controls, dependencies, historical change records, remotes, or
user-owned files.

## Acceptance criteria

1. `git diff v0.1.0...<candidate> -- skills` is empty, and every committed
   marketplace skill is byte-identical to its canonical `skills/**/SKILL.md`.
2. Both committed catalogs contain exactly one `level7-dev-loop` entry in the
   `level7-engineering` marketplace and resolve only to the physical
   `./plugins/level7-dev-loop` payload.
3. The clean payload contains exactly both host manifests, all 12 skills,
   README, changelog, and license; it contains no harness,
   executable, hook, MCP configuration, telemetry, or symlink.
4. Descriptor, changelog, root and payload manifests, legacy catalog, generated
   host catalogs, package metadata, archive names, and checksum sidecars bind
   canonical stable version `0.1.1`.
5. `make distribution-check` fails on stale catalog or manifest metadata,
   missing or changed payload bytes, an extra payload file, a source escape, or
   any source-skill drift; existing archive, reproducibility, containment,
   reinstall, upgrade, rollback, conflict, and removal checks continue to pass.
6. In disposable host state, Codex CLI `0.151.0` and Claude Code `2.1.247`
   validate and complete marketplace add, discovery, install, list/details,
   uninstall, and marketplace removal for the exact candidate package. Any
   provider invocation or remote Git/tag smoke is separately disclosed and
   authorized before it runs.
7. Compatibility metadata records only observations actually established for
   `0.1.1`; unobserved cells remain `NOT_RUN`, formal support remains
   `WITHHELD`, and offline qualification cannot authorize publication.
8. Focused tests, race tests, full `make ci`, declared macOS cross-builds,
   plugin/JSON validation, diff hygiene, and final solo self-review pass.
9. README installation requires no clone, Go toolchain, build, ZIP extraction,
   or manual file copying. Update and removal commands are explicit, and users
   are told when to start a new Codex task or reload Claude plugins.
10. Publication remains a separate exact-candidate action using an annotated
    `v0.1.1` tag and a GitHub Release with two ZIPs and their two SHA-256 files.

## Risks and mitigations

- **Payload drift:** deterministic checks bind both copied manifests and every
  copied skill to canonical source bytes and reject missing or extra files.
- **Oversized installation:** both catalogs target the bounded physical plugin
  subdirectory rather than the repository root.
- **Host-schema divergence:** the Codex and Claude catalogs retain their native
  schemas while resolving the same self-contained plugin payload.
- **False compatibility promotion:** package-manager and provider observations
  are separated; unchanged skill content does not silently create a new live
  provider result.
- **Scope growth:** directory listing assets, hosted policy pages, analytics,
  new skills, and automation are deferred until the direct marketplace path is
  complete.

## Rollback

Before publication, discard this isolated branch. After publication, revert the
`v0.1.1` distribution commit and direct users to the immutable `v0.1.0` release;
do not move or rewrite an existing tag. Users can uninstall
`level7-dev-loop@level7-engineering` and remove only the `level7-engineering`
marketplace. The plugin owns no service, database, migration, or persistent
runtime data.
