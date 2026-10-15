---
# devilGo-kx0l
title: Light table generation — LightTables[16][256], falloffs[16][128], cone[8][8][16][16] (L2)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:04Z
updated_at: 2026-10-15T22:42:22Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-b782
    - devilGo-bodd
    - devilGo-m18k
---

# Light table generation — byte-exact port of MakeLightTable()

## Goal
Generate all 4 tables exactly as C++ light_tables.cpp:24-119: 16x256 shade tables with leveltype branches (normal/hell/nest+crypt), falloffs with quadratic vs linear formulas (float32 arithmetic), cone interpolations (sqrt double promotion), TRN loads, color cycling rotate.

## Research
- Source/engine/light_tables.cpp:24-119 (full function)
- Source/engine/light_tables.hpp:10-17 (extern declarations)
- Source/engine/lighting_defs.hpp:14-17 (LightTableSize=256, NumLightingLevels=16, NumLightRadiuses=16)
- Constants: VolumeScale=3321.9281, MillibelMin=-10000, StereoSeparation=6000 (from soundsample.cpp, also used in light? No - that's audio; lighting uses LightFalloffs only)
- LightFalloffs formula: factor=d/maxDist; NEST/CRYPT: brightness=radius*1.25; scaled=factor^2*brightness + (15-brightness); else scaled=factor*15; scaled+=0.5; uint8
- LightConeInterpolations: a=8*x-offX; b=8*y-offY; C=uint8(sqrt(double(a*a+b*b))) — double promotion!
- Hell branch: color = ((15<<4)/15)*(15-i)/16*(j+1); color=1+(color>>4); symm j+1 & 31-j
- NEST/CRYPT branch: iota(0..15) on first 16 entries; table[15][0]=0; table[15][1..15]=1
- Color cycling: rotate(begin+1, begin+2, begin+32) per table, called every frame

## Scope
Package: `internal/light/tables.go`
- `func MakeLightTables(levelType DungeonType) (LightTables [16][256]uint8, Falloffs [16][128]uint8, Cone [8][8][16][16]uint8, FullyLit, FullyDark *[256]uint8)`
- Exact int8/uint8 arithmetic; int divisions first, then float32 operations
- sqrt uses math.Sqrt(float64()) matching C++ double promotion
- levelType switch for three branches (normal/hell/nest+crypt)
- Return nil for FullyLit/FullyDark when disabled (hell/nest+crypt)
- ColorCycle(offset int) rotates slice[1:32] of each table

## Acceptance criteria
- [ ] `TestLightTablesIdentity` shade 0 table = identity (0..255); table15 all zeros (unless hell/nest+crypt branch mutates)
- [ ] `TestFalloffsGolden` all 2048 bytes match C++ fixture (`test/fixtures/light_render_benchmark/falloffs.bin`)
- [ ] `TestConeGolden` all 16384 bytes match C++ fixture (`test/fixtures/light_render_benchmark/cone.bin`)
- [ ] `TestHellBranchColors` leveltype=HELL → tables 0..14 have blood-wall color override, FullyLit=nil, table[15] unchanged
- [ ] `TestNestCryptBranch` leveltype=NEST/CRYPT → first 16 entries = 0..15; table15[0]=0, [1..15]=1; FullyDark=nil
- [ ] `TestColorCycling` rotate(1,2,32) applied each tick → 31-step period, byte-dump after N frames matches C++

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Dependencies & blockers
- devilGo-n89r (palette/TRN decoder for table index→color pipeline)
- devilGo-b782 (palette registry level-specific)

## Open questions
- None - formulas transcribed verbatim from light_tables.cpp:24-119 with integer/float32 order.
