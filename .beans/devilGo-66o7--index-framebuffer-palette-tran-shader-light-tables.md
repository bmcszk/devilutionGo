---
# devilGo-66o7
title: Index framebuffer + palette-tran shader (light tables, TRNs, blend LUT)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# Paletted renderer core

## Goal
Replace the Volume I §7 `ColorScale` plan with the byte-exact architecture: world rendered into an *index* buffer (RGBA8 with the palette index in R), 16 palette-variant rows for lighting, per-TRN palette variants for uniques/infravision/pause, 256×256 blend LUT for transparency, one final palette-lookup pass.

## Research
- /tmp/dg-graphics.md G4 (light table generation transcription, TRN census, combo counts), G5 (lighting mapping + budget), G8 (blend LUT call sites and the 6 blit flavours), G10 D2/D4.
- Source/engine/light_tables.cpp:19-70 (MakeLightTable incl. hell/nest/crypt overrides, FullyLit/FullyDark nulling)
- Source/engine/render/blit_impl.hpp:20-190 (BlitDirect/WithMap/Blended/BlendedWithMap/Lightmap variants — the exact byte ops to mirror)
- Source/engine/render/clx_render.cpp:556-590 + Source/engine/render/clx_render.hpp:40/:76/:81/:89/:107 (the draw flavours), Source/engine/render/clx_render.cpp:393 ClxApplyTrans
- Source/utils/palette_blending.cpp:30-88 (BlendColors integer halving, symmetry, skipFrom/skipTo), Source/utils/palette_kd_tree.hpp (tie-breaking!)
- Source/engine/palette.cpp:52-90 (CycleColors/CycleColorsReverse rotate palette AND LUT rows+columns), :140 ApplyFadeLevel, :188 LoadPaletteAndInitBlending
- Source/engine/render/scrollrt.cpp:249 (ClxDrawLight), :264 (ClxDrawLightBlended), :378-405 (monster TRN tree), :468-495 (player light bias `max(l-5,0)` at :489)
- Kage docs https://ebitengine.org/en/documents/shader.html ; batching https://ebitengine.org/en/documents/performancetips.html ; verified there is NO upstream examples/palette (https://github.com/hajimehoshi/ebiten/tree/v2.9.9/examples) — model the shader on examples/masking and examples/blend.
- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Scope
internal/render/palette.go (system palette, tone-map, fades, cycling, `[]color.RGBA` upload), internal/render/lighttable.go (port MakeLightTable verbatim incl. level overrides), internal/render/blendlut.go (port GenerateBlendedLookupTable + KD-tree), internal/render/shaders/palette.kage + blend.kage, the world index target + final pass, per-(trn,level) palette-variant cache with an LRU cap.

## Acceptance criteria
- [ ] `TestMakeLightTable` equals a C++ byte dump for all 16 tables × 6 level types (256×16×6 bytes)
- [ ] `TestLightFalloffs` (16×128) and `TestLightConeInterpolations` (8×8×16×16, incl. perfect-square boundaries) byte-equal
- [ ] `TestBlendingLut` byte-equal for town / L1 / crypt palettes with the correct skip ranges
- [ ] Cycling advances palette AND re-derives the 16 light variants and the LUT (test asserts a known after-state hash)
- [ ] Sprites draw with index-0 keying while tiles write index 0 as black (explicit regression test for the keying trap in G4)
- [ ] Final pass is 1 draw command; measured GPU draw commands for a full frame < 40 (`-tags=ebitenginedebug` output captured in the test log artifact)
- [ ] No ColorScale/ColorMatrix anywhere on the world index path (lint via a grep guard in the test)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — shader + offscreen target patterns
/home/blaze/.agents/skills/go-dev/SKILL.md — pure-Go reference blitters as the parity oracle
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden byte comparison
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest change that still hits byte-exactness (no new render graph)

## Dependencies & blockers
blocked_by dg-task-007. Blocks dg-task-011 (which becomes 'blend + automap transparency' only) and dg-task-010.

## Open questions
- Keep a pure-Go software index rasteriser as reference implementation (recommended, gives GPU-free CI parity)? Yes — decide and record here.
- Uniform-per-draw (palette row index) vs one image per variant: measure batch splits with ebitenginedebug and pick.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Kage shader implements palette-row lookup (index texture + 16 palette rows + uniform paletteOrigin per G4 decision D-palette); renders to an *ebiten.Image representing the 640×480 index framebuffer.
- [ ] Palette fade/cycle implemented CPU-side per G8 (ReplacePixels the 16 palette-row textures only); sprite pixels never touched.
- [ ] Draw-call merging verified: same-atlas + same-blend + same-filter runs batch per `ebitengine.org performancetips`; counted with `-tags=ebitenginedebug` and documented in PoW.


## Resolution gate (grilling decision #5)
- [ ] 640×480 world logical surface rendered correctly, AND 2x window (1280×960) integer-scaled without artifacts — parity golden frames at BOTH sizes.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:duplication-seam (MakeLightTable port + falloff/cone golden ACs claimed by both 66o7 and kx0l; kx0l canonical) + 4 cite drifts
