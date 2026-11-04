package levels

// This file ports the hell layout passes: the quadrant FirstRoom carves and the
// mirrored, walled result MakeDmt turns into mega tiles
// (levels/drlg_l4.cpp:186-330). The hell layout is built in one quadrant and
// mirrored into the other three, so every room pass below only writes
// coordinates below DMAXX/2 and DMAXY/2.

// initHellDungeonFlags is InitDungeonFlags: clear the masks, then fill the
// whole layer with the hell rock tile (levels/drlg_l4.cpp:186-192).
func (d *Dungeon) initHellDungeonFlags() {
	d.Mask.Reset()
	d.Protected.Reset()

	for x := range DMAXX {
		for y := range DMAXY {
			d.Tiles[x][y] = hellRock
		}
	}
}

// mapRoom marks a room of the quadrant as carved, clipped to the quadrant
// (levels/drlg_l4.cpp:194-203, MapRoom).
func (d *Dungeon) mapHellRoom(room Rectangle) {
	for y := 0; y < room.Size.Height && y+room.Position.Y < DMAXY/2; y++ {
		for x := 0; x < room.Size.Width && x+room.Position.X < DMAXX/2; x++ {
			d.Mask.Set(room.Position.X+x, room.Position.Y+y, true)
		}
	}
}

// checkRoom reports whether a room fits the quadrant without touching an
// already carved cell (levels/drlg_l4.cpp:205-225, CheckRoom).
//
// The position test is the upstream `<= 0` one, so a room that starts on the
// quadrant's first row or column never fits.
func (d *Dungeon) checkHellRoom(room Rectangle) bool {
	if room.Position.X <= 0 || room.Position.Y <= 0 {
		return false
	}

	return d.hellRoomFits(room)
}

// hellRoomFits is CheckRoom's cell scan: every cell must lie inside the quadrant
// and be uncarved (levels/drlg_l4.cpp:211-221).
func (d *Dungeon) hellRoomFits(room Rectangle) bool {
	for y := range room.Size.Height {
		for x := range room.Size.Width {
			if !d.hellRoomCellFree(x+room.Position.X, y+room.Position.Y) {
				return false
			}
		}
	}

	return true
}

// hellRoomCellFree reports whether one cell of a candidate room lies inside the
// quadrant and is still uncarved (levels/drlg_l4.cpp:211-221).
func (d *Dungeon) hellRoomCellFree(px, py int) bool {
	if px < 0 || px >= DMAXX/2 || py < 0 || py >= DMAXY/2 {
		return false
	}

	return !d.Mask.Test(px, py)
}

// hellRoom1Fits is the probe the first candidate room must pass: it tests the
// room grown by one shell of cells (levels/drlg_l4.cpp:238-247).
//
// upstream quirk: the vertical form passes the room height as the probe width
// and adds its one spare column to the height, marked BUGFIX upstream
// (levels/drlg_l4.cpp:239).
//
//nolint:revive // verticalLayout is upstream GenerateRoom's parameter
func (d *Dungeon) hellRoom1Fits(room Rectangle, verticalLayout bool) bool {
	if verticalLayout {
		return d.checkHellRoom(Rectangle{
			Position: room.Position.Add(Point{X: -1, Y: -1}),
			Size:     Size{Width: room.Size.Height + 2, Height: room.Size.Width + 1},
		})
	}

	return d.checkHellRoom(Rectangle{
		Position: room.Position.Add(Point{X: -1, Y: -1}),
		Size:     Size{Width: room.Size.Width + 2, Height: room.Size.Height + 1},
	})
}

// generateRoom subdivides area into two rooms and recurses into each of them
// (levels/drlg_l4.cpp:227-272, GenerateRoom). The first room is probed with a
// swapped width/height check and the second slides into the neighbouring
// quadrant half; both placements are kept independently.
//
//nolint:cyclop,funlen,revive // 1:1 port of levels/drlg_l4.cpp:227-272; the room probes are upstream's
func (d *Dungeon) generateHellRoom(area Rectangle, verticalLayout bool) {
	rotate := !d.Rnd.FlipCoin(4)
	verticalLayout = (!verticalLayout && rotate) || (verticalLayout && !rotate)

	var room1 Rectangle

	var placeRoom1 bool

	for range 20 {
		randomWidth := (d.Rnd.GenerateRnd(5) + 2) &^ 1
		randomHeight := (d.Rnd.GenerateRnd(5) + 2) &^ 1
		room1.Size = Size{Width: int(randomWidth), Height: int(randomHeight)}
		room1.Position = area.Position

		if verticalLayout {
			room1.Position = room1.Position.Add(Point{
				X: -room1.Size.Width,
				Y: area.Size.Height/2 - room1.Size.Height/2,
			})
			placeRoom1 = d.hellRoom1Fits(room1, true)
		} else {
			room1.Position = room1.Position.Add(Point{
				X: area.Size.Width/2 - room1.Size.Width/2,
				Y: -room1.Size.Height,
			})
			placeRoom1 = d.hellRoom1Fits(room1, false)
		}

		if placeRoom1 {
			break
		}
	}

	if placeRoom1 {
		// upstream quirk: the height clip is taken against DMAXX, not DMAXY;
		// the two are equal so the value is unchanged (levels/drlg_l4.cpp:251).
		d.mapHellRoom(Rectangle{
			Position: room1.Position,
			Size: Size{
				Width:  min(DMAXX-room1.Position.X, room1.Size.Width),
				Height: min(DMAXX-room1.Position.Y, room1.Size.Height),
			},
		})
	}

	room2 := room1

	var placeRoom2 bool

	if verticalLayout {
		room2.Position.X = area.Position.X + area.Size.Width
		placeRoom2 = d.checkHellRoom(Rectangle{
			Position: room2.Position.Add(Point{Y: -1}),
			Size:     Size{Width: room2.Size.Width + 1, Height: room2.Size.Height + 2},
		})
	} else {
		room2.Position.Y = area.Position.Y + area.Size.Height
		placeRoom2 = d.checkHellRoom(Rectangle{
			Position: room2.Position.Add(Point{X: -1}),
			Size:     Size{Width: room2.Size.Width + 2, Height: room2.Size.Height + 1},
		})
	}

	if placeRoom2 {
		d.mapHellRoom(room2)
	}

	if placeRoom1 {
		d.generateHellRoom(room1, verticalLayout)
	}

	if placeRoom2 {
		d.generateHellRoom(room2, verticalLayout)
	}
}

// firstHellRoom is FirstRoom: it sizes and places the first room of the first
// quadrant, records the level 16 quadrant anchor and marks the set piece room
// the warlord quest protects (levels/drlg_l4.cpp:274-303).
func (d *Dungeon) firstHellRoom(opts Options) {
	room := Rectangle{Size: Size{Width: hellFirstRoomSide, Height: hellFirstRoomSide}}

	if d.CurrLevel != 16 {
		if d.CurrLevel == warlordLevel && opts.Quests.Warlord {
			room.Size = Size{Width: 11, Height: 11}
		} else {
			room.Size = Size{
				Width:  int(d.Rnd.GenerateRnd(5) + 2),
				Height: int(d.Rnd.GenerateRnd(5) + 2),
			}
		}
	}

	xmin := (DMAXX/2 - room.Size.Width) / 2
	xmax := DMAXX/2 - 1 - room.Size.Width
	ymin := (DMAXY/2 - room.Size.Height) / 2
	ymax := DMAXY/2 - 1 - room.Size.Height

	room.Position = Point{
		X: int(d.Rnd.GenerateRnd(int32(xmax-xmin+1))) + int(xmin),
		Y: int(d.Rnd.GenerateRnd(int32(ymax-ymin+1))) + int(ymin),
	}

	if d.CurrLevel == 16 {
		d.HoldL4 = room.Position
	}

	if opts.Quests.Warlord {
		d.SetPieceRoom = Rectangle{
			Position: room.Position.Add(Point{X: 1, Y: 1}),
			Size:     Size{Width: room.Size.Width + 1, Height: room.Size.Height + 1},
		}
	} else {
		d.SetPieceRoom = Rectangle{}
	}

	d.mapHellRoom(room)
	d.generateHellRoom(room, !d.Rnd.FlipCoin(2))
}

// closeOuterBorders clears the first row and column of the quadrant so the
// mirrored layout has no room hanging off the map edge
// (levels/drlg_l4.cpp:1041-1050).
func (d *Dungeon) closeOuterBorders() {
	for x := range DMAXX / 2 {
		d.Mask.ResetAt(x, 0)
	}

	for y := range DMAXY / 2 {
		d.Mask.ResetAt(0, y)
	}
}

// hellArea is FindArea: the carved quadrant area scaled to the mirrored layer
// (levels/drlg_l4.cpp:1052-1059).
func (d *Dungeon) hellArea() int { return d.Mask.Count() * 4 }

// mirrorHellLayout mirrors the first quadrant into the other three
// (levels/drlg_l4.cpp:1052-1071, MirrorDungeonLayout).
func (d *Dungeon) mirrorHellLayout() {
	for y := range DMAXY / 2 {
		for x := range DMAXX / 2 {
			if !d.Mask.Test(x, y) {
				continue
			}

			d.Mask.Set(x, DMAXY-1-y, true)
			d.Mask.Set(DMAXX-1-x, y, true)
			d.Mask.Set(DMAXX-1-x, DMAXY-1-y, true)
		}
	}
}

// makeHellDmt turns the mirrored mask into mega tiles, one 2x2 wall pattern per
// cell (levels/drlg_l4.cpp:1073-1082, MakeDmt).
func (d *Dungeon) makeHellDmt() {
	for y := range DMAXY - 1 {
		for x := range DMAXX - 1 {
			d.Tiles[x][y] = l4ConvTbl[d.hellMaskPattern(x, y)]
		}
	}
}

// hellMaskPattern is the 2x2 wall pattern MakeDmt looks the mega tile up with:
// the four mask bits around (x, y), low to high north-west, south-west,
// north-east, south-east (levels/drlg_l4.cpp:1077).
func (d *Dungeon) hellMaskPattern(x, y int) int {
	corners := [4]Point{{X: 1, Y: 1}, {X: 0, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}}

	pattern := 0
	for bit, corner := range corners {
		if d.Mask.Test(x+corner.X, y+corner.Y) {
			pattern |= 1 << (3 - bit)
		}
	}

	return pattern
}

// prepareHellInnerBorders is PrepareInnerBorders: it opens the two inner rows and
// columns of the quadrant that would otherwise be sealed off once the layout is
// mirrored (levels/drlg_l4.cpp:873-944).
//
// hallok is upstream's file-static array, reused for the rows in the first half
// and for the columns in the second; it stays a local because nothing outside
// this pass reads it.
func (d *Dungeon) prepareHellInnerBorders() {
	var hallok [DMAXY / 2]bool

	d.markHellOpenRows(&hallok)
	d.openHellInnerRow(&hallok)
	d.markHellOpenCols(&hallok)
	d.openHellInnerCol(&hallok)
}

// markHellOpenRows fills hallok for every quadrant row: a row is open when the
// first carved cell the scan reaches has floor below it and no room diagonally
// below (levels/drlg_l4.cpp:878-889).
func (d *Dungeon) markHellOpenRows(hallok *[DMAXY / 2]bool) {
	for y := DMAXY/2 - 1; y >= 0; y-- {
		hallok[y] = d.hellRowOpen(y)
	}
}

// hellRowOpen reports whether the first carved cell of row y leaves a path open
// below it, or false when the row holds no carved cell
// (levels/drlg_l4.cpp:881-889).
func (d *Dungeon) hellRowOpen(y int) bool {
	for x := DMAXX/2 - 1; x >= 0; x-- {
		if d.Mask.Test(x, y) {
			return d.hellFloorPathBelow(x, y)
		}
	}

	return false
}

// hellFloorPathBelow is the row test: the cell must have floor below it and no
// room diagonally below, which is what the mirror turns into a path
// (levels/drlg_l4.cpp:884-885).
func (d *Dungeon) hellFloorPathBelow(x, y int) bool {
	hasSouthWestRoom := y+1 < DMAXY/2 && d.Mask.Test(x, y+1)
	hasSouthRoom := x+1 < DMAXX/2 && y+1 < DMAXY/2 && d.Mask.Test(x+1, y+1)

	return hasSouthWestRoom && !hasSouthRoom
}

// markHellOpenCols is the mirrored scan the second half of PrepareInnerBorders
// runs on the columns, into the same array (levels/drlg_l4.cpp:911-924).
func (d *Dungeon) markHellOpenCols(hallok *[DMAXY / 2]bool) {
	for x := DMAXX/2 - 1; x >= 0; x-- {
		hallok[x] = d.hellColOpen(x)
	}
}

// hellColOpen reports whether the first carved cell of column x leaves a path
// open to its right (levels/drlg_l4.cpp:914-924).
func (d *Dungeon) hellColOpen(x int) bool {
	for y := DMAXY/2 - 1; y >= 0; y-- {
		if d.Mask.Test(x, y) {
			return d.hellFloorPathRight(x, y)
		}
	}

	return false
}

// hellFloorPathRight is the column test (levels/drlg_l4.cpp:917-918).
func (d *Dungeon) hellFloorPathRight(x, y int) bool {
	hasSouthEastRoom := x+1 < DMAXX/2 && d.Mask.Test(x+1, y)
	hasSouthRoom := x+1 < DMAXX/2 && y+1 < DMAXY/2 && d.Mask.Test(x+1, y+1)

	return hasSouthEastRoom && !hasSouthRoom
}

// openHellInnerRow is the first do-while of PrepareInnerBorders: it walks to the
// next open row and carves it until it runs into the quadrant's own rooms
// (levels/drlg_l4.cpp:890-909).
func (d *Dungeon) openHellInnerRow(hallok *[DMAXY / 2]bool) {
	ry := int(d.Rnd.GenerateRnd(DMAXY/2-1)) + 1

	for ry != 0 {
		ry = d.openHellRowAt(hallok, ry)
	}
}

// openHellRowAt is one iteration of that do-while: a row the scan marked open is
// carved across, otherwise the walk steps to the next row and wraps at the
// quadrant edge. A carved row reports 0, which is where the loop stops.
func (d *Dungeon) openHellRowAt(hallok *[DMAXY / 2]bool, ry int) int {
	if !hallok[ry] {
		if ry+1 == DMAXY/2 {
			return 1
		}

		return ry + 1
	}

	for x := DMAXX/2 - 1; x >= 0; x-- {
		if d.Mask.Test(x, ry) {
			return 0
		}

		d.Mask.Set(x, ry, true)
		d.Mask.Set(x, ry+1, true)
	}

	return ry
}

// openHellInnerCol is the mirrored do-while for the columns
// (levels/drlg_l4.cpp:921-940).
func (d *Dungeon) openHellInnerCol(hallok *[DMAXY / 2]bool) {
	rx := int(d.Rnd.GenerateRnd(DMAXX/2-1)) + 1

	for rx != 0 {
		rx = d.openHellColAt(hallok, rx)
	}
}

// openHellColAt is one iteration of the column do-while
// (levels/drlg_l4.cpp:921-940).
func (d *Dungeon) openHellColAt(hallok *[DMAXY / 2]bool, rx int) int {
	if !hallok[rx] {
		if rx+1 == DMAXX/2 {
			return 1
		}

		return rx + 1
	}

	for y := DMAXY/2 - 1; y >= 0; y-- {
		if d.Mask.Test(rx, y) {
			return 0
		}

		d.Mask.Set(rx, y, true)
		d.Mask.Set(rx+1, y, true)
	}

	return rx
}
