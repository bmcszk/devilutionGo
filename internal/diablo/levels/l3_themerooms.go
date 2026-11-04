package levels

// This file ports the caves branch of the theme-room pass:
// DRLG_PlaceThemeRooms(5, 10, 7, 0, false) and the caves arm of
// CreateThemeRoom (levels/drlg_l3.cpp:2062, levels/gendung.cpp:96-160).

// caveThemeMin and caveThemeMax are the minSize/maxSize the caves generator
// places theme rooms with (levels/drlg_l3.cpp:2062).
const (
	caveThemeMin = 5
	caveThemeMax = 10
)

// Caves theme-room wall, corner and door tiles
// (levels/gendung.cpp:100-160).
const (
	caveThemeWallHorizontal Tile = 134
	caveThemeWallVertical   Tile = 137
	caveThemeCornerTopLeft  Tile = 150
	caveThemeCornerTopRight Tile = 151
	caveThemeCornerBottomL  Tile = 152
	caveThemeCornerBottomR  Tile = 138
	caveThemeDoorVertical   Tile = 147
	caveThemeDoorHorizontal Tile = 146
)

// caveThemeSpec is the DRLG_PlaceThemeRooms(5, 10, 7, 0, false) call the caves
// generator makes (levels/drlg_l3.cpp:2062).
var caveThemeSpec = themeRoomSpec{
	floor:   caveFloor,
	minSize: caveThemeMin,
	maxSize: caveThemeMax,
	paint:   (*Dungeon).paintCaveThemeRoom,
}

// placeCaveThemeRooms paints the caves theme rooms and opens a transparency
// region over each (levels/gendung.cpp:552-556).
func (d *Dungeon) placeCaveThemeRooms() {
	d.placeThemeRooms(caveThemeSpec)
}

// paintCaveThemeRoom is the caves arm of CreateThemeRoom: it walls the room,
// stamps the four corners and carves one door
// (levels/gendung.cpp:96-160). The transparency region is written before the
// paint, as upstream does.
func (d *Dungeon) paintCaveThemeRoom(room Rectangle) {
	d.RectTrans(Rectangle{
		Position: room.Position.Add(catacombThemeOffset).MegaToWorld(),
		Size:     Size{Width: room.Size.Width*2 - 5, Height: room.Size.Height*2 - 5},
	})

	lx, ly := room.Position.X, room.Position.Y
	hx := lx + room.Size.Width
	hy := ly + room.Size.Height

	d.paintCaveThemeWalls(lx, ly, hx, hy)
	d.stampCaveThemeCorners(lx, ly, hx, hy)
	d.carveCaveThemeDoor(lx, ly, hx, hy)
}

// paintCaveThemeWalls fills the room with floor and its border with the two
// wall faces (levels/gendung.cpp:100-115).
func (d *Dungeon) paintCaveThemeWalls(lx, ly, hx, hy int) {
	for yy := ly; yy < hy; yy++ {
		for xx := lx; xx < hx; xx++ {
			d.Tiles[xx][yy] = caveThemeWalledTile(xx, yy, Rectangle{
				Position: Point{X: lx, Y: ly},
				Size:     Size{Width: hx - lx, Height: hy - ly},
			})
		}
	}
}

// caveThemeWalledTile is the caves wall/floor choice for one tile of the room:
// the first and last row become the horizontal face, the first and last column
// the vertical one, the rest floor (levels/gendung.cpp:103-114).
func caveThemeWalledTile(xx, yy int, room Rectangle) Tile {
	lastColumn := room.Position.X + room.Size.Width - 1
	lastRow := room.Position.Y + room.Size.Height - 1

	switch {
	case yy == room.Position.Y || yy == lastRow:
		return caveThemeWallHorizontal
	case xx == room.Position.X || xx == lastColumn:
		return caveThemeWallVertical
	default:
		return caveFloor
	}
}

// stampCaveThemeCorners writes the four corner tiles
// (levels/gendung.cpp:141-147).
func (d *Dungeon) stampCaveThemeCorners(lx, ly, hx, hy int) {
	d.Tiles[lx][ly] = caveThemeCornerTopLeft
	d.Tiles[hx-1][ly] = caveThemeCornerTopRight
	d.Tiles[lx][hy-1] = caveThemeCornerBottomL
	d.Tiles[hx-1][hy-1] = caveThemeCornerBottomR
}

// carveCaveThemeDoor cuts the room's single door into one of the two middle
// walls (levels/gendung.cpp:149-152).
func (d *Dungeon) carveCaveThemeDoor(lx, ly, hx, hy int) {
	if d.Rnd.FlipCoin(2) {
		d.Tiles[hx-1][(ly+hy)/2] = caveThemeDoorVertical

		return
	}

	d.Tiles[(lx+hx)/2][hy-1] = caveThemeDoorHorizontal
}
