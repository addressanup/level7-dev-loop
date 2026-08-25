# Level 7 Dev Loop — Wave 1 Design

| Field | Value |
|---|---|
| Artifact ID | `L7-W01-DES-001` |
| Artifact type | Proposed Wave 1 technical and delivery design |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Wave | 1 — Scope, traceability, and build-control transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — AWAITING ACCOUNTABLE-OWNER APPROVAL** |
| Change contract | [`L7-W01-CC-001`](wave-01-change-contract.md) 0.1.0, SHA-256 `f53d06d2b02760bcf6ca958b72e4d2473cc52edc3f4a2cb1471cadbd4ab42afc` |
| Specification | [`L7-W01-SPEC-001`](wave-01-specification.md) 0.1.0, SHA-256 `8715388fbe0185a3ae24d4c13d30704305a2393526fefcc71a82fce9bba119cc` |
| Proposal approval | Anup Pandey approved the exact presented pair in the current conversation on 2026-08-25; the persisted reference is `AP0` and grants no implementation authority |
| Canonical root | `/Users/anuppandey/Desktop/level7-dev-loop` |
| Source identity | Local `main` commit `ee181b759c346055b0fb5b2fa1b3b1e676dd83e4`; tree `2f23a0810660995b6f562c361ab38cd4faafa3b3`; the approved contract/specification are the only pre-design untracked files |
| Primary change class | Infrastructure or configuration change |
| Secondary change classes | Architecture/modernization change; security/authorization-governance change |
| Risk / maximum later effect | `R3 — high` / A2 local repository change only under a later exact implementation approval |
| Current effect | A1 addition of this design proposal only; no existing-file, test, dependency, Git, or external effect |
| Feature/exposure state | `NOT_APPLICABLE`; no user-visible behavior is created |
| Next gate | Accountable-owner approval or revision of this exact design; implementation remains separately unauthorized |

## 1. Decision requested

The accountable owner is asked to approve or revise this exact design. Approval confirms the proposed architecture, exact path envelope, interfaces, serial slices, verification effects, audit seam, and recovery model. It does not itself authorize implementation, an existing-file edit, branch creation, staging, a commit, a merge, a test run, a dependency, a network/provider/host action, publication, release, deployment, exposure, or continuation to Wave 2.

Before implementation can be considered, the owner must additionally:

1. resolve the root-module identity gate in §11;
2. approve the exact local branch/commit/test-effect envelope in §§13–15 against a freshly reproduced source identity; and
3. confirm the implementation candidate has no overlapping user change.

## 2. Design drivers and constraints

The design optimizes for the following approved priorities, in order:

1. **Fail-closed scope transition.** The current Step 5 product-path denial remains historical evidence while a successor becomes the active build-control entry point.
2. **Source-derived truth.** Requirement counts, ownership, allocations, prototype inventory, and candidate path deltas are derived from authoritative inputs instead of copied summaries.
3. **No product plane.** Wave 1 adds build-control logic and governance data only. It creates no `cmd/l7`, product `internal/*`, `semantic/`, schema, adapter, package, provider, or mutation path.
4. **No new dependency.** All new executable validation uses Go 1.26.7 standard library only. Go 1.27.0 remains shadow evidence.
5. **One writer and exact paths.** One short-lived local branch and one integration owner own the candidate. The reviewer is read-only and cannot remediate its own findings.
6. **Historical preservation.** Existing approved artifacts, manifests, prototype skills, host manifests, and `scripts/harness/check-foundation-scope.sh` remain byte-preserved.
7. **Truthful evidence.** Local checks, hosted checks, module ownership, grant semantics, and support claims retain separate states; no result is inherited across candidates or hosts.

## 3. Proposed component shape

```text
approved requirements/backlog ──> trace parser/check ───────┐
protected prototype assets ─────> claim/disposition check ──┤
base snapshot + path policy ─────> phase/scope check ────────┼─> deterministic decision-first report
shared-control registry ─────────> ownership check ──────────┤
module/import locks ─────────────> existing import check ────┘

Makefile / configured CI invoke the same local checks
                         │
                         └─ no product behavior, authority, network, or release verdict
```

The new `internal/harness/buildcontrol` command is a harness-only validator. It reads fixed repository-owned inputs, returns deterministic results, and has no authority-bearing interface. It does not import product code, spawn processes, use the network, load credentials, consult environment-driven policy, mutate the repository, or auto-repair an input.

## 4. Exact path design

### 4.1 Existing files proposed for modification

| Path | Intended delta | Owner |
|---|---|---|
| `Makefile` | Make the new build-control command the active `policy-check`; add explicit targeted and candidate-check targets; preserve offline/cache/toolchain constraints | Harness/build integrator |
| `.github/workflows/harness.yml` | Rename inert-step wording and invoke the same `make ci`; retain digest-pinned checkout, read-only permissions, no secret, baseline blocking, shadow nonblocking | Harness/build integrator |
| `README.md` | Report Wave 1 build-control status, commands, limitations, provisional module state, and no-product/no-support boundary | Wave integration owner |
| `scripts/harness/check-import-boundaries.sh` | Treat `internal/harness` and every descendant as test/build-control-only so future product packages cannot import the new command | Harness/build integrator |
| `go.mod` | **Conditional only:** update the root module path if the owner approves a replacement under §11; otherwise byte-preserved | Harness/build integrator under owner module decision |
| `harness/modules.lock.tsv` | **Conditional only:** mirror the approved root-module identity; keep updater `reserved`/`UNSET` | Harness/build integrator under owner module decision |

No other existing path may change under this design. In particular, `scripts/harness/check-foundation-scope.sh`, `harness/foundation-inputs.sha256`, `harness/import-boundaries.tsv`, all approved `docs/artifacts/` inputs, all prototype skills/manifests, `.go-version`, dependency/action/toolchain locks, and the inert proving test remain byte-identical.

If the owner selects a module replacement whose impact requires another existing path, this design is stale and must be revised before implementation.

### 4.2 New build-control files

| Path | Purpose |
|---|---|
| `harness/phases.tsv` | Exactly one active phase, exact base commit/tree/manifest, and active path-policy identity |
| `harness/wave-01-base.sha256` | SHA-256 snapshot of every tracked file at the approved implementation base |
| `harness/wave-01-paths.tsv` | Exact add/modify/conditional/audit-only path allowlist, owner, and rule ID |
| `harness/support-matrix.tsv` | Machine-validatable current/v1 support, effect, proof-profile, and claim states |
| `harness/prototype-dispositions.tsv` | Exactly one conform/replace/deprecate/exclude disposition for each user-invocable prototype skill |
| `harness/control-ownership.tsv` | Unique writer/integrator, reviewer rule, and earliest transition for shared control classes |
| `internal/harness/buildcontrol/main.go` | No-argument command entry point and exit handling |
| `internal/harness/buildcontrol/load.go` | Root resolution, strict UTF-8/LF/final-newline reads, TSV loading, and fixed-input inventory |
| `internal/harness/buildcontrol/markdown.go` | Bounded Markdown table/heading extraction for approved requirement/backlog formats |
| `internal/harness/buildcontrol/trace.go` | Normative ID grammar/range expansion and owner/allocation reduction |
| `internal/harness/buildcontrol/claims.go` | Support-matrix and prototype-disposition validation |
| `internal/harness/buildcontrol/policy.go` | Base snapshot, active phase, exact path delta, file-type, protected-input, and current harness-control validation |
| `internal/harness/buildcontrol/ownership.go` | Shared-control ownership and path-owner consistency validation |
| `internal/harness/buildcontrol/report.go` | Stable rule IDs, deterministic sorting, decision-first output, and bounded diagnostics |
| `internal/harness/buildcontrol/nlink_unix.go` | Darwin/Linux regular-file single-link verification for the already supported harness hosts |
| `internal/harness/buildcontrol/testutil_test.go` | In-memory fixture builder and deterministic assertion helpers |
| `internal/harness/buildcontrol/trace_test.go` | Positive, missing, duplicate, malformed, unknown, overlapping-owner, and allocation-drift cases |
| `internal/harness/buildcontrol/claims_test.go` | Support, effect, profile, prototype inventory/disposition, and false-claim cases |
| `internal/harness/buildcontrol/policy_test.go` | Phase, path, base, manifest, file-type, stale-input, reserved-module, and bypass cases |
| `internal/harness/buildcontrol/ownership_test.go` | Missing, duplicate, conflicting, protected, and unauthorized owner cases |

The test fixtures are table-driven immutable repository tests using `testing/fstest.MapFS`; no temporary fixture files, external process, network, clock, randomness, or ambient home state are needed.

### 4.3 New Wave 1 governance/evidence files

| Path | Writer / lifecycle |
|---|---|
| `docs/artifacts/wave-01-change-contract.md` | Already approved proposal; byte-preserved |
| `docs/artifacts/wave-01-specification.md` | Already approved proposal; byte-preserved |
| `docs/artifacts/wave-01-design.md` | This proposal; byte-preserved after approval |
| `docs/artifacts/wave-01-approval.md` | Implementation-pass scribe records exact contract/spec/design hashes and fresh owner authority as AP0-at-rest; cannot grant authority |
| `docs/artifacts/wave-01-module-identity-decision.md` | Exact owner choice, evidence, impact, and limits for the root module; required before implementation |
| `docs/artifacts/wave-01-grant-ladder-amendment.md` | Inert technology/backlog amendment proposal; separate security audit and approval required before activation |
| `docs/artifacts/wave-01-evidence.md` | Commands, effects, environment, results, failures, limits, candidate/manifest identities, and checkpoint state |
| `docs/artifacts/wave-01-candidate.sha256` | Exact changed-file candidate manifest; excludes itself and the later audit record; its own digest is recorded in evidence |
| `docs/artifacts/wave-01-audit.md` | Separate read-only reviewer record added only in a separately authorized audit pass after candidate freeze |

The implementation writer may not create or edit `wave-01-audit.md`. A reviewer may not modify any candidate file. Audit-record Git binding, candidate integration, and merge each remain separate effects.

## 5. Build-control command contract

### 5.1 Invocation

The production-like gate is deliberately narrow:

```text
<pinned-go> run -mod=readonly ./internal/harness/buildcontrol
```

It accepts no command-line options, configuration path, environment-selected phase, repair mode, network location, output path, or suppression flag. It requires the current directory to resolve to the canonical repository root by checking fixed root markers and the active phase record. Tests call pure functions directly with an in-memory filesystem.

`Makefile` invokes the command with the pinned `$(GO)` binary after `toolchain-check`. The command may read only the fixed paths named in this design plus the complete repository file inventory needed for scope validation. It writes nothing; Go's compile cache remains the already declared repository `.cache` effect owned by `Makefile`.

### 5.2 Output and exit behavior

Successful output is one stable decision-first line followed by sorted component summaries:

```text
PASS rule=BCTL-000 phase=wave-01 requirements=163 allocation=v1.0:140,v1.x:18,later:5 prototypes=12
```

A policy/validation failure is nonzero and begins:

```text
BLOCKED rule=<stable-rule-id> subject=<bounded-identifier> message=<bounded-text> next=<one-action>
```

Internal parse or I/O failure is nonzero, uses a stable rule, and never falls back to a broader phase. Diagnostics are ASCII, single-line per finding, lexically sorted, free of terminal controls and raw environment values, and capped by count and byte limits. When multiple findings exist, every finding is reported up to the cap; aggregate success cannot hide a safety failure.

### 5.3 Stable rule families

| Prefix | Meaning |
|---|---|
| `BCTL-0xx` | Root, input format, determinism, and internal command contract |
| `TRACE-1xx` | Requirement grammar, uniqueness, ownership, allocation, and totals |
| `CLAIM-2xx` | Support, effect ceiling, proof profile, priority, prototype disposition, and release wording |
| `SCOPE-3xx` | Base snapshot, active phase, path delta, protected bytes, file type, module, and predecessor state |
| `OWN-4xx` | Shared-control writer/reviewer/transition ownership |
| `BND-xxx` | Existing module/import boundaries retained from `harness/import-boundaries.tsv` |

Rule identifiers and their meaning are public build-control contracts. Changing or reusing one requires an impact note and owner-approved control change.

## 6. Strict data formats

Every new TSV uses UTF-8, LF, one final newline, tab separators, a single exact header, no quoting, no blank data row, no leading/trailing field whitespace, and `#` comments only before the header. Unknown columns, enums, duplicate keys, extra fields, control characters, non-ASCII paths, `.`/`..`, absolute paths, and path aliases fail closed.

### 6.1 `harness/phases.tsv`

```text
phase	state	base_commit	base_tree	base_manifest	path_policy
wave-01	active	<40-hex>	<40-hex>	harness/wave-01-base.sha256	harness/wave-01-paths.tsv
```

Exactly one row and one `active` state are permitted in Wave 1. The base commit/tree must match the fresh owner-approved implementation source identity. A later wave adds a newly reviewed successor row and changes active state through its own approved transition; it cannot edit history to pretend the earlier phase never existed.

### 6.2 `harness/wave-01-paths.tsv`

```text
change	path	owner	rule
add|modify|conditional|audit-only	<exact-relative-path>	<owner-id>	<SCOPE-rule>
```

All paths in §4 appear exactly once. Deletion is not an allowed Wave 1 operation. `conditional` is valid only for `go.mod` and `harness/modules.lock.tsv` and is resolved by the module decision before implementation. `audit-only` is valid only for `docs/artifacts/wave-01-audit.md` and is unavailable to the implementation writer.

### 6.3 `harness/support-matrix.tsv`

```text
id	surface	current_state	v1_ceiling	claim_state	owner
```

Required IDs cover Codex advisory package, Claude advisory package, Controlled Client, generic profile, feature/behavior-change profile, behavior-preserving-refactor profile, A3/A4 handoff, A5 absence, dual-host support, and stable version claim. Enums distinguish `prototype`, `planned`, `development`, `unsupported`, and `absent` from `supported`; Wave 1 permits no `supported` or stable-release row.

### 6.4 `harness/prototype-dispositions.tsv`

```text
skill	disposition	target_owner	cutover
```

The validator derives the current user-invocable inventory from `skills/*/SKILL.md` frontmatter and requires exactly these approved dispositions:

| Skill | Disposition | Target owner / cutover intent |
|---|---|---|
| `l7-next` | `conform` | `BL-007`; sole generated conductor at Wave 7 |
| `l7-constitution` | `replace` | `BL-002/BL-007`; internal invariant/frame obligations |
| `l7-build` | `replace` | `BL-007/BL-008/BL-009`; internal generic/feature/refactor path |
| `l7-change` | `replace` | `BL-011` plus applicable future profiles; local candidate and handoff only |
| `l7-review` | `replace` | `BL-010`; evidence-gap and candidate-assurance flow |
| `l7-release` | `replace` | `BL-010/BL-042`; structurally separate assurance/release flow |
| `l7-deploy` | `exclude` | `BL-011`; data-only A3/A4 handoff, no executable v1 surface |
| `l7-greenfield` | `exclude` | Future `BL-016` after v1.0 |
| `l7-ops` | `exclude` | Future `BL-022` profiles |
| `l7-experience` | `exclude` | Future `BL-019` plus P0 status-experience obligations |
| `l7-geometry` | `deprecate` | Remove universal/perfection contract; future product-specific UX criteria only |
| `l7-storybook` | `exclude` | Future conditional `BL-029` profile |

Inventory mismatch, missing/duplicate skill, unknown disposition, or public-cutover ambiguity is `BLOCKED`. Wave 1 records this plan but does not edit any prototype asset.

### 6.5 `harness/control-ownership.tsv`

```text
control	path_kind	path	writer	reviewer	change_gate
```

`path_kind` is `exact` or non-overlapping `prefix`. Every shared control class in `L7-ORC-001` §10 appears once. Protected evaluator, grant, signing, AP2/AP3, promotion, and release controls use writer `external-denied` for the candidate repository. Overlapping path rules, multiple writers, missing reviewer rules, or an implementation-writer claim over an external/protected control fail.

## 7. Trace-validation design

The parser does not attempt general Markdown interpretation. It recognizes only the approved bootstrap table shapes:

1. In `requirements.md`, it enters the normative region at `## 9. Functional requirements`, continues through `## 10. Nonfunctional requirements`, and stops at `## 11.`. A definition row must have a first cell containing exactly one backticked ID matching `L7-[A-Z]+-[0-9]{3}`.
2. In `feature-backlog.md`, it enters `## 8. Normative requirement ownership and release allocation` and stops at `## 9.`. It parses only four-cell ownership rows and expands comma-separated exact IDs and zero-padded inclusive ranges within one prefix.
3. It rejects duplicate definitions, repeated/overlapping ownership expressions, reversed ranges, unknown prefixes/IDs, missing IDs, multiple owners, unknown allocation values, and totals that differ from the derived set.
4. The displayed summary totals are checked against the derived result but never used to generate it.

The successful current result must be exactly 163 unique definitions and ownership entries, allocated `140 V1.0 / 18 V1.x / 5 Later`. A future legitimate requirement change must update the authoritative sources through their own approval and then update the expected gate in a new source-bound candidate; a developer cannot edit only the generated total.

## 8. Phase and scope-validation design

### 8.1 Immutable base

Immediately before a later implementation authorization, the integrator reproduces the exact clean source identity. `harness/wave-01-base.sha256` is generated deterministically from every regular tracked file at that base, sorted by bytewise path. It records content SHA-256 and relative path, excludes `.git` and `.cache`, and includes every approved historical artifact and prototype asset.

The base manifest is not permission. `harness/phases.tsv` binds its digest, commit, and tree; the implementation approval binds the same tuple. A mismatch is `SCOPE-301` and blocks before any candidate check can pass.

### 8.2 Candidate delta

The phase checker walks the repository without following links and excludes only top-level `.git` and `.cache`. It then proves:

- every base path still exists as a regular single-link file;
- every base path not marked `modify` or resolved `conditional` remains byte-identical to the base manifest;
- every changed or added path is listed exactly once in `wave-01-paths.tsv` with the current writer role;
- no deletion, symlink, special node, non-ASCII alias, unowned path, or extra non-cache file exists;
- the approved contract, specification, and design match their bound hashes;
- `foundation-inputs.sha256` still verifies every protected foundation/prototype input;
- root module and updater states match the separately approved module decision;
- action/toolchain/signing/import locks and configured CI safety properties retain their Step 5 invariants; and
- product paths remain absent in Wave 1.

`scripts/harness/check-foundation-scope.sh` remains unchanged as historical Step 5 source. It is no longer the active `policy-check`; the successor reproduces all still-applicable controls and proves the predecessor bytes through the base and protected manifests. Historical Step 5 evidence remains attached to its original candidate and is never relabeled as a Wave 1 run.

### 8.3 Future transition seam

The policy contains no hard-coded allowance for future product directories. Wave 2 or later must add its own source-bound phase/path policy and negative fixtures through a new approved transition. Unknown phases remain denied. This prevents Wave 1 from creating a dormant universal bypass.

## 9. Claim and ownership validation

The claim checker compares the strict TSV records with approved constants derived from `L7-BL-001`, `L7-ORC-001`, and the current prototype inventory. It rejects:

- a plugin-install-to-mutation implication;
- A3/A4 execution or any A5 interface;
- Codex evidence inherited by Claude or the reverse;
- stable `1.0.0`, dual-host, enforcement, containment, compatibility, release, or security claims;
- a generic proof pass where a material specialist profile is absent;
- a missing/duplicate prototype disposition; or
- P0/P1/P2 drift without an approved impact decision.

The ownership checker cross-validates `control-ownership.tsv` against the path policy and the orchestration ownership table. It ensures the implementation writer owns only Wave 1 build-control and governance paths, generated/candidate outputs have one integrator, and protected evaluator/grant/signing/release classes remain outside candidate authority.

## 10. Existing import-boundary integration

`scripts/harness/check-import-boundaries.sh` remains the Go package-graph gate. Its only proposed semantic change is to reject imports of the exact `internal/harness` package **or any descendant**, including `internal/harness/buildcontrol`, from every non-harness package. The command itself remains standard-library-only and is never linked into product code.

The current rules for executor/transaction/receipt separation, pure kernel/policy/semantic closures, updater reservation, external-module detours, and `unsafe` remain unchanged. Since Wave 1 introduces no product package, the positive package graph is still harness-only; negative tests prove future bypass attempts fail.

## 11. Module-identity decision gate

The design cannot infer domain or vanity-module control. Before implementation, Anup Pandey must make one exact owner decision recorded in `wave-01-module-identity-decision.md`:

1. **Confirm current identity:** attest control and intended use of `continuallabs.ltd/level7-dev-loop`. `go.mod` and `harness/modules.lock.tsv` remain unchanged; the decision records evidence and publication limitations.
2. **Approve replacement:** name one exact replacement module and its ownership evidence. `go.mod` and `harness/modules.lock.tsv` change together; `Makefile` derives the harness import path from the module registry rather than retaining a conflicting literal.

If neither decision is made, implementation is `BLOCKED`. If a replacement affects any path not listed in §4.1 or introduces a dependency, redirect, network lookup, publication, or compatibility commitment, this design must be revised and reapproved. The updater stays `reserved`/`UNSET` in either case.

## 12. Grant-ladder amendment boundary

`wave-01-grant-ladder-amendment.md` is a data-only successor proposal to the conflicting portion of `TDR-013`. It contains:

- the exact clauses and backlog/orchestration assumptions affected;
- distinct schemas and trust policies for `qualification`, `evaluation`, `pilot`, and `stable`;
- audience, purpose, issuer, candidate, host/model/platform, target root/class, effect, expiry, nonce, revocation, and policy bindings;
- non-interchangeability and arbitrary-user-root denial invariants;
- lower-ceiling behavior when a grant, environment, or proof is absent;
- migration, compatibility, removal, compromise, and rollback consequences; and
- a compact R3 assurance case plus required independent security/boundary audit inputs.

The artifact status remains `PROPOSED — INERT`. No code parses it, no test flag activates it, and Wave 1 cannot issue/sign/install/use a grant. A separate digest-bound technology/backlog approval must explicitly supersede `TDR-013` before the amendment becomes normative. Failure leaves C2 synthetic mutation, pilot mutation, and stable mutation at their documented blocked/lower ceilings.

## 13. Serial implementation slices

All slices are proposed only; none is authorized by design approval. The later implementation action uses one writer and keeps every completed commit green.

| Slice | Paths / result | Verification before next slice | Proposed conventional commit |
|---|---|---|---|
| 0 — Bind decisions | Add approved planning/approval/module-decision records; resolve conditional module paths; generate exact base/path policies | Hash/source/path checks; no product path; no dependency | `docs(wave-01): bind approved build-control plan` |
| 1 — Trace and claim contracts | Add strict data files, Markdown/TSV loaders, trace/claim logic, and complete table-driven fixtures | Targeted Go tests; current 163/140/18/5 and 12-skill results | `test(build-control): enforce scope and claim contracts` |
| 2 — Phase and ownership gate | Add phase/base/path/ownership policies, scope/ownership logic, file-type controls, and adversarial tests | Targeted tests plus current repository gate; old protected bytes match | `feat(build-control): add fail-closed phase gate` |
| 3 — Harness integration | Update import checker, `Makefile`, configured CI wording, and README | Baseline full local verify; shadow full local verify; configured workflow static checks | `chore(harness): activate wave 01 controls` |
| 4 — Amendment and freeze | Add inert grant amendment; freeze exact candidate manifest and evidence record | Manifest reproduction, diff/path/secret/dependency audit, `git diff --check` | `docs(wave-01): freeze candidate evidence` |
| 5 — Independent audit | Separate read-only reviewer inspects the exact Slice 4 commit and creates only the audit record under separate authority | Zero unresolved Blocker/Critical/High/Medium; candidate bytes unchanged | Separate reviewer-owned `docs(audit): review wave 01 scope transition` |

No API or UI slice is invented because Wave 1 has neither. Schema here means strict build-control TSV/Markdown input contracts, not a product artifact schema. The grant amendment is governance data, not an authorization implementation.

## 14. Verification plan and effects

### 14.1 Targeted commands

The later approved test envelope is expected to authorize exactly these local commands from the canonical root:

```text
make policy-check GO_VERSION=1.26.7
make import-check GO_VERSION=1.26.7
make test GO_VERSION=1.26.7
make verify GO_VERSION=1.26.7
make verify GO_VERSION=1.27.0
git diff --check
git status --porcelain=v2 --branch --untracked-files=all
```

The exact final implementation approval may narrow but may not silently broaden this list. `make bootstrap`, dependency installation with network access, hosted CI dispatch, provider/host invocation, remote Git, root operation, and ambient cleanup are excluded. If either pinned local toolchain is missing, the applicable check is `BLOCKED`; no download is inferred.

### 14.2 Declared local effects

`make`/Go commands may write only to the already ignored repository-scoped paths under `.cache/go/`, `.cache/repro/`, and `.cache/toolchains/` as documented by the current harness. They may update repository-scoped Go telemetry mode state fixed to `off`. They do not authorize writes to ambient Go telemetry, home/config directories, credentials, package managers, remotes, or external systems.

Tests use no network, credential, real provider, real host plugin, clock-sensitive oracle, process sandbox, user repository fixture, or hidden evaluator. Hosted Linux remains `NOT_RUN` unless separately authorized later. Local macOS results remain development evidence and cannot promote the controlled Ubuntu or dual-host matrix.

### 14.3 Required negative cases

The table-driven suite includes at least:

- missing/duplicate/malformed/unknown/reversed/overlapping requirement IDs;
- zero/two owners and all allocation-total drifts;
- summary-total tampering without source-definition change;
- missing/duplicate/unknown prototype skill or disposition;
- plugin-implies-mutation, cross-host inheritance, A3/A4/A5 execution, stable/dual/enforcement, and generic-specialist false claims;
- missing/duplicate active phase, stale base, malformed manifest, unauthorized add/modify/delete, symlink/special/hardlinked file, non-ASCII path, and protected-byte drift;
- reserved updater, wrong root module, external-module detour, harness import, `unsafe`, and forbidden transitive import;
- missing/duplicate/overlapping owner and candidate ownership of protected controls; and
- diagnostic-order, cap, exit-code, no-repair, and repeat-run determinism.

Each broken candidate must fail for its intended stable rule. Blanket failure is not proof that the correct control detected the fault.

## 15. Git, review, and integration design

### 15.1 Proposed local branch strategy

After a separate exact authorization, use one short-lived branch `feat/wave-01-build-control` in the canonical worktree from a freshly reproduced clean `main` base. Do not create a remote or secondary writer worktree. The approved planning files become part of the first bounded commit; untracked or modified paths outside `wave-01-paths.tsv` block branch work.

Branch creation, each commit, audit-record commit, and final integration are explicit local Git effects. This design does not authorize them. Commit author metadata is not approval or identity evidence.

### 15.2 Candidate freeze

After Slice 4:

1. full baseline and shadow local verification complete;
2. `wave-01-candidate.sha256` lists every changed regular file except itself and the absent later audit, sorted by path;
3. `wave-01-evidence.md` records the manifest digest, commit/tree/parent, commands/effects/results, environment, and all limits/`NOT_RUN` states;
4. the branch is frozen for reviewer read-only access; and
5. any candidate edit invalidates verification and review.

### 15.3 Independent audit and integration

The reviewer receives the exact Slice 4 commit, read-only candidate authority, the planning/decision chain, base/path manifests, negative-fixture inventory, verification evidence, and the R3 assurance case. The reviewer may add only `wave-01-audit.md` in a later separately authorized pass and cannot remediate findings.

Any Blocker, Critical, High, or Medium finding returns to a new remediator candidate and requires fresh full verification and re-audit. A favorable model audit is development evidence only; it is not AP2/AP3, qualified security approval, product support, or release `GO`.

Merging/integrating the audited branch into `main` is a separate owner decision. After any approved integration, rerun the complete baseline and shadow local harness and confirm the candidate files remain byte-identical apart from the separately bound audit record. No Wave 2 work starts automatically.

## 16. Recovery and interruption design

| Failure point | Safe state | Recovery |
|---|---|---|
| Before branch creation | `main` plus approved untracked planning files; no implementation | Stop; revise design/authority if needed |
| Slice fails before commit | Branch dirty; overlapping writers frozen | Restore exact touched preimages using a reviewed inverse patch or discard only under explicit owner authority; never use destructive reset implicitly |
| Committed slice fails later verification | Branch remains isolated; `main` unchanged | Add a narrow conventional remediation commit; rerun from the first affected gate |
| Phase gate partially wired | Old `main` remains authoritative; candidate is `BLOCKED` | Restore active `policy-check` through a narrow remediation commit before further slice work |
| Unexpected path/dependency/network/credential effect | Candidate fails; authority stops | Preserve evidence, assess exposure, revoke affected authority if applicable, and require a new scoped plan |
| Audit finding | Candidate frozen and not integrated | Remediator creates a new digest/commit; reviewer re-audits exact successor |
| Interruption/resume | No approval or result is inherited from mutable files | Reproduce root, base, branch, status, path manifest, proposal/design hashes, authority, and last green commit before continuing |

No cleanup of `.cache`, ambient telemetry, branches, commits, or unexpected files is automatic. Material deletion or branch discard requires explicit owner direction.

## 17. Security, privacy, and performance self-review design

### 17.1 Security boundaries

- Build-control data and Markdown are untrusted bytes; parsers are bounded and reject ambiguity.
- The validator has no `os/exec`, `net`, credential, environment-policy, repair, approval, signing, or mutation interface.
- Root resolution is canonical; repository walk does not follow links; allowed paths are exact ASCII relative names.
- Every admitted file is regular and single-link on the supported harness OSes.
- Protected inputs are checked by content digest, not presence.
- Candidate-author ownership is denied for evaluator, grant, signing, promotion, AP2/AP3, and release controls.
- The grant amendment cannot be loaded by runtime code or test mode.

### 17.2 Privacy

No personal data or secret is required. Reports include repository-relative paths, public artifact IDs, digests, tool/environment versions, and bounded diagnostics only. They do not persist raw environment values, credentials, user prompts, session transcripts, or hidden chain-of-thought.

### 17.3 Boundedness

The validator scans only fixed authoritative files plus one complete non-`.git`/non-`.cache` repository inventory. The implementation constants cap file count, total bytes, Markdown line length, TSV rows, expanded IDs, findings, and output bytes. Exceeding any cap is `BLOCKED`, never truncation with success. No retry, daemon, watcher, goroutine fan-out, or background process is needed.

Performance is a developer-harness budget, not a release benchmark. The evidence record reports observed duration and repository size; no hard speed claim is made in Wave 1.

## 18. Acceptance mapping

| Design element | Specification acceptance |
|---|---|
| Strict Markdown trace parser and `trace.go` | `W01-AC-001` |
| Support/prototype TSVs and `claims.go` | `W01-AC-002`–`005` |
| Phase/base/path manifests and `policy.go` | `W01-AC-006`–`007`, `W01-AC-013` |
| Module-decision gate | `W01-AC-008` |
| Inert grant amendment | `W01-AC-009` |
| Ownership TSV/checker | `W01-AC-010` |
| Makefile/CI integration and exact verification plan | `W01-AC-011` |
| Candidate manifest, evidence, and separate audit | `W01-AC-012` |

No aggregate result can waive a failed acceptance criterion, kill condition, or unresolved module/grant/review gate.

## 19. Residual risks and rejected alternatives

| Item | Disposition |
|---|---|
| Markdown bootstrap parsing is format-coupled | Accepted for Wave 1 with exact heading/table contracts and adversarial tests; canonical artifact schemas arrive in later waves |
| Local Git/filesystem and same-user artifacts are mutable | Explicit residual risk; source hashes, isolated branch, manifests, and read-only review improve evidence but do not create external trust |
| Hosted Linux and controlled Ubuntu evidence absent | Remains `NOT_RUN`; cannot affect Wave 1 build-control-only checkpoint |
| Module ownership unconfirmed | Hard pre-implementation block under §11 |
| Grant ladder unapproved | Amendment remains inert; mutation ceilings remain blocked/lower |
| Reusing `check-foundation-scope.sh` as the permanent active gate | Rejected because its unconditional product-path denial cannot express later approved phases |
| Editing the old Step 5 checker/manifests in place | Rejected because it destroys historical candidate identity |
| Environment-selected phase or repository feature flag | Rejected because it creates a bypassable admission path |
| General YAML/JSON schema dependency | Rejected in Wave 1 to preserve the zero-production-dependency harness; strict TSV/approved Markdown contracts suffice for this bootstrap transition |
| Shell-only range/Markdown parser | Rejected due portability and ambiguity risk; standard-library Go provides bounded typed parsing and table-driven tests |
| Git-only scope detection | Rejected as the sole gate; complete base/path manifests also detect unexpected non-Git files and protected-byte drift |
| Parallel writers | Rejected for this shared, high-risk harness transition |

## 20. Approval and stopping point

No design approval or implementation authority is embedded here. The accountable owner may approve this exact design, request revision, or reject it.

If approved, the next action is **not automatic implementation**. The owner must first resolve the module identity and then issue a fresh exact authorization for the proposed branch, files, slices, local Git effects, and test/cache effects against the then-current source identity. Until that happens, every implementation path remains blocked.
