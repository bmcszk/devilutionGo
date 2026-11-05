---
# devilGo-t1lp
title: 'Town DRLG goldens: probe town subcommand + inventory doc'
status: in-progress
type: task
priority: high
created_at: 2026-11-05T23:48:01Z
updated_at: 2026-11-05T23:55:30Z
parent: devilGo-97mn
---

Lane dg-town s1 (goldens only, no Go port).

## Goal
RED-gate feedstock for CreateTown (town.cpp:359): committed CSV goldens plus a
linked inventory of the upstream town surface.

## Deliverables
- [ ] docs/research/dg-town-inventory.md
- [ ] dvg_probe town subcommand (dump-only, additive patch)
- [ ] internal/testdata/drlg/town/*.csv (8 seeds x main/prev)
- [ ] README correction: committed seeds are seeds.txt, not {0,1,2,77,...}
- [ ] gates + push

## Out of scope
Go port of CreateTown, parity test (later sessions).


## Triage (2026-11)

Disk-verified findings that change the contract:

1. Seed set: the lane contract named `{0,1,2,77,2588,5193,7777,13131333}` but
   the committed corpus convention is `tools/probe/seeds.txt` =
   `2588,743271966,1383137027,844660068,19770182,1522546307,125121312,1511478689`.
   Use seeds.txt (8 seeds) — the same set the 256 dPiece goldens use.
2. Town RNG: `CreateTown` (town.cpp:359) takes no rseed, and
   `SetRndSeedForDungeonLevel()` (diablo.cpp:3149) runs *after* `CreateLevel`
   (diablo.cpp:3392-3395), so at the CreateTown call site upstream never pins
   the RNG. The only RNG consumer is `PickRandomlyAmong({1,2,3,4})` in the
   `!IsWarpOpen(DTYPE_HELL)` branch (town.cpp:220-227, 10 calls). The probe must
   pin `SetRndSeed(seed)` itself; that pin is a probe decision, documented.
3. dTransVal is all-zero for town: `DRLG_InitTrans()` (gendung.cpp:385) memsets
   it inside `InitGlobals` and no town path writes it (town's `PlaceDunTiles`
   call at town.cpp:208 uses the 2-arg overload, which does not touch it).
   Dump it anyway as a zero-check gate.
4. Promoted to todo: scope is fully specified, no open questions.

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Progress log

- 2026-11-06 inventory: docs/research/dg-town-inventory.md (2aea670)
- 2026-11-06 probe `town` subcommand + `dunconv town` encoder; 8 dumps, entry-invariant
- 2026-11-06 8 CSVs under internal/testdata/drlg/town + assets/town.til.csv

## Proof of Work

(see the closing comment; gates re-run before completion)
