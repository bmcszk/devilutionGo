package levels

// initCatacombsPieces is InitDungeonPieces (levels/drlg_l2.cpp:1663-1689): two
// passes over the finished world-tile grid, the first marking the door jamb
// tiles themselves, the second the two cells each arch reserves beside it.
func initCatacombsPieces(grid *PieceGrid) {
	markCatacombsJambs(grid)
	markCatacombsArches(grid)
}

// markCatacombsJambs is the first loop of InitDungeonPieces
// (levels/drlg_l2.cpp:1665-1676).
func markCatacombsJambs(grid *PieceGrid) {
	for y := range MAXDUNY {
		for x := range MAXDUNX {
			switch grid.Piece[x][y] {
			case 540, 177, 550: // drlg_l2.cpp:1668
				grid.Special[x][y] = 5
			case 541, 552: // drlg_l2.cpp:1670
				grid.Special[x][y] = 6
			default:
			}
		}
	}
}

// markCatacombsArches is the second loop of InitDungeonPieces
// (levels/drlg_l2.cpp:1678-1688). Upstream writes dSpecial[i][j + 1] and
// dSpecial[i + 1][j] unguarded; at the far edge those reads land outside the
// declared array, so the port clamps them instead of reproducing the overrun.
// The y-outer order is load-bearing: two neighbouring arches can claim the same
// cell, and there the later write wins.
func markCatacombsArches(grid *PieceGrid) {
	for y := range MAXDUNY {
		for x := range MAXDUNX {
			switch grid.Piece[x][y] {
			case 131: // drlg_l2.cpp:1680
				setSpecial(grid, x, y+1, 2)
				setSpecial(grid, x, y+2, 1)
			case 134, 138: // drlg_l2.cpp:1683
				setSpecial(grid, x+1, y, 3)
				setSpecial(grid, x+2, y, 4)
			default:
			}
		}
	}
}

// setSpecial writes one dSpecial marker when the cell is inside the grid.
func setSpecial(grid *PieceGrid, x, y int, pc int8) {
	if x >= MAXDUNX || y >= MAXDUNY {
		return
	}
	grid.Special[x][y] = pc
}
