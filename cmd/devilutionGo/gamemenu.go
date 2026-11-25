//go:build windowed

package main

import (
	"context"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// clearOutcomes disarms every flow: the Exit-to-Main-Menu row resets the
// lane's pending hooks before the flow walks back.
func (l *deviceLane) clearOutcomes() {
	for i := range l.outcomes {
		l.outcomes[i].pending, l.outcomes[i].hook = false, nil
	}
}

// gamemenu row indices inside the root table, upstream order
// (X:gamemenu.cpp:49-57 sgSingleMenu).
const (
	gamemenuRowOptions = iota
	gamemenuRowSave
	gamemenuRowLoad
	gamemenuRowExit
	gamemenuRowQuit
)

// errOptionsPending is the named stand-in for the Options row until S5
// replaces it with the real submenu: activating it surfaces this refusal as
// the screen's error row, the "not available" stub the bean's scope asks for.
var errOptionsPending = fmt.Errorf("game menu: options pending (S5)")

// gameMenuScreen is the in-game Esc menu (X:gamemenu.cpp, devilGo-1a4r S4/S5):
// cmd composition over the lane (submit site + tick), the persist archive
// (HasGame for the Load row) and the front flow (LeaveGame / RequestQuit).
// The Diablo-neutral table model lives in internal/game/ui; this screen keys
// its actions by row index.
type gameMenuScreen struct {
	lane    *deviceLane
	flow    *scene.FrontFlow
	stores  *persist.HeroStoreSet
	slot    int
	seed    uint32
	onRate  func(int)
	enabled func() bool // the dead-player gate stand-in (today: constant alive)

	// active marks the menu open (gmenu_is_active): the flow's MenuOpen
	// closure reads it and ProcessInput pauses the sim on it.
	active bool

	// table is the active table model; the root table until the Options
	// submenu swaps it in (X:gamemenu.cpp:185-191 GamemenuOptions).
	table *ui.MenuTable

	// lastError surfaces a refused action as the screen's error row.
	lastError error
}

// newGameMenuScreen builds the screen over the boot-owned surfaces. slot is
// the hero slot threaded from the selhero entry (the save target); seed is the
// session's level seed (0 for a fresh town — the load handler does not
// validate B for an unloaded town record, disclosed in the lane). enabled is
// the dead-player gate; nil means constant alive (disclosed stand-in — death
// has not landed). onRate is the SetTickRate presentation callback; nil for
// none.
func newGameMenuScreen(
	lane *deviceLane, flow *scene.FrontFlow, stores *persist.HeroStoreSet,
	slot int, onRate func(int),
) *gameMenuScreen {
	return &gameMenuScreen{
		lane: lane, flow: flow, stores: stores,
		slot: slot, onRate: onRate,
		table: newGameMenuRootTable(),
	}
}

// newGameMenuRootTable builds the single-player root table in upstream order
// (X:gamemenu.cpp:49-57); enabled flags refresh at open().
func newGameMenuRootTable() *ui.MenuTable {
	return ui.NewMenuTable([]ui.MenuItem{
		{Label: "Options", Enabled: true},
		{Label: "Save Game", Enabled: true},
		{Label: "Load Game", Enabled: true},
		{Label: "Exit to Main Menu", Enabled: true},
		{Label: "Quit Game", Enabled: true},
	})
}

// isOpen reports the menu's active state (gmenu_is_active): the flow's
// MenuOpen closure and the window's movement gate read it.
func (m *gameMenuScreen) isOpen() bool { return m.active }

// LastError reports the surfaced refusal of the last action, or nil.
func (m *gameMenuScreen) LastError() error { return m.lastError }

// Rows exposes the active table's rows to the render pass.
func (m *gameMenuScreen) Rows() []ui.MenuItem { return m.table.Rows() }

// Focused reports the focused row index.
func (m *gameMenuScreen) Focused() int { return m.table.Focused() }

// open activates the menu: the update-fn role (X:gamemenu.cpp:91-99
// GamemenuUpdateSingle) recomputes the enabled flags — Load iff the slot
// holds a game — then focus lands on the FIRST ENABLED row
// (X:gmenu.cpp:238-244).
func (m *gameMenuScreen) open(slot int) {
	m.slot = slot
	m.table = newGameMenuRootTable()

	// Load Game enabled iff the slot holds a frozen session (gbValidSaveFile's
	// Go equivalent: persist.SessionArchive.HasGame, selhero.go:370-372 shape).
	archive := persist.NewSessionArchive(m.stores, lanePackSource)
	has, err := archive.HasGame(context.Background(), m.slot, string(persist.KindSingle))
	if err != nil {
		m.lastError = fmt.Errorf("game menu load check: %w", err)
		has = false
	}
	rows := m.table.Rows()
	rows[gamemenuRowLoad].Enabled = has

	// Options enabled iff the player is alive (X:gamemenu.cpp:97-99); the
	// death gate has not landed, so the stand-in is constant alive.
	if m.enabled != nil && !m.enabled() {
		rows[gamemenuRowOptions].Enabled = false
	}

	// Save Game enabled stands in for the hand-cursor gate
	// (X:gamemenu.cpp:333-335 pcurs != CURSOR_HAND refusal) — constant true
	// until the cursor lands. DISCLOSED stand-in.

	m.table.Focus()
	m.active = true
}

// close deactivates the menu only (X:gmenu.cpp:290-292: Esc closes the
// menu; nothing else).
func (m *gameMenuScreen) close() { m.active = false }

// HandleKey routes the menu keys: navigation, activation, Esc close
// (X:gmenu.cpp:280-311 gmenu_presskeys). KeyUnknown is an idle frame and is
// never a press.
func (m *gameMenuScreen) HandleKey(key ui.Key) {
	if key == ui.KeyUnknown {
		return
	}

	if !m.active {
		// Esc on FrontGame with the menu closed opens it (X:diablo.cpp:546-551).
		if key == ui.KeyEscape {
			m.open(m.slot)
		}

		return
	}

	switch key {
	case ui.KeyEscape:
		m.close()
	case ui.KeyEnter:
		m.selectRow(m.table.Focused())
	case ui.KeyArrowDown:
		m.table.MoveDown()
	case ui.KeyArrowUp:
		m.table.MoveUp()
	default:
	}
}

// selectRow runs the focused row's action (X:gmenu.cpp:285-289: Enter on an
// enabled row only).
func (m *gameMenuScreen) selectRow(idx int) {
	if !m.table.Rows()[idx].Enabled {
		return
	}

	switch idx {
	case gamemenuRowOptions:
		m.lastError = errOptionsPending // REPLACED in S5 by the submenu
	case gamemenuRowSave:
		m.saveGame()
	case gamemenuRowLoad:
		m.loadGame()
	case gamemenuRowExit:
		m.exitToMainMenu()
	case gamemenuRowQuit:
		m.quitGame()
	}
}

// saveGame submits KindSaveSlot{A: slot, B: seed} through the lane's bridge
// session (X:gamemenu.cpp:331-364 gamemenu_save_game's SaveGame verb) and
// arms the lane's save outcome hook.
func (m *gameMenuScreen) saveGame() {
	if m.lane == nil {
		return
	}

	cmd := kernel.Command{
		Kind: kernel.KindSaveSlot,
		Tick: m.lane.game.TickCount,
		A:    int32(m.slot),
		B:    int32(m.seed),
	}

	if !m.lane.input.SubmitDecision(cmd) {
		return
	}

	m.lane.armOutcome(kernel.KindSaveSlot, kernel.KindSavedSlot, func(ev kernel.Event) {
		if ev.Kind != kernel.KindSavedSlot {
			m.lastError = fmt.Errorf("game menu save: kernel refused the slot")

			return
		}

		m.close()
	})
}

// loadGame submits KindLoadSlot{A: slot, B: seed} with the selhero load's
// peek (selhero.go:421-428) and arms the load outcome hook.
func (m *gameMenuScreen) loadGame() {
	if m.lane == nil {
		return
	}

	archive := persist.NewSessionArchive(m.stores, lanePackSource)

	peek, err := archive.LoadSlot(m.slot)
	if err != nil {
		m.lastError = fmt.Errorf("game menu load peek: %w", err)

		return
	}

	cmd := kernel.Command{
		Kind: kernel.KindLoadSlot,
		Tick: m.lane.game.TickCount,
		A:    int32(m.slot),
		B:    int32(peek.LevelSeed),
	}

	if !m.lane.input.SubmitDecision(cmd) {
		return
	}

	m.lane.armOutcome(kernel.KindLoadSlot, kernel.KindLoadedSlot, func(ev kernel.Event) {
		if ev.Kind != kernel.KindLoadedSlot {
			m.lastError = fmt.Errorf("game menu load: kernel refused the slot")

			return
		}

		m.close()
	})
}

// exitToMainMenu clears the lane's hooks, closes the menu and walks the flow
// back to the main menu (X:gamemenu.cpp:106-121 GamemenuNewGame's
// gbRunGame = false arm).
func (m *gameMenuScreen) exitToMainMenu() {
	m.lane.clearOutcomes()
	m.close()
	m.flow.LeaveGame()
}

// quitGame asks the flow to end the boot rc-0 (X:gamemenu.cpp:287-295
// gamemenu_quit_game).
func (m *gameMenuScreen) quitGame() {
	m.close()
	m.flow.RequestQuit()
}
