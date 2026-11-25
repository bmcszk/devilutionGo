//go:build windowed

package scene_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// --- S5: the game-menu render branch (behavioural legs need no assets) ---

func TestFrontGameGameMenuRows(t *testing.T) {
	// given: a flow with a game-menu content stub installed and entered
	flow := scene.NewFrontFlow()
	stub := &menuContentStub{rows: []string{"Options", "Save Game", "Load Game", "Exit to Main Menu", "Quit Game"}, focused: 1, open: true}
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

// menuContentStub is the GameMenuContent the render test drives.
type menuContentStub struct {
	rows    []string
	focused int
	open    bool
	slider  int // 0..46, -1 = row is no slider
}

func (s *menuContentStub) HandleKey(ui.Key)       {}
func (s *menuContentStub) GameMenuOpen() bool     { return s.open }
func (s *menuContentStub) GameMenuRows() []string { return s.rows }
func (s *menuContentStub) GameMenuFocused() int   { return s.focused }
func (s *menuContentStub) GameMenuSlider() float64 {
	if s.slider < 0 {
		return -1
	}

	return float64(s.slider) / 46
}
