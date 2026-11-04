package levels

// This file ports the hard-coded minisets the cathedral generator places
// (levels/drlg_l1.cpp:39-95, levels/crypt.cpp:20-34). They are upstream data,
// not fixture data, so they are Go values with the source line cited.

// stairSup is the stairs-up miniset on a corner wall
// (levels/drlg_l1.cpp:39-46).
var stairSup = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{catFloor, catFloor, catFloor, catFloor},
		{catHWall, catHWall, catHWall, catHWall},
		{catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor},
	},
	Replace: [6][6]Tile{
		{0, 66, catVWallEnd, 0},
		{63, 64, 65, 0},
		{0, 67, 68, 0},
		{0, 0, 0, 0},
	},
}

// l5StairSup is the stairs-up miniset of the original cathedral, whose search
// pattern runs through the dirt layer (levels/crypt.cpp:20-34).
var l5StairSup = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{catDirt, catDirt, catDirt, catDirt},
		{catHWall, catHWall, catHWall, catHWall},
		{catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor},
	},
	Replace: [6][6]Tile{
		{0, 66, catDirtHwallEnd, 0},
		{63, 64, 65, 0},
		{0, 67, 68, 0},
		{0, 0, 0, 0},
	},
}

// stairDown is the stairs-down miniset (levels/drlg_l1.cpp:55-67).
var stairDown = Miniset{
	Size: Size{Width: 4, Height: 3},
	Search: [6][6]Tile{
		{catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor},
	},
	Replace: [6][6]Tile{
		{62, 57, 58, 0},
		{61, 59, 60, 0},
		{0, 0, 0, 0},
	},
}

// pwaterIn is the Poisoned Water Supply entrance miniset
// (levels/drlg_l1.cpp:81-95).
var pwaterIn = Miniset{
	Size: Size{Width: 6, Height: 6},
	Search: [6][6]Tile{
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
		{catFloor, catFloor, catFloor, catFloor, catFloor, catFloor},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0, 0, 0},
		{0, 202, 200, 200, catVWallEnd2, 0},
		{0, 199, 203, 203, 83, 0},
		{0, 85, 206, 80, 81, 0},
		{0, 0, 134, 135, 0, 0},
		{0, 0, 0, 0, 0, 0},
	},
}
