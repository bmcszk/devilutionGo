# DRLG L1 cathedral + Hellfire crypt goldens (devilGo-76tu)

Committed, Go-owned parity fixtures for `CreateL5Dungeon` (`drlg_l1.cpp:1299-1345`)
on the two level types it serves: `DTYPE_CATHEDRAL` (level 1) and `DTYPE_CRYPT`
(Hellfire levels 21-24).

Tests read only these files. They never read MPQs, the devilutionX tree, or the
network; a missing or malformed fixture fails the test.

## Corpus

`l<level>-<entry>-<seed>.csv`, 16 variants × 8 pinned seeds = 128 files:

| level | leveltype | source |
|---|---|---|
| 1-4 | `DTYPE_CATHEDRAL` (1) | `drlg_l1.cpp` cathedral branch |
| 21-24 | `DTYPE_CRYPT` (6) | `crypt.cpp` variant |

Levels 2-4 are included because they are the cathedral levels (`CreateL5Dungeon`
serves all of `DTYPE_CATHEDRAL`): 2 exercises the Butcher set piece, 3 the
Skeleton King set piece, 4 the banner set piece plus the Poisoned Water Supply
entrance and its stairs rule.

Each CSV holds four records (the format `internal/diablo/testsupport.EncodeDunCSV`
writes): the two tile-layer dimensions, the row-major 40×40 `dungeon` tile ids,
the row-major 80×80 `dTransVal` window (rows 16..95, columns 16..95 of the
112×112 array), and `ViewX,ViewY,leveltype`.

## Set-piece data

`setpieces/<name>.csv` is the dimension header plus the tile layer of the `.dun`
files the generator loads by name (`drlg_l1.cpp:386-394`, `crypt.cpp:690,701`):
`rnd6`/`skngdo`/`banner2` from `DIABDAT.MPQ levels/l1data`, `cornerstone`/
`uberroom` from devilutionX `mods/hf` (`assets/hf/nlevels/l5data`).

## Producer

Offline, once, against the pinned commit `ca91d778f`:

1. `git -C devilutionX worktree add /tmp/dvlx-pinned ca91d778f`
2. add `add_subdirectory(${DVG_PROBE_DIR} dvg-probe)` to `CMake/Tests.cmake`,
   configure with `-DBUILD_TESTING=ON -DDVG_PROBE_DIR=<repo>/tools/probe`, build
   the `dvg_probe` target, copy `DIABDAT.MPQ`, `hellfire.mpq` and
   `devilutionx.mpq` next to it.
3. `dvg_probe drlg <out> tools/probe/seeds.txt` — one `.dun` + `.meta` per
   variant and seed, dumped **before** `CreateThemeRooms()` (theme rooms are
   devilGo-foug's scope, so these goldens pin only the generator).
4. Convert each pair to CSV (dims, tiles, transparency window, view position,
   leveltype). `manifest.txt` carries the sha256 of every committed file.

The same build reproduces all 256 committed `dun/` artifacts byte-identically
(manifest sha256 check), which is the provenance check for the producer.

## Set-piece format (fixed 2026-11-04)

A set-piece CSV holds the `.dun` tile layer as **uint16 tile values**, one
decimal per cell, after a `width,height` first record. `encodeDun` in
`l1_parity_test.go` rebuilds the file as `uint16 w, uint16 h, w*h uint16 LE`,
which is what `GetDunSize`/`PlaceDunTiles` read.

The five CSVs first committed with this corpus were wrong: they held the raw
**byte** stream of the tile layer, so every value looked like `4,0,7,0,...` and
`encodeDun` rebuilt a layer whose tiles were `4,0,7,0,...` instead of
`4,7,7,4,...` — a set piece the C++ generator never placed. The goldens in this
directory are unaffected (they come from the real `.dun` inside the C++ run),
so the defect only showed up as a cathedral parity mismatch on the levels that
use a set piece (2, 3, 4).

All six CSVs are now regenerated from the pinned assets and verified to
round-trip: `encodeDun` output equals the first `4+2*w*h` bytes of the real
`.dun` byte for byte (`cornerstone`, `uberroom` checked explicitly;
`rnd6`/`skngdo`/`banner2`/`banner1` read from `DIABDAT.MPQ` with the ported Go
MPQ reader).

`banner1` (`levels/l1data/banner1.dun`) is added because
`DRLG_CheckQuests`→`DrawLTBanner` (levels/drlg_quests.cpp:56-80) loads it on
level 4. It writes `pdungeon`, not `dungeon`, so it does not change the tile
goldens.
