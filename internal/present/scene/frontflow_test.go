package scene_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// keySpy is the ui.Screen stub that records the keys it saw.
type keySpy struct {
	keys []ui.Key
}

func (s *keySpy) HandleKey(key ui.Key) { s.keys = append(s.keys, key) }

func TestFrontGameRoutesKeysToTheGameScreen(t *testing.T) {
	// given: the flow entered the game with a spy screen installed
	flow := scene.NewFrontFlow()
	spy := &keySpy{}
	flow.SetGameScreen(spy)
	flow.EnterGame()

	// when: Escape reaches the flow
	feedKeys(flow, ui.KeyEscape)

	// then: the game screen saw it
	if len(spy.keys) != 1 || spy.keys[0] != ui.KeyEscape {
		t.Fatalf("stub saw %v keys, want [KeyEscape]", spy.keys)
	}
}

func TestFrontGameRoutesArrowsToTheGameScreen(t *testing.T) {
	// given: the flow in the game with a spy screen
	flow := scene.NewFrontFlow()
	spy := &keySpy{}
	flow.SetGameScreen(spy)
	flow.EnterGame()

	// when: the arrows reach the flow
	feedKeys(flow, ui.KeyArrowUp, ui.KeyArrowDown, ui.KeyArrowLeft, ui.KeyArrowRight)

	// then: the game screen saw all four
	if len(spy.keys) != 4 {
		t.Fatalf("stub saw %d keys, want 4", len(spy.keys))
	}
}

func TestFrontGameWithoutAScreenStaysInert(t *testing.T) {
	// given: the flow in the game with NO game screen installed
	flow := scene.NewFrontFlow()
	flow.EnterGame()

	// when: keys reach the flow
	feedKeys(flow, ui.KeyEscape, ui.KeyArrowUp, ui.KeyUnknown)

	// then: nothing happened — no quit, no screen change
	if flow.Quit() {
		t.Fatal("Quit() = true on a bare FrontGame, want inert")
	}
	if flow.Screen() != scene.FrontGame {
		t.Fatalf("Screen() = %v, want FrontGame (keys dropped)", flow.Screen())
	}
}

func TestFrontGameLeaveGameReturnsToMainMenu(t *testing.T) {
	// given: the flow in the game
	flow := scene.NewFrontFlow()
	flow.SetGameScreen(&keySpy{})
	flow.EnterGame()

	// when: the game-menu's Exit row asks to leave
	flow.LeaveGame()

	// then: the flow is back on the main menu (the reverse arm the Exit row
	// owns; the menu's own walk-back is untouched)
	if flow.Screen() != scene.FrontMainMenu {
		t.Fatalf("Screen() = %v, want FrontMainMenu after LeaveGame", flow.Screen())
	}
}

func TestFrontGameRequestQuitSetsTheQuitFlag(t *testing.T) {
	// given: the flow in the game
	flow := scene.NewFrontFlow()
	flow.SetGameScreen(&keySpy{})
	flow.EnterGame()

	// when: the game-menu's Quit row asks to quit
	flow.RequestQuit()

	// then: the window's quit probe reports it
	if !flow.Quit() {
		t.Fatal("Quit() = false after RequestQuit, want true")
	}
}

func TestFrontGameIdleFrameNeverChangesTheScreen(t *testing.T) {
	// given: the flow in the game with a spy screen
	flow := scene.NewFrontFlow()
	spy := &keySpy{}
	flow.SetGameScreen(spy)
	flow.EnterGame()

	// when: idle frames feed KeyUnknown between armed keys
	feedKeys(flow, ui.KeyUnknown, ui.KeyEscape, ui.KeyUnknown)

	// then: the flow itself never moved on an idle frame
	if flow.Screen() != scene.FrontGame {
		t.Fatalf("Screen() = %v after an idle frame, want FrontGame", flow.Screen())
	}
	// and the idle keys pass through to the screen without the flow itself
	// moving
	if flow.Screen() != scene.FrontGame {
		t.Fatalf("Screen() = %v after an idle frame, want FrontGame", flow.Screen())
	}
	// then: the idle frame never sets the quit flag
	if flow.Quit() {
		t.Fatal("an idle frame must not set the quit flag")
	}
}

// feedKeys hands the flow its keys in order.
func feedKeys(flow *scene.FrontFlow, keys ...ui.Key) {
	for _, key := range keys {
		flow.HandleKey(key)
	}
}
