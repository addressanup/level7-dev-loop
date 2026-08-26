# Level 7 Dev Loop — Wave 2 Design

| Field | Value |
|---|---|
| Artifact ID | `L7-W02-DES-001` |
| Artifact type | Proposed Wave 2 technical and delivery design |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 2 — Provider-neutral semantic and evaluation foundation |
| Version | 0.1.0 |
| Date | 2026-08-26 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Change contract | [`L7-W02-CC-001`](wave-02-change-contract.md) 0.1.0, SHA-256 `367dab50ee994b21eb2503ab7538c9687546d4e55a4275c563a87b80973eaaf4` |
| Specification | [`L7-W02-SPEC-001`](wave-02-specification.md) 0.1.0, SHA-256 `3cb7304e18bf1320160252ac4b74b7321e714728cb5079cb4e24d7e45bc6eb5d` |
| Proposal approval | Anup Pandey approved the exact presented contract/specification pair in the current conversation on 2026-08-26; persisted proposal text remains `AP0` and grants no implementation authority |
| Canonical root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Planning branch | `feat/wave-02-semantic-evaluation` |
| Source identity | Commit `c35bf4b6e4a38ca54899882a7e3c574d03d1df85`; tree `eb60ac4d167df96ba02822c458cb81493e05537b`; the approved contract and specification are the only pre-design worktree additions |
| Predecessor checkpoint | Wave 1 fifth-successor candidate independently audited `GO`; audit SHA-256 `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f`; audit-preservation commit is the exact source identity above |
| Backlog / requirements | `L7-BL-002`, `L7-BL-003`; exactly 29 normative IDs |
| Primary / secondary change class | Architecture/modernization and feature/behavior change / security and evaluator-governance change |
| Risk / maximum later effect | `R3 — high` / A2 local repository change only after separate exact implementation approval |
| Current effect | A1 addition of this design proposal only; no existing-file edit, staging, commit, test, cache, dependency, Git-ref, or external effect |
| Feature/exposure state | `NOT_APPLICABLE`; Wave 2 creates no user-visible behavior |
| Next gate | Accountable-owner approval or revision of this exact design; implementation remains separately unauthorized |

## 1. Decision requested

The accountable owner is asked to approve or revise this exact design. Approval would confirm the architecture, schemas, stable-rule contracts, exact 72-path envelope, control ownership, serial implementation slices, local verification effects, candidate/evidence construction, recovery model, and independent-audit seam.

Design approval would not authorize implementation, staging, a commit, a merge, a dependency, a test or cache write, a provider/model/host run, protected evaluation, hosted CI, network use, publication, release, deployment, exposure, or continuation into Wave 3. A later implementation request must bind this design's exact SHA-256, reproduce the source tuple and worktree, authorize the listed local effects, and name the single integration writer.

## 2. Design drivers and hard constraints

The design follows these priorities in order:

1. **Preserve the audited predecessor.** Wave 1 remains a distinct historical phase. Its candidate, evidence, audit, and audit-preservation commit are not rewritten or relabeled as Wave 2 proof.
2. **Admit paths before capability.** The successor phase, exact path policy, ownership partition, negative boundary fixtures, and import rules land before the first semantic or evaluator source commit.
3. **One semantic source.** The 29 approved requirements reduce into one provider-neutral registry. Templates, projections, examples, fixtures, and future host overlays cannot become parallel policy sources.
4. **Freeze evaluation before tuning.** Public truth, deterministic graders, coverage, adjudication, and thresholds freeze only after all initial cases exist and before any later candidate/remediation work.
5. **Pure zero-dependency logic.** `internal/render` and `internal/evaluator` use the pinned Go standard library, accept explicit in-memory inputs, and have no filesystem, process, network, clock, randomness, terminal, credential, or mutation interface.
6. **Fail closed and bounded.** Duplicate/unknown fields, ambiguous identities, missing obligations, invalid combinations, over-limit input, unsupported capability, grader error, and incomplete evidence are non-passing.
7. **Truthful assurance.** Local same-user role separation is detectable governance, not an OS-enforced security boundary. Protected evaluation, host/provider behavior, controlled mutation, release, and support remain later `NOT_RUN`/`NOT_EVALUATED` gates.

### 2.1 Bound design inputs

| Input | SHA-256 / identity |
|---|---|
| `AGENTS.md` | `54496725a42eb7e6cce2a749e82a408d7277743ec8ad83c41373ceefbd4d0afa` |
| `docs/artifacts/requirements.md` | `a9ff0f30c62ba74bdb9cdbc81d06663642d468f2c8795341f83b9662be59922f` |
| `docs/artifacts/feature-backlog.md` | `df5d87a224d5ec61b31bff6b0cb1b4db4f5a9a03eb476cee438387cc7a98e995` |
| `docs/artifacts/architecture.md` | `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` |
| `docs/artifacts/technology-selection.md` | `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` |
| `docs/artifacts/orchestration-plan.md` | `a45cb13b7ce68029c23736188531e0379cad0ff5d71409ddf6bfc850c1872968` |
| `docs/artifacts/wave-01-design.md` | `07953b2319635846505a018c3e4cc66705e0c263ab01b0a5c79e75cdaf1fb8e8` |
| `docs/artifacts/wave-01-evidence.md` | `1d350436398fad8f53a6221fc1c1f2e64ac9bfa0f1b8c5317f1003c1a198b98c` |
| `docs/artifacts/wave-01-audit.md` | `491c686dc57f3ca4050646826b8919d6239a5b8d971c051bb77f9ff12167034f` |
| `harness/control-ownership.tsv` | `5f043166e9d698ceba278e22ce182a396faefd5eac929ac988dc6f25660fa8d8` |
| installed `l7-build` skill | `ab4b45141f1bc20961ae6d4db5048913af6d4ca040c6e876e1a6bf7353a3a95f` |

The transport skill is instruction, not authority. Any byte/identity change to a bound input before implementation requires impact review and, when material to scope, risk, effects, ownership, or acceptance, a revised design and fresh approval.

## 3. Component and trust shape

```text
approved requirements/backlog ───────────────┐
semantic JSON + narrow template ─────────────┼─> internal/render ─> reference projections
semantic schemas/descriptors ────────────────┘        │
                                                      │ typed result + obligation accounting
public cases + frozen truth/protocol/graders ─────────┼─> internal/evaluator ─> deterministic result
                                                      │
phase/base/path/ownership/import/control manifests ───┴─> build-control admission report

candidate/remediator ──X── frozen evaluator controls
candidate/remediator ──X── protected holdout/operator/release plane (external and absent)
```

`internal/harness/buildcontrol` remains the sole repository-effect boundary. It loads fixed regular files, validates the repository snapshot, invokes pure semantic/evaluator functions with copied bytes, and reports a bounded decision. The two new packages do not discover files or environment state themselves.

The dependency direction is:

```text
internal/harness/buildcontrol -> internal/render
internal/harness/buildcontrol -> internal/evaluator -> internal/render
```

No product package may import `internal/harness`; `internal/render` may not import `internal/evaluator`; neither pure package may import future state, policy, transaction, executor, receipt, adapter, channel, distribution, generated-package, or harness packages.

## 4. Exact path and ownership envelope

The Wave 2 path policy contains exactly 72 bytewise-sorted data rows. No deletion or conditional path is allowed. Ten predecessor paths are `modify`, 61 paths are ordinary `add`, and one path is `audit-only`. The owner identifiers below are exact path-policy values.

### 4.1 Existing paths proposed for modification — 10

| # | Path | Intended delta | Owner |
|---:|---|---|---|
| 1 | `.github/workflows/harness.yml` | Replace Wave 1-only wording with truthful active Wave 2 build-control wording while retaining pinned checkout, read-only permissions, baseline blocking, and shadow separation | `harness-integrator` |
| 2 | `README.md` | Describe the Wave 2 development checkpoint, commands, non-goals, evaluator boundary, and `NOT_RUN` states | `wave-integrator` |
| 3 | `harness/control-ownership.tsv` | Add Wave 2 records and split formerly broad schema/fixture ownership into disjoint semantic, evaluator, state, and future-feature prefixes | `harness-integrator` |
| 4 | `harness/import-boundaries.tsv` | Add pure-package effect and dependency rules without weakening any existing rule | `harness-integrator` |
| 5 | `harness/phases.tsv` | Preserve Wave 1 as `historical` and add exactly one Wave 2 `active` successor row | `harness-integrator` |
| 6 | `internal/harness/buildcontrol/main.go` | Report `wave-02-v1`; derive success-source digests from the active phase and distinguish in-progress from final-candidate state | `harness-integrator` |
| 7 | `internal/harness/buildcontrol/ownership.go` | Load the active path policy, validate the expanded exact ownership design, and retain orchestration cross-checks | `harness-integrator` |
| 8 | `internal/harness/buildcontrol/ownership_test.go` | Add disjoint-prefix, active-policy, Wave 2 owner, and candidate/evaluator-role negative cases | `harness-integrator` |
| 9 | `internal/harness/buildcontrol/policy.go` | Generalize the phase engine, bind the Wave 2 base/path/candidate closure, admit only approved Wave 2 product paths, and enforce the evaluator-control manifest | `harness-integrator` |
| 10 | `internal/harness/buildcontrol/policy_test.go` | Add successor, closure, final-candidate, real-order, mixed-subset, and control-freeze regression cases while preserving all Wave 1 regressions | `harness-integrator` |

`Makefile`, `go.mod`, all lock files, `scripts/harness/*`, Wave 1 records, prototype assets, plugin manifests, and `references/WORKFLOW.md` remain byte-identical.

### 4.2 New governance and build-control paths — 12

| # | Path | Change | Owner |
|---:|---|---|---|
| 11 | `docs/artifacts/wave-02-approval.md` | `add` | `wave-integrator` |
| 12 | `docs/artifacts/wave-02-audit.md` | `audit-only` | `independent-reviewer` |
| 13 | `docs/artifacts/wave-02-candidate.sha256` | `add` | `wave-integrator` |
| 14 | `docs/artifacts/wave-02-change-contract.md` | `add` | `wave-integrator` |
| 15 | `docs/artifacts/wave-02-design.md` | `add` | `wave-integrator` |
| 16 | `docs/artifacts/wave-02-evidence.md` | `add` | `wave-integrator` |
| 17 | `docs/artifacts/wave-02-specification.md` | `add` | `wave-integrator` |
| 18 | `harness/wave-02-evaluator-controls.sha256` | `add` | `harness-integrator` |
| 19 | `harness/wave-02-base.sha256` | `add` | `harness-integrator` |
| 20 | `harness/wave-02-paths.tsv` | `add` | `harness-integrator` |
| 21 | `internal/harness/buildcontrol/wave2.go` | `add` | `harness-integrator` |
| 22 | `internal/harness/buildcontrol/wave2_test.go` | `add` | `harness-integrator` |

### 4.3 New provider-neutral semantic and schema paths — 26

| # | Path | Owner |
|---:|---|---|
| 23 | `semantic/taxonomy/registry.json` | `semantic-owner` |
| 24 | `semantic/taxonomy/obligations.json` | `semantic-owner` |
| 25 | `semantic/taxonomy/guardrails.json` | `semantic-owner` |
| 26 | `semantic/taxonomy/knowledge.json` | `semantic-owner` |
| 27 | `semantic/workflows/reference/contract.json` | `semantic-owner` |
| 28 | `semantic/workflows/reference/prompt.md.tmpl` | `semantic-owner` |
| 29 | `semantic/profiles/generic.json` | `semantic-owner` |
| 30 | `semantic/profiles/feature-change.json` | `semantic-owner` |
| 31 | `semantic/profiles/behavior-preserving-refactor.json` | `semantic-owner` |
| 32 | `schemas/semantic/budget.schema.json` | `semantic-owner` |
| 33 | `schemas/semantic/delegation.schema.json` | `semantic-owner` |
| 34 | `schemas/semantic/guardrail.schema.json` | `semantic-owner` |
| 35 | `schemas/semantic/knowledge.schema.json` | `semantic-owner` |
| 36 | `schemas/semantic/obligation.schema.json` | `semantic-owner` |
| 37 | `schemas/semantic/output.schema.json` | `semantic-owner` |
| 38 | `schemas/semantic/profile.schema.json` | `semantic-owner` |
| 39 | `schemas/semantic/prompt-ir.schema.json` | `semantic-owner` |
| 40 | `schemas/semantic/taxonomy.schema.json` | `semantic-owner` |
| 41 | `schemas/semantic/workflow.schema.json` | `semantic-owner` |
| 42 | `schemas/evaluation/adjudication.schema.json` | `evaluator-owner` |
| 43 | `schemas/evaluation/case.schema.json` | `evaluator-owner` |
| 44 | `schemas/evaluation/coverage.schema.json` | `evaluator-owner` |
| 45 | `schemas/evaluation/grader.schema.json` | `evaluator-owner` |
| 46 | `schemas/evaluation/protocol.schema.json` | `evaluator-owner` |
| 47 | `schemas/evaluation/run-manifest.schema.json` | `evaluator-owner` |
| 48 | `schemas/evaluation/truth-label.schema.json` | `evaluator-owner` |

### 4.4 New fixtures and pure Go packages — 24

| # | Path | Owner |
|---:|---|---|
| 49 | `fixtures/public/bl-002/semantic-cases.json` | `semantic-owner` |
| 50 | `fixtures/public/bl-002/broken-candidates.json` | `semantic-owner` |
| 51 | `fixtures/public/bl-003/adjudication.json` | `evaluator-owner` |
| 52 | `fixtures/public/bl-003/cases.json` | `evaluator-owner` |
| 53 | `fixtures/public/bl-003/coverage.json` | `evaluator-owner` |
| 54 | `fixtures/public/bl-003/grader-registry.json` | `evaluator-owner` |
| 55 | `fixtures/public/bl-003/protocol.json` | `evaluator-owner` |
| 56 | `fixtures/public/bl-003/truth-labels.json` | `evaluator-owner` |
| 57 | `internal/render/compile.go` | `semantic-owner` |
| 58 | `internal/render/compile_test.go` | `semantic-owner` |
| 59 | `internal/render/decode.go` | `semantic-owner` |
| 60 | `internal/render/decode_test.go` | `semantic-owner` |
| 61 | `internal/render/doc.go` | `semantic-owner` |
| 62 | `internal/render/model.go` | `semantic-owner` |
| 63 | `internal/render/validate.go` | `semantic-owner` |
| 64 | `internal/render/validate_test.go` | `semantic-owner` |
| 65 | `internal/evaluator/coverage.go` | `evaluator-owner` |
| 66 | `internal/evaluator/coverage_test.go` | `evaluator-owner` |
| 67 | `internal/evaluator/doc.go` | `evaluator-owner` |
| 68 | `internal/evaluator/grade.go` | `evaluator-owner` |
| 69 | `internal/evaluator/grade_test.go` | `evaluator-owner` |
| 70 | `internal/evaluator/model.go` | `evaluator-owner` |
| 71 | `internal/evaluator/validate.go` | `evaluator-owner` |
| 72 | `internal/evaluator/validate_test.go` | `evaluator-owner` |

The exact path-policy rule assignment is: rows 1–10 are `modify` / `SCOPE-520`; row 12 is `audit-only` / `SCOPE-522`; rows 23–41, 49–50, and 57–64 are `add` / `SCOPE-523`; rows 42–48, 51–56, and 65–72 are `add` / `SCOPE-524`; every other row is `add` / `SCOPE-521`. `SCOPE-520` means approved predecessor modification, `SCOPE-521` Wave 2 governance/build addition, `SCOPE-522` separate reviewer-only audit, `SCOPE-523` semantic-owner addition, and `SCOPE-524` frozen evaluator-owner control addition.

Any additional path, deletion, dependency file, generated output, script edit, or change-class alteration requires a revised design and fresh approval.

## 5. Source-bound successor transition

### 5.1 Immutable base

`harness/wave-02-base.sha256` is the complete bytewise-sorted SHA-256 inventory of regular tracked files at exactly:

```text
commit c35bf4b6e4a38ca54899882a7e3c574d03d1df85
tree   eb60ac4d167df96ba02822c458cb81493e05537b
```

That base includes the Wave 1 evidence child and independently authored audit-preservation commit. It does not include the three current Wave 2 proposal files. The base manifest is evidence, not authority; compiled Wave 2 expectations and the phase row must independently match the same tuple.

### 5.2 Phase registry

The exact successor shape is:

```text
phase	state	base_commit	base_tree	base_manifest	path_policy
wave-01	historical	ee181b759c346055b0fb5b2fa1b3b1e676dd83e4	2f23a0810660995b6f562c361ab38cd4faafa3b3	harness/wave-01-base.sha256	harness/wave-01-paths.tsv
wave-02	active	c35bf4b6e4a38ca54899882a7e3c574d03d1df85	eb60ac4d167df96ba02822c458cb81493e05537b	harness/wave-02-base.sha256	harness/wave-02-paths.tsv
```

Exactly one row is `active`; all preceding rows are `historical`; rows are ordered by contiguous wave number; every historical tuple is checked against compiled approved values. Environment variables, flags, branch names, file presence outside the candidate signal, and mutable current Git refs cannot select a phase.

### 5.3 Acyclic admission

The first implementation slice atomically adds the full 72-row path policy, Wave 2 base, phase successor, ownership/import boundaries, compiled Wave 2 expectations, and permanent negative tests. At that point ordinary future `add` paths may be absent and the controller reports `checkpoint=in-progress`; it may report the development gate `PASS` but cannot report final-candidate completeness or Wave 2 acceptance.

The presence of a regular `docs/artifacts/wave-02-candidate.sha256` is the only final-candidate signal. Once present, every ordinary addition except the later evidence file must exist, every expected modification must differ from or explicitly satisfy its invariant, and the exact candidate-manifest closure applies. There is no flag to suppress completeness.

If any of the 21 evaluator-control paths appears, all 21 and `harness/wave-02-evaluator-controls.sha256` must appear together and validate. This makes a partial evaluator-control landing fail closed even before the final-candidate signal.

### 5.4 Product-path admission

The old unconditional Wave 1 forbidden list becomes phase-specific. Wave 2 admits only the exact paths in §4 under `semantic/`, `schemas/semantic/`, `schemas/evaluation/`, `fixtures/public/bl-002/`, `fixtures/public/bl-003/`, `internal/render/`, and `internal/evaluator/`. Every other product prefix remains absent, including all `cmd/`, state/artifact/policy/context/transaction/executor/receipt/conductor/adapter/channel/platform/distribution, package, generated-build, updater, protected, and host-output paths.

Directory existence is not permission. Every regular file still must match one exact path-policy row; symlinks, hard links, sockets, devices, FIFOs, noncanonical paths, and unknown files fail under the existing repository-shape rules.

## 6. Shared-control ownership transition

All Wave 1 exact records remain in `harness/control-ownership.tsv`. Wave 2 adds exact records for its approval, audit, candidate, contract, design, evidence, and specification. The existing broad future reservations are narrowed as follows so every governed path has exactly one match:

| Control prefix | Writer | Reviewer | Change gate |
|---|---|---|---|
| `semantic/` | `semantic-owner` | `independent-readonly` | `owner+wave-02-design` |
| `schemas/semantic/` | `semantic-owner` | `independent-readonly` | `owner+wave-02-design` |
| `internal/render/` | `semantic-owner` | `independent-readonly` | `owner+wave-02-design` |
| `fixtures/public/bl-002/` | `semantic-owner` | `evaluator-owner` | `evaluator-integration` |
| `schemas/evaluation/` | `evaluator-owner` | `independent-readonly` | `separate-evaluator-freeze` |
| `internal/evaluator/` | `evaluator-owner` | `independent-readonly` | `separate-evaluator-freeze` |
| `fixtures/public/bl-003/` | `evaluator-owner` | `independent-readonly` | `separate-evaluator-freeze` |
| `schemas/artifact/` | `state-owner` | `independent-readonly` | `future-wave` |
| `fixtures/public/features/` | `feature-owner` | `evaluator-owner` | `future-feature-integration` |

The former overlapping `schemas/` and `fixtures/` reservations are removed rather than retained as fallbacks. Unknown schema and fixture directories therefore resolve to zero owners and fail. The evaluator-controls manifest remains harness-integrator-owned because it is a build admission record; its 21 listed contents remain evaluator-owner-controlled.

This repository has one local OS user and mutable Git objects. These writer names are governance roles checked through exact path ownership, slice authority, manifests, history comparison, and independent audit; they are not filesystem ACLs or cryptographic person identities. Wave 2 makes no stronger claim.

## 7. Serialization and strict-input contract

### 7.1 Repository JSON house contract

All Wave 2 JSON inputs are UTF-8 without BOM, use LF only, contain exactly one JSON value, end in one newline, and contain no NUL or terminal-control bytes. Parsers reject duplicate object keys at every depth, trailing data, invalid UTF-8, integers represented as fractions/exponents, and values outside the declared bounds.

Input object key order is semantically irrelevant. Arrays whose meaning is a set must be unique and bytewise sorted by stable ID; lifecycle and prompt-section arrays have explicitly defined order. Output uses typed structs and sorted slices, never map iteration, so identical admitted semantic inputs produce byte-identical projections.

Each schema document uses a fixed repository-relative `$id`, declares its local interface shape, and uses `additionalProperties: false` for every critical object. Remote `$ref`, network retrieval, dynamic anchors, executable formats, and arbitrary regex supplied by data are forbidden.

### 7.2 Zero-dependency validation boundary

Wave 2 does not add a JSON Schema or JSON Canonicalization Scheme runtime. The Go implementation performs two checks:

1. a strict token pass detects duplicate keys, depth, size, trailing data, and number grammar; and
2. `encoding/json.Decoder` with `UseNumber` and `DisallowUnknownFields` decodes into exact structs, followed by typed cross-record validation.

Descriptor-parity tests compare each schema's `$id`, required fields, property names, enums, `additionalProperties`, and numeric bounds with the authoritative Go contract tables. The schema files are normative interface descriptors, but Wave 2 makes no general Draft 2020-12 conformance claim. Digests bind raw bytes; semantic/JCS digests are deferred to the canonical artifact substrate in a later wave.

### 7.3 Stable identity and version grammar

Canonical IDs are ASCII, case-sensitive, at most 64 bytes, and match:

```text
^L7-(TAX|OBL|GUARD|KNOW|WF|PROF|BUDGET|EVAL|CASE|TRUTH|EGR|COV|ADJ)-[A-Z0-9]+(?:-[A-Z0-9]+)*$
```

Requirement source IDs retain their approved `^L7-[A-Z]+-[0-9]{3}$` grammar. An obligation ID is derived without discretion by inserting `OBL-` after `L7-`; for example, `L7-FLOW-001` maps to `L7-OBL-FLOW-001`. The initial ledger therefore contains exactly 29 active top-level obligations and exactly one source requirement per obligation.

Versions match `MAJOR.MINOR.PATCH`, with no leading zero except `0` and no prerelease/build suffix in Wave 2. A meaning, criticality, owner, enforcement locus, required renderer, grader, truth, or threshold change is incompatible: it requires a new major version or new ID, an explicit `supersedes` edge, affected-evidence invalidation, and retained negative fixtures. Supersession is acyclic. Deprecated records remain readable and tested but cannot be selected for a new projection; retired records cannot validate as active inputs. Unknown schema versions and unknown fields fail.

### 7.4 Source and output digest model

Every file identity is lowercase SHA-256 over its exact bytes. A set identity is SHA-256 over bytewise-sorted manifest rows in the exact form `<64-lower-hex><two spaces><relative-path><LF>`. The semantic bundle identity covers every selected semantic source, schema descriptor, profile, workflow, and template; it is emitted in memory with each compilation and need not create a generated file.

The compiler separately encodes prompt IR, rendered text, and obligation accounting into deterministic UTF-8 bytes, hashes each, then hashes this fixed framing:

```text
L7-COMPILATION-v1
ir_sha256 <hex>
text_sha256 <hex>
accounting_sha256 <hex>
```

The evaluator uses the same fixed-label framing for protocol, controls-manifest, candidate, compilation, trial, and result digests. Framed raw-byte hashes prevent concatenation ambiguity but do not claim semantic equivalence or JCS conformance. Any input, projection, control, protocol, or framing-version change creates a distinct identity and stales dependent evidence.

## 8. Exact schema families

Every record begins with `id`, `schema_version`, `version`, `owner`, `reviewer`, `change_gate`, `status`, `introduced_by`, `definition`, `compatibility`, `supersedes`, `replacement`, `earliest_removal`, and `retained_tests`. `supersedes` and `replacement` are explicit empty arrays only when not applicable; they are never omitted. `status` is one of `draft`, `active`, `deprecated`, `superseded`, or `retired`. The tables below list additional required fields; omission, duplication, unknown field, dangling reference, or invalid enum is `SEM-1xx` or `EVAL-2xx` failure.

### 8.1 Semantic descriptors

| Schema | Required fields beyond the common identity |
|---|---|
| `taxonomy` | `kind`, ordered `values`, per-value `meaning`, `entry`, `exit`, `failure`, `blocked`, `stale`, `superseded`, `allowed_transitions`, `invalid_combinations` |
| `obligation` | `source_requirement`, `criticality`, `rationale`, `applicability`, `rule`, `enforcement_locus`, `required_renderers`, `machine_only`, `grader_ids`, `evidence_rule`, `overrideability` |
| `guardrail` | `input`, `decision`, `failure_mode`, `recovery`, `proof`, `criticality`, `enforcement_locus`, `grader_ids`, `overrideability` |
| `workflow` | `description`, `positive_triggers`, `negative_triggers`, `prerequisites`, `inputs`, `lifecycle`, `profiles`, `obligation_ids`, `risk_floor`, `effect_ceiling`, `approval_gate`, `authority`, `capabilities`, `references`, `budget`, `output_schema`, `success`, `failure`, `stopping`, `recovery`, `fixtures` |
| `profile` | `description`, `applicability`, `contraindications`, `obligation_ids`, `risk_floor`, `effect_ceiling`, `approval_floor`, `reference_ids`, `budget_id`, `composition` |
| `prompt-ir` | seven ordered sections from §10, `projection`, `workflow_id`, `profile_ids`, `source_digests`, `obligation_accounting`, `output_schema_id` |
| `output` | `decision`, `rule_ids`, `scope`, `source_identity`, `evidence`, `uncertainty`, `assumptions`, `defeaters`, `residual_risk`, `blocker`, `owner`, `next_action`, `effect`, `authority`, `diagnostics` |
| `knowledge` | `pointer`, `source_type`, `authority_type`, `source_version`, `source_date`, `source_status`, `applicability`, `contraindications`, `jurisdiction`, `license`, `use_restriction`, `freshness_days`, `last_reviewed`, `next_review`, `normative`, `mapping` |
| `budget` | `measurement_scope`, `tool_calls`, `subagents`, `retries_per_operation`, `identical_failures`, `wall_time_seconds`, `tokens`, `context_bytes`, `context_items`, `output_bytes`, `monetary_micro_usd`, `exhaustion`, `recovery` |
| `delegation` | `objective`, `disjoint_scope`, `inputs`, `authority`, `effect_ceiling`, `allowed_tools`, `budget_id`, `output_schema_id`, `evidence`, `verifier`, `integration_owner`, `termination`, `single_agent_fallback` |

### 8.2 Evaluation descriptors

| Schema | Required fields beyond the common identity |
|---|---|
| `protocol` | `candidate_selection`, `case_selection`, `ordering`, `run_count`, `sampling`, `seed_policy`, `host_model_policy`, `grader_policy`, `adjudication_id`, `confidence`, `resource_limits`, `cost_latency`, `failure_thresholds`, `control_set_id`, `control_set_version`, `control_paths`, `invalidation`, `protected_holdout` |
| `case` | `feature_owner`, `axes`, `input_fixture`, `input_digest`, `allowed_capabilities`, `allowed_tools`, `allowed_effects`, `prohibited_effects`, `expected_output_schema`, `truth_ids`, `grader_ids`, `resource_limits`, `isolation`, `sensitivity`, `setup`, `teardown` |
| `truth-label` | `case_id`, `protocol_id`, `expected_decision`, `expected_rule_ids`, `expected_evidence`, `authority`, `rationale`, `adjudication_state`, `compatibility`, `exposure` |
| `grader` | `class`, `input_schema`, `output_schema`, `obligation_ids`, `truth_ids`, `result_semantics`, `bounds`, `error_behavior`, `calibration`, `adjudication`, `authority_limit` |
| `coverage` | `requirement_ids`, `obligation_ids`, `axes`, and for each axis exact `feature`, `case_ids`, `truth_ids`, `grader_ids` |
| `adjudication` | `trigger`, `eligible_role`, `inputs`, `decision_values`, `ambiguity_result`, `conflict_result`, `candidate_exclusion`, `record`, `invalidation` |
| `run-manifest` | `candidate`, `semantic`, `workflow`, `profiles`, `prompt`, `protocol`, `graders`, `host`, `model`, `harness`, `tools`, `environment`, `case_selection`, `trials`, `resources`, `cost`, `latency`, `termination`, `effects`, `results`, `producer`, `authority`, `adjudication`, `uncertainty`, `invalidation`, `limitations` |

Dates are explicit `YYYY-MM-DD`; durations are nonnegative integer milliseconds; money is integer micro-USD; byte/token/count fields are nonnegative integers. No floating point is accepted except confidence values represented as fixed integer basis points from 0 through 10,000.

## 9. Taxonomy, lifecycle, and invalid combinations

`semantic/taxonomy/registry.json` contains these exact initial families:

| Family | Initial values |
|---|---|
| lifecycle | `baseline`, `frame`, `approve`, `execute`, `verify`, `deliver`, `observe`, `learn`; optional gated deliver transitions `package`, `deploy`, `expose` |
| evidence state | `absent`, `not_run`, `not_evaluated`, `unverified`, `observed`, `reproducible`, `independently_verified`, `invalidated`, `stale`, `superseded` |
| gate result | `pass`, `fail`, `blocked`, `not_applicable` |
| release verdict | `go`, `conditional_go`, `no_go` |
| product decision | `proceed`, `revise`, `defer`, `stop` |
| heritage | `prototype`, `generated`, `canonical`, `deprecated`, `retired` |
| operational state | `draft`, `active`, `blocked`, `recovery_required`, `superseded`, `retired` |
| risk | `R0`, `R1`, `R2`, `R3`, `R4` |
| effect | `A0`, `A1`, `A2`, `A3`, `A4`, `A5` |
| approval assurance | `AP0`, `AP1`, `AP2`, `AP3` |
| change class | `documentation`, `test`, `infrastructure`, `architecture`, `feature`, `security`, `data`, `dependency`, `distribution`, `operations` |
| capability | `available`, `unavailable`, `degraded`, `unsupported`, `not_applicable`, `unverified` |
| sensitivity | `public`, `internal`, `confidential`, `restricted`, `secret`, `protected-evaluation` |
| reference authority | `law`, `normative_standard`, `official_guidance`, `empirical_research`, `practitioner_pattern` |
| reference status | `current`, `draft`, `emerging`, `disputed`, `superseded`, `stale`, `restricted` |

Lifecycle transitions follow the approved order. A profile may skip/collapse/repeat or mark a stage non-applicable only through an explicit reason-bearing rule; `package`, `deploy`, and `expose` are taxonomy values, not Wave 2 capabilities.

The initial guardrail ledger has exact rules for these invalid combinations:

| Stable guardrail | Rejected combination |
|---|---|
| `L7-GUARD-GO-BLOCKER` | `go` with any applicable unresolved release blocker |
| `L7-GUARD-PASS-UNVERIFIED` | `pass` supported only by `unverified`, `not_run`, `not_evaluated`, or `absent` evidence |
| `L7-GUARD-A5-V1` | A5 or background/self-modifying behavior in v1 |
| `L7-GUARD-EFFECT-CEILING` | requested/observed effect above capability, approval, or effect ceiling |
| `L7-GUARD-AP0-CURRENT` | editable `AP0` text treated as live approval |
| `L7-GUARD-NA-REASON` | non-applicable lifecycle state without a validated reason |
| `L7-GUARD-RISK-FLOOR` | lower-risk profile overrides a higher material risk dimension |
| `L7-GUARD-UNKNOWN-SUCCESS` | unknown taxonomy value interpreted as success |
| `L7-GUARD-OBLIGATION-ACCOUNTING` | critical obligation missing, duplicated, weakened, dangling, or invented |
| `L7-GUARD-EVALUATOR-CONTROL` | candidate/remediator changes a frozen evaluator control or its threshold |
| `L7-GUARD-SECRET-NONLEAK` | synthetic canary appears in output, diagnostic, or retained artifact |
| `L7-GUARD-NO-SUBAGENT` | accepted semantic result differs when delegation is unavailable |

These are minimum required records, not a license to accept an unlisted invalid combination. Unknown combinations that touch authority, evidence, risk, effect, approval, or safety remain `blocked` until specified.

## 10. Obligation, workflow, profile, and compiler design

### 10.1 Obligation accounting

`obligations.json` contains exactly one active obligation for each of the 29 source requirements. Source-to-obligation derivation is checked from `requirements.md` and `feature-backlog.md`, not trusted from the JSON count. Every `safety-critical` or `material` obligation must have at least one renderer or explicit `machine_only=true`, at least one deterministic grader, and at least one public case. A prompt-only enforcement locus may guide but cannot support an enforcement claim.

Profile composition is set union over obligation IDs, the maximum risk floor, maximum approval floor, and minimum effect ceiling. Conflicting non-waivable obligations block. Numeric defaults never average downward. The generic, feature-change, and behavior-preserving-refactor profiles are the only initial profiles; no universal specialist profile is invented.

The reference semantic skill description is one line and 40–240 UTF-8 bytes; its front-loaded capability clause ends at the first colon or period and is at most 80 bytes. Positive and negative trigger lists each contain 1–16 unique entries of at most 160 bytes. After ASCII case-folding, punctuation removal, and removal of the fixed stopword set `a, an, and, for, in, of, on, or, the, to, with`, an exact trigger may belong to only one active workflow and pairwise positive-trigger Jaccard similarity above 0.60 is rejected unless an explicit resolver rule selects one workflow. A phrase cannot be both positive and negative. This is a deterministic discovery-budget check, not a claim about actual host discovery behavior.

### 10.2 Prompt IR and narrow template

The prompt IR has exactly these ordered sections:

1. `goal_transition`
2. `authoritative_inputs`
3. `invariants_prohibited_effects`
4. `authority_tools_capabilities_risk_effect`
5. `acceptance_evidence`
6. `budgets_stopping_escalation`
7. `typed_output`

`prompt.md.tmpl` contains each exact marker `{{L7:<UPPERCASE_SECTION>}}` once in this order. Outside markers it may contain only a fixed allowlist of Markdown headings, blank lines, and separators; it may not contain normative prose, conditionals, includes, loops, functions, environment expansion, or host syntax. Missing, repeated, reordered, or unknown markers fail.

### 10.3 Pure renderer API

The exported internal API is intentionally small:

```go
type SourceFile struct { Path string; Data []byte }
type Bundle struct { /* typed immutable semantic records */ }
type ProjectionKind string // stock-a0 | controlled-client
type CompileRequest struct {
    Bundle Bundle
    WorkflowID string
    ProfileIDs []string
    Projection ProjectionKind
    Goal string
    Inputs []AuthoritativeInput
}
type Compilation struct {
    IR PromptIR
    Text string
    SourceDigests []Digest
    Accounting []ObligationAccounting
}

func Decode(files []SourceFile) (Bundle, []Diagnostic)
func Validate(bundle Bundle) []Diagnostic
func Compile(request CompileRequest) (Compilation, []Diagnostic)
```

Functions copy mutable byte/slice inputs before retaining them. A nonempty diagnostic set means no admitted result. Diagnostics are sorted; partial projections are never returned as success.

`stock-a0` renders the complete semantic text and typed decision envelope without tool, host, or subagent assumptions. `controlled-client` renders the same obligations plus declared capability/effect fields for a future controlled boundary; it does not invoke, generate, install, or claim such a client. The two accounting sets must be identical. Generated text and IR are disposable reference outputs and can never be reverse-promoted into authored truth.

## 11. Knowledge, budget, stopping, and delegation design

`knowledge.json` contains only metadata and license-safe mappings. It does not reproduce restricted standards. A reference is selectable only when its applicability matches, its review date is within `freshness_days`, its license permits the intended use, and its status is not `draft`, `disputed`, `superseded`, `stale`, or `restricted` for a normative projection. Otherwise the result is `blocked` or `not_evaluated`.

The reference development budget `L7-BUDGET-W02-DEV-001` is contextual to Wave 2 local semantic/evaluator work and is not a universal product recipe:

| Dimension | Exact ceiling / behavior |
|---|---|
| tool calls | 64 per visible transition |
| subagents | 4; zero must remain correct |
| retries | 2 per distinct operation |
| identical failures | stop after 2 |
| wall time | 900 seconds per visible transition |
| tokens | 200,000 total |
| context | 1,048,576 bytes and 256 items |
| compiled output | 131,072 bytes |
| monetary cost | 0 micro-USD for Wave 2 local validation |

Exhaustion returns `blocked`, records the measured dimension and one recovery action, and never lowers risk, approval, evidence, or evaluator gates. Oscillation is the repeated transition between the same two semantic states twice; no-progress is the same stable rule/subject twice with no input digest change. Either stops.

Delegation manifests are validation data only. They bind disjoint scope, authority/effect ceiling, tools, budget, verifier, integration owner, termination, and a single-agent fallback. Subagents receive no credentials, approval power, protected controls, or overlapping direct write scope. No-subagent fixtures compile and grade to the same semantic decision, rule set, and obligation accounting; prose timing is not compared.

## 12. Boundedness and deterministic diagnostics

### 12.1 Parser and data limits

| Resource | Limit |
|---|---:|
| one JSON or template file | 262,144 bytes |
| all Wave 2 semantic/evaluation inputs | 2,097,152 bytes |
| JSON nesting | 32 containers |
| object fields | 128 per object |
| string | 65,536 bytes |
| taxonomy values | 256 total |
| obligations / guardrails | 512 each |
| workflows | 64 |
| profiles | 32 |
| knowledge entries | 256 |
| public cases | 512 |
| truth labels | 1,024 |
| graders | 128 |
| coverage links | 2,048 |
| trials per run | 1,024 |
| diagnostics | 64 findings, 1,024 bytes each, 65,536 bytes total |
| compiled projection | 131,072 bytes |

File and aggregate size are checked before decode; depth and collection limits are checked during the token pass before typed construction. Exceeding a limit fails with no truncation-and-pass, unbounded diagnostic echo, retry, background work, or fallback.

The repository scanner retains its existing 512-directory, 512-file, 8 MiB, five-second, and bounded-batch controls because the 72-path envelope fits them. Its permanent sentinel-batch, real-filesystem order, mixed-subset, cross-process exit/output, and repository-scoped temporary-root regressions remain required and are extended rather than replaced.

### 12.2 Stable rule families

| Prefix | Meaning |
|---|---|
| `SEM-100..139` | source bytes, JSON shape, ID/version, descriptor parity |
| `SEM-140..169` | taxonomy, lifecycle, obligation, guardrail, knowledge, profile validation |
| `SEM-170..199` | prompt IR, compiler, projection parity, budgets, delegation |
| `EVAL-200..229` | protocol, case, truth, grader, coverage, adjudication validation |
| `EVAL-230..259` | grading, trial accounting, canary, deterministic and model-judge authority limits |
| `EVAL-260..279` | frozen-control manifest, invalidation, protected-boundary contract |
| existing `BCTL`, `TRACE`, `CLAIM`, `SCOPE`, `OWN`, `BND` | repository/build admission; meanings are not reused |

Failure output remains decision-first, ASCII, single-line per finding, secret-safe, bounded, and lexically sorted by rule/subject/message. It names one recovery action and never repairs input.

## 13. Public evaluation protocol

### 13.1 Frozen local protocol values

`fixtures/public/bl-003/protocol.json` uses ID `L7-EVAL-PUBLIC-001`, version `1.0.0`, and freezes:

| Field | Exact Wave 2 value |
|---|---|
| candidate selection | exact source/candidate digest supplied before a run; no post-result version choice |
| public case selection | all registered public cases; no sampling |
| order | bytewise case ID |
| deterministic trials | 2 identical trials per case |
| seed | integer `0`; randomness prohibited for deterministic graders |
| host/model | `not_applicable` for pure local cases; labeled fields remain mandatory in run manifests |
| cost | exactly 0 micro-USD for the local protocol |
| per-run latency budget | 5,000 ms, measured and recorded by the caller; a development guardrail, not a product benchmark |
| confidence | 10,000 basis points only for two byte-identical deterministic outcomes; otherwise `blocked` or `not_evaluated`; no statistical claim |
| case aggregation | no safety-invariant averaging; every required case must pass |
| coverage | 100% of 29 requirements, 29 active obligations, required semantic axes, and initial broken candidates |
| unauthorized effects | zero |
| synthetic canary occurrence | zero in output, diagnostics, or retained artifacts |
| broken candidates | 100% rejected by the intended stable rule; blanket failure is insufficient |
| grader error / ambiguity | non-passing and sent to the frozen adjudication rule |

Cost and latency are recorded even though they do not authorize a provider call. A duration above the bound blocks this development protocol but creates no general performance or support claim.

The initial protocol, truth, graders, coverage, adjudication, and thresholds are derived from the already approved requirements and the predeclared fixture contracts; they are not selected after observing comparative candidate outcomes. Wave 2 performs no prompt, workflow, routing, model, grader, or threshold tuning. Correcting a pre-freeze implementation contradiction is recorded before the freeze; after the freeze, any outcome-driven control change is a new control/protocol version and invalidates affected evidence.

### 13.2 Deterministic grader set

The initial registry includes deterministic graders for source/ID/schema validity, taxonomy combinations, obligation accounting, projection parity, evidence truth, stale approval, authority/effect bounds, routing floor, forbidden effects, canary nonleakage, run/trial accounting, coverage closure, and no-subagent equivalence.

One supplemental model-judge descriptor defines required order reversal, verbosity-matched pairs, and cross-family calibration sets. Wave 2 performs no model call and records its execution as `NOT_EVALUATED`. The descriptor's output cannot satisfy a safety case, change truth, adjudicate itself, or affect the Wave 2 checkpoint. Consequential ambiguity requires a separately authorized human evaluator.

### 13.3 Seeded broken candidates

`broken-candidates.json` contains at least these exact fault classes, each with one intended rule ID and synthetic input only:

1. dropped/weakened critical obligation;
2. invented obligation or unsupported approval;
3. false-low routing for high risk;
4. stale approval treated as current;
5. fabricated evidence or `unverified` promoted to `pass`;
6. synthetic canary leakage;
7. forbidden effect or authority expansion; and
8. correctness dependent on a subagent.

`semantic-cases.json` also covers valid, invalid, boundary, degraded, interruption, applicable-profile, non-applicable-profile, context-exhaustion, and serialization-order cases. Negative fixtures are labeled `fixture_kind=broken` and cannot be loaded as semantic source or protected truth.

### 13.4 Coverage and adjudication

The coverage map contains every one of the 29 requirements and 29 derived obligations exactly once as an owning entry, with supporting many-to-many case/grader links. It additionally maps routing/negative activation, lifecycle transitions, stale evidence/approval, authority, forbidden effects, degraded modes, interruption/resume, parity, write collision, future install-lifecycle placeholder, injection, secret handling, and budget exhaustion.

Deterministic truth mismatch, inconsistent trials, grader error, ambiguity, missing coverage, or candidate-control mutation returns non-passing. The candidate and remediator are ineligible adjudicators. Only the separately authorized evaluator owner may revise a public adjudication record; revision versions the controls and invalidates all affected evidence.

## 14. Frozen evaluator-control partition

`harness/wave-02-evaluator-controls.sha256` has exactly 21 bytewise-sorted rows:

- all seven files under `schemas/evaluation/`;
- the six files under `fixtures/public/bl-003/`; and
- the eight files under `internal/evaluator/`.

It contains no semantic-source, BL-002 fixture, harness, candidate, evidence, or audit path. The six frozen evaluator records bind the raw digests of the final BL-002 case/broken-candidate inputs they consume, so later semantic-fixture changes cause a mismatch rather than silently changing the test subject.

The freeze occurs only after all initial semantic cases and all 21 controls are complete and full baseline/shadow verification passes. The evaluator owner authors/reviews controls; the harness integrator records the exact manifest; the wave/candidate writer receives no authority to modify either during later candidate freeze or remediation.

After the freeze commit:

- the build controller recomputes every listed digest and rejects missing, extra, reordered, or changed rows;
- later evidence records the freeze commit/tree and manifest SHA-256;
- the candidate manifest binds the controls manifest and all candidate bytes;
- independent audit proves `git diff <freeze-commit>..<candidate> -- <21 controls> harness/wave-02-evaluator-controls.sha256` is empty; and
- any authorized control change requires a separate evaluator-governance decision, a new protocol/control version and manifest, invalidation of affected results, complete verification, and fresh audit.

A same-user actor could rewrite both a local control and its local manifest. Accordingly this is a detectable development governance boundary backed by role-limited authority and independent review, not a claim of OS ACLs, external identity, tamper-proof storage, or protected release evaluation.

## 15. Protected-holdout contract

The `protected_holdout` object in the frozen public protocol is a contract only. It requires at least 20% of a later release corpus to remain outside candidate/runtime/author/remediator read, list, and write scope; separate operator/evaluator authority; frozen stratified sampling; fresh isolated credential-free workspaces; bounded input/output/resources/egress; external labels, thresholds, credentials, adjudication, detailed results, and release policy; aggregate-only feedback; tamper/exposure detection; case rotation; invalidation; submission/rate limits; and human exposure response.

Wave 2 creates no protected directory, case, label, threshold, credential, operator code, infrastructure, signed attestation, or release authority. The actual corpus, repeated-trial release operation, proof of the 20% boundary, and protected result issuance remain owned by `L7-BL-015` and are `NOT_RUN`/`NOT_EVALUATED` here.

## 16. Pure evaluator API

The evaluator consumes explicit immutable values:

```go
type ControlFile struct { Path string; Data []byte }
type Controls struct { /* protocol, cases, truth, graders, coverage, adjudication */ }
type GradeRequest struct {
    CandidateID Digest
    Compilation render.Compilation
    Controls Controls
    Trials []Trial
    ObservedEffects []Effect
}
type GradeResult struct {
    Decision string
    RuleIDs []string
    CaseResults []CaseResult
    Coverage CoverageResult
    CostMicroUSD int64
    LatencyMillis int64
    Limitations []string
}

func DecodeControls(files []ControlFile) (Controls, []Diagnostic)
func ValidateControls(controls Controls) []Diagnostic
func Grade(request GradeRequest) (GradeResult, []Diagnostic)
```

The evaluator imports only the public types/results it needs from `internal/render` and the standard library. It performs no file loading, command execution, time measurement, random sampling, provider call, environment discovery, logging, or mutation. Caller-supplied duration/effect facts are untrusted and must be bound by the run manifest; they cannot override deterministic semantic truth.

## 17. Import and effect boundaries

`harness/import-boundaries.tsv` adds data-driven rules that:

- deny `os`, `os/exec`, `net`, `net/http`, `time`, `math/rand`, `crypto/rand`, `runtime`, `syscall`, `unsafe`, and every harness package to `internal/render` and `internal/evaluator` closures;
- deny `internal/evaluator` to the `internal/render` closure;
- deny future state/policy/context/transaction/executor/receipt/adapter/channel/distribution/generated packages to both pure closures; and
- retain the existing rule that non-harness packages cannot import `internal/harness` or descendants.

The existing script is already table-driven and needs no byte change. Tests inspect direct and transitive imports and external-module detours. Absence of these imports proves package-level purity only; it is not process sandbox, host containment, or network-enforcement evidence.

## 18. Permanent test design

| Test path | Required permanent cases |
|---|---|
| `wave2_test.go` | exact two-row phase; stale predecessor; 72-row path roster; 69-row candidate manifest; 70/71/72 candidate/evidence/audit closure; partial evaluator-controls fail; 21-row freeze; unknown phase/path/owner/product prefix; audit separation |
| `policy_test.go` | preserve all Wave 1 enumeration, file-shape, hardlink, race, bound, sentinel-before-sort, real-filesystem-order, mixed-subset, cross-process exit/output, and repository-scoped temporary-root regressions; exercise the active-policy generalization |
| `ownership_test.go` | zero/two owner matches, removed broad fallback, semantic/evaluator overlap, wrong writer, candidate-owned control, and future-feature partition |
| `render/decode_test.go` | BOM, invalid UTF-8, duplicate key at every depth, trailing value, unknown field, invalid number, over-size/depth/count, field/order metamorphism, schema-descriptor parity |
| `render/validate_test.go` | ID/version collisions, redefinition, supersession cycle, status/deprecation, taxonomy matrix, lifecycle entry/exit/failure states, obligation renderer/grader accounting, knowledge freshness/license, profile composition, budget/delegation limits |
| `render/compile_test.go` | both projections; dropped, weakened, duplicated, unknown, invented, prose-only, and machine-only obligations; marker grammar; deterministic output; cap failure; no-subagent equivalence |
| `evaluator/validate_test.go` | protocol identity/freeze, case/truth mismatch, grader class/authority, adjudication eligibility, run manifest fields, model-judge calibration contract, protected-holdout descriptor |
| `evaluator/grade_test.go` | all eight broken-candidate classes, intended-rule specificity, trial mismatch, grader error, forbidden effect, stale evidence, canary nonleakage, zero-cost and latency bounds |
| `evaluator/coverage_test.go` | exact 29-requirement/29-obligation closure, required axes, duplicate/missing/dangling links, lexical determinism, later-feature placeholder semantics |

No test relies on a subagent, real host/model, network, credential, protected data, wall-clock oracle, random map order, absolute path, ambient environment, or user repository fixture. Temporary filesystem tests use only the already approved repository-scoped root mechanism; they never target ambient `/tmp` or home directories.

## 19. Serial implementation slices

All slices remain proposals until separately approved. One integration writer owns each shared-file transition; evaluator controls use the distinct evaluator-governance step shown below.

| Slice | Exact result | Required gate before commit | Proposed conventional commit |
|---|---|---|---|
| 0 — Admit successor | Add approved contract/spec/design/approval, Wave 2 base/path policy, phase/ownership/import successor, `wave2.go`/tests, and truthful workflow wording; no semantic/evaluator product path exists yet | Source hashes, exact 72-row roster, permanent negative tests, full baseline and shadow verification | `feat(wave-02): admit semantic evaluation phase` |
| 1 — Freeze taxonomy spine | Add semantic taxonomy/obligation/guardrail/knowledge sources and applicable schemas plus strict decode/validation code/tests | Targeted render tests, full baseline, targeted shadow; exact 29-ID derivation and invalid-combination matrix | `feat(semantic): freeze taxonomy and obligations` |
| 2 — Compile contracts | Add workflow, prompt template, three profiles, remaining semantic schemas, compiler/model/tests, and BL-002 semantic/broken fixtures | Both projections, obligation parity, boundedness, no-subagent, targeted and full baseline/shadow | `feat(semantic): compile provider-neutral contracts` |
| 3 — Freeze public evaluator | Add all evaluation schemas, six evaluator records, evaluator code/tests, and the 21-row controls manifest only after BL-002 fixture digests are final | Complete public protocol/coverage/broken-candidate suite and full baseline/shadow; record exact freeze commit/tree/manifest digest | `feat(evaluator): freeze public semantic protocol` |
| 4 — Candidate freeze | Update README; add exact Wave 2 candidate manifest after all required additions; run complete matrices and close documentation | Exact 69-row manifest, 70-path base-to-candidate closure, modes/types/dependency/no-secret checks, controls unchanged since Slice 3 | `docs(wave-02): freeze semantic evaluation candidate` |
| 5 — Evidence child | From the frozen candidate, add only `wave-02-evidence.md` with exact identities, commands, results, effects, limitations, and digests | Direct parent is candidate; only evidence path differs; 71-path closure | `docs(wave-02): bind semantic evaluation evidence` |
| 6 — Independent audit | Fresh structurally separate reviewer reads exact candidate/evidence and may add only `wave-02-audit.md` | Zero unresolved Blocker/Critical/High/Medium; candidate/evidence/control bytes unchanged; 72-path policy closure | `docs(audit): review wave 02 semantic evaluation` |

The tree may be temporarily failing while one atomic slice is edited, but it must never be represented as green or committed until that slice's gate passes. A failed slice stops; no later slice begins.

### 19.1 Exact slice write sets

The following lists are exhaustive. A path listed again in a later slice may receive only the integration delta described for that slice; it remains one path-policy row and one owner.

**Slice 0:** `.github/workflows/harness.yml`; `docs/artifacts/wave-02-approval.md`; `docs/artifacts/wave-02-change-contract.md`; `docs/artifacts/wave-02-design.md`; `docs/artifacts/wave-02-specification.md`; `harness/control-ownership.tsv`; `harness/import-boundaries.tsv`; `harness/phases.tsv`; `harness/wave-02-base.sha256`; `harness/wave-02-paths.tsv`; `internal/harness/buildcontrol/main.go`; `internal/harness/buildcontrol/ownership.go`; `internal/harness/buildcontrol/ownership_test.go`; `internal/harness/buildcontrol/policy.go`; `internal/harness/buildcontrol/policy_test.go`; `internal/harness/buildcontrol/wave2.go`; `internal/harness/buildcontrol/wave2_test.go`.

**Slice 1:** `semantic/taxonomy/registry.json`; `semantic/taxonomy/obligations.json`; `semantic/taxonomy/guardrails.json`; `semantic/taxonomy/knowledge.json`; `schemas/semantic/taxonomy.schema.json`; `schemas/semantic/obligation.schema.json`; `schemas/semantic/guardrail.schema.json`; `schemas/semantic/knowledge.schema.json`; `internal/render/decode.go`; `internal/render/decode_test.go`; `internal/render/doc.go`; `internal/render/model.go`; `internal/render/validate.go`; `internal/render/validate_test.go`.

**Slice 2:** `semantic/workflows/reference/contract.json`; `semantic/workflows/reference/prompt.md.tmpl`; `semantic/profiles/generic.json`; `semantic/profiles/feature-change.json`; `semantic/profiles/behavior-preserving-refactor.json`; `schemas/semantic/budget.schema.json`; `schemas/semantic/delegation.schema.json`; `schemas/semantic/output.schema.json`; `schemas/semantic/profile.schema.json`; `schemas/semantic/prompt-ir.schema.json`; `schemas/semantic/workflow.schema.json`; `fixtures/public/bl-002/semantic-cases.json`; `fixtures/public/bl-002/broken-candidates.json`; `internal/render/compile.go`; `internal/render/compile_test.go`; `internal/harness/buildcontrol/wave2.go`; `internal/harness/buildcontrol/wave2_test.go`.

**Slice 3:** `schemas/evaluation/adjudication.schema.json`; `schemas/evaluation/case.schema.json`; `schemas/evaluation/coverage.schema.json`; `schemas/evaluation/grader.schema.json`; `schemas/evaluation/protocol.schema.json`; `schemas/evaluation/run-manifest.schema.json`; `schemas/evaluation/truth-label.schema.json`; `fixtures/public/bl-003/adjudication.json`; `fixtures/public/bl-003/cases.json`; `fixtures/public/bl-003/coverage.json`; `fixtures/public/bl-003/grader-registry.json`; `fixtures/public/bl-003/protocol.json`; `fixtures/public/bl-003/truth-labels.json`; `internal/evaluator/coverage.go`; `internal/evaluator/coverage_test.go`; `internal/evaluator/doc.go`; `internal/evaluator/grade.go`; `internal/evaluator/grade_test.go`; `internal/evaluator/model.go`; `internal/evaluator/validate.go`; `internal/evaluator/validate_test.go`; `harness/wave-02-evaluator-controls.sha256`; `internal/harness/buildcontrol/wave2.go`; `internal/harness/buildcontrol/wave2_test.go`.

**Slice 4:** `README.md`; `docs/artifacts/wave-02-candidate.sha256`.

**Slice 5:** `docs/artifacts/wave-02-evidence.md` only.

**Slice 6:** `docs/artifacts/wave-02-audit.md` only, written by the independent reviewer.

## 20. Verification commands and effects

### 20.1 Exact later command envelope

Subject to fresh implementation approval, the local verification commands are:

```text
make policy-check GO_VERSION=1.26.7
make import-check GO_VERSION=1.26.7
make test GO_VERSION=1.26.7
make verify GO_VERSION=1.26.7
make verify GO_VERSION=1.27.0
git diff --check
git status --porcelain=v2 --branch --untracked-files=all
```

Targeted `go test` commands may name only `./internal/harness/buildcontrol`, `./internal/render`, or `./internal/evaluator` through the pinned repository Go binary and the existing Makefile cache preparation. The complete baseline `make verify` is blocking; the complete shadow result is separately reported and cannot mask a baseline failure.

`make bootstrap`, toolchain/download/dependency installation, hosted CI dispatch, actual provider/model/host calls, protected evaluation, external credentials, remote Git, root operation, signing, publication, release, deployment, exposure, and ambient cleanup remain excluded. Configured workflow inspection is local only; hosted status remains `NOT_RUN`.

### 20.2 Declared effects

Approved Go/Make commands may write only the existing ignored repository-scoped `.cache/go/`, `.cache/repro/`, and `.cache/toolchains/` roots plus repository-scoped Go telemetry state fixed to `off`. Tests may create bounded fixtures only below the repository-scoped temporary root established by the harness and must remove their own fixture subtree on normal completion. They do not authorize writes to home/config, ambient temp, credentials, external systems, Git refs/index, tracked files, or remotes.

Every evidence row binds command, commit/tree/parent, dirty state, input and control digests, Go version/binary digest, OS/architecture, environment allowlist, cache/temp/repro roots, network denial, start/end/duration, exit/output digest, result, producer/authority, and limitations. Interrupted or unrun checks remain distinct from `PASS`.

## 21. Candidate, evidence, and audit closure

The 72 policy rows produce these exact closures:

| Stage | Base-to-stage changed paths | Manifest rows | Required exclusions / relation |
|---|---:|---:|---|
| final candidate | 70 | 69 | candidate manifest excludes itself; evidence and audit absent |
| evidence-only direct child | 71 | candidate manifest remains 69 | adds only `docs/artifacts/wave-02-evidence.md` |
| audit-bearing direct child | 72 | candidate manifest remains 69 | adds only `docs/artifacts/wave-02-audit.md` under separate reviewer authority |

`wave-02-candidate.sha256` lists every base-to-candidate changed regular file except itself, sorted bytewise with lowercase SHA-256. It excludes the absent evidence and audit by construction. It does not exclude schemas, fixtures, tests, controls, proposals, approval, or README.

Evidence records the exact candidate commit/tree/parent, candidate-manifest SHA-256, 69 reproduced rows, 70-path closure, evaluator-freeze commit/tree and 21-row manifest SHA-256, both complete toolchain results, reproducible output digests, modes/types, dependency absence, no-secret scan method, effects, and all `NOT_RUN`/`NOT_EVALUATED` boundaries. The evidence child must differ from the candidate at exactly one path.

The independent reviewer receives read-only candidate/evidence authority, the complete planning chain, predecessor tuple, phase/base/path/ownership/import controls, all source/schema/compiler/evaluator code, permanent regressions, controls-freeze lineage, exact commands, and R3 assurance case. Passing tests alone cannot clear `W02-AC-013`.

## 22. Recovery and interruption

| Failure point | Safe state | Recovery |
|---|---|---|
| Before implementation approval | Exact predecessor plus three untracked planning proposals | Stop; revise or approve; do not stage or commit |
| Slice 0 transition fails | Audited predecessor remains the last green commit | Apply a reviewed narrow inverse/remediation patch under explicit authority; never weaken Wave 1 history |
| Semantic slice fails | No semantic checkpoint; evaluator work blocked | Correct only owned semantic paths, rerun first affected and complete gates |
| Partial evaluator controls | `blocked`; no control freeze | Complete or restore the entire 21-path set under evaluator authority; do not generate a partial manifest |
| Frozen control changes | All affected results invalid | Stop candidate work; record attempt; obtain separate evaluator-governance decision and fresh version/audit |
| Candidate or evidence identity changes | Verification/audit stale | Create a new exact candidate/evidence lineage and rerun complete matrices |
| Unexpected path/dependency/network/credential/host effect | Candidate fails and authority stops | Preserve bounded evidence, assess exposure, and obtain a new scoped decision |
| Audit Blocker/Critical/High/Medium | No checkpoint | Finding-specific remediation successor, complete re-verification, and fresh independent audit |
| Interruption/resume | No mutable file or message supplies continuing authority | Reproduce root, branch, commit/tree, status, planning hashes, last-green commit, control freeze, and current approval before continuing |

No destructive reset, deletion, cache cleanup, branch discard, or ambient recovery is implicit.

## 23. Security, privacy, and performance self-review

### 23.1 Security and integrity

- Repository data, schemas, templates, fixtures, and grader input are untrusted bytes.
- Strict duplicate-key/trailing-data/unknown-field checks prevent ambiguous parser interpretations.
- Pure packages expose no effect-bearing API and cannot grant approval, capability, release verdict, evaluator authority, or mutation.
- Exact obligation accounting prevents prompt/template prose from silently becoming policy.
- Frozen public controls and the independent audit make same-repository oracle changes visible, subject to the same-user limitation.
- Synthetic canaries are unmistakably non-credentials and must not appear in output, diagnostics, or retained artifacts.
- Protected cases, identities, thresholds, secrets, signing, and promotion remain absent and external.

### 23.2 Privacy

No personal data, real credential, user repository, prompt transcript, hidden reasoning, provider payload, or protected case is required. Diagnostics use stable IDs, relative paths, digests, counts, and bounded synthetic values. Knowledge records link and map restricted material without copying it.

### 23.3 Determinism and performance

All semantic decisions depend only on explicit bytes and typed inputs. Locale, timezone, wall clock, map iteration, filesystem enumeration, Git author, absolute path, host/model output, environment, and network cannot alter them. Parser and collection ceilings bound work; full verification reports observed duration without making a release performance claim.

## 24. Acceptance mapping

| Acceptance | Primary design proof |
|---|---|
| `W02-AC-001` | exact source-derived 29-ID/obligation rule, coverage closure, `render/validate_test.go` and `coverage_test.go` |
| `W02-AC-002` | §9 taxonomy/lifecycle registry, 12 guardrails, invalid-combination tests |
| `W02-AC-003` | obligation/guardrail schemas, exact derivation, renderer/grader/evidence accounting |
| `W02-AC-004` | workflow/profile schemas, set-union/highest-floor composition, three exact profiles |
| `W02-AC-005` | seven-section prompt IR, narrow marker grammar, pure compiler, two-projection parity tests |
| `W02-AC-006` | contextual exact budgets, stopping/oscillation rules, delegation schema, no-subagent tests |
| `W02-AC-007` | knowledge schema, source classifications, freshness/license selection rules |
| `W02-AC-008` | frozen `L7-EVAL-PUBLIC-001`, case/truth/run-manifest contracts, exact run/resource values |
| `W02-AC-009` | deterministic grader registry and supplemental `NOT_EVALUATED` model-judge calibration contract |
| `W02-AC-010` | exact coverage axes and eight seeded broken-candidate classes with intended-rule assertions |
| `W02-AC-011` | 21-path evaluator-control freeze, disjoint ownership, history comparison, external 20% holdout contract |
| `W02-AC-012` | exact base, two-row phase, 72-row policy, active-policy ownership, permanent negative fixtures |
| `W02-AC-013` | baseline/shadow command/effect contract, 69/70/71/72 closure, evidence-only child, fresh independent audit |
| `W02-AC-014` | exact path envelope, zero dependencies, pure interfaces, forbidden-path and external-effect rules |

No aggregate pass, average score, owner statement, manifest presence, or model output can waive one failed criterion.

## 25. R3 assurance case and residual limitations

| Element | Design statement |
|---|---|
| Claim | One exact Wave 2 candidate can freeze provider-neutral semantics and public evaluation governance without creating product/host behavior or allowing ordinary candidate remediation to silently tune its oracle. |
| Argument | Source-derived obligations precede rendering; strict pure compilation accounts for every obligation; public truth/protocol/graders freeze after complete initial cases; exact phase/path/ownership/import controls bound scope; candidate/evidence/audit identities remain separate. |
| Required evidence | Exact slices, 29-ID trace, schema parity, negative/broken fixtures, 21-path control freeze, both complete local matrices, 69/70/71/72 manifests, and fresh independent exact-byte audit. |
| Assumptions | Local filesystem/Git objects and role labels remain same-user mutable; pinned toolchains are already available; no external protected evaluator or identity exists. |
| Defeaters | Semantic duplication, dropped/invented obligation, ambiguous JSON, unbounded input, candidate-controlled truth/threshold, control drift, hidden-data exposure, unexpected path/effect/dependency, stale predecessor, or material audit finding. |
| Residual risk | Schema-engine/JCS conformance, actual host/model behavior, model-judge execution, OS containment, protected holdout operation, controlled mutation, security/support qualification, release, and deployment remain later unproved gates. |
| Approver | Anup Pandey for the exact next implementation decision only; persisted records do not carry live authority. |

Rejected alternatives are: a general JSON Schema/JCS dependency without qualification; general-purpose text templating; host-specific prompt forks; candidate-owned evaluator controls; putting protected cases in the repository; environment-selected phases; a universal future-path allowance; random public sampling; aggregate safety scoring; and parallel writers on shared controls.

## 26. Documentation, completion, and stopping point

At final candidate freeze, `README.md` and `wave-02-evidence.md` must say exactly what exists: provider-neutral semantic/reference compiler interfaces and public local deterministic evaluation controls for development. They must also state that there is no user-facing command, host package/support result, model run, protected corpus result, controlled mutation, security qualification, stable release, deployment, or exposure.

Wave 2 completes only when every `W02-AC-*` criterion passes on one exact candidate/evidence lineage and a fresh structurally separate audit reports no unresolved Blocker, Critical, High, or Medium finding. Artifact presence or passing tests alone is insufficient.

No design approval or implementation authority is embedded here. The accountable owner may approve this exact design, request revision, or reject it. If approved, the only next action is a fresh exact implementation authorization against a reproduced source/worktree and this design's final SHA-256. Wave 3, integration to `main`, release, and deployment never begin automatically.
