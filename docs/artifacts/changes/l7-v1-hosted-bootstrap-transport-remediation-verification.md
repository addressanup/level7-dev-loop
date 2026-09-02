# Level 7 v1.0 Hosted Bootstrap Transport Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-bootstrap-transport-remediation` |
| Risk tier | `3` |
| Approved brief commit | `e27676ddb4f8875cf9a88ff3c2ef2a26a85fdfa1` |
| Base commit | `66777352538a514b990ffca8fa290ca6de9584fd` |
| Base tree | `055583cef8181be59405443c2bb0ee14fc5e7690` |
| Candidate commit | `248bf52b15d381dea06c8d67d7f7c8505c53f504` |
| Candidate tree | `2f80134747575a92d1f5892cdcd61a5926142b80` |
| Result | `PASS` |
| Verification scope | Local technical verification only |
| Reviewer | `codex-root` (implementer-run verification) |
| Verified at | `2026-09-01T17:45:24Z` |
| Local host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Authorization, lineage, and scope | PASS — the active-user envelope binds Product Owner Anup Pandey, implementer `codex-root`, this change ID, and the exact approved brief. The candidate is the direct child of that brief and changes exactly the four authorized non-evidence paths; base-to-candidate changes are those four paths plus the brief, with no deletion or historical-record rewrite |
| Bootstrap transport boundary | PASS — the focused check exercised 16 offline fixtures: receive failure then success, persistent receive failure, timeout, every allowed retryable HTTP status, non-retryable HTTP/TLS/local-write failures, the 600-second aggregate deadline, exact attempts and delays, unique same-directory temporary files, atomic install, cleanup, pre-existing output preservation, symlink refusal, signal status, and ambient `.curlrc` isolation. No external request, nested curl retry, blanket retry, alternate mirror, swallowed error, or candidate-controlled policy was accepted |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | PASS from the clean exact candidate in one final authorized attempt — the controller reported Tier 3 team state `building`, base `6677735…`, head `248bf52…`, tree `2f8013…`, and five changed paths. Offline module integrity, import closure and boundaries, formatting, shell parsing, vet, typecheck, actual-host compilation, full tests, race checks, all eight fixed fuzz targets, harness/CLI reproducibility, and distribution compatibility passed without target reduction or retry-until-green |
| Fuzz and deterministic boundaries | PASS — strict configuration and provider terminal each completed literal `10000x`; the six remaining fixed targets completed their five-second budgets. The exact eight-target inventory, offline toolchain, disposable corpus roots, and direct failure propagation remained intact |
| `make v1-candidate-check GO_VERSION=1.26.7` | PASS in exactly one invocation after full verification passed — Darwin arm64/amd64 binaries and both host archives reproduced byte-for-byte; native arm64 Codex and Claude lifecycle, CLI, MCP, upgrade, rollback, removal, and disposable-root conformance passed; the unsigned release remained blocked |
| Workflow validation | PASS — `actionlint .github/workflows/release.yml` passed on the unchanged candidate bytes. Inspection confirmed the only workflow diff from the base is the exact release-base rebinding to `66777352538a514b990ffca8fa290ca6de9584fd`; manual dispatch, two-parent topology, checks, reviews, environments, signing, provider evidence, owner production approval, and publish-once gates remain fail closed |
| Environment-failure lineage | PRESERVED — an initial full attempt stopped on a missing isolated module cache under `GOPROXY=off`; separately authorized pinned module hydration then passed. The next authorized full attempt stopped during typecheck with `no space left on device`. Cleanup of only authorized ignored build/repro caches left 4.2 GiB, below the required threshold, so no test ran. After separate authority removed one exact 2.2 GiB ignored repro cache while preserving its registered worktree, 6.1 GiB was available and the single final full verification and conditional single v1 candidate check passed. No failed attempt was swallowed or repeated without fresh authority |
| Workspace hygiene | PASS — the candidate worktree and index remained clean. All registered worktrees, branches, tracked files, hydrated module/toolchain caches, and implementation bytes were preserved. The original checkout's unrelated foundation audit remained untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884` |

## Reproducible identities

| Output | SHA-256 |
|---|---|
| `.github/workflows/release.yml` | `2b5b0bb43a71065ad15675704fb1cdd25ffa5af91ba0b97e9d117e31c3aa83ef` |
| `Makefile` | `4030f9ce0b389ab7ebacf344b2b5a7c482719a9622409f766735ed21f1e68e01` |
| `scripts/harness/bootstrap-go.sh` | `787fc131f89be430001855033d56f24ed963ef81911624396686190f2314fd20` |
| `scripts/harness/check-bootstrap-go.sh` | `6e712f7e472f07e8e5b0db6f0e771c032b62b6a8d66a5483abd6fb3ce3915e6c` |
| Reproducible harness test binary | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Reproducible native CLI | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| Codex `1.0.0-dev` package | `2553f635039021c5709bf69f6b4ea16a60e20f7320b41184a6b365fa10a871b2` |
| Claude `1.0.0-dev` package | `35eea6c532027819ecbde77c4d54d8ccd292af4e8d2f2fb4546e19d18a1491e8` |
| Codex v0.1.1 distribution package | `58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba` |
| Claude v0.1.1 distribution package | `0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49` |

## Verification boundary

| Boundary | State |
|---|---|
| Actual Codex and Claude host/provider trials, model use, and provider network | `NOT_RUN` |
| macOS amd64 native execution | `NOT_RUN`; binaries were cross-built, structurally validated, and reproduced only |
| Hosted Harness, paired benchmark, trusted policy, PR, review, and accountable-owner GitHub envelopes | `NOT_RUN`; predecessor failures remain historical and local evidence does not replace exact-head hosted evidence |
| Benchmark-regression acceptance | `NOT_RUN`; no exception was requested or accepted |
| Apple credentials, signing, hardened-runtime verification, and notarization | `NOT_RUN` |
| Push, merge, release workflow dispatch, tag, GitHub release, publication, real-host installation, or deployment | `NOT_RUN` |

This implementer-run local `PASS` establishes only that the exact implementation
candidate meets the authorized technical boundary. It is not an independent
audit, hosted readiness decision, owner GO, merge authority, release authority,
or publication authority. Any implementation-byte change invalidates this
record.

## Rollback and next transition

No external release state exists from this verification. Before audit, an
authorized rollback reverts this record, then implementation commit
`248bf52b15d381dea06c8d67d7f7c8505c53f504`, then brief commit
`e27676ddb4f8875cf9a88ff3c2ef2a26a85fdfa1`, and confirms exact base tree
`055583cef8181be59405443c2bb0ee14fc5e7690`; history is preserved. After audit,
its record is reverted first.

The only next Level 7 transition is separately authorized commissioning of
`apbusinessidentity-tech` for one independent read-only Tier 3 audit at
`docs/artifacts/changes/l7-v1-hosted-bootstrap-transport-remediation-audit.md`,
bound to this verification-record successor commit and tree. This record grants
no such authority.
