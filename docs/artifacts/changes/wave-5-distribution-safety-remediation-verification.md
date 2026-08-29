# Wave 5 Distribution Safety Remediation — Successor Verification

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Candidate commit | `6180a095d25eb2ecf55051b6d695fb2ac7b6f61e` |
| Candidate tree | `8a8ca8fae9d7fb6c1bf702916734c297f0b3f279` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-29T19:46:33Z` |
| Host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the external approval envelope binds owner `active-user`, implementer `codex-root`, and unchanged brief-only commit `555c00ac489cc009671219898e4c8964c76f4f75`. The active user separately authorized this fresh `l7-release` run. `make policy-check GO_VERSION=1.26.7` selected Tier 3, exact base `79a0739202942970536cc29782ad1b3952e7d15e`, implementation candidate and tree above, five changed paths, and truthful pre-verification `building` state. |
| Lineage and scope | PASS — the exact base is an ancestor of the implementation candidate. The base-to-candidate diff contains only the approved brief and four declared distribution Go files. Historical verification `6e18768…` and NO_GO audit `c3b2ab5…` remain in Git history; reopening commit `ea55f95…` removed those stale records from the current tree before durability remediation `6180a09…`. They are not reused as current evidence. |
| Focused distribution suite | PASS — repository-pinned `go test -mod=readonly -count=1 -timeout=2m ./internal/harness/distribution` completed in `14.352s`. It covered compatibility, package identity, receipt binding, install/removal recovery, path ownership and containment, rollback, and the new namespace-durability cases. |
| Focused race suite | PASS — the same distribution package passed with CGO race instrumentation in `18.562s`. |
| Namespace durability protocol | PASS as implementation evidence — tests observed durable-record publication before dependent mutation; staged-file and directory publication; same-directory package publication; destination-before-source cross-directory quarantine rename barriers; barriers after owned-file, tree, receipt, and journal deletion; durable root-component publication; and replay barriers for already-visible or already-missing recovery state. Injected sync errors failed closed and recovered in-process after the hook was restored. |
| Full repository verification | PASS — `make verify GO_VERSION=1.26.7` passed exact-candidate policy, offline module integrity and tidy diff, import/effect boundaries, formatting, shell syntax, vet, type compilation, compile-only tagged actual-host coverage, the complete test suite, harness and CLI reproducibility, and deterministic distribution checks. The distribution package completed within the repository's two-minute test bound in `21.083s`. |
| Full race suite | PASS — repository-pinned Go `1.26.7` with CGO race instrumentation, exact harness identity link values, and a three-minute bound passed `./...` on Darwin arm64. |
| Cross-platform compile boundary | PASS — `make cli-cross-build GO_VERSION=1.26.7` produced the declared Darwin arm64 and amd64 CLI binaries. The changed distribution package also compiled for `linux/amd64` with `CGO_ENABLED=0`; that was compile-only evidence, not execution on Linux. |
| Security and compatibility controls | PASS within scope — vet, race, import/effect boundaries, physical-directory identity checks, containment, symlink rejection, malformed-state rejection, exact ordered compatibility tuples, archive/manifest identity binding, and ownership-preserving failure cases passed. This is not a broader security certification. |
| Data, migration, and operational behavior | NOT APPLICABLE — this change modifies build-time validation and disposable fixture lifecycle state only. It adds no production data, schema, migration, runtime provider behavior, deployment, or external effect. |
| Performance | NOT APPLICABLE as a product benchmark — product CLI/runtime paths are unchanged. Durability work is confined to bounded fixture lifecycle operations, and the focused/full suites remained within their configured limits. No throughput or latency claim is made. |
| Distribution reproducibility | PASS — both development archives retained their exact historical SHA-256 identities, and `actual_host=NOT_RUN` remained explicit. |
| Diff and workspace hygiene | PASS — the implementation candidate and index were clean throughout candidate-bound verification; `git diff --check` passed. The only workspace dirt was the unrelated user-owned untracked `docs/artifacts/foundation-rebaseline-admission-audit.md`, which remained untouched and unstaged. |
| Artifact budget | PASS — the implementation candidate contains one brief and no current evidence record. This sole successor verification record and one future audit remain within the Tier 3 maximum of three current artifact paths. |

## Deterministic identities

| Output | SHA-256 |
|---|---|
| Codex development package | `9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0` |
| Claude development package | `718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb` |
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Residual and claim boundaries

- The tests prove requested file/directory synchronization calls, their order,
  fail-closed error propagation, and in-process recovery. They do not simulate
  a process crash, reboot, storage-controller failure, or power loss and do not
  prove physical persistence on every filesystem or device.
- A hard crash before a durable record's temporary file is renamed can leave an
  ignored `.write-*` orphan. Such a file is not a journal or receipt and cannot
  authorize lifecycle mutation, but repeated crashes can leave inert metadata
  files requiring fixture-root cleanup.
- Rollback atomically replaces one receipt and has no dependent destructive
  package mutation. If receipt publication becomes visible but its directory
  sync reports failure, old or new receipt state can be valid and visible;
  callers must recover or inspect before blindly retrying the toggling rollback.
- Concurrent hostile mutation outside the checked physical-directory identity
  windows is not qualified.
- Actual Codex or Claude execution, real host installation/removal, signing,
  notarization, publication, deployment, hosted exact-candidate CI, release,
  and support remain `NOT_RUN`, unverified, unauthorized, or `WITHHELD`.

## Rollback proof

The verified implementation remains repository-local. Before audit, reverse
this verification-only successor, then remediation `6180a09…`, original
lifecycle implementation `e1acb5c…`, compatibility implementation `db4a20b…`,
and brief `555c00a…` in that order. After audit, reverse the audit-only successor
first. Historical verification, NO_GO, and reopening commits remain preserved
and net to no current artifact. Use ordinary reviewed reverts, fail closed on
conflicts or unexpected paths, and confirm the restored tree equals exact base
tree `49400a5f97355b2ccbab27d165074d6be0a24757`. No external state or migration
must be unwound.

## Verification boundary

This is implementer-run technical verification, not the independent audit,
owner review, merge readiness, release, installation, publication, deployment,
or support authority. Any implementation change invalidates this record. The
authorized next transition is a distinct read-only `l7-release` audit of the
verification-only successor.
