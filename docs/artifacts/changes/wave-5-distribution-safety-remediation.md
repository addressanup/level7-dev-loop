# Wave 5 Distribution Safety Remediation

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Risk tier | `3` — package identity and destructive fixture-lifecycle controls |
| Status | `approved` for this bounded repository-local implementation through the active user interaction |
| Base commit | `79a0739202942970536cc29782ad1b3952e7d15e` |
| Base tree | `49400a5f97355b2ccbab27d165074d6be0a24757` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `codex-root` |
| Runtime feature flag | Not applicable — this changes build-time validation and disposable fixture behavior only |

## Problem

The merged Wave 5 distribution foundation is reproducible and offline, but a
read-only review found bounded fail-closed gaps. An interruption after receipt
publication can leave an otherwise committed install permanently blocked by
its pending record. Removal accepts unowned empty directories and has no
transaction record from which to resume after a late filesystem failure.
Lifecycle input validation proves archive integrity without binding the caller's
host and version labels to the archive metadata and host manifest, and an
inactive known digest can be relabelled in a later receipt. Compatibility
validation also accepts arbitrary non-empty host tuples rather than the exact
development claims qualified by the repository.

Wave 5 already generates one inventory-covered `DISTRIBUTION.json` in each host
archive even though the historical brief's prose omitted it from the archive
member list. That historical record remains immutable. This successor
prospectively retains and validates the useful metadata member; its
`source_digest` binds authored inputs and the pre-derived package payload, not
the final archive digest.

The historical Wave 5 approval chronology cannot be reconstructed from current
repository evidence. This remediation neither rewrites nor claims to repair
that evidence limitation.

## Scope

Implement only these fail-closed corrections:

1. Make install completion idempotent when the exact pending package is already
   the active, receipt-bound transaction. Remove only the matching stale pending
   record, and reject any conflicting receipt or known-digest relabelling.
2. Reject every package directory not implied by an owned file path. Journal a
   validated removal before mutation, make retries tolerate only already-removed
   owned paths, revalidate every remaining path, and clear the journal only
   after package bytes and the receipt are gone.
3. Bind lifecycle package host and version labels to exactly one strict
   `DISTRIBUTION.json` and the corresponding Codex or Claude manifest. Reject
   cross-host, cross-version, duplicate, missing, conflicting, or promoted
   metadata on a clean fixture root.
4. Validate the complete, ordered Codex and Claude compatibility tuples against
   the exact repository-qualified development boundary, including product,
   surface, versions, targets, platforms, capabilities, degraded behavior,
   rollback, evidence states, and withheld support.
5. Add focused interruption, retry, directory-conflict, identity-substitution,
   known-digest, and compatibility-mutation tests.

Each Codex and Claude archive and expanded marketplace package must retain
exactly one generated `DISTRIBUTION.json`. It remains schema 1 and binds package
name, prerelease version, development channel, host, manifest path, catalog
path, deterministic source digest, builder identity,
`support_claim=WITHHELD`, and `actual_host_gate=NOT_RUN`. It remains
inventory-covered and cannot be removed, renamed, promoted, or interpreted as
signing, authenticity, release, provider, or actual-host evidence.

Do not change authored distribution data, generated manifests, generated
archive content, runtime or provider code, CI or policy controls, dependencies,
historical records, external hosts, remotes, or user/global configuration. No
real plugin install or removal is authorized.

## Exact implementation file set

Add:

- `docs/artifacts/changes/wave-5-distribution-safety-remediation.md`
- `docs/artifacts/changes/wave-5-distribution-safety-remediation-verification.md`
- `docs/artifacts/changes/wave-5-distribution-safety-remediation-audit.md`

Modify:

- `internal/harness/distribution/lifecycle.go`
- `internal/harness/distribution/lifecycle_test.go`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`

No other path is authorized. In particular, all earlier briefs, verification
records, audits, `distribution/*.json`, generated root manifests, workflows,
the build controller, and the user-owned untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md` remain unchanged.

## Acceptance criteria

1. The implementation is a successor of exact base
   `79a0739202942970536cc29782ad1b3952e7d15e`, and only the declared paths
   differ from that base.
2. A fault after writing the new install receipt but before deleting its pending
   record is recoverable. Recovery accepts only the exact already-committed
   transaction, deletes the pending record, preserves the receipt and package,
   and permits subsequent lifecycle operations.
3. Receipt completion never overwrites different version or ownership metadata
   for a digest already known to the receipt, whether active or inactive.
4. Prepare-removal rejects an unowned empty directory and any other undeclared
   filesystem entry without changing package, receipt, or unowned bytes.
5. Removal writes a host-bound pending record before its first deletion. An
   injected mid-removal or post-receipt interruption resumes idempotently;
   missing owned paths are accepted only under that valid record, while every
   remaining path is still ownership-, metadata-, content-, and containment-
   checked before deletion.
6. Install, rollback, and removal fail closed while a conflicting lifecycle
   transaction is pending. Malformed, substituted, cross-host, symlinked, or
   escaping pending state cannot authorize a mutation.
7. A package validates only when its caller host and version match its strict
   distribution metadata and matching host manifest. A correctly checksummed
   archive relabelled to the other host or another valid prerelease fails on a
   clean root without writing lifecycle state.
8. Each package contains exactly one inventory-covered `DISTRIBUTION.json` with
   the retained development-only contract above and no opposite-host manifest.
9. Compatibility validation rejects every mutation of the exact Codex and
   Claude tuples, including arbitrary non-empty values and reordered or extra
   capabilities.
10. Package construction remains byte-identical to the base: Codex archive
    digest `9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0` and
    Claude archive digest
    `718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb`.
11. Focused distribution tests, race-enabled distribution tests, repository
    verification, cross-builds, policy checks, and diff hygiene pass on the
    immutable implementation candidate. Unexecuted hosted or actual-host checks
    remain explicitly unverified.
12. The Tier 3 candidate receives one successor verification record and a
    separately authorized independent read-only `l7-release` audit before any
    readiness, merge, release, deployment, or publication decision.

## Risks and mitigations

- **Destructive retry risk:** the removal journal contains a previously
  validated receipt, but retries still validate all remaining paths before
  deleting them and never follow symlinks.
- **State-machine ambiguity:** install and removal use separate host-bound
  records and mutually exclude each other. Only exact committed state is
  idempotent; conflicting state fails closed.
- **Compatibility drift:** exact tuples intentionally require a validator and
  test change when declarative claims change.
- **Behavioral regression:** tests inject failures at each durable boundary and
  the full offline distribution digests must remain identical.
- **Scope or evidence expansion:** protected external actions and the historical
  approval limitation remain out of scope and unclaimed.

## Verification strategy

Run the focused distribution package first, including fault-injection and
mutation cases, then its race-enabled suite. Rebuild/check both development
packages twice and compare the exact base digests. Run the repository-pinned
full verification, declared CLI cross-builds, policy controller, and
`git diff --check`; inspect the exact base-to-candidate diff and tracked/index
cleanliness. Record only observed results in the one verification successor.

## Rollback

Before merge, discard or revert the implementation commit and then this brief
commit. After merge, revert the future audit, verification, implementation, and
brief commits in reverse order. This restores the exact base tree; no data
migration, external state, host installation, publication, or production
rollback is involved. Disposable fixtures may be deleted in their entirety.
