---
# devilGo-oqho
title: ItemPack/PlayerPack + SaveGameData/SaveLevel streams
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - save
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-6xu2
blocked_by:
    - devilGo-w67z
---

# Serializers

## Goal
All pack/unpack + game/level stream writers with exact endianness.

## Research
pack.h structs; loadsave.cpp:1503,1902-2900; MaxMissilesForSaveGame 125 ordering rules; padding to MaxMonsters.

## Scope
internal/persist/{pack,loadsave}.go with SaveHelper-equivalent (WriteLE/WriteBE/Skip).

## Acceptance criteria
- [ ] pack_test.cpp:441-868 all cases green
- [ ] 5 scripted states hex-equal SaveGameData
- [ ] LoadLevel regenerates identical StateHash (round-trip)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit byte order per field, no binary.Read shortcuts

## Dependencies & blockers
dg-task-034, dg-epic-04 structs.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Persistence epic serializers; blocked_by w67z (matches tree dg-task-035←034 — codec before serializers is right: packs get encrypted at pfile layer).
2 local: pack.h ItemPack struct verified at :20-31 (bean cites :20-30 ✓); loadsave.cpp:1503 SaveMonster verified; file 2947 lines so ":1902-2900" range valid; RETL/SHAR/HELF/SHLF magics verified :2308-2318 (load) and :2768-2770 (write); MaxMissilesForSaveGame=125 verified loadsave.cpp:55; pack_test.cpp 1459 lines, :441 TEST_F(PackTest, UnPackItem_diablo) verified (441-868 range = the pack test cluster).
3 online: n/a.
4 solution: none structural — structs dependency ("dg-epic-04" prose = world epic) is covered at epic level: 6xu2 IS blocked_by devilGo-97mn (verified), so world structs exist before any persistence task promotes. Cleanup: remap prose "dg-task-034, dg-epic-04" → w67z + 97mn.
5 understanding: persistence layer core; SaveHelper WriteLE/WriteBE port is the right primitive; hex-equal AC for 5 scripted states + StateHash round-trip are strong testable gates.
6 repro: `beans show devilGo-oqho`; `sed -n '20,31p' .../Source/pack.h`; `grep -n "MaxMissilesForSaveGame" .../Source/loadsave.cpp | head -2`; `sed -n '441,446p' ~/github/bmcszk/devilutionX/test/pack_test.cpp`; `beans show devilGo-6xu2 | grep -A2 blocked_by`.
7 fix: none required (optional prose id remap at promotion).
8 validation: n/a.
9 related: w67z, xwtt, 97mn children, 6xu2.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
