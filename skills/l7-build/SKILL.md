---
name: l7-build
description: >
  Implement one approved feature or one wave using spec, design, small commits,
  tests, and documentation. Trigger for start wave, implement feature, build P0,
  or continue feature development.
user-invocable: true
---

# Build Loop

Implement one feature unless the user explicitly says `Start Wave N` and the orchestration plan exists.

## Preconditions

- Harness is green
- Feature exists in the backlog
- Dependencies for this feature/wave are done
- Current branch strategy is clear

Stop if any precondition fails.

## Per feature

1. Write spec. Wait for approval.
2. Write design. Wait for approval.
3. Implement chunks: schema → logic → API → UI → tests
4. After each chunk: lint, types, tests, show diff, wait if high risk
5. Security and performance self-review
6. Update README/API docs/changelog and `docs/artifacts/`
7. Commit with `feat(scope): ...`

## Wave mode

If the user said `Start Wave N`:

- Execute features in that wave in dependency order
- Pause after each feature unless the user said continue
- After the wave: merge, full test suite, wave report

## Post-launch features

If the product is already live, do not use this skill first. Use `/l7-change`.
If the user already implemented ad hoc, use `/l7-review`.

When the wave or feature is complete, tell the user to run `/l7-next`.
