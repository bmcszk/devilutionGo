---
# devilGo-1e1b
title: 'Objects data package: objdat (109 rows) + shrine list + interaction switch'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-15T18:29:02Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
---

# Objects: 109 rows, 34 shrines, interaction switch

## Goal
Transcribe `objdat` and the object-behaviour enumerations that live in C++ rather than in a TSV —
the 34 shrine effects, the 17 themes, the book/tome/elixir/rune/oil effects — and shape the
interaction dispatch as a data-driven switch.

## Research
- `/tmp/dg-requirements-append.md` **V4.1** (objdat 109 rows; loader `tables/objdat.cpp:259`),
  **V4.2** (interaction algorithm, `Source/objects.cpp` 4928 lines), **V4.3** (all 34
  `shrine_type` values, `objects.cpp:65-100`), **V4.4** (book/tome/elixir/rune/oil),
  **V11.3** (themes), **V12.10** (shrine pool 34 vs 26 code branch).
- TSV: `assets/txtdata/objects/objdat.tsv` — **no** hf override (`mods/hf/txtdata/` has no
  objects dir). Columns include `id|file|minLevel|maxLevel|levelType|theme|quest|flags|animDelay|
  animLen|animWidth|selectionRegion` (verified from the header). 109 data rows; note the file's
  last line has no trailing newline (110 physical lines), and distinct `file` values = 62.
  Distinct `theme` atoms used: 12 `THEME_*` values (not all 17 themes appear as objdat tags).
- Structs/enums: `ObjectData` (objdat.h:100-218), `theme_id` (objdat.h:18-37 = **17 themes +
  THEME_NONE=-1**), `_object_id` (objdat.h:42+, order == objdat rows), `ObjectFlag`
  (Animated, Solid, Light, Breakable, Trap, MissilesPassThrough, ...).
- Shrine enum verified on disk: `objects.cpp:65-99` lists **34** `Shrine*` values; the 35th
  token is the `NumberOfShrineTypes` sentinel (V4.3's "34" is correct).
- Interaction: `HandleObjects` flags driving anim/solid/light/trap/breakable; `InteractObject` ->
  `OperateObject` switch on `_otype` with `selectionRegion` cleared after use (`:2999`);
  `OperateShrine` `:2989-3105` reseeds `DiabloGenerator rng(shrine._oRndSeed)` (`:2996`) and
  switches on `_oVar1` (`:3004-3080`, Murphys at `:3101`); `shrineavail[]` `:183-209`
  (Any|Single|Multi — Single = Thaumaturgic/Solar, Multi-excluded = Spooky/Tainted); pool size
  `gbIsHellfire ? NumberOfShrineTypes : 26` (`:1302`); doors `:1766-1775`, lever `:1828`/`:1833`
  (Uber `UberRow/UberCol`), chest `:2024`/trap `:2066`, sarcophagus `:2194`, pedestal
  `:2229-2241`; `MAXOBJECTS 127`, `dObject`, `ObjMem`/`Objects`/`ActiveObjects`
  (objects.h:32-80).

## Scope
`internal/gamedata/objdat/`: `ObjectData` table (109), `ObjectID` enum in row order, `ThemeID`
(17 + NONE), `ObjectFlag` bitset, `ShrineType` (34), `shrineavail[]` table, and an
`OperateFunc` registry `map[ObjectID]Operator` replacing the `_otype` switch — including the
34 shrine handlers' *effect descriptors* (data-shaped where possible: stat deltas, durability
deltas, item lists; code-shaped for the `rng`-driven picks). Books/tomes/elixirs/oils cross-ref
`miscId` atoms from dg-task-102.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Objdat` = **109** rows base **and** hf (assert the
      absence of an hf override as a positive test, so a future hf file cannot silently drift),
      plus `ThemeID` = **17** + `THEME_NONE`, `ShrineType` = **34** (sentinel excluded).
- [ ] Worked example (shrine effects, deterministic per `_oRndSeed`): ShrineMysterious =
      -1 Str/Mag/Dex/Vit then +6 to one chosen by `rng.generateRnd(4)`; ShrineHidden = +10
      `_iDurability`/`_iMaxDur` to every equipped non-indestructible item (else random stat
      down); ShrineGloomy = weapons lose 1 `_iMaxDam` (never below `_iMinDam`), armour loses 1
      `_iMaxAC`. All three asserted with a fixed seed.
- [ ] Shrine availability table parity: `Single` set == {Thaumaturgic, Solar}, `Multi`-excluded
      set == {Spooky, Tainted}, and pool size is 26 (Diablo) vs 34 (Hellfire) for the same seed.
- [ ] Selection-region + flags: every objdat row's `selectionRegion`/`flags` decode to the same
      bitset as the C++ (`SelectionRegion` None/Bottom/Middle/Top, cursor_defs.hpp:14-19);
      `Animated` rows carry `animDelay`/`animLen`/`animWidth` used by `HandleObjects`.
- [ ] Operator registry completeness: every `_object_id` that the C++ `OperateObject` switch
      handles has a registry entry or an explicit documented `nil` (no silent default).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — registry + bitset flags
- /home/blaze/.agents/skills/coding-rules/SKILL.md — effect descriptors (data) vs handlers
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded shrine effect tests

## Dependencies & blockers
blocked_by dg-task-101. Blocks dg-task-115 (themes/setmaps place objects), dg-task-112 (shrine
town portal/quest interplay), dg-task-118 (audit), and existing dg-task-019/020.

## Open questions
- Only 12 of 17 `THEME_*` values appear in objdat's `theme` column; the other 5 (BARREL,
  MONSTPIT, TREASURE, CAULDRON, BRNCROSS family) are spawned by themes.cpp code. Confirm before
  deciding whether themes belong in this bean or dg-task-115.
- Are shrine effects representable as pure data for all 34, or will some need code? V4.3 lists
  exact bodies — recommend data where possible, code otherwise, each recorded.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
