# Level 7 Dev Loop — Wave 1 Grant-Ladder Amendment Proposal

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-W01-GRANT-001` |
| Artifact type | Data-only technology/backlog successor proposal |
| Artifact schema | Bootstrap/pre-schema; a later approved schema transition is required before machine use |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **PROPOSED — INERT; NOT APPROVED OR ACTIVE** |
| Proposes to amend | `L7-TDR-001` `TDR-013`; `L7-BL-001` items `L7-BL-005`, `009`, `015`, `041`, and `042`; corresponding `L7-ORC-001` wave dependencies |
| Current authority/effect | Proposal data only (`AP0`, `A0`); grants no authority and permits no mutation |
| Required next gates | Independent security/boundary audit, digest-bound technology/backlog approval, implementation design, and separate implementation authority |

## 1. Decision proposed

Replace the single grant concept in `TDR-013` with four purpose-specific, non-interchangeable grant kinds:

1. `qualification` for Level-7-owned disposable synthetic roots;
2. `evaluation` for fresh evaluator-owned isolated case roots;
3. `pilot` for explicitly enrolled real-user roots after controlled conformance; and
4. `stable` for an exact independently approved release.

This proposal resolves the sequencing conflict in which a release-signed grant is required for all controlled mutation while C2 qualification, C5 protected evaluation, and C6 pilot mutation must occur before the C7 stable release decision. It does not weaken the stable release-capability design. Each earlier kind has a narrower issuer, audience, purpose, target class, evidence meaning, effect boundary, lifetime, and trust policy, and can never verify as a later kind.

Until a separate approval explicitly makes a successor normative, `TDR-013` remains unchanged, pre-release controlled mutation remains blocked, and this file is not an authorization input.

## 2. Exact affected assumptions

| Existing record | Affected assumption | Proposed successor treatment |
|---|---|---|
| `L7-TDR-001` `TDR-013` | Every controlled local mutation requires one release-signed capability grant. | Retain the release-signed rule for `stable`; introduce narrower pre-release grant kinds for qualification, evaluation, and pilot only. |
| Technology selection §11.3 | `controlled_local_mutation` consumes only `urn:level7:attestation:release-capability:v1`. | Verify one exact kind-specific predicate and trust policy; no predicate, signature, or evidence is substitutable across kinds. |
| `L7-ORC-001` Wave 5 | Controlled-boundary work depends on an approved grant-ladder resolution. | Permit only development of the verifier and negative fixtures; mutation still needs the matching approved kind and environment. |
| `L7-ORC-001` Wave 8 / `L7-BL-009` | C2 must mutate synthetic fixtures before a stable release exists. | Require `qualification`; evidence is development-only and cannot satisfy C5, C6, or C7. |
| `L7-ORC-001` Wave 11 / `L7-BL-015` | Protected mutation cases run under evaluator authority. | Require `evaluation`; candidate authors cannot request, select, inspect, issue, or score the grant or target. |
| `L7-ORC-001` Wave 12 / `L7-BL-041` | Consented pilot users require controlled local mutation before stable promotion. | Require `pilot`; bind the exact C5 candidate, cohort, root enrollment, consent protocol, environment, and expiry. |
| `L7-ORC-001` Wave 13 / `L7-BL-042` | Stable mutation follows an independent `GO` and release authorization. | Require `stable`; preserve the release-verdict, authorization, signing, TUF, root-policy, freshness, and revocation chain. |

No other requirement, backlog allocation, effect ceiling, approval-assurance rule, protected-evaluator boundary, release threshold, or updater separation is proposed to change.

## 3. Common signed envelope

A future implementation may accept a grant only as a DSSE envelope whose in-toto Statement subject is the exact controlled-client payload and whose predicate type is the one registered for its grant kind. The signed predicate must contain all of these fields; omission, ambiguity, an unknown field in a security-critical object, or an unsupported schema version fails closed:

| Field group | Required binding |
|---|---|
| Identity | Schema ID/version, exact grant kind, unique 256-bit random grant ID/nonce, issuer identity and workflow/key version |
| Audience and purpose | One named verifier audience and one kind-specific purpose; no wildcard or multiple-purpose value |
| Candidate | Exact source, package, controlled-client, adapter, and evidence-bundle digests; no tag, branch, mutable URL, or source-only substitution |
| Host/model/platform | Exact host and adapter versions, provider/model identifier and observed service identity, OS/architecture, image/kernel, containment profile, and policy/schema versions |
| Target | Canonical target class, root selector/identity, root-creation or enrollment evidence, owner class, and prohibited target classes |
| Effect | Maximum `A0`, `A1`, or `A2`; allowed action classes; explicit denial of `A3`, `A4`, `A5`, background, self-modifying, release, and publication effects |
| Time and replay | Issued-at, not-before, not-after, maximum lifetime, monotonic serial, revocation sequence, and trusted-time/freshness requirements |
| Policy and guardrails | Exact trust-policy, root-policy where applicable, risk/AP matrix, evaluator or cohort protocol, guardrail thresholds, failure behavior, and removal/expiry bindings |
| Lineage | Required predecessor grant/evidence, qualification or conformance result where applicable, issuer authorization, and superseded/revoked grant IDs |

The effective ceiling is always the minimum of the binary-declared maximum, achieved assurance, grant ceiling, containment proof, current environment, current action approval, and root-owned local policy where one is required. A local policy may narrow a grant and cannot widen it.

The action capability remains a separate unexported, in-memory, one-use capability bound to the exact action envelope, root handle, pre-state, executor/writer identity, deadline, and a fresh action nonce. A grant is not AP1 and cannot be presented as the user's action confirmation.

## 4. Kind-specific contracts

### 4.1 `qualification`

| Property | Contract |
|---|---|
| Predicate type | `urn:level7:attestation:qualification-capability:v1` |
| Purpose | Development proof of A1/A2 boundary behavior on disposable public synthetic fixtures before C5 |
| Issuer | Predicate-restricted qualification issuer outside candidate write authority; cannot issue evaluation, pilot, stable, verdict, authorization, signing, or promotion predicates |
| Audience | Exact qualification build of the controlled client and qualification supervisor only |
| Candidate prerequisite | Exact development candidate plus approved public qualification protocol; no release or support status |
| Target | One launcher-created, Level-7-owned disposable synthetic repository under the fixed qualification root class; arbitrary paths, existing user roots, mounted user data, home directories, and reusable workspaces are forbidden |
| Environment | Exact frozen Ubuntu 24.04 x86_64 image/kernel/Bubblewrap/provider/host/model tuple approved for qualification |
| Effect | At most A2 inside the admitted synthetic root; no external, Git remote, credential, package-manager, host-install, or publication effect |
| Lifetime | At most 24 hours and one qualification run/batch; the admitted action capabilities remain one-use |
| Evidence meaning | Development qualification only; never controlled evaluation, pilot, stable, compatibility, security, support, or release evidence |

The launcher, not the candidate or model, chooses and creates the root. The verifier must prove the canonical root is under the fixed qualification parent, newly created for the run, contains only approved synthetic inputs, is not a link/mount alias to user data, and is destroyed or quarantined after bounded evidence collection. Destruction is not represented as guaranteed storage, backup, log, or provider erasure.

### 4.2 `evaluation`

| Property | Contract |
|---|---|
| Predicate type | `urn:level7:attestation:evaluation-capability:v1` |
| Purpose | Candidate-inaccessible public/protected mutation cases for C5 controlled differential conformance |
| Issuer | Evaluator-controlled predicate-restricted issuer, separate from candidate, launcher, result, verdict, authorization, grant, signing, and promotion identities |
| Audience | Exact evaluator launcher/supervisor and evaluated payload digest |
| Candidate prerequisite | Exact C4 package bytes, frozen evaluator protocol, case-set commitment, environment, and result-issuer policy |
| Target | One fresh evaluator-owned isolated case root selected by the evaluator; the candidate cannot list, select, address, persist, or reuse hidden roots or case data |
| Environment | Exact frozen Ubuntu 24.04 x86_64 evaluation AMI/image/kernel/Bubblewrap/provider gateway/host/model tuple |
| Effect | At most A2 inside the current case root; no evaluator-control, threshold, oracle, credential, remote, promotion, or cross-case effect |
| Lifetime | At most 24 hours and one declared case batch; each case receives fresh root and action capability state |
| Evidence meaning | Evaluation input only; results require the separate result issuer and can never serve as pilot or stable authority |

Candidate bytes and candidate authors have no read, list, write, scoring, issuance, or revocation path to protected cases, grant material, truth labels, thresholds, evaluator credentials, or root-selection metadata. Any exposure invalidates the affected evaluation and triggers holdout rotation and a fresh candidate-independent run.

### 4.3 `pilot`

| Property | Contract |
|---|---|
| Predicate type | `urn:level7:attestation:pilot-capability:v1` |
| Purpose | Consented, staged C6 mutation by an exact C5-controlled candidate in explicitly enrolled real-user roots |
| Issuer | Pilot authority operating outside candidate write authority and separate from evaluator result, release verdict, stable authorization, artifact signing, and promotion identities |
| Audience | Exact pilot controlled-client package, pilot cohort, administrator-owned installation, and supported host/model/platform tuple |
| Candidate prerequisite | Exact C5 candidate and complete passing controlled-conformance packet; approved consent, privacy, incident, support, rollback, and observation protocol |
| Target | Explicitly enrolled canonical root identity and pseudonymous cohort selector in root-owned local policy; no wildcard, ancestor, arbitrary path, ambient workspace, or unenrolled user root |
| Environment | Exact supported pilot matrix entry; nonqualified or mismatched tuples remain A0 |
| Effect | At most A2 for approved local action classes after fresh AP1; no A3/A4 execution, background work, publication, release, or cross-root effect |
| Lifetime | At most 30 days, further bounded by cohort window, local-policy expiry, trusted freshness, guardrails, and revocation sequence |
| Evidence meaning | Pilot/adoption evidence only; cannot establish stable support, stable compatibility, release `GO`, or stable mutation authority |

An administrator enrolls the exact root outside the model session. The root-owned local policy records accountable owner, consent/protocol version, target identity, default OFF, enable/expiry, guardrails, failure behavior, privacy rationale, and removal work. Every mutation still requires current action-specific AP1. Pilot expiry or disablement blocks new actions but permits only digest-bound recovery of an already admitted transaction.

### 4.4 `stable`

| Property | Contract |
|---|---|
| Predicate type | `urn:level7:attestation:release-capability:v1` |
| Purpose | Post-C7 controlled A1/A2 use of an exact evidence-qualified stable release |
| Issuer | Existing predicate-restricted release capability-grant issuer after exact independent `GO` and separate release authorization |
| Audience | Exact released controlled client, adapters, host/model/platform tuple, TUF target, installation receipt, cohort, and root-owned policy |
| Candidate prerequisite | Exact C7 `GO`, release authorization, reproducible payload, protected/public results, pilot/adoption evidence, provenance, SBOM/license, signatures, rollback/revocation, and release packet |
| Target | Root-owned policy selects exact canonical roots or a narrowly defined pseudonymous cohort; repository/model/same-user data cannot enroll or widen a target |
| Environment | Only the evidence-qualified supported matrix; all other tuples remain A0 |
| Effect | At most authorized A2 after current AP1; A3/A4 remain plan/handoff only and A5 remains absent |
| Lifetime | At most 30 days with TUF timestamp freshness no older than 7 days, monotonic revocation, and local-policy expiry |
| Evidence meaning | Stable mutation authority only for the exact bound release and tuple; not a general support, compliance, or future-version grant |

All existing §11.3 stable rules remain: root-owned installation and anti-rollback state, fixed updater path, authenticated TUF metadata, exact receipt/grant/policy checks at session start and immediately before mutation, default OFF, bounded recovery, explainable OFF reason, and a separately approved removal/default-ON decision.

## 5. Non-interchangeability and verifier rules

1. Each kind has a distinct predicate type, issuer allowlist, audience, purpose enum, target-class enum, and evidence-state output. The verifier dispatches by the signed predicate type and never by a repository field, filename, CLI flag, environment variable, model claim, or requested effect.
2. Trust policies contain no common issuer authorized for more than one kind. A cryptographically valid signature from the wrong role is invalid evidence.
3. `qualification < evaluation < pilot < stable` is sequencing notation, not a widening inheritance chain. No lower kind verifies as, converts to, or supplies the authority of a higher kind.
4. A higher kind is also unusable in a lower environment unless its exact audience, purpose, target class, tuple, and policy explicitly match; there is no generic “at least this grant level” comparison.
5. Grant predicate, candidate, environment, target, local-policy, receipt, and action-envelope digests must all match current bytes. Any mismatch is OFF before AP1.
6. Repository data, unsigned JSON, self-signing, test mode, conversational approval, host session grant, hook output, model output, argv, ordinary environment, and same-user plugin/config state cannot issue, select, install, verify, or widen a grant.
7. Qualification and evaluation binaries/profiles are structurally compiled/configured without arbitrary-root admission. Their supervisors accept inherited canonical root handles from their protected launchers, not user/model-selected path strings.
8. Pilot and stable require root-owned installation, receipt, anti-rollback state, and local policy. Same-user-only or no-admin installations remain A0.
9. Failure is monotone: an unknown kind/schema/issuer, missing field, stale time, clock rollback, revoked serial, unreadable state, unsupported tuple, guardrail trip, or ambiguous root lowers to A0 and never retries by selecting another kind.
10. No grant authorizes evaluator modification, grant issuance, signing, promotion, deployment, publication, secrets, provider credentials, remote Git, package installation, or background/self-modifying work.

## 6. Expiry, revocation, compromise, and recovery

- Each issuer maintains a monotonic serial and revocation sequence in protected infrastructure. Pilot/stable updater state persists the highest observed values under root ownership so rollback cannot restore an older grant.
- Qualification/evaluation launchers keep equivalent run-scoped anti-replay state outside candidate control. Reusing a grant ID, case root, action nonce, expired batch, or superseded candidate fails.
- Revocation or expiry blocks new mutations. Only recovery already bound to a previously admitted transaction may proceed, and only to its declared safe state; it cannot admit new scope or a new candidate.
- Issuer compromise stops issuance, advances revocation, rotates the role-specific trust root through a separately approved process, invalidates affected evidence, and requires requalification/re-evaluation/re-enrollment as applicable. Compromise of one role does not authorize another role.
- Candidate or evaluator-data exposure invalidates the affected runs and grants. Protected cases and credentials are rotated; negative results and exposure scope remain recorded.
- Local disablement is immediate for new actions. Remote revocation is bounded by the declared trusted-time/freshness contract and must not be advertised as an instantaneous kill switch.
- Removal of pilot/stable software preserves the minimum root-owned anti-rollback/revocation tombstone required to prevent downgrade. Purging it is a distinct privileged destructive action with explicit preview and authority.

## 7. Migration and compatibility

1. This proposal is bootstrap Markdown only. No Wave 1 source reads it, and no filename, field, or status in it is an implemented interface.
2. After separate approval, the technology decision and backlog receive digest-bound successor records. Historical `TDR-013` remains immutable and is marked superseded only by reference, never edited in place.
3. Wave 5 must design canonical machine schemas, issuer policies, verifier state, typed OFF reasons, migration, and negative fixtures before any controlled mutation trial.
4. Existing development artifacts without a kind-specific signed grant do not migrate and remain A0. There is no grandfathering, conversion, or unsigned compatibility mode.
5. Changing a predicate schema, issuer, candidate, host/model/platform tuple, target class, effect, lifetime, revocation policy, evaluator boundary, or local-policy contract invalidates affected evidence and requires a newly authorized transition.
6. Stable-reader compatibility is fail-closed: an older binary that cannot interpret current trust or revocation state remains A0. Default-ON and removal remain later architecture/release decisions.

## 8. Required independent security/boundary audit inputs

The separate audit must receive and bind:

- this exact proposal digest and predecessor technology/backlog/orchestration digests;
- a clause-by-clause impact diff for `TDR-013`, §§11.3/12.3, and Waves 5/8/11/12/13;
- threat models for issuer compromise, wrong-role signatures, predicate confusion, downgrade, replay, trusted-time failure, root aliasing, mount/symlink/hardlink/TOCTOU, candidate/evaluator escape, credential/signing oracles, and grant laundering;
- exact future predicate schemas and trust-policy issuer/audience mappings;
- qualification/evaluation arbitrary-user-root denial proofs and launcher/supervisor interface design;
- pilot/stable root-policy, enrollment, receipt, updater, anti-rollback, expiry, revocation, recovery, and uninstall design;
- negative fixtures proving every lower-to-higher and cross-kind substitution fails for the intended rule;
- exact environment, containment, network, credential, host/model/provider, and target-root evidence contracts;
- separation-of-duty evidence for candidate, launcher, evaluator, result, verdict, authorization, issuer, signer, promoter, environment owner, and accountable owner;
- privacy/consent/retention/deletion analysis for protected cases and pilot roots; and
- rollback, compromise, key rotation, holdout rotation, incident, and evidence-invalidation procedures.

Any unresolved Blocker, Critical, High, or Medium finding leaves the amendment unapproved and every affected mutation ceiling at A0.

## 9. Compact R3 assurance case

| Element | Statement |
|---|---|
| Claim | Pre-release boundary work can obtain only the minimum kind-specific A1/A2 authority needed to produce qualification, evaluation, or pilot evidence without creating a path to stable authority or arbitrary user-root mutation. |
| Argument | Predicate, issuer, audience, purpose, candidate, environment, target class, effect, expiry, revocation, local policy, and evidence meaning are all kind-specific and jointly verified; there is no ordinal substitution or repository-selected mode. |
| Evidence required | Exact schemas/trust policies; issuer and root separation; arbitrary-root denial; wrong-kind/replay/downgrade fixtures; environment containment; protected evaluation; pilot enrollment/consent; stable release lineage; independent audit. |
| Defeaters | Shared issuer/key, generic predicate, wildcard audience/target, candidate-selected root, unsigned/test bypass, stale time, rollback, cross-kind acceptance, same-user pilot/stable install, evaluator exposure, or any unresolved material audit finding. |
| Residual risk | Root or protected-infrastructure compromise, provider/platform defects, delayed offline revocation, local same-machine evidence mutability, imperfect erasure, and future implementation error remain; each limits claims to the exact evidence and environment achieved. |

## 10. Approval and stopping boundary

This file is an inert proposal created as a Wave 1 deliverable. It does not issue, sign, install, verify, consume, simulate as active, or authorize any grant. It does not amend current technology or backlog records by its own presence, and the Wave 1 build-control command deliberately has no parser or activation path for it.

The only permissible next decision for this proposal is separate independent security/boundary review. Normative adoption, implementation, qualification, evaluation, pilot, stable use, publication, release, deployment, or exposure each requires its own later authority and evidence.
