---
name: l7-experience
description: >
  Level 7 Principal Design Engineer loop for a working product: constitution,
  experience audit, experience system, improvement wave, and design QA. Trigger
  for UI, UX, interaction design, accessibility, IA, content design, or make it
  look and feel better.
user-invocable: true
---

# Experience Loop

This product already works. Diagnose before decorating.

## Constitution (new design session)

Cover the full stack, not just visuals: strategy, research, IA, content, interaction, UI, motion, accessibility, human factors, service design, conversion ethics, trust/privacy, design system, cross-platform, performance-as-UX, AI UX if present, and design QA.

Do not change pixels without a problem, evidence, and acceptance criteria.

## Sequence

1. Audit — landing, web, mobile, handoffs. Heuristics, IA, content, IxD, visual, a11y (WCAG 2.2 AA), mobile ergonomics, motion/performance, trust/conversion, system fidelity. Findings EXP-###, waves 1-4.
2. System — principles, voice, object language, tokens, required states, cross-surface rules. Extract from code first.
3. Wave implement — one approved wave. Pattern-level chunks. Preserve working critical paths.
4. Design QA — PASS / PASS WITH FIXES / FAIL / UNVERIFIED. Keyboard, contrast, states, fidelity.

After the system is approved, recommend `/l7-geometry` before or with Wave 1 if the user wants geometric/visual perfection.

Then `/l7-next`.

If a wave changes a critical journey, auth, payments, or IA, require the matching engineering release gate.
