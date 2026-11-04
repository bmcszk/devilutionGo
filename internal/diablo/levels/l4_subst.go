package levels

// This file ports the two closing decoration passes of the hell generator: the
// shadow tiles the wall bases get and the random substitution that varies the
// wall and floor tiles (levels/drlg_l4.cpp:152-166, 838-866).

// applyHellShadowPatterns is ApplyShadowsPatterns: the two tiles south-west of a
// wall base are drawn as shadows (levels/drlg_l4.cpp:152-166).
//
// The inner loop is bounded by DMAXY, not DMAXX (upstream quirk, harmless while
// the layer is square).
func (d *Dungeon) applyHellShadowPatterns() {
	for y := 1; y < DMAXY; y++ {
		for x := 1; x < DMAXY; x++ {
			d.shadeHellTile(x, y)
		}
	}
}

// shadeHellTile draws the two shadow tiles south-west of a wall base
// (levels/drlg_l4.cpp:157-165).
func (d *Dungeon) shadeHellTile(x, y int) {
	if !isAnyOfTile(d.Tiles[x][y], 3, 4, 8, 15) {
		return
	}

	if d.Tiles[x-1][y] == hellFloor {
		d.Tiles[x-1][y] = 47
	}

	if d.Tiles[x-1][y-1] == hellFloor {
		d.Tiles[x-1][y-1] = 48
	}
}

// hellSubstitution is Substitution: one pass replaces walls and floors with a
// random tile of the same base type, the next turns some wall bases into skulls
// (levels/drlg_l4.cpp:838-866).
func (d *Dungeon) hellSubstitution() {
	for y := range DMAXY {
		for x := range DMAXX {
			d.substituteHellTile(x, y)
		}
	}

	for y := range DMAXY {
		for x := range DMAXX {
			d.skullHellTile(x, y)
		}
	}
}

// substituteHellTile is Substitution's first draw: a tile with a base type is
// replaced by the (GenerateRnd(16)+1)-th table entry sharing that base
// (levels/drlg_l4.cpp:841-859).
func (d *Dungeon) substituteHellTile(x, y int) {
	if !d.Rnd.FlipCoin(3) {
		return
	}

	base := hellBaseType(d.hellTile(x, y))
	if base == 0 || d.hellProtected(x, y) {
		return
	}

	d.setHellTile(x, y, d.drawHellBaseVariant(base))
}

// drawHellBaseVariant picks the (GenerateRnd(16)+1)-th table entry that shares
// base, wrapping at the end of the table (levels/drlg_l4.cpp:845-856).
func (d *Dungeon) drawHellBaseVariant(base Tile) Tile {
	remaining := d.Rnd.GenerateRnd(16)
	index := -1

	for remaining >= 0 {
		index++
		if index == len(l4BTypes) {
			index = 0
		}

		if base == l4BTypes[index] {
			remaining--
		}
	}

	return Tile(index)
}

// skullHellTile is Substitution's second draw: some wall bases become one of the
// three skull tiles (levels/drlg_l4.cpp:860-866). The pick is PickRandomlyAmong
// (engine/random.hpp:377-381), which clamps the raw roll to 0, so it draws
// through RandomIntLessThan and not GenerateRnd.
func (d *Dungeon) skullHellTile(x, y int) {
	if !d.Rnd.FlipCoin(10) {
		return
	}

	if hellBaseType(d.hellTile(x, y)) != hellFloor || d.hellProtected(x, y) {
		return
	}

	d.setHellTile(x, y, hellSkulls[d.Rnd.RandomIntLessThan(int32(len(hellSkulls)))])
}

// hellBaseType is L4BTYPES[tile]: the undecorated tile a variant maps to. Tiles
// past the table have no base; upstream reads the neighbouring static there,
// which no reachable tile id does (the generator's highest write is 136).
func hellBaseType(tile Tile) Tile {
	if int(tile) >= len(l4BTypes) {
		return 0
	}

	return l4BTypes[tile]
}
