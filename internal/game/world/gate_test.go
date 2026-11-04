// Package world_test proves the gameplay gate rules (ADR-018 law 6) and the
// movement handlers that own the grid slice (laws 2 and 3).
package world_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// heroX/heroY is the tile the player starts on, the slice's constructor origin.
const (
	heroX = 0
	heroY = 0
)

// scriptedView is the gate's state view with a target table behind it: the grid
// slice holds no monsters yet, so the live-target rule needs a scripted view.
// Range stays the real grid-slice measurement.
type scriptedView struct {
	state   *world.State
	targets map[world.Tile]bool
}

func (v scriptedView) Range(tile world.Tile) int { return v.state.Range(tile) }

func (v scriptedView) TargetAt(tile world.Tile) (alive, ok bool) {
	alive, ok = v.targets[tile]

	return alive, ok
}

func TestGate_RejectsClickBeyondReach(t *testing.T) {
	// given
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when
	atReach := gate.Reject(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + world.StepReach, Y: heroY}))
	beyond := gate.Reject(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + world.StepReach + 1, Y: heroY}))

	// then
	if atReach != world.RuleAccepted {
		t.Fatalf("click at reach %d rejected with %q, want accepted", world.StepReach, atReach)
	}

	if beyond != world.RuleOutOfRange {
		t.Fatalf("click past reach rejected with %q, want %q", beyond, world.RuleOutOfRange)
	}
}

func TestGate_DropsDeadTargetClickBeforeAnyHandler(t *testing.T) {
	// given
	view := scriptedView{
		state:   world.NewState(),
		targets: map[world.Tile]bool{{X: heroX + 1, Y: heroY}: false},
	}
	runs := 0

	k := kernel.New(world.NewGate(view, 0), nil)
	k.Register(kernel.KindPlayerAttack, func(kernel.Command, *kernel.Streams) []kernel.Event {
		runs++

		return []kernel.Event{{Kind: kernel.KindPlayerStruck}}
	})

	// when
	accepted := k.Submit(tileCommand(kernel.KindPlayerAttack, 0, world.Tile{X: heroX + 1, Y: heroY}))
	events := k.Pass(0)

	// then
	if accepted {
		t.Fatal("gate accepted a click on a dead target, want it dropped")
	}

	if runs != 0 {
		t.Fatalf("dropped click reached a handler %d times, want 0", runs)
	}

	if len(events) != 0 || len(k.Trace()) != 0 {
		t.Fatalf("dropped click emitted %d events, want 0 (no change, no event)", len(k.Trace()))
	}

	if hero := view.state.Hero(); hero != (world.Tile{X: heroX, Y: heroY}) {
		t.Fatalf("dropped click moved the player to %v, want %v", hero, world.Tile{X: heroX, Y: heroY})
	}

	if flags := view.state.Grid().DFlags[heroY][heroX]; flags != world.FlagNone {
		t.Fatalf("dropped click wrote grid flags %d, want none", flags)
	}
}

func TestGate_DropsClicksInsideTheCooldownWindow(t *testing.T) {
	// given
	gate := world.NewGate(scriptedView{state: world.NewState()}, 5)
	ticks := []uint32{10, 13, 15}
	want := []world.GateRule{world.RuleAccepted, world.RuleCooldown, world.RuleAccepted}

	// when
	got := make([]world.GateRule, 0, len(ticks))

	for _, tick := range ticks {
		got = append(got, gate.Reject(tileCommand(kernel.KindPlayerMove, tick, world.Tile{X: heroX + 1, Y: heroY})))
	}

	// then
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("click at tick %d judged %q, want %q", ticks[i], got[i], want[i])
		}
	}
}

func TestGate_AcceptsWalkInReachAndAttackOnLiveTarget(t *testing.T) {
	// given
	view := scriptedView{
		state:   world.NewState(),
		targets: map[world.Tile]bool{{X: heroX + 1, Y: heroY}: true},
	}
	gate := world.NewGate(view, 0)

	// when
	walk := gate.Reject(tileCommand(kernel.KindPlayerMove, 0, world.Tile{X: heroX + 1, Y: heroY}))
	attack := gate.Reject(tileCommand(kernel.KindPlayerAttack, 0, world.Tile{X: heroX + 1, Y: heroY}))

	// then
	if walk != world.RuleAccepted || attack != world.RuleAccepted {
		t.Fatalf("walk judged %q, attack judged %q, want both accepted", walk, attack)
	}
}

func TestGate_PassesNonGameplayCommandsUntouched(t *testing.T) {
	// given
	gate := world.NewGate(scriptedView{state: world.NewState()}, 5)

	// when
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindKey, A: 27})

	// then
	if rule != world.RuleAccepted {
		t.Fatalf("key command judged %q, want accepted: the gate judges only gameplay commands", rule)
	}
}

// tileCommand builds a command aimed at one tile.
func tileCommand(kind kernel.Kind, tick uint32, tile world.Tile) kernel.Command {
	return kernel.Command{Tick: tick, Kind: kind, A: int32(tile.X), B: int32(tile.Y)}
}
