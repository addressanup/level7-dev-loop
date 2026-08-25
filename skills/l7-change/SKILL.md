---
name: l7-change
description: >
  Post-launch feature loop: intake, architecture impact, implementation with
  feature flag, progressive release, and outcome review. Trigger for add a feature
  after launch, new request, rollout, or flag cleanup.
user-invocable: true
---

# Live Change Loop

This is not greenfield. Preserve contracts, data, and SLOs.
Run the next incomplete step only.

## Sequence

1. Intake — Feature Change Record. Decision: BUILD NOW / LATER / EXPERIMENT / DO NOT BUILD. Require evidence and RICE. Feature flag planned; production default OFF.
2. Impact — Architecture compatibility, validation tier 1/2/3, rollback, flag design.
3. Implement — Spec, design, small chunks, flag OFF/ON tests, docs.
4. Validation
   - Tier 1: targeted tests + CI
   - Tier 2: staging + integration/E2E
   - Tier 3: `/l7-release` audit
5. Progressive release — dark deploy, internal, 1-5%, 25%, 50%, 100%. Hard rollback = flag OFF first.
6. Outcome — compare metrics, keep/iterate/retire, remove flag and dead paths.

## Rules

- Do not code during intake or impact.
- Do not expose users before the validation tier passes.
- Write artifacts: `feature-change-record.md`, `architecture-impact-review.md`, implementation and release reports.

Then run `/l7-next`.
