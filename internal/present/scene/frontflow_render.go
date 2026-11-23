//go:build windowed

package scene

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
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

	for i, y := range frontLabelRows(flow) {
		index := frontTextIndex
		if i == flow.menu.FocusedIndex() && flow.Screen() == FrontMainMenu {
			index = frontFocusIndex
		}
		drawFrontLabel(surface, y, uint8(index))
	}

	return nil
}

// frontLabelRows reports each label's surface y, centered in the band by the
// screen's row count.
func frontLabelRows(flow *FrontFlow) []int {
	labels := frontLines(flow)
	top := frontBandY + ((frontBandH - len(labels)*frontFocusH) / 2)

	rows := make([]int, len(labels))
	for i := range labels {
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
	case FrontSelHero:
		return []int{0, 1, 2}
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

// keep the key surface honest: the flow routes the widget kit's keys.
var _ ui.Key = ui.KeyEnter
