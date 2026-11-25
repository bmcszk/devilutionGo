---
# devilGo-v13s
title: 'selhero: hero-row Load Game / New Game over a production SessionArchive'
status: draft
type: task
priority: normal
created_at: 2026-11-25T12:12:09Z
updated_at: 2026-11-25T12:12:09Z
parent: devilGo-xih1
blocked_by:
    - devilGo-dcob
---

# selhero: hero-row Load Game / New Game over a production SessionArchive

## Goal
Activating a hero ROW with a saved game offers Load Game / New Game (upstream SelheroListSelect, selhero.cpp:217-236 @e1ad0c2b6); Load submits KindLoadSlot and the next settled tick logs KindLoadedSlot; a hero with no saved game goes straight to new-game (:239). Includes migrating the boot's hero list to the ADR-003 per-hero single_N.db store model (D1 decision recorded in the dcob triage block).

## Research
- Brainer triage hl1 (task-hl1-triage, 2026-11-25; .hermes/lane/m2-hero-load/result.md section dcob 2d): upstream flow verified at e1ad0c2b6 (anchor ca91d778f stays canonical); gbLoadGame from SELHERO_CONTINUE (menu.cpp:130).
- Seams that exist: kernel verbs KindSaveSlot/KindLoadSlot + events KindSavedSlot/KindLoadedSlot (internal/game/kernel/command.go:125-150), world.RegisterSaveLoad handlers (internal/game/world/session.go:233-281), SessionArchive seam (session.go:87-97). Missing: production SessionArchive (only the test-local slotArchive in internal/integration/session_stage_test.go), RegisterSaveLoad not registered in bootDeviceLane (device_lane.go:108-115), a boot LevelSource.
- ADR-003 "one hero = one .db file" (docs/adrs/adr-003.md:12); ADR-008 world owns no storage.

## Scope
- Production SessionArchive adapter over persist.SQLiteStore in the diablo layer (promote the tested slotArchive shape out of internal/integration); NOT internal/game, NOT internal/persist importing world.
- RegisterSaveLoad in the boot; windowed-boot LevelSource disclosure updated.
- Per-hero single_N.db enumeration (first-free-slot = first missing file number; MAX_CHARACTERS 99 cap enforced at enumeration).
- Hero-row load/new-game list in the selhero UI + compose draw site.

## Acceptance criteria
- [ ] Production SessionArchive roundtrip test (save -> close -> reopen -> load -> hash-match) beside the store
- [ ] Hero row with a saved game offers Load/New Game; Load submits KindLoadSlot; KindLoadedSlot in the tick log; no-save hero goes straight to new game
- [ ] Hero list enumerates per-hero single_N.db files per ADR-003; first-free-slot + 99 cap enforced at enumeration
- [ ] Full gates green incl. -tags windowed for the windowed surfaces; RED evidence per TDD law

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Split from dcob by the hl1 brainer triage (task-hl1-triage, 2026-11-25; parent-verified). Blocked-by dcob (D1-D10 land first, RF1 included).
