---
name: l7-geometry
description: >
  Achieve geometric and visual perfection of UI components using a 4/8 spacing
  rhythm, shared control heights, nested radii, baseline type, and optical
  alignment. Trigger for pixel perfect, spacing, grid, alignment, or visual polish.
user-invocable: true
---

# Geometric Perfection

Measurement beats taste.

## Standard

- 4px subgrid / 8px rhythm
- Internal padding <= external group spacing
- Type line-heights on the baseline unit
- Shared heights for sibling controls; inputs match adjacent buttons
- Nested radius = outer radius - padding
- Icons optically centered in 16/20/24/32 graphic sizes
- Hit targets >= 24 CSS px, 44 on iOS-class controls
- Same geometry in empty/loading/error/disabled/selected
- Off-scale values are defects unless documented as 0.5-1px optical nudges

## Sequence

1. Inventory magic numbers → GEO-###
2. Lock tokens (ask approval)
3. Perfect primitives before screens
4. Compose screens on the grid
5. Fill the measurement table

Verdict must be PERFECT or PASS WITH LISTED EXCEPTIONS.

Write `docs/artifacts/geometric-perfection-report.md`.
Then `/l7-experience` Design QA or `/l7-next`.
