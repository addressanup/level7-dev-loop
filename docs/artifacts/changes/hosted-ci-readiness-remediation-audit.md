# Hosted CI Readiness Remediation — Independent Audit

| Field | Value |
|---|---|
| Change ID | `hosted-ci-readiness-remediation` |
| Candidate commit | `9007ff0dd4bd516bf2ea68b370511db5ec7fa643` |
| Candidate tree | `5fd541a21f2e55df968342975bc03c8e54ac22f4` |
| Result | `GO` |
| Reviewer | `l7-release-independent-auditor` |
| Audited at | `2026-08-28T10:11:19Z` |
| Verified implementation | `228af7bbd29e1836ff8db59b2aab256af0b9fb9f` |
| Implementation tree | `04283aa24cfb010d539afad18c2a8e4ccf439fdf` |
| Brief commit | `55e97c93b65c96f4f281ea090816b3205784cae5` |
| Base commit | `481adaaec967ac34b6b27cf78509b85d5c068abc` |
| Base tree | `d57a334696487b1d15557c9980e8a55c2dc4c930` |

## Decision

`GO` for the exact offline remediation bytes and faithful materialization of
this sole audit record.

The approved brief, external approval envelope, implementation, verification
successor, object identities, direct ancestry, exact scope, artifact budget,
test-only Git identity change, five-tuple lock control, unchanged authenticated
bootstrap, historical-record preservation, provider-control boundary, disclosed
deprecated whole-tree-checker boundary, and state-complete rollback were
independently inspected.

This is a stage-bounded decision. It does not satisfy or waive the actual Intel
bootstrap or exact-audit-successor hosted checks. Those gates remain `NOT_RUN`
because the approved sequence deliberately places publication and hosted
execution after this audit under separate authorization. The missing distinct
GitHub accountable-owner and auditor/reviewer identities also remain a
fail-closed readiness blocker. This `GO` is not a merge, release, deployment,
publication, provider-compatibility, or hosted-readiness decision.

## Acceptance map

| Criterion | Independent assessment |
|---|---|
| 1. Exact base and preserved history | PASS — base `481adaaec967ac34b6b27cf78509b85d5c068abc` resolves to tree `d57a334696487b1d15557c9980e8a55c2dc4c930`. Brief `55e97c93…`, implementation `228af7bb…`, and verification `9007ff0d…` form a direct three-commit chain. The worktree is on separate branch `hosted-ci-readiness-remediation`. Its merge base with audited provider head `f0e9f54c053e9cc2ef93c98b05b9b07b42d5ffc3`, tree `82bace9a1bcb4fb4423badb4aed83dc1a91e0fbb`, is the exact declared base. The four files relevant to the hosted failures are byte-identical between that base and provider evidence head, so the remediation remains isolated from the audited provider lineage. |
| 2. Hermetic test identity | PASS by source inspection and bound verification evidence — only the test-owned `commit-tree` call gains command-local `user.useConfigOnly=true`, `user.name=Level Seven`, and `user.email=l7@example.invalid`. The verification record reports the focused test, full Git adapter package, and clean-runner `make ci` passing with author/committer environment removed and global/system Git configuration disabled. No repository, global, or system configuration write is introduced. |
| 3. Production Git behavior unchanged | PASS — production repository, commit, and merge adapter blobs are byte-identical to the base. The shared `runGit` and benchmark helpers are unchanged. The modification is confined to one test call, so production discovery, snapshot, identity, authorization, and merge behavior is unaffected. |
| 4. Exact Intel lock record | PASS — implementation adds exactly one tab-separated baseline `1.26.7` `darwin/amd64` row with filename `go1.26.7.darwin-amd64.tar.gz`, size `67852067`, SHA-256 `92e8b34bff3c89ab16404c595669ac8cb004cc2f676dcbd1f5b87a6b8def3b47`, archive URL `https://go.dev/dl/go1.26.7.darwin-amd64.tar.gz`, and detached-signature URL `https://go.dev/dl/go1.26.7.darwin-amd64.tar.gz.asc`. No existing row changes. Actual authentication of that archive remains criterion 6. |
| 5. Exact five-tuple control | PASS with the declared historical-checker boundary — the checker requires five non-comment records and exactly one of each approved role/version/OS/architecture key. Five total records plus exact multiplicity rejects missing, duplicate, shifted, or extra tuples, while its existing field, digest, and URL validation rejects malformed rows. Verification records successful valid, missing, duplicate, malformed, shifted, and extra-row fixtures in a disposable copy of original valid foundation commit `08c38b69a2cd63b4adf27873756a09e363e0c5a4`. Independent hash inspection confirms the deprecated current-tree checker stops earlier because `foundation-inputs.sha256` retains the historical Step-5 `AGENTS.md` digest; the same mismatch predates this change. The checker is not called by current `make verify` or hosted `make ci`. This disclosed limitation does not convert the historical checker into an active whole-tree gate or justify rewriting historical manifests. |
| 6. Actual Intel bootstrap | `NOT_RUN` — required future hosted gate. Static inspection confirms the unchanged bootstrap selects exactly one version/OS/architecture tuple, restricts downloads and redirects to HTTPS, checks archive size and SHA-256, pins the Google primary and signing fingerprints, requires a matching `gpgv` `VALIDSIG`, rejects unsafe archive members, and checks the extracted Go version. The unchanged workflow selects `macos-15-intel`, then runs bootstrap, `make ci`, and the declared cross-build. No Intel archive, signature, or Intel host was available to this offline audit, so no actual-host result is inferred. |
| 7. Historical records unchanged | PASS — base-to-implementation comparison shows no change to any historical `harness/*.sha256` or `docs/artifacts/*.sha256` manifest or prior brief, verification, or audit record. Their old digests remain historical candidate bindings rather than live inventories. |
| 8. Workflow and protection preservation | PASS for repository-controlled scope — both workflows, action digests, permissions, runner matrix, timeouts, baseline blocking status, shadow nonblocking status, benchmark logic, trusted-base evaluation, `Makefile`, and bootstrap are byte-identical to the base. No check is skipped, renamed, removed, or made nonblocking. Branch protection is external state and was not polled under the audit’s no-network boundary; this candidate contains no mechanism or authorized action that changes it. |
| 9. Exact audit-successor hosted checks | `NOT_RUN` — required future hosted gate. Ubuntu baseline, Ubuntu shadow, macOS arm64, macOS Intel/amd64, paired benchmark, and trusted `evaluate` are all unrun for this remediation lineage. The audit-only successor does not yet exist. The brief expressly requires these checks against that future exact successor and separately authorizes neither publication nor hosted execution at this stage. |
| 10. Trusted external authority boundary | PASS as a fail-closed boundary; not ready — repository policy remains unchanged and cannot derive owner or auditor authority from repository prose or technical success. Trusted `evaluate` must remain blocked until exact-head GitHub approvals supply a real accountable owner and a real independent auditor/reviewer distinct from the PR author and from each other. This local implementation and audit supply neither identity and make no `evaluate` success claim. |
| 11. Bound offline verification | PASS on the sole implementer record with independent corroboration — verification binds implementation `228af7bbd29e1836ff8db59b2aab256af0b9fb9f`, tree `04283aa24cfb010d539afad18c2a8e4ccf439fdf`, and reports the focused clean-config test, full adapter package, clean-runner `make ci`, repository-pinned `make verify`, applicable race suite, shell syntax, matrix fixtures, and Darwin arm64/amd64 cross-builds passing. Existing outputs independently hash to the recorded values: harness test binary `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da`, arm64 CLI `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb`, and amd64 CLI `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9`. The local receipt binds the previously authenticated arm64 Go `1.26.7` archive and existing signing fingerprints; it does not prove the new Intel row. |
| 12. Independent audit | PASS — this separately authorized review is bound to verification commit `9007ff0dd4bd516bf2ea68b370511db5ec7fa643`, tree `5fd541a21f2e55df968342975bc03c8e54ac22f4`. Auditor `l7-release-independent-auditor` is distinct from owner `accountable-owner` and implementer `codex-root`. The external approval envelope has schema `1`, exact change ID, exact brief commit `55e97c93b65c96f4f281ea090816b3205784cae5`, source `active-user-interaction`, and distinct owner/implementer fields. The approved brief blob remains unchanged. |
| 13. Scope and artifact budget | PASS — base-to-implementation adds only the approved brief and modifies only `harness/toolchains.lock.tsv`, `internal/l7/adapter/git/repository_test.go`, and `scripts/harness/check-foundation-scope.sh`. Implementation-to-verification adds only the sole verification record. The two current artifacts plus faithful addition of this sole audit record equal the Tier 3 maximum of three. Modes are preserved, and `git diff --check` reports no hygiene error. |
| 14. Claim and effect boundary | PASS — no provider executable, provider version/help surface, prompt/stdin, model session, retry, fallback, provider installation, compatibility promotion, global configuration, external CI, remote mutation, merge, release, deployment, or publication occurred or is claimed. Production provider code and tests are byte-identical to the base. Historical Gate 1 and Gate 2 evidence remains candidate-bound, and the audited no-model parser gates remain `NOT_RUN`. |

## Authenticated toolchain boundary

The implementation changes only lock data; it does not alter bootstrap code.

For a future Intel runner, the unchanged bootstrap:

1. derives `darwin/amd64` from the host;
2. requires exactly one matching version/OS/architecture record;
3. restricts archive and signature locations to the locked Go HTTPS endpoint;
4. verifies exact archive byte length and SHA-256;
5. loads the single locked Google signing identity;
6. requires both the pinned primary and signing fingerprints;
7. validates the detached signature and its signing/primary identity relationship;
8. rejects archive members outside the `go/` root or containing traversal;
9. verifies exact extracted Go version; and
10. emits a platform-bound receipt used for fail-closed cache reuse.

The existing `make toolchain-check` then checks version, GOROOT, tool directory,
target and host OS/architecture, local toolchain mode, offline dependency
configuration, CGO mode, and telemetry state. These controls are unchanged.
Their application to the new Intel archive remains an actual-host requirement,
not offline audit evidence.

## Deprecated whole-tree checker boundary

`scripts/harness/check-foundation-scope.sh` is preserved as deprecated Step-5
source and is not the active product policy or CI gate. Its historical
`harness/foundation-inputs.sha256` still matches the original foundation
candidate’s `AGENTS.md` SHA-256
`54496725a42eb7e6cce2a749e82a408d7277743ec8ad83c41373ceefbd4d0afa`;
the current base `AGENTS.md` hashes differently. Consequently, direct current-tree
execution fails before reaching the matrix block, as it already did at the
declared base.

The verification method—overlaying only the candidate checker and lock onto a
disposable archive of original valid foundation commit `08c38b69…`—is
appropriate for isolating the changed historical assertion. It does not claim
that the deprecated whole-tree checker is active or that current historical
manifests describe the new tree. The active bootstrap independently consumes
the lock and fails closed on missing or duplicate host tuples.

## Provider and hosted boundaries

Historical Codex Gate 1 remains `PASS` only for failed candidate
`8fba20512d1b5710104ec4b031ae9ee0f41d16a5`, tree
`7943f38db45705ce9cc1da01fb600f57e518215f`.

Historical Claude Gate 2 remains `NO_GO`: both exact implementer and reviewer
help invocations succeeded, both unknown-option parser controls unexpectedly
exited successfully, both invalid `--max-turns not-an-integer` controls failed
as required, and neither help surface advertised `--max-turns`. Help
advertisement remains non-dispositive; the successful unknown-option controls
remain dispositive.

The audited Codex and Claude no-model parser gates remain `NOT_RUN`. This
remediation does not weaken unknown-option rejection, typed `--max-turns`
enforcement, argument construction, permissions, output schemas, cancellation,
cleanup, reviewer immutability, scope, or containment.

The following remain blocking future transition gates:

- actual GitHub-hosted Intel bootstrap and technical execution;
- all required technical checks on the exact audit successor; and
- exact-head GitHub approvals from distinct real accountable-owner and
  independent auditor/reviewer actors.

They are expected sequencing gates, not defects in the inspected offline
candidate.

## Rollback proof

| State | Required reverse sequence | Independent result |
|---|---|---|
| Pre-verification implementation `228af7bb…` | Implementation, then brief | PASS — implementation is the direct child of brief `55e97c93…`, which is the direct child of base `481adaae…`; reverse-order ordinary reverts restore exact base tree `d57a334696487b1d15557c9980e8a55c2dc4c930`. |
| Post-verification `9007ff0d…` | Verification, implementation, then brief | PASS — the verification record is the direct and only successor after implementation. Reverting it first removes bound evidence before its candidate and restores the same exact base tree after the remaining two reversions. |
| Post-audit audit-only successor | Audit, verification, implementation, then brief | PASS — faithful materialization changes only this audit path. Reverting that record first leaves the already-proved direct chain and removes every change-specific artifact before restoring the exact base tree. |

Every rollback must preserve history with ordinary revert commits, fail closed
on conflicts or extra paths, and confirm the final tree equals
`d57a334696487b1d15557c9980e8a55c2dc4c930`. The audited provider branch and
PR, public-repository visibility, branch protection, reviews, hosted checks,
download caches, and other external state remain outside repository-tree
rollback and must not be represented as reversed by it.

## Findings and severity

No unresolved candidate findings.

| Severity | Count |
|---|---|
| BLOCKER | 0 |
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 0 |
| LOW | 0 |

Required future hosted and external-authority gates are recorded separately
above and remain blocking for readiness; they are not waived by this decision.

## Read-only and materialization boundary

The auditor used only read-only Git object, tree, ancestry, diff, source,
workflow, manifest, approval-envelope, cached-output hash, receipt, artifact,
and status inspection. It did not edit files, index, refs, envelopes,
configuration, remotes, or external systems. It did not run tests, builds,
`make`, the controller, provider executables, version/help probes, prompt/stdin,
model sessions, network, CI, retries, fallbacks, merge, release, deployment, or
publication.

The remediation worktree and index are clean. Status inspection of the primary
worktree reported only
`docs/artifacts/foundation-rebaseline-admission-audit.md` as untracked; its
contents were not inspected or touched.

This `GO` authorizes only faithful mechanical creation and commitment of this
exact audit record at
`docs/artifacts/changes/hosted-ci-readiness-remediation-audit.md` as the sole
repository-tree change after
`9007ff0dd4bd516bf2ea68b370511db5ec7fa643`, followed by creation of the matching
external audit envelope with:

- schema `1`;
- change ID `hosted-ci-readiness-remediation`;
- actor `l7-release-independent-auditor`;
- candidate commit `9007ff0dd4bd516bf2ea68b370511db5ec7fa643`;
- the resulting audit commit; and
- source `independent-agent`.

It authorizes no implementation, remediation, rollback, re-verification,
provider activity, remote publication, PR creation or update, hosted CI,
external review request, branch-protection or repository-setting change,
configuration change, merge, release, deployment, or publication. After
faithful materialization and envelope creation, stop.
