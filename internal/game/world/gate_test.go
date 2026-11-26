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

// scriptedSession is the SessionView a test scripts: a level word plus a
// blocked-tile table, so the create rule is provable without a level.
type scriptedSession struct {
	state   *world.State
	level   int
	kind    int
	loaded  bool
	hero    bool
	blocked map[world.Tile]bool
}

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

func (s scriptedSession) Range(tile world.Tile) int { return s.state.Range(tile) }

func (scriptedSession) TargetAt(world.Tile) (alive, ok bool) { return false, false }

func (s scriptedSession) Level() (level, levelType int, ok bool) { return s.level, s.kind, s.loaded }

func (s scriptedSession) SpawnAllowed(tile world.Tile) bool { return !s.blocked[tile] }

func (s scriptedSession) HeroRaised() bool { return s.hero }

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

func TestGate_DropsACreateOnABlockedSpawn(t *testing.T) {
	// given a gate whose session says the spawn tile cannot host a hero
	spawn := world.Tile{X: heroX + 2, Y: heroY + 2}
	session := scriptedSession{state: world.NewState(), blocked: map[world.Tile]bool{spawn: true}}
	gate := world.NewSessionGate(session, session, 0)

	// when
	rule := gate.Reject(kernel.Command{
		Tick: 0, Kind: kernel.KindCreatePack, A: 0, B: int32(spawn.X), C: int32(spawn.Y),
	})

	// then
	if rule != world.RuleBlockedSpawn {
		t.Fatalf("create on a blocked spawn judged %q, want %q", rule, world.RuleBlockedSpawn)
	}
}

func TestGate_DropsASaveWithNoHero(t *testing.T) {
	// given a gate whose session has raised no hero
	session := scriptedSession{state: world.NewState()}
	gate := world.NewSessionGate(session, session, 0)

	// when
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: 0, B: 2588})

	// then
	if rule != world.RuleNoHero {
		t.Fatalf("save with no hero judged %q, want %q", rule, world.RuleNoHero)
	}
}

func TestGate_AcceptsASaveFromASessionWithAHero(t *testing.T) {
	// given a gate whose session has raised a hero
	session := scriptedSession{state: world.NewState(), hero: true}
	gate := world.NewSessionGate(session, session, 0)

	// when
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: 0, B: 2588})

	// then
	if rule != world.RuleAccepted {
		t.Fatalf("save from a session with a hero judged %q, want accepted", rule)
	}
}

func TestGate_AcceptsACreateOnAWalkableSpawn(t *testing.T) {
	// given a gate whose session lets a hero stand anywhere
	session := scriptedSession{state: world.NewState()}
	gate := world.NewSessionGate(session, session, 0)

	// when
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindCreatePack, A: 0, B: 1, C: 1})

	// then
	if rule != world.RuleAccepted {
		t.Fatalf("create on a walkable spawn judged %q, want accepted", rule)
	}
}

func TestGate_RejectsAMouseWalkBeyondReach(t *testing.T) {
	// given a gate over the real slice, with the hero at the origin
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when the bridge's walk row names a tile at reach and one step further
	atReach := gate.Reject(tileCommand(kernel.KindMouseWalk, 0, world.Tile{X: heroX + world.StepReach, Y: heroY}))
	beyond := gate.Reject(tileCommand(kernel.KindMouseWalk, 0, world.Tile{X: heroX + world.StepReach + 1, Y: heroY}))

	// then the walk row is judged like the click it is: the reach rule owns it
	if atReach != world.RuleAccepted || beyond != world.RuleOutOfRange {
		t.Fatalf("mouse walk at reach judged %q, beyond reach judged %q, want %q and %q",
			atReach, beyond, world.RuleAccepted, world.RuleOutOfRange)
	}
}

func TestGate_AcceptsAFirmStickPushAndDropsAGentleOne(t *testing.T) {
	// given a gate over the real slice
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when a firm push and a gentle push name the same sector
	firm := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindStickPush, A: 2, B: 2, C: 0})
	gentle := gate.Reject(kernel.Command{Tick: 1, Kind: kernel.KindStickPush, A: 1, B: 2, C: 0})

	// then only the firm push may step: a gentle one only re-faces upstream, and
	// this slice carries no facing, so the pass could not change anything
	if firm != world.RuleAccepted || gentle != world.RuleNoStep {
		t.Fatalf("firm push judged %q, gentle push judged %q, want %q and %q",
			firm, gentle, world.RuleAccepted, world.RuleNoStep)
	}
}

func TestGate_DropsANoneNoneStickPush(t *testing.T) {
	// given a gate over the real slice
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when a firm push with no x or y deflection arrives (the stick at rest)
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindStickPush, A: 2, B: 0, C: 0})

	// then the gate drops it: upstream's WalkInDir returns before reading
	// FaceDir for a NONE/NONE push (plrctrls.cpp:1747 @ca91d778f), so the
	// table's South cell must never become a step
	if rule != world.RuleNoStep {
		t.Fatalf("NONE/NONE firm push judged %q, want %q", rule, world.RuleNoStep)
	}
}

func TestGate_DropsAStickPushWithAForeignSector(t *testing.T) {
	// given a gate over the real slice
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when a hand-edited log word names an axis ordinal no sector has
	rule := gate.Reject(kernel.Command{Tick: 0, Kind: kernel.KindStickPush, A: 2, B: 7, C: 0})

	// then it names no step, so the pass cannot change anything
	if rule != world.RuleNoStep {
		t.Fatalf("a foreign sector judged %q, want %q", rule, world.RuleNoStep)
	}
}

// tileCommand builds a command aimed at one tile.

func TestGate_DropsTickRateOutsideTheUpstreamRange(t *testing.T) {
	// given: the plain gate
	gate := world.NewGate(scriptedView{state: world.NewState()}, 0)

	// when: out-of-range rates are judged
	low := gate.Reject(kernel.Command{Kind: kernel.KindSetTickRate, A: 19})
	high := gate.Reject(kernel.Command{Kind: kernel.KindSetTickRate, A: 51})

	// then: both are dropped by the range rule; 50 passes
	if low != world.RuleTickRateRange {
		t.Fatalf("Reject(A:19) = %q, want %q", low, world.RuleTickRateRange)
	}
	if high != world.RuleTickRateRange {
		t.Fatalf("Reject(A:51) = %q, want %q", high, world.RuleTickRateRange)
	}
	if rule := gate.Reject(kernel.Command{Kind: kernel.KindSetTickRate, A: 50}); rule != world.RuleAccepted {
		t.Fatalf("Reject(A:50) = %q, want accepted", rule)
	}
}

func tileCommand(kind kernel.Kind, tick uint32, tile world.Tile) kernel.Command {
	return kernel.Command{Tick: tick, Kind: kind, A: int32(tile.X), B: int32(tile.Y)}
}
