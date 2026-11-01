// Package missiles: the Golem's add body.
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// AddGolem ports AddGolem (missiles.cpp:2620-2660). The missile always deletes
// itself: a live golem belonging to the caster is killed instead, otherwise the
// body looks for a free, missile-reachable tile within 5 of the aim
// (cpp:2636-2638) and — only for the local player — spawns the golem, or asks
// the level owner to spawn it, so multiplayer stays in sync (cpp:2642-2658).
func AddGolem(w *HitWorld, m *Missile, p *AddParameter) {
	m.DelFlag = true // cpp:2622

	player := m.Source

	if w.ProcessCallbacks.FindGolemForPlayer(player) { // cpp:2626-2630
		w.ProcessCallbacks.KillGolem(player)

		return
	}

	spawn, ok := findGolemSpawn(w, m, p.Dst)
	if !ok || !w.ProcessCallbacks.IsMyPlayer(player) { // cpp:2636-2643
		return
	}

	if !w.ProcessCallbacks.IsLevelOwnedByLocalClient(player) { // cpp:2647-2654
		w.ProcessCallbacks.RequestSpawnGolem(spawn)

		return
	}

	w.ProcessCallbacks.SpawnGolem(player, spawn, m.SpellLevel) // cpp:2656
}

// findGolemSpawn ports the FindClosestValidPosition predicate (missiles.cpp:2634-2638):
// an unoccupied tile with a clear missile line from the launch tile.
func findGolemSpawn(w *HitWorld, m *Missile, dst TilePosition) (TilePosition, bool) {
	start := m.Position.Start

	posOk := func(candidate engine.Point) bool {
		return !w.ProcessCallbacks.IsTileOccupied(candidate.X, candidate.Y) &&
			LineClearMissile(w, start, TilePosition{X: candidate.X, Y: candidate.Y})
	}

	spot, found := engine.FindClosestValidPosition(posOk, engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}
