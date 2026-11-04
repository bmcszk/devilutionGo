package levels

// This file ports the crypt's hard-coded minisets and search/replace tables
// (levels/crypt.cpp:20-34, 56-190, 190-198, 357-500). Upstream data, so Go
// values with the source line cited.

// cryptDirtShape is the raw mega-tile id 22 the crypt stair minisets search
// for. The crypt tile enum has no name for it, so upstream writes the literal
// (levels/crypt.cpp:26-27).
const cryptDirtShape Tile = 22

// cryptStairsTown is the stairs-up miniset of the first crypt level
// (levels/crypt.cpp:56-76).
var cryptStairsTown = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{cryptDirtShape, cryptDirtShape, cryptDirtShape, cryptDirtShape},
		{cryptDirtShape, cryptDirtShape, cryptDirtShape, cryptDirtShape},
		{cryptHWall, cryptHWall, cryptHWall, cryptHWall},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
	},
	Replace: [6][6]Tile{
		{0, 62, cryptDirtHWallEnd, 0},
		{0, 61, cryptDirtHwall, 0},
		{63, 64, 65, 0},
		{66, 67, 68, 0},
		{0, 0, 0, 0},
	},
}

// cryptStairsUpHF is the stairs-up miniset of the Hellfire crypt levels
// (levels/crypt.cpp:36-54).
var cryptStairsUpHF = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{cryptDirtShape, cryptDirtShape, cryptDirtShape, cryptDirtShape},
		{cryptDirtShape, cryptDirtShape, cryptDirtShape, cryptDirtShape},
		{cryptHWall, cryptHWall, cryptHWall, cryptHWall},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
	},
	Replace: [6][6]Tile{
		{0, 54, cryptDirtHWallEnd, 0},
		{0, 53, cryptDirtHwall, 0},
		{55, 56, 57, 0},
		{58, 59, 60, 0},
		{0, 0, 0, 0},
	},
}

// cryptStairDown is the crypt stairs-down miniset (levels/crypt.cpp:77-96).
var cryptStairDown = Miniset{
	Size: Size{Width: 4, Height: 5},
	Search: [6][6]Tile{
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor, cryptFloor},
	},
	Replace: [6][6]Tile{
		{0, 0, 52, 0},
		{0, 48, 51, 0},
		{0, 47, 50, 0},
		{45, 46, 49, 0},
		{0, 0, 0, 0},
	},
}

// cryptVWallSection is the vertical wall variation miniset
// (levels/crypt.cpp:98-108).
var cryptVWallSection = Miniset{
	Size: Size{Width: 1, Height: 3},
	Search: [6][6]Tile{
		{cryptVWall},
		{cryptVWall},
		{cryptVWall},
	},
	Replace: [6][6]Tile{
		{cryptVWall7},
		{cryptVWall6},
		{cryptVWall5},
	},
}

// cryptHWallSection is the horizontal wall variation miniset
// (levels/crypt.cpp:109-113).
var cryptHWallSection = Miniset{
	Size:    Size{Width: 3, Height: 1},
	Search:  [6][6]Tile{{cryptHWall, cryptHWall, cryptHWall}},
	Replace: [6][6]Tile{{cryptHWall7, cryptHWall6, cryptHWall5}},
}

// cryptFloorLave is the lava pool miniset (levels/crypt.cpp:114-126).
var cryptFloorLave = Miniset{
	Size: Size{Width: 3, Height: 3},
	Search: [6][6]Tile{
		{cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor},
		{cryptFloor, cryptFloor, cryptFloor},
	},
	Replace: [6][6]Tile{
		{0, 0, 0},
		{0, cryptFloor10, 0},
		{0, 0, 0},
	},
}

// cryptSubstitutions2Sets are the pillar and star minisets CryptSubstitions2
// places in upstream order (levels/crypt.cpp:653-666). Only 167 has a name in
// the crypt tile enum; 168-172 are the unnamed pillar and star ids upstream
// writes as literals (levels/crypt.cpp:130-189).
var cryptSubstitutions2Sets = []Miniset{
	cryptPillarMiniset(cryptPillarHalf),
	cryptPillarMiniset(168),
	cryptPillarMiniset(169),
	cryptPillarMiniset(170),
	cryptPillarMiniset(171),
	cryptPillarMiniset(172),
}

// cryptStatues are the statue substitutions (levels/crypt.cpp:362-369).
var cryptStatues = []replaceTile{
	{cryptVWall, cryptVDemon},
	{cryptVWall, cryptVSuccubus},
	{cryptHWall, cryptHDemon},
	{cryptHWall, cryptHSuccubus},
}

// cryptCrackedTiles, cryptBrokenTiles and cryptLeakingTiles are the three
// damage families the per-level passes apply (levels/crypt.cpp:371-436).
var cryptCrackedTiles = []replaceTile{
	{cryptVWall, cryptVWall2},
	{cryptHWall, cryptHWall2},
	{cryptCorner, cryptCorner2},
	{cryptDWall, cryptDWall2},
	{cryptDArch, cryptDArch2},
	{cryptVWallEnd, cryptVWallEnd2},
	{cryptHWallEnd, cryptHWallEnd2},
	{cryptHArchEnd, cryptHArchEnd2},
	{cryptVArchEnd, cryptVArchEnd2},
	{cryptHArchVWall, cryptHArchVWall2},
	{cryptVArch, cryptVArch2},
	{cryptHArch, cryptHArch2},
	{cryptFloor, cryptFloor2},
	{cryptHWallVArch, cryptHWallVArch2},
	{cryptPillar, cryptPillar3},
	{cryptPillar1, cryptPillar4},
	{cryptPillar2, cryptPillar5},
}

var cryptBrokenTiles = []replaceTile{
	{cryptVWall, cryptVWall3},
	{cryptHWall, cryptHWall3},
	{cryptCorner, cryptCorner3},
	{cryptDWall, cryptDWall3},
	{cryptDArch, cryptDArch3},
	{cryptVWallEnd, cryptVWallEnd3},
	{cryptHWallEnd, cryptHWallEnd3},
	{cryptHArchEnd, cryptHArchEnd3},
	{cryptVArchEnd, cryptVArchEnd3},
	{cryptHArchVWall, cryptHArchVWall3},
	{cryptVArch, cryptVArch3},
	{cryptHArch, cryptHArch3},
	{cryptFloor, cryptFloor3},
	{cryptHWallVArch, cryptHWallVArch3},
	{cryptPillar, cryptPillar6},
	{cryptPillar1, cryptPillar7},
	{cryptPillar2, cryptPillar8},
}

var cryptLeakingTiles = []replaceTile{
	{cryptVWall, cryptVWall4},
	{cryptHWall, cryptHWall4},
	{cryptCorner, cryptCorner4},
	{cryptDWall, cryptDWall4},
	{cryptDArch, cryptDArch4},
	{cryptVWallEnd, cryptVWallEnd4},
	{cryptHWallEnd, cryptHWallEnd4},
	{cryptHArchEnd, cryptHArchEnd4},
	{cryptVArchEnd, cryptVArchEnd4},
	{cryptHArchVWall, cryptHArchVWall4},
	{cryptVArch, cryptVArch4},
	{cryptHArch, cryptHArch4},
	{cryptFloor, cryptFloor4},
	{cryptHWallVArch, cryptHWallVArch4},
	{cryptPillar, cryptPillar9},
	{cryptPillar1, cryptPillar10},
	{cryptPillar2, cryptPillar11},
}

// cryptSubstitutions1Tiles is the first variant family
// (levels/crypt.cpp:438-466).
var cryptSubstitutions1Tiles = []replaceTile{
	{cryptVArch, cryptVArch6},
	{cryptHArch, cryptHArch6},
	{cryptVArch, cryptVArch7},
	{cryptHArch, cryptHArch7},
	{cryptVWall5, cryptVWall8},
	{cryptVWall5, cryptVWall9},
	{cryptVWall6, cryptVWall10},
	{cryptVWall6, cryptVWall11},
	{cryptVWall7, cryptVWall12},
	{cryptVWall7, cryptVWall13},
	{cryptHWall5, cryptHWall8},
	{cryptHWall5, cryptHWall9},
	{cryptHWall5, cryptHWall10},
	{cryptHWall5, cryptHWall11},
	{cryptHWall5, cryptHWall12},
	{cryptHWall5, cryptHWall13},
	{cryptFloor7, cryptFloor15},
	{cryptFloor7, cryptFloor16},
	{cryptFloor6, cryptFloor17},
	{cryptPillar, cryptPillar12},
	{cryptFloor8, cryptFloor18},
	{cryptFloor8, cryptFloor19},
	{cryptFloor9, cryptFloor20},
	{cryptFloor10, cryptFloor21},
	{cryptFloor10, cryptFloor22},
	{cryptFloor10, cryptFloor23},
}

// cryptSubstitution1Floor is the floor family CryptSubstitions2 applies
// (levels/crypt.cpp:468-475).
var cryptSubstitution1Floor = []replaceTile{
	{cryptFloor, cryptFloor11},
	{cryptFloor, cryptFloor12},
	{cryptFloor, cryptFloor13},
	{cryptFloor, cryptFloor14},
}

// cryptSubstitution2Floor is the plain floor family CryptFloor applies
// (levels/crypt.cpp:477-484).
var cryptSubstitution2Floor = []replaceTile{
	{cryptFloor, cryptFloor6},
	{cryptFloor, cryptFloor7},
	{cryptFloor, cryptFloor8},
	{cryptFloor, cryptFloor9},
}

// replaceTile is upstream `struct ReplaceTile`: one 1x1 substitution
// (levels/crypt.cpp:357-360).
type replaceTile struct {
	search  Tile
	replace Tile
}

// cryptPillarMiniset builds one of the six 3x3 pillar minisets
// (levels/crypt.cpp:127-189).
func cryptPillarMiniset(tile Tile) Miniset {
	return Miniset{
		Size: Size{Width: 3, Height: 3},
		Search: [6][6]Tile{
			{cryptFloor, cryptFloor, cryptFloor},
			{cryptFloor, cryptFloor, cryptFloor},
			{cryptFloor, cryptFloor, cryptFloor},
		},
		Replace: [6][6]Tile{
			{0, 0, 0},
			{0, tile, 0},
			{0, 0, 0},
		},
	}
}
