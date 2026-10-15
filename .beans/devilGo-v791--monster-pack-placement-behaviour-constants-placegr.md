---
# devilGo-v791
title: Monster pack placement + behaviour constants (PlaceGroup parity)
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-15T18:29:02Z
parent: devilGo-97mn
blocked_by:
    - devilGo-778h
---

# Pack placement and monster behaviour constants

## Goal
Port the placement/pack rules and the monster behaviour constants that sit outside the per-row
tables: `PlaceGroup`, `AddMonsterType`/`AddMonster` type lists, `PlaceUniqueMonst`, and the
shared AI primitives' constants.

## Research
- `/tmp/dg-requirements-append.md` **V2.4** (pack/spawn placement) and **V12.3** (pointer to it),
  **V2.3** shared primitives, **V12.9** (spawn gating for monsters).
- C++ `Source/monster.cpp`: `PlaceGroup(typeIndex, num, leader, leashed)` `:308-380` — outer
  `try1 < 10` full-group rewind (`:312-319`); anchor from leader `leader.tile +
  Direction(GenerateRnd(8))` (`:321-325`) else `do { x = 16+GenerateRnd(80); y = 16+
  GenerateRnd(80) } while (!CanPlaceMonster)` (`:326-331`, the 40x40 playable band); budget clamp
  `num + ActiveMonsterCount > totalmonsters` (`:333-335`); scatter up to 100 tries stepping
  `Displacement(Direction(GenerateRnd(8)))` in x **and the vanilla bug `yp +=
  Displacement(...).deltaX`** (`:338` — keep verbatim); gates `CanPlaceMonster`, same
  `dTransVal` as anchor (`:341`), leashed `|dx|<4 && |dy|<4` (`:342`); leader minions get
  `maxHitPoints *= 2`, `hitPoints = maxHitPoints`, `intelligence = leader->intelligence`
  (`:352-355`), `setLeader` when leashed (`:357-359`), extra handling when `ai != Gargoyle`
  (`:360`); `PlaceGroup` refuses to overwrite Gargoyle.
- `AddMonsterType`/`AddMonster` `:566-603`, `:1592` (`PLACE_*` flags);
  `IsMonsterAvailable` `:3159-3167` (rejects `availability == Retail` in spawn);
  `PlaceUniqueMonst(uniqindex, minionType, bosspacksize)` `:463` and `PlaceUniqueMonsters()`
  `:504-610` (`PlaceUniqueMonst(SkeletonKing, i, 30)` `:541`, Warlord/Lachdanan scatter
  `:566-569`, Lazarus+RedVex `:576-579`, BlackJade/Na-Krul hf branches `:580`, `:598`,
  `:601-607`); Diablo/quest gates `:3688`, `:3701-3704`, `:4087`.
- Shared AI primitives: `AiDelay` `:755`, `StartAttack` `:823`, `StartRangedAttack` `:832`,
  `RandomWalk` intelligence rule `v < 2*intelligence+33` (`+83` while mid-move) `:3050-3058`,
  `MonstCheckDoors`, `IsRelativeMoveOK` `:3132+`, `MFLAG_*` (`MFLAG_TARGETS_MONSTER` in
  `AiRanged` `:1984`), `checkStandAnimationIsLoaded` `:1938`, `:3063`.
- Containers: `MaxMonsters`/`ActiveMonsters` (`Source/monsters.h`), `dMonster` 4-tile trick.

## Scope
`internal/gamedata/monsterspawn/`: `PackRule`/`PlaceRule` tables per level type, `PlaceGroup`
with the RNG call sequence frozen, unique-monster placement table, behaviour constants
(`AI_DELAY_*`, `MINION_HP_FACTOR=2`, leash radius 4, scatter tries 100, rewind tries 10, band
16..96) as named ints, and the `yp += ...deltaX` bug behind a documented flag that defaults to
"on" (parity) with a recorded decision if ever turned off.

## Acceptance criteria
- [ ] TDD first: placement-sequence parity — for 3 fixed seeds x 4 level types, the Go
      `PlaceGroup` RNG call tape (order and count of `GenerateRnd` calls) equals the C++ probe
      output; group size and every tile coordinate match, including the 26 appended hf monster
      types being placeable.
- [ ] Worked example: leader-minion transform asserted exactly — `maxHitPoints *= 2`,
      `hitPoints = maxHitPoints`, `intelligence` copied from leader, leash radius `|dx|<4 &&
      |dy|<4`, and the same-`dTransVal` constraint rejecting cross-transparency placements.
- [ ] The `yp += Displacement(...).deltaX` vanilla bug is reproduced byte-for-byte in the tape
      test (a "fixed" Go version must fail this test).
- [ ] Budget clamp test: `num + ActiveMonsterCount > totalmonsters` truncates to
      `totalmonsters - ActiveMonsterCount` and never panics at the boundary.
- [ ] Spawn gating: with `spawn` flag on, no `availability == Retail` type is ever returned by
      `IsMonsterAvailable` (V12.9), tested over the full 112-row table.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — deterministic RNG call order discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — tape comparison vs C++ probe
- /home/blaze/.agents/skills/coding-rules/SKILL.md — constants table separated from placement loop

## Dependencies & blockers
blocked_by dg-task-104. Also needs dg-task-002 (RNG) and dg-task-015 (tile properties /
`CanPlaceMonster`, `dTransVal`). Blocks existing dg-task-020 (themes + population).

## Open questions
- Does `totalmonsters` come from a table or a per-level constant? V2.4 does not pin it — read
  `monster.cpp` before implementing the clamp.
- Where do `PLACE_*` flags live per level (hard-coded lists in `AddMonsterType`)? If they are not
  in a TSV, the inventory has no row list for them; confirm this is code, not data.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
