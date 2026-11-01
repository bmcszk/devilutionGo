// Package missiles: the lightning control bolt's per-tick body and the
// SpawnLightning tail it shares with the Lightning row.
package missiles

import (
	"slices"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// The MT_STORM family ordinals (tables/monstdat.h:262-265); monstdat.tsv file
// lines 78-81 give data rows 76-79 on the MT_NZOMBIE = 0 base.
const (
	monsterTypeStorm      = 76
	monsterTypeStormRider = 77
	monsterTypeStormLord  = 78
	monsterTypeMaelstrom  = 79
)

// stormMonsterTypes is the MT_STORM family that forks thin lightning instead of
// the regular bolt (missiles.cpp:826-827; tables/monstdat.h:262-265).
var stormMonsterTypes = []int{monsterTypeStorm, monsterTypeStormRider, monsterTypeStormLord, monsterTypeMaelstrom}

// AddLightningControl ports AddLightningControl (missiles.cpp:2782-2790): the
// bolt remembers its launch tile, flies at 32 pixels per tick toward the aim
// tile, starts on a random frame of 1..8 and lives 256 ticks.
func AddLightningControl(w *HitWorld, m *Missile, p *AddParameter) {
	m.Var1 = m.Position.Start.X // cpp:2784
	m.Var2 = m.Position.Start.Y // cpp:2785

	UpdateMissileVelocity(m, p.Dst, 32) // cpp:2786

	m.AnimFrame = w.ProcessCallbacks.GenerateRnd(8) + 1 // cpp:2787 RandomIntBetween(1, 8)
	m.Duration = 256                                    // cpp:2788
}

// ProcessLightningControl ports ProcessLightningControl (missiles.cpp:3238-3262):
// the damage is re-rolled every tick — currlevel-based for a trap, level-based
// for a player, twice the monster's range otherwise — and handed to
// SpawnLightning, which walks the bolt and forks the visible lightning.
func ProcessLightningControl(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3240

	SpawnLightning(w, m, lightningControlDamage(w, m)) // cpp:3242-3260
}

// SpawnLightning ports SpawnLightning (missiles.cpp:802-849): the bolt walks
// one tick forward, stopping dead on a blocking tile, and then forks the
// visible Lightning (ThinLightning for the storm monsters) back at its launch
// tile, remembering the tile it reached.
func SpawnLightning(w *HitWorld, m *Missile, dam int) {
	m.Duration-- // cpp:804

	walkLightningControl(w, m) // cpp:805-817

	position := m.Position.Tile
	if lightningControlForks(w, m, position) { // cpp:820-821
		spawnID := Lightning
		if forksThinLightning(w, m) { // cpp:826-827
			spawnID = ThinLightning
		}

		w.ProcessCallbacks.SpawnMissile(position, m.Position.Start, South, spawnID, // cpp:829-838
			m.Caster, m.Source, dam, m.SpellLevel, m)

		m.Var1 = position.X // cpp:839-840
		m.Var2 = position.Y
	}

	if m.Duration == 0 { // cpp:846-848
		m.DelFlag = true
	}
}

// walkLightningControl ports the MoveMissile call of SpawnLightning
// (missiles.cpp:805-817): a trap's own launch tile never blocks the bolt, any
// other blocking tile kills it on the spot.
func walkLightningControl(w *HitWorld, m *Missile) {
	MoveMissile(w, m, func(tile TilePosition) bool {
		if m.IsTrap() && tile == m.Position.Start {
			return true
		}

		if w.Grid.TileHasAny(tile.X, tile.Y, world.TileBlockMissile) {
			m.Duration = 0

			return false
		}

		return true
	}, DontMoveOnFailedProbe(false))
}

// lightningControlForks ports the fork guard (missiles.cpp:820-821): the tile
// must not block missile flight, must differ from the remembered tile and must
// be inside the dungeon.
func lightningControlForks(w *HitWorld, m *Missile, position TilePosition) bool {
	if w.Grid.TileHasAny(position.X, position.Y, world.TileBlockMissile) {
		return false
	}

	return position != (TilePosition{X: m.Var1, Y: m.Var2}) && world.InDungeonBounds(position.X, position.Y)
}

// forksThinLightning ports the storm-monster test of SpawnLightning
// (missiles.cpp:825-828): a monster source whose type is in the MT_STORM
// family uses the thin bolt.
func forksThinLightning(w *HitWorld, m *Missile) bool {
	if m.IsTrap() || m.Caster == int(TargetMonsters) {
		return false
	}

	stats := w.AddSource.monsterStats(m.Source)

	return slices.Contains(stormMonsterTypes, stats.Type)
}

// lightningControlDamage ports the three damage branches (missiles.cpp:3243-3259).
func lightningControlDamage(w *HitWorld, m *Missile) int {
	switch {
	case m.IsTrap(): // cpp:3244-3246
		return w.ProcessCallbacks.GenerateRnd(w.ProcessCallbacks.Currlevel) +
			2*w.ProcessCallbacks.Currlevel
	case m.Caster == int(TargetMonsters): // cpp:3247-3250
		return (w.ProcessCallbacks.GenerateRnd(2) +
			w.ProcessCallbacks.GenerateRnd(w.AddSource.playerStats(m.Source).Level) + 2) << 6
	default: // cpp:3251-3253
		stats := w.AddSource.monsterStats(m.Source)

		return 2 * randomIntBetween(w, stats.MinDamage, stats.MaxDamage)
	}
}
