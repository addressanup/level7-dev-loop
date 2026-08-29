# Level 7 Dev Loop

Level 7 is a lean, risk-proportionate development workflow for Codex and Claude
Code. Its common path is deliberately ordinary:

`brief → implement → test → review → merge`

Working software, Git identity, automated verification, and normal review carry
the evidence. Process expands only when risk does.

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

The reverse provider order is also modeled. Tier 3 prompts in an active terminal
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

### Exact remote candidate integration

The repository also contains an operator-only path for the exceptional case
where GitHub's enabled pull-request merge methods would replace an already
verified and audited commit identity. It is not part of the product CLI and must
not be used until its own Tier 3 verification, independent audit, and a separate
live-effect authorization are current.

`scripts/harness/exact-fast-forward-integration.sh` accepts only a full
repository name, pull-request number, base commit, candidate commit, candidate
tree, accountable owner, and auditor. It verifies the live pull request,
distinct exact-head approvals, latest required GitHub Actions checks, strict
branch protection, sole-administrator identity, source ref, tree, and
fast-forward ancestry. It then requires an active terminal and the complete
candidate SHA. The only effect path temporarily disables administrator
enforcement, performs one expected-old lease-bound fast-forward to
`refs/heads/main`, and restores enforcement before postflight. It has no squash,
rebase, merge, pull-request edit, branch deletion, generic force, or local-ref
path.

The full protection response is snapshotted and compared before and after the
effect. Restoration also runs on command failure and handled interruption. A
restoration failure stops in a recovery state whose sole next action is to
re-enable administrator enforcement; the tool never resets or force-pushes
`main`.

The offline contract harness uses fake `gh` and `git` executables and cannot
reach GitHub:

```sh
make exact-fast-forward-integration-check
```

Any real invocation must supply every binding explicitly; there are no defaults,
abbreviated SHAs, nonterminal confirmations, retries, or fallback merge methods.
Passing this harness is technical evidence only and grants no authority to run
the operator.

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

## Risk model

| Tier | Examples | Required process |
|---|---|---|
| 1 — routine | Docs, tests, refactors, cleanup, low-risk fixes | Concise task, implementation, relevant tests, clean diff, normal review. Zero governance artifacts. |
| 2 — product change | Features, meaningful UX, public interfaces, persistence | One `docs/artifacts/changes/<change-id>.md` brief; default-OFF flag when appropriate; tests and normal review. |
| 3 — high risk or release | Authorization/security, destructive behavior, material migrations, production release, governance controls | Brief, external owner approval, bound verification record, independent read-only audit, rollback. At most three artifacts. |

Only Tier 3 requires independent audit. Repository prose and passing tests never
constitute approval.

## Controller

The Go controller compares an exact Git base with a candidate commit/tree,
validates declared scope, elevates protected controls to Tier 3, enforces the
artifact budget, and validates external approval/audit bindings. It reports a
small state and one executable next action.

Local Tier 1 example:

```sh
L7_RISK_TIER=1 \
L7_BASE_REF=<base-commit> \
L7_SCOPE='docs/guide.md,internal/example_test.go' \
make policy-check
```

Tier 2/3 base, tier, change ID, and scope come from the one change brief. Explicit
local authority is stored outside tracked repository text under `.git/l7/`. CI
uses trusted review/event data.

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

Repository rules are part of the installation contract. They must restrict risk
labels to trusted maintainers; require the non-experimental `Harness` jobs
`Go 1.26.7 (baseline)`, `CLI macOS 15 (arm64)`, `CLI macOS 15 (amd64)`, and
`CLI paired benchmark gate`, plus the `Trusted policy` evaluation; dismiss stale
reviews; require at least one non-author approval; and require CODEOWNER/owner
review for protected paths. Workflow YAML does not make a check blocking by
itself. Installation and upgrade must verify those required checks against the
live repository ruleset before claiming they are blocking. Trusted policy reads
exact-head check and review identities before it reports `ready`; it does not
infer them from candidate text.

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

Start with `l7-next`. The 12 skills share the same risk tiers and artifact budget;
`l7-release` is reserved for Tier 3 and production release validation.

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
make cli-benchmark-check L7_BENCHMARK_BASE_ROOT=/absolute/path/to/base-checkout
```

The baseline, paired benchmark, and both native macOS architecture jobs are
required by the installation contract above; the configured shadow toolchain
remains non-blocking. Each macOS job runs the same offline suite plus declared
macOS cross-builds. The actual-host target compiles tagged probes with
`-run '^$'`; it never launches a provider. This local repository has no remote or
live ruleset evidence, and no hosted or Intel job was run for this candidate, so
it does not claim those jobs are currently blocking or that Intel runtime support
was observed. Trusted policy remains a separate required gate so technical CI
neither needs nor manufactures approval. These checks are evidence, not approval.
