package levels

// This file ports the tile fix passes the cathedral generator runs after the
// stairs are placed: the transparency patch-ups, the dirt and corner tile
// fixes, the decoration substitution, the lamp scattering and the floor
// variation (levels/drlg_l1.cpp:414-429, 996-1154, 1282-1289).

// dirtFixes are FixDirtTiles' rules in upstream order, applied to (i,j) with
// (i+1,j) and (i,j+1) as the neighbour scans.
var dirtFixes = []dirtFix{
	{catHDirtCorner, Point{X: 1}, catDirtVwall, catDirtCorner2},
	{catDirtVwall, Point{X: 1}, catDirtVwall, catDirtVWall2},
	{catDirtVwallEnd, Point{X: 1}, catDirtVwall, catDirtVWallEnd2},
	{catDirtHwall, Point{Y: 1}, catDirtHwall, catDirtHWall2},
	{catHDirtCorner, Point{Y: 1}, catDirtHwall, catDirtCorner2},
	{catDirtHwallEnd, Point{Y: 1}, catDirtHwall, catDirtHWallEnd2},
}

// cornerFixes are FixCornerTiles' three rules in upstream order
// (levels/drlg_l1.cpp:1140-1152).
var cornerFixes = []cornerFix{
	{
		from: catHCorner, to: catVCorner, protected: true,
		first:  tileProbe{offset: Point{X: -1}, want: catFloor},
		second: tileProbe{offset: Point{Y: -1}, want: catVWall},
	},
	{
		from: catDirtCorner2, to: catHArchEnd,
		first:  tileProbe{offset: Point{X: 1}, want: catFloor},
		second: tileProbe{offset: Point{Y: 1}, want: catVWall},
	},
	{
		from: catDirtCorner2, to: catVArchEnd,
		first:  tileProbe{offset: Point{Y: 1}, want: catFloor},
		second: tileProbe{offset: Point{X: 1}, want: catHWall},
	},
}

// dirtFix rewrites one dirt tile whose neighbour does not continue its face
// (levels/drlg_l1.cpp:1111-1136).
type dirtFix struct {
	from      Tile
	neighbour Point
	was       Tile
	to        Tile
}

// tileProbe is one neighbouring tile a corner fix requires.
type tileProbe struct {
	offset Point
	want   Tile
}

// cornerFix rewrites one corner shape into the corner tile the art needs
// (levels/drlg_l1.cpp:1139-1153).
type cornerFix struct {
	from      Tile
	to        Tile
	first     tileProbe
	second    tileProbe
	protected bool
}

// fixTransparency copies the transparency region over the wall tiles that were
// widened into a second world tile, so the whole doubled tile belongs to the
// room behind it (levels/drlg_l1.cpp:1076-1110).
//
//nolint:cyclop,revive // the five tile cases are upstream's sequential if chain
func (d *Dungeon) fixTransparency() {
	yy := 16

	for j := range DMAXY {
		xx := 16

		for i := range DMAXX {
			// BUGFIX: the upstream check is `j > 0` first (drlg_l1.cpp:1083).
			if d.Tiles[i][j] == catDirtHwallEnd && j > 0 && d.Tiles[i][j-1] == catDirtHwall {
				d.Trans[xx+1][yy] = d.Trans[xx][yy]
				d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
			}

			// BUGFIX: the upstream check is `i + 1 < DMAXY` (drlg_l1.cpp:1088).
			if d.Tiles[i][j] == catDirtVwallEnd && i+1 < DMAXY && d.Tiles[i+1][j] == catDirtVwall {
				d.Trans[xx][yy+1] = d.Trans[xx][yy]
				d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
			}

			if d.Tiles[i][j] == catDirtHwall {
				d.Trans[xx+1][yy] = d.Trans[xx][yy]
				d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
			}

			if d.Tiles[i][j] == catDirtVwall {
				d.Trans[xx][yy+1] = d.Trans[xx][yy]
				d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
			}

			if d.Tiles[i][j] == catVDirtCorner {
				d.Trans[xx+1][yy] = d.Trans[xx][yy]
				d.Trans[xx][yy+1] = d.Trans[xx][yy]
				d.Trans[xx+1][yy+1] = d.Trans[xx][yy]
			}

			xx += 2
		}

		yy += 2
	}
}

// fixDirtTiles closes the dirt faces that end mid-wall
// (levels/drlg_l1.cpp:1111-1136).
func (d *Dungeon) fixDirtTiles() {
	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			for _, fix := range dirtFixes {
				d.applyDirtFix(i, j, fix)
			}
		}
	}
}

// applyDirtFix rewrites one tile when it holds the rule's dirt face and its
// neighbour no longer continues that face
// (levels/drlg_l1.cpp:1115-1133).
func (d *Dungeon) applyDirtFix(i, j int, fix dirtFix) {
	if d.Tiles[i][j] != fix.from {
		return
	}

	if d.Tiles[fix.neighbour.X+i][fix.neighbour.Y+j] != fix.was {
		d.Tiles[i][j] = fix.to
	}
}

// fixCornerTiles turns the corner shapes the wall pass left into the corner
// variants the tile art needs (levels/drlg_l1.cpp:1137-1154).
func (d *Dungeon) fixCornerTiles() {
	for j := 1; j < DMAXY-1; j++ {
		for i := 1; i < DMAXX-1; i++ {
			for _, fix := range cornerFixes {
				d.applyCornerFix(i, j, fix)
			}
		}
	}
}

// applyCornerFix rewrites one tile when its corner shape and both neighbours
// match (levels/drlg_l1.cpp:1140-1152).
func (d *Dungeon) applyCornerFix(i, j int, fix cornerFix) {
	if fix.protected && d.Protected.Test(i, j) {
		return
	}

	if d.Tiles[i][j] != fix.from {
		return
	}

	first := fix.first.offset
	if d.Tiles[i+first.X][j+first.Y] != fix.first.want {
		return
	}

	second := fix.second.offset
	if d.Tiles[i+second.X][j+second.Y] != fix.second.want {
		return
	}

	// BUGFIX: upstream notes the HCorner case should also protect the tile
	// (drlg_l1.cpp:1142).
	d.Tiles[i][j] = fix.to
}

// substitution swaps decorated tiles for a random member of their undecorated
// family, extending the family over the neighbouring tile when the chosen
// variant is a corner piece (levels/drlg_l1.cpp:996-1035).
func (d *Dungeon) substitution() {
	for y := range DMAXY {
		for x := range DMAXX {
			if d.Rnd.FlipCoin(4) {
				d.substituteTile(x, y)
			}
		}
	}
}

// substituteTile applies the substitution decision to one tile
// (levels/drlg_l1.cpp:1000-1031).
//
//nolint:cyclop,revive // the corner-family fixups are upstream's sequential branches
func (d *Dungeon) substituteTile(x, y int) {
	family := tileDecorations[d.Tiles[x][y]]
	if family == 0 || d.Protected.Test(x, y) {
		return
	}

	index := d.pickDecoration(family)

	// BUGFIX: upstream adds the `y > 0` guard (drlg_l1.cpp:1019-1026).
	if index == catVWall4 && y > 0 {
		if tileDecorations[d.Tiles[x][y-1]] != catVWall2 || d.Protected.Test(x, y-1) {
			index = catVWall2
		} else {
			d.Tiles[x][y-1] = catVWall5
		}
	}

	// BUGFIX: upstream adds the `x + 1 < DMAXX` guard (drlg_l1.cpp:1028-1035).
	if index == catHWall4 && x+1 < DMAXX {
		if tileDecorations[d.Tiles[x+1][y]] != catHWall2 || d.Protected.Test(x+1, y) {
			index = catHWall2
		} else {
			d.Tiles[x+1][y] = catHWall5
		}
	}

	d.Tiles[x][y] = index
}

// pickDecoration walks the decoration table until the drawn occurrence of the
// tile's family is reached, wrapping at the end of the table
// (levels/drlg_l1.cpp:1004-1013). The table is indexed by tile id, so the walk
// position is also the tile it returns.
func (d *Dungeon) pickDecoration(family Tile) Tile {
	remaining := d.Rnd.GenerateRnd(16)
	index := -1

	for remaining >= 0 {
		index++

		if index == len(tileDecorations) {
			index = 0
		}

		if family == tileDecorations[index] {
			remaining--
		}
	}

	return Tile(index)
}

// placeLamps scatters the candlestick miniset over the level
// (levels/drlg_l1.cpp:1282-1286).
func (d *Dungeon) placeLamps() {
	numt := int(d.Rnd.GenerateRnd(5)) + 5
	for range numt {
		d.PlaceMiniSet(lamps, DMAXX*DMAXY, true)
	}
}

// fillFloor varies the plain floor tiles with the two decorated floor variants
// (levels/drlg_l1.cpp:414-429).
func (d *Dungeon) fillFloor() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.fillFloorTile(i, j)
		}
	}
}

// fillFloorTile varies one plain, unprotected floor tile
// (levels/drlg_l1.cpp:419-427).
func (d *Dungeon) fillFloorTile(i, j int) {
	if d.Tiles[i][j] != catFloor || d.Protected.Test(i, j) {
		return
	}

	switch d.Rnd.RandomIntLessThan(3) {
	case 1:
		d.Tiles[i][j] = catFloor22
	case 2:
		d.Tiles[i][j] = catFloor23
	default:
	}
}

// copyToPrevTiles mirrors the finished layer into pdungeon, the copy the
// automap and the set pieces read (levels/drlg_l1.cpp:1292).
func (d *Dungeon) copyToPrevTiles() {
	d.PrevTiles = d.Tiles
}
