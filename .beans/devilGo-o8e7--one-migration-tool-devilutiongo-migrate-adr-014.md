---
# devilGo-o8e7
title: One migration tool devilutionGo-migrate (ADR-014)
status: todo
type: feature
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-i56j
---

# One migration tool: devilutionGo-migrate

## Goal

One CLI, `cmd/devilutionGo-migrate`, per [ADR-014](../docs/adrs/adr-014.md):
assets from original MPQs into the assets dir (layout law
[ADR-005](../docs/adrs/adr-005.md)), legacy saves into SQLite (law
[ADR-003](../docs/adrs/adr-003.md)), and configs plus registries into YAML
flat / CSV tabular files (law [ADR-013](../docs/adrs/adr-013.md)). The tool
is a thin CLI over already-round-trip-proven producers; nothing new has to
invent a format at migration time.

## Research

- internal/assets already decodes 100% of records from retail MPQs
  (exhaustive-extraction proof in the asset verification PoW).
- internal/persist offers SaveGameData/ItemPack/PlayerPack codecs, a sqlite
  store, plus importers; cmd/gamedatagoldens shows the golden-comparison
  pattern to reuse for table output.
- internal/engine/ini_import reads diablo.ini; txtdata TSV loaders decode the
  gamedata tables; [ADR-004](../docs/adrs/adr-004.md) keeps INI as the
  settings format, migration only produces the flat YAML view for registries.

## Scope

Children:

1. Migrate assets: MPQ into assets dir + YAML asset lists.
2. Migrate saves: legacy .sv into SQLite through the persist importers.
3. Migrate configs/registries: INI view into YAML flat, tables into CSV tabular.
4. Migration docs + CI fixture path.

## Acceptance criteria

- [ ] One binary covers all three lanes with checkpointed, re-runnable steps.
- [ ] No MPQ or save binaries ever enter the repository; outputs go to the configured dirs.
- [ ] Each lane prints a final checksum summary and exits non-zero on any mismatch.

## Libraries

- modernc.org/sqlite (already the store driver per the accepted stack record)
- No new dependencies.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md

## Dependencies & blockers

- Lane 2 leans on the persistence store work that is in progress; it reuses
  its importers unchanged.

## Open questions

- None blocking; YAML schema of the asset manifest should follow the graphics
  manifest feature so the two do not fork.

## Triage

TRIAGE brainer evidence 2026-11-03: ADR-014 accepted in the architecture PR
with this exact lane list, after reading all prior records plus the library
catalog. Bean-store lane mapping here mirrors the accepted text verbatim.
