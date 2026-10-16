---
# devilGo-v791
title: Monster pack placement + behaviour constants (PlaceGroup parity)
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-97mn
blocked_by:
    - devilGo-778h
    - devilGo-vti7
---

# Pack placement and monster behaviour constants

## Goal
Port the placement/pack rules and the monster behaviour constants that sit outside the per-row
tables: `PlaceGroup`, `AddMonsterType`/`AddMonster` type lists, `PlaceUniqueMonst`, and the
shared AI primitives' constants.

## Research
- `docs/research/dg-requirements-append.md` **V2.4** (pack/spawn placement) and **V12.3** (pointer to it),
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — deterministic RNG call order discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — tape comparison vs C++ probe
- /home/blaze/.agents/skills/coding-rules/SKILL.md — constants table separated from placement loop

## Dependencies & blockers
blocked_by devilGo-778h. Also needs devilGo-m2x7 (RNG) and devilGo-vti7 (tile properties /
`CanPlaceMonster`, `dTransVal`). Blocks existing devilGo-foug (themes + population).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: totalmonsters source (table vs per-level constant) + PLACE_* flag lists read from monster.cpp as step 1, before clamp implementation.
- Does `totalmonsters` come from a table or a per-level constant? V2.4 does not pin it — read
  `monster.cpp` before implementing the clamp.
- Where do `PLACE_*` flags live per level (hard-coded lists in `AddMonsterType`)? If they are not
  in a TSV, the inventory has no row list for them; confirm this is code, not data.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 2 missing edges (vti7, foug); content verified strong incl. the vanilla-bug AC.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 task (devilGo-v791): PlaceGroup/AddMonsterType/PlaceUniqueMonst + AI constants with frozen RNG tape; draft; wire blockedBy [778h] ✓ = body's devilGo-778h.
2 local: STRONG verification: PlaceGroup decl monster.cpp:308 ✓; try1<10 rewind :312-315 ✓; budget clamp verified at :336-337 (body ":333-335", ~3-line drift); the vanilla `yp += ...deltaX` bug verified VERBATIM at :343 with `/// BUGFIX: `yp += Point.y`` comment ✓ (body ":338"); leader-minion `maxHitPoints *= 2; hitPoints = maxHitPoints; intelligence = leader->intelligence` verified :352-355 ✓ exact; leash `>= 4` rejects ✓ (:344); RandomWalk `v < 2*intelligence+33` / `+83` :3050-3058 ✓ exact; IsMonsterAvailable :3154 rejects `gbIsSpawn && Retail` ✓ (body ":3159-3167", 5-line drift); PlaceUniqueMonst decl :463 ✓, SkeletonKing `i, 30` :541 ✓; checkStandAnimationIsLoaded :1937 (body :1938 ✓). OQ "totalmonsters?" answerable: global `size_t totalmonsters` monster.cpp:132 set during level gen — code, not table (fine to leave as coder note). Wire gap: body "needs devilGo-m2x7 (m2x7), devilGo-vti7 (vti7)" and "blocks devilGo-foug (foug)" — m2x7/vti7 not in wire (not transitive via 778h→jx08 chain), foug has no v791 edge.
3 online: n/a.
4 solution: wire vti7 (brings m2x7 transitively) and foug; refresh 3 drifted line cites (optional).
5 understanding: behaviour-constants + placement layer between 778h tables and foug population; the RNG-call-tape AC incl. reproducing the yp bug is exactly the 1:1 mandate done right.
6 repro: `sed -n '308,315p;336,345p;350,356p' devilutionX/Source/monster.cpp; sed -n '3050,3058p' devilutionX/Source/monster.cpp; grep -n "size_t totalmonsters" devilutionX/Source/monster.cpp`
7 fix: `beans update devilGo-v791 --blocked-by devilGo-vti7`; `beans update devilGo-foug --blocked-by devilGo-v791`; optional cite refreshes (`:333-335`→`:336-337`, `:338`→`:343`, `:3159-3167`→`:3154-3167`).
8 validation: v791 blockedByIds includes vti7; foug's includes v791.
9 related: 778h (dep), foug (dependent), vti7/m2x7 (needed), 9o2h (dTransVal consumer).
VERDICT: NEEDS-WORK:wire-deps — 2 missing edges (vti7, foug); content verified strong incl. the vanilla-bug AC.

Fix pass 2026-10-16: Dep-wiring applied (fix pass): vti7 + foug edges wired.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).


