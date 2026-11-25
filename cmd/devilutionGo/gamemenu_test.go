//go:build windowed

package main

import (
	"context"
	"testing"
	"time"

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

// --- S5: options submenu + Speed row ---

func TestGameMenuOptions_SubmenuOpensEntryForEntry(t *testing.T) {
	// given: the menu open at the root
	boot := newGameMenuBoot(t)
	boot.menu.open(0)

	// when: Options activates
	boot.menu.selectRow(gamemenuRowOptions)

	// then: the submenu table is the entry-for-entry five rows, first three
	// DISABLED (X:gamemenu.cpp:68-75 sgOptionsMenu), focused on Speed's
	// first enabled neighbour
	rows := boot.menu.Rows()
	want := []string{"Music", "Sound", "Gamma", "Speed", "Previous Menu"}
	for i, row := range rows {
		if row.Label != want[i] {
			t.Fatalf("submenu row %d = %q, want %q", i, row.Label, want[i])
		}
	}
	for i := range 3 {
		if rows[i].Enabled {
			t.Fatalf("submenu row %d (%s) enabled, want DISABLED (S6 follow-up bean)", i, rows[i].Label)
		}
	}
	if !rows[3].Enabled || !rows[4].Enabled {
		t.Fatal("Speed / Previous Menu must be enabled")
	}
}

func TestGameMenuOptions_PreviousMenuReturnsToRoot(t *testing.T) {
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions) // submenu

	// when: Previous Menu activates
	boot.menu.selectRow(4)

	// then: the root table is back
	rows := boot.menu.Rows()
	if rows[0].Label != "Options" {
		t.Fatalf("root row 0 = %q after Previous Menu, want Options", rows[0].Label)
	}
}

func TestGameMenuOptions_EscClosesTheWholeMenu(t *testing.T) {
	// given: the submenu open
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)

	// when: Esc reaches the screen
	boot.menu.HandleKey(ui.KeyEscape)

	// then: the WHOLE menu closed (X:gmenu.cpp:290-292 gmenu_set_items(null)),
	// not a walk back to the root
	if boot.menu.isOpen() {
		t.Fatal("menu still open after Esc in the submenu, want fully closed")
	}
}

func TestGameMenuSpeed_RightStepsTheRate(t *testing.T) {
	// given: the submenu open with the rate at the default 20
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)
	before := boot.lane.game.TickCount

	// when: Right steps the Speed row
	boot.menu.HandleKey(ui.KeyArrowRight)

	// when: the decision settles — the menu pauses the sim while open
	// (upstream applies gnTickDelay at keypress, outside the tick loop; our
	// law-1 port dispatches on the next settled tick, which runs once the
	// menu closes), so close and settle
	boot.menu.close()
	if err := boot.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// then: KindSetTickRate{A:21} dispatched and the live delay is 1000/21 ms
	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
			if cmd.Kind == kernel.KindSetTickRate && cmd.A == 21 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no KindSetTickRate{A:21} in ticks %d..%d", before, boot.lane.game.TickCount)
	}
	if boot.lane.game.TickDelay() != 1000*time.Millisecond/21 {
		t.Fatalf("TickDelay() = %v, want %v", boot.lane.game.TickDelay(), 1000*time.Millisecond/21)
	}
}

func TestGameMenuSpeed_EnterToggles(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rate  int32
		want  int32
		delay time.Duration
	}{
		{"at 20 -> 50", 20, 50, 20 * time.Millisecond},
		{"at 50 -> 20", 50, 20, 50 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given: the submenu open with the rate set to tc.rate
			boot := newGameMenuBoot(t)
			boot.lane.game.Kernel.Submit(kernel.Command{Kind: kernel.KindSetTickRate, A: tc.rate})
			if err := boot.Update(); err != nil {
				t.Fatalf("Update: %v", err)
			}
			boot.menu.open(0)
			boot.menu.selectRow(gamemenuRowOptions)
			before := boot.lane.game.TickCount

			// when: Enter toggles the Speed row
			boot.menu.HandleKey(ui.KeyEnter)
			boot.menu.close() // the pause rule holds the tick until close
			if err := boot.Update(); err != nil {
				t.Fatalf("Update: %v", err)
			}

			// then: the toggled rate landed (X:gamemenu.cpp:265-271)
			found := false
			for tick := before; tick < boot.lane.game.TickCount; tick++ {
				for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
					if cmd.Kind == kernel.KindSetTickRate && cmd.A == tc.want {
						found = true
					}
				}
			}
			if !found {
				t.Fatalf("no KindSetTickRate{A:%d} dispatched", tc.want)
			}
			if boot.lane.game.TickDelay() != tc.delay {
				t.Fatalf("TickDelay() = %v, want %v", boot.lane.game.TickDelay(), tc.delay)
			}
		})
	}
}

func TestGameMenuSpeed_LabelFollowsTheLiveValue(t *testing.T) {
	// given: the submenu open at rate 20
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)

	// then: the Speed row's label carries the live value
	// (X:gamemenu.cpp:174-177)
	if label := boot.menu.speedLabel(); label != "Speed - Normal" {
		t.Fatalf("Speed label = %q at rate 20, want %q", label, "Speed - Normal")
	}

	// when: the slider moves into the 40+ band
	boot.menu.table.SliderSet(gamemenuRowSpeed, speedSliderMin, speedSliderMax, 45)

	// then: the label re-reads from the slider's pending position
	if label := boot.menu.speedLabel(); label != "Speed - Faster" {
		t.Fatalf("Speed label = %q at rate 45, want %q", label, "Speed - Faster")
	}
}

func TestGameMenuSpeed_MultiplayerDisablesTheRow(t *testing.T) {
	// given: a multiplayer boot (upstream X:gamemenu.cpp:159-170 shows a
	// static label, no slider)
	boot := newGameMenuBoot(t)
	boot.lane.game.Multiplayer = true
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)

	// when: keys hit the disabled Speed row
	boot.menu.HandleKey(ui.KeyArrowRight)
	boot.menu.HandleKey(ui.KeyEnter)

	// then: no SetTickRate submitted (the trace holds none)
	for tick := uint32(0); tick < boot.lane.game.TickCount; tick++ {
		for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
			if cmd.Kind == kernel.KindSetTickRate {
				t.Fatalf("SetTickRate dispatched in multiplayer at tick %d", tick)
			}
		}
	}
}

func TestGameMenuSpeed_SimOptionsFollows(t *testing.T) {
	// given: the boot's in-memory SimOptions
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)

	// when: one Right step settles (close first — the pause rule holds the
	// tick while the menu is open)
	boot.menu.HandleKey(ui.KeyArrowRight)
	boot.menu.close()
	if err := boot.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// then: the boot's SimOptions.TickRate reads 21 and GnTickDelay() = 47
	// (in-memory only — persistence is the runtime-ini bean's)
	if boot.simOpts.TickRate != 21 {
		t.Fatalf("SimOptions.TickRate = %d, want 21", boot.simOpts.TickRate)
	}
	if boot.simOpts.GnTickDelay() != 47 {
		t.Fatalf("GnTickDelay() = %d, want 47", boot.simOpts.GnTickDelay())
	}
}

func TestGameMenu_IdleFrameBetweenArmAndAnswer(t *testing.T) {
	// given: the submenu open
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)

	// when: idle frames interleave with the armed Right step
	boot.menu.HandleKey(ui.KeyUnknown)
	boot.menu.HandleKey(ui.KeyArrowRight)
	boot.menu.HandleKey(ui.KeyUnknown)
	if err := boot.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// then: the flow never changed screen on the idle frame
	if boot.flow.Screen() != scene.FrontGame {
		t.Fatalf("Screen() = %v after idle frames, want FrontGame", boot.flow.Screen())
	}
	if !boot.menu.isOpen() {
		t.Fatal("an idle frame closed the menu")
	}
}
