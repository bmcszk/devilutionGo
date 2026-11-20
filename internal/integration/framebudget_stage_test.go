// The AC-3 frame-budget stage (devilGo-prcc): the parity scene the scripted walk
// freezes at — the committed cathedral L1 layout at the walk's own seed, the hero
// standing on the level's own entry tile, its light cone around it — rasterised
// through the landed render surfaces and presented through the palette pass one
// frame at a time, so the bench and the budget pin read the same pipeline.
//
// Every step is a landed public entry: the committed DRLG layout for the level,
// the committed tiles corpus for the MIN/SOL/CEL records (through the same
// LoadMinPieces → DecodeDungeonPieces → ReencodeDungeonCels chain the town compose
// runs), levels.DoLighting for the hero's dLight cone, render.DrawWorld for the
// world pass, Surface.DrawDirect for the panel stage, render.DrawString for the
// text stage and Presenter.Draw for the palette lookup. Nothing on the render path
// is rewritten or instrumented: this file only wraps the entries in a frame clock.
package integration_test

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// budgetFrameMillis is the AC-3 budget: 60 fps, 16.67 ms per frame
	// (devilGo-ftbu AC-3, "no frame >16ms p95").
	budgetFrameMillis = 16.67
	// frameSamples is how many frames one measurement series holds. A p95 over
	// 128 samples is the 122nd value, so one scheduling spike cannot pass for a
	// distribution.
	frameSamples = 128
	// frameWarmup is how many frames run before sampling: first-frame
	// allocation, texture upload and shader warm-up never land in a percentile.
	frameWarmup = 8
	// tilesCorpusDir is the committed tiles corpus the scene rasterises.
	tilesCorpusDir = "../testdata/render/tiles"
	// drlgCorpusDir is the committed DRLG corpus (the level seeds' tile layers).
	drlgCorpusDir = "../testdata/drlg"
	// dpieceCorpusDir is the committed dPiece corpus under the DRLG tree.
	dpieceCorpusDir = drlgCorpusDir + "/dpiece"
	// celHeadFrames is the frame count of the committed cathedral CEL head
	// (tools/goldens tiles: `l1.cel` cut to 96 frames).
	celHeadFrames = 96
	// paletteFixture is the committed cathedral L1 palette dump.
	paletteFixture = "../testdata/render/pal/l1_1.pal.hex"
	// panelFixture is the committed CLX panel sheet the panel stage draws.
	panelFixture = "../testdata/gfx/clx/data_panel8bucp.clx.hex"
	// fontFixture is the committed 12-05 font sheet the text stage draws.
	fontFixture = "../testdata/assets/fonts/12-05.clx.hex"
	// heroLightRadius is the player's starting light radius (player.cpp:2317
	// `player._pLightRad = 10`).
	heroLightRadius uint8 = 10
	// frameLabel is the text stage's line. The committed 12-05 sheet carries
	// pixels for exactly 27 of its 256 slots (codes 0xD0..0xEA); the rest are
	// blank headers. The label is written with the drawable codes so the stage
	// draws a full line of real glyphs instead of a no-op over blanks — the
	// sheet's own corpus note says as much
	// (internal/testdata/assets/fonts/README.md).
	frameLabel = "\u00d0\u00d1\u00d2\u00d3\u00d4\u00d5\u00d6\u00d7\u00d8\u00d9\u00da\u00db\u00dc\u00dd" +
		"\u00de\u00df\u00e0\u00e1\u00e2\u00e3\u00e4\u00e5\u00e6\u00e7\u00e8\u00e9\u00ea"
)

// The frame pipeline's stage names, in draw order.
const (
	stageWorld   = "world"
	stagePanel   = "panel"
	stageText    = "text"
	stagePresent = "present"
)

// The MIN block's bit layout (dun_tile.hpp:91, :100): the frame index in the low
// 12 bits, the 3-bit tile shape above it.
const (
	blockFrameMask = 0x0FFF
	blockTypeMask  = 0xF000
)

// frameScene is the parity scene's loaded inputs: the frame surface, the injected
// world view, the view geometry, the palette state, the decoded pieces, the panel
// draw list and the font. Loaded once per test or bench, so fixture decode never
// lands in a sample.
type frameScene struct {
	surface render.Surface
	world   render.World
	view    render.View
	state   render.PaletteState
	pieces  levelPieces
	spots   []render.Clip
	font    *assets.Font
	// tables is the cathedral light-table set the tile rasteriser keys on.
	tables light.LightTables
	// pass is the palette presenter, built on first use: building it eagerly
	// would charge its texture upload to the first frame.
	pass *render.Presenter
}

// frameLayout is the parity scene's level image: the committed piece grid, its SOL
// classification record and the entry tile the walk's descend lands on.
type frameLayout struct {
	// piece is dPiece flattened x-major (index x*MAXDUNY+y), the corpus' own
	// indexing (world.LevelLayout.Piece).
	piece []int
	// sol is the cathedral SOL record with its hand patches applied, one entry per
	// piece id in the record's own bit layout (levels/dun_tile.hpp:106-114).
	sol []render.TileProperties
	// entry is the level's ViewPosition, the tile the render centres on.
	entry world.Tile
}

// levelPieces is the decoded tiles corpus the rasteriser reads: the micro blocks
// per piece, the re-encoded CEL and the level's MicroTileLen.
type levelPieces struct {
	micros  render.DungeonMicros
	cel     []byte
	tileLen uint8
}

// frameStage names one timed stage of the pipeline and the frame part it runs. The
// world stage is the frame's first layer and clears the surface itself; the panel,
// text and present stages run over the frame it left.
type frameStage struct {
	name string
	run  func(*frameScene)
}

// frameTimes is one measurement's per-stage series plus the pipeline total.
type frameTimes struct {
	stages map[string][]time.Duration
	total  []time.Duration
}

// frameSink rasterises the world pass into the frame surface through the landed
// tile rasteriser: floors blit their diamond pair, cells their micro stack,
// out-of-bounds tiles the black diamond (scrollrt.cpp:611-706, :1080-1116).
type frameSink struct {
	scene *frameScene
}

// newFrameScene builds the scene: the walk's cathedral seed and entry tile, the
// hero's dLight cone, the level's decoded tile records, the panel sheet and the
// font sheet.
func newFrameScene(tb testing.TB) *frameScene {
	tb.Helper()

	layout := loadFrameLayout(tb)
	hero := render.TilePos{X: layout.entry.X, Y: layout.entry.Y}

	pieces, err := decodeLevelPieces()
	if err != nil {
		tb.Fatalf("decode committed cathedral pieces: %v", err)
	}

	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)

	return &frameScene{
		surface: render.NewSurface(render.WorldLogicalWidth, render.WorldLogicalHeight),
		world:   frameWorld(layout, heroLitDungeon(hero, tables), hero),
		view:    frameView(hero, pieces.tileLen),
		state:   render.NewPaletteState(*loadFramePalette(tb), tables, assets.BlendingFor(0)),
		tables:  tables,
		pieces:  pieces,
		spots:   loadPanelSpots(tb),
		font:    loadFrameFont(tb),
	}
}

// loadFrameLayout reads one committed cathedral layout for any test or bench: the
// dPiece fixture, the retail SOL record with the cathedral patches, and the
// fixture's ViewPosition record. It mirrors dtest.LoadL1Layout but takes a
// testing.TB, so a benchmark can load the same image the tests load.
func loadFrameLayout(tb testing.TB) frameLayout {
	tb.Helper()

	return frameLayout{
		piece: readFramePieces(tb),
		sol:   readFrameSOL(tb),
		entry: readFrameEntry(tb),
	}
}

// readFramePieces parses the committed dPiece record of the walk's seed: the
// header record, then MAXDUNX*MAXDUNY x-major piece ids (dpiece_golden_test.go).
func readFramePieces(tb testing.TB) []int {
	tb.Helper()

	path := filepath.Join(dpieceCorpusDir, fmt.Sprintf("l1-dpiece-main-%d.csv", scenarioSeed))
	records := readFrameCSV(tb, path)
	if len(records) < 2 {
		tb.Fatalf("%s: %d records, want at least 2", path, len(records))
	}

	fields := records[1]
	if len(fields) != world.MAXDUNX*world.MAXDUNY {
		tb.Fatalf("%s: %d pieces, want %d", path, len(fields), world.MAXDUNX*world.MAXDUNY)
	}

	pieces := make([]int, len(fields))

	for i, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			tb.Fatalf("%s: piece %d %q: %v", path, i, field, err)
		}

		pieces[i] = value
	}

	return pieces
}

// readFrameSOL decodes the retail cathedral SOL record and applies the cathedral
// hand patches (dun_tile_data.cpp:103-158).
func readFrameSOL(tb testing.TB) []render.TileProperties {
	tb.Helper()

	solData, err := readHexFixture(filepath.Join(tilesCorpusDir, "sol", "l1.sol.hex"))
	if err != nil {
		tb.Fatalf("read cathedral SOL record: %v", err)
	}

	sol, err := render.LoadSOLData(solData)
	if err != nil {
		tb.Fatalf("LoadSOLData: %v", err)
	}

	patched := render.ApplySOLHandPatches(sol, render.LevelCathedral)

	return patched[:]
}

// readFrameEntry reads one tile-layer fixture's ViewPosition record: the fourth
// record's `ViewX,ViewY,leveltype` row (golden_test.go:99-105).
func readFrameEntry(tb testing.TB) world.Tile {
	tb.Helper()

	path := filepath.Join(drlgCorpusDir, fmt.Sprintf("l1-main-%d.csv", scenarioSeed))
	records := readFrameCSV(tb, path)
	if len(records) != 4 || len(records[3]) != 3 {
		tb.Fatalf("%s: %d records of %d last-record fields, want 4 of 3", path, len(records),
			len(records[len(records)-1]))
	}

	x, err := strconv.Atoi(records[3][0])
	if err != nil {
		tb.Fatalf("%s: parse ViewX: %v", path, err)
	}

	y, err := strconv.Atoi(records[3][1])
	if err != nil {
		tb.Fatalf("%s: parse ViewY: %v", path, err)
	}

	return world.Tile{X: x, Y: y}
}

// readFrameCSV parses one committed fixture's CSV records.
func readFrameCSV(tb testing.TB, path string) [][]string {
	tb.Helper()

	file, err := os.Open(path)
	if err != nil {
		tb.Fatalf("open fixture %s: %v", path, err)
	}

	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		tb.Fatalf("parse fixture %s: %v", path, err)
	}

	return records
}

// decodeLevelPieces runs the level tile decode the renderer consumes:
// LoadSOLData → ApplySOLHandPatches → LoadMinPieces → DecodeDungeonPieces →
// ReencodeDungeonCels, with the MIN blocks remapped onto the committed CEL head.
//
// The committed cathedral CEL is the retail file's first 96 frames; the level
// references 1119, and the rest of the record is not committed (no test reads an
// MPQ). Each block is therefore pointed at a committed frame of its own tile type,
// so every draw is a real retail RLE payload of the right shape and the frame
// table the rasteriser indexes is the re-encoder's own output.
func decodeLevelPieces() (levelPieces, error) {
	minData, err := readHexFixture(filepath.Join(tilesCorpusDir, "min", "l1.min.hex"))
	if err != nil {
		return levelPieces{}, err
	}

	solData, err := readHexFixture(filepath.Join(tilesCorpusDir, "sol", "l1.sol.hex"))
	if err != nil {
		return levelPieces{}, err
	}

	celData, err := readHexFixture(filepath.Join(tilesCorpusDir, "cel", "l1-head.hex"))
	if err != nil {
		return levelPieces{}, err
	}

	decoded, err := decodeDungeonLevel(minData, solData)
	if err != nil {
		return levelPieces{}, err
	}

	head := celHeadFramesOf(decoded.Frames)

	return levelPieces{
		micros:  remapOntoHead(decoded.Micros, head),
		cel:     render.ReencodeDungeonCels(celData, head),
		tileLen: decoded.TileLen,
	}, nil
}

// decodeDungeonLevel runs the MIN decode of one level over its SOL record.
func decodeDungeonLevel(minData, solData []byte) (render.DecodedPieces, error) {
	sol, err := render.LoadSOLData(solData)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("LoadSOLData: %w", err)
	}

	sol = render.ApplySOLHandPatches(sol, render.LevelCathedral)

	minPieces, err := render.LoadMinPieces(minData)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("LoadMinPieces: %w", err)
	}

	decoded, err := render.DecodeDungeonPieces(minPieces, render.LevelCathedral, sol)
	if err != nil {
		return render.DecodedPieces{}, fmt.Errorf("DecodeDungeonPieces: %w", err)
	}

	return decoded, nil
}

// celHeadFramesOf keeps the decoded frames the committed CEL head carries.
func celHeadFramesOf(frames []render.DunFrame) []render.DunFrame {
	head := make([]render.DunFrame, 0, celHeadFrames)
	for _, frame := range frames {
		if frame.Frame <= celHeadFrames {
			head = append(head, frame)
		}
	}

	return head
}

// remapOntoHead points every micro block at a committed frame of its own tile type
// (round-robin over the head's frames of that type), then renumbers the blocks onto
// the re-encoded table. A block whose type the head carries no frame of keeps its
// frame: the rasteriser then draws nothing for it.
func remapOntoHead(micros render.DungeonMicros, head []render.DunFrame) render.DungeonMicros {
	byType := framesByType(head)
	picks := make(map[render.TileType]int, len(byType))
	remapped := make(render.DungeonMicros, len(micros))

	for piece, blocks := range micros {
		remapped[piece] = blocks
		remapPieceBlocks(&remapped[piece], byType, picks)
	}

	render.RenumberLevelCelBlocks(remapped, render.ComputeCelBlockAdjustments(head))

	return remapped
}

// framesByType groups the head's frames by their MIN tile shape.
func framesByType(head []render.DunFrame) map[render.TileType][]uint16 {
	byType := make(map[render.TileType][]uint16, len(head))
	for _, frame := range head {
		byType[frame.Info.Type] = append(byType[frame.Info.Type], frame.Frame)
	}

	return byType
}

// remapPieceBlocks rewrites one piece's used blocks onto the head's frames,
// advancing the per-type round-robin cursor.
func remapPieceBlocks(blocks *render.MICROS, byType map[render.TileType][]uint16, picks map[render.TileType]int) {
	for block, celBlock := range blocks {
		if !celBlock.HasValue() {
			continue
		}

		candidates := byType[celBlock.Type()]
		if len(candidates) == 0 {
			continue
		}

		pick := picks[celBlock.Type()]
		picks[celBlock.Type()] = pick + 1

		blocks[block] = withFrame(celBlock, candidates[pick%len(candidates)])
	}
}

// withFrame replaces a MIN block's frame index and keeps its tile shape.
func withFrame(block render.LevelCelBlock, frame uint16) render.LevelCelBlock {
	return render.LevelCelBlock(uint16(block)&blockTypeMask | frame&blockFrameMask)
}

// heroLitDungeon is the level's dLight grid with the hero's cone lit: InitLight
// parks every tile at the dark default (levels/gendung.cpp:301-307) and DoLighting
// stamps the radius-10 cone the player carries (lighting.cpp:117-180,
// player.cpp:2317).
func heroLitDungeon(hero render.TilePos, tables light.LightTables) *levels.Dungeon {
	dungeon := levels.NewDungeon(levels.Cathedral, dungeonLevel)
	dungeon.InitLight()

	levels.DoLighting(dungeon, tables, levels.Point{X: hero.X, Y: hero.Y}, heroLightRadius, levels.Offset{})

	return dungeon
}

// frameWorld wraps the committed layout as the renderer's injected grid view: the
// dPiece/dSpecial/dLight lookups DrawWorld reads (scrollrt.cpp:784-953), with the
// hero standing on the entry tile and no monster on any tile — the state the walk
// freezes at dungeon-enter.
func frameWorld(layout frameLayout, dungeon *levels.Dungeon, hero render.TilePos) render.World {
	piece := framePieceLookup(layout.piece)
	properties := framePropertyLookup(piece, layout.sol)

	return render.World{
		InBounds: inFrameBounds,
		IsFloor: func(p render.TilePos) bool {
			return properties(p)&(render.TilePropertiesSolid|render.TilePropertiesBlockMissile) == 0
		},
		IsWall: func(p render.TilePos) bool {
			return properties(p)&(render.TilePropertiesSolid|render.TilePropertiesBlockMissile) != 0
		},
		NotSolid:    func(p render.TilePos) bool { return properties(p)&render.TilePropertiesSolid == 0 },
		Light:       frameLight(dungeon),
		Piece:       func(p render.TilePos) uint16 { return uint16(piece(p)) },
		Special:     func(render.TilePos) int8 { return 0 },
		Transparent: func(render.TilePos) bool { return false },
		Corpse:      func(render.TilePos) int8 { return 0 },
		Item:        func(render.TilePos) int8 { return 0 },
		ItemPostDraw: func(int8) bool {
			return false
		},
		DeadPlayer: func(render.TilePos) bool { return false },
		Object:     func(render.TilePos) (render.ObjectRef, bool) { return render.ObjectRef{}, false },
		Player:     framePlayer(hero),
		Monster:    func(render.TilePos) (render.Occupant, bool) { return render.Occupant{}, false },
	}
}

// framePieceLookup returns dPiece for one tile, 0 outside the grid (the
// x-major corpus indexing, world.LevelLayout.Piece).
func framePieceLookup(pieces []int) func(render.TilePos) int {
	return func(p render.TilePos) int {
		if !inFrameBounds(p) {
			return 0
		}

		return pieces[p.X*world.MAXDUNY+p.Y]
	}
}

// framePropertyLookup returns the tile's SOL word, the classification every tile
// predicate reads (levels/tile_properties.cpp:11-45); an unclassified tile is
// solid.
func framePropertyLookup(
	piece func(render.TilePos) int,
	sol []render.TileProperties,
) func(render.TilePos) render.TileProperties {
	return func(p render.TilePos) render.TileProperties {
		id := piece(p)
		if id <= 0 || id >= len(sol) {
			return render.TilePropertiesSolid
		}

		return sol[id]
	}
}

// frameLight is the World.Light lookup: dLight, dark outside the grid
// (levels/gendung.cpp:301-307).
func frameLight(dungeon *levels.Dungeon) func(render.TilePos) uint8 {
	return func(p render.TilePos) uint8 {
		if !inFrameBounds(p) {
			return light.LightsMax
		}

		return dungeon.Light[p.X][p.Y]
	}
}

// framePlayer is the World.Player lookup: the standing hero on its own tile and
// nothing anywhere else (player.cpp:2040-2050).
func framePlayer(hero render.TilePos) func(render.TilePos) (render.Occupant, bool) {
	return func(p render.TilePos) (render.Occupant, bool) {
		if p != hero {
			return render.Occupant{}, false
		}

		return render.Occupant{ID: 0, Encoded: 1, Mode: render.ModeStand}, true
	}
}

// inFrameBounds is InDungeonBounds over the level grid.
func inFrameBounds(p render.TilePos) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < world.MAXDUNX && p.Y < world.MAXDUNY
}

// frameView is the frame's tile walk: the 640x480 viewport geometry
// CalcViewportGeometry derives around the hero (scrollrt.cpp:1675-1721), walked
// MicroTileLen rows past the visible area so the tall wall cels are drawn.
func frameView(hero render.TilePos, microTiles uint8) render.View {
	const (
		viewRows    = 18
		viewColumns = 11
		viewOriginX = -32
		viewOriginY = 160
	)

	return render.View{
		Start:       render.TilePos{X: hero.X - (viewColumns+1)/2 + 1, Y: hero.Y - (viewRows+1)/2 + 1},
		Origin:      render.PixelPos{X: viewOriginX, Y: viewOriginY},
		Rows:        viewRows,
		Columns:     viewColumns,
		ScreenWidth: render.WorldLogicalWidth,
		MicroTiles:  int(microTiles),
	}
}

// frameStages is the pipeline in draw order: the world pass (DrawWorld: floor,
// tile content, out-of-bounds fill), the panel/sprite stage, the text stage, then
// the palette-lookup presentation pass.
func frameStages() []frameStage {
	return []frameStage{
		{name: stageWorld, run: drawWorldStage},
		{name: stagePanel, run: drawPanelStage},
		{name: stageText, run: drawTextStage},
		{name: stagePresent, run: drawPresentStage},
	}
}

// newFrameTimes allocates the series one measurement fills.
func newFrameTimes(samples int) *frameTimes {
	times := &frameTimes{stages: make(map[string][]time.Duration, len(frameStages()))}
	for _, stage := range frameStages() {
		times.stages[stage.name] = make([]time.Duration, 0, samples)
	}

	times.total = make([]time.Duration, 0, samples)

	return times
}

// measureFrames samples the full pipeline samples times: every sample runs all
// four stages in draw order and is timed as a whole, and each stage keeps its own
// series for the per-stage table.
func measureFrames(scene *frameScene, samples int) *frameTimes {
	stages := frameStages()

	for range frameWarmup {
		runFrame(scene, stages, nil)
	}

	times := newFrameTimes(samples)
	for range samples {
		times.total = append(times.total, runFrame(scene, stages, times))
	}

	return times
}

// runFrame runs one whole frame and returns its wall time; times, when non-nil,
// records each stage's own time.
func runFrame(scene *frameScene, stages []frameStage, times *frameTimes) time.Duration {
	begin := time.Now()

	for _, stage := range stages {
		stageBegin := time.Now()
		stage.run(scene)

		if times != nil {
			times.stages[stage.name] = append(times.stages[stage.name], time.Since(stageBegin))
		}
	}

	return time.Since(begin)
}

// p95 is one series' 95th percentile in milliseconds.
func p95(series []time.Duration) float64 {
	return float64(percentile(series, 95).Microseconds()) / 1000
}

// percentile returns the p-th percentile of a duration series by nearest rank,
// with no interpolation. An empty series is a broken harness, not a pass.
func percentile(series []time.Duration, p int) time.Duration {
	if len(series) == 0 {
		panic("percentile of an empty frame series")
	}

	sorted := slices.Clone(series)
	slices.Sort(sorted)

	return sorted[min(max((p*len(sorted)+99)/100, 1), len(sorted))-1]
}

// drawWorldStage rasterises the world pass into the frame surface.
func drawWorldStage(scene *frameScene) {
	clear(scene.surface.Pixels)
	render.DrawWorld(&frameSink{scene: scene}, scene.world, scene.view)
}

// drawPanelStage draws the panel sheet's frames onto the frame, the sprite layer
// over the world (ClxDraw, clx_render.cpp).
func drawPanelStage(scene *frameScene) {
	for _, clip := range scene.spots {
		scene.surface.DrawDirect(clip)
	}
}

// drawTextStage draws the frame's label line (text_render.cpp DrawString).
func drawTextStage(scene *frameScene) {
	render.DrawString(scene.surface, scene.font, frameLabel,
		render.Rect{X: 16, Y: 16, W: 320, H: 40}, render.Options{NoTable: true})
}

// drawPresentStage runs the palette-lookup pass: the whole 640x480 index surface
// through one DrawTrianglesShader (docs/research/dg-graphics.md G10 D2).
func drawPresentStage(scene *frameScene) {
	scene.presenter().Draw(scene.surface, nil, render.KeyTile)
}

// presenter builds the palette pass once per scene, the way the window boot does
// (NewPresenter + WithFadeLevel(256)).
func (s *frameScene) presenter() *render.Presenter {
	if s.pass == nil {
		s.pass = render.NewPresenter(s.state, render.WithFadeLevel(256))
	}

	return s.pass
}

// DrawFloorTile blits the tile's floor diamond pair (scrollrt.cpp:687-706).
func (s *frameSink) DrawFloorTile(target render.DrawTarget) {
	s.drawMicro(target, 0)
	s.drawMicroRight(target, 1)
}

// DrawCell blits one tile's micro stack: the floor pair and the tall rows
// (scrollrt.cpp:611-666). Floor tiles draw in the floor pass, so they return.
func (s *frameSink) DrawCell(target render.DrawTarget) {
	micro, ok := s.microsOf(target.Tile)
	if !ok || s.scene.world.IsFloor(target.Tile) {
		return
	}

	s.drawCellFloorPair(target, micro)
	s.drawCellTallRows(target, micro)
}

// DrawBlackTile fills the out-of-bounds black diamond (dun_render.cpp:1080-1116).
func (s *frameSink) DrawBlackTile(at render.PixelPos) {
	render.WorldDrawBlackTile(s.scene.surface, at.X, at.Y)
}

// The entity, missile and arch layers draw nothing here: the parity scene is the
// walk's standing hero before any kill or loot, and the arch decode is a separate
// landed slice (the town compose's sink leaves the same methods empty).
func (*frameSink) DrawMissilePre(render.DrawTarget)        {}
func (*frameSink) DrawCorpse(render.DrawTarget)            {}
func (*frameSink) DrawObjectPre(render.DrawTarget)         {}
func (*frameSink) DrawItemPre(render.DrawTarget)           {}
func (*frameSink) DrawDeadPlayer(render.DrawTarget)        {}
func (*frameSink) DrawMonster(render.DrawTarget, int)      {}
func (*frameSink) DrawItemPost(render.DrawTarget)          {}
func (*frameSink) DrawObjectPost(render.DrawTarget)        {}
func (*frameSink) DrawMissilePost(render.DrawTarget)       {}
func (*frameSink) DrawArch(render.DrawTarget, int8)        {}
func (*frameSink) DrawArchBlended(render.DrawTarget, int8) {}
func (*frameSink) DrawArchDirect(render.DrawTarget, int8)  {}

// DrawPlayer draws nothing: the hero sprite sheet is not part of the committed
// corpus this gate measures, and the frame budget holds without it.
func (*frameSink) DrawPlayer(render.DrawTarget, int) {}

// drawCellFloorPair blits a wall tile's two floor halves, the right one displaced
// a frame width east (scrollrt.cpp:618-641).
func (s *frameSink) drawCellFloorPair(target render.DrawTarget, micro render.MICROS) {
	for _, block := range []int{0, 1} {
		at := target.At
		if block == 1 {
			at.X += render.DunFrameWidth
		}

		s.drawBlock(render.DrawTarget{Tile: target.Tile, At: at, Light: target.Light}, micro[block])
	}
}

// drawCellTallRows blits micro blocks 2.., one pair per row walking up
// (scrollrt.cpp:640-666).
func (s *frameSink) drawCellTallRows(target render.DrawTarget, micro render.MICROS) {
	at := target.At
	for i := 2; i+1 < len(micro); i += 2 {
		at.Y -= render.TileHeight

		for _, block := range []int{i, i + 1} {
			s.drawBlock(render.DrawTarget{Tile: target.Tile, At: at, Light: target.Light}, micro[block])
		}
	}
}

// drawMicro blits one half of the floor diamond pair.
func (s *frameSink) drawMicro(target render.DrawTarget, block int) {
	micro, ok := s.microsOf(target.Tile)
	if !ok {
		return
	}

	s.drawBlock(target, micro[block])
}

// drawMicroRight blits the pair's right half one frame width east
// (scrollrt.cpp RightFrameDisplacement).
func (s *frameSink) drawMicroRight(target render.DrawTarget, block int) {
	target.At.X += render.DunFrameWidth
	s.drawMicro(target, block)
}

// microsOf resolves one tile's decoded micro blocks through its piece id.
func (s *frameSink) microsOf(tile render.TilePos) (render.MICROS, bool) {
	piece := int(s.scene.world.Piece(tile))
	if piece <= 0 || piece >= len(s.scene.pieces.micros) {
		return render.MICROS{}, false
	}

	return s.scene.pieces.micros[piece], true
}

// drawBlock blits one micro block's re-encoded frame; the triangle shapes are 31
// rows tall and every other shape 32 (dun_render.hpp:148-157).
func (s *frameSink) drawBlock(target render.DrawTarget, block render.LevelCelBlock) {
	if !block.HasValue() {
		return
	}

	height := render.DunFrameHeight
	if block.Type() == render.TileLeftTriangle || block.Type() == render.TileRightTriangle {
		height = render.TriangleHeight
	}

	lighting := render.TileLighting{
		Level:  target.Light,
		Blend:  s.scene.state.LUT,
		Tables: &s.scene.tables.Tables,
	}
	render.RenderTileFrame(s.scene.surface, target.At, block.Type(),
		dunFrame(s.scene.pieces.cel, block.Frame()), height, render.MaskSolid, lighting)
}

// dunFrame is GetDunFrame (dun_render.hpp:123-127): the frame table lookup over
// the re-encoded CEL.
func dunFrame(cel []byte, frame uint16) []byte {
	return cel[binary.LittleEndian.Uint32(cel[4*frame:]):]
}

// readHexFixture decodes one committed hex-dump fixture (16 bytes per line).
func readHexFixture(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(raw)), ""))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return decoded, nil
}

// loadFramePalette reads the committed cathedral palette dump.
func loadFramePalette(tb testing.TB) *assets.Palette {
	tb.Helper()

	raw, err := readHexFixture(paletteFixture)
	if err != nil {
		tb.Fatalf("load cathedral palette: %v", err)
	}

	palette, err := assets.LoadPaletteFromBytes(raw)
	if err != nil {
		tb.Fatalf("decode cathedral palette: %v", err)
	}

	return palette
}

// loadPanelSpots decodes the committed panel sheet and lays its frames along the
// bottom bar, the panel stage's draw list.
func loadPanelSpots(tb testing.TB) []render.Clip {
	tb.Helper()

	raw, err := readHexFixture(panelFixture)
	if err != nil {
		tb.Fatalf("load panel sheet: %v", err)
	}

	sheet, err := assets.DecodeCLX(raw)
	if err != nil {
		tb.Fatalf("decode panel sheet: %v", err)
	}

	spots := make([]render.Clip, 0, len(sheet.Frames))
	for i, frame := range sheet.Frames {
		spots = append(spots, render.Clip{
			X: 16 + i*frame.Width, Y: render.WorldLogicalHeight - frame.Height,
			Width: frame.Width, Height: frame.Height, Pixels: frame.Pixels,
		})
	}

	return spots
}

// loadFrameFont decodes the committed 12-05 font sheet.
func loadFrameFont(tb testing.TB) *assets.Font {
	tb.Helper()

	raw, err := readHexFixture(fontFixture)
	if err != nil {
		tb.Fatalf("load font sheet: %v", err)
	}

	font, err := assets.DecodeFontSheet(raw)
	if err != nil {
		tb.Fatalf("decode font sheet: %v", err)
	}

	return font
}
