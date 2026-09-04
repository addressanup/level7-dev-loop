# Level 7 v1.0.0-dev Explicitly Unsigned Prerelease — Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-explicitly-unsigned-prerelease` |
| Candidate commit | `4e782916e6b0342559bbd787f752528ad72a53cd` |
| Candidate tree | `dc5bc73be09034bd0a6b71b1d1df5be0b0310ca6` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |
| Audited at | `2026-09-04T04:01:36Z` |
| Base commit | `5c23038e38a07b4f91f8ef38bbf163e061857910` |
| Base tree | `240bc9509a02a1a71616b62e88068ed7783bc65b` |
| Approved brief commit | `92d750c2d3be83321f65c3ed9c9f0ce4f9dc50e7` |
| Approved brief tree | `acd604b9be5fb6bb5bdadcf561fe95ca2d9494a5` |
| Verified implementation | `1a9af15ea59c7f9243e47e5c9b3504380fae456f` |
| Implementation tree | `a72a97c0135500c3980aab9d901a5c7171009e87` |
| Verification record SHA-256 | `e260aa42364301fd525c314b6aa492b15d14fbc387fada6cfd67a9cb178a0936` |
| Owner approval envelope SHA-256 | `7c12edd386c7423a58d0f8c7394c1516200358ecc8c2bf12a675b1568c3fcc3a` |

## Decision and independence

`GO` is granted only for candidate
`4e782916e6b0342559bbd787f752528ad72a53cd`, tree
`dc5bc73be09034bd0a6b71b1d1df5be0b0310ca6`. The approved brief, complete
ancestry, exact implementation scope, stable-lineage rebinding, unsigned
prerelease workflow, user-facing warnings, verification record, approval
envelope, rollback, and external-effect boundaries were independently
inspected read-only.

Reviewer `apbusinessidentity-tech` is a real GitHub collaborator with write
permission and is distinct from Product Owner `addressanup`, implementer and
verifier `codex-root`, and intended PR author and release operator
`anup19950725`. The active GitHub CLI identity remained `addressanup`; it was
not switched or used to mutate hosted state during this audit. This record is
not a GitHub review, owner approval, merge authorization, workflow-dispatch
authorization, environment approval, or publication authorization.

## Independently checked evidence

| Area | Assessment |
|---|---|
| Candidate identity and topology | PASS — the worktree was clean on branch `codex/l7-v1-explicitly-unsigned-prerelease` at the exact candidate and tree above. Base `5c23038e...` is an ancestor. The history is linear through brief `92d750c...`, implementation `71a4b7c...`, preserved failure `a27f654...`, lint remediation `f3eb24f...`, preserved wrapper failure `517474c...`, evidence-lineage repair `1a9af15...`, and verification successor `4e78291...`. No implementation commit follows verification. |
| Proposal and authorized scope | PASS — the proposal is the direct child of the exact base and adds only its brief. Relative to the approved proposal, implementation `1a9af15...` changes exactly `.github/workflows/release.yml`, `.github/workflows/unsigned-prerelease.yml`, `CHANGELOG.md`, `README.md`, and `docs/releases/v1.0.0-dev.md`. Candidate `4e78291...` adds only the canonical verification record after that tested implementation. |
| Approval and evidence lineage | PASS — the non-versioned schema-1 owner envelope exactly binds actor `Anup Pandey (addressanup)`, implementer `codex-root`, this change ID, brief `92d750c...`, and source `active-user-interaction`; its SHA-256 is exact. Historical failure records remain recoverable at commits `a27f654...` and `517474c...` with SHA-256 values `5a434102108b992cc591318748c6fa8d68b7b90ec2655ae28789ba26c8740eb0` and `ed9a176299944bd86db420065c4b13a0cf4da36d7b12719fff8047b36318c3a5`. The current tree contains one canonical PASS record and no audit record or audit envelope before this decision. |
| Stable release isolation | PASS — the complete base-to-candidate diff for `.github/workflows/release.yml` is exactly one line: `L7_RELEASE_BASE` changes from `c634e092b2f938ad3038be0632d162b2bdde41f3` to `5c23038e38a07b4f91f8ef38bbf163e061857910`. Every stable version, signing, notarization, provider, asset, production-environment, and publication control remains byte-identical to the base. The new constant correctly anticipates the required two-parent merge with current main as first parent and fails closed after later main drift. |
| Dispatch and control boundary | PASS by static inspection — the new workflow has only `workflow_dispatch`, fixed concurrency with cancellation disabled, bounded jobs, explicit least permissions, and five action uses pinned to full commit hashes. It requires repository, actor `anup19950725`, `main`, exact candidate commit/tree, current remote main, and run attempt one. It rejects duplicate exact-candidate dispatches, predecessor artifacts, existing `v1.0.0-dev` or `v1.0.0` refs/releases, drifted v0.1.0/v0.1.1 releases, and a missing or non-exact `v1-prerelease` environment before any environment job is eligible. |
| Hosted lineage gates | PASS by static inspection — preparation repeatedly binds the unique merged PR, exact two-parent merge ordering, PR-head tree, PR author, sole `l7-risk-tier-3` label, one-attempt PR Harness, successful baseline/shadow/arm64/amd64/paired-benchmark jobs, latest successful Trusted policy/evaluate result, distinct exact-head owner and auditor approvals, and one-attempt post-merge Harness. Publication revalidates those exact run and review IDs immediately before mutation. |
| Build and unsigned-signature boundary | PASS by static inspection — source is materialized by `git archive`; pinned Go `1.26.7` is bootstrapped once; the unchanged Makefile forces offline module/build controls after bootstrap. Two complete builds compare all four Mach-O inputs and both host ZIPs byte-for-byte, retain the exact arm64/amd64 inventory, and run the unchanged package check. `codesign` is used only for display; the workflow accepts only observed `AD_HOC` or `NONE`, rejects authority, Team ID, trusted timestamp, or Developer ID evidence, and contains no Apple credential name, signing command, keychain import, notarization command, fallback, architecture reduction, `continue-on-error`, or retry path. |
| Assets, manifest, and attestations | PASS by static inspection — preparation closes the asset inventory to the two `1.0.0-dev` host ZIPs, `SHA256SUMS`, and `UNSIGNED-PRERELEASE-MANIFEST.json`; rejects non-files, symlinks in extracted packages, oversize assets, and extra files; records exact sizes and SHA-256 hashes; and attests all four paths before one artifact upload with overwrite disabled, compression zero, and seven-day retention. The manifest fixes the unsigned, unnotarized, evaluation-only, provider-not-run, support-withheld state and contains no candidate-selected release authority. |
| Host and accountable-owner boundary | PASS by static inspection — the emitted schema binds a single raw merged-PR trial comment to the exact candidate, run, attempt, artifact, architecture, archive sizes and hashes, bounded command lists, host versions, transcript hashes, lifecycle results, other-architecture `NOT_RUN`, provider execution `NOT_RUN`, and support `WITHHELD`. Publication requires exactly one later, unedited `addressanup` owner comment byte-equal to the emitted asset-bound sentence and later than the trial comment, plus the protected environment approval. Repository bytes cannot synthesize either external decision. |
| Environment and credential boundary | PASS by static inspection — the non-environment control job can read the exact environment policy with Actions/read. The publish job receives only `RELEASE_ADMIN_TOKEN`, verifies it belongs to `addressanup`, and uses it only for administrative read checks: exact environment policy, sole secret name, immutable-release setting, main protection, and retained release state. Tag and release mutations use the job-scoped Contents/write token. No secret value is printed, copied into an artifact, or used for publication. Current GitHub documentation confirms the Actions/read environment-inspection permission, self-review prevention semantics, and Administration/read immutable-release inspection used by this design. |
| Publication and immutability | PASS by static inspection — after one final contiguous read-only revalidation, publication creates an annotated `v1.0.0-dev` tag object and ref, creates a draft prerelease against that tag, uploads exactly four files, validates remote size and digest, download-compares every byte, and publishes once with `draft:false`, `prerelease:true`, and `make_latest:"false"`. Final checks require `.immutable == true`, exact tag target/title/body/assets, unchanged main/environment/protection/retained releases, absent stable `v1.0.0`, and `v0.1.1` still returned as latest. Partial external state is left untouched for separately authorized recovery. |
| User-facing truth | PASS — release notes, README, and changelog consistently identify `v1.0.0-dev` as unsigned, unnotarized, evaluation-only, potentially blocked by Gatekeeper, provider execution `NOT_RUN`, and support `WITHHELD`; distinguish GitHub source archives from installable assets; prohibit global Gatekeeper disablement; preserve v0.1.1 as latest stable rollback; and keep future signed/notarized `v1.0.0` separate. |
| Hygiene and protected unrelated state | PASS — current implementation file SHA-256 values match the verification record. Remote main remained the exact base during audit; the change branch and both v1 tags were absent. The primary checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Evidence relied on without rerun

The following exact-candidate results were reviewed in the canonical
verification record and were not rerun:

- the sole pinned `make bootstrap-ci GO_VERSION=1.26.7` success;
- the preserved original workflow-lint failure, output-equivalent lint-only
  remediation, preserved evidence-wrapper failure, and one successful
  replacement `actionlint` invocation;
- the consumed successful `git diff --check` and semantic workflow review;
- the successful replacement
  `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7`, including module,
  import, formatting, vet, compilation, test, race, fuzz, reproducibility, and
  distribution coverage; and
- the sole successful `make v1-candidate-check GO_VERSION=1.26.7`, including
  arm64/amd64 reproduction and Codex/Claude package and native CLI/MCP
  conformance.

Those results bind tested implementation `1a9af15...`, tree `a72a97c...`.
Verification successor `4e78291...` adds only its evidence record, so no tested
implementation byte changed. This audit did not run `make`, bootstrap,
`actionlint`, a build, a test, a provider, or a workflow.

## Findings and residual risks

No unresolved implementation findings were identified.

| Severity | Count |
|---|---:|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |

The following remain future external gates, not facts established by this
local audit:

- exact-head hosted Harness, benchmark, Trusted policy, and owner/auditor
  reviews on the eventual PR;
- the separately authorized protected-control merge transaction, exact
  two-parent merge, restored main protection, and post-merge Harness;
- creation and verification of `v1-prerelease`, its sole encrypted secret,
  immutable-release enablement, and the token's live suitability;
- one-attempt hosted preparation, observed signature classifications, exact
  prepared asset hashes, manifest, artifact identity, and four attestations;
- exact-asset Codex and Claude host-only trials, removal and residue evidence,
  and the later asset-bound Product Owner sentence and deployment approval;
  and
- annotated tag creation and immutable public prerelease publication.

macOS amd64 native execution remains `NOT_RUN`; provider/model execution stays
`NOT_RUN`; Developer ID signing and Apple notarization stay `NOT_PERFORMED`;
formal support stays `WITHHELD`. The workflow's external trial comment is
schema-validated evidence supplied by the operator, not independently replayed
host telemetry; the later owner decision must evaluate that limitation.

## Preservation and rollback

This audit changes no implementation byte and preserves every historical
record. Before merge, rollback is an ordinary reverse-order revert of this
audit record, verification `4e78291...`, evidence-lineage repair `1a9af15...`,
wrapper-failure record `517474c...`, lint remediation `f3eb24f...`, original
failure record `a27f654...`, implementation `71a4b7c...`, and brief
`92d750c...`, without rewriting history; the resulting tree must equal base
tree `240bc9509a02a1a71616b62e88068ed7783bc65b`.

A failed preparation publishes nothing. Any unexpected tag, draft, release,
asset, artifact, or other hosted state must be reported and left untouched
pending separate cleanup authority. After immutable publication, never move,
delete, replace, overwrite, or promote `v1.0.0-dev`; correction requires a new,
distinctly versioned, fully reviewed prerelease. Future stable `v1.0.0` uses
new signed/notarized bytes and evidence. If remote main advances beyond the
exact prerelease merge first, a new stable-lineage remediation is required.

## Exact next transition

`codex-root` must validate that this committed audit is the direct child of
candidate `4e782916e6b0342559bbd787f752528ad72a53cd`, changes only this audit
path, and retains the exact candidate/tree and `GO` fields. Under separate
explicit authority, it may then create the non-versioned schema-1 audit
envelope binding actor `apbusinessidentity-tech`, candidate commit
`4e782916e6b0342559bbd787f752528ad72a53cd`, this resulting audit commit, and
source `independent-agent`.

No audit envelope, push, PR, hosted review, check, merge, branch-protection
change, environment or secret change, workflow dispatch, host trial, tag,
release, publication, installation, deployment, cleanup, or other external
effect is authorized by this decision.
