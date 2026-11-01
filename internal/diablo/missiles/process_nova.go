// Package missiles: the nova wave's per-tick body.
package missiles

// novaQuarterRadius ports the quarterRadius array of ProcessNovaCommon
// (missiles.cpp:3293): the nine tile offsets of one quadrant, {4,0}..{0,4}.
var novaQuarterRadius = [9]TilePosition{
	{X: 4, Y: 0},
	{X: 4, Y: 1},
	{X: 4, Y: 2},
	{X: 4, Y: 3},
	{X: 4, Y: 4},
	{X: 3, Y: 4},
	{X: 2, Y: 4},
	{X: 1, Y: 4},
	{X: 0, Y: 4},
}

// ProcessNova ports ProcessNova (missiles.cpp:3309-3312), which delegates to
// ProcessNovaCommon with the NovaBall projectile (cpp:3280-3304): a full circle
// is fired as 9 quarter offsets x 4 flips = 36 missiles aimed at the caster's
// own facing, then the nova counts down and deletes on the tick reaching 0.
func ProcessNova(w *HitWorld, m *Missile) {
	novaWave(w, m)
}

// novaWave ports the ProcessNovaCommon body (cpp:3282-3303).
func novaWave(w *HitWorld, m *Missile) {
	src := m.Position.Tile
	dir, target := novaAim(w, m)

	for _, offset := range novaQuarterRadius {
		for _, aimed := range novaFlips(offset) { // cpp:3297-3300
			w.ProcessCallbacks.SpawnMissile(src, TilePosition{X: src.X + aimed.X, Y: src.Y + aimed.Y},
				dir, NovaBall, target, m.Source, m.Dam, m.SpellLevel, nil)
		}
	}

	m.Duration-- // cpp:3301-3303

	if m.Duration == 0 {
		m.DelFlag = true
	}
}

// novaAim ports the direction/target pick of the wave (missiles.cpp:3286-3290):
// a trap fires south at the players, a cast nova fires at the monsters along the
// caster's facing.
func novaAim(w *HitWorld, m *Missile) (Direction, int) {
	if m.IsTrap() {
		return South, int(TargetPlayers)
	}

	return w.ProcessCallbacks.PlayerDirection(m.Source), int(TargetMonsters)
}

// novaFlips ports the four offsets of one quarter step (missiles.cpp:3297):
// the offset itself, its XY flip, its X flip and its Y flip — the four
// quadrants of the ring.
func novaFlips(offset TilePosition) [4]TilePosition {
	return [4]TilePosition{
		offset,
		{X: offset.Y, Y: offset.X},
		{X: -offset.X, Y: offset.Y},
		{X: offset.X, Y: -offset.Y},
	}
}
