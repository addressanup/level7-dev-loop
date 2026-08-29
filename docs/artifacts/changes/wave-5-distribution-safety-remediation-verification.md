# Wave 5 Distribution Safety Remediation — Verification

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Candidate commit | `e1acb5c9e3b69c11026c7935dc9174f72828ea76` |
| Candidate tree | `58f6920b9dfbeb2c7491e6f2fcb57ab7a3ef324b` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T19:03:38Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — external approval binds owner `active-user`, implementer `codex-root`, and exact brief-only commit `555c00ac489cc009671219898e4c8964c76f4f75`. `make policy-check GO_VERSION=1.26.7` selected Tier 3, exact base `79a0739202942970536cc29782ad1b3952e7d15e`, candidate and tree above, five changed paths, and truthful `building` state. |
| Lineage and scope | PASS — the exact base is the direct ancestor of brief `555c00a…`, compatibility implementation `db4a20b…`, and lifecycle implementation candidate `e1acb5c…`. The base-to-candidate diff contains only the approved brief and four declared distribution Go files. |
| Focused distribution suite | PASS — repository-pinned `go test -mod=readonly -count=1 ./internal/harness/distribution` covered exact compatibility mutations, package identity substitution, install and removal transaction faults, receipt binding, abandoned stages, unowned conflicts, containment, and rollback. |
| Focused race suite | PASS — CGO race instrumentation passed `./internal/harness/distribution` after the final candidate was committed. |
| Full repository verification | PASS — `make verify GO_VERSION=1.26.7` passed policy, offline module integrity and tidy diff, import/effect boundaries, formatting, shell syntax, vet, type compilation, compile-only tagged actual-host coverage, the complete test suite, harness and CLI reproducibility, and deterministic distribution checks. |
| Full race suite | PASS — repository-pinned Go `1.26.7` with CGO race instrumentation, exact harness identity link values, and a three-minute bound passed `./...` on Darwin arm64. |
| Cross-builds | PASS — `make cli-cross-build GO_VERSION=1.26.7` produced the declared Darwin arm64 and amd64 CLI binaries with the pinned offline toolchain. |
| Compatibility and package identity | PASS — Codex and Claude compatibility projections require their exact ordered development tuples, distinguish `null` from an empty capability list, and reject every tested field mutation. Package validation binds caller host/version, strict `DISTRIBUTION.json`, the matching sole host manifest, permissions, provenance, SBOM, and the exact ordered inventory. |
| Install recovery | PASS — deterministic journal-owned stages tolerate missing and partial transaction-owned files, reject unexpected or symlinked state, and bind pending installs to canonical hashes of both base and committed receipts. Recovery fails closed for changed bindings, missing committed packages, and modified or missing base packages. |
| Removal recovery | PASS — removal journals precede mutation; all five durable fault points resume idempotently. Unowned files/directories, conflicting source and quarantine trees, symlinked paths, changed bytes, malformed/cross-host/escaping journals, and receipt mismatch fail closed without deleting conflicting bytes. |
| Distribution reproducibility | PASS — two development archives retained their exact historical SHA-256 identities, and every generated identity projection remains inventory-covered. |
| Diff and workspace hygiene | PASS — `git diff --check` passed; the implementation candidate and index were clean after verification. The only workspace dirt was the unrelated user-owned untracked `docs/artifacts/foundation-rebaseline-admission-audit.md`, which remained untouched and unstaged. |
| Artifact budget | PASS — the implementation candidate contains one brief and no evidence record. This sole verification record and one future audit remain within the Tier 3 maximum of three artifact paths. |

## Deterministic identities

| Output | SHA-256 |
|---|---|
| Codex development package | `9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0` |
| Claude development package | `718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb` |
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Applicability and residual boundaries

- No production data, schema, or migration exists in this change. Lifecycle
  state is confined to disposable fixture roots, so production migration and
  operational rollback checks are not applicable.
- Performance benchmarking is not applicable: product CLI/runtime paths are
  unchanged and archive, file, receipt, and JSON inputs remain explicitly
  bounded. No benchmark claim is made.
- Security evidence is limited to the exercised ownership, containment,
  symlink, malformed-state, import/effect-boundary, race, and vet checks. This
  is not a broader security certification, and concurrent external filesystem
  mutation is not qualified.
- Actual Codex or Claude execution, real host installation/removal, signing,
  notarization, publication, deployment, and hosted exact-candidate CI remain
  `NOT_RUN` or unverified. Support remains `WITHHELD`.

## Rollback proof

The verified local chain is exact base `79a0739202942970536cc29782ad1b3952e7d15e`
(tree `49400a5f97355b2ccbab27d165074d6be0a24757`), brief `555c00a…`,
compatibility implementation `db4a20b…`, and lifecycle implementation
`e1acb5c…` (tree `58f6920b9dfbeb2c7491e6f2fcb57ab7a3ef324b`).
Before audit, reverse the verification record, lifecycle implementation,
compatibility implementation, and brief in that order. After audit, reverse
the audit record first. Use ordinary reviewed reverts, fail closed on conflicts
or unexpected paths, and confirm the restored tree equals the exact base tree.
No external state or migration must be unwound.

## Verification boundary

This is implementer-run technical verification, not the independent audit,
owner review, merge readiness, release, installation, publication, deployment,
or support authority. Any implementation change invalidates this record. The
authorized next transition is the distinct read-only `l7-release` audit of the
verification-only successor.
