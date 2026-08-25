# Level 7 Dev Loop — Foundation Harness

| Field | Value |
|---|---|
| Artifact ID | `L7-HAR-001` |
| Artifact type | Repository harness and local verification evidence |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Foundation step | 5 — Harness |
| Status | Candidate complete locally; separate-context audit `GO`; owner approval required before Step 6 |
| Version | 0.1.0 |
| Date | 2026-08-24 |
| Input decision | Approved [`L7-TEC-001`](technology-selection.md) 0.2.0 and [`L7-APR-TEC-001`](technology-selection-approval.md) |
| Approval-record SHA-256 | `62b8f52fe3075e99c0ca891bcf48d49f0ece067547de19008418a615bec6fac7` |
| Candidate | [`harness-candidate.sha256`](harness-candidate.sha256), 20 exact files |
| Candidate-manifest SHA-256 | `64bba1fcfe347d27a2b05df545b753bf7dc181383d99630493db4d7a47233592` |
| Audit | [`L7-AUD-HAR-001`](harness-audit.md) 0.1.0 — exact-manifest `GO` |
| Audit SHA-256 | `ff1616af337a8101fb2df53b026c65ff342a6d6587897f00973ccd476e99c445` |
| Effect | Bounded A2 repository harness plus A1 governance/evidence records; no product runtime or Codex/Claude host behavior; local verification effects are documented in §6 |
| Scope identity | Non-Git workspace snapshot; no commit, branch, remote, or hosted-CI identity exists |
| Next authorization if approved | Foundation Step 6 orchestration plan only |

## 1. Outcome

Foundation Step 5 is locally complete for the exact candidate manifest. It supplies a minimal, inert, zero-production-dependency Go harness with one proving test, authenticated repository-scoped compiler bootstrap, formatting/lint/type/test gates, live import-policy scaffolding, CI configuration, structured logging proof, an inert environment example, and an honest README.

No product command, controlled client, updater, runtime behavior, prompt, product/semantic workflow, skill, plugin/marketplace manifest, host package, provider call, actual-host experiment, deployment, publication, release, or autonomous/self-healing behavior was added. The approved §18 product tree remains a documented target rather than empty directories that could falsely imply implementation.

## 2. Actual Step 5 layout

```text
.github/workflows/harness.yml     # read-only-token baseline + allowed-to-fail shadow configuration
.env.example                      # documentation only; no secrets or authority
.gitattributes / .gitignore       # normalized source and ignored repository caches
.go-version / go.mod              # exact baseline; zero require directives
Makefile                          # bootstrap/install/lint/typecheck/test/repeat-build/CI gates
harness/                          # tool/action/identity/module/boundary/input locks
internal/harness/                 # package declaration + exactly one inert proving test
scripts/harness/                  # authenticated bootstrap and policy checks
docs/artifacts/                   # approval, candidate, audit, and this evidence record
README.md                         # prototype status, commands, boundaries, and limitations
```

`cmd/l7`, `cmd/l7up`, product `internal/*`, `semantic/`, `schemas/`, `fixtures/`, `packages/`, and `build/generated/` remain absent. `harness/modules.lock.tsv` marks the separately required updater module `reserved` with identity `UNSET`; creating its path fails until a later approved module-aware boundary harness exists.

## 3. Acceptance result

| Step 5 requirement | State | Evidence |
|---|---|---|
| Repository layout | `PASS` | Minimum harness only; protected scope checker rejects every reserved product path. |
| Dependency install | `PASS` | `make install`; zero modules to download; `go mod verify`; `go mod tidy -diff`; no `go.sum` or `vendor/`. |
| Formatting and lint | `PASS` | exact `gofmt`, `go vet`, POSIX `sh -n`, ShellCheck 0.11.0, policy and boundary checks. |
| Type/compile check | `PASS` | `go test -run '^$'` with exact version/OS/architecture link binding. |
| Tests | `PASS` | exactly one inert Go test on baseline and shadow. |
| CI scaffold | Configuration `PASS`; hosted run `NOT_RUN` | actionlint/YAML pass; automatic token restricted to read-only and not persisted, no user-configured secret, no `pull_request_target`, exact action commit, exact toolchain locks. No Git remote exists. |
| Logging | `PASS` for inert proof | one in-memory `slog` JSON record with fixed schema/status/effect/telemetry fields; timestamp removed; no persistent sink. |
| Environment example | `PASS` | not auto-loaded; no secret-shaped value, approval, authority, provider credential, or mutation toggle. |
| README | `PASS` | leads with prototype status and distinguishes local proof, missing hosted CI, and missing product/security/release proof. |
| Production dependency pins | `PASS` by empty graph | production module count is exactly zero; future selected modules remain uninstalled decisions. |

## 4. Frozen inputs

### 4.1 Go distributions

| Role | Version | Platform | Size | SHA-256 | Local result |
|---|---:|---|---:|---|---|
| Baseline | 1.26.7 | darwin/arm64 | 64,772,572 | `020a1e8224811be75163e920bc77e0926a1390a6aeea19bdcf23f74b9d749f6d` | archive, fingerprint, detached signature, version: `PASS` |
| Baseline | 1.26.7 | linux/amd64 | 66,890,901 | `ffb5f8de10c62550dfddab66b36b57030721e0a44a3218e9e1181d7b59f121ca` | locked; hosted Linux execution `NOT_RUN` |
| Shadow | 1.27.0 | darwin/arm64 | 68,303,667 | `90493b3bbd5e10f91d12153198bf1994fd756399b4fec93b49b0c6e2acdeeb3e` | archive, fingerprint, detached signature, version: `PASS` |
| Shadow | 1.27.0 | linux/amd64 | 70,523,269 | `675c26c449cbb18fc24b74650de1eabbae6e16f64326fd85a283fb3b58280685` | locked; hosted Linux execution `NOT_RUN` |

The values were rechecked against the official [Go downloads feed](https://go.dev/dl/). Bootstrap pins Google primary fingerprint `EB4C1BFD4F042F6DDDCCEC917721F63BD38B4796` and signing subkey `0E225917414670F4442C250DFD533C07C264648F`. The detached signature is useful defense in depth but uses Google's shared Linux Packages Signing Authority, not a Go-exclusive identity. A reused extracted tree remains same-user writable and is not complete gate-bearing evidence even though its archive/signature is reauthenticated.

### 4.2 CI action and module graph

- `actions/checkout` `v7.0.1` is fixed to commit `3d3c42e5aac5ba805825da76410c181273ba90b1`; the tag-to-commit mapping was rechecked against the official repository.
- No setup action, cache action, artifact upload, user-configured secret, or write permission is used. Checkout may use GitHub's automatic ephemeral token with `contents: read`; `persist-credentials: false` prevents it from being retained in the checkout.
- The provisional root module is `continuallabs.ltd/level7-dev-loop`. Domain control and vanity-module resolution are `UNPROVED` and must be owner-confirmed before product imports or publication.
- Production module count is zero. Staticcheck, gosec, govulncheck, Gitleaks, Syft, and Cosign were not provisioned or invoked by Step 5; Syft exists in the ambient host but is `NOT_RUN`. Future go-tuf, JSON Schema, JCS, and `x/text` dependencies were not added. Nothing is silently fetched as `latest`.

## 5. Local verification evidence

Observed host: macOS 26.5.2 build 25F84, Darwin 25.5.0, arm64. This is development evidence only and is not the selected controlled Ubuntu production tuple.

| Command/check | Result |
|---|---|
| `make bootstrap GO_VERSION=1.26.7` | `PASS`; official archive size/SHA-256/fingerprints/signature; exact darwin/arm64 binary. |
| `make bootstrap GO_VERSION=1.27.0` | `PASS`; same checks for the shadow archive. |
| `make install GO_VERSION=1.26.7` | `PASS`; zero dependency downloads, module verification, tidy diff. |
| `make lint GO_VERSION=1.26.7` | `PASS`; scope, import, format, shell syntax, `go vet`. |
| `make typecheck GO_VERSION=1.26.7` | `PASS`; compile-only test invocation. |
| `make test GO_VERSION=1.26.7` | `PASS`; one inert test. |
| `make reproducible GO_VERSION=1.26.7` | `PASS`; two invocation-private build caches, identical bytes. |
| `make ci GO_VERSION=1.26.7` | local command `PASS`; not a hosted workflow result. |
| `make verify GO_VERSION=1.27.0` | `PASS`; complete shadow gate. |
| simultaneous baseline + shadow `make verify` | both `PASS`; prior cross-version output race is corrected. |
| hostile inherited `GOROOT`/`GOOS`/`GOARCH`/`GOFLAGS` | normalized; exact root/tool/host/target assertions `PASS`. |
| ShellCheck + actionlint + YAML parse | `PASS`. |
| protected input digests + exact-one-test + zero-module assertions | `PASS`. |
| transient `cmd/l7up` reserved-path negative probe | correctly rejected by `BND-000`. |
| transient pure `internal/kernel -> os` negative probe | correctly rejected by `BND-004`; probe removed. |
| 20-entry candidate manifest | every file `PASS`; manifest digest independently rebound by the auditor. |

Same-machine repeat-build test-binary hashes are stable in the observed runs:

- Go 1.26.7: `1507927db3fb1508ce732e2f717b4e850e015140f8f956e12f713ad656a4032f`
- Go 1.27.0: `9edf1547c82d18a356c0d85d00c848bed9159f1461c469cf1b49ed3085b420c3`

These are local smoke evidence, not the two protected clean-builder proof required by `SP-11`.

## 6. Isolation, logging, and observed side effect

All Go verification commands bind the exact compiler root/tool directory, native OS/architecture, `GOTOOLCHAIN=local`, `GOENV=off`, `GOWORK=off`, `CGO_ENABLED=0`, repository-scoped caches/temp/telemetry, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, and `GOAUTH=off`. This closes Go's module/VCS resolution for the current zero-dependency command graph; it is not an OS network sandbox.

The sole test logs one fixed event into memory and validates exactly six fields. It reads no project file, prompt, provider credential, secret, user environment value, network resource, or persistent sink. `slog` internally obtains a clock value when creating the record, but the test removes it before encoding and neither persists nor asserts it.

During the first pre-correction verification attempt, `GOTELEMETRY=off` proved ineffective and Go created 14 local telemetry counter/marker files under `/Users/anuppandey/Library/Application Support/go/telemetry/local` between 22:17 and 22:18 local time. No upload-directory file was observed, but no network capture was performed. The files were not deleted or changed without separate user authority. The corrected harness redirects both exact compilers to repository-scoped telemetry with mode `off`; subsequent runs created no newer file in the ambient directory. This incident is preserved as evidence rather than hidden.

## 7. Boundary policy

- `internal/executor`, `transaction`, and `receipt` cannot transitively reach render, host adapters, or semantic prompt/workflow packages.
- `cmd/l7` cannot transitively reach the privileged channel.
- A root-module `cmd/l7up` is restricted to one entry package, standard library, and `internal/channel`; the reserved-module gate blocks the selected separate updater module until module-aware enforcement is approved.
- Future `internal/kernel`, `internal/policy`, and `semantic/*` closures reject filesystem/process/network/clock/randomness/runtime/syscall families and external-module detours.
- Product packages cannot import `unsafe` or the test-only harness.

The current positive scan covers one inert package, so these are live harness policies—not proof of future product architecture. Permanent negative fixtures and exact updater dependency/transitive allowlists are mandatory before the first applicable implementation.

## 8. Explicitly deferred / `NOT_RUN`

- Git initialization/import, branch protection, a conventional commit, remote CI, exact runner image identity, protected builders, and release roles;
- Linux/Ubuntu execution, race/fuzz/property/fault suites, vendoring, SBOM, license/notices, vulnerability/security/secret scanners, signing, and provenance;
- every `SP-01`–`SP-16` actual-host/platform/product acceptance spike;
- Codex/Claude installation, discovery, lifecycle, context, prompt/skill parity, provider/model, AP1, containment, relay, path, transaction, recovery, TUF, rollout, and postdeployment evidence;
- all product source trees and behavior; and
- self-healing/self-maintaining behavior, which remains a later governed capability rather than a Step 5 feature.

## 9. Owner gate

The product owner may approve this exact Step 5 candidate or request revision. Approval authorizes **Foundation Step 6 orchestration planning only**: waves, shared-file ownership, dependency order, evidence gates, and safe parallelism limits. It does not authorize product code, prompt/skill or manifest changes, dependencies, Git/repository publication, actual-host/provider experiments, deployment, exposure, release, or cleanup of the observed ambient telemetry files.

If approved, run the same `l7-greenfield` skill for Step 6. After the orchestration plan itself is approved, the next implementation skill will be `l7-build` for Wave 1.
