Offline fixture producer
========================

`oracle.cpp` is compiled by hand (never by `go test` or CI) to obtain the
reference bytes the Go port must match. It `#include`s the pinned upstream TUs
so the algorithm producing the goldens is upstream's own code:

- `utils/palette_blending.cpp` — `GenerateBlendedLookupTable`, the LUT under test.
- `utils/palette_kd_tree.cpp` — the k-d tree and its nearest-neighbour search.
- `engine/render/blit_impl.hpp` — all six blit flavours.
- `engine/render/light_render.cpp:499-511` — `Lightmap`'s constructor, copied
  verbatim because the rest of that TU pulls in the tile rasteriser the blit
  flavours never touch.

The only hand-written logic is the variant mapping
(`variant[shade][i] = palette[tables[shade][i]]`, the G4 definition the shader
consumes) and the two cycle rotations, which are copied line-for-line from
`engine/palette.cpp:52-90`.

Build and run:

```bash
g++ -std=c++20 -O1 -I/tmp/dvlx-pinned/Source -I/usr/include/SDL2 \
    $(pkg-config --cflags sdl2) oracle.cpp \
    /tmp/dvlx-pinned/Source/utils/str_cat.cpp -o /tmp/oracle -lSDL2
/tmp/oracle <outdir> <checkout>/internal/testdata/gamedata/lighttables/cathedral_tables.tsv \
    <paldir> town:-1:-1 l1_1:-1:-1 l5base:1:31 diablo:-1:-1
```

`<paldir>` must hold `town.pal`, `l1_1.pal`, `l5base.pal`, `diablo.pal` (768 bytes
each) and `pause.trn` (256 bytes) — the palette files dumped from the retail
records, the TRN from `internal/testdata/gfx/trn/gendata_pause.trn.hex`.

`wraporacle.cpp` is the wrap-layout producer (od1d, bean wrap-layout oracle
goldens from the offline C++ probe): verbatim copies of WordWrapString +
GetLineWidth + their direct helpers from the PINNED upstream
Source/engine/render/text_render.cpp (checkout /tmp/dvlx-pin-ca91 @
ca91d778f), fed the exact corpus internal/render/text_test.go pins and glyph
widths decoded from the committed 12-05.clx.hex golden. `#include`ing the
whole TU was rejected: its dependency surface (font loading, Surface,
app_fatal) is unbounded for an offline probe. Run once, offline:

```bash
g++ -std=c++20 -O2 wraporacle.cpp -o /tmp/wraporacle
/tmp/wraporacle internal/testdata/assets/fonts/12-05.clx.hex \
    > internal/testdata/assets/fonts/wrap_golden.tsv
```

Output TSV rows: `stringIdx<TAB>lineIdx<TAB>lineWidth<TAB>startX<TAB>lineText`
(stringIdx indexes the corpus array inside wraporacle.cpp; lines 0-19 are the
LineWidth-parity strings measured whole, lines 20-30 are the wrap cases with
one row per wrapped line; lineText is escaped).
