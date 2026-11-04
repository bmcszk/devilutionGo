package levels

// This file holds the three hard-coded tile tables the cathedral generator's
// decoration passes read. They are upstream data, not fixture data: ported as Go
// values with the source region cited (levels/drlg_l1.cpp:183-293).

// shadowPatterns is ShadowPatterns, in upstream order (levels/drlg_l1.cpp:183-225).
var shadowPatterns = []shadowPattern{
	{strig: catHWallEnd, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor17, nv2: 0, nv3: catFloor15},
	{strig: catVCorner, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor17, nv2: 0, nv3: catFloor15},
	{strig: catPillar, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor18, nv2: 0, nv3: catFloor15},
	{strig: catDArch, s1: catFloor, s2: catFloor, s3: catFloor, nv1: catFloor21, nv2: catFloor13, nv3: catFloor12},
	{strig: catDArch, s1: catFloor, s2: catVWall, s3: catFloor, nv1: catFloor16, nv2: catVWall17, nv3: catFloor12},
	{strig: catDArch, s1: catFloor, s2: catFloor, s3: catHWall, nv1: catFloor16, nv2: catFloor13, nv3: catHWallShadow},
	{strig: catDArch, s1: 0, s2: catVWall, s3: catHWall, nv1: 0, nv2: catVWall17, nv3: catHWallShadow},
	{strig: catDArch, s1: catFloor, s2: catVArch, s3: catFloor, nv1: catFloor16, nv2: catVArch5, nv3: catFloor12},
	{strig: catDArch, s1: catFloor, s2: catFloor, s3: catHArch, nv1: catFloor16, nv2: catFloor13, nv3: catHArchShadow},
	{strig: catDArch, s1: catFloor, s2: catVArch, s3: catHArch, nv1: catFloor19, nv2: catVArch5, nv3: catHArchShadow},
	{strig: catDArch, s1: catFloor, s2: catVWall, s3: catHArch, nv1: catFloor16, nv2: catVWall17, nv3: catHArchShadow},
	{strig: catDArch, s1: catFloor, s2: catVArch, s3: catHWall, nv1: catFloor16, nv2: catVArch5, nv3: catHWallShadow},
	{strig: catVArchEnd, s1: catFloor, s2: catFloor, s3: catFloor, nv1: catFloor17, nv2: catFloor13, nv3: catFloor15},
	{strig: catVArchEnd, s1: catFloor, s2: catVWall, s3: catFloor, nv1: catFloor17, nv2: catVWall17, nv3: catFloor15},
	{strig: catVArchEnd, s1: catFloor, s2: catVArch, s3: catFloor, nv1: catFloor20, nv2: catVArch5, nv3: catFloor15},
	{strig: catHArchEnd, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor17, nv2: 0, nv3: catFloor12},
	{strig: catHArchEnd, s1: catFloor, s2: 0, s3: catHArch, nv1: catFloor16, nv2: 0, nv3: catHArchShadow},
	{strig: catHArchEnd, s1: 0, s2: 0, s3: catHWall, nv1: 0, nv2: 0, nv3: catHWallShadow},
	{strig: catVArch, s1: 0, s2: 0, s3: catFloor, nv1: 0, nv2: 0, nv3: catFloor12},
	{strig: catVArch, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor12, nv2: 0, nv3: catFloor12},
	{strig: catVArch, s1: catHWall, s2: 0, s3: catFloor, nv1: catHWallShadow, nv2: 0, nv3: catFloor12},
	{strig: catVArch, s1: catHArch, s2: 0, s3: catFloor, nv1: catHArchShadow, nv2: 0, nv3: catFloor12},
	{strig: catVArch, s1: catFloor, s2: catVArch, s3: catHArch, nv1: catFloor12, nv2: 0, nv3: catHArchShadow},
	{strig: catHWallVArch, s1: 0, s2: 0, s3: catFloor, nv1: 0, nv2: 0, nv3: catFloor12},
	{strig: catHWallVArch, s1: catFloor, s2: 0, s3: catFloor, nv1: catFloor12, nv2: 0, nv3: catFloor12},
	{strig: catHWallVArch, s1: catHWall, s2: 0, s3: catFloor, nv1: catHWallShadow, nv2: 0, nv3: catFloor12},
	{strig: catHWallVArch, s1: catHArch, s2: 0, s3: catFloor, nv1: catHArchShadow, nv2: 0, nv3: catFloor12},
	{strig: catHWallVArch, s1: catFloor, s2: catVArch, s3: catHArch, nv1: catFloor12, nv2: 0, nv3: catHArchShadow},
	{strig: catHArchVWall, s1: 0, s2: catFloor, s3: 0, nv1: 0, nv2: catFloor13, nv3: 0},
	{strig: catHArchVWall, s1: catFloor, s2: catFloor, s3: 0, nv1: catFloor13, nv2: catFloor13, nv3: 0},
	{strig: catHArchVWall, s1: 0, s2: catVWall, s3: 0, nv1: 0, nv2: catVWall17, nv3: 0},
	{strig: catHArchVWall, s1: catFloor, s2: catVArch, s3: 0, nv1: catFloor13, nv2: catVArch5, nv3: 0},
	{strig: catHArch, s1: 0, s2: catFloor, s3: 0, nv1: 0, nv2: catFloor13, nv3: 0},
	{strig: catHArch, s1: catFloor, s2: catFloor, s3: 0, nv1: catFloor13, nv2: catFloor13, nv3: 0},
	{strig: catHArch, s1: 0, s2: catVWall, s3: 0, nv1: 0, nv2: catVWall17, nv3: 0},
	{strig: catHArch, s1: catFloor, s2: catVArch, s3: 0, nv1: catFloor13, nv2: catVArch5, nv3: 0},
	{strig: catCorner, s1: catFloor, s2: catVArch, s3: catHArch, nv1: catFloor19, nv2: 0, nv3: 0},
}

// baseTypes is BaseTypes, the base tile of every cathedral tile id (levels/drlg_l1.cpp:227-254).
var baseTypes = [207]Tile{
	0, catVWall, catHWall, catCorner, catDWall, catDArch, catVWallEnd, catHWallEnd,
	catHArchEnd, catVArchEnd, catHArchVWall, catVArch, catHArch, catFloor, catHWallVArch, catPillar,
	catVCorner, catHCorner, 0, 0, 0, 0, 0, 0,
	0, catVWall, catHWall, catHArchVWall, catDWall, catDArch, catVWallEnd, catHWallEnd,
	catHArchEnd, catVArchEnd, catHArchVWall, catVArch, catHArch, catHWallVArch, catDArch, catHWallVArch,
	catHArchVWall, catDWall, catHWallVArch, catDWall, catDArch, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, catVWall,
	catHWall, catCorner, catDWall, catVWall, catVWallEnd, catHWallEnd, catVCorner, catHCorner,
	catHWall, catVWall, catVWall, catHWall, catHWall, catVWall, catVWall, catHWall,
	catHWall, catHWall, catHWall, catHWall, catVWall, catVWall, catVArch, catVWall,
	catFloor, catFloor, catFloor, catVWall, catHWall, catVWall, catHWall, catVWall,
	catHWall, catVWall, catHWall, catHWall, catHWall, catHWall, catHArch, 0,
	0, catVArch, catVWall, catVArch, catVWall, catFloor, 0, 0,
	0, 0, 0, 0, 0, catFloor, catFloor, catFloor,
	catFloor, catFloor, catFloor, catFloor, catFloor, catFloor, catFloor, catFloor,
	catFloor, catFloor, catVWall, catVArch, catHWall, catHArch, catFloor, catFloor,
	catFloor, catHArch, catHWall, catVWall, catHWall, catHWall, catDWall, catHWallVArch,
	catDWall, catHArchVWall, catFloor, catFloor, catDWall, catDWall, catVWall, catVWall,
	catDWall, catHWall, catHWall, catFloor, catFloor, catFloor, catFloor, catVDoor,
	catHDoor, catHDoorVDoor, catVDoorEnd, catHDoorEnd, catDWall2, catDWall3, catHArchVWall3, catDWall2,
	catHWallVArch2, catDWall3, catVDoor, catDWall2, catDWall3, catHDoorVDoor, catHDoorVDoor, catVWall,
	catHWall, catVDoor, catHDoor, catDirt, catDirt, catVDoor, catHDoor, 0,
	0, 0, 0, 0, 0, 0, 0,
}

// tileDecorations is TileDecorations: the undecorated tile every cathedral
// tile id collapses to (levels/drlg_l1.cpp:255-292).
var tileDecorations = [207]Tile{
	0, catVWall, catHWall, catCorner, catDWall, catDArch, catVWallEnd, catHWallEnd,
	catHArchEnd, catVArchEnd, catHArchVWall, catVArch, catHArch, catFloor, catHWallVArch, catPillar,
	catVCorner, catHCorner, 0, 0, 0, 0, 0, 0,
	0, catVDoor, catHDoor, 0, catHDoorVDoor, 0, catVDoorEnd, catHDoorEnd,
	0, 0, 0, 0, 0, 0, 0, 0,
	catHArchVWall3, catDWall2, catHWallVArch2, catDWall3, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, catVWall2,
	catHWall2, 0, catDWall4, 0, 0, 0, 0, 0,
	0, catVWall2, 0, catHWall2, 0, 0, catVWall2, catHWall2,
	0, catHWall, catHWall, catHWall, catVWall, catVWall, catVArch, catVDoor,
	catFloor, catFloor, catFloor, catVWall, catHWall, catVWall, catHWall, catVWall,
	catHWall, catVWall, catHWall, catHWall, catHWall, catHWall, catHArch, 0,
	0, catVArch, catVWall, catVArch, catVWall, catFloor, 0, 0,
	0, 0, 0, 0, 0, catFloor, catFloor, catFloor,
	catFloor, catFloor, catFloor, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0,
}

// shadowPattern is one ShadowStruct entry: the tile the pattern keys on, the
// three base-type neighbours it requires (0 = any) and the three shadow tiles
// it writes (0 = leave alone) (levels/drlg_l1.cpp:151-181).
type shadowPattern struct {
	strig Tile
	s1    Tile
	s2    Tile
	s3    Tile
	nv1   Tile
	nv2   Tile
	nv3   Tile
}
