---
# devilGo-97mn
title: 'World & entities: grids, DRLG L1-L4/town/crypt/setmaps, player/monster/missile/object/item/towner'
status: todo
type: feature
priority: normal
tags:
    - plan:epic
    - phase-1
    - phase-2
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-29T15:20:51Z
parent: devilGo-i56j
blocked_by:
    - devilGo-5wsb
---

# World simulation core

## Goal
All world state structures and the deterministic generators + entity containers with tick processing, so gameplay (epic-05) has a living world.

## Research
- Grids & flags: Source/levels/dun_tile_data.hpp:26-120 (DungeonFlag, dPiece/dFlags/dPlayer/dMonster/dObject/dCorpse/dSpecial, SOLData[1379], MICROS 16 micros); DMAXX/Y 40, MAXDUN 112 gendung_defs.hpp:5-11; LevelCelBlock encoding dun_tile.hpp:83-104; TileType encodings :12-71.
- Generators: CreateL5Dungeon drlg_l1.cpp:1298-1312 (+crypt branches :1214,1221; lamps :1231-1234; PlaceMiniSetRandom :1249-1266; SelectChamber :1268-1294); L2 drlg_l2.cpp:1780-1786 RoomMin/Max + :2830 CreateL2Dungeon + :2839 LoadPreL2Dungeon; L3 :2185; L4 :1228; town.cpp warps :210-234; setmaps.cpp:112-165; themes.cpp:349,377-449; trigs; tile_properties.cpp:11-41.
- Entities: Player player.h:190-389 (+918 vector); Monster monster.h:181-310 (+38 MaxMonsters 200, AI enum tables/monstdat.h:22-58); Missile missiles.h:25-200 (+238 list); Object objects.h:32-80 (+30 127); Item items.h:189-230 (+25 127); Towner towners.h:48-63; corpses dCorpse encoding dun_tile_data.hpp:110-116.
- Animation: engine/animationinfo.h:112-131 (128 fractions), .cpp:37-115.
- Population: monster.cpp:566-603 AddMonsterType flags, :195 InitMonster, :282-292 Place; items.cpp:3422 SpawnItem, :3195 SpawnUnique; towners.cpp:760,811; player.cpp CalcPlrInv funnel :450-494, CreatePlayer :2300.
- Tests: drlg_l1..l4_test with seed fixtures; tile_properties; dead; townerdat.
- Study: docs/research/dg-requirements.md §2, §3, Appendices D.2/J.2.

## Scope
internal/world (grids, drlg_*, town, crypt, setmaps, themes, trigs, minisets) + internal/entity (structs, containers, tick processors, animation, position).

## Acceptance criteria
- [ ] Seed parity: every test/fixtures/diablo/*.dun and hellfire/*.dun reproduced byte-identically (dungeon grid; dPiece for L1 setmaps)
- [ ] Theme rooms + quest set-pieces match fixtures on 5 fixed seeds per level type
- [ ] Player walk: path via ported A*, 8-dir, interpolation progress exact (mock tests)
- [ ] Monster spawn lists per level seed match C++ dump for 3 seeds (type indices + positions)
- [ ] Missile motion: fixed-point 16.16 velocities, CheckMissileCol port passes missiles_test.cpp cases
- [ ] Animation: frame-to-tick distribution equals AnimationInfo algorithm on 30 cases
- [ ] dCorpse/dObject/dPlayer occupancy rules identical (unit tests)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — array-based state, deterministic iteration
/home/blaze/.agents/skills/coding-rules/SKILL.md — generator functions read like the C++ phases
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture-driven

## Dependencies & blockers
blocked_by dg-epic-02 (tables, sprites optional for logic). Blocks dg-epic-05, dg-epic-08 (level save/load), dg-epic-12.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: missile list = slice preserving C++ order (parity test on order), pool only if profiling demands later.
- Missile list order port strategy (pool vs slice) — decide in devilGo-r462 with order parity test.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: World state + deterministic generators + entity containers. Wired blockedBy 5wsb ✓ (body "dg-epic-02" = 5wsb via id-map), blocks nxxo/6xu2 ✓. 14 children.
2 local: Cites verified: gendung_defs.hpp:5-11 DMAXX/Y 40, MAXDUN 112 (=16+40*2+16) ✓; drlg_l1.cpp CreateL5Dungeon:1298 ✓, SelectChamber:1270 (body 1268-1294 ✓), PlaceMiniSetRandom:1251 ✓; questdat.hpp:14 MAXQUESTS 24 ✓; DungeonSeeds[15] quests.cpp:93 ✓.
3 online: n/a (no external claims).
4 solution: none — ready. dg-epic label cleanup optional.
5 understanding: game/world+entity layers; .dun fixture seed-parity AC is the exact 1:1 evidence form the mandate requires (byte-identical grids).
6 repro: `sed -n '5,11p' Source/levels/gendung_defs.hpp`; `grep -n "CreateL5Dungeon" Source/levels/drlg_l1.cpp`
7 fix: none.
8 validation: children (vti7, 76tu, 4ybb, yeia, foug, r462, 778h, v791, 1e1b, bodd, 9o2h, h6u0, pu3r, fpkz) triaged in batch-06.
9 related: blocker 5wsb; blocked nxxo, 6xu2; sibling qbiy (dun_render pairing).
VERDICT: TRIAGED-OK

## TRIAGE 2026-10-18 (pi wave, batch 2)

devilGo-97mn** (World & entities epic, parent i56j, draft)
1 Root: world state + DRLG L1-L4/town/crypt/setmaps + entity containers — IMPL NOTES consistent with body.
2 Local: YAML `blocked_by: [devilGo-5wsb]` ✓ ("dg-epic-02"). gendung_defs.hpp:5-11 DMAXX/Y=40, MAXDUN=16+80+16=112 ✓. CreateL5Dungeon:1298 ✓, PlaceMiniSetRandom:1251 (body 1249-1266 range ✓), SelectChamber:1270 (body 1268-1294 ✓). All 14 children exist ✓ (vti7…fpkz).
3 Online: SKIPPED — pure C++ source claims, all verified on disk.
4 Scope-fit: no dup — dungeon types split to fpkz, tile props to vti7, per-level to 76tu/4ybb/yeia; missile-list open question delegated to r462.
5 DDD: game/world + game/entity layers; grids feed render (qbiy) and gameplay (nxxo) — correct seams.
6 Dep-order: 5wsb→97mn→{nxxo,6xu2} chain wired in YAML of each ✓.
7 AC-gaps: 7 ACs all fixture/testable; missile slice-vs-pool question RESOLVED (recorded slice+order-parity) — no open gap.
8 Validation: `.dun` byte-identity run; monster spawn dump diff on 3 seeds; AnimationInfo 30-case port.
9 Related: blocker 5wsb; blocked nxxo/6xu2; sibling qbiy (dun_render pairing).
**TRIAGED-OK:devilGo-97mn**


Wave outputs: /tmp/dg-triage/out-b2.md

## Libraries

- Phase rule (DECIDED 2026-10-17): pure-Go only, no CGO. Phase-0 packages ran stdlib-only; concrete deps (modernc.org/sqlite, suprsokr/go-mpq, gopher-lua, ebiten) attach at the beans that consume them — never substituted by a worker.
