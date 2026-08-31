---
name: l7-cyber
description: >
  Audit an authorized repository's attack surface and exploit hypotheses,
  optionally confirm them in approved disposable isolation, and prepare a
  separate remediation brief.
user-invocable: true
---

# Level 7 Cyber

Prefer local MCP `l7_v1_cyber`; fall back to `l7 cyber --json`. Always begin
with the read-only audit. Report trust boundaries, security findings,
classification, confidence, bounded reproduction, evidence digest,
remediation, and verification tests.

Use active confirmation only when the user explicitly requests it and the
tracked policy enables it. Invoke `active: true` or `l7 cyber --active --json`.
Fail closed unless the signed, digest-pinned image can run non-root in a
disposable Docker/OrbStack-compatible container without host credentials,
sockets, Internet, or the original checkout.

Auditing never edits source. When remediation is requested, use the exact
report ID with the remediation action. The resulting Tier 3 brief requires
fresh named owner approval before any patching. Export Markdown or redacted
JSON only when explicitly requested.
