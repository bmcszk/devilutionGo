// Package missiles: the Hellfire mana add bodies — AddMana, AddMagi and the
// trap's AddStealMana.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// stealManaRadius is the FindClosestValidPosition radius of AddStealMana
// (missiles.cpp:1465: 0, 2).
const stealManaRadius = 2

// AddMana ports AddMana (missiles.cpp:1655-1674): a d10 plus a d4 per character
// level plus a d6 per spell level, all in the 26.6 scale, scaled by the class's
// splRestoreMana, added to the pool and to the base pool with both capped at
// their maxima; the row then deletes itself and the mana panel is redrawn.
func AddMana(w *HitWorld, m *Missile, _ *AddParameter) {
	player := w.AddSource.playerStats(m.Source)

	manaAmount := (w.ProcessCallbacks.GenerateRnd(10) + 1) << 6 // cpp:1659

	for range player.Level { // cpp:1660-1662
		manaAmount += randomIntBetween(w, 1, 4) << 6
	}

	for range m.SpellLevel { // cpp:1663-1665
		manaAmount += (w.ProcessCallbacks.GenerateRnd(6) + 1) << 6
	}

	manaAmount = int(int64(manaAmount) * int64(w.ProcessCallbacks.ClassManaRestore(player.Class)) >> 6) // cpp:1671

	pool := w.ProcessCallbacks.PlayerMana(m.Source) // cpp:1666-1669
	pool.Mana = min(pool.Mana+manaAmount, pool.MaxMana)
	pool.ManaBase = min(pool.ManaBase+manaAmount, pool.MaxManaBase)
	w.ProcessCallbacks.SetPlayerMana(m.Source, pool.Mana, pool.ManaBase)

	m.DelFlag = true                // cpp:1670
	w.ProcessCallbacks.RedrawMana() // cpp:1673
}

// AddMagi ports AddMagi (missiles.cpp:1676-1684): refill both mana pools to
// their maxima, delete the row and redraw the mana panel.
func AddMagi(w *HitWorld, m *Missile, _ *AddParameter) {
	pool := w.ProcessCallbacks.PlayerMana(m.Source) // cpp:1678-1680
	w.ProcessCallbacks.SetPlayerMana(m.Source, pool.MaxMana, pool.MaxManaBase)

	m.DelFlag = true                // cpp:1681
	w.ProcessCallbacks.RedrawMana() // cpp:1682
}

// AddStealMana ports AddStealMana (missiles.cpp:1458-1476): the closest tile
// within 2 of the launch tile that holds a player loses its whole mana pool,
// with the base pool rebased to keep the equipment bonus, then the item values
// recalculate, the mana panel redraws and the "Pig" sound plays at that tile.
// The row always deletes itself.
func AddStealMana(w *HitWorld, m *Missile, _ *AddParameter) {
	spot, found := findTrappedPlayerTile(w, m.Position.Start)
	if found {
		player := int(absI16(w.DPlayer[spot.Y][spot.X])) - 1

		pool := w.ProcessCallbacks.PlayerMana(player)               // cpp:1468
		pool.Mana = 0                                               // cpp:1468
		pool.ManaBase = pool.Mana + pool.MaxManaBase - pool.MaxMana // cpp:1469
		w.ProcessCallbacks.SetPlayerMana(player, pool.Mana, pool.ManaBase)
		w.ProcessCallbacks.CalcPlayerInv(player)             // cpp:1470
		w.ProcessCallbacks.RedrawMana()                      // cpp:1471
		w.ProcessCallbacks.PlaySfxLoc("Pig", spot.X, spot.Y) // cpp:1473
	}

	m.DelFlag = true // cpp:1475
}

// findTrappedPlayerTile ports the FindClosestValidPosition call of AddStealMana
// (missiles.cpp:1460-1466): an in-dungeon tile within 2 of start that a player
// occupies.
func findTrappedPlayerTile(w *HitWorld, start TilePosition) (TilePosition, bool) {
	spot, found := engine.FindClosestValidPosition(
		func(target engine.Point) bool {
			return world.InDungeonBounds(target.X, target.Y) && w.DPlayer[target.Y][target.X] != 0
		},
		engine.Point{X: start.X, Y: start.Y}, 0, stealManaRadius,
	)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}
