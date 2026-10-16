---
# devilGo-66o7
title: Index framebuffer + palette-tran shader (light tables, TRNs, blend LUT)
status: todo
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-1spo
---

# Paletted renderer core

## Goal
Replace the Volume I §7 `ColorScale` plan with the byte-exact architecture: world rendered into an *index* buffer (RGBA8 with the palette index in R), 16 palette-variant rows for lighting, per-TRN palette variants for uniques/infravision/pause, 256×256 blend LUT for transparency, one final palette-lookup pass.

## Research
- docs/research/dg-graphics.md G4 (light table generation transcription, TRN census, combo counts), G5 (lighting mapping + budget), G8 (blend LUT call sites and the 6 blit flavours), G10 D2/D4.
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — shader + offscreen target patterns
/home/blaze/.agents/skills/go-dev/SKILL.md — pure-Go reference blitters as the parity oracle
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden byte comparison
/home/blaze/.agents/skills/ponytail/SKILL.md — smallest change that still hits byte-exactness (no new render graph)

## Dependencies & blockers
blocked_by devilGo-n89r. Blocks devilGo-j9be (which becomes 'blend + automap transparency' only) and devilGo-ru6l.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: SPIKE-AT-CLAIM: software rasteriser reference = YES (GPU-free CI parity, decided). Uniform-per-draw vs per-variant images: measured with ebitenginedebug in a committed spike before the render-path bean, decision recorded here.
- Keep a pure-Go software index rasteriser as reference implementation (recommended, gives GPU-free CI parity)? Yes — decide and record here.
- Uniform-per-draw (palette row index) vs one image per variant: measure batch splits with ebitenginedebug and pick.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Kage shader implements palette-row lookup (index texture + 16 palette rows + uniform paletteOrigin per G4 decision D-palette); renders to an *ebiten.Image representing the 640×480 index framebuffer.
- [ ] Palette fade/cycle implemented CPU-side per G8 (ReplacePixels the 16 palette-row textures only); sprite pixels never touched.
- [ ] Draw-call merging verified: same-atlas + same-blend + same-filter runs batch per `ebitengine.org performancetips`; counted with `-tags=ebitenginedebug` and documented in PoW.


## Resolution gate (grilling decision #5)
- [ ] 640×480 world logical surface rendered correctly, AND 2x window (1280×960) integer-scaled without artifacts — parity golden frames at BOTH sizes.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:duplication-seam (MakeLightTable port + falloff/cone golden ACs claimed by both 66o7 and kx0l; kx0l canonical) + 4 cite drifts


## IMPL NOTES (triage 2026-10-16)
1 root: Graphics-wave task (devilGo-66o7): the paletted-renderer core. Wire: blockedBy [n89r] — matches body devilGo-n89r ✓. BUT the lighting wave later split table generation OUT into devilGo-kx0l (L2) without trimming this body.
2 local: DUPLICATION verified: 66o7 scope still contains "internal/render/lighttable.go (port MakeLightTable verbatim incl. level overrides)" and ACs `TestMakeLightTable` (16×256×6), `TestLightFalloffs` (16×128), `TestLightConeInterpolations` (8×8×16×16) — kx0l's body claims the SAME port (internal/light/tables.go, MakeLightTables) with stricter fixture ACs (falloffs.bin/cone.bin golden, hell/nest branch tests, cycling). Both beans cannot own MakeLightTable. Cites otherwise verified: light_tables.cpp MakeLightTable at :20 (bean 19-70 ✓ range); palette.cpp CycleColors :52-56 ✓, LoadPaletteAndInitBlending :188 ✓; scrollrt.cpp player light bias `std::max(lightTableIndex-5,0)` ✓ but at :492 (bean :489, drift 3), ClxDrawLight :252 (bean :249), ClxDrawLightBlended :267 (bean :264); ClxApplyTrans at :423 (bean :393, drift 30); blit_impl.hpp/palette_blending.cpp/palette_kd_tree.hpp all exist ✓.
3 online: Ebiten claims (no upstream examples/palette; masking/blend as models; batching rules) — verified online during graphics wave per milestone log ✓.
4 solution: Trim to infra: drop lighttable.go from scope and the three table-generation ACs (kx0l is canonical; keep an upload round-trip test "palette-row textures == kx0l tables"); keep blend LUT (blendlut.go — no other bean claims generation; j9be/qeyu only consume), variant cache, final pass, keying-trap and draw-count ACs. Do NOT add a hard 66o7→kx0l edge yet: wire currently has kx0l blockedBy m18k and m18k blockedBy 66o7 — adding it creates a cycle; the kx0l→m18k edge looks wrong (pure data gen doesn't need the shader) and belongs to batch-05's kx0l/m18k triage.
5 understanding: engine/render core implementing G10 D-palette decision (index framebuffer + 16 palette rows + Kage shader); retrofitted gates + resolution gate present ✓.
6 repro: `beans show devilGo-66o7 --body-only | grep -n "lighttable\|TestMakeLightTable"; beans show devilGo-kx0l --body-only | grep -n "MakeLightTables\|falloffs.bin"; sed -n '20p' Source/engine/light_tables.cpp; grep -n "lightTableIndex - 5" Source/engine/render/scrollrt.cpp`
7 fix: `beans update devilGo-66o7 --body-replace-old "internal/render/lighttable.go (port MakeLightTable verbatim incl. level overrides), internal/render/blendlut.go" --body-replace-new "internal/render/blendlut.go"` ; `--body-replace-old "- [ ] \`TestMakeLightTable\` equals a C++ byte dump for all 16 tables × 6 level types (256×16×6 bytes)
- [ ] \`TestLightFalloffs\` (16×128) and \`TestLightConeInterpolations\` (8×8×16×16, incl. perfect-square boundaries) byte-equal" --body-replace-new "- [ ] Table byte-parity owned by devilGo-kx0l (TestFalloffsGolden/TestConeGolden/TestHellBranchColors); this bean asserts the 16 uploaded palette-row textures equal kx0l's generated tables (upload round-trip)"` ; refresh drifted cites (249→252, 264→267, 489→492, 393→423).
8 validation: scope no longer mentions MakeLightTable; AC list has no TestMakeLightTable/TestLightFalloffs; kx0l remains sole generation owner; no cycle introduced.
9 related: kx0l (generation owner — seam), m18k (shader application, blockedBy this), j9be/qeyu (LUT consumers), 7aul (parity harness dep), b782 (palette data dep, batch-03 flagged missing edge b782→66o7).
VERDICT: NEEDS-WORK:duplication-seam (MakeLightTable port + falloff/cone golden ACs claimed by both 66o7 and kx0l; kx0l canonical) + 4 cite drifts

---

## Summary

| id | verdict |
|---|---|
| devilGo-qeyu | NEEDS-WORK:wire-deps (66o7 LUT + zdj7 edges missing; DrawWallConnections :615, AmLineLength :74) |
| devilGo-zdj7 | NEEDS-WORK:seam-undeclared (automap_test port AC duplicated with qeyu; one-line note fixes) |
| devilGo-v1hy | TRIAGED-OK |
| devilGo-pdpc | NEEDS-WORK:content-defect (LightType has PartiallyLit+PerPixel at :112-117, PerPixel dispatch branch missing from scope/ACs; body deps ≠ wire; Lightmap struct at light_render.hpp:14) |
| devilGo-gk5e | TRIAGED-OK (SOL hand-patch list verified index-exact; RenderTileFrame/black_tile cites drifted ~5 lines) |
| devilGo-7aul | NEEDS-WORK:wire-deps (zdj7+v1hy not wired for automap-on/text-compose ACs; deps otherwise exact) |
| devilGo-66o7 | NEEDS-WORK:duplication-seam with devilGo-kx0l (MakeLightTable scope+3 ACs duplicated; do not add 66o7→kx0l edge until kx0l's suspicious blockedBy-m18k edge is fixed — flag to batch-05) |

Counts: TRIAGED-OK 2 · NEEDS-WORK 5 · DUPLICATE-OF 0 · STALE 0 · DECISION-NEEDED 0.
Coverage: 7/7 ids from batch-04-ids.txt triaged exactly once; all are devilGo-qbiy children; none reassigned (note: contract's batch-05 prose overlapped 4 of these ids — batch-04-ids.txt is authoritative).

BATCH-04-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Seam resolution (triage): MakeLightTable port + falloff/cone goldens CANONICAL at devilGo-bodd; this bean consumes them.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).


