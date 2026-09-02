# Level 7 v1.0 Hosted amd64 CI Time-Budget and Release-Lineage Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation` |
| Risk tier | `3` |
| Approved brief commit | `4d56fe6729d679da3002ca86572c5e7d5838fb85` |
| Base commit | `c634e092b2f938ad3038be0632d162b2bdde41f3` |
| Base tree | `4a28b3ec2495566554cda8ab2462b3b41043b474` |
| Candidate commit | `333ed7abec9d34eacc9cc8481b2194f4506db87a` |
| Candidate tree | `fa09d1581fcea377cedee133635e9ce534ddf682` |
| Result | `PASS` |
| Verification scope | Local technical verification; pinned bootstrap network only, downstream verification offline |
| Reviewer | `codex-root` (implementer-run verification) |
| Verified at | `2026-09-02T07:06:17Z` |
| Local host | `darwin/arm64` (macOS `26.5.2`) |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Authorization, lineage, and scope | PASS — Product Owner Anup Pandey's active-user approval binds implementer `codex-root`, this change ID, approved brief `4d56fe6…`, base `c634e09…`/tree `4a28b3e…`, and the two implementation paths. The schema-1 external approval envelope is a regular, non-symlink file with SHA-256 `e2f6a6a5bfc8694c7ee3bb28fe2e02f3447b00e177de7900b45493ba1d1431fb`. The candidate is the brief's direct child; its only paths are `.github/workflows/harness.yml` and `.github/workflows/release.yml`. Base-to-candidate changes are those two paths plus the brief, with no historical-record rewrite |
| Workflow controls | PASS — `actionlint` passed both workflows. Harness changes only the macOS timeout selection: the existing matrix has literal `arm64=15` and `amd64=25`, and the job consumes that value. The release workflow changes only `L7_RELEASE_BASE` from `66777352538a514b990ffca8fa290ca6de9584fd` to `c634e092b2f938ad3038be0632d162b2bdde41f3`. Runners, job/check names, steps, commands, targets, ordering, fuzz and benchmark controls, actors, topology, reviews, environments, signing, provider, asset, and publication gates remain unchanged |
| Original prerequisite failure | PRESERVED — the first exact `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` invocation exited `2` immediately at the Makefile toolchain check with `missing pinned Go 1.26.7; run: make bootstrap GO_VERSION=1.26.7`. No test, build-control evaluation, or product build started; only a 4 KiB ignored cache scaffold was created. The failure was not retried or represented as a technical result without fresh Product Owner authority |
| Authorized storage recovery | PASS — only 38 individually prevalidated, Git-ignored, untracked, non-symlink, regenerable cache directory leaves were permanently removed: 20 obsolete generator/reproduction scratch leaves (16 old-worktree leaves and four original-checkout Wave 5 temporary leaves), 11 old-worktree Go build/race leaves, and seven old-worktree pinned Go `1.26.7` toolchain duplicates. The first 31 leaves produced `5,350,540 KiB` free, still below the then-authorized 6.5 GiB floor, so work stopped. After separate authority, the seven toolchain duplicates produced `7,014,060 KiB` free. No registered worktree, ref, tracked/index byte, protected file, source, release artifact, or user file was removed. These deleted caches are recoverable only by regeneration |
| Pinned bootstrap | PASS — exactly one `make bootstrap-ci GO_VERSION=1.26.7` ran from `2026-09-02T06:46:09Z` through `06:46:24Z` and exited `0`. The authenticated Darwin/arm64 archive was 64,772,572 bytes with SHA-256 `020a1e8224811be75163e920bc77e0926a1390a6aeea19bdcf23f74b9d749f6d`; its detached signature SHA-256 was `58dc70df936443877736e795a5861538189b6c576fd96d6f82423a98a5301ff6`. Emitted `google-linux-signing-key.pub` has SHA-256 `54dea5f6c2a26091578cf52a999cebc6b64df478d37ad4dce96376b711e3b27c`; the accepted primary fingerprint is `EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796` and signing subkey is `0E225917414670F4442C250DFD533C07C264648F`. The module cache was hydrated through the pinned bootstrap and verified for subsequent offline use; no partial, staging, temporary, or symlink residue remained |
| Evidence-path typo and resumed preflight | PRESERVED/PASS — the first read-only post-bootstrap evidence shell exited `1` solely because `codex-root` requested nonexistent `linux_signing_key.pub` instead of the emitted `google-linux-signing-key.pub`; its preceding free-space and Go-version checks had passed. It did not mutate the candidate or start verification. Work stopped; independent corrected read-only integrity evidence then confirmed the bootstrap and candidate, and fresh Product Owner authority preceded the final bounded preflight and verification. That preflight confirmed candidate `333ed7a…`/tree `fa09d15…`, the exact approval envelope, a clean tracked/index state, and `6,613,288 KiB` free, above the authorized 6.1 GiB floor. Therefore the optional original-checkout `.cache/toolchains/go1.27.0` cache was not deleted |
| Final `make verify` | PASS — exactly one separately authorized replacement `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` ran from `2026-09-02T06:59:26Z` through `07:05:44Z` (6m18s) and exited `0`. The controller bound Tier 3/team state `building` to base `c634e09…`, head `333ed7a…`, tree `fa09d15…`, and three base-to-candidate paths. Pinned toolchain/module integrity, three host/arm64/amd64 import closures, 32 package boundaries, bootstrap transport/module negative fixtures, formatting and static checks, typecheck, actual-host compilation, unit and race tests, all eight fixed fuzz targets, harness/CLI reproducibility, and distribution passed without target reduction or retry |
| Fuzz and deterministic boundaries | PASS — strict configuration and provider terminal each completed literal `10000x`; the other six fixed targets completed five-second budgets. The fixed eight-target inventory, offline toolchain, disposable corpus roots, and direct failure propagation remained intact. Harness A/B and CLI A/B outputs were byte-identical |
| `make v1-candidate-check GO_VERSION=1.26.7` | PASS — exactly one invocation ran only after the full verification passed, from `2026-09-02T07:05:52Z` through `07:06:17Z` (25s), and exited `0`. Darwin arm64/amd64 binaries and both host packages reproduced; native arm64 Codex and Claude lifecycle, CLI, MCP, stable upgrade, rollback, removal, and disposable-root checks passed. Both archives identify `1.0.0-dev`, `development-candidate`, `release_blocked=true`, and `authority=external-only`; unsigned release remained blocked |
| Workspace and external boundary | PASS — the exact implementation candidate's tracked worktree and index remained clean through verification; this record was then the sole authorized ordinary-untracked addition. Ignored outputs are confined to `.cache/` and `build/`; no generated file is tracked, ordinarily untracked, or symlinked. The original checkout's unrelated foundation audit remains untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. Live remote `main` remains `c634e09…`; no remediation remote branch, PR, or associated Actions run exists, and the `v1.0.0` tag and GitHub release remain absent. No push, hosted run, PR mutation, tag, release, publication, installation, or deployment occurred |

## Reproducible identities

| Output | Size (bytes) | Architecture | SHA-256 |
|---|---:|---|---|
| `.github/workflows/harness.yml` | — | text | `a11cfeb74a99c5ad6fc9b41577440036ef867127b2469d87a74499fec13f5b88` |
| `.github/workflows/release.yml` | — | text | `f46e9f018f6e71ced0836ce5c7cbd99e4b40ff55c7e8c2c3a1c05b31a61d6542` |
| Reproducible harness test binary A/B | `4,732,658` each | Darwin arm64 | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Reproducible CLI A/B | `18,939,762` each | Darwin arm64 | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| `build/v1-inputs/darwin-arm64/l7` | `18,939,762` | Darwin arm64 | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| `build/v1-inputs/darwin-arm64/l7-embed` | `94,600` | Darwin arm64 | `898db0787c397cb15508e0f91b8081f458522cb3734abd8a1bfdc953b920bc3b` |
| `build/v1-inputs/darwin-amd64/l7` | `20,351,392` | Darwin x86_64 | `412a1d09f657066aab55fdab33c4b111e1a64f3cdbf731ed11b00cc45c5babb4` |
| `build/v1-inputs/darwin-amd64/l7-embed` | `56,624` | Darwin x86_64 | `c7c33ff35e05ad2b5a037374276c01d8c610401e7300136db8c08e118884ef7d` |
| Codex `level7-dev-loop-1.0.0-dev-codex.zip` | `22,097,080` | arm64 + amd64 package | `2553f635039021c5709bf69f6b4ea16a60e20f732b41184a6b365fa10a871b2` |
| Claude `level7-dev-loop-1.0.0-dev-claude.zip` | `22,096,887` | arm64 + amd64 package | `35eea6c532027819ecbde77c4d54d8ccd292af4e8d2f2fb4546e19d18a1491e8` |
| Codex v0.1.1 distribution package | — | host package | `58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba` |
| Claude v0.1.1 distribution package | — | host package | `0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49` |

Both candidate ZIPs passed separate bounded read-only archive integrity checks. Every embedded
binary size and hash matched its source input and `CHECKSUMS.json` entry.

## Verification boundary

| Boundary | State |
|---|---|
| macOS amd64 native execution | `NOT_RUN`; amd64 binaries were cross-built, structurally inspected, and reproduced only |
| Hosted Harness, paired benchmark, trusted policy, PR, reviews, and accountable-owner GitHub envelopes | `NOT_RUN`; local verification does not replace exact-head hosted evidence |
| Benchmark-regression acceptance | `NOT_RUN`; no exception was requested or accepted |
| Actual Codex and Claude host/provider trials, model execution, and provider network | `NOT_RUN` |
| Apple credentials, signing, hardened-runtime verification, and notarization | `NOT_RUN` |
| Push, merge, release workflow dispatch, tag, GitHub release, publication, production installation, or deployment | `NOT_RUN` |

This implementer-run local `PASS` establishes only that the exact implementation
candidate meets the authorized technical boundary. It is not an independent
audit, hosted readiness decision, owner GO, merge authority, release authority,
or publication authority. Any implementation-byte change invalidates this
record.

## Rollback and next transition

No external release state exists from this verification. Before audit, an
authorized rollback reverts this record's commit, then implementation commit
`333ed7abec9d34eacc9cc8481b2194f4506db87a`, then brief commit
`4d56fe6729d679da3002ca86572c5e7d5838fb85`, and confirms exact base tree
`4a28b3ec2495566554cda8ab2462b3b41043b474`; history is preserved. After audit,
its record is reverted first.

The only next Level 7 transition is separately authorized commissioning of
`apbusinessidentity-tech` for one independent read-only Tier 3 audit at
`docs/artifacts/changes/l7-v1-hosted-amd64-ci-time-budget-release-lineage-remediation-audit.md`,
bound to this verification-record successor commit and tree. This record grants
no such authority.
