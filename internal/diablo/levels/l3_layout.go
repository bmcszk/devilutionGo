package levels

// This file ports the caves layout half of GenerateLevel: the room and block
// maker, the four fill passes, the floor-area gate and Lockout's reachability
// check (levels/drlg_l3.cpp:711-1014, 1890-1927).

// caveMinArea is the minimum floor count a caves layout must reach before the
// stairs are attempted (levels/drlg_l3.cpp:2032).
const caveMinArea = 600

// caveFillRun is FillStraights' hard-coded inner scan length; it stops three
// tiles short of the map edge (levels/drlg_l3.cpp:896, 921, 941, 961).
const caveFillRun = 37

// l3StraightRules are FillStraights' four sweeps in upstream order
// (levels/drlg_l3.cpp:895-971): each rule scans the 37-long runs of a row pair
// (vertical=false) or column pair, and fills the near or the far side of the
// pair with a random dirt/floor draw once a run passes three tiles.
var l3StraightRules = [...]straightRule{
	{vertical: false, near: 0, far: 1, writeFar: false},
	{vertical: false, near: 1, far: 0, writeFar: true},
	{vertical: true, near: 0, far: 1, writeFar: false},
	{vertical: true, near: 1, far: 0, writeFar: true},
}

// straightRule is one FillStraights sweep: the tile pair it looks for on a
// row or column pair, and which side of the pair it fills.
type straightRule struct {
	vertical bool
	near     Tile
	far      Tile
	writeFar bool
}

// initDungeonFlags is InitDungeonFlags: it clears the tile layer and the
// protected mask before a layout attempt (levels/drlg_l3.cpp:711-715).
func (d *Dungeon) initCaveDungeonFlags() {
	d.Tiles = [DMAXX][DMAXY]Tile{}
	d.Protected.Reset()
}

// fillRoom fills a block's interior and stamps a random border, reporting
// whether the area was free (levels/drlg_l3.cpp:717-757).
func (d *Dungeon) fillRoom(x1, y1, x2, y2 int) bool {
	if x1 <= 1 || x2 >= 34 || y1 <= 1 || y2 >= 38 {
		return false
	}

	if d.roomAreaSum(x1, y1, x2, y2) != 0 {
		return false
	}

	d.fillRoomInterior(x1, y1, x2, y2)
	d.stampRoomBorder(x1, y1, x2, y2)

	return true
}

// roomAreaSum is FillRoom's overlap test: the sum of every tile in the area.
func (d *Dungeon) roomAreaSum(x1, y1, x2, y2 int) int {
	sum := 0
	for j := y1; j <= y2; j++ {
		for i := x1; i <= x2; i++ {
			sum += int(d.Tiles[i][j])
		}
	}

	return sum
}

// fillRoomInterior sets the area minus its one-tile border to floor.
func (d *Dungeon) fillRoomInterior(x1, y1, x2, y2 int) {
	for j := y1 + 1; j < y2; j++ {
		for i := x1 + 1; i < x2; i++ {
			d.Tiles[i][j] = 1
		}
	}
}

// stampRoomBorder draws the block outline, each border tile a coin flip away
// from being left as dirt (levels/drlg_l3.cpp:744-755).
func (d *Dungeon) stampRoomBorder(x1, y1, x2, y2 int) {
	for j := y1; j <= y2; j++ {
		d.stampBorderTile(x1, j)
		d.stampBorderTile(x2, j)
	}

	for i := x1; i <= x2; i++ {
		d.stampBorderTile(i, y1)
		d.stampBorderTile(i, y2)
	}
}

// stampBorderTile floors one border tile unless the coin keeps it as dirt
// (levels/drlg_l3.cpp:746-754).
func (d *Dungeon) stampBorderTile(i, j int) {
	if !d.Rnd.FlipCoin(2) {
		d.Tiles[i][j] = 1
	}
}

// createBlock grows one block next to point and recurses on its neighbours
// (levels/drlg_l3.cpp:759-843).
func (d *Dungeon) createBlock(point Point, obs, dir int) {
	blockSizeX := int(d.Rnd.RandomIntLessThan(2)) + 3
	blockSizeY := int(d.Rnd.RandomIntLessThan(2)) + 3

	low, high := d.blockCorners(point, obs, dir, blockSizeX, blockSizeY)
	if !d.fillRoom(low.X, low.Y, high.X, high.Y) {
		return
	}

	if d.Rnd.FlipCoin(4) {
		return
	}

	d.growBlockNeighbours(dir, low.X, low.Y, high.X, high.Y, blockSizeX, blockSizeY)
}

// growBlockNeighbours is the recursive half of CreateBlock: each of the four
// sides grows a block of its own unless it is the side we came from
// (levels/drlg_l3.cpp:836-842).
func (d *Dungeon) growBlockNeighbours(dir, x1, y1, x2, y2, blockSizeX, blockSizeY int) {
	if dir != 2 {
		d.createBlock(Point{X: x1, Y: y1}, blockSizeY, 0)
	}
	if dir != 3 {
		d.createBlock(Point{X: x2, Y: y1}, blockSizeX, 1)
	}
	if dir != 0 {
		d.createBlock(Point{X: x1, Y: y2}, blockSizeY, 2)
	}
	if dir != 1 {
		d.createBlock(Point{X: x1, Y: y1}, blockSizeX, 3)
	}
}

// blockCorners is CreateBlock's per-direction geometry: the two corners of the
// new block, placed against point on the side dir names
// (levels/drlg_l3.cpp:764-822). obs decides how far the block slides along the
// perpendicular axis.
func (d *Dungeon) blockCorners(point Point, obs, dir, blockSizeX, blockSizeY int) (low, high Point) {
	x1, y1, x2, y2 := 0, 0, 0, 0
	switch dir {
	case 0:
		y2 = point.Y - 1
		y1 = y2 - blockSizeY
		x1 = d.blockOffset(point.X, blockSizeX, obs)
		x2 = blockSizeX + x1
	case 1:
		x1 = point.X + 1
		x2 = x1 + blockSizeX
		y1 = d.blockOffset(point.Y, blockSizeY, obs)
		y2 = y1 + blockSizeY
	case 2:
		y1 = point.Y + 1
		y2 = y1 + blockSizeY
		x1 = d.blockOffset(point.X, blockSizeX, obs)
		x2 = blockSizeX + x1
	case 3:
		x2 = point.X - 1
		x1 = x2 - blockSizeX
		y1 = d.blockOffset(point.Y, blockSizeY, obs)
		y2 = y1 + blockSizeY
	default:
		// dir is one of 0..3 at every call site, upstream has no default
	}

	return Point{X: x1, Y: y1}, Point{X: x2, Y: y2}
}

// blockOffset is the shared slide of CreateBlock's three-branch axis
// placement (levels/drlg_l3.cpp:770-778): a block smaller than the observed
// room starts inside it, a larger one straddles the origin.
func (d *Dungeon) blockOffset(origin, blockSize, obs int) int {
	switch {
	case blockSize < obs:
		return int(d.Rnd.GenerateRnd(int32(blockSize))) + origin
	case blockSize == obs:
		return origin
	default:
		return origin - int(d.Rnd.GenerateRnd(int32(blockSize)))
	}
}

// floorArea paints a filled floor rectangle (levels/drlg_l3.cpp:845-852).
func (d *Dungeon) floorArea(x1, y1, x2, y2 int) {
	for j := y1; j <= y2; j++ {
		for i := x1; i <= x2; i++ {
			d.Tiles[i][j] = 1
		}
	}
}

// fillDiagonals closes the two-tile diagonal gaps of the 2x2 pattern sum
// (levels/drlg_l3.cpp:854-875).
func (d *Dungeon) fillDiagonals() {
	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			d.fillDiagonalGap(i, j)
		}
	}
}

// fillDiagonalGap closes the diagonal gap the 2x2 pattern sum names at i,j
// (levels/drlg_l3.cpp:856-873).
func (d *Dungeon) fillDiagonalGap(i, j int) {
	switch d.diagonalPattern(i, j) {
	case 6:
		d.fillOneDiagonal(i, j, Point{X: 0, Y: 0}, Point{X: 1, Y: 1})
	case 9:
		d.fillOneDiagonal(i, j, Point{X: 1, Y: 0}, Point{X: 0, Y: 1})
	default:
	}
}

// fillOneDiagonal floors the corner the coin names, measured from i,j.
func (d *Dungeon) fillOneDiagonal(i, j int, first, second Point) {
	corner := second
	if d.Rnd.FlipCoin(2) {
		corner = first
	}

	d.Tiles[i+corner.X][j+corner.Y] = 1
}

// diagonalPattern is FillDiagonals' 2x2 bit sum, north-west bit last.
func (d *Dungeon) diagonalPattern(i, j int) int {
	return int(d.Tiles[i+1][j+1]) + 2*int(d.Tiles[i][j+1]) +
		4*int(d.Tiles[i+1][j]) + 8*int(d.Tiles[i][j])
}

// fillSingles turns an isolated dirt cell into floor when all eight of its
// neighbours are floor (levels/drlg_l3.cpp:877-889).
func (d *Dungeon) fillSingles() {
	for j := 1; j < DMAXY-1; j++ {
		for i := 1; i < DMAXX-1; i++ {
			if d.isSingleHole(i, j) {
				d.Tiles[i][j] = 1
			}
		}
	}
}

// isSingleHole reports whether the cell and its neighbourhood are the
// surrounded-single pattern FillSingles looks for.
func (d *Dungeon) isSingleHole(i, j int) bool {
	if d.Tiles[i][j] != 0 {
		return false
	}

	rowAbove := d.Tiles[i][j-1] + d.Tiles[i-1][j-1] + d.Tiles[i+1][j-1]
	rowBeside := d.Tiles[i+1][j] + d.Tiles[i-1][j]
	rowBelow := d.Tiles[i][j+1] + d.Tiles[i-1][j+1] + d.Tiles[i+1][j+1]

	return rowAbove == 3 && rowBeside == 2 && rowBelow == 3
}

// fillStraights closes the long straight gaps FillSingles cannot see
// (levels/drlg_l3.cpp:891-972).
func (d *Dungeon) fillStraights() {
	for _, rule := range l3StraightRules {
		d.fillStraightLines(rule)
	}
}

// fillStraightLines scans one orientation of row or column pairs. Both
// orientations run over 39 lines because DMAXX-1 == DMAXY-1.
func (d *Dungeon) fillStraightLines(rule straightRule) {
	for line := range DMAXY - 1 {
		d.fillStraightLine(rule, line)
	}
}

// fillStraightLine is one line of a FillStraights pass: it measures each run of
// the pattern and fills it when the run ends. A run that reaches the scan
// limit is dropped, as upstream never flushes it
// (levels/drlg_l3.cpp:896-919).
func (d *Dungeon) fillStraightLine(rule straightRule, line int) {
	runStart := 0
	runLen := 0

	for index := range caveFillRun {
		if d.straightNear(rule, line, index) == rule.near &&
			d.straightFar(rule, line, index) == rule.far {
			if runLen == 0 {
				runStart = index
			}
			runLen++

			continue
		}

		d.fillStraightRun(rule, line, runStart, index, runLen)
		runStart, runLen = 0, 0
	}
}

// fillStraightRun fills a measured run with random dirt or floor ids.
func (d *Dungeon) fillStraightRun(rule straightRule, line, start, end, runLen int) {
	if runLen <= 3 || d.Rnd.FlipCoin(2) {
		return
	}

	for k := start; k < end; k++ {
		d.setStraightCell(rule, line, k)
	}
}

// straightNear reads the near cell of the pair at line,index.
func (d *Dungeon) straightNear(rule straightRule, line, index int) Tile {
	if rule.vertical {
		return d.Tiles[line][index]
	}

	return d.Tiles[index][line]
}

// straightFar reads the far cell of the pair at line,index.
func (d *Dungeon) straightFar(rule straightRule, line, index int) Tile {
	if rule.vertical {
		return d.Tiles[line+1][index]
	}

	return d.Tiles[index][line+1]
}

// setStraightCell writes the random fill into the side of the pair the rule
// names (levels/drlg_l3.cpp:904, 929, 949, 969).
func (d *Dungeon) setStraightCell(rule straightRule, line, index int) {
	tile := Tile(d.Rnd.GenerateRnd(2))

	if rule.vertical {
		if rule.writeFar {
			d.Tiles[line+1][index] = tile
		} else {
			d.Tiles[line][index] = tile
		}

		return
	}

	if rule.writeFar {
		d.Tiles[index][line+1] = tile
	} else {
		d.Tiles[index][line] = tile
	}
}

// clearEdges keeps the last row and column of the raw layer empty
// (levels/drlg_l3.cpp:974-982).
func (d *Dungeon) clearEdges() {
	for j := range DMAXY {
		d.Tiles[DMAXX-1][j] = 0
	}

	for i := range DMAXX {
		d.Tiles[i][DMAXY-1] = 0
	}
}

// getFloorArea counts the floor cells of the raw layer
// (levels/drlg_l3.cpp:984-995).
func (d *Dungeon) getFloorArea() int {
	area := 0
	for j := range DMAXY {
		for i := range DMAXX {
			area += int(d.Tiles[i][j])
		}
	}

	return area
}

// makeMegas converts the raw 0/1 layer into mega-tile ids through L3ConvTbl
// (levels/drlg_l3.cpp:997-1015).
func (d *Dungeon) makeMegas() {
	for j := range DMAXY - 1 {
		for i := range DMAXX - 1 {
			d.Tiles[i][j] = d.convertedMega(i, j)
		}
	}

	d.edgeMegas()
}

// convertedMega is MakeMegas' 2x2 pattern lookup. Patterns 6 and 9 are drawn
// from a pair of substitutes to break up the largest cave walls
// (levels/drlg_l3.cpp:1000-1010).
func (d *Dungeon) convertedMega(i, j int) Tile {
	v := d.diagonalPattern(i, j)

	if v == 6 {
		v = d.pickRandomlyAmong([2]int{12, 5})
	}
	if v == 9 {
		v = d.pickRandomlyAmong([2]int{13, 14})
	}

	return l3ConvTbl[v]
}

// edgeMegas marks the last row and column of the mega layer as east wall
// (levels/drlg_l3.cpp:1011-1014).
func (d *Dungeon) edgeMegas() {
	for j := range DMAXY {
		d.Tiles[DMAXX-1][j] = caveCeiling
	}

	for i := range DMAXX {
		d.Tiles[i][DMAXY-1] = caveCeiling
	}
}

// lockout reports whether every floor cell of the layout is reachable from the
// last one scanned (levels/drlg_l3.cpp:1904-1927).
func (d *Dungeon) lockout() bool {
	d.Mask.Reset()

	total := 0
	last := Point{}

	for j := range DMAXY {
		for i := range DMAXX {
			if d.Tiles[i][j] == 0 {
				continue
			}

			d.Mask.Set(i, j, true)
			last = Point{X: i, Y: j}
			total++
		}
	}

	return total == d.lockRectangle(last.X, last.Y)
}

// lockRectangle floods the layout mask from x,y and returns how many cells the
// flood reached (levels/drlg_l3.cpp:1890-1902).
//
// The four neighbours are walked in upstream's order; the bounds tests keep
// the recursion inside the mask because LockRectangle itself has none.
func (d *Dungeon) lockRectangle(x, y int) int {
	if x < 0 || x >= DMAXX || y < 0 || y >= DMAXY || !d.Mask.Test(x, y) {
		return 0
	}

	d.Mask.Set(x, y, false)

	count := 1
	count += d.lockRectangle(x+1, y)
	count += d.lockRectangle(x-1, y)
	count += d.lockRectangle(x, y+1)
	count += d.lockRectangle(x, y-1)

	return count
}
