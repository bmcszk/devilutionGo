package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

func TestRegisterMovement_StepMarksTheTileExploredAndEmitsPlayerMoved(t *testing.T) {
	// given
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileNone

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when
	k.Submit(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + 1, Y: heroY}))
	events := k.Pass(0)

	// then
	if hero := state.Hero(); hero != (world.Tile{X: heroX + 1, Y: heroY}) {
		t.Fatalf("player tile after the step = %v, want %v", hero, world.Tile{X: heroX + 1, Y: heroY})
	}

	if flags := state.Grid().DFlags[heroY][heroX+1]; flags&world.FlagExplored == 0 {
		t.Fatalf("stepped tile flags = %d, want %d set", flags, world.FlagExplored)
	}

	if len(events) != 1 || events[0].Kind != kernel.KindPlayerMoved {
		t.Fatalf("events = %v, want one %q", events, kernel.KindPlayerMoved)
	}

	if events[0].A != heroX+1 || events[0].B != heroY {
		t.Fatalf("event payload = (%d,%d), want (%d,%d)", events[0].A, events[0].B, heroX+1, heroY)
	}
}

func TestRegisterMovement_BlockedStepChangesNothing(t *testing.T) {
	// given
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileSolid

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when
	k.Submit(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + 1, Y: heroY}))
	events := k.Pass(0)

	// then
	if hero := state.Hero(); hero != (world.Tile{X: heroX, Y: heroY}) {
		t.Fatalf("blocked step moved the player to %v, want %v", hero, world.Tile{X: heroX, Y: heroY})
	}

	if len(events) != 0 {
		t.Fatalf("blocked step emitted %d events, want 0 (no change, no event)", len(events))
	}
}

func TestRegisterMovement_DistantTileChangesNothing(t *testing.T) {
	// given
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileNone

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when
	k.Submit(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + 2, Y: heroY}))
	events := k.Pass(0)

	// then
	if hero := state.Hero(); hero != (world.Tile{X: heroX, Y: heroY}) {
		t.Fatalf("walk past one step moved the player to %v, want %v", hero, world.Tile{X: heroX, Y: heroY})
	}

	if len(events) != 0 {
		t.Fatalf("walk past one step emitted %d events, want 0", len(events))
	}
}

func TestRegisterMovement_LaterHandlerObservesTheMove(t *testing.T) {
	// given
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileNone

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	var observed world.Tile

	k.Register(kernel.KindPlayerMove, func(kernel.Command, *kernel.Streams) []kernel.Event {
		observed = state.Hero()

		return []kernel.Event{{Kind: kernel.KindPlayerStruck}}
	})

	// when
	k.Submit(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + 1, Y: heroY}))
	events := k.Pass(0)

	// then
	if observed != (world.Tile{X: heroX + 1, Y: heroY}) {
		t.Fatalf("second handler saw player tile %v, want the already moved %v", observed,
			world.Tile{X: heroX + 1, Y: heroY})
	}

	want := []kernel.Kind{kernel.KindPlayerMoved, kernel.KindPlayerStruck}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v (registration order)", events, want)
	}

	for i, kind := range want {
		if events[i].Kind != kind {
			t.Fatalf("event %d = %q, want %q (registration order)", i, events[i].Kind, kind)
		}
	}
}
