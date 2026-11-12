package paritytest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
	"github.com/bmcszk/devilutionGo/internal/render/parity"
)

const (
	// paletteFixture is the committed Cathedral palette dump the state loads.
	paletteFixture = "l1_1.pal.hex"
	// paletteDir is the committed palette fixture directory.
	paletteDir = "pal"
	// sceneWidth is the world-tile width of the scripted scene.
	sceneWidth = 40
	// sceneHeight is the world-tile height of the scripted scene.
	sceneHeight = 40
	// sceneAnchor is the tile the view is centred on.
	sceneAnchor = 20
	// viewRows is the tile-row count of the captured view.
	viewRows = 7
	// viewColumns is the tile-column count of the captured view.
	viewColumns = 9
	// viewMicroTiles is MicroTileLen of the captured view.
	viewMicroTiles = 4
	// viewOriginX is the target-buffer origin x of the captured view.
	viewOriginX = 320
	// viewOriginY is the target-buffer origin y of the captured view.
	viewOriginY = 200
)

// sceneTile is one scripted world tile: the grids under one dPiece coordinate.
type sceneTile struct {
	floor    bool
	wall     bool
	notSolid bool
	special  int8
	corpse   int8
	item     int8
	object   bool
	light    uint8
}

// scene is the scripted dungeon behind the captured frame.
type scene struct {
	tiles  map[render.TilePos]sceneTile
	width  int
	height int
}

// rasterSink paints every layer draw into an index surface: a later draw
// overwrites an earlier one, so the surface pins the draw order as well as the
// geometry.
type rasterSink struct {
	surface render.Surface
}

// World builds the scripted minimal world: a deterministic mixed terrain of
// floors, walls, arches, corpses, items and objects, walked as one 640x480
// frame. SetRndSeed runs immediately before assembly, freezing the LCG for
// everything the assembly consumes.
func World() render.World {
	engine.SetRndSeed(Seed)

	return scriptedWorld().world()
}

// RandomStateAfter returns the engine LCG state after assembling the scripted
// world, so a test can pin the freeze point.
func RandomStateAfter(_ render.World) uint32 {
	return engine.GetLCGEngineState()
}

// PaletteState builds the presenter state of the scripted world: the Cathedral
// palette, its light tables and blend ranges, cycling untouched.
func PaletteState() render.PaletteState {
	return render.NewPaletteState(loadPalette(), light.MakeLightTables(leveldat.DTCathedral), assets.BlendingFor(0))
}

// SoftwareSurface rasterises the scripted world on the CPU: the tier-1 oracle
// the captured frame's index bytes are compared against.
func SoftwareSurface(world render.World) render.Surface {
	return walkWorld(world)
}

// PaletteFromVariants expands the state's undimmed variants row back to palette
// indices: shade 0 is the bright row (dLight 0 = fully lit), so each index's
// colour is variant[0][index] — the palette row the C++ dumper writes for a
// fade-256 frame with cycling off.
func PaletteFromVariants(state render.PaletteState) parity.Palette {
	var palette parity.Palette
	for index := range palette {
		rgb := state.Variants[0][index]
		palette[index] = parity.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xFF}
	}

	return palette
}

// scriptedWorld assembles the deterministic scene: every tile's state derives
// from its coordinates, every entity kind appears at least once, and the view
// is fixed. Runs after World's SetRndSeed freeze.
func scriptedWorld() *scene {
	world := &scene{tiles: make(map[render.TilePos]sceneTile), width: sceneWidth, height: sceneHeight}

	for x := 15; x < 30; x++ {
		for y := 15; y < 30; y++ {
			world.tiles[render.TilePos{X: x, Y: y}] = scriptedTile(x, y)
		}
	}

	anchor := render.TilePos{X: sceneAnchor, Y: sceneAnchor}
	world.tiles[anchor] = sceneTile{
		floor: true, light: 3, special: 1,
		corpse: 2, item: 2, object: true,
	}

	return world
}

// scriptedTile derives one tile's state from its coordinates.
func scriptedTile(x, y int) sceneTile {
	state := sceneTile{floor: true}
	if (x+y)%3 == 0 {
		state.wall, state.notSolid = true, true
	}

	if (x*7+y*13)%5 == 0 {
		state.special = 1
	}

	if (x+y)%4 == 0 {
		state.corpse, state.light = 1, 9
	}

	if (x*3+y*5)%7 == 0 {
		state.item = 1
	}

	if (x+y)%6 == 0 {
		state.object = true
	}

	return state
}

// world wraps the scene as the renderer's injected grid view.
func (s *scene) world() render.World {
	return render.World{
		InBounds: func(tile render.TilePos) bool {
			return tile.X >= 0 && tile.Y >= 0 && tile.X < s.width && tile.Y < s.height
		},
		IsFloor:  func(tile render.TilePos) bool { return s.at(tile).floor },
		IsWall:   func(tile render.TilePos) bool { return s.at(tile).wall },
		NotSolid: func(tile render.TilePos) bool { return s.at(tile).notSolid },
		Light:    func(tile render.TilePos) uint8 { return s.at(tile).light },
		Piece:    func(render.TilePos) uint16 { return 0 },
		Special:  func(tile render.TilePos) int8 { return s.at(tile).special },
		Transparent: func(render.TilePos) bool {
			return false
		},
		Corpse: func(tile render.TilePos) int8 { return s.at(tile).corpse },
		Item:   func(tile render.TilePos) int8 { return s.at(tile).item },
		ItemPostDraw: func(int8) bool {
			return false
		},
		DeadPlayer: func(render.TilePos) bool { return false },
		Object: func(tile render.TilePos) (render.ObjectRef, bool) {
			return render.ObjectRef{}, s.at(tile).object
		},
		Player:     func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		Monster:    func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		MissilePre: false,
		Town:       false,
	}
}

// at returns tile's scripted state, the zero value outside the script.
func (s *scene) at(tile render.TilePos) sceneTile { return s.tiles[tile] }

// walkWorld rasterises one frame of world into a 640x480 index surface: the
// same three passes DrawGame runs, in upstream order.
func walkWorld(world render.World) render.Surface {
	sink := &rasterSink{surface: render.NewSurface(render.WorldLogicalWidth, render.WorldLogicalHeight)}

	render.DrawWorld(sink, world, render.View{
		Start:       render.TilePos{X: sceneAnchor, Y: sceneAnchor},
		Origin:      render.PixelPos{X: viewOriginX, Y: viewOriginY},
		Rows:        viewRows,
		Columns:     viewColumns,
		ScreenWidth: render.WorldLogicalWidth,
		MicroTiles:  viewMicroTiles,
	})

	return sink.surface
}

// DrawFloorTile stamps the floor layer.
func (s *rasterSink) DrawFloorTile(t render.DrawTarget) { s.stamp(t, 1) }

// DrawCell stamps the wall/cell layer.
func (s *rasterSink) DrawCell(t render.DrawTarget) { s.stamp(t, 2) }

// DrawMissilePre stamps the pre-missile layer.
func (s *rasterSink) DrawMissilePre(t render.DrawTarget) { s.stamp(t, 3) }

// DrawCorpse stamps the corpse layer.
func (s *rasterSink) DrawCorpse(t render.DrawTarget) { s.stamp(t, 4) }

// DrawObjectPre stamps the pre-object layer.
func (s *rasterSink) DrawObjectPre(t render.DrawTarget) { s.stamp(t, 5) }

// DrawItemPre stamps the pre-item layer.
func (s *rasterSink) DrawItemPre(t render.DrawTarget) { s.stamp(t, 6) }

// DrawDeadPlayer stamps the dead-player layer.
func (s *rasterSink) DrawDeadPlayer(t render.DrawTarget) { s.stamp(t, 7) }

// DrawPlayer stamps the player layer.
func (s *rasterSink) DrawPlayer(t render.DrawTarget, _ int) { s.stamp(t, 8) }

// DrawMonster stamps the monster layer.
func (s *rasterSink) DrawMonster(t render.DrawTarget, _ int) { s.stamp(t, 9) }

// DrawItemPost stamps the post-item layer.
func (s *rasterSink) DrawItemPost(t render.DrawTarget) { s.stamp(t, 10) }

// DrawObjectPost stamps the post-object layer.
func (s *rasterSink) DrawObjectPost(t render.DrawTarget) { s.stamp(t, 11) }

// DrawMissilePost stamps the post-missile layer.
func (s *rasterSink) DrawMissilePost(t render.DrawTarget) { s.stamp(t, 12) }

// DrawArch stamps the arch layer.
func (s *rasterSink) DrawArch(t render.DrawTarget, _ int8) { s.stamp(t, 13) }

// DrawArchBlended stamps the blended-arch layer.
func (s *rasterSink) DrawArchBlended(t render.DrawTarget, _ int8) { s.stamp(t, 14) }

// DrawArchDirect stamps the direct-arch layer.
func (s *rasterSink) DrawArchDirect(t render.DrawTarget, _ int8) { s.stamp(t, 15) }

// DrawBlackTile stamps the town OOB tile.
func (s *rasterSink) DrawBlackTile(at render.PixelPos) {
	s.stamp(render.DrawTarget{At: at}, 16)
}

// stamp marks a three-pixel run at the draw position with the layer's index.
func (s *rasterSink) stamp(target render.DrawTarget, index uint8) {
	for i := range 3 {
		x, y := target.At.X+i, target.At.Y
		if x < 0 || y < 0 || x >= s.surface.Width || y >= s.surface.Height {
			continue
		}

		s.surface.Set(x, y, index)
	}
}

// loadPalette reads the committed Cathedral palette dump the tests load.
func loadPalette() assets.Palette {
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "render", paletteDir, paletteFixture))
	if err != nil {
		panic(fmt.Errorf("paritytest: load %s: %w", paletteFixture, err))
	}

	pal, err := assets.LoadPaletteFromBytes(body)
	if err != nil {
		panic(fmt.Errorf("paritytest: decode %s: %w", paletteFixture, err))
	}

	return *pal
}
