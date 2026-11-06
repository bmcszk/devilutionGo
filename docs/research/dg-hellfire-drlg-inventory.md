# Hellfire DRLG inventory: hive and crypt generator surfaces

Scout for the Hellfire dungeon families. Pinned upstream: `/tmp/dvlx-pinned` at
`ca91d778f`. Every claim is a `file:line` read of that tree or of this worktree.

## Headline

There is no `Source/levels/drlg_l5.cpp` and no `Source/levels/drlg_l6.cpp` in
the pinned tree (`ls Source/levels/drlg_l*.cpp` lists only l1, l2, l3, l4).
Hellfire reuses the base generators:

| Levels | Level type | Generator | Entry point |
|---|---|---|---|
| 17-20 | `DTYPE_NEST` | caves, `drlg_l3` | `gendung.cpp:371-372` -> `CreateL3Dungeon` (`drlg_l3.cpp:2192-2200`) |
| 21-24 | `DTYPE_CRYPT` | cathedral, `drlg_l1` | `gendung.cpp:364-365` -> `CreateL5Dungeon` (`drlg_l1.cpp:1298-1313`) |

Level type comes from the level number at `gendung.cpp:335` (`GetLevelType`,
nest at levels 17-20 and crypt at 21-24).

**Both families call the same `DRLG_LPass3` seam**, so the shared
`ExpandPieces` port already covers their expansion step:

- hive: `CreateL3Dungeon` calls `Pass3()` (`drlg_l3.cpp:2198`), which is
  `DRLG_LPass3(8 - 1)` (`drlg_l3.cpp:2149-2152`). Identical to caves, `lv` = 7.
- crypt: `CreateL5Dungeon` calls `Pass3()` (`drlg_l1.cpp:1307`), which is
  `DRLG_LPass3(Dirt - 1)` (`drlg_l1.cpp:1239-1241`) with `Dirt = 22`
  (`drlg_l1.cpp:120`), so `lv` = 21. `LoadL1Dungeon` does the same at
  `drlg_l1.cpp:1335`.

`DRLG_LPass3` itself (`gendung.cpp:591-624`, expansion loop `:610-623`) expands `pMegaTiles[lv]` over the
40x40 buffer. Only the table differs per family (`.til`, not `.min`), so
`ExpandPieces` needs no new algorithm. New port work is layout branches and
data, not expansion.

## Crypt (levels 21-24)

`CreateL5Dungeon` runs the cathedral `GenerateLevel` then the crypt post steps
(`drlg_l1.cpp:1309-1312`): `PlaceCryptLights` and `SetCryptSetPieceRoom`.

Crypt branches inside the l1 generator:

- `drlg_l1.cpp:1019-1024` room choice: `SetCryptRoom` on level 24, else
  `SetCornerRoom`
- `drlg_l1.cpp:1159-1160` `PlaceCryptStairs`, `:1213-1214` `FixCryptDirtTiles`
- `drlg_l1.cpp:1220-1221` `CryptSubstitution`, `:1243-1244` `InitCryptPieces`
- `drlg_l1.cpp:1337-1339` custom map load: `AddCryptObjects`, `PlaceCryptLights`

`crypt.cpp` (840 lines) owns the crypt-only code: shadow patterns
(`crypt.cpp:488`), cracked/broken/leaking/substitution minisets
(`crypt.cpp:612-666`), `InitCryptPieces` (`:668`), `SetCryptRoom` (`:681`),
`SetCornerRoom` (`:697`), `FixCryptDirtTiles` (`:708`), `PlaceCryptStairs`
(`:726`), `CryptSubstitution` (`:751`), `SetCryptSetPieceRoom` (`:802`),
`PlaceCryptLights` (`:817-838`). Two `.dun` set pieces: `nlevels\l5data\uberroom.dun`
(`crypt.cpp:690`) and `nlevels\l5data\cornerstone.dun` (`crypt.cpp:701`).
`AddCryptObjects` lives in `objects.cpp:3793`.

## Hive (levels 17-20)

`CreateL3Dungeon` runs the caves `GenerateLevel`, then `Pass3` and `PlaceLights`
(`drlg_l3.cpp:2196-2199`). Nest branches inside the generator:

- `drlg_l3.cpp:1499-1501` pool: `PlaceSlimePool` (`:1421`) instead of lava
- `drlg_l3.cpp:1982-1989` stairs: `PlaceNestStairs` (`:1959`)
- `drlg_l3.cpp:2034-2042` isles: `L6ISLE1..L6ISLE5` (`:587`, `:600`, `:611`,
  `:624`, `:635`) at 70/70/30/30/100/100/90 percent plus one 1x1 miniset
- `drlg_l3.cpp:2182-2187` lights: `PlaceHiveLights` (`:2169-2178`, radius 9 on
  tiles 381-456). The hive has no crypt-style set piece room

## Assets each family loads

| Purpose | Hive (`l6data`) | Crypt (`l5data`) |
|---|---|---|
| Graphics | `nlevels\l6data\l6.cel` + `l6.til` (`diablo.cpp:1409-1413`) | `nlevels\l5data\l5.cel` + `l5.til` (`diablo.cpp:1414-1418`) |
| Stairs sheet | `levels\l1data\l1s` (`diablo.cpp:1413`) | `nlevels\l5data\l5s` (`diablo.cpp:1418`) |
| Mega tiles `pMegaTiles` | `l6.til` | `l5.til` |
| Micro pieces | `l6.min` (`dun_tile_data.cpp:88-89`) | `l5.min` (`dun_tile_data.cpp:90-91`) |
| Transparency | `l6.sol` (`dun_tile_data.cpp:154`) | `l5.sol` (`dun_tile_data.cpp:157`) |
| Palette | `nlevels\l6data\l6base{1..5}.pal`, alternate art shifts the pick (`palette.cpp:219-223`) | `nlevels\l5data\l5base.pal`, single (`palette.cpp:212-215`) |
| Set pieces | none | `uberroom.dun`, `cornerstone.dun` |

Hive uses `nlevels\` (new graphics) while caves use `levels\`, so a hive port
cannot reuse `l3.til`. Go coverage today:

- Crypt layout, post steps and set pieces: `internal/diablo/levels/crypt.go`
  (`uberroom.dun` at `crypt.go:190`, `cornerstone.dun` at `crypt.go:205`),
  gated by 64 dungeon-layer goldens (levels 21-24, two entries, eight seeds)
  through `crypt_parity_test.go:31`
- Hellfire palettes for both families: `internal/assets/palette.go:50-56`
- `ExpandPieces` scripted crypt coverage: `piece_test.go:32`, `:59`
- Crypt mega-tile index: `piece.go:71` maps `Crypt` to the l1 dirt index, which
  matches `Dirt - 1` = 21

Missing:

- No hive generator branch, no hive fixtures (`internal/testdata/drlg/` has no
  `l17`-`l20` CSV), no `l6.*` asset CSV
- No crypt `dPiece` golden: `internal/testdata/drlg/dpiece/` holds 256 fixtures
  for levels 1-16 only (`dpiece_parity_test.go:13-53`), so crypt Pass3 is
  scripted-tested but not parity-gated. `l5.til` is absent from
  `internal/testdata/drlg/assets/`, which stops at `l1..l4` plus `town`
- `piece.go:65-68` states "Town and Nest have no Pass3 generator". False for
  Nest: `drlg_l3.cpp:2149-2152` is unconditional and `CreateL3Dungeon` calls it
  for nest levels too. Nest should resolve to the caves index 7, not the
  default 0. Filed as an observation, not patched here

## Probe state

`tools/probe/patches/` touches `drlg_l2.cpp`, `drlg_l3.cpp`, `drlg_l4.cpp` and
`gendung.cpp` only (`+++ b/` lines). A hive run would still emit caves stage
dumps: the nest isle hook sits inside the shared `GenerateLevel`
(`devilutionx-ca91d778f-l3l4stages.patch`, before `drlg_l3.cpp:2034`) and dump
names key by `currlevel`. The crypt l1 path dumps too, so a crypt `dPiece`
golden needs only `l5.til`.

## Order suggestion

1. Crypt `dPiece` gate with `l5.til` at index 21, no new algorithm.
2. Hive layout port: the five nest branches in `drlg_l3.cpp` (stairs, pool,
   isles, lights, plus `l6.min`/`l6.sol`), then a `l6.til` gate at index 7. Fix
   the `piece.go` nest comment with it and add the data CSVs to the manifest.
