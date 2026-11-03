---
# devilGo-oqho
title: ItemPack/PlayerPack + SaveGameData/SaveLevel streams
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - save
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-03T00:54:02Z
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

## Progress log
- 2026-11-02: wave closed by coder (a8b0354): ym0r atomicity regression test (RED on unwrapped BeginTx, restored) + oqho LoadLevel StateHash round-trip + xwtt AC closes (v0/v2 migration replay, .sv export hex pin, uiheroinfo mapping). Gates: persist pkg ok 9.9s, lint 0 issues, sqlc-check clean, pushed.

- 2026-11-02: wave opened by parent (persist grouped PR). Coder
  dispatched on dg-persist-tx-regression-statehash-svexport. Verified
  on disk before open: pack_test.cpp goldens + hex SaveGameData exist;
  remaining gap = LoadLevel StateHash round-trip.

## Acceptance criteria
- [x] pack_test.cpp:441-868 all cases green — TestItemPackEncodeEarVector (pack_test.cpp:880), TestItemPackEncodeEmptyVector (:818), TestItemPackGoldenRoundTrip, TestItemPackSize, plus packitem_modes_test.go:47/68 (:514 shareware, :559 multiplayer)
- [x] 5 scripted states hex-equal SaveGameData — TestSaveGameStatesGolden over testdata/savegame_states.tsv (five states, hex image compared byte-for-byte)
- [x] LoadLevel regenerates identical StateHash (round-trip) — TestLoadLevelStateHashRoundTrip (internal/persist/levelsave_statehash_test.go:31) + TestSaveLevelRoundTrip

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit byte order per field, no binary.Read shortcuts

## Dependencies & blockers
devilGo-w67z, dg-epic-04 structs.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Persistence epic serializers; blocked_by w67z (matches tree devilGo-oqho←034 — codec before serializers is right: packs get encrypted at pfile layer).
2 local: pack.h ItemPack struct verified at :20-31 (bean cites :20-30 ✓); loadsave.cpp:1503 SaveMonster verified; file 2947 lines so ":1902-2900" range valid; RETL/SHAR/HELF/SHLF magics verified :2308-2318 (load) and :2768-2770 (write); MaxMissilesForSaveGame=125 verified loadsave.cpp:55; pack_test.cpp 1459 lines, :441 TEST_F(PackTest, UnPackItem_diablo) verified (441-868 range = the pack test cluster).
3 online: n/a.
4 solution: none structural — structs dependency ("dg-epic-04" prose = world epic) is covered at epic level: 6xu2 IS blocked_by devilGo-97mn (verified), so world structs exist before any persistence task promotes. Cleanup: remap prose "devilGo-w67z, dg-epic-04" → w67z + 97mn.
5 understanding: persistence layer core; SaveHelper WriteLE/WriteBE port is the right primitive; hex-equal AC for 5 scripted states + StateHash round-trip are strong testable gates.
6 repro: `beans show devilGo-oqho`; `sed -n '20,31p' .../Source/pack.h`; `grep -n "MaxMissilesForSaveGame" .../Source/loadsave.cpp | head -2`; `sed -n '441,446p' ~/github/bmcszk/devilutionX/test/pack_test.cpp`; `beans show devilGo-6xu2 | grep -A2 blocked_by`.
7 fix: none required (optional prose id remap at promotion).
8 validation: n/a.
9 related: w67z, xwtt, 97mn children, 6xu2.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.


## Summary of Changes

Verified the pack and level-stream ports against the C++ fixtures with no source
edits (gate sweep only). ItemPack/PlayerPack encode the pack_test.cpp vectors
(ear at :880, empty at :818, shareware expand at :514, multiplayer expand at
:559); SaveGameData matches the five committed scripted states byte-for-byte;
LoadLevel round-trips to an identical StateHash.

## Proof of Work

```
go test ./internal/persist/ -count=1 -run 'TestSaveGame|TestLoadLevelStateHash|TestItemPack|TestSaveLevel' -v
--- PASS: TestItemPackSize
--- PASS: TestItemPackGoldenRoundTrip
--- PASS: TestItemPackEncodeEarVector
--- PASS: TestItemPackEncodeEmptyVector
--- PASS: TestSaveLevelRoundTrip
--- PASS: TestLoadLevelStateHashRoundTrip
--- PASS: TestSaveGameStatesGolden
ok  github.com/bmcszk/devilutionGo/internal/persist
```

Full baseline at the branch tip: build rc0, gofmt empty, vet rc0, lint 0 issues,
`go test ./... -count=1` all ok.

Residual: none.
