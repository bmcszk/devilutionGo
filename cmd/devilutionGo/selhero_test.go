//go:build windowed

package main

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// committedClassTable reads one committed per-class class table (the
// internal/testdata corpus) or fails the test: the fixture pipeline law keeps
// the bytes in the tree, so a missing file FAILS rather than skips.
func committedClassTable(t *testing.T, class, name string) *enginetxtdata.Table {
	t.Helper()

	data, err := scene.ReadCommittedFixture(class, name)
	if err != nil {
		t.Fatalf("read committed %s/%s: %v", class, name, err)
	}

	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		t.Fatalf("parse %s/%s: %v", class, name, err)
	}

	return tbl
}

// TestSelHeroListFromStore drives the front boot's selhero screen over a real
// SQLite store: the hero list the screen exposes mirrors persist.ListHeroes
// slot for slot and name for name.
func TestSelHeroListFromStore(t *testing.T) {
	// given the front boot over a store with one warrior on slot 0 and one
	// rogue on slot 1
	boot := newSelHeroTestBoot(t)
	store := boot.store

	ctx := context.Background()
	warrior := persist.HeroRecord{
		Slot: 0, Name: "Warrior One", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}
	rogue := persist.HeroRecord{
		Slot: 1, Name: "Rogue One", ClassID: 1, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := store.SaveHero(ctx, warrior); err != nil {
		t.Fatalf("SaveHero warrior: %v", err)
	}

	if err := store.SaveHero(ctx, rogue); err != nil {
		t.Fatalf("SaveHero rogue: %v", err)
	}

	// when the flow walks title -> menu -> selhero and the screen loads its
	// list from the store
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if got := boot.flow.Screen(); got != scene.FrontSelHero {
		t.Fatalf("screen = %v, want FrontSelHero", got)
	}

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("selhero refresh: %v", err)
	}

	// then the screen's hero list carries both rows in slot order
	names := boot.selhero.heroNames()
	if !slices.Equal(names, []string{"Warrior One", "Rogue One"}) {
		t.Fatalf("hero list = %v, want [Warrior One Rogue One]", names)
	}

	// and the rows match what the store itself reports
	rows, err := store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 2 || rows[0].Name != "Warrior One" || rows[1].Name != "Rogue One" {
		t.Fatalf("store rows = %+v, want the two saved heroes", rows)
	}
}

// TestSelHeroCreateSubmitsCreatePack drives the create flow: name entry, class
// pick, then the create select submits KindCreatePack through the LIVE kernel
// and the next settled tick's log carries KindCreatedPack.
func TestSelHeroCreateSubmitsCreatePack(t *testing.T) {
	// given the front boot on the selhero screen, in create mode at the class
	// pick step, focused on the warrior class
	boot := newSelHeroTestBoot(t)
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if got := boot.flow.Screen(); got != scene.FrontSelHero {
		t.Fatalf("screen = %v, want FrontSelHero", got)
	}

	// when the user opens the create flow, enters a name and picks a class
	boot.handleKey(ui.KeyEnter) // Create Hero

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Brave Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()

	// then the class preview panel reads the committed warrior attributes
	// before the create fires, so the player sees what they raise
	attrs, err := boot.selhero.classPreview(0)
	if err != nil {
		t.Fatalf("classPreview warrior: %v", err)
	}

	if attrs.BaseStr != 30 || attrs.BaseVit != 25 {
		t.Fatalf("warrior preview = %+v, want baseStr 30 baseVit 25", attrs)
	}

	// when the class is confirmed the screen submits KindCreatePack through
	// the live kernel and the frame's settled tick passes it
	boot.handleKey(ui.KeyEnter) // Warrior

	boot.frame()

	// then the tick log carries CreatedPack for the warrior class
	events := boot.createdEvents()
	if len(events) != 1 {
		t.Fatalf("CreatedPack events = %d (%v), want exactly one", len(events), events)
	}

	if events[0].A != 0 {
		t.Fatalf("CreatedPack class = %d, want 0 (warrior)", events[0].A)
	}

	// and the persisted hero row exists under the entered name
	rows, err := boot.store.ListHeroes(context.Background())
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 1 || rows[0].Name != "Brave Hero" || rows[0].ClassID != 0 {
		t.Fatalf("store rows = %+v, want one warrior named Brave Hero", rows)
	}
}

// TestSelHeroDeleteHitsDeleteHero drives the delete path: a seeded hero row is
// deleted through the screen and the store no longer reports it.
func TestSelHeroDeleteHitsDeleteHero(t *testing.T) {
	// given the front boot over a store with one hero on slot 0
	boot := newSelHeroTestBoot(t)

	ctx := context.Background()
	hero := persist.HeroRecord{
		Slot: 0, Name: "Doomed Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := boot.store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.refresh(ctx); err != nil {
		t.Fatalf("selhero refresh: %v", err)
	}

	// when the user walks down to the Delete Hero action row, activates it and
	// answers the Yes/No confirm (dcob D5): the menu is hero row (0), Create
	// (1), Delete (2), Quit (3)
	boot.handleKey(ui.KeyArrowDown)
	boot.handleKey(ui.KeyArrowDown) // Delete Hero
	boot.handleKey(ui.KeyEnter)     // the confirm step arms
	boot.handleKey(ui.KeyEnter)     // Yes

	// then the store row is gone
	rows, err := boot.store.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 0 {
		t.Fatalf("store rows = %+v, want the hero deleted", rows)
	}
}

// TestSelHeroEscapeWalksBack pins the screen's back-walk: Escape on the
// selhero screen returns to the main menu.
func TestSelHeroEscapeWalksBack(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)
	boot.handleKey(ui.KeyEscape)

	if got := boot.flow.Screen(); got != scene.FrontMainMenu {
		t.Fatalf("screen = %v, want FrontMainMenu", got)
	}
}

// TestSelHeroClassPreviewFromCommittedLoaders pins the preview's data source:
// the warrior preview reads the committed attributes.tsv through the typed
// loader, so the panel shows the class's real stat baselines.
func TestSelHeroClassPreviewFromCommittedLoaders(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	attrs, err := boot.selhero.classPreview(0)
	if err != nil {
		t.Fatalf("classPreview: %v", err)
	}

	// the committed warrior row's own values
	want, _, err := txtdata.LoadClassAttributes(
		committedClassTable(t, "warrior", "attributes.tsv"))
	if err != nil {
		t.Fatalf("LoadClassAttributes: %v", err)
	}

	if attrs != want {
		t.Fatalf("preview attrs = %+v, want the loader's %+v", attrs, want)
	}
}

// selHeroTestBoot is the windowless selhero boot: the front flow over a real
// SQLite store in a temp dir and the live kernel. No mocks beyond the tick
// source the frameClock already drives.
type selHeroTestBoot struct {
	*frontBoot

	store   *persist.SQLiteStore
	selhero *selHeroScreen
}

// newSelHeroTestBoot wires the selhero boot the windowed run uses: the front
// flow, a temp-dir store and the device lane's live kernel, with the selhero
// screen consuming both.
func newSelHeroTestBoot(t *testing.T) *selHeroTestBoot {
	t.Helper()

	store, err := persist.OpenSQLiteStore(
		context.Background(), filepath.Join(t.TempDir(), "selhero.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	boot := &selHeroTestBoot{frontBoot: newFrontBoot(), store: store}
	// The screen's Load path needs the save/load verbs on the lane's kernel:
	// the same attach the windowed boot does (buildWindowGame's surface).
	boot.lane.attachSaveLoad(
		persist.NewSessionArchive(store, bootPackSource{}), refusingLevelSource{})
	boot.selhero = newSelHeroScreen(store, boot.lane, boot.flow.RequestBackWalk)
	boot.flow.SetSelHero(boot.selhero, boot.selhero, func() {
		_ = boot.selhero.refresh(context.Background())
	})

	return boot
}

// frame runs one window frame: read input, then the settled tick and its
// kernel pass — the same order windowGame.Update drives.
func (b *selHeroTestBoot) frame() {
	b.lane.frame(func() {})
}

// createdEvents reports the CreatedPack events the lane's kernel trace holds.
func (b *selHeroTestBoot) createdEvents() []kernel.Event {
	return createdPackEvents(b.lane.game.Kernel.Trace())
}

// createdPackEvents filters a trace down to its CreatedPack events.
func createdPackEvents(trace []kernel.Event) []kernel.Event {
	out := make([]kernel.Event, 0, 1)

	for _, ev := range trace {
		if ev.Kind == kernel.KindCreatedPack {
			out = append(out, ev)
		}
	}

	return out
}
