---
# devilGo-lz2m
title: 'boot: per-hero single_N.db store model (ADR-003 D1 migration)'
status: draft
type: task
priority: normal
created_at: 2026-11-25T14:53:02Z
updated_at: 2026-11-25T14:53:12Z
parent: devilGo-xih1
blocked_by:
    - devilGo-v13s
---

# boot: per-hero single_N.db store model (ADR-003 D1 migration)

## Goal
Migrate the windowed boot's hero storage from the single multi-row single_0.db to ADR-003's per-hero model — one .db file per hero (single_N.db), the hero list probing slots 0..98 like upstream pfile_ui_set_hero_infos (pfile.cpp:695-726 @e1ad0c2b6: OpenSaveArchive(i) per slot), first-free-slot and the 99-cap enforced at file enumeration.

## Research
- hl2 triage (task-hl2-load-triage, 2026-11-25; .hermes/lane/m2-hero-load/v13s-triage-result.md section 1b): the migration is a UX + storage-handle redesign — selHeroScreen.store re-shapes from one *persist.SQLiteStore into a slot->handle map, touching refresh (selhero.go:158-176), firstFreeSlot (:467-489), performFocusedDelete (:240-254); the Load path (v13s) never touches any of it, hence the split.
- D1 decision (dcob triage block): ADR-003 "one hero = one .db file" (docs/adrs/adr-003.md:12) stands.

## Scope
- selHeroScreen storage: slot->handle map over per-hero single_N.db files.
- Hero-list enumeration + Load/New-Game flows read through the map; first-free-slot = first missing file number; 99-cap at enumeration.
- Load path (v13s) keeps working unchanged across the migration (its tests pin it).

## Acceptance criteria
- [ ] Boot opens per-hero single_N.db files; the hero list reflects files on disk (integration test with real temp files)
- [ ] first-free-slot = first missing file number; 99-cap refused at enumeration
- [ ] v13s load-path tests stay green across the migration
- [ ] Full gates incl. -tags windowed; RED evidence per TDD law

## Libraries
stdlib + modernc.org/sqlite (existing).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split from v13s by the hl2 brainer triage (2026-11-25, parent-verified); blocked-by v13s (the load path lands first on the single-db simplification, then this re-shapes the handle). Deep triage owed before todo: schema/handle-lifetime decisions (open-99 vs LRU handles) need a pass.
