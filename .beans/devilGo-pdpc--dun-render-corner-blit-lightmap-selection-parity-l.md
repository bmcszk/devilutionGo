---
# devilGo-pdpc
title: Dun_render corner-blit lightmap selection parity (L4)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-15T22:42:22Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-gk5e
    - devilGo-n89r
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
---

# Dun_render corner-blit lightmap selection parity

## Goal
Port the exact per-tile lightmap selection from dun_render.cpp:967-998 and LightType dispatch (FullyDark/FullyLit/Normal) so each micro-tile's 4 corner dLight values produce the same palette-row selection per pixel row as C++.

## Research
- Source/engine/render/dun_render.cpp:113-130 LightType enum (Normal/FullyDark/FullyLit)
- :123/129 RenderLineOpaque/Transparent FullyDark writes 0 (black), FullyLit copies src raw, Normal uses tbl+lightmap
- :967-998 dispatch: `isFullyDarkLightTable(tbl) -> FullyDark; isFullyLitLightTable -> FullyLit; else Normal`
- Lightmap structure: 4 corner dLight → interpolated per-row blend across 32×32 floor / wall height pixels
- Graphics G10/G5 already decided: index framebuffer + 16 palette-row textures + Kage shader for Normal path
- Shader must replicate the per-row corner-weight table C++ caches ('corners' struct) → pre-bake or compute on GPU

## Scope
Package: `internal/render/lightmap.go` + shader (`shaders/lightmap.kage`)
- Lightmap build: per tile read 4 corners (TL/TR/BL/BR) from dLight → build 32-row weight matrix (floor) or wall-column matrix
- `isFullyDarkLightTable` / `isFullyLitLightTable` checks (compare tbl pointer to LightTables[0] and [15] data)
- Kage shader: uPaletteTex (16×256 palette), uLightmap (interpolated weights), uTileCornerLights (4 corners)
- Fast paths: FullyDark (write 0), FullyLit (copy src) — skip palette lookup entirely
- CPU fallback: `func ApplyLightmap(dst *ebiten.Image, src *ebiten.Image, corners [4]uint8, tbl *uint8)` exact per-pixel math

## Acceptance criteria
- [ ] `TestLightmapCorners` 4-corner combos (0/7/15/3 etc) → shader output matches C++ CPU fallback dump (fixture `test/fixtures/light_render_benchmark/lightmap_*.bin`)
- [ ] `TestFullyDarkFastPath` tbl==LightTables[15] → output all black without palette read
- [ ] `TestFullyLitFastPath` tbl==LightTables[0] (non-hell) → output src raw
- [ ] `TestHellBranch` leveltype=HELL → FullyLit=nil (no fast path), palette animation applies
- [ ] `TestTileMicroGrid` every micro-tile type (floor/wall/arch) → lightmap weights match C++ per-row blend

## Skills for coder
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md

## Dependencies & blockers
- devilGo-gk5e (tile geometry + dLight array layout)
- devilGo-66o7 (index framebuffer + palette shader infra)
- devilGo-j9be (transparency blend LUT)

## Open questions
- Lightmap struct internals in dun_render.hpp/cpp not fully read this session — targeted re-read required to extract exact weight-table dimensions and cache layout
