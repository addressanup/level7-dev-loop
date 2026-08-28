# Hosted CI Readiness Remediation

| Field | Value |
|---|---|
| Change ID | `hosted-ci-readiness-remediation` |
| Risk tier | `3` — protected harness and toolchain controls |
| Status | `proposed`; implementation is not approved |
| Base commit | `481adaaec967ac34b6b27cf78509b85d5c068abc` |
| Base tree | `d57a334696487b1d15557c9980e8a55c2dc4c930` |
| Audited evidence head | `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3` |
| Audited evidence tree | `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb` |
| Evidence PR | `addressanup/level7-dev-loop#1`; no merge is authorized |
| Selected disposition | Isolated clean-baseline remediation; preserve the audited provider branch and its history unchanged |
| Accountable owner | Active user approved investigation and this brief only; implementation requires fresh approval bound to this brief commit |
| Implementer | `codex-root` |
| Feature flag | Not applicable; no production behavior changes |

## Problem

Public PR `addressanup/level7-dev-loop#1` exposes two technical failures that
already exist at its exact base. Hosted run `33159727486` against audited head
`f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3` reported:

- `CLI macOS 15 (arm64)` and `CLI paired benchmark gate`: `PASS`;
- `Go 1.26.7 (baseline)`: `FAIL`;
- nonblocking `Go 1.27.0 (shadow)`: `FAIL` for the same test defect; and
- `CLI macOS 15 (amd64)`: `FAIL` before test execution.

The Ubuntu jobs fail at
`TestSnapshotRejectsInvalidMissingAndNonAncestorBases` because its direct
`git commit-tree HEAD^{tree} -m unrelated` test setup relies on ambient Git
author identity. A clean runner has none, so the test stops at
`internal/l7/adapter/git/repository_test.go:74`. The production adapter is not
implicated. The failure is reproducible without altering repository or global
configuration by supplying `user.useConfigOnly=true` and disabling ambient Git
configuration; the inline-identified initial commit succeeds and the
unidentified `commit-tree` call fails.

The Intel macOS job fails closed in `scripts/harness/bootstrap-go.sh` because
`harness/toolchains.lock.tsv` contains no unique baseline Go `1.26.7`
`darwin/amd64` record. The workflow intentionally requests the existing
`macos-15-intel` runner. Official Go download metadata identifies the missing
archive as `go1.26.7.darwin-amd64.tar.gz`, size `67852067`, SHA-256
`92e8b34bff3c89ab16404c595669ac8cb004cc2f676dcbd1f5b87a6b8def3b47`,
with its archive and detached signature under `https://go.dev/dl/`.

An exact base-to-evidence-head comparison confirms that
`internal/l7/adapter/git/repository_test.go`, `harness/toolchains.lock.tsv`,
`scripts/harness/check-foundation-scope.sh`, and
`.github/workflows/harness.yml` are unchanged. These defects therefore predate
the provider no-model gate harness and must not be folded into, blamed on, or
used to rewrite its audited commits.

Trusted-policy run `33159727689` is a separate authority blocker. It returned
`AUTH-001` because repository variable `L7_ACCOUNTABLE_OWNER` is absent and no
current external reviews exist. At proposal time the only repository
collaborator is PR author `addressanup`. Tier 3 hosted readiness requires an
accountable-owner GitHub approval and an auditor/reviewer GitHub approval on the
exact head, with both actors distinct from the PR author and from each other.
Repository code, local approval, passing technical checks, or a fabricated
identity cannot satisfy that requirement.

## Preserved provider disposition

The original provider base
`51191ad6edc670a0e73c3d152484bd57785144f7`, clean-baseline disposition head
`a3b40cbeebc81e89a469bdf3540fcbd1f83d2a7a`, failed brief
`438375b2d8edcec0983f9ce4eb4654a222cabd68`, and failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5` with tree
`7943f38db45705ce9cc1da01fb600f57e518215f` remain historical evidence.

Codex actual-host Gate 1 passed only for that failed candidate. Claude
actual-host Gate 2 remains `NO_GO`: both exact implementer and reviewer help
invocations succeeded, both unknown-option parser controls unexpectedly exited
successfully, both invalid `--max-turns not-an-integer` controls failed as
required, and neither help surface advertised `--max-turns`. Help advertisement
alone remains non-dispositive; the successful unknown-option controls remain
dispositive.

The audited no-model Codex and Claude parser gates at evidence head `f0e9f54c…`
remain `NOT_RUN`. This remediation must not weaken or remove unknown-option
rejection, typed `--max-turns` enforcement, or any argument, permission,
output-schema, cancellation, cleanup, reviewer-immutability, scope, or
containment control. It cannot transfer historical evidence, promote a provider
version, or create a compatibility claim.

## Scope

Apply one isolated clean-baseline remediation from exact base
`481adaaec967ac34b6b27cf78509b85d5c068abc`:

1. Give only the failing test-owned `commit-tree` operation an explicit,
   deterministic test identity. Do not change production Git behavior, the
   shared runner's environment contract, local/global Git configuration, or any
   identity/authorization policy.
2. Add exactly one baseline Go `1.26.7` `darwin/amd64` lock record using the
   official filename, byte size, SHA-256, archive URL, and detached-signature
   URL above. The existing bootstrap must continue to enforce one matching
   tuple, HTTPS-only Go download URLs, size and SHA-256 equality, pinned signing
   identity, detached-signature validity, safe archive members, and exact
   extracted Go version.
3. Update the foundation harness assertion from its historical four-record
   matrix to the exact five-record matrix. It must reject malformed, duplicate,
   missing, or additional records and preserve the existing baseline/shadow
   roles and versions.

Do not modify either workflow, any branch-protection rule, provider code or
test, compatibility profile, dependency, plugin or skill, production source,
historical manifest, prior brief/verification/audit record, remote, or global
configuration. Do not amend, rebase, cherry-pick into, or push the audited
provider branch as part of this change.

## Exact implementation file set

Add:

- `docs/artifacts/changes/hosted-ci-readiness-remediation.md`
- `docs/artifacts/changes/hosted-ci-readiness-remediation-verification.md`
- `docs/artifacts/changes/hosted-ci-readiness-remediation-audit.md`

Modify:

- `internal/l7/adapter/git/repository_test.go`
- `harness/toolchains.lock.tsv`
- `scripts/harness/check-foundation-scope.sh`

No other path is authorized. In particular,
`.github/workflows/harness.yml`, `.github/workflows/policy.yml`, `Makefile`,
`scripts/harness/bootstrap-go.sh`, all historical `*.sha256` manifests, the
audited provider-harness commits and artifacts, and the user-owned untracked
`docs/artifacts/foundation-rebaseline-admission-audit.md` remain unchanged.

## Acceptance criteria

1. The remediation starts from exact base
   `481adaaec967ac34b6b27cf78509b85d5c068abc`, tree
   `d57a334696487b1d15557c9980e8a55c2dc4c930`, on a separate branch. Audited
   evidence head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`, tree
   `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb`, and all original provider
   implementation and governance history remain unchanged.
2. The failing test creates its unrelated commit with explicit test-only author
   and committer identity. The focused test passes with global/system Git
   configuration disabled and `user.useConfigOnly=true`, without writing any
   repository, global, or system Git configuration.
3. Production repository discovery, snapshot, commit, review, merge, identity,
   and authorization behavior is byte-identical to the base; the test fix does
   not mask a product identity failure or alter the general `runGit` helper.
4. `harness/toolchains.lock.tsv` adds exactly this tab-separated record and no
   other semantic change:
   `baseline`, `1.26.7`, `darwin`, `amd64`,
   `go1.26.7.darwin-amd64.tar.gz`, `67852067`,
   `92e8b34bff3c89ab16404c595669ac8cb004cc2f676dcbd1f5b87a6b8def3b47`,
   `https://go.dev/dl/go1.26.7.darwin-amd64.tar.gz`, and
   `https://go.dev/dl/go1.26.7.darwin-amd64.tar.gz.asc`.
5. The foundation scope check requires exactly five unique declared tuples:
   baseline `1.26.7` for `darwin/amd64`, `darwin/arm64`, and `linux/amd64`, plus
   shadow `1.27.0` for `darwin/arm64` and `linux/amd64`. Missing, duplicate,
   malformed, version/role-shifted, or extra rows fail closed.
6. On an actual GitHub-hosted Intel macOS runner, the unchanged bootstrap
   selects exactly the new row, downloads only the pinned HTTPS resources,
   validates byte size and SHA-256, authenticates the detached signature to the
   existing locked Google signing identity, safely extracts the archive, and
   observes exact version `go1.26.7` and host `darwin/amd64` before `make ci`
   and the declared cross-build pass.
7. Historical digest manifests and governance records remain byte-for-byte
   unchanged. They continue to describe their bound historical candidates and
   are not rewritten to describe this remediation.
8. Both workflows, action digests, permissions, runner matrix, timeouts,
   baseline blocking status, shadow nonblocking status, benchmark policy,
   trusted-base evaluation, and main-branch protection remain unchanged. No
   failed check is skipped, allowed to fail, renamed, or removed.
9. Against the exact future audit successor head, required technical hosted
   checks `Go 1.26.7 (baseline)`, `CLI macOS 15 (arm64)`,
   `CLI macOS 15 (amd64)`, and `CLI paired benchmark gate` pass. The shadow job
   is recorded truthfully but remains nonblocking.
10. Trusted `evaluate` remains blocked unless external GitHub evidence supplies
    an exact-head accountable-owner approval and an exact-head independent
    auditor/reviewer approval whose actors satisfy Tier 3 separation. This code
    change neither supplies nor weakens those identities.
11. Repository-pinned targeted tests, clean-Git-config reproduction, shell
    syntax, foundation matrix controls, `make verify`, the complete applicable
    race suite, declared cross-builds, diff hygiene, ancestry, artifact budget,
    and tracked/index cleanliness pass before the sole verification record.
12. A separately authorized independent read-only `l7-release` audit maps every
    criterion, validates the authenticated toolchain boundary and test-only Git
    identity fix, and checks the complete rollback sequence before recording
    `GO` or `NO_GO`.
13. Base-to-implementation changes only this brief and the three declared
    implementation files. Verification and audit successors add only their own
    records, keeping the Tier 3 artifact total at three.
14. No provider executable, version/help interface, prompt/stdin, model session,
    retry, fallback, provider installation, compatibility promotion, global
    configuration, merge, release, deployment, or publication occurs or is
    claimed by implementation or offline verification.

## Test strategy and hosted-runner gates

Before implementation, retain the clean-config failing reproduction as
diagnostic evidence. After implementation, run the same focused test with
global and system Git configuration disabled and `user.useConfigOnly=true`, the
complete `./internal/l7/adapter/git` package, the repository-pinned
`make verify`, the applicable CGO-enabled race suite, and
`make cli-cross-build`. Run `sh -n` on the modified foundation check and the
unchanged bootstrap, then exercise the foundation check with the valid matrix
plus isolated missing, duplicate, malformed, shifted, and extra-row negative
fixtures. Confirm exact changed paths, immutable base/brief bytes, production
blob equality, historical manifest equality, `git diff --check`, artifact
budget, and clean tracked/index state.

Hosted-runner gates are separate from provider actual-host gates. Subject to
fresh external authorization, publish only the exact audited successor and
require:

- Ubuntu baseline Go `1.26.7`: blocking `PASS`;
- macOS 15 arm64 CLI: blocking `PASS`;
- macOS 15 Intel/amd64 CLI, including authenticated bootstrap: blocking `PASS`;
- paired CLI benchmark: blocking `PASS`, unless the existing accountable-owner
  exact-head regression acceptance mechanism is truthfully invoked;
- Ubuntu shadow Go `1.27.0`: record its result without changing its existing
  nonblocking status; and
- trusted `evaluate`: `PASS` only after exact-head technical checks and the
  distinct external owner/auditor review envelopes exist.

Provider actual-host execution is outside scope. Historical Gate 1/Gate 2 facts
remain candidate-bound, and both audited no-model parser gates remain `NOT_RUN`.

## Risks and mitigations

- **Ambient identity leaks into tests.** Limit deterministic identity to the one
  test-owned commit construction and prove the focused test under clean config;
  do not configure the repository or change production Git code.
- **Wrong or substituted Intel archive.** Bind official filename, length,
  SHA-256, HTTPS URLs, detached signature, existing primary/subkey fingerprints,
  safe archive layout, and extracted version before use.
- **Lock expansion weakens the matrix.** Require exactly the five named unique
  tuples and preserve every existing malformed-record and exact-match failure.
- **Historical manifests are mistaken for live inventories.** Leave all
  predecessor/candidate manifests immutable; use Git and this new Tier 3 lineage
  for current identity.
- **Technical success is mistaken for governance readiness.** Keep `evaluate`
  and branch protection unchanged; report the missing external identities as a
  separate blocker until real distinct actors act on the exact head.
- **The audited provider candidate is silently changed.** Work only from the
  clean baseline on a separate branch. Any future replay or requalification of
  PR #1 is a distinct transition with fresh exact-head verification and audit.
- **Supply-chain or network behavior expands.** Use only the existing bootstrap
  and official locked endpoints; add no dependency, action, secret, credential,
  retry class, or fallback.
- **The unrelated user artifact is staged.** Use exact pathspecs and leave it
  untouched and unstaged.

## Rollback

Rollback is history-preserving and restores exact base tree
`d57a334696487b1d15557c9980e8a55c2dc4c930`:

1. Before verification exists, revert the single implementation commit and
   then this brief commit.
2. After verification but before audit, revert the verification-record commit,
   implementation commit, and brief commit in that order.
3. After audit exists, revert the audit-record commit, verification record,
   implementation, and brief commits in that order.

Every sequence uses ordinary revert commits, rejects conflicts or extra paths,
and confirms the final tree equals the declared base. Historical manifests,
the audited provider branch/PR, public-repository state, branch protection,
reviews, and other remote evidence are outside this code rollback and remain
untouched. A downloaded cache may be discarded separately; it is ignored,
contains no source of authority, and is never required to reconstruct state.

If the clean-baseline remediation is later merged, PR #1 remains independently
candidate-bound. Rebasing, cherry-picking, merging the remediation into it, or
opening a successor provider PR changes its head and requires a separately
authorized, freshly verified and audited transition; this brief does not
authorize one.

## Commit sequence and approval boundary

1. `481adaaec967ac34b6b27cf78509b85d5c068abc` — exact clean baseline.
2. `docs(ci): define hosted CI readiness remediation` — add only this brief.
3. Stop for fresh accountable-owner implementation approval bound to the exact
   brief commit.

After fresh approval:

4. `test(ci): make hosted baseline checks hermetic` — modify only the three
   declared implementation files.
5. Run offline verification and freeze the exact implementation commit/tree.
6. `test(ci): record hosted CI readiness verification` — add the sole
   verification record.
7. Obtain separate authorization for one independent read-only `l7-release`
   audit bound to the verification successor.
8. `docs(ci): record hosted CI readiness audit` — add the sole audit record.
9. Stop. Remote publication, a new PR, hosted-runner execution, GitHub owner and
   auditor reviews, integration, and any update to PR #1 require their own
   explicit authorization and exact-head bindings.

The active user authorized only evidence inspection and creation and commitment
of this proposal. Implementation, approval-envelope creation, provider or
provider-interface execution, prompt/stdin, model sessions, retry, fallback,
verification record, audit, remote mutation, hosted CI, review, merge, release,
deployment, publication, installation, and global configuration remain
unauthorized. The next executable transition is fresh explicit approval for
implementation of this exact committed brief.
