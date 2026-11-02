// Package missiles_test pins the search-failure and local-view arms of
// ProcessTeleport (missiles.cpp:3673-3674, 3689-3691).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// teleportViewSpy records the relocation, light, vision and view calls.
type teleportViewSpy struct {
	moves   []missiles.TilePosition
	lights  []missiles.TilePosition
	visions []missiles.TilePosition
	views   []missiles.TilePosition
}

// teleportWorldConfig selects the fixtures's search and view flags.
type teleportWorldConfig struct {
	allBlocked bool
	local      bool
}

// TestProcessTeleportFailedSearchReturnsIdle pins the missing-destination arm
// (missiles.cpp:3673-3674): the tick only burns the countdown — no delete, no
// relocation, no light, vision or view call.
func TestProcessTeleportFailedSearchReturnsIdle(t *testing.T) {
	// given
	spy := &teleportViewSpy{}
	w := teleportViewWorld(teleportWorldConfig{allBlocked: true}, spy)
	m := teleportViewMissile()

	// when
	missiles.ProcessTeleport(w, &m)

	// then
	if m.Duration != 1 {
		t.Fatalf("duration = %d, want 1", m.Duration)
	}

	if m.Var1 != 0 {
		t.Fatalf("var1 = %d, want the untouched 0", m.Var1)
	}

	if m.DelFlag {
		t.Fatalf("del = %v, want false", m.DelFlag)
	}

	if len(spy.moves) != 0 || len(spy.views) != 0 {
		t.Fatalf("moves/views = %v/%v, want none", spy.moves, spy.views)
	}
}

// TestProcessTeleportLocalPlayerViewFollows pins the last arm
// (missiles.cpp:3689-3691): the local player's relocation also parks the view
// on the new tile.
func TestProcessTeleportLocalPlayerViewFollows(t *testing.T) {
	// given
	spy := &teleportViewSpy{}
	w := teleportViewWorld(teleportWorldConfig{local: true}, spy)
	m := teleportViewMissile()

	// when
	missiles.ProcessTeleport(w, &m)

	// then
	if m.Var1 != 1 {
		t.Fatalf("var1 = %d, want 1", m.Var1)
	}

	if len(spy.moves) != 1 {
		t.Fatalf("moves = %v, want one", spy.moves)
	}

	if len(spy.views) != 1 {
		t.Fatalf("views = %v, want one", spy.views)
	}

	if spy.views[0] != spy.moves[0] {
		t.Fatalf("view = %v, want the move target %v", spy.views[0], spy.moves[0])
	}

	if m.DelFlag {
		t.Fatalf("del = %v, want false", m.DelFlag)
	}
}

// TestProcessTeleportRemotePlayerViewStays pins the non-local arm: a remote
// player relocates but the local view does not move (missiles.cpp:3689).
func TestProcessTeleportRemotePlayerViewStays(t *testing.T) {
	// given
	spy := &teleportViewSpy{}
	w := teleportViewWorld(teleportWorldConfig{}, spy)
	m := teleportViewMissile()

	// when
	missiles.ProcessTeleport(w, &m)

	// then
	if len(spy.moves) != 1 {
		t.Fatalf("moves = %v, want one", spy.moves)
	}

	if len(spy.views) != 0 {
		t.Fatalf("views = %v, want none", spy.views)
	}
}

// teleportViewMissile is a teleport body on its first relocation tick.
func teleportViewMissile() missiles.Missile {
	m := missiles.Missile{Type: missiles.Teleport, Source: 0, Duration: 2}
	m.Position.Tile = missiles.TilePosition{X: 10, Y: 12}

	return m
}

// teleportViewWorld wires the search predicate and the relocation sinks.
func teleportViewWorld(cfg teleportWorldConfig, spy *teleportViewSpy) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.PlayerAtPosition = func(missiles.TilePosition) int {
		if cfg.allBlocked {
			return 5 // a live second player on every tile
		}

		return -1
	}
	cb.IsMyPlayer = func(int) bool { return cfg.local }
	cb.RelocatePlayer = func(_ int, position missiles.TilePosition) {
		spy.moves = append(spy.moves, position)
	}
	cb.ChangeLightXY = func(_ int, position missiles.TilePosition) {
		spy.lights = append(spy.lights, position)
	}
	cb.ChangeVisionXY = func(_ int, position missiles.TilePosition) {
		spy.visions = append(spy.visions, position)
	}
	cb.SetViewPosition = func(position missiles.TilePosition) {
		spy.views = append(spy.views, position)
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}
}
