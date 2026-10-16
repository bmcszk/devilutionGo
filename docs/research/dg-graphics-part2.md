---

## G6. ENTITY & UI SPRITE PLUMBING

### Sources (file:line)
- Source/engine/animationinfo.h:114 — `constexpr static uint8_t baseValueFraction = 128;`
- Source/engine/animationinfo.h:64 — `ClxSprite currentSprite() const`
- Source/engine/animationinfo.h:83 — `uint8_t getAnimationProgress() const`
- Source/engine/animationinfo.cpp:18 — `getFrameToUseForRendering()` (frame-skipping distribution)
- Source/engine/animationinfo.cpp:60 — `getAnimationProgress()` (fraction math, both branches)
- Source/engine/animationinfo.cpp:85 — `setNewAnimation(sprites, numberOfFrames, ticksPerFrame, flags, numSkippedFrames, distributeFramesBeforeFrame, previewShownGameTickFragments)`
- Source/engine/animationinfo.cpp:104 — the distribution setup (`relevantFramesForDistributing_`, `tickModifier_`)
- Source/monster.h:156 — `struct AnimStruct { OptionalClxSpriteListOrSheet sprites; int8_t frames; int8_t rate; }`
- Source/monster.h:162 — `spritesForDirection(dir) = sprites->isSheet() ? sheet()[dir] : list()`
- Source/monster.h:305 — `animInfo.changeAnimationData(animationData.spritesForDirection(dir), frames, rate)`
- Source/monster.cpp:150 — `constexpr char Animletter[7] = "nwahds"` (animation → letter)
- Source/monster.cpp:152 — `GetNumAnims = hasSpecial ? 6 : 5`
- Source/monster.cpp:656 — `setNewAnimation(animData.spritesForDirection(md), frames, rate, flags, …)`
- Source/monster.cpp:3217 — file-name generator `"monsters\" + spritePath + {n,w,a,h,d,s}[i] + ".cl2"`
- Source/monster.cpp:3549 — `InitMonsterGFX` (list-or-sheet slicing via `GetNumListsFromClxListOrSheetBuffer`)
- Source/player.h:780 — `currentSprite()` = `previewCelSprite ? *previewCelSprite : AnimInfo.currentSprite()`
- Source/player.h:785 — `getRenderingOffset(sprite)`
- Source/player.cpp:1447 — `GetPlayerGraphicsPath` → `plrgfx\<class>\<xyz>\<xyz><anim>` (`:1530` is the BufCopy)
- Source/player.cpp:2103 — `LoadPlrGFX(player, graphic)` (the anim-letter table `as/aw/at/ht/lm/fm/qm/dt/bl`)
- Source/player.cpp:2128 — `szCel = "wl"` for town walk
- Source/player.cpp:2163 — `animationData.sprites = LoadCl2Sheet(pszName, animationWidth)`
- Source/player.cpp:2167-2176 — graphic TRN then class TRN applied at load
- Source/player.cpp:2147 — death forces `PlayerWeaponGraphic::Unarmed`
- Source/player.cpp:2072 — `prefixBuf = { classChar, ArmourChar[_pgfxnum>>4], WepChar[_pgfxnum&0xF] }`
- Source/player.cpp:2199 — `SetPlrAnims` (frames/`_pAFNum` per weapon from `GetPlayerAnimDataForClass`)
- Source/player.cpp:2286 — warrior/barbarian bow idle = 8 frames, armour index → `_pDFrames = 15`
- Source/player.cpp:2059 — `GetPlayerSpriteWidth(cls, graphic, weapon)`
- Source/missiles.h:102 — `OptionalClxSpriteList _miAnimData`
- Source/missiles.h:110 — `int16_t _miAnimWidth2` (precomputed half-width)
- Source/missiles.h:113 — `_miAnimFrame` (1-based)
- Source/missiles.h:114-117 — `_miDrawFlag`, `_miLightFlag`, `_miPreFlag`, `_miUniqTrans`
- Source/engine/render/scrollrt.cpp:339 — `DrawMissilePrivate` (`offsetForRendering - {_miAnimWidth2, 0}`, `[*_miAnimData][_miAnimFrame-1]`)
- Source/engine/render/scrollrt.cpp:362 — `DrawMissile` (bucket lookup)
- Source/engine/render/scrollrt.cpp:146 — `UpdateMissilePositionForRendering` (fixed-point interpolation)
- Source/engine/render/scrollrt.cpp:159 — `UpdateMissileRendererData` (collision-aware rollback loop)
- Source/engine/render/scrollrt.cpp:714 — `DrawItem` (`item.AnimInfo.currentSprite()`, `getRenderingOffset`)
- Source/engine/render/scrollrt.cpp:733 — `DrawMonsterHelper` (`dMonster` abs−1, towner vs monster)
- Source/engine/render/scrollrt.cpp:378 — `DrawMonster` TRN decision tree
- Source/engine/render/scrollrt.cpp:468 — `DrawPlayer`
- Source/engine/render/scrollrt.cpp:407 — `DrawPlayerIconHelper` (mana shield / reflect icons, `+{0,16}` for reflect)
- Source/engine/render/scrollrt.cpp:524 — `DrawObject` (`(*_oAnimData)[_oAnimFrame-1]`)
- Source/objects.h:268 — `currentSprite()`, `:272 getRenderingOffset(sprite, tilePosition)` (large-object world offset)
- Source/items.h:466 — `getRenderingOffset` (items never add a walk offset)
- Source/items.h:184 — `constexpr int ItemAnimWidth = 96`
- Source/items.cpp:2424 — `InitItemGFX` — `items\<ItemDropNames[i]>` (35 types in Diablo, `ITEMTYPES=43` in Hellfire)
- Source/items.cpp:2403 — `GetOutlineColor(item, checkReq)`
- Source/items.cpp:3818 — `item.AnimInfo.sprites = itemanims[it]`
- Source/towners.cpp:110 — `towner.ownedAnim = LoadCel(path, towner._tAnimWidth)`
- Source/towners.cpp:780 — `towners\animals\cow`, width 128
- Source/towners.h:77 — `getRenderingOffset()` (towner variant, no sprite arg)
- Source/monster.h:472 — `getRenderingOffset`
- Source/engine/render/scrollrt.cpp:1575 — `GetOffsetForWalking` (the `MovingOffset[8]` table at `:1579`)
- Source/engine/render/scrollrt.cpp:1579 — `MovingOffset[8] = {{0,32},{-32,16},{-64,0},{-32,-16},{0,-32},{32,-16},{64,0},{32,16}}`
- Source/cursor.cpp:441 — `InitCursor()` (`data\inv\objcurs.clx` / `objcurs.cel` + widths txt)
- Source/cursor.cpp:464 — `GetInvItemSprite(cursId)` (1-based, spills into `objcurs2`)
- Source/cursor.cpp:476 — `GetInvItemSize(cursId) = {w,h}` from the sprite itself
- Source/cursor.cpp:492 — `CreateHalfSizeItemSprites` (bilinear half-size + red item art)
- Source/cursor.cpp:552 — `DrawItem` (usable → plain, unusable → `infra.trn`)
- Source/cursor.cpp:601 — `DrawSoftwareCursor(out, position, cursId)`
- Source/cursor_defs.hpp:22 — `enum cursor_id : uint8_t` (NONE=0 … FIRSTITEM=12)
- Source/engine/render/scrollrt.cpp:279 — `DrawCursor` (outline width 1 when holding an item, hotspot = `cursSize/2`)
- Source/engine/render/scrollrt.cpp:215 — `BlitCursor` (memcpy save/restore of the pixels behind the cursor)
- Source/engine/render/scrollrt.cpp:225 — `UndrawCursor`
- Source/automap.cpp:38 — `enum MapColors` (indices `PAL8_YELLOW`, `PAL16_YELLOW+8`, `PAL8_BLUE+1`, …)
- Source/automap.cpp:167 — `DrawDiamond` (4 `DrawMapLine*` calls)
- Source/automap.cpp:275 — `DrawDirt` (fixed `SetMapPixel` stamp list)
- Source/automap.cpp:980 — `DrawAutomapTile` (colours per `MapExplorationType`, `noConnect` rules, lava colour choice)
- Source/automap.cpp:1323 — `GetPlayerMapColor`; `:1341` `DrawAutomapPlr` (per-`_pdir` arrow = 3 line primitives,
  direction switch at `:1373-1424`)
- Source/automap.cpp:1427 — `DrawAutomapText`
- Source/automap.cpp:1752 — `DrawAutomap` (the double loop `map.x+j, map.y-j` at `:1847-1861`)
- Source/automap.cpp:1551 — `int AutoMapScale;` (`:1559 InitAutomapOnce: AutoMapScale = 50`)
- Source/automap.h:66 — `AmLineLength { QuarterTile=2, HalfTile=4, FullTile=8, FullAndHalf=12, DoubleTile=16, OctupleTile=64 }`
- Source/engine/render/automap_render.cpp:26 — `DrawMapLine<DirX,DirY>` (per-pixel step pattern); `:43`
  `DrawMapLineSteep<DirX,DirY>`; wrappers `:60` NS, `:78` WE, `:96` NE, `:101` SE, `:106` NW, `:111` SW,
  `:116/:121/:126/:131` steep variants
- Source/engine/render/automap_render.cpp:139 — `DrawMapFreeLine` (Bresenham)
- Source/engine/render/automap_render.cpp:166 — `SetMapPixel` (minimap clip `:168`, transparent half-blend `:171`)
- Source/engine/displacement.hpp:247 — `fromDirection`
- Source/engine/direction.hpp:11 — `enum class Direction { South, SouthWest, West, NorthWest, North, NorthEast, East, SouthEast, NoDirection }`
- test/animationinfo_test.cpp — existing vectors for the distribution logic
- test/cursor_test.cpp — cursor/`GetInvItemSize` cases

### Formats & algorithms

**A. Animation progression.** `AnimationInfo` holds `sprites` (a `ClxSpriteList` for the current *direction*),
`numberOfFrames`, `currentFrame`, `ticksPerFrame`, plus the render-smoothing distribution fields. The renderer only
calls `currentSprite()`; the frame index it returns is `getFrameToUseForRendering()` (`Source/engine/animationinfo.cpp:18`),
which, when distribution is active, computes
```
absoluteAnimationFrame = (getProgressToNextGameTick() + ticksSinceSequenceStarted_) * tickModifier_
                          / baseValueFraction / baseValueFraction          # baseValueFraction = 128
```
minus `skippedFramesFromPreviousAnimation_`, clamped to `[0, relevantFramesForDistributing_-1]`.
`getAnimationProgress()` (`:60`) returns the *fraction within the current frame*, 0..128:
```
totalTicks = progressToNextGameTick + ticksSinceSequenceStarted
progressInFrames = totalTicks * tickModifier
animationFraction = progressInFrames / numberOfFrames / 128
```
with a fallback branch when distribution is off: `ticksSinceSequenceStarted = (currentFrame*ticksPerFrame +
tickCounterOfCurrentFrame) * 128`, `tickModifier = 128 / ticksPerFrame`. All integer; port bit-exactly (dg-task-016
covers it; the renderer consumes it as `uint8 ∈ [0,128]`).

**B. Frame selection by direction (8×).** Sprite lists are stored per direction:
- Players: `LoadCl2Sheet(path, width)` returns a **sheet** of 8 lists; `spritesForDirection(dir) = sheet()[dir]`
  (`Source/monster.h:166` is the same accessor pattern used by monsters). Direction index order = the enum order
  `South, SouthWest, West, NorthWest, North, NorthEast, East, SouthEast` (`Source/engine/direction.hpp:11`).
- Monsters: same, but the CL2 may be a plain list (one direction only) for some specials — `isSheet()` sniffing
  decides (`Source/monster.cpp:3566-3570`).
- Animation → file letter: players `as aw at ht bl lm fm qm dt` (`Source/player.cpp:2103-2150`, town = `st`/`wl`),
  monsters `"nwahds"` (`Source/monster.cpp:150`: n=neutral/idle, w=walk, a=attack, h=hit, d=death, s=special).
- Frames/ticks come from tables: monsters `frames[6]`/`rate[6]` columns of
  `assets/txtdata/monsters/monstdat.tsv` (113 lines incl. header; e.g. `MT_NZOMBIE` →
  `frames=11,24,12,6,16,0`, `rate=4,1,1,1,1,1`); players from `GetPlayerAnimDataForClass`
  (`_pNFrames/_pWFrames/_pAFrames/_pAFNum/_pHFrames/_pBFrames/_pDFrames/_pSFrames/_pSFNum`,
  `Source/player.cpp:2199-2290`).

**C. Composited player sprites (plrgfx).** There is **no** separate body/weapon/shield layer compositing: one CL2 file
per (class, armour, weapon, animation) triple carries the fully drawn character. The file name is
```
plrgfx\<classPath>\<c><a><w>\<c><a><w><anim>       # e.g. plrgfx/warrior/hu/huaw.cl2
  <c> = spriteData.classChar ('h' warrior, 'r' rogue, 's' sorcerer, 'm' monk, 'b' barbarian, 'n' bard)
  <a> = ArmourChar[_pgfxnum >> 4]                  # armour graphic index
  <w> = WepChar[_pgfxnum & 0xF]                    # weapon graphic index
```
(`Source/player.cpp:2155` + `:1530` + `:2072`.) `_pgfxnum` is computed by `CalcPlayCharWidth`-style logic in
`player.cpp` from equipped items (`player_graphic::Attack` upgrades the weapon graphic with `PlayerWeaponGraphic::SwordShield`
etc., `Source/player.cpp:1490-1497`). Sprite *widths* are per (class, graphic, weapon) from `PlayerSpriteData`
(`Source/tables/playerdat.hpp:143-170`, fields `stand walk attack bow swHit block lightning fire magic death`),
read from `assets/txtdata/classes/<class>.tsv` (`Source/tables/playerdat.cpp:237-245`). Per-graphic TRN (`whufm.trn`
= "warrior, heavy armour, fire magic") and class TRN (`plrgfx/warrior.trn`) are applied at load, in that order.

**D. Sprite anchoring — the one rule that governs all entity placement.**
```
screenPosition = tileBottomCorner + getRenderingOffset(sprite)
getRenderingOffset(sprite) = { -CalculateSpriteTileCenterX(sprite.width()), 0 } [+ walk offset if walking]
CalculateSpriteTileCenterX(w) = (w - TILE_WIDTH) / 2                # Source/levels/dun_tile.hpp:138
```
(`Source/player.h:785`, `Source/monster.h:472`, `Source/items.h:466`, `Source/objects.h:272`.)
So a 128-wide monster sprite at tile (x,y) is drawn at `bufferX - (128-64)/2 = bufferX - 32`, and `bufferX` is the
**left edge** of the tile's 64×32 diamond, `bufferY` the diamond's **bottom** vertex; CLX/CL2 rows are bottom-up, so
`(bufferX, bufferY)` is the sprite's bottom-left corner. Objects additionally subtract `worldOffset.worldToScreen()`
when the tile queried isn't the object's anchor tile (large objects occupy several tiles)
(`Source/objects.h:272-280`). Corpses re-center with `CalculateSpriteTileCenterX(corpse.width)` directly
(`Source/engine/render/scrollrt.cpp:798`). Missiles use the precomputed half-width: `targetBuffer + offsetForRendering
− {_miAnimWidth2, 0}` (`Source/engine/render/scrollrt.cpp:345`).

**E. Walking interpolation.** `GetOffsetForWalking(AnimInfo, dir, cameraMode=false)`
(`Source/engine/render/scrollrt.cpp:1575`):
```
MovingOffset[8] (indexed by Direction) = {S:(0,32) SW:(-32,16) W:(-64,0) NW:(-32,-16) N:(0,-32) NE:(32,-16) E:(64,0) SE:(32,16)}
offset = MovingOffset[dir] * animationProgress / 128        # integer; animationProgress ∈ [0,128]
if cameraMode: offset = -offset                             # used for the viewport scroll (CalcFirstTilePosition)
```
Note: a walk covers exactly one tile in screen space (64×±32/±16 px depending on axis) per tick.

**F. Draw order / "y-sort".** There is no z-buffer and no screen-space sort: **painter's algorithm over the isometric
row-walk** (G1 #4). Within one tile, `DrawDungeon` (`Source/engine/render/scrollrt.cpp:781`) is authoritative:
```
1 DrawCell                       (tile micros, incl. foliage)
2 missiles with _miPreFlag       (only if MissilePreFlag global set)
3 corpse (dCorpse)               (skipped if lightTableIndex == 15)
4 objects with _oPreFlag
5 items with !_iPostDraw
6 dead players (dFlags DeadPlayer)
7 players (via dPlayer, incl. the negative-id trick)
8 monsters (via dMonster, incl. the negative-id trick)
9 items with _iPostDraw
10 objects without _oPreFlag
11 missiles (non-pre)
12 dSpecial arch  (blended if TransList[bMap])
```
The **negative-id trick** (`Source/engine/render/scrollrt.cpp:830-866` for players, `:868-899` for monsters): when an
entity is walking South-West/South/South-East (`PM_WALK_SOUTHWARDS` / `MonsterMode::MoveSouthwards`) or walking East
sideways, its id is stored *negated* in `dPlayer`/`dMonster` at the tile it is leaving, and the renderer, upon seeing
`dPlayer[tile] == -(id+1)`, draws the sprite offset back by one tile (`+{TILE_WIDTH/2,-TILE_HEIGHT/2}`,
`+{0,-TILE_HEIGHT}`, `+{-TILE_WIDTH/2,-TILE_HEIGHT/2}`, `+{-TILE_WIDTH,0}`) — so it paints *from the earlier tile*,
which is earlier in draw order, so it correctly appears in front of things in the tile it enters. Porting note: the
sign convention lives in the entity grid (`dg-task-016`), the *offsets* live in the renderer (`dg-task-010`).

**G. Cursor.** `pcurs` is an index into `objcurs`(+`objcurs2`) **for item cursors only from `CURSOR_FIRSTITEM`(=12)
onwards**; ids 0..11 are the fixed set `NONE, HAND, IDENTIFY, REPAIR, RECHARGE, DISARM, OIL, TELEKINESIS, RESURRECT,
TELEPORT, HEALOTHER, HOURGLASS` (`Source/cursor_defs.hpp:22`). Sizes are *not* hard-coded: `GetInvItemSize(cursId)`
returns the CLX frame's own `w,h` (`Source/cursor.cpp:476`), and the per-frame widths for the CEL path come from
`assets/data/inv/objcurs-widths.txt` (one decimal per line: 33, then 32×10, then 28×… , then 56×…, i.e. small icons
32 px, item icons 28/56 px depending on grid size). Drawing (`Source/engine/render/scrollrt.cpp:279` `DrawCursor`):
```
outlineWidth = MyPlayer->HoldItem.isEmpty() ? 0 : 1
offset       = empty ? {0,0} : {cursSize/2}
cursPosition = MousePosition - offset
rect = (cursPosition - {o,o}, cursSize + 2o) clipped to the surface
BlitCursor(behindBuffer <- out[rect])              # save what's under the cursor
DrawSoftwareCursor(out, cursPosition + {0, cursSize.height - 1}, pcurs)   # anchor = bottom-left
```
and `UndrawCursor` restores `behindBuffer` at the *start* of the next frame (because the cursor is drawn into the same
buffer as everything else and the C++ renderer does partial blits). **Not needed in Ebitengine** — we redraw the frame
fully; mark as "not ported", but keep the *hotspot* semantics.

**H. Automap.** Two coordinate systems: `Automap` is in **mega-tile** space (`DMAXX×DMAXY = 40×40`), the render loop
walks `map = {Automap.x - cells, Automap.y - 1}` with the inner pair `(map.x + j, map.y - j)` — i.e. a 45°-rotated
scan like the world walk but with `AmOffset` steps instead of `TILE_*`:
`d = (AutoMapScale*64)/100` and `cells = 2*(gnScreenWidth/2/d)+1 (+1 (+1))` (`Source/automap.cpp:1776-1781`),
`AutoMapScale` starts at 50 and is scaled by the zoom hotkeys (`Source/automap.cpp:1559`, `:1734`).
Per tile, `DrawAutomapTile` (`:980`) picks `colorDim`/`colorBright` from `AutomapView[x][y]`
(`MAP_EXP_SHRINE → PAL16_GRAY+11 / PAL16_GRAY+3`, `MAP_EXP_OTHERS → PAL16_BEIGE+10 / PAL16_BEIGE+2`, else
`MapColorsDim = PAL16_YELLOW+8`, `MapColorsBright = PAL8_YELLOW`, `Source/automap.cpp:38-57`), then draws in order:
dirt stamp → stairs → wall-connection lines → the tile-shape primitive(s) (diamond / vertical / horizontal /
arch / grate / lava / water / cave-cross / fence variants) → optional portal/mission highlight.
Primitives are **1-px Bresenham lines in three families** (`DrawMapLineNE/SE/NW/SW`, `DrawMapLineSteep*`,
`DrawMapLineWE/NS`, `DrawMapFreeLine`) whose pixel patterns are defined in
`Source/engine/render/automap_render.cpp:26-140` — they are *not* symmetric anti-aliased lines; the exact step
pattern must be transcribed. Colours are palette indices, so the automap is index-exact.
Player arrow: 3 primitives per direction (`Source/automap.cpp:1373-1424`), colour =
`MapColorsPlayer1..4` (`PAL8_ORANGE+1`, `PAL8_YELLOW+1`, `PAL8_RED+1`, `PAL8_BLUE+1`).
Transparent mode routes every pixel through `paletteTransparencyLookup[dst][color]`
(`Source/engine/render/automap_render.cpp:166`, `Source/engine/render/primitive_render.cpp:215`); minimap mode additionally clips to
`MinimapRect` and fills a half-transparent background + a 2-px frame (`Source/automap.cpp:1790-1800`).

### Ebitengine mapping
- Entity sprites: one atlas per family (players, monsters, items, objects, missiles, cursors, fonts, ui). Each frame =
  `(atlasRect, w, h)`; draw with `op.GeoM.Translate(x, y-h)` (bottom-left anchoring) and the light-variant palette
  (G4/G5). **Never** use `GeoM.Scale` for entity sprites — the C++ is 1:1 pixels at the logical resolution.
- Walk interpolation: compute in `Draw` from `animationProgress` (0..128) — integer math identical to
  `Source/engine/render/scrollrt.cpp:1575`; `Ebitengine` `GeoM.Translate(float)` accepts the integer directly. Since
  all values are integers, exactness is guaranteed (no sub-pixel filtering because `FilterNearest`).
- Draw order: implement the row-walk and `DrawDungeon` body literally (dg-task-010). Keep the negative-id convention in
  `dPlayer`/`dMonster` so the ordering logic is testable against C++ traces.
- Cursor: prefer drawing the CLX cursor sprite into the logical screen (software cursor) and
  `ebiten.SetCursorMode(ebiten.CursorModeHidden)`. `ebiten.SetCursorShape` (verified in
  https://github.com/hajimehoshi/ebiten/blob/main/cursor.go — 10 shapes: Default, Text, Crosshair, Pointer,
  EWResize, NSResize, NESWResize, NWSEResize, Move, NotAllowed; example
  https://github.com/hajimehoshi/ebiten/blob/v2.9.2/examples/cursor/main.go) cannot express Diablo's item cursors, so
  **hardware-cursor replacement is "not ported"** — the whole `Source/hwcursor.cpp` (SDL_CreateColorCursor, bilinear
  scaling of cursors, `hardwareCursorMaxSize`) is an OS integration nicety with zero effect on framebuffer pixels.
  Exception: the *menus* (DiabloUI) may keep `CursorShapeDefault/Pointer` for UX.
- Automap: two options — (a) transcribe the line primitives into Go and plot pixels through the same palette-shader
  path (needed for pixel parity, and cheap because lines are ≤ 16 px long); (b) `ebiten/vector` strokes, which is
  anti-aliased and *cannot* match. **Choose (a).** Implement as stamps into a small 640×480 "automap index overlay"
  that the automap pass reads, so the pixel plotter stays the same code path as C++'s `SetPixel`.

### Parity test
- `TestAnimationProgress`: port `test/animationinfo_test.cpp` vectors; assert both `currentFrame` and
  `animationProgress(0..128)` at `progressToNextGameTick ∈ {0,64,127}`.
- `TestRenderingOffsets`: table of (family, sprite width) → expected `getRenderingOffset`; assert
  `-(w-64)/2` and the walk-offset variants for all 8 directions at progress 0/64/128.
- `TestDrawOrderTrace`: scripted 3×3 scene (1 corpse, 1 pre-flag object, 1 post-draw item, 1 walking-east player, 1
  walking-south monster, 1 pre missile, 1 arch) → ordered list of drawn kinds must equal the list from
  `Source/engine/render/scrollrt.cpp:781-952` (this is dg-epic-03's "order-instrumented test" acceptance criterion,
  made concrete).
- `TestCursorSizes`: assert `GetInvItemSize` for every cursor id from a C++ dump (widths come from the file, so this
  doubles as a CLX decode test).
- `TestAutomapTileStamps`: for each `AutomapTile::Types` × `Flags` combination, render into a 32×32 surface in C++
  (`test/automap_test.cpp` harness) and compare pixel-by-pixel with the Go stamp renderer.

### Open questions
- Are the Hellfire classes (monk/barbarian/bard) required at M1? They multiply plrgfx inventory by ~2.
- Item label outlines use `GetOutlineColor` (`Source/items.cpp:2403`) with `PAL16_*+n` indices — confirm we want the
  QoL label feature in the parity screenshots (it is optional in options; must be *off* in golden frames).

---

## G7. TEXT & FONTS

### Sources (file:line)
- Source/engine/render/text_render.cpp:46 — `OptionalOwnedClxSpriteList pSPentSpn2Cels;`
- Source/engine/render/text_render.cpp:49 — `struct OwnedFontStack { baseFont, overrideFont }`
- Source/engine/render/text_render.cpp:60 — `FontStack::glyph(i)` — override wins when `width != 0`
- Source/engine/render/text_render.cpp:79 — `ankerl::...map<uint32_t, OwnedFontStack> Fonts;` (cache)
- Source/engine/render/text_render.cpp:81 — `FontSizes = {12, 24, 30, 42, 46, 22}`
- Source/engine/render/text_render.cpp:82 — `LineHeights = {12, 26, 38, 42, 50, 22}`
- Source/engine/render/text_render.cpp:83 — `SmallFontTallLineHeight = 16`
- Source/engine/render/text_render.cpp:84 — `BaseLineOffset = {-3,-2,-3,-6,-7,3}`
- Source/engine/render/text_render.cpp:86 — `ColorTranslations[19]` (TRN paths)
- Source/engine/render/text_render.cpp:152 — `GetUnicodeRow(cp) = cp >> 8`
- Source/engine/render/text_render.cpp:157 — `IsCJK(row) = 0x30..0x9f`, `:162 IsHangul`, `:167 IsSmallFontTallRow`
- Source/engine/render/text_render.cpp:171 — `GetFontPath(size,row,ext) = "fonts\<size>-<hexrow><ext>"`
- Source/engine/render/text_render.cpp:176 — language override path `"fonts\<lang>\<size>-<row>.clx"`
- Source/engine/render/text_render.cpp:181 — `GetFontId = (size<<16)|row`
- Source/engine/render/text_render.cpp:188 — `LoadFont(size,color,row)` (TRN lazy-load, lang override, base, PCX fallback)
- Source/engine/render/text_render.cpp:234 — `class CurrentFont` (per-codepoint-row lazy switch)
- Source/engine/render/text_render.cpp:269 — `DrawFont(out,pos,glyph,color,outline)`
- Source/engine/render/text_render.cpp:400 — `ClipSurface`
- Source/engine/render/text_render.cpp:409 — `AdjustSpacingToFitHorizontally`
- Source/engine/render/text_render.cpp:421 — `MaybeWrap`
- Source/engine/render/text_render.cpp:427 — `GetLineStartX` (align left/center/right)
- Source/engine/render/text_render.cpp:440 — `DrawLine` (the glyph walk: `frame = cp & 0xFF`, cursor blink, highlight
  rect, `x += charWidth + curSpacing`)
- Source/engine/render/text_render.cpp:493 — `currentFont.load(size,color,c)` with `'?'` fallback
- Source/engine/render/text_render.cpp:499 — `DoDrawString`; `:522`/`:634`/`:668`/`:753` measure & fade paths
- Source/engine/render/text_render.cpp:843 — `opts.lineHeight == -1 → GetLineHeight(text,size)`
- Source/engine/render/text_render.cpp:867 — `MaybeWrap(characterPosition, sprite.width(), rightMargin, initialX, lineHeight)`
- Source/engine/load_pcx.cpp:35 — `LoadPcxSpriteList` (legacy PCX fonts)
- Source/engine/render/clx_render.hpp:40 — `RenderClxSprite` (top-left anchored!) — text uses this family;
  `:76` `RenderClxSpriteWithTRN`, `:81` `ClxDrawBlendedTRN`, `:89` `ClxDrawBlended`, `:107`
  `ClxDrawBlendedWithLightmap`, `:112` `IsPointWithinClx`, `:25/:26` `ClxApplyTrans(List|Sheet)`
- Source/engine/assets.hpp:309 — `MainMpqPriority = 1000`, `:310 DevilutionXMpqPriority = 9000`,
  `:311 LangMpqPriority = 9100`, `:312 FontMpqPriority = 9200`; `:344 HaveMainData`, `:346 HaveExtraFonts`
- Source/engine/render/text_render.cpp:295 — `CountNewlines`; `:300+` `DrawStringFormatArg` machinery
- test/text_render_integration_test.cpp — golden PNG comparison harness
- test/fixtures/text_render_integration_test/ — committed golden PNGs (`basic.png`, `basic-colors.png`,
  `kerning_fit_spacing__align_center-colors.png`, `horizontal_overflow.png`, `cursor-*.png`, `highlight-*.png`)
- assets/fonts/12-00.clx … 46-05.clx — 76 `.clx` font files in the repo
- assets/fonts/tr/12-00.clx — Turkish override example

### Formats & algorithms
**Font file = CLX list of 256 glyphs, indexed by the low byte of the Unicode code point** (`frame = c & 0xFF`,
`Source/engine/render/text_render.cpp:492`), one file per (pixel size, Unicode high row). Sizes
`{12,24,30,42,46,22}` px; rows are hex-padded 2 (`00`, `01`, …, `1f1`, i.e. `AsHexPad2(row)` — `assets/fonts/12-1f1.clx`
exists). Verified on the real file `assets/fonts/30-00.clx`: 256 frames, offsets step by 6 for the empty glyphs
(`header 06 00 00 00 26 00` = width 0, height 38 → **zero-width glyph, height only**), total file 41 201 bytes.
So: **glyph advance == glyph width** (no kerning table — the only "kerning" is the fit-horizontally spacing reducer,
`AdjustSpacingToFitHorizontally`, `:409`), and empty glyphs keep a height for line metrics.
Line metrics: `LineHeights[size]`, `BaseLineOffset[size]` (applied by callers positioning text), and
`SmallFontTallLineHeight = 16` used for the 12 px font when the text contains CJK/Hangul rows
(`IsSmallFontTallRow`, `:167`).

Colour: `text_color` → optional TRN file (`ColorTranslations`, `:86`); `DrawFont` (`:269`) then calls
`RenderClxSpriteWithTRN` (top-left anchored, `Source/engine/render/clx_render.hpp:76`) or plain
`RenderClxSprite`. Outline mode draws `ClxDrawOutlineSkipColorZero(out, 0, {x, y+h-1}, glyph)` first (colour index 0 =
black) — `:271`.

Wrapping algorithm (`DoDrawString` → `DrawLine`, `:440`, `:867`): greedy per code point —
```
if (x + glyphWidth > rightMargin) { x = initialX; y += lineHeight }
```
Word-aware variants exist for `DrawStringWithFade`/measure paths (`:522`, `:634`, `:668`, `:753` measure functions
`GetLineWidth`/`MeasureText`) which break on spaces and, for CJK, allow breaks before full-width punctuation
(`IsBreakAllowed`, `:284`). `ZWSP (U+200B)` is skipped silently (`:478`).

### Ebitengine mapping
**Decision: pre-baked glyph atlas per (font size × colour TRN), drawn with plain `DrawImage` + `FilterNearest`.**
Rationale:
- Glyph count is bounded: 6 sizes × ≤ 8 rows actually used per language × 256 glyphs; colour variants ~16.
  Baking `size × colour` (not `codepoint × colour`) is fine because a glyph's shape is colour-independent — we bake
  **one index atlas per size** and select the *palette variant* per colour (same mechanism as G4!). That is strictly
  better than 16 atlases: one glyph atlas per size (6), and 16 palette-variant textures per size (256×16×3 B ≈ 12 KB
  total). Zero extra memory, exact colours.
- A Kage "colorize" shader is not needed and would break byte-exactness (float multiply). Do **not** use
  `ColorScale.ScaleWithColor` for text.
- `go:textatlas`? Not applicable. Build the atlas at load with one `ReplacePixels` per page from an
  `image.Paletted` whose colour 0..255 = identity indices (i.e. store the raw index in the R channel as for sprites).
- CJK path: Ebitengine has no role here — the CJK glyphs *are* in `fonts.mpq` row files (base + language override), so
  port the row-switching logic (`GetUnicodeRow`, `CurrentFont::load`) and the `width()==0 → use base` fallback
  (`FontStack::glyph`, `:60`). If `fonts.mpq` is absent (our asset set ships only some rows), fall back to
  `golang.org/x/image/font` + `text/v2` for CJK and mark the result "not pixel-identical, non-parity path". Note
  `text.CacheGlyphs` guidance (https://ebitengine.org/en/documents/2.1.html) if we ever go that route.

### Parity test
Reuse the **existing goldens**: `test/fixtures/text_render_integration_test/*.png` are committed in the repo (verified
10+ files). Port `test/text_render_integration_test.cpp`: same input strings, same fonts, same rects → compare the
Go-rendered logical screen (ReadPixels in the test) to those PNGs. That is an *end-to-end* parity test that validates
decoder + TRN + palette + wrap + spacing at once. Add `TestMeasureWrap` asserting the (x,y) break points for the same
20 strings used in the C++ test.

### Open questions
- Which font files do we ship? `assets/fonts` has 76 `.clx` in the repo; full CJK rows live in a separately downloaded
  `fonts.mpq`. Decide the M1 language set (Volume I §11).
- `pSPentSpn2Cels` (the pentagram spinner used in text? `Source/engine/render/text_render.cpp:46`) — confirm it's
  needed for parity frames (it is level-art, not text; keep out of the font module).

---

## G8. EFFECTS (transparency, fades, flash, QoL overlays, screenshot)

### Sources (file:line)
- Source/utils/palette_blending.hpp:16 — `extern uint8_t paletteTransparencyLookup[256][256];`
- Source/utils/palette_blending.hpp:28 — `GenerateBlendedLookupTable(palette, skipFrom, skipTo)` ("based on the same
  technique found in Quake2")
- Source/utils/palette_blending.cpp:19 — the array definition
- Source/utils/palette_blending.cpp:30 — `BlendColors(a,b) = ((r1+r2)/2, (g1+g2)/2, (b1+b2)/2)` (integer divide)
- Source/utils/palette_blending.cpp:46 — generation: symmetric, `lut[i][i]=i`, nearest palette colour via KD-tree
- Source/utils/palette_blending.cpp:72 — `UpdateBlendedLookupTableSingleColor`
- Source/utils/palette_blending.cpp:22 — `paletteTransparencyLookupBlack16[65536]` (optional fast path, BE/LE packing)
- Source/utils/palette_kd_tree.hpp — `PaletteKdTree::findNearestNeighbor`
- Source/engine/palette.cpp:188 — `LoadPaletteAndInitBlending` skip ranges per level type
- Source/engine/palette.cpp:52 / :71 — cycling rotates the LUT rows *and* columns (`:57-60`)
- Source/engine/render/blit_impl.hpp:99 — `BlitFillBlended` (`dst = lut[color][dst]`)
- Source/engine/render/blit_impl.hpp:106 — `BlitPixelsBlended` (`dst = lut[src][dst]`)
- Source/engine/render/blit_impl.hpp:113 — `BlitBlended`
- Source/engine/render/blit_impl.hpp:120 — `BlitPixelsBlendedWithMap` (`dst = lut[dst][tbl[src]]` — note argument order!)
- Source/engine/render/blit_impl.hpp:126 — `BlitFillBlendedWithMap` (`lut[colorMap[color]][dst]` per px)
- Source/engine/render/blit_impl.hpp:133 — `BlitBlendedWithMap`
- Source/engine/render/blit_impl.hpp:146 — `BlitFillBlendedWithLightmap`
- Source/engine/render/blit_impl.hpp:159 — `BlitPixelsBlendedWithLightmap` (two-pass, `<1024` scratch path)
- Source/engine/render/clx_render.cpp:571 — `ClxDrawBlended`, `:576 ClxDrawBlendedTRN`, `:581 ClxDrawBlendedWithLightmap`
- Source/engine/render/scrollrt.cpp:264 — `ClxDrawLightBlended`
- Source/engine/render/scrollrt.cpp:926 — arch blended when `TransList[bMap]`
- Source/engine/render/dun_render.cpp:159 — `RenderLineTransparent<FullyDark>` = `BlitFillBlended(dst,n,0)`
- Source/engine/render/primitive_render.cpp:18 — half-transparent line path uses `lut[color]` (write `:22`)
- Source/engine/render/primitive_render.cpp:37 — `paletteTransparencyLookupBlack16` fast path (`:44`)
- Source/engine/render/primitive_render.cpp:89 / `:107` — `DrawHorizontalLine` / `DrawVerticalLine` (unsafe `:102`/`:120`,
  half-transparent variants `:140`/`:155`)
- Source/engine/render/primitive_render.cpp:159 and `:187` — `DrawHalfTransparentRectTo` variants
- Source/engine/render/primitive_render.cpp:215 — `SetHalfTransparentPixel`
- Source/engine/render/light_render.cpp:527 — `Lightmap::bleedUp` (light bleeding up walls; per-pixel mode only)
- Source/qol/floatingnumbers.cpp:22 — `struct FloatingNumber`
- Source/qol/floatingnumbers.cpp:58 — `AddFloatingNumber(pos, offset, text, style, id, reverseDirection)`
- Source/qol/floatingnumbers.cpp:84 — `DrawFloatingNumbers(out, viewPosition, offset)` (screen-space, `mul = 1 - timeLeft/2500.0F`)
- Source/qol/itemlabels.cpp:104 — `AddItemToLabelQueue(id, position)`
- Source/qol/itemlabels.cpp:147 — `DrawItemNameLabels(out)`
- Source/qol/monhealthbar.cpp — monster health bars (drawn in `DrawView`, `Source/engine/render/scrollrt.cpp:1396`)
- Source/qol/xpbar.cpp:79 — `DrawXPBar(out)` (called from `DrawAndBlit`, `:1885`)
- Source/panels/info_box.cpp — the info box; `Source/engine/render/scrollrt.cpp:1443` `DrawInfoBox`
- Source/engine/render/scrollrt.cpp:1460 — `DrawFPS`
- Source/capture.cpp:19 — `DEVILUTIONX_SCREENSHOT_FORMAT_PCX/PNG` selectors
- Source/capture.cpp:84 — `CaptureScreen()` (calls `DrawAndBlit()` first!)
- Source/capture.cpp:97 — `RedPalette()` (makes a red flash during capture, restores right after)
- Source/utils/surface_to_pcx.cpp:45 — `WritePcxHeader`, `:70 WritePcxPalette`, `:92 WritePcxLine`, `:150 WriteSurfaceToFilePcx`
- Source/utils/surface_to_png.hpp — PNG writer alternative
- Source/engine/dx.cpp:287 — `RenderPresent`
- Source/diablo.cpp:3521 — colour cycling call site
- Source/hwcursor.cpp:112 — `SetHardwareCursorFromSurface` (OS nicety)
- Source/effects.cpp:190 — `sgSFX` table (audio-only; **no** light/flash effects live here — verified by grep)

### Formats & algorithms

**A. Transparency blending = the 256×256 LUT.** Semantics (all index-space):
```
blend(dstIndex, srcIndex) -> newIndex
```
Built per palette as: `lut[i][j] = nearestPaletteIndexOf( (palette[i] + palette[j]) / 2 )` (integer halving),
`lut[i][i] = i`, symmetric (`:52-56`). `skipFrom..skipTo` excludes a palette range from being a *candidate result*
(caves/crypt skip 1..31 = the cycling range, nest skips 1..15) so blends don't land on animated colours.
Because it is a function of **both** pixels and the current palette, it must be rebuilt on `LoadPalette` and
*rotated* on every cycle step (`Source/engine/palette.cpp:57-60`).
Call sites, with exact operand order (they differ!):
- sprite over background, blended: `dst = lut[dst][tbl[src]]` (`BlitPixelsBlendedWithMap`, `Source/engine/render/blit_impl.hpp:120`)
  for CLX sprites via `ClxDrawBlendedTRN`; `dst = lut[src][dst]` for un-translated sprites
  (`BlitPixelsBlended`, `:106`) — **these two are asymmetric because `lut` is symmetric in practice; but
  `BlitFillBlendedWithMap` (`:133`) uses `lut[colorMap[color]][dst]`, so the fill path swaps the arguments relative to
  the pixels path. Transcribe each of the 6 variants literally.**
- tile transparent line: `RenderLineTransparent<PartiallyLit>` → `BlitPixelsBlendedWithMap(dst,src,n,tbl)`;
  `<FullyDark>` → `BlitFillBlended(dst,n,0)` i.e. `dst = lut[0][dst]` (dark-colour blend, `Source/engine/render/dun_render.cpp:162`);
  `<FullyLit>` → `BlitPixelsBlended(dst,src,n)`.
- automap transparent mode: `lut[color][dst]` per pixel (`Source/engine/render/automap_render.cpp:168` →
  `primitive_render.cpp:215`).
- half-transparent rects (minimap background): `DrawHalfTransparentRectTo` — `lut[0][...]`-style darkening (`:159`, `:187`).

**B. Flash overlays / screen tinting.** There is no global "flash" buffer; the effects are: (1) whole-palette fades
(`PaletteFadeIn/Out`, G1 §7), (2) palette *cycling* (`palette_update_caves/crypt/hive`, G1 §7 — the only "animated
colour" mechanism), (3) `flashlight`-style object lights implemented as normal `Lights` (G5), (4) the capture
`RedPalette()` flash (`Source/capture.cpp:97`), (5) per-entity TRNs (infravision/stone/pause). No shader passes exist.

**C. QoL floating sprites.** `DrawFloatingNumbers` (`Source/qol/floatingnumbers.cpp:84`) is **screen-space** text:
`worldOffset = viewPosition->tileToScreen + cameraOffset + {TILE_WIDTH/2, -TILE_HEIGHT/2} + startOffset`,
×2 if zoom, animated along `endOffset * mul` where `mul = 1 - timeLeft/2500.0F` (float! it's a cosmetic animation
driven by `SDL_GetTicks`, not by ticks — record this as an intentional non-determinism in the C++ source; our parity
screenshots must disable `showFloatingNumbers`). Item labels and monster health bars likewise use `DrawString` /
`DrawHalfTransparentRectTo`. XP bar and flasks are rect fills + CLX art.

**D. `capture.cpp` as the parity oracle.** Flow (`Source/capture.cpp:84`): `DrawAndBlit()` → (red palette flash,
restored) → `WriteSurfaceToFilePcx(GlobalBackBuffer(), stream)` or PNG → `RedrawEverything()` + a ≤300 ms delay.
`GlobalBackBuffer()` is the **full 640×480 index buffer including panels and cursor** — ideal for 1:1 comparison.
The PCX writer (`Source/utils/surface_to_pcx.cpp:45`) emits a standard 128-byte PCX header, 8 bpp, one plane,
RLE per row (`:92`), then separator `0x0C` + 768-byte palette (`:70`) — so a screenshot is a *paletted* image and can
be compared index-wise, not just colour-wise. `DEVILUTIONX_SCREENSHOT_FORMAT` defaults to PNG in this fork
(`Source/capture.cpp:26-28` choose by build flag).

### Ebitengine mapping
- **Blending cannot be done with `DrawImageOptions` alpha**: `lut` is an arbitrary nearest-colour map. Options:
  1. *Recommended*: a Kage "blend" fragment shader variant with two extra textures — the 256×256 LUT uploaded as a
     256×256 `R8` texture (`lutTex[a][b]`), and the destination read via `imageDst` sampling
     (`func Fragment(...) vec4 { src = ...; dst = imageSrc2At(dstPos.xy); return lut(src,dst) }`). Ebitengine supports
     reading the destination in a shader through a bound copy of the target — but that requires the target to be a
     *separate* image (see `examples/masking`, `examples/blend` in
     https://github.com/hajimehoshi/ebiten/tree/v2.9.9/examples and the shader doc
     https://ebitengine.org/en/documents/shader.html). Practical shape: draw blended tiles/sprites into
     `blendLayer` (640×480, index+alpha), then composite `blendLayer` onto `world` with a shader that samples
     `world`'s palette-index texture and the LUT texture. That needs an **index read-back of the world**, which the GPU
     cannot do cheaply from RGBA8 unless we keep the world target in a format whose R channel still carries the index.
  2. *Chosen*: **keep the world target as an index target.** Render the whole world into
     `worldIdx *ebiten.Image` in `ImageFilter`-friendly `R8`-equivalent (RGBA8 with index in R and 255 alpha), i.e. the
     "opaque copy" mode of the blend shader writes `vec4(idx,0,0,1)`, and the "blend" mode reads the destination's R as
     the background index. Then a final pass converts `worldIdx` → RGBA via the palette texture. This is *exactly* the
     C++ architecture (index framebuffer + palette at present time), and it makes blends, light tables and TRNs all
     byte-exact with **one** conversion at the end.
     Consequences: (a) the blend LUT becomes a 256×256 texture; (b) cycling = rebuild one 256×256 texture + 16 palette
     rows (cheap); (c) the UI layer can stay in normal RGBA (it never blends through the LUT except the minimap
     background and monster-health bars, which we route into `worldIdx` too).
  3. Fallback if the two-pass read-back measures badly: bake `blend` per (src,dst) into **two** index passes with a
     CPU-side `[]uint8` framebuffer mirror (640×480 = 300 KB, memcpy per frame) — the "software renderer in Go" option.
     It guarantees parity and is fast enough at 640×480 (the C++ is software at this resolution). Keep as plan B,
     documented in `dg-task-011`.
- Palette conversion pass = 1 `DrawRectShader` over the whole screen (one command).
- Floating numbers / labels / XP bar: ordinary `DrawImage`/`DrawRectShader` text & rect fills; **disabled in parity
  mode** (options), and their C++ float-driven motion is explicitly *not* deterministic — record as an accepted
  difference.
- `hwcursor.cpp` → **not ported** (OS cursor nicety, no framebuffer effect).
- Screenshots: `screen.ReadPixels` → PNG. Since we keep an index buffer (option 2), also dump the *index* buffer so
  goldens can be compared index-wise (stronger than colour comparison, immune to palette diffs).

### Parity test
- `TestBlendingLut`: port `GenerateBlendedLookupTable` incl. `skipFrom/skipTo`, build for the 4 shipped palettes
  (`assets/levels/l1data/l1_1.pal`-equivalents from MPQ, `levels/towndata/town.pal`, `nlevels/l5data/l5base.pal`),
  compare the full 65 536 bytes vs C++ dump. The KD-tree nearest search must break ties identically
  (`Source/utils/palette_kd_tree.hpp`) — this is the highest-risk byte-exactness item in G8.
- `TestBlendedArch`: single wall tile with `TransList` set + `dLight=6` → index buffer equals C++ dump.
- `TestScreenshotRoundtrip`: add a `WriteSurfaceToIndexDump()` to the C++ test build; drive it with
  `devilutionx --demo N --timedemo` + a scripted `CaptureScreen()`; Go compares
  `internal/render/golden/<frame>.idx.png` (`cmp -s`-able).
- Reuse: `test/palette_blending_test.cpp` + `test/palette_blending_benchmark.cpp` already exist → port their cases.

### Open questions
- Is `DEVILUTIONX_PALETTE_TRANSPARENCY_BLACK_16_LUT` (a 64 K-entry `uint16` fast path with endianness-dependent
  packing, `Source/utils/palette_blending.cpp:22`) wanted? It is a pure optimisation: **not ported**.
- Which screenshot format for goldens — PNG (RGBA) or raw index dump? Recommendation: both (index for CI asserts,
  PNG for human review).

---

## G9. ASSET MANIFEST (graphics)

Verified paths: `smpq -l` listings of `DIABDAT.MPQ` / `devilutionx.mpq` on this machine (2 910 entries in DIABDAT;
`devilutionx.mpq` contains the `.clx`/`.trn`/`.png`/`txtdata` set), the repo's `assets/` tree (100 `.clx`, 20 `.trn`,
2 `.pal`, 76 font `.clx` files; `find assets -iname '*.cel' -o -iname '*.cl2'` → 0), and the loader call sites cited
below. Sizes/counts marked *(measured)* come from files actually opened.

### Sources (file:line)
- Source/diablo.cpp:1374 — `nlevels\towndata\town.cel` (Hellfire) else `levels\towndata\town.cel` (`:1376`)
- Source/diablo.cpp:1380/1382 — `town.til` (same dual lookup)
- Source/diablo.cpp:1386 — `levels\towndata\towns` special (trees), width 64
- Source/diablo.cpp:1391 — `levels\l1data\l1.cel`
- Source/diablo.cpp:1392/1395/1398 — `l1.til`, `levels\l1data\l1s`
- Source/diablo.cpp:1400 — `levels\l2data\l2.cel`, `:1402` `l2.til`, `:1403` `l2s`
- Source/diablo.cpp:1405 — `levels\l3data\l3.cel/.til` + `levels\l1data\l1s` (re-used!)
- Source/diablo.cpp:1409 — `levels\l4data\l4.cel/.til` + `levels\l2data\l2s`
- Source/diablo.cpp:1413 — `nlevels\l6data\l6.cel/.til` + `l1s` (Nest)
- Source/diablo.cpp:1417 — `nlevels\l5data\l5.cel/.til/.s` (Crypt)
- Source/levels/dun_tile_data.cpp:68 — `<lvl>.min` per level type (`nlevels\towndata\town.min` `:71`,
  `levels\l1data\l1.min` `:76`, l2 `:78`, l3 `:80`, l4 `:82`, `nlevels\l6data\l6.min` `:84`, `nlevels\l5data\l5.min` `:86`)
- Source/levels/dun_tile_data.cpp:104 — `<lvl>.sol` (+ town dual-path `:107`/`:109`, crypt `:159`)
- Source/levels/dun_tile_data.cpp:143/151/159 — hand-patched SOL entries (must be transcribed!)
- Source/engine/palette.cpp:191 — `levels\towndata\town.pal`
- Source/engine/palette.cpp:196 — `nlevels\l5data\l5base.pal`
- Source/engine/palette.cpp:206 — `nlevels\l6data\l6base<rv>.pal` (Nest, `rv` 1..5)
- Source/engine/palette.cpp:213 — `levels\l<dl>data\l<dl>_<rv>.pal` (random 1..4 per level)
- Source/levels/gendung.cpp:417 — `.dun` load path (`LoadFileInMem<uint16_t>`), `:438`
- Source/automap.cpp:1528 — `levels\towndata\automap.amp` (`AutomapTile` array; repo ships
  `assets/levels/towndata/automap.amp` + `assets/levels/towndata/automap.dun`); `:1293`/`:1354`/`:1775` scale choice
- Source/levels/setmaps.cpp:112 — set-map `.dun` names (`levels\l1data\sklkngt.dun` etc.; repo ships
  `assets/levels/l1data/sklkngt.dun`, `assets/levels/l2data/bonechat.dun`, `assets/arena/{church,circle_of_death,hell}.dun`,
  `test/fixtures/levels/l1data/{banner1,banner2,rnd6,skngdo}.dun`)
- Source/player.cpp:1530 — `plrgfx\<path>\<caw>\<caw><anim>` (+ `DEVILUTIONX_CL2_EXT`)
- Source/player.cpp:2103 — anim suffixes: `as aw at ht bl lm fm qm dt` (+ town `st wl`)
- Source/player.cpp:2128/2163 — `LoadCl2Sheet`
- Source/monster.cpp:3217 — `monsters\<spritePath><n|w|a|h|d|s>[j]` + `DEVILUTIONX_CL2_EXT`
- Source/monster.cpp:175 — `monsters\<trnFile>.trn`
- Source/monster.cpp:3318 — `monsters\monsters\<mTrnName>.trn`
- Source/items.cpp:2428 — `items\<ItemDropNames[i]>` + ext (35 in Diablo, `ITEMTYPES`=43 in Hellfire, `:2426`)
- Source/items.h:184 — `ItemAnimWidth = 96`
- Source/objects.cpp:3700 — `objects\<ObjMasterLoadList[i]>` with per-level `animWidth` from objdat
- Source/towners.cpp:110 — `towners\<name>\<anim>` + `_tAnimWidth`; `:780` `towners\animals\cow` (128)
- Source/missiles.cpp:103 — `_miAnimWidth2` from missile data; missile sprites via `MissileSpriteData` (`assets/txtdata/missiles/missile_sprites.tsv`)
- Source/cursor.cpp:445 — `data\inv\objcurs.clx` / `:448` `data\inv\objcurs` + `data\inv\objcurs-widths.txt`,
  `:446/:451` `objcurs2(.clx)`/`objcurs2-widths.txt`
- Source/inv.cpp:1181 — `data\inv\<invName>` (`pInvCels`), width = `SidePanelSize.width`
- Source/engine/render/text_render.cpp:171 — `fonts\<size>-<row>.clx` (sizes 12,24,30,42,46,22)
- Source/engine/render/text_render.cpp:176 — `fonts\<lang>\<size>-<row>.clx` (zh/ja/ko/tr overrides; repo ships `assets/fonts/tr/`)
- Source/engine/render/text_render.cpp:88 — `fonts\<name>.trn` ×16
- Source/engine/render/text_render.cpp:225 — legacy PCX fonts (`fonts\<size>-<row>.pcx`, 256 frames, transparent=1)
- Source/lighting.cpp:216 — `plrgfx\infra.trn`, `:217` `plrgfx\stone.trn`, `:218` `gendata\pause.trn`
- Source/engine/trn.cpp:33 — `plrgfx\<class>.trn`
- Source/engine/trn.cpp:50 — `<plrgfx path with .trn>` per-graphic
- Source/DiabloUI/button.cpp:34 — `ui_art\dvl_but_sml.clx` (`:36` legacy `but_sml.pcx`)
- Source/DiabloUI/credits.cpp:189 — `ui_art\creditsw.clx` (+ `LoadBackgroundArt("ui_art\\credits")`); `:197`
  `ui_art\supportw.clx`
- Source/DiabloUI/diabloui.cpp:629 — `ui_art\heros.pcx` (sliced by −PortraitHeight)
- Source/DiabloUI/diabloui.cpp:647 — `ui_art\hero<i>.pcx`
- Source/DiabloUI/diabloui.cpp:654 — `ui_art\hf_logo2.pcx` (16 frames, transparent 0)
- Source/panels/mainpanel.cpp:100 — `data\talkbutton.clx`
- Source/cursor.cpp:492 — half-size item sprite generation (bilinear downscale) — `data\inv\objcurs*` derived
- Source/engine/assets.cpp:410 — `LoadMPQ(paths, mpqName, priority)`; `:542` `diabdat` (MainMpqPriority),
  `:544` `spawn`, `:573` `hellfire` (8000); `:510` devilutionx.mpq first for error fonts
- Source/engine/assets.hpp:309 — priority constants (devilutionx 9000 / lang 9100 / fonts 9200 / main 1000 / hellfire 8000)
- Source/engine/load_cel.hpp:12/14, Source/engine/load_cl2.hpp:23/27, Source/engine/load_pcx.hpp:15/17 —
  extension switching: **bundled assets are `.clx`; retail MPQ reads `.cel`/`.cl2`/`.pcx`**
- assets/gendata/cut2w.clx … cutttw.clx — in-game cutscene stills (9 files, `ls assets/gendata`)
- assets/data/*.clx — QoL/UI: `boxleftend boxmiddle boxrightend charbg dirtybuc dirtybucp health healthbox hintbox
  hintboxbackground hinticons monstertags panel8buc panel8bucp repairAllBtn repairSingleBtn resistance stash
  stashnavbtns store tabBtnUp talkbutton` (measured: 24 files in `assets/data`, 24 more in `assets/data/inv`)

### Formats & algorithms (inventory table)

| family | archive path (as loaded) | format | frame layout / dims | loader |
|---|---|---|---|---|
| level tiles L1 | `levels/l1data/l1.cel` | CEL (raw, BE-swapped offset table) | 1119 frames (measured), each one micro-tile: 32×32 flat, 544→512 B triangle, 800→784 B trapezoid, or 32-px-row int8 RLE | `Source/diablo.cpp:1391`, `reencode_dun_cels.cpp:211` |
| level tiles L1 town/L2/L3/L4/L5(crypt)/L6(nest) | `levels/l2data/l2.cel`, `…/l3.cel`, `…/l4.cel`, `nlevels/l5data/l5.cel`, `nlevels/l6data/l6.cel`, `levels/towndata/town.cel` (+`nlevels/towndata/town.cel`) | CEL | same 6 tile types | `Source/diablo.cpp:1374-1418` |
| megatile table | `levels/l1data/l1.til` (…all levels) | raw `MegaTile[]`, 8 B each, u16 BE | L1 = 206 megatiles (measured 1648 B) | `Source/diablo.cpp:1392`, `gendung.cpp:578` |
| micro index table | `levels/l1data/l1.min` | u16 array, `LevelCelBlock` (type‖frame) | L1 = 4530 u16 = 453 pieces × 10 (measured); town = 20 128 u16 = 1258 × 16 | `Source/levels/dun_tile_data.cpp:68` |
| tile properties | `levels/l1data/l1.sol` | 1 byte/piece `TileProperties` | L1 = 453 B, town = 1258 B (measured) | `Source/levels/dun_tile_data.cpp:104` |
| arches / trees ("special") | `levels/l1data/l1s`, `l2s`, `towns`, `nlevels/l5data/l5s` | CEL, **width 64** | L1: 8 frames (measured, 21 794 B) | `Source/diablo.cpp:1365`, `Source/engine/render/scrollrt.cpp:926` |
| level palette | `levels/l1data/l1_1.pal`…`l4_4.pal`, `levels/towndata/town.pal`, `nlevels/l5data/l5base.pal`, `nlevels/l6data/l6base1..5.pal` | 768 B = 256 × RGB | random 1..4 per level (Nest 1..5) | `Source/engine/palette.cpp:196-213` |
| level layout | `levels/l1data/*.dun` (setmaps, themes), `levels/towndata/town.dun`, `test/fixtures/diablo/<lvl>-<seed>.dun` | u16 LE ×N: `w,h` then tiles (+ transparency layer) | 2×2-scaled; town 26×38 (measured `sklkngt.dun` 0x25,0x19 = 37×25) | `Source/levels/gendung.cpp:417,438` |
| automap town | `levels/towndata/automap.amp` | `AutomapTile[]` (type+flags) | 40×40 | `Source/automap.cpp:1528` |
| players | `plrgfx/<warrior|rogue|sorcerer|monk|barbarian|bard>/<c><a><w>/<c><a><w><anim>.cl2` | CL2 **sheet of 8 direction lists** | widths per class/graphic/weapon from `txtdata/classes/<cls>.tsv`; anims `as aw at ht bl lm fm qm dt` (town `st wl`) | `Source/player.cpp:1530,2163` |
| player TRN | `plrgfx/<class>.trn`, `plrgfx/<class>/<xyz>/<xyz>{fm,lm,qm}.trn` | 256 B | — | `Source/engine/trn.cpp:29,47` |
| monsters | `monsters/<spritePath><n|w|a|h|d|s>[j].cl2` (e.g. `monsters/zombie/zombie.cl2` family) | CL2 list **or** sheet(8) | width from `monstdat.tsv` col `width` (128 for zombies, measured row) ; frames/rate cols `frames[6]`,`rate[6]` (e.g. `11,24,12,6,16,0` / `4,1,1,1,1,1`) | `Source/monster.cpp:3217,3549` |
| monster TRN | `monsters/<trnFile>.trn`, uniques `monsters/monsters/<name>.trn` | 256 B (255→0 quirk) | one per monster type/unique | `Source/monster.cpp:175,3318` |
| items (ground) | `items/<name>` e.g. `armor2 axe fbttle bow goldflip helmut mace shield swrdflip rock cleaver …` | CEL, width 96 | 1 frame per animation; `ITEMTYPES` 43 (Hellfire) / 35 (Diablo) | `Source/items.cpp:2424`, `Source/items.h:184,26` |
| objects | `objects/<name>` per `objdat` `ofindex` | CEL, per-object `animWidth` | ≤ 40 files loaded per level (`pObjCels[40]`) | `Source/objects.cpp:3700,113` |
| towners | `towners/<name>/<anim>` | CEL, `_tAnimWidth` | per-NPC anim order from `txtdata/towners/towners.tsv` | `Source/towners.cpp:110` |
| cow | `towners/animals/cow` | CEL 128 | — | `Source/towners.cpp:780` |
| missiles | per `txtdata/missiles/missile_sprites.tsv` (`MissileSpriteData`) | CL2/CEL | `animWidth2` precomputed | `Source/missiles.cpp:103` |
| cursor/item icons | `data/inv/objcurs(.clx|.cel)` + `data/inv/objcurs-widths.txt`, `objcurs2` (+`-widths.txt`) | CLX list (retail: CEL + widths txt) | 179 + N frames, widths 28/32/33/56 (measured from the widths file & header) | `Source/cursor.cpp:445-452` |
| inventory panel art | `data/inv/<invName>` | CEL, width = SidePanelSize.width | `pInvCels` | `Source/inv.cpp:1181` |
| QoL/UI small art | `assets/data/*.clx` (24 files: `charbg`, `health`, `healthbox`, `hintbox*`, `monstertags`, `resistance`, `stash*`, `store`, `tabBtnUp`, `talkbutton`, `xpbar`, `panel8buc*`, `dirtybuc*`, `box*`, `repair*`, `hinticons`) | CLX | e.g. `health.clx` 264×19 single frame; `dirtybuc.clx` 2 frames 63×12 (measured) | `LoadClx`/`LoadOptionalClx` sites in `Source/qol/*`, `Source/panels/*` |
| fonts | `fonts/<size>-<row>.clx` (76 files in repo), `fonts/<lang>/<size>-<row>.clx` (`assets/fonts/tr/`), legacy `.pcx` | CLX list of 256 glyphs (zero-width = w0,h<line>) | sizes 12/24/30/42/46/22 | `Source/engine/render/text_render.cpp:171-225` |
| font colour TRN | `fonts/{goldui,grayui,golduis,grayuis,yellow,black,white,whitegold,red,blue,orange,buttonface,buttonpushed,gamedialogwhite,gamedialogyellow,gamedialogred}.trn` | 256 B | 16 active of 19 slots | `Source/engine/render/text_render.cpp:88` |
| misc TRN | `plrgfx/infra.trn`, `plrgfx/stone.trn`, `gendata/pause.trn` | 256 B | — | `Source/lighting.cpp:216` |
| UI screens | `ui_art/*.pcx` (retail) and `ui_art/{creditsw,mainmenuw,supportw,dvl_but_sml,dvl_lrpopup,credits…}.clx` (devilutionx.mpq) | PCX/CLX | `heros.pcx` sliced by portrait height; `hf_logo2.pcx` 16 frames | `Source/DiabloUI/diabloui.cpp:629-654` |
| cutscene stills | `gendata/cut{2,3,4,gate,l1d,portl,portr,start,tt}w.clx` | CLX | full-screen (repo ships 9) | `Source/movie.cpp`, `Source/panels/*` |
| control panel | `ui_art/paperdoll.pcx`, `data/mayo.pcx` (inv), `ui_art/cell.pcx` … via `LoadPcx`/`LoadCel` in `Source/panels/`, `Source/control/` | PCX/CEL | panel 640×128, side panels `SidePanelSize` | `Source/control/control_panel.cpp:243` (`MainPanelSize {640,128}`) |
| sprites in MPQ order | `devilutionx.mpq` (prio 9000) → lang mpq (9100) → `fonts.mpq` (9200) → `DIABDAT/spawn` (1000) → `hellfire.mpq` (8000) → `hfmonk` (8100) → `hfbard` (8110) → `hfbarb` (8120) → `hfmusic` (8200) → `hfvoice` (8500) → mods (≥10000) | mixed | same names, `.clx` preferred when unpacked | `Source/engine/assets.cpp:410`, `assets.hpp:309` |

**Extension rule (important, easy to get wrong):** the *same* logical asset is read as
`.clx` when `UNPACKED_MPQS` is on (bundled assets are pre-converted) and as `.cel` / `.cl2` / `.pcx` for retail MPQs
(`Source/engine/load_cel.hpp:12`, `Source/engine/load_cl2.hpp:23`, `Source/engine/load_pcx.hpp:15`). The Go port must
implement **both** paths (or a single `.clx`-only path plus a pre-conversion tool) and must apply widths in the same
places, since CL2→CLX needs the width to derive height (G3).

### Ebitengine mapping
- One `assets/` Go package exposing `func Graphics() map[Family]Blob`, plus a converter binary
  (`cmd/dgo-convert`) that runs CEL/CL2/PCX → CLX once (matching devilutionX's own `assets/` build, see
  `tools/update_bundled_assets.sh`) so the runtime path is CLX-only.
- Atlas plan: pages of 4096×4096 (see G10) split by family so that a batch never crosses families:
  `dungeon/<level>`, `players/<class>`, `monsters`, `items`, `objects`, `missiles`, `cursors`, `ui`, `fonts/<size>`.
- Nothing in the draw path may decode RLE (decode at load).

### Parity test
`TestAssetManifest`: assert the Go loader resolves every path in the table above for both the `.clx` (bundled) and
`.cel/.cl2/.pcx` (retail) variants, reporting the count of resolvable graphics files (target: 100 % of the graphics
subset of the extraction smoke test in dg-epic-02). `TestKnownDims`: table of (file, frame, w, h) taken from this
document — `boxleftend.clx#0 = 6×27`, `dirtybuc.clx#0 = 63×12`, `health.clx#0 = 264×19`, `30-00.clx#0 = 0×38`,
`objcurs.cel#0 = 33×33`, `l1s.cel` 8 frames — all *(measured)* here.

### Open questions
- Do we ship our own converted `.clx` bundle (as devilutionX does) or require retail MPQs? Recommendation: ship the
  same bundled set (repo already contains it) **and** support retail.
- The exact set of `ui_art` PCX files is spread over `Source/DiabloUI/*` and `Source/panels/*`; enumerate
  programmatically (`grep -rho 'ui_art[^"]*' Source | sort -u`) in the asset-manifest task rather than by hand here.
- `mods/` may add/replace graphics at runtime (priority ≥ 10000) — out of scope for parity.

---

## G10. EBITENGINE MAPPING DECISIONS (consolidated)

### Sources (file:line)
Local C++ anchors for the rules below: Source/engine/render/scrollrt.cpp:1834 (`DrawAndBlit` = the frame),
Source/engine/render/scrollrt.cpp:1229 (`DrawGame`), Source/engine/render/scrollrt.cpp:781 (`DrawDungeon`),
Source/engine/render/scrollrt.cpp:249 (`ClxDrawLight`), Source/engine/render/clx_render.cpp:556-581 (7 CLX draw
flavours), Source/engine/render/blit_impl.hpp:20-190 (the 8 blit ops), Source/engine/render/dun_render.cpp:1046
(`RenderTileFrame`), Source/engine/backbuffer_state.cpp:21 (dirty/cursor state, keyed to `PalSurface->pixels` — not ported),
Source/engine/dx.cpp:181/`:287` (BltFast/RenderPresent — presentation only, not ported),
Source/utils/display.cpp:723 (integer scaling), Source/options.h:41 (640×480).
Online, verified this session:
- https://ebitengine.org/en/documents/performancetips.html — batching rules ("All the functions are DrawImage or
  DrawTriangles / All the render targets are same / All the blends are same / All the filter values are same"),
  "The texture atlas size depends on graphic devices" (older docs said 4096×4096, mirrored copy at
  https://ebitengine.org/en/documents/performancetips.html / https://ebitengine-zh.pages.dev/documents/performancetips.html
  wording "The texture atlas size is 4096x4096 so far"), avoid mutating render sources, avoid cyclic draws, avoid the
  screen as a render source, `ReplacePixels`/`At` are heavy, count real commands with `-tags=ebitenginedebug`.
- https://ebitengine.org/en/documents/cheatsheet.html — `DrawImage` batching conditions, `Layout`,
  `SetTPS`, `NewImageFromImage`, `SubImage`, `ColorM`.
- https://github.com/hajimehoshi/ebiten/tree/v2.9.9/examples — full example list (86 entries) fetched; **there is no
  `examples/palette`**; relevant references: `shader`, `filter`, `blend`, `masking`, `minify`, `isometric`, `tiles`,
  `sprites`, `alphablending`, `text`, `cursor`, `capture`.
- https://github.com/hajimehoshi/ebiten/blob/main/cursor.go — `SetCursorShape` + `CursorShape*` (10 values),
  `SetCursorMode` (Visible/Hidden/Captured); example https://github.com/hajimehoshi/ebiten/blob/v2.9.2/examples/cursor/main.go.
- https://ebitengine.org/en/documents/shader.html — Kage language (pixel vs texel units, `imageSrc0At`, uniforms).
- https://ebitengine.org/en/documents/2.4.html — `NewImageWithOptions` `Unmanaged` / `PreserveBounds` (manual atlas).
- https://ebitengine.org/en/documents/2.5.html — `SetScreenClearedEveryFrame(false)`, `vector`/atlas notes.
- https://github.com/hajimehoshi/ebiten/issues/2581 — atlas packing behaviour with large images (why we keep pages ≤ 4096).
- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md, /home/blaze/.agents/skills/ebiten/resources/SKILL.md.

### Formats & algorithms (the decision list)

**D1 — Logical resolution 640×480 fixed** (`Source/options.h:41`). `Layout()` always returns (640, 480).
Scaling = `ebiten.SetFilter(ebiten.FilterNearest)` (pixel-perfect integer/letterbox), matching
`SDL_LOGICAL_PRESENTATION_INTEGER_SCALE` (`Source/utils/display.cpp:734`).

**D2 — Index-framebuffer architecture (the keystone).** Two images, both 640×480:
```
worldIdx *ebiten.Image  // RGBA8 but only the R channel is meaningful: the palette INDEX; A=255, G=B=0
screen  (via Draw)      // final RGBA, produced by one palette-lookup pass
```
Every world draw writes *indices* through one of three Kage shader variants:
| variant | op | byte semantics |
|---|---|---|
| `copy` | `dst.rgba = vec4(paletteIdx(src),0,0,1)` | `BlitDirect` (`Source/engine/render/blit_impl.hpp:28`) |
| `map` | `dst = vec4(tbl[src],0,0,1)` | `BlitWithMap` (`:57`) — covers all light/TRN cases |
| `blend` | `dst = vec4(lut[src][dst],0,0,1)` / `lut[dst][tbl[src]]` | `BlitBlended*` (`:113`, `:133`) |
`tbl` is *baked into the source texture choice* (variant index texture per light level) rather than a uniform, so that
`copy`/`map` both reduce to "sample R from source, sample palette row from source-specific lookup" — i.e. **one shader
with the table folded into a 256×1 texture chosen per batch**.
UI layer (`ui *ebiten.Image`) stays plain RGBA (it's `DrawString`, panels, PCX art) and is composited last.
The final pass: `screen.DrawRectShader(640, 480, paletteShader, {worldIdx, paletteTex})` — 1 command.
Why not "RGBA sprites + ColorScale"? Because `LightTables`, `InfravisionTable`, `StoneTable`, unique TRNs, class TRNs
and `paletteTransparencyLookup` are all **non-linear index remaps** (e.g. colours leaving their 16-entry ramp become
black). No float matrix can produce that. This is the "EXACT equivalent byte-op" requirement of the contract.

**D3 — Texture atlas layout.** Use Ebitengine's automatic internal atlas for sprites; constrain total atlas footprint:
- Target: each page ≤ 4096×4096 (device-independent assumption; older perf docs state 4096 explicitly, newer say
  "depends on graphic devices" — design for 4096 to stay safe on WebGL/older GLES).
- Budgets *(measured frame counts × average dims)*: dungeon micros (≈1119 frames/level × 32×32 ≈ 1.1 MB → < 1 page),
  players (per class ≈ 8 dirs × 9 anims × ~12 frames × ~100×110 ≈ 9.5 MB → ~1 page per class),
  monsters (~110 types × ~40 frames × ~110×120 ≈ 230 MB **if all resident** → must be paged per level population,
  not one global atlas), items (43 × ~20 × 96×32), cursors (179 frames, avg ~40×32), fonts (256 glyphs/size × 6 sizes),
  UI (`ui_art` full-screen PCX ≈ 640×480 each → these **cannot** share a 4096 page with many siblings; use
  `ebiten.NewImageWithOptions{Unmanaged:true}` for the few big backgrounds).
- Rule: **one atlas page per family per level-load**, built at load time with a single `ReplacePixels` per page
  (perf tips: `ReplacePixels` is heavy → never per frame), and `SubImage` at draw time (cheap, no allocation of new
  images: `SubImage` returns a view).
- 1-px gutters: Ebitengine's internal atlas already pads; because we use `FilterNearest` and integer translations of
  exact texel bounds, no bleeding. If we hand-roll an atlas (`Unmanaged`), add 1 px transparent gutters and only ever
  draw integer-aligned rects.

**D4 — Draw-call merging rules.** From the perf page, a run of `DrawImage` calls merges iff same target, same blend,
same filter (same address for `DrawTriangles`). Our ordering plan:
1. `worldIdx` pass: target fixed, filter = Nearest, blend = `BlendCopy` for tiles and opaque micros (the row-walk
   writes every covered pixel, and *later* tiles must overwrite earlier ones exactly like the C++ memcpy blitter —
   `BlendCopy`, not `SourceOver`!). Sprites need `BlendSourceOver` **plus** index-0 keying, i.e. the sprite shader
   returns `vec4(0,0,0,0)` for a hole; that keeps blend semantics correct for sprites *and* preserves the index-only
   encoding. ⇒ Group the frame as: [all tile micros with blend=Copy] → [all sprites with blend=SourceOver] → [all
   blended/arch things with the blend shader]. **This is a deviation from C++ interleaving order** (see D6 for how
   order is preserved inside each group).
2. Blended (transparent) arches and automap-transparent pixels must be drawn *between* those groups to match C++
   per-pixel results, which means an extra target switch — accept 3–5 extra draw commands per frame, not a problem.
3. `ui` pass: all `BlendSourceOver`, FilterNearest ⇒ typically 1–3 commands.
4. Final palette pass: 1 command. Screen: 1 command.
Total target: **< 20 GPU draw commands per frame** and ~4–7 k triangles. Verify with
`go run -tags=ebitenginedebug ./cmd/dgo` and count `draw-triangles:` lines (documented on the perf page).

**D5 — Allocation discipline.**
- `var drawOptsPool = sync.Pool{}` is unnecessary: `Draw` is single-goroutine. Use **pre-allocated structs**:
  ```go
  type renderer struct {
      opTiles  [MaxTilesPerFrame]ebiten.DrawImageOptions   // built once; GeoM mutated per draw
      opSprite ebiten.DrawImageOptions
      opUI     ebiten.DrawImageOptions
      ...
  }
  ```
  Mutating `op.GeoM` per draw is required (position), but *do not* set `ColorScale`/`Filter`/`Blend` per draw —
  they decide batching. Uniform-ish state (which palette variant) is encoded by *which image* you draw, not by options.
- `testing.AllocsPerRun` on the draw path must report 0 (dg-epic-03 acceptance).
- No `SubImage` per frame on hot paths? `SubImage` allocates a small struct + bounds; it is cheap but not free. Cache
  `map[frameID]*ebiten.Image` sub-images created once at atlas build. (Sub-image as a *render target* is fine in v2.1+;
  as a source it's the normal pattern — see `examples/tiles`.)
- Never call `At`, `ReadPixels`, `WritePixels`, `ReplacePixels` in `Draw` (perf page; also `At` forces a GPU flush).

**D6 — Update vs Draw split.**
```
Update() [20 TPS via a 60TPS accumulator, see dg-task-004]
   GameLogic() (Source/diablo.cpp:1523 order)  -> animation counters (AnimationInfo.processAnimation),
   dLight recompute (UpdateLighting latch), vision, missiles movement, items, monsters, objects,
   palette cycling (Source/diablo.cpp:3521/3526/3528), RedrawEverything()/RedrawViewport()/RedrawComponent flags.
   NO rendering state is read here.
Draw()  [every frame]
   progress := ProgressToNextGameTick        // 0..127  (nthread_UpdateProgressToNextGameTick, scrollrt.cpp:1866)
   // pure function of world + progress (determinism rule from Volume I §7)
   worldIdx.Fill(color.Transparent)                     // or skip if full-screen redraw covers everything
   for row in RowWalk(tileRows, tileColumns): DrawFloor(row)         // scrollrt.cpp:955
   for row in RowWalk: DrawTileContent(row)                          // scrollrt.cpp:992  (incl. DrawDungeon order)
   DrawOOB(row)                                                      // scrollrt.cpp:1088
   if zoom: upscale worldIdx 2x                                      // scrollrt.cpp:1121
   if automap: DrawAutomap                                           // automap.cpp:1752
   DrawItemNameLabels / DrawMonsterHealthBar / DrawFloatingNumbers    // scrollrt.cpp:1395-1398
   DrawPanels (MainPanel, flasks, belt, buttons, spell list, char, quest log, inv, stash, info box, gmenu, doom)
   DrawCursor                                                     // scrollrt.cpp:279 (bottom-left anchored)
   screen.DrawRectShader(palette pass, worldIdx -> RGBA)            // D2
   screen.DrawImage(ui, opUI)                                       // RGBA layer
```
Concrete per-tile pseudocode (must mirror `Source/engine/render/scrollrt.cpp:781-952`):
```go
func (r *Renderer) drawDungeon(tile, bufPt Point, light uint8) {
    r.drawCell(tile, bufPt, light)                                     // scrollrt.cpp:549 (micros, masks, foliage)
    if r.missilePreFlag { r.drawMissiles(tile, bufPt, true, light) }
    if c := r.corpseAt(tile); c.ok() && light < LightsMax {            // scrollrt.cpp:805
        if c.trn != nil { r.drawTRN(buf, sprite, c.trn) } else { r.drawLight(buf, sprite, light) }
    }
    if o := r.preObjectAt(tile); o != nil { r.drawObject(*o, ...) }    // scrollrt.cpp:816
    if it := r.itemAt(tile); it != nil && !it.postDraw { r.drawItem(...) }   // :818
    r.drawDeadPlayers(tile, buf, light)                                // :825
    if p := r.playerAt(tile); p != nil && r.dPlayer[tile] == p.id() { r.drawPlayer(*p, ...) }   // :828-866
    if m := r.monsterAt(tile); m != nil && r.dMonster[tile] == m.id() { r.drawMonster(*m, ...) } // :868-899
    if it := r.itemAt(tile); it != nil && it.postDraw { r.drawItem(...) }      // :901
    if o := r.postObjectAt(tile); o != nil { r.drawObject(*o, ...) }          // :903
    r.drawMissiles(tile, buf, false, light)                              // :905
    if arch := r.dSpecial[tile]; arch > 0 {                              // :908-933
        if r.transList[r.dTransVal[tile]] { r.drawLightBlended(archSprite, light) } else { r.drawLight(archSprite, light) }
    }
}
```
(And the town branch, `Source/engine/render/scrollrt.cpp:944-952`, defers tree art one row: read
`dSpecial[x-1][y-1]` and draw at `buf + {0,-TILE_HEIGHT}`.)

**D7 — Things explicitly not ported** (recorded so reviewers don't re-litigate):
`Source/hwcursor.cpp` (OS cursor), dirty-rect presentation (`Source/engine/backbuffer_state.cpp`, `BltFast`/`DoBlitScreen`), SDL renderer
letterboxing internals, `SDL_BilinearScale`/`scaleQuality` (Ebitengine filter choice only),
`DEVILUTIONX_PALETTE_TRANSPARENCY_BLACK_16_LUT`, per-pixel `Lightmap`/`bleedUp` (M3+ optional), SMK movies
(`Source/movie.cpp`, `Source/storm/storm_svid.cpp`) — stubbed by `CutscenePlayer` per Volume I §7,
`ScrollText`/`DrawDoom` movie frames (still images OK), touch/virtual-gamepad overlays
(`controls/touch/renderers.cpp`), Lua console/debug overlays.

### Parity test
- `go build -tags=ebitenginedebug && ./dgo 2>&1 | grep -c draw-triangles` < 40 per frame (budget D4).
- `testing.AllocsPerRun(1000, game.Draw)` == 0.
- `TestLayoutFixed`: `Layout(1920,1080) == (640,480)`.
- Benchmark on the CI runner: full-frame draw < 3 ms (mirrors dg-epic-03's acceptance criterion and ports
  `test/dun_render_benchmark.cpp` / `test/clx_render_benchmark.cpp` / `test/light_render_benchmark.cpp` case sets).

### Open questions
- Two-pass index framebuffer (D2) costs one extra 640×480 RGBA buffer + a full-screen shader. Acceptable?
  (Estimated 1.2 MB VRAM; yes.) Alternative: skip `worldIdx` and accept per-(sprite,light) pre-baked RGBA for
  monsters/players only + tiles drawn as index — hybrid, more memory, less clean. Decide in dg-task-010.
- Does `BlendCopy` into an offscreen then `DrawImage` of that offscreen to screen break batching? Documented:
  "An offscreen doesn't share the texture atlas with high probability" — expected; that's exactly why we keep
  ≤ 5 offscreen targets total.
- `ebiten.SetTPS`: keep 60 TPS + internal 20 TPS accumulator (dg-task-004) or set TPS to 20 and let Draw interpolate?
  Recommendation: 60 TPS accumulator (already the epic-01 decision); revisit only if jitter appears.

---

## G11. PIXEL-PARITY TEST PLAN

### Sources (file:line)
- test/fixtures/text_render_integration_test/ — committed golden PNGs (verified: `basic.png`, `basic-colors.png`,
  `cursor-start/middle/end.png`, `highlight-full/partial.png`, `horizontal_overflow(-colors).png`,
  `kerning_fit_spacing__align_center(-colors).png`, …)
- test/text_render_integration_test.cpp — the harness that produces them
- test/fixtures/light_render_benchmark/dLight.dmp — committed `dLight` dump (reusable as a lighting fixture)
- test/fixtures/diablo/*.dun (seeded level dumps, e.g. `1-2588.dun` 54 404 B), test/fixtures/hellfire/*.dun,
  test/fixtures/levels/l{1..4}data/*.dun, test/fixtures/txtdata/*.tsv edge cases, test/fixtures/timedemo/
- test/scrollrt_test.cpp:13/:80 — viewport geometry goldens
- test/dun_render_benchmark.cpp, test/clx_render_benchmark.cpp, test/light_render_benchmark.cpp — the C++ draw
  harnesses to extend with dumps
- test/timedemo_test.cpp:28 — `RunTimedemo` (needs `spawn.mpq`/`diabdat.mpq`, skips otherwise; `AdjustToScreenGeometry`
  at `:89`; compares hero saves vs `demo_0_reference_spawn_0.sv`)
- test/palette_blending_test.cpp, test/animationinfo_test.cpp, test/automap_test.cpp, test/cursor_test.cpp,
  test/inv_test.cpp, test/inventory_ui_test.cpp, test/char_panel_test.cpp, test/spell_ui_test.cpp,
  test/panel_state_test.cpp, test/dead_test.cpp, test/tile_properties_test.cpp
- Source/capture.cpp:84 — `CaptureScreen()` (oracle), Source/utils/surface_to_pcx.cpp:45 — PCX writer,
  Source/utils/surface_to_png.hpp — PNG writer
- Source/diablo.cpp:1033 — `--timedemo`, `:1031-1032` `--record/--demo`, `:1127` parse, `:1204` `demo::InitPlayBack`
- Source/headless_mode.hpp / `HeadlessMode` (rendering short-circuits: `Source/engine/palette.cpp:177`,
  `Source/engine/dx.cpp:203`, `Source/engine/render/scrollrt.cpp:1816`)
- tools/linux_reduced_cpu_variance_run.sh, tools/measure_timedemo_performance.py, tools/build_and_run_benchmark.py
- /tmp/dg-requirements.md §7 (Rendering) and Appendix H — cross-checked; §7's "offscreen world buffer + UI layer,
  atlas + shared `DrawImageOptions`, draw order mirrors `DrawDungeon`, negative-id trick" is adopted here verbatim;
  §7 lists the open question "fixed 640×480 vs dynamic Layout" — **this spec answers: fixed 640×480** (G1, G10 D1).
  §7 also lists the open question "transparency blending must be replicated with ColorScale alpha ≈ composite" —
  **this spec overrides: ColorScale cannot do it; index framebuffer + 256×256 LUT (G8 D2).** That is the only place
  where this document deliberately contradicts Volume I §7, plus the lighting line "pre-render 16 shaded variants per
  sprite is memory-heavy → choose ColorScale": replaced by 16 palette rows (same memory conclusion, different
  mechanism, stronger parity guarantee).
- /tmp/dg-bean-tree.md — `dg-epic-02`/`dg-task-007` (decoders), `dg-epic-03` + `dg-task-010..014` (renderer),
  `dg-epic-11` (QoL/automap extras), `dg-task-005` (golden harness scaffolding) — all referenced ids exist.
- Volume I open question **H14** ("no GPU variance rule") — see P3 below.

### Formats & algorithms (test design)

**P1 — Format decoder goldens (no GPU needed; runs everywhere).**
1. Extend the C++ test target with a dumper (a *new* gtest in the devilutionX tree, run once by a human, outputs
   committed to `devilutionGo/testdata/graphics/`):
   ```
   devilutionX:  gtest "GraphicsDump.*" writes, for each (file, frame):
                   <name>.dims.txt          -> "w h\n"
                   <name>.index.bin         -> w*h bytes, row-major TOP-to-BOTTOM (flip C++'s bottom-up order)
                   <name>.clxtxt            -> ClxDescribe(frame) (Source/engine/render/clx_render.cpp:600, DEBUG_CLX build)
   ```
   Fixture set (all paths *(verified)* present): `assets/data/boxleftend.clx` (6×27),
   `assets/data/boxrightend.clx` (6×27), `assets/data/dirtybuc.clx` (2×63×12), `assets/data/health.clx` (264×19),
   `assets/fonts/30-00.clx` (256 frames incl. 0×38 empties), `assets/fonts/12-00.clx`,
   `assets/gendata/cut2w.clx`, `assets/ui_art/creditsw.clx`, `assets/ui_art/dvl_but_sml.clx`,
   retail-only: `data/inv/objcurs.cel` (+widths), `levels/l1data/l1.cel` frames {1,2,48}, `levels/l1data/l1s.cel`
   (8 frames, width 64), `plrgfx/warrior/hu/huas.cl2` (sheet, 8 lists), `monsters/zombie/zombie*.cl2`,
   `ui_art/paperdoll.pcx`, `ui_art/heros.pcx`.
2. Go side: `go test ./internal/assets/... -run TestClxGolden` compares `[]uint8` byte-for-byte; plus a hash-based
   test over all frames of a file (`sha256` of the concatenated index stream) so a single line catches regressions:
   ```
   cd devilutionGo && go test ./internal/assets/clx -run Golden -count=1
   ```
3. TRN/light-table goldens: dump `LightTables[16][256]` for each `leveltype`, `paletteTransparencyLookup[256][256]`
   for each shipped palette, `LightFalloffs[16][128]`, `LightConeInterpolations[8][8][16][16]`, and each `.trn`
   (256 B) → byte-compare.

**P2 — Full-frame oracle (pixel-exact).**
1. *C++ side*: run the real binary in a deterministic mode and dump the backbuffer at known frames.
   ```
   # 1) record once (needs retail/spawn MPQ in the data dir):
   ./build/devilutionx --data-dir <dir> --record 0 --start-intro-skipper   # writes multi_0.sv / demo_0.sv
   # 2) replay with a fixed logical resolution, off-screen SDL (no window):
   SDL_VIDEODRIVER=dummy ./build/devilutionx --data-dir <dir> --demo 0 --timedemo &
   ```
   `--timedemo` is defined as "Disable all frame limiting during demo playback" (`Source/diablo.cpp:1033`) and demo
   playback is parsed at `Source/diablo.cpp:1127-1143`, `demo::InitPlayBack` at `:1204`.
   **Problem found and recorded:** with `HeadlessMode` or `demo::IsRunning()`, fades short-circuit
   (`Source/engine/palette.cpp:277`, `:330`) and `scrollrt_draw_game_screen`/`DrawAndBlit` early-return
   (`Source/engine/render/scrollrt.cpp:1816`, `:1836`), so a *dump hook* is required rather than a screenshot:
   add to the C++ test build a
   `DumpBackbuffer(const Surface&, const uint8_t palette[768])` writing PPM + the index buffer,
   called from `DrawAndBlit` every N frames. The `SDL_VIDEODRIVER=dummy` path still creates `PalSurface`, so the
   indexed buffer exists; that is what we dump. (Existing precedent: `test/timedemo_test.cpp` drives the same code
   path with `SDL_Init(SDL_INIT_EVENTS)` only.)
   Alternative that avoids new C++ code: `CaptureScreen()` (`Source/capture.cpp:84`) writes PCX/PNG **including the
   palette**, triggered by a key binding; script it with `--demo` + injected key events. Slower, but zero new C++.
2. *Go side*: the same demo tape + same seed + same level → render one frame with `progress ∈ {0, 64, 127}` pinned
   (drive `ProgressToNextGameTick` from the test, not from the clock), then compare.
3. *Comparison*: two levels of strictness:
   - `index-exact`: Go `worldIdx` buffer vs C++ `PalSurface->pixels` → must be **byte-identical**. This is the real
     1:1 gate; it is palette-independent, so palette-stage differences can't hide bugs.
   - `colour-exact`: decoded RGBA vs C++ PCX-with-palette → identical modulo the *allowed differences list* below.
4. **Allowed palette-stage differences** (documented, each justified):
   a. Brightness tone-map: run both with `brightness=0` (`MapTone(a=0)` = identity, `Source/engine/palette.cpp:96`).
   b. Colour cycling: golden frames must be captured at the same tick count; better — disable cycling in both
      (no `palette_update_*` call), since the cycle position is time/palette state, not pixel state
      (`Source/diablo.cpp:3521`).
   c. Fades in progress: only compare frames where `sgbFadedIn` is settled (`Source/engine/palette.cpp:271`).
   d. `SetLogicalPaletteColor`-modified palettes (mods/Lua): not used in goldens.
   e. Anything under `DrawFPS`, `DrawConsole`, `_DEBUG` overlays: must be off in both.
   f. QoL text overlays (labels, health bars, floating numbers, XP bar) — off in both; they are float/timer driven
      (`Source/qol/floatingnumbers.cpp:99` uses `SDL_GetTicks` and a float multiplier).
   g. Hardware cursor: irrelevant (not in the framebuffer).
5. **Diff tooling**: `cmd/dgo-parity` with flags
   `--cpp <dump> --go <dump> --max-diff 0 --report-diff-png out.png`, printing first differing (x,y) and a per-region
   histogram; exit non-zero on any diff when the region is not in the allow-list.

**P3 — Seed-dungeon golden frames in CI (no GPU variance rule).**
Rule (Volume I H14): **CI must never depend on GPU rounding or driver behaviour.** Implementation:
1. Goldens are generated once by a human/privileged job and committed as `testdata/graphics/frames/<seed>-<level>-<frame>{.idx.zst,.png}`.
2. CI runs the comparison with `ebiten.NewImage`-free paths only — i.e. **P3 asserts the CPU-side index buffers**
   (`internal/render` keeps a CPU mirror of `worldIdx` for tests, updated by the same blitter functions that fill
   GPU textures? No: simpler and chosen — implement the blitters as pure Go index-buffer functions operating on
   `[]uint8` and have the GPU path upload that buffer per dirty tile). This makes the parity gate
   GPU-independent and also gives us the "plan B" software renderer from G8.
   ⇒ Concretely: the port keeps **one software index-rasteriser as the reference implementation** (fast enough at
   640×480: the C++ does it every frame) and the GPU path is an optimisation validated against it in tests.
3. Deterministic level states: build them from seeds, not from saves — reuse `test/fixtures/diablo/*.dun`
   (e.g. `1-2588.dun`) which dg-task-005 already parses, plus `test/fixtures/levels/l*data/*.dun` minisets and
   `assets/arena/*.dun`. Fixed content: place player at (20,20), `dLight` from the `dLight.dmp` fixture, one monster,
   one item, one object, one arch; `progress` pinned.
4. CI matrix: `go test -run Parity ./internal/render/...` on ubuntu-latest (no display) + macos + windows
   (headless, no ebiten init in that test binary — guard with the `headless` build tag).
5. `tools/linux_reduced_cpu_variance_run.sh` precedent for CPU-sensitive benches — reuse the idea for
   `just parity-bench`.
6. Optional nightly job with a real GPU (`ebitenginedebug` draw-command counting + screenshot diff) so GPU-specific
   regressions (atlas splits, filter mistakes) are caught without blocking PRs.

**P4 — Order & performance regression tests** (from dg-epic-03): instrumented order list (G6 test),
`DunRenderStats`-equivalent counters (`Source/engine/render/dun_render.cpp:1015` guarded by `DUN_RENDER_STATS`)
ported as `render_stats_test.go` to prove we blit the same number of micros per frame (a strong indirect parity signal).

### Ebitengine mapping
- Test-only helper `internal/render/testutil`: `DumpIndex() []uint8`, `LoadGolden(name)`, `Diff(a,b)`.
- `go:generate` step to regenerate goldens: `go run ./cmd/dgo-parity -record -dir testdata/graphics/frames`
  (fails CI if goldens change unexpectedly).
- Screenshots in Go: `screen.ReadPixels` (only in tests / `F1` handler) → PNG via `image/png`, plus the index dump.
  Note `ReadPixels` is a sync point — acceptable outside the steady-state draw.

### Parity test
(This section *is* the test plan; the acceptance mapping is: dg-task-007 ← P1, dg-task-010 ← P2/P3 + P4,
dg-task-011 ← P1 light/blending goldens + P2 blended-arch frames, dg-task-013 ← automap stamps goldens,
dg-task-014 ← `test/fixtures/text_render_integration_test` goldens. Extra gaps → EXTRA TASKS below.)

### Open questions
- Can CI legally fetch the retail MPQs for P1's CEL/CL2/PCX fixtures? If not, we need the derived-fixture approach
  (small, committed, generated locally). Flag for the operator.
- `SDL_VIDEODRIVER=dummy` + `PalSurface` creation: needs verification on the actual build (not attempted here — no
  C++ build performed in this read-only contract). If the dummy driver cannot create the 8-bit surface, P2 must use
  `OwnedSurface` directly in a test target instead of the real window path.

---

## EXTRA TASKS

Gaps found against /tmp/dg-bean-tree.md `dg-epic-02`/`dg-epic-03`/`dg-epic-11` and /tmp/dg-requirements.md §7:
(1) Volume I §7 prescribes `ColorScale`-based lighting/blending, which this extraction proves is *not* pixel-exact —
the index-framebuffer + palette-row + 256×256-LUT architecture needs its own task. (2) No task covers the dungeon
CEL **re-encode** + MIN/TIL/SOL loading pipeline (dg-task-007 covers sprites only). (3) The `*.pal` +
`paletteTransparencyLookup` + cycling pipeline has no owner. (4) Full-frame parity goldens (P2/P3) are named in
acceptance criteria but no task builds the dumper/CLI. (5) Automap line primitives (`automap_render.cpp` Bresenham
families) are non-obvious and are not called out in dg-task-013's research. (6) Asset-manifest/coverage tooling
(G9) is unowned.

```json
{"id": "dg-task-201", "type": "task", "title": "Dungeon tile pipeline: SOL/MIN/TIL load + CEL re-encode + micro renderer", "parent": "dg-epic-03", "blocked_by": ["dg-task-007", "dg-task-015"], "tags": ["phase-1", "render", "tiles"], "priority": "critical", "body": "# Dungeon tile pipeline\n\n## Goal\nLoad `*.sol`/`*.min`/`*.til`/`*.cel` for a level type, re-encode dungeon CELs exactly like `ReencodeDungeonCels`, and render micro tiles (6 TileTypes × 4 MaskTypes) at the correct screen positions with the tile's light table.\n\n## Research\n- /tmp/dg-graphics.md G2 (byte layouts, worked examples, the MIN index scramble, mask ASCII art, foliage split, re-encode sizes 512/784/1024).\n- Source/levels/dun_tile.hpp:19 (TileType doc comments), :83 (LevelCelBlock bits), :106 (TileProperties), :119 (MICROS), :131/132 (ReencodedTriangle/TrapezoidFrameSize)\n- Source/levels/dun_tile_data.cpp:68 (LoadMinData), :104 (LoadLevelSOLData + hand-patched SOL entries for L1/L3/L4/Crypt — transcribe the exact indices!), :165 (SetDungeonMicros incl. `blocks - 2 + (block & 1) - (block & 0xE)`)\n- Source/levels/reencode_dun_cels.cpp:24-180 (every re-encode helper incl. RenderTransparentSquare and ExtractFoliage*)\n- Source/engine/render/dun_render.hpp:28 (MaskType), :123 (GetDunFrame/Swap32LE), :147/:162; Source/engine/render/dun_render.cpp:92-104 (prefix increments), :466/:481 (unrolled triangle line widths/src offsets), :1046 (RenderTileFrame), :1085 (world_draw_black_tile)\n- Source/engine/render/scrollrt.cpp:549 (DrawCell incl. mask selection), :680 (DrawFloorTile), :955/:992/:1088 (row walk + DrawOOB/DrawDirtTile)\n- /tmp/dg-requirements.md §2 and §7; dg-epic-04 dg-task-015 provides the grids.\n- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md\n\n## Scope\ninternal/world/dungeoncel.go (sol/min/til load + micro scramble + re-encode + frame renumbering), internal/render/tile.go (RenderTileFrame for all 6 types × all masks, bottom-up row walking, foliage pass), atlas upload of decoded micros, `DrawFloor`/`DrawCell`/`DrawOOB`/`DrawDirtTile` wired into the row walk. NOT in scope: lighting tables (dg-task-202), entity order (dg-task-010).\n\n## Acceptance criteria\n- [ ] `TestMinScramble`: for blocks∈{10,16} the decoded `LevelCelBlock` list equals a C++ dump for the first 200 pieces of l1 and town\n- [ ] `TestReencode`: output buffer size equals `GetReencodedSize` and frame sizes are 512 (tri), 784 (trap), 1024 (square); foliage frames extracted count matches C++\n- [ ] `TestRenderTileGolden`: per (TileType × MaskType × lightLevel {0,8,15}) a 64×64 index buffer equals the C++ dump (add a dump gtest to devilutionX test/dun_render_benchmark harness)\n- [ ] `TestRowWalk`: 5×5 map produces the identical (tile, bufferX, bufferY) sequence as C++ (dg-epic-03 first criterion, owned here)\n- [ ] SOL hand-patches (l1 idx 9,15,16,20,21,27,28,51,56,58,61,63,65,72,208,247,253,257,323,403,24,450; l3 170,481,487; l4 210; crypt 142) are encoded as data and asserted\n- [ ] `world_draw_black_tile` and `DrawDirtTile` OOB fallbacks implemented and unit-tested\n- [ ] Draw path allocates 0 (testing.AllocsPerRun)\n\n## Skills for coder\n/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — SubImage/atlas rules for micro tiles\n/home/blaze/.agents/skills/go-dev/SKILL.md — index math and bounds-checked byte walking\n/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden-file comparisons\n/home/blaze/.agents/skills/coding-rules/SKILL.md — keep the re-encode helpers 1:1 named with C++ for reviewability\n\n## Dependencies & blockers\nblocked_by dg-task-007 (CLX decode) and dg-task-015 (grids). Blocks dg-task-010, dg-task-011.\n\n## Open questions\n- Endianness of the dungeon CEL frame table: mirror `Swap32LE` (BE-correct) or read LE? See /tmp/dg-graphics.md G2 open questions; decide with the operator and cover with the BE CI job (tools/run_big_endian_tests.sh).\n- Ship retail `.cel` fixtures for CI or only derived dumps?"}
```

```json
{"id": "dg-task-202", "type": "task", "title": "Index framebuffer + palette-tran shader (light tables, TRNs, blend LUT)", "parent": "dg-epic-03", "blocked_by": ["dg-task-007"], "tags": ["phase-1", "render", "palette", "parity"], "priority": "critical", "body": "# Paletted renderer core\n\n## Goal\nReplace the Volume I §7 `ColorScale` plan with the byte-exact architecture: world rendered into an *index* buffer (RGBA8 with the palette index in R), 16 palette-variant rows for lighting, per-TRN palette variants for uniques/infravision/pause, 256×256 blend LUT for transparency, one final palette-lookup pass.\n\n## Research\n- /tmp/dg-graphics.md G4 (light table generation transcription, TRN census, combo counts), G5 (lighting mapping + budget), G8 (blend LUT call sites and the 6 blit flavours), G10 D2/D4.\n- Source/engine/light_tables.cpp:19-70 (MakeLightTable incl. hell/nest/crypt overrides, FullyLit/FullyDark nulling)\n- Source/engine/render/blit_impl.hpp:20-190 (BlitDirect/WithMap/Blended/BlendedWithMap/Lightmap variants — the exact byte ops to mirror)\n- Source/engine/render/clx_render.cpp:556-590 + Source/engine/render/clx_render.hpp:40/:76/:81/:89/:107 (the draw flavours), Source/engine/render/clx_render.cpp:393 ClxApplyTrans\n- Source/utils/palette_blending.cpp:30-88 (BlendColors integer halving, symmetry, skipFrom/skipTo), Source/utils/palette_kd_tree.hpp (tie-breaking!)\n- Source/engine/palette.cpp:52-90 (CycleColors/CycleColorsReverse rotate palette AND LUT rows+columns), :140 ApplyFadeLevel, :188 LoadPaletteAndInitBlending\n- Source/engine/render/scrollrt.cpp:249 (ClxDrawLight), :264 (ClxDrawLightBlended), :378-405 (monster TRN tree), :468-495 (player light bias `max(l-5,0)` at :489)\n- Kage docs https://ebitengine.org/en/documents/shader.html ; batching https://ebitengine.org/en/documents/performancetips.html ; verified there is NO upstream examples/palette (https://github.com/hajimehoshi/ebiten/tree/v2.9.9/examples) — model the shader on examples/masking and examples/blend.\n- Skill: /home/blaze/.agents/skills/ebiten/rendering/SKILL.md\n\n## Scope\ninternal/render/palette.go (system palette, tone-map, fades, cycling, `[]color.RGBA` upload), internal/render/lighttable.go (port MakeLightTable verbatim incl. level overrides), internal/render/blendlut.go (port GenerateBlendedLookupTable + KD-tree), internal/render/shaders/palette.kage + blend.kage, the world index target + final pass, per-(trn,level) palette-variant cache with an LRU cap.\n\n## Acceptance criteria\n- [ ] `TestMakeLightTable` equals a C++ byte dump for all 16 tables × 6 level types (256×16×6 bytes)\n- [ ] `TestLightFalloffs` (16×128) and `TestLightConeInterpolations` (8×8×16×16, incl. perfect-square boundaries) byte-equal\n- [ ] `TestBlendingLut` byte-equal for town / L1 / crypt palettes with the correct skip ranges\n- [ ] Cycling advances palette AND re-derives the 16 light variants and the LUT (test asserts a known after-state hash)\n- [ ] Sprites draw with index-0 keying while tiles write index 0 as black (explicit regression test for the keying trap in G4)\n- [ ] Final pass is 1 draw command; measured GPU draw commands for a full frame < 40 (`-tags=ebitenginedebug` output captured in the test log artifact)\n- [ ] No ColorScale/ColorMatrix anywhere on the world index path (lint via a grep guard in the test)\n\n## Skills for coder\n/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — shader + offscreen target patterns\n/home/blaze/.agents/skills/go-dev/SKILL.md — pure-Go reference blitters as the parity oracle\n/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden byte comparison\n/home/blaze/.agents/skills/ponytail/SKILL.md — smallest change that still hits byte-exactness (no new render graph)\n\n## Dependencies & blockers\nblocked_by dg-task-007. Blocks dg-task-011 (which becomes 'blend + automap transparency' only) and dg-task-010.\n\n## Open questions\n- Keep a pure-Go software index rasteriser as reference implementation (recommended, gives GPU-free CI parity)? Yes — decide and record here.\n- Uniform-per-draw (palette row index) vs one image per variant: measure batch splits with ebitenginedebug and pick."}
```

```json
{"id": "dg-task-203", "type": "task", "title": "Full-frame parity harness: C++ backbuffer dumper + Go comparator + seed-dungeon goldens", "parent": "dg-epic-03", "blocked_by": ["dg-task-005", "dg-task-010", "dg-task-201", "dg-task-202"], "tags": ["phase-1", "testing", "parity"], "priority": "critical", "body": "# Pixel-parity harness\n\n## Goal\nMake 'pixel-exact' falsifiable: dump C++ indexed backbuffers for deterministic scenes, assert Go produces byte-identical index buffers (and colour-identical PNGs modulo a documented allow-list), runnable in CI without a GPU.\n\n## Research\n- /tmp/dg-graphics.md G11 (plan P1-P4, allowed palette-stage differences a-g), G1 (frame order), G8 (capture.cpp flow).\n- Source/capture.cpp:84 CaptureScreen, :97 RedPalette; Source/utils/surface_to_pcx.cpp:45/70/92/150; Source/utils/surface_to_png.hpp\n- Source/engine/render/scrollrt.cpp:1814 `scrollrt_draw_game_screen` (early-return at `:1816`) and `:1834`
  `DrawAndBlit` (early-return at `:1836`) — this is why a dump hook is needed rather than a screenshot loop\n- Source/engine/palette.cpp:271/324 (short-circuits at `:273`/`:275` and `:326`/`:328`), `:363`/`:373`/`:391` cycling\n- Source/diablo.cpp:1031-1033 (--record/--demo/--timedemo), :1127 parse, :1204 InitPlayBack\n- test/timedemo_test.cpp:28-110 (drives the same code path with only SDL_INIT_EVENTS), :89 AdjustToScreenGeometry\n- Fixtures to reuse: test/fixtures/diablo/*.dun, test/fixtures/hellfire/*.dun, test/fixtures/levels/l{1..4}data/*.dun, test/fixtures/light_render_benchmark/dLight.dmp, test/fixtures/text_render_integration_test/*.png, assets/arena/*.dun\n- tools/linux_reduced_cpu_variance_run.sh, tools/measure_timedemo_performance.py\n- /tmp/dg-requirements.md Appendix H H14 (no GPU variance rule), Appendix K (C++ test → Go test matrix)\n\n## Scope\n(a) a gtest in the devilutionX tree (`test/backbuffer_dump_test.cpp`, documented in devilutionGo/docs/parity.md, NOT committed to this repo by this task) writing `<frame>.idx.bin` + `<frame>.pal.bin` + `<frame>.ppm`;\n(b) internal/render/parity: loader for the dumps, `Diff`, per-region allow-list, `cmd/dgo-parity` CLI (`-record`, `-compare`, exit code);\n(c) 3 seeds × 3 view positions × progress∈{0,64,127} golden frames generated from .dun fixtures with a scripted minimal world (player, monster, item, object, arch, automap on/off);\n(d) `just parity` target and CI wiring (ubuntu/macos/windows, no display).\n\n## Acceptance criteria\n- [ ] `just parity` passes on a headless runner and fails if any golden index byte changes (diff report names the first differing (x,y) and the owning tile/sprite)\n- [ ] Index-exact gate present AND separate colour gate present, with the allow-list encoded as data (brightness=0, cycling off, fades settled, QoL overlays off, debug overlays off)\n- [ ] At least one golden frame each for: town (trees deferred-one-row rule), L1 cathedral (arches + transparency + foliage), L4 hell (light-table override), crypt (nest/crypt override + lava brightness)\n- [ ] Automap-on golden frame compares (dg-task-013 dependency)\n- [ ] Text goldens reused from test/fixtures/text_render_integration_test/*.png via the same comparator (proves the pipeline composes)\n- [ ] Documentation: how to regenerate the C++ dumps (commands, SDL_VIDEODRIVER=dummy requirement, MPQ location) verified by running them once\n\n## Skills for coder\n/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden-file + fixture matrix\n/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — ReadPixels/WritePixels only inside test helpers\n/home/blaze/.agents/skills/go-dev/SKILL.md — CLI structure, no CGO, build tags for headless\n/home/blaze/.agents/skills/beans/SKILL.md — keep this bean's todo list current\n\n## Dependencies & blockers\nblocked_by dg-task-010, dg-task-201, dg-task-202, dg-task-005. Blocks the dg-epic-03 'golden frames' acceptance criterion and milestone M2 sign-off.\n\n## Open questions\n- Legal/CI question: can the job fetch retail MPQs? If no, all goldens must come from the committed fixtures + bundled assets path.\n- Do we need Windows/macOS goldens (fonts/driver differences) or is one platform + 'CPU variance' rule enough?"}
```

```json
{"id": "dg-task-204", "type": "task", "title": "Palette/TRN registry: level palettes, cycling, all TRN application points", "parent": "dg-epic-02", "blocked_by": ["dg-task-007"], "tags": ["phase-1", "assets", "palette"], "priority": "high", "body": "# Palette & TRN registry\n\n## Goal\nOwn the *data* side of colour: per-level palette selection (incl. random 1..4 / nest 1..5), all `.trn` files and the exact order they are applied, baked-vs-per-draw TRN classification, and the RNG call that picks the palette.\n\n## Research\n- /tmp/dg-graphics.md G4 (TRN census table), G9 (asset manifest rows for palettes/TRNs).\n- Source/engine/palette.cpp:175 LoadPalette, :188 LoadPaletteAndInitBlending (skip ranges per leveltype), :196-213 LoadRndLvlPal (town/crypt special cases, `RandomIntBetween(1,4)`, nest `alternateNestArt` +1)\n- Source/engine/trn.cpp:29 GetClassTRN (`plrgfx\\<class>.trn` via `spriteData.trn`), :47 GetPlayerGraphicTRN, :14/19/24 infra/stone/pause\n- Source/monster.cpp:173-192 InitMonsterTRN (255→0 replace quirk, applied to list OR sheet, skip index 1 for Councilor/Magistr/Cabalist/Advocate!), :3318 unique TRN\n- Source/player.cpp:2167-2176 (graphic TRN then class TRN order)\n- Source/lighting.cpp:214 LoadTrns\n- Source/engine/render/text_render.cpp:88 ColorTranslations[19] (3 null slots) and :188 lazy load\n- assets/txtdata/monsters/monstdat.tsv column `trnFile`; assets/plrgfx/warrior/whu/whu{fm,lm,qm}.trn (repo ships 20 .trn — verified `find assets -iname '*.trn' | wc -l`)\n- /tmp/dg-requirements.md §6 (asset pipeline), Appendix D.5\n\n## Scope\ninternal/assets/palette.go (Load + naming + random pick with the shared RNG), internal/assets/trn.go (registry: name → 256 bytes → application mode baked/per-draw), golden tests. Feeds dg-task-202's variant cache and dg-task-010/011's draw decisions.\n\n## Acceptance criteria\n- [ ] Palette file name selection reproduces C++ for every leveltype including `alternateNestArt` and the RNG advance (asserted with seeded RNG vectors)\n- [ ] All 19 ColorTranslations slots resolved with the 3 null entries mapped to 'no TRN'\n- [ ] `ClxApplyTrans` equivalent applied in the same order as C++ for players (graphic then class) and monsters (type), and NOT applied for the 4 skipped demon animations (Councilor/Magistr/Cabalist/Advocate index 1)\n- [ ] Monster TRN 255→0 quirk covered by a test using a real 256-byte fixture\n- [ ] Every `.trn` in the bundled asset set loads with size 256 and round-trips through the registry (count assertion)\n\n## Skills for coder\n/home/blaze/.agents/skills/ebiten/resources/SKILL.md — embed + lazy load\n/home/blaze/.agents/skills/go-dev/SKILL.md — registry/typed errors, no globals in exported API\n/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded tables\n\n## Dependencies & blockers\nblocked_by dg-task-007. Blocks dg-task-202 (variant cache) and dg-epic-05 item/monster colouring.\n\n## Open questions\n- Where does `_pgfxnum` (armour/weapon graphic index) get computed — entity or assets? Cross-repo with dg-task-016."}
```

```json
{"id": "dg-task-205", "type": "task", "title": "Automap line primitives + transparent/minimap blend modes", "parent": "dg-epic-03", "blocked_by": ["dg-task-010"], "tags": ["phase-1", "render", "automap"], "priority": "normal", "body": "# Automap primitives\n\n## Goal\nPort the exact pixel-plotted line/stamp primitives that make the automap look identical, plus its two blend modes; dg-task-013 covers the layout/text/arrow, this covers the raster primitives and blending.\n\n## Research\n- /tmp/dg-graphics.md G6 §H (scan loop, colours, stamp families), G8 §A (blend call sites).\n- Source/engine/render/automap_render.cpp:26 DrawMapLine<DirX,DirY>, :43 DrawMapLineSteep<DirX,DirY>, :60/:78/:96/:101/:106/:111/:116/:121/:126/:131 wrappers, :139 DrawMapFreeLine (Bresenham), :166 SetMapPixel (minimap clip at :168 + transparent half-blend at :171)\n- Source/automap.cpp:167 DrawDiamond, :200/:250 DrawWallConnections (NE/SE variants), :275 DrawDirt (explicit SetMapPixel list), :302-330 other stamps, :38-57 MapColors enum, :980 DrawAutomapTile (per-exploration colours, noConnect rules, lava/acid/water colour choice), :1776-1781 cells/scale computation, :1847-1861 double loop\n- Source/automap.h:66 AmLineLength values, AmWidthOffset/AmHeightOffset tables (:41-64)\n- Source/engine/render/primitive_render.cpp:18/:37/:89/:107/:140/:155/:159/:187/:215 half-transparent line/rect/pixel paths\n- Source/levels/gendung.cpp:424 (.dun transparency layer → dTransVal), Source/automap.cpp:1528 automap.amp\n- test/automap_test.cpp\n\n## Scope\ninternal/render/automapstamps.go (all line families + dirt/stairs/diamond/grate/arch/cave-cross/fence stamps), blend-mode routing through dg-task-202's LUT, minimap rect clipping, scale/zoom handling. Extend dg-task-013's renderer to use them.\n\n## Acceptance criteria\n- [ ] For each AmLineLength × each of the 8 line functions, a 24×24 index-buffer stamp equals a C++ dump\n- [ ] `DrawAutomapTile` golden for every AutomapTile::Types value with representative flag combos (incl. the noConnect special cases for L2/L3 doors and arches)\n- [ ] AutomapType::Transparent compares equal to C++ `lut[color][dst]` results, not to an alpha blend\n- [ ] Minimap clipping test (pixels outside MinimapRect are not written)\n- [ ] automap_test.cpp ported cases pass\n\n## Skills for coder\n/home/blaze/.agents/skills/ebiten/rendering/SKILL.md\n/home/blaze/.agents/skills/go-dev/SKILL.md — small integer loops, no float\n/home/blaze/.agents/skills/go-fluent-testing/SKILL.md\n\n## Dependencies & blockers\nblocked_by dg-task-010; needs dg-task-202 for the blend LUT. Blocks dg-epic-11 automap QoL items.\n\n## Open questions\n- Do we ever draw the automap through the sprite atlas instead of pixel stamps? No — stamps are required for parity; document the decision."}
```

```json
{"id": "dg-task-206", "type": "task", "title": "Graphics asset manifest, coverage checker and CLX conversion tooling", "parent": "dg-epic-02", "blocked_by": ["dg-task-006", "dg-task-007"], "tags": ["phase-1", "assets", "tooling"], "priority": "medium", "body": "# Graphics asset manifest & tooling\n\n## Goal\nMachine-checked inventory of every graphics asset the renderer needs (family → archive path → format → frame layout), so 'missing art' is a build-time error, not a runtime surprise, and so retail `.cel/.cl2/.pcx` can be pre-converted to `.clx` for the bundled path.\n\n## Research\n- /tmp/dg-graphics.md G9 (the table + extension rules + measured counts/sizes), G3 (formats), G1 §extension rule.\n- Source/engine/load_cel.hpp:12/14, Source/engine/load_cl2.hpp:23/27, Source/engine/load_pcx.hpp:15/17 (`.clx` vs `.cel/.cl2/.pcx` by build flag)\n- Source/diablo.cpp:1359 LoadLvlGFX (all level triples), Source/levels/dun_tile_data.cpp:68/:104 (min/sol), Source/cursor.cpp:445-452 (+ `-widths.txt`), Source/items.cpp:2424, Source/objects.cpp:3700, Source/towners.cpp:110/:780, Source/player.cpp:1530/:2163, Source/monster.cpp:3217, Source/panels/mainpanel.cpp:100, Source/DiabloUI/{button.cpp:34,credits.cpp:189,diabloui.cpp:629-654}.cpp\n- Source/engine/assets.cpp:410 LoadMPQ + assets.hpp:309 priorities\n- tools/update_bundled_assets.sh (how devilutionX produces the bundled `.clx` set), 3rdParty/mpqfs\n- assets/ tree: 100 `.clx`, 20 `.trn`, 2 `.pal`, 76 font `.clx`, `assets/data/inv/objcurs-widths.txt` (verified counts)\n\n## Scope\ngraphics_manifest.tsv (hand-curated from G9, machine-readable), `cmd/dgo-assets` with subcommands `list`, `coverage`, `convert` (CEL/CL2/PCX → CLX using the Go decoders, with the same width sources as C++), CI job asserting every manifest entry resolves from either the bundled dir or a provided MPQ, and a report of unresolvable names.\n\n## Acceptance criteria\n- [ ] Manifest covers all 7 level types (cel/til/min/sol/special/pal), players (6 classes × anims × 8 dirs), monsters (all 113 monstdat rows with a sprite), items (35 vs 43 under spawn/hellfire), objects, towners, missiles, cursors (+widths), fonts (6 sizes × available rows), UI (`ui_art/*`), gendata cutscenes\n- [ ] `coverage` exits non-zero on a missing entry and prints the loader file:line that asked for it\n- [ ] `convert` output for the 100 bundled `.clx` files is byte-identical to the committed files (strongest possible decoder cross-validation, uses repo assets only)\n- [ ] Widths sourcing matches C++ for every family (no hard-coded width where C++ reads a `-widths.txt`)\n- [ ] Docs/graphics.md generated from the manifest\n\n## Skills for coder\n/home/blaze/.agents/skills/ebiten/resources/SKILL.md — embed + asset lifecycle\n/home/blaze/.agents/skills/go-dev/SKILL.md — CLI + table-driven codegen\n/home/blaze/.agents/skills/go-setup/SKILL.md — where tooling lives in the tree\n/home/blaze/.agents/skills/coding-rules/SKILL.md — data over code\n\n## Dependencies & blockers\nblocked_by dg-task-006, dg-task-007. Blocks dg-task-010/dg-task-012/dg-task-013/dg-task-014 content checks ('all art present').\n\n## Open questions\n- Where do the `ui_art` PCX names come from for a full enumeration (grep + DevilutionX `devilutionx.mpq` build)?\n- Should `convert` be part of a release pipeline or developer-only?"}
```

---

## VERIFY (commands actually run)

```
$ wc -l /tmp/dg-graphics.md
<see value printed below>

$ grep -c '^## G' /tmp/dg-graphics.md        # 11 section headers expected
11

$ grep -oE 'Source/[A-Za-z0-9_/.-]+:[0-9]+' /tmp/dg-graphics.md | sort -u | wc -l
455

$ grep -oE 'assets/[A-Za-z0-9_./-]+' /tmp/dg-graphics.md | sort -u | wc -l
18

$ grep -oE 'test/[A-Za-z0-9_./-]+' /tmp/dg-graphics.md | sort -u | wc -l
21

$ grep -c '^```' /tmp/dg-graphics.md
94            (>= 15 required)

$ python3 -c "import json,re;s=open('/tmp/dg-graphics.md').read();\
  b=re.findall(r'\`\`\`json\n(.*?)\`\`\`',s,re.S);\
  ids=set(re.findall(r'\"dg-[a-z]+-\d+\"', open('/tmp/dg-bean-tree.md').read()));\
  [print(j['id'], j['parent'], j['parent'] in ids, len(j['body'])) for j in map(json.loads,b)]"
dg-task-201 dg-epic-03 True 3268
dg-task-202 dg-epic-03 True 3236
dg-task-203 dg-epic-03 True 2987
dg-task-204 dg-epic-02 True 2466
dg-task-205 dg-epic-03 True 1939
dg-task-206 dg-epic-02 True 2310
```

All 6 EXTRA TASK blocks parse as JSON, use the required schema (`id/type/title/parent/blocked_by/tags/priority/body`
with `Goal/Research/Scope/Acceptance criteria (≥3 checkboxes)/Skills for coder (real paths)/Dependencies & blockers/
Open questions`) and every `parent` id exists in `/tmp/dg-bean-tree.md`.

Verification commands used while extracting (spot list, all executed and their output used above):
`git rev-parse HEAD`, `wc -l` over ~40 source files, `python3 /tmp/dg_gl.py <regex> <files>` (a line-number grep
helper written to /tmp), and Python decoders run against `assets/data/boxleftend.clx`, `boxrightend.clx`,
`dirtybuc.clx`, `health.clx`, `assets/fonts/30-00.clx`, plus read-only `smpq -e` extraction of
`levels/l1data/l1.{cel,til,min,sol}`, `levels/l1data/l1_1.pal`, `levels/l1data/l1s.cel`,
`levels/towndata/town.{min,sol,pal}`, `data/inv/objcurs.cel`, `plrgfx/infra.trn`, `plrgfx/stone.trn` from a local
`DIABDAT.MPQ` into `/tmp/dgx/`. Repo writes: none. Git write commands: none. Beans CLI: not used.

Notes on contract items that do not exist in this tree (stated rather than invented):
- `PresentPrimarySurface` → `RenderPresent` / `BltFast` (`Source/engine/dx.cpp:287`, `:181`).
- `SetPalette` → `LoadPalette`/`UpdateSystemPalette`/`SystemPaletteUpdated` (`Source/engine/palette.cpp:175/152/170`);
  no gamma exists in this fork.
- `level_cel_decode` / `gendung rendering tables (dPiece placement)` → `reencode_dun_cels.cpp` +
  `SetDungeonMicros` + `DRLG_LPass3` (`Source/levels/gendung.cpp:578`).
- `run_length_walk` / `run_bin_walk` → `ClxBlitInfo` / `SkipRestOfLineWithOverrun` / `GetSkipSize`
  (`Source/engine/render/clx_render.cpp:73`, `:94`, `Source/utils/clx_decode.hpp:49`).
- `blixmap` tables → `paletteTransparencyLookup[256][256]` (`Source/utils/palette_blending.cpp:19`).
- item sparkle sprites → none exist; the mechanism is palette cycling + outlines + `infra.trn`.
- `dun_tile_data.cpp` "micro-tile formats" → the SOL/MIN/TIL loaders in that file (G2/G9).

GRAPHICS-CONTRACT-COMPLETE
