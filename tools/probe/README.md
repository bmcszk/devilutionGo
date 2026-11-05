# dvg-probe — devilGo-vsx3 parity-fixture dumper

C++ probe that generates every parity fixture the devilutionGo port consumes,
byte-identically, from devilutionX pinned at commit
`ca91d778fba92a1f680afb1fb2cdf09337ff1b42`.

## Prerequisites

- devilutionX checkout at the pinned commit
- CMake ≥ 3.22, Ninja, a C++23 compiler
- System deps: SDL2, libsodium, libpng, libbz2, GTest (bundled Lua 5.4 is used;
  force `-DDEVILUTIONX_SYSTEM_LUA=OFF` if the system Lua is not 5.4)
- **Assets**: `devilutionx.mpq` next to the build output. The `dun`, `frames`
  and `save` subcommands additionally need `DIABDAT.MPQ` (or `spawn.mpq`) in
  the same directory at run time. MPQs are copyrighted and never committed.

## Build

From the devilutionGo repo root:

```sh
DVLX=~/github/bmcszk/devilutionX            # pinned at ca91d778f
git -C "$DVLX" apply tools/probe/patches/devilutionx-ca91d778f-dvg-probe.patch
cmake -S"$DVLX" -B"$DVLX/build-probe" -G Ninja \
    -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=ON \
    -DDEVILUTIONX_SYSTEM_LUA=OFF \
    -DDVG_PROBE_DIR="$PWD/tools/probe"
cmake --build "$DVLX/build-probe" --target dvg_probe -j$(nproc)
cp your-devilutionx.mpq "$DVLX/build-probe/"   # required
cp DIABDAT.MPQ "$DVLX/build-probe/"            # required for dun/frames/save
```

The patch only adds one `add_subdirectory("${DVG_PROBE_DIR}" dvg-probe)` line
to `CMake/Tests.cmake`; all probe sources stay in this directory. Revert with
`git -C "$DVLX" apply -R tools/probe/patches/...`.

### Stage-dump patches (diagnostics only)

Three patches add dump-only hooks to the generators; they edit no game logic
(every hunk is an added `ProbeDump*` call or the dumper itself):

| Patch | Hooks |
|---|---|
| `devilutionx-ca91d778f-l2layout.patch` | `drlg_l2.cpp` — full layout layer per attempt |
| `devilutionx-ca91d778f-l3l4stages.patch` | `drlg_l3.cpp` caves + `drlg_l4.cpp` hell — one `dungeon[40][40]` dump per generator stage |

Stage dumps are written only when a stage dir is set (the `l2layout` and
`stages` subcommands below) and are never committed.

## Subcommands

| Command | Produces | Notes |
|---|---|---|
| `dvg_probe rng <out> tools/probe/seeds.txt` | `rng/<seed>.bin` | 1000 raw `GenerateRandomNumber()` outputs per seed, LE u32 |
| `dvg_probe lighttables <out>` | `lighttables/{cathedral,hell,nest,crypt}_tables.bin`, `falloffs.bin`, `cone.bin` | all four `MakeLightTable()` leveltype branches; tables 16×256, falloffs 16×128, cone 8×8×16×16 |
| `dvg_probe dun <out> tools/probe/seeds.txt` | `dun/l<level>-<entry>-<seed>.dun` + `.meta` | seeded DRLG; diablo L1–L4 + hellfire L17–L20, main+prev entries; .dun layout = 2×u16 LE dims then `dungeon[40][40]` then the 80×80 `dTransVal` window; .meta = ViewPosition (2×u16 LE) + leveltype byte |
| `dvg_probe drlg <out> tools/probe/seeds.txt` | `drlg/l<level>-<entry>-<seed>.dun` + `.meta` | generator goldens only, dumped **before** the probe's `CreateThemeRooms()`; level matrix 1-4 cathedral (`drlg_l1.cpp`), 5-8 catacombs (`drlg_l2.cpp`), 9-12 caves (`drlg_l3.cpp`), 13-16 hell (`drlg_l4.cpp`), 21-24 Hellfire crypt; .dun layout identical to `dun`. Convert with `go run ./tools/probe/dunconv <out>/drlg internal/testdata/drlg` |
| `dvg_probe dpiece <out> tools/probe/seeds.txt` | `dpiece/l<level>-dpiece-<entry>-<seed>.bin` | the two tile-scale layers the generators produce: `dPiece` (mega-tile expansion, `DRLG_LPass3`, `gendung.cpp:591`) then `dSpecial`, both 112x112 x-major. Layout = 2x u16 LE dims, 112*112 u16 LE `dPiece`, 112*112 int8 `dSpecial`. Levels 1-16, both entries; the run installs each level's real `lN.til` through `LoadLvlGFX` (the `ProbeLoadLvlGFX` shim) first, because `DRLG_LPass3` expands whatever `pMegaTiles` holds. Convert with `go run ./tools/probe/dunconv dpiece <out>/dpiece internal/testdata/drlg/dpiece` |
| `dvg_probe stages <out> tools/probe/seeds.txt` | `<out>/<stage>-<level>-<seed>.bin` | per-attempt stage `dungeon[40][40]` dumps (levels 9-16, main entry); diagnostics for the port, not committed |
| `dvg_probe frames <out> <run-dir> <demo-file> [frames]` | `frames/frame_<n>.idx.bin` + `.pal.bin` | scripted timedemo run; `.idx.bin` = 2×u16 LE dims then row-major 8-bit back buffer; `.pal.bin` = 256 `SDL_Color` RGBA |
| `dvg_probe save <out> <run-dir> <demo-file>` | `save/single_0.sv` | scripted run ends with `pfile_write_hero(true)` |
| `dvg_probe manifest <out>` | `manifest.txt` | sha256 per artifact; determinism gate |

`<run-dir>` is scratch space (config path, demo copy, save output).

### Demo file

`demo_0.dmo` is `test/fixtures/timedemo/WarriorLevel1to2/demo_0.dmo` from the
pinned devilutionX tree with the trailing 14 bytes removed (the final
`MoveTo`-6 message and the `QuitEvent`). The probe removes the original two-byte `QuitEvent` and appends the same
pinned `QuitEvent` after the final rendering message, so the scripted loop
ends through the normal SDL quit path. The file is committed next to this
README as `demo_0.dmo`.

## Determinism contract

- All outputs depend only on `seeds.txt`, `demo_0.dmo`, and the pinned commit.
- `save` writes `PlayerPack.dwLowDateTime/dwHighDateTime` as zero
  (`pack.h:34-35`); the engine never sets them for new saves, so the pin holds.
- Two clean rebuilds + full rerun must yield identical `manifest.txt`
  (`tools/probe/out/manifest.txt` in the devilutionGo repo).
- Artifacts are written under `tools/probe/out/` in the devilutionGo repo
  (gitignored except the manifest); `test/fixtures/` is owned by devilGo-1spo.

## Golden refresher tooling (devkit)

Three in-repo programs turn probe artifacts or committed TSVs into committed
fixtures. ADR-019 classifies them as fixture refreshers: test instrumentation,
not conversion tools, and not for users (`docs/adrs/adr-019.md`, "Test
instrumentation is not part of the one program"). Nothing in `go test` or CI
runs them; a refresh ends with a normal test run that reads the committed bytes.

| Refresher | Reads | Writes | Gate |
|---|---|---|---|
| `tools/probe/dunconv` | `dvg_probe drlg` `.dun`+`.meta` pairs, or `dvg_probe dpiece` `.bin` dumps | `internal/testdata/drlg/*.csv`, `internal/testdata/drlg/dpiece/*.csv` | `internal/diablo/levels`: `TestL1CathedralParity`, `TestL2CatacombsParity`, `TestL3CavesParity`, `TestL4HellParity`, `TestCryptParity`, `TestDPieceParity` |
| `cmd/gamedatagoldens` | committed TSVs under `internal/testdata/gamedata` | `internal/testdata/gamedata/goldens/<table>-<variant>.json` | `internal/diablo/gamedata/audit`: `TestRowCountsMatchManifest`, `TestFingerprintsMatch` |

### dunconv

Run the patched probe first (see Build above), then convert both artifact sets:

```sh
DVLX=~/github/bmcszk/devilutionX
"$DVLX/build-probe/dvg_probe" drlg tools/probe/out/drlg tools/probe/seeds.txt
"$DVLX/build-probe/dvg_probe" dpiece tools/probe/out/dpiece tools/probe/seeds.txt

go run ./tools/probe/dunconv tools/probe/out/drlg internal/testdata/drlg
go run ./tools/probe/dunconv dpiece tools/probe/out/dpiece internal/testdata/drlg/dpiece
```

`dunconv` encodes through `testsupport.EncodeDunCSV`, so a refresh cannot drift
from the format the tests parse. The `dpiece` form writes into the `dpiece/`
subdirectory, next to the `.dun` corpus those layers expand. Both forms take
positional arguments only, `<artifact-dir> <corpus-dir>`, and fail on an empty
artifact dir rather than writing an empty corpus.

### gamedatagoldens

```sh
go run ./cmd/gamedatagoldens
```

It reads only the committed TSVs, so it never builds devilutionX, reads an MPQ
or touches the network. It rewrites one `parity.Fixture` JSON per table listed
in `targets` (`cmd/gamedatagoldens/main.go:26`), and each output names the TSV it
came from in `sourceTSV`. A table joins coverage by one line in `targets`.

The third refresher, `tools/probe/affix_probe`, dumps the `SelectAffix` pool into
`internal/testdata/gamedata/items/affix_pool_probe.tsv`; its build and dump steps
live in `tools/probe/affix_probe/README.md`.

All three refreshers overwrite files git tracks, so `git diff` shows exactly what a
refresh changed and the gate above either accepts the new bytes or names the row
that moved.

## Consumers

- `devilGo-m2x7` — rng vectors
- `devilGo-kx0l` — falloffs.bin / cone.bin
- `devilGo-xcdz` — dungeon dumps
- `devilGo-7aul` — frames pairs
- `devilGo-1spo` — storage + hash manifest
- `devilGo-6xu2`/`devilGo-oqho` — save roundtrip
