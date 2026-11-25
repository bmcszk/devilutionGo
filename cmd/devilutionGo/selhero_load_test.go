//go:build windowed

package main

import (
	"context"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// seedSavedHero writes the create flow's hero row and freezes a session into
// it through the production archive: the "Save File Exists" state
// (selhero.cpp:217-236 hassaved branch).
func seedSavedHero(t *testing.T, boot *selHeroTestBoot, name string) {
	t.Helper()

	ctx := context.Background()

	if err := boot.store.SaveHero(ctx, persist.HeroRecord{
		Slot: 0, Name: name, ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	archive := persist.NewSessionArchive(boot.store, bootPackSource{})

	// the town session the freeze writes: level 0, Loaded false, the hero on
	// {76,68} — NOT the bootstrap spawn tile (75,68), so a load that restores
	// nothing cannot satisfy the saved-tile assertions
	rec := world.SessionRecord{
		Slot: 0, LevelSeed: loadSeed, Level: 0, LevelType: 0, Loaded: false,
		Hero: world.Tile{X: 76, Y: 68},
		Pack: warriorPack(),
	}

	if err := archive.SaveSlot(rec); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}
}

// TestSelHeroHeroRowWithSavedGameOffersLoadNew drives the hero row's split
// (selhero.cpp:217-236): a hero row with a frozen game opens the Load/New
// list instead of the create flow.
func TestSelHeroHeroRowWithSavedGameOffersLoadNew(t *testing.T) {
	// given the boot over a store whose hero holds a frozen session
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "Saved Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// when the hero row is activated
	boot.handleKey(ui.KeyEnter)

	// then the load step is up: two rows, no create flow
	if got := boot.selhero.loadMenu.ItemCount(); got != 2 {
		t.Fatalf("load rows = %d, want 2 (Load Game / New Game)", got)
	}

	if boot.selhero.creating != nil {
		t.Fatalf("hero row with a saved game opened the create flow, want the load step")
	}

	if boot.selhero.loadStep == nil {
		t.Fatalf("load step did not open on a hero row with a game")
	}
}

// TestSelHeroHeroRowWithoutGameStartsNewGame pins the no-save branch
// (selhero.cpp:239 SelheroLoadSelect(1)): a hero row without a frozen game
// starts a new game for THAT hero, not the create flow.
func TestSelHeroHeroRowWithoutGameStartsNewGame(t *testing.T) {
	// given the boot over a store whose hero has no frozen session
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()

	if err := boot.store.SaveHero(ctx, persist.HeroRecord{
		Slot: 0, Name: "Bare Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// when the hero row is activated
	boot.handleKey(ui.KeyEnter)

	// then the create verb is submitted for the hero's class — not the name
	// step, not the load list
	if !boot.lane.pendingCreate {
		t.Fatalf("hero row without a game did not start a new game (pendingCreate=false)")
	}

	if boot.selhero.creating != nil {
		t.Fatalf("hero row without a game opened the create flow, want the direct start")
	}
}

// TestSelHeroLoadSubmitsKindLoadSlotAndSettles drives the Load row: Enter on
// Load Game peeks the seed and submits KindLoadSlot; the settled pass then
// announces KindLoadedSlot for the slot.
func TestSelHeroLoadSubmitsKindLoadSlotAndSettles(t *testing.T) {
	// given the boot at the load step of a hero with a frozen session
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "Loaded Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter) // hero row -> load step

	// when Load Game is confirmed and the frame settles
	boot.handleKey(ui.KeyEnter)
	boot.frame()

	// then the kernel trace holds the LoadedSlot announcement for slot 0
	loaded := loadedSlotEvents(boot.lane.game.Kernel.Trace())
	if len(loaded) != 1 {
		t.Fatalf("LoadedSlot events = %d (%v), want exactly one", len(loaded), loaded)
	}

	if loaded[0].A != 0 {
		t.Fatalf("LoadedSlot slot = %d, want 0", loaded[0].A)
	}

	// and the world holds the restored hero: the saved tile
	pack, tile, raised := boot.lane.state.RaisedPack()
	if !raised {
		t.Fatalf("load did not raise a hero")
	}

	if pack.Class != 0 {
		t.Fatalf("restored pack class = %d, want 0 (warrior)", pack.Class)
	}

	// the saved tile is {76,68}, not the bootstrap spawn (75,68): only the
	// record's restore can put the hero there (Loaded=false restores directly,
	// state.go restoreSession — no relocation search runs).
	want := world.Tile{X: 76, Y: 68}
	if tile != want {
		t.Fatalf("restored tile = %v, want the saved %v", tile, want)
	}
}

// TestSelHeroLoadEscapeWalksBack pins the load step's cancel: Escape returns
// to the hero list (upstream UiFocusNavigationEsc, selhero.cpp:232).
func TestSelHeroLoadEscapeWalksBack(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "Esc Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter) // hero row -> load step
	boot.handleKey(ui.KeyEscape)

	if boot.selhero.loadStep != nil {
		t.Fatalf("Escape did not leave the load step")
	}

	if got := boot.selhero.list.FocusedIndex(); got != 0 {
		t.Fatalf("focus after escape = %d, want the hero row (0)", got)
	}
}

// TestSelHeroNewGameRowStartsTheGame pins the New Game row (selhero.cpp:368-391
// value 1): a new game for the EXISTING hero submits the create verb with the
// hero's own class.
func TestSelHeroNewGameRowStartsTheGame(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "New Game Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter) // hero row -> load step
	boot.handleKey(ui.KeyArrowDown)
	boot.handleKey(ui.KeyEnter) // New Game

	if !boot.lane.pendingCreate {
		t.Fatalf("New Game did not submit the create verb for the existing hero")
	}

	if boot.selhero.creating != nil {
		t.Fatalf("New Game opened the create flow, want the direct start")
	}
}

// loadRowsNames reports the load step's row labels: the test reads what the
// player sees.
func loadRowsNames(boot *selHeroTestBoot) []string {
	rows := make([]string, 0, boot.selhero.loadMenu.ItemCount())
	for i := range boot.selhero.loadMenu.ItemCount() {
		rows = append(rows, boot.selhero.loadRows[i])
	}

	return rows
}

// loadRowsIsLoadNew pins the row labels: upstream order, Load Game first.
func TestSelHeroLoadRowsAreLoadNew(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "Rows Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter)

	if got := loadRowsNames(boot); !slices.Equal(got, []string{"Load Game", "New Game"}) {
		t.Fatalf("load rows = %v, want [Load Game New Game]", got)
	}
}

// loadSeed is the level seed the test freezes: a value the zero blob cannot
// carry, so the peek path is exercised honestly.
const loadSeed = uint32(0x00C0FFEE)

// warriorPack is the starting pack the boot source serves: the scripted
// warrior's frozen pack (device_lane.go bootPackSource).
func warriorPack() world.StartingPack {
	pack, _ := bootPackSource{}.StartingPack(0)

	return pack
}

// loadedSlotEvents filters a trace down to its LoadedSlot events.
func loadedSlotEvents(trace []kernel.Event) []kernel.Event {
	out := make([]kernel.Event, 0, 1)

	for _, ev := range trace {
		if ev.Kind == kernel.KindLoadedSlot {
			out = append(out, ev)
		}
	}

	return out
}

// TestSelHeroLoadStepRendersTwoRows pins the load step's draw site (X3
// precedent): at the load step the composed front frame draws the two
// Load/New rows — NOT the hero/action list.
func TestSelHeroLoadStepRendersTwoRows(t *testing.T) {
	// given the boot at the load step of a hero with a frozen session
	boot := newSelHeroTestBoot(t)
	seedSavedHero(t, boot, "Render Hero")

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEnter) // hero row -> load step

	// when the front frame composes
	surface := &render.Surface{Width: 640, Height: 480, Pixels: make([]uint8, 640*480)}
	if err := scene.ComposeFrontFrame(surface, boot.flow); err != nil {
		t.Fatalf("ComposeFrontFrame: %v", err)
	}

	// then the band carries the 2 load rows
	if rows := countFrontLabelRows(surface, 200, 440); rows != 2 {
		t.Fatalf("load-step band rows = %d, want 2 (Load Game / New Game)", rows)
	}
}
