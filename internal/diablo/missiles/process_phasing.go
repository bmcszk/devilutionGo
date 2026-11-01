// Package missiles: the Phasing bodies and the player-placement predicate they
// share with Teleport.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// PosOkPlayer ports PosOkPlayer (player.cpp:3089-3112): inside the dungeon, on
// a walkable tile, with no other living player and no living monster on it —
// asleep monsters (negative dMonster) and dead ones never block, and in town
// any monster at all blocks.
func PosOkPlayer(w *HitWorld, player int, position TilePosition) bool {
	if !inDungeon(position) { // cpp:3091-3092
		return false
	}

	if !w.Grid.IsTileWalkable(position.X, position.Y) { // cpp:3093-3094
		return false
	}

	if other := w.ProcessCallbacks.PlayerAtPosition(position); // cpp:3095-3097
	other >= 0 && other != player && !w.ProcessCallbacks.PlayerHasNoLife(other) {
		return false
	}

	return posOkPlayerMonsterFree(w, position)
}

// AddPhasing ports AddPhasing (missiles.cpp:2900-2930). A TARGET_BOTH cast —
// the caster's own phasing — just parks the missile on the aim tile, nudged to
// a legal spot when the aim is blocked; otherwise the missile lands on a random
// tile in the ring 4..6 around the launch tile, and the spell fizzles when the
// ring holds no legal tile at all.
func AddPhasing(w *HitWorld, m *Missile, p *AddParameter) {
	m.Duration = 2 // cpp:2902

	if m.Caster == int(TargetBoth) { // cpp:2906-2911
		m.Position.Tile = p.Dst
		if !PosOkPlayer(w, m.Source, p.Dst) {
			updateVileMissPos(w, m, p.Dst)
		}

		return
	}

	targets := phasingRing(w, m)
	if len(targets) == 0 { // cpp:2924-2927
		m.DelFlag = true

		return
	}

	// cpp:2929: GenerateRnd(count), floored at 0.
	m.Position.Tile = targets[maxInt(w.ProcessCallbacks.GenerateRnd(len(targets)), 0)]
}

// posOkPlayerMonsterFree ports the dMonster half of PosOkPlayer
// (player.cpp:3099-3110).
func posOkPlayerMonsterFree(w *HitWorld, position TilePosition) bool {
	monster := int(w.DMonster[position.Y][position.X])
	if monster == 0 {
		return true
	}

	if w.ProcessCallbacks.LevelType == levelTypeTown { // cpp:3100-3102
		return false
	}

	if monster < 0 { // cpp:3103-3105: asleep monsters never block
		return false
	}

	return w.ProcessCallbacks.MonsterHasNoLife(monster - 1) // cpp:3106-3108
}

// phasingRing ports the ring scan (missiles.cpp:2913-2923): every 13x13 tile
// around the launch tile outside the middle 7x7 that the caster may stand on.
func phasingRing(w *HitWorld, m *Missile) []TilePosition {
	targets := make([]TilePosition, 0, 4*9)

	for y := -6; y <= 6; y++ {
		targets = appendPhasingRow(w, m, targets, y)
	}

	return targets
}

// appendPhasingRow collects the legal tiles of one row of the scan
// (missiles.cpp:2914-2923).
func appendPhasingRow(w *HitWorld, m *Missile, targets []TilePosition, y int) []TilePosition {
	for x := -6; x <= 6; x++ {
		target := TilePosition{X: m.Position.Start.X + x, Y: m.Position.Start.Y + y}
		if phasingRingTile(w, m.Source, target, x, y) {
			targets = append(targets, target)
		}
	}

	return targets
}

// phasingRingTile applies the skip-centre and occupancy tests of one scanned
// tile (missiles.cpp:2916-2921).
func phasingRingTile(w *HitWorld, player int, target TilePosition, x, y int) bool {
	if (x >= -3 && x <= 3) || (y >= -3 && y <= 3) { // cpp:2916-2918: skip centre
		return false
	}

	return PosOkPlayer(w, player, target)
}

// updateVileMissPos ports UpdateVileMissPos (missiles.cpp:2932-2947): the
// nearest tile to dst the local player may stand on, searched ring by ring out
// to 49; the missile keeps its old tile when nothing is found.
func updateVileMissPos(w *HitWorld, m *Missile, dst TilePosition) {
	for k := 1; k < 50; k++ {
		spot, found := vileRingSpot(w, w.ProcessCallbacks.MyPlayer(), dst, k)
		if found {
			m.Position.Tile = spot

			return
		}
	}
}

// vileRingSpot scans the square ring of radius k around dst.
func vileRingSpot(w *HitWorld, player int, dst TilePosition, k int) (TilePosition, bool) {
	for j := -k; j <= k; j++ {
		for i := -k; i <= k; i++ {
			target := TilePosition{X: dst.X + i, Y: dst.Y + j}
			if PosOkPlayer(w, player, target) {
				return target, true
			}
		}
	}

	return TilePosition{}, false
}

// inDungeon is the InDungeonBounds guard.
func inDungeon(position TilePosition) bool {
	return world.InDungeonBounds(position.X, position.Y)
}
