---
# devilGo-baxb
title: Affix selection + drop-rule parity (V1.6/V1.7 worked examples as tests)
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-2j8j
---

# Affix generation and drop rules

## Goal
Port the item *generation* algorithms that consume the dg-task-102 tables: affix pool weighting,
prefix/suffix allocation with the affix lock, item power application, value calc, and the
monster-drop decision tree — with the inventory's worked examples as executable tests.

## Research
- `/tmp/dg-requirements-append.md` **V1.6** (affix selection, verified line by line) and
  **V1.7** (drop rules: magic/unique/no-drop), **V12.4** (drop tables by monster level),
  **V12.3** (pack rules cross-ref), **V12.7** (durability jitter).
- C++: `Source/items.cpp` — `SelectAffix` `:1063-1093` (type/minlvl/maxlvl/onlygood/good-evil/
  staff-charges filters, then the affix is appended `PLChance` times into
  `StaticVector<const PLStruct*,256>`, pick = `eligibleAffixes[GenerateRnd(size)]`, empty pool =>
  `nullopt`); `GetItemPowerPrefixAndSuffix` `:1179-1207` (`allocatePrefix = FlipCoin(4)` 25%,
  `allocateSuffix = !FlipCoin(3)` 66.67%, force one with `FlipCoin()`, `!onlygood &&
  !FlipCoin(3) => onlygood = true`, prefix drawn first and its `alignment` written into local
  `goe` at `:1195` **before** the suffix draw at `:1201` = the affix lock); `SaveItemAffix`
  `:1048-1061` (`multVal` -> `_iVMult2` when `param2` set else `_iVMult1`); `GetItemPower`
  `:1216-1224`; `GenerateMagicItemName` `:1103-1113`; `CalcItemValue` `:600-610`
  (`v = _iVMult1+_iVMult2; v>0 => v *= _ivalue; v<0 => v = _ivalue/v; _iIvalue =
  max(_iVAdd1+_iVAdd2+v, 1)`); `GetItemBonus` `:1309-1344` (`minlvl = min(minlvl,25)`);
  `GetStaffSpell`/staff path `:1236+` incl. vanilla-only `!gbIsHellfire && FlipCoin(4)` `:1239`.
- Drop path: `SpawnItem` `:3422-3481` (treasure bits `tables/monstdat.h:86-88`
  `T_MASK=0xFFF`, `T_NODROP=0x4000`, `T_UNIQ=0x8000`; SP special-treasure short-circuit to
  `SpawnUnique` `:3195`; mushroom brain override `:3435-3453`; `uper = isUnique ? 15 : 1` `:3464`);
  `RndItemForMonsterLevel` `:3256-3266` (`GenerateRnd(100) > 40 => IDI_NONE`, then
  `> 25 => IDI_GOLD`, else `GetItemIndexForDroppableItem(considerDropRate=true, ...)`
  `:1333-1375`); `GetItemBLevel` `:1480-1491`; `CheckUnique` `:1431-1453`
  (`GenerateRnd(100) > uper => invalid`, `GetValidUniques` `:1417-1429`,
  `validUniques[size-1-uidOffset]` = highest-lvl-eligible first); `TryRandomUniqueItem`
  `:3315-3360`; `RndAllItems` `:1392-1401`, `RndTypeItems` `:1403-1413`, `ItemRndDur`
  `:1475-1478`; quality tiers NORMAL/MAGIC/UNIQUE only (`items.h:41-45`).
- RNG dependency: dg-task-002 (`GenerateRnd`, `FlipCoin`, `RandomIntBetween` random.hpp:403).

## Scope
`internal/gamedata/itemgen/`: affix pool + weighting (materialise the duplicate-expanded vector,
keep tie order), prefix/suffix allocation + afflock, `SaveItemAffix`/`GetItemPower`,
`CalcItemValue`, `GetItemBonus` type dispatch, `SpawnItem` decision tree, unique roll.
Registry-free, pure functions over dg-task-102 tables + injected RNG.

## Acceptance criteria
- [ ] TDD first: V1.6 worked example is a failing test before implementation — Sword,
      `minlvl=1, maxlvl=6, onlygood=false, goe=GOE_ANY`, eligible set {Tin x2, Brass x2,
      sharpness x1, mending x1} => pool length **6**, `GenerateRnd(6)` index 0 => **Tin**
      (TOHIT_CURSE, v1=6 v2=10, multVal -3, alignment Evil); then `goe := GOE_EVIL` and the
      suffix draw must drop every `alignment=Good` suffix (and `mayhem`, lvl 7, by the
      minlvl>maxlvl filter).
- [ ] Row-count/weight parity: expansion factor test asserts `sum(PLChance)` over the eligible
      pool equals the C++ `StaticVector` size for 1000 sampled (type, minlvl, maxlvl, onlygood,
      goe) tuples — cross-checked against a C++ probe.
- [ ] V1.7 drop-rate parity: `RndItemForMonsterLevel` distribution over 1e6 seeds reproduces
      59.x% no-drop (`>40`), 37.5% of drops gold, remainder `dropRate`-weighted (weights only
      0/1/2 in the file) within chi-square p>0.01 of the C++ histogram.
- [ ] Magic/unique gates: `GetItemBLevel` roll order (rnd100<=10 OR rnd100<=lvl OR onlygood OR
      misc in {STAFF,RING,AMULET}, `uper15 => lvl+4`) and `CheckUnique` (`uper=15` => 15%
      unique-monster chance, `uper=1` => 1%) each have a deterministic seeded test;
      `DiscardRandomValues(1)` consumption asserted.
- [ ] `FlipCoin` denominators 4 / 3 / 2 pinned by test (prefix 25%, suffix 66.67%, onlygood 66.67%)
      and the prefix-before-suffix `goe` ordering asserted (an Evil prefix must never yield a
      Good suffix).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — injectable RNG, pure functions, no globals
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded tape + distribution tests
- /home/blaze/.agents/skills/coding-rules/SKILL.md — each C++ step reads as its own function

## Dependencies & blockers
blocked_by dg-task-102. Also needs dg-task-002 (RNG port) for `GenerateRnd`/`FlipCoin`
semantics. Blocks dg-task-108 (store stock calls `SpawnSmith`/`SpawnWitch`/`SpawnBoy`/
`SpawnPremium`) and existing dg-task-023.

## Open questions
- `StaticVector<const PLStruct*,256>` capacity: is 256 ever exceeded by the real tables
  (`sum(PLChance)` over a full pool)? Measure it; if it can overflow, the C++ behaviour is part
  of parity and must be replicated, not fixed.
- `dropRate` widened to 255 by commit `1412e25ef` but only 0/1/2 appear in the data — keep the
  general algorithm (recommended) or trim to observed values?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; otherwise exemplary — worked-example ACs verified against C++)
