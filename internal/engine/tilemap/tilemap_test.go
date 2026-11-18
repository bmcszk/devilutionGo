// Package tilemap ports the fork's isometric screen-to-tile cursor math
// (Source/engine/render/scrollrt.cpp:1599-1674 ShiftGrid/RowsCoveredByPanel/
// CalcTileOffset/TilesInView; Source/cursor.cpp:732-761 ConvertToTileGrid and
// cursor.cpp:763-781 ShiftToDiamondGridAlignment @ fork ca91d778f).
package tilemap_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/tilemap"
)

// shiftGridCases pins ShiftGrid's diamond offset (scrollrt.cpp:1599-1603):
// x += vertical + horizontal, y += vertical - horizontal.
func TestShiftGrid(t *testing.T) {
	cases := []struct {
		name                 string
		startX, startY       int
		horizontal, vertical int
		wantX, wantY         int
	}{
		{
			name:       "view shift by negative half counts",
			startX:     50,
			startY:     50,
			horizontal: -5,
			vertical:   -7,
			wantX:      38,
			wantY:      48,
		},
		{name: "east pixel steps add both", startX: 38, startY: 48, horizontal: 5, vertical: 7, wantX: 50, wantY: 50},
		{name: "west pixel steps", startX: 38, startY: 48, horizontal: 4, vertical: 7, wantX: 49, wantY: 51},
		{name: "zero step is identity", startX: 12, startY: 34, horizontal: 0, vertical: 0, wantX: 12, wantY: 34},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotX, gotY := tilemap.ShiftGrid(tc.startX, tc.startY, tc.horizontal, tc.vertical)

			// then
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("ShiftGrid(%d,%d,h=%d,v=%d) = (%d,%d), want (%d,%d)",
					tc.startX, tc.startY, tc.horizontal, tc.vertical, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

// TestTilesInView pins TilesInView (scrollrt.cpp:1645-1674): ceil division of
// the viewport over the tile size.
func TestTilesInView(t *testing.T) {
	cases := []struct {
		name                  string
		width, height         int
		wantColumns, wantRows int
	}{
		{name: "640x480 divides exactly", width: 640, height: 480, wantColumns: 10, wantRows: 15},
		{name: "644x482 rounds up", width: 644, height: 482, wantColumns: 11, wantRows: 16},
		{name: "tiny viewport", width: 65, height: 33, wantColumns: 2, wantRows: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotColumns, gotRows := tilemap.TilesInView(tc.width, tc.height)

			// then
			if gotColumns != tc.wantColumns || gotRows != tc.wantRows {
				t.Fatalf("TilesInView(%d,%d) = (%d,%d), want (%d,%d)",
					tc.width, tc.height, gotColumns, gotRows, tc.wantColumns, tc.wantRows)
			}
		})
	}
}

// TestTilesInViewZoomed pins the zoom half (scrollrt.cpp:1668-1673): half the
// unzoomed counts, rounded up per axis.
func TestTilesInViewZoomed(t *testing.T) {
	// when
	gotColumns, gotRows := tilemap.TilesInViewZoomed(640, 480)

	// then
	if gotColumns != 5 || gotRows != 8 {
		t.Fatalf("TilesInViewZoomed(640,480) = (%d,%d), want (5,8)", gotColumns, gotRows)
	}
}

// rowsCoveredByPanelCases pins RowsCoveredByPanel (scrollrt.cpp:1605-1618).
func TestRowsCoveredByPanel(t *testing.T) {
	cases := []struct {
		name                    string
		screenWidth, panelWidth int
		panelHeight             int
		want                    int
	}{
		{name: "narrow screen covers view", screenWidth: 480, panelWidth: 640, panelHeight: 128, want: 0},
		{name: "wider screen shows 4 rows", screenWidth: 800, panelWidth: 640, panelHeight: 128, want: 4},
		{name: "odd panel rows floor", screenWidth: 800, panelWidth: 640, panelHeight: 140, want: 4},
		{name: "screen equal to panel width covers", screenWidth: 640, panelWidth: 640, panelHeight: 128, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := tilemap.RowsCoveredByPanel(tc.screenWidth, tc.panelWidth, tc.panelHeight)

			// then
			if got != tc.want {
				t.Fatalf("RowsCoveredByPanel(%d,%d,%d) = %d, want %d",
					tc.screenWidth, tc.panelWidth, tc.panelHeight, got, tc.want)
			}
		})
	}
}

// TestCalcTileOffset pins CalcTileOffset (scrollrt.cpp:1620-1642).
func TestCalcTileOffset(t *testing.T) {
	cases := []struct {
		name          string
		width, height int
		wantX, wantY  int
	}{
		{name: "640x480 divides exactly", width: 640, height: 480, wantX: 0, wantY: 0},
		{name: "remainder centers in", width: 656, height: 500, wantX: 24, wantY: 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotX, gotY := tilemap.CalcTileOffset(tc.width, tc.height)

			// then
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("CalcTileOffset(%d,%d) = (%d,%d), want (%d,%d)",
					tc.width, tc.height, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

// TestCalcTileOffsetZoomed pins the zoom half (scrollrt.cpp:1631-1634): the
// halved viewport's remainders.
func TestCalcTileOffsetZoomed(t *testing.T) {
	// when
	gotX, gotY := tilemap.CalcTileOffsetZoomed(640, 480)

	// then
	if gotX != 0 || gotY != 8 {
		t.Fatalf("CalcTileOffsetZoomed(640,480) = (%d,%d), want (0,8)", gotX, gotY)
	}
}

// convertToTileGridCases pins ConvertToTileGrid (cursor.cpp:732-761) against
// the 640x480 no-zoom geometry (columns 10, lowRows 15 — the even-columns/
// odd-rows alignment case, which applies NO alignment adjustment).
func TestConvertToTileGrid(t *testing.T) {
	cases := []struct {
		name         string
		x, y         int
		wantX, wantY int
	}{
		// Base tile (38,48); each pixel adds ShiftGrid(px/64, py/32).
		{name: "screen centre lands on the view tile", x: 320, y: 240, wantX: 50, wantY: 50},
		{name: "one tile east of centre", x: 384, y: 240, wantX: 51, wantY: 49},
		{name: "one tile west of centre", x: 256, y: 240, wantX: 49, wantY: 51},
		{name: "one tile south of centre", x: 320, y: 272, wantX: 51, wantY: 51},
		{name: "top-left corner", x: 0, y: 0, wantX: 38, wantY: 48},
		{name: "bottom-right corner", x: 639, y: 479, wantX: 61, wantY: 53},
		{name: "off-by-one past the south edge", x: 320, y: 480, wantX: 58, wantY: 58},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			view := tilemap.View{Columns: 10, LowRows: 15, ViewX: 50, ViewY: 50}

			// when
			gotX, gotY := tilemap.ConvertToTileGrid(view, tc.x, tc.y)

			// then
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("ConvertToTileGrid(%d,%d) = (%d,%d), want (%d,%d)",
					tc.x, tc.y, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

// TestConvertToTileGridZoom pins the zoom adjustment (cursor.cpp:753-755):
// the pixel rises half a tile before the grid step, so the same pixel resolves
// one tile further north than the unzoomed run.
func TestConvertToTileGridZoom(t *testing.T) {
	// given
	view := tilemap.View{Columns: 5, LowRows: 8, Zoom: true, ViewX: 50, ViewY: 50}

	// when
	gotX, gotY := tilemap.ConvertToTileGrid(view, 320, 256)

	// then
	// Base ShiftGrid(50,50,-2,-4) = (44,48); odd-columns/even-rows branch
	// bumps tileY to 49; zoom raises the pixel 8 -> ty = 248/32 = 7, tx = 5;
	// ShiftGrid(44,49,5,7) = (56,51).
	if gotX != 56 || gotY != 51 {
		t.Fatalf("ConvertToTileGrid(320,256,zoom) = (%d,%d), want (56,51)", gotX, gotY)
	}
}

// TestConvertToTileGridOddRowsEvenColumns pins the third alignment branch
// (cursor.cpp:749-750): odd columns with even low rows bump the base tile's
// row before the pixel steps.
func TestConvertToTileGridOddRowsEvenColumns(t *testing.T) {
	// given
	view := tilemap.View{Columns: 11, LowRows: 14, ViewX: 60, ViewY: 60}

	// when
	gotX, gotY := tilemap.ConvertToTileGrid(view, 0, 0)

	// then
	// Base (48,58) with the row bump to 59; pixel (0,0) with the +16 y shift
	// stays at ty=0 — the base tile itself.
	if gotX != 48 || gotY != 59 {
		t.Fatalf("ConvertToTileGrid(0,0) = (%d,%d), want (48,59)", gotX, gotY)
	}
}

// shiftToDiamondGridCases pins ShiftToDiamondGridAlignment
// (cursor.cpp:763-781): the tile-edge flips and the MAXDUN clamp.
func TestShiftToDiamondGrid(t *testing.T) {
	cases := []struct {
		name           string
		x, y           int
		startX, startY int
		wantX, wantY   int
		wantFlip       bool
	}{
		// px = x%64, py = y%32; flipy: py < px/2; flipx: py >= 32-px/2;
		// flipflag = (flipy&&flipx) || ((flipy||flipx) && px < 32).
		{name: "no flip at tile origin", x: 0, y: 0, startX: 45, startY: 58, wantX: 45, wantY: 58, wantFlip: false},
		{name: "top wedge flips y and flags", x: 8, y: 2, startX: 45, startY: 58, wantX: 45, wantY: 57, wantFlip: true},
		{
			name:     "left-centre wedge flips nothing",
			x:        8,
			y:        24,
			startX:   45,
			startY:   58,
			wantX:    45,
			wantY:    58,
			wantFlip: false,
		},
		{
			name:     "bottom-right wedge flips x only, no flag",
			x:        32,
			y:        30,
			startX:   45,
			startY:   58,
			wantX:    46,
			wantY:    58,
			wantFlip: false,
		},
		{name: "clamps x to MAXDUN-1", x: 8, y: 2, startX: 111, startY: 111, wantX: 111, wantY: 110, wantFlip: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotX, gotY, gotFlip := tilemap.ShiftToDiamondGridAlignment(tc.x, tc.y, tc.startX, tc.startY, 112, 112)

			// then
			if gotX != tc.wantX || gotY != tc.wantY || gotFlip != tc.wantFlip {
				t.Fatalf("ShiftToDiamondGrid(px=%d,py=%d,tile=%d,%d) = (%d,%d,%v), want (%d,%d,%v)",
					tc.x, tc.y, tc.startX, tc.startY, gotX, gotY, gotFlip, tc.wantX, tc.wantY, tc.wantFlip)
			}
		})
	}
}

// TestShiftToDiamondGridCorners pins the remaining wedge shapes
// (cursor.cpp:763-781): the flipy&&flipx wedge and negative pixels.
func TestShiftToDiamondGridCorners(t *testing.T) {
	cases := []struct {
		name           string
		x, y           int
		startX, startY int
		wantX, wantY   int
		wantFlip       bool
	}{
		// flipy&&flipx wedge: px>32 with py<px/2 and py>=32-px/2.
		{
			name:     "narrow wedge between the flips flags both",
			x:        48,
			y:        10,
			startX:   45,
			startY:   58,
			wantX:    46,
			wantY:    57,
			wantFlip: true,
		},
		{
			// Negative pixels: Go % truncates toward zero like C++11
			// (-8%64=-8), so px/2=-4 makes both wedges miss.
			name:     "negative pixel truncates toward zero, no flip",
			x:        -8,
			y:        0,
			startX:   45,
			startY:   58,
			wantX:    45,
			wantY:    58,
			wantFlip: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotX, gotY, gotFlip := tilemap.ShiftToDiamondGridAlignment(tc.x, tc.y, tc.startX, tc.startY, 112, 112)

			// then
			if gotX != tc.wantX || gotY != tc.wantY || gotFlip != tc.wantFlip {
				t.Fatalf("ShiftToDiamondGrid(px=%d,py=%d,tile=%d,%d) = (%d,%d,%v), want (%d,%d,%v)",
					tc.x, tc.y, tc.startX, tc.startY, gotX, gotY, gotFlip, tc.wantX, tc.wantY, tc.wantFlip)
			}
		})
	}
}

// screenToTileCases compose the full seam (cursor.cpp:902-931 pipeline, the
// player/walk offsets resolved by the caller into the pixel) for one spread
// of pixels: the tile the mouse path commands against.
func TestScreenToTile(t *testing.T) {
	cases := []struct {
		name         string
		x, y         int
		wantX, wantY int
	}{
		// 640x480, no zoom: columns 10, lowRows 15, view (50,50).
		{name: "centre", x: 320, y: 240, wantX: 50, wantY: 50},
		{name: "west edge", x: 0, y: 240, wantX: 45, wantY: 55},
		{name: "east edge crosses the diamond boundary", x: 639, y: 240, wantX: 55, wantY: 45},
		{name: "north edge", x: 320, y: 0, wantX: 43, wantY: 43},
		{name: "south edge", x: 320, y: 479, wantX: 57, wantY: 57},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			view := tilemap.View{Columns: 10, LowRows: 15, ViewX: 50, ViewY: 50}
			gotX, gotY := tilemap.ScreenToTile(view, tc.x, tc.y, 112, 112)

			// then
			if gotX != tc.wantX || gotY != tc.wantY {
				t.Fatalf("ScreenToTile(%d,%d) = (%d,%d), want (%d,%d)", tc.x, tc.y, gotX, gotY, tc.wantX, tc.wantY)
			}
		})
	}
}

// TestScreenToTileMutatedPixel pins the pin's non-const-ref semantics
// (cursor.cpp:732, :908-910): with even Columns && even LowRows the
// +TileHeight/2 parity shift mutates the pixel, and the diamond snap must
// read the MUTATED pixel. With the mutated py=18 the top-wedge flipy
// misses — tile stays (43,43). The un-mutated py=2 flips y: (43,42) —
// a different tile.
func TestScreenToTileMutatedPixel(t *testing.T) {
	// given
	view := tilemap.View{Columns: 10, LowRows: 14, ViewX: 50, ViewY: 50}

	// when
	gotX, gotY := tilemap.ScreenToTile(view, 328, 2, 112, 112)

	// then
	if gotX != 43 || gotY != 43 {
		t.Fatalf("ScreenToTile(328,2) = (%d,%d), want (43,43)", gotX, gotY)
	}
}
