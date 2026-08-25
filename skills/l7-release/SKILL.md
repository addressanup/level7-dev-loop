---
name: l7-release
description: >
  Release validation: integration testing, independent read-only audit, and
  evidence-based remediation. Trigger before deploy, audit the codebase,
  GO/NO-GO, fix audit findings, or release candidate review.
user-invocable: true
---

# Release Validation

Choose one mode. Do not mix them in the same turn.

## Mode A — Integration testing

Run or design E2E, API, data integrity, performance, security, and regression checks.
Write `docs/artifacts/integration-testing-report.md`.
End with Ready / Not Ready.

## Mode B — Independent audit (default if user says audit)

You are a read-only Principal Engineer auditor.

MUST:
- Map the repo before concluding
- Cite files, commands, and tests
- Score BLOCKER / CRITICAL / HIGH / MEDIUM / LOW / INFO
- Decide GO / CONDITIONAL GO / NO-GO

MUST NOT:
- Edit code, config, git, or environments
- Invent vulnerabilities
- Approve because tests happen to pass

If a host safety filter blocks security wording, continue with correctness, data integrity, reliability, authorization gaps, and operational readiness. Do not drop the GO/NO-GO decision.

Write `docs/artifacts/principal-engineer-release-audit.md`.

## Mode C — Remediation

Only after Mode B.

- Confirm each `AUD-###` before changing code
- One finding per commit: `fix(audit-AUD-001): ...`
- Add regression proof
- Do not self-issue GO
- Write `docs/artifacts/release-audit-remediation.md`
- Tell the user to run Mode B in a fresh session

Ask which mode to run if the user did not specify.
After completion, recommend `/l7-deploy` only on GO or cleared CONDITIONAL GO.
