// Package missiles: the Hellfire Big Explosion add body.
package missiles

// bigExplosionRadius is the Rectangle radius of the AddBigExplosion damage
// sweep (missiles.cpp:1611: Rectangle{missile.position.tile, 1}).
const bigExplosionRadius = 1

// AddBigExplosion ports AddBigExplosion (missiles.cpp:1603-1618): a player
// cast rolls 2 * (level + 2d10) + 4, scales it by the spell level, stores it in
// _midam and sweeps the 3x3 square around the missile, col-major; then the
// blast seeds an 8-radius light at the launch tile, resets to the default
// frame group and lives _miAnimLen - 1 ticks.
func AddBigExplosion(w *HitWorld, m *Missile, _ *AddParameter) {
	if m.SourceOf() == SourcePlayer { // cpp:1605
		level := w.AddSource.playerStats(m.Source).Level
		dmg := 2*(level+generateRndSum(w, 10, 2)) + 4 // cpp:1606
		dmg = scaleSpellEffect(dmg, m.SpellLevel)     // cpp:1607
		m.Dam = dmg                                   // cpp:1609

		sweepBigExplosion(w, m, dmg) // cpp:1611-1613
	}

	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 8) // cpp:1615
	m.setDefaultFrameGroup()                                                       // cpp:1616
	m.Duration = m.AnimLen - 1                                                     // cpp:1617
}

// sweepBigExplosion ports the PointsInRectangleColMajor(Rectangle{tile, 1})
// loop of AddBigExplosion (missiles.cpp:1611-1613): 3x3 tiles around the
// missile, x major then y minor (engine/points_in_rectangle_range.hpp:258-262).
func sweepBigExplosion(w *HitWorld, m *Missile, dmg int) {
	damageType := DamageTypeFor(m.Type)

	for dx := -bigExplosionRadius; dx <= bigExplosionRadius; dx++ {
		for dy := -bigExplosionRadius; dy <= bigExplosionRadius; dy++ {
			position := TilePosition{
				X: m.Position.Tile.X + dx,
				Y: m.Position.Tile.Y + dy,
			}
			CheckMissileCol(w, m, position, MissileColProbe{
				DamageType:     damageType,
				MinDamage:      dmg,
				MaxDamage:      dmg,
				Shifted:        false,
				Keep:           true,
				OnlyHitWalking: nil,
			})
		}
	}
}
