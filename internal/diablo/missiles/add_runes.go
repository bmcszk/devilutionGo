// Package missiles: the Hellfire rune add bodies — AddRune and the per-runon
// rows that select which missile the rune detonates.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// runeSearchRadius is the maximum radius of the rune placement search
// (missiles.cpp:683: FindClosestValidPosition(..., dst, 0, 9)).
const runeSearchRadius = 9

// AddRuneOfFire ports AddRuneOfFire (missiles.cpp:1250-1253): the rune
// detonates a BigExplosion.
func AddRuneOfFire(w *HitWorld, m *Missile, p *AddParameter) {
	addRune(w, m, p.Dst, BigExplosion)
}

// AddRuneOfLight ports AddRuneOfLight (missiles.cpp:1255-1261): the damage is
// 16 × (2d10 + the caster's level + 2), scaled in before the rune placement, so
// a fizzled rune still carries it; the rune detonates a LightningWall.
func AddRuneOfLight(w *HitWorld, m *Missile, p *AddParameter) {
	level := 0
	if m.SourceOf() == SourcePlayer { // cpp:1257
		level = w.AddSource.playerStats(m.Source).Level
	}

	m.Dam = 16 * (generateRndSum(w, 10, 2) + level + 2) // cpp:1258

	addRune(w, m, p.Dst, LightningWall) // cpp:1259
}

// AddRuneOfNova ports AddRuneOfNova (missiles.cpp:1263-1266): the rune
// detonates a Nova.
func AddRuneOfNova(w *HitWorld, m *Missile, p *AddParameter) {
	addRune(w, m, p.Dst, Nova)
}

// AddRuneOfImmolation ports AddRuneOfImmolation (missiles.cpp:1268-1271): the
// rune detonates an Immolation.
func AddRuneOfImmolation(w *HitWorld, m *Missile, p *AddParameter) {
	addRune(w, m, p.Dst, Immolation)
}

// AddRuneOfStone ports AddRuneOfStone (missiles.cpp:1273-1276): the rune
// detonates a StoneCurse.
func AddRuneOfStone(w *HitWorld, m *Missile, p *AddParameter) {
	addRune(w, m, p.Dst, StoneCurse)
}

// addRune ports AddRune (missiles.cpp:668-693): with a clear missile line to
// the aim, the rune snaps to the closest free tile within 9 of it, remembers
// its own missile id in var1 (int8 upstream) and seeds an 8-radius light;
// otherwise it deletes itself.
func addRune(w *HitWorld, m *Missile, dst TilePosition, missileID MissileID) {
	if !LineClearMissile(w, m.Position.Start, dst) { // cpp:670
		m.DelFlag = true

		return
	}

	spot, found := findRunePosition(w, dst)
	if !found {
		m.DelFlag = true // cpp:691

		return
	}

	m.Position.Tile = spot                                 // cpp:686
	m.Var1 = int(int8(missileID))                          // cpp:687
	m.Lid = w.ProcessCallbacks.AddLight(spot.X, spot.Y, 8) // cpp:688
}

// findRunePosition ports the FindClosestValidPosition predicate of AddRune
// (missiles.cpp:671-684): in the dungeon, no object, no missile and no solid
// piece on the tile.
func findRunePosition(w *HitWorld, dst TilePosition) (TilePosition, bool) {
	spot, found := engine.FindClosestValidPosition(
		func(target engine.Point) bool {
			if !world.InDungeonBounds(target.X, target.Y) { // cpp:673-675
				return false
			}

			tile := TilePosition{X: target.X, Y: target.Y}
			if w.ProcessCallbacks.IsObjectAtPosition(tile) { // cpp:676-678
				return false
			}

			if w.ProcessCallbacks.TileContainsMissile(tile) { // cpp:679-681
				return false
			}

			return !w.Grid.TileHasAny(target.X, target.Y, world.TileSolid) // cpp:682-684
		},
		engine.Point{X: dst.X, Y: dst.Y}, 0, runeSearchRadius,
	)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}
