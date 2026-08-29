# Wave 5 Distribution Safety Remediation — Independent Audit

| Field | Value |
|---|---|
| Change ID | `wave-5-distribution-safety-remediation` |
| Candidate commit | `6e1876884c16550c11a1ddbd22f46c7e9ec3cc5b` |
| Candidate tree | `35f335feb94d8d1249af9a9665629068de851937` |
| Result | `NO_GO` |
| Reviewer | `codex-release-auditor-1` |

## Decision and separation

`NO_GO`. A distinct read-only audit agent reviewed the exact verification
successor above. The auditor identity is different from accountable owner
`active-user` and implementer/verifier `codex-root`. The implementer/verifier
`codex-root` materialized the returned decision without changing its
classification or claiming that implementer self-review was independent.

This decision is limited to the approved repository-local remediation. It does
not authorize or perform merge, installation, publication, deployment, release,
provider execution, or any external mutation.

## Bound lineage and evidence

| Evidence | Audit result |
|---|---|
| Approval boundary | PASS — the external approval envelope binds owner `active-user`, implementer `codex-root`, and unchanged brief-only commit `555c00ac489cc009671219898e4c8964c76f4f75`. |
| Implementation candidate | PASS — implementation `e1acb5c9e3b69c11026c7935dc9174f72828ea76`, tree `58f6920b9dfbeb2c7491e6f2fcb57ab7a3ef324b`, is a linear successor of exact base `79a0739202942970536cc29782ad1b3952e7d15e`; its five changed paths are within the approved file set. |
| Verified audit target | PASS — candidate `6e1876884c16550c11a1ddbd22f46c7e9ec3cc5b`, tree `35f335feb94d8d1249af9a9665629068de851937`, adds only the permitted verification record above the implementation candidate. The controller reported `awaiting-independent-audit`. |
| Technical checks | PASS as technical evidence — focused, focused-race, full repository, and full-race suites passed; both Darwin cross-builds passed; distribution digests remained exact. The audit did not treat passing tests as proof of storage-crash durability. |
| Claim boundaries | PASS — actual provider execution, real-host lifecycle, hosted exact-head evidence, security certification, migration, benchmark, merge, release, and support remain withheld, unverified, `NOT_RUN`, or not applicable as recorded. |

## Findings

| ID | Severity | Finding | Required remediation |
|---|---|---|---|
| `W5R-AUD-001` | High | Lifecycle transactions are recoverable after injected process returns but are not proven crash-durable. `writeRegular` syncs temporary file contents and then renames the file without a durability barrier for the containing directory. Install/removal package-tree renames and later file, receipt, and journal deletions likewise lack explicit durable namespace ordering. A storage or power crash can therefore persist a destructive namespace mutation without the pending journal, or expose receipt/package state in a different order than the recovery protocol assumes. Recovery can return successfully when the journal rename was not durable. The tests inject ordinary function exits and do not establish filesystem persistence across a crash/reboot. Consequently, the verification record's statement that all five removal fault points are durable is unsupported. | Define and implement a platform-supported durable namespace protocol for journal publication and every dependent rename/deletion, with the journal durably visible before destructive mutation and receipt/package/journal ordering made explicit. Add fault/storage-order tests, then create fresh verification and obtain a fresh distinct audit. Narrowing the accepted crash-durability behavior would instead require a new owner-approved brief and is not remediation of this candidate. |

No other Critical, High, Medium, or Low finding was reported. Compatibility,
package identity, static ownership/containment and symlink checks, rollback
lineage, exact archive digests, and withheld actual-host/hosted claims were
otherwise consistent with the approved brief.

## Consequence and next boundary

This High finding blocks readiness, merge, release, installation, publication,
deployment, and support claims. A new implementation commit would invalidate
this verification and audit. Preserve this decision in history; before
remediation, reverse the audit-record successor and verification-record
successor in that order, then return to bounded `l7-build` work. Any remediated
candidate requires the complete verification and distinct audit sequence again.

## Rollback

No external state or migration exists. Reversing this audit-only successor
restores exact verification commit `6e1876884c16550c11a1ddbd22f46c7e9ec3cc5b`.
Reversing the verification successor then restores implementation candidate
`e1acb5c9e3b69c11026c7935dc9174f72828ea76`. Further reverse-order reverts of
the two implementation commits and brief restore exact base tree
`49400a5f97355b2ccbab27d165074d6be0a24757`. Use ordinary reviewed reverts and
fail closed on conflicts or unexpected paths.
