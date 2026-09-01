---
name: l7-onboard
description: >
  Start Level 7 in an existing repository by inspecting project, Git,
  provider, policy, and memory state, then applying only explicitly approved
  orchestration setup.
user-invocable: true
---

# Level 7 Onboard

Use the local MCP tool `l7_v1_onboard` with `action: status` first. If the MCP
bridge is unavailable, run the plugin-relative executable as
`l7 onboard --status --json`.

Explain the returned ordered transitions and its executable `next` action.
Status is read-only. Run `action: apply` or `l7 onboard --apply --json` only
when the user explicitly asks to apply onboarding; that mutation creates or
updates tracked `.l7/orchestration.json` while preserving the strict legacy
`.l7/config.json` schema.

Never copy credentials into policy. Gateway credentials may only be referenced
by environment-variable name or macOS Keychain service/account. Continue with
provider probing and Sync only when the accepted state names those transitions.
