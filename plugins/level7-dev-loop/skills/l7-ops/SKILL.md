---
name: l7-ops
description: >
  Operate a live product using SLOs, incidents, feedback, and prioritized fixes
  without mandatory report generation.
user-invocable: true
---

# Lean Operations

Inspect SLOs, alerts, incidents, bugs, product signals, and cost. Update existing
operational records only when the information will be used; do not create a
governance report by default.

Route routine fixes as Tier 1, product changes as Tier 2, and security,
destructive, migration, or release work as Tier 3. For a specific live-product
change, apply `l7-change` inside the active conductor without another approval
pause; for an incident, prioritize containment and rollback before process
documentation. Require explicit authority before a live destructive effect.
