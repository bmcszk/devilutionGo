package levels

// floodSeed is one entry of the span-fill stack
// (levels/gendung.cpp:227-233).
type floodSeed struct {
	scanStart int
	scanEnd   int
	y         int
	dy        int
}

// IsFloor reports whether the world-tile position holds floorID, reading the
// mega-tile layer at half scale (levels/gendung.cpp:184-193). The truncation
// towards zero is upstream's: a world x < 16 maps to mega tile 0, not -1.
func (d *Dungeon) IsFloor(p Point, floorID Tile) bool {
	i := (p.X - 16) / 2
	j := (p.Y - 16) / 2

	if i < 0 || i >= DMAXX {
		return false
	}

	if j < 0 || j >= DMAXY {
		return false
	}

	return d.Tiles[i][j] == floorID
}

// fillTransparencyValues marks one floor tile and the non-floor tiles around
// it, because the span filler never visits those
// (levels/gendung.cpp:195-217).
func (d *Dungeon) fillTransparencyValues(floor Point, floorID Tile) {
	// North, South, East, West, NorthEast, NorthWest, SouthEast, SouthWest.
	for _, dir := range [8]Point{
		{X: 0, Y: -1},
		{X: 0, Y: 1},
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 1, Y: -1},
		{X: -1, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
	} {
		adjacent := floor.Add(dir)
		if !d.IsFloor(adjacent, floorID) {
			d.Trans[adjacent.X][adjacent.Y] = d.TransVal
		}
	}

	d.Trans[floor.X][floor.Y] = d.TransVal
}

// FindTransparencyValues span-fills the transparency region that contains
// floor (levels/gendung.cpp:219-291). The algorithm is the adapted Wikipedia
// span fill: it also seeds diagonally adjacent unvisited floor tiles.
//
//nolint:cyclop,funlen,revive // 1:1 port of levels/gendung.cpp:219-291; size is upstream's
func (d *Dungeon) FindTransparencyValues(floor Point, floorID Tile) {
	seedStack := []floodSeed{{scanStart: floor.X, scanEnd: floor.X + 1, y: floor.Y, dy: 1}}

	isInside := func(x, y int) bool {
		if d.Trans[x][y] != 0 {
			return false
		}

		return d.IsFloor(Point{X: x, Y: y}, floorID)
	}

	set := func(x, y int) { d.fillTransparencyValues(Point{X: x, Y: y}, floorID) }

	checkDiagonals := func(p, direction Point) {
		up := p.Add(Point{X: 0, Y: -1})
		upOver := up.Add(direction)
		if !isInside(up.X, up.Y) && isInside(upOver.X, upOver.Y) {
			seedStack = append(seedStack, floodSeed{scanStart: upOver.X, scanEnd: upOver.X + 1, y: upOver.Y, dy: -1})
		}

		down := p.Add(Point{X: 0, Y: 1})
		downOver := down.Add(direction)
		if !isInside(down.X, down.Y) && isInside(downOver.X, downOver.Y) {
			seedStack = append(seedStack, floodSeed{scanStart: downOver.X, scanEnd: downOver.X + 1, y: downOver.Y, dy: 1})
		}
	}

	for len(seedStack) > 0 {
		seed := seedStack[len(seedStack)-1]
		seedStack = seedStack[:len(seedStack)-1]

		scanLeft := seed.scanStart
		if isInside(scanLeft, seed.y) {
			for isInside(scanLeft-1, seed.y) {
				set(scanLeft-1, seed.y)
				scanLeft--
			}

			checkDiagonals(Point{X: scanLeft, Y: seed.y}, Point{X: -1, Y: 0})
		}

		if scanLeft < seed.scanStart {
			seedStack = append(seedStack,
				floodSeed{scanStart: scanLeft, scanEnd: seed.scanStart - 1, y: seed.y - seed.dy, dy: -seed.dy})
		}

		scanRight := seed.scanStart
		for scanRight < seed.scanEnd {
			for isInside(scanRight, seed.y) {
				set(scanRight, seed.y)
				scanRight++
			}

			seedStack = append(seedStack,
				floodSeed{scanStart: scanLeft, scanEnd: scanRight - 1, y: seed.y + seed.dy, dy: seed.dy})
			if scanRight-1 > seed.scanEnd {
				seedStack = append(seedStack,
					floodSeed{scanStart: seed.scanEnd + 1, scanEnd: scanRight - 1, y: seed.y - seed.dy, dy: -seed.dy})
			}

			if scanLeft < scanRight {
				checkDiagonals(Point{X: scanRight - 1, Y: seed.y}, Point{X: 1, Y: 0})
			}

			for scanRight < seed.scanEnd && !isInside(scanRight, seed.y) {
				scanRight++
			}

			scanLeft = scanRight
			if scanLeft < seed.scanEnd {
				checkDiagonals(Point{X: scanLeft, Y: seed.y}, Point{X: -1, Y: 0})
			}
		}
	}
}

// InitTrans clears the transparency window and opens region 1
// (levels/gendung.cpp:375-380, DRLG_InitTrans).
func (d *Dungeon) InitTrans() {
	d.Trans = [MAXDUNX][MAXDUNY]int8{}
	d.TransList = [256]bool{}
	d.TransVal = 1
}

// RectTrans writes TransVal over an inclusive rectangle and advances the
// region id (levels/gendung.cpp:382-394, DRLG_RectTrans).
func (d *Dungeon) RectTrans(area Rectangle) {
	for j := area.Position.Y; j <= area.Position.Y+area.Size.Height; j++ {
		for i := area.Position.X; i <= area.Position.X+area.Size.Width; i++ {
			d.Trans[i][j] = d.TransVal
		}
	}

	d.TransVal++
}

// MRectTrans is DRLG_MRectTrans over a mega-tile rectangle: the region covers
// the doubled tiles minus the last row and column
// (levels/gendung.cpp:396-399).
func (d *Dungeon) MRectTrans(area Rectangle) {
	d.RectTrans(Rectangle{
		Position: area.Position.MegaToWorld().Add(Point{X: 1, Y: 1}),
		Size:     Size{Width: area.Size.Width*2 - 1, Height: area.Size.Height*2 - 1},
	})
}

// MRectTransExtent is the extent overload of DRLG_MRectTrans
// (levels/gendung.cpp:401-404).
func (d *Dungeon) MRectTransExtent(origin, extent Point) {
	d.MRectTrans(Rectangle{
		Position: origin,
		Size:     Size{Width: extent.X - origin.X, Height: extent.Y - origin.Y},
	})
}

// CopyTrans copies one transparency value (levels/gendung.cpp:406-409).
func (d *Dungeon) CopyTrans(sx, sy, dx, dy int) {
	d.Trans[dx][dy] = d.Trans[sx][sy]
}

// FloodTransparencyValues assigns one region id to every connected area of
// floorID (levels/gendung.cpp:631-649).
func (d *Dungeon) FloodTransparencyValues(floorID Tile) {
	yy := 16
	for j := range DMAXY {
		xx := 16
		for i := range DMAXX {
			if d.Tiles[i][j] == floorID && d.Trans[xx][yy] == 0 {
				d.FindTransparencyValues(Point{X: xx, Y: yy}, floorID)
				d.TransVal++
			}

			xx += 2
		}

		yy += 2
	}
}
