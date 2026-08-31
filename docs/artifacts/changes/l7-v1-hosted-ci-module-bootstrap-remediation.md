# Level 7 v1 Hosted CI Module Bootstrap Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-hosted-ci-module-bootstrap-remediation` |
| Risk tier | `3` |
| Base commit | `f10486aee9b361f9b2ec7e92782f30247b7b7a32` |
| Base tree | `94a6f1b66a8831a29ef8c12e82cf679e4f5a1f42` |
| Product Owner | Anup Pandey (`addressanup`) |
| Implementer | `codex-root` |
| Trigger | PR #15, Harness run `33427651169` |

## Problem

PR #15 correctly binds the audited orchestration candidate to head
`f10486aee9b361f9b2ec7e92782f30247b7b7a32`, but its first hosted Harness run
fails before technical verification. Fresh Linux and macOS runners authenticate
the pinned Go archive and then enter `make ci` with an empty repository-local
module cache. The intended offline boundary (`GOPROXY=off`, `GOSUMDB=off`, and
`GOVCS=*:off`) therefore rejects the pinned
`github.com/odvcencio/gotreesitter@v0.24.0` lookup.

The benchmark job passed, but the baseline and both required macOS jobs failed
with the same missing-module error. Rerunning unchanged cannot repair an empty
cache. Relaxing the offline verification targets would weaken the supply-chain
boundary and is not acceptable.

## Scope

Add one explicit CI bootstrap phase that uses the authenticated pinned Go
toolchain to populate only the repository-local module cache from the fixed Go
module proxy and checksum database. The phase must verify the existing
`go.mod`/`go.sum` identities, disable direct VCS and credential use, and end
before the existing offline verification graph starts.

Add a network-free regression harness for the bootstrap command, environment,
failure behavior, and source-file immutability. Wire every hosted Harness job
that may compile candidate code through the same bootstrap entry point.

Out of scope are dependency or product changes, cache persistence between jobs,
secrets, private modules, proxy fallback, direct VCS, policy weakening, prior
record edits, PR mutation, hosted reruns, reviews, merge, signing, publication,
installation, release, and deployment.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation.md`
- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-hosted-ci-module-bootstrap-remediation-audit.md`
- `scripts/harness/bootstrap-modules.sh`
- `scripts/harness/check-bootstrap-modules.sh`

Modify:

- `.github/workflows/harness.yml`
- `Makefile`

Delete: none.

## Acceptance criteria

1. The proposal is a direct child of the exact base commit and adds only this
   brief. Product Owner approval must bind the resulting brief commit before
   any other declared path changes.
2. A dedicated CI bootstrap target authenticates the existing pinned Go
   toolchain first, then downloads the complete locked module graph into the
   repository-local module cache using only the fixed HTTPS Go proxy and the
   public Go checksum database. There is no `direct` fallback.
3. Module priming forces the local toolchain and workspace-off mode, disables
   telemetry, credentials, private-module exceptions, insecure transport, and
   all VCS access, and exposes no secret or ambient authentication input.
4. `go.mod` and `go.sum` are immutable inputs and remain outside this scope.
   Priming fails if either is missing or changes, if a required sum is absent or
   mismatched, or if download or checksum verification fails.
5. After priming, all existing build, lint, test, race, fuzz, reproducibility,
   distribution, package, CLI, MCP, lifecycle, and conformance targets retain
   `GOPROXY=off`, `GOSUMDB=off`, `GOVCS=*:off`, and `GOAUTH=off`. No technical
   verification target gains network access.
6. A network-free disposable regression harness uses a fake Go executable to
   prove the exact priming commands and environment, success behavior, failure
   propagation, dependency-file immutability, and rejection of unexpected
   inputs. Shell syntax and the regression harness are mandatory local gates.
7. Linux baseline/shadow, macOS arm64/amd64, and paired benchmark jobs use the
   same CI bootstrap entry point. Workflow permissions remain read-only,
   checkout credentials remain disabled, actions remain digest-pinned, and no
   secret, cache action, write permission, or `pull_request_target` execution is
   added to the Harness workflow.
8. Bounded local verification must exercise an empty disposable module cache,
   authenticate the locked module through the fixed endpoints, then run the
   unchanged offline gates. It must record commands, platform, outcomes, and
   exact candidate identity without crediting a hosted run that did not occur.
9. A later separately authorized push must preserve the exact verified blobs.
   Hosted success requires fresh exact-head baseline, macOS arm64, macOS amd64,
   benchmark, and trusted-policy checks; an edited workflow is not evidence of
   a hosted pass.
10. The existing local dry run against PR base
    `84bd69f90d366356b0ce1e1a392f906258f3de91` reports `GIT-003` because the
    active predecessor remediation declares a later base. This bootstrap-only
    brief does not alter or waive that separate trusted-policy lineage blocker.
    Any correction requires distinct exact scope and approval.
11. Existing orchestration behavior, default-OFF flags, package contents,
    development version, unsigned release-blocked provenance, frozen v0.1.1
    artifacts, prior briefs, verification records, audits, and external
    authority envelopes remain unchanged.
12. The original checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and outside this change.

## Risks and mitigations

- **Network access leaks into verification:** isolate network use in the named
  bootstrap target and assert that every downstream gate retains the existing
  offline environment.
- **Dependency substitution or fallback:** require existing sums, the public
  checksum database, a fixed HTTPS proxy without `direct`, the pinned local
  toolchain, VCS-off behavior, and unchanged dependency files.
- **Ambient credential exposure:** clear private/proxy bypass variables, keep
  `GOAUTH=off`, forbid workflow secrets, and test the effective environment
  with a fake executable.
- **Cache poisoning or stale reuse:** keep the cache repository-local, verify
  module content after priming, and exercise a fresh disposable cache during
  verification. Do not add a shared hosted cache action.
- **False promotion claim:** preserve failed run `33427651169` as historical
  evidence, require a fresh exact-head hosted run later, and keep the separate
  `GIT-003` lineage blocker fail-closed.
- **Stale assurance:** any implementation or workflow change invalidates the
  current verification, audit, and readiness result for promotion.

## Rollback

Before implementation, revert only the brief proposal commit to restore exact
base commit `f10486aee9b361f9b2ec7e92782f30247b7b7a32` and tree
`94a6f1b66a8831a29ef8c12e82cf679e4f5a1f42`.

After implementation, use ordinary additive reverts in reverse order: audit,
verification, implementation, then brief. Stop on any conflict or unexpected
path and confirm the intended Git tree after each step. Repository-local
test-owned caches may be discarded; historical records, remote PR #15, failed
run `33427651169`, user state, credentials, releases, and production state must
remain untouched.

## Current transition

1. Commit only this brief as a child of the exact audited base.
2. Stop for explicit Product Owner approval bound to that brief commit.
3. After approval, record the matching external approval envelope and implement
   only the declared paths using disposable roots and the fixed bootstrap
   endpoints.
4. If bounded local verification passes, commit only the declared verification
   record and stop for separate authority to commission an independent audit.
5. Push, PR mutation, hosted execution, review, owner GO, merge, signing,
   release, publication, installation, and deployment remain separately gated.
