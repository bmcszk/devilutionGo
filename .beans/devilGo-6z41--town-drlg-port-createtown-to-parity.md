---
# devilGo-6z41
title: 'Town DRLG: port CreateTown to parity'
status: completed
type: task
priority: normal
created_at: 2026-11-06T00:04:32Z
updated_at: 2026-11-06T00:22:30Z
parent: devilGo-97mn
blocking:
    - devilGo-yeia
---

Subtask of the DRLG family. Red gate landed by the dg-town s2 lane: the parity
test exists and fails against HEAD (RED is the contract for this bean).

## RED state at branch dg-town

- `internal/diablo/levels/town_parity_test.go` — `TestTownParity` (per-seed
  matrix) + `TestTownCorpus` (corpus size), diffing `dPiece`, `dSpecial` and the
  `dungeon` automap tile layer of the 8 committed
  `internal/testdata/drlg/town/l0-town-main-<seed>.csv` fixtures.
- `internal/diablo/levels/town.go` — WIP plumbing only: `TownSector1..4` /
  `TownAutomap` resource paths, `TownOptions`, and a `CreateTown` stub that
  returns a zeroed `PieceGrid`.
- Observed failure (all 8 seeds, from the committed corpus):
  `dPiece[0][0] = 0, want 217`.
- Second RED finding: the five `.dun` layers `CreateTown` loads
  (`town.cpp:202-207`) are **not committed**. `TestTownParity` reports
  `inputs-committed` missing fixtures:
  `internal/testdata/drlg/town/inputs/{sector1s,sector2s,sector3s,sector4s,automap}.csv`.

## Contract

Upstream (pinned `ca91d778f`, read-only reference): `Source/levels/town.cpp`
`:191-246` (`DrlgTPass3`: 426 fill, four `FillSector` at fixed origins,
`PlaceDunTiles(automap.dun)`, closed-warp patches incl. the ten
`PickRandomlyAmong({1,2,3,4})` draws, poisoned-water fountain), `:359-387`
(`CreateTown`: view bounds + entry `ViewPosition`), `:149-186`
(`InitTownPieces`: `dPiece` micro id -> `dSpecial` code). Surface map:
`docs/research/dg-town-inventory.md`.

Entry point is already fixed by the gate: `levels.CreateTown(d *Dungeon, opts
TownOptions) PieceGrid` — it writes `d.Tiles` (`dungeon`) and returns
`PieceGrid{Piece, Special}`. Inputs are injected (`TownOptions.Files`,
`MegaTiles`, `Warps`, `Hellfire`, `Quests`; RNG is `d.Rnd`, seeded by the caller).

## Acceptance criteria

- [x] Commit the five `.dun` inputs as setpieces-format CSVs (`width,height`
  then `width*height` uint16 tile values) under
  `internal/testdata/drlg/town/inputs/`, with a provenance note (retail
  `DIABDAT.MPQ` is not a legal test input — AGENTS.md zero-dependency rule)
- [x] Port `DrlgTPass3` + `CreateTown` + `InitTownPieces` 1:1; every ported
  function has at least one test
- [x] `TestTownParity` green on all 8 seeds for `dPiece`, `dSpecial` and
  `dungeon`; `dTransVal` stays all zero
- [x] gofmt -l empty, go vet clean, `./scripts/lint.sh` 0 issues,
  `go test ./internal/... -count=1` all ok

## Libraries

- stdlib only (no third-party deps at this stage)

## Triage 2026-11-06 (dg-town s2, disk-verified, delegated triage in lieu of a brainer pass)

Census before any claim (this session's brainer step ran disk-first, per the
wave recipe, because the town surface was already researched in
docs/research/dg-town-inventory.md):

- `grep -rn "func CreateTown" internal/` finds only the WIP stub in
  `internal/diablo/levels/town.go` — no town generator exists on disk.
- The corpus is 8 files x 5 records (`awk 'END{print NR}'` = 5; row 2/3 are
  12544 cells, row 5 is 1600), matching the inventory's array shapes.
- `ls internal/testdata/drlg/town/inputs` fails: the `.dun` input corpus is
  absent, so the gate currently fails before it can diff a real generator.
- `internal/testdata/drlg/assets/town.til.csv` is committed (376 mega tiles) and
  is the `pMegaTiles` table the gate already loads — the only town input that
  exists today.
- Scope is one generator, no retry loop, one RNG consumer: the smallest DRLG
  slice in the family.

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting

## IMPL NOTES (dg-town s3)

Bisect ledger, oldest signature first (each MOVE = progress):

1. `dPiece[0][0] = 0, want 217` (all 8 seeds) + `missing town .dun fixtures
   [.../town/inputs/{sector1s..4s,automap}.csv]` — RED at 8616252, stub
   CreateTown + absent input corpus.
2. `dungeon tile (15,35) = 47, want 7` (all 8 seeds) — first real divergence
   after the port: `dPiece` and `dSpecial` were already cell-exact, so the
   remaining mismatch sat in the hell-gate patch, which upstream writes as
   `dungeon[15..17][35]` and the port wrote transposed as `dungeon[35][15..17]`.
3. GREEN 8/8.

Corpus derivation was verified offline before the port: expanding sector4s ..
sector1s through `town.til` reproduced every committed dPiece cell except the 35
patch cells (caves 8, hell 20 + the x=46 spill of `FillTile(45,78,..)`, catacombs
4, poisoned water 2), and the automap layer plus its patch cells matched record 5.

Scope notes: `TownOpenHive/OpenGrave/CloseHive/CloseGrave` are ported as literal
cell tables (each coordinate carries its closed and open value) and covered by
`TestCreateTownHellfireDoors` driving `CreateTown` with `Hellfire: true` in both
door states, because every committed golden runs with Hellfire off.
`ENTRY_TWARPUP` still only moves `ViewPosition` via `TWarpFrom` game-flow state
that `TownOptions` does not carry; the two entries the fixtures drive are
`ENTRY_MAIN` and `ENTRY_PREV`. Renderer-side `.min`/`SetDungeonMicros` remains
deferred per the lane scope guard.

## Proof of Work

Commits on `dg-town`:

- `cdabfd7` town goldens: five fixed `.dun` inputs + shared tile-layer codec
  (`testsupport.EncodeDunTilesCSV`/`DecodeDunTilesCSV`, `dunconv town-inputs`)
- `21df9e2` town DRLG: port CreateTown, TestTownParity 8/8 green
- `add9ca9` town DRLG: cover the Hellfire hive/grave door patches

```
$ go test ./internal/diablo/levels/ -run 'TestTown' -v -count=1
=== RUN   TestTownParity
=== RUN   TestTownParity/seed-125121312
=== RUN   TestTownParity/seed-1383137027
=== RUN   TestTownParity/seed-1511478689
=== RUN   TestTownParity/seed-1522546307
=== RUN   TestTownParity/seed-19770182
=== RUN   TestTownParity/seed-2588
=== RUN   TestTownParity/seed-743271966
=== RUN   TestTownParity/seed-844660068
--- PASS: TestTownParity (0.01s)
--- PASS: TestTownCorpus (0.01s)
ok  github.com/bmcszk/devilutionGo/internal/diablo/levels
```

```
$ go build ./...            # (no output)
$ gofmt -l .                # (no output)
$ go vet ./internal/...     # (no output)
$ ./scripts/lint.sh         # 0 issues
$ go test ./... -count=1    # 43 lines "ok", 0 FAIL
$ python3 scripts/beans-lint.py   # beans-lint: OK (176 beans)
$ just sqlc-check           # no raw SQL outside sqlc-generated code
```

Residual: none for the 8 committed seeds. Follow-ups none opened; the deferred
renderer-side `.min` work stays on its existing bean.

## IMPL NOTE addendum (ec632f9)

The fountain predicate was reusing `Quests.PWATER`, documented as the level-gated `Quest::IsAvailable()` (`levels/tables/questdat.cpp:64-75`: false whenever `currlevel != _qlevel`), while `CreateTown` reads `_qactive != QUEST_DONE && != QUEST_NOTAVAIL` (`levels/town.cpp:240-244`). `Quests` gains `PWATERFountain` for the town predicate; `PWATER` stays the `IsAvailable()` result `drlg_l1` reads. `TestTownParity` pins `PWATERFountain: true` (the probe state), gates re-run: build/gofmt/vet clean, lint 0 issues, `go test ./...` 43 ok 0 FAIL.
