// Package missiles: the Teleport bodies (shared with Phasing).
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// AddTeleport ports AddTeleport (missiles.cpp:1924-1942): the destination is
// the nearest tile the caster may stand on within 5 of the aim; the missile
// then sits on it with a 2-tick life, and a failed search fizzles the spell.
func AddTeleport(w *HitWorld, m *Missile, p *AddParameter) {
	spot, ok := teleportDestination(w, m.Source, p.Dst)
	if !ok { // cpp:1938-1941
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.Position.Tile = spot // cpp:1935-1936
	m.Position.Start = spot
	m.Duration = 2 // cpp:1937
}

// ProcessTeleport ports ProcessTeleport (missiles.cpp:3656-3692): the first
// tick only counts down, then the body relocates the caster onto the tile the
// search finds — clearing the old dPlayer slot, moving the player's tile,
// future and old positions, re-marking the trans tiles, re-occupying the new
// tile, and moving the light and vision (town keeps them still) — and finally
// parks the local player's view on the new tile.
func ProcessTeleport(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3658

	if m.Duration <= 0 { // cpp:3659-3662
		m.DelFlag = true

		return
	}

	player := m.Source

	spot, ok := teleportDestination(w, player, m.Position.Tile)
	if !ok { // cpp:3673-3674
		return
	}

	w.ProcessCallbacks.RelocatePlayer(player, spot) // cpp:3676-3684
	m.Var1 = 1                                      // cpp:3683

	if w.ProcessCallbacks.LevelType != levelTypeTown { // cpp:3685-3688
		w.ProcessCallbacks.ChangeLightXY(player, spot)
		w.ProcessCallbacks.ChangeVisionXY(player, spot)
	}

	if w.ProcessCallbacks.IsMyPlayer(player) { // cpp:3689-3691
		w.ProcessCallbacks.SetViewPosition(spot)
	}
}

// teleportDestination ports the FindClosestValidPosition(PosOkPlayer) search of
// both bodies (missiles.cpp:1928-1932, 3667-3671).
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
