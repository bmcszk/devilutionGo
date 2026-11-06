package levels

// The nest level range and the pool count GenerateLevel demands
// (levels/gendung.cpp:335, levels/drlg_l3.cpp:1970, 1437).
const (
	// hiveFirstLevel and hiveLowestLevel are the first and last DTYPE_NEST
	// level: level 17 places the town warp instead of stairs up, level 20 has
	// no stairs down (levels/drlg_l3.cpp:1964, 1970).
	hiveFirstLevel  = 17
	hiveLowestLevel = 20
	// hiveSlimePoolMinimum is the pattern count PlaceSlimePool requires before
	// it reports a usable level (levels/drlg_l3.cpp:1437).
	hiveSlimePoolMinimum = 3
)

// hiveLightFirstTile, hiveLightLastTile and hiveLightRadius are the PlaceHiveLights
// band and radius (levels/drlg_l3.cpp:2172-2173). The ids are dPiece values, not
// mega-tile ids, so they exceed the uint8 Tile range.
const (
	hiveLightFirstTile uint16 = 381
	hiveLightLastTile  uint16 = 456
	hiveLightRadius    uint8  = 9
)

// hiveSlimePools are the pool patterns PlaceSlimePool stamps, in upstream order
// (levels/drlg_l3.cpp:1423-1435).
var hiveSlimePools = [...]hivePattern{
	{miniset: hivePattern41, chance: 30},
	{miniset: hivePattern42, chance: 40},
	{miniset: hivePattern39, chance: 50},
	{miniset: hivePattern40, chance: 60},
}

// hiveIsles are the isle minisets the nest decoration scatters, with the
// placement percentage of each call (levels/drlg_l3.cpp:2035-2041).
var hiveIsles = [...]hivePattern{
	{miniset: hiveIsle1, chance: 70},
	{miniset: hiveIsle2, chance: 70},
	{miniset: hiveIsle3, chance: 30},
	{miniset: hiveIsle4, chance: 30},
	{miniset: hiveIsle1, chance: 100},
	{miniset: hiveIsle2, chance: 100},
	{miniset: hiveIsle5, chance: 90},
}

// hiveFloorSubs are the four floor substitutions that follow the isles
// (levels/drlg_l3.cpp:2042-2045).
var hiveFloorSubs = [...]hiveSub{
	{search: 8, replace: 25, chance: 20},
	{search: 8, replace: 26, chance: 20},
	{search: 8, replace: 27, chance: 20},
	{search: 8, replace: 28, chance: 20},
}

// hiveTunnelRise is the first tunnel sweep, the odds climbing from pattern 29
// to 38 (levels/drlg_l3.cpp:2046-2055).
var hiveTunnelRise = [...]hivePattern{
	{miniset: hivePattern29, chance: 10},
	{miniset: hivePattern30, chance: 15},
	{miniset: hivePattern31, chance: 20},
	{miniset: hivePattern32, chance: 25},
	{miniset: hivePattern33, chance: 30},
	{miniset: hivePattern34, chance: 35},
	{miniset: hivePattern35, chance: 40},
	{miniset: hivePattern36, chance: 45},
	{miniset: hivePattern37, chance: 50},
	{miniset: hivePattern38, chance: 55},
}

// hiveTunnelFall is the second tunnel sweep: the same patterns back down
// from 38 to 29 (levels/drlg_l3.cpp:2056-2065).
var hiveTunnelFall = [...]hivePattern{
	{miniset: hivePattern38, chance: 10},
	{miniset: hivePattern37, chance: 15},
	{miniset: hivePattern36, chance: 20},
	{miniset: hivePattern35, chance: 25},
	{miniset: hivePattern34, chance: 30},
	{miniset: hivePattern33, chance: 35},
	{miniset: hivePattern32, chance: 40},
	{miniset: hivePattern31, chance: 45},
	{miniset: hivePattern30, chance: 50},
	{miniset: hivePattern29, chance: 55},
}

// hiveCentres are the two hive centre pieces that close the pattern block
// (levels/drlg_l3.cpp:2066-2067).
var hiveCentres = [...]hivePattern{
	{miniset: hivePattern9, chance: 40},
	{miniset: hivePattern10, chance: 45},
}

// hiveSurfaceSubs are the wall and floor rim substitutions that close the nest
// decoration pass (levels/drlg_l3.cpp:2068-2089).
var hiveSurfaceSubs = [...]hiveSub{
	{search: 7, replace: 29, chance: 25},
	{search: 7, replace: 30, chance: 25},
	{search: 7, replace: 31, chance: 25},
	{search: 7, replace: 32, chance: 25},
	{search: 9, replace: 33, chance: 25},
	{search: 9, replace: 34, chance: 25},
	{search: 9, replace: 35, chance: 25},
	{search: 9, replace: 36, chance: 25},
	{search: 9, replace: 37, chance: 25},
	{search: 10, replace: 39, chance: 25},
	{search: 10, replace: 40, chance: 25},
	{search: 10, replace: 41, chance: 25},
	{search: 10, replace: 42, chance: 25},
	{search: 10, replace: 43, chance: 25},
	{search: 9, replace: 45, chance: 25},
	{search: 9, replace: 46, chance: 25},
	{search: 10, replace: 47, chance: 25},
	{search: 10, replace: 48, chance: 25},
	{search: 11, replace: 38, chance: 25},
	{search: 11, replace: 44, chance: 25},
	{search: 11, replace: 49, chance: 25},
	{search: 11, replace: 50, chance: 25},
}

// hivePattern is one miniset the nest decoration stamps and the percentage of
// its call site; hiveSub is one 1x1 substitution of the same pass
// (levels/drlg_l3.cpp:1423-1435, 2035-2089).
type (
	hivePattern struct {
		miniset Miniset
		chance  int32
	}
	hiveSub struct {
		search  Tile
		replace Tile
		chance  int32
	}
)
