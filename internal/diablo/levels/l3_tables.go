package levels

// This file holds the caves generator's lookup tables: the 2x2 pattern
// converter MakeMegas indexes, the flood-fill direction masks Spawn and
// SpawnEdge read, and the lava pool substitution alphabet
// (levels/drlg_l3.cpp).

// The two cave tile ids every caves pass keys on: the open floor the level is
// mostly made of, and the dirt ceiling/mass that fills the rest
// (levels/drlg_l3.cpp:717-760, 1442-1500).
const (
	caveFloor   Tile = 7
	caveCeiling Tile = 8
	// caveLavaGround is tile 33, the ground lava PoolFix writes over a ceiling
	// tile that touches a pool (levels/drlg_l3.cpp:1520).
	caveLavaGround Tile = 33
)

// The ground lava tile range: floor lava runs from 25 to 41 and only those
// tiles count as pool rim (levels/drlg_l3.cpp:1519).
const (
	l3LavaGroundMin Tile = 25
	l3LavaGroundMax Tile = 41
)

// l3ConvTbl maps the sixteen possible patterns of a 2x2 area, where each cell
// either contains a wall or it does not, to a mega tile
// (levels/drlg_l3.cpp:31-34).
var l3ConvTbl = [16]Tile{8, 11, 3, 10, 1, 9, 12, 12, 6, 13, 4, 13, 2, 14, 5, 7}

// l3SpawnWarpDirs is Spawn's spawntable: bit 3 is north, 2 south, 1 east and 0
// west for the steps that cross into the neighbouring pocket, bits 7..4 the
// same four directions for the steps that stay in the dirt layer
// (levels/drlg_l3.cpp:1321-1322).
var l3SpawnWarpDirs = [15]uint8{
	0x00, 0x0A, 0x03, 0x05, 0x0C, 0x06, 0x09, 0x00,
	0x00, 0x0C, 0x03, 0x06, 0x09, 0x0A, 0x05,
}

// l3SpawnEdgeDirs is SpawnEdge's spawntable, the same encoding for the pockets
// that touch the map edge (levels/drlg_l3.cpp:1272-1273).
var l3SpawnEdgeDirs = [15]uint8{
	0x00, 0x0A, 0x43, 0x05, 0x2C, 0x06, 0x09, 0x00,
	0x00, 0x1C, 0x83, 0x06, 0x09, 0x0A, 0x05,
}

// l3PoolSub replaces each dirt tile of a small disconnected pocket with the
// lava pool tile the table maps it to; 0 means "not a pool tile"
// (levels/drlg_l3.cpp:1444).
// The sixteenth entry is a port guard: upstream's table ends at index 14 and a
// raw tile id of 15 would read past it. MakeMegas only emits the ids of
// l3ConvTbl, so 15 never reaches the lookup (levels/drlg_l3.cpp:1444).
var l3PoolSub = [16]Tile{0, 35, 26, 36, 25, 29, 34, 7, 33, 28, 27, 37, 32, 31, 30, 0}
