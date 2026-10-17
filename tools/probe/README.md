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

## Subcommands

| Command | Produces | Notes |
|---|---|---|
| `dvg_probe rng <out> tools/probe/seeds.txt` | `rng/<seed>.bin` | 1000 raw `GenerateRandomNumber()` outputs per seed, LE u32 |
| `dvg_probe lighttables <out>` | `lighttables/{cathedral,hell,nest,crypt}_tables.bin`, `falloffs.bin`, `cone.bin` | all four `MakeLightTable()` leveltype branches; tables 16×256, falloffs 16×128, cone 8×8×16×16 |
| `dvg_probe dun <out> tools/probe/seeds.txt` | `dun/l<level>-<entry>-<seed>.dun` + `.meta` | seeded DRLG; diablo L1–L4 + hellfire L17–L20, main+prev entries; .dun layout = 2×u16 LE dims then `dungeon[40][40]` then the 80×80 `dTransVal` window; .meta = ViewPosition (2×u16 LE) + leveltype byte |
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

## Consumers

- `devilGo-m2x7` — rng vectors
- `devilGo-kx0l` — falloffs.bin / cone.bin
- `devilGo-xcdz` — dungeon dumps
- `devilGo-7aul` — frames pairs
- `devilGo-1spo` — storage + hash manifest
- `devilGo-6xu2`/`devilGo-oqho` — save roundtrip
