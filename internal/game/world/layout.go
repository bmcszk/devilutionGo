package world

// LevelLayout is one dungeon level's world-tile image: the piece grid the
// family generator expanded (levels/gendung.cpp:591-624 DRLG_LPass3), the SOL
// record that classifies each piece (LoadLevelSOLData, dun_tile_data.cpp:107)
// and the items the load leaves on the floor (InitItems, diablo.cpp:3155).
//
// It is a plain value so the diablo layer builds it from the committed corpus
// and this layer consumes it without knowing where a level comes from.
type LevelLayout struct {
	// Piece is dPiece in its own [x][y] indexing (dun_tile_data.hpp:99),
	// flattened x-major: index x*MAXDUNY+y.
	Piece []int
	// SOL is one record byte per piece id — the TileProperties flags a piece
	// carries, still in the SOL record's own bit layout
	// (levels/dun_tile.hpp:106-114).
	SOL []byte
	// Items are the floor items the level starts with, already at their
	// resolved Items indices.
	Items []FloorItem
}

// FloorItem ties one item index to the tile it lies on.
type FloorItem struct {
	// Index is the 0-based Items index; dItem stores it 1-based
	// (items.cpp:100).
	Index int
	// Tile is where the item lies.
	Tile Tile
}

// properties converts the SOL record into the grid's property word, so the
// tile predicates read the flags they test.
func (l LevelLayout) properties() []TileProperties {
	props := make([]TileProperties, len(l.SOL))
	for i, sol := range l.SOL {
		props[i] = SOLTileProperties(sol)
	}

	return props
}
