---
# devilGo-w0nb
title: Migration recipes and CI fixture path
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-13T00:33:35Z
parent: devilGo-o8e7
---

---
# devilGo-w0nb
title: Migration recipes and CI fixture path
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T14:23:09Z
parent: devilGo-o8e7
---

# Migration recipes and CI fixture path

## Goal

Close the lane set: a recipes page for `devilutionGo-migrate` (all three
subcommands, flags, outputs, checksum summaries) and a documented CI path for
verification without MPQs or saves in the repository: synthetic fixtures run
in CI, retail verification recorded as an environment amendment, never a
silent skip.

## Research

- Existing environment-amendment convention in the asset verification PoW.
- docs/recipes layout conventions.
- Zero-dependency precedent already in tree: cmd/gamedatagoldens reads only
  committed TSVs and never touches an MPQ.

## Scope

- docs/recipes/devilutiongo-migrate page; CI jobs on synthetic fixtures;
  skip documentation that names the required retail dirs.

## Acceptance criteria

- [ ] Recipes page covers all three subcommands with copy-paste examples.
- [ ] CI runs the migration tests on synthetic fixtures and skips retail paths only with the documented amendment message.

## Libraries

- None.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- The three lane children merged first; docs describe what shipped, not plans.

## Open questions

- None.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Save story (ADR-003 + triage TD-4): SQLite rows (SaveHero/SaveGame/
   SaveLevel) are THE session save; the tick log is the replay/desync oracle
   validated against rows (upstream pfile.cpp:657/:665 model) — never a
   competing loader.
2. Data access (repo law): every DB read/write goes through sqlc-generated
   queries; raw SQL is a contract breach. New schema file => added to
   sqlc.yaml in the same commit; `just sqlc-check` in the gate list.
