package levels

// This file declares the rule shape of the FixTilesPatterns table: one rule is
// one `if` of levels/drlg_l4.cpp:489-836.

// l4TileRule is one rule of FixTilesPatterns: the origin cell must hold match,
// every entry of cond must hold, then Tiles[i+dx][j+dy] = write. The offsets are
// relative to the scanned cell and read through the layer pair, exactly like
// the upstream array indexing (levels/dun_tile_data.cpp:31-32).
type l4TileRule struct {
	match Tile
	cond  []l4Hit
	dx    int
	dy    int
	write Tile
}

// l4Hit is one neighbour test of a rule at Tiles[i+dx][j+dy]: the tile there
// must equal tile, or must differ from it when not is set.
type l4Hit struct {
	dx   int
	dy   int
	tile Tile
	not  bool
}

// l4TilePatterns is FixTilesPatterns' six sequential sweeps of the layer
// (levels/drlg_l4.cpp:489-836). The sweeps cannot be flattened into one list:
// each rule reads the tiles the rules of the previous sweep already wrote.
