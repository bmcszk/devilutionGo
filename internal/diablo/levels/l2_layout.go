package levels

// This file ports the catacombs layout builder: the ASCII room/hall layer
// CreateDungeon fills before any tile exists, and the passes that turn it into
// tiles (levels/drlg_l2.cpp:1685-2508).

// Hall directions, indexed the way DirAdd is (levels/drlg_l2.cpp:33-39, 53-59).
const (
	hallNone hallDirection = iota
	hallUp
	hallRight
	hallDown
	hallLeft
)

// Room grid limits (levels/drlg_l2.cpp:1773-1786).
const (
	// catacombAreaMin is CreateRoom's AreaMin: the smallest rectangle it splits.
	catacombAreaMin = 2
	// catacombRoomMax and catacombRoomMin bound a drawn room (RoomMax, RoomMin).
	catacombRoomMax = 10
	catacombRoomMin = 4
	// catacombRoomListMax is the RoomList length guard (`nRoomCnt >= 80`).
	catacombRoomListMax = 80
)

// Hall border clamps: ConnectHall turns a hall round when it reaches these
// coordinates (levels/drlg_l2.cpp:1886-1890).
const (
	catacombHallEdge   = 38
	catacombHallOrigin = 1
)

// quadrantTurns is TurnAtQuadrant's table: each room corner turns a hall
// heading one of two ways away from it (levels/drlg_l2.cpp:1896-1903).
var quadrantTurns = map[layoutCell]struct {
	against [2]hallDirection
	to      hallDirection
}{
	cellQuadC: {against: [2]hallDirection{hallUp, hallLeft}, to: hallRight},
	cellQuadB: {against: [2]hallDirection{hallUp, hallRight}, to: hallDown},
	cellQuadE: {against: [2]hallDirection{hallLeft, hallDown}, to: hallUp},
	cellQuadA: {against: [2]hallDirection{hallRight, hallDown}, to: hallLeft},
}

// patternCellAllowed is DoPatternCheck's code table: each pattern code lists the
// layout cells it accepts; code 0 is the wildcard
// (levels/drlg_l2.cpp:1996-2046).
var patternCellAllowed = map[Tile][]layoutCell{
	1: {cellWall},
	2: {cellRoom},
	3: {cellDoor},
	4: {cellEmpty},
	5: {cellDoor, cellRoom},
	6: {cellDoor, cellWall},
	7: {cellEmpty, cellRoom},
	8: {cellDoor, cellWall, cellRoom},
}

// ringOffsets are the eight neighbours of a cell, in row-major order.
var ringOffsets = [8]Point{
	{X: -1, Y: -1},
	{Y: -1},
	{X: 1, Y: -1},
	{X: -1},
	{X: 1},
	{X: -1, Y: 1},
	{Y: 1},
	{X: 1, Y: 1},
}

// hallSides selects the optional wide-hall tiles ConnectHall asks for: minus is
// the left/top side of the walked direction, plus the right/bottom side
// (levels/drlg_l2.cpp:1912-1922).
type hallSides struct{ minus, plus bool }

// hallDirection is HallDirection: the axis a hall runs along.
type hallDirection int8

// hallNode is HallNode, one pending connection between two rooms
// (levels/drlg_l2.cpp:41-45).
type hallNode struct {
	beginning Point
	end       Point
	direction hallDirection
}

// roomNode is RoomNode, one drawn room rectangle (levels/drlg_l2.cpp:47-50).
type roomNode struct {
	topLeft     Point
	bottomRight Point
}

// catacombs is the catacombs generator's layout state: predungeon, the room
// list, the pending hall list and the room counter
// (levels/drlg_l2.cpp:52-60).
type catacombs struct {
	d       *Dungeon
	cells   [DMAXX][DMAXY]layoutCell
	roomCnt int
	rooms   [catacombRoomListMax + 1]roomNode
	halls   []hallNode
}

// newCatacombs is InitDungeonFlags: it clears the protected mask and fills the
// layout layer with empty cells (levels/drlg_l2.cpp:1685-1689).
func newCatacombs(d *Dungeon) *catacombs {
	d.Protected.Reset()

	c := &catacombs{d: d}
	for i := range DMAXX {
		for j := range DMAXY {
			c.cells[i][j] = cellEmpty
		}
	}

	return c
}

// cell reads the layout layer. Cells outside the layer read as empty: upstream
// indexes a global `char predungeon[DMAXX][DMAXY]` whose neighbours are the
// zeroed flag storage that follows it (levels/dun_tile_data.cpp:36-38), and no
// reachable call site depends on those values.
func (c *catacombs) cell(p Point) layoutCell {
	if p.X < 0 || p.Y < 0 || p.X >= DMAXX || p.Y >= DMAXY {
		return cellEmpty
	}

	return c.cells[p.X][p.Y]
}

// setCell writes the layout layer, ignoring positions outside it.
func (c *catacombs) setCell(p Point, value layoutCell) {
	if p.X < 0 || p.Y < 0 || p.X >= DMAXX || p.Y >= DMAXY {
		return
	}

	c.cells[p.X][p.Y] = value
}

// placeHallExt marks one extra hall tile beside a hall
// (levels/drlg_l2.cpp:1912-1916, PlaceHallExt).
func (c *catacombs) placeHallExt(p Point) {
	if c.cell(p) == cellEmpty {
		c.setCell(p, cellHall)
	}
}

// createDoorType turns one hall tile into a door unless a door or a room
// quadrant is already there (levels/drlg_l2.cpp:1906-1910, CreateDoorType).
func (c *catacombs) createDoorType(p Point) {
	if c.cell(Point{X: p.X - 1, Y: p.Y}) == cellDoor {
		return
	}
	if c.cell(Point{X: p.X + 1, Y: p.Y}) == cellDoor {
		return
	}
	if c.cell(Point{X: p.X, Y: p.Y - 1}) == cellDoor {
		return
	}
	if c.cell(Point{X: p.X, Y: p.Y + 1}) == cellDoor {
		return
	}
	if isQuadrant(c.cell(p)) {
		return
	}

	c.setCell(p, cellDoor)
}

// isQuadrant reports whether v is one of the four room corner markers
// (levels/drlg_l2.cpp:1708-1713).
func isQuadrant(v layoutCell) bool {
	return v == cellQuadA || v == cellQuadB || v == cellQuadC || v == cellQuadE
}

// mapRoom paints one room rectangle: interior floor, walls on the border
// (levels/drlg_l2.cpp:1691-1706).
func (c *catacombs) mapRoom(x1, y1, x2, y2 int) {
	for jj := y1; jj <= y2; jj++ {
		for ii := x1; ii <= x2; ii++ {
			c.setCell(Point{X: ii, Y: jj}, cellRoom)
		}
	}

	for jj := y1; jj <= y2; jj++ {
		c.setCell(Point{X: x1, Y: jj}, cellWall)
		c.setCell(Point{X: x2, Y: jj}, cellWall)
	}

	for ii := x1; ii <= x2; ii++ {
		c.setCell(Point{X: ii, Y: y1}, cellWall)
		c.setCell(Point{X: ii, Y: y2}, cellWall)
	}
}

// defineRoom records one drawn room, marks its corner quadrants and paints
// the rectangle's border and interior (levels/drlg_l2.cpp:1708-1739).
func (c *catacombs) defineRoom(topLeft, bottomRight Point) {
	c.setCell(topLeft, cellQuadC)
	c.setCell(Point{X: topLeft.X, Y: bottomRight.Y}, cellQuadE)
	c.setCell(Point{X: bottomRight.X, Y: topLeft.Y}, cellQuadB)
	c.setCell(Point{X: bottomRight.X, Y: bottomRight.Y}, cellQuadA)

	c.roomCnt++
	c.rooms[c.roomCnt] = roomNode{topLeft: topLeft, bottomRight: bottomRight}

	for i := topLeft.X + 1; i <= bottomRight.X-1; i++ {
		c.setCell(Point{X: i, Y: topLeft.Y}, cellWall)
		c.setCell(Point{X: i, Y: bottomRight.Y}, cellWall)
	}

	bottomRight.Y--
	for j := topLeft.Y + 1; j <= bottomRight.Y; j++ {
		c.setCell(Point{X: topLeft.X, Y: j}, cellWall)
		c.setCell(Point{X: bottomRight.X, Y: j}, cellWall)

		for i := topLeft.X + 1; i < bottomRight.X; i++ {
			c.setCell(Point{X: i, Y: j}, cellRoom)
		}
	}
}

// protectSetPieceRoom is DefineRoom's forceHW branch, which only the
// quest-sized set-piece room takes: it marks the room's top row protected so
// FixLockout leaves it alone (levels/drlg_l2.cpp:1725-1731).
func (c *catacombs) protectSetPieceRoom(topLeft, bottomRight Point) {
	// BUGFIX: upstream fixes this, the original looped j instead of i.
	for i := topLeft.X; i < bottomRight.X; i++ {
		for i < bottomRight.Y {
			c.d.Protected.Set(i, topLeft.Y, true)
			i++
		}
	}
}

// createRoom draws a random room rectangle inside the passed area and recurses
// into the four remainders (levels/drlg_l2.cpp:1771-1867).
func (c *catacombs) createRoom(topLeft, bottomRight Point, rDest int, hDir hallDirection, size *Size) {
	if c.roomCnt >= catacombRoomListMax || topLeft.X+catacombAreaMin > bottomRight.X ||
		topLeft.Y+catacombAreaMin > bottomRight.Y {
		return
	}

	area := Size{Width: bottomRight.X - topLeft.X, Height: bottomRight.Y - topLeft.Y}

	roomSize := area
	if area.Width > catacombRoomMin {
		roomSize.Width = int(c.d.Rnd.GenerateRnd(int32(min(area.Width, catacombRoomMax)-catacombRoomMin))) +
			catacombRoomMin
	}
	if area.Height > catacombRoomMin {
		roomSize.Height = int(c.d.Rnd.GenerateRnd(int32(min(area.Height, catacombRoomMax)-catacombRoomMin))) +
			catacombRoomMin
	}

	if size != nil {
		roomSize = *size
	}

	roomTopLeft, roomBottomRight := c.placeRoom(topLeft, bottomRight, roomSize)
	c.defineRoom(roomTopLeft, roomBottomRight)
	if size != nil {
		c.protectSetPieceRoom(roomTopLeft, roomBottomRight)
	}

	c.connectToParent(roomTopLeft, roomBottomRight, roomSize, rDest, hDir, size)
	c.splitArea(topLeft, bottomRight, roomTopLeft, roomBottomRight, roomSize, c.roomCnt)
}

// placeRoom draws the room's origin inside one area and clamps both corners
// to the level bounds (levels/drlg_l2.cpp:1791-1816).
func (c *catacombs) placeRoom(topLeft, bottomRight Point, roomSize Size) (roomTopLeft, roomBottomRight Point) {
	area := Point{X: bottomRight.X - topLeft.X, Y: bottomRight.Y - topLeft.Y}

	randomWidth := c.d.Rnd.GenerateRnd(int32(area.X))
	randomHeight := c.d.Rnd.GenerateRnd(int32(area.Y))

	roomTopLeft = Point{X: topLeft.X + int(randomWidth), Y: topLeft.Y + int(randomHeight)}
	roomBottomRight = Point{X: roomTopLeft.X + roomSize.Width, Y: roomTopLeft.Y + roomSize.Height}

	if roomBottomRight.X > bottomRight.X {
		roomBottomRight.X = bottomRight.X
		roomTopLeft.X = bottomRight.X - roomSize.Width
	}
	if roomBottomRight.Y > bottomRight.Y {
		roomBottomRight.Y = bottomRight.Y
		roomTopLeft.Y = bottomRight.Y - roomSize.Height
	}

	roomTopLeft.X = clampCoord(roomTopLeft.X)
	roomTopLeft.Y = clampCoord(roomTopLeft.Y)
	roomBottomRight.X = clampCoord(roomBottomRight.X)
	roomBottomRight.Y = clampCoord(roomBottomRight.Y)

	return roomTopLeft, roomBottomRight
}

// clampCoord is std::clamp<WorldTileCoord>(v, 1, 38) (levels/drlg_l2.cpp:1813-1816).
func clampCoord(v int) int {
	return min(max(v, 1), DMAXX-2)
}

// connectToParent records the set-piece exclusion rectangle and queues the
// hall that joins this room to its parent (levels/drlg_l2.cpp:1820-1852).
func (c *catacombs) connectToParent(
	roomTopLeft, roomBottomRight Point, roomSize Size, rDest int, hDir hallDirection, size *Size,
) {
	standoff := Point{X: 2, Y: 2}

	if size != nil {
		c.d.SetPieceRoom = Rectangle{
			Position: Point{X: roomTopLeft.X + standoff.X, Y: roomTopLeft.Y + standoff.Y},
			Size:     Size{Width: roomSize.Width - 1, Height: roomSize.Height - 1},
		}
	}

	if rDest == 0 {
		return
	}

	c.halls = append(c.halls, c.hallToParent(roomTopLeft, roomBottomRight, roomSize, rDest, hDir))
}

// hallToParent picks the two wall openings one hall connects
// (levels/drlg_l2.cpp:1824-1851).
func (c *catacombs) hallToParent(
	roomTopLeft, roomBottomRight Point, roomSize Size, rDest int, hDir hallDirection,
) hallNode {
	parent := c.rooms[rDest]
	node := hallNode{direction: hDir}

	switch hDir {
	case hallUp:
		node.beginning = Point{
			X: int(c.d.Rnd.GenerateRnd(int32(roomSize.Width-2))) + roomTopLeft.X + 1,
			Y: roomTopLeft.Y,
		}
		node.end = Point{
			X: int(c.d.Rnd.GenerateRnd(int32(parent.bottomRight.X-parent.topLeft.X-2))) + parent.topLeft.X + 1,
			Y: parent.bottomRight.Y,
		}
	case hallDown:
		node.beginning = Point{
			X: int(c.d.Rnd.GenerateRnd(int32(roomSize.Width-2))) + roomTopLeft.X + 1,
			Y: roomBottomRight.Y,
		}
		node.end = Point{
			X: int(c.d.Rnd.GenerateRnd(int32(parent.bottomRight.X-parent.topLeft.X-2))) + parent.topLeft.X + 1,
			Y: parent.topLeft.Y,
		}
	case hallRight:
		node.beginning = Point{
			X: roomBottomRight.X,
			Y: int(c.d.Rnd.GenerateRnd(int32(roomSize.Height-2))) + roomTopLeft.Y + 1,
		}
		node.end = Point{
			X: parent.topLeft.X,
			Y: int(c.d.Rnd.GenerateRnd(int32(parent.bottomRight.Y-parent.topLeft.Y-2))) + parent.topLeft.Y + 1,
		}
	case hallLeft:
		node.beginning = Point{
			X: roomTopLeft.X,
			Y: int(c.d.Rnd.GenerateRnd(int32(roomSize.Height-2))) + roomTopLeft.Y + 1,
		}
		node.end = Point{
			X: parent.bottomRight.X,
			Y: int(c.d.Rnd.GenerateRnd(int32(parent.bottomRight.Y-parent.topLeft.Y-2))) + parent.topLeft.Y + 1,
		}
	default:
		// rDest is non-zero here, so upstream always matches one of the four
		// ordinals; hallNone keeps the zero node (levels/drlg_l2.cpp:1824-1851).
	}

	return node
}

// splitArea recurses into the four remainders around the new room, splitting
// the longer axis first (levels/drlg_l2.cpp:1854-1867).
func (c *catacombs) splitArea(topLeft, bottomRight, roomTopLeft, roomBottomRight Point, roomSize Size, rID int) {
	standoff := Point{X: 2, Y: 2}

	roomBottomLeft := Point{X: roomTopLeft.X, Y: roomBottomRight.Y}
	roomTopRight := Point{X: roomBottomRight.X, Y: roomTopLeft.Y}

	if roomSize.Height > roomSize.Width {
		c.createRoom(addPoint(topLeft, standoff), subPoint(roomBottomLeft, standoff), rID, hallRight, nil)
		c.createRoom(addPoint(roomTopRight, standoff), subPoint(bottomRight, standoff), rID, hallLeft, nil)
		c.createRoom(
			addPoint(Point{X: topLeft.X, Y: roomBottomRight.Y}, standoff),
			subPoint(Point{X: roomBottomRight.X, Y: bottomRight.Y}, standoff), rID, hallUp, nil,
		)
		c.createRoom(
			addPoint(Point{X: roomTopLeft.X, Y: topLeft.Y}, standoff),
			subPoint(Point{X: bottomRight.X, Y: roomTopLeft.Y}, standoff), rID, hallDown, nil,
		)

		return
	}

	c.createRoom(addPoint(topLeft, standoff), subPoint(roomTopRight, standoff), rID, hallDown, nil)
	c.createRoom(addPoint(roomBottomLeft, standoff), subPoint(bottomRight, standoff), rID, hallUp, nil)
	c.createRoom(
		addPoint(Point{X: topLeft.X, Y: roomTopLeft.Y}, standoff),
		subPoint(Point{X: roomTopLeft.X, Y: bottomRight.Y}, standoff), rID, hallRight, nil,
	)
	c.createRoom(
		addPoint(Point{X: roomBottomRight.X, Y: topLeft.Y}, standoff),
		subPoint(Point{X: bottomRight.X, Y: roomBottomRight.Y}, standoff), rID, hallLeft, nil,
	)
}

// addPoint is WorldTilePosition + WorldTileDisplacement.
func addPoint(p, d Point) Point {
	return Point{X: p.X + d.X, Y: p.Y + d.Y}
}

// subPoint is WorldTilePosition - WorldTileDisplacement.
func subPoint(p, d Point) Point {
	return Point{X: p.X - d.X, Y: p.Y - d.Y}
}

// connectHall walks a hall from one room opening to the other, turning at walls
// and stamping doors where it enters and leaves a room
// (levels/drlg_l2.cpp:1869-1982).
func (c *catacombs) connectHall(node hallNode) {
	fMinusFlag := c.d.Rnd.GenerateRnd(100) < 50
	fPlusFlag := c.d.Rnd.GenerateRnd(100) < 50

	beginning := node.beginning
	end := node.end

	c.createDoorType(beginning)
	c.createDoorType(end)

	direction := node.direction
	end = subPoint(end, dirAdd[direction])
	c.setCell(end, cellHall)

	fInRoom := false
	for {
		direction = bounce(beginning, direction)
		direction = c.turnAtQuadrant(beginning, direction)
		beginning = addPoint(beginning, dirAdd[direction])
		if fInRoom {
			fInRoom = c.placeHallStepInRoom(beginning, direction)
		} else {
			fInRoom = c.placeHallStepInHall(beginning, direction, hallSides{minus: fMinusFlag, plus: fPlusFlag})
		}
		direction = c.steer(beginning, end, node.beginning, direction)

		if beginning == end {
			return
		}
	}
}

// bounce flips a hall that reached the level edge back into the level
// (levels/drlg_l2.cpp:1886-1890).
func bounce(beginning Point, direction hallDirection) hallDirection {
	if !hallPastEdge(beginning, direction) {
		return direction
	}

	return hallOpposite(direction)
}

// hallPastEdge reports whether the hall origin sits on the border its direction
// runs into (levels/drlg_l2.cpp:1886-1889).
func hallPastEdge(beginning Point, direction hallDirection) bool {
	switch direction {
	case hallUp:
		return beginning.Y <= catacombHallOrigin
	case hallDown:
		return beginning.Y >= catacombHallEdge
	case hallLeft:
		return beginning.X <= catacombHallOrigin
	case hallRight:
		return beginning.X >= catacombHallEdge
	default:
		return false
	}
}

// hallOpposite is the direction a bounced hall takes.
func hallOpposite(direction hallDirection) hallDirection {
	switch direction {
	case hallUp:
		return hallDown
	case hallDown:
		return hallUp
	case hallLeft:
		return hallRight
	case hallRight:
		return hallLeft
	default:
		return direction
	}
}

// turnAtQuadrant turns a hall away from the room corner its tile carries
// (levels/drlg_l2.cpp:1896-1903).
func (c *catacombs) turnAtQuadrant(beginning Point, direction hallDirection) hallDirection {
	turn, ok := quadrantTurns[c.cell(beginning)]
	if ok && (direction == turn.against[0] || direction == turn.against[1]) {
		return turn.to
	}

	return direction
}

// placeHallStepInRoom is ConnectHall's step while the hall is still crossing a
// room: the first empty tile it reaches gets the room's exit door and the walk
// leaves the room (levels/drlg_l2.cpp:1905-1910, 1923-1925).
func (c *catacombs) placeHallStepInRoom(beginning Point, direction hallDirection) bool {
	if c.cell(beginning) != cellEmpty {
		return true
	}

	c.createDoorType(subPoint(beginning, dirAdd[direction]))
	c.setCell(beginning, cellHall)

	return false
}

// placeHallStepInHall is ConnectHall's step between rooms: a tile that is not
// empty either opens a door in a wall or reports that the hall has run into a
// room, and an empty tile takes the optional wide-hall side tiles
// (levels/drlg_l2.cpp:1911-1927).
func (c *catacombs) placeHallStepInHall(beginning Point, direction hallDirection, sides hallSides) bool {
	if c.cell(beginning) != cellEmpty {
		if c.cell(beginning) == cellWall {
			c.createDoorType(beginning)
		}

		return c.cell(beginning) != cellHall
	}

	c.placeHallSides(beginning, direction, sides)
	c.setCell(beginning, cellHall)

	return false
}

// placeHallSides adds the wide-hall tiles beside a freshly placed hall tile
// (levels/drlg_l2.cpp:1912-1922).
func (c *catacombs) placeHallSides(beginning Point, direction hallDirection, sides hallSides) {
	if sides.minus {
		c.placeHallSide(beginning, direction, -1)
	}
	if sides.plus {
		c.placeHallSide(beginning, direction, 1)
	}
}

// placeHallSide places one wide-hall tile beside the walked direction: a hall
// running vertically takes its extra tile on the left/right face and a
// horizontal one on the top/bottom face, with sign -1 for the first side and 1
// for the second (levels/drlg_l2.cpp:1913-1922).
func (c *catacombs) placeHallSide(beginning Point, direction hallDirection, sign int) {
	if direction == hallUp || direction == hallDown {
		c.placeHallExt(addPoint(beginning, Point{X: sign}))

		return
	}

	c.placeHallExt(addPoint(beginning, Point{Y: sign}))
}

// steer re-aims the hall at its target after each step
// (levels/drlg_l2.cpp:1929-1981).
//
//nolint:cyclop,revive // 1:1 port of ConnectHall's steering chain; the guards are upstream's, quirks included
func (c *catacombs) steer(beginning, end, nodeBeginning Point, direction hallDirection) hallDirection {
	nDx := absInt(end.X - beginning.X)
	nDy := absInt(end.Y - beginning.Y)

	if nDx > nDy {
		if c.d.Rnd.GenerateRnd(100) < int32(min(2*nDx, 30)) {
			direction = steerX(end, beginning)
		}
	} else if c.d.Rnd.GenerateRnd(100) < int32(min(5*nDy, 80)) {
		direction = steerY(end, beginning)
	}

	if nDy < 10 && beginning.X == end.X && isHorizontal(direction) {
		direction = steerY(end, beginning)
	}
	if nDx < 10 && beginning.Y == end.Y && isVertical(direction) {
		direction = steerX(end, beginning)
	}
	if nDy == 1 && nDx > 1 && isVertical(direction) {
		direction = steerX(end, beginning)
	}
	if nDx == 1 && nDy > 1 && isHorizontal(direction) {
		// Upstream tests beginning.x here, not beginning.y.
		if end.Y <= beginning.Y || beginning.X >= DMAXX {
			direction = hallUp
		} else {
			direction = hallDown
		}
	}
	if nDx == 0 && c.cell(beginning) != cellEmpty && isHorizontal(direction) {
		if end.X <= nodeBeginning.X || beginning.X >= DMAXX {
			direction = hallUp
		} else {
			direction = hallDown
		}
	}
	if nDy == 0 && c.cell(beginning) != cellEmpty && isVertical(direction) {
		if end.Y <= nodeBeginning.Y || beginning.Y >= DMAXY {
			direction = hallLeft
		} else {
			direction = hallRight
		}
	}

	return direction
}

// steerX aims a hall horizontally at its target.
func steerX(end, beginning Point) hallDirection {
	if end.X <= beginning.X || beginning.X >= DMAXX {
		return hallLeft
	}

	return hallRight
}

// steerY aims a hall vertically at its target.
func steerY(end, beginning Point) hallDirection {
	if end.Y <= beginning.Y || beginning.Y >= DMAXY {
		return hallUp
	}

	return hallDown
}

// isVertical reports whether a hall runs along the y axis.
func isVertical(direction hallDirection) bool {
	return direction == hallUp || direction == hallDown
}

// isHorizontal reports whether a hall runs along the x axis.
func isHorizontal(direction hallDirection) bool {
	return direction == hallLeft || direction == hallRight
}

// absInt is std::abs for the hall distance arithmetic.
func absInt(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

// countEmptyTiles counts the layout cells FillVoids has left
// (levels/drlg_l2.cpp:2119-2131).
func (c *catacombs) countEmptyTiles() int {
	t := 0
	for jj := range DMAXY {
		for ii := range DMAXX {
			if c.cells[ii][jj] == cellEmpty {
				t++
			}
		}
	}

	return t
}

// knockWalls opens a room border where the tiles on both sides of it are floor
// (levels/drlg_l2.cpp:2133-2163).
func (c *catacombs) knockWalls(x1, y1, x2, y2 int) {
	for ii := x1 + 1; ii < x2; ii++ {
		c.knockWallRowAt(ii, y1, y2)
	}

	for jj := y1 + 1; jj < y2; jj++ {
		c.knockWallColumnAt(jj, x1, x2)
	}
}

// knockWallRowAt opens the two horizontal borders of a room rectangle where a
// hall passes through them (levels/drlg_l2.cpp:2136-2147).
func (c *catacombs) knockWallRowAt(ii, y1, y2 int) {
	if c.cell(Point{X: ii, Y: y1 - 1}) == cellRoom && c.cell(Point{X: ii, Y: y1 + 1}) == cellRoom {
		c.setCell(Point{X: ii, Y: y1}, cellRoom)
	}
	if c.cell(Point{X: ii, Y: y2 - 1}) == cellRoom && c.cell(Point{X: ii, Y: y2 + 1}) == cellRoom {
		c.setCell(Point{X: ii, Y: y2}, cellRoom)
	}
	if c.cell(Point{X: ii, Y: y1 - 1}) == cellDoor {
		c.setCell(Point{X: ii, Y: y1 - 1}, cellRoom)
	}
	if c.cell(Point{X: ii, Y: y2 + 1}) == cellDoor {
		c.setCell(Point{X: ii, Y: y2 + 1}, cellRoom)
	}
}

// knockWallColumnAt is the vertical half of knockWalls
// (levels/drlg_l2.cpp:2150-2161).
func (c *catacombs) knockWallColumnAt(jj, x1, x2 int) {
	if c.cell(Point{X: x1 - 1, Y: jj}) == cellRoom && c.cell(Point{X: x1 + 1, Y: jj}) == cellRoom {
		c.setCell(Point{X: x1, Y: jj}, cellRoom)
	}
	if c.cell(Point{X: x2 - 1, Y: jj}) == cellRoom && c.cell(Point{X: x2 + 1, Y: jj}) == cellRoom {
		c.setCell(Point{X: x2, Y: jj}, cellRoom)
	}
	if c.cell(Point{X: x1 - 1, Y: jj}) == cellDoor {
		c.setCell(Point{X: x1 - 1, Y: jj}, cellRoom)
	}
	if c.cell(Point{X: x2 + 1, Y: jj}) == cellDoor {
		c.setCell(Point{X: x2 + 1, Y: jj}, cellRoom)
	}
}

// doPatternCheck turns one layout cell into the tile the pattern table picks
// for its 3x3 window (levels/drlg_l2.cpp:1984-2051).
func (c *catacombs) doPatternCheck(i, j int) {
	for k := range l2Patterns {
		if l2Patterns[k][4] == 255 {
			return
		}

		if c.patternFits(l2Patterns[k], i, j) {
			c.d.Tiles[i][j] = l2Patterns[k][9]
		}
	}
}

// patternFits reports whether the 3x3 layout window at (i,j) matches one
// pattern row (levels/drlg_l2.cpp:1988-2051).
func (c *catacombs) patternFits(pattern [10]Tile, i, j int) bool {
	for l := range 9 {
		x := i - 1 + l%3
		y := j - 1 + l/3

		if !c.patternCellFits(pattern[l], Point{X: x, Y: y}) {
			return false
		}
	}

	return true
}

// patternCellFits applies one pattern cell: 0 is a wildcard, the rest name the
// layout cells the position may hold (levels/drlg_l2.cpp:1996-2046). Positions
// outside the layer count as a match, as upstream's bounds branch does.
func (c *catacombs) patternCellFits(code Tile, p Point) bool {
	if !patternPosInLayer(p) {
		return true
	}
	if code == 0 {
		return true
	}

	allowed, ok := patternCellAllowed[code]
	if !ok {
		return false
	}

	return containsCell(allowed, c.cells[p.X][p.Y])
}

// patternPosInLayer reports whether p lies inside the layout layer.
func patternPosInLayer(p Point) bool {
	return p.X >= 0 && p.X < DMAXX && p.Y >= 0 && p.Y < DMAXY
}

func containsCell(cells []layoutCell, cell layoutCell) bool {
	for _, candidate := range cells {
		if cell == candidate {
			return true
		}
	}

	return false
}

// build runs the layout passes in upstream order and reports whether the
// result is usable (levels/drlg_l2.cpp:2445-2508).
func (c *catacombs) build(quests Quests) bool {
	c.createRoom(Point{X: 2, Y: 2}, Point{X: DMAXX - 1, Y: DMAXY - 1}, 0, hallNone, setPieceSize(quests, c.d.CurrLevel))

	for len(c.halls) > 0 {
		c.connectHall(c.halls[0])
		c.halls = c.halls[1:]
	}

	c.closeQuadrants()

	if !c.fillVoids() {
		return false
	}

	for j := range DMAXY {
		for i := range DMAXX {
			c.doPatternCheck(i, j)
		}
	}

	return true
}

// setPieceSize is CreateDungeon's quest-sized room: the level whose quest is
// available gets its room pinned to the set piece footprint, the others draw a
// random room (levels/drlg_l2.cpp:2447-2462).
func setPieceSize(quests Quests, level int) *Size {
	switch {
	case level == 5 && quests.Blood:
		return &Size{Width: 14, Height: 20}
	case level == 6 && quests.Schamb:
		return &Size{Width: 10, Height: 10}
	case level == 7 && quests.Blind:
		return &Size{Width: 15, Height: 15}
	default:
		return nil
	}
}

// closeQuadrants turns the room corner markers into walls and opens a wall ring
// around every hall cell (levels/drlg_l2.cpp:2470-2495).
func (c *catacombs) closeQuadrants() {
	for j := range DMAXY {
		for i := range DMAXX {
			c.closeQuadrantAt(i, j)
		}
	}
}

// closeQuadrantAt turns a leftover room corner into a wall, and a hall cell into
// room while walling the empty ring around it (levels/drlg_l2.cpp:2470-2495).
func (c *catacombs) closeQuadrantAt(i, j int) {
	if isQuadrant(c.cells[i][j]) {
		c.cells[i][j] = cellWall
	}
	if c.cells[i][j] != cellHall {
		return
	}

	c.cells[i][j] = cellRoom
	c.wallEmptyRingAround(i, j)
}

// wallEmptyRingAround walls the eight neighbours of a hall cell that are still
// empty (levels/drlg_l2.cpp:2484-2493).
func (c *catacombs) wallEmptyRingAround(i, j int) {
	for _, offset := range ringOffsets {
		p := addPoint(Point{X: i, Y: j}, offset)
		if c.cell(p) == cellEmpty {
			c.setCell(p, cellWall)
		}
	}
}
