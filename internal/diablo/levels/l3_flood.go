package levels

// This file ports the caves river: River() and the walk that finds it
// (levels/drlg_l3.cpp:1017-1266). A river starts on one of the four lava
// ceiling tiles and walks the dirt layer until it reaches the map border, then
// a bridge is carved into the middle of the path.

// River's limits (levels/drlg_l3.cpp:1021-1026, 1233).
const (
	caveRiverLen     = 100
	caveRiverTries   = 200
	caveRiverMax     = 4
	caveRiverMinBail = 7
)

// l3RiverStarts classifies the ceiling tile a river starts on: the outward
// direction, the direction back towards the level, and the first path tile
// (levels/drlg_l3.cpp:1061-1077). Tiles outside 25..28 park the walk, which is
// what the zero value means; riverStart never classifies one of those.
var l3RiverStarts = [256]riverStart{
	25: {dir: 3, nodir: 2, tile: 40},
	26: {dir: 0, nodir: 1, tile: 38},
	27: {dir: 1, nodir: 0, tile: 41},
	28: {dir: 2, nodir: 3, tile: 39},
}

// l3RiverCorners are River's eight turn cases: the corner tile written two
// steps back and the direction the path may not take next
// (levels/drlg_l3.cpp:1118-1168).
var l3RiverCorners = [...]struct {
	dir    int
	pdir   int
	tile   int
	nodir2 int
}{
	{dir: 0, pdir: 2, tile: 22, nodir2: 1},
	{dir: 3, pdir: 1, tile: 22, nodir2: 2},
	{dir: 0, pdir: 3, tile: 21, nodir2: 1},
	{dir: 2, pdir: 1, tile: 21, nodir2: 3},
	{dir: 1, pdir: 2, tile: 20, nodir2: 0},
	{dir: 3, pdir: 0, tile: 20, nodir2: 2},
	{dir: 1, pdir: 3, tile: 19, nodir2: 0},
	{dir: 2, pdir: 0, tile: 19, nodir2: 3},
}

// l3RiverReaches are River's four border checks: the wall face that must sit
// one step ahead of the path and the ceiling behind it, which is what turns the
// river into a pool inlet (levels/drlg_l3.cpp:1170-1229).
var l3RiverReaches = [...]riverReach{
	{dir: 0, front: 10, behind: 8, tile: 24, offset: Point{X: 0, Y: -1}},
	{dir: 1, front: 2, behind: 8, tile: 42, offset: Point{X: 0, Y: 1}},
	{dir: 2, front: 4, behind: 8, tile: 43, offset: Point{X: 1, Y: 0}},
	{dir: 3, front: 9, behind: 8, tile: 23, offset: Point{X: -1, Y: 0}},
}

// caveRiverWalk is the state River carries from one attempt to the next: the
// current direction, the direction the path came from and the direction it may
// not turn to. Upstream keeps all three in River's locals.
type caveRiverWalk struct {
	path  caveRiver
	dir   int
	pdir  int
	nodir int
}

// riverStart is one entry of l3RiverStarts.
type riverStart struct {
	dir   int
	nodir int
	tile  int
}

// riverReach is one entry of l3RiverReaches.
type riverReach struct {
	dir    int
	front  Tile
	behind Tile
	tile   int
	offset Point
}

// caveRiver is River's path buffer: the x, y and tile id of every step.
type caveRiver struct {
	x      [caveRiverLen]int
	y      [caveRiverLen]int
	tile   [caveRiverLen]int
	amount int
}

// begin restarts the path at start, keeping the start tile riverStart
// classified (levels/drlg_l3.cpp:1078-1081).
func (r *caveRiver) begin(start Point) {
	r.x[0] = start.X
	r.y[0] = start.Y
	r.amount = 1
}

// append adds one step to the path (levels/drlg_l3.cpp:1115-1117).
func (r *caveRiver) append(step Point, tile int) {
	r.x[r.amount] = step.X
	r.y[r.amount] = step.Y
	r.tile[r.amount] = tile
	r.amount++
}

// river finds the rivers of a caves level and writes them into the tile layer
// (levels/drlg_l3.cpp:1017-1266).
func (d *Dungeon) river() {
	walk := caveRiverWalk{pdir: -1}
	tries := 0
	rivers := 0

	for tries < caveRiverTries && rivers < caveRiverMax {
		stop, bridge := d.riverPass(&walk, &tries)
		if stop {
			return
		}

		if bridge {
			rivers++
		}
	}
}

// riverPass is one pass of River's outer loop: walk attempts until one path
// reaches the border, then carve its bridge. stop is upstream's `return`, which
// ends the whole river pass (levels/drlg_l3.cpp:1027-1262).
func (d *Dungeon) riverPass(w *caveRiverWalk, tries *int) (stop, bridge bool) {
	reached := false

	for !reached && *tries < caveRiverTries {
		*tries++

		giveUp := false
		reached, giveUp = d.riverAttempt(w)
		if giveUp {
			return true, false
		}
	}

	if !reached || w.path.amount < caveRiverMinBail {
		return false, false
	}

	return false, w.path.placeBridge(d)
}

// riverAttempt is one iteration of River's inner loop: pick a start, walk it.
// stop is upstream's `return` after a hundred failed starts.
func (d *Dungeon) riverAttempt(w *caveRiverWalk) (reached, stop bool) {
	start, found, stop := d.riverStart(w)
	if stop || !found {
		return false, stop
	}

	w.path.begin(start)

	return w.walkFrom(d, start), false
}

// riverStart picks the lava ceiling tile the path starts on
// (levels/drlg_l3.cpp:1030-1077). ok is false when the scan fell off the map,
// stop when it ran the whole scan budget.
func (d *Dungeon) riverStart(w *caveRiverWalk) (start Point, ok, stop bool) {
	p, scans := d.riverScan()

	switch {
	case p.Y >= DMAXY:
		return Point{}, false, false
	case scans >= caveRiverLen:
		return Point{}, false, true
	}

	startTile := l3RiverStarts[d.Tiles[p.X][p.Y]]
	w.dir, w.nodir, w.path.tile[0] = startTile.dir, startTile.nodir, startTile.tile

	return p, true, false
}

// riverScan draws random starts until one row holds a lava ceiling tile, then
// scans that row forward for it, returning the hit and the number of starts
// spent (levels/drlg_l3.cpp:1030-1059).
func (d *Dungeon) riverScan() (Point, int) {
	p := Point{}
	scans := 0

	for (p.Y >= DMAXY || !d.isRiverStartTile(p)) && scans < caveRiverLen {
		p = Point{X: int(d.Rnd.GenerateRnd(DMAXX)), Y: int(d.Rnd.GenerateRnd(DMAXY))}
		scans++

		p = d.scanRiverForward(p)
	}

	return p, scans
}

// scanRiverForward walks east along p's row until it finds a lava ceiling tile
// or runs off the row (levels/drlg_l3.cpp:1043-1054).
func (d *Dungeon) scanRiverForward(p Point) Point {
	for p.Y < DMAXY && !d.isRiverStartTile(p) {
		p.X++
		if p.X >= DMAXX {
			p.X = 0
			p.Y++
		}
	}

	return p
}

// isRiverStartTile reports whether p holds one of the four lava ceiling tiles a
// river starts on (levels/drlg_l3.cpp:1052).
func (d *Dungeon) isRiverStartTile(p Point) bool {
	return d.Tiles[p.X][p.Y] >= 25 && d.Tiles[p.X][p.Y] <= 28
}

// walkFrom traces the river path from start and reports whether it reached the
// map border (levels/drlg_l3.cpp:1080-1231).
func (w *caveRiverWalk) walkFrom(d *Dungeon, start Point) bool {
	pos := start
	dircheck := 0
	nodir2 := 4

	for dircheck < 4 && w.path.amount < caveRiverLen {
		previous := pos

		dircheck = w.turn(d, dircheck, nodir2)
		pos = stepRiver(w.dir, pos)

		if d.flatTile(pos.X, pos.Y) != caveFloor {
			pos = previous

			continue
		}

		dircheck = 0
		w.path.append(pos, d.riverStepTile(w.dir))
		nodir2 = w.noteCorner(nodir2)
	}

	return w.reachedBorder(d, pos)
}

// turn is River's direction pick: random on the first step of a run, then a
// right-hand turn that skips the two forbidden directions
// (levels/drlg_l3.cpp:1083-1093).
func (w *caveRiverWalk) turn(d *Dungeon, dircheck, nodir2 int) int {
	if dircheck == 0 {
		w.dir = int(d.Rnd.GenerateRnd(4))
	} else {
		w.dir = (w.dir + 1) & 3
	}

	dircheck++
	for w.dir == w.nodir || w.dir == nodir2 {
		w.dir = (w.dir + 1) & 3
		dircheck++
	}

	return dircheck
}

// noteCorner applies the turn case for the step just taken and returns the
// direction the next step may not take (levels/drlg_l3.cpp:1118-1168).
func (w *caveRiverWalk) noteCorner(nodir2 int) int {
	tile, next, ok := riverCorner(w.dir, w.pdir)
	if !ok {
		w.pdir = w.dir

		return nodir2
	}

	if w.path.amount > 2 {
		w.path.tile[w.path.amount-2] = tile
	}

	w.pdir = w.dir

	return next
}

// riverCorner looks up the turn case for a (direction, previous direction)
// pair.
func riverCorner(dir, pdir int) (tile, nodir2 int, ok bool) {
	for _, corner := range l3RiverCorners {
		if corner.dir == dir && corner.pdir == pdir {
			return corner.tile, corner.nodir2, true
		}
	}

	return 0, 0, false
}

// stepRiver moves one step in dir, straight onto the map edge if the step
// leaves the layer: the flat address the C array reads then lands in the
// neighbouring array (levels/drlg_l3.cpp:1095-1105).
func stepRiver(dir int, pos Point) Point {
	switch dir {
	case 0:
		pos.Y = max(pos.Y-1, 0)
	case 1:
		pos.Y = min(pos.Y+1, DMAXY)
	case 2:
		pos.X = min(pos.X+1, DMAXX)
	case 3:
		pos.X = max(pos.X-1, 0)
	default:
	}

	return pos
}

// riverStepTile is the tile a path step gets from its direction: a vertical
// river face for a north/south step, a horizontal one otherwise
// (levels/drlg_l3.cpp:1107-1113).
func (d *Dungeon) riverStepTile(dir int) int {
	if dir < 2 {
		return d.pickRandomlyAmong([2]int{17, 18})
	}

	return d.pickRandomlyAmong([2]int{15, 16})
}

// reachedBorder reports whether the path ended against a pool wall. On a hit it
// appends the inlet step past the path (not counted, like upstream) and retiles
// the last real step from the direction the path turned to
// (levels/drlg_l3.cpp:1170-1229).
func (w *caveRiverWalk) reachedBorder(d *Dungeon, pos Point) bool {
	for _, reach := range l3RiverReaches {
		if w.dir != reach.dir {
			continue
		}

		front, ok := d.riverInlet(reach, pos)
		if !ok {
			continue
		}

		w.path.append(front, reach.tile)
		w.retilePreviousStep()

		return true
	}

	return false
}

// riverInlet reports whether pos ends against reach's pool wall and returns the
// inlet step one further along (levels/drlg_l3.cpp:1181-1229).
func (d *Dungeon) riverInlet(reach riverReach, pos Point) (Point, bool) {
	front := pos.Add(reach.offset)
	behind := pos.Add(Point{X: 2 * reach.offset.X, Y: 2 * reach.offset.Y})

	if !inCaveArea(front) || !inCaveArea(behind) {
		return Point{}, false
	}

	if d.Tiles[front.X][front.Y] != reach.front || d.Tiles[behind.X][behind.Y] != reach.behind {
		return Point{}, false
	}

	return front, true
}

// inCaveArea reports whether a point is inside the raw mega-tile layer.
func inCaveArea(p Point) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < DMAXX && p.Y < DMAXY
}

// retilePreviousStep re-tiles the step before the inlet from the direction the
// path came from (levels/drlg_l3.cpp:1178-1225).
func (w *caveRiverWalk) retilePreviousStep() {
	if tile, _, ok := riverCorner(w.dir, w.pdir); ok {
		w.path.tile[w.path.amount-2] = tile
	}
}

// placeBridge finds a straight run of the path that has open floor on both
// sides, turns it into a bridge tile and writes the whole path into the layer.
// It reports whether a bridge was found
// (levels/drlg_l3.cpp:1235-1262).
func (r *caveRiver) placeBridge(d *Dungeon) bool {
	if !r.bridgeFound(d) {
		return false
	}

	for i := range r.amount {
		d.Tiles[r.x[i]][r.y[i]] = Tile(r.tile[i])
	}

	return true
}

// bridgeFound scans random steps of the path for one that can carry a bridge,
// as upstream's thirty-try loop does (levels/drlg_l3.cpp:1237-1257).
func (r *caveRiver) bridgeFound(d *Dungeon) bool {
	for range 30 {
		bridge := int(d.Rnd.GenerateRnd(int32(r.amount - 1)))

		found := r.bridgeAt(d, bridge)
		if found == 0 {
			continue
		}

		r.tile[bridge] = found

		return true
	}

	return false
}

// bridgeAt reports which bridge tile fits at path step bridge: 44 for a
// north/south bridge, 45 for an east/west one, 0 for none
// (levels/drlg_l3.cpp:1240-1257).
func (r *caveRiver) bridgeAt(d *Dungeon, bridge int) int {
	x, y := r.x[bridge], r.y[bridge]

	switch r.tile[bridge] {
	case 15, 16:
		if d.bridgeRunsAlong(x, y, Point{Y: 1}) && !r.pathCrosses(bridge, Point{Y: 1}) {
			return 44
		}
	case 17, 18:
		if d.bridgeRunsAlong(x, y, Point{X: 1}) && !r.pathCrosses(bridge, Point{X: 1}) {
			return 45
		}
	default:
	}

	return 0
}

// bridgeRunsAlong reports whether the floor tile sits open on both sides of
// (x, y) along step (levels/drlg_l3.cpp:1242, 1250).
func (d *Dungeon) bridgeRunsAlong(x, y int, step Point) bool {
	return d.flatTile(x-step.X, y-step.Y) == caveFloor && d.flatTile(x+step.X, y+step.Y) == caveFloor
}

// pathCrosses reports whether another path step sits one step either way from
// bridge, which would put the bridge across the river rather than along it
// (levels/drlg_l3.cpp:1248-1261).
func (r *caveRiver) pathCrosses(bridge int, step Point) bool {
	for i := range r.amount {
		if r.x[i] == r.x[bridge]+step.X && r.y[i] == r.y[bridge]+step.Y {
			return true
		}

		if r.x[i] == r.x[bridge]-step.X && r.y[i] == r.y[bridge]-step.Y {
			return true
		}
	}

	return false
}
