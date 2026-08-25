# Level 7 Dev Loop — Wave 1 Audit Remediation

| Field | Value |
|---|---|
| Artifact ID | `L7-REM-W01-003` |
| Artifact type | Finding-specific Level 7 Release Validation Mode C remediation record |
| Version | 0.3.0 |
| Date | 2026-08-26 |
| Status | **REMEDIATED IN THIRD SUCCESSOR CANDIDATE — EXACT VERIFICATION AND FRESH INDEPENDENT MODE B AUDIT REQUIRED** |
| Supersedes | `L7-REM-W01-002` only for the third successor; prior bytes remain at candidate commit `64eee79` |
| Source audit | `L7-AUD-W01-003` at `docs/artifacts/wave-01-audit.md`, SHA-256 `27b255f6cbdaaf050c7268828f26cd24be9b5120aa38db85ea2fb63fc9288039` |
| Source audit commit | `910c21406b717da13a7bdfe8ac73357b5078b251` — reviewer-authored second-successor NO-GO preserved before candidate mutation |
| Audited candidate | Commit `64eee794519e381a69d284c32cc35ac58897aa2f`; tree `5829f97b1ef9e29a9c61d5872ed725920be65f84` |
| Audited evidence child | Commit `8f82512fd4eb3e24ba6033427badaa3439e06450`; tree `a005a29728114fc7cf7f7d3243b4c988619d5836` |
| Authority | Accountable-owner authorization in the current conversation for the recommended Mode C remediation of `AUD-W01-012` and `AUD-W01-013` |
| Disposition | Findings corrected locally with separate commits and regression proof; not self-cleared and no `GO` issued |

## 1. Scope and preservation

This Mode C pass confirms and remediates only `AUD-W01-012` and `AUD-W01-013`. It preserves the audited candidate, evidence-only child, and third NO-GO audit in Git history. All earlier Wave 1 audit and remediation records remain available at their bound historical commits.

The generic skill output paths remain protected historical inputs. `docs/artifacts/release-audit-remediation.md` remains SHA-256 `d124659ce655415252255e5e50cd117ba964d634720fd4d9ccd98a0dd9872ca2`, and `docs/artifacts/principal-engineer-release-audit.md` remains SHA-256 `e606e7ad8e756667c0bf560463f296232cbf8f74e7108c4bd31afd1c647ad24c`. Repository design reserves this registered Wave 1 path for the current remediation record, so neither generic record was edited.

This pass does not authorize or perform merge, release, deployment, publication, remote creation, hosted CI, provider/host execution, grant activation, product work, or Wave 2 work.

## 2. Finding-specific remediation

| Finding | Confirmation | Remediation | Regression proof | Commit |
|---|---|---|---|---|
| `AUD-W01-012` — MEDIUM | Confirmed: `loadSkillInventory` opened the fixed `skills/` root with plain `os.Open` before policy scanning, so a FIFO could block and a symlink could expose external directory entries | Added canonical rooted directory acquisition, component/type checks, no-follow/nonblocking directory open, bounded enumeration, and pre/open/post identity, mode, time, and size validation; factored the controller path for deterministic end-to-end testing | `TestSkillInventoryRejectsSymlinkFIFOAndReplacementWithoutBlocking` and `TestControllerRejectsSkillInventorySymlinkFIFOAndReplacementWithoutBlocking` exercise real symlink, FIFO, and in-window replacement fixtures and require `BCTL-022`/`BCTL-023` within one second | `b9a48d5f55abbab9eeab1f0a4f1a536351a13a6e` |
| `AUD-W01-013` — MEDIUM | Confirmed: `filepath.WalkDir` materialized each directory before the callback could enforce the `512` directory/file counts | Replaced whole-directory walking with deterministic rooted `ReadDir(n)` batches capped at `1,027` entries, retained the separate global `512` directory, `512` file, and 8 MiB byte ceilings, hardened repository-file reads with rooted no-follow/nonblocking identity checks, and added a fail-closed five-second scan deadline (`SCOPE-339`) | `TestRepositoryScanCapsSingleDirectoryReadBeforeEnumeration` proves the pre-enumeration request ceiling and stable `SCOPE-346`; `TestRepositoryScanDeadlineFailsClosedBeforeFurtherIO` proves timeout cannot pass or continue I/O; existing exact count/byte tests remain active | `47f8e7af4e942964f8a8046864fdcb8dc267ffa1` |

Each actionable finding has its own conventional fix commit. The later candidate-binding commit contains only this remediation record and the regenerated self-excluded candidate manifest.

## 3. Verification evidence

The finding-specific regression selections passed under both pinned local toolchains before candidate binding:

| Toolchain | Targeted result |
|---|---|
| Go 1.26.7 | Inventory-root direct/controller adversaries plus repository enumeration/count/deadline/byte tests: `PASS` |
| Go 1.27.0 | Inventory-root direct/controller adversaries plus repository enumeration/count/deadline/byte tests: `PASS` shadow development evidence |

The complete `make verify GO_VERSION=1.26.7` and `make verify GO_VERSION=1.27.0` matrices are intentionally executed only after the containing candidate commit and tree are frozen. Their exact results and reproducible test-binary SHA-256 values belong in the later evidence-only child, not circularly in this candidate record. Until both matrices pass and that child is frozen, the third successor is not ready for independent review.

All local Go commands remain offline and use only the approved ignored repository-scoped cache paths. Hosted Ubuntu CI, actual Codex/Claude conformance, Controlled Client qualification, protected evaluation, grant security review, AP2/AP3, signing, publication, release, and deployment remain `NOT_RUN` or absent.

## 4. Successor binding and decision boundary

The containing third-successor candidate commit/tree and regenerated manifest digest are recorded non-circularly in the evidence-only child created after candidate freeze. This record does not embed its own digest or containing commit.

All prior NO-GO audits remain correct for their exact audited candidates. This remediator does not issue `GO`, `CONDITIONAL GO`, or acceptance. After exact verification and evidence freeze, a fresh structurally separate reviewer must reproduce the new candidate/evidence identities, inspect both finding dispositions and regression proofs, and issue a new Mode B verdict. Any candidate or evidence byte change invalidates that later review.

No merge, release, deployment, publication, external action, Wave 2 start, or autonomous continuation follows from remediation success.
