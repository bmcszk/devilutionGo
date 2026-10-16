# Graphics Pipeline — 1:1 Ebitengine Port Spec

Extracted read-only from devilutionX @ `ca91d778fba92a1f680afb1fb2cdf09337ff1b42` (cwd `/home/blaze/github/bmcszk/devilutionX`).
Every `file:line` citation below was opened and read during this extraction. Byte-level examples were produced by
decoding real asset files (`assets/data/*.clx`, and fixtures extracted read-only from a local `DIABDAT.MPQ` into
`/tmp/dgx/`), never from memory.

**Terminology used throughout**
- *micro tile / frame* = one 32×32 (or 32×31) tile drawing primitive inside a dungeon CEL frame.
- *tile* = one dungeon tile = 64×32 px = 2×2 micro tiles (`TILE_WIDTH 64`, `TILE_HEIGHT 32`, `Source/levels/dun_tile.hpp:7`).
- *index* = an 8-bit palette index. devilutionX renders into an **8-bit indexed backbuffer** and only the palette
  (`logical_palette`/`system_palette`, 256 × RGB8) turns indices into colours at present time.
- *TRN* = a 256-byte palette→palette lookup (`dst = trn[src]`), including the 16 light tables.

**The single most important parity fact:** in the C++ game, the entire world is drawn as *indices*, and lighting,
unique-monster colours, class colours, petrification, infravision and 50 % transparency are all **exact 8-bit table
lookups**, not blends. Any Go renderer that reaches for `ColorScale`/alpha compositing instead of reproducing those
tables *will not* be pixel-exact. Section G4/G5/G8 define the required Go equivalents; the recommended architecture
is "index textures + one palette-lookup Kage fragment shader" (see G10).

---

## G1. FRAME ARCHITECTURE

### Sources (file:line)
- Source/engine/render/scrollrt.cpp:1834 — `DrawAndBlit()` (top-level per-frame render entry)
- Source/engine/render/scrollrt.cpp:1313 — `DrawView(out, startPosition)`
- Source/engine/render/scrollrt.cpp:1229 — `DrawGame(fullOut, position, offset)`
- Source/engine/render/scrollrt.cpp:955 — `DrawFloor(...)`
- Source/engine/render/scrollrt.cpp:992 — `DrawTileContent(...)`
- Source/engine/render/scrollrt.cpp:1088 — `DrawOOB(...)`
- Source/engine/render/scrollrt.cpp:781 — `DrawDungeon(...)` (per-tile content order)
- Source/engine/render/scrollrt.cpp:1121 — `Zoom(out)` (2× nearest upscale in-place)
- Source/engine/render/scrollrt.cpp:1675 — `CalcViewportGeometry()` (tileRows/tileColumns/tileShift/tileOffset)
- Source/engine/render/scrollrt.cpp:1718 — tileRows formula
- Source/engine/render/scrollrt.cpp:1719 — tileColumns formula
- Source/engine/render/scrollrt.cpp:1620 — `CalcTileOffset`
- Source/engine/render/scrollrt.cpp:1645 — `TilesInView`
- Source/engine/render/scrollrt.cpp:1490 — `DoBlitScreen(Rectangle area)`
- Source/engine/render/scrollrt.cpp:1515 — `DrawMain(dwHgt, drawDesc, drawHp, drawMana, drawSbar, drawBtn)`
- Source/engine/render/scrollrt.cpp:1575 — `GetOffsetForWalking` (camera-mode flag)
- Source/engine/render/scrollrt.cpp:1738 — `ClearScreenBuffer`
- Source/engine/dx.cpp:58 — `SDL_Surface *PalSurface;` (the 8-bit backbuffer)
- Source/engine/dx.cpp:50 — `SDL_Renderer *renderer;`
- Source/engine/dx.cpp:53 — `SDLTextureUniquePtr texture;`
- Source/engine/dx.cpp:56 — `SDLSurfaceUniquePtr RendererTextureSurface;` (24-bit)
- Source/engine/dx.cpp:181 — `BltFast(srcRect, dstRect)`
- Source/engine/dx.cpp:201 — `Blit(src, srcRect, dstRect)`
- Source/engine/dx.cpp:287 — `RenderPresent()` (this is the "PresentPrimarySurface" of this fork)
- Source/engine/backbuffer_state.hpp:17 — `struct BackbufferState { RedrawState redrawState; DrawnCursor cursor; };`
- Source/engine/backbuffer_state.hpp:22 — `struct BackbufferPtrAndState { void *ptr; BackbufferState state; };`
- Source/engine/backbuffer_state.hpp:41 — `void RedrawEverything();`
- Source/engine/backbuffer_state.hpp:43 — `bool IsRedrawViewport();`
- Source/engine/backbuffer_state.hpp:47 — `DrawnCursor &GetDrawnCursor();`
- Source/engine/palette.cpp:152 — `UpdateSystemPalette`
- Source/engine/palette.cpp:170 — `void SystemPaletteUpdated(int first, int ncolor)`
- Source/engine/palette.cpp:175 — `LoadPalette(path)`
- Source/engine/palette.cpp:271 — `PaletteFadeIn(fr, srcPalette)`
- Source/engine/palette.cpp:324 — `PaletteFadeOut`
- Source/engine/palette.cpp:363 — `palette_update_caves()`
- Source/utils/display.cpp:137 — `CalculateUIRectangle()` (`constexpr Size UISize { 640, 480 }`)
- Source/utils/display.cpp:263 — `sizes.emplace_back(640, 480)` (vanilla resolution entry)
- Source/utils/display.cpp:432 — `AdjustToScreenGeometry` sets `gnScreenWidth/gnScreenHeight`
- Source/utils/display.cpp:734 — `ReinitializeIntegerScale`: `SDL_LOGICAL_PRESENTATION_INTEGER_SCALE` vs `LETTERBOX`
- Source/utils/display.cpp:779 — `SDL_SetRenderLogicalPresentation(renderer, gnScreenWidth, gnScreenHeight, ...)`
- Source/utils/display.cpp:939 — `bool OutputRequiresScaling()`
- Source/utils/display.cpp:950 — `void ScaleOutputRect(SDL_Rect *rect)` (integer scaling math)
- Source/options.h:41 — `#define DEFAULT_WIDTH 640`
- Source/options.h:44 — `#define DEFAULT_HEIGHT 480`
- Source/control/control_panel.cpp:274 — `gnViewportHeight = gnScreenHeight;` (reduced when the main panel covers)
- Source/engine/surface.hpp:30 — `struct Surface { SDL_Surface *surface; ... }` (the render target abstraction)
- Source/engine/surface.hpp:98 — `uint16_t pitch()`
- Source/engine/surface.hpp:111 — `Surface subregion(int x, int y, int w, int h)`
- Source/engine/surface.hpp:130 — `Surface subregionY(int y, int h)`
- Source/engine/render/automap_render.cpp:168 — `SetMapPixel` checks `MinimapRect` / `AutomapType::Transparent`
- Source/engine/light_tables.cpp:14 — `std::array<std::array<uint8_t, LightTableSize>, NumLightingLevels> LightTables;`
- test/scrollrt_test.cpp:13 — `TEST(Scroll_rt, calc_tiles_in_view_original)` (golden tile counts)
- test/scrollrt_test.cpp:80 — `TEST(Scroll_rt, calc_tile_offset_original)` (golden tile offsets)

### Formats & algorithms

**1. There is exactly one logical framebuffer: an 8-bit indexed surface of `gnScreenWidth × gnScreenHeight`
(default 640×480), `PalSurface`.** Everything — world, panels, text, cursor, automap — is blitted into that single
buffer. There is no world-vs-UI layer separation in C++; ordering is purely "later draw wins".

```
Source/engine/dx.cpp:58   SDL_Surface *PalSurface;              // 8-bit INDEX8 surface, we render to this
Source/engine/dx.cpp:122   ... SDL_CreateSurface/... /*depth=*/8, SDL_PIXELFORMAT_INDEX8)
```
Present path (`Source/engine/dx.cpp:287` `RenderPresent`):
```
if (renderer != nullptr) {              // SDL2/3 GPU path
    SDL_SetRenderDrawColor(renderer, 0,0,0,255); SDL_RenderClear(renderer);
    SDL_UpdateTexture(texture, nullptr, surface->pixels, surface->pitch);   // 8-bit + palette -> RGBA
    SDL_RenderTexture(renderer, texture, nullptr, nullptr);                 // logical presentation scales
    SDL_RenderPresent(renderer);
} else {                                 // surface path (SDL1 / direct)
    UpdateOutputSurface();               // possibly SDL_UpdateWindowSurfaceRects(dirtyRects)
}
```
`BltFast` (`Source/engine/dx.cpp:181`) is the "copy PalSurface → window" step, called *before* `RenderPresent`;
with a renderer it is a no-op full-update, with the surface path it records up to 16 dirty rects
(`Source/engine/dx.cpp:60` `MaxDirtyRects = 16`). **Dirty-rect bookkeeping is a presentation optimisation only and is
explicitly not ported** (Ebitengine redraws the whole logical screen each frame).

**2. Frame sequence.** `DrawAndBlit()` (`Source/engine/render/scrollrt.cpp:1834`) per call:
```
hgt = 0
drawHealth/drawMana/drawControlButtons/drawBelt from backbuffer_state component flags
if (gnScreenWidth > mainPanel.width || IsRedrawEverything()) { full redraw: hgt = gnScreenHeight, drawCtrlPan=true }
else if (IsRedrawViewport()) { hgt = gnViewportHeight; drawCtrlPan = false }
out = GlobalBackBuffer()                      // Surface over PalSurface, full screen
UndrawCursor(out)                             // restore the pixels saved behind the cursor (see below)
nthread_UpdateProgressToNextGameTick()        // ProgressToNextGameTick <- 0..128  (rendering interpolation!)
DrawView(out, ViewPosition)                   // world + automap + QoL labels + panels + flasks-upper
if (drawCtrlPan) DrawMainPanel(out)
if (drawHealth)  DrawLifeFlaskLower(out, ...)
if (drawMana)    DrawManaFlaskLower(out, ...); DrawSpell(out)
if (drawControlButtons) DrawMainPanelButtons(out)
if (drawBelt)    DrawInvBelt(out)
if (drawChatInput) DrawChatBox(out)
DrawXPBar(out); [DrawFlaskValues(...)]
DrawCursor(out)                               // saves pixels behind cursor, then blits cursor sprite
DrawFPS(out)
DrawMain(hgt, drawInfoBox, drawHealth, drawMana, drawBelt, drawControlButtons)   // BltFast of regions
RedrawComplete(); clear component flags
RenderPresent()
```

**3. `DrawView` (`Source/engine/render/scrollrt.cpp:1313`) order** — this is the definitive "draw order" the port
must reproduce:
```
offset = {}; CalcFirstTilePosition(startPosition, offset);
DrawGame(out, startPosition, offset);
if (AutomapActive) DrawAutomap(out.subregionY(0, gnViewportHeight));
[_DEBUG grid / debug text]
DrawItemNameLabels(out)        // QoL item labels
DrawMonsterHealthBar(out)
DrawFloatingNumbers(out, startPosition, offset)   // damage numbers
[store text, inventory, spellbook, durability, level button, char, quest log, stash, visual store,
 unique info box, quest text, spell list, gold split/withdraw, help, chat log, death screen, pause menu,
 diablo msg, plr msg, gamemap menu, help screen (doom_draw), info box]
UpdateLifeManaPercent()
DrawLifeFlaskUpper(out); DrawManaFlaskUpper(out)
```
`DrawGame` (`Source/engine/render/scrollrt.cpp:1229`) does, in order:
```
out = fullOut.subregionY(0, zoom ? (gnViewportHeight+1)/2 : gnViewportHeight)   // zoom renders at half res
columns = tileColumns; rows = tileRows
if (CanPanelsCoverView() && (leftPanel||rightPanel)) columns -= zoom ? 2 : 4
UpdateMissilesRendererData()                       // interpolates missiles + buckets by rendering tile
if (MyPlayer->isWalking()) expand rows/columns by direction  (North/South: rows+=2; NE: cols+1 rows+2; E/W: cols+1;
                                                              SE/SW/NW: cols+1 rows+1)
Lightmap lightmap = Lightmap::build(perPixelLighting, position, offset, gnScreenWidth, gnViewportHeight,
                                    rows, columns, out.at(0,0), out.pitch(), LightTables,
                                    FullyLitLightTable, FullyDarkLightTable, dLight, MicroTileLen)
DrawFloor(out, lightmap, position, offset, rows, columns)
DrawTileContent(out, lightmap, position, offset, rows, columns)
DrawOOB(out, lightmap, position, offset, rows, columns)
if (zoom) Zoom(fullOut.subregionY(0, gnViewportHeight))
```

**4. Row-walk (the isometric scan order).** Identical in `DrawFloor`/`DrawTileContent`/`DrawOOB`
(`Source/engine/render/scrollrt.cpp:955`, `:992`, `:1088`):
```
for i in 0..rows-1:
  for j in 0..columns-1:
     tilePos  += Direction::East        # (+1,-1)
     bufferPos.x += TILE_WIDTH
  tilePos  += Direction::West * columns # rewind
  bufferPos.x -= columns * TILE_WIDTH
  bufferPos.y += TILE_HEIGHT/2          # 16
  if (i & 1) == 1: tilePos.x++; columns--;          bufferPos.x += TILE_WIDTH/2   #  32
  else:            tilePos.y++; columns++;          bufferPos.x -= TILE_WIDTH/2   # -32
```
`Direction::East = {+1,-1}`, `West = {-1,+1}`, `North = {0,-1}`, `South = {0,+1}`, `NorthEast = {+1,0}`,
`NorthWest = {0,-1}`… conversion table: `Source/engine/displacement.hpp:247` `fromDirection`.
`DrawTileContent` first does `rows += MicroTileLen` (`Source/engine/render/scrollrt.cpp:995`) so tall walls from
rows below can still overwrite. `DrawOOB` uses `rows + 5` (`Source/engine/render/scrollrt.cpp:1089`).

**5. Viewport geometry** (`CalcViewportGeometry`, `Source/engine/render/scrollrt.cpp:1675`, at 640×480/zoom off):
```
zoomFactor = zoom ? 2 : 1
screenWidth  = gnScreenWidth  / zoomFactor      # 640
screenHeight = gnScreenHeight / zoomFactor      # 480
panelHeight  = mainPanel.height / zoomFactor    # 128
pixelsToPanel = screenHeight - panelHeight      # 352
playerPosition = { screenWidth/2 , pixelsToPanel/2 }          # {320, 176}
tilesToTop  = (playerPosition.y + TILE_HEIGHT - 1) / TILE_HEIGHT  # (176+31)/32 = 6
tilesToLeft = (playerPosition.x + TILE_WIDTH - 1) / TILE_WIDTH  # (320+63)/64 = 5
startPosition = playerPosition - { tilesToLeft*64, tilesToTop*32 }   # {-0, -16} → {0,-16}
tileShift = North*tilesToTop + West*tilesToLeft
if (tilesToLeft*64 >= playerPosition.x) { startPosition += {32,-16}; tileShift += NorthEast }
elif (tilesToTop*32 < playerPosition.y) { startPosition += {0,-32}; tileShift += North }
tileOffset = { startPosition.x - 32, startPosition.y + 16 - 1 }
renderStart = startPosition - {32, 16}
tileRows    = (viewportHeight - renderStart.y + 16 - 1) / 16
tileColumns = (screenWidth     - renderStart.x + 64 - 1) / 64
```
These exact formulas have golden values already in `test/scrollrt_test.cpp:13` (tiles-in-view) and
`:80` (tile offset) — **port the test first, the formula second.**

**6. Zoom mode.** `Graphics.zoom` renders the world into a 320×240 region and then `Zoom()`
(`Source/engine/render/scrollrt.cpp:1121`) doubles it *in place, back-to-front* (byte-wise: for each of the top
`h/2` source rows, duplicate every pixel horizontally, then `memcpy` the doubled row into the row below). In
Ebitengine this is simply "draw the world buffer with `FilterNearest` at 2×", but note the C++ zoom only doubles the
*viewport*, not the panels, and the mouse coordinates are halved (`AlterMousePositionViaZoom`,
`Source/cursor.cpp:682`).

**7. Palette lifecycle.**
- `LoadPalette(path)` (`Source/engine/palette.cpp:175`): reads a 768-byte `.pal` (256 × RGB, 3 bytes each,
  verified on `/tmp/dgx/levels/l1data/l1_1.pal` → `size 768, first 6 bytes 00 00 00 68 68 7A`) into
  `logical_palette` (`std::array<SDL_Color,256>`).
- `UpdateSystemPalette` (`:152`) applies **brightness tone mapping** (`ApplyGlobalBrightness`, `:118`) producing
  `system_palette`, then `SystemPaletteUpdated(first,ncolor)` pushes it to the SDL surface palette and
  `RedrawEverything()`.
  Tone curve: `a = -(brightness/100)*2.0`; `f(x) = a*x² + (1-a)*x` over `x = c/255`, output
  `uint8((clamp(f,0,1)*255)+0.5)` (`Source/engine/palette.cpp:96-110`). Default brightness → identity.
- Fades (`:271`/`:324`): time-driven; `ApplyFadeLevel(fadeval, dst, src)` = `dst[i] = (fadeval*src[i])/256` per
  channel (`:140`). Loop: `for (uint32_t i = 0; i < 256; i = fr * (elapsed_ms) / 50)` with `fr *= 3`; each step
  re-sets the palette and presents. **Fade = whole-palette multiply, not a screen overlay.**
- Cycling (`palette_update_caves` `:363`, `_crypt` `:373`, `_hive` `:391`): `CycleColors(from,to)` rotates
  `logical_palette[from..to]`, `system_palette[from..to]`, **all rows and columns of
  `paletteTransparencyLookup`** in the same range (`Source/engine/palette.cpp:52-64`). Called from the tick loop
  (`Source/diablo.cpp:3521`), i.e. after ticks, not after frames.
- There is **no gamma** in this fork: only the brightness tone-map above.
- **No `SetPalette` function exists in this tree**; the equivalents are `LoadPalette` (`Source/engine/palette.cpp:175`),
  `UpdateSystemPalette` (`:152`), `SystemPaletteUpdated` (`:170`) and `SetLogicalPaletteColor(i,color)` (`:406`,
  which also refreshes one row of the blending LUT).

### Ebitengine mapping
- `Layout(outsideW, outsideH) (640, 480)` **fixed** (`Source/options.h:41`). Ebitengine then letterboxes/scales the
  logical screen; set `ebiten.SetScreenScalingPosition(...)`/`SetFilter(ebiten.FilterNearest)` for the vanilla
  integer-scale behaviour (`SDL_LOGICAL_PRESENTATION_INTEGER_SCALE`, `Source/utils/display.cpp:734`).
  For user-selected non-integer scaling use `ebiten.FilterPixelated` (v2.9+) which is "nearest with crisp edges";
  the C++ "bilinear scaling" quality option has no exact 8-bit equivalent and is declared *cosmetic, not ported*.
- Buffers (all allocated once at startup, never per frame):
  1. `world *ebiten.Image` 640×480 — floor + tile content + entities + automap (the `out` region passed to
     `DrawGame` is a *subregion* of the same buffer in C++; we can use one buffer and skip the panel rows).
  2. `ui *ebiten.Image` 640×480 — panels, flasks, menus, text, cursor. Separating world/UI is a **Go-only**
     optimisation: it lets us skip re-drawing the world when only a panel component is dirty (mirrors the
     `PanelDrawComponent` flags of `Source/engine/backbuffer_state.hpp:41`). It must not change pixels.
  3. `final *ebiten.Image` (or draw straight to `screen`) = `world` then `ui`, one `DrawImage` each.
- The C++ "index8 buffer + palette" becomes: keep an `indices []uint8` CPU mirror? **No.** Keep
  *index textures on the GPU* and one `palette *ebiten.Image` (256×1 RGBA) updated on every palette change
  (`UpdateSystemPalette`, cycling, fades). See G4/G10 for the shader contract.
- `Draw()` must contain **zero** allocation and zero `At()/ReadPixels()` (perf tips:
  https://ebitengine.org/en/documents/performancetips.html "Don't call (*Image).ReplacePixels too much").
- Frame pacing: Ebitengine `Update()` is the game tick driver (20 TPS logic inside, see dg-epic-01); `Draw()` runs
  at 60 FPS and computes `progress = ProgressToNextGameTick ∈ [0,128)` — exactly the C++ model where
  `nthread_UpdateProgressToNextGameTick()` is called from the draw path (`Source/engine/render/scrollrt.cpp:1866`).

### Parity test
`TestViewportGeometry` table-driven over the four cases in `test/scrollrt_test.cpp` (640×480, 640×480 zoom,
960×540, 640×512, 768×480 zoom) asserting `tileRows`, `tileColumns`, `tileOffset.x/y` equal to the C++ numbers.
`TestFrameOrder`: instrumented renderer recording an ordered list of (layer, kind) ops; assert the list equals the
C++ order encoded from `DrawAndBlit` + `DrawView` above.
`TestPaletteFadeLevel`: port `ApplyFadeLevel` and assert `dst[i] == uint8(fadeval*src[i]/256)` for all
`fadeval ∈ {0,1,255,256}` against a C++-dumped golden (1024-byte array per step).

### Open questions
- Does the Go port need the 1920×1080 "big UI" mode (Volume I §7 mentions it)? The 640×480 decision here assumes no.
- `PanelDrawComponent` partial redraws save CPU in C++ surface mode; in Go the world re-draw is one batch — confirm
  we still want the world/ui split (adds a 640×480 offscreen; offscreens don't batch with the atlas).

---

## G2. LEVEL TILES (dungeon micro-tiles, transparency, tall tiles)

### Sources (file:line)
- Source/levels/dun_tile.hpp:7 — `#define TILE_WIDTH 64`
- Source/levels/dun_tile.hpp:8 — `#define TILE_HEIGHT 32`
- Source/levels/dun_tile.hpp:19 — `enum class TileType : uint8_t` (Square, TransparentSquare, LeftTriangle,
  RightTriangle, LeftTrapezoid, RightTrapezoid)
- Source/levels/dun_tile.hpp:83 — `struct LevelCelBlock { uint16_t data; }`
- Source/levels/dun_tile.hpp:93 — `type() = (data & 0x7000) >> 12`
- Source/levels/dun_tile.hpp:100 — `frame() = data & 0xFFF` (1-based frame index)
- Source/levels/dun_tile.hpp:106 — `enum class TileProperties` (Solid 1, BlockLight 2, BlockMissile 4,
  Transparent 8, TransparentLeft 0x10, TransparentRight 0x20, Trap 0x80)
- Source/levels/dun_tile.hpp:119 — `struct MICROS { LevelCelBlock mt[16]; }`
- Source/levels/dun_tile.hpp:123 — `DunFrameWidth = TILE_WIDTH/2 = 32`
- Source/levels/dun_tile.hpp:126 — `DunFrameHeight = TILE_HEIGHT = 32`
- Source/levels/dun_tile.hpp:128 — `DunFrameTriangleHeight = 31`
- Source/levels/dun_tile.hpp:131 — `ReencodedTriangleFrameSize = 544 - 32` (=512)
- Source/levels/dun_tile.hpp:132 — `ReencodedTrapezoidFrameSize = 800 - 16` (=784)
- Source/levels/dun_tile.hpp:138 — `CalculateSpriteTileCenterX(width) = (width - TILE_WIDTH)/2`
- Source/levels/dun_tile_data.hpp:29 — `#define MAXTILES 1379`
- Source/levels/dun_tile_data.hpp:33 — `enum class DungeonFlag` (Missile 1, Visible 2, DeadPlayer 4, Populated 8,
  MissileFireWall 0x10, MissileLightningWall 0x20, Lit 0x40, Explored 0x80)
- Source/levels/dun_tile_data.hpp:57 — `struct MegaTile { uint16_t micro1..micro4; }`
- Source/levels/dun_tile_data.hpp:80 — `OptionalOwnedClxSpriteList pSpecialCels;`
- Source/levels/dun_tile_data.hpp:82 — `std::unique_ptr<MegaTile[]> pMegaTiles;`
- Source/levels/dun_tile_data.hpp:83 — `std::unique_ptr<std::byte[]> pDungeonCels;`
- Source/levels/dun_tile_data.hpp:87 — `TileProperties SOLData[MAXTILES];`
- Source/levels/dun_tile_data.hpp:99 — `uint_fast8_t MicroTileLen;`
- Source/levels/dun_tile_data.hpp:101 — `int8_t TransVal;`
- Source/levels/dun_tile_data.hpp:103 — `std::array<bool, 256> TransList;`
- Source/levels/dun_tile_data.hpp:105 — `uint16_t dPiece[MAXDUNX][MAXDUNY];`
- Source/levels/dun_tile_data.hpp:107 — `MICROS DPieceMicros[MAXTILES];`
- Source/levels/dun_tile_data.hpp:108 — `int8_t dTransVal[][]`, `:110 uint8_t dLight[][]`, `:112 dPreLight`,
  `:114 dFlags`, `:116 int8_t dPlayer`, `:119 dMonster (int16)`, `:124 int8_t dCorpse`, `:129 dObject`,
  `:135 int8_t dSpecial`
- Source/levels/dun_tile_data.hpp:145 — `InDungeonBounds`
- Source/levels/dun_tile_data.hpp:200 — `IsTileLit` (dFlags & Lit)
- Source/levels/dun_tile_data.hpp:212 — `TileHasAny(coords, prop) = HasAnyOf(SOLData[dPiece[x][y]], prop)`
- Source/levels/dun_tile_data.cpp:68 — `LoadMinData` (`levels\l1data\l1.min` etc.)
- Source/levels/dun_tile_data.cpp:104 — `LoadLevelSOLData` (`l1.sol` + per-level hand fixes)
- Source/levels/dun_tile_data.cpp:165 — `SetDungeonMicros`
- Source/levels/dun_tile_data.cpp:167 — `microTileLen = 10; blocks = 10;`
- Source/levels/dun_tile_data.cpp:172 — town: `microTileLen = 16; blocks = 16;`
- Source/levels/dun_tile_data.cpp:175 — hell: `microTileLen = 12; blocks = 16;`
- Source/levels/dun_tile_data.cpp:184 — the MIN→DPieceMicros loop (the `blocks - 2 + (block & 1) - (block & 0xE)`
  index scramble) + `frameToTypeMap`
- Source/levels/dun_tile_data.cpp:199 — `ReencodeDungeonCels(...)`
- Source/levels/dun_tile_data.cpp:201 — `ComputeCelBlockAdjustments` + frame-number fixups
- Source/levels/reencode_dun_cels.cpp:24 — `ReencodeDungeonCelsLeftTriangleLower`
- Source/levels/reencode_dun_cels.cpp:38 — `ReencodeDungeonCelsLeftTriangle`
- Source/levels/reencode_dun_cels.cpp:57 — `ReencodeDungeonCelsRightTriangleLower`
- Source/levels/reencode_dun_cels.cpp:72 — `ReencodeDungeonCelsRightTriangle`
- Source/levels/reencode_dun_cels.cpp:89 — `ReencodeDungeonCelsLeftTrapezoid`
- Source/levels/reencode_dun_cels.cpp:96 — `ReencodeDungeonCelsRightTrapezoid`
- Source/levels/reencode_dun_cels.cpp:103 — `RenderTransparentSquare` (the 32×32 RLE→flat decode)
- Source/levels/reencode_dun_cels.cpp:118 — `ExtractFoliageLeftTriangle`
- Source/levels/reencode_dun_cels.cpp:133 — `ExtractFoliageRightTriangle`
- Source/levels/reencode_dun_cels.cpp:146 — `ExtractFoliageTransparentSquare`
- Source/levels/reencode_dun_cels.cpp:172 — `ReencodeFloorWithFoliage`
- Source/levels/reencode_dun_cels.cpp:183 — `GetReencodedSize`
- Source/levels/reencode_dun_cels.cpp:211 — `ReencodeDungeonCels` (writes the new LE frame table)
- Source/levels/reencode_dun_cels.cpp:304 — `ComputeCelBlockAdjustments`
- Source/engine/render/dun_render.hpp:28 — `enum class MaskType { Solid, Transparent, Right, Left }` (with ASCII art
  of the exact triangle masks at `:36` and `:58`)
- Source/engine/render/dun_render.hpp:117 — `RenderTileFrame(out, lightmap, position, tile, src, height, maskType, tbl)`
- Source/engine/render/dun_render.hpp:123 — `GetDunFrame(dungeonCelData, frame)` = `Swap32LE(table[frame])`
- Source/engine/render/dun_render.hpp:132 — `GetDunFrameFoliage` = frame + `ReencodedTriangleFrameSize`
- Source/engine/render/dun_render.hpp:147 — `RenderTile` (height = 31 for triangles else 32)
- Source/engine/render/dun_render.hpp:162 — `RenderTileFoliage` (draws at `y-16`, height 16)
- Source/engine/render/dun_render.cpp:41 — `Width = 32`, `:44 Height = 32`, `:47 LowerHeight = 16`,
  `:50 TriangleUpperHeight = 15`, `:54 XStep = 2`
- Source/engine/render/dun_render.cpp:92 — `PrefixIncrement<Left> = 2`, `:95 <Right> = -2`
- Source/engine/render/dun_render.cpp:98 — `InitialPrefix<Mask> = incr >= 0 ? -32 : 64`
- Source/engine/render/dun_render.cpp:102 — `InitPrefix(y) = InitialPrefix + incr*y`
- Source/engine/render/dun_render.cpp:107 — `enum class LightType { FullyDark, PartiallyLit, FullyLit, PerPixel }`
- Source/engine/render/dun_render.cpp:124 — `RenderLineOpaque<FullyDark>` → `FillBytesUpTo32(dst,n,0)`
- Source/engine/render/dun_render.cpp:130 — `<FullyLit>` → memcpy
- Source/engine/render/dun_render.cpp:140 — `<PartiallyLit>` → `BlitPixelsWithMap(dst,src,n,tbl)`
- Source/engine/render/dun_render.cpp:148 — `<PerPixel>` → `BlitPixelsWithLightmap`
- Source/engine/render/dun_render.cpp:159 — `RenderLineTransparent<FullyDark>` → `BlitFillBlended(dst,n,0)`
- Source/engine/render/dun_render.cpp:166 — `<FullyLit>` → `BlitPixelsBlended`
- Source/engine/render/dun_render.cpp:172 — `<PartiallyLit>` → `BlitPixelsBlendedWithMap`
- Source/engine/render/dun_render.cpp:204 — `RenderLineTransparentAndOpaque<OpaqueFirst>`
- Source/engine/render/dun_render.cpp:220 — `RenderLine<Light,Mask>` (prefix → transparent/opaque split)
- Source/engine/render/dun_render.cpp:252 — `struct Clip`, `:261 CalculateClip`
- Source/engine/render/dun_render.cpp:276 — `RenderSquareFull` (`dst -= dstPitch` per row → bottom-to-top)
- Source/engine/render/dun_render.cpp:299 — `RenderTransparentSquareFull` (int8 RLE walk)
- Source/engine/render/dun_render.cpp:333 — `RenderTransparentSquareClipped` (prefix/overshoot handling)
- Source/engine/render/dun_render.cpp:422 — `CalculateDiamondClipY`
- Source/engine/render/dun_render.cpp:447 — `CalculateTriangleSourceSkipLowerBottom = 2*n*(n+1)/2`
- Source/engine/render/dun_render.cpp:466 — `RenderTriangleLower` (unrolled: widths 2,4,…,32; src offsets
  0,2,6,12,20,…,240; `src += 272` at the end)
- Source/engine/render/dun_render.cpp:481 — `RenderTriangleUpper` (widths 30,28,…,2; src 0,30,58,…,238)
- Source/engine/render/dun_render.cpp:500 — `RenderLeftTriangleLower` (`dst += XStep*(LowerHeight-1)`,
  line offset `dstPitch + XStep`)
- Source/engine/render/dun_render.cpp:908 — `RenderTileType` dispatch
- Source/engine/render/dun_render.cpp:1046 — `RenderTileFrame`
- Source/engine/render/dun_render.cpp:1085 — `world_draw_black_tile`
- Source/engine/render/scrollrt.cpp:549 — `DrawCell` (the 16-micro loop, masks, foliage)
- Source/engine/render/scrollrt.cpp:599 — `LevelCelBlock{pMap->mt[0]}` first-micro + `getFirstTileMaskLeft`
- Source/engine/render/scrollrt.cpp:615 — second micro with `RightFrameDisplacement = {32,0}`
- Source/engine/render/scrollrt.cpp:628 — `targetBufferPosition.y -= TILE_HEIGHT;`
- Source/engine/render/scrollrt.cpp:630 — `for (i = 2, n = MicroTileLen; i < n; i += 2)` (pairs)
- Source/engine/render/scrollrt.cpp:680 — `DrawFloorTile` (mt[0] LeftTriangle + mt[1] RightTriangle,
  `GetDunFrame`, `DunFrameTriangleHeight`)
- Source/engine/render/scrollrt.cpp:949 — tall-tile/`dSpecial` arch pass (`pSpecialCels`)
- Source/engine/render/scrollrt.cpp:1047 — `DrawDirtTile` (out-of-bounds fallback)
- Source/engine/render/scrollrt.cpp:1088 — `DrawOOB`
- Source/engine/render/blit_impl.hpp:20 — `BlitFillDirect` / `:26 BlitPixelsDirect`
- Source/engine/render/blit_impl.hpp:45 — `BlitFillWithMap` (`memset(dst, colorMap[color], length)`)
- Source/engine/render/blit_impl.hpp:50 — `BlitPixelsWithMap` (`dst[i]=colorMap[src[i]]`)
- Source/levels/gendung.cpp:375 — `DRLG_InitTrans` (`memset dTransVal`, `TransList={}`, `TransVal=1`)
- Source/levels/gendung.cpp:382 — `DRLG_RectTrans`
- Source/levels/gendung.cpp:396 — `DRLG_MRectTrans`
- Source/levels/gendung.cpp:205 — `FindTransparencyValues` flood-fill of `dTransVal`
- Source/levels/gendung.cpp:424 — `.dun` transparent layer fills `dTransVal[16+i][16+j]`
- Source/levels/gendung.cpp:636 — `FloodTransparencyValues`
- Source/objects.cpp:1066 — `dSpecial[pos] = 7` (door arch indices), `:1071 =8`, `:1076 =5`, `:1080 =6`
- Source/diablo.cpp:1359 — `LoadLvlGFX()` — the exact per-level `*.cel/*.til/*s` triple
- Source/levels/dun_tile_data.cpp:143 — L3 SOL patch; `:151` L4 `SOLData[210] = None`; `:159` crypt `SOLData[142]`

### Formats & algorithms

**1. Dungeon CEL (macro→micro frame blob, big-endian frame table).**
Verified on the real file `/tmp/dgx/levels/l1data/l1.cel`:
```
bytes 0..3    u32 LE  = 0x0000045f = 1119   -> frame count (LE!)
bytes 4..(4+4*n)  u32 LE frame offsets, ascending, last == file size (738836)  [verified: ascending? True,
                                         offs[n]==size]
frame i begins at offs[i]; NO per-frame header (raw pixel runs only)
```
Caveat: `Source/levels/reencode_dun_cels.cpp:188` reads these offsets with **`Swap32LE`** (i.e. big-endian swap)
because the retail `*.cel` frame table is documented as big-endian in vanilla, while devilutionX's
`GetDunFrame` (`Source/engine/render/dun_render.hpp:123`) also uses `Swap32LE` — but after `ReencodeDungeonCels`
rewrites the table with `WriteLE32` (`Source/levels/reencode_dun_cels.cpp:216`). **This is a real trap: the input
table is BE-swapped, the re-encoded output table is LE, and `GetDunFrame` still swaps.** Read
`GetReencodedSize`/`ReencodeDungeonCels` again when porting; do not "fix" it. Empirically on our GOG file the table
is already ascending-LE and `Swap32LE` yields the same values only because… it doesn't. See "Open questions".

Frame pixel formats (`Source/levels/dun_tile.hpp:19` doc comments, all **bottom row first**):

| TileType | encoding | size after re-encode |
|---|---|---|
| `Square` (0) | flat `32*32` index bytes | 1024 |
| `TransparentSquare` (1) | 32 rows of int8 RLE, runs never cross a row: `v>0` → copy v pixels; `v<0` → skip `-v` transparent | as-is (`Source/levels/reencode_dun_cels.cpp:226`) |
| `LeftTriangle` (2) | 31 rows, widths 2,4,…,32,…,2; **2 unused padding bytes before every even row (bottom-counted)** | 512 (`544-32`) |
| `RightTriangle` (3) | same, padding **after** every even row | 512 |
| `LeftTrapezoid` (4) | triangle-lower part (16 rows, widths 2..32) with padding, then a flat 32×16 rect on top | 784 (`800-16`) |
| `RightTrapezoid` (5) | same with trailing padding | 784 |

**2. Worked byte example — retail `LeftTriangle` frame #48 of `l1.cel`** (`/tmp/dgx/levels/l1data/l1.cel`,
offset 34044, length 544):
```
0000: 00 00 29 29 2a 79 29 29 00 00 28 77 77 29 77 77 29 75 75 74 75 76 76 77 ...
      ^^^^^^^^ padding(2)  ^^^^^^^^ row0 width 2 (29 29) ... wait: width sequence is
      i=0: skip 2 bytes, read 2 px = 29 29      (row 0, bottom-most, width 2)
      i=0 cont: read 2 px = 2a 79               (width 4? no: two rows per loop, see below)
```
Use the exact decoder (`Source/levels/reencode_dun_cels.cpp:24`) instead of eyeballing:
```
width = 0
for i in 0..7:                 # lower half, growing 2→32
    src += 2                   # discard padding
    width += 2                 # 2
    copy width bytes           # -> bytes [2,4)   = 29 29
    src += width
    width += 2                 # 4
    copy width bytes           # -> bytes [8,12)  = 28 77 77 29 ... (verify with the real dump)
    src += width
for i in 0..6:                 # upper half, shrinking 30→2
    src += 2
    width -= 2
    copy width
    src += width
    width -= 2
    copy width
    src += width
src += 2; width -= 2; copy width   # final row (top, width 2)
```
Total copied = 2·(1+2+…+16) + 2·(15+14+…+1) + 2 = 544 bytes with 32 padding bytes dropped → **512 bytes output**
(`ReencodedTriangleFrameSize`, `Source/levels/dun_tile.hpp:131`). Rendering walks those 512 bytes with
`RenderTriangleLower` (`Source/engine/render/dun_render.cpp:466`) whose literal src offsets are
`0, 2, 6, 12, 20, 30, 42, 56, 72, 90, 110, 132, 156, 182, 210, 240` (= `i*(i+1)`) with widths `2,4,…,32`
— i.e. **row k is drawn at screen `dst - k*(pitch) + ... ` stepping up-left/right by `XStep=2` per row**.

**3. Worked byte example — `TransparentSquare` (foliage) decode** (`Source/levels/reencode_dun_cels.cpp:103`):
```
for i in 0..31:                       # 32 rows, bottom-to-top
    dst -= 2*DunFrameWidth            # rewind to row start (dst was left after previous row)
    drawWidth = 32
    while drawWidth > 0:
        v = int8(*src++)
        if v > 0: memcpy(dst, src, v); src += v
        else:     v = -v              # transparent: nothing read
        dst += v; drawWidth -= v
```
So a `TransparentSquare` run stream for one row of 32 px is e.g.
`[-4][+10 px…][-2][+8 px…][-8]` = 4 transparent, 10 solid, 2 transparent, 8 solid, 8 transparent.

**4. Foliage extraction (`ReencodeFloorWithFoliage`, `Source/levels/reencode_dun_cels.cpp:172`).** A floor tile
whose block is `TransparentSquare` is *split*: the solid pixels in its top 16 rows are re-encoded as a new
triangle-shaped frame (left or right per `info.isFloorLeft()`) plus a remaining `TransparentSquare` RLE, and the
original frame's pixels below the diagonal are zeroed. At draw time `DrawCell` renders the floor triangle and then,
*only if the tile is a floor*, renders the extracted foliage via `RenderTileFoliage` at `y-16`, height 16,
`MaskType::Solid` (`Source/engine/render/dun_render.hpp:162`, `Source/engine/render/scrollrt.cpp:697`).

**5. MIN (micro index) file → `DPieceMicros`.** `*.min` = array of `u16` **little-endian-swapped** values
(`Swap16LE`, `Source/levels/dun_tile_data.cpp:186`), `blocks` per piece where blocks = 10 (dungeon levels) or 16
(town/hell). Each u16 *is* a `LevelCelBlock`:
```
bits 0..11  frame index, 1-based (0 = no block)
bits 12..14 TileType
```
The scramble `pieces[blocks - 2 + (block & 1) - (block & 0xE)]` (`Source/levels/dun_tile_data.cpp:186`) is the
on-disk ↔ in-memory reordering and must be transcribed **literally** (it is `blocks-2+(b&1)-(b&~0xE)`… evaluate for
blocks=10: b=0→8, b=1→9, b=2→2, b=3→3, … b=8→0, b=9→1). Verified counts on real files:
`/tmp/dgx/levels/l1data/l1.min` = 9060 bytes = 4530 u16 = **453 pieces × 10 blocks** and
`l1.sol` = 453 bytes (one `TileProperties` per piece) — i.e. `tileCount/blocks == sol size`.
`town.min` = 40256 bytes = 20128 u16 = 1258 pieces × 16 blocks; `town.sol` = 1258 bytes. Same relationship.

**6. `.til` (megatile) file** — `/tmp/dgx/levels/l1data/l1.til` is 1648 bytes = 206 `MegaTile`s × 8 bytes;
`MegaTile{micro1..micro4}` (`Source/levels/dun_tile_data.hpp:57`) are u16 (BE-swapped at use:
`Swap16LE(mega.micro1)`, `Source/levels/gendung.cpp:585`). `DRLG_LPass3` (`Source/levels/gendung.cpp:578`) stamps a
megatile into `dPiece` 2×2: `dPiece[i][j]=v1, dPiece[i+1][j]=v2, dPiece[i][j+1]=v3, dPiece[i+1][j+1]=v4` for the
whole map (`:590`) then patches the `.dun`-defined area at `xx=16+2*i, yy=16+2*j` (`:602`).

**7. Tile → screen transform (the y-offset for tall tiles).**
```
Displacement::worldToScreen() = { (dy - dx) * 32, -(dy + dx) * 16 }   Source/engine/displacement.hpp:151
screenToWorld()               = { (2*dy + dx)/-64, (2*dy - dx)/-64 }  Source/engine/displacement.hpp:164
```
`targetBufferPosition` for a tile is maintained **incrementally** by the row-walk (G1 #4), not recomputed.
For tile (dx,dy) relative to the first tile: `screenX = (dTile.x - dTile.y) * TILE_WIDTH/2`,
`screenY = (dTile.x + dTile.y) * TILE_HEIGHT/2`, and the *bottom* corner of the diamond is
`(tileOffset + screenY)`, so the tile's 64×32 bbox spans `y-31 .. y`.
Tall tiles: a wall's micro frames are stacked **upwards** — `DrawCell` starts at the tile's bottom pair `mt[0],mt[1]`
and then `targetBufferPosition.y -= TILE_HEIGHT` once per pair (`Source/engine/render/scrollrt.cpp:628` and `:643`),
iterating `i = 2 .. MicroTileLen-1` in steps of 2. So a cathedral wall of `MicroTileLen = 10` occupies
`10*32 = 320` px of vertical screen space above the tile's bottom corner, which is why
`DrawTileContent` adds `rows += MicroTileLen` and why the lightmap buffer is
`viewportHeight + TILE_HEIGHT*(microTileLen/2 + 1)` rows tall (`Source/engine/render/light_render.cpp:435`).

**8. Floor vs wall vs arch.**
- *floor test*: `IsFloor(p) = !TileHasAny(p, Solid|BlockMissile)` (`Source/engine/render/scrollrt.cpp:111`);
  `DrawFloor` only draws `mt[0]/mt[1]` of the piece as the two 32×31 triangles (`DrawFloorTile`, `:680`) using the
  tile's own `dLight` table.
- *wall test*: `IsWall(p) = !IsFloor(p) || dSpecial[p] != 0` (`Source/engine/render/scrollrt.cpp:115`).
- *arch/tall "special"*: `dSpecial[x][y]-1` indexes `pSpecialCels` (64 px wide frames, `levels/l1data/l1s` etc.)
  drawn **after all entities** in the same tile with `ClxDrawLight`/`ClxDrawLightBlended`
  (`Source/engine/render/scrollrt.cpp:907-943`). In town there are no arches; trees from `levels/towndata/towns`
  are drawn **one row later** to overlap correctly (`Source/engine/render/scrollrt.cpp:944-952`, reads
  `dSpecial[x-1][y-1]` at `bufferPos + {0,-TILE_HEIGHT}`).
- *wall-behind pre-pass* (`Source/engine/render/scrollrt.cpp:1012-1030`): if the tile is part of an x-axis wall and
  the two tiles behind it (`+{1,-1}`, `+{0,-1}`) are non-solid, draw the East neighbour's content first and set
  `skipNext` so the following tile in the row is skipped. This is what keeps sprites from poking through walls.

**9. Transparency, byte-level.** Two conditions AND together (`Source/engine/render/scrollrt.cpp:575`):
```
transparency = TileHasAny(tile, TileProperties::Transparent) && TransList[dTransVal[tile]]
```
`TransList` is a 256-entry bool array; `dTransVal` is a per-tile room id assigned by `DRLG_*Trans`
(`Source/levels/gendung.cpp:375-430`), flooded by `FindTransparencyValues` (`:205`) / `FloodTransparencyValues`
(`:636`). `DoVision` marks a room "see-through" the moment a ray passes it (`Source/lighting.cpp:200-203`:
`if (trans != 0) TransList[trans] = true;`). So **transparency is per-room and dynamic** — a room whose far side the
player can see renders its walls blended.

When `transparency` is on, each micro is rendered with `MaskType::Transparent` (whole frame blended) except the two
first micros of the tile which use the directional masks:
```
getFirstTileMaskLeft:  LeftTrapezoid|TransparentSquare -> (TransparentLeft ? MaskType::Left  : Solid)
                       LeftTriangle                    -> Solid
                       otherwise                       -> Transparent
getFirstTileMaskRight: RightTrapezoid|TransparentSquare-> (TransparentRight? MaskType::Right : Solid)
                       RightTriangle                   -> Solid
                       otherwise                       -> Transparent
```
(`Source/engine/render/scrollrt.cpp:577-601`.) `MaskType::Left`/`Right` are the diagonal-split masks documented
with ASCII art in `Source/engine/render/dun_render.hpp:36-75`; implemented by a *prefix* counter that starts at
`-32` (Left) / `64` (Right) and changes by `+2` / `-2` per row (`:92-104`), splitting each 32-px row into a
transparent part and an opaque part (`RenderLine`, `:220`, `RenderLineTransparentAndOpaque`, `:204`).

The actual per-pixel op for a transparent line is
`dstColor = paletteTransparencyLookup[dstColor][tbl[srcColor]]`
(`BlitPixelsBlendedWithMap`, `Source/engine/render/blit_impl.hpp:126`) — a **double table lookup, no float**.

**10. `reencode_dun_cels.cpp` purpose.** One-time level-load pass that (a) strips the 32/16 padding bytes from
triangle and trapezoid frames so the renderer's line walking is a straight `dstPitch`-per-row loop, (b) turns floor
`TransparentSquare` frames into (triangle + foliage) pairs, (c) compacts the frame table to only the frames actually
referenced (hence `ComputeCelBlockAdjustments`, `Source/levels/reencode_dun_cels.cpp:304`, which renumbers
`LevelCelBlock.frame()` in `DPieceMicros` afterwards, `Source/levels/dun_tile_data.cpp:202-215`).
It is a **performance/normalisation step, not a semantic one**, but it *changes frame indices*, so a Go port either
runs the identical re-encode (recommended — then `LevelCelBlock.frame()` values match C++ and golden tile tests are
directly comparable) or must apply the same adjustment table.

### Ebitengine mapping
- At load: for every dungeon CEL frame, decode micro frames into **one paletted (index8) `*ebiten.Image` per micro**
  laid out in a shared atlas, size 32×31 or 32×32, with the rows *flipped* (C++ stores bottom row first, Ebitengine
  images are top-row-first). Do the flip at decode time so the draw path is trivial.
- Draw: `SubImage` of the atlas + one `DrawImage` per micro with
  `op.GeoM = translate(targetX, targetY - h + 1)` and `op.ColorScale` identity; the palette lookup happens in a
  custom Kage fragment shader with two uniform-free inputs: the index texture (R channel = index) and a
  256×1 palette texture that already has `LightTables[dLight]` *baked in* for that light level. Two practical
  options:
  1. **16 palette textures** (one per light level), each `256×1`, regenerated when the logical palette changes
     (`palette[i] → palette[LightTables[level][i]]`). Choosing the palette per tile = swapping a bound image; a
     whole run of tiles with the same light level stays in one batch (perf tips: batch requires same blend+filter
     and same target; source images across the atlas are fine).
  2. index texture + palette texture + *third* texture = the 16×256 light table as a 256×16 image, sampled at
     `(index, lightLevel)`; then only one palette texture is needed and the light level is a per-draw uniform
     (`op.ColorScale.ScaleWithColor(...)` abusing an unused channel, or a `DrawRectShader` with per-tile uniforms —
     which **breaks batching**). Prefer option 1: no per-draw uniforms → maximal batching.
- `MaskType::Left/Right/Transparent` cannot be done with `ColorScale`; it needs either (a) per-mask atlases with an
  alpha channel pre-baked as "0 = opaque, 1 = blend" and a Kage shader that does `mix(dst, blend, mask)`, or
  (b) splitting each row into up to 2 `DrawImage`s (opaque part then blended part) which is exactly what C++ does.
  **Recommendation: (b) — geometric row splitting keeps the op-count predictable (≤ 2 per micro row, and only for the
  16 upper rows of the first pair) and needs no new shader**; but implement blending via the 256×256 blend LUT
  uploaded as a texture (G8) so the arithmetic stays byte-exact.
- Draw-call budget at 640×480: `tileRows ≈ 30`, `tileColumns ≈ 11` at worst (`:1718`), `MicroTileLen` up to 16, so
  worst-case micros = rows*columns*(MicroTileLen/2) ≈ but only non-solid floor/wall tiles draw → measure with the
  ported `DunRenderStats` counter (`Source/engine/render/dun_render.cpp:1015`). Target: < 4000 `DrawImage` calls in
  the world pass; with a single atlas and constant blend/filter this collapses to a handful of GPU commands.
- Tall-tile overlap: keep the C++ order exactly (tile loop, then `y -= TILE_HEIGHT` walking up). Do **not** try to
  sort by sprite screen position; the negative-`dPlayer`/`dMonster` trick (G6) exists precisely because ordering is
  tile-based.

### Parity test
`TestMinScramble`: for `blocks ∈ {10,16}`, assert `LevelCelBlock(blockIdx) == Swap16LE(min[blocks-2+(b&1)-(b&0xE)])`
for the first 200 pieces of a committed copy of `l1.min`/`town.min`.
`TestReencodeSizes`: assert output frame sizes 512/784/1024 and total buffer size equals
`GetReencodedSize` (`Source/levels/reencode_dun_cels.cpp:183`) for the whole `l1.cel` frame set.
`TestRenderTileGolden`: render one micro of each `TileType` × each `MaskType` × light levels {0, 8, 15} into a 64×64
`Surface` in C++ (a small gtest added to `test/dun_render_benchmark.cpp`'s harness) dumping the index buffer; the Go
test decodes the same region from the GPU via a one-off `ReadPixels` in the test (never in `Draw`) and compares bytes.
`TestRowWalk`: 5×5 map → the ordered list of (tile, bufferX, bufferY) must equal C++ (mirrors the dg-epic-03 first
acceptance criterion).

### Open questions
- **Endianness of the dungeon CEL frame table.** `GetDunFrame`/`reencode_dun_cels.cpp:188` use `Swap32LE`, but the
  GOG `l1.cel` frame table is already ascending little-endian (`offs == [4484, 4836, 5860, … , 738836] == filesize`),
  so on a little-endian host `Swap32LE` is a no-op and it works; on big-endian it would swap. The Go port should read
  **LE with an explicit comment** and add the BE CI job (`.github` runs `tools/run_big_endian_tests.sh`) as a parity
  check. Needs a decision: mirror `Swap32LE` semantics or standardise on LE.
- Does the port need `perPixelLighting` (`Source/engine/render/light_render.cpp`, marching-squares + half-space
  triangle rasteriser, ~350 lines)? Volume I §7 marks it optional; recommend: **not ported for M2**, and
  `Lightmap` in Go is a no-op struct with the table path only.

---

## G3. SPRITE FORMATS (CEL / CL2 / CLX / PCX)

### Sources (file:line)
- Source/engine/clx_sprite.hpp:17 — the 6-byte CLX frame header table (header size, width, height, all u16 LE)
- Source/engine/clx_sprite.hpp:38 — `class ClxSprite`
- Source/engine/clx_sprite.hpp:55 — `class ClxSprite`; `width()` = `LoadLE16(&data_[2])`, `height()` = `LoadLE16(&data_[4])`
- Source/engine/clx_sprite.hpp:67 — `pixelData() = &data_[LoadLE16(data_)]`
- Source/engine/clx_sprite.hpp:74 — `pixelDataSize() = dataSize - LoadLE16(data)` (set in ctor, `:41`)
- Source/engine/clx_sprite.hpp:95 — `class ClxSpriteList`
- Source/engine/clx_sprite.hpp:103 — `numSprites() = LoadLE32(data_)`
- Source/engine/clx_sprite.hpp:109 — `operator[](i)` = `ClxSprite{ &data_[begin], end-begin }`
- Source/engine/clx_sprite.hpp:116 — `spriteOffset(i) = LoadLE32(&data_[4 + i*4])`
- Source/engine/clx_sprite.hpp:122 — `dataSize() = LoadLE32(&data_[4 + numSprites*4])`
- Source/engine/clx_sprite.hpp:187 — `class ClxSpriteSheet` (list of lists)
- Source/engine/clx_sprite.hpp:205 — `sheetOffset(i) = LoadLE32(&data_[4*i])`
- Source/engine/clx_sprite.hpp:404 — `GetNumListsFromClxListOrSheetBuffer` (list vs sheet sniff)
- Source/engine/clx_sprite.hpp:418 — `class ClxSpriteListOrSheet`
- Source/engine/clx_sprite.hpp:288 — `OwnedClxSpriteList` (owns `unique_ptr<uint8_t[]>`)
- Source/utils/clx_decode.hpp:12 — `IsClxOpaque(c) = c >= 0x80`
- Source/utils/clx_decode.hpp:17 — `GetClxOpaquePixelsWidth(c) = -int8(c)` (1..65)
- Source/utils/clx_decode.hpp:23 — `IsClxOpaqueFill(c) = c <= 0xBE`
- Source/utils/clx_decode.hpp:28 — `GetClxOpaqueFillWidth(c) = 0xBF - c` (1..63)
- Source/utils/clx_decode.hpp:42 — `struct SkipSize { wholeLines; xOffset; }`
- Source/utils/clx_decode.hpp:49 — `GetSkipSize(remainingWidth, srcWidth)` — the overrun/line-wrap rule
- Source/utils/clx_encode.hpp:14 — `ClxFrameHeaderSize = 6`
- Source/utils/clx_encode.hpp:16 — `AppendClxTransparentRun` (0x7F-chunked)
- Source/utils/clx_encode.hpp:29 — `AppendClxFillRun` (`0x80`+color chunks of 0x3F, tail `0xBF-width`)
- Source/utils/clx_encode.hpp:43 — `AppendClxPixelsRun` (`0xBF`+0x41 px chunks, tail `256-width`)
- Source/utils/clx_encode.hpp:60 — `AppendClxPixelsOrFillRun` (`MinFillRunLength = 3`, tail threshold 2)
- Source/utils/cel_to_clx.cpp:25 — `IsCelTransparent(c) = c >= 0x80`, `:30 GetCelTransparentWidth = -int8(c)`
- Source/utils/cel_to_clx.cpp:35 — `CelToClx` entry
- Source/utils/cel_to_clx.cpp:41 — group sniff: `LoadLE32(&data[maybeNumFrames*4+4]) != size` ⇒ it's a *group list*
- Source/utils/cel_to_clx.cpp:81 — `CelFrameHeaderSize = 10`, `celFrameHasHeader = LoadLE16(src)==10`
- Source/utils/cel_to_clx.cpp:95 — per-line loop `for (remainingCelWidth = frameWidth; remainingCelWidth != 0;)`
- Source/utils/cel_to_clx.cpp:104 — `++frameHeight` per line (CEL: **height = number of lines, no overflow**)
- Source/utils/cl2_to_clx.cpp:17 — `Cl2ToClx(data, size, widthOrWidths, clxData)`
- Source/utils/cl2_to_clx.cpp:23 — same group sniff as CEL
- Source/utils/cl2_to_clx.cpp:55 — `frameWidth = widthOrWidths.HoldsPointer() ? ptr[frame-1] : value`
- Source/utils/cl2_to_clx.cpp:63 — CLX header write (`ClxFrameHeaderSize`, width)
- Source/utils/cl2_to_clx.cpp:68 — `src = frameBegin + LoadLE16(frameBegin)` (skip the CL2 frame header)
- Source/utils/cl2_to_clx.cpp:71-93 — the 3-way run decode + `xOffset`/`remainingWidth` walk
- Source/utils/cl2_to_clx.cpp:95 — `GetSkipSize(remainingWidth, frameWidth)`; `frameHeight += skipSize.wholeLines`
- Source/utils/cl2_to_clx.cpp:100 — flush trailing pixels + trailing transparent run
- Source/engine/load_cel.cpp:26 — `LoadCelListOrSheetWithStatus(name, widthOrWidths)` → `name + DEVILUTIONX_CEL_EXT`
- Source/engine/load_cel.hpp:12 — `#define DEVILUTIONX_CEL_EXT ".clx"` (unpacked builds) / `:14` `".cel"`
- Source/engine/load_cl2.cpp:22 — `LoadCl2ListOrSheetWithStatus` → `name + DEVILUTIONX_CL2_EXT` → `Cl2ToClx`
- Source/engine/load_clx.cpp:13 — `LoadOptionalClxListOrSheet` → `OwnedClxSpriteListOrSheet::FromBuffer`
- Source/engine/load_pcx.cpp:35 — `LoadPcxSpriteList(filename, numFramesOrFrameHeight, transparentColor, outPalette)`
- Source/utils/pcx.hpp:8 — `struct PCXHeader` (128 bytes, `PcxHeaderSize` `:30`)
- Source/utils/pcx_to_clx.cpp:47 — `LoadPcxMeta`: `width = Swap16LE(Xmax)-Swap16LE(Xmin)+1`, same for height, `bpp`
- Source/utils/pcx_to_clx.cpp:61 — `PcxToClx(handle, fileSize, numFramesOrFrameHeight, transparentColor, outPalette)`
- Source/utils/pcx_to_clx.cpp:71 — frame splitting: positive arg = frame count, negative = frame height
- Source/utils/pcx_to_clx.cpp:107 — `srcSkip = width % 2`
- Source/utils/pcx_to_clx.cpp:118 — PCX RLE: `byte <= 0xBF` → literal pixel; else `runLength = byte & 0x3F` + colour
- Source/utils/pcx_to_clx.cpp:136 — lines are reversed (`frameBuffer[(frameHeight-(line+1))*width]`)
- Source/utils/pcx_to_clx.cpp:140 — transparent-colour mode builds transparent/solid runs
- Source/utils/pcx_to_clx.cpp:167 — palette block: separator `0x0C` then 256×RGB
- Source/engine/render/clx_render.cpp:73 — `ClxBlitInfo(src) -> (srcEnd, length)` (the "run walk")
- Source/engine/render/clx_render.cpp:94 — `SkipRestOfLineWithOverrun` (the "bin walk")
- Source/engine/render/clx_render.cpp:104 — `SkipLinesForRenderBackwardsWithOverrun`
- Source/engine/render/clx_render.cpp:115 — `DoRenderBackwardsClipY`
- Source/engine/render/clx_render.cpp:148 — `DoRenderBackwardsClipXY`
- Source/engine/render/clx_render.cpp:225 — `DoRenderBackwards` (early-out + `CalculateClipX`)
- Source/engine/render/clx_render.cpp:556 — `ClxDraw` (`BlitDirect`)
- Source/engine/render/clx_render.cpp:561 — `ClxDrawTRN` (`BlitWithMap`)
- Source/engine/render/clx_render.cpp:566 — `ClxDrawWithLightmap`
- Source/engine/render/clx_render.cpp:571 — `ClxDrawBlended`
- Source/engine/render/clx_render.cpp:576 — `ClxDrawBlendedTRN`
- Source/engine/render/clx_render.cpp:581 — `ClxDrawBlendedWithLightmap`
- Source/engine/render/clx_render.cpp:393 — `ClxApplyTrans(sprite, trn)` — in-place index rewrite
- Source/engine/render/clx_render.cpp:441 — `IsPointWithinClx` (mouse hit test on the RLE stream)
- Source/engine/render/clx_render.cpp:487 — `ClxMeasureSolidHorizontalBounds`
- Source/engine/render/clx_render.cpp:250 — `GetOutline<SkipColorIndexZero>` (outline pixel extraction + 1-entry cache
  `:230`)
- Source/engine/render/clx_render.hpp:74 — `RenderClxSprite` = top-left-anchored variant of `ClxDrawTRN`
- Source/engine/surface.hpp:69 — `at(x,y)`, `:88 SetPixel`, `:98 pitch()`

### Formats & algorithms

**Runtime canonical format: CLX.** CEL (retail + many devilutionX assets) and CL2 (player art) and PCX (UI) are all
converted to CLX at load. CLX = `u32 LE numFrames`, `numFrames+1` × `u32 LE` offsets (relative to the list header
start; the last one is the list size), then frames; a *sheet* is a list-of-lists (`ClxSpriteSheet`,
`Source/engine/clx_sprite.hpp:187`) with its own `u32 LE[numLists]` offset table.
Frame = `u16 headerSize(=6)`, `u16 width`, `u16 height`, then CL2-style RLE pixel data.

**CL2/CLX pixel RLE (exact):**
```
control = *src++
if (control < 0x80)                 -> TRANSPARENT run of `control` px        (0x01..0x7F, max 127)
elif (control <= 0xBE)              -> FILL: width = 0xBF - control (1..63); color = *src++
else                                -> PIXELS: width = 256 - control (1..65); width bytes follow
```
Runs **cross line boundaries**. Line advance: after consuming runs until `remainingWidth <= 0`,
```
overrun   = -remainingWidth                    (>=0)
wholeLines = overrun / srcWidth + 1   (if overrun == 0 -> 1 line, xOffset 0)
xOffset    = overrun % srcWidth
```
(`Source/utils/clx_decode.hpp:49`; the "remainingWidth == 0 fast path" comment at `:53` is a perf note, keep semantics).
Frame height is **not stored in CL2** — `Cl2ToClx` computes it by running the decoder and summing `wholeLines`
(`Source/utils/cl2_to_clx.cpp:95-99`). That is why `widthOrWidths` must be exactly right: a wrong width yields a wrong
height and garbage pixels.

**Worked example 1 — CLX, real file** `assets/data/boxleftend.clx` (205 bytes, verified by script):
```
0000: 01 00 00 00        numFrames = 1
0004: 0c 00 00 00        frame0 @ +12  (right after 1 offset + 1 size u32)
0008: cd 00 00 00        frame1 @ +205 == file size  (terminator)
000c: 06 00              headerSize = 6
000e: 06 00              width  = 6
0010: 1b 00              height = 27
0012: pixel data (187 bytes), first runs:
      fa f5 f7 f8 f7 f7 f8   PIXELS 6  colors f5 f7 f8 f7 f7 f8   (7 bytes)
      fe f5 cc               PIXELS 2  colors f5 cc               (3 bytes)
      bb cb                  FILL 4  color cb                     (2 bytes)
      fa f7 cb c1 c2 c2 c3   PIXELS 6                             (7 bytes)
      ...
```
Run widths sum to 162 = 6×27 exactly. Because CLX rows go **bottom-to-top**, the first run
(`f5 f7 f8 f7 f7 f8`) is the *bottom* row of the 6×27 sprite; decoded image rows (bottom→top, printed by the
verification script): `f5f7f8f7f7f8`, `f5cccbcbcbcb`, `f7cbc1c2c2c3`, …, `fafdfbfdfdfc`.
Frame 0 of `assets/data/boxrightend.clx` is 6×27 too, and its second run is
`fd f5 cc` = `fd` → 256-0xfd = **3** pixels `f5 cc ff` (note: 3 px, 4 bytes).

**Worked example 2 — CEL (retail), real file** `/tmp/dgx/data/inv/objcurs.cel`, frame 1, bytes 724..1217:
```
0a 00 00 00 00 00 00 00 00 00    -> u16 at frame start == 0x000A == CelFrameHeaderSize(10)
                                     => header present, skip 10 bytes (Source/utils/cel_to_clx.cpp:82)
eb 01                             -> 0xeb >= 0x80 => TRANSPARENT 256-0xeb = 21 px
fd f5 ec 03                       -> 0xfd => TRANSPARENT 3 px  ... etc
af                                 -> 0x0a? no: 0xaf >= 0x80 => TRANSPARENT 81? -> read from dump: run 'P',10
```
CEL control encoding is different from CL2: **`c >= 0x80` means transparent `256-c` px; `c < 0x80` means `c` literal
pixels follow** (`Source/utils/cel_to_clx.cpp:25-31`). CEL is strictly **one line per iteration of the outer loop,
line length == frameWidth, never crosses lines**, and the frame's height is simply the number of lines consumed
(`Source/utils/cel_to_clx.cpp:104`). Verified: `objcurs.cel` has 179 frames, `u32 LE` count + ascending offsets, and
the widths come from a companion text file `assets/data/inv/objcurs-widths.txt` — one decimal width per line, e.g.
`33, 32, 32, …, 28 (many), 56 (many), 5…` (`ReadWidths(FindAsset("data\\inv\\objcurs-widths.txt"))`,
`Source/cursor.cpp:448`).
Retail dungeon CELs (`l1.cel` frame 1) have **no** 10-byte header (`20 14 12 …`: `0x20 = 32` → 32 literal pixels
follow), and `l1s.cel` (special/arch) has 8 frames of width 64 (`LoadCelWithStatus(special, 64)`,
`Source/diablo.cpp:1365`).

**CEL→CLX transcode** (`Source/utils/cel_to_clx.cpp:35`) re-emits every line as CLX runs:
`AppendClxTransparentRun` + `AppendClxPixelsOrFillRun`, and computes `height` by counting lines. Note it is a strict
per-line conversion (no cross-line merging), so a converted CEL keeps the same visual rows.

**PCX** (`Source/utils/pcx_to_clx.cpp:61`): 128-byte header (`Source/utils/pcx.hpp:8`, fields `Xmin/Ymin/Xmax/Ymax`,
`BitsPerPixel`, `BytesPerLine`), `bpp == 8` asserted (`:67`), rows decoded with classic PCX RLE
(`<=0xBF` literal; `>=0xC0` → `count = byte & 0x3F` repeats of the next byte), row stride padded by `width % 2`
(`:107`), lines **reversed** into a frame buffer (`:136`, because CLX is bottom-up), then re-emitted as CLX runs,
optionally with a `transparentColor` index. Frames are sliced vertically: `numFramesOrFrameHeight > 0` = frame count,
`< 0` = frame height (`:71`). A trailing `0x0C` byte + 768-byte palette is returned via `outPalette` (`:167`).

**`run_length_walk` / `run_bin_walk`:** those helper names do not exist in this tree (verified:
`grep -rn "RunLengthWalk|run_bin_walk|RunBinWalk" Source` → no matches; `git log -S run_bin_walk` → no commits).
The equivalents, which the spec means when referring to run walking:
`ClxBlitInfo` (measure a run, `Source/engine/render/clx_render.cpp:73`),
`SkipRestOfLineWithOverrun` (advance a whole line, `:94`),
`GetSkipSize` (line-wrap arithmetic, `Source/utils/clx_decode.hpp:49`), and
`ClxMeasureSolidHorizontalBounds` (`:487`). Port these four as `clxRunInfo`, `clxSkipLine`, `clxSkipSize`,
`clxMeasureSolidX`.

**Memory model / structs.**
```cpp
class ClxSprite      { const uint8_t* data_; uint32_t pixel_data_size_; }   // 16 bytes, points into the buffer
class ClxSpriteList  { const uint8_t* data_; }                              // header read on the fly
class ClxSpriteSheet { const uint8_t* data_; uint16_t num_lists_; }
class OwnedClxSpriteList { std::unique_ptr<uint8_t[]> data_; }              // owns it
class OwnedClxSpriteListOrSheet { std::unique_ptr<uint8_t[]> data_; uint16_t num_lists_; }
```
(`Source/engine/clx_sprite.hpp:38,95,187,288,418,520`). List-or-sheet is *sniffed*, not flagged:
`GetNumListsFromClxListOrSheetBuffer(data,size)` (`Source/engine/clx_sprite.hpp:404`) reads the first u32 as a frame
count, and if `offsets[numFrames+1] != size`, treats the first u32 as a *byte offset* to the first group ⇒
`numLists = maybeNumFrames/4`. Port this heuristic verbatim — several assets depend on it.

### Ebitengine mapping
- Decoder output type: `type Frame struct { W, H uint16; Index []byte /*len W*H, row-major top-to-bottom*/ }`.
  Decode CLX pixel data into `Index` **once at load** (`ReplacePixels` of an `image.Paletted` with a 256-colour
  palette whose entries encode the *index* → `ebiten.NewImageFromImage` path, or better: pack the index into a RGBA
  texture's R channel and use the palette shader — G10). Decoding to a plain `[]byte` keeps the golden-hash tests
  independent of the GPU.
- Frame count sanity test: decode the first frame of `assets/data/boxleftend.clx` in Go and assert the 27 rows
  printed above (the strings in this document are the golden).
- Store the decoded `Index` alongside the created `*ebiten.Image`? No — after upload, drop it; keep only
  `(atlasRect, w, h)`. Except for `IsPointWithinClx` (`Source/engine/render/clx_render.cpp:441`) which the *cursor
  hit-testing* uses for items: keep a per-item-sprite boolean "solid mask" (1 bit/px) built at decode time instead of
  walking RLE per click. `ClxMeasureSolidHorizontalBounds` results can also be cached at decode.
- Outlines (`ClxDrawOutlineSkipColorZero`, `:590`) are needed for hover highlighting: precompute the outline pixel
  list at decode (C++ caches exactly one sprite: `OutlinePixelsCache`, `:230`) — in Go, cache it in the frame record;
  it is a few hundred bytes/sprite and computed once.
- Allocation discipline: all atlas construction at load (`ReplacePixels` once per atlas page, not per sprite —
  perf tips: "Don't call ReplacePixels too much").

### Parity test
1. **Decoder goldens** — for each format, decode N frames to index arrays and compare against C++ dumps:
   - CLX: `assets/data/boxleftend.clx` (1 frame), `boxrightend.clx`, `dirtybuc.clx` (2 frames, 63×12),
     `health.clx` (264×19), `assets/fonts/30-00.clx` (256 frames, many 0×38 empty), plus ≥ 10 frames of
     `levels/l1data/l1s.cel` after conversion.
   - CEL: `data/inv/objcurs.cel` frame 1 (header-skip path) and `levels/l1data/l1.cel` frame 1 (no header).
   - CL2: `plrgfx/warrior/hu/huas.cl2` (8 direction lists).
   - PCX: `ui_art/heros.pcx` sliced by `-PortraitHeight` (`Source/DiabloUI/diabloui.cpp:629`).
   Produce the C++ side with a tiny gtest that writes PNGs of decoded frames (`test/fixtures/` already hosts PNGs for
   `text_render_integration_test`), or dump raw index bytes. Command sketch:
   `cmake -S . -B build -DBUILD_TESTING=ON -DENABLE_TESTS=ON && cmake --build build --target devilutionx-tests && ./build/test/devilutionx-tests --gtest_filter='*Clx*'`
2. **Structural test** — implement the group/list/sheet sniff (`:404`) against a synthetic corpus: single list, sheet,
   and a file whose last offset equals size (must be read as a list).
3. **Fuzz** — `go test -fuzz=FuzzClxDecode` on truncated/oversized control bytes; must error, never panic
   (mirrors dg-sub-0003).

### Open questions
- Where do we source binary `.cel/.cl2/.pcx` fixtures for CI? The repo ships only `.clx` in `assets/` (verified:
  `find assets -iname '*.cel' -o -iname '*.cl2'` → 0 results) and retail MPQs are not in the repo. Decision needed:
  commit a small derived fixture set (license note: assets are retail; the repo already ships `assets/*.clx` under
  the project's own provenance rules) or run decoder goldens only in the "has MPQ" CI job.
- Should the Go port keep the CL2→CLX conversion (re-encode) or decode CL2 directly to GPU? Keeping the re-encode
  makes every other pass (TRN application, `ClxApplyTrans`, frame index arithmetic) byte-identical to C++.

---

## G4. PALETTES & TRN (translation tables)

### Sources (file:line)
- Source/engine/palette.h:14 — `extern std::array<SDL_Color, 256> logical_palette;`
- Source/engine/palette.h:15 — `extern std::array<SDL_Color, 256> system_palette;`
- Source/engine/palette.cpp:52 — `CycleColors(from,to)`
- Source/engine/palette.cpp:71 — `CycleColorsReverse(from,to)`
- Source/engine/palette.cpp:175 — `LoadPalette(path)` (768-byte RGB file)
- Source/engine/palette.cpp:188 — `LoadPaletteAndInitBlending(path)` (calls `GenerateBlendedLookupTable` with
  per-level skip ranges: caves/crypt `skipFrom=1,skipTo=31`, nest `1..15`, else full)
- Source/engine/palette.cpp:214 — `LoadRndLvlPal(l)` — the palette file name selection rules
- Source/engine/palette.cpp:118 — `ApplyGlobalBrightness` (tone-map LUT, 256 entries)
- Source/engine/palette.cpp:140 — `ApplyFadeLevel(fadeval,dst,src)` = `(fadeval*c)/256`
- Source/engine/trn.cpp:14 — `GetInfravisionTRN()` → `InfravisionTable`
- Source/engine/trn.cpp:19 — `GetStoneTRN()` → `StoneTable`
- Source/engine/trn.cpp:24 — `GetPauseTRN()` → `PauseTable`
- Source/engine/trn.cpp:29 — `GetClassTRN(player)` → `plrgfx\<class>.trn`
- Source/engine/trn.cpp:47 — `GetPlayerGraphicTRN(name)` → `<anim>.trn` next to the CL2
- Source/lighting.cpp:37 — `std::array<uint8_t, 256> InfravisionTable; StoneTable; PauseTable;`
- Source/lighting.cpp:214 — `LoadTrns()` — `plrgfx\infra.trn`, `plrgfx\stone.trn`, `gendata\pause.trn`
- Source/monster.cpp:173 — `InitMonsterTRN(CMonster&)` → `monsters\<trnFile>.trn`, then `std::replace(255→0)`
- Source/monster.cpp:186 — applies it with `ClxApplyTrans` over every animation (list or sheet)
- Source/monster.cpp:3318 — unique monster TRN: `monsters\monsters\<mTrnName>.trn` → `monster.uniqueMonsterTRN`
- Source/monster.cpp:3572 — `if (!monsterData.trnFile.empty()) InitMonsterTRN(...)`
- Source/player.cpp:2167 — `GetPlayerGraphicTRN(pszName)` → `ClxApplyTrans` on the loaded CL2 sheet
- Source/player.cpp:2172 — `GetClassTRN(player)` → `ClxApplyTrans` afterwards (order: graphic TRN, then class TRN)
- Source/engine/light_tables.cpp:19 — `MakeLightTable()`
- Source/engine/light_tables.cpp:22-46 — the light-table generation loop (steps 16×8 then 8×4 then 16×6)
- Source/engine/light_tables.cpp:48 — `LightTables[15] = {}` (fully dark)
- Source/engine/light_tables.cpp:53 — hell: blood-wall remap of indices 1..16/31..16 per level
- Source/engine/light_tables.cpp:68 — nest/crypt: lava kept fully bright (`iota` over first 16)
- Source/engine/light_tables.hpp:18 — `LightTables[16][256]`, `FullyLitLightTable`, `FullyDarkLightTable`
- Source/engine/render/clx_render.cpp:393 — `ClxApplyTrans(ClxSprite, trn)` (rewrites indices **in place**)
- Source/engine/render/clx_render.cpp:415 — `ClxApplyTrans(ClxSpriteList)`, `:421` `(ClxSpriteSheet)`
- Source/engine/render/blit_impl.hpp:45 — `BlitFillWithMap`: `memset(dst, colorMap[color], length)`
- Source/engine/render/blit_impl.hpp:50 — `BlitPixelsWithMap`: `dst[i] = colorMap[src[i]]`
- Source/engine/render/clx_render.cpp:561 — `ClxDrawTRN`
- Source/engine/render/scrollrt.cpp:249 — `ClxDrawLight` (`LightTables[i]` used *as* the TRN)
- Source/engine/render/scrollrt.cpp:378 — `DrawMonster`: TRN choice (infravision/stone/unique)
- Source/engine/render/scrollrt.cpp:781 — `DrawDungeon`: corpse `uniqueMonsterTRN` branch (`:802`)
- Source/engine/render/text_render.cpp:88 — `ColorTranslations[19]` TRN paths
- Source/engine/render/text_render.cpp:269 — `DrawFont` → `RenderClxSpriteWithTRN`
- Source/engine/palette.cpp:406 — `SetLogicalPaletteColor(i,color)` (mods can rewrite a single colour)
- Source/utils/palette_blending.hpp:16 — `extern uint8_t paletteTransparencyLookup[256][256];`
- Source/utils/palette_blending.cpp:46 — `GenerateBlendedLookupTable` (nearest-neighbour in a KD-tree of the palette)
- Source/utils/palette_blending.cpp:30 — `BlendColors(a,b) = ((a.r+b.r)/2, (a.g+b.g)/2, (a.b+b.b)/2)`

### Formats & algorithms

**The model:** 256-entry palette of RGB8. Sprites and tiles store *indices*. All colour effects are index→index
lookups. There are exactly 4 families of tables:

1. **Light tables** `LightTables[16][256]` (`Source/engine/light_tables.cpp:19`). Generation, transcribed:
```
shade = 0
for each of 16 tables:
    colorIndex = 0
    for steps in {16,16,16,16,16,16,16,16, 8,8,8,8, 16,16,16,16,16,16}:   # sums to 256
        shading = shade * steps / 16                       # integer division
        shadeStart = colorIndex; shadeEnd = shadeStart + steps - 1
        for step in 0..steps-1:
            if colorIndex == 0: table[0] = 0; colorIndex++; continue     # black stays black
            color = shadeStart + step + shading
            if color > shadeEnd or color == 255: color = 0               # wraps to black
            table[colorIndex++] = color
    shade++
LightTables[15] = all zeros     # fully dark
```
   (Note the loop over 18 `steps` entries covers 8·16 + 4·8 + 6·16 = 256 indices; `colorIndex` therefore walks the
   whole 256-byte table.) Meaning: index ranges are the palette's 16-colour *ramps*; a light level shifts a colour
   `shade*steps/16` positions down its own ramp and colours falling outside their ramp become black.
   Level overrides: hell (`:53`) rewrites indices 1..16 and 31..16 (mirrored) per level with
   `color = ((15)<<4)/shades*(shades-i)/Range*(j+1); color = 1 + (color>>4)`;
   nest/crypt (`:68`) set `iota(0..15)` for the first 16 of *every* level and `LightTables[15][0]=0`,
   `LightTables[15][1..15]=1`, and **clear `FullyDarkLightTable`** so no tile is pure black.
   `FullyLitLightTable` is cleared in hell (`:66`) because even fully-lit tiles then need a mapping.
2. **TRN files** — plain 256-byte arrays, loaded with `LoadFileInMem<uint8_t>(path, array)`:
   - `plrgfx/infra.trn` (infravision / unlit / unusable-items red), `plrgfx/stone.trn` (petrified),
     `gendata/pause.trn` (debug path overlay). Verified 256 bytes each (`/tmp/dgx/plrgfx/infra.trn`).
   - `plrgfx/<class>.trn` (class tinting) — `warrior.trn`, `rogue.trn`, `sorcerer.trn` (+ `monk`, `barbarian` in
     Hellfire). Applied *after* the per-graphic TRN (`Source/player.cpp:2167-2176`), both via `ClxApplyTrans`.
   - `plrgfx/<cls>/<xyz>/<xyz><anim>.trn` per-graphic TRNs, e.g. `plrgfx/warrior/whu/whufm.trn`, `whulm.trn`,
     `whuqm.trn` (fire/lightning/magic spell animations recoloured) — 256 bytes each (verified from
     `devilutionx.mpq` listing).
   - `monsters/<trnFile>.trn` per monster *type* (e.g. `zombie/bluered`, `zombie/grey` from
     `assets/txtdata/monsters/monstdat.tsv` column `trnFile`), applied at type load time, with the quirk
     `std::replace(255 → 0)` first (`Source/monster.cpp:177`): **an entry of 255 in the file means "map to black"**.
   - `monsters/monsters/<mTrnName>.trn` for uniques, kept per monster instance and used at draw
     (`Source/monster.cpp:3318`, `Source/engine/render/scrollrt.cpp:802`, `:378`).
   - `fonts/*.trn` for text colours: 19 entries in `ColorTranslations` (`goldui, grayui, golduis, grayuis, —, —,
     yellow, —, black, white, whitegold, red, blue, orange, buttonface, buttonpushed, gamedialogwhite,
     gamedialogyellow, gamedialogred`) — verified present in `devilutionx.mpq` (20 `.trn` entries listed).
3. **`paletteTransparencyLookup[256][256]`** — see G8.
4. **`system_palette` vs `logical_palette`** — logical is the file content; system = tone-mapped (brightness) version
   actually uploaded to the GPU/SDL palette. Fades multiply `system_palette`.

**Byte-level draw-time translation.** `dstIndex = trn[srcIndex]` — *only for opaque pixels*; transparent runs write
nothing, so background indices survive. Fill runs translate the repeated colour once
(`BlitFillWithMap`: `memset(dst, colorMap[color], n)`). Because `ClxApplyTrans` exists (`Source/engine/render/clx_render.cpp:393`),
some TRNs are baked into the sprite data **at load** (monster type TRN, player class/graphic TRN) and some are applied
**per draw** (light tables, infravision, stone, unique TRN for corpses/missiles, font colour TRN).

**TRN census (what must exist in Go):**

| family | when applied | count | source |
|---|---|---|---|
| light tables | per draw (per tile's `dLight`) | 16 | `Source/engine/light_tables.cpp:19` |
| monster type TRN | baked at type load | one per monster with `trnFile` set (monstdat.tsv) | `Source/monster.cpp:173` |
| unique monster TRN | per draw | one per unique (`UniqueMonstersData`) | `Source/monster.cpp:3318` |
| player class TRN | baked at load | ≤ 6 (classes) | `Source/engine/trn.cpp:29` |
| player graphic TRN | baked at load | 3 per spell-casting class (fm/lm/qm) | `Source/engine/trn.cpp:47` |
| infra/stone/pause | per draw | 3 | `Source/lighting.cpp:214` |
| font colour TRN | per draw | 16 (19 slots, 3 null) | `Source/engine/render/text_render.cpp:88` |
| transparency LUT | per pixel | 256×256 | `Source/utils/palette_blending.cpp:46` |

**Combos count estimate for the "pre-bake RGBA per (sprite, palette)" option:**
- light levels 16 × player frames: 4 classes × (8 dirs × ~10 anims × ~10 frames) ≈ 4·800 = 3200 frames →
  3200×16 = 51 k blits of the same sprites — memory blow-up (player frames are ~90×110 avg ≈ 10 KB index each).
- dungeon micros: `l1.cel` 1119 frames × 16 = 17 904; all 7 tilesets ≈ 6–7 k frames total → ≈ 110 k variants.
- Conclusion: **do not pre-bake per light level.** Pre-bake only where the set is small and stable: monster-type TRN
  (bake into the decoded indices, exactly like C++), player class/graphic TRN (same), font colour
  (16 palettes × glyph atlas). Use palette-swap textures for light levels and the rare per-draw TRNs.

### Ebitengine mapping — decision
**Recommended: (b) paletted textures + one Kage "index→palette" fragment shader**, with 16 palette textures.

Rationale:
1. Byte-exactness is *provable*: the shader does `idx = texSrc.r*255` → `rgb = texPalette[vec2((idx+0.5)/256, 0.5)]`.
   Because the palette texture is filled with the *exact* RGB bytes from `system_palette` (uploaded as
   `image.NewPaletted`/`ReplacePixels`, no filtering), the result is bit-identical to SDL's INDEX8→RGB blit. No float
   colour maths enter the pipeline at all, so the "float RGBA rounding" hazard disappears (Ebitengine stores RGBA8 =
   8 bits/channel, and `FilterNearest` sampling is a pure texel fetch).
2. Light: bind palette variant `P[level]` where `P[level][i] = palette[LightTables[level][i]]` — computed once on
   palette change (16 × 256 × 3 bytes = 12 KB, recomputed on `LoadPalette`, `UpdateSystemPalette`, cycling).
   A tile run with constant `dLight` binds the same texture → keeps the batch
   (https://ebitengine.org/en/documents/performancetips.html — batching requires same target/blend/filter; source
   images in the same internal atlas keep it to one command).
   Cycling changes only 31 palette entries: re-upload 16 textures of 256×1 (cheap, `ReplacePixels` on tiny images at
   most 4×/s — the cycle runs on ticks, `Source/diablo.cpp:3521`).
3. Per-draw TRNs (infravision, stone, unique) = additional palette variants of the *current* level, i.e.
   `P_trn[level][i] = palette[LightTables[level][trn[i]]]`. Generate lazily and cache by `(trn, level)`; realistic
   working set: 3 TRNs × 16 + per-unique TRN × 16 ≈ a few hundred 768-byte tables ≈ < 1 MB GPU.
   **This is exactly the C++ composition order** (`ClxDrawTRN(out,pos,sprite,LightTables[i].data())`,
   `Source/engine/render/scrollrt.cpp:249`) — note light table applied *after* the sprite's baked TRN, never before.
4. Fades: multiply all 16 palette variants (or multiply the base palette once and re-derive; cheaper) — matches
   `ApplyFadeLevel` (`Source/engine/palette.cpp:140`).
5. Alternative (a) "pre-baked RGBA per (sprite, palette)" would need 8× the atlas memory for players/monsters and ~110 k
   tile variants (estimate above) and would still not solve `paletteTransparencyLookup` blending (needs the *destination*
   colour, not just the source). Rejected.
6. Regarding the upstream examples: **there is no `examples/palette` in the ebitengine examples tree** — verified by
   listing https://github.com/hajimehoshi/ebiten/tree/v2.9.9/examples (86 entries; closest are `filter`, `hsv`, `hue`,
   `mandelbrot`, `squiral`, `shader`, `blend`, `masking`, `minify`). The palette-lookup technique must be written as a
   custom Kage shader (docs: https://ebitengine.org/en/documents/shader.html). Also verified
   `ebiten.SetCursorShape` exists with 10 shape constants (https://github.com/hajimehoshi/ebiten/blob/main/cursor.go,
   example https://github.com/hajimehoshi/ebiten/blob/v2.9.2/examples/cursor/main.go).

```kage (pixel unit, no ColorMatrix):
//kage:unit pixels
package main
var PaletteOrigin vec2      // texel of palette row for this batch (one row per variant)
func Fragment(dstPos vec4, srcTexEl vec2, color vec4) vec4 {
        idx := imageSrc0At(srcTexEl).r * 255.0            // index stored in R channel
        if idx < 0.5 { return vec4(0,0,0,0) }             // index 0 == transparent (sprites' "hole")
        c := imageSample0? ... -> use a dedicated palette image bound as imageSrc1
        return vec4(imageSrc1At(vec2((idx+0.5)/256.0, PaletteOrigin.y)).rgb, 1.0)
}
```
(Exact Go/Kage spelling to be finalised in the task; the *semantics* are the contract: R-channel index, texel fetch
from a 256-column palette row, index 0 → alpha 0. Note index 0 is only "transparent" for **sprites** (their RLE
skips holes so a hole never appears as a written pixel); dungeon micros have no alpha — `Square`/triangle frames write
every pixel of their shape, index 0 included, which is legitimate black. **Therefore the shader must not key on
index 0 for tiles** — use two shaders or a uniform-free variant: the tile shader writes alpha 1 always, the sprite
shader keys on 0. This is a real parity trap; flagged in G11 too.)

### Parity test
- `TestMakeLightTable`: port `MakeLightTable` and compare all 16×256 bytes against a C++ dump for
  `leveltype ∈ {cathedral, catacombs, caves, hell, nest, crypt}` (the hell/nest/crypt branches change the tables!).
- `TestTrnApplication`: load `plrgfx/infra.trn` (256 bytes, present in both retail and `devilutionx.mpq`) and
  `gendata/pause.trn` from `assets/`; assert `frame.Index` after `ClxApplyTrans` equals a C++-dumped frame.
- `TestPaletteRoundtrip`: for a fixture frame, produce the Go `*ebiten.Image` pixels via a one-off
  `ReadPixels` in the test and compare to `system_palette[LightTables[l][idx]]` computed in Go. This is where the
  float issue is checked: Ebitengine's `ReadPixels` on an unmodified texture returns the texel bytes; assert
  `r,g,b` exactly equal the palette bytes and `a == 255` (or 0). **Rule: never let a ColorMatrix touch a
  pixel-parity-checked draw.**
- `TestFadeLevelBytes`: `ApplyFadeLevel` parity (integer `(fadeval*c)/256`).

### Open questions
- Palette texture layout: one 256×1 image per (light level × trn × fade) variant is many small images (each becomes
  its own atlas region; the atlas is ~4096×4096 and these are tiny, fine) — or one 256×N sheet + `PaletteOrigin`
  uniform (a per-draw uniform **does not** break Ebitengine batching since uniforms are folded into vertices? Not
  guaranteed — verify empirically with `-tags=ebitenginedebug` draw-command counting, see perf tips). Recommend:
  start with separate images, measure.
- Do we support `SetLogicalPaletteColor` (mods rewriting a palette entry, `Source/engine/palette.cpp:406`)? Only if
  the Lua/mod surface is ported.

---

## G5. LIGHTING

### Sources (file:line)
- Source/engine/light_tables.cpp:78 — `LightFalloffs[16][128]` generation (the falloff formula)
- Source/engine/light_tables.cpp:98 — linear branch
- Source/engine/light_tables.cpp:89 — nest/crypt quadratic branch
- Source/engine/light_tables.cpp:107 — `LightConeInterpolations[8][8][16][16]` generation
- Source/engine/light_tables.hpp:20 — `extern uint8_t LightFalloffs[NumLightRadiuses][128];`
- Source/engine/light_tables.hpp:21 — `extern uint8_t LightConeInterpolations[8][8][16][16];`
- Source/engine/lighting_defs.hpp:8 — `#define MAXLIGHTS 32`
- Source/engine/lighting_defs.hpp:11 — `constexpr char LightsMax = 15;`
- Source/engine/lighting_defs.hpp:14 — `LightTableSize = 256`, `:17 NumLightingLevels = 16`
- Source/lighting.cpp:33 — `Light VisionList[MAXVISION]`, `:34 Lights[MAXLIGHTS]`, `:35 ActiveLights`, `:36 ActiveLightCount`
- Source/lighting.cpp:48 — `RotateRadius` (the 4-quadrant walk)
- Source/lighting.cpp:70 — `SetLight` (writes `dLight` or `dPreLight` while loading)
- Source/lighting.cpp:83 — `TileAllowsLight` = `!BlockLight`
- Source/lighting.cpp:90 — `DoVisionFlags` (`| Lit`, `| Visible`, `Explored`)
- Source/lighting.cpp:104 — `DoUnLight` (radius+2 restore to `dPreLight`)
- Source/lighting.cpp:117 — `DoLighting(position, radius, offset)` — the core loop
- Source/lighting.cpp:161-171 — the inner loop using `LightConeInterpolations` and `LightFalloffs`
- Source/lighting.cpp:194 — `DoVision` (ray-based, marks `TransList` too)
- Source/lighting.cpp:239 — `InitLighting`
- Source/lighting.cpp:251 — `AddLight` (MAXLIGHTS pool, sets `UpdateLighting`)
- Source/lighting.cpp:269 — `AddUnLight`
- Source/lighting.cpp:290/308/326/347 — `ChangeLightRadius/XY/Offset/ChangeLight`
- Source/engine/render/scrollrt.cpp:249 — `ClxDrawLight`
- Source/engine/render/scrollrt.cpp:786 — `const int lightTableIndex = dLight[x][y]` in `DrawDungeon`
- Source/engine/render/scrollrt.cpp:681 — same in `DrawFloorTile`
- Source/engine/render/scrollrt.cpp:489 — player light bias `max(lightTableIndex-5, 0)`
- Source/engine/render/scrollrt.cpp:397 — `!IsTileLit → GetInfravisionTRN` for monsters
- Source/engine/render/scrollrt.cpp:401 — `_pInfraFlag && lightTableIndex > 8` → infravision
- Source/engine/render/scrollrt.cpp:481 — player: unlit or (`infra && level>8`) → infravision TRN
- Source/engine/render/scrollrt.cpp:806 — corpses only drawn if `lightTableIndex < LightsMax`
- Source/engine/render/scrollrt.cpp:823 — objects only found if `lightTableIndex < LightsMax`
- Source/engine/render/light_render.cpp:24 — `RenderFullTile` (diamond fill into the per-pixel lightmap)
- Source/engine/render/light_render.cpp:44 — `RenderTriangle` (half-space rasteriser, 28.4 fixed point)
- Source/engine/render/light_render.cpp:150 — `InterpTable[n][d] = ((n<<4)+8)/d`
- Source/engine/render/light_render.cpp:162 — `RenderCell` (marching squares over 4 corner light levels)
- Source/engine/render/light_render.cpp:432 — `BuildLightmap`
- Source/engine/render/light_render.cpp:513 — `Lightmap::build`
- Source/engine/render/light_render.cpp:527 — `Lightmap::bleedUp`
- Source/engine/render/light_render.hpp:30 — `Lightmap::adjustColor(color, lightLevel) = lightTables[lightLevel][color]`
- Source/engine/render/light_render.hpp:35 — `getLightingAt(outLoc)` (maps a backbuffer address to a lightmap texel)
- Source/engine/render/blit_impl.hpp:60 — `BlitFillWithLightmap` / `:70 BlitPixelsWithLightmap`
- Source/diablo.cpp:3521 — `palette_update_caves()` call site (tick-driven)
- Source/diablo.cpp:3526 — `palette_update_hive()`
- Source/diablo.cpp:3528 — `palette_update_crypt()`
- test/light_render_benchmark.cpp:23 — the 16×256 light tables used by the benchmark
- test/fixtures/light_render_benchmark/dLight.dmp — committed fixture: a `dLight` map dump (reusable!)

### Formats & algorithms

**A. Per-tile light level `dLight[x][y] ∈ [0,15]`** (0 = fully lit, 15 = black). `MakeLightTable` derives the
16 translation tables (G4). Rendering never "darkens": it looks up `LightTables[dLight]` and blits indices through it.

**B. Light falloff tables.** For radius index `r ∈ [0,16)` (r = the light's `radius` field, 0..15):
```
maxDistance = (r + 1) * 8
for distance in 0..127:
    if distance > maxDistance:  LightFalloffs[r][distance] = 15
    else:
        factor = distance / maxDistance                      # float
        if leveltype in {NEST, CRYPT}:
            brightness = r * 1.25
            scaled = factor*factor*brightness + (15 - brightness)
            scaled = max(0.0, scaled)                        # over-exposure clamps to fully bright
        else:
            scaled = factor * 15.0                           # linear
        scaled += 0.5                                        # round half-up
        LightFalloffs[r][distance] = uint8(scaled)
```
(`Source/engine/light_tables.cpp:78-105`. Verified the two branches and the `+0.5F` rounding note in the source
comments "Leaner falloff" / "quardratic falloff with over exposure".)

**C. Distance metric — `LightConeInterpolations[8][8][16][16]`** (`Source/engine/light_tables.cpp:107`):
```
for offsetY in 0..7: for offsetX in 0..7: for y in 0..15: for x in 0..15:
    a = 8*x - offsetX;  b = 8*y - offsetY
    table[offsetX][offsetY][x][y] = uint8( sqrtf(a*a + b*b) )      # float sqrt, truncating cast
```
i.e. distance in 1/8-tile units from a sub-tile offset, pre-quantised to an integer index into `LightFalloffs`.

**D. `DoLighting(position, radius, offset)`** (`Source/lighting.cpp:117`), the exact algorithm:
```
if offset.x < 0 { offset.x += 8; position.x -= 1 }
if offset.y < 0 { offset.y += 8; position.y -= 1 }
dist = offset
minX/maxX/minY/maxY = 15 clamped to map bounds (position±15)
if leveltype in {NEST, CRYPT}: if GetLight(pos) > LightFalloffs[radius][0]: SetLight(pos, LightFalloffs[radius][0])
else:                          SetLight(position, 0)
for i in 0..3:                                   # four quadrant sweeps, rotated
    yBound = (i>0 && i<3) ? maxY : minY
    xBound = (i<2) ? maxX : minX
    for y in 0..yBound-1:
        for x in 1..xBound-1:                     # NOTE x starts at 1, y at 0
            linearDistance = LightConeInterpolations[offset.x][offset.y][x + block.x][y + block.y]
            if linearDistance >= 128: continue
            temp = position + Displacement{x,y}.Rotate(-i)
            v = LightFalloffs[radius][linearDistance]
            if !InDungeonBounds(temp): continue
            if v < GetLight(temp): SetLight(temp, v)      # min-merge, never brightens
    RotateRadius(offset, dist, light, block)     # rotate the 1/8-offset frame; wraps negative coords, sets block flags
```
`RotateRadius` (`:48`) is `dist = {7-dist.y, dist.x}`, same for `light`, `offset = dist - light`, then wrap +8 and set
`block = {1,0}/{0,1}` for negatives. `DoUnLight` (`:104`) restores a `(radius+2)` box from `dPreLight`.
**Lights are 1/8-tile positioned** (`ChangeLightOffset`, `:326`) so a torch on a wall can bleed half-tiles.

**E. Vision (which tiles are "lit" at all, a boolean separate from `dLight`).** `DoVision` (`:194`) raycasts from the
player with radius = `_pLightRad` (a `VisionList` per player, `MAXVISION = 4`), sets `DungeonFlag::Lit` + `Visible`
(`:90`), marks `Explored`/`AutomapView`, and *sets `TransList[dTransVal[t]] = true`* along transparent corridors
(`:200`). Effects used by the renderer:
- `DrawMonsterHelper`: `if (!IsTileLit(tile) && !(MyPlayer->_pInfraFlag && IsFloor(tile))) return;`
  (`Source/engine/render/scrollrt.cpp:752`) — monster not drawn at all.
- `DrawMonster`: `if (!IsTileLit(tilePosition)) → ClxDrawTRN(infravision)` (`:397`) — drawn red instead
  (this branch is only reached for `MFLAG_...`/other-level cases; keep both).
- `DrawPlayer`: hidden if unlit and no infravision and not arena/town (`:471`); red (`GetInfravisionTRN`) if unlit or
  (`infra && level>8`) (`:481`); else `lightTableIndex = max(level-5, 0)` (`:487`) — **players are always drawn
  brighter than their tile**.
- `DrawDungeon`: corpses (`:806`) and objects (`:823`) skipped entirely at `lightTableIndex == LightsMax`.

**F. Entity halos.** A light is an `(position, radius)` pair only — no halo sprite. `AddLight`/`ChangeLightXY`
(`:251`, `:308`) with `MAXLIGHTS = 32` active lights pooled via `ActiveLights` + `ActiveLightCount`;
`UpdateLighting` latches and the whole `dLight` is recomputed from `dPreLight` when dirty. Items that "shine"
(rings, stags) are `lightTableIndex`-driven too — the item light radius comes from the item table, applied in
`items.cpp`, not here. **`dPreLight` is the static base from the level's own light map (set while `LoadingMapObjects`,
`Source/lighting.cpp:70`).**

**G. Palette fades** — G1 §7 (`ApplyFadeLevel` = integer `(fadeval*c)/256` per channel). Black/green/red fades in this
fork are all achieved by *rewriting the palette file/ranges*: there is no separate "green fade"; the town/crypt/etc.
`.pal` files define the tint and `PaletteFadeIn/Out` scale whatever the current palette is. Item sparkle: **no
sparkle overlay sprite exists in this fork** (verified: `grep -rn "sparkle" Source` → no matches); "shimmering
items" are rendered (a) with the `infra.trn` red tint if unusable (`Source/cursor.cpp:560` `DrawItem`), and (b) with a
1-px outline of a colour from `GetOutlineColor` (`Source/items.cpp:2403`) when hovered, plus (c) the QoL name label.

**H. Per-pixel lighting (optional mode).** `Lightmap::build` (`Source/engine/render/light_render.cpp:513`) rasterises
`dLight` into an 8-bit *light level per pixel* buffer sized `viewportWidth × (viewportHeight + 32*(microTileLen/2+1))`,
by marching squares over each tile *quad* (corner values `q0..q3` = 4 neighbouring tiles' `dLight`) and filling
contours from `min(maxLight, 14)` down to `minLight` (`:459`). Triangle interpolation factors use
`InterpTable[n][d] = (n*16+8)/d` (`:150`) — integer, quantised to 1/16. `bleedUp` (`:527`) copies the bottom row of
the base diamond up into a scratch `TILE_WIDTH*TILE_HEIGHT` buffer so tall wall micros receive light from the floor
they stand on. Then `BlitPixelsWithLightmap` does `dst[i] = lightTables[light[i]][src[i]]`
(`Source/engine/render/blit_impl.hpp:70`). **Recommendation: not ported for the first renderer milestone** (matches
Volume I §7 "optional flag, default table-based"); the byte-op is stated here so the decision is documented.

### Ebitengine mapping — decision + budget
- **Decision: per-tile light via palette-variant textures (16 palette rows), no ColorScale, no light shader.**
  `ColorScale` multiply is *rejected* for parity: it is a float `ColorM` on premultiplied RGBA and cannot reproduce
  `LightTables[i][j]` (a non-monotone ramp remap that sends colours to black when they leave their ramp).
  A Kage light shader is *unnecessary*: the palette-variant approach already covers both tiles and sprites with one
  mechanism (G4). `Lightmap`/per-pixel mode, if ever needed, would be a 640×~960 `R8` texture +
  `DrawRectShader`; out of scope, recorded as such.
- Per-tile table lookup: `dLight` becomes a `[MAXDUNX][MAXDUNY]uint8` (or `[112][112]byte`) grid — already required
  by dg-epic-04. At draw, `level := dLight[tile]` picks the palette variant for the batch: group the row-walk so
  consecutive micros with the same level share one batch (they mostly do — the walk is row-major over tiles).
- Vision (`Lit`/`Visible`) is logic-side; the renderer just reads `dFlags` (`Source/levels/dun_tile_data.hpp:114`).
- Frame budget at 640×480:
  - micros per frame: worst case ≈ `rows(≈30) × cols(≈11) = 330` tiles × up to 8 micro pairs, but only visible
    floor/wall micros draw → in practice 330 tiles × (2 floor micros + a few wall micros) ≈ 1.5–3 k `DrawImage`s.
  - entities ≈ up to 200 monsters + 4 players + items/objects/missiles on screen ⇒ < 500 sprite draws.
  - each `DrawImage` = 1 quad = 2 triangles ⇒ ≤ ~7 k triangles/frame: trivial for any GPU, and Ebitengine merges
    same-atlas/same-blend/same-filter runs into a handful of draw commands.
  - palette variant swaps do *not* change blend/filter, so they are cheap; if measurement shows they split batches,
    switch to the "one 256×16 palette sheet + PaletteOrigin uniform" variant and count commands with
    `go build -tags=ebitenginedebug` (perf tips page documents exactly this).
- Fades: multiply the base palette on CPU per step (256×3 bytes × 16 variants = 12 KB, at most once per frame) and
  `ReplacePixels` the palette textures; never touch the sprites.

### Parity test
- `TestLightFalloffs`: port the formula; compare all 16×128 bytes with a C++ dump, both branches
  (`leveltype` = caves vs crypt). Watch the float: Go `float32` arithmetic must mirror `static_cast<float>` —
  compute with `float32` explicitly, not `float64`.
- `TestLightConeInterpolations`: all 8×8×16×16 = 16 384 bytes; `uint8(math.Sqrt(float64(a*a+b*b)))` vs C++
  `sqrt((a*a)+(b*b))` where the args are `int` and the result `double` — check truncation matches on the boundary
  values (e.g. `a²+b²` a perfect square ±1). Add a deliberate boundary case list.
- `TestDoLighting`: use the committed fixture `test/fixtures/light_render_benchmark/dLight.dmp`
  (`ls test/fixtures/light_render_benchmark` → `dLight.dmp`) as input state; assert byte-equal `dLight` after a
  scripted sequence of `AddLight/ChangeLightXY/DoUnLight`.
- `TestMakeLightTableLevelVariants`: as G4.
- `TestNoUnlitCorpses`: assert a tile with `dLight==15` draws no corpse/object (order-instrumented test).

### Open questions
- Exact `sqrt` semantics: C++ `sqrt((a*a)+(b*b))` promotes to `double`; Go `math.Sqrt` is also `float64` — parity
  expected, but the `uint8()` truncation of values like `sqrt(63)=7.937…` must be checked exhaustively (cheap: 16 384
  values, just compare in the test).
- Item "sparkle": none in this fork — confirm the product requirement doesn't want vanilla's shimmering gold
  (which is a *palette cycle*, i.e. `palette_update_*` + specific indices, already covered).
