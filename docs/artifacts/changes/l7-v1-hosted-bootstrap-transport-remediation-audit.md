# Level 7 v1.0 Hosted Bootstrap Transport Remediation - Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-bootstrap-transport-remediation` |
| Candidate commit | `62c7f26a252b87eac8143c0756b080f69406853e` |
| Candidate tree | `1be2f125d0d8841ccc074893f7699f758d78c82d` |
| Result | `GO` |
| Reviewer | `apbusinessidentity-tech` |
| Audited at | `2026-09-01T17:55:45Z` |
| Implementation commit | `248bf52b15d381dea06c8d67d7f7c8505c53f504` |
| Implementation tree | `2f80134747575a92d1f5892cdcd61a5926142b80` |
| Approved brief | `e27676ddb4f8875cf9a88ff3c2ef2a26a85fdfa1` |
| Base commit | `66777352538a514b990ffca8fa290ca6de9584fd` |
| Base tree | `055583cef8181be59405443c2bb0ee14fc5e7690` |

## Decision and independence

`GO` for only the exact verification commit and tree above. The approved brief
and implementer verification were treated as claims, not conclusions. The
retry policy, cache-install boundary, authentication chain, offline regression
checker, workflow lineage rebinding, protected controls, historical failures,
policy state, rollback, and external-effect boundaries were independently
inspected.

Product Owner Anup Pandey (`addressanup`) commissioned this audit with reviewer
`apbusinessidentity-tech`, distinct from implementer/verifier `codex-root` and
PR author/release operator `anup19950725`. This repository decision is not a
GitHub review, hosted result, owner `GO`, merge-readiness determination,
benchmark-regression acceptance, or release authority.

## Requirement and evidence map

| Area | Independent assessment |
|---|---|
| Exact candidate | PASS - branch `codex/l7-v1-hosted-bootstrap-transport-remediation` began at clean HEAD `62c7f26a252b87eac8143c0756b080f69406853e`, tree `1be2f125d0d8841ccc074893f7699f758d78c82d`. The audit path and matching external audit envelope were absent. |
| Authority and topology | PASS - the strict local approval envelope binds Product Owner Anup Pandey, implementer `codex-root`, this change ID, and brief `e27676ddb4f8875cf9a88ff3c2ef2a26a85fdfa1`. Base `66777352538a514b990ffca8fa290ca6de9584fd` is followed by the sole-path brief, four-path implementation `248bf52b15d381dea06c8d67d7f7c8505c53f504`, and sole-path verification record in a direct linear chain. No prior approval, audit, review, or owner decision transfers. |
| Exact implementation scope | PASS - implementation modifies only `.github/workflows/release.yml`, `Makefile`, and `scripts/harness/bootstrap-go.sh`, and adds only `scripts/harness/check-bootstrap-go.sh`. There are no deletions, undeclared paths, or historical-record changes. Base through verification contains only those four paths plus the brief and verification record. `git diff --check` passes. |
| Historical failure lineage | PASS by read-only hosted inspection - Harness run `33526156162` failed at exact main head `66777352538a514b990ffca8fa290ca6de9584fd`. Arm64 jobs `99917396518` and `99923803648` both stopped in `make bootstrap-ci GO_VERSION=1.26.7` with curl status 56 and HTTP 500 before offline CLI checks. Attempt-one baseline, shadow, and amd64 jobs passed; its benchmark was skipped on main push. Trusted-policy runs `33527633971` and `33528069463` were skipped workflow-run records. None is credited as successor success. |
| Retry policy | PASS - each locked-file download has a literal four-attempt ceiling, a closed retry set of curl 28 or 56 and curl 22 only with final HTTP 408, 429, 500, 502, 503, or 504, fixed delays of 1, 2, and 4 seconds, and a fixed 600-second aggregate deadline passed to curl as the decreasing `--max-time`. Curl ambient configuration is disabled first. There is no nested curl retry, all-error retry, mirror fallback, workflow retry, environment override, swallowed transport failure, or success-on-timeout path. |
| Atomic cache and cleanup | PASS - every attempt uses a unique `mktemp` regular nonsymlink file adjacent to the final path. A successful closed transfer creates the absent final name atomically by same-directory hard link and removes the temporary link. Failure, exhaustion, and HUP/INT/TERM remove the temporary; an existing regular cache entry is preserved, symlinks and non-regular entries fail closed, and an output appearing during transfer is not overwritten. Authentication and extraction cannot continue after terminal transport failure. |
| Authentication and containment | PASS - locked HTTPS hosts and redirects, TLS 1.2 floor, archive size and SHA-256, exact Google primary and signing-subkey fingerprints, detached signature, archive-member paths, toolchain version, nonsymlink caches, disposable staging, and offline post-bootstrap execution remain effective. Toolchain and signing lock blobs are unchanged. Downloaded bytes cannot assert their own acceptance or change retry limits. |
| Focused independent diagnostic | PASS - from an exact Git archive of the verification head in disposable `/private/tmp` roots, `sh -n` passed for both changed scripts and the offline checker passed all 16 deterministic fixtures. The fixtures cover receive failure then success, exhaustion, timeout, all six allowed HTTP statuses, non-retryable HTTP/TLS/local-write failure, exact delays and calls, deadline, unique adjacent temporary files, atomic install, cleanup, pre-existing cache preservation, symlink refusal, signal status, and ambient policy isolation. `shellcheck` also passed. No external request occurred. |
| Makefile integration | PASS - `bootstrap-go-check` is the sole new phony target and runs the focused checker after existing cache preparation. It is added only to `technical-lint`; all prior targets, toolchain settings, fuzz inventories, benchmark commands, release commands, and direct failure propagation remain. |
| Release workflow | PASS - base-to-implementation workflow diff is exactly one replacement: `L7_RELEASE_BASE` changes from `e88b18ef1cbfd4f811efb1f0ab1b12a27a770503` to merged predecessor `66777352538a514b990ffca8fa290ca6de9584fd`. `actionlint` passes. Manual dispatch, exact two-parent/tree topology, actor, label, checks, distinct reviews, protected environments, signing/notarization, provider trials, owner authorization, absent tag/release, exact assets, attestations, and publish-once gates remain byte-identical. |
| Protected controls | PASS - Harness and trusted-policy workflows, paired benchmark checker, controller policy and tests, dependencies, toolchain/signing locks, target inventories, benchmark threshold and minimum, architectures, and production code are byte-identical to the base. No exception, threshold weakening, target reduction, policy bypass, or candidate-controlled benchmark acceptance is present. |
| Controller behavior | PASS - an offline pinned-toolchain run with disposable build/temp roots and the exact audit-request ref reported Tier 3 team state `awaiting-independent-audit`, candidate/tree above, six changed base-visible paths, and next action `record the bound independent decision`. Repository policy requires a strict non-versioned audit envelope whose actor matches this record and remains distinct from owner and implementer. |
| Implementer evidence boundary | PASS for the claims independently checked - record hashes for all four implementation files match. The record truthfully preserves earlier module-cache and disk-space failures and binds the final local `PASS` to implementation `248bf52b15d381dea06c8d67d7f7c8505c53f504`, tree `2f80134747575a92d1f5892cdcd61a5926142b80`. Its full verify, eight fuzz targets, package lifecycle, cross-build, and reproducibility evidence were not rerun by this audit. |
| User-owned state | PASS - the original checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged at SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

## Findings

No unresolved BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW finding remains within
the approved bootstrap-transport remediation scope and its local evidence
boundary.

The remediation preserves the two historical receive failures rather than
turning a rerun green by chance. Its retries remain resource-bounded and
status-selective, while every downloaded byte still crosses the independent
size, digest, signature, identity, member-path, and version gates.

## Residual risks and claim boundary

- No hosted workflow ran on this exact candidate. The fake-curl fixtures prove
  control flow, not that the live go.dev transport will recover on future
  GitHub macOS runners. Fresh exact-head baseline, shadow, arm64, amd64, paired
  benchmark, and trusted-policy results remain required.
- The aggregate deadline uses the runner's integer wall clock. Curl receives
  only the computed remaining time and timeout remains terminal, but abnormal
  host clock discontinuity is not simulated by the focused checker.
- The workflow base rebinding is locally exact but has not processed a future
  two-parent remediation merge, hosted reviews, environments, Apple secrets,
  provider evidence, or release assets.
- Actual Codex and Claude hosts/providers, model/network use, native amd64
  execution, signing, notarization, tag/release creation, publication,
  installation, and deployment remain `NOT_RUN` and unauthorized.
- A later benchmark regression remains blocking unless the existing separate
  exact-head owner decision is made. This audit accepts no exception and grants
  no owner, merge, signing, or release authority.

## Rollback and preservation

No remote or production effect was created. Before merge, rollback proceeds by
ordinary history-preserving reverts in reverse order: audit record,
verification `62c7f26a252b87eac8143c0756b080f69406853e`, implementation
`248bf52b15d381dea06c8d67d7f7c8505c53f504`, then brief
`e27676ddb4f8875cf9a88ff3c2ef2a26a85fdfa1`. Stop on any conflict or extra
path and confirm restoration of exact base tree
`055583cef8181be59405443c2bb0ee14fc5e7690`.

If a future remediation merge exists while v1.0.0 remains unpublished, revert
that merge only under separate authority. A failed release preparation must
publish nothing and leave unexpected external state untouched pending
separate cleanup authority. An immutable published release must never be moved
or overwritten; correction then requires a separately approved patch release.

## Next executable transition

After this sole-path audit successor and its matching non-versioned envelope
validate as `reviewed`, stop for separate Product Owner authority to publish
the exact audit head to a fresh Tier 3 PR. That future PR must retain the sole
trusted label `l7-risk-tier-3` and earn fresh exact-head Harness, paired
benchmark, trusted-policy, independent review, accountable-owner approval, and
owner `GO` evidence before merge.

This decision authorizes no push, PR mutation or review, hosted execution or
rerun, merge, release dispatch, provider trial, signing, notarization, tag,
publication, installation, or deployment.
