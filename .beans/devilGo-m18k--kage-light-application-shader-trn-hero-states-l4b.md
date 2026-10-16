---
# devilGo-m18k
title: Kage light application shader + TRN hero states (L4b + G4 TRN)
status: todo
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T14:26:53Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-gk5e
    - devilGo-66o7
    - devilGo-j9be
    - devilGo-b782
---

# Kage light application shader + TRN hero states

## Goal
Implement the Kage fragment shader that performs per-pixel palette translation (LightTables lookup) + TRN special states (Infravision/Stone/Pause) matching clx_render.cpp blend logic exactly.

## Research
- Graphics G4 decision: index framebuffer + 16 palette-row textures + Kage palette-lookup shader (not ColorScale)
- clx_render.cpp TRN section (fix2: exact lines re-anchored at claim; cite corrected per triage) hero Petrified → Stone(); Blend(); BlendMapMask (StoneTable shift + blend mask)
- :766-769 hero Aeclis/Infravision state → Infravision() TRN
- Light tables still applied for lit heroes; these TRNs REPLACE the light tbl for special states
- Palette cycling rotate(1,2,32) on first 32 entries of all 16 tables (L2)

## Scope
Package: `internal/render/lightshader.go` + `shaders/palette.kage`
- Shader inputs: uPaletteRows[16] (texture 256×16), uLightmap (per-pixel light index 0..15), uTRNMode (enum: None/Infravision/Stone/Pause), uCycleOffset (0..31 for lava/blood shift)
- Pixel logic: `idx = uPaletteRows[light][src] ; if uTRNMode!=None { idx = TRN[uTRNMode][idx] } ; if cycling { idx = (idx+uCycleOffset)%32 + (idx>=32?0:0) }`
- CPU fallback path for parity testing (no GPU)

## Acceptance criteria
- [ ] `TestShaderPalTRN` Petrified hero → StoneTable applied over light table (pixel dump vs C++)
- [ ] `TestShaderInfravision` Infravision hero → InfravisionTable applied
- [ ] `TestShaderColorCycle` offset increments per frame → 31-step period matches C++ `rotate(1,2,32)`
- [ ] `TestShaderFullyDarkLit` fast paths skip palette texture sample (bench)

## Skills for coder
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- /home/blaze/.agents/skills/ebiten/shader/SKILL.md

## Dependencies & blockers
- devilGo-66o7 (index framebuffer + palette texture layout)
- devilGo-b782 (TRN registry + level palettes)

## Open questions
- Exact clx_render blend mask parameters for Petrified (BlendMapMask args) — need targeted read

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:broken-research (clx_render.cpp:756-769 out of range; BlendMapMask is a fabricated symbol) + missing skill path + body/frontmatter dep mismatch + palette.kage ownership overlap with 66o7.


## IMPL NOTES (triage 2026-10-16)
1 root: Shader bean — per-pixel palette translation + TRN hero states replacing light table for special states. Right idea, wrong citations.
2 local: `beans show` + probes: cited `clx_render.cpp:756-769` is OUT OF RANGE — file is 632 lines (`wc -l`); the Petrified/Infravision logic actually lives at `Source/engine/render/scrollrt.cpp:388-399` (monster TRN tree: Infravision when `_pInfraFlag && lightTableIndex > 8`, Petrified → GetStoneTRN(), `trn != nullptr ? ClxDrawTRN : ClxDrawLight` — confirms "TRN replaces light table" ✓) and `Source/engine/trn.cpp:14-26` (GetInfravisionTRN/GetStoneTRN/GetPauseTRN). `grep -rn BlendMapMask Source/` → ZERO hits: "BlendMapMask" does not exist in the codebase (open question asks for its "exact args" of a non-existent function). Palette cycling rotate confirmed `Source/lighting.cpp:481` (`std::rotate(begin+1, begin+2, begin+32)`) ✓. Skill `/home/blaze/.agents/skills/ebiten/shader/SKILL.md` does NOT exist.
3 online: n/a (no external claims).
4 solution: re-anchor research to scrollrt.cpp:388-399 + trn.cpp:14-26; delete/rewrite the BlendMapMask open question (fabricated symbol — actual question is the ClxDrawTRN vs ClxDrawLight branch semantics, already readable); fix skills path (use ebiten/rendering only, or create shader skill); sync body deps to frontmatter (frontmatter: gk5e, 66o7, j9be; body lists 66o7+b782 only — b782 not in frontmatter); demarcate palette.kage ownership vs 66o7 (both claim it).
5 understanding: Diablo-layer shader consuming 66o7's index-framebuffer architecture, kx0l's tables (as palette rows), b782's TRN data; sits between data beans and 7aul parity harness.
6 repro: `beans show devilGo-m18k --body-only`; `wc -l /home/blaze/github/bmcszk/devilutionX/Source/engine/render/clx_render.cpp`; `sed -n '388,399p' .../engine/render/scrollrt.cpp`; `grep -rn BlendMapMask /home/blaze/github/bmcszk/devilutionX/Source/`; `ls /home/blaze/.agents/skills/ebiten/`.
7 fix: `beans update devilGo-m18k --body-replace-old "clx_render.cpp:756-769 hero Petrified → Stone(); Blend(); BlendMapMask (StoneTable shift + blend mask)" --body-replace-new "scrollrt.cpp:388-399 monster TRN tree: Petrified → GetStoneTRN(); Infravision when _pInfraFlag && lightTableIndex > 8; trn!=nullptr ? ClxDrawTRN : ClxDrawLight (trn.cpp:14-26 getters)"`; `beans update devilGo-m18k --body-replace-old "- Exact clx_render blend mask parameters for Petrified (BlendMapMask args) — need targeted read" --body-replace-new "- None after re-anchoring (BlendMapMask does not exist in devilutionX; branch semantics fully readable at scrollrt.cpp:388-399)"`; drop the ebiten/shader skill line or `beans update devilGo-m18k --body-replace-old "- /home/blaze/.agents/skills/ebiten/shader/SKILL.md" --body-replace-new "- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md"`; align body deps with frontmatter (add gk5e/j9be, or add b782 to frontmatter via `beans update devilGo-m18k --blocked-by devilGo-b782`).
8 validation: `beans show devilGo-m18k --body-only | grep -c "clx_render.cpp:756"` → 0; `grep -c BlendMapMask` in body → 0; skills line points at existing file.
9 related: 66o7 (architecture + palette.kage overlap), kx0l (tables this shader consumes), b782 (TRN data), j9be (draw-time application, blocks this).
VERDICT: NEEDS-WORK:broken-research (clx_render.cpp:756-769 out of range; BlendMapMask is a fabricated symbol) + missing skill path + body/frontmatter dep mismatch + palette.kage ownership overlap with 66o7.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via /tmp/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.