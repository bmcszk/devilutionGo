// Package missiles: the Guardian sentry's add body.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// AddGuardian ports AddGuardian (missiles.cpp:2770-2800): the guardian is
// placed on the closest free tile within five steps of the aim (cpp:2773-2789)
// and fizzles when the whole neighbourhood is blocked (cpp:2789-2793);
// otherwise it takes a radius-1 light (cpp:2794) and a life of
// min(spell level + level/2, 30) << 4 floored at 30 (cpp:2795-2798).
func AddGuardian(w *HitWorld, m *Missile, p *AddParameter) {
	spawnPosition, ok := findGuardianSpot(w, m.Position.Start, p.Dst)
	if !ok { // cpp:2789-2793
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.DelFlag = false                                                        // cpp:2791
	m.Position.Tile = spawnPosition                                          // cpp:2792
	m.Position.Start = spawnPosition                                         // cpp:2793
	m.Lid = w.ProcessCallbacks.AddLight(spawnPosition.X, spawnPosition.Y, 1) // cpp:2794

	level := w.AddSource.playerStats(m.Source).Level
	m.Duration = minInt(m.SpellLevel+level/2, 30) // cpp:2795-2796
	m.Duration <<= 4                              // cpp:2797
	m.Duration = maxInt(m.Duration, 30)           // cpp:2798
	m.Var1 = m.Duration - m.AnimLen               // cpp:2799
	m.Var3 = 1                                    // cpp:2800
}

// findGuardianSpot ports the placement search (missiles.cpp:2773-2788): a tile
// inside the dungeon, with no monster, object or missile on it, nothing solid
// or missile-blocking, and a clear missile line back to the caster.
func findGuardianSpot(w *HitWorld, start, dst TilePosition) (TilePosition, bool) {
	posOk := func(candidate engine.Point) bool {
		target := TilePosition{X: candidate.X, Y: candidate.Y}

		return guardianSpotFree(w, target) && LineClearMissile(w, start, target) // cpp:2788
	}

	spot, found := engine.FindClosestValidPosition(posOk,
		engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}

// guardianSpotFree ports the per-tile half of the search predicate
// (missiles.cpp:2776-2787): inside the dungeon, no monster, object or missile
// on the tile, and nothing solid or missile-blocking.
func guardianSpotFree(w *HitWorld, target TilePosition) bool {
	if !world.InDungeonBounds(target.X, target.Y) { // cpp:2776-2778
		return false
	}

	if w.DMonster[target.Y][target.X] != 0 { // cpp:2779-2781
		return false
	}

	if w.Grid.FindObjectAtPosition(target.X, target.Y) != nil { // cpp:2782-2784
		return false
	}

	if w.ProcessCallbacks.TileContainsMissile(target) { // cpp:2785-2787
		return false
	}

	return !w.Grid.TileHasAny(target.X, target.Y, world.TileSolid|world.TileBlockMissile)
}

// maxInt is upstream's std::max for the guardian life floor.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
