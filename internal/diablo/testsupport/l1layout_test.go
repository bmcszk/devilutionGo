package testsupport_test

// Pins for the L1 layout loader (devilGo-sj5a): the walk's dungeon stage must run
// on the committed cathedral layout the l1_parity gate pins, not a synthetic
// grid. The assertions read the fixture itself, so a corpus refresh that moved
// the layout would fail here rather than silently changing the walk.

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// The walk's cathedral level and the pinned scenario seed.
const (
	l1Level        = 1
	l1ScenarioSeed = 2588
)

// The SOL record bits the cathedral patch table writes (levels/dun_tile.hpp:106-114).
// The grid's property word drops BlockLight, so the patch pin reads the record
// byte itself.
const (
	solBlockLight   byte = 1 << 1
	solBlockMissile byte = 1 << 2
)

func TestLoadL1Layout_ClassifiesRealWallsAndFloors(t *testing.T) {
	// given the committed cathedral level 1 layout at the scenario seed
	layout := testsupport.LoadL1Layout(t, l1Level, l1ScenarioSeed)

	// when the piece layer is counted by classification
	solid, walkable := countClasses(layout.Image)

	// then the layout is a dungeon, not an empty room: both kinds of tile exist
	if solid == 0 {
		t.Fatal("the loaded L1 layout has no solid tile, want real walls")
	}

	if walkable == 0 {
		t.Fatal("the loaded L1 layout has no walkable tile, want real floors")
	}
}

func TestLoadL1Layout_EntryIsAWalkableTile(t *testing.T) {
	// given
	layout := testsupport.LoadL1Layout(t, l1Level, l1ScenarioSeed)

	// when the entry tile is classified through the SOL record
	entryWalkable := layout.EntryWalkable

	// then the level's own ViewPosition is where the player may arrive
	if !entryWalkable {
		t.Fatalf("entry %v is a solid tile, want the level's own walkable ViewPosition", layout.Entry)
	}
}

func TestLoadL1Layout_CathedralPatchesLandOnTheirSlots(t *testing.T) {
	// given the loaded layout: the retail l1.sol record with the cathedral hand
	// patches applied (dun_tile_data.cpp:109-133)
	layout := testsupport.LoadL1Layout(t, l1Level, l1ScenarioSeed)

	// when an arched slot, a light-only slot and an unpatched solid slot are read
	arched := layout.Image.SOL[9]
	lightOnly := layout.Image.SOL[403]
	unpatched := layout.Image.SOL[47]

	// then the arched slot carries both flags the arched-tile patch ORs in
	if arched&(solBlockLight|solBlockMissile) != solBlockLight|solBlockMissile {
		t.Fatalf("SOL[9] = %#x, want BlockLight|BlockMissile from the arched-tile patch", arched)
	}

	// and the patch reaches the grid word the predicates read
	if got := world.SOLTileProperties(arched); got&world.TileBlockMissile == 0 {
		t.Fatalf("SOL[9] = %#x maps to %#x, want BlockMissile on the grid word", arched, got)
	}

	// and the wall slot upstream patches with BlockLight alone gains only that bit
	if lightOnly&solBlockLight == 0 {
		t.Fatalf("SOL[403] = %#x, want BlockLight from the wall-tile patch", lightOnly)
	}

	// and a solid slot no patch names keeps the retail record's byte
	if unpatched != 0x29 {
		t.Fatalf("SOL[47] = %#x, want the retail 0x29: no cathedral patch touches that slot", unpatched)
	}
}

func TestLoadL1Layout_WallsBlockAStepOnTheRealGrid(t *testing.T) {
	// given a session standing on the loaded level, at the first walkable tile
	// that has a solid cardinal neighbour
	layout := testsupport.LoadL1Layout(t, l1Level, l1ScenarioSeed)

	origin, blocked := walkableWithSolidNeighbour(t, layout.Image)

	state := world.NewState()
	if _, ok := state.LoadLevel(layout.Image, origin, l1Level, 1); !ok {
		t.Fatal("load the L1 layout: no walkable tile at the probe window")
	}

	// when the step onto that solid neighbour is attempted
	stepped := state.StepTo(blocked)

	// then the real grid refuses it: a synthetic all-walkable grid never would
	if stepped {
		t.Fatalf("step from %v onto the solid tile %v was allowed, want real walls to block it", origin, blocked)
	}
}

// walkableWithSolidNeighbour finds a walkable tile and the solid cardinal
// neighbour that a step from it would target.
func walkableWithSolidNeighbour(t *testing.T, image *world.LevelLayout) (origin, target world.Tile) {
	t.Helper()

	for y := 1; y < world.MAXDUNY-1; y++ {
		for x := 1; x < world.MAXDUNX-1; x++ {
			if origin, target, ok := wallPairAt(image, x, y); ok {
				return origin, target
			}
		}
	}

	t.Fatal("the loaded L1 layout has no solid tile next to a walkable one, want real walls")

	return origin, target
}

// wallPairAt reports the walkable/solid cardinal pair anchored at (x, y).
func wallPairAt(image *world.LevelLayout, x, y int) (origin, target world.Tile, ok bool) {
	origin = world.Tile{X: x, Y: y}
	if solid(image, origin) {
		return origin, target, false
	}

	for _, step := range [][2]int{{1, 0}, {0, 1}} {
		target := world.Tile{X: x + step[0], Y: y + step[1]}
		if solid(image, target) {
			return origin, target, true
		}
	}

	return origin, target, false
}

// solid classifies one tile through the image's own SOL record.
func solid(image *world.LevelLayout, tile world.Tile) bool {
	piece := image.Piece[tile.X*world.MAXDUNY+tile.Y]

	return world.SOLTileProperties(image.SOL[piece])&world.TileSolid != 0
}

// countClasses counts the image's tiles by classification, both in tile order.
func countClasses(image *world.LevelLayout) (solidCount, walkable int) {
	for y := range world.MAXDUNY {
		for x := range world.MAXDUNX {
			if solid(image, world.Tile{X: x, Y: y}) {
				solidCount++

				continue
			}

			walkable++
		}
	}

	return solidCount, walkable
}
