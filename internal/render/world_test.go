package render_test

import (
	"hash/fnv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
)

// sceneHeight is the world-tile height of every scripted scene.
const sceneHeight = 40

// sceneTile is one scripted world tile: the grids under one dPiece coordinate.
type sceneTile struct {
	floor       bool
	wall        bool
	notSolid    bool
	transparent bool
	deadPlayer  bool
	object      bool
	objectPre   bool
	light       uint8
	piece       uint16
	special     int8
	corpse      int8
	item        int8
	player      *render.Occupant
	monster     *render.Occupant
}

// scene is the scripted dungeon a test hands to the renderer as a render.World.
type scene struct {
	tiles      map[render.TilePos]sceneTile
	postItems  map[int8]bool
	width      int
	height     int
	town       bool
	missilePre bool
}

// traceSink records one label per layer draw, in call order.
type traceSink struct{ calls []string }

// indexSink paints every layer draw into an index surface: a later draw
// overwrites an earlier one, so the hash of the buffer pins the draw order as
// well as the geometry.
type indexSink struct{ surface render.Surface }

// TestDrawDungeonDrawsLayersInUpstreamOrder pins the twelve-layer order of
// DrawDungeon (Source/engine/render/scrollrt.cpp:782-953): cell, missiles(pre),
// corpse, object(pre), item(pre), dead player, player, missile(post),
// object(post), item(post), arch.
func TestDrawDungeonDrawsLayersInUpstreamOrder(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.missilePre = true
	world.set(tile, sceneTile{
		light:      5,
		special:    2,
		corpse:     3,
		item:       4,
		object:     true,
		objectPre:  true,
		floor:      true,
		notSolid:   true,
		deadPlayer: true,
		player:     &render.Occupant{ID: 0, Encoded: 1},
	})
	sink := &traceSink{}
	at := render.PixelPos{X: 100, Y: 100}

	// when
	render.DrawDungeon(sink, world.world(), tile, at)

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(100, 100)",
		"missilePre (20, 20)@(100, 100)",
		"corpse (20, 20)@(100, 100)",
		"objectPre (20, 20)@(100, 100)",
		"itemPre (20, 20)@(100, 100)",
		"deadPlayer (20, 20)@(100, 100)",
		"player (20, 20)@(100, 100)",
		"missilePost (20, 20)@(100, 100)",
		"arch (20, 20)@(100, 100)",
	})
}

// TestDrawDungeonDrawsMonsterBelowPostDrawItem pins the monster-before-post-item
// half of the order: the monster paints before the post-draw items and the arch
// (scrollrt.cpp:864-908).
func TestDrawDungeonDrawsMonsterBelowPostDrawItem(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.postItems[5] = true
	world.set(tile, sceneTile{
		item:    5,
		special: 1,
		monster: &render.Occupant{ID: 3, Encoded: 4},
	})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(0, 0)",
		"monster (20, 20)@(0, 0)",
		"missilePost (20, 20)@(0, 0)",
		"itemPost (20, 20)@(0, 0)",
		"arch (20, 20)@(0, 0)",
	})
}

// TestDrawDungeonSkipsUnlitObjectsAndCorpses pins the dLight gate: at or above
// LightsMax a tile draws neither corpse nor object (scrollrt.cpp:801, :814).
func TestDrawDungeonSkipsUnlitObjectsAndCorpses(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.set(tile, sceneTile{light: 15, corpse: 1, object: true})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(0, 0)",
		"missilePost (20, 20)@(0, 0)",
	})
}

// TestDrawDungeonDrawsBlendedArchForTransparentTile pins the arch transparency
// branch (scrollrt.cpp:910-943): a transparent region blends the arch cel.
func TestDrawDungeonDrawsBlendedArchForTransparentTile(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.set(tile, sceneTile{special: 1, transparent: true})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(0, 0)",
		"missilePost (20, 20)@(0, 0)",
		"archBlended (20, 20)@(0, 0)",
	})
}

// TestDrawDungeonDefersTownLeavesOneRow pins the town branch: tree leaves are
// drawn from the north-west neighbour one tile up, direct and unlit
// (scrollrt.cpp:944-952).
func TestDrawDungeonDefersTownLeavesOneRow(t *testing.T) {
	// given
	world := newScene(40)
	world.town = true
	world.set(render.TilePos{X: 19, Y: 19}, sceneTile{special: 1})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), render.TilePos{X: 20, Y: 20}, render.PixelPos{X: 100, Y: 100})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(100, 100)",
		"missilePost (20, 20)@(100, 100)",
		"archDirect (19, 19)@(100, 68)",
	})
}

// TestDrawDungeonDrawsSouthwardsPlayerFromDepartedTile pins the negative-id
// priority: a player walking south is registered on its destination tile but
// drawn from the tile it is leaving, half a screen tile back
// (scrollrt.cpp:827-862).
func TestDrawDungeonDrawsSouthwardsPlayerFromDepartedTile(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.set(tile, sceneTile{
		player: &render.Occupant{ID: 1, Encoded: -2, Mode: render.ModeMoveSouthwards, Dir: render.South},
	})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{X: 100, Y: 100})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(100, 100)",
		"player (19, 19)@(100, 68)",
		"missilePost (20, 20)@(100, 100)",
	})
}

// TestDrawDungeonDrawsEastwalkingPlayerBelowStaticObject pins the y-sort of an
// eastwards walking player against a static object on its destination tile: the
// player paints from its departed tile, so the object's later draw covers it
// (scrollrt.cpp:827-862, :903-905).
func TestDrawDungeonDrawsEastwalkingPlayerBelowStaticObject(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.set(tile, sceneTile{
		object: true,
		player: &render.Occupant{ID: 0, Encoded: -1, Mode: render.ModeMoveSideways, Dir: render.East},
	})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{X: 100, Y: 100})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(100, 100)",
		"player (19, 21)@(36, 100)",
		"missilePost (20, 20)@(100, 100)",
		"objectPost (20, 20)@(100, 100)",
	})
}

// TestDrawDungeonSkipsPlayerWhoseGridIDDisagrees pins the second half of the
// priority trick: a southwards walker registered with a positive dPlayer id is
// not drawn at all (scrollrt.cpp:834-836).
func TestDrawDungeonSkipsPlayerWhoseGridIDDisagrees(t *testing.T) {
	// given
	world := newScene(40)
	tile := render.TilePos{X: 20, Y: 20}
	world.set(tile, sceneTile{
		player: &render.Occupant{ID: 0, Encoded: 1, Mode: render.ModeMoveSouthwards, Dir: render.South},
	})
	sink := &traceSink{}

	// when
	render.DrawDungeon(sink, world.world(), tile, render.PixelPos{})

	// then
	expectCalls(t, sink.calls, []string{
		"cell (20, 20)@(0, 0)",
		"missilePost (20, 20)@(0, 0)",
	})
}

// TestDrawFloorWalksTheIsoDiamond pins the shared row walk: one row runs east
// (x+1, y-1) per column, then the odd rows drop the column count and shift the
// cursor half a screen tile (scrollrt.cpp:955-981).
func TestDrawFloorWalksTheIsoDiamond(t *testing.T) {
	// given
	world := newScene(40)
	for x := range 40 {
		for y := range 40 {
			world.set(render.TilePos{X: x, Y: y}, sceneTile{floor: true})
		}
	}
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: 10, Y: 10}, Rows: 2, Columns: 2}

	// when
	render.DrawFloor(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{
		"floor (10, 10)@(0, 0)",
		"floor (11, 9)@(64, 0)",
		"floor (10, 11)@(-32, 16)",
		"floor (11, 10)@(32, 16)",
		"floor (12, 9)@(96, 16)",
	})
}

// TestDrawFloorSkipsOutOfBoundsTiles pins InDungeonBounds: the walk visits
// positions outside the grid and skips them (scrollrt.cpp:959-963).
func TestDrawFloorSkipsOutOfBoundsTiles(t *testing.T) {
	// given
	world := newScene(12)
	for y := range 40 {
		world.set(render.TilePos{X: 11, Y: y}, sceneTile{floor: true})
	}
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: 11, Y: 5}, Rows: 2, Columns: 2}

	// when
	render.DrawFloor(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{
		"floor (11, 5)@(0, 0)",
		"floor (11, 6)@(-32, 16)",
	})
}

// TestDrawTileContentDrawsWallBehindFirst pins the wall pre-pass: a wall aligned
// with its eastern neighbour that has walkable floor behind it draws the tile
// one east first, then itself, and the next tile's own draw is skipped
// (scrollrt.cpp:1009-1024).
func TestDrawTileContentDrawsWallBehindFirst(t *testing.T) {
	// given
	world := newScene(40)
	world.set(render.TilePos{X: 10, Y: 10}, sceneTile{wall: true})
	world.set(render.TilePos{X: 11, Y: 10}, sceneTile{wall: true})
	world.set(render.TilePos{X: 11, Y: 9}, sceneTile{notSolid: true})
	world.set(render.TilePos{X: 10, Y: 9}, sceneTile{notSolid: true})
	sink := &traceSink{}
	view := render.View{
		Start:       render.TilePos{X: 10, Y: 10},
		Origin:      render.PixelPos{X: 100, Y: 100},
		Rows:        1,
		Columns:     3,
		ScreenWidth: 640,
	}

	// when
	render.DrawTileContent(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{
		"cell (11, 9)@(164, 100)",
		"missilePost (11, 9)@(164, 100)",
		"cell (10, 10)@(100, 100)",
		"missilePost (10, 10)@(100, 100)",
		"cell (12, 8)@(228, 100)",
		"missilePost (12, 8)@(228, 100)",
	})
}

// TestDrawTileContentAddsMicroTileRows pins the extra rows tall walls need:
// DrawTileContent walks MicroTileLen rows past the view (scrollrt.cpp:995).
func TestDrawTileContentAddsMicroTileRows(t *testing.T) {
	// given
	world := newScene(40)
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: 10, Y: 10}, Rows: 1, Columns: 1, MicroTiles: 2}

	// when
	render.DrawTileContent(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{
		"cell (10, 10)@(0, 0)",
		"missilePost (10, 10)@(0, 0)",
		"cell (10, 11)@(-32, 16)",
		"missilePost (10, 11)@(-32, 16)",
		"cell (11, 10)@(32, 16)",
		"missilePost (11, 10)@(32, 16)",
		"cell (11, 11)@(0, 32)",
		"missilePost (11, 11)@(0, 32)",
	})
}

// TestDrawOOBPaintsDirtPatternTile pins the out-of-bounds fallback: an OOB tile
// draws the dirt pattern cell matching its grid parity on both axes
// (scrollrt.cpp:1047-1078).
func TestDrawOOBPaintsDirtPatternTile(t *testing.T) {
	// given
	world := newScene(40)
	world.set(render.TilePos{X: 1, Y: 1}, sceneTile{floor: true, piece: 5, light: 7})
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: -1, Y: -1}, Rows: -4, Columns: 1}

	// when
	render.DrawOOB(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{"cell (1, 1)@(0, 0)"})
}

// TestDrawOOBFallsBackToBlackTile pins the dirt sampler's failsafe: an empty
// pattern cell draws a black tile (scrollrt.cpp:1059-1062).
func TestDrawOOBFallsBackToBlackTile(t *testing.T) {
	// given
	world := newScene(40)
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: -1, Y: -1}, Rows: -4, Columns: 1}

	// when
	render.DrawOOB(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{"black (0, 0)"})
}

// TestDrawOOBFillsTownGapsWithBlackTiles pins the town branch: town has no dirt
// fallback, so every OOB tile is black (scrollrt.cpp:1093-1097).
func TestDrawOOBFillsTownGapsWithBlackTiles(t *testing.T) {
	// given
	world := newScene(40)
	world.town = true
	world.set(render.TilePos{X: 1, Y: 1}, sceneTile{floor: true, piece: 5})
	sink := &traceSink{}
	view := render.View{Start: render.TilePos{X: -1, Y: -1}, Rows: -4, Columns: 1}

	// when
	render.DrawOOB(sink, world.world(), view)

	// then
	expectCalls(t, sink.calls, []string{"black (0, 0)"})
}

// TestWorldToScreenMatchesUpstreamMatrix pins the isometric projection
// (Source/engine/displacement.hpp:151-155): the -135° rotation scaled to a
// 64×32 tile.
func TestWorldToScreenMatchesUpstreamMatrix(t *testing.T) {
	// given
	cases := []struct {
		dx, dy int
		want   render.PixelPos
	}{
		{dx: 0, dy: 0, want: render.PixelPos{X: 0, Y: 0}},
		{dx: 1, dy: 0, want: render.PixelPos{X: -32, Y: -16}},
		{dx: 0, dy: 1, want: render.PixelPos{X: 32, Y: -16}},
		{dx: 2, dy: 0, want: render.PixelPos{X: -64, Y: -32}},
		{dx: -3, dy: 4, want: render.PixelPos{X: 224, Y: -16}},
	}

	for _, row := range cases {
		// when
		got := render.WorldToScreen(row.dx, row.dy)

		// then
		if got != row.want {
			t.Fatalf("worldToScreen(%d, %d) = %v, want %v", row.dx, row.dy, got, row.want)
		}
	}
}

// TestWalkOffsetMatchesUpstreamTable pins the walking camera offset table, its
// fixed-point progress scale and the camera-mode negation
// (scrollrt.cpp:1575-1597).
func TestWalkOffsetMatchesUpstreamTable(t *testing.T) {
	// given
	cases := []struct {
		dir      render.Direction
		progress int
		want     render.PixelPos
	}{
		{dir: render.South, progress: 128, want: render.PixelPos{X: 0, Y: -32}},
		{dir: render.South, progress: 0, want: render.PixelPos{X: 0, Y: 0}},
		{dir: render.South, progress: 64, want: render.PixelPos{X: 0, Y: -16}},
		{dir: render.East, progress: 128, want: render.PixelPos{X: -64, Y: 0}},
		{dir: render.East, progress: 64, want: render.PixelPos{X: -32, Y: 0}},
		{dir: render.North, progress: 128, want: render.PixelPos{X: 0, Y: 32}},
		{dir: render.SouthWest, progress: 64, want: render.PixelPos{X: 16, Y: -8}},
	}

	for _, row := range cases {
		// when
		got := render.WalkOffset(row.dir, row.progress)

		// then
		if got != row.want {
			t.Fatalf("WalkOffset(%d, %d) = %v, want %v", row.dir, row.progress, got, row.want)
		}
	}
}

// TestDrawWorldMatchesGoldenIndexHash pins the whole frame walk — floor, tile
// content and OOB — as one index-buffer hash at three animation progresses of a
// seeded scene, so any change in order, geometry or layer selection shows up.
func TestDrawWorldMatchesGoldenIndexHash(t *testing.T) {
	// given
	world, view := seededScene()
	want := map[int]uint64{0: 0xd3ca1c79f13f69b6, 64: 0x90d157a95cf0c46d, 128: 0x8107c5511ed6566d}

	for progress, wantHash := range want {
		sink := newIndexSink(640, 480)
		walked := view
		walked.Origin = walked.Origin.Add(render.WalkOffset(render.SouthEast, progress))

		// when
		render.DrawWorld(sink, world, walked)

		// then
		if got := sink.hash(); got != wantHash {
			t.Fatalf("progress %d: index hash %#x, want %#x", progress, got, wantHash)
		}
	}
}

// TestDrawWorldAllocatesNothing pins the software world path as allocation-free:
// the per-frame walk runs on preallocated state only.
func TestDrawWorldAllocatesNothing(t *testing.T) {
	// given
	world, view := seededScene()
	sink := newIndexSink(640, 480)

	// when
	allocs := testing.AllocsPerRun(20, func() { render.DrawWorld(sink, world, view) })

	// then
	if allocs != 0 {
		t.Fatalf("DrawWorld allocates %v times per frame, want 0", allocs)
	}
}

// seededScene builds a deterministic 9×7 view with every layer populated, so the
// golden hash covers all twelve layers of the tile order.
func seededScene() (render.World, render.View) {
	world := newScene(40)
	world.missilePre = true
	world.postItems[2] = true
	seedTiles(world)
	seedOccupants(world)

	return world.world(), render.View{
		Start:       render.TilePos{X: 20, Y: 20},
		Origin:      render.PixelPos{X: 320, Y: 200},
		Rows:        7,
		Columns:     9,
		ScreenWidth: 640,
		MicroTiles:  4,
	}
}

// seedTiles paints the deterministic mixed terrain of the golden scene.
func seedTiles(world *scene) {
	for x := 15; x < 30; x++ {
		for y := 15; y < 30; y++ {
			world.set(render.TilePos{X: x, Y: y}, seededTile(x, y))
		}
	}
}

// seedOccupants puts one of every entity on the view's anchor tile.
func seedOccupants(world *scene) {
	world.set(render.TilePos{X: 20, Y: 20}, sceneTile{
		floor:   true,
		light:   3,
		special: 1,
		corpse:  2,
		item:    2,
		object:  true,
		player:  &render.Occupant{ID: 0, Encoded: -1, Mode: render.ModeMoveSouthwards, Dir: render.SouthEast},
		monster: &render.Occupant{ID: 0, Encoded: 1},
	})
}

// seededTile derives one tile's state from its coordinates.
func seededTile(x, y int) sceneTile {
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
		state.object, state.objectPre = true, true
	}

	return state
}

// newScene builds an empty scene of width tiles, sceneHeight tall.
func newScene(width int) *scene {
	return &scene{
		tiles:     make(map[render.TilePos]sceneTile),
		postItems: make(map[int8]bool),
		width:     width,
		height:    sceneHeight,
	}
}

// newIndexSink allocates the surface the sink paints into.
func newIndexSink(width, height int) *indexSink {
	return &indexSink{surface: render.NewSurface(width, height)}
}

// expectCalls asserts the recorded layer sequence.
func expectCalls(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("draw sequence has %d calls, want %d\n got: %v\nwant: %v", len(got), len(want), got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("draw call %d = %q, want %q\n got: %v\nwant: %v", i, got[i], want[i], got, want)
		}
	}
}

// set scripts one tile.
func (s *scene) set(tile render.TilePos, value sceneTile) { s.tiles[tile] = value }

// world wraps the scene as the renderer's injected grid view.
func (s *scene) world() render.World {
	return render.World{
		InBounds: func(tile render.TilePos) bool {
			return tile.X >= 0 && tile.Y >= 0 && tile.X < s.width && tile.Y < s.height
		},
		IsFloor:      func(tile render.TilePos) bool { return s.at(tile).floor },
		IsWall:       func(tile render.TilePos) bool { return s.at(tile).wall },
		NotSolid:     func(tile render.TilePos) bool { return s.at(tile).notSolid },
		Light:        func(tile render.TilePos) uint8 { return s.at(tile).light },
		Piece:        func(tile render.TilePos) uint16 { return s.at(tile).piece },
		Special:      func(tile render.TilePos) int8 { return s.at(tile).special },
		Transparent:  func(tile render.TilePos) bool { return s.at(tile).transparent },
		Corpse:       func(tile render.TilePos) int8 { return s.at(tile).corpse },
		Item:         func(tile render.TilePos) int8 { return s.at(tile).item },
		ItemPostDraw: func(item int8) bool { return s.postItems[item] },
		DeadPlayer:   func(tile render.TilePos) bool { return s.at(tile).deadPlayer },
		Object: func(tile render.TilePos) (render.ObjectRef, bool) {
			value := s.at(tile)

			return render.ObjectRef{Pre: value.objectPre}, value.object
		},
		Player: func(tile render.TilePos) (render.Occupant, bool) {
			if occupant := s.at(tile).player; occupant != nil {
				return *occupant, true
			}

			return render.Occupant{}, false
		},
		Monster: func(tile render.TilePos) (render.Occupant, bool) {
			if occupant := s.at(tile).monster; occupant != nil {
				return *occupant, true
			}

			return render.Occupant{}, false
		},
		MissilePre: s.missilePre,
		Town:       s.town,
	}
}

// at returns tile's scripted state, the zero value outside the script.
func (s *scene) at(tile render.TilePos) sceneTile { return s.tiles[tile] }

func (s *traceSink) DrawFloorTile(t render.DrawTarget) { s.record("floor", t) }

func (s *traceSink) DrawCell(t render.DrawTarget) { s.record("cell", t) }

func (s *traceSink) DrawMissilePre(t render.DrawTarget) { s.record("missilePre", t) }

func (s *traceSink) DrawCorpse(t render.DrawTarget) { s.record("corpse", t) }

func (s *traceSink) DrawObjectPre(t render.DrawTarget) { s.record("objectPre", t) }

func (s *traceSink) DrawItemPre(t render.DrawTarget) { s.record("itemPre", t) }

func (s *traceSink) DrawDeadPlayer(t render.DrawTarget) { s.record("deadPlayer", t) }

func (s *traceSink) DrawPlayer(t render.DrawTarget, _ int) { s.record("player", t) }

func (s *traceSink) DrawMonster(t render.DrawTarget, _ int) { s.record("monster", t) }

func (s *traceSink) DrawItemPost(t render.DrawTarget) { s.record("itemPost", t) }

func (s *traceSink) DrawObjectPost(t render.DrawTarget) { s.record("objectPost", t) }

func (s *traceSink) DrawMissilePost(t render.DrawTarget) { s.record("missilePost", t) }

func (s *traceSink) DrawArch(t render.DrawTarget, _ int8) { s.record("arch", t) }

func (s *traceSink) DrawArchBlended(t render.DrawTarget, _ int8) { s.record("archBlended", t) }

func (s *traceSink) DrawArchDirect(t render.DrawTarget, _ int8) { s.record("archDirect", t) }

func (s *traceSink) DrawBlackTile(at render.PixelPos) {
	s.calls = append(s.calls, "black "+at.String())
}

// record appends one labelled draw target.
func (s *traceSink) record(label string, target render.DrawTarget) {
	s.calls = append(s.calls, label+" "+target.Tile.String()+"@"+target.At.String())
}

func (s *indexSink) DrawFloorTile(t render.DrawTarget) { s.stamp(t, 1) }

func (s *indexSink) DrawCell(t render.DrawTarget) { s.stamp(t, 2) }

func (s *indexSink) DrawMissilePre(t render.DrawTarget) { s.stamp(t, 3) }

func (s *indexSink) DrawCorpse(t render.DrawTarget) { s.stamp(t, 4) }

func (s *indexSink) DrawObjectPre(t render.DrawTarget) { s.stamp(t, 5) }

func (s *indexSink) DrawItemPre(t render.DrawTarget) { s.stamp(t, 6) }

func (s *indexSink) DrawDeadPlayer(t render.DrawTarget) { s.stamp(t, 7) }

func (s *indexSink) DrawPlayer(t render.DrawTarget, _ int) { s.stamp(t, 8) }

func (s *indexSink) DrawMonster(t render.DrawTarget, _ int) { s.stamp(t, 9) }

func (s *indexSink) DrawItemPost(t render.DrawTarget) { s.stamp(t, 10) }

func (s *indexSink) DrawObjectPost(t render.DrawTarget) { s.stamp(t, 11) }

func (s *indexSink) DrawMissilePost(t render.DrawTarget) { s.stamp(t, 12) }

func (s *indexSink) DrawArch(t render.DrawTarget, _ int8) { s.stamp(t, 13) }

func (s *indexSink) DrawArchBlended(t render.DrawTarget, _ int8) { s.stamp(t, 14) }

func (s *indexSink) DrawArchDirect(t render.DrawTarget, _ int8) { s.stamp(t, 15) }

func (s *indexSink) DrawBlackTile(at render.PixelPos) {
	s.stamp(render.DrawTarget{At: at}, 16)
}

// stamp marks a three-pixel run at the draw position with the layer's index.
func (s *indexSink) stamp(target render.DrawTarget, index uint8) {
	for i := range 3 {
		x, y := target.At.X+i, target.At.Y
		if x < 0 || y < 0 || x >= s.surface.Width || y >= s.surface.Height {
			continue
		}

		s.surface.Set(x, y, index)
	}
}

// hash is the FNV-1a digest of the painted index buffer.
func (s *indexSink) hash() uint64 {
	digest := fnv.New64a()
	digest.Write(s.surface.Pixels)

	return digest.Sum64()
}
