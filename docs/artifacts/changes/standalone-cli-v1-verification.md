# Standalone CLI v1 Wave 1 — Verification

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1` |
| Candidate commit | `5e65b272e454e4b446847453ade2f042d9d631d8` |
| Candidate tree | `24b4c277f2b9c51f247941267c2203a1063b0072` |
| Result | `PASS` |
| Reviewer | `local-verifier` |

## Checks

| Check | Result |
|---|---|
| `make policy-check` | PASS — Tier 3 `building`; 13 base-to-candidate paths, all within the approved Wave 1 scope |
| `make verify` | PASS — offline module verification, policy, eight-package import boundaries, formatting, shell syntax, vet, typecheck, all tests, and repeat-build checks |
| CLI unit/application/presentation tests | PASS — help/version/status, text/JSON, malformed and bounded arguments, cancellation, unknown commands/flags, short/failed writes, writer-independent output, truthful unavailable state |
| Import-boundary fail-closed probe | PASS — a temporary forbidden `os` import in `internal/l7/app` was rejected by `BND-602`; the probe was removed and the clean policy passed again |
| CLI smoke test | PASS — help and JSON version return `0`; unavailable status returns `2` and cannot claim working lifecycle behavior |
| `make cli-cross-build` | PASS — macOS arm64 and amd64 binaries built with the pinned Go 1.26.7 toolchain |
| `actionlint .github/workflows/harness.yml .github/workflows/policy.yml` | PASS |
| `jq empty plugin.json .codex-plugin/plugin.json .claude-plugin/plugin.json marketplace.json` | PASS |
| `git diff --check 2c0e1c97c0344e423a75b01fa3d1a0dc423a2b9d..HEAD` | PASS |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| Host macOS arm64 `l7` binary | `d0db0728b261ed5b7023f3e0b5bdc7ca003da8902e7068075fc7456740236c9b` |
| Cross-built macOS arm64 `l7` binary | `d0db0728b261ed5b7023f3e0b5bdc7ca003da8902e7068075fc7456740236c9b` |
| Cross-built macOS amd64 `l7` binary | `87446ec4289785a3495d1f6f6178317d1c70c7c462ac8f6a8328f4a83fce3fb9` |

The arm64 host and cross-built identities match. The amd64 result is build
evidence only; Wave 1 does not claim an amd64 runtime test, provider
compatibility, release readiness, or product lifecycle capability. This record is
verification evidence, not owner approval or independent-audit authority.
