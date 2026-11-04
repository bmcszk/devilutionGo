package levels

// This file ports the room-layout stage of the cathedral generator: the
// boolean mega-tile mask FirstRoom builds, and the checks and placement the
// recursive room generator runs against it (levels/drlg_l1.cpp:434-560). The
// mask is the only generator output this stage produces; MakeDmt turns it into
// tiles in a later pass.

// chamberSet is where FirstRoom lays out the three chambers and the hallway,
// in chamber order.
type chamberSet struct {
	Chambers [3]Rectangle
	Present  [3]bool
	Hallway  Rectangle
}

// firstRoom resets the layout mask and generates the cathedral's boolean room
// layout: three chambers, the hallway joining them, and the rooms that grow
// off them (levels/drlg_l1.cpp:507-551).
func (d *Dungeon) firstRoom() {
	d.Mask.Reset()
	d.rollLayoutFlags()

	rooms := d.layoutRooms()

	// The array order is upstream's call order: chamber 1, 2, 3, then the
	// hallway (levels/drlg_l1.cpp:536-543).
	for i, chamber := range rooms.Chambers {
		if rooms.Present[i] {
			d.mapRoom(chamber)
		}
	}

	d.mapRoom(rooms.Hallway)

	for i, chamber := range rooms.Chambers {
		if rooms.Present[i] {
			d.generateRoom(chamber, d.VerticalLayout)
		}
	}
}

// rollLayoutFlags draws FirstRoom's four coins and forces chamber 2 whenever a
// side chamber is missing (levels/drlg_l1.cpp:511-517).
func (d *Dungeon) rollLayoutFlags() {
	// FlipCoin() defaults to FlipCoin(2) (engine/random.hpp:368).
	d.VerticalLayout = d.Rnd.FlipCoin(2)
	d.HasChamber1 = !d.Rnd.FlipCoin(2)
	d.HasChamber2 = !d.Rnd.FlipCoin(2)
	d.HasChamber3 = !d.Rnd.FlipCoin(2)

	if !d.HasChamber1 || !d.HasChamber3 {
		d.HasChamber2 = true
	}
}

// mapRoom sets every mega tile of room in the layout mask
// (levels/drlg_l1.cpp:434-441).
func (d *Dungeon) mapRoom(room Rectangle) {
	for y := range room.Size.Height {
		for x := range room.Size.Width {
			d.Mask.Set(room.Position.X+x, room.Position.Y+y, true)
		}
	}
}

// checkRoom reports whether room lies inside the layer and touches no claimed
// mega tile (levels/drlg_l1.cpp:443-457).
func (d *Dungeon) checkRoom(room Rectangle) bool {
	for j := range room.Size.Height {
		for i := range room.Size.Width {
			if !d.unclaimedCell(i+room.Position.X, j+room.Position.Y) {
				return false
			}
		}
	}

	return true
}

// unclaimedCell reports whether a mega tile is inside the layer and still free
// (levels/drlg_l1.cpp:447-452).
func (d *Dungeon) unclaimedCell(x, y int) bool {
	if x < 0 || x >= DMAXX || y < 0 || y >= DMAXY {
		return false
	}

	return !d.Mask.Test(x, y)
}

// generateRoom searches for one room on the near side of area, carves it, and
// recurses into both sides so the layout grows outward from the chambers
// (levels/drlg_l1.cpp:459-502).
//
//nolint:funlen,cyclop,revive // 1:1 port of levels/drlg_l1.cpp:459-502; verticalLayout is upstream's parameter
func (d *Dungeon) generateRoom(area Rectangle, verticalLayout bool) {
	rotate := d.Rnd.FlipCoin(4)
	verticalLayout = (!verticalLayout && rotate) || (verticalLayout && !rotate)

	var room1 Rectangle
	var placeRoom1 bool

	for range 20 {
		width := int(d.Rnd.GenerateRnd(5)+2) &^ 1
		height := int(d.Rnd.GenerateRnd(5)+2) &^ 1
		room1 = Rectangle{Position: area.Position, Size: Size{Width: width, Height: height}}

		if verticalLayout {
			room1.Position = room1.Position.Add(Point{
				X: -room1.Size.Width,
				Y: area.Size.Height/2 - room1.Size.Height/2,
			})
			// BUGFIX: upstream passes the swapped size {height + 2, width + 1}
			// here, only p1 should be used; fixing it breaks compatibility
			// (levels/drlg_l1.cpp:474).
			placeRoom1 = d.checkRoom(Rectangle{
				Position: room1.Position.Add(Point{X: -1, Y: -1}),
				Size:     Size{Width: room1.Size.Height + 2, Height: room1.Size.Width + 1},
			})
		} else {
			room1.Position = room1.Position.Add(Point{
				X: area.Size.Width/2 - room1.Size.Width/2,
				Y: -room1.Size.Height,
			})
			placeRoom1 = d.checkRoom(Rectangle{
				Position: room1.Position.Add(Point{X: -1, Y: -1}),
				Size:     Size{Width: room1.Size.Width + 2, Height: room1.Size.Height + 1},
			})
		}

		if placeRoom1 {
			break
		}
	}

	if placeRoom1 {
		d.mapRoom(Rectangle{
			Position: room1.Position,
			Size: Size{
				Width:  min(DMAXX-room1.Position.X, room1.Size.Width),
				Height: min(DMAXY-room1.Position.Y, room1.Size.Height),
			},
		})
	}

	room2 := room1
	var placeRoom2 bool

	if verticalLayout {
		room2.Position.X = area.Position.X + area.Size.Width
		placeRoom2 = d.checkRoom(Rectangle{
			Position: room2.Position.Add(Point{X: 0, Y: -1}),
			Size:     Size{Width: room2.Size.Width + 1, Height: room2.Size.Height + 2},
		})
	} else {
		room2.Position.Y = area.Position.Y + area.Size.Height
		placeRoom2 = d.checkRoom(Rectangle{
			Position: room2.Position.Add(Point{X: -1, Y: 0}),
			Size:     Size{Width: room2.Size.Width + 2, Height: room2.Size.Height + 1},
		})
	}

	if placeRoom2 {
		d.mapRoom(room2)
	}

	if placeRoom1 {
		d.generateRoom(room1, !verticalLayout)
	}

	if placeRoom2 {
		d.generateRoom(room2, !verticalLayout)
	}
}

// layoutRooms builds and trims FirstRoom's rectangles: the hallway spans only
// the chambers that exist and both are transposed for a vertical layout
// (levels/drlg_l1.cpp:519-534). Chamber 2 is never transposed, like upstream.
func (d *Dungeon) layoutRooms() chamberSet {
	rooms := chamberSet{
		Chambers: [3]Rectangle{
			{Position: Point{X: 1, Y: 15}, Size: Size{Width: 10, Height: 10}},
			{Position: Point{X: 15, Y: 15}, Size: Size{Width: 10, Height: 10}},
			{Position: Point{X: 29, Y: 15}, Size: Size{Width: 10, Height: 10}},
		},
		Present: [3]bool{d.HasChamber1, d.HasChamber2, d.HasChamber3},
		Hallway: Rectangle{Position: Point{X: 1, Y: 17}, Size: Size{Width: 38, Height: 6}},
	}

	if !d.HasChamber1 {
		rooms.Hallway.Position.X += 17
		rooms.Hallway.Size.Width -= 17
	}

	if !d.HasChamber3 {
		rooms.Hallway.Size.Width -= 16
	}

	if d.VerticalLayout {
		swapXY(&rooms.Chambers[0].Position)
		swapXY(&rooms.Chambers[2].Position)
		swapXY(&rooms.Hallway.Position)
		rooms.Hallway.Size.Width, rooms.Hallway.Size.Height = rooms.Hallway.Size.Height, rooms.Hallway.Size.Width
	}

	return rooms
}

// swapXY transposes p in place, which is what upstream's std::swap of the x and
// y members does (levels/drlg_l1.cpp:530-533).
func swapXY(p *Point) { p.X, p.Y = p.Y, p.X }

// findArea returns the number of mega tiles the layout covers
// (levels/drlg_l1.cpp:556-559).
func (d *Dungeon) findArea() int { return d.Mask.Count() }
