# Level 7 v1.0.0 Stable Release — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-stable-release` |
| Risk tier | `3` |
| Approved brief commit | `19ab613894a0b8ae4a8524d208477f0388ec7d28` |
| Base commit | `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` |
| Base tree | `9b6550c11e9666ab25166af6a5fb7560cdcc8cdf` |
| Candidate commit | `8911e179f95165347946214420f93dbfa9a9f7dc` |
| Candidate tree | `1a2b5fa792def7950cf974d7f56f3c11f36f257d` |
| Result | `PASS` |
| Verification scope | Local offline technical verification only |
| Reviewer | `codex-root` (implementer-run verification) |
| Verified at | `2026-09-01T13:42:41Z` |
| Local host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Authorization, lineage, and scope | PASS — active-user authorization names the exact candidate commit/tree and permits only this verification record. The candidate is the direct child of the approved brief and changes exactly the nine declared non-evidence paths: two additions and seven modifications, with no deletion or historical-record rewrite |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | PASS from the clean exact candidate — the controller truthfully reported pre-record state `building`; offline module integrity, import/effect boundaries, formatting, shell parsing, vet, compilation, the full unit and race suites, all eight fixed fuzz targets, and harness/CLI reproducibility passed without target reduction or retry-until-green |
| Focused package and validator coverage | PASS — exact identity-pair validation, version/architecture substitution rejection, mixed-identity rejection before output or host access, host-local catalog binding, release-authority exclusion, strict archive/metadata parsing, closed inventory, offline execution, disposable lifecycle, upgrade/rollback/removal, and symlink/owned-file containment are exercised by the passing suites |
| `make v1-candidate-check GO_VERSION=1.26.7` | PASS — ordinary `1.0.0-dev` Codex and Claude candidates passed native arm64 CLI/MCP lifecycle conformance; Darwin arm64/amd64 binaries and both host archives reproduced byte-for-byte; unsigned release remained blocked |
| Explicit stable package validation | PASS — `make distribution`, `make v1-inputs`, `make v1-package`, and `make v1-package-check` with exact `L7_CLI_VERSION=1.0.0` and `L7_PACKAGE_CHANNEL=stable` validated both stable archives, native arm64 CLI/MCP behavior, v0.1.1 upgrade/rollback/removal, and disposable cleanup offline |
| Stable unsigned reproducibility | PASS — an additional `make v1-candidate-check` invocation with the exact stable identity rebuilt Darwin arm64/amd64 inputs and both stable host archives byte-for-byte. Its conformance-script substep selected the ordinary development archives by default; stable lifecycle evidence therefore comes only from the separately parameterized stable package check above |
| `actionlint .github/workflows/release.yml` | PASS — local static validation only; the manual workflow was not dispatched |
| Workflow and release-boundary inspection | PASS — the candidate retains manual-only execution, exact merged-main/check/review gates, two-build unsigned comparison, disposable signing state, four-binary signature validation, exactly one notarization submission per host, exact asset inventories/digests/attestations, a separate owner-gated production job, exact provider-trial evidence binding, and fail-closed refusal of existing tag/release state. It contains no benchmark exception, automatic trigger, signing fallback, or candidate-controlled release authority |
| Workspace hygiene | PASS — the exact candidate worktree and index were clean before this record. Ignored build roots were disposable verification output; no implementation byte, remote, PR, credential, tag, release, or hosted setting changed. The original checkout's unrelated foundation audit remained untouched and unstaged at its declared SHA-256 |

## Deterministic identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Reproducible native CLI / Darwin arm64 | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| Codex `1.0.0-dev` package | `2553f635039021c5709bf69f6b4ea16a60e20f7320b41184a6b365fa10a871b2` |
| Claude `1.0.0-dev` package | `35eea6c532027819ecbde77c4d54d8ccd292af4e8d2f2fb4546e19d18a1491e8` |
| Codex unsigned stable `1.0.0` package | `6f67333bc1117acee1a3fc44a1f1b819e9a131b1446f32902596c87ae8fb733c` |
| Claude unsigned stable `1.0.0` package | `9a0369ab68362fb29f9811d389fbf92f99141646f523145956f185941702351c` |
| Codex v0.1.1 distribution package | `58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba` |
| Claude v0.1.1 distribution package | `0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49` |

## Verification boundary

| Boundary | State |
|---|---|
| Actual Codex and Claude host/provider trials, model use, and provider network | `NOT_RUN` |
| macOS amd64 native execution | `NOT_RUN`; binaries were cross-built, structurally validated, and reproduced only |
| Apple credentials, signing, hardened-runtime verification, and notarization | `NOT_RUN` |
| GitHub environment configuration or approval, workflow dispatch, push, PR/review mutation, and fresh exact-head hosted checks | `NOT_RUN` |
| Paired benchmark gate or benchmark-regression acceptance | `NOT_RUN`; predecessor evidence does not transfer and no exception is accepted |
| Tag, GitHub release, publication, installation on real hosts, or deployment | `NOT_RUN` |

The local `PASS` establishes only that the exact implementation candidate meets
the authorized offline technical boundary. It does not establish provider
support, Intel runtime compatibility, signed/notarized asset authenticity,
hosted policy readiness, release readiness, owner GO, merge authority, or
publication authority. Any implementation-byte change invalidates this record.

## Rollback and next transition

No external release state exists from this verification. Before audit, an
authorized rollback reverts this record, then implementation commit
`8911e179f95165347946214420f93dbfa9a9f7dc`, then brief commit
`19ab613894a0b8ae4a8524d208477f0388ec7d28`, and confirms exact base tree
`9b6550c11e9666ab25166af6a5fb7560cdcc8cdf`; history is preserved. After audit,
its record is reverted first.

The only next Level 7 transition is separately authorized commissioning of
`apbusinessidentity-tech` for one independent read-only Tier 3 audit at
`docs/artifacts/changes/l7-v1-stable-release-audit.md`, bound to the exact
verification successor commit and tree. This record grants no such authority.
