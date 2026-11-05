package levels

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
	// Special is `dSpecial`: the door and arch markers the level-1 and level-2
	// `InitDungeonPieces` write (levels/drlg_l1.cpp:402-429,
	// levels/drlg_l2.cpp:1657-1681).
	Special [MAXDUNX][MAXDUNY]int8
}

// ExpandPieces is DRLG_LPass3 (levels/gendung.cpp:581-621), the expansion every
// generator runs last (levels/drlg_l1.cpp:1241, drlg_l2.cpp:2823,
// drlg_l3.cpp:2144, drlg_l4.cpp:1223). It is a stub: the port lands next
// session, so it produces an empty grid, which is what the dPiece parity gate
// reports.
func ExpandPieces(d *Dungeon, megaTiles []MegaTile) PieceGrid {
	_, _ = d, megaTiles

	return PieceGrid{}
}
