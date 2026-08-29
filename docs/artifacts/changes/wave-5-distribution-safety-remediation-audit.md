# Wave 5 Distribution Safety Remediation — Independent Successor Audit

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Candidate commit | `26e7222cf435f09fecb94efcffa083ce4072a5e0` |
| Candidate tree | `7a768b46dfa8a60ede8cec49c9a5a79485ae1c9a` |
| Implementation commit | `6180a095d25eb2ecf55051b6d695fb2ac7b6f61e` |
| Implementation tree | `8a8ca8fae9d7fb6c1bf702916734c297f0b3f279` |
| Result | `GO` |
| Reviewer | `codex-release-auditor-2` |

## Decision and separation

`GO`, limited to the approved repository-local Wave 5 remediation.

The read-only auditor `codex-release-auditor-2` is distinct from accountable
owner `active-user` and implementer/verifier `codex-root`. The audit made no
file, Git, envelope, remote, host-installation, publication, deployment, or
other external mutation. Implementer-run verification was validated as
evidence but was not treated as independent review.

## Bound evidence

- Exact base `79a0739202942970536cc29782ad1b3952e7d15e` is an ancestor of
  the verified target.
- Brief commit `555c00ac489cc009671219898e4c8964c76f4f75` remains
  byte-identical to its approved version.
- Target `26e7222…` is the direct verification-only successor of implementation
  `6180a09…`; its only delta is the authorized successor verification record.
- The base-to-target diff contains only the brief, verification record, and
  four declared distribution Go paths.
- Historical verification `6e18768…` and NO_GO audit `c3b2ab5…` remain
  preserved in history. Reopening commit `ea55f95…` removes their stale
  current-tree evidence before remediation.
- The approval envelope binds owner `active-user`, implementer `codex-root`,
  and brief `555c00a…`. The active user separately authorized this audit.
- With the audit request bound to the exact target, the controller returned
  Tier 3 state `awaiting-independent-audit`, six changed paths, and next
  transition `record the bound independent decision`.

## W5R-AUD-001 disposition

`REMEDIATED`.

The implementation now:

- publishes lifecycle JSON by syncing file content and metadata, renaming, then
  syncing the containing directory;
- durably publishes newly created directory components leaf-to-root;
- syncs package publication before receipt replacement;
- syncs the destination directory before the source directory for
  cross-directory quarantine renames;
- syncs containing directories after owned-file, package-tree, receipt, and
  journal deletion;
- replays barriers for already-visible renames and already-missing paths during
  recovery; and
- stops dependent mutation when any required synchronization reports failure.

The auditor confirmed that pinned Darwin Go `1.26.7` implements
`os.File.Sync` with `F_FULLFSYNC` and an `ENOTSUP` fallback to `fsync`. The
changed package avoids prohibited raw syscall imports.

The tests truthfully establish requested barrier ordering, fail-closed error
propagation, and in-process recovery. They do not claim simulated power-loss
persistence.

## Independent checks

- Focused distribution suite: PASS, `14.517s`.
- Focused namespace-durability and recovery cases repeated five times: PASS,
  `18.678s`.
- Race-enabled distribution suite: PASS, `17.913s`.
- Deterministic distribution check: PASS:
  - Codex:
    `9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0`.
  - Claude:
    `718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb`.
  - Actual host: `NOT_RUN`.
- Exact-target policy check: PASS.
- Diff hygiene, commit/tree binding, scope, and clean tracked/index state: PASS.
- The unrelated user-owned untracked audit file remained untouched.

## Findings

| Severity | Finding | Required remediation |
|---|---|---|
| Critical | None | None |
| High | None | None |
| Medium | None | None |
| Low | None | None |

## Residual limitations

- Barrier-order tests do not simulate a process crash, reboot, controller
  failure, or power loss and cannot prove physical persistence on every
  filesystem or device.
- A crash before temporary-file rename can leave an inert `.write-*` file.
- Rollback receipt replacement has no destructive package mutation, but a
  reported directory-sync failure after the replacement becomes visible
  requires state inspection before retrying the toggling rollback.
- Concurrent hostile namespace mutation outside checked identity windows is
  not qualified.
- Linux evidence is compile-only. Actual Codex/Claude execution, real host
  lifecycle, signing, notarization, publication, deployment, hosted
  exact-successor CI, release, and support remain `NOT_RUN`, unverified,
  unauthorized, or `WITHHELD`.

## Rollback

No external state or migration requires reversal. Revert the audit-only
successor first, then verification `26e7222…`, remediation `6180a09…`,
lifecycle implementation `e1acb5c…`, compatibility implementation `db4a20b…`,
and brief `555c00a…`, using reviewed ordinary reverts and failing closed on
conflicts. Historical verification, NO_GO, and reopening commits remain
preserved and net to no current artifact. The restored tree must equal base
tree `49400a5f97355b2ccbab27d165074d6be0a24757`.

## Next executable transition

Record this bound independent `GO`, then evaluate exact-head readiness. This
decision does not itself authorize merge, installation, publication,
deployment, release, or support claims.
