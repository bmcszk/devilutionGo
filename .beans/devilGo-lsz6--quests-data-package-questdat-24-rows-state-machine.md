---
# devilGo-lsz6
title: 'Quests data package: questdat (24 rows) + state machine + per-class/per-mode matrix'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:39:19Z
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
- `docs/research/dg-requirements-append.md` **V6.1** (24 rows, loader `tables/questdat.cpp:80`), **V6.2**
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
per-quest `qdmsg`/`qlstr` references into devilGo-orkj text ids. FSM *runtime* stays devilGo-xw8k.

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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — small FSM + hook table, no reflection
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded pool tapes
- /home/blaze/.agents/skills/coding-rules/SKILL.md — data table vs lifecycle split

## Dependencies & blockers
blocked_by devilGo-jx08. Needs devilGo-778h (`CheckQuestKill` monster ids), devilGo-1e1b
(`OBJ_ALTBOY`), devilGo-rctk (quest dialog cells), devilGo-orkj (`qdmsg` text ids). Blocks
devilGo-h6u0, devilGo-fpkz, devilGo-i2cr, and existing devilGo-xw8k.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: qslvl enum owned here, set-level placement in the levels-content bean (single-source split). QS_TOMESPAWN: port the desync-comment behavior verbatim with a parity test on seed 0.
- Is `qslvl` -> set-level mapping owned here or in devilGo-fpkz? Recommend: enum here, placement
  there.
- MP mushroom tome pre-spawn (`QS_TOMESPAWNED`) interacts with the desync comment at
  `items.cpp:3446-3453`; confirm whether the Go port needs the "every client sees it" double
  spawn or the loopback design removes it.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:dep-typo-116-should-be-112/h6wen + remap-deps


## IMPL NOTES (triage 2026-10-16)
1 root: questdat transcription + lifecycle/availability matrix; = devilGo-lsz6.
2 local: beans show ✓; questdat.tsv 24 rows ✓; NO hf override ✓ (mods/hf/txtdata/quests/ does not exist); InitQuests mushroom forcing quests.cpp:60-62 ✓ verbatim (TOWN_HEALER/Q_MUSHROOM/TEXT_NONE + TOWN_WITCH/TEXT_MUSH9); InitialiseQuestPools from level-15 seed :88-92 + body :114 ✓ (DiabloGenerator, pickRandomlyAmong, seed 988045466 case); questdat.cpp:80 LoadQuestData ✓.
3 online: n/a
4 solution: dep typo — "devilGo-orkj (qdmsg text ids)" should be devilGo-h6wn = devilGo-h6wn (textdat owns speech ids; 116/orkj is misc enums: cursors/gold/XP caps). Remap (101→jx08; also needs 104→778h, 107→1e1b, 108→rctk); S2.
5 understanding: diablo-layer quest tables + initial-state matrix; FSM runtime stays in xw8k (documented from this side already).
6 repro: beans show devilGo-lsz6; sed -n '60,64p;88,95p;114,120p' devilutionX/Source/quests.cpp; ls devilutionX/mods/hf/txtdata/ (no quests dir)
7 fix: beans update devilGo-lsz6 --blocked-by devilGo-jx08; body-replace-old "devilGo-orkj (qdmsg text ids)" --body-replace-new "devilGo-h6wn / devilGo-h6wn (qdmsg text ids)".
8 validation: beans show devilGo-lsz6 shows corrected dep + edge.
9 related: devilGo-xw8k, devilGo-rctk, devilGo-h6wn, devilGo-778h, devilGo-1e1b
VERDICT: NEEDS-WORK:dep-typo-116-should-be-112/h6wen + remap-deps

Fix pass 2026-10-16: Dep fix applied (fix pass): dep-typo corrected (116 → 112/h6wn) + remap to devilGo ids.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


