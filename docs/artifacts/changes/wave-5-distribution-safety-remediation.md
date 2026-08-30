# Wave 5 Distribution Safety Remediation — Solo Integration

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Risk tier | `3` — package identity and destructive fixture-lifecycle controls |
| Status | `approved` for this bounded repository-local implementation by the active user |
| Base commit | `17e7abfd32db6fa7d554d1942c5c734500c33e8a` |
| Base tree | `d2b8d90a76e53103fb40931c9acc09346fa66e97` |
| Assurance | `solo`; truthful self-review with no independent-audit claim |
| Runtime feature flag | Not applicable — build-time validation and disposable fixtures only |

## Problem

PR #7 contains a completed lifecycle-safety remediation for the development
plugin packages, but it was built before the solo-policy cutover and carries an
obsolete verification/audit commit chain. Its product changes have not reached
`main`.

The underlying gaps remain relevant: interrupted install completion can strand
valid state; removal needs a durable, resumable transaction; caller host/version
labels must bind to strict package metadata; inactive known digests must not be
relabelled; and compatibility claims must match the exact qualified development
tuples rather than merely being non-empty.

## Scope

Port only the final product-code state from implementation commits
`db4a20ba6c5c4b88d406dd20595b5667d8eecfcc`,
`e1acb5c9e3b69c11026c7935dc9174f72828ea76`, and
`6180a095d25eb2ecf55051b6d695fb2ac7b6f61e` onto the current solo mainline:

1. Bind lifecycle host and version inputs to exactly one strict
   `DISTRIBUTION.json` and the corresponding host manifest.
2. Make exact interrupted install completion idempotent while rejecting
   conflicting receipts and known-digest relabelling.
3. Journal removal before mutation, reject undeclared directories or files, and
   resume only a valid host-bound transaction while revalidating every remaining
   path.
4. Order directory durability barriers around journal, receipt, package,
   quarantine, and cleanup namespace mutations.
5. Validate the complete ordered Codex and Claude compatibility tuples.
6. Retain the current solo conductor descriptions and starter prompts when
   resolving distribution-generator conflicts.

No real host install/removal, provider execution, network access, publication,
release, deployment, repository-rule change, remote mutation, or support claim
is authorized. Provider execution and actual-host lifecycle remain `NOT_RUN`.

## Exact implementation file set

Add:

- `docs/artifacts/changes/wave-5-distribution-safety-remediation.md`

Modify:

- `internal/harness/distribution/lifecycle.go`
- `internal/harness/distribution/lifecycle_test.go`
- `internal/harness/distribution/main.go`
- `internal/harness/distribution/main_test.go`

Do not add verification or audit records. Git, exact-head checks, and truthful
self-review carry the evidence. All historical artifacts, workflows, skills,
manifests, authored distribution data, runtime/provider code, dependencies,
remotes, repository rules, and user-owned files remain unchanged.

## Acceptance criteria

1. The final code behavior from the three named implementation commits is
   present on top of the exact solo base without their evidence-only successors.
2. Recovery accepts only an exact already-committed install transaction,
   removes its stale pending record, and preserves the receipt and package.
3. Receipt completion cannot overwrite different version or ownership metadata
   for any active or inactive known digest.
4. Prepare-removal rejects undeclared filesystem entries without mutating owned
   or unowned bytes.
5. Removal persists and durably publishes a host-bound journal before deletion;
   mid-removal and post-receipt failures resume idempotently while every
   remaining path is revalidated.
6. Malformed, substituted, cross-host, symlinked, escaping, or conflicting
   lifecycle state fails closed.
7. Package validation rejects caller/metadata host or version mismatch and
   duplicate, missing, conflicting, or promoted distribution metadata.
8. Compatibility validation rejects every mutation of the exact Codex and
   Claude development tuples, including reordered or extra capabilities.
9. The base package bytes remain deterministic: Codex
   `02b9baddf6dbe43207aea7d85142ec16afa1ef1db771306f5b63ee4a6ffdf5d5`
   and Claude
   `da5f2f706b6793103069fbdaddef79b1e8b1dac404b60c88ba2f3a9ea5f64471`.
10. Current solo manifest descriptions and `l7-next` default prompts remain
    unchanged.
11. Focused distribution tests, race tests, full repository verification,
    declared cross-builds, exact-head solo readiness, and diff hygiene pass.
12. The implementation leaves no verification/audit artifacts and makes no
    unexecuted host, provider, publication, support, or release claim.

## Risks and mitigations

- **Destructive retry:** journals authorize only previously validated owned
  paths; retries revalidate all surviving content and never follow symlinks.
- **Durability ordering:** fault-injection tests assert every required namespace
  barrier and fail before later mutations when a barrier fails.
- **Port drift:** the final tree is compared with PR #7's implementation state
  for the four code files, except for the intentional solo generator metadata.
- **Claim drift:** package digests and compatibility states remain bound to the
  current development-only base and support stays withheld.

## Rollback

Before publication, discard this isolated branch. After any later authorized
integration, revert the implementation commits and then this brief as one
reviewed unit. No host, provider, migration, deployment, publication, or user
configuration state needs cleanup.
