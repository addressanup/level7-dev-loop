# README discoverability and onboarding rewrite

| Field | Value |
|---|---|
| Change ID | `readme-discoverability-rewrite` |
| Risk tier | Tier 2 — meaningful public onboarding and product-positioning change |
| Assurance | Solo implementation with automated checks and truthful self-review |
| Base | `origin/main` at `6757cc34d3bf9e05599387101bfc643e1fe4fee3` |
| Behavior change | None; documentation and one explanatory SVG only |

## Problem

The root README explains the repository in depth, but it makes a new visitor
read through internal CLI, harness, policy, and historical material before they
can quickly answer five basic questions: what Level 7 is, why it is useful, how
to install it, what it can access, and whether it fits their environment.

The rewrite must make the released instruction plugin understandable to
developers at any experience level, persuasive without hype, and easy for
search engines and answer systems to quote accurately.

## Scope

- Rewrite the root `README.md` around the stable v0.1.1 instruction plugin.
- Put a direct product definition, value proposition, and both host quickstarts
  near the top.
- Add clear workflow, audience, risk-tier, skill, permissions, compatibility,
  update, removal, troubleshooting, FAQ, contribution, and license sections.
- Add one accessible, repository-native visual with desktop and mobile SVG
  variants that explain the development loop.
- Preserve exact release commands and conservative compatibility boundaries.
- Use natural, descriptive language rather than keyword repetition or
  unsupported ranking claims.

Do not change skills, manifests, catalogs, the standalone CLI, workflows,
policy controls, dependencies, release assets, or remote repository settings.

## Acceptance criteria

1. The first two sentences identify Level 7 as a skills-only AI development
   workflow plugin for Codex and Claude Code and explain its primary outcome.
2. A new user can install and invoke the plugin without reading maintainer
   internals.
3. The README explains the one-intent loop, risk tiers, all 12 skills, effect
   boundaries, and the difference between the plugin and experimental Go CLI.
4. Claims about permissions, privacy, compatibility, host testing, and support
   remain no broader than the v0.1.1 release evidence.
5. Both SVG variants have a title and description; the responsive README image
   has meaningful alternative text, readable mobile text, sufficient contrast,
   no external requests, and no animation.
6. Links are descriptive and valid, installation snippets remain exact, and
   relevant distribution and repository checks pass.
7. The final diff contains only this brief, `README.md`, and the two SVG
   variants.

## Rollback

Revert the documentation commit. The plugin payload and runtime behavior are
unchanged, so rollback requires no migration or user action.
