# Committed render goldens (devilGo-66o7 tier 1)

Byte-exact reference data for the pure-Go palette-lookup core in
`internal/render`. No Go test reads a devilutionX checkout, a retail MPQ, or a
`DG_*` variable: the tests read these files and **fail** (`t.Fatalf`) when one is
missing or short.

## Provenance

- Source tree: `/tmp/dvlx-pinned`, commit `ca91d778fba92a1f680afb1fb2cdf09337ff1b42`.
- Palettes: `devilutionx.mpq` / `DIABDAT.MPQ` retail records read through
  `internal/assets.MPQStack`, dumped once offline (the Go stack is the reader, so
  no second decoder is in the path).
- `lut/`, `blit/`, `cycle/`: produced by the offline C++ producer kept in
  `producer/oracle.cpp`, which `#include`s the pinned
  `utils/palette_blending.cpp`, `utils/palette_kd_tree.cpp` and
  `engine/render/blit_impl.hpp` **unmodified** — the algorithm under test is the
  upstream algorithm, not a transcription.

## Layout

| Path | Contents |
|---|---|
| `pal/<level>.pal.hex` | 768 bytes, hex, 96 chars per line — the palette as loaded |
| `lut/<level>.txt` | 65536 bytes, hex, 512 chars per line — the full `paletteTransparencyLookup[256][256]` row-major |
| `lut/diablo.txt` | same shape, for `ui_art/diablo.pal` with the full-range blend |
| `lut/variants_diablo.txt` | 12288 bytes, hex — 16 light-shade rows of 256 RGB triples |
| `blit/blit_trace.tsv` | one row per blit flavour: `flavour, length, src, dst, light, fillColor, want` in hex |
| `cycle/cycle_steps.tsv` | cycle after-state: `step, palette hex, variants sha256, lut sha256` |

## Producer commands (run once, offline; never by tests or CI)

```bash
# 1. palettes, through the Go MPQ stack
DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire
#    levels\towndata\town.pal      -> pal/town.pal.hex
#    levels\l1data\l1_1.pal        -> pal/l1_1.pal.hex
#    nlevels\l5data\l5base.pal     -> pal/l5base.pal.hex
#    nlevels\l6data\l6base1.pal    -> pal/l6base1.pal.hex

# 2. LUTs, variants, blit trace, cycle hashes
g++ -std=c++20 -O1 -I/tmp/dvlx-pinned/Source -I/usr/include/SDL2 \
    $(pkg-config --cflags sdl2) producer/oracle.cpp \
    /tmp/dvlx-pinned/Source/utils/str_cat.cpp -o /tmp/oracle -lSDL2
/tmp/oracle <outdir> internal/testdata/gamedata/lighttables/cathedral_tables.tsv <paldir> \
    town:-1:-1 l1_1:-1:-1 l5base:1:31 diablo:-1:-1
```

`-1:-1` is `GenerateBlendedLookupTable`'s default (blend the whole palette), the
range `LoadPaletteAndInitBlending` uses for town, cathedral, catacombs and hell;
`1:31` is the caves/crypt range (`palette.cpp:188-200`).

## Skill/coding-rules notes

The LUT is stored as hex text rather than a raw `.bin` so it reviews in a diff:
65536 bytes is small enough to commit and large enough to catch any change in the
k-d tree search, the median rounding or the tie-break.
