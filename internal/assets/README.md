# assets: txtdata loader + font glyphs + sfx catalog (executed, not oriented)

What landed in this package, and the deviations a reader must know about.

## `txtdata.go` — tab-separated table loader

`Load(io.Reader) (*Table, error)` normalizes a table (UTF-8 BOM strip, CRLF and
lone CR folded to LF), parses the header row, and asserts an exact field count
on every data row. `Row.Int/Bool/String/Byte` convert on demand and report
failures instead of returning silent zero values.

Row-count goldens live in `internal/testdata/assets/txtdata/golden/` (113-row
monstdat schema, 50-row Experience table); `TestUpstreamTablesParseFully` parses
a devilutionX checkout when one is next to the worktree and asserts the loader
consumes every non-empty line exactly once.

**Deviation (documented, not silent):** `internal/testdata/assets/txtdata/README.md`
records that upstream treats a lone CR as a bad record terminator
(`data/parser.hpp:115-130`). This loader folds a lone CR to LF instead, as the
txtdata contract asks, so `cr.tsv` parses rather than fails. Parity therefore
holds for CRLF/LF/BOM exports and deliberately diverges on CR-only input.

## `font.go` — CLX font sheet decode

`DecodeFontSheet` decodes the 256 glyphs of a CLX font sheet. Unlike
`DecodeCLX` it accepts zero-width frames: 229 of the 256 glyphs in the shipped
12-05 sheet carry a 6-byte header and no pixel stream. `Glyph`, `Width`,
`LineWidth` (ported `text_render.cpp:631-659`) and `Digest` (sha256 of every
glyph's pixels in sheet order) are the UI lane's API.

## `internal/diablo/gamedata/audio` — sfx catalog

`SfxID` is 917 generated constants over the effects.tsv id column
(`scripts/gen_sfx_ids.py`). `SfxCatalog` indexes rows by id, refuses empty or
duplicate rows at construction, and loads bytes lazily once per id through an
injected `SfxSource`, tagging samples with `DefaultSampleRate` = 22050
(`options.cpp:55`).

**Deviation (documented, not silent):** ebiten audio is not yet a dependency of
this module, so the catalog stops at container bytes. The PCM seam stays
injectable and the engine audio layer owns the mixer context.
