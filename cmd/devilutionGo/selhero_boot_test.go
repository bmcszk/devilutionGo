//go:build windowed

package main

import (
	"context"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestBootWiresSelHeroScreen drives the production boot path: runWindowed's
// construction surface must build the persist store AND install the real
// selhero screen on the flow, so a player picking Single Player gets the
// screen — not the Escape-only placeholder (review F1).
func TestBootWiresSelHeroScreen(t *testing.T) {
	boot := newWindowBoot(t)

	// when the flow walks title -> menu -> Single Player
	boot.flow.HandleKey(ui.KeySpace)
	boot.flow.HandleKey(ui.KeyEnter)

	// then the boot built a persist store and installed the real screen:
	// the entry-time refresh already populated the action rows.
	if boot.stores == nil {
		t.Fatalf("window boot built no persist store")
	}

	if got := boot.flow.Screen(); got != scene.FrontSelHero {
		t.Fatalf("screen = %v, want FrontSelHero", got)
	}

	if got := boot.selhero.list.ItemCount(); got != 3 {
		t.Fatalf("entry-time menu items = %d, want the 3 action rows", got)
	}
}

// windowBoot is the production-shape windowed boot under test: the full
// construction path runWindowed uses, over a temp-dir store.
type windowBoot struct {
	*windowGame

	selhero *selHeroScreen
}

// newWindowBoot builds the boot through the production wiring path
// (buildWindowGame's surface), not a test-local copy of it.
func newWindowBoot(t *testing.T) *windowBoot {
	t.Helper()

	w, err := buildWindowGame(t.TempDir())
	if err != nil {
		t.Fatalf("buildWindowGame: %v", err)
	}

	boot := &windowBoot{windowGame: w}
	boot.selhero = w.selhero

	return boot
}

// TestSelHeroListPopulatesOnEntry pins the entry-time refresh: walking into
// the selhero screen must populate the menu from the store — not only after
// a delete (review F2).
func TestSelHeroListPopulatesOnEntry(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	if err := boot.stores.SaveHero(context.Background(), persist.HeroRecord{
		Slot: 0, Name: "Entry Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	// when the flow walks title -> menu -> selhero (keys only)
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// then the screen's list already carries the hero row
	if names := boot.selhero.heroNames(); !slices.Equal(names, []string{"Entry Hero"}) {
		t.Fatalf("hero list on entry = %v, want [Entry Hero]", names)
	}
}

// TestSelHeroFrameRendersHeroRows pins the draw path: with two heroes, the
// composed front frame draws one band row per list entry — heroes first, then
// the actions — instead of the placeholder's hardcoded 3 rows (review F2).
func TestSelHeroFrameRendersHeroRows(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	for slot, name := range []string{"Hero One", "Hero Two"} {
		if err := boot.stores.SaveHero(ctx, persist.HeroRecord{
			Slot: slot, Name: name, ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
		}); err != nil {
			t.Fatalf("SaveHero %s: %v", name, err)
		}
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// the opening walk must leave no armed create hook or queued pack: with
	// heroes stored, Enter on hero row 0 runs startNewGame, so any leftover
	// onCreateEvent/pendingCreate here is latent state, not flow output
	if boot.lane.onCreateEvent != nil || boot.lane.pendingCreate {
		t.Fatalf("opening keys left latent create state (hook=%v pending=%v)",
			boot.lane.onCreateEvent != nil, boot.lane.pendingCreate)
	}

	// the preview panel needs loaded class baselines: open create and confirm
	// the name, which loads the classdat pick list and the warrior preview
	// create white-box (the row 0 key no longer routes to create once heroes
	// exist — startNewGame takes over; create opens only from the Create row)
	_ = boot.selhero.beginCreate()
	_ = boot.selhero.typeName("Preview Hero")
	boot.selhero.confirmName()
	boot.handleKey(ui.KeyEscape) // back out; the preview data stays loaded
	boot.handleKey(ui.KeyEnter)  // re-enter from the main menu

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("selhero refresh: %v", err)
	}

	// when the front frame composes for the selhero screen
	surface := &render.Surface{Width: 640, Height: 480, Pixels: make([]uint8, 640*480)}
	if err := scene.ComposeFrontFrame(surface, boot.flow); err != nil {
		t.Fatalf("ComposeFrontFrame: %v", err)
	}

	// then the band carries 5 list rows: 2 heroes + Create/Delete/Quit
	if rows := countFrontLabelRows(surface, 200, 440); rows != 5 {
		t.Fatalf("front frame list rows = %d, want 5 (2 heroes + 3 actions)", rows)
	}

	// and the stats preview panel is its own draw site: 5 value rows right of
	// the list (selhero.cpp:527-546)
	if rows := countFrontLabelRows(surface, 460, 620); rows != 5 {
		t.Fatalf("preview panel rows = %d, want 5", rows)
	}
}

// countFrontLabelRows counts the distinct label bars in the front band: the
// bar rows a draw site painted between xFrom and xTo.
func countFrontLabelRows(surface *render.Surface, xFrom, xTo int) int {
	rows := 0
	inRow := false

	for y := 96; y < 96+240; y++ {
		painted := false
		row := y * surface.Width
		for x := xFrom; x < xTo; x++ {
			if surface.Pixels[row+x] != 0 {
				painted = true

				break
			}
		}

		if painted && !inRow {
			rows++
		}

		inRow = painted
	}

	return rows
}

// TestSelHeroCreateOpensByKey pins the Enter-on-Create-Hero key path (review
// F8): with the entry refresh populating the real rows, the Create action is
// reachable by navigation + Enter alone — no white-box beginCreate call.
func TestSelHeroCreateOpensByKey(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	if err := boot.stores.SaveHero(context.Background(), persist.HeroRecord{
		Slot: 0, Name: "Existing Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	// when the flow walks into the screen and down to the Create Hero row
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyArrowDown) // hero row -> Create Hero
	boot.handleKey(ui.KeyEnter)

	// then create mode opened and the class pick loaded from the committed
	// classdat rows as soon as the name step confirmed
	if boot.selhero.creating == nil {
		t.Fatalf("Enter on Create Hero did not open the create flow")
	}

	boot.handleKey(ui.KeyEnter) // confirm the empty name: the class pick opens

	if got := boot.selhero.classMenu.ItemCount(); got != 6 {
		t.Fatalf("class list = %d rows, want the 6 committed classes", got)
	}

	if boot.selhero.SelHeroPreviewRows() != 5 {
		t.Fatalf("preview rows = %d, want the 5 stat rows", boot.selhero.SelHeroPreviewRows())
	}
}

// TestSelHeroFocusedDelete pins the delete target with a DISCRIMINATING walk:
// with two heroes, walking UP-wrap from First (through Quit and Delete, which
// never cross a hero row) leaves First the last-highlighted hero, so Enter on
// the Delete action removes FIRST — not the row the offset math would name
// (review F6).
func TestSelHeroFocusedDelete(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	first := persist.HeroRecord{
		Slot: 0, Name: "First Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}
	second := persist.HeroRecord{
		Slot: 1, Name: "Second Hero", ClassID: 1, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := boot.stores.SaveHero(ctx, first); err != nil {
		t.Fatalf("SaveHero first: %v", err)
	}

	if err := boot.stores.SaveHero(ctx, second); err != nil {
		t.Fatalf("SaveHero second: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("selhero refresh: %v", err)
	}

	// the menu: First(0) Second(1) Create(2) Delete(3) Quit(4). Walking
	// UP-wrap from First crosses no hero row, so First stays the
	// last-highlighted hero; then one ArrowDown lands on Delete and Enter
	// removes FOCUSED First (selhero.cpp:139-146: the delete arms on the hero
	// row's focus, not on the action's position).
	boot.handleKey(ui.KeyArrowUp)   // wrap: First -> Quit(4)
	boot.handleKey(ui.KeyArrowUp)   // Quit -> Delete(3)
	boot.handleKey(ui.KeyArrowUp)   // Delete -> Create(2)
	boot.handleKey(ui.KeyArrowDown) // Create -> Delete(3)
	boot.handleKey(ui.KeyEnter)     // Delete: the confirm step arms (dcob D5)
	boot.handleKey(ui.KeyEnter)     // Yes

	rows, err := boot.stores.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 1 || rows[0].Name != "Second Hero" {
		t.Fatalf("store rows after delete = %+v, want only Second Hero (First was focused and deleted)", rows)
	}
}

// TestSelHeroRefusedCreateLeavesNoRow pins the ghost-row fix: a create the
// kernel refuses (an ordinal outside the committed classdat rows; the real
// MPQ asset gate is the recorded deviation) must NOT leave a store row
// (review F5).
func TestSelHeroRefusedCreateLeavesNoRow(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// open create, confirm the name, then force a refusal: the data-built
	// pack source serves every committed class (dcob D7), so the ghost-row
	// pin drives pickClass with an ordinal no committed classdat row has
	boot.handleKey(ui.KeyEnter) // Create Hero

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Ghost Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()

	boot.selhero.pickClass(len(classDirs)) // out of range: the source refuses

	boot.frame()

	events := boot.createdEvents()
	if len(events) != 0 {
		t.Fatalf("CreatedPack events = %v, want none (kernel refuses an unknown class)", events)
	}

	rows, err := boot.stores.ListHeroes(context.Background())
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 0 {
		t.Fatalf("store rows after refused create = %+v, want NO ghost row", rows)
	}
}
