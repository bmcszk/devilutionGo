---
# devilGo-fpkz
title: 'Levels content: dungeon types, trigs, themes, setmaps, unique level features'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
---

# Levels content: trigs, themes, setmaps, unique features

## Goal
Turn the V11 level *content* inventory into Go data + tables: dungeon-type bands, trigger piece
lists with their tile rules, the 17 themes and their effects, the set-level map list with
entry tiles/palettes/transparency, and the unique-feature checklist. Generators themselves stay
in dg-task-017/018/019.

## Research
- `/tmp/dg-requirements-append.md` **V11.1** (dungeon types / level bands,
  `Source/levels/gendung_defs.hpp:16-25`), **V11.2** (trigs), **V11.3** (themes), **V11.4**
  (setmaps), **V11.5** (unique features), **V4.1** (objdat `theme` column), **V12.2** (enums).
- Enums: `dungeon_type` TOWN/CATHEDRAL/CATACOMBS/CAVES/HELL/NEST/CRYPT (`DTYPE_LAST=CRYPT`,
  `DTYPE_NONE=-1`); `_setlevels` (`gendung_defs.hpp:47-60`) NONE, SKELKING, BONECHAMB, MAZE,
  POISONWATER, VILEBETRAYER, ARENA_CHURCH, ARENA_HELL, ARENA_CIRCLE_OF_LIFE + sentinels
  (`SL_FIRST_ARENA`, `SL_LAST`); `theme_id` = **17** + `THEME_NONE=-1` (objdat.h:18-37);
  `lvl_entry` (`:28-37`) MAIN/PREV/SETLVL/RTNLVL/LOAD/WARPLVL/TWARPDN/TWARPUP.
- Trigs (`Source/levels/trigs.cpp`): piece lists `:31-59` — `TownDownList {715,714,718,719,720,
  722,723,724,725,726}`, `TownWarp1List {1170..1178,1180,1182,1184}`, `TownCryptList
  {1330..1337}`, `TownHiveList {1306,1307,1308,1309}`, `L1UpList {126,128,129,130,131,132,134,
  136,137,138,139}`, `L1DownList {105,106,107,108,109,111,113,114,117}`, `L2UpList {265,266}`,
  `L2DownList {268,269,270,271}`, `L2TWarpUpList {557,558}`, `L3UpList {169..182}`, L3/L4 lists
  `:48-59`; `trigs[MAXTRIGGERS]`, `numtrigs`, `TWarpFrom`, `trigflag` (`:27-29`). Per-type rules:
  `InitTownTriggers` `:109-149` stairs at **(25,29)** + conditional warps Catacombs **(49,21)**
  `_tlvl=5`, Caves **(17,69)** `9`, Hell **(41,80)** `13`, Nest **(80,62)** `17`, Crypt **(36,24)**
  `21` gated by `IsWarpOpen`; L1 `128`->PREV / `114`->NEXT; L2 `266`->PREV (skipped on the
  Chamber-of-Bone quest tile) / `558`->TWARPUP `_tlvl=0` / `270`->NEXT; L3 `170`/`167`/`548`;
  L4 `82`/`421`/`119` and second pass `369`->NEXT **only when `Q_BETRAYER == QUEST_DONE`**;
  Hive `65`/`62`/`79`; Crypt `183`/`157`/`125`; plus `InitSKingTriggers`, `InitSChambTriggers`,
  `InitPWaterTriggers`, `InitVileTriggers`, `InitPlayerTriggers`.
- Themes (`Source/levels/themes.cpp`): 17 `Theme_*` init fns (`:377`-`:785`, listed in V11.3);
  `InitThemes()` `:349-449` per-level `themeCount` + `ttval`; special rooms use
  `ThemeGood[GenerateRnd(4)]` retrying `SpecialThemeFit` else `GenerateRnd(17)` (`:880-886`);
  `HoldThemeRooms()` `:891-916` marks `Populated` (skips level 16 + NEST/CRYPT);
  `CreateThemeRooms()` `:918-980` + `UpdateL4Trans()` for HELL; `THEME_NONE` -> `app_fatal`
  (`:977`).
- Setmaps (`Source/levels/setmaps.cpp:112-165`): per V11.4 — sklkng1/2.dun entry {83,44} +
  `sklkngt.dun`/`l1data\l1_2.pal`; bonecha2/1 entry {70,40} + `bonechat.dun`/`l2_2.pal`;
  foulwatr entry {31,83} + `LoadPWaterPalette`/`UpdatePWaterPalette` (`quests.cpp:326-347`);
  vile1/2 entry {35,36} + `vile1.dun` transparency + return portal (35,32); arena church
  {29,22} (visual {28,20}), arena hell {34,26} ({33,26}), circle_of_death {30,26} ({29,26});
  `GetArenaLevelType` (setmaps.h:19-30) church=CATHEDRAL, hell/circle=HELL; `LoadSetMap` flips
  `Q_SKELKING` INIT->ACTIVE `_qvar1=1` (`:121-125`).
- Unique features (V11.5): Foul Water `WaterDone` bitmask (`quests.cpp:38`) +
  `PureWaterPalette` (`:48`) + `StartPWaterPurify` (`:50-58`) + auto-complete at 4 monsters
  (`:176-187`); Chamber of Bone (`Q_SCHAMB`, suppressed up-stair trigger); King Leoric's Tomb
  (`LeoricAi`, coffin -> `fprst3.smk`, reward `Uniq(SKCROWN)`); Vile betrayer (`OBJ_ALTBOY` MP,
  RedPortals, L4 tile 369); Butcher; Halls of the Blind; Anvil of Fury (`IDI_ANVIL` at
  `SetPiece.pos + (11,11)`, `items.cpp:2455`); Gharbad's cave + goat shrine theme; Zhar's nest +
  Snotspill; Cornerstone of the World (`Q_CORNSTN`); The Defiler (`Q_DEFILER`); Na-Krul Uber
  rooms (lever `objects.cpp:1828-1833`, `UberRow`/`UberCol`, `IsUberRoomOpened`, death -> 3 magic
  weapons + stream SFX `monster.cpp:931-940`); Hork Demon -> Theodore (`bTheoQuest`,
  `IDI_THEODORE`); Little Girl/Celia (`Q_GIRL`, `TOWN_GIRL`, L17); Wandering Trader (`Q_TRADER`,
  Nest light radius 9); Farmer's Orchard/Hive + cow variant (dg-task-109); Ghoul Dance/graveyard/
  Crypt (`Q_GRAVE`); arena set levels; Nest/Crypt palettes + cycling (`diablo.cpp:3509-3529`).

## Scope
`internal/gamedata/leveldat/`: `DungeonType`/`SetLevelID`/`ThemeID`/`LevelEntry` enums, trigger
tables (`piece list -> {message, _tlvl}` per dungeon type, incl. the conditional L4 `369` and the
L2 quest-tile suppression), town warp tile table, theme effect table (`theme_id -> object/monster
spawn spec`), setmap manifest (`files, entry tile, visual entry, transparency file, palette,
objs fn, trigs fn`), and the unique-feature checklist as data with quest-state links.

## Acceptance criteria
- [ ] TDD first: content-count parity — `ThemeID` = **17** + `THEME_NONE`, `SetLevelID` = 8
      values + sentinels, `DungeonType` = 7 + `DTYPE_NONE`, and the trigger tables cover all 8
      `Init*Triggers` families with the exact piece-list cardinalities from `trigs.cpp:31-59`
      (e.g. `TownDownList` 10 ids, `TownCryptList` 8, `TownHiveList` 4, `L1UpList` 11,
      `L1DownList` 9).
- [ ] Worked example: standing on town tile **(25,29)** emits `WM_DIABNEXTLVL`; **(80,62)** ->
      level 17 only when `IsWarpOpen(DTYPE_NEST)`; **(36,24)** -> level 21 only for the Crypt;
      and on L4, piece **369** yields NEXT *only* when `Q_BETRAYER == QUEST_DONE` (test both).
- [ ] Setmap manifest byte-check: every `LoadSetMap` case's file list, entry tile and visual
      tile equal V11.4 (e.g. SL_ARENA_CHURCH {29,22} visual {28,20}), and the palette/
      transparency files exist in the asset manifest (dg-task-206).
- [ ] Theme effect test: each of the 17 themes spawns exactly the documented object/monster set
      from objdat's `theme` tag (12 themes tag objdat rows; the remaining 5 are code-spawned
      monster/bomb clusters — assert both lists and their union == 17).
- [ ] Unique-feature checklist parity: the V11.5 features each have a data entry keyed by quest
      id (Q_SKELKING, Q_BUTCHER, Q_POISONWATER/Q_SCHAMB, Q_BETRAYER, Q_GARBUD, Q_ZHAR, Q_ANVIL,
      Q_CORNSTN, Q_DEFILER, Q_GIRL, Q_TRADER, Q_GRAVE, Q_MUSHROOM, Q_FARMER/Q_JERSEY) and the
      Na-Krul Uber + Hork/Theodore branches are represented; `HoldThemeRooms` skips level 16 and
      NEST/CRYPT (tested).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — data-driven tables, no switch soup
- /home/blaze/.agents/skills/coding-rules/SKILL.md — content manifest vs generator split
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — palette/transparency references

## Dependencies & blockers
blocked_by dg-task-101. Needs dg-task-107 (objdat theme tags/object ids), dg-task-110 (quest ids
for conditional triggers), dg-task-104 (theme monster types). Blocks dg-task-109 (Hive/Crypt
tiles shared), dg-task-118 (audit), existing dg-task-019/020 and dg-task-206 (asset manifest).

## Open questions
- Are the Nest/Crypt `.dun` files in `mods/hf/nlevels/` or in `hellfire.mpq`, and which does the
  Go port read? Both paths are cited in V11/V12.10; needs one decision.
- V11.2 lists `L3UpList {169..182}` collapsed — expand it from `trigs.cpp:48-59` before writing
  the test, don't infer the range.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps + 1 wrong cite (InitThemes :349-449 → :826) — content otherwise verified exact.
