---
# devilGo-aix0
title: 'Spells data package: spelldat (36 rows) + mana/damage formulas'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
    - devilGo-pu3r
---

# Spells: 36 rows + formulas

## Goal
Transcribe `spelldat.tsv` with the prepended `Null` spell and port the mana-cost/damage/staff
charge formulas that read it.

## Research
- `/tmp/dg-requirements-append.md` **V7.1** (36 rows; loader `tables/spelldat.cpp:232-256`),
  **V7.2** (formulas), **V12.10** (hf deltas), **V12.12** (counts, "spelldat (+Null row) 37 /
  52"), **V12.9** (spawn skip lists).
- TSVs: `assets/txtdata/spells/spelldat.tsv` (36 data rows) and `mods/hf/txtdata/spells/spelldat.tsv`
  (**51** data rows: +15 appended, 2 modified — note V12.12's "52" = 51 + the prepended Null).
  Columns `id|name|soundId|bookCost10|staffCost10|manaCost|flags|bookLevel|staffLevel|
  minIntelligence|missiles|manaMultiplier|minMana|staffMin|staffMax`. First two ids are `Firebolt`,
  `Healing`; last base row `BoneSpirit`; last hf row `RuneOfStone`.
- **Null spell is prepended by code**, not present in the TSV: `AddNullSpell()`
  `spelldat.cpp:20-35` (all costs 0, flags Fire, `sMissiles` Null, `sStaffMin/Max = 40/80`), so
  `SpellsData.size() == 37` and `GetSpellData(id) == SpellsData[id]`.
- Structs/enums: `SpellData` (spelldat.h:225-252), `SpellID` (spelldat.h:30-87, Null ... RuneOfStone,
  `Invalid=-1`), `SpellType` (spelldat.h:20: Skill, Spell, Invalid, Book, Scroll, Staff, Charge),
  `SpellDataFlags` (spelldat.cpp:48-55: Fire, Lightning, Magic, Targeted, AllowedInTown).
  `missiles` -> `MissileID sMissiles[2]` with `Null` fill (`:250`); only `Flash` uses two
  (`FlashBottom, FlashTop`).
- Mana formula `GetManaAmount` (`Source/spells.cpp:104-140`): `sl = max(spellLevel-1, 0)`;
  `adj = sl * sManaAdj`; Firebolt halves `adj` (`:117-119`); Resurrect `adj = sl * (sManaCost/8)`
  (`:120-122`); Healing/HealOther `ma = Healing.sManaCost + 2*level - adj` (`:124`);
  `sManaCost == 255 => ma = (_pMaxManaBase >> 6) - adj` (`:125-126`); else `ma = sManaCost - adj`;
  `ma = max(ma,0) << 6` then class multiplier `ma = ma * GetClassAttributes(cls).manaCost >> 6`
  (`:136-138`) — the V10.3 `manaCost` column.
- Damage `CalculateSpellDamage` (`Source/spells.cpp`) feeding `AddMissile(..., midam, spllvl)`;
  spell levels in `pSplLvl[]` (37 wide in saves, `pack.h:33-88`); `LevelUpSpell`.
- Book/staff availability `GetSpellBookLevel`/`GetSpellStaffLevel` `spells.cpp:280-345` with the
  `gbIsSpawn` skip lists (StoneCurse/Guardian/Golem/Elemental/BloodStar/BoneSpirit, +Apocalypse
  for staff); `bookLevel/staffLevel = -1` => no book/staff exists. Scroll rows: 27 `IMISC_SCROLL`
  itemdat rows (dg-task-102); `IsReadiedSpellValid` `spells.cpp:26-45`; staff path
  `items.cpp:1236-1270` + `!gbIsHellfire && FlipCoin(4)` rule (`:1239`).
- 15 hf-only spells named in V12.10: Mana, Magi, Jester, LightningWall, Immolation, Warp,
  Reflect, Berserk, RingOfFire, Search, RuneOfFire, RuneOfLight, RuneOfNova, RuneOfImmolation,
  RuneOfStone.

## Scope
`internal/gamedata/spelldat/`: `SpellData` table (37 base incl. code-injected Null, 52 hf),
`SpellID`/`SpellType`/`SpellDataFlag` enums, `sMissiles[2]` with Null fill, mana/damage/staff
formula functions as pure table+RNG consumers, `AllowedInTown` gate set, book/staff level
tables with the spawn skip list, and the class `manaCost` multiplier hook (values from
dg-task-114).

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Spelldat` = **37** rows base (36 TSV + prepended `Null`)
      and **52** with hf (51 TSV + Null); +15 appended / 2 modified asserted by name against the
      V12.10 spell list; `SpellID` enum size matches the table size exactly.
- [ ] Mana-formula worked examples with `sManaAdj`/`sManaCost` from the table: Firebolt at
      spell level 3 halves `adj`; `sManaCost == 255` path yields `(_pMaxManaBase >> 6) - adj`
      then `max(ma,0) << 6` then `* manaCost >> 6`; Healing uses
      `Healing.sManaCost + 2*level - adj`; Resurrect uses `sl * (sManaCost/8)`. All integer,
      exact-shift asserted.
- [ ] `Null` row fidelity: index 0 has all costs 0, flags Fire, `sMissiles = {Null, Null}`,
      `staffMin/staffMax = 40/80` — matching `AddNullSpell()`, and `GetSpellData(id)` is a pure
      index (no map lookup).
- [ ] Only `Flash` has two missiles (`FlashBottom`, `FlashTop`): test asserts every other row's
      `sMissiles[1] == Null` across base and hf.
- [ ] Book/staff gating: for each spell, `bookLevel`/`staffLevel == -1` => no book/staff item;
      with `spawn` on, StoneCurse/Guardian/Golem/Elemental/BloodStar/BoneSpirit (+Apocalypse for
      staff) are excluded — asserted against the full 37-row table.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — int shift math, explicit fixed-point units
- /home/blaze/.agents/skills/coding-rules/SKILL.md — formulas read as the C++ step order
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-driven formula tests

## Dependencies & blockers
blocked_by dg-task-101. Needs dg-task-106 (`sMissiles` ids must exist) and dg-task-114
(class `manaCost` multiplier). Blocks dg-task-118 (audit) and existing dg-task-022 (missile
engine), dg-task-023 (staff/scroll item generation).

## Open questions
- V12.12 says spelldat hf = 52 while the file has 51 rows; the extra is the injected Null. Keep
  both numbers documented so the audit (dg-task-118) does not report a mismatch.
- Do the 2 modified hf spelldat rows change `manaCost` or only flags/levels? Diff them and record.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; all numeric claims disk-verified)


## IMPL NOTES (triage 2026-10-16)
1 root: spelldat transcription + mana/damage/staff formulas; = dg-task-111.
2 local: beans show ✓; spelldat.tsv 36 base / 51 hf rows ✓ EXACT; AddNullSpell spelldat.cpp:20-35 ✓ verbatim (costs 0, flags Fire, sMissiles Null×2, sStaffMin 40/sStaffMax 80 — the "prepended Null" claim is real code); GetManaAmount spells.cpp:104 ✓; GetSpellBookLevel :302 / GetSpellStaffLevel :325 ✓ (cited :280-345 covers); staff rule items.cpp:1239 `!gbIsHellfire && FlipCoin(4)` ✓ exact line; V7.1 @ :1621, V7.2 @ :1674 ✓; V12.12 "52" vs 51-row file reconciled in open questions ✓.
3 online: n/a
4 solution: remap deps (101→jx08; needs 106→vhgd, 114→pu3r) + edge; S2. Open questions are recorded and audit-safe.
5 understanding: diablo-layer spell tables + pure formulas; sMissiles ids constrained by vhgd; class manaCost multiplier from pu3r.
6 repro: beans show devilGo-aix0; wc -l devilutionX/assets/txtdata/spells/spelldat.tsv devilutionX/mods/hf/txtdata/spells/spelldat.tsv; sed -n '20,35p' devilutionX/Source/tables/spelldat.cpp; sed -n '1239p' devilutionX/Source/items.cpp
7 fix: beans update devilGo-aix0 --blocked-by devilGo-jx08; body-append id map.
8 validation: beans show lists blocked_by devilGo-jx08.
9 related: devilGo-jx08, devilGo-vhgd, devilGo-pu3r, devilGo-mwlm, devilGo-onlm
VERDICT: NEEDS-WORK:remap-deps (S1/S2 only; all numeric claims disk-verified)

Fix pass 2026-10-16: Dep-wiring applied (fix pass): declared edges wired; numeric claims all disk-verified (triage).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via /tmp/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.