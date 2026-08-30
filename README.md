# Level 7 Dev Loop

Level 7 is a solo-first, risk-proportionate development workflow for Codex and
Claude Code. Its default path is one uninterrupted conductor loop:

`intent → inspect → implement → test → repair → self-review → handoff`

Working software, Git identity, automated verification, and truthful self-review
carry the evidence. Process expands at real risk and effect boundaries, not
because a solo developer lacks a second identity.

## Solo fast plugin loop

`l7-next` is the default conductor even when it is not named. One concrete user
request authorizes ordinary repository-local reversible inspection,
implementation, testing, repair, and self-review. The conductor may apply the
other Level 7 skills internally; it does not ask the user to approve skill
transitions.

Solo assurance is the default. Tier 1 creates no governance artifact; Tier 2 and
Tier 3 use at most one concise brief. Solo verification and review stay in Git
and CI, so they do not create evidence-only commits or require a fabricated
independent auditor. `L7_ASSURANCE_MODE=team` explicitly opts a repository into
distinct owner/auditor controls when those real identities exist.

External, destructive, irreversible, credentialed, production, publication,
deployment, release, and protected-branch merge effects still require explicit
authority at the actual boundary.

## Install the v0.1.1 plugin

The release packages the 12 Level 7 instruction skills for Codex and Claude
Code. It does not install the experimental standalone `l7` Go binary, an MCP
server, hooks, executables, telemetry, or network behavior of its own. The
standalone CLI remains a separate `0.1.0-dev` preview.

For Codex CLI, add the tag-pinned GitHub marketplace and install Level 7:

```sh
codex plugin marketplace add addressanup/level7-dev-loop --ref v0.1.1
codex plugin add level7-dev-loop@level7-engineering
```

Start a new Codex task, then give one concrete objective:

```text
$l7-next Implement and verify <your objective>.
```

Remove only the Level 7 plugin and its marketplace source with:

```sh
codex plugin remove level7-dev-loop@level7-engineering
codex plugin marketplace remove level7-engineering
```

For Claude Code, add the same tag-pinned GitHub repository at user scope and
install Level 7:

```sh
claude plugin marketplace add addressanup/level7-dev-loop@v0.1.1 --scope user
claude plugin install level7-dev-loop@level7-engineering --scope user
```

Start a new Claude Code session, or run `/reload-plugins`, then invoke:

```text
/level7-dev-loop:l7-next Implement and verify <your objective>.
```

Remove it with:

```sh
claude plugin uninstall level7-dev-loop@level7-engineering --scope user
claude plugin marketplace remove level7-engineering --scope user
```

These commands select the immutable `v0.1.1` release tag. To move to a later
release, remove the installed plugin and pinned marketplace, add the new tag,
and install again. Start a new Codex task or reload Claude plugins after every
installation.

The v0.1.1 compatibility boundary is intentionally narrow: package build is
checked for macOS arm64 and amd64, while validation, discovery, installation,
inspection, removal, and cleanup through the committed local marketplace are
smoke-tested only on macOS arm64 with Codex CLI 0.151.0 and Claude Code 2.1.247.
Remote tag fetching and provider/model invocation for the `0.1.1` packages are
`NOT_RUN`; other host versions and architectures are unqualified, and formal
support is withheld. See `distribution/compatibility.json` for the exact
boundary.

## Wave 4 local readiness preview

The repository includes a default-OFF local lifecycle and execution preview. It
can adopt an existing Git worktree, record one proportionate change, reconstruct
its scope and identity from Git, run explicit repository checks, create one
controlled candidate commit, and coordinate implementation and review through
isolated Codex or Claude Code adapters. Wave 4 can additionally evaluate exact-
candidate readiness and, after immediate human confirmation, advance one
explicit local branch with an atomic fast-forward compare-and-swap.

```sh
make build
./build/bin/l7 help
./build/bin/l7 version --json
./build/bin/l7 status
```

Enable the preview explicitly in the target repository, then commit its tracked
configuration before opening a change:

```sh
l7 adopt --enable-local-lifecycle
git add .l7/config.json
git commit -m "chore: adopt Level 7"
```

Tier 1 stores only disposable context under Git's common directory and creates
no governance document:

```sh
l7 brief \
  --id routine-fix \
  --tier 1 \
  --problem "Fix a low-risk internal defect." \
  --scope 'internal/example/**'
```

Tier 2 and Tier 3 require problem, scope, acceptance criteria, risk, and rollback
input. They create exactly one tracked brief, which must be committed before
provider execution:

```sh
l7 brief \
  --id product-feature \
  --tier 2 \
  --problem "Add a bounded product feature." \
  --scope 'internal/product/**' \
  --accept "Relevant tests pass." \
  --risk "State could become stale." \
  --rollback "Revert the candidate."

git add docs/artifacts/changes/product-feature.md
git commit -m "docs(product): add product feature brief"
l7 status --json
```

Configure at least one repository-defined verification command as exact argv;
the CLI never inserts a shell. For example, the tracked configuration may use:

```json
"verification": [
  {"name": "test", "argv": ["make", "test"], "benchmark": false}
]
```

Then run the bounded lifecycle explicitly:

```sh
l7 run --agent codex --message "feat(product): implement product feature"
l7 verify
l7 review --agent claude
l7 ready
l7 status --json
```

The reverse provider order is also modeled. This standalone Wave 4 CLI preview
predates the solo plugin conductor and exposes an optional team-style lifecycle.
Its Tier 3 path prompts in an active terminal
for approval bound to the exact committed brief and selected implementer, then
requires the other provider for read-only review. It creates at most the one
brief, one verification record, and one independent audit record. Tier 1 creates
no governance artifact; Tier 2 keeps assurance in repository-local state.

The preview supports non-bare macOS Git worktrees with an initial commit. It
rejects dirty intake or index state, undeclared scope expansion, provider-created
commits, reviewer mutation, stale identities, protected controls below Tier 3,
unsafe configuration/state files, executable replacement, and concurrent Level
7 mutations. Git commit and tree IDs remain canonical; bounded recovery evidence
lives under `.git/l7/product/`. Cancellation, timeout, and aggregate output limits
terminate the inherited subprocess group. A bounded pipe-drain delay prevents a
session-escaped descendant from retaining the command wait and mutation lock;
the escaped process itself can still survive. This is process containment, not a
general OS sandbox. Controlled commits run repository hooks with system/global
Git configuration disabled and a bounded environment, so hooks may require
repository-local identity and must not depend on ambient configuration.

`l7 ready` recomputes the candidate, tree, tracked configuration digest,
verification checks, independent `GO`, identities, audit, and benchmark facts
under the repository mutation lock. A changed fact invalidates old readiness.
Tier 3 requires owner, implementer, and reviewer identities to be distinct and
requires a passing check explicitly designated as a benchmark by the trusted
repository configuration. The readiness receipt is repository-local evidence;
Git and the tracked configuration remain canonical.

Trusted automation may invoke `l7 ready --headless --json` with one strict,
size-bounded JSON document on stdin. This path composes only the pure CI decoder
and domain evaluator: it does not locate a repository, prompt, launch a provider,
write product state, access the network, or merge. The envelope and evaluator
must come from the trusted base; candidate prose and passing tests do not confer
authority.

After readiness, local integration is explicit and interactive:

```sh
l7 merge --target release-candidate
```

The target must already be a local branch at the recorded readiness base, must
not be checked out in another worktree, and must be an ancestor of the exact
candidate. Immediately before the effect, the CLI displays the old and new full
identities and requires the operator to type the full candidate SHA in an active
terminal. It then revalidates under lock and performs only
`git update-ref --no-deref <ref> <candidate> <expected-old>`. It never checks
out, fetches, pushes, rebases, resets, deletes, creates a merge commit, or touches
a remote. If receipt persistence is interrupted after the ref update, status
consults Git and the next interactive merge invocation can record recovery; it
never auto-resets the ref.

The adapters currently recognize only the provisional fixture baselines Codex
`codex-cli 0.149.1` and Claude Code `2.1.241`. Real provider launches and
actual-host trials are `NOT_RUN` and separately gated, so Wave 4 makes no current
provider-support claim. Build-tagged actual-host probes and both provider-order
orchestration compile during verification but are skipped; each real order also
requires an exact source-candidate authorization and live terminal confirmation
inside a disposable no-remote repository. Deployment, release, publication,
network orchestration, provider installation, and global configuration remain
unavailable. This binary is a development candidate, not a released or supported
CLI.

## v0.1 dual-host plugin packages

One strict descriptor at `distribution/package.json` generates the Codex,
Claude, root-plugin, and legacy marketplace metadata. The two host packages
share stable version `0.1.1` but have separate manifests, catalogs,
compatibility entries, inventories, digests, and archives.

```sh
make bootstrap
make distribution-check
make distribution
```

`make distribution-check` compares the tracked generated manifests with the
authored descriptor, builds each archive twice, compares exact bytes, reinspects
every allowlisted entry, and runs disposable filesystem fixtures for install,
identical reinstall, interrupted upgrade recovery, rollback, conflict preview,
and removal. The fixtures preserve unowned files and canonical project artifacts
and block on missing or stale ownership receipts. They do not invoke a host
package manager and are not actual-host lifecycle evidence.

For machine consumption, the underlying checker also accepts `--check --json`.
Its canonical report records Level 7's internal archive/catalog binding and the
internal structure, reproducibility, and fixture-lifecycle results. This remains
offline build evidence: neither host is documented to consume
`DISTRIBUTION.json` or enforce its digest. The package compatibility projection
separately carries the exact version-bounded macOS arm64 smoke observation. A
fresh offline check does not
rerun either host, evaluate release authority, sign, or publish, so its live host
cells remain `NOT_RUN`, authority remains `NOT_EVALUATED`, support remains
withheld, and `release_ready` remains false.

`make distribution` writes ignored release output under
`build/distributions/`. The Codex layout uses `.codex-plugin/plugin.json` and a
repo-marketplace catalog at `.agents/plugins/marketplace.json`; the Claude layout
uses `.claude-plugin/plugin.json` for its manifest and
`.claude-plugin/marketplace.json` for its catalog. Skills remain at each plugin
root.
See the [official OpenAI packaging contract](https://developers.openai.com/plugins/build/plugins)
and [official Claude plugin reference](https://code.claude.com/docs/en/plugins-reference).

Each package includes matching changelog and MIT license text, an explicit inert
permission declaration, a claim-withheld compatibility entry, a complete file
inventory, an SPDX SBOM, and an unsigned provenance input. Version `0.1.1` is a
distribution-only instruction-plugin patch, not a signature, notarization,
updater, or broad provider-support claim. A package result for one host cannot
promote the other. Any repeat host trial must bind the exact archive bytes and
use disposable host state, including an isolated `CODEX_HOME` or Claude
configuration, plugin-cache, and temporary roots; macOS Keychain and
provider-side request logs remain outside those filesystem isolation controls.

## Risk model

| Tier | Examples | Required process |
|---|---|---|
| 1 — routine | Docs, tests, refactors, cleanup, low-risk fixes | Concise task, relevant tests, clean diff, truthful self-review. Zero governance artifacts. |
| 2 — product change | Features, meaningful UX, public interfaces, persistence | One `docs/artifacts/changes/<change-id>.md` brief; default-OFF flag when appropriate; tests and self-review. |
| 3 — high risk or release | Authorization/security, destructive behavior, material migrations, production release, governance controls | One brief, stronger relevant verification, self-review, rollback, and explicit authority at the actual effect boundary. |

Solo mode never requires or claims independent audit. Team assurance is an
explicit trusted opt-in for repositories with genuinely distinct owner and
reviewer identities. Risk tier and collaboration topology remain separate.

## Controller

The Go controller compares an exact Git base with a candidate commit/tree,
validates declared scope, elevates protected controls to Tier 3, enforces the
assurance-specific artifact budget, and reports the selected assurance mode with
one executable next action. Solo is the default; team mode retains bound
owner/auditor validation for compatibility.

Local Tier 1 example:

```sh
L7_RISK_TIER=1 \
L7_ASSURANCE_MODE=solo \
L7_BASE_REF=<base-commit> \
L7_SCOPE='docs/guide.md,internal/example_test.go' \
make policy-check
```

Tier 2/3 base, tier, change ID, and scope come from the one change brief. Solo
mode uses Git/CI evidence and no approval/audit records. Team-mode authority is
stored outside tracked repository text under `.git/l7/`, and CI derives it from
trusted forge events.

```sh
make policy-check
make verify
make ready-check  # merge/release gate
```

The trusted-policy workflow builds the controller from the pull request's base
revision and evaluates the candidate read-only. Protected control changes
therefore cannot weaken the evaluator used by the merge gate.

Trusted CI requires exactly one maintainer-controlled label:
`l7-risk-tier-1`, `l7-risk-tier-2`, or `l7-risk-tier-3`. Tier 1 also requires one
explicit `L7-Scope: path,pattern` line in PR metadata; scope is never generated
from the candidate diff. Tier 2/3 scope remains in the bound brief. Missing or
conflicting classification fails closed, and authorization, security, migration,
deployment, workflow, harness, controller, skill, and plugin-control paths force
Tier 3.

Repository rules are part of the installation contract. Both modes must restrict
risk labels to trusted maintainers and require the non-experimental `Harness` jobs
`Go 1.26.7 (baseline)`, `CLI macOS 15 (arm64)`, `CLI macOS 15 (amd64)`, and
`CLI paired benchmark gate`, plus the exact-head `evaluate` check published by
Trusted policy. Solo mode sets required approving reviews to zero; team mode
dismisses stale reviews and requires its configured non-author/CODEOWNER review.
Workflow YAML does not make a check blocking by itself. Installation and upgrade
must verify those required checks against the live repository ruleset before
claiming they are blocking.

Harness runs for pull requests and pushes to `main`, so publishing a feature
branch does not start a redundant push suite. Trusted policy is triggered again
after the pull-request Harness run completes. Its base-revision controller remains
read-only with respect to the candidate; the workflow publishes the final
`evaluate` result against the exact PR head.

### One-time solo-policy cutover

GitHub evaluates a trusted policy workflow from the base branch, so the pull
request that first installs this solo policy is still subject to the previous
team-only gate. A solo maintainer must not fabricate an auditor to cross that
bootstrap boundary. After the exact cutover head passes Harness and self-review,
the safe migration is one explicit administrator merge/bypass for that immutable
head, if repository rules permit it. That protected-branch effect requires the
maintainer's specific authority and is not granted by ordinary implementation
instructions. Subsequent pull requests use the new solo default.

The benchmark job takes five alternating base/candidate samples on one macOS
host with one pinned toolchain, compares medians, reports every raw `ns/op`
sample, and blocks regressions above 10%. A failed regression can become current
only when the configured accountable owner submits an exact-head approved review
containing this exact line:

```text
L7-Benchmark-Regression-Accepted: <full-candidate-sha>
```

That external review marker cannot waive a missing, cancelled, or stale
benchmark check.

## Skills

`l7-next` is the only default entry point and conducts the complete local loop.
The other skills are internal execution lenses, not user approval checkpoints.
`l7-release` is used only for a real release boundary or opt-in team assurance.

## Historical records

The repository retains 64 pre-lean governance artifacts and their Git history.
Legacy phase registries, path rosters, ownership ledgers, candidate SHA manifests,
approval receipts, and repeated audit chains are deprecated: they remain evidence
of earlier work but are not active inputs and should not be continually updated.

No product runtime or supported production capability is introduced by the lean
workflow. Capability claims must remain tied to working code and verification.

## Harness

The harness keeps its pinned, repository-scoped Go toolchain, offline module
settings, import boundaries, lint/type/test gates, and repeat-build check.

```sh
make bootstrap
make policy-check
make verify
make build
make cli-cross-build
make cli-actual-host-compile
make distribution-check
make distribution
make cli-benchmark-check L7_BENCHMARK_BASE_ROOT=/absolute/path/to/base-checkout
```

The baseline, paired benchmark, and both native macOS architecture jobs are
required by the installation contract above; the configured shadow toolchain
remains non-blocking. Feature branches run this suite once through the pull
request event; pushes to `main` retain post-merge coverage. Each macOS job runs the same offline suite plus declared
macOS cross-builds. The actual-host target compiles tagged probes with
`-run '^$'`; it never launches a provider. This local repository has no remote or
live ruleset evidence, and no hosted or Intel job was run for this candidate, so
it does not claim those jobs are currently blocking or that Intel runtime support
was observed. Trusted policy remains a separate required gate so technical CI
neither needs nor manufactures approval. These checks are evidence, not approval.
