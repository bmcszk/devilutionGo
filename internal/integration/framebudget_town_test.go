// The AC-3 town row (devilGo-ht4l): the milestone AC reads "60fps rendering of
// town + dungeon", and the cathedral row alone evidences only the dungeon half.
// This file builds the town scene the game actually opens on — the committed
// CreateTown layout at the AC-1 walk's own seed, the hero standing on the town
// entry tile, the town palette — and runs it through the frame pipeline the
// cathedral row measures, so both rows report per-stage plus total p95 against
// the same budget class.
//
// Scene inputs are landed public entries only: the committed town dPiece/dSpecial
// fixture, the committed town MIN/SOL/pal records (through the same
// LoadSOLData → ApplySOLHandPatches → LoadMinPieces → DecodeDungeonPieces →
// ReencodeDungeonCels chain the town compose runs), levels.CreateTown's own
// light init, render.DrawWorld, Surface.DrawDirect, render.DrawString and
// Presenter.Draw. Nothing on the render path is rewritten or instrumented.
//
// Disclosed limits, identical to the cathedral row's: the committed CEL head is
// the cathedral's first 96 frames and the town MIN references 3803, so every
// town block is pointed at a committed frame of its own tile shape before the
// re-encode (remapOntoHead). The hero sprite sheet is not part of the committed
// corpus, so the town frame is the standing hero-less world pass — the same
// state the town compose renders (internal/present/scene/town_render.go leaves
// DrawPlayer empty).
package integration_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// townCorpusDir is the committed CreateTown corpus the town dPiece and
	// dSpecial grids come from (internal/testdata/drlg/README.md, "Town").
	townCorpusDir = drlgCorpusDir + "/town"
	// townMinFixture is the retail town MIN record: nlevels\towndata\town.min
	// (dun_tile_data.cpp:73), 1379 pieces of 16 blocks.
	townMinFixture = tilesCorpusDir + "/min/town.min.hex"
	// townSOLFixture is the retail town SOL record: levels\towndata\town.sol
	// (dun_tile_data.cpp:107).
	townSOLFixture = tilesCorpusDir + "/sol/town.sol.hex"
	// townPaletteFixture is the committed town palette dump: levels\towndata\
	// town.pal (LoadLvlGFX DTYPE_TOWN, diablo.cpp:1371-1387).
	townPaletteFixture = "../testdata/render/pal/town.pal.hex"
	// townDunRecords is the town fixture's record count: the 112×112 header, the
	// dPiece grid, the dSpecial grid, the 40×40 automap tiles and
	// `ViewX,ViewY,leveltype`.
	townDunRecords = 5
)

// The town entry tile: ViewPosition of ENTRY_MAIN (town.cpp:363-365), the tile
// the player lands on at a new game and the tile the town compose centres on
// (present/scene town_compose.go heroTile).
const (
	townHeroX = 75
	townHeroY = 68
)

// townLayout is one committed town image: dPiece and dSpecial flattened x-major
// (cell (x,y) at x*MAXDUNY+y, the corpus' own indexing) plus the level's SOL
// record with its town hand patches applied.
type townLayout struct {
	piece   []int
	special []int8
	sol     []render.TileProperties
}

// TestM1TownFrameBudgetP95 pins the town half of AC-3: the town scene's p95 frame
// time, stage by stage and in total, stays under the budget of the runner class it
// measures — the same budget class and the same nearest-rank p95 the cathedral row
// asserts. The row runs in its own process for the reason
// measureRowInOwnProcess documents.
func TestM1TownFrameBudgetP95(t *testing.T) {
	// when the town row is measured in its own process
	table := measureRowInOwnProcess(t, townRowLabel)

	// then the row reported its per-stage table
	t.Logf("%s row p95 over %d frames: %s", townRowLabel, budgetSamples(), table)
}

// TestM1TownFrameSceneIsNotEmpty pins that the town row measures the town scene:
// the world pass paints a real town frame (many palette indices over the tile
// spread, the panel and text stages over it) and the resulting frame is not the
// cathedral row's frame, so a copy of the dungeon scene cannot pass as the town
// row.
func TestM1TownFrameSceneIsNotEmpty(t *testing.T) {
	// given the town scene and the cathedral scene the other row measures
	town := newTownFrameScene(t)
	cathedral := newFrameScene(t)

	// when both run one whole frame in draw order
	runFrame(town, frameStages(), nil)
	runFrame(cathedral, frameStages(), nil)

	// then the town world pass painted the layout: a real town frame carries the
	// terrain, house and foliage palette indices, not a flat fill
	if indices := distinctIndices(town.surface); indices < 16 {
		t.Fatalf("town frame carries %d distinct palette indices, want the layout's spread", indices)
	}

	// and the town row is its own scene, not the dungeon row's frame
	if slices.Equal(town.surface.Pixels, cathedral.surface.Pixels) {
		t.Fatal("town frame equals the cathedral frame: the town row is measuring the dungeon scene")
	}

	t.Logf("town distinct indices: %d, cathedral distinct indices: %d",
		distinctIndices(town.surface), distinctIndices(cathedral.surface))
}

// newTownFrameScene builds the town scene the bench and the budget pin measure:
// the committed town layout at the parity seed, the town MIN/SOL decode remapped
// onto the committed CEL head, the town palette and light tables, and the hero
// standing on the town entry tile.
func newTownFrameScene(tb testing.TB) *frameScene {
	tb.Helper()

	layout := loadTownFrameLayout(tb)
	pieces := decodeTownFramePieces(tb)
	hero := render.TilePos{X: townHeroX, Y: townHeroY}
	tables := light.MakeLightTables(light.DTYPE_TOWN)

	return &frameScene{
		surface: render.NewSurface(render.WorldLogicalWidth, render.WorldLogicalHeight),
		world:   townFrameWorld(layout),
		view:    frameView(hero, pieces.tileLen),
		state:   render.NewPaletteState(*loadTownPalette(tb), tables, assets.BlendingFor(0)),
		tables:  tables,
		pieces:  pieces,
		spots:   loadPanelSpots(tb),
		font:    loadFrameFont(tb),
	}
}

// loadTownFrameLayout reads one committed town layout at the AC-1 walk's seed:
// the dPiece and dSpecial grids, and the retail town SOL record with its hand
// patches (dun_tile_data.cpp:155-158).
func loadTownFrameLayout(tb testing.TB) townLayout {
	tb.Helper()

	path := filepath.Join(townCorpusDir, fmt.Sprintf("l0-town-main-%d.csv", scenarioSeed))
	records := readFrameCSV(tb, path)
	if len(records) != townDunRecords {
		tb.Fatalf("%s: %d records, want %d", path, len(records), townDunRecords)
	}

	cells := worldCellCount(tb, path, records[0])

	return townLayout{
		piece:   parseTownPieces(tb, path, records[1], cells),
		special: parseTownSpecial(tb, path, records[2], cells),
		sol:     readTownSOL(tb),
	}
}

// worldCellCount validates the fixture's "MAXDUNX,MAXDUNY" header and returns
// the cell count its grids must carry.
func worldCellCount(tb testing.TB, path string, header []string) int {
	tb.Helper()

	if len(header) != 2 {
		tb.Fatalf("%s: header %v, want 2 fields", path, header)
	}

	x, err := strconv.Atoi(header[0])
	if err != nil {
		tb.Fatalf("%s: parse header width: %v", path, err)
	}

	y, err := strconv.Atoi(header[1])
	if err != nil {
		tb.Fatalf("%s: parse header height: %v", path, err)
	}

	if x != levels.MAXDUNX || y != levels.MAXDUNY {
		tb.Fatalf("%s: header %v, want %d,%d", path, header, levels.MAXDUNX, levels.MAXDUNY)
	}

	return x * y
}

// parseTownPieces decodes the fixture's dPiece record.
func parseTownPieces(tb testing.TB, path string, record []string, cells int) []int {
	tb.Helper()

	if len(record) != cells {
		tb.Fatalf("%s: %d pieces, want %d", path, len(record), cells)
	}

	pieces := make([]int, len(record))
	for i, field := range record {
		value, err := strconv.Atoi(field)
		if err != nil {
			tb.Fatalf("%s: piece %d %q: %v", path, i, field, err)
		}

		pieces[i] = value
	}

	return pieces
}

// parseTownSpecial decodes the fixture's dSpecial record: the door and arch
// markers InitDungeonPieces derives (levels/town.go initTownPieces).
func parseTownSpecial(tb testing.TB, path string, record []string, cells int) []int8 {
	tb.Helper()

	if len(record) != cells {
		tb.Fatalf("%s: %d special cells, want %d", path, len(record), cells)
	}

	special := make([]int8, len(record))
	for i, field := range record {
		value, err := strconv.Atoi(field)
		if err != nil {
			tb.Fatalf("%s: special %d %q: %v", path, i, field, err)
		}

		special[i] = int8(value)
	}

	return special
}

// readTownSOL decodes the retail town SOL record and applies the town hand
// patches (dun_tile_data.cpp:155-158).
func readTownSOL(tb testing.TB) []render.TileProperties {
	tb.Helper()

	solData, err := readHexFixture(townSOLFixture)
	if err != nil {
		tb.Fatalf("read town SOL record: %v", err)
	}

	sol, err := render.LoadSOLData(solData)
	if err != nil {
		tb.Fatalf("LoadSOLData: %v", err)
	}

	patched := render.ApplySOLHandPatches(sol, render.LevelTown)

	return patched[:]
}

// loadTownPalette reads the committed town palette dump.
func loadTownPalette(tb testing.TB) *assets.Palette {
	tb.Helper()

	raw, err := readHexFixture(townPaletteFixture)
	if err != nil {
		tb.Fatalf("load town palette: %v", err)
	}

	palette, err := assets.LoadPaletteFromBytes(raw)
	if err != nil {
		tb.Fatalf("decode town palette: %v", err)
	}

	return palette
}

// decodeTownFramePieces runs the level tile decode the renderer consumes for the
// town: the retail town MIN/SOL records through the 16-block town scramble
// (dun_tile_data.cpp:168-177), with the MIN blocks remapped onto the committed
// CEL head exactly as the cathedral row remaps its own.
func decodeTownFramePieces(tb testing.TB) levelPieces {
	tb.Helper()

	minData, err := readHexFixture(townMinFixture)
	if err != nil {
		tb.Fatalf("read town MIN record: %v", err)
	}

	solData, err := readHexFixture(townSOLFixture)
	if err != nil {
		tb.Fatalf("read town SOL record: %v", err)
	}

	celData, err := readHexFixture(filepath.Join(tilesCorpusDir, "cel", "l1-head.hex"))
	if err != nil {
		tb.Fatalf("read committed CEL head: %v", err)
	}

	decoded, err := decodeTownPiecesFrom(minData, solData)
	if err != nil {
		tb.Fatalf("decode committed town pieces: %v", err)
	}

	head := townCelHeadFrames(tb)

	return levelPieces{
		micros:  remapOntoHead(decoded.Micros, head),
		cel:     render.ReencodeDungeonCels(celData, head),
		tileLen: decoded.TileLen,
	}
}

// decodeTownPiecesFrom runs the MIN decode of the town over its SOL record.
func decodeTownPiecesFrom(minData, solData []byte) (render.DecodedPieces, error) {
	sol, err := render.LoadSOLData(solData)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("LoadSOLData: %w", err)
	}

	sol = render.ApplySOLHandPatches(sol, render.LevelTown)

	minPieces, err := render.LoadMinPieces(minData)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("LoadMinPieces: %w", err)
	}

	decoded, err := render.DecodeDungeonPieces(minPieces, render.LevelTown, sol)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("DecodeDungeonPieces: %w", err)
	}

	return decoded, nil
}

// townCelHeadFrames is the frame list the committed CEL head carries: the
// cathedral MIN decode's frames up to the head's cut (celHeadFrames), which is
// the table every scene remaps onto.
func townCelHeadFrames(tb testing.TB) []render.DunFrame {
	tb.Helper()

	minData, err := readHexFixture(filepath.Join(tilesCorpusDir, "min", "l1.min.hex"))
	if err != nil {
		tb.Fatalf("read cathedral MIN record: %v", err)
	}

	solData, err := readHexFixture(filepath.Join(tilesCorpusDir, "sol", "l1.sol.hex"))
	if err != nil {
		tb.Fatalf("read cathedral SOL record: %v", err)
	}

	decoded, err := decodeDungeonLevel(minData, solData)
	if err != nil {
		tb.Fatalf("decode cathedral head frames: %v", err)
	}

	return celHeadFramesOf(decoded.Frames)
}

// townFrameWorld wraps the committed town grids as the renderer's injected grid
// view: the dPiece/dSpecial/dLight lookups DrawWorld reads (scrollrt.cpp:784-953)
// with the town classification the town compose's adapter uses
// (internal/present/scene/town_render.go), the fully lit dLight a fresh town
// carries (levels/gendung.cpp:301-307) and no entity on any tile.
func townFrameWorld(layout townLayout) render.World {
	lightGrid := townFrameLight()

	return render.World{
		InBounds:    inFrameBounds,
		IsFloor:     townFloorLookup(layout),
		IsWall:      townWallLookup(layout),
		NotSolid:    townNotSolidLookup(layout),
		Light:       townLightLookup(lightGrid),
		Piece:       func(p render.TilePos) uint16 { return uint16(townPieceAt(layout, p)) },
		Special:     func(p render.TilePos) int8 { return layout.special[p.X*levels.MAXDUNY+p.Y] },
		Transparent: func(render.TilePos) bool { return false },
		Corpse:      func(render.TilePos) int8 { return 0 },
		Item:        func(render.TilePos) int8 { return 0 },
		ItemPostDraw: func(int8) bool {
			return false
		},
		DeadPlayer: func(render.TilePos) bool { return false },
		Object:     func(render.TilePos) (render.ObjectRef, bool) { return render.ObjectRef{}, false },
		Player:     func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		Monster:    func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
		Town:       true,
	}
}

// townFrameLight is the town's dLight grid after InitLight: town's default is 0,
// fully lit (levels/light.go InitLight, levels/gendung.cpp:301-307).
func townFrameLight() [levels.MAXDUNX][levels.MAXDUNY]uint8 {
	dungeon := levels.NewDungeon(levels.Town, 0)
	dungeon.InitLight()

	return dungeon.Light
}

// townPieceAt is dPiece for one tile, 0 outside the grid.
func townPieceAt(layout townLayout, p render.TilePos) int {
	if !inFrameBounds(p) {
		return 0
	}

	return layout.piece[p.X*levels.MAXDUNY+p.Y]
}

// townPropertiesAt is the tile's SOL word through its dPiece id, "solid" for an
// unclassified cell (levels/tile_properties.cpp:11-45).
func townPropertiesAt(layout townLayout, p render.TilePos) render.TileProperties {
	id := townPieceAt(layout, p)
	if id <= 0 || id >= len(layout.sol) {
		return render.TilePropertiesSolid
	}

	return layout.sol[id]
}

// townFloorLookup is IsFloor: `!TileHasAny(tile, Solid|BlockMissile)`
// (scrollrt.cpp:114-117).
func townFloorLookup(layout townLayout) func(render.TilePos) bool {
	return func(p render.TilePos) bool {
		blocked := render.TilePropertiesSolid | render.TilePropertiesBlockMissile

		return townPropertiesAt(layout, p)&blocked == 0
	}
}

// townWallLookup is IsWall: `!IsFloor(tile) || dSpecial[tile] != 0`
// (scrollrt.cpp:119-122).
func townWallLookup(layout townLayout) func(render.TilePos) bool {
	floor := townFloorLookup(layout)

	return func(p render.TilePos) bool {
		return !floor(p) || layout.special[p.X*levels.MAXDUNY+p.Y] != 0
	}
}

// townNotSolidLookup is IsTileNotSolid (levels/tile_properties.cpp:11-19), the
// wall pre-pass's test (scrollrt.cpp:1009-1024).
func townNotSolidLookup(layout townLayout) func(render.TilePos) bool {
	return func(p render.TilePos) bool {
		return townPropertiesAt(layout, p)&render.TilePropertiesSolid == 0
	}
}

// townLightLookup is the World.Light lookup: dLight, dark outside the grid
// (levels/gendung.cpp:301-307).
func townLightLookup(grid [levels.MAXDUNX][levels.MAXDUNY]uint8) func(render.TilePos) uint8 {
	return func(p render.TilePos) uint8 {
		if !inFrameBounds(p) {
			return light.LightsMax
		}

		return grid[p.X][p.Y]
	}
}
