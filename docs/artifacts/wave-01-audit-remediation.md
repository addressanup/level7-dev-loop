# Level 7 Dev Loop — Wave 1 Audit Remediation

| Field | Value |
|---|---|
| Artifact ID | `L7-REM-W01-002` |
| Artifact type | Finding-specific Level 7 Release Validation Mode C remediation record |
| Version | 0.2.0 |
| Date | 2026-08-26 |
| Status | **REMEDIATED IN SECOND SUCCESSOR CANDIDATE — FRESH INDEPENDENT MODE B AUDIT REQUIRED** |
| Supersedes | `L7-REM-W01-001` only for the second successor; prior bytes remain at candidate commit `808fc16` |
| Source audit | `L7-AUD-W01-002` at `docs/artifacts/wave-01-audit.md`, SHA-256 `bc58479c7626dd88ad4937eabfa0482b8c1e11a95f2a9c95ded114b379a1ef1b` |
| Source audit commit | `4a0685a` — reviewer-authored successor NO-GO preserved before candidate mutation |
| Audited candidate | Commit `808fc16b1c13c96e30c66f08a97dd2a014b31db0`; tree `bd914fdcb88aad54be5acf19c60dbcbb35f2d709` |
| Audited evidence child | Commit `c0f834f7318081d99d0dcbbc2b82d907961be633`; tree `55536582355517969b802be11856398fafe3b3e5` |
| Authority | Accountable-owner authorization in the current conversation for the recommended Mode C remediation of `AUD-W01-008` and `AUD-W01-009` |
| Disposition | Findings corrected locally with separate commits and regression proof; not self-cleared and no `GO` issued |

## 1. Scope and preservation

This Mode C pass confirms and remediates only `AUD-W01-008` and `AUD-W01-009`. It preserves the audited candidate, evidence-only child, and successor NO-GO audit in Git history. The earlier `L7-AUD-W01-001` and `L7-REM-W01-001` records also remain available at their bound historical commits.

The generic skill output path `docs/artifacts/release-audit-remediation.md` is a protected historical base input with SHA-256 `d124659ce655415252255e5e50cd117ba964d634720fd4d9ccd98a0dd9872ca2`. Repository design reserves this distinguishable, registered Wave 1 path for the current remediation record, so the protected generic record was not edited.

This pass does not authorize or perform merge, release, deployment, publication, remote creation, hosted CI, provider/host execution, grant activation, product work, or Wave 2 work.

## 2. Finding-specific remediation

| Finding | Confirmation | Remediation | Regression proof | Commit |
|---|---|---|---|---|
| `AUD-W01-008` — MEDIUM | Confirmed: `appendFindings` retained the first 51 findings before `renderFindings` sorted them, so unordered map traversal selected an unstable failure subset | Kept the 51-finding memory ceiling while retaining the lexicographically smallest bounded set independent of arrival order; centralized the comparison used for collection and rendering | A greater-than-cap failing trace is generated from an unordered 163-entry map in two separate test processes; both exit nonzero with byte-identical output, the exact expected retained boundary, and `BCTL-099` | `86a0a47` |
| `AUD-W01-009` — MEDIUM | Confirmed: `readStrictFile` called `os.Open` before any rooted shape/identity check, so a symlink could be consumed and a FIFO could block before policy rejection | Added canonical rooted-relative validation, rooted `os.Root` traversal, pre-read component/regular/single-link checks, no-follow nonblocking opening, and opened/path pre/post identity, mode, link, time, and size validation | Real filesystem fixtures reject final/intermediate symlinks, FIFO, and hardlinks before consumption; a deterministic in-window mutation fails with `BCTL-023`; FIFO completion is time-bounded | `f5f197c` |

Each actionable finding has its own conventional fix commit. The later candidate-binding commit contains only this remediation record and the regenerated self-excluded candidate manifest.

## 3. Verification evidence

After both finding commits and the refreshed candidate-manifest bytes, the complete local matrix passed:

| Command | Result | Reproducible test-binary SHA-256 |
|---|---|---|
| `make verify GO_VERSION=1.26.7` | `PASS` | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| `make verify GO_VERSION=1.27.0` | `PASS` shadow development evidence | `da0ff13d148e68a648a4ee23fa35c4e173f8145bd97a5d1beddcc9422000f85a` |

Both runs remained offline, reported zero module dependencies, passed the phase-aware candidate and two-package import gates, and passed compile, typecheck, unit, adversarial, vet, formatting, shell, and repeat-build checks. The exact committed candidate is rerun with both matrices before the evidence-only child is finalized. The verifier writes only the ignored repository-scoped Go/reproducibility cache paths permitted by the approved design.

Hosted Ubuntu CI, actual Codex/Claude conformance, Controlled Client qualification, protected evaluation, grant security review, AP2/AP3, signing, publication, release, and deployment remain `NOT_RUN` or absent.

## 4. Successor binding and decision boundary

The containing second-successor candidate commit/tree and regenerated manifest digest are recorded non-circularly in the evidence-only child created after candidate freeze. This record does not embed its own digest or containing commit.

Both NO-GO audits remain correct for their exact audited candidates. This remediator does not issue `GO`, `CONDITIONAL GO`, or acceptance. A fresh structurally separate reviewer must reproduce the new candidate/evidence identities, inspect both finding dispositions and regression proofs, and issue a new Mode B verdict. Any candidate or evidence byte change invalidates that later review.

No merge, release, deployment, publication, external action, Wave 2 start, or autonomous continuation follows from remediation success.
