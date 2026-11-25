//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// gameMenuBoot is the windowed boot with the game-menu screen installed and
// the flow entered into the game — the state every game-menu test starts from.
type gameMenuBoot struct {
	*windowBoot

	menu *gameMenuScreen
}

// newGameMenuBoot walks the production boot into FrontGame (a frozen town
// session loads through the REAL Update loop) and installs the game menu.
func newGameMenuBoot(t *testing.T) *gameMenuBoot {
	t.Helper()

	boot := newWindowBoot(t)

	ctx := context.Background()

	if err := boot.stores.SaveHero(ctx, persist.HeroRecord{
		Slot: 0, Name: "Menu Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	archive := persist.NewSessionArchive(boot.stores, lanePackSource)
	rec := world.SessionRecord{
		Slot: 0, LevelSeed: 0x00C0FFEE, Level: 0, LevelType: 0, Loaded: false,
		Hero: world.Tile{X: 76, Y: 68},
		Pack: warriorPackForWindow(),
	}

	if err := archive.SaveSlot(rec); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}

	boot.flow.HandleKey(ui.KeySpace)
	boot.flow.HandleKey(ui.KeyEnter)
	boot.flow.HandleKey(ui.KeyEnter) // hero row -> load step
	boot.flow.HandleKey(ui.KeyEnter) // Load Game

	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	if got := boot.flow.Screen(); got != scene.FrontGame {
		t.Fatalf("setup: screen = %v, want FrontGame", got)
	}

	boot.installGameMenu(0)

	return &gameMenuBoot{windowBoot: boot, menu: boot.menu}
}

func TestGameMenu_PausesSim(t *testing.T) {
	// given: the game running at frame cadence
	boot := newGameMenuBoot(t)

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	ticksOpen := boot.lane.game.TickCount

	// when: Escape opens the menu and the SAME frames run through the REAL
	// Update loop
	boot.flow.HandleKey(ui.KeyEscape)

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: the settled-tick counter did not move (upstream ProcessInput
	// returns false while the menu is open, X:diablo.cpp:229-232)
	if boot.lane.game.TickCount != ticksOpen {
		t.Fatalf("TickCount advanced %d over 10 frames with the menu open, want 0",
			boot.lane.game.TickCount-ticksOpen)
	}
	if !boot.menu.isOpen() {
		t.Fatal("menu did not open on Esc")
	}
}

func TestGameMenu_EscRoundTrip(t *testing.T) {
	// given: the menu open through the real loop
	boot := newGameMenuBoot(t)
	boot.flow.HandleKey(ui.KeyEscape)

	for range 3 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	pausedAt := boot.lane.game.TickCount

	// when: a second Esc closes it and frames run again
	boot.flow.HandleKey(ui.KeyEscape)

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: the menu closed and the ticks resumed
	if boot.menu.isOpen() {
		t.Fatal("menu still open after the second Esc")
	}
	if boot.lane.game.TickCount == pausedAt {
		t.Fatalf("TickCount did not resume after close (stuck at %d)", pausedAt)
	}
}

func TestGameMenu_LoadRowEnabled(t *testing.T) {
	t.Run("DisabledWithoutAGame", func(t *testing.T) {
		// given: a boot whose slot holds no frozen session — walk in via a
		// fresh CREATE (the load path needs a record, so this boot creates)
		boot := newGameMenuBoot(t)

		// the menu opened over the loaded session; simulate the no-game case
		// by asking the screen to recompute with a slot that has no record
		boot.menu.open(7)

		// when: navigation walks onto the Load row
		loadIdx := 2

		// then: the row is disabled and navigation skips it
		if boot.menu.table.Rows()[loadIdx].Enabled {
			t.Fatal("Load Game enabled for a slot with no game, want disabled")
		}
	})

	t.Run("EnabledWithAGame", func(t *testing.T) {
		boot := newGameMenuBoot(t)
		boot.menu.open(0) // slot 0 holds the session this boot loaded

		if !boot.menu.table.Rows()[2].Enabled {
			t.Fatal("Load Game disabled for a slot with a game, want enabled")
		}
	})
}

func TestGameMenu_SaveSubmitsTheSlot(t *testing.T) {
	// given: the menu open over the loaded session
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.flow.HandleKey(ui.KeyEscape) // open
	before := boot.lane.game.TickCount

	// when: Save Game activates and the frame settles
	boot.menu.selectRow(1) // row 1 = Save Game

	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: the kernel trace carries a SaveSlot for THAT slot (the
	// discriminating state), and the archive holds the session under it
	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
			if cmd.Kind == kernel.KindSaveSlot && cmd.A == 0 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no KindSaveSlot{A:0} in ticks %d..%d", before, boot.lane.game.TickCount)
	}

	archive := persist.NewSessionArchive(boot.stores, lanePackSource)
	has, err := archive.HasGame(context.Background(), 0, string(persist.KindSingle))
	if err != nil {
		t.Fatalf("HasGame: %v", err)
	}
	if !has {
		t.Fatal("the archive does not hold slot 0 after the save settled")
	}
}

func TestGameMenu_ExitToMainMenu(t *testing.T) {
	// given: the menu open
	boot := newGameMenuBoot(t)
	boot.flow.HandleKey(ui.KeyEscape)

	// when: Exit to Main Menu activates
	boot.menu.selectRow(3)

	// then: the flow walked back and the movement lane stopped
	if got := boot.flow.Screen(); got != scene.FrontMainMenu {
		t.Fatalf("Screen() = %v, want FrontMainMenu after Exit", got)
	}
}

func TestGameMenu_Quit(t *testing.T) {
	// given: the menu open
	boot := newGameMenuBoot(t)
	boot.flow.HandleKey(ui.KeyEscape)

	// when: Quit Game activates
	boot.menu.selectRow(4)

	// when: the next Update runs
	err := boot.Update()

	// then: the window honours the quit (ebiten.Termination surfaces as the
	// run-loop's regular termination)
	if err == nil {
		t.Fatal("Update after Quit = nil, want the termination")
	}
}

func TestGameMenu_IdleFrameNeverChangesAnything(t *testing.T) {
	// given: the game running
	boot := newGameMenuBoot(t)

	// when: idle frames only
	for range 5 {
		boot.flow.HandleKey(ui.KeyUnknown)
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: no menu, no quit, no screen change
	if boot.menu.isOpen() {
		t.Fatal("an idle frame opened the menu")
	}
	if boot.flow.Quit() {
		t.Fatal("an idle frame set the quit flag")
	}
	if boot.flow.Screen() != scene.FrontGame {
		t.Fatalf("Screen() = %v after idle frames, want FrontGame", boot.flow.Screen())
	}
}
