package levels

// This file ports the hell wall passes: AddWall finds corridors that end in a
// wall and turns them into proper horizontal or vertical walls
// (levels/drlg_l4.cpp:334-486).

// l4HorizontalWallStarts are the tiles AddWall probes with HorizontalWallOk, in
// upstream order (levels/drlg_l4.cpp:442).
var l4HorizontalWallStarts = []Tile{10, 12, 13, 15, 16, 21, 22}

// l4VerticalWallStarts are the tiles AddWall probes with VerticalWallOk
// (levels/drlg_l4.cpp:452).
var l4VerticalWallStarts = []Tile{8, 9, 11, 14, 15, 16, 21, 23}

// addHellWalls is AddWall: every unprotected cell whose tile starts a wall gets
// the wall built through it, horizontal candidates before vertical ones
// (levels/drlg_l4.cpp:437-458).
func (d *Dungeon) addHellWalls() {
	for j := range DMAXY {
		for i := range DMAXX {
			if d.Protected.Test(i, j) {
				continue
			}

			d.tryHellHorizontalWall(i, j)
			d.tryHellVerticalWall(i, j)
		}
	}
}

// tryHellHorizontalWall runs the horizontal candidates at one cell. Upstream
// re-reads the cell for every candidate, because a wall built by the first match
// changes the tile the next candidate is compared against
// (levels/drlg_l4.cpp:442-451).
func (d *Dungeon) tryHellHorizontalWall(i, j int) {
	for _, start := range l4HorizontalWallStarts {
		if d.hellTile(i, j) != start {
			continue
		}

		d.Rnd.Discard(1)

		if x := d.hellHorizontalWallOk(i, j); x != -1 {
			d.hellHorizontalWall(i, j, x)
		}
	}
}

// tryHellVerticalWall is the vertical candidate list, run after the horizontal
// one (levels/drlg_l4.cpp:452-461).
func (d *Dungeon) tryHellVerticalWall(i, j int) {
	for _, start := range l4VerticalWallStarts {
		if d.hellTile(i, j) != start {
			continue
		}

		d.Rnd.Discard(1)

		if y := d.hellVerticalWallOk(i, j); y != -1 {
			d.hellVerticalWall(i, j, y)
		}
	}
}

// hellHorizontalWallOk measures the corridor to the right of (i, j): it runs
// while the cell is open floor with floor above and below, and the wall is only
// built for a run longer than three that ends in a wall tile
// (levels/drlg_l4.cpp:334-350).
func (d *Dungeon) hellHorizontalWallOk(i, j int) int {
	x := d.hellHorizontalRun(i, j)

	if x > 3 && isAnyOfTile(d.hellTile(i+x, j), l4HorizontalWallStarts...) {
		return x
	}

	return -1
}

// hellHorizontalRun measures the open corridor to the right of (i, j): the run
// stops at a protected cell or at a cell whose row above or below is not floor
// (levels/drlg_l4.cpp:336-347).
func (d *Dungeon) hellHorizontalRun(i, j int) int {
	x := 1
	for ; d.hellTile(i+x, j) == hellFloor; x++ {
		if d.hellProtected(i+x, j) || d.hellTile(i+x, j-1) != hellFloor || d.hellTile(i+x, j+1) != hellFloor {
			break
		}
	}

	return x
}

// hellVerticalWallOk is the same probe downwards (levels/drlg_l4.cpp:352-371).
func (d *Dungeon) hellVerticalWallOk(i, j int) int {
	y := d.hellVerticalRun(i, j)

	if y > 3 && isAnyOfTile(d.hellTile(i, j+y), l4VerticalWallStarts...) {
		return y
	}

	return -1
}

// hellVerticalRun measures the open corridor below (i, j)
// (levels/drlg_l4.cpp:354-368).
func (d *Dungeon) hellVerticalRun(i, j int) int {
	y := 1
	for ; d.hellTile(i, j+y) == hellFloor; y++ {
		if d.hellProtected(i, j+y) || d.hellTile(i-1, j+y) != hellFloor || d.hellTile(i+1, j+y) != hellFloor {
			break
		}
	}

	return y
}

// hellHorizontalWall turns the open corridor at (i, j) into a horizontal wall of
// dx cells, caps the far cell and opens a random doorway
// (levels/drlg_l4.cpp:373-413).
func (d *Dungeon) hellHorizontalWall(i, j, dx int) {
	d.setHellTile(i, j, hellWallHead(d.hellTile(i, j)))

	for x := 1; x < dx; x++ {
		d.setHellTile(i+x, j, hellHWall)
	}

	d.setHellTile(i+dx, j, hellWallEnd(d.hellTile(i+dx, j)))

	xx := int(d.Rnd.GenerateRnd(int32(dx-3))) + 1
	d.setHellTile(i+xx, j, 57)
	d.setHellTile(i+xx+2, j, 56)
	d.setHellTile(i+xx+1, j, 60)

	if d.hellTile(i+xx, j-1) == hellFloor {
		d.setHellTile(i+xx, j-1, 58)
	}

	if d.hellTile(i+xx+1, j-1) == hellFloor {
		d.setHellTile(i+xx+1, j-1, 59)
	}
}

// hellVerticalWall is the vertical form: the corridor is filled with floor and
// closed with a doorway (levels/drlg_l4.cpp:415-470).
func (d *Dungeon) hellVerticalWall(i, j, dy int) {
	d.setHellTile(i, j, hellWallTop(d.hellTile(i, j)))

	for y := 1; y < dy; y++ {
		d.setHellTile(i, j+y, hellVWall)
	}

	d.setHellTile(i, j+dy, hellWallBottom(d.hellTile(i, j+dy)))

	yy := int(d.Rnd.GenerateRnd(int32(dy-3))) + 1
	d.setHellTile(i, j+yy, 53)
	d.setHellTile(i, j+yy+2, 52)
	d.setHellTile(i, j+yy+1, hellFloor)

	if d.hellTile(i-1, j+yy) == hellFloor {
		d.setHellTile(i-1, j+yy, 54)
	}

	if d.hellTile(i-1, j+yy-1) == hellFloor {
		d.setHellTile(i-1, j+yy-1, 55)
	}
}

// hellWallHead reshapes the tile the wall starts on: a corridor corner becomes
// the matching wall piece (levels/drlg_l4.cpp:374-382). The three rules cannot
// chain, because no target is one of the sources.
func hellWallHead(tile Tile) Tile {
	switch tile {
	case 13:
		return 17
	case 16:
		return 11
	case 12:
		return 14
	default:
		return tile
	}
}

// hellWallTop is the vertical form of hellWallHead (levels/drlg_l4.cpp:416-424).
func hellWallTop(tile Tile) Tile {
	switch tile {
	case 14:
		return 17
	case 8:
		return 9
	case 15:
		return 10
	default:
		return tile
	}
}

// hellWallEnd shapes the tile the horizontal wall runs into
// (levels/drlg_l4.cpp:393-402).
func hellWallEnd(tile Tile) Tile {
	switch tile {
	case 15:
		return 14
	case 10:
		return 17
	case 21:
		return 23
	case 22:
		return 29
	default:
		return tile
	}
}

// hellWallBottom shapes the tile the vertical wall runs into
// (levels/drlg_l4.cpp:434-446).
func hellWallBottom(tile Tile) Tile {
	switch tile {
	case 11:
		return 17
	case 9:
		return 10
	case 16:
		return 13
	case 21:
		return 22
	case 23:
		return 29
	default:
		return tile
	}
}
