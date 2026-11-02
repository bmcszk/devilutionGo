// Package missiles: the Hellfire bow add bodies — Immolation (fireball bow),
// Lightning Bow, Charged Bolt Bow and the rogue's Spectral Arrow.
package missiles

// AddSpectralArrow ports AddSpectralArrow (missiles.cpp:1479-1505): the rogue
// arrow speed bonus is the class level step (Rogue /4, Warrior and Bard /8)
// plus the item fast-attack bits (1, 2, 4, 8); the arrow lives one tick and
// carries its aim tile in var1/var2 and the bonus in var3.
func AddSpectralArrow(w *HitWorld, m *Missile, p *AddParameter) {
	av := spectralArrowBonus(w, m)

	m.Duration = 1 // cpp:1501
	m.Var1 = p.Dst.X
	m.Var2 = p.Dst.Y
	m.Var3 = av
}

// AddImmolation ports AddImmolation (missiles.cpp:1620-1634): the fireball bow
// shot leaves at speed 16 plus one per spell level (capped at 34) for player
// casters, faces its aim, burns 256 ticks and seeds an 8-radius light.
func AddImmolation(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1622-1624
		dst = shiftTile(dst, p.Midir)
	}

	sp := 16
	if m.Caster == int(TargetMonsters) { // cpp:1626-1628
		sp += min(m.SpellLevel, 34)
	}

	UpdateMissileVelocity(m, dst, sp)                        // cpp:1629
	SetDirection16(m, GetDirection16(m.Position.Start, dst)) // cpp:1630
	m.Duration = 256                                         // cpp:1631
	m.Lid = addMissileLight(w, m)                            // cpp:1632
}

// AddLightningBow ports AddLightningBow (missiles.cpp:1636-1653): speed 32, a
// random animation start frame, 255 ticks, the caster's tile in var1/var2 (the
// launch tile for a trap) and the pre-rolled damage moved to the 26.6 scale.
func AddLightningBow(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1638-1640
		dst = shiftTile(dst, p.Midir)
	}

	UpdateMissileVelocity(m, dst, 32)       // cpp:1641
	m.AnimFrame = randomIntBetween(w, 1, 8) // cpp:1642
	m.Duration = 255                        // cpp:1643

	if m.Source < 0 { // cpp:1644-1650
		m.Var1 = m.Position.Start.X
		m.Var2 = m.Position.Start.Y
	} else {
		tile := w.AddSource.playerStats(m.Source).Tile
		m.Var1 = tile.X
		m.Var2 = tile.Y
	}

	m.Dam <<= 6 // cpp:1651
}

// AddChargedBoltBow ports AddChargedBoltBow (missiles.cpp:1716-1731): the
// random bolt count, the flat 15 damage of a non-player caster, the shifted aim
// tile, a random animation start frame, a 5-radius light, speed 8, var1 = 5,
// the aim direction in var2 and 256 ticks.
func AddChargedBoltBow(w *HitWorld, m *Missile, p *AddParameter) {
	m.Mirnd = uint8(w.ProcessCallbacks.GenerateRnd(15) + 1) // cpp:1718
	if m.Caster != int(TargetMonsters) {                    // cpp:1719-1721
		m.Dam = 15
	}

	dst := p.Dst
	if m.Position.Start == dst { // cpp:1723-1725
		dst = shiftTile(dst, p.Midir)
	}

	m.AnimFrame = randomIntBetween(w, 1, 8)                                        // cpp:1726
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 5) // cpp:1727
	UpdateMissileVelocity(m, dst, 8)                                               // cpp:1728
	m.Var1 = 5
	m.Var2 = int(p.Midir)
	m.Duration = 256 // cpp:1731
}

// spectralArrowBonus ports the speed roll of AddSpectralArrow
// (missiles.cpp:1481-1499): the class level step and the four item
// attack-speed bits, or 0 for a non-player caster.
func spectralArrowBonus(w *HitWorld, m *Missile) int {
	if m.SourceOf() != SourcePlayer { // cpp:1483
		return 0
	}

	player := w.AddSource.playerStats(m.Source)

	av := 0
	switch player.Class {
	case HeroClassRogue: // cpp:1486-1487
		av += (player.Level - 1) / 4
	case HeroClassWarrior, HeroClassBard: // cpp:1488-1489
		av += (player.Level - 1) / 8
	default: // every other class has no level step
	}

	if player.hasItemFlag(ItemSpecialEffectQuickAttack) { // cpp:1491-1492
		av++
	}

	if player.hasItemFlag(ItemSpecialEffectFastAttack) { // cpp:1493-1494
		av += 2
	}

	if player.hasItemFlag(ItemSpecialEffectFasterAttack) { // cpp:1495-1496
		av += 4
	}

	if player.hasItemFlag(ItemSpecialEffectFastestAttack) { // cpp:1497-1498
		av += 8
	}

	return av
}
