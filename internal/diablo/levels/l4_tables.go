package levels

// This file holds the two lookup tables the hell generator reads: the 2x2 wall
// pattern table MakeDmt indexes and the base-tile table Substitution resolves
// random variants through (levels/drlg_l4.cpp:43, 132-148).

// l4ConvTbl is the 16 possible 2x2 SW-wall patterns mapped to their mega tile
// (levels/drlg_l4.cpp:43, L4ConvTbl). The index bits are, low to high:
// north-west, north-east, south-west, south-east.
var l4ConvTbl = [16]Tile{
	30, 6, 1, 6,
	2, 6, 6, 6,
	9, 6, 1, 6,
	2, 6, 3, 6,
}

// hellSkulls are the three skull tiles the second substitution pass draws from
// (levels/drlg_l4.cpp:863).
var hellSkulls = [3]Tile{95, 96, 97}

// l4BTypes maps a hell tile id to the undecorated tile it is a variant of
// (levels/drlg_l4.cpp:132-148, L4BTYPES). Substitution only substitutes tiles
// with a non-zero base, and picks the replacement among the ids that map to the
// same base. The table is 140 entries because that is upstream's size; every
// id above 89 has no base type.
var l4BTypes = [140]Tile{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	10, 11, 12, 13, 14, 15, 16, 17, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 6,
	6, 6, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 2, 1, 2, 1, 2, 1, 1, 2,
	2, 0, 0, 0, 0, 0, 0, 15, 16, 9,
	12, 4, 5, 7,
}
