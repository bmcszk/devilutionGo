package world_test

// Pins for the session verbs (devilGo-sj5a): descend loads a level image and
// announces it, loot takes the floor item under the player. Both own their
// state through the handler only (ADR-018 laws 2 and 3).

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// The cathedral level and dungeon_type the fixtures use (gendung.cpp:326-344:
// level 1 is DTYPE_CATHEDRAL, ordinal 1 in gendung_defs.hpp:15-26).
const (
	cathLevel = 1
	cathType  = 1
)

// scriptedLevels is the LevelSource a test scripts: one level image per level
// number, plus the entry tile the load starts from.
type scriptedLevels struct {
	images map[int]*world.LevelLayout
	spawn  world.Tile
}

func (s scriptedLevels) LevelImage(level, _ int) (*world.LevelLayout, world.Tile, bool) {
	image, ok := s.images[level]

	return image, s.spawn, ok
}

func TestDescend_LoadsTheLevelAndAnnouncesIt(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}, spawn: world.Tile{X: 40, Y: 40}}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	accepted := k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(0)

	// then the gate passed it, one event announced it, and the slice moved
	if !accepted {
		t.Fatal("gate dropped the first descend, want it dispatched")
	}

	assertDescendedEvent(t, events, cathLevel, cathType)
	assertLevelState(t, state, source.spawn, cathLevel, cathType)
}

func TestDescend_FoldsTheLayoutIntoTheGrid(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: solidLayout(world.Tile{X: 10, Y: 10})}}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// then the piece layer and its SOL classification both landed
	if !state.Grid().IsTileSolid(10, 10) {
		t.Fatal("descend did not load the level's solid tile, want the layout folded into the grid")
	}

	if !state.Grid().IsTileWalkable(11, 11) {
		t.Fatal("descend marked a walkable tile solid, want the layout's own classification")
	}
}

func TestDescend_UnknownLevelChangesNothing(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{}}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("unloadable level emitted %d events, want 0 (no change, no event)", len(events))
	}

	if _, _, ok := state.Level(); ok {
		t.Fatal("unloadable level recorded a loaded level, want the session untouched")
	}
}

func TestDescend_GateDropsAReloadOfTheSameLevel(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}}
	gate := world.NewSessionGate(state, state, 0)

	k := kernel.New(gate, nil)
	world.RegisterSessionVerbs(k, state, source)

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// when the same level is asked for again
	rule := gate.Reject(kernel.Command{Tick: 1, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	accepted := k.Submit(kernel.Command{Tick: 1, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(1)

	// then
	if rule != world.RuleSameLevel {
		t.Fatalf("reload judged %q, want %q", rule, world.RuleSameLevel)
	}

	if accepted || len(events) != 0 {
		t.Fatalf("reload accepted=%v events=%d, want the gate to drop it before dispatch", accepted, len(events))
	}
}

func TestLoot_TakesTheFloorItemAndAnnouncesIt(t *testing.T) {
	// given a session standing on a level whose entry tile holds an item
	const itemIndex = 7

	tile := world.Tile{X: 20, Y: 21}
	layout := openLayout()
	layout.Items = []world.FloorItem{{Index: itemIndex, Tile: tile}}

	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, scriptedLevels{
		images: map[int]*world.LevelLayout{cathLevel: layout}, spawn: tile,
	})

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// when
	k.Submit(kernel.Command{Tick: 1, Kind: kernel.KindLoot})
	events := k.Pass(1)

	// then
	if len(events) != 1 || events[0].Kind != kernel.KindLooted {
		t.Fatalf("events = %v, want one %q", events, kernel.KindLooted)
	}

	if events[0].A != itemIndex || events[0].B != int32(tile.X) || events[0].C != int32(tile.Y) {
		t.Fatalf("event payload = (%d,%d,%d), want (%d,%d,%d)",
			events[0].A, events[0].B, events[0].C, itemIndex, tile.X, tile.Y)
	}

	if got := state.Grid().DItem[tile.Y][tile.X]; got != 0 {
		t.Fatalf("dItem at the looted tile = %d, want 0 (the item left the floor)", got)
	}
}

func TestLoot_EmptyTileChangesNothing(t *testing.T) {
	// given
	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}})

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindLoot})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("loot on an empty tile emitted %d events, want 0 (no change, no event)", len(events))
	}
}

// openLayout is a level image whose whole 112x112 grid is walkable.
func openLayout() *world.LevelLayout {
	layout := &world.LevelLayout{
		Piece: make([]int, world.MAXDUNX*world.MAXDUNY),
		SOL:   make([]byte, 1),
	}

	return layout
}

// solidLayout is openLayout with one tile made solid.
func solidLayout(tile world.Tile) *world.LevelLayout {
	layout := openLayout()
	layout.Piece[tile.X*world.MAXDUNY+tile.Y] = 1
	layout.SOL = []byte{0, 0x01}

	return layout
}

// assertDescendedEvent pins the announce event's kind and payload.
func assertDescendedEvent(t *testing.T, events []kernel.Event, level, levelType int) {
	t.Helper()

	if len(events) != 1 || events[0].Kind != kernel.KindDescended {
		t.Fatalf("events = %v, want one %q", events, kernel.KindDescended)
	}

	if events[0].A != int32(level) || events[0].B != int32(levelType) {
		t.Fatalf("event payload = (%d,%d), want (%d,%d)", events[0].A, events[0].B, level, levelType)
	}
}

// assertLevelState pins where the load put the player and what it recorded.
func assertLevelState(t *testing.T, state *world.State, hero world.Tile, level, levelType int) {
	t.Helper()

	if got := state.Hero(); got != hero {
		t.Fatalf("hero at %v after the descend, want the level's entry %v", got, hero)
	}

	if gotLevel, kind, ok := state.Level(); !ok || gotLevel != level || kind != levelType {
		t.Fatalf("level state = (%d,%d,%v), want (%d,%d,true)", gotLevel, kind, ok, level, levelType)
	}
}
