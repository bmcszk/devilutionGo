// Package missiles_test pins AddTownPortal (missiles.cpp:2066-2115) and
// ProcessTownPortal (cpp:4070-4098): the portal that lands on a free tile with
// no trigger beside it, ages from open to idle frames, ramps its light and
// warps the players standing on it.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// townPortalIdleGroupForTest is PortalFrame::Idle (missiles.h:79-82).
const townPortalIdleGroupForTest = 1

// townPortalCases pin the AddTownPortal exits.
var townPortalCases = []townPortalCase{
	{name: "town parks the portal on the aim tile", town: true, want: missiles.TilePosition{X: 10, Y: 12}},
	{name: "the dungeon searches for a free tile", want: missiles.TilePosition{X: 10, Y: 12}},
	{name: "a blocked search deletes the portal", empty: true, wantDel: true},
}

// townPortalCase pins one AddTownPortal cast.
type townPortalCase struct {
	name    string
	town    bool
	empty   bool
	want    missiles.TilePosition
	wantDel bool
}

// townPortalTick pins one ProcessTownPortal configuration.
type townPortalTick struct {
	idle     bool
	ramp     bool
	stander  bool
	floored  bool
	expiring bool
}

// townPortalWarps records the warp traffic of one tick.
type townPortalWarps struct {
	cleared int
	sent    int
}

// townPortalSpy records the add body's side effects.
type townPortalSpy struct {
	expired   int
	activated []missiles.TilePosition
}

// TestAddTownPortal pins the destination, the 100-tick life and the var1
// threshold (cpp:2126-2128).
func TestAddTownPortal(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(townPortalCases))
	for _, tc := range townPortalCases {
		subtests = append(subtests, subtest{tc.name, runTownPortalAdd(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessTownPortal pins the countdown, the idle flip, the light ramp and
// the warp of the players standing on the tile (cpp:4070-4098).
func TestProcessTownPortal(t *testing.T) {
	// when
	subtests := []subtest{
		{"the first tick only counts down", runTownPortalTick(townPortalTick{})},
		{"the idle threshold swaps the frames", runTownPortalTick(townPortalTick{idle: true})},
		{"the light ramp steps once per tick", runTownPortalTick(townPortalTick{ramp: true})},
		{"a player on the tile is warped", runTownPortalTick(townPortalTick{stander: true})},
		{"the countdown floors at one", runTownPortalTick(townPortalTick{floored: true})},
		{"a zeroed portal deletes itself", runTownPortalTick(townPortalTick{expiring: true})},
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// townPortalWorld wires the search predicate inputs and the portal sinks.
func townPortalWorld(spy *townPortalSpy, tc townPortalCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.LevelType = 1
	if tc.town {
		cb.LevelType = 0
	}

	cb.IsMyPlayer = func(int) bool { return false }
	cb.ExpireSiblingTownPortals = func(*missiles.Missile) { spy.expired++ }
	cb.ActivatePortal = func(position missiles.TilePosition) { spy.activated = append(spy.activated, position) }
	cb.TrigNear = func(missiles.TilePosition) bool { return false }
	cb.PlayerAtPosition = func(missiles.TilePosition) int { return -1 }
	if tc.empty {
		cb.TrigNear = func(missiles.TilePosition) bool { return true }
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}

// runTownPortalAdd drives one AddTownPortal.
func runTownPortalAdd(tc townPortalCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &townPortalSpy{}
		w := townPortalWorld(spy, tc)
		m := missiles.Missile{Type: missiles.TownPortal, Source: 0, AnimLen: 12}
		m.SetTraveledFromTile(10, 10)
		m.Position.Start = missiles.TilePosition{X: 20, Y: 20}

		// when
		missiles.AddTownPortal(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 12}})

		// then
		assertTownPortalAdd(t, m, spy, tc)
	}
}

// assertTownPortalAdd checks the destination, life, threshold and the portal
// activation of the local player.
func assertTownPortalAdd(t *testing.T, m missiles.Missile, spy *townPortalSpy, tc townPortalCase) {
	t.Helper()

	if m.DelFlag != tc.wantDel {
		t.Fatalf("del = %v, want %v", m.DelFlag, tc.wantDel)
	}

	if tc.wantDel {
		return
	}

	assertTownPortalLanding(t, m, tc)
	assertTownPortalSiblings(t, spy)
}

// assertTownPortalLanding checks the destination and the 100-tick life
// (cpp:2126-2128: the idle swap fires at 100 - AnimLen).
func assertTownPortalLanding(t *testing.T, m missiles.Missile, tc townPortalCase) {
	t.Helper()

	if m.Position.Tile != tc.want || m.Position.Start != tc.want {
		t.Fatalf("tile/start = %v/%v, want %v twice", m.Position.Tile, m.Position.Start, tc.want)
	}

	if m.Duration != 100 || m.Var1 != 88 {
		t.Fatalf("duration/var1 = %d/%d, want 100/88", m.Duration, m.Var1)
	}
}

// assertTownPortalSiblings checks the older-portal expiry and the activation
// gate (cpp:2128-2142).
func assertTownPortalSiblings(t *testing.T, spy *townPortalSpy) {
	t.Helper()

	if spy.expired != 1 {
		t.Fatalf("expired siblings = %d, want the older portal zeroed", spy.expired)
	}

	if len(spy.activated) != 0 {
		t.Fatalf("activated = %v, want none for a non-local caster", spy.activated)
	}
}

// runTownPortalTick drives one ProcessTownPortal tick.
func runTownPortalTick(cfg townPortalTick) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &townPortalSpy{}
		w := townPortalWorld(spy, townPortalCase{})
		warped := &townPortalWarps{}
		w.ProcessCallbacks.ActivePlayers = func() []int { return []int{0} }
		w.ProcessCallbacks.IsMyPlayer = func(int) bool { return true }
		w.ProcessCallbacks.PortalReadyPlayer = func(int, missiles.TilePosition) bool { return cfg.stander }
		w.ProcessCallbacks.ClearPlayerPath = func(int) { warped.cleared++ }
		w.ProcessCallbacks.SendWarpCmd = func(int) { warped.sent++ }
		m := townPortalTickMissile(cfg)

		// when
		missiles.ProcessTownPortal(w, &m)

		// then
		assertTownPortalTick(t, m, warped, cfg)
	}
}

// townPortalTickMissile is the portal one tick starts from.
func townPortalTickMissile(cfg townPortalTick) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.TownPortal, Source: 0, Duration: 40, Var1: 88, Lid: 3,
		AnimLen: 12,
	}
	m.Position.Tile = missiles.TilePosition{X: 10, Y: 12}
	if cfg.idle {
		m.Duration = 89
	}

	if cfg.floored {
		m.Duration = 1
	}

	if cfg.expiring {
		m.Duration = 0
	}

	return m
}

// assertTownPortalTick checks the countdown, the frame swap, the ramp, the warp
// and the delete.
func assertTownPortalTick(t *testing.T, m missiles.Missile, warped *townPortalWarps, cfg townPortalTick) {
	t.Helper()

	if cfg.floored || cfg.expiring {
		assertTownPortalFloor(t, m, cfg)

		return
	}

	assertTownPortalFrames(t, m, warped, cfg)
}

// assertTownPortalFrames checks the idle flip, the ramp step and the warp.
func assertTownPortalFrames(t *testing.T, m missiles.Missile, warped *townPortalWarps, cfg townPortalTick) {
	t.Helper()

	if cfg.idle && m.MFNum != townPortalIdleGroupForTest {
		t.Fatalf("frame group = %d, want the idle group 1", m.MFNum)
	}

	assertTownPortalRampAndWarp(t, m, warped, cfg)
}

// assertTownPortalRampAndWarp checks the expLight step and the warp of the
// standing player.
func assertTownPortalRampAndWarp(t *testing.T, m missiles.Missile, warped *townPortalWarps, cfg townPortalTick) {
	t.Helper()

	if cfg.stander && (warped.cleared != 1 || warped.sent != 1) {
		t.Fatalf("warp = %d cleared / %d sent, want one of each", warped.cleared, warped.sent)
	}

	if cfg.ramp && m.Var2 != 1 {
		t.Fatalf("ramp step = %d, want one step", m.Var2)
	}
}

// assertTownPortalFloor checks the floored countdown and the zeroed portal.
func assertTownPortalFloor(t *testing.T, m missiles.Missile, cfg townPortalTick) {
	t.Helper()

	if cfg.floored && (m.DelFlag || m.Duration != 1) {
		t.Fatalf("del/duration = %v/%d, want a portal that holds its last tick",
			m.DelFlag, m.Duration)
	}

	if cfg.expiring && !m.DelFlag {
		t.Fatal("a portal whose duration reached zero must delete itself")
	}
}
