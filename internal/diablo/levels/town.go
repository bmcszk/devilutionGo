package levels

// The fixed .dun resources CreateTown loads, by upstream path
// (levels/town.cpp:202-207). The caller supplies their bytes, replacing the
// upstream LoadFileInMem calls.
const (
	// TownSector1 is the 25x25 north-east quadrant, placed at (46,46).
	TownSector1 = `levels\towndata\sector1s.dun`
	// TownSector2 is the 25x23 north-west quadrant, placed at (46,0).
	TownSector2 = `levels\towndata\sector2s.dun`
	// TownSector3 is the 23x25 south-east quadrant, placed at (0,46).
	TownSector3 = `levels\towndata\sector3s.dun`
	// TownSector4 is the 23x23 south-west quadrant, placed at (0,0).
	TownSector4 = `levels\towndata\sector4s.dun`
	// TownAutomap is the 40x40 automap tile layer (`dungeon`).
	TownAutomap = `levels\towndata\automap.dun`
)

// Town mega-tile ids the pass writes directly (levels/town.cpp:191-244).
const (
	// townRockTile fills every dPiece cell before the sectors land
	// (town.cpp:193-199).
	townRockTile = 426
	// townEmptySectorTile is the dPiece value FillSector leaves for a sector
	// cell whose tile id is 0 (town.cpp:35-38).
	townEmptySectorTile = 218
	// townAutomapFloor is PlaceDunTiles' floorId for the automap layer; zero
	// keeps the cells a sector left alone (town.cpp:208).
	townAutomapFloor = 0
)

// townSpecials maps one dPiece value to its dSpecial code; anything else keeps
// the zero InitTownPieces leaves (levels/town.cpp:151-184).
var townSpecials = map[uint16]int8{
	359: 1,
	357: 2,
	128: 6,
	129: 7,
	127: 8,
	116: 9,
	156: 10,
	157: 11,
	155: 12,
	161: 13,
	159: 14,
	213: 15,
	211: 16,
	216: 17,
	215: 18,
}

// TownOptions is the input of CreateTown: the resources it loads by path, the
// level's mega-tile table and the warp/quest snapshot its closed-warp patches
// read. The RNG is d.Rnd, which the caller seeds (levels/town.cpp:224-226 is the
// only draw in the level, so the seed is an explicit input).
type TownOptions struct {
	// Entry is lvl_entry; it only moves ViewPosition (levels/town.cpp:364-384).
	Entry Entry
	// Files maps the TownSector*/TownAutomap paths to their `.dun` bytes.
	Files map[string][]byte
	// MegaTiles is the town pMegaTiles table (levels/towndata/town.til), which
	// FillSector indexes by tile id minus one (levels/town.cpp:39-44).
	MegaTiles []MegaTile
	// Warps is IsWarpOpen(dungeon_type); a missing key is a closed warp, so
	// every patch runs (levels/town.cpp:210-238).
	Warps map[LevelType]bool
	// Hellfire is gbIsHellfire: it adds the hive and grave patches
	// (levels/town.cpp:228-238).
	Hellfire bool
	// Quests is the quest snapshot; Quests.PWATER selects the live poisoned
	// water fountain (levels/town.cpp:240-244).
	Quests Quests
}

// CreateTown is the town generator (levels/town.cpp:359-387, DrlgTPass3): it
// writes the `dungeon` tile layer into d and returns the dPiece/dSpecial grids
// the same pass builds.
//
// It is not a generator in the CreateL5Dungeon sense: one pass, no attempts, no
// rejection loop, and no SetRndSeed inside (docs/research/dg-town-inventory.md).
func CreateTown(d *Dungeon, opts TownOptions) PieceGrid {
	setTownViewBounds(d, opts.Entry)

	var grid PieceGrid
	fillTownRock(&grid)
	fillTownSectors(&grid, opts)
	placeTownAutomap(d, opts)
	applyTownWarpPatches(d, &grid, opts)
	applyTownHellfirePatches(d, &grid, opts)
	fillTownPoisonedWater(&grid, opts)
	initTownPieces(&grid)

	return grid
}

// setTownViewBounds sets the play rectangle every town map shares and the
// per-entry camera position (levels/town.cpp:361-385).
func setTownViewBounds(d *Dungeon, entry Entry) {
	switch entry {
	case EntryMain: // new game
		d.ViewPosition = Point{X: 75, Y: 68}
	case EntryPrev: // cathedral
		d.ViewPosition = Point{X: 25, Y: 31}
	default:
		// ENTRY_TWARPUP reads TWarpFrom, game-flow state the generator does not
		// receive here; the two entries above are the ones the fixtures drive.
	}
}

// fillTownRock seeds every dPiece cell with solid rock, which the four sector
// fills then overwrite except for the 16-tile border (levels/town.cpp:193-199).
func fillTownRock(grid *PieceGrid) {
	for x := range grid.Piece {
		for y := range grid.Piece[x] {
			grid.Piece[x][y] = townRockTile
		}
	}
}

// fillTownSectors places the four town quadrants in upstream order
// (levels/town.cpp:202-205).
func fillTownSectors(grid *PieceGrid, opts TownOptions) {
	placements := []struct {
		path string
		at   Point
	}{
		{TownSector1, Point{X: 46, Y: 46}},
		{TownSector2, Point{X: 46, Y: 0}},
		{TownSector3, Point{X: 0, Y: 46}},
		{TownSector4, Point{X: 0, Y: 0}},
	}

	for _, placement := range placements {
		fillSector(grid, opts.Files[placement.path], placement.at, opts.MegaTiles)
	}
}

// fillSector expands one sector's mega tiles into its 2x2 world-tile blocks
// (levels/town.cpp:26-58). A tile id of 0 leaves the empty-sector value.
func fillSector(grid *PieceGrid, dunData []byte, at Point, megaTiles []MegaTile) {
	size := GetDunSize(dunData)

	for j := range size.Height {
		x := at.X
		for i := range size.Width {
			mega := MegaTile{
				Micro1: townEmptySectorTile,
				Micro2: townEmptySectorTile,
				Micro3: townEmptySectorTile,
				Micro4: townEmptySectorTile,
			}

			if tileID := int(dunTileAt(dunData, size, i, j)) - 1; tileID >= 0 {
				// Upstream indexes pMegaTiles unchecked; a fixture tile id past
				// the table is the same out-of-bounds read there.
				mega = megaTiles[tileID]
			}

			grid.Piece[x][at.Y+2*j] = mega.Micro1
			grid.Piece[x+1][at.Y+2*j] = mega.Micro2
			grid.Piece[x][at.Y+2*j+1] = mega.Micro3
			grid.Piece[x+1][at.Y+2*j+1] = mega.Micro4
			x += 2
		}
	}
}

// dunTileAt reads one mega-tile id of a `.dun` tile layer (levels/town.cpp:31).
func dunTileAt(dunData []byte, size Size, i, j int) uint16 {
	offset := 4 + 2*(j*size.Width+i)

	return uint16(dunData[offset]) | uint16(dunData[offset+1])<<8
}

// placeTownAutomap copies the automap tile layer into `dungeon`
// (levels/town.cpp:207-208).
func placeTownAutomap(d *Dungeon, opts TownOptions) {
	d.PlaceDunTiles(opts.Files[TownAutomap], Point{}, townAutomapFloor)
}

// applyTownWarpPatches draws the closed-warp entrances: the catacombs stairs,
// the caves entrance and the hell gate (levels/town.cpp:210-227).
func applyTownWarpPatches(d *Dungeon, grid *PieceGrid, opts TownOptions) {
	if !opts.Warps[Catacombs] {
		d.Tiles[20][7] = 10
		d.Tiles[20][6] = 8
		fillTile(grid, Point{X: 48, Y: 20}, 320, opts.MegaTiles)
	}

	if !opts.Warps[Caves] {
		d.Tiles[4][30] = 8
		fillTile(grid, Point{X: 16, Y: 68}, 332, opts.MegaTiles)
		fillTile(grid, Point{X: 16, Y: 70}, 331, opts.MegaTiles)
	}

	if !opts.Warps[Hell] {
		for x := 15; x <= 17; x++ {
			d.Tiles[x][35] = 7
		}

		for x := 36; x < 46; x++ {
			fillTile(grid, Point{X: x, Y: 78}, int(d.Rnd.GenerateRnd(4))+1, opts.MegaTiles)
		}
	}
}

// applyTownHellfirePatches draws the hive and grave doors, open or closed
// (levels/town.cpp:228-238).
func applyTownHellfirePatches(d *Dungeon, grid *PieceGrid, opts TownOptions) {
	if !opts.Hellfire {
		return
	}

	if opts.Warps[Nest] {
		townOpenHive(d, grid)
	} else {
		townCloseHive(d, grid)
	}

	if opts.Warps[Crypt] {
		townOpenGrave(d, grid)
	} else {
		townCloseGrave(d, grid)
	}
}

// fillTownPoisonedWater writes the fountain the Poisoned Water Supply quest
// replaces most of the game (levels/town.cpp:240-244).
func fillTownPoisonedWater(grid *PieceGrid, opts TownOptions) {
	if opts.Quests.PWATER {
		fillTile(grid, Point{X: 60, Y: 70}, 342, opts.MegaTiles)

		return
	}

	fillTile(grid, Point{X: 60, Y: 70}, 71, opts.MegaTiles)
}

// fillTile expands one mega tile into a 2x2 dPiece block
// (levels/town.cpp:66-79).
func fillTile(grid *PieceGrid, at Point, tileID int, megaTiles []MegaTile) {
	mega := megaTiles[tileID-1]

	grid.Piece[at.X][at.Y] = mega.Micro1
	grid.Piece[at.X+1][at.Y] = mega.Micro2
	grid.Piece[at.X][at.Y+1] = mega.Micro3
	grid.Piece[at.X+1][at.Y+1] = mega.Micro4
}

// initTownPieces derives dSpecial from the dPiece values the sectors placed
// (levels/town.cpp:149-186).
func initTownPieces(grid *PieceGrid) {
	for x := range grid.Piece {
		for y := range grid.Piece[x] {
			grid.Special[x][y] = townSpecialFor(grid.Piece[x][y])
		}
	}
}

// townSpecialFor maps one dPiece value to its dSpecial code, zero for the
// pieces InitTownPieces ignores (levels/town.cpp:151-184).
func townSpecialFor(piece uint16) int8 {
	return townSpecials[piece]
}
