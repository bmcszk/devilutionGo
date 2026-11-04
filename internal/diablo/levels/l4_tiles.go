package levels

// This file holds the flat view of the tile layer the hell passes read and
// write through. Upstream declares `dungeon` and `pdungeon` as adjacent globals
// (levels/dun_tile_data.cpp:31-32), so a neighbour index that walks past the
// last row of `dungeon` lands in `pdungeon`: the hell passes rely on that
// (levels/drlg_l4.cpp:433-844, 906-920), and reproducing it means modelling the
// pair as one array of 2*DMAXX*DMAXY tiles.

import "slices"

// hellTile reads the tile at (x, y) of the layer pair. Cells left of the first
// row or past the doubled pair read as zero, which is what the upstream globals
// hold before the generator first writes them.
func (d *Dungeon) hellTile(x, y int) Tile { return d.flatTile(x, y) }

// setHellTile writes the tile at (x, y) of the layer pair, wrapping into
// pdungeon when the index walks past the last row of the layer itself.
func (d *Dungeon) setHellTile(x, y int, tile Tile) {
	index := x*DMAXY + y

	switch {
	case index < 0:
		return
	case index < DMAXX*DMAXY:
		d.Tiles[index/DMAXY][index%DMAXY] = tile
	case index < 2*DMAXX*DMAXY:
		prev := index - DMAXX*DMAXY
		d.PrevTiles[prev/DMAXY][prev%DMAXY] = tile
	default:
	}
}

// hellProtected reports whether the Protected bit is set for a cell that may sit
// outside the layer: the hell passes ask about the cell past the last column,
// where upstream reads the bit next to the bitset. Such a cell is treated as
// unprotected, which is the value the upstream layout leaves there.
func (d *Dungeon) hellProtected(x, y int) bool {
	if x < 0 || x >= DMAXX || y < 0 || y >= DMAXY {
		return false
	}

	return d.Protected.Test(x, y)
}

// isAnyOfTile reports whether tile is one of candidates (utils/is_of.hpp IsAnyOf).
func isAnyOfTile(tile Tile, candidates ...Tile) bool {
	return slices.Contains(candidates, tile)
}
