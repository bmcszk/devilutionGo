# Committed font goldens

CLX font sheets for the text renderer. No Go test reads a devilutionX checkout,
a retail MPQ, or a `DG_*` variable: the tests read the files below and fail
(`t.Fatalf`) when one is missing.

## Provenance

- Source tree: `github.com/diasurgical/devilutionX`, commit
  `ca91d778fba92a1f680afb1fb2cdf09337ff1b42`.
- Path: `assets/fonts/12-05.clx` — the repo's own unpacked assets, **not**
  retail `DIABDAT.MPQ` bytes.
- License: Sustainable Use License, see `devilutionX/LICENSE.md` in that repo.
- `12-05.clx` is the smallest committed font sheet (4304 bytes, 256 frames) and
  therefore the cheapest corpus that still exercises the real layout: 229 of the
  256 glyphs are zero-width (`w=0,h=20`, a 6-byte header and no pixel stream).

## Layout

| Path | Contents |
|---|---|
| `12-05.clx.hex` | full `.clx` file bytes, 16 bytes per line |
| `12-05.sha256` | `<sha256>  <index bytes>  <frames>  <zero-width frames>  <glyph row heights>` for the concatenated decoded glyph indices |
| `12-05.glyph-<key>.hex` | full index dump of one glyph: `d0` = first non-empty glyph (8x20), `de` = widest glyph (10x20), `00` = a zero-width glyph (empty file) |

## Producer commands (run once, offline)

The producer reads the checkout and writes the files above; it is never run by
tests or CI. The CL2 command table it uses is the one pinned by
`internal/testdata/gfx/README.md` (`<0x80` transparent run, `0x80..0xBE` fill of
`0xBF-c`, `0xBF` literal of 65, `0xC0..0xFF` literal of `256-c`).

```bash
cd <devilutionX checkout>
python3 - <<'EOF'
import struct, hashlib, os
OUT = "<devilutionGo>/internal/testdata/assets/fonts"
SRC = "assets/fonts/12-05.clx"

def decode(src, w, h):
    px = bytearray(w*h); pos = 0; i = 0
    while i < len(src) and pos < len(px):
        c = src[i]; i += 1
        if c < 0x80:
            pos += c
        elif c <= 0xBE:
            run = 0xBF - c; v = src[i]; i += 1
            for j in range(run):
                if pos+j < len(px): px[pos+j] = v
            pos += run
        else:
            run = 256 - c if c > 0xBF else 65
            for j in range(run):
                if pos < len(px): px[pos] = src[i+j]
                pos += 1
            i += run
    return bytes(px), pos

d = open(SRC, 'rb').read()
n = struct.unpack_from('<I', d, 0)[0]
offs = struct.unpack_from('<%dI' % (n+1), d, 4)
open(f"{OUT}/12-05.clx.hex", "w").write(
    "".join(d[i:i+16].hex() + "\n" for i in range(0, len(d), 16)))
allpx = bytearray(); zero = 0; rows = set(); glyphs = []
for i in range(n):
    fr = d[offs[i]:offs[i+1]]
    hs, w, h = struct.unpack_from('<HHH', fr, 0)
    rows.add(h)
    if w == 0: zero += 1
    px, pos = decode(fr[hs:], w, h)
    assert pos == w*h, (i, w, h, pos)
    glyphs.append(px); allpx += px
digest = hashlib.sha256(bytes(allpx)).hexdigest()
open(f"{OUT}/12-05.sha256", "w").write(
    f"{digest}  {len(allpx)}  {n}  {zero}  {','.join(map(str, sorted(rows)))}\n")
EOF
```
