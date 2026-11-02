// Package missiles_test pins the collection-level ProcessMissiles driver and
// its two DeleteMissiles sweeps (missiles.cpp:4194, :4219, :4243).
package missiles_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

func TestProcessMissiles_SweepsRowThatLeftTheDungeonBeforeThePass(t *testing.T) {
	// given
	c := missiles.NewMissileCollection()
	c.AddMissile(driverMissile(1, -1, 10, 5)) // out of bounds
	c.AddMissile(driverMissile(2, 64, 64, 5))
	w := newSilentWorld()

	// when
	c.ProcessMissiles(w)

	// then
	if got := sources(c); !slices.Equal(got, []int{2}) {
		t.Fatalf("survivors = %v, want [2]", got)
	}

	if got := c.Missiles()[0].Duration; got != 4 {
		t.Fatalf("in-bounds row duration = %d, want 4 (the swept row never ran)", got)
	}
}

func TestProcessMissiles_SweepsRowFlaggedByThePass(t *testing.T) {
	// given
	c := missiles.NewMissileCollection()
	c.AddMissile(driverMissile(1, 64, 64, 1)) // burns out this tick
	c.AddMissile(driverMissile(2, 65, 64, 5))
	w := newSilentWorld()

	// when
	c.ProcessMissiles(w)

	// then
	if got := sources(c); !slices.Equal(got, []int{2}) {
		t.Fatalf("survivors = %v, want [2]", got)
	}

	if got := c.Missiles()[0].Duration; got != 4 {
		t.Fatalf("survivor duration = %d, want 4", got)
	}
}

func TestProcessMissiles_KeepsInsertionOrderOfSurvivors(t *testing.T) {
	// given
	c := missiles.NewMissileCollection()
	c.AddMissile(driverMissile(1, 60, 60, 5))
	c.AddMissile(driverMissile(2, 61, 61, 1)) // flagged by the pass
	c.AddMissile(driverMissile(3, 62, 62, 5))
	c.AddMissile(driverMissile(4, 10, -1, 5)) // swept before the pass
	w := newSilentWorld()

	// when
	c.ProcessMissiles(w)

	// then
	if got := sources(c); !slices.Equal(got, []int{1, 3}) {
		t.Fatalf("survivors = %v, want [1 3] in insertion order", got)
	}
}

func TestProcessMissiles_ClearsMissileTileFlagsBeforeThePass(t *testing.T) {
	// given
	w := newSilentWorld()
	w.Grid.DFlags[70][66] = world.FlagMissile | world.FlagMissileFire | world.FlagMissileLight | world.FlagVisible
	c := missiles.NewMissileCollection()
	c.AddMissile(driverMissile(1, 66, 70, 5))

	// when
	c.ProcessMissiles(w)

	// then: the sweep cleared the three missile bits; the row's own PutMissile
	// then raised FlagMissile again (missiles.cpp:4222-4223, 412).
	if got := w.Grid.DFlags[70][66]; got != world.FlagVisible|world.FlagMissile {
		t.Fatalf("tile flags = %b, want the visible bit plus FlagMissile re-raised by the pass", got)
	}
}

func TestProcessMissiles_ResetsMissilePreFlagPerFrame(t *testing.T) {
	// given
	c := missiles.NewMissileCollection()
	c.AddMissile(driverMissile(1, 64, 64, 5))
	w := newSilentWorld()
	w.ProcessCallbacks.MissilePreFlag = true

	// when
	c.ProcessMissiles(w)

	// then
	if w.ProcessCallbacks.MissilePreFlag {
		t.Fatal("MissilePreFlag still set after the pass, want cleared (missiles.cpp:4231)")
	}
}

// driverMissile builds a ResurrectBeam row at a tile: its processFn only
// decrements the duration and flags the row at zero (missiles.cpp:4164-4170),
// so the duration is an observable "the pass ran this row" marker.
func driverMissile(source, x, y, duration int) missiles.Missile {
	m := missiles.Missile{Type: missiles.ResurrectBeam, Source: source, Duration: duration}
	m.SetTraveledFromTile(x, y)

	return m
}
