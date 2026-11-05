# Town DRLG inventory — `CreateTown` surface

Upstream: devilutionX @ `ca91d778fba92a1f680afb1fb2cdf09337ff1b42`.
Scope: what `CreateTown` reads, writes and depends on — the ground truth for the
town goldens (`internal/testdata/drlg/town/`) and for the later Go port.

## Entry points

| Symbol | Location | Role |
|---|---|---|
| `CreateDungeon(uint32_t rseed, lvl_entry)` | `Source/levels/gendung.cpp:355` | dispatch; `DTYPE_TOWN` → `CreateTown(entry)` at `:361` |
| `GetLevelType(0)` | `Source/levels/gendung.cpp:335-338` | returns `DTYPE_TOWN` |
| `CreateLevel(lvl_entry)` | `Source/diablo.cpp:1443` | `CreateDungeon(DungeonSeeds[currlevel], entry)` then `InitTownTriggers()` |
| `CreateTown(lvl_entry)` | `Source/levels/town.cpp:359-387` | sets view bounds, then `DrlgTPass3()` (`:386`) |
| `DrlgTPass3()` | `Source/levels/town.cpp:191-246` | the whole generator (no loops, no retries) |
| `FillSector` / `FillTile` | `town.cpp:26-58` / `town.cpp:66-79` | mega-tile writers |
| `InitTownPieces()` | `town.cpp:149-186` | writes `dSpecial` from `dPiece` values |

`CreateTown` does **not** receive `rseed`. It is not a generator in the
`CreateL5Dungeon` sense: one pass, no attempts, no rejection loop, no
`SetRndSeed` inside.

## Upstream files town loads

Generator reads (all inside `DrlgTPass3`):

| Path | `file:line` | Size (retail) | Dims (tiles) |
|---|---|---|---|
| `levels\towndata\sector1s.dun` | `town.cpp:202` | 21254 | 25×25 |
| `levels\towndata\sector2s.dun` | `town.cpp:203` | 19554 | 25×23 |
| `levels\towndata\sector3s.dun` | `town.cpp:204` | 19554 | 23×25 |
| `levels\towndata\sector4s.dun` | `town.cpp:205` | 13758 | 23×23 |
| `levels\towndata\automap.dun` | `town.cpp:207` | 54404 | 40×40 |

The four sector files tile the piece grid exactly: 23+23 = 46 and 25+25 = 50, so
the quadrants cover pieces `0..95` on both axes. `dPiece` cells `96..111` keep the
`426` fill (`town.cpp:195-198`) — 16 rows and 16 columns of solid rock that no
sector overwrites. `automap.dun` is the 40×40 tile layer (`dungeon`), placed by
`PlaceDunTiles(dunData, {0,0})` (`gendung.cpp:512`), the 2-arg overload that
touches no `dTransVal`.

The sector `.dun` files carry more than the tile layer: after
`2 + 2*w*h` bytes each holds further dPiece-scale layers (`LoadTransparency`
reads layer 4, `gendung.cpp:420-437`). Town never reads them — `FillSector`
uses only `dunData[2 ... 2+w*h]` (`town.cpp:30-31`).

Asset files the same level installs (outside the generator):

| Path | `file:line` | Note |
|---|---|---|
| `nlevels\towndata\town.til` → `levels\towndata\town.til` | `diablo.cpp:1380-1383` | `pMegaTiles`; 376 vs 342 entries |
| `nlevels\towndata\town.cel` → `levels\towndata\town.cel` | `diablo.cpp:1374-1378` | `pDungeonCels` |
| `levels\towndata\towns` (width 64) | `diablo.cpp:1386` | `pSpecialCels` (trees) |
| `nlevels\towndata\town.min` → `levels\towndata\town.min` | `dun_tile_data.cpp:73-77` | `SetDungeonMicros` input |
| `levels\towndata\own.pal` | `engine/palette.cpp:208` | `LoadRndLvlPal` town branch |
| `levels\towndata\automap.amp` | `automap.cpp:1528` | automap tile table |

`LoadLvlGFX` prefers the Hellfire `nlevels` variant whenever hellfire assets are
loaded and only falls back to `levels`. `FillSector`/`FillTile` index
`pMegaTiles[tileId-1]`, so **every town `dPiece` value depends on which of the
two `town.til` files is installed**. The committed goldens pin the `nlevels`
one (the probe loads the `hf` archives); a run without hellfire.mpq produces a
different `dPiece` for tile ids ≥ 342 only if the values differ, which is
exactly what the goldens would catch.

## Piece assembly order

1. `InitGlobals()` (`gendung.cpp:302-331`): zeroes `dFlags`, `dPlayer`,
   `dMonster`, `dCorpse`, `dItem`, `dObject`, `dSpecial`; fills `dLight` with
   `0` for town (`:311-316`); `DRLG_InitTrans()` zeroes `dTransVal` (`:384-389`);
   `dminPosition`/`dmaxPosition` reset.
2. `CreateTown` sets `dminPosition = {10,10}`, `dmaxPosition = {84,84}`
   (`town.cpp:361-362`) and `ViewPosition` per entry (`:364-384`).
3. `DrlgTPass3` fills the whole 112×112 `dPiece` with `426`, 2×2 blockwise
   (`town.cpp:193-200`).
4. `FillSector` × 4 at fixed origins: `(46,46)`, `(46,0)`, `(0,46)`, `(0,0)`
   (`town.cpp:202-205`). Per dun cell: `tileId = tile - 1`; `tileId >= 0` →
   write the mega tile's four micros as the 2×2 block, else all four cells get
   the `218` sentinel (`town.cpp:36-56`).
5. `automap.dun` → `dungeon[0..39][0..39]`, skipping zero tiles
   (`gendung.cpp:512-531`).
6. Closed-warp patches: catacombs (`town.cpp:210-214`), caves (`:215-219`),
   hell (`:220-227`) rewrite a few `dungeon` cells plus `FillTile` blocks, so
   the stair/entrance art matches the closed state.
7. Hellfire only: `TownOpenHive`/`TownCloseHive`, `TownOpenGrave`/
   `TownCloseGrave` (`town.cpp:228-238`) — direct `dPiece` writes on top.
8. Poisoned-water fountain: `FillTile(60, 70, 342)` while `Q_PWATER` is neither
   done nor unavailable, else tile `71` (`town.cpp:240-244`).
9. `InitTownPieces()` maps `dPiece` micro ids to `dSpecial` codes
   (`town.cpp:149-190`).

## Arrays town writes

| Array | Written by | Shape |
|---|---|---|
| `dPiece[112][112]` | steps 3-8 | complete (426 fill + sectors + patches) |
| `dSpecial[112][112]` | `InitTownPieces` only | sparse; `1,2,6..18` |
| `dungeon[40][40]` | `automap.dun` + patches | complete-ish (zero tiles skipped) |
| `dTransVal[112][112]` | nobody — zeroed by `DRLG_InitTrans` | all `0` |
| `dFlags[112][112]` | `InitGlobals` only | all `None` |
| `dLight[112][112]` | `InitGlobals` | all `0` (town is fully lit) |

`InitTownPieces` key → code pairs (`town.cpp:153-182`): `359→1`, `357→2`,
`128→6`, `129→7`, `127→8`, `116→9`, `156→10`, `157→11`, `155→12`, `161→13`,
`159→14`, `213→15`, `211→16`, `216→17`, `215→18`. Keys are **micro ids already
in `dPiece`**, not tile ids — a port must run this pass after the sector/patch
writes, matching `:245`. The renderer reads the codes as an arch index into
`towns` CELs (`scrollrt.cpp:121`, `:912`, `:939`), so a wrong code shows up as a
wrong tree/wall overlay rather than a wrong floor.

## How `CreateTown` differs from the cathedral floors

| Aspect | Cathedral (`CreateL5Dungeon`) | Town (`CreateTown`) |
|---|---|---|
| Seed input | `SetRndSeed(rseed)`, three layout attempts | none; `rseed` is dropped at `gendung.cpp:361` |
| Layout source | procedural (rooms, halls, minisets) | 5 fixed `.dun` files |
| Variation | seed-dependent | only the closed-warp patches |
| Retry loop | yes (`CreateL5Dungeon` loop) | no |
| Transparency | `DRLG_RectTrans`/`LoadTransparency` non-zero | `dTransVal` stays zero |
| Themes / monsters | `CreateThemeRooms`, `GetLevelMTypes` | neither (`diablo.cpp:3379-3381`) |
| Doors / specials | generator writes `dSpecial` | `InitTownPieces` derives it from `dPiece` |
| Triggers | `InitL1Triggers` | `InitTownTriggers`, static coords (`trigs.cpp:109-152`) |

No `TownSqs` (or any other data file) feeds doors or triggers: town triggers are
hard-coded positions gated by the same `IsWarpOpen` predicates the layout uses
(`trigs.cpp:115-150`), and they are installed by `CreateLevel` *after*
`CreateDungeon` returns — outside this unit's surface.

## RNG

Exactly one RNG consumer: `PickRandomlyAmong({1,2,3,4})` ten times inside
`for (x = 36; x < 46; x++) FillTile(x, 78, …)` (`town.cpp:224-226`), reached only
while `!IsWarpOpen(DTYPE_HELL)` — i.e. a hero below level 20 with no hell warp
flag, and never in multiplayer (`trigs.cpp:76-105`). Everything else is
deterministic.

Upstream never pins the seed at this call site: `SetRndSeedForDungeonLevel()`
(`diablo.cpp:3149-3157`) runs *after* `CreateLevel` (`diablo.cpp:3373`) in
`LoadGameLevelStandardLevel` (`diablo.cpp:3371-3377`), so
the ten draws consume whatever ambient state the caller left. The probe pins
`SetRndSeed(seed)` before `CreateDungeon` to make the dump a pure function of
`(entry, seed)`; the port must take the RNG as an explicit input for the same
reason.

`dSpecial` codes depend only on `dPiece`, so the ten random tiles are the only
seed-sensitive content in the whole town golden. Their reachable tile ids are
`{1,2,3,4}` (`town.cpp:225`), i.e. pieces from the first four `town.til` mega
tiles at `x 72..91, y 78..79`.

## Consumers inside this unit

- `PlaceDunTiles` also sets `Protected` (`gendung.cpp:520`), the automap
  "revealed" bitset — town marks all non-zero `automap.dun` cells.
- `dLight` all-zero means fully lit; `CreateTown` never touches lighting.
- Goldens must therefore assert `dPiece`, `dSpecial` and `dTransVal`; `dFlags`
  and `dLight` add no information beyond "InitGlobals ran".

## Golden set this analysis implies

- Seeds: `tools/probe/seeds.txt` (8), the corpus-wide convention — not the
  illustrative list a scoping note might carry. All eight dumps differ, and the
  difference is exactly the ten `x 36..45, y 78` tile columns.
- Entries: `ENTRY_MAIN` and `ENTRY_PREV` produce identical `dPiece`, `dSpecial`,
  `dTransVal` and `dungeon`; only `ViewPosition` moves (`town.cpp:364-384`).
  Eight files, not sixteen; the probe fails if the two ever diverge.
- Mega-tile table: `nlevels/towndata/town.til` (376 tiles) and
  `levels/towndata/town.til` (342 tiles) share the first 342 tiles byte for
  byte, so the goldens do not silently depend on whether the Hellfire archives
  are loaded. `assets/town.til.csv` pins the larger table because
  `GetTileCount(DTYPE_TOWN)` returns 376 (`dvg_probe.cpp`).
- Probe state: single player, fresh hero, `fullQuests`, no Hellfire quest
  progress — every warp is closed, so the catacombs/caves/hell patches and the
  RNG branch all run, and the hive/grave patches do not.
