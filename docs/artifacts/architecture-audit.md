# Level 7 Dev Loop — Architecture Audit Record

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-ARC-001` |
| Artifact type | Separate-context architecture and threat review record |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Status | Complete — separate-context model audit PASS; owner architecture decision pending |
| Version | 0.2.0 |
| Date | 2026-08-24 |
| Candidate | [`L7-ARC-001`](architecture.md) 0.2.0 |
| Candidate SHA-256 | `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` |
| Review mode | Three read-only, separate-context model reviewers; no reviewer edited the candidate |
| Assurance label | Separate-context model review; not qualified independent human/domain/release audit |
| Effect and risk | A1 audit-record write; no product/runtime mutation |
| Owner decision | Architecture approval remains pending |

## 1. Scope and method

The review tested the Step 3 architecture against approved [`L7-REQ-001`](requirements.md), approved [`L7-BKL-001`](feature-backlog.md), current official host/package constraints, and the architecture's own invariants and fitness functions.

Three separate read-only reviewers were used:

| Reviewer task | Review lens | Mutation authority |
|---|---|---|
| `/root/requirements_red_team` | Adversarial safety, trust, authority, evidence authenticity, recovery, evaluator/release isolation, future autonomy | None; read-only review |
| `/root/requirements_inventory` | Codex/Claude package realism, host file/data/lifecycle boundaries, current official documentation, prototype dispositions | None; read-only review |
| `/root/requirements_product` | Option quality/scoring, product journey, P0 traceability, accessibility/comprehension, technology neutrality | None; read-only review |

The reviewers did not author or patch the candidate and did not close their own findings. The primary agent performed remediation, after which the same reviewers re-examined the changed candidate. This supplies a structurally separate model review only. It is not AP2/AP3, a qualified human review, a security certification, or the release independence required by `L7-BL-042`.

## 2. Review history

| Round | Candidate identity | Reviewer result | Evidence disposition |
|---|---|---|---|
| 1 — initial candidate | Pre-remediation draft; exact digest was not retained and is not used for final binding | Safety `BLOCKED`; host `REVISE`; product conditional | Diagnostic only. One blocker, multiple high, and medium findings required changes. |
| 2 — full revised candidate | SHA-256 `59d9c7299eb860067a0886c5978ae8d6ee8e760436f09bcd480bd7cc59c03de2` | Host `PASS`; safety `PASS_WITH_CONDITIONS`; product found one approval-scope inconsistency | Prior blocker/highs closed. Remaining medium consistency findings were remediated. |
| 3 — targeted corrected candidate | SHA-256 `734584bef4554ba46f5987e93db54a05628f76564fa6ef7bc9e4acf3b5fd155b` | Safety `PASS`; product `PASS` | Critical-risk references, A1 R/W/V permissions, and evaluator→verdict→authorization→promotion separation verified. |
| 4 — final candidate record | SHA-256 `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` | Safety reviewer `PASS` after evidence-integrity check | Candidate change from round 3 was limited to version/status and architecture §21. Reviewer verified this record's existence, exact binding, chronology, findings, and assurance limitations. |

The absence of an exact Round 1 digest is disclosed rather than reconstructed. Only the exact final candidate identity above is eligible for the final audit conclusion.

## 3. Material findings and corrections

| Finding | Initial severity | Architecture correction | Final state |
|---|---:|---|---|
| Schema-valid/self-consistent forged execution evidence could advance state. | Blocker | Every gate-bearing execution record now requires a kernel-verifiable producer receipt rooted outside editable fields or fresh admitted reproduction; manual claims remain `USER_ASSERTED`/`UNVERIFIED`; `AR-011` is critical. | Corrected; technology feasibility unproved. |
| Capability closure covered A2 more clearly than A1. | High | `AR-001` now covers runtime/data discovery and every A1/A2 path with an A0→A1→A2 fail-closed capability ladder; failure blocks the applicable effect and stable C7. | Corrected; actual-host proof unproved. |
| User/host/model-bound inputs could bypass context minimization. | High | All Level-7-controlled user, attachment, host, tool, retrieval, summary, memory, log, artifact, and subagent payloads pass the context gateway; pre-plugin provider ingress is disclosed under `AR-002`. | Corrected; actual-host mediation unproved. |
| Adapter and install/update/removal authorities were missing from the permission model, and uninstall callbacks were assumed. | High | Confirmation bridge and lifecycle controller are explicit principals; lifecycle has target/approval/CAS/recovery rules; `prepare-removal` precedes official host-manager removal; skipped preparation is safe; no callback is assumed. | Corrected; host lifecycle evidence remains a Step 4/C−1 gate. |
| Public evaluator controls were candidate-writable and protected-evaluator human boundaries were incomplete. | High | Frozen public controls require separate governance; candidate/remediator paths categorically deny writes; hidden cases/labels/thresholds are non-readable/listable/writable; exposure invalidates and rotates evidence. | Corrected. |
| Builder, signer, evaluator, release verdict, accountable authorization, and promotion were collapsed or ambiguous. | High | Logical principals and permissions are separated; exact-byte flow is candidate→builder→signer→evaluator→independent verdict→accountable authorization→promoter. | Corrected. |
| Recovery preimages could duplicate unknown secrets. | High | Journals default to digests/metadata only; secret-bearing/unclassifiable targets need an approved user-owned/non-persisting or protected recovery method or block. | Corrected. |
| Codex runtime/kernel/plugin-data invocation was undocumented. | High | Package inventory and `AR-001` require actual proof of runtime location/invocation/writable data; undocumented plugin-data is optional and non-canonical. | Corrected as explicit feasibility risk. |
| OpenAI publication might expose untested non-CLI surfaces. | High | Channel and every exposed surface are part of support; universal publication blocks unless untested surfaces are excluded or added to requirements/conformance. | Corrected as explicit feasibility risk `AR-013`. |
| Package inventory, reproducibility, compaction handling, all-skill dispositions, and legal-source disposition were incomplete. | Medium | Both self-contained packages have required logical inventory; normalized payload is byte-reproducible; build/host compaction tests do not assume callbacks; all 12 skills and missing legal source have explicit dispositions. | Corrected; host re-audit PASS. |
| P0 traceability, Observe/Learn/Close, generic profile, and accessible decision-first status were incomplete. | Medium | Every P0 maps to a component and seam; runtime closes the outcome loop; E2 includes generic/feature/refactor; QA/fitness/experience criteria cover accessible status and comprehension. | Corrected; product re-audit PASS. |
| Critical risks, A1 permission, and evaluator/release flow were inconsistent after remediation. | Medium | `AR-001`/`AR-002`/`AR-011` now appear consistently in selection and approval; A1 writer is capability-scoped R/W/V; diagram separates evaluator, verdict, authorization, and promoter. | Corrected; safety and product targeted re-audits PASS. |

## 4. Residual conditions

No blocker, high, or medium architecture-document defect remained after Round 3. The following are deliberately unresolved **feasibility conditions**, not audit waivers or evidence of support:

- `AR-001`: prove kernel/runtime/writable-data discovery and non-bypassable A1/A2 admission on each actual host;
- `AR-002`: prove supported approval provenance and model-bound context mediation while disclosing pre-plugin host/provider ingress;
- `AR-011`: select and prove a producer-attestation or fresh-reproduction boundary inaccessible to editable artifact/proposal forgery;
- `AR-003`–`AR-010`, `AR-012`, and `AR-013`: retain their owners and later gates exactly as recorded in architecture §20.

Failure of `AR-001`, `AR-002`, or `AR-011` blocks the affected path and the approved stable dual-host v1.0 journey. Owner approval of architecture authorizes technology comparison only; it does not mark any feasibility risk PASS.

## 5. Final audit gate

| Check | State |
|---|---|
| Exactly three options, arithmetic, hard gates, and technology neutrality | `PASS` |
| Product journey, experience, and all 18 P0 ownership/seam mappings | `PASS` |
| Current official host/package assumptions and all 12 prototype dispositions | `PASS` |
| A1/A2 authority, evidence authenticity, recovery, context, evaluator, and release separation | `PASS` on architecture design; implementation feasibility remains explicitly unproved |
| Exact final candidate evidence-integrity confirmation | `PASS` |

**Final verdict:** `PASS` for the Step 3 architecture document. No architecture blocker, high, or medium finding is waived. The verdict is bound to the exact candidate digest above and remains a separate-context model review—not proof that the critical technology/actual-host feasibility conditions pass.

## 6. Owner gate

After the final evidence-integrity confirmation is recorded, the product owner may approve or request revision of [`L7-ARC-001`](architecture.md). Approval authorizes foundation Step 4—technology selection—only.
