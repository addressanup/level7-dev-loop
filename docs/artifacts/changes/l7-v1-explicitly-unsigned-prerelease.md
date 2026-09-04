# Level 7 v1.0.0-dev Explicitly Unsigned (No Developer ID) Prerelease

| Field | Value |
|---|---|
| Change ID | `l7-v1-explicitly-unsigned-prerelease` |
| Risk tier | `3` — public unsigned executable distribution, protected release control, and future stable-release lineage |
| Status | `proposed`; implementation requires Product Owner approval of this exact brief commit |
| Base commit | `5c23038e38a07b4f91f8ef38bbf163e061857910` |
| Base tree | `240bc9509a02a1a71616b62e88068ed7783bc65b` |
| Product lineage | PR #18 merged by a two-parent commit whose parents are `c634e092b2f938ad3038be0632d162b2bdde41f3` and `f0f901e4a39228f73c91651d9b645ad4e5aa7531` |
| Passed lineage evidence | exact-head Harness `33608424877`; Trusted policy `33612345925`; merged-main Harness `33639534103`; merged-main policy run `33640801067` was the designed non-applicable push-event skip |
| Intended prerelease | annotated tag and immutable GitHub prerelease `v1.0.0-dev`; both tag and release are absent |
| Stable boundary | `v1.0.0` remains absent, signed/notarized-only, and is never created, moved, replaced, or weakened by this change |
| Proposal branch | `codex/l7-v1-explicitly-unsigned-prerelease`, rooted directly at the exact base |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root` |
| PR author / prerelease operator | `anup19950725`; must remain distinct from the accountable owner at hosted approval boundaries |
| Independent auditor | `apbusinessidentity-tech` |
| Hosted assurance | team mode; sole trusted PR label `l7-risk-tier-3` |
| Next executable transition | Commit only this brief, then stop for explicit Product Owner approval bound to the exact proposal commit |

## Problem

The merged v1 candidate is ready for its governed stable-release preparation,
but the Apple Developer enrollment needed to issue an exportable Developer ID
Application identity and a team notarization key is still being processed. The
existing release workflow correctly fails closed without those credentials: it
can publish only signed, notarized `v1.0.0` assets after two Accepted Apple
submissions. Disabling or bypassing that path would break the approved stable
contract.

The repository already produces deterministic `1.0.0-dev` /
`development-candidate` Codex and Claude archives containing the complete
macOS arm64 and amd64 inventory. Their embedded provenance deliberately says
`release_blocked:true`; candidate bytes cannot authorize their own promotion.
There is no hosted path that can expose those bytes as a clearly labeled,
evaluation-only GitHub prerelease under external Tier 3 controls. The current
README also says an unsigned development archive is not a release, and the
stable workflow's exact first-parent constant would become stale if a new
prerelease change were merged without rebinding it.

The smallest honest interim distribution is a separate, manual-only,
explicitly unsigned `v1.0.0-dev` prerelease. It must remain distinguishable
from stable in its version, channel, tag, release flags, title, notes,
manifest, assets, installation guidance, and retained lack of support claims.
It is not an updateable placeholder for `v1.0.0`: later Apple approval creates
new signed and notarized stable bytes with new hashes and evidence.

## Scope

After approval of this exact brief, implementation may:

1. add one manual-only, immutable-action-pinned unsigned-prerelease workflow
   fixed to version `1.0.0-dev`, channel `development-candidate`, tag
   `v1.0.0-dev`, and an exact merged-main candidate rooted at this brief's
   base;
2. preserve the existing packager, validators, product code, target inventory,
   and package-internal `release_blocked:true` provenance byte-for-byte; that
   field continues to deny candidate-controlled stable promotion while the new
   external workflow governs only an evaluation prerelease;
3. materialize source with `git archive`, bootstrap pinned Go `1.26.7`, build
   the unchanged distribution plus two clean unsigned input and package sets,
   compare all four Mach-O inputs and both host ZIPs byte-for-byte, and run the
   unchanged v1 package check without network, retry, or reduced inventory;
4. inspect all four Mach-O inputs with read-only signature tools and record
   whether each linker-produced state is `AD_HOC` or `NONE`; fail if any input
   carries a Developer ID authority, Team ID, trusted signing timestamp, or
   other identity signature. No `codesign --sign`, `codesign --force`,
   keychain import, or other signing mutation is permitted;
5. prepare and attest exactly four files:
   `level7-dev-loop-1.0.0-dev-codex.zip`,
   `level7-dev-loop-1.0.0-dev-claude.zip`, `SHA256SUMS`, and
   `UNSIGNED-PRERELEASE-MANIFEST.json`;
6. make the manifest bind the exact base, merge commit, tree, pull request,
   workflow repository/run/attempt, unsigned-input hashes,
   archive names/sizes/SHA-256 values, checksum and release-note hashes, exact
   channel/tag, and the explicit states `unsigned-prerelease-prepared`,
   `developer_id_signing:"NOT_PERFORMED"`, per-binary observed signature
   state, notarization `NOT_PERFORMED`,
   `actual_host_lifecycle_at_preparation:"NOT_RUN"`, provider execution
   `NOT_RUN`, support `WITHHELD`, and a requirement for external
   post-preparation host evidence before publication;
7. upload the prepared artifact once with retention fixed at seven days,
   compression level zero, and overwrite disabled; attest every file, bind the
   returned artifact ID in the handoff rather than the pre-upload manifest,
   and stop after emitting one exact asset-bound Product Owner authorization
   sentence;
8. under separate host-trial authority, download only that artifact and run
   one bounded Codex trial and one bounded Claude trial from disposable roots
   on one honestly recorded declared package architecture (`darwin/arm64` or
   `darwin/amd64`). Each trial verifies
   the archive hash, marketplace validation/add, installation, discovery,
   native CLI/MCP startup, uninstall, marketplace removal, and residue; records
   every exact command, host version, architecture, results, and a redacted
   transcript hash; leaves the other architecture `NOT_RUN`; and performs no
   provider/model execution. `anup19950725` must post exactly one raw
   workflow-schema trial JSON comment on the merged pull request;
9. allow the publication job to run only after the host-only trials pass, the
   exact emitted sentence is posted by `addressanup` on the merged pull
   request, and `addressanup` approves the pending deployment in a separately
   protected `v1-prerelease` environment with self-review prevented;
10. after final read-only revalidation, create the annotated tag object and
    then `refs/tags/v1.0.0-dev` pointing to that object; create a draft release
    only against that existing annotated tag; upload and download-compare the
    exact four files; and publish once with `draft:false`, `prerelease:true`,
    and `make_latest:"false"`;
11. use the exact release title
    `Level 7 Dev Loop v1.0.0-dev — UNSIGNED (NO DEVELOPER ID), NOT NOTARIZED`;
    make the remote release body byte-equal to `docs/releases/v1.0.0-dev.md`;
    and update README/changelog language to say prominently that the binaries
    have no Developer ID signature or Apple notarization,
    macOS Gatekeeper may block them, they are for evaluation only, actual
    provider/model qualification is not claimed, GitHub-generated source
    archives are not installable assets, and users must never be instructed to
    disable Gatekeeper globally; and
12. change only `L7_RELEASE_BASE` in the existing stable workflow from
   `c634e092b2f938ad3038be0632d162b2bdde41f3` to this exact base
   `5c23038e38a07b4f91f8ef38bbf163e061857910`. This preserves the future
   signed `v1.0.0` path for the exact two-parent merge produced by this change;
   every other stable signing, notarization, provider-trial, environment,
   owner-approval, asset, and publication gate remains byte-identical.

The workflow is inert until separately authorized repository administration
creates only `v1-prerelease` with protected branches enabled, custom branch
policies disabled, exactly one required reviewer `addressanup`, and
`prevent_self_review=true`; stores only `RELEASE_ADMIN_TOKEN` in that
environment; and enables immutable GitHub releases. That secret must be a
short-lived fine-grained token owned by `addressanup`, limited to this
repository, with only Administration/read, Actions/read, Contents/read, and
Environments/read.
The publish mutation uses its job-scoped `GITHUB_TOKEN` with Contents/write,
never the administration-read token.

A separately authorized administrator preflight must verify the exact
environment policy, the sole environment-secret name without its value, the
repository-wide immutable setting, and the unchanged v0.1.0/v0.1.1 releases
before dispatch. The workflow must begin with a non-environment control job
that uses its job-scoped token to prove `v1-prerelease` already exists with the
exact policy before any job referencing it can become eligible. That job also
requires `GITHUB_RUN_ATTEMPT=1`, proves the current run is the sole dispatch for
the exact candidate/tag, and finds no prior prepared run or artifact. The
environment-gated publish job reverifies every available control immediately
before any tag or release mutation. No Apple credential,
`v1-signing` configuration, `v1-production` approval, signing, notarization, or
stable release is needed or permitted for this prerelease.

The later merge requires separate protected-control authority because current
`main` requires linear history while both release workflows require an exact
two-parent merge. That transaction must snapshot the full normalized
protection, disable only `required_linear_history`, revalidate base/head/tree,
make exactly one GitHub `merge_method=merge` attempt with the expected head,
then restore `required_linear_history=true` after success or failure and prove
the full protection equals the snapshot. Restoration may be attempted at most
three times solely to recover the protection boundary; unverifiable
restoration is a critical stop.

Actual Codex/Claude provider execution remains `NOT_RUN`, and formal host or
provider support remains `WITHHELD`. Passing host-only lifecycle trials records
installability only for the one observed architecture; it is not provider/model
qualification and cannot promote the other architecture beyond offline
validation. No production installation or deployment follows publication.

Production code, package generators and validators, dependencies, the stable
release asset contract, Harness and policy workflows, benchmark comparator and
10% threshold, branch protection, historical governance records, v0.1.0 and
v0.1.1 tags/releases/assets, and provider or Apple credentials are outside
scope. There is no automatic trigger, `skip_signing` input, stable fallback,
architecture bypass, target reduction, rerun-until-green, overwrite, or
mutable-asset update path.

## Exact implementation file set

Declared path count: 8 (5 Add, 3 Modify, 0 Delete).

Add:

- `.github/workflows/unsigned-prerelease.yml`
- `docs/artifacts/changes/l7-v1-explicitly-unsigned-prerelease.md`
- `docs/artifacts/changes/l7-v1-explicitly-unsigned-prerelease-verification.md`
- `docs/artifacts/changes/l7-v1-explicitly-unsigned-prerelease-audit.md`
- `docs/releases/v1.0.0-dev.md`

Modify:

- `.github/workflows/release.yml`
- `CHANGELOG.md`
- `README.md`

Delete:

- None.

## Acceptance criteria

1. The proposal commit is the direct child of exact base `5c23038e...`, adds
   only this brief, and leaves every implementation byte, historical record,
   primary worktree, remote, pull request, review, check, protection, secret,
   environment, tag, release, asset, and deployment state unchanged.
2. Fresh Product Owner approval names this change ID, the exact proposal
   commit, base/tree, implementer `codex-root`, all five non-evidence paths,
   exact `v1.0.0-dev` identity, explicitly unsigned boundary, and sole stable
   workflow constant rebinding. No prior release authority transfers.
3. Implementation descends without history rewriting from the approved brief
   and changes only the five declared non-evidence paths. Verification and
   audit records are later single-purpose commits; historical records remain
   byte-identical.
4. The existing packager, candidate validator, Makefile, product code, tests,
   dependencies, target inventory, binary identity, and provenance logic
   remain unchanged. Package differences are limited to the declared README
   and changelog truth updates. Only exact `1.0.0-dev` /
   `development-candidate` packages are accepted. Stable version, channel,
   filename, manifest identity, tag, or release mutation is never used as a
   prerelease workflow input or output; documentation may reference
   `v1.0.0` only to explain the separate future stable boundary.
5. The stable workflow diff is exactly one replacement of `L7_RELEASE_BASE`
   from `c634e092...` to `5c23038e...`. It continues to require Apple signing,
   four verified hardened-runtime signatures with timestamps, two Accepted
   notarizations, provider trials, a protected production approval, exact
   stable assets, and one immutable `v1.0.0` publication.
6. The prerelease workflow has only `workflow_dispatch`, fixed concurrency
   without cancellation, bounded jobs, least permissions, and action uses
   pinned to full commit hashes. It accepts only actor `anup19950725`, exact
   inputs, and one attempt authorized separately. It has no Apple secret name,
   signing command, keychain import, `notarytool`, signing fallback, automatic
   trigger, retry, `continue-on-error`, or production environment. Read-only
   `codesign` inspection is allowed only to classify `AD_HOC` or `NONE` and
   reject any Developer ID identity, Team ID, or trusted signing timestamp.
7. Before any environment job is eligible, a non-environment control job
   verifies that `v1-prerelease` already exists with protected branches true,
   custom policies false, no wait timer or custom protection app, and exactly
   one required-reviewer rule for `addressanup` with self-review prevented. It
   requires run attempt one and proves this is the sole workflow dispatch for
   the exact candidate/tag with no prior prepared run or artifact. A failure
   cannot be retried or replaced by a fresh dispatch under the same candidate
   authority.
8. The later merge occurs only under separate exact authority that snapshots
   all normalized `main` protection, disables only required linear history,
   revalidates exact base/head/tree, makes one GitHub merge-method merge
   attempt with the expected head, and restores required linear history after
   success or failure. Full snapshot equality is mandatory; restoration gets
   at most three attempts solely to recover the protection boundary.
9. Immediately before the merge, preparation, and publication, remote `main`
   has the required exact value for that stage. The release candidate is the
   unique two-parent merge whose ordered parents are `[5c23038e..., exact
   reviewed PR head]`, whose tree equals that
   PR head's tree, and whose PR is closed/merged, authored by `anup19950725`,
   based on exact `5c23038e...`, and carries only `l7-risk-tier-3`.
10. The latest exact-head PR checks for baseline, Go shadow, macOS arm64, macOS
   amd64, paired benchmark, and Trusted policy all complete successfully. The
   latest owner and independent-auditor reviews are distinct, APPROVED, and
   bound to that exact PR head. A fresh automatic post-merge Harness must also
   pass before dispatch; no check, review, or predecessor evidence is inferred.
11. Two clean archived-source builds using pinned Go `1.26.7` reproduce all four
   unsigned binaries and both archives byte-for-byte. Separately, the clean
   exact implementation checkout passes the authorized local invocations of
   the unchanged full verification and v1 candidate check once each, without
   reduced architectures, targets, time budgets, or retries. Hosted Harness
   executions remain independent required evidence rather than retries of
   those local invocations.
12. The prepared artifact contains exactly the two `1.0.0-dev` ZIPs,
    `SHA256SUMS`, and `UNSIGNED-PRERELEASE-MANIFEST.json`, with no symlink,
    Apple ticket, credential, source archive, stable filename, extra file, or
    mutable reference. All four receive verifiable GitHub attestations.
13. The manifest and release notes state, without placeholder or ambiguity,
    that no Developer ID signing or Apple notarization was performed; record
    each binary's observed ad-hoc-or-none state; record actual-host lifecycle
    as `NOT_RUN` specifically at preparation time; and say that the assets are
    evaluation-only, potentially blocked by Gatekeeper, not provider-qualified,
    unsupported, and incapable of satisfying stable `v1.0.0` evidence or
    authority. The later run-bound pull-request comment is the sole
    post-preparation host evidence and must be revalidated before publication.
    The manifest and notes include no `release_ready`, signing-success,
    notarization-success, support-success, or candidate-selected approval
    field.
14. One separately authorized host-only trial attempt verifies the exact
    prepared Codex and Claude archives from disposable roots on one honestly
    named declared package architecture (`darwin/arm64` or `darwin/amd64`).
    Every command, archive hash, host
    version, operation result, architecture, residue result, and redacted
    transcript hash is recorded in exactly one raw JSON pull-request comment.
    Provider/model execution and the other architecture remain `NOT_RUN`;
    neither may be inferred. Any trial failure, residue, hash or architecture
    mismatch, or host ambiguity stops publication without retry.
15. `actionlint` passes both changed workflows, `git diff --check` passes, and
    semantic inspection proves the prerelease identity/asset/environment
    inventory and exact stable-lineage-only diff. The exact implementation
    candidate's authorized local verification invokes
    `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` and
    `make v1-candidate-check GO_VERSION=1.26.7` exactly once each.
16. One verification record binds PASS or FAIL to the exact implementation
    commit/tree and observed evidence. Only PASS may advance to a separately
    commissioned independent read-only audit by `apbusinessidentity-tech`,
    which binds GO or NO_GO to the exact verification successor. A separate
    exact-head owner review remains required before merge.
17. Repository administration changes only the `v1-prerelease` environment,
    its exact policy, sole `RELEASE_ADMIN_TOKEN` secret, and the repository-wide
    immutable-release setting under separate authority. The short-lived token
    is owned by `addressanup`, limited to this repository, has only
    Administration/read, Actions/read, Contents/read, and Environments/read,
    and is never used for publication. The administrator preflight verifies
    the exact environment policy and secret-name set without values, immutable
    setting, and unchanged
    v0.1.0/v0.1.1 refs/releases/assets before dispatch. The protected publish
    job reverifies the controls immediately before mutation and uses only its
    job-scoped Contents/write token to publish. No secret value appears in
    logs, artifacts, Git, comments, or chat.
18. The prepared artifact is uploaded once with seven-day retention,
    compression level zero, and overwrite false. Expiry, deletion, replacement,
    or missing attestation is terminal rather than rebuild authority.
    Preparation emits the exact future owner sentence binding candidate
    commit/tree, run/attempt, artifact ID, all four asset names/sizes/hashes,
    no-Developer-ID/unnotarized/unsupported state, tag `v1.0.0-dev`,
    `prerelease:true`, and `make_latest:"false"`; then stops. Generic approval,
    this brief approval, an environment click, or a package manifest cannot
    substitute for that later sentence.
19. Publication refuses any pre-existing `v1.0.0-dev` tag, release, draft, or
    asset and revalidates that `v1.0.0` remains absent and untouched. It first
    creates exactly one annotated tag object and its ref at the candidate, then
    creates a draft against that existing tag, uploads and download-compares
    exactly four assets, and publishes once. The final release title is exactly
    `Level 7 Dev Loop v1.0.0-dev — UNSIGNED (NO DEVELOPER ID), NOT NOTARIZED`;
    its body is byte-equal to the committed notes; and it is immutable,
    non-draft, prerelease, and created with `make_latest:"false"`. Before and
    after publication, GitHub's latest-release endpoint resolves exactly to
    retained stable tag `v0.1.1`.
20. Any identity, ref, tree, label, review, check, protection, environment, run,
    artifact, digest, attestation, flag, or asset drift; failure, cancellation,
    timeout, unexpected skip; or unexpected external state is terminal. Do not
    rerun, overwrite, move, delete, replace, or clean it up without new
    authority.
21. The stable workflow rebinding remains valid only while the exact
    prerelease-change merge is remote `main`. Any later main commit before
    Apple-approved stable dispatch invalidates that topology and requires a new
    reviewed stable-lineage remediation; no ref widening or evidence transfer
    is allowed.
22. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and unstaged at SHA-256
    `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`.

## Risks and mitigations

- **No-Developer-ID bytes are mistaken for stable:** isolate every identity
  and control, use the exact all-caps warning in the release title and notes,
  retain `release_blocked:true`, withhold support, keep v0.1.1 latest, and
  never create or touch `v1.0.0`.
- **Ad-hoc linker signatures are called unsigned without evidence:** allow
  only read-only signature inspection, record `AD_HOC` or `NONE` per binary,
  and reject every Developer ID authority, Team ID, or trusted timestamp.
- **Gatekeeper blocks or users bypass macOS safety:** disclose the expected
  limitation before download, provide no global Gatekeeper-disable command,
  and direct users who require normal trust behavior to wait for signed
  `v1.0.0`.
- **Candidate-controlled publication:** require exact Git/PR/check/review
  identity, a protected environment, a later digest-bound owner sentence, and
  a distinct release operator; repository bytes remain evidentiary only.
- **A missing environment is silently auto-created:** require the initial
  non-environment job and external administrator preflight to prove the exact
  protected `v1-prerelease` environment exists before any deployment job can
  become eligible.
- **Duplicate preparation changes the bytes or evidence:** require attempt one,
  exactly one dispatch for the candidate/tag, no predecessor artifact, fixed
  concurrency, and no rerun or replacement after failure.
- **Wrong or mutable assets:** compare two clean builds, use a closed four-file
  inventory, bind every byte by size and SHA-256, attest it, stage and
  download-compare it, enable immutability, and refuse existing state.
- **Prerelease strands stable lineage:** rebind only the stable workflow's exact
  first-parent constant to this base, then require the future prerelease-change
  merge to have that ordered first parent; any later main drift requires a new
  reviewed remediation instead of widening the check.
- **Protected linear history blocks the required topology:** use one separately
  authorized, snapshot-bound, single-control transaction and restore the full
  protection boundary after the sole merge attempt.
- **Host installability becomes a provider claim:** bind the one-architecture
  host-only trials to exact archives and record the other architecture plus
  provider/model execution as `NOT_RUN`; keep support `WITHHELD`.
- **The owner pause outlives its artifact:** use fixed seven-day retention and
  treat expiry, deletion, or missing attestation as terminal instead of
  rebuilding under stale authority.
- **Partial or failed publication:** publish only after complete staged
  verification; if any external state appears unexpectedly, stop and preserve
  it for separately authorized recovery rather than retrying or deleting it.
- **Credential or historical-state damage:** use only the scoped GitHub token,
  never access Apple credentials, preserve all prior records/releases, and
  isolate implementation from the unrelated primary checkout.

## Rollback

Before implementation, revert or discard only this proposal commit. Before
merge, revert later audit and verification records first, then the scoped
implementation and brief; confirm that the exact base tree and stable workflow
constant are restored. No remote history rewriting is permitted.

A failed preparation publishes nothing. Any unexpected tag, draft, release,
asset, artifact, or other external state is reported and left untouched until
separate cleanup authority is granted. After immutable publication, never move,
delete, replace, or overwrite `v1.0.0-dev`; disclose a defect and use a newly
reviewed, distinctly versioned prerelease if correction is necessary.

Apple approval later resumes a separate stable `v1.0.0` path. It produces new
signed/notarized packages, hashes, manifest, attestations, provider evidence,
and an asset-bound owner decision; it does not update the prerelease in place
or inherit its withheld support. Users move to stable through a clean
uninstall/reinstall under the later signed release instructions. If remote
`main` advances beyond the exact prerelease-change merge first, the stable
workflow must receive a new Tier 3 lineage remediation before dispatch.

## Current transition

Commit only this brief as a direct child of exact base `5c23038e...`, then stop
for explicit Product Owner approval of that exact proposal commit. That
approval must bind the change ID, base/tree, implementer, five non-evidence
paths, exact no-Developer-ID prerelease identity/title/assets, protected
`v1-prerelease` boundary, host-only trial boundary, protected-control merge
transaction, and one stable-lineage constant update. It grants no
implementation, verification, audit, push, PR, review, workflow dispatch,
environment or secret change, tag, release, publication, installation,
deployment, or cleanup authority.
