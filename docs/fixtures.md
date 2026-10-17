# Parity fixture production

`tools/probe/dvg_probe` is the C++ oracle for parity fixtures. It is built
against devilutionX commit `ca91d778fba92a1f680afb1fb2cdf09337ff1b42` using
the patch in `tools/probe/patches/`.

| Artifact | Producer | Consumer |
|---|---|---|
| `rng/<seed>.bin` | `dvg_probe rng` | devilGo-m2x7 |
| `lighttables/*_tables.bin` | `dvg_probe lighttables` | lighting parity tests |
| `lighttables/falloffs.bin`, `cone.bin` | `dvg_probe lighttables` | devilGo-kx0l |
| `dun/*.dun`, `*.meta` | `dvg_probe dun` | devilGo-xcdz |
| `frames/*.idx.bin`, `*.pal.bin` | `dvg_probe frames` | devilGo-7aul |
| `save/spawn_0.sv` | `dvg_probe save` | devilGo-6xu2 / devilGo-oqho |
| `manifest.txt` | `dvg_probe manifest` | devilGo-1spo |

The producer writes only to its requested output directory; it never writes to
`test/fixtures/`. `tools/probe/out/` is the conventional ignored output
location. The pinned seed list and demo input are committed under
`tools/probe/`.

## Format notes

- RNG files contain 1000 little-endian `uint32_t` values per seed.
- Light tables are the native contiguous C++ arrays: 16×256 table bytes,
  16×128 falloff bytes, and 8×8×16×16 cone bytes.
- Dungeon files contain two little-endian `uint16_t` dimensions, the 40×40
  dungeon tile array, then the 80×80 transparency window. The `.meta` sidecar
  contains the little-endian view position followed by one level-type byte.
- Frame index files contain two little-endian dimensions followed by row-major
  8-bit pixels. Palette files contain 256 `SDL_Color` values (RGBA).
- Saves are vanilla spawn `.sv` archives. New-player `PlayerPack` timestamp
  fields `dwLowDateTime` and `dwHighDateTime` from `pack.h:34-35` are zeroed by
  the engine's `PackPlayer` initialization.

## Reproduction and determinism

Apply the patch to a clean checkout of the pinned devilutionX commit, configure
with `BUILD_TESTING=ON` and `-DDVG_PROBE_DIR=$PWD/tools/probe`, then build the
`dvg_probe` target. Run each subcommand twice from clean output/run directories
and compare `sha256sum` output or the generated manifest. MPQ assets are local
copyrighted inputs and are not committed.
