// Package missiles_test pins AddTeleport (missiles.cpp:2950-2970) and
// ProcessTeleport (cpp:3900-3960), the shared relocation body both Teleport and
// Phasing run: the search for a legal tile around the aim, then the one-shot
// player move with its dPlayer/trans/light/vision bookkeeping.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// teleportCases pin the AddTeleport exits.
var teleportCases = []teleportCase{
	{name: "a legal aim tile arms the missile", want: missiles.TilePosition{X: 10, Y: 12}},
	{name: "a blocked aim tile fizzles", empty: true, wantDel: true},
}

// teleportCase pins one AddTeleport cast.
type teleportCase struct {
	name    string
	empty   bool
	want    missiles.TilePosition
	wantDel bool
}

// tcExpiry pins one ProcessTeleport configuration.
type tcExpiry struct {
	move bool
	town bool
}

// teleportSpy records the relocation the process body performed.
type teleportSpy struct {
	moves   []missiles.TilePosition
	lights  []missiles.TilePosition
	visions []missiles.TilePosition
	views   []missiles.TilePosition
}

// TestAddTeleport pins the destination search and the 2-tick life.
func TestAddTeleport(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(teleportCases))
	for _, tc := range teleportCases {
		subtests = append(subtests, subtest{tc.name, runTeleportAdd(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessTeleport pins the relocation and the expiry (cpp:3900-3960).
func TestProcessTeleport(t *testing.T) {
	// when
	subtests := []subtest{
		{"the last tick deletes the missile", runTeleportProcess(tcExpiry{})},
		{"the move clears the old tile and moves the light and vision", runTeleportProcess(
			tcExpiry{move: true})},
		{"town leaves the light and vision alone", runTeleportProcess(tcExpiry{move: true, town: true})},
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// teleportWorld wires the search predicate and the relocation sinks.
func teleportWorld(spy *teleportSpy) *missiles.HitWorld {
	return newTeleportWorld(spy, 1)
}

// teleportTownWorld is the DTYPE_TOWN variant: the move leaves light and
// vision alone.
func teleportTownWorld(spy *teleportSpy) *missiles.HitWorld {
	return newTeleportWorld(spy, 0)
}

// newTeleportWorld wires the search predicate and the relocation sinks for one
// level type.
func newTeleportWorld(spy *teleportSpy, levelType int) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.LevelType = levelType

	cb.MyPlayer = func() int { return 0 }
	cb.PlayerAtPosition = func(missiles.TilePosition) int { return -1 }
	cb.PlayerHasNoLife = func(int) bool { return false }
	cb.RelocatePlayer = func(_ int, position missiles.TilePosition) {
		spy.moves = append(spy.moves, position)
	}
	cb.ChangeLightXY = func(_ int, position missiles.TilePosition) {
		spy.lights = append(spy.lights, position)
	}
	cb.ChangeVisionXY = func(_ int, position missiles.TilePosition) {
		spy.visions = append(spy.visions, position)
	}
	cb.SetViewPosition = func(position missiles.TilePosition) { spy.views = append(spy.views, position) }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// blockAllTiles makes PosOkPlayer fail everywhere for the fizzle case.
func blockAllTiles(w *missiles.HitWorld, tc teleportCase) {
	if tc.empty {
		w.ProcessCallbacks.PlayerAtPosition = func(missiles.TilePosition) int { return 5 }
	}
}

// runTeleportAdd drives one AddTeleport.
func runTeleportAdd(tc teleportCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		w := teleportWorld(&teleportSpy{})
		blockAllTiles(w, tc)

		m := missiles.Missile{Type: missiles.Teleport, Source: 0}
		m.SetTraveledFromTile(10, 10)
		m.Position.Start = missiles.TilePosition{X: 10, Y: 10}
		p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 12}}

		// when
		missiles.AddTeleport(w, &m, p)

		// then
		if m.DelFlag != tc.wantDel {
			t.Fatalf("del = %v, want %v", m.DelFlag, tc.wantDel)
		}

		if tc.wantDel {
			assertTeleportFizzled(t, p)

			return
		}

		assertTeleportArmed(t, m, tc)
	}
}

// assertTeleportFizzled checks the spell-fizzle flag of a failed search.
func assertTeleportFizzled(t *testing.T, p *missiles.AddParameter) {
	t.Helper()

	if !p.SpellFizzled {
		t.Fatal("a fizzled teleport must raise spellFizzled")
	}
}

// assertTeleportArmed checks the found destination and the 2-tick life.
func assertTeleportArmed(t *testing.T, m missiles.Missile, tc teleportCase) {
	t.Helper()

	if m.Position.Tile != tc.want || m.Position.Start != tc.want || m.Duration != 2 {
		t.Fatalf("tile/start/duration = %v/%v/%d, want %v twice and 2",
			m.Position.Tile, m.Position.Start, m.Duration, tc.want)
	}
}

// runTeleportProcess drives one ProcessTeleport pair.
func runTeleportProcess(cfg tcExpiry) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &teleportSpy{}
		w := teleportWorld(spy)
		if cfg.town {
			w = teleportTownWorld(spy)
		}
		m := missiles.Missile{Type: missiles.Teleport, Source: 0, Duration: 2}
		m.Position.Tile = missiles.TilePosition{X: 10, Y: 12}
		if !cfg.move {
			m.Duration = 1
		}

		// when
		missiles.ProcessTeleport(w, &m)

		// then
		if !cfg.move {
			assertTeleportExpiry(t, m, spy)

			return
		}

		assertTeleportMove(t, m, spy, cfg)
	}
}

// assertTeleportExpiry checks the expiry tick.
func assertTeleportExpiry(t *testing.T, m missiles.Missile, spy *teleportSpy) {
	t.Helper()

	if m.Duration != 0 || !m.DelFlag {
		t.Fatalf("duration/del = %d/%v, want the countdown to delete the missile", m.Duration, m.DelFlag)
	}

	if len(spy.moves) != 0 {
		t.Fatalf("moves = %v, want none", spy.moves)
	}
}

// assertTeleportMove checks the relocation and its bookkeeping.
func assertTeleportMove(t *testing.T, m missiles.Missile, spy *teleportSpy, cfg tcExpiry) {
	t.Helper()

	want := missiles.TilePosition{X: 10, Y: 12}
	if len(spy.moves) != 1 || spy.moves[0] != want || m.Var1 != 1 {
		t.Fatalf("moves/var1 = %v/%d, want [%v]/1", spy.moves, m.Var1, want)
	}

	assertTeleportLighting(t, spy, cfg)
}

// assertTeleportLighting checks the light/vision/view traffic of one move.
func assertTeleportLighting(t *testing.T, spy *teleportSpy, cfg tcExpiry) {
	t.Helper()

	if cfg.town {
		if len(spy.lights) != 0 || len(spy.visions) != 0 {
			t.Fatalf("town touched light/vision: %v/%v", spy.lights, spy.visions)
		}

		return
	}

	// the view only moves for MyPlayer, and the fixture's local player is -1
	if len(spy.lights) != 1 || len(spy.visions) != 1 || len(spy.views) != 0 {
		t.Fatalf("light/vision/view = %v/%v/%v, want one light, one vision, no view",
			spy.lights, spy.visions, spy.views)
	}
}
