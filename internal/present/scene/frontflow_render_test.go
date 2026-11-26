//go:build windowed

package scene_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// --- S5: the game-menu render branch (behavioural legs need no assets) ---

func TestFrontGameGameMenuRows(t *testing.T) {
	// given: a flow with a game-menu content stub installed and entered
	flow := scene.NewFrontFlow()
	stub := &menuContentStub{
		rows:    []string{"Options", "Save Game", "Load Game", "Exit to Main Menu", "Quit Game"},
		focused: 1,
		open:    true,
	}
	flow.SetGameScreen(stub)
	flow.SetGameMenuContent(stub)
	flow.EnterGame()

	// when: the band's row geometry is asked for
	rows := scene.FrontGameMenuRows(flow)

	// then: five rows centered by the root table's row count
	if len(rows) != 5 {
		t.Fatalf("FrontGameMenuRows = %d rows, want 5", len(rows))
	}
	top := rows[0]
	for i, y := range rows {
		if want := top + i*24; y != want {
			t.Fatalf("row %d y = %d, want %d (frontFocusH stride)", i, y, want)
		}
	}
}

func TestGameMenuRenderSliderFillGrowsPerStep(t *testing.T) {
	// review F6 shape: the fill bar grows by (x1-x0)/46 px per Speed step —
	// composed through the real render pass, no assets needed
	composeWithFill := func(fill float64) int {
		flow := scene.NewFrontFlow()
		stub := &menuContentStub{
			rows:    []string{"Speed"},
			focused: 0,
			open:    true,
			slider:  float64(fill),
		}
		flow.SetGameScreen(stub)
		flow.SetGameMenuContent(stub)
		flow.EnterGame()

		surface := &render.Surface{Width: 640, Height: 480, Pixels: make([]uint8, 640*480)}
		if err := scene.ComposeFrontFrame(surface, flow); err != nil {
			t.Fatalf("ComposeFrontFrame: %v", err)
		}

		n := 0
		for _, px := range surface.Pixels {
			if px == 176 { // sliderFillIndex
				n++
			}
		}

		return n
	}

	n0 := composeWithFill(0)
	n1 := composeWithFill(1.0 / 46)

	// one step's worth of columns is span/46 = 3.04 -> the int cast draws 3
	// columns, times the 8-pixel bar height = 24 pixels
	want := (140 / 46) * frontFocusBarH // 3 * 8
	if n0 != 0 {
		t.Fatalf("fill 0 drew %d pixels, want 0", n0)
	}
	if got := n1 - n0; got < want || got > want+1 {
		t.Fatalf(
			"fill step 0->1/46 drew %d pixels (total %d), want ~%d (span 140 / 46 steps)",
			got,
			n1,
			want,
		)
	}
}

// frontFocusBarH is the fill bar's drawn height (drawSliderFill's dy loop,
// frontFocusH-8-8).
const frontFocusBarH = 8

// menuContentStub is the GameMenuContent the render test drives.
type menuContentStub struct {
	rows    []string
	focused int
	open    bool
	slider  float64 // the fill 0..1 the row reports; negative = no slider
}

func (s *menuContentStub) HandleKey(ui.Key)       {}
func (s *menuContentStub) GameMenuOpen() bool     { return s.open }
func (s *menuContentStub) GameMenuRows() []string { return s.rows }
func (s *menuContentStub) GameMenuFocused() int   { return s.focused }
func (s *menuContentStub) GameMenuSlider() float64 {
	return s.slider
}
