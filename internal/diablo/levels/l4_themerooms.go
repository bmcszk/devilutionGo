package levels

// This file ports the hell theme rooms: DRLG_PlaceThemeRooms(7, 10, 6, 8, true)
// and the hell arm of CreateThemeRoom (levels/gendung.cpp:96-183, 532-561,
// levels/drlg_l4.cpp:1193-1195).

// hellThemeSpec is the DRLG_PlaceThemeRooms call the hell generator makes: rooms
// grow on floor tile 6, the drawn size is narrowed to a random 5..8 and the flip
// frequency is 8 (levels/drlg_l4.cpp:1194).
var hellThemeSpec = themeRoomSpec{
	floor:   hellFloor,
	minSize: hellThemeMin,
	maxSize: hellThemeMax,
	freq:    hellThemeFreq,
	rndSize: true,
	paint:   (*Dungeon).paintHellThemeRoom,
}

// placeHellThemeRooms paints the hell theme rooms (levels/drlg_l4.cpp:1193-1195).
func (d *Dungeon) placeHellThemeRooms() {
	d.placeThemeRooms(hellThemeSpec)
}

// paintHellThemeRoom is the hell arm of CreateThemeRoom: it opens the room's
// transparency region, walls its border, stamps the four corners and carves one
// doorway (levels/gendung.cpp:96-183).
func (d *Dungeon) paintHellThemeRoom(room Rectangle) {
	d.MRectTrans(Rectangle{
		Position: room.Position,
		Size:     Size{Width: room.Size.Width - 1, Height: room.Size.Height - 1},
	})

	lx, ly := room.Position.X, room.Position.Y
	hx := lx + room.Size.Width
	hy := ly + room.Size.Height

	d.fillHellThemeRoomBorder(lx, ly, hx, hy)
	d.stampHellThemeRoomCorners(lx, ly, hx, hy)
	d.carveHellThemeRoomDoor(lx, ly, hx, hy)
}

// fillHellThemeRoomBorder is CreateThemeRoom's HELL tile fill
// (levels/gendung.cpp:135-144).
func (d *Dungeon) fillHellThemeRoomBorder(lx, ly, hx, hy int) {
	for yy := ly; yy < hy; yy++ {
		for xx := lx; xx < hx; xx++ {
			d.Tiles[xx][yy] = hellThemeBorderTile(xx, yy, lx, ly, hx, hy)
		}
	}
}

func hellThemeBorderTile(xx, yy, lx, ly, hx, hy int) Tile {
	switch {
	case yy == ly || yy == hy-1:
		return hellHWall
	case xx == lx || xx == hx-1:
		return hellVWall
	default:
		return hellFloor
	}
}

// stampHellThemeRoomCorners writes the four fixup tiles that turn the painted
// border into the hell room's rounded outline (levels/gendung.cpp:156-161).
func (d *Dungeon) stampHellThemeRoomCorners(lx, ly, hx, hy int) {
	d.Tiles[lx][ly] = 9
	d.Tiles[hx-1][ly] = 16
	d.Tiles[lx][hy-1] = 15
	d.Tiles[hx-1][hy-1] = 12
}

// carveHellThemeRoomDoor opens one doorway with its two shadow tiles, on a
// randomly chosen side (levels/gendung.cpp:175-190).
func (d *Dungeon) carveHellThemeRoomDoor(lx, ly, hx, hy int) {
	if d.Rnd.FlipCoin(2) {
		yy := (ly + hy) / 2
		d.Tiles[hx-1][yy-1] = 53
		d.Tiles[hx-1][yy] = hellFloor
		d.Tiles[hx-1][yy+1] = 52
		d.Tiles[hx-2][yy-1] = 54

		return
	}

	xx := (lx + hx) / 2
	d.Tiles[xx-1][hy-1] = 57
	d.Tiles[xx][hy-1] = hellFloor
	d.Tiles[xx+1][hy-1] = 56
	d.Tiles[xx][hy-2] = 59
	d.Tiles[xx-1][hy-2] = 58
}
