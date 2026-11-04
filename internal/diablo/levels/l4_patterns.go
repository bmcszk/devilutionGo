package levels

// This file ports the hell tile-pattern and transparency passes: the six
// FixTilesPatterns sweeps (through l4TilePatterns), the transparency window the
// half-covered wall tiles need, the corner-tile post pass and the final fix.

// fixHellTilesPatterns is FixTilesPatterns: six full sweeps over the layer, in
// upstream order (levels/drlg_l4.cpp:489-836).
func (d *Dungeon) fixHellTilesPatterns() {
	for _, sweep := range l4TilePatterns {
		for j := range DMAXY {
			for i := range DMAXX {
				d.applyHellTileRules(sweep, i, j)
			}
		}
	}
}

// applyHellTileRules runs one sweep's rules at (i, j), in table order, so a rule
// sees the writes of the rules before it (levels/drlg_l4.cpp:489-836).
func (d *Dungeon) applyHellTileRules(sweep []l4TileRule, i, j int) {
	for _, rule := range sweep {
		if d.hellTile(i, j) != rule.match || !d.hellRuleHolds(rule, i, j) {
			continue
		}

		d.setHellTile(i+rule.dx, j+rule.dy, rule.write)
	}
}

// hellRuleHolds reports whether every neighbour test of a rule passes.
func (d *Dungeon) hellRuleHolds(rule l4TileRule, i, j int) bool {
	for _, hit := range rule.cond {
		if (d.hellTile(i+hit.dx, j+hit.dy) == hit.tile) == hit.not {
			return false
		}
	}

	return true
}

// fixHellTransparency is FixTransparency: the half-covered wall tiles copy the
// region id of the tile they are drawn against onto the doubled tiles next to
// them (levels/drlg_l4.cpp:1021-1062).
//
// Upstream reads the tile above and to the right of the cell before checking
// the index bounds (marked BUGFIX there); the reads are modelled through the
// layer pair, so the j == 0 read lands on the last tile of the previous row
// exactly like the upstream array walk.
func (d *Dungeon) fixHellTransparency() {
	yy := hellTransOrigin
	for j := range DMAXY {
		xx := hellTransOrigin
		for i := range DMAXX {
			d.fixHellTransparencyAt(i, j, xx, yy)

			xx += 2
		}

		yy += 2
	}
}

// fixHellTransparencyAt is the per-tile half of FixTransparency: the two guarded
// half-wall copies, then the tile's own spills (levels/drlg_l4.cpp:1027-1060).
func (d *Dungeon) fixHellTransparencyAt(i, j, xx, yy int) {
	tile := d.hellTile(i, j)

	d.copyHellHalfWallTrans(tile, i, j, xx, yy)
	d.copyHellTileTrans(tile, xx, yy)
}

// copyHellHalfWallTrans copies the region id onto the doubled tiles a half-cover
// wall leaves uncovered, when the tile it is drawn against is the double-wide
// wall it belongs to (levels/drlg_l4.cpp:1029-1039).
func (d *Dungeon) copyHellHalfWallTrans(tile Tile, i, j, xx, yy int) {
	if isDURightWall(tile) && d.hellTile(i, j-1) == hellDWall {
		d.Trans[xx+1][yy] = d.Trans[xx][yy]
		d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
	}

	if isDLLeftWall(tile) && d.hellTile(i+1, j) == hellDWallB {
		d.Trans[xx][yy+1] = d.Trans[xx][yy]
		d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
	}
}

// copyHellTileTrans is FixTransparency's per-tile switch: the wall and doorway
// tiles spill their region onto the doubled tiles they only half cover
// (levels/drlg_l4.cpp:1040-1058).
func (d *Dungeon) copyHellTileTrans(tile Tile, xx, yy int) {
	switch tile {
	case hellDWall:
		d.Trans[xx+1][yy] = d.Trans[xx][yy]
		d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
	case hellDWallB:
		d.Trans[xx][yy+1] = d.Trans[xx][yy]
		d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
	case 24:
		d.Trans[xx+1][yy] = d.Trans[xx][yy]
		d.Trans[xx][yy+1] = d.Trans[xx][yy]
		d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
	case hellDoorH:
		d.Trans[xx-1][yy] = d.Trans[xx][yy+1]
		d.Trans[xx][yy] = d.Trans[xx][yy+1]
	case hellDoorV:
		d.Trans[xx][yy-1] = d.Trans[xx+1][yy]
		d.Trans[xx][yy] = d.Trans[xx+1][yy]
	default:
	}
}

// isDURightWall reports the walls that cover the right half of their tile
// (levels/drlg_l4.cpp:991-1004).
func isDURightWall(tile Tile) bool { return isAnyOfTile(tile, 25, 28, 23) }

// isDLLeftWall reports the walls that cover the left half of their tile
// (levels/drlg_l4.cpp:1006-1019).
func isDLLeftWall(tile Tile) bool { return isAnyOfTile(tile, 27, 26, 22) }

// fixHellCornerTiles is FixCornerTiles: a corner wall tile whose right or lower
// neighbour is a rock-edge tile is redrawn with the +98 corner variant
// (levels/drlg_l4.cpp:1064-1074).
func (d *Dungeon) fixHellCornerTiles() {
	for j := 1; j < DMAXY-1; j++ {
		for i := 1; i < DMAXX-1; i++ {
			if d.hellNeedsCornerTile(i, j) {
				d.Tiles[i][j] += hellCornerOffset
			}
		}
	}
}

// hellNeedsCornerTile reports whether the cell holds a wall tile that is drawn as
// a corner because the tile right of it or below it is bare rock
// (levels/drlg_l4.cpp:1068-1070).
func (d *Dungeon) hellNeedsCornerTile(i, j int) bool {
	tile := d.Tiles[i][j]
	if tile < hellEdgeWall || tile > hellRock {
		return false
	}

	return d.Tiles[i+1][j] < hellEdgeWall || d.Tiles[i][j+1] < hellEdgeWall
}

// generalFix is GeneralFix: a doorway tile between floor and wall is redrawn as
// the plain wall corner (levels/drlg_l4.cpp:1090-1098).
func (d *Dungeon) generalFix() {
	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			if d.hellDoorNeedsCorner(i, j) {
				d.Tiles[i][j] = 17
			}
		}
	}
}

// hellDoorNeedsCorner reports whether the cell is a doorway tile standing on a
// corridor floor with open floor below it (levels/drlg_l4.cpp:1094).
func (d *Dungeon) hellDoorNeedsCorner(i, j int) bool {
	if !isAnyOfTile(d.Tiles[i][j], 24, 122) {
		return false
	}

	return d.Tiles[i+1][j] == hellHWall && d.Tiles[i][j+1] == 5
}
