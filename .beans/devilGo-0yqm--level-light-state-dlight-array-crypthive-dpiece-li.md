---
# devilGo-0yqm
title: Level light state — dLight array + crypt/hive dPiece light sources (PlaceCryptLights + PlaceHiveLights)
status: in-progress
type: task
priority: high
created_at: 2026-11-06T23:11:39Z
updated_at: 2026-11-06T23:11:39Z
---

## Goal
Land the level-state half of the lighting subsystem for the two dPiece-driven light families, grouped in one unit: PlaceCryptLights (levels/crypt.cpp:817-838, radius 3 over an 83-entry lava tile set) and PlaceHiveLights (levels/drlg_l3.cpp:2162-2170, radius 9 over dPiece 381-456). Both call DoLighting directly, so this bean also lands the dLight array and the DoLighting update pass (Source/lighting.cpp:117-180) that gives the two families an observable effect.

## Why grouped
User law: group features in one PR. The two families share exactly one mechanism (DoLighting writing dLight); splitting them would land the same primitive twice.

## Out of scope (separate beans)
- Dynamic light list: AddLight / ChangeLightRadius / ProcessLightList (Source/lighting.cpp:254-430) - bean "Light source inventory data package".
- Vision: DoVision / ProcessVisionList - bean "Vision system".
- Renderer consumption of dLight, LightTables shader application - bean "Lighting + transparency at draw time".
- SetCryptSetPieceRoom + InitCryptPieces (need UberRow/CornerStone globals, not dLight).

## Pre-check evidence
- grep -rn 'dLight|dSpecial' internal/: comments only (crypt.go:9-12 deferral, hive.go:117-126 unwired sink, dSpecial in PieceGrid). No dLight state.
- internal/light EXISTS (MakeLightTables, Falloffs[16][128], Cone[8][8][16][16], golden-pinned). No DoLighting anywhere in the repo.
- internal/diablo/levels has ZERO non-test consumers, and internal/game has no renderer/lighting to wire into. So "wired to visibility rendering" degrades to: wired to Dungeon state via one build-time entry point; documented in the PR.

## Acceptance criteria
- [ ] Dungeon.Light is dLight [MAXDUNX][MAXDUNY]uint8, initialised to InitGlobals' default (gendung.cpp:301-307: 0 in town, 15 elsewhere)
- [ ] DoLighting ports lighting.cpp:117-180 exactly: 15-tile quadrant bounds, cone distance gate at 128, min-with-current write, crypt/nest dim base vs set-to-0 for other level types
- [ ] PlaceCryptLights returns one radius-3 source per tile whose dPiece is in the 83-entry lava set (crypt.cpp:819-826), first-match break per tile
- [ ] PlaceHiveLights returns one radius-9 source per tile with dPiece in 381..456 (drlg_l3.cpp:2165-2168)
- [ ] One build-time entry point stamps both families into dLight (ApplyLights)
- [ ] Unit tests: source sets recomputed independently from the committed crypt/hive dPiece goldens; DoLighting dim-base vs plain, min semantics, bounds, cone falloff
- [ ] Gates: build, gofmt clean, golangci-lint 0, vet, go test ./... -count=1 green
