# Level 7 v1.0.0 Stable Release

| Field | Value |
|---|---|
| Change ID | `l7-v1-stable-release` |
| Risk tier | `3` — release workflow, signing inputs, public package identity, and publication boundary |
| Status | `proposed`; implementation requires Product Owner approval of this exact brief commit |
| Base commit | `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` |
| Base tree | `9b6550c11e9666ab25166af6a5fb7560cdcc8cdf` |
| Product lineage | PR #15 merged by a two-parent commit whose parents are `84bd69f90d366356b0ce1e1a392f906258f3de91` and `97b16e8d29282e63b8a63742567c6693ffd13aac` |
| Passed lineage evidence | exact-head Harness `33498861738`; Trusted policy `33501335526`; merged-main Harness `33504902561` |
| Intended release | exact tag and GitHub release `v1.0.0`; no tag or release exists yet |
| Proposal branch | `codex/l7-v1-stable-release`, rooted directly at the exact base |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root` |
| PR author / release operator | `anup19950725`; must remain distinct from the accountable owner at hosted approval boundaries |
| Independent auditor | `apbusinessidentity-tech` |
| Hosted assurance | team mode; sole trusted PR label `l7-risk-tier-3` |
| Next executable transition | Stop for explicit Product Owner approval bound to this exact proposal commit |

## Problem

The v1 orchestration implementation is merged, but the distributable remains
truthfully identified as `1.0.0-dev`. Its two deterministic macOS packages are
unsigned development candidates, their provenance blocks release, and they do
not contain a self-contained local marketplace catalog that both claimed hosts
can install from the extracted release archive. The repository has no release
workflow, protected release environment, Apple signing/notarization identity,
or `v1.0.0` release record.

The passed PR and mainline runs establish the merged product lineage only. A
new release workflow, package identity, catalog, documentation, or signed byte
invalidates exact-byte verification. The mainline push workflow also skipped
the paired benchmark by design; the exact PR benchmark passed, but a fresh
release candidate must pass the unchanged gate. Historical provider trials do
not transfer: actual Codex and Claude marketplace installation and provider
execution must be observed again on the exact release candidate before either
support claim is promoted.

The smallest honest release change keeps product architecture and behavior
unchanged, makes stable package inputs explicit and strictly validated, adds
host-installable catalogs, and introduces one manual fail-closed release path.
Signing, notarization, provider network use, environment/secret configuration,
tag creation, and publication remain later external effects.

## Scope

After exact-brief approval, implementation may:

1. preserve `1.0.0-dev` as the ordinary default while accepting an explicit,
   canonical version/channel pair for release builds; reject mixed, ambient,
   malformed, or unsupported identities;
2. have the existing `l7pack` architecture produce the two exact host archives
   with strict, host-specific local marketplace catalogs, closed inventories,
   checksums, SPDX SBOMs, and input provenance that cannot assert approval,
   notarization, publication, or release readiness;
3. extend the existing candidate validator and lifecycle tests to cover both
   the development default and exact stable `1.0.0`, including catalog
   substitution, cross-host pairing, version/channel mismatch, unsafe paths,
   upgrade from `0.1.1`, rollback, removal, native CLI, and MCP behavior;
4. add a manually dispatched, immutable-action-pinned release workflow that
   runs only for an exact merged `main` commit, builds unsigned inputs twice,
   compares them before signing, imports Apple credentials only into a
   disposable keychain, signs all four Mach-O inputs with hardened runtime and
   secure timestamps, verifies the signatures, submits both final ZIPs with
   `notarytool --wait`, and fails unless both submissions are accepted;
5. make the workflow publish in a separate `v1-production` environment job:
   download the exact prepared artifact, verify its manifest, digests and
   GitHub attestations, create only absent annotated tag `v1.0.0`, stage and
   verify the exact release assets, and publish once without overwrite,
   replacement, retry-until-green, or partial-success promotion; and
6. replace development release wording with exact stable installation,
   verification, compatibility, permission, update, uninstall, and rollback
   instructions plus release notes that distinguish the bundled v1 engine from
   the still-available v0.1.1 skills-only package.

The unsigned build and validation path remains offline, deterministic, bounded,
and disposable. Signing and provider trials are explicit network boundaries;
they never run in the ordinary harness. Final signed bytes are not claimed to
be reproducible because secure timestamps change them: reproducibility applies
to the unsigned inputs, while final assets are bound by digests, notarization
results, a release manifest, and GitHub artifact attestations.

The workflow is inert until repository administrators separately create and
protect its signing and production environments, load Apple credentials, set
`addressanup` as the required production reviewer with self-review prevented,
and enable the intended immutable-release controls. Implementation must not
create or change those external resources. No brief, repository file, workflow
input, package metadata, verification record, or candidate-controlled marker
may substitute for the external reviewer or release authority.

Production orchestration, adapters, dependencies, root v0.1.1 marketplace
manifests, trusted policy, existing Harness workflow, benchmark code/comparator,
10% threshold, branch protection, historical records, and provider credentials
are outside scope. There is no automatic trigger on push or tag, updater,
installation, deployment, benchmark acceptance, architecture bypass, target
reduction, swallowed failure, broad timeout increase, or signing fallback.

## Exact implementation file set

Declared path count: 12 (5 Add, 7 Modify, 0 Delete).

Add:

- `.github/workflows/release.yml`
- `docs/artifacts/changes/l7-v1-stable-release.md`
- `docs/artifacts/changes/l7-v1-stable-release-verification.md`
- `docs/artifacts/changes/l7-v1-stable-release-audit.md`
- `docs/releases/v1.0.0.md`

Modify:

- `CHANGELOG.md`
- `Makefile`
- `README.md`
- `cmd/l7pack/main.go`
- `cmd/l7pack/main_test.go`
- `internal/harness/v1candidate/main.go`
- `internal/harness/v1candidate/main_test.go`

Delete:

- None.

## Acceptance criteria

1. This proposal commit has exact sole parent/base `e88b18e…`, adds only this
   brief, and leaves every implementation byte, historical record, worktree,
   remote, PR, tag, release, credential, and hosted setting unchanged.
2. Fresh external approval names this change ID, the exact proposal commit,
   base/tree, implementer, and nine non-evidence implementation paths. No prior
   approval, verification, audit, review, GO, or hosted result transfers.
3. Implementation descends without history rewriting from the approved brief
   and changes only the nine declared non-evidence paths. Verification and
   audit records are later single-purpose commits; all other declared paths and
   all historical governance records remain byte-identical.
4. Ordinary `make v1-candidate` still produces only `1.0.0-dev` development
   artifacts with release-blocked input provenance. Stable mode requires the
   exact explicit pair `1.0.0`/`stable`; conflicting filenames, manifests,
   catalogs, checksums, SBOMs, binary version output, or provenance fail closed.
5. Each archive has one closed, sorted inventory and exactly one host-specific
   manifest and local marketplace catalog. The catalog resolves only the
   extracted package itself. Missing, extra, cross-host, symlinked, traversing,
   oversized, version-mismatched, or digest-changed content is rejected.
6. Package metadata reports observed inputs only and has no `release_ready`,
   approval, exception, signing-success, notarization-success, or publication
   field that candidate bytes can use to authorize promotion.
7. Development and stable unsigned inputs build from the pinned offline
   toolchain, clean archived source, disposable roots, and hard size/time
   limits. Two clean stable builds match byte-for-byte before any signing.
8. The exact implementation candidate passes focused race/adversarial/package
   tests, `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7`, and
   `make v1-candidate-check GO_VERSION=1.26.7`. Any failure stops; there is no
   retry-until-green or reduced target inventory.
9. A fresh PR for the exact implementation head passes baseline, Go shadow,
   macOS arm64, macOS amd64, the unchanged paired benchmark threshold/minimum,
   and trusted policy with the sole label `l7-risk-tier-3`, an independent
   non-author review, and exact-head accountable-owner approval.
10. Under separate network/model authorization, the exact stable archives pass
    actual Codex and Claude marketplace validation, add, install, discovery,
    named skill invocation, provider execution, uninstall, marketplace removal,
    and residue checks from disposable host roots. Results bind host version,
    architecture, archive SHA-256, commands, and cleanup. A failed or unrun cell
    remains `NO_GO` or explicitly `NOT_RUN`; it is never inferred from help,
    compilation, the synthetic lifecycle, or predecessor evidence.
11. Under separate signing/notarization authorization, the prepare job uses only
    protected secrets, deletes its disposable keychain on success or failure,
    verifies Developer ID signatures on all four Mach-O inputs, receives
    `Accepted` for both ZIP submissions, emits the exact asset manifest and
    digests, and produces verifiable GitHub artifact attestations. Secrets,
    certificates, private keys, tokens, and notarization logs containing secret
    material never enter Git or release assets.
12. The publish job cannot start without a distinct `addressanup` approval in
    the protected `v1-production` environment and a separately stated release
    authorization bound to the exact commit/tree, workflow run, manifest, and
    asset digests. It stops on base/head/tree/check/review/tag/release/asset or
    environment drift and refuses any existing `v1.0.0` ref or release.
13. Publication attaches exactly the two host ZIPs, their checksums, the release
    manifest, and release notes; verifies the final asset set and attestations;
    then publishes once. GitHub-generated source archives are not represented
    as installable v1 packages. No installation or deployment follows.
14. One verification commit may add only this change's verification record and
    bind `PASS` or `FAIL` to the exact implementation commit/tree and observed
    evidence. Only after separate commissioning may `apbusinessidentity-tech`
    add one read-only audit record binding `GO` or `NO_GO` to the exact
    verification successor. A separate owner GO is still required before merge.
15. A benchmark regression above the unchanged 10% threshold blocks. Any later
    acceptance is a separate exact-head accountable-owner GitHub decision; it
    is not implementation authority and cannot be encoded in candidate bytes or
    used to bypass the trusted policy.
16. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and unstaged with SHA-256
    `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

## Risks and mitigations

- **Candidate-controlled release:** manual dispatch, exact commit/tag checks,
  protected secrets, a distinct production reviewer, attestations, and a final
  separately authorized effect keep repository metadata evidentiary only.
- **Signing ambiguity:** compare unsigned inputs first, sign every architecture
  binary with secure timestamps, verify locally, require two accepted Apple
  submissions, and bind the non-reproducible final bytes by digest.
- **Uninstallable assets:** include strict self-contained catalogs and require
  actual package-manager lifecycle trials on the exact ZIPs before support or
  publication.
- **Provider drift or cost:** pin and record host versions, require explicit
  model/network authority, bound each trial, and report failures or unrun cells
  without retries or broadened claims.
- **Partial or duplicate publication:** separate preparation from owner-gated
  publication, refuse pre-existing refs/releases/assets, verify the complete
  staged set, and publish only after every prerequisite passes.
- **Credential or user-state leakage:** use disposable keychain and host roots,
  minimal workflow permissions, masked secrets, guarded cleanup, and residue
  checks; stop rather than use ambient credentials or global configuration.
- **Historical or performance drift:** preserve every predecessor record and
  threshold, require fresh exact-head evidence, and stop on any undeclared path.

## Rollback

Before merge, revert or discard only this release change and remove ignored
build output with exact guarded paths; `main` remains at the last accepted
product tree. A failed preparation must publish nothing, revoke no credentials,
delete its disposable keychain and host roots, and leave any retained workflow
artifact clearly non-release. Any unexpected draft, tag, asset, or external
state is reported and left untouched until separately authorized cleanup.

After an immutable public release, do not move or delete `v1.0.0` and do not
replace assets in place. Correct defects with a new reviewed patch release and,
when necessary, a public revocation/security notice and Apple credential or
ticket response under separate authority. Uninstall removes only the extracted
marketplace/plugin and Level 7-owned disposable state; it never edits a user's
repository history or provider credentials.
