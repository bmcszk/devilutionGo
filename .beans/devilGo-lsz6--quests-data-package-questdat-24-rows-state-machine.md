---
# devilGo-lsz6
title: 'Quests data package: questdat (24 rows) + state machine + per-class/per-mode matrix'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
---

# Quests: 24 rows, state machine, availability matrix

## Goal
Transcribe `questdat.tsv` and port the quest lifecycle (`InitQuests`/`CheckQuests`/
`CheckQuestKill`/`ResyncQuests`) including the seed-derived quest pools and the SP/MP/spawn/
Hellfire availability matrix.

## Research
- `/tmp/dg-requirements-append.md` **V6.1** (24 rows, loader `tables/questdat.cpp:80`), **V6.2**
  (state machine, `Source/quests.cpp` 623 lines), **V5.2** (quest dialog matrix), **V12.9**
  (spawn wipe), **V12.10** (`bCowQuest`/`bTheoQuest` still code-branched).
- TSV: `assets/txtdata/quests/questdat.tsv`, **no** hf override. Header
  `qdlvl|qdmultlvl|qlvlt|bookOrder|qdrnd|qslvl|isSinglePlayerOnly|qdmsg|qlstr`; `qdmultlvl = -1`
  means unavailable in MP-shared mode; `qslvl` is a `_setlevels` id; `bookOrder` is the quest-log
  sort key; `qdrnd` the random-pool weight.
- Enums: `quest_id` (`tables/questdat.hpp:74-99`, 24 + `Q_INVALID=-1`), `quest_state`
  (`:62+`: NOTAVAIL, INIT, ACTIVE, DONE + hf `QUEST_HIVE_TEASE1/TEASE2/ACTIVE/DONE`),
  `MAXQUESTS 24` (`:14`), `QuestData` (`:117-129`).
- `InitQuests()` `quests.cpp:60-110`: forced mushroom cells (`:61-62`), `WaterDone = 0` (`:64`),
  per-quest copy of `_qlvltype/_qslvl/_qdmsg`, `_qactive = NOTAVAIL`, `position = (0,0)`; SP =>
  `_qlevel = _qdlvl` + `INIT`; MP => `_qlevel = _qdmultlvl`, skipped when `isSinglePlayerOnly`
  (`:76-85`); then `!MP && randomizeQuests` => **`InitialiseQuestPools(DungeonSeeds[15], Quests)`**
  (`:88-92`, body `:114-135`) — quest order derives from the **level-15 seed**; `gbIsSpawn` wipes
  all to NOTAVAIL (`:94-99`); tail fixes `Q_SKELKING._qvar2 = 2`, `Q_ROCK._qvar2 = 2` when
  unavailable, `Q_LTBANNER._qvar1 = 1` always, MP `Q_BETRAYER._qvar1 = 2`, MP mushroom pre-spawns
  the tome (`QS_TOMESPAWNED`) (`:101-110`).
- `CheckQuests()` `:137-201`: MP betrayer spawns `OBJ_ALTBOY` at `SetPiece.pos + (4,6)` when
  `_qvar1 == 2`; SP betrayer portal `AddMissile(...RedPortal)` + `_qvar2 = 1`; return portal at
  **(35,32)** in `SL_VILEBETRAYER` when `_qvar2 == 4`; Poison Water auto-completes when
  `ActiveMonsterCount == 4`; set-level entry when standing on `quest.position` with
  `_qslvl != 0` => `StartNewLvl(WM_DIABSETLVL, _qslvl)`.
- `CheckQuestKill()` `:227-287` (early return when `gbIsSpawn` `:229`): `MT_SKING` -> Q_SKELKING
  DONE + `RestWellLeoricIllFindYourSon`; `MT_CLEAVER` -> Q_BUTCHER DONE +
  `TheSpiritsOfTheDeadAreNowAvenged`; `Garbud` -> Q_GARBUD DONE + `ImNotImpressed`; `Zhar` ->
  Q_ZHAR DONE + `ImSorryDidIBreakYourConcentration`; `Lazarus` -> Q_BETRAYER DONE +
  `YourMadnessEndsHereBetrayer` **and `Q_DIABLO -> ACTIVE`**.
- `ResyncQuests()` `:397`; `_qvar1`/`_qvar2` `QS_*` constants (`QS_MUSHGIVEN`, `QS_BRAINSPAWNED`,
  `QS_TOMESPAWNED`); mushroom-drop interaction with `SpawnItem` (V1.7 `:3435-3453`).

## Scope
`internal/gamedata/questdat/`: `QuestData` table (24 rows, all 9 columns), `QuestID`/
`QuestState` enums, `QVar` constant table, availability matrix function
`InitialQuestState(q, mode spawn|retail, sp|mp, randomize, hf)`, `InitialiseQuestPools` with
`DungeonSeeds[15]`, kill-hook table `map[MonsterID]QuestAdvance` (instead of an if-chain), and
per-quest `qdmsg`/`qlstr` references into dg-task-116 text ids. FSM *runtime* stays dg-task-026.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Questdat` = **24** rows base **and** hf (assert no hf
      override file exists), `QuestID` = 24 + `Q_INVALID`, `QuestState` = 4 + 4 hive states;
      `bookOrder` values are a permutation with no duplicates.
- [ ] Seed-derived pool worked example: `InitialiseQuestPools(DungeonSeeds[15], ...)` for 10
      fixed level-15 seeds reproduces the C++ quest-availability order exactly (tape test),
      including `randomizeQuests` off => deterministic `qdlvl` ordering.
- [ ] Availability matrix table test: for each of the 24 quests assert the
      (SP, MP, spawn, hf) state quadruple — e.g. `isSinglePlayerOnly` quests are NOTAVAIL in MP,
      `qdmultlvl == -1` quests are unavailable in MP, and **all 24** are NOTAVAIL when `spawn` is
      on (V12.9).
- [ ] Kill-hook worked example: killing `MT_SKING`, `MT_CLEAVER`, Garbud, Zhar, Lazarus each
      advances exactly the documented quest + emits the documented `HeroSpeech`, and Lazarus also
      flips `Q_DIABLO` to ACTIVE; no other monster kill changes any quest.
- [ ] Poison Water auto-completion fires only when `ActiveMonsterCount == 4` in the set level
      (boundary 3/4/5 tested); betrayer portal coordinates (35,32) asserted.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — small FSM + hook table, no reflection
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded pool tapes
- /home/blaze/.agents/skills/coding-rules/SKILL.md — data table vs lifecycle split

## Dependencies & blockers
blocked_by dg-task-101. Needs dg-task-104 (`CheckQuestKill` monster ids), dg-task-107
(`OBJ_ALTBOY`), dg-task-108 (quest dialog cells), dg-task-116 (`qdmsg` text ids). Blocks
dg-task-109, dg-task-115, dg-task-118, and existing dg-task-026.

## Open questions
- Is `qslvl` -> set-level mapping owned here or in dg-task-115? Recommend: enum here, placement
  there.
- MP mushroom tome pre-spawn (`QS_TOMESPAWNED`) interacts with the desync comment at
  `items.cpp:3446-3453`; confirm whether the Go port needs the "every client sees it" double
  spawn or the loopback design removes it.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:dep-typo-116-should-be-112/h6wen + remap-deps
