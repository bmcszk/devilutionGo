// Package missiles: the Apocalypse box walk's bodies.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// AddApocalypse ports AddApocalypse (missiles.cpp:2652-2665): the spell fills
// a 16x16 box around the caster's tile, clamped to the dungeon
// (cpp:2657-2661), rolls GenerateRndSum(6, level) + level as the boom damage
// (cpp:2663) and lives 255 ticks (cpp:2664).
func AddApocalypse(w *HitWorld, m *Missile, _ *AddParameter) {
	m.Var1 = 8                                             // cpp:2656
	m.Var2 = maxInt(m.Position.Start.Y-8, 1)               // cpp:2657
	m.Var3 = minInt(m.Position.Start.Y+8, world.MAXDUNY-1) // cpp:2658
	m.Var4 = maxInt(m.Position.Start.X-8, 1)               // cpp:2659
	m.Var5 = minInt(m.Position.Start.X+8, world.MAXDUNX-1) // cpp:2660
	m.Var6 = m.Var4                                        // cpp:2661

	level := w.AddSource.playerStats(m.Source).Level
	m.Dam = generateRndSum(w, 6, level) + level // cpp:2663
	m.Duration = 255                            // cpp:2664
}

// ProcessApocalypse ports ProcessApocalypse (missiles.cpp:3850-3872): each
// tick walks the box from the remembered cursor and drops one ApocalypseBoom
// on the first target it finds, then parks the cursor just past it so the next
// tick resumes where this one stopped (cpp:3866-3870). An exhausted box
// deletes the missile (cpp:3872).
func ProcessApocalypse(w *HitWorld, m *Missile) {
	for j := m.Var2; j < m.Var3; j++ { // cpp:3852
		for k := m.Var4; k < m.Var5; k++ { // cpp:3853
			if apocalypseSkip(w, m, k, j) {
				continue
			}

			w.ProcessCallbacks.SpawnMissile(TilePosition{X: k, Y: j}, TilePosition{X: k, Y: j}, // cpp:3865
				w.ProcessCallbacks.PlayerDirection(m.Source), ApocalypseBoom,
				int(TargetMonsters), m.Source, m.Dam, 0, nil)

			m.Var2 = j     // cpp:3866
			m.Var4 = k + 1 // cpp:3867

			return
		}

		m.Var4 = m.Var6 // cpp:3870
	}

	m.DelFlag = true // cpp:3872
}

// apocalypseSkip reports whether a box tile is not a boom target: no monster,
// a player minion, a solid tile, or — under Hellfire — a monster the missile
// cannot reach in a straight line (missiles.cpp:3854-3862).
func apocalypseSkip(w *HitWorld, m *Missile, x, y int) bool {
	idx := int(w.DMonster[y][x]) - 1 // cpp:3854-3856
	if idx < 0 {
		return true
	}

	if w.ProcessCallbacks.MonsterIsPlayerMinion(idx) { // cpp:3857-3858
		return true
	}

	if w.Grid.TileHasAny(x, y, world.TileSolid) { // cpp:3859-3860
		return true
	}

	// cpp:3861-3862
	return w.ProcessCallbacks.IsHellfire() &&
		!LineClearMissile(w, m.Position.Tile, TilePosition{X: x, Y: y})
}
