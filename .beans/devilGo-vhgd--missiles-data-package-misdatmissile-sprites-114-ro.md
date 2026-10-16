---
# devilGo-vhgd
title: 'Missiles data package: misdat+missile_sprites (114 rows) + Add/Process registry'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
---

# Missiles: 114 rows + behaviour registry

## Goal
Transcribe `misdat` (68) + `missile_sprites` (46) and build the Go registry that replaces the
C++ function-pointer tables: every `addFn`/`processFn` name referenced by the data resolves to a
registered implementation.

## Research
- `/tmp/dg-requirements-append.md` **V3.1** (misdat columns; loader `tables/misdat.cpp:373-396`),
  **V3.2** (behaviour classes / Go registry), **V3.3** (missile_sprites; loader `misdat.cpp:134`),
  **V12.10** (hf deltas), **V12.12** (counts).
- TSVs: `assets/txtdata/missiles/{misdat,missile_sprites}.tsv`,
  `mods/hf/txtdata/missiles/{misdat,missile_sprites}.tsv`. misdat columns
  `id|addFn|processFn|castSound|hitSound|graphic|flags|movementDistribution`; sprite columns
  `id|width|width2|name|numFrames|flags|frameDelay[16]|frameLength[16]`.
- **Registry sizes (measured, not assumed)**: base misdat references **51** distinct `addFn` and
  **37** distinct `processFn`; the hf file is a strict superset at **74** distinct `addFn` and
  **46** distinct `processFn`. All 74 add names and all 46 process names resolve to real
  definitions in `Source/missiles.cpp` (74 `void Add...` definitions, 48 `void Process...`
  definitions; every table name matched). V3.2's "74 Add functions" is therefore the **hf
  superset** count, not the base count — the base registry is 51.
- `MissileID` (`Source/tables/spelldat.h:98-176`): 111 entries; 68 Diablo (`LastDiablo =
  DiabloApocalypse` idx 67), Hellfire band idx 69-108 ending `OrangeExplosion`, `Null = -1`.
- Sprite resolution: `MissileFileData::LoadGFX()` `misdat.cpp:400-430` — `animFAmt==1` single
  `missiles\<name>` CL2, else 8/16-direction sheet; `ToIndex` de-dup of delay/length arrays
  `misdat.cpp:47-55`; flags `ParseMissileGraphicsFlag` `:57+`.
- Motion/collision core: `AddMissile(...)` `missiles.cpp:278+`, `CheckMissileCol` `:479`,
  `MoveMissileAndCheckMissileCol` `:636` with 16.16 `position.traveled`, `Blockable` deflection
  `:648`, light radius at draw `:1338` (`DTYPE_NEST ? 9 : 3`), `addFn` dispatch `:2840` region,
  `ProcessMissiles` driver `:4198`, `ProcessRedPortal` `:4162-4180`.

## Scope
`internal/gamedata/misdat/`: `MissileData` (misdat.h:166-196), `MissileFileData` (sprite table),
`MissileID` enum in TSV order (111 entries incl. Null), `MissileGraphicsFlags`/
`MissileDataFlags`/`MissileMovementDistribution` enums, and
`Registry map[MissileID]struct{ Add AddFn; Process ProcessFn }`. Implementations port
`AddOpenNest` (`:1238`, 4 `BigExplosion` at (80/81, 62/63) then self-delete), `AddRune*`
(`:667`, `:1248-1271`), `AddReflect` `:1276`, `AddBerserk` `:1293`, `AddHorkSpawn` `:1344`,
`AddJester` `:1352`, `AddStealPotions` `:1390`, `AddStealMana` `:1456`, `AddSpectralArrow`
`:1477` (hf damage fix behind the flag), `AddWarp` `:1505`, `AddLightningWall` `:1580`,
`AddBigExplosion` `:1601`, `AddImmolation` `:1618`, `AddLightningBow` `:1634`, `AddMana`
`:1653`, `AddRedPortal` `:2773` — the runtime port itself stays in dg-task-022.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Misdat` = **68** base / **108** hf (+40 appended, 0
      modified), `MissileSprites` = **46** base / **59** hf (+13 appended); missile-domain total
      **114** base rows; 40 appended hf ids exactly the V12.10 list (Mana ... OrangeExplosion).
- [ ] Registry completeness: base build registers **51** distinct Add + **37** Process fns;
      hf build registers **74** + **46**; the test loads the TSV name sets and fails on any name
      without an implementation or any implementation not named by a row.
- [ ] Worked example: `AddOpenNest` spawns 4 `BigExplosion` missiles at (80,62) (81,62) (80,63)
      (81,63) and self-deletes — scripted test with a stub world.
- [ ] Sprite arrays: `frameDelay[16]`/`frameLength[16]` de-dup into shared tables with the same
      index assignment as `ToIndex`; assert distinct-array count for all 46 base rows.
- [ ] 16.16 fixed-point motion: `position.traveled` arithmetic property-tested against a C++
      probe for every distinct `movementDistribution` value (no float anywhere).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern, no reflection
- /home/blaze/.agents/skills/coding-rules/SKILL.md — table (data) vs Add/Process (behaviour)
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — property tests for fixed point
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — sprite sheet/animFAmt direction layout

## Dependencies & blockers
blocked_by dg-task-101. Blocks dg-task-111 (spelldat `missiles` column references these ids),
dg-task-104 (monster `GetMissileType` targets), dg-task-118 (audit), and existing dg-task-022
(missile engine), whose AC "misdat.tsv row count == registry size" this bean sharpens to the
51/37 base and 74/46 hf name-set counts.

## Open questions
- `MissileID` has 111 entries but misdat has 68 base rows and hf `Null=-1` — which ids are
  intentionally table-less (e.g. `ThinLightningControl` variants)? Enumerate before asserting.
- Do the two extra `Process*` definitions in `Source/missiles.cpp` (48 definitions vs 46 table
  names) matter? They look like internal helpers; confirm they are not registry entries.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; every measured claim re-verified true)
