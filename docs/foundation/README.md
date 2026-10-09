# Product foundation status

This is the single status record for the six-step product foundation defined
by `l7-greenfield`. Evidence decides each status; a status label is never
evidence by itself. Re-assess a step when its cited evidence changes after the
commit recorded below, or when an earlier step changes.

| Step | Status | Assessed complete at |
|---|---|---|
| 1. Requirements | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| 2. Feature backlog | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| 3. Architecture | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| 4. Technology selection | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| 5. Harness | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |
| 6. Orchestration plan | `complete` | `d82db9e7b0284653d22d5669dcbdb87293a692d0` |

First assessed on 2026-10-09. The cited `docs/artifacts/` documents are
historical records: they are referenced, not edited. New or extended
foundation documents belong in this directory.

## 1. Requirements

| Criterion | Evidence |
|---|---|
| Problem | `docs/artifacts/requirements.md` §3; `docs/artifacts/concept-brief.md` |
| Target users | `docs/artifacts/requirements.md` §1 (approved primary-user assumption), §7 |
| Functional requirements | `docs/artifacts/requirements.md` §9.1–§9.10 |
| Non-functional requirements | `docs/artifacts/requirements.md` §10 (`L7-NFR-001`–`035`) |
| Constraints | `docs/artifacts/requirements.md` §13 |
| Measurable success metrics | `docs/artifacts/requirements.md` §12 |
| Risks | `docs/artifacts/requirements.md` §14 |

Gaps: none against the criteria.

Currency: the requirements predate the shipped v1 orchestration engine and the
solo-assurance cutover. §11 lists any Level 7-executed external mutation as a
v1.0 non-goal and §13 assumes one worktree at a time. For the opt-in crew,
`requirements-amendment-crew-delivery.md` amends both: parallel worktrees,
pull-request delivery, and owner-confirmed merge, each with its limits. Any
other work that pushes, publishes pull requests, or merges on a forge still
needs its own recorded amendment here before it starts.

## 2. Feature backlog

| Criterion | Evidence |
|---|---|
| P0 features with dependencies, effort, and acceptance criteria | `docs/artifacts/feature-backlog.md` §4 (summary table and per-item acceptance criteria) |
| P1 features with dependencies, effort, and acceptance criteria | `docs/artifacts/feature-backlog.md` §5 |
| P2 features with dependencies, effort, and acceptance criteria | `docs/artifacts/feature-backlog.md` §6 |

Gaps: none against the criteria.

Currency: features shipped in `1.0.0-dev` (onboard, sync, cyber, headless) map
to P1 items `L7-BL-024` and `L7-BL-026` rather than to discrete entries.
Parallel multi-agent work belongs to `L7-BL-026`; remote delivery and automatic
merging belong to P2 items `L7-BL-035` and `L7-BL-036`, which require an
autonomy charter before promotion.

## 3. Architecture

| Criterion | Evidence |
|---|---|
| Three options scored against the requirements | `docs/artifacts/architecture.md` §4 (options A, B, C), §5 (weighted totals 70.4, 83.4, 74.4) and §5.1 (hard gates) |
| Selected design with rationale | `docs/artifacts/architecture.md` §5.2 (option B) |
| Failure modes | `docs/artifacts/architecture.md` §4 (failure boundary per option), §14 |

Gaps: none against the criteria.

Currency: the shipped engine realises the selected local-kernel design with a
different package layout from §7.2: `cmd/l7` and
`internal/l7/{domain,app,adapter,presentation}`, with dependency direction
enforced by `make l7-import-closure-check`. Planned packages such as
`internal/kernel` and `internal/policy` do not exist.

## 4. Technology selection

| Criterion | Evidence |
|---|---|
| Scored candidates | `docs/artifacts/technology-selection.md` §4.1 (Go 84.6, Rust 78.0, TypeScript/Node 44.2), §5, §6 |
| Selected stack | `docs/artifacts/technology-selection.md` §4.2, §7 |
| Compatibility | `docs/artifacts/technology-selection.md` §19; `distribution/compatibility.json` |
| Pinned production dependency versions | `go.mod` (`toolchain go1.26.7`, `github.com/odvcencio/gotreesitter v0.24.0`), `go.sum`, `.go-version`, `harness/toolchains.lock.tsv`; checked by `make install` (`go mod verify`, `go mod tidy -diff`) |

Gaps: none against the criteria.

Currency: §7 recorded zero production modules. The `gotreesitter` dependency
used by codebase memory was added later and is not yet recorded in any
technology decision.

## 5. Harness

| Criterion | Evidence |
|---|---|
| Repository layout | `README.md` (Development and contributing) |
| Lint | `Makefile` targets `lint` and `technical-lint` (format, import boundaries, shell syntax, `go vet`) |
| Type checking | `Makefile` target `typecheck` |
| Tests | `Makefile` targets `test`, `race-check`, `fuzz-check` |
| CI | `.github/workflows/harness.yml` |
| Logging | `L7_LOG_FORMAT`/`L7_LOG_LEVEL` in `Makefile` and `.env.example`; structured-logging proof in `docs/artifacts/harness.md` §3 and `internal/harness/proving_test.go` |
| README | `README.md` |
| `.env.example` | `.env.example` (`L7_LOG_*`, `L7_ASSURANCE_MODE`, `L7_TELEMETRY`, `L7_NETWORK`) |
| Install, lint, typecheck, and test commands verified passing | Hosted Harness run [37815110512](https://github.com/addressanup/level7-dev-loop/actions/runs/37815110512) at `d82db9e` ran `make bootstrap-ci` and `make ci` (install, technical lint, typecheck, test, race, fuzz, reproducibility, distribution); all four jobs passed |

Gaps: none against the criteria. Local `make` was not run during this
assessment because the host's Xcode license prompt blocks `git`, `make`, and
`python3`; the hosted run at the exact commit is the evidence.

Currency: `docs/artifacts/harness.md` describes the original zero-dependency
harness; the `Makefile` and workflow above are the current harness.

## 6. Orchestration plan

| Criterion | Evidence |
|---|---|
| Waves | `docs/artifacts/orchestration-plan.md` §7 |
| Shared files | `docs/artifacts/orchestration-plan.md` §10 |
| Parallelism limits | `docs/artifacts/orchestration-plan.md` §11 (after Git setup: at most two disjoint writer worktrees plus one read-only reviewer, with serial integration) |

Gaps: none against the criteria.

Currency: the wave sequence describes the original v1.0 plan, and §11's
"Git is absent" statement is superseded because the repository now uses Git.
Its post-Git parallelism and serial-integration rules still apply to
development of this repository.
