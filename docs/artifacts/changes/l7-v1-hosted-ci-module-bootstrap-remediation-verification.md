# Level 7 v1 Hosted CI Module Bootstrap Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-ci-module-bootstrap-remediation` |
| Candidate commit | `bb8ce70a236d2122117c03c7be34ff796e3bf0f4` |
| Candidate tree | `65b12fd2d826422cd5854bed166c2cc3c4412f6b` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-31T19:40:40Z` |
| Brief commit | `68fe20ef599b5e3bf92670da7ea76cd111665454` |
| Trigger | PR #15, Harness run `33427651169` |
| Implementation commits | `10dda82c52c443009a17e9b72d27bd6f99f41194`, `bb8ce70a236d2122117c03c7be34ff796e3bf0f4` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the external active-user approval envelope names Product Owner Anup Pandey, implementer `codex-root`, this change ID, and exact brief commit `68fe20ef599b5e3bf92670da7ea76cd111665454`. Team policy selected Tier 3 and accepted exact candidate `bb8ce70a236d2122117c03c7be34ff796e3bf0f4`, tree `65b12fd2d826422cd5854bed166c2cc3c4412f6b`, in `building` state. |
| Exact scope | PASS — base `f10486aee9b361f9b2ec7e92782f30247b7b7a32` through the candidate changes exactly five paths: the approved brief and four declared implementation paths. The verification record was absent while tests ran; the declared audit record remains absent. |
| Dependency identity | PASS — `go.mod` and `go.sum` have no base-to-candidate diff and retain blobs `ca94980ebe74be3c6cd03a58d5e6904c99636a57` and `9db7fe4d2dbd9289aabc0bef42e2a8f5c2f3c64e`. Bootstrap fails if either input changes type or bytes. |
| Network-free bootstrap regression | PASS — `make bootstrap-modules-check GO_VERSION=1.26.7` exercises the exact three Go commands plus six negative fixtures with fake tooling and no network. It verifies the fixed proxy/checksum environment, downstream offline environment, ambient credential/proxy removal, disposable HOME cleanup, failure propagation, missing/symlinked inputs, argument rejection, and dependency-file immutability. |
| Empty-cache authenticated bootstrap | PASS — an archived exact-candidate checkout began with an empty repository-local module cache. The pinned Go 1.26.7 archive signature reauthenticated, the fixed HTTPS Go proxy and public checksum database populated only the locked module, `go.mod`/`go.sum` stayed byte-identical, and subsequent `go mod verify` passed with `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, and `GOAUTH=off`. |
| Clean-clone hosted-sequence simulation | PASS — a clean detached local clone at the exact candidate began without a module cache, ran `make bootstrap-ci GO_VERSION=1.26.7`, then ran the unchanged `make ci GO_VERSION=1.26.7` offline. Toolchain authentication, module priming, lint, vet, compile, unit, race, eight fixed fuzz targets, reproducibility, and frozen distribution compatibility all passed. This is local simulation, not hosted evidence. |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | PASS — exact-head policy, module integrity, import/effect boundaries, formatting, shell syntax, bootstrap regressions, vet, type and actual-provider compile-only checks, complete unit tests, serialized CGO race coverage, all eight fixed five-second fuzz targets, reproducible harness/CLI builds, and frozen v0.1.1 distribution compatibility passed. |
| `L7_ASSURANCE_MODE=team make v1-candidate-check GO_VERSION=1.26.7` | PASS — macOS arm64/amd64 binaries and both host archives reproduced; native arm64 Codex and Claude CLI/MCP, disposable install, stable upgrade, exact rollback, removal, cleanup, and OS-network-denied conformance passed. The packages remain unsigned and release-blocked. |
| Workflow boundary | PASS by inspection and regression — Linux baseline/shadow, macOS arm64/amd64, and benchmark jobs use the same `bootstrap-ci` entry point. Harness permissions remain read-only, checkout credentials remain disabled, actions remain digest-pinned, and no secret, cache action, write permission, or `pull_request_target` trigger was added. |
| Historical hosted state | PRESERVED — PR #15 remains open at old head `f10486aee9b361f9b2ec7e92782f30247b7b7a32`; no push, label, review, rerun, or PR edit occurred. Run `33427651169` remains the truthful failed historical run. |
| Trusted-policy lineage | BLOCKED separately as declared — a local exact-controller invocation with PR base `84bd69f90d366356b0ce1e1a392f906258f3de91` still returns `GIT-003` because this brief declares later base `f10486aee9b361f9b2ec7e92782f30247b7b7a32`. No policy code, scope rule, base identity, or finding was changed or waived. |
| User-owned state | PASS — the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged with SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

The first implementation candidate
`10dda82c52c443009a17e9b72d27bd6f99f41194` used `go mod download all` in an
empty cache. It failed closed because Go attempted to add unlocked transitive
sums to `go.sum`; the immutability guard rejected the change. No verification
record was created for that candidate. Additive commit
`bb8ce70a236d2122117c03c7be34ff796e3bf0f4` narrowed priming to the same locked
direct-module operation used by `install`, retained offline verification, and
passed every final check above.

## Reproducible identities

| Output | SHA-256 |
|---|---|
| `.github/workflows/harness.yml` | `2643cd436e7e45d769325a161d388b3752f883fcfb0d25817294f52e6effac90` |
| `Makefile` | `bc99b94b10c19d7933de05de610ee4ea20865a89230fa17027ba4863a7074223` |
| `scripts/harness/bootstrap-modules.sh` | `47e96bda11a36141e996ff71242bd9e3e6b7936172d417732bd583b06a4c691a` |
| `scripts/harness/check-bootstrap-modules.sh` | `96367adb73d1d2aa323bc359cda8386a404a773ac110d3cf3682673a5ed2670d` |
| Harness test binary | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Reproducible native CLI | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| `level7-dev-loop-1.0.0-dev-codex.zip` | `ce6ead70d4cbec718c2737bddcd0abe7e5cc984b81549f8188af5008c7b4f1fd` |
| `level7-dev-loop-1.0.0-dev-claude.zip` | `d7ae491f869ee0346a5437500272d26fd90a472c494d141ea129d6addab4df3d` |
| Frozen `level7-dev-loop-codex-0.1.1.zip` | `58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba` |
| Frozen `level7-dev-loop-claude-0.1.1.zip` | `0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49` |

## Platform and evidence boundary

| Evidence | Status |
|---|---|
| macOS 26.5.2 arm64, Go 1.26.7, Swift 6.3.3 native execution | PASS |
| Darwin amd64 cross-build, archive identity, and reproducibility | PASS |
| Darwin amd64 native execution | `NOT_RUN` |
| Hosted CI execution on the new candidate | `NOT_RUN` |
| Go 1.27.0 shadow execution on the new candidate | `NOT_RUN` |
| Live Codex, Claude, gateway, or Cyber provider calls | `NOT_RUN` |
| Real Codex/Claude installation, signing, publication, release, or deployment | `NOT_RUN` |

This record is implementer-run verification, not an independent audit, hosted
result, owner GO, merge authorization, or release authority. It makes no claim
that the separate trusted-policy lineage blocker is closed.

## Next executable transition

Stop for separate Product Owner authority to commission one independent
read-only audit of exact verification-record successor commit/tree by a reviewer
distinct from Product Owner Anup Pandey and implementer `codex-root`. Push, PR
mutation, hosted execution, review, owner GO, merge, signing, release,
publication, installation, and deployment remain unauthorized. The declared
`GIT-003` lineage blocker requires a separately scoped and approved correction.
