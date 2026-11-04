package levels

// This file ports the theme-room pass the catacombs generator runs first:
// DRLG_PlaceThemeRooms(6, 10, 3, 0, false) and the helpers it reads
// (levels/gendung.cpp:38-137, 522-558). Only the catacombs branch of
// CreateThemeRoom is ported; the caves, nest and hell branches cannot run for
// this level type.

// catacombThemeMin and catacombThemeMax are the minSize/maxSize
// DRLG_PlaceThemeRooms is called with (levels/drlg_l2.cpp:2704).
const (
	catacombThemeMin = 6
	catacombThemeMax = 10
)

// Catacombs theme-room corner and door tiles (levels/gendung.cpp:136-139, 141-147).
const (
	catacombThemeCornerTopLeft     Tile = 8
	catacombThemeCornerTopRight    Tile = 7
	catacombThemeCornerBottomLeft  Tile = 9
	catacombThemeCornerBottomRight Tile = 6
	catacombThemeDoorSide          Tile = 4
	catacombThemeDoorEnd           Tile = 5
)

// catacombThemeSpec is the DRLG_PlaceThemeRooms(6, 10, 3, 0, false) call the
// catacombs generator makes (levels/drlg_l2.cpp:2704).
var catacombThemeSpec = themeRoomSpec{
	floor:   catacombFloor,
	minSize: catacombThemeMin,
	maxSize: catacombThemeMax,
	paint:   (*Dungeon).paintCatacombThemeRoom,
}

// catacombThemeOffset is Direction::South, which the world-tile grid maps to
// the isometric step {1, 1} (levels/gendung.cpp:551, engine/displacement.hpp:250-251).
var catacombThemeOffset = Point{X: 1, Y: 1}

// themeRoomSpec is the parameter set DRLG_PlaceThemeRooms, GetSizeForThemeRoom
// and CreateThemeRoom read: the floor tile a room grows on, its size bounds and
// the level-type branch that paints it (levels/gendung.cpp:38-137, 522-558).
type themeRoomSpec struct {
	floor   Tile
	minSize int
	maxSize int
	paint   func(d *Dungeon, room Rectangle)
}

// placeCatacombThemeRooms paints the theme rooms the level has room for
// (levels/gendung.cpp:522-558).
func (d *Dungeon) placeCatacombThemeRooms() {
	d.placeThemeRooms(catacombThemeSpec)
}

// placeThemeRooms is DRLG_PlaceThemeRooms: the scan drops the previous rooms,
// then paints one at every origin the spec accepts (levels/gendung.cpp:522-558).
// The painted rooms stay on the dungeon because IsNearThemeRoom reads them in
// later passes, exactly as upstream's themeLoc does.
func (d *Dungeon) placeThemeRooms(spec themeRoomSpec) {
	d.ThemeRooms = nil

	for j := range DMAXY {
		for i := range DMAXX {
			d.ThemeRooms = d.placeThemeRoomAt(d.ThemeRooms, Point{X: i, Y: j}, spec)
		}
	}
}

// placeCatacombThemeRoomAt is one iteration of the DRLG_PlaceThemeRooms scan:
// it paints a theme room rooted at origin and returns the enlarged room list,
// or the list unchanged when the spot is ineligible (levels/gendung.cpp:525-556).
func (d *Dungeon) placeThemeRoomAt(themes []Rectangle, origin Point, spec themeRoomSpec) []Rectangle {
	// FlipCoin(0) is upstream's unconditional frequency.
	if d.Tiles[origin.X][origin.Y] != spec.floor || !d.Rnd.FlipCoin(0) {
		return themes
	}

	room, ok := d.themeRoomSize(themes, origin, spec)
	if !ok {
		return themes
	}

	room.Position = room.Position.Add(catacombThemeOffset)
	spec.paint(d, room)

	return append(themes, room)
}

// themeRoomSize is GetSizeForThemeRoom: the largest free floor rectangle
// around origin, minus its two-tile border, or "" when the area is too small
// (levels/gendung.cpp:38-95). The bounds check is upstream's, which only
// rejects the lower-right corner when both axes overflow.
func (d *Dungeon) themeRoomSize(themes []Rectangle, origin Point, spec themeRoomSpec) (Rectangle, bool) {
	if !themeRoomOriginAllowed(themes, origin, spec) {
		return Rectangle{}, false
	}

	maxWidth := min(spec.maxSize, DMAXX-origin.X)
	maxHeight := min(spec.maxSize, DMAXY-origin.Y)

	room := Size{Width: maxWidth, Height: maxHeight}
	room, ok := d.themeRoomSweep(origin, room, Size{Width: maxWidth, Height: maxHeight}, spec)
	if !ok {
		return Rectangle{}, false
	}

	return Rectangle{
		Position: origin,
		Size:     Size{Width: room.Width - 2, Height: room.Height - 2},
	}, true
}

// themeRoomSweep is the shrinking scan of GetSizeForThemeRoom: each ring i
// clips the room to the floor runs found on row and column i, and the scan
// stops as soon as a run undershoots the minimum unless it is still the first
// minSize rings, where an undershoot rejects the whole room
// (levels/gendung.cpp:63-93).
func (d *Dungeon) themeRoomSweep(origin Point, room, bounds Size, spec themeRoomSpec) (Size, bool) {
	for i := range spec.maxSize {
		width := d.themeRoomRunWidth(origin, room, bounds.Height, i, spec.floor)
		height := d.themeRoomRunHeight(origin, room, bounds.Width, i, spec.floor)

		if width < spec.minSize || height < spec.minSize {
			if i < spec.minSize {
				return Size{}, false
			}

			break
		}

		room = Size{Width: min(room.Width, width), Height: min(room.Height, height)}
	}

	return room, true
}

// themeRoomOriginAllowed is GetSizeForThemeRoom's two early rejections: the
// original broken corner check and the already-used halo (levels/gendung.cpp:53-58).
func themeRoomOriginAllowed(themes []Rectangle, origin Point, spec themeRoomSpec) bool {
	if origin.X+spec.maxSize > DMAXX && origin.Y+spec.maxSize > DMAXY {
		return false
	}

	return !isNearThemeRoom(themes, origin)
}

// themeRoomRunWidth is the horizontal probe of the GetSizeForThemeRoom loop:
// the free floor run leftwards-to-rightwards on row origin.Y+i, seeded with i
// when the current room still exceeds it (levels/gendung.cpp:65-73).
func (d *Dungeon) themeRoomRunWidth(origin Point, room Size, maxHeight, i int, floor Tile) int {
	width := 0
	if i < room.Height {
		width = i
	}
	if i < maxHeight {
		width = d.floorRun(origin, width, room.Width, false, i, floor)
	}

	return width
}

// themeRoomRunHeight is the vertical probe of the same loop, on column
// origin.X+i (levels/gendung.cpp:75-83).
func (d *Dungeon) themeRoomRunHeight(origin Point, room Size, maxWidth, i int, floor Tile) int {
	height := 0
	if i < room.Width {
		height = i
	}
	if i < maxWidth {
		height = d.floorRun(origin, height, room.Height, true, i, floor)
	}

	return height
}

// floorRun extends one row or column run of floor tiles from origin
// (levels/gendung.cpp:65-82). fixedX reads down the column at origin.X+i, the
// other form reads across the row at origin.Y+i.
//
//nolint:revive // fixedX is the axis GetSizeForThemeRoom reads, not control coupling
func (d *Dungeon) floorRun(origin Point, run, limit int, fixedX bool, i int, floor Tile) int {
	for run < limit {
		if fixedX {
			if d.Tiles[origin.X+i][origin.Y+run] != floor {
				break
			}
		} else if d.Tiles[origin.X+run][origin.Y+i] != floor {
			break
		}

		run++
	}

	return run
}

// isNearThemeRoom reports whether origin falls inside the two-tile halo of an
// already placed theme room (levels/gendung.cpp:616-623).
func isNearThemeRoom(themes []Rectangle, origin Point) bool {
	for _, theme := range themes {
		halo := Rectangle{
			Position: Point{X: theme.Position.X - 2, Y: theme.Position.Y - 2},
			Size:     Size{Width: theme.Size.Width + 5, Height: theme.Size.Height + 5},
		}
		if halo.Contains(origin) {
			return true
		}
	}

	return false
}

// paintCatacombThemeRoom is the catacombs branch of CreateThemeRoom: it opens
// the room's transparency region, walls its border, floors its interior and
// stamps the four corner tiles (levels/gendung.cpp:96-137, 556-559).
func (d *Dungeon) paintCatacombThemeRoom(room Rectangle) {
	d.MRectTrans(Rectangle{
		Position: room.Position,
		Size:     Size{Width: room.Size.Width - 1, Height: room.Size.Height - 1},
	})

	lx, ly := room.Position.X, room.Position.Y
	hx := lx + room.Size.Width
	hy := ly + room.Size.Height

	d.fillCatacombThemeRoomBorder(lx, ly, hx, hy)
	d.stampCatacombThemeRoomCorners(lx, ly, hx, hy)
	d.carveCatacombThemeRoomDoor(lx, ly, hx, hy)
}

// fillCatacombThemeRoomBorder is CreateThemeRoom's CATACOMBS tile fill: the
// first/last row become horizontal wall, the first/last column vertical wall,
// the rest floor (levels/gendung.cpp:103-114).
func (d *Dungeon) fillCatacombThemeRoomBorder(lx, ly, hx, hy int) {
	for yy := ly; yy < hy; yy++ {
		for xx := lx; xx < hx; xx++ {
			d.Tiles[xx][yy] = catacombThemeBorderTile(xx, yy, lx, ly, hx, hy)
		}
	}
}

func catacombThemeBorderTile(xx, yy, lx, ly, hx, hy int) Tile {
	switch {
	case yy == ly || yy == hy-1:
		return catacombHWall
	case xx == lx || xx == hx-1:
		return catacombVWall
	default:
		return catacombFloor
	}
}

// stampCatacombThemeRoomCorners writes the four fixup tiles that turn the
// painted border into the catacombs' rounded room outline
// (levels/gendung.cpp:135-140).
func (d *Dungeon) stampCatacombThemeRoomCorners(lx, ly, hx, hy int) {
	d.Tiles[lx][ly] = catacombThemeCornerTopLeft
	d.Tiles[hx-1][ly] = catacombThemeCornerTopRight
	d.Tiles[lx][hy-1] = catacombThemeCornerBottomLeft
	d.Tiles[hx-1][hy-1] = catacombThemeCornerBottomRight
}

// carveCatacombThemeRoomDoor opens one door on a randomly chosen side
// (levels/gendung.cpp:141-147).
func (d *Dungeon) carveCatacombThemeRoomDoor(lx, ly, hx, hy int) {
	if d.Rnd.FlipCoin(2) {
		d.Tiles[hx-1][(ly+hy)/2] = catacombThemeDoorSide

		return
	}

	d.Tiles[(lx+hx)/2][hy-1] = catacombThemeDoorEnd
}
