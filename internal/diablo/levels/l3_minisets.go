package levels

// This file holds the caves minisets: the stairs, the isles River() floods,
// the stalagmite and wall-crack decorations, and the crevices the caves
// decoration pass scatters (levels/drlg_l3.cpp:35-430). Search and replace
// are indexed [y][x] with 0 as the wildcard, exactly as upstream.

// l3StairsUp is the caves "Stairs up" miniset (levels/drlg_l3.cpp:35).
var l3StairsUp = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 8, 0},
		{10, 10, 0},
		{7, 7, 0},
	},
	Replace: [6][6]Tile{
		{51, 50, 0},
		{48, 49, 0},
		{0, 0, 0},
	},
}

// l3StairsDown is the caves "Stairs down" miniset (levels/drlg_l3.cpp:62).
var l3StairsDown = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 9, 7},
		{8, 9, 7},
		{0, 0, 0},
	},
	Replace: [6][6]Tile{
		{0, 47, 0},
		{0, 46, 0},
		{0, 0, 0},
	},
}

// l3TownWarp is the caves "Stairs up to town" miniset (levels/drlg_l3.cpp:89).
var l3TownWarp = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 8, 0},
		{10, 10, 0},
		{7, 7, 0},
	},
	Replace: [6][6]Tile{
		{125, 125, 0},
		{125, 125, 0},
		{0, 0, 0},
	},
}

// l3Tite1 is the caves "Stalagmite white stalactite 1" miniset (levels/drlg_l3.cpp:116).
var l3Tite1 = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{0, 57, 58, 0},
		{0, 56, 55, 0},
		{0, 0, 0, 0},
	},
}

// l3Tite2 is the caves "Stalagmite white stalactite 2" miniset (levels/drlg_l3.cpp:132).
var l3Tite2 = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{0, 61, 62, 0},
		{0, 60, 59, 0},
		{0, 0, 0, 0},
	},
}

// l3Tite3 is the caves "Stalagmite white stalactite 3" miniset (levels/drlg_l3.cpp:148).
var l3Tite3 = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{0, 65, 66, 0},
		{0, 64, 63, 0},
		{0, 0, 0, 0},
	},
}

// l3Tite6 is the caves "Stalagmite white stalactite horizontal" miniset (levels/drlg_l3.cpp:164).
var l3Tite6 = Miniset{
	Size: Size{Width: 5, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7, 7},
		{7, 7, 7, 0, 7},
		{7, 7, 7, 0, 7},
		{7, 7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0, 0},
		{0, 77, 78, 0, 0},
		{0, 76, 74, 75, 0},
		{0, 0, 0, 0, 0},
	},
}

// l3Tite7 is the caves "Stalagmite white stalactite vertical" miniset (levels/drlg_l3.cpp:180).
var l3Tite7 = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 0, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0, 0},
		{0, 83, 0, 0},
		{0, 82, 80, 0},
		{0, 81, 79, 0},
		{0, 0, 0, 0},
	},
}

// l3Tite8 is the caves "Stalagmite 1" miniset (levels/drlg_l3.cpp:198).
var l3Tite8 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 52, 0},
		{0, 0, 0},
	},
}

// l3Tite9 is the caves "Stalagmite 2" miniset (levels/drlg_l3.cpp:212).
var l3Tite9 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 53, 0},
		{0, 0, 0},
	},
}

// l3Tite10 is the caves "Stalagmite 3" miniset (levels/drlg_l3.cpp:226).
var l3Tite10 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 54, 0},
		{0, 0, 0},
	},
}

// l3Tite11 is the caves "Stalagmite 4" miniset (levels/drlg_l3.cpp:240).
var l3Tite11 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 67, 0},
		{0, 0, 0},
	},
}

// l3Tite12 is the caves "Stalagmite on vertical wall" miniset (levels/drlg_l3.cpp:254).
var l3Tite12 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{9, 7},
	},
	Replace: [6][6]Tile{
		{68, 0},
	},
}

// l3Tite13 is the caves "Stalagmite on horizontal wall" miniset (levels/drlg_l3.cpp:260).
var l3Tite13 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{10},
		{7},
	},
	Replace: [6][6]Tile{
		{69},
		{0},
	},
}

// l3Crev1 is the caves "Cracked vertical wall 1" miniset (levels/drlg_l3.cpp:272).
var l3Crev1 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{8, 7},
	},
	Replace: [6][6]Tile{
		{84, 85},
	},
}

// l3Crev2 is the caves "Cracked vertical wall - north corner" miniset (levels/drlg_l3.cpp:278).
var l3Crev2 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{8, 11},
	},
	Replace: [6][6]Tile{
		{86, 87},
	},
}

// l3Crev3 is the caves "Cracked horizontal wall 1" miniset (levels/drlg_l3.cpp:284).
var l3Crev3 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{8},
		{10},
	},
	Replace: [6][6]Tile{
		{89},
		{88},
	},
}

// l3Crev4 is the caves "Cracked vertical wall 2" miniset (levels/drlg_l3.cpp:296).
var l3Crev4 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{8, 7},
	},
	Replace: [6][6]Tile{
		{90, 91},
	},
}

// l3Crev5 is the caves "Cracked horizontal wall - north corner" miniset (levels/drlg_l3.cpp:302).
var l3Crev5 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{8},
		{11},
	},
	Replace: [6][6]Tile{
		{92},
		{93},
	},
}

// l3Crev6 is the caves "Cracked horizontal wall 2" miniset (levels/drlg_l3.cpp:314).
var l3Crev6 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{8},
		{10},
	},
	Replace: [6][6]Tile{
		{95},
		{94},
	},
}

// l3Crev7 is the caves "Cracked vertical wall - west corner" miniset (levels/drlg_l3.cpp:326).
var l3Crev7 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{8, 7},
	},
	Replace: [6][6]Tile{
		{96, 101},
	},
}

// l3Crev8 is the caves "Cracked horizontal wall - north" miniset (levels/drlg_l3.cpp:332).
var l3Crev8 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{2},
		{8},
	},
	Replace: [6][6]Tile{
		{102},
		{97},
	},
}

// l3Crev9 is the caves "Cracked vertical wall - east corner" miniset (levels/drlg_l3.cpp:344).
var l3Crev9 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{3, 8},
	},
	Replace: [6][6]Tile{
		{103, 98},
	},
}

// l3Crev10 is the caves "Cracked vertical wall - west" miniset (levels/drlg_l3.cpp:350).
var l3Crev10 = Miniset{
	Size: Size{Width: 2, Height: 1},
	Search: [6][6]Tile{
		{4, 8},
	},
	Replace: [6][6]Tile{
		{104, 99},
	},
}

// l3Crev11 is the caves "Cracked horizontal wall - south corner" miniset (levels/drlg_l3.cpp:356).
var l3Crev11 = Miniset{
	Size: Size{Width: 1, Height: 2},
	Search: [6][6]Tile{
		{6},
		{8},
	},
	Replace: [6][6]Tile{
		{105},
		{100},
	},
}

// l3Isle1 is the caves "Replace broken wall with floor 1" miniset (levels/drlg_l3.cpp:368).
var l3Isle1 = Miniset{
	Size: Size{Width: 2, Height: 3},
	Search: [6][6]Tile{
		{5, 14},
		{4, 9},
		{13, 12},
	},
	Replace: [6][6]Tile{
		{7, 7},
		{7, 7},
		{7, 7},
	},
}

// l3Isle2 is the caves "Replace small wall with floor 2" miniset (levels/drlg_l3.cpp:382).
var l3Isle2 = Miniset{
	Size: Size{Width: 3, Height: 2},
	Search: [6][6]Tile{
		{5, 2, 14},
		{13, 10, 12},
	},
	Replace: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
	},
}

// l3Isle3 is the caves "Replace small wall with lava 1" miniset (levels/drlg_l3.cpp:394).
var l3Isle3 = Miniset{
	Size: Size{Width: 2, Height: 3},
	Search: [6][6]Tile{
		{5, 14},
		{4, 9},
		{13, 12},
	},
	Replace: [6][6]Tile{
		{29, 30},
		{25, 28},
		{31, 32},
	},
}

// l3Isle4 is the caves "Replace small wall with lava 2" miniset (levels/drlg_l3.cpp:408).
var l3Isle4 = Miniset{
	Size: Size{Width: 3, Height: 2},
	Search: [6][6]Tile{
		{5, 2, 14},
		{13, 10, 12},
	},
	Replace: [6][6]Tile{
		{29, 26, 30},
		{31, 27, 32},
	},
}

// l3Isle5 is the caves "Replace small wall with floor 3" miniset (levels/drlg_l3.cpp:420).
var l3Isle5 = Miniset{
	Size: Size{Width: 2, Height: 2},
	Search: [6][6]Tile{
		{5, 14},
		{13, 12},
	},
	Replace: [6][6]Tile{
		{7, 7},
		{7, 7},
	},
}
