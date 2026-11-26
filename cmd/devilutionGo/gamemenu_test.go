//go:build windowed

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
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

	// a second HERO row exists in slot 1 with NO game behind it: the save
	// proof needs a slot the fixture does not pre-seed (review F2) — a hero
	// row is not a game (SessionArchive.HasGame reads the game entry only)
	if err := boot.stores.SaveHero(ctx, persist.HeroRecord{
		Slot: 1, Name: "Second Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero slot 1: %v", err)
	}

	return &gameMenuBoot{windowBoot: boot, menu: boot.menu}
}

func TestGameMenu_PausesSim(t *testing.T) {
	// given: the game running at frame cadence, with a world-tick probe
	boot := newGameMenuBoot(t)
	worldTicks := 0
	boot.lane.game.OnTick = func() { worldTicks++ }

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// when: Escape opens the menu and the SAME frames run through the REAL
	// Update loop
	boot.flow.HandleKey(ui.KeyEscape)

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: NO world tick ran while the menu was open
	if worldTicks != 10 {
		t.Fatalf("world ticks = %d after 10 open frames, want 10 (the pre-open cadence only)", worldTicks)
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

func TestGameMenu_WorldKeepsRunningInMultiplayer(t *testing.T) {
	// given: a multiplayer boot (upstream does not pause there: ProcessInput
	// returns true with the menu open, X:diablo.cpp:231) with a world probe
	boot := newGameMenuBoot(t)
	boot.lane.game.Multiplayer = true

	worldTicks := 0
	boot.lane.game.OnTick = func() { worldTicks++ }

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	beforeTicks, beforeWorld := boot.lane.game.TickCount, worldTicks

	// when: Escape opens the menu and ten frames run through the REAL Update
	// loop — the menu-open WINDOW frame must not impose its own pause
	boot.flow.HandleKey(ui.KeyEscape)

	for range 10 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: the world kept running while the menu was open (review F11: the
	// menu-open frame used to call framePaused unconditionally and froze MP)
	if !boot.menu.isOpen() {
		t.Fatal("menu did not open on Esc")
	}

	if got := boot.lane.game.TickCount - beforeTicks; got != 10 {
		t.Fatalf("TickCount advanced %d over 10 open MP frames, want 10", got)
	}

	if got := worldTicks - beforeWorld; got != 10 {
		t.Fatalf("world ticks = %d over 10 open MP frames, want 10", got)
	}
}

func TestGameMenu_LoadRowEnabled(t *testing.T) {
	t.Run("DisabledWithoutAGame", func(t *testing.T) {
		// given: a boot whose focused slot holds no frozen session, computed
		// by reopening the screen over slot 7 (the fixture's slot 0 has one)
		boot := newGameMenuBoot(t)
		boot.menu.open(7)

		// then: the Load row is disabled
		if boot.menu.table.Rows()[gamemenuRowLoad].Enabled {
			t.Fatal("Load Game enabled for a slot with no game, want disabled")
		}

		// and: navigation from focus 0 skips the disabled row (AC6's skip leg)
		boot.menu.HandleKey(ui.KeyArrowDown)
		boot.menu.HandleKey(ui.KeyArrowDown)
		if got := boot.menu.Focused(); got != 3 {
			t.Fatalf("Focused after two Downs = %d, want 3 (disabled row 2 skipped)", got)
		}

		// and: Enter on the disabled row is inert (AC6's inert leg) — scan the
		// trace from the walk-in, excluding the boot's own load at tick 0
		boot.menu.selectRow(gamemenuRowLoad)
		for tick := uint32(1); tick < boot.lane.game.TickCount; tick++ {
			for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
				if cmd.Kind == kernel.KindLoadSlot {
					t.Fatalf("KindLoadSlot dispatched at tick %d from a disabled row", tick)
				}
			}
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
	// given: the menu OPEN over the loaded session, installed for a NON-ZERO
	// slot the fixture does NOT pre-seed (slot 1 carries a hero row but no
	// game), carrying the seed threaded from the walk (review F2/F3)
	boot := newGameMenuBoot(t)
	boot.menu.slot = 1
	boot.menu.seed = 0x00C0FFEE
	boot.flow.HandleKey(ui.KeyEscape) // open
	if !boot.menu.isOpen() {
		t.Fatal("setup: menu not open before the Save row activates")
	}

	archive := persist.NewSessionArchive(boot.stores, lanePackSource)
	has, err := archive.HasGame(context.Background(), 1, string(persist.KindSingle))
	if err != nil {
		t.Fatalf("HasGame before: %v", err)
	}
	if has {
		t.Fatal("setup: slot 1 already holds a game before the save")
	}
	before := boot.lane.game.TickCount

	// when: Save Game activates (upstream gamemenu_off precedes SaveGame,
	// X:gamemenu.cpp:331-336) and the now-running frames settle
	boot.menu.selectRow(gamemenuRowSave)

	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: the menu closed on activation and the kernel trace carries
	// SavedSlot{A:1} with the seed word — the discriminating state
	if boot.menu.isOpen() {
		t.Fatal("menu still open after Save Game activated")
	}
	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, ev := range boot.lane.game.Kernel.TraceEventsAt(tick) {
			if ev.Kind == kernel.KindSavedSlot && ev.A == 1 && ev.C == int32(0x00C0FFEE) {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no KindSavedSlot{A:1, C:0x00C0FFEE} in ticks %d..%d", before, boot.lane.game.TickCount)
	}

	// and the archive NOW holds slot 1 — the save wrote it, not the fixture
	has, err = archive.HasGame(context.Background(), 1, string(persist.KindSingle))
	if err != nil {
		t.Fatalf("HasGame after: %v", err)
	}
	if !has {
		t.Fatal("the archive does not hold slot 1 after the save settled")
	}
}

func TestGameMenu_ThreadedSlotAndSeedPreserveTheSave(t *testing.T) {
	// review F3 consequence: the walk in threads the selhero slot and the
	// peeked seed into the menu; a save keeps both (hero in slot 0 here, the
	// threaded values are what the install carried)
	boot := newGameMenuBoot(t)
	if boot.menu.slot != 0 || boot.menu.seed != 0x00C0FFEE {
		t.Fatalf("threaded slot/seed = (%d, %#x), want (0, 0xc0ffee)", boot.menu.slot, boot.menu.seed)
	}

	boot.flow.HandleKey(ui.KeyEscape)
	before := boot.lane.game.TickCount
	boot.menu.selectRow(gamemenuRowSave)

	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, ev := range boot.lane.game.Kernel.TraceEventsAt(tick) {
			if ev.Kind == kernel.KindSavedSlot && ev.A == 0 && ev.C == int32(0x00C0FFEE) {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no KindSavedSlot{A:0, C:0xc0ffee}: the threaded seed did not survive the save")
	}

	rec, err := persist.NewSessionArchive(boot.stores, lanePackSource).LoadSlot(0)
	if err != nil {
		t.Fatalf("LoadSlot after save: %v", err)
	}
	if rec.LevelSeed != 0x00C0FFEE {
		t.Fatalf("saved record seed = %#x, want 0xc0ffee (seed tail preserved)", rec.LevelSeed)
	}
}

func TestGameMenu_LoadSubmitsAndSettles(t *testing.T) {
	// the Load half of AC7: activating Load closes the menu (upstream
	// gamemenu_off-first), the command dispatches on the running tick and
	// the world restores — LoadedSlot announces the slot
	boot := newGameMenuBoot(t)
	boot.flow.HandleKey(ui.KeyEscape)
	if !boot.menu.table.Rows()[gamemenuRowLoad].Enabled {
		t.Fatal("setup: Load Game disabled for the slot the boot loaded")
	}
	before := boot.lane.game.TickCount

	boot.menu.selectRow(gamemenuRowLoad)

	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	if boot.menu.isOpen() {
		t.Fatal("menu still open after Load Game activated")
	}
	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, ev := range boot.lane.game.Kernel.TraceEventsAt(tick) {
			if ev.Kind == kernel.KindLoadedSlot && ev.A == 0 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no KindLoadedSlot{A:0} in ticks %d..%d", before, boot.lane.game.TickCount)
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

func TestGameMenu_DoubleOpenStaysOpen(t *testing.T) {
	// upstream twin: DoubleOpen_StillOpen (X:test/game_menu_test.cpp:128-135)
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.open(0)

	if !boot.menu.isOpen() {
		t.Fatal("a second open deactivated the menu, want still open")
	}
}

func TestGameMenu_MenuStateConsistency(t *testing.T) {
	// upstream twin: MenuStateConsistency (X:test/game_menu_test.cpp:137-157):
	// gmenu_is_active mirrors sgpCurrentMenu being set. DEVIATION (owned):
	// the Go screen keeps its root table alive while closed (the render gate
	// is GameMenuOpen), so the closed leg asserts only the active flag and
	// the submenu state upstream's gmenu_set_items(nullptr) clears.
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	if !boot.menu.isOpen() || boot.menu.table == nil {
		t.Fatal("open menu: active/table mismatch")
	}

	boot.menu.close()
	if boot.menu.isOpen() || boot.menu.inSubmenu {
		t.Fatal("closed menu: active or submenu state left set")
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

// --- S1: the menu close persists diablo.ini ---

// TestGameMenu_ClosePersistsTheSettingsDocument pins the save hook: closing the
// WHOLE menu calls SaveOptions upstream (gmenu_set_items(nullptr),
// X:gmenu.cpp:246-248), so the boot writes its settings document on that event
// and NOT on a walk back to the root table (the menu is still up).
func TestGameMenu_ClosePersistsTheSettingsDocument(t *testing.T) {
	// given a boot whose settings dir is a fresh temp dir
	settingsDir := t.TempDir()

	boot, err := buildWindowGame(t.TempDir(), settingsDir)
	if err != nil {
		t.Fatalf("buildWindowGame: %v", err)
	}

	iniPath := filepath.Join(settingsDir, engine.SettingsIniName)

	// when the menu opens and walks into the submenu, then back to the root
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)
	boot.menu.selectRow(gamemenuRowPrevious)

	// then nothing was written yet: the menu never closed
	if _, err := os.Stat(iniPath); !os.IsNotExist(err) {
		t.Fatalf("stat after submenu walk = %v, want no written ini", err)
	}

	// when Esc closes the whole menu
	boot.menu.HandleKey(ui.KeyEscape)
	if boot.menu.isOpen() {
		t.Fatal("menu still open after Esc")
	}

	// then the document landed carrying the session's rate and the upstream
	// key names (Source/options.cpp:514-516, :796, :843)
	data, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatalf("read written ini: %v", err)
	}

	want := []string{
		"[Audio]", "Sound Volume=0", "Music Volume=0", "Sample Rate=22050",
		"[Game]", "Speed=20", "[Graphics]", "Brightness Correction=0",
	}
	for _, key := range want {
		if !strings.Contains(string(data), key) {
			t.Fatalf("written ini missing %q:\n%s", key, data)
		}
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

	// when: Right steps the Speed row; frames run with the menu OPEN —
	// shape (b1) keeps the kernel pass running under the pause, so the
	// rate applies without closing (review F1 fix)
	boot.menu.HandleKey(ui.KeyArrowRight)
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

			// when: Enter toggles the Speed row (shape b1: the rate applies
			// while the menu stays open)
			boot.menu.HandleKey(ui.KeyEnter)
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

func TestGameMenuSpeed_AppliesWhileTheMenuIsOpen(t *testing.T) {
	// review F1, shape (b1): the kernel pass runs while the menu is open
	// (only the world/sim step is gated), so the rate lands on a tick that
	// actually runs — the label truthfully names the APPLIED rate
	boot := newGameMenuBoot(t)
	boot.menu.open(0)
	boot.menu.selectRow(gamemenuRowOptions)
	before := boot.lane.game.TickCount

	// when: Right steps the Speed row and frames run with the menu OPEN
	boot.menu.HandleKey(ui.KeyArrowRight)
	for range 3 {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	// then: KindSetTickRate{A:21} dispatched WITHOUT closing the menu and
	// the live delay followed
	if !boot.menu.isOpen() {
		t.Fatal("the frames closed the menu; the rate must apply while open (shape b1)")
	}
	found := false
	for tick := before; tick < boot.lane.game.TickCount; tick++ {
		for _, cmd := range boot.lane.game.Kernel.TraceCommandsAt(tick) {
			if cmd.Kind == kernel.KindSetTickRate && cmd.A == 21 {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no KindSetTickRate{A:21} dispatched in ticks %d..%d with the menu open", before, boot.lane.game.TickCount)
	}
	if boot.lane.game.TickDelay() != 1000*time.Millisecond/21 {
		t.Fatalf("TickDelay() = %v with the menu open, want %v", boot.lane.game.TickDelay(), 1000*time.Millisecond/21)
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

	// when: one Right step settles (shape b1: the pass runs while open —
	// no close needed)
	boot.menu.HandleKey(ui.KeyArrowRight)
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
