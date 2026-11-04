package levels

// This file ports FixTilesPatterns (levels/drlg_l1.cpp:875-995): three full
// scans of the mega-tile layer that close the tile vocabulary up around the
// walls the carve passes left. Upstream writes the rules as sequential ifs; the
// port keeps the exact rule order as data so the order-dependent cases (a rule
// reading what an earlier rule just wrote) behave identically.

// fixPatternsPass1 is the first FixTilesPatterns scan
// (levels/drlg_l1.cpp:880-906). It tags the dirt next to each wall face.
var fixPatternsPass1 = []tilePattern{
	// i + 1 < DMAXX (drlg_l1.cpp:882-893)
	{catHWall, Point{X: 1}, catDirt, catDirtHwallEnd},
	{catFloor, Point{X: 1}, catDirt, catDirtHwall},
	{catFloor, Point{X: 1}, catHWall, catHWallEnd},
	{catVWallEnd, Point{X: 1}, catDirt, catDirtVwallEnd},
	// j + 1 < DMAXY (drlg_l1.cpp:895-904)
	{catVWall, Point{Y: 1}, catDirt, catDirtVwallEnd},
	{catFloor, Point{Y: 1}, catVWall, catVWallEnd},
	{catFloor, Point{Y: 1}, catDirt, catDirtVwall},
}

// fixPatternsPass2 is the second FixTilesPatterns scan
// (levels/drlg_l1.cpp:908-975). It turns the tagged dirt into corners and wall
// ends.
//
//nolint:gochecknoglobals // the rule table is the ported constant data of one upstream function
var fixPatternsPass2 = []tilePattern{
	// i + 1 < DMAXX (drlg_l1.cpp:911-943)
	{catFloor, Point{X: 1}, catDirtVwall, catHDirtCorner},
	{catFloor, Point{X: 1}, catDirt, catVDirtCorner},
	{catHWallEnd, Point{X: 1}, catDirt, catDirtHwallEnd},
	{catFloor, Point{X: 1}, catDirtVwallEnd, catHDirtCorner},
	{catDirtVwall, Point{X: 1}, catDirt, catVDirtCorner},
	{catHWall, Point{X: 1}, catDirtVwall, catHDirtCorner},
	{catDirtVwall, Point{X: 1}, catVWall, catVWallEnd},
	{catHWallEnd, Point{X: 1}, catDirtVwall, catHDirtCorner},
	{catHWall, Point{X: 1}, catVWall, catVWallEnd},
	{catCorner, Point{X: 1}, catDirt, catDirtVwallEnd},
	{catHDirtCorner, Point{X: 1}, catVWall, catVWallEnd},
	{catHWallEnd, Point{X: 1}, catVWall, catVWallEnd},
	{catHWallEnd, Point{X: 1}, catDirtVwallEnd, catHDirtCorner},
	{catDWall, Point{X: 1}, catVCorner, catHCorner},
	{catHWallEnd, Point{X: 1}, catFloor, catHCorner},
	{catHWall, Point{X: 1}, catDirtVwallEnd, catHDirtCorner},
	{catHWall, Point{X: 1}, catFloor, catHCorner},
	// i > 0 (drlg_l1.cpp:944-956)
	{catDirtHwallEnd, Point{X: -1}, catDirt, catDirtVwall},
	{catDirtVwall, Point{X: -1}, catDirtHwallEnd, catHDirtCorner},
	{catVWallEnd, Point{X: -1}, catDirt, catDirtVwallEnd},
	{catVWallEnd, Point{X: -1}, catDirtHwallEnd, catHDirtCorner},
	// j + 1 < DMAXY (drlg_l1.cpp:957-971)
	{catVWall, Point{Y: 1}, catHWall, catHWallEnd},
	{catVWallEnd, Point{Y: 1}, catDirtHwall, catHDirtCorner},
	{catDirtHwall, Point{Y: 1}, catHWall, catHWallEnd},
	{catVWallEnd, Point{Y: 1}, catHWall, catHWallEnd},
	{catHDirtCorner, Point{Y: 1}, catHWall, catHWallEnd},
	{catVWallEnd, Point{Y: 1}, catDirt, catDirtVwallEnd},
	{catVWallEnd, Point{Y: 1}, catFloor, catVCorner},
	{catVWall, Point{Y: 1}, catFloor, catVCorner},
	{catFloor, Point{Y: 1}, catVCorner, catHCorner},
	// j > 0 (drlg_l1.cpp:972-981)
	//
	// The first two rules share their condition; upstream's second only fires
	// when the first did not, because it reads the tile the first rewrote.
	{catVWallEnd, Point{Y: -1}, catDirt, catHWallEnd},
	{catVWallEnd, Point{Y: -1}, catDirt, catDirtVwallEnd},
	{catHWallEnd, Point{Y: -1}, catDirtVwallEnd, catHDirtCorner},
	{catDirtHwall, Point{Y: -1}, catDirtVwallEnd, catHDirtCorner},
}

// fixPatternsPass3 is the third FixTilesPatterns scan
// (levels/drlg_l1.cpp:986-994).
var fixPatternsPass3 = []tilePattern{
	{catDWall, Point{Y: 1}, catHWall, catHWallEnd},
	{catHWall, Point{X: 1}, catDirtVwall, catHDirtCorner},
	{catDirtHwall, Point{Y: 1}, catDirt, catVDirtCorner},
}

// tilePattern is one FixTilesPatterns rule: when the anchor tile holds From,
// the tile at Offset is rewritten from Was to To. Rules are applied in table
// order and each one re-reads the layer, which is what upstream's sequential
// ifs do.
type tilePattern struct {
	From   Tile
	Offset Point
	Was    Tile
	To     Tile
}

// isInsideLayer reports whether p is a mega tile of the layer
// (levels/drlg_l1.cpp:878-879).
func isInsideLayer(p Point) bool {
	return p.X >= 0 && p.X < DMAXX && p.Y >= 0 && p.Y < DMAXY
}

// fixTilesPatterns runs the three FixTilesPatterns scans in order
// (levels/drlg_l1.cpp:875-995).
func (d *Dungeon) fixTilesPatterns() {
	d.applyTilePatterns(fixPatternsPass1)
	d.applyTilePatterns(fixPatternsPass2)
	d.applyTilePatterns(fixPatternsPass3)
}

// applyTilePatterns scans the layer in upstream's j-then-i order, applying
// every rule of one pass at each anchor tile (levels/drlg_l1.cpp:880-994).
func (d *Dungeon) applyTilePatterns(patterns []tilePattern) {
	for j := range DMAXY {
		for i := range DMAXX {
			d.applyTilePatternsAt(Point{X: i, Y: j}, patterns)
		}
	}
}

// applyTilePatternsAt applies one pass's rules at one anchor tile, skipping
// rules whose target lies outside the layer, which is upstream's BUGFIX bounds
// check (levels/drlg_l1.cpp:878-879).
func (d *Dungeon) applyTilePatternsAt(anchor Point, patterns []tilePattern) {
	for _, pattern := range patterns {
		if d.Tiles[anchor.X][anchor.Y] != pattern.From {
			continue
		}

		target := anchor.Add(pattern.Offset)
		if !isInsideLayer(target) {
			continue
		}

		if d.Tiles[target.X][target.Y] == pattern.Was {
			d.Tiles[target.X][target.Y] = pattern.To
		}
	}
}
