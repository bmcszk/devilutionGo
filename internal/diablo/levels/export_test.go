//nolint:testpackage // the external tests reach the ported generator tables through this seam
package levels

// Test seams for the external tests. Nothing here is part of the package API:
// golangci's testpackage linter skips export_test.go, which keeps the generator
// surface free of test-only accessors while still letting the external tests
// pin the ported data tables (levels/drlg_l1.cpp:183-293).

// TileTableSizes reports the lengths of the three hard-coded generator tables:
// BaseTypes, TileDecorations and ShadowPatterns.
func TileTableSizes() (base, decorations, shadows int) {
	return len(baseTypesLen()), len(tileDecorationsLen()), len(shadowPatterns)
}

// baseTypesLen is the BaseTypes table (levels/drlg_l1.cpp:227-254).
func baseTypesLen() []Tile {
	return baseTypes[:]
}

// tileDecorationsLen is the TileDecorations table (levels/drlg_l1.cpp:255-292).
func tileDecorationsLen() []Tile {
	return tileDecorations[:]
}
