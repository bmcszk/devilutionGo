//go:build windowed

package scene

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// townSurfaceWidth/Height is the world logical surface the presenter
	// scales from (G10 renderer binding; town_compose.go's 640x480 frame).
	townSurfaceWidth  = 640
	townSurfaceHeight = 480

	// frontBandY is the text band's top and frontBandH its height: a dark
	// panel behind the screen's labels, the DiabloUI full-black backdrop
	// stand-in until the h6wn text lane lands the retail font mapping.
	frontBandY = 96
	frontBandH = 240

	// frontFocusH is the focus bar's height: one label row.
	frontFocusH = 24

	// frontIndex draws the band, frontTextIndex the labels, frontFocusIndex
	// the focused row's marker; plain palette indices picked for visibility
	// in the town palette (0 black band, 215 orange text, 250 gray marker).
	frontIndex      = 0
	frontTextIndex  = 215
	frontFocusIndex = 250

	// The stats preview panel's own draw site: the right-hand column the
	// selhero screen's class baselines land in (review F2).
	frontPreviewX     = 460
	frontPreviewY     = 160
	frontPreviewW     = 160
	frontPreviewIndex = 178
)

// mainMenuItems mirrors the flow's menu order (mainmenu.cpp:72-79).
var mainMenuItems = []string{"Single Player", "Exit"}

// ComposeFrontFrame draws the active front screen into the town frame's
// surface, in front of the terminal town render (devilGo-uuyh): a dark band
// with the screen's label rows, the focused menu row marked. The labels draw
// as index bars until the h6wn text lane lands the retail font mapping (the
// retail 12-05 sheet carries no ASCII glyphs, so string labels cannot render
// yet).
func ComposeFrontFrame(surface *render.Surface, flow *FrontFlow) error {
	if surface == nil || surface.Pixels == nil {
		return fmt.Errorf("compose front frame: no town surface behind the flow")
	}

	drawFrontBand(surface)

	if flow.Screen() == FrontSelHero && flow.selheroContent != nil {
		composeSelHeroLabels(surface, flow.selheroContent)

		return nil
	}

	for i, y := range frontLabelRows(flow) {
		index := frontTextIndex
		if i == flow.menu.FocusedIndex() && flow.Screen() == FrontMainMenu {
			index = frontFocusIndex
		}

		drawFrontLabel(surface, y, uint8(index))
	}

	return nil
}

// ComposeSelHeroRows reports the selhero screen's label rows through the
// flow's installed screen: the rows the band draws and the row count the
// render tests read (review F2). The screen must expose the rows its menu
// actually carries — hero names first, then the actions.
func ComposeSelHeroRows(flow *FrontFlow) ([]string, error) {
	if flow.selheroContent == nil {
		return nil, fmt.Errorf("compose selhero rows: no selhero screen installed")
	}

	return flow.selheroContent.SelHeroRows(), nil
}

// composeSelHeroLabels draws the selhero screen's own rows: one label bar
// per list row (hero rows then the actions), the focused row marked, plus
// the stats preview panel's value rows (review F2). The panel is a second
// draw site, so the screen's preview data reaches the frame.
func composeSelHeroLabels(surface *render.Surface, content SelHeroContent) {
	drawSelHeroList(surface, content)
	drawSelHeroPreview(surface, content.SelHeroPreviewRows())
}

// drawSelHeroList draws the screen's list rows, the focused row marked.
func drawSelHeroList(surface *render.Surface, content SelHeroContent) {
	rows := content.SelHeroRows()
	focused := content.SelHeroFocused()

	for i, y := range frontLabelRowsFor(len(rows)) {
		index := frontTextIndex
		if i == focused {
			index = frontFocusIndex
		}

		drawFrontLabel(surface, y, uint8(index))
	}
}

// drawSelHeroPreview draws the stats preview panel: one value bar per class
// baseline the screen loaded, right of the list (selhero.cpp:527-546).
func drawSelHeroPreview(surface *render.Surface, rows int) {
	for i := range rows {
		y := frontPreviewY + i*frontFocusH
		for dy := 4; dy < frontFocusH-4; dy++ {
			row := (y + dy) * surface.Width
			for x := frontPreviewX; x < frontPreviewX+frontPreviewW; x++ {
				surface.Pixels[row+x] = frontPreviewIndex
			}
		}
	}
}

// frontLabelRows reports each label's surface y, centered in the band by the
// screen's row count.
func frontLabelRows(flow *FrontFlow) []int {
	labels := frontLines(flow)

	return frontLabelRowsFor(len(labels))
}

// frontLabelRowsFor centers rowCount label rows in the front band.
func frontLabelRowsFor(rowCount int) []int {
	top := frontBandY + ((frontBandH - rowCount*frontFocusH) / 2)

	rows := make([]int, rowCount)
	for i := range rows {
		rows[i] = top + i*frontFocusH
	}

	return rows
}

// frontLines renders the active screen's label row count: one row per label.
func frontLines(flow *FrontFlow) []int {
	switch flow.Screen() {
	case FrontTitle:
		return []int{0, 1}
	case FrontMainMenu:
		return make([]int, len(mainMenuItems))
	default:
		return nil
	}
}

// drawFrontBand fills the band rect with the backdrop index.
func drawFrontBand(surface *render.Surface) {
	for y := frontBandY; y < frontBandY+frontBandH; y++ {
		row := y * surface.Width
		for x := 160; x < surface.Width-160; x++ {
			surface.Pixels[row+x] = frontIndex
		}
	}
}

// drawFrontLabel draws one label row: a bar at the row's indent, wider when
// it carries the focus marker.
func drawFrontLabel(surface *render.Surface, y int, index uint8) {
	for dy := 4; dy < frontFocusH-4; dy++ {
		row := (y + dy) * surface.Width
		for x := 200; x < 440; x++ {
			surface.Pixels[row+x] = index
		}
	}
}
