// Package missiles_test: the slice-1 wiring gate — each of the five wired
// ids must resolve to the real addFn of its misdat row, not the no-op stub
// (registry.go).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

func TestWiredIdsHaveRealAddBodies(t *testing.T) {
	// given
	ids := []missiles.MissileID{
		missiles.Arrow,
		missiles.Firebolt,
		missiles.MagmaBall,
		missiles.BloodStar,
		missiles.Acid,
	}

	// when
	behaviors := make([]missiles.Behavior, 0, len(ids))
	for _, id := range ids {
		behavior, ok := missiles.BehaviorFor(id)
		if !ok {
			t.Fatalf("no behavior for missile id %d", id)
		}

		behaviors = append(behaviors, behavior)
	}

	// then
	for i, id := range ids {
		t.Run(idName(id), func(t *testing.T) {
			assertRealAddBody(t, id, behaviors[i])
			assertRealProcessBody(t, id, behaviors[i])
		})
	}
}

// assertRealProcessBody dispatches one wired id's Process side and asserts the
// ported body ran: a one-tick missile that expires deletes itself and removes
// its light (missiles.cpp:2960-2964 ProcessArrow, 2970-2980
// ProcessGenericProjectile), which the no-op stub never does.
func assertRealProcessBody(t *testing.T, id missiles.MissileID, behavior missiles.Behavior) {
	t.Helper()

	m := &missiles.Missile{Type: id, Caster: 0, Source: 0, Duration: 1, Lid: 7}
	m.SetTraveledFromTile(64, 64)

	// when
	behavior.Process(m)

	// then
	if !m.DelFlag {
		t.Fatalf("missile id %d still dispatches the no-op process stub", id)
	}
}

// assertRealAddBody dispatches one wired id and asserts a real add body ran:
// every ported body sets a duration, the no-op stub leaves 0.
func assertRealAddBody(t *testing.T, id missiles.MissileID, behavior missiles.Behavior) {
	t.Helper()

	m := &missiles.Missile{Type: id, Caster: 0, Source: 0}
	m.SetTraveledFromTile(0, 0)
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 10}, Midir: missiles.South}

	// when
	behavior.Add(m, p)

	// then
	if m.Duration == 0 {
		t.Fatalf("missile id %d still dispatches the no-op stub", id)
	}
}

// idName labels a sub-test with the misdat row name.
func idName(id missiles.MissileID) string {
	switch id {
	case missiles.Arrow:
		return "Arrow"
	case missiles.Firebolt:
		return "Firebolt"
	case missiles.MagmaBall:
		return "MagmaBall"
	case missiles.BloodStar:
		return "BloodStar"
	case missiles.Acid:
		return "Acid"
	default:
		return "unknown"
	}
}
