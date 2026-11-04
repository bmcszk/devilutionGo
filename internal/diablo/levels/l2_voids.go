package levels

// This file ports the catacombs void filler: the passes that grow room-sized
// holes out of the empty layer FillVoids finds, and the two passes that turn a
// grown rectangle into a room (levels/drlg_l2.cpp:2165-2443, 1984-2051).

// voidRect is the rectangle FillVoid starts from: the wall cell plus one step
// into every empty side (levels/drlg_l2.cpp:2165-2180).
type voidRect struct {
	x1, y1 int
	x2, y2 int
}

// voidEdge is one growing side of a void run: where it currently is and
// whether it is still moving (levels/drlg_l2.cpp:2184-2203).
type voidEdge struct {
	pos  int
	open bool
}

// voidGrowth is the four fill flags FillVoid takes as arguments upstream: the
// sides the wall cell has empty space on (levels/drlg_l2.cpp:2165).
type voidGrowth struct {
	xf1, yf1, xf2, yf2 bool
}

// opensAnySide reports whether FillVoids found any side to grow into.
func (g voidGrowth) opensAnySide() bool {
	return g.xf1 || g.yf1 || g.xf2 || g.yf2
}

// voidRectAt is FillVoid's opening step: it opens the rect one cell into each
// side that has empty space (levels/drlg_l2.cpp:2166-2180).
func voidRectAt(g voidGrowth, xx, yy int) voidRect {
	r := voidRect{x1: xx, y1: yy, x2: xx, y2: yy}

	if g.xf1 {
		r.x1--
	}
	if g.xf2 {
		r.x2++
	}
	if g.yf1 {
		r.y1--
	}
	if g.yf2 {
		r.y2++
	}

	return r
}

// fillVoid grows one room into the empty space around a wall cell, widening
// the longest free run first. The four branches are upstream's if/else-if
// chain: each returns, so the first open side wins
// (levels/drlg_l2.cpp:2165-2384).
func (c *catacombs) fillVoid(g voidGrowth, xx, yy int) {
	r := voidRectAt(g, xx, yy)

	if !g.xf1 {
		c.fillVoidFromRight(r, g)

		return
	}
	if !g.xf2 {
		c.fillVoidFromLeft(r, g)

		return
	}
	if !g.yf1 {
		c.fillVoidFromBelow(r, g)

		return
	}
	if !g.yf2 {
		c.fillVoidFromAbove(r, g)
	}
}

// fillVoidFromRight crawls the free rows beside the wall cell, then widens the
// run to the right (levels/drlg_l2.cpp:2181-2229).
func (c *catacombs) fillVoidFromRight(r voidRect, g voidGrowth) {
	y1, y2 := c.growRows(r.x2, r.y1, r.y2, g)
	if y2-y1 <= 5 {
		return
	}

	x2 := c.growColsRight(r.x1, r.x2, y1, y2, g)
	c.paintVoidRoom(r.x1, x2, y1, y2)
}

// fillVoidFromLeft crawls the free rows beside the wall cell, then widens the
// run to the left (levels/drlg_l2.cpp:2239-2286).
func (c *catacombs) fillVoidFromLeft(r voidRect, g voidGrowth) {
	y1, y2 := c.growRows(r.x1, r.y1, r.y2, g)
	if y2-y1 <= 5 {
		return
	}

	x1 := c.growColsLeft(r.x1, r.x2, y1, y2, g)
	c.paintVoidRoom(x1, r.x2, y1, y2)
}

// fillVoidFromBelow crawls the free columns beside the wall cell, then lowers
// the run (levels/drlg_l2.cpp:2276-2329).
func (c *catacombs) fillVoidFromBelow(r voidRect, g voidGrowth) {
	x1, x2 := c.growColumns(r.y2, r.x1, r.x2, g)
	if x2-x1 <= 5 {
		return
	}

	y2 := c.growRowsDown(x1, x2, r.y1, r.y2, g)
	c.paintVoidRoom(x1, x2, r.y1, y2)
}

// fillVoidFromAbove crawls the free columns beside the wall cell, then raises
// the run (levels/drlg_l2.cpp:2319-2372).
func (c *catacombs) fillVoidFromAbove(r voidRect, g voidGrowth) {
	x1, x2 := c.growColumns(r.y1, r.x1, r.x2, g)
	if x2-x1 <= 5 {
		return
	}

	y1 := c.growRowsUp(x1, x2, r.y1, r.y2, g)
	c.paintVoidRoom(x1, x2, y1, r.y2)
}

// paintVoidRoom turns one grown rectangle into a room once both axes are wide
// enough, then opens the walls its floor now spans
// (levels/drlg_l2.cpp:2165-2384).
func (c *catacombs) paintVoidRoom(x1, x2, y1, y2 int) {
	if x2-x1 <= 5 || y2-y1 <= 5 {
		return
	}

	c.mapRoom(x1, y1, x2, y2)
	c.knockWalls(x1, y1, x2, y2)
}

// growRows widens the void along y while the column at x stays empty
// (levels/drlg_l2.cpp:2182-2203, 2240-2261). The 14-cell cap is upstream's.
func (c *catacombs) growRows(x, y1, y2 int, g voidGrowth) (rowMin, rowMax int) {
	up := voidEdge{pos: y1, open: g.yf1}
	down := voidEdge{pos: y2, open: g.yf2}

	for (up.open || down.open) && down.pos-up.pos < 14 {
		up = c.extendRowUp(x, up)
		down = c.extendRowDown(x, down)
	}

	return up.pos + 2, down.pos - 2
}

// growColumns widens the void along x over the fixed row
// (levels/drlg_l2.cpp:2277-2303, 2320-2346).
func (c *catacombs) growColumns(row, x1, x2 int, g voidGrowth) (colMin, colMax int) {
	left := voidEdge{pos: x1, open: g.xf1}
	right := voidEdge{pos: x2, open: g.xf2}

	for (left.open || right.open) && right.pos-left.pos < 14 {
		left = c.extendColLeft(row, left)
		right = c.extendColRight(row, right)
	}

	return left.pos + 2, right.pos - 2
}

// extendRowUp moves a void's open edge one cell up and stops when that cell is
// taken (levels/drlg_l2.cpp:2196-2203).
func (c *catacombs) extendRowUp(x int, e voidEdge) voidEdge {
	if !e.open || e.pos == 0 {
		return voidEdge{pos: e.pos}
	}

	pos := e.pos - 1

	return voidEdge{pos: pos, open: c.cell(Point{X: x, Y: pos}) == cellEmpty}
}

// extendRowDown moves a void's open edge one cell down and stops when that cell
// is taken (levels/drlg_l2.cpp:2196-2203).
func (c *catacombs) extendRowDown(x int, e voidEdge) voidEdge {
	if !e.open || e.pos == DMAXY-1 {
		return voidEdge{pos: e.pos}
	}

	pos := e.pos + 1

	return voidEdge{pos: pos, open: c.cell(Point{X: x, Y: pos}) == cellEmpty}
}

// extendColLeft moves a void's open edge one cell left and stops when that cell
// is taken (levels/drlg_l2.cpp:2290-2297).
func (c *catacombs) extendColLeft(row int, e voidEdge) voidEdge {
	if !e.open || e.pos == 0 {
		return voidEdge{pos: e.pos}
	}

	pos := e.pos - 1

	return voidEdge{pos: pos, open: c.cell(Point{X: pos, Y: row}) == cellEmpty}
}

// extendColRight moves a void's open edge one cell right and stops when that
// cell is taken (levels/drlg_l2.cpp:2290-2297).
func (c *catacombs) extendColRight(row int, e voidEdge) voidEdge {
	if !e.open || e.pos == DMAXX-1 {
		return voidEdge{pos: e.pos}
	}

	pos := e.pos + 1

	return voidEdge{pos: pos, open: c.cell(Point{X: pos, Y: row}) == cellEmpty}
}

// voidRowClear reports whether every layout cell of a row run is empty.
func (c *catacombs) voidRowClear(x1, x2, y int) bool {
	for ii := x1; ii <= x2; ii++ {
		if c.cell(Point{X: ii, Y: y}) != cellEmpty {
			return false
		}
	}

	return true
}

// voidColumnClear reports whether every layout cell of a column run is empty.
func (c *catacombs) voidColumnClear(y1, y2, x int) bool {
	for jj := y1; jj <= y2; jj++ {
		if c.cell(Point{X: x, Y: jj}) != cellEmpty {
			return false
		}
	}

	return true
}

// growColsRight extends a void to the right while its rows stay empty
// (levels/drlg_l2.cpp:2205-2229).
func (c *catacombs) growColsRight(x1, x2, y1, y2 int, g voidGrowth) int {
	if !g.xf2 {
		return x2 - 2
	}

	for x2 != DMAXX-1 && x2-x1 < 12 && c.voidColumnClear(y1, y2, x2) {
		x2++
	}

	return x2 - 2
}

// growColsLeft extends a void to the left while its rows stay empty
// (levels/drlg_l2.cpp:2263-2286).
func (c *catacombs) growColsLeft(x1, x2, y1, y2 int, g voidGrowth) int {
	if !g.xf1 {
		return x1 + 2
	}

	for x1 != 0 && x2-x1 < 12 && c.voidColumnClear(y1, y2, x1) {
		x1--
	}

	return x1 + 2
}

// growRowsDown lowers a void's floor while the columns stay empty
// (levels/drlg_l2.cpp:2306-2329).
func (c *catacombs) growRowsDown(x1, x2, y1, y2 int, g voidGrowth) int {
	if !g.yf2 {
		return y2 - 2
	}

	for y2 != DMAXY-1 && y2-y1 < 12 && c.voidRowClear(x1, x2, y2) {
		y2++
	}

	return y2 - 2
}

// growRowsUp raises a void's ceiling while the columns stay empty
// (levels/drlg_l2.cpp:2349-2372).
func (c *catacombs) growRowsUp(x1, x2, y1, y2 int, g voidGrowth) int {
	if !g.yf1 {
		return y1 + 2
	}

	for y1 != 0 && y2-y1 < 12 && c.voidRowClear(x1, x2, y1) {
		y1--
	}

	return y1 + 2
}

// fillVoids grows rooms into the leftover void until the open space is small
// enough, at most 100 wall probes (levels/drlg_l2.cpp:2386-2443).
func (c *catacombs) fillVoids() bool {
	for to := 0; c.countEmptyTiles() > 700 && to < 100; {
		xx := int(c.d.Rnd.GenerateRnd(38)) + 1
		yy := int(c.d.Rnd.GenerateRnd(38)) + 1

		if c.cells[xx][yy] != cellWall {
			continue
		}

		if g := c.voidShape(xx, yy); g.opensAnySide() {
			c.fillVoid(g, xx, yy)
		}

		to++
	}

	return c.countEmptyTiles() <= 700
}

// voidShape classifies the free space around one wall cell into the four
// growth directions (levels/drlg_l2.cpp:2397-2437).
func (c *catacombs) voidShape(xx, yy int) voidGrowth {
	if c.voidGapOnLeft(xx, yy) {
		return c.leftGapShape(xx, yy)
	}
	if c.voidGapOnRight(xx, yy) {
		return c.rightGapShape(xx, yy)
	}
	if c.voidGapAbove(xx, yy) {
		return c.aboveGapShape(xx, yy)
	}
	if c.voidGapBelow(xx, yy) {
		return c.belowGapShape(xx, yy)
	}

	return voidGrowth{}
}

// voidGapOnLeft is FillVoid's first outer test: empty to the left, room to the
// right (levels/drlg_l2.cpp:2399).
func (c *catacombs) voidGapOnLeft(xx, yy int) bool {
	return c.layoutAt(Point{X: xx - 1, Y: yy}) == cellEmpty &&
		c.layoutAt(Point{X: xx + 1, Y: yy}) == cellRoom
}

// voidGapOnRight is FillVoid's second outer test (levels/drlg_l2.cpp:2408).
func (c *catacombs) voidGapOnRight(xx, yy int) bool {
	return c.layoutAt(Point{X: xx + 1, Y: yy}) == cellEmpty &&
		c.layoutAt(Point{X: xx - 1, Y: yy}) == cellRoom
}

// voidGapAbove is FillVoid's third outer test (levels/drlg_l2.cpp:2417).
func (c *catacombs) voidGapAbove(xx, yy int) bool {
	return c.layoutAt(Point{X: xx, Y: yy - 1}) == cellEmpty &&
		c.layoutAt(Point{X: xx, Y: yy + 1}) == cellRoom
}

// voidGapBelow is FillVoid's fourth outer test (levels/drlg_l2.cpp:2426).
func (c *catacombs) voidGapBelow(xx, yy int) bool {
	return c.layoutAt(Point{X: xx, Y: yy + 1}) == cellEmpty &&
		c.layoutAt(Point{X: xx, Y: yy - 1}) == cellRoom
}

// leftGapShape is FillVoid's first shape: the cell opens to the left and the
// corners confirm the gap runs up and down (levels/drlg_l2.cpp:2399-2406).
func (c *catacombs) leftGapShape(xx, yy int) voidGrowth {
	if c.layoutAt(Point{X: xx + 1, Y: yy - 1}) != cellRoom ||
		c.layoutAt(Point{X: xx + 1, Y: yy + 1}) != cellRoom ||
		c.layoutAt(Point{X: xx - 1, Y: yy - 1}) != cellEmpty ||
		c.layoutAt(Point{X: xx - 1, Y: yy + 1}) != cellEmpty {
		return voidGrowth{}
	}

	return voidGrowth{xf1: true, yf1: true, yf2: true}
}

// rightGapShape is FillVoid's second shape: the cell opens to the right
// (levels/drlg_l2.cpp:2408-2415).
func (c *catacombs) rightGapShape(xx, yy int) voidGrowth {
	if c.layoutAt(Point{X: xx - 1, Y: yy - 1}) != cellRoom ||
		c.layoutAt(Point{X: xx - 1, Y: yy + 1}) != cellRoom ||
		c.layoutAt(Point{X: xx + 1, Y: yy - 1}) != cellEmpty ||
		c.layoutAt(Point{X: xx + 1, Y: yy + 1}) != cellEmpty {
		return voidGrowth{}
	}

	return voidGrowth{xf2: true, yf1: true, yf2: true}
}

// aboveGapShape is FillVoid's third shape: the cell opens upward
// (levels/drlg_l2.cpp:2417-2424).
func (c *catacombs) aboveGapShape(xx, yy int) voidGrowth {
	if c.layoutAt(Point{X: xx - 1, Y: yy + 1}) != cellRoom ||
		c.layoutAt(Point{X: xx + 1, Y: yy + 1}) != cellRoom ||
		c.layoutAt(Point{X: xx - 1, Y: yy - 1}) != cellEmpty ||
		c.layoutAt(Point{X: xx + 1, Y: yy - 1}) != cellEmpty {
		return voidGrowth{}
	}

	return voidGrowth{xf1: true, xf2: true, yf1: true}
}

// belowGapShape is FillVoid's fourth shape: the cell opens downward
// (levels/drlg_l2.cpp:2426-2433).
func (c *catacombs) belowGapShape(xx, yy int) voidGrowth {
	if c.layoutAt(Point{X: xx - 1, Y: yy - 1}) != cellRoom ||
		c.layoutAt(Point{X: xx + 1, Y: yy - 1}) != cellRoom ||
		c.layoutAt(Point{X: xx - 1, Y: yy + 1}) != cellEmpty ||
		c.layoutAt(Point{X: xx + 1, Y: yy + 1}) != cellEmpty {
		return voidGrowth{}
	}

	return voidGrowth{xf1: true, xf2: true, yf2: true}
}

// layoutAt reads the layout layer at a bare coordinate pair.
func (c *catacombs) layoutAt(p Point) layoutCell {
	return c.cell(p)
}

// build is CreateDungeon: it draws the room tree, digs the halls, closes the
// leftover quadrants and turns the layout into tiles
// (levels/drlg_l2.cpp:2445-2508).
