package levels

// The Hellfire hive/grave door patches. Each cell carries the closed and the
// open value CreateTown writes for it (-1: that state leaves the cell alone);
// the coordinates and values are transcribed from levels/town.cpp:281-360.

// townHiveTiles is the dungeon cells of both hive door states.
var townHiveTiles = []townDoorCell{
	{35, 27, 18, -1},
	{36, 27, 63, 47},
}

// townHivePieces is the dPiece cells of both hive door states.
var townHivePieces = []townDoorCell{
	{78, 60, 1161, 1161},
	{79, 60, 1258, 1162},
	{78, 61, 1259, 1163},
	{79, 61, 1260, 1293},
	{78, 62, 1261, 1261},
	{79, 62, 1262, 1295},
	{78, 63, 1263, 1263},
	{79, 63, 1264, 1296},
	{78, 64, 1265, 281},
	{79, 64, 1266, 1297},
	{78, 65, 1267, 283},
	{80, 60, 1268, 1298},
	{81, 60, 1269, -1},
	{80, 61, 1270, 1300},
	{81, 61, 1271, 1301},
	{82, 60, 1272, 1302},
	{83, 60, 1273, 1303},
	{82, 61, 1274, 1304},
	{83, 61, 1275, 1305},
	{80, 62, 1276, 1306},
	{81, 62, 1277, 1307},
	{80, 63, 1278, 1308},
	{81, 63, 1279, 1309},
	{80, 64, 1280, 1310},
	{81, 64, 1281, 1311},
	{80, 65, 1282, 1312},
	{81, 65, 1283, 1313},
	{82, 64, 1288, 1318},
	{83, 64, 1289, 1319},
	{82, 65, 1290, 1320},
	{83, 65, 1291, 1321},
	{82, 62, 1284, 1314},
	{83, 62, 1285, 1315},
	{82, 63, 1286, 1316},
	{83, 63, 1287, 1317},
	{84, 61, 279, 279},
	{84, 62, 280, 280},
	{84, 63, 279, 279},
	{84, 64, 10, 10},
	{85, 60, 11, 11},
	{85, 61, 12, 12},
	{85, 62, 13, 13},
	{85, 63, 14, 14},
	{85, 64, 15, 15},
	{86, 60, 16, 16},
	{86, 61, 17, 17},
	{79, 65, -1, 284},
}

// townGraveTiles is the dungeon cells of both grave door states.
var townGraveTiles = []townDoorCell{
	{14, 8, -1, 47},
	{14, 7, -1, 47},
}

// townGravePieces is the dPiece cells of both grave door states.
var townGravePieces = []townDoorCell{
	{36, 21, 1322, 1330},
	{37, 21, 1323, 1331},
	{36, 22, 1324, 1332},
	{37, 22, 1325, 1333},
	{36, 23, 1326, 1334},
	{37, 23, 1327, 1335},
	{36, 24, 1328, 1336},
	{37, 24, 1329, 1337},
	{35, 21, 1338, 1338},
	{34, 21, 1339, 1339},
}

// townDoorCell is one cell a Hellfire door patch writes, with the value each
// door state stores there (-1: not written).
type townDoorCell struct {
	X, Y         int
	Closed, Open int
}

// townCloseHive draws the closed hive door (levels/town.cpp:281-329).
func townCloseHive(d *Dungeon, grid *PieceGrid) {
	applyTownDoor(d, grid, townHiveTiles, townHivePieces, townDoorCell.closed)
}

// townOpenHive draws the opened hive door (levels/town.cpp:283-333).
func townOpenHive(d *Dungeon, grid *PieceGrid) {
	applyTownDoor(d, grid, townHiveTiles, townHivePieces, townDoorCell.open)
}

// townCloseGrave draws the closed grave door (levels/town.cpp:333-360).
func townCloseGrave(d *Dungeon, grid *PieceGrid) {
	applyTownDoor(d, grid, townGraveTiles, townGravePieces, townDoorCell.closed)
}

// townOpenGrave draws the opened grave door (levels/town.cpp:335-360).
func townOpenGrave(d *Dungeon, grid *PieceGrid) {
	applyTownDoor(d, grid, townGraveTiles, townGravePieces, townDoorCell.open)
}

// applyTownDoor writes one door state's tile and piece cells. The dPiece writes
// are the single-cell form of FillTile's block (levels/town.cpp:66-79).
func applyTownDoor(d *Dungeon, grid *PieceGrid, tiles, pieces []townDoorCell, pick func(townDoorCell) (int, bool)) {
	for _, cell := range tiles {
		if tile, ok := pick(cell); ok {
			d.Tiles[cell.X][cell.Y] = Tile(tile)
		}
	}

	for _, cell := range pieces {
		if piece, ok := pick(cell); ok {
			grid.Piece[cell.X][cell.Y] = uint16(piece)
		}
	}
}

// closed returns the closed door's value for one cell, false when that state
// writes nothing.
func (c townDoorCell) closed() (int, bool) {
	return c.Closed, c.Closed >= 0
}

// open returns the open door's value for one cell, false when that state writes
// nothing.
func (c townDoorCell) open() (int, bool) {
	return c.Open, c.Open >= 0
}
