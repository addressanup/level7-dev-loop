# Level 7 v1.0 Multi-Host Orchestration — Scope Successor Independent Audit

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration-successor` |
| Candidate commit | `f97b2445bb966660fb3cb403521cae78f0367cbe` |
| Candidate tree | `4aac837742726a63d8103db8da4a8fb5c6c32db8` |
| Result | `NO_GO` |
| Reviewer | `apbusinessidentity-tech` |

## Decision and independence

`NO_GO` for the exact verification successor above. Product Owner Anup Pandey
separately commissioned this audit and designated `apbusinessidentity-tech` as
the independent reviewer. The reviewer is distinct from implementer and
verification author `codex-root`; the implementation and verification
conclusions were treated only as claims to check, not as audit results.

This is a local, independently reasoned, read-only review of the verified Git
candidate. It is not a hosted forge review, organizational-independence
attestation, owner GO, or authorization for any push, signing, release,
publication, protected-branch merge, installation, or deployment. Any future
hosted team-assurance transition must independently bind the real exact-head
forge identities and reject an auditor who is the configured owner or PR
author.

## Exact candidate and evidence checked

| Area | Independent assessment |
|---|---|
| Git envelope | PASS — the audit began and ended its read-only phase on branch `codex/l7-v1-orchestration` at exact HEAD `f97b2445bb966660fb3cb403521cae78f0367cbe`, tree `4aac837742726a63d8103db8da4a8fb5c6c32db8`, with clean tracked, index, and untracked state. The candidate is unsigned local Git evidence; no signature or remote-readiness claim is inferred. |
| Approval boundary | PASS for the local implementation authority checked — the external approval envelope binds actor `Anup Pandey (Product Owner)`, implementer `codex-root`, change ID `l7-v1-multi-host-orchestration-successor`, and exact brief addition `493612797bb7a23bdc9e48c3d21685cda05534a6`. Historical approval of the predecessor brief was not reused. |
| Lineage | PASS — the successor proposal is the direct child of historical brief `a9038285d612dec6d1c496c3f9a69fed9ca75f74`; integration `a60006145ccf410813f971485c8eead561d7d01b` is the expected two-parent merge of exact target `0047641d6b825c3272ee6fd57aca93885e6aa75f` and successor `493612797bb7a23bdc9e48c3d21685cda05534a6`; the audited head is its direct verification-only child. |
| Scope and blob identity | PASS — base-to-integration contains the successor brief plus exactly 56 declared implementation paths. Every non-governance blob matches implementation commit `dc1fdb6a88d173a78e375bb8d54691c999d12e39`; integration-to-audit-target adds only the declared verification record. `git diff --check` passed. |
| Default and authority controls | PARTIAL — orchestration, Sync, Cyber active mode, and Headless are default OFF; active Cyber additionally requires an image digest, and Headless requires a digest-bound confirmation. The two security findings below prevent reliance on the gateway mutation and endpoint boundaries. |
| Implementation and tests | CHECKED — routing, native and gateway discovery/workers, configuration, broker, private memory, Cyber isolation, Headless checkpoints/worktrees/review/local-CAS merge, MCP/CLI surfaces, state stores, import boundaries, and their relevant tests were inspected. A targeted non-mutating comparison between broker validation and Git patch acceptance confirmed Finding `L7-ORCH-AUD-001`. |
| Verification evidence | PARTIAL — the implementer record truthfully binds `make verify` and `make v1-candidate-check` to integration commit `a60006145ccf410813f971485c8eead561d7d01b`, tree `3f47da61bc01d20ae60f08363d62fc39464ce60e`. Cached outputs independently match the recorded harness, CLI, Codex ZIP, and Claude ZIP SHA-256 values. Finding `L7-ORCH-AUD-003` identifies mandatory coverage not established by those gates. |
| Packaging and reproducibility | PASS for deterministic development artifacts only — both macOS architecture binaries and helpers are present in each host archive; permissions are explicit; telemetry and network default OFF; SBOM, checksums, and unsigned release-blocked provenance are present. Codex ZIP `efb0dbd1d6ed75d158860675d50332c69f30ed163999f35ebcbbaca1b3ae1eed` and Claude ZIP `58b3d6ad462abb2e25355b58a35621434144f6d89c4c270c54aad56faa8380fe` match the verification record. |
| CI and policy | NOT CREDITED as exact-head readiness — no remote branch contains this local audit target, and no hosted exact-head Harness, trusted-policy result, owner approval, or auditor approval was observed or claimed. Local policy evidence cannot authorize a remote transition. |
| Live behavior | NOT RUN — no live Codex, Claude, configured gateway, Docker/OrbStack Cyber image, real plugin installation, signing, publication, deployment, or production check was performed during this audit. Historical provider observations do not transfer. |

## Findings

| ID | Severity | Finding and required remediation |
|---|---|---|
| `L7-ORCH-AUD-001` | BLOCKER | The gateway tool broker does not derive and validate the complete affected-path set for every patch operation understood by Git. A valid multi-operation patch can therefore contain a path-changing operation absent from the broker's authorized path set, violating the promised fail-closed repository-scope boundary and permitting out-of-scope mutation. Remediation must use one canonical Git-aware, no-effect preflight to enumerate every affected source and destination path; reject unsupported, ambiguous, binary, copy, rename, mode, secret, protected, symlink, and out-of-scope operations; recheck the actual post-apply diff before accepting success; and add negative regression coverage for every path-bearing patch form. |
| `L7-ORCH-AUD-002` | HIGH | The configuration's plaintext loopback exception validates endpoint text by prefix rather than parsing and comparing the exact URL authority. It can therefore admit a non-loopback authority under the loopback-only HTTP policy, risking credential transmission over plaintext to an unintended host. Remediation must parse URLs, reject user information and ambiguous authorities, require HTTPS generally, allow HTTP only for an exact normalized loopback host/IP with a valid port, apply the same rule to catalog URLs, and add adversarial authority-normalization tests. |
| `L7-ORCH-AUD-003` | BLOCKER (evidence) | The successor keeps the original product, security, packaging, performance, and conformance criteria mandatory, but the recorded gates do not establish all of them. The standard verification recipe does not run a race suite or sustained fuzz campaign, while `v1-candidate-check` rebuilds and byte-compares development artifacts without exercising required v1 install, CLI, MCP, upgrade, and rollback conformance for both host packages. No hosted exact-head result fills that gap. Remediation must add and pass bounded, reproducible race/fuzz/adversarial and dual-host package lifecycle conformance on the remediated exact head, then record the exact commands, platforms, results, and artifact identities without promoting unperformed live-provider claims. |

The first two findings are independently sufficient to reject the security
boundary. The verification-coverage gap independently prevents a release or
owner-GO recommendation even after source remediation.

## Residual risk and boundaries

- Provider catalogs, aliases, sessions, quota recovery, gateway protocol
  compatibility, and real host authentication remain unqualified for this exact
  head because no live provider check was performed.
- The Cyber image digest is intentionally unset by default, and no signed image
  or real disposable-container run was checked. Active Cyber remains blocked.
- Headless multi-session crash recovery, natural quota waiting, provider
  failover, reviewer separation, and local merges have fixture/unit evidence,
  not a long-running real-provider qualification.
- The development packages are unsigned and release-blocked. Reproducible bytes,
  checksums, and an SBOM do not establish signing identity, notarization,
  provenance, host compatibility, support, installation safety, or release
  authority.
- No remote, protected branch, user installation, credential store, provider
  account, container runtime, production system, release, or deployment state
  changed during this audit.

## Rollback and preservation

No live or remote effect requires rollback. This audit record is deliberately
left untracked and unstaged. If the local candidate is abandoned after later
record commits, use ordinary revert commits in reverse dependency order: audit,
verification `f97b2445bb966660fb3cb403521cae78f0367cbe`, then integration merge
`a60006145ccf410813f971485c8eead561d7d01b` with its implementation-side first
parent retained, confirming restoration of exact target tree
`e314a0437a9ae0bfbd2e1812bb07c26b60ea385f`. Any deeper rollback must continue
through the preserved historical commits in reverse order, fail closed on a
conflict or unexpected path, and confirm the intended Git tree. Do not delete or
rewrite historical governance records.

If the candidate is remediated instead, preserve this `NO_GO` record and all
current commits. A source change invalidates verification and this audit for the
new candidate; it must proceed through a fresh approved scope, exact-head
verification, and a newly commissioned independent audit.

## Next executable transition

Do not request owner GO or perform any remote or release effect. Create one
concise Tier 3 remediation brief covering `L7-ORCH-AUD-001` through
`L7-ORCH-AUD-003`, obtain fresh Product Owner approval bound to that exact brief,
implement the bounded remediation, run the missing targeted and conformance
checks on the new exact head, create a fresh verification record, and commission
a different exact-head independent audit transition. Push, signing, release,
publication, protected-branch merge, installation, and deployment remain
separately unauthorized.
