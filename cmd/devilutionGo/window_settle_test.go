//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// TestWindowUpdateSettlesTheCreatedHero drives (*windowGame).Update — the REAL
// window loop — through a create: after the pick and N Updates, the store row
// EXISTS (dcob D10 / RF1). The front-flow screens must not starve the lane's
// frame, or the pending create never settles.
func TestWindowUpdateSettlesTheCreatedHero(t *testing.T) {
	boot := newWindowBoot(t)

	// walk title -> menu -> selhero and open the create flow
	boot.flow.HandleKey(ui.KeySpace)
	boot.flow.HandleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Settled Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()
	boot.selhero.classMenu.HandleKey(ui.KeyEnter) // pick warrior

	// drive the REAL Update loop: the create settles within N frames
	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	rows, err := boot.stores.ListHeroes(context.Background())
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 1 || rows[0].Name != "Settled Hero" {
		t.Fatalf("store rows after create + %d Updates = %v, want 1 (Settled Hero)", settleFrames, rows)
	}
}

// TestWindowUpdateLoadEntersTheGame drives the REAL window loop through a
// load: at the load step, Enter on Load Game settles and the flow leaves the
// front screens into FrontGame with the restored hero (S8; upstream
// SELHERO_CONTINUE -> StartGame, menu.cpp:69/:130).
func TestWindowUpdateLoadEntersTheGame(t *testing.T) {
	// given the boot over a store whose hero holds a frozen session
	boot := newWindowBoot(t)

	ctx := context.Background()

	if err := boot.stores.SaveHero(ctx, persist.HeroRecord{
		Slot: 0, Name: "Loaded Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	archive := persist.NewSessionArchive(boot.stores, lanePackSource)

	// the record freezes the hero on {76,68}, NOT the bootstrap spawn tile
	// (75,68): the bare grid only checks bounds, so the tile is walkable and
	// only the record's restore can put the hero there.
	rec := world.SessionRecord{
		Slot: 0, LevelSeed: 0x00C0FFEE, Level: 0, LevelType: 0, Loaded: false,
		Hero: world.Tile{X: 76, Y: 68},
		Pack: warriorPackForWindow(),
	}

	if err := archive.SaveSlot(rec); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}

	// when the flow walks title -> menu -> selhero -> (hero row) load step
	// -> Load Game, and the REAL Update loop settles the load
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

	// then the flow entered the game and the world holds the restored hero
	if got := boot.flow.Screen(); got != scene.FrontGame {
		t.Fatalf("screen after load = %v, want scene.FrontGame", got)
	}

	if _, _, raised := boot.lane.state.RaisedPack(); !raised {
		t.Fatalf("the load did not raise the hero into the world")
	}

	// and the restored hero sits on the record's saved tile: {76,68} can only
	// come from the record — the bootstrap raise puts the hero on (75,68), and
	// the town record (Loaded=false) restores the hero directly with no
	// SyncInitPlrPos search (state.go restoreSession) that could relocate it.
	if got := boot.lane.state.Hero(); got != (world.Tile{X: 76, Y: 68}) {
		t.Fatalf("restored hero tile = %v, want the record's {76,68}", got)
	}
}

// warriorPackForWindow is the pack the lane's classdat source serves for the
// warrior: the value the save froze and the restore must put back.
func warriorPackForWindow() world.StartingPack {
	pack, _ := lanePackSource.StartingPack(0)

	return pack
}
