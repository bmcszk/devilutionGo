package levels

// This file holds the nest minisets: the stairs, the isles the nest
// decoration scatters and the slime pool patterns PlaceSlimePool stamps
// (levels/drlg_l3.cpp:48-115, 431-456, 457-645, 646-710). Search and replace
// are indexed [y][x] with 0 as the wildcard, exactly as upstream.

// hiveStairsUp is the nest "stairs up" miniset (levels/drlg_l3.cpp:48).
var hiveStairsUp = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 8, 0},
		{10, 10, 0},
		{7, 7, 0},
	},
	Replace: [6][6]Tile{
		{20, 19, 0},
		{17, 18, 0},
		{0, 0, 0},
	},
}

// hiveStairsDown is the nest "stairs down" miniset (levels/drlg_l3.cpp:75).
var hiveStairsDown = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 9, 7},
		{8, 9, 7},
		{0, 0, 0},
	},
	Replace: [6][6]Tile{
		{0, 16, 0},
		{0, 15, 0},
		{0, 0, 0},
	},
}

// hiveTownWarp is the nest "stairs up to town" miniset (levels/drlg_l3.cpp:102).
var hiveTownWarp = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{8, 8, 0},
		{10, 10, 0},
		{7, 7, 0},
	},
	Replace: [6][6]Tile{
		{24, 23, 0},
		{21, 22, 0},
		{0, 0, 0},
	},
}

// hivePattern9 is the nest decoration pattern HivePattern9 (levels/drlg_l3.cpp:431).
var hivePattern9 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 126, 0},
		{0, 0, 0},
	},
}

// hivePattern10 is the nest decoration pattern HivePattern10 (levels/drlg_l3.cpp:444).
var hivePattern10 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, 124, 0},
		{0, 0, 0},
	},
}

// hivePattern29 is the nest tunnel pattern HivePattern29 (levels/drlg_l3.cpp:457).
var hivePattern29 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{67, 0, 0},
		{66, 51, 0},
		{0, 0, 0},
	},
}

// hivePattern30 is the nest tunnel pattern HivePattern30 (levels/drlg_l3.cpp:470).
var hivePattern30 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{69, 0, 0},
		{68, 52, 0},
		{0, 0, 0},
	},
}

// hivePattern31 is the nest tunnel pattern HivePattern31 (levels/drlg_l3.cpp:483).
var hivePattern31 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{70, 0, 0},
		{71, 53, 0},
		{0, 0, 0},
	},
}

// hivePattern32 is the nest tunnel pattern HivePattern32 (levels/drlg_l3.cpp:496).
var hivePattern32 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{73, 0, 0},
		{72, 54, 0},
		{0, 0, 0},
	},
}

// hivePattern33 is the nest tunnel pattern HivePattern33 (levels/drlg_l3.cpp:509).
var hivePattern33 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{75, 0, 0},
		{74, 55, 0},
		{0, 0, 0},
	},
}

// hivePattern34 is the nest tunnel pattern HivePattern34 (levels/drlg_l3.cpp:522).
var hivePattern34 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{77, 0, 0},
		{76, 56, 0},
		{0, 0, 0},
	},
}

// hivePattern35 is the nest tunnel pattern HivePattern35 (levels/drlg_l3.cpp:535).
var hivePattern35 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{79, 0, 0},
		{78, 57, 0},
		{0, 0, 0},
	},
}

// hivePattern36 is the nest tunnel pattern HivePattern36 (levels/drlg_l3.cpp:548).
var hivePattern36 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{81, 0, 0},
		{80, 58, 0},
		{0, 0, 0},
	},
}

// hivePattern37 is the nest tunnel pattern HivePattern37 (levels/drlg_l3.cpp:561).
var hivePattern37 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{83, 0, 0},
		{82, 59, 0},
		{0, 0, 0},
	},
}

// hivePattern38 is the nest tunnel pattern HivePattern38 (levels/drlg_l3.cpp:574).
var hivePattern38 = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{7, 7, 7},
		{7, 7, 7},
		{7, 7, 7},
	},
	Replace: [6][6]Tile{
		{84, 0, 0},
		{85, 60, 0},
		{0, 0, 0},
	},
}

// hiveIsle1 is the nest isle miniset L6ISLE1 (levels/drlg_l3.cpp:587).
var hiveIsle1 = Miniset{
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

// hiveIsle2 is the nest isle miniset L6ISLE2 (levels/drlg_l3.cpp:600).
var hiveIsle2 = Miniset{
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

// hiveIsle3 is the nest isle miniset L6ISLE3 (levels/drlg_l3.cpp:611).
var hiveIsle3 = Miniset{
	Size: Size{Width: 2, Height: 3},
	Search: [6][6]Tile{
		{5, 14},
		{4, 9},
		{13, 12},
	},
	Replace: [6][6]Tile{
		{107, 115},
		{119, 122},
		{131, 123},
	},
}

// hiveIsle4 is the nest isle miniset L6ISLE4 (levels/drlg_l3.cpp:624).
var hiveIsle4 = Miniset{
	Size: Size{Width: 3, Height: 2},
	Search: [6][6]Tile{
		{5, 2, 14},
		{13, 10, 12},
	},
	Replace: [6][6]Tile{
		{107, 120, 115},
		{131, 121, 123},
	},
}

// hiveIsle5 is the nest isle miniset L6ISLE5 (levels/drlg_l3.cpp:635).
var hiveIsle5 = Miniset{
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

// hivePattern39 is the nest slime pool pattern HivePattern39 (levels/drlg_l3.cpp:646).
var hivePattern39 = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 107, 115, 7},
		{7, 131, 123, 7},
		{7, 7, 7, 7},
	},
}

// hivePattern40 is the nest slime pool pattern HivePattern40 (levels/drlg_l3.cpp:661).
var hivePattern40 = Miniset{
	Size: Size{Width: 4, Height: 4},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 108, 7},
		{7, 109, 112, 7},
		{7, 7, 7, 7},
	},
}

// hivePattern41 is the nest slime pool pattern HivePattern41 (levels/drlg_l3.cpp:676).
var hivePattern41 = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 107, 115, 7},
		{7, 119, 122, 7},
		{7, 131, 123, 7},
		{7, 7, 7, 7},
	},
}

// hivePattern42 is the nest slime pool pattern HivePattern42 (levels/drlg_l3.cpp:693).
var hivePattern42 = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
		{7, 7, 7, 7},
	},
	Replace: [6][6]Tile{
		{7, 7, 7, 7},
		{7, 126, 108, 7},
		{7, 7, 117, 7},
		{7, 109, 112, 7},
		{7, 7, 7, 7},
	},
}
