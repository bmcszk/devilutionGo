package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// stickSector is one classified stick sector and the tile its step lands on:
// the expectations are the upstream tables transcribed (plrctrls.cpp:100-105
// FaceDir composed with engine/displacement.hpp:247-275 fromDirection), so a
// sector mapped to the wrong delta fails by name.
type stickSector struct {
	name     string
	x, y     int32
	wantTile world.Tile
}

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

func TestRegisterMovement_StickPushWalksOneSectorTile(t *testing.T) {
	// given the nine sectors of a firm push, in axis-ordinal order
	sectors := []stickSector{
		{"NONE/NONE walks south", 0, 0, world.Tile{X: spawnTile.X + 1, Y: spawnTile.Y + 1}},
		{"NONE/UP walks north", 0, 1, world.Tile{X: spawnTile.X - 1, Y: spawnTile.Y - 1}},
		{"NONE/DOWN walks south", 0, 2, world.Tile{X: spawnTile.X + 1, Y: spawnTile.Y + 1}},
		{"LEFT/NONE walks west", 1, 0, world.Tile{X: spawnTile.X - 1, Y: spawnTile.Y + 1}},
		{"LEFT/UP walks north-west", 1, 1, world.Tile{X: spawnTile.X - 1, Y: spawnTile.Y}},
		{"LEFT/DOWN walks south-west", 1, 2, world.Tile{X: spawnTile.X, Y: spawnTile.Y + 1}},
		{"RIGHT/NONE walks east", 2, 0, world.Tile{X: spawnTile.X + 1, Y: spawnTile.Y - 1}},
		{"RIGHT/UP walks north-east", 2, 1, world.Tile{X: spawnTile.X, Y: spawnTile.Y - 1}},
		{"RIGHT/DOWN walks south-east", 2, 2, world.Tile{X: spawnTile.X + 1, Y: spawnTile.Y}},
	}

	// when each sector's push runs
	for _, sector := range sectors {
		// then each walks the tile the upstream tables name
		t.Run(sector.name, func(t *testing.T) { assertStickPushWalksOneTile(t, sector) })
	}
}

func TestRegisterMovement_GentleStickPushChangesNothing(t *testing.T) {
	// given a hero standing mid-level with every tile walkable
	state := world.NewState()

	if _, ok := state.LoadLevel(openLayout(), spawnTile, cathLevel, cathType); !ok {
		t.Fatalf("LoadLevel on the open layout failed, want the hero at %v", spawnTile)
	}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when a gentle push arrives: upstream only re-faces (WalkInDir runs on a
	// firm push alone, Movement plrctrls.cpp:1845-1850) and this slice carries
	// no facing, so the pass cannot change anything
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindStickPush, A: 1, B: 2, C: 0})
	events := k.Pass(0)

	// then nobody moved and nothing was announced
	if hero := state.Hero(); hero != spawnTile {
		t.Fatalf("a gentle push moved the hero to %v, want %v", hero, spawnTile)
	}

	if len(events) != 0 {
		t.Fatalf("a gentle push emitted %d events, want 0 (no change, no event)", len(events))
	}
}

func TestRegisterMovement_MouseWalkStepsAndAnnouncesTheMove(t *testing.T) {
	// given a hero at the slice origin with a walkable neighbour
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileNone

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when the bridge's walk row names the neighbour tile (KindMouseWalk A=x,
	// B=y of cursPosition, command.go/:81-83)
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindMouseWalk, A: heroX + 1, B: heroY + 1})
	events := k.Pass(0)

	// then the hero stepped and the kernel announced the new tile
	if hero := state.Hero(); hero != (world.Tile{X: heroX + 1, Y: heroY + 1}) {
		t.Fatalf("hero tile after the mouse walk = %v, want %v", hero, world.Tile{X: heroX + 1, Y: heroY + 1})
	}

	if len(events) != 1 || events[0].Kind != kernel.KindPlayerMoved {
		t.Fatalf("events = %v, want one %q", events, kernel.KindPlayerMoved)
	}

	if events[0].A != heroX+1 || events[0].B != heroY+1 {
		t.Fatalf("event payload = (%d,%d), want (%d,%d)", events[0].A, events[0].B, heroX+1, heroY+1)
	}
}

func TestRegisterMovement_BlockedMouseWalkChangesNothing(t *testing.T) {
	// given a hero whose neighbour is a wall
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileSolid

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when the walk row names that wall
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindMouseWalk, A: heroX + 1, B: heroY})
	events := k.Pass(0)

	// then the refused step moved nobody and emitted nothing
	if hero := state.Hero(); hero != (world.Tile{X: heroX, Y: heroY}) {
		t.Fatalf("a blocked mouse walk moved the hero to %v, want %v", hero, world.Tile{X: heroX, Y: heroY})
	}

	if len(events) != 0 {
		t.Fatalf("a blocked mouse walk emitted %d events, want 0 (no change, no event)", len(events))
	}
}

// assertStickPushWalksOneTile is one sector's given/when/then: a hero standing
// mid-level steps one tile toward the pushed sector and the pass says so.
func assertStickPushWalksOneTile(t *testing.T, sector stickSector) {
	t.Helper()

	// given a hero standing mid-level with every tile walkable
	state := world.NewState()

	if _, ok := state.LoadLevel(openLayout(), spawnTile, cathLevel, cathType); !ok {
		t.Fatalf("LoadLevel on the open layout failed, want the hero at %v", spawnTile)
	}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterMovement(k, state)

	// when a firm push names that sector (StickPush::Firm is 2,
	// controller_motion.h:32-37; the axis words are the AxisDirectionX/Y
	// ordinals, axis_direction.h:7-16)
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindStickPush, A: 2, B: sector.x, C: sector.y})
	events := k.Pass(0)

	// then the hero walked one tile toward the push and the pass said so
	if hero := state.Hero(); hero != sector.wantTile {
		t.Fatalf("hero tile after the push = %v, want %v", hero, sector.wantTile)
	}

	if len(events) != 1 || events[0].Kind != kernel.KindPlayerMoved {
		t.Fatalf("events = %v, want one %q", events, kernel.KindPlayerMoved)
	}

	if events[0].A != int32(sector.wantTile.X) || events[0].B != int32(sector.wantTile.Y) {
		t.Fatalf("event payload = (%d,%d), want (%d,%d)",
			events[0].A, events[0].B, sector.wantTile.X, sector.wantTile.Y)
	}
}
