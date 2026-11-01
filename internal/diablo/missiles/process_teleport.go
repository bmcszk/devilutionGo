// Package missiles: the Teleport bodies (shared with Phasing).
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// AddTeleport ports AddTeleport (missiles.cpp:2950-2970): the destination is
// the nearest tile the caster may stand on within 5 of the aim; the missile
// then sits on it with a 2-tick life, and a failed search fizzles the spell.
func AddTeleport(w *HitWorld, m *Missile, p *AddParameter) {
	spot, ok := teleportDestination(w, m.Source, p.Dst)
	if !ok { // cpp:2965-2968
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.Position.Tile = spot  // cpp:2959-2961
	m.Position.Start = spot // cpp:2960-2961
	m.Duration = 2          // cpp:2962
}

// ProcessTeleport ports ProcessTeleport (missiles.cpp:3900-3960): the first
// tick only counts down, then the body relocates the caster onto the tile the
// search finds — clearing the old dPlayer slot, moving the player's tile,
// future and old positions, re-marking the trans tiles, re-occupying the new
// tile, and moving the light and vision (town keeps them still) — and finally
// parks the local player's view on the new tile.
func ProcessTeleport(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3902

	if m.Duration <= 0 { // cpp:3903-3906
		m.DelFlag = true

		return
	}

	player := m.Source

	spot, ok := teleportDestination(w, player, m.Position.Tile)
	if !ok { // cpp:3916-3917
		return
	}

	w.ProcessCallbacks.RelocatePlayer(player, spot) // cpp:3919-3928
	m.Var1 = 1                                      // cpp:3929

	if w.ProcessCallbacks.LevelType != levelTypeTown { // cpp:3931-3934
		w.ProcessCallbacks.ChangeLightXY(player, spot)
		w.ProcessCallbacks.ChangeVisionXY(player, spot)
	}

	if w.ProcessCallbacks.IsMyPlayer(player) { // cpp:3935-3937
		w.ProcessCallbacks.SetViewPosition(spot)
	}
}

// teleportDestination ports the FindClosestValidPosition(PosOkPlayer) search of
// both bodies (missiles.cpp:2951-2957, 3909-3915).
func teleportDestination(w *HitWorld, player int, dst TilePosition) (TilePosition, bool) {
	posOk := func(candidate engine.Point) bool {
		return PosOkPlayer(w, player, TilePosition{X: candidate.X, Y: candidate.Y})
	}

	spot, found := engine.FindClosestValidPosition(posOk, engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}
