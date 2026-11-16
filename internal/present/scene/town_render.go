package scene

import (
	"encoding/binary"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// townWorldView carries the adapter state one composed world needs: the
// grids, the decoded micro blocks and the level SOL properties.
type townWorldView struct {
	grid   levels.PieceGrid
	micros render.DungeonMicros
	sol    render.SOLData
	light  [levels.MAXDUNX][levels.MAXDUNY]uint8
}

// townRenderer is the raster sink: the floor and cell draws decode their
// micro block's re-encoded frame and blit it through RenderTileFrame; the
// entity and arch layers stay empty until their beans land (devilGo-s7x0
// adds the player layer).
type townRenderer struct {
	surface render.Surface
	view    *townWorldView
	cel     []byte
	state   render.PaletteState
	tables  light.LightTables
}

// DrawFloorTile renders the floor diamond pair of one tile
// (scrollrt.cpp:687-706 DrawFloorTile).
func (r *townRenderer) DrawFloorTile(t render.DrawTarget) {
	r.drawMicro(t, 0)
	r.drawMicroRight(t, 1)
}

// DrawCell renders one tile's wall micro stack: the floor pair (walls only;
// floors render in the floor pass) then the tall rows
// (scrollrt.cpp:611-666 DrawCell).
func (r *townRenderer) DrawCell(t render.DrawTarget) {
	micro, ok := r.microsOf(t.Tile)
	if !ok || r.view.floorOf(t.Tile) {
		return
	}

	r.drawFloorPair(t, micro)
	r.drawTallRows(t, micro)
}

// DrawBlackTile stamps the town OOB black diamond
// (scrollrt.cpp:1090-1093, tile.go WorldDrawBlackTile).
func (r *townRenderer) DrawBlackTile(at render.PixelPos) {
	render.WorldDrawBlackTile(r.surface, at.X, at.Y)
}

func (*townRenderer) DrawMissilePre(render.DrawTarget)        {}
func (*townRenderer) DrawCorpse(render.DrawTarget)            {}
func (*townRenderer) DrawObjectPre(render.DrawTarget)         {}
func (*townRenderer) DrawItemPre(render.DrawTarget)           {}
func (*townRenderer) DrawDeadPlayer(render.DrawTarget)        {}
func (*townRenderer) DrawPlayer(render.DrawTarget, int)       {}
func (*townRenderer) DrawMonster(render.DrawTarget, int)      {}
func (*townRenderer) DrawItemPost(render.DrawTarget)          {}
func (*townRenderer) DrawObjectPost(render.DrawTarget)        {}
func (*townRenderer) DrawMissilePost(render.DrawTarget)       {}
func (*townRenderer) DrawArch(render.DrawTarget, int8)        {}
func (*townRenderer) DrawArchBlended(render.DrawTarget, int8) {}

// DrawArchDirect draws the town tree leaves unlit (scrollrt.cpp:944-952
// ClxDraw branch); the special-cel decode lands with the arch bean.
func (*townRenderer) DrawArchDirect(render.DrawTarget, int8) {}

// solOf is the SOL lookup upstream does through SOLData[dPiece]
// (scrollrt.cpp:611-666 dispatch on tile properties).
func (w *townWorldView) solOf(tile render.TilePos) render.TileProperties {
	piece := int(w.grid.Piece[tile.X][tile.Y])
	if piece <= 0 || piece >= len(w.sol) {
		return 0
	}

	return w.sol[piece]
}

// floorOf is IsFloor: !TileHasAny(tile, Solid|BlockMissile)
// (scrollrt.cpp:114-117).
func (w *townWorldView) floorOf(tile render.TilePos) bool {
	return w.solOf(tile)&(render.TilePropertiesSolid|render.TilePropertiesBlockMissile) == 0
}

// world is the render.World view over the town grids (scrollrt.cpp:784-953
// reads the dPiece/dLight/dSpecial/TransList globals directly; this adapter
// injects them).
func (w *townWorldView) world() render.World {
	return render.World{
		InBounds: func(tile render.TilePos) bool {
			return tile.X >= 0 && tile.X < levels.MAXDUNX && tile.Y >= 0 && tile.Y < levels.MAXDUNY
		},
		IsFloor:  w.floorOf,
		IsWall:   func(tile render.TilePos) bool { return !w.floorOf(tile) || w.grid.Special[tile.X][tile.Y] != 0 },
		NotSolid: func(tile render.TilePos) bool { return w.solOf(tile)&render.TilePropertiesSolid == 0 },
		Light:    func(tile render.TilePos) uint8 { return w.light[tile.X][tile.Y] },
		Piece:    func(tile render.TilePos) uint16 { return w.grid.Piece[tile.X][tile.Y] },
		Special:  func(tile render.TilePos) int8 { return w.grid.Special[tile.X][tile.Y] },
		Transparent: func(render.TilePos) bool {
			return false
		},
		Corpse:       func(render.TilePos) int8 { return 0 },
		Item:         func(render.TilePos) int8 { return 0 },
		ItemPostDraw: func(int8) bool { return false },
		DeadPlayer:   func(render.TilePos) bool { return false },
		Object:       func(render.TilePos) (render.ObjectRef, bool) { return render.ObjectRef{}, false },
		Player:       func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		Monster:      func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		Town:         true,
	}
}

// drawMicro renders one half of the floor diamond pair: a triangle of
// DunFrameTriangleHeight rows at the tile's bottom-corner position; the
// right half displaces one DunFrameWidth east
// (scrollrt.cpp RightFrameDisplacement {DunFrameWidth, 0}).
func (r *townRenderer) drawMicro(t render.DrawTarget, block int) {
	micro, ok := r.microsOf(t.Tile)
	if !ok || !micro[block].HasValue() {
		return
	}

	r.drawBlock(t, micro[block])
}

func (r *townRenderer) drawMicroRight(t render.DrawTarget, block int) {
	t.At.X += render.DunFrameWidth
	r.drawMicro(t, block)
}

// drawFloorPair renders micro tiles 0/1 for a wall tile, the right half
// displaced one DunFrameWidth east (scrollrt.cpp:618-641).
func (r *townRenderer) drawFloorPair(t render.DrawTarget, micro render.MICROS) {
	for _, block := range []int{0, 1} {
		if !micro[block].HasValue() {
			continue
		}

		at := t.At
		if block == 1 {
			at.X += render.DunFrameWidth
		}

		r.drawBlock(render.DrawTarget{Tile: t.Tile, At: at, Light: t.Light}, micro[block])
	}
}

// drawTallRows renders micro tiles 2.., one pair per row walking up
// (scrollrt.cpp:640-666).
func (r *townRenderer) drawTallRows(t render.DrawTarget, micro render.MICROS) {
	at := t.At
	for i := 2; i+1 < len(micro); i += 2 {
		at.Y -= render.TileHeight

		for _, block := range []int{i, i + 1} {
			if !micro[block].HasValue() {
				continue
			}

			r.drawBlock(render.DrawTarget{Tile: t.Tile, At: at, Light: t.Light}, micro[block])
		}
	}
}

// microsOf resolves one tile's decoded micro blocks, false past the table.
func (r *townRenderer) microsOf(tile render.TilePos) (render.MICROS, bool) {
	piece := int(r.view.grid.Piece[tile.X][tile.Y])
	if piece <= 0 || piece >= len(r.view.micros) {
		return render.MICROS{}, false
	}

	return r.view.micros[piece], true
}

// drawBlock blits one micro block's re-encoded frame
// (dun_render.hpp:148-157 RenderTile: triangle frames are 31 rows, every
// other shape 32; the draw is Solid — transparency dispatch is the
// deferred clip seam of tile.go).
func (r *townRenderer) drawBlock(t render.DrawTarget, block render.LevelCelBlock) {
	src := dunFrame(r.cel, block.Frame())

	frameHeight := render.DunFrameHeight
	if block.Type() == render.TileLeftTriangle || block.Type() == render.TileRightTriangle {
		frameHeight = render.TriangleHeight
	}

	lighting := render.TileLighting{Level: t.Light, Blend: r.state.LUT, Tables: &r.tables.Tables}

	render.RenderTileFrame(r.surface, t.At, block.Type(), src, frameHeight, render.MaskSolid, lighting)
}

// dunFrame is GetDunFrame (dun_render.hpp:123-127): the frame table lookup
// over the re-encoded CEL.
func dunFrame(cel []byte, frame uint16) []byte {
	offset := int(binary.LittleEndian.Uint32(cel[4*frame:]))

	return cel[offset:]
}
