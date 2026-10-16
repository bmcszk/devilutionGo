---
# devilGo-778h
title: 'Monsters data package: monstdat+unique_monstdat (212 rows) + AiProc registry'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
---

# Monsters: 212 rows + AI family registry

## Goal
Transcribe `monstdat` (112) and `unique_monstdat` (100) into Go tables **and** the `AiProc[]`
function-pointer table as a Go registry keyed by `MonsterAIID`, with the exact null/non-null
shape of the C++.

## Research
- `/tmp/dg-requirements-append.md` **V2.1** (41 columns; loader `Source/tables/monstdat.cpp:409`;
  `MonsterData` monstdat.h:100-160), **V2.2** (100 unique rows; loader `:459`), **V2.3**
  (AiProc table), **V2.4** (pack/spawn rules), **V12.2** (enums), **V12.12** (counts).
- TSVs: `assets/txtdata/monsters/{monstdat,unique_monstdat}.tsv`,
  `mods/hf/txtdata/monsters/monstdat.tsv` (unique_monstdat has **no** hf override).
- Enums (`Source/tables/monstdat.h`): `MonsterAIID` `:22-63` = **40 named entries** +
  `Invalid = -1`; `MonsterClass` `:66-70` (Undead, Demon, Animal); `MonsterAvailability`
  `:92-95`; treasure bits `:86-88` (`T_MASK=0xFFF`, `T_NODROP=0x4000`, `T_UNIQ=0x8000`);
  `Uniq(...)` written by `monstdat.cpp:39`.
- **Correction to V2.3**: `Source/monster.cpp:3089-3130` `AiProc[]` has **40 entries**
  (`grep -c MonsterAIID` on that range = 40), 39 real function pointers plus **one `nullptr`**
  at `MonsterAIID::FireMan` (`:3107` — elemental behaviour lives in the missile code). The
  inventory's V2.3 table lists 41 rows because it counts the `Invalid = -1` sentinel as
  `AiProc[40]`; there is no index 40.
- Distinct `ai` values actually referenced: **24** in base monstdat, **31** in hf monstdat.
- Ranged missile selection `GetMissileType(MonsterAIID)` (`monster.cpp:1941-1972`) and
  `toHitSpecial` / `min/maxDamageSpecial` columns.

## Scope
`internal/gamedata/monstdat/`: `MonsterData` (all parsed columns, exact order),
`UniqueMonsterData`, `MonsterAIID`/`MonsterClass`/`MonsterAvailability` enums, treasure-bit
helpers, and `var AiRegistry = map[MonsterAIID]AIRuntime` where the `FireMan` key maps to a
**nil** implementation (so "no AI" is data, not an accident) and every registry entry has
`Name()` for diagnostics. Behaviour implementations themselves land in dg-task-105/epic-04;
this bean ships the registry *shape* + names.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Monstdat` = **112** base / **138** hf (+26 appended,
      0 modified), `UniqueMonstdat` = **100** base = hf (no override); monster-domain total
      **212** base rows; the 26 hf rows are exactly the V12.10 list (`MT_HELLBOAR` ...
      `MT_NAKRUL`).
- [ ] AI registry parity: `AiProc` has **40** entries with exactly **one** nil (`FireMan`), and
      every distinct `ai` value used by base+hf monstdat (31) resolves to a registry key; a test
      enumerates `MonsterAIID` and fails on any unmapped or extra entry.
- [ ] Worked example: the two monstdat rows carrying `Uniq(...)` treasure (e.g. `MT_CLEAVER`)
      decode to `T_UNIQ | id` and `dropsSpecialTreasure` is true; `T_NODROP` rows produce no
      loot — asserted from the table, not from generation code.
- [ ] hf diff content test: hf monstdat == base monstdat for the first 112 rows (0 modified) and
      appends 26, verified by a byte-level diff against both files.
- [ ] `GetMissileType` mapping table asserted for all documented pairs (GoatRanged->Arrow,
      Succubus/LazarusSuccubus->BloodStar, Acid/AcidUnique->Acid, FireBat->Firebolt,
      Torchant->Fireball, Lich->OrangeFlare, ArchLich->YellowFlare, Psychorb->BlueFlare,
      Necromorb->RedFlare, Magma->MagmaBall, Storm->ThinLightningControl,
      Diablo->DiabloApocalypse, BoneDemon->BlueFlare2, default->Arrow).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern replacing fn pointers
- /home/blaze/.agents/skills/coding-rules/SKILL.md — registry (data) vs AI (behaviour) separation
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-driven enum tests

## Dependencies & blockers
blocked_by dg-task-101. Blocks dg-task-105 (pack placement consumes these tables), dg-task-107
(missile ids referenced by monster AI must exist), dg-task-113 (monster voices -> audio),
dg-task-118 (audit), and existing dg-task-020 (population).

## Open questions
- Should `unique_monstdat` get an hf variant later? None exists on disk (100 = 100) — confirm
  Hellfire uniques are all already in the base table rather than missing data.
- `MonsterAvailability` Hellfire band naming in V12.2 ("(Hellfire band)") is vague; read
  monstdat.h:92-95 and pin the exact identifiers.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 3 declared blocks edges missing; row counts and AiProc shape verified exact.
