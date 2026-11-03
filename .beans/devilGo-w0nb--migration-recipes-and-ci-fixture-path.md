---
# devilGo-w0nb
title: Migration recipes and CI fixture path
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-o8e7
---

# Migration docs and CI fixture path

## Goal

Close the lane set: a recipes page for `devilutionGo-migrate` (all three
lanes, flags, outputs, checksum summaries) and a documented CI path for
verification without MPQs or saves in the repository: synthetic fixtures run
in CI, retail verification recorded as an environment amendment, never a
silent skip.

## Research

- Existing environment-amendment convention in the asset verification PoW.
- docs/recipes layout conventions.

## Scope

- docs/recipes/devilutiongo-migrate; CI jobs on synthetic fixtures; skip
  documentation that names the required retail dirs.

## Acceptance criteria

- [ ] Recipes page covers all three lanes with copy-paste examples.
- [ ] CI runs the migration tests on synthetic fixtures and skips retail paths only with the documented amendment message.

## Libraries

- None.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Lanes 1-3 merged first; docs describe what shipped, not plans.

## Open questions

- None.

## Triage

TRIAGE brainer evidence 2026-11-03: laws mapped from the accepted
ADR-013/014/015/018 set onto the surveyed tree; code seams verified live
(file list in Research); workers follow existing given/when/then test law.
