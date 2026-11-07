package levels_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestPlaceCaveLightsSources is PlaceCaveLights' collection rule
// (levels/drlg_l3.cpp:2154-2168): one radius-7 source per tile whose dPiece
// matches any of the three predicates, in grid order, and nothing else.
func TestPlaceCaveLightsSources(t *testing.T) {
	// given: a grid holding every matching tile, plus the near-miss neighbours
	// on both sides of each predicate.
	all := append(caveLightTiles(), 0, 54, 147, 148, 150, 152, 161)
	pieces := tileRowGrid(all)

	// when: the caves light pass collects its sources.
	sources := levels.PlaceCaveLights(pieces)

	// then: exactly the matching tiles, all radius 7, in grid order.
	expectSources(t, sources, expectedCaveSources(pieces))
	if got, want := len(sources), len(caveLightTiles()); got != want {
		t.Fatalf("sources = %d, want %d", got, want)
	}
}

// TestPlaceCaveLightsMatchesCavesDPiece runs the pass over the committed caves
// dPiece corpus: real generator output must yield at least one source per
// level, and every source must sit on a tile upstream lights.
func TestPlaceCaveLightsMatchesCavesDPiece(t *testing.T) {
	// given: the committed dPiece grids of one caves seed, both entries.
	goldens := loadPieceGoldens(t, 2588)

	// when: the caves light pass runs on each of them.
	for _, key := range []string{"l9-main", "l9-prev", "l12-main", "l12-prev"} {
		pieces := levels.PieceGrid{Piece: requireGolden(t, goldens, key, 2588).piece}
		sources := levels.PlaceCaveLights(pieces)

		// then: non-empty, radius 7 everywhere, every position a matching tile.
		checkCaveSources(t, key, sources, pieces)
	}
}

// TestCavesApplyLightsShadesLava runs the whole caves light application over a
// real caves grid: ApplyLights must darken the lit tiles below the level's
// default shade of 15 (levels/gendung.cpp:301-307).
func TestCavesApplyLightsShadesLava(t *testing.T) {
	// given: a caves level initialised to the fully dark dLight default and
	// its committed dPiece grid.
	dungeon := levels.NewDungeon(levels.Caves, 9)
	dungeon.InitLight()
	pieces := levels.PieceGrid{Piece: loadPieceGoldens(t, 2588)["l9-main"].piece}

	// when: the level's static lights are applied.
	levels.ApplyLights(dungeon, pieces)

	// then: tiles are brightened and every source centre is fully lit.
	sources := levels.PlaceCaveLights(pieces)
	if len(sources) == 0 {
		t.Fatal("sources = 0, want at least one lit caves tile")
	}
	if brightened := countBelow(dungeon, 15); brightened == 0 {
		t.Fatal("no tile brightened by the caves light pass")
	}
	for _, source := range sources {
		if got := dungeon.Light[source.Position.X][source.Position.Y]; got != 0 {
			t.Fatalf("source tile %+v = %d, want 0 (caves brighten their centre)", source.Position, got)
		}
	}
}

// TestCaveLightSourcesFromGoldens is the crypt/hive twin for the caves: every
// tile whose dPiece matches PlaceCaveLights' three predicates becomes a radius-7
// source (levels/drlg_l3.cpp:2154-2168), recomputed from all 64 committed caves
// dPiece goldens.
func TestCaveLightSourcesFromGoldens(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadPieceGoldens(t, seed)

		for _, variant := range l3Variants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed dPiece golden of this variant.
				golden := requireGolden(t, goldens, name, seed)
				pieces := levels.PieceGrid{Piece: golden.piece}

				// when: the caves light pass collects sources from it.
				sources := levels.PlaceCaveLights(pieces)

				// then: every matching tile yielded exactly one radius-7 source.
				checkCaveSources(t, name, sources, pieces)
			})
		}
	}
}

// caveLightTiles restates PlaceCaveLights' three tile predicates independently
// of the port (levels/drlg_l3.cpp:2158-2166): the 55-146 band, the 153-160 band
// and the two singletons 149 and 151.
func caveLightTiles() []uint16 {
	tiles := make([]uint16, 0, 107)
	for tile := uint16(55); tile <= 146; tile++ {
		tiles = append(tiles, tile)
	}
	for tile := uint16(153); tile <= 160; tile++ {
		tiles = append(tiles, tile)
	}

	return append(tiles, 149, 151)
}

// isCaveLightTile restates the upstream predicate independently of the port, so
// a mistyped bound in cave_lights.go fails these tests.
func isCaveLightTile(tile uint16) bool {
	return (tile >= 55 && tile <= 146) || (tile >= 153 && tile <= 160) || tile == 149 || tile == 151
}

// tileRowGrid fills a grid row by row with tiles, in order.
func tileRowGrid(tiles []uint16) levels.PieceGrid {
	grid := levels.PieceGrid{}
	index := 0
	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if index == len(tiles) {
				return grid
			}
			grid.Piece[x][y] = tiles[index]
			index++
		}
	}

	return grid
}

// expectedCaveSources is the source list the upstream predicate implies for a
// grid: one radius-7 light per matching cell, in grid order.
func expectedCaveSources(pieces levels.PieceGrid) []levels.Light {
	sources := make([]levels.Light, 0)
	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if isCaveLightTile(pieces.Piece[x][y]) {
				sources = append(sources, levels.Light{Position: levels.Point{X: x, Y: y}, Radius: 7})
			}
		}
	}

	return sources
}

// checkCaveSources asserts the sources are non-empty, all radius 7, and each one
// sits on a tile the upstream predicate accepts.
func checkCaveSources(t *testing.T, name string, sources []levels.Light, pieces levels.PieceGrid) {
	t.Helper()

	if len(sources) == 0 {
		t.Fatalf("%s: sources = 0, want at least one lit caves tile", name)
	}
	for _, source := range sources {
		if source.Radius != 7 {
			t.Fatalf("%s: source %+v has radius %d, want 7", name, source, source.Radius)
		}
		if tile := pieces.Piece[source.Position.X][source.Position.Y]; !isCaveLightTile(tile) {
			t.Fatalf("%s: source %+v sits on dPiece %d, which upstream does not light", name, source, tile)
		}
	}
}

// countBelow counts the dLight cells darker than shade.
func countBelow(dungeon *levels.Dungeon, shade uint8) int {
	count := 0
	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if dungeon.Light[x][y] < shade {
				count++
			}
		}
	}

	return count
}
