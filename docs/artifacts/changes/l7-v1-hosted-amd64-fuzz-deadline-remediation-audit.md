# Level 7 v1.0 Hosted amd64 Fuzz Deadline Remediation - Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-amd64-fuzz-deadline-remediation` |
| Candidate commit | `ef34d34c419a048c57bfdbe61c752eef52bd45bd` |
| Candidate tree | `1fd495c015c6fb7b9ffa9e8679cc2a8f01d1c124` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |

## Decision and independence

`GO` for the exact hosted amd64 fuzz-deadline remediation verification commit
and tree above. The repair replaces wall-clock cancellation for only
`FuzzStrictConfigurationDecode` with a fixed execution count, bounds random
mutations below the production size ceiling, and adds deterministic
production-boundary coverage without changing production code or accepting a
timeout as success.

Product Owner Anup Pandey separately commissioned this Tier 3 team-assurance
audit after the forge reviewer login was resolved read-only as exactly
`apbusinessidentity-tech`. The reviewer is distinct from Product Owner
`addressanup`, implementer and verification author `codex-root`, and PR author
`anup19950725`. The approved brief and implementer verification were treated
as claims to test, not as audit conclusions.

This is an independently reasoned repository-read-only audit. It is not a
GitHub PR review, owner `GO`, hosted result, merge-readiness decision, or
authorization for any push, PR mutation, hosted run, protected merge, signing,
release, publication, installation, or deployment.

## Requirement and evidence map

| Area | Independent assessment |
|---|---|
| Exact audit candidate | PASS - the audit began and ended at clean branch `codex/l7-v1-orchestration` HEAD `ef34d34c419a048c57bfdbe61c752eef52bd45bd`, tree `1fd495c015c6fb7b9ffa9e8679cc2a8f01d1c124`. Its sole parent is integration commit `830cec9037e1852deee0a1ca5ebd391dbaa06c1f`, and it adds only `docs/artifacts/changes/l7-v1-hosted-amd64-fuzz-deadline-remediation-verification.md`. The declared audit path is absent. |
| Fresh proposal and authority | PASS - brief commit `f5691891874ca9b88e6090ec2854f45ed7d49e3f`, tree `55443dfae72e030f499908cb1eb53841b0ad485e`, has exact sole parent/base `84bd69f90d366356b0ce1e1a392f906258f3de91`, tree `84c2a105227f98089ea001f97473af79933bd743`, and adds only this successor brief. The strict external active-user approval envelope binds Product Owner Anup Pandey, implementer `codex-root`, the new change ID, and that exact brief commit. No predecessor approval, verification, audit, review, or owner decision is reused. |
| Integration topology | PASS - `830cec9037e1852deee0a1ca5ebd391dbaa06c1f`, tree `567b0920c49beef20d03b491fc4713360e0bd067`, is an additive two-parent merge. Its exact first parent is target `53bbe3b106f6bcea89078e474c2a63daf6c28d56`, tree `643de970a7fb68a66ccc1eed90871e34ea0a4358`; its exact second parent is approved proposal `f5691891874ca9b88e6090ec2854f45ed7d49e3f`. Both remain ancestors, with no rebase, reset, amend, squash, force update, or history rewrite. |
| Exact 69-path contract | PASS - the brief declares 48 additions, 18 modifications, and three live-tree retirements: 69 exact paths. Target-to-integration changes exactly six paths: the new brief, the three declared retired predecessor records, and only the two approved remediation files. Base-to-integration changes exactly 64 paths: the new brief plus 63 declared non-governance paths. The verification successor changes exactly 65 base-visible paths by adding only its record; this audit is the sole remaining declared path. `git diff --check` passed. |
| Implementation preservation | PASS - apart from `internal/l7/adapter/orchestrationconfig/config_test.go` and `scripts/harness/check-l7-fuzz.sh`, target-to-integration contains no non-governance change. All other 61 non-governance modes and blobs are therefore byte-identical to exact target `53bbe3b106f6bcea89078e474c2a63daf6c28d56`. Production code, dependencies, workflows, package inputs, default-OFF behavior, and product configuration semantics did not drift. |
| Random fuzz bound | PASS - the only fuzz callback change introduces literal test-only constant `fuzzConfigurationMaxBytes = 64 << 10` and skips random mutated payloads above that bound. Production `MaxBytes = 256 << 10`, `Load`, `localfile.Read`, strict JSON decoding, validation, schema, and admission behavior are untouched. The seed corpus remains valid and duplicate-field JSON. |
| Deterministic production boundary | PASS - `TestStrictLoadEnforcesProductionSizeBoundary` writes through a disposable test root, accepts valid JSON padded with legal trailing whitespace to exactly 256 KiB, rejects duplicate-field JSON at exactly 256 KiB, and rejects the valid exact-boundary payload after one additional byte. Inspection confirms `localfile.Read` accepts exactly the limit and rejects `limit+1`, while strict decoding rejects duplicate keys. |
| Fuzz driver contract | PASS - the literal inventory still contains all eight exact targets. Only `FuzzStrictConfigurationDecode` uses `10000x`; the other seven retain `5s`. Source assertions fail if either budget count changes. The pinned local toolchain, `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, `GOAUTH=off`, clean exact-head precondition, archived source, disposable cache/corpus/temp/telemetry roots, `CGO_ENABLED=1`, `-parallel=1`, and two-minute hard timeout remain. `set -eu` and direct command status propagation keep every fuzz or setup failure terminal; there is no retry, swallowed error, environment budget override, or architecture bypass. Shell syntax passed. |
| Independent focused diagnostics | PASS - from an exact `ef34d34c419a048c57bfdbe61c752eef52bd45bd` Git archive in disposable `/private/tmp` roots, the exact production-boundary unit test passed. A separate pinned Go 1.26.7 offline run of `FuzzStrictConfigurationDecode` with `CGO_ENABLED=1`, `-parallel=1`, `-fuzztime=10000x`, and `-timeout=2m` completed all 10,000 executions and passed. No repository path was written. |
| Implementer verification | PASS for the exact evidence checked - the verification record binds `PASS` to integration commit `830cec9037e1852deee0a1ca5ebd391dbaa06c1f`, tree `567b0920c49beef20d03b491fc4713360e0bd067`. It records 20 consecutive affected-target passes completing all 200,000 requested fuzz executions, final exact-head team `make verify`, and `make v1-candidate-check`. Heavyweight unchanged verification was not rerun for this audit. |
| Controller and policy integrity | PASS - `internal/harness/buildcontrol/policy.go` retains blob `3cde1e7b4bdd1ee80e4a6015f3e4c45e79da6c75`, and `.github/workflows/policy.yml` retains blob `e8afb29b310d5ab6270f667b45efcd4d22a4e604`, identically at PR base, integration target, and candidate. The recorded fresh-envelope Tier 3 team dry run against explicit base `84bd69f90d366356b0ce1e1a392f906258f3de91` passed with `changed=64` and no Git, scope, artifact, stale-authority, or policy-weakening finding. |
| Historical and artifact preservation | PASS - the three retired PR-base-lineage records exist at exact first-parent target `53bbe3b106f6bcea89078e474c2a63daf6c28d56`, are absent from the integration live tree, and remain reachable through immutable history. No retired record or external envelope was edited or extended. The base-visible verification candidate contains only the new brief and verification record under the three-record Tier 3 team budget; this audit is the third and final permitted artifact. |
| Reproducible identities | PASS for cached evidence - independent SHA-256 checks match the verification record for both changed files, the Harness workflow, Makefile, bootstrap scripts, reproducible harness and CLI binaries, both v1.0-dev host packages, and both frozen v0.1.1 packages. Package bytes remain unchanged, unsigned, and release-blocked. |
| Hosted and remote boundary | NOT RUN / PRESERVED - hosted execution of this exact head, including required macOS amd64, remains `NOT_RUN`. The local remote-tracking feature ref remains old target `53bbe3b106f6bcea89078e474c2a63daf6c28d56`. Failed Harness run `33481638065` and Trusted policy run `33482573666` remain historical failures; local results do not replace them. No push, PR mutation or review, hosted run or rerun, or remote success was inferred. |
| User-owned state | PASS - the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched and unstaged with SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Findings

No unresolved BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW finding remains within
the approved hosted amd64 fuzz-deadline remediation scope and its local
evidence boundary.

The remediation does not hide the historical hosted failure. It removes the
affected target's wall-clock success race while preserving terminal failure
semantics, a fixed minimum execution count, and deterministic production-limit
coverage. The independently repeated exact boundary and affected-target checks
support the implementer evidence.

## Residual risks and claim boundary

- Hosted macOS amd64 execution of the exact remediated head remains `NOT_RUN`.
  Runner-specific scheduling, process, or architecture behavior is therefore
  unqualified until a separately authorized fresh exact-head hosted run passes.
- The 64 KiB fuzz-only ceiling intentionally reduces random mutation coverage
  between 64 KiB and the 256 KiB production limit. Exact-limit valid, strict
  invalid, and oversized admission cases are deterministic, but they do not
  provide randomized coverage throughout that upper interval.
- A fixed `10000x` count removes wall-clock cancellation from the affected
  target but does not guarantee a fixed duration. The retained two-minute hard
  timeout correctly fails closed if individual executions or the corpus stall.
- The current hosted `AUTH-001` review-envelope failure is not supplied by this
  local audit record. A future GitHub review must bind the real distinct forge
  reviewer to the later exact pushed head.
- Darwin amd64 native execution, live provider and gateway calls, real host
  installation, signing, notarization, publication, release, and deployment
  remain `NOT_RUN`. The unchanged packages remain unsigned and release-blocked.
- No repository file, index, ref, envelope, remote, PR, workflow, host, release,
  or production state changed during this review. These reviewer-authored bytes
  have not been written, staged, or committed by the reviewer.

## Rollback and preservation

No live or remote effect requires rollback. If the later authorized audit-only
commit is abandoned, ordinarily revert only that audit commit. Deeper rollback
must proceed in reverse order: verification
`ef34d34c419a048c57bfdbe61c752eef52bd45bd`, then integration merge
`830cec9037e1852deee0a1ca5ebd391dbaa06c1f` using target
`53bbe3b106f6bcea89078e474c2a63daf6c28d56` as mainline. Confirm restoration
of exact target tree `643de970a7fb68a66ccc1eed90871e34ea0a4358` and stop on
any conflict or unexpected path. Preserve the base-rooted proposal, all
retired records and commits, external envelopes, PR #15, failed workflow runs,
packages, credentials, user installations, releases, and production state.

## Next executable transition

Return these exact reviewer-authored bytes to `codex-root` for sole-record
validation. Under the Product Owner's existing bounded authorization,
`codex-root` may materialize this record byte-for-byte, record the matching
external audit envelope, commit only this audit record, and then stop.
Byte-for-byte transport does not transfer or dilute reviewer authorship. This
audit does not authorize owner `GO`, push, PR mutation or review, hosted
execution or rerun, protected-branch merge, signing, release, publication,
installation, or deployment.
