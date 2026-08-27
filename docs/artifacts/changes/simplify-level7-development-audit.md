# Simplify Level 7 Development — Independent Audit

| Field | Value |
|---|---|
| Change ID | `simplify-level7-development` |
| Candidate commit | `6c93ee9399c91825b8bd4f37e6b2cd2651d64e77` |
| Candidate tree | `f1e5dceb26812ad88148edcde01da738317f2985` |
| Result | `NO_GO` |
| Reviewer | `independent-auditor` |

## Findings

| ID | Severity | Finding | Evidence | Required remediation |
|---|---|---|---|---|
| `AUD-001` | BLOCKER | Trusted CI does not read a trusted declared risk tier. Absence of a candidate-authored brief is treated as Tier 1, and CI constructs the declared scope from the complete candidate diff. A security, authorization, destructive, migration, or other non-protected high-risk change can therefore omit the brief and receive the Tier 1 fast path as long as it does not touch the hard-coded governance paths. This defeats the central tier-enforcement control and does not fail closed on under-classification. | `.github/workflows/policy.yml:83-95` sets `risk=1`, treats `brief == ""` as Tier 1, and passes `git diff --name-only` back as `--scope`; it does not consume a trusted label or PR field despite subscribing to label events. `internal/harness/buildcontrol/policy.go:69-70` intentionally skips brief discovery for explicit Tier 1, while `policy.go:180-190` only elevates the hard-coded protected paths. The brief requires tier declaration rather than downward inference and says uncertain security, destructive, data, and release risk fails closed. | Require an explicit risk-tier declaration from trusted PR/event metadata for every CI evaluation. Bind Tier 1 scope to trusted task/PR metadata rather than deriving authorization from the candidate diff. Reject absent, conflicting, or insufficient trusted classification, while retaining automatic Tier 3 elevation for protected paths. Add regression tests showing an unprotected authorization/migration change cannot become Tier 1 by omitting a brief. |
| `AUD-002` | BLOCKER | The implemented evaluator does not implement the documented state machine or failed-audit remediation path. A legitimate `NO_GO` audit is treated as an invalid audit while the evaluator remains at `awaiting-independent-audit`; it does not return to `building`. Several declared valid states are also unreachable from evaluation, so the transition-table test proves only a standalone lookup table, not executable controller transitions. | `internal/harness/buildcontrol/policy.go:281-288` returns `awaiting-independent-audit` for every audit validation failure and jumps directly to `ready` on GO. `policy.go:355-363` rejects every result except GO. `policy.go:257-264` jumps from `verified` directly to `ready` when both environment references match, so `reviewed` is unreachable for Tier 1/2. `internal/harness/buildcontrol/change_test.go:5-32` tests `nextState`/`validateTransition` in isolation; the lifecycle test in `policy_test.go` covers only the GO path. This conflicts with the brief and `references/WORKFLOW.md`, which state that failed audit returns to `building`, and with the success criterion that no valid state is unreachable. | Make evaluated states and declared transitions agree. Represent a bound NO_GO as `building` with a remediation next action; ensure `verified` and `reviewed` are either reachable controller states or remove them from the declared model. Add end-to-end controller tests for NO_GO, remediation, re-verification, re-audit, and each retained state/transition. |
| `AUD-003` | HIGH | Tier 1/2 review readiness is not enforced by the trusted policy workflow. The controller can model `ready` only from process environment references, but trusted CI never derives or passes those references and does not require readiness for Tier 1/2. The documented normal-review gate is therefore external and unstated in the executable workflow. | `internal/harness/buildcontrol/policy.go:257-264` relies on `L7_VERIFIED_REF` and `L7_REVIEW_REF`. `.github/workflows/policy.yml:89-95` invokes `--require-ready` only for Tier 3 and never sets those variables. `README.md:48-51` says repository rules must require trusted-policy and verification checks, but does not identify a required normal-review rule. | Derive verification/review identity from trusted CI/forge inputs and enforce `ready`, or explicitly make required branch-review protection part of the installation contract and test/document that dependency. |

## Evidence

| Check | Result |
|---|---|
| Candidate identity | PASS — `git rev-parse 6c93ee9399c91825b8bd4f37e6b2cd2651d64e77^{tree}` returned `f1e5dceb26812ad88148edcde01da738317f2985`. |
| Scope and history | PASS — the base-to-candidate diff contains 47 declared paths; the only changed historical-artifact paths are the permitted brief and verification record. Base has 64 tracked artifact files and candidate has 66. The unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` was preserved. |
| Owner approval binding | PASS — the external `.git/l7/approvals/simplify-level7-development.json` binds distinct actors `accountable-user` and `codex-root` to brief commit `b914c1c4b007bd1dc81aee5d18b9018720272616`; the candidate controller verifies byte-identical brief content and a brief-only planning boundary. |
| Verification binding | PASS — the verification record binds implementation commit `61f24dfde3f5b36b1c82dff84f45c7f79e9f5a51` and tree `91d16bc34d2d00a39e29fe5bee82d20ff6e00f50`; the audit candidate is its audit-only successor `6c93ee9399c91825b8bd4f37e6b2cd2651d64e77`. |
| `make policy-check` | PASS — Tier 3, state `awaiting-independent-audit`, 47 changed paths. |
| `make verify` | PASS — imports, formatting, vet/typecheck, all Go tests, and repeat-build comparison; binary digest matched the verification record. |
| `make ready-check` before audit | Expected BLOCKED — `STATE-002`, next action `record the bound independent decision`. |
| Workflow and manifest syntax | PASS — `actionlint` passed both workflows; `jq empty` passed all four plugin/marketplace manifests; `git diff --check` passed. |

## Decision

`NO_GO`. Automated verification is green and the candidate materially reduces
ceremony, but `AUD-001` allows risk-tier bypass and `AUD-002` leaves the core
state/remediation model inconsistent with the approved brief. Do not merge or
deploy this candidate. Remediation must produce a new implementation candidate,
fresh verification, and a fresh independent audit.
