//nolint:testpackage // the external tests reach the ported generator tables and passes through this seam
package levels

// Test seams for the external tests. Nothing here is part of the package API:
// golangci's testpackage linter skips export_test.go, which keeps the generator
// surface free of test-only accessors while still letting the external tests pin
// the ported data tables and drive one pass at a time
// (levels/drlg_l1.cpp:183-293).

// Cathedral tile ids the pass tests drive (tile_cathedral.go, drlg_l1.cpp:97-178).
const (
	TestFloor         = catFloor
	TestFloor22       = catFloor22
	TestFloor23       = catFloor23
	TestVWall         = catVWall
	TestHWall         = catHWall
	TestVCorner       = catVCorner
	TestHCorner       = catHCorner
	TestVDirtCorner   = catVDirtCorner
	TestHDirtCorner   = catHDirtCorner
	TestDirtCorner2   = catDirtCorner2
	TestDirtHWall     = catDirtHwall
	TestDirtVWall     = catDirtVwall
	TestDirtHWall2    = catDirtHWall2
	TestDirtVWall2    = catDirtVWall2
	TestDirtHWallEnd  = catDirtHwallEnd
	TestDirtVWallEnd  = catDirtVwallEnd
	TestDirtHWallEnd2 = catDirtHWallEnd2
	TestDirtVWallEnd2 = catDirtVWallEnd2
	TestHArchEnd      = catHArchEnd
	TestVArchEnd      = catVArchEnd
	TestVWall2        = catVWall2
)

// FixDirtTilesForTest runs the FixDirtTiles pass (levels/drlg_l1.cpp:1111-1136).
func (d *Dungeon) FixDirtTilesForTest() { d.fixDirtTiles() }

// FixCornerTilesForTest runs the FixCornerTiles pass (levels/drlg_l1.cpp:1137-1154).
func (d *Dungeon) FixCornerTilesForTest() { d.fixCornerTiles() }

// FixTransparencyForTest runs the FixTransparency pass
// (levels/drlg_l1.cpp:1076-1110).
func (d *Dungeon) FixTransparencyForTest() { d.fixTransparency() }

// FillFloorForTest runs the FillFloor pass (levels/drlg_l1.cpp:414-429).
func (d *Dungeon) FillFloorForTest() { d.fillFloor() }

// SubstitutionForTest runs the Substitution pass (levels/drlg_l1.cpp:996-1035).
func (d *Dungeon) SubstitutionForTest() { d.substitution() }

// CheckQuestsForTest runs the DRLG_CheckQuests call site
// (levels/drlg_quests.cpp:89-118).
func (d *Dungeon) CheckQuestsForTest(quests Quests) { d.checkQuests(quests) }

// TileTableSizes reports the lengths of the three hard-coded generator tables:
// BaseTypes, TileDecorations and ShadowPatterns.
func TileTableSizes() (base, decorations, shadows int) {
	return len(baseTypesLen()), len(tileDecorationsLen()), len(shadowPatterns)
}

// DecorationFamily reports the decoration family a tile belongs to
// (TileDecorations) so a test can assert Substitution stays inside it.
func DecorationFamily(tile Tile) Tile { return tileDecorations[tile] }

// baseTypesLen is the BaseTypes table (levels/drlg_l1.cpp:227-254).
func baseTypesLen() []Tile {
	return baseTypes[:]
}

// tileDecorationsLen is the TileDecorations table (levels/drlg_l1.cpp:255-292).
func tileDecorationsLen() []Tile {
	return tileDecorations[:]
}
