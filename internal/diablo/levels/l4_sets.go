package levels

// This file holds the five hell minisets PlaceStairs stamps: the two stair
// cases, the pentagram, its portal form and the stairs up to town
// (levels/drlg_l4.cpp:45-129).

// l4UStairs is the stairs up (levels/drlg_l4.cpp:45-58, L4USTAIRS).
var l4UStairs = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{36, 38, 35, 0},
		{37, 34, 33, 32},
		{0, 0, 31, 0},
		{0, 0, 0, 0},
	},
}

// l4TWarp is the stairs up to town (levels/drlg_l4.cpp:59-73, L4TWARP).
var l4TWarp = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
		{6, 6, 6, 6},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{134, 136, 133, 0},
		{135, 132, 131, 130},
		{0, 0, 129, 0},
		{0, 0, 0, 0},
	},
}

// l4DStairs is the stairs down (levels/drlg_l4.cpp:74-89, L4DSTAIRS).
var l4DStairs = Miniset{
	Size: Size{Width: 5, Height: 5},
	Search: [6][6]Tile{
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0, 0},
		{0, 0, 45, 41, 0},
		{0, 44, 43, 40, 0},
		{0, 46, 42, 39, 0},
		{0, 0, 0, 0, 0},
	},
}

// l4Penta is the pentagram the closed hell gate is drawn with
// (levels/drlg_l4.cpp:90-104, L4PENTA).
var l4Penta = Miniset{
	Size: Size{Width: 5, Height: 5},
	Search: [6][6]Tile{
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0, 0},
		{0, 98, 100, 103, 0},
		{0, 99, 102, 105, 0},
		{0, 101, 104, 106, 0},
		{0, 0, 0, 0, 0},
	},
}

// l4Penta2 is the pentagram portal the open hell gate is drawn with
// (levels/drlg_l4.cpp:105-119, L4PENTA2).
var l4Penta2 = Miniset{
	Size: Size{Width: 5, Height: 5},
	Search: [6][6]Tile{
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
		{6, 6, 6, 6, 6},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0, 0},
		{0, 107, 109, 112, 0},
		{0, 108, 111, 114, 0},
		{0, 110, 113, 115, 0},
		{0, 0, 0, 0, 0},
	},
}
