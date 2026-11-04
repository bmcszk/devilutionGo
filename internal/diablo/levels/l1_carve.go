package levels

// This file ports the carve stage of the cathedral generator: the passes that
// turn the boolean room mask into the mega-tile layer and then decorate the
// walls (levels/drlg_l1.cpp:473-995, 1036-1075, 1155-1201, 1212-1256).

// wallRunRules is AddWall's test chain in upstream order. The tests are not
// exclusive: a run rewrites the tile, so a later test sees the new shape.
var wallRunRules = []wallRunRule{
	{tile: catCorner, vertical: false, startTile: catHWall},
	{tile: catCorner, vertical: true, startTile: catVWall},
	{tile: catVWallEnd, vertical: false, startTile: catDWall},
	{tile: catHWallEnd, vertical: true, startTile: catDWall},
	{tile: catHWall, vertical: false, startTile: catHWall},
	{tile: catVWall, vertical: true, startTile: catVWall},
}

// wallRunRule is one entry of AddWall's sequential test chain: the tile shape
// that triggers a run, the direction it runs in, and the start tile it draws
// (levels/drlg_l1.cpp:758-797).
type wallRunRule struct {
	tile      Tile
	vertical  bool
	startTile Tile
}

// wallStyle is the tile set one wall run draws with
// (levels/drlg_l1.cpp:669-695, 711-737).
type wallStyle struct {
	wall  Tile
	door  Tile
	start Tile
}

// initDungeonFlags resets the tile layer to dirt and clears the two flag masks
// (levels/drlg_l1.cpp:473-479).
func (d *Dungeon) initDungeonFlags() {
	for i := range DMAXX {
		for j := range DMAXY {
			d.Tiles[i][j] = catDirt
		}
	}

	d.Protected.Reset()
	d.Chamber.Reset()
}

// makeDmt turns the layout mask into the mega-tile layer: floor for masked
// cells, a wall tile for the neighbours, keeping diagonal corners open
// (levels/drlg_l1.cpp:607-628). Note the DMAXX-1/DMAXY-1 bounds: the last row
// and column of the layer stay dirt.
func (d *Dungeon) makeDmt() {
	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			d.Tiles[i][j] = d.makeDmtTile(i, j)
		}
	}
}

// makeDmtTile is MakeDmt's per-cell decision (levels/drlg_l1.cpp:609-625).
func (d *Dungeon) makeDmtTile(i, j int) Tile {
	switch {
	case d.Mask.Test(i, j):
		return catFloor
	case d.isEmptyDiagonalCorner(i, j):
		return catFloor // Remove diagonal corners
	case d.isMaskCorner(i, j):
		return catVCorner
	case d.Mask.Test(i, j+1):
		return catHWall
	case d.Mask.Test(i+1, j):
		return catVWall
	case d.Mask.Test(i+1, j+1):
		return catDWall
	default:
		return catDirt
	}
}

// isEmptyDiagonalCorner reports the mask shape MakeDmt fills as floor to keep
// the diagonal open (levels/drlg_l1.cpp:612-613).
func (d *Dungeon) isEmptyDiagonalCorner(i, j int) bool {
	return !d.Mask.Test(i+1, j+1) && d.Mask.Test(i, j+1) && d.Mask.Test(i+1, j)
}

// isMaskCorner reports the mask shape MakeDmt turns into a corner tile
// (levels/drlg_l1.cpp:614-615).
func (d *Dungeon) isMaskCorner(i, j int) bool {
	return d.Mask.Test(i+1, j+1) && d.Mask.Test(i, j+1) && d.Mask.Test(i+1, j)
}

// tileInLayer reads one mega tile, reporting false outside the layer. Upstream
// indexes `dungeon` without bounds checks in the wall scans; every reachable
// out-of-range read there lands on the untouched last row/column, which
// initDungeonFlags leaves at Dirt, so "outside the layer" and "not a Floor
// tile" are the same answer for those scans (levels/drlg_l1.cpp:629-666).
func (d *Dungeon) tileInLayer(x, y int) (Tile, bool) {
	if x < 0 || x >= DMAXX || y < 0 || y >= DMAXY {
		return catDirt, false
	}

	return d.Tiles[x][y], true
}

// isWallReplaceable reports whether a wall run may end on tileID
// (levels/drlg_l1.cpp:639-641).
func isWallReplaceable(tileID Tile) bool {
	switch tileID {
	case catCorner, catDWall, catDArch, catVWallEnd, catHWallEnd, catVCorner, catHCorner,
		catDirtHwall, catDirtVwall, catVDirtCorner, catHDirtCorner, catDirtHwallEnd, catDirtVwallEnd:
		return true
	default:
		return false
	}
}

// horizontalWallOk reports the usable run length of a horizontal wall at
// position, or -1 (levels/drlg_l1.cpp:629-646).
func (d *Dungeon) horizontalWallOk(position Point) int {
	return d.wallRunOk(position, Point{X: 1}, Point{Y: -1}, Point{Y: 1})
}

// verticalWallOk reports the usable run length of a vertical wall at position,
// or -1 (levels/drlg_l1.cpp:648-665).
func (d *Dungeon) verticalWallOk(position Point) int {
	return d.wallRunOk(position, Point{Y: 1}, Point{X: -1}, Point{X: 1})
}

// wallRunOk walks a wall run from position in direction step, counting the
// floor cells whose two neighbouring cells (sideA, sideB) stay floor and are
// neither protected nor chamber. It returns the run length the wall may span,
// or -1 when the run ends on a tile a wall may not abut
// (levels/drlg_l1.cpp:629-665).
func (d *Dungeon) wallRunOk(position, step, sideA, sideB Point) int {
	length := 1

	for d.wallRunContinues(position, step, sideA, sideB, length) {
		length++
	}

	if length == 1 {
		return -1
	}

	end := position.Add(Point{X: step.X * length, Y: step.Y * length})
	if tile, _ := d.tileInLayer(end.X, end.Y); !isWallReplaceable(tile) {
		return -1
	}

	return length
}

// wallRunContinues reports whether the run still extends over cell length
// (levels/drlg_l1.cpp:631-642, 650-661).
func (d *Dungeon) wallRunContinues(position, step, sideA, sideB Point, length int) bool {
	cell := position.Add(Point{X: step.X * length, Y: step.Y * length})

	if tile, ok := d.tileInLayer(cell.X, cell.Y); !ok || tile != catFloor {
		return false
	}

	nearA, _ := d.tileInLayer(cell.X+sideA.X, cell.Y+sideA.Y)
	nearB, _ := d.tileInLayer(cell.X+sideB.X, cell.Y+sideB.Y)

	return nearA == catFloor && nearB == catFloor &&
		!d.Protected.Test(cell.X, cell.Y) && !d.Chamber.Test(cell.X, cell.Y)
}

// horizontalWallStyle draws the arch/fence variant of a horizontal run
// (levels/drlg_l1.cpp:669-695).
func (d *Dungeon) horizontalWallStyle(start Tile) wallStyle {
	style := wallStyle{wall: catHWall, door: catHDoor, start: start}

	switch d.Rnd.GenerateRnd(4) {
	case 2: // Add arch
		style.wall = catHArch
		style.door = catHArch
		style.start = skirtTile(start, catHWall, catHArch, catHArchVWall)
	case 3: // Add fence
		style.wall = catHFence
		style.start = skirtTile(start, catHWall, catHFence, catHFenceVWall)
	default:
	}

	if d.Rnd.GenerateRnd(6) == 5 {
		style.door = catHArch
	}

	return style
}

// verticalWallStyle draws the arch/fence variant of a vertical run
// (levels/drlg_l1.cpp:711-737).
func (d *Dungeon) verticalWallStyle(start Tile) wallStyle {
	style := wallStyle{wall: catVWall, door: catVDoor, start: start}

	switch d.Rnd.GenerateRnd(4) {
	case 2: // Add arch
		style.wall = catVArch
		style.door = catVArch
		style.start = skirtTile(start, catVWall, catVArch, catHWallVArch)
	case 3: // Add fence
		style.wall = catVFence
		style.start = skirtTile(start, catVWall, catVFence, catHWallVFence)
	default:
	}

	if d.Rnd.GenerateRnd(6) == 5 {
		style.door = catVArch
	}

	return style
}

// skirtTile is the start-tile half of the HorizontalWall/VerticalWall style
// switch: a plain wall start follows the run's style, a dirt-wall start keeps
// its own arch or fence variant (levels/drlg_l1.cpp:672-693, 714-735).
func skirtTile(start, plainWall, styleWall, dirtStart Tile) Tile {
	switch start {
	case plainWall:
		return styleWall
	case catDWall:
		return dirtStart
	default:
		return start
	}
}

// horizontalWall writes one horizontal wall run: the start tile, the body and
// the door, protecting the door when it is a plain one
// (levels/drlg_l1.cpp:667-707).
func (d *Dungeon) horizontalWall(position Point, start Tile, maxX int) {
	style := d.horizontalWallStyle(start)

	d.Tiles[position.X][position.Y] = style.start
	for x := 1; x < maxX; x++ {
		d.Tiles[position.X+x][position.Y] = style.wall
	}

	offset := int(d.Rnd.GenerateRnd(int32(maxX-1))) + 1
	d.Tiles[position.X+offset][position.Y] = style.door

	if style.door == catHDoor {
		d.Protected.Set(position.X+offset, position.Y, true)
	}
}

// verticalWall writes one vertical wall run (levels/drlg_l1.cpp:709-749).
func (d *Dungeon) verticalWall(position Point, start Tile, maxY int) {
	style := d.verticalWallStyle(start)

	d.Tiles[position.X][position.Y] = style.start
	for y := 1; y < maxY; y++ {
		d.Tiles[position.X][position.Y+y] = style.wall
	}

	offset := int(d.Rnd.GenerateRnd(int32(maxY-1))) + 1
	d.Tiles[position.X][position.Y+offset] = style.door

	if style.door == catVDoor {
		d.Protected.Set(position.X, position.Y+offset, true)
	}
}

// addWallRuns decorates one tile with the wall runs its shape allows
// (levels/drlg_l1.cpp:755-799). Every rule that fires discards one random value
// before it measures the run, whether or not a run is placed.
func (d *Dungeon) addWallRuns(i, j int) {
	position := Point{X: i, Y: j}

	for _, rule := range wallRunRules {
		if d.Tiles[i][j] != rule.tile {
			continue
		}

		d.Rnd.Discard(1)
		d.raiseWall(position, rule)
	}
}

// raiseWall places the wall run one rule asks for, if the layer allows it
// (levels/drlg_l1.cpp:761-796).
func (d *Dungeon) raiseWall(position Point, rule wallRunRule) {
	if rule.vertical {
		if maxY := d.verticalWallOk(position); maxY != -1 {
			d.verticalWall(position, rule.startTile, maxY)
		}

		return
	}

	if maxX := d.horizontalWallOk(position); maxX != -1 {
		d.horizontalWall(position, rule.startTile, maxX)
	}
}

// addWall raises the wall runs on every unprotected, non-chamber tile
// (levels/drlg_l1.cpp:751-803).
func (d *Dungeon) addWall() {
	for j := range DMAXY {
		for i := range DMAXX {
			if d.Protected.Test(i, j) || d.Chamber.Test(i, j) {
				continue
			}

			d.addWallRuns(i, j)
		}
	}
}
