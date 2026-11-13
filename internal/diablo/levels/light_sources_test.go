package levels_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/light"
)

// upstreamLavaTiles is an independent copy of PlaceCryptLights' lava tile list
// (levels/crypt.cpp:819-826). It is kept here so a truncated or mistyped port of
// the ported table fails a test instead of silently emitting fewer lights.
var upstreamLavaTiles = []uint16{
	124, 128, 130, 132, 133, 134, 135, 139, 141, 143, 145, 156, 164, 166,
	167, 168, 169, 170, 182, 190, 192, 195, 196, 199, 200, 254, 266, 273,
	276, 281, 282, 283, 284, 285, 286, 287, 288, 290, 302, 316, 434, 435,
	436, 437, 445, 446, 447, 453, 457, 460, 466, 470, 477, 479, 484, 485,
	486, 490, 507, 537, 557, 559, 561, 563, 564, 568, 569, 572, 578, 580,
	584, 585, 589, 592, 593, 594, 595, 596, 597, 598, 599, 600, 601,
}

// TestCryptLavaTiles pins the ported lava tile table against upstream's list
// (levels/crypt.cpp:819-826): 83 entries, same order, no duplicates.
func TestCryptLavaTiles(t *testing.T) {
	// given: the ported table and a hand copy of the upstream literal.
	ported := levels.CryptLavaTilesForTest()

	// when: the ported table is read back.
	want := upstreamLavaTiles

	// then: same length, same values in upstream order.
	if len(ported) != len(want) {
		t.Fatalf("lava tiles = %d entries, want %d", len(ported), len(want))
	}
	for i, wantTile := range want {
		if ported[i] != wantTile {
			t.Fatalf("lavaTiles[%d] = %d, want %d", i, ported[i], wantTile)
		}
	}
}

// TestPlaceCryptLightsSources is PlaceCryptLights' collection rule
// (levels/crypt.cpp:828-837): one radius-3 source per tile whose dPiece is a
// lava tile, and nothing else.
func TestPlaceCryptLightsSources(t *testing.T) {
	// given: a grid with three lava tiles and two ordinary ones.
	pieces := levels.PieceGrid{}
	pieces.Piece[20][30] = 124
	pieces.Piece[20][31] = 601
	pieces.Piece[0][0] = 595
	pieces.Piece[20][32] = 123
	pieces.Piece[111][111] = 1

	// when: the crypt light pass collects its sources.
	sources := levels.PlaceCryptLights(pieces)

	// then: exactly the three lava tiles, in x-major order, all radius 3.
	expectSources(t, sources, []levels.Light{
		{Position: levels.Point{X: 0, Y: 0}, Radius: 3},
		{Position: levels.Point{X: 20, Y: 30}, Radius: 3},
		{Position: levels.Point{X: 20, Y: 31}, Radius: 3},
	})
}

// TestPlaceHiveLightsSources is PlaceHiveLights' collection rule
// (levels/drlg_l3.cpp:2165-2168): one radius-9 source per tile whose dPiece is
// in 381..456 inclusive, and nothing else.
func TestPlaceHiveLightsSources(t *testing.T) {
	// given: the band edges, one tile inside and one outside on each side.
	pieces := levels.PieceGrid{}
	pieces.Piece[10][10] = 381
	pieces.Piece[10][11] = 456
	pieces.Piece[10][12] = 400
	pieces.Piece[10][13] = 380
	pieces.Piece[10][14] = 457

	// when: the hive light pass collects its sources.
	sources := levels.PlaceHiveLights(pieces)

	// then: the three in-band tiles, all radius 9.
	expectSources(t, sources, []levels.Light{
		{Position: levels.Point{X: 10, Y: 10}, Radius: 9},
		{Position: levels.Point{X: 10, Y: 11}, Radius: 9},
		{Position: levels.Point{X: 10, Y: 12}, Radius: 9},
	})
}

// TestDoLightingBrightensCentre is the non-crypt branch of DoLighting
// (Source/lighting.cpp:157-159): the light's own tile is forced to full
// brightness, whatever it held before.
func TestDoLightingBrightensCentre(t *testing.T) {
	// given: a cathedral level, whose dLight starts fully dark.
	dungeon := levels.NewDungeon(levels.Cathedral, 1)
	dungeon.InitLight()
	centre := levels.Point{X: 40, Y: 40}

	// when: a radius-5 light is stamped on the centre tile.
	levels.DoLighting(dungeon, lightTablesFor(levels.Cathedral), centre, 5, levels.Offset{})

	// then: the centre is lit to 0 and its neighbours are dimmer than the
	// untouched default of 15.
	if got := dungeon.Light[centre.X][centre.Y]; got != 0 {
		t.Fatalf("centre light = %d, want 0", got)
	}
	if got := dungeon.Light[centre.X+2][centre.Y]; got >= 15 {
		t.Fatalf("neighbour light = %d, want < 15", got)
	}
}

// TestDoLightingDimCryptCentre is the crypt/nest branch of DoLighting
// (Source/lighting.cpp:153-156): the centre is only lowered to the radius-0
// falloff, never all the way to 0, so a crypt lava tile stays dim.
func TestDoLightingDimCryptCentre(t *testing.T) {
	// given: a crypt level whose centre tile is fully dark.
	dungeon := levels.NewDungeon(levels.Crypt, 21)
	dungeon.InitLight()
	centre := levels.Point{X: 40, Y: 40}
	tables := lightTablesFor(levels.Crypt)
	want := tables.Falloffs[5][0]
	dungeon.Light[centre.X][centre.Y] = 15

	// when: a radius-5 light is stamped on it.
	levels.DoLighting(dungeon, tables, centre, 5, levels.Offset{})

	// then: the centre holds the radius-0 falloff rather than 0.
	if got := dungeon.Light[centre.X][centre.Y]; got != want {
		t.Fatalf("centre light = %d, want %d", got, want)
	}
	if want == 0 {
		t.Fatal("crypt falloffs[5][0] = 0, the dim branch is untested")
	}
}

// TestDoLightingOnlyBrightens is DoLighting's min semantics
// (Source/lighting.cpp:173-175): the write is `if v < GetLight`, so a tile
// already brighter than the incoming falloff keeps its value. A lower light
// value means a brighter tile.
func TestDoLightingOnlyBrightens(t *testing.T) {
	// given: a cathedral level with one neighbour tile already fully lit.
	dungeon := levels.NewDungeon(levels.Cathedral, 1)
	dungeon.InitLight()
	centre := levels.Point{X: 40, Y: 40}
	dungeon.Light[centre.X+1][centre.Y] = 0

	// when: a radius-5 light covers it.
	levels.DoLighting(dungeon, lightTablesFor(levels.Cathedral), centre, 5, levels.Offset{})

	// then: it stays fully lit instead of taking the dimmer falloff.
	if got := dungeon.Light[centre.X+1][centre.Y]; got != 0 {
		t.Fatalf("pre-lit tile = %d, want 0", got)
	}
}

// TestDoLightingCornerStaysDark pins the cone bounds
// (Source/lighting.cpp:136-151): a light in the north-west dungeon corner
// floods no tile outside the grid and leaves the far corner untouched.
func TestDoLightingCornerStaysDark(t *testing.T) {
	// given: a level lit at the north-west corner.
	dungeon := levels.NewDungeon(levels.Cathedral, 1)
	dungeon.InitLight()

	// when: a radius-15 light sits on tile (0,0).
	levels.DoLighting(dungeon, lightTablesFor(levels.Cathedral), levels.Point{}, 15, levels.Offset{})

	// then: the corner itself is lit and the far south-east corner is untouched.
	if got := dungeon.Light[0][0]; got != 0 {
		t.Fatalf("corner light = %d, want 0", got)
	}
	if got := dungeon.Light[levels.MAXDUNX-1][levels.MAXDUNY-1]; got != 15 {
		t.Fatalf("far corner light = %d, want 15", got)
	}
}

// TestInitLightDefault is InitGlobals' dLight default (levels/gendung.cpp:301-307):
// town starts fully lit, every other level type fully dark.
func TestInitLightDefault(t *testing.T) {
	cases := []struct {
		name      string
		levelType levels.LevelType
		want      uint8
	}{
		{"town", levels.Town, 0},
		{"cathedral", levels.Cathedral, 15},
		{"crypt", levels.Crypt, 15},
		{"nest", levels.Nest, 15},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when: a fresh dungeon of this level type initialises its light grid.
			dungeon := levels.NewDungeon(tc.levelType, 1)
			dungeon.InitLight()

			// then: every tile holds the level type's default.
			expectUniformLight(t, dungeon, tc.want)
		})
	}
}

// TestApplyLightsStampsSources is the build-time wiring: ApplyLights lights the
// collected family sources into dLight, so a level ends up with lit tiles at
// the emitting tiles and only there.
func TestApplyLightsStampsSources(t *testing.T) {
	// given: a crypt level with one lava tile and one ordinary tile.
	dungeon := levels.NewDungeon(levels.Crypt, 21)
	dungeon.InitLight()
	pieces := levels.PieceGrid{}
	pieces.Piece[40][40] = 124

	// when: the level's static lights are applied.
	levels.ApplyLights(dungeon, pieces)

	// then: the lava tile and its neighbour are lit, a far tile is not.
	if got := dungeon.Light[40][40]; got >= 15 {
		t.Fatalf("lava tile light = %d, want < 15", got)
	}
	if got := dungeon.Light[41][40]; got >= 15 {
		t.Fatalf("lava neighbour light = %d, want < 15", got)
	}
	if got := dungeon.Light[60][40]; got != 15 {
		t.Fatalf("far tile light = %d, want 15", got)
	}
}

// TestApplyLightsIgnoresOtherLevels pins the family dispatch: a level type this
// port has no static light pass for leaves dLight untouched.
func TestApplyLightsIgnoresOtherLevels(t *testing.T) {
	// given: a cathedral level on a grid full of lava and nest light tiles.
	dungeon := levels.NewDungeon(levels.Cathedral, 1)
	dungeon.InitLight()
	pieces := levels.PieceGrid{}
	pieces.Piece[40][40] = 124
	pieces.Piece[41][41] = 400

	// when: the static lights are applied.
	levels.ApplyLights(dungeon, pieces)

	// then: nothing was lit.
	expectUniformLight(t, dungeon, 15)
}

// TestApplyLightsAfterRealGeneration drives the documented three-step level
// build (GenerateLevel, Pass3 = ExpandPieces, then the static light pass) on a
// real caves, crypt and nest level, and checks the result is a lit level rather
// than the fully dark InitGlobals grid.
func TestApplyLightsAfterRealGeneration(t *testing.T) {
	cases := []struct {
		name      string
		levelType levels.LevelType
		level     int
		family    string
		generate  func(*levels.Dungeon, levels.Options) error
	}{
		{"caves-9", levels.Caves, 9, "l3", levels.GenerateL3},
		{"crypt-21", levels.Crypt, 21, "l5", levels.GenerateL1},
		{"nest-17", levels.Nest, 17, "l6", levels.GenerateNest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when: the level is generated, expanded and its lights applied.
			dungeon, pieces := generateLitLevel(t, tc.levelType, tc.level, tc.family, tc.generate)

			// then: the level is lit, and only near collected sources.
			checkLitTilesAreNearSources(t, dungeon, pieces, tc.levelType)
		})
	}
}

// TestCryptLightSourcesFromGoldens recomputes the crypt light sources straight
// from the committed C++ probe dPiece goldens instead of from the Go generator:
// every tile whose dPiece is a lava id becomes a radius-3 source
// (levels/crypt.cpp:828-837). It pins the source count and radii of all 64
// crypt level/entry/seed variants without depending on the Go generator.
func TestCryptLightSourcesFromGoldens(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadPieceGoldens(t, seed)

		for _, variant := range cryptVariants {
			name := fmt.Sprintf("l%d-%s", variant.level, variant.name)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed dPiece golden of this variant.
				golden := requireGolden(t, goldens, name, seed)

				// when: the crypt light pass collects sources from it.
				sources := levels.PlaceCryptLights(levels.PieceGrid{Piece: golden.piece})

				// then: every lava tile yielded exactly one radius-3 source.
				checkCryptLightSources(t, sources, golden.piece)
			})
		}
	}
}

// TestHiveLightSourcesFromGoldens is the nest twin: every tile whose dPiece is
// in 381..456 becomes a radius-9 source (levels/drlg_l3.cpp:2165-2168),
// recomputed from the 64 committed nest dPiece goldens.
func TestHiveLightSourcesFromGoldens(t *testing.T) {
	for _, seed := range committedSeeds {
		goldens := loadPieceGoldens(t, seed)

		for _, variant := range hiveVariants {
			name := hiveFixtureName(variant)
			t.Run(fmt.Sprintf("%s-seed-%d", name, seed), func(t *testing.T) {
				// given: the committed dPiece golden of this variant.
				golden := requireGolden(t, goldens, name, seed)

				// when: the hive light pass collects sources from it.
				sources := levels.PlaceHiveLights(levels.PieceGrid{Piece: golden.piece})

				// then: every in-band tile yielded exactly one radius-9 source.
				checkHiveLightSources(t, sources, golden.piece)
			})
		}
	}
}

// generateLitLevel builds one level and applies its static light pass.
func generateLitLevel(
	t *testing.T, levelType levels.LevelType, level int, family string,
	generate func(*levels.Dungeon, levels.Options) error,
) (*levels.Dungeon, levels.PieceGrid) {
	t.Helper()

	dungeon := levels.NewDungeon(levelType, level)
	dungeon.Rnd.Seed(committedSeeds[0])
	dungeon.InitLight()

	if err := generate(dungeon, levels.Options{Entry: levels.EntryMain}); err != nil {
		t.Fatalf("generate: %v", err)
	}

	pieces := levels.ExpandPieces(dungeon, loadMegaTiles(t, family))
	levels.ApplyLights(dungeon, pieces)

	return dungeon, pieces
}

// checkLitTilesAreNearSources asserts every lit tile sits within 15 world tiles
// of a collected source, and that the pass lit something at all.
func checkLitTilesAreNearSources(
	t *testing.T, dungeon *levels.Dungeon, pieces levels.PieceGrid, levelType levels.LevelType,
) {
	t.Helper()

	sources := levels.PlaceHiveLights(pieces)
	if levelType == levels.Crypt {
		sources = levels.PlaceCryptLights(pieces)
	}
	if len(sources) == 0 {
		t.Fatal("no light sources collected")
	}

	lit := litTiles(dungeon)
	if len(lit) == 0 {
		t.Fatal("no tile lit")
	}

	for _, tile := range lit {
		if !nearAnySource(tile, sources, 15) {
			t.Fatalf("lit tile (%d,%d) is not near any source", tile.X, tile.Y)
		}
	}
}

// litTiles lists every tile dLight marks as brighter than the fully dark
// default.
func litTiles(dungeon *levels.Dungeon) []levels.Point {
	var lit []levels.Point

	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if dungeon.Light[x][y] != 15 {
				lit = append(lit, levels.Point{X: x, Y: y})
			}
		}
	}

	return lit
}

// lightTablesFor builds the level's light tables the way ApplyLights does.
func lightTablesFor(levelType levels.LevelType) light.LightTables {
	return light.MakeLightTables(levels.LevelLightType(levelType))
}

// expectSources compares a collected source list against the expected one,
// position and radius both.
func expectSources(t *testing.T, got, want []levels.Light) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("sources = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sources[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// expectUniformLight asserts every tile of the grid holds want.
func expectUniformLight(t *testing.T, dungeon *levels.Dungeon, want uint8) {
	t.Helper()

	for x := range levels.MAXDUNX {
		for y := range levels.MAXDUNY {
			if got := dungeon.Light[x][y]; got != want {
				t.Fatalf("Light[%d][%d] = %d, want %d", x, y, got, want)
			}
		}
	}
}

// requireGolden fetches one committed dPiece golden or fails the sub-test.
func requireGolden(t *testing.T, goldens map[string]*pieceGolden, name string, seed uint32) *pieceGolden {
	t.Helper()

	golden := goldens[name]
	if golden == nil {
		t.Fatalf("missing committed fixture for %s seed %d", name, seed)
	}

	return golden
}

// checkCryptLightSources compares PlaceCryptLights' output with an independent
// scan of the fixture grid for lava ids.
func checkCryptLightSources(
	t *testing.T, sources []levels.Light, piece [levels.MAXDUNX][levels.MAXDUNY]uint16,
) {
	t.Helper()

	lava := make(map[uint16]bool, len(upstreamLavaTiles))
	for _, tile := range upstreamLavaTiles {
		lava[tile] = true
	}

	if want := countTiles(piece, func(tile uint16) bool { return lava[tile] }); len(sources) != want {
		t.Fatalf("sources = %d, want %d", len(sources), want)
	}

	for _, source := range sources {
		if source.Radius != 3 {
			t.Fatalf("source %+v: radius = %d, want 3", source, source.Radius)
		}
		if !lava[piece[source.Position.X][source.Position.Y]] {
			t.Fatalf("source %+v is not on a lava tile", source)
		}
	}
}

// checkHiveLightSources compares PlaceHiveLights' output with an independent
// scan of the fixture grid for the 381..456 band.
func checkHiveLightSources(
	t *testing.T, sources []levels.Light, piece [levels.MAXDUNX][levels.MAXDUNY]uint16,
) {
	t.Helper()

	inBand := func(tile uint16) bool { return tile >= 381 && tile <= 456 }
	if want := countTiles(piece, inBand); len(sources) != want {
		t.Fatalf("sources = %d, want %d", len(sources), want)
	}

	for _, source := range sources {
		if source.Radius != 9 {
			t.Fatalf("source %+v: radius = %d, want 9", source, source.Radius)
		}
	}
}

// countTiles counts the tiles of a fixture grid that match.
func countTiles(piece [levels.MAXDUNX][levels.MAXDUNY]uint16, match func(uint16) bool) int {
	count := 0
	for _, row := range piece {
		for _, tile := range row {
			if match(tile) {
				count++
			}
		}
	}

	return count
}

// nearAnySource reports whether p is within radius world tiles of a source on
// either axis.
func nearAnySource(p levels.Point, sources []levels.Light, radius int) bool {
	for _, source := range sources {
		dx, dy := p.X-source.Position.X, p.Y-source.Position.Y
		if dx >= -radius && dx <= radius && dy >= -radius && dy <= radius {
			return true
		}
	}

	return false
}
