# Level 7 Dev Loop — Technology Selection Audit Record

| Field | Value |
|---|---|
| Artifact ID | `L7-AUD-TEC-001` |
| Artifact type | Separate-context technology, host-boundary, and supply-chain review record |
| Artifact schema | Bootstrap/pre-schema; migrate when the canonical artifact schema ships |
| Status | Complete — separate-context model audit `PASS`; owner technology decision pending |
| Version | 0.1.0 |
| Date | 2026-08-24 |
| Candidate | [`L7-TEC-001`](technology-selection.md) 0.2.0 |
| Candidate SHA-256 | `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` |
| Approved architecture | [`L7-ARC-001`](architecture.md) 0.2.0, SHA-256 `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` |
| Review mode | Three read-only specialist model reviews; no reviewer edited the candidate |
| Assurance label | Separate-context model review; not qualified human security/legal/domain review, AP2/AP3, actual-host proof, or release independence |
| Effect and risk | A1 audit-record write only; no source/runtime/host mutation |
| Owner decision | Foundation Step 4 approval remains pending |

## 1. Scope and method

The review tested Step 4 against approved requirements, backlog, architecture, the architecture approval/audit binding, current official host/tool documentation, exact local help/version observations, scored candidate arithmetic, P0 traceability, and the technology artifact's own fail-closed gates.

Three separate read-only reviewers contributed:

| Reviewer task | Review lens | Final-candidate role |
|---|---|---|
| `/root/technology_host` | Codex/Claude plugin, permission, sandbox, hook, MCP, distribution, lifecycle, and trusted-client realism | Established the stock-plugin A0 ceiling and companion requirement; diagnostic input, superseded by the adversarial final host review |
| `/root/technology_host_redteam` | One-shot argv/config, outbound request admission, relay/TLS, credential separation, process teardown, AP1, and model/service identity | Re-reviewed exact final digest; `GO` with no blocker/high |
| `/root/technology_consistency_audit` | Architecture/backlog consistency, protected release plane, predicate identities, TUF, rollout flag, retention/deletion, owner scope, and hidden operational cost | Re-reviewed exact final digest; `GO` with no blocker/high |

The primary agent performed all remediation. Reviewers had no mutation authority and did not close their own findings by editing the candidate. This is useful structural separation for a planning artifact only; it is not the independent release verdict required by `L7-BL-042`.

## 2. Review history

| Round | Candidate identity | Result | Disposition |
|---|---|---|---|
| 1 — host research | Initial pre-remediation draft; digest not retained for final binding | Stock-package A0 feasible; stock A1/A2 not proven; controlled companion recommended | Diagnostic only; triggered the explicit two-product decision and actual-host kill gates |
| 2 — consistency and host red-team | Intermediate drafts; exact digests not retained and not used for final binding | `NO_GO`; multiple blocker/high findings | Required remediation; nothing waived |
| 3 — targeted remediation loops | Intermediate candidates | Reviewers found remaining high inconsistencies in relay admission, argv/context, AP1 teardown, evidence issuers, TUF policy, deletion authority, verdict states, and owner scope | Each finding corrected and rechecked before final binding |
| 4 — exact final candidate | SHA-256 `d9ed43644d36a529c8cbc18806c2738d362c32ef16b0b3197a81c6e27d1dadc5` | Host red-team `GO`; consistency/security/data/supply-chain audit `GO` | Eligible for owner Step 4 decision; release readiness remains `NO_GO` |

The absence of retained intermediate digests is disclosed rather than reconstructed. Only the exact final candidate digest above is covered by the final verdict.

## 3. Material findings and corrections

| Finding | Initial severity | Candidate correction | Final state |
|---|---:|---|---|
| A normal Codex/Claude plugin cannot close every Level-7-issued A1/A2 path or mint the required approval receipt. | Blocker | Split the product into advisory A0 host packages and a separately installed, root-owned Controlled Client; plugin installation alone never enables mutation. | Corrected as an explicit product decision; actual closure remains unproved. |
| Long-lived host sessions and real credentials in a child made context, approval, and credential separation implausible. | Blocker/High | Select fresh one-shot host children inside Ubuntu containment, a parent-owned auth-injecting relay, and complete child teardown before AP1. | Corrected as a testable candidate; no support claim. |
| Codex has no documented empty-tools switch or pre-execution callback. | Blocker | Relay parses the outbound request before forwarding and rejects nonempty tools or admitted-context drift; failure to emit zero tools removes Controlled Codex. | Corrected with a hard kill result. |
| Relay checks covered routing but not semantic request equality, response identity, DNS/TLS, retry, compression, or cost bounds. | High | Strict bounded JSON admission, one request, exact requested model/context/tools, response identity validation before proposal admission, fixed upstream/TLS rules, and hard resource ceilings. | Corrected; SP-03/09 remain empirical gates. |
| Codex and Claude candidate invocations left optional context/tool/settings surfaces and result-envelope ambiguity. | High | Freeze exact flags, rebuilt environments, feature inventories, owned instruction/schema inputs, empty tools/MCP, version-qualified Claude `structured_output`, and unknown/default-on fail behavior. | Corrected; exact-version behavior remains unproved. |
| Process-group exit alone could leave descendants able to spoof confirmation. | High | Before AP1: close relay, reach stdout/stderr EOF, reap namespace init, prove dedicated cgroup empty, sanitize output, then render on the trusted terminal. | Corrected; SP-04 remains empirical. |
| Provider model names were treated too much like immutable implementation identity. | High | Separate requested-model admission from response-reported service identity; record observation interval and explicitly disclaim immutable weights unless verifiable. | Corrected. |
| Protected build/evaluation/release roles and signatures were collapsed or cryptographically ambiguous. | Blocker/High | Separate credential-free builders, provenance, evaluator launch/result, independent verdict, authorization, capability grant, artifact signing, and promotion; map every predicate to one permitted OIDC subject. | Corrected; concrete repositories/roles remain release blockers. |
| Release verdict omitted `CONDITIONAL_GO` and the complete frozen packet. | High | Bind the three verdicts (`GO`, `CONDITIONAL_GO`, `NO_GO`) to the full packet; invariant/gap failures force `NO_GO`; only fresh `GO` can authorize/promote. | Corrected. |
| Companion bootstrap/update/rollback and flag revocation lacked a complete trust lifecycle. | High | External Cosign bootstrap, narrow `l7up`, TUF 1.0.31 wire profile audited against 1.0.36, anti-rollback state, signed effect ceiling plus root policy, 7-day refresh and 30-day grant bounds. | Corrected; TUF conformance remains release-blocking. |
| TUF protected the stable delegation without explicitly constraining top-level `targets`. | High | Distinct 2-of-3 offline root, top-level-targets, and stable-targets roles; stable paths require the terminating delegation; bypass cases are mandatory. | Corrected; significant owner-approved custody burden remains. |
| Record history, secrets, legal hold, deletion, and tombstones were underspecified. | High | Versioned handling contracts, bounded retention, secret prohibition, `PENDING` tombstones, optional digest retention, truthful sink limits, and A4 handoff for irreversible/destructive purge. | Corrected; no unsupported erasure claim. |
| Unknown fields/future schemas could be dropped during cross-host writes. | Medium | Preserve/quarantine unknown core fields and future majors, round-trip namespaced extensions, and block critical unknowns. | Corrected. |
| P0 traceability, exact compatibility prerequisites, and rollout-flag lifecycle were incomplete. | Medium/High | Literal mapping for every P0 ID, exact host/model/platform/grant matrix, two-authority default-OFF ceiling, expiry/revocation/recovery/removal rules, and new spikes. | Corrected. |
| The owner gate hid material product, provider-key, platform, cloud, hardware-key, freshness, and Step 5 authorization costs. | High | Approval clause now enumerates the two-product UX, Ubuntu-only controlled target, per-session API key, GitHub Enterprise/AWS plane, three 2-of-3 offline key sets, refresh/renewal degradation, and inert Step 5-only scope. | Corrected. |

## 4. Residual conditions

No blocker or high technology-document finding remains. These deliberately unresolved feasibility conditions are not waivers or evidence of support:

- `AR-001`, `AR-002`, `AR-003`, and `AR-011` remain `UNPROVED`;
- Controlled Codex and Claude have no qualified model/service tuple, and either host is removed if its request/context/tool/relay contract fails;
- Ubuntu Bubblewrap, kernel, namespace, cgroup, path, transaction, recovery, and AP1 behavior require actual-platform proof;
- TUF 1.0.31-profile conformance against the 1.0.36 audit target, bootstrap, threshold custody, update, rollback, revocation, and removal remain unproved;
- GitHub Enterprise/AWS repositories, identities, protected environments, AMI, gateway, evaluators, and costs are not configured;
- all `SP-01`–`SP-16` results remain future evidence, not Step 4 claims; and
- licensing, human security/legal review, pilot evidence, and independent C7 release decision remain open.

Any failed hard gate causes the artifact's documented degraded result. Owner approval cannot convert `UNPROVED` into `PASS` or authorize a stable/release claim.

## 5. Final audit gate

| Check | State |
|---|---|
| Candidate comparisons, score arithmetic, and selected stack | `PASS` |
| Approved architecture and all P0 backlog IDs traced | `PASS` |
| Stock-A0/controlled-client product split and host realism | `PASS` as conditional technology design; support remains unproved |
| Outbound semantic admission, credential relay, containment, and AP1 design | `PASS` as a fail-closed experiment contract; SP-03/04/09 remain unproved |
| Records, paths, recovery, privacy, deletion, and unknown-field policy | `PASS` at technology-selection level |
| Protected build/evaluation/release identity and exact-byte chain | `PASS` at technology-selection level; infrastructure absent |
| Channel/TUF/feature-flag lifecycle and owner-visible operating burden | `PASS` at technology-selection level; conformance absent |
| Exact final candidate digest independently rechecked | `PASS` by both final reviewers |

**Final verdict:** `PASS` for the exact Step 4 conditional technology-selection artifact above. No blocker/high finding is waived. This is not implementation evidence or release readiness.

## 6. Owner gate

The product owner may approve or request revision of [`L7-TEC-001`](technology-selection.md). Approval accepts the material decisions and burdens in its §28 and authorizes Foundation Step 5 harness construction only, within that section's explicit exclusions.
