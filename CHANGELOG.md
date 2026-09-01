# Changelog

All notable changes to the Level 7 instruction plugin packages are recorded
here. Host observations and support claims remain scoped by
`distribution/compatibility.json`.

## 1.0.0

- Add explicit, fail-closed `1.0.0`/`stable` package identity while preserving
  `1.0.0-dev`/`development-candidate` as the ordinary default.
- Add self-contained, host-specific local marketplace catalogs to the Codex
  and Claude packages without changing the v0.1.1 catalogs or rollback bytes.
- Validate stable and development archives through the same closed inventory,
  checksums, SPDX SBOM, offline native CLI/MCP, upgrade, rollback, removal,
  path-safety, and disposable-root boundaries.
- Add a manual-only release workflow that compares clean unsigned inputs before
  signing, verifies four Developer ID signatures, requires two accepted Apple
  notarization submissions, and attests the exact prepared assets.
- Require fresh exact-head checks and reviews, exact-archive provider trials,
  a hosted digest-bound owner authorization, and a protected production
  approval before creating the absent annotated tag and immutable release.
- Document signed-asset verification, local marketplace installation,
  permissions, compatibility limits, update, uninstall, and v0.1.1 rollback.

## 1.0.0-dev

- Add a default-off, bundled macOS arm64/amd64 orchestration engine and local
  MCP bridge generated for Codex and Claude Code from one canonical source.
- Discover authenticated Codex app-server and Claude Code capabilities, probe
  configured OpenAI Responses/Anthropic Messages gateways, and persist
  explainable, fail-closed route decisions with effort escalation and reviewer
  separation.
- Add `l7-onboard`, `l7-sync`, `l7-cyber`, and `l7-headless` skills and CLI/MCP
  surfaces.
- Add private Git-bound codebase memory, isolated read-only-first security
  audits, and durable multi-wave Headless execution with crash recovery,
  provider failover, natural quota waiting, exact-ref local merges, and a hard
  stop before push, release, or deployment.
- Add reproducible unsigned development packages, checksums, SPDX SBOMs, and
  explicit release-blocked provenance. Stable `v1.0.0` is not published by
  this change.

## 0.1.1

- Add standard, tag-pinnable Git marketplace catalogs for Codex and Claude
  Code.
- Add one clean dual-host plugin payload so installation no longer requires a
  source build, ZIP extraction, or copying the development harness.
- Bind the committed marketplace payload to the canonical manifests and all 12
  source skills during offline distribution verification.
- Keep every Level 7 skill byte-for-byte unchanged from `v0.1.0`.

## 0.1.0

- Make `l7-next` a one-intent solo conductor instead of a skill router.
- Default protected repository work to truthful solo assurance with no mandatory
  independent auditor or evidence-only verification/audit commits.
- Retain opt-in team assurance and bind its audit identity to the actual forge
  reviewer before evaluation.
- Avoid duplicate feature-branch Harness runs and publish trusted policy after
  exact-head pull-request checks complete.
- Add deterministic, separate Codex and Claude development package assembly.
- Add offline package structure and reversible filesystem-lifecycle checks.
- Bind each marketplace catalog to its package identity and emit a canonical
  offline qualification result that cannot claim release readiness.
- Smoke-test installation, discovery, explicit `l7-next` invocation, and removal
  on Codex CLI 0.151.0 and Claude Code 2.1.247 on macOS arm64.
- Keep formal support withheld and signing/publication outside the offline
  qualification boundary.
