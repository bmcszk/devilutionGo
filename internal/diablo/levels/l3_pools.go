package levels

// This file ports the caves lava pool: the dirt flood that measures a pocket
// and the substitution that turns it into lava, plus PoolFix which repairs the
// ceiling tiles River leaves behind (levels/drlg_l3.cpp:1268-1369, 1442-1528).

// cavePocketMax is the largest pocket PlaceLavaPool turns into a pool
// (levels/drlg_l3.cpp:1275, 1319).
const cavePocketMax = 40

// cavePoolChance is the one-in-four chance a measured pocket is kept
// (levels/drlg_l3.cpp:1466).
const cavePoolChance = 25

// cavePoolMinArea is the smallest pocket worth a pool
// (levels/drlg_l3.cpp:1477).
const cavePoolMinArea = 4

const (
	floodEnter floodVerdict = iota
	floodSkip
	floodEscape
)

const (
	pocketClosed pocketState = iota
	pocketOpen
)

// caveFloodDirs are the eight directions Spawn and SpawnEdge test, in upstream
// order. The edge directions stay in the pocket, the far ones cross into the
// neighbouring dirt (levels/drlg_l3.cpp:1288-1315, 1337-1364).
var caveFloodDirs = [...]struct {
	bit  uint8
	step Point
	edge bool
}{
	{bit: 8, step: Point{X: 0, Y: -1}, edge: true},
	{bit: 4, step: Point{X: 0, Y: 1}, edge: true},
	{bit: 2, step: Point{X: 1, Y: 0}, edge: true},
	{bit: 1, step: Point{X: -1, Y: 0}, edge: true},
	{bit: 0x80, step: Point{X: 0, Y: -1}, edge: false},
	{bit: 0x40, step: Point{X: 0, Y: 1}, edge: false},
	{bit: 0x20, step: Point{X: 1, Y: 0}, edge: false},
	{bit: 0x10, step: Point{X: -1, Y: 0}, edge: false},
}

// floodVerdict is what Spawn's entry checks decide for a candidate tile: enter
// it, skip it, or stop the whole flood (levels/drlg_l3.cpp:1275-1286,
// 1325-1336).
type floodVerdict int

// pocketState records whether the flood escaped the pocket it measured: an open
// pocket touches the map edge or grew past the area limit, which disqualifies
// it as a pool (levels/drlg_l3.cpp:1452-1464).
type pocketState int

// poolSide is one of the three nudge directions Spawn tests after the edge
// flood: the tile to flood and whether that tile is on the map
// (levels/drlg_l3.cpp:1456-1464).
type poolSide struct {
	at    Point
	onMap bool
}

// spawnEdge measures a pocket that touches the map edge
// (levels/drlg_l3.cpp:1270-1317).
func (d *Dungeon) spawnEdge(x, y int, total *int) bool {
	return d.floodDirt(x, y, total, &l3SpawnEdgeDirs, d.spawnEdge)
}

// spawn measures a pocket closed inside the layer; its low direction bits step
// through SpawnEdge and its high bits stay in Spawn
// (levels/drlg_l3.cpp:1319-1369).
func (d *Dungeon) spawn(x, y int, total *int) bool {
	return d.floodDirt(x, y, total, &l3SpawnWarpDirs, d.spawnEdge)
}

// floodDirt is the shared body of Spawn and SpawnEdge: mark the tile, count it,
// and recurse along the directions the table opens. A true result means the
// pocket grew past the area limit or ran into the map edge, which is what
// disqualifies it as a pool (levels/drlg_l3.cpp:1275-1364).
func (d *Dungeon) floodDirt(x, y int, total *int, table *[15]uint8, edgeFlood func(int, int, *int) bool) bool {
	switch d.floodVerdict(x, y, *total) {
	case floodSkip:
		return false
	case floodEscape:
		return true
	}

	tile := d.Tiles[x][y]
	d.Tiles[x][y] |= 0x80
	*total = *total + 1

	if tile == caveCeiling {
		return d.floodCeilingNeighbours(x, y, total)
	}

	// Tile 15 has no flood directions; MakeMegas never produces it, and the
	// table stops at 14.
	if int(tile) >= len(table) {
		return true
	}

	return d.floodNeighbours(x, y, total, table[tile], edgeFlood)
}

// floodVerdict applies Spawn's four entry checks in upstream order
// (levels/drlg_l3.cpp:1275-1286).
func (d *Dungeon) floodVerdict(x, y, total int) floodVerdict {
	switch {
	case total > cavePocketMax, !inCaveArea(Point{X: x, Y: y}):
		return floodEscape
	case d.Tiles[x][y]&0x80 != 0:
		return floodSkip
	case d.Tiles[x][y] > 15:
		return floodEscape
	default:
		return floodEnter
	}
}

// floodNeighbours recurses into the directions the tile's bitmask opens. The
// low bits step into the neighbouring pocket through SpawnEdge, the high bits
// stay in Spawn (levels/drlg_l3.cpp:1338-1364).
func (d *Dungeon) floodNeighbours(x, y int, total *int, bits uint8, edgeFlood func(int, int, *int) bool) bool {
	for _, dir := range caveFloodDirs {
		if bits&dir.bit == 0 {
			continue
		}

		recurse := d.spawn
		if dir.edge {
			recurse = edgeFlood
		}

		if recurse(x+dir.step.X, y+dir.step.Y, total) {
			return true
		}
	}

	return false
}

// floodCeilingNeighbours continues a flood that reached an open ceiling tile.
// Spawn ignores the direction table there and always steps into all four
// neighbours, east first, while SpawnEdge keeps reading its table
// (levels/drlg_l3.cpp:1346-1361).
func (d *Dungeon) floodCeilingNeighbours(x, y int, total *int) bool {
	for _, step := range [...]Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		if d.spawn(x+step.X, y+step.Y, total) {
			return true
		}
	}

	return false
}

// placeLavaPool floods every ceiling tile of the layer, keeps the pockets of at
// most forty tiles that do not touch the edge, and turns them into lava
// (levels/drlg_l3.cpp:1442-1495). It reports whether any pool was placed.
func (d *Dungeon) placeLavaPool() bool {
	placed := false

	// duny bounds both loops: upstream reads DMAXY for the x axis too
	// (levels/drlg_l3.cpp:1447-1448).
	for dunY := range DMAXY {
		for dunX := range DMAXY {
			if d.measurePool(dunX, dunY) {
				placed = true
			}
		}
	}

	return placed
}

// measurePool floods the pocket of one ceiling tile and paints it when it is
// small and closed (levels/drlg_l3.cpp:1449-1490).
func (d *Dungeon) measurePool(x, y int) bool {
	if d.Tiles[x][y] != caveCeiling {
		return false
	}

	d.Tiles[x][y] |= 0x80
	area := 1
	state := d.floodPoolSides(x, y, &area)

	return d.paintPool(x, y, area, state)
}

// floodPoolSides runs PlaceLavaPool's four side nudges in upstream order. Each
// side is skipped once the flood has already escaped
// (levels/drlg_l3.cpp:1452-1464).
func (d *Dungeon) floodPoolSides(x, y int, area *int) pocketState {
	escaped := true

	if x+1 < DMAXX {
		escaped = d.spawn(x+1, y, area)
	}

	sides := [...]poolSide{
		{at: Point{X: x - 1, Y: y}, onMap: x-1 > 0},
		{at: Point{X: x, Y: y + 1}, onMap: y+1 < DMAXY},
		{at: Point{X: x, Y: y - 1}, onMap: y-1 > 0},
	}

	for _, side := range sides {
		if escaped {
			// upstream's else branch: a pocket that escaped stays escaped.
			continue
		}

		escaped = d.spawnPoolSide(side, area)
	}

	if escaped {
		return pocketOpen
	}

	return pocketClosed
}

// spawnPoolSide floods one side of the nudge. A side off the map counts as an
// escaped pocket, which is what upstream's else branch sets
// (levels/drlg_l3.cpp:1456-1464).
func (d *Dungeon) spawnPoolSide(side poolSide, area *int) bool {
	if !side.onMap {
		return true
	}

	return d.spawn(side.at.X, side.at.Y, area)
}

// paintPool clears the measured pocket's markers and, when the pocket is closed
// and the dice agree, substitutes the lava tile for every dirt tile in it
// (levels/drlg_l3.cpp:1466-1490).
func (d *Dungeon) paintPool(x, y, area int, state pocketState) bool {
	if d.Rnd.GenerateRnd(100) < cavePoolChance {
		return d.placePoolTiles(x, y, area, state)
	}

	d.rejectPool(x, y, area)

	return false
}

// placePoolTiles swaps the dirt of a kept pocket for the matching lava tile
// (levels/drlg_l3.cpp:1477-1487).
func (d *Dungeon) placePoolTiles(x, y, area int, state pocketState) bool {
	placed := false

	for j := max(y-area, 0); j < min(y+area, DMAXY); j++ {
		for i := max(x-area, 0); i < min(x+area, DMAXX); i++ {
			if d.clearPoolMarkerAndPlace(i, j, area, state) {
				placed = true
			}
		}
	}

	return placed
}

// rejectPool clears the markers of a pocket the dice refused, leaving the dirt
// as it is (levels/drlg_l3.cpp:1477-1487).
func (d *Dungeon) rejectPool(x, y, area int) {
	for j := max(y-area, 0); j < min(y+area, DMAXY); j++ {
		for i := max(x-area, 0); i < min(x+area, DMAXX); i++ {
			d.clearPoolMarker(i, j)
		}
	}
}

// clearPoolMarker strips the flood marker of one pocket tile and leaves the
// tile as it was because the dice rejected the pool
// (levels/drlg_l3.cpp:1477-1487).
func (d *Dungeon) clearPoolMarker(i, j int) {
	if d.Tiles[i][j]&0x80 == 0 {
		return
	}

	d.Tiles[i][j] &^= 0x80
}

// clearPoolMarkerAndPlace strips the flood marker of one pocket tile and, when
// the pocket qualifies, swaps its dirt for the matching lava tile
// (levels/drlg_l3.cpp:1477-1487).
func (d *Dungeon) clearPoolMarkerAndPlace(i, j, area int, state pocketState) bool {
	if d.Tiles[i][j]&0x80 == 0 {
		return false
	}

	d.Tiles[i][j] &^= 0x80
	if area <= cavePoolMinArea || state != pocketClosed {
		return false
	}

	d.Tiles[i][j] = poolSubstitute(d.Tiles[i][j])

	return true
}

// poolSubstitute is PlaceLavaPool's tile swap: a dirt tile becomes the lava
// tile of its shape, anything the table rejects stays as it was
// (levels/drlg_l3.cpp:1478-1482).
func poolSubstitute(tile Tile) Tile {
	if int(tile) >= len(l3PoolSub) {
		return tile
	}

	sub := l3PoolSub[tile]
	if sub == 0 || sub > 37 {
		return tile
	}

	return sub
}

// poolFix turns every ceiling tile that touches a ground lava tile into a
// ground lava tile itself, because a pool rim was missed by River
// (levels/drlg_l3.cpp:1509-1528).
func (d *Dungeon) poolFix() {
	for y := 1; y < DMAXY-1; y++ {
		for x := 1; x < DMAXX-1; x++ {
			d.fixPoolCeiling(x, y)
		}
	}
}

// fixPoolCeiling turns one ceiling tile next to a pool into ground lava
// (levels/drlg_l3.cpp:1514-1523).
func (d *Dungeon) fixPoolCeiling(x, y int) {
	if d.Tiles[x][y] != caveCeiling {
		return
	}

	if d.touchesLava(x, y) {
		d.Tiles[x][y] = caveLavaGround
	}
}

// touchesLava reports whether the 3x3 neighbourhood of x,y holds a ground lava
// tile (levels/drlg_l3.cpp:1517-1524).
func (d *Dungeon) touchesLava(x, y int) bool {
	for j := y - 1; j <= y+1; j++ {
		for i := x - 1; i <= x+1; i++ {
			if d.Tiles[i][j] >= l3LavaGroundMin && d.Tiles[i][j] <= l3LavaGroundMax {
				return true
			}
		}
	}

	return false
}
