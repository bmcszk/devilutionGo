// Package tilemap ports the fork's isometric screen-to-tile cursor math —
// the generic half of the mouse click-to-walk seam (devilGo-c6wt s2). The
// Diablo-specific verb decision lives in internal/diablo/controls; this
// package carries only the diamond-grid geometry (no diablo imports, ADR-008).
//
// Upstream anchors (fork ca91d778f): ShiftGrid, RowsCoveredByPanel,
// CalcTileOffset, TilesInView (Source/engine/render/scrollrt.cpp:1599-1674);
// ConvertToTileGrid, ShiftToDiamondGridAlignment (Source/cursor.cpp:732-781,
// called from CheckCursMove cursor.cpp:902-931). All functions take explicit
// inputs (house pattern): the screen geometry and view position are resolved
// by the caller, so the math is testable headless.
package tilemap

// Tile geometry constants (levels/dun_tile.hpp:7-8; gendung_defs.hpp:6-9).
const (
	// TileWidth is TILE_WIDTH, the isometric tile sprite width.
	TileWidth = 64
	// TileHeight is TILE_HEIGHT, half the width.
	TileHeight = 32
)

// adjustedPixel is the pixel AFTER the parity/zoom mutations (the pin's
// non-const Point& escapes them, cursor.cpp:732, :908-910) — what the
// diamond snap must read.
type adjustedPixel struct {
	x, y int
}

// View is the frame state ConvertToTileGrid resolves a pixel against: the
// tile-grid projection of the camera and the viewport geometry, exactly the
// inputs upstream's ConvertToTileGrid reads from the render globals
// (cursor.cpp:733-735: TilesInView, RowsCoveredByPanel, ViewPosition).
type View struct {
	// Columns is TilesInView's column count (scrollrt.cpp:1645).
	Columns int
	// LowRows is rows minus RowsCoveredByPanel — the rows the world viewport
	// really shows under any open panel (cursor.cpp:736: lrow).
	LowRows int
	// Zoom is Graphics.zoom (cursor.cpp:753).
	Zoom bool
	// ViewX, ViewY is ViewPosition, the tile the camera centres on.
	ViewX, ViewY int
}

// ShiftGrid ports ShiftGrid (scrollrt.cpp:1599-1603): move a tile coordinate
// along the diamond axes — horizontal steps run south-east, vertical steps
// south-west. Returns the shifted (x, y).
func ShiftGrid(x, y, horizontal, vertical int) (gridX, gridY int) {
	return x + vertical + horizontal, y + vertical - horizontal
}

// TilesInView ports TilesInView (scrollrt.cpp:1645-1674): the tile counts the
// viewport spans, rounded up.
func TilesInView(width, height int) (columns, rows int) {
	columns = width / TileWidth
	if width%TileWidth != 0 {
		columns++
	}
	rows = height / TileHeight
	if height%TileHeight != 0 {
		rows++
	}
	return columns, rows
}

// TilesInViewZoomed ports the zoom half of TilesInView (scrollrt.cpp:1668-1673):
// half the unzoomed counts, rounded up per axis.
func TilesInViewZoomed(width, height int) (columns, rows int) {
	columns, rows = TilesInView(width, height)
	if columns&1 != 0 {
		columns++
	}
	columns /= 2
	if rows&1 != 0 {
		rows++
	}
	rows /= 2
	return columns, rows
}

// RowsCoveredByPanel ports RowsCoveredByPanel (scrollrt.cpp:1605-1618): the
// world rows the open main panel hides — zero when the screen is narrower
// than the panel, else the panel height in tiles (halved under zoom).
func RowsCoveredByPanel(screenWidth, panelWidth, panelHeight int) int {
	if screenWidth <= panelWidth {
		return 0
	}

	return panelHeight / TileHeight
}

// RowsCoveredByPanelZoomed ports the zoom half of RowsCoveredByPanel
// (scrollrt.cpp:1614-1616): the unzoomed count halved.
func RowsCoveredByPanelZoomed(screenWidth, panelWidth, panelHeight int) int {
	return RowsCoveredByPanel(screenWidth, panelWidth, panelHeight) / 2
}

// CalcTileOffset ports CalcTileOffset (scrollrt.cpp:1620-1642): the pixel
// remainder that centres the half-tile fringe. Returns (offsetX, offsetY) —
// the amounts AlterMousePositionViaPlayer adds back to the mouse pixel
// (cursor.cpp:703-706).
func CalcTileOffset(width, height int) (offsetX, offsetY int) {
	return tileOffset(width%TileWidth, height%TileHeight)
}

// CalcTileOffsetZoomed ports the zoom half (scrollrt.cpp:1631-1634): the
// halved viewport's remainders.
func CalcTileOffsetZoomed(width, height int) (offsetX, offsetY int) {
	return tileOffset((width/2)%TileWidth, (height/2)%TileHeight)
}

// ConvertToTileGrid ports ConvertToTileGrid (cursor.cpp:732-761): map the
// adjusted screen pixel to the dungeon tile it lands on, relative to the
// view. The caller has already applied the panel, scrolling, zoom and
// player-walk pixel adjustments of cursor.cpp:907-917.
func ConvertToTileGrid(view View, screenX, screenY int) (tileX, tileY int) {
	tileX, tileY, _ = convertToTileGridAdjusted(view, screenX, screenY)
	return tileX, tileY
}

// ShiftToDiamondGridAlignment ports ShiftToDiamondGridAlignment
// (cursor.cpp:763-781): snap the tile to the diamond cell the pixel really
// falls in and report flipflag — the pixel-in-tile wedge the target
// resolution reads. maxDunX/maxDunY are the clamp bounds.
func ShiftToDiamondGridAlignment(
	screenX, screenY, tileX, tileY, maxDunX, maxDunY int,
) (snappedX, snappedY int, flipFlag bool) {
	px := screenX % TileWidth
	py := screenY % TileHeight

	// flipy: py < px/2 -> tile.y-- (cursor.cpp:768-771).
	deltaY := 0
	if py < px/2 {
		deltaY = -1
	}
	// flipx: py >= TH - px/2 -> tile.x++ (cursor.cpp:772-775).
	deltaX := 0
	if py >= TileHeight-px/2 {
		deltaX = 1
	}

	snappedX = clamp(tileX+deltaX, 0, maxDunX-1)
	snappedY = clamp(tileY+deltaY, 0, maxDunY-1)

	flipy := deltaY == -1
	flipx := deltaX == 1
	flipFlag = (flipy && flipx) || ((flipy || flipx) && px < TileWidth/2)

	return snappedX, snappedY, flipFlag
}

// ScreenToTile composes the cursor seam (cursor.cpp:902-931) for the frame
// the caller measured: ConvertToTileGrid then ShiftToDiamondGridAlignment.
// The pixel must already carry the panel/scroll/zoom/player adjustments
// (cursor.cpp:910-917); the caller resolves those from presentation state.
func ScreenToTile(view View, screenX, screenY, maxDunX, maxDunY int) (tileX, tileY int) {
	// cursor.cpp:908-910: the snap reads the pixel AFTER ConvertToTileGrid's
	// parity/zoom mutations escape through the Point&.
	tileX, tileY, pixel := convertToTileGridAdjusted(view, screenX, screenY)
	tileX, tileY, _ = ShiftToDiamondGridAlignment(
		pixel.x, pixel.y, tileX, tileY, maxDunX, maxDunY,
	)

	return tileX, tileY
}

// convertToTileGridAdjusted ports the same rows with the pixel mutation that
// the parity and zoom branches apply upstream: screenPosition is a non-const
// ref (cursor.cpp:732), so the +TH/2 alignment shift and the -TH/4 zoom
// shift escape into the caller's pixel and ShiftToDiamondGridAlignment sees
// the mutated value (cursor.cpp:908-910).
func convertToTileGridAdjusted(view View, screenX, screenY int) (tileX, tileY int, pixel adjustedPixel) {
	// Center player tile on screen (cursor.cpp:738-740).
	tileX, tileY = ShiftGrid(view.ViewX, view.ViewY, -view.Columns/2, -view.LowRows/2)

	// Align grid (cursor.cpp:742-750): the parity of the tile counts decides
	// which half-tile the screen origin sits in.
	if view.Columns%2 == 0 && view.LowRows%2 == 0 {
		screenY += TileHeight / 2
	} else if view.Columns%2 != 0 && view.LowRows%2 != 0 {
		screenX -= TileWidth / 2
	} else if view.Columns%2 != 0 && view.LowRows%2 == 0 {
		tileY++
	}

	if view.Zoom {
		screenY -= TileHeight / 4 // cursor.cpp:753-755
	}

	tx := screenX / TileWidth
	ty := screenY / TileHeight

	tileX, tileY = ShiftGrid(tileX, tileY, tx, ty) // cursor.cpp:757-759
	return tileX, tileY, adjustedPixel{x: screenX, y: screenY}
}

// clamp is std::clamp (cursor.cpp:776-777).
func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

// tileOffset is the shared centering math (scrollrt.cpp:1636-1642).
func tileOffset(remX, remY int) (offsetX, offsetY int) {
	x, y := remX, remY

	if x != 0 {
		x = (TileWidth - x) / 2
	}
	if y != 0 {
		y = (TileHeight - y) / 2
	}

	return x, y
}
