package levels

// This file ports the caves fence pass: the two decorator scans that wall off
// long floor runs and then re-tile the wall lines, plus the door fix-ups that
// close the gaps the two scans leave (levels/drlg_l3.cpp:1530-1800).

// l3FenceVerticalSub and l3FenceHorizontalSub are Fence's line substitutions,
// indexed by the tile that is found. Zero means "not a fence tile", the same
// sentinel the fill tables use (levels/drlg_l3.cpp:1728-1743, 1766-1778).
var (
	l3FenceVerticalSub   = [256]Tile{2: 139, 10: 131, 126: 133, 129: 133, 134: 138, 136: 138}
	l3FenceHorizontalSub = [256]Tile{4: 140, 9: 130, 121: 132, 124: 132, 135: 138, 137: 138}
)

// fenceDoorKeep are the wall tiles that let a fence door stay: CreateThemeRoom
// and the wall runs produce them (levels/drlg_l3.cpp:1617-1630).
var (
	l3FenceDoorLeft  = [...]Tile{130, 132, 133, 134, 136, 138, 140}
	l3FenceDoorRight = [...]Tile{131, 132, 133, 135, 137, 138, 139}
)

// fence is Fence: it walls off long dirt runs, re-tiles the wall lines it finds,
// then closes the doors that no longer fit (levels/drlg_l3.cpp:1634-1798).
func (d *Dungeon) fence() {
	d.fenceRuns()
	d.fenceLines()

	d.addFenceDoors()
	d.fenceDoorFix()
}

// fenceRuns is Fence's first scan: it turns every long horizontal or vertical
// dirt run into the fence tiles (levels/drlg_l3.cpp:1636-1704).
func (d *Dungeon) fenceRuns() {
	// The 1 and -1 bounds are upstream's BUGFIX marks: the scan used to start
	// at the map edge, which the wall helpers cannot read.
	for j := 1; j < DMAXY-1; j++ {
		for i := 1; i < DMAXX-1; i++ {
			d.fenceRunAt(i, j)
		}
	}
}

// fenceRunAt is one cell of Fence's first scan
// (levels/drlg_l3.cpp:1639-1702).
func (d *Dungeon) fenceRunAt(i, j int) {
	if d.Tiles[i][j] == 10 && !d.Rnd.FlipCoin(2) {
		d.carveFenceRow(i, j)
	}

	if d.Tiles[i][j] == 9 && !d.Rnd.FlipCoin(2) {
		d.carveFenceColumn(i, j)
	}

	if d.Tiles[i][j] == 11 && d.Tiles[i+1][j] == 10 && d.Tiles[i][j+1] == 9 && !d.Rnd.FlipCoin(2) {
		d.carveFenceCorner(i, j)
	}
}

// carveFenceRow walls a horizontal dirt run: a start tile, a random middle and
// an end tile (levels/drlg_l3.cpp:1641-1651).
func (d *Dungeon) carveFenceRow(i, j int) {
	end := d.runEnd(Point{X: i, Y: j}, Point{X: 1, Y: 0}, 10)
	if end-i <= 0 {
		return
	}

	d.Tiles[i][j] = 127
	for xx := i + 1; xx < end; xx++ {
		d.Tiles[xx][j] = Tile(d.pickRandomlyAmong([2]int{129, 126}))
	}

	d.Tiles[end][j] = 128
}

// carveFenceColumn walls a vertical dirt run (levels/drlg_l3.cpp:1653-1663).
func (d *Dungeon) carveFenceColumn(i, j int) {
	end := d.runEnd(Point{X: i, Y: j}, Point{X: 0, Y: 1}, 9)
	if end-j <= 0 {
		return
	}

	d.Tiles[i][j] = 123
	for yy := j + 1; yy < end; yy++ {
		d.Tiles[i][yy] = Tile(d.pickRandomlyAmong([2]int{124, 121}))
	}

	d.Tiles[i][end] = 122
}

// carveFenceCorner walls the corner of a horizontal and a vertical dirt run
// (levels/drlg_l3.cpp:1665-1701).
func (d *Dungeon) carveFenceCorner(i, j int) {
	d.Tiles[i][j] = 125

	endX := d.runEnd(Point{X: i + 1, Y: j}, Point{X: 1, Y: 0}, 10)
	for xx := i + 1; xx < endX; xx++ {
		d.Tiles[xx][j] = Tile(d.pickRandomlyAmong([2]int{129, 126}))
	}
	d.Tiles[endX][j] = 128

	endY := d.runEnd(Point{X: i, Y: j + 1}, Point{X: 0, Y: 1}, 9)
	for yy := j + 1; yy < endY; yy++ {
		d.Tiles[i][yy] = Tile(d.pickRandomlyAmong([2]int{124, 121}))
	}
	d.Tiles[i][endY] = 122
}

// runEnd finds the last tile of a run of tile from start, stepping by step. The
// scan may leave the layer, where upstream's flat address lands in the
// neighbouring array (levels/drlg_l3.cpp:1643-1646).
func (d *Dungeon) runEnd(start, step Point, tile Tile) int {
	at := start
	for d.flatTile(at.X, at.Y) == tile {
		at = at.Add(step)
	}

	if step.X != 0 {
		return at.X - 1
	}

	return at.Y - 1
}

// fenceLines is Fence's second scan: where a floor tile sits away from a theme
// room it walls the line it belongs to and leaves one passable tile in it
// (levels/drlg_l3.cpp:1706-1794).
func (d *Dungeon) fenceLines() {
	for j := 1; j < DMAXY; j++ {
		for i := 1; i < DMAXX; i++ {
			d.fenceLineAt(i, j)
		}
	}
}

// fenceLineAt is one cell of Fence's second scan. The RNG discard runs before
// the theme-room test because upstream uses the comma operator
// (levels/drlg_l3.cpp:1708-1710).
func (d *Dungeon) fenceLineAt(i, j int) {
	if d.Tiles[i][j] != caveFloor {
		return
	}

	d.Rnd.Discard(1)
	if isNearThemeRoom(d.ThemeRooms, Point{X: i, Y: j}) {
		return
	}

	if d.Rnd.FlipCoin(2) {
		d.fenceVerticalLine(i, j)

		return
	}

	d.fenceHorizontalLine(i, j)
}

// fenceVerticalLine walls the vertical line through i,j, leaving the floor tile
// the dice pick (levels/drlg_l3.cpp:1711-1746).
func (d *Dungeon) fenceVerticalLine(i, j int) {
	y1, y2, ok := d.fenceVerticalSpan(i, j)
	if !ok {
		return
	}

	pass := int(d.Rnd.GenerateRnd(int32(y2-y1-1))) + y1 + 1
	for y := y1; y <= y2; y++ {
		if y != pass {
			d.Tiles[i][y] = d.fenceVerticalTile(d.Tiles[i][y])
		}
	}
}

// fenceVerticalSpan is the walled run the vertical fence line covers, or false
// when the run is too short or already floor at either end
// (levels/drlg_l3.cpp:1713-1724).
func (d *Dungeon) fenceVerticalSpan(i, j int) (y1, y2 int, ok bool) {
	y1 = j
	for y1 >= 0 && d.fenceVerticalUp(i, y1) {
		y1--
	}
	y1++

	y2 = j
	for y2 < DMAXY && d.fenceVerticalDown(i, y2) {
		y2++
	}
	y2--

	if y2-y1 <= 1 || d.Tiles[i][y1] == caveFloor || d.Tiles[i][y2] == caveFloor {
		return 0, 0, false
	}

	return y1, y2, true
}

// fenceHorizontalLine walls the horizontal line through i,j
// (levels/drlg_l3.cpp:1747-1780).
func (d *Dungeon) fenceHorizontalLine(i, j int) {
	x1, x2, ok := d.fenceHorizontalSpan(i, j)
	if !ok {
		return
	}

	pass := int(d.Rnd.GenerateRnd(int32(x2-x1-1))) + x1 + 1
	for x := x1; x <= x2; x++ {
		if x != pass {
			d.Tiles[x][j] = d.fenceHorizontalTile(d.Tiles[x][j])
		}
	}
}

// fenceHorizontalSpan is the walled run the horizontal fence line covers, or
// false when the run is too short or already floor at either end
// (levels/drlg_l3.cpp:1749-1760).
func (d *Dungeon) fenceHorizontalSpan(i, j int) (x1, x2 int, ok bool) {
	x1 = i
	for x1 >= 0 && d.fenceHorizontalLeft(x1, j) {
		x1--
	}
	x1++

	x2 = i
	for x2 < DMAXX && d.fenceHorizontalRight(x2, j) {
		x2++
	}
	x2--

	if x2-x1 <= 1 || d.Tiles[x1][j] == caveFloor || d.Tiles[x2][j] == caveFloor {
		return 0, 0, false
	}

	return x1, x2, true
}

// fenceVerticalTile is the vertical line substitution: a floor tile becomes one
// of the two fence faces, every other candidate keeps its own replacement
// (levels/drlg_l3.cpp:1728-1743).
func (d *Dungeon) fenceVerticalTile(tile Tile) Tile {
	if tile == caveFloor {
		return Tile(d.pickRandomlyAmong([2]int{137, 135}))
	}

	if sub := l3FenceVerticalSub[tile]; sub != 0 {
		return sub
	}

	return tile
}

// fenceHorizontalTile is the horizontal line substitution
// (levels/drlg_l3.cpp:1766-1778).
func (d *Dungeon) fenceHorizontalTile(tile Tile) Tile {
	if tile == caveFloor {
		return Tile(d.pickRandomlyAmong([2]int{136, 134}))
	}

	if sub := l3FenceHorizontalSub[tile]; sub != 0 {
		return sub
	}

	return tile
}

// fenceVerticalUp reports whether the line above i,y can still be walled: the
// tiles either side must not be fence tiles themselves
// (levels/drlg_l3.cpp:1530-1540).
func (d *Dungeon) fenceVerticalUp(i, y int) bool {
	if isFenceTile(d.Tiles[i+1][y]) || isFenceTile(d.Tiles[i-1][y]) {
		return false
	}

	return d.Tiles[i][y] == 7 || d.Tiles[i][y] == 10 || d.Tiles[i][y] == 126 ||
		d.Tiles[i][y] == 129 || d.Tiles[i][y] == 134 || d.Tiles[i][y] == 136
}

// fenceVerticalDown is the downwards counterpart (levels/drlg_l3.cpp:1542-1552).
func (d *Dungeon) fenceVerticalDown(i, y int) bool {
	if isFenceTile(d.Tiles[i+1][y]) || isFenceTile(d.Tiles[i-1][y]) {
		return false
	}

	return d.Tiles[i][y] == 2 || d.Tiles[i][y] == 7 ||
		d.Tiles[i][y] == 134 || d.Tiles[i][y] == 136
}

// fenceHorizontalLeft reports whether the line left of x,j can still be walled
// (levels/drlg_l3.cpp:1554-1564).
func (d *Dungeon) fenceHorizontalLeft(x, j int) bool {
	if isFenceTile(d.Tiles[x][j+1]) || isFenceTile(d.Tiles[x][j-1]) {
		return false
	}

	return d.Tiles[x][j] == 7 || d.Tiles[x][j] == 9 || d.Tiles[x][j] == 121 ||
		d.Tiles[x][j] == 124 || d.Tiles[x][j] == 135 || d.Tiles[x][j] == 137
}

// fenceHorizontalRight is the rightwards counterpart
// (levels/drlg_l3.cpp:1566-1576).
func (d *Dungeon) fenceHorizontalRight(x, j int) bool {
	if isFenceTile(d.Tiles[x][j+1]) || isFenceTile(d.Tiles[x][j-1]) {
		return false
	}

	return d.Tiles[x][j] == 4 || d.Tiles[x][j] == 7 ||
		d.Tiles[x][j] == 135 || d.Tiles[x][j] == 137
}

// isFenceTile reports whether tile is inside the fence tile range
// (levels/drlg_l3.cpp:1532-1533).
func isFenceTile(tile Tile) bool {
	return tile >= 130 && tile <= 152
}

// addFenceDoors turns a floor tile walled in on both sides into a door tile
// (levels/drlg_l3.cpp:1578-1598).
func (d *Dungeon) addFenceDoors() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.addFenceDoorAt(i, j)
		}
	}
}

// addFenceDoorAt is one cell of AddFenceDoors. Neighbours are read through the
// flat address because the scan starts at the map edge, where upstream reads
// the array that precedes or follows the tile layer
// (levels/drlg_l3.cpp:1580-1597).
func (d *Dungeon) addFenceDoorAt(i, j int) {
	if d.Tiles[i][j] != caveFloor {
		return
	}

	if isFenceTile(d.flatTile(i-1, j)) && isFenceTile(d.flatTile(i+1, j)) {
		d.Tiles[i][j] = caveThemeDoorHorizontal

		return
	}

	if isFenceTile(d.flatTile(i, j-1)) && isFenceTile(d.flatTile(i, j+1)) {
		d.Tiles[i][j] = caveThemeDoorVertical
	}
}

// fenceDoorFix drops every door tile that no longer has walls on both sides
// (levels/drlg_l3.cpp:1600-1632).
func (d *Dungeon) fenceDoorFix() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.fenceDoorFixAt(i, j)
		}
	}
}

// fenceDoorFixAt is one cell of FenceDoorFix: a horizontal door needs fence
// tiles left and right, a vertical one needs them above and below, and each
// door is dropped again when the wall set around it cannot carry it
// (levels/drlg_l3.cpp:1602-1630).
func (d *Dungeon) fenceDoorFixAt(i, j int) {
	switch d.Tiles[i][j] {
	case caveThemeDoorHorizontal:
		d.fixHorizontalFenceDoor(i, j)
	case caveThemeDoorVertical:
		d.fixVerticalFenceDoor(i, j)
	default:
	}
}

// fixHorizontalFenceDoor drops a horizontal door whose two side neighbours are
// not both fence, then drops it again unless one of them carries the wall set
// (levels/drlg_l3.cpp:1604-1612).
func (d *Dungeon) fixHorizontalFenceDoor(i, j int) {
	if !d.fenceTilePair(i-1, j, i+1, j) {
		d.Tiles[i][j] = caveFloor

		return
	}

	if !holdsAny(d.flatTile(i-1, j), l3FenceDoorLeft[:]) &&
		!holdsAny(d.flatTile(i+1, j), l3FenceDoorLeft[:]) {
		d.Tiles[i][j] = caveFloor
	}
}

// fixVerticalFenceDoor is the vertical counterpart of fixHorizontalFenceDoor
// (levels/drlg_l3.cpp:1613-1622).
func (d *Dungeon) fixVerticalFenceDoor(i, j int) {
	if !d.fenceTilePair(i, j-1, i, j+1) {
		d.Tiles[i][j] = caveFloor

		return
	}

	if !holdsAny(d.flatTile(i, j-1), l3FenceDoorRight[:]) &&
		!holdsAny(d.flatTile(i, j+1), l3FenceDoorRight[:]) {
		d.Tiles[i][j] = caveFloor
	}
}

// fenceTilePair reports whether both flat-addressed neighbours are fence tiles.
func (d *Dungeon) fenceTilePair(x1, y1, x2, y2 int) bool {
	return isFenceTile(d.flatTile(x1, y1)) && isFenceTile(d.flatTile(x2, y2))
}

// holdsAny reports whether tile is one of set.
func holdsAny(tile Tile, set []Tile) bool {
	for _, candidate := range set {
		if tile == candidate {
			return true
		}
	}

	return false
}
