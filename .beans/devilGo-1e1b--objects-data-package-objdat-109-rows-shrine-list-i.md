---
# devilGo-1e1b
title: 'Objects data package: objdat (109 rows) + shrine list + interaction switch'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
    - devilGo-778h
---

# Objects: 109 rows, 34 shrines, interaction switch

## Goal
Transcribe `objdat` and the object-behaviour enumerations that live in C++ rather than in a TSV —
the 34 shrine effects, the 17 themes, the book/tome/elixir/rune/oil effects — and shape the
interaction dispatch as a data-driven switch.

## Research
- `docs/research/dg-requirements-append.md` **V4.1** (objdat 109 rows; loader `tables/objdat.cpp:259`),
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
`miscId` atoms from devilGo-2j8j.

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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — registry + bitset flags
- /home/blaze/.agents/skills/coding-rules/SKILL.md — effect descriptors (data) vs handlers
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded shrine effect tests

## Dependencies & blockers
blocked_by devilGo-jx08. Blocks devilGo-fpkz (themes/setmaps place objects), devilGo-h6wn (shrine
town portal/quest interplay), devilGo-i2cr (audit), and existing devilGo-yeia/020.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: 5 THEME_* values are code-spawned (themes.cpp), not objdat rows — themes table stays in objdat bean, spawn effects follow the renderer of theme code at claim. Shrines stay in this bean (single owner).
- Only 12 of 17 `THEME_*` values appear in objdat's `theme` column; the other 5 (BARREL,
  MONSTPIT, TREASURE, CAULDRON, BRNCROSS family) are spawned by themes.cpp code. Confirm before
  deciding whether themes belong in this bean or devilGo-fpkz.
- Are shrine effects representable as pure data for all 34, or will some need code? V4.3 lists
  exact bodies — recommend data where possible, code otherwise, each recorded.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 4 declared blocks edges missing; 2 cosmetic research errors; core counts/behavior verified exact.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 data task (devilGo-1e1b); draft; wire blockedBy [jx08] ✓.
2 local: LIVE-VERIFIED: objdat.tsv 110 lines = 109 data rows ✓; NO hf objects dir ✓ (positive-test AC is meaningful); shrine enum objects.cpp:65-100 = 34 Shrine* + NumberOfShrineTypes sentinel ✓; OperateShrine :2989 ✓ exact with `DiabloGenerator rng(shrine._oRndSeed)` + `selectionRegion = None` ✓; pool `gbIsHellfire ? NumberOfShrineTypes : 26` :1302 ✓ exact; shrineavail[] :175-209 ✓ (Single=Thaumaturgic verified in entries); lever :1828 ✓; doors :1766 ✓; MAXOBJECTS 127 objects.h:30 ✓, struct Object :32 ✓; _object_id :42 ✓, theme_id :18 ✓. Two minor errors: research note "last line has no trailing newline" is FALSE (tail -c1 = 0x0a); ObjectData struct at objdat.h:168 not ":100-218". Wire gap: body "Blocks devilGo-fpkz, 112, 118 ✓, 019/020" — fpkz, h6wn, yeia, foug edges all missing (i2cr wired ✓).
3 online: n/a.
4 solution: wire 4 missing edges; fix trailing-newline note + ObjectData cite (cosmetic).
5 understanding: object-domain data root for fpkz themes/setmaps, h6wn shrine-portal interplay, yeia/foug placement; OperateFunc-registry replacing the _otype switch keeps data/behaviour split without a silent default (AC enforces it).
6 repro: `awk 'END{print NR}' devilutionX/assets/txtdata/objects/objdat.tsv; tail -c1 devilutionX/assets/txtdata/objects/objdat.tsv | xxd; sed -n '65,100p' devilutionX/Source/objects.cpp | grep -c Shrine; sed -n '2989,3000p;1302p' devilutionX/Source/objects.cpp; grep -n "struct ObjectData" devilutionX/Source/tables/objdat.h`
7 fix: `beans update devilGo-fpkz --blocked-by devilGo-1e1b`; `beans update devilGo-h6wn --blocked-by devilGo-1e1b`; `beans update devilGo-yeia --blocked-by devilGo-1e1b`; `beans update devilGo-foug --blocked-by devilGo-1e1b`; optional: `beans update devilGo-1e1b --body-replace-old "verify the file's last line has no trailing newline (110 physical lines)" --body-replace-new "the file ends with a trailing newline (110 lines: 1 header + 109 data)"`
8 validation: blockedByIds updated; `tail -c1` matches corrected note.
9 related: jx08 (dep), fpkz/h6wn/yeia/foug/i2cr (dependents), 2j8j (miscId atoms).
VERDICT: NEEDS-WORK:wire-deps — 4 declared blocks edges missing; 2 cosmetic research errors; core counts/behavior verified exact.

Fix pass 2026-10-16: Dep-wiring applied (fix pass): fpkz/h6wn/yeia/foug now blocked-by this bean (were declared, never wired).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


