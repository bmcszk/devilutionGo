package levels

// cathedralSpecial maps a `dPiece` value to the `dSpecial` marker the cathedral
// InitDungeonPieces writes for it (levels/drlg_l1.cpp:407-425). The upstream
// IsAnyOf chains are disjoint, so one table is faithful.
var cathedralSpecial = map[uint16]int8{
	11:  1, // drlg_l1.cpp:407
	70:  1,
	320: 1,
	210: 1,
	340: 1,
	417: 1,
	10:  2, // drlg_l1.cpp:409
	248: 2,
	324: 2,
	343: 2,
	330: 2,
	420: 2,
	252: 3, // drlg_l1.cpp:411
	254: 4, // drlg_l1.cpp:413
	258: 5, // drlg_l1.cpp:415
	266: 6, // drlg_l1.cpp:417
}

// initCathedralPieces is InitDungeonPieces (levels/drlg_l1.cpp:402-429): it
// reads the finished world-tile grid back to find the doors and arches.
func initCathedralPieces(grid *PieceGrid) {
	for x := range MAXDUNX {
		for y := range MAXDUNY {
			if pc, ok := cathedralSpecial[grid.Piece[x][y]]; ok {
				grid.Special[x][y] = pc
			}
		}
	}
}

// initCryptPieces is InitCryptPieces (levels/crypt.cpp:668-679): the Hellfire
// crypt has its own two markers.
func initCryptPieces(grid *PieceGrid) {
	for x := range MAXDUNX {
		for y := range MAXDUNY {
			switch grid.Piece[x][y] {
			case 76: // crypt.cpp:672
				grid.Special[x][y] = 1
			case 79: // crypt.cpp:674
				grid.Special[x][y] = 2
			default:
			}
		}
	}
}
