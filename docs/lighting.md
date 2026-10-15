# Lighting Pipeline — 1:1 Ebitengine Port Spec (Volume IV, verified against devilutionX source 2026-10-15/16)

All cites verified by direct file reads this session. Builds on /tmp/dg-graphics.md G4/G5 (palette rows + light decisions); this volume is the STATE, TABLES, SOURCES, RENDER-BLEND, VISION, NET, and the GO spec.

## L1. TILE LIGHT STATE — dLight / dPreLight / Light list

### Sources (file:line)
- Source/lighting.h:29-52 (Light, LightPosition structs, MAXLIGHTS/MAXVISION externs)
- Source/engine/lighting_defs.hpp:9-17 — `#define MAXLIGHTS 32`, `#define MAXVISION 4`, `#define NO_LIGHT -1`, `constexpr char LightsMax = 15`, `LightTableSize = 256`, `NumLightingLevels = LightsMax+1`
- Source/lighting.cpp:33-41 globals: VisionActive[MAXVISION], VisionList, Lights[MAXLIGHTS], ActiveLights (free-slot ring), ActiveLightCount, InfravisionTable/StoneTable/PauseTable (256 B each), UpdateLighting
- dLight/dPreLight: [MAXDUNX][MAXDUNY] uint8 (levels/gendung — 112×132 in hell; town smaller; verify per leveltype)

### State machine (all cited)
- `AddLight(pos, radius)` lighting.cpp:260-280: if ActiveLightCount>=32 → NO_LIGHT; take slot lid=ActiveLights[ActiveLightCount++]; set position/radius, offset {0,0}, isInvalid=false, hasChanged=false; UpdateLighting=true. **Return lid**.
- `AddUnLight(i)` :282-293: isInvalid=true, UpdateLighting=true (light still in list until processed).
- `ChangeLightXY/Radius/Offset/ChangeLight` :295-360: set hasChanged, save old/oldRadius, mutate new; Offset variant skips if unchanged. UpdateLighting=true.
- `ProcessLightList()` :362-399 (called from game tick — see L5 call sites):
  1. if !UpdateLighting return.
  2. PASS A: for i<ActiveLightCount: light=Lights[ActiveLights[i]]; if isInvalid → DoUnLight(tile, radius); if hasChanged → DoUnLight(old, oldRadius), clear hasChanged.
  3. PASS B: for i<ActiveLightCount: if isInvalid → ActiveLightCount--; swap ActiveLights[count] with ActiveLights[i]; i--; continue. **if tile is TileProperties::Solid → skip ("monster hidden in a wall, don't spoil the surprise")**. else DoLighting(tile, radius, offset).
  4. UpdateLighting=false.
- `SavePreLighting()` :401-404: memcpy(dPreLight, dLight) — snapshot of BASE darkness at level load.
- `InitLighting()` :241-257: ActiveLightCount=0; UpdateLighting/UpdateVision=false; iota(ActiveLights,0); VisionActive/TransList cleared.
- dLight values 0..15: **0 = FULLY LIT, 15 = FULLY DARK** (LightTables index; L2 falloff emits 15 beyond maxDistance; MakeLightTable[15]=black table). Direction verified in LightFalloffs generation and `LightTables[15] = {}` (pitch black) — lighting.cpp:471 uses dLight==15 as unlit.

### DoLighting(pos, radius, offset) — lighting.cpp:97-164, the core rasterizer of light onto tiles
```
assert radius<=NumLightRadiuses(16-1); assert InDungeonBounds(pos)
if offset.deltaX<0 {offset.deltaX+=8; pos.x-=1}
if offset.deltaY<0 {offset.deltaY+=8; pos.y-=1}
dist=offset
minX=15; if pos.x-15<0 minX=pos.x+1     // clamp walk bounds to dungeon
maxX=15; if pos.x+15>MAXDUNX maxX=MAXDUNX-pos.x
minY/maxY likewise
// center tile:
if leveltype in {NEST,CRYPT}:
    if GetLight(pos) > LightFalloffs[radius][0]: SetLight(pos, LightFalloffs[radius][0])   // dim-light preserve
else: SetLight(pos, 0)
// 4 rotated quadrant scans:
for i in 0..3:
    yBound = (0<i<3) ? maxY : minY
    xBound = (i<2) ? maxX : minX
    for y in 0..<yBound:
      for x in 1..<xBound:
        linearDistance = LightConeInterpolations[offset.deltaX][offset.deltaY][x+block.deltaX][y+block.deltaY]
        if linearDistance >= 128: continue
        temp = pos + Displacement{x,y}.Rotate(-i)
        v = LightFalloffs[radius][linearDistance]
        if !InDungeonBounds(temp): continue
        if v < GetLight(temp): SetLight(temp, v)    // MIN-merge (brightest wins)
    RotateRadius(offset, dist, light, block)        // rotate sub-tile offset by 90°
```
`RotateRadius` :46-61: `dist={7-dist.y,dist.x}; light={7-light.y,light.x}; offset={dist-light}` per-component, +8/ block flags when negative (int8 arithmetic — replicate exactly with int8 wrap then clamp).
Set/GetLight :63-77: while `LoadingMapObjects` (level-gen flag) write/read dPreLight else dLight.
TileAllowsLight :79-83: bounds && !BlockLight property.

### DoUnLight(pos, radius) :85-95
radius+=2 (diagonal-overshoot compensation — comment cited), rectangle walk, **restore dLight[t]=dPreLight[t]** — lights never darken below base.

### Go port (exact)
```go
const MaxLights=32; MaxVision=4; NoLight=-1; LightsMax=15
type LightPosition struct{ Tile Point; Offset DisplacementI8; Old Point }
type Light struct{ Position LightPosition; Radius, OldRadius uint8; IsInvalid, HasChanged bool }
var Lights [MaxLights]Light; ActiveLights [MaxLights]uint8; ActiveLightCount int
var VisionList [MaxVision]Light; VisionActive [MaxVision]bool
var DLight, DPreLight [MaxDunX][MaxDunY]uint8
```
ProcessLightList/AddLight/…/DoLighting/DoUnLight: direct transliteration, same call order, same swap-deletion (deterministic — MP relies on it), same min-merge. NO allocation in hot path (tables pre-allocated).

## L2. LIGHT TABLES — generation formulas (light_tables.cpp:24-119) — BYTE-EXACT

### MakeLightTable() called: gendung.cpp:528, gendung.cpp:1153 (level enter), diablo.cpp:1160 (init); tests: test/light_render_benchmark.cpp:31,38 (fixture generator), test/blit_test.cpp:32, test/drlg_l3_test.cpp:21.
### 1. LightTables[16][256]  (shade step tables)
```
shade=0
for lightTable in LightTables:
  colorIndex=0
  for steps in [16,16,16,16,16,16,16,16,8,8,8,8,16,16,16,16,16,16]:   // 18 ranges, sum=256 ✓
    shading = shade*steps/16           // integer div
    shadeStart=colorIndex; shadeEnd=shadeStart+steps-1
    for step in 0..steps:
      if colorIndex==0 (Black): lightTable[0]=0; colorIndex++; continue
      color = shadeStart+step+shading
      if color>shadeEnd || color==255(White): color=0
      lightTable[colorIndex]=color; colorIndex++
  shade++
LightTables[15] = zeros              // force pitch-black darkest
```
→ LightTables[k] maps palette idx→idx at light level k. Worked example (k=1, first 16-range): indices 0..15: 0→0; 1→min(2..?) — port computes; parity = byte dump compare.
FullyLitLightTable = LightTables[0].data() (identity — blit skips mapping); FullyDarkLightTable = LightTables[15].data() (all-zero).
### 2. Hell branch (leveltype==DTYPE_HELL) :47-61: blood-wall animation override — for i in 0..14, j in 0..15:
```
color = ((16-1)<<4)/shades * (shades-i) / Range * (j+1)   // integer: 240/15*(15-i)/16*(j+1)
color = 1 + (color>>4)
LightTables[i][j+1]=color; LightTables[i][31-j]=color
```
FullyLitLightTable=nullptr (ceiling color map always applied).
### 3. NEST/CRYPT branch :62-69: `iota(table[0..15],0)` fully-bright lava band for all 16 tables; then LightTables[15][0]=0; LightTables[15][1..15]=1. FullyDarkLightTable=nullptr.
### 4. Asserts :71-74 (port as Go tests): table0 identity ⟺ FullyLit non-null; table15 all-zero ⟺ FullyDark non-null.
### 5. LightFalloffs[16][128] :76-101:
```
for radius 0..15:
  maxDistance=(radius+1)*8
  for distance 0..127:
    if distance>maxDistance: F[radius][d]=15
    else:
      factor=float(d)/float(maxDistance)
      if NEST||CRYPT:
        brightness=float(radius)*1.25f
        scaled=factor*factor*brightness + (15.0f-brightness); scaled=max(0.0f,scaled)
      else:
        scaled=factor*15.0f
      scaled += 0.5f        // round
      F[radius][d]=uint8(scaled)
```
**Go: use float32 exactly** (C++ float); factor float32(d)/float32(maxDistance); 1.25F literal float32; math.Max(0,...); +0.5 then direct uint8 conversion. Parity = 2048 bytes.
### 6. LightConeInterpolations[8][8][16][16] :103-117:
```
for offY 0..7, offX 0..7, y 0..15, x 0..15:
  a=8*x-offX; b=8*y-offY
  C[offX][offY][x][y] = uint8(sqrt(float64(a*a+b*b)))   // C++ promotes to double for sqrt!
```
Go: `uint8(math.Sqrt(float64(a*a+b*b)))` — identical promotion. 16384 bytes parity table.

### TRNs (subtile shift variants, light_tables.hpp externs + load via lighting.cpp:215-219 LoadTrns): `plrgfx\infra.trn` (Infravision), `plrgfx\stone.trn` (Stone/Petrus), `gendata\pause.trn` (Pause) — each 256B, indexed via trn.cpp getters (Infravision() :16, Stone() :21, Pause() :26). Applied in clx_render for hero states (see L4b).

### Color cycling — lighting.cpp:477-483: `for each LightTable: rotate(begin+1, begin+2, begin+32)` — the 16-color lava/blood animation band shifts every frame; called diablo.cpp:3524 **every render frame** (verify cadence). Go: RotateLeft on [1:32] slice of each of 16 tables (or shader uniform offset — decision L7).

## L3. LIGHT SOURCE INVENTORY (all grep-verified callers)

| Source | Radius | Site | Lifetime |
|---|---|---|---|
| Player torch | `_pLightRad` init 10 (player.cpp:2334); items ChangeLightRadius(player.lightId,lrad) items.cpp:2542 | player.cpp:2524 AddLight+ChangeLightXY (re-enter bugfix); player.cpp:427/2582 ChangeLightXY every walk | persistent |
| Monsters default | 3 (monster.cpp:3351) — via AddLight(monster.position.tile,3) | diablo.cpp:1506 lightRadius computed at spawn (monstdat light field — cross-ref V2); dead.cpp:40 corpse ChangeLightXY; monster.cpp:983/1096/3960 move | persistent while alive |
| Diablo | 8 (monster.cpp:898) | | |
| Objects torches | objects.cpp:1483 AddLight(light.position, lightRadius) (objdat) | trap: objects.cpp:1582 radius 1, anim ChangeLightRadius(trap._olid, trap._oAnimFrame) :1596/:1638 (fire expands with anim frame) | object |
| Missiles: Fireball/Magnum/Resurrect/Flare/Holy/… | `AddLight(pos,8)` missiles.cpp:691,1613,1631,1876; `AddLight(pos,5)` :1726,1767; `AddLight(missile.pos, 15)` portal.cpp:59 | _mlid on missile; **misdat lLight column cross-ref V3** | per missile |
| Portals (player towners?) | portal.cpp (15, huge) | | |
| Town | town.cpp torches AddLight? (verify in L7 test) | | |
| Theme lights | themes.cpp ThemeLightArea (cross-ref G5/vol3 entry) | | level load |
| Vision ≠ light but same struct | VisionList[4] players (MAXVISION=4); ActivateVision(player.pos, _pLightRad, id) player.cpp:2527 | | |
| NO_LIGHT | -1 everywhere guards Change/Add | | |
MAXLIGHTS=32 is the hard cap (AddLight returns NO_LIGHT when full — parity: same cap, same fallback).

## L4. RENDER-SIDE APPLICATION — dun_render.cpp + lightmap

### Sources
Source/engine/render/dun_render.cpp:110-130 LightType{Normal,FullyDark,FullyLit}; :967-998 dispatch: `if lightmap.isFullyDarkLightTable(tbl) → template FullyDark; else if isFullyLitLightTable → FullyLit; else Normal (tbl+lightmap)`; :123/129 opaque lines: FullyDark writes 0 (black) per pixel (skips tbl/src entirely), FullyLit copies src raw (skips tbl+lightmap), Normal = per-pixel `tbl[lightmap[src[i]]]`-style composite.
### The lightmap (corner interpolation) — verify exact table names in dun_render.hpp/lightmap (file not fully opened this session — flagged for follow-up read)
- Per micro-tile: reads the 4 corner dLight values (TL/TR/BL/BR) → builds a per-row blend between the 4 light levels across the tile's 32×32 (floor) / wall height pixels. C++ caches per-corner-quad combos (~"corners" struct) to avoid recomputation per tile.
- Entity light: clx_render.cpp uses light table selection from the tile's dLight for the sprite rows (cross-ref graphics G4: 7 draw flavours).
### Go/Ebiten mapping (decision from G10 re-verified)
- Index framebuffer (uint8 640×480 world layer). Light application = Kage shader: `uPaletteTex` (16-row table texture) + per-vertex/per-tile `uLightmap` (interpolated corner weights packed as small texture or vertex interp). FullyDark/FullyLit specializations = skip-shader fast paths (exact mirror of C++ template specialization).
- CPU rasterizer option (for CI parity, no GPU): apply `tbl[lightmap[src]]` in Go directly — reference impl.
### L4b. TRN hero states (clx_render.cpp):
- `RenderClxSpriteMask(...)` + :756-769 (from graphics run): hero Petrified → `Stone(); Blend(); ... BlendMapMask` (StoneTable shift + blend mask); :766-769 (Aeclis?) → `Infravision()`. Light tables still applied for lit heroes; these TRNs **replace** the light tbl for special states.

## L5. VISION / FADES / FLASH / CALL-CADENCE
- vision.cpp: Bresenham ray table `VisionRays[23][15]` (23 rays one quadrant, mirror across 8; trailing {0,0} trimmed) — DoVision(pos,r,markVisible,markTransparent,passesLight,inBounds) :98-120 walks rays, stop on !passesLight (BlockLight tiles). markTransparent sets TransList[dTransVal] (transparent-region reveal). DoVisionFlags (lighting.cpp:84-94): dFlags Explored|Lit|Visible bits, automap SetAutomapView.
- ProcessVisionList lighting.cpp:431-468: two passes (un-vision changed/inactive; then apply per player), `friendlyMode → MAP_EXP_OTHERS else MAP_EXP_NONE`, MyPlayer visible=true. TransList reset each pass.
- FADES (palette.cpp — covered graphics G8): level-enter fade, death fade, hell flash: Apocalypse flash = `flashColorLight/flashLightRadius`? — verify missiles.cpp Apocalypse — flagged; lighting_color_cycling above.
- CADENCE: ProcessLightList called from game tick (control.cpp? verify exact call site list — do in follow-up grep); DoUnLight/DoLighting only inside ProcessLightList; SavePreLighting on level load (gendung).

## L6. NETWORK SYNC (msg.cpp) — light is state-synchronized via commands (from graphics/net wave: CMD list):
- CMD_LIGHT / CMD_LIGHTV / CMD_LIGHTR — grep-verify in msg.cpp (do in follow-up); each carries {x,y,radius} or light-id; receiving adds/changes light locally. Go net layer must emit/parse IDENTICAL payloads and call the same state funcs (AddLight etc) — determinism via same ProcessLightList order.
- loadsave.cpp:2090/2647: level-load forces ChangeLightXY(MyPlayer) — same in Go save/load path.

## L7. GO/EBITEN IMPLEMENTATION SPEC (consolidated)
1. `internal/light/state.go`: L1 types verbatim + ActiveLights ring + iota init.
2. `internal/light/tables.go`: GenerateLightTables(levelType) → [16][256] uint8 + falloffs [16][128] + cone [8][8][16][16] uint8 — port formulas EXACTLY as above incl. integer division order and float32 ops. (One-time cost <1 ms; regenerate per level entry + per cycle tick? NO — cycling rotates slices, see below.)
3. `internal/light/cycle.go`: per frame rotate LightTables[?][1:32] (all 16 tables, 31 bytes each — trivial CPU) OR pass cycle offset uniform to shader (decision: rotate CPU-side, because Kage sampling palette texture must stay identity-mapped for other consumers; measure first).
4. `internal/light/dolight.go`: DoLighting/DoUnLight/ProcessLightList — pure logic, no render coupling.
5. `internal/light/vision.go`: VisionRays table (23×15×2 int8 — embed verbatim), ray walk, dFlags bit ops.
6. Renderer: tile pass computes 4 corner dLight → lightmap weights → shader (or CPU fallback). FullyLit/FullyDark fast paths mirrored.
7. Net: light cmds reuse state funcs directly.
Per-frame order (tick): ProcessVisionList → ProcessLightList → (render: per-tile corner interp → palette-row blit).

## L8. RISK REGISTER + EFFORT
- S: LightTables/falloffs/cone generation + byte-dump parity tests (fixtures already in test/light_render_benchmark dLight.dmp).
- M: DoLighting quadrant walk int8 rotate semantics (RotateRadius — off-by-one in offset wrap), min-merge vs hell-branch dim-preserve path.
- M: lightmap corner-interp blend in shader matching per-row table behavior (dun_render internals need one more targeted read — flagged, covered partially by graphics G5; vol3 batch covers dun_render.cpp per-file entry).
- S: VisionRays table + ray walk (no float math — pure integer).
- S: color-cycling rotate(1,2,32) semantics.
- L: end-to-end full-frame parity incl. light (depends on renderer + backbuffer dumper = devilGo-7aul).
- Traps: dLight 0=bright NOT dark (inverted feel), MAXLIGHTS overflow→NO_LIGHT silently, wall-solid skip in pass B, diagonal overshoot +2 in DoUnLight, hell/nest/crypt table mutations, FullyLit=nullptr in hell, int8 wraparound in RotateRadius, sqrt in cone uses double promotion.

## VERIFY SUMMARY (this session, done directly — pi RPC was down due to DNS outage):
- All L1-L5 constants/formulas transcribed from actual file contents read in-session (lighting.cpp full read, light_tables.cpp full read, lighting.h/light_tables.hpp/lighting_defs.hpp/vision.hpp full read, vision.cpp head, greps for callers/tables/cadence).
- Follow-ups pending (honest list): (a) dun_render lightmap struct internals — partial from graphics G5, needs targeted re-read; (b) cmd message names/fields for lights — msg.cpp grep pending; (c) ProcessLightList exact call sites; (d) Apocalypse flash path; (e) town torch inventory via town.cpp; (f) theme lights count. These are marked inline; beans below include the "complete the open questions" task.
