# Level 7 v1.0.0-dev Explicitly Unsigned (No Developer ID) Prerelease — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-explicitly-unsigned-prerelease` |
| Risk tier | `3` |
| Approved brief commit | `92d750c2d3be83321f65c3ed9c9f0ce4f9dc50e7` |
| Base commit | `5c23038e38a07b4f91f8ef38bbf163e061857910` |
| Base tree | `240bc9509a02a1a71616b62e88068ed7783bc65b` |
| Original implementation commit | `71a4b7cd60ac6545852dcb2156897d42411a8025` |
| Lint remediation commit | `f3eb24f592a77791cab3c55c485a0924be89ad5b` |
| Candidate commit | `1a9af15ea59c7f9243e47e5c9b3504380fae456f` |
| Candidate tree | `a72a97c0135500c3980aab9d901a5c7171009e87` |
| Result | `PASS` |
| Verification scope | Local Tier 3 technical verification; authenticated pinned-bootstrap network only, with final verification offline and no hosted or release mutation |
| Reviewer | `codex-root` |
| Completed at | `2026-09-02T17:26:07Z` |
| Local host | `darwin/arm64` (macOS `26.5.2`) |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Authorization, lineage, and scope | PASS — active-user Product Owner approval binds Anup Pandey (`addressanup`), implementer `codex-root`, change `l7-v1-explicitly-unsigned-prerelease`, and approved brief `92d750c...`. Candidate `1a9af15...` / tree `a72a97c...` descends without rewriting from exact base `5c23038...`. Relative to the approved proposal, its implementation diff contains exactly `.github/workflows/release.yml`, `.github/workflows/unsigned-prerelease.yml`, `CHANGELOG.md`, `README.md`, and `docs/releases/v1.0.0-dev.md` |
| External approval envelope | PASS — `.git/l7/approvals/l7-v1-explicitly-unsigned-prerelease.json` has schema `1`, the exact actor, implementer, change, brief, and `active-user-interaction` binding, at SHA-256 `7c12edd386c7423a58d0f8c7394c1516200358ecc8c2bf12a675b1568c3fcc3a` |
| Historical failure preservation | PASS — the first verification FAIL remains recoverable at commit `a27f65410568b1dd3f3ba1b0156958cf5a2943c0`, record SHA-256 `5a434102108b992cc591318748c6fa8d68b7b90ec2655ae28789ba26c8740eb0`. The evidence-wrapper FAIL remains recoverable at commit `517474c2a0410cd3d071dba500d35d5de89f3d93`, record SHA-256 `ed9a176299944bd86db420065c4b13a0cf4da36d7b12719fff8047b36318c3a5`. Candidate `1a9af15...` removed both failed record paths from its tree in one two-deletion commit so this canonical PASS record is added after the tested candidate. Both failure commits remain ancestors; no reset, rebase, amend, or history rewrite occurred |
| Pinned bootstrap | PASS — the sole `make bootstrap-ci GO_VERSION=1.26.7` ran from `2026-09-02T16:11:34Z` through `16:11:50Z` and exited `0`. GPG verification accepted Google signing subkey `0E225917414670F4442C250DFD533C07C264648F`; repository-local Go `1.26.7 darwin/arm64` and the offline module cache were established. Bootstrap was not rerun |
| Workflow lint chronology | PASS — the original candidate's sole lint run failed with `SC2129` and four `SC2016` findings and is preserved at `a27f654...`. Remediation `f3eb24f...` changed only output-redirection grouping and shell-safe format-string quoting, with emitted handoff bytes unchanged. A later lint evidence wrapper failed because zsh parameter `status` is read-only; that exact failure is preserved at `517474c...`. One separately authorized replacement `actionlint .github/workflows/release.yml .github/workflows/unsigned-prerelease.yml` ran from `2026-09-02T16:45:00Z` through `16:45:01Z`, emitted no findings, and exited `0`. It was not rerun |
| `git diff --check` | PASS — one read-only `git diff --check 92d750c2d3be83321f65c3ed9c9f0ce4f9dc50e7..HEAD --` exited `0` during the bound inspection; no timestamp was recorded. Candidate `1a9af15...` subsequently changed only by deleting the two historical Markdown record paths; all five implementation blobs remained exact. The command was not rerun |
| Semantic workflow verification | PASS — read-only review found the five implementation files aligned with the approved brief: manual-only fixed `v1.0.0-dev` identity, full action pins, exact actor/ref/tree/run-attempt and environment controls, closed arm64/amd64 and asset inventories, read-only unsigned signature classification, no Apple/signing/notary fallback, exact attestation/handoff/publication gates, truthful unsigned documentation, and only the authorized stable `L7_RELEASE_BASE` replacement. The lint remediation is output-byte equivalent |
| Superseded policy-only full-verification attempt | PRESERVED — one earlier `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` ran from `2026-09-02T17:02:41Z` through `17:02:47Z` and exited `2` before technical tests because the then-current tree contained an out-of-budget remediation record and a stale FAIL record (`ART-003`, `SCOPE-002`, `VERIFY-002`). That result caused candidate `1a9af15...`; it was not treated as technical success or silently retried against the same candidate |
| Full local verification | PASS — exactly one replacement `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` ran against candidate `1a9af15...` from `2026-09-02T17:13:36Z` through `17:21:00Z` and exited `0`. Build control reported Tier 3 team mode, exact base/head/tree, `state=building`, and six changed paths. Module integrity, import closures and boundaries, bootstrap fixtures, formatting, vet, type compilation, full tests, race checks, eight fixed fuzz targets, binary reproducibility, CLI reproducibility, and distribution checks passed. Reproducible harness-test SHA-256 was `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70`; reproducible arm64 CLI SHA-256 was `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| Authorized storage recovery | PASS — earlier exact authority removed only the `7,374,760 KiB` Xcode DerivedData leaf `laalee-goneifujaobgbbfdzbylbzthnzit` and the `571,220 KiB` Xcode `ModuleCache.noindex`; both are regenerable. After full verification, free space again fell below the `6,396,314 KiB` candidate-check floor and the candidate check remained unused. Separate exact authority then removed only the two completed reproducibility roots (`380,680 KiB` and `149,264 KiB`) and `.cache/go/build` (`691,528 KiB`). All three were Git-ignored, regenerable, non-symlink directories with no nested Git repository or open file. Their absence was verified and free space became `6,978,588 KiB`; the pinned toolchain and module/download caches were preserved. No other cleanup occurred |
| v1 candidate check | PASS — the sole `make v1-candidate-check GO_VERSION=1.26.7` ran from `2026-09-02T17:25:32Z` through `17:26:07Z` and exited `0`. Both clean Darwin arm64 and amd64 binaries, both Codex/Claude `1.0.0-dev` packages, native CLI/MCP lifecycle conformance, stable upgrade/rollback/removal in disposable roots, and a second full binary/package reproduction passed. The command explicitly retained the unsigned-release block |
| Workspace and external boundary | PASS — candidate commit/tree and tracked cleanliness remained exact through verification. Generated outputs stayed in ignored `.cache/` and `build/`. The primary checkout's unrelated `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. No audit, push, PR, hosted run, review, merge, repository setting, environment, secret, dispatch, attestation, tag, release, publication, production installation, or deployment occurred |

## Candidate file identities

| Path | SHA-256 |
|---|---|
| `.github/workflows/unsigned-prerelease.yml` | `41cab4094f16ac305146e9f63532843147d2b6dc05f8f7d98618708d7f2f7924` |
| `docs/releases/v1.0.0-dev.md` | `9a288d082f50e810b6e0834e473d510451182b1f6ca17d9d7f75b2a7702a7fa9` |
| `.github/workflows/release.yml` | `669bf0db8c400f2a543c76a256dd048f1202327561c3f9757aacaffde6883f51` |
| `CHANGELOG.md` | `45a2272ec68372c958ec64478b7c1d59fc3e218270bc4b12d9abfc4c7a3d5614` |
| `README.md` | `80469a189a7fdc98e24f63aa8513d159c49f67dd104196d72596d6b68ed12b29` |

## Verified local build identities

| Path | Size | SHA-256 |
|---|---:|---|
| `build/v1-inputs/darwin-amd64/l7` | `20,351,392` | `412a1d09f657066aab55fdab33c4b111e1a64f3cdbf731ed11b00cc45c5babb4` |
| `build/v1-inputs/darwin-amd64/l7-embed` | `56,624` | `c7c33ff35e05ad2b5a037374276c01d8c610401e7300136db8c08e118884ef7d` |
| `build/v1-inputs/darwin-arm64/l7` | `18,939,762` | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| `build/v1-inputs/darwin-arm64/l7-embed` | `94,600` | `898db0787c397cb15508e0f91b8081f458522cb3734abd8a1bfdc953b920bc3b` |
| `build/v1-candidate/level7-dev-loop-1.0.0-dev-codex.zip` | `22,098,273` | `1f1905e22798910e85e9dfc23a380c06725894e5c6366dd85f7666cf742d55e8` |
| `build/v1-candidate/level7-dev-loop-1.0.0-dev-claude.zip` | `22,098,082` | `3cfaaa52fccf464ce200e71025568a7c9a0e370390b24377e4539159f83b5b1d` |

These are local verification outputs, not prepared workflow artifacts or
published release assets. No release manifest, checksum file, attestation, or
GitHub artifact identity was created by this local sequence.

## Verification boundary

| Boundary | State |
|---|---|
| Corrected workflow lint, diff hygiene, and semantic controls | `PASS` |
| Full local Tier 3 verification | `PASS` |
| Darwin arm64/amd64 reproducible binaries and host packages | `PASS` |
| Local native Codex/Claude package lifecycle and CLI/MCP conformance | `PASS` |
| macOS amd64 native execution | `NOT_RUN`; amd64 binaries were cross-built, structurally validated, and reproduced only |
| Actual post-preparation Codex/Claude host trial and provider execution | `NOT_RUN` |
| Hosted Harness, benchmark, Trusted policy, PR, reviews, and merge | `NOT_RUN` |
| Prerelease workflow dispatch, signature classification, manifest, checksums, attestations, and artifact upload | `NOT_RUN` |
| Apple Developer ID signing and notarization | `NOT_PERFORMED` and outside this explicitly unsigned prerelease |
| Tag, GitHub prerelease, publication, production installation, or deployment | `NOT_RUN` |

This implementer-run `PASS` is bound only to candidate
`1a9af15ea59c7f9243e47e5c9b3504380fae456f` / tree
`a72a97c0135500c3980aab9d901a5c7171009e87`. It is not an independent audit,
hosted-readiness decision, owner GO, merge authority, release authority, or
publication authority. Any implementation-byte change invalidates it.

## Rollback and next transition

No remote or release state exists from this verification. Reverting this
verification-record successor returns to the exact tested candidate
`1a9af15...`. The two historical FAIL records remain recoverable from commits
`a27f654...` and `517474c...`; reverting candidate `1a9af15...` restores their
current-tree presence without rewriting history.

The only next Level 7 transition is a separately commissioned independent
read-only Tier 3 audit bound to this verification-record successor commit and
tree. This record grants no audit, push, PR, merge, release, publication,
installation, deployment, or cleanup authority.
