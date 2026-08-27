# Standalone CLI v1 — Product Foundation Brief

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1` |
| Risk tier | `3` — initial wave changes protected CI and import-policy controls |
| Status | `planned`; awaiting Tier 3 implementation approval |
| Base commit | `2c0e1c97c0344e423a75b01fa3d1a0dc423a2b9d` |
| Primary user | Solo product founder |
| Primary interface | Standalone macOS CLI |
| Requirements decision | Accepted by the user in the active interaction on `2026-08-27`; this is not implementation authority |
| Backlog decision | Accepted by the user in the active interaction on `2026-08-27`; this is not implementation authority |
| Architecture decision | Option A accepted by the user in the active interaction on `2026-08-27`; this is not implementation authority |
| Technology decision | Go 1.26.7 accepted by the user in the active interaction on `2026-08-27`; this is not implementation authority |
| Harness decision | Accepted by the user in the active interaction on `2026-08-27`; this is not implementation authority |
| Orchestration decision | Accepted by the user in the active interaction on `2026-08-27`; planning freeze authorized, implementation not authorized |

## Problem

A solo product founder using coding agents must carry a consequential feature
from an ambiguous idea to a merge decision, often across interruptions. Codex and
Claude Code can implement quickly, but neither chat history nor generated prose
alone reliably preserves current intent, scope, authority, verification, or the
reason a candidate is safe to accept.

The product must cover the whole journey—framing, implementation, interruption
recovery, verification, review, and merge decision—without recreating the
artifact-heavy governance process that this repository has deprecated.

## Users and operating context

The primary v1 user is a solo product founder who owns product intent, repository
quality, and release consequences while delegating implementation work to Codex
and Claude Code. The founder may allow broad local automation but remains the
accountable authority for merge and high-risk boundaries.

The initial supported environment is macOS. The CLI must work across technology
stacks and repository sizes, including large monorepos, by using Git and
repository-defined commands instead of assuming a language or framework.

## Product outcome

The founder can begin with incomplete intent and reach a high-quality,
performance-conscious merge candidate while the CLI:

- preserves the active problem, scope, decisions, and evidence in the repository;
- launches and coordinates Codex and Claude Code;
- resumes the correct workflow state after interruption;
- verifies the exact Git candidate with repository-defined checks;
- uses cross-provider independent review for high-risk work when available; and
- asks the founder before every merge or other explicitly protected action.

Speed of resuming a session is not a success measure. Correct state recovery,
software quality, and the performance of both the CLI and produced software are.

## Functional requirements

| ID | Requirement | Acceptance boundary |
|---|---|---|
| `FR-001` | Initialize or adopt a Git repository without replacing its existing engineering controls. | Existing commands, CI, instructions, and history remain intact; generated state is explicit and reviewable. |
| `FR-002` | Capture a concise problem, scope, acceptance criteria, risk tier, and rollback before implementation when the change requires a brief. | Tier 1 needs no artifact; Tier 2/3 uses one change brief as defined by the lean workflow. |
| `FR-003` | Derive workflow and candidate identity from Git. | Status names the exact base, head, tree, changed scope, current state, and one executable next action without a candidate SHA manifest. |
| `FR-004` | Launch and orchestrate installed Codex and Claude Code clients. | The founder can select either implementer; lifecycle state remains provider-neutral and survives switching clients. |
| `FR-005` | Persist durable workflow state locally in the repository. | A fresh CLI process reconstructs the accepted state without chat history or a hosted account. Sensitive transient data and credentials are not committed. |
| `FR-006` | Perform scoped local work autonomously. | Within granted scope, the CLI may read, edit, run tests, and create conventional Git commits; scope expansion fails closed. |
| `FR-007` | Use repository-defined verification. | The CLI discovers or accepts explicit lint, type, test, build, benchmark, and CI commands and records their current result without language-specific assumptions. |
| `FR-008` | Protect quality and performance. | A merge recommendation identifies current failures, test coverage relevant to the change, benchmark regressions, uncertainty, and residual risk. Passing tests never imply approval. |
| `FR-009` | Enforce proportionate review. | Tier 1/2 receives normal review. Tier 3 requires distinct owner approval and independent read-only review; Codex preferably audits Claude work and Claude audits Codex work, with a human reviewer permitted. |
| `FR-010` | Require founder confirmation before merge. | The CLI never merges solely because checks or reviews pass; confirmation is bound to the current Git candidate and becomes stale after candidate changes. |
| `FR-011` | Protect consequential boundaries. | Authorization/security changes, destructive actions, material migrations, governance-controller changes, and production deployment require explicit founder approval. |
| `FR-012` | Support headless CI. | Non-interactive commands return stable machine-readable output and meaningful exit codes, require explicit trusted inputs, and never infer authority from repository prose. |
| `FR-013` | Remain recoverable. | Failed verification or review returns to an executable remediation state; interruption, rejection, and rollback cannot leave an accepted deadlocked state. |
| `FR-014` | Make truthful capability claims. | The CLI distinguishes planned, available, verified, reviewed, and ready behavior and never upgrades a claim from text, agent confidence, or tests alone. |

## Nonfunctional requirements

| ID | Requirement | Measure |
|---|---|---|
| `NFR-001` | Local-first privacy | The Level 7 CLI makes no network calls. Only an explicitly invoked Codex or Claude client may use its own configured network access. |
| `NFR-002` | macOS compatibility | Supported macOS versions and CPU architectures are declared and exercised in CI before release. |
| `NFR-003` | Technology neutrality | Core behavior has no dependency on a product repository's language, framework, package manager, or CI provider. |
| `NFR-004` | Scale proportionality | Routine status and policy work is based on Git metadata and changed scope, not unconditional full-content repository scans. Large inputs are streamed or bounded and fail with actionable diagnostics. |
| `NFR-005` | CLI performance | Benchmarks cover small, medium, and large-monorepo fixtures. A greater than 10% regression from the accepted baseline blocks release unless the founder explicitly accepts the tradeoff. |
| `NFR-006` | Produced-software performance | Repository-defined performance checks are first-class verification results. A material regression blocks a ready recommendation unless explicitly accepted. |
| `NFR-007` | Determinism | Given the same repository bytes, Git identity, configuration, and trusted external inputs, policy state and machine-readable output are stable. |
| `NFR-008` | Security | Arguments, paths, environment, subprocesses, symlinks, output size, credentials, and untrusted agent output are treated as hostile inputs and tested accordingly. |
| `NFR-009` | Observability | Human-readable decisions lead with outcome and next action; structured output exposes rule IDs without storing hidden reasoning or unnecessary personal data. |
| `NFR-010` | Reversibility | Repository mutations are represented by reviewable Git commits; incomplete orchestration can be stopped without corrupting workflow state. |

## Success measures

v1 succeeds only when quality and performance evidence supports it. Time-to-resume
is deliberately excluded.

| Measure | Initial target |
|---|---|
| False `ready` decisions in adversarial lifecycle tests | `0` |
| Merge recommendations lacking evidence for the exact current Git candidate | `0` |
| Tier 3 approval bypasses or accepted self-audits | `0` |
| Valid accepted states without an executable next transition | `0` |
| Interruption fixtures that reconstruct the wrong accepted state | `0` |
| CLI benchmark regression against the accepted fixture baseline | `≤10%`; otherwise block or record explicit acceptance |
| Material produced-software performance regression | `0` unacknowledged regressions |
| Change reopens/reverts attributable to lost intent, scope, or stale evidence | No worse than repository baseline in pilot use; improvement required before claiming a quality benefit |

## Constraints and explicit boundaries

- State is repository-local; v1 has no hosted account or service.
- The CLI supports Codex and Claude Code at launch and does not claim identical
  provider capabilities where they differ.
- Local read, edit, test, and scoped commit actions may be automated.
- Every merge requires founder confirmation.
- Production deployment remains a separately approved Tier 3 action.
- The CLI does not replace Git, host sandboxes, CI, code review, package managers,
  or repository-specific test tools.
- v1 does not promise defect-free software, guaranteed productivity, regulatory
  compliance, or performance improvement.
- Windows and Linux support are outside the v1 commitment.

## Smallest viable backlog

Effort uses relative points (`1`, `2`, `3`, `5`, `8`). P0 is the smallest
complete founder journey; an item is not complete merely because its happy path
works.

### P0 — usable walking skeleton

| ID | Outcome | Depends on | Effort | Acceptance criteria |
|---|---|---|---:|---|
| `CLI-001` | Stable CLI shell and repository adoption | — | 3 | A versioned command grammar provides human and JSON output, stable exit codes, configuration discovery, cancellation, and non-destructive adoption of an existing Git repository. |
| `CLI-002` | Git-derived change and lifecycle core | `CLI-001` | 5 | Tier, scope, base/head/tree, changed paths, state, and executable next action are derived deterministically; scope expansion, stale state, and invalid transitions fail closed. |
| `CLI-003` | Repository-local continuity | `CLI-002` | 3 | A fresh process resumes the exact accepted state after interruption; corrupt, partial, conflicting, or stale local state is diagnosed without inventing progress. |
| `CLI-004` | Technology-neutral verification runner | `CLI-001`, `CLI-002` | 5 | Explicit repository commands run with bounded environment, output, timeout, and cancellation; results bind to the current Git candidate and become stale after change. |
| `CLI-005` | Codex adapter | `CLI-001`, `CLI-002` | 5 | Detects the installed client, launches a scoped session, passes provider-neutral task context, observes bounded lifecycle results, and reports unsupported capabilities truthfully. |
| `CLI-006` | Claude Code adapter | `CLI-001`, `CLI-002` | 5 | Meets the same adapter contract without claiming Codex/Claude flag, permission, or session parity. |
| `CLI-007` | Implement-and-review orchestration | `CLI-003`–`CLI-006` | 8 | Founder selects an implementer; the CLI can run implementation and verification, then prefer the other provider for Tier 3 read-only audit; implementer, reviewer, and exact candidate identities remain distinct and bound. |
| `CLI-008` | Founder-controlled ready and merge flow | `CLI-002`, `CLI-004`, `CLI-007` | 5 | The CLI cannot report ready without current required evidence and review, never treats tests as approval, invalidates stale confirmation, and asks the founder immediately before an exact-candidate merge. |
| `CLI-009` | Headless CI policy mode | `CLI-002`, `CLI-004`, `CLI-008` | 3 | Non-interactive evaluation accepts only explicit trusted metadata, emits stable JSON and exit codes, performs no prompts or network calls, and cannot be weakened by candidate-controlled policy. |
| `CLI-010` | Quality, performance, and adversarial release gate | `CLI-001`–`CLI-009` | 5 | Small/medium/monorepo benchmarks establish the accepted baseline; lifecycle, subprocess, filesystem, malformed-input, authority, self-review, interruption, and false-ready tests meet all success measures. |

P0 is complete only when one fixture repository can execute this sequence on
macOS:

`adopt → brief → Codex-or-Claude implement → repository verification → normal or cross-provider review → founder-confirmed merge`

The inverse provider order must also pass. Tier 1 must remain artifact-free, and
the fixture must demonstrate interruption and resume between every adjacent
state.

### P1 — quality and daily-use depth

| ID | Outcome | Depends on | Effort | Acceptance criteria |
|---|---|---|---:|---|
| `CLI-101` | Explain and diagnose | P0 | 3 | Every block explains the failed rule, evidence source, remediation, and next executable command without exposing hidden reasoning. |
| `CLI-102` | Repository command profiles | `CLI-004` | 3 | Teams can declare reusable test/build/benchmark profiles with bounded inputs and deterministic precedence. |
| `CLI-103` | Monorepo-aware targeting and caching | `CLI-002`, `CLI-004`, `CLI-010` | 5 | Changed components and applicable verification are selected without unconditional full-repository content scans; cache reuse is Git-bound and cannot hide stale failures. |
| `CLI-104` | Provider capability diagnostics | `CLI-005`, `CLI-006` | 3 | `doctor` reports installed versions, supported operations, permission/network boundaries, and safe degradation before orchestration. |
| `CLI-105` | Outcome-quality reporting | `CLI-008` | 5 | Local, privacy-preserving summaries compare reopens, reverts, performance regressions, false-ready blocks, and explicitly accepted risks without scoring people. |

### P2 — explicit post-v1 candidates

| ID | Candidate | Dependency | Promotion condition |
|---|---|---|---|
| `CLI-201` | Linux support | Stable macOS P0/P1 | Platform CI, installer, subprocess, path, permission, and filesystem behavior meet the same acceptance boundaries. |
| `CLI-202` | Windows support | Stable adapter and filesystem contracts | Native process, path, shell, permission, and Git behavior are designed and independently verified rather than emulated by claim. |
| `CLI-203` | Additional coding-agent adapters | Stable provider-neutral adapter contract | A real user need and capability-specific conformance suite justify each adapter. |
| `CLI-204` | Optional visual interface | Stable CLI and local data contract | User evidence shows a visual surface improves decisions without splitting authority or state. |
| `CLI-205` | Optional remote collaboration | Proven local-first product | Separate privacy, identity, threat-model, data lifecycle, and service-operability decisions are approved; no silent hosted dependency. |

## Backlog sequencing rules

- Implement P0 in dependency order; do not parallelize changes that share the
  lifecycle core, repository state, or command contract.
- Provider adapters may be built in isolated branches after their common adapter
  contract is accepted, then integrated sequentially.
- `CLI-010` is not a final testing phase: its fixtures and benchmarks begin with
  `CLI-001` and grow with every item.
- Each implementation slice receives its own risk classification. Provider
  invocation and ordinary CLI behavior are Tier 2 unless a protected Tier 3
  boundary is changed.
- Defer P1 and P2 until the complete P0 journey works in both provider orders.
- No backlog item authorizes automatic merge, production deployment, hidden
  network access, or claims of quality/performance benefit without evidence.

## Architecture options

Scores use `1` (poor) through `5` (strong). Weighted totals are out of `100`.

| Criterion | Weight | A — short-lived modular CLI | B — CLI plus local daemon | C — thin script/plugin coordinator |
|---|---:|---:|---:|---:|
| Safety and deterministic testability | 25 | 5 | 4 | 2 |
| Installation and operational simplicity | 20 | 5 | 2 | 4 |
| Small-to-monorepo performance | 15 | 4 | 5 | 2 |
| Provider process isolation | 15 | 4 | 5 | 2 |
| Headless CI suitability | 10 | 5 | 3 | 4 |
| Repository-local operation and recovery | 10 | 5 | 4 | 4 |
| Future extensibility | 5 | 4 | 5 | 3 |
| **Weighted total** | **100** | **93** | **77** | **57** |

### Option A — short-lived modular CLI with subprocess adapters

Each invocation reconstructs canonical state from Git, the one permitted brief,
trusted external authority, and bounded local runtime data. A pure domain core is
surrounded by ports for Git, filesystem, process execution, time, terminal, and
authority. Codex and Claude are child-process adapters. There is no background
service, database, or product-owned network client.

This option best fits repository locality, CI, uninstallability, deterministic
tests, and the solo-founder operating model. Its main cost is careful persistence
and cancellation across multiple short-lived commands.

### Option B — CLI client plus persistent local daemon

A repository-aware daemon owns an event store, agent processes, scheduling, and
resume behavior while a thin CLI communicates over a local socket. This makes
long-running orchestration, live progress, concurrency, and future visual clients
easier.

It adds daemon lifecycle, socket authentication, upgrades, stale-process
recovery, background resource consumption, version skew, and a second authority
surface before v1 has evidence that those costs are needed.

### Option C — thin CLI coordinating repository scripts and provider plugins

The CLI delegates workflow behavior to repository-owned shell commands and agent
plugin skills. It is quick to extend and naturally technology-neutral.

Core safety and semantics would be distributed across mutable scripts, shells,
provider prompts, and candidate-controlled files. Cross-platform quoting,
deterministic state, bounded execution, authority separation, and truthful
capability handling would be substantially harder to verify.

## Recommended architecture

Select **Option A: a short-lived modular CLI with subprocess adapters**.

The selected structure is a ports-and-adapters design with one dependency rule:
outer adapters may depend inward; the domain and application layers never import
filesystem, Git, terminal, environment, clock, network, or provider-specific
packages.

```text
Command/JSON interface
        |
Application use cases ── status, adopt, run, verify, review, merge
        |
Pure domain core ─────── risk, scope, lifecycle, candidate, evidence
        |
Ports ───────────────── Git, state, process, authority, terminal, clock
        |
Adapters ────────────── local Git/files, command runner, Codex, Claude, CI
```

### Component responsibilities

| Component | Owns | Must not own |
|---|---|---|
| Command interface | Argument parsing, human/JSON rendering, exit codes, confirmation UX | Policy decisions, shell evaluation, provider-specific lifecycle rules |
| Application layer | Use-case ordering, cancellation, transaction boundaries, progress events | Direct filesystem/process/environment access |
| Domain core | Risk tiers, scope, states/transitions, Git candidate identity, evidence freshness, readiness | I/O, clocks, credentials, agent invocation, mutable global state |
| Git adapter | Repository discovery, refs/trees/diffs, commits, merge preconditions | Approval inference or destructive history rewriting |
| State adapter | Strict schema/version parsing, atomic local writes, recovery and migration | Duplicating data already authoritative in Git or CI |
| Verification adapter | Explicit argv execution, bounds, cancellation, result capture | Shell interpolation, implicit command discovery with side effects |
| Provider adapters | Installed-client discovery, capability mapping, bounded launch, provider-neutral handoff | Network implementation, credential custody, approval, merge authority |
| Authority adapter | External owner/reviewer identity and exact-candidate binding | Reading authority from candidate-controlled prose |
| CI adapter | Trusted non-interactive metadata and stable result publication | Prompts, local user assumptions, candidate-selected evaluator weakening |

### Repository-local data model

- Git commits and trees are canonical for candidate identity and history.
- `docs/artifacts/changes/<change-id>.md` remains the only planning artifact for
  Tier 2; Tier 3 may add its one verification record and one audit record.
- One versioned, strict repository configuration file declares explicit
  verification commands and non-secret defaults. Its exact name and schema are
  selected during technology selection.
- Transient session input, bounded output, cancellation state, and external
  authority envelopes live under the repository's Git common directory, not in
  tracked product files. They are disposable and must never contain credentials.
- Accepted workflow state is reconstructed; it is not advanced by editing a
  mutable status field or replaying free-form agent output.
- Local writes use create-temp, flush, close, and atomic rename on the same
  filesystem. Readers reject symlinks, non-regular files, oversized values,
  unknown schema fields, and conflicting records.

### Provider and process boundary

The CLI invokes installed Codex and Claude executables using explicit argument
arrays and a minimal allowlisted environment. It does not embed provider SDKs or
make network requests. The adapter maps a shared request—role, scope, candidate,
task, acceptance criteria, permissions, and output contract—to capabilities the
installed client actually exposes.

Provider output is untrusted data. Structured results may supply observations and
references, but never owner approval, reviewer independence, command authority,
or merge permission. Unsupported or changed client behavior produces a truthful
blocked/degraded result, not a guessed fallback.

### Concurrency and transaction model

- One mutating Level 7 operation may hold the repository lock at a time.
- Read-only status operations may run concurrently from immutable snapshots.
- Implementation and audit never run concurrently on the same candidate.
- Provider adapters may be developed in parallel only on isolated branches after
  the shared contract is frozen; integration remains sequential.
- A candidate change invalidates verification, review, and pending merge
  confirmation atomically on the next state reconstruction.
- Cancellation terminates the supervised child process group, records no success,
  releases the lock, and leaves the previous accepted state recoverable.

### Trust boundaries

1. Repository files are untrusted inputs, including instructions and agent text.
2. Git object identity proves bytes and ancestry, not human authority.
3. Repository-defined verification commands are explicitly authorized code
   execution and run only at the requested step.
4. Codex and Claude own their network and credential behavior; Level 7 passes no
   credentials and makes no network calls itself.
5. Owner approval, independent-review identity, and merge confirmation originate
   outside candidate-controlled text and bind the exact current candidate.
6. Trusted CI evaluates protected policy from the base revision so the candidate
   cannot weaken its own merge gate.

## Architecture failure modes

| Failure | Required behavior |
|---|---|
| Git repository missing, ambiguous, corrupt, or changes during evaluation | Stop without mutation; report the failed identity check and retry action. |
| Local state is partial, stale, oversized, symlinked, or schema-incompatible | Ignore no error and invent no state; quarantine or replace only after explicit recovery. |
| Codex or Claude is absent or its interface is incompatible | Report the unavailable capability; allow the other provider only where independence requirements still hold. |
| Agent hangs, floods output, forks children, or ignores cancellation | Enforce time/output/process-group bounds; terminate and record failure without advancing state. |
| Repository verification command is missing or fails | Remain `building` or return to remediation; never recommend merge. |
| Candidate changes during or after verification/review | Mark evidence and confirmation stale and require fresh checks. |
| Implementer attempts to issue owner approval or independent audit | Reject the identity collision regardless of prose or passing tests. |
| Cross-provider reviewer is unavailable | Block Tier 3 pending the other provider or a qualified human; never downgrade silently. |
| Concurrent mutation starts | Reject or queue before side effects; never merge two mutable session histories. |
| Merge preconditions change before confirmation | Recompute and request confirmation for the new exact candidate. |
| CLI crashes mid-write or mid-commit | Atomic state writes and Git recovery expose either the prior accepted state or an explicit incomplete operation. |
| Large monorepo exceeds configured resource bounds | Stop with measured diagnostics and a bounded targeting remedy; never fall back to an unbounded scan. |

## Technology candidates

The comparison assumes the approved short-lived modular CLI architecture and the
existing repository. Scores use `1` (poor) through `5` (strong); weighted totals
are out of `100`.

| Criterion | Weight | Go | Rust | Swift | TypeScript/Node |
|---|---:|---:|---:|---:|---:|
| Reuse of current verified code and harness | 20 | 5 | 1 | 1 | 1 |
| Deterministic safety and testability | 20 | 5 | 5 | 4 | 3 |
| Single-binary macOS distribution | 15 | 5 | 5 | 5 | 2 |
| Child-process and cancellation control | 15 | 5 | 5 | 4 | 4 |
| Small-to-monorepo performance | 15 | 5 | 5 | 4 | 3 |
| Build, test, cross-compile, and CI fit | 10 | 5 | 3 | 3 | 4 |
| Production dependency surface | 5 | 5 | 4 | 5 | 2 |
| **Weighted total** | **100** | **100** | **79** | **70** | **53** |

### Candidate assessment

- **Go** already has a pinned offline toolchain, deterministic harness,
  repository/Git controller code, pure evaluator/renderer packages, tests, CI,
  and reproducible builds. It produces a small operational footprint and has
  sufficient process/filesystem primitives without third-party packages.
- **Rust** offers excellent safety and performance but would introduce a second
  toolchain, rewrite verified Go behavior, expand supply-chain inputs, and delay
  the founder journey without a demonstrated requirement Go cannot meet.
- **Swift** is native to the macOS launch platform and locally available, but it
  reduces reuse, complicates Linux-based CI/tooling assumptions, and creates an
  early platform lock that conflicts with future portability.
- **TypeScript/Node** is productive for adapters but requires a runtime and
  dependency ecosystem for distribution, increases startup/memory cost, and
  weakens the single-binary/offline simplicity of the approved architecture.

## Selected technology

Select **Go 1.26.7**, compiled with `CGO_ENABLED=0`, using the standard library
for production code.

The repository already pins the baseline toolchain in `.go-version`, `go.mod`,
and `harness/toolchains.lock.tsv`. Product builds must use that repository-owned
toolchain path rather than an ambient `go` executable. A newer shadow toolchain
may detect future incompatibility but cannot silently define release output.

### Stack

| Concern | Selection | Constraint |
|---|---|---|
| Language/runtime | Go `1.26.7` | No runtime installation required for end users; production binary is static within supported macOS constraints. |
| CLI parsing | Go standard library | Versioned explicit subcommands and flags; no reflection-heavy CLI framework or hidden environment precedence. |
| Structured data | Strict JSON with integer schema versions | Reject unknown, duplicate, trailing, oversized, symlinked, and non-regular inputs; no YAML execution/tag surface. |
| Git integration | Supervised `git` subprocess with explicit argv | Git remains canonical; no reimplementation of object/ref semantics and no shell command construction. |
| Process supervision | `os/exec`, contexts, process groups, bounded pipes | Cancellation and timeout terminate the complete child process group; output and environment are capped. |
| Persistence | Atomic regular files under tracked `.l7/` config and the Git common directory for transient state | No database, daemon, global state service, credential store, or mutable status ledger. |
| Provider integration | Installed `codex` and `claude` executables | No provider SDK dependency and no Level 7 network client; adapters expose only tested capabilities. |
| Testing | Standard `testing`, fake executables, temporary Git repositories, golden JSON only where semantic | No network in blocking tests; adversarial filesystem/process/state cases are mandatory. |
| Benchmarking | Go benchmarks plus generated small/medium/monorepo Git fixtures | Baselines record host/toolchain context; comparisons never pretend unlike hosts are equivalent. |
| Logging/output | Decision-first text and versioned JSON/JSONL | Structured events contain outcomes and references, not hidden reasoning, credentials, or full agent transcripts by default. |
| Distribution | Signed/notarized macOS archives containing one `l7` binary and notices | Release mechanics remain Tier 3 and are not authorized by this selection. |

### Initial command contract

The v1 command surface is intentionally small:

```text
l7 adopt
l7 status [--json]
l7 brief
l7 run --agent codex|claude
l7 verify
l7 review --agent codex|claude|human
l7 ready [--json]
l7 merge
l7 doctor [--json]
```

`l7 merge` always recomputes readiness and prompts for the exact candidate. CI
uses non-interactive `status`, `verify`, and `ready`; it cannot invoke an
interactive confirmation path or manufacture approval.

### Configuration and local state

- Track one `.l7/config.json` containing schema version, explicit verification
  argv arrays, command time/output limits, protected path additions, and
  non-secret provider adapter preferences.
- Store external approval, review bindings, locks, bounded run metadata, and
  disposable session handoff under `$(git rev-parse --git-common-dir)/l7/`.
- Do not store access tokens, provider credentials, raw hidden reasoning, full
  environment dumps, or unbounded stdout/stderr.
- Treat user and repository configuration as untrusted input. Repository config
  may narrow automation but cannot weaken built-in protected boundaries.
- Schema migration is explicit, version-to-version, atomic, and covered by
  forward/backward incompatibility tests; unknown future versions fail closed.

### Codex compatibility baseline

The locally inspected baseline is `codex-cli 0.149.1`. Its current non-interactive
surface exposes `codex exec`, stdin prompts, `--json` JSONL events,
`--output-schema`, explicit working directory, sandbox modes, ephemeral sessions,
resume/fork, and `exec review`.

The adapter will use only a tested subset and will never pass dangerous bypass
flags. Write work defaults to the narrowest compatible workspace sandbox; audit
work requires read-only access. User configuration is not assumed stable or
equivalent to Level 7 authority.

### Claude Code compatibility baseline

The locally inspected baseline is Claude Code `2.1.241`. Its current
non-interactive surface exposes `--print`, JSON/stream-JSON output, JSON Schema,
explicit tools and tool denials, permission modes, session IDs/resume, bare/safe
modes, and worktree support.

The adapter will use only a tested subset, explicit tool bounds, and a permission
mode appropriate to implement or audit. It will never pass
`--dangerously-skip-permissions`. Because non-interactive mode may skip workspace
trust UI and silently ignore invalid settings, Level 7 must validate its own
generated settings and trusted-directory preconditions before launch.

### Provider compatibility policy

- The two locally observed versions are evidence baselines, not timeless minimum
  versions or claims about every installation.
- `l7 doctor` records executable path, exact version, supported tested
  capabilities, and any degradation before a session starts.
- Unknown versions may run only after a local compatibility probe passes the
  adapter's inert contract tests. Mutation and independent-audit capability fail
  closed when required semantics are unknown.
- Provider-specific argv construction is isolated and table-tested. Shared domain
  code never switches behavior on provider prose.
- Fake executables make blocking CI deterministic. Actual-host provider tests are
  separately labeled, non-secret, and required before claiming a version
  supported; network/model success is never assumed from `--help` output.
- Provider auto-update is outside Level 7. A changed executable invalidates its
  cached capability result and requires a fresh probe.

### Build and dependency policy

- Keep production dependencies at zero for the initial walking skeleton. Any new
  module requires a concrete missing capability, version pin, license review,
  provenance record, and rollback decision in the active change brief.
- Commit `go.mod`/`go.sum` changes together; `go mod tidy -diff`, `go mod verify`,
  formatting, vet, typecheck, tests, and reproducible build remain blocking.
- Build macOS `arm64` and `amd64`; execute blocking tests on supported real macOS
  runners rather than treating cross-compilation as runtime compatibility.
- Embed version, commit, and dirty-state metadata from the trusted build process;
  runtime output must distinguish development, verified, and released binaries.
- Release signing, notarization, checksums, SBOM/notices, and publication are
  Tier 3 supply-chain controls, not ordinary development artifacts.

## Product harness

The existing pinned Go harness is the baseline rather than a second build system.
It already runs offline module checks, formatting, shell syntax, trusted policy,
import boundaries, vet, typecheck, all Go tests, and a repeat-build comparison.
Because its Go commands target `./...`, new product packages enter the standard
gates automatically.

Baseline verification on `2026-08-27` at commit
`2c0e1c97c0344e423a75b01fa3d1a0dc423a2b9d`:

| Command | Result |
|---|---|
| `make install` | PASS — no module dependencies; module files verified and tidy |
| `make lint` | PASS — lean policy, four-package import check, formatting, shell syntax, and vet |
| `make typecheck` | PASS — evaluator, harness, buildcontrol, and renderer packages |
| `make test` | PASS — all current Go tests |

These results prove the current foundation harness only. They do not prove the
future CLI, macOS compatibility, provider invocation, or product performance.

### Planned repository layout

```text
cmd/l7/                         executable composition root only
internal/l7/domain/             pure risk, lifecycle, scope, candidate, evidence
internal/l7/app/                use cases and effect-port interfaces
internal/l7/adapter/config/     strict tracked configuration
internal/l7/adapter/git/        Git identity, diff, commit, merge preconditions
internal/l7/adapter/state/      atomic Git-common-dir runtime state
internal/l7/adapter/process/    bounded child-process supervision
internal/l7/adapter/provider/   shared provider contract and fake adapter
internal/l7/adapter/codex/      Codex argv/events/capability mapping
internal/l7/adapter/claude/     Claude argv/events/capability mapping
internal/l7/adapter/verify/     repository-defined command execution
internal/l7/adapter/authority/  external owner/reviewer bindings
internal/l7/presentation/       text and versioned JSON/JSONL output
internal/l7/testkit/            test-only repositories and fake executables
testdata/l7/                    bounded static malformed/config/output fixtures
```

The domain package has no side-effect imports. The application package depends
only on domain types and interfaces it owns. Adapters depend inward and are wired
only by `cmd/l7`. Production packages never import `internal/harness`; harness
code may test product packages from the outside.

### Test layers

| Layer | Purpose | Blocking boundary |
|---|---|---|
| Domain unit tests | Exhaustive risk/state transitions, scope, freshness, readiness, and failure semantics | Every accepted state has a next action; invalid transitions and stale evidence always fail. |
| Application tests | Use-case sequencing with fake ports, cancellation, confirmation, and rollback | No effect occurs before its authorization/precondition; failures preserve the prior accepted state. |
| Adapter contract tests | Strict config/state parsing, Git behavior, process groups, output bounds, Codex/Claude argv and event decoding | Malformed, unknown, partial, symlinked, oversized, and incompatible inputs fail predictably. |
| Integration tests | Temporary real Git repositories plus fake provider and verification executables | Adopt through ready works in both provider orders without network or credentials. |
| CLI end-to-end tests | Compiled `l7` process, terminal/nonterminal modes, JSON schema, exit codes, signals | Human and machine contracts remain compatible and decision-first. |
| Adversarial tests | Shell injection, path traversal, race/change-after-check, forked children, self-review, stale approval, corrupt state | Zero authority bypasses, false-ready outcomes, unbounded output, or deadlocked accepted states. |
| Fuzz tests | JSON/event parsers, path/scope input, state reconstruction | No panic, uncontrolled allocation, silent acceptance, or nondeterministic result. |
| Benchmarks | Small, medium, and generated monorepo status/policy/config/event workloads | Regressions above 10% require an explicit accepted tradeoff; memory and file-count behavior are reported. |
| Actual-host compatibility | Real macOS arm64/amd64 binary and separately authorized installed Codex/Claude probes | Required before claiming an OS/provider version supported; cross-compilation and fake clients are insufficient. |

Tests use deterministic clocks, explicit environment maps, temporary directories,
fake executables, and bounded fixtures. They do not depend on network, user home
configuration, ambient credentials, provider billing, or mutable global caches.

### CI and merge gates

- Retain the blocking Ubuntu baseline harness and non-blocking shadow Go job.
- Add a blocking macOS baseline job when the proving CLI package lands; it runs
  install, lint, typecheck, tests, end-to-end CLI tests, and benchmarks within a
  declared variance policy for both release architectures where runners permit.
- Retain trusted-base policy evaluation and non-author review as merge gates.
- Add selected product-layer import/effect boundaries before domain/application
  implementation. Because CI workflow and harness policy files are protected,
  that harness-enabling change is Tier 3 and requires explicit owner approval and
  independent audit.
- Keep fake-provider tests blocking and credential-free. Actual Codex/Claude
  network/model trials run only in a separately authorized protected context and
  publish capability observations, never reusable credentials or blanket quality
  claims.
- Pin external CI actions by immutable commit. Do not grant write permissions to
  candidate verification jobs.
- A release candidate additionally requires clean-tree reproducibility, macOS
  runtime compatibility, provider matrix results, notices, signing/notarization,
  rollback, and independent Tier 3 audit.

### Import and effect gates

The product boundary checker must enforce at least:

- `internal/l7/domain` imports no filesystem, process, environment, network,
  clock, randomness, terminal, adapter, or harness package;
- `internal/l7/app` imports only domain plus standard-library pure types and never
  concrete adapters;
- provider adapters do not import one another;
- only the process adapter invokes arbitrary child processes;
- only the Git adapter invokes `git` or mutates Git state;
- only the state adapter writes Level 7 runtime files;
- only the command composition root constructs concrete adapters;
- no production package imports testkit, fixtures, or harness packages; and
- no package other than provider adapters references provider-specific flags or
  event schemas.

### Logging and diagnostics contract

- Default human output leads with `PASS`, `BLOCKED`, `FAILED`, or `CANCELLED`,
  then current state, reason, and one executable next action.
- `--json` emits one versioned final object; streaming operations may emit
  versioned JSONL events followed by exactly one terminal result.
- Diagnostics go to stderr and never alter machine-readable stdout.
- Every failure has a stable rule/code, bounded subject, safe message, and
  remediation. Error text does not include secrets, full environment values, or
  unbounded provider output.
- Logs record Level 7 observations and effect summaries, not hidden model
  reasoning or full transcripts. Persistent run records are opt-in, bounded,
  disposable, and stored under the Git common directory.
- Timestamps and duration are observational only and cannot affect domain state.

### Environment and configuration

The current `.env.example` remains documentation-only and must never be
auto-loaded. It contains no provider credentials or authority values. Product
configuration moves to strict `.l7/config.json`; provider clients retain custody
of their own authentication.

Environment variables are limited to documented non-secret CI overrides with
deterministic precedence. Approval, reviewer identity, risk, scope, executable
paths, shell fragments, and dangerous permission bypasses cannot be smuggled
through ambient environment inheritance.

### Harness acceptance criteria

- Current `make install`, `make lint`, `make typecheck`, and `make test` remain
  green before and after each slice.
- The first proving slice adds only a version/help/status composition path plus
  tests; it cannot launch agents, edit repositories, commit, review, or merge.
- Selected package boundaries are enforced before effects are implemented.
- Product tests are offline, deterministic, parallel-safe, and leave no process,
  lock, Git, or filesystem residue.
- At least one end-to-end fixture exercises every P0 state and interruption edge
  in both provider orders before P0 completion.
- macOS runtime and actual provider compatibility are verified before the
  corresponding support claim, not inferred from compilation or help text.
- CI, tests, and repository records remain evidence only; founder approval and
  independent-review authority stay external and exact-candidate-bound.
- README documents only working commands and verified support. `.env.example`
  remains secret-free and no product dependency is added without the approved
  dependency policy.

## Implementation waves

The P0 journey is delivered in four sequential waves. Each wave after Wave 1 gets
its own concise brief against the merged predecessor; this foundation brief is
not continually updated into another governance chain.

| Wave | Backlog scope | Risk | Outcome | Exit gate |
|---|---|---|---|---|
| **1 — trusted harness and inert CLI** | Proving subset of `CLI-001`, foundations for `CLI-010` | Tier 3 | Protected import/CI gates exist before effectful product code; `l7 version`, help, and truthful not-yet-available status compile and render stable text/JSON. | Owner approval before build; offline verification; independent audit; no agent, repository mutation, commit, review, or merge capability. |
| **2 — local lifecycle walking core** | `CLI-001`–`CLI-003` | Tier 2 unless scope changes | Adopt, brief, Git-derived status, pure lifecycle, strict config, atomic local state, interruption recovery, and conventional local commit preparation work without agent execution. | Complete unit/application/integration/adversarial tests; normal independent code review; one Wave 2 brief. |
| **3 — verification and provider execution** | `CLI-004`–`CLI-007` | Tier 3 | Bounded repository commands, Git commits, Codex/Claude adapters, capability probes, cancellation, and implement/review orchestration work in both provider orders. | Explicit owner approval, fake-provider and actual-host compatibility evidence, security/process audit, no merge capability. |
| **4 — ready, merge, CI, and P0 quality gate** | `CLI-008`–`CLI-010` | Tier 3 and release when shipped | Exact-candidate readiness, founder-confirmed merge, headless CI, macOS matrix, benchmarks, adversarial lifecycle coverage, support claims, and rollback complete P0. | Full verification, cross-provider and independent release audit, founder merge/release decision; no automatic deployment. |

P1 and P2 are not pre-authorized. After P0 outcome review, each promoted item
enters `l7-change` with its own proportional tier and brief only when required.

## Wave dependency graph

```text
Wave 1: trusted harness + inert CLI
   |
Wave 2: pure lifecycle + local continuity
   |
Wave 3: verification + Git effects + Codex/Claude orchestration
   |
Wave 4: readiness + confirmed merge + CI/performance/release evidence
```

No wave may begin from an unmerged or unaudited required predecessor. A failed
test or audit returns to the same wave for remediation and fresh verification;
it does not authorize downstream work.

## Wave ownership and shared files

| Area | Primary wave owner | Shared-file rule |
|---|---|---|
| `cmd/l7` composition | One integrator per wave | Updated only after inward contracts pass; never used as a policy implementation package. |
| `internal/l7/domain` | Wave 1 seed, Wave 2 owner | Wave 3/4 consume accepted contracts; material lifecycle changes return to Wave 2 ownership and invalidate downstream evidence. |
| `internal/l7/app` | Wave 1 seed, Wave 2 owner | Effect ports freeze before adapters; downstream changes require application tests first. |
| Provider contract | Wave 3 integrator | Freeze shared request/result/capability types before separate adapter work. |
| Codex adapter | Isolated Wave 3 branch/agent | Must not edit Claude adapter or shared domain while parallel work is active. |
| Claude adapter | Isolated Wave 3 branch/agent | Must not edit Codex adapter or shared domain while parallel work is active. |
| Git/process/state/verification adapters | One sequential owner | These share effect and transaction semantics; do not parallelize them in one worktree. |
| `Makefile`, workflows, harness policy | Wave 1 only until a separately approved control change | Protected files cannot be casually edited by product waves. |
| README/support claims | Integrator after evidence | Document only commands and compatibility demonstrated by the current candidate. |
| Change brief, verification, audit | Brief owner, verifier, independent auditor respectively | Stay within the tier artifact budget; audit writer is read-only and separate from implementer/remediator. |

## Parallelism limits

- Default execution is sequential Wave 1 → 2 → 3 → 4.
- Do not run multiple mutating agents in the same worktree.
- Only the Codex and Claude adapter implementations may proceed in parallel, and
  only after the provider-neutral contract is committed and each uses an isolated
  branch/worktree with disjoint files.
- The adapter integrator does not act as the Tier 3 independent auditor.
- Tests, read-only analysis, and benchmark fixture generation may run in parallel
  when they do not mutate shared caches or candidate files.
- Merge one adapter at a time, rerun the complete Wave 3 suite after each merge,
  then perform cross-provider behavior tests on the integrated candidate.
- If a supposedly isolated change needs a shared contract edit, stop parallel
  work, integrate or discard the branch safely, revise the contract sequentially,
  and rebase both adapters on the new accepted commit.

## Commit and review cadence

Each wave uses small conventional commits whose tests pass at the commit boundary.
The intended cadence is:

1. brief-only planning commit;
2. pure types/contracts and tests;
3. one adapter or use case with its tests;
4. integration and adversarial tests;
5. documentation/support-claim update after behavior works;
6. verification record when required;
7. independent audit-only successor when required.

Do not combine generated evidence with implementation, amend reviewed commits, or
continue implementation after verification without invalidating and recreating
the evidence for the new candidate.

## Wave 1 objective and non-goals

Wave 1 proves that the selected architecture can enter the repository under the
required quality gates. It adds an inert, deterministic command surface and the
minimum protected harness changes needed to test future boundaries.

Wave 1 explicitly does not:

- discover or modify a target repository;
- create `.l7/config.json` or runtime state;
- invoke Git, Codex, Claude, shells, verification commands, or network clients;
- create commits, approvals, reviews, audits, merges, or releases;
- claim macOS/provider support beyond the exact tests actually executed; or
- modify the existing lean governance controller.

## Wave 1 acceptance criteria

- `l7 version`, help, and unavailable `status` behavior are deterministic in text
  and JSON, with stable exit codes and decision-first output.
- The pure seed domain/application packages have no side-effect imports and
  expose no future effect as working.
- The import checker enforces the selected inward dependency boundaries before
  effect adapters arrive.
- The baseline and macOS jobs build and test the inert CLI; repeat builds produce
  identical bytes under the declared build inputs.
- Tests cover malformed argv, unknown commands/flags, terminal vs nonterminal
  output, cancellation before execution, and truthful unsupported status.
- `make policy-check`, `make verify`, action/workflow lint, manifest validation,
  and diff hygiene pass.
- The candidate adds no production dependency and stays within the exact scope
  below.
- A distinct independent auditor confirms the protected harness cannot be
  weakened by the candidate it evaluates and returns `GO` before merge.

## Scope

This foundation step defines v1 requirements, backlog, approved architecture,
technology, harness, and the proposed implementation waves. It also defines the
exact Wave 1 scope below. Later waves use new concise briefs rather than modifying
this foundation record. No separate requirements, design, architecture, approval,
or candidate-manifest artifacts will be created for Wave 1.

## Exact implementation file set

Add:

- `cmd/l7/main.go`
- `cmd/l7/main_test.go`
- `internal/l7/domain/result.go`
- `internal/l7/domain/result_test.go`
- `internal/l7/app/app.go`
- `internal/l7/app/app_test.go`
- `internal/l7/presentation/output.go`
- `internal/l7/presentation/output_test.go`
- `docs/artifacts/changes/standalone-cli-v1-verification.md`
- `docs/artifacts/changes/standalone-cli-v1-audit.md` (independent reviewer only)

Modify:

- `.github/workflows/harness.yml`
- `Makefile`
- `README.md`
- `harness/import-boundaries.tsv`

No other Wave 1 path is authorized. If implementation demonstrates that one
listed source split is unnecessary, omit it rather than creating a replacement
path. Any additional or renamed path requires an explicit brief revision and
fresh owner approval before it changes.

## Acceptance criteria

- The requirements reflect the founder's decisions: solo-founder focus;
  standalone macOS CLI; Codex and Claude support; repository-local state;
  technology-neutral verification; small-to-monorepo scale; local autonomous
  editing/testing/commits; cross-provider Tier 3 review; explicit confirmation
  for protected actions and every merge; and headless CI support.
- Functional and nonfunctional requirements are independently testable or name a
  later measurable baseline.
- Quality and both forms of performance are measured without making unsupported
  benefit claims.
- The document does not authorize product implementation or create another
  governance chain.

## Risks and mitigations

- **Scope breadth:** one end-to-end journey across two providers and arbitrary
  stacks can exceed v1 capacity. The backlog must select a walking skeleton and
  defer depth without weakening the required acceptance boundaries.
- **Provider mismatch:** Codex and Claude expose different invocation, permission,
  and session semantics. Use adapters with an explicit capability contract and
  report unsupported behavior truthfully.
- **Local-state tampering:** repository-controlled state is reviewable but not an
  authority source. Bind decisions to Git and obtain approval/review identity from
  trusted interaction or CI inputs.
- **Agent subprocess risk:** bound arguments, environment, filesystem scope,
  output, cancellation, and credentials; never interpret agent prose as command
  authority.
- **Monorepo cost:** prefer Git-index and changed-scope operations, benchmark
  representative fixtures, cap resource use, and degrade with actionable errors.
- **Quality theater:** measure false-ready, reopen/revert, benchmark, and bypass
  outcomes rather than document volume or agent activity.

## Rollback

Before implementation, rollback is reverting this planning-only commit. During
implementation, each scoped conventional commit must remain
independently revertible. Repository adoption must provide an uninstall path that
removes Level 7-generated configuration and state without altering product code
or Git history. Production rollback remains repository-specific and explicitly
approved.

## Foundation decision

The requirements, backlog, Option A architecture, Go selection, harness,
four-wave plan, and exact Tier 3 Wave 1 scope were accepted in the active
interaction. This authorizes freezing the brief as the planning boundary only.
It does not authorize implementation, merge, release, or deployment.

The next transition is `l7-build` for Wave 1. Before implementation, a fresh
external owner approval must bind the committed brief and name a distinct
implementer. Passing tests or this recorded decision cannot supply that authority.
