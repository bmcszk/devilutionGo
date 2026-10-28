# Committed gfx goldens

These are the only committed gfx bytes in the repo. No Go test reads a
devilutionX checkout, a retail MPQ, or a `DG_*` variable: the tests read the
files below and fail (`t.Fatalf`) when one is missing.

## Provenance

- Source tree: `github.com/diasurgical/devilutionX`, commit
  `ca91d778fba92a1f680afb1fb2cdf09337ff1b42`.
- Paths: `<checkout>/assets/` — the repo's own unpacked assets, **not** retail
  `DIABDAT.MPQ` bytes.
- License: Sustainable Use License, see `devilutionX/LICENSE.md` in that repo.

## Layout

| Path | Contents |
|---|---|
| `pal/diablo.pal.hex` | 768 bytes, 48 lines of 16 bytes (`ui_art/diablo.pal`) |
| `trn/<flat>.trn.hex` | 256 bytes, 16 lines of 16 bytes; `gendata/pause.trn`, `plrgfx/warrior/whu/whu{fm,lm,qm}.trn`, `fonts/red.trn`, `fonts/whitegold.trn` |
| `clx/<flat>.clx.hex` | full `.clx` file bytes, 16 bytes per line |
| `clx/<flat>.sha256` | `<sha256>  <index bytes>  <w>x<h>  <frames>` |
| `clx/<flat>.frame0.hex` | full frame-0 index dump for three small files |
| `pcx/four_by_three.pcx.hex` | 4x3 8-bit PCX whose file rows are constant `0x10`, `0x20`, `0x30` |

`<flat>` is the checkout path without `.clx`/`.trn`, `/` replaced by `_`
(e.g. `data/health.clx` → `data_health`).

## Producer commands (run once, offline)

The producer reads the checkout and writes the files above; it is never run by
tests or CI.

```bash
DX=/home/blaze/github/bmcszk/devilutionX/assets
OUT=internal/testdata/gfx
# palette and TRNs: straight 16-bytes-per-line hexdumps
xxd -p -c 16 "$DX/ui_art/diablo.pal" | tr -d '\n'   # → pal/diablo.pal.hex (same shape)
```

For the CLX corpus a `python3` heredoc decodes each file independently of the
Go code (CLX list header: `u32 count`, `count+1` `u32` offsets, last offset ==
file size, otherwise a sheet of per-list offsets; frame header
`u16 headerSize, u16 width, u16 height`; CL2 commands per `clx_decode.hpp:23-35`:
`<0x80` transparent run of `c`, `0x80..0xBE` fill of `0xBF-c` (width 1..63,
`clx_decode.hpp:25` sets `ClxFillMax = 0xBE`), `0xBF` literal of `256-0xBF = 65`,
`0xC0..0xFF` literal of `256-c` (width 1..64)). It writes the `.clx.hex` bytes, hashes the concatenated decoded
indices into `.sha256`, and dumps frame 0 for `data_health`,
`data_talkbutton`, `data_boxrightend`.

Corpus selection (20 smallest non-font files with uniform frame dims):
`data/{boxrightend,boxleftend,hintbox,monstertags,repairSingleBtn,dirtybucp,
dirtybuc,healthbox,hintboxbackground,talkbutton,xpbar,stashnavbtns,resistance,
repairAllBtn,panel8buc,health,panel8bucp,tabBtnUp,hinticons}.clx` and
`ui_art/dvl_but_sml.clx`. Font files are excluded because glyph sheets contain
zero-width frames.

The PCX fixture is synthetic (no PCX ships in the devilutionX repo): a 128-byte
header, `BytesPerLine` 4, and RLE rows `c4 10 c4 20 c4 30`. It pins the C++
bottom-up emission order (`pcx_to_clx.cpp:169`) with asymmetric rows.

Byte-goldens for CEL/CL2/PCX decode against the C++ are tracked in
`devilGo-l5fp`.

The producer asserts, per file: the header parses, every frame has the same
dimensions, `sum(w*h) == len(indices)`, first and last frames are non-empty.
`TestCLXDecodeMatchesCommittedHashes` re-asserts the same invariants and the
digest from the committed bytes, so a Go decoder that drifts from the Python
reference (an independent implementation, not a second Go run) fails.
