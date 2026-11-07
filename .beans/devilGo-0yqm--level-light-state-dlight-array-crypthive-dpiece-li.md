---
# devilGo-0yqm
title: Level light state — dLight array + crypt/hive dPiece light sources (PlaceCryptLights + PlaceHiveLights)
status: completed
type: task
priority: high
created_at: 2026-11-06T23:11:39Z
updated_at: 2026-11-06T23:21:12Z
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
- [x] Dungeon.Light is dLight [MAXDUNX][MAXDUNY]uint8, initialised to InitGlobals' default (gendung.cpp:301-307: 0 in town, 15 elsewhere)
- [x] DoLighting ports lighting.cpp:117-180 exactly: 15-tile quadrant bounds, cone distance gate at 128, min-with-current write, crypt/nest dim base vs set-to-0 for other level types
- [x] PlaceCryptLights returns one radius-3 source per tile whose dPiece is in the 83-entry lava set (crypt.cpp:819-826), first-match break per tile
- [x] PlaceHiveLights returns one radius-9 source per tile with dPiece in 381..456 (drlg_l3.cpp:2165-2168)
- [x] One build-time entry point stamps both families into dLight (ApplyLights)
- [x] Unit tests: source sets recomputed independently from the committed crypt/hive dPiece goldens; DoLighting dim-base vs plain, min semantics, bounds, cone falloff
- [x] Gates: build, gofmt clean, golangci-lint 0, vet, go test ./... -count=1 green


## Progress log
Claimed and implemented in one session. Pre-check found the contract's premise partly stale: internal/light already exists (tables golden-pinned), there is no DoLighting anywhere, and internal/diablo/levels has zero non-test consumers. The lane therefore landed the level-state half (dLight + DoLighting + both source families) and stopped at the renderer boundary, as the scope guard requires.

## Proof of Work

### Gates (verbatim, real output)
```
$ gofmt -l .
(empty)

$ go vet ./...
vet rc=0

$ golangci-lint run ./...
0 issues.

$ go test ./... -count=1
44 packages ok, 0 FAIL
ok  github.com/bmcszk/devilutionGo/internal/diablo/levels  9.251s

$ just sqlc-check
no raw SQL outside sqlc-generated code
```

### RED -> GREEN
RED (before implementation):
```
internal/diablo/levels/light_sources_test.go:25:19: undefined: levels.CryptLavaTilesForTest
internal/diablo/levels/light_sources_test.go:51:20: undefined: levels.PlaceCryptLights
internal/diablo/levels/light_sources_test.go:54:19: undefined: levels.Light
internal/diablo/levels/light_sources_test.go:82:13: levels.PlaceHiveLights(pieces) (no value) used as value
FAIL github.com/bmcszk/devilutionGo/internal/diablo/levels [build failed]
```
GREEN: all of the above pass, plus the parity tests against committed goldens.

### Two test premises corrected (honest record)
DoLighting only ever brightens; my first crypt/only-brightens tests assumed a lower value was darker. Upstream is `if (v < GetLight) SetLight`, so a LOWER dLight value is a BRIGHTER tile (lighting.cpp:173-175). Tests renamed and re-derived (`TestDoLightingOnlyBrightens`, `TestDoLightingDimCryptCentre`); the port was not changed to match my wrong premise.

### Upstream line drift found
- Task preamble cites `Source/lighting.cpp` "117-160" for DoLighting; the function actually spans 117-180.
- Task preamble calls the file `lightandvillage.cpp`; it is `lighting.cpp`.
- hive.go's own comment cited `drlg_l3.cpp:2169-2178` for PlaceHiveLights; the function body is 2162-2170 and 2169-2178 is the CreateL3Dungeon PlaceLights call site + tail. Both cited correctly in the new code.
- crypt.go cited a non-existent bean `devilGo-1v0l`; corrected to name the deferred functions instead of a dead ID.

### Deviation table
| Upstream | Port | Why |
|---|---|---|
| DoLighting reads the `leveltype` and `LightFalloffs` globals | tables + level type injected into Dungeon/args | AGENTS.md: no package-level mutable state; same arithmetic |
| PlaceHiveLights calls a DoLighting callback via AddLight-like sink | returns []Light, callers apply | one shape for both families; sink was unwired and had no consumer |
| InitGlobals memsets dLight for the whole grid | Dungeon.InitLight explicit method | generator-agnostic, testable without a level build |
| static passes run inside CreateL5Dungeon/CreateL3Dungeon | ApplyLights is the single entry point, called after ExpandPieces | this package's Generate* does not own Pass3, so the composer is the caller |

## Skills for coder
- /home/blaze/.agents/skills/beans/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-tests/SKILL.md — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
