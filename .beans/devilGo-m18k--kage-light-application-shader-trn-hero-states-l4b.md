---
# devilGo-m18k
title: Kage light application shader + TRN hero states (L4b + G4 TRN)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-15T22:42:03Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-gk5e
    - devilGo-66o7
    - devilGo-j9be
---

# Kage light application shader + TRN hero states

## Goal
Implement the Kage fragment shader that performs per-pixel palette translation (LightTables lookup) + TRN special states (Infravision/Stone/Pause) matching clx_render.cpp blend logic exactly.

## Research
- Graphics G4 decision: index framebuffer + 16 palette-row textures + Kage palette-lookup shader (not ColorScale)
- clx_render.cpp:756-769 hero Petrified → Stone(); Blend(); BlendMapMask (StoneTable shift + blend mask)
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
