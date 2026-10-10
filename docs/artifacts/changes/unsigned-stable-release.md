# Owner-Authorized Unsigned Stable Release — Change Brief

| Field | Value |
|---|---|
| Change ID | `unsigned-stable-release` |
| Risk tier | `3` — release policy and irreversible publication |
| Status | Owner explicitly authorized unsigned stable v1.0.0 and all four waivers on 2026-10-10 |
| Base commit | `2432079234724a3f3a61ca73229305fc8a330f43` |
| Base tree | `74eb2c99f33325b857885dea860fd0613056d9a3` |
| Accountable owner | `addressanup`, the actual authenticated repository owner |
| Assurance | `solo`; truthful self-review, no independent audit |
| Authority | Prepare and publish unsigned stable v1.0.0 under the owner's explicit waiver decision; no installation, deployment, or changes to hosted protection settings |

## Problem

The signed release path cannot run without credentials, hosted controls, and
evaluation/provider evidence that do not exist. The owner changed the release
decision instead of requesting those prerequisites.

The owner authorized the release, then explicitly selected these four waivers
and **Unsigned stable v1.0.0** rather than a new prerelease:

1. Apple signing and notarization.
2. Protected holdout evaluation.
3. Separate release operator and protected approval.
4. Exact-asset Codex and Claude provider trials.

This deliberately replaces the signed/qualified publication requirements in
`release-readiness` for this release only. It does not invent successful
evidence, amend the public evaluation results, or claim that CI substitutes
for the waived gates. Their evidence remains `NOT_RUN`; formal support is
`WITHHELD`. macOS Gatekeeper may block the binaries. Never advise disabling it
globally.

All six product foundation steps remain complete. Runtime behavior, default-
OFF feature flags, Phase 4, dependencies, corpus, and evaluator are unchanged.

## Scope

Replace only the stable release policy and its qualification claims with the
explicit owner-authorized unsigned path. Preserve exact candidate checks,
reproducible packaging, truthful disclosures, attestation verification,
owner identity, and immutable publication. No installation, deployment,
provider execution, evaluator/corpus changes, or hosted setting changes.

## Exact implementation file set

Modify:

- `.github/workflows/release.yml`
- `scripts/harness/check-stable-release-preflight.sh`
- `README.md`
- `CHANGELOG.md`
- `docs/releases/v1.0.0.md`

Add:

- `scripts/harness/check-stable-release-preflight.py`
- `scripts/harness/publish-unsigned-stable.sh`
- `docs/artifacts/changes/unsigned-stable-release.md`

Delete: none. Preserve the user's untracked foundation audit, immutable
v1.0.0-dev, frozen v0.1.1 bytes, and all historical records.

## Acceptance criteria

1. Publish only v1.0.0, non-prerelease, explicitly titled UNSIGNED and NOT
   NOTARIZED. Notes and manifest disclose all four waivers, absent publisher
   identity/notarization/evaluation/provider evidence, and withheld support.
2. Require exact squash lineage, unchanged main and tested PR tree, one Tier 3
   label, latest exact-head baseline/shadow/both macOS/benchmark/trusted policy
   checks, and successful post-merge baseline/shadow/both macOS checks.
3. Only the real owner may dispatch or publish. Require one raw exact-candidate
   waiver authorization on the merged PR and the explicit dispatch
   acknowledgement. Reject drift, ambiguous authorization, duplicates, reruns,
   previous prepared artifacts, and any existing stable tag/release.
4. Compare two clean builds of all four native inputs and both stable ZIPs.
   Validate inventory, checksums/SBOMs, native CLI/MCP, upgrade, rollback,
   uninstall, and residue. Do not call these actual provider trials.
5. Actions prepares and attests only; it has no contents-write permission and
   receives no owner/admin credential. A local owner publisher uses the
   existing forge login, checks live immutable-release settings and trusted
   variables, and verifies exact workflow/run/artifact/digest/attestation
   provenance before any remote write. Do not modify hosted environments.
6. Attach exactly two installable ZIPs, SHA256SUMS, and RELEASE-MANIFEST.json.
   Validate archive extraction before use. Re-download the uploaded draft
   assets and compare every byte. Recheck owner authority, CI, main, prepared
   artifact, annotated tag, draft identity, and immutability immediately before
   publishing. Verify final release is immutable and non-prerelease.
7. Offline tests must reject invented qualification, actor/candidate/CI drift,
   duplicate/reused artifacts, modified assets, unsafe extraction, disabled
   immutability, and unexpected tag/draft state. Run full CI, workflow lint,
   stable package checks, and hosted checks on the final implementation.

## Risks and mitigations

Waived signing and qualification lower assurance by explicit owner choice.
Stable channel describes publication intent, not Apple trust or formal
support. Checksums and GitHub attestations verify integrity/provenance only.

## Rollback

Before publication, revert this bounded policy change normally. Once a tag
or draft exists, failures stop without deleting, retrying, moving tags, or
replacing assets; obtain separate cleanup authority. After immutable
publication, use a reviewed patch release or separately authorized withdrawal.
Never replace v1.0.0-dev or the v0.1.1 rollback.
