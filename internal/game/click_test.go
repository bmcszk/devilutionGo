package game_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// leftClickButton is SDL_BUTTON_LEFT, the recorded left mouse button.
const leftClickButton = 1

// recordCase is one demo record and the name the test reports it under.
type recordCase struct {
	name string
	rec  demo.Record
}

// clickView is the state view a click maps against: the grid slice measures
// range, the target table says what stands on a tile.
type clickView struct {
	state   *world.State
	targets map[world.Tile]bool
}

func (v clickView) Range(tile world.Tile) int { return v.state.Range(tile) }

func (v clickView) TargetAt(tile world.Tile) (alive, ok bool) {
	alive, ok = v.targets[tile]

	return alive, ok
}

func TestClickCommand_LeftPressWalksToThePointerTile(t *testing.T) {
	// given
	state := world.NewState()
	// The camera is presentation state: the host resolver lands every recorded
	// pointer position on this tile.
	tile := world.Tile{X: 1, Y: 0}
	router := game.NewClickRouter(clickView{state: state}, fixedTile(tile))

	// when
	cmd, ok := router.Command(demo.MouseButtonRecord{Button: leftClickButton, X: 40, Y: 20, Down: true})

	// then
	if !ok {
		t.Fatal("left press raised no command, want a walk command")
	}

	if cmd.Kind != kernel.KindPlayerMove {
		t.Fatalf("command kind = %q, want %q on an empty tile", cmd.Kind, kernel.KindPlayerMove)
	}

	if cmd.A != int32(tile.X) || cmd.B != int32(tile.Y) {
		t.Fatalf("command payload = (%d,%d), want (%d,%d)", cmd.A, cmd.B, tile.X, tile.Y)
	}
}

func TestClickCommand_LeftPressOnLiveTargetAttacks(t *testing.T) {
	// given
	tile := world.Tile{X: 1, Y: 0}
	view := clickView{state: world.NewState(), targets: map[world.Tile]bool{tile: true}}
	router := game.NewClickRouter(view, fixedTile(tile))

	// when
	cmd, ok := router.Command(demo.MouseButtonRecord{Button: leftClickButton, X: 40, Y: 20, Down: true})

	// then
	if !ok {
		t.Fatal("left press on a live target raised no command, want an attack")
	}

	if cmd.Kind != kernel.KindPlayerAttack {
		t.Fatalf("command kind = %q, want %q over a live target", cmd.Kind, kernel.KindPlayerAttack)
	}
}

func TestClickCommand_IgnoresNonPressRecords(t *testing.T) {
	// given
	cases := []recordCase{
		{name: "button release", rec: demo.MouseButtonRecord{Button: leftClickButton, Down: false}},
		{name: "other button press", rec: demo.MouseButtonRecord{Button: 2, Down: true}},
		{name: "pointer motion", rec: demo.MouseMotionRecord{X: 4, Y: 4}},
		{name: "tick record", rec: demo.TickRecord{Progress: 40}},
		{name: "key press", rec: demo.KeyRecord{Sym: 27, Down: true}},
	}
	router := game.NewClickRouter(clickView{state: world.NewState()}, fixedTile(world.Tile{}))

	// when
	raised := mapRecords(router, cases)

	// then
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if raised[i] {
				t.Fatalf("%s raised a gameplay command, want none", tc.name)
			}
		})
	}
}

func TestClickCommand_WalkClickMovesThePlayerThroughTheGate(t *testing.T) {
	// given
	state := world.NewState()
	state.Grid().SOLData[0] = world.TileNone
	tile := world.Tile{X: 1, Y: 0}
	view := clickView{state: state}

	k := kernel.New(world.NewGate(view, 0), nil)
	world.RegisterMovement(k, state)

	// when
	cmd, ok := game.NewClickRouter(view, fixedTile(tile)).
		Command(demo.MouseButtonRecord{Button: leftClickButton, X: 40, Y: 20, Down: true})
	accepted := ok && k.Submit(cmd)
	events := k.Pass(0)

	// then
	if !accepted {
		t.Fatalf("walk click was dropped by the gate (%v), want it dispatched", cmd)
	}

	if hero := state.Hero(); hero != tile {
		t.Fatalf("player tile = %v, want %v (state changes only through the handler)", hero, tile)
	}

	if len(events) != 1 || events[0].Kind != kernel.KindPlayerMoved {
		t.Fatalf("events = %v, want one %q announcing the move", events, kernel.KindPlayerMoved)
	}
}

func TestClickCommand_DeadTargetClickNeverReachesAHandler(t *testing.T) {
	// given: the click was predicted against a live target, and the target died
	// before the tick boundary dispatched the buffered command.
	state := world.NewState()
	tile := world.Tile{X: 1, Y: 0}
	producer := clickView{state: state, targets: map[world.Tile]bool{tile: true}}
	dispatched := clickView{state: state, targets: map[world.Tile]bool{tile: false}}
	runs := 0

	k := kernel.New(world.NewGate(dispatched, 0), nil)
	k.Register(kernel.KindPlayerAttack, func(kernel.Command, *kernel.Streams) []kernel.Event {
		runs++

		return []kernel.Event{{Kind: kernel.KindPlayerStruck}}
	})

	// when
	cmd, ok := game.NewClickRouter(producer, fixedTile(tile)).
		Command(demo.MouseButtonRecord{Button: leftClickButton, X: 40, Y: 20, Down: true})
	accepted := ok && k.Submit(cmd)
	events := k.Pass(0)

	// then
	if !ok || cmd.Kind != kernel.KindPlayerAttack {
		t.Fatalf("click raised %v (ok=%v), want an attack command", cmd, ok)
	}

	if accepted {
		t.Fatal("gate accepted a click on a dead target, want it dropped")
	}

	if runs != 0 || len(events) != 0 || len(k.Trace()) != 0 {
		t.Fatalf("dropped click ran %d handlers and emitted %d events, want none", runs, len(events))
	}
}

// mapRecords reports which demo records raised a gameplay command.
func mapRecords(router *game.ClickRouter, cases []recordCase) []bool {
	raised := make([]bool, len(cases))

	for i, tc := range cases {
		_, raised[i] = router.Command(tc.rec)
	}

	return raised
}

// fixedTile is a host resolver that lands every pointer position on one tile.
func fixedTile(tile world.Tile) game.TileAt {
	return func(uint16, uint16) world.Tile { return tile }
}
