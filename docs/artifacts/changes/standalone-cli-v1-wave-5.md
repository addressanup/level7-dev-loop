# Standalone CLI v1 — Wave 5 Dual-Host Distribution Foundation

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-5` |
| Risk tier | `3` — plugin manifests, distribution controls, and hosted CI are protected surfaces |
| Status | `planned`; awaiting accountable-owner approval of the exact brief commit |
| Base commit | `f92c560cbe89e8d318e5521d9fc620f6153e9e14` |
| Base tree | `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0` |
| Proposed accountable owner | `apbusinessidentity-tech` |
| Implementer | `addressanup` using the active Codex implementation session |
| Runtime feature flag | Not applicable — this wave adds build-time package tooling and no runtime command or effect |

## Problem

The repository has a working default-OFF development CLI and hand-maintained
Codex, Claude, root-plugin, and marketplace manifests. It cannot yet prove that
two distinct host packages came from one versioned source, reproduce their
bytes, reject manifest drift, describe an exact compatibility boundary, or
exercise non-destructive package lifecycle rules. The existing `1.0.0`
manifest metadata can also be mistaken for a stable product claim even though
no supported installation, signing, publication, pilot, or release exists.

## Scope

Implement only the first bounded C4 distribution work package for the current
lean product:

- define one authored development-package descriptor with common metadata and
  explicit Codex and Claude overlays;
- render separate deterministic Codex and Claude package archives from exact
  allowlists, with canonical ordering, modes, timestamps, inventories, and
  digests;
- generate and drift-check the four existing repository manifest/catalog files
  from that descriptor;
- include matching prerelease version, changelog, MIT license text, permission
  declaration, compatibility metadata, SPDX SBOM, and unsigned provenance input
  in each development package;
- verify package structure, path safety, digest binding, cross-host separation,
  two-build reproducibility, and absence of mutable runtime references;
- exercise install, same-version reinstall, upgrade, rollback, prepare-removal,
  and removal semantics only in disposable filesystem fixtures, including
  conflicts, interruption, missing receipts, and preservation of unowned files
  and canonical repository artifacts; and
- run the offline distribution checks in the existing harness without creating
  a new support, signing, promotion, or release gate.

The package version is a shared prerelease development identity, not stable
`1.0.0`. Compatibility entries remain independently qualified and default to
`NOT_RUN` or development-only evidence. A pass for one host never promotes the
other or establishes a dual-host support claim.

This wave does **not** run Codex, Claude, a model, a provider network, host help
or version probes, or either build-tagged actual-host gate. It does not install,
enable, update, disable, or remove a real host plugin; touch user or global host
configuration; fetch a dependency; sign or notarize bytes; create an updater or
channel; publish an archive; create a GitHub release; deploy; pilot; or issue a
v1.0 decision. Production adapters remain pinned to Codex CLI `0.149.1` and
Claude Code `2.1.241`. Unknown-option rejection, typed `--max-turns`
enforcement, argument, permission, schema, cancellation, cleanup, and
containment controls are unchanged and may not be weakened.

## Architecture and contracts

- `distribution/package.json` is the only authored package metadata source. It
  contains common identity plus explicit host overlays; generated manifests are
  outputs and may not feed the source in reverse.
- `distribution/compatibility.json` is declarative claim data. Every tuple names
  host, plugin, schema, OS/architecture, required and optional capability,
  degradation, lifecycle evidence, rollback, and claim state. Unknown or
  incomplete fields fail closed.
- The distribution assembler is build-time harness code. It cannot import the
  product application, provider adapters, authority, Git mutation, CI envelope,
  or state packages and performs no network or host-process invocation.
- Each archive is built from a fixed tracked-file allowlist. Absolute paths,
  traversal, symlinks, duplicate names, special files, mutable remote content,
  and undeclared files are rejected.
- Package qualification is byte- and filesystem-based. A receipt binds host,
  version, package digest, owned paths, and prior digest. Conflict handling uses
  compare-before-write/remove semantics and never deletes an unowned or changed
  file.
- Generated SBOM and provenance input are development evidence only. They are
  unsigned and cannot satisfy package authenticity, promotion, or release
  requirements.

## Exact implementation file set

Add:

- `docs/artifacts/changes/standalone-cli-v1-wave-5.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-5-verification.md`
- `docs/artifacts/changes/standalone-cli-v1-wave-5-audit.md`
- `distribution/package.json`
- `distribution/compatibility.json`
- `LICENSE`
- `CHANGELOG.md`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`
- `internal/harness/distribution/archive.go`
- `internal/harness/distribution/archive_test.go`
- `internal/harness/distribution/lifecycle.go`
- `internal/harness/distribution/lifecycle_test.go`
- `scripts/harness/check-distribution.sh`

Modify:

- `.codex-plugin/plugin.json`
- `.claude-plugin/plugin.json`
- `.github/workflows/harness.yml`
- `plugin.json`
- `marketplace.json`
- `README.md`
- `Makefile`
- `harness/support-matrix.tsv`
- `harness/import-boundaries.tsv`
- `scripts/harness/check-import-boundaries.sh`

No other path is authorized. In particular, `.l7/config.json`, `go.mod`,
`go.sum`, product runtime packages, provider adapters and qualification tests,
skills, the trusted-policy workflow/controller, existing change records, local
or remote Git refs, host configuration, and the user-owned untracked foundation
audit remain unchanged. Scope expansion or a renamed path requires a revised
brief and fresh accountable-owner approval.

## Acceptance criteria

1. One strict authored descriptor deterministically produces distinct Codex and
   Claude manifests and catalogs; CI rejects hand-edited output drift.
2. Two clean builds from the same exact source produce byte-identical host
   archives and identical SHA-256 digests on the declared harness platforms.
3. Each archive contains only its allowlisted host manifest, the shared skills,
   changelog, license, permissions, compatibility entry, inventory, SPDX SBOM,
   and unsigned provenance input; it contains no other host's manifest.
4. Archive validation rejects traversal, absolute or duplicate paths, symlinks,
   special files, unsafe modes, mutable references, malformed metadata, digest
   mismatch, package substitution, and undeclared content.
5. Disposable lifecycle fixtures cover clean install, identical reinstall,
   upgrade, rollback, prepare-removal, removal, interruption, stale or missing
   receipt, and changed/unowned-file conflicts without modifying files outside
   the fixture or deleting canonical project artifacts.
6. Compatibility output names exact observed and target tuples while reporting
   provider execution and real host lifecycle as `NOT_RUN`; no host, dual-host,
   Intel-runtime, security, signing, release, or stable-v1 claim is inferred.
7. The CLI runtime behavior, default-OFF lifecycle flag, production provider
   pins, provider argv, parser controls, authority checks, and actual-host gates
   are byte-for-byte unchanged.
8. Distribution build and qualification make no network call, prompt, provider
   invocation, host installation, global write, repository-state write, Git
   mutation, signing, publication, or deployment effect.
9. Production dependencies remain zero. Package tooling uses the pinned Go
   standard library, emits bounded diagnostics, and observes the existing import
   and effect boundaries.
10. Formatting, lint, type checks, full tests, race tests, reproducibility,
    macOS cross-builds, distribution drift/reproducibility checks, protected CI,
    scope inspection, and `git diff --check` pass for the exact candidate.

## Risks and mitigations

- Development archives may look releasable. Use an explicit prerelease version,
  development channel/status, unsigned provenance label, and claim-withheld
  compatibility entries; documentation continues to block publication.
- A generator could silently package unintended files. Use a closed allowlist,
  normalized paths, source-digest binding, archive reinspection, and two-build
  equality.
- Lifecycle tests could overstate real package-manager safety. Label them
  filesystem fixtures and keep every actual-host lifecycle cell `NOT_RUN` until
  separately authorized and observed.
- Removal logic could destroy user data. Require owned-path receipts and exact
  digest comparison; changed, unknown, or missing ownership blocks removal and
  preserves the file.
- Generated metadata could drift from repository manifests. Treat the authored
  descriptor as one-way truth and fail CI on any byte difference.
- SBOM/provenance files could be mistaken for authenticity. Mark them unsigned
  build inputs; signing, channel trust, revocation, and promotion remain later
  Tier 3 gates.

## Rollback

Before integration, revert the small conventional Wave 5 commits in reverse
order. Generated archives live only under ignored `build/` paths and can be
discarded without touching host state. No real installation, migration,
provider session, remote ref, signing key, release, deployment, or publication
exists to roll back. After any later separately authorized integration, use a
new reviewed revert; never rewrite preserved history or remove user artifacts.

## Planned commit sequence

1. `docs(distribution): define standalone cli wave five`
2. `test(harness): reserve distribution effect boundaries`
3. `feat(distribution): generate deterministic host packages`
4. `test(distribution): qualify reversible fixture lifecycle`
5. `ci(distribution): verify dual-host development packages`
6. one verification-only commit, then one independent audit-only commit after
   their separately authorized transitions.
