# Standalone CLI v1 Wave 1 — Verification

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1` |
| Candidate commit | `34ba6284121374fdf5e8efc13a4872f29fa6ab0c` |
| Candidate tree | `19c711edffba61929995b4894e70e9807698dec4` |
| Result | `PASS` |
| Reviewer | `local-verifier` |

## Remediation lineage

This record supersedes the verification at commit
`2004817cd115c2997f13aeb78ea81ce243084065` after the independent `NO_GO` audit
recorded at `c9cc9f8`. Both prior records remain in Git history. The remediated
candidate closes `CLI-AUD-001` and `CLI-AUD-002`; no product capability was added.

## Checks

| Check | Result |
|---|---|
| `make policy-check` | PASS — Tier 3 `building`; 13 base-to-candidate implementation paths, all within the owner-approved Wave 1 scope |
| `make verify` | PASS — offline module verification, policy, formatting, shell syntax, vet, typecheck, 12 CLI/product tests, existing harness tests, and repeat-build checks |
| CLI import closure | PASS — domain, application, and presentation direct imports match explicit allowlists on the host, Darwin arm64, and Darwin amd64; local dependency closures permit only the domain package |
| Checked-in import-policy probes | PASS — the gate's non-domain repository and indirect-filesystem probes both exercise and confirm rejection on every `make import-check` run |
| Adversarial source probes | PASS — temporary blank imports of `internal/evaluator` and `io/ioutil` in `internal/l7/app` each failed `l7-import-closure-check`; both probes were removed with no candidate diff |
| Required-check contract | PASS — README names both non-experimental Harness jobs and Trusted policy as required installation checks and truthfully withholds any live blocking claim until a repository ruleset is verified |
| CLI smoke test | PASS — help and JSON version return `0`; unavailable status returns `2` and cannot claim working lifecycle behavior |
| `make cli-cross-build` | PASS — macOS arm64 and amd64 binaries built with the pinned Go 1.26.7 toolchain and report the expected Mach-O architectures |
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
evidence only; Wave 1 does not claim an amd64 runtime test, live repository-rule
installation, provider compatibility, release readiness, or product lifecycle
capability. This record is verification evidence, not owner approval or
independent-audit authority.
