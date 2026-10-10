# Verified CLI Installation and Updates — Change Brief

| Field | Value |
|---|---|
| Change ID | `plugin-cli-install-update` |
| Risk tier | `3` — installer, updater, and distribution trust boundary; mutates user-local host plugin registrations |
| Status | implemented locally; review-ready, not published |
| Base commit | `ed1457f2c54e2308f757d7ea3bf4fcfa113575ec` |
| Accountable owner | `addressanup` |
| Assurance | `solo`; truthful self-review, no independent audit claimed |
| Authority | repository-local implementation and disposable tests only; no push, PR, release, hosted-setting change, or real-user installation |

## Problem

Installing v1.0.0 requires `gh`, `jq`, a hand-typed verification recipe, manual
ZIP extraction, and two host-specific marketplace commands; updating repeats
marketplace removal and re-registration by hand. The Git marketplace payload
still identifies as the frozen v0.1.1 instruction plugin, so refreshing it
never installs v1. Nothing keeps the `l7` CLI, the Codex plugin, and the
Claude plugin coherent, and nothing retains a verified previous version for
an offline rollback.

## Scope

Add `l7 plugin install|update|status|rollback|uninstall` plus a one-time,
reviewable bootstrap script that installs a verified `l7` into a user-owned
prefix and then runs the same install logic. Shared release resolution,
verification, versioned storage, locking, journaling, and state live in
`internal/l7/adapter/plugin*`; Codex and Claude differ only in small command
adapters that reuse the hosts' own marketplace and plugin commands.

Trust root: the fixed repository `addressanup/level7-dev-loop`, the allowlisted
release workflows, `refs/heads/main`, and GitHub artifact attestations
verified by `gh attestation verify --bundle` (gh ≥ 2.67.0, no GitHub login).
Bundles come from the public attestations API; assets come from the release
download URLs over bounded HTTPS with an explicit redirect host allowlist and
no credentials. The manifest, `SHA256SUMS`, the selected archive, and the
extracted CLI binary are each bound by digest and attestation to the exact
source commit before anything is executed or registered.

Explicit boundaries kept: stable is the default channel; prereleases, exact
versions, downgrades, unsigned binaries, and withheld-support releases each
need their own flag; `--yes` only confirms non-trust prompts. Nothing installs
or updates implicitly, no telemetry or credentials are stored, Gatekeeper and
quarantine are never touched, and the immutable v1.0.0, v1.0.0-dev, and v0.1.1
releases plus the frozen Git marketplace are unchanged. The two existing
release manifests are consumed as-is; no new release is published.

Declared boundary change: `internal/l7/adapter/pluginrelease` becomes the
second allowed `net/http` owner under `BND-607`, with positive controls in
`scripts/harness/check-import-boundaries.sh`.

## Acceptance criteria

1. After the bootstrap, `l7` is on a user-owned path and the documented
   commands install, update, inspect, roll back, and remove each host's plugin
   without manual ZIP, digest, or marketplace steps.
2. Every network operation is explicit (`install`, `update`, `status --check`);
   `status`, `rollback`, and `uninstall` work offline. `L7_NETWORK=off` blocks
   the network path with a precise message.
3. Byte integrity and attestation provenance (repository, workflow, source
   ref, exact commit, subject digest) are verified for the manifest, checksums,
   archive, and CLI binary before extraction, execution, or registration.
   Unsafe archives, inventory drift, wrong host/OS/architecture, bad digests,
   missing assets, and malformed state fail closed without mutation.
4. Unsigned or unqualified releases are refused until `--accept-unsigned` and
   `--accept-unqualified` are passed; the refusal names the exact flags.
5. Installations are immutable versioned directories with ownership receipts;
   a lock and a journal make host registration recoverable, a failed or
   interrupted update restores the previous registration, and rollback uses a
   retained verified version without network access. Uninstall removes only
   owned registrations and unreferenced versions.
6. `update --all` touches only already-managed hosts, preserves pins, reports
   per-host results, and treats an already-current host as a no-op.
7. Tests run with disposable homes, fake `codex`/`claude`/`gh` executables, and
   fixture release servers; `make ci` passes from a clean scratch checkout.
   Real host adapters are exercised only in an explicitly isolated profile.

## Risks and mitigations

- A compromised network path: fixed repository, HTTPS-only allowlisted hosts,
  bounded sizes, digest plus attestation binding to the exact source commit.
- Host command drift: minimum host versions are checked, every mutation is
  verified through the hosts' own list commands, and failures restore the
  previous registration from the journal.
- Partial state: atomic writes, digest-bound state and receipts, and fail-closed
  loading; the next explicit mutation attempts recovery before proceeding.
- Overstated trust: the CLI repeats the release's `NOT_RUN`/`WITHHELD` facts in
  status and consent errors and never claims signing, notarization, or support.

## Rollback

Revert this change normally before publication. Installed users keep working
installations: the per-host `previous` version is retained for
`l7 plugin rollback`, and the hand-run host commands in the README remain valid
for removal. No immutable release asset, tag, or hosted setting is affected.
