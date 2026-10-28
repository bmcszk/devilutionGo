---
# devilGo-n89r
title: CLX/CL2/CEL/PCX decoders + TRN + palettes
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - sprites
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-27T19:12:34Z
parent: devilGo-5wsb
---

# Sprite decode

## Goal
Decode all sprite formats to paletted index frames; convert to *ebiten.Image with palette/TRN variants.

## Research
Source/engine/clx_sprite.hpp:15-27 (header), CL2 RLE per https://github.com/savagesteel/d1-file-formats/blob/master/PC-Mac/CL2.md; load_cel.cpp (CelToClx path), load_clx.cpp, load_pcx.cpp; trn.cpp; palette.cpp:175 LoadPalette; light_tables (16×256) engine/light_tables.hpp; dun tile re-encode ReencodeDungeonCels (levels/reencode_dun_cels.*).

## Scope
internal/assets/{clx,cl2,cel,pcx,trn,palette}.go; SpriteSet type: frames [][]byte indices + built ebiten images per (palette, trn) variant lazily; atlas builder.

## Acceptance criteria
- [x] Golden decode tests for ≥20 sampled frames — **CLX only**: 20 committed files
      (`internal/testdata/gfx/clx/`), decode digest asserted and re-derived in Python (20/20
      match). CEL/CL2/PCX are pinned by SYNTHETIC streams built by in-test helpers that mirror
      the decoder's own assumptions (no committed goldens; no repo-shipped `.cel`/`.cl2`/`.pcx`
      exists — retail-byte decision). That synthetic pinning missed the `0xBF` opcode bug
      (audit 2026-10-28, fixed via `devilGo-0jjm`) and the PCX row-order/BytesPerLine bugs
      (`devilGo-aysp`); real byte-goldens for those formats → `devilGo-l5fp`.
- [ ] TRN variant image equals reference re-palettized pixels — residual → renderer seam
      (`*ebiten.Image` variants; `ebiten` is not a dependency yet).
- [x] Palette load for l1_2.pal etc. from MPQ — `LoadPaletteFromStack` + `AssetSource`, real
      768-byte `ui_art/diablo.pal` golden; retail `l1_2.pal` path derivation is
      `LevelPaletteName` (retail bytes stay out of the tree).
- [ ] Decode benchmarks: full plrgfx warrior sheet < 100ms cold — residual → `devilGo-l5fp`
      (plrgfx bytes are retail-only). `BenchmarkCLXDecode` runs on the committed corpus:
      77978 ns/op (10x) on this machine.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — image creation rules (NewImageFromImage at load, SubImage at draw)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden files

## Dependencies & blockers
devilGo-vof6.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: SPIKE-AT-CLAIM: TRN loading eager for uniques, lazy for light variants — benchmark at claim decides only the light-16 case.
- Lazy vs eager TRN variants: eager for uniques (few), lazy for light (16) — benchmark decides.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before
      implementation; PoW includes the failing→passing run.
      _Reason left open (2026-10-28 audit): RED-first holds for the CEL defects (`cel_test.go:470`
      "CEL: data too short for 536870913 frame offsets", 10-byte header mismatch) but the CLX
      boundary paths were never RED — the `0xBF` opcode had no test at all (`devilGo-0jjm`) and
      the goldens passed against the wrong opcode. Treated as NOT MET until `devilGo-0jjm` /
      `devilGo-aysp` land their captured RED runs in PR #3._
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip,
      or row-exact data table (pick per domain; no approximations).
      _Evidence: 20/20 CLX digest re-derivation + byte-identical upstream `.clx.hex`; TRN/PAL
      goldens byte-identical to upstream; CEL RED→GREEN runs recorded above. Caveat on the
      re-derivation's `0xBF` boundary documented in Proof of Work (2026-10-28).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Decoders byte-exact: fixture CEL/CL2/CLX/PCX → Go decode vs C++ decode → golden PNG hash match (G3 worked examples).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-1 decoder task (devilGo-n89r); draft, blocked_by devilGo-vof6 (wired ✓); all 6 downstream dependents wired (ru6l, mlv2, v1hy, 66o7, b782, l5fp all blocked_by n89r ✓).
2 local: `dX/Source/engine/clx_sprite.hpp:15-27` CLX 6-byte frame header + CL2 ref URL (cited URL exists inside the C++ header itself, line 22) ✓; load_cel/clx/pcx.cpp, trn.cpp, palette.cpp (`:175 LoadPalette` ✓), reencode_dun_cels.* all exist ✓; light_tables.hpp NumLightRadiuses=16 ✓. Skills ebiten/rendering + go-fluent-testing exist ✓. TDD-first golden-decode AC (≥20 frames/format, byte-equal vs C++ dump) present and testable; G10/G11 gate present.
3 online: n/a (CL2.md URL is quoted from the in-tree header, not an independent claim).
4 solution: none — ready.
5 understanding: engine-layer asset decode producing paletted index frames + lazy (palette,TRN) ebiten variants; foundation for every renderer/UI task.
6 repro: `sed -n '15,27p' dX/Source/engine/clx_sprite.hpp`; `ls dX/Source/engine/load_{cel,clx,pcx}.cpp dX/Source/levels/reencode_dun_cels.cpp`; `beans show devilGo-ru6l --json | jq .blocked_by`.
7 fix: none.
8 validation: n/a.
9 related: vof6 (bytes source), l5fp (manifest), b782 (palette/TRN data), ru6l/66o7 (consumers).
VERDICT: TRIAGED-OK.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.


## Progress log
- 2026-10-23 claimed (wave-1 first unblocked task; blocker devilGo-vof6 already completed).

## Proof of Work
Branch `assets/palette-trn-registry` (PR #3).

CEL fidelity defects found in triage and fixed (both 1:1 against the pinned C++):
- multi-group offsets read from the zero slot (`cel_to_clx.cpp:60-66`) — RED
  `cel_test.go:470 CEL: data too short for 536870913 frame offsets`.
- 10-byte frame header skipped, width from the caller, height from line count
  (`cel_to_clx.cpp:76-111`) — RED `cel_test.go:109 CEL frame header size mismatch
  (expected 6, got 10)`; API is now `DecodeCEL(data []byte, widths ...uint16)`.

Gates (parent-verified on the rebased tip, not a worker self-report):
- `go build ./...` ok; `go test -count=1 ./...` 26 packages ok, 10859 pass events, 0 FAIL.
- `scripts/lint.sh ./...` → `0 issues.`
- `gofmt -l` over tracked non-vendor files → empty.
- `go test -run xxx -bench BenchmarkCLXDecode -benchtime 10x ./internal/assets/` →
  `77978 ns/op`.

Golden corpus independence (corrected 2026-10-28 after the brainer audit): the committed
`<flat>.sha256` digests were re-derived with a separate Python CLX decoder (fill/literal/
transparent semantics from `clx_decode.hpp:13-38`) over the same committed `.clx.hex` bytes —
20/20 digests, dims and frame counts match, and all 20 hex files are byte-identical to
`<checkout>/assets/**`. **Caveat:** the re-derivation took the same `0x80..0xBF fill` boundary
the Go decoder had (the C++ truth is `ClxFillMax = 0xBE`, `clx_decode.hpp:25`; `0xBF` is a
65-byte literal, emitted by `clx_encode.hpp:44-49`), so the "independent" check shared the
decoder's wrong spec at that boundary. It still did not catch the bug because no committed
golden exercises `0xBF` at a command position (it occurs in 5 real frames, never as a command).
The boundary is fixed and pinned by tests via `devilGo-0jjm`; `internal/testdata/gfx/README.md`
is corrected in the same round. TRN and palette hex goldens are
byte-identical to `<checkout>/assets/{gendata,fonts,plrgfx,ui_art}`.

Residual (re-scoped, not skipped):
- CEL/CL2/PCX ≥20-frame goldens → `devilGo-l5fp`: the repo ships 0 `.cel`/`.cl2`/`.pcx`, so
  any corpus is retail-byte dependent; CLX corpus (100 repo files) is the committable surface.
- TRN variant images + lazily built `*ebiten.Image` → renderer seam (67xx/qbiy), needs the
  ebiten dependency.
- plrgfx 100 ms decode benchmark → `devilGo-l5fp` (retail bytes, local-only perf check).
- DecodeCEL keeps decoded indices instead of re-encoding to CL2: same pixels, different
  container; recorded here as the deviation (`cel_to_clx.cpp` re-encodes only for its own
  runtime sprite type).
