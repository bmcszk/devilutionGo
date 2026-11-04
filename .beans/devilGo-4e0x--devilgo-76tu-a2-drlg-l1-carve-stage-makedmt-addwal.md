---
# devilGo-4e0x
title: 'devilGo-76tu A2: DRLG L1 carve stage — MakeDmt, AddWall, FillChambers, FixTilesPatterns, stairs, GenerateLevel loop'
status: in-progress
type: task
priority: normal
created_at: 2026-11-04T04:08:20Z
updated_at: 2026-11-04T04:08:45Z
parent: devilGo-97mn
---

Subtask of devilGo-76tu (DRLG L1 cathedral + crypt). Parent state (verified): the room-layout stage (FirstRoom/GenerateRoom/CheckRoom/MapRoom/FindArea) is byte-identical to the C++ stage oracle on all 16 variants x 8 seeds (TestDevStage01FirstRoom green). `GenerateL1` in internal/diablo/levels/l1.go is still a stub, so TestL1CathedralParity is red at tile (0,0).

## Goal
Port the carve stage of the cathedral generator so the port state equals the C++ stage dumps `02-predstairs-aN` and `03-stairs` byte for byte, for every variant/seed/attempt, and wire `GenerateL1` to run that far.

## Upstream reference (READ-ONLY)
/tmp/dvlx-pinned = devilutionX @ ca91d778f. Re-clone if pruned. Canonical: Source/levels/drlg_l1.cpp.

C++ line map to port (1:1, cite `file:line` in comments):
- :28-95    inline Miniset tables: STAIRSUP (:39), L5STAIRSUP (:47), STAIRSDOWN (:55), LAMPS (:69), PWATERIN (:81) — hardcoded upstream data, port as Go values
- :430-447  InitSetPiece
- :448-472  InitDungeonPieces (port only if a caller in the fixture path needs it — GenerateLevel does not call it; document if skipped)
- :473-479  InitDungeonFlags (memset dungeon to Dirt, reset Protected/Chamber)
- :607-628  MakeDmt (note the DMAXX-1/DMAXY-1 bounds: last row/col stay Dirt)
- :629-750  HorizontalWallOk / VerticalWallOk / HorizontalWall / VerticalWall (RNG: GenerateRnd(4), GenerateRnd(6), GenerateRnd(maxN-1)+1)
- :751-803  AddWall
- :804-874  GenerateChamber / GenerateHall
- :875-995  FixTilesPatterns (two passes, BUGFIX bounds checks)
- :1036-1075 FillChambers (calls GenerateChamber/GenerateHall, then InitSetPiece at :1072)
- :1155-1202 PlaceCathedralStairs (PlaceStairs at :1203 dispatches on leveltype; crypt deferred to unit C)
- :1212-1298 GenerateLevel: the `do { LevelSeeds[currlevel] = GetLCGEngineState(); FirstRoom(); } while (FindArea() < minarea)` retry loop plus `DRLG_InitTrans(); attempt++`, then InitDungeonFlags/MakeDmt/FillChambers/FixTilesPatterns/AddWall/FloodTransparencyValues(13), then PlaceStairs -> retry if it fails, then the EntranceStairs trans-copy loop, then stop (the 04.. passes belong to unit B).
- :1330-1356 SelectChamber

## Requirements
- Style matches the landed package: unexported methods on *Dungeon (see l1_rooms.go). `d.Tiles`, `d.Mask`, `d.Chamber`, `d.Protected`, `d.Rnd.{GenerateRnd,FlipCoin}`, `d.InitTrans`, `d.FloodTransparencyValues`, `d.PlaceMiniSet`, `d.PlaceDunTiles`, `GetDunSize`, `CanReplaceTile` already exist — reuse, do not re-port.
- `GenerateL1(d *Dungeon, opts Options) error` becomes the real body: minarea by level (1:533, 2:693, default 761), the retry loop, and the pass order above. Options.Quests gates InitSetPiece/PlaceCathedralStairs; Options.SetPieces replaces LoadFileInMem. A `LevelSeed` state field models `LevelSeeds[currlevel]` (`if LevelSeed != 0 { re-seed }`) — the probe resets it per variant, so document that.
- SelectChamber AC (binding): table-driven tests covering ALL 8 availability combinations (HasChamber1 x HasChamber2 x HasChamber3) x VerticalLayout true/false, pinning the returned WorldTilePosition and the RNG draw sequence for each.
- Every ported function gets at least one test. Table-driven with t.Run.
- No new deps, no go.mod edits, no fixture edits, no .beans writes, no git state changes.

## Acceptance criteria
- [ ] AC 1: Extend the gitignored dev oracle harness `internal/diablo/levels/zzdev_stage_test.go` (package levels, skip when /tmp/drlg-stages absent) with diffs for stages `02-predstairs-aN` and `03-stairs`: tiles, trans window, mask, chamber, protected, view, setpiece — for all 16 variants (levels 1-4, 21-24 x main/prev) x 8 seeds x every dumped attempt. This test must be GREEN.
- [ ] AC 2: `gofmt -l internal/diablo/levels` empty; `go vet ./internal/...` rc0; `./scripts/lint.sh` 0 issues; `go build ./...`.
- [ ] AC 3: `go test ./internal/diablo/levels/ -count=1 -run TestL1CathedralParity` is still expected RED (units B/C remaining) — paste the first failing cell in the report.
- [ ] AC 4: Report: files touched, C++ line citations, gate output tails, and the exact oracle diff for any stage cell you could not make green.

## Stop rules
Do not touch fixtures to make tests pass. Do not stub passes to fake green. If the oracle diff cannot be closed after sustained debugging, report the concrete cell (variant/seed/attempt, coordinate, got/want) instead of claiming success.

## Triage (brainer, 2026-11-04)
Delegated slice of a parent already triaged TRIAGED-OK (devilGo-76tu): the
upstream symbols, fixture path and gate are all cited in the body above and were
verified on disk (`Source/levels/drlg_l1.cpp` @ ca91d778f, committed corpus under
`internal/testdata/drlg/`, oracle dumps under `/tmp/drlg-stages`). No open
question blocks work; no new dependency is introduced.

## Libraries
- stdlib only (no third-party deps; no go.mod edits)

## Skills for coder
- `/home/blaze/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `/home/blaze/.agents/skills/go-dev/SKILL.md` — RNG-call-order discipline, 1:1 port rules
- `/home/blaze/.agents/skills/go-tests/SKILL.md` — table-driven tests, fixture matrix

## Progress log
- 2026-11-04 (session 5): claimed; implementing MakeDmt/AddWall/GenerateLevel carve stage in
  internal/diablo/levels/l1.go, oracle stages 02-predstairs/03-stairs as the gate.
