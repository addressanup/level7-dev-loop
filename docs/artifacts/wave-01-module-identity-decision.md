# Level 7 Dev Loop — Wave 1 Module Identity Decision

| Field | Value |
|---|---|
| Artifact ID | `L7-DEC-W01-MOD-001` |
| Artifact type | Root Go module identity decision |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **APPROVED FOR THE WAVE 1 LOCAL CANDIDATE** |
| Accountable owner | Anup Pandey |
| Selected module | `github.com/addressanup/level7-dev-loop` |
| Replaces | Provisional `continuallabs.ltd/level7-dev-loop` |
| Owner evidence | `USER_ASSERTED`: Anup Pandey stated that the project will live in the personal GitHub account `addressanup` |
| Effect | Local module/import identity update only; no GitHub, network, publication, compatibility, or support effect |

## Decision

Wave 1 will use `github.com/addressanup/level7-dev-loop` as the root Go module path. The implementation must change `go.mod` and the active core entry in `harness/modules.lock.tsv` together and must derive the harness link-time import path from the selected module rather than retain the provisional literal.

The separately privileged updater remains `reserved` at `cmd/l7up` with identity `UNSET`. This decision does not select or activate an updater module.

## Evidence and limits

The GitHub-account ownership and intended future repository location are current-session user assertions. They are sufficient for this local development candidate but are not external identity attestation, proof that a GitHub repository exists, permission to create or publish one, a remote binding, a compatibility commitment, or a release claim.

Before any publication or distribution claim, a later gate must authenticate the exact repository/organization ownership, remote, package source, release identities, and generated import/package lineage. If `addressanup`, the repository name, or the intended hosting location changes, the module decision and every bound candidate must be reassessed.
