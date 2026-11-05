package levels

// pass3DirtTile is the pMegaTiles index each generator passes to DRLG_LPass3 as
// `lv`, resolved from the call site's own dirt tile id minus one.
const (
	// Dirt == 22 (levels/drlg_l1.cpp:120), passed as Dirt - 1 (drlg_l1.cpp:1241).
	pass3DirtTileCathedral = 22 - 1
	// 12 - 1 (levels/drlg_l2.cpp:2858).
	pass3DirtTileCatacombs = 12 - 1
	// 8 - 1 (levels/drlg_l3.cpp:2151).
	pass3DirtTileCaves = 8 - 1
	// 30 - 1 (levels/drlg_l4.cpp:1231).
	pass3DirtTileHell = 30 - 1
	// pass3Border is the world-tile border DRLG_LPass3 leaves at dirt
	// (levels/gendung.cpp:616-617).
	pass3Border = 16
)

// MegaTile is one pMegaTiles entry (levels/dun_tile_data.hpp:56-61): the four
// world tiles a mega tile expands to.
type MegaTile struct {
	Micro1 uint16
	Micro2 uint16
	Micro3 uint16
	Micro4 uint16
}

// PieceGrid is the Pass3 output: `dPiece` and `dSpecial`, both [x][y] like
// every other engine grid (levels/dun_tile_data.cpp:51,62).
type PieceGrid struct {
	// Piece is `dPiece`: the world-tile ids of the whole level.
	Piece [MAXDUNX][MAXDUNY]uint16
	// Special is `dSpecial`: the door and arch markers the family
	// InitDungeonPieces writes (levels/drlg_l1.cpp:402-429,
	// levels/drlg_l2.cpp:1663-1689, levels/crypt.cpp:668-679).
	Special [MAXDUNX][MAXDUNY]int8
}

// ExpandPieces is the level's Pass3: DRLG_LPass3 (levels/gendung.cpp:591-624)
// expands the mega-tile layer into the world-tile grid, then the family's
// InitDungeonPieces (levels/drlg_l1.cpp:1239-1247, drlg_l2.cpp:2858-2861)
// marks the door and arch cells off that grid.
func ExpandPieces(d *Dungeon, megaTiles []MegaTile) PieceGrid {
	grid := PieceGrid{}

	fillDirtPieces(&grid, dirtMegaTile(d.LevelType, megaTiles))
	expandMegaTiles(&grid, d.Tiles, megaTiles)
	initSpecialPieces(&grid, d.LevelType)

	return grid
}

// dirtMegaTile resolves the `lv` argument of the generator's DRLG_LPass3 call
// (levels/gendung.cpp:593) in the level's own pMegaTiles. Town and Nest have no
// Pass3 generator, so they fall through to index 0.
func dirtMegaTile(levelType LevelType, megaTiles []MegaTile) MegaTile {
	index := pass3DirtIndex(levelType)
	if index >= len(megaTiles) {
		return MegaTile{}
	}

	return megaTiles[index]
}

// pass3DirtIndex is the `lv` argument dirtMegaTile resolves: the level's own
// dirt tile id minus one (levels/gendung.cpp:593, drlg_l1.cpp:1241,
// drlg_l2.cpp:2858, drlg_l3.cpp:2151, drlg_l4.cpp:1231). Town and Nest have no
// Pass3 generator, so they fall through to index 0.
func pass3DirtIndex(levelType LevelType) int {
	switch levelType {
	case Cathedral, Crypt:
		return pass3DirtTileCathedral
	case Catacombs:
		return pass3DirtTileCatacombs
	case Caves:
		return pass3DirtTileCaves
	case Hell:
		return pass3DirtTileHell
	default:
		return 0
	}
}

// fillDirtPieces is the first block of DRLG_LPass3 (levels/gendung.cpp:594-612):
// every world tile of the level starts as the dirt mega tile.
func fillDirtPieces(grid *PieceGrid, dirt MegaTile) {
	for y := range MAXDUNY / 2 {
		for x := range MAXDUNX / 2 {
			placeMegaTile(grid, 2*x, 2*y, dirt)
		}
	}
}

// expandMegaTiles is the second block of DRLG_LPass3 (levels/gendung.cpp:614-621):
// the 40x40 mega-tile layer overwrites the middle of the grid, two world tiles
// per mega tile starting at the 16-tile border.
func expandMegaTiles(grid *PieceGrid, tiles [DMAXX][DMAXY]Tile, megaTiles []MegaTile) {
	for j := range DMAXY {
		for i := range DMAXX {
			// Upstream indexes pMegaTiles[dungeon[i][j] - 1] unguarded; a zero
			// mega tile never reaches Pass3 in the game, so the port offsets the
			// out-of-range read instead of reproducing it.
			id := int(tiles[i][j]) - 1
			if id < 0 || id >= len(megaTiles) {
				continue
			}
			placeMegaTile(grid, 2*i+pass3Border, 2*j+pass3Border, megaTiles[id])
		}
	}
}

// placeMegaTile writes one mega tile's four micro tiles at the world-tile
// position (levels/gendung.cpp:600-603).
func placeMegaTile(grid *PieceGrid, x, y int, mega MegaTile) {
	grid.Piece[x][y] = mega.Micro1
	grid.Piece[x+1][y] = mega.Micro2
	grid.Piece[x][y+1] = mega.Micro3
	grid.Piece[x+1][y+1] = mega.Micro4
}

// initSpecialPieces is the family InitDungeonPieces each generator runs right
// after its DRLG_LPass3 call. Caves and hell have none: their Pass3 is the
// expansion alone (levels/drlg_l3.cpp:2148-2152, levels/drlg_l4.cpp:1229-1233).
func initSpecialPieces(grid *PieceGrid, levelType LevelType) {
	switch levelType {
	case Cathedral:
		initCathedralPieces(grid)
	case Crypt:
		initCryptPieces(grid)
	case Catacombs:
		initCatacombsPieces(grid)
	default:
	}
}
